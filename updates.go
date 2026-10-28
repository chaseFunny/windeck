package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"windeck/locales"
)

// Updates: Windeck checks its GitHub releases once a day, and from the
// menu bar. mygo.Updater finds the new version, checks its signature
// against the key built into the app, and replaces the app; Windeck's
// own window, in the design's look, asks first, unless the user chose to
// install updates without asking, which then run from the next launch.

// The update window's states.
const (
	updChecking = iota
	updAvailable
	updDownloading
	updReady
	updLatest
	updFailed
	updInstallFailed
	updDisabled
)

const (
	updateInterval = 24 * time.Hour
	updateRetry    = time.Hour // after a check that failed, as offline
	updateWidth    = 520
)

// updateState is the updates', kept in updates.json for what outlives a
// launch.
type updateState struct {
	LastCheck time.Time `json:"lastCheck"`
	Skipped   string    `json:"skipped"` // a version the user skipped

	enabled bool   // this copy can update itself
	ready   string // a version installed, which runs from the next launch
	busy    bool   // a check or a download under way
	user    bool   // the user asked for the check: the window shows its outcome
	phase   int
	found   *mygo.Update
	errText string
	got     int64
	total   int64
	cancel  context.CancelFunc
	timer   *time.Timer
	win     *mygo.Window
	notes   ui.ScrollState
}

// startUpdates finds whether this copy can update itself, off the main
// thread as it writes next to the app to know, then plans the checks.
func (a *App) startUpdates() {
	if b, err := os.ReadFile(a.path("updates.json")); err == nil {
		json.Unmarshal(b, &a.upd)
	}
	go func() {
		on := mygo.Updater.Enabled()
		a.update(func() {
			a.upd.enabled = on
			a.scheduleUpdateCheck()
			if a.setWin != nil {
				a.setWin.Invalidate()
			}
		})
	}()
}

func (a *App) saveUpdates() {
	b, err := json.Marshal(&a.upd)
	if err != nil {
		return
	}
	os.MkdirAll(a.dir, 0o700)
	os.WriteFile(a.path("updates.json"), b, 0o600)
}

// scheduleUpdateCheck plans the next check in the background: a day
// after the last, and soon after launch when that was longer ago. It is
// one timer, not a loop.
func (a *App) scheduleUpdateCheck() {
	if a.upd.timer != nil {
		a.upd.timer.Stop()
		a.upd.timer = nil
	}
	if !a.upd.enabled || !a.set.AutoUpdate {
		return
	}
	a.updateCheckIn(max(updateInterval-time.Since(a.upd.LastCheck), 10*time.Second))
}

func (a *App) updateCheckIn(d time.Duration) {
	if a.upd.timer != nil {
		a.upd.timer.Stop()
	}
	a.upd.timer = time.AfterFunc(d, func() {
		a.update(func() {
			a.upd.timer = nil
			a.checkUpdates(false)
		})
	})
}

// checkForUpdates checks as the user asked, from the menu bar or
// Settings: the window shows at once and says how it went.
func (a *App) checkForUpdates() {
	if !a.upd.enabled {
		a.upd.phase = updDisabled
		a.openUpdateWindow(true)
		return
	}
	if a.upd.phase == updDownloading && a.upd.busy {
		a.openUpdateWindow(true)
		return
	}
	if a.upd.ready != "" {
		a.upd.phase = updReady
		a.openUpdateWindow(true)
		return
	}
	a.checkUpdates(true)
}

// checkUpdates asks GitHub for a newer version, on another goroutine.
func (a *App) checkUpdates(user bool) {
	if user {
		a.upd.user = true
		a.upd.phase = updChecking
		a.openUpdateWindow(true)
	}
	if a.upd.busy {
		return
	}
	a.upd.busy = true
	go func() {
		ctx, cancel := timeoutContext(30 * time.Second)
		up, err := mygo.Updater.Check(ctx)
		cancel()
		a.update(func() { a.checked(up, err) })
	}()
}

