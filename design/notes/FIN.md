# FIN review (18.11.2026)

Cross-screen review of all 259 boards (headless Chrome 1440 + 1280, plus grep). Findings became jobs FIN01–FIN16 in queue.md.

## Generator chain (rebuild cascades)
- s12 imports s11; s13 imports s11 + s7s and rewrites the S7 boards; s14 imports s11 + s4t; s17 imports s14s; s18 imports s17 + s3; s19 imports s18; s20, s21, s22 import s19.
- sstate_css.py feeds s2s, s6s, s7s, s8, s9, s10, s13, s14s, s15, s15s, s16, s18.
- After a shared-helper edit, re-run the downstream rows and diff the boards. Rebuild S3, S10, S11, S14 only with their own latest scripts (see HANDOFF). FIN02–FIN05 share generators: run them one after another.

## Already consistent (no job)
Every Dialog is role="alertdialog" with focus on Anulează; every Lamp has a word; phone Toasts sit in one place; crumb markup and SidePanel heads match; screen CSS has no raw colours or px (except board viewport sizes); ♀/♂ carry their colour except the two in FIN08.

## Closed backlog items
"No board shows Anulează after a save" (all but S12, moved to FIN02); S8 "♂ :" → "♂:"; S9 desktop CycleTrack blocks fine at 1280; SYS11 nits b, c fixed in S20P (a unchecked); SYS12 load-failed reason + --up flip done; S2 "nimic lipsă" before-12 drawn (De verificat empty still open, FIN16); LoginCard README has the §46 layout.

## Unclear, carried in jobs
SectionTree aria-controls (FIN13); EntryCardPlus/HallCard section-first naming; S16 boards not drawn: Dialog prefill, cycle-change Dialog, focus return.

## Decisions for the user to review
D110–D121 (Claude, recommended); defaults D15–D109 "to review"; flagged D24, D25, D27, D53, D59, D63, D64/D91 (viewer prices vs spec §1), D78/D118 (password departs from spec A2), D95/D02 (guide week formula).

## Generator risk (found in FIN02, fixed by FIN02G)
- s8_build / s8s_build rewrite canvas.json on every run (even without --canvas, exit 0) and move the S8 row to y 41633; s10 --canvas does the same to S10.
- s15/s15s/s16/s16s stop on a row assert even without --canvas (pin S15 y 27292, S16 y 29020).
- s18–s22 base builds fail their row assert with --canvas: run them without it, then the ..s build with --canvas.
- s8/s9/s10/s12/s13 need their heights file as the 2nd argument (else s10 rewrites S10-vanzare's height).
