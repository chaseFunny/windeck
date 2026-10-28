package main

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"windeck/internal/settings"
	"windeck/internal/switcher"
	"windeck/internal/wm"
	"windeck/locales"
)

// The first launch: the logo and the name come in with a burst of
// confetti, drawn by MyGo's poses; then the permissions, if they are not
// given yet, and the switcher itself under a guide to its keys.

const (
	splashW, splashH = 720, 520
	splashFor        = 2600 * time.Millisecond
	splashOut        = 300 * time.Millisecond
)

type splashState struct {
	win   *mygo.Window
	start time.Time
	// shown is when the first frame showed, from which the poses head
	// for where they end: a pose starts at the first one it gets.
	shown  time.Time
	pieces []confetto
}

// confetto is a piece of the burst: where it flies, how it turns, and how
// it looks.
type confetto struct {
	to         ui.Pose
	spin, tilt float32
	w, h       float32
	round      bool
	color      ui.Color
}

// showSplash shows the first launch's animation, then calls then.
func (a *App) showSplash(then func()) {
	pal := &lightPalette
	if a.set.Appearance == "dark" {
		pal = &darkPalette
	}
	a.splash = splashState{start: time.Now(), pieces: newBurst(pal)}
	d := mygo.Screen.DisplayNearestPoint(mygo.Screen.CursorScreenPoint())
	w := mygo.NewWindow(mygo.WindowOptions{
		Title:       "Windeck",
		Panel:       true,
		Frameless:   true,
		Transparent: true,
		Width:       splashW,
		Height:      splashH,
		Content:     ui.View(a.splashView),
	})
	w.SetPosition(d.WorkArea.X+(d.WorkArea.Width-splashW)/2, d.WorkArea.Y+(d.WorkArea.Height-splashH)/2)
	w.ShowInactive()
	a.splash.win = w
	time.AfterFunc(splashFor, func() {
		a.update(func() {
			if a.splash.win != nil {
				a.splash.win.Close()
				a.splash.win = nil
			}
			then()
		})
	})
}

// newBurst makes the confetti: the brand's violets and the success green,
// thrown up and out from behind the logo, falling as they slow.
func newBurst(pal *palette) []confetto {
	colors := []ui.Color{pal.brand, pal.brand, pal.brandHover, darkPalette.brandText, pal.success, pal.warning}
	out := make([]confetto, 70)
	for i := range out {
		dir := -math.Pi/2 + (rand.Float64()-0.5)*math.Pi*1.4
		dist := 140 + rand.Float64()*170
		c := confetto{
			to: ui.Pose{
				X: float32(math.Cos(dir) * dist),
				Y: float32(math.Sin(dir)*dist) + 120 + rand.Float32()*80,
			},
			spin:  (rand.Float32() - 0.5) * 900,
			tilt:  30 + 50*rand.Float32(),
			color: colors[rand.IntN(len(colors))],
			round: rand.IntN(3) == 0,
		}
		c.w, c.h = 10, 10
		if !c.round {
			c.w, c.h = 14, 7+5*rand.Float32()
		}
		out[i] = c
	}
	return out
}

func (a *App) splashView(c *ui.Context) {
	pal := themed(c)
	c.Root().Background(ui.Transparent)
	s := &a.splash
	now := c.Now()
	root := ui.Box(c.Key("stage")).Fill().PassThrough()
	if s.shown.IsZero() {
		if root.Bounds().W > 0 {
			s.shown = now
		}
		c.AnimationFrame()
	}
	in := !s.shown.IsZero()
	leaving := in && now.Sub(s.start) > splashFor-splashOut
	if leaving {
		c.AnimationFrame()
	}
	root.Children(func() {
		// The burst, behind the card: it bursts from the logo's middle.
		burst := ui.Box(c.Key("burst")).Absolute().Left(splashW/2).Top(splashH/2-40).Size(1, 1).PassThrough()
		burst.Children(func() {
			fly := ui.Animation{Duration: 1600 * time.Millisecond, Ease: ui.CubicBezier(0.1, 0.7, 0.3, 1), Delay: 180 * time.Millisecond}
			for i, p := range s.pieces {
				var to, turn ui.Pose
				if in {
					// Flying out, turning, and fading only once they fall.
					to, turn = p.to, ui.Pose{Rotate: p.spin, Fade: 1}
				}
				e := ui.Box(c.Key(i)).Absolute().Left(-p.w/2).Top(-p.h/2).Size(p.w, p.h).PassThrough()
				e.Transform(ui.RotationX(p.tilt))
				e.AnimatePose(to, fly)
				e.Children(func() {
					piece := ui.Box(c).Size(p.w, p.h).Background(p.color).PassThrough()
					if p.round {
						piece.Radius(p.w / 2)
					}
					piece.AnimatePose(turn, ui.Animation{Duration: 2000 * time.Millisecond, Ease: ui.EaseIn, Delay: 180 * time.Millisecond})
				})
			}
		})
		// The card: the logo pops in with a little bounce, the name and
		// the line under it rise after.
		card := ui.Column(c.Key("card")).Absolute().Left(splashW/2-180).Top(splashH/2-150).Width(360).
			Padding(40, 24).Gap(16).AlignItems(ui.Center).Radius(radiusCard).
			Background(pal.bgElevated).Border(1, pal.border).PassThrough()
		cardPose := ui.Pose{Scale: 0.92, Fade: 1}
		if in {
			cardPose = ui.Pose{}
		}
		if leaving {
			cardPose = ui.Pose{Scale: 0.96, Fade: 1}
		}
		card.AnimatePose(cardPose, ui.Animation{Duration: 220 * time.Millisecond, Ease: easeOut})
		card.Children(func() {
			logo := ui.Box(c.Key("logo")).Size(88, 88)
			logoPose := ui.Pose{Scale: 0.4, Fade: 1}
			if in {
				logoPose = ui.Pose{}
			}
			logo.AnimatePose(logoPose, ui.Bouncy)
			logo.Children(func() {
				if b := a.appLogo(); b != nil {
					ui.Image(c, b).Size(88, 88)
				}
			})
			rise := func(delay time.Duration) ui.Animation {
				return ui.Animation{Duration: 360 * time.Millisecond, Ease: easeOut, Delay: delay}
			}
			name := ui.Box(c.Key("name")).Padding(8, 20).Radius(16, 16, 16, 4).Background(pal.brandSubtle)
			namePose := ui.Pose{Y: 12, Fade: 1}
			if in {
				namePose = ui.Pose{}
			}
			name.AnimatePose(namePose, rise(220*time.Millisecond))
			name.Children(func() {
				ui.Text(c, locales.AppName()).Font(fontDisplay).FontSize(28).FontWeight(700).TextColor(pal.text)
			})
			tag := ui.Text(c.Key("tag"), locales.SplashTagline()).FontSize(14).TextColor(pal.textMuted).TextAlign(ui.Center)
			tagPose := ui.Pose{Y: 8, Fade: 1}
			if in {
				tagPose = ui.Pose{}
			}
			tag.AnimatePose(tagPose, rise(360*time.Millisecond))
		})
	})
}

