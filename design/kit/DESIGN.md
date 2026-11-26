---
name: Impex Line Panel
colors:
  steel-0: "#ffffff"
  steel-50: "#f4f6f7"
  steel-100: "#e8ebed"
  steel-150: "#dce1e4"
  steel-200: "#d3d9dd"
  steel-300: "#c3cad0"
  steel-400: "#93a2ac"
  steel-500: "#76828b"
  steel-600: "#56626b"
  steel-700: "#434e57"
  steel-800: "#1f272d"
  steel-900: "#13181c"
  ground: "{colors.steel-100}"
  panel: "{colors.steel-50}"
  raised: "{colors.steel-0}"
  sunken: "{colors.steel-150}"
  rule: "{colors.steel-200}"
  line: "{colors.steel-300}"
  line-strong: "{colors.steel-500}"
  ink: "{colors.steel-900}"
  ink-secondary: "{colors.steel-700}"
  ink-muted: "{colors.steel-600}"
  ink-disabled: "{colors.steel-400}"
  control: "{colors.steel-800}"
  control-hover: "#2c363e"
  control-pressed: "#0c1013"
  on-control: "#ffffff"
  focus: "#17703f"
  readout: "{colors.steel-900}"
  readout-line: "#263038"
  readout-ink: "#e6eef1"
  readout-muted: "#93a2ac"
  readout-guide: "#7e8d97"
  readout-band: "#1e2a26"
  readout-ok: "#53dd8c"
  readout-warn: "#f2b544"
  readout-fault: "#ff7d72"
  readout-2: "#62c3f5"
  readout-3: "#c69cf2"
  readout-4: "#d9c58a"
  ok: "#17703f"
  ok-lamp: "#1a8a4c"
  ok-bg: "#daf0e2"
  warn: "#7f4e00"
  warn-lamp: "#a86c05"
  warn-bg: "#fbeaca"
  fault: "#b0261e"
  fault-bg: "#f8dedb"
  info: "#285a82"
  info-bg: "#dce8f2"
  scrim: "#13181c80"
  rec-hen: "#b0246a"
  rec-hen-bg: "#f9e2ee"
  rec-rooster: "#1f5bb5"
  rec-rooster-bg: "#e1eaf8"
  rec-mortality: "#3a424a"
  rec-mortality-bg: "#e3e6e9"
  rec-cull: "#7c2f86"
  rec-cull-bg: "#f1e2f3"
  rec-placement: "#b4450f"
  rec-placement-bg: "#fbe6da"
  rec-transfer: "#5b3cc4"
  rec-transfer-bg: "#e9e4fa"
  rec-sale: "#56680c"
  rec-sale-bg: "#ecf0d8"
  rec-feed: "#7f5c00"
  rec-feed-bg: "#f6edc8"
  rec-weighing: "#0a6a8f"
  rec-weighing-bg: "#ddeef5"
  rec-eggs: "#0d6f69"
  rec-eggs-bg: "#dcf0ee"
typography:
  title:
    fontFamily: '"IBM Plex Sans", system-ui, sans-serif'
    fontSize: 24px
    lineHeight: 30px
    fontWeight: 500
  title-sm:
    fontFamily: '"IBM Plex Sans", system-ui, sans-serif'
    fontSize: 20px
    lineHeight: 26px
    fontWeight: 500
  heading:
    fontFamily: '"IBM Plex Sans", system-ui, sans-serif'
    fontSize: 16px
    lineHeight: 22px
    fontWeight: 500
  body:
    fontFamily: '"IBM Plex Sans", system-ui, sans-serif'
    fontSize: 14px
    lineHeight: 20px
    fontWeight: 400
  body-strong:
    fontFamily: '"IBM Plex Sans", system-ui, sans-serif'
    fontSize: 14px
    lineHeight: 20px
    fontWeight: 500
  input:
    fontFamily: '"IBM Plex Sans", system-ui, sans-serif'
    fontSize: 17px
    lineHeight: 24px
    fontWeight: 400
  small:
    fontFamily: '"IBM Plex Sans", system-ui, sans-serif'
    fontSize: 12px
    lineHeight: 16px
    fontWeight: 400
  caption:
    fontFamily: '"IBM Plex Sans", system-ui, sans-serif'
    fontSize: 11px
    lineHeight: 14px
    fontWeight: 500
    letterSpacing: 0.06em
  readout-lg:
    fontFamily: '"IBM Plex Mono", ui-monospace, monospace'
    fontSize: 32px
    lineHeight: 36px
    fontWeight: 500
  readout:
    fontFamily: '"IBM Plex Mono", ui-monospace, monospace'
    fontSize: 20px
    lineHeight: 24px
    fontWeight: 500
  figure:
    fontFamily: '"IBM Plex Mono", ui-monospace, monospace'
    fontSize: 14px
    lineHeight: 20px
    fontWeight: 400
  figure-sm:
    fontFamily: '"IBM Plex Mono", ui-monospace, monospace'
    fontSize: 12px
    lineHeight: 16px
    fontWeight: 400
