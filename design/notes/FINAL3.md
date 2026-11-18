# FINAL3 closing check (19.11.2026, after FY01–FY04)
Clean: Dialogs (14), S18 ranges, S2 touch targets, no overflow at board size/1280, scroll h = canvas h (S11 +1px fallback fonts), row gaps 120 (Main→S4 240), docstring sums, FIN09/D143, no new CSS duplicates.
Findings → FZ01: S22 resetare-esuata Banner lead not alone (D147); 7 stale shots (Sidebar chevron sub-pixel).
- FZ01 · 1 (D147): S22 resetare eșuată + OneTimePassword preview: lead alone "Parola lui Ion Rusu nu s-a resetat." ¶ "Verifică conexiunea la internet, apoi apasă Reîncearcă. Parola veche merge în continuare." (sys12_build, s22s_build; UserRow/OneTimePassword README, components.md quote); canvas.json unchanged, s22s rerun byte-identical; S22-20 (+1280) retaken.
- FZ01 · 2: S17-6/7/8, S18-7 (each +1280), S14-7-1280 retaken by scratch render (headless Chrome, viewport = canvas.json board, 1280 for -1280, fonts.ready); only diff was the Sidebar chevron (28 px); no heights files touched.
