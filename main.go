// Windeck (窗签) is a window switcher for macOS: hold Option and press Tab,
// and the user's windows show as a browser's tabs, grouped by app down
// the left, with the highlighted one large in the middle. Releasing Option
// goes there; a quick press and release goes back to the previous window
// without anything showing. Space or slash searches.
//
// It lives in the menu bar. The windows' pictures, the user's habits and
// what was learned of them stay on this Mac.
package main

import (
	"log"

	"github.com/egoist/mygo"

	"windeck/locales"
)

func main() {
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}
	// A menu bar app: no Dock icon, no app switcher entry of its own.
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
	mygo.App.OnWindowAllClosed(func() {})
	registerFonts()
	a := newApp()
	mygo.Theme.SetSource(themeSource(a.set.Appearance))
	mygo.App.SetLocale(a.set.Language)
	mygo.App.OnLocaleChanged(func(string) {
		if a.tray != nil {
			a.tray.SetMenu(a.trayMenu())
		}
		if a.setWin != nil {
			a.setWin.SetTitle(locales.SettingsTitle())
		}
	})
	mygo.App.OnSecondInstance(func([]string, string) { a.openSettings() })
	mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) {
		a.quitting = true
		a.shutdown()
	})
	mygo.App.WhenReady(a.start)
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
