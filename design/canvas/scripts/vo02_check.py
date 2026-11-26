"""vo02_check.py · the VO02 key check (D199, spec §8 › Where). Renders every screen board (canvas w × h, headless Chrome as the
scans) and checks the voice key:
- every board except S1 holds exactly one button.ip-voicekey (S1 none);
- phone: a drawn key sits inside the frame, 16 from the right edge of its screen, never over the BottomTabBar, the SaveBar or
  the full-screen sheet's foot (bounding boxes; space-4 above the lowest bar unless it is on a BottomSheet scrim), and covers
  no data: no visible text, field, icon, lamp or table cell of the screen under it (elements under a scrim excepted). A fixed
  frame clipped mid-scroll (its form continues under the bar, the view not at its end) is listed apart: there the key floats
  over the content below as in the app (the 88 px clearance holds at a view's end and for a focused field, spec §8);
- desktop: the key is the last key of the first PageHeader, inside the frame;
- the frame holds its content: the root's scroll height ≤ the board's h (a board that grew is listed with its new height).
python3 -I vo02_check.py <repo_root> [--dir DIR] [STEM …]   (DIR: render copies instead of design/canvas/project)
Prints one line per problem; exit 1 if any."""
import sys, json, asyncio, pathlib
from playwright.async_api import async_playwright

JS = """() => {
  const keys = [...document.querySelectorAll('button.ip-voicekey')];
  const res = {n: keys.length, probs: [], sh: document.documentElement.scrollHeight};
  if (!keys.length) return res;
  const k = keys[0]; const kr = k.getBoundingClientRect(); res.key = [kr.x, kr.y, kr.width, kr.height];
  res.hidden = k.hidden; res.inert = !!k.closest('[inert]');
  const vw = document.documentElement.clientWidth;
  if (k.classList.contains('ip-voicekey--fab')) {
    res.kind = 'phone';
    if (k.hidden) return res;
    const scr = k.closest('.ip-screen, .ip-sidepanel--sheet, .ip-scrim--sheet');
    const sr = scr.getBoundingClientRect();
    res.right = Math.round(sr.right - kr.right);
    const onScrim = !!k.closest('.ip-scrim--sheet');
    const ov = (r) => !(r.right <= kr.left || r.left >= kr.right || r.bottom <= kr.top || r.top >= kr.bottom);
    for (const sel of ['.ip-screen__tabbar', '.ip-screen__savebar', '.ip-sidepanel__foot', '.ip-sheet']) {
      for (const b of document.querySelectorAll(sel)) { const r = b.getBoundingClientRect(); if (r.width && ov(r)) res.probs.push('over ' + sel); }
    }
    if (!onScrim) {
      const bar = [...document.querySelectorAll('.ip-screen__tabbar, .ip-screen__savebar, .ip-sidepanel--sheet > .ip-sidepanel__foot')]
        .filter(b => scr.contains(b)).map(b => b.getBoundingClientRect()).filter(r => r.width);
      if (bar.length) res.gap = Math.round(Math.min(...bar.map(r => r.top)) - kr.bottom);
    } else { const sh = k.parentElement.querySelector('.ip-sheet'); if (sh) res.gap = Math.round(sh.getBoundingClientRect().top - kr.bottom); }
    if (kr.bottom > window.innerHeight + 0.5 || kr.top < 0) res.probs.push('outside frame');
    // data under the key (not on a scrim: the scrim itself covers the page)
    if (!onScrim && !res.inert) {
      const vis = e => { for (let x = e; x; x = x.parentElement) { const s = getComputedStyle(x); if (s.display === 'none' || s.visibility === 'hidden' || x.classList.contains('ip-sr') || x.hidden) return false; } return true; };
      const clip = e => { let r = e.getBoundingClientRect(); for (let x = e.parentElement; x; x = x.parentElement) { const s = getComputedStyle(x);
          if (s.overflow !== 'visible' || s.overflowX !== 'visible' || s.overflowY !== 'visible') { const q = x.getBoundingClientRect();
            r = {left: Math.max(r.left, q.left), right: Math.min(r.right, q.right), top: Math.max(r.top, q.top), bottom: Math.min(r.bottom, q.bottom)};
            if (r.right <= r.left || r.bottom <= r.top) return null; } } return r; };
      const hits = new Set();
      const w = document.createTreeWalker(scr, NodeFilter.SHOW_TEXT); let t;
      while ((t = w.nextNode())) {
        if (!t.textContent.trim()) continue; const el = t.parentElement; if (k.contains(el) || !vis(el)) continue;
        if (el.closest('.ip-scrim')) continue;
        const rg = document.createRange(); rg.selectNodeContents(t);
        for (const q of rg.getClientRects()) { if (q.width < 1) continue; const cr = clip(el); if (!cr) continue;
          const r = {left: Math.max(q.left, cr.left), right: Math.min(q.right, cr.right), top: Math.max(q.top, cr.top), bottom: Math.min(q.bottom, cr.bottom)};
          if (r.right > r.left && r.bottom > r.top && ov(r)) hits.add('text ' + JSON.stringify(t.textContent.trim().slice(0, 30))); }
      }
      for (const el of scr.querySelectorAll('input, textarea, select, td, svg, .ip-input, .ip-lamp')) {
        if (k.contains(el) || el === k || !vis(el) || el.closest('.ip-scrim')) continue; const cr = clip(el); if (!cr) continue;
        if (ov(cr)) hits.add(el.tagName.toLowerCase() + '.' + (el.className.baseVal ?? el.className).split(' ')[0]);
      }
      const view = k.parentElement.querySelector(':scope > .ip-screen__view'); const main = view && view.querySelector(':scope > .ip-screen__main');
      const mid = view && main && !view.classList.contains('ip-screen__view--end') && main.getBoundingClientRect().bottom > view.getBoundingClientRect().bottom + 0.5;
      if (hits.size) (mid ? (res.mid = []) : res.probs).push('covers ' + [...hits].slice(0, 4).join(' | '));
    }
  } else {
    res.kind = 'desk';
    const h = document.querySelector('header.ip-pagehead');
    if (!h || !h.contains(k)) res.probs.push('not in the first PageHeader');
    else { const ks = [...h.querySelectorAll('button, a[href]')].filter(b => b.getBoundingClientRect().width); if (ks[ks.length - 1] !== k) res.probs.push('not the last key'); }
    if (kr.right > vw + 0.5 || kr.width < 1) res.probs.push('outside frame');
  }
  return res; }"""


