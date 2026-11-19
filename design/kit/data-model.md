# Impex Line: data model v2 (cycles, sections, feed)

Shared by both designs (ledger and control panel). This replaces §0 of the first spec.

## Tables

| Table | Fields | Notes |
|---|---|---|
| `FARM` | `id, name, type, guide_id` | `type`: `growth` or `laying`. `guide_id` is the guide this farm follows; it must have the same type. |
| `SECTION` | `id, farm_id, code, parent_id?, active` | For example Telița H1, H5-1; Hîrbovăț halls H5, H6, H7 with compartments H5.1 H5.2 H6.1 H6.2 H7.1 H7.2 (round 2, D152; the old H1/H3/H4 codes map to H5/H6/H7). `parent_id` groups compartments under a hall when eggs are counted per hall. |
| `GUIDE` | `id, name, type` | `type`: `growth` or `laying`. |
| `GUIDE_WEEK` | `guide_id, week, sex, feed_g, body_weight_g, mortality_pct?, lay_pct?, hatching_pct?, egg_weight_g?` | Targets by **age week** of the birds (SAPT), not by cycle week: a laying guide starts at its first age week (e.g. 19), with no offset (round 2, D151; D26's "age week − 18" is superseded). `body_weight_g` = weight at the end of that week (week 1 = day 7, D157). `hatching_pct` = % incubabile (hatching eggs of all eggs), not hatchability (round 2, D153). `mortality_pct` = cumulative % (round 2). |
| `SERIES` | `id, name, year, status, age_at_placement_days` | For example "Seria 2026". The yearly flow: growth → laying → sale. `age_at_placement_days` (round 2, D151): the birds' age on the placement day, whole days, default 0 (day-old chicks); "Vârstă la populare (zile)" on Serie nouă. The series' placement date = the date of its first `placement` event (round 2). |
| `CYCLE` | `id, series_id, farm_id, guide_id, start_date, end_date?, started_by` | One stay of one series in one farm. `started_by`: `auto` or `manual`. A cycle covers a span of age weeks (round 2): Seria 2026 at Telița 07.10.2025–28.02.2026 = SAPT 1–21; at Hîrbovăț from 22.02.2026 = SAPT 20, open (SAPT 59 on 18.11.2026). |
| `BIRD_EVENT` | `id, cycle_id, section_id, date, sex, type, qty, transfer_id?, note?, created_by, created_at` | `type`: `placement`, `mortality`, `cull`, `sexing_error`, `transfer`, `sale`, `adjustment`. `qty` is signed. `reason?` (for `cull`): `boală`, `șchiopătare`, `altul` (added 16.11.2026, D36). |
| `FEED_TYPE` | `id, name, for_sex?, active` | For example Starter, Grower, Breeder 1, Cocoși. |
| `FEED_DELIVERY` | `id, farm_id, feed_type_id, date, qty_kg, note?, created_by` · `price_per_kg?, priced_by?, priced_at?` | The manager enters type, date and kg. The owner adds the price later. |
| `FEED_ALLOCATION` | `id, section_id, date, sex, feed_type_id, g_per_bird, created_by` | The vet's daily grams per hen or rooster, per section. |
| `WEIGHING` | `id, section_id, date, sex, avg_weight_g, cv_pct?, uniformity_pct?, sample_size?, created_by` | |
| `EGG_PRODUCTION` | `id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs, avg_egg_weight_g?, created_by` | Recorded on the hall section (the parent) when compartments share an egg count. |
| `USER` | `id, name, login, password_hash, role, active` | `role`: `vet`, `manager`, `viewer`, `owner`. |
| `USER_FARM` | `user_id, farm_id` | Every role except owner is limited to its farms. |
| `AUDIT_LOG` | `id, table, row_id, action, old_values, new_values, reason, user_id, at` | Changes to past data and price entries. |

## Cycle rules

**Auto start**
- When a section of a farm receives birds (placement or transfer in) while the whole farm is at 0, a new `CYCLE` starts.
- The cycle gets that farm's series, its current `guide_id` and `start_date` = that date.

**Manual start**
- The **Începe ciclu nou** button (manager or owner) closes the farm's open cycle and opens a new one from a chosen date.
- Use it when birds arrive in a farm that wasn't fully emptied, or to restart the guide.

**End**
- A cycle ends automatically on the day the farm's birds reach 0, which is after a transfer out or a sale.
- It can also be ended by the manual button.

**Age and week (round 2, D151; replaces the cycle-start "guide week" of D02)**
- `age_days(date) = series.age_at_placement_days + (date − series placement date)`. Day-old chicks are 0 days old on their placement day. (round 2)
- `week (SAPT) = floor(age_days / 7) + 1`: a 13-day bird is SAPT 2, a 14-day bird SAPT 3. (round 2)
- Seria 2026 (placed 07.10.2025, age 0): 18.11.2026 = age 407 = **SAPT 59**; weeks run Tue–Mon (07.10.2025 was a Tuesday). (round 2)
- The week belongs to the birds, not the farm: it carries over when birds move farms (a transfer keeps their age) and a new cycle (auto or manual) does not restart it. (round 2)
- Every chart, table, report and guide comparison uses this week; guides are looked up by it (`GUIDE_WEEK.week` = SAPT). (round 2)
- The workbook's `Zi`/`Zile` = age_days + 1 and its `Sapt` = SAPT, so the app's age is one less than the sheet's day number (18.11.2026: age 407, sheet Zile 408). (round 2)

**Two farms at once**
- A series has at most one open cycle per farm.
- Two series can each be in a different farm at the same time: Seria 2026 laying at Hîrbovăț while Seria 2027 grows at Telița.

**Guide changes**
- Changing a farm's guide applies to new cycles only.
- An open cycle keeps the guide it started with, unless the manager changes it on the cycle.

## Calculated, not stored

**Birds alive** for a section, sex and date
- = `SUM(BIRD_EVENT.qty)` for that section and sex, up to that date.
- Farm totals = the sum of its sections.
- Never below 0.

**Consumption (kg)** for a section, sex and day
- = `FEED_ALLOCATION.g_per_bird × birds alive / 1000`.
- A day with no allocation shows as **missing**. It isn't carried over from the day before.

**Feed stock** for a farm and feed type
- = `SUM(FEED_DELIVERY.qty_kg)` − `SUM(consumption)` for allocations of that type in that farm's sections.
- Days left = stock ÷ average consumption over the last 3 days.

**Feed cost**
- Only for deliveries that have a price.
- Unpriced deliveries show as "fără preț" and are counted separately ("8 000 kg fără preț").

**Sales vs deaths (round 2, D155)**
- Birds that leave a farm to a buyer are `sale` events, never `mortality`; Mortalitate counts real deaths only. Birds alive are the same either way. (round 2)
- Import of the Hîrbovăț register: from 12.11.2026, a compartment-day with ≥ 40 birds removed is a sale (the sheet booked them as mortality): 12–18.11.2026 ♀ 12 801 · ♂ 1 182 sold, real deaths on 18.11.2026 ♀ 3 · ♂ 0. (round 2)
- Mortality % excludes sales. A cycle closes when sales (or transfers out) take the farm to 0. (round 2)

**Percentages**
- Cumulative mortality % = deaths since the cycle start ÷ birds at the cycle start (placed or arrived), per sex (Hîrbovăț to 18.11.2026: ♀ 1 521 ÷ 26 883 = 5,7%). (round 2)
- **Daily lay %** = eggs that day ÷ hens alive at the **start** of the day (round 2, D156): 18.11.2026 5 914 ÷ 14 395 = 41,1% (the sheet's 47,1% divides by the end-of-day count and is not used).
- **Weekly lay %** = the week's eggs ÷ hen-days, where hen-days = Σ hens alive at the **start** of each day of the week, the same basis as the daily % (round 2, D156, D159): SAPT 59 (17–18.11) 11 806 ÷ 29 221 = 40,4%, between its days (39,7% and 41,1%). Never the mean of the daily %, and not the Productivitate sheet's end-of-day count (43,8%, rejected). Farm and hall figures use section sums, not the sheet's Total rows. (round 2)
- % incubabile and % jos = incubabile (or jos) ÷ eggs that day or week. (round 2)
- All of them are compared with `GUIDE_WEEK` by age week (SAPT) and sex. (round 2)
- **Weighing vs guide (round 2, D157):** a weighing is labelled with its SAPT but compared with the guide row of the week just ended (`GUIDE_WEEK.week` = SAPT − 1), because a guide weight is the weight at the end of its week: 24.02.2026 (age 140, SAPT 21) Telița H1 ♀ 2 553 g vs guide week 20 2 370 g = +7,7%, abatere.
