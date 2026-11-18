"""Skeleton markup shared by the loading boards (FIN11G, additions.css section 54): every skeleton takes the loaded block's
own boxes, so nothing moves when data arrives (craft S6P: the "se încarcă" skeleton is the loaded geometry).
- readout(): a Readout's own line elements, one bar each (StockReadout default: cap, value, 3 kv rows, sub line);
  chart_readout(): the same for a Readout holding a GuideChart (its empty svg at the loaded viewBox);
  stat(): a Stat's own lines, guide lines as wide as the loaded ones (so they wrap alike);
- track(compact): the CycleTrack skeleton at the drawn track's height (full 120, compact 90);
- field(box): the loaded field's label line, its own (empty, sunken) box and any hint lines;
- line(): one text line of bars; widths(): the column shares a loaded table and its loading skeleton both declare.
Used by s2s, s8s, s11 (S12/S13 through it), s16s, s6s, s7s, s18/s18s, s19/s19s, s20/s20s, s21, s22."""


def bar(cls=''):
    return f'<span class="ip-skel{(" " + cls) if cls else ""}"></span>'


def readout(variant='stock', kv=3, sub=1, live=False):
    """div.ip-readout.ip-readout--<variant>.ip-readout--skel. live=True: cap, value and sub inside .ip-readout__live,
    then the kv list (S8 comparison Readout); otherwise cap, value, kv list, sub lines (StockReadout)."""
    cap = f'<div class="ip-readout__cap"><span>{bar("ip-skel--w2")}</span></div>'
    val = f'<div class="ip-readout__val">{bar("ip-skel--w3")}<small>{bar()}</small></div>'
    subs = f'<div class="ip-readout__sub">{bar("ip-skel--w1")}</div>' * sub
    kvs = '<dl class="ip-readout__kv">' + f'<div><dt>{bar("ip-skel--w3")}</dt><dd>{bar()}</dd></div>' * kv + '</dl>'
    body = f'<div class="ip-readout__live">{cap}{val}{subs}</div>{kvs}' if live else f'{cap}{val}{kvs}{subs}'
    v = f' ip-readout--{variant}' if variant else ''
    return f'<div class="ip-readout{v} ip-readout--skel" aria-hidden="true">{body}</div>'


def chart_readout(cls, viewbox):
    """A Readout holding a GuideChart (S4 / S14 key Readout): cap, value, sub, then the chart's own empty svg (the loaded
    viewBox, so it is as tall as the drawn chart at any width; sunken), the keyboard read line and the legend line."""
    return (f'<div class="ip-readout {cls} ip-readout--skel" aria-hidden="true"><div class="ip-readout__cap"><span>{bar("ip-skel--w2")}</span></div>'
            f'<div class="ip-readout__val">{bar("ip-skel--w3")}<small>{bar()}</small></div><div class="ip-readout__sub">{bar("ip-skel--w1")}</div>'
            f'<svg class="ip-chart" viewBox="{viewbox}"></svg><div class="ip-chart__read"><span>{bar("ip-skel--w2")}</span></div>'
            f'<div class="ip-chart__legend"><div>{bar("ip-skel--w1")}</div></div></div>')


def stat(guides):
    """div.ip-stat.ip-stat--skel: the Stat's own lines (cap, value, its guide lines). guides: per guide line, the loaded
    line's items (px, their one-line width, measured): each item is two touching bars of that total width with a break
    opportunity (wbr) between them, so a guide line wraps at the same width as the loaded text (S14 at 1280)."""
    def item(w):
        a = round(w / 8, 2); b = round(w / 4 - a, 2)
        return (f'<span><span class="ip-skel" style="width:calc(var(--space-1) * {a:g})"></span><wbr/>'
                f'<span class="ip-skel" style="width:calc(var(--space-1) * {b:g})"></span></span>')
    g = ''.join('<span class="ip-stat__guide">' + ''.join(item(w) for w in line) + '</span>' for line in guides)
    return (f'<div class="ip-stat ip-stat--skel" aria-hidden="true"><span class="ip-cap">{bar("ip-skel--w2")}</span>'
            f'<span class="ip-stat__val">{bar("ip-skel--w3")}</span>{g}</div>')


def track(compact=False):
    return f'<div class="ip-trackskel{" ip-trackskel--compact" if compact else ""}" aria-hidden="true">{bar() * 3}</div>'


