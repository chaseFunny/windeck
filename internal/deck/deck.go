// Package deck is the switcher's model of the user's windows: grouped by
// app in an order that does not move, and remembered in the order they
// were used (MRU), from which the Tab key's order is drawn.
//
// The list keeps places stable so the user's spatial memory holds: apps
// stay in the order Windeck first saw them, windows in the order they
// appeared. Only the highlight jumps, along the cycle: the most recently
// used windows first.
package deck

import (
	"slices"

	"windeck/internal/wm"
)

// Deck is the windows and their history. It is not safe for concurrent
// use: the app keeps it on the main thread.
type Deck struct {
	windows map[wm.WindowID]wm.Window
	// seq is when each window and each app was first seen, which orders
	// the list.
	winSeq map[wm.WindowID]int
	appSeq map[string]int
	next   int
	// mru is the windows, the one the user is in first.
	mru []wm.WindowID
}

// New returns an empty deck.
func New() *Deck {
	return &Deck{windows: map[wm.WindowID]wm.Window{}, winSeq: map[wm.WindowID]int{}, appSeq: map[string]int{}}
}

// AppKey is what groups a window with its app's: its bundle ID, or its
// process for an app without one.
func AppKey(w wm.Window) string {
	if w.BundleID != "" {
		return w.BundleID
	}
	return "pid:" + itoa(int(w.PID))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// SeedApps orders apps before their windows are listed: as they
// launched.
func (d *Deck) SeedApps(keys []string) {
	for _, k := range keys {
		if _, ok := d.appSeq[k]; !ok {
			d.appSeq[k] = d.next
			d.next++
		}
	}
}

// SeedHistory starts the history from the windows' stacking order, front
// to back, for windows not used since Windeck started.
func (d *Deck) SeedHistory(order []wm.WindowID) {
	for _, id := range order {
		if !slices.Contains(d.mru, id) {
			d.mru = append(d.mru, id)
		}
	}
}

// Update replaces the windows with those listed now: new ones join the
// end of their app's group, and gone ones leave the history.
func (d *Deck) Update(ws []wm.Window) {
	live := make(map[wm.WindowID]wm.Window, len(ws))
	for _, w := range ws {
		live[w.ID] = w
		if _, ok := d.appSeq[AppKey(w)]; !ok {
			d.appSeq[AppKey(w)] = d.next
			d.next++
		}
		if _, ok := d.winSeq[w.ID]; !ok {
			d.winSeq[w.ID] = d.next
			d.next++
		}
	}
	for id := range d.winSeq {
		if _, ok := live[id]; !ok {
			delete(d.winSeq, id)
		}
	}
	d.windows = live
	d.mru = slices.DeleteFunc(d.mru, func(id wm.WindowID) bool { _, ok := live[id]; return !ok })
	// Windows never used go after those that were, in the list's order.
	var unused []wm.WindowID
	for _, w := range ws {
		if !slices.Contains(d.mru, w.ID) {
			unused = append(unused, w.ID)
		}
	}
	slices.SortFunc(unused, func(a, b wm.WindowID) int { return d.winSeq[a] - d.winSeq[b] })
	d.mru = append(d.mru, unused...)
}

// Remove drops a window at once, as one the user just closed.
func (d *Deck) Remove(id wm.WindowID) {
	delete(d.windows, id)
	d.mru = slices.DeleteFunc(d.mru, func(x wm.WindowID) bool { return x == id })
}

// Patch changes a window in place, as one just minimized.
func (d *Deck) Patch(id wm.WindowID, fn func(*wm.Window)) {
	if w, ok := d.windows[id]; ok {
		fn(&w)
		d.windows[id] = w
	}
}

// Focus records that the user went to a window.
func (d *Deck) Focus(id wm.WindowID) {
	if id == 0 {
		return
	}
	i := slices.Index(d.mru, id)
	if i == 0 {
		return
	}
	if i > 0 {
		d.mru = slices.Delete(d.mru, i, i+1)
	}
	d.mru = slices.Insert(d.mru, 0, id)
}

// Window returns a window by its ID.
func (d *Deck) Window(id wm.WindowID) (wm.Window, bool) {
	w, ok := d.windows[id]
	return w, ok
}

// Len is the number of windows.
func (d *Deck) Len() int { return len(d.windows) }

// Current is the window the user is in: the most recently used.
func (d *Deck) Current() wm.WindowID {
	if len(d.mru) == 0 {
		return 0
	}
	return d.mru[0]
}

// MRU returns the windows, the most recently used first.
func (d *Deck) MRU() []wm.WindowID { return slices.Clone(d.mru) }

// Cycle is the order Tab goes through the windows: the window before the
// current one first, always, then the predicted one if any, then the
// others from the most recently used, and the current window last.
// focused is the window the user is in: when it is not the most recently
// used, as on the desktop, the cycle starts at that one instead.
//
// The first place is never the prediction's: one Tab always goes back to
// where the user just was.
func (d *Deck) Cycle(focused, predicted wm.WindowID) []wm.WindowID {
	if len(d.mru) == 0 {
		return nil
	}
	if focused != d.mru[0] {
		rest := slices.Clone(d.mru)
		if predicted != 0 && len(rest) > 1 && predicted != rest[0] {
			if i := slices.Index(rest, predicted); i > 1 {
				rest = slices.Delete(rest, i, i+1)
				rest = slices.Insert(rest, 1, predicted)
			}
		}
		return rest
	}
	cur := d.mru[0]
	rest := slices.Clone(d.mru[1:])
	if predicted != 0 && len(rest) > 1 && predicted != rest[0] && predicted != cur {
		if i := slices.Index(rest, predicted); i > 1 {
			rest = slices.Delete(rest, i, i+1)
			rest = slices.Insert(rest, 1, predicted)
		}
	}
	return append(rest, cur)
}

// Group is an app's windows, in the list's order.
type Group struct {
	Key     string
	Name    string
	Path    string
	PID     int32
	Windows []wm.Window
}

// Groups returns the windows by app, apps in the order first seen and
// windows in the order they appeared.
func (d *Deck) Groups() []Group {
	byApp := map[string]*Group{}
	var keys []string
	for _, w := range d.windows {
		k := AppKey(w)
		g := byApp[k]
		if g == nil {
			g = &Group{Key: k, Name: w.AppName, Path: w.AppPath, PID: w.PID}
			byApp[k] = g
			keys = append(keys, k)
		}
		g.Windows = append(g.Windows, w)
	}
	slices.SortFunc(keys, func(a, b string) int { return d.appSeq[a] - d.appSeq[b] })
	out := make([]Group, 0, len(keys))
	for _, k := range keys {
		g := byApp[k]
		slices.SortFunc(g.Windows, func(a, b wm.Window) int { return d.winSeq[a.ID] - d.winSeq[b.ID] })
		out = append(out, *g)
	}
	return out
}
