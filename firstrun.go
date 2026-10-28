package main

import (
	"time"

	liquid "github.com/egoist/mygo/plugins/glass"
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

// guideOverlay dims the panel and shows its keys on a pane of glass over
// it: what is under it shows through, frosted. The keys read as a table,
// in two groups, the keys as keycaps on the left, one thing they do a row.
func (a *App) guideOverlay(c *ui.Context, pal *palette) {
	scrim := ui.Box(c.Key("guide")).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(panelRadius).
		Background(pal.scrim).Center()
	scrim.Children(func() {
		card := ui.Column(c.Key("guide-card")).Width(540).Padding(24).Gap(16).Radius(radiusCard).
			Material(liquid.Glass{Tint: pal.bgElevated.Alpha(0.35)}).Border(1, pal.border).AlignItems(ui.Stretch)
		card.Children(func() {
			ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
				if b := a.appLogo(); b != nil {
					ui.Image(c, b).Size(48, 48).Shrink(0)
				}
				ui.Column(c).Gap(2).Grow(1).MinWidth(0).Children(func() {
					ui.Text(c, locales.GuideTitle()).Font(fontDisplay).FontSize(20).FontWeight(700).TextColor(pal.text)
					ui.Text(c, locales.GuideDetail()).FontSize(13).TextColor(pal.textMuted)
				})
			})
			mod := a.modKey()
			type line struct {
				keys   []string // keycaps; "+" joins, "|" is "or", "~text" is a word
				action string
			}
			group := func(title string, lines []line) {
				ui.Column(c).Gap(8).AlignItems(ui.Stretch).Children(func() {
					ui.Text(c, title).FontSize(12).FontWeight(500).TextColor(pal.textMuted)
					ui.Column(c).Radius(radiusCtl).Background(pal.bgElevated.Alpha(0.7)).Border(1, pal.border).
						Dividers(1, pal.border).AlignItems(ui.Stretch).Children(func() {
						for _, l := range lines {
							ui.Row(c).Height(40).Padding(0, 16).Gap(16).AlignItems(ui.Center).Children(func() {
								ui.Row(c).Width(170).Shrink(0).Gap(6).AlignItems(ui.Center).Children(func() {
									for _, k := range l.keys {
										switch {
										case k == "+":
											ui.Text(c, "+").FontSize(13).TextColor(pal.textFaint)
										case k == "|":
											ui.Text(c, locales.ShortcutsOr()).FontSize(12).TextColor(pal.textFaint)
										case len(k) > 1 && k[0] == '~':
											ui.Text(c, k[1:]).FontSize(13).TextColor(pal.textMuted)
										default:
											guideKey(c, pal, k)
										}
									}
								})
								ui.Text(c, l.action).FontSize(14).TextColor(pal.text).Grow(1).MinWidth(0)
							})
						}
					})
				})
			}
			group(locales.GuideSwitching(), []line{
				{[]string{mod, "+", "Tab"}, locales.GuideOpen()},
				{[]string{"~" + locales.GuideLetGo(), mod}, locales.GuideRelease()},
				{[]string{"~" + locales.GuideTap(), mod, "+", "Tab"}, locales.GuideQuick()},
			})
			panel := []line{{[]string{locales.KeySpace(), "|", "/"}, locales.GuideSearch()}}
			if a.set.ActionKeys {
				panel = append(panel,
					line{[]string{"W"}, locales.HintClose()},
					line{[]string{"M"}, locales.HintMinimize()},
					line{[]string{"Q"}, locales.HintQuit()})
			}
			panel = append(panel, line{[]string{"esc", "|", mod, "+", "Tab"}, locales.GuideClose()})
			group(locales.GuideInPanel(), panel)
			ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
				ui.Text(c, locales.GuideNote()).FontSize(12).TextColor(pal.textMuted).Grow(1).MinWidth(0)
				ui.PrimaryButton(c.Key("guide-ok"), locales.GuideStart()).Height(36).Shrink(0).OnClick(a.hideGuide)
			})
		})
	})
}

// guideKey draws a key as a key looks: raised a little by a thicker edge
// at its bottom.
func guideKey(c *ui.Context, pal *palette, label string) {
	ui.Box(c).MinWidth(28).Height(28).Padding(0, 8).Radius(radiusTag).Center().
		Background(pal.bgPage).BorderWidth(1, 1, 2, 1).BorderColor(pal.borderBold).Children(func() {
		ui.Text(c, label).FontSize(13).FontWeight(500).TextColor(pal.text).SingleLine()
	})
}

// hideGuide leaves the switcher open, to try.
func (a *App) hideGuide() {
	a.ses.guide = false
	a.ses.focusSearch = true
	a.invalidatePanel()
}
