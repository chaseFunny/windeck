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
		Width:                680,
		Height:               640,
		DisableResize:        true,
		DisableMaximize:      true,
		TitleBarStyle:        mygo.TitleBarHiddenInset,
		TitleBarHeight:       44,
		TrafficLightPosition: &mygo.Point{X: 16, Y: 16},
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
	pal := themed(c)
	c.Root().Background(pal.bgPage)
	ui.Column(c).Fill().AlignItems(ui.Stretch).Children(func() {
		ui.Box(c).Height(44).Absolute().Top(0).Left(0).Right(0).DragWindow()
		ui.Column(c).Grow(1).Padding(64, 48, 0, 48).Gap(32).AlignItems(ui.Stretch).Children(func() {
			ui.Column(c).AlignItems(ui.Center).Gap(16).Children(func() {
				if b := a.appLogo(); b != nil {
					ui.Image(c, b).Size(64, 64)
				}
				title, sub := locales.OnboardWelcome(), locales.OnboardWelcomeDetail()
				if a.ob.step == 1 {
					title, sub = locales.OnboardShortcut(), locales.OnboardShortcutDetail()
				}
				// The screen's one big bubble holds its one sentence that
				// matters, in the rounded face of welcomes.
				ui.Box(c).Padding(12, 24).Radius(16, 16, 16, 4).Background(pal.brandSubtle).Children(func() {
					ui.Text(c, title).Font(fontDisplay).FontSize(28).FontWeight(700).TextColor(pal.text).TextAlign(ui.Center)
				})
				ui.Text(c, sub).FontSize(14).TextColor(pal.textMuted).TextAlign(ui.Center).MaxWidth(440)
			})
			ui.Column(c).Gap(12).Children(func() {
				if a.ob.step == 0 {
					ui.Column(c).Radius(radiusCard).Background(pal.bgPage).Border(1, pal.border).Padding(0, 24).Dividers(1, pal.border).Children(func() {
						a.permissionCard(c, pal, accessIcon, locales.PermAccessibility(), locales.PermAccessibilityWhy(), true, a.trusted, urlAccessibility)
						a.permissionCard(c, pal, screenIcon, locales.PermScreen(), locales.PermScreenWhy(), false, a.canCapture, urlScreen)
					})
					if !a.trusted {
						ui.Text(c, locales.PermStaleNote()).FontSize(13).TextColor(pal.textMuted).AlignSelf(ui.Center).TextAlign(ui.Center).MaxWidth(520)
					}
					if a.ob.captureAsked && !a.canCapture {
						ui.Row(c).Gap(12).Padding(12, 16).Radius(radiusCtl).Background(pal.brandSubtle).AlignItems(ui.Center).Children(func() {
							ui.Icon(c, infoIcon).Size(20, 20).TextColor(pal.brandText)
							ui.Text(c, locales.PermScreenRestart()).FontSize(14).TextColor(pal.brandOnSub).Grow(1)
							ui.Button(c.Key("relaunch"), locales.PermRelaunch()).OnClick(func() {
								a.shutdown()
								mygo.App.Relaunch()
							})
						})
					}
					return
				}
				a.shortcutCards(c, pal)
				ui.Text(c, locales.OnboardShortcutNote()).FontSize(13).TextColor(pal.textMuted).AlignSelf(ui.Center).Margin(8, 0, 0, 0)
			})
		})
		// The footer: where you are, and the way on. One primary button
		// on the screen: Allow until Accessibility is, Continue after.
		ui.Row(c).Height(64).Shrink(0).Padding(0, 24).Gap(8).AlignItems(ui.Center).BorderWidth(1, 0, 0, 0).BorderColor(pal.border).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Label(locales.OnboardStep(a.ob.step+1, 2)).Children(func() {
				for i := range 2 {
					dot := ui.Box(c).Size(8, 8).Radius(4).Background(pal.borderBold)
					if i == a.ob.step {
						dot.Background(pal.brand).Width(24)
					}
					dot.Transition(ui.ElementTransition{Duration: 200 * time.Millisecond, Ease: easeOut})
				}
			})
			ui.Spacer(c)
			if a.ob.step == 0 {
				if a.trusted {
					ui.PrimaryButton(c.Key("next"), locales.OnboardNext()).Height(36).OnClick(func() { a.ob.step = 1 })
				} else {
					ui.Button(c.Key("next"), locales.OnboardNext()).Height(36).Disabled(true)
				}
				return
			}
			ui.Button(c.Key("back"), locales.OnboardBack()).Height(36).OnClick(func() { a.ob.step = 0 })
			ui.PrimaryButton(c.Key("done"), locales.OnboardDone()).Height(36).OnClick(a.finishOnboarding)
		})
	})
	if c.Shortcut(0, ui.KeyEnter) {
		switch {
		case a.ob.step == 0 && a.trusted:
			a.ob.step = 1
		case a.ob.step == 1:
			a.finishOnboarding()
		}
	}
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
	// Then the switcher, as it shows, with its keys over it.
	a.showGuide()
}

// askPermission clears what macOS kept of an earlier build's grant, then
// asks again and opens the System Settings pane: the list then holds one
// Windeck, and the switch the user turns on is this build's.
func (a *App) askPermission(accessibility bool, url string) {
	if accessibility {
		wm.ResetPermission(wm.ServiceAccessibility)
		wm.Trusted(true)
	} else {
		wm.ResetPermission(wm.ServiceScreen)
		wm.CanCapture(true)
	}
	mygo.Shell.OpenExternal(url)
	a.watchPermissions()
}

// permissionCard is a permission's row: what it is for, and its button,
// or the check once it is allowed.
func (a *App) permissionCard(c *ui.Context, pal *palette, svg *ui.SVG, title, why string, required, granted bool, url string) {
	ui.Row(c.Key(url)).Gap(16).Padding(24, 0).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Size(40, 40).Shrink(0).Radius(radiusCtl).Background(pal.bgSubtle).Center().Children(func() {
			ui.Icon(c, svg).Size(24, 24).TextColor(pal.textMuted)
		})
		ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, title).FontSize(16).FontWeight(500).TextColor(pal.text)
				if required {
					tag(c, pal, locales.PermRequired(), true)
				} else {
					tag(c, pal, locales.PermOptional(), false)
				}
			})
			ui.Text(c, why).FontSize(13).TextColor(pal.textMuted)
		})
		if granted {
			statusDot(c, pal, locales.PermGranted())
			return
		}
		var b ui.Element
		if required {
			b = ui.PrimaryButton(c.Key("open"), locales.PermAllow()).Height(36)
		} else {
			b = ui.Button(c.Key("open"), locales.PermAllow()).Height(36)
		}
		b.OnClick(func() {
			if !required {
				a.ob.captureAsked = true
			}
			a.askPermission(required, url)
		})
	})
}
