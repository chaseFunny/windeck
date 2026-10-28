package main

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"windeck/internal/deck"
	"windeck/internal/license"
	"windeck/internal/settings"
	"windeck/internal/smart"
	"windeck/internal/snapshot"
	"windeck/internal/switcher"
	"windeck/internal/wm"
)

// App is Windeck's state. It lives on the main thread: other goroutines
// change it through update.
type App struct {
	quitting bool   // Windeck is quitting: the panel closes too
	dir      string // Windeck's folder of Application Support
	set      settings.Settings
	lic      license.State
	model    *smart.Model
	pool     *snapshot.Pool
	deck     *deck.Deck
	sw       switcher.Switcher
	// apps are the running apps, for those without windows.
	apps []wm.App
	// holds are what Windeck just did to windows, over the listings for a
	// moment, and heldWin the windows as they were then.
	intro   introState
	holds   map[wm.WindowID]held
	heldWin map[wm.WindowID]wm.Window

	trusted    bool // the accessibility permission
	canCapture bool // the screen recording permission
	tapOn      bool

	panel      *mygo.Window
	setWin     *mygo.Window
	onboardWin *mygo.Window
	tray       *mygo.Tray

	ses     session // the switch under way
	icons   iconCache
	shots   previewCache
	refresh refresher

	ss  settingsState
	ob  onboardingState
	upd updateState
}

// session is a switch, from the shortcut pressed to the switch made.
type session struct {
	rounded   bool // the panel has its corners
	query     string
	searchBuf string // what the search field holds
	focused   wm.WindowID
	predicted wm.WindowID
	keys      int
	timer     *time.Timer
	collapsed map[string]bool
	rows      []row
	scroll    ui.ScrollState
	scrollTo  bool // the highlight moved by keys: show it
	search    ui.Handle
	hovered   wm.WindowID
	display   mygo.Display
	// generation counts sessions, so that a timer of one that ended does
	// nothing; focusSearch asks the view to focus the search field.
	generation  int
	focusSearch bool
	// guide shows the keys over the panel, after the first launch.
	guide bool
	// pointer is where the pointer was as the panel showed, or last
	// highlighted a row: hovering highlights only once it moves.
	pointer mygo.Point
}

func newApp() *App {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, "Library", "Application Support", "Windeck")
	}
	if devShow() != "" {
		dir = filepath.Join(os.TempDir(), "windeck-dev")
	}
	a := &App{
		dir:   dir,
		pool:  snapshot.New(snapshot.DefaultBudget),
		deck:  deck.New(),
		icons: iconCache{byPath: map[string]*ui.Bitmap{}, loading: map[string]bool{}},
		shots: previewCache{byID: map[wm.WindowID]*preview{}, capturing: map[wm.WindowID]bool{}, failed: map[wm.WindowID]bool{}},
	}
	a.ses.collapsed = map[string]bool{}
	a.holds = map[wm.WindowID]held{}
	a.heldWin = map[wm.WindowID]wm.Window{}
	a.set = settings.Load(a.path("settings.json"))
	a.lic = license.Load(a.path("license.json"), time.Now())
	a.model = smart.Load(a.path("learned.json"))
	return a
}

func (a *App) path(name string) string { return filepath.Join(a.dir, name) }

func osGetpid() int { return os.Getpid() }

