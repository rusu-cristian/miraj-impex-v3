# Impex Line: navigation and screens v2 (shared)

Every screen is an app screen (Impeccable mode **Operate**) and is built code-first.
⭐ marks a Moment section: the cycle track at full width, or a guide chart.
The record-type code (icon + colour) appears on every event, column, chart line and filter.

## Navigation map
```
Autentificare
└─ App
   ├─ Acasă                         (content by role)
   ├─ Înregistrează
   │  ├─ Zilnic                     (vet+)
   │  ├─ Hrană zilnică              (vet+)
   │  ├─ Cântărire                  (vet+)
   │  ├─ Transfer                   (manager+)
   │  ├─ Vânzare                    (manager+)
   │  ├─ Livrare hrană              (manager+)
   │  └─ Introdus azi
   ├─ Serie [series selector]       (viewer+)
   │  └─ Seria 2026 › Fermă [ciclu] › Secțiune
   │     tabs: Rezumat · Păsări · Hrană · Greutate · Producție*
   ├─ Rapoarte                      (viewer+)
   │  ├─ Producție
   │  └─ Hrană (Stoc · Livrări · Consum · Cost)
   │     └─ Prețuri hrană           (owner)
   └─ Setări
      ├─ Serii (+ Serie nouă)       (owner)
      ├─ Ferme și secțiuni          (owner)
      ├─ Ghiduri                    (owner)
      ├─ Tipuri de hrană            (owner)
      ├─ Utilizatori                (owner)
      └─ Istoric modificări         (manager+)
```
\* Producție appears only for laying cycles.

**Desktop sidebar:** wordmark, series selector, Acasă, Înregistrează (expands), Serie, Rapoarte, Setări, then the account.

**Phone tab bar:**
- vet: Acasă · Înregistrează (centre key) · Cont
- other roles: Acasă · Serie · Înregistrează · Rapoarte · Mai mult

## Screens

### S1 · Autentificare
Wordmark, Login, Parolă, Intră. States are in spec A1.

