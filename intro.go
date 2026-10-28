package main

import (
	"math"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/shader"

	"windeck/internal/wm"
	"windeck/locales"
)

// The first launch's intro, over the whole screen with only a breathing
// aurora behind it: "Windeck" is written out by hand; each letter turns
// into a window card, the cards stack into the logo's deck, the violet
// square opens behind them and the tab comes out at the side; the name
// rises under the logo; then the logo flies to its place in what comes
// next (the permissions' window, or the switcher under its guide), which
// takes over at the same spot. A click, Return or Escape skips it. All of
// it is drawn with the painter and a shader, from the time since it began.

// Timeline, in milliseconds from the first frame.
const (
	inGlowIn      = 700
	inWriteStart  = 350
	inLetterGap   = 170
	inLetterDur   = 600
	inCardStart   = 2250 // the letters become cards, one after another
	inCardStagger = 45
	inCardDur     = 420
	inStackStart  = 2500 // the cards fly into the deck
	inStackDur    = 760
	inSquareStart = 3050 // the violet square opens behind them
	inSquareDur   = 700
	inDetailStart = 3500 // the front card's title bar and lines
	inMarkStart   = 3650 // the name under the logo
	inExpand      = 4700 // the logo flies to its place
	inExpandDur   = 700
	inTotal       = inExpand + inExpandDur
	// The next window shows as the logo comes to rest: showing one holds
	// the main thread for a frame or two, which a moving logo would show.
	inHandOver = inExpandDur - 90
)

// Sizes, in points.
const (
	inWordScale = 2.2
	inLogoSide  = 200 // the violet square's side
	inLogoRise  = 40  // the logo sits this far above the middle, the name under it
	inBreath    = 3.4 // seconds a breath of the aurora
	iconGrid    = 824 // the square's side in the icon's 1024 grid
	iconOffset  = 100 // and where it starts
)

// The logo's deck, in the square's 824 units (internal/tools/appicon):
// the two cards behind, the front one, and the tab.
var deckRects = [4]struct{ x, y, w, h, r float32 }{
	{162, 70, 420, 330, 44},
	{132, 114, 460, 360, 44},
	{102, 164, 500, 392, 48},
	{40, 202, 104, 150, 26},
}

// letterSlot is the card each letter of "Windeck" becomes: W and i the
// back card, n and d the middle one, e and c the front one, k the tab.
var letterSlot = [7]int{0, 0, 1, 1, 2, 2, 3}

// aurora is the background: a soft veil behind the lettering, lighter
// (darker, in dark mode) than the desktop, so it reads on any wallpaper,
// and violet lights drifting around it, all breathing slowly; clear
// toward the screen's edges.
var aurora = shader.MustCompile(`package main

var Time float
var Level float
var Breath float
var Center vec2
var Dark float
var Colors [5]vec4

func Fragment(dst vec4, pos vec2) vec4 {
	size := rectSize()
	uv := (pos - Center) / size.y

	v := length(uv * vec2(0.55, 1.0))
	veil := exp(-v*v*9.0) * (0.5 + 0.12*Breath)

	glow := vec3(0)
	w := 0.0
	for i := 0; i < 5; i++ {
		f := float(i)
		c := vec2(cos(Time*0.33+f*1.2566), sin(Time*0.27+f*2.1)) * vec2(0.42, 0.2)
		d := length((uv - c) * vec2(0.7, 1.0))
		k := exp(-d * d * 11.0)
		glow += Colors[i].rgb * k
		w += k
	}
	m := length(uv * vec2(0.5, 0.9))
	mask := exp(-m * m * 4.0)
	ga := clamp(w, 0.0, 1.0) * (0.2 + 0.2*Breath) * mask
	gc := glow / max(w, 0.0001)

	veilColor := mix(vec3(1.0), vec3(0.07, 0.07, 0.09), Dark)
	rgb := gc*ga + veilColor*veil*(1.0-ga)
	a := ga + veil*(1.0-ga)
	return vec4(rgb, a) * Level
}
`)

// auroraColors are the hz violets, light to deep.
var auroraColors = []ui.Color{lightPalette.brand, darkPalette.brand, darkPalette.brandText, lightPalette.brandHover, darkPalette.brandOnSub}

