"""python3 s6_render.py [height] : render the S6 board (390 wide, from /tmp/s6) to design/shots/S6.png at the board height,
and report the natural content height (for s6_heights.json), sideways overflow, duplicate ids and broken ARIA/label refs."""
import asyncio, pathlib, sys
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
H = int(sys.argv[1]) if len(sys.argv) > 1 else 844
CHECK = """() => {
 const ids = [...document.querySelectorAll('[id]')].map(e => e.id); const dup = ids.filter((x, i) => ids.indexOf(x) !== i);
 const refs = [...document.querySelectorAll('[aria-describedby],[aria-labelledby],[for],[aria-controls]')].flatMap(e =>
   ['aria-describedby','aria-labelledby','for','aria-controls'].flatMap(a => (e.getAttribute(a) || '').split(/\\s+/).filter(Boolean)))
   .filter(r => !document.getElementById(r));
 const main = document.querySelector('.ip-screen__main'); const sb = document.querySelector('.ip-savebar');
 const over = [...document.querySelectorAll('.ip-screen__main *')].filter(e => e.getBoundingClientRect().right > 390.5).map(e => e.className).slice(0, 5);
 return {dup, broken: refs, sw: document.documentElement.scrollWidth, over,
   natural: (() => { const r = document.querySelector('.ip-root'); const mh = r.style.minHeight; r.style.minHeight = '0px';
     const h = Math.ceil(document.querySelector('.ip-screen').getBoundingClientRect().height); r.style.minHeight = mh; return h; })(),
   mainBottom: Math.round(main.getBoundingClientRect().bottom), barH: Math.round(sb.getBoundingClientRect().height),
   docH: document.documentElement.scrollHeight};
}"""
async def main():
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        pg = await b.new_page(viewport={'width': 390, 'height': H})
        await pg.goto('file:///tmp/s6/S6-zilnic.html'); await pg.wait_for_timeout(1000)
        print(await pg.evaluate(CHECK))
        await pg.screenshot(path=str(R / 'design/shots/S6.png'), full_page=True)
        await b.close()
asyncio.run(main())
