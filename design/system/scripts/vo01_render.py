"""Render and check the VO01 voice previews at 1440: python3 -I vo01_render.py [<tag>] [--board]
Writes /tmp/vo01/<Comp>-<tag>.png (full page) and prints, per preview: content height (dsCard = + 48, D135), overflow, control
sizes (56 phone key, 32 desktop key, 48 touch chips / examples / close), broken id references and duplicate ids, text
contrast of the voice parts on their own surface, the key's icon on its fill per state, and (VoicePanel) the demo run
(key → listening with a live meter → auto-stop → thinking (aria-busy) → Toast; Esc cancels; reduced motion: popover and
sheet distances at rest). --board: the canvas board V1-voce.dc.html at 1440 (its height for BOARD_H)."""
import sys, asyncio, pathlib, re
from playwright.async_api import async_playwright
ROOT = pathlib.Path(__file__).resolve().parents[1]           # design/system
REPO = ROOT.parents[1]
OUT = pathlib.Path('/tmp/vo01'); OUT.mkdir(exist_ok=True)
HEAD = (f'<link rel="stylesheet" href="file://{ROOT}/tokens.css"><link rel="stylesheet" href="file://{ROOT}/components/bundle.css">'
        f'<link rel="stylesheet" href="file://{ROOT}/components/additions.css">')

CONTRAST = r"""() => {
  const lum = c => { const m = c.match(/[\d.]+/g).map(Number); const f = v => { v/=255; return v<=0.03928? v/12.92 : Math.pow((v+0.055)/1.055,2.4) };
    return 0.2126*f(m[0])+0.7152*f(m[1])+0.0722*f(m[2]); };
  const bg = e => { for (let n=e; n; n=n.parentElement){ const b=getComputedStyle(n).backgroundColor; const m=b.match(/[\d.]+/g);
    if (m && (m.length<4 || +m[3]>0.5)) return b; } return 'rgb(255,255,255)'; };
  const ratio = (a,b) => { const x=lum(a), y=lum(b); return ((Math.max(x,y)+0.05)/(Math.min(x,y)+0.05)); };
  const sel = '.ip-voicemark,.ip-voicepanel__hint,.ip-voicepanel__note,.ip-voicepanel__time,.ip-voicepanel__said,.ip-voicepanel__cmd,.ip-voiceex,'
    + '.ip-voicepanel__opened,.ip-voicepanel__q,.ip-voicechip,.ip-voicepanel__state .ip-lamp,.ip-voicekey,.ip-toast__txt,.ip-voicepanel__ex > p';
  const out = {};
  document.querySelectorAll(sel).forEach(e => { if (!e.getClientRects().length) return;
    const k = e.className.baseVal === undefined ? e.className.split(' ').filter(c=>c.startsWith('ip-')).slice(-1)[0] + (e.getAttribute('aria-pressed')==='true'?'[pressed]':'') + (e.getAttribute('aria-disabled')==='true'?'[disabled]':'') + (e.classList.contains('is-partial')?'[partial]':'') : '?';
    const r = ratio(getComputedStyle(e).color, bg(e)); out[k] = Math.min(out[k] ?? 99, r); });
  return Object.fromEntries(Object.entries(out).map(([k,v]) => [k, +v.toFixed(2)])); }"""
SIZES = r"""() => { const o = {}; document.querySelectorAll('.ip-voicekey,.ip-voicechip,.ip-voiceex,.ip-voicepanel__close,.ip-voicepanel__foot .ip-btn,.ip-voicepanel__keys .ip-btn,.ip-banner__key,.ip-voicepanel__said').forEach(e => {
  const r = e.getBoundingClientRect(); if (!r.width) return; const k = e.className.split(' ').filter(c => /ip-(voice|btn--touch|iconbtn--touch|banner__key|chip--touch)/.test(c)).join('.');
  (o[k] = o[k] || new Set()).add(Math.round(r.width) + 'x' + Math.round(r.height)); });
  return Object.fromEntries(Object.entries(o).map(([k, v]) => [k, [...v].slice(0, 6)])); }"""
OVER = r"""() => [...document.querySelectorAll('.ip-voicepanel,.ip-voicepanel__head,.ip-voicepanel__body > *,.ip-voicepanel__foot,.ip-banner,.ip-chips,.ip-label,.ip-toast,.vo-phone,.ip-savebar,.ip-entry__grid > *')]
  .filter(e => e.scrollWidth > e.clientWidth + 1 || (e.parentElement && e.getBoundingClientRect().right > e.parentElement.getBoundingClientRect().right + 1 && !e.closest('.vo-desk--page')))
  .map(e => e.className.split(' ')[0] + ':' + (e.textContent || '').slice(0, 30))"""
