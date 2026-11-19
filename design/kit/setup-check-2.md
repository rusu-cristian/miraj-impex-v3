# Setup check 2 — new workbook `Raport_NOU_2026.xlsx` + "table first" request (19.11.2026)

Read-only audit (scripts in /tmp/audit, `python3 -I`, openpyxl read_only/data_only). No kit file edited.

## 1. New workbook (18 sheets; all date templates run to 2027, real data ends 18.11.2026)
| Sheet | Size | Holds | Grain · real range |
|---|---|---|---|
| `Hirbovat` (= `BD Hirbovat`, `Telegram BOT` copies) | 410×56 | Seria 2026 at Hîrbovăț: mortality + alive per compartment × sex, eggs per hall and farm | daily + weekly `Total` rows · 17.02–18.11.2026 (H7 to 13.11) |
| `Hrana Hirbovat` | 360×38 | guide g ♀/♂ (weekly), norm g/bird per compartment × sex, deliveries kg ♀/♂ feed + lei, kg per compartment × sex, stock ♀/♂ | daily · 17.02–19.11.2026 (deliveries to 13.11) |
| `%%` | 410×15 | % lay, % incubabile, % jos per farm and hall | daily + Total |
| `Masa` | 410×30 | egg weight, Cat 1/2/3 %, incubabile % per farm and hall | daily + Total · to 18.11 |
| `Productivitate` | 50×13 | per age week 23–58: hen-days, avg hens, eggs week/cumul, hatching eggs, % lay real vs ghid vs dif, hatching eggs/hen real vs ghid | weekly |
| `RealGhid` | 46×35 | laying guide by age week 23–62 (% lay, mortality %, eggs/hen cumul, hatching eggs/hen) + plan vs real per week (from 10.03.2026) + hatchability guide | weekly |
| `Incubare` | 161 rows | incubator batches 1–118: eggs set, set date, hatch date, chicks, hatch % vs guide by age week, egg age per hall (labelled H1/H3/H4) | per batch · 20.03–22.11.2026 |
| `BD Incubator` | 542×9 | incubation history (set, transfer, hatch, chicks, %, hall H5/H6/H7) | per batch · 2022–12.2026 |
| `Raspunsuri` | 259×34 | Google-Form entries: per compartment C/G mortality, per hall Jos/Cat1/Cat2/Cat3/Incubabil/Masa | daily · 10.03–18.11.2026 |
| `Telita`, `Hrana Telita` | 209×30, 177×26 | as before; data identical by date to the old file; new `Sapt`/`Zi` columns, guide split G/C, rows shifted | daily · 07.10.2025–27.04.2026 |
| `Total Telita` | 28×15 | weekly Telița mortality ♀/♂ and cumulative % | weekly 1–21 |
| `TG Notif`, `verificare`, `Foaie10` | – | Telegram summary template, check sheet (zeros), empty | – |

**The "Excel habit" (`Hirbovat`):** one row per day (`Data · Zile · Sapt`), a `Total` subtotal row after each Monday (weeks Tue–Mon = age weeks). Column groups left→right: **Total Fermă** [♂ mort · ♂ în viață · ♀ mort · ♀ în viață · Ouă · % · Incubabile · % · Jos · %], Total incubabil (cumul), then **Hala 5 | Hala 6 | Hala 7**, each = 2 compartments × [♂ mort · nr · ♀ mort · nr] + hall [Ouă · % · inc · % · jos · %]. Total rows carry week sums and % (lay %, cumulative mortality %). Feed, egg weight/categories and % live in sibling sheets with the same Data/Zile/Sapt spine.

