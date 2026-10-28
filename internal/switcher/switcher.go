// Package switcher is the switcher's state machine: what each key does
// while the shortcut is held (browsing) and once the user starts typing
// (searching), and when the panel shows at all.
//
// The machine has no clock and draws nothing: it returns Actions, which
// the app carries out, so that every rule here is tested without a
// window.
package switcher

import (
	"slices"

	"windeck/internal/wm"
)

// Phase is where the switcher is.
type Phase int

const (
	// Idle: the shortcut is not held, nothing shows.
	Idle Phase = iota
	// Pending: the shortcut was pressed, and the panel waits a moment
	// before it shows, so that a quick press and release switches without
	// the panel flashing up.
	Pending
	// Browsing (state A): the panel shows while the modifier is held;
	// releasing it switches.
	Browsing
	// Searching (state B): the panel stays without the modifier, and keys
	// type into its search field.
	Searching
)

// ActionKind is what the app is to do.
type ActionKind int

const (
	None ActionKind = iota
	// StartTimer: wait the quick-switch delay, then call Timer.
	StartTimer
	// Show the panel.
	Show
	// Cancel: hide the panel, switching nowhere.
	Cancel
	// Switch to Window, hiding the panel. Quick tells that the panel never
	// showed.
	Switch
	// Search: the panel takes the keyboard for its search field, and
	// stays without the modifier.
	Search
	// CloseWindow, MinimizeWindow and QuitApp act on Window, the panel
	// staying.
	CloseWindow
	MinimizeWindow
	QuitApp
	// Redraw: the selection moved.
	Redraw
)

// Action is something for the app to do.
type Action struct {
	Kind   ActionKind
	Window wm.WindowID
	Quick  bool
	// Reveal: show the panel first, as a key other than Tab does while
	// the switcher still waits to show.
	Reveal bool
}

// Switcher is the state machine.
type Switcher struct {
	phase Phase
	// cycle is the order of Tab, fixed as the switcher opens.
	cycle []wm.WindowID
	// list is the windows in the order the panel shows them, the order of
	// the arrow keys, and group the index of each one's app group.
	list  []wm.WindowID
	group []int
	sel   wm.WindowID
	shown bool
}

// Phase reports where the switcher is.
func (s *Switcher) Phase() Phase { return s.phase }

// Selected is the highlighted window.
func (s *Switcher) Selected() wm.WindowID { return s.sel }

// Cycle is the order of Tab, for the panel's 1, 2, 3 hints.
func (s *Switcher) Cycle() []wm.WindowID { return s.cycle }

// Shown reports whether the panel shows.
func (s *Switcher) Shown() bool { return s.shown }

// Open starts a switch as the shortcut is pressed: the highlight goes to
// the first window of the cycle, or the last with Shift.
func (s *Switcher) Open(cycle []wm.WindowID, shift bool) Action {
	if s.phase != Idle {
		return s.tab(shift)
	}
	if len(cycle) == 0 {
		return Action{}
	}
	s.cycle = slices.Clone(cycle)
	s.phase = Pending
	s.shown = false
	s.sel = s.cycle[0]
	if shift {
		s.sel = s.cycle[len(s.cycle)-1]
	}
	return Action{Kind: StartTimer}
}

// SetList gives the windows in the panel's order, with each one's group;
// the highlight stays on its window, or goes to the first when that one
// left the list.
func (s *Switcher) SetList(ids []wm.WindowID, groups []int) {
	s.list, s.group = ids, groups
	if s.phase == Idle {
		return
	}
	if len(ids) == 0 {
		s.sel = 0
		return
	}
	if !slices.Contains(ids, s.sel) {
		s.sel = ids[0]
	}
}

// Timer is the quick-switch delay elapsed: the panel shows, unless the
// modifier went up meanwhile without the event saying so (held false).
func (s *Switcher) Timer(held bool) Action {
	if s.phase != Pending {
		return Action{}
	}
	if !held {
		return s.Release()
	}
	return s.show(Browsing)
}

