# Impex Line

## Users
- **Vets** enter daily data per section on phones: mortality, culls, sexing errors, grams per hen and rooster, weighings, eggs.
- **Farm managers** record feed deliveries (no price), transfers, sales and new cycles, view their farms and correct past data.
- **Viewers** read the views of their farms.
- **The owner** sees all farms, prices feed deliveries, and manages farms, guides and users.

## Product
An internal web app that replaces the per-series Excel and Google Forms workflow.
- **Series and cycles:** it tracks yearly series through farm cycles (growth at Telița → transfer → laying at Hîrbovăț → sale). Each farm follows its linked guide from its cycle start.
- **Counts:** calculated from events per section.
- **Feed:** consumption comes from the vet's daily grams, stock is kept per feed type, and cost comes from owner-priced deliveries.
- **Navigation:** every view drills down from series to farm (cycle) to section.
- **Record-type code:** every record type has a fixed icon and colour.

## Constraints
- Interface in Romanian only.
- Vue / Nuxt.
- Online only.
- Phone first for entry, desktop for views.
- WCAG 2.2 AA.
- Light theme only.
- Incubator and offline entry are out of scope.

## Ethical stance
- No dark patterns.
- No tracking beyond the audit log of data changes.
- Vets can't see feed prices.
- The change log is visible to managers and the owner, and entries are attributed to named users.

## Success
- 95% of hall-days are entered by 12:00.
- The median daily entry takes under 2 minutes.
- Bird counts match the spreadsheet during the parallel run.
- The owner answers "how is a hall doing against the guide" in 3 clicks or fewer, and stops opening the workbook.

## Open questions
1. Should a farm only ever hold one series at a time? The default allows a manual new cycle while birds remain.
2. Is the egg total formula right? The default is Cat. 1 + Cat. 2 + Cat. 3 + Incubabile, with "Jos" kept separate. To confirm with Petrea.
3. Is the Hîrbovăț productivity guide available, and in what format?
4. How will the Seria 2026 history be imported, and from which date does the app become the source of truth?
5. Are the alert thresholds right? Defaults: mortality above 3× the hall's 7-day average, weight ±5%, lay ±3 points, feed running out within 5 days.
6. Should copy use the informal "tu" (the default) or the formal "dumneavoastră"?
