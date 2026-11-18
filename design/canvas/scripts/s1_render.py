"""python3 s1_render.py [stems…] : render the S1 boards listed by s1_build.py in /tmp/s1/boards.json at their board width
and height (from /tmp/s1/<stem>.html, after document.fonts.ready) to design/shots/S1.png, S1-1.png, … (board order), with
s2_render.py's checks (ids, ARIA refs, names, contrast, targets, overflow) plus the natural height (must fit 1024 / 844)."""
import asyncio, json, pathlib, sys, re
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
src = (R / 'design/canvas/scripts/s2_render.py').read_text(encoding='utf-8')
CHECK = re.search(r'CHECK = r"""(.*?)"""', src, re.S).group(1)
ONLY = sys.argv[1:]
async def main():
    boards = json.loads(pathlib.Path('/tmp/s1/boards.json').read_text())
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for i, bd in enumerate(boards):
            st, w, h = bd['stem'], bd['w'], bd['h']
            name = 'S1' + (f'-{i}' if i else '') + '.png'
            if ONLY and not any(st.endswith(o) or name == o for o in ONLY): continue
            pg = await b.new_page(viewport={'width': w, 'height': h})
            await pg.goto(f'file:///tmp/s1/{st}.html'); await pg.evaluate('document.fonts.ready.then(() => true)'); await pg.wait_for_timeout(800)
            r = await pg.evaluate(CHECK); r.pop('order', None)
            print(name, st, w, json.dumps(r, ensure_ascii=False), 'FITS' if r['natural'] <= h else 'TALLER')
            await pg.screenshot(path=str(R / 'design/shots' / name), full_page=False)
            await pg.close()
        await b.close()
asyncio.run(main())
