# Impex Line: functional spec v2 (shared by both designs)

Status: 16.11.2026, round 2 updates 19.11.2026 (today on the boards = 18.11.2026, D152). This replaces spec v1.
- **Data model:** `data-model.md`.
- **Interface:** Romanian with full diacritics.
- **Sample values:** from the Seria 2026 workbook, marked *sample*. Names of people and later-dated figures are illustrative.

## 1. Shared rules

**Roles**

| Role | Enters | Sees | Edits past data | Scope |
|---|---|---|---|---|
| `vet` | daily mortality, culls, sexing errors, daily feed grams, weighings, eggs | today's entry screens and "Introdus azi" | own entries, same day only | assigned farms |
| `viewer` | – | all views and reports, no prices | – | assigned farms |
| `manager` | everything the vet enters, plus feed deliveries (no price), transfers, sales, new cycle | all views and reports; prices read-only | yes, with a reason | assigned farms |
| `owner` | everything, plus feed prices, users, farms, guides | everything | yes, with a reason | all farms |

**Cycles and the week (round 2, D151)**
- Every week in the app is the birds' **age week** (SAPT): `age_days = age at placement + (date − placement date)`, `SAPT = floor(age_days / 7) + 1`. Day-old chicks are 0 days old on their placement day. (round 2)
- Seria 2026 (placed 07.10.2025, age 0): 18.11.2026 = age 407 = SAPT 59; its weeks run Tue–Mon (SAPT 59 = 17.11–23.11.2026). (round 2)
- The week carries over when birds move farms and doesn't restart with a new cycle; guides are read by SAPT. (round 2)
- A farm follows its linked guide during its current cycle.
- A cycle starts automatically when birds arrive in an empty farm, or with **Începe ciclu nou**.
- A cycle ends when the farm reaches 0 (by sale or transfer out; emptying is a sale, D155).
- See `data-model.md` › Cycle rules.

**Birds alive**
- Birds alive is the sum of bird events, per section and sex.
- No save may make a section go below 0 on any day.

**Feed**
- Daily kg = grams × birds alive. There's one entry per section, sex and day; a missing day is flagged, not carried over.
- Stock is kept per farm and feed type.
- Cost counts priced deliveries only. Unpriced kg are shown separately.

**Guide comparison**
- Comparisons are by age week (SAPT) and sex; a weighing is compared with the guide of the week just ended, SAPT − 1 (round 2, D157).
- Tolerance: weight ±5%, lay ±3 points, feed ±10% `default`. Mortality is flagged when it's above the guide.

**Record-type code**
- Each record type has a fixed icon + colour, used everywhere (chips, rows, chart lines, filters, legends):
  - ♀ găini, ♂ cocoși
  - mortalitate, sacrificare, eroare sexare
  - populare, transfer, vânzare
  - hrană (daily: wheat icon · delivery: package icon · price: coins icon)
  - cântărire, ouă
- Status (în normă / abatere / refuzat / informație) is always shown separately, as a word with its own marker (a lamp or a tag). It's never shown with a record colour.

