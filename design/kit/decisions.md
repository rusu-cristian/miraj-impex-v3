# Decisions

## 2026-11-16 · Bird counts are calculated from events, not stored
- Decided by: user
- Why: one `BIRD_EVENT` ledger (placement, mortality, sexing error, transfer, adjustment) means counts can't drift, and a transfer is one pair of rows.
- Rejected: one entity per kind of change; a daily wide row per hall like the spreadsheet.
- Rests on: the user's request for events/records instead of separate entities.

## 2026-11-16 · Data model v2: farm types, sections, cycles, feed types, priced-later deliveries, daily grams
- Decided by: user
- Why: each farm follows its own guide from the day birds arrive. The yearly series moves growth → laying → sale. The manager records deliveries and the owner adds prices later. The vet enters grams per section every day.
- Rejected: the guide week counted from the series placement day; the norm-until-changed model; the price entered with the delivery.
- Rests on: the whole spec v2, the cycle track, the feed report.

## 2026-11-16 · Roles: vet / viewer / manager / owner
- Decided by: user
- Rejected: operator / supervisor / manager / owner.

## 2026-11-16 · A fixed record-type code (icon + colour), shared by both designs
- Decided by: user (request); default (the hues)
- Why: so each kind of record is recognisable at a glance on a screen full of data.
- Values: hen #B0246A, rooster #1F5BB5, mortality #3A424A, cull #7C2F86, placement #B4450F, transfer #5B3CC4, sale #56680C, feed #7F5C00, weighing #0A6A8F, eggs #0D6F69. Sexing error is a split hen/rooster tile.
- Rejected: grey for placement (the user found it hard to spot).
- Known risk: rooster vs weighing and cull vs transfer are close. Icons and words carry the difference.

## 2026-11-16 · (superseded) The schema has 9 tables, plus users, user farms and an audit log
- Decided by: user (9 tables); default (USER, USER_FARM, AUDIT_LOG)
- Why: the user found the 14-entity version overcomplicated. Roles and past-data edits need the three additions.
- Rejected: SERIES_STAGE, FEED_TYPE, FEED_EVENT, Guide/GuideWeek split, a stored consumption ledger, the incubator.
- Rests on: the role model and the correction feature.

## 2026-11-16 · Feed price lives only on purchases; consumption and stock are in kg
- Decided by: user
- Why: price matters for the "food bought" view. Feeding is only about the stockpile.
- Rejected: weighted-average or FIFO costing on consumption.
- Rests on: the feed report and stock warnings.

## 2026-11-16 · New design system, not the existing "Impex Line" (walnut/abacus)
- Decided by: user
- Why: the user chose to start fresh.
- Rejected: reusing the 15.11.2026 system and its React components.
- Rests on: everything visual.

## 2026-11-16 · Direction: the farm ledger
- Decided by: user
- Why: warm paper, graphite and one barn-red ink fit an exact daily register.
- Rejected: eggshell and straw; a hall control panel.
- Rests on: all tokens.

## 2026-11-16 · Signature: the age ruler
- Decided by: user
- Why: every figure in this domain is read against the week of age.
- Rejected: a guide tick; the G · C pair as the signature (it stays a convention).
- Rests on: the series header, the Acasă header and the report headers.

## 2026-11-16 · Light theme only; mixed density
- Decided by: user
- Why: office PCs and bright halls; it's simpler to build and print. Roomy phone forms, compact desktop tables.
- Rejected: light and dark; a single density.

## 2026-11-16 · Danger is crimson #9E1F3F, separate from the barn-red accent
- Decided by: default
- Why: the accent is red-brown, so danger needed a cooler, darker hue, always paired with an icon and words.
- Rejected: using the accent for danger; an orange danger.
- Rests on: Tag, Banner, refusal states, destructive buttons.

