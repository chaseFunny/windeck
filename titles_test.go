package main

import "testing"

func TestShortTitle(t *testing.T) {
	for _, tc := range []struct{ title, app, want string }{
		{"Onyx AI - Google Chrome - 蜗牛", "Google Chrome", "Onyx AI"},
		{"net/http - Google Chrome", "Google Chrome", "net/http"},
		{"活动监视器 – 所有进程", "活动监视器", "所有进程"},
		{"Google Chrome", "Google Chrome", "Google Chrome"},
		{"a - b - c", "Notes", "a - b - c"},
		{"server.go — api", "Code", "server.go — api"},
		{"", "Finder", ""},
	} {
		if got := shortTitle(tc.title, tc.app); got != tc.want {
			t.Errorf("shortTitle(%q, %q) = %q, want %q", tc.title, tc.app, got, tc.want)
		}
	}
}