func timeoutContext(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

// update runs fn on the main thread, and draws the panel.
func (a *App) update(fn func()) {
	if a.panel != nil {
		a.panel.Update(fn)
		return
	}
	mygo.App.WhenReady(fn)
}

// redraw draws the windows that show what changed.
func (a *App) redraw() {
	for _, w := range []*mygo.Window{a.panel, a.setWin, a.onboardWin} {
		if w != nil {
			w.Invalidate()
		}
	}
}

func (a *App) start() {
	a.panel = newPanel(a)
	a.tray = newTray(a)
	a.trusted = wm.Trusted(false)
	a.canCapture = wm.CanCapture(false)
	if err := wm.Observe(a.changed); err != nil {
		a.trusted = false
	}
	mygo.App.SetOpenAtLogin(a.set.OpenAtLogin)
	// The apps in the order they launched order the list's groups, and
	// the windows front to back start the history.
	var keys []string
	for _, ap := range wm.RunningApps() {
		keys = append(keys, appKey(ap))
	}
	a.deck.SeedApps(keys)
	a.deck.SeedHistory(wm.StackingOrder())
	a.ses.focused = wm.FocusedWindow()
	a.deck.Focus(a.ses.focused)
	a.refreshNow()
	if dev := devShow(); dev != "" {
		a.showForDev(dev)
	} else {
		a.startTap()
		switch {
		case !a.set.Onboarded:
			a.showIntro()
		case !a.trusted:
			a.openOnboarding()
		}
	}
	go a.validateLicense()
	if devShow() == "" {
		a.startUpdates()
	}
	// Learned data is written now and then, not on every switch.
	go func() {
		for range time.Tick(5 * time.Minute) {
			a.model.Save()
		}
	}()
}

func appKey(ap wm.App) string {
	return deck.AppKey(wm.Window{BundleID: ap.BundleID, PID: ap.PID})
}

// startTap starts the shortcut, once the user allowed Windeck to.
func (a *App) startTap() {
	if a.tapOn {
		wm.SetTapModifier(a.modifier())
		wm.SetSystemSwitcher(a.set.Shortcut != settings.ShortcutCommand)
		return
	}
	if !a.trusted {
		return
	}
	if err := wm.StartTap(a.modifier(), a.key); err != nil {
		return
	}
	a.tapOn = true
	wm.SetSystemSwitcher(a.set.Shortcut != settings.ShortcutCommand)
}

func (a *App) modifier() wm.Modifier {
	if a.set.Shortcut == settings.ShortcutCommand {
		return wm.ModCommand
	}
	return wm.ModOption
}

func (a *App) shutdown() {
	// Give macOS its own Cmd-Tab back.
	wm.SetSystemSwitcher(true)
	a.model.Save()
}

// saveSettings writes the settings and applies those that act at once.
func (a *App) saveSettings() {
	a.set.Save(a.path("settings.json"))
	mygo.Theme.SetSource(themeSource(a.set.Appearance))
	mygo.App.SetLocale(a.set.Language)
	mygo.App.SetOpenAtLogin(a.set.OpenAtLogin)
	a.startTap()
	a.scheduleUpdateCheck()
	a.redraw()
}

// pro reports whether Pro's features are on: licensed or in the trial.
func (a *App) pro() bool { return a.lic.Status(time.Now()).Pro() }

func themeSource(appearance string) mygo.ThemeSource {
	switch appearance {
	case "dark":
		return mygo.ThemeDark
	case "system":
		return mygo.ThemeSystem
	}
	return mygo.ThemeLight
}

// changed hears what changed in the user's windows, on any goroutine.
func (a *App) changed(c wm.Change) {
	switch c.Event {
	case wm.EventFocus:
		a.update(func() { a.focusChanged(c.Leaving, c.Window) })
	case wm.EventChanged, wm.EventSpace:
		a.refresh.soon(a)
	}
}

// focusChanged learns that the user went from one window to another, and
// keeps the picture of the one left, for when it can no longer be
// captured.
func (a *App) focusChanged(from, to wm.WindowID) {
	if to == 0 {
		// An app without windows, or the desktop.
		if a.sw.Phase() == switcher.Idle {
			a.ses.focused = 0
		}
		if fw, ok := a.deck.Window(from); ok {
			a.captureLater(fw)
		}
		return
	}
	a.deck.Focus(to)
	if a.sw.Phase() == switcher.Idle {
		a.ses.focused = to
	}
	fw, okFrom := a.deck.Window(from)
	tw, okTo := a.deck.Window(to)
	if okTo && !a.set.IsExcluded(tw.BundleID) {
		var fi smart.Item
		if okFrom && !a.set.IsExcluded(fw.BundleID) {
			fi = itemOf(fw)
		}
		a.model.Visit(fi, itemOf(tw), time.Now())
	}
	if okFrom {
		a.captureLater(fw)
	}
	if !okTo {
		// A window Windeck has not listed yet.
		a.refresh.soon(a)
	}
}

func itemOf(w wm.Window) smart.Item {
	return smart.Item{ID: uint32(w.ID), App: deck.AppKey(w), Name: w.AppName, Title: w.Title}
}

// refresher lists the windows again, one listing at a time, soon after
// changes, which come in bursts.
type refresher struct {
	running bool
	again   bool
	timer   *time.Timer
}

// soon lists the windows after a short pause, from any goroutine.
func (r *refresher) soon(a *App) {
	a.update(func() {
		if r.timer != nil {
			return
		}
		r.timer = time.AfterFunc(120*time.Millisecond, func() {
			a.update(func() {
				r.timer = nil
				a.refreshNow()
			})
		})
	})
}

// refreshNow lists the windows now, on another goroutine.
func (a *App) refreshNow() {
	r := &a.refresh
	if r.running {
		r.again = true
		return
	}
	r.running = true
	go func() {
		ws := wm.Windows()
		apps := wm.RunningApps()
		a.update(func() {
			r.running = false
			a.listed(ws, apps)
			if r.again {
				r.again = false
				a.refreshNow()
			}
		})
	}()
}

// held is what Windeck itself just did to a window, which the listings
// may not show yet: a window takes a moment to go to the Dock, come back
// or close, and a listing in between would move its row back and forth.
type held struct {
	until     time.Time
	minimized bool
	gone      bool
}

// holdFor is how long what Windeck did outweighs the listings.
const holdFor = 1200 * time.Millisecond

// hold keeps what Windeck did to a window over the listings for a moment,
// then lists the windows again, to take what they say from then on.
func (a *App) hold(id wm.WindowID, h held) {
	h.until = time.Now().Add(holdFor)
	if w, ok := a.deck.Window(id); ok {
		a.heldWin[id] = w
	}
	a.holds[id] = h
	time.AfterFunc(holdFor+50*time.Millisecond, func() { a.update(a.refreshNow) })
}

// applyHolds makes a listing say what Windeck just did.
func (a *App) applyHolds(ws []wm.Window) []wm.Window {
	now := time.Now()
	for id, h := range a.holds {
		if now.After(h.until) {
			delete(a.holds, id)
			delete(a.heldWin, id)
			continue
		}
		i := slices.IndexFunc(ws, func(w wm.Window) bool { return w.ID == id })
		switch {
		case h.gone && i >= 0:
			ws = slices.Delete(ws, i, i+1)
		case h.gone:
		case i >= 0:
			ws[i].Minimized = h.minimized
		default:
			// Missed by this listing, on its way: kept as it was.
			if w, ok := a.heldWin[id]; ok {
				w.Minimized = h.minimized
				ws = append(ws, w)
			}
		}
	}
	return ws
}

// listed takes a new listing of the windows.
func (a *App) listed(ws []wm.Window, apps []wm.App) {
	// Windeck's own windows are not switched to.
	pid := int32(os.Getpid())
	ws = slices.DeleteFunc(ws, func(w wm.Window) bool { return w.PID == pid })
	ws = a.applyHolds(ws)
	a.deck.Update(ws)
	a.apps = apps
	a.pool.Keep(func(id uint32) bool { _, ok := a.deck.Window(wm.WindowID(id)); return ok })
	a.shots.keep(a)
	a.icons.load(a, ws)
	if a.sw.Phase() != switcher.Idle {
		a.rebuildRows()
	}
}

// validateLicense checks the license with Lemon Squeezy now and then, as
// after a refund; offline, it stays.
func (a *App) validateLicense() {
	st := a.lic
	if st.Key == "" || time.Since(st.Checked) < 7*24*time.Hour {
		return
	}
	ctx, cancel := timeoutContext(15 * time.Second)
	defer cancel()
	err := st.Validate(ctx, time.Now())
	if err != nil && err != license.ErrInvalid {
		return
	}
	a.update(func() {
		a.lic = st
		a.lic.Save(a.path("license.json"))
	})
}

// resetLearned forgets what was learned and every picture kept.
func (a *App) resetLearned() {
	a.model.Reset()
	a.pool.Clear()
	a.shots.clear()
	a.redraw()
}

// devShow is WINDECK_DEV_SHOW: panel, settings or onboarding, to show
// that window at launch and look at it, without the shortcut taking the
// keyboard and with the data in a folder of its own.
func devShow() string { return os.Getenv("WINDECK_DEV_SHOW") }

func (a *App) showForDev(what string) {
	switch what {
	case "panel", "search":
		// Once the windows are listed.
		time.AfterFunc(time.Second, func() {
			a.update(func() {
				if what == "search" {
					a.openSearch()
					return
				}
				if a.open(false) {
					a.do(a.sw.Timer(true))
				}
			})
		})
	case "settings":
		a.openSettings()
	case "intro":
		a.showIntro()
	case "guide":
		time.AfterFunc(time.Second, func() { a.update(a.showGuide) })
	default:
		a.openOnboarding()
	}
}
