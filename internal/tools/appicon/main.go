// Command appicon draws Windeck's icon: a deck of window cards with a
// tab at their side, flat, on the hz violet (#5A4BC6) of a macOS icon's
// rounded square. It writes resources/icon.png and the website's icons.
//
//	go run ./internal/tools/appicon
package main

import (
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/egoist/mygo/ui"
	"golang.org/x/image/draw"
)

var (
	brand       = ui.Hex("#5A4BC6")
	brandLight  = ui.Hex("#B5ACF1")
	brandSubtle = ui.Hex("#F4F3FD")
	muted       = ui.Hex("#E2E2EA")
)

func view(c *ui.Context) {
	c.Root().Background(ui.Transparent)
	// Apple's grid: the shape is 824 of 1024, centered, with the shadow
	// every icon of the Dock has.
	ui.Box(c).Fill().Center().Children(func() {
		ui.Box(c).Size(824, 824).Radius(186).Background(brand).
			Shadow(0, 14, 32, 0, ui.RGBA(28, 28, 35, 0.22)).
			Children(func() {
				card := func(x, y, w, h float32, alpha float32) {
					ui.Box(c).Absolute().Left(x).Top(y).Size(w, h).Radius(44).Background(ui.RGBA(255, 255, 255, alpha))
				}
				// The deck: two cards behind, fainter, and the front one.
				card(262, 170, 420, 330, 0.30)
				card(232, 214, 460, 360, 0.55)
				ui.Box(c).Absolute().Left(202).Top(264).Size(500, 392).Radius(48).Background(ui.Hex("#ffffff")).
					Children(func() {
						// A title bar's three dots and two lines of text.
						for i := range 3 {
							ui.Box(c).Absolute().Left(float32(76+i*36)).Top(40).Size(20, 20).Radius(10).Background(muted)
						}
						ui.Box(c).Absolute().Left(76).Top(104).Size(290, 26).Radius(13).Background(muted)
						ui.Box(c).Absolute().Left(76).Top(156).Size(210, 26).Radius(13).Background(muted)
						ui.Box(c).Absolute().Left(76).Top(208).Size(380, 120).Radius(22).Background(brandSubtle)
					})
				// The tab at the deck's side, as a bookmark's.
				ui.Box(c).Absolute().Left(140).Top(302).Size(104, 150).Radius(26).Background(brandLight)
			})
	})
}

func write(path string, img image.Image) {
	os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}

func scaled(src image.Image, n int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, n, n))
	draw.CatmullRom.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

// card is the site's share image, 1200 by 630: the icon, the name and
// the one sentence in the big bubble, on white.
func card(icon *ui.Bitmap) func(c *ui.Context) {
	return func(c *ui.Context) {
		c.Root().Background(ui.Hex("#FFFFFF"))
		ui.Row(c).Fill().Padding(96).Gap(64).AlignItems(ui.Center).Children(func() {
			ui.Image(c, icon).Size(280, 280).Shrink(0)
			ui.Column(c).Gap(24).Children(func() {
				ui.Text(c, "Windeck · 窗签").Font("Nunito").FontSize(64).FontWeight(700).TextColor(ui.Hex("#1C1C23"))
				ui.Box(c).Padding(16, 24).Radius(16, 16, 16, 4).Background(brandSubtle).Children(func() {
					ui.Text(c, "Switch windows like browser tabs.").Font("Inter").FontSize(32).FontWeight(500).TextColor(ui.Hex("#1C1C23"))
				})
				ui.Text(c, "A window switcher for macOS · macOS 窗口切换器").Font("Inter").FontSize(24).TextColor(ui.Hex("#5E5E6E"))
			})
		})
	}
}

func main() {
	for _, f := range []struct{ file, family string }{
		{"Inter-Regular.ttf", "Inter"}, {"Inter-Medium.ttf", "Inter"}, {"Inter-Bold.ttf", "Inter"}, {"Nunito.ttf", "Nunito"},
	} {
		if b, err := os.ReadFile(filepath.Join("resources/fonts", f.file)); err == nil {
			ui.RegisterFont(b, f.family)
		}
	}
	img := ui.Render(view, 1024, 1024, 1)
	write("resources/icon.png", img)
	write("website/public/icon.png", scaled(img, 512))
	write("website/public/favicon.png", scaled(img, 64))
	write("website/public/og.png", ui.Render(card(ui.NewBitmap(img)), 1200, 630, 1))
}
