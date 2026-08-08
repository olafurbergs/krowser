package tui

import "charm.land/lipgloss/v2"

// Monokai is the default theme.
var Monokai = Theme{
	Name:       "Monokai",
	Dark:       true,
	Primary:    lipgloss.Color("#ae81ff"),
	Accent:     lipgloss.Color("#66d9ef"),
	Good:       lipgloss.Color("#a6e22e"),
	Warn:       lipgloss.Color("#e6db74"),
	Bad:        lipgloss.Color("#f92672"),
	Text:       lipgloss.Color("#f8f8f2"),
	Muted:      lipgloss.Color("#75715e"),
	Dimmed:     lipgloss.Color("#3e3d32"),
	SelectedBG: lipgloss.Color("#49483e"),
	HeaderFG:   lipgloss.Color("#f8f8f2"),
	HeaderBG:   lipgloss.Color("#272822"),
}

// OneDark is a popular dark editor theme.
var OneDark = Theme{
	Name:       "One Dark",
	Dark:       true,
	Primary:    lipgloss.Color("#c678dd"),
	Accent:     lipgloss.Color("#56b6c2"),
	Good:       lipgloss.Color("#98c379"),
	Warn:       lipgloss.Color("#e5c07b"),
	Bad:        lipgloss.Color("#e06c75"),
	Text:       lipgloss.Color("#abb2bf"),
	Muted:      lipgloss.Color("#5c6370"),
	Dimmed:     lipgloss.Color("#3b4048"),
	SelectedBG: lipgloss.Color("#2a2e35"),
	HeaderFG:   lipgloss.Color("#abb2bf"),
	HeaderBG:   lipgloss.Color("#282c34"),
}

// Dracula is the popular dark Dracula theme.
var Dracula = Theme{
	Name:       "Dracula",
	Dark:       true,
	Primary:    lipgloss.Color("#bd93f9"),
	Accent:     lipgloss.Color("#8be9fd"),
	Good:       lipgloss.Color("#50fa7b"),
	Warn:       lipgloss.Color("#f1fa8c"),
	Bad:        lipgloss.Color("#ff5555"),
	Text:       lipgloss.Color("#f8f8f2"),
	Muted:      lipgloss.Color("#6272a4"),
	Dimmed:     lipgloss.Color("#44475a"),
	SelectedBG: lipgloss.Color("#44475a"),
	HeaderFG:   lipgloss.Color("#f8f8f2"),
	HeaderBG:   lipgloss.Color("#282a36"),
}

// Nord is the arctic, north-bluish theme.
var Nord = Theme{
	Name:       "Nord",
	Dark:       true,
	Primary:    lipgloss.Color("#88c0d0"),
	Accent:     lipgloss.Color("#81a1c1"),
	Good:       lipgloss.Color("#a3be8c"),
	Warn:       lipgloss.Color("#ebcb8b"),
	Bad:        lipgloss.Color("#bf616a"),
	Text:       lipgloss.Color("#d8dee9"),
	Muted:      lipgloss.Color("#616e88"),
	Dimmed:     lipgloss.Color("#3b4252"),
	SelectedBG: lipgloss.Color("#3b4252"),
	HeaderFG:   lipgloss.Color("#d8dee9"),
	HeaderBG:   lipgloss.Color("#2e3440"),
}

// GruvboxDark is the earthy retro-groove theme.
var GruvboxDark = Theme{
	Name:       "Gruvbox Dark",
	Dark:       true,
	Primary:    lipgloss.Color("#d3869b"),
	Accent:     lipgloss.Color("#83a598"),
	Good:       lipgloss.Color("#b8bb26"),
	Warn:       lipgloss.Color("#fabd2f"),
	Bad:        lipgloss.Color("#fb4934"),
	Text:       lipgloss.Color("#ebdbb2"),
	Muted:      lipgloss.Color("#928374"),
	Dimmed:     lipgloss.Color("#3c3836"),
	SelectedBG: lipgloss.Color("#3c3836"),
	HeaderFG:   lipgloss.Color("#ebdbb2"),
	HeaderBG:   lipgloss.Color("#282828"),
}

