# Tabs

View tabs. The selected tab is `ink` with a 2px `ink` underline and a small `ok-lamp` dot. Others are `ink-secondary`.

## Project note (repo copy only, SYS20, 26.11.2026)
- **Equal tabs (D195, additions.css §74):** `div.ip-tabs` is a grid as wide as n × the widest tab (each tab padded `space-4` each side, label centred, one line); its line runs under the tabs. The selected tab's dot hangs left of its label (negative margin = dot + gap), so a tab's width never depends on which tab is selected. On a phone (`.ip-screen .ip-tabs`) the list spans the page and the tabs share it equally (S4: 5 × 71.6 in 358), the dot in the label line. Spec: components.md › Additions › SYS20. To copy into the system artifact (HANDOFF).
