package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"slices"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"windeck/internal/license"
	"windeck/internal/settings"
	"windeck/internal/wm"
	"windeck/locales"
)

// The Settings window, laid out as the design's desktop windows are: a
// 44-point title bar, a 240-point sidebar on the subtle gray, its chosen
// page on the subtle violet, and the page at a reading width of 720, its
// settings in cards, the label at the left and the control at the right.
// Changes apply as they are made; there is no Save.

const (
	pageGeneral = iota
	pageShortcuts
	pagePrivacy
	pageSmart
	pageLicense
	pageAbout
)

const (
	navWidth     = 240
	titleBarTall = 44
	readingWidth = 720
)

// settingsState is the Settings window's own.
type settingsState struct {
	page    int
	applied settings.Settings // the settings as last saved
	scroll  ui.ScrollState

	confirmReset bool
	installed    map[string]installedApp // the privacy list's apps, looked up once
	showAway     bool                    // the privacy list shows the apps not installed

	key        string // the license key typed
	activating bool
	licenseMsg string
	licenseErr bool
}

type installedApp struct {
	name, path string
	ok         bool
}

// Website is Windeck's site, where it is bought.
const Website = "https://windeck.islumi.com"

func (a *App) openSettings() {
	if a.setWin != nil {
		focusApp(a.setWin)
		return
	}
	a.ss.applied = a.set.Clone()
	a.canCapture = wm.CanCapture(false)
	a.trusted = wm.Trusted(false)
	a.setWin = mygo.NewWindow(mygo.WindowOptions{
		Title:                locales.SettingsTitle(),
		Width:                880,
		Height:               620,
		MinWidth:             760,
		MinHeight:            480,
		TitleBarStyle:        mygo.TitleBarHiddenInset,
		TitleBarHeight:       titleBarTall,
		TrafficLightPosition: &mygo.Point{X: 16, Y: 16},
		DisableMaximize:      true,
		StateKey:             "settings",
		Content:              ui.View(a.settingsView),
	})
	a.setWin.OnClosed(func() { a.setWin = nil })
	focusApp(a.setWin)
}

type navPage struct {
	icon  *ui.SVG
	label string
}

func navPages() []navPage {
	return []navPage{
		{gearIcon, locales.PageGeneral()},
		{keyboardIcon, locales.PageShortcuts()},
		{handIcon, locales.PagePrivacy()},
		{sparkleIcon, locales.PageSmart()},
		{keyIcon, locales.PageLicense()},
		{infoIcon, locales.PageAbout()},
	}
}

func (a *App) settingsView(c *ui.Context) {
	pal := themed(c)
	c.Root().Background(pal.bgPage)
	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		a.settingsNav(c, pal)
		ui.Column(c.Key("page")).Grow(1).MinWidth(0).AlignItems(ui.Stretch).Background(pal.bgPage).Children(func() {
			ui.Box(c).Height(titleBarTall).Shrink(0).DragWindow()
			ui.Scroll(c.Key(a.ss.page)).Grow(1).MinHeight(0).TrackScroll(&a.ss.scroll).Padding(0, 32, 32, 32).Children(func() {
				ui.Column(c).MaxWidth(readingWidth).Gap(32).AlignItems(ui.Stretch).Children(func() {
					ui.Text(c, navPages()[a.ss.page].label).FontSize(20).FontWeight(700).TextColor(pal.text).Margin(0, 0, -8, 0)
					switch a.ss.page {
					case pageGeneral:
						a.generalPage(c, pal)
					case pageShortcuts:
						a.shortcutsPage(c, pal)
					case pagePrivacy:
						a.privacyPage(c, pal)
					case pageSmart:
						a.smartPage(c, pal)
					case pageLicense:
						a.licensePage(c, pal)
					case pageAbout:
						a.aboutPage(c, pal)
					}
				})
			})
		})
	})
	// ⌘1 to ⌘6 go to the pages, as the tabs of a settings window do.
	for i, k := range []ui.Key{ui.Key1, ui.Key2, ui.Key3, ui.Key4, ui.Key5, ui.Key6} {
		if c.Shortcut(ui.Cmd, k) {
			a.goPage(i)
		}
	}
	if ui.AlertDialog(c, &a.ss.confirmReset, locales.ResetTitle(), locales.ResetDetail(), locales.ResetConfirm(), locales.CommonCancel()) == 0 {
		a.resetLearned()
	}
	if !reflect.DeepEqual(a.set, a.ss.applied) {
		if !a.set.Previews && a.ss.applied.Previews {
			// Previews off: the pictures kept go too.
			a.pool.Keep(func(uint32) bool { return false })
		}
		a.ss.applied = a.set.Clone()
		a.saveSettings()
		if a.tray != nil {
			a.tray.SetMenu(a.trayMenu())
		}
	}
}

