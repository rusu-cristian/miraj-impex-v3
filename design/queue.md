# Queue

Target: Design canvas "Impex Line Panel · Ecrane" (slug e44dd7fb-2990-41f9-a791-2841f1b1dcd1), artboards built from the Impex Line Panel design system (repo copy in design/system/, additions in design/system/components/ + components.md "Additions"). Recipe: design/kit/canvas.md.

**Always in context (every build/review job, not repeated below):** design/kit/craft.md · design/kit/decisions-current.md · design/kit/canvas.md · design/system/INDEX.md · design/data/extracts/README.md (screens only).

Screen jobs build the default artboard; `..S` jobs add the spec's state artboards; `..P` jobs are /impeccable polish + audit (with the /unslop template test). S5 (section level) is not in the requested list and is not queued.

| ID | Type | Name | Depends on | Context | Status | Commit |
|----|------|------|------------|---------|--------|--------|
| DATA01 | data | Extracts: structure, placement, transfers, Telița mortality/feed, deliveries, feed types, weighings, guides, cycles | — | design/kit/setup-check.md §1 §4; design/data/Seria_2026.xlsx; data-model.md | done | b71679a |
| DATA02 | data | Extracts: Hîrbovăț previous-flock eggs, counts-latest, illustrative sample-hirbovat-2026 (01.03–16.11.2026) | DATA01 | design/kit/setup-check.md §1 §4 §5; data-model.md#calculated-not-stored; spec.md#1-shared-rules; design/data/extracts/guide-laying.json | done | 3b9cb85 |
| SYS00 | system | Repo system integration: additions.css on canvas, components.md "Additions" section, rule clarifications (Stat value = `readout`; uppercase exceptions; green nav/time cues; AllocationRow lamp; "fără date în ghid" is a guide line, not a Tag) | — | components.md#stat, #tag, #allocationrow, #lamp; DESIGN.md#typography; design/system/components/bundle.css | done | c60d4d3 |
| SYS01 | system | For S4: CycleTrack `farm` variant (one lane, cycles of any series); SectionTile totals strip; AlertRow (De verificat); Stat + GuideChart "fără date în ghid"/"fără ghid" states; Sidebar series selector; FilterBar "Interval" range | SYS00 | components.md#cycletrack, #sectiontile, #stat, #guidechart, #sidebar, #filterbar, #eventrow, #lamp; screens.md#s4; spec.md#d1 | done | 3ef8e34 |
| S4 | screen | S4 Fermă · Rezumat (desktop), Hîrbovăț ciclu Seria 2026 | SYS01, DATA02 | screens.md#s4; spec.md#d1, #b4; components.md#pageheader, #cycleselector, #cycletrack, #filterbar, #recordchip, #tabs, #readout, #guidechart, #stat, #sectiontile, #eventrow, #sidebar, #lamp, #tag, #sexcount; extracts: farms-sections, cycles, counts-latest, sample-hirbovat-2026, guide-laying, eggs-hirbovat-prev-weekly | done | 87cac50 |
| S4S | screen | S4 states: se încarcă, fără date în perioadă, fără ghid, fără acces, fermă goală între cicluri, CycleSelector open, Ciclu ▾ dialog Începe ciclu nou | S4 | spec.md#d1 (States), #b4; components.md#skeleton, #emptystate, #dialog, #cycleselector, #banner | done | 40702f8 |
| S4P | polish | S4 + S4S polish + audit (incl. % format per D46: "0 %" axis label → "0%"; move `.s4-chipsplit` and `.s4-ded` into additions.css as system classes, D47/D55) | S4S | screens.md#s4; spec.md#d1; DESIGN.md#dos-and-donts | done | d8edc16 |
| SYS02 | system | For S7: FeedTotals block (per feed type: kg today, stock after, days left, lamp) | SYS00 | components.md#allocationrow, #readout, #lamp, #savebar; screens.md#s7; spec.md#c2 | done | 52fa6f6 |
| S7 | screen | S7 Hrană zilnică (phone 390); fix FeedTotals compact SaveBar lamp word "Cocoși 3,9 zile" → "abatere" (D42) | SYS02, S4P | screens.md#s7; spec.md#c2; components.md#formheader, #datefield, #allocationrow, #select, #textfield, #button, #savebar, #toast, #banner; extracts: farms-sections, counts-latest, sample-hirbovat-2026, guide-laying, feed-types | done | 63abd62 |
| S7S | screen | S7 states: fără zi anterioară, avertisment abatere, stoc sub 0, salvat (toast), eroare cu reîncercare | S7 | spec.md#c2 (Rules, States), #1-shared-rules (Connectivity) | done | d865d10 |
| S7P | polish | S7 polish + audit (move `s7-*` AllocationRow phone layout into additions.css as `ip-alloc--phone`, D57) | S7S | screens.md#s7; spec.md#c2; DESIGN.md#dos-and-donts | done | 0b05160 |
| SYS03 | system | For S3: FarmSummary panel | SYS00 | components.md#stat, #sexcount, #lamp, #tag; screens.md#s3; spec.md#d1 | done | 36313aa |
| S3 | screen | S3 Seria 2026 · nivel serie, cycle track full (desktop) + states | SYS03, S7P | screens.md#s3; spec.md#d1, #b3, #b4; components.md#cycletrack, #pageheader, #tabs, #readout; extracts: cycles, placement-seria2026, transfers-seria2026, counts-latest, sample-hirbovat-2026 | done | 33b2847 |
| S3P | polish | S3 polish + audit | S3 | screens.md#s3; spec.md#d1; DESIGN.md#dos-and-donts | done | 45aa9be |
| SYS04 | system | For S11: Readout stock variant (row of stock readouts), Cost block | SYS00 | components.md#readout, #lamp, #datatable; screens.md#s11; spec.md#d4 | done | 1bb0a25 |
| S11 | screen | S11 Rapoarte · Hrană (Stoc · Livrări · Consum · Cost), owner view + manager/viewer read-only note | SYS04, S3P | screens.md#s11; spec.md#d4, #c5, #c6; components.md#datatable, #tag, #textfield, #pageheader, #tabs, #filterbar; extracts: feed-deliveries, feed-daily-telita, feed-types, sample-hirbovat-2026 | done | 6c15ca0 |
| S11P | polish | S11 polish + audit | S11 | screens.md#s11; spec.md#d4; DESIGN.md#dos-and-donts | done | a884384 |
| S12 | screen | S12 Prețuri hrană (owner) + "Toate livrările au preț" + error | S11P | screens.md#s12; spec.md#c6; components.md#datatable, #textfield, #button, #emptystate, #toast; extracts: feed-deliveries | done | 7b723f8 |
| S12P | polish | S12 polish + audit | S12 | screens.md#s12; spec.md#c6 | done | b2f47e9 |
| S11M | screen | S11 Livrări: owner changes a saved price ("Modifică" inline row + required Motiv + "rămâne în jurnal", D69) + validation error; align Lei/kg & Total lei header text with the input edge (S12P) | S12P | screens.md#s11; spec.md#c6, #d4; components.md#datatable, #textfield, #button, #toast; design/notes/S11P.md, S12P.md; extracts: feed-deliveries, sample-hirbovat-2026 | done | b58c66d |
| S13 | screen | S13 Livrare hrană (manager; phone + desktop side panel) + saved toast | S11M | screens.md#s13; spec.md#c5; components.md#select, #datefield, #textfield, #button, #sidepanel, #toast; extracts: feed-deliveries, feed-types, sample-hirbovat-2026 | done | c6bff04 |
| S13P | polish | S13 polish + audit | S13 | screens.md#s13; spec.md#c5 | done | 222b8e9 |
| S11X | screen | S11 boards: Consum table (1113px) and Livrări table scroll inside their own focusable, caption-named box where the content area is narrower (as S13P `.s13-tscroll`); regenerate S11 + S13 boards | S13P | design/notes/S11M.md, S13P.md; design/canvas/scripts/s11_build.py, s13_build.py; screens.md#s11 | done | c7e65c9 |
| SYS05 | system | For S6: EntryCard yesterday-mortality hint + hall card with Ouă group | SYS00 | components.md#entrycard, #countinput, #checkbox; screens.md#s6; spec.md#c1 | done | 40702f8 |
| S6 | screen | S6 Zilnic (phone 390), Hîrbovăț | SYS05, S13P | screens.md#s6; spec.md#c1; components.md#formheader, #datefield, #entrycard, #countinput, #select, #checkbox, #savebar, #toast; extracts: farms-sections, counts-latest, sample-hirbovat-2026 | done | 964a0e6 |
| S6S | screen | S6 states: se încarcă, deja introdus (Editează), zi trecută pentru vet, avertisment 3×, refuz peste vii, eroare cu reîncercare, fermă fără păsări, salvat | S6 | spec.md#c1 (Rules, States) | done | 615c3a6 |
| S6P | polish | S6 polish + audit | S6S | screens.md#s6; spec.md#c1; reviewer nits from S6: "(prag ±3)" figure in Mono; H3 "din" / "♀ 8 475" line break; field accessible names lack the section (system-level, may go to FIN) | done | 7a663c3 |
| S6M | screen | S6 9th state board: Mai mult open (Eroare sexare) + editable Motiv Select with "Motiv lipsă" refusal, D39 example H1.1 ♀ Sacrificare 2 (D79) | S6P | screens.md#s6; spec.md#c1; components.md Additions › SYS05 (EntryCardPlus); design/notes/S6S.md, S6P.md; design/canvas/scripts/s6_build.py, s6s_build.py, sstate_css.py | done | 18aa6a8 |
| SYS06 | system | For S4 tabs: align GuideChart/GuideMissing previews to role="group" (D44); GuideChart variants (mortality %, kg/day, CV/uniformity, hatching, egg weight, small curve); RecordCode adjustment + cycle open/close marks | SYS00 | components.md#guidechart, #recordcode, #recordtile; spec.md#d1 (Tabs) | done | 63abd62 |
| S4T | screen | S4 tabs: Păsări (type-filtered event log), Hrană, Greutate, Producție | SYS06, S6M | screens.md#s4; spec.md#d1 (Tabs, Filters); components.md#eventrow, #recordchip, #filterbar, #datatable, #guidechart, #readout; extracts: sample-hirbovat-2026, guide-laying, weighings-weekly, eggs-hirbovat-prev-weekly | done | fca4947 |
| S4TP | polish | S4 tabs polish + audit | S4T | spec.md#d1; DESIGN.md#dos-and-donts; also align S4 Rezumat: 15.11 "Mortalitate ♂ −4" includes the H3.1 cull (split as Păsări: ♂ −2 + Sacrificare ♂ −2), round .5 up (D82), Readout legend → SYS06 colour key | done | 46027dd |
| S9 | screen | S9 Transfer (phone + desktop) incl. cycle open/close preview, refusal | S4TP | screens.md#s9; spec.md#c3; components.md#sectionselect, #countinput, #checkbox, #effectpreview, #banner, #button, #toast; extracts: farms-sections, counts-latest, transfers-seria2026 | done | bb7eddf |
| S9P | polish | S9 polish + audit | S9 | screens.md#s9; spec.md#c3; from S9 (D83): add a D39 phone board "ciclu nou" (move onto an empty farm, "Ciclu nou la …, sapt. 1"); draw the phone SectionSelect BottomSheet open; error board: preview no longer under the SaveBar; S4 page narrow layout beside a panel (s4s_build, keep other boards identical) or FIN | done | a3ed7ac |
| SYS07 | system | SexCountRow: section ♀·♂ count row with "Toate", add/remove (S10, S18) | SYS00 | components.md#countinput, #sexcount, #checkbox, #iconbutton; screens.md#s10, #s18; spec.md#c4, #b3 | done | d865d10 |
| S10 | screen | S10 Vânzare + confirmation dialog (farm empties) | SYS07, S9P | screens.md#s10; spec.md#c4; components.md#effectpreview, #dialog, #textfield, #datefield, #button; extracts: counts-latest, sample-hirbovat-2026 | done | 8bd72bb |
| S10P | polish | S10 polish + audit | S10 | screens.md#s10; spec.md#c4; from S10 (D84): ghost "Toate secțiunile" key in the SexCountRow head on sales (new additions.css § + components.md entry), redraw default/confirmation boards with it; add a partial-sale D39 board (cycle continues); refuzat SaveBar line 2 wraps to 3 lines | building | |
| SYS08 | system | Acasă blocks: MissingRow (Lipsesc date), EntryKeys, UnpricedCount card | SYS00, SYS01 | components.md#eventrow, #recordtile, #button, #lamp, #bottomsheet; screens.md#s2; spec.md#d2, #c9 | done | d1bd1e3 |
| S2 | screen | S2 Acasă: vet (phone), manager (phone + desktop), owner (desktop) | SYS08, S10P | screens.md#s2; spec.md#d2, #c8, #c9; components.md#cycletrack, #bottomtabbar, #sidebar, #readout; extracts: all needed | todo | |
| S2P | polish | S2 polish + audit | S2 | screens.md#s2; spec.md#d2 | todo | |
| S14 | screen | S14 Rapoarte · Producție | S2P | screens.md#s14; spec.md#d3; components.md#datatable, #filterbar, #pageheader; extracts: eggs-hirbovat-prev-weekly, sample-hirbovat-2026, guide-laying | todo | |
| S14P | polish | S14 polish + audit | S14 | screens.md#s14; spec.md#d3 | todo | |
| S8 | screen | S8 Cântărire (phone) + live comparison states | S14P | screens.md#s8; spec.md#c7; components.md#select, #segmentedcontrol, #textfield, #formheader, #savebar; extracts: weighings-weekly, guide-growth | todo | |
| S8P | polish | S8 polish + audit | S8 | screens.md#s8; spec.md#c7 | todo | |
| SYS09 | system | EventRow actions (Editează / Șterge / Corectează, "modificat de"), ChangeRow (old → new) | SYS06 | components.md#eventrow, #iconbutton, #tag; spec.md#c8, #e1, #e2 | done | 4c71c06 |
| S15 | screen | S15 Introdus azi (phone) + empty | SYS09, S8P | screens.md#s15; spec.md#c8; components.md#eventrow, #emptystate, #dialog | todo | |
| S15P | polish | S15 polish + audit | S15 | screens.md#s15; spec.md#c8 | todo | |
| S16 | screen | S16 Corectează (desktop side panel + phone sheet) + refused + delete dialog | S15P | screens.md#s16; spec.md#e1; components.md#sidepanel, #effectpreview, #banner, #dialog, #textfield, #countinput | todo | |
| S16P | polish | S16 polish + audit | S16 | screens.md#s16; spec.md#e1 | todo | |
| S17 | screen | S17 Istoric modificări | S16P | screens.md#s17; spec.md#e2; components.md#filterbar, #datatable, #recordchip, #select | todo | |
| S17P | polish | S17 polish + audit | S17 | screens.md#s17; spec.md#e2 | todo | |
| S18 | screen | S18 Serii + Serie nouă (incl. "Telița are încă…" refusal) | SYS07, S17P | screens.md#s18; spec.md#b3; components.md#datatable, #cycletrack, #textfield, #select, #datefield; extracts: placement-seria2026, cycles | todo | |
| S18P | polish | S18 polish + audit | S18 | screens.md#s18; spec.md#b3 | todo | |
| SYS10 | system | SectionTree editor (S19) | SYS00 | components.md#datatable, #iconbutton, #tag; screens.md#s19; spec.md#b1 | done | dc5ed3d |
| S19 | screen | S19 Ferme și secțiuni + states | SYS10, S18P | screens.md#s19; spec.md#b1; components.md#sidepanel, #select, #segmentedcontrol, #banner, #emptystate; extracts: farms-sections | todo | |
| S19P | polish | S19 polish + audit | S19 | screens.md#s19; spec.md#b1 | todo | |
| SYS11 | system | GuideImportPreview (DataTable with row errors) | SYS06 | components.md#datatable, #filedrop, #banner; screens.md#s20; spec.md#b2 | done | 94c1d59 |
| S20 | screen | S20 Ghiduri: list, import panel, guide detail + states | SYS11, S19P | screens.md#s20; spec.md#b2; components.md#guidechart, #datatable, #filedrop; extracts: guide-growth, guide-laying | todo | |
| S20P | polish | S20 polish + audit | S20 | screens.md#s20; spec.md#b2 | todo | |
| S21 | screen | S21 Tipuri de hrană | S20P | screens.md#s21; data-model.md#tables; components.md#datatable, #sidepanel, #segmentedcontrol, #checkbox; extracts: feed-types | todo | |
| S21P | polish | S21 polish + audit | S21 | screens.md#s21 | todo | |
| SYS12 | system | UserRow list + one-time password display | SYS00 | components.md#datatable, #tag, #lamp, #textfield; spec.md#a2 | done | 618b6cc |
| S22 | screen | S22 Utilizatori + form + deactivate dialog + states | SYS12, S21P | screens.md#s22; spec.md#a2; components.md#sidepanel, #select, #checkbox, #dialog, #banner, #emptystate | todo | |
| S22P | polish | S22 polish + audit | S22 | screens.md#s22; spec.md#a2 | todo | |
| SYS13 | system | LoginCard + Wordmark | SYS00 | components.md#sidebar, #textfield, #button, #banner; spec.md#a1 | done | 01ae64f |
| S1 | screen | S1 Autentificare (phone + desktop) + 4 error states | SYS13, S22P | screens.md#s1; spec.md#a1 | todo | |
| S1P | polish | S1 polish + audit | S1 | screens.md#s1; spec.md#a1 | todo | |
| FIN | polish | Cross-screen consistency review | S1P | all artboards; design/system/; craft.md | todo | |
