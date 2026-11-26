# Impex Line: functional spec v2 (shared by both designs)

Status: 16.11.2026, round 2 updates 19.11.2026 (today on the boards = 18.11.2026, D152); round 3 adds §8 Voice (26.11.2026, D199). This replaces spec v1.
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
- **Laying figures (round 2):** % ouat comes from the workbook's Productivitate sheet, col I, by age week 23–62 (D184; RealGhid col B is no longer used for lay %); cumulative mortality % stays RealGhid. % incubabile is hatching eggs of all eggs, not hatchability (D153). *Sample:* SAPT 59 ♀ 157 g · ♂ 156 g, % ouat 57,6% (SAPT 56 61,1%).
- **Rules:**
  - Duplicate week + sex rows are refused.
  - A guide used by a cycle can be replaced but not deleted.
  - A laying guide whose first row is week 1 imports with one file-level warn Banner in the import preview, saveable (D180): "Ghidul de ouat începe la SAPT 1. Săptămâna 1 e prima săptămână de viață. Ghidurile de ouat încep de obicei la SAPT 19–23; verifică dacă fișierul numără săptămânile de la transfer." (round 2)
- **States:** parse errors per row (D180: a % over 100 is a refused row, Save blocked; the import example is the Productivitate laying guide pasted with week 53 typed "645" for 64,5%, refused, D185); a partial guide shows `–` with "fără date în ghid" for every SAPT outside its rows (Hîrbovăț's laying guide has no body weight; growth weights end at SAPT 25).

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
  - Total eggs = Cat. 1 + 2 + 3 + Incubabile; Jos is kept separate (D18).
  - **Egg categories (round 2, D183):** Cat. 1, Cat. 2 and Cat. 3 are counts per hall per day, entered with Incubabile (supersedes D164's placeholder). *Sample (18.11.2026):* H5 Cat. 1 13 · Cat. 2 417 · Cat. 3 27 · Incubabile 3 415 = 3 872; H6 8 · 110 · 31 · 1 893 = 2 042.
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
  - **order (D159):** newest week first; inside a week the days run by date (Ma 17.11, Mi 18.11 …). The table opens on the last 4 weeks of the period (SAPT 56–59 on 18.11.2026); the default period is "Ciclul curent" (D191, supersedes D159's "4 săpt." default)
  - **load-more row (D190, D192):** the last body row, after the last loaded week: one cell across the table, row height, faded fill, muted text "Încarcă încă 4 săptămâni · SAPT 56–59 · 4 din 40" (40 = the cycle's weeks, SAPT 20–59); the whole row is the button; each press appends the next 4 older weeks open with their days, keeps focus and a polite live region says "Au fost adăugate SAPT 52–55."; with a period filter set the press widens it 4 weeks back and the FilterBar follows ("4 săpt." → "8 săpt.", Interval start −4 weeks); at the cycle's first week the row becomes the muted note "Începutul ciclului · SAPT 20"
  - a foot row "Total SAPT 56–59" (D191: the weeks on screen, labelled by their span, growing as weeks load: sums, end values at the last loaded day, rates over the loaded weeks)
- **Columns:**
  - the spine: Data · Vârstă (zile) · SAPT (the first column sticks when the table scrolls sideways)
  - then **column groups by record type**, chosen by the FilterBar RecordChips: a chip off hides its columns (and its rows in the event log under the table). Each group is split ♀ · ♂ where the record has a sex. **Order (D167):** "În viață" (vii) is the first group, right after the spine; "Vânzare" (de vânzare) is always the last group; the other record groups sit between them
  - per section (D159): the farm page's tabs carry a **Detaliu** switch above the table, Fermă · Hale · Compartimente, default **Fermă** (the "Total fermă" group only); Hale adds one group per hall (H5 · H6 · H7), Compartimente one per compartment (H5.1 … H7.2) after "Total fermă"; the table scrolls sideways when they don't fit. Producție has eggs per hall only, so on it Compartimente shows the halls' groups with "ouăle se numără pe hală". Rezumat has no switch (Total fermă only). Series page = "Total serie"; section page = that section only, no switch
- **Tabs and their tables:**
  - **Rezumat:** the "Total fermă" table first, no section groups; its column groups go by record type (the tags toggle record groups), each split ♀ · ♂ where the record has a sex (D165): În viață ♀ · ♂ · Morți ♀ · ♂ (culls included, as the register; Păsări splits them off) · Ouat (ouă · % ouat · ghid · incubabile · % incubabile · jos · % jos; D167 adds the incubabile count and % jos) · Hrană kg ♀ · ♂ · Vânzare ♀ · ♂ (only when the period has sales; last, D167). A growth farm swaps the egg columns for g/pasăre ♀ · ♂ (efectiv · ghid) and Greutate ♀ (weighing weeks). Then, in this order: the key Readout with its guide chart and the Stats vs guide, the section overview (tiles), De verificat, the latest records. *Sample (Hîrbovăț, SAPT 59 week row, 17.11–18.11, 2 din 7 zile):* în viață ♀ 12 561 · ♂ 794 · morți ♀ 10 · ♂ 0 · ouă 11 806 · % ouat 40,4% (ghid 57,6%, D184) · incubabile 10 918 (92,5%) · jos 311 (2,6%) · hrană ♀ 4 675 kg · ♂ 330 kg · vândute ♀ 2 255 · ♂ 433. SAPT 58 row: în viață ♀ 14 826 · ♂ 1 227 · morți ♀ 36 · ♂ 8 · ouă 47 512 · % ouat 30,4% (ghid 58,8%) · incubabile 45 745 (96,3%) · jos 1 221 (2,6%) · vândute ♀ 10 546 · ♂ 749.
  - **Păsări:** per section × sex. "În viață ♀ · ♂" (end of day; week row = end of week) is always shown, as the first group (D167). Chips → groups: Mortalitate (morți ♀ · ♂, week row also cumulative %), Sacrificare (♀ · ♂, reason in the event log), Eroare sexare (♀ · ♂, signed as recorded), Transfer (♀ · ♂, signed: + in, − out), Populare, Vânzare (vândute ♀ · ♂, the last group, D167). Then the alive and mortality charts, then the event log with the type chips. *Sample (18.11.2026, Total fermă):* morți ♀ 3 · ♂ 0 · vândute ♀ 1 831 · ♂ 289 · în viață ♀ 12 561 · ♂ 794.
  - **Hrană:** per section × sex: g/pasăre efectiv · ghid, kg; farm group also Livrări kg (♀ feed · ♂ feed), lei (owner, manager; never viewer, D64) and Stoc kg at the end of the day; week row = kg and deliveries summed, g/pasăre = kg ÷ bird-days (D38), stock at the week's end. Chip Hrană shows or hides them all. Then the g/pasăre vs guide chart, kg/zi, stock and days left (farm level). *Sample (18.11.2026):* ♀ 2 303,2 kg · 160 g (ghid 157 g) · ♂ 154,7 kg; stoc ♀ feed 31 664,9 kg · ♂ feed 1 781 kg.
  - **Greutate:** per section × sex: medie g, ghid (SAPT − 1, D157), față de ghid %, CV %, uniformitate %. A week row shows that week's weighing; its open rows list only the days with a weighing (the other days hold nothing to read), and a week without one opens, like every week, to one muted row "fără cântărire în această săptămână" (D167: no week looks collapsed). Then weight vs guide chart. Hîrbovăț has no real weighings (illustrative ones only, DATA03).
  - **Producție (laying only):** per hall (eggs are counted per hall; compartment detail shows its hall's eggs with "pe hală"): găini (start of day) · ouă · % ouat efectiv · ghid · față de ghid · stare · % Cat. 1 · % Cat. 2 · % Cat. 3 · incubabile n · % · jos n · % · masă ou g. The categories are shown as % of the eggs (count ÷ ouă), never as counts, on week and day rows, for the farm and each hall, in the Ouat group's tint (D183; replaces D153's "not shown"). Then the % ouat and egg-weight charts. *Sample (18.11.2026):* ouă 5 914 · incubabile 5 308 (89,8%) · jos 158 · % ouat 41,1% (ghid 57,6%, D184) · masă ou 66,4 g; SAPT 59 % Cat. 1 0,3% · % Cat. 2 6,2% · % Cat. 3 1,1%; SAPT 58 0,3% · 2,6% · 0,8%.
- **Series page (S3, round 2):** the same tabs, table first, with no FilterBar (D63): every week open by default (D166), each closes with its key; every column group is shown; the period is the whole series (SAPT 1–59): the table opens on its last 4 weeks (newest first, SAPT 59–56) and its load-more row reads "· 4 din 59", down to "Începutul seriei · SAPT 1" (D190, D192; replaces "Arată toate săptămânile", D160); the groups are the series' types (Eroare sexare and Populare included, D45), so they stay put while weeks load; Exportă in the header exports the whole series (D191). Columns: one "Total serie" group plus a "Fermă" text column after SAPT (Telița SAPT 1–21, Hîrbovăț from SAPT 20, both in the transfer weeks SAPT 20–21); per-farm detail is the farm page. Order (D174): header → cycle track → tabs (table with its load-more row + notes → charts) → "Ferme" (the farm summaries). In the transfer weeks each bird is counted once (in Telița until its batch arrives at Hîrbovăț). On a phone the record tags are single-choice as on the farm page (default Mortalitate) and the farms ride on the sticky cell's second line in the transfer weeks.
- **Section page (S5, round 2):** the farm page's tabs and FilterBar, no Detaliu switch, one group (the section, ♀ · ♂); Rezumat = the section's table, then its Stats, guide charts and latest records (no tiles). A compartment's Producție shows its hall's eggs with "ouăle se numără pe hală (H5)".
- **Filters:** period (Ciclul curent · 4 săpt. · 13 săpt. · Interval; default Ciclul curent, D191; the load-more row widens a set period 4 weeks back, "4 săpt." → "8 săpt.", D192), the record-type chips = the period's types (D45: the cycle adds Eroare sexare and Transfer), grain (Zi / Săptămână), record types, export .xlsx. Record-type tags toggle several column groups on desktop; on phones they are single-choice, one tag = one column group at a time (D160). The period counts whole age weeks ("4 săpt." = the current week and the 3 before it). The Detaliu switch sits with the table, not in the FilterBar.
- **Export .xlsx = the whole chosen period (D191):** every week of the period, not only the loaded ones (open weeks with their days, closed weeks as week rows), with the column groups, chips, Detaliu level and order on screen, one sheet per tab, numbers as numbers; head lines name the farm, cycle, period and chips; a `–` cell exports empty and its reason goes in the sheet's note line.
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
- Columns: hens alive, eggs, % lay vs guide, % Cat. 1 · % Cat. 2 · % Cat. 3 (of the eggs, D183; farm week and day rows and hall week rows; also on the closed previous cycle (Seria 2025, from its extract, D187); on the phone the Cat. columns are in the tag «Incubabile · Cat.» (D186, D188)), hatching eggs and %, egg weight, mortality, feed g/hen. Guide lay % = Productivitate col I (D184).
- **Daily rows (round 2, D154):** each week is one block (PeriodTable): the farm's week row (total-row style, "SAPT 58 · 10.11–16.11", expand key), then, when open, its 7 daily farm rows (Data · Vârstă (zile) · SAPT), then the hall week rows H5 · H6 · H7 (D92, child rows; days are farm-level only, hall days live on the farm page's Producție tab). Grain Zi opens every week and is the default (D166: every week open with its days); Săpt. shows week rows only. Newest week first, days by date inside it; default period "Ciclul curent", the table opening on its last 4 weeks with the load-more row "Încarcă încă 4 săptămâni · SAPT 56–59 · 4 din 40" before the foot "Total SAPT 56–59" (D190–D192).
- Rates: daily % ouat = eggs ÷ hens at the start of the day, weekly = eggs ÷ hen-days counted from start-of-day hens (D156, D159); hall and farm figures from section sums, never the sheet's Total row. Mortality = real deaths, sales in their own column "Vândute ♀" in weeks that have them (D155).
- *Sample (Hîrbovăț, real, newest first):* SAPT 59 (17.11–18.11, 2 din 7 zile) ouă 11 806 · 40,4% (ghid 57,6%) · Cat. 1–3 0,3% · 6,2% · 1,1% · incubabile 92,5% · vândute ♀ 2 255; SAPT 58 (10.11–16.11) ouă 47 512 · 30,4% (ghid 58,8%) · Cat. 1–3 0,3% · 2,6% · 0,8% · incubabile 96,3% · vândute ♀ 10 546; SAPT 57 (03.11–09.11) ouă 49 225 · 27,6% (ghid 59,9%) · incubabile 98,7%; 18.11.2026 day row (farm) ouă 5 914 · 41,1% (ghid 57,6%). Hall eggs by day (18.11.2026: H5 3 872 · H6 2 042 · H7 – goală din 14.11) are on the farm page's Producție tab with Detaliu Hale; here the hall rows are weekly.
- Export .xlsx = the whole chosen period (D1, D191).

### D4 · Hrană report — MVP
1. **Stoc:** farm × feed type: kg, average per day, days left, run-out date.
2. **Livrări:** all deliveries with their price status (with price / fără preț). The owner can edit inline.
3. **Consum:** per week × section: kg, g/bird vs guide. Round 2 (D154): each week is a PeriodTable block: the farm's week row (total-row style, "SAPT 58 · 10.11–16.11", kg ♀ 22 230,9 · ♂ 1 682,7), its 7 daily farm rows when open (Data · Vârstă (zile) · SAPT, e.g. "Mi 18.11 · 407 · 59" ♀ 2 303,2 kg · 160 g, ghid 157 g · ♂ 154,7 kg), then the section week rows (as now). Weekly g/bird = kg ÷ bird-days (D38); grain, newest-first order, the "Ciclul curent" default with the last 4 weeks and the load-more row as D1 (D190–D192; replaces "Vezi toate săptămânile"). Export = the whole chosen period (D191).
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
- the incubator (a separate module later); round 2 (D182): the workbook's `Incubare` batches and `BD Incubator` history get no table, entry form or report
- offline entry
- notifications
- importing the Seria 2026 history (a one-time developer migration)
- voice (round 3, §8): a wake word or always-on listening, speech in any language but Romanian, voice before sign-in

## 8. Voice (round 3, 26.11.2026, D199)

### F1 · Comenzi vocale — MVP (round 3)
- **Job:** When my hands are busy in the hall, or I don't know where something lives, I want to say what I want to see or enter, so the app opens it, or fills the form for me to check and save.
- **Who:** every signed-in role, only within its role, its farms and its time windows (§1). Voice is a second way to press what the screen already offers; it never adds a power the screen doesn't give that user.
- **Decided by:** user (D199): show/filter/navigate runs at once; anything that changes data opens the right form prefilled and waits for Salvează; phone floating mic above the BottomTabBar, desktop mic key in the page header + shortcut; tap to start, stops by itself after a pause or on a second tap. Everything else here: Claude, to review (see "Decision points" at the end).

**Where**
- **Every signed-in screen and state:** S2–S22, their tabs, loading, empty, error and saved states, open SidePanels, the phone full-screen sheet (S16), BottomSheets (Înregistrează menu, FilterBar sheet, SectionSelect) and open menus (pressing the key closes the menu first).
- **Not before sign-in:** S1 has no voice key and the shortcut does nothing (no user, no role; a password is never spoken). *To review.*
- **Dialogs (D199 note):** voice is off inside an open Dialog: the key sits under its scrim (`inert`, hidden from the reader) and the shortcut does nothing. Voice can still *open* two kinds differently:
  - **form Dialogs** (Începe ciclu nou, Schimbă ghidul ciclului, the series' Vârstă la populare change): opened by voice, prefilled like any form (VoiceMarks, Motiv only if spoken), focus on the first filled field; the user presses its key.
  - **confirm Dialogs** (the Vânzare that empties the farm, Șterge, Dezactivează, Reactivează, Închide seria, Retrage, Resetează parola): never opened by voice; voice opens the record or form with the key that opens the Dialog focused, and the user presses it. *To review.*
- **Phone:** a round floating key (VoiceButton, 56 px) at the bottom right, `space-4` from the screen edge and `space-4` above the lowest bar on screen: the BottomTabBar, or the SaveBar on a form, or the foot of the full-screen sheet; it never covers them. Over a BottomSheet it sits above the sheet's top edge on the scrim. Every scrolling phone view sets `scroll-padding-bottom` and an equal bottom padding of 88 px (56 + `space-4` × 2) on top of its bar's own (SaveBar FX05), so a focused field, the last row or the load-more row scrolls clear of the key (WCAG 2.4.11). While the soft keyboard is open (a field focused, read from the visual viewport) the floating key is hidden (`hidden`, out of the reading order) and comes back when the keyboard closes; the shortcut still works from a hardware keyboard; the VoicePanel's own field is the exception (the key stays docked in its foot). A phone Toast stacks above the key. *To review.*
- **Desktop:** a mic IconButton, the last key on the right of the PageHeader actions on every page (alone where a page has no actions), the same place on every screen; with a SidePanel open it stays in the PageHeader. Shortcut **Ctrl + Shift + Space** (all platforms, ⌃⇧Space on a Mac keyboard).

**Flow**
1. **Start:** tap the key (or the shortcut). The VoicePanel opens (phone: a BottomSheet; desktop: a popover under the header key) and the key latches (`aria-pressed="true"`).
2. **Microphone permission** (first use on the device): the panel says why before the browser asks.
3. **Ascultă:** Lamp `ok` + "Ascult…", a live level meter, the elapsed time in Mono ("0:04"), and the live transcript (`ink-secondary` while it can still change, `ink` once final).
4. **Stop:** by itself after ~1,5 s of silence once speech has started; after 20 s at most ("Am oprit după 20 de secunde."); on a second tap of the key (stops and runs what was said); if nothing is heard for 5 s, it stops with the neînțeles state ("Nu am auzit nimic.").
5. **Transcriere → se gândește:** the final transcript stays on screen, Lamp `info` + "Mă gândesc…", the key shows the Button spinner (`aria-busy`). Tapping the transcript text now stops the run and turns it into an editable field (see Transcript).
6. **Result**, one of: executat · formular precompletat · răspuns · clarificare · neînțeles · nepermis · fără microfon · fără conexiune/eroare (below). A run with no result after 10 s becomes the error state.
7. **Cancel at any step:** Esc, Închide (the panel's IconButton) or the panel's Anulează; tapping outside the popover (desktop) or on the scrim (phone). Nothing runs, the key unlatches, focus returns to where it was before the start (the key, or the field the shortcut was pressed in), the app-level status region says "Comanda a fost anulată.". Esc cancels voice first: while the panel is open Esc closes only it; a SidePanel, the S16 sheet or a menu under it stays open until the next Esc. A second tap of the key while listening is *stop*, not cancel (step 4); a tap on it while "Mă gândesc…" cancels.

**Transcript (shown live, editable before it runs)**
- The transcript is plain text in Plex Sans in „…”, digits as heard ("pune 3 găini moarte la H5.1"); below it, once understood, the normalised command in the app's own format (Mono figures, ♀/♂ signs, section codes): "Zilnic · Hîrbovăț · 18.11.2026 · H5.1 · ♀ Mortalitate 3".
- **Fix it:** tap (or Enter on) the transcript → it becomes a TextField holding the text, the run stops, keys **Execută** (primary) and Anulează. Execută runs the edited text exactly as a spoken one.
- **Show/filter/navigate commands don't wait for Execută** (D199): they run as soon as they are understood; the fix window is the "Mă gândesc…" time, and Anulează on the result Toast undoes them.
- **Scrie comanda:** the panel's foot always holds a TextField "Scrie comanda" + Execută (Enter runs it), so every command works without a microphone, with a screen reader, or in a loud hall. A typed command skips steps 2–4.

**States and copy**
- **Idle:** no panel; the key at rest. Its tooltip (desktop) and description: "Comandă vocală · Ctrl+Shift+Space".
- **Cerere permisiune microfon:** Lamp `info` "aștept permisiunea"; "Impex Line are nevoie de microfon ca să-ți audă comanda. Apasă Permite în fereastra browserului."; the text field is already usable.
- **Ascultă:** Lamp `ok` "Ascult…" · level meter · "0:04" · live transcript · hint "Se oprește singur după o pauză. Apasă din nou ca să termini." · examples for this screen ("Poți spune: „arată mortalitatea pe ultimele 8 săptămâni”") until the first word arrives.
- **Auto-stop:** silence ~1,5 s after speech; 20 s cap; 5 s with no speech. The meter stops, the key shows the spinner.
- **Transcriere / se gândește:** Lamp `info` "Mă gândesc…"; transcript final and tappable.
- **Executat (show, filter, navigate, export):** the panel closes and the app shows the result exactly as if the user had pressed the keys (URL, FilterBar, chips, tabs, focus on the page h1 as after any navigation, FX02). Toast (no record icon, no figures beyond the scope; it is also the announcement, as the panel has closed): "Am deschis Rapoarte · Producție, ultimele 8 săptămâni" / "Am deschis Hîrbovăț › H5 · Păsări: Mortalitate, 8 săpt." / "Am încărcat sapt. 52–55" + **Anulează**, which returns to the previous screen with its filters, loaded weeks and scroll (a form's draft is kept, §1 Connectivity). Several steps in one command (open a page, then filter it) are one Anulează. **Export** is the exception: its Toast has no Anulează (a download can't be taken back): "Export .xlsx descărcat · Hîrbovăț · Păsări · sapt. 20–59".
- **Răspuns (questions):** the view that shows the figure opens behind the panel and the panel stays open with a short written answer in the panel's result area: figures in Mono, ♀/♂ signs, the week and date, the status as Lamp + word. "La Hîrbovăț sunt azi ♀ 12 561 găini și ♂ 794 cocoși (18.11.2026, sapt. 59)." / "Sapt. 59 (2 din 7 zile): ouat 40,4%, ghid 57,6%, sub ghid cu 17,2 puncte." + Lamp `warn` "abatere". Under it "Am deschis Rapoarte · Producție" + Anulează (link). The answer is never spoken by the app (no speech synthesis); a screen reader reads it from the live region. *To review.* An answer only uses figures the user may see on that screen (prices never for vet/viewer, D06/D64).
- **Formular precompletat (data changes):** the panel closes; the right form opens for the right farm, date and section (scrolled to the first filled card or row), filled with what was understood; nothing is saved.
  - Each filled field carries the VoiceMark "din voce" until the user edits it by hand, presses Anulează completarea, or saves.
  - An info Banner with action (FIN03) as the form's first block under the FormHeader / PageHeader: "Am completat din ce ai spus. Verifică și salvează." then a second line listing what was filled ("2 câmpuri: H5.1 · ♀ Mortalitate 3 · ♂ Mortalitate 1"), key **Anulează completarea** (restores the form as it was before the command).
  - Focus goes to the first filled field. The SaveBar works as usual: its summary counts the filled values; warnings and refusals run as if typed (a count above birds alive blocks Salvează with the form's own copy).
  - A field the command did not name keeps its value (or stays empty: empty is null, never 0). A field that already held a typed value different from the spoken one is replaced, and its mark says what it held: "din voce · înainte 2". *To review.*
  - A required field the command didn't give (Motiv of a Sacrificare, the Motiv of a correction) stays empty with the form's own refusal; if the choices are few it is asked first as a clarification (below).
  - Yesterday's-grams suggestions (D13): a row the command filled counts as confirmed (the user said it), and carries the mark. *To review.*
  - The day is already saved: **Editează** only for the user's own save from today (D07; a vet until 23:59); **Corectează** (S16, Motiv required, filled only if spoken) for a past save, and for a manager or owner changing someone else's save from today (D70 B); a vet on someone else's save gets the nepermis state (D71).
  - A form Dialog (Începe ciclu nou, Schimbă ghidul ciclului, Vârstă la populare) opens prefilled as a form; a confirm Dialog the form leads to (the Vânzare that empties the farm) opens only from the user's own press of the form's key (Where › Dialogs).
- **Clarificare (ambiguous):** a question and up to 3 RecordChip-styled choice chips (touch size on phone) + "Spune din nou". "La care secțiune din H5?" → H5.1 · H5.2; "La ce fermă?" → Telița · Hîrbovăț; "Ce motiv are sacrificarea?" → Boală · Șchiopătare · Altul. More than 3 candidates: the 3 most likely + "Alta…", which opens the screen's own selector (SectionSelect, Select). Choosing a chip runs the command; Esc cancels. Chips are a `role=radiogroup` named by the question.
- **Neînțeles:** "Nu am înțeles „…”." / "Încearcă așa: „deschide raportul de producție” sau „pune 3 găini moarte la H5.1”." (two examples chosen for this screen and role) + keys Spune din nou (primary) · Scrie comanda (focuses the field with the text kept).
- **Nepermis (role, farm, time window):** nothing opens; the panel shows the same refusal the screen would give, in the same Banner type: "Pentru zilele trecute cere managerului." (vet, past day, info, D77) · "Nu ai acces la această fermă." · "Nu ai acces la Istoric modificări." (D104) · the vet's own entry after 23:59: the reason the S15 key carries (D99) · a voice-only refusal for the never list (below): "Vocea nu salvează. Verifică formularul și apasă Salvează.", "Vocea nu șterge. Deschide înregistrarea și apasă Șterge.", "Parolele nu se spun cu voce." Where the screen would let the user go somewhere useful, the panel offers it as a chip ("Deschide Introdus azi").
- **Fără microfon (permission denied, no microphone, browser without capture):** warn Banner "Microfonul e blocat pentru Impex Line." + how to allow it, for the device in use: phone Chrome "Apasă lacătul de lângă adresă › Permisiuni › Microfon › Permite."; iPhone "Setări › Safari › Microfon › Permite." (home-screen app: "Setări › Impex Line › Microfon" when listed); desktop "Apasă pictograma microfonului din bara de adrese › Permite." + "Poți scrie comanda mai jos." The text field takes the focus. No microphone / unsupported: "Dispozitivul nu are microfon disponibil. Scrie comanda mai jos."
- **Fără conexiune / eroare (transcription or AI failed, timeout):** the §49 failure Banner (FX04/FY04), `role=alert`: "Comanda nu s-a trimis: „pune 3 găini moarte la H5.1”." / "Verifică conexiunea la internet, apoi apasă Reîncearcă. Ce ai spus rămâne scris mai jos." Key **Reîncearcă** (takes the focus) resends the transcript (or, if transcription itself failed, the audio still held in memory); the transcript is in the text field to fix. Offline before starting: the key is `aria-disabled` with the reason "Fără conexiune" and a press shows that Banner.
- **Anulare:** see Flow 7.

**How a command is understood**
- **Context sent with every command:** the screen (S-number, tab, open panel or sheet), its scope (series, farm, cycle, section), FilterBar state, the open form and its fields, the user's role and farms, today's date and SAPT. Defaults come from it: "aici" / no farm named = the current scope, else the user's only farm, else a clarification; no date = azi; "săptămâna asta" = the current age week (SAPT 59 on 18.11.2026), "săptămâna trecută" = SAPT 58 (§1: weeks are age weeks).
- **Allowed actions only:** the AI returns one action from the catalogue below with its parameters; the app checks role, farm and time window before it acts, exactly as the screen does, and then performs it through the same screen logic. A phrase that maps to no catalogue action is neînțeles; the AI never writes data and never invents a figure (questions are answered from the app's own figures). The backend spec for the endpoint and the catalogue follows this design (D199).
- **One command, one screen:** a phrase that names two forms ("pune mortalitatea și hrana") fills the first; the second is offered as a chip on the Banner ("Continuă cu Hrană zilnică"). Several fields of one form in one phrase are fine. *To review.*
- **Romanian, spoken farm vocabulary:**
  - numbers as words or digits: "trei", "două sute cincizeci", "o mie opt sute treizeci și unu", "doi virgulă patruzeci și unu" (2,41), "o sută șaizeci de grame", "nouă lei douăzeci și doi" (9,22 lei); "zero"/"niciuna" = 0, never empty
  - sections: "ha cinci punct unu", "hala cinci compartimentul unu", "H cinci unu" → H5.1 (Hîrbovăț); "ha cinci liniuță unu" → H5-1 (Telița); "hala cinci" / "H5" → the hall (a clarification where a compartment is needed)
  - farms: "Hîrbovăț" (also heard "Hirbovat"), "Telița"; sex: găini / găină / femele → ♀, cocoși / cocoș / masculi → ♂; "păsări" = both
  - weeks and dates: "SAPT 58", "săptămâna cincizeci și opt", "azi", "ieri", "alaltăieri", "pe 13 noiembrie", "de pe 1 până pe 15 noiembrie"
  - record words: moarte / mortalitate / au murit; sacrificate / sacrificare; eroare de sexare; transfer / mută; vânzare / am vândut; hrană / grame / porție; livrare / am primit; cântărire / greutate / CV / uniformitate / probă; ouă / categoria unu / incubabile / jos / masa medie; preț / lei pe kilogram
  - diacritics are restored in the transcript ("Telita" → "Telița")
- **Numbers are never trusted blind:** every figure goes into a form for the user to check (that is why data changes always prefill); in the normalised line and the Banner the figure is written in Mono next to its section and sex so a misheard number is seen before Salvează.

**Rules: what voice never does**
- Press Salvează or any save key (Mută, Înregistrează vânzarea, Adaugă, Creează seria, Salvează corecția, Salvează prețurile).
- Delete, deactivate, reactivate, retire, close a series, reset a password or undo a saved record; open or confirm a confirm Dialog (Where › Dialogs). Voice opens the record with that key focused.
- Set, reset, reveal or speak a password; sign in or sign out without the panel's one-tap confirmation.
- Anything outside the user's role, farms or time window (§1); prices for a vet or viewer, a price with a manager's delivery ("Prețul îl adaugă proprietarul în Prețuri hrană.").
- Act on another user's entry for a vet (D71); bulk actions the screens don't have.

**Accessibility**
- The key is a toggle `button` named "Comandă vocală" (`aria-label`; desktop tooltip adds the shortcut): `aria-pressed="true"` while listening (and while the panel is open on a result), `aria-busy="true"` while thinking, `aria-keyshortcuts="Control+Shift+Space"`; no `aria-expanded`/`aria-controls` (one state, pressed; the panel is a labelled region "Comandă vocală" announced through the live regions).
- **Shortcut Ctrl + Shift + Space:** starts listening from anywhere, even in a field (it types nothing) and **doesn't move focus**; pressed again while listening it stops; Esc cancels (before it closes a SidePanel, the S16 sheet or a menu, Flow 7). Chosen because: a modifier chord types nothing in fields; no default in Chrome, Edge, Firefox or Safari; not a screen-reader key (NVDA/JAWS use Insert/CapsLock, VoiceOver Ctrl+Option); "/" types in fields, is Firefox's quick find and is eaten by screen readers in browse mode; Ctrl/⌘ + Space switch input sources on Windows and macOS. *To review* (check on PCs where Ctrl+Shift switches keyboard layout, and ChromeOS, where Ctrl+Shift+Space switches input method: open question 20).
- **Live regions:** while the panel is open its polite region announces each state once: "Ascult." · "Mă gândesc." · an answer · a clarification question. Results that close the panel are announced outside it: executat by its Toast (the Toast's own status role: "Am deschis Rapoarte, Producție, ultimele 8 săptămâni."); formular precompletat and cancel by one app-level polite status region in the shell ("Am completat 2 câmpuri din ce ai spus. Verifică și salvează." / "Comanda a fost anulată."), as focus lands. Errors and refusals are `role=alert` in the panel. The meter is `aria-hidden` (the word carries the state).
- **Screen readers:** "Ascult." is announced before the microphone opens, and the panel says once "Cu cititorul de ecran pornit, folosește căștile sau scrie comanda." (its speech would be recorded); the text field gives the same results with no audio.
- **Focus:** a start by the key leaves focus on the key (so Space/Enter stops); a start by the shortcut leaves focus where it was; the panel is not a modal trap (desktop popover, phone sheet with scrim: focus moves into it only when a clarification, error or answer needs a choice); after navigation focus goes to the page h1; after prefill to the first filled field; on cancel back to where it was.
- Touch target 56 px phone, desktop key `control-h` (≥24 px); chips touch size on phone (§40).
- **Reduced motion:** the panel fades instead of rising/growing, in its container's own time (sheet 280 ms, popover 120 ms); the key's latch is colour only; the meter keeps lighting segments (a colour change, no movement).

**Motion (D196)**
- Phone panel = BottomSheet motion (rises its height, 280 ms); desktop popover = menu motion (scale 0.95 + opacity from the key's bottom-right corner, 120 ms); key press 0.97; latch and unlatch colour 120 ms; Banner and Toast as their components. The level meter is a live readout of the microphone, not an animation: segments light and go dark at the audio's pace (≤ 15 updates a second), no easing, no pulse; nothing else loops. The spinner is the Button's.

**Privacy**
- Audio is streamed for transcription and never stored: not on the device after the panel closes, not on the server, not by the provider (a contract condition, open question 13).
- Transcripts are not stored for show/filter/navigate commands or questions. When a save still carries at least one "din voce" mark, its `AUDIT_LOG` entry records `sursa: voce` and the transcript next to the saved values (so "who entered what" includes "said what"); a form whose marks were all edited away saves as typed (D199 note). On first use the panel shows one line: "Ce spui nu se păstrează. Dacă salvezi un formular completat din voce, textul comenzii rămâne în Istoric modificări." *To review* (open question 18).
- A usage count per user and day (no text) is kept for cost.

**Action catalogue (81 intents)**

Mode: **imediat** = runs at once, Toast + Anulează (export: no Anulează) · **răspuns** = written answer + the view that shows it · **precompletează** = the form opens filled, nothing saved until Salvează · **confirmă** = one tap in the panel. Roles: `vet` · `viewer` (Observator) · `manager` · `owner`; "viewer+" = viewer, manager, owner. Every intent also accepts the context defaults above. Example figures are the 18.11.2026 sample where real, else illustrative.

*Navigare*

| # | Intent | Examples | Parameters | Opens | Mode | Roles |
|---|---|---|---|---|---|---|
| N01 | Acasă | „du-mă acasă”, „deschide Acasă” | – | S2 | imediat | all |
| N02 | Înregistrează | „ce pot înregistra?”, „deschide Înregistrează” | – | phone: Înregistrează BottomSheet; desktop: Sidebar group open, focus on its first item | imediat | vet, manager, owner |
| N03 | Open an entry form, blank | „deschide Zilnic”, „vreau să dau hrana” | form, farm, date | S6–S10, S13 empty | imediat | per form (§1) |
| N04 | Seria | „deschide Seria 2026”, „seria la Producție” | series, tab | S3 (tab) | imediat | viewer+ |
| N05 | Fermă | „deschide Hîrbovăț”, „Telița ciclul trecut” | farm, cycle, tab | S4 (tab) | imediat | viewer+ |
| N06 | Secțiune | „deschide ha cinci punct unu”, „H6.2 la Păsări” | farm, section, tab | S5 (tab) | imediat | viewer+ |
| N07 | Tab | „treci la Hrană”, „arată Greutatea” | tab (Rezumat · Păsări · Hrană · Greutate · Producție) | the current scope page | imediat | viewer+ |
| N08 | Rapoarte · Producție | „deschide raportul de producție” | farm, period | S14 | imediat | viewer+ |
| N09 | Rapoarte · Hrană | „arată stocul de hrană”, „deschide livrările”, „consumul pe săptămâni”, „costul hranei” | part (Stoc · Livrări · Consum · Cost) | S11 at that part | imediat | viewer+ (Cost and lei: manager, owner, D64) |
| N10 | Prețuri hrană | „deschide prețurile” | – | S12 | imediat | owner |
| N11 | Introdus azi | „ce am introdus azi?”, „deschide Introdus azi” | – | S15 | imediat | vet, manager, owner |
| N12 | Lipsesc date | „arată ce lipsește” | farm, day | S2 Lipsesc date (focus on its heading); owner/viewer: De verificat | imediat | all |
| N13 | Istoric modificări | „deschide istoricul modificărilor” | – | S17 | imediat | manager, owner |
| N14 | Setări page | „deschide utilizatorii”, „ghidurile”, „tipurile de hrană”, „fermele”, „seriile” | page | S18–S22 | imediat | owner |
| N15 | Setări item | „deschide ghidul Ross 308 ouat”, „utilizatorul Ion Rusu”, „ferma Telița în setări” | item | S19 panel · S20 detail · S21 panel · S22 panel · S18 series | imediat | owner |
| N16 | Înapoi / închide | „înapoi”, „închide panoul” | – | the previous screen; closes a SidePanel or sheet (a form keeps its draft) | imediat | all |
| N17 | Cont / Mai mult | „deschide contul”, „mai mult” | – | phone: the tab bar's Cont (vet) or Mai mult (other roles); desktop: the Sidebar account | imediat | all |

*Filtre și tabel*

| # | Intent | Examples | Parameters | Opens | Mode | Roles |
|---|---|---|---|---|---|---|
| F01 | Perioadă | „ultimele 8 săptămâni”, „tot ciclul”, „13 săptămâni”, „de pe 1 până pe 15 noiembrie” | Ciclul curent · 4 · 13 săpt. · n săpt. · Interval | FilterBar (S4, S5, S14, S11 Consum); "8 săpt." = 4 săpt. widened once (D192); on S3 (no FilterBar, D63) a clarification Telița · Hîrbovăț opens that farm with the period | imediat | viewer+ |
| F02 | Pas | „pe zile”, „doar săptămânile” | Zi · Săpt. | FilterBar grain | imediat | viewer+ |
| F03 | Detaliu | „pe hale”, „pe compartimente”, „toată ferma” | Fermă · Hale · Compartimente (+ phone section) | the Detaliu switch (farm tabs except Rezumat) | imediat | viewer+ |
| F04 | Tipuri | „doar mortalitatea”, „arată și sacrificările”, „ascunde vânzările” | record types on/off (phone: one) | RecordChips | imediat | viewer+ |
| F05 | Ciclu | „ciclul trecut”, „ciclul din 2025 de la Hîrbovăț” | farm, cycle | CycleSelector (rescopes S4) | imediat | viewer+ |
| F06 | Serie | „treci pe Seria 2025” | series | series selector | imediat | viewer+ |
| F07 | Fermă în rapoarte | „producția de la Hîrbovăț” | farm | S14 / S11 farm filter | imediat | viewer+ |
| F08 | Încarcă săptămâni | „încă patru săptămâni”, „mai multe săptămâni”, „până la începutul ciclului” | ×4 or up to the first week | the load-more row pressed n times (D190); live "Au fost adăugate SAPT 52–55." | imediat | viewer+ |
| F09 | Deschide / închide săptămâna | „deschide SAPT 58”, „închide săptămâna 57” | week | its expand key | imediat | viewer+ |
| F10 | Mergi la o zi | „arată ziua de 14 noiembrie”, „ce s-a întâmplat ieri la H6.1” | date, section, tab | the scope page with that day's row loaded, open and focused | imediat | viewer+ |
| F11 | Deschide o înregistrare | „deschide vânzarea de azi”, „livrarea din 13 noiembrie” | type, date, section | its row in the Păsări log or S11 Livrări, focused (current row) | imediat | viewer+ |
| F12 | Filtre istoric | „modificările lui Ion Rusu din noiembrie” | user, farm, type, period | S17 filtered | imediat | manager, owner |
| F13 | Fără preț | „livrările fără preț” | – | S11 Livrări "fără preț" rows (owner: S12) | imediat | manager, owner |
| F14 | Resetează filtrele | „resetează filtrele”, „înapoi la ciclul curent” | – | Ciclul curent · Zi · Fermă · every chip | imediat | viewer+ |
| F15 | Mai vechi | „arată înregistrările mai vechi”, „mai vechi” | – | the Păsări log's Mai vechi (D81), the next page focused | imediat | viewer+ |

*Întrebări*

| # | Intent | Examples | Parameters | Opens | Mode | Roles |
|---|---|---|---|---|---|---|
| Q01 | Păsări în viață | „câte găini avem la Hîrbovăț?”, „câți cocoși sunt în H6.2?” | scope, sex, date | S4/S5 Rezumat ("♀ 12 561 · ♂ 794 · SAPT 59") | răspuns | viewer+; vet: its sections today, opens Zilnic |
| Q02 | Ouat față de ghid | „care e procentul de ouat săptămâna asta față de ghid?”, „cât a fost ouatul ieri?” | farm, hall, week/day | S14 ("Sapt. 59 (2 din 7 zile): ouat 40,4%, ghid 57,6% …") | răspuns | viewer+ |
| Q03 | Mortalitate | „câte găini au murit săptămâna asta?”, „mortalitatea cumulată față de ghid” | scope, period, sex | S4 Păsări, chip Mortalitate | răspuns | viewer+ |
| Q04 | Hrană dată | „câte grame dăm la găini azi?”, „cât s-a consumat ieri?” | scope, date, sex | S4 Hrană ("♀ 160 g · ghid 157 g") | răspuns | viewer+; vet: opens S7 |
| Q05 | Stoc | „cât stoc avem la Hîrbovăț?”, „pe câte zile mai ajunge hrana?” | farm, feed type | S11 Stoc | răspuns | viewer+ |
| Q06 | Greutate | „cât cântăresc găinile din H1 față de ghid?” | section, sex | S5 Greutate | răspuns | viewer+ |
| Q07 | Ouă și categorii | „câte ouă incubabile am avut ieri?”, „ce procent de categoria a doua?” | farm, hall, date, category | S4 Producție, Detaliu Hale | răspuns | viewer+ |
| Q08 | Cost | „cât ne-a costat hrana pe ciclul ăsta?”, „cât costă un ou?” | farm, cycle | S11 Cost | răspuns | manager, owner |
| Q09 | Săptămâna și vârsta | „în ce săptămână suntem?”, „câte zile au păsările?” | scope | no change ("SAPT 59 · vârstă 407 zile · 18.11.2026") | răspuns | all |
| Q10 | Ce lipsește | „ce lipsește azi?”, „unde n-am completat?” | farm, day | S2 Lipsesc date / De verificat | răspuns | all |
| Q11 | De verificat | „ce e de verificat?” | farm | S2 / S4 Rezumat De verificat | răspuns | viewer+ |
| Q12 | Livrări fără preț | „câte livrări n-au preț?” | – | S12 (manager: S11 Livrări) | răspuns | manager, owner |
| Q13 | Ce am salvat | „am salvat hrana azi?” | form, date | S15 | răspuns | vet, manager, owner |
| Q14 | Cine a modificat | „cine a modificat mortalitatea de pe 14 noiembrie?” | record | S17 filtered | răspuns | manager, owner |
| Q15 | Ciclu și ghid | „ce ghid are Hîrbovăț?”, „când a început ciclul?” | farm | S4 header and CycleSelector | răspuns | viewer+ |
| Q16 | Sacrificări, vânzări, transferuri | „câte găini am sacrificat luna asta?”, „câte păsări am vândut săptămâna trecută?”, „ce transferuri au fost în noiembrie?” | scope, type, period, sex | S4/S5 Păsări with that chip and period ("Sapt. 58: vândute ♀ 10 546 · ♂ 749.") | răspuns | viewer+ |
| Q17 | Livrări în perioadă | „câtă hrană am primit în noiembrie?”, „ce livrări au fost săptămâna asta?” | farm, feed type, period | S11 Livrări (lei for manager, owner) ("13.11.2026: 25 640 kg.") | răspuns | viewer+ |

*Formulare precompletate (nothing saved without Salvează)*

| # | Intent | Examples | Parameters | Opens | Mode | Roles |
|---|---|---|---|---|---|---|
| P01 | Mortalitate | „pune 3 găini moarte la ha cinci punct unu”, „la H6.2 au murit două găini și un cocoș” | farm, date, section, sex, count | S6, that EntryCard | precompletează | vet (today), manager, owner |
| P02 | Sacrificare + Motiv | „am sacrificat 2 găini la H5.2 pentru șchiopătare” | + Motiv (boală · șchiopătare · altul) | S6 Sacrificare + Motiv (missing: clarification) | precompletează | vet (today), manager, owner |
| P03 | Eroare sexare | „la H6.1 erau 4 cocoși printre găini” | section, sex, signed count | S6, Mai mult open | precompletează | vet (today), manager, owner |
| P04 | Ouă | „la hala 5: categoria unu 13, categoria doi 417, categoria trei 27, incubabile 3 415” | hall, Jos, Cat. 1–3, Incubabile, Masa medie | S6 hall card Ouă group (Total ouă live: 3 872) | precompletează | vet (today), manager, owner |
| P05 | Fără pierderi | „fără pierderi azi la H6.1”, „azi n-a murit nimic” | sections or all | S6 "Fără pierderi azi" ticked | precompletează | vet (today), manager, owner |
| P06 | Grame pe pasăre | „dă 160 de grame la toate găinile”, „la H5.1 cocoși 150 de grame” | farm, date, rows (section × sex or a whole sex), g 1–300 | S7 rows filled, kg live | precompletează | vet (today), manager, owner |
| P07 | Copiază ieri | „la fel ca ieri”, „copiază ieri” | – | S7 Copiază ieri pressed | precompletează | vet (today), manager, owner |
| P08 | Tip de hrană | „la cocoși pune hrana Cocoși” | rows, feed type | S7 Select | precompletează | vet (today), manager, owner |
| P09 | Cântărire | „cântărire H1 găini 2 553 de grame, CV 8, uniformitate 85, probă 100” | section, sex, date, g, CV %, uniformitate %, probă | S8 + live guide comparison | precompletează | vet (today), manager, owner |
| P10 | Transfer | „mută 7 001 găini din H2 în H7.1” | date, din, în, ♀, ♂, Toate | S9 + effect preview | precompletează | manager, owner |
| P11 | Vânzare | „vânzare azi la Hîrbovăț, toate păsările din H6” | farm, date, rows, Toate, buyer, note | S10 + effect preview; a total not split by section is listed on the Banner ("Ai spus ♀ 1 831 · ♂ 289 în total; scrie-le pe secțiuni.") | precompletează | manager, owner |
| P12 | Livrare hrană | „am primit 25 640 de kilograme pentru găini la Hîrbovăț pe 13 noiembrie” | farm, date, feed type, kg, note | S13 + stock after | precompletează | manager, owner |
| P13 | Preț livrare | „livrarea din 13 noiembrie: 9 lei 22 pe kilogram” | delivery, lei/kg or total | S12 row filled, the other value calculated | precompletează | owner |
| P14 | Modifică preț | „schimbă prețul livrării din 24 octombrie la 9,24 lei, motiv factură corectată” | delivery, price, Motiv | S11 Livrări row in Modifică (D69) | precompletează | owner |
| P15 | Corectează | „corectează mortalitatea de pe 14 noiembrie la H6.1, găini 5 în loc de 3, motiv numărate greșit” | record, field, new value, Motiv | S16 whole save, figure marked "modificat" + VoiceMark; also someone else's save from today (D70 B) | precompletează | manager, owner |
| P16 | Editează azi | „schimbă mortalitatea de azi la H5.1 în 4” | as P01–P09 | the form in Editează, only for the user's own save from today; someone else's save → P15 (D70 B) | precompletează | vet (own, until 23:59), manager (own), owner (own) |
| P17 | Începe ciclu nou | „începe un ciclu nou la Hîrbovăț de azi” | farm, date, guide, Motiv | B4 form Dialog opened prefilled; confirmed by hand | precompletează | manager, owner |
| P18 | Schimbă ghidul ciclului | „schimbă ghidul ciclului la Hîrbovăț” | farm, guide | its form Dialog opened prefilled | precompletează | manager, owner |
| P19 | Serie nouă | „serie nouă Seria 2027, la Telița pe 6 octombrie 2027, vârstă zero zile, H1 găini 14 080, H2 găini 14 000, H5-1 cocoși 3 510” | name, year, farm, date, Vârstă la populare 0–140, rows | S18 form + live SAPT line | precompletează | owner |
| P20 | Vârstă la populare | „schimbă vârsta la populare a Seriei 2026 la 7 zile” | series, days, Motiv | S18, its form Dialog opened prefilled (age + Motiv), confirmed by hand | precompletează | owner |
| P21 | Utilizator nou | „adaugă utilizatorul Petrea, login petrea, rol vet la Hîrbovăț” | name, login, role, farms (no password, D78) | S22 add panel | precompletează | owner |
| P22 | Rol și ferme | „dă-i lui Ion Rusu și ferma Telița”, „fă-l pe Petrea observator” | user, role, farms | S22 user panel | precompletează | owner |
| P23 | Fermă și secțiuni | „adaugă compartimentul H7.3 în hala H7”, „fermă nouă de creștere” | farm, type, guide, section code, parent | S19 panel / SectionTree | precompletează | owner |
| P24 | Ghid nou | „importă un ghid de ouat numit Ross 308 ouat 2027” | name, type | S20 Importă ghid (file or paste by hand) | precompletează | owner |
| P25 | Tip de hrană nou | „adaugă hrana Breeder 3 pentru găini” | name, sex | S21 panel | precompletează | owner |
| P26 | Închide seria | „închide Seria 2025” | series | S18 with Închide seria focused (confirm Dialog never by voice) | precompletează | owner |
| P27 | Editează tip de hrană | „tipul Cocoși e pentru cocoși”, „redenumește Breeder 3 în Breeder 3 plus” | feed type, name, for sex (retiring a type goes through its confirm step: never by voice) | S21 panel filled | precompletează | owner |
| P28 | Înlocuiește ghidul | „înlocuiește ghidul Ross 308 ouat” | guide | S20 › the guide › Înlocuiește (type fixed; file or paste by hand) | precompletează | owner |

*Altele*

| # | Intent | Examples | Parameters | Opens | Mode | Roles |
|---|---|---|---|---|---|---|
| X01 | Exportă | „exportă tabelul”, „descarcă în Excel” | – (the screen's filters) | Exportă .xlsx of the whole chosen period (D191); Toast without Anulează | imediat | viewer+ where Exportă exists |
| X02 | Ieși din cont | „ieși din cont”, „deconectează-mă” | – | the panel asks "Ieși din cont?" with chips Ieși · Rămâi → S1 | confirmă | all |
| X03 | Anulează ultima comandă | „anulează”, „nu asta” | – | the last voice result's Anulează / Anulează completarea | imediat | all |
| X04 | Ajutor | „ce pot să spun?”, „ajutor” | – | the panel lists 4 examples for this screen and role | răspuns | all |

*Never (refused in the panel, Rules above):* Salvează and every save key · Șterge, Dezactivează, Reactivează, Retrage, Închide seria, the Vânzare that empties a farm (no confirm Dialog is opened or confirmed by voice; the record opens with its key focused) · undo of a saved record · passwords (set, reset, show, say) · sign-in · prices for vet/viewer, price on a manager's delivery · another user's entry for a vet · anything outside role, farm or time window.

**Decision points (taken, to review)**
1. Sign-in screen has no voice; voice starts after sign-in.
2. Dialogs: no voice inside; form Dialogs may open prefilled, confirm Dialogs never open by voice (the record opens with the key focused).
3. Shortcut Ctrl + Shift + Space on every platform.
4. Second tap while listening = stop and run; cancel = Esc, Închide, Anulează, a tap while thinking.
5. Show commands run with no Execută step; the fix window is "Mă gândesc…" + Anulează on the Toast.
6. Answers are written only (no speech synthesis).
7. Log out by voice needs one tap in the panel; export runs at once.
8. A voice-filled suggestion row in Hrană zilnică counts as confirmed; a typed value replaced by voice shows "înainte n".
9. One form per command; a second form is offered as a chip.
10. More than 3 candidates: 3 chips + "Alta…" (the screen's selector).
11. Transcript kept in `AUDIT_LOG` only when a save still carries a "din voce" mark; one-line notice on first use.
12. Look: phone key round (an exception to "round = lamps and the centre key"), 56 px, `raised` + `line-strong` + `shadow-overlay` at rest, graphite `control` latched while listening; phone Toast stacks above it; desktop key an IconButton at the end of the PageHeader actions (components › Additions › VO00).
13. Level meter: 12 steel segments (lit `ink-secondary`, unlit `line`, D199 note 3), live at the audio's pace, kept under reduced motion.
14. The key is a toggle (`aria-pressed`), no `aria-expanded`; the shortcut doesn't move focus; Esc cancels voice before it closes a SidePanel or S16.
15. Phone: the floating key hides while the soft keyboard is open; 88 px `scroll-padding-bottom` on every phone view.


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