## 2. Old vs new
- **Real data now to 18.11.2026** (old: 28.02.2026). Hîrbovăț sections renamed **H5.1 H5.2 H6.1 H6.2 H7.1 H7.2 (Hala 5/6/7)**; old H1.1…H4.2 map by position (H1→H5, H3→H6, H4→H7; DATA01 had flagged the H5.1G… row-2 labels as "wrong" — they were right). No new farm or series; Telița unchanged.
- Arrivals in `Hirbovat`: ♂ 3 024 on 22.02 (498 ×5, H7.2 534), ♀ 26 883 on 25–28.02 (4 500 ×5, H7.2 4 383) vs extract ♀ 26 765 · ♂ 3 060 (D28). `Hirbovat`/`Hrana` start 17.02.2026 (old Hrana Hîrbovăț started 10.02 = D21/D25's start).
- **Story change:** lay fell from ~73% (wk 55) to ~26–28% (wk 57, 03–09.11); depopulation from 12.11: Hala 7 empty 14.11, H6.2 almost empty 18.11, H6.1 started 18.11. Alive 16.11 ♀ 14 826 · ♂ 1 227 (sample: ♀ 25 252 · ♂ 2 685); 18.11 ♀ 12 561 · ♂ 794. Removals are booked as mortality (as the previous flock, D92).
- Gone from the new file: `Sheet9`, H-sheets (`H1 ( h41,h42)`, `H2`, `H3…`, `H6…`, `H51.`, `H52`), `sapt`, `import date`, `DataHirbovat`, `datefinal`. **No weighings at all** in the new file.
- Extracts: **stale** → `cycles.json`, `transfers-seria2026.json` (Hîrbovăț side), `farms-sections.json` (Hîrbovăț codes), `feed-deliveries.json` (+ real Hîrbovăț deliveries, lei), `guide-laying.json` (real lay % / mortality now), `counts-latest.json`, `sample-hirbovat-2026.json`. **Still valid** → `placement-seria2026.json`, `mortality-telita-daily.json`, `feed-daily-telita.json` (identical by date), `feed-types.json`, `guide-growth.json`, `weighings-weekly.json` and `eggs-hirbovat-prev-*` (sources only in the old file — keep it as the source for those).
- Scripts: `data01_build.py` hardcodes `Sheet9`, `Hrana Hîrbovăț`, H-sheet names, `HIR = H1.1…H4.2`, fixed rows → needs a new Hîrbovăț reader (rewrite that half; Telița half works if rows are found by date). `data02_prev.py` reads `import date`/`datefinal` → keep pointed at the old file. `data02_sample.py` → retire for 17.02–18.11.
- **sample-hirbovat-2026.json is replaceable** for mortality, alive, eggs (total/inc/jos, Cat 1–3, egg weight), g/bird and kg per compartment × sex, deliveries + lei, stock. Not in the workbook (keep illustrative or drop): culls + reasons, sexing errors, weighings, event log/users/audit.

## 3. Bird age and week
- The workbook states age per row: `Zi`/`Zile` = 1 on the placement day 07.10.2025 (day-old chicks), `Sapt` = ceil(Zi/7). Equivalent to the user's rule `floor(age_days/7)+1` with age_days = date − 07.10.2025 (age 0 on arrival): e.g. 21.10.2025 = age 14 → Sapt 3 ✓; 16.11.2026 = age 405 → 58 ✓; 17.11 = 406 → 59 ✓. No explicit "age at placement" field. Confirm with the user that a chick is 0 days old on its placement day (if they read `Zi` as age, Zi 14 = Sapt 2 in their own file).
- Open cycle Hîrbovăț Seria 2026: 16.11.2026 → user rule **58** (workbook 58) · D02 from 10.02 **40** · D02 from first arrival 22.02 **39**. 19.11.2026 → user **59** (workbook 59) · D02/10.02 **41** · D02/22.02 **39**. D02 + 10.02 = age week − 18 (D26), so only the label changes, not the guide row — but the guides in the file are indexed by age week, so D26's offset can go.
- Telița (closed): D02 from placement = user rule. **D95:** 24.02.2026 = age 140 → **SAPT 21** (workbook `Telita` Sapt 21), confirming D95's label/guide 2 525 g; the old H1 sheet's wk-20 comparison isn't in the new file. Residual question only if the guide's body weight means "at end of week N" (then day 140 = end of wk 20).
- Impact: D02, D23/D24, D26, data-model "Guide week", SERIES/CYCLE need an age at placement; ~153 boards print "sapt./SAPT 40"; ~34 boards print ♀ 25 252.

## 4. Where the table goes
- Data today: S4 tabs Rezumat (default) · Păsări · Hrană · Greutate · Producție; S3 Serie tabs Păsări · Hrană · Greutate · Producție (no FilterBar, D63); S5 same tabs per section. S4 FilterBar: period (Ciclul curent · 4 săpt. · 13 săpt. · Interval), grain Zi/Săpt., RecordChips **Mortalitate · Sacrificare · Eroare sexare · Transfer · Hrană · Cântărire · Ouă** (spec also lists Populare, Vânzare), Exportă .xlsx.
- S14 table: week groups (SAPT n + dates) with a farm row then hall rows; columns Găini · Ouă · Ouat (Efectiv/Ghid/Față de ghid/Stare) · Incubabile (Ouă/%) · Masă ou · Mortalitate ♀/♂ · Hrană ♀ g/pasăre (Efectiv/Ghid); SAPT 33–40, export.
- **(a) "Tabel" first tab on S4 (and S3, S5):** S14 grouping, week row = subtotal (like Excel `Total`), expands to 7 day rows; grain Zi = all expanded; RecordChips choose column groups (Mortalitate → ♀/♂ mort + în viață, Ouă → ouă/%/inc/jos/masă, Hrană → g + kg …). Rezumat becomes tab 2. ~10–12 boards (S4 default, expanded, chip subset, past cycle, phone, 3–4 states; S3 ×2; S5 ×1) + 1 DataGrid component.
- **(b) Extend S14 into "Rapoarte · Tabel"** (all record groups, daily rows, chips, growth farm too): ~6–8 boards; farm pages stay chart-first, so the habit isn't met where people land.
- **(c) Excel mirror on S4 Tabel:** halls as column groups (Total fermă | Hala 5 | 6 | 7) with daily rows and inline weekly Total rows, hall chips + record chips; closest to the habit, wide scroll; ~8–10 boards.

## 5. Doesn't fit data-model.md
- Incubation batches (`Incubare`, `BD Incubator`): set/hatch dates, eggs set, chicks, hatch % vs guide, egg age per hall — no table. *Out of scope (D182).*
- Age at placement / age-week guides (see 3); GUIDE_WEEK lacks eggs/hen (week, cumul), hatching eggs/hen, hatchability; `hatching_pct` is ambiguous.
- Two lay% guides disagree (RealGhid col B wk 56 56,9% vs Productivitate col I 61,1%; RealGhid wk 53 typo "605"). *Resolved: Productivitate col I (D184).* The extract's "illustrative lay_pct" (D29) is the hatchability guide shifted one week (Incubare wk 55 76,5 · 57 74,1).
- Deliveries carry total lei and a ♀/♂ feed split, not a feed type or lei/kg; depopulation booked as mortality (needs sale/transfer mapping); egg categories Cat 1–3 come as counts (form) and % (Masa). *Categories resolved: counts on Zilnic, % in reports (D183).*
