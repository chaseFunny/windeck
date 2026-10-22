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

## The look

Windeck follows the hz design system (github.com/hz7002/hz-design): white, near-black text and one violet, #5A4BC6, kept to about a twentieth of what shows; grays with a touch of violet; flat colors, layers told apart by background and a 1-point border, shadows only on what floats. `theme.go` holds its tokens and its steps (corners 6, 10, 16; spacing 4 to 48; type 20, 16, 14, 13, 12; weights 400, 500, 700), and turns MyGo's widgets violet with `brandTheme`; every view starts with `themed(c)`. Inter and Nunito are in the app (`resources/fonts`, OFL), Chinese falls back to the system's font. Icons are Lucide's, with a 1.5 line. A screen has one primary button, and one big bubble at most: the onboarding's title. The panel is cut to continuous corners by `wm.RoundWindow`, so the system's shadow follows them.

In the list, every title starts on one line and every icon on another: a group's windows have no icon, their titles where the group's name is. Titles drop what repeats their app's name (`shortTitle`); the preview's header keeps the whole title. Search makes the run of letters it matched bold.

To look at a window as it really shows, material and all, without the shortcut taking the keyboard: `WINDECK_DEV_SHOW=panel|search|settings|onboarding go run .`, which keeps its data in a temporary folder.
