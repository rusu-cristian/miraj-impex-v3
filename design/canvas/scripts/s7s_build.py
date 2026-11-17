"""Build the S7 · Hrană zilnică state artboards (S7S) from the extracts, reusing s7_build.py.
python3 -I s7s_build.py <repo_root>
Imports s7_build (it rewrites the S7 default board at its height 3162, S7P) and reuses its helpers,
header, date field, CSS and data. Every state board is one 390×844 phone viewport: FormHeader pinned at the top,
SaveBar at the foot, the form scrolled to the part the state concerns (scroll offsets measured at 390; all line heights
are fixed px tokens, so they do not depend on the font). Writes design/canvas/project/S7-hrana-zilnica--<state>.dc.html
and /tmp/s7s/<stem>.html (plain wrappers for renders)."""
import sys, pathlib, importlib, datetime as dt

ROOT = pathlib.Path(sys.argv[1])
sys.argv = ['s7_build.py', str(ROOT), '3162']
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
B = importlib.import_module('s7_build')
import sstate_css as SC  # viewport / Toast-undo / Banner-action / skeleton CSS, shared with s6s_build
n, dd, icon, sexv, pair, lamp = B.n, B.dd, B.icon, B.sexv, B.pair, B.lamp
NB, MINUS, TODAY, YDAY, WEEK, START = B.NB, B.MINUS, B.TODAY, B.YDAY, B.WEEK, B.START
GUIDE, FTN, FTS, HB, ROWS = B.GUIDE, B.FTN, B.FTS, B.HB_TYPES, B.ROWS
B.IC.update({
 'alert': '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>',   # Banner fault preview
 'check': '<path d="M20 6 9 17l-5-5"/>',                                                    # Toast preview
})
VH = 844

