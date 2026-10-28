package main

import (
	"slices"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"windeck/internal/settings"
	"windeck/internal/smart"
	"windeck/internal/switcher"
	"windeck/internal/wm"
	"windeck/locales"
)

// The panel: the windows as tabs down the left, grouped by app, and the
// highlighted window large at the right. Tabs are for finding, the
// preview for making sure.
//
// It floats flat, as the design has what floats: one color, a 1-point
// border and the large shadow, cut to the continuous corners of macOS's
// panels. Every title in the list starts on one line, the icons on
// another, so the eye runs straight down them.

const (
	panelRadius  = radiusCard
	sidebarWidth = 352
	listInset    = 8  // the list's padding at the sides
	rowInset     = 8  // a row's, before its icon
	rowIcon      = 20 // the icons of the list
	rowGap       = 12 // between a row's icon and its title
	titleStart   = rowInset + rowIcon + rowGap

	// The preview's header lines up with the search field, 12 from the
	// top and as tall; the picture takes the rest, with a small margin.
	previewTop   = 12
	previewPad   = 16
	headerHeight = 36
	headerGap    = 12

	// The list layout's.
	listWidth     = 720
	listRowHeight = 48
	listPad       = 8
	searchBlock   = 12 + 36 + 8 // the search field with its margins
	guideHeight   = 560         // the list panel's, under the guide
)

// The design's motion: 120 ms for hover, easing out.
var (
	easeOut     = ui.CubicBezier(0.2, 0, 0, 1)
	hoverMotion = ui.ElementTransition{Colors: true, Duration: 120 * time.Millisecond, Ease: easeOut}
)

