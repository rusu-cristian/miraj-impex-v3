# Impex Line: functional spec v2 (shared by both designs)

Status: 16.11.2026. This replaces spec v1.
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

**Cycles and the guide week**
- A farm follows its linked guide from the start of its current cycle. The guide week is `floor((date − cycle.start) / 7) + 1`.
- A cycle starts automatically when birds arrive in an empty farm, or with **Începe ciclu nou**.
- A cycle ends when the farm reaches 0.
- See `data-model.md` › Cycle rules.

**Birds alive**
- Birds alive is the sum of bird events, per section and sex.
- No save may make a section go below 0 on any day.

**Feed**
- Daily kg = grams × birds alive. There's one entry per section, sex and day; a missing day is flagged, not carried over.
- Stock is kept per farm and feed type.
- Cost counts priced deliveries only. Unpriced kg are shown separately.

**Guide comparison**
- Comparisons are by guide week and sex.
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
- Dates: `dd.mm.yyyy · sapt. 41` (guide week of the farm's cycle).

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
- **Rules:**
  - Duplicate week + sex rows are refused.
  - A guide used by a cycle can be replaced but not deleted.
- **States:** parse errors per row; a partial guide shows `–` with "fără date în ghid".

### B3 · Series — MVP
- **Job:** When a new year's flock arrives, I want to start a series and place the birds into the growth farm's sections.
- **Flow:**
  1. Name and year (*sample:* Seria 2027 · 2027).
  2. Placement: the growth farm, date, then rows of section · ♀ · ♂ (*sample:* H1 ♀ 14 080 · H2 ♀ 14 000 · H5-1 ♂ 3 510).
  3. Creează seria.
- **Result:** the placement events, plus a new cycle in the growth farm (auto) with its guide.
- **Rules:** if the farm still holds birds from another series: "Telița are încă 26 881 păsări din Seria 2026. Mută-le sau vinde-le înainte, sau începe un ciclu nou manual."
- **Closing:** Închide seria (owner) after its last cycle ends. A closed series is read-only.

### B4 · Cycles — MVP (manager, owner)
- **Job:** When birds arrive or a farm restarts, I want the farm to follow its guide from week 1, so comparisons line up.
- **Entry points:** a farm page › Ciclu (cycle panel); the cycle track (click a cycle block).
- **Automatic:**
  - A cycle opens when birds arrive in an empty farm. The Toast reads "Ciclu nou la Hîrbovăț · Seria 2026 · sapt. 1 de la 10.02.2026 · ghid Ross 308 ouat."
  - It closes when the farm reaches 0.
- **Manual:**
  - **Începe ciclu nou:** a dialog with the date (default today), the guide (default the farm's) and the note "Ciclul curent (sapt. 41) se închide pe 15.11.2026. Săptămâna ghidului pornește de la 1." It requires a reason.
  - **Schimbă ghidul ciclului:** applies from the cycle start, so charts are recompared.
- **Rules:**
  - One open cycle per farm.
  - A cycle can't start before the farm's last event of the previous cycle.
- **States:**
  - farm empty, no cycle: "Ferma e goală. Ciclul pornește la prima populare sau transfer."

## 4. Entry

All forms work at 360px width. Each one shows the farm, the date and the guide week, and has a sticky Salvează.

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
  - *sample:* H1.1 ♀ 9 402 × 165 g = 1 551,3 kg
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
  - If the destination farm is empty, a cycle opens there; the preview says so ("Ciclu nou la Hîrbovăț, sapt. 1").
  - If the source farm reaches 0, its cycle closes; the preview says so.
  - Undo is available the same day (manager).

### C4 · Vânzare (manager) — MVP
- **Job:** When the laying flock is sold, I want to record how many birds left, so the farm empties and its cycle closes.
- **Flow:**
  1. Farm, then date.
  2. One row per section with ♀ and ♂ counts. "Toate" fills everything that's alive.
  3. Optional buyer and note.
  4. An effect preview ("Hîrbovăț: ♀ 26 881 · ♂ 2 994 → 0 · ciclul se închide").
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
- **Live comparison:** "ghid sapt. 41: 3 460 g · 3 412 g (−1,4%)".

### C8 · Introdus azi — MVP
- The user's entries today, grouped by record type with their icons. Edit or delete until 23:59.

### C9 · Lipsesc date — MVP
- Sections with birds but no Zilnic entry or no Hrană zilnică entry for a day. Shown from 12:00, for up to 7 days back.

## 5. Views (viewer and up)

### D1 · Scope pages: Seria › Fermă (ciclu) › Secțiune — MVP
- **Header:** breadcrumbs, title, alive ♀ · ♂, guide week.
- **Series level:** the cycle track (both farms, all cycles of the series), then a summary per farm.
- **Farm level:**
  - the cycle selector (the current cycle by default, past cycles listed)
  - a section overview: one tile per section, showing:
    - ♀ · ♂ alive
    - mortality this week
    - grams today
    - weight vs guide
    - % lay (laying)
  - farm totals = the sum of the sections
- **Section level:** the same tabs, scoped to one section.
- **Tabs:**
  - **Rezumat:** key figures vs guide, alerts, children.
  - **Păsări:** alive and mortality per day or week, culls, transfers, sales, the event log with type filter chips.
  - **Hrană:** g/bird vs guide, kg per day, stock and days left at farm level.
  - **Greutate:** weight vs guide, CV, uniformity.
  - **Producție:** laying only. % lay, % hatching, egg weight, categories.
- **Filters:** period (Ciclul curent · 4 săpt. · 13 săpt. · Interval), grain (Zi / Săptămână), record types, export .xlsx.
- **States:**
  - loading
  - no data in the period
  - no guide
  - no access
  - an empty farm between cycles: "Ferma e goală din 30.09.2026. Ultimul ciclu: Seria 2025, 64 săpt."

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
- A weekly table per farm › section, by guide week.
- Columns: hens alive, eggs, % lay vs guide, hatching eggs and %, egg weight, mortality, feed g/hen.
- Export.

### D4 · Hrană report — MVP
1. **Stoc:** farm × feed type: kg, average per day, days left, run-out date.
2. **Livrări:** all deliveries with their price status (with price / fără preț). The owner can edit inline.
3. **Consum:** per week × section: kg, g/bird vs guide.
4. **Cost:** lei per cycle, per bird, per egg, from priced deliveries, plus a line for the "X kg fără preț" that isn't included.
- Prices are hidden from vets. Viewers and managers see them read-only.

### D5 · Export — MVP
- Every table can be exported to .xlsx with the current filters.

### D6 · Compare cycles — later
- The same farm, different series, by guide week.

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
  1. Open the entry and change the values.
  2. A reason is required (≤ 200 characters).
  3. A preview shows the effect: "H2 · găini în viață de pe 03.11.2026 se schimbă cu +12."
  4. Salvează corecția.
- **Actions:** Șterge, with the reason required and a confirmation dialog.
- **Rules:**
  - The same "never below 0" check runs over every later day.
  - The old and new values go to `AUDIT_LOG`.
- **States:** refused: "Corecția ar lăsa H6 cu −4 găini pe 16.02.2026."

### E2 · Change log — MVP
- **Entry points:** Setări › Istoric modificări, and on each entry ("modificat de Ion · 16.11.2026, 09:40").
- **Content:** a list with date, user, table, hall, old → new, reason. Filter by user, farm and period.
