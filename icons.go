package main

import "github.com/egoist/mygo/ui"

// icon draws an icon of Windeck's own on a 16 grid, in the text color, as
// SF Symbols draw.
func icon(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">` + shapes + `</svg>`))
}

var (
	searchIcon     = icon(`<circle cx="7" cy="7" r="4.5"/><path d="M10.4 10.4 14 14"/>`)
	closeIcon      = icon(`<path d="M4 4l8 8M12 4l-8 8"/>`)
	minimizeIcon   = icon(`<path d="M3.5 8h9"/>`)
	chevronRight   = icon(`<path d="M6 3.5 10.5 8 6 12.5"/>`)
	chevronDown    = icon(`<path d="M3.5 6 8 10.5 12.5 6"/>`)
	starIcon       = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor"><path d="M8 1.2l1.5 4.6 4.8.1-3.9 2.9 1.4 4.6L8 10.6l-3.8 2.8 1.4-4.6-3.9-2.9 4.8-.1z"/></svg>`))
	displayIcon    = icon(`<rect x="2" y="2.8" width="12" height="8.4" rx="1.4"/><path d="M6 13.6h4M8 11.2v2.4"/>`)
	desktopIcon    = icon(`<rect x="2" y="3" width="12" height="10" rx="1.6"/><path d="M2 6h12"/>`)
	minimizedIcon  = icon(`<path d="M8 2.5v7M5 6.8 8 9.8l3-3"/><path d="M3 13h10"/>`)
	hiddenIcon     = icon(`<path d="M2 8s2.2-4 6-4 6 4 6 4-2.2 4-6 4-6-4-6-4z"/><path d="M3 13 13 3"/>`)
	fullscreenIcon = icon(`<path d="M2.5 6V2.5H6M10 2.5h3.5V6M13.5 10v3.5H10M6 13.5H2.5V10"/>`)
	gearIcon       = icon(`<circle cx="8" cy="8" r="2.2"/><path d="M8 1.8v1.6M8 12.6v1.6M3.6 3.6l1.1 1.1M11.3 11.3l1.1 1.1M1.8 8h1.6M12.6 8h1.6M3.6 12.4l1.1-1.1M11.3 4.7l1.1-1.1"/>`)
	lockIcon       = icon(`<rect x="3" y="7" width="10" height="7" rx="1.6"/><path d="M5.2 7V5a2.8 2.8 0 0 1 5.6 0v2"/>`)
	sparkleIcon    = icon(`<path d="M8 2v3M8 11v3M2 8h3M11 8h3M4 4l1.8 1.8M10.2 10.2 12 12M4 12l1.8-1.8M10.2 5.8 12 4"/>`)
	keyIcon        = icon(`<circle cx="5.5" cy="10.5" r="3"/><path d="M7.7 8.3 13.5 2.5M11.5 4.5l1.5 1.5M10 6l1.2 1.2"/>`)
	infoIcon       = icon(`<circle cx="8" cy="8" r="6"/><path d="M8 7.2v4M8 4.8v.1"/>`)
	checkIcon      = icon(`<path d="M3.5 8.4 6.5 11.3 12.5 4.8"/>`)
	plusIcon       = icon(`<path d="M8 3v10M3 8h10"/>`)
	appIcon        = icon(`<rect x="2.5" y="2.5" width="11" height="11" rx="3"/>`)
	keyboardIcon   = icon(`<rect x="1.5" y="4" width="13" height="8" rx="1.6"/><path d="M4 6.6h.1M6.6 6.6h.1M9.3 6.6h.1M12 6.6h.1M5 9.4h6"/>`)
	shieldIcon     = icon(`<path d="M8 1.8 13 3.6v4c0 3.2-2.2 5.4-5 6.6-2.8-1.2-5-3.4-5-6.6v-4z"/>`)
	screenIcon     = icon(`<rect x="2" y="3" width="12" height="8.5" rx="1.4"/><circle cx="8" cy="7.2" r="1.6"/><path d="M5.5 13.5h5"/>`)
)
