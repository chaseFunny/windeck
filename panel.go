package main

import (
	"slices"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"windeck/internal/switcher"
	"windeck/internal/wm"
	"windeck/locales"
)

// The panel: the windows as tabs down the left, grouped by app, and the
// highlighted window large in the middle. Tabs are for finding, the
// preview for making sure.

const (
	panelInset   = 16  // room around the card for its shadow
	sidebarWidth = 360 // the list's
	previewPad   = 26
	headerHeight = 56
)

var hoverMotion = ui.ElementTransition{Colors: true, Duration: 110 * time.Millisecond}

func newPanel(a *App) *mygo.Window {
	w := mygo.NewWindow(mygo.WindowOptions{
		Title:         "Windeck",
		Panel:         true,
		Hidden:        true,
		Frameless:     true,
		Transparent:   true,
		DisableShadow: true,
		Width:         1100,
		Height:        700,
		Content:       ui.View(a.panelView),
	})
	// Hidden rather than closed, to show at once the next time: on
	// Escape the view does not take, and as the user goes elsewhere.
	w.OnClose(func(e *mygo.CloseEvent) {
		e.PreventDefault()
		a.do(a.sw.Dismiss())
	})
	w.OnBlur(func() {
		if a.sw.Phase() == switcher.Searching {
			a.do(a.sw.Dismiss())
		}
	})
	return w
}

func (a *App) panelView(c *ui.Context) {
	pal := paletteFor(c)
	c.Root().Background(ui.Transparent)
	if a.sw.Phase() == switcher.Idle {
		return
	}
	ui.Box(c).Fill().Padding(panelInset).Children(func() {
		card := ui.Row(c.Key("card")).Fill().Radius(18).Clip().AlignItems(ui.Stretch).
			Background(pal.card).Border(1, pal.cardBorder).
			Shadow(0, 18, 48, 0, pal.shadow).Shadow(0, 2, 6, 0, pal.shadow2)
		card.Children(func() {
			a.sidebarView(c, pal)
			a.previewView(c, pal)
		})
	})
	a.panelKeys(c)
}

// panelKeys are the keys of state B, which the panel's search field has:
// in state A the event tap takes them before any app does.
func (a *App) panelKeys(c *ui.Context) {
	if a.sw.Phase() != switcher.Searching {
		return
	}
	if c.Shortcut(0, ui.KeyDown) || c.Shortcut(0, ui.KeyTab) || c.Shortcut(ui.Ctrl, ui.KeyN) {
		a.do(a.sw.Move(1))
	}
	if c.Shortcut(0, ui.KeyUp) || c.Shortcut(ui.Shift, ui.KeyTab) || c.Shortcut(ui.Ctrl, ui.KeyP) {
		a.do(a.sw.Move(-1))
	}
	if c.Shortcut(0, ui.KeyEnter) {
		a.do(a.sw.Enter())
	}
	if c.Shortcut(0, ui.KeyEscape) {
		empty := a.ses.searchBuf == ""
		a.ses.searchBuf = ""
		a.setQuery("")
		a.do(a.sw.Escape(empty))
	}
}

// --- the list ---

func (a *App) sidebarView(c *ui.Context, pal *palette) {
	col := ui.Column(c.Key("sidebar")).Width(sidebarWidth).Shrink(0).AlignItems(ui.Stretch).
		Background(pal.sidebar).BorderWidth(0, 1, 0, 0).BorderColor(pal.divider)
	col.Children(func() {
		a.searchView(c, pal)
		list := ui.Scroll(c.Key("list")).Grow(1).MinHeight(0).TrackScroll(&a.ses.scroll).Padding(2, 10, 10, 10).Gap(2)
		list.Children(func() {
			if len(a.ses.rows) == 0 {
				ui.Text(c, locales.ListEmpty()).FontSize(13).TextColor(pal.textMuted).Margin(24, 8)
			}
			for _, r := range a.ses.rows {
				switch r.kind {
				case rowGroup:
					a.groupRow(c, pal, r)
				default:
					a.windowRow(c, pal, r)
				}
			}
		})
		a.scrolled()
		a.hintsView(c, pal)
	})
}

// scrolled ends the frame's scrolling into view, which the highlight
// asks for once.
func (a *App) scrolled() { a.ses.scrollTo = false }

