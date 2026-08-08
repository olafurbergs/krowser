package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"
)

// helpOverlay renders the full keybinding help as a centered overlay.
func helpOverlay(km KeyMap, theme Theme, width, height int) string {
	h := help.New()
	h.SetWidth(width - 6)
	full := h.FullHelpView(km.FullHelp())
	body := theme.Title("Help — key bindings") + "\n\n" + full
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Padding(1, 2).
		Width(min(width-4, 90)).
		Render(body)
	backdrop := lipgloss.NewStyle().Background(theme.Dimmed)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceStyle(backdrop))
}
