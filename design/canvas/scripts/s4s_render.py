"""python3 s4s_render.py [--heights] [--w 1280] [stem ...] : render the S4 Rezumat boards (/tmp/s4s/<stem>.html, plain
wrappers written by s4_build.py / s4s_build.py) at 1440 (or --w), shot design/shots/S4.png (default board) and
S4S-<n>.png (states, s4s_build order; -1280 suffix at another width), print the board height (main content + 24, min 1024).
--heights writes the state heights into s4s_heights.json and prints the default board's int for s4_build.py (README).
Run only to re-measure (README › Never run for a row); then rebuild the row with --canvas."""
import sys, json, asyncio, pathlib
from playwright.async_api import async_playwright
HERE = pathlib.Path(__file__).resolve().parent; R = HERE.parents[2]
B = 'S4-ferma-rezumat--'
ORDER = ['S4-ferma-rezumat'] + [B + s for s in ('se-incarca', 'fara-date-in-perioada', 'fara-ghid', 'fara-acces', 'ferma-goala',
                                                'ciclu-meniu', 'ciclu-ales', 'ciclu-nou')]
async def main():
    a = sys.argv[1:]; W = 1440
    if '--w' in a: i = a.index('--w'); W = int(a[i + 1]); del a[i:i + 2]
    save = '--heights' in a; stems = [x for x in a if not x.startswith('--')] or ORDER
    hs = {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for st in stems:
            pg = await b.new_page(viewport={'width': W, 'height': 1024})
            await pg.goto(f'file:///tmp/s4s/{st}.html'); await pg.wait_for_timeout(1200)
            mh = await pg.evaluate("Math.ceil(document.querySelector('.ip-shell__main').lastElementChild.getBoundingClientRect().bottom + 24)")
            clip = await pg.evaluate("""[...document.querySelectorAll('.ip-readout *,.ip-stat *,.ip-ev *,.ip-stile *,.ip-stotals *')].filter(e=>e.scrollWidth>e.clientWidth+1 && getComputedStyle(e).overflow!='visible').map(e=>e.className+':'+e.textContent.slice(0,40)).slice(0,8)""")
            wraps = await pg.evaluate("""[...document.querySelectorAll('.ip-stat__guide,.ip-stile__kv,.ip-stotals__sub,.ip-ev__meta')].filter(e=>e.getBoundingClientRect().height>24).map(e=>e.textContent.slice(0,50)).slice(0,10)""")
            print(st, W, 'board h', max(mh, 1024), 'clip', clip, 'wrapped', wraps)
            hs[st] = max(mh, 1024)
            k = ORDER.index(st); name = ('S4' if k == 0 else f'S4S-{k}') + ('' if W == 1440 else f'-{W}') + '.png'
            await pg.screenshot(path=str(R / 'design/shots' / name), full_page=True)
            await pg.close()
        await b.close()
    if save and W == 1440:
        hp = HERE / 's4s_heights.json'; old = json.loads(hp.read_text()) if hp.exists() else {}
        old.update({k: v for k, v in hs.items() if k != ORDER[0]})
        hp.write_text(json.dumps(old, indent=1, ensure_ascii=False) + '\n')
        if ORDER[0] in hs: print('s4_build.py height (README int):', hs[ORDER[0]])
asyncio.run(main())