# ---------- one AllocationRow in a given state ----------
# mode: ok / warn (computed from the deviation) · sugg (yesterday, unconfirmed) · empty (no yesterday) · error (out of 1–300) · ro (saved)
def row(r, mode, g=None, uid=''):
    sx, sec, alive = r['sex'], r['section'], r['alive']
    c = 'hen' if sx == 'F' else 'rooster'; sign = '♀' if sx == 'F' else '♂'
    word = 'găini' if sx == 'F' else 'cocoși'
    gid = f'{uid}{sec.replace(".", "-")}-{sx}'
    guide = r['guide_g']; assert guide == GUIDE[sx]
    if g is None: g = r['today_g'] if mode not in ('sugg',) else r['yesterday_g']
    sr = f'<span class="ip-sr">, {sec} {word}</span>'
    opts = ''.join(f'<option{" selected" if t == r["feed_type_id"] else ""}>{FTN[t]}</option>' for t in HB if FTS[t] == sx)
    if mode == 'ro':
        sel = f'<label class="ip-field" for="t-{gid}"><span class="ip-label">Tip hrană{sr}</span><span class="ip-input ip-input--touch is-readonly"><input id="t-{gid}" type="text" value="{FTN[r["feed_type_id"]]}" readonly/></span></label>'
    else:
        sel = f'<label class="ip-field"><span class="ip-label">Tip hrană{sr}</span><span class="ip-input ip-input--touch"><select>{opts}</select>{icon("chev", 16)}</span></label>'
    kg = round(g * alive / 1000, 1) if mode in ('ok', 'warn', 'sugg', 'ro') else None
    dev = (g - guide) / guide * 100 if kg is not None else 0
    if mode in ('ok', 'ro', 'warn'): mode_eff = 'warn' if abs(dev) > 10 and mode != 'ro' else mode
    else: mode_eff = mode
    devtxt = f' · <span class="ip-num">{"+" if dev > 0 else MINUS}{abs(dev):.0f}%</span>' if abs(dev) > 10 else ''
    st = lamp('warn', 'abatere') if abs(dev) > 10 else lamp('ok', 'în normă')
    msg = ''
    if mode_eff == 'sugg':
        inp = f'<span class="ip-input ip-input--touch ip-num is-suggested"><input id="g-{gid}" type="text" inputmode="decimal" value="{g}" aria-describedby="s-{gid}"/><span class="suffix">ieri</span></span>'
        right = f'<div class="ip-alloc__kg is-suggested">{n(kg, 1)}{NB}kg<small>ghid {guide}{NB}g</small><small id="s-{gid}">de confirmat</small></div>'
    elif mode_eff == 'empty':
        inp = f'<span class="ip-input ip-input--touch ip-num"><input id="g-{gid}" type="text" inputmode="decimal" value="" placeholder="{guide}" aria-describedby="s-{gid}"/><span class="suffix">g</span></span>'
        right = f'<div class="ip-alloc__kg is-suggested">–<small>ghid {guide}{NB}g</small><small id="s-{gid}">de completat</small></div>'
    elif mode_eff == 'error':
        inp = f'<span class="ip-input ip-input--touch ip-num is-error"><input id="g-{gid}" type="text" inputmode="decimal" value="{g}" aria-invalid="true" aria-describedby="e-{gid}"/><span class="suffix">g</span></span>'
        right = f'<div class="ip-alloc__kg">–<small>ghid {guide}{NB}g</small>{lamp("fault", "refuzat")}</div>'
        msg = f'<p class="ip-err ip-alloc__msg" id="e-{gid}">{icon("alert", 14)}<span>{sec} <span class="ip-sex ip-sex--{c}">{sign}</span>: <span class="ip-num">{g}{NB}g</span> e peste limită. Scrie o valoare între <span class="ip-num">1</span> și <span class="ip-num">300</span>{NB}g.</span></p>'
    elif mode_eff == 'warn':
        inp = f'<span class="ip-input ip-input--touch ip-num is-warn"><input id="g-{gid}" type="text" inputmode="decimal" value="{g}" aria-describedby="w-{gid}"/><span class="suffix">g</span></span>'
        right = f'<div class="ip-alloc__kg">{n(kg, 1)}{NB}kg<small>ghid {guide}{NB}g{devtxt}</small>{st}</div>'
        msg = f'<p class="ip-warnline ip-alloc__msg" id="w-{gid}">{icon("warn", 14)}<span>{sec} <span class="ip-sex ip-sex--{c}">{sign}</span>: <span class="ip-num">{g}{NB}g</span>, cu <span class="ip-num">{abs(dev):.0f}%</span> {"peste" if dev > 0 else "sub"} ghidul de <span class="ip-num">{guide}{NB}g</span>. Verifică gramajul; se poate salva.</span></p>'
    elif mode_eff == 'ro':
        inp = f'<span class="ip-input ip-input--touch ip-num is-readonly"><input id="g-{gid}" type="text" value="{g}" readonly/><span class="suffix">g</span></span>'
        right = f'<div class="ip-alloc__kg">{n(kg, 1)}{NB}kg<small>ghid {guide}{NB}g{devtxt}</small>{st}</div>'
    else:
        inp = f'<span class="ip-input ip-input--touch ip-num"><input id="g-{gid}" type="text" inputmode="decimal" value="{g}"/><span class="suffix">g</span></span>'
        right = f'<div class="ip-alloc__kg">{n(kg, 1)}{NB}kg<small>ghid {guide}{NB}g</small>{st}</div>'
    ginp = f'<label class="ip-field" for="g-{gid}"><span class="ip-label">g/pasăre{sr}</span>{inp}</label>'
    html = f'''<div class="ip-alloc ip-alloc--phone" role="group" aria-label="{sec}, {word}, {n(alive)} în viață" id="r-{gid}">
<div class="ip-alloc__who"><span class="ip-sex ip-sex--{c} ip-t-heading" aria-hidden="true">{sign}</span><div><div class="ip-t-body-strong">{sec}</div><div class="ip-note"><span class="ip-num">{n(alive)}</span> {word}</div></div></div>
{right}
<div class="ip-alloc__ctl">{sel}{ginp}</div>
{msg}
</div>'''
    confirmed = mode_eff in ('ok', 'warn', 'ro')
    return dict(html=html, conf=confirmed, kg=kg if confirmed else None, t=r['feed_type_id'], sx=sx, sec=sec)

