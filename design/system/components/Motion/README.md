# Motion

The system's motion (D196, additions.css §75 + the motion block at the top of additions.css). One easing `--ip-ease`
(cubic-bezier(0.2, 0, 0, 1)), three durations (`--ip-dur-key` 120 / `--ip-dur-panel` 200 / `--ip-dur-page` 280 ms), and the
distances `--ip-press-scale` 0.97, `--ip-pop-scale` 0.95, `--ip-modal-scale` 0.96, `--ip-off` 100%, `--ip-shift` 8 px,
`--ip-stagger` 40 ms. Transform and opacity only, transitions only (no keyframes).

- **Hook:** the script that opens or closes an element sets `data-state="open"` / `"closed"` on it. Open enters through
  `@starting-style` (also on insertion); closed plays the same path back and ends `display:none`. Elements without the
  attribute are drawn open and still, so static boards and state previews never move. Loaded weeks: `tbody.ip-ptable__week.is-new`
  with `--ip-i` 0–3; loaded content: `[data-state="loaded"]`.
- **Moves:** key press (Button, IconButton, BottomTabBar key) · Dialog (scrim fade + scale 0.96) · BottomSheet and phone
  SidePanel sheet (up from the foot, 280) · docked SidePanel (from the right edge) · Toast (up from the foot) · menus and
  Select lists (scale 0.95 from the trigger, 120) · Banner (8 px from above) · PeriodTable week open/close (rows fade, no
  height) and load 4 more weeks (8 px + fade, 40 ms apart) · skeleton → content (fade).
- **Colour only:** RecordChip, SegmentedControl, Checkbox, Tabs. **Never:** figures, Readouts, GuideChart, CycleTrack,
  lamps, table cells, skeleton bars (no shimmer), the focus ring, Sidebar and tab bar navigation, the SaveBar.
- **Reduced motion:** `prefers-reduced-motion: reduce` (or `.ip-motion--reduce` on a subtree) sets the distances to rest;
  every fade keeps its timing.

`preview.html` is a full page (not a dsCard fragment) with a small script: each card's keys play its motion; the switch at the
top simulates reduced motion. Built with the canvas board M1 · Mișcare by `design/system/scripts/mo01_build.py <repo_root> [--canvas]`.
To feel-check, slow it down 5× (DevTools › Animations, or CDP `Animation.setPlaybackRate` 0.2).
