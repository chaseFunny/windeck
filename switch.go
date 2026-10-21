package main

import (
	"slices"
	"time"

	"github.com/egoist/mygo"

	"windeck/internal/deck"
	"windeck/internal/smart"
	"windeck/internal/switcher"
	"windeck/internal/wm"
)

// The switch: the event tap hands keys to key, on the main thread; the
// state machine (package switcher) says what they do, and do does it.

// appBit marks the IDs of apps without windows, which the list shows at
// its end, apart from windows' numbers.
const appBit = 0x80000000

func appID(pid int32) wm.WindowID { return wm.WindowID(appBit | uint32(pid)) }

func isAppID(id wm.WindowID) (int32, bool) {
	if id&appBit == 0 {
		return 0, false
	}
	return int32(id &^ appBit), true
}

// key handles a key the event tap saw, and tells whether to keep it from
// the app in front.
func (a *App) key(e wm.KeyEvent) bool {
	if a.sw.Phase() == switcher.Idle {
		if e.Key != wm.KeyTab || e.Released {
			return false
		}
		return a.open(e.Shift)
	}
	act, keep := a.sw.Key(e)
	if keep {
		a.ses.keys++
	}
	a.do(act)
	return keep
}

// open starts a switch as the shortcut is pressed. Everything it needs is
// at hand: the windows were listed as they changed, and the prediction is
// a lookup.
func (a *App) open(shift bool) bool {
	if a.deck.Len() == 0 {
		return false
	}
	focused := a.ses.focused
	a.ses.predicted = 0
	if a.set.Predict && a.pro() {
		plain := a.deck.Cycle(focused, 0)
		if cur, ok := a.deck.Window(focused); ok && !a.set.IsExcluded(cur.BundleID) {
			var cands []smart.Item
			for _, id := range plain {
				if w, ok := a.deck.Window(id); ok && id != focused && !a.set.IsExcluded(w.BundleID) {
					cands = append(cands, itemOf(w))
				}
			}
			a.ses.predicted = wm.WindowID(a.model.Predict(itemOf(cur), cands, time.Now()))
		}
	}
	cycle := a.deck.Cycle(focused, a.ses.predicted)
	a.ses.keys = 1
	a.ses.query, a.ses.searchBuf = "", ""
	a.ses.hovered = 0
	a.ses.scroll.Y = 0
	a.ses.scrollTo = true
	act := a.sw.Open(cycle, shift)
	if act.Kind != switcher.StartTimer {
		return false
	}
	wm.SetCapturing(true)
	a.rebuildRows()
	gen := a.ses.gen()
	a.ses.timer = time.AfterFunc(time.Duration(a.set.QuickDelay)*time.Millisecond, func() {
		held := wm.ModifierDown()
		a.update(func() {
			if a.ses.generation == gen {
				a.do(a.sw.Timer(held))
			}
		})
	})
	// Fresh pictures of the windows the user most likely goes to, and a
	// fresh listing, both off the main thread.
	for _, id := range cycle[:min(len(cycle), 4)] {
		if w, ok := a.deck.Window(id); ok {
			a.captureLater(w)
		}
	}
	a.refreshNow()
	return true
}

// do carries out what the state machine decided.
func (a *App) do(act switcher.Action) {
	if act.Reveal {
		a.showPanel(false)
	}
	switch act.Kind {
	case switcher.Show:
		a.showPanel(false)
	case switcher.Redraw:
		a.ses.scrollTo = true
		a.selectionMoved()
	case switcher.Search:
		wm.SetCapturing(false)
		a.showPanel(true)
	case switcher.Switch:
		a.switchTo(act.Window, act.Quick)
	case switcher.CloseWindow:
		a.closeWindow(act.Window)
	case switcher.MinimizeWindow:
		a.minimizeWindow(act.Window)
	case switcher.QuitApp:
		a.quitApp(act.Window)
	}
	if a.sw.Phase() == switcher.Idle {
		a.endSession()
	} else {
		a.invalidatePanel()
	}
}

// selectionMoved fetches a fresh picture of the highlighted window.
func (a *App) selectionMoved() {
	if w, ok := a.deck.Window(a.sw.Selected()); ok {
		a.captureLater(w)
	}
	a.invalidatePanel()
}

func (s *session) gen() int {
	s.generation++
	return s.generation
}

// endSession hides the panel and stops listening to every key.
func (a *App) endSession() {
	wm.SetCapturing(false)
	if a.ses.timer != nil {
		a.ses.timer.Stop()
		a.ses.timer = nil
	}
	a.ses.generation++
	a.ses.query, a.ses.searchBuf = "", ""
	a.hidePanel()
}

