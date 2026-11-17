# Handoff
Updated: 2026-11-17 18:10 · Last commit: bb7eddf
Repo: rusu-cristian/miraj-impex-v3 · Branch: main · Target: Design canvas "Impex Line Panel · Ecrane" (https://claude.ai/artifact/VC8kqksPgxxYo9kPpK14yv, slug e44dd7fb-2990-41f9-a791-2841f1b1dcd1, live version 1794856515-f4a7) built from the Impex Line Panel design system (https://claude.ai/artifact/KWvk2DHkGgVsDKdhwybfgd) · Setup: full, canvas mode (canvas read-only from this session, see below)

## Progress
Done 39 / 67 · Screens done: S4, S7, S3, S11 (+S11M), S12, S13 (all polished), S6 + 9 states, S4 tabs (polished), S9 (polish next) · Building: S9P; all SYS jobs done

## Just finished
- S9 Transfer: phone replay 28.02.2026 Telița H2 → Hîrbovăț H4.2 + salvat, refuzat (D39), eroare; desktop SidePanel beside S4; row y 17590; side-panel/rec/effect CSS now shared in sstate_css.py; D83; 2 fix rounds (bb7eddf)
- S4TP S4 tabs polish: Rezumat 15.11 deaths only (cull on its own row), .5 rounds up, SYS06 legends on Rezumat/fără ghid/ciclu ales, 1280 Stats on one line; 13 S4 boards rebuilt (fără ghid h 1938); 2 fix rounds (46027dd)
- S4T S4 tabs Păsări (2529), Hrană, Greutate, Producție at y 800 x 13680…18240; §32 opt-in DataTable row headers + framed scroll; D81, D82; 2 fix rounds (fca4947)
- S6M 9th Zilnic board: H1.1 ♀ Sacrificare 2 without Motiv (refuzat, D39 example) + Mai mult open, x 4230 (18aa6a8)
- SYS13 LoginCard + text Wordmark, §31; D80 (show/hide key, no reset link, proposed hint/empty-field copy); 1 fix round (01ae64f)
- S6P Zilnic polish: section-first accessible names, "±3" Mono, "din ♀ n" kept together, se încarcă skeleton = loaded geometry; all 9 S6 boards rebuilt; D79 queues S6M (7a663c3)
- SYS12 UserRow list + OneTimePassword, §30; D78 (app generates the password, shown once; departs from spec A2 field list, flagged); 2 fix rounds (password field a11y, announced failures) (618b6cc)
- S6S Zilnic 8 state boards, row y 12591 x 470…3760 (D77: one Editează for the form); shared state CSS sstate_css.py with S7S; 1 fix round (615c3a6)
- SYS11 GuideImportPreview, 9 states, §29; D76 (refused rows block saving; one preview row per file row); passed first review (94c1d59)
- S6 Zilnic default board: Hîrbovăț 16.11.2026, Dr. Ceban's round typed before save, H1/H3/H4 × 2 EntryCards + Ouă groups, row y 12591 x 0 (964a0e6)
- S11X S11 Livrări/Consum tables scroll in a shared `.s11-tscroll` box (S13 reuses it); 1 fix round (c7e65c9)
- S13P Livrare hrană polish: S11 tables scroll in focusable caption-named boxes beside the panel (1280 clipped two columns); salvat stock line 3 lines (222b8e9)
- SYS10 SectionTree editor, 16 states, §28; D75 (tree edits save one at a time); 1 fix round (dc5ed3d)
- S13 merged from design-loop/s13-session-0139 (D73, D74)

## Next up
- S9P → S10 → S10P → S2 → S2P (all system jobs done; screens only from here)

## Waiting on the user
- None blocking. 17.11.2026 11:03: user said "go on with what you recommend, I will review later": take helpers' recommended options, log them as decisions "to review", no gates. 16.11.2026 resume: user chose to keep defaults and continue without screen gates; review everything later (flagged: D24, D25, D27, D53, D59, D63, D64; D78 departs from spec A2).

## Unpublished on the canvas (publish with publish_files, mode patch, "as" = project/<file>)
- Since 16.11.2026 sessions have network access off and did not make the canvas, so no tool here can write to it (get_artifact → network_off_limit). Publish from a chat opened from the canvas page, or from a session with network access on.
- project/additions.css (§23–25 SYS08, §26–27 SYS09, §28 SYS10, §29 SYS11, §30 SYS12, §31 SYS13, §32 S4T); project/canvas.json
- project/S11-rapoarte-hrana{,--se-incarca,--manager,--observator,--fara-livrari-in-perioada,--modifica-pret,--modifica-pret--eroare}.dc.html (all 7 rebuilt by S11M)
- project/S12-preturi-hrana{,--se-incarca,--in-lucru,--toate-livrarile-au-pret,--eroare}.dc.html (S12 row now y 10303)
- project/S13-livrare-hrana{,--salvat,--cantitate-lipsa,--eroare-la-salvare,--desktop}.dc.html (row y 11447; S13P rebuilt all but --cantitate-lipsa; S11X rebuilt all S11, S12 boards and S13 --desktop)
- project/S6-zilnic.dc.html (row y 12591, x 0, 390×4284) + S6-zilnic--{se-incarca,deja-introdus,zi-trecuta,abatere,refuzat,eroare-la-salvare,ferma-fara-pasari,salvat,motiv-lipsa}.dc.html (x 470…4230)
- project/S4-ferma--{pasari,hrana,greutate,productie}.dc.html (y 800, x 13680/15200/16720/18240) + project/S4-ferma-rezumat{,--*}.dc.html (all 9 rebuilt by S4TP; fără ghid h 1938) + project/canvas.json: every row below S4 moved +595 (redo on the live canvas.json, then add the 4 boards)
- project/S9-transfer{,--salvat,--refuzat,--eroare-la-salvare,--desktop}.dc.html (row y 17590, x 0…1880) + canvas.json
- Add every later board and each additions.css / canvas.json change here.

## Things to remember
- One screen at a time; SYS jobs can run in parallel and append to additions.css / components.md › Additions in their own headed block (re-read before each edit). Commit only a job's own files.
- Helpers may be refused the impeccable/intent skills; they then work from craft.md, PRODUCT.md and DESIGN.md.
- Generators live in design/canvas/scripts/ (s4s_build.py shared desktop shell; s7_build.py phone pattern; s11_build.py price/error classes reused by S12; s13_build.py imports s7s/s7/s11/s4s and rewrites S7 boards byte-identical). Shots use fallback fonts.
- New components: components.md › Additions and design/system/components/additions.css (§1–32).
- FIN backlog: 10px SVG track text; S3's local th.num rule redundant; ip-crumbs links 14px tall (D46 wants ≥24px); S3/S11 row-header classes → one DataTable class; FeedTotals preview summary text (D62); SectionTree aria-controls → [role=menu] like SYS09, menu + disabled Retrage share one aria-describedby note (read twice); S13 panel Fermă select offers Telița (no open cycle), no board for that case; S11 tables → system `.ip-tablescroll` (sticky first column; tall Consum shows its scrollbar only at the bottom); SaveBar summary class `ip-ftot-sum` (FeedTotals name, used by S6/S7) → neutral SaveBar-summary class; SYS11 nits a–c in notes/SYS11.md; Toast-with-undo and Banner-with-action layouts (sstate_css.py) → additions.css; SYS12 nits in notes/SYS12.md; S3/S11/ip-cost/ip-sparktable row-header copies → §32 .ip-table--rowheads; S4 ciclu ales wraps 2 Stats + sub-line at 1280; Producție Pe hale scrolls ~6px at 1280; Rezumat lists are div links (tabs use ul/li); S6P FIN: "din ♀ n" nowrap into HallCard .ip-eggs__sub (drop .s6-nobr), EntryCardPlus/HallCard previews adopt section-first aria-labelledby naming.
