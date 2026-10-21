// Command appicon draws Windeck's icon: a deck of window cards with a
// bookmark tab at their side, on the blue of a macOS icon's rounded
// square. It writes resources/icon.png and the website's icons.
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

func view(c *ui.Context) {
	c.Root().Background(ui.Transparent)
	// Apple's grid: the shape is 824 of 1024, centered.
	ui.Box(c).Fill().Center().Children(func() {
		ui.Box(c).Size(824, 824).Radius(186).
			LinearGradient(ui.LinearGradient{Angle: 160, Oklab: true, Stops: []ui.GradientStop{
				{Color: ui.Hex("#5b8cff"), At: 0}, {Color: ui.Hex("#2f6bff"), At: 0.55}, {Color: ui.Hex("#1f4fd8"), At: 1},
			}}).
			Shadow(0, 18, 40, 0, ui.RGBA(10, 30, 90, 0.35)).
			InsetShadow(0, 2, 0, 0, ui.RGBA(255, 255, 255, 0.25)).
			Children(func() {
				card := func(x, y, w, h float32, alpha float32) {
					ui.Box(c).Absolute().Left(x).Top(y).Size(w, h).Radius(44).
						Background(ui.RGBA(255, 255, 255, alpha)).
						Shadow(0, 14, 30, 0, ui.RGBA(10, 30, 90, 0.28))
				}
				// The deck: two cards behind, fainter, and the front one.
				card(250, 168, 420, 330, 0.42)
				card(222, 212, 460, 360, 0.68)
				ui.Box(c).Absolute().Left(192).Top(262).Size(500, 392).Radius(48).
					Background(ui.Hex("#ffffff")).
					Shadow(0, 20, 40, 0, ui.RGBA(10, 30, 90, 0.32)).
					Children(func() {
						// A title bar's three dots and two lines of text.
						for i := range 3 {
							ui.Box(c).Absolute().Left(float32(76+i*36)).Top(40).Size(20, 20).Radius(10).Background(ui.Hex("#d6dcea"))
						}
						ui.Box(c).Absolute().Left(76).Top(104).Size(290, 26).Radius(13).Background(ui.Hex("#e6eaf3"))
						ui.Box(c).Absolute().Left(76).Top(156).Size(210, 26).Radius(13).Background(ui.Hex("#eef1f7"))
						ui.Box(c).Absolute().Left(76).Top(208).Size(380, 120).Radius(22).Background(ui.Hex("#f1f4fa"))
					})
				// The tab at the deck's side, as a bookmark's.
				ui.Box(c).Absolute().Left(132).Top(300).Size(104, 150).Radius(26).
					LinearGradient(ui.LinearGradient{Angle: 180, Oklab: true, Stops: []ui.GradientStop{
						{Color: ui.Hex("#ffd166"), At: 0}, {Color: ui.Hex("#ffb020"), At: 1},
					}}).
					Shadow(0, 10, 22, 0, ui.RGBA(120, 70, 0, 0.35))
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

func main() {
	img := ui.Render(view, 1024, 1024, 1)
	write("resources/icon.png", img)
	write("website/public/icon.png", scaled(img, 512))
	write("website/public/favicon.png", scaled(img, 64))
}
