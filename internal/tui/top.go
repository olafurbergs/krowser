package tui

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/olafurb/krowser/internal/k8s"
)

// topScreen shows live CPU/memory usage for pods or nodes as progress bars
// against their limits and requests, refreshing every second.
type topScreen struct {
	client *k8s.Client
	theme  Theme

	res   *k8s.Resource
	ns    string
	allNs bool

	entries []k8s.TopEntry
	loading bool
	err     error

	table   table.Model
	width   int
	height  int
	spinner spinner.Model
	palette []color.Color
}

func newTopScreen(client *k8s.Client, theme Theme, res *k8s.Resource, ns string, allNs bool) *topScreen {
	t := table.New()
	t.KeyMap.GotoTop = key.NewBinding(key.WithDisabled())
	t.KeyMap.LineUp = key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up"))
	t.KeyMap.LineDown = key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down"))
	t.SetStyles(table.Styles{
		Header: lipgloss.NewStyle().Bold(true).Foreground(theme.Primary).Padding(0, 1),
		Cell:   lipgloss.NewStyle().Foreground(theme.Text).Padding(0, 1),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Accent),
	})
	s := &topScreen{
		client: client,
		theme:  theme,
		res:    res,
		ns:     ns,
		allNs:  allNs,
		table:  t,
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(theme.Accent)),
		),
	}
	s.updatePalette()
	return s
}

// updatePalette rebuilds the green-to-red gradient used for the gauges from
// the current theme colors.
func (s *topScreen) updatePalette() {
	s.palette = lipgloss.Blend1D(12, s.theme.Good, s.theme.Warn, s.theme.Bad)
}

func (s *topScreen) Init() tea.Cmd {
	s.loading = true
	return tea.Batch(s.loadCmd(), s.tickCmd(), func() tea.Msg { return s.spinner.Tick() })
}

func (s *topScreen) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return topTickMsg{} })
}

func (s *topScreen) loadCmd() tea.Cmd {
	client, res, ns := s.client, s.res, s.ns
	return func() tea.Msg {
		entries, err := client.Top(tuiCtx, *res, ns)
		if err != nil {
			return loadErrorMsg{err: err}
		}
		return loadedTopMsg{entries: entries}
	}
}

func (s *topScreen) Update(msg tea.Msg) (screenView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return s.handleKey(msg)
	case tea.MouseWheelMsg:
		switch msg.Mouse().Button {
		case tea.MouseWheelUp:
			s.table.MoveUp(1)
		case tea.MouseWheelDown:
			s.table.MoveDown(1)
		}
		return s, nil
	case topTickMsg:
		cmd := s.tickCmd()
		if !s.loading {
			s.loading = true
			cmd = tea.Batch(cmd, s.loadCmd())
		}
		return s, cmd
	case loadedTopMsg:
		s.entries = msg.entries
		s.loading = false
		s.err = nil
		s.applyTable()
		return s, nil
	case loadErrorMsg:
		s.loading = false
		s.err = msg.err
		return s, nil
	case spinner.TickMsg:
		if s.loading {
			var cmd tea.Cmd
			s.spinner, cmd = s.spinner.Update(msg)
			return s, cmd
		}
		return s, nil
	}
	t, cmd := s.table.Update(msg)
	s.table = t
	return s, cmd
}

