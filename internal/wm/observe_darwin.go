package wm

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// What changes, from NSWorkspace's notifications (apps launched, quit,
// activated, hidden; the Space changed) and accessibility observers on
// every app (a window focused, created, closed, minimized, retitled).
// Nothing is polled.

var (
	onChange  func(Change)
	lastFocus atomic.Uint32

	observersMu sync.Mutex
	observers   = map[int32]observer{}

	axCallback = purego.NewCallback(axNotified)
)

type observer struct{ ref, source uintptr }

func emit(c Change) {
	if onChange != nil {
		onChange(c)
	}
}

// noteFocus records that a window came to the front.
func noteFocus(id WindowID) {
	if id == 0 {
		return
	}
	prev := WindowID(lastFocus.Swap(uint32(id)))
	if prev != id {
		emit(Change{Event: EventFocus, Window: id, Leaving: prev})
	}
}

var axNotifications = []string{
	"AXFocusedWindowChanged", "AXMainWindowChanged", "AXWindowCreated", "AXUIElementDestroyed",
	"AXWindowMiniaturized", "AXWindowDeminiaturized", "AXApplicationHidden", "AXApplicationShown",
	"AXTitleChanged",
}

func axNotified(_, el, notification, refcon uintptr) uintptr {
	pid := int32(refcon)
	switch goString(notification) {
	case "AXFocusedWindowChanged", "AXMainWindowChanged":
		if FrontmostPID() == pid {
			noteFocus(axWindowID(el))
		}
	case "AXTitleChanged":
		// Titles change all the time, as a terminal's: only windows'.
		if axWindowID(el) != 0 {
			emit(Change{Event: EventChanged, PID: pid, Window: axWindowID(el)})
		}
	default:
		emit(Change{Event: EventChanged, PID: pid})
	}
	return 0
}

// observeApp follows an app's windows. An app that just launched may not
// answer yet: it is asked again a few times.
func observeApp(pid int32) {
	go func() {
		for _, wait := range []time.Duration{0, 500 * time.Millisecond, 2 * time.Second, 5 * time.Second} {
			time.Sleep(wait)
			if tryObserve(pid) {
				return
			}
		}
	}()
}

func tryObserve(pid int32) bool {
	observersMu.Lock()
	_, ok := observers[pid]
	observersMu.Unlock()
	if ok {
		return true
	}
	var ref uintptr
	if int32(call(axObserverCreate, uintptr(pid), axCallback, uintptr(unsafe.Pointer(&ref)))) != axErrSuccess || ref == 0 {
		return false
	}
	app := newAXApp(pid)
	defer release(app)
	added := 0
	for _, n := range axNotifications {
		if int32(call(axObserverAddNotif, ref, app, cfstr(n), uintptr(pid))) == axErrSuccess {
			added++
		}
	}
	if added == 0 {
		release(ref)
		return false
	}
	src := call(axObserverGetSource, ref)
	addToMainRunLoop(src)
	observersMu.Lock()
	observers[pid] = observer{ref, src}
	observersMu.Unlock()
	return true
}

func unobserveApp(pid int32) {
	observersMu.Lock()
	o, ok := observers[pid]
	delete(observers, pid)
	observersMu.Unlock()
	if ok {
		removeFromMainRunLoop(o.source)
		release(o.ref)
	}
}

var (
	selDefaultCenter = objc.RegisterName("notificationCenter")
	selAddObserver   = objc.RegisterName("addObserver:selector:name:object:")
	selNotify        = objc.RegisterName("windeckNotify:")
	selName          = objc.RegisterName("name")
	selUserInfo      = objc.RegisterName("userInfo")
	selObjectForKey  = objc.RegisterName("objectForKey:")
	selAlloc         = objc.RegisterName("alloc")
	selInit          = objc.RegisterName("init")
	selStringWith    = objc.RegisterName("stringWithUTF8String:")
	classNSString    = class("NSString")

	workspaceObserver objc.ID
)

func nsstr(s string) objc.ID {
	b := append([]byte(s), 0)
	return classNSString.Send(selStringWith, unsafe.Pointer(&b[0]))
}

func workspaceNotified(_ objc.ID, _ objc.SEL, note objc.ID) {
	name := nsString(note.Send(selName))
	var pid int32
	if info := note.Send(selUserInfo); info != 0 {
		if app := info.Send(selObjectForKey, nsstr("NSWorkspaceApplicationKey")); app != 0 {
			pid = objc.Send[int32](app, selPID)
		}
	}
	switch name {
	case "NSWorkspaceDidLaunchApplicationNotification":
		observeApp(pid)
		emit(Change{Event: EventChanged, PID: pid})
	case "NSWorkspaceDidTerminateApplicationNotification":
		unobserveApp(pid)
		emit(Change{Event: EventChanged, PID: pid})
	case "NSWorkspaceDidActivateApplicationNotification":
		go func() {
			if id := focusedWindowOf(pid); id != 0 {
				noteFocus(id)
			} else if prev := WindowID(lastFocus.Swap(0)); prev != 0 {
				// An app without windows: the user left the window.
				emit(Change{Event: EventFocus, Leaving: prev})
			}
		}()
	case "NSWorkspaceActiveSpaceDidChangeNotification":
		emit(Change{Event: EventSpace})
		emit(Change{Event: EventChanged})
	default:
		emit(Change{Event: EventChanged, PID: pid})
	}
}

func focusedWindowOf(pid int32) WindowID {
	ax := newAXApp(pid)
	if ax == 0 {
		return 0
	}
	defer release(ax)
	el := axAttr(ax, "AXFocusedWindow")
	if el == 0 {
		return 0
	}
	defer release(el)
	return axWindowID(el)
}

// Observe reports what changes to fn, from the main thread or from other
// goroutines. Call it once, on the main thread, with the app running.
func Observe(fn func(Change)) error {
	onChange = fn
	class, err := objc.RegisterClass("WindeckWorkspaceObserver", objc.GetClass("NSObject"), nil, nil, []objc.MethodDef{
		{Cmd: selNotify, Fn: workspaceNotified},
	})
	if err != nil {
		return err
	}
	workspaceObserver = objc.ID(class).Send(selAlloc).Send(selInit)
	center := workspace().Send(selDefaultCenter)
	for _, name := range []string{
		"NSWorkspaceDidLaunchApplicationNotification", "NSWorkspaceDidTerminateApplicationNotification",
		"NSWorkspaceDidActivateApplicationNotification", "NSWorkspaceDidHideApplicationNotification",
		"NSWorkspaceDidUnhideApplicationNotification", "NSWorkspaceActiveSpaceDidChangeNotification",
	} {
		center.Send(selAddObserver, workspaceObserver, selNotify, nsstr(name), objc.ID(0))
	}
	ObserveApps()
	if id := FocusedWindow(); id != 0 {
		lastFocus.Store(uint32(id))
	}
	return nil
}

// ObserveApps follows the windows of the apps running, as once the user
// allows Windeck to, which the observers need.
func ObserveApps() {
	for _, a := range RunningApps() {
		observeApp(a.PID)
	}
}