**Formats**
- Numbers: `14 080` (thin no-break space), `2,41 kg`, true minus `−82`.
- Unknown values: `–` with a reason.
- Dates: `dd.mm.yyyy · sapt. 59` (the birds' age week, round 2). In tables the day carries its age in days too, column "Vârstă (zile)": `18.11.2026 · 407 · SAPT 59` (round 2, D159).

**Edits**
- Vets can change or delete their own entries until 23:59 the same day.
- Managers and the owner can change past data with a reason. Every change is written to `AUDIT_LOG`.

**Connectivity:** online only. A form keeps its draft in the browser until it's saved, and offers retry on failure.

**Copy voice**
- Informal "tu" `default`. Verbs on buttons. Errors name the section, the number and the fix.
- No exclamation marks or emoji.

## 2. Access
- **A1 · Sign in:** see the appendix.
- **A2 · Users and access (owner):** see the appendix, with the roles vet / viewer / manager / owner.

## 3. Setup (owner; manager for their farms where noted)

### B1 · Farms and sections — MVP
- **Job:** When we set up a farm, I want to give it a type, its sections and the guide it follows, so each cycle is compared with the right targets.
- **Entry points:** Setări › Ferme.
- **Fields:**
  - farm name
  - type: Creștere or Ouat
  - guide: must have the same type
  - sections: a code, with an optional parent for compartments
- **Rules:**
  - The guide must match the farm's type: "Ghidul „Ross 308 ouat” este pentru ouat. Ferma Telița este de creștere."
  - A section with birds can't be retired.
  - Changing the guide applies to the next cycle. The current cycle is changed only from the cycle panel (B4).
- **States:**
  - no farms: "Adaugă prima fermă."
  - farm without a guide: Tag "fără ghid". Its charts show actuals only.

### B2 · Guides — MVP
- **Fields:** name, type (Creștere / Ouat), and weekly rows (week × sex: feed g, weight g, mortality %, lay %, hatching %, egg weight g), imported by paste or .xlsx/.csv.
- **Weeks are age weeks (round 2, D151):** the `Săpt.` column of a guide is the birds' age week (SAPT), the same week every screen shows. A laying guide starts at its first age week, not at 1: "Ross 308 ouat" covers SAPT 19–65 (lay % from SAPT 23), "Ross 308 creștere" SAPT 1–25. No offset is applied (D26's "SAPT = vârstă − 18" is gone, also from the Ghiduri detail note). (round 2)
- **Weight rows (round 2, D157):** a guide weight is the weight at the end of its week (week 1 = day 7); a weighing labelled SAPT n is compared with guide week n − 1. The detail table says it in its foot note: "Greutatea din ghid = la sfârșitul săptămânii. O cântărire din SAPT n se compară cu rândul n − 1." (round 2)
- **Laying figures (round 2):** % ouat and cumulative mortality % come from the workbook's RealGhid sheet (the guide the app uses; the Productivitate curve is not imported, see open questions). % incubabile is hatching eggs of all eggs, not hatchability (D153). *Sample:* SAPT 59 ♀ 157 g · ♂ 156 g, % ouat 53,3%.
- **Rules:**
  - Duplicate week + sex rows are refused.
  - A guide used by a cycle can be replaced but not deleted.
  - A laying guide whose first row is week 1 imports with a warning (saveable) on that row: "Săptămâna 1 e prima săptămână de viață. Ghidurile de ouat încep de obicei la SAPT 19–23; verifică dacă fișierul numără săptămânile de la transfer." (round 2)
- **States:** parse errors per row; a partial guide shows `–` with "fără date în ghid" for every SAPT outside its rows (Hîrbovăț's laying guide has no body weight; growth weights end at SAPT 25).

### B3 · Series — MVP
- **Job:** When a new year's flock arrives, I want to start a series and place the birds into the growth farm's sections.
- **Flow:**
  1. Name and year (*sample:* Seria 2027 · 2027).
  2. Placement: the growth farm, date, **Vârstă la populare (zile)** (round 2), then rows of section · ♀ · ♂ (*sample:* H1 ♀ 14 080 · H2 ♀ 14 000 · H5-1 ♂ 3 510).
  3. Creează seria.
- **Vârstă la populare (zile)** (round 2, D151):
  - CountInput, whole days, **0–140**, default **0** (day-old chicks). 140 days = 20 weeks, the age Seria 2026 left the growth farm; older birds would skip the growth guide.
  - Hint: "0 pentru pui de o zi. Săptămâna păsărilor (SAPT) se numără de la această vârstă."
  - Live line under the field (Mono figures), from the date and the age: "La populare: vârstă 0 zile · SAPT 1." (21 → "vârstă 21 de zile · SAPT 4").
  - Errors, on blur and on save: emptied field "Scrie vârsta la populare în zile, 0 pentru pui de o zi."; out of range "Vârsta la populare e 160 de zile. Scrie o valoare între 0 și 140."; not a whole number "Scrie un număr întreg de zile, de exemplu 0." Each is a refusal (Creează seria disabled until fixed).
  - Later change: owner only, from the series in Serii, with a reason (E1, `AUDIT_LOG`). The Dialog says what moves: "Toate săptămânile Seriei 2026 se renumerotează (0 → 7 zile: azi SAPT 59 → SAPT 60) și se recompară cu ghidurile." (round 2)
- **Weeks in the series list (round 2):** each series row shows its SAPT span and today's age: open "SAPT 1–59 · azi 407 zile" (Seria 2026), closed "SAPT 19–65" (Seria 2025, start inferred); each cycle row shows its own span and dates: "Telița · 07.10.2025–28.02.2026 · SAPT 1–21", "Hîrbovăț · din 22.02.2026 · SAPT 20–59 · deschis". A column "Vârstă la populare" shows "0 zile".
- **Result:** the placement events, plus a new cycle in the growth farm (auto) with its guide.
- **Rules:** if the farm still holds birds from another series: "Telița are încă 26 881 păsări din Seria 2026. Mută-le sau vinde-le înainte, sau începe un ciclu nou manual."
- **Closing:** Închide seria (owner) after its last cycle ends. A closed series is read-only.

### B4 · Cycles — MVP (manager, owner)
- **Job:** When birds arrive or a farm restarts, I want the farm to follow its guide at the birds' age week, so comparisons line up. (round 2, D151: a cycle doesn't restart the week.)
- **Entry points:** a farm page › Ciclu (cycle panel); the cycle track (click a cycle block).
- **Automatic:**
  - A cycle opens when birds arrive in an empty farm. The Toast reads "Ciclu nou la Hîrbovăț · Seria 2026 · sapt. 20 de la 22.02.2026 · ghid Ross 308 ouat." (round 2: the first arrival, ♂ 22.02.2026, age 138 = SAPT 20)
  - It closes when the farm reaches 0, by sale or transfer out (D155).
