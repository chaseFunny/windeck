package main

import (
	"os"
	"strings"
	"unicode/utf8"
)

// shortTitle drops from a window's title what repeats its app's name,
// which the list shows anyway: Chrome's "Page - Google Chrome - Work"
// reads "Page", Activity Monitor's "Activity Monitor – CPU" reads "CPU".
// A title that is only the app's name is kept.
func shortTitle(title, app string) string {
	if app == "" || title == "" {
		return title
	}
	for _, sep := range []string{" - ", " — ", " – "} {
		// "Page - App", or "Page - App - Profile".
		if i := strings.LastIndex(title, sep+app); i > 0 {
			rest := title[i+len(sep)+len(app):]
			if rest == "" || strings.HasPrefix(rest, sep) && !strings.Contains(rest[len(sep):], sep) && utf8.RuneCountInString(rest) <= 32 {
				return strings.TrimSpace(title[:i])
			}
		}
		// "App - Page".
		if after, ok := strings.CutPrefix(title, app+sep); ok && strings.TrimSpace(after) != "" {
			return strings.TrimSpace(after)
		}
	}
	return title
}

// compareFold orders names as Finder does, without regard to case.
func compareFold(a, b string) int {
	return strings.Compare(strings.ToLower(a), strings.ToLower(b))
}

// tildePath writes a path in the home folder as Finder's Go menu does:
// ~/Library/Application Support/Windeck.
func tildePath(p string) string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if rest, ok := strings.CutPrefix(p, home); ok && (rest == "" || rest[0] == '/') {
			return "~" + rest
		}
	}
	return p
}
