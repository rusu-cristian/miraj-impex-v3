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
- **Loading:** a spinner, the label stays. While busy (`aria-busy="true"`, usually with `aria-disabled="true"`) the key keeps its variant's colours, not the disabled grey (Additions › S14P, section 37).
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

- **Which widget (FIN10, D130):** when the options carry state (a Lamp + word, a ♀ · ♂ count, a reason an option is disabled, a dates line, a mark such as "ghidul fermei"), the Select is the APG select-only combobox `.ip-select` (additions.css §53, Additions › FIN10); a plain list of names stays a native `<select>` inside `.ip-input`. Closed, the two look the same (value 400, 17/24 on touch, `ink-secondary` chevron). One field is one widget on every board.


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
- **Record-type columns:** the header carries the record icon in its colour, and ♀/♂ columns are coloured. Head icon: `span.ip-table__ico` (+ `ip-rec ip-rec--<type>`), 14px on the caption line (additions.css §32, FIN14B).
- **Groups:** `sunken` rows.
- **Totals:** a `line-strong` top border, one edge wherever the row sits (after a body row, in a following tbody or a tfoot: the row above drops its rule; §32, FIN14B).
- **Inline edit cells** (owner prices) use a medium TextField.
- **Phone:** horizontal scroll with the first column sticky.
- **Row headers, wrapping heads, scroll region** (FIN14, additions.css §32/§36): rows named by `th[scope=row]` use `ip-table--rowheads` (tbody and tfoot); figure heads may wrap with `ip-table--wraphead`; a table that can outgrow its box sits in `div.ip-tablescroll` (role=region, tabindex=0, aria-label = caption; `--framed` draws the box). No screen copies (see Additions › FIN14).
- **Current row** (S22P §45, FIN05 §51): the record open in a SidePanel beside the list: `tr.is-current` + `aria-current="true"` on the row's key (not drawn pressed); `sunken` with a hair `line-strong` frame inside, steel only, no lamp; distinct from hover (no frame) and focus (ring on the key).


## EventRow

One record in a log: RecordTile, a title with section and figures (♀/♂ coloured), meta (date, time · SAPT · author: "16.11.2026, 08:34 · SAPT 40 · Dr. Ceban", FIN09), then a status lamp or tag.

- **Row:** height ≥ 48px, separated by `rule`; hover `sunken`.
- **Filtering:** by RecordChips in the FilterBar.
- **Locked rows** show a "blocat" tag.
- **Current row** (FIN05 §51): the record open in a SidePanel beside the log: `li.ip-ev.is-current` + `aria-current="true"` on its key (Corectează, not drawn pressed); the DataTable current-row look (`sunken`, hair `line-strong` frame), steel only.


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

**With action (FIN03, additions.css §49):** every Banner that carries a key (Reîncearcă, a Button link, an `ip-link` such as Alege ghidul) is `div.ip-banner.ip-banner--act` > 16px icon, `div.ip-banner__body` (the sentences; `p`s space-2 apart), then the key with `.ip-banner__key`. The icon stays on the first text line (§47); the body wraps; the key sits at the end, centred on the Banner's height. When the body would get narrower than space-12 × 6 (288) beside it (phone, SidePanel), the key wraps under the text, at the text's start, space-3 below. Text that stays one line beside a Button key adds `ip-banner--1l` (text, icon and key on one centre line). A load error says what failed and the fix in two sentences: "<what> nu s-au încărcat." / "Verifică conexiunea la internet, apoi reîncearcă." (+ what still works, "Poți … și acum."), `role="alert"`.


## Toast

A save confirmation in `control` (graphite), with an optional undo. Not used for errors.

**Placement (FIN02, additions.css §48):** phone: `div.ip-toastwrap` inset `space-4` at the foot of the screen's view, above the SaveBar or BottomTabBar (in the phone full-screen sheet: above its foot); desktop: `ip-toastwrap--page`, bottom right of the page, `space-8` from both edges, at most `panel-w` wide. A save Toast carries Anulează: `span.ip-toast__txt` + `button.ip-link.ip-toast__undo`.


## EmptyState

An unlit lamp with one sentence and a next step, e.g. "Ferma e goală din 30.09.2026." or "Toate livrările au preț.".


## Skeleton

Static `sunken` bars, no shimmer.

- **System skeletons (Additions):** EventRow rows `ip-ev--skel` (section 23), DataTable rows `tr.ip-table__skel` (section 42), Readout `div.ip-readout.ip-readout--skel`, CycleTrack `div.ip-trackskel` (`--compact`), field `div.ip-field.ip-field--skel` and text line `span.ip-skel-line` (section 54). Bar widths: `ip-skel--w1|w2|w3` (192/144/72), `--ch1…4` (as wide as 1–4 figures, in `ch` of the cell's Mono), `--key` (an IconButton, `control-h`), `--end` (at the cell's right edge). Screens draw these, not screen copies (FIN11); generators take the markup from `design/canvas/scripts/skel.py`.
- **The skeleton is the loaded geometry (craft S6P, FIN11G):** a skeleton uses the loaded block's own boxes, so nothing moves when data arrives. Readout: the Readout's own line elements, one bar each (StockReadout: cap, value, 3 kv rows, sub; with a GuideChart: its empty svg at the loaded viewBox); Stat `ip-stat--skel`: its own lines, guide lines as wide as the loaded ones. CycleTrack: the drawn track's height (full 120, `--compact` 90) and its legend line. Field: the field's own label line, box (sunken, no edge) and hint lines. Table: the loaded and the loading table share one head that declares the column shares (`th[style="--ip-colw:n%;--ip-colmin:…"]`, one text column takes the rest, the loaded min-content as a strut); skeleton bars never widen a column.


## Dialog

Confirms destructive and cycle-changing actions.

- **Style:** `panel`, radius `radius-lg`, `shadow-overlay`, on `scrim`.
- **Title:** names the action and the object.
- **Fields:** a reason where required.
- **Focus:** starts on Anulează.
- **Placement (FIN04, additions.css §35/§50):** one scrim per frame, the Dialog centred both ways in it: desktop `div.ip-scrim.ip-scrim--page` fills the page (Sidebar, page and any open SidePanel, above the Toast), space-6 round the Dialog; phone `ip-scrim--phone` fills the phone screen, space-4 round it. The page and panel under it are `inert`; `role="alertdialog"` `aria-modal="true"`; body line `p.ip-dialog__note`, keys `div.ip-dialog__keys` (phone `--stack`), Anulează first.


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

- **Breadcrumbs hit area (FIN06, additions.css §52):** each parent link is a target of its own without moving the line: desktop ≥24px tall (`space-6`, D46; word + 5px each side, e.g. "SERIA 2026" 79×24); its focus ring is drawn round the word by `::after` and stops at the h1's top edge. Phone page heads use `nav.ip-crumbs.ip-crumbs--touch` with the word in `span.ip-crumbs__txt`: ≥44px (46 tall, word + 8 wide, "SETĂRI" 50×46); the word takes the focus ring.


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
- **Sidebar series selector** (`SeriesSelector/`): APG select-only combobox, since FIN10/FIN13 the section 53 Select `div.ip-select.ip-series > div.ip-input.ip-select__btn[role=combobox][tabindex=0]` (`aria-expanded`, `aria-controls`, `aria-activedescendant`) under the wordmark: name, Lamp + word ("deschisă" ok / "închisă" off), chevron; `listbox` (`ip-menu`) over the nav, options with a check slot, name + Lamp, dates in Mono ("până la 28.01.2026"). Keyboard: ↓/Enter/Space/Alt+↓ open, ↑↓ Home End PgUp PgDn and type-ahead move, Enter/Space/Alt+↑ select, Esc close, Tab selects and moves on; focus stays on the combobox. `a.ip-nav` has no underline.
- **FilterBar Interval** (`FilterInterval/`): with Interval selected, `.ip-interval` holds "de la" and "până la" DateFields (labels above, Mono `dd.mm.yyyy` at `10ch` wide, 24px calendar key), default cycle start → today; at 1180px the bar stays one row with Exportă (`.ip-filterbar__end`). Future dates: typed → error row in flow under the bar (`.ip-filterbar__err`) "Până la 20.11.2026 e în viitor. Alege o dată până la 16.11.2026."; calendar = APG date-picker dialog with `table[role=grid]`, `aria-selected` on the gridcell, `aria-current="date"` on today, days after today `aria-disabled` (`sunken` + `ink-disabled`), next month disabled; selected day `control`. Keys: arrows, Home/End, PgUp/PgDn, Enter picks, Esc closes; focus returns to the calendar key.
- **Helpers:** `.ip-sr` (visually hidden), vars `--ip-hair`, `--ip-edge-strong`, `--ip-mark-w`, `--ip-fs-*/--ip-lh-*` (type sizes from tokens.css) so additions carry no raw values.

### SYS02 · FeedTotals for S7 (Hrană zilnică), 16.11.2026
Folder `design/system/components/FeedTotals/` (README + preview), CSS in additions.css section 8. Decided by default (user away), to review.

- **Purpose:** the per-feed-type totals under the AllocationRows (spec C2 step 4): what today's confirmed rows take from each feed type's stock and how many days are left. Plus a compact variant for the SaveBar summary.
- **Anatomy:** `section.ip-ftot` = SectionTotals' strip (`panel`, 1px `line`, top `line-strong`, `radius-md`, rows split by `rule`); head `ip-cap` "Total pe tip de hrană" + "dd.mm.yyyy · doar rânduri confirmate"; one `.ip-ftot__row[role=group]` per feed type used today at the farm: RecordTile wheat (`rec-feed`) · type name + for-sex sign (`rec-hen`/`rec-rooster`) · "n din m rânduri confirmate" · Azi (kg, one decimal, Plex Mono 16/20 500 like the AllocationRow kg) · Stoc după azi (whole kg, "din <stoc acum> kg") · Zile rămase (stoc după azi ÷ average of the last 3 days, one decimal, "până pe dd.mm.yyyy") · Lamp + word. Foot `ip-note` names the formula, the averaging window and the threshold. Phone by default (type line + lamp, then three right-aligned figure cells); `ip-ftot--wide` for desktop (one line per type, `aria-hidden` column header, 120px status column).
- **States:** în normă (`ok`, ≥ 5 zile) · abatere (`warn`, < 5 zile, D19) · stock below 0 (`warn` lamp "abatere", D50, with "stoc sub 0" as the Stoc după azi sub-line, + the warn Banner with the spec sentence "Stocul de Breeder 2 la Hîrbovăț ar ajunge la −1 240 kg. Verifică livrările.", days "0,0 · se termină azi"; a warning, Salvează stays enabled) · nothing confirmed (Azi "–" "nimic confirmat", stock = "stoc acum", days from it) · rows left out named under Azi ("fără ♂ H3.2").
- **Compact (SaveBar):** `.ip-savebar__sum.ip-ftot-sum` (FIN15: now `.ip-savebar__sum--2l` > `.ip-savebar__line`), two lines: wheat + "10 din 12 rânduri · 3 691,3 kg"; "2 de confirmat ·" + one stock Lamp, worst type first ("Cocoși 3,9 zile" warn / stock below 0: warn "abatere" + detail "Breeder 2 · stoc sub 0" (D50) / "stoc în normă" ok). Counts rows (section × sex), not sections. Nothing confirmed: "0 din 12 rânduri · – kg", Salvează disabled.
- `ip-ftot--wide` is for desktop only.
- **Tokens:** existing only; new var `--ip-fs-alloc/--ip-lh-alloc` (16/20) names bundle.css's `.ip-alloc__kg` size.
- **A11y:** static (no targets); type rows are labelled groups ("Breeder 2, pentru găini"); per-cell keys stay for screen readers in the wide layout; the stock warning is `role=status`; kg figures are not live regions (no announcement per keystroke); every lamp has its word.

### SYS03 · FarmSummary for S3 (series) and S2 (Acasă, owner), 16.11.2026
Folder `design/system/components/FarmSummary/` (README + preview), CSS in additions.css section 9. Decided by default (user away), to review.