// switchTo goes to a window, or brings an app without windows forward.
func (a *App) switchTo(id wm.WindowID, quick bool) {
	place := slices.Index(a.sw.Cycle(), id)
	a.model.Switched(time.Now(), !quick, place, id != 0 && id == a.ses.predicted, a.ses.keys)
	if pid, ok := isAppID(id); ok {
		go wm.ActivateApp(pid)
		return
	}
	w, ok := a.deck.Window(id)
	if !ok {
		return
	}
	a.deck.Focus(id)
	a.ses.focused = id
	go wm.Focus(w)
}

func (a *App) closeWindow(id wm.WindowID) {
	if pid, ok := isAppID(id); ok {
		a.quitPID(pid)
		return
	}
	if _, ok := a.deck.Window(id); !ok {
		return
	}
	go wm.Close(id)
	a.deck.Remove(id)
	a.pool.Remove(uint32(id))
	a.sw.Removed(id)
	a.rebuildRows()
	a.refresh.soon(a)
}

func (a *App) minimizeWindow(id wm.WindowID) {
	w, ok := a.deck.Window(id)
	if !ok || w.Minimized {
		return
	}
	// Its last picture first: once in the Dock it cannot be captured.
	allowed := a.mayCapture(w)
	go func() {
		if allowed {
			if img := wm.CaptureWindow(id, previewSide); img != nil {
				a.pool.Put(uint32(id), img, time.Now())
			}
		}
		wm.Minimize(id)
	}()
	a.deck.Patch(id, func(w *wm.Window) { w.Minimized = true })
	a.sw.Next()
	a.rebuildRows()
	a.refresh.soon(a)
}

func (a *App) quitApp(id wm.WindowID) {
	if pid, ok := isAppID(id); ok {
		a.quitPID(pid)
		return
	}
	if w, ok := a.deck.Window(id); ok {
		a.quitPID(w.PID)
	}
}

func (a *App) quitPID(pid int32) {
	go wm.QuitApp(pid)
	for _, g := range a.deck.Groups() {
		if g.PID != pid {
			continue
		}
		for _, w := range g.Windows {
			a.deck.Remove(w.ID)
			a.sw.Removed(w.ID)
		}
	}
	a.sw.Removed(appID(pid))
	a.apps = slices.DeleteFunc(a.apps, func(ap wm.App) bool { return ap.PID == pid })
	a.rebuildRows()
	a.refresh.soon(a)
}

// rowKind is what a row of the list is.
type rowKind int

const (
	rowWindow rowKind = iota
	rowGroup          // the header of an app's windows
	rowApp            // a running app without windows
)

// row is a row of the panel's list.
type row struct {
	kind  rowKind
	id    wm.WindowID // the window, or the app's ID (appID)
	win   wm.Window
	key   string // the app's
	name  string
	path  string
	pid   int32
	count int
	child bool // in an app's group, under its header
	group int
	open  bool // a group's header: its windows show
}

// rebuildRows lays out the list: by app, in places that do not move, or,
// for a query, the matches best first.
func (a *App) rebuildRows() {
	var rows []row
	if a.ses.query == "" {
		rows = a.groupRows()
	} else {
		rows = a.resultRows()
	}
	a.ses.rows = rows
	var ids []wm.WindowID
	var groups []int
	for _, r := range rows {
		if r.kind != rowGroup {
			ids = append(ids, r.id)
			groups = append(groups, r.group)
		}
	}
	a.sw.SetList(ids, groups)
}

func (a *App) groupRows() []row {
	var rows []row
	sel, _ := a.deck.Window(a.sw.Selected())
	have := map[int32]bool{}
	for gi, g := range a.deck.Groups() {
		have[g.PID] = true
		if len(g.Windows) == 1 {
			w := g.Windows[0]
			rows = append(rows, row{kind: rowWindow, id: w.ID, win: w, key: g.Key, name: g.Name, path: g.Path, pid: g.PID, group: gi})
			continue
		}
		// A collapsed group opens while the highlight is in it.
		open := !a.ses.collapsed[g.Key] || deck.AppKey(sel) == g.Key
		rows = append(rows, row{kind: rowGroup, key: g.Key, name: g.Name, path: g.Path, pid: g.PID, count: len(g.Windows), group: gi, open: open})
		if !open {
			continue
		}
		for _, w := range g.Windows {
			rows = append(rows, row{kind: rowWindow, id: w.ID, win: w, key: g.Key, name: g.Name, path: g.Path, pid: g.PID, child: true, group: gi})
		}
	}
	if a.set.ShowWindowless {
		n := len(rows)
		for _, ap := range a.windowless(have) {
			rows = append(rows, row{kind: rowApp, id: appID(ap.PID), key: appKey(ap), name: ap.Name, path: ap.Path, pid: ap.PID, group: n})
			n++
		}
	}
	return rows
}

