# Handoff
Updated: 2026-11-16 22:15 · Last commit: 6c15ca0
Repo: rusu-cristian/miraj-impex-v3 · Branch: main · Target: Design canvas "Impex Line Panel · Ecrane" (https://claude.ai/artifact/VC8kqksPgxxYo9kPpK14yv, slug e44dd7fb-2990-41f9-a791-2841f1b1dcd1, live version 1794856515-f4a7) built from the Impex Line Panel design system (https://claude.ai/artifact/KWvk2DHkGgVsDKdhwybfgd) · Setup: full, canvas mode

## Progress
Done 25 / 64 · Screens done: S4 (+8 states, polished), S7 (+7 states, polished), S3 (+3 states, polished), S11 (+4 states; S11P next)

## Just finished
- S11 Rapoarte · Hrană (owner, manager, observator without prices per D64, loading, empty period)
- S3P published by the coordinator (helper publish approval went unanswered)

## Next up
- S11P (polish + audit), then S12 → S12P → S13 → S13P; SYS05 is done so S6 can follow S13P; SYS07 done (S10/S18 ready once their order comes)

## Waiting on the user
- The user is away (since 16.11.2026 17:48) and asked for every choice to be made by default and reviewed later: D14–D64 in design/kit/decisions-current.md are marked "default (user away) — to review". Flagged for attention: D24 (sapt. 40 not 41), D25 (Hîrbovăț cycle starts 10.02.2026 though transfers are 22–28.02), D27 (illustrative 2026 sample after 28.02.2026), D53 (literal 3× mortality rule), D59 (Corecție vs Mortalitate tiles look alike in grey), D63 (Telița "Diferență registru" ♀ −611 · ♂ −210), D64 (viewers see no prices).
- No screen gates were held (D33); the user reviews all artboards on return.

## Things to remember
- Run loop jobs one screen at a time; system jobs (SYS..) can run in parallel with a screen. Canvas publishes are serial: helpers fetch the version first and retry on conflict; if a helper's publish needs an approval that doesn't come, the coordinator publishes the listed files with publish_files (paths under design/canvas/project/, "as" = project/<file>).
- Helpers may be refused loading the impeccable/intent skills; they then work from craft.md, PRODUCT.md and DESIGN.md.
- Generators live in design/canvas/scripts/ (s4s_build.py holds the shared desktop shell; any shell change regenerates S3/S4/S11 boards). Phone screens: s7_build.py pattern. Shots use fallback fonts (Google Fonts blocked here).
- New components so far (copy into the system artifact later; specs in components.md › Additions, CSS in design/system/components/additions.css): CycleTrackFarm (+ .ded), SectionTotals, AlertRow, GuideMissing, SeriesSelector, FilterInterval, FeedTotals, FarmSummary (+ --stretch), StockReadout, CostBlock, EntryCardPlus, HallCard, GuideChartVariants (phone/wide drawings), RecordCodePlus (Corecție, cycle marks), SexCountRow (--pick), chip --split, ip-alloc--phone, interaction states.
- FIN backlog: system-wide 10px SVG track text and Mono th.num (noted in S3P.md); S5 section level not queued (not requested).
