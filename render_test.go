package main

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	render(t, dir, "panel-light.png", a.panelView, false)
	render(t, dir, "panel-dark.png", a.panelView, true)
	a.sw.StartSearch()
	a.ses.searchBuf = "go"
	a.setQuery("go")
	render(t, dir, "panel-search.png", a.panelView, false)
	for i, name := range []string{"general", "privacy", "smart", "license"} {
		a.ss.page = i
		render(t, dir, "settings-"+name+".png", a.settingsView, false)
	}
	a.ob.step = 0
	render(t, dir, "onboarding-1.png", a.onboardingView, false)
	a.ob.step = 1
	render(t, dir, "onboarding-2.png", a.onboardingView, false)
}

func testApp(t *testing.T) *App {
	a := &App{
		dir:   t.TempDir(),
		set:   settings.Default(),
		pool:  snapshot.New(snapshot.DefaultBudget),
		deck:  deck.New(),
		icons: iconCache{byPath: map[string]*ui.Bitmap{}, loading: map[string]bool{}},
		shots: previewCache{byID: map[wm.WindowID]*preview{}, capturing: map[wm.WindowID]bool{}},
	}
	a.ses.collapsed = map[string]bool{}
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

func render(t *testing.T, dir, name string, view func(*ui.Context), dark bool) {
	t.Helper()
	img := ui.Render(func(c *ui.Context) {
		if dark {
			c.SetTheme(ui.DarkTheme())
		}
		view(c)
	}, 1180, 740, 2)
	os.MkdirAll(dir, 0o755)
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	png.Encode(f, img)
}

func newTestModel() *smart.Model { return smart.New("") }
