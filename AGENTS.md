# AGENTS.md

Windeck is a macOS window switcher built with MyGo (pure Go, native UI). Read docs/architecture.md first.

## Rules

- No cgo: everything builds with `CGO_ENABLED=0`. Native code goes through purego, in `internal/wm` only.
- Nothing may slow the shortcut down: no I/O, IPC or captures on the main thread in `open`; list windows and take pictures on goroutines and publish through `App.update`.
- The cycle's first place is always the previous window. Predictions may take the second place only.
- No polling of windows or screens: work comes from events. The only timer loop watches permissions while the onboarding or Settings window is open.
- Window titles, pictures and app lists never leave the Mac.
- Strings live in `locales/en.json` and `locales/zh-Hans.json`; run `go tool mygo generate` after changing them.

## Commands

```sh
go build ./... && GOOS=linux go build ./...
go vet ./... && go tool mygo vet .
go test ./...
WINDECK_LIVE=1 go test ./internal/wm -run TestLive -v      # this Mac's windows, from a terminal with the permissions
WINDECK_RENDER=/tmp/windeck go test -run TestRenderLive .  # the views drawn to PNGs
go tool mygo build                                         # build/darwin-arm64/Windeck.app and .dmg
go run ./internal/tools/appicon                            # redraw the icon
bun install && bun run --cwd website dev                   # the website
```
