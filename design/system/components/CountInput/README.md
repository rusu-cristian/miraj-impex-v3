# CountInput

Whole-number input for birds and eggs. The sex is marked by a 3px inner left bar in `rec-hen` or `rec-rooster`, and by ♀ / ♂ in the label.

- **Empty:** "–", meaning null, not 0.
- **Large:** height `count-input-h`, style `readout` (Plex Mono 20).
- **States:** warning (`warn-lamp` border + a line on `warn-bg`), error (`fault`), saved (read-only).
- **Input:** `inputmode="numeric"`. Enter moves to the next field.
