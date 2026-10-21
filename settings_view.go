package main

import (
	"context"
	"errors"
	"fmt"
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

// The Settings window: a sidebar of pages, and the page.

const (
	pageGeneral = iota
	pagePrivacy
	pageSmart
	pageLicense
	pageAbout
)

// settingsState is the Settings window's own.
type settingsState struct {
	page    int
	applied settings.Settings // the settings as last saved
	scroll  ui.ScrollState

	pickApp      string // the app chosen to exclude
	confirmReset bool

	key        string // the license key typed
	activating bool
	licenseMsg string
	licenseErr bool
}

// Website is Windeck's site, where it is bought.
const Website = "https://windeck.app"

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
		Width:                760,
		Height:               560,
		MinWidth:             640,
		MinHeight:            440,
		TitleBarStyle:        mygo.TitleBarHiddenInset,
		TitleBarHeight:       52,
		TrafficLightPosition: &mygo.Point{X: 18, Y: 19},
		DisableMaximize:      true,
		StateKey:             "settings",
		Content:              ui.View(a.settingsView),
	})
	a.setWin.OnClosed(func() { a.setWin = nil })
	focusApp(a.setWin)
}

func (a *App) settingsView(c *ui.Context) {
	pal := paletteFor(c)
	c.Root().Background(pal.windowBack)
	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		nav := ui.Column(c.Key("nav")).Width(196).Shrink(0).Padding(56, 12, 16, 12).Gap(2).Background(pal.sidebar).
			BorderWidth(0, 1, 0, 0).BorderColor(pal.divider)
		nav.Children(func() {
			ui.Box(c).Height(56).Absolute().Top(0).Left(0).Right(0).DragWindow()
			for i, p := range []struct {
				svg   *ui.SVG
				label string
			}{
				{gearIcon, locales.PageGeneral()},
				{lockIcon, locales.PagePrivacy()},
				{sparkleIcon, locales.PageSmart()},
				{keyIcon, locales.PageLicense()},
				{infoIcon, locales.PageAbout()},
			} {
				b := ui.ButtonBase(c.Key(i)).Padding(7, 10).Radius(8).AlignItems(ui.Start).Label(p.label)
				sel := a.ss.page == i
				if sel {
					b.Background(pal.rowSelected).Shadow(0, 1, 2, 0, pal.rowShadow)
				} else if b.Hovered() {
					b.Background(pal.rowHover)
				}
				b.Children(func() {
					ui.Row(c).Gap(9).AlignItems(ui.Center).Children(func() {
						col := pal.textMuted
						if sel {
							col = pal.accent
						}
						ui.Icon(c, p.svg).Size(15, 15).TextColor(col)
						ui.Text(c, p.label).FontSize(13).FontWeight(500).TextColor(pal.text)
					})
				})
				page := i
				b.OnClick(func() { a.ss.page = page; a.ss.scroll.Y = 0 })
			}
		})
		ui.Column(c.Key("page")).Grow(1).MinWidth(0).AlignItems(ui.Stretch).Children(func() {
			ui.Box(c).Height(52).Shrink(0).DragWindow().Padding(0, 28).Justify(ui.Center).Children(func() {
				ui.Text(c, a.pageTitle()).FontSize(15).FontWeight(650).TextColor(pal.text)
			})
			ui.Scroll(c.Key(a.ss.page)).Grow(1).MinHeight(0).TrackScroll(&a.ss.scroll).Padding(4, 28, 28, 28).Gap(18).Children(func() {
				switch a.ss.page {
				case pageGeneral:
					a.generalPage(c, pal)
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
	if ui.AlertDialog(c, &a.ss.confirmReset, locales.ResetTitle(), locales.ResetDetail(), locales.ResetConfirm(), locales.CommonCancel()) == 0 {
		a.resetLearned()
	}
	if !reflect.DeepEqual(a.set, a.ss.applied) {
		a.ss.applied = a.set.Clone()
		a.saveSettings()
		if a.tray != nil {
			a.tray.SetMenu(a.trayMenu())
		}
	}
}

func (a *App) pageTitle() string {
	switch a.ss.page {
	case pagePrivacy:
		return locales.PagePrivacy()
	case pageSmart:
		return locales.PageSmart()
	case pageLicense:
		return locales.PageLicense()
	case pageAbout:
		return locales.PageAbout()
	}
	return locales.PageGeneral()
}

// section is a titled card of settings rows.
func section(c *ui.Context, pal *palette, title string, fn func()) {
	ui.Column(c).Gap(8).Children(func() {
		if title != "" {
			ui.Text(c, title).FontSize(12).FontWeight(600).TextColor(pal.textMuted).Margin(0, 4)
		}
		ui.Column(c).Radius(12).Background(pal.surface).Border(1, pal.divider).Padding(4, 16).Dividers(1, pal.divider).Children(fn)
	})
}

// settingRow is a row of a section: a label and its description at the
// left, the control at the right.
func settingRow(c *ui.Context, pal *palette, label, detail string, control func()) {
	ui.Row(c).Gap(16).Padding(12, 0).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
			ui.Text(c, label).FontSize(13).FontWeight(500).TextColor(pal.text)
			if detail != "" {
				ui.Text(c, detail).FontSize(11.5).TextColor(pal.textMuted)
			}
		})
		ui.Row(c).Shrink(0).AlignItems(ui.Center).Gap(8).Children(control)
	})
}

func proChip(c *ui.Context, pal *palette) {
	ui.Box(c).Padding(1, 6).Radius(6).Background(pal.accentSoft).Children(func() {
		ui.Text(c, "PRO").FontSize(10).FontWeight(700).TextColor(pal.accent)
	})
}

func (a *App) generalPage(c *ui.Context, pal *palette) {
	section(c, pal, locales.SettingsShortcut(), func() {
		ui.Box(c).Padding(14, 0).Children(func() { a.shortcutCards(c, pal) })
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
		settingRow(c, pal, locales.SettingsQuickDelay(), locales.SettingsQuickDelayDetail(), func() {
			v := float64(a.set.QuickDelay)
			ui.StepSlider(c.Key("delay"), &v, 60, 400, 10).Width(150)
			a.set.QuickDelay = int(v)
			ui.Text(c, fmt.Sprintf("%d ms", a.set.QuickDelay)).FontSize(12).TextColor(pal.textMuted).Width(52).TextAlign(ui.End)
		})
		settingRow(c, pal, locales.SettingsWindowless(), locales.SettingsWindowlessDetail(), func() {
			ui.Switch(c.Key("windowless"), &a.set.ShowWindowless)
		})
		settingRow(c, pal, locales.SettingsPrimaryDisplay(), locales.SettingsPrimaryDisplayDetail(), func() {
			ui.Switch(c.Key("primary"), &a.set.PrimaryDisplay)
		})
		settingRow(c, pal, locales.SettingsOpenAtLogin(), "", func() {
			ui.Switch(c.Key("login"), &a.set.OpenAtLogin)
		})
	})
}

// shortcutCards are the two shortcuts to choose from, as large cards.
func (a *App) shortcutCards(c *ui.Context, pal *palette) {
	ui.Row(c).Gap(12).AlignItems(ui.Stretch).Children(func() {
		card := func(mode, keys, title, detail string, recommended bool) {
			sel := a.set.Shortcut == mode
			b := ui.ButtonBase(c.Key(mode)).Grow(1).Basis(0).Padding(16).Radius(12).Label(title)
			b.Background(pal.surface).Border(1, pal.divider)
			if sel {
				b.Border(2, pal.accent).Background(pal.accentSoft)
			} else if b.Hovered() {
				b.Border(1, pal.textFaint)
			}
			b.Children(func() {
				ui.Column(c).Gap(10).Children(func() {
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						ui.Text(c, keys).FontSize(22).FontWeight(650).TextColor(pal.text)
						ui.Spacer(c)
						if recommended {
							ui.Box(c).Padding(2, 7).Radius(6).Background(pal.accent).Children(func() {
								ui.Text(c, locales.ShortcutRecommended()).FontSize(10.5).FontWeight(650).TextColor(ui.RGB(255, 255, 255))
							})
						}
					})
					ui.Text(c, title).FontSize(13.5).FontWeight(600).TextColor(pal.text)
					ui.Text(c, detail).FontSize(12).TextColor(pal.textMuted)
				})
			})
			b.OnClick(func() { a.set.Shortcut = mode })
		}
		card(settings.ShortcutOption, "⌥ Tab", locales.ShortcutOptionTitle(), locales.ShortcutOptionDetail(), true)
		card(settings.ShortcutCommand, "⌘ Tab", locales.ShortcutCommandTitle(), locales.ShortcutCommandDetail(), false)
	})
}

