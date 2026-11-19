"""python3 s4s_render.py [--heights] [--w 1280] [stem ...] : render the S4 Rezumat boards (/tmp/s4s/<stem>.html, plain
wrappers written by s4_build.py / s4s_build.py) at 1440 (or --w), shot design/shots/S4.png (default board) and
S4S-<n>.png (states, s4s_build order; -1280 suffix at another width), print the board height (main content + 24, min 1024).
--heights writes the state heights into s4s_heights.json and prints the default board's int for s4_build.py (README).
D167 (S4NG): each board renders at its canvas width (board_widths.json, else 1440) and prints the width its tables need
(1440 + the widest table's overflow, up to the next 8 px); --widths renders at 1440 and writes those needs into
board_widths.json (any /tmp/s4s stem, e.g. the S9/S10 desktop boards; shots only for the S4 Rezumat boards).
Run only to re-measure (README › Never run for a row); then rebuild the row with --canvas."""
import sys, json, asyncio, pathlib
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from playwright.async_api import async_playwright
HERE = pathlib.Path(__file__).resolve().parent; R = HERE.parents[2]
B = 'S4-ferma-rezumat--'
ORDER = ['S4-ferma-rezumat'] + [B + s for s in ('se-incarca', 'fara-date-in-perioada', 'fara-ghid', 'fara-acces', 'ferma-goala',
                                                'ciclu-meniu', 'ciclu-ales', 'ciclu-nou')]
async def main():
    a = sys.argv[1:]; W = 1440
    if '--w' in a: i = a.index('--w'); W = int(a[i + 1]); del a[i:i + 2]
    save = '--heights' in a; wid = '--widths' in a; stems = [x for x in a if not x.startswith('--')] or ORDER
    import canvas_rows as CR
    bw = dict(CR.WIDTHS)
    hs = {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for st in stems:
            vw = 1440 if wid else (CR.width(st, 1440) if W == 1440 else W)
            pg = await b.new_page(viewport={'width': vw, 'height': 1024})
            await pg.goto(f'file:///tmp/s4s/{st}.html'); await pg.wait_for_timeout(1200)
            if wid: await pg.evaluate("document.querySelectorAll('.s4s-wide').forEach(e => e.classList.remove('s4s-wide'))")  # measure as the 1440 page
            mh = await pg.evaluate("Math.ceil(document.querySelector('.ip-shell__main').lastElementChild.getBoundingClientRect().bottom + 24)")
            clip = await pg.evaluate("""[...document.querySelectorAll('.ip-readout *,.ip-stat *,.ip-ev *,.ip-stile *,.ip-stotals *')].filter(e=>e.scrollWidth>e.clientWidth+1 && getComputedStyle(e).overflow!='visible').map(e=>e.className+':'+e.textContent.slice(0,40)).slice(0,8)""")
            wraps = await pg.evaluate("""[...document.querySelectorAll('.ip-stat__guide,.ip-stile__kv,.ip-stotals__sub,.ip-ev__meta')].filter(e=>e.getBoundingClientRect().height>24).map(e=>e.textContent.slice(0,50)).slice(0,10)""")
            over = await pg.evaluate("Math.max(0, ...[...document.querySelectorAll('.ip-tablescroll')].map(r => r.scrollWidth - r.clientWidth))")
            need = vw + (-(-over // 8) * 8 if over else 0)
            print(st, vw, 'board h', max(mh, 1024), 'table overflow', over, 'needs w', need, 'clip', clip, 'wrapped', wraps)
            if wid:
                if need > 1440: bw[st] = need
                else: bw.pop(st, None)
            hs[st] = max(mh, 1024)
            if st in ORDER and not wid:
                k = ORDER.index(st); name = ('S4' if k == 0 else f'S4S-{k}') + ('' if W == 1440 else f'-{W}') + '.png'
                await pg.screenshot(path=str(R / 'design/shots' / name), full_page=True)
            await pg.close()
        await b.close()
    if wid:
        (HERE / 'board_widths.json').write_text(json.dumps(dict(sorted(bw.items())), indent=1, ensure_ascii=False) + '\n'); print('board_widths.json', bw)
    if save and W == 1440 and not wid:
        hp = HERE / 's4s_heights.json'; old = json.loads(hp.read_text()) if hp.exists() else {}
        old.update({k: v for k, v in hs.items() if k != ORDER[0]})
        hp.write_text(json.dumps(old, indent=1, ensure_ascii=False) + '\n')
        if ORDER[0] in hs: print('s4_build.py height (README int):', hs[ORDER[0]])
asyncio.run(main())