## 2026-11-16 · Fonts: Source Serif 4 for titles, Source Sans 3 for the interface
- Decided by: default
- Why: a humanist ledger feel, good Romanian diacritics, tabular figures, free and self-hostable.
- Rejected: one family only; a monospace face for counts.

## 2026-11-16 · Online only; drafts kept in the browser until saved
- Decided by: user (online only); default (drafts)
- Why: connectivity in the halls is reliable enough.
- Rejected: offline sync.

## 2026-11-16 · Total eggs = Cat. 1 + Cat. 2 + Cat. 3 + Incubabile; "Jos" kept separate
- Decided by: default (approved in the spec)
- Why: the sample rows in DataHirbovat suggest it.
- Rests on: HallEntryCard, the production report.

## 2026-11-16 · Operators don't see feed prices
- Decided by: default (approved in the spec)
- Rests on: C4, the feed report.

## 2026-11-16 · Alert thresholds
- Decided by: default (approved in the spec)
- Values: mortality above 3× the hall's 7-day average; weight ±5%; lay ±3 points; feed below 5 days; missing entries shown from 12:00.

## 2026-11-16 · Operators can edit their own entries until 23:59 the same day
- Decided by: default
- Why: typos get fixed without a manager, and history stays trustworthy.
- Rests on: C6, E1.

## 2026-11-16 · The Seria 2026 history is imported once by the developer
- Decided by: default (approved in the spec)
- Rejected: an import screen.

## 2026-11-16 · Second design: the barn control panel (Impex Line Panel)
- Decided by: user
- Why: to compare with the ledger design. Light steel panels, IBM Plex Sans and Mono, graphite keys, status lamps, dark readouts.
- Rejected: eggshell and straw; a veterinary chart.
- Rests on: the Impex Line Panel design system.

## 2026-11-16 · Panel: light theme only; green means "normal / current", never a record type
- Decided by: user (light); default (the meaning of green)
- Why: in a control panel a green lamp means running and normal. Primary keys are graphite rather than green, so green stays a state.
- Rejected: a green primary button; dark theme first.

## 2026-11-16 · Panel signature: the cycle track
- Decided by: user
- Rejected: a section map; a guide gauge.

## 2026-11-16 · Vet feed entry: yesterday's grams are a suggestion that must be confirmed
- Decided by: default
- Why: grams are entered daily, and carrying a value over silently would hide days nobody entered.
- Rests on: AllocationRow, C2.

