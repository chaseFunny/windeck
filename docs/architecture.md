# Architecture

Windeck is one Go binary built with MyGo, pure Go with no cgo: native APIs are called through purego.

| Package | Is |
|---|---|
| `internal/wm` | the window collection layer: windows of every app on every Space (accessibility API, the window server's list, and elements found by remote token for other Spaces), focus and raise (SkyLight), captures (`SLSHWCaptureWindowList`), the event tap of the shortcut, and observers of what changes. Everything macOS specific is here. |
| `internal/deck` | the model: windows grouped by app in places that do not move, the most-recently-used history, and the Tab cycle (the previous window always first). |
| `internal/switcher` | the state machine: pending (quick switch), browsing (state A), searching (state B), as Actions the app carries out. |
| `internal/smart` | frecency, app and window transitions, the second-place prediction, and search matching and ranking. |
| `internal/snapshot` | windows' pictures in memory, least recently used out past a budget of bytes. |
| `internal/settings`, `internal/license` | the settings; the trial and Lemon Squeezy activation. |
| the root package | the app: the panel, Settings, onboarding and the menu bar item, in MyGo's native UI. |

## Nothing on the critical path

The windows are listed as they change (NSWorkspace notifications and accessibility observers), off the main thread, so pressing the shortcut reads what is already known. The prediction is a lookup. A quick press and release switches before the panel would show. Pictures are taken as the user leaves a window, and when the panel shows, on other goroutines.