func (a *App) goPage(i int) {
	if a.ss.page != i {
		a.ss.page, a.ss.scroll.Y = i, 0
	}
}

// settingsNav is the sidebar: the pages, the chosen one on the subtle
// violet, as the design's navigation is.
func (a *App) settingsNav(c *ui.Context, pal *palette) {
	nav := ui.Column(c.Key("nav")).Width(navWidth).Shrink(0).Padding(titleBarTall+8, 12, 16, 12).Gap(4).
		Background(pal.bgSubtle).BorderWidth(0, 1, 0, 0).BorderColor(pal.border)
	nav.Children(func() {
		ui.Box(c).Height(titleBarTall).Absolute().Top(0).Left(0).Right(0).DragWindow()
		for i, p := range navPages() {
			b := ui.ButtonBase(c.Key(i)).Height(32).Padding(0, 12).Gap(12).Radius(radiusCtl).Justify(ui.Start).Label(p.label)
			sel := a.ss.page == i
			text, glyph := pal.text, pal.textMuted
			switch {
			case sel:
				b.Background(pal.brandSubtle)
				text, glyph = pal.brandOnSub, pal.brandText
			case b.Pressed():
				b.Background(pal.borderBold)
			case b.Hovered():
				b.Background(pal.bgMuted)
			}
			b.Transition(hoverMotion)
			b.Children(func() {
				ui.Icon(c, p.icon).Size(16, 16).TextColor(glyph)
				w := 400
				if sel {
					w = 500
				}
				ui.Text(c, p.label).FontSize(14).FontWeight(w).TextColor(text).SingleLine()
			})
			page := i
			b.OnClick(func() { a.goPage(page) })
		}
	})
}

// section is a titled card of settings rows.
func section(c *ui.Context, pal *palette, title string, fn func()) {
	ui.Column(c).Gap(12).Children(func() {
		if title != "" {
			ui.Text(c, title).FontSize(16).FontWeight(500).TextColor(pal.text)
		}
		ui.Column(c).Radius(radiusCard).Background(pal.bgPage).Border(1, pal.border).Padding(0, 24).Dividers(1, pal.border).Children(fn)
	})
}

// footnote is the small print under a card.
func footnote(c *ui.Context, pal *palette, s string) {
	ui.Text(c, s).FontSize(12).TextColor(pal.textMuted).Margin(-20, 0, 0, 0)
}

// settingRow is a row of a section: a label and its description at the
// left, the control at the right.
func settingRow(c *ui.Context, pal *palette, label, detail string, control func()) {
	ui.Row(c).Gap(16).MinHeight(56).Padding(12, 0).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, label).FontSize(14).TextColor(pal.text)
			if detail != "" {
				ui.Text(c, detail).FontSize(13).TextColor(pal.textMuted)
			}
		})
		ui.Row(c).Shrink(0).AlignItems(ui.Center).Gap(8).Children(control)
	})
}

// tag is the design's tag: 22 high, 12-point text, on the subtle violet,
// or gray when it says nothing of the brand.
func tag(c *ui.Context, pal *palette, label string, brand bool) {
	bg, fg := pal.bgMuted, pal.textMuted
	if brand {
		bg, fg = pal.brandSubtle, pal.brandOnSub
	}
	ui.Box(c).Height(22).Padding(0, 8).Radius(radiusTag).Center().Shrink(0).Background(bg).Children(func() {
		ui.Text(c, label).FontSize(12).FontWeight(500).TextColor(fg).SingleLine()
	})
}

func proChip(c *ui.Context, pal *palette) { tag(c, pal, "PRO", true) }

