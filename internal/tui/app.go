package tui

import (
	"context"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/olafurb/krowser/internal/k8s"
)

// tuiCtx is the base context for all Kubernetes API calls.
var tuiCtx = context.Background()

// cmdMsg wraps a message as a command.
func cmdMsg(m tea.Msg) tea.Cmd { return func() tea.Msg { return m } }

// quitCmd returns a command that quits the program.
func quitCmd() tea.Cmd { return func() tea.Msg { return tea.Quit() } }

type screen int

// Screen identifiers.
const (
	screenResources screen = iota
	screenDetail
	screenLogs
	screenPickerNamespaces
	screenPickerContexts
	screenPickerResources
	screenPickerContainers
	screenPickerThemes
	screenTop
	screenForward
)

// screen is any renderable, updateable view managed by the root model.
type screenView interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (screenView, tea.Cmd)
	View() string
	Resize(width, height int)
}

// Config carries startup options into the model.
type Config struct {
	Client        *k8s.Client
	Contexts      []k8s.ContextInfo
	Kubeconfig    string
	Namespace     string
	AllNamespaces bool
	ContextPicker bool
	Theme         string // theme name, empty for the default (Monokai)
	ThemeFile     string // path to persist the selected theme, empty for the default location
}

// Model is the root Bubble Tea model.
type Model struct {
	client     *k8s.Client
	contexts   []k8s.ContextInfo
	kubeconfig string
	context    string

	theme     Theme
	themeFile string
	km        KeyMap
	helpM     help.Model

	width  int
	height int

	res   *k8s.Resource
	ns    string
	allNs bool

	cur  screenView
	prev screenView

	resource *resourceScreen
	detail   *detailScreen
	logs     *logsScreen
	picker   *pickerScreen
	top      *topScreen

	toasts *toastManager
	dlg    *dialog

	helpVisible bool
	statusText  string
}

// New builds the root model.
func New(cfg Config) *Model {
	themeFile := cfg.ThemeFile
	if themeFile == "" {
		themeFile = ThemeConfigPath()
	}
	themeName := cfg.Theme
	if themeName == "" {
		themeName = LoadSavedTheme(themeFile)
	}
	theme, ok := ThemeByName(themeName)
	if !ok {
		theme = DefaultTheme
	}
	m := &Model{
		client:     cfg.Client,
		contexts:   cfg.Contexts,
		kubeconfig: cfg.Kubeconfig,
		context:    cfg.Client.Context(),
		theme:      theme,
		themeFile:  themeFile,
		km:         DefaultKeyMap(),
		helpM:      help.New(),
		ns:         cfg.Namespace,
		allNs:      cfg.AllNamespaces,
		toasts:     newToastManager(theme),
		dlg:        newDialog(),
	}
	if m.ns == "" {
		m.ns = "default"
	}
	if cfg.ContextPicker {
		m.openPicker(pickerContexts, true)
		return m
	}
	m.openResourceScreen()
	return m
}

func (m *Model) openResourceScreen() {
	m.prev = nil
	if m.res == nil {
		m.res = &k8s.Resources[2] // Pods
	}
	ns := m.ns
	if m.allNs || !m.res.Namespaced {
		ns = ""
	}
	m.resource = newResourceScreen(m.client, m.theme, m.res, ns, m.allNs)
	m.resource.Resize(m.width, m.height)
	m.cur = m.resource
}

func (m *Model) openPicker(mode pickerMode, start bool) {
	switch mode {
	case pickerContexts:
		m.picker = newContextPicker(m.theme, m.contexts)
	case pickerNamespaces:
		m.picker = newNamespacePicker(m.client, m.theme, nil)
	case pickerResources:
		m.picker = newResourcePicker(m.client, m.theme)
	case pickerContainers:
		if m.logs != nil {
			m.picker = newContainerPicker(m.theme, m.logs.containers)
		} else {
			m.picker = newContainerPicker(m.theme, nil)
		}
	case pickerThemes:
		m.picker = newThemePicker(m.theme)
	}
	if start {
		m.cur = m.picker
	} else {
		m.prev = m.cur
		m.cur = m.picker
	}
	m.picker.Resize(m.width, m.height)
}

// Init starts the application.
func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	if m.cur != nil {
		cmds = append(cmds, m.cur.Init())
	}
	return tea.Batch(cmds...)
}

