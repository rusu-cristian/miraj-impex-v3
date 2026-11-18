"""python3 s12_render.py <out_heights.json> <stem>... : render /tmp/s4s/<stem>.html at 1440, shot design/shots/S12[-n].png, record main height (min 1024)."""
import sys, json, asyncio, pathlib
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
ORDER = ['S12-preturi-hrana', 'S12-preturi-hrana--se-incarca', 'S12-preturi-hrana--in-lucru', 'S12-preturi-hrana--toate-livrarile-au-pret', 'S12-preturi-hrana--eroare']
async def main():
    hp = pathlib.Path(sys.argv[1]); hs = json.loads(hp.read_text()) if hp.exists() else {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for st in sys.argv[2:]:
            pg = await b.new_page(viewport={'width': 1440, 'height': 1024})
            await pg.goto(f'file:///tmp/s4s/{st}.html'); await pg.wait_for_timeout(1200)
            mh = await pg.evaluate("Math.ceil(document.querySelector('.ip-shell__main').lastElementChild.getBoundingClientRect().bottom + 24)")
            tw = await pg.evaluate("[...document.querySelectorAll('.ip-table')].map(t=>Math.round(t.getBoundingClientRect().width)).concat([document.documentElement.scrollWidth])")
            print(st, 'main h', mh, 'widths', tw)
            hs[st] = max(mh, 1024)
            k = ORDER.index(st); name = 'S12.png' if k == 0 else f'S12-{k}.png'
            await pg.screenshot(path=str(R / 'design/shots' / name), full_page=True)
            await pg.close()
        await b.close()
    hp.write_text(json.dumps(hs, indent=1))
asyncio.run(main())