type introState struct {
	win      *mygo.Window
	shownAt  time.Time
	start    time.Time // the first frame's time
	finished bool
	handed   bool
	// toOnboarding: the logo lands on the permissions' window's, else it
	// fades into the switcher.
	toOnboarding bool
	// land is where the logo lands in the window, and side the square's
	// side there.
	land     ui.Point
	landSide float32
	display  mygo.Display

	word   []glyph
	offset []float32
	boxes  []box
}

// showIntro plays the intro on the display with the pointer, then goes
// on to the permissions or the switcher.
func (a *App) showIntro() {
	if a.intro.win != nil {
		return // playing
	}
	in := &a.intro
	*in = introState{word: wordGlyphs()}
	var x float32
	for _, g := range in.word {
		in.offset = append(in.offset, x)
		in.boxes = append(in.boxes, glyphBox(g))
		x += g.advance
	}
	d := mygo.Screen.DisplayNearestPoint(mygo.Screen.CursorScreenPoint())
	in.display = d
	b, wa := d.Bounds, d.WorkArea
	in.toOnboarding = !wm.Trusted(false)
	if in.toOnboarding {
		// The onboarding window, centered in the work area, shows its
		// 64-point icon 64 from its top: the square is 824/1024 of it.
		x, y := onboardingOrigin(d)
		in.land = ui.Point{X: float32(x - b.X + onboardW/2), Y: float32(y - b.Y + 64 + 32)}
		in.landSide = 64 * iconGrid / 1024
	} else {
		// Into the switcher, which shows in the middle.
		in.land = ui.Point{X: float32(wa.X - b.X + wa.Width/2), Y: float32(wa.Y - b.Y + wa.Height/2)}
		in.landSide = 24
	}
	in.win = mygo.NewWindow(mygo.WindowOptions{
		Title:         "Windeck",
		Panel:         true,
		Hidden:        true,
		Frameless:     true,
		Transparent:   true,
		DisableShadow: true,
		DisableResize: true,
		Width:         b.Width,
		Height:        b.Height,
		Content:       ui.View(a.introView),
	})
	in.win.SetPosition(b.X, b.Y)
	in.win.Show()
	in.win.Focus()
	in.shownAt = time.Now()
	// In case no frames come (the window covered), or they stop.
	time.AfterFunc(2*time.Second, func() {
		a.update(func() {
			if a.intro.start.IsZero() {
				a.finishIntro()
			}
		})
	})
	time.AfterFunc(inTotal*time.Millisecond+2500*time.Millisecond, func() { a.update(a.finishIntro) })
}

// introHandOver shows what comes after the intro, under it.
func (a *App) introHandOver() {
	in := &a.intro
	if in.handed {
		return
	}
	in.handed = true
	if in.toOnboarding {
		a.ob.introLogo = true
		a.ob.display = &in.display
		a.openOnboarding()
		return
	}
	a.set.Onboarded = true
	a.saveSettings()
	if a.tray != nil {
		a.tray.SetMenu(a.trayMenu())
	}
	a.showGuide()
}

// finishIntro ends the intro, skipped or played: what comes next shows,
// with its own logo back, and the intro closes over it.
func (a *App) finishIntro() {
	in := &a.intro
	if in.finished {
		return
	}
	in.finished = true
	a.introHandOver()
	a.ob.introLogo = false
	if a.onboardWin != nil {
		a.onboardWin.Invalidate()
	}
	win := in.win
	time.AfterFunc(140*time.Millisecond, func() {
		a.update(func() {
			if win != nil {
				win.Close()
			}
			if a.intro.win == win {
				a.intro.win = nil
			}
		})
	})
}

func (a *App) introView(c *ui.Context) {
	pal := themed(c)
	c.Root().Background(ui.Transparent)
	if c.Shortcut(0, ui.KeyEscape) || c.Shortcut(0, ui.KeyEnter) {
		a.update(a.finishIntro)
	}
	stage := ui.Box(c).Fill()
	if stage.Clicked() {
		a.update(a.finishIntro)
	}
	stage.Draw(func(p *ui.Painter, r ui.Rect) {
		in := &a.intro
		if in.start.IsZero() {
			in.start = p.Now()
		}
		ms := float32(p.Now().Sub(in.start).Milliseconds())
		if ms < inTotal {
			p.AnimationFrame()
		} else {
			a.update(a.finishIntro)
		}
		if ms >= inExpand+inHandOver && !in.handed {
			a.update(a.introHandOver)
		}
		a.paintIntro(p, r, pal, ms)
	})
}