func (a *App) generalPage(c *ui.Context, pal *palette) {
	ui.Column(c).Gap(12).Children(func() {
		ui.Text(c, locales.SettingsLayout()).FontSize(16).FontWeight(500).TextColor(pal.text)
		a.layoutCards(c, pal)
	})
	section(c, pal, locales.SettingsLook(), func() {
		settingRow(c, pal, locales.SettingsAppearance(), "", func() {
			opts := []string{"light", "dark", "system"}
			i := max(slices.Index(opts, a.set.Appearance), 0)
			ui.Segmented(c.Key("appearance"), &i, locales.AppearanceLight(), locales.AppearanceDark(), locales.AppearanceSystem())
			a.set.Appearance = opts[i]
		})
		settingRow(c, pal, locales.SettingsLanguage(), "", func() {
			opts := []string{"", "en", "zh-Hans"}
			i := max(slices.Index(opts, a.set.Language), 0)
			ui.Segmented(c.Key("language"), &i, locales.LanguageSystem(), "English", "简体中文")
			a.set.Language = opts[i]
		})
	})
	section(c, pal, locales.SettingsBehavior(), func() {
		settingRow(c, pal, locales.SettingsWindowless(), locales.SettingsWindowlessDetail(), func() {
			ui.Switch(c.Key("windowless"), &a.set.ShowWindowless).Label(locales.SettingsWindowless())
		})
		settingRow(c, pal, locales.SettingsPrimaryDisplay(), locales.SettingsPrimaryDisplayDetail(), func() {
			ui.Switch(c.Key("primary"), &a.set.PrimaryDisplay).Label(locales.SettingsPrimaryDisplay())
		})
		settingRow(c, pal, locales.SettingsOpenAtLogin(), "", func() {
			ui.Switch(c.Key("login"), &a.set.OpenAtLogin).Label(locales.SettingsOpenAtLogin())
		})
	})
}

// shortcutsPage is the keys: the shortcut that opens the switcher and how
// long it is held, then the keys of the switcher, drawn as keys, and the
// window actions, which can be turned off.
func (a *App) shortcutsPage(c *ui.Context, pal *palette) {
	mod := "⌥"
	if a.set.Shortcut == settings.ShortcutCommand {
		mod = "⌘"
	}
	ui.Column(c).Gap(12).Children(func() {
		ui.Text(c, locales.ShortcutsOpen()).FontSize(16).FontWeight(500).TextColor(pal.text)
		a.shortcutCards(c, pal)
	})
	section(c, pal, "", func() {
		settingRow(c, pal, locales.SettingsQuickDelay(), locales.SettingsQuickDelayDetail(), func() {
			v := float64(a.set.QuickDelay)
			ui.Slider(c.Key("delay"), &v, 60, 400).Width(140).Label(locales.SettingsQuickDelay())
			a.set.QuickDelay = int(math.Round(v/10) * 10)
			ui.Text(c, fmt.Sprintf("%d ms", a.set.QuickDelay)).FontSize(13).TextColor(pal.textMuted).Width(52).TextAlign(ui.End)
		})
	})
	section(c, pal, locales.ShortcutsInSwitcher(), func() {
		keyRow(c, pal, locales.ShortcutsNext(), true, "tab")
		keyRow(c, pal, locales.ShortcutsPrevious(), true, "⇧", "tab")
		keyRow(c, pal, locales.ShortcutsMove(), true, "↑", "↓")
		keyRow(c, pal, locales.ShortcutsJump(), true, "←", "→")
		keyRow(c, pal, locales.ShortcutsSwitch(), true, locales.ShortcutsRelease(mod), "|", "↩")
		keyRow(c, pal, locales.ShortcutsSearch(), true, locales.KeySpace(), "|", "/")
		keyRow(c, pal, locales.ShortcutsCancel(), true, "esc")
	})
	footnote(c, pal, locales.ShortcutsSearchNote())
	section(c, pal, locales.ShortcutsActions(), func() {
		settingRow(c, pal, locales.ShortcutsActionKeys(), locales.ShortcutsActionKeysDetail(), func() {
			ui.Switch(c.Key("actions"), &a.set.ActionKeys).Label(locales.ShortcutsActionKeys())
		})
		on := a.set.ActionKeys
		keyRow(c, pal, locales.ShortcutsClose(), on, "W")
		keyRow(c, pal, locales.ShortcutsMinimize(), on, "M")
		keyRow(c, pal, locales.ShortcutsQuit(), on, "Q")
	})
}

// keyRow is a row of what a key does: the words at the left, the keys at
// the right, "|" between two keys that do the same. Off, it is faint.
func keyRow(c *ui.Context, pal *palette, label string, on bool, keys ...string) {
	text := pal.text
	if !on {
		text = pal.textFaint
	}
	ui.Row(c).Gap(16).MinHeight(48).Padding(8, 0).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).FontSize(14).TextColor(text).Grow(1).MinWidth(0)
		ui.Row(c).Shrink(0).Gap(4).AlignItems(ui.Center).Children(func() {
			for _, k := range keys {
				if k == "|" {
					ui.Text(c, locales.ShortcutsOr()).FontSize(12).TextColor(pal.textFaint).Margin(0, 4)
					continue
				}
				box := ui.Box(c).MinWidth(28).Height(28).Padding(0, 8).Radius(radiusTag).Center().
					Background(pal.bgSubtle).Border(1, pal.border)
				box.Children(func() {
					ui.Text(c, k).FontSize(13).FontWeight(500).TextColor(text).SingleLine()
				})
			}
		})
	})
}

