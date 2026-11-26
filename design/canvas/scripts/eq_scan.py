"""eq_scan.py · the SYS20 check (D195, additions.css §74): every control group's items have one rendered width (±1 px).
Renders in headless Chrome (playwright, /usr/bin/google-chrome, as head_scan.py) every board (design/canvas/project/*.dc.html
at its canvas.json w × h) and every system preview (design/system/components/*/preview.html at 1440).
python3 -I eq_scan.py <repo_root>                 boards + previews
python3 -I eq_scan.py <repo_root> FILE|STEM …      only these boards / previews
Groups: SegmentedControl (.ip-seg > .ip-seg__item), Tabs (.ip-tabs > .ip-tab), BottomTabBar (.ip-tabbar > .ip-tabbar__item)
and RecordChip groups (any element holding two or more .ip-chip children); only drawn items count (a group needs two).
A group is reported when its widest and narrowest item differ by more than 1 px, or when an item's label overflows its
item or the group runs past its parent's right edge. Prints file · kind · labels · widths; exit 1 if any."""
import sys, json, asyncio, pathlib
from playwright.async_api import async_playwright

SCAN = """() => {
  const vis = e => { const r = e.getBoundingClientRect(); if (r.width < 0.5 || r.height < 0.5) return false;
    for (let x = e; x; x = x.parentElement) { const s = getComputedStyle(x); if (s.display === 'none' || s.visibility === 'hidden') return false; } return true; };
  const out = []; let n = 0;
  const check = (kind, g, items) => { items = items.filter(vis); if (items.length < 2 || !vis(g)) return; n++;
    const w = items.map(i => i.getBoundingClientRect().width); const gr = g.getBoundingClientRect(), pr = g.parentElement.getBoundingClientRect();
    const over = items.some(i => i.scrollWidth > i.clientWidth + 1) || gr.right > pr.right + 1;
    if (Math.max(...w) - Math.min(...w) > 1 || over)
      out.push([kind, items.map(i => i.textContent.trim().replace(/\\s+/g, ' ').slice(0, 20)).join(' | '), w.map(x => x.toFixed(1)).join(' '), over]); };
  const kids = (g, c) => [...g.children].filter(x => x.classList.contains(c));
  document.querySelectorAll('.ip-seg').forEach(g => check('seg', g, kids(g, 'ip-seg__item')));
  document.querySelectorAll('.ip-tabs').forEach(g => check('tabs', g, kids(g, 'ip-tab')));
  document.querySelectorAll('.ip-tabbar').forEach(g => check('tabbar', g, kids(g, 'ip-tabbar__item')));
  new Set([...document.querySelectorAll('.ip-chip')].map(c => c.parentElement)).forEach(g => check('chips', g, kids(g, 'ip-chip')));
  return {out, n}; }"""


def targets(root, picks):
    proj = root / 'design/canvas/project'; sysd = root / 'design/system'
    out = [(proj / k, (proj / k).as_uri(), v['w'], v['h'])
           for k, v in json.loads((proj / 'canvas.json').read_text(encoding='utf-8'))['boards'].items() if (proj / k).exists()]
    for f in sorted(sysd.glob('components/*/preview.html')):
        wrap = pathlib.Path('/tmp/eq_scan') / (f.parent.name + '.html'); wrap.parent.mkdir(exist_ok=True)
        wrap.write_text(f'''<!doctype html><html lang="ro"><head><meta charset="utf-8">
<link rel="stylesheet" href="{(sysd / 'tokens.css').as_uri()}"><link rel="stylesheet" href="{(sysd / 'components/bundle.css').as_uri()}">
<link rel="stylesheet" href="{(sysd / 'components/additions.css').as_uri()}">
<style>body{{margin:0;padding:24px;background:var(--ground)}}</style></head><body>{f.read_text(encoding="utf-8")}</body></html>''', encoding='utf-8')
        out.append((f, wrap.as_uri(), 1440, 1024))
    if picks:
        out = [t for t in out if any(p in str(t[0]) or t[0].name.startswith(p) for p in picks)]
    return out


async def main(argv):
    root = pathlib.Path(argv[1]).resolve()
    ts = targets(root, argv[2:]); hits = []; groups = 0; sem = asyncio.Semaphore(4)
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        async def one(t):
            f, url, w, h = t
            async with sem:
                pg = await b.new_page(viewport={'width': w, 'height': min(h, 1200)})
                await pg.goto(url); await pg.wait_for_timeout(600)
                res = await pg.evaluate(SCAN); await pg.close()
            return f, res
        for f, res in await asyncio.gather(*(one(t) for t in ts)):
            groups += res['n']
            for kind, labels, widths, over in res['out']:
                hits.append(f'{f.relative_to(root)} · {kind} · {labels} · {widths}{" · overflow" if over else ""}')
    print('\n'.join(hits))
    print(f'eq_scan: {len(ts)} files, {groups} control groups, {len(hits)} mismatches')
    return 1 if hits else 0


if __name__ == '__main__':
    sys.exit(asyncio.run(main(sys.argv)))
