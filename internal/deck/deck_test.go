package deck

import (
	"slices"
	"testing"

	"windeck/internal/wm"
)

func win(id wm.WindowID, app string) wm.Window {
	return wm.Window{ID: id, BundleID: app, AppName: app, Title: app + "-" + itoa(int(id))}
}

func TestGroupsKeepTheirPlaces(t *testing.T) {
	d := New()
	d.SeedApps([]string{"code", "chrome", "term"})
	d.Update([]wm.Window{win(3, "term"), win(1, "chrome"), win(2, "code"), win(4, "code")})
	d.Focus(4)
	d.Focus(1)
	gs := d.Groups()
	var names []string
	for _, g := range gs {
		names = append(names, g.Key)
	}
	if !slices.Equal(names, []string{"code", "chrome", "term"}) {
		t.Fatalf("groups %v, want the order apps launched", names)
	}
	if gs[0].Windows[0].ID != 2 || gs[0].Windows[1].ID != 4 {
		t.Fatalf("code's windows %v, want the order they appeared", gs[0].Windows)
	}
	// Using windows moves the history, never the list.
	d.Focus(2)
	if g := d.Groups(); g[0].Windows[0].ID != 2 || g[0].Windows[1].ID != 4 {
		t.Fatal("focusing reordered the list")
	}
	// A new window goes to the end of its app's group.
	d.Update([]wm.Window{win(3, "term"), win(1, "chrome"), win(2, "code"), win(4, "code"), win(9, "code")})
	if g := d.Groups(); g[0].Windows[2].ID != 9 {
		t.Fatalf("new window at %v", g[0].Windows)
	}
}

func TestCycleFirstPlaceIsAlwaysThePreviousWindow(t *testing.T) {
	d := New()
	d.Update([]wm.Window{win(1, "a"), win(2, "b"), win(3, "c"), win(4, "d")})
	for _, id := range []wm.WindowID{4, 3, 2, 1} {
		d.Focus(id)
	}
	// MRU: 1 2 3 4; the user is in 1.
	if got := d.Cycle(1, 0); !slices.Equal(got, []wm.WindowID{2, 3, 4, 1}) {
		t.Fatalf("cycle %v", got)
	}
	// A prediction takes the second place, never the first.
	if got := d.Cycle(1, 4); !slices.Equal(got, []wm.WindowID{2, 4, 3, 1}) {
		t.Fatalf("cycle with prediction %v", got)
	}
	if got := d.Cycle(1, 2); !slices.Equal(got, []wm.WindowID{2, 3, 4, 1}) {
		t.Fatalf("a prediction of the first place moved it: %v", got)
	}
	// On the desktop, no window is current: the cycle starts at the
	// most recently used one.
	if got := d.Cycle(0, 0); !slices.Equal(got, []wm.WindowID{1, 2, 3, 4}) {
		t.Fatalf("cycle from the desktop %v", got)
	}
}

func TestClosedWindowsLeaveTheHistory(t *testing.T) {
	d := New()
	d.Update([]wm.Window{win(1, "a"), win(2, "b")})
	d.Focus(2)
	d.Update([]wm.Window{win(1, "a")})
	if got := d.MRU(); !slices.Equal(got, []wm.WindowID{1}) {
		t.Fatalf("mru %v", got)
	}
	d.Remove(1)
	if d.Len() != 0 || len(d.MRU()) != 0 {
		t.Fatal("Remove kept the window")
	}
}

func TestSeedHistory(t *testing.T) {
	d := New()
	d.SeedHistory([]wm.WindowID{3, 1})
	d.Update([]wm.Window{win(1, "a"), win(2, "b"), win(3, "c")})
	if got := d.MRU(); !slices.Equal(got, []wm.WindowID{3, 1, 2}) {
		t.Fatalf("mru %v", got)
	}
}
