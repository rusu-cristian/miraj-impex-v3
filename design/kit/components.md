# Impex Line Panel: components

Each entry lists its purpose, states and the tokens it uses. Live previews are in the Impex Line Panel design system; markup and styles are in `components/bundle.css` (`ip-` classes).

**Shared behaviour for every component:**
- `focus-visible` is a 2px `focus` ring with a 2px offset.
- Touch targets are at least 44px.
- Phone controls use `control-h-touch`.
- Icons are decorative unless the component has no text, in which case it needs an `aria-label`.


## RecordCode

The product-wide code: every record type has one icon and one colour, used on every screen, so a screen full of data can be scanned by type.

| Type | Token | Icon |
|---|---|---|
| Găini | `rec-hen` | ♀ (venus) |
| Cocoși | `rec-rooster` | ♂ (mars) |
| Mortalitate | `rec-mortality` | skull |
| Sacrificare | `rec-cull` | ban |
| Eroare sexare | `rec-hen` + `rec-rooster` | two rings, split tile |
| Populare | `rec-placement` | log-in |
| Transfer | `rec-transfer` | arrows left-right |
| Vânzare | `rec-sale` | truck |
| Hrană | `rec-feed` | wheat (daily), package (delivery), coins (price) |
| Cântărire | `rec-weighing` | scale |
| Ouă | `rec-eggs` | egg |

**Rules:**
- The colour always comes with its icon, and with a word or a ♀/♂ sign. It never carries meaning alone.
- Every `rec-*` colour is ≥4.5:1 on `ground`, `panel` and its own `-bg` tint.
- Record colours mark *what* a record is. Status (*how* it is) uses lamps and the `ok` / `warn` / `fault` / `info` tokens, never a record colour.
- ♀ figures are always `rec-hen`; ♂ figures are always `rec-rooster`. This applies in rows, inputs, chart lines and legends.
- **Known close pairs:** cocoși blue and cântărire cerulean; sacrificare plum and transfer violet. They're told apart by icon and word.


## RecordTile

The square icon tile that leads every event row, entry card and alert.

- **Size:** `record-tile` (28px), radius `radius-sm`.
- **Colours:** background `rec-*-bg`, icon `rec-*` (18px, 1.75 stroke).
- **Eroare sexare:** the split ♀/♂ tile.
- **Feed** uses the wheat icon (daily), the package icon (delivery) or the coins icon (price), all in `rec-feed`.
- The tile is decorative (`aria-hidden`). The row's text names the type.


## RecordChip

A labelled record type, used in legends and as filter chips in the FilterBar.

- **Style:** background `rec-*-bg`, text and icon `rec-*`, 1px border at 30% of the colour, radius `radius-sm`, height 28px.
- **Filter chips:**
  - on: the full chip
  - off: `raised` with `ink-muted` text, keeping the coloured icon
- **Behaviour:** `role="checkbox"` with `aria-checked`.


## SexCount

How every hen and rooster figure is written: ♀ in `rec-hen`, ♂ in `rec-rooster`, figures in IBM Plex Mono.

- **Order:** ♀ always comes first, separated from ♂ by `·`.
- **Missing sex:** a sex the scope doesn't hold is left out.
- **Signs:** removals use the true minus (−).
- **Sizes:** `figure` in rows, `readout` in inputs and stats, `readout-lg` for the key figure.


## Lamp

Status as an indicator lamp with a word: în normă / curent, abatere, refuzat / defect, informație, oprit.

- **Lamp:** an 8px dot in `ok-lamp`, `warn-lamp`, `fault` or `info`. The `off` lamp is a hollow `line-strong` ring.
- **Text:** in `ok`, `warn`, `fault` or `info`.
- **On readouts:** `readout-ok`, `readout-warn`, `readout-fault`.
- The word is required. Lamps never use record colours.


## Tag

A filled status label for tables and stats: `ok`, `warn`, `fault`, `info`, `neutral`.

- **Style:** the `-bg` tint with the status text colour, radius `radius-xs`, 12px weight 500.
- **Length:** ≤ 16 characters.


## Button

Panel keys. `primary` (graphite `control`, one per view), `secondary`, `ghost` and `destructive`.

- **Primary:** `control` with `on-control` text and `shadow-key`.
  - hover: `control-hover`
  - pressed: `control-pressed` (the key sinks; no shadow)
- **Secondary:** `raised` with 1px `line-strong` and `shadow-key`.
- **Destructive:** `raised`, with the border and text in `fault`.
- **Disabled:** `sunken` + `ink-disabled`.
- **Loading:** a spinner, the label stays.
- **Height:** `control-h` on desktop, `control-h-touch` on phone. Radius `radius-sm`.
- **Labels:** verbs: Salvează, Mută, Înregistrează vânzarea, Salvează prețurile.