// shortcutCards are the two shortcuts to choose from: the keys drawn as
// keys, and a radio button saying which is on. Cards that click: hover
// darkens the border and adds the small shadow, no more.
func (a *App) shortcutCards(c *ui.Context, pal *palette) {
	ui.Row(c).Gap(16).AlignItems(ui.Stretch).Children(func() {
		card := func(mode, mod, title, detail string, recommended bool) {
			sel := a.set.Shortcut == mode
			b := ui.ButtonBase(c.Key(mode)).Grow(1).Basis(0).Padding(24).Radius(radiusCard).
				AlignItems(ui.Stretch).Justify(ui.Start).Label(title).Background(pal.bgPage).Border(1, pal.border)
			switch {
			case sel:
				b.Border(1, pal.brand).Shadow(0, 0, 0, 1, pal.brand)
			case b.Hovered():
				b.Border(1, pal.borderBold).Shadow(0, 1, 2, 0, pal.shadow.Alpha(0.5))
			}
			b.Transition(hoverMotion)
			b.Children(func() {
				ui.Column(c).Grow(1).Gap(4).AlignItems(ui.Start).Children(func() {
					ui.Row(c).Gap(4).AlignItems(ui.Center).AlignSelf(ui.Stretch).Margin(0, 0, 12, 0).Children(func() {
						keycap(c, pal, mod, 32)
						ui.Text(c, "+").FontSize(13).TextColor(pal.textFaint).Margin(0, 4)
						keycap(c, pal, "tab", 48)
						ui.Spacer(c)
						if recommended {
							tag(c, pal, locales.ShortcutRecommended(), true)
						}
						radioMark(c, pal, sel)
					})
					ui.Text(c, title).FontSize(14).FontWeight(500).TextColor(pal.text)
					ui.Text(c, detail).FontSize(13).TextColor(pal.textMuted)
				})
			})
			b.OnClick(func() { a.set.Shortcut = mode })
		}
		card(settings.ShortcutOption, "⌥", locales.ShortcutOptionTitle(), locales.ShortcutOptionDetail(), true)
		card(settings.ShortcutCommand, "⌘", locales.ShortcutCommandTitle(), locales.ShortcutCommandDetail(), false)
	})
}

// layoutCards are the panel's two layouts to choose from, each drawn
// small, as the shortcut's cards are.
func (a *App) layoutCards(c *ui.Context, pal *palette) {
	ui.Row(c).Gap(16).AlignItems(ui.Stretch).Children(func() {
		card := func(mode, title, detail string, draw func()) {
			sel := a.set.Layout == mode
			b := ui.ButtonBase(c.Key("layout:"+mode)).Grow(1).Basis(0).Padding(24).Radius(radiusCard).
				AlignItems(ui.Stretch).Justify(ui.Start).Label(title).Background(pal.bgPage).Border(1, pal.border)
			switch {
			case sel:
				b.Border(1, pal.brand).Shadow(0, 0, 0, 1, pal.brand)
			case b.Hovered():
				b.Border(1, pal.borderBold).Shadow(0, 1, 2, 0, pal.shadow.Alpha(0.5))
			}
			b.Transition(hoverMotion)
			b.Children(func() {
				ui.Column(c).Grow(1).Gap(4).AlignItems(ui.Stretch).Children(func() {
					ui.Row(c).Gap(12).AlignItems(ui.Start).Margin(0, 0, 12, 0).Children(func() {
						ui.Box(c).Grow(1).Height(72).Radius(radiusTag).Background(pal.bgSubtle).Border(1, pal.border).Clip().Children(draw)
						radioMark(c, pal, sel)
					})
					ui.Text(c, title).FontSize(14).FontWeight(500).TextColor(pal.text)
					ui.Text(c, detail).FontSize(13).TextColor(pal.textMuted)
				})
			})
			b.OnClick(func() { a.set.Layout = mode })
		}
		line := func(w float32, chosen bool) {
			col := pal.borderBold
			if chosen {
				col = pal.brand
			}
			ui.Box(c).Width(w).Height(4).Radius(2).Background(col)
		}
		card(settings.LayoutTabs, locales.LayoutTabs(), locales.LayoutTabsDetail(), func() {
			ui.Row(c).Fill().Children(func() {
				ui.Column(c).Width(56).Padding(8).Gap(8).BorderWidth(0, 1, 0, 0).BorderColor(pal.border).Children(func() {
					line(36, false)
					line(28, true)
					line(32, false)
					line(24, false)
				})
				ui.Column(c).Grow(1).Padding(8).AlignItems(ui.Stretch).Children(func() {
					ui.Box(c).Height(54).Radius(4).Background(pal.bgPage).Border(1, pal.border)
				})
			})
		})
		card(settings.LayoutList, locales.LayoutList(), locales.LayoutListDetail(), func() {
			ui.Column(c).Fill().Padding(8).Gap(6).Children(func() {
				for i, w := range []float32{96, 120, 80, 104} {
					ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
						ui.Box(c).Size(8, 8).Radius(2).Background(pal.borderBold)
						line(w, i == 1)
					})
				}
			})
		})
	})
}

