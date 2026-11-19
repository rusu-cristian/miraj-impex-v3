# KIT03 · Kit update for D182–D185
- **data-model.md:** Tables › GUIDE_WEEK (lay_pct = Productivitate col I, wk 23–62, D184), EGG_PRODUCTION (cat1–3 counts, total per D18/D183), new "Not modelled" line after the table (incubation, D182); Calculated › Percentages (+ Cat. 1–3 %).
- **spec.md:** B2 Laying figures (source + 57,6% sample) and States (paste "645" example, D185); C1 Rules (+ Egg categories, D183); D1 Producție tab (columns, sample) and Rezumat sample ghid 53,3/54,5 → 57,6/58,8; D3 Columns + sample (ghid 57,6/58,8/59,9, Cat %).
- **screens.md:** S4 Other tabs, S6 Ouă group, S14, S20. **components.md:** SYS05 HallCard (+ Categories), PeriodTable › Props `tags` (+ Producție column set).
- **open-questions.md** 8/9/10 resolved (D182/D183/D184, 10 also names D185); 11 open. **setup-check-2.md §5** marked.
- **Producție columns added (Ouat group, after "stare", before incub.):** `% Cat. 1` · `% Cat. 2` · `% Cat. 3` (prefix form as `% ouat`/`% incub.`/`% jos`); % of the row's eggs, one decimal, `–` with no eggs; eggs tint.
- Samples: SAPT 59 0,3/6,2/1,1 % · SAPT 58 0,3/2,6/0,8 %; 18.11 farm (computed, not in DATA06) 0,4/8,9/1,0 %; H5 0,3/10,8/0,7; H6 0,4/5,4/1,5.
- **S6C:** s6_build reads `R.day(d,h)['cat1'..'cat3']` (H5 13·417·27, H6 8·110·31); drop D164 `ec1 = eggs − inc`.
- **S4C:** add the 3 columns to sys14 Producție (S4 tab all Detaliu levels; S14 farm week/day + hall week rows); S14 day rows farm-level only.
- **RF14:** components.md SYS11 "S20N amendments" now reads Productivitate col I + refusal „645” (D185): align sys11/s20 TYPO_RAW/READ and the message text. spec §7 Out of scope now names incubation (D182).
