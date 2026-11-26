"""vkey.py · VO02: the voice key on every signed-in board (spec §8 › Where, screens.md › Voice placement table, D199).
`add(markup)` takes a board's root markup and places the one VoiceButton the shell carries; every board writer calls it on
the markup it writes (the .dc.html and its /tmp render wrapper), so a rebuilt row always holds the key. S1 (signed out)
never calls it. The key itself is VO01's `key()` (design/system/scripts/vo01_build.py, imported, not copied).
- Desktop (div.ip-shell): the header key, the last key of the first PageHeader's actions (span.ip-voiceanchor; in the key
  row when the actions are a key row over a note, S20; a page with no actions gets one div.ip-row holding it); it stays
  there with a SidePanel docked; under a Dialog scrim it is inert.
- Phone (div.ip-screen, or the S16 / S18 full-screen sheet aside.ip-sidepanel--sheet):
  - BottomSheet open (div.ip-scrim--sheet > .ip-sheet: Înregistrează, FilterBar sheet, SectionSelect): the first child of the
    scrim, space-4 above the sheet's top edge (§78, no measuring);
  - a form: the floating key is the SaveBar's last child, anchored space-4 above the bar itself (§78, VO03: any bar height);
  - else the floating key (.ip-voicekey--float), the last child of the full-screen sheet (space-4 above its foot) or of the
    screen (space-4 above the BottomTabBar, §78 --ip-voice-bar); a screen that is not fixed to its frame
    (a tall board drawn whole) is the board itself: it is positioned (ip-scrimhost) and the key placed in it (absolute;
    the app uses fixed);
  - a Dialog open (div.ip-scrim--phone): the key stays at its place under the scrim, inert (voice is off inside Dialogs);
  - a field focused (span.ip-input with is-focus holding an input or textarea, CountInput included): the soft keyboard is
    open, the key is `hidden` (§78), as the app does.
Idempotent: markup that already holds a key (the VO boards, which place their own) is returned unchanged."""
import re, sys, pathlib, importlib.util

HERE = pathlib.Path(__file__).resolve().parent
_spec = importlib.util.spec_from_file_location('vo01_build', HERE.parents[1] / 'system/scripts/vo01_build.py')
VO = importlib.util.module_from_spec(_spec)
_argv = sys.argv[:]
sys.argv = [sys.argv[0], str(HERE.parents[2])]   # vo01_build reads the repo root from argv[1]
_spec.loader.exec_module(VO)
sys.argv = _argv

UID = 'vk'


def _close(html, i, tag):
    """index just past the element that opens at html[i] (<tag …>), counting nested <tag / </tag>"""
    depth = 0
    for m in re.finditer(rf'<(/?){tag}\b[^>]*>', html[i:]):
        depth += -1 if m.group(1) else 1
        if depth == 0:
            return i + m.end()
    raise ValueError(f'unclosed <{tag}>')


def _keyboard(html):
    """a focused text field (not a CountInput, not a select-only combobox): the phone's soft keyboard is open"""
    for m in re.finditer(r'<span class="(ip-input\b[^"]*)"[^>]*>', html):
        cls = m.group(1).split()
        if 'is-focus' not in cls or 'ip-select__btn' in cls:
            continue
        inner = html[m.end():_close(html, m.start(), 'span')]
        if re.search(r'<(input|textarea)\b(?![^>]*type="(checkbox|radio)")', inner):
            return True
    return False


def phone_key(html, inert=False, absolute=False, hidden=False, floating=True):
    k = VO.key('fab', 'rest', UID, float_=floating, hidden=hidden)
    if absolute:
        k = k.replace('<button type="button" class="ip-voicekey', '<button type="button" style="position:absolute" class="ip-voicekey', 1)
    if inert:
        k = k.replace('<button type="button"', '<button type="button" inert', 1)
    return k


def desk_key(inert=False):
    return f'<span class="ip-voiceanchor"{" inert" if inert else ""}>{VO.key("head", "rest", UID)}</span>'


def add(html):
    if 'ip-voicekey' in html:
        return html
    if 'class="ip-screen' in html or 'ip-sidepanel--sheet' in html:
        return _phone(html)
    if '<header class="ip-pagehead' in html:
        return _desk(html)
    return html


def _phone(html):
    dialog = 'ip-scrim--phone' in html
    kb = _keyboard(html) and not dialog
    m = re.search(r'<div class="ip-scrim ip-scrim--sheet\b[^"]*"[^>]*>', html)
    if m and re.search(r'class="ip-sheet\b', html[m.end():_close(html, m.start(), 'div')]):
        # a BottomSheet is open: the key on its scrim, above the sheet's top edge (not floating, §78)
        return html[:m.end()] + phone_key(html, floating=False) + html[m.end():]
    m = re.search(r'<aside class="ip-sidepanel ip-sidepanel--sheet\b[^"]*"[^>]*>', html)
    if m:
        end = _close(html, m.start(), 'aside') - len('</aside>')
        return html[:end] + phone_key(html, inert=dialog, hidden=kb) + html[end:]
    if re.search(r'<div (?:inert )?class="ip-savebar ip-screen__savebar', html):
        return in_savebar(html, phone_key(html, inert=dialog, hidden=kb))
    m = re.search(r'<div class="ip-screen\b([^"]*)"[^>]*>', html)
    fixed = 'ip-screen--fixed' in m.group(1)
    end = _close(html, m.start(), 'div') - len('</div>')
    out = html[:end] + phone_key(html, inert=dialog, absolute=not fixed, hidden=kb) + html[end:]
    if not fixed and 'ip-scrimhost' not in m.group(1):
        tag = m.group(0).replace('class="ip-screen', 'class="ip-screen ip-scrimhost', 1)
        out = out[:m.start()] + tag + out[m.end():]
    return out


def in_savebar(html, key):
    """the key as the SaveBar's last child: §78 anchors it space-4 above the bar's top edge, whatever the bar's height"""
    m = re.search(r'<div (?:inert )?class="ip-savebar ip-screen__savebar\b[^"]*"[^>]*>', html)
    end = _close(html, m.start(), 'div') - len('</div>')
    return html[:end] + key + html[end:]


def _desk(html):
    inert = 'ip-scrim--page' in html
    m = re.search(r'<header class="ip-pagehead[^"]*"[^>]*>', html)
    end = _close(html, m.start(), 'header') - len('</header>')
    head = html[m.end():end]
    # the header's top-level children: the title block, then (if any) the actions block
    kids, i = [], 0
    while True:
        n = re.search(r'<(div|nav|span)\b', head[i:])
        if not n:
            break
        s = i + n.start(); e = _close(head, s, n.group(1)); kids.append((s, e)); i = e
    if len(kids) >= 2:
        s, e = kids[-1]
        cut = m.end() + e - len('</div>')
        row = re.match(r'<div class="[^"]*">\s*(<div class="[^"]*keyrow[^"]*">)', head[s:e])   # S20: a key row over a note
        if row:
            r0 = m.end() + s + row.start(1); cut = _close(html, r0, 'div') - len('</div>')
        assert html[cut:cut + 6] == '</div>', html[m.end() + s:m.end() + e][:120]
        return html[:cut] + desk_key(inert) + html[cut:]
    return html[:end] + f'<div class="ip-row">{desk_key(inert)}</div>' + html[end:]