def halls(built):
    hs = {}
    for r, b in zip(ROWS, built): hs.setdefault(r['section'].split('.')[0], []).append((r, b['html']))
    return '<div class="ip-halls">' + ''.join(
        f'''<section class="ip-hall" aria-labelledby="h-{h}"><header class="ip-hall__head"><span class="ip-cap">Hală</span><h2 class="ip-hall__code" id="h-{h}">{h}</h2><span class="ip-note">secțiuni <span class="ip-num">{" · ".join(dict.fromkeys(r["section"] for r, _ in rs))}</span></span>{pair(n(sum(r["alive"] for r, _ in rs if r["sex"] == "F")), n(sum(r["alive"] for r, _ in rs if r["sex"] == "M")), sr="în viață ")}</header>
<div class="ip-hall__sections">{"".join(b for _, b in rs)}</div></section>''' for h, rs in hs.items()) + '</div>'

# ---------- FeedTotals (phone) for a set of rows; start override = D39 state example ----------
fbt = {(f['date'], f['feed_type_id']): f for f in B.S['feed_by_type_daily']}
STK = {s['feed_type_id']: s for s in B.S['stock_today']}
def ftot(built, start_override=None):
    out = []
    for t in HB:
        sx = FTS[t]; rs = [b for b in built if b['t'] == t]
        conf = [b for b in rs if b['conf']]; miss = [b for b in rs if not b['conf']]
        start = (start_override or {}).get(t, fbt[(YDAY.isoformat(), t)]['stock_end_kg'])
        azi = round(sum(b['kg'] for b in conf), 1) if conf else None
        after = start - (azi or 0)
        avgk = STK[t]['avg_daily_kg_last3']
        days = max(after, 0) / avgk
        below0 = after < 0; ok = days >= 5
        sign = '♀' if sx == 'F' else '♂'; c = 'hen' if sx == 'F' else 'rooster'
        if azi is None: kgcell = '<span class="ip-ftot__val">–</span><span class="ip-ftot__sub">nimic confirmat</span>'
        else:
            left = f'fără {sign} ' + ', '.join(b['sec'] for b in miss) if miss else 'toate confirmate'
            kgcell = f'<span class="ip-ftot__val">{n(azi, 1)}<small>kg</small></span><span class="ip-ftot__sub">{left}</span>'
        stsub = ('<span class="ip-ftot__sub">stoc sub 0</span>' if below0 else '') + (f'<span class="ip-ftot__sub">din <span class="ip-num">{n(round(start))}</span> kg</span>' if azi is not None else '<span class="ip-ftot__sub">stoc acum</span>')
        daysub = 'se termină azi' if below0 else f'până pe <span class="ip-num">{dd(TODAY + dt.timedelta(days=int(days)))}</span>'
        banner = f'<div class="ip-banner ip-banner--warn ip-ftot__msg" role="status">{icon("warn", 16)}<div>Stocul de {FTN[t]} la Hîrbovăț ar ajunge la <span class="ip-num">{n(round(after))}</span> kg. Verifică livrările.</div></div>' if below0 else ''
        out.append(dict(t=t, days=days, ok=ok, below0=below0, after=after, html=f'''<div class="ip-ftot__row" role="group" aria-label="{FTN[t]}, pentru {"găini" if sx == "F" else "cocoși"}"><div class="ip-ftot__type"><span class="ip-tile ip-rec ip-rec--feed" aria-hidden="true">{icon("feed")}</span><div><span class="ip-ftot__name">{FTN[t]} <span class="ip-sex ip-sex--{c}" aria-hidden="true">{sign}</span></span><span class="ip-ftot__sub">{len(conf)} din {len(rs)} rânduri confirmate</span></div></div><div class="ip-ftot__cell ip-ftot__cell--kg"><span class="ip-ftot__key">Azi</span>{kgcell}</div><div class="ip-ftot__cell ip-ftot__cell--stock"><span class="ip-ftot__key">Stoc după azi</span><span class="ip-ftot__val">{n(round(after))}<small>kg</small></span>{stsub}</div><div class="ip-ftot__cell ip-ftot__cell--days"><span class="ip-ftot__key">Zile rămase</span><span class="ip-ftot__val">{n(days, 1)}<small>zile</small></span><span class="ip-ftot__sub">{daysub}</span></div><div class="ip-ftot__status">{lamp("ok", "în normă") if ok else lamp("warn", "abatere")}</div>{banner}</div>'''))
    w0, w1 = STK[HB[0]]['avg_window']
    avg = ', '.join(f'{FTN[t]}: <span class="ip-num">{n(STK[t]["avg_daily_kg_last3"])}</span> kg/zi' for t in HB)
    html = f'''<section class="ip-ftot" aria-labelledby="ft-h"><div class="ip-ftot__head"><h2 class="ip-cap s7-m0" id="ft-h">Total pe tip de hrană</h2><span class="ip-note"><span class="ip-num">{dd(TODAY)}</span> · doar rânduri confirmate</span></div>
{"".join(f["html"] for f in out)}
<p class="ip-note ip-ftot__foot">Zile rămase = stocul după azi ÷ consumul mediu <span class="ip-num">{dt.date.fromisoformat(w0).strftime("%d")}–{dd(dt.date.fromisoformat(w1))}</span> ({avg}). Abatere sub 5 zile.</p></section>'''
    return html, out