rounded:
  none: 0px
  xs: 2px
  sm: 3px
  md: 4px
  lg: 6px
  round: 50%
spacing:
  1: 4px
  2: 8px
  3: 12px
  4: 16px
  5: 20px
  6: 24px
  8: 32px
  12: 48px
  control-h: 32px
  control-h-touch: 48px
  count-input-h: 56px
  row-h: 32px
  record-tile: 28px
  lamp: 8px
  header-h: 56px
  tabbar-h: 64px
  sidebar-w: 232px
  panel-w: 440px
components:
  button-primary:
    backgroundColor: "{colors.control}"
    textColor: "{colors.on-control}"
    typography: "{typography.body-strong}"
    rounded: "{rounded.sm}"
    padding: "0 16px"
    height: "{spacing.control-h}"
  button-secondary:
    backgroundColor: "{colors.raised}"
    textColor: "{colors.ink}"
    typography: "{typography.body-strong}"
    rounded: "{rounded.sm}"
    padding: "0 16px"
    height: "{spacing.control-h}"
  button-destructive:
    backgroundColor: "{colors.raised}"
    textColor: "{colors.fault}"
    typography: "{typography.body-strong}"
    rounded: "{rounded.sm}"
    height: "{spacing.control-h}"
  button-touch:
    typography: "{typography.input}"
    height: "{spacing.control-h-touch}"
    padding: "0 20px"
  input:
    backgroundColor: "{colors.raised}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.sm}"
    padding: "0 12px"
    height: "{spacing.control-h}"
  count-input:
    backgroundColor: "{colors.raised}"
    textColor: "{colors.ink}"
    typography: "{typography.readout}"
    rounded: "{rounded.sm}"
    padding: "0 12px"
    height: "{spacing.count-input-h}"
  record-tile:
    rounded: "{rounded.sm}"
    size: "{spacing.record-tile}"
  record-tile-hen:
    backgroundColor: "{colors.rec-hen-bg}"
    textColor: "{colors.rec-hen}"
  record-tile-rooster:
    backgroundColor: "{colors.rec-rooster-bg}"
    textColor: "{colors.rec-rooster}"
  record-tile-mortality:
    backgroundColor: "{colors.rec-mortality-bg}"
    textColor: "{colors.rec-mortality}"
  record-tile-cull:
    backgroundColor: "{colors.rec-cull-bg}"
    textColor: "{colors.rec-cull}"
  record-tile-placement:
    backgroundColor: "{colors.rec-placement-bg}"
    textColor: "{colors.rec-placement}"
  record-tile-transfer:
    backgroundColor: "{colors.rec-transfer-bg}"
    textColor: "{colors.rec-transfer}"
  record-tile-sale:
    backgroundColor: "{colors.rec-sale-bg}"
    textColor: "{colors.rec-sale}"
  record-tile-feed:
    backgroundColor: "{colors.rec-feed-bg}"
    textColor: "{colors.rec-feed}"
  record-tile-weighing:
    backgroundColor: "{colors.rec-weighing-bg}"
    textColor: "{colors.rec-weighing}"
  record-tile-eggs:
    backgroundColor: "{colors.rec-eggs-bg}"
    textColor: "{colors.rec-eggs}"
  tag-ok:
    backgroundColor: "{colors.ok-bg}"
    textColor: "{colors.ok}"
    typography: "{typography.small}"
    rounded: "{rounded.xs}"
    padding: "2px 6px"
  tag-warn:
    backgroundColor: "{colors.warn-bg}"
    textColor: "{colors.warn}"
    typography: "{typography.small}"
    rounded: "{rounded.xs}"
    padding: "2px 6px"
  tag-fault:
    backgroundColor: "{colors.fault-bg}"
    textColor: "{colors.fault}"
    typography: "{typography.small}"
    rounded: "{rounded.xs}"
    padding: "2px 6px"
  tag-info:
    backgroundColor: "{colors.info-bg}"
    textColor: "{colors.info}"
    typography: "{typography.small}"
    rounded: "{rounded.xs}"
    padding: "2px 6px"
  panel:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.md}"
    padding: "16px"
  section-tile:
    backgroundColor: "{colors.raised}"
    textColor: "{colors.ink}"
    rounded: "{rounded.md}"
    padding: "12px"
  readout:
    backgroundColor: "{colors.readout}"
    textColor: "{colors.readout-ink}"
    typography: "{typography.readout-lg}"
    rounded: "{rounded.md}"
    padding: "12px 16px"
  table-header:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink-muted}"
    typography: "{typography.caption}"
    padding: "8px 12px"
  table-row:
    backgroundColor: "{colors.raised}"
    textColor: "{colors.ink}"
    typography: "{typography.small}"
    padding: "6px 12px"
    height: "{spacing.row-h}"
  sidebar:
    backgroundColor: "{colors.panel}"
    width: "{spacing.sidebar-w}"
    padding: "12px"
  tab-bar:
    backgroundColor: "{colors.panel}"
    height: "{spacing.tabbar-h}"
  form-header:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    typography: "{typography.heading}"
    height: "{spacing.header-h}"
  save-bar:
    backgroundColor: "{colors.panel}"
    padding: "12px 16px"
  toast:
    backgroundColor: "{colors.control}"
    textColor: "{colors.on-control}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: "12px 16px"
  dialog:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.lg}"
    padding: "24px"
    width: "{spacing.panel-w}"
  cycle-track-open:
    backgroundColor: "{colors.raised}"
    textColor: "{colors.ink}"
    rounded: "{rounded.xs}"
  cycle-track-closed:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.xs}"
