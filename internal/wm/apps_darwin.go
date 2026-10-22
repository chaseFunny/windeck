package wm

import (
	"runtime"
	"strings"
	"unsafe"

	"github.com/ebitengine/purego/objc"
)

// Running apps, through NSWorkspace and NSRunningApplication.

var (
	poolPush = sym(libObj, "objc_autoreleasePoolPush")
	poolPop  = sym(libObj, "objc_autoreleasePoolPop")

	selSharedWorkspace = objc.RegisterName("sharedWorkspace")
	selRunningApps     = objc.RegisterName("runningApplications")
	selCount           = objc.RegisterName("count")
	selObjectAtIndex   = objc.RegisterName("objectAtIndex:")
	selPolicy          = objc.RegisterName("activationPolicy")
	selPID             = objc.RegisterName("processIdentifier")
	selLocalizedName   = objc.RegisterName("localizedName")
	selBundleID        = objc.RegisterName("bundleIdentifier")
	selBundleURL       = objc.RegisterName("bundleURL")
	selExecutableURL   = objc.RegisterName("executableURL")
	selPath            = objc.RegisterName("path")
	selIsHidden        = objc.RegisterName("isHidden")
	selUTF8String      = objc.RegisterName("UTF8String")
	selAppWithPID      = objc.RegisterName("runningApplicationWithProcessIdentifier:")
	selActivate        = objc.RegisterName("activateWithOptions:")
	selHide            = objc.RegisterName("hide")
	selUnhide          = objc.RegisterName("unhide")
	selTerminate       = objc.RegisterName("terminate")
	selFrontmostApp    = objc.RegisterName("frontmostApplication")
	selIsTerminated    = objc.RegisterName("isTerminated")

	classWorkspace  = class("NSWorkspace")
	classRunningApp = class("NSRunningApplication")
)

// withPool runs fn in an autorelease pool, on one thread: a pool belongs
// to the thread that pushed it.
func withPool(fn func()) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	p := call(poolPush)
	defer call(poolPop, p)
	fn()
}

func nsString(id objc.ID) string {
	if id == 0 {
		return ""
	}
	p := objc.Send[uintptr](id, selUTF8String)
	if p == 0 {
		return ""
	}
	base := ptr(p)
	n := 0
	for *(*byte)(unsafe.Add(base, n)) != 0 {
		n++
	}
	return string(unsafe.Slice((*byte)(base), n))
}

func workspace() objc.ID { return classWorkspace.Send(selSharedWorkspace) }

// appOf describes a running application object.
func appOf(app objc.ID) App {
	a := App{
		PID:      objc.Send[int32](app, selPID),
		Name:     nsString(app.Send(selLocalizedName)),
		BundleID: nsString(app.Send(selBundleID)),
		Hidden:   objc.Send[bool](app, selIsHidden),
	}
	if u := app.Send(selBundleURL); u != 0 {
		a.Path = nsString(u.Send(selPath))
	} else if u := app.Send(selExecutableURL); u != 0 {
		a.Path = nsString(u.Send(selPath))
	}
	return a
}

// RunningApps returns the apps that may have windows to switch to, in the
// order they launched: those of the Dock, and accessory apps, as a menu
// bar app with a settings window.
func RunningApps() []App {
	var out []App
	withPool(func() {
		arr := workspace().Send(selRunningApps)
		n := objc.Send[uint](arr, selCount)
		for i := range n {
			app := arr.Send(selObjectAtIndex, i)
			// NSApplicationActivationPolicyProhibited (2) apps have no UI.
			policy := objc.Send[int](app, selPolicy)
			if policy == 2 {
				continue
			}
			a := appOf(app)
			a.Regular = policy == 0
			if a.PID <= 0 {
				continue
			}
			out = append(out, a)
		}
	})
	return out
}

// FrontmostPID returns the process of the app the user is in.
func FrontmostPID() int32 {
	var pid int32
	withPool(func() {
		if app := workspace().Send(selFrontmostApp); app != 0 {
			pid = objc.Send[int32](app, selPID)
		}
	})
	return pid
}

func withApp(pid int32, fn func(app objc.ID)) {
	withPool(func() {
		if app := classRunningApp.Send(selAppWithPID, pid); app != 0 {
			fn(app)
		}
	})
}

// ActivateApp brings an app to the front, unhiding it.
func ActivateApp(pid int32) {
	withApp(pid, func(app objc.ID) {
		if objc.Send[bool](app, selIsHidden) {
			app.Send(selUnhide)
		}
		// NSApplicationActivateIgnoringOtherApps
		objc.Send[bool](app, selActivate, uint(2))
	})
}

// QuitApp asks an app to quit, as its Quit menu item does.
func QuitApp(pid int32) {
	withApp(pid, func(app objc.ID) { objc.Send[bool](app, selTerminate) })
}

// HideApp hides an app, as Cmd-H does.
func HideApp(pid int32) {
	withApp(pid, func(app objc.ID) { objc.Send[bool](app, selHide) })
}

var (
	selURLForBundle   = objc.RegisterName("URLForApplicationWithBundleIdentifier:")
	selDefaultManager = objc.RegisterName("defaultManager")
	selDisplayName    = objc.RegisterName("displayNameAtPath:")

	classFileManager = class("NSFileManager")
)

// InstalledApp finds the app of a bundle ID on this Mac, running or not:
// its name as Finder shows it, and its bundle. ok is false when it is not
// installed.
func InstalledApp(bundleID string) (name, path string, ok bool) {
	if bundleID == "" {
		return "", "", false
	}
	withPool(func() {
		u := workspace().Send(selURLForBundle, objc.ID(cfstr(bundleID)))
		if u == 0 {
			return
		}
		path = nsString(u.Send(selPath))
		if path == "" {
			return
		}
		ns := objc.ID(cfstr(path))
		name = nsString(classFileManager.Send(selDefaultManager).Send(selDisplayName, ns))
		name = strings.TrimSuffix(name, ".app")
		ok = true
	})
	return name, path, ok
}
