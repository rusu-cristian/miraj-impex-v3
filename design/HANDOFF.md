# Handoff
Updated: 2026-11-16 23:05 · Last commit: d1bd1e3
Repo: rusu-cristian/miraj-impex-v3 · Branch: main · Target: Design canvas "Impex Line Panel · Ecrane" (https://claude.ai/artifact/VC8kqksPgxxYo9kPpK14yv, slug e44dd7fb-2990-41f9-a791-2841f1b1dcd1, live version 1794856515-f4a7) built from the Impex Line Panel design system (https://claude.ai/artifact/KWvk2DHkGgVsDKdhwybfgd) · Setup: full, canvas mode (canvas read-only from this session, see below)

## Progress
Done 27 / 64 · Screens done: S4, S7, S3, S11 (all polished) · Building: S12 (fix round 1)

## Just finished
- SYS08 Acasă blocks: MissingRow, EntryKeys, UnpricedCount (2 fix rounds; defaults D65, D66); no canvas boards, system previews in design/system/components/<Name>/
- S11P Rapoarte · Hrană polish + audit (1 fix round: dates/counts in ip-num, observer Livrări table fit); additions.css §22 th.num back to Sans (commits with SYS08)

## Next up
- S12 → S12P → S13 → S13P → S6 (SYS05 done)

## Waiting on the user
- None blocking. 16.11.2026 resume: user chose to keep defaults D14–D64 and continue without screen gates; review everything later (flagged: D24, D25, D27, D53, D59, D63, D64).

## Unpublished on the canvas (publish with publish_files, mode patch, "as" = project/<file>)
- Since 16.11.2026 this session has network access off and did not make the canvas, so no tool here can write to it (get_artifact → network_off_limit). Publish from a chat opened from the canvas page, or from a session with network access on.
- project/additions.css now also holds §23–25 (SYS08)
- project/S11-rapoarte-hrana.dc.html, S11-rapoarte-hrana--se-incarca.dc.html, --manager, --observator, --fara-livrari-in-perioada; project/additions.css; project/canvas.json
- Add every later board (S12…) and each additions.css / canvas.json change here.

## Things to remember
- One screen at a time; SYS jobs can run in parallel and append to additions.css / components.md › Additions in their own headed block (re-read before each edit). Commit only a job's own files.
- Helpers may be refused the impeccable/intent skills; they then work from craft.md, PRODUCT.md and DESIGN.md.
- Generators live in design/canvas/scripts/ (s4s_build.py holds the shared desktop shell; a shell change regenerates S3/S4/S11). Phone screens: s7_build.py pattern. Shots use fallback fonts.
- New components so far: see components.md › Additions and design/system/components/additions.css (§1–25).
- FIN backlog: system-wide 10px SVG track text; S3's local th.num rule now redundant; ip-crumbs links 14px tall (D46 wants ≥24px); S3/S11 separate row-header classes → one DataTable class (S11P.md).