## 2026-11-16 · Setup-check resolutions (Decided by: default (user away) — to review)
- **D24 Guide week today.** Today in screens is 16.11.2026; Hîrbovăț cycle from 10.02.2026 → **sapt. 40** by the data-model formula. The spec's "sapt. 41" examples are illustrative and are rewritten to 40 (the B4 dialog note becomes "Ciclul curent (sapt. 40) se închide pe 16.11.2026"). Rejected: moving today to 17.11.2026.
- **D25 Hîrbovăț cycle start.** Keep 10.02.2026 (first Hrana Hîrbovăț day) as the laying cycle start, although transfer rows are dated 22–28.02.2026; Telița's cycle closes 28.02.2026. Leftover Telița counts (e.g. H3 −9) are treated as data errors, not shown. Rejected: starting the cycle on 22.02.2026.
- **D26 Laying guide index.** Workbook guides are by age week; cycle week = age week − 18 (sapt. 40 = age wk 58: ♀ 158 g, ♂ 155 g). Growth cycle week = age week.
- **D27 Sample data.** The workbook ends 28.02.2026 (one feed row 10.03.2026). Seria 2026 Hîrbovăț figures after that come from design/data/extracts/sample-hirbovat-2026.json, generated deterministically and labelled illustrative; the canvas "00 · Index" board lists which figures are sample. Screens stay clean (no "exemplu" badges in the UI). The cycle selector's past Hîrbovăț cycle uses the real previous-flock egg data (04.04.2025–28.01.2026).
- **D28 Workbook wins over kit samples.** Where kit examples contradict the workbook (H1.1 ♀ 9 402; ♀ 26 881 · ♂ 2 994; "8 secțiuni"), screens use workbook-derived figures (arrival ♀ 26 765 · ♂ 3 060; 12 allocation rows), and sample-derived counts for later dates.
- **D29 Missing guide data.** No guide for laying mortality, hatching, egg weight, body weight after wk 23, CV/uniformity: show "–" with "fără date în ghid" (a guide line, not a Tag). The Sheet9 col P series is used as the illustrative % lay guide (age wk 25–63).
- **D30 Feed types.** Illustrative names: Hîrbovăț "Breeder 2" (♀) and "Cocoși" (♂); Telița "Starter", "Grower", "Pre-ouat" mapped from workbook notes (rost, razvitie, preouat). No Hîrbovăț deliveries exist in the workbook: Hîrbovăț stock/deliveries are sample, using the spec's 24 000 kg Breeder 2 example.
- **D31 Component rule clarifications.** Stat value uses the `readout` token (20/24), not a raw 24px. Uppercase also allowed for the mono Readout caption, the "IMPEX LINE" wordmark and CycleTrack lane labels. Green without a word is allowed only for non-status cues (Sidebar current item, Tab dot, "azi" line, focus ring). 44px targets apply on touch screens; desktop keeps 32px controls. AllocationRow uses a deviation lamp + word (component spec), not a tag. Guide week: `SAPT 40` in mono context lines/Readouts, `sapt. 40` in running text and toasts.
- **D32 New components.** The design-system artifact can't be edited from this session, so new components the screens need are specified by default (in the system's own language, from existing tokens), appended to components.md under "Additions", built in design/system/components/<Name>/ with CSS in design/system/components/additions.css, published to the canvas, and listed in HANDOFF for copying into the system artifact.
- **D33 No user screen gates while the user is away.** Screens are built, reviewed and polished, then marked done; the user reviews all artboards and these decisions on return.
- **D34 Interaction states (SYS00).** Decided by: default (user away) — to review. Real :hover/:active/:disabled live in design/system/components/additions.css, published at canvas `project/additions.css` (the `ds/` folder is install-only). Pressed for ghost buttons, IconButton and EventRow = one steel step deeper (`sunken` → `rule`); RecordChip hover shows its edge in its own record colour. Artboards link tokens.css → bundle.css → additions.css.
- **D35 Data mapping (DATA01).** Decided by: default (user away) — to review. The "norponin" delivery (29.11.2025) maps to Grower. The previous Hîrbovăț cycle starts 04.04.2025 (earliest data, age wk 23, start inferred). Workbook anomalies (H3 −9, H6 −12, H5-2 −2; transfer pairing vs sheet names; H3/H6 weighing years; H52 wk 15–19) are kept in extract meta and never shown as real counts below zero.
- **D34a Pressed states (SYS00 review).** Decided by: default (user away) — to review. Secondary pressed = `rule` (one step deeper than hover `sunken`), like ghost/IconButton/EventRow. Destructive pressed = `color-mix(in srgb, var(--fault) 8%, var(--fault-bg))`. Chip pressed = `panel`. While pressed (on `rule`), `ink-muted` text and lamp words inside interactive rows switch to `ink-secondary` to stay ≥4.5:1.
- **D36 Previous Hîrbovăț cycle & culls (DATA02).** Decided by: default (user away) — to review. The previous flock's cycle week = age week − 18 (as D26), so its inferred cycle start is 07.03.2025 (age wk 19), replacing D35's 04.04.2025; its data starts 04.04.2025 (cycle wk 5). BIRD_EVENT gets type `cull` with a `reason` (boală, șchiopătare, altul), because spec C1 needs culls with a reason.
- **D37 S4 components (SYS01).** Decided by: default (user away) — to review. Extensions live in new folders (CycleTrackFarm, SectionTotals, AlertRow, GuideMissing, SeriesSelector, FilterInterval) so the system copies stay identical to the artifact. Farm track drops the lane label (page header names the farm; aria-label keeps it) and has a legend for its marks. Selected cycle = ink underline + aria-current. AlertRow: link then status lamp+word in a fixed right column. SectionTotals = panel strip with line-strong top edge, names any section missing today ("medie, fără H1.2"). No-guide farm = neutral Tag "fără ghid"; on Readouts a hollow off lamp with "FĂRĂ GHID"/"FĂRĂ DATE ÎN GHID". Series selector trigger = lamp + "deschisă"/"închisă". Interval = inline labelled date fields; typed future date error "Alege o dată până la 16.11.2026."
- **D38 SYS01 review calls.** Decided by: default (user away) — to review. Cycle blocks are links, so closed blocks use an `ink-muted` edge (≥3:1 on the sunken lane) in both CycleTrack variants (additions.css), instead of `line-strong` (2.99:1). Farm/section averages of grams weight ♀ by hens and ♂ by roosters. No-guide demos use a farm labelled as an example; the partial-guide example is Telița's growth cycle (weight guide ends at SAPT 23 there).
- **D38a Partial-guide demo.** Decided by: default (user away) — to review. Telița's Seria 2026 cycle only reaches SAPT 21, so the partial-guide component demo is labelled "Fermă exemplu · creștere" (sample actuals SAPT 1–25, real growth guide ending SAPT 23), not Telița. Screens show partial guides only where their own data reaches past the guide's end. Interval date labels sit above the fields (bar grows ~20px when Interval is on).
- **D39 State-example rows.** Decided by: default (user away) — to review. A component preview may keep one row that isn't in the extracts when it's the only way to show a spec state (AlertRow "refuzat": H4.2 ♂ −500 sale refused, ♂ 476 alive from the extract), with a footnote saying it's a state example. Screens never do this: their states use extract data or are separate state artboards.
- **D40 FeedTotals (SYS02).** Decided by: default (user away) — to review. "Stoc sub 0" is a warning, not a refusal: warn lamp "stoc sub 0" + warn Banner with the C2 sentence; Salvează stays enabled (overrides the builder's fault lamp). SaveBar summary counts section × sex rows ("10 din 12 rânduri · 3 691,3 kg"), never "secțiuni". Today's kg one decimal, stocks whole kg. Nothing confirmed: Azi "–", stock now and days from it. Desktop layout via `ip-ftot--wide` modifier. The −1 240 kg row is a D39 state example.
- **D41 S4 layout calls.** Decided by: default (user away) — to review. % ouat shown per hall inside each tile ("Ouat H1 · ghid 74,1 %"). Section overview = 3 hall columns with 2 tiles stacked (6 in a row is too tight). FilterBar shows the 5 record types with records in the period (all on) and grain label "Săpt.", so it stays one row. CycleSelector trigger = lamp + "deschis", no SAPT. Feed-stock alerts lead with the package (delivery) tile. EventRows carry a tag only where it adds information ("lipsă 2", "corectat"); the latest-records log shows 15–16.11 (9 rows). Section tiles link to S5 (not built in this run).
- **D42 Static artboards.** Decided by: default (user away) — to review. Canvas artboards carry no UI scripts (canvas.md), so behaviours like the GuideChart ←/→ week readout are specified in markup/ARIA (role="group", aria-roledescription="grafic", live region) and the component spec, and are wired in the Vue build. S4 keeps % ouat in the Readout (week), the Stat (today) and the totals strip, as the spec lists it in the stats. Status words are only the Lamp vocabulary (în normă / abatere / refuzat / informație / lipsă / oprit), never ad-hoc phrases like "sub 5 zile".
- **D43 FarmSummary (SYS03).** Decided by: default (user away) — to review. A closed cycle shows its placement ("Populat · 07.10.2025") instead of 0 alive. S2 compact shows % ouat + mortality. Guide-name tags may exceed the 16-character Tag limit ("Ross 308 creștere"); every other Tag keeps ≤16. Narrow by default, `--wide` on S3, `--touch` on phone. S2 shows each farm today (Telița empty); S3 shows the farm's cycle in that series (Telița closed).
- **D44 S4 fix calls.** Decided by: default (user away) — to review. A SectionTile may show two lamps when two states hold (H3.2: "lipsă" for feed and "abatere" for lay). The S4 Readout chart is 548×220 so the Readout matches the three Stat rows. Charts use role="group" + aria-roledescription="grafic" (D42); the GuideMissing/GuideChart previews still using role="img" are aligned in SYS06.
- **D45 FilterBar chips (amends D41).** Decided by: default (user away) — to review. The FilterBar lists every record type with records in the selected period (on S4 "Ciclul curent" this adds Eroare sexare and Transfer); the bar may wrap to a second row rather than hide types.
- **D46 SYS03 review calls.** Decided by: default (user away) — to review. Feed vs guide compares like with like: Σ daily g/pasăre over the days with feed rows against the guide over the same days (Telița ♀ 8 401 · ♂ 10 142 g vs ♀ 9 065 · ♂ 10 835 g → în normă). Percentages are written without a space before % ("74,0%", "−1,4%"), product-wide, as craft.md/spec C7. Cycle-state lamp words "deschis", "închis", "goală" are allowed outside the D42 status vocabulary (state, not status). Desktop text links keep a ≥24px target.
- **D47 S4 final calls.** Decided by: default (user away) — to review. FilterBar on S4 is two rows (period/grain/export, then 7 record chips). The Eroare sexare chip is a neutral chip led by the split ♀/♂ tile (moves from screen CSS `.s4-chipsplit` into additions.css in SYS06). Totals guide line may compact to "ghid 158/155 g" where space is tight.
- **D48 S4 states (S4S).** Decided by: default (user away) — to review. No-guide: a "fără ghid" Tag only in the header; Stats/tiles use the guide line "ghid: – fără ghid". EmptyState lamp words are state labels ("fără date", "fără acces", "goală"). Empty farm hides the filter bar and tabs (they act on a cycle) and keeps only "Mută păsări"; next step "Vezi ultimul ciclu". Viewer role label in Romanian = "Observator" (nav: Acasă · Serie · Rapoarte). No data in period: no type chips, Exportă disabled, next step "Alege ciclul curent". Open CycleSelector = listbox of cycles, then a separate "Ciclu" group of action buttons (manager+).
- **D49 Stock & cost (SYS04).** Decided by: default (user away) — to review. Lei/pasăre uses birds at cycle start (placed or arrived by transfer). A stock with no recent consumption shows the off lamp "oprit". CostBlock on phone scrolls with a sticky first column; desktop `ip-cost--wide`. Coins icon = Lucide "coins" path (the system names it but ships none). One read-only markup for owner/manager/viewer; vets never get the section. Stock readouts carry no record icon (rec-feed is 2,92:1 on the dark readout); the caption word names the feed type.
- **D50 "Stoc sub 0" wording (amends D40).** Decided by: default (user away) — to review. Lamp words stay within D42: a stock that would go below 0 shows the warn lamp "abatere", with "stoc sub 0 · …" on the sub-line/detail — in StockReadout and FeedTotals alike.
- **D51 S4 state boards (S4S review).** Decided by: default (user away) — to review. A separate state artboard may show the real farm in a hypothetical state (e.g. Hîrbovăț "fără ghid"); component demos still use "Fermă exemplu" (D38). "Ciclu ales" = the page rescoped to a chosen past cycle (Seria 2025 at Hîrbovăț, real previous-flock data); the open selector is its own board "ciclu · meniu". An empty farm's selector shows "Niciun ciclu deschis" with the off lamp "goală"; past cycles stay menu options.
