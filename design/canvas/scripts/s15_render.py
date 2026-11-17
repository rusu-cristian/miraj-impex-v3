"""python3 s15_render.py [--heights] : render the S15 boards listed by s15_build.py in /tmp/s15/boards.json (390 wide,
at their board heights) to design/shots/S15.png, S15-1.png, …, with s2_render.py's checks (sideways scroll, duplicate
ids, broken ARIA references, unnamed controls, text under 4.5:1, touch targets under 44, overflowing row text, tab order).
--heights writes the natural height of the full-page board(s) (max 844 and the content, rounded up to 4) to
s15_heights.json (then rebuild with s15_build.py <repo> s15_heights.json)."""
import asyncio, json, pathlib, sys, re
from playwright.async_api import async_playwright
R = pathlib.Path('/home/claude/miraj-impex-v3')
HJ = R / 'design/canvas/scripts/s15_heights.json'
src = (R / 'design/canvas/scripts/s2_render.py').read_text(encoding='utf-8')
CHECK = re.search(r'CHECK = r"""(.*?)"""', src, re.S).group(1)  # the same checks as S2
async def main():
    boards = json.loads(pathlib.Path('/tmp/s15/boards.json').read_text())
    hs = json.loads(HJ.read_text()) if HJ.exists() else {}
    new = {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for i, bd in enumerate(boards):
            st, h = bd['stem'], bd['h']
            pg = await b.new_page(viewport={'width': 390, 'height': 844 if '--heights' in sys.argv else h})
            await pg.goto(f'file:///tmp/s15/{st}.html'); await pg.wait_for_timeout(800)
            r = await pg.evaluate(CHECK); print(st, json.dumps(r, ensure_ascii=False))
            if not bd['fixed']: new[st] = max(844, int(-(-r['natural'] // 4) * 4))
            await pg.screenshot(path=str(R / 'design/shots' / (f'S15-{i}.png' if i else 'S15.png')), full_page=not bd['fixed'])
            await pg.close()
        await b.close()
    if '--heights' in sys.argv:
        HJ.write_text(json.dumps({**hs, **new}, indent=1)); print('heights', new)
asyncio.run(main())