- **Manual:**
  - **Începe ciclu nou:** a dialog with the date (default today), the guide (default the farm's) and the note "Ciclul curent (sapt. 20–59) se închide pe 18.11.2026. Săptămâna păsărilor continuă: ciclul nou începe la sapt. 59." It requires a reason. (round 2, D151/D152: the week is the birds' age, it doesn't restart)
  - **Schimbă ghidul ciclului:** applies from the cycle start, so charts are recompared by SAPT (round 2).
- **Rules:**
  - One open cycle per farm.
  - A cycle can't start before the farm's last event of the previous cycle.
- **States:**
  - farm empty, no cycle: "Ferma e goală. Ciclul pornește la prima populare sau transfer."

## 4. Entry

All forms work at 360px width. Each one shows the farm, the date and the birds' week (SAPT, round 2), and has a sticky Salvează.

### C1 · Zilnic (vet) — MVP
- **Job:** After the morning round, I want to enter what happened in every section in one go.
- **Flow:**
  1. Farm, then date (Azi / Ieri / alege).
  2. One card per section with birds. Each card shows alive ♀ · ♂ and yesterday's mortality as a hint.
  3. Fields per sex present:
     - **Mortalitate**
     - **Sacrificare:** requires a reason (boală, șchiopătare, altul)
     - **Eroare sexare:** under "Mai mult"
     - at a laying farm, on the hall's card: **Ouă** (Jos, Cat. 1–3, Incubabile, Masa medie)
  4. "Fără pierderi azi" fills 0 for that section.
  5. Salvează. The Toast reads "Salvat · 6 secțiuni · ♀ −23 · ♂ −5".
- **Rules:**
  - Empty is null, not 0.
  - A second save replaces the earlier entry for that section, day and type.
  - **Warning:** mortality above 3× the section's 7-day average.
  - **Block:** a count above birds alive.
  - Total eggs = Cat. 1 + 2 + 3 + Incubabile; Jos is kept separate `default`.
- **States:**
  - loading
  - already entered (Editează)
  - past day for a vet: "Pentru zilele trecute cere managerului."
  - save error with retry
  - farm without birds

### C2 · Hrană zilnică (vet) — MVP
- **Job:** Each day, I want to enter how many grams each hen and rooster gets in every section, so consumption and stock are calculated.
- **Flow:**
  1. Farm, then date.
  2. A grid with one row per section × sex present. Each row shows:
     - birds alive
     - the feed type (default: the last one used in that section)
     - g/pasăre: prefilled with yesterday's value as a suggestion, highlighted until confirmed
     - kg calculated live
     - the guide figure for this week
  3. **Copiază ieri** fills every row with yesterday's grams.
  4. The total kg per feed type and the remaining stock are shown.
  5. Salvează.
- **Data:**
  - g: decimal from 1 to 300
  - feed type: from `FEED_TYPE`, filtered by sex when set
  - *sample (round 2, real):* Hîrbovăț 18.11.2026, every ♀ row 160 g: ♀ 14 395 × 160 g = 2 303,2 kg (guide SAPT 59 157 g)
- **Rules:**
  - A row left as an unconfirmed suggestion isn't saved. The section shows as missing.
  - **Warning:** a deviation from the guide above 10%.
  - **Warning:** stock that would go below 0: "Stocul de Breeder 2 la Hîrbovăț ar ajunge la −1 240 kg. Verifică livrările."
- **States:**
  - no previous day: empty fields with the guide figure as a placeholder
  - saved
  - missing sections listed on Acasă

### C3 · Transfer (manager) — MVP
- **Flow:**
  1. Date.
  2. **Din:** section (any farm).
  3. **În:** section (any farm).
  4. ♀ and ♂ counts. "Toate" fills everything that's alive.
  5. An effect preview.
  6. Mută.
- **Rules:**
  - No section may go below 0 on any day.
  - If the destination farm is empty, a cycle opens there; the preview says so ("Ciclu nou la Hîrbovăț, sapt. 20", round 2: the birds' age week, D151, as B4).
  - If the source farm reaches 0, its cycle closes; the preview says so.
  - Undo is available the same day (manager).

### C4 · Vânzare (manager) — MVP
- **Job:** When the laying flock is sold, I want to record how many birds left, so the farm empties and its cycle closes.
- **Flow:**
  1. Farm, then date.
  2. One row per section with ♀ and ♂ counts. "Toate" fills everything that's alive.
  3. Optional buyer and note.
  4. An effect preview ("Hîrbovăț: ♀ 12 561 · ♂ 794 → 0 · ciclul se închide", round 2: the 18.11.2026 counts, D158b).
  5. Înregistrează vânzarea.
- **Confirmation:** a dialog when the sale empties the farm.

### C5 · Livrare hrană (manager) — MVP
- **Fields:** farm, date, feed type, kg, optional note.
- **No price.**
- **Saved:** "Adăugat · 24 000 kg Breeder 2 · stoc Hîrbovăț: 31 450 kg · preț de adăugat de proprietar."

### C6 · Prețuri hrană (owner) — MVP
- **Job:** When invoices arrive, I want to add prices to the deliveries, so feed cost is complete.
- **Entry points:** Acasă (owner) shows "4 livrări fără preț"; Rapoarte › Hrană › Fără preț.
- **Flow:**
  1. A list of unpriced deliveries, oldest first: date, farm, type, kg.
  2. Enter lei/kg or the total lei; the other is calculated.
  3. Salvează prețurile saves all edited rows.
- **Rules:**
  - A price can be changed later by the owner only, with a reason.
  - It's logged.
- **States:** none unpriced: "Toate livrările au preț."

### C7 · Cântărire (vet) — MVP
- **Fields:** section, sex, date, weight (g), CV %, uniformity %, sample size.
- **Live comparison** (round 2, D157: a weighing in SAPT n is compared with guide week n − 1): "ghid la sfârșitul SAPT 20: 2 370 g · 2 553 g (+7,7%)" (Telița H1 ♀, 24.02.2026, SAPT 21 · abatere).

### C8 · Introdus azi — MVP
- The user's entries today, grouped by record type with their icons. Edit or delete until 23:59.

### C9 · Lipsesc date — MVP
- Sections with birds but no Zilnic entry or no Hrană zilnică entry for a day. Shown from 12:00, for up to 7 days back.

## 5. Views (viewer and up)

### D1 · Scope pages: Seria › Fermă (ciclu) › Secțiune — MVP
- **Header:** breadcrumbs, title, alive ♀ · ♂, the birds' week (SAPT, round 2: Hîrbovăț 18.11.2026 "♀ 12 561 · ♂ 794 · SAPT 59").
- **Series level:** the cycle track (both farms, all cycles of the series), then the tabs (below: table, notes, charts), then a summary per farm (D174).
- **Farm level:**
  - the cycle selector (the current cycle by default, past cycles listed)
  - the filter bar, then the tabs
  - in Rezumat, after its table (round 2), a section overview: one tile per section, showing:
    - ♀ · ♂ alive
    - mortality this week
    - grams today
    - weight vs guide
    - % lay (laying)
  - farm totals = the sum of the sections
- **Section level:** the same tabs, scoped to one section.

**Table first (round 2, D154).** Every tab opens on its table (components › Additions › PeriodTable); its charts and lists come after the table. The table is the record as the workbook keeps it, by week with the days inside, without copying the sheet 1:1.

- **Rows (every tab):**
  - one **week row** per age week in the period, total-row style: "SAPT 58 · 10.11–16.11"; it holds the week's sums (deaths, sales, eggs, kg), its end values (alive at the end of the week, stock) and its rates (weekly lay % = eggs ÷ hen-days counted from start-of-day hens, D156/D159; % incubabile of the week's eggs)
  - under it its **7 daily rows**: Data · Vârstă (zile) · SAPT (D159: the column is the bird age per D151, 407 on 18.11.2026, not the workbook's Zile 408), e.g. "Ma 17.11 · 406 · 59", "Mi 18.11 · 407 · 59" (weekday shown because weeks run on age: Tue–Mon for Seria 2026)
  - **grain** (FilterBar Zi / Săpt.): Săpt. = week rows only, each with its expand key; Zi = every week open. Default Zi (D166): every week of the period opens expanded, its week row followed by its day rows, newest first, no click needed; each expand key can still close its week. On every board and preview every table shows every week open (D167), loading and state boards included. Opening or closing one week doesn't change the grain.
  - **order (D159):** newest week first; inside a week the days run by date (Ma 17.11, Mi 18.11 …). Default period = the last 4 weeks ("4 săpt.": SAPT 56–59 on 18.11.2026)
  - a foot row "Total perioadă" (sums, end values at the period's end, rates over the period)
- **Columns:**
  - the spine: Data · Vârstă (zile) · SAPT (the first column sticks when the table scrolls sideways)
  - then **column groups by record type**, chosen by the FilterBar RecordChips: a chip off hides its columns (and its rows in the event log under the table). Each group is split ♀ · ♂ where the record has a sex. **Order (D167):** "În viață" (vii) is the first group, right after the spine; "Vânzare" (de vânzare) is always the last group; the other record groups sit between them
  - per section (D159): the farm page's tabs carry a **Detaliu** switch above the table, Fermă · Hale · Compartimente, default **Fermă** (the "Total fermă" group only); Hale adds one group per hall (H5 · H6 · H7), Compartimente one per compartment (H5.1 … H7.2) after "Total fermă"; the table scrolls sideways when they don't fit. Producție has eggs per hall only, so on it Compartimente shows the halls' groups with "ouăle se numără pe hală". Rezumat has no switch (Total fermă only). Series page = "Total serie"; section page = that section only, no switch
- **Tabs and their tables:**
  - **Rezumat:** the "Total fermă" table first, no section groups; its column groups go by record type (the tags toggle record groups), each split ♀ · ♂ where the record has a sex (D165): În viață ♀ · ♂ · Morți ♀ · ♂ (culls included, as the register; Păsări splits them off) · Ouat (ouă · % ouat · ghid · incubabile · % incubabile · jos · % jos; D167 adds the incubabile count and % jos) · Hrană kg ♀ · ♂ · Vânzare ♀ · ♂ (only when the period has sales; last, D167). A growth farm swaps the egg columns for g/pasăre ♀ · ♂ (efectiv · ghid) and Greutate ♀ (weighing weeks). Then, in this order: the key Readout with its guide chart and the Stats vs guide, the section overview (tiles), De verificat, the latest records. *Sample (Hîrbovăț, SAPT 59 week row, 17.11–18.11, 2 din 7 zile):* în viață ♀ 12 561 · ♂ 794 · morți ♀ 10 · ♂ 0 · ouă 11 806 · % ouat 40,4% (ghid 53,3%) · incubabile 10 918 (92,5%) · jos 311 (2,6%) · hrană ♀ 4 675 kg · ♂ 330 kg · vândute ♀ 2 255 · ♂ 433. SAPT 58 row: în viață ♀ 14 826 · ♂ 1 227 · morți ♀ 36 · ♂ 8 · ouă 47 512 · % ouat 30,4% (ghid 54,5%) · incubabile 45 745 (96,3%) · jos 1 221 (2,6%) · vândute ♀ 10 546 · ♂ 749.
  - **Păsări:** per section × sex. "În viață ♀ · ♂" (end of day; week row = end of week) is always shown, as the first group (D167). Chips → groups: Mortalitate (morți ♀ · ♂, week row also cumulative %), Sacrificare (♀ · ♂, reason in the event log), Eroare sexare (♀ · ♂, signed as recorded), Transfer (♀ · ♂, signed: + in, − out), Populare, Vânzare (vândute ♀ · ♂, the last group, D167). Then the alive and mortality charts, then the event log with the type chips. *Sample (18.11.2026, Total fermă):* morți ♀ 3 · ♂ 0 · vândute ♀ 1 831 · ♂ 289 · în viață ♀ 12 561 · ♂ 794.
  - **Hrană:** per section × sex: g/pasăre efectiv · ghid, kg; farm group also Livrări kg (♀ feed · ♂ feed), lei (owner, manager; never viewer, D64) and Stoc kg at the end of the day; week row = kg and deliveries summed, g/pasăre = kg ÷ bird-days (D38), stock at the week's end. Chip Hrană shows or hides them all. Then the g/pasăre vs guide chart, kg/zi, stock and days left (farm level). *Sample (18.11.2026):* ♀ 2 303,2 kg · 160 g (ghid 157 g) · ♂ 154,7 kg; stoc ♀ feed 31 664,9 kg · ♂ feed 1 781 kg.
  - **Greutate:** per section × sex: medie g, ghid (SAPT − 1, D157), față de ghid %, CV %, uniformitate %. A week row shows that week's weighing; its open rows list only the days with a weighing (the other days hold nothing to read), and a week without one opens, like every week, to one muted row "fără cântărire în această săptămână" (D167: no week looks collapsed). Then weight vs guide chart. Hîrbovăț has no real weighings (illustrative ones only, DATA03).
  - **Producție (laying only):** per hall (eggs are counted per hall; compartment detail shows its hall's eggs with "pe hală"): găini (start of day) · ouă · % ouat efectiv · ghid · față de ghid · stare · incubabile n · % · jos n · % · masă ou g. Categories Cat. 1–3 are not shown this round (D153). Then the % ouat and egg-weight charts. *Sample (18.11.2026):* ouă 5 914 · incubabile 5 308 (89,8%) · jos 158 · % ouat 41,1% (ghid 53,3%) · masă ou 66,4 g.
- **Series page (S3, round 2):** the same tabs, table first, with no FilterBar (D63): every week open by default (D166), each closes with its key; every column group is shown; the table opens on the last 4 weeks (newest first, SAPT 59–56) with an "Arată toate săptămânile" key under it that shows SAPT 59 → 1 (D160); Exportă in the header exports what's on screen. Columns: one "Total serie" group plus a "Fermă" text column after SAPT (Telița SAPT 1–21, Hîrbovăț from SAPT 20, both in the transfer weeks SAPT 20–21); per-farm detail is the farm page. Order (D174): header → cycle track → tabs (table + "Arată toate săptămânile" + notes → charts) → "Ferme" (the farm summaries). In the transfer weeks each bird is counted once (in Telița until its batch arrives at Hîrbovăț). On a phone the record tags are single-choice as on the farm page (default Mortalitate) and the farms ride on the sticky cell's second line in the transfer weeks.
- **Section page (S5, round 2):** the farm page's tabs and FilterBar, no Detaliu switch, one group (the section, ♀ · ♂); Rezumat = the section's table, then its Stats, guide charts and latest records (no tiles). A compartment's Producție shows its hall's eggs with "ouăle se numără pe hală (H5)".
- **Filters:** period (Ciclul curent · 4 săpt. · 13 săpt. · Interval; default 4 săpt., D159), grain (Zi / Săptămână), record types, export .xlsx. Record-type tags toggle several column groups on desktop; on phones they are single-choice, one tag = one column group at a time (D160). The period counts whole age weeks ("4 săpt." = the current week and the 3 before it). The Detaliu switch sits with the table, not in the FilterBar.
- **Export .xlsx = what's on screen (round 2):** the same rows (open weeks with their days, closed weeks as week rows), column groups, chips, Detaliu level and order, one sheet per tab, numbers as numbers; head lines name the farm, cycle, period and chips; a `–` cell exports empty and its reason goes in the sheet's note line.
- **States:**
  - loading: the table's skeleton at its loaded geometry (week rows with bars, D132), charts after it
  - no data in the period: EmptyState instead of the table; no type chips, Exportă disabled, next step "Alege ciclul curent" (D48)
  - no guide: the guide columns (ghid, față de ghid, stare) drop out (D92), header Tag "fără ghid"; a partial guide shows `–` in the ghid cell and "fără date în ghid" in the table's foot note (D29)
  - no access
  - an empty farm between cycles: no filter bar, tabs or table (D48), EmptyState "Ferma e goală din 28.02.2026. Ultimul ciclu: Seria 2026, SAPT 1–21." (Telița, round 2)
  - **partial current week:** the week row says how many days it holds, "SAPT 59 · 17.11–18.11 · 2 din 7 zile"; sums cover those days, rates use them only; the week is open by default
  - **missing entries:** the day's cell is `–` with the Lamp "lipsă"; the week row's sum covers the entered days and its row header says "6 din 7 zile"
  - **emptying weeks with sales (D155):** Mortalitate holds real deaths only, the removals sit in Vânzare (SAPT 58 ♀ 10 546 · ♂ 749; 18.11.2026 ♀ 1 831 · ♂ 289); alive falls by both; a section emptied in the period shows `–` after its last day and its group head says "goală din 14.11" (H7, D158); weekly lay % on hen-days from start-of-day hens, so it sits between its days (SAPT 59 40,4%: 17.11 39,7%, 18.11 41,1%, D159)
  - past cycle (D52): period "Tot ciclul", rows only where the register has data, emptying not shown as mortality (D54)

### D2 · Acasă — MVP
- **Vet:**
  - missing sections today
  - the entry shortcuts (Zilnic, Hrană zilnică, Cântărire)
  - Introdus azi
- **Manager:**
  - the above, plus alerts for their farms
  - feed days left
  - Transfer, Vânzare and Livrare shortcuts
- **Owner:**
  - the cycle track for the active series
  - alerts across farms
  - "livrări fără preț"
  - feed stock

### D3 · Producție report — MVP
- A weekly table per farm › section, by the birds' age week (SAPT, round 2).
- Columns: hens alive, eggs, % lay vs guide, hatching eggs and %, egg weight, mortality, feed g/hen.
- **Daily rows (round 2, D154):** each week is one block (PeriodTable): the farm's week row (total-row style, "SAPT 58 · 10.11–16.11", expand key), then, when open, its 7 daily farm rows (Data · Vârstă (zile) · SAPT), then the hall week rows H5 · H6 · H7 (D92, child rows; days are farm-level only, hall days live on the farm page's Producție tab). Grain Zi opens every week and is the default (D166: every week open with its days); Săpt. shows week rows only. Newest week first, days by date inside it, default period the last 4 weeks (D159).
- Rates: daily % ouat = eggs ÷ hens at the start of the day, weekly = eggs ÷ hen-days counted from start-of-day hens (D156, D159); hall and farm figures from section sums, never the sheet's Total row. Mortality = real deaths, sales in their own column "Vândute ♀" in weeks that have them (D155).
- *Sample (Hîrbovăț, real, newest first):* SAPT 59 (17.11–18.11, 2 din 7 zile) ouă 11 806 · 40,4% (ghid 53,3%) · incubabile 92,5% · vândute ♀ 2 255; SAPT 58 (10.11–16.11) ouă 47 512 · 30,4% (ghid 54,5%) · incubabile 96,3% · vândute ♀ 10 546; SAPT 57 (03.11–09.11) ouă 49 225 · 27,6% (ghid 55,7%) · incubabile 98,7%; 18.11.2026 day row (farm) ouă 5 914 · 41,1% (ghid 53,3%). Hall eggs by day (18.11.2026: H5 3 872 · H6 2 042 · H7 – goală din 14.11) are on the farm page's Producție tab with Detaliu Hale; here the hall rows are weekly.
- Export .xlsx = what's on screen (D1).

### D4 · Hrană report — MVP
1. **Stoc:** farm × feed type: kg, average per day, days left, run-out date.
2. **Livrări:** all deliveries with their price status (with price / fără preț). The owner can edit inline.
3. **Consum:** per week × section: kg, g/bird vs guide. Round 2 (D154): each week is a PeriodTable block: the farm's week row (total-row style, "SAPT 58 · 10.11–16.11", kg ♀ 22 230,9 · ♂ 1 682,7), its 7 daily farm rows when open (Data · Vârstă (zile) · SAPT, e.g. "Mi 18.11 · 407 · 59" ♀ 2 303,2 kg · 160 g, ghid 157 g · ♂ 154,7 kg), then the section week rows (as now). Weekly g/bird = kg ÷ bird-days (D38); grain, newest-first order and the 4-week default as D1 (D159). Export = what's on screen.
4. **Cost:** lei per cycle, per bird, per egg, from priced deliveries, plus a line for the "X kg fără preț" that isn't included.
- Prices are hidden from vets. Viewers and managers see them read-only.

### D5 · Export — MVP
- Every table can be exported to .xlsx with the current filters.

### D6 · Compare cycles — later
- The same farm, different series, by the birds' age week (SAPT, round 2).

## 6. Corrections
- **E1 · Corectează (manager, owner):** see the appendix, with a required reason, an effect preview, the never-below-0 check across every later day, and an audit entry.
- **E2 · Istoric modificări:** see the appendix, filterable by record type.

## 7. Out of scope
- the incubator (a separate module later)
- offline entry
- notifications
- importing the Seria 2026 history (a one-time developer migration)


## Appendix: unchanged features from spec v1

### A1 · Sign in — MVP
- **Job:** When I open the app, I want to get in quickly on my phone, so I can enter today's numbers.
- **Entry points:** app URL, or the home-screen shortcut (a PWA install `default`).
- **Flow:**
  1. Enter login and password.
  2. Tap Intră.
  3. You land on your role's start page (see Stage 4).
  4. The session lasts 30 days on that device `default`.
- **Data:**
  - `login`: text, required, 3–40 characters. *Sample:* `petrea`
  - `password`: required, minimum 8 characters.
- **Actions:**
  - Primary: Intră.
  - There's no self-service password reset. The owner resets passwords (A2).
- **Rules:** after 5 failed attempts, sign-in waits 5 minutes.
- **States:**
  - Wrong credentials: "Login sau parolă greșită."
  - Locked: "Prea multe încercări. Încearcă din nou peste 5 minute."
  - Deactivated user: "Contul tău este dezactivat. Vorbește cu administratorul."
  - Offline: "Nu există conexiune. Verifică internetul și încearcă din nou."
- **Copy:** title "Impex Line", fields "Login" and "Parolă", button "Intră".
- **Success signal:** vets stay signed in and don't ask for their password each week.

### A2 · Users and access — MVP (owner)
- **Job:** When someone joins or leaves, I want to give or remove access to specific farms, so each person sees only what they need.
- **Entry points:** Setări › Utilizatori.
- **Flow:**
  1. The list of users shows name, role (vet / viewer / manager / owner), farms, last sign-in and whether the account is active.
  2. Adaugă utilizator opens a form: name, login, temporary password, role, farms. Saving creates the user.
  3. Opening a user lets you change role or farms, reset the password, or deactivate the account.
- **Data:**
  - `name`: required, ≤ 60 characters.
  - `login`: unique, `[a-z0-9.]`, 3–40 characters.
  - `role`: required.
  - `farms`: at least one, unless the role is owner.
- **Actions:**
  - Primary: Salvează.
  - Secondary: Resetează parola. Shows the new temporary password once.
  - Destructive: Dezactivează. Confirmed in a dialog ("Petrea nu va mai putea intra. Datele introduse de el rămân."). Can be undone with Reactivează.
  - Users are never deleted, because their entries keep pointing to them.
- **Rules:**
  - At least one active owner must always exist.
  - You can't deactivate yourself.
- **States:**
  - Empty: "Doar tu ai acces. Adaugă primul utilizator."
  - Duplicate login: "Loginul „petrea” este deja folosit."
- **Success signal:** access changes take under a minute, with no developer involved.

### E1 · Edit past data — MVP
- **Job:** When a past number turns out wrong, I want to fix it with a reason, so the counts are right and we know what changed.
- **Entry points:** any entry in an event log or table › Corectează.
- **Flow:**
  1. Open the entry: the whole save (D172, one save = one record, D168), e.g. a Zilnic save shows every compartment × sex in it as the Zilnic form does; any figure can be changed, a changed figure is marked ("modificat", the saved value under it).
  2. A reason is required (≤ 200 characters).
  3. A preview shows the effect of each change and the totals that follow: "H2 · găini în viață de pe 03.11.2026 se schimbă cu +12."
  4. Salvează corecția.
- **Actions:** Șterge (the whole save), with the reason required and a confirmation dialog.
- **Rules:**
  - The same "never below 0" check runs over every later day.
  - The old and new values go to `AUDIT_LOG`.
- **States:** refused: "Corecția ar lăsa H6 cu −4 găini pe 16.02.2026."

### E2 · Change log — MVP
- **Entry points:** Setări › Istoric modificări, and on each entry ("modificat de Ion · 16.11.2026, 09:40").
- **Content:** a list with date, user, table, hall, old → new, reason. Filter by user, farm and period.
