package wm

import (
	"os/exec"

	"github.com/ebitengine/purego/objc"
)

var (
	selMainBundle = objc.RegisterName("mainBundle")
	classBundle   = class("NSBundle")
)

// ResetPermission forgets what macOS remembers of Windeck for a service.
// macOS ties a grant to the build that got it: after an update the old
// entry can stay switched on in System Settings while the new build is not
// allowed, and the user cannot tell the two apart. Removing it before
// asking leaves one entry, the right one. It runs tccutil, so it belongs on
// a click, never on the shortcut's way; a bundle-less dev build skips it.
func ResetPermission(service string) {
	var id string
	withPool(func() { id = nsString(classBundle.Send(selMainBundle).Send(selBundleID)) })
	if id == "" {
		return
	}
	_ = exec.Command("/usr/bin/tccutil", "reset", service, id).Run()
}