// windowless are the running apps of the Dock without windows.
func (a *App) windowless(have map[int32]bool) []wm.App {
	var out []wm.App
	self := int32(osGetpid())
	for _, ap := range a.apps {
		if ap.Regular && !have[ap.PID] && ap.PID != self {
			out = append(out, ap)
		}
	}
	return out
}

func (a *App) resultRows() []row {
	var items []smart.Item
	byID := map[uint32]row{}
	have := map[int32]bool{}
	for gi, g := range a.deck.Groups() {
		have[g.PID] = true
		for _, w := range g.Windows {
			it := itemOf(w)
			items = append(items, it)
			byID[it.ID] = row{kind: rowWindow, id: w.ID, win: w, key: g.Key, name: g.Name, path: g.Path, pid: g.PID, group: gi}
		}
	}
	if a.set.ShowWindowless {
		for _, ap := range a.windowless(have) {
			id := uint32(appID(ap.PID))
			items = append(items, smart.Item{ID: id, App: appKey(ap), Name: ap.Name})
			byID[id] = row{kind: rowApp, id: wm.WindowID(id), key: appKey(ap), name: ap.Name, path: ap.Path, pid: ap.PID}
		}
	}
	personal := a.set.Personal && a.pro()
	now := time.Now().Unix()
	var rows []row
	for i, r := range smart.Rank(a.model, a.ses.query, items, personal, func() int64 { return now }) {
		rw := byID[r.Item.ID]
		rw.group = i
		rows = append(rows, rw)
	}
	return rows
}

// setQuery filters the list as the user types.
func (a *App) setQuery(q string) {
	if q == a.ses.query {
		return
	}
	a.ses.query = q
	a.rebuildRows()
	// The best match takes the highlight.
	for _, r := range a.ses.rows {
		if r.kind != rowGroup {
			a.sw.Hover(r.id)
			break
		}
	}
	a.ses.scroll.Y = 0
	a.selectionMoved()
}

// showPanel shows the panel on the display the user works on; search
// gives it the keyboard, for its search field.
func (a *App) showPanel(search bool) {
	if !a.panel.IsVisible() {
		d := a.panelDisplay()
		a.ses.display = d
		w := min(int(float64(d.WorkArea.Width)*0.84), 1360)
		h := min(int(float64(d.WorkArea.Height)*0.8), 860)
		w, h = max(w, min(760, d.WorkArea.Width)), max(h, min(480, d.WorkArea.Height))
		a.panel.SetBounds(mygo.Rectangle{
			X: d.WorkArea.X + (d.WorkArea.Width-w)/2, Y: d.WorkArea.Y + (d.WorkArea.Height-h)/2,
			Width: w, Height: h,
		})
		a.ses.pointer = mygo.Screen.CursorScreenPoint()
		a.selectionMoved()
	}
	if search {
		a.panel.Show()
		a.panel.Focus()
		a.ses.focusSearch = true
	} else if !a.panel.IsVisible() {
		a.panel.ShowInactive()
	}
	a.invalidatePanel()
}

// panelDisplay is where the panel shows: the display of the window the
// user is in, or of the pointer, or the main one as the settings say.
func (a *App) panelDisplay() mygo.Display {
	if a.set.PrimaryDisplay {
		return mygo.Screen.PrimaryDisplay()
	}
	if w, ok := a.deck.Window(a.ses.focused); ok && !w.Minimized && w.Bounds.W > 0 {
		return mygo.Screen.DisplayMatching(mygo.Rectangle{X: int(w.Bounds.X), Y: int(w.Bounds.Y), Width: int(w.Bounds.W), Height: int(w.Bounds.H)})
	}
	return mygo.Screen.DisplayNearestPoint(mygo.Screen.CursorScreenPoint())
}

func (a *App) hidePanel() {
	if a.panel != nil && a.panel.IsVisible() {
		a.panel.Hide()
	}
	a.ses.hovered = 0
	a.ses.focusSearch = false
}

// invalidatePanel draws the panel anew.
func (a *App) invalidatePanel() {
	if a.panel != nil {
		a.panel.Invalidate()
	}
}