// checked takes a check's outcome.
func (a *App) checked(up *mygo.Update, err error) {
	a.upd.busy = false
	user := a.upd.user
	a.upd.user = false
	if err != nil {
		if user {
			a.upd.phase = updFailed
			a.updateWindowChanged()
		}
		if a.set.AutoUpdate {
			a.updateCheckIn(updateRetry)
		}
		return
	}
	a.upd.LastCheck = time.Now()
	a.saveUpdates()
	a.scheduleUpdateCheck()
	if a.setWin != nil {
		a.setWin.Invalidate()
	}
	switch {
	case up == nil:
		if user {
			a.upd.phase = updLatest
			a.updateWindowChanged()
		}
	case up.Version == a.upd.ready:
		if user {
			a.upd.phase = updReady
			a.updateWindowChanged()
		}
	case !user && up.Version == a.upd.Skipped:
		// Skipped: the next version shows, not this one.
	case !user && a.set.AutoInstall:
		a.upd.found = up
		a.installUpdate(false)
	default:
		a.upd.found = up
		a.upd.phase = updAvailable
		a.upd.notes.Y = 0
		a.openUpdateWindow(user)
		a.updateWindowChanged()
	}
}

// installUpdate downloads the version found and puts it in the app's
// place; asked by the user, Windeck relaunches into it.
func (a *App) installUpdate(user bool) {
	up := a.upd.found
	if up == nil || a.upd.busy {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.upd.busy, a.upd.cancel = true, cancel
	a.upd.got, a.upd.total = 0, 0
	a.upd.phase = updDownloading
	a.updateWindowChanged()
	go func() {
		err := up.Install(ctx, func(got, total int64) {
			a.update(func() {
				a.upd.got, a.upd.total = got, total
				if a.upd.win != nil {
					a.upd.win.Invalidate()
				}
			})
		})
		a.update(func() {
			a.upd.busy, a.upd.cancel = false, nil
			cancel()
			switch {
			case errors.Is(err, context.Canceled):
				a.upd.phase = updAvailable
			case err != nil:
				a.upd.phase = updInstallFailed
				a.upd.errText = err.Error()
				if !user && a.upd.win == nil {
					return // tried again tomorrow
				}
			default:
				a.upd.ready = up.Version
				a.upd.phase = updReady
				if a.tray != nil {
					a.tray.SetMenu(a.trayMenu())
				}
				if user {
					mygo.App.Relaunch()
					return
				}
			}
			a.updateWindowChanged()
		})
	}()
}

// openUpdateWindow shows the update window: in front when the user asked,
// else behind what they work in, for when they look.
func (a *App) openUpdateWindow(front bool) {
	w, h := a.updateWindowSize()
	if a.upd.win != nil {
		a.upd.win.SetContentSize(w, h)
		if front {
			focusApp(a.upd.win)
		}
		a.upd.win.Invalidate()
		return
	}
	a.upd.win = mygo.NewWindow(mygo.WindowOptions{
		Title:                locales.UpdatesTitle(),
		Width:                w,
		Height:               h,
		Hidden:               true,
		DisableResize:        true,
		DisableMaximize:      true,
		TitleBarStyle:        mygo.TitleBarHiddenInset,
		TitleBarHeight:       titleBarTall,
		TrafficLightPosition: &mygo.Point{X: 16, Y: 16},
		Vibrancy:             mygo.VibrancyUnderWindow,
		Content:              ui.View(a.updateView),
	})
	a.upd.win.Center()
	a.upd.win.OnClosed(func() {
		a.upd.win = nil
		// Closing the window stops a download the user started.
		if a.upd.cancel != nil {
			a.upd.cancel()
		}
	})
	if front {
		focusApp(a.upd.win)
	} else {
		a.upd.win.ShowInactive()
	}
}

// updateWindowChanged sizes the window to its state and draws it.
func (a *App) updateWindowChanged() {
	if a.upd.win == nil {
		return
	}
	a.upd.win.SetContentSize(a.updateWindowSize())
	a.upd.win.Invalidate()
}

func (a *App) updateWindowSize() (int, int) {
	switch a.upd.phase {
	case updAvailable:
		return updateWidth, 460
	case updDownloading:
		return updateWidth, 236
	case updInstallFailed:
		return updateWidth, 240
	}
	return updateWidth, 200
}

func (a *App) closeUpdateWindow() {
	if a.upd.win != nil {
		a.upd.win.Close()
	}
}

// updateView is the update window: the app's icon and what is going on,
// the release notes of a new version, the buttons at the bottom right,
// the one that goes on the primary.
func (a *App) updateView(c *ui.Context) {
	pal := themed(c)
	c.Root().Background(glass(c, pal, pal.bgPage))
	u := &a.upd
	cur := mygo.App.Version()
	version := ""
	if u.found != nil {
		version = u.found.Version
	}
	if u.phase == updReady {
		version = u.ready
	}
	var title, detail string
	switch u.phase {
	case updChecking:
		title = locales.UpdatesChecking()
	case updAvailable:
		title = locales.UpdatesAvailable(version)
		if cur != "" {
			detail = locales.UpdatesAvailableDetail(cur)
		}
	case updDownloading:
		title = locales.UpdatesDownloading(version)
	case updReady:
		title, detail = locales.UpdatesReady(version), locales.UpdatesReadyDetail()
	case updLatest:
		title, detail = locales.UpdatesLatest(), locales.UpdatesLatestDetail(cur)
	case updFailed:
		title, detail = locales.UpdatesFailed(), locales.UpdatesFailedDetail()
	case updInstallFailed:
		title, detail = locales.UpdatesInstallFailed(), u.errText
	case updDisabled:
		title, detail = locales.UpdatesTitle(), locales.UpdatesDisabled()
	}
	ui.Column(c).Fill().Padding(titleBarTall, 32, 24, 32).Gap(16).AlignItems(ui.Stretch).Children(func() {
		ui.Box(c).Height(titleBarTall).Absolute().Top(0).Left(0).Right(0).DragWindow()
		ui.Row(c).Gap(16).AlignItems(ui.Center).Shrink(0).Children(func() {
			if b := a.appLogo(); b != nil {
				ui.Image(c, b).Size(56, 56)
			}
			ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
				ui.Text(c, title).FontSize(16).FontWeight(700).TextColor(pal.text)
				if detail != "" {
					ui.Text(c, detail).FontSize(13).TextColor(pal.textMuted).MaxLines(3).Ellipsis("…")
				}
			})
		})
		switch u.phase {
		case updChecking:
			ui.Progress(c, -1)
		case updAvailable:
			ui.Text(c, locales.UpdatesNotes()).FontSize(13).FontWeight(500).TextColor(pal.text).Margin(4, 0, -8, 0)
			ui.Scroll(c.Key("notes")).Grow(1).MinHeight(0).TrackScroll(&u.notes).Radius(radiusCtl).
				Background(pal.bgSubtle).Border(1, pal.border).Padding(16).Children(func() {
				notesView(c, pal, u.found.Notes)
			})
			ui.Text(c, locales.UpdatesPermissionNote()).FontSize(12).TextColor(pal.textMuted)
		case updDownloading:
			p := -1.0
			if u.total > 0 {
				p = float64(u.got) / float64(u.total)
			}
			ui.Progress(c, p)
			if u.total > 0 {
				ui.Text(c, megabytes(u.got)+" / "+megabytes(u.total)).FontSize(12).TextColor(pal.textMuted).Margin(-8, 0, 0, 0)
			}
		}
		if u.phase != updAvailable {
			ui.Spacer(c) // the notes take the room there
		}
		ui.Row(c).Gap(8).Shrink(0).AlignItems(ui.Center).Children(func() { a.updateButtons(c) })
	})
	if c.Shortcut(0, ui.KeyEscape) && u.phase != updDownloading {
		a.closeUpdateWindow()
	}
}

