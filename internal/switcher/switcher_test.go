package switcher

import (
	"testing"

	"windeck/internal/wm"
)

var (
	cycle  = []wm.WindowID{2, 3, 4, 1}
	list   = []wm.WindowID{1, 2, 3, 4}
	groups = []int{0, 0, 1, 2}
)

func open(t *testing.T) *Switcher {
	t.Helper()
	s := &Switcher{}
	s.SetList(list, groups)
	if a := s.Open(cycle, false); a.Kind != StartTimer {
		t.Fatalf("Open: %+v", a)
	}
	return s
}

func key(s *Switcher, k wm.Key) Action {
	a, _ := s.Key(wm.KeyEvent{Key: k, Mod: true})
	return a
}

func TestQuickReleaseSwitchesWithoutThePanel(t *testing.T) {
	s := open(t)
	a := s.Release()
	if a.Kind != Switch || a.Window != 2 || !a.Quick {
		t.Fatalf("release: %+v", a)
	}
	if s.Phase() != Idle {
		t.Fatal("not idle")
	}
	// The timer firing afterwards does nothing.
	if a := s.Timer(true); a.Kind != None {
		t.Fatalf("late timer: %+v", a)
	}
}

func TestTabsBeforeTheDelayStayHidden(t *testing.T) {
	s := open(t)
	key(s, wm.KeyTab)
	if s.Shown() {
		t.Fatal("a second Tab showed the panel")
	}
	if a := s.Release(); a.Window != 3 || !a.Quick {
		t.Fatalf("release: %+v", a)
	}
}

func TestHoldingShowsThePanel(t *testing.T) {
	s := open(t)
	if a := s.Timer(true); a.Kind != Show {
		t.Fatalf("timer: %+v", a)
	}
	key(s, wm.KeyTab)
	key(s, wm.KeyTab)
	if s.Selected() != 4 {
		t.Fatalf("selected %d", s.Selected())
	}
	a, _ := s.Key(wm.KeyEvent{Key: wm.KeyTab, Shift: true, Mod: true})
	if a.Kind != Redraw || s.Selected() != 3 {
		t.Fatalf("shift-tab: %+v, %d", a, s.Selected())
	}
	if a := s.Release(); a.Kind != Switch || a.Window != 3 || a.Quick {
		t.Fatalf("release: %+v", a)
	}
}

func TestTimerWithoutTheModifierSwitches(t *testing.T) {
	s := open(t)
	if a := s.Timer(false); a.Kind != Switch || a.Window != 2 {
		t.Fatalf("timer with the modifier up: %+v", a)
	}
}

func TestArrowsFollowTheList(t *testing.T) {
	s := open(t)
	a := key(s, wm.KeyDown) // 2 → 3, in the list's order
	if !a.Reveal || s.Selected() != 3 {
		t.Fatalf("down: %+v, %d", a, s.Selected())
	}
	key(s, wm.KeyDown)
	key(s, wm.KeyDown) // wraps to 1
	if s.Selected() != 1 {
		t.Fatalf("selected %d", s.Selected())
	}
	key(s, wm.KeyUp)
	if s.Selected() != 4 {
		t.Fatalf("up wrapped to %d", s.Selected())
	}
	// Tab picks the cycle up from where the arrows left the highlight.
	key(s, wm.KeyTab)
	if s.Selected() != 1 {
		t.Fatalf("tab after arrows: %d", s.Selected())
	}
}

func TestGroupJumps(t *testing.T) {
	s := open(t) // on 2, in group 0
	key(s, wm.KeyRight)
	if s.Selected() != 3 {
		t.Fatalf("right: %d", s.Selected())
	}
	key(s, wm.KeyRight)
	key(s, wm.KeyRight) // wraps to group 0
	if s.Selected() != 1 {
		t.Fatalf("right wrapped to %d", s.Selected())
	}
	key(s, wm.KeyLeft) // back to the last group
	if s.Selected() != 4 {
		t.Fatalf("left wrapped to %d", s.Selected())
	}
	key(s, wm.KeyLeft)
	if s.Selected() != 3 {
		t.Fatalf("left: %d", s.Selected())
	}
}