func (a *App) searchView(c *ui.Context, pal *palette) {
	box := ui.Row(c.Key("search")).Height(38).Shrink(0).Margin(12, 12, 8, 12).Padding(0, 10).Gap(8).Radius(10).AlignItems(ui.Center)
	if a.sw.Phase() == switcher.Searching {
		box.Background(pal.fieldFocus).Border(1, pal.accent.Alpha(0.55)).Shadow(0, 0, 0, 3, pal.accentSoft)
		box.Children(func() {
			ui.Icon(c, searchIcon).Size(15, 15).TextColor(pal.textMuted)
			in := ui.TextInputBase(c.Key("q"), &a.ses.searchBuf).Grow(1).FontSize(14).TextColor(pal.text).
				Placeholder(locales.SearchPlaceholder()).Bind(&a.ses.search).Label(locales.SearchPlaceholder())
			if a.ses.focusSearch {
				a.ses.search.Focus(c)
				a.ses.focusSearch = false
			}
			if in.Changed() {
				a.setQuery(a.ses.searchBuf)
			}
		})
		return
	}
	box.Background(pal.field).Cursor(ui.CursorText).Label(locales.SearchPlaceholder())
	box.Children(func() {
		ui.Icon(c, searchIcon).Size(15, 15).TextColor(pal.textFaint)
		ui.Text(c, locales.SearchPlaceholder()).FontSize(14).TextColor(pal.textFaint).Grow(1).SingleLine()
		kbd(c, pal, locales.KeySpace())
	})
	if box.Clicked() {
		a.do(a.sw.StartSearch())
	}
}

// kbd draws a key, as the hints name them.
func kbd(c *ui.Context, pal *palette, label string) {
	ui.Box(c).Padding(1, 6).Radius(5).Background(pal.kbd).Border(1, pal.kbdBorder).Children(func() {
		ui.Text(c, label).FontSize(11).FontWeight(500).TextColor(pal.textMuted).SingleLine()
	})
}

func (a *App) hintsView(c *ui.Context, pal *palette) {
	ui.Row(c.Key("hints")).Shrink(0).Wrap().GapX(12).GapY(6).Padding(10, 14, 12, 14).BorderWidth(1, 0, 0, 0).BorderColor(pal.divider).
		AlignItems(ui.Center).Children(func() {
		hint := func(k, label string) {
			ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
				kbd(c, pal, k)
				ui.Text(c, label).FontSize(11.5).TextColor(pal.textMuted).SingleLine()
			})
		}
		if a.sw.Phase() == switcher.Searching {
			hint("↑↓", locales.HintMove())
			hint("↵", locales.HintSwitch())
			hint("esc", locales.HintClear())
			return
		}
		hint("⇥", locales.HintNext())
		hint("↑↓", locales.HintMove())
		hint("W", locales.HintClose())
		hint("M", locales.HintMinimize())
		hint("Q", locales.HintQuit())
	})
}

func (a *App) groupRow(c *ui.Context, pal *palette, r row) {
	h := ui.Row(c.Key("g:"+r.key)).Padding(8, 8, 4, 6).Gap(7).AlignItems(ui.Center).Radius(8).Cursor(ui.CursorPointer).
		Label(r.name)
	if h.Hovered() {
		h.Background(pal.rowHover)
	}
	h.Transition(hoverMotion)
	h.Children(func() {
		chev := chevronDown
		if !r.open {
			chev = chevronRight
		}
		ui.Icon(c, chev).Size(11, 11).TextColor(pal.textFaint)
		a.appIconView(c, pal, r.path, 18)
		ui.Text(c, r.name).FontSize(12.5).FontWeight(600).TextColor(pal.textMuted).SingleLine().Ellipsis("…").Shrink(1).MinWidth(0)
		ui.Box(c).Padding(0, 6).Radius(8).Background(pal.chip).Children(func() {
			ui.Text(c, itoa(r.count)).FontSize(11).FontWeight(500).TextColor(pal.chipText)
		})
	})
	key, open := r.key, r.open
	h.OnClick(func() {
		a.ses.collapsed[key] = open
		a.rebuildRows()
	})
}

