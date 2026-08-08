package tui

import (
	"bufio"
	"io"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/olafurb/krowser/internal/k8s"
	"github.com/olafurb/krowser/internal/tui/loghl"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const maxLogLines = 2000

// logsScreen streams and displays pod container logs.
type logsScreen struct {
	client *k8s.Client
	theme  Theme

	res        *k8s.Resource
	obj        *unstructured.Unstructured
	container  string
	containers []string
	follow     bool
	showTS     bool
	tail       int

	lines     []string
	hlLines   []string
	hl        *loghl.Highlighter
	vp        viewport.Model
	reader    *bufio.Reader
	stream    io.ReadCloser
	streaming bool
	err       error
	stopped   bool
}

func newLogsScreen(client *k8s.Client, theme Theme, res *k8s.Resource, obj *unstructured.Unstructured) *logsScreen {
	return &logsScreen{
		client: client,
		theme:  theme,
		res:    res,
		obj:    obj,
		follow: true,
		tail:   200,
		hl:     logHighlighter(theme),
		vp: viewport.New(
			viewport.WithWidth(80),
			viewport.WithHeight(20),
		),
	}
}

// logHighlighter builds a syntax highlighter colored by the theme.
func logHighlighter(th Theme) *loghl.Highlighter {
	return loghl.NewHighlighter(loghl.Palette{
		Primary: themeHex(th.Primary),
		Accent:  themeHex(th.Accent),
		Good:    themeHex(th.Good),
		Warn:    themeHex(th.Warn),
		Bad:     themeHex(th.Bad),
		Muted:   themeHex(th.Muted),
	})
}

// Init loads the container list then starts the stream.
func (s *logsScreen) Init() tea.Cmd {
	client := s.client
	ns := s.obj.GetNamespace()
	name := s.obj.GetName()
	return func() tea.Msg {
		containers, err := client.PodContainers(tuiCtx, ns, name)
		if err != nil {
			return logStreamEndMsg{err: err}
		}
		return loadedContainersMsg{containers: containers}
	}
}

// SetContainer switches the container being tailed.
func (s *logsScreen) SetContainer(name string) tea.Cmd {
	if name == s.container {
		return nil
	}
	s.container = name
	s.lines = nil
	s.hlLines = nil
	s.err = nil
	return s.startStream()
}

func (s *logsScreen) startStream() tea.Cmd {
	ns := s.obj.GetNamespace()
	name := s.obj.GetName()
	rc, err := s.client.OpenLogStream(tuiCtx, ns, name, s.container, s.follow, s.tail, true)
	if err != nil {
		return func() tea.Msg { return logStreamEndMsg{err: err} }
	}
	s.stream = rc
	s.reader = bufio.NewReader(rc)
	s.streaming = true
	s.stopped = false
	return s.readLine()
}

func (s *logsScreen) readLine() tea.Cmd {
	r := s.reader
	return func() tea.Msg {
		line, err := r.ReadString('\n')
		if err != nil {
			return logStreamEndMsg{err: err}
		}
		return logLineMsg{line: strings.TrimRight(line, "\r\n")}
	}
}

func (s *logsScreen) Update(msg tea.Msg) (screenView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "left":
			s.closeStream()
			return s, cmdMsg(backMsg{})
		case "ctrl+c":
			s.closeStream()
			return s, quitCmd()
		case "f":
			s.follow = !s.follow
			return s, s.startStream()
		case "t":
			s.showTS = !s.showTS
			s.rehighlight()
			s.renderContent()
			return s, nil
		case "c":
			s.closeStream()
			return s, cmdMsg(setScreenMsg{screen: screenPickerContainers})
		}
	case loadedContainersMsg:
		s.containers = msg.containers
		if len(msg.containers) > 0 {
			s.container = msg.containers[len(msg.containers)-1] // last = app container (init containers first)
		}
		return s, s.startStream()
	case logLineMsg:
		line := msg.line
		if !s.showTS {
			line = stripTimestamp(line)
		}
		s.lines = append(s.lines, line)
		s.hlLines = append(s.hlLines, s.hl.Highlight(line))
		if len(s.lines) > maxLogLines {
			s.lines = s.lines[len(s.lines)-maxLogLines:]
			s.hlLines = s.hlLines[len(s.hlLines)-maxLogLines:]
		}
		s.renderContent()
		if s.follow {
			s.vp.GotoBottom()
		}
		return s, s.readLine()
	case logStreamEndMsg:
		s.streaming = false
		if s.stream != nil {
			s.stream.Close()
			s.stream = nil
		}
		if msg.err != nil && !isEOF(msg.err) {
			s.err = msg.err
		}
		if s.follow && !s.stopped {
			return s, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return logReconnectMsg{} })
		}
		return s, nil
	case logReconnectMsg:
		if s.follow && !s.stopped {
			return s, s.startStream()
		}
		return s, nil
	}
	vp, cmd := s.vp.Update(msg)
	s.vp = vp
	return s, cmd
}

func (s *logsScreen) closeStream() {
	s.stopped = true
	if s.stream != nil {
		s.stream.Close()
		s.stream = nil
	}
}

func (s *logsScreen) renderContent() {
	s.vp.SetContent(strings.Join(s.hlLines, "\n"))
}

// rehighlight rebuilds the highlighted buffer from the raw lines, for example
// when timestamps are toggled.
func (s *logsScreen) rehighlight() {
	s.hlLines = s.hlLines[:0]
	for _, line := range s.lines {
		if !s.showTS {
			line = stripTimestamp(line)
		}
		s.hlLines = append(s.hlLines, s.hl.Highlight(line))
	}
}

func (s *logsScreen) Resize(width, height int) {
	s.vp.SetWidth(width)
	s.vp.SetHeight(height)
}

func (s *logsScreen) View() string {
	var b strings.Builder
	container := s.container
	if container == "" {
		container = "(all)"
	}
	status := "⏹ paused"
	if s.follow {
		status = "▶ following"
	}
	header := lipgloss.NewStyle().Bold(true).Foreground(s.theme.Primary).
		Render("Logs · "+s.obj.GetName()+" · "+container) +
		"  " + lipgloss.NewStyle().Foreground(s.theme.Accent).Render(status)
	b.WriteString(header)
	b.WriteString("\n")
	if s.err != nil {
		b.WriteString(s.theme.BadStyle().Render(s.err.Error()))
		b.WriteString("\n")
	}
	b.WriteString(s.vp.View())
	return b.String()
}

// stripTimestamp removes the RFC3339 prefix that the API adds when
// timestamps are requested.
func stripTimestamp(line string) string {
	// timestamp form: "2006-01-02T15:04:05.000000000Z "
	if len(line) < 20 {
		return line
	}
	if line[4] == '-' && line[7] == '-' && line[10] == 'T' {
		if idx := strings.IndexByte(line, ' '); idx >= 0 {
			return line[idx+1:]
		}
	}
	return line
}

func isEOF(err error) bool {
	return err == io.EOF
}
