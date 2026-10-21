package wm

import (
	"os"
	"testing"
	"time"
)

// TestLive lists this computer's windows and captures the first: run with
// WINDECK_LIVE=1 in a desktop session, with the permissions given.
func TestLive(t *testing.T) {
	if os.Getenv("WINDECK_LIVE") == "" {
		t.Skip("WINDECK_LIVE not set")
	}
	start := time.Now()
	ws := Windows()
	t.Logf("%d windows in %v (trusted %v, capture %v)", len(ws), time.Since(start), Trusted(false), CanCapture(false))
	for _, w := range ws {
		t.Logf("%6d %-20s %-40q min=%v hid=%v fs=%v space=%d cur=%v disp=%d ext=%v", w.ID, w.AppName, w.Title, w.Minimized, w.Hidden, w.FullScreen, w.Space, w.OnCurrentSpace, w.Display, w.External)
	}
	start = time.Now()
	t.Logf("second listing: %d in %v", len(Windows()), time.Since(start))
	t.Logf("stacking: %v focused: %v", StackingOrder(), FocusedWindow())
	for _, w := range ws {
		start = time.Now()
		img := CaptureWindow(w.ID, 1600)
		if img != nil {
			t.Logf("captured %d (%s): %v in %v", w.ID, w.AppName, img.Bounds(), time.Since(start))
		} else {
			t.Logf("no capture for %d (%s, min=%v)", w.ID, w.AppName, w.Minimized)
		}
	}
	t.Logf("displays: %+v", Displays())
}
