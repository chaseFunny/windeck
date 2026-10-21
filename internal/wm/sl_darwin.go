package wm

import (
	"image"
	"unsafe"

	"github.com/ebitengine/purego"
)

// The window server: Core Graphics' window list and displays, and
// SkyLight's Spaces, captures and focus, which no public API offers.

var (
	cgWindowListCopyInfo = sym(libCG, "CGWindowListCopyWindowInfo")
	cgRectFromDict       = sym(libCG, "CGRectMakeWithDictionaryRepresentation")
	cgGetActiveDisplays  = sym(libCG, "CGGetActiveDisplayList")
	cgDisplayIsBuiltin   = sym(libCG, "CGDisplayIsBuiltin")
	cgMainDisplayID      = sym(libCG, "CGMainDisplayID")
	cgPreflightCapture   = sym(libCG, "CGPreflightScreenCaptureAccess")
	cgRequestCapture     = sym(libCG, "CGRequestScreenCaptureAccess")
	cgImageGetWidth      = sym(libCG, "CGImageGetWidth")
	cgImageGetHeight     = sym(libCG, "CGImageGetHeight")
	cgColorSpaceSRGB     = sym(libCG, "CGColorSpaceCreateWithName")
	cgColorSpaceRelease  = sym(libCG, "CGColorSpaceRelease")
	cgBitmapContext      = sym(libCG, "CGBitmapContextCreate")
	cgContextRelease     = sym(libCG, "CGContextRelease")
	cgContextInterp      = sym(libCG, "CGContextSetInterpolationQuality")
	kCGColorSpaceSRGB    = global(libCG, "kCGColorSpaceSRGB")

	slsMainConnectionID     = sym(libSL, "SLSMainConnectionID")
	slsCopySpacesForWindows = sym(libSL, "SLSCopySpacesForWindows")
	slsCopyDisplaySpaces    = sym(libSL, "SLSCopyManagedDisplaySpaces")
	slsHWCaptureWindowList  = sym(libSL, "SLSHWCaptureWindowList")
	slpsSetFrontProcess     = sym(libSL, "_SLPSSetFrontProcessWithOptions")
	slpsPostEventRecordTo   = sym(libSL, "SLPSPostEventRecordTo")
	cgsSetSymbolicHotKey    = sym(libSL, "CGSSetSymbolicHotKeyEnabled")
	getProcessForPID        = sym(libAS, "GetProcessForPID")

	cgContextDrawImage func(ctx uintptr, rect cgRect, image uintptr)
	cgDisplayBounds    func(display uint32) cgRect
)

type cgRect struct{ X, Y, W, H float64 }

func init() {
	purego.RegisterLibFunc(&cgContextDrawImage, libCG, "CGContextDrawImage")
	purego.RegisterLibFunc(&cgDisplayBounds, libCG, "CGDisplayBounds")
}

func connection() uintptr { return call(slsMainConnectionID) }

// CanCapture reports whether the user allowed Windeck to record the
// screen, which the previews need; with prompt, macOS asks them to.
func CanCapture(prompt bool) bool {
	if call(cgPreflightCapture)&0xff != 0 {
		return true
	}
	if prompt {
		return call(cgRequestCapture)&0xff != 0
	}
	return false
}

// cgWindow is a window as the window server lists it.
type cgWindow struct {
	id       WindowID
	pid      int32
	layer    int
	title    string
	bounds   Rect
	onscreen bool
	alpha    float64
}

const (
	cgListAll            = 0
	cgListExcludeDesktop = 1 << 4
	cgNullWindow         = 0
	minimumWindowSize    = 50
)

