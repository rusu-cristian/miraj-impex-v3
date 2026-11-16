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
  1. header (today, farm, guide week)
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
   - the guide-week readout per open cycle
   - clicking a block opens that farm at that cycle
3. **Farm summaries:** one panel per farm with its type, guide, cycle (sapt. N), alive ♀ · ♂ and key figures vs the guide.
4. **Tabs:** Păsări · Hrană · Greutate · Producție, aggregated per farm.

### S4 · Scope page · farm level (Hîrbovăț · ciclu Seria 2026) ← first build
1. **Page header:**
   - breadcrumbs, title "Hîrbovăț"
   - type and guide tags
   - meta: ♀ · ♂ alive · sapt. 41 · from 10.02.2026
   - actions: Mută păsări, Vânzare, and "Ciclu ▾" (Începe ciclu nou, Schimbă ghidul ciclului)
2. **Cycle selector + compact cycle track for this farm:** past cycles of any series. Selecting one rescopes the page.
3. **Filter bar:** period, grain, record-type chips, export.
4. **Tabs.**
5. **Rezumat contents:**
   - ⭐ **a. Key readout:** a dark display with the farm's key figure for the cycle type (laying: % lay vs guide; growth: weight ♀ vs guide), plus a small guide chart.
   - **b. Stats vs guide:**
     - alive ♀ · ♂
     - cumulative mortality %
     - weight ♀ / ♂
     - feed g/bird today
     - feed days left
     - % lay (laying)
   - **c. Section overview:**
     - one tile per section (SectionTile): code, ♀ · ♂ alive, mortality this week, grams today, weight vs guide, % lay
     - a missing-data lamp
     - grouped by hall when compartments exist
     - clicking a tile opens the section
     - a totals row
   - **d. De verificat:** alerts for this farm.
   - **e. Latest records:** the event log with record icons.
- **Other tabs:** as in spec D1.

### S5 · Scope page · section level
Same template. Rezumat shows the section's stats, the latest records and the guide charts. There are no child tiles.

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
5. **Save bar:** "8 secțiuni · 3 412 kg", Salvează.

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
4. **Consum:** a weekly table per section with g/bird vs guide.
5. **Cost:** per cycle, per bird and per egg from priced deliveries, plus a line for the unpriced kg.

### S12 · Prețuri hrană (owner)
A list of unpriced deliveries, lei/kg ↔ total inputs, Salvează prețurile, and a "toate au preț" empty state.

### S13 · Livrare hrană (manager)
Farm, date, feed type, kg, note, the stock-after line, Adaugă.

### S14 · Producție report
A weekly table by guide week, per section, vs the guide. Export.

### S15 · Introdus azi
Today's entries grouped by record type; edit or delete.

### S16 · Corectează (side panel or sheet)
The entry with its record icon, editable fields, an effect preview, the reason, Salvează corecția / Șterge.

### S17 · Istoric modificări
Filters (user, farm, record type, period), then the change list with record icons.

### S18 · Serii (+ Serie nouă)
The series list with each series' cycles. The new-series form: name, year, growth farm, date, placement rows per section ♀ · ♂, Creează seria.

### S19 · Ferme și secțiuni
- Farm list with type and guide.
- Farm panel: name, type, guide (filtered by type), section tree, add section, retire.

### S20 · Ghiduri
Guide list with type. Import panel (paste or file, then a preview with row errors). Guide detail as a weekly table plus small curves.

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
