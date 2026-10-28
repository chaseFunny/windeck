package main

import (
	"time"

	"github.com/egoist/mygo/ui"

	"windeck/internal/settings"
	"windeck/internal/switcher"
	"windeck/locales"
)

// After the first launch's intro (intro.go): the switcher, pinned, under
// a guide to its keys.

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