// progress is how far ms is through [from, from+dur], from 0 to 1.
func progress(ms, from, dur float32) float32 { return max(0, min(1, (ms-from)/dur)) }

func easeOutCubic(t float32) float32 { u := 1 - t; return 1 - u*u*u }

func easeInOutCubic(t float32) float32 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	u := -2*t + 2
	return 1 - u*u*u/2
}

// easeOutBack overshoots a little before it settles.
func easeOutBack(t float32) float32 {
	const c1, c3 = 1.4, 2.4
	u := t - 1
	return 1 + c3*u*u*u + c1*u*u
}

func lerp(a, b, t float32) float32 { return a + (b-a)*t }

func lerpRect(a, b ui.Rect, t float32) ui.Rect {
	return ui.Rect{X: lerp(a.X, b.X, t), Y: lerp(a.Y, b.Y, t), W: lerp(a.W, b.W, t), H: lerp(a.H, b.H, t)}
}

// bump rises from 0 to 1 and back over [from, from+dur].
func bump(ms, from, dur float32) float32 {
	return float32(math.Sin(float64(progress(ms, from, dur)) * math.Pi))
}

// paintIntro draws the intro at ms into r, the window.
func (a *App) paintIntro(p *ui.Painter, r ui.Rect, pal *palette, ms float32) {
	in := &a.intro
	b, wa := in.display.Bounds, in.display.WorkArea
	cx := r.X + float32(wa.X-b.X) + float32(wa.Width)/2
	cy := r.Y + float32(wa.Y-b.Y) + float32(wa.Height)/2
	ex := easeInOutCubic(progress(ms, inExpand, inExpandDur))

	// The aurora: in, a swell as the logo settles, out as it flies.
	if p.Accelerated() {
		secs := float64(ms) / 1000
		breath := 0.5 + 0.5*math.Sin(secs*2*math.Pi/inBreath-math.Pi/2)
		level := easeOutCubic(progress(ms, 0, inGlowIn)) * (1 + 0.35*bump(ms, inSquareStart+inSquareDur-250, 900)) *
			(1 - easeOutCubic(progress(ms, inExpand, inExpandDur*0.8)))
		dark := float32(0)
		if pal.dark {
			dark = 1
		}
		colors := make([]float32, 0, 4*len(auroraColors))
		for _, c := range auroraColors {
			colors = append(colors, float32(c.R)/255, float32(c.G)/255, float32(c.B)/255, 1)
		}
		aurora.Draw(p, r, shader.Options{Uniforms: map[string]any{
			"Time": float32(secs) + 2, "Level": level, "Breath": float32(breath),
			"Center": []float32{cx - r.X, cy - r.Y - inLogoRise/2}, "Dark": dark, "Colors": colors,
		}})
	}

	// The logo's place and size, flying at the end.
	side := float32(inLogoSide) * (1 + 0.04*bump(ms, inSquareStart+inSquareDur-120, 420))
	lx, ly := cx, cy-inLogoRise
	side = lerp(side, in.landSide, ex)
	lx = lerp(lx, r.X+in.land.X, ex)
	ly = lerp(ly, r.Y+in.land.Y, ex)
	fade := float32(1)
	if !in.toOnboarding {
		fade = 1 - easeOutCubic(progress(ms, inExpand+inExpandDur*0.4, inExpandDur*0.6))
	}
	u := side / iconGrid
	ox, oy := lx-side/2, ly-side/2
	slot := func(i int) ui.Rect {
		d := deckRects[i]
		return ui.Rect{X: ox + d.x*u, Y: oy + d.y*u, W: d.w * u, H: d.h * u}
	}

	// The violet square, opening behind the deck with a little bounce.
	sq := progress(ms, inSquareStart, inSquareDur)
	if sq > 0 {
		s := side * (0.5 + 0.5*easeOutBack(sq))
		alpha := easeOutCubic(min(1, sq*2.5)) * fade
		p.Fill(ui.Rect{X: lx - s/2, Y: ly - s/2, W: s, H: s}, pal.brand.Alpha(alpha), 186*s/iconGrid)
	}
	white := lightPalette.bgPage
	slotColor := [4]ui.Color{white.Alpha(0.30), white.Alpha(0.55), white, darkPalette.brandText}
	turn := easeOutCubic(progress(ms, inSquareStart+120, inSquareDur-120))

	// The word, written letter by letter, drawing slowly nearer; then each
	// letter gives way to its card, and the cards fly into the deck.
	ws := float32(inWordScale) * (1 + 0.035*easeOutCubic(progress(ms, inWriteStart, 2400)))
	ww := wordWidth(in.word) * ws
	word := xform{s: ws, x: cx - ww/2, y: cy - inLogoRise + wordXHeight/2*ws}
	stroke := wordStroke * ws
	ink := pal.text
	// Cards are drawn back to front: the slot's order, the tab last.
	var cards [7]struct {
		rect   ui.Rect
		radius float32
		color  ui.Color
		show   bool
	}
	for i, g := range in.word {
		f := xform{s: ws, x: word.x + in.offset[i]*ws, y: word.y}
		card := easeOutCubic(progress(ms, float32(inCardStart+i*inCardStagger), inCardDur))
		if drawn := easeInOutCubic(progress(ms, float32(inWriteStart+i*inLetterGap), inLetterDur)); drawn > 0 && card < 1 {
			var path ui.Path
			for _, s := range g.strokes {
				s.appendTo(&path, f)
			}
			p.StrokePath(path.Trim(0, drawn), stroke, ink.Alpha(1-card))
			if g.dot != nil {
				if pop := progress(ms, float32(inWriteStart+i*inLetterGap+inLetterDur-140), 260); pop > 0 {
					x, y := f.at(*g.dot)
					var dot ui.Path
					p.FillPath(dot.Circle(x, y, stroke*0.62*easeOutBack(pop)), ink.Alpha(1-card))
				}
			}
		}
		if card <= 0 {
			continue
		}
		// The card starts as the letter's box, growing out of it.
		bx := in.boxes[i]
		pad := 4 * ws
		x0, y0 := f.at(pt{bx.x, bx.y})
		from := ui.Rect{X: x0 - pad, Y: y0 - pad, W: bx.w*ws + 2*pad, H: bx.h*ws + 2*pad}
		grow := 0.6 + 0.4*card
		from = ui.Rect{X: from.X + from.W*(1-grow)/2, Y: from.Y + from.H*(1-grow)/2, W: from.W * grow, H: from.H * grow}
		st := easeInOutCubic(progress(ms, float32(inStackStart+i*35), inStackDur))
		k := letterSlot[i]
		cards[i].rect = lerpRect(from, slot(k), st)
		cards[i].radius = lerp(8*ws/2.2, deckRects[k].r*u, st)
		cards[i].color = pal.brand.Mix(slotColor[k], turn).Alpha(card * fade)
		cards[i].show = true
	}
	for k := range 4 {
		for i := range cards {
			if cards[i].show && letterSlot[i] == k {
				p.Fill(cards[i].rect, cards[i].color, cards[i].radius)
			}
		}
		if k == 2 {
			// The front card's title bar and lines, once it is white.
			if det := easeOutCubic(progress(ms, inDetailStart, 400)) * fade; det > 0 {
				fr := slot(2)
				at := func(x, y, w, h, rad float32, c ui.Color) {
					p.Fill(ui.Rect{X: fr.X + x*u, Y: fr.Y + y*u, W: w * u, H: h * u}, c.Alpha(det), rad*u)
				}
				muted := lightPalette.border
				for d := range 3 {
					at(float32(76+d*36), 40, 20, 20, 10, muted)
				}
				at(76, 104, 290*det, 26, 13, muted)
				at(76, 156, 210*det, 26, 13, muted)
				at(76, 208, 380, 120, 22, lightPalette.brandSubtle)
			}
		}
	}

	// The name and the line under the logo, until it flies.
	mark := easeOutCubic(progress(ms, inMarkStart, 420)) * (1 - easeOutCubic(progress(ms, inExpand, 260)))
	if mark > 0 {
		name := ui.Span{Text: locales.AppName(), Font: fontDisplay, Size: 36, Weight: 700, Color: ink.Alpha(mark)}
		w, _ := p.MeasureText(0, name)
		top := cy - inLogoRise + inLogoSide/2 + 28 + 10*(1-mark)
		p.RichText(cx-w/2, top, 0, name)
		tag := ui.Span{Text: locales.SplashTagline(), Size: 15, Weight: 500, Color: pal.textMuted.Alpha(mark)}
		w, _ = p.MeasureText(0, tag)
		p.RichText(cx-w/2, top+54, 0, tag)
	}
}
