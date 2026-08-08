package tui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const toastLifetime = 6 * time.Second

// toast is a single notification entry.
type toast struct {
	id      int
	kind    toastKind
	title   string
	body    string
	created time.Time
}

// toastManager holds the stack of visible toasts.
type toastManager struct {
	toasts []toast
	nextID int
	theme  Theme
	width  int
}

func newToastManager(theme Theme) *toastManager {
	return &toastManager{theme: theme}
}

// push adds a toast and returns a command that schedules its dismissal.
func (t *toastManager) push(kind toastKind, title, body string) tea.Cmd {
	t.toasts = append(t.toasts, toast{
		id:      t.nextID,
		kind:    kind,
		title:   title,
		body:    body,
		created: time.Now(),
	})
	t.nextID++
	return tea.Tick(toastLifetime+500*time.Millisecond, func(time.Time) tea.Msg {
		return toastDismissMsg{}
	})
}

// dismiss removes the oldest toast.
func (t *toastManager) dismiss() {
	if len(t.toasts) > 0 {
		t.toasts = t.toasts[1:]
	}
}

// height returns the number of lines the toast stack occupies.
func (t *toastManager) height() int {
	return len(t.toasts)
}

func (t *toastManager) style(k toastKind) lipgloss.Style {
	color := t.theme.Accent
	switch k {
	case ToastSuccess:
		color = t.theme.Good
	case ToastWarning:
		color = t.theme.Warn
	case ToastError:
		color = t.theme.Bad
	}
	return lipgloss.NewStyle().
		Foreground(color).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 1)
}

// View renders the toast stack right-aligned, one line per toast.
func (t *toastManager) View() string {
	if len(t.toasts) == 0 {
		return ""
	}
	var lines []string
	for _, tn := range t.toasts {
		prefix := "●"
		label := tn.title
		if tn.body != "" {
			label += " " + tn.body
		}
		line := t.style(tn.kind).Render(prefix + " " + label)
		if t.width > 0 {
			line = lipgloss.PlaceHorizontal(t.width, lipgloss.Right, line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