func TestActionsKeepThePanel(t *testing.T) {
	s := open(t)
	s.Timer(true)
	a := key(s, wm.KeyW)
	if a.Kind != CloseWindow || a.Window != 2 || s.Phase() != Browsing {
		t.Fatalf("w: %+v", a)
	}
	s.Removed(2)
	if s.Selected() != 3 {
		t.Fatalf("after close, selected %d, want the next", s.Selected())
	}
	if a := key(s, wm.KeyM); a.Kind != MinimizeWindow || a.Window != 3 {
		t.Fatalf("m: %+v", a)
	}
	if a := key(s, wm.KeyQ); a.Kind != QuitApp {
		t.Fatalf("q: %+v", a)
	}
}

// A held W repeats: only its first press closes, or the highlight, gone on
// to the next window, would be closed too, and the app's windows after it.
func TestHeldActionKeysActOnce(t *testing.T) {
	s := open(t)
	s.Timer(true)
	if a := key(s, wm.KeyW); a.Kind != CloseWindow || a.Window != 2 {
		t.Fatalf("w: %+v", a)
	}
	s.Removed(2)
	for _, k := range []wm.Key{wm.KeyW, wm.KeyM, wm.KeyQ} {
		a, keep := s.Key(wm.KeyEvent{Key: k, Repeat: true, Mod: true})
		if a.Kind != None || !keep {
			t.Fatalf("repeated %v: %+v, kept %v", k, a, keep)
		}
	}
	if s.Selected() != 3 || s.Phase() != Browsing {
		t.Fatalf("after the repeats: selected %d, phase %v", s.Selected(), s.Phase())
	}
	// Tab held still goes on.
	if a, _ := s.Key(wm.KeyEvent{Key: wm.KeyTab, Repeat: true, Mod: true}); s.Selected() == 3 {
		t.Fatalf("held tab did not move: %+v", a)
	}
}

func TestSearchState(t *testing.T) {
	s := open(t)
	a := key(s, wm.KeySpace)
	if a.Kind != Search || !a.Reveal || s.Phase() != Searching {
		t.Fatalf("space: %+v", a)
	}
	// The modifier going up no longer switches, and keys go to the field.
	if a := s.Release(); a.Kind != None {
		t.Fatalf("release while searching: %+v", a)
	}
	if _, keep := s.Key(wm.KeyEvent{Key: wm.KeyW}); keep {
		t.Fatal("the tap kept a key typed into the search")
	}
	s.SetList([]wm.WindowID{4, 1}, []int{0, 1})
	if s.Selected() != 4 {
		t.Fatalf("results: selected %d", s.Selected())
	}
	s.Move(1)
	if a := s.Enter(); a.Kind != Switch || a.Window != 1 {
		t.Fatalf("enter: %+v", a)
	}
}

func TestEscapeInSearchClearsFirst(t *testing.T) {
	s := open(t)
	s.StartSearch()
	if a := s.Escape(false); a.Kind != Redraw || s.Phase() != Searching {
		t.Fatalf("escape with a query: %+v", a)
	}
	if a := s.Escape(true); a.Kind != Cancel || s.Phase() != Idle {
		t.Fatalf("escape without: %+v", a)
	}
}

func TestEscapeCancels(t *testing.T) {
	s := open(t)
	if a := key(s, wm.KeyEscape); a.Kind != Cancel {
		t.Fatalf("escape: %+v", a)
	}
}

func TestClickAndHover(t *testing.T) {
	s := open(t)
	s.Timer(true)
	s.Hover(4)
	if s.Selected() != 4 {
		t.Fatal("hover")
	}
	if a := s.Click(1); a.Kind != Switch || a.Window != 1 {
		t.Fatalf("click: %+v", a)
	}
}

func TestOtherKeysAreKeptWhileBrowsing(t *testing.T) {
	s := open(t)
	if _, keep := s.Key(wm.KeyEvent{Key: wm.KeyOther, Mod: true}); !keep {
		t.Fatal("a key typed with Option held reached the app behind")
	}
}

func TestEmptyCycleDoesNothing(t *testing.T) {
	s := &Switcher{}
	if a := s.Open(nil, false); a.Kind != None || s.Phase() != Idle {
		t.Fatalf("open empty: %+v", a)
	}
}
