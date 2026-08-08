// Package tui implements the Bubble Tea application for krowser.
package tui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Theme holds the palette and shared styles for the application. Each theme
// is a named, selectable palette.
type Theme struct {
	Name string
	Dark bool

	Primary    color.Color
	Accent     color.Color
	Good       color.Color
	Warn       color.Color
	Bad        color.Color
	Text       color.Color
	Muted      color.Color
	Dimmed     color.Color
	SelectedBG color.Color
	HeaderFG   color.Color
	HeaderBG   color.Color
}

// gradientColors returns the two stops used for brand gradients.
func (t Theme) gradientColors() []color.Color {
	return []color.Color{t.Primary, t.Accent}
}

// Header renders a gradient-branded title bar.
func (t Theme) Header(title string) string {
	return t.gradientText(title, lipgloss.NewStyle().Bold(true))
}

// gradientText renders each rune with a color sampled from the theme gradient.
func (t Theme) gradientText(s string, base lipgloss.Style) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return ""
	}
	colors := lipgloss.Blend1D(len(runes), t.gradientColors()...)
	var b strings.Builder
	for i, r := range runes {
		c := colors[i]
		if c == nil {
			c = t.Primary
		}
		b.WriteString(base.Foreground(c).Render(string(r)))
	}
	return b.String()
}

// themeHex renders a color as an "#rrggbb" string.
func themeHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

// Title returns a style for section titles.
func (t Theme) Title(s string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(t.Primary).Render(s)
}

// AccentText styles a string with the accent color.
func (t Theme) AccentText(s string) string {
	return lipgloss.NewStyle().Foreground(t.Accent).Render(s)
}

// MutedText styles a string with the muted color.
func (t Theme) MutedText(s string) string {
	return lipgloss.NewStyle().Foreground(t.Muted).Render(s)
}

// StatusColor maps a Kubernetes status string to a color.
func (t Theme) StatusColor(status string) color.Color {
	switch strings.ToLower(status) {
	case "running", "ready", "succeeded", "active", "true", "available", "bound":
		return t.Good
	case "pending", "containercreating", "terminating", "init:0/1", "waiting", "unknown", "unavailable":
		return t.Warn
	case "error", "failed", "crashloopbackoff", "errorbackoff", "notready", "terminated", "evicted", "false":
		return t.Bad
	default:
		return t.Muted
	}
}

// Box renders a bordered panel with an optional title.
func (t Theme) Box(width, height int, title, content string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Dimmed).
		Width(width).
		Height(height).
		Render(t.Title(title) + "\n" + content)
}

// BadStyle styles an error message.
func (t Theme) BadStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Bad)
}

// FilterStyle styles the filter input bar.
func (t Theme) FilterStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Padding(0, 1)
}
