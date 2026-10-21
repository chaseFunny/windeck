package wm

import (
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// Core Foundation, Core Graphics, the accessibility API and SkyLight, the
// window server's private framework that AltTab and the like call too,
// loaded through purego: Windeck builds with CGO_ENABLED=0.

func dlopen(path string) uintptr {
	h, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic("windeck: " + err.Error())
	}
	return h
}

var (
	libCF     = dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation")
	libCG     = dlopen("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics")
	libAS     = dlopen("/System/Library/Frameworks/ApplicationServices.framework/ApplicationServices")
	libSL     = dlopen("/System/Library/PrivateFrameworks/SkyLight.framework/SkyLight")
	libObj    = dlopen("/usr/lib/libobjc.A.dylib")
	libAppKit = dlopen("/System/Library/Frameworks/AppKit.framework/AppKit")
)

// class returns an Objective-C class, once AppKit, which has most of them,
// is loaded: naming libAppKit orders the package's initialization so.
func class(name string) objc.ID {
	_ = libAppKit
	return objc.ID(objc.GetClass(name))
}

// sym returns a function's address, 0 when the system has none.
func sym(lib uintptr, name string) uintptr {
	a, err := purego.Dlsym(lib, name)
	if err != nil {
		return 0
	}
	return a
}

// global reads a pointer-sized constant a framework exports, as
// kCFBooleanTrue.
func global(lib uintptr, name string) uintptr {
	a := sym(lib, name)
	if a == 0 {
		return 0
	}
	return *(*uintptr)(ptr(a))
}

// ptr turns an address C returned into a pointer, as go vet accepts.
func ptr(a uintptr) unsafe.Pointer { return *(*unsafe.Pointer)(unsafe.Pointer(&a)) }

func call(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, args...)
	return r
}

var (
	cfRelease                = sym(libCF, "CFRelease")
	cfRetain                 = sym(libCF, "CFRetain")
	cfArrayGetCount          = sym(libCF, "CFArrayGetCount")
	cfArrayGetValueAtIndex   = sym(libCF, "CFArrayGetValueAtIndex")
	cfArrayCreate            = sym(libCF, "CFArrayCreate")
	cfDictionaryGetValue     = sym(libCF, "CFDictionaryGetValue")
	cfDictionaryCreate       = sym(libCF, "CFDictionaryCreate")
	cfStringCreateWithBytes  = sym(libCF, "CFStringCreateWithBytes")
	cfStringGetLength        = sym(libCF, "CFStringGetLength")
	cfStringGetCString       = sym(libCF, "CFStringGetCString")
	cfStringGetMaxSize       = sym(libCF, "CFStringGetMaximumSizeForEncoding")
	cfNumberGetValue         = sym(libCF, "CFNumberGetValue")
	cfNumberCreate           = sym(libCF, "CFNumberCreate")
	cfGetTypeID              = sym(libCF, "CFGetTypeID")
	cfStringGetTypeID        = sym(libCF, "CFStringGetTypeID")
	cfNumberGetTypeID        = sym(libCF, "CFNumberGetTypeID")
	cfBooleanGetTypeID       = sym(libCF, "CFBooleanGetTypeID")
	cfArrayGetTypeID         = sym(libCF, "CFArrayGetTypeID")
	cfBooleanGetValue        = sym(libCF, "CFBooleanGetValue")
	cfDataCreate             = sym(libCF, "CFDataCreate")
	cfRunLoopGetMain         = sym(libCF, "CFRunLoopGetMain")
	cfRunLoopAddSource       = sym(libCF, "CFRunLoopAddSource")
	cfRunLoopRemoveSource    = sym(libCF, "CFRunLoopRemoveSource")
	cfMachPortCreateRLSource = sym(libCF, "CFMachPortCreateRunLoopSource")

	cfBooleanTrue         = global(libCF, "kCFBooleanTrue")
	cfBooleanFalse        = global(libCF, "kCFBooleanFalse")
	cfRunLoopCommonModes  = global(libCF, "kCFRunLoopCommonModes")
	cfTypeArrayCallBacks  = sym(libCF, "kCFTypeArrayCallBacks")
	cfTypeDictKeyCallBack = sym(libCF, "kCFTypeDictionaryKeyCallBacks")
	cfTypeDictValCallBack = sym(libCF, "kCFTypeDictionaryValueCallBacks")
)

const (
	cfStringEncodingUTF8 = 0x08000100
	cfNumberSInt32Type   = 3
	cfNumberSInt64Type   = 4
	cfNumberDoubleType   = 13
)

func release(ref uintptr) {
	if ref != 0 {
		call(cfRelease, ref)
	}
}

