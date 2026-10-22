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