---

# Impex Line Panel: design

## Overview

**Impex Line Panel** is the barn-control-panel design of Impex Line. The interface is set like the hall's climate controller:
- light steel panels
- engraved uppercase captions
- graphite keys
- status lamps
- dark readout windows for the figures that matter

**Character:** alert, exact, mechanical.

**Colour has three separate jobs:**
1. **Record type:** a fixed icon + colour (`rec-*`), the same on every screen.
2. **Status:** a lamp + a word (`ok`, `warn`, `fault`, `info`).
3. **Steel neutrals** for everything else.

**Signature: the cycle track.** One lane per farm.
- Each cycle is a block.
- The transfer is a `rec-transfer` link between lanes; the placement and sale marks are at the ends.
- The green "azi" line runs through all lanes, with a guide-week readout per open cycle.

**Moment: dark Readouts** for the key figure and the guide charts.

**Template test:** with a generic farm SaaS's name swapped in, the direction doesn't fit. That product wouldn't have farm lanes with cycles, LED status words or a fixed record-type code.

**Theme and density:** light only. Mixed density: 32px desktop rows and controls; 48px and 56px phone controls.

## Colors

- **Steel ramp:** `steel-0` … `steel-900`.
- **Surface roles:** `ground` (page), `panel`, `raised`, `sunken`, `rule`, `line`, `line-strong` (3.3:1).
- **Text:** `ink` 14.9:1, `ink-secondary` 7.1:1, `ink-muted` 5.2:1, `ink-disabled`.
- **Keys:** `control` (graphite) is the primary key, with `on-control` text.
- **Status:**
  - `ok` + `ok-lamp` (normal / current, also the focus ring)
  - `warn` + `warn-lamp`
  - `fault`
  - `info`
  - each with a `-bg` tint
- **Readout:** `readout`, `readout-line`, `readout-ink`, `readout-muted`, `readout-guide`, `readout-band`, `readout-ok`, `readout-warn`, `readout-fault`, plus `readout-2` … `readout-4` for extra series.
- **Record-type code** (each ≥4.5:1 on ground, panel and its tint):
  - hen #B0246A, rooster #1F5BB5
  - mortality #3A424A, cull #7C2F86
  - placement #B4450F, transfer #5B3CC4, sale #56680C
  - feed #7F5C00, weighing #0A6A8F, eggs #0D6F69
  - each with a `-bg` tint

## Typography

- **Families:** IBM Plex Sans for words; IBM Plex Mono for every number, week and readout.
- **Styles:** `title` 24/30, `title-sm` 20/26, `heading` 16/22, `body` 14/20, `body-strong`, `input` 17/24, `small` 12/16, `caption` 11/14 uppercase +0.06em.
- **Readout figures:** `readout-lg` 32/36, `readout` 20/24, `figure` 14/20, `figure-sm` 12/16.
- **Weights:** 400 and 500.
- **Uppercase** is allowed only in `caption`.

