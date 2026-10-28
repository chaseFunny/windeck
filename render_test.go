package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"windeck/internal/deck"
	"windeck/internal/settings"
	"windeck/internal/smart"
	"windeck/internal/snapshot"
	"windeck/internal/wm"
)

// TestRenderLive draws the panel, Settings and onboarding with this Mac's
// windows into PNGs, to look at: WINDECK_RENDER=<dir>.
func TestRenderLive(t *testing.T) {
	dir := os.Getenv("WINDECK_RENDER")
	if dir == "" {
		t.Skip("WINDECK_RENDER not set")
	}
	registerFonts()
	a := testApp(t)
	ws := wm.Windows()
	a.listed(ws, wm.RunningApps())
	a.deck.SeedHistory(wm.StackingOrder())
	a.ses.focused = a.deck.Current()
	for _, w := range ws[:min(len(ws), 8)] {
		if img := wm.CaptureWindow(w.ID, previewSide); img != nil {
			a.pool.Put(uint32(w.ID), img, time.Now())
		}
	}
	a.canCapture = true
	a.sw.Open(a.deck.Cycle(a.ses.focused, 0), false)
	a.rebuildRows()
	a.sw.Timer(true)
	// The panel as a 14-inch MacBook Pro shows it, Settings and the
	// onboarding at their windows' sizes.
	const pw, ph = 1270, 755
	render(t, dir, "panel-light.png", a.panelView, false, pw, ph)
	render(t, dir, "panel-dark.png", a.panelView, true, pw, ph)
	a.set.Previews = false
	render(t, dir, "panel-nopicture.png", a.panelView, false, pw, ph)
	a.set.Previews = true
	// The list layout, as tall as its rows.
	a.set.Layout = settings.LayoutList
	a.rebuildRows()
	lh := min(2*listPad+len(a.ses.rows)*listRowHeight, ph)
	render(t, dir, "panel-list.png", a.panelView, false, listWidth, lh)
	render(t, dir, "panel-list-dark.png", a.panelView, true, listWidth, lh)
	a.set.Layout = settings.LayoutTabs
	a.rebuildRows()
	a.sw.StartSearch()
	a.ses.searchBuf = "go"
	a.setQuery("go")
	render(t, dir, "panel-search.png", a.panelView, false, pw, ph)
	a.set.Layout = settings.LayoutList
	render(t, dir, "panel-list-search.png", a.panelView, false, listWidth, min(lh+searchBlock, ph))
	a.set.Layout = settings.LayoutTabs
	a.ses.searchBuf = "zzzz"
	a.setQuery("zzzz")
	render(t, dir, "panel-search-empty.png", a.panelView, false, pw, ph)
	// The guide over the panel, after the first launch.
	a.ses.searchBuf = ""
	a.setQuery("")
	a.ses.guide = true
	render(t, dir, "panel-guide.png", a.panelView, false, pw, ph)
	render(t, dir, "panel-guide-dark.png", a.panelView, true, pw, ph)
	a.set.Layout = settings.LayoutList
	a.rebuildRows()
	render(t, dir, "panel-list-guide.png", a.panelView, false, listWidth, guideHeight)
	a.set.Layout = settings.LayoutTabs
	a.rebuildRows()
	a.ses.guide = false
	for i, name := range []string{"general", "shortcuts", "privacy", "smart", "license", "about"} {
		a.ss.page = i
		render(t, dir, "settings-"+name+".png", a.settingsView, false, 880, 620)
	}
	a.ss.page = pagePrivacy
	a.ss.scroll.Y = 520
	render(t, dir, "settings-privacy-2.png", a.settingsView, false, 880, 620)
	a.ss.scroll.Y = 0
	a.ss.page = pageGeneral
	render(t, dir, "settings-general-dark.png", a.settingsView, true, 880, 620)
	a.ob.step = 0
	a.trusted, a.canCapture = false, false
	render(t, dir, "onboarding-1.png", a.onboardingView, false, 680, 640)
	a.trusted, a.canCapture = true, true
	render(t, dir, "onboarding-1-done.png", a.onboardingView, false, 680, 640)
	a.ob.step = 1
	render(t, dir, "onboarding-2.png", a.onboardingView, false, 680, 640)
}