## IconButton

An icon-only action with a required `aria-label`. `control-h` square (`control-h-touch` on phone), icon `ink-secondary`; hover `sunken`.


## TextField

Text, decimal, password and note inputs.

- **Field:** `raised`, 1px `line-strong`, radius `radius-sm`. The label sits above in 12px `ink-secondary`.
- **Units** are a suffix: g, kg, lei/kg, %.
- **States:**
  - focus: `ink` border + `focus` ring
  - error: `fault` border + an error line
  - read-only: `sunken`
  - disabled
- **Phone:** height `control-h-touch`, 17px text.


## CountInput

Whole-number input for birds and eggs. The sex is marked by a 3px inner left bar in `rec-hen` or `rec-rooster`, and by ♀ / ♂ in the label.

- **Empty:** "–", meaning null, not 0.
- **Large:** height `count-input-h`, style `readout` (Plex Mono 20).
- **States:** warning (`warn-lamp` border + a line on `warn-bg`), error (`fault`), saved (read-only).
- **Input:** `inputmode="numeric"`. Enter moves to the next field.


## Select

One choice from a list: feed type, guide, farm, section. The menu is `raised` with `shadow-overlay`; the active option is `sunken`. Group headings are captions. Options can lead with a RecordTile or a SexCount.


## SectionSelect

Choose a farm › section, with birds alive shown as ♀ · ♂. Grouped by farm with type captions. A disabled option shows its reason ("sursa"); an empty section says "goală". Opens as a BottomSheet on phone.


## Checkbox

Binary choice: "Fără pierderi azi", "Toate", farm access. An 18px box; checked = a `control` fill with a white check. 44px hit area.


## SegmentedControl

2–4 exclusive options: period, grain, ♀/♂.

- **Track:** `sunken`.
- **Selected segment:** `raised` with 1px `line-strong`.
- **Sex segments** carry their ♀ / ♂ colour.


## DateField

The entry date: Azi / Ieri chips, the field `dd.mm.yyyy`, and the guide week of the farm's cycle in mono ("SAPT 41 · ciclu din 10.02.2026"). Future dates are disabled.


## FileDrop

Paste or upload a breed guide (.xlsx / .csv): a dashed `line-strong` border on `raised`. Row errors show in a fault Banner.


## Stat

One indicator compared with the guide.

- **Card:** `panel`, 1px `line`, radius `radius-md`.
- **Contents:** a caption with a RecordTile icon, the value in Plex Mono 24, the guide line, and a status Tag or Lamp.
- **Colour:** the value stays `ink`. Use ♀/♂ colours only for sex figures.


## DataTable

Compact tables on `raised` with uppercase caption headers on `panel`.

- **Rows:** `row-h`, separated by `rule`. Numbers are right-aligned in Plex Mono.
- **Record-type columns:** the header carries the record icon in its colour, and ♀/♂ columns are coloured.
- **Groups:** `sunken` rows.
- **Totals:** a `line-strong` top border.
- **Inline edit cells** (owner prices) use a medium TextField.
- **Phone:** horizontal scroll with the first column sticky.


## EventRow

One record in a log: RecordTile, a title with section and figures (♀/♂ coloured), meta (date · SAPT · author), then a status lamp or tag.

- **Row:** height ≥ 48px, separated by `rule`; hover `sunken`.
- **Filtering:** by RecordChips in the FilterBar.
- **Locked rows** show a "blocat" tag.


## SectionTile

One section on a farm overview.

- **Contents:**
  - the code, and a status lamp
  - ♀ · ♂ alive
  - a ♀/♂ proportion bar
  - key/value lines with record icons: mortality this week, grams today, weight vs guide, % lay
- **Card:** `raised`, 1px `line`, radius `radius-md`.
- **Hover:** `line-strong` border.
- **Missing data:** a `warn-lamp` border and the lamp "lipsă".
- **Interaction:** clicking opens the section.
- **Layout:** grouped by hall when there are compartments.


## CycleSelector

Choose which cycle a farm page shows.

- **Menu:** the open cycle first (lamp `ok`), then past cycles with their series, dates and length.
- **Actions** (manager+): Începe ciclu nou and Schimbă ghidul ciclului.


## Readout

The Moment: a dark display set into the light UI for the page's key figure and its guide chart.

- **Style:** `readout` background, radius `radius-md`.
- **Caption:** Plex Mono uppercase in `readout-muted`, with a status lamp on the right.
- **Value:** `readout-lg` in `readout-ink`.
- **Sub-line:** guide and deviation.
- **Use:** at most one per section of a page. Stock readouts on the feed report may repeat in a row.


