"""python3 s9_render.py : render the S9 boards (phone 390×844 from /tmp/s9, desktop 1440×1024 + 1280 from /tmp/s4s) to
design/shots/S9.png, S9-1 … S9-4.png (canvas order) + S9-4-1280.png, and report sideways scroll, duplicate ids, broken ARIA
references, unnamed form controls, the form end vs the SaveBar, and page elements under the open SidePanel."""
import asyncio, pathlib
from playwright.async_api import async_playwright
R = pathlib.Path('/home/claude/miraj-impex-v3')
ORDER = [('S9-transfer', 390, 844, 's9'), ('S9-transfer--salvat', 390, 844, 's9'), ('S9-transfer--refuzat', 390, 844, 's9'),
         ('S9-transfer--eroare-la-salvare', 390, 844, 's9'), ('S9-transfer--desktop', 1440, 1024, 's4s'), ('S9-transfer--desktop', 1280, 1024, 's4s')]
CHECK = """() => {
 const ids = [...document.querySelectorAll('[id]')].map(e => e.id); const dup = ids.filter((x, i) => ids.indexOf(x) !== i);
 const refs = [...document.querySelectorAll('[aria-describedby],[aria-labelledby],[for],[aria-controls],[aria-activedescendant]')].flatMap(e =>
   ['aria-describedby','aria-labelledby','for','aria-controls','aria-activedescendant'].flatMap(a => (e.getAttribute(a) || '').split(/\\s+/).filter(Boolean)))
   .filter(r => !document.getElementById(r));
 const unnamed = [...document.querySelectorAll('input:not([type=checkbox]),select,textarea,[role=combobox]')].filter(e => !(e.labels && e.labels.length) && !e.getAttribute('aria-labelledby') && !e.getAttribute('aria-label')).map(e => e.id || e.outerHTML.slice(0, 60));
 const main = document.querySelector('.s7-main'); const sb = document.querySelector('.ip-savebar');
 const pnl = document.querySelector('.ip-sidepanel'); let under = null;
 if (pnl) { const px = pnl.getBoundingClientRect().left; under = [...document.querySelectorAll('.s13-vp *')].filter(e => { const r = e.getBoundingClientRect(); return r.width && r.right > px + 1 && r.top < 1024; }).map(e => e.className.baseVal ?? e.className).slice(0, 4); }
 const fx = document.querySelector('#fx, #pfx'); const tall = [...document.querySelectorAll('.s7-main > *')].map(e => Math.round(e.getBoundingClientRect().bottom));
 return {dup, broken: refs, unnamed, sw: document.documentElement.scrollWidth, under,
   mainBottom: main ? Math.round(main.getBoundingClientRect().bottom) : null, parts: tall, barTop: sb ? Math.round(sb.getBoundingClientRect().top) : null,
   barH: sb ? Math.round(sb.getBoundingClientRect().height) : null,
   body: (() => { const b = document.querySelector('.s13-pbody'); return b ? [b.scrollHeight, b.clientHeight] : null })()};
}"""
async def main():
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for k, (st, w, h, d) in enumerate(ORDER):
            pg = await b.new_page(viewport={'width': w, 'height': h})
            await pg.goto(f'file:///tmp/{d}/{st}.html'); await pg.wait_for_timeout(1000)
            print(st, w, await pg.evaluate(CHECK))
            await pg.screenshot(path=str(R / 'design/shots' / ('S9.png' if k == 0 else 'S9-4-1280.png' if w == 1280 else f'S9-{k}.png')))
            await pg.close()
        await b.close()
asyncio.run(main())
