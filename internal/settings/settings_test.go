package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultsAndRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s := Load(path)
	if s.Shortcut != ShortcutOption || s.Appearance != "light" || s.QuickDelay != 150 || !s.IsExcluded("com.1password.1password") {
		t.Fatalf("defaults %+v", s)
	}
	s.Shortcut = ShortcutCommand
	s.Excluded = append(s.Excluded, "com.example.bank")
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	l := Load(path)
	if l.Shortcut != ShortcutCommand || !l.IsExcluded("com.example.bank") {
		t.Fatalf("loaded %+v", l)
	}
}

func TestNormalizes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(path, []byte(`{"shortcut":"hyper","appearance":"neon","quickDelay":9999,"language":"xx"}`), 0o600)
	s := Load(path)
	if s.Shortcut != ShortcutOption || s.Appearance != "light" || s.QuickDelay != 600 || s.Language != "" {
		t.Fatalf("normalized %+v", s)
	}
}
