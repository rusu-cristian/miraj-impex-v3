"""python3 s13_render.py : render the S13 boards (phone 390×844 from /tmp/s13, desktop 1440×1024 from /tmp/s4s) to
design/shots/S13.png, S13-1 … S13-4.png (canvas order) + S13-4-1280.png, and report overflow, duplicate ids and broken ARIA references."""
import asyncio, pathlib
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
ORDER = [('S13-livrare-hrana', 390, 844, 's13'), ('S13-livrare-hrana--salvat', 390, 844, 's13'),
         ('S13-livrare-hrana--cantitate-lipsa', 390, 844, 's13'), ('S13-livrare-hrana--eroare-la-salvare', 390, 844, 's13'),
         ('S13-livrare-hrana--desktop', 1440, 1024, 's4s'), ('S13-livrare-hrana--desktop', 1280, 1024, 's4s')]  # last: S13-4-1280.png
CHECK = """() => {
 const ids = [...document.querySelectorAll('[id]')].map(e => e.id); const dup = ids.filter((x, i) => ids.indexOf(x) !== i);
 const refs = [...document.querySelectorAll('[aria-describedby],[aria-labelledby],[for],[aria-controls]')].flatMap(e =>
   ['aria-describedby','aria-labelledby','for','aria-controls'].flatMap(a => (e.getAttribute(a) || '').split(/\\s+/).filter(Boolean)))
   .filter(r => !document.getElementById(r));
 const main = document.querySelector('.ip-screen__main'); const sb = document.querySelector('.ip-savebar');
 return {dup, broken: refs, sw: document.documentElement.scrollWidth,
   mainBottom: main ? Math.round(main.getBoundingClientRect().bottom) : null, barTop: sb ? Math.round(sb.getBoundingClientRect().top) : null,
   body: (() => { const b = document.querySelector('.ip-sidepanel__body--form'); return b ? [b.scrollHeight, b.clientHeight] : null })()};
}"""
async def main():
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for k, (st, w, h, d) in enumerate(ORDER):
            pg = await b.new_page(viewport={'width': w, 'height': h})
            await pg.goto(f'file:///tmp/{d}/{st}.html'); await pg.wait_for_timeout(1000)
            print(st, await pg.evaluate(CHECK))
            await pg.screenshot(path=str(R / 'design/shots' / ('S13.png' if k == 0 else 'S13-4-1280.png' if w == 1280 else f'S13-{k}.png')))
            await pg.close()
        await b.close()
asyncio.run(main())