// Update dispatches messages to the active screen and handles global state.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.toasts.width = msg.Width
		m.helpM.SetWidth(msg.Width)
		if m.cur != nil {
			m.cur.Resize(msg.Width, msg.Height)
		}
		return m, nil
	case tea.KeyPressMsg:
		if m.dlg.active {
			return m.handleDialogKey(msg)
		}
		switch msg.String() {
		case "ctrl+c":
			return m, quitCmd()
		case "?":
			m.helpVisible = !m.helpVisible
			return m, nil
		}
	}

	if m.dlg.active {
		return m.handleDialogUpdate(msg)
	}
	if m.helpVisible {
		if _, ok := msg.(tea.KeyPressMsg); ok {
			m.helpVisible = false
		}
		return m, nil
	}

	return m.handleScreenMessage(msg)
}

// handleScreenMessage forwards messages to the active screen and handles
// navigation messages emitted by screens.
func (m *Model) handleScreenMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.cur == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case helpToggleMsg:
		m.helpVisible = !m.helpVisible
		return m, nil
	case backMsg:
		return m.back()
	case setScreenMsg:
		switch msg.screen {
		case screenPickerContexts:
			m.openPicker(pickerContexts, false)
		case screenPickerNamespaces:
			m.openPicker(pickerNamespaces, false)
		case screenPickerResources:
			m.openPicker(pickerResources, false)
		case screenPickerContainers:
			m.openPicker(pickerContainers, false)
		case screenPickerThemes:
			m.openPicker(pickerThemes, false)
		case screenDetail, screenLogs, screenTop, screenForward:
			return m, nil
		}
		return m, m.cur.Init()
	case setResourceMsg:
		m.res = msg.res
		m.openResourceScreen()
		return m, m.cur.Init()
	case setNamespaceMsg:
		m.allNs = msg.all
		m.ns = msg.ns
		if m.ns == "" && !m.allNs {
			m.ns = "default"
		}
		m.openResourceScreen()
		return m, m.cur.Init()
	case openDetailMsg:
		m.detail = newDetailScreen(m.client, m.theme, msg.res, msg.obj, msg.kind)
		m.detail.Resize(m.width, m.height)
		m.prev = m.cur
		m.cur = m.detail
		return m, m.cur.Init()
	case openLogsMsg:
		m.logs = newLogsScreen(m.client, m.theme, msg.res, msg.obj)
		m.logs.Resize(m.width, m.height)
		m.prev = m.cur
		m.cur = m.logs
		return m, m.cur.Init()
	case openTopMsg:
		m.top = newTopScreen(m.client, m.theme, msg.res, msg.ns, msg.allNs)
		m.top.Resize(m.width, m.height)
		m.prev = m.cur
		m.cur = m.top
		return m, m.cur.Init()
	case openDialogMsg:
		m.dlg.show(msg.req)
		return m, nil
	case toastMsg:
		cmd := m.toasts.push(msg.kind, msg.title, msg.body)
		return m, cmd
	case toastDismissMsg:
		m.toasts.dismiss()
		return m, nil
	case actionResultMsg:
		return m.handleActionResult(msg)
	case selectedContextMsg:
		return m.switchContext(msg.name)
	case selectedNamespaceMsg:
		m.ns = msg.name
		m.allNs = false
		m.openResourceScreen()
		return m, m.cur.Init()
	case selectedResourceMsg:
		m.res = msg.res
		m.openResourceScreen()
		return m, m.cur.Init()
	case selectedContainerMsg:
		if m.logs != nil {
			cmd := m.logs.SetContainer(msg.name)
			m.cur = m.logs
			if m.resource != nil {
				m.prev = m.resource
			}
			return m, cmd
		}
		return m, nil
	case selectedThemeMsg:
		if t, ok := ThemeByName(msg.name); ok {
			m.applyTheme(t)
			if err := SaveTheme(m.themeFile, t.Name); err != nil {
				m.toasts.push(ToastError, "settings", "failed to save theme: "+err.Error())
			}
		}
		m.openResourceScreen()
		return m, m.cur.Init()
	case statusMsg:
		m.statusText = msg.text
		return m, nil
	}

	next, cmd := m.cur.Update(msg)
	if next != m.cur {
		m.cur = next
	}
	return m, cmd
}

func (m *Model) back() (tea.Model, tea.Cmd) {
	if m.prev != nil {
		m.cur = m.prev
		m.prev = nil
		return m, nil
	}
	m.openResourceScreen()
	return m, m.cur.Init()
}

