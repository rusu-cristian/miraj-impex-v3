"""sexsign.py · the ♀/♂ sign markup (SX01, D170, additions.css §72): one helper for every row and system preview that writes
a sign outside a SexCount figure span ("♀ <b>12 561</b>" stays s4s_build.hen/roo). No font in the stack has the signs, so a
plain glyph drew ~5 px: every sign is the §72 hook, the glyph aria-hidden + the word for screen readers (as sys11 `sexg`).
- sign('♀') / sign('F') → <span class="ip-sex ip-sex--hen" aria-hidden="true">♀</span><span class="ip-sr">găini</span>;
  sr='' leaves the word out (the word is written beside the sign), sr='…' says another word.
- figure(sx, fig) → sign + " " + fig in .ip-nowrap (the icon is an atomic inline: a line may break right after it).
- sexify(html) → the same string with every bare ♀/♂ of its text drawn as sign(): tags, attribute values, signs already in a
  hook (span.ip-sex / .is-hen / .is-rooster, td.ip-gimp__sex) and visually hidden text (.ip-sr, svg, title) are left as they are. No sr word
  where "găini"/"cocoși" is written right beside the sign ("♀ găini", "Găini ♀"), so screen readers do not hear it twice. A sign followed by its figure is kept with it (.ip-nowrap, keep()).
Pure string helpers: no data, imports nothing of the design. Checked by sex_scan.py (no plain sign on boards/previews)."""
import re

SEX = {'♀': ('hen', 'găini'), '♂': ('rooster', 'cocoși'), 'F': ('hen', 'găini'), 'M': ('rooster', 'cocoși')}
GLYPH = {'hen': '♀', 'rooster': '♂'}


def sign(sx, sr=None):
    """the §72 sign: glyph aria-hidden (drawn as the icon) + its word for screen readers (sr='' = none)"""
    tone, word = SEX[sx]
    w = word if sr is None else sr
    return f'<span class="ip-sex ip-sex--{tone}" aria-hidden="true">{GLYPH[tone]}</span>' + (f'<span class="ip-sr">{w}</span>' if w else '')


def figure(sx, fig, sr=None, num=False):
    """a sign with its figure written after it ("♀ 26 883"), kept on one line: the icon is an atomic inline, so a line may
    break after it even before a no-break space (CSS Text 3); .ip-nowrap holds the pair together (num: the figure in .ip-num)"""
    if num:
        fig = f'<span class="ip-num">{fig}</span>'
    return f'<span class="ip-nowrap">{sign(sx, sr)} {fig}</span>'


_SIGNH = r'<span class="ip-sex ip-sex--(?:hen|rooster)" aria-hidden="true">[\u2640\u2642]</span>(?:<span class="ip-sr">[^<]*</span>)?'
_PAIR = re.compile(r'(?<!<span class="ip-nowrap">)(' + _SIGNH + r'(?:[ \u00a0\u202f]|&#160;|&nbsp;)+'
                   r'(?:<span class="ip-num">[^<]*</span>|[\u2212+\-]?\d(?:[\d,.]|[\u202f\u00a0](?=\d))*%?))')


def keep(html):
    """every sign followed by its figure (plain digits or an .ip-num span) wrapped in .ip-nowrap, once"""
    return _PAIR.sub(r'<span class="ip-nowrap">\1</span>', html)


_TOK = re.compile(r'(<[^>]*>)')
_GL = re.compile('[♀♂]')
_OPEN = re.compile(r'<([a-zA-Z][\w-]*)([^>]*)>')
_VOID = {'area', 'base', 'br', 'col', 'embed', 'hr', 'img', 'input', 'link', 'meta', 'source', 'track', 'wbr'}
_SKIPTAG = {'svg', 'script', 'style', 'title', 'textarea', 'option'}
_HOOK = re.compile(r'''class\s*=\s*["'][^"']*\b(?:ip-sex(?:--hen|--rooster)?|is-hen|is-rooster|ip-gimp__sex|ip-sr)\b''')
_WORD = re.compile(r'(?i)(găini|cocoși)')


def _sr(text, i, nxt):
    """'' when the sign's word is written right beside it (after it, or just before it), else None (the default word)"""
    after = text[i + 1:].lstrip('   ')
    if not after and nxt is not None:
        after = nxt.lstrip('   ')
    before = text[:i].rstrip('   ')
    return '' if _WORD.match(after) or re.search(r'(?i)(găini|cocoși)$', before) else None


def sexify(html):
    parts = _TOK.split(html)
    out, stack = [], []
    for k, p in enumerate(parts):
        if k % 2:                      # a tag
            out.append(p)
            if p.startswith('</'):
                if stack:
                    stack.pop()
                continue
            m = _OPEN.match(p)
            if not m or p.startswith('<!') or p.endswith('/>') or m.group(1).lower() in _VOID:
                continue
            skip = (stack[-1] if stack else False) or m.group(1).lower() in _SKIPTAG or bool(_HOOK.search(m.group(2)))
            stack.append(skip)
            continue
        if (stack and stack[-1]) or not _GL.search(p):
            out.append(p)
            continue
        nxt = parts[k + 2] if k + 2 < len(parts) else None
        s, last = [], 0
        for g in _GL.finditer(p):
            s.append(p[last:g.start()]); s.append(sign(g.group(), _sr(p, g.start(), nxt))); last = g.end()
        s.append(p[last:])
        out.append(''.join(s))
    return keep(''.join(out))


if __name__ == '__main__':
    assert sign('F') == '<span class="ip-sex ip-sex--hen" aria-hidden="true">♀</span><span class="ip-sr">găini</span>'
    assert sexify('<b>♀</b> găini') == '<b><span class="ip-sex ip-sex--hen" aria-hidden="true">♀</span></b> găini'
    assert sexify('Găini ♀') == 'Găini <span class="ip-sex ip-sex--hen" aria-hidden="true">♀</span>'
    assert sexify('<span class="ip-sex ip-sex--rooster">♂ <b>3</b></span>') == '<span class="ip-sex ip-sex--rooster">♂ <b>3</b></span>'
    assert sexify('<i aria-label="♀">x</i><span class="ip-sr">♀</span>') == '<i aria-label="♀">x</i><span class="ip-sr">♀</span>'
    t = sexify('x ♀ <span class="ip-num">4 281</span> · ♂&#160;0 · ♀ 12 561<small>%</small>')
    assert t.count('ip-nowrap') == 3 and sexify(t) == t and keep(t) == t, t
    print('ok')