def stocklamp(fts):
    w = next((f for f in fts if f['below0']), None) or min(fts, key=lambda f: f['days'])
    if w['below0']: return lamp('warn', 'abatere') + f'<span>{FTN[w["t"]]} · stoc sub 0</span>'
    if not w['ok']: return lamp('warn', 'abatere') + f'<span>{FTN[w["t"]]} <span class="ip-num">{n(w["days"], 1)}</span> zile</span>'
    return lamp('ok', 'în normă') + '<span>stoc</span>'

def savebar(built, line2, key, lead=''):  # the one SaveBar shell (s7_build.savebar / sum_rows)
    nconf = sum(1 for b in built if b['conf']); tot = round(sum(b['kg'] for b in built if b['conf']), 1)
    return B.savebar(B.sum_rows(nconf, len(built), tot if nconf else None, lead), line2, key), nconf, tot

SAVE_KEY = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Salvează</button>'
SAVE_OFF = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="sb-sum">Salvează</button>'
GLINE = f'<p class="ip-note s7-m0">ghid <span class="ip-num">SAPT {WEEK}</span>: {pair(f"{GUIDE['F']}{NB}g", f"{GUIDE['M']}{NB}g")}</p>'
def copy(disabled_reason=None):
    if disabled_reason:
        return f'''<div class="s7-copy">
<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch" disabled aria-describedby="cp-why">{icon("copy", 16)}Copiază ieri</button>
<p class="ip-note s7-m0" id="cp-why">{disabled_reason}</p>
{GLINE}
</div>'''
    return B.COPY

def default_modes():  # the S7 default: confirmed rows as entered, the two missing rows as yesterday's suggestion
    return [('ok', None) if r['today_g'] is not None else ('sugg', None) for r in ROWS]

def idx(sec, sx): return next(i for i, r in enumerate(ROWS) if r['section'] == sec and r['sex'] == sx)

def screen(main_parts, save, view_cls='', scroll=0, overlay=''):
    style = f' style="margin-top:calc(-1 * {scroll}px)"' if scroll else ''
    return f'''<div class="s7-screen s7s-vp">
{B.HEAD}
<div class="s7s-view{(" " + view_cls) if view_cls else ""}">
<main class="s7-main"{style}>
{"".join(main_parts)}
</main>
{overlay}
</div>
{save}
</div>'''

boards = []

# 1 · fără zi anterioară — no 15.11 grams: empty fields, guide as placeholder, Copiază ieri disabled with its reason
b1 = [row(r, 'empty') for r in ROWS]
ft1, f1 = ftot(b1)
sv1, nc, _ = savebar(b1, stocklamp(f1), SAVE_OFF)
assert nc == 0 and [n(f['days'], 1) for f in f1] == ['7,9', '4,8']
reason = f'Nimic de copiat: pe <span class="ip-num">{dd(YDAY)}</span> nu sunt grame introduse. Scrie gramele pentru fiecare rând.'
boards.append(('fara-zi-anterioara', 'fără zi anterioară', screen([B.DATE, copy(reason), halls(b1), ft1], sv1)))

# 2 · abatere — H1.2 ♂ 172 g typed (guide 155, +11%): warn field + warnline, Salvează enabled
m = default_modes(); b2 = [row(r, *mm) for r, mm in zip(ROWS, m)]
assert ROWS[idx('H1.2', 'M')]['today_g'] == 172
ft2, f2 = ftot(b2)
sv2, nc, tot = savebar(b2, stocklamp(f2), SAVE_KEY)
assert (nc, n(tot, 1)) == (10, f'3{NB}691,3')
SCROLL2 = 329  # measured at 390 (S7P): H1.2 ♂ row (with its warn line) ends 16 px above the SaveBar; H1.2 ♀ and part of H1.1 ♂ above it
boards.append(('abatere', 'abatere', screen([B.DATE, B.COPY, halls(b2), ft2], sv2, scroll=SCROLL2)))

