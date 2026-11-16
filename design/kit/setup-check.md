# Setup check (16.11.2026)

An audit of `design/kit/` and `design/data/Seria_2026.xlsx` before screens are built. Nothing in the kit was changed except this file and `decisions-current.md`. Scripts used are outside the repo (`/home/claude/scripts/`).

Note: while this audit ran, other files appeared (`kit/canvas.md`, `kit/craft.md`, `design/system/components/*`, `design/canvas/`). They were not audited.

## 0. What the workbook holds

| Sheet | What it is | Real data range |
|---|---|---|
| Telita | Telița daily: mortality per section (H1, H2, H3, H6, H5-1, H5-2), alive per section, totals ♀/♂, "Corectie" (moves between sections and transfers out), "Erori sexare". Row 3 = placement. No cull column. | 07.10.2025 – 28.02.2026 (template dates to 27.04.2026) |
| sapt | Telița weekly roll-up of the above, by age week 1–21 | age wk 1–21 |
| Hrana Telita | Daily: guide feed g ♀/♂ (weekly, "Ghid"), norma g/bird per section, kg per section, deliveries (kg, one column "Gaina"), stock ("Sold"), cost lei per delivery. Feed-type notes in col Y. | 07.10.2025 – 26.02.2026 |
| H1 ( h41,h42), H2, H3 ( h11,h12), H6( h31,h32) | Weekly weight charts, hens (Px): mortality wk/cum %, sexing errors, birds in hall, **Furaj standard g/b/d**, Furaj efectiv (empty), **Greutate Standard**, Efectiv, Diferenta, CV %, uniformity. Name = Telița hall → Hîrbovăț sections. | Standard wk 1–23; actual wk 1–20 (to 24.02.2026). H3 and H6 carry 2024/2025 dates (copied from the previous flock). |
| H51., H52 | Same for roosters (Cx) | Standard wk 1–23; actual to wk 19 (17.02.2026) |
| Hrana Hîrbovăț | Daily template: **Ghid feed g ♀/♂** weekly, norma per section, kg per section H1.1…H4.2, received kg, stock | Guide age wk 19–65 (10.02.2026 – 12.01.2027). Actuals: one row only (10.03.2026, H1.1 508,5 kg ♀ / 48,8 kg ♂). No deliveries. |
| Sheet9 | Hîrbovăț section × sex grid (G11 C11 … G42 C42): arrival counts, forward-filled; col P unlabelled = **% lay guide** | Arrivals 22–28.02.2026; % lay age wk 25–63; template to 18.01.2027 |
| import date | Eggs per hall (Hala 1, 3, 4): Jos, Cat 1–3, Incubabil, Total ou, Masa; mortality per compartment | 04.04.2025 – 28.01.2026 (**previous flock**, not Seria 2026) |
| DataHirbovat | Daily per section: mortality ♀/♂, nutret g, ouat, incubabil, greutate ou (some labels wrong: "H2.1", "H2" inside hall 1/3) | 04.04.2025 – 28.01.2026 (previous flock); 2026 rows empty |
| datefinal | Weekly roll-up of DataHirbovat by age week; egg weight is a weekly sum of daily averages (÷ 7) | age wk 23–64 (previous flock) |

**Last date the workbook covers with real data: 28.02.2026** (one feed row on 10.03.2026). Seria 2026 at Hîrbovăț has no real mortality, eggs, weighings, grams or deliveries after that. Every figure dated after 28.02.2026 (10.03.2026 for feed) in a screen dated 16.11.2026 must be labelled *sample* / illustrative, or shown as waiting / "–". Guide rows and template dates run to 18.01.2027 but are not actuals.