## Layout

- **Spacing:** a 4px base.
- **Breakpoints:** `sm` < 640, `md` 640–1023, `lg` ≥ 1024.
- **Desktop:**
  - a sidebar (`sidebar-w` 232) on `panel`
  - page edge 24px
  - scope pages: PageHeader → CycleSelector and compact CycleTrack → FilterBar with RecordChips → Tabs → content
  - the farm Rezumat: a Readout with a GuideChart, Stats, SectionTiles grouped by hall, alerts, EventRows
- **Phone:**
  - page edge 16px
  - a BottomTabBar with a graphite centre key
  - forms are FormHeader (with a RecordTile) → cards or rows → SaveBar

## Elevation & Depth

- **Surfaces are flat,** separated by 1px edges.
- **Shadows:**
  - `shadow-key`: the lower edge of buttons only
  - `shadow-overlay`: menus, dialogs, toasts, side panel only
- **Readouts** sit flush in panels as dark windows.
- **Focus:** a 2px `focus` ring (green) with a 2px offset.

## Shapes

- **Radii:** `xs` 2 (tags, cycle blocks), `sm` 3 (controls, record tiles), `md` 4 (panels, readouts), `lg` 6 (dialogs, sheets).
- **Round shapes:** status lamps and the phone centre key only.
- **Nesting:** the inner radius is the outer radius minus the padding.

## Components

The full specs are in `components.md`. Live previews are in the Impex Line Panel design system.

**Inventory:**
- **Record types:** RecordCode, RecordTile, RecordChip, SexCount.
- **Status:** Lamp, Tag.
- **Actions:** Button, IconButton.
- **Inputs:** TextField, CountInput, Select, SectionSelect, Checkbox, SegmentedControl, DateField, FileDrop.
- **Data:** Stat, DataTable, EventRow, SectionTile, CycleSelector.
- **Moment:** Readout, GuideChart.
- **Signature:** CycleTrack.
- **Entry:** EntryCard, AllocationRow, EffectPreview.
- **Feedback:** Banner, Toast, EmptyState, Skeleton.
- **Containers:** Dialog, SidePanel, BottomSheet.
- **App shell:** Sidebar, BottomTabBar, FormHeader, SaveBar, PageHeader, Tabs, FilterBar.

**Motion (D196, additions.css §75):**
- One easing: `cubic-bezier(0.2, 0, 0, 1)`. Transform and opacity only, as transitions (no keyframes).
- 120ms for key presses (scale 0.97), menus, toggles and table rows; 200ms for dialogs, the side panel, toasts and banners; 280ms for phone sheets and page-level changes.
- Things enter from where they come (sheets and toasts from the foot, the panel from the right, menus from their field) and leave the same way.
- Lamps don't blink. Charts, numbers, table cells and navigation don't animate. Reduced motion keeps the fades, drops the movement.
- Tabs and SegmentedControl (D198, §77): one selected mark (tab underline / segment thumb) slides to the new item, 200ms, label colour in step; instant from the keyboard and under reduced motion. Loaded content replaces its skeleton with a reveal: blocks fade in and rise 8 px, 40ms apart (≤4 steps).
- Live updates (D197, §76) are the one place data moves: a new row rolls in 8 px from above + opacity (200ms) and its record tint fades once over 1.6s (linear, `--ip-dur-tint`), "Nou" until seen; a changed value crossfades 4 px (never counts) and tints its cell and totals; a new transfer mark grows 0.96 → 1 at the azi line; a reader scrolled away gets "N înregistrări noi · Arată" instead.

**Icons:** line icons, 1.75 stroke, at 14/16/18/22px. Record icons are fixed per type and reserved.

## Do's and Don'ts

**Do:**
- Lead every record with its RecordTile.
- Colour every ♀ figure `rec-hen` and every ♂ figure `rec-rooster`.
- Pair every lamp with a word.
- Put numbers in Plex Mono.
- Write the guide week as `SAPT 41` from the cycle start.
- Show unpriced feed as "fără preț".

**Don't:**
- Use a record colour to show status, or a status colour to show a record type.
- Use green for primary buttons.
- Use more than one Readout per page section (except feed stock).
- Use dark panels outside Readouts.
- Use gradients, glows, neon, glassmorphism or blur.
- Build KPI-card grids with big icons.
- Use emoji or illustrations.
- Use exclamation marks or greeting headlines.
- Make lamps blink.
- Bring in anything from the ledger look (serif titles, barn red, paper).