// cgWindows lists the windows of every app, on every Space, front to back
// for those that show.
func cgWindows() []cgWindow {
	arr := call(cgWindowListCopyInfo, cgListAll|cgListExcludeDesktop, cgNullWindow)
	if arr == 0 {
		return nil
	}
	defer release(arr)
	n := arrayCount(arr)
	out := make([]cgWindow, 0, n)
	for i := range n {
		d := arrayAt(arr, i)
		layer, _ := goInt(dictGet(d, "kCGWindowLayer"))
		if layer != 0 {
			continue
		}
		id, _ := goInt(dictGet(d, "kCGWindowNumber"))
		pid, _ := goInt(dictGet(d, "kCGWindowOwnerPID"))
		w := cgWindow{
			id:       WindowID(id),
			pid:      int32(pid),
			title:    goString(dictGet(d, "kCGWindowName")),
			onscreen: goBool(dictGet(d, "kCGWindowIsOnscreen")),
			alpha:    goFloat(dictGet(d, "kCGWindowAlpha")),
		}
		var r cgRect
		if b := dictGet(d, "kCGWindowBounds"); b != 0 && call(cgRectFromDict, b, uintptr(unsafe.Pointer(&r)))&0xff != 0 {
			w.bounds = Rect{r.X, r.Y, r.W, r.H}
		}
		out = append(out, w)
	}
	return out
}

// Displays lists the connected displays, the main one first.
func Displays() []Display {
	var ids [16]uint32
	var n uint32
	call(cgGetActiveDisplays, 16, uintptr(unsafe.Pointer(&ids[0])), uintptr(unsafe.Pointer(&n)))
	main := uint32(call(cgMainDisplayID))
	out := make([]Display, 0, n)
	for _, id := range ids[:n] {
		r := cgDisplayBounds(id)
		d := Display{ID: id, Bounds: Rect{r.X, r.Y, r.W, r.H}, BuiltIn: call(cgDisplayIsBuiltin, uintptr(id))&0xff != 0, Main: id == main}
		if d.Main {
			out = append([]Display{d}, out...)
		} else {
			out = append(out, d)
		}
	}
	return out
}

// displayOf returns the index of the display that holds most of r.
func displayOf(ds []Display, r Rect) int {
	best, area := 0, -1.0
	for i, d := range ds {
		x0, y0 := max(r.X, d.Bounds.X), max(r.Y, d.Bounds.Y)
		x1, y1 := min(r.X+r.W, d.Bounds.X+d.Bounds.W), min(r.Y+r.H, d.Bounds.Y+d.Bounds.H)
		if a := max(x1-x0, 0) * max(y1-y0, 0); a > area {
			best, area = i, a
		}
	}
	return best
}

// spaceInfo is a Space: the desktop number Mission Control gives it on its
// display (0 for a full-screen app's), and whether it shows now.
type spaceInfo struct {
	number     int
	fullscreen bool
	current    bool
}

// spaces describes the Spaces of every display, by id.
func spaces() map[uint64]spaceInfo {
	out := map[uint64]spaceInfo{}
	if slsCopyDisplaySpaces == 0 {
		return out
	}
	arr := call(slsCopyDisplaySpaces, connection())
	if arr == 0 {
		return out
	}
	defer release(arr)
	for i := range arrayCount(arr) {
		d := arrayAt(arr, i)
		cur, _ := goInt(dictGet(dictGet(d, "Current Space"), "ManagedSpaceID"))
		list := dictGet(d, "Spaces")
		desktop := 0
		for j := range arrayCount(list) {
			s := arrayAt(list, j)
			id, _ := goInt(dictGet(s, "ManagedSpaceID"))
			typ, _ := goInt(dictGet(s, "type"))
			info := spaceInfo{current: id == cur, fullscreen: typ == 4}
			if typ == 0 {
				desktop++
				info.number = desktop
			}
			out[uint64(id)] = info
		}
	}
	return out
}

// spaceOf returns the Space a window is on, 0 when it is on none, as a
// minimized window, or on all of them.
func spaceOf(id WindowID) uint64 {
	if slsCopySpacesForWindows == 0 {
		return 0
	}
	num := newCFNumber32(int32(id))
	ids := newCFArray([]uintptr{num})
	release(num)
	defer release(ids)
	arr := call(slsCopySpacesForWindows, connection(), 7, ids)
	if arr == 0 {
		return 0
	}
	defer release(arr)
	if arrayCount(arr) != 1 {
		return 0
	}
	v, _ := goInt(arrayAt(arr, 0))
	return uint64(v)
}

