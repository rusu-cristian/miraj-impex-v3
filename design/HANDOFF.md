# Handoff
Updated: 2026-11-17 20:40 · Last commit: e0b5e58
Repo: rusu-cristian/miraj-impex-v3 · Branch: main · Target: Design canvas "Impex Line Panel · Ecrane" (https://claude.ai/artifact/VC8kqksPgxxYo9kPpK14yv, slug e44dd7fb-2990-41f9-a791-2841f1b1dcd1, live version 1794856515-f4a7) built from the Impex Line Panel design system (https://claude.ai/artifact/KWvk2DHkGgVsDKdhwybfgd) · Setup: full, canvas mode (canvas read-only from this session, see below)

## Progress
Done 54 / 72 · Screens done: S15 + 7 states, S4, S7, S3, S11 (+S11M), S12, S13 (all polished), S6 + 9 states, S4 tabs (polished), S9 (polished), S10 (polished), S2 + 12 states (polished) + Observator (S2V), S14 + 4 states + filter sheet (polished), S8 + 6 states (polished) · Building: none; all SYS jobs done

## Just finished
- S15S Introdus azi states: 6 phones x 1410…3760 (se încarcă, eroare la încărcare, Acțiuni deschis, după ștergere, după 23:59 + modificat la 09:52, manager 10.11) + vet desktop x 4230 1440×1492 (Sidebar Înregistrează expanded); D99; 1 fix round (e0b5e58)
- S15 Introdus azi: 3 phone boards, row y 27292, x 0/470/940 (vet Dr. Ceban 16.11, 22 record rows by type, h 1636; manager Ion Rusu empty; delete Dialog Mortalitate H4.2); abatere Lamp on Ouă H3 + Hrană H1.2 ♂ rows; tab bar marks Înregistrează; D98; S15S queued for leftover states; 1 fix round (320a02f)
- S8P Cântărire polish: §38 (`ip-readout__live`, `ip-seg--fill`, `ip-fieldmsg`), weight-field descriptions read the lamp word once, "Nescrisă la cântărire." hint, author-name rule (D96), copy fixes; heights valoare refuzată 959, zi trecută 1056; D97; 1 fix round (0c9eae1)
- S8S Cântărire states: se încarcă, alege secțiunea (390×844 sheet), zi trecută (Telița 17.02), fermă fără păsări, deja introdus, gol Telița; x 3290…5640 in the S8 row; D96; passed first review (01b6deb)
- S8 Cântărire: 7 phone boards, row y 26059, x 0…2820 step 470 (default Hîrbovăț 16.11 H1.1 ♀ 3 967 g, gol, în normă + abatere as Telița replays, valoare refuzată, salvat, eroare); D95; 2 fix rounds (live comparison announced once with its lamp word) (2596972)
- S14P Producție polish: §36 tablescroll relative, §37 busy buttons keep colours; phone filter BottomSheet board `--telefon-filtre` (x 12630, 390×844); cycle Stats (62,1 g / 90,6% / 2025 62,0 g); orphan words glued at 1280; D94; 2 fix rounds (271f2e7)
- S14S Producție states: observator · se exportă (x 7600), eroare la încărcare (x 9120), fără fermă de ouat (x 10640, D39), telefon 390×2709 (x 12160; next row starts y 26043); D93; 2 fix rounds (phone crumb hit area, busy key) (3f1c662)
- S14 Rapoarte · Producție: 5 desktop boards, row y 23214, x 0…6080 (default 16.11.2026, se încarcă, fără date în perioadă, fără ghid, Seria 2025 closed cycle h 1716); D92; 1 fix round (emptying-only weeks "–", not 0) (5621133). s14_build.py --canvas needs the heights file passed too
- S2V Observator Acasă: phone (x 13820, 390×2172) + desktop (x 14290, 1440×1116) in the S2 row; phone CycleTrack drawn at 358px; D91 (D06 viewer prices read-only stays; spec §1 table says "no prices" — user to correct); passed first review (4f42545)
- S2P Acasă polish: §35 (`.ip-list--alerts` phone alert row, BottomTabBar disabled key, `.ip-scrim--phone` + `.ip-dialog__keys`, `.ip-readouts__skel`, `.ip-trackskel`); all 16 S2 boards rebuilt (vet h 900, manager phone h 1876); D90; D89 queues S2V (Observator Acasă); 1 fix round + shots retaken (9d23a4c)
- S2S Acasă: manager Introdus azi (empty, D87), "2 FERME", 12 state boards (8 phones from x 3980, 4 desktops from x 7740, row y 20922); D88; 1 fix round (menu focus, disabled tab key) (3219a79). Rebuild the row with s2s_build.py --canvas, never s2_build.py --canvas
- S2 Acasă: vet phone (x 0), manager phone (x 470), manager desktop (x 940), owner desktop (x 2460), row y 20922, all 16.11.2026 figures from extracts; D86, D87 (manager also gets Introdus azi → S2S); passed first review (c041aea)
- S10P Vânzare polish: ghost "Toate secțiunile" key (§34, aria-pressed tri-state), partial-sale D39 board (H4 sold, cycle continues) at x 2350, desktop → x 2820; SaveBars 2 lines; D85; passed first review (9ac7251)
- S10 Vânzare: whole Hîrbovăț flock sold 16.11 (D39) + confirmation dialog (cycle closes sapt. 40), refuzat, salvat, eroare, desktop panel; row y 18734; sstate_css gained FX/MODAL/OPT; D84 ("Toate secțiunile" key queued for S10P); 1 fix round (8bd72bb)
- S9P Transfer polish: "alege secțiunea" phone BottomSheet (§33, components.md entry), "ciclu nou" D39 board, error board Banner + preview above SaveBar; S9 row now 7 boards, desktop x 2820; 1 fix round (a3ed7ac)
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
- S15P → S16 → S16P → S17 (all system jobs done; screens only from here)

## Waiting on the user
- Non-blocking: spec §1 roles table says the viewer sees "no prices", D4/D06 say read-only (D91 keeps D06) — correct the table before S11 viewer boards are revisited.
- Non-blocking: D95 guide week — app formula D02 puts 24.02.2026 in SAPT 21 (2 525 g, în normă); the workbook compares with week 20 (abatere). Confirm D02.
- None blocking otherwise. 17.11.2026 11:03: user said "go on with what you recommend, I will review later": take helpers' recommended options, log them as decisions "to review", no gates. 16.11.2026 resume: user chose to keep defaults and continue without screen gates; review everything later (flagged: D24, D25, D27, D53, D59, D63, D64; D78 departs from spec A2).

## Unpublished on the canvas (publish with publish_files, mode patch, "as" = project/<file>)
- Since 16.11.2026 sessions have network access off and did not make the canvas, so no tool here can write to it (get_artifact → network_off_limit). Publish from a chat opened from the canvas page, or from a session with network access on.
- project/additions.css (§23–25 SYS08, §26–27 SYS09, §28 SYS10, §29 SYS11, §30 SYS12, §31 SYS13, §32 S4T, §33 S9P); project/canvas.json
- project/S11-rapoarte-hrana{,--se-incarca,--manager,--observator,--fara-livrari-in-perioada,--modifica-pret,--modifica-pret--eroare}.dc.html (all 7 rebuilt by S11M)
- project/S12-preturi-hrana{,--se-incarca,--in-lucru,--toate-livrarile-au-pret,--eroare}.dc.html (S12 row now y 10303)
- project/S13-livrare-hrana{,--salvat,--cantitate-lipsa,--eroare-la-salvare,--desktop}.dc.html (row y 11447; S13P rebuilt all but --cantitate-lipsa; S11X rebuilt all S11, S12 boards and S13 --desktop)
- project/S6-zilnic.dc.html (row y 12591, x 0, 390×4284) + S6-zilnic--{se-incarca,deja-introdus,zi-trecuta,abatere,refuzat,eroare-la-salvare,ferma-fara-pasari,salvat,motiv-lipsa}.dc.html (x 470…4230)
- project/S4-ferma--{pasari,hrana,greutate,productie}.dc.html (y 800, x 13680/15200/16720/18240) + project/S4-ferma-rezumat{,--*}.dc.html (all 9 rebuilt by S4TP; fără ghid h 1938) + project/canvas.json: every row below S4 moved +595 (redo on the live canvas.json, then add the 4 boards)
- project/S9-transfer{,--salvat,--refuzat,--eroare-la-salvare,--alege-sectiunea,--ciclu-nou,--desktop}.dc.html (row y 17590, phones x 0…2350, desktop x 2820) + canvas.json + additions.css (§33)
- project/S10-vanzare{,--confirmare,--refuzat,--salvat,--eroare-la-salvare,--desktop}.dc.html (row y 18734, x 0…2350) + canvas.json
- S10P: project/additions.css (§34) + project/S10-vanzare{,--confirmare,--refuzat,--salvat,--eroare-la-salvare,--partiala,--desktop}.dc.html (desktop now x 2820) + canvas.json
- S2: project/S2-acasa{,--manager,--manager-desktop,--proprietar-desktop}.dc.html (row y 20922) + canvas.json + S2S: S2-acasa--{se-incarca,inainte-de-12,zi-trecuta-blocat,sterge-confirmare,dupa-2359,ferma-goala,fara-acces,eroare-la-incarcare,manager-desktop--se-incarca,manager-desktop--zi-trecuta-stoc-sub-0,proprietar-desktop--se-incarca,proprietar-desktop--toate-cu-pret}.dc.html (manager + owner defaults rebuilt)
- S2P: project/additions.css (§35) + all 16 S2-acasa*.dc.html + canvas.json (heights)
- S2V: project/S2-acasa--observator{,-desktop}.dc.html + canvas.json
- S14: project/S14-rapoarte-productie{,--se-incarca,--fara-date-in-perioada,--fara-ghid,--seria-2025}.dc.html + canvas.json
- S14S: project/S14-rapoarte-productie--{observator-se-exporta,eroare-la-incarcare,fara-ferma-de-ouat,telefon}.dc.html + canvas.json
- S14P: project/additions.css (§36–37) + all 10 S14-rapoarte-productie*.dc.html (incl. new --telefon-filtre) + canvas.json
- S8: project/S8-cantarire{,--gol,--in-norma,--abatere,--valoare-refuzata,--salvat,--eroare-la-salvare}.dc.html + canvas.json
- S8S: project/S8-cantarire--{se-incarca,alege-sectiunea,zi-trecuta,ferma-fara-pasari,deja-introdus,gol-telita}.dc.html + canvas.json
- S8P: project/additions.css (§38) + all 13 S8-cantarire*.dc.html + canvas.json (heights)
- S15: project/S15-introdus-azi{,--gol,--sterge-confirmare}.dc.html + canvas.json
- S15S: project/S15-introdus-azi--{se-incarca,eroare-la-incarcare,actiuni-deschis,dupa-stergere,dupa-2359,manager-10-11,desktop}.dc.html + canvas.json
- Add every later board and each additions.css / canvas.json change here.

## Things to remember
- 17.11.2026 session pattern: every screen gets a ..S states job when the builder lists leftover states (S2S, S14S, S8S added); polish jobs move screen CSS that is really a system state into a new additions.css § (now §38). Reviewers recompute figures and measure a11y in Chrome AX tree; shots must be taken at board heights, not the 844 heights viewport.
- One screen at a time; SYS jobs can run in parallel and append to additions.css / components.md › Additions in their own headed block (re-read before each edit). Commit only a job's own files.
- Helpers may be refused the impeccable/intent skills; they then work from craft.md, PRODUCT.md and DESIGN.md.
- Generators live in design/canvas/scripts/ (s4s_build.py shared desktop shell; s7_build.py phone pattern; s11_build.py price/error classes reused by S12; s13_build.py imports s7s/s7/s11/s4s and rewrites S7 boards byte-identical). Shots use fallback fonts.
- New components: components.md › Additions and design/system/components/additions.css (§1–33).
- FIN backlog: 10px SVG track text; S3's local th.num rule redundant; ip-crumbs links 14px tall (D46 wants ≥24px); S3/S11 row-header classes → one DataTable class; FeedTotals preview summary text (D62); SectionTree aria-controls → [role=menu] like SYS09, menu + disabled Retrage share one aria-describedby note (read twice); S13 panel Fermă select offers Telița (no open cycle), no board for that case; S11 tables → system `.ip-tablescroll` (sticky first column; tall Consum shows its scrollbar only at the bottom); SaveBar summary class `ip-ftot-sum` (FeedTotals name, used by S6/S7) → neutral SaveBar-summary class; SYS11 nits a–c in notes/SYS11.md; Toast-with-undo and Banner-with-action layouts (sstate_css.py) → additions.css; SYS12 nits in notes/SYS12.md; S3/S11/ip-cost/ip-sparktable row-header copies → §32 .ip-table--rowheads; S4 ciclu ales wraps 2 Stats + sub-line at 1280; Producție Pe hale scrolls ~6px at 1280; Rezumat lists are div links (tabs use ul/li); S4 needs a narrow layout beside a SidePanel (CycleTrack blocks 21px at 1280 on S9 desktop; S9 board rule for now); no board shows Anulează after a save; S6P FIN: "din ♀ n" nowrap into HallCard .ip-eggs__sub (drop .s6-nobr), EntryCardPlus/HallCard previews adopt section-first aria-labelledby naming; §34 key box 1px right of the rows' Toate box (subtract --ip-hair in the negative margin); S10 "mixed" key state out of view on the 844 refuzat/parțială boards; SYS08 EntryKeys disabled links focusable like SYS09 (tabindex=0); S2 "nimic lipsă" before-12 variant + De verificat empty not drawn; system rule "meta breaks at ·" (S2 markup only; S4 s4s.alerts() wraps the same); s4s Dialog/skeleton/empty-box + sstate_css Banner/Toast-undo (s7s-*) classes still screen CSS; S8 abatere warn line reads "♂ :" (space before colon); phone FilterBar key + filter sheet body and phone Readout without GuideChart → system variants; S4 Seria 2025 "Masă medie ou · SAPT 47" Stat repeats its table row; CycleTrack "phone" variant (S2V track_svg phone) → components.md Additions entry, axis start lacks a month label, two transfer dots merge at 358px.
