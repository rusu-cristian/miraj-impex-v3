# Product

<!-- impeccable:product-schema 1 -->

## Platform
adaptive: phone for daily entry, desktop for views and reports. It must work at 360px and up.

## Stack
Vue / Nuxt. Online only, with a PWA install for the phone home screen. The data model is in `spec.md` §0.

## Users
- **Vet** (veterinarul fermei). Walks the sections every morning and enters, on a phone, for each section:
  - mortality, culls (sacrificare) and sexing errors
  - the grams given per hen and per rooster
  - weighings
  - eggs, at the laying farm

  Today this is spread across an Excel workbook and Google Forms. A typo silently breaks every formula after it.
- **Farm manager.** Records feed deliveries (type, date, kg), transfers between farms and sales, and starts a new cycle when needed. Checks their farms against the guide and corrects past data.
- **Viewer.** Read-only access to the views of assigned farms.
- **Owner.** Sees every farm and adds feed prices once invoices arrive. Wants to answer "how is this farm doing against its guide this cycle" and "how much feed did we buy, use and still have, and what did it cost" without opening the workbook.

## Product Purpose
Replace the per-series spreadsheets (Seria 2026: Telita, Hrana Telita, H1…H52 weight charts, sapt, Hrana Hîrbovăț, DataHirbovat, Raport Nou 2026) with one app.
- **Farms** have a type (growth or laying), are split into sections, and are linked to a guide of the same type.
- **Counts:** the app keeps a correct running count of hens and roosters per section from daily events.
- **Cycles:** a cycle starts when birds first arrive in an empty farm, or from a manual "Începe ciclu nou". From that day the farm follows its guide from week 1.
- **Feed:** the app calculates consumption from the vet's daily grams per section, and stock per feed type from the manager's deliveries. Cost comes once the owner has priced them.
- **Comparison:** weight, feed, mortality and production are compared with the guide for every cycle, the same way every year.

## Positioning
An internal tool for one poultry business. It isn't sold, so there's no pricing and no marketing site. It's different from the spreadsheet because:
- counts are calculated from events, so they can't drift
- each new series starts empty instead of being copied
- every view can be narrowed from the series to the farm to the hall

## Operating Context
- **Farms:**
  - Telița: growth, sections H1, H2, H3, H6, H5-1, H5-2
  - Hîrbovăț: laying, halls H1, H3 and H4, each split into sections such as H1.1 and H1.2
- **The yearly cycle:** each year's series is placed at Telița, transferred to Hîrbovăț around week 19, then sold. When Hîrbovăț receives the birds, Telița can receive the next series, so two series can run at the same time in different farms.
- **Daily entry:** happens in the morning, inside or next to the halls. Connectivity is reliable enough for online-only.
- **Views:** checked on an office PC and on phones.
- **Incubator:** a separate module, later.

## Capabilities and Constraints
- **Interface language:** Romanian only, with full diacritics.
- **Roles:** vet, viewer, manager and owner. Each role except the owner is limited to its assigned farms.
- **Counts** are never allowed below 0 on any day.
- **Edits to past data** need a manager, a reason and an audit entry.
- **Feed prices** are entered by the owner after delivery and hidden from vets.
- **Out of scope:** offline entry, notifications, the incubator. The Seria 2026 history import is a one-time developer migration.

## Brand Commitments
- **Name:** Impex Line.
- **No logo exists.** The name is used as a wordmark.
- **Copy:** plain Romanian, informal "tu", verbs on buttons. No hype, no exclamation marks.

## Evidence on Hand
- **The `Seria_2026.xlsx` workbook:**
  - real hall codes and field structures
  - placement figures (H1 G 14 080, H2 G 14 000, H5-1 C 3 510)
  - daily mortality
  - feed deliveries (07.10.2025, 8 000 kg, 66 800 lei)
  - weekly weighings
  - Hîrbovăț egg data
- **Missing:** the Hîrbovăț productivity guide sheet, the "Raport Nou 2026" sheet, and the incubator data. There's also no logo, no photos, and no user feedback yet. None of these may be invented.

## Product Principles
1. **Counts come from events.** Every number can be traced back to who entered what, and when.
2. **Every record type is recognisable at a glance**, by the same icon and colour everywhere.
3. **Every figure is placed in its cycle** (date and guide week) and **against the farm's guide.**
4. **Entry is fast and forgiving on a phone.** Empty isn't zero, warnings don't block, and refusals say exactly why.
5. **One layout at every level.** Drilling down narrows the scope; it never changes the page.
6. **The owner never needs the spreadsheet again.** Everything exports to .xlsx.

## Accessibility & Inclusion
- **Standard:** WCAG 2.2 AA.
- **Contrast:** 4.5:1 for text, 3:1 for controls and focus.
- **Touch targets:** at least 44px on touch screens.
- **Meaning** is never carried by colour alone.
- **Charts** have keyboard readouts and a data table for screen readers.
- **Motion:** reduced-motion is respected.
- **Locale:** Romanian, with no RTL needs.