func retain(ref uintptr) uintptr {
	if ref != 0 {
		call(cfRetain, ref)
	}
	return ref
}

var (
	strMu    sync.Mutex
	strCache = map[string]uintptr{}
)

// cfstr returns a CFString of s that lives as long as the process: for the
// constant keys and attribute names, made once.
func cfstr(s string) uintptr {
	strMu.Lock()
	defer strMu.Unlock()
	if r, ok := strCache[s]; ok {
		return r
	}
	r := newCFString(s)
	strCache[s] = r
	return r
}

// newCFString returns an owned CFString of s.
func newCFString(s string) uintptr {
	b := []byte(s)
	var p uintptr
	if len(b) > 0 {
		p = uintptr(unsafe.Pointer(&b[0]))
	}
	return call(cfStringCreateWithBytes, 0, p, uintptr(len(b)), cfStringEncodingUTF8, 0)
}

func typeID(ref uintptr) uintptr { return call(cfGetTypeID, ref) }

var (
	stringType = call(cfStringGetTypeID)
	numberType = call(cfNumberGetTypeID)
	boolType   = call(cfBooleanGetTypeID)
	arrayType  = call(cfArrayGetTypeID)
)

// goString converts a CFString, "" for anything else.
func goString(ref uintptr) string {
	if ref == 0 || typeID(ref) != stringType {
		return ""
	}
	n := call(cfStringGetLength, ref)
	if n == 0 {
		return ""
	}
	size := call(cfStringGetMaxSize, n, cfStringEncodingUTF8) + 1
	buf := make([]byte, size)
	if call(cfStringGetCString, ref, uintptr(unsafe.Pointer(&buf[0])), size, cfStringEncodingUTF8)&0xff == 0 {
		return ""
	}
	for i, c := range buf {
		if c == 0 {
			return string(buf[:i])
		}
	}
	return string(buf)
}

// goInt converts a CFNumber, or a CFBoolean as 0 or 1.
func goInt(ref uintptr) (int64, bool) {
	if ref == 0 {
		return 0, false
	}
	switch typeID(ref) {
	case numberType:
		var v int64
		call(cfNumberGetValue, ref, cfNumberSInt64Type, uintptr(unsafe.Pointer(&v)))
		return v, true
	case boolType:
		return int64(call(cfBooleanGetValue, ref) & 0xff), true
	}
	return 0, false
}

func goFloat(ref uintptr) float64 {
	if ref == 0 || typeID(ref) != numberType {
		return 0
	}
	var v float64
	call(cfNumberGetValue, ref, cfNumberDoubleType, uintptr(unsafe.Pointer(&v)))
	return v
}

func goBool(ref uintptr) bool {
	v, _ := goInt(ref)
	return v != 0
}

func arrayCount(arr uintptr) int {
	if arr == 0 || typeID(arr) != arrayType {
		return 0
	}
	return int(call(cfArrayGetCount, arr))
}

func arrayAt(arr uintptr, i int) uintptr { return call(cfArrayGetValueAtIndex, arr, uintptr(i)) }

func dictGet(dict uintptr, key string) uintptr {
	if dict == 0 {
		return 0
	}
	return call(cfDictionaryGetValue, dict, cfstr(key))
}

// newCFNumber32 returns an owned CFNumber.
func newCFNumber32(v int32) uintptr {
	return call(cfNumberCreate, 0, cfNumberSInt32Type, uintptr(unsafe.Pointer(&v)))
}

// newCFArray returns an owned CFArray of the values, which it retains.
func newCFArray(values []uintptr) uintptr {
	var p uintptr
	if len(values) > 0 {
		p = uintptr(unsafe.Pointer(&values[0]))
	}
	return call(cfArrayCreate, 0, p, uintptr(len(values)), cfTypeArrayCallBacks)
}

// newCFDict returns an owned dictionary of one key and value.
func newCFDict(key, value uintptr) uintptr {
	k, v := [1]uintptr{key}, [1]uintptr{value}
	return call(cfDictionaryCreate, 0, uintptr(unsafe.Pointer(&k[0])), uintptr(unsafe.Pointer(&v[0])), 1,
		cfTypeDictKeyCallBack, cfTypeDictValCallBack)
}

func newCFData(b []byte) uintptr {
	return call(cfDataCreate, 0, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)))
}

// addToMainRunLoop adds a run loop source to the main thread's run loop,
// in the common modes, so that it fires during menus and drags too.
func addToMainRunLoop(source uintptr) {
	call(cfRunLoopAddSource, call(cfRunLoopGetMain), source, cfRunLoopCommonModes)
}

func removeFromMainRunLoop(source uintptr) {
	call(cfRunLoopRemoveSource, call(cfRunLoopGetMain), source, cfRunLoopCommonModes)
}
