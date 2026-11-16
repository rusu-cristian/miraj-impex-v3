# Impex Line Panel

Impex Line Panel is the control-panel design of Impex Line. Impex Line tracks each breeder-flock series through its cycles: growth at Telița, transfer, laying at Hîrbovăț, sale. Each farm follows its linked guide from the start of its cycle.

The interface is set like the hall's climate controller:
- light steel panels
- engraved uppercase captions
- graphite keys
- status lamps
- dark readout windows for the figures that matter

**Character:** alert, exact, mechanical.

## The three jobs of colour

Colour does exactly three things, and they never mix.

1. **Record type: what a record is.** Each type has a fixed icon and colour (`rec-*`), the same on every screen. A screen full of data can be scanned by type.
2. **Status: how it is.** A lamp plus a word:
   - `ok` / `ok-lamp`: în normă, current
   - `warn` / `warn-lamp`: deviation from the guide, unusual value
   - `fault`: refused, error
   - `info`: neutral
3. **Steel neutrals** for everything else.

### Record-type code

| Type | Token | Icon | Where it appears |
|---|---|---|---|
| Găini ♀ | `rec-hen` | venus | every hen figure, input bar, column, chart line |
| Cocoși ♂ | `rec-rooster` | mars | every rooster figure, input bar, column, chart line |
| Mortalitate | `rec-mortality` | skull | mortality records, columns, filters |
| Sacrificare | `rec-cull` | ban | cull records |
| Eroare sexare | `rec-hen` + `rec-rooster` | two rings on a split tile | sexing errors |
| Populare | `rec-placement` | log-in | placement, cycle start marks |
| Transfer | `rec-transfer` | arrows | transfers, the link on the cycle track |
| Vânzare | `rec-sale` | truck | sales, cycle end marks |
| Hrană | `rec-feed` | wheat · package · coins | daily grams · deliveries · prices |
| Cântărire | `rec-weighing` | scale | weighings, body weight |
| Ouă | `rec-eggs` | egg | egg production, lay % |

- The icon and the word (or the ♀/♂ sign) always come with the colour.
- Every `rec-*` colour is ≥4.5:1 on `ground`, `panel` and its own `-bg` tint.
- **Known close pairs:** rooster blue and weighing cerulean; cull plum and transfer violet. They're told apart by icon and word.

## Content

- **Language:** Romanian with full diacritics, in sentence case.
  - **Exception:** engraved captions (`caption` style) are UPPERCASE: column heads, panel captions, breadcrumbs and readout captions.
- **Address:** informal "tu".
- **Buttons are verbs:** Salvează, Mută, Înregistrează vânzarea, Salvează prețurile, Începe ciclu nou.
- **Errors name the section, the number and the fix:** "H1.1 are 1 012 cocoși. Nu poți scădea 1 100."
- **Counts:** `♀ 14 080 · ♂ 3 510`, in IBM Plex Mono, with ♀ in `rec-hen` and ♂ in `rec-rooster`.
  - Thousands are separated by a thin no-break space.
  - Use the true minus (−).
- **Dates and weeks:**
  - Dates are `dd.mm.yyyy`.
  - The guide week of the farm's cycle is written `SAPT 41` in mono.
  - The week counts from the cycle start, not from hatching.
- **Unknown values** are `–` with a reason ("fără ghid", "fără preț").
- **No** exclamation marks, emoji or hype.

## Surfaces

- **Surface roles:**
  - `ground`: the page
  - `panel`: cards, panels, stats, the sidebar
  - `raised`: inputs, tables, event rows, section tiles
  - `sunken`: tracks, hover, read-only
- **Edges:** separated by 1px `line` edges and `rule` row hairlines.
- **Shadows:** `shadow-overlay` is only for menus, dialogs, toasts and the side panel. `shadow-key` is only for buttons.
- **Text:**
  - `ink` (14.9:1) for text and numbers
  - `ink-secondary` (7.1:1) for meta and labels
  - `ink-muted` (5.2:1) for captions and hints
  - `ink-disabled` for disabled controls only
