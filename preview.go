package main

import (
	"slices"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"windeck/internal/wm"
)

// previewSide is the largest side of a window's picture, in pixels: the
// preview's size on a Retina display.
const previewSide = 1600

// maxBitmaps is how many pictures are kept as bitmaps, ready to draw: the
// windows Tab goes through.
const maxBitmaps = 10

// preview is a window's picture, as a bitmap, and when it was taken.
type preview struct {
	bmp *ui.Bitmap
	at  time.Time
}

// previewCache keeps the bitmaps of the pictures in the snapshot pool
// that were shown lately, and which windows are being captured.
type previewCache struct {
	byID      map[wm.WindowID]*preview
	order     []wm.WindowID // the most recently shown first
	capturing map[wm.WindowID]bool
}

// mayCapture reports whether Windeck may take a window's picture: the
// user allowed it, the app is not on the privacy list, and the window
// server has it to give.
func (a *App) mayCapture(w wm.Window) bool {
	return a.set.Previews && a.canCapture && !a.set.IsExcluded(w.BundleID) && !w.Minimized && !w.Hidden
}

// captureLater takes a window's picture on another goroutine, into the
// pool.
func (a *App) captureLater(w wm.Window) {
	if !a.mayCapture(w) || a.shots.capturing[w.ID] {
		return
	}
	if s, ok := a.pool.Get(uint32(w.ID)); ok && time.Since(s.At) < 400*time.Millisecond {
		return
	}
	a.shots.capturing[w.ID] = true
	go func() {
		img := wm.CaptureWindow(w.ID, previewSide)
		at := time.Now()
		a.update(func() {
			delete(a.shots.capturing, w.ID)
			if img == nil {
				return
			}
			if _, ok := a.deck.Window(w.ID); ok {
				a.pool.Put(uint32(w.ID), img, at)
			}
		})
	}()
}

// previewOf returns a window's picture as a bitmap, nil when there is
// none.
func (a *App) previewOf(id wm.WindowID) *ui.Bitmap {
	c := &a.shots
	shot, ok := a.pool.Get(uint32(id))
	if !ok {
		return nil
	}
	p := c.byID[id]
	if p == nil || !p.at.Equal(shot.At) {
		if p != nil {
			if w, h := p.bmp.Size(); w == shot.Image.Rect.Dx() && h == shot.Image.Rect.Dy() {
				p.bmp.Update(shot.Image)
				p.at = shot.At
			} else {
				p = nil
			}
		}
		if p == nil {
			p = &preview{bmp: ui.NewBitmap(shot.Image), at: shot.At}
			c.byID[id] = p
		}
	}
	c.order = slices.DeleteFunc(c.order, func(x wm.WindowID) bool { return x == id })
	c.order = slices.Insert(c.order, 0, id)
	for len(c.order) > maxBitmaps {
		delete(c.byID, c.order[len(c.order)-1])
		c.order = c.order[:len(c.order)-1]
	}
	return p.bmp
}

// keep forgets the bitmaps of windows that are gone.
func (c *previewCache) keep(a *App) {
	c.order = slices.DeleteFunc(c.order, func(id wm.WindowID) bool {
		if _, ok := a.deck.Window(id); !ok {
			delete(c.byID, id)
			return true
		}
		return false
	})
}

func (c *previewCache) clear() {
	c.byID = map[wm.WindowID]*preview{}
	c.order = nil
}

// iconCache is the apps' icons, by their bundles' paths and size.
type iconCache struct {
	byPath  map[string]*ui.Bitmap
	loading map[string]bool
}

// load fetches the icons of the windows' apps that are not fetched yet.
func (ic *iconCache) load(a *App, ws []wm.Window) {
	for _, w := range ws {
		a.icon(w.AppPath, 64)
	}
	for _, ap := range a.apps {
		if ap.Regular {
			a.icon(ap.Path, 64)
		}
	}
}

// icon returns an app's icon of px pixels, fetching it when it is not
// yet: nil until then.
func (a *App) icon(path string, px int) *ui.Bitmap {
	if path == "" {
		return nil
	}
	key := path + "@" + itoa(px)
	ic := &a.icons
	if b := ic.byPath[key]; b != nil || ic.loading[key] {
		return b
	}
	ic.loading[key] = true
	go func() {
		png, err := mygo.Shell.FileIcon(path, px)
		var b *ui.Bitmap
		if err == nil {
			b, _ = ui.DecodeBitmap(png)
		}
		a.update(func() {
			if b != nil {
				ic.byPath[key] = b
			}
		})
	}()
	return nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
