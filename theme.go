package main

import (
	_ "embed"

	"github.com/egoist/mygo/ui"
)

// Windeck's look is the hz design system's: white, near-black text and
// one violet, used sparingly; grays with a touch of violet; flat colors,
// layers told apart by background and a 1-point border, shadows on what
// floats alone. Its sizes come from fixed steps:
//
//	corners  6 (tags), 10 (buttons, fields, rows), 16 (cards, the panel)
//	spacing  4, 8, 12, 16, 24, 32, 48
//	type     20/28 page titles, 16/24 sections, 14/22 body, 13/20 secondary, 12/16 small
//	weights  400, 500, 700
//
// Every color below is one of its tokens; the views take them from here,
// never as values of their own.

const (
	radiusTag  = 6
	radiusCtl  = 10
	radiusCard = 16

	fontUI      = "Inter"  // with the system's Chinese font after it
	fontDisplay = "Nunito" // the welcome's title and empty states alone
)

type palette struct {
	dark bool

	bgPage     ui.Color // --bg-page
	bgSubtle   ui.Color // --bg-subtle: sidebars, the window behind cards
	bgMuted    ui.Color // --bg-muted: hover, fields, gray tags
	bgElevated ui.Color // --bg-elevated: the panel
	border     ui.Color // --border
	borderBold ui.Color // --border-strong

	text      ui.Color // --text-primary
	textMuted ui.Color // --text-secondary
	textFaint ui.Color // --text-tertiary: placeholders and notes alone

	brand       ui.Color // --brand
	brandHover  ui.Color
	brandText   ui.Color // --brand-text
	brandSubtle ui.Color // --brand-subtle: the chosen row, tags, the big bubble
	brandOnSub  ui.Color // --brand-subtle-text
	onBrand     ui.Color

	success ui.Color
	warning ui.Color
	danger  ui.Color
	focus   ui.Color // --focus-ring

	shadow ui.Color // the panel's, lg
	scrim  ui.Color // what dims the panel under a guide
}

var lightPalette = palette{
	bgPage:     ui.Hex("#FFFFFF"),
	bgSubtle:   ui.Hex("#F7F7FA"),
	bgMuted:    ui.Hex("#EFEFF4"),
	bgElevated: ui.Hex("#FFFFFF"),
	border:     ui.Hex("#E2E2EA"),
	borderBold: ui.Hex("#C9C9D4"),

	text:      ui.Hex("#1C1C23"),
	textMuted: ui.Hex("#5E5E6E"),
	textFaint: ui.Hex("#7C7C8C"),

	brand:       ui.Hex("#5A4BC6"),
	brandHover:  ui.Hex("#4A3CA8"),
	brandText:   ui.Hex("#5A4BC6"),
	brandSubtle: ui.Hex("#F4F3FD"),
	brandOnSub:  ui.Hex("#4A3CA8"),
	onBrand:     ui.Hex("#FFFFFF"),

	success: ui.Hex("#2E9E6A"),
	warning: ui.Hex("#E07A2F"),
	danger:  ui.Hex("#D9433F"),
	focus:   ui.RGBA(90, 75, 198, 0.35),

	shadow: ui.RGBA(28, 28, 35, 0.12),
	scrim:  ui.RGBA(28, 28, 35, 0.40),
}

var darkPalette = palette{
	dark:       true,
	bgPage:     ui.Hex("#121216"),
	bgSubtle:   ui.Hex("#1C1C23"),
	bgMuted:    ui.Hex("#2E2E38"),
	bgElevated: ui.Hex("#24242C"),
	border:     ui.Hex("#2E2E38"),
	borderBold: ui.Hex("#464654"),

	text:      ui.Hex("#F7F7FA"),
	textMuted: ui.Hex("#A2A2B0"),
	textFaint: ui.Hex("#7C7C8C"),

	brand:       ui.Hex("#7366D8"),
	brandHover:  ui.Hex("#6456D0"),
	brandText:   ui.Hex("#B5ACF1"),
	brandSubtle: ui.Hex("#282160"),
	brandOnSub:  ui.Hex("#D3CDF7"),
	onBrand:     ui.Hex("#FFFFFF"),

	success: ui.Hex("#5FC795"),
	warning: ui.Hex("#F0A46A"),
	danger:  ui.Hex("#F07C78"),
	focus:   ui.RGBA(181, 172, 241, 0.45),

	shadow: ui.RGBA(0, 0, 0, 0.45),
	scrim:  ui.RGBA(0, 0, 0, 0.55),
}

// brandTheme is MyGo's widgets in the same colors: buttons, switches,
// segmented controls and focus rings in the violet, Inter for text.
func brandTheme(dark bool) *ui.Theme {
	p, t := &lightPalette, ui.LightTheme()
	if dark {
		p, t = &darkPalette, ui.DarkTheme()
	}
	t.Background = p.bgPage
	t.Surface, t.SurfaceHover, t.SurfacePressed = p.bgPage, p.bgSubtle, p.bgMuted
	t.Border = p.borderBold
	t.Text, t.TextMuted = p.text, p.textMuted
	t.Accent, t.AccentHover, t.AccentPressed, t.AccentText = p.brand, p.brandHover, p.brandHover, p.onBrand
	t.Danger, t.Warning, t.Success = p.danger, p.warning, p.success
	t.Selection = p.brand.Alpha(0.22)
	t.Focus = p.focus
	t.Radius = radiusCtl
	t.FontSize = 14
	t.Font = fontUI
	return t
}

// themed sets a window's theme for the frame and returns its palette:
// every view starts with it.
func themed(c *ui.Context) *palette {
	dark := c.Theme().Dark
	c.SetTheme(brandTheme(dark))
	if dark {
		return &darkPalette
	}
	return &lightPalette
}

func paletteFor(c *ui.Context) *palette {
	if c.Theme().Dark {
		return &darkPalette
	}
	return &lightPalette
}

var (
	//go:embed resources/fonts/Inter-Regular.ttf
	interRegular []byte
	//go:embed resources/fonts/Inter-Medium.ttf
	interMedium []byte
	//go:embed resources/fonts/Inter-Bold.ttf
	interBold []byte
	//go:embed resources/fonts/Nunito.ttf
	nunito []byte
)

// registerFonts makes Inter and Nunito ready for text: they are in the
// app, not on every Mac. Chinese text falls back to the system's font,
// as the design's font stack does.
func registerFonts() {
	for _, f := range [][]byte{interRegular, interMedium, interBold} {
		ui.RegisterFont(f, fontUI)
	}
	ui.RegisterFont(nunito, fontDisplay)
}
