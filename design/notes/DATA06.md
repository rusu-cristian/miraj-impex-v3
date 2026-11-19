# DATA06 · Guide lay % from Productivitate (D184) · Cat. 1–3 counts (D183)

**Changed.** `data03_build.py` takes the lay % guide from Productivitate col I (wk 23–62; col A = age week per D151, its hen-days/eggs equal the extracts' per-SAPT sums; rows 63–67 repeat 53,7 as a template and are not taken). Mortality and eggs/hen guides stay RealGhid. It reads Cat. 1–3 per hall and day from `Raspunsuri` (the form). Rebuilt: guide-laying.json (`lay_pct`; `alt_realghid` replaces `alt_productivitate`), hirbovat-daily.json (`guide_lay_pct`; `cat1–cat3` on hall_day and farm_day; `meta.cat_form_unused`), hirbovat-weekly.json (`guide.lay_pct`, `lay_vs_guide_pts`, `guide.lay_pct_realghid`; `cat1–3` and `cat*_pct` = % of eggs per farm and hall), key-figures-18-11.md, README. `data03_check.py` asserts both (checks: OK).
**D18:** Total ouă = Cat 1 + 2 + 3 + Incubabile holds on all 757 hall-days with eggs (10.03–18.11), 0 mismatches. Masa's Cat % = these counts ÷ eggs (0 differences). 4 of 258 form submissions are not used: 2 with the year typed as 2027, one 23.06 entry that is really 23.08, one duplicate. The used one per day matches the Hirbovat sheet.

**Old → new on boards**
- Guide lay % today, SAPT 59: **53,3 → 57,6 %**. 18.11 farm 41,1 % vs guide: −12,2 → **−16,5 pts**; H5 45,3 %: −8,0 → −12,3; H6 34,9 %: −18,4 → −22,7.
- Weekly guide SAPT 56/57/58/59: 56,9/55,7/54,5/53,3 → **61,1/59,9/58,8/57,6**. SAPT 59 lay 40,4 % vs guide: −12,9 → −17,2 pts. Peak SAPT 30 guide 86,7 → 88,1. Verdicts above +3 move into ±3 ("abatere" → "în normă"): farm SAPT 30 +3,65 → +2,25 and SAPT 31 +4,27 → +2,67. By hall: 27 H5, 30 H5/H6/H7, 31 H6/H7, 32 H7, and 56 H6 +5,11 → +0,91. Days in the default 4 weeks: 31.10 H6 → în normă; 01.11 H7 → în normă; 01.11 farm −0,8 → −5,0 and H5 +0,6 → −3,6 → abatere. 86 hall/farm-days flip over the cycle, mostly in SAPT 27–32.
- Lay alert "sub ghid din": H5 02.11 → **01.11**; H6 stays 01.11.
- Zilnic Ouă 18.11 (replaces the D164 placeholder H5 457 / H6 149 in Cat. 1): **H5 Cat. 1 13 · Cat. 2 417 · Cat. 3 27**; **H6 8 · 110 · 31** (+ inc 3 415 / 1 893 = 3 872 / 2 042).
- Reports Cat % of eggs, farm (Cat 1/2/3): SAPT 58 0,3/2,6/0,8 % · SAPT 59 0,3/6,2/1,1 %.

**Generators**
- These read guide lay % from the extracts through realdata.py (`guide_lay`, `day()['guide_lay_pct']`, `week()['guide']`, `lay_below_since`/`alerts`), so they pick up the new values on rerun: s2, s2s, s4s, s6, s15, s16s, sys14 (and s4t, s14 through sys14/realdata). s6_build has the comment "RealGhid; 53,3 at SAPT 59", and the docstrings of s14/s20/s20s/sys11 still say "RealGhid".
- **Boards showing these verdicts (RF14):** sys14 PeriodTable Producție "Stare" lamp (|lay − guide| > 3 → abatere), used by S4 Producție/Rezumat (s4s), S14 (s14), S3 Serie, Secțiune. The default 4 weeks show week 56 H6 and the 31.10/01.11 days (Hale/Compartimente). Weeks 27–32 show only under "Arată toate" / whole-cycle views. The s4s/s4t Producție chart's ±3 band now holds the SAPT 30–31 peak. s6, s15 and s2 verdicts are for 18.11 only (still abatere, wider gap).
- **Will break on rerun:** sys11_build (`laying_typed_rows` expects wk 53 lay = 60,5 from RealGhid's "605" typo) and s20_build (via sys11 helpers, plus its assert at l.336). With D184, wk 53 = 64,5 and has no typo. Needs a decision (below).
- Cat: only **s6_build** shows Cat values, and it hard-codes D164 (`ec1 = eggs − inc`, Cat. 2/3 null). It should read `R.day(d, h)['cat1'..'cat3']`. s4s l.808 is only a legend text. No S4 Producție or S14 table shows Cat % yet (D183 asks for it).

**D185 (user):** the S20/SYS11 paste example = Productivitate col I with week 53 typed "645", refused. No data change; RF14 updates sys11/s20 (TYPO_RAW/READ) to fix the break above.