### S2 · Acasă
- **Vet:**
  1. header (today, farm, the birds' week; round 2: "18.11.2026 · Hîrbovăț · SAPT 59", D151/D152)
  2. Lipsesc date (per section, with the record icon for Zilnic or Hrană)
  3. entry keys: Zilnic, Hrană zilnică, Cântărire
  4. Introdus azi
- **Manager:**
  1. header
  2. De verificat (alerts)
  3. feed days left per farm × type
  4. entry keys (adds Transfer, Vânzare, Livrare hrană)
  5. Lipsesc date
- **Owner:**
  1. ⭐ compact cycle track for the active series
  2. De verificat
  3. Livrări fără preț (count + link to S11)
  4. feed stock
  5. farm summaries

### S3 · Scope page · series level (Seria 2026)
1. **Page header:** breadcrumbs, title, totals ♀ · ♂ across farms.
2. ⭐ **Cycle track (full):**
   - one lane per farm, with each cycle as a block
   - transfer links between lanes
   - the "azi" line
   - the SAPT readout per open cycle (the birds' age week, round 2, D151: "SAPT 59")
   - clicking a block opens that farm at that cycle
3. **Farm summaries:** one panel per farm with its type, guide, cycle (round 2: its SAPT span, "SAPT 1–21" / "SAPT 20–59"), alive ♀ · ♂ and key figures vs the guide.
4. **Tabs:** Păsări · Hrană · Greutate · Producție, aggregated per farm. Round 2 (D154, spec D1 › Series page), each tab in this order:
   1. ⭐ **Table** (PeriodTable, no Detaliu switch): the last 4 weeks, SAPT 59–56 (newest first, D159/D160), with their days behind the expand key, a "Fermă" column, one "Total serie" group; no FilterBar (D63), Exportă in the header; under the table the key "Arată toate săptămânile" (SAPT 59 → 1, D160)
   2. the tab's charts

### S4 · Scope page · farm level (Hîrbovăț · ciclu Seria 2026) ← first build
1. **Page header:**
   - breadcrumbs, title "Hîrbovăț"
   - type and guide tags
   - meta: ♀ · ♂ alive · SAPT 59 · from 22.02.2026 (round 2: ♀ 12 561 · ♂ 794, the birds' age week, D151)
   - actions: Mută păsări, Vânzare, and "Ciclu ▾" (Începe ciclu nou, Schimbă ghidul ciclului)
2. **Cycle selector + compact cycle track for this farm:** past cycles of any series. Selecting one rescopes the page.
3. **Filter bar:** period (default 4 săpt., D159), grain, record-type chips, export.
4. **Tabs** (Rezumat · Păsări · Hrană · Greutate · Producție; every tab opens on its table, round 2, D154).
5. **Rezumat contents** (round 2 order):
   - ⭐ **a. Table** (PeriodTable, "Total fermă" columns, no Detaliu switch): week rows newest first with their days (Data · Vârstă (zile) · SAPT), last 4 weeks, foot "Total perioadă" (spec D1 › Rezumat, D159).
   - ⭐ **b. Key readout:** a dark display with the farm's key figure for the cycle type (laying: % lay vs guide; growth: weight ♀ vs guide), plus a small guide chart.
   - **c. Stats vs guide:**
     - alive ♀ · ♂
     - cumulative mortality %
     - weight ♀ / ♂
     - feed g/bird today
     - feed days left
     - % lay (laying)
   - **d. Section overview:**
     - one tile per section (SectionTile): code, ♀ · ♂ alive, mortality this week, grams today, weight vs guide, % lay
     - a missing-data lamp
     - grouped by hall when compartments exist
     - clicking a tile opens the section
     - a totals row
   - **e. De verificat:** alerts for this farm.
   - **f. Latest records:** the event log with record icons.
- **Other tabs** (round 2): 1. the Detaliu switch (Fermă · Hale · Compartimente, default Fermă, D159) and the table (PeriodTable: Total fermă, + hall or compartment groups; chips choose column groups; newest week first, last 4 weeks by default); 2. the tab's charts; 3. Păsări only: the event log with the type chips. Contents as in spec D1.

### S5 · Scope page · section level
Same template. Rezumat shows the section's stats, the latest records and the guide charts. There are no child tiles.
Round 2 (D154): every tab opens on its table (PeriodTable, one group = this section ♀ · ♂); Rezumat order = table · Stats · guide charts · latest records; other tabs = table · charts (· event log on Păsări).

### S6 · Zilnic (phone first)
1. **Form header:** "Zilnic", farm, date, sapt.
2. **Section cards** (EntryCard):
   - alive ♀ · ♂
   - count inputs coloured by sex for Mortalitate and Sacrificare (with a reason)
   - "Mai mult": Eroare sexare
   - Ouă group (laying, on the hall card)
   - "Fără pierderi azi"
3. **Save bar:** a summary with record icons, then Salvează.

### S7 · Hrană zilnică (phone first) ← build early
1. **Form header:** "Hrană zilnică", farm, date, sapt.
2. **Copiază ieri** key and the guide hint for this week.
3. **Allocation rows:** section × sex. Each row has:
   - the sex icon and colour
   - birds alive
   - the feed type select
   - a g/pasăre input, with the suggestion state until confirmed
   - kg live
   - the guide figure
   - a deviation tag
4. **Totals per feed type:** kg today, stock after, days left.
5. **Save bar:** "8 rânduri · 2 457,9 kg", Salvează (round 2: 18.11.2026, the 4 compartments with birds H5.1 H5.2 H6.1 H6.2 × ♀ · ♂, H7 empty; ♀ 2 303,2 + ♂ 154,7 kg; rows, never "secțiuni", D40).

### S8 · Cântărire
Section, sex, values, live guide comparison, Salvează.

### S9 · Transfer
Din (section select with alive counts), În, ♀/♂ counts (Toate), effect preview including cycle open/close, Mută.

### S10 · Vânzare
Farm, date, rows per section with ♀/♂ counts (Toate), buyer, note, effect preview, a confirmation dialog when it empties the farm.

### S11 · Hrană report (Stoc · Livrări · Consum · Cost) ← build early
1. **Page header:** "Hrană" and the action Livrare hrană (manager+).
2. ⭐ **Stock readouts:** one per farm × feed type, showing kg, days left, run-out date and a lamp.
3. **Livrări table:**
   - date, farm, type, kg, lei/kg, total, priced by
   - unpriced rows show the info tag "fără preț"
   - the owner edits inline
4. **Consum:** a weekly table per section with g/bird vs guide. Round 2 (D154): week row (farm) → its 7 daily farm rows when open → section week rows; grain and order as spec D1.
5. **Cost:** per cycle, per bird and per egg from priced deliveries, plus a line for the unpriced kg.

### S12 · Prețuri hrană (owner)
A list of unpriced deliveries, lei/kg ↔ total inputs, Salvează prețurile, and a "toate au preț" empty state.

### S13 · Livrare hrană (manager)
Farm, date, feed type, kg, note, the stock-after line, Adaugă.

### S14 · Producție report
A weekly table by the birds' age week (SAPT), per section, vs the guide. Export.
Round 2 (D154, spec D3), in this order: 1. filters (Fermă, period, grain Zi / Săpt.); 2. ⭐ the table (newest week first, last 4 weeks by default, D159): per week the farm's week row with expand key → its 7 daily rows when open → hall week rows H5 · H6 · H7; foot "Total perioadă"; 3. the % ouat Readout and Stats (D92, after the table). Export = what's on screen.

### S15 · Introdus azi
Today's entries grouped by record type; edit or delete.

### S16 · Corectează (side panel or sheet)
The entry with its record icon, editable fields, an effect preview, the reason, Salvează corecția / Șterge.

### S17 · Istoric modificări
Filters (user, farm, record type, period), then the change list with record icons.

### S18 · Serii (+ Serie nouă)
The series list with each series' cycles. The new-series form: name, year, growth farm, date, placement rows per section ♀ · ♂, Creează seria.
Round 2 (D151, spec B3): the list shows each series' SAPT span and today's age ("SAPT 1–59 · azi 407 zile") and each cycle's span ("Hîrbovăț · din 22.02.2026 · SAPT 20–59 · deschis"), plus a "Vârstă la populare" column. Form order: name, year, growth farm, date, **Vârstă la populare (zile)** (default 0, hint, live line "La populare: vârstă 0 zile · SAPT 1."), placement rows, Creează seria.

### S19 · Ferme și secțiuni
- Farm list with type and guide.
- Farm panel: name, type, guide (filtered by type), section tree, add section, retire.

### S20 · Ghiduri
Guide list with type. Import panel (paste or file, then a preview with row errors). Guide detail as a weekly table plus small curves.
Round 2 (D151, spec B2): weeks are age weeks: the list shows each guide's span ("Ross 308 ouat · SAPT 19–65", "Ross 308 creștere · SAPT 1–25"); the detail table's SAPT column is the age week (no "SAPT = vârstă − 18" note); the weight foot note names the SAPT − 1 comparison (D157); the import preview warns on a laying guide that starts at week 1.

### S21 · Tipuri de hrană
List (name, for sex, active). Add or edit.

### S22 · Utilizatori
See spec appendix A2, with the roles vet / viewer / manager / owner.

## Build order
1. **S4** farm level, Rezumat (desktop): section overview, readout, stats, filter chips, cycle selector.
2. **S7** Hrană zilnică (phone).
3. **S3** series level with the full cycle track.
4. **S11** Hrană report, then **S12** Prețuri hrană, then **S13** Livrare hrană.
5. **S6** Zilnic (phone).
6. **S4** remaining tabs (Păsări with the type-filtered event log, Greutate, Producție).
7. **S9** Transfer, then **S10** Vânzare.
8. **S2** Acasă (vet, manager, owner).
9. **S14**, **S8**, **S15**, **S16**, **S17**.
10. **S18**–**S22**, then **S1**.
