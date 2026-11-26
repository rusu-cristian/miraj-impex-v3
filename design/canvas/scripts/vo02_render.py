"""vo02_render.py · VO02 renders (headless Chrome, fallback fonts, as the sNN_render scripts).
python3 -I vo02_render.py <repo_root>                 shots of the 13 VO boards → design/shots/VO-1.png … VO-13.png (canvas w × h)
python3 -I vo02_render.py <repo_root> --measure       VO-5: the scroll that keeps the focused field 16 above the key; VO-9: the first popover's height; VO-13's and VO-8's heights (the S14 page's
                                                       natural height) → vo02_heights.json; rebuild vo02_build.py afterwards
python3 -I vo02_render.py <repo_root> --sample        the 6 sample shell boards (S2 phone, S4 desktop, S6 phone, S14 phone, S16
                                                       desktop, S22 desktop) → design/shots/VO-S*.png (S16 desktop at its top 1024)"""
import sys, json, asyncio, pathlib
from playwright.async_api import async_playwright

ROOT = pathlib.Path(sys.argv[1]).resolve()
PROJ = ROOT / 'design/canvas/project'; SHOTS = ROOT / 'design/shots'; HJ = pathlib.Path(__file__).with_name('vo02_heights.json')
C = json.loads((PROJ / 'canvas.json').read_text(encoding='utf-8'))['boards']
SAMPLE = ['S2-acasa', 'S4-ferma-rezumat', 'S6-zilnic', 'S14-rapoarte-productie--telefon', 'S16-corecteaza--desktop', 'S22-utilizatori']


async def main():
    mode = sys.argv[2] if len(sys.argv) > 2 else ''
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        async def page(stem, h=None):
            v = C[f'{stem}.dc.html']; pg = await b.new_page(viewport={'width': v['w'], 'height': h or v['h']})
            await pg.goto((PROJ / f'{stem}.dc.html').as_uri()); await pg.wait_for_timeout(500); return pg
        if mode == '--measure':
            hs = json.loads(HJ.read_text()) if HJ.exists() else {}
            pg = await page('VO-13-ferma-productie-toast', 844)
            hs['VO-13-ferma-productie-toast'] = await pg.evaluate("document.querySelector('.ip-root').style.minHeight='844px', document.documentElement.scrollHeight"); await pg.close()
            hs.pop('VO-13-ferma-productie-toast@scroll', None)
            pg = await page('VO-5-hrana-zilnica-precompletat')
            r = await pg.evaluate("""() => { const i = document.querySelector('.ip-input.is-focus').getBoundingClientRect(), k = document.querySelector('.ip-voicekey--float').getBoundingClientRect();
              const m = document.querySelector('.ip-screen__main'); return {d: i.bottom + 16 - k.top, cur: -(parseFloat(getComputedStyle(m).marginTop) || 0)}; }""")
            hs['VO-5-hrana-zilnica-precompletat@scroll'] = max(0, round(r['cur'] + r['d'])); await pg.close()
            pg = await page('VO-9-livrare-se-gandeste')
            hs['VO-9-livrare-se-gandeste@pop1'] = round(await pg.evaluate("document.querySelector('.ip-voiceanchor .ip-voicepanel--pop').getBoundingClientRect().height")); await pg.close()
            pg = await page('VO-8-productie-raspuns', 1024)
            hs['VO-8-productie-raspuns'] = max(1024, await pg.evaluate('document.documentElement.scrollHeight')); await pg.close()
            HJ.write_text(json.dumps(hs, indent=1) + '\n'); print(hs); return
        stems = SAMPLE if mode == '--sample' else [k[:-8] for k in C if k.startswith('VO-')]
        for s in stems:
            h = 1024 if s == 'S16-corecteaza--desktop' else None
            pg = await page(s, h)
            name = ('VO-S' + s.split('-')[0][1:] + ('-telefon' if C[f'{s}.dc.html']['w'] <= 600 or 'telefon' in s else '')) if mode == '--sample' else s.split('-')[0] + '-' + s.split('-')[1]
            await pg.screenshot(path=str(SHOTS / f'{name}.png'), full_page=False); await pg.close(); print('shot', name)
asyncio.run(main())