func (a *App) privacyPage(c *ui.Context, pal *palette) {
	ui.Row(c).Gap(10).Padding(12, 14).Radius(12).Background(pal.accentSoft).AlignItems(ui.Center).Children(func() {
		ui.Icon(c, shieldIcon).Size(18, 18).TextColor(pal.accent)
		ui.Text(c, locales.PrivacyPromise()).FontSize(12.5).TextColor(pal.text).Grow(1)
	})
	section(c, pal, locales.PrivacyPermissions(), func() {
		a.permissionRow(c, pal, locales.PermAccessibility(), locales.PermAccessibilityDetail(), a.trusted, urlAccessibility, true)
		a.permissionRow(c, pal, locales.PermScreen(), locales.PermScreenDetail(), a.canCapture, urlScreen, false)
		settingRow(c, pal, locales.SettingsPreviews(), locales.SettingsPreviewsDetail(), func() {
			ui.Switch(c.Key("previews"), &a.set.Previews)
		})
	})
	section(c, pal, locales.PrivacyExcluded(), func() {
		ui.Text(c, locales.PrivacyExcludedDetail()).FontSize(11.5).TextColor(pal.textMuted).Padding(12, 0, 4, 0)
		names := a.appNames()
		for _, id := range a.set.Excluded {
			ui.Row(c.Key(id)).Gap(10).Padding(8, 0).AlignItems(ui.Center).Children(func() {
				name, path := names[id].name, names[id].path
				a.appIconView(c, pal, path, 20)
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					if name != "" {
						ui.Text(c, name).FontSize(13).TextColor(pal.text).SingleLine()
					}
					ui.Text(c, id).FontSize(11).TextColor(pal.textFaint).SingleLine().Ellipsis("…")
				})
				smallButton(c, pal, "remove", closeIcon, locales.PrivacyRemove(), false).OnClick(func() {
					a.set.Excluded = slices.DeleteFunc(slices.Clone(a.set.Excluded), func(x string) bool { return x == id })
				})
			})
		}
		settingRow(c, pal, locales.PrivacyAdd(), "", func() {
			var opts []string
			byName := map[string]string{}
			for _, ap := range a.apps {
				if ap.BundleID == "" || !ap.Regular || a.set.IsExcluded(ap.BundleID) || byName[ap.Name] != "" {
					continue
				}
				opts = append(opts, ap.Name)
				byName[ap.Name] = ap.BundleID
			}
			slices.Sort(opts)
			a.ss.pickApp = ""
			ui.Select(c.Key("add"), &a.ss.pickApp, opts).Width(200).Label(locales.PrivacyAdd())
			if id := byName[a.ss.pickApp]; id != "" {
				a.set.Excluded = append(slices.Clone(a.set.Excluded), id)
				a.pool.Keep(func(w uint32) bool {
					win, ok := a.deck.Window(wmID(w))
					return !ok || win.BundleID != id
				})
			}
		})
	})
	section(c, pal, locales.PrivacyData(), func() {
		settingRow(c, pal, locales.PrivacyDataFolder(), a.dir, func() {
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

const (
	urlAccessibility = "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"
	urlScreen        = "x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture"
)

func (a *App) permissionRow(c *ui.Context, pal *palette, label, detail string, granted bool, url string, accessibility bool) {
	settingRow(c, pal, label, detail, func() {
		if granted {
			ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, checkIcon).Size(14, 14).TextColor(ui.Hex("#2fa84f"))
				ui.Text(c, locales.PermGranted()).FontSize(12.5).TextColor(pal.textMuted)
			})
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

type appName struct{ name, path string }

// appNames are the running apps' names and bundles, by bundle ID.
func (a *App) appNames() map[string]appName {
	out := map[string]appName{}
	for _, ap := range a.apps {
		if ap.BundleID != "" {
			out[ap.BundleID] = appName{ap.Name, ap.Path}
		}
	}
	return out
}

func wmID(id uint32) wm.WindowID { return wm.WindowID(id) }

func (a *App) smartPage(c *ui.Context, pal *palette) {
	pro := a.pro()
	section(c, pal, "", func() {
		settingRow(c, pal, locales.SmartPredict(), locales.SmartPredictDetail(), func() {
			proChip(c, pal)
			ui.Switch(c.Key("predict"), &a.set.Predict).Disabled(!pro)
		})
		settingRow(c, pal, locales.SmartPersonal(), locales.SmartPersonalDetail(), func() {
			proChip(c, pal)
			ui.Switch(c.Key("personal"), &a.set.Personal).Disabled(!pro)
		})
	})
	if !pro {
		ui.Text(c, locales.SmartNeedsPro()).FontSize(12).TextColor(pal.textMuted).Margin(0, 4)
	}
	m := a.model.Metrics()
	section(c, pal, locales.SmartStats(), func() {
		ui.Row(c).Padding(16, 0).Gap(12).Children(func() {
			stat := func(value, label string) {
				ui.Column(c).Grow(1).Basis(0).Gap(4).Children(func() {
					ui.Text(c, value).FontSize(24).FontWeight(650).TextColor(pal.text)
					ui.Text(c, label).FontSize(11.5).TextColor(pal.textMuted)
				})
			}
			today := 0
			if m.Day == time.Now().Format("2006-01-02") {
				today = m.Switches
			}
			stat(itoa(today), locales.StatToday())
			hit, keys := "–", "–"
			if m.Shown > 0 {
				hit = fmt.Sprintf("%d%%", m.TopTwo*100/m.Shown)
				keys = fmt.Sprintf("%.1f", float64(m.Keys)/float64(m.Shown))
			}
			stat(hit, locales.StatTopTwo())
			stat(keys, locales.StatKeys())
		})
	})
	ui.Text(c, locales.SmartLocalNote()).FontSize(11.5).TextColor(pal.textMuted).Margin(0, 4)
}

func (a *App) licensePage(c *ui.Context, pal *palette) {
	st := a.lic.Status(time.Now())
	ui.Column(c).Gap(6).Padding(20).Radius(14).Background(pal.surface).Border(1, pal.divider).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "Windeck Pro").FontSize(18).FontWeight(700).TextColor(pal.text)
			switch {
			case st.Licensed:
				chipText(c, pal, locales.LicenseActive(), pal.accent)
			case st.Trial:
				chipText(c, pal, locales.LicenseTrial(st.DaysLeft), pal.accent)
			default:
				chipText(c, pal, locales.LicenseFree(), pal.textMuted)
			}
		})
		ui.Text(c, locales.LicensePitch()).FontSize(12.5).TextColor(pal.textMuted)
		if a.lic.Email != "" {
			ui.Text(c, locales.LicenseOwner(a.lic.Email)).FontSize(12).TextColor(pal.textMuted)
		}
	})
	if st.Licensed {
		section(c, pal, "", func() {
			settingRow(c, pal, locales.LicenseDeactivate(), locales.LicenseDeactivateDetail(), func() {
				ui.Button(c.Key("deactivate"), locales.LicenseDeactivate()).Disabled(a.ss.activating).OnClick(a.deactivateLicense)
			})
		})
	} else {
		section(c, pal, locales.LicenseEnter(), func() {
			ui.Row(c).Gap(10).Padding(14, 0).AlignItems(ui.Center).Children(func() {
				ui.TextInput(c.Key("key"), &a.ss.key).Grow(1).Placeholder("XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX").Label(locales.LicenseEnter())
				ui.PrimaryButton(c.Key("activate"), locales.LicenseActivate()).Disabled(a.ss.activating || a.ss.key == "").OnClick(a.activateLicense)
			})
		})
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Button(c.Key("buy"), locales.LicenseBuy()).OnClick(func() { mygo.Shell.OpenExternal(Website + "/#pricing") })
			ui.Text(c, locales.LicensePrice()).FontSize(12).TextColor(pal.textMuted)
		})
	}
	if a.ss.licenseMsg != "" {
		col := pal.textMuted
		if a.ss.licenseErr {
			col = pal.danger
		}
		ui.Text(c, a.ss.licenseMsg).FontSize(12.5).TextColor(col).Margin(0, 4)
	}
}

func chipText(c *ui.Context, pal *palette, label string, color ui.Color) {
	ui.Box(c).Padding(2, 8).Radius(7).Background(color.Alpha(0.12)).Children(func() {
		ui.Text(c, label).FontSize(11.5).FontWeight(600).TextColor(color)
	})
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
	ui.Column(c).AlignItems(ui.Center).Gap(10).Padding(28, 0).Children(func() {
		if b := a.appLogo(); b != nil {
			ui.Image(c, b).Size(96, 96)
		}
		ui.Text(c, "Windeck · 窗签").FontSize(22).FontWeight(700).TextColor(pal.text)
		ui.Text(c, locales.AboutVersion(mygo.App.Version())).FontSize(12.5).TextColor(pal.textMuted)
		ui.Text(c, locales.AboutTagline()).FontSize(14).TextColor(pal.text).TextAlign(ui.Center).Margin(8, 0, 0, 0)
		ui.Row(c).Gap(10).Margin(14, 0, 0, 0).Children(func() {
			ui.Button(c.Key("site"), locales.AboutWebsite()).OnClick(func() { mygo.Shell.OpenExternal(Website) })
			ui.Button(c.Key("onboard"), locales.TrayPermissions()).OnClick(a.openOnboarding)
		})
	})
}