func (m *Model) switchContext(name string) (tea.Model, tea.Cmd) {
	if name == "" || name == m.context {
		m.openResourceScreen()
		return m, m.cur.Init()
	}
	client, err := k8s.New(m.kubeconfig, name)
	if err != nil {
		m.toasts.push(ToastError, "context", "failed to switch: "+err.Error())
		m.openResourceScreen()
		return m, m.cur.Init()
	}
	m.client = client
	m.context = name
	m.openResourceScreen()
	return m, m.cur.Init()
}

func (m *Model) handleDialogKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.dlg.req.input {
		switch msg.String() {
		case "esc":
			m.dlg.hide()
			return m, nil
		case "enter":
			value := m.dlg.input.Value()
			run := m.dlg.req.run
			m.dlg.hide()
			return m, runDialogAction(run, value)
		}
		t, cmd := m.dlg.input.Update(msg)
		m.dlg.input = t
		return m, cmd
	}
	switch msg.String() {
	case "y", "enter":
		run := m.dlg.req.run
		m.dlg.hide()
		return m, runDialogAction(run, "")
	case "n", "esc":
		m.dlg.hide()
		return m, nil
	}
	return m, nil
}

func (m *Model) handleDialogUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dlg.req.input {
		t, cmd := m.dlg.input.Update(msg)
		m.dlg.input = t
		return m, cmd
	}
	return m, nil
}

func runDialogAction(run func(string) error, value string) tea.Cmd {
	if run == nil {
		return nil
	}
	return func() tea.Msg {
		if err := run(value); err != nil {
			return actionResultMsg{success: false, title: "Action failed", message: err.Error()}
		}
		return actionResultMsg{success: true, title: "Success", message: "action applied"}
	}
}

func (m *Model) handleActionResult(msg actionResultMsg) (tea.Model, tea.Cmd) {
	if msg.success {
		m.toasts.push(ToastSuccess, msg.title, msg.message)
	} else {
		m.toasts.push(ToastError, msg.title, msg.message)
	}
	if m.resource != nil {
		return m, m.resource.Reload()
	}
	return m, nil
}

// applyTheme switches the active theme across all screens.
func (m *Model) applyTheme(t Theme) {
	m.theme = t
	m.toasts.theme = t
	if m.resource != nil {
		m.resource.theme = t
	}
	if m.detail != nil {
		m.detail.theme = t
	}
	if m.logs != nil {
		m.logs.theme = t
		m.logs.hl = logHighlighter(t)
		m.logs.rehighlight()
		m.logs.renderContent()
	}
	if m.picker != nil {
		m.picker.theme = t
	}
	if m.top != nil {
		m.top.theme = t
		m.top.updatePalette()
	}
}

// View renders the current screen with the shared chrome.
func (m *Model) View() tea.View {
	v := tea.NewView(m.layout())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = "krowser · " + m.context + " · " + m.resourceTitle()
	return v
}

func (m *Model) resourceTitle() string {
	if m.res == nil {
		return ""
	}
	t := m.res.Title
	if m.res.Namespaced {
		if m.allNs {
			t += " · all-namespaces"
		} else {
			t += " · " + m.ns
		}
	}
	return t
}

// layout composes header, content, toasts, status and help bars.
func (m *Model) layout() string {
	content := ""
	if m.cur != nil {
		content = m.cur.View()
	}

	header := m.theme.Header("krowser") +
		"  " + m.theme.AccentText(m.context) +
		"  " + m.theme.MutedText(m.resourceTitle())

	var parts []string
	parts = append(parts, header)

	toastLines := m.toasts.height()
	bodyHeight := max(1, m.height-len(strings.Split(header, "\n"))-toastLines-2)
	body := lipgloss.NewStyle().Width(m.width).MaxHeight(bodyHeight).Render(content)
	parts = append(parts, body)

	if t := m.toasts.View(); t != "" {
		parts = append(parts, t)
	}

	status := m.statusLine()
	parts = append(parts, status)

	parts = append(parts, m.helpM.View(m.km))

	out := strings.Join(parts, "\n")

	if m.dlg.active {
		out = m.dlg.View(m.width, m.height, m.theme)
	} else if m.helpVisible {
		out = helpOverlay(m.km, m.theme, m.width, m.height)
	}
	return out
}

func (m *Model) statusLine() string {
	var b strings.Builder
	b.WriteString(m.theme.MutedText(m.context))
	if m.cur == m.resource && m.resource != nil {
		total := len(m.resource.rows)
		b.WriteString("  " + m.theme.MutedText(m.resourceTitle()))
		b.WriteString("  " + m.theme.AccentText(m.statusText))
		b.WriteString("  " + m.theme.MutedText("("+itoa(total)+" items)"))
	}
	return b.String()
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