def legend():
    """The CycleTrack legend line under a track skeleton (its own ip-tracklegend box, one bar), so what follows keeps its place."""
    return f'<div class="ip-tracklegend" aria-hidden="true"><div class="ip-skel-line">{bar("ip-skel--w1")}</div></div>'


BOX = {'ctl': 'ip-input', 'touch': 'ip-input ip-input--touch', 'count': 'ip-input ip-count',
       'area': 'ip-input ip-input--area ip-input--touch'}


def field(box='touch', hints=0, label=True, cls='', rows=3):
    """div.ip-field.ip-field--skel: label line (bar), the loaded field's box (empty; an area holds `rows` empty lines,
    as its textarea), then one hint of `hints` lines (bar each). cls: extra classes the loaded field has (grid placement)."""
    inner = ('<span>' + line('') * rows + '</span>') if box == 'area' else ''
    lab = f'<span class="ip-label">{bar()}</span>' if label else ''
    hs = ('<span class="ip-hint">' + line('') * hints + '</span>') if hints else ''  # one hint, `hints` lines
    c = f' {cls}' if cls else ''
    return f'<div class="ip-field ip-field--skel{c}" aria-hidden="true">{lab}<span class="{BOX[box]}">{inner}</span>{hs}</div>'


def line(*classes, cls=''):
    c = f' {cls}' if cls else ''
    return f'<span class="ip-skel-line{c}">{"".join(bar(k) for k in classes)}</span>'


def widths(head, pcts, mins=None):
    """The loaded table's column shares, declared on its head: each leaf head cell (colspan 1) of `head` (a thead, its
    rows, or a row's th cells) gets style="--ip-colw:<n>%" (additions.css §54 sets the width from it, at zero specificity,
    so a screen can turn it off where its columns fold beside a SidePanel), one share per column in column order
    (rowspan/colspan aware).
    The loaded and the loading table use the same head, so both lay their columns out the same (§54, FIN11G); a column a
    container query hides takes its share with it. One share is None: that column (the widest, text) takes the rest.
    mins: each column's loaded min-content width (px, measured, without the cell padding), set as --ip-colmin (in
    space-1 units): §54 puts a zero-height strut of that width in the head cell, so where the shares are narrower than
    the loaded content (1280) the loading columns grow exactly as the loaded ones do."""
    import re
    assert pcts.count(None) == 1 and sum(p for p in pcts if p) < 100, pcts
    assert mins is None or len(mins) == len(pcts), (mins, pcts)
    rows = re.findall(r'<tr[^>]*>(.*?)</tr>', head, re.S) or [head]
    taken, col_of = {}, []          # taken[(row, col)] by a rowspan from above; col_of: (row, cell index) -> first column
    for r, row in enumerate(rows):
        c = 0
        for i, m in enumerate(re.finditer(r'<th\b[^>]*>', row)):
            while taken.get((r, c)): c += 1
            cs = int((re.search(r'colspan="(\d+)"', m.group(0)) or [0, 1])[1]); rs = int((re.search(r'rowspan="(\d+)"', m.group(0)) or [0, 1])[1])
            for dr in range(rs):
                for dc in range(cs): taken[(r + dr, c + dc)] = True
            col_of.append((r, i, c if cs == 1 else None)); c += cs
    ncol = max(c for (_, c) in taken) + 1
    assert ncol == len(pcts), (ncol, len(pcts))
    leaf = {(r, i): c for r, i, c in col_of if c is not None}
    assert sorted(leaf.values()) == list(range(ncol)), leaf
    out, pos = [], 0
    for r, row in enumerate(rows):
        def put(m, r=r, it=iter(range(10 ** 6))):
            i = next(it); c = leaf.get((r, i))
            if c is None: return m.group(0)
            st = ([f'--ip-colw:{pcts[c]:g}%'] if pcts[c] is not None else []) + ([f'--ip-colmin:calc(var(--space-1) * {round(mins[c] / 4, 2):g})'] if mins else [])
            return m.group(0)[:-1] + f' style="{";".join(st)}">' if st else m.group(0)
        new = re.sub(r'<th\b[^>]*>', put, row)
        k = head.index(row, pos); out.append(head[pos:k] + new); pos = k + len(row)
    return ''.join(out) + head[pos:]
