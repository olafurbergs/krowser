// Package loghl brings syntax highlighting to pod logs.
//
// The highlighting engine in this package is adapted from loglit
// (https://github.com/madmaxieee/loglit, MIT license, Copyright (c) 2025
// Tsng, Kahiok). The engine has been vendored into krowser because loglit
// keeps it in Go `internal/` packages, which cannot be imported from another
// module. See the LICENSE file in this directory.
package loghl

// Palette carries the base colors used to build a highlighter theme. Hex
// strings are expected in "#rrggbb" form.
type Palette struct {
	Primary, Accent string
	Good, Warn, Bad string
	Muted           string
}

// Highlighter highlights individual log lines.
type Highlighter struct {
	r *Renderer
}

// NewHighlighter builds a highlighter whose colors are derived from Palette.
func NewHighlighter(p Palette) *Highlighter {
	r, err := New(GetDefaultConfig(), themeFromPalette(p))
	if err != nil {
		return &Highlighter{}
	}
	return &Highlighter{r: r}
}

// Highlight returns the line with syntax highlighted via ANSI escapes. It
// returns the input unchanged when highlighting fails or is unavailable.
func (h *Highlighter) Highlight(line string) string {
	if h.r == nil {
		return line
	}
	out, err := h.r.Render(line)
	if err != nil {
		return line
	}
	return out
}

// themeFromPalette maps the base syntax groups to the palette colors. The
// built-in config links every log group to these base groups, so overriding
// just the base groups recolors the whole log output.
func themeFromPalette(p Palette) Theme {
	th := Theme{
		Name:         "krowser",
		HighlightMap: map[string]*Highlight{},
	}
	set := func(group, hex string, opts ...func(*Highlight)) {
		hl := &Highlight{Group: group}
		if hex != "" {
			hl.Fg = fg(hex)
		}
		for _, o := range opts {
			o(hl)
		}
		th.HighlightMap[group] = hl
	}
	set("Constant", p.Accent)
	set("Number", p.Accent)
	set("Float", p.Accent)
	set("Special", p.Primary)
	set("Comment", p.Muted, func(h *Highlight) { h.Italic = true })
	set("Boolean", p.Good)
	set("String", p.Good)
	set("Type", p.Primary)
	set("Operator", p.Accent)
	set("Statement", p.Primary)
	set("Function", p.Accent)
	set("Underlined", p.Accent, func(h *Highlight) { h.Underline = true })
	set("Label", p.Primary)
	set("Structure", p.Primary)
	set("ErrorMsg", p.Bad)
	set("WarningMsg", p.Warn)
	set("Exception", p.Warn)
	set("Debug", p.Muted)
	set("LogGreen", p.Good)
	set("LogBlue", p.Accent)
	return th
}
