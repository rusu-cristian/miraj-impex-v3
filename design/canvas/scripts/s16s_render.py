"""python3 s16s_render.py [--measure] [stems…] : render the S16S boards listed by s16s_build.py in /tmp/s16/states.json at their
board width and height to design/shots/S16-5.png, S16-6.png, … (board order), with s16_render.py's checks (ids, ARIA refs,
names, contrast, targets, tab order, sheet body overflow, nothing under the panel).
--measure: on the desktop-dupa-stergere board, at 1440 and 1280, print the focused Corectează's top/bottom and the page's
top/bottom; writes nothing. FX06: that board's page sits at its end by CSS (s16s_build), at every width, so no scroll is stored
(was "S16-corecteaza--desktop-dupa-stergere@scroll" in s16_heights.json, which left the key at y 1076 at 1280)."""
import asyncio, json, pathlib, sys, re
from playwright.async_api import async_playwright
R = pathlib.Path(__file__).resolve().parents[3]
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
                for vw in (w, 1280):
                    await pg.set_viewport_size({'width': vw, 'height': h}); await pg.wait_for_timeout(200)
                    print(vw, await pg.evaluate('[".ip-focus", "main.ip-shell__main"].map(s => { const r = document.querySelector(s).getBoundingClientRect(); return [s, Math.round(r.top), Math.round(r.bottom)]; })'))
                await pg.close(); break
            await pg.screenshot(path=str(R / 'design/shots' / f'S16-{i}.png'))
            await pg.close()
        await b.close()
asyncio.run(main())
