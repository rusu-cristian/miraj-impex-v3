"""python3 s16s_render.py [--measure] [stems…] : render the S16S boards listed by s16s_build.py in /tmp/s16/states.json at their
board width and height to design/shots/S16-5.png, S16-6.png, … (board order), with s16_render.py's checks (ids, ARIA refs,
names, contrast, targets, tab order, sheet body overflow, nothing under the panel).
--measure: on the desktop-dupa-stergere board (built with scroll 0) read the top of the Jurnal block, write scroll = top −
space-6 (or page height − board height, if smaller: the page never ends above the board's foot) into s16_heights.json ("S16-corecteaza--desktop-dupa-stergere@scroll"); rebuild with s16s_build.py <repo> after."""
import asyncio, json, pathlib, sys, re
from playwright.async_api import async_playwright
R = pathlib.Path('/home/claude/miraj-impex-v3')
HJ = R / 'design/canvas/scripts/s16_heights.json'
src = (R / 'design/canvas/scripts/s16_render.py').read_text(encoding='utf-8')
CHECK = re.search(r'CHECK = r"""(.*?)"""', src, re.S).group(1)  # the same checks as S16
MS = 'S16-corecteaza--desktop-dupa-stergere'
async def main():
    measure = '--measure' in sys.argv; only = [a for a in sys.argv[1:] if not a.startswith('--')]
    boards = json.loads(pathlib.Path('/tmp/s16/states.json').read_text())
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        for i, bd in enumerate(boards, 5):
            st, w, h = bd['stem'], bd['w'], bd['h']
            if only and not any(o == st.split('--')[-1] or o == st for o in only): continue
            if measure and st != MS: continue
            pg = await b.new_page(viewport={'width': w, 'height': h})
            await pg.goto(f'file:///tmp/s16/{st}.html'); await pg.wait_for_timeout(800)
            r = await pg.evaluate(CHECK); print(st, w, json.dumps(r, ensure_ascii=False))
            if measure:
                sh = await pg.evaluate('document.querySelector(".s13-vp").scrollHeight')  # page height (the board clips it)
                hs = json.loads(HJ.read_text()); hs[MS + '@scroll'] = min(r['jurnal'] - 24, sh - h); HJ.write_text(json.dumps(hs, indent=1) + '\n')
                print('scroll', hs[MS + '@scroll']); await pg.close(); break
            await pg.screenshot(path=str(R / 'design/shots' / f'S16-{i}.png'))
            await pg.close()
        await b.close()
asyncio.run(main())