IDS = r"""() => { const bad = []; document.querySelectorAll('[aria-describedby],[aria-labelledby],[aria-controls]').forEach(e => {
  ['aria-describedby','aria-labelledby','aria-controls'].forEach(a => { const v = e.getAttribute(a); if (v) v.split(' ').forEach(i => { if (!document.getElementById(i)) bad.push(a + ':' + i); }); }); });
  document.querySelectorAll('label[for]').forEach(l => { if (!document.getElementById(l.htmlFor)) bad.push('for:' + l.htmlFor); });
  const all = [...document.querySelectorAll('[id]')].map(e => e.id); return { bad, dup: [...new Set(all.filter((x, i) => all.indexOf(x) !== i))] }; }"""
ARIA = r"""() => { const k = [...document.querySelectorAll('.ip-voicekey')];
  return { keys: k.length, named: k.every(e => e.getAttribute('aria-label') === 'Comandă vocală'), toggles: k.every(e => e.hasAttribute('aria-pressed')),
    shortcut: k.every(e => e.getAttribute('aria-keyshortcuts') === 'Control+Shift+Space'), noExpanded: k.every(e => !e.hasAttribute('aria-expanded')),
    busyArePressed: k.filter(e => e.getAttribute('aria-busy') === 'true').every(e => e.getAttribute('aria-pressed') === 'true'),
    meterHidden: [...document.querySelectorAll('.ip-voicemeter')].every(e => e.getAttribute('aria-hidden') === 'true'),
    statusPerPanel: [...document.querySelectorAll('.ip-voicepanel')].every(p => p.querySelectorAll('[role=status]').length === 1),
    radiogroupsNamed: [...document.querySelectorAll('.ip-voicechips[role=radiogroup]')].every(g => g.getAttribute('aria-labelledby')),
    marksHidden: [...document.querySelectorAll('.ip-voicemark')].every(m => m.getAttribute('aria-hidden') === 'true' && document.querySelector('[aria-describedby~="' + m.id + '"]')),
    failAlerts: [...document.querySelectorAll('.ip-voicepanel .ip-banner')].every(b => b.getAttribute('role') === 'alert'),
    hiddenKeyGone: [...document.querySelectorAll('.ip-voicekey[hidden]')].every(e => !e.getClientRects().length) }; }"""
GEOM = r"""() => [...document.querySelectorAll('.vo-phone')].map(ph => { const k = ph.querySelector('.ip-voicekey--float'); if (!k || !k.getClientRects().length) return null;
  const kr = k.getBoundingClientRect(), pr = ph.getBoundingClientRect(); const bar = ph.querySelector('.ip-screen__tabbar,.ip-screen__savebar');
  const t = ph.querySelector('.ip-toastwrap .ip-toast'); const br = bar && bar.getBoundingClientRect();
  return { right: Math.round(pr.right - kr.right), aboveBar: br ? Math.round(br.top - kr.bottom) : null, toastGap: t ? Math.round(kr.top - t.getBoundingClientRect().bottom) : null }; }).filter(Boolean)"""
SHEETKEY = r"""() => [...document.querySelectorAll('.ip-scrim--sheet > .ip-voicekey')].map(k => Math.round(k.parentElement.querySelector(':scope > .ip-sheet').getBoundingClientRect().top - k.getBoundingClientRect().bottom))
  .concat([...document.querySelectorAll('.ip-sidepanel--sheet > .ip-voicekey')].map(k => 'S16 foot ' + Math.round(k.parentElement.querySelector('.ip-sidepanel__foot').getBoundingClientRect().top - k.getBoundingClientRect().bottom)))
  .concat([...document.querySelectorAll('.ip-entry__grid:has(.ip-voicemark)')].map(g => 'pair inputs Δ ' + [...g.querySelectorAll(':scope > .ip-field > .ip-input')].map(e => Math.round(e.getBoundingClientRect().top)).join('/')))"""


async def check(b, name, html, tag, reduce=False):
    pf = OUT / f'{name}.html'
    pf.write_text(f'<!doctype html><html lang="ro"><head><meta charset="utf-8">{HEAD}<style>body{{margin:0;padding:24px;background:var(--ground)}}</style></head>'
                  f'<body>{html}</body></html>', encoding='utf-8')
    pg = await b.new_page(viewport={'width': 1440, 'height': 900}, reduced_motion='reduce' if reduce else 'no-preference')
    await pg.goto(f'file://{pf}'); await pg.wait_for_timeout(500)
    h = await pg.evaluate("Math.ceil(document.querySelector('body > .ip-root, body > div').getBoundingClientRect().height)")
    print(f'== {name}: content {h} → dsCard {h + 48}')
    print('  overflow:', (await pg.evaluate(OVER))[:8])
    print('  sizes:', await pg.evaluate(SIZES))
    print('  refs:', await pg.evaluate(IDS))
    print('  aria:', await pg.evaluate(ARIA))
    print('  contrast:', await pg.evaluate(CONTRAST))
    g = await pg.evaluate(GEOM)
    if g: print('  phone key (right / above bar / toast gap):', g)
    s = await pg.evaluate(SHEETKEY)
    if s: print('  key above sheet edge:', s)
    await pg.screenshot(path=str(OUT / f'{name}-{tag}.png'), full_page=True)
    return pg