- **Purpose:** one panel per farm for a series (S3 §3, spec D1 "a summary per farm"), and the owner's farm summaries on Acasă (S2, compact).
- **Anatomy:** `article.ip-fsum` (`panel`, 1px `line`, `radius-md`, 16px padding). Head: farm name `h3` 16/22 500 as the panel's one link (ink, underlined; to the farm page at that cycle; inline-flex, min-height `--space-6` = 24px target in narrow/`--wide`/`--compact`, 48px in `--touch`, D46) · type Tag and guide Tag (`neutral`) · cycle Lamp + word at the right (`ok` deschis / `off` închis / `off` goală; these cycle-state words are allowed outside the D42 status vocabulary, state not status, D46). Mono context line: "ciclu din 10.02.2026 · SAPT 40" or "07.10.2025–28.02.2026 · 21 săpt.". Alive: caption + SexCount at `readout` size + "total n" (closed cycle: "Populat · 07.10.2025" with the placement, since alive at close is 0). Key figures: `dl` of rows split by `rule`, SectionTile's kv language: record icon 14 + key · value Plex Mono 14/20 500 `ink` right-aligned (♀ · ♂ as SexCount) · sub-line 12px `ink-muted` (guide, deviation, week, exclusions or `ghid: – fără date în ghid`) · Lamp + word only where a guide or D19 threshold exists, words from the D42 Lamp vocabulary only ("în normă" / "abatere" …); the detail ("+6,8%", "sub 5 zile") goes in the sub-line. Every ♀/♂ sign, keys and sub-lines included, is an `ip-sex` span. Percentages without a space ("74,1%", "−7,3%"; value unit `small.ip-fsum__pct`, D46).
- **Figures:** laying, open: Ouat azi vs ghid · Mortalitate cumulată ♀ · ♂ % · Hrană azi ♀ · ♂ g vs ghid (D38 weighting, names left-out rows) · Zile stoc ♀ · ♂ (package; `warn` "abatere", sub-line "Cocoși până pe 19.11.2026 · sub 5 zile", the short type named). Growth, closed: Greutate ♀ la transfer vs ghid (±5%; outside → `warn` "abatere", "+6,8%" in the sub-line) · Mortalitate cumulată · Hrană totală g pe pasăre = Σ daily g/pasăre over the days with feed rows vs the guide over the same days (D46; Telița ♀ 8 401 · ♂ 10 142 vs 9 065 · 10 835 g, 143 days → `ok` "în normă"), total kg ♀ · ♂ on the sub-line · Transferate ♀ · ♂ (arrivals, D28).
- **Variants:** narrow by default (phone; Lamp right of the sub-line) · `ip-fsum--wide` (S3 desktop, `.ip-fsums` 2 columns; Lamps in a fixed 96px column like AlertRow) · `ip-fsum--compact` (S2 owner: 12px padding, alive inline at figure size, 2 figures: Ouat azi and Mort. cumulată, because feed stock has its own S2 block) · `ip-fsum--touch` (phone: links reach 48px).
- **Empty farm between cycles:** Lamp `off` "goală" + `.ip-fsum__empty`: "Ferma e goală din 28.02.2026. Ultimul ciclu: Seria 2026, 21 săpt." and the next step link "Vezi ultimul ciclu". S2 shows a farm as it is today; S3 shows the farm's cycle in that series.
- **Rules:** no Readout inside (one per section; S3's moment is the CycleTrack), no big icons, values stay `ink` except ♀/♂; unknown = "–" with its reason. Guide Tag shows the guide's full name even past 16 characters ("Ross 308 creștere", 17).
- **Tokens:** existing only; new var `--ip-fs-heading/--ip-lh-heading` (16/22) names tokens.css `.ip-t-heading`.

### SYS04 · StockReadout, 16.11.2026
Folder `design/system/components/StockReadout/` (README + preview), CSS in additions.css section 10. Decided by default (user away), to review.

- **Purpose:** S11 › Stoc (spec D4 §1): one dark Readout per farm × feed type, repeated in a row (the craft.md exception to one Readout per section).
- **Anatomy:** `ul.ip-readouts` (StockReadoutRow) > `li` > `div.ip-readout.ip-readout--stock[role=group]`: caption "HÎRBOVĂȚ · BREEDER 2" (Mono uppercase `readout-muted`, one line; the for-sex follows as hidden `ip-sr` text ", pentru găini", no `aria-label` on the span) with `ip-rlamp` + word on the right · value = kg in stock (`readout-lg`, whole kg) · `dl.ip-readout__kv` above a `readout-line` hairline: zile rămase (one decimal) · medie/zi 13.11–15.11 (last 3 complete days) · se termină (run-out date) · optional `ip-readout__sub` with the reason.
- **States:** `ok` în normă (≥ 5 zile) · `warn` abatere (< 5 zile, D19; "sub pragul de 5 zile") · stock below 0: `warn` abatere (D40 warning, D50 wording; sub-line "stoc sub 0 · verifică livrările", "epuizat pe") · `off` oprit with "– fără consum în ultimele 3 zile" · `off` oprit for a farm without an open cycle (`ip-readout--off`: one readout per farm, value "–" in `readout-muted`, "fără ciclu deschis"; Telița never shows its −716 kg workbook Sold).
- **Row:** `auto-fill` columns ≥ 240px: 4 per row at 1136px (2–3 readouts keep that width), one column on phone. An `ip-note` under it names the formula, window and threshold.
- **No record icon** on the readout: `rec-feed` is 2,92:1 on `readout` (< 3:1); the caption word names the type.

### SYS04 · CostBlock, 16.11.2026
Folder `design/system/components/CostBlock/` (README + preview), CSS in additions.css section 11. Decided by default (user away), to review.

- **Purpose:** S11 › Cost (spec D4 §4): lei per cycle from priced deliveries, lei per bird, lei per egg, and the "X kg fără preț" not included. A compact DataTable, not a KPI grid.
- **Anatomy:** `section.ip-cost` > `.ip-cost__head` (`panel`, caption + "din livrările cu preț · până pe dd.mm.yyyy") > `.ip-cost__scroll` (`role=region`, `tabindex=0`, `aria-label` = caption, every layout since FIN14B) > `table.ip-table.ip-cost__table` (hidden `caption.ip-sr` = the head title; Ciclu · Kg · Lei · Lei/pasăre · Lei/ou), one `tbody` per cycle: cycle row (`th[scope=row]` name + sub-line, Mono right-aligned figures) and `tr.ip-cost__unpriced` (hidden `ip-sr` cycle prefix "Hîrbovăț · Seria 2026, ", coins `rec-feed` + info Tag "fără preț", unpriced kg in the Kg column, "n livrări, dates · neincluse în cost"; none unpriced: "Toate livrările au preț.") · totals row `is-total` (§32 totals edge, FIN14B) "Total Seria 2026" naming the excluded kg · `ip-cost__foot` note naming the bases.
- **Bases:** Lei/pasăre per bird at the start of the cycle (placed or arrived by transfer), not alive; Lei/ou = lei ÷ total eggs (D18) to date, laying only (growth "–" "creștere, fără ouă"). Whole lei for totals, 2 decimals for unit costs.
- **Layouts:** default = phone (DataTable rule: horizontal scroll, sticky 160px first column with a `rule` edge, focusable labelled region; no `overflow:hidden` on the block so the focus ring shows whole, head/foot round their own `radius-md` corners); `ip-cost--wide` for desktop (one-line row headers; the same sticky first column with its `rule` edge and the focusable labelled region, FIN14B).
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

- **MissingRow · Lipsesc date** (spec C9): `section.ip-missing` > head (`h2.ip-cap` "Lipsesc date" + scope note "Hîrbovăț · azi de la 12:00 · ultimele 7 zile", time and count in `ip-num`) > `.ip-list` of `a.ip-ev.ip-ev--missing`, one row per section × entry × day, newest day first. Anatomy: RecordTile (Zilnic = mortality, Hrană zilnică = feed wheat) · `.ip-ev__title` 500 ("H3.2 · Hrană zilnică ♂", type + sign in `.ip-ev__type`, no-wrap; farm first when the user has several farms; hidden "Completează: " and sex words) · meta Mono "16.11.2026 · SAPT 40 · ♂ 426 în viață · ieri 157 g" · `.ip-ev__status` warn Lamp **lipsă** · `.ip-ev__go` 16px chevron. Variants: link (today; any day for manager/owner) · locked `div.is-locked` for a vet's past day (D07: Tag "blocat" after the title, meta "zi trecută · o completează managerul", no chevron, lamp kept). States: hover `sunken`, pressed `rule` (chevron + lamp word `ink-secondary`), focus ring, empty after 12:00 (off Lamp "nimic lipsă" + "Toate secțiunile cu păsări au Zilnic și Hrană zilnică în ultimele 7 zile."), before 12:00 with nothing missing ("Zilele 09.11–15.11.2026 sunt complete. Lipsurile de azi apar de la 12:00."), before 12:00 with past days missing (past-day rows only, today not listed yet), loading (`ip-ev--skel`, `aria-busy`); in the empty sentences "7" and "12:00" are `ip-num`. Only sections with birds. A11y: each row one link with a full name; rows ≥ 48px. Pressed rows with ♀ figures are 4.46:1 (D66, accepted momentary state, flagged for the system artifact).
- **EntryKeys** (spec D2): `section.ip-entrykeys` (size container) > `h2.ip-cap` "Înregistrează" > `ul.ip-entrykeys__list` > `a.ip-btn.ip-btn--secondary.ip-entrykey`: RecordTile + label + hint (`ip-note`, what the form records). Vet: Zilnic · Hrană zilnică · Cântărire; manager adds Transfer · Vânzare · Livrare hrană (package). All secondary (no primary on Acasă). Grid ≥ 168px columns: 6 in a row at 1136, 2 columns on phone with an odd last key across; keys ≥ 48px tall on every size. States: hover, pressed (hint → `ink-secondary`), focus, disabled (`a[role=link][tabindex=0][aria-disabled=true]` no href, focusable (FIN13), tile loses its record colour, one note under the grid with the reason, referenced by `aria-describedby` on each disabled key). Owner: no EntryKeys on Acasă (Înregistrează tab, D66).
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

- **SectionTree** (`div.ip-stree` > list head + `table.ip-table.ip-stree__t`, the DataTable unchanged): farm row = DataTable group row (`tr.is-group.ip-stree__farm`: name, guide Tag `neutral` or "fără ghid" on `rule`, note "ouat · 3 hale · 6 compartimente"); section/hall rows `ip-stree--l2` (hall note "2 compartimente"); compartment rows `ip-stree--l3`, one level only, hung from a hair `line` rail. Columns Secțiune · Stare (Lamp ok "cu păsări" / off "goală"; Tag neutral "retrasă") · one IconButton "Acțiuni" (`aria-controls` → the `role=menu` element, FIN13) → ip-menu: Redenumește · Mută în sus · Mută în jos · Adaugă compartiment (sections and halls) · Retrage (`aria-haspopup="dialog"`), with a note about Retrage tied by `aria-describedby` to the Retrage item only (read once; FIN13). Add row: ghost "Adaugă secțiune" per farm. Edit row (add/rename): visible `ip-label` above the field ("Compartiment nou în H1", "Cod nou pentru H5-2") + medium TextField 96px + Salvează (secondary) + Anulează (ghost), error line under it. Live region `p.ip-sr[role=status]` announces moves ("H3 mutată în sus: a 2-a din 6 secțiuni la Telița.").
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
- **S20P amendments (18.11.2026, D115):** percentages keep one decimal (`fmt(v, '%')`: "89,0", as the S20 detail tables); se citește = six section 42 rows (`tr.ip-table__skel`, `aria-hidden`), each bar as wide as its column's figures in `ch` of the Mono cell (`ip-skel--ch1…4` since FIN11, were `ip-gimp__sk1…sk4`: Rând, SAPT, Vârstă 2 · Sex 1 · hrană 3 · greutate, ouat % 4), right-aligned on numbers, Verificare empty, no hidden text row (live region "Se citește <fișier>."); running text "sapt. 24–25" (labels, links and Mono cells keep SAPT; the duplicate reason reads "Rândul 8 are deja sapt. 4 ♀. Păstrează un singur rând."); no-header fault "… nu are antetul cu coloanele SAPT și Sex. Pune pe primul rând numele coloanelor de sub chenar și încarcă fișierul din nou."; empty file: one "Alege alt fișier" (the EmptyState's; the source line has no key, counts keep the `space-3` inset); scroll box max-height adds its two hair edges; the FileDrop is the paste target (section 43). Helpers `skel_body`, `empty_file`, `NOHEAD`, `EMPTY_NOTE`, `DROP` live above the preview blocks so S20/S20S reuse them.
- **Open (default applied):** A — any refused row blocks saving the whole guide (fix the file, re-import) (rejected: save the valid rows and skip the refused ones). B — one preview row per file row (week × sex), the file's own line numbers (rejected: one row per week with ♀/♂ column pairs, as the guide detail). Small defaults: ouat files read by age week, SAPT = vârstă − 18 shown beside (D26); only the guide columns the file has are shown, the rest listed as fără date în ghid; no editing in the preview; G/C, F/M, ♀/♂ accepted for sex.

### SYS12 · UserRow list + OneTimePassword for S22 (Utilizatori), 17.11.2026
Folders `design/system/components/UserRow/`, `OneTimePassword/` (README + preview, desktop: list 1056px, dialogs 440px), CSS in additions.css section 30. For S22 (spec A2, owner). Builder/render: `design/system/scripts/sys12_build.py`, `sys12_render.py`. Decided by default (user away), to review; open point A below.

