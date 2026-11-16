# Decisions in force (16.11.2026)

Built from `decisions.md`, `spec.md` v2 and `data-model.md` v2. Latest wins. Entries that a later decision or spec v2 overturned are listed at the end, not kept.

## In force

**D01 · Bird counts come from events, not stored counts**
- Decided by: user
- One `BIRD_EVENT` ledger (placement, mortality, sexing error, transfer, sale, adjustment). Counts can't drift; a transfer is one pair of rows.
- Rejected: one entity per kind of change; a daily wide row per hall like the spreadsheet.

**D02 · Data model v2: farm types, sections, cycles, feed types, priced-later deliveries, daily grams**
- Decided by: user
- Each farm has a type (growth / laying), sections and a linked guide of the same type. The guide week is counted from the farm's cycle start: `floor((date − cycle.start) / 7) + 1`. The yearly series moves growth → laying → sale. 15 tables as in `data-model.md`.
- Rejected: the guide week counted from the series placement day; the norm-until-changed model; the price entered with the delivery; the 9-table schema.

**D03 · Roles: vet / viewer / manager / owner**
- Decided by: user
- Every role except owner is limited to its farms (`USER_FARM`).
- Rejected: operator / supervisor / manager / owner.

**D04 · A fixed record-type code (icon + colour)**
- Decided by: user (request); default (the hues)
- hen #B0246A, rooster #1F5BB5, mortality #3A424A, cull #7C2F86, placement #B4450F, transfer #5B3CC4, sale #56680C, feed #7F5C00, weighing #0A6A8F, eggs #0D6F69. Sexing error is a split hen/rooster tile. Status never uses a record colour.
- Rejected: grey for placement.
- Known risk: rooster vs weighing and cull vs transfer are close; icon and word carry the difference.

**D05 · Feed price lives only on deliveries, added later by the owner; consumption and stock are in kg**
- Decided by: user
- The manager enters type, date and kg; the owner adds the price. Cost counts priced deliveries only; unpriced kg are shown apart ("fără preț").
- Rejected: weighted-average or FIFO costing; the price entered with the delivery.

**D06 · Vets don't see feed prices; viewers and managers see them read-only**
- Decided by: default (approved in spec v2)

**D07 · Vets can change or delete their own entries until 23:59 the same day; past data needs a manager or the owner, a reason and an `AUDIT_LOG` entry**
- Decided by: default (approved in spec v2)

**D08 · Online only; a form keeps its draft in the browser until saved**
- Decided by: user (online only); default (drafts)
- Rejected: offline sync.

**D09 · The Seria 2026 history is imported once by the developer**
- Decided by: default (approved in the spec)
- Rejected: an import screen.

**D10 · Design system: Impex Line Panel (the barn control panel), a new system**
- Decided by: user
- Light steel panels, IBM Plex Sans and Mono, graphite keys, status lamps with words, dark Readouts. Not the 15.11.2026 walnut/abacus system, and not the ledger direction.
- Rejected: reusing the 15.11.2026 system; the farm ledger; eggshell and straw; a veterinary chart.

**D11 · Panel: light theme only, mixed density; green means "normal / current", never a record type; primary keys are graphite**
- Decided by: user (light, mixed density); default (the meaning of green)
- Rejected: light and dark; a green primary button.

**D12 · Panel signature: the cycle track**
- Decided by: user
- Rejected: a section map; a guide gauge; the age ruler.

**D13 · Vet feed entry: yesterday's grams are a suggestion that must be confirmed**
- Decided by: default
- An unconfirmed row isn't saved and the section shows as missing.

## Taken today (16.11.2026)

**D14 · Repository: rusu-cristian/miraj-impex-v3, branch main**
- Decided by: user

**D15 · Canvas mode: screens are artboards on a Design canvas built from the approved Impex Line Panel design system; the system is not redesigned**
- Decided by: default (user away) — to review

**D16 · The design system artifact can't be edited from this session**
- Decided by: default (user away) — to review
- Any new or changed component is added to the repo copy in `design/system/` and to `components.md`, and listed for copying into the system artifact.

**D17 · A manual new cycle (Începe ciclu nou) is allowed while birds remain in the farm**
- Decided by: default (user away) — to review

**D18 · Total eggs = Cat. 1 + Cat. 2 + Cat. 3 + Incubabile; "Jos" is kept separate**
- Decided by: default (user away) — to review
- Supported by the workbook: in "import date", Total ou = Cat 1 + Cat 2 + Cat 3 + Incubabil on all 900 hall-days, Jos excluded.

**D19 · Alert thresholds**
- Decided by: default (user away) — to review
- Mortality above 3× the section's 7-day average; weight ±5%; lay ±3 points; feed ±10%; stock under 5 days; missing entries shown from 12:00.

**D20 · Copy uses the informal "tu"**
- Decided by: default (user away) — to review

**D21 · Transfer to Hîrbovăț on 10.02.2026, so the laying cycle starts at sapt. 1 that day**
- Decided by: default (user away) — to review
- Note: the workbook's bird movements (Telita "Corectie", Sheet9) show the transfer on 22–28.02.2026; 10.02.2026 is the first row of "Hrana Hîrbovăț" (age week 19).

**D22 · Guide names are illustrative ("Ross 308 ouat", "Ross 308 creștere"); guide figures are shown only where the workbook has them, otherwise the "fără date în ghid" state**
- Decided by: default (user away) — to review

**D23 · Today in screens is 16.11.2026; Hîrbovăț is in sapt. 41 of its cycle from 10.02.2026**
- Decided by: default (user away) — to review
- Note: by the D02 formula, 16.11.2026 is sapt. 40 of a cycle from 10.02.2026 (day 279); sapt. 41 starts 17.11.2026. See `setup-check.md` › Open points.

## Superseded and dropped
- 9-table schema (+ users, user farms, audit log): superseded by data model v2 (D02).
- Direction: the farm ledger (warm paper, graphite, barn red): superseded by the Panel (D10).
- Signature: the age ruler: superseded by the cycle track (D12) and the cycle-based guide week (D02).
- Light theme only; mixed density (ledger version): restated for the Panel (D11).
- Danger is crimson #9E1F3F: superseded by the Panel `fault` #B0261E.
- Fonts: Source Serif 4 / Source Sans 3: superseded by IBM Plex Sans / Mono (D10).
- Second design "to compare with the ledger": now the approved system (D10, D15).
- "Operators don't see feed prices": role name superseded; restated as D06.
- "Operators can edit their own entries until 23:59": role name superseded; restated as D07.
- Total eggs (old entry, "rests on HallEntryCard"): restated as D18; HallEntryCard no longer exists.
- Alert thresholds (old entry, per "hall", no feed ±10%): replaced by D19.
- "New design system, not walnut/abacus": merged into D10.

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
