package main

import "github.com/egoist/mygo/ui"

// palette is Windeck's colors, for the light appearance, its default, and
// the dark one.
type palette struct {
	dark bool

	card       ui.Color // the panel
	cardBorder ui.Color
	shadow     ui.Color
	shadow2    ui.Color
	sidebar    ui.Color // the list's side of the panel
	divider    ui.Color

	text      ui.Color
	textMuted ui.Color // app names, hints
	textFaint ui.Color // the cycle's numbers, counts

	rowHover    ui.Color
	rowSelected ui.Color
	rowShadow   ui.Color
	guide       ui.Color // the line down a group's windows

	field      ui.Color // the search field
	fieldFocus ui.Color
	chip       ui.Color // status badges
	chipText   ui.Color
	kbd        ui.Color
	kbdBorder  ui.Color

	accent     ui.Color
	accentSoft ui.Color
	danger     ui.Color
	star       ui.Color // the predicted window's mark

	stage      ui.Color // behind the preview
	frame      ui.Color // around the preview
	windowBack ui.Color // the settings and onboarding windows
	surface    ui.Color // their cards
}

var lightPalette = palette{
	card:       ui.RGBA(250, 250, 252, 0.97),
	cardBorder: ui.RGBA(0, 0, 0, 0.08),
	shadow:     ui.RGBA(18, 22, 40, 0.20),
	shadow2:    ui.RGBA(18, 22, 40, 0.08),
	sidebar:    ui.RGBA(240, 242, 246, 0.9),
	divider:    ui.RGBA(0, 0, 0, 0.07),

	text:      ui.Hex("#1d1d1f"),
	textMuted: ui.Hex("#6b6b76"),
	textFaint: ui.Hex("#a1a1ab"),

	rowHover:    ui.RGBA(0, 0, 0, 0.045),
	rowSelected: ui.Hex("#ffffff"),
	rowShadow:   ui.RGBA(20, 24, 50, 0.10),
	guide:       ui.RGBA(0, 0, 0, 0.10),

	field:      ui.RGBA(0, 0, 0, 0.045),
	fieldFocus: ui.Hex("#ffffff"),
	chip:       ui.RGBA(0, 0, 0, 0.055),
	chipText:   ui.Hex("#5d5d68"),
	kbd:        ui.Hex("#ffffff"),
	kbdBorder:  ui.RGBA(0, 0, 0, 0.12),

	accent:     ui.Hex("#2f6bff"),
	accentSoft: ui.RGBA(47, 107, 255, 0.12),
	danger:     ui.Hex("#e5484d"),
	star:       ui.Hex("#f5a524"),

	stage:      ui.RGBA(0, 0, 0, 0.025),
	frame:      ui.RGBA(0, 0, 0, 0.10),
	windowBack: ui.Hex("#f5f6f8"),
	surface:    ui.Hex("#ffffff"),
}

var darkPalette = palette{
	dark:       true,
	card:       ui.RGBA(30, 30, 34, 0.97),
	cardBorder: ui.RGBA(255, 255, 255, 0.09),
	shadow:     ui.RGBA(0, 0, 0, 0.5),
	shadow2:    ui.RGBA(0, 0, 0, 0.25),
	sidebar:    ui.RGBA(24, 24, 28, 0.9),
	divider:    ui.RGBA(255, 255, 255, 0.07),

	text:      ui.Hex("#ececf0"),
	textMuted: ui.Hex("#a3a3ae"),
	textFaint: ui.Hex("#73737f"),

	rowHover:    ui.RGBA(255, 255, 255, 0.05),
	rowSelected: ui.RGBA(255, 255, 255, 0.10),
	rowShadow:   ui.RGBA(0, 0, 0, 0.3),
	guide:       ui.RGBA(255, 255, 255, 0.12),

	field:      ui.RGBA(255, 255, 255, 0.06),
	fieldFocus: ui.RGBA(255, 255, 255, 0.09),
	chip:       ui.RGBA(255, 255, 255, 0.08),
	chipText:   ui.Hex("#b8b8c2"),
	kbd:        ui.RGBA(255, 255, 255, 0.08),
	kbdBorder:  ui.RGBA(255, 255, 255, 0.14),

	accent:     ui.Hex("#6b9bff"),
	accentSoft: ui.RGBA(107, 155, 255, 0.16),
	danger:     ui.Hex("#ff6b6b"),
	star:       ui.Hex("#ffc04d"),

	stage:      ui.RGBA(255, 255, 255, 0.025),
	frame:      ui.RGBA(255, 255, 255, 0.10),
	windowBack: ui.Hex("#1c1c20"),
	surface:    ui.Hex("#26262b"),
}

func paletteFor(c *ui.Context) *palette {
	if c.Theme().Dark {
		return &darkPalette
	}
	return &lightPalette
}
