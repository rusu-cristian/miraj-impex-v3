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

### SYS01 · components for S4 (farm page), 16.11.2026
Folders in `design/system/components/` (README + preview each), CSS in additions.css sections 2–7. Decided by default (user away), to review.

- **CycleTrack `farm` variant** (`CycleTrackFarm/`): `svg.ip-track.ip-track--farm`, one `sunken` lane for one farm, every cycle of any series over calendar time, no lane label. Closed blocks `panel` + `ink-muted` edge (blocks are links: ≥3:1 on the lane; D38, also for the full CycleTrack's link blocks); open `raised` + 1.5px `ink`. Block y 16 h 18. Start mark: placement `rec-placement` or transfer in `rec-transfer` (`.xfer`); end mark: sale `rec-sale` or transfer out `rec-transfer`; a legend names the marks. "azi" line `ok-lamp` with a small "SAPT 40" Readout when a cycle is open. Blocks are `<a>` links with `aria-label`/`<title>`; the SVG has a `<desc>` summary. Selecting a block selects that cycle; the selected block has `aria-current="true"` and a 2px `ink` underline in the lane at y 37, ≥2px clear of the block (open or closed). Hover: closed `raised` + `ink-secondary` edge, open `panel`. Focus: a 2px `focus` ring rect 2px off the block.
- **SectionTotals** (`SectionTotals/`): `.ip-stotals` strip under the SectionTile grid; `panel`, 1px `line`, top `line-strong`, `radius-md`, cells split by `rule`. Cells in SectionTile order with its record icons: Total fermă · ♀ · ♂ alive · Mort. săpt. (total + ♀ · ♂) · Hrană azi as SexCount ♀ · ♂ g (♀ weighted by hens, ♂ by roosters, D38; sub-line "ghid 158 / 155 g") · Greutate ♀ vs guide (hen-weighted) · Ouat vs guide (laying). Values Plex Mono 14 500 `ink`; sub-line names guide, exclusions per sex ("medie fără ♀ H4.2", "fără ♂ H3.2") or the reason for `–`. Grouped numbers use U+202F everywhere.
- **AlertRow "De verificat"** (`AlertRow/`): `.ip-ev.ip-ev--alert`: RecordTile of the record type concerned · title "section · indicator figure · guide/limit" ("H3 · ouat 70,1 % · ghid 74,1 %") · meta date · SAPT · since when · optional "Deschide secțiunea" link · status Lamp + word at the right edge in a ≥96px column (`warn` abatere / lipsă, `fault` refuzat, `info` informație). Never a record colour for status. Empty: off lamp "nimic de verificat" + one sentence.
- **Guide missing** (`GuideMissing/`): Stat with no guide figure: `ghid: – fără date în ghid`, no Tag/Lamp. Farm without a guide: neutral Tag "fără ghid" in the guide slot. GuideChart partial (demo, D38a: "Fermă exemplu · creștere", weight ♀, real growth guide to SAPT 23): guide + band stop at the last guide week with an end cap (`.gend`) and the note "fără date în ghid după SAPT 23"; GuideChart no guide: actual only, no band. No-guide demos use "Fermă exemplu". Every focusable `svg.ip-chart` has the keyboard week readout: ←/→ week cursor (`.cur`), Home/End, `.ip-chart__read[role=status]` naming week, actual, guide, deviation. Both use the new readout off lamp `ip-rlamp--off` (hollow `readout-muted` ring) with "FĂRĂ DATE ÎN GHID" / "FĂRĂ GHID".
- **Sidebar series selector** (`SeriesSelector/`): APG select-only combobox `div.ip-input.ip-series__btn[role=combobox][tabindex=0]` (`aria-expanded`, `aria-controls`, `aria-activedescendant`) under the wordmark: name, Lamp + word ("deschisă" ok / "închisă" off), chevron; `listbox` (`ip-menu`) over the nav, options with a check slot, name + Lamp, dates in Mono ("până la 28.01.2026"). Keyboard: ↓/Enter/Space/Alt+↓ open, ↑↓ Home End PgUp PgDn and type-ahead move, Enter/Space/Alt+↑ select, Esc close, Tab selects and moves on; focus stays on the combobox. `a.ip-nav` has no underline.
- **FilterBar Interval** (`FilterInterval/`): with Interval selected, `.ip-interval` holds "de la" and "până la" DateFields (labels above, Mono `dd.mm.yyyy` at `10ch` wide, 24px calendar key), default cycle start → today; at 1180px the bar stays one row with Exportă (`.ip-filterbar__end`). Future dates: typed → error row in flow under the bar (`.ip-filterbar__err`) "Până la 20.11.2026 e în viitor. Alege o dată până la 16.11.2026."; calendar = APG date-picker dialog with `table[role=grid]`, `aria-selected` on the gridcell, `aria-current="date"` on today, days after today `aria-disabled` (`sunken` + `ink-disabled`), next month disabled; selected day `control`. Keys: arrows, Home/End, PgUp/PgDn, Enter picks, Esc closes; focus returns to the calendar key.
- **Helpers:** `.ip-sr` (visually hidden), vars `--ip-hair`, `--ip-edge-strong`, `--ip-mark-w`, `--ip-fs-*/--ip-lh-*` (type sizes from tokens.css) so additions carry no raw values.

### SYS02 · FeedTotals for S7 (Hrană zilnică), 16.11.2026
Folder `design/system/components/FeedTotals/` (README + preview), CSS in additions.css section 8. Decided by default (user away), to review.

- **Purpose:** the per-feed-type totals under the AllocationRows (spec C2 step 4): what today's confirmed rows take from each feed type's stock and how many days are left. Plus a compact variant for the SaveBar summary.
- **Anatomy:** `section.ip-ftot` = SectionTotals' strip (`panel`, 1px `line`, top `line-strong`, `radius-md`, rows split by `rule`); head `ip-cap` "Total pe tip de hrană" + "dd.mm.yyyy · doar rânduri confirmate"; one `.ip-ftot__row[role=group]` per feed type used today at the farm: RecordTile wheat (`rec-feed`) · type name + for-sex sign (`rec-hen`/`rec-rooster`) · "n din m rânduri confirmate" · Azi (kg, one decimal, Plex Mono 16/20 500 like the AllocationRow kg) · Stoc după azi (whole kg, "din <stoc acum> kg") · Zile rămase (stoc după azi ÷ average of the last 3 days, one decimal, "până pe dd.mm.yyyy") · Lamp + word. Foot `ip-note` names the formula, the averaging window and the threshold. Phone by default (type line + lamp, then three right-aligned figure cells); `ip-ftot--wide` for desktop (one line per type, `aria-hidden` column header, 120px status column).
- **States:** în normă (`ok`, ≥ 5 zile) · abatere (`warn`, < 5 zile, D19) · stock below 0 (`warn` lamp "abatere", D50, with "stoc sub 0" as the Stoc după azi sub-line, + the warn Banner with the spec sentence "Stocul de Breeder 2 la Hîrbovăț ar ajunge la −1 240 kg. Verifică livrările.", days "0,0 · se termină azi"; a warning, Salvează stays enabled) · nothing confirmed (Azi "–" "nimic confirmat", stock = "stoc acum", days from it) · rows left out named under Azi ("fără ♂ H3.2").
- **Compact (SaveBar):** `.ip-savebar__sum.ip-ftot-sum`, two lines: wheat + "10 din 12 rânduri · 3 691,3 kg"; "2 de confirmat ·" + one stock Lamp, worst type first ("Cocoși 3,9 zile" warn / stock below 0: warn "abatere" + detail "Breeder 2 · stoc sub 0" (D50) / "stoc în normă" ok). Counts rows (section × sex), not sections. Nothing confirmed: "0 din 12 rânduri · – kg", Salvează disabled.
- `ip-ftot--wide` is for desktop only.
- **Tokens:** existing only; new var `--ip-fs-alloc/--ip-lh-alloc` (16/20) names bundle.css's `.ip-alloc__kg` size.
- **A11y:** static (no targets); type rows are labelled groups ("Breeder 2, pentru găini"); per-cell keys stay for screen readers in the wide layout; the stock warning is `role=status`; kg figures are not live regions (no announcement per keystroke); every lamp has its word.

### SYS03 · FarmSummary for S3 (series) and S2 (Acasă, owner), 16.11.2026
Folder `design/system/components/FarmSummary/` (README + preview), CSS in additions.css section 9. Decided by default (user away), to review.

- **Purpose:** one panel per farm for a series (S3 §3, spec D1 "a summary per farm"), and the owner's farm summaries on Acasă (S2, compact).
- **Anatomy:** `article.ip-fsum` (`panel`, 1px `line`, `radius-md`, 16px padding). Head: farm name `h3` 16/22 500 as the panel's one link (ink, underlined; to the farm page at that cycle; inline-flex, min-height `--space-6` = 24px target in narrow/`--wide`/`--compact`, 48px in `--touch`, D46) · type Tag and guide Tag (`neutral`) · cycle Lamp + word at the right (`ok` deschis / `off` închis / `off` goală; these cycle-state words are allowed outside the D42 status vocabulary, state not status, D46). Mono context line: "ciclu din 10.02.2026 · SAPT 40" or "07.10.2025 – 28.02.2026 · 21 săpt.". Alive: caption + SexCount at `readout` size + "total n" (closed cycle: "Populat · 07.10.2025" with the placement, since alive at close is 0). Key figures: `dl` of rows split by `rule`, SectionTile's kv language: record icon 14 + key · value Plex Mono 14/20 500 `ink` right-aligned (♀ · ♂ as SexCount) · sub-line 12px `ink-muted` (guide, deviation, week, exclusions or `ghid: – fără date în ghid`) · Lamp + word only where a guide or D19 threshold exists, words from the D42 Lamp vocabulary only ("în normă" / "abatere" …); the detail ("+6,8%", "sub 5 zile") goes in the sub-line. Every ♀/♂ sign, keys and sub-lines included, is an `ip-sex` span. Percentages without a space ("74,1%", "−7,3%"; value unit `small.ip-fsum__pct`, D46).
- **Figures:** laying, open: Ouat azi vs ghid · Mortalitate cumulată ♀ · ♂ % · Hrană azi ♀ · ♂ g vs ghid (D38 weighting, names left-out rows) · Zile stoc ♀ · ♂ (package; `warn` "abatere", sub-line "Cocoși până pe 19.11.2026 · sub 5 zile", the short type named). Growth, closed: Greutate ♀ la transfer vs ghid (±5%; outside → `warn` "abatere", "+6,8%" in the sub-line) · Mortalitate cumulată · Hrană totală g pe pasăre = Σ daily g/pasăre over the days with feed rows vs the guide over the same days (D46; Telița ♀ 8 401 · ♂ 10 142 vs 9 065 · 10 835 g, 143 days → `ok` "în normă"), total kg ♀ · ♂ on the sub-line · Transferate ♀ · ♂ (arrivals, D28).
- **Variants:** narrow by default (phone; Lamp right of the sub-line) · `ip-fsum--wide` (S3 desktop, `.ip-fsums` 2 columns; Lamps in a fixed 96px column like AlertRow) · `ip-fsum--compact` (S2 owner: 12px padding, alive inline at figure size, 2 figures: Ouat azi and Mort. cumulată, because feed stock has its own S2 block) · `ip-fsum--touch` (phone: links reach 48px).
- **Empty farm between cycles:** Lamp `off` "goală" + `.ip-fsum__empty`: "Ferma e goală din 28.02.2026. Ultimul ciclu: Seria 2026, 21 săpt." and the next step link "Vezi ultimul ciclu". S2 shows a farm as it is today; S3 shows the farm's cycle in that series.
- **Rules:** no Readout inside (one per section; S3's moment is the CycleTrack), no big icons, values stay `ink` except ♀/♂; unknown = "–" with its reason. Guide Tag shows the guide's full name even past 16 characters ("Ross 308 creștere", 17).
- **Tokens:** existing only; new var `--ip-fs-heading/--ip-lh-heading` (16/22) names tokens.css `.ip-t-heading`.

### SYS04 · StockReadout, 16.11.2026
Folder `design/system/components/StockReadout/` (README + preview), CSS in additions.css section 10. Decided by default (user away), to review.

- **Purpose:** S11 › Stoc (spec D4 §1): one dark Readout per farm × feed type, repeated in a row (the craft.md exception to one Readout per section).
- **Anatomy:** `ul.ip-readouts` (StockReadoutRow) > `li` > `div.ip-readout.ip-readout--stock[role=group]`: caption "HÎRBOVĂȚ · BREEDER 2" (Mono uppercase `readout-muted`, one line; the for-sex follows as hidden `ip-sr` text ", pentru găini", no `aria-label` on the span) with `ip-rlamp` + word on the right · value = kg in stock (`readout-lg`, whole kg) · `dl.ip-readout__kv` above a `readout-line` hairline: zile rămase (one decimal) · medie/zi 13–15.11 (last 3 complete days) · se termină (run-out date) · optional `ip-readout__sub` with the reason.
- **States:** `ok` în normă (≥ 5 zile) · `warn` abatere (< 5 zile, D19; "sub pragul de 5 zile") · stock below 0: `warn` abatere (D40 warning, D50 wording; sub-line "stoc sub 0 · verifică livrările", "epuizat pe") · `off` oprit with "– fără consum în ultimele 3 zile" · `off` oprit for a farm without an open cycle (`ip-readout--off`: one readout per farm, value "–" in `readout-muted`, "fără ciclu deschis"; Telița never shows its −716 kg workbook Sold).
- **Row:** `auto-fill` columns ≥ 240px: 4 per row at 1136px (2–3 readouts keep that width), one column on phone. An `ip-note` under it names the formula, window and threshold.
- **No record icon** on the readout: `rec-feed` is 2,92:1 on `readout` (< 3:1); the caption word names the type.

### SYS04 · CostBlock, 16.11.2026
Folder `design/system/components/CostBlock/` (README + preview), CSS in additions.css section 11. Decided by default (user away), to review.

- **Purpose:** S11 › Cost (spec D4 §4): lei per cycle from priced deliveries, lei per bird, lei per egg, and the "X kg fără preț" not included. A compact DataTable, not a KPI grid.
- **Anatomy:** `section.ip-cost` > `.ip-cost__head` (`panel`, caption + "din livrările cu preț · până pe dd.mm.yyyy") > `.ip-cost__scroll` > `table.ip-table.ip-cost__table` (hidden `caption.ip-sr` = the head title; Ciclu · Kg · Lei · Lei/pasăre · Lei/ou), one `tbody` per cycle: cycle row (`th[scope=row]` name + sub-line, Mono right-aligned figures) and `tr.ip-cost__unpriced` (hidden `ip-sr` cycle prefix "Hîrbovăț · Seria 2026, ", coins `rec-feed` + info Tag "fără preț", unpriced kg in the Kg column, "n livrări, dates · neincluse în cost"; none unpriced: "Toate livrările au preț.") · totals row `is-total` (`line-strong` edge) "Total Seria 2026" naming the excluded kg · `ip-cost__foot` note naming the bases.
- **Bases:** Lei/pasăre per bird at the start of the cycle (placed or arrived by transfer), not alive; Lei/ou = lei ÷ total eggs (D18) to date, laying only (growth "–" "creștere, fără ouă"). Whole lei for totals, 2 decimals for unit costs.
- **Layouts:** default = phone (DataTable rule: horizontal scroll, sticky 160px first column with a `rule` edge, focusable labelled region; no `overflow:hidden` on the block so the focus ring shows whole, head/foot round their own `radius-md` corners); `ip-cost--wide` for desktop.
- **Roles:** owner, manager, viewer: same read-only markup. Vets: the section is not rendered (D06).
- **Tokens:** existing only; new var `--ip-fs-table/--ip-lh-table` (13/18, bundle.css `.ip-table`). The coins icon is Lucide "coins" (the system names it but ships no path).

### SYS05 · EntryCard additions, 16.11.2026
Folder `design/system/components/EntryCardPlus/` (README + preview), CSS in additions.css section 12. Decided by default (user away), to review.

- **Purpose:** what S6 Zilnic (spec C1) needs on top of EntryCard: the yesterday hint, the warning and refusal states, the cull reason, "Fără pierderi azi" filled and the saved state.
- **Anatomy:** `section.ip-panel.ip-entry` > head (code `h4` 16/22, status Tag right, "în viață" ♀ · ♂ at the start of the day) > `.ip-entry__grid` (one column per sex present, ♀ left, ♂ right; rows Mortalitate, Sacrificare, Motiv; one-sex sections `--one`) > ghost key "Mai mult · Eroare sexare" > foot (Checkbox, or saved meta + Editează). CountInputs hold real `input[inputmode=numeric]` with placeholder "–" (null, not 0).
- **Yesterday hint:** under each sex's Mortalitate, "ieri ♀ 3 · media 7 zile 1,0" (`ink-muted` 12/16, figures Plex Mono, sex sign in its record colour); mean of the 7 days before the chosen day, one decimal; in `aria-describedby`.
- **Warning (saveable):** Mortalitate > 3 × the section's 7-day mean for that sex (D19, literal: a 0,0 mean warns on any death). `is-warn` + full-row warn line naming section, number and mean ("H1.1: 4 găini moarte, de peste 3 ori media pe 7 zile (1,0). Verifică numărătoarea; poți salva."), Tag `warn` "abatere", Salvează enabled.
- **Refusal:** Mortalitate + Sacrificare + Eroare sexare of one sex > alive at the start of the day. `is-error` + error naming section, number, alive, day and fix ("H1.2: 4 370 depășește cei 437 cocoși în viață pe 16.11.2026. Introdu cel mult 437."), Tag `fault` "refuzat", Salvează disabled; SaveBar Lamp + what it concerns.
- **Cull reason:** Sacrificare > 0 shows a required Select "Motiv ♀|♂" (boală, șchiopătare, altul; D36) under it in that sex's column; placeholder "Alege". Missing on save: error "H1.1: alege motivul pentru 2 găini sacrificate.", Tag "refuzat", save blocked, focus to the Select.
- **Fără pierderi azi:** fills 0 in Mortalitate and Sacrificare for the sexes present, read-only (`sunken`, sex bar kept), Tag `neutral` "fără pierderi"; unchecking restores the draft.
- **Saved:** read-only fields, Tag `neutral` "salvat", "Salvat la 08:34 · Dr. Ceban · poți modifica până la 23:59" + secondary Editează (D07).
- **Messages span the grid row** under the field they belong to (a 157px phone column can't hold them). Tags: one at a time, refuzat > abatere > fără pierderi / salvat; none while empty or unsaved.

### SYS05 · HallCard, 16.11.2026
Folder `design/system/components/HallCard/` (README + preview), CSS in additions.css section 13. Decided by default (user away), to review.

- **Purpose:** laying farm only: groups a hall's compartment EntryCards and carries the hall's Ouă entry (spec C1, EGG_PRODUCTION on the parent section).
- **Anatomy:** `section.ip-hall` > head on the ground with a `line-strong` top edge ("Hală" caption · code `h3` · "secțiuni H1.1 · H1.2" · ♀ · ♂ alive) > section EntryCards > `section.ip-panel.ip-eggs` "Ouă · H1" led by the eggs RecordTile: CountInputs Jos, Cat. 1, Cat. 2, Cat. 3, Incubabile (whole, numeric keypad) and TextField Masa medie (decimal, g, at CountInput height in this group) > read-only `dl.ip-eggs__calc`.
- **Computed:** Total ouă = Cat. 1 + 2 + 3 + Incubabile, "Jos separat (345)" (D18), only when all four are filled; Ouat azi = total ÷ hens alive at the start of the day vs the guide week ("75,8%", "ghid 74,1% · +1,7 puncte") with Lamp `ok` "în normă" / `warn` "abatere" beyond ±3 points (D19). Empty: "–" with its reason, no Lamp. Not a live region.
- **Layouts:** phone one column; `ip-hall--wide` (desktop): section cards in 2 columns, Ouă across with 6 fields in a row and the two computed figures side by side; 32px controls.
- **Tokens:** existing only; no new colours or values.

### S4P · RecordChip split + CycleTrack inferred start, 16.11.2026
CSS in additions.css sections 14–15 (moved out of the S4 screen CSS, D47/D55). Decided by default (user away), to review.

- **RecordChip · split** (`.ip-chip--split`, README note in `RecordChip/`): the Eroare sexare filter chip. Sexing error has no single record colour (D04), so the chip is neutral steel led by the split RecordTile: `button.ip-chip.ip-rec.ip-chip--split[role=checkbox]` > `span.ip-tile.ip-tile--split[aria-hidden]` (20px, `radius-xs`, the two rings ♀ `rec-hen` / ♂ `rec-rooster` at 16px) + "Eroare sexare". On: `sunken` + `ink` (edge `ink` at 30%); off (`.ip-chip--off`): `raised` + `ink-muted`, split tile kept. Hover, pressed, disabled and focus as every chip (section 1). Order in the FilterBar = RecordCode order (after Sacrificare).
- **CycleTrack · inferred start** (`.ip-track .ded`, legend key `.ip-tracklegend svg.ded`; README note in `CycleTrackFarm/`): a cycle whose start has no recorded placement or transfer (D54, e.g. Hîrbovăț Seria 2025 "început dedus 07.03.2025") gets no record mark; a dashed `ink-secondary` line (`ip-edge-strong`, dash `ip-mark-w`) across the lane (y 14–40), 3px before the block, inside the block's `<a>`. The legend names it with a 3×12 dashed key + "început dedus"; the block's `aria-label`/`<title>` and the SVG `<desc>` say "început dedus dd.mm.yyyy". Works in both CycleTrack variants. ≥3:1 on the `sunken` lane.

### SYS06 · GuideChart variants, 16.11.2026
Folder `design/system/components/GuideChartVariants/` (README + preview), CSS in additions.css section 16. GuideChart and GuideMissing previews aligned to D44. Decided by default (user away), to review.

- **Markup for every focusable chart (D42/D44):** `svg.ip-chart[role=group][aria-roledescription=grafic][tabindex=0]` + `aria-label` + `aria-describedby` (hidden help + readout); drawing in `<g aria-hidden="true">`; `div.ip-chart__read[role=status][aria-live=polite]` (←/→, Home/End, starts on today); hidden `table.ip-sr` with `thead th[scope=col]`, `th[scope=row]` per SAPT. The GuideChart preview now shows Telița ♀ weight (real) vs the growth guide; GuideMissing previews use the same markup and D46 percent format.
- **Two-sex lines:** ♀ `.act` (`readout-ok`, solid), ♂ `.s2` (`readout-2`, dash 8 3), dots `.pt`/`.pt2`. Measured: `rec-hen` 2,81:1 and `rec-rooster` 2,74:1 on `readout` (< 3:1), so RecordCode sex colours are not used on dark charts; the legend (`.ip-chart__legend`: 24×8 swatch + ♀/♂ sign in `readout-ink` + word) carries the sex. Extra series `.s3` (`readout-3`, dash 2 3), `.s4` (`readout-4`, dash 12 3 2 3).
- **Variants:** Mortalitate % cumulată ♀ · ♂ (% of placement; no guide, D29: `ip-rlamp--off` "FĂRĂ DATE ÎN GHID") · Hrană kg/zi by feed type (two lines, not stacked; no guide line, since the guide is g/pasăre; caption lamp = kg vs guide g × birds on the same rows, ±10%) · Uniformitate + CV ♀ (one % axis, no guide) · Incubabile % (incubabile ÷ total ouă, D18; guide if present, else no-guide) · Masa oului g (no guide).
- **Small curve** `svg.ip-spark` 160×48: actual + guide only, no axes/band, last point; `role=img` + `aria-label`, not focusable. Light steel by default (actual `ink`, ♀/♂ `rec-hen`/`rec-rooster`, guide dashed `ink-muted`) for DataTable cells (`.ip-sparktable`) and S20 guide detail; readout tokens inside `.ip-readout`.
- **Tokens:** existing only; new vars `--ip-line-w` (1.75), `--ip-dash-guide|2|3|4` name bundle.css values and the readout-3/4 dashes.
- **Fix round 1 · phone charts:** every GuideChart (base, GuideMissing, variants) = `div.ip-chartbox` (`container-type:inline-size`) with `svg.ip-chart--wide` (viewBox 600×150) and `svg.ip-chart--phone` (viewBox 320×150, plot x 40–308, axis text 11px caption size, fewer ticks: X first · middle · last, ≤3–4 Y); `@container (max-width: 599.98px)` shows the phone one, the other is `display:none` (out of tab order and the a11y tree). 600px = the wide viewBox width, the point where its 10px labels would shrink. A lone `svg.ip-chart` keeps working (wide only).
- **Fix round 1 · no sideways scroll:** `table.ip-sr{display:block}` (a hidden table no longer widens the page); wide DataTables in previews go in `div.ip-tablescroll[tabindex=0][role=region][aria-label]` (DataTable phone rule as a wrapper: `overflow-x:auto`, first column sticky with a `rule` edge) — `.ip-sparktable` and the RecordCodePlus demo table.
- **Fix round 1 · legends:** base GuideChart and GuideMissing use `.ip-chart__legend` swatches (act, guide, band) like the variants; `ip-readout__sub` keeps only context, no "— efectiv · - - ghid" text legend. GuideMissing partial chart has a `.today` line at SAPT 25.
- **Readout no-guide caption (D59):** on a dark Readout caption only, the off lamp (`ip-rlamp--off`, hollow ring in `readout-muted`) may carry "FĂRĂ GHID" (cycle has no guide) or "FĂRĂ DATE ÎN GHID" (guide has no figure for this indicator/week) — a state, not a status (D37/D46). The one exception to "fără date în ghid is never a Tag or Lamp" (D29); everywhere else it stays the guide line `ghid: – fără date în ghid`.

### SYS06 · RecordCode adjustment + cycle marks, 16.11.2026
Folder `design/system/components/RecordCodePlus/` (README + preview), CSS in additions.css section 17. Decided by default (user away), to review.

- **Corecție (adjustment)** `ip-rec--adjust`: Lucide "pencil-line" in `ink-secondary` on `sunken` (6,46:1 tint, 7,11 ground, 7,86 panel, 5,98 rule). Steel, not a status colour; close pair with Mortalitate (icon + word). RecordCode table row: Corecție · `ink-secondary` on `sunken` · pencil-line; FilterBar chip after Eroare sexare. An edited record keeps its type + Tag "corectat" (D41); only a count change with no other cause is a Corecție.
- **Cycle marks** `span.ip-cmark.ip-cmark--start|--end` in `.ip-ev__meta` + word "ciclu nou" / "ciclu închis": the CycleTrack 3×12 mark bar beside a mini block (open: `raised` + 1,5px `ink`; closed: `panel` + `ink-muted`), bar colour from the record that caused it (`.place`, `.xfer`, `.sale`, `.ded` = `svg.ded` dashed `ink-secondary` line, as the CycleTrack inferred start (fix 1: drawn as SVG, two dashes, so it never reads as the solid bar); none = `ink`). No new colour, never a Tag/Lamp. Cycle-only rows (Începe ciclu nou, golire) lead with the hollow `ip-tile--cycle` (`raised` + 1px `line`).

### SYS07 · SexCountRow, 16.11.2026
Folder `design/system/components/SexCountRow/` (README + preview), CSS in additions.css section 18. Decided by default (user away), to review.

- **Purpose:** the section rows with ♀ and ♂ counts in S10 Vânzare (C4), S18 Serie nouă placement (B3) and S9 Transfer (C3).
- **Anatomy:** `section.ip-panel.ip-sexrows` > head (caption + farm · date) > `ol.ip-sexrows__list` > `li.ip-sexrow`: code (or a SectionSelect combobox when the user adds the row) · "în viață" SexCount where birds leave (Vânzare, Transfer) · remove IconButton (`aria-label` "Elimină H1", Lucide x) on user-added rows only · ♀ and ♂ CountInputs (sex bar; label = visible ♀/♂ + hidden section and "găini"/"cocoși"; `inputmode=numeric`; "–" empty = null) · full-row error · Checkbox "Toate" (fills both counts with everything alive, inputs native `readonly` while checked with the Toate label's id added to their `aria-describedby`; not in placement). Ghost key "Adaugă secțiune" under user-added lists. Totals line (`line-strong` edge): caption + sub-line (rows filled, alive) + ♀ · ♂ sum as a SexCount at readout size; on desktop the ♀ sum sits in grid area `f` and the ♂ sum in area `m`, right-aligned under their columns (phone keeps one pair).
- **States:** a sex left "–" isn't required (a row needs one count); refusal above alive "H3.1: 500 depășește cei 437 cocoși în viață pe 16.11.2026. Introdu cel mult 437." (save disabled); duplicate section "H2 e deja în listă. Alege altă secțiune sau elimină rândul." (save disabled); SectionSelect disables sections already listed with the reason "în listă", empty ones say "goală".
- **Layouts:** phone stacked (section line; ♀ | ♂ at 56; Toate); `ip-sexrows--wide` one line per row with an `aria-hidden` column head, via a container query at ≥ 640px; `ip-sexrows--add` for user-added rows (Secțiune · ♀ · ♂ · remove); `ip-sexrows--pick` for S9 Transfer: one "Din" row, SectionSelect (farm › section, alive in the menu, 192 = `space-12` × 4) · În viață · ♀ · ♂ · Toate (areas `sec alive f m all`), no remove key, no Adaugă, no totals line.
- **Tokens:** existing only; new vars `--ip-sexrow-in|cols|areas` from space tokens.

### S7P · AllocationRow phone layout + SaveBar summary, 16.11.2026
CSS in additions.css sections 19–20 (moved out of the S7 screen CSS `s7-*`, D57). README note in `AllocationRow/`. Decided by default (user away), to review.

- **AllocationRow · phone** (`.ip-alloc.ip-alloc--phone`): the stacked row for a 358px card (390 page, 16px edges). Line 1: `.ip-alloc__who` (♀/♂ sign 16px in a 16px column, section code `body-strong`, "n găini|cocoși" `ip-note`) left; `.ip-alloc__kg` right as a column (kg Plex Mono 16/20 500 · "ghid 158 g [· +11%]" 12/16 · Lamp + word, or "de confirmat" / "de completat"; gaps 4). Line 1 shares one baseline: section code = kg figure. Line 2: `.ip-alloc__ctl` = Tip hrană Select (1fr) + g/pasăre TextField (120px), both `ip-input--touch` (48) with `ip-label` above (hidden ", H1.1 găini" suffix), so labels, fields and figures line up down the list. Optional `.ip-alloc__msg` (ip-warnline / ip-err) spans the card under the fields; placed by `grid-column`, so a row without one has no empty track (card = 12 + 60 + 12 + 68 + 12 + 2 = 166px). Suggested kg: `.ip-alloc__kg.is-suggested` (`ink-muted`). Row gap 12; tokens only.
- **SaveBar summary (amends SYS02 › Compact):** two lines at 390. Line 1: wheat · "10 din 12 rânduri ·" · "3 691,3 kg"; line 2: one status, Lamp + word + detail ("abatere · Cocoși 3,9 zile", "abatere · Breeder 2 · stoc sub 0", "refuzat · H1.1 ♀ 350 g · scrie 1–300 g") or, when saved, "salvat · 12 rânduri · 4 398,9 kg" / "Dr. Ceban · poți modifica până la 23:59". "n de confirmat" is no longer repeated (it is m − n from line 1, and each such row says "de confirmat"). Each segment is one no-wrap flex item (§20), so under 390 the bar wraps between segments, never between a figure and its unit (360: up to 4 lines).
- **Focus not obscured (2.4.11):** a screen with this SaveBar sets `scroll-padding-bottom` to the bar height + 5 (field border 1, ring offset 2, ring 2): 78px at ≥ 390 (bar 73), 106px below 390 (bar up to 101 at 360).

### S3P · FarmSummary row stretch, 16.11.2026
CSS in additions.css section 21 (moved out of the S3 screen CSS `.s3-fsums`). To copy into the system's FarmSummary README with HANDOFF. Decided by default (user away), to review.

- **FarmSummary · stretch** (`.ip-fsums.ip-fsums--stretch`): panels that share a grid row take the row's height, so side-by-side farms end on one line (S3 desktop: Telița and Hîrbovăț, real, empty or no-access panels alike). Default `.ip-fsums` keeps `align-items:start` (phone stacks, where equal heights mean nothing). S2 compact rows may use it. Tokens only, no new values.

### S11P · DataTable numeric headers, 16.11.2026
CSS in additions.css section 22. To copy into the system's DataTable README with HANDOFF. Decided by default (user away), to review.

- **DataTable · numeric headers** (`.ip-table th.num`): right-aligned like their column, but set in the caption face (Plex Sans 11/14 500, uppercase, `ink-muted`) like every other header. The bundle's `.ip-table .num` (Mono, tabular) now applies to cells only. Affects S3, S11 and the DataTable, CostBlock, GuideChart-variants and RecordCode-plus previews; the S3 screen rule `.s3-table th.num` font-family is now redundant (its `white-space:normal` stays). No new values.

### SYS08 · Acasă blocks: MissingRow, EntryKeys, UnpricedCount, 16.11.2026
Folders `design/system/components/MissingRow/`, `EntryKeys/`, `UnpricedCount/` (README + preview, 1440 and 390), CSS in additions.css sections 23–25. For S2 Acasă (vet, manager, owner). Decided by default (user away), to review.

- **MissingRow · Lipsesc date** (spec C9): `section.ip-missing` > head (`h2.ip-cap` "Lipsesc date" + scope note "Hîrbovăț · azi de la 12:00 · ultimele 7 zile", time and count in `ip-num`) > `.ip-list` of `a.ip-ev.ip-ev--missing`, one row per section × entry × day, newest day first. Anatomy: RecordTile (Zilnic = mortality, Hrană zilnică = feed wheat) · `.ip-ev__title` 500 ("H3.2 · Hrană zilnică ♂", type + sign in `.ip-ev__type`, no-wrap; farm first when the user has several farms; hidden "Completează: " and sex words) · meta Mono "16.11.2026 · SAPT 40 · ♂ 426 în viață · ieri 157 g" · `.ip-ev__status` warn Lamp **lipsă** · `.ip-ev__go` 16px chevron. Variants: link (today; any day for manager/owner) · locked `div.is-locked` for a vet's past day (D07: Tag "blocat" after the title, meta "zi trecută · o completează managerul", no chevron, lamp kept). States: hover `sunken`, pressed `rule` (chevron + lamp word `ink-secondary`), focus ring, empty after 12:00 (off Lamp "nimic lipsă" + "Toate secțiunile cu păsări au Zilnic și Hrană zilnică în ultimele 7 zile."), before 12:00 with nothing missing ("Zilele 09.11.2026–15.11.2026 sunt complete. Lipsurile de azi apar de la 12:00."), before 12:00 with past days missing (past-day rows only, today not listed yet), loading (`ip-ev--skel`, `aria-busy`); in the empty sentences "7" and "12:00" are `ip-num`. Only sections with birds. A11y: each row one link with a full name; rows ≥ 48px. Pressed rows with ♀ figures are 4.46:1 (D66, accepted momentary state, flagged for the system artifact).
- **EntryKeys** (spec D2): `section.ip-entrykeys` (size container) > `h2.ip-cap` "Înregistrează" > `ul.ip-entrykeys__list` > `a.ip-btn.ip-btn--secondary.ip-entrykey`: RecordTile + label + hint (`ip-note`, what the form records). Vet: Zilnic · Hrană zilnică · Cântărire; manager adds Transfer · Vânzare · Livrare hrană (package). All secondary (no primary on Acasă). Grid ≥ 168px columns: 6 in a row at 1136, 2 columns on phone with an odd last key across; keys ≥ 48px tall on every size. States: hover, pressed (hint → `ink-secondary`), focus, disabled (`a[role=link][aria-disabled=true]` no href, tile loses its record colour, one note under the grid with the reason, referenced by `aria-describedby` on each disabled key). Owner: no EntryKeys on Acasă (Înregistrează tab, D66).
- **UnpricedCount · Livrări fără preț** (spec D2 owner): `section.ip-unpriced` (`panel`, 1px `line`, `radius-md`, 16px) > head (coins RecordTile + `h2.ip-cap`) > count in the readout style ("4<small> livrări</small>", real space) + total kg (Mono 14 `ink-secondary`) > one line per farm × feed type ("Hîrbovăț · Breeder 2 | 3 livrări · 72 068 kg", count and kg in `ip-num`, split by `rule`) > note (oldest date; unpriced kg are not in cost) > secondary key "Adaugă prețurile" → S11 › Livrări (fără preț). States: default, none ("0 livrări · Toate livrările au preț." + "Vezi livrările"), loading (`.ip-unpriced__skel` stack of static `ip-skel`, `aria-busy`). No lamp (no threshold). Owner only (D06, D64).
- **Tokens:** existing only; no new vars. One container breakpoint (EntryKeys, 519.98px = where a third 168px column fits). Chevron = Lucide "chevron-right".

### SYS09 · EventRow actions + ChangeRow, 16.11.2026
Folders `design/system/components/EventRowActions/`, `ChangeRow/` (README + preview, desktop and 390), CSS in additions.css sections 26–27. For S15 Introdus azi, S16 Corectează, S17 Istoric modificări and every event log. Builder/render: `design/system/scripts/sys09_build.py`, `sys09_render.py`. Decided by default (user away), to review; open points A–D below.

- **EventRow · actions** (`ul.ip-list.ip-list--act` > `li.ip-ev.ip-ev--act`): RecordTile · title + meta + optional `.ip-ev__mod` · action slot. Keys by role and day (C8, E1, D07): author same day until 23:59 → Editează (link to the form) + Șterge (button → Dialog); manager/owner past day → Corectează (link to S16) + Șterge (reason + Dialog); vet past own entry → locked, Tag "blocat" + reason in the meta ("zi trecută · o corectează managerul"); vet on someone else's entry → no slot (D71); viewer → no slot. S15 list head `div.ip-list__head` (caption title + note with the 23:59 window). Layout from the list width (container): ≥ 480px ghost keys at `control-h` with hidden record names; < 480px one touch IconButton "Acțiuni" (ellipsis-vertical) → `div.ip-menu.ip-ev__menu` (`hidden` while closed) > `div.ip-ev__items[role=menu]` of 48px menuitems + a sibling `p.ip-note` tied by `aria-describedby` (menu and disabled items). Marks: Tag `neutral` "corectat" + link "modificat de Ion Rusu · 15.11.2026, 10:12" to Istoric (24px / 48px hit area); same-day author edit = "modificat la hh:mm" in the meta, no Tag. States: key hover/pressed/focus; disabled with reason (window closed while open: focusable `aria-disabled` keys, Editează `a[role=link][tabindex=0]`, + `aria-describedby` to the meta reason; phone menu items `ink-disabled` + note); locked; deleted `.ip-ev--deleted` (steel tile, `<del>` figures, Tag "șters", "șters de …"). Rows are not targets (no row hover, also on mark-link hover). Mono date-times and "♀ și ♂" (`.ip-ev__sexes`) never split.
- **ChangeRow** (`ul.ip-list.ip-list--change` > `li.ip-ev.ip-ev--change`): when (Mono) · RecordTile (spec "table" = record type) · record link "Type · section ♀/♂" + "ziua dd.mm.yyyy · SAPT · farm", stretched over the row, to the section at that day · values, one per field: `<del>` old → `<ins>` new ("–" = șters) · "Motiv: …" ("– în aceeași zi, fără motiv" when none) · who + role · Tag `neutral` corectat / șters / editat · chevron. Phone stacked; ≥ 720px one line under an `aria-hidden` column head. States: hover `sunken` (Tag → `rule`), pressed `rule`, focus ring on the row, empty `li.ip-list__empty` (off Lamp "fără modificări" + sentence + "Schimbă perioada sau filtrele."), loading (`ip-ev--skel`).
- **Tokens:** existing only; no new vars. Container breakpoints 480px (EventRow keys) and 720px (ChangeRow line). New generic helpers: `.ip-list:is(ul,ol)` reset, `.ip-list__empty`, `.ip-list__head`.
- **Open (defaults applied):** A — the author's same-day edits/deletes are logged in Istoric (Tag "editat"/"șters", no reason) and the row shows "modificat la hh:mm" without "corectat". B — a manager/owner changing someone else's entry of today uses Corectează (reason + audit), Editează only for one's own same-day entries. C — phone actions behind one "Acțiuni" menu key, desktop inline keys. D — a deleted record leaves logs and totals; the struck row shows only where history is shown (the section-day page opened from a ChangeRow).

### SYS10 · SectionTree for S19 (Ferme și secțiuni), 17.11.2026
Folder `design/system/components/SectionTree/` (README + preview, desktop: 440px farm panel and page width), CSS in additions.css section 28. For the S19 farm panel (spec B1). Builder/render: `design/system/scripts/sys10_build.py`, `sys10_render.py`. Decided by default (user away), to review; open point A below.

- **SectionTree** (`div.ip-stree` > list head + `table.ip-table.ip-stree__t`, the DataTable unchanged): farm row = DataTable group row (`tr.is-group.ip-stree__farm`: name, guide Tag `neutral` or "fără ghid" on `rule`, note "ouat · 3 hale · 6 compartimente"); section/hall rows `ip-stree--l2` (hall note "2 compartimente"); compartment rows `ip-stree--l3`, one level only, hung from a hair `line` rail. Columns Secțiune · Stare (Lamp ok "cu păsări" / off "goală"; Tag neutral "retrasă") · one IconButton "Acțiuni" → ip-menu: Redenumește · Mută în sus · Mută în jos · Adaugă compartiment (sections and halls) · Retrage (`aria-haspopup="dialog"`), with a note tied by `aria-describedby`. Add row: ghost "Adaugă secțiune" per farm. Edit row (add/rename): visible `ip-label` above the field ("Compartiment nou în H1", "Cod nou pentru H5-2") + medium TextField 96px + Salvează (secondary) + Anulează (ghost), error line under it. Live region `p.ip-sr[role=status]` announces moves ("H3 mutată în sus: a 2-a din 6 secțiuni la Telița.").
- **Keyboard:** menu keys ↑/↓, Home/End, Enter, Esc returns focus to the key (APG menu button); disabled items focusable but inert. Edit row: focus to the field on open, Enter = Salvează, Esc = Anulează, then focus back on that row's Acțiuni key (the new row's key after adding). Mută în sus/jos: focus stays on the moved row's key + live announcement. After the Retrage Dialog: focus on the next active row's key.
- **Rules:** B1 "a section with birds can't be retired" → Retrage `aria-disabled` with the reason ("H1.1 are păsări. O poți retrage când e goală."); a hall with active compartments likewise. Retire never deletes: records stay in reports; Dialog "Retrage secțiunea H5-2 de la Telița", focus on Anulează, destructive "Retrage secțiunea". Retired rows stay listed last, no key, no reactivation. Rename allowed with birds (same section, records follow the code). Reorder within siblings only; the parent is fixed once saved.
- **States:** key hover/pressed/focus/open; menu item current/pressed/disabled; duplicate code, empty code, save failed (fault Banner, typed code kept), load failed (Banner "… nu s-au încărcat. Verifică conexiunea și apasă Reîncearcă." + Reîncearcă), loading (skeleton, farm row `is-group`, `aria-busy`), farm without sections (off Lamp "fără secțiuni"), no farms (EmptyState "fără ferme" + "Adaugă prima fermă"), farm without guide (Tag "fără ghid").
- **Tokens:** existing only; no new vars. Desktop only (S19 lists no phone layout).
- **Open (default applied, D75):** A — each tree action saves on its own (inline Salvează per row, Retrage through its Dialog); the panel's Salvează covers only name, type and guide (rejected for now: batching tree edits into the panel's Salvează with "nesalvat" marks).

### SYS11 · GuideImportPreview for S20 (Ghiduri import), 17.11.2026
Folder `design/system/components/GuideImportPreview/` (README + preview, desktop 672px), CSS in additions.css section 29. For the S20 import panel (spec B2). Builder/render: `design/system/scripts/sys11_build.py`, `sys11_render.py`. Decided by default (user away), to review; open points A–B below.

- **GuideImportPreview** (`section.ip-gimp`): source line (file name or "Tabel lipit" + format · type, counts "51 rânduri citite · 48 valide · 3 refuzate", ghost "Alege alt fișier") → fault Banner listing refused rows as links to the refused cell → info Banner "Nu blochează salvarea." (blank cells, absent guide columns: `–` fără date în ghid) or ok Banner when all valid → SegmentedControl Toate / Refuzate (toggle buttons) → DataTable (`ip-table`, unchanged) in a focusable scroll box (12 rows, sticky header), one row per file row: Rând · (Vârstă, ouat) · SAPT · Sex · guide columns present · Verificare → primary "Salvează ghidul" (`aria-disabled` + reason note while rows are refused) + Anulează.
- **Refused cell** `td.ip-gimp__bad`: value as read, fault text on `fault-bg` with a hair fault edge, `tabindex=-1`, `aria-invalid`, `aria-describedby` → the reason after the fault Lamp "refuzat" in Verificare. Duplicate week + sex marks SAPT + Sex; non-number and missing sex mark their cell. Partial guide: `–` + "fără date în ghid: greutatea" guide line (D29).
- **Keyboard:** error links → refused cell (focus ring inside the cell) → Tab on to Salvează; filter = toggle buttons; table box is a focusable region; Salvează stays focusable while disabled and reads its reason; a polite live region announces the counts when reading ends; the table is described by the counts line.
- **States:** se citește (skeleton, `aria-busy`), toate valide (creștere file, ouat paste), rânduri refuzate, filtru Refuzate + focused cell, format greșit (.pdf → fault Banner + FileDrop), necitibil (no SAPT/Sex header), gol (EmptyState "fără date").
- **Tokens:** existing only; no new vars. Desktop only, ≥672px.
- **Open (default applied):** A — any refused row blocks saving the whole guide (fix the file, re-import) (rejected: save the valid rows and skip the refused ones). B — one preview row per file row (week × sex), the file's own line numbers (rejected: one row per week with ♀/♂ column pairs, as the guide detail). Small defaults: ouat files read by age week, SAPT = vârstă − 18 shown beside (D26); only the guide columns the file has are shown, the rest listed as fără date în ghid; no editing in the preview; G/C, F/M, ♀/♂ accepted for sex.
