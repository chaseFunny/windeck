package wm

import (
	"sync/atomic"

	"github.com/ebitengine/purego"
)

// The switcher's shortcut: an event tap on the session's keyboard events,
// as only a tap sees the modifier go up, which switches, and can keep the
// keys the switcher handles from the app in front.

var (
	cgEventTapCreate   = sym(libCG, "CGEventTapCreate")
	cgEventTapEnable   = sym(libCG, "CGEventTapEnable")
	cgEventGetIntField = sym(libCG, "CGEventGetIntegerValueField")
	cgEventGetFlags    = sym(libCG, "CGEventGetFlags")
	tapCallback        = purego.NewCallback(tapped)
	tapPort            uintptr
	tapHandler         func(KeyEvent) bool
	tapMod             atomic.Int32
	tapCapturing       atomic.Bool
)

const (
	evKeyDown          = 10
	evFlagsChanged     = 12
	evTapDisabledTime  = 0xFFFFFFFE
	evTapDisabledInput = 0xFFFFFFFF

	flagShift   = 0x20000
	flagControl = 0x40000
	flagOption  = 0x80000
	flagCommand = 0x100000

	fieldKeycode    = 9
	fieldAutorepeat = 8
)

var keycodes = map[int64]Key{
	48: KeyTab, 53: KeyEscape, 49: KeySpace, 44: KeySlash, 36: KeyReturn, 76: KeyReturn,
	126: KeyUp, 125: KeyDown, 123: KeyLeft, 124: KeyRight, 13: KeyW, 46: KeyM, 12: KeyQ, 50: KeyGrave,
}

func modFlags() (mine, other uint64) {
	if Modifier(tapMod.Load()) == ModCommand {
		return flagCommand, flagOption | flagControl
	}
	return flagOption, flagCommand | flagControl
}

func tapped(_ uintptr, typ uintptr, ev uintptr, _ uintptr) uintptr {
	switch uint32(typ) {
	case evTapDisabledTime, evTapDisabledInput:
		// macOS turns a tap off that took too long: on again.
		call(cgEventTapEnable, tapPort, 1)
		return ev
	case evFlagsChanged:
		if !tapCapturing.Load() {
			return ev
		}
		mine, _ := modFlags()
		flags := uint64(call(cgEventGetFlags, ev))
		if flags&mine == 0 {
			tapHandler(KeyEvent{Released: true})
		}
		return ev
	case evKeyDown:
		flags := uint64(call(cgEventGetFlags, ev))
		mine, other := modFlags()
		code := int64(call(cgEventGetIntField, ev, fieldKeycode))
		k := keycodes[code]
		e := KeyEvent{
			Key:    k,
			Shift:  flags&flagShift != 0,
			Repeat: call(cgEventGetIntField, ev, fieldAutorepeat) != 0,
			Mod:    flags&mine != 0,
			Other:  flags&other != 0,
		}
		// The shortcut itself, or any key while the switcher has it.
		if !tapCapturing.Load() && !(k == KeyTab && e.Mod && !e.Other) {
			return ev
		}
		if tapHandler(e) {
			return 0
		}
	}
	return ev
}

// StartTap starts reporting the shortcut (the modifier and Tab) and, while
// capturing, every key and the modifier going up, to fn, which returns
// whether to keep the key from the app in front. Call it on the main
// thread, whose run loop the tap runs on, once.
func StartTap(mod Modifier, fn func(KeyEvent) bool) error {
	tapHandler = fn
	SetTapModifier(mod)
	if tapPort != 0 {
		return nil
	}
	mask := uint64(1<<evKeyDown | 1<<evFlagsChanged)
	// kCGSessionEventTap, kCGHeadInsertEventTap, kCGEventTapOptionDefault
	port := call(cgEventTapCreate, 1, 0, 0, uintptr(mask), tapCallback, 0)
	if port == 0 {
		return ErrNotTrusted
	}
	tapPort = port
	addToMainRunLoop(call(cfMachPortCreateRLSource, 0, port, 0))
	call(cgEventTapEnable, port, 1)
	return nil
}

// SetTapModifier changes the shortcut's modifier.
func SetTapModifier(mod Modifier) { tapMod.Store(int32(mod)) }

// SetCapturing has the tap report every key, and the modifier going up,
// while the switcher shows.
func SetCapturing(on bool) { tapCapturing.Store(on) }

// ModifierDown reports whether the shortcut's modifier is held now.
func ModifierDown() bool {
	mine, _ := modFlags()
	// kCGEventSourceStateCombinedSessionState
	return uint64(call(cgEventSourceFlagsState, 0))&mine != 0
}

var cgEventSourceFlagsState = sym(libCG, "CGEventSourceFlagsState")
