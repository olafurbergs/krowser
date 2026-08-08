package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

// dialogRequest describes an action awaiting confirmation.
type dialogRequest struct {
	title       string
	message     string
	action      string
	input       bool
	placeholder string
	run         func(value string) error
}

// dialog is a modal dialog rendered as an overlay, optionally collecting
// a single text input value before running the action.
type dialog struct {
	active bool
	req    dialogRequest
	input  textinput.Model
}

func newDialog() *dialog {
	return &dialog{input: textinput.New()}
}

// show activates the dialog with the given request.
func (d *dialog) show(req dialogRequest) {
	d.active = true
	d.req = req
	if req.input {
		d.input = textinput.New()
		d.input.Placeholder = req.placeholder
		d.input.Prompt = "➤ "
		d.input.SetWidth(20)
		d.input.Focus()
	} else {
		d.input.Blur()
	}
}

// hide deactivates the dialog.
func (d *dialog) hide() {
	d.active = false
	d.req = dialogRequest{}
	d.input.Blur()
}

// View renders a centered modal over a dimmed backdrop.
func (d *dialog) View(width, height int, theme Theme) string {
	if !d.active {
		return ""
	}
	req := d.req
	var b strings.Builder
	b.WriteString(theme.Title(req.title))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(theme.HeaderFG).Render(req.message))
	b.WriteString("\n")
	if req.input {
		b.WriteString(d.input.View())
		b.WriteString("\n")
	}
	confirm := lipgloss.NewStyle().Foreground(theme.Good).Bold(true).Render("[ " + req.action + " ]")
	cancel := lipgloss.NewStyle().Foreground(theme.Muted).Render("[ cancel ]")
	b.WriteString(confirm + "   " + cancel)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Accent).
		Padding(1, 2).
		Width(min(56, max(24, width-8))).
		Render(b.String())

	backdrop := lipgloss.NewStyle().Background(theme.Dimmed)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceStyle(backdrop))
}
