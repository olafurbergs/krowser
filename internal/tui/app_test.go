package tui

import (
	"path/filepath"
	"testing"

	"github.com/olafurb/krowser/internal/k8s"
)

func TestThemeSelectionApplies(t *testing.T) {
	m := New(Config{Client: &k8s.Client{}, Namespace: "default", ThemeFile: filepath.Join(t.TempDir(), "theme")})
	if m.theme.Name != "Monokai" {
		t.Fatalf("default theme = %q, want Monokai", m.theme.Name)
	}

	// Open the theme picker.
	if _, cmd := m.handleScreenMessage(setScreenMsg{screen: screenPickerThemes}); cmd != nil {
		t.Fatalf("unexpected command while opening picker: %v", cmd)
	}
	if m.cur != m.picker {
		t.Fatalf("expected theme picker to be current, got %T", m.cur)
	}

	// Emit a selection for the Dracula item.
	var item *pickerItem
	for i := range m.picker.items {
		if m.picker.items[i].title == "● Dracula" || m.picker.items[i].title == "  Dracula" {
			item = &m.picker.items[i]
			break
		}
	}
	if item == nil {
		t.Fatal("Dracula item not found in theme picker")
	}
	msg := m.picker.emit(*item)()
	sel, ok := msg.(selectedThemeMsg)
	if !ok {
		t.Fatalf("expected selectedThemeMsg, got %T", msg)
	}

	// Dispatch the selection through the root model. The returned command is
	// the resource screen's reload, which is expected after navigating back.
	if _, cmd := m.Update(sel); cmd == nil {
		t.Fatal("expected a reload command after theme selection")
	}
	if m.theme.Name != "Dracula" {
		t.Errorf("theme after selection = %q, want Dracula", m.theme.Name)
	}
	if m.resource.theme.Name != "Dracula" {
		t.Errorf("resource screen theme = %q, want Dracula", m.resource.theme.Name)
	}
	if got := LoadSavedTheme(m.themeFile); got != "Dracula" {
		t.Errorf("persisted theme = %q, want Dracula", got)
	}
}

func TestThemeLoadedFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme")
	if err := SaveTheme(path, "Nord"); err != nil {
		t.Fatal(err)
	}
	m := New(Config{Client: &k8s.Client{}, Namespace: "default", ThemeFile: path})
	if m.theme.Name != "Nord" {
		t.Errorf("theme from saved file = %q, want Nord", m.theme.Name)
	}
}

func TestThemeFlagOverridesSaved(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme")
	if err := SaveTheme(path, "Nord"); err != nil {
		t.Fatal(err)
	}
	m := New(Config{Client: &k8s.Client{}, Namespace: "default", ThemeFile: path, Theme: "Gruvbox Dark"})
	if m.theme.Name != "Gruvbox Dark" {
		t.Errorf("theme = %q, want Gruvbox Dark (flag wins)", m.theme.Name)
	}
}
