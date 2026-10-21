package main

import (
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"windeck/internal/wm"
	"windeck/locales"
)

// The first launch: the permissions Windeck needs, then the shortcut.

type onboardingState struct {
	step     int
	watching bool
	// captureAsked tells that the user went to allow screen recording,
	// which takes effect once Windeck starts again.
	captureAsked bool
}

func (a *App) openOnboarding() {
	if a.onboardWin != nil {
		focusApp(a.onboardWin)
		return
	}
	a.ob.step = 0
	a.onboardWin = mygo.NewWindow(mygo.WindowOptions{
		Title:                locales.OnboardTitle(),
		Width:                640,
		Height:               600,
		DisableResize:        true,
		DisableMaximize:      true,
		TitleBarStyle:        mygo.TitleBarHiddenInset,
		TitleBarHeight:       40,
		TrafficLightPosition: &mygo.Point{X: 16, Y: 14},
		Content:              ui.View(a.onboardingView),
	})
	a.onboardWin.Center()
	a.onboardWin.OnClosed(func() { a.onboardWin = nil })
	focusApp(a.onboardWin)
	a.watchPermissions()
}

// watchPermissions checks the permissions every second while the user is
// in System Settings, as macOS tells no one when they change: only while
// the onboarding or Settings window is open.
func (a *App) watchPermissions() {
	if a.ob.watching {
		return
	}
	a.ob.watching = true
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for range t.C {
			trusted, capture := wm.Trusted(false), wm.CanCapture(false)
			done := false
			a.update(func() {
				if trusted && !a.trusted {
					a.trusted = true
					wm.ObserveApps()
					a.startTap()
					a.refreshNow()
				}
				if capture != a.canCapture {
					a.canCapture = capture
				}
				a.redraw()
				if a.onboardWin == nil && a.setWin == nil {
					a.ob.watching = false
					done = true
				}
			})
			if done {
				return
			}
		}
	}()
}

func (a *App) onboardingView(c *ui.Context) {
	pal := paletteFor(c)
	c.Root().Background(pal.windowBack)
	ui.Column(c).Fill().Padding(52, 40, 28, 40).Gap(18).AlignItems(ui.Stretch).Children(func() {
		ui.Box(c).Height(40).Absolute().Top(0).Left(0).Right(0).DragWindow()
		ui.Column(c).AlignItems(ui.Center).Gap(8).Children(func() {
			if b := a.appLogo(); b != nil {
				ui.Image(c, b).Size(64, 64)
			}
			title, sub := locales.OnboardWelcome(), locales.OnboardWelcomeDetail()
			if a.ob.step == 1 {
				title, sub = locales.OnboardShortcut(), locales.OnboardShortcutDetail()
			}
			ui.Text(c, title).FontSize(22).FontWeight(700).TextColor(pal.text).TextAlign(ui.Center)
			ui.Text(c, sub).FontSize(13).TextColor(pal.textMuted).TextAlign(ui.Center).MaxWidth(470)
		})
		ui.Column(c).Grow(1).Gap(12).Children(func() {
			if a.ob.step == 0 {
				a.permissionCard(c, pal, shieldIcon, locales.PermAccessibility(), locales.PermAccessibilityWhy(), true, a.trusted, urlAccessibility)
				a.permissionCard(c, pal, screenIcon, locales.PermScreen(), locales.PermScreenWhy(), false, a.canCapture, urlScreen)
				if a.ob.captureAsked && !a.canCapture {
					ui.Text(c, locales.PermScreenRestart()).FontSize(12).TextColor(pal.textMuted)
				}
				return
			}
			a.shortcutCards(c, pal)
			ui.Text(c, locales.OnboardShortcutNote()).FontSize(12).TextColor(pal.textMuted).Margin(6, 2)
		})
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			for i := range 2 {
				dot := ui.Box(c).Size(7, 7).Radius(4).Background(pal.guide)
				if i == a.ob.step {
					dot.Background(pal.accent).Width(18)
				}
			}
			ui.Spacer(c)
			if a.ob.step == 0 {
				ui.PrimaryButton(c.Key("next"), locales.OnboardNext()).Disabled(!a.trusted).OnClick(func() { a.ob.step = 1 })
				return
			}
			ui.Button(c.Key("back"), locales.OnboardBack()).OnClick(func() { a.ob.step = 0 })
			ui.PrimaryButton(c.Key("done"), locales.OnboardDone()).OnClick(a.finishOnboarding)
		})
	})
}

func (a *App) finishOnboarding() {
	a.set.Onboarded = true
	a.saveSettings()
	if a.tray != nil {
		a.tray.SetMenu(a.trayMenu())
	}
	if a.onboardWin != nil {
		a.onboardWin.Close()
	}
}

func (a *App) permissionCard(c *ui.Context, pal *palette, svg *ui.SVG, title, why string, required, granted bool, url string) {
	card := ui.Row(c.Key(url)).Gap(14).Padding(16).Radius(14).Background(pal.surface).Border(1, pal.divider).AlignItems(ui.Center)
	card.Children(func() {
		ui.Box(c).Size(40, 40).Radius(10).Background(pal.accentSoft).Center().Children(func() {
			ui.Icon(c, svg).Size(20, 20).TextColor(pal.accent)
		})
		ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, title).FontSize(14).FontWeight(600).TextColor(pal.text)
				label := locales.PermOptional()
				if required {
					label = locales.PermRequired()
				}
				chipText(c, pal, label, pal.textMuted)
			})
			ui.Text(c, why).FontSize(12).TextColor(pal.textMuted)
		})
		if granted {
			ui.Row(c).Gap(5).Shrink(0).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, checkIcon).Size(15, 15).TextColor(ui.Hex("#2fa84f"))
				ui.Text(c, locales.PermGranted()).FontSize(12.5).TextColor(pal.textMuted)
			})
			return
		}
		ui.Button(c.Key("open"), locales.PermAllow()).OnClick(func() {
			if required {
				wm.Trusted(true)
			} else {
				a.ob.captureAsked = true
				wm.CanCapture(true)
			}
			mygo.Shell.OpenExternal(url)
		})
	})
	if !required && a.ob.captureAsked && !granted {
		ui.Row(c).Children(func() {
			ui.Button(c.Key("relaunch"), locales.PermRelaunch()).OnClick(func() {
				a.shutdown()
				mygo.App.Relaunch()
			})
		})
	}
}