// CaptureWindow returns a window's picture, scaled down to fit maxSide
// pixels; nil when the window server has none, as for a window minimized
// before Windeck started.
func CaptureWindow(id WindowID, maxSide int) Capture {
	if slsHWCaptureWindowList == 0 {
		return nil
	}
	wid := uint32(id)
	// kCGSCaptureIgnoreGlobalClipShape | kCGSCaptureBestResolution: the
	// window whole, behind others and on other Spaces, at its pixels.
	arr := call(slsHWCaptureWindowList, connection(), uintptr(unsafe.Pointer(&wid)), 1, 1<<11|1<<8)
	if arr == 0 {
		return nil
	}
	defer release(arr)
	if arrayCount(arr) == 0 {
		return nil
	}
	return toRGBA(arrayAt(arr, 0), maxSide)
}

// toRGBA draws a CGImage into an RGBA image no larger than maxSide.
func toRGBA(img uintptr, maxSide int) *image.RGBA {
	w, h := int(call(cgImageGetWidth, img)), int(call(cgImageGetHeight, img))
	if w <= 1 || h <= 1 {
		return nil
	}
	if s := max(w, h); maxSide > 0 && s > maxSide {
		w, h = max(w*maxSide/s, 1), max(h*maxSide/s, 1)
	}
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	cs := call(cgColorSpaceSRGB, kCGColorSpaceSRGB)
	defer call(cgColorSpaceRelease, cs)
	// kCGImageAlphaPremultipliedLast | kCGBitmapByteOrder32Big: RGBA, as
	// image.RGBA keeps it.
	ctx := call(cgBitmapContext, uintptr(unsafe.Pointer(&out.Pix[0])), uintptr(w), uintptr(h), 8, uintptr(out.Stride), cs, 1|4<<12)
	if ctx == 0 {
		return nil
	}
	defer call(cgContextRelease, ctx)
	call(cgContextInterp, ctx, 3) // kCGInterpolationHigh
	cgContextDrawImage(ctx, cgRect{0, 0, float64(w), float64(h)}, img)
	return out
}

type psn struct{ high, low uint32 }

// focusWindow makes a window the key window of its app and its app the
// front one, going to its Space, as AltTab does: the window server's own
// calls, which the public API has no counterpart of.
func focusWindow(pid int32, id WindowID) bool {
	var p psn
	if getProcessForPID == 0 || int32(call(getProcessForPID, uintptr(pid), uintptr(unsafe.Pointer(&p)))) != 0 {
		return false
	}
	// kCPSUserGenerated
	if int32(call(slpsSetFrontProcess, uintptr(unsafe.Pointer(&p)), uintptr(id), 0x200)) != 0 {
		return false
	}
	var rec [0xf8]byte
	rec[0x04] = 0xf8
	rec[0x3a] = 0x10
	*(*uint32)(unsafe.Pointer(&rec[0x3c])) = uint32(id)
	for i := 0x20; i < 0x30; i++ {
		rec[i] = 0xff
	}
	rec[0x08] = 0x01
	call(slpsPostEventRecordTo, uintptr(unsafe.Pointer(&p)), uintptr(unsafe.Pointer(&rec[0])))
	rec[0x08] = 0x02
	call(slpsPostEventRecordTo, uintptr(unsafe.Pointer(&p)), uintptr(unsafe.Pointer(&rec[0])))
	return true
}

// SetSystemSwitcher turns macOS's own Cmd-Tab and Cmd-Shift-Tab on or
// off, for Windeck to take them over.
func SetSystemSwitcher(on bool) {
	if cgsSetSymbolicHotKey == 0 {
		return
	}
	v := uintptr(0)
	if on {
		v = 1
	}
	call(cgsSetSymbolicHotKey, 1, v) // Cmd-Tab
	call(cgsSetSymbolicHotKey, 2, v) // Cmd-Shift-Tab
}
