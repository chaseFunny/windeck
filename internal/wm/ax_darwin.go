package wm

import (
	"encoding/binary"
	"unsafe"

	"github.com/ebitengine/purego"
)

// The accessibility API: what lists an app's windows, says whether one
// is minimized or full screen, and raises, closes and minimizes it.

var (
	axCreateApplication  = sym(libAS, "AXUIElementCreateApplication")
	axCopyAttributeValue = sym(libAS, "AXUIElementCopyAttributeValue")
	axSetAttributeValue  = sym(libAS, "AXUIElementSetAttributeValue")
	axPerformAction      = sym(libAS, "AXUIElementPerformAction")
	axGetWindow          = sym(libAS, "_AXUIElementGetWindow")
	axCreateWithToken    = sym(libAS, "_AXUIElementCreateWithRemoteToken")
	axIsProcessTrusted   = sym(libAS, "AXIsProcessTrusted")
	axIsTrustedWithOpts  = sym(libAS, "AXIsProcessTrustedWithOptions")
	axObserverCreate     = sym(libAS, "AXObserverCreate")
	axObserverAddNotif   = sym(libAS, "AXObserverAddNotification")
	axObserverGetSource  = sym(libAS, "AXObserverGetRunLoopSource")

	axSetMessagingTimeout func(el uintptr, seconds float32) int32
)

func init() {
	purego.RegisterLibFunc(&axSetMessagingTimeout, libAS, "AXUIElementSetMessagingTimeout")
}

const axErrSuccess = 0

// axAttr returns an owned value of an element's attribute, 0 when it has
// none or the app does not answer.
func axAttr(el uintptr, name string) uintptr {
	var out uintptr
	if int32(call(axCopyAttributeValue, el, cfstr(name), uintptr(unsafe.Pointer(&out)))) != axErrSuccess {
		return 0
	}
	return out
}

func axString(el uintptr, name string) string {
	v := axAttr(el, name)
	defer release(v)
	return goString(v)
}

func axBool(el uintptr, name string) bool {
	v := axAttr(el, name)
	defer release(v)
	return goBool(v)
}

func axSetBool(el uintptr, name string, b bool) bool {
	v := cfBooleanFalse
	if b {
		v = cfBooleanTrue
	}
	return int32(call(axSetAttributeValue, el, cfstr(name), v)) == axErrSuccess
}

func axAction(el uintptr, action string) bool {
	return int32(call(axPerformAction, el, cfstr(action))) == axErrSuccess
}

// axWindowID returns the window server's number of a window element.
func axWindowID(el uintptr) WindowID {
	var id uint32
	if int32(call(axGetWindow, el, uintptr(unsafe.Pointer(&id)))) != axErrSuccess {
		return 0
	}
	return WindowID(id)
}

// newAXApp returns an owned element of an app, which gives up on an app
// that does not answer within a tenth of a second: a hung app must never
// hold the switcher up.
func newAXApp(pid int32) uintptr {
	el := call(axCreateApplication, uintptr(pid))
	if el != 0 {
		axSetMessagingTimeout(el, 0.1)
	}
	return el
}

// axWindowsOf returns owned elements of an app's windows on the current
// Spaces, minimized ones included.
func axWindowsOf(app uintptr) []uintptr {
	arr := axAttr(app, "AXWindows")
	if arr == 0 {
		return nil
	}
	defer release(arr)
	n := arrayCount(arr)
	out := make([]uintptr, 0, n)
	for i := range n {
		out = append(out, retain(arrayAt(arr, i)))
	}
	return out
}

// axWindowsByToken finds the elements of an app's windows on other
// Spaces, which AXWindows leaves out, by asking for the elements of the
// app's first ids one by one, as AltTab does. want are the windows to
// find: it stops once it has them all. The elements are owned.
func axWindowsByToken(pid int32, want map[WindowID]bool) map[WindowID]uintptr {
	found := map[WindowID]uintptr{}
	if axCreateWithToken == 0 || len(want) == 0 {
		return found
	}
	token := make([]byte, 20)
	binary.LittleEndian.PutUint32(token[0:], uint32(pid))
	binary.LittleEndian.PutUint32(token[4:], 0)
	binary.LittleEndian.PutUint32(token[8:], 0x636f636f) // "coco"
	for id := uint64(0); id < 1000 && len(found) < len(want); id++ {
		binary.LittleEndian.PutUint64(token[12:], id)
		data := newCFData(token)
		el := call(axCreateWithToken, data)
		release(data)
		if el == 0 {
			continue
		}
		if wid := axWindowID(el); wid != 0 && want[wid] && found[wid] == 0 {
			if sub := axString(el, "AXSubrole"); sub == "AXStandardWindow" || sub == "AXDialog" {
				found[wid] = el
				continue
			}
		}
		release(el)
	}
	return found
}

// Trusted reports whether the user allowed Windeck to control the
// computer (Privacy & Security › Accessibility); with prompt, macOS asks
// them to, once.
func Trusted(prompt bool) bool {
	if !prompt {
		return call(axIsProcessTrusted)&0xff != 0
	}
	opts := newCFDict(cfstr("AXTrustedCheckOptionPrompt"), cfBooleanTrue)
	defer release(opts)
	return call(axIsTrustedWithOpts, opts)&0xff != 0
}
