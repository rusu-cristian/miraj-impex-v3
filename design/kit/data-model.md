# Impex Line: data model v2 (cycles, sections, feed)

Shared by both designs (ledger and control panel). This replaces §0 of the first spec.

## Tables

| Table | Fields | Notes |
|---|---|---|
| `FARM` | `id, name, type, guide_id` | `type`: `growth` or `laying`. `guide_id` is the guide this farm follows; it must have the same type. |
| `SECTION` | `id, farm_id, code, parent_id?, active` | For example H1, H5-1, H1.1. `parent_id` groups compartments under a hall when eggs are counted per hall. |
| `GUIDE` | `id, name, type` | `type`: `growth` or `laying`. |
| `GUIDE_WEEK` | `guide_id, week, sex, feed_g, body_weight_g, mortality_pct?, lay_pct?, hatching_pct?, egg_weight_g?` | General targets for week 1, 2, … of a cycle. |
| `SERIES` | `id, name, year, status` | For example "Seria 2026". The yearly flow: growth → laying → sale. |
| `CYCLE` | `id, series_id, farm_id, guide_id, start_date, end_date?, started_by` | One stay of one series in one farm. `started_by`: `auto` or `manual`. |
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

**Guide week**
- Guide week = `floor((date − cycle.start_date) / 7) + 1`.
- Every chart, report and guide comparison for that farm uses this week.

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

**Percentages**
- Mortality %, % lay and % hatching eggs are relative to birds alive at the start of the cycle or at the start of the day, as the guide defines them.
- All of them are compared with `GUIDE_WEEK` by guide week and sex.
