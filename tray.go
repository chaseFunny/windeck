package main

import (
	"bytes"
	"image/png"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"windeck/locales"
)

// trayGlyph is Windeck's mark in the menu bar: a deck of cards with a tab
// at its side, drawn black for macOS to tint.
var trayGlyph = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 18 18" fill="none" stroke="#000" stroke-width="1.5" stroke-linejoin="round"><rect x="3.5" y="5" width="11" height="9.5" rx="2"/><path d="M5.5 3h7" stroke-linecap="round"/><path d="M3.5 8.2h3.2" stroke-linecap="round"/></svg>`))

// trayIcon renders the glyph at 18 points on a Retina display.
func trayIcon() []byte {
	img := ui.Render(func(c *ui.Context) {
		c.Root().Background(ui.Transparent)
		ui.Box(c).Fill().Center().Children(func() {
			ui.Image(c, trayGlyph).Size(18, 18)
		})
	}, 18, 18, 2)
	var buf bytes.Buffer
	if png.Encode(&buf, img) != nil {
		return nil
	}
	return buf.Bytes()
}

func newTray(a *App) *mygo.Tray {
	t, err := mygo.NewTray(mygo.TrayOptions{
		ID:             "windeck",
		Icon:           trayIcon(),
		IconIsTemplate: true,
		ToolTip:        "Windeck",
		Menu:           a.trayMenu(),
	})
	if err != nil {
		return nil
	}
	return t
}

func (a *App) trayMenu() *mygo.Menu {
	shortcut := "⌥⇥"
	if a.set.Shortcut == "command" {
		shortcut = "⌘⇥"
	}
	return mygo.NewMenu([]*mygo.MenuItem{
		{Label: locales.TraySearch(shortcut), Click: func(*mygo.MenuItem, *mygo.Window) { a.openSearch() }},
		mygo.Separator(),
		{Label: locales.TraySettings(), Accelerator: "Cmd+,", Click: func(*mygo.MenuItem, *mygo.Window) { a.openSettings() }},
		{Label: locales.TrayPermissions(), Click: func(*mygo.MenuItem, *mygo.Window) { a.openOnboarding() }},
		mygo.Separator(),
		{Label: locales.TrayQuit(), Accelerator: "Cmd+Q", Click: func(*mygo.MenuItem, *mygo.Window) { mygo.App.Quit() }},
	})
}

// openSearch opens the switcher straight in search, from the menu bar.
func (a *App) openSearch() {
	if a.deck.Len() == 0 {
		return
	}
	if !a.open(false) {
		return
	}
	a.do(a.sw.StartSearch())
}

// focusApp brings Windeck's own windows forward: a menu bar app is not
// active until it is asked to be.
func focusApp(w *mygo.Window) {
	mygo.App.Focus()
	w.Show()
	w.Focus()
}
