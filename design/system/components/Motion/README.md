# Motion

The system's motion (D196, additions.css §75; live updates D197, §76 + the motion block at the top of additions.css). One easing `--ip-ease`
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
  height) and load 4 more weeks (8 px + fade, 40 ms apart) · skeleton → content (MO04: `div.ip-reveal`, the block `[data-state=loaded]` 8 px + fade 200 ms, `--ip-i` 0–3 40 ms apart, the skeleton `[data-state=closed]` fades out under it) ·
  Tabs underline and SegmentedControl thumb (MO04, §77: one `::before` per group slides, translateX, `--ip-dur-slide` 200 ms; labels' colour in step; arrow keys and reduced motion instant) · PeriodTable week key fills its cell (§71).
- **Colour only:** RecordChip, Checkbox. **Never** (outside live updates, below): figures, Readouts, GuideChart, CycleTrack,
  lamps, table cells, skeleton bars (no shimmer), the focus ring, Sidebar and tab bar navigation, the SaveBar.
- **Live updates (MO03, D197, §76):** the one place data moves. Hooks set by the live script: `data-live="new"` (+ `data-state="open"` on insert) on a row → 8 px from above + opacity, 200 ms, rows below take its room at once; `data-live="changed"` on a cell → `span.ip-live__swap` (old figure 4 px up and out, new 4 px up and in, 200 ms, never counting); `data-live-rec=<type>` + `data-live-tint` on → fade → removed: the record tint (`--ip-live-mix` 10 % of `rec-*` over the surface's own fill) fading once over `--ip-dur-tint` 1600 ms, **linear** (the one exception to `--ip-ease`: a timer of how recent, not a movement; `--ip-ease` spent ~80 % of it in the first 0.6 s); "Nou" (`span.ip-tag--info.ip-live__mark`) until seen; CycleTrack `g.ip-track__live` opacity + scale 0.96 → 1 from its anchor (a state change: SVG has no `@starting-style`); `div.ip-livedock > button.ip-livebar[data-state]` "N înregistrări noi · Arată" (as a Banner, floating, 0-high sticky dock); `span.ip-live__at` "Actualizat hh:mm"; one polite announcement per batch. The preview's last card drives them on real figures (18.11.2026 08:15 Zilnic, 08:27 + 08:44 as a batch while scrolled, the 10:12 correction 17.11 H5.2 ♀ 1 → 3, a replay (D39) of the 28.02.2026 transfer H2 → H7.1, H7.2); a second press reverts (demo only), so a retarget can be watched.
- **For the product (MO03 audit):** remove `data-live-tint` on the element's `transitionend` with `propertyName` `background-color` (never on a timer, so a fade is never cut) and the old figure (`.ip-live__old`) on its own `transitionend` (`opacity`); the swap's old figure leaves at key speed (120) and the new one starts 60 ms later, so two figures are never legible at once; insert a CycleTrack mark with `data-state="closed"`, flush style (`getComputedStyle(el).opacity`), then set `"open"` (SVG gets no `@starting-style`); the "noi" bar closes itself, the rows going in, when the reader scrolls the list back to its top. The preview and M1 do exactly this (`data-lk` names each tinted element's key).
- **Reduced motion:** `prefers-reduced-motion: reduce` (or `.ip-motion--reduce` on a subtree) sets the distances to rest
  and `--ip-dur-slide` to 0 (tabs and segments change at once); every fade keeps its timing.

`preview.html` is a full page (not a dsCard fragment) with a small script: each card's keys play its motion; the switch at the
top simulates reduced motion. Built with the canvas board M1 · Mișcare by `design/system/scripts/mo01_build.py <repo_root> [--canvas]`.
To feel-check, slow it down 5× (DevTools › Animations, or CDP `Animation.setPlaybackRate` 0.2).
