# FINAL cross-screen review (18.11.2026, after FIN16)
257 boards at board size + 158 desktops at 1280, fallback fonts, scripted greps. Findings → FX01–FX07.
Already consistent: FIN09 formats, SAPT/sapt. (D128/D134), combobox, current row, Dialog/Toast markup, type weights (except S3 `<b>–</b>`), no overflow at 390/1440/1280 except S10-vanzare--desktop.
1 S10 desktop scrolls 1281 in a 1024 board: `ip-sr` text in the docked panel body positions against `.ip-root` → FX01.
2 S2/S15 desktop block heads 11px caption vs 16px ip-sechead elsewhere; S15 h1 2px; S20 salvat focus moves the head → FX02.
3 Phone form gaps 24 / 20 / 16 → FX03.
4–5 Failure Banner placement, lead and retry copy; Toast lead order → FX04 (D141, D143).
6 Screen CSS repeating system rules (EventRow nowrap, ip-skel block, SaveBar scroll-padding) → FX05.
7 S16 focus below fold at 1280, S4 unit repeated, S3 bold dash, missing tabindex on 7 regions → FX06. Navigation parent (D142) → FX07.
