// Package settings is the user's settings, kept as JSON in Windeck's
// folder of Application Support.
package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

// Shortcut modes: Option-Tab beside the system's Cmd-Tab, or Cmd-Tab in
// its place.
const (
	ShortcutOption  = "option"
	ShortcutCommand = "command"
)

// Settings are the user's choices.
type Settings struct {
	// Onboarded tells that the user went through the first launch's
	// steps.
	Onboarded bool `json:"onboarded"`
	// Shortcut is ShortcutOption or ShortcutCommand.
	Shortcut string `json:"shortcut"`
	// Appearance is "light" (the default), "dark" or "system".
	Appearance string `json:"appearance"`
	// Language is "" for the system's, "en" or "zh-Hans".
	Language string `json:"language"`
	// QuickDelay is how long the shortcut is held, in milliseconds, before
	// the panel shows: a release before switches without it.
	QuickDelay int `json:"quickDelay"`
	// Predict puts the predicted window in the cycle's second place.
	Predict bool `json:"predict"`
	// Personal ranks search results by the user's habits.
	Personal bool `json:"personal"`
	// Previews shows windows' pictures; off, the list alone.
	Previews bool `json:"previews"`
	// ShowWindowless lists running apps without windows at the end.
	ShowWindowless bool `json:"showWindowless"`
	// PrimaryDisplay shows the panel on the main display, rather than on
	// the one the user works on.
	PrimaryDisplay bool `json:"primaryDisplay"`
	// Excluded are apps (bundle IDs) whose windows are never captured,
	// whose titles are never kept, and which nothing is learned of.
	Excluded []string `json:"excluded"`
	// OpenAtLogin starts Windeck as the user logs in.
	OpenAtLogin bool `json:"openAtLogin"`
}

// DefaultExcluded are password managers, the keychain and banking apps.
var DefaultExcluded = []string{
	"com.1password.1password",
	"com.agilebits.onepassword7",
	"com.agilebits.onepassword-osx",
	"com.bitwarden.desktop",
	"com.apple.keychainaccess",
	"com.apple.Passwords",
	"org.keepassxc.keepassxc",
	"com.dashlane.dashlanephonefinal",
	"com.lastpass.LastPass",
	"com.alipay.mobile.client",
	"com.cmbchina.CMBMobileBank",
}

// Default returns the settings of a new user.
func Default() Settings {
	return Settings{
		Shortcut:   ShortcutOption,
		Appearance: "light",
		QuickDelay: 150,
		Predict:    true,
		Personal:   true,
		Previews:   true,
		Excluded:   slices.Clone(DefaultExcluded),
	}
}

// Load reads the settings saved at path, or the defaults.
func Load(path string) Settings {
	s := Default()
	if b, err := os.ReadFile(path); err == nil {
		json.Unmarshal(b, &s)
	}
	return s.normalized()
}

func (s Settings) normalized() Settings {
	if s.Shortcut != ShortcutCommand {
		s.Shortcut = ShortcutOption
	}
	switch s.Appearance {
	case "light", "dark", "system":
	default:
		s.Appearance = "light"
	}
	switch s.Language {
	case "", "en", "zh-Hans":
	default:
		s.Language = ""
	}
	s.QuickDelay = min(max(s.QuickDelay, 0), 600)
	if s.Excluded == nil {
		s.Excluded = []string{}
	}
	return s
}

// Save writes the settings to path.
func (s Settings) Save(path string) error {
	b, err := json.MarshalIndent(s.normalized(), "", "  ")
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(path), 0o700)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// IsExcluded reports whether an app is on the privacy list.
func (s Settings) IsExcluded(bundleID string) bool {
	return bundleID != "" && slices.Contains(s.Excluded, bundleID)
}

// Clone returns a copy that shares nothing with s.
func (s Settings) Clone() Settings {
	s.Excluded = slices.Clone(s.Excluded)
	return s
}
