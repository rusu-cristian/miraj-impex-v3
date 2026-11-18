"""python3 s17_render.py [--measure] [--w 1280] [stems…] : render the S17 boards listed by s17_build.py in /tmp/s17/boards.json
at their board width and height (desktops from /tmp/s4s, the phone from /tmp/s2) to design/shots/S17.png, S17-1.png, …
(board order), with s2_render.py's checks (ids, ARIA refs, names, contrast, targets, overflow). --measure writes each
desktop board's natural height at 1440 (rounded up to 4, min 1024) into s17_heights.json (then rebuild with --canvas).
--w 1280 renders the desktops at 1280 as S17-<i>-1280.png (no heights written)."""
import asyncio, json, pathlib, sys, re
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
HJ = R / 'design/canvas/scripts/s17_heights.json'
src = (R / 'design/canvas/scripts/s2_render.py').read_text(encoding='utf-8')
CHECK = re.search(r'CHECK = r"""(.*?)"""', src, re.S).group(1)
A = sys.argv[1:]
W = int(A[A.index('--w') + 1]) if '--w' in A else None
ONLY = [a for i, a in enumerate(A) if not a.startswith('--') and not (i and A[i - 1] == '--w')]
async def main():
    boards = json.loads(pathlib.Path('/tmp/s17/boards.json').read_text())
    hs = json.loads(HJ.read_text()) if HJ.exists() else {}
    new = {}
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for i, bd in enumerate(boards):
            st, w, h = bd['stem'], bd['w'], bd['h']
            if ONLY and not any(st.endswith(o) for o in ONLY): continue
            if W and w != 1440: continue
            vw = W or w
            pg = await b.new_page(viewport={'width': vw, 'height': h})
            await pg.goto(f'file://{"/tmp/s2" if w == 390 else "/tmp/s4s"}/{st}.html'); await pg.wait_for_timeout(800)
            r = await pg.evaluate(CHECK); r.pop('order', None); print(st, vw, json.dumps(r, ensure_ascii=False))
            if w == 1440 and not W:
                nat = await pg.evaluate('document.querySelector(".ip-root").scrollHeight')
                new[st] = max(1024, int(-(-nat // 4) * 4))
            name = 'S17' + (f'-{i}' if i else '') + (f'-{W}' if W else '') + '.png'
            await pg.screenshot(path=str(R / 'design/shots' / name), full_page=False)
            await pg.close()
        await b.close()
    if '--measure' in A:
        HJ.write_text(json.dumps({**hs, **new}, indent=1)); print('heights', new)
asyncio.run(main())