// updateButtons are the window's buttons for its state: one primary.
func (a *App) updateButtons(c *ui.Context) {
	u := &a.upd
	ok := func() {
		ui.Spacer(c)
		ui.PrimaryButton(c.Key("ok"), locales.CommonOK()).OnClick(a.closeUpdateWindow)
	}
	switch u.phase {
	case updChecking:
		ui.Spacer(c)
		ui.Button(c.Key("cancel"), locales.CommonCancel()).OnClick(a.closeUpdateWindow)
	case updAvailable:
		ui.Button(c.Key("skip"), locales.UpdatesSkip()).OnClick(func() {
			u.Skipped = u.found.Version
			a.saveUpdates()
			a.closeUpdateWindow()
		})
		ui.Spacer(c)
		ui.Button(c.Key("later"), locales.UpdatesLater()).OnClick(a.closeUpdateWindow)
		ui.PrimaryButton(c.Key("install"), locales.UpdatesInstallButton()).OnClick(func() { a.installUpdate(true) })
	case updDownloading:
		ui.Spacer(c)
		ui.Button(c.Key("cancel"), locales.CommonCancel()).OnClick(func() {
			if u.cancel != nil {
				u.cancel()
			}
		})
	case updReady:
		ui.Spacer(c)
		ui.Button(c.Key("later"), locales.UpdatesLater()).OnClick(a.closeUpdateWindow)
		ui.PrimaryButton(c.Key("relaunch"), locales.UpdatesRelaunch()).OnClick(mygo.App.Relaunch)
	case updInstallFailed, updDisabled:
		ui.Button(c.Key("site"), locales.AboutWebsite()).OnClick(func() { mygo.Shell.OpenExternal(Website) })
		ok()
	default:
		ok()
	}
}