func (a *App) windowRow(c *ui.Context, pal *palette, r row) {
	id := r.id
	sel := id == a.sw.Selected()
	title := r.win.Title
	if r.kind == rowApp {
		title = r.name
	} else if title == "" {
		title = r.name
	}
	el := ui.Row(c.Key(id)).Gap(10).AlignItems(ui.Center).Radius(9).MinHeight(38).Cursor(ui.CursorPointer).
		Label(title + ", " + r.name)
	left := float32(10)
	if r.child {
		left = 21
	}
	el.Padding(7, 8, 7, left)
	hovered := el.Hovered()
	switch {
	case sel:
		el.Background(pal.rowSelected).Shadow(0, 1, 3, 0, pal.rowShadow).Border(1, pal.accent.Alpha(0.35))
	case hovered:
		el.Background(pal.rowHover)
	}
	el.Transition(hoverMotion)
	if sel && a.ses.scrollTo {
		el.ScrollIntoView()
	}
	a.hovering(id, hovered)
	el.Children(func() {
		if r.child {
			// The line down a group's windows, in place of their icon.
			ui.Box(c).Width(1.5).Height(22).Radius(1).Background(pal.guide).Margin(0, 6, 0, 2)
		} else {
			a.appIconView(c, pal, r.path, 22)
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(1).Children(func() {
			t := ui.Text(c, title).FontSize(13).TextColor(pal.text).MaxLines(2).Ellipsis("…")
			if sel {
				t.FontWeight(600)
			} else {
				t.FontWeight(450)
			}
			if !r.child && r.kind == rowWindow && r.win.Title != "" && r.win.Title != r.name {
				ui.Text(c, r.name).FontSize(11).TextColor(pal.textMuted).SingleLine().Ellipsis("…")
			}
		})
		if hovered && r.kind == rowWindow {
			a.rowButtons(c, pal, r)
			return
		}
		ui.Row(c).Gap(4).Shrink(0).AlignItems(ui.Center).Children(func() {
			a.badges(c, pal, r, true)
			a.ordinal(c, pal, id)
		})
	})
	el.OnClick(func() { a.do(a.sw.Click(id)) })
}

// hovering follows the pointer from row to row: the row under it is
// highlighted, as long as the pointer moved since the panel showed, so
// that a panel showing under a still pointer keeps the Tab's choice.
func (a *App) hovering(id wm.WindowID, hovered bool) {
	if !hovered || a.ses.hovered == id {
		if !hovered && a.ses.hovered == id {
			a.ses.hovered = 0
		}
		return
	}
	a.ses.hovered = id
	p := mygo.Screen.CursorScreenPoint()
	if p == a.ses.pointer {
		return
	}
	a.ses.pointer = p
	if a.sw.Selected() != id {
		a.sw.Hover(id)
		a.selectionMoved()
	}
}

// rowButtons are the minimize and close buttons a hovered row shows.
func (a *App) rowButtons(c *ui.Context, pal *palette, r row) {
	id := r.id
	ui.Row(c).Gap(2).Shrink(0).Children(func() {
		if !r.win.Minimized {
			smallButton(c, pal, "min", minimizeIcon, locales.RowMinimize(), false).OnClick(func() { a.minimizeWindow(id) })
		}
		smallButton(c, pal, "close", closeIcon, locales.RowClose(), true).OnClick(func() { a.closeWindow(id) })
	})
}

func smallButton(c *ui.Context, pal *palette, key string, svg *ui.SVG, label string, danger bool) ui.Element {
	b := ui.ButtonBase(c.Key(key)).Size(24, 24).Radius(6).Center().TextColor(pal.textMuted).Label(label).Tooltip(label)
	if b.Hovered() {
		if danger {
			b.Background(pal.danger).TextColor(ui.RGB(255, 255, 255))
		} else {
			b.Background(pal.chip).TextColor(pal.text)
		}
	}
	b.Children(func() { ui.Icon(c, svg).Size(13, 13) })
	return b
}

// ordinal shows a window's place in the cycle, 1 to 3, faint, and marks
// the predicted one.
func (a *App) ordinal(c *ui.Context, pal *palette, id wm.WindowID) {
	if a.ses.query != "" {
		return
	}
	i := slices.Index(a.sw.Cycle(), id)
	if id == a.ses.predicted && i == 1 {
		ui.Icon(c, starIcon).Size(11, 11).TextColor(pal.star).Tooltip(locales.PredictedTip()).Label(locales.PredictedTip())
	}
	if i >= 0 && i < 3 {
		ui.Text(c, itoa(i+1)).FontSize(11).FontWeight(600).TextColor(pal.textFaint).Width(10).TextAlign(ui.End)
	}
}

// badges are a window's states: where it is and how it shows.
func (a *App) badges(c *ui.Context, pal *palette, r row, compact bool) {
	chip := func(svg *ui.SVG, label string) {
		b := ui.Row(c).Gap(3).Padding(2, 5).Radius(6).Background(pal.chip).AlignItems(ui.Center).Shrink(0).Label(label)
		if compact {
			b.Tooltip(label)
		}
		b.Children(func() {
			ui.Icon(c, svg).Size(11, 11).TextColor(pal.chipText)
			if !compact {
				ui.Text(c, label).FontSize(11).TextColor(pal.chipText).SingleLine()
			}
		})
	}
	text := func(label string) {
		ui.Box(c).Padding(1, 5).Radius(6).Background(pal.chip).Shrink(0).Children(func() {
			ui.Text(c, label).FontSize(10.5).FontWeight(500).TextColor(pal.chipText).SingleLine()
		})
	}
	if r.kind == rowApp {
		text(locales.BadgeNoWindows())
		return
	}
	w := r.win
	if w.Space > 0 && !w.OnCurrentSpace && !w.Minimized {
		if compact {
			text(locales.BadgeDesktopShort(w.Space))
		} else {
			chip(desktopIcon, locales.BadgeDesktop(w.Space))
		}
	}
	if w.External {
		chip(displayIcon, locales.BadgeDisplay())
	}
	if w.FullScreen {
		chip(fullscreenIcon, locales.BadgeFullScreen())
	}
	if w.Minimized {
		chip(minimizedIcon, locales.BadgeMinimized())
	}
	if w.Hidden {
		chip(hiddenIcon, locales.BadgeHidden())
	}
}

func (a *App) appIconView(c *ui.Context, pal *palette, path string, size float32) {
	px := 64
	if size > 32 {
		px = 256
	}
	if b := a.icon(path, px); b != nil {
		ui.Image(c, b).Size(size, size).Shrink(0)
		return
	}
	ui.Box(c).Size(size, size).Shrink(0).Radius(size/4.5).Background(pal.chip)
}

// --- the preview ---

func (a *App) previewView(c *ui.Context, pal *palette) {
	id := a.sw.Selected()
	r, ok := a.rowOf(id)
	col := ui.Column(c.Key("preview")).Grow(1).MinWidth(0).Padding(previewPad-6, previewPad, previewPad, previewPad).Gap(14).AlignItems(ui.Stretch)
	col.Children(func() {
		if !ok {
			ui.Box(c).Grow(1).Center().Children(func() {
				ui.Text(c, locales.ListEmpty()).FontSize(14).TextColor(pal.textMuted)
			})
			return
		}
		a.previewHeader(c, pal, r)
		stage := ui.Box(c.Key("stage")).Grow(1).MinHeight(0).Center().Radius(14).Background(pal.stage)
		stage.Children(func() {
			var bmp *ui.Bitmap
			if r.kind == rowWindow {
				bmp = a.previewOf(id)
			}
			if bmp == nil {
				a.fallbackCard(c, pal, r)
				return
			}
			// Fit the picture to the stage, as large as it goes.
			W, H := c.Size()
			sw := W - 2*panelInset - sidebarWidth - 2*previewPad - 24
			sh := H - 2*panelInset - headerHeight - 14 - previewPad*2 - 24
			bw, bh := bmp.Size()
			if bw <= 0 || bh <= 0 || sw <= 0 || sh <= 0 {
				return
			}
			s := min(sw/float32(bw), sh/float32(bh))
			ui.Image(c, bmp).Size(float32(bw)*s, float32(bh)*s).Radius(10).Border(1, pal.frame).
				Shadow(0, 10, 30, 0, pal.shadow2).Label(locales.PreviewOf(r.win.Title))
		})
	})
}

func (a *App) rowOf(id wm.WindowID) (row, bool) {
	for _, r := range a.ses.rows {
		if r.id == id && r.kind != rowGroup {
			return r, true
		}
	}
	return row{}, false
}

func (a *App) previewHeader(c *ui.Context, pal *palette, r row) {
	ui.Row(c.Key("head")).Height(headerHeight).Shrink(0).Gap(12).AlignItems(ui.Center).Children(func() {
		a.appIconView(c, pal, r.path, 34)
		ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
			title := r.win.Title
			if title == "" || r.kind == rowApp {
				title = r.name
			}
			ui.Text(c, title).FontSize(17).FontWeight(650).TextColor(pal.text).MaxLines(2).Ellipsis("…")
			ui.Text(c, r.name).FontSize(12.5).TextColor(pal.textMuted).SingleLine()
		})
		ui.Row(c).Gap(5).Shrink(0).Children(func() { a.badges(c, pal, r, false) })
	})
}