// keycap draws a key of the keyboard, as the shortcut's.
func keycap(c *ui.Context, pal *palette, label string, width float32) {
	ui.Box(c).Size(width, 32).Radius(radiusTag).Center().Background(pal.bgSubtle).Border(1, pal.border).Children(func() {
		ui.Text(c, label).FontSize(14).FontWeight(500).TextColor(pal.text)
	})
}

// radioMark is a radio button's look, on or off.
func radioMark(c *ui.Context, pal *palette, on bool) {
	box := ui.Box(c).Size(16, 16).Radius(8).Margin(0, 0, 0, 8).Center()
	if on {
		box.Background(pal.brand).Children(func() {
			ui.Box(c).Size(6, 6).Radius(3).Background(pal.onBrand)
		})
		return
	}
	box.Border(1, pal.borderBold).Background(pal.bgPage)
}

func (a *App) privacyPage(c *ui.Context, pal *palette) {
	// Information is the subtle violet's.
	ui.Row(c).Gap(12).Padding(16).Radius(radiusCtl).Background(pal.brandSubtle).AlignItems(ui.Start).Children(func() {
		ui.Icon(c, handIcon).Size(20, 20).TextColor(pal.brandText)
		ui.Text(c, locales.PrivacyPromise()).FontSize(14).TextColor(pal.brandOnSub).Grow(1)
	})
	section(c, pal, locales.PrivacyPermissions(), func() {
		a.permissionRow(c, pal, locales.PermAccessibility(), locales.PermAccessibilityDetail(), a.trusted, urlAccessibility, true)
		a.permissionRow(c, pal, locales.PermScreen(), locales.PermScreenDetail(), a.canCapture, urlScreen, false)
		settingRow(c, pal, locales.SettingsPreviews(), locales.SettingsPreviewsDetail(), func() {
			ui.Switch(c.Key("previews"), &a.set.Previews).Label(locales.SettingsPreviews())
		})
	})
	section(c, pal, locales.PrivacyExcluded(), func() {
		if len(a.set.Excluded) == 0 {
			ui.Text(c, locales.PrivacyEmpty()).FontSize(13).TextColor(pal.textMuted).Padding(16, 0)
		}
		// The apps not on this Mac fold into one row: they wait there in
		// case they are installed, and say little until then.
		var away []string
		for _, id := range a.excludedSorted() {
			if !a.installedApp(id).ok && !a.ss.showAway {
				away = append(away, id)
				continue
			}
			a.excludedRow(c, pal, id)
		}
		if len(away) > 0 {
			b := ui.ButtonBase(c.Key("away")).MinHeight(48).Justify(ui.Start).Gap(12).Label(locales.PrivacyNotInstalledMore(len(away)))
			b.Children(func() {
				ui.Box(c).Width(24).Center().Children(func() {
					ui.Icon(c, chevronRight).Size(16, 16).TextColor(pal.textMuted)
				})
				col := pal.textMuted
				if b.Hovered() {
					col = pal.text
				}
				ui.Text(c, locales.PrivacyNotInstalledMore(len(away))).FontSize(13).TextColor(col)
			})
			b.OnClick(func() { a.ss.showAway = true })
		}
		// Adding is a menu of the running apps, as macOS's lists add.
		ui.Row(c).MinHeight(56).AlignItems(ui.Center).Children(func() {
			var apps []wm.App
			seen := map[string]bool{}
			for _, ap := range a.apps {
				if ap.BundleID == "" || !ap.Regular || a.set.IsExcluded(ap.BundleID) || seen[ap.BundleID] || ap.PID == int32(osGetpid()) {
					continue
				}
				seen[ap.BundleID] = true
				apps = append(apps, ap)
			}
			slices.SortFunc(apps, func(x, y wm.App) int { return compareFold(x.Name, y.Name) })
			ui.MenuButton(c.Key("add"), locales.PrivacyAdd(), func(m *ui.Menu) {
				for _, ap := range apps {
					if m.Item(ap.Name).Chosen() {
						a.exclude(ap.BundleID)
					}
				}
			}).Disabled(len(apps) == 0)
		})
	})
	footnote(c, pal, locales.PrivacyExcludedDetail())
	section(c, pal, locales.PrivacyData(), func() {
		settingRow(c, pal, locales.PrivacyDataFolder(), tildePath(a.dir), func() {
			ui.Button(c.Key("reveal"), locales.PrivacyReveal()).OnClick(func() {
				os.MkdirAll(a.dir, 0o700)
				mygo.Shell.OpenPath(a.dir)
			})
		})
		settingRow(c, pal, locales.ResetTitle(), locales.ResetDetail(), func() {
			ui.Button(c.Key("reset"), locales.ResetButton()).OnClick(func() { a.ss.confirmReset = true })
		})
	})
}

