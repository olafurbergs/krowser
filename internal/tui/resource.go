package tui

import (
	"errors"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/olafurb/krowser/internal/k8s"
	"github.com/sahilm/fuzzy"
)

var errInvalidReplicas = errors.New("invalid replica count")

// resourceScreen is the main resource table view.
type resourceScreen struct {
	client *k8s.Client
	theme  Theme

	res   *k8s.Resource
	ns    string
	allNs bool

	rows     []k8s.Row
	filtered []k8s.Row
	loading  bool
	err      error

	table    table.Model
	spinner  spinner.Model
	filter   textinput.Model
	filterOn bool
}

func newResourceScreen(client *k8s.Client, theme Theme, res *k8s.Resource, ns string, allNs bool) *resourceScreen {
	s := &resourceScreen{
		client: client,
		theme:  theme,
		res:    res,
		ns:     ns,
		allNs:  allNs,
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(theme.Accent)),
		),
		filter: textinput.New(),
	}
	s.filter.Placeholder = "filter…"
	s.filter.Prompt = "🔍 "
	s.filter.SetWidth(30)
	s.filter.SetStyles(s.filter.Styles())
	s.table = table.New()
	s.table.SetStyles(table.Styles{
		Header: lipgloss.NewStyle().Bold(true).Foreground(theme.Primary).Padding(0, 1),
		Cell:   lipgloss.NewStyle().Foreground(theme.Text).Padding(0, 1),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Accent),
	})
	s.table.KeyMap.GotoTop = key.NewBinding(key.WithDisabled())
	s.table.KeyMap.LineUp = key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up"))
	s.table.KeyMap.LineDown = key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down"))
	return s
}

// SetResource points the view at a resource type and namespace, then reloads.
func (s *resourceScreen) SetResource(res *k8s.Resource, ns string, allNs bool) tea.Cmd {
	s.res = res
	s.ns = ns
	s.allNs = allNs
	s.rows = nil
	s.filtered = nil
	return s.startLoad()
}

// SetNamespace changes the namespace and reloads.
func (s *resourceScreen) SetNamespace(ns string, allNs bool) tea.Cmd {
	s.ns = ns
	s.allNs = allNs
	s.rows = nil
	s.filtered = nil
	return s.startLoad()
}

// Reload triggers a fresh list.
func (s *resourceScreen) Reload() tea.Cmd {
	return s.startLoad()
}

func (s *resourceScreen) startLoad() tea.Cmd {
	s.loading = true
	s.err = nil
	return tea.Batch(s.loadCmd(), func() tea.Msg { return s.spinner.Tick() })
}

func (s *resourceScreen) loadCmd() tea.Cmd {
	client := s.client
	res := s.res
	ns := s.ns
	return func() tea.Msg {
		rows, err := client.List(tuiCtx, *res, ns)
		if err != nil {
			return loadErrorMsg{err: err}
		}
		return loadedRowsMsg{res: res, rows: rows}
	}
}

func (s *resourceScreen) Init() tea.Cmd {
	return s.startLoad()
}

func (s *resourceScreen) Update(msg tea.Msg) (screenView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if s.filterOn {
			return s.handleFilterKey(msg)
		}
		return s.handleKey(msg)
	case tea.MouseWheelMsg:
		switch msg.Mouse().Button {
		case tea.MouseWheelUp:
			s.table.MoveUp(1)
		case tea.MouseWheelDown:
			s.table.MoveDown(1)
		}
		return s, nil
	case spinner.TickMsg:
		if s.loading {
			var cmd tea.Cmd
			s.spinner, cmd = s.spinner.Update(msg)
			return s, cmd
		}
		return s, nil
	case loadedRowsMsg:
		if msg.res != s.res {
			return s, nil // stale response for an old resource
		}
		s.rows = msg.rows
		s.loading = false
		s.err = nil
		s.applyFilter()
		return s, nil
	case loadErrorMsg:
		s.loading = false
		s.err = msg.err
		return s, nil
	}
	if s.filterOn {
		var cmd tea.Cmd
		s.filter, cmd = s.filter.Update(msg)
		s.applyFilter()
		return s, cmd
	}
	t, cmd := s.table.Update(msg)
	s.table = t
	return s, cmd
}

func (s *resourceScreen) handleFilterKey(msg tea.KeyPressMsg) (screenView, tea.Cmd) {
	switch msg.String() {
	case "esc":
		s.filterOn = false
		s.filter.Blur()
		s.filter.SetValue("")
		s.applyFilter()
		return s, nil
	case "enter":
		s.filterOn = false
		s.filter.Blur()
		return s, nil
	case "ctrl+c":
		return s, s.quit()
	}
	var cmd tea.Cmd
	s.filter, cmd = s.filter.Update(msg)
	s.applyFilter()
	return s, cmd
}