# a SaveBar whose two-line summary wraps to three lines is 8 px taller than §78's 73 (FX05 assumes two at 390): the key keeps
# its 73-based place, 8 above the bar instead of 16 (reported, not over the bar)
KNOWN_GAP = {'S6-zilnic--abatere.dc.html': 8}
# taller than its h before VO02 (2352 in 2304, desktop: no key change in height)
KNOWN_TALL = {'S17-istoric-modificari--observator-marcaj-text.dc.html'}


async def main(argv):
    root = pathlib.Path(argv[1]).resolve(); rest = argv[2:]
    d = root / 'design/canvas/project'
    if '--dir' in rest:
        i = rest.index('--dir'); d = pathlib.Path(rest[i + 1]); del rest[i:i + 2]
    c = json.loads((root / 'design/canvas/project/canvas.json').read_text(encoding='utf-8'))
    ts = []
    for k in c['order']:
        if not k.startswith('S'): continue   # the screen boards (VO boards draw several keys: key strip, two phones)
        f = d / k
        if not f.exists(): f = d / k.replace('.dc.html', '.html')
        if not f.exists(): continue
        if rest and not any(k.startswith(s) for s in rest): continue
        ts.append((k, f, c['boards'][k]['w'], c['boards'][k]['h']))
    probs = []; mids = []; stats = {'phone': 0, 'desk': 0, 'hidden': 0, 'inert': 0, 'none': 0}; grown = []
    sem = asyncio.Semaphore(6)
    async with async_playwright() as p:
        b = await p.chromium.launch(executable_path='/usr/bin/google-chrome')
        async def one(t):
            k, f, w, h = t
            async with sem:
                pg = await b.new_page(viewport={'width': w, 'height': h})
                await pg.goto(f.as_uri()); await pg.wait_for_timeout(300)
                r = await pg.evaluate(JS); await pg.close()
            return k, w, h, r
        for k, w, h, r in await asyncio.gather(*(one(t) for t in ts)):
            s1 = k.startswith('S1-')
            if s1:
                if r['n']: probs.append(f'{k}: S1 holds {r["n"]} keys')
                stats['none'] += 1
            elif r['n'] != 1:
                probs.append(f'{k}: {r["n"]} keys')
            if r['n']:
                stats[r['kind']] += 1; stats['hidden'] += bool(r.get('hidden')); stats['inert'] += bool(r.get('inert'))
                if r['kind'] == 'phone' and not r.get('hidden'):
                    if r.get('right') != 16: r['probs'].append(f'right {r.get("right")}')
                    if r.get('gap') not in (None, 16) and not (k in KNOWN_GAP and r.get('gap') == KNOWN_GAP[k]):
                        r['probs'].append(f'gap to bar {r.get("gap")}')
                for q in r['probs']: probs.append(f'{k}: {q}')
                for q in r.get('mid') or []: mids.append(f'{k}: mid-scroll frame, the key floats over the form below it · {q}')
            if r['sh'] > h + 0.5 and k not in KNOWN_TALL: grown.append(f'{k}: content {r["sh"]} > h {h}')
    print('\n'.join(mids)); print('\n'.join(probs)); print('\n'.join(grown))
    print(f'vo02_check: {len(ts)} boards · keys phone {stats["phone"]} (hidden {stats["hidden"]}) desktop {stats["desk"]} · inert {stats["inert"]} · S1 {stats["none"]} · {len(mids)} mid-scroll frames (key over the form that continues below, as in the app) · {len(probs)} problems · {len(grown)} taller than h')
    return 1 if probs or grown else 0


if __name__ == '__main__':
    sys.exit(asyncio.run(main(sys.argv)))
