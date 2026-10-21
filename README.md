# Windeck 窗签

A window switcher for macOS that works like browser tabs.

Hold <kbd>⌥</kbd> and press <kbd>Tab</kbd>. Your windows show as tabs down the left, grouped by app, with the highlighted one large in the middle. Let go of <kbd>⌥</kbd> to switch.

- **Tap to go back.** A quick press and release returns to the previous window, and nothing shows on screen.
- **Readable tabs, a large preview.** Full titles grouped by app, and a big picture to confirm the window.
- **Every window.** Minimized, hidden, on other desktops and on other displays, each with a badge.
- **Search.** Press <kbd>Space</kbd> or <kbd>/</kbd>, let go of <kbd>⌥</kbd>, and type.
- **Smart, on your Mac only.** The first Tab always goes to the previous window. The second place is predicted from your habits, and search is ranked by them.

See the [docs](docs/index.md).

## Development

Windeck is a [MyGo](https://github.com/egoist/mygo) app: pure Go (no cgo), with native UI. It needs Go 1.27+.

```sh
go tool mygo dev          # run Windeck Dev with live reload
go tool mygo build        # build/darwin-arm64/Windeck.app and the .dmg
go test ./...             # unit tests
bun install && bun run dev   # the website, in website/
```

Windeck needs the Accessibility permission, and Screen Recording for previews. A `go run .` from a terminal that already has them inherits them.

## License

Proprietary. © 2026 Windeck.
