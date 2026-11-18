"""python3 s15s_render.py [--heights] [stems…] : render the S15S boards listed by s15s_build.py in /tmp/s15/states.json at their
board width and height to design/shots/S15-3.png, S15-4.png, … (in board order), with s2_render.py's checks (as
s15_render.py). --heights writes the desktop board's natural height at 1440 (rounded up to 4, min 1024) into
s15_heights.json (then rebuild with s15s_build.py <repo> --canvas)."""
import asyncio, json, pathlib, sys, re
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
HJ = R / 'design/canvas/scripts/s15_heights.json'
src = (R / 'design/canvas/scripts/s2_render.py').read_text(encoding='utf-8')
CHECK = re.search(r'CHECK = r"""(.*?)"""', src, re.S).group(1)  # the same checks as S2 / S15
ONLY = [a for a in sys.argv[1:] if not a.startswith('--')]
async def main():
    boards = json.loads(pathlib.Path('/tmp/s15/states.json').read_text())
    hs = json.loads(HJ.read_text()) if HJ.exists() else {}
    new = {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for i, bd in enumerate(boards, 3):
            st, w, h = bd['stem'], bd['w'], bd['h']
            if ONLY and not any(o in st for o in ONLY): continue
            src_dir = '/tmp/s15' if w == 390 else '/tmp/s4s'
            pg = await b.new_page(viewport={'width': w, 'height': h})
            await pg.goto(f'file://{src_dir}/{st}.html'); await pg.wait_for_timeout(800)
            r = await pg.evaluate(CHECK); print(st, json.dumps(r, ensure_ascii=False))
            if w == 1440:
                nat = await pg.evaluate('document.querySelector(".ip-root").scrollHeight')
                new[st] = max(1024, int(-(-nat // 4) * 4)); print('natural', nat)
            await pg.screenshot(path=str(R / 'design/shots' / f'S15-{i}.png'), full_page=False)
            await pg.close()
        await b.close()
    if '--heights' in sys.argv:
        HJ.write_text(json.dumps({**hs, **new}, indent=1)); print('heights', new)
asyncio.run(main())