func testApp(t *testing.T) *App {
	a := &App{
		dir:   t.TempDir(),
		set:   settings.Default(),
		pool:  snapshot.New(snapshot.DefaultBudget),
		deck:  deck.New(),
		icons: iconCache{byPath: map[string]*ui.Bitmap{}, loading: map[string]bool{}},
		shots: previewCache{byID: map[wm.WindowID]*preview{}, capturing: map[wm.WindowID]bool{}, failed: map[wm.WindowID]bool{}},
	}
	a.ses.collapsed = map[string]bool{}
	a.holds = map[wm.WindowID]held{}
	a.heldWin = map[wm.WindowID]wm.Window{}
	a.model = newTestModel()
	a.lic.TrialStart = time.Now()
	// Icons come from the running app; here every app has Windeck's.
	logo := a.appLogo()
	for _, w := range wm.RunningApps() {
		for _, px := range []int{64, 256} {
			a.icons.byPath[w.Path+"@"+itoa(px)] = logo
		}
	}
	a.trusted = true
	a.ss.applied = a.set.Clone()
	return a
}

func render(t *testing.T, dir, name string, view func(*ui.Context), dark bool, w, h int) {
	t.Helper()
	img := ui.Render(func(c *ui.Context) {
		if dark {
			c.SetTheme(ui.DarkTheme())
		}
		view(c)
	}, w, h, 2)
	os.MkdirAll(dir, 0o755)
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	png.Encode(f, img)
}

func newTestModel() *smart.Model { return smart.New("") }

// TestRenderUpdates draws the update window's states: WINDECK_RENDER=<dir>.
func TestRenderUpdates(t *testing.T) {
	dir := os.Getenv("WINDECK_RENDER")
	if dir == "" {
		t.Skip("WINDECK_RENDER not set")
	}
	registerFonts()
	a := testApp(t)
	a.upd.found = testUpdate()
	for _, st := range []struct {
		name  string
		phase int
	}{{"available", updAvailable}, {"downloading", updDownloading}, {"latest", updLatest}, {"checking", updChecking}} {
		a.upd.phase = st.phase
		a.upd.got, a.upd.total = 1_800_000, 4_900_000
		w, h := a.updateWindowSize()
		render(t, dir, "update-"+st.name+".png", a.updateView, false, w, h)
	}
	a.upd.phase = updAvailable
	w, h := a.updateWindowSize()
	render(t, dir, "update-available-dark.png", a.updateView, true, w, h)
}

func testUpdate() *mygo.Update {
	return &mygo.Update{Version: "1.1.0", Notes: "### Added\n\n- A second layout, **List**: icons and titles alone.\n- A Shortcuts page in Settings.\n- Updates from the menu bar.\n\n### Fixed\n\n- The switcher no longer flashes as Tab is held."}
}

// TestRenderIntro draws the first launch's intro at moments of its
// timeline, on a 1512 by 982 screen (without the GPU's aurora).
func TestRenderIntro(t *testing.T) {
	dir := os.Getenv("WINDECK_RENDER")
	if dir == "" {
		t.Skip("WINDECK_RENDER not set")
	}
	registerFonts()
	a := testApp(t)
	a.intro = introState{word: wordGlyphs(), toOnboarding: true}
	var x float32
	for _, g := range a.intro.word {
		a.intro.offset = append(a.intro.offset, x)
		a.intro.boxes = append(a.intro.boxes, glyphBox(g))
		x += g.advance
	}
	d := mygo.Display{Bounds: mygo.Rectangle{Width: 1512, Height: 982}, WorkArea: mygo.Rectangle{Y: 33, Width: 1512, Height: 949}}
	a.intro.display = d
	ox, oy := onboardingOrigin(d)
	a.intro.land = ui.Point{X: float32(ox + onboardW/2), Y: float32(oy + 96)}
	a.intro.landSide = 64 * iconGrid / 1024
	for _, ms := range []float32{700, 1500, 2150, 2450, 2750, 3100, 3400, 3900, 4600, 5050} {
		for _, dark := range []bool{false} {
			name := fmt.Sprintf("intro-%04d.png", int(ms))
			render(t, dir, name, func(c *ui.Context) {
				pal := themed(c)
				c.Root().Background(pal.bgSubtle)
				ui.Box(c).Fill().Draw(func(p *ui.Painter, r ui.Rect) { a.paintIntro(p, r, pal, ms) })
			}, dark, 1512, 982)
		}
	}
}

// The intro plays once a Mac: recorded as it starts, in a file of its own.
func TestIntroPlaysOnce(t *testing.T) {
	a := testApp(t)
	if a.introPlayed() {
		t.Fatal("a new Mac has played the intro")
	}
	a.markIntroPlayed()
	if !a.introPlayed() {
		t.Fatal("the intro is not recorded")
	}
	a.resetLearned()
	if !a.introPlayed() {
		t.Fatal("resetting what was learned forgot the intro")
	}
}