// exclude puts an app on the privacy list, and forgets its pictures.
func (a *App) exclude(id string) {
	a.set.Excluded = append(slices.Clone(a.set.Excluded), id)
	a.pool.Keep(func(w uint32) bool {
		win, ok := a.deck.Window(wmID(w))
		return !ok || win.BundleID != id
	})
}

// installedApp looks an excluded app up on this Mac, once.
func (a *App) installedApp(id string) installedApp {
	if a.ss.installed == nil {
		a.ss.installed = map[string]installedApp{}
	}
	if ap, ok := a.ss.installed[id]; ok {
		return ap
	}
	var ap installedApp
	for _, r := range a.apps {
		if r.BundleID == id {
			ap = installedApp{r.Name, r.Path, true}
		}
	}
	if !ap.ok {
		ap.name, ap.path, ap.ok = wm.InstalledApp(id)
	}
	a.ss.installed[id] = ap
	return ap
}

// excludedSorted is the privacy list, the apps of this Mac first, by
// name, then the others, kept for when they are installed.
func (a *App) excludedSorted() []string {
	ids := slices.Clone(a.set.Excluded)
	slices.SortStableFunc(ids, func(x, y string) int {
		ax, ay := a.installedApp(x), a.installedApp(y)
		switch {
		case ax.ok != ay.ok:
			if ax.ok {
				return -1
			}
			return 1
		case ax.ok:
			return compareFold(ax.name, ay.name)
		}
		return 0
	})
	return ids
}

func (a *App) excludedRow(c *ui.Context, pal *palette, id string) {
	ap := a.installedApp(id)
	r := ui.Row(c.Key(id)).Gap(12).MinHeight(56).Padding(8, 0).AlignItems(ui.Center)
	hovered := r.Hovered()
	r.Children(func() {
		if ap.ok {
			a.appIconView(c, pal, ap.path, 24)
		} else {
			ui.Box(c).Size(24, 24).Shrink(0).Radius(radiusTag).Background(pal.bgMuted).Center().Children(func() {
				ui.Icon(c, appIcon).Size(16, 16).TextColor(pal.textFaint)
			})
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(1).Children(func() {
			if ap.ok {
				ui.Text(c, ap.name).FontSize(14).TextColor(pal.text).SingleLine().Ellipsis("…")
				ui.Text(c, id).FontSize(12).TextColor(pal.textMuted).SingleLine().Ellipsis("…")
				return
			}
			ui.Text(c, id).FontSize(14).TextColor(pal.textMuted).SingleLine().Ellipsis("…")
			ui.Text(c, locales.PrivacyNotInstalled()).FontSize(12).TextColor(pal.textMuted)
		})
		// The remove button shows as the pointer comes to the row, as
		// macOS's lists show their row actions.
		b := smallButton(c, pal, "remove", minusIcon, locales.PrivacyRemove(), false)
		if !hovered && !b.Focused() {
			b.TextColor(ui.Transparent)
		}
		b.OnClick(func() {
			a.set.Excluded = slices.DeleteFunc(slices.Clone(a.set.Excluded), func(x string) bool { return x == id })
		})
	})
}

const (
	urlAccessibility = "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"
	urlScreen        = "x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture"
)

// statusDot is a permission's state: the success color's check and
// "Allowed".
func statusDot(c *ui.Context, pal *palette, label string) {
	ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
		ui.Icon(c, checkIcon).Size(16, 16).TextColor(pal.success)
		ui.Text(c, label).FontSize(13).TextColor(pal.textMuted)
	})
}

