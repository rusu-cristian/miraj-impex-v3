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
