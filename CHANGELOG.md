# Changelog

## 1.2.1

- The intro plays once on every Mac, on the first launch of a version that has it, also where Windeck was set up before; Windeck records that it played (intro.json, beside the settings) and never plays it again. 1.2.0 played it only on a first setup.

## 1.2.0

- An intro on the first launch, over the whole screen with a breathing violet aurora: "Windeck" is written out by hand, each letter turns into a window card, the cards stack into the logo's deck as its violet square opens, the name rises under it, and the logo flies to the permissions' window, or into the switcher shown under a guide to its keys. A click, Return or Escape skips it.
- The pointer works in the switcher opened by the shortcut: hovering a row shows its buttons, as it did only in search before.
- Tab goes down the list once the panel shows, one row a press, instead of jumping along the history; before it shows, a quick press still goes back to the previous window.
- Minimizing no longer shakes the list: the row stays where it is, with the highlight on it.
- A minimized window's row has a restore button and its close button; M restores it too.
- The shortcut closes the panel kept open by search.
- The preview's header lines up with the search field, and the picture is larger.
- Translucent, as popMind and the system's own panels: the switcher floats on the popover material, Settings on the sidebar material as System Settings does, the onboarding and update windows on the window's; the blurred desktop shows through, and cards keep their plain colors.
- Fixed: holding W a moment too long closed every window of the app, one after the other, as the key repeated and the highlight went on to the next; and Q held quit app after app. W, M and Q now act once a press.
- The panel's hints say what W and Q act on: "Close window" and "Quit app".

## 1.1.1

- Permissions survive updates: Windeck is now signed with its own certificate, the same for every version, so macOS knows an update as the app it allowed. Updating from 1.1.0 or earlier asks for Accessibility once more; after that, never again.
- Allowed already, and still asked? Allow… now clears the entry an earlier version left in System Settings before asking, so the list holds one Windeck and its switch is this version's.

## 1.1.0

- Updates: Windeck checks its releases once a day and from the menu bar's "Check for Updates…", shows what is new and installs the signed update, then relaunches; or, chosen in Settings → About, installs it without asking for the next launch. Updates are signed, and only those signed with Windeck's key are installed.
- A Shortcuts page in Settings: the shortcut and how long it is held, every key of the switcher, and a switch to turn W, M and Q off.
- The switcher no longer flashes: a window whose picture is on the way shows its outline until the picture fills it, instead of its app's icon; the pictures of the windows around the highlight are taken ahead; the highlight moves with the keys at once; the panel shows and hides without the system's fade and never shows the last switch's frame.
- A second layout, List: the windows as one column of icons and titles, the most recent first, the panel as tall as its rows. Choose it in Settings → General → Layout, or from the menu bar's Layout menu.
- The hz design system throughout, the app and the site: the violet #5A4BC6, Inter and Nunito, Lucide icons, flat cards, one primary button a screen; a new flat icon; the onboarding's title in a bubble; the hz mark in About and the site's footer.
- The site follows the hz landing page: one big bubble, the product on a gray block, three cards, text and pictures taking turns, facts that are measured, a dark ending.
- The panel floats on the system's material with continuous corners and the system's shadow; the highlighted window is a tab, as a browser's.
- Titles in the list drop the app's name they repeat ("Page - Google Chrome - Work" reads "Page") and line up on one edge; search shows the letters it matched in bold, and says so when nothing matches.
- Settings look like System Settings: colored icons in a sidebar on its material, ⌘1 to ⌘6 for the pages, the shortcut as keys to choose between, the privacy list with the apps' names and icons, apps not installed folded away, and a menu to add one.
- The onboarding's permissions in one group, with the step shown in the footer; Return goes on.
- Fixed: Quit in the menu bar, and SIGTERM, did not quit.
- Fixed: turning previews off left the pictures kept showing, and kept.

## 1.0.0

The first version.

- ⌥Tab (or ⌘Tab) shows the windows of every app as tabs grouped by app, with a large preview; releasing switches, a quick press and release goes back without the panel.
- Minimized, hidden, full-screen windows and those of other desktops and displays, with badges.
- W, M and Q close, minimize and quit from the panel; Space or / searches.
- The second place of Tab is predicted from your habits, and search is ranked by them (Pro).
- Pictures of windows kept in memory for those that can no longer be captured, within a memory budget.
- A privacy list, a reset of everything learned, and first-launch steps for the permissions and the shortcut.
- English and Simplified Chinese.
