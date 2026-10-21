package wm

import "sync"

// The elements of the windows listed, kept to act on them: an element
// stays valid as long as its window lives.
var (
	elemMu sync.Mutex
	elems  = map[WindowID]uintptr{} // owned
	// notWindows are windows of the window server, on other Spaces, that
	// the search by token did not find: an app's invisible helpers, which
	// are not looked for again.
	notWindows = map[WindowID]bool{}
)

func keepElement(id WindowID, el uintptr) {
	elemMu.Lock()
	defer elemMu.Unlock()
	if old := elems[id]; old != 0 && old != el {
		release(old)
	}
	elems[id] = el
}

func element(id WindowID) uintptr {
	elemMu.Lock()
	defer elemMu.Unlock()
	return retain(elems[id])
}

// forget releases the elements of windows that are gone.
func forget(live map[WindowID]bool) {
	elemMu.Lock()
	defer elemMu.Unlock()
	for id, el := range elems {
		if !live[id] {
			release(el)
			delete(elems, id)
		}
	}
}

// axWindow is what the accessibility API says of a window.
type axWindow struct {
	title      string
	minimized  bool
	fullscreen bool
}

func describe(el uintptr) (axWindow, bool) {
	sub := axString(el, "AXSubrole")
	if sub != "AXStandardWindow" && sub != "AXDialog" {
		return axWindow{}, false
	}
	return axWindow{
		title:      axString(el, "AXTitle"),
		minimized:  axBool(el, "AXMinimized"),
		fullscreen: axBool(el, "AXFullScreen"),
	}, true
}

// Windows lists the windows of the running apps, on every Space and
// display, minimized and hidden ones too. It asks every app, so call it
// off the main thread; an app that does not answer is skipped after a
// tenth of a second. skip are apps (bundle IDs) whose windows are listed
// without asking them for their titles.
func Windows() []Window {
	apps := RunningApps()
	cg := cgWindows()
	ds := Displays()
	sp := spaces()
	trusted := Trusted(false)

	byPID := map[int32][]cgWindow{}
	for _, w := range cg {
		byPID[w.pid] = append(byPID[w.pid], w)
	}

	type result struct {
		app  App
		wins []Window
	}
	results := make([]result, len(apps))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i, app := range apps {
		results[i].app = app
		if len(byPID[app.PID]) == 0 {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() { <-sem; wg.Done() }()
			results[i].wins = appWindows(app, byPID[app.PID], trusted)
		}()
	}
	wg.Wait()

	var out []Window
	live := map[WindowID]bool{}
	for _, r := range results {
		for _, w := range r.wins {
			live[w.ID] = true
			w.PID, w.AppName, w.BundleID, w.AppPath, w.Hidden = r.app.PID, r.app.Name, r.app.BundleID, r.app.Path, r.app.Hidden
			w.Display = displayOf(ds, w.Bounds)
			if w.Display < len(ds) {
				w.External = !ds[w.Display].BuiltIn
			}
			if !w.Minimized {
				if s, ok := sp[spaceOf(w.ID)]; ok {
					w.Space = s.number
					w.OnCurrentSpace = s.current
					if s.fullscreen {
						w.FullScreen = true
					}
				} else {
					w.OnCurrentSpace = true // on every Space
				}
			}
			out = append(out, w)
		}
	}
	forget(live)
	return out
}