- **Keys:** the primary button is a graphite key (`control`) with a `shadow-key` lower edge. There's one per view.
- **Theme:** light only.

## Readouts (the Moment)

A **Readout** is a dark display set into the light UI.
- **Colours:** `readout` background, `readout-ink` figures, `readout-muted` captions, and the `readout-ok` / `readout-warn` / `readout-fault` lamps.
- **Use:** the key figure of a page, guide charts (actual `readout-ok`, guide dashed `readout-guide`, band `readout-band`, extra series `readout-2` … `readout-4`), feed stock, and the guide-week counters on the cycle track.
- **Limit:** at most one per page section. Feed stock readouts may repeat in a row.

## The signature: the cycle track

One lane per farm, along calendar time. Each cycle is a block:
- closed cycles: `panel`
- open cycles: `raised` with an `ink` edge

Markers on the track:
- **Placement start:** a `rec-placement` mark.
- **Transfer:** a `rec-transfer` link between lanes.
- **Sale end:** a `rec-sale` mark.
- **Today:** an `ok-lamp` "azi" line through all lanes, with a readout per open cycle showing its guide week.

Use the `full` variant on the series page and `compact` on the owner's Acasă and in farm headers.

## Type

- **IBM Plex Sans** for words.
- **IBM Plex Mono** for every number, date-week and readout.
- **Styles:** `title` 24/30, `title-sm` 20/26, `heading` 16/22, `body` 14/20, `input` 17/24 (phone), `small` 12/16, `caption` 11/14 uppercase +0.06em.
- **Readout figures:** `readout-lg` 32/36, `readout` 20/24, `figure` 14/20, `figure-sm` 12/16.
- **Weights:** 400 and 500.

## Space, size, shape

- **Spacing:** steps of 4.
  - Desktop: page edge `space-6`, panels padded `space-4`, compact rows `row-h` (32px), controls `control-h` (32px).
  - Phone: controls `control-h-touch` (48px), count inputs `count-input-h` (56px), page edge `space-4`.
- **Radii:**
  - `radius-xs` (2): tags, cycle blocks
  - `radius-sm` (3): controls, record tiles
  - `radius-md` (4): panels, readouts
  - `radius-lg` (6): dialogs, sheets
  - `radius-round`: lamps and the phone Înregistrează key only
- **Focus:** a 2px `focus` (green) outline with a 2px offset.

## Motion

- One easing: `cubic-bezier(0.2, 0, 0, 1)`.
- **Durations:** 120ms for feedback (key press), 200ms for panels, sheets and toasts, 280ms for page-level changes.
- **Lamps don't blink.** Charts, numbers and skeletons never animate.

## Layout

- **Desktop:**
  - Sidebar (`sidebar-w`) on `panel`. The current item has a green lamp.
  - Scope pages: PageHeader → CycleSelector and compact CycleTrack (farm level) → FilterBar with RecordChips → Tabs → content.
  - The farm Rezumat shows the Readout, Stats, SectionTiles, alerts and the latest EventRows.
- **Phone:**
  - BottomTabBar with the graphite Înregistrează key in the centre. Vets see 3 items.
  - Forms are FormHeader (with a RecordTile) → cards → SaveBar.

## Iconography

- **Icons:** Lucide-style line icons, 1.75 stroke, at 14/16/18/22px, in `currentColor`.
- **Record icons** are fixed per type and are never reused for anything else.
- **No photos or illustrations.**
- **No logo.** The wordmark is "IMPEX LINE" in Plex Sans 500, uppercase, tracked.

## Components

Components are specified as HTML/CSS with `ip-` classes in `components/bundle.css`. The app is built in Vue/Nuxt.

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

**Rules the app must enforce:**
- A section never goes below 0.
- Cycles open on the first birds into an empty farm and close at 0.
- The guide week counts from the cycle start.
- Unconfirmed feed suggestions aren't saved.
- Prices are hidden from vets.