// SolarizedDark is the low-contrast Solarized dark theme.
var SolarizedDark = Theme{
	Name:       "Solarized Dark",
	Dark:       true,
	Primary:    lipgloss.Color("#d33682"),
	Accent:     lipgloss.Color("#2aa198"),
	Good:       lipgloss.Color("#859900"),
	Warn:       lipgloss.Color("#b58900"),
	Bad:        lipgloss.Color("#dc322f"),
	Text:       lipgloss.Color("#839496"),
	Muted:      lipgloss.Color("#657b83"),
	Dimmed:     lipgloss.Color("#073642"),
	SelectedBG: lipgloss.Color("#073642"),
	HeaderFG:   lipgloss.Color("#93a1a1"),
	HeaderBG:   lipgloss.Color("#002b36"),
}

// CatppuccinMocha is the popular pastel Mocha flavor.
var CatppuccinMocha = Theme{
	Name:       "Catppuccin Mocha",
	Dark:       true,
	Primary:    lipgloss.Color("#cba6f7"),
	Accent:     lipgloss.Color("#89b4fa"),
	Good:       lipgloss.Color("#a6e3a1"),
	Warn:       lipgloss.Color("#f9e2af"),
	Bad:        lipgloss.Color("#f38ba8"),
	Text:       lipgloss.Color("#cdd6f4"),
	Muted:      lipgloss.Color("#a6adc8"),
	Dimmed:     lipgloss.Color("#313244"),
	SelectedBG: lipgloss.Color("#313244"),
	HeaderFG:   lipgloss.Color("#cdd6f4"),
	HeaderBG:   lipgloss.Color("#1e1e2e"),
}

// TokyoNight is the Tokyo Night dark theme.
var TokyoNight = Theme{
	Name:       "Tokyo Night",
	Dark:       true,
	Primary:    lipgloss.Color("#bb9af7"),
	Accent:     lipgloss.Color("#7dcfff"),
	Good:       lipgloss.Color("#9ece6a"),
	Warn:       lipgloss.Color("#e0af68"),
	Bad:        lipgloss.Color("#f7768e"),
	Text:       lipgloss.Color("#c0caf5"),
	Muted:      lipgloss.Color("#565f89"),
	Dimmed:     lipgloss.Color("#24283b"),
	SelectedBG: lipgloss.Color("#24283b"),
	HeaderFG:   lipgloss.Color("#c0caf5"),
	HeaderBG:   lipgloss.Color("#1a1b26"),
}

// SolarizedLight is the light Solarized variant.
var SolarizedLight = Theme{
	Name:       "Solarized Light",
	Dark:       false,
	Primary:    lipgloss.Color("#d33682"),
	Accent:     lipgloss.Color("#2aa198"),
	Good:       lipgloss.Color("#859900"),
	Warn:       lipgloss.Color("#b58900"),
	Bad:        lipgloss.Color("#dc322f"),
	Text:       lipgloss.Color("#657b83"),
	Muted:      lipgloss.Color("#93a1a1"),
	Dimmed:     lipgloss.Color("#e4e4e4"),
	SelectedBG: lipgloss.Color("#eee8d5"),
	HeaderFG:   lipgloss.Color("#657b83"),
	HeaderBG:   lipgloss.Color("#eee8d5"),
}

// CatppuccinLatte is the light pastel flavor.
var CatppuccinLatte = Theme{
	Name:       "Catppuccin Latte",
	Dark:       false,
	Primary:    lipgloss.Color("#8839ef"),
	Accent:     lipgloss.Color("#1e66f5"),
	Good:       lipgloss.Color("#40a02b"),
	Warn:       lipgloss.Color("#df8e1d"),
	Bad:        lipgloss.Color("#d20f39"),
	Text:       lipgloss.Color("#4c4f69"),
	Muted:      lipgloss.Color("#8c8fa1"),
	Dimmed:     lipgloss.Color("#e6e9ef"),
	SelectedBG: lipgloss.Color("#e6e9ef"),
	HeaderFG:   lipgloss.Color("#4c4f69"),
	HeaderBG:   lipgloss.Color("#eff1f5"),
}

// Themes is the list of selectable themes.
var Themes = []Theme{
	Monokai,
	OneDark,
	Dracula,
	Nord,
	GruvboxDark,
	SolarizedDark,
	CatppuccinMocha,
	TokyoNight,
	SolarizedLight,
	CatppuccinLatte,
}

// DefaultTheme is the theme used when none is selected.
var DefaultTheme = Monokai

// ThemeByName returns the theme registered under name.
func ThemeByName(name string) (Theme, bool) {
	for _, t := range Themes {
		if t.Name == name {
			return t, true
		}
	}
	return Theme{}, false
}

// DarkTheme is an alias for Monokai, the default theme.
var DarkTheme = Monokai