Facts worth keeping:
- Placement 07.10.2025: H1 ♀ 14 080, H2 ♀ 14 000, H5-1 ♂ 3 510 (confirmed).
- Hîrbovăț arrivals (Sheet9): H1.1, H1.2, H3.1, H3.2 ♀ 4 500 · ♂ 498 each; H4.1 ♀ 4 382 · ♂ 534; H4.2 ♀ 4 383 · ♂ 534. Total ♀ 26 765 · ♂ 3 060.
- Telița before transfer (21.02.2026): ♀ 27 381 · ♂ 3 286. Telița after (27.02.2026) is left with residuals 23 / 30 / −9 / 0 / 454 / 244 (a negative count at H3, and ♂ 698 that don't reconcile with ♂ 3 060 at Hîrbovăț).
- 13 deliveries to Telița, 07.10.2025 – 10.02.2026, 263 740 kg, 2 031 768 lei (7,30 – 9,11 lei/kg). None to Hîrbovăț.
- Feed-type notes (Hrana Telita col Y): rost (24.10.2025), razvitie (14.11.2025), norponin (29.11.2025), preouat (20.01.2026). No rooster-only deliveries.
- Hîrbovăț on 16.11.2026: age week 58; feed guide ♀ 158 g · ♂ 155 g; % lay guide 74,1.

## 1. Missing data

Already known as missing in the kit: the Hîrbovăț productivity guide, the "Raport Nou 2026" sheet, incubator data, real guide tables.

**Guide figures that do exist (partial, by age week, unnamed):**
- Growth: weight standard ♀ and ♂ wk 1–23 and feed standard g/b/d wk 1–23 (H-sheets); feed guide ♀/♂ wk 1–25 (Hrana Telita "Ghid").
- Laying: feed guide ♀/♂ age wk 19–65 (Hrana Hîrbovăț "Ghid"); % lay age wk 25–63 (Sheet9 col P, unlabelled, inferred).
- No guide anywhere for: mortality %, hatching %, egg weight, body weight after wk 23 (the whole laying period), CV / uniformity targets.
- All guide rows are indexed by **age week**, not by cycle week (see Open points).

**Also missing (not yet listed in the kit):**

| Missing | Screens that need it |
|---|---|
| Seria 2026 laying actuals at Hîrbovăț after 28.02.2026: daily mortality, culls, eggs by category, egg weight, grams per section, weighings | S2, S3 farm summaries, S4 (readout, stats, tiles, De verificat, latest records, all tabs), S5, S6 (yesterday hint), S7 (yesterday's grams), S8, S11 Consum, S14, S15 |
| Hîrbovăț feed deliveries and stock | S7 totals per type, S11 stock readouts + Livrări, S12, S13, S2 (days left, stock) |
| Feed type list for Hîrbovăț and sex assignment (the kit's "Breeder 1/2", "Starter", "Grower", "Cocoși" are not in the workbook) | S7, S11, S13, S21 |
| Culls (sacrificare) with reasons: no column in any sheet | S4 Păsări, S6, S15, S16 |
| Weighing sample size | S8, S4 Greutate |
| Any Seria 2027 placement at Telița (spec samples assume it) | S3 (two series at once), S2 owner track, S18 |
| Sales of any flock | S10, S3 sale mark |
| Start date of the previous Hîrbovăț cycle (data begins 04.04.2025, mid-cycle at age wk 23) | S4 cycle selector (past cycle) |
| Users and people's names | S1, S22, S17, EventRow meta (illustrative per the kit) |
| Price status of deliveries (the workbook stores lei on every Telița delivery; "fără preț" rows would be illustrative) | S11, S12, S2 owner |

## 2. Mismatches

**components.md vs spec.md / screens.md / data-model.md**

| Component | What components.md says | What the spec / screens say | Screens |
|---|---|---|---|
| AllocationRow | deviation **lamp** on the right | screens S7: a deviation **tag** | S7 |
| AllocationRow / SaveBar | – | S7 needs "Totals per feed type: kg today, stock after, days left": no component | S7 |
| Tag | ≤ 16 characters | the required state "fără date în ghid" is 17 | S4, S5, S7, S8, S20 |
| GuideChart | variants weight, lay, feed g, mortality %, egg weight | also hatching % (D1 Producție), kg/day (Hrană tab), CV / uniformity (Greutate tab), alive/mortality over time (Păsări tab, no guide); no "no guide" (actuals only) or "partial guide" state; no small-curve variant for guide detail | S4, S5, S20 |
| CycleTrack | series' cycles, one lane per farm; `compact` = Acasă owner + farm header | S4 needs one farm's lane with past cycles of **any series**: a farm-scoped variant is not defined | S4 |
| SectionTile | tiles grouped by hall | S4 c also needs a totals row; no hall-group header defined | S4 |
| EntryCard | one section; Ouă group "on the hall card" | spec C1: each card also shows yesterday's mortality as a hint (missing); "hall card" is never defined (data model: eggs on the parent section) | S6 |
| Readout | key figure + guide chart; stock readouts "may repeat" | S11 stock readout needs kg, days left, run-out date and a lamp: no stock variant spec | S11, S2 |
| RecordCode | 11 types | data model `BIRD_EVENT.type` has `adjustment` (correction) and cycles open/close appear in logs and toasts: no icon/colour | S4 Păsări, S16, S17 |
| EventRow | tile, title, meta, lamp/tag; "blocat" tag | S15 needs edit/delete actions, E1 "Corectează" entry point, E2 "modificat de … " line and old → new diff | S15, S16, S17 |
| DataTable | inline edit for owner prices | S20 guide import preview needs per-row parse errors; FileDrop puts row errors in one Banner | S20 |
| FilterBar / SegmentedControl / DateField | period options; single date | period "Interval" needs a date-range picker | S3, S4, S5, S11, S14, S17 |
| Sidebar | "Series selector, then the nav items" | the series selector itself is not specified | all desktop |
| components.md header | "markup and styles are in `components/bundle.css`" | at audit start `design/system/components/` was empty; it is being filled now | – |

**Components a screen needs that don't exist in components.md**
- S1: a login card and a standalone wordmark (the wordmark exists only inside Sidebar).
- S2 / S4 d: an alert row/list for "De verificat" (lamp + record tile + scope + link).
- S2 / C9: a "Lipsesc date" row (section, day, record icon Zilnic or Hrană).
- S2: entry keys (Zilnic, Hrană zilnică, Cântărire …); a "Livrări fără preț" count card.
- S2 owner / S3: a farm summary panel (type, guide, sapt. N, alive, key figures vs guide).
- S7: the feed totals per type block (see above).
- S10 / S18: a per-section ♀ · ♂ count row with "Toate" and add/remove (sale rows, series placement rows).
- S19: a section tree editor (parent / compartments, add, retire).
- S20: a guide import preview with row errors; guide detail small curves.
- S22: a user list (role, farms, last sign-in, active) and a one-time temporary password display.
- S21 / S22: an active / inactive toggle (Checkbox may do; not stated).

**Old-structure leftovers**
- components.md itself has no operator/supervisor, age-ruler or ledger leftovers.
- decisions.md: operator role names; "Rests on: HallEntryCard" (no such component); "Rests on: C4" and "C6, E1" point to spec v1 numbering (C4 is now Vânzare, C6 Prețuri hrană); ledger direction, age ruler, crimson danger, Source fonts, 9-table schema (all dropped in `decisions-current.md`).
- data-model.md and spec.md: "shared by both designs (ledger and control panel)".
- spec A1: "see Stage 4" (stale reference). spec E2: list column "hall" (should be section).
- screens.md: the vet tab bar has "Cont" and the sidebar "the account", other roles have "Mai mult"; no screen covers either.

**Kit samples that contradict the workbook**
- spec C2 sample "H1.1 ♀ 9 402 × 165 g": H1.1 received ♀ 4 500; the guide for age wk 58 is 158 g.
- spec C4 / B3 "Hîrbovăț ♀ 26 881 · ♂ 2 994" and "Telița are încă 26 881 păsări": the workbook arrival total is ♀ 26 765 · ♂ 3 060.
- screens S7 "8 secțiuni": Hîrbovăț has 6 sections × 2 sexes = 12 rows.
- kit-direction "Transfer to Hîrbovăț on 10.02.2026": movements are 22–28.02.2026 (kept as default D21).
- PRODUCT "Hîrbovăț egg data": it belongs to the previous flock (04.04.2025 – 28.01.2026), not Seria 2026.
- "sapt. 41 from 10.02.2026" on 16.11.2026: the formula gives sapt. 40.

## 3. Decisions
Written to `decisions-current.md` (D01–D13 in force from earlier, D14–D23 taken today, superseded list at the end).

## 4. Data extracts to create in `design/data/extracts/`

| File | Source | Range | Screens |
|---|---|---|---|
| `farms-sections.json` | Telita row 2; Hrana Hîrbovăț row 1; Sheet9 row 2; H-sheet names (Telița H1 → H4.1/H4.2, H3 → H1.1/H1.2, H6 → H3.1/H3.2) | static | S3–S7, S9, S10, S19 |
| `placement-seria2026.json` | Telita row 3; H-sheets row 4 | 07.10.2025 | S3, S4 (past cycle), S18 |
| `transfers-seria2026.json` | Telita "Corectie" cols R–W; Sheet9 first non-zero per section | 05.11.2025 – 28.02.2026 | S3 track, S4 Păsări, S9 sample |
| `mortality-telita-daily.json` | Telita cols D–I (mortality), J–O (alive), X–AC (sexing errors) | 07.10.2025 – 28.02.2026 | S3, S4/S5 Telița cycle, S2 |
| `counts-latest.json` | Telita 21.02.2026; Sheet9 28.02.2026 | to 28.02.2026 | S3, S4, S6, S7 |
| `feed-deliveries.json` | Hrana Telita col L (kg), col X (lei) | 07.10.2025 – 10.02.2026 (Telița only) | S11, S12, S13, S2 |
| `feed-daily-telita.json` | Hrana Telita cols F–K (g/bird), N–S (kg), V (stock) | 09.10.2025 – 26.02.2026 | S7 (shape), S11 Consum, S4 Hrană |
| `feed-types.json` | Hrana Telita col Y notes | 24.10.2025 – 20.01.2026 | S7, S11, S13, S21 |
| `weighings-weekly.json` | H1, H2, H51., H52 (weight, CV, uniformity, mortality, standard); H3/H6 re-dated or skipped | age wk 1–20 (to 24.02.2026) | S4/S5 Greutate (Telița cycle), S8, S3 |
| `guide-growth.json` | H-sheets Standard + Furaj standard; Hrana Telita "Ghid" | age wk 1–25 | S20, Telița GuideCharts |
| `guide-laying.json` | Hrana Hîrbovăț "Ghid" (feed ♀/♂); Sheet9 col P (% lay) | age wk 19–65 / 25–63 | S4 readout, S7 guide figure, S14, S20 |
| `eggs-hirbovat-prev-daily.json` | import date (per hall Jos, Cat 1–3, Incubabil, Total, Masa); DataHirbovat (mortality, nutret per section) | 04.04.2025 – 28.01.2026 | S4 past cycle, S14, shape for S6 Ouă and S4 samples |
| `eggs-hirbovat-prev-weekly.json` | datefinal | age wk 23–64 | S14, lay GuideChart |
| `cycles.json` | derived | Telița 07.10.2025 – 28.02.2026; Hîrbovăț from 10.02.2026 (D21); previous Hîrbovăț cycle to 28.01.2026 | S2, S3, S4 |
| `sample-hirbovat-2026.json` | illustrative, labelled sample | 01.03.2026 – 16.11.2026 | S2, S4–S8, S11–S15 |

## 5. Open points that block the first three screens

**S4 · Hîrbovăț Rezumat (desktop)**
1. Week number: 16.11.2026 is sapt. 40 by the formula, sapt. 41 by D23. Default: show sapt. 40 (the formula drives every comparison) and update the copy samples, or move "today" to 17.11.2026 if sapt. 41 must stay.
2. Guide indexing: workbook guides are by age week; the app counts from cycle start (Hîrbovăț cycle wk 1 = age wk 19). Default: import the laying guide re-indexed (cycle wk = age wk − 18) and keep age week out of the UI.
3. No Seria 2026 laying actuals. Default: current-cycle figures from `sample-hirbovat-2026.json`, labelled sample, shaped on the previous flock at the same age; the past cycle in the selector uses the real 2025 data.
4. Weight and mortality guide absent for laying weeks. Default: the Stat shows the actual (sample) with "fără date în ghid"; allow Tag text over 16 characters or use "fără ghid" + a note.
5. Alive counts. Default: start from Sheet9 (♀ 26 765 · ♂ 3 060), minus sample mortality; drop the spec's 26 881 · 2 994.

**S7 · Hrană zilnică (phone)**
1. Feed type names at Hîrbovăț. Default: "Breeder 2" for ♀ and "Cocoși" for ♂, marked illustrative.
2. Rows: 12 (6 sections × ♀/♂), guide ♀ 158 g · ♂ 155 g (age wk 58, real). Default: yesterday's grams and stock are sample; fix the "8 secțiuni" and "9 402" samples.
3. The totals-per-type block has no component. Default: build it from Stat + DataTable totals and add it to `design/system/` and components.md (D16).

**S3 · Seria 2026 cycle track**
1. Transfer date: D21 says 10.02.2026, workbook moves 22–28.02.2026. Default: Hîrbovăț block starts 10.02.2026; one transfer link on 10.02.2026; Telița block closes 28.02.2026 (overlap shown).
2. Telița never reaches 0 in the workbook (residuals, H3 −9). Default: treat the residuals as data errors and close the Telița cycle on 28.02.2026.
3. Seria 2027 at Telița: not in the workbook. Default: S3 shows Seria 2026 only; Telița's lane is empty after 28.02.2026; no Seria 2027 block.