// afterSplash goes on from the animation: to the permissions while
// Accessibility is not given, else to the switcher and its guide.
func (a *App) afterSplash() {
	a.trusted = wm.Trusted(false)
	if !a.trusted {
		a.openOnboarding()
		return
	}
	a.set.Onboarded = true
	a.saveSettings()
	if a.tray != nil {
		a.tray.SetMenu(a.trayMenu())
	}
	a.showGuide()
}

// showGuide opens the switcher, pinned, with its keys over it: the user
// sees what shows as they press the shortcut, and how to work it.
func (a *App) showGuide() {
	if a.deck.Len() == 0 {
		// Nothing listed yet, as right after Accessibility was given.
		a.refreshNow()
		time.AfterFunc(600*time.Millisecond, func() {
			a.update(func() {
				if a.deck.Len() > 0 {
					a.showGuide()
				}
			})
		})
		return
	}
	a.ses.guide = true
	a.openSearch()
	if a.sw.Phase() == switcher.Idle {
		a.ses.guide = false
	}
}

// modKey is the shortcut's modifier, as a key shows it.
func (a *App) modKey() string {
	if a.set.Shortcut == settings.ShortcutCommand {
		return "⌘"
	}
	return "⌥"
}

// guideOverlay dims the panel and shows the keys in a card over it.
func (a *App) guideOverlay(c *ui.Context, pal *palette) {
	scrim := ui.Box(c.Key("guide")).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(panelRadius).
		Background(pal.scrim).Center()
	scrim.AnimatePose(ui.Pose{}, ui.Animation{Duration: 200 * time.Millisecond, Ease: easeOut})
	scrim.Children(func() {
		card := ui.Column(c.Key("guide-card")).Width(460).Padding(24).Gap(16).Radius(radiusCard).
			Background(pal.bgElevated).Border(1, pal.border).AlignItems(ui.Stretch)
		card.Children(func() {
			ui.Column(c).Gap(4).Children(func() {
				ui.Text(c, locales.GuideTitle()).Font(fontDisplay).FontSize(20).FontWeight(700).TextColor(pal.text)
				ui.Text(c, locales.GuideDetail()).FontSize(13).TextColor(pal.textMuted)
			})
			mod := a.modKey()
			ui.Column(c).Gap(10).Children(func() {
				guideRow(c, pal, locales.GuideOpen(), mod, "Tab")
				guideRow(c, pal, locales.GuideRelease(), locales.ShortcutsRelease(mod))
				guideRow(c, pal, locales.GuideQuick(), mod, "Tab")
				guideRow(c, pal, locales.GuideSearch(), locales.KeySpace(), "/")
				if a.set.ActionKeys {
					guideRow(c, pal, locales.GuideActions(), "W", "M", "Q")
				}
				guideRow(c, pal, locales.GuideClose(), "esc", mod+"Tab")
			})
			ui.Text(c, locales.GuideNote()).FontSize(12).TextColor(pal.textMuted)
			ui.Row(c).Children(func() {
				ui.Spacer(c)
				ui.PrimaryButton(c.Key("guide-ok"), locales.GuideStart()).Height(36).OnClick(a.hideGuide)
			})
		})
	})
}

// guideRow is a line of the guide: its keys at the left, what they do.
func guideRow(c *ui.Context, pal *palette, label string, keys ...string) {
	ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Row(c).Width(104).Shrink(0).Gap(4).AlignItems(ui.Center).Children(func() {
			for _, k := range keys {
				kbd(c, pal, k)
			}
		})
		ui.Text(c, label).FontSize(13).TextColor(pal.text).Grow(1).MinWidth(0)
	})
}

// hideGuide leaves the switcher open, to try.
func (a *App) hideGuide() {
	a.ses.guide = false
	a.ses.focusSearch = true
	a.invalidatePanel()
}