func newPanel(a *App) *mygo.Window {
	w := mygo.NewWindow(mygo.WindowOptions{
		Title:       "Windeck",
		Panel:       true,
		Hidden:      true,
		Frameless:   true,
		Transparent: true,
		Width:       1100,
		Height:      700,
		Content:     ui.View(a.panelView),
	})
	// Hidden rather than closed, to show at once the next time: on
	// Escape the view does not take, and as the user goes elsewhere. As
	// Windeck quits, it closes like any window, or quitting would stop.
	w.OnClose(func(e *mygo.CloseEvent) {
		if a.quitting {
			return
		}
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
	pal := themed(c)
	c.Root().Background(ui.Transparent)
	if a.sw.Phase() == switcher.Idle {
		return
	}
	if a.ses.guide {
		defer a.guideOverlay(c, pal)
	}
	if a.set.Layout == settings.LayoutList {
		a.listPanelView(c, pal)
		a.panelKeys(c)
		return
	}
	card := ui.Row(c.Key("card")).Fill().Radius(panelRadius).Clip().AlignItems(ui.Stretch).
		Background(pal.bgElevated).Border(1, pal.border)
	card.Children(func() {
		a.sidebarView(c, pal)
		a.previewView(c, pal)
	})
	a.panelKeys(c)
}

// --- the list layout ---

// listPanelView is the list layout: one column of the windows, the most
// recent first, each its app's icon and its title, nothing else, as the
// app switchers of old showed them. Search adds its field at the top.
func (a *App) listPanelView(c *ui.Context, pal *palette) {
	card := ui.Column(c.Key("card")).Fill().Radius(panelRadius).Clip().AlignItems(ui.Stretch).
		Background(pal.bgElevated).Border(1, pal.border)
	card.Children(func() {
		if a.sw.Phase() == switcher.Searching {
			a.searchView(c, pal)
		}
		list := ui.Scroll(c.Key("list")).Grow(1).MinHeight(0).TrackScroll(&a.ses.scroll).Padding(listPad)
		list.Children(func() {
			if len(a.ses.rows) == 0 {
				a.emptyList(c, pal)
			}
			for _, r := range a.ses.rows {
				a.listRow(c, pal, r)
			}
		})
		a.scrolled()
	})
}

// listRow is a window in the list layout: the app's icon and the title,
// with the window's states and its place in the cycle at the right.
func (a *App) listRow(c *ui.Context, pal *palette, r row) {
	id := r.id
	sel := id == a.sw.Selected()
	title := rowTitle(r)
	el := ui.Row(c.Key(id)).Height(listRowHeight).Padding(0, 12).Gap(12).AlignItems(ui.Center).Radius(radiusCtl).
		Cursor(ui.CursorPointer).Label(title + ", " + r.name)
	hovered := el.Hovered()
	switch {
	case sel:
		el.Background(pal.brandSubtle)
	case hovered:
		el.Background(pal.bgMuted)
	}
	if hovered && !sel {
		// Only the pointer's hover eases in: the highlight moves with
		// the keys at once, leaving no trail as Tab is held.
		el.Transition(hoverMotion)
	}
	if sel && a.ses.scrollTo {
		el.ScrollIntoView()
	}
	a.hovering(id, hovered)
	el.Children(func() {
		a.appIconView(c, pal, r.path, 32)
		ui.Box(c).Grow(1).MinWidth(0).Children(func() {
			a.titleText(c, pal, title, 16, 400, sel)
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

// panelKeys are the keys of state B, which the panel's search field has:
// in state A the event tap takes them before any app does.
func (a *App) panelKeys(c *ui.Context) {
	if a.sw.Phase() != switcher.Searching {
		return
	}
	if a.ses.guide {
		// The guide's button is the one Return presses.
		if c.Shortcut(0, ui.KeyEnter) {
			a.hideGuide()
		}
		if c.Shortcut(0, ui.KeyEscape) {
			a.do(a.sw.Dismiss())
		}
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
		Background(pal.bgSubtle).BorderWidth(0, 1, 0, 0).BorderColor(pal.border)
	col.Children(func() {
		a.searchView(c, pal)
		list := ui.Scroll(c.Key("list")).Grow(1).MinHeight(0).TrackScroll(&a.ses.scroll).Padding(0, listInset, 8, listInset).Gap(0)
		list.Children(func() {
			if len(a.ses.rows) == 0 {
				a.emptyList(c, pal)
			}
			for i, r := range a.ses.rows {
				switch r.kind {
				case rowGroup:
					a.groupRow(c, pal, r, i == 0)
				default:
					a.windowRow(c, pal, r)
				}
			}
		})
		a.scrolled()
		a.hintsView(c, pal)
	})
}

// emptyList says why the list is empty: nothing matches, or there is
// nothing to switch to. An empty state, as the design draws them: a line
// drawing, a rounded title and one sentence.
func (a *App) emptyList(c *ui.Context, pal *palette) {
	ui.Column(c).AlignItems(ui.Center).Gap(8).Padding(48, 24).Children(func() {
		ui.Icon(c, searchIcon).Size(48, 48).TextColor(pal.borderBold).Margin(0, 0, 8, 0)
		title, detail := locales.ListEmpty(), ""
		if a.ses.query != "" {
			title, detail = locales.SearchNoResults(a.ses.query), locales.SearchNoResultsHint()
		}
		ui.Text(c, title).Font(fontDisplay).FontSize(16).FontWeight(700).TextColor(pal.text).TextAlign(ui.Center).MaxLines(2).Ellipsis("…")
		if detail != "" {
			ui.Text(c, detail).FontSize(13).TextColor(pal.textMuted).TextAlign(ui.Center)
		}
	})
}

// scrolled ends the frame's scrolling into view, which the highlight
// asks for once.
func (a *App) scrolled() { a.ses.scrollTo = false }

func (a *App) searchView(c *ui.Context, pal *palette) {
	box := ui.Row(c.Key("search")).Height(36).Shrink(0).Margin(12, 12, 8, 12).Padding(0, 12).Gap(8).Radius(radiusCtl).AlignItems(ui.Center)
	if a.sw.Phase() == switcher.Searching {
		// A focused field: the violet border and the 3-point focus ring.
		box.Background(pal.bgPage).Border(1, pal.brand).Shadow(0, 0, 0, 3, pal.focus)
		box.Children(func() {
			ui.Icon(c, searchIcon).Size(16, 16).TextColor(pal.textMuted)
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
	box.Background(pal.bgPage).Border(1, pal.border).Cursor(ui.CursorText).Label(locales.SearchPlaceholder())
	if box.Hovered() {
		box.Border(1, pal.borderBold)
	}
	box.Transition(hoverMotion)
	box.Children(func() {
		ui.Icon(c, searchIcon).Size(16, 16).TextColor(pal.textFaint)
		ui.Text(c, locales.SearchPlaceholder()).FontSize(14).TextColor(pal.textFaint).Grow(1).SingleLine()
		kbd(c, pal, locales.KeySpace())
	})
	if box.Clicked() {
		a.do(a.sw.StartSearch())
	}
}

// kbd draws a key, as the hints name them.
func kbd(c *ui.Context, pal *palette, label string) {
	ui.Box(c).MinWidth(20).Height(20).Padding(0, 4).Radius(radiusTag).Center().Background(pal.bgPage).Border(1, pal.border).Children(func() {
		ui.Text(c, label).FontSize(12).FontWeight(500).TextColor(pal.textMuted).SingleLine()
	})
}

func (a *App) hintsView(c *ui.Context, pal *palette) {
	ui.Row(c.Key("hints")).Shrink(0).Height(40).Gap(12).Padding(0, 16).BorderWidth(1, 0, 0, 0).BorderColor(pal.border).
		AlignItems(ui.Center).Clip().Children(func() {
		hint := func(k, label string) {
			ui.Row(c).Gap(5).Shrink(0).AlignItems(ui.Center).Children(func() {
				kbd(c, pal, k)
				ui.Text(c, label).FontSize(12).TextColor(pal.textMuted).SingleLine()
			})
		}
		if a.sw.Phase() == switcher.Searching {
			hint("↑↓", locales.HintMove())
			hint("↩", locales.HintSwitch())
			hint("esc", locales.HintClear())
			return
		}
		if !a.set.ActionKeys {
			hint("⇥", locales.HintNext())
			hint("/", locales.ShortcutsSearch())
			return
		}
		// The keys that act, named for what they act on: W a window, Q
		// its whole app. Tab, which everyone holding it knows, gives
		// them its room.
		hint("W", locales.HintClose())
		hint("M", locales.HintMinimize())
		hint("Q", locales.HintQuit())
	})
}

// groupRow is the header of an app's windows: its icon and name where
// the rows have theirs, the count and the arrow at the right, as a
// sidebar section of Finder's.
func (a *App) groupRow(c *ui.Context, pal *palette, r row, first bool) {
	h := ui.Row(c.Key("g:"+r.key)).Height(32).Padding(0, 8, 0, rowInset).Gap(rowGap).AlignItems(ui.Center).Radius(radiusCtl).
		Cursor(ui.CursorPointer).Label(r.name)
	if !first {
		h.Margin(8, 0, 0, 0)
	}
	hovered := h.Hovered()
	if hovered {
		h.Background(pal.bgMuted)
	}
	h.Transition(hoverMotion)
	h.Children(func() {
		a.appIconView(c, pal, r.path, rowIcon)
		ui.Text(c, r.name).FontSize(13).FontWeight(500).TextColor(pal.textMuted).SingleLine().Ellipsis("…").Grow(1).MinWidth(0)
		ui.Text(c, itoa(r.count)).FontSize(12).TextColor(pal.textFaint).Shrink(0)
		chev := chevronDown
		if !r.open {
			chev = chevronRight
		}
		col := pal.textFaint
		if hovered {
			col = pal.textMuted
		}
		ui.Icon(c, chev).Size(16, 16).TextColor(col).Shrink(0)
	})
	key, open := r.key, r.open
	h.OnClick(func() {
		a.ses.collapsed[key] = open
		a.rebuildRows()
	})
}

// rowTitle is what a row says: the window's title without its app's
// name, which the row shows anyway, or the app's name.
func rowTitle(r row) string {
	if r.kind == rowApp || r.win.Title == "" {
		return r.name
	}
	return shortTitle(r.win.Title, r.name)
}

func (a *App) windowRow(c *ui.Context, pal *palette, r row) {
	id := r.id
	sel := id == a.sw.Selected()
	title := rowTitle(r)
	sub := !r.child && r.kind == rowWindow && title != r.name
	el := ui.Row(c.Key(id)).Gap(rowGap).AlignItems(ui.Center).Radius(radiusCtl).Cursor(ui.CursorPointer).
		Label(title + ", " + r.name)
	if sub {
		el.MinHeight(48).Padding(4, 8, 4, rowInset)
	} else {
		el.MinHeight(36).Padding(4, 8, 4, rowInset)
	}
	if r.child {
		el.PaddingStart(titleStart)
	}
	hovered := el.Hovered()
	switch {
	case sel:
		el.Background(pal.brandSubtle)
	case hovered:
		el.Background(pal.bgMuted)
	}
	if hovered && !sel {
		// Only the pointer's hover eases in: the highlight moves with
		// the keys at once, leaving no trail as Tab is held.
		el.Transition(hoverMotion)
	}
	if sel && a.ses.scrollTo {
		el.ScrollIntoView()
	}
	a.hovering(id, hovered)
	el.Children(func() {
		if !r.child {
			a.appIconView(c, pal, r.path, rowIcon)
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(1).Children(func() {
			a.titleText(c, pal, title, 14, 400, sel)
			if sub {
				ui.Text(c, r.name).FontSize(12).TextColor(pal.textMuted).SingleLine().Ellipsis("…")
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

// titleText is a title on one line, its letters the search matched in
// bold, as Spotlight shows them. The chosen row's title is the
// subtle-violet's text color.
func (a *App) titleText(c *ui.Context, pal *palette, title string, size float32, weight int, sel bool) {
	color := pal.text
	if sel {
		color = pal.brandOnSub
	}
	at := smart.Positions(a.ses.query, title)
	// Letters strewn about read as noise: only a run of them is bold.
	if len(at) > 0 && at[len(at)-1]-at[0] != len(at)-1 {
		at = nil
	}
	if len(at) == 0 {
		ui.Text(c, title).FontSize(size).FontWeight(weight).TextColor(color).SingleLine().Ellipsis("…")
		return
	}
	var spans []ui.Span
	runes := []rune(title)
	start, bold := 0, slices.Contains(at, 0)
	for i := 1; i <= len(runes); i++ {
		b := i < len(runes) && slices.Contains(at, i)
		if i == len(runes) || b != bold {
			s := ui.Span{Text: string(runes[start:i]), Size: size, Weight: weight, Color: color}
			if bold {
				s.Weight = 700
			}
			spans = append(spans, s)
			start, bold = i, b
		}
	}
	ui.RichText(c, spans...).SingleLine().Ellipsis("…")
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
		if r.win.Minimized {
			// In the Dock: back, as its Dock picture does.
			smallButton(c, pal, "restore", restoreIcon, locales.RowRestore(), false).OnClick(func() { a.restoreWindow(id) })
		} else {
			smallButton(c, pal, "min", minimizeIcon, locales.RowMinimize(), false).OnClick(func() { a.minimizeWindow(id) })
		}
		smallButton(c, pal, "close", closeIcon, locales.RowClose(), true).OnClick(func() { a.closeWindow(id) })
	})
}

func smallButton(c *ui.Context, pal *palette, key string, svg *ui.SVG, label string, danger bool) ui.Element {
	b := ui.ButtonBase(c.Key(key)).Size(24, 24).Radius(radiusTag).Center().TextColor(pal.textMuted).Label(label).Tooltip(label)
	switch {
	case b.Pressed() && danger:
		b.Background(pal.danger.Alpha(0.85)).TextColor(pal.onBrand)
	case b.Hovered() && danger:
		b.Background(pal.danger).TextColor(pal.onBrand)
	case b.Pressed():
		b.Background(pal.borderBold).TextColor(pal.text)
	case b.Hovered():
		b.Background(pal.bgMuted).TextColor(pal.text)
	}
	b.Transition(hoverMotion)
	b.Children(func() { ui.Icon(c, svg).Size(16, 16) })
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
		ui.Icon(c, starIcon).Size(12, 12).TextColor(pal.brandText).Tooltip(locales.PredictedTip()).Label(locales.PredictedTip())
	}
	// The presses of Tab a window is away: the list layout's rows are in
	// the cycle's order, so its numbers count down the rows. The tabs'
	// are by app, where Tab goes down the rows too and numbers of the
	// cycle would point elsewhere.
	if a.set.Layout == settings.LayoutList && i >= 0 && i < 3 {
		ui.Text(c, itoa(i+1)).FontSize(12).FontWeight(500).TextColor(pal.textFaint).Width(12).TextAlign(ui.End)
	}
}

// badges are a window's states: where it is and how it shows. The list
// shows them as icons, with the words as a tooltip; the preview's header
// spells them out.
func (a *App) badges(c *ui.Context, pal *palette, r row, compact bool) {
	chip := func(svg *ui.SVG, short, label string) {
		// A tag, as the design's: 22 high, 12-point text, gray.
		b := ui.Row(c).Gap(4).Height(22).Padding(0, 4).Radius(radiusTag).Background(pal.bgMuted).AlignItems(ui.Center).Shrink(0).Label(label)
		if compact {
			b.Tooltip(label)
		} else {
			b.Padding(0, 8)
		}
		b.Children(func() {
			ui.Icon(c, svg).Size(14, 14).TextColor(pal.textMuted)
			text := label
			if compact {
				text = short
			}
			if text != "" {
				ui.Text(c, text).FontSize(12).TextColor(pal.textMuted).SingleLine()
			}
		})
	}
	if r.kind == rowApp {
		ui.Text(c, locales.BadgeNoWindows()).FontSize(12).TextColor(pal.textFaint).SingleLine()
		return
	}
	w := r.win
	if w.Space > 0 && !w.OnCurrentSpace && !w.Minimized {
		chip(desktopIcon, itoa(w.Space), locales.BadgeDesktop(w.Space))
	}
	if w.External {
		chip(displayIcon, "", locales.BadgeDisplay())
	}
	if w.FullScreen {
		chip(fullscreenIcon, "", locales.BadgeFullScreen())
	}
	if w.Minimized {
		chip(minimizedIcon, "", locales.BadgeMinimized())
	}
	if w.Hidden {
		chip(hiddenIcon, "", locales.BadgeHidden())
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
	// Until the icon loads, the shape of one.
	ui.Box(c).Size(size, size).Shrink(0).Radius(size / 4.5).Background(pal.bgMuted)
}

// --- the preview ---

func (a *App) previewView(c *ui.Context, pal *palette) {
	id := a.sw.Selected()
	r, ok := a.rowOf(id)
	col := ui.Column(c.Key("preview")).Grow(1).MinWidth(0).Padding(previewTop, previewPad, previewPad, previewPad).Gap(headerGap).AlignItems(ui.Stretch)
	col.Children(func() {
		if !ok {
			return
		}
		a.previewHeader(c, pal, r)
		stage := ui.Box(c.Key("stage")).Grow(1).MinHeight(0).Center()
		stage.Children(func() {
			var bmp *ui.Bitmap
			if r.kind == rowWindow && a.showsPicture(r.win) {
				bmp = a.previewOf(id)
			}
			// The picture as large as the stage takes it, never larger
			// than the window is: a small window stays small.
			W, H := c.Size()
			sw := W - sidebarWidth - 2*previewPad
			sh := H - previewTop - headerHeight - headerGap - previewPad
			if bmp == nil {
				if a.picturePending(r) && sw > 0 && sh > 0 {
					// Its picture is on the way: the window's shape, in
					// the panel's own color, where the picture comes, so
					// that it fills in rather than replacing a card. Tab
					// held through windows shows no flashing.
					bw, bh := float32(r.win.Bounds.W), float32(r.win.Bounds.H)
					s := min(sw/bw, sh/bh, 1)
					ui.Box(c).Size(bw*s, bh*s).Radius(radiusCard).Background(pal.bgSubtle).Border(1, pal.border)
					return
				}
				a.fallbackCard(c, pal, r)
				return
			}
			bw, bh := bmp.Size()
			if bw <= 0 || bh <= 0 || sw <= 0 || sh <= 0 {
				return
			}
			s := min(sw/float32(bw), sh/float32(bh))
			if r.win.Bounds.W > 0 {
				s = min(s, float32(r.win.Bounds.W)/float32(bw))
			}
			// A picture: the design's 16 corners and a border, no shadow.
			ui.Image(c, bmp).Size(float32(bw)*s, float32(bh)*s).Radius(radiusCard).Border(1, pal.border).
				Label(locales.PreviewOf(r.win.Title))
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
		a.appIconView(c, pal, r.path, 28)
		ui.Column(c).Grow(1).MinWidth(0).Gap(0).Children(func() {
			title := r.win.Title
			if title == "" || r.kind == rowApp {
				title = r.name
			}
			ui.Text(c, title).FontSize(14).FontWeight(500).TextColor(pal.text).SingleLine().Ellipsis("…")
			if title != r.name {
				ui.Text(c, r.name).FontSize(12).TextColor(pal.textMuted).SingleLine()
			}
		})
		ui.Row(c).Gap(8).Shrink(0).Children(func() { a.badges(c, pal, r, false) })
	})
}

// fallbackCard stands in for a picture there is none of: the app's large
// icon and why, never a blank.
func (a *App) fallbackCard(c *ui.Context, pal *palette, r row) {
	reason := a.noPreviewReason(r)
	ui.Column(c).AlignItems(ui.Center).Gap(8).Padding(24).MaxWidth(400).Children(func() {
		a.appIconView(c, pal, r.path, 96)
		ui.Text(c, rowTitle(r)).FontSize(16).FontWeight(500).TextColor(pal.text).TextAlign(ui.Center).MaxLines(2).Ellipsis("…").Margin(16, 0, 0, 0)
		if reason != "" {
			ui.Text(c, reason).FontSize(14).TextColor(pal.textMuted).TextAlign(ui.Center)
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
