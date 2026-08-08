package loghl

import (
	"strings"
	"testing"
)

func testPalette() Palette {
	return Palette{
		Primary: "#ae81ff",
		Accent:  "#66d9ef",
		Good:    "#a6e22e",
		Warn:    "#e6db74",
		Bad:     "#f92672",
		Muted:   "#75715e",
	}
}

func TestHighlighterColorsLevelsAndDates(t *testing.T) {
	h := NewHighlighter(testPalette())
	out := h.Highlight("2024-01-02T03:04:05Z ERROR something")
	if !strings.Contains(out, "38;2;249;38;114") {
		t.Errorf("ERROR should use Bad color, got %q", out)
	}
	if !strings.Contains(out, "38;2;174;129;255") {
		t.Errorf("timestamp should use Primary color, got %q", out)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Error("expected ANSI escapes")
	}
}

func TestHighlighterLevelColors(t *testing.T) {
	h := NewHighlighter(testPalette())
	cases := map[string]string{
		"INFO":  "38;2;102;217;239", // Accent
		"WARN":  "38;2;230;219;116", // Warn
		"ERROR": "38;2;249;38;114",  // Bad
	}
	for in, want := range cases {
		out := h.Highlight(in)
		if !strings.Contains(out, want) {
			t.Errorf("highlight(%q) should contain %q, got %q", in, want, out)
		}
	}
}

func TestHighlighterColorsFollowPalette(t *testing.T) {
	a := NewHighlighter(testPalette())
	b := NewHighlighter(Palette{
		Primary: "#c678dd",
		Accent:  "#56b6c2",
		Good:    "#98c379",
		Warn:    "#e5c07b",
		Bad:     "#e06c75",
		Muted:   "#5c6370",
	})
	if a.Highlight("ERROR") == b.Highlight("ERROR") {
		t.Error("level color should differ between palettes")
	}
	if !strings.Contains(b.Highlight("ERROR"), "38;2;224;108;117") {
		t.Errorf("ERROR should use the second palette's Bad color, got %q", b.Highlight("ERROR"))
	}
}