func (a *App) permissionRow(c *ui.Context, pal *palette, label, detail string, granted bool, url string, accessibility bool) {
	settingRow(c, pal, label, detail, func() {
		if granted {
			statusDot(c, pal, locales.PermGranted())
			return
		}
		ui.Button(c.Key(url), locales.PermOpen()).OnClick(func() {
			if accessibility {
				wm.Trusted(true)
			} else {
				wm.CanCapture(true)
			}
			mygo.Shell.OpenExternal(url)
			a.watchPermissions()
		})
	})
}

func wmID(id uint32) wm.WindowID { return wm.WindowID(id) }

func (a *App) smartPage(c *ui.Context, pal *palette) {
	pro := a.pro()
	if !pro {
		ui.Row(c).Gap(12).Padding(16).Radius(radiusCtl).Background(pal.brandSubtle).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, infoIcon).Size(20, 20).TextColor(pal.brandText)
			ui.Text(c, locales.SmartNeedsPro()).FontSize(14).TextColor(pal.brandOnSub).Grow(1)
			ui.PrimaryButton(c.Key("upgrade"), locales.LicenseBuy()).OnClick(func() { a.goPage(pageLicense) })
		})
	}
	section(c, pal, "", func() {
		settingRow(c, pal, locales.SmartPredict(), locales.SmartPredictDetail(), func() {
			proChip(c, pal)
			ui.Switch(c.Key("predict"), &a.set.Predict).Disabled(!pro).Label(locales.SmartPredict())
		})
		settingRow(c, pal, locales.SmartPersonal(), locales.SmartPersonalDetail(), func() {
			proChip(c, pal)
			ui.Switch(c.Key("personal"), &a.set.Personal).Disabled(!pro).Label(locales.SmartPersonal())
		})
	})
	m := a.model.Metrics()
	section(c, pal, locales.SmartStats(), func() {
		ui.Row(c).Padding(24, 0).AlignItems(ui.Stretch).Dividers(1, pal.border).Children(func() {
			// The key numbers are the violet's.
			stat := func(value, label string) {
				ui.Column(c).Grow(1).Basis(0).Gap(4).AlignItems(ui.Center).Children(func() {
					ui.Text(c, value).FontSize(28).FontWeight(700).TextColor(pal.brandText)
					ui.Text(c, label).FontSize(13).TextColor(pal.textMuted).TextAlign(ui.Center)
				})
			}
			today := 0
			if m.Day == time.Now().Format("2006-01-02") {
				today = m.Switches
			}
			stat(itoa(today), locales.StatToday())
			hit, keys := "—", "—"
			if m.Shown > 0 {
				hit = fmt.Sprintf("%d%%", m.TopTwo*100/m.Shown)
				keys = fmt.Sprintf("%.1f", float64(m.Keys)/float64(m.Shown))
			}
			stat(hit, locales.StatTopTwo())
			stat(keys, locales.StatKeys())
		})
	})
	footnote(c, pal, locales.SmartLocalNote())
}

func (a *App) licensePage(c *ui.Context, pal *palette) {
	st := a.lic.Status(time.Now())
	ui.Row(c).Gap(16).Padding(24).Radius(radiusCard).Background(pal.bgPage).Border(1, pal.border).AlignItems(ui.Center).Children(func() {
		if b := a.appLogo(); b != nil {
			ui.Image(c, b).Size(48, 48).Shrink(0)
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "Windeck Pro").FontSize(20).FontWeight(700).TextColor(pal.text)
				switch {
				case st.Licensed:
					tag(c, pal, locales.LicenseActive(), true)
				case st.Trial:
					tag(c, pal, locales.LicenseTrial(st.DaysLeft), true)
				default:
					tag(c, pal, locales.LicenseFree(), false)
				}
			})
			ui.Text(c, locales.LicensePitch()).FontSize(14).TextColor(pal.textMuted)
			if a.lic.Email != "" {
				ui.Text(c, locales.LicenseOwner(a.lic.Email)).FontSize(13).TextColor(pal.textMuted)
			}
		})
	})
	if st.Licensed {
		section(c, pal, "", func() {
			settingRow(c, pal, locales.LicenseDeactivate(), locales.LicenseDeactivateDetail(), func() {
				if a.ss.activating {
					ui.Spinner(c).Size(16, 16)
				}
				ui.Button(c.Key("deactivate"), locales.LicenseDeactivate()).Disabled(a.ss.activating).OnClick(a.deactivateLicense)
			})
		})
	} else {
		section(c, pal, locales.LicenseEnter(), func() {
			ui.Row(c).Gap(8).Padding(24, 0).AlignItems(ui.Center).Children(func() {
				in := ui.TextInput(c.Key("key"), &a.ss.key).Grow(1).Height(36).Placeholder("XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX").Label(locales.LicenseEnter())
				if in.Submitted() && a.ss.key != "" && !a.ss.activating {
					a.activateLicense()
				}
				if a.ss.activating {
					ui.Spinner(c).Size(16, 16)
				}
				ui.PrimaryButton(c.Key("activate"), locales.LicenseActivate()).Height(36).Disabled(a.ss.activating || a.ss.key == "").OnClick(a.activateLicense)
			})
		})
		ui.Row(c).Gap(12).AlignItems(ui.Center).Margin(-20, 0, 0, 0).Children(func() {
			ui.Link(c, locales.LicenseBuy(), Website+"/#pricing").FontSize(14).TextColor(pal.brandText)
			ui.Text(c, locales.LicensePrice()).FontSize(13).TextColor(pal.textMuted)
		})
	}
	if a.ss.licenseMsg != "" {
		// What happened, on the success or the error color's soft
		// background.
		fg, bg, icon := pal.success, pal.success.Alpha(0.12), checkIcon
		if a.ss.licenseErr {
			fg, bg, icon = pal.danger, pal.danger.Alpha(0.12), infoIcon
		}
		ui.Row(c).Gap(12).Padding(12, 16).Radius(radiusCtl).Background(bg).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, icon).Size(20, 20).TextColor(fg)
			ui.Text(c, a.ss.licenseMsg).FontSize(14).TextColor(pal.text).Grow(1)
		})
	}
}

