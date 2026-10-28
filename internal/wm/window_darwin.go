package wm

import "github.com/ebitengine/purego/objc"

// Windeck's own windows.

var (
	selContentView     = objc.RegisterName("contentView")
	selSubviews        = objc.RegisterName("subviews")
	selSetWantsLayer   = objc.RegisterName("setWantsLayer:")
	selLayer           = objc.RegisterName("layer")
	selSetCornerRadius = objc.RegisterName("setCornerRadius:")
	selSetCornerCurve  = objc.RegisterName("setCornerCurve:")
	selSetMasksToBound = objc.RegisterName("setMasksToBounds:")
	selIsKindOfClass   = objc.RegisterName("isKindOfClass:")
	selInvalidShadow   = objc.RegisterName("invalidateShadow")
	selSetAnimation    = objc.RegisterName("setAnimationBehavior:")

	classEffectView = class("NSVisualEffectView")
)

// RoundWindow cuts a window, its content and the material behind it, to
// corners of radius with the continuous curve of the system's panels:
// the shadow, which macOS makes of what the window draws, follows them.
// handle is the NSWindow; call it on the main thread.
func RoundWindow(handle uintptr, radius float64) {
	if handle == 0 {
		return
	}
	withPool(func() {
		win := objc.ID(handle)
		view := win.Send(selContentView)
		if view == 0 {
			return
		}
		round := func(v objc.ID) {
			v.Send(selSetWantsLayer, true)
			l := v.Send(selLayer)
			if l == 0 {
				return
			}
			l.Send(selSetCornerRadius, radius)
			l.Send(selSetCornerCurve, objc.ID(cfstr("continuous")))
			l.Send(selSetMasksToBound, true)
		}
		round(view)
		subs := view.Send(selSubviews)
		for i := range objc.Send[uint](subs, selCount) {
			if v := subs.Send(selObjectAtIndex, i); objc.Send[bool](v, selIsKindOfClass, classEffectView) {
				round(v)
			}
		}
		win.Send(selInvalidShadow)
	})
}

// NoWindowAnimation has a window show and hide at once, without the fade
// macOS gives panels: the switcher comes and goes with the keys.
func NoWindowAnimation(handle uintptr) {
	if handle == 0 {
		return
	}
	// NSWindowAnimationBehaviorNone
	objc.ID(handle).Send(selSetAnimation, 2)
}

var (
	selInitTracking     = objc.RegisterName("initWithRect:options:owner:userInfo:")
	selAddTrackingArea  = objc.RegisterName("addTrackingArea:")
	classTrackingArea   = class("NSTrackingArea")
	trackedPointerViews = map[objc.ID]bool{}
)

type nsRect struct{ X, Y, W, H float64 }

// TrackPointerAlways has a window's content hear the pointer while
// Windeck is not the active app. MyGo tracks it in the active app only,
// and the switcher's panel never activates Windeck: without this, its
// rows never see the pointer come, and show no buttons. handle is the
// NSWindow; call it on the main thread, once its content is made.
func TrackPointerAlways(handle uintptr) {
	// MyGo registers its view's class as it makes the first window.
	surface := class("MyGoSurfaceView")
	if handle == 0 || classTrackingArea == 0 || surface == 0 {
		return
	}
	withPool(func() {
		var visit func(v objc.ID)
		visit = func(v objc.ID) {
			if v == 0 {
				return
			}
			if objc.Send[bool](v, selIsKindOfClass, surface) && !trackedPointerViews[v] {
				// Entered and exited, moved, always, in the visible rect.
				const options = 0x01 | 0x02 | 0x80 | 0x200
				area := classTrackingArea.Send(selAlloc)
				area = area.Send(selInitTracking, nsRect{}, uint(options), v, objc.ID(0))
				v.Send(selAddTrackingArea, area)
				trackedPointerViews[v] = true
			}
			subs := v.Send(selSubviews)
			for i := range objc.Send[uint](subs, selCount) {
				visit(subs.Send(selObjectAtIndex, i))
			}
		}
		visit(objc.ID(handle).Send(selContentView))
	})
}