async def demo(pg):
    """the VoicePanel step-through: the real key, the absolute popover, the meter from values, auto-stop, thinking, Toast"""
    k = pg.locator('#vo-demo-key'); pop = pg.locator('[data-vo=pop]')
    await k.click(); await pg.wait_for_timeout(450)
    lv = [await pg.evaluate("document.querySelector('[data-vo=pop] .ip-voicemeter').dataset.level")]
    await pg.wait_for_timeout(300); lv.append(await pg.evaluate("document.querySelector('[data-vo=pop] .ip-voicemeter').dataset.level"))
    st = await pop.get_attribute('data-state'); pr = await k.get_attribute('aria-pressed')
    kb, pb = await k.bounding_box(), await pop.bounding_box()
    print('  demo listen:', st, 'pressed', pr, 'levels', lv, 'focus on key', await pg.evaluate("document.activeElement.id"),
          'popover right − key right', round(pb['x'] + pb['width'] - kb['x'] - kb['width']), 'gap', round(pb['y'] - kb['y'] - kb['height']), 'w', round(pb['width']))
    await pg.screenshot(path=str(OUT / 'demo-listen.png'), clip={'x': 0, 'y': 0, 'width': 1440, 'height': 760})
    await pg.wait_for_timeout(2400)
    print('  demo after phrase:', await pg.evaluate("document.querySelector('[data-vo=state]').textContent"), 'busy', await k.get_attribute('aria-busy'))
    await pg.wait_for_timeout(2600)
    print('  demo done:', await pop.get_attribute('data-state'), 'toast', await pg.evaluate("document.querySelector('[data-vo=toast]').dataset.state"),
          'pressed', await k.get_attribute('aria-pressed'))
    await k.click(); await pg.wait_for_timeout(300); await pg.keyboard.press('Escape'); await pg.wait_for_timeout(300)
    print('  demo Esc:', await pop.get_attribute('data-state'), await pg.evaluate("document.querySelector('[data-vo=app]').textContent"),
          'focus', await pg.evaluate("document.activeElement.id"))
    await pg.keyboard.press('Control+Shift+Space'); await pg.wait_for_timeout(200)
    print('  shortcut:', await pop.get_attribute('data-state'), await k.get_attribute('aria-pressed'), 'focus stays', await pg.evaluate("document.activeElement.id"))
    await pg.keyboard.press('Escape')


async def main():
    tag = next((a for a in sys.argv[1:] if not a.startswith('--')), 'r')
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        if '--board' in sys.argv:
            pg = await b.new_page(viewport={'width': 1440, 'height': 900})
            await pg.goto(f'file://{REPO}/design/canvas/project/V1-voce.dc.html'.replace('V1', 'V1'))
            t = (REPO / 'design/canvas/project/V1-voce.dc.html').read_text(encoding='utf-8')
            body = t[t.index('<div class="ip-root"'):t.index('</x-dc>')]
            pf = OUT / 'board.html'
            pf.write_text(f'<!doctype html><html lang="ro"><head><meta charset="utf-8">{HEAD}<style>body{{margin:0}}</style></head><body>{body}</body></html>', encoding='utf-8')
            await pg.goto(f'file://{pf}'); await pg.wait_for_timeout(500)
            print('board content height:', await pg.evaluate("Math.ceil(document.querySelector('.ip-root').scrollHeight)"))
            await b.close(); return
        for comp in ('VoiceButton', 'VoicePanel', 'VoiceMark'):
            frag = (ROOT / 'components' / comp / 'preview.html').read_text(encoding='utf-8')
            pg = await check(b, comp, frag, tag)
            if comp == 'VoicePanel':
                await demo(pg)
            await pg.close()
        # reduced motion: the distances at rest, the fades kept
        pg = await b.new_page(viewport={'width': 1440, 'height': 900}, reduced_motion='reduce')
        pf = OUT / 'VoicePanel.html'; await pg.goto(f'file://{pf}'); await pg.wait_for_timeout(300)
        print('reduced motion:', await pg.evaluate("""(() => { const s = getComputedStyle(document.documentElement);
          const pop = document.querySelector('[data-vo=pop]'); const ps = getComputedStyle(pop);
          return { pop: s.getPropertyValue('--ip-pop-scale'), off: s.getPropertyValue('--ip-off'), press: s.getPropertyValue('--ip-press-scale'),
                   popTransition: ps.transitionProperty + ' ' + ps.transitionDuration,
                   meterTransition: getComputedStyle(document.querySelector('.ip-voicemeter > i')).transitionDuration,
                   keyLatch: getComputedStyle(document.querySelector('.ip-voicekey')).transitionProperty }; })()"""))
        await b.close()

asyncio.run(main())