# 3 · valoare în afara limitei — H1.1 ♀ 350 g: error (D57 copy), refuzat, Salvează disabled until fixed
m = default_modes(); m[idx('H1.1', 'F')] = ('error', 350)
b3 = [row(r, *mm) for r, mm in zip(ROWS, m)]
ft3, f3 = ftot(b3)
sv3, nc, tot = savebar(b3, f'{lamp("fault", "refuzat")}<span>H1.1 <span class="ip-sex ip-sex--hen">♀</span> <span class="ip-num">350{NB}g</span> ·</span><span>scrie <span class="ip-num">1–300</span>{NB}g</span>', SAVE_OFF)
assert (nc, n(tot, 1)) == (9, f'3{NB}015,1')
boards.append(('valoare-in-afara-limitei', 'valoare în afara limitei', screen([B.DATE, B.COPY, halls(b3), ft3], sv3)))

# 4 · stoc sub 0 — spec C2 sentence; D39 state example (start-of-day Breeder 2 stock 2 093 kg so that today's 3 332,5 kg gives −1 240)
m = default_modes(); b4 = [row(r, *mm) for r, mm in zip(ROWS, m)]
ft4, f4 = ftot(b4, start_override={'ft-breeder2': 2093})
assert n(round(f4[0]['after'])) == f'{MINUS}1{NB}240' and f4[0]['below0']
sv4, _, _ = savebar(b4, stocklamp(f4), SAVE_KEY)
boards.append(('stoc-sub-0', 'stoc sub 0 · exemplu de stare (D39)', screen([B.DATE, B.COPY, halls(b4), ft4], sv4, view_cls='s7s-view--end')))

# 5 · salvat — all 12 rows confirmed (H3.2 ♂ 157, H4.2 ♀ 156 from yesterday) and saved: read-only, Toast with Anulează, Editează
m = [('ro', r['today_g'] if r['today_g'] is not None else r['yesterday_g']) for r in ROWS]
b5 = [row(r, *mm) for r, mm in zip(ROWS, m)]
ft5, f5 = ftot(b5)
tot5 = round(sum(b['kg'] for b in b5), 1); br2 = round(f5[0]['after'])
assert n(tot5, 1) == f'4{NB}398,9' and br2 == 27611 and n(f5[1]['days'], 1) == '3,8'
sv5, _, _ = savebar(b5, '<span>Dr. Ceban ·</span><span>poți modifica până la <span class="ip-num">23:59</span></span>', '<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Editează</button>', lead='salvat')
toast = f'''<div class="s7s-toastwrap"><div class="ip-toast s7s-toast" role="status">{icon("check", 16)}<span class="s7s-toast__txt">Salvat · <span class="ip-num">12</span> rânduri · <span class="ip-num">{n(tot5, 1)}</span> kg · stoc Breeder 2: <span class="ip-num">{n(br2)}</span> kg</span><button type="button" class="ip-link s7s-undo">Anulează</button></div></div>'''
boards.append(('salvat', 'salvat', screen([B.DATE, f'<div class="s7-copy">{GLINE}</div>', halls(b5), ft5], sv5, overlay=toast)))

# 6 · eroare la salvare — connectivity: fault Banner + Reîncearcă, typed values kept (draft in the browser)
m = default_modes(); b6 = [row(r, *mm) for r, mm in zip(ROWS, m)]
ft6, f6 = ftot(b6)
sv6, nc6, tot6 = savebar(b6, stocklamp(f6), SAVE_KEY)
assert (nc6, n(tot6, 1)) == (10, f'3{NB}691,3')
ban = f'''<div class="ip-banner ip-banner--fault s7s-banner" role="alert">{icon("alert", 16)}<div class="s7s-banner__body"><p class="s7-m0">Nu s-au salvat cele <span class="ip-num">{nc6}</span> rânduri din <span class="ip-num">{dd(TODAY)}</span> (<span class="ip-num">{n(tot6, 1)}</span> kg). Verifică conexiunea și apasă Reîncearcă.</p><p class="s7-m0">Ce ai scris rămâne păstrat în browser până la salvare.</p><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Reîncearcă</button></div></div>'''
boards.append(('eroare-la-salvare', 'eroare la salvare', screen([ban, B.DATE, B.COPY, halls(b6), ft6], sv6)))

