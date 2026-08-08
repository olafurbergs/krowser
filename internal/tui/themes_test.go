package tui

import "testing"

func TestThemeByName(t *testing.T) {
	if t_, ok := ThemeByName("Monokai"); !ok || !t_.Dark {
		t.Error("Monokai should resolve as a dark theme")
	}
	if t_, ok := ThemeByName("Catppuccin Latte"); !ok || t_.Dark {
		t.Error("Catppuccin Latte should resolve as a light theme")
	}
	if _, ok := ThemeByName("nope"); ok {
		t.Error("unknown theme should not resolve")
	}
}

func TestDefaultTheme(t *testing.T) {
	if DefaultTheme.Name != "Monokai" {
		t.Errorf("DefaultTheme = %q, want Monokai", DefaultTheme.Name)
	}
	if DarkTheme.Name != "Monokai" {
		t.Errorf("DarkTheme alias = %q, want Monokai", DarkTheme.Name)
	}
}