func (s *Switcher) show(p Phase) Action {
	s.phase = p
	if s.shown {
		return Action{Kind: Redraw}
	}
	s.shown = true
	return Action{Kind: Show}
}

// Release is the modifier going up: in state A, switch.
func (s *Switcher) Release() Action {
	switch s.phase {
	case Pending, Browsing:
		quick := !s.shown
		return s.finish(Action{Kind: Switch, Window: s.sel, Quick: quick})
	}
	return Action{}
}

func (s *Switcher) finish(a Action) Action {
	s.phase = Idle
	s.shown = false
	s.cycle = nil
	if a.Kind == Switch && a.Window == 0 {
		a.Kind = Cancel
	}
	return a
}

// tab moves the highlight. Before the panel shows, along the cycle: a
// blind press goes back to the previous window, the next to the one
// before. Once it shows, down the list the user sees, one row a press,
// never jumping about; Shift goes up.
func (s *Switcher) tab(back bool) Action {
	if s.shown && len(s.list) > 0 {
		if back {
			return s.move(-1)
		}
		return s.move(1)
	}
	if len(s.cycle) == 0 {
		return Action{}
	}
	i := slices.Index(s.cycle, s.sel)
	switch {
	case i < 0 && back:
		i = len(s.cycle) - 1
	case i < 0:
		i = 0
	case back:
		i = (i - 1 + len(s.cycle)) % len(s.cycle)
	default:
		i = (i + 1) % len(s.cycle)
	}
	s.sel = s.cycle[i]
	return Action{Kind: Redraw}
}

// move moves the highlight along the list, by the panel's order.
func (s *Switcher) move(delta int) Action {
	if len(s.list) == 0 {
		return Action{}
	}
	i := slices.Index(s.list, s.sel)
	if i < 0 {
		i = 0
	} else {
		i = (i + delta + len(s.list)) % len(s.list)
	}
	s.sel = s.list[i]
	return Action{Kind: Redraw}
}

// jumpGroup moves the highlight to the first window of the app group
// before or after the highlighted one's.
func (s *Switcher) jumpGroup(delta int) Action {
	if len(s.list) == 0 || len(s.group) != len(s.list) {
		return Action{}
	}
	i := max(slices.Index(s.list, s.sel), 0)
	g := s.group[i]
	if delta > 0 {
		for j := i + 1; j < len(s.list); j++ {
			if s.group[j] != g {
				s.sel = s.list[j]
				return Action{Kind: Redraw}
			}
		}
		s.sel = s.list[0]
		return Action{Kind: Redraw}
	}
	// Back: the first window of the group before.
	j := i
	for j > 0 && s.group[j-1] == g {
		j--
	}
	if j == 0 {
		j = len(s.list) // wrap to the last group
	}
	prev := s.group[j-1]
	for j > 0 && s.group[j-1] == prev {
		j--
	}
	s.sel = s.list[j]
	return Action{Kind: Redraw}
}

