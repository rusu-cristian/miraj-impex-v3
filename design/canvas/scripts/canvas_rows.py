"""canvas_rows.py · shared argv + canvas.json row placement for the sNN generators (FIN02G); see README.md.
- args(): <repo_root> [heights.json] [--canvas]; the heights file defaults to this folder's sNN_heights.json.
- row_y(): a row keeps the y of its existing entries; only a row with no entry yet goes below the lowest other row.
- place(): existing boards keep x and y (w, h, title refreshed in place); a run that adds a board lays the given sequence out
  from x0 (80 px gaps), moving the listed existing boards with it. Only called under --canvas.
- save(): refuses overlapping boards; writes only when the text changes (same json format as before).
- relayout() / `python3 -I canvas_rows.py relayout <repo_root> [--canvas]` (RF01): re-stack the rows when board heights
  change. A row = the boards sharing one y. Rows above ROW0 (the 00 · Index board) and the first row at ROW0 keep their y;
  every later row gets y = previous row y + its tallest h + GAP_Y (120). x, w, h never change. Prints each moved row; writes
  canvas.json only with --canvas (through save(), so overlapping boards stop it). Run it after a row's --canvas rebuild.
  S4NG (D167): it also lays each row out along x: a board that grew wider pushes the boards to its right so every gap stays
  at least GAP_X (80); boards never move left.
- Wide boards (D167 › 4): board_widths.json {stem: w} lists the boards drawn wider than their base width so a wide table
  shows whole (no clipped sideways scroll on the canvas). place() takes w from it, s4s_build.write the board's $preview width."""
import sys, json, pathlib

HERE = pathlib.Path(__file__).resolve().parent
GAP_X, GAP_Y = 80, 120
WIDTHS = json.loads((HERE / 'board_widths.json').read_text(encoding='utf-8')) if (HERE / 'board_widths.json').exists() else {}
def width(stem, w):
    """the board's drawn width: its base width (1440 / 390) or the wider one board_widths.json gives it (D167)"""
    return max(w, WIDTHS.get(stem, w))


def args(heights=None, argv=None):
    """(root, heights path or None, canvas flag) from sys.argv; '--canvas' may sit anywhere after the root."""
    argv = sys.argv if argv is None else argv
    a = [x for x in argv[1:] if x != '--canvas']
    hp = pathlib.Path(a[1]) if len(a) > 1 else (HERE / heights if heights else None)
    return pathlib.Path(a[0]), hp, '--canvas' in argv[1:]


def path(root):
    return pathlib.Path(root) / 'design/canvas/project/canvas.json'


def load(root):
    return json.loads(path(root).read_text(encoding='utf-8'))


def in_row(key, stem):
    return key == f'{stem}.dc.html' or key.startswith(f'{stem}--')


def below_lowest(c, stem):
    others = [v for k, v in c['boards'].items() if not in_row(k, stem)]
    ly = max(v['y'] for v in others)
    return ly + max(v['h'] for v in others if v['y'] == ly) + GAP_Y


def row_y(c, stem):
    """y of the row of `stem` (its default board's entry, else any entry of the row), else below the lowest other row."""
    b = c['boards']
    if f'{stem}.dc.html' in b: return b[f'{stem}.dc.html']['y']
    ys = [v['y'] for k, v in b.items() if in_row(k, stem)]
    return ys[0] if ys else below_lowest(c, stem)


def place(c, stem, seq, x0=0):
    """seq: (board_stem, w, h, title) for boards this script writes, or a bare board_stem for an existing board it only
    positions. Returns the x after the last board."""
    B = c['boards']; y = row_y(c, stem)
    fresh = any(not isinstance(s, str) and f'{s[0]}.dc.html' not in B for s in seq)
    x = x0
    for s in seq:
        if isinstance(s, str):
            f = f'{s}.dc.html'; assert f in B, f'{f} has no canvas entry: run its base build with --canvas first'
            e = B[f]
        else:
            s, w, h, t = s; f = f'{s}.dc.html'; e = B.get(f); w = width(s, w)
            if e is None:
                e = B[f] = {'x': x, 'y': y, 'w': w, 'h': h, **({'expand': 'fill'} if w == 1440 else {}), 'title': t}
            else:
                e.update(w=w, h=h, title=t)  # in place: key order (and so the file) unchanged
        if fresh: e['x'] = x
        if f not in c['order']: c['order'].append(f)
        x += e['w'] + GAP_X
    spread(c, [k for k, v in B.items() if v['y'] == y])   # D167: a board drawn wider pushes its row's later boards right
    return x


def overlaps(c):
    v = sorted(c['boards'].items(), key=lambda kv: (kv[1]['y'], kv[1]['x']))
    bad = []
    for i, (k, a) in enumerate(v):
        for l, b in v[i + 1:]:
            if b['y'] >= a['y'] + a['h']: break
            if a['x'] < b['x'] + b['w'] and b['x'] < a['x'] + a['w']: bad.append((k, l))
    return bad


def save(root, c):
    bad = overlaps(c); assert not bad, f'canvas.json boards overlap: {bad[:3]}'
    txt = json.dumps(c, ensure_ascii=False, indent=2) + '\n'
    p = path(root)
    if p.read_text(encoding='utf-8') != txt: p.write_text(txt, encoding='utf-8')
    return txt


ROW0 = 800  # first screen row (canvas.md › Sizes)


def rows(c):
    """[(y, [board keys])] in y order."""
    by = {}
    for k, v in c['boards'].items(): by.setdefault(v['y'], []).append(k)
    return sorted(by.items())


def spread(c, keys):
    """D167: within one row, push boards right so each starts at least GAP_X after the one before it (wider boards);
    returns [(board, old x, new x)]"""
    out, end = [], None
    for k in sorted(keys, key=lambda k: c['boards'][k]['x']):
        e = c['boards'][k]
        if end is not None and e['x'] < end + GAP_X:
            out.append((k, e['x'], end + GAP_X)); e['x'] = end + GAP_X
        end = e['x'] + e['w']
    return out


def relayout(c):
    """Re-stack the rows from ROW0 down in place; returns [(first board of the row, old y, new y)] for the rows that moved
    (and, as (board, 'x', old x, new x), the boards a wider neighbour pushed right)."""
    moved, prev = [], None
    for y, keys in rows(c):
        if y < ROW0: continue
        moved += [(k, 'x', a, b) for k, a, b in spread(c, keys)]
        ny = y if prev is None else prev[0] + prev[1] + GAP_Y
        if ny != y:
            for k in keys: c['boards'][k]['y'] = ny
            moved.append((min(keys, key=lambda k: c['boards'][k]['x']), y, ny))
        prev = (ny, max(c['boards'][k]['h'] for k in keys))
    return moved


if __name__ == '__main__' and sys.argv[1:2] == ['relayout']:
    root = pathlib.Path(next(a for a in sys.argv[2:] if a != '--canvas')); c = load(root)
    for m in relayout(c):
        if m[1] == 'x': print(f'{m[0]}: x {m[2]} → {m[3]} ({m[3] - m[2]:+d})')
        else: k, a, b = m; print(f'{k}: y {a} → {b} ({b - a:+d})')
    bad = overlaps(c)
    if bad: sys.exit(f'overlap: {bad[:3]}')
    print('rows', len([y for y, _ in rows(c) if y >= ROW0]), 'ok' if '--canvas' not in sys.argv else 'saved')
    if '--canvas' in sys.argv: save(root, c)