func (s *resourceScreen) handleKey(msg tea.KeyPressMsg) (screenView, tea.Cmd) {
	km := s.table.KeyMap
	switch msg.String() {
	case "ctrl+c":
		return s, s.quit()
	case "?":
		return s, cmdMsg(helpToggleMsg{})
	case "/":
		s.filterOn = true
		return s, s.filter.Focus()
	case "q":
		return s, s.quit()
	case "enter", "right", "y", "d":
		row, ok := s.Selected()
		if !ok {
			return s, nil
		}
		kind := "yaml"
		if msg.String() == "d" {
			kind = "describe"
		}
		return s, cmdMsg(openDetailMsg{res: s.res, obj: row.Obj, kind: kind})
	case "l":
		if !s.res.SupportsLogs() {
			return s, s.toast(ToastWarning, "logs", "logs are only available for pods")
		}
		row, ok := s.Selected()
		if !ok {
			return s, nil
		}
		return s, cmdMsg(openLogsMsg{res: s.res, obj: row.Obj})
	case "g":
		if !s.res.Namespaced {
			return s, nil
		}
		return s, cmdMsg(setNamespaceMsg{ns: "", all: !s.allNs})
	case "n":
		return s, cmdMsg(setScreenMsg{screen: screenPickerNamespaces})
	case "u":
		return s, cmdMsg(setScreenMsg{screen: screenPickerContexts})
	case "k":
		return s, cmdMsg(setScreenMsg{screen: screenPickerResources})
	case "r":
		return s, s.Reload()
	case "x", "R", "s":
		row, ok := s.Selected()
		if !ok {
			return s, nil
		}
		return s.actionDialog(msg.String(), row)
	case "t":
		if !s.res.SupportsTop() {
			return s, s.toast(ToastWarning, "top", "top is only available for pods and nodes")
		}
		return s, cmdMsg(openTopMsg{res: s.res, ns: s.ns, allNs: s.allNs})
	case "T":
		return s, cmdMsg(setScreenMsg{screen: screenPickerThemes})
	case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-", "=", ",":
		if res := k8s.FindResource(msg.String()[0]); res != nil {
			return s, cmdMsg(setResourceMsg{res: res})
		}
	}
	switch {
	case key.Matches(msg, km.LineUp):
		s.table.MoveUp(1)
	case key.Matches(msg, km.LineDown):
		s.table.MoveDown(1)
	case key.Matches(msg, km.PageUp):
		s.table.MoveUp(s.table.Height())
	case key.Matches(msg, km.PageDown):
		s.table.MoveDown(s.table.Height())
	case key.Matches(msg, km.GotoBottom):
		s.table.GotoBottom()
	}
	return s, nil
}

func (s *resourceScreen) actionDialog(key string, row k8s.Row) (screenView, tea.Cmd) {
	name, ns := rowName(row), rowNS(row, s.ns)
	switch key {
	case "x":
		return s, cmdMsg(openDialogMsg{req: dialogRequest{
			title:   "Delete",
			message: "Delete " + s.res.Title + " " + name + "?",
			action:  "DELETE",
			run: func(string) error {
				return s.client.Delete(tuiCtx, *s.res, ns, name)
			},
		}})
	case "R":
		if !s.res.SupportsRestart() {
			return s, s.toast(ToastWarning, "restart", "restart is not supported for "+s.res.Title)
		}
		return s, cmdMsg(openDialogMsg{req: dialogRequest{
			title:   "Restart",
			message: "Rollout restart " + s.res.Title + " " + name + "?",
			action:  "RESTART",
			run: func(string) error {
				return s.client.Restart(tuiCtx, *s.res, ns, name)
			},
		}})
	case "s":
		if !s.res.SupportsScale() {
			return s, s.toast(ToastWarning, "scale", "scaling is not supported for "+s.res.Title)
		}
		return s, cmdMsg(openDialogMsg{req: dialogRequest{
			title:       "Scale",
			message:     "Scale " + s.res.Title + " " + name + " to:",
			action:      "SCALE",
			input:       true,
			placeholder: "replicas",
			run: func(value string) error {
				replicas, err := parseReplicas(value)
				if err != nil {
					return err
				}
				return s.client.Scale(tuiCtx, *s.res, ns, name, replicas)
			},
		}})
	}
	return s, nil
}

// Selected returns the currently highlighted resource row.
func (s *resourceScreen) Selected() (k8s.Row, bool) {
	if len(s.filtered) == 0 {
		return k8s.Row{}, false
	}
	idx := s.table.Cursor()
	if idx < 0 || idx >= len(s.filtered) {
		return k8s.Row{}, false
	}
	return s.filtered[idx], true
}

func (s *resourceScreen) applyFilter() {
	query := strings.TrimSpace(s.filter.Value())
	all := s.rows
	if query == "" {
		s.filtered = all
	} else {
		sources := make([]string, len(all))
		for i, r := range all {
			sources[i] = strings.Join(s.displayCols(r), " ")
		}
		matches := fuzzy.Find(query, sources)
		s.filtered = make([]k8s.Row, 0, len(matches))
		for _, m := range matches {
			s.filtered = append(s.filtered, all[m.Index])
		}
	}
	s.table.SetRows(s.renderRows(s.filtered))
}

// searchCols returns the columns used for filtering, including the namespace
// when browsing all namespaces so entries are findable by namespace.
func (s *resourceScreen) searchCols(r k8s.Row) []string {
	return s.displayCols(r)
}

