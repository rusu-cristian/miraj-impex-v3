"""python3 s21_render.py [--measure] [--w 1280] [stems…] : render the S21 boards listed by s21_build.py in /tmp/s21/boards.json
at their board width and height (from /tmp/s21/<stem>.html) to design/shots/S21.png, S21-1.png, … (board order), with
s2_render.py's checks (ids, ARIA refs, names, contrast, targets, overflow). --measure writes each board's natural height
(rounded up to 4; desktop min 1024, phone min 844) into s21_heights.json, then rebuild with --canvas: the list-only boards
and the phone by their page height; the SidePanel boards and --salvat keep 1024 (the page clipped beside the panel).
--w 1280 renders the desktops at 1280 as S21-<i>-1280.png (no heights written)."""
import asyncio, json, pathlib, sys, re
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
HJ = R / 'design/canvas/scripts/s21_heights.json'
src = (R / 'design/canvas/scripts/s2_render.py').read_text(encoding='utf-8')
CHECK = re.search(r'CHECK = r"""(.*?)"""', src, re.S).group(1)
A = sys.argv[1:]
W = int(A[A.index('--w') + 1]) if '--w' in A else None
ONLY = [a for i, a in enumerate(A) if not a.startswith('--') and not (i and A[i - 1] == '--w')]
NAT = '''() => { const r = document.querySelector('.ip-root'); const b = document.querySelector('.ip-sidepanel--sheet .ip-sidepanel__body');
  if (b) return r.clientHeight + (b.scrollHeight - b.clientHeight); return r.scrollHeight; }'''
PANEL = ('--breeder-2', '--starter', '--retragere-refuzata', '--tip-nou', '--nume-gol', '--nume-folosit', '--salvat',
         '--eroare-la-salvare', '--sterge', '--retras-stoc-0', '--reactivat')
async def main():
    boards = json.loads(pathlib.Path('/tmp/s21/boards.json').read_text())
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
            await pg.goto(f'file:///tmp/s21/{st}.html'); await pg.evaluate('document.fonts.ready.then(() => true)'); await pg.wait_for_timeout(800)
            r = await pg.evaluate(CHECK); r.pop('order', None); print(st, vw, json.dumps(r, ensure_ascii=False))
            if not W and not any(st.endswith(x) for x in PANEL):
                nat = await pg.evaluate(NAT)
                new[st] = max(1024 if w == 1440 else 844, int(-(-nat // 4) * 4))
            name = 'S21' + (f'-{i}' if i else '') + (f'-{W}' if W else '') + '.png'
            await pg.screenshot(path=str(R / 'design/shots' / name), full_page=False)
            await pg.close()
        await b.close()
    print('natural heights', new)
    if '--measure' in A:
        HJ.write_text(json.dumps({**hs, **new}, indent=1)); print('heights written')
asyncio.run(main())