# 7 · se încarcă — static Skeleton rows (no shimmer); header and date are known from the route
def skel(w, h=None): return f'<div class="ip-skel" style="width:{w}{";height:" + h if h else ""}"></div>'
# Skeleton geometry = the loaded geometry, so nothing shifts when rows arrive (S7P): row line 1 = code 20 + alive 16 |
# kg 20 + guide 16 + lamp 16 (gaps 4); line 2 = label 16 + field 48; hall head = heading line 22; Copiază ieri = key 48 +
# guide line 16; SaveBar summary = 2 lines of 16 (gap 4).
SK16, SK20 = 'var(--ip-lh-small)', 'var(--ip-lh-alloc)'
skrow = f'''<div class="ip-alloc ip-alloc--phone s7s-skrow" aria-hidden="true"><div class="ip-alloc__who s7s-skcol">{skel("40%", SK20)}{skel("60%", SK16)}</div><div class="ip-alloc__kg s7s-skcol s7s-skcol--end">{skel("calc(var(--space-12) * 1.5)", SK20)}{skel("var(--space-12)", SK16)}{skel("calc(var(--space-12) + var(--space-4))", SK16)}</div><div class="ip-alloc__ctl"><div class="s7s-skcol">{skel("40%", SK16)}{skel("100%", "var(--control-h-touch)")}</div><div class="s7s-skcol">{skel("60%", SK16)}{skel("100%", "var(--control-h-touch)")}</div></div></div>'''
skhall = f'''<section class="ip-hall" aria-hidden="true"><div class="ip-hall__head">{skel("30%", "var(--ip-lh-heading)")}</div><div class="ip-hall__sections">{skrow * 4}</div></section>'''
sk = f'''<p class="ip-sr" role="status">Se încarcă rândurile pentru <span class="ip-num">{dd(TODAY)}</span>…</p>
<div class="s7s-skblock" aria-busy="true">
<div class="s7-copy" aria-hidden="true">{skel("45%", "var(--control-h-touch)")}{skel("55%", SK16)}</div>
<div class="ip-halls">{skhall * 2}</div>
</div>'''
sv7 = f'''<div class="ip-savebar s7-savebar"><span class="ip-savebar__sum s7s-skcol" aria-hidden="true">{skel("70%", SK16)}{skel("45%", SK16)}</span><button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled>Salvează</button></div>'''
boards.append(('se-incarca', 'se încarcă', screen([B.DATE, sk], sv7)))

CSS = B.CSS + SC.VIEW + SC.TOAST + SC.BANNER
CSS_EXTRA = {'se-incarca': SC.SKEL}  # skeleton rules only on the board that uses them

STEM0, TITLE0 = 'S7-hrana-zilnica', 'S7 · Hrană zilnică'
SYS = ROOT / 'design/system'
pathlib.Path('/tmp/s7s').mkdir(exist_ok=True)
for slug, state, body in boards:
    stem, title = f'{STEM0}--{slug}', f'{TITLE0} · {state}'
    root = f'<div class="ip-root" style="padding:0; width:390px; height:{VH}px; overflow:hidden">{body}</div>'
    dc = f'''<!doctype html>
<html lang="ro">
<head>
<meta charset="utf-8">
<title>{title}</title>
<script src="./support.js"></script>
<link rel="stylesheet" href="ds/impexlinepanel/tokens.css">
<link rel="stylesheet" href="ds/impexlinepanel/components/bundle.css">
<link rel="stylesheet" href="additions.css">
</head>
<body>
<x-dc>
<helmet>
{B.FONTS}
<style>{CSS}{CSS_EXTRA.get(slug, '')}</style>
</helmet>
{root}
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":390,"height":{VH}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
    (ROOT / f'design/canvas/project/{stem}.dc.html').write_text(dc, encoding='utf-8')
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{B.FONTS}<style>{CSS}{CSS_EXTRA.get(slug, '')}</style></head><body>{root}</body></html>'''
    pathlib.Path(f'/tmp/s7s/{stem}.html').write_text(wrap, encoding='utf-8')
    print('ok', stem, title, len(dc))
