package tui

import (
	"os"
	"path/filepath"
	"strings"
)

// ThemeConfigPath returns the file where the selected theme is persisted.
// It lives under the user data directory (XDG_DATA_HOME or ~/.local/share).
func ThemeConfigPath() string {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dir, "krowser", "theme")
}

// LoadSavedTheme reads the persisted theme name. It returns "" when the file
// is missing, unreadable, or empty.
func LoadSavedTheme(path string) string {
	if path == "" {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// SaveTheme writes the theme name so it can be restored on the next run.
func SaveTheme(path, name string) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(name+"\n"), 0o600)
}