// Key handles a key while the shortcut's modifier is held (state A, or
// still pending), and reports whether to keep it from the app in front.
func (s *Switcher) Key(e wm.KeyEvent) (Action, bool) {
	if s.phase != Pending && s.phase != Browsing {
		return Action{}, false
	}
	if e.Released {
		return s.Release(), false
	}
	// Any key but Tab shows the panel at once: the user is looking.
	reveal := func(a Action) Action {
		if !s.shown {
			s.phase = Browsing
			s.shown = true
			a.Reveal = true
		}
		return a
	}
	if e.Repeat && (e.Key == wm.KeyW || e.Key == wm.KeyM || e.Key == wm.KeyQ) {
		// A held key repeats, and the highlight goes on to the next window
		// after each: W held would close the app's windows one after the
		// other, Q quit app after app. One press, one act.
		return Action{}, true
	}
	switch e.Key {
	case wm.KeyTab:
		return s.tab(e.Shift), true
	case wm.KeyGrave:
		return s.tab(true), true
	case wm.KeyEscape:
		return s.finish(Action{Kind: Cancel}), true
	case wm.KeyReturn:
		return s.finish(Action{Kind: Switch, Window: s.sel, Quick: !s.shown}), true
	case wm.KeyUp:
		return reveal(s.move(-1)), true
	case wm.KeyDown:
		return reveal(s.move(1)), true
	case wm.KeyLeft:
		return reveal(s.jumpGroup(-1)), true
	case wm.KeyRight:
		return reveal(s.jumpGroup(1)), true
	case wm.KeyW:
		return reveal(Action{Kind: CloseWindow, Window: s.sel}), true
	case wm.KeyM:
		return reveal(Action{Kind: MinimizeWindow, Window: s.sel}), true
	case wm.KeyQ:
		return reveal(Action{Kind: QuitApp, Window: s.sel}), true
	case wm.KeySpace, wm.KeySlash:
		a := reveal(Action{Kind: Search})
		s.phase = Searching
		return a, true
	}
	// Other keys with the modifier held would type into the app behind:
	// the switcher keeps them.
	return Action{}, true
}

// Removed takes a window out of the cycle and the list, as one just
// closed, the highlight going to the next in the list.
func (s *Switcher) Removed(id wm.WindowID) {
	if i := slices.Index(s.list, id); i >= 0 {
		next := wm.WindowID(0)
		if i+1 < len(s.list) {
			next = s.list[i+1]
		} else if i > 0 {
			next = s.list[i-1]
		}
		s.list = slices.Delete(slices.Clone(s.list), i, i+1)
		if len(s.group) > i {
			s.group = slices.Delete(slices.Clone(s.group), i, i+1)
		}
		if s.sel == id {
			s.sel = next
		}
	}
	s.cycle = slices.DeleteFunc(s.cycle, func(x wm.WindowID) bool { return x == id })
	if s.sel == id {
		s.sel = 0
		if len(s.cycle) > 0 {
			s.sel = s.cycle[0]
		}
	}
}

// Next moves the highlight to the next window of the list, as after one
// was minimized.
func (s *Switcher) Next() {
	if s.phase != Idle {
		s.move(1)
	}
}

// Hover highlights the window under the pointer.
func (s *Switcher) Hover(id wm.WindowID) Action {
	if s.phase == Idle || s.sel == id {
		return Action{}
	}
	s.sel = id
	return Action{Kind: Redraw}
}

// Click switches to the window clicked.
func (s *Switcher) Click(id wm.WindowID) Action {
	if s.phase == Idle {
		return Action{}
	}
	return s.finish(Action{Kind: Switch, Window: id})
}

// StartSearch goes to state B, as a click in the search field does.
func (s *Switcher) StartSearch() Action {
	if s.phase == Idle || s.phase == Searching {
		return Action{}
	}
	a := Action{Kind: Search, Reveal: !s.shown}
	s.phase = Searching
	s.shown = true
	return a
}

// Move moves the highlight in the results, in state B (Up, Down, Tab).
func (s *Switcher) Move(delta int) Action {
	if s.phase != Searching {
		return Action{}
	}
	return s.move(delta)
}

// Enter switches to the highlighted result, in state B.
func (s *Switcher) Enter() Action {
	if s.phase != Searching {
		return Action{}
	}
	return s.finish(Action{Kind: Switch, Window: s.sel})
}

// Escape, in state B, clears the query when there is one (the app does,
// as queryEmpty false says) or else closes the switcher.
func (s *Switcher) Escape(queryEmpty bool) Action {
	if s.phase == Idle {
		return Action{}
	}
	if s.phase == Searching && !queryEmpty {
		return Action{Kind: Redraw}
	}
	return s.finish(Action{Kind: Cancel})
}

// Dismiss closes the switcher, as the panel losing the keyboard does.
func (s *Switcher) Dismiss() Action {
	if s.phase == Idle {
		return Action{}
	}
	return s.finish(Action{Kind: Cancel})
}
