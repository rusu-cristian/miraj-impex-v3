# AlertRow · "De verificat" (project addition, SYS01)

An EventRow variant for alerts on the Rezumat tab (S4 d.).

- **Markup:** `.ip-list` > `div.ip-ev.ip-ev--alert` > RecordTile (`aria-hidden`) · `.ip-ev__main` (title + `.ip-ev__meta`) · optional `a.ip-link.ip-ev__act` · `.ip-ev__status` (Lamp + word).
- **RecordTile:** the record type the alert is about (eggs for lay, skull for mortality, wheat for missing feed, truck for a refused sale, package for an unpriced delivery).
- **Title:** section · indicator figure · guide or limit, figures in Plex Mono 500 (♀/♂ in their colours): "H3 · ouat 70,1 % · ghid 74,1 %". The indicator word names the type, since the tile is decorative.
- **Meta:** date · SAPT · since when ("sub ghid de 3 zile", "de la 12:00", "de 1 zi"), author where one exists.
- **Status (right edge, fixed column ≥96px so lamps align):** `ip-lamp--warn` "abatere" (or "lipsă" for missing entries), `ip-lamp--fault` "refuzat", `ip-lamp--info` "informație". Never a record colour.
- **Action (optional):** "Deschide secțiunea" link before the status; hovering it fills the row `sunken`. Rows without an action keep the status column aligned.
- **Empty:** EmptyState inside the list: off lamp "nimic de verificat" + one sentence naming farm and date.
- Thresholds per D19.
