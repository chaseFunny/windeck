package main

import (
	_ "embed"

	"github.com/egoist/mygo/ui"
)

//go:embed resources/icon.png
var logoPNG []byte

var logoBitmap *ui.Bitmap

// appLogo is Windeck's icon, for its own windows.
func (a *App) appLogo() *ui.Bitmap {
	if logoBitmap == nil {
		logoBitmap, _ = ui.DecodeBitmap(logoPNG)
	}
	return logoBitmap
}