func (s *topScreen) handleKey(msg tea.KeyPressMsg) (screenView, tea.Cmd) {
	km := s.table.KeyMap
	switch msg.String() {
	case "ctrl+c":
		return s, quitCmd()
	case "q", "esc", "left":
		return s, cmdMsg(backMsg{})
	case "?":
		return s, cmdMsg(helpToggleMsg{})
	case "r":
		s.loading = true
		return s, s.loadCmd()
	case "enter", "right":
		if len(s.entries) == 0 || s.table.Cursor() < 0 {
			return s, nil
		}
		return s, s.openDetail()
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

func (s *topScreen) Resize(width, height int) {
	s.width = width
	s.height = height
	s.table.SetWidth(width)
	s.table.SetHeight(max(1, height-1))
	s.applyTable()
}

// applyTable rebuilds the table rows and sizes each column to fit its longest
// string while filling the available width.
func (s *topScreen) applyTable() {
	cols := []k8s.Column{{Name: "NAME"}}
	if s.allNs {
		cols = append(cols, k8s.Column{Name: "NAMESPACE"})
	}
	cols = append(cols, k8s.Column{Name: "CPU"}, k8s.Column{Name: "MEM"})

	cells := make([][]string, 0, len(s.entries))
	rows := make([]table.Row, 0, len(s.entries))
	for _, e := range s.entries {
		row := []string{e.Name}
		if s.allNs {
			row = append(row, e.Namespace)
		}
		row = append(row, s.cell(e, true), s.cell(e, false))
		cells = append(cells, row)
		rows = append(rows, table.Row(row))
	}

	s.table.SetColumns(fitColumns(cols, cells, s.width))
	s.table.SetWidth(s.width)
	s.table.SetRows(rows)
}

func (s *topScreen) View() string {
	if s.err != nil {
		return s.theme.Box(0, 0, "Error", s.theme.BadStyle().Render(s.err.Error()))
	}
	if s.loading && len(s.entries) == 0 {
		return s.theme.MutedText(s.spinner.View() + " loading " + s.res.Title + " metrics…")
	}

	title := s.res.Title
	if s.res.Namespaced {
		if s.allNs {
			title += " · all-namespaces"
		} else {
			title += " · " + s.ns
		}
	}
	return s.theme.Title(title) +
		"  " + s.theme.MutedText("auto-refresh 1s · r refresh") +
		"\n" + s.table.View()
}

// cell renders the gauge and label for one metric (cpu=true) of an entry.
func (s *topScreen) cell(e k8s.TopEntry, cpu bool) string {
	var used, lim, req float64
	if cpu {
		used, lim, req = e.CPU, e.CPULim, e.CPUReq
	} else {
		used, lim, req = e.Mem, e.MemLim, e.MemReq
	}
	format := formatBytes
	if cpu {
		format = formatCPU
	}

	frac := 0.0
	if lim > 0 {
		frac = used / lim
	} else if req > 0 {
		frac = used / req
	}

	label := ""
	if s.res.Plural == "nodes" {
		pct := 0.0
		if lim > 0 {
			pct = used / lim * 100
		}
		label = fmt.Sprintf("%s/%s %d%%", format(used), format(lim), int(math.Round(pct)))
	} else {
		label = format(used) + "/"
		if lim > 0 {
			label += format(lim)
		} else {
			label += "–"
		}
		if req > 0 {
			label += " rq" + format(req)
		}
	}
	return s.gauge(frac) + " " + label
}

// openDetail fetches the selected pod or node and opens its detail view.
func (s *topScreen) openDetail() tea.Cmd {
	e := s.entries[s.table.Cursor()]
	res := *s.res
	return func() tea.Msg {
		obj, err := s.client.Get(tuiCtx, res, e.Namespace, e.Name)
		if err != nil {
			return toastMsg{kind: ToastError, title: "open", body: err.Error()}
		}
		return openDetailMsg{res: &res, obj: obj, kind: "yaml"}
	}
}

// gauge renders a progress bar whose filled blocks are individually colored
// along a green-to-red gradient, turning redder as usage approaches and
// exceeds the limit.
func (s *topScreen) gauge(frac float64) string {
	if len(s.palette) != 12 {
		s.updatePalette()
	}
	filled := int(math.Round(frac * 12))
	if filled < 0 {
		filled = 0
	}
	if filled > 12 {
		filled = 12
	}
	var b strings.Builder
	for i := 0; i < 12; i++ {
		if i < filled {
			b.WriteString(lipgloss.NewStyle().Foreground(s.palette[i]).Render("█"))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(s.theme.Dimmed).Render("░"))
		}
	}
	return b.String()
}

func formatCPU(m float64) string {
	if m >= 1000 {
		return trimFrac(m / 1000)
	}
	return trimFrac(m) + "m"
}

func formatBytes(b float64) string {
	units := []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei"}
	i := 0
	v := b
	for v >= 1024 && i < len(units)-1 {
		v /= 1024
		i++
	}
	return trimFrac(v) + units[i]
}

// trimFrac formats v with at most two decimals, stripping trailing zeros.
func trimFrac(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}