func (a *App) activateLicense() {
	key := a.ss.key
	st := a.lic
	a.ss.activating, a.ss.licenseMsg = true, ""
	host, _ := os.Hostname()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		err := st.Activate(ctx, key, host, time.Now())
		a.setWin.Update(func() {
			a.ss.activating = false
			if err != nil {
				a.ss.licenseErr = true
				if errors.Is(err, license.ErrInvalid) {
					a.ss.licenseMsg = locales.LicenseInvalid()
				} else {
					a.ss.licenseMsg = locales.LicenseNetwork()
				}
				return
			}
			a.lic = st
			a.lic.Save(a.path("license.json"))
			a.ss.key, a.ss.licenseErr, a.ss.licenseMsg = "", false, locales.LicenseThanks()
		})
	}()
}

func (a *App) deactivateLicense() {
	st := a.lic
	a.ss.activating = true
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		err := st.Deactivate(ctx)
		a.setWin.Update(func() {
			a.ss.activating = false
			if err != nil {
				a.ss.licenseErr, a.ss.licenseMsg = true, locales.LicenseNetwork()
				return
			}
			a.lic = st
			a.lic.Save(a.path("license.json"))
			a.ss.licenseErr, a.ss.licenseMsg = false, locales.LicenseDeactivated()
		})
	}()
}

func (a *App) aboutPage(c *ui.Context, pal *palette) {
	ui.Column(c).AlignItems(ui.Center).Gap(8).Padding(32, 0, 16, 0).Children(func() {
		if b := a.appLogo(); b != nil {
			ui.Image(c, b).Size(96, 96).Margin(0, 0, 8, 0)
		}
		ui.Text(c, "Windeck").FontSize(20).FontWeight(700).TextColor(pal.text)
		if v := mygo.App.Version(); v != "" {
			ui.Text(c, locales.AboutVersion(v)).FontSize(13).TextColor(pal.textMuted)
		}
		ui.Text(c, locales.AboutTagline()).FontSize(14).TextColor(pal.text).TextAlign(ui.Center).Margin(8, 0, 0, 0)
		ui.Row(c).Gap(8).Margin(16, 0, 0, 0).Children(func() {
			ui.Button(c.Key("site"), locales.AboutWebsite()).OnClick(func() { mygo.Shell.OpenExternal(Website) })
			ui.Button(c.Key("onboard"), locales.TrayPermissions()).OnClick(a.openOnboarding)
		})
	})
	a.updatesSection(c, pal)
	hzMark(c, pal)
}

// hzMark is the maker's mark the design signs with: the small solid
// bubble, its corner at the bottom left, and "hz", the h in the text's
// color and the z in the violet.
func hzMark(c *ui.Context, pal *palette) {
	ui.Row(c).Gap(8).AlignItems(ui.Center).AlignSelf(ui.Center).Label("hz").Children(func() {
		ui.Box(c).Size(20, 16).Radius(8, 8, 8, 2).Background(pal.brand)
		ui.RichText(c,
			ui.Span{Text: "h", Size: 14, Weight: 700, Color: pal.text, Font: fontDisplay},
			ui.Span{Text: "z", Size: 14, Weight: 700, Color: pal.brandText, Font: fontDisplay},
		)
	})
}
