"""python3 s19_render.py [--measure] [--w 1280] [stems…] : render the S19 boards listed by s19_build.py in /tmp/s19/boards.json
at their board width and height (from /tmp/s19/<stem>.html) to design/shots/S19.png, S19-1.png, … (board order), with
s2_render.py's checks (ids, ARIA refs, names, contrast, targets, overflow). --measure writes each board's natural height
(rounded up to 4; desktop min 1024, phone min 844) into s19_heights.json, then rebuild with --canvas: the desktop list,
the no-farms board and the phone list by their page height; the SidePanel boards keep 1024 (the page is clipped beside the
panel, the panel body scrolls).
--w 1280 renders the desktops at 1280 as S19-<i>-1280.png (no heights written)."""
import asyncio, json, pathlib, sys, re
from playwright.async_api import async_playwright
R = pathlib.Path('/home/claude/miraj-impex-v3')
HJ = R / 'design/canvas/scripts/s19_heights.json'
src = (R / 'design/canvas/scripts/s2_render.py').read_text(encoding='utf-8')
CHECK = re.search(r'CHECK = r"""(.*?)"""', src, re.S).group(1)
A = sys.argv[1:]
W = int(A[A.index('--w') + 1]) if '--w' in A else None
ONLY = [a for i, a in enumerate(A) if not a.startswith('--') and not (i and A[i - 1] == '--w')]
NAT = '''() => { const r = document.querySelector('.ip-root'); const b = document.querySelector('.ip-sidepanel--sheet .ip-sidepanel__body');
  if (b) return r.clientHeight + (b.scrollHeight - b.clientHeight); return r.scrollHeight; }'''
async def main():
    boards = json.loads(pathlib.Path('/tmp/s19/boards.json').read_text())
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
            await pg.goto(f'file:///tmp/s19/{st}.html'); await pg.evaluate('document.fonts.ready.then(() => true)'); await pg.wait_for_timeout(800)
            r = await pg.evaluate(CHECK); r.pop('order', None); print(st, vw, json.dumps(r, ensure_ascii=False))
            if not W and not (w == 1440 and st != 'S19-ferme-sectiuni' and not st.endswith('--fara-ferme')):
                nat = await pg.evaluate(NAT)
                new[st] = max(1024 if w == 1440 else 844, int(-(-nat // 4) * 4))
            name = 'S19' + (f'-{i}' if i else '') + (f'-{W}' if W else '') + '.png'
            await pg.screenshot(path=str(R / 'design/shots' / name), full_page=False)
            await pg.close()
        await b.close()
    print('natural heights', new)
    if '--measure' in A:
        HJ.write_text(json.dumps({**hs, **new}, indent=1)); print('heights written')
asyncio.run(main())
