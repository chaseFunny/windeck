package main

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// The lettering of the first launch's intro: "Windeck" written as single
// strokes, a rounded monoline hand leaning forward, which the intro writes
// out and then turns, letter by letter, into the cards of the logo's deck.

type pt struct{ X, Y float32 }

func lerpPt(a, b pt, t float32) pt { return pt{a.X + (b.X-a.X)*t, a.Y + (b.Y-a.Y)*t} }

// cubic is a Bézier segment after the current point.
type cubic struct{ c1, c2, to pt }

// stroke is a continuous line: a start and the segments after it.
type stroke struct {
	from pt
	segs []cubic
}

// glyph is a letter's strokes, drawn one after another, and the dot of
// an i.
type glyph struct {
	strokes []stroke
	advance float32
	dot     *pt
}

type pen struct{ s stroke }

func penAt(x, y float32) *pen { return &pen{s: stroke{from: pt{x, y}}} }

func (p *pen) last() pt {
	if n := len(p.s.segs); n > 0 {
		return p.s.segs[n-1].to
	}
	return p.s.from
}

// line is a straight segment, kept as a cubic so that every segment
// samples alike.
func (p *pen) line(x, y float32) *pen {
	a, b := p.last(), pt{x, y}
	p.s.segs = append(p.s.segs, cubic{lerpPt(a, b, 1.0/3), lerpPt(a, b, 2.0/3), b})
	return p
}

func (p *pen) curve(c1x, c1y, c2x, c2y, x, y float32) *pen {
	p.s.segs = append(p.s.segs, cubic{pt{c1x, c1y}, pt{c2x, c2y}, pt{x, y}})
	return p
}

// kappa draws a quarter of an ellipse with a cubic.
const kappa = 0.5523

// arc adds a quarter of an ellipse from the current point to (x, y);
// horizontalFirst says it leaves the top or the bottom.
func (p *pen) arc(x, y float32, horizontalFirst bool) *pen {
	a := p.last()
	dx, dy := x-a.X, y-a.Y
	if horizontalFirst {
		return p.curve(a.X+dx*kappa, a.Y, x, y-dy*kappa, x, y)
	}
	return p.curve(a.X, a.Y+dy*kappa, x-dx*kappa, y, x, y)
}

func (p *pen) done() stroke { return p.s }

// Letter metrics, in letter units: the baseline at 0, the x-height at
// -40, capitals at -62, ascenders at -64.
const (
	wordXHeight = 40
	wordStroke  = 6.2
	wordSlant   = 0.12 // the letters lean forward, as handwriting does
)

// wordGlyphs is "Windeck".
func wordGlyphs() []glyph {
	w := glyph{strokes: []stroke{penAt(1, -62).line(10, -2).
		curve(11, 1, 13, 1, 14, -2).line(21, -44).line(28, -2).
		curve(29, 1, 31, 1, 32, -2).line(41, -62).done()}, advance: 48}
	dot := pt{4, -55}
	i := glyph{strokes: []stroke{penAt(4, -38).line(4, 0).done()}, advance: 15, dot: &dot}
	n := glyph{strokes: []stroke{
		penAt(2, -40).line(2, 0).done(),
		penAt(2, -25).curve(4.5, -36, 10, -40, 16.5, -40).curve(25, -40, 29, -34.5, 29, -25).line(29, 0).done(),
	}, advance: 38}
	d := glyph{strokes: []stroke{
		penAt(29, -20).arc(15, -40, false).arc(1, -20, true).arc(15, 0, false).arc(29, -20, true).done(),
		penAt(29, -64).line(29, 0).done(),
	}, advance: 38}
	e := glyph{strokes: []stroke{penAt(2, -20).line(30, -20).
		curve(30, -32, 24, -40, 16, -40).curve(7, -40, 1, -31, 1, -20).
		curve(1, -8, 8, 0, 17, 0).curve(23, 0, 27, -3, 30, -7).done()}, advance: 37}
	c := glyph{strokes: []stroke{penAt(29, -32).
		curve(26, -37, 21, -40, 16, -40).curve(7, -40, 1, -31, 1, -20).
		curve(1, -8, 8, 0, 17, 0).curve(23, 0, 27, -3, 30, -8).done()}, advance: 36}
	k := glyph{strokes: []stroke{
		penAt(3, -64).line(3, 0).done(),
		penAt(28, -40).line(4, -16).done(),
		penAt(12, -24).line(29, 0).done(),
	}, advance: 32}
	gs := []glyph{w, i, n, d, e, c, k}
	lean := func(p pt) pt { return pt{p.X - p.Y*wordSlant, p.Y} }
	for gi := range gs {
		g := &gs[gi]
		for si := range g.strokes {
			s := &g.strokes[si]
			s.from = lean(s.from)
			for ci := range s.segs {
				c := &s.segs[ci]
				c.c1, c.c2, c.to = lean(c.c1), lean(c.c2), lean(c.to)
			}
		}
		if g.dot != nil {
			d := lean(*g.dot)
			g.dot = &d
		}
	}
	return gs
}

// wordWidth is the word's advance, in letter units, the last letter's
// spacing left out.
func wordWidth(gs []glyph) float32 {
	var w float32
	for _, g := range gs {
		w += g.advance
	}
	return w - 4
}

// xform maps letter units to the window: scale, then offset.
type xform struct{ s, x, y float32 }

func (f xform) at(p pt) (float32, float32) { return f.x + p.X*f.s, f.y + p.Y*f.s }

func (s stroke) appendTo(path *ui.Path, f xform) {
	path.MoveTo(f.at(s.from))
	for _, c := range s.segs {
		x1, y1 := f.at(c.c1)
		x2, y2 := f.at(c.c2)
		x, y := f.at(c.to)
		path.CubeTo(x1, y1, x2, y2, x, y)
	}
}

func (c cubic) eval(from pt, t float32) pt {
	u := 1 - t
	a, b, d, e := u*u*u, 3*u*u*t, 3*u*t*t, t*t*t
	return pt{a*from.X + b*c.c1.X + d*c.c2.X + e*c.to.X, a*from.Y + b*c.c1.Y + d*c.c2.Y + e*c.to.Y}
}

// box is a rectangle in letter units.
type box struct{ x, y, w, h float32 }

// glyphBox is what a letter's strokes and dot cover, in letter units.
func glyphBox(g glyph) box {
	x0, y0 := float32(math.MaxFloat32), float32(math.MaxFloat32)
	x1, y1 := -x0, -y0
	add := func(p pt) {
		x0, y0, x1, y1 = min(x0, p.X), min(y0, p.Y), max(x1, p.X), max(y1, p.Y)
	}
	for _, s := range g.strokes {
		cur := s.from
		add(cur)
		for _, c := range s.segs {
			for k := 1; k <= 12; k++ {
				add(c.eval(cur, float32(k)/12))
			}
			cur = c.to
		}
	}
	if g.dot != nil {
		add(*g.dot)
	}
	return box{x0, y0, x1 - x0, y1 - y0}
}