- **UserRow** (`section.ip-users` > `div.ip-list__head` + `table.ip-table.ip-users__t`, the DataTable unchanged): head = caption "Utilizatori" · counts note "5 utilizatori · 4 activi · 1 dezactivat" · primary "Adaugă utilizator" (the page's one primary). Row `tr.ip-user`: Nume (500; "(tu)" note on your row) · Login (Mono) · Rol (text: Proprietar · Manager · Veterinar · Observator, D48) · Ferme ("toate fermele" for the owner) · Ultima intrare (Mono `dd.mm.yyyy, hh:mm`; "– încă n-a intrat") · Stare (Lamp ok "activ" / off "dezactivat", state words as D46) · one IconButton "Acțiuni pentru …" → ip-menu (SYS10 pattern; opens downward, `--up` only when the room below the key to the viewport or scroll container bottom is less than the menu and the room above is larger, measured on open; S22P): Editează · Resetează parola · Dezactivează, or for a deactivated user Editează · Resetează parola (disabled, "Contul e dezactivat. Reactivează-l ca să-i dai o parolă nouă.") · Reactivează; own row: Dezactivează disabled ("Nu îți poți dezactiva propriul cont."). Deactivated rows `is-off` (`ink-secondary`), listed last; active rows by role then name.
- **Dialogs** (every SYS12 `dialog()` is `role="alertdialog"`, S22P, D120)**:** "Resetează parola lui Ion Rusu" (focus Anulează; primary "Resetează parola"; failure = fault Banner "Parola nu s-a resetat; cea veche merge în continuare. …" + Anulează + primary Reîncearcă, focus moves to Reîncearcă, Esc = Anulează) turns into the OneTimePassword; "Dezactivează contul lui Petrea" with the A2 sentence + "Poți reactiva contul oricând." (destructive key) → Toast; Reactivează has no dialog, Toast only.
- **OneTimePassword** (`div.ip-otp` in a Dialog): context line ("Parolă resetată · login ion.rusu" / "Cont nou · Observator · Hîrbovăț · login observator") · `label[for]` "Parolă temporară" · a real read-only TextField `input[type=text][readonly].ip-input.is-readonly.ip-otp__pw` (`value` = the password, `autocomplete=off spellcheck=false autocapitalize=off`, `aria-describedby` → once Banner + error line) full width, Mono 500 at the readout size (20/24), 12 characters spaced by `letter-spacing: space-1`, no groups (screen readers read the value; ←/→ read it character by character); `::selection` = `focus` fill + `raised` text · secondary "Copiază parola" + ok Lamp "copiată" when done · info Banner "Se arată o singură dată. Dă-i-o lui … acum. După ce închizi n-o mai poți vedea; o poți doar reseta din nou." · primary "Închide". Failed copy: `input.focus()` + `input.select()` on the password (preview: `is-selected`, preview-only: it fills the whole field where `::selection` fills only the text, so screens never use it as a field style), error line `p.ip-err.ip-fieldmsg` (alert icon) "Copierea nu a mers. Parola e selectată: apasă Ctrl+C." Generated: 12 characters, a–z without i l o + 2–9.
- **Keyboard:** Tab: Adaugă utilizator → one Acțiuni key per row (rows aren't targets). Menu: APG menu button (Enter/Space/↓ open on the first item, ↑ on the last; ↑/↓, Home/End, type-ahead; Esc/Tab close, focus back on the key); disabled items focusable, inert, reason via `aria-describedby`. Confirm dialogs start on Anulează (reset failed: Reîncearcă; Esc = Anulează; the fault Banner `#…-why` has no role, the dialog and Reîncearcă point `aria-describedby` at it, so focus reads "Reîncearcă, buton, Parola nu s-a resetat; …"); the password step starts on Copiază parola; Tab trapped (password → Copiază → Închide); Esc closes; focus returns to the row's key (the new row's key after adding). Live regions: copy result (`p.ip-sr[role=status]`), Toasts for deactivate/reactivate, list load failed (the §49 fault Banner is `role=alert`, the list's `p.ip-sr[role=status]` stays empty, so one announcement; FIN13).
- **States:** default, menu open, own row, deactivated row, never signed in, empty (own row + off Lamp "doar tu" + "Doar tu ai acces. Adaugă primul utilizator."), loading (two §42 `tr.ip-table__skel` rows since FIN07, `aria-busy`), load failed (head note "– lista nu s-a încărcat", the §49 Banner with action `role=alert` + Reîncearcă (FIN13), no table, status region empty); current row (the user open in a panel, §45 below; on the S22 page "Adaugă utilizator" is always the PageHeader's secondary key since FIN07, so the panel keys are the one primary); password after reset, for a new account, copied, copy failed; reset failed; deactivate confirm + Toast; reactivate Toast.
- **Copy failed announcement (S22S fix 1, D119):** as the SYS13 rule, focus moves to the password field, which carries the error line (aria-describedby); the copy status region stays empty then and announces only "Parola a fost copiată." (in SYS12 `otp()` since S22P: preview and every S22 board; the S22S screen-side override is gone).
- **Tokens:** existing only; no new vars. Desktop only (S22 lists no phone layout).
- **Open (default applied):** A — on Adaugă utilizator the app generates the temporary password and shows it once after Salvează, exactly like a reset; the add form has no password field, only the note "Parola temporară se generează la salvare și se arată o singură dată." (rejected: the owner types the temporary password in the form, A2's literal field list — weak, reused passwords and two flows). Small defaults: Romanian role names Proprietar / Manager / Veterinar (Observator per D48); reset asks for confirmation first (the old password stops working); no forced change at first sign-in is promised in the copy (A1 has no such flow); closing never asks again; no Toast after the password dialog; Reactivează without a dialog; deactivated users listed last; the list head owns "Adaugă utilizator".

### SYS13 · LoginCard + Wordmark for S1 (Autentificare), 17.11.2026
Folders `design/system/components/LoginCard/`, `Wordmark/` (README + preview: desktop page band, 9 card states at 440, 3 phone boards at 390 × 844), CSS in additions.css section 31. For S1 (spec A1). Builder/render: `design/system/scripts/sys13_build.py`, `sys13_render.py`. Decided by default (user away), to review.

- **Wordmark:** text only, written "Impex Line", drawn by bundle.css `.ip-wordmark` (uppercase, Sans 500, tracked; PRODUCT.md: no logo). Sidebar: `div.ip-wordmark` 15/20, unchanged. Login: `h1.ip-wordmark.ip-wordmark--login` at the title style 24/30 (`--ip-fs-title`/`--ip-lh-title`), the S1 page's only heading, same on phone. No icon, no tagline.
- **LoginCard** (`section.ip-panel.ip-login` > h1 wordmark > `form.ip-login__form[novalidate]`): [fault Banner summary `#…-msg[tabindex=-1]`, no role] · Login TextField (`autocomplete=username`, `autocapitalize=none`, `spellcheck=false`, max 40) · Parolă TextField (`type=password`, `autocomplete=current-password`) with a show/hide IconButton inside the field (`aria-label="Arată parola"`, `aria-pressed`, `aria-controls`; inset focus ring) · primary "Intră" full width (`type=submit`) · hint "Ai uitat parola? Vorbește cu administratorul." · `p.ip-sr[role=status]`. Page `main.ip-login-page`: desktop card at `panel-w`, top of a centred column on `ground`; phone `--phone` full width, touch sizes (48, 17px input).
- **States:** default (focus Login) · filled · password shown · signing in (spinner in Intră, label kept, form `aria-busy`, repeat presses ignored; if Intră itself gets `aria-busy`/`aria-disabled`, section 37 keeps it graphite, not grey) · empty fields ("Scrie loginul." / "Scrie parola.", checked on Intră only) · A1's four refusals as fault Banners, copy word for word: wrong credentials (both fields `aria-invalid` + described by the Banner, password emptied, focus Parolă) · too many attempts (Intră `aria-disabled` for the 5 minutes, focus Banner) · deactivated (focus Banner) · offline (values kept, focus Banner). `aria-invalid` alone draws the fault edge inside `.ip-login`.
- **Keyboard / announcements:** Login → Parolă → Arată parola → Intră; Enter in a field = Intră. Focus goes to the first field in error, or to the Banner when no field can fix it; the message is read once through that focus (field `aria-describedby` / the Banner's text). Only if the target already had focus (Enter in that field) does the status region say it instead. Never `role=alert` plus a live region.
- **Tokens:** existing only; two named type vars (`--ip-fs-title`, `--ip-lh-title`, copies of tokens.css `.ip-t-title`).
- **Small defaults:** show/hide password included (A1 doesn't forbid it; helps phone typing); no "forgot password" link, one hint line pointing to the administrator (A1: owner resets); no "rămâi conectat" (A1: 30 days by default); autofocus on Login; every refusal is fault (offline included); only emptiness is checked client-side (length rules are the server's "Login sau parolă greșită."); the card is centred in a top-aligned column, its content left-aligned; Intră full width on both layouts.

### S4T · DataTable row headers + framed scroll region, 17.11.2026
CSS in additions.css section 32 (both copies, identical). To copy into the system's DataTable README with HANDOFF (D32). Decided by default (user away), to review.

- **DataTable · row headers** (`table.ip-table.ip-table--rowheads`, opt-in): for tables whose rows are named by `th[scope=row]` (group rows by `th[scope=rowgroup]`). Row headers read as body text (13/18 Sans 400, `ink`, `raised`, `row-h`), not as the uppercase caption headers of thead. Group rows (`tr.is-group`) and totals (`tr.is-total`) give their th what bundle.css gives their td (`sunken` / 500; `line-strong` top edge on `panel`). `th.ip-table__child` indents a row under its group row (`space-6`). Two-row headers: a first column spanning both rows (`th[rowspan]`) sits at the bottom; colgroup heads align right over their figures on a `rule`; inside `.ip-tablescroll` the second head row's first cell does not stick.
- **DataTable · framed scroll region** (`div.ip-tablescroll.ip-tablescroll--framed`): the scroll region draws the table's box (1px `line`, `radius-md`, clipped by the region); the last row drops its rule against the frame. Focus stays the system ring on the region. Used by the S4 Fermă tabs (Păsări, Hrană, Greutate, Producție); since FIN14 also S11, S12, S13 (the older copies `.ip-cost__table`, `.ip-sparktable`, S3/S11/S12 screen rules are folded in, see FIN14). Tokens only, no new values.

### S9P · SectionSelect phone BottomSheet, 17.11.2026
CSS in additions.css section 33 (both copies, identical). Used by S9 Transfer (board "S9 · Transfer · alege secțiunea · exemplu de stare (D39)"); to copy into the system's SectionSelect README with HANDOFF (D32). Decided by default (user away), to review.

- **Anatomy:** `div.ip-scrim.ip-scrim--sheet` (fills the phone screen, sheet at the bottom) > `div.ip-sheet.ip-sheet--select[role=dialog][aria-modal=true]` (labelled by its title, described by its note) > handle · head (title 16/22 "În: alege secțiunea" + note "O fermă goală începe un ciclu nou.", IconButton touch "Închide lista", Lucide "x") · `div.ip-sheet__list[role=listbox][tabindex=0][aria-activedescendant]` of `[role=group]` labelled by their `ip-menu__group` caption (farm · type) > `div.ip-menu__item[role=option]`: section code (500, hidden farm prefix, so options read "Hîrbovăț H1.1 …") then the ♀ · ♂ alive SexCount pair or a reason word.
- **Combobox:** on phone the SectionSelect `div.ip-input[role=combobox][tabindex=0]` has `aria-haspopup="dialog"` and `aria-expanded`; `aria-controls` points at the sheet only while it is open (APG), and no empty hidden listbox stays in the page. No `aria-activedescendant` on the combobox (it moves to the sheet's listbox). Desktop keeps the APG select-only listbox combobox (SexCountRow).
- **Options:** rows 48 (`control-h-touch`). Reason words instead of figures: "sursa" (the Din section) and "în listă" are `aria-disabled` (`ink-disabled`); "goală" is `ink-muted` and stays enabled where an empty section is a valid choice (a transfer onto an empty farm opens a cycle, C3), disabled where it isn't (Din).
- **Focus and keys:** on open focus moves to the listbox, whose `aria-activedescendant` names the active option (`is-active`, `sunken`): the chosen one, else the first enabled. Its ring is inset (2px `focus`, offset −2px) so the sheet head and the screen edge don't clip it. ↑/↓ move the active option, Enter or a tap picks it and closes the sheet. Tab cycles Închide ↔ listbox only (focus trap); Esc or Închide closes without a change; focus returns to the combobox.
- **Background:** the FormHeader, form and SaveBar behind the scrim are `inert` (out of the tab order and the accessibility tree).
- **Tokens:** existing only (`panel`, `rule`, `sunken`, `ink*`, `focus`, `control-h-touch`, `space-*`); vars from earlier sections (`--ip-hair`, `--ip-fs-figure`, `--ip-focus-w`).

### S10P · SexCountRow "Toate secțiunile" key, 17.11.2026
CSS in additions.css section 34 (both copies, identical). Used by S10 Vânzare (D84); to copy into the system's SexCountRow README with HANDOFF (D32). Decided by default (user, 17.11.2026: "go on with what you recommend"), to review.

- **Anatomy:** `button.ip-btn.ip-btn--ghost.ip-sexrows__allkey[aria-pressed]` right after `.ip-sexrows__head`, before the list: the system Checkbox box `span.ip-check__box[aria-hidden]` (Lucide check when pressed, Lucide minus when mixed, empty when off) + "Toate secțiunile" + hidden " din Hîrbovăț". Phone `ip-btn--touch` (48), desktop 32. The key pulls left by its padding (`space-3`) and its hair border (FIN13) so its box sits in the rows' Toate column; a `rule` splits it from the first row (`space-4` above the row).
- **Behaviour:** `aria-pressed` = "true" when every row has Toate on, "mixed" when some do, "false" when none. From "false"/"mixed" it turns Toate on in every row (counts = all alive on the chosen day, inputs read-only as row Toate); from "true" it turns every row off and the rows get their drafts back. A row's own Toate still works and updates the key. Only on Vânzare (not Transfer `--pick`, not placement); hidden on the saved, read-only form.
- **Tokens:** existing only (`control`, `rule`, `space-*`, `--ip-hair`); no new values.

### S2P · Acasă states: AlertRow narrow, BottomTabBar key off, phone Dialog, skeletons, 17.11.2026
CSS in additions.css section 35 (both copies, identical), moved out of the S2 screen CSS. Used by S2 Acasă; to copy into the system's EventRow (AlertRow), BottomTabBar, Dialog and Skeleton READMEs with HANDOFF (D32). Decided by default (user, 17.11.2026: "go on with what you recommend"), to review.

- **AlertRow · narrow list** (`div.ip-list.ip-list--alerts`, opt-in size container): under 480px the text takes the first line after the tile; the action link (48px target, under the text) and the status Lamp (right) share a second line (D86). From 480px one line as section 4. Screens break the meta only at its `·` separators.
  Title figures keep their unit: `.ip-ev__main > span.ip-num` and `.ip-sex` are `nowrap` (fix 1, as the S4 shell; figure + unit in one span).
- **BottomTabBar:** items fill the bar height (≥ 44px target, content centred), link items not underlined. Disabled key: `a.ip-tabbar__item[role=link][aria-disabled=true][tabindex=0]`, no href, `aria-describedby` = the reason shown on the page (EntryKeys note or EmptyState); label `ink-disabled`, centre key `sunken` + `ink-disabled`, no `shadow-key`; keeps focus.
- **Dialog · phone:** `div.ip-scrim.ip-scrim--phone` fills the positioned phone screen (tab bar included, page under it `inert`), Dialog centred with `space-4` round it; `p.ip-dialog__note` body line (`ink-secondary`); `div.ip-dialog__keys` keys right-aligned on one line, `--stack` full width under each other, Anulează first (focus starts there), destructive last.
- **Skeletons:** StockReadout `ul.ip-readouts > li > div.ip-readouts__skel[aria-hidden]` (light `panel` box, 5 static `ip-skel` bars: caption, value, two lines, a shorter last); CycleTrack `div.ip-trackskel[aria-hidden]`, one bar per lane then the month axis bar. The block keeps `aria-busy` and the page a hidden `role=status` line.
- **Tokens:** existing only (`panel`, `line`, `sunken`, `ink-*`, `space-*`, `control-h-touch`, `record-tile`, `radius-md`, `--ip-hair`); no new values.

### S14P · DataTable scroll region + Button busy, 17.11.2026
CSS in additions.css sections 36–37 (both copies, identical). Used by S14 Rapoarte · Producție; to copy into the system's DataTable and Button READMEs with HANDOFF (D32). Decided by default (user, 17.11.2026: "go with recommended options"), to review.

- **DataTable · scroll region holds its hidden text** (section 32 amendment, section 36): `.ip-tablescroll{position:relative}`. The region is the containing block of the hidden `.ip-sr` words in head cells (e.g. "găini"/"cocoși" after ♀/♂), so they are clipped by the region instead of widening the page (S14 at 1280 scrolled sideways). No z-index (no stacking context; the sticky first column keeps its own). Replaces the S14 screen rule. S4 Fermă tab boards render identically (1440 and 1280).
- **Button · busy** (section 37): `button.ip-btn[aria-busy="true"]` (spinner `span.ip-spin`, label kept, usually also `aria-disabled="true"` so repeat presses are ignored) keeps its variant's surface, text and edge (primary `control`/`on-control` + `shadow-key`; secondary `raised`/`ink`/`line-strong` + `shadow-key`; ghost transparent/`ink`; destructive `raised`/`fault`) over section 1's aria-disabled grey, with `cursor: progress`; hover/pressed stay off through the aria-disabled guard. Replaces the S14S board rule. LoginCard is unaffected (its busy flag sits on the form).
- **Tokens:** existing only; no new values.

### S8P · Readout live comparison, SegmentedControl fill, wrapping field message, 17.11.2026
CSS in additions.css section 38 (both copies, identical), moved out of the S8 screen CSS (`s8-live`, `s8-seg`, `s8-msg`). Used by S8 Cântărire; to copy into the system's Readout, SegmentedControl and TextField READMEs with HANDOFF (D32). Decided by default (user, 17.11.2026: "go with recommended options"), to review.

- **Readout · live comparison** (`div.ip-readout__live[aria-live=polite][aria-atomic=true]` inside `.ip-readout`, around caption, value and sub-line; lays out as the Readout's own column): one live region per page. The value and the caption's section · sex · SAPT are `aria-hidden`, so it announces the lamp word + sub-line once ("ÎN NORMĂ ghid SAPT 21: 2 525 g · 2 553 g (+1,1%)", "FĂRĂ DATE ÎN GHID SAPT 40"). The lamp has an id (`ro-lamp`) and the field that feeds it is described by "ro-lamp ro-sub"; on a warning by "ro-lamp" + its own warn line ("ABATERE" + the warn line, no doubled comparison). The `dl.ip-readout__kv` stays outside the live region.
- **SegmentedControl · fill** (`.ip-seg--fill`): the track spans its field and the segments share it equally, label centred (phone ♀/♂ choice). Touch height unchanged (`ip-seg--touch`).
- **Field message that may wrap** (`p.ip-err.ip-fieldmsg` / `p.ip-warnline.ip-fieldmsg` under a field): no paragraph margin, icon keeps its size on the first line (as section 12's `.ip-entry__msg`).
- **Tokens:** existing only (`space-*`, `--ip-hair`); no new values. No other board uses these classes, so nothing else re-renders.

### S16 · SidePanel phone full-screen sheet, 17.11.2026
CSS in additions.css section 39 (both copies, identical). Used by S16 Corectează (phone boards); to copy into the system's SidePanel README with HANDOFF (D32). Decided by default (user, 17.11.2026: "go on with what you recommend"), to review.

- **Anatomy:** `aside.ip-sidepanel.ip-sidepanel--sheet[role=dialog][aria-modal=true]` (labelled by its title) filling the phone screen: head (record tile · title 16/22 "Corectează Mortalitate H4.1" + Mono note "HÎRBOVĂȚ · 14.11.2026 · SAPT 40" · touch IconButton "Închide Corectează", Lucide "x") · body (scrolls) · foot (Șterge destructive, then the primary, both touch, sharing the width).
- **Desktop** keeps the system SidePanel beside the page (non-modal, page reflowed, nothing under it, as S9/S10/S13).
- **Over it:** the Șterge Dialog sits in `ip-scrim--phone` (section 35, keys stacked, focus on Anulează); the sheet is `inert` meanwhile. Its Motiv is required and is prefilled from the sheet's Motiv when one was typed there (S16P), otherwise empty; while empty its Șterge key is `disabled` with `aria-describedby` = the Motiv hint ("Obligatoriu, cel mult 200 de caractere · 0/200 · scrie motivul ca să poți șterge") (spec E1; fix round 1).
- **Motiv soft limit (S16P, D102):** on the sheet / panel and in the Dialog the Motiv `textarea` has no `maxlength` (a paste is never cut); the hint keeps the counter ("… · 59/200 · apare în Istoric modificări"). Over 200 the field turns error (`is-error`, `aria-invalid`) with "Motivul are 212 caractere; cel mult 200. Scurtează-l cu 12 ca să poți salva." (Dialog: "… ca să poți șterge") and Salvează corecția / Șterge are `disabled` with `aria-describedby` = that error.
- **Focus (S16P):** on open → the first CountInput (Mortalitate ♀; desktop board `S16-corecteaza--desktop`); while the record loads the keys are `disabled` with `aria-describedby` = the visible status line ("Se încarcă rândul din 14.11.2026 · Șterge și Salvează corecția se activează după încărcare.") and focus stays on Închide. On close (Închide, Anulează, Esc) → back to the Corectează key that opened it. Dialog: on open → Anulează; on Anulează / Esc → back to the Șterge key of the sheet. After a failed save → Reîncearcă in the fault Banner (`--eroare-la-salvare`). After deletion the sheet / panel closes and the deleted row's key is gone → focus goes to Corectează of the row now in its place (the next row; the previous one if it was the last; the list heading if the list is empty) (`--desktop-dupa-stergere`: Mortalitate H4.2 · 14.11.2026); the Toast (`role=status`) announces it and keeps Anulează in the tab order.
- **Head note:** the Mono context line is in capitals like the S9/S10/S13 panel heads: "HÎRBOVĂȚ · 14.11.2026 · SAPT 40", "TELIȚA → HÎRBOVĂȚ · 28.02.2026".
- **Tokens:** existing only (`panel`, `line`, `control-h-touch`, `space-*`); no new values. No other board uses the class.

### S17S · RecordChip touch, 17.11.2026
CSS in additions.css section 40 (both copies, identical). Used by the S17 phone filter BottomSheet (`S17-istoric-modificari--telefon-filtre`); to copy into the system's RecordChip README with HANDOFF (D32). Decided by default (user, 17.11.2026: "go on with what you recommend"), to review.

- **Rule:** on a touch screen a filter chip is a phone control, so it gets the touch height: `button.ip-chip.ip-rec.ip-chip--touch[role=checkbox]` = `control-h-touch` (48) tall, side padding `space-3` / `space-4`, icon gap `space-2`; the split chip (`ip-chip--split`) keeps a `space-2` left edge for its 20px tile. Colours, on/off, hover/pressed/disabled/focus and the 13/16 label are the chip's own (sections 1, 15). Chips in a group sit `space-2` apart, wrapping: the screen sets that gap on its chip group (`.s4-chips`, the same group as the desktop FilterBar chips), not §40, which sizes only the chip (S17P; there is no system chip-group class). Rejected: the 28px chip with an invisible 48px hit area (needs 20px row gaps, and the face no longer shows the target); 28px as on desktop (under the 44px touch floor, craft §5).
- **Where:** inside the phone FilterBar BottomSheet, under a visible label "Tipuri de înregistrări" (`role=group` + `aria-labelledby`). Desktop FilterBars keep the 28px chip.
- **Tokens:** existing only (`control-h-touch`, `space-*`); no new values. No other board uses the class.

### S19P · SegmentedControl error, DataTable skeleton row, 17.11.2026
CSS in additions.css sections 41–42 (both copies, identical). Used by S19 Ferme și secțiuni (`--nume-gol`, `--se-incarca`); to copy into the system's SegmentedControl, DataTable and Skeleton READMEs with HANDOFF (D32). Decided by: Claude (recommended, user to review).

- **SegmentedControl · error** (section 41): a required choice left unset after save, `span.ip-seg[role=radiogroup][aria-invalid=true][aria-describedby=<error id>]` (demo `.is-error`): the track takes the TextField error edge, a hair `fault` border, its padding cut by the same hair, so the control keeps its size and the segments stay put. The error line under it is the field message (`p.ip-err.ip-fieldmsg`, section 38, alert icon): "Alege tipul fermei: Creștere sau Ouat. Lista de ghiduri depinde de el." Segments, selected look and focus are unchanged. Replaces D111's "error line only".
- **DataTable · skeleton row** (section 42): while a DataTable loads, its real head row, then two `tr.ip-table__skel` rows with one static `ip-skel` bar per cell (`--w1…w3` by the column's usual content), centred in the row; `.ip-skel--end` puts a bar at the right edge (numbers, the action column). Region `aria-hidden`, block `aria-busy`, a hidden `role=status` line says what loads. Replaces the S19S screen rule `.s19-skt`.
- **SectionTree edit row error** (SYS10): the error line carries the alert icon like every field message (`p.ip-err.ip-fieldmsg`, icon + span). Drawn screen-side in S19 (s19s_build); SYS10's helper and preview still draw text only (FIN).
- **Not moved (FIN):** the desktop Toast placement (`.s19-toastwrap`, bottom right, `space-8` in; same as S18S) and the Banner's trailing action key (`.s19-bkey`) stay screen CSS, to move together with the Toast-undo and Banner-action layouts they wrap (`s7s-toast`, `s7s-banner`, sstate_css.py).
- **Tokens:** existing only (`fault`, `--ip-hair`, `space-*`); no new values. No other board uses these classes or a SegmentedControl with `aria-invalid`, so nothing else re-renders.

### S20P · FileDrop paste target, 18.11.2026
CSS in additions.css section 43 (both copies, identical); SYS11 amendments in section 29 (see SYS11 above). Used by S20 Ghiduri (Importă ghid, format greșit, necitibil, nume gol/folosit) and the SYS11 preview; to copy into the system's FileDrop README with HANDOFF (D32). Decided by: Claude (recommended, user to review), D115.

- **FileDrop · paste target** (section 43): where a table can be pasted, the FileDrop itself takes focus so Ctrl+V lands in it: `div.ip-drop.ip-drop--paste[tabindex=0][role=group][aria-label="Lipește tabelul aici (Ctrl+V) sau alege un fișier .xlsx sau .csv"]`, after a wrong file also `aria-describedby` = the file fault Banner (`role=alert`). Focus = the system ring (`focus`, 2px, 2px offset) round the dashed box (`:focus-visible`, demo `.ip-focus`); look, icon, sentence and column note are the FileDrop's own. "alege un fișier" inside is a real `button.ip-link` drawn as the link, inline in its sentence (rule moved from section 29's `.ip-gimp` scope to `.ip-drop`). Markup from SYS11's `DROP` helper; S20 no longer adds the attributes itself.
- **Tokens:** existing only (`focus`, `--ip-focus-w`, `--ip-focus-offset`); no new values. No board outside S20 uses `ip-drop` or `ip-gimp`, so nothing else re-renders.
- **Not moved (FIN):** `.s20-toastwrap` / `.s20-bkey` equal `.s19-toastwrap` / `.s19-bkey`; they move with `s7s-toast` / `s7s-banner` (sstate_css.py).

### S21P · Checkbox disabled, error and form field, 18.11.2026
CSS in additions.css section 44 (both copies, identical). Used by S21 Tipuri de hrană (the Activ field on every type panel; the error on `--retragere-refuzata`); to copy into the system's Checkbox README with HANDOFF (D32). Decided by: Claude (recommended, user to review), D117.

- **Checkbox · disabled** (section 44): only where the box truly can't change, with the reason under it as the hint (`aria-describedby`): `input[disabled]`, or `input[aria-disabled=true]` kept focusable while a script ignores clicks (demo `.is-disabled`). The craft rule, `sunken` + `ink-disabled`: box `sunken` with a `line` edge, a ticked box keeps its check in `ink-disabled` (no `control` fill), label `ink-disabled`, cursor not-allowed. Replaces D116's "the system Checkbox has no disabled look". No S21 board draws it: every Activ box there can change.
- **Checkbox · error** (section 44): a choice the save refuses, `input[aria-invalid=true][aria-describedby=<error id>]` (demo `.is-error`): the box border turns `fault` at its own hair width (the TextField error edge); fill, check and focus ring unchanged; the error line under it is the field message (`p.ip-err.ip-fieldmsg`, section 38) naming the fix: "Breeder 2 nu se poate retrage: Hîrbovăț mai are 28 252 kg în stoc. Bifează Activ la loc; îl retragi când stocul ajunge la 0 kg." A refused box stays live (the user ticks it back); the save key is `aria-disabled` and described by the same line. A ticked box in error keeps its `control` fill, so the fault edge is ~1.4:1 there; no board draws it, and a ticked refusal needs an outer fault ring when a screen first needs it (D117).
- **Checkbox as a form field** (section 44): `div.ip-field > label.ip-check` + hint or error line. The label keeps its 44px hit area but takes `-space-3` above and below, so its 20px text line sits in the form rhythm (the form gap above, `space-1` to the hint, as a TextField's label, input and hint), and it hugs its text (`align-self: start`); `position: relative` paints it over the hint below, so all 44px stay the target (S21P fix round 1). Replaces the S21 screen rules `.s21-act`. Checkboxes outside an `ip-field` (SexCountRow Toate, EntryCard foot) are unchanged.
- **Tokens:** existing only (`sunken`, `line`, `ink-disabled`, `fault`, `space-3`); no new values. No board outside S21 has a Checkbox inside an `ip-field`, disabled or `aria-invalid` (S6, S9, S10 use it in the EntryCard foot and SexCountRow), so nothing else re-renders.
- **Not moved (FIN):** `.s21-toastwrap` / `.s21-bkey` equal `.s19-toastwrap` / `.s19-bkey`; they move with `s7s-toast` / `s7s-banner` (sstate_css.py).

### S22P · DataTable current row; SYS12 fixes, 18.11.2026
CSS in additions.css section 45 (both copies, identical); SYS12 amendments in `sys12_build.py` (above). Used by S22 Utilizatori (the user panel boards); to copy into the system's DataTable README with HANDOFF (D32). Decided by: Claude (recommended, user to review), D120.

- **DataTable · current row** (section 45): the row whose record is open beside the list in a non-modal SidePanel. The row's own key or link carries `aria-current="true"` (S22: the Acțiuni IconButton; demo `tr.is-current`); the row takes `sunken` with a hair `line-strong` frame drawn inside it (outline, offset −hair), so it reads apart from hover (`sunken`, no frame). Steel only, no lamp: no colour bar, no green; text, lamps and keys unchanged (`ink-muted` 4.7:1, the ok word 4.6:1 on `sunken`). The fill sits on the `tr` (cells transparent), so a positioned action cell doesn't paint over the frame.
- **One primary per view with a panel open:** the list head's primary ("Adaugă utilizator") is drawn `secondary` while any SidePanel is open, so the panel's Salvează / Adaugă utilizatorul is the one graphite key; while its own add panel is open it is also pressed and `aria-expanded="true"`. Screen-side in S22 (Button variants unchanged).
- **SYS12:** `dialog()` → `role="alertdialog"`, always with `aria-describedby` (the confirm sentence, the fault Banner, or for the password step the once Banner `#…-once`); copy failed leaves the status region empty and the error line is a field message with the alert icon; preview load failed head note "– lista nu s-a încărcat"; the row menu opens by room (`--up` only without room below), preview blocks keep room under their open menus.
- **Tokens:** existing only (`sunken`, `line-strong`, `--ip-hair`); no new values. No board outside S22 has a table row with `aria-current` or `is-current`, so nothing else re-renders.
- **Not moved (FIN):** S19 Ferme și secțiuni and S21 Tipuri de hrană already show one primary with a panel open and mark the edited row by its pressed, expanded Editează key (ghost); S22 rows have no Editează key (it sits in the row menu), hence §45. FIN: add §45 to the S19/S21 edited rows too (key kept pressed), so every list marks the open record the same way.

### S1P · Button spinner, LoginCard Banner icon and page position, 18.11.2026
CSS in additions.css section 46 (both copies, identical). Used by S1 Autentificare (`--se-intra`, the four Banner states, the two desktops); to copy into the system's Button and LoginCard READMEs with HANDOFF (D32). Decided by: Claude (recommended, user to review), D121.

- **Button · loading spinner** (section 46, system-wide): `span.ip-spin` is a quarter arc in the label colour on a full quiet track, the arc at the top right (track `line` on light keys, `ink-muted` on the graphite primary), same 14px box and 2px stroke, static (no loop, craft §4). bundle.css drew a 3/4 ring open on the right that read as the letter "C" before the label on a still board. The label stays (Button › Loading; rejected: "Se intră…" in place of "Intră", against "the label stays"). Other boards that draw a spinner: S14 `S14-rapoarte-productie--observator-se-exporta` (ghost Exportă .xlsx), plus the Button and LoginCard previews; their HTML is unchanged.
- **LoginCard · Banner icon** (section 46): `.ip-login__msg > .ip-icon` dropped by `--ip-hair` so its 16px icon centred on the first 18px text line (as section 38's field message). Moved to the system-wide Banner rule, section 47 (FIN01).
- **Login page · desktop position** (section 31 amendment): the card's top edge at `space-12` × 4 (192), so the card rests on the upper third of a 1024 page; still top-anchored (a Banner grows the card downward). Phone unchanged. Ground stays `ground`; no tagline, footer or version line (craft §4, D121).
- **Fix round 1:** signing in, Intră itself carries `aria-busy="true" aria-disabled="true"` (Button › Loading, as the S14 busy key; section 37 keeps it graphite, look unchanged), set by SYS13 `card(busy=True)`; the form keeps its `aria-busy`. The LoginCard preview's desktop band grows to `space-12` × 16 (768; was × 12, 576) so the card sits 192 from its top with room below, and its README says `space-12` × 4 (192).
- **Tokens:** existing only (`line`, `ink-muted`, `--ip-hair`, `space-12`); no new values.

### FIN01 · Banner icon on its first line, 18.11.2026
CSS in additions.css section 47 (both copies, identical); replaces section 46's LoginCard-only rule. Used by every Banner with a 16px icon (53 boards on 19 screens; Banner, EffectPreview, FeedTotals, GuideImportPreview, LoginCard, OneTimePassword, UserRow previews); Banner README line added. Decided by: FIN review (FIN.md), Claude (recommended, user to review).

- **Banner · icon** (section 47, system-wide): `.ip-banner > .ip-icon--16{margin-block:var(--ip-hair)}`. bundle.css set the 16px icon at the top of the first 18px line, 1px high; a hair above and below centres it on that line. Both sides, so a Banner that centres its row (`s4s-banner`, align-items center beside a key) keeps its icon where it was, and the 18px margin box never grows a Banner. `s7s-banner`'s own `margin-top: hair` gives the same value (no double shift). 18px icons (SectionTree) are already on their line and untouched.
- **UserTable · load failed** (section 47): `.ip-users__bnrow` (sentence + Reîncearcă) aligns to its top and the sentence takes `(control-h − 1lh) / 2` above and below, so its first line sits on the key's centre line beside the key and at the top when the key wraps under it; the icon drops by the same amount plus the hair. Was 8px above the sentence's line; height unchanged at 1440 (+14px when the key wraps under the sentence, 390). (Superseded by D134 (FIN13): the UserRow preview uses the §49 Banner; the `.ip-users__bnrow` rules are removed.)
- **Measured** (icon centre − first-line centre, 1440 and 390): bare Banner −1 → 0 (21 boards), FeedTotals −1 → 0, UserTable −8 → 0, LoginCard 0 → 0, `s7s-banner` 0 → 0, `s4s-banner` unchanged (0 / 0.5 on one line).
- **Not moved (FIN):** `s4s-banner` centres the icon on the whole Banner, so where its text wraps (S14, S17, S18 `--eroare-la-incarcare` at 1440: two lines) the icon sits 9px under the first line. Screen CSS (sstate/s4s); move to the first line there if wanted.
- **Tokens:** existing only (`--ip-hair`, `control-h`); no new values. No board HTML changes.

### FIN02 · Toast placement and undo layout, 18.11.2026
CSS in additions.css section 48 (both copies, identical). Used by every Toast (18 boards on 14 screens); Toast README line added. Decided by: FIN review (FIN.md), Claude (recommended, user to review), D122.

- **Toast · placement** (section 48, system-wide): `div.ip-toastwrap > div.ip-toast[role=status]`, absolute in the screen's positioned area, `z-index` under a Dialog scrim. Phone (default): `left/right/bottom: space-4` inside the view that ends above the SaveBar or BottomTabBar (unchanged from the `s7s-toastwrap` every phone used); in `.ip-sidepanel--sheet` it sits `space-4` above the sheet foot (touch keys + `space-3` padding × 2 + hair). Desktop: `ip-toastwrap--page`, bottom right of the page at `space-8`, `max-width: min(panel-w, 100% − 2 × space-8)`: the right column where the SidePanel opens; longer text wraps. Was bottom-left at `space-6` beside the Sidebar (S12), right at 24 (S16), right at 32 without a cap (S18–S22).
- **Toast with Anulează** (section 48): `.ip-toast__txt` (flex 1, wraps, `margin: 0` so a `p` works) and `.ip-toast__undo` (on-control link, weight 500, 48px touch target via `min-height: control-h-touch` with a negative block margin back to the 20px line, centred); a Toast holding `__txt` aligns to its top and its icon drops a hair onto the first line. Same values as the `s7s-toast` / `s7s-undo` screen copies they replace.
- **S12 "Prețuri salvate"** gets Anulează (owner, same day), as every save Toast (D111/D116/D118).
- **Replaced screen CSS:** `s7s-toastwrap`, `s7s-toast`, `s7s-toast__txt`, `s7s-undo` (sstate_css.py TOAST, removed), `s12-toastwrap`, `s16s-toastwrap`, `s16s-toastwrap--desk`, `s18s-toastwrap`, `s19-toastwrap`, `s20-toastwrap`, `s21-toastwrap`, `s22-toastwrap`. The "Not moved (FIN)" Toast notes under S19P, S20P and S21P are closed by this entry (the `-bkey` / `s7s-banner` parts stay open).
- **Tokens:** existing only (`space-4`, `space-8`, `space-3`, `space-2`, `space-5`, `panel-w`, `control-h-touch`, `--ip-hair`); no new values.

### FIN03 · Banner with action, 18.11.2026
CSS in additions.css section 49 (both copies, identical); section 46's header no longer names the LoginCard Banner icon (moved to section 47 by FIN01). Used by every Banner with a key: 25 boards on 18 screens (S2, S4, S6–S10, S12–S22); Banner README line added. Decided by: FIN review (FIN.md), Claude (recommended, user to review), D123.

- **Banner · with action** (section 49, system-wide): `.ip-banner--act` wraps (`flex-wrap`), aligns icon and body to the top and puts the key (`.ip-banner__key`, `flex: none`, `align-self: center`) at the end. `.ip-banner__body`: `flex: 1 1 calc(space-12 × 6)`, `min-width: 0`, a column of `p` (margin 0) space-2 apart. The icon sits in the start padding (padding-left space-3 + space-4 + space-2, icon margin-left −(space-4 + space-2)), so the text starts where it did and a wrapped key's line starts at the text. Row gap space-3 (text to wrapped key, as the `s7s-banner` it replaces). A link key keeps its 24px target (D46) with a block margin of (space-6 − 1lh)/2, so a one-line Banner stays 18px + padding.
- **Replaced screen CSS and markup:** `s4s-banner` (icon centred on the whole Banner: S4/S14 fără ghid, S14/S17/S18 eroare la încărcare, S17 din „modificat de”), `s7s-banner` + `s7s-banner__body` with a key (phone and SidePanel save/load errors: S2, S6–S10, S13, S15, S16, S18–S22), `s19/s20/s21-bkey` (S19–S21 load errors), bare Banners with a key (S12 eroare; S22 eroare la încărcare via `.ip-users__bnrow`). Removed: `.s4s-banner` rules (s4s_build CSS_S4/CSS_STATES), `.s7s-banner__body > .ip-btn` (sstate_css; `s7s-banner` stays for keyless Banners: refuzat, zi trecută, deja introdus, S19 compartiment), `.s19/.s20/.s21-bkey`. Section 47's `.ip-users__bnrow` rules now serve only the SYS12 UserRow preview. (Superseded by D134 (FIN13): the UserRow preview uses the §49 Banner; the `.ip-users__bnrow` rules are removed.) Closes FIN01's "Not moved" `s4s-banner` note and the `-bkey` / `s7s-banner` parts of the S19P–S21P notes.
- **S22 eroare la încărcare:** `role="alert"` and the S19–S21 copy: "Utilizatorii nu s-au încărcat." / "Verifică conexiunea la internet, apoi reîncearcă. Poți adăuga un utilizator nou și acum." (was one line with a bold lead, no role). The list's status line (`ip-sr`, role status) stays but empty: it repeated the first sentence the alert now announces.
- **Measured** (fallback fonts; 390 / 1280 / 1440): icon centre − first-line centre 0 on all 25 boards (S14/S17/S18 load errors were 9–9.5); key centre − Banner centre 0 where it sits beside the text; wrapped keys start at the text (0) 12px under it; no overflow. Phone and SidePanel Banners unchanged in size (shots byte-identical); desktop two-paragraph Banners +8 (paragraph gap; 54 → 62, S22 50 → 62); S4/S14 fără ghid 42 → 36; S20 eroare la salvare (864 wide) puts Reîncearcă beside the text, 98 → 54. Board heights and canvas.json unchanged.
- **Fix 1 · one-line modifier** `ip-banner--1l` (section 49): for text that stays one line beside the key at every desktop width, the body is `control-h` tall with its line centred and the icon drops by (control-h − 1lh)/2 + hair (section 47's UserTable value; that rule stays for the SYS12 UserRow preview's different markup). Used on S17 din „modificat de” only (the other one-line action Banners, S4/S14 fără ghid, have a link key and already measure 0). Banner 50 high as at HEAD; icon − line 0.5 (HEAD 0.5), key − Banner 0; shot S17-6 byte-identical to HEAD. CSS cannot tell one line from two, so the modifier is set by the screen (D123).
- **Tokens:** existing only (`space-2/3/4/6/12`, `--ip-hair` via section 47); no new values.

### FIN04 · One Dialog scrim, 18.11.2026
CSS in additions.css section 50 (both copies, identical). Used by every desktop Dialog: 13 boards on 7 screens (S4, S16, S18–S22); S10 confirmare (phone) moves to section 35. Dialog README line added. Decided by: FIN review (FIN.md), Claude (recommended, user to review), D124.

- **Dialog · page scrim** (section 50, system-wide): `div.ip-scrim.ip-scrim--page` > `div.ip-dialog[role=alertdialog][aria-modal=true]`, the desktop twin of section 35's `ip-scrim--phone`: `position:absolute; inset:0` in the positioned page (the board's `.ip-root`), `z-index` 5 (over any open SidePanel and the Toast of section 48), no radius, Dialog centred both ways with bundle.css's space-6 round it (one grid row). A board taller than its viewport sets `--board-vp-h` (its viewport height) on its root: the scrim covers the whole board and the Dialog centres in that first viewport, as at scroll 0 (fix round 1). The page shell and the SidePanel under it are `inert` (out of the tab order and the accessibility tree; S4/S16 add `aria-hidden`); focus starts on Anulează (S22's OneTimePassword Dialogs keep their SYS12 focus: Copiază parola, the password field, Reîncearcă). Body line `p.ip-dialog__note`, keys `div.ip-dialog__keys` (section 35). In the app the scrim is the viewport.
- **Moved:** S4 ciclu nou and S16 desktop Șterge sat 192px under the top (`s4s-scrim`, padding-top space-12 × 4); now centred (S4's 1874-tall board: in its first 1024px via `--board-vp-h: 1024px` on the root, Dialog top 270, as the app's fixed scrim at scroll 0). S18–S22 were already centred by the same values under screen names. S10 confirmare (phone 390) used the desktop rule at 192 and the `s10-dlgfoot` key rule: now `ip-scrim--phone` with `ip-dialog__keys--stack`, as S2/S15/S16 (Dialog 342 → 358 wide, centred).
- **Replaced screen CSS and markup:** `s4s-scrim`, `s4s-dialog`, `s4s-dlgnote` (→ `ip-dialog__note`), `ip-row s4s-dlgfoot` (→ `ip-dialog__keys`) in s4s_build CSS_STATES (also picked by s10/s16s), `s10-dlgfoot`, `s18s-scrim`, `s19-scrim` (S19–S22). `s9-modal` is the phone BottomSheet anchor (`position:relative`), not a Dialog scrim: unchanged. Component previews (Dialog card, SYS10 SectionTree, SYS12) keep the bare `ip-scrim` card.
- **Measured** (fallback fonts; 1440, 1280, 390): Dialog centre − scrim centre 0 / 0 on all 14 Dialog boards; scrim = page box, over the SidePanel (hit test in the scrim) on S16, S19, S21 and S22 parola resetată / copiere eșuată; no focusable element outside the scrim that is not inert; every Dialog `alertdialog` + `aria-modal`, autofocus Anulează (S22 password Dialogs as SYS12).
- **Tokens:** existing only (bundle.css space-6, `scrim`); no new values.

### FIN05 · One current-row mark, 18.11.2026
CSS in additions.css section 51 (both copies, identical), an amendment of section 45. Used by S16 Corectează (desktop Jurnal, 4 boards), S19 Ferme și secțiuni (12 panel boards), S21 Tipuri de hrană (5); S22 unchanged (its rows match through `:has()`). DataTable and EventRow lines above and READMEs updated. Decided by: Claude (recommended, user to review), D125.

- **EventRow · current row** (section 51): `ul.ip-list > li.ip-ev.is-current`, its own key or link `aria-current="true"` (`:has()` fallback as section 45): `sunken` + hair `line-strong` frame inside (outline, offset −hair). Same look, markup and rule as the DataTable current row: `tr.is-current` + `aria-current` on the row's key.
- **Key:** the Editează / Corectează key that opened the panel keeps `aria-expanded="true"` + `aria-controls` and takes `aria-current="true"`, but is not drawn pressed (the row carries the mark). The list head's add key stays pressed while its own add panel is open (no row yet).
- **Inside the marked row** (tables and lists): a neutral Tag on `rule`; ghost keys and IconButtons hover `rule`, pressed `line` (hover `sunken` would vanish on the fill). Row header `th[scope=row]` takes `sunken`; in `ip-tablescroll` the sticky first cell draws its part of the frame (same outline, `clip-path` at its right edge), so the frame is unbroken and the cell stays opaque while scrolling.
- **Measured:** frame 3.9:1 against neighbouring raised rows; text on `sunken` ink-muted 4.7, ♀ 4.8, ♂ 5.0, link 5.5, ok word 4.6:1; Tag text on `rule` 6.0:1; focus ring 4.6:1. Hover: `sunken`, no frame; focus: the ring on the key.
- **Tokens:** existing only (`sunken`, `rule`, `line`, `line-strong`, `--ip-hair`); no new values.

### FIN06 · Breadcrumb hit area, 18.11.2026
CSS in additions.css section 52 (both copies, identical). Used by every board with a breadcrumb link: 139 desktop boards (S4, S9, S10–S22; S3's 4 crumbs have no link) take it from CSS alone; the 9 phone boards (S14 telefon, telefon-filtre; S17 telefon, telefon-filtre; S18, S19, S20, S21, S22 telefon) switch to the touch class. To copy into the system's PageHeader README with HANDOFF (D32). Decided by: Claude (recommended, user to review), D126.

- **Desktop** (section 52): `.ip-crumbs a` is an inline-flex box `space-6` (24) tall, padded (`space-6` − 1lh)/2 = 5px at the sides, with the same negative margins, so the 14px crumb line, the h1 and the meta line stay where they were (element rects identical to HEAD on all 148 boards with a crumb link). `position: relative` paints it over the h1's top margin (the box overhangs it by 3px), so the whole box takes the click. Focus (fix round 1): the link's outline is dropped and `a::after` draws the ring (border `--ip-focus-w` `focus`, `radius-xs`, `pointer-events: none`), inset so it sits 2px off the word's 14px line at the sides and top and flush with the line below: its lower edge stops at the h1's top edge. The old 2px-offset outline overlapped the Î circumflex of the 8 S20 "Înlocuiește …" titles by 2px (already at HEAD; §52's `position: relative` would have painted it over the glyph). Ring 20px tall (was 22), widths unchanged; 4.65:1 or more on every steel surface.
- **Phone** (section 52): `nav.ip-crumbs.ip-crumbs--touch` > `a` > `span.ip-crumbs__txt`: the link reaches `space-4` above and below and `space-1` at the sides (46 tall, word + 8 wide), negative margins keep the line; the word's span takes the ring (2px `focus`, 2px offset, `radius-xs`), the link draws none, so the ring neither crosses the title nor clips at the screen top. Replaces the screen rules `.s14-phead .ip-crumbs a`, `.s14-crumb` (S14S/S14P) and S19P's 50×46 inline padding (also used by S20–S22). S17/S18 "Setări" grew from 42×46 to 50×46.
- **Measured** (fallback fonts, hit test by elementFromPoint): desktop 79×24 "Seria 2026", 74×24 "Rapoarte", 53×24 "Hrană", 51×24 "Acasă", 52×24 "Setări", 60×24 "Ghiduri", 134/106×24 Ross 308 guides; phone 72×46 "Rapoarte", 50×46 "Setări". No pixel change at rest on any board. Focused (139 desktop boards, 169 links): the ring covers no ink of the word (Ș comma in "ROSS 308 CREȘTERE" included) or the title; ring foot to the h1's first ink 0px on the 8 "Înlocuiește" boards, 1px or more elsewhere (5px on most).
- **Tokens:** existing only (`space-1`, `space-4`, `space-6`, `focus`, `radius-xs`, `--ip-focus-w`, `--ip-focus-offset`); no new values.

### FIN07 · S22 list page; UserRow skeleton rows, 18.11.2026
SYS12 amendment in `sys12_build.py` (`SK`) and additions.css section 30 (both copies, identical). Used by S22 Utilizatori (`--se-incarca`) and the UserRow preview. Decided by: Claude (recommended, user to review), D127.

- **UserRow · loading** (section 30 → 42): the skeleton rows are the DataTable skeleton row, two `tr.ip-table__skel` (aria-hidden), one `ip-skel` bar per cell, the Acțiuni cell `td.ip-users__act` with a `--w3 --end` bar; the SYS12-only rule `.ip-users__skel .ip-skel{display:block}` is removed (section 42 already sets it, plus `max-width` and the middle alignment). Preview: 4 → 2 rows, card height 2432 → 2368 (measured 2366). S22 `--se-incarca` follows S19S/S21: hidden `role=status` line, block `aria-busy`, box `aria-hidden`, no count in the list head.
- **UserRow on a page** (S22, screen-side): the list sits in the page's `section.ip-block` with the `div.ip-sechead` (FIN15/FIN15B names) (`h2.ip-t-heading` "Utilizatori" + the count note `#…-sum` on the right), and "Adaugă utilizator" is the PageHeader's secondary key (pressed + `aria-expanded` + `aria-controls` while its own add panel is open), as S18/S19/S21; the panel keys are the view's one primary. The standalone `section.ip-users` + `div.ip-list__head` (caption, counts, key) stays the component preview's shape. README updated.
- **Add keys are verbs** (D127, craft §6): "Adaugă serie" (S18), "Adaugă fermă" (S19), "Adaugă tip" (S21), "Adaugă utilizator" (S22); panel titles stay nouns. Button markup unchanged.
- **Tokens:** existing only; no new values. No board outside S22 uses `ip-users__skel`, so nothing else re-renders from the CSS change.

### FIN09 · Date, time, range and number format, 18.11.2026
Copy rule for every board, component preview and generator (craft §6 › Copy, which it extends). No CSS. Decided by: Claude (recommended, user to review), D129.

- **Date + time:** `dd.mm.yyyy, hh:mm` (comma, as ChangeRow and UserRow already wrote it): EventRow meta "16.11.2026, 08:34 · SAPT 40 · Dr. Ceban", "cântărit 16.11.2026, 10:10", "modificat de Ion Rusu · 15.11.2026, 10:12". The `·` stays the separator between fields; it never joins a date to its time. Time alone (today) stays "hh:mm" ("salvat la 08:34", "până la 23:59"). Text the user typed (a Motiv sample) is left as typed.
- **Range:** one form, `start–end` with an en dash and no spaces, for figures, weeks, dates and months: "SAPT 3–40", "1–15 din 1 098", "2 600–5 000 g", "10.11–16.11.2026", "07.10.2025–28.02.2026", "10.2025–12.2026", "10.02.2026–azi". The end carries the full date; the start drops only the year when it is the same ("10.02–16.11.2026", not "10.02.2026 – 16.11.2026"), never the month ("10.11–16.11.2026", not "10–16.11.2026"). Two exceptions drop the year at the end too (D129 amendment): an in-context day window of the current week, whose year the screen already gives (S13 effect "la media 07.11–09.11", StockReadout "medie/zi 13.11–15.11"); and a table row whose column or head gives the series or period, which drops both years even across a year boundary (S14 "SAPT 40 10.11–16.11", S14 --seria-2025 "SAPT 43 26.12–01.01"). Prose may say "de la … până la …".
- **Thousands:** every grouped figure uses U+202F (narrow no-break space), from 1 000 up, in visible text, aria-labels and hidden captions alike ("1–15 din 1 098", "29 825 păsări", "24 000 kg"). Generators format through the shared `n()` (s4s_build/s7_build, sys11 `NNBSP`); a literal figure in a string is written with `n()` too, never with a typed space. U+00A0 only ties a unit, a sign or a `·` to its figure; it is never the thousands separator.

### FIN10 · One Select widget, 18.11.2026
CSS in additions.css section 53 (both copies, identical). Replaces four copies of one widget: the SeriesSelector `ip-series__btn/__menu/__opt` (Sidebar, 155 boards), the SexCountRow SectionSelect `ip-sexrow__btn/__menu` (S9, S18), S19's Ghid `s19-combo/s19-pop/s19-anchor` and S4's CycleSelector `s4-cycsel` + `s4s-anchor/-cycmenu/-list/-mitem/-gap` (screen CSS removed). The old §6 and §18 select rules served only the SeriesSelector and SexCountRow previews (sys07_build); FIN13 moved both previews to §53 and removed those rules. Decided by: Claude (recommended, user to review), D130.

- **Anatomy:** `div.ip-select` (anchor; the Sidebar adds `ip-series` for its slot) > `div.ip-input.ip-select__btn[role=combobox][tabindex=0]` (`aria-haspopup=listbox`, `aria-expanded`, `aria-controls`, `aria-labelledby` the visible label's id, `aria-describedby` hint or error, `aria-activedescendant` while open) > `span.ip-select__val` (+ Lamp + word) + Lucide chevron (chevron-up while open), then `ul.ip-menu.ip-select__menu[role=listbox][tabindex=-1]` (`hidden` while closed): `li.ip-menu__group[role=presentation]` captions and `li.ip-menu__item.ip-select__opt[role=option][aria-selected]` = `span.ip-select__chk` (aria-hidden; Lucide check on the chosen option, every option keeps the 16px slot) + `span.ip-select__txt` > `span.ip-select__top` (name, then its state at the end: Lamp + word, SexCount pair or reason `ip-note`) + optional `ip-num` line (dates, week). A popup with actions (CycleSelector) = `div.ip-menu.ip-select__menu` > `ul.ip-select__list[role=listbox]` + caption + `button.ip-menu__item.ip-select__act` (leading icon or empty `ip-select__chk`).
- **Keys (APG select-only):** focus stays on the combobox; ↓/Enter/Space/Alt+↓ open; ↑↓ Home End PgUp PgDn and type-ahead move the active option (`is-active`, `sunken`); Enter/Space/Alt+↑ choose and close; Esc closes unchanged; Tab chooses and moves on. `aria-disabled` options (`ink-disabled`, reason word kept) are never chosen.
- **Phone:** the same combobox with `ip-input--touch` opens the section 33 BottomSheet: `aria-haspopup="dialog"`, `aria-controls` only while open, no hidden listbox in the page (now also S18's placement rows and the S10/S18 farm fields). Select inside a sheet: the outer sheet goes inert, Esc closes only the top sheet, focus returns to the field (pf-farm; S14 filter sheet, FIN12, D133).
- **States:** placeholder `is-empty` (`ink-muted`); open = `ink` edge; disabled = `aria-disabled` + `is-disabled` (bundle look), still focusable with the system ring, the hint says why, no list; error = `is-error` + `aria-invalid` + `aria-describedby` the field message. Value 400 like a field value; a value standing in for its row or region's code (Sidebar series, SexCountRow section) is `b` (500). No font reset, so touch values are 17/24 like a native touch select (S9 Din/În and S18 placement sections were 14px).
- **Unified look:** chevron `ink-secondary` everywhere (the Sidebar series and the CycleSelector had `ink`); the menu sits `space-1` under the field, as wide as it (`min-width: 100%`), z-index 3; the chosen option's check in the leading slot (S19 Ghid had it at the end; S9's Din list gains the empty slot).
- **Rule applied:** S4 ciclu nou "Ghid" (Fără ghid · Ghiduri de ouat · Ross 308 ouat "ghidul fermei", the same field as S19's Ghid, D112), S18 "Ferma de creștere" (Telița with its state: "goală din 28.02.2026", "6 secțiuni" or the ♀ · ♂ alive on the refused board) and S10 "Fermă" (Hîrbovăț with ♀ · ♂ alive, Telița disabled "goală") became `.ip-select`; S8 "Secțiune" (phone, opens the section 33 BottomSheet of sections with ♀ · ♂ alive or "goală") moved from its own `div.ip-input[role=combobox]` + `s8-grow` to `.ip-select` (fix round 1). Native selects kept (plain names, from a grep of the boards): S6 mortality Motiv (boală, șchiopătare, altul), S7 Tip hrană per section, S11/S17 Fermă filter (Toate fermele, Hîrbovăț, Telița), S17 Utilizator filter, S13 Tip hrană, S22 Rol. FIN12 (D133) brought in the last two: S14/S14S Fermă filter (desktop listbox, phone filter sheet; Telița disabled, reason `ip-note` "creștere, fără ouă") and S13 Fermă (phone boards and desktop panel; Telița disabled "fără ciclu deschis"). Native selects are now only the plain lists above. The select chevron rule lives only in additions.css (fix round 1 dropped the screen copies).
- **Tokens:** existing only (`ink`, `ink-secondary`, `ink-muted`, `space-1|2|4`, `--ip-fs-small`, `--ip-focus-*`); no new values.

### FIN11 · Skeleton consistency, 18.11.2026
CSS in additions.css sections 29, 35 and 42 (both copies, identical). Every skeleton the boards draw is now a system class; to copy into the system's Skeleton and DataTable READMEs with HANDOFF (D32). Decided by: Claude (recommended, user to review), D131.

- **Width modifiers** (section 42, any `ip-skel`): `ip-skel--ch1…4` (1–4 `ch` of the inherited font, the Mono of a figure cell) replace SYS11's `ip-gimp__sk1…4` (section 29 rules removed; S20 se citește and the preview unchanged in pixels); `ip-skel--key` (`control-h`, an IconButton's width) for an action column holding one.
- **UserRow · loading:** the Acțiuni bar is `--key --end` (was `--w3`, 72px in the 40px icon column, which widened it to 80); the column keeps its loaded 40px (S22 `--se-incarca`, the other columns take the 40px back).
- **S11 · se încarcă:** the three Stoc readouts are the section 35 StockReadout skeleton (light `panel` box, 5 bars), as S2 and S8, replacing the dark `.s11-skro` box (screen CSS removed from S11–S13); the box is 158 tall as on S2 (was 192; a loaded StockReadout row is 197), board height 1241 → 1207.
- **S18 · se încarcă:** the CycleTrack skeleton is plain `ip-trackskel` (the `s18s-trackskel` hook had no rule) and the table rows are `tr.ip-table__skel` (group row `is-group ip-table__skel`), replacing the `.s18s-skt` copy of section 42 (pixels unchanged).
- **Not moved:** S16S's field skeletons (`s16s-sk*`: a label bar `space-2` above a CountInput, Motiv area or effect box, full width) hold no figure width for `--ch`; they stay screen CSS with S6S/S7S/S8S's `s7s-skcol` until one system field skeleton covers all four (D131). Content-sized columns of loading tables (S11, S18–S22) still differ from their loaded widths.
- **Tokens:** existing only (`control-h`, `ch`); no new values.

### FIN11G · Skeletons keep the loaded geometry, 18.11.2026
CSS in additions.css section 54 (both copies, identical; section 35's StockReadout and CycleTrack skeleton rules moved there). Markup helper `design/canvas/scripts/skel.py`; to copy into the system's Skeleton, Readout, CycleTrack, TextField and DataTable READMEs with HANDOFF (D32). Decided by: Claude, to review (D132).

- **Readout skeleton** `div.ip-readout(.ip-readout--stock).ip-readout--skel[aria-hidden]`: the loaded Readout's own lines (`__cap > span`, `__val` + `small`, `dl.__kv` rows `dt`/`dd`, `__sub`; S8: `__live` then two kv rows), one bar each, so each line keeps its font and line height; light `panel` fill, hair `line` edge as an inset outline (no layout), kv rule hidden. StockReadout = cap, value, 3 kv rows, sub (a readout row stretches to its tallest, which has the sub line): 197 as loaded (was 158 on S2, S11; S8 161 as loaded, was 158). With a GuideChart (S4, S14 key Readout) the chart's own `svg.ip-chart` with the loaded viewBox (sunken), its read and legend lines (was the screen box `.s4s-skreadout`, 338). The Stats beside it are `div.ip-stat.ip-stat--skel` (fix 1): the Stat's own cap, value and guide lines, each guide item two touching bars (`wbr` between) as wide as the loaded item, so a guide line wraps where the loaded text does; the Readout row is then as loaded at both widths: 364 at 1440, 364 (S4) / 404 (S14, guide lines wrap) at 1280 (was 334 at 1280; replaces `.s4s-skstat`). Replaces `.ip-readouts__skel`.
- **CycleTrack skeleton** `div.ip-trackskel` (full, series page) 120 tall, `--compact` (Acasă, farm header) 90: the drawn track's viewBox height; the two lane bars sit where the cycle blocks do, the axis bar at the foot (was 68 for both; S18 120, S2 90). Under it the legend line as loaded (`div.ip-tracklegend > div.ip-skel-line`, fix 1; S18's table no longer drops 24px on load).
- **Field skeleton** `div.ip-field.ip-field--skel[aria-hidden]` > `span.ip-label` (bar) + the loaded field's own empty box (`ip-input` + `--touch` / `ip-count` / `--area`, an area holding one empty line per textarea row) + `span.ip-hint` (one `ip-skel-line` per hint line): sunken box, no edge, no record edge, no hover. Any other loaded text line (SaveBar summary lines, an AllocationRow's who/kg lines, an EntryCard head, an EffectPreview line) is `span.ip-skel-line` > bar in that block's own element. Replaces S16S `s16s-sk*` and S6S/S7S/S8S `s7s-skcol` (screen CSS removed); S16's effect box is the EffectPreview's own lines with the record icon (known from the row).
- **DataTable columns:** a table that loads declares its column shares once on the head the loaded and the loading table share: each leaf `th[style="--ip-colw:n%"]` (rowspan/colspan aware), measured from the loaded table at 1440; one text column declares none and takes the rest. Fix 1: each head cell also carries the loaded column's min-content width, `--ip-colmin` (measured, `space-1` units, without the cell padding), drawn as a zero-height `th::after` strut, so where the loaded content outgrows a share (1280, a table that scrolls) the loading column grows as the loaded one does. Rule `:where(.ip-table th[style*="--ip-colw"]){width:var(--ip-colw)}` (zero specificity): S19, S21 and S22 (container query that folds columns beside a SidePanel) and S18 (`.s18-vp`) set `th{width:auto}` and `th::after{content:none}` there, so that layout stays content-sized. S12 keeps its colgroup (two fixed edit columns) with the four content columns in %. Skeleton bars `--w1…w3` inside `tr.ip-table__skel` are capped at the cell (`min(…, 100%)`, out of the min-content width). Tables: S11 Livrări (owner), Consum, Cost; S12; S14 Pe săptămâni (desktop); S18; S19; S20; S21; S22.
- **Loaded boards:** columns now keep their place across states (e.g. S18 salvat moves up to 56px, S22 dezactivat/parola-cont-nou up to 48px, S11 modifică preț 30px against their old content-sized layout; at 1280 S11 10.6, S22 doar-tu 10.2, S20 7.2); default boards move ≤ 1.8px at 1440; panel boards unchanged.
- **Not changed (follow-up):** blocks around the skeletons that are data-sized (S2 track head link 24 vs 16, list rows, stock note, FarmSummary skeletons; S16 status line 2 lines vs the 1-line "Introdus de"; S18 table rows 32 vs 45/50 and its notes); the S3 se încarcă track and FarmSummary skeletons stay screen CSS (`s3-sk*`; the track box is already 120 as loaded); loading tables draw 2–6 rows of 32 against taller loaded rows.
- **Tokens:** existing only (`panel`, `line`, `sunken`, `space-*`, `radius-sm`, `--ip-hair`); the % shares are measured column widths; no new values.

### FIN13 · System docs and previews, 18.11.2026
CSS in additions.css sections 29, 34, 47 and the new section 55 (both copies, identical); §6 and §18 lose their old select rules. Previews, READMEs and scripts in `design/system/`. Decided by: Claude, to review (D134).

- **CycleTrack · track text** (section 55, system-wide): `.ip-track` azi label, Readout figure and month labels at the caption size (11px; bundle.css drew 10); the S2/S3/S18 screen copies are gone. The S4 farm track's Readout box 48 → 56 ("SAPT 40" at 11px Mono, as S3); CycleTrack/CycleTrackFarm previews likewise. CycleTrack README gains the phone variant (D91) and the track text line.
- **Hidden week text** (D115, D128 follow-up): CycleTrack block names, titles and `desc`, and GuideChart / sparkline `aria-label`s read "sapt. n" ("azi sapt. 40", "sapt. 3–40") on S2–S4, S9, S10, S14, S16–S18 and in the SYS01/SYS06 previews; table and region names keep SAPT (captions, column names), drawn Readouts and Mono lines keep SAPT.
- **SYS11:** refused-row reasons and error-list links use the `ip-sex` glyph (aria-hidden) + `ip-sr` word ("Rândul 8 are deja sapt. 4 ♀."); section 29 keeps the 12/16 line in Verificare.
- **SexCountRow · Toate secțiunile** (section 34): the key also pulls left by its hair border, so its box sits on the rows' Toate boxes (was 1px right). The SexCountRow preview draws the key "mixed" (H1.1 Toate, the rest off) and its SectionSelect as §53; SeriesSelector preview likewise §53.
- **SectionTree:** the edit-row error is the §38 field message with its icon (helper; S19 unchanged); the Retrage note is described once, by Retrage only.
- **Other:** EntryKeys disabled keys `tabindex=0` (preview, S2S); UserRow load error = §49 Banner `role=alert`, `.ip-users__bnrow` rules removed; Checkbox preview/README draw §44; Button README §46; FeedTotals preview adds the refused and saved SaveBar summaries (D62); FIN09 ranges in FarmSummary, StockReadout, CostBlock, FeedTotals, EntryCardPlus; LoginCard and SexCountRow card heights measured.
- **Tokens:** existing only (`--ip-fs-caption`, `--ip-hair`); no new values.

### FIN13B · System previews follow-up, 18.11.2026
CSS in additions.css section 9 (both copies, identical); previews and scripts in `design/system/`. Decided by: Claude, to review (D135).

- **FarmSummary · narrow figure row** (section 9): columns `minmax(0,1fr) auto auto`, areas "key val val" / "sub sub status" (2 columns when the row has no Lamp); the sub-line reaches the Lamp, not the value's width. The 390 Telița Hrană totală sub-line fits (was 25px over); `--wide` and `--compact` unchanged; S2/S3 boards identical in pixels.
- **Preview card heights:** dsCard = content height at 1440 + 24 above and below (LoginCard's rule); 34 previews raised (scripts sys04, sys06_gen, sys08–sys11 carry the new values).
- **Scripts:** sys04/05/08/09/10/13 builds and sys05–sys10/12/13 renders find their paths from `__file__`; s1_build and s19_build pass `__file__` to the sys13/sys10 exec (as s20, s22). Outputs byte-identical.
- **Tokens:** existing only; no new values.

### FIN14 · DataTable copies folded into the system, 18.11.2026
CSS in additions.css sections 11, 16, 32, 36 and 42 (both copies, identical). Generators s3, s4t, s11 (S12, S13 through it), s12, s13, s20; previews CostBlock (sys04_build), GuideChartVariants (sys06_gen). To copy into the system's DataTable, CostBlock and GuideChartVariants READMEs with HANDOFF (D32). Decided by: Claude, to review (D136).

- **Row headers** (§32): every table whose rows are named by `th[scope=row]` is `ip-table--rowheads`; the rule now covers `tfoot` (S12 "Total · 4 livrări"). Replaced: S3 `s3-rowh`/`s3-group` (→ `th.ip-table__child`, group th), S11 `.s11-table th[scope=row|rowgroup]`, `.s11-rh`, S12 `.s12-toth`, CostBlock's own row-header base (§11 keeps only weight 500, top alignment, `space-2` padding), `.ip-sparktable tbody th` (S4 Greutate, S20 Pe indicatori, preview). Look: S3 child words "găini/cocoși" `ink-secondary` → `ink` and S3 group rows 40 → `row-h` 32 (board 1134 → 1117, seria închisă 1028 → 1024); S20 rows without a curve 35 → 32 (−11px under the table; stored heights kept, s20_render cannot shrink a board).
- **Wrapping heads** (§32, opt-in `ip-table--wraphead`): `thead th.num` may wrap, heads sit at the bottom; replaces S3's local `th.num` rule. S4 Producție "Pe hale" uses it: at 1280 "Față de ghid" and "Masă ou" wrap and the 6px sideways scroll is gone ("Cat. 1–3" keep a no-break space); 1440 unchanged. Not for heads with `--ip-colmin` struts (§54).
- **Scroll region** (§36, moved from §16): S11 Livrări/Consum, S12 and S13 (beside the SidePanel) use `div.ip-tablescroll.ip-tablescroll--framed[role=region][tabindex=0][aria-label=caption]` instead of `.s11-tablebox.s11-tscroll` (loading copies `aria-hidden`, no tab stop); S12's price table becomes a named region too. The first column is now sticky on S11/S12/S13 (known issue closed) with the `rule` edge. Amendments: a first cell spanning columns does not stick or draw the edge (S11 Preț nou/Motiv rows, S12 error row, S18/S19/S20 group rows, S4 Greutate group rows); the first column's `raised` fill has zero specificity of its own, so row states keep their fill (S11 edit rows `sunken`, bundle group rows).
- **Skeleton bars in group rows** (§42): `tr.is-group .ip-skel` in `rule`, so they show on `sunken` (S3's screen copy removed; now visible on S14, S18 se încarcă and the SectionTree preview).
- **Side effects:** collapsed → separate borders on S11/S12 tables: S11 content +0.5px (stored heights kept, inside the foot padding; no --canvas step for S11); S12 totals edge is `rule` + `line-strong` like S4. FIN11G columns: loading vs loaded ≤0.3px at 1440 and 1280 on S11, S12, S14 (as before).
- **Left:** the S11 Consum horizontal scrollbar still sits at the table's foot below ~1180px wide (needs a height-capped region, D136); S3 table frame (`.s3-table` border/radius) and the 14px head-icon copies (`.s4-ico`, `.s4t-thico`, `.s11-thico`) are still screen CSS; `.ip-cost__unpriced > th` (400, `ink-secondary`) loses to the row-header rule, as before.
- **Tokens:** existing only; no new values.

### FIN14B · DataTable follow-ups: totals edge, head icon, S3 region, CostBlock region, 18.11.2026
CSS in additions.css sections 11 and 32 (both copies, identical). Generators s3, s4t (S14, S16, S17 through it), s11 (S13), s12; previews CostBlock (sys04_build), DataTable. To copy into the system's DataTable and CostBlock READMEs with HANDOFF (D32). Decided by: Claude, to review (D137).

- **Totals edge** (§32): one rule for every `tr.is-total` (after a body row, first in a following tbody, in a tfoot): the row above drops its own `rule`, the totals row keeps its `line-strong` top edge. S4 Fermă tabs (Păsări, Hrană, Producție; S17 copy) and S12 lose the 2px rule + line-strong double edge; S12's last row is 1px shorter (content-sized). CostBlock's own totals copy is gone (its cycle separator skips the totals row). The DataTable preview (collapsed borders) now shows line-strong instead of the rule.
- **Head icon** (§32): `span.ip-table__ico` replaces `.s4t-thico`, `.s11-thico` and `.s3-table th .s4-ico` (same look: inline-flex, top, `space-1`, record colour from `ip-rec--*`); the DataTable preview uses it (its icons move from bundle.css's 2px to `space-1` + 2px, top-aligned). S3's group-line icons keep their spacing through `.s3-group .s4-ico`.
- **S3 table frame**: the Păsări table sits in `div.ip-tablescroll.ip-tablescroll--framed` (region named by the caption; loading copy plain), replacing `.s3-table` border/radius/overflow and its last-row rule. Look: the sticky first column's `rule` edge, as S4 Păsări; heights unchanged. The 1280 figure padding (container query) stays screen CSS: S3-only density, not a DataTable rule.
- **CostBlock**: `.ip-cost__scroll` is a region in every layout (`role=region`, `tabindex=0`, `aria-label` = caption); not `--framed`, since the block draws the box with its head and foot. Dead `.ip-cost__unpriced > th` removed (the row header was already 500 `ink`).
- **CostBlock wide** (fix 1): `ip-cost--wide` no longer turns the sticky first column off; it keeps the `rule` edge as the phone layout and every §36 table. S13 at 1280: the 282px column (cycle names, "Total Seria 2026") stays while 137px scroll; S11 at 1440 and the preview only gain the edge.
- **Left:** the S11 Consum horizontal scrollbar still sits at the table's foot (D136).
- **Tokens:** existing only; no new values.

### FIN15 · Shared shell classes folded into the system, 18.11.2026
CSS in additions.css sections 56–60 (new) and 20 (renamed in place) (both copies, identical). Every generator that drew a page shell (s2–s4, s4t, s6–s22 builds, s3/s4t/s6/s6s/s8–s14/s16 renders, sstate_css); previews FeedTotals, HallCard (hand-written). To copy into the system's Sidebar, BottomTabBar, SidePanel, EmptyState, Dialog/BottomSheet, SaveBar, FeedTotals and HallCard READMEs with HANDOFF (D32). Decided by: Claude, to review (D138).

- **Shell** (§56, desktop): `div.ip-shell` > `nav.ip-sidebar.ip-shell__side` + `main.ip-shell__main` (space-6 padding and gaps); Sidebar foot `div.ip-note.ip-sidebar__account` (role in ink 500, Ieși ≥24px, D46); PageHeader meta line `div.ip-row.ip-pagehead__meta`; section head `div.ip-sechead` (h2 left, note or text link right on one baseline; the link ≥24px, D46 — was S4/S9/S10 screen CSS, now every section-head link). Replaces s4-shell/side/main/account/meta/head.
- **Sidebar:** the account foot is `ip-sidebar__account`; the spacer before it is `span.ip-grow` (FIN15B, §61).
- **BottomTabBar / phone screen** (§57): `div.ip-screen` (column on ground, frame-wide, no 390px of its own) > `main.ip-screen__main` (space-4 edges, space-8 foot, space-6 gaps) + `nav.ip-tabbar.ip-screen__tabbar` (sticky at the screen's foot, full width). Replaces s7-screen/s7-main, s2-phone (dropped: its only rule was the min-height ip-screen has) and s2-tabbar.
- **SidePanel** (§58): `aside.ip-sidepanel.ip-sidepanel--dock` (non-modal, D73: fixed to the right of the positioned page, full height, panel-w); head `div.ip-row.ip-sidepanel__title` (RecordTile + title block, space-2, no wrap; also in the phone sheet); foot `div.ip-row.ip-sidepanel__keys`; utility `.ip-m0` (no margin on a heading/paragraph in a gapped block). Replaces s13-panel/phead/pkeys/m0. The page beside it (`ip-dockpage`) and the scrolling form body (`ip-sidepanel__body--form`) are §62 (FIN15B).
- **EmptyState** (§59): `div.ip-emptybox > div.ip-empty` (panel, 1px line, radius-md) with `p.ip-empty__line` (first ink, the rest ink-secondary); the next step (`.ip-empty > .ip-btn|.ip-link`) sits space-2 lower, now on every EmptyState. Look: S6 Zilnic and S8 Cântărire "fermă fără păsări" keys 8px lower (were the only EmptyStates without it); GuideImportPreview preview key +8. Replaces s4s-emptybox/s4s-empty-p.
- **Dialog / BottomSheet on a phone** (§60): `.ip-scrimhost` on the phone screen positions it, so the `ip-scrim--phone` / `--sheet` scrim fills it. Replaces s9-modal.
- **SaveBar** (§20, renamed): the two-line summary is the SaveBar's own `span.ip-savebar__sum.ip-savebar__sum--2l` > two `span.ip-savebar__line` (segments never break inside); used by S6, S7, S8 (was FeedTotals' `ip-ftot-sum`/`__line`).
- **HallCard:** Ouat azi sub-line "din ♀ 8 556" is `span.ip-eggs__base`, kept on one line by the system rule (was S6's `.s6-nobr`); the sub-line still wraps between segments (a whole-line nowrap would overflow: 29px on S6 abatere at 390, 86px in the preview).
- **Left:** `s4-block`, `s4-grow`, `s13-vp`, `s13-pbody`, `s7s-vp/view`, `s7-savebar` were still screen copies shared by many generators: moved by FIN15B.
- **Tokens:** existing only; no new values.

### FIN15B · The last shell copies folded into the system, 18.11.2026
CSS in additions.css sections 61–65 (new; both copies, identical). Generators: every build that drew a desktop page, a docked SidePanel or a fixed phone screen (33 builds: s2, s2s, s3, s4s, s4t, s6, s6s, s7, s7s, s8, s8s, s9–s22 and their s builds, plus sstate_css; 7 renders: s6s, s8, s9, s10, s13, s16, s16s). To copy into the system's Sidebar, SidePanel, BottomTabBar/SaveBar READMEs and the utilities list with HANDOFF (D32). Decided by: Claude, to review (D139).

- **Page block** (§61): `section.ip-block` = one page block (`div.ip-sechead` + its table, list, tiles or chart): column, space-3 inside, `min-width:0` so a wide table scrolls in its region instead of widening the page. Replaces s4-block (S2–S4, S9–S22).
- **Spacer** (§61): utility `.ip-grow` (`flex:1`): the Sidebar's `span.ip-grow` before the account foot, a Banner's text block beside its icon (S19S). Replaces s4-grow.
- **Docked page** (§62): `div.ip-dockpage` > `div.ip-shell`, beside `aside.ip-sidepanel--dock` (§58): absolute in the positioned page frame (`.ip-root`), `inset: 0 panel-w 0 0`, clips at the frame's foot (in the app it scrolls). Its height is the frame's (the board root's min-height), never its content: no 1024px in the system. A page with the panel closed sets `right:0` (screen modifier, S16S). Replaces s13-vp (S9, S10, S13, S16, S18–S22).
- **SidePanel form body** (§62): `div.ip-sidepanel__body.ip-sidepanel__body--form` takes the height between head and foot keys and scrolls, fields space-4 apart (bundle body: space-3); docked and as the phone sheet. Replaces s13-pbody.
- **Fixed phone screen** (§63): `div.ip-screen.ip-screen--fixed` is exactly its frame's height (`height:100%`; the board root is `height:844px; overflow:hidden`, now also on S14/S17 telefon-filtre and S8 alege-sectiunea, which had only min-height), so the FormHeader stays on top and the SaveBar or BottomTabBar at the foot; `div.ip-screen__view` between them takes the rest and clips the form (its `main.ip-screen__main` keeps its own height), `--end` shows the form's end. `div.ip-savebar.ip-screen__savebar` = the SaveBar sticky at the screen's foot (as `ip-screen__tabbar`, §57). With `.ip-scrimhost` (§60) for a modal. Replaces s7s-vp, s7s-view(--end), s7-savebar (S2, S6–S10, S13–S15, S17).
- **Utilities** (§64): `.ip-nowrap` (one line: a figure with its unit or lamp, a date range, a name) replaces sNN-nw on S2, S3, S11, S13, S14, S18, S19, S21, S22; `.ip-muted` (ink-muted: a secondary figure, the unknown "–" with its reason) replaces s4t-muted, s11-muted, s3/s18/s20-dash, s20-none; `.ip-m0` (§58) also replaces s4s-, s7-, s2-, s3-, s11-, s12-m0.
- **CycleTrack width box** (§65, as `ip-chartbox`): `div.ip-trackbox` > one `svg.ip-track.ip-track--1160|1000|720|560` per viewBox width; a container query shows the widest drawing the box has that fits (1160 from 1160px, then 1000; 720 below 1000 and 560 below 720 only where the box holds them), so track text stays 11px and blocks keep their targets. Hidden drawings are display:none (out of tab order and the accessibility tree). Replaces s2-, s3-, s18-trackbox and their track--N rules (S2, S3, S18–S22); the 358 phone drawings keep their screen class. FIN16 uses it for the S4/S9/S10 track text at 1280.
- **Not moved** (same declarations, different jobs; D139): inline-flex + space-2 (Lamp cells s4t/s11-state, s11-type, s11-farm, s3-diff), table foot line (s4t-foot, s11-foot), column + space-6 (s4-panel, s14-page, s7s-skblock), 24px targets (s11-lk, s20-flink; s3-gname is a span), other responsive display switches, table widths.
- **Look:** none. All 255 boards at board size + 158 desktop at 1280 pixel-identical to before; canvas.json and heights unchanged.
- **Tokens:** existing only; no new values.

### FIN16 · S4 leftovers: narrow farm page, farm track width box, list rows in li, De verificat empty, 18.11.2026
CSS in additions.css section 1 (amended) and the section 65 note (both copies, identical); S4 page layout in the S4 screen CSS (s4s_build, s4t_build). Generators s4s, s4t, s2, s2s, s9, s10, s14, s16, s16s; preview CycleTrack (hand-written). To copy into the system's EventRow/AlertRow, CycleTrack and CycleTrackFarm READMEs with HANDOFF (D32). Decided by: Claude, to review (D140).

- **S4 page width** (screen CSS): `main.ip-shell__main.s4-main` is a named size container (`s4main`); the page lays itself out from its own width, not from the board. From 1000 up (1440: 1160, 1280: 1000) as before, with the CycleSelector column 416 (was 336) so the track box is a CycleTrack step: 720 at 1440, 560 at 1280. Under 1000 (beside a SidePanel: 720 / 560) one column: CycleSelector field (kept 416 wide) over the track (page width), Readout over the Stats, halls, De verificat over Ultimele înregistrări, the S4T two-block rows. Replaces the S9/S10/S16 board rule `.ip-dockpage .s4-cycle,.ip-dockpage .s4-top` (+ S16's `.s16-vp .s4t-split`), which also caught the full-width S16 "după ștergere" page.
- **CycleTrack farm in the width box** (§65): `div.ip-trackbox` > `svg.ip-track.ip-track--farm.ip-track--1160|1000|720|560`, each drawn 1:1 (same marks, labels and ARIA; only x positions change), so at 1440 and 1280, with or without a SidePanel, the track text is 11px and the block targets 30px (was one 800 drawing scaled to 640 / 560: 8.8 / 7.7px text, 24 / 21px blocks).
- **EventRow links in a list** (§1): a list of row links is `ul.ip-list > li > a.ip-ev` (S4 Rezumat Ultimele înregistrări, as the tabs' `ul/li` lists); the li draws the rule between rows (`li + li`, hair `rule`) and the link keeps the list's outer corners and the focus ring above the next row. AlertRows are `li.ip-ev.ip-ev--alert` in `ul.ip-list` (S4 Rezumat, S2 De verificat). Look unchanged (pixel-identical).
- **AlertRow empty** (SYS01 "Empty"): the list's own empty row `li.ip-list__empty`: off Lamp "nimic de verificat" + one sentence naming why ("Ciclul Seria 2025 la Hîrbovăț e închis din 28.01.2026; nu are alerte." on S4 ciclu ales; "Telița e goală din 28.02.2026; alertele apar cu ciclul nou." on the new S2 board), as "nimic introdus" (SYS09) and "nimic lipsă"; no next step. Replaces S4's screen box `.s4s-none`.
- **CycleTrack block labels:** a label must fit its block; shorten it (drop the length, then `S 2026 · …`, then no label), the link's `aria-label` keeps it whole. The CycleTrack preview now draws the real Telița cycle (07.10.2025–28.02.2026, 21 săpt.) and transfer (22.02–28.02.2026) and Hîrbovăț SAPT 40.
- **Tokens:** existing only (`--space-8`, `--space-12`, `--ip-hair`); 416 = space-12 × 8 + space-8; no new values.