// notesView shows release notes, the version's section of the changelog
// in Markdown: its headings, list items and paragraphs, as plain text.
func notesView(c *ui.Context, pal *palette, md string) {
	plain := strings.NewReplacer("**", "", "`", "").Replace
	ui.Column(c).Gap(6).AlignItems(ui.Stretch).Children(func() {
		for i, line := range strings.Split(md, "\n") {
			line = strings.TrimSpace(line)
			switch {
			case line == "":
			case strings.HasPrefix(line, "#"):
				t := plain(strings.TrimSpace(strings.TrimLeft(line, "#")))
				ui.Text(c.Key(i), t).FontSize(13).FontWeight(700).TextColor(pal.text).Margin(6, 0, 0, 0)
			case strings.HasPrefix(line, "- "), strings.HasPrefix(line, "* "):
				ui.Row(c.Key(i)).Gap(8).AlignItems(ui.Start).Children(func() {
					ui.Text(c, "•").FontSize(13).TextColor(pal.textMuted)
					ui.Text(c, plain(line[2:])).FontSize(13).TextColor(pal.text).Grow(1).MinWidth(0)
				})
			default:
				ui.Text(c.Key(i), plain(line)).FontSize(13).TextColor(pal.text)
			}
		}
	})
}

func megabytes(n int64) string { return fmt.Sprintf("%.1f MB", float64(n)/1e6) }

// updatesSection is the About page's: the two choices and the last check.
func (a *App) updatesSection(c *ui.Context, pal *palette) {
	section(c, pal, locales.SettingsUpdates(), func() {
		settingRow(c, pal, locales.UpdatesAuto(), locales.UpdatesAutoDetail(), func() {
			ui.Switch(c.Key("autoUpdate"), &a.set.AutoUpdate).Label(locales.UpdatesAuto())
		})
		settingRow(c, pal, locales.UpdatesInstall(), locales.UpdatesInstallDetail(), func() {
			ui.Switch(c.Key("autoInstall"), &a.set.AutoInstall).Label(locales.UpdatesInstall()).Disabled(!a.set.AutoUpdate)
		})
		last := locales.UpdatesNever()
		if !a.upd.LastCheck.IsZero() {
			last = locales.UpdatesLast(a.upd.LastCheck.Local().Format("2006-01-02 15:04"))
		}
		detail := ""
		if !a.upd.enabled {
			detail = locales.UpdatesDisabled()
		}
		settingRow(c, pal, last, detail, func() {
			ui.Button(c.Key("checkNow"), locales.UpdatesCheckNow()).OnClick(a.checkForUpdates)
		})
	})
}