// displayColumns returns the table columns to render, prepending a NAMESPACE
// column when browsing all namespaces.
func (s *resourceScreen) displayColumns() []k8s.Column {
	cols := s.res.Columns
	if !s.allNs || !s.res.Namespaced {
		return cols
	}
	out := make([]k8s.Column, 0, len(cols)+1)
	out = append(out, k8s.Column{Name: "NAMESPACE"})
	out = append(out, cols...)
	return out
}

// displayCols returns the rendered cells for a row, prepending the object's
// namespace when browsing all namespaces.
func (s *resourceScreen) displayCols(r k8s.Row) []string {
	if s.allNs && s.res.Namespaced {
		return append([]string{r.Obj.GetNamespace()}, r.Cols...)
	}
	out := make([]string, len(r.Cols))
	copy(out, r.Cols)
	return out
}

func (s *resourceScreen) renderRows(rows []k8s.Row) []table.Row {
	out := make([]table.Row, 0, len(rows))
	statusCol := s.res.StatusCol
	if s.allNs && s.res.Namespaced && statusCol >= 0 {
		statusCol++ // NAMESPACE column shifts the status column right
	}
	for _, r := range rows {
		cols := s.displayCols(r)
		if sc := statusCol; sc >= 0 && sc < len(cols) {
			status := cols[sc]
			cols[sc] = lipgloss.NewStyle().Foreground(s.theme.StatusColor(status)).Render(status)
		}
		out = append(out, table.Row(cols))
	}
	return out
}

func (s *resourceScreen) Resize(width, height int) {
	s.setTableWidth(width)
	s.table.SetHeight(height)
	s.filter.SetWidth(max(20, width/4))
}

func (s *resourceScreen) setTableWidth(width int) {
	displayRows := make([][]string, len(s.filtered))
	for i, r := range s.filtered {
		displayRows[i] = s.displayCols(r)
	}
	s.table.SetColumns(fitColumns(s.displayColumns(), displayRows, width))
	s.table.SetWidth(width)
}

func (s *resourceScreen) View() string {
	if s.err != nil {
		return s.theme.Box(0, 0, "Error", s.theme.BadStyle().Render(s.err.Error()))
	}
	if s.loading && len(s.rows) == 0 {
		return s.theme.MutedText(s.spinner.View() + " loading " + s.res.Title + "…")
	}
	var b strings.Builder
	if s.filterOn {
		b.WriteString(s.theme.FilterStyle().Render(s.filter.View()))
		b.WriteString("\n")
	}
	b.WriteString(s.table.View())
	return b.String()
}

func (s *resourceScreen) toast(kind toastKind, title, body string) tea.Cmd {
	return cmdMsg(toastMsg{kind: kind, title: title, body: body})
}

func (s *resourceScreen) quit() tea.Cmd {
	return func() tea.Msg { return tea.Quit() }
}

func rowName(r k8s.Row) string {
	if len(r.Cols) == 0 {
		return "<unknown>"
	}
	return r.Cols[0]
}

func rowNS(r k8s.Row, fallback string) string {
	if ns := r.Obj.GetNamespace(); ns != "" {
		return ns
	}
	return fallback
}

func parseReplicas(s string) (int32, error) {
	if s == "" {
		return 0, errInvalidReplicas
	}
	var n int32
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errInvalidReplicas
		}
		if n > (1<<31-1)/10 {
			return 0, errInvalidReplicas
		}
		n = n*10 + int32(c-'0')
	}
	return n, nil
}

// fitColumns sizes table columns so they exactly fill the available width.
func fitColumns(cols []k8s.Column, rows [][]string, width int) []table.Column {
	n := len(cols)
	if n == 0 {
		return nil
	}
	mins := make([]int, n)
	for i, c := range cols {
		mins[i] = lipgloss.Width(c.Name) + 2
	}
	for _, r := range rows {
		for i := 0; i < n && i < len(r); i++ {
			if w := lipgloss.Width(r[i]) + 2; w > mins[i] {
				mins[i] = w
			}
		}
	}
	total := 0
	for _, m := range mins {
		total += m
	}
	if width <= 0 {
		width = total
	}
	colsOut := make([]table.Column, n)
	if total <= width {
		extra := (width - total) / n
		for i, m := range mins {
			colsOut[i] = table.Column{Title: cols[i].Name, Width: m + extra}
		}
		return colsOut
	}
	floor := 8
	base := width - n*floor
	if base < 0 {
		base = 0
	}
	scale := float64(0)
	if total > 0 {
		scale = float64(base) / float64(total)
	}
	alloc := 0
	for i, m := range mins {
		w := floor + int(float64(m)*scale)
		if w < 1 {
			w = 1
		}
		colsOut[i] = table.Column{Title: cols[i].Name, Width: w}
		alloc += w
	}
	if diff := width - alloc; diff != 0 {
		colsOut[n-1].Width += diff
		if colsOut[n-1].Width < 1 {
			colsOut[n-1].Width = 1
		}
	}
	return colsOut
}
