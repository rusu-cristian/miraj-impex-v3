"""head_scan.py · the SYS17 check (D189): every table head cell on one line. Renders in headless Chrome (playwright,
/usr/bin/google-chrome, as the sNN_render scripts) every board that holds a <table> (design/canvas/project/*.dc.html at its
canvas.json w × h) and every system preview with a table (design/system/components/*/preview.html at 1440, as sys14_render).
python3 -I head_scan.py <repo_root>                    boards + previews
python3 -I head_scan.py <repo_root> --w 1280           desktop boards (w ≥ 1440) at 1280 instead of their canvas width
python3 -I head_scan.py <repo_root> FILE|STEM …        only these boards / previews
A th is reported when any of its text runs over more than one line. Its text is grouped by the box that lays it out (the
th itself, or a block / flex item inside it): a designed second line (a head note span.ip-ptable__note, a phone sub-line
span.ip-ptable__sub, a companion rate span.ip-ptable__pct) is its own box and counts as one line; a box whose text sits on
two line boxes is a wrapped head. Hidden text (.ip-sr, [hidden], display:none, the §72 signs at font-size 0) is skipped,
as are th not drawn (zero box). Every th is checked (column, group and section heads; row heads: PeriodTable week/day/foot,
DataTable row heads) except the phone CostBlock's cycle row heads (additions §11: `.ip-cost` without `--wide`, a 160 px
sticky column whose sub-line wraps by design; only the CostBlock preview draws it), counted apart as "excluded".
Prints file · th text · lines; exit 1 if any."""
import sys, json, asyncio, pathlib, re
from playwright.async_api import async_playwright

SCAN = """() => {
  const out = []; let excl = 0;
  const vis = e => { for (let x = e; x; x = x.parentElement) { if (x.hidden || x.classList.contains('ip-sr')) return false;
    const s = getComputedStyle(x); if (s.display === 'none' || s.visibility === 'hidden') return false; } return true; };
  const box = (e, th) => { for (let x = e; x && x !== th; x = x.parentElement) { if (!getComputedStyle(x).display.startsWith('inline')) return x; } return th; };
  for (const th of document.querySelectorAll('th')) {
    const r = th.getBoundingClientRect(); if (r.width < 1 || r.height < 1 || !vis(th)) continue;
    if (th.matches('.ip-cost:not(.ip-cost--wide) .ip-cost__table tbody th')) { excl++; continue; }
    const groups = new Map(); const w = document.createTreeWalker(th, NodeFilter.SHOW_TEXT); let n;
    while ((n = w.nextNode())) {
      if (!n.textContent.trim()) continue; const el = n.parentElement;
      if (!vis(el) || parseFloat(getComputedStyle(el).fontSize) === 0) continue;
      const rg = document.createRange(); rg.selectNodeContents(n);
      const rs = [...rg.getClientRects()].filter(q => q.width > 0.5 && q.height > 0.5);
      const b = box(el, th); if (!groups.has(b)) groups.set(b, []); groups.get(b).push(...rs);
    }
    let worst = 0;
    for (const rs of groups.values()) {
      const mids = rs.map(q => [q.top + q.height / 2, q.height]).sort((a, b) => a[0] - b[0]); let lines = 0, last = -1e9;
      for (const [m, h] of mids) { if (m - last > Math.max(4, h * 0.6)) { lines++; last = m; } }
      worst = Math.max(worst, lines);
    }
    if (worst > 1) out.push([worst, Math.round(r.height), th.textContent.replace(/\\s+/g, ' ').trim().slice(0, 60)]);
  }
  return {out, excl}; }"""


def targets(root, picks, w1280):
    proj = root / 'design/canvas/project'; sysd = root / 'design/system'
    c = json.loads((proj / 'canvas.json').read_text(encoding='utf-8'))['boards']
    out = []
    for k, v in c.items():
        f = proj / k
        if not f.exists() or '<table' not in f.read_text(encoding='utf-8'): continue
        w = v['w']
        if w1280 and w >= 1440: w = 1280
        out.append((f, f.as_uri(), w, v['h']))
    for f in sorted(sysd.glob('components/*/preview.html')):
        if '<table' not in f.read_text(encoding='utf-8'): continue
        wrap = pathlib.Path('/tmp/head_scan') / (f.parent.name + '.html'); wrap.parent.mkdir(exist_ok=True)
        wrap.write_text(f'''<!doctype html><html lang="ro"><head><meta charset="utf-8">
<link rel="stylesheet" href="{(sysd / 'tokens.css').as_uri()}"><link rel="stylesheet" href="{(sysd / 'components/bundle.css').as_uri()}">
<link rel="stylesheet" href="{(sysd / 'components/additions.css').as_uri()}">
<style>body{{margin:0;padding:24px;background:var(--ground)}}</style></head><body>{f.read_text(encoding="utf-8")}</body></html>''', encoding='utf-8')
        out.append((f, wrap.as_uri(), 1440, 1024))
    if picks:
        out = [t for t in out if any(p in str(t[0]) or t[0].name.startswith(p) for p in picks)]
    return out


async def main(argv):
    root = pathlib.Path(argv[1]).resolve(); rest = argv[2:]
    w1280 = '--w' in rest
    if w1280: i = rest.index('--w'); assert rest[i + 1] == '1280'; del rest[i:i + 2]
    ts = targets(root, rest, w1280)
    hits = []; sem = asyncio.Semaphore(4)
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        async def one(t):
            f, url, w, h = t
            async with sem:
                pg = await b.new_page(viewport={'width': w, 'height': min(h, 1200)})
                await pg.goto(url); await pg.wait_for_timeout(600)
                res = await pg.evaluate(SCAN); await pg.close()
            return f, w, res
        excl = 0
        for f, w, res in await asyncio.gather(*(one(t) for t in ts)):
            excl += res['excl']
            for lines, hh, txt in res['out']:
                hits.append(f'{f.relative_to(root)} @{w} · {txt!r} · {lines} lines ({hh}px)')
    print('\n'.join(hits))
    print(f'head_scan: {len(ts)} files with a table, {len(hits)} wrapped head cells ({excl} phone CostBlock row heads excluded)')
    return 1 if hits else 0


if __name__ == '__main__':
    sys.exit(asyncio.run(main(sys.argv)))