// fallbackCard stands in for a picture there is none of: the app's large
// icon and why, never a blank.
func (a *App) fallbackCard(c *ui.Context, pal *palette, r row) {
	reason := a.noPreviewReason(r)
	ui.Column(c).Center().Gap(14).Padding(32).MaxWidth(440).Children(func() {
		a.appIconView(c, pal, r.path, 112)
		title := r.win.Title
		if title == "" || r.kind == rowApp {
			title = r.name
		}
		ui.Text(c, title).FontSize(16).FontWeight(600).TextColor(pal.text).TextAlign(ui.Center).MaxLines(3).Ellipsis("…")
		if reason != "" {
			ui.Text(c, reason).FontSize(13).TextColor(pal.textMuted).TextAlign(ui.Center)
		}
	})
}

func (a *App) noPreviewReason(r row) string {
	w := r.win
	switch {
	case r.kind == rowApp:
		return locales.PreviewApp()
	case a.set.IsExcluded(w.BundleID):
		return locales.PreviewExcluded()
	case !a.set.Previews:
		return locales.PreviewOff()
	case !a.canCapture:
		return locales.PreviewNoPermission()
	case w.Minimized:
		return locales.PreviewMinimized()
	case w.Hidden:
		return locales.PreviewHidden()
	}
	return ""
}