## GuideChart

Actual vs guide, drawn on a Readout.

- **Grid:** `readout-line`.
- **Guide:** dashed `readout-guide`, with its tolerance band in `readout-band`.
- **Actual:** `readout-ok`.
- **Extra sections:** `readout-2` … `readout-4`, each with its own dash.
- **Today:** a faint `readout-ok` line, with the last point marked.
- **X axis:** guide weeks of the cycle.
- **Accessibility:** keyboard readout, a hidden data table, no animation.
- **Variants:** weight, lay, feed g, mortality %, egg weight.


## CycleTrack

The signature: the series' cycles over calendar time, one lane per farm.

- **Lanes:** `sunken`, with uppercase farm labels.
- **Cycle blocks:**
  - closed: `panel` + `line-strong`
  - open: `raised` + a 1.5px `ink` edge, label weight 500
- **Markers:**
  - a placement start mark in `rec-placement`
  - the transfer link and dot in `rec-transfer`
  - a sale end mark in `rec-sale`
- **Today:** the "azi" line in `ok-lamp` across both lanes, with a small Readout per open cycle ("SAPT 41").
- **Interaction:** clicking a block opens that farm at that cycle.
- **Variants:**
  - `full`: series page
  - `compact`: Acasă for the owner, the farm header
- **Accessibility:** the SVG carries a text summary. Blocks are focusable links.


## EntryCard

One section in the Zilnic form.

- **Header:** the code, ♀ · ♂ alive, a status tag.
- **Count inputs:** ♀/♂-barred CountInputs for Mortalitate and Sacrificare (Sacrificare requires a reason).
- **"Mai mult":** Eroare sexare.
- **Ouă group:** on the hall card at a laying farm.
- **"Fără pierderi azi"** checkbox.
- **Card:** `panel`, 1px `line`, radius `radius-md`.
- **Layout:** one column on phone; a 2-column grid on desktop.


## AllocationRow

One section × sex row in Hrană zilnică.

- **Left:** the ♀/♂ sign and the section code, with birds alive.
- **Middle:** the feed type select and the g/pasăre input. Yesterday's grams appear as a *suggestion* (dashed `sunken` field, "ieri") until confirmed with a tap or by typing.
- **Right:** the live kg, plus the guide figure and a deviation lamp.
- **Unconfirmed rows** aren't saved.
- **Card:** `raised`, 1px `line`, radius `radius-md`.


## EffectPreview

Shows the result of a transfer, sale or correction before saving, including any cycle that opens or closes.

- **Style:** `sunken`, radius `radius-sm`.
- **Contents:** the record icons and lines with ♀/♂ figures.
- **Refused** becomes a fault Banner.


## Banner

An inline message (info, warn, fault, ok): the status tint with a 1px status border and icon. The text stays `ink`.


## Toast

A save confirmation in `control` (graphite), with an optional undo. Not used for errors.


## EmptyState

An unlit lamp with one sentence and a next step, e.g. "Ferma e goală din 30.09.2026." or "Toate livrările au preț.".


## Skeleton

Static `sunken` bars, no shimmer.


## Dialog

Confirms destructive and cycle-changing actions.

- **Style:** `panel`, radius `radius-lg`, `shadow-overlay`, on `scrim`.
- **Title:** names the action and the object.
- **Fields:** a reason where required.
- **Focus:** starts on Anulează.


## SidePanel

Details and corrections beside the page on desktop: `panel`, `panel-w`, `shadow-overlay`, a sticky footer. A full-screen sheet on phone.


## BottomSheet

Phone choice lists, such as the Înregistrează menu. Each entry is a RecordTile, then a label and a hint.


## Sidebar

Desktop navigation on `panel`.

- **Wordmark:** "IMPEX LINE" in uppercase, tracked.
- **Series selector**, then the nav items.
- **Current item:** `raised` with a 1px `line` and a green `ok-lamp` on the right, like an active channel.
- **Înregistrează** expands into entry items, each with its record icon in its colour.


## BottomTabBar

Phone navigation on `panel`. The centre Înregistrează key is a 48px graphite circle with `shadow-key`. Vets see 3 items; the other roles see 5.


## FormHeader

Phone form header: back, the title with its RecordTile, and a context line in mono (farm · date · SAPT).


## SaveBar

The sticky save area of a phone form: a summary with record icons and ♀/♂ figures, then the primary key. Disabled until there's something to save.


## PageHeader

Breadcrumbs as uppercase captions, then the title (24px), type and guide tags, a meta line (♀ · ♂ · SAPT · cycle start) and actions on the right.


