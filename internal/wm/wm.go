// Package wm is Windeck's window collection layer: it lists the windows
// of every app across Spaces and displays, minimized and hidden ones too,
// brings one to the front, closes or minimizes it, captures its picture,
// and reports what changes. Everything platform specific lives behind
// it, so a macOS release that moves the ground under it is fixed here.
//
// Only macOS is supported; elsewhere the functions report nothing, so
// that the rest of the app builds and its tests run.
package wm

import (
	"errors"
	"image"
)

// WindowID is the window server's number of a window (CGWindowID).
type WindowID uint32

// Rect is a rectangle in global display points, top-left origin.
type Rect struct{ X, Y, W, H float64 }

// Window is a window as the switcher lists it.
type Window struct {
	ID       WindowID
	PID      int32
	BundleID string
	AppName  string
	AppPath  string
	Title    string

	Minimized  bool
	Hidden     bool // its app is hidden (Cmd-H)
	FullScreen bool
	// Space is the number of the desktop it is on, counted per display as
	// Mission Control does, 0 when unknown; OnCurrentSpace tells that it
	// is on a Space the user sees now.
	Space          int
	OnCurrentSpace bool
	// Display is the index of the display it is on, in the order of
	// Displays; External that it is not the built-in one.
	Display  int
	External bool
	Bounds   Rect
}

// App is a running app the user can switch to.
type App struct {
	PID      int32
	BundleID string
	Name     string
	Path     string
	Hidden   bool
	// Regular tells an app of the Dock from an accessory one, as a menu
	// bar app.
	Regular bool
}

// Event is something that changed in the windows the user has.
type Event int

const (
	// EventFocus is a window coming to the front: Focused names it.
	EventFocus Event = iota
	// EventChanged is windows created, closed, minimized, retitled, apps
	// launched or quit: the list should be read again.
	EventChanged
	// EventSpace is the user going to another Space.
	EventSpace
)

// Change is an Event, with the window it concerns where it has one.
type Change struct {
	Event   Event
	Window  WindowID
	PID     int32
	Leaving WindowID // the window that had the focus, for EventFocus
}

// Key is a key the switcher handles while its shortcut is held.
type Key int

const (
	KeyOther Key = iota
	KeyTab
	KeyEscape
	KeySpace
	KeySlash
	KeyReturn
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyW
	KeyM
	KeyQ
	KeyGrave
)

// Modifier is the modifier of the switcher's shortcut.
type Modifier int

const (
	ModOption Modifier = iota
	ModCommand
)

// KeyEvent is a key press the event tap saw, or the shortcut's modifier
// being released (Released).
type KeyEvent struct {
	Key      Key
	Shift    bool
	Repeat   bool
	Released bool // the modifier went up
	// Mod tells that the shortcut's modifier is down, Other that another
	// modifier (Control, or the other of Option and Command) is.
	Mod, Other bool
}

// Capture is a window's picture.
type Capture = *image.RGBA

// Display is a connected display.
type Display struct {
	ID      uint32
	Bounds  Rect
	BuiltIn bool
	Main    bool
}

// ErrNotTrusted is the event tap failing for want of the accessibility
// permission.
var ErrNotTrusted = errors.New("windeck: the accessibility permission is needed")

// Permission services, as tccutil names them.
const (
	ServiceAccessibility = "Accessibility"
	ServiceScreen        = "ScreenCapture"
)
