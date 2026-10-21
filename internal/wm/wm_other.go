//go:build !darwin

package wm

func Trusted(prompt bool) bool                            { return false }
func CanCapture(prompt bool) bool                         { return false }
func RunningApps() []App                                  { return nil }
func FrontmostPID() int32                                 { return 0 }
func ActivateApp(pid int32)                               {}
func QuitApp(pid int32)                                   {}
func HideApp(pid int32)                                   {}
func Windows() []Window                                   { return nil }
func StackingOrder() []WindowID                           { return nil }
func Focus(w Window)                                      {}
func Close(id WindowID) bool                              { return false }
func Minimize(id WindowID) bool                           { return false }
func FocusedWindow() WindowID                             { return 0 }
func CaptureWindow(id WindowID, maxSide int) Capture      { return nil }
func SetSystemSwitcher(on bool)                           {}
func Displays() []Display                                 { return nil }
func Observe(fn func(Change)) error                       { return nil }
func ObserveApps()                                        {}
func StartTap(mod Modifier, fn func(KeyEvent) bool) error { return ErrNotTrusted }
func SetTapModifier(mod Modifier)                         {}
func SetCapturing(on bool)                                {}
func ModifierDown() bool                                  { return false }