## Tabs

View tabs. The selected tab is `ink` with a 2px `ink` underline and a small `ok-lamp` dot. Others are `ink-secondary`.


## FilterBar

Period, grain, record-type chips and export. On phone it collapses into one button that opens a BottomSheet.


## Patterns
- **Forms:** one column, label above. "(opțional)" marks optional fields. Validation runs on blur and on save; the first error gets focus; typed values are kept; the draft is kept until it's saved.
- **Warnings vs refusals:** warnings let you save. Refusals disable save and name the section, the number and the day.
- **Destructive and cycle-changing actions:** a Dialog that names the object; a reason is required for manager edits.
- **Loading:** static skeletons; spinners only inside buttons.
- **Empty:** an EmptyState with an unlit lamp and a next step.
- **No permission:** "Nu ai acces la această fermă."

## Additions (project, 2026-11-16)

The Impex Line Panel system artifact can't be edited from this session (D16), so per D32 every component, state or rule this project adds or clarifies is recorded here, in the system's own language and from existing tokens. Its CSS lives in `design/system/components/additions.css` (loaded after `bundle.css`; on the canvas at `project/additions.css`), and each entry is listed in HANDOFF to be copied into the system artifact later. Entries below override nothing in the sections above unless they say so; they settle what those sections left open (D31).

- **Stat value size:** the value uses the `readout` type style (Plex Mono 20/24, weight 500), not a raw 24px. additions.css sets this on `.ip-stat__val` (overriding bundle.css's 24/28); builders change nothing.
- **Uppercase exceptions:** besides `caption`, uppercase is allowed only for the mono Readout caption (`ip-readout__cap`), the "IMPEX LINE" wordmark and CycleTrack lane labels (and breadcrumbs, as bundle.css already draws them).
- **Green without a word:** allowed only for non-status cues: the Sidebar current-item dot, the selected Tab dot, the "azi" line on the CycleTrack and GuideChart, and the focus ring. Any status in green is a Lamp or Tag with its word.
- **AllocationRow deviation:** shown as a Lamp plus a word (e.g. `ip-lamp--warn` "peste ghid"), never a Tag.
- **"fără date în ghid":** a guide line, written `ghid: – fără date în ghid` in `ip-stat__guide` / the row's guide slot (D29), never a Tag or Lamp.
- **Interaction states:** real `:hover`, `:active` (pressed: the key sinks, no shadow) and `:disabled` for Button, IconButton, RecordChip, Tabs, TextField/Select (`.ip-input`, incl. native `<input>`/`<select>` inside it and `:focus-within` ring), menu items, SectionTile and EventRow now live in additions.css, mirroring the `is-hover`/`is-pressed` demo classes with the same tokens. Pressed on steel surfaces (Button secondary and ghost, IconButton, EventRow) deepens one steel step, `sunken` hover → `rule` pressed; secondary also drops its `shadow-key` (D34a). While pressed on `rule`, `ink-muted` text (`ip-cap`, `ip-note`, `ip-hint`, `ip-stat__guide`, `ip-pair .sep`) and lamp words inside these controls switch to `ink-secondary` (ink-muted is 4.4:1 and the ok word 4.3:1 on `rule`); lamp dots keep their status colour. Demo classes `is-pressed`/`is-disabled` were added where bundle.css had none. Motion vars `--ip-ease` (`cubic-bezier(0.2,0,0,1)`), `--ip-dur-key|panel|page` (120/200/280ms) and a `prefers-reduced-motion` guard are defined there. Preview: `design/system/components/_Interaction/preview.html`.
- **Interaction states, fix round 1 (SYS00):** every `:hover`/`:active` rule also skips `[aria-disabled="true"]` (and `:disabled`, `.is-disabled`), so a disabled link or control never changes on hover or press. Destructive pressed = `color-mix(in srgb, var(--fault) 8%, var(--fault-bg))` (fix round 2, D34a), fault text (4.6:1) and edge kept, one step past its `fault-bg` hover. Ghost pressed stays `rule` (D34). RecordChip pressed = `panel` with its record-colour edge (rec-* text stays ≥4.5:1). IconButton disabled = `sunken` + `ink-disabled`, like Button. A `.ip-list` holding `a`/`button` EventRows is `overflow: visible` (bundle.css clips it), with `radius-md` outer corners on its first and last rows, so the focus ring (2px, 2px offset) shows.
- **Disabled exception, Tabs and EventRow:** disabled is text-only, `ink-disabled` with no `sunken` fill (a filled strip would read as a selected or pressed row). Every other control keeps the craft rule: `sunken` + `ink-disabled`.