// appWindows lists an app's windows: those the accessibility API lists,
// then those of the window server it leaves out, on other Spaces, which
// its elements are found for by token.
func appWindows(app App, cg []cgWindow, trusted bool) []Window {
	var out []Window
	seen := map[WindowID]bool{}
	bounds := map[WindowID]cgWindow{}
	for _, w := range cg {
		bounds[w.id] = w
	}
	if !trusted {
		// Without the accessibility permission, the window server's list
		// alone: the windows that show, untitled unless Windeck may
		// record the screen.
		for _, w := range cg {
			if w.onscreen && w.bounds.W >= minimumWindowSize && w.bounds.H >= minimumWindowSize && w.alpha > 0 {
				out = append(out, Window{ID: w.id, Title: w.title, Bounds: w.bounds})
			}
		}
		return out
	}
	ax := newAXApp(app.PID)
	if ax == 0 {
		return nil
	}
	defer release(ax)
	for _, el := range axWindowsOf(ax) {
		id := axWindowID(el)
		d, ok := describe(el)
		if id == 0 || !ok || seen[id] {
			release(el)
			continue
		}
		seen[id] = true
		keepElement(id, el)
		w := Window{ID: id, Title: d.title, Minimized: d.minimized, FullScreen: d.fullscreen, Bounds: bounds[id].bounds}
		if w.Title == "" {
			w.Title = bounds[id].title
		}
		out = append(out, w)
	}
	// Windows of the window server the app did not list: on other Spaces.
	want := map[WindowID]bool{}
	elemMu.Lock()
	for _, w := range cg {
		if seen[w.id] || w.onscreen || notWindows[w.id] || w.bounds.W < minimumWindowSize || w.bounds.H < minimumWindowSize {
			continue
		}
		if el := elems[w.id]; el != 0 {
			continue // found before: described below
		}
		want[w.id] = true
	}
	elemMu.Unlock()
	found := axWindowsByToken(app.PID, want)
	elemMu.Lock()
	for id := range want {
		if found[id] == 0 {
			notWindows[id] = true
		}
	}
	elemMu.Unlock()
	for id, el := range found {
		keepElement(id, el)
	}
	for _, w := range cg {
		if seen[w.id] || w.onscreen {
			continue
		}
		el := element(w.id)
		if el == 0 {
			continue
		}
		d, ok := describe(el)
		release(el)
		if !ok {
			continue
		}
		seen[w.id] = true
		t := d.title
		if t == "" {
			t = w.title
		}
		out = append(out, Window{ID: w.id, Title: t, Minimized: d.minimized, FullScreen: d.fullscreen, Bounds: w.bounds})
	}
	return out
}

// StackingOrder returns the windows that show, front to back: the most
// recently used first, as the switcher's history starts.
func StackingOrder() []WindowID {
	var out []WindowID
	for _, w := range cgWindows() {
		if w.onscreen && w.alpha > 0 && w.bounds.W >= minimumWindowSize && w.bounds.H >= minimumWindowSize {
			out = append(out, w.id)
		}
	}
	return out
}

// Focus brings a window to the front, from whichever Space, minimized or
// hidden it is.
func Focus(w Window) {
	el := element(w.ID)
	defer release(el)
	if w.Hidden {
		ActivateApp(w.PID)
	}
	if el != 0 && w.Minimized {
		axSetBool(el, "AXMinimized", false)
	}
	if !focusWindow(w.PID, w.ID) {
		ActivateApp(w.PID)
	}
	if el != 0 {
		axAction(el, "AXRaise")
	}
	noteFocus(w.ID)
}

// Close closes a window, as its close button does.
func Close(id WindowID) bool {
	el := element(id)
	if el == 0 {
		return false
	}
	defer release(el)
	btn := axAttr(el, "AXCloseButton")
	if btn == 0 {
		return false
	}
	defer release(btn)
	return axAction(btn, "AXPress")
}

// Minimize sends a window to the Dock.
func Minimize(id WindowID) bool {
	el := element(id)
	if el == 0 {
		return false
	}
	defer release(el)
	return axSetBool(el, "AXMinimized", true)
}

// FocusedWindow returns the window the user is in: the focused window of
// the frontmost app.
func FocusedWindow() WindowID {
	pid := FrontmostPID()
	if pid == 0 {
		return 0
	}
	ax := newAXApp(pid)
	if ax == 0 {
		return 0
	}
	defer release(ax)
	el := axAttr(ax, "AXFocusedWindow")
	if el == 0 {
		return 0
	}
	defer release(el)
	return axWindowID(el)
}
