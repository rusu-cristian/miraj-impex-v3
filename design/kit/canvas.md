# Design canvas: Impex Line Panel · Ecrane · runda 2

- **Current canvas** (created 19.11.2026 (round 2) from the FV03 coordinator session; old canvas e5d5275e… is stale, do not use):
  slug `c70b5e5e-ac76-40a2-ac41-f308b82af36b` · URL https://claude.ai/artifact/RaaG1ADkqBNH9w7DtfZcre · title "Impex Line Panel · Ecrane · runda 2".
  Made with `create_from_type` (Design type), filled from `design/canvas/project/` (271 files, 2 publish_files batches, patch) +
  `project/ds/impexlinepanel/tokens.json` by `copy_file` from the system artifact `95f22f4a-b2df-421d-b638-4d5e9d0656e2` (v `1794843774-ade9`)
  → 272 own files. Version after publish `1795123576-0152` (always re-fetch: `get_artifact`). canvas.json title now "… · runda 2".
  FINAL5 (20.11.2026): patch over `1795128533-4c97` with the 79 files changed since (77 boards, canvas.json, additions.css; D182–D189) → version `1795132564-e1da`, 272 own files.
  SYS18 (20.11.2026): patch over `1795160506-29b6` with 52 files (47 changed + 5 new boards incl. Main, canvas.json, additions.css; D190–D192 load-more row), remove `project/S3-seria-2026--toate-saptamanile.dc.html` (renamed `--inceputul-seriei`) → version `1795162838-7551`, 276 own files; hashes checked with list_files.
  SYS19 (26.11.2026): patch over `1795635990-4d48` with 4 files (additions.css §73, tokens.css `col-rule`; plus S11L's unpublished `S11-rapoarte-hrana--se-incarca.dc.html` and canvas.json, D193) → version `1795652130-c58a`, 276 own files; every local project file's sha256 = live (list_files).
  SYS20 (26.11.2026): patch over `1795652130-c58a` with 180 files (179 boards + additions.css §74; D195 equal-width control groups; canvas.json unchanged) → version `1795695290-d903`, 276 own files; every local project file's sha256 = live (list_files).
  VO02/VO03 (26.11.2026): published by the coordinator after the user's yes (subagent publish was refused; the artifact has outside guests): 3 patches over `1795703555-313e` — 24 files (13 VO boards, V1, S9 ×7, Main, additions.css §78, canvas.json), 128 boards, 128 boards → version `1795725498-f022`; every local project file's sha256 = live (list_files; tokens.json live-only as before).
  MO01 (26.11.2026): patch over `1795695290-d903` with 4 files (additions.css §75 + motion tokens + gentle reduced motion, D196; new interactive board `M1-miscare.dc.html` "M1 · Mișcare" as its own row at y 800 under Main, every screen row moved down 3592 by relayout; Main.dc.html lists it; canvas.json) → version `1795699223-f2dd`, 277 own files; every local project file's sha256 = live (list_files).
  MO02 (26.11.2026, audit fixes): patch over `1795699223-f2dd` with additions.css (§75: popover origins top left / top right incl. `.ip-users__menu`, SidePanel slides without a fade outside reduced motion, press skips `aria-busy`) + `M1-miscare.dc.html` (tabs' selected state) → version `1795699746-d349`, 277 own files.
  MO03 (26.11.2026, live updates D197): patch over `1795699746-d349` with 7 files (additions.css §76 + `--ip-dur-tint`/`--ip-live-mix`; `M1-miscare.dc.html` with the live card, h 3472 → 4144, every screen row moved down 672 by relayout; 3 new boards `S4-ferma-rezumat--rand-nou`, `S2-acasa--proprietar-desktop--inregistrari-noi`, `S3-seria-2026--transfer-azi` (phones after them moved right); Main.dc.html; canvas.json) → version `1795701704-d3a2`, 280 own files.
  MO03 fix 1 (26.11.2026, audit): patch over `1795701704-d3a2` with additions.css (§76: old figure out at 120 ms, new in 60 ms later; transitionend / closed→open notes) + `M1-miscare.dc.html` (tint and old figure removed on transitionend, bar closes at the list top) → version `1795702064-4234`, 280 own files.
  MO04 (26.11.2026, D198): patch over `1795702064-4234` with 3 files (additions.css §77 Tabs underline / segment thumb slide + `--ip-dur-slide`, §71 full-cell week key, §75 skeleton → content reveal; `M1-miscare.dc.html` h 4144 → 4692 with the phone strip and the Reîncarcă card, every screen row moved down 548 by relayout; canvas.json; no other board changed: CSS only) → version `1795703555-313e`, 280 own files.
  Only the session that made it can publish to it (`network_off_limit` elsewhere): later sessions hand files to whoever can.
- STALE (do not use) — previous canvas: slug `e5d5275e-4850-45ce-afa4-d035b0a5bd2b` · URL https://claude.ai/artifact/VP5iig8NNXWPVRw1yS9WUi (created 19.11.2026 by a design-loop session and filled from `design/canvas/project/`; version after the full publish `1795075243-a466`).
- FV03 (19.11.2026): live version `1795092797-ff63`; FV03's patch (245 changed + 12 added files, remove `project/S2-acasa--zi-trecuta-blocat.dc.html`, `project/S15-introdus-azi--manager-10-11.dc.html`) is NOT published: the helper session got `network_off_limit` (it did not create the artifact), so the coordinator publishes it (2 batches ≤ 256, mode "patch").
- Old canvas (16–17.11.2026, stale since 16.11, read-only from later sessions): slug `e44dd7fb-2990-41f9-a791-2841f1b1dcd1`, https://claude.ai/artifact/VC8kqksPgxxYo9kPpK14yv. Do not publish there.
- Type: Design (canvas), `bcd2878f-a88d-4094-a616-8e39d8a5fd8a`. Version at setup: `1794844216-3062` (always re-fetch: `get_artifact`).
- Design system installed: Impex Line Panel (`95f22f4a-…`, v `1794843774-ade9`) as folder `impexlinepanel`.
- Setup sub-agent could call publish_files and copy_file: **yes**. Caveat: the server only lets a session change artifacts it made
  (`network_off_limit`); if a helper's publish is refused for that reason, hand the files to the coordinator to publish.

## File layout (only files under `project/` show; local mirror: `design/canvas/project/`)
- `project/canvas.json` — index (content.index). `boards` keyed by file name, `order`, `designSystems` (leave as is).
- `project/Main.dc.html` — artboard "00 · Index" at x 0, y 0, 880×560.
- `project/ds/impexlinepanel/tokens.json` (copied from the system artifact's `project/tokens.json`), `project/ds/impexlinepanel/tokens.css` (CSS vars + type classes with `.ip-t-` aliases; published from the repo's `design/system/tokens.css`, the single source), `tokens.json` (Theme menu), `components/bundle.css` (`ip-` classes; @imports Plex fonts).
- `project/additions.css`: the project's additions (D32: interaction states, motion vars, new components), published from `design/system/components/additions.css`. Kept outside `ds/` because the type reserves `project/ds/<folder>/` for the system install.
- One screen = one file `project/<stem>.dc.html`, flat (no subfolders). Stem: ASCII, letters/digits/`_`/`-`/`.`, unique.
  "S4 · Fermă · Rezumat" → `S4-ferma-rezumat.dc.html`; state "S4 · Fermă · Rezumat · se încarcă" → `S4-ferma-rezumat--se-incarca.dc.html`.
  The human name (with diacritics) goes in the board's `title`.

## Sizes
- Desktop (PAGE): `w` 1440, `"expand":"fill"`; root fluid (no px width; `max-width` container, %/fr), `h` = its height at 1440 (e.g. 1024; over-tall beats clipped).
- Phone: 390×844, root FIXED `width:390px;height:844px` (taller `h` only for a long scrolling form). No fake status bar.
- `$preview` in data-props = the board's `w`×`h`. Gaps: 80 px between boards in a row, 120 px between rows.
- Wide tables (D167, S4NG): the canvas is an example, so no board shows a clipped sideways scroll. When a table is wider than
  its page at 1440, the board grows wider: `w` = 1440 + the table's overflow, up to the next 8 px (listed in
  `design/canvas/scripts/board_widths.json`, measured by `s4s_render.py --widths`; `canvas_rows.place` and `s4s_build.write`
  read it). The page keeps its 1440 layout (beside a docked SidePanel: 1440 less the panel) and the table region grows past
  the page's right edge to the table's full width (board CSS `.s4s-wide`, canvas only); the product keeps its scroll region
  and sticky Data column. `canvas_rows.py relayout` pushes the later boards of the row right so every gap stays ≥ 80.
- Layout: one row per screen, starting y = 800. Default state at x 0, its states to the right (x += w + 80), phone boards after the desktop ones. Next row y = row y + tallest h + 120.

## Turning a self-contained screen HTML into an artboard (`.dc.html`)
Shared CSS is NOT re-published per screen: link the three shared files in this order, tokens.css → bundle.css → additions.css
(relative, flat path, no `../`; exact hrefs `ds/impexlinepanel/tokens.css`, `ds/impexlinepanel/components/bundle.css`, `additions.css`):
```html
<!doctype html>
<html lang="ro">
<head>
<meta charset="utf-8">
<title>S4 · Fermă · Rezumat</title>
<script src="./support.js"></script>
<link rel="stylesheet" href="ds/impexlinepanel/tokens.css">
<link rel="stylesheet" href="ds/impexlinepanel/components/bundle.css">
<link rel="stylesheet" href="additions.css">
</head>
<body>
<x-dc>
<helmet>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500&amp;family=IBM+Plex+Mono:wght@400;500&amp;display=swap">
<style>body{margin:0} /* + screen-only CSS rules from the source <style>, if any */</style>
</helmet>
<div class="ip-root" style="padding:0; min-height:1024px"> …the screen's <body> markup… </div>
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{"$preview":{"width":1440,"height":1024}}'>
class Component extends DCLogic { renderVals() { return {}; } }
</script>
</body>
</html>
```
Rules (each fails silently): keep the `support.js` line exactly; drop the source's inlined tokens/bundle CSS (linked above) and any
`<script>` that builds UI; close every element, quote every attribute; no remote images, iframes, emoji; icons = inline stroke SVG
(Lucide-style, 1.75); real `<button>`/`<a href>`/`<input>`+`<label>`. Static comps: no `is_interactive`.
Type: use the `ip-t-<name>` classes (`ip-t-title`, `ip-t-body`, `ip-t-small`, `ip-t-caption`, `ip-t-readout`, …), not the bare `.title`/`.body`.
States on artboards: real `:hover`/`:active`/`:disabled` come from additions.css; to show a state statically use the demo classes (`is-hover`, `is-pressed`, `is-disabled`, `ip-focus`/`is-focus`).

## Adding one screen (one publish per call, sequential — never two publishes in parallel)
1. `get_artifact` slug → live `version`. `read_file` slug, path `project/canvas.json` → write it to `design/canvas/project/canvas.json`.
2. Write `design/canvas/project/S4-ferma-rezumat.dc.html`. In canvas.json add (keep every other key/entry untouched):
   `"boards": {"S4-ferma-rezumat.dc.html": {"x":0,"y":800,"w":1440,"h":1024,"expand":"fill","title":"S4 · Fermă · Rezumat"}}`
   (phone: `{"x":…,"y":…,"w":390,"h":844,"title":"S4 · Fermă · Rezumat · telefon"}`) and push the path onto `order`.
   Optionally update Main.dc.html's list.
3. publish_files:
```json
{"slug":"c70b5e5e-ac76-40a2-ac41-f308b82af36b","base_version":"<live version>","mode":"patch",
 "files":[{"as":"project/S4-ferma-rezumat.dc.html","path":"/home/claude/miraj-impex-v3/design/canvas/project/S4-ferma-rezumat.dc.html"},
          {"as":"project/canvas.json","path":"/home/claude/miraj-impex-v3/design/canvas/project/canvas.json"}]}
```
   Keep the returned `version` for the next call. Replacing a screen = same `as`, no canvas.json unless layout changes.
   Remove = `"remove":["project/<stem>.dc.html"]` plus taking it out of `boards` and `order`.
4. Conflict: re-run step 1, redo the canvas.json edit on the fresh copy, retry (stop after 3). Do not render or read back to verify.
- Never use publish_version/mode "replace" on this canvas. Updating shared CSS: edit the repo file first (`design/system/tokens.css` or
  `design/system/components/additions.css`), copy it to the mirror, then publish it at `project/ds/impexlinepanel/tokens.css` or `project/additions.css`.
  bundle.css is the system's and stays unchanged; additions and fixes go in additions.css. (SYS00 published both; canvas version after SYS00 fix 2: `1794845755-a7ea`; after SYS01: `1794846357-5576`; after SYS01 fix 1: `1794847363-9139`; after SYS01 fix 2: `1794847872-f8cd`; after SYS02: `1794848329-7c69`; after SYS03 fix 1: `1794849537-0345`; after SYS04: `1794849991-b4dd`; after SYS04 fix 1: `1794850418-ec52`; after SYS05: `1794850956-bcf8`; after SYS06: `1794852762-3f0c`; after SYS06 fix 1: `1794853378-53b6`; after SYS07 fix 1: `1794854256-1a8a`.)
