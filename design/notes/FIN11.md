# FIN11 · polish · Skeleton consistency · 18.11.2026 · not published (D131)
- additions.css (both copies): §42 gains `ip-skel--ch1…4` (ch of the inherited font) and `--key` (control-h); §29 `ip-gimp__sk1…4` rules removed; §35 comment lists its users.
- SYS11 `skel_body` → `--ch{n}` (preview rebuilt, pixels same); SYS12 `SK` Acțiuni bar `--w3` → `--key --end` (S22 icon column 80 → 40, as loaded; preview rebuilt). READMEs: UserRow, GuideImportPreview.
- s11_build: Stoc skeleton → §35 `ip-readouts__skel` (5 bars), `.s11-skro` CSS removed (S11–S13 styles). s18s_build: `ip-trackskel` without the dead `s18s-trackskel`; rows → §42 `tr.ip-table__skel`, `.s18s-skt` copy removed.
- Not changed: `s16s-sk*` (field footprints, no figure width; D131, kept with s7s-skcol for one later field skeleton).
- Rebuilt the whole canvas in README order (/tmp/fin11/rebuild.sh). 23 boards changed (S11 ×7, S12 ×5, S13 ×1, S18 ×8, S20 ×1, S22 ×1); normalising diff (styles stripped, old→new map) leaves nothing else; styles lose only the removed rules.
- Pixels (old vs new at board size; S18/S20 also 1280): only S11 se-incarca (light §35 boxes, page 34px shorter) and S22 se-incarca (column widths) change.
- Height: S11 se-incarca 1241 → 1207 (s11_heights.json + canvas.json h only; x/y unchanged).
- Shots retaken (3): S11-1 (s11_render, heights to /tmp), S22-12, S22-12-1280 (s22_render). Second full run byte-identical (boards, previews, heights).
- Open (D131): §35 boxes shorter than loaded readouts/tracks; content-sized loading columns (S11, S18–S22) differ from loaded widths.
