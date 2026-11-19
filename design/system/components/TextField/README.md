# TextField

Text, decimal, password and note inputs.

- **Field:** `raised`, 1px `line-strong`, radius `radius-sm`. The label sits above in 12px `ink-secondary`.
- **Units** are a suffix: g, kg, lei/kg, %.
- **States:**
  - focus: the 2px `focus` ring on the field's own edge, inside its box (the field keeps its size; error/warn keep their hair outermost, the ring one hair inside), additions.css §68 (FW02, D149)
  - error: `fault` border + an error line
  - read-only: `sunken`
  - disabled
- **Phone:** height `control-h-touch`, 17px text.
- **In a table row** (DataTable inline edit, FW03, D150): no label above (the column head and the field's `aria-label` name it); its hint and error line go in the row's detail row (`tr.ip-table__detail`, additions.css §69), the error with the alert icon and `aria-errormessage`; the row keeps its height.
