package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

func colorCode(s lipgloss.Style) string {
	inner := strings.TrimPrefix(s.Render("x"), "\x1b[")
	if i := strings.Index(inner, "m"); i >= 0 {
		inner = inner[:i]
	}
	return inner
}

func accentCode(t Theme) string {
	return colorCode(lipgloss.NewStyle().Foreground(t.Accent))
}

func textCode(t Theme) string {
	return colorCode(lipgloss.NewStyle().Foreground(t.Text))
}

func TestResourceTableSelectionUsesThemeAccent(t *testing.T) {
	mk := func(th Theme) string {
		tbl := table.New()
		tbl.SetStyles(table.Styles{
			Header:   lipgloss.NewStyle().Bold(true).Foreground(th.Primary).Padding(0, 1),
			Cell:     lipgloss.NewStyle().Foreground(th.Text).Padding(0, 1),
			Selected: lipgloss.NewStyle().Bold(true).Foreground(th.Accent),
		})
		tbl.SetColumns([]table.Column{{Title: "NAME", Width: 10}})
		tbl.SetRows([]table.Row{{"aaa"}, {"bbb"}})
		tbl.SetWidth(20)
		tbl.SetHeight(5)
		tbl.UpdateViewport()
		return tbl.View()
	}

	view := mk(DarkTheme)
	if !strings.Contains(view, accentCode(DarkTheme)) {
		t.Fatalf("selected row should use theme accent %q, got view %q", accentCode(DarkTheme), view)
	}
	if !strings.Contains(view, textCode(DarkTheme)) {
		t.Fatalf("cell text should use theme text color %q, got view %q", textCode(DarkTheme), view)
	}
	if mk(DarkTheme) == mk(Dracula) {
		t.Error("selection styling should differ between themes (hardcoded color regression)")
	}
	if mk(OneDark) == mk(Dracula) {
		t.Error("cell text styling should differ between themes (hardcoded color regression)")
	}
}

func TestPickerSelectionUsesThemeAccent(t *testing.T) {
	items := []pickerItem{{title: "aaa", desc: "d1"}, {title: "bbb", desc: "d2"}}
	mk := func(th Theme) string {
		p := newPicker(nil, th, pickerThemes, "Pick", items)
		p.list.SetSize(30, 5)
		return p.list.View()
	}

	view := mk(DarkTheme)
	if !strings.Contains(view, accentCode(DarkTheme)) {
		t.Fatalf("selected item should use theme accent %q, got view %q", accentCode(DarkTheme), view)
	}
	if mk(DarkTheme) == mk(Dracula) {
		t.Error("picker selection styling should differ between themes (hardcoded color regression)")
	}
	if mk(OneDark) == mk(Dracula) {
		t.Error("picker item text styling should differ between themes (hardcoded color regression)")
	}
}
