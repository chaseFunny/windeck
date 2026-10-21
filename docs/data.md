# Privacy and local data

Everything Windeck keeps is in `~/Library/Application Support/Windeck`, in plain JSON you can read. Window pictures are kept in memory only.

| File | Holds |
|---|---|
| `settings.json` | your settings, including the privacy list (`excluded`, bundle IDs) |
| `learned.json` | what Windeck learned: see below |
| `license.json` | the trial's start, and the license key and activation id once activated |
| `window-state.json` | where the Settings window was |

## learned.json

```json
{
  "apps":   { "com.apple.Safari": { "s": 4.2, "t": 1792000000 } },
  "titles": { "com.apple.Safari\u001fGo Packages": { "s": 1.7, "t": 1792000000 } },
  "transitions": { "com.microsoft.VSCode": { "com.googlecode.iterm2": { "s": 9.1, "t": 1792000000 } } },
  "metrics": { "day": "2026-10-22", "switches": 84, "shown": 30, "topTwo": 22, "predicted": 6, "keys": 61, "total": 1210 }
}
```

- Each `{ "s", "t" }` is a score as of a Unix time, which halves every 72 hours; entries that decay to nothing are dropped.
- `titles` are kept to rank search. Apps on the privacy list are never recorded, nor captured.
- Settings › Privacy › Reset forgets all of it and every picture kept.

## What leaves your Mac

Nothing about your windows: no titles, pictures, file names or app lists. Activating a license sends the key and the Mac's name to Lemon Squeezy, which checks it.
