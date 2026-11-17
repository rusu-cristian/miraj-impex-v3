"""Build the S4 state artboards (S4S) from the extracts. The shared shell/header/sections are functions with a no-guide
flag; % is written without a space (D46). Since S4P the default board (s4_build.py) is drawn by these same functions,
so the shell is identical on all 9 boards; the inferred-start mark and the split chip are system classes (additions.css).
python3 -I s4s_build.py <repo_root>
Writes design/canvas/project/S4-ferma-rezumat--<state>.dc.html and /tmp/s4s/<state>.html (plain wrappers)."""
import sys, json, pathlib, datetime as dt

ROOT = pathlib.Path(sys.argv[1])
EX = ROOT / 'design/data/extracts'
S = json.load(open(EX / 'sample-hirbovat-2026.json', encoding='utf-8'))
CL = json.load(open(EX / 'counts-latest.json', encoding='utf-8'))
GL = json.load(open(EX / 'guide-laying.json', encoding='utf-8'))
CY = json.load(open(EX / 'cycles.json', encoding='utf-8'))

NB = ' '; MINUS = '−'

# S4TP (D82): .5 rounds up (away from zero), never to even, as s4t_build; float noise is cut at 1e-9 first.
from decimal import Decimal, ROUND_HALF_UP
def rnd(v, dec=0):
    d = Decimal(repr(float(v))).quantize(Decimal('1e-9')).quantize(Decimal(1).scaleb(-dec), rounding=ROUND_HALF_UP)
    return float(d) if dec else int(d)

# S16P fix 1: DROP = (date, section, sex) of a deleted Mortalitate record, set on the module before it runs (only the S16
# after-delete board does, through s4t_build); unset, nothing changes. The record's deaths go back into the birds alive
# from that day on, and every stored total that counts them is recomputed in memory (the extracts stay untouched).
DROP = globals().get('DROP')
def drop_mortality(S, CL, day, sec, sx):
    rows = sorted((r for r in S['daily'] if r['section'] == sec and r['sex'] == sx and r['date'] >= day), key=lambda r: r['date'])
    q = rows[0]['mortality']; assert rows[0]['date'] == day and q > 0
    for r in rows:
        if r['date'] == day: r['mortality'] = 0
        else: r['alive_start'] += q
        r['alive_end'] += q
    def pcts(d):  # mortality_pct / removals_pct of a stored total, 2 decimals as stored
        d['mortality_pct'] = round(d['mortality'] / d['start'] * 100, 2)
        d['removals_pct'] = round((d['mortality'] + d['culls'] + d['sexing_errors']) / d['start'] * 100, 2)
    st = next(r for r in S['section_totals'] if r['section'] == sec and r['sex'] == sx)
    st['mortality'] -= q; st['alive'] += q; pcts(st)
    # feed rows (allocations, kg) and egg rows are records of their own, saved with that day's count: they stay as entered
    fm = CL['farms'][0]
    cs = next(x for x in fm['sections'] if x['section'] == sec); cs[sx] += q
    cs['detail'][sx]['mortality'] -= q; cs['detail'][sx]['alive'] += q; pcts(cs['detail'][sx])
    next(h for h in fm['halls'] if h['hall'] == sec.split('.')[0])[sx] += q
    fm['totals'][sx] += q; fm['totals']['all'] += q
    cy = fm['cycle']; rm = cy['removals'][sx]; rm['mortality'] -= q; st0 = cy['start_counts'][sx]
    cy['mortality_pct'][sx] = round(rm['mortality'] / st0 * 100, 2)
    cy['removals_pct'][sx] = round((rm['mortality'] + rm['culls'] + rm['sexing_errors']) / st0 * 100, 2)
    sm = S['summary_today']; sm['alive'][sx] += q; sm['alive']['total'] += q
    sm['cumulative_mortality_pct'][sx] = cy['mortality_pct'][sx]; sm['cumulative_removals_pct'][sx] = cy['removals_pct'][sx]
    return q
DROPQ = drop_mortality(S, CL, *DROP) if DROP else 0

def n(v, dec=0):
    v = rnd(v, dec)
    s = f'{abs(v):,.{dec}f}'.replace(',', 'X').replace('.', ',').replace('X', NB)
    return (MINUS if v < 0 else '') + s

def sgn(v, dec=1):
    v = rnd(v, dec)
    return ('+' if v > 0 else '') + n(v, dec)

def ddate(iso):
    y, m, d = iso.split('-'); return f'{d}.{m}.{y}'

# ---------- icons (copied path data from RecordTile / Sidebar / PageHeader previews) ----------
IC = {
 'mortality': '<path d="M12 3a7 7 0 0 0-7 7v4l2 2v3h10v-3l2-2v-4a7 7 0 0 0-7-7z"/><circle cx="9.5" cy="11" r="1.2"/><circle cx="14.5" cy="11" r="1.2"/><path d="M11 19v-2M13 19v-2"/>',
 'cull': '<circle cx="12" cy="12" r="9"/><path d="M5.7 5.7l12.6 12.6"/>',
 'transfer': '<path d="m16 3 4 4-4 4"/><path d="M20 7H4"/><path d="m8 21-4-4 4-4"/><path d="M4 17h16"/>',
 'sale': '<path d="M3 7h11v9H3zM14 10h4l3 3v3h-7"/><circle cx="7" cy="18" r="2"/><circle cx="17" cy="18" r="2"/>',
 'feed': '<path d="M12 21V9M12 9c-3 0-4-2-4-5 3 0 4 2 4 5zM12 9c3 0 4-2 4-5-3 0-4 2-4 5zM12 15c-3 0-4-2-4-5 3 0 4 2 4 5zM12 15c3 0 4-2 4-5-3 0-4 2-4 5z"/>',
 'delivery': '<path d="M12 3 3 7.5v9L12 21l9-4.5v-9zM3 7.5 12 12l9-4.5M12 12v9"/>',
 'price': '<ellipse cx="9" cy="7" rx="6" ry="3"/><path d="M3 7v4c0 1.7 2.7 3 6 3s6-1.3 6-3V7"/><path d="M9 14v3c0 1.7 2.7 3 6 3s6-1.3 6-3v-4c0-1.7-2.7-3-6-3"/>',
 'weighing': '<path d="M12 3v18"/><path d="M7 21h10"/><path d="M3 7h18"/><path d="m6 7-3 7a3 3 0 0 0 6 0z"/><path d="m18 7-3 7a3 3 0 0 0 6 0z"/>',
 'eggs': '<path d="M12 22c4 0 7-3 7-7.5C19 9 16 2 12 2S5 9 5 14.5C5 19 8 22 12 22z"/>',
 'sexerr': '<circle cx="8" cy="12" r="4" stroke="var(--rec-hen)"/><circle cx="16" cy="12" r="4" stroke="var(--rec-rooster)"/>',
 'chev': '<path d="m6 9 6 6 6-6"/>',
 'cycle': '<path d="M21 12a9 9 0 0 1-15.5 6.2L3 16M3 12a9 9 0 0 1 15.5-6.2L21 8"/><path d="M21 3v5h-5M3 21v-5h5"/>',
 'download': '<path d="M12 3v12"/><path d="m7 10 5 5 5-5"/><path d="M5 21h14"/>',
 'home': '<path d="M3 10.5 12 3l9 7.5V21H3z"/><path d="M9 21v-6h6v6"/>',
 'plus': '<path d="M5 12h14"/><path d="M12 5v14"/>',
 'layers': '<path d="m12 2 10 5-10 5L2 7z"/><path d="m2 12 10 5 10-5"/><path d="m2 17 10 5 10-5"/>',
 'chart': '<path d="M3 3v18h18"/><path d="m7 15 4-4 3 3 5-6"/>',
 'info': '<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>',
 'cal': '<rect x="3" y="4" width="18" height="18" rx="2"/><path d="M16 2v4M8 2v4M3 10h18"/>',
 'check': '<path d="M20 6 9 17l-5-5"/>',
 'settings': '<path d="M4 6h10"/><path d="M18 6h2"/><circle cx="16" cy="6" r="2"/><path d="M4 18h4"/><path d="M12 18h8"/><circle cx="10" cy="18" r="2"/>',
}
REC = {'mortality': 'mortality', 'cull': 'cull', 'transfer': 'transfer', 'sale': 'sale', 'feed': 'feed',
       'delivery': 'feed', 'price': 'feed', 'weighing': 'weighing', 'eggs': 'eggs'}

def icon(k, size=None):
    cls = 'ip-icon' + (f' ip-icon--{size}' if size else '')
    return f'<svg class="{cls}" viewBox="0 0 24 24" aria-hidden="true">{IC[k]}</svg>'

def tile(k):
    return f'<span class="ip-tile ip-rec ip-rec--{REC[k]}" aria-hidden="true">{icon(k)}</span>'

def ico(k):  # small record icon in its colour, for captions and key/value lines
    return f'<span class="ip-rec ip-rec--{REC[k]} s4-ico">{icon(k, 14)}</span>'

def hen_v(v):  # hen figure without the sign (the label already carries ♀)
    return f'<span class="ip-sex ip-sex--hen"><b>{v}</b></span>'

def hen(v, unit=''):
    return f'<span class="ip-sex ip-sex--hen">♀ <b>{v}{unit}</b></span>'

def roo(v, unit=''):
    return f'<span class="ip-sex ip-sex--rooster">♂ <b>{v}{unit}</b></span>'

def sg(sx):  # bare ♀/♂ sign in its record colour
    return '<span class="ip-sex ip-sex--hen">♀</span>' if sx == 'F' else '<span class="ip-sex ip-sex--rooster">♂</span>'
F, M = sg('F'), sg('M')

def pair(f, m, unit=''):
    return f'<span class="ip-pair">{hen(f, unit)}<span class="sep">·</span>{roo(m, unit)}</span>'

def lamp(kind, word):
    return f'<span class="ip-lamp ip-lamp--{kind}">{word}</span>'

# ---------- data ----------
farm = CL['farms'][0]
alive = {(x['section'], sx): x[sx] for x in farm['sections'] for sx in 'FM'}
SECTIONS = ['H1.1', 'H1.2', 'H3.1', 'H3.2', 'H4.1', 'H4.2']
TODAY = '2026-11-16'; WK = 40
wk_rows = [r for r in S['daily'] if '2026-11-10' <= r['date'] <= TODAY]
mort_wk = {s: sum(r['mortality'] for r in wk_rows if r['section'] == s) for s in SECTIONS}
mort_wk_sex = {sx: sum(r['mortality'] for r in wk_rows if r['sex'] == sx) for sx in 'FM'}
alloc = {(a['section'], a['sex']): a for a in S['allocations_today']}
wt = {(w['section'], w['sex']): w['avg_weight_g'] for w in S['weighings'] if w['date'] == TODAY}
def wavg(d, sx):
    keys = [k for k in d if k[1] == sx and d[k] is not None]
    return sum(d[k] * alive[k] for k in keys) / sum(alive[k] for k in keys)
w_f = rnd(wavg(wt, 'F')); w_m = rnd(wavg(wt, 'M'))
g_today = {k: a['today_g'] for k, a in alloc.items()}
g_f = rnd(wavg(g_today, 'F')); g_m = rnd(wavg(g_today, 'M'))
guide = {(r['cycle_week'], r['sex']): r for r in GL['rows']}
gf_g = guide[(WK, 'F')]['feed_g']; gm_g = guide[(WK, 'M')]['feed_g']
eggs_today = {r['hall']: r for r in S['eggs_daily'] if r['date'] == TODAY}
lay_all = {r['cycle_week']: r for r in S['eggs_weekly'] if r['hall'] == 'all'}
st = S['summary_today']
lay_today = st['lay_pct_today']; lay_guide = st['guide_lay_pct']
stocks = {s['feed_type_id']: s for s in S['stock_today']}
unpriced = [d for d in S['deliveries'] if not d['priced']]
unpriced_kg = sum(d['qty_kg'] for d in unpriced)
cum = farm['cycle']
tot = farm['totals']

def pct(v, dec=1):  # D46: no space before %
    return n(v, dec) + '%'
c25 = next(c for c in CY['cycles'] if c['id'] == 'cyc-hirbovat-2025')
c26 = next(c for c in CY['cycles'] if c['id'] == 'cyc-hirbovat-2026')
assert c26['week_today'] == WK and c26['start_date'] == '2026-02-10'
PREV_WEEKS = c25['weeks']; PREV_END = c25['end_date']  # 47, 2026-01-28
assert (PREV_WEEKS, PREV_END) == (47, '2026-01-28')

# ---------- sidebar ----------
def sidebar(role='Proprietar', viewer=False, series='2026'):
    s26 = series == '2026'
    cur = f'<span class="name">Seria 2026</span>{lamp("ok", "deschisă")}' if s26 else f'<span class="name">Seria 2025</span>{lamp("off", "închisă")}'
    nav =f'<a class="ip-nav" href="#acasa">{icon("home")}Acasă</a>'
    if not viewer: nav += f'<a class="ip-nav" href="#inregistreaza">{icon("plus")}Înregistrează</a>'
    nav += f'<a class="ip-nav is-current" href="#serie" aria-current="page">{icon("layers")}Serie</a><a class="ip-nav" href="#rapoarte">{icon("chart")}Rapoarte</a>'
    if not viewer: nav += f'<a class="ip-nav" href="#setari">{icon("settings")}Setări</a>'
    return f'''<nav class="ip-sidebar s4-side" aria-label="Navigare">
<div class="ip-wordmark">Impex Line</div>
<div class="ip-series"><span class="ip-sr" id="ser-lbl">Serie</span><div class="ip-input ip-series__btn" id="ser-combo" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="false" aria-controls="ser-list" aria-labelledby="ser-lbl">{cur}{icon('chev', 16)}</div>
<ul class="ip-menu ip-series__menu" role="listbox" id="ser-list" aria-labelledby="ser-lbl" tabindex="-1" hidden="hidden"><li class="ip-menu__group" role="presentation">Serii</li><li class="ip-menu__item ip-series__opt" role="option" id="ser-2026" aria-selected="{'true' if s26 else 'false'}"><span class="chk"></span><span class="txt"><span class="top"><b>Seria 2026</b>{lamp('ok', 'deschisă')}</span><span class="ip-num">din 07.10.2025</span></span></li><li class="ip-menu__item ip-series__opt" role="option" id="ser-2025" aria-selected="{'false' if s26 else 'true'}"><span class="chk"></span><span class="txt"><span class="top"><b>Seria 2025</b>{lamp('off', 'închisă')}</span><span class="ip-num">până la 28.01.2026</span></span></li></ul></div>
{nav}
<span class="s4-grow"></span>
<div class="ip-note s4-account"><b>{role}</b><a class="ip-link" href="#iesi">Ieși</a></div>
</nav>'''

# ---------- page header ----------
def crumbs(series='2026'):
    return f'<nav class="ip-crumbs" aria-label="Cale"><a href="#seria-{series}">Seria {series}</a><span aria-hidden="true">›</span><span aria-current="page">Hîrbovăț</span></nav>'
CRUMBS = crumbs()
def header(mode='default'):
    if mode == 'noaccess':  # no counts, no guide, no week, no actions: nothing about the farm leaks
        return f'<header class="ip-pagehead"><div>{CRUMBS}<h1>Hîrbovăț</h1></div></header>'
    if mode == 'past':  # chosen closed cycle Seria 2025 (D51): no alive today (D43); start inferred (D36), no flock count in the workbook
        meta = (f'<span class="ip-tag ip-tag--neutral">Ouat</span><span class="ip-tag ip-tag--neutral">Ross 308 ouat</span>{lamp("off", "închis")}'
                f'<span class="ip-note">Început dedus <span class="ip-num">{ddate(c25["start_date"])}</span> · încheiat prin golire <span class="ip-num">{ddate(PREV_END)}</span> · <span class="ip-num">{PREV_WEEKS} săpt.</span></span>'
                f'<span class="ip-note">efectiv – fără date în registru</span>')
        acts = '<a class="ip-btn ip-btn--secondary" href="#ciclu-seria-2026">Înapoi la ciclul curent</a>'
        return f'''<header class="ip-pagehead">
<div>
{crumbs('2025')}
<h1>Hîrbovăț</h1>
<div class="ip-row s4-meta">{meta}</div>
</div>
<div class="ip-row">
{acts}
</div>
</header>'''
    if mode == 'empty':  # 30.01.2026: no birds, no open cycle; only the move that can start a cycle
        meta = f'<span class="ip-tag ip-tag--neutral">Ouat</span><span class="ip-tag ip-tag--neutral">Ross 308 ouat</span>{lamp("off", "goală")}<span class="ip-note ip-num">din {ddate(PREV_END)} · niciun ciclu deschis</span>'
        acts = f'<button type="button" class="ip-btn ip-btn--secondary">{icon("transfer", 16)}Mută păsări</button>'
    else:
        gtag = '<span class="ip-tag ip-tag--neutral">fără ghid</span>' if mode == 'noguide' else '<span class="ip-tag ip-tag--neutral">Ross 308 ouat</span>'
        meta = f'<span class="ip-tag ip-tag--neutral">Ouat</span>{gtag}{pair(n(tot["F"]), n(tot["M"]))}<span class="ip-note ip-num">SAPT {WK} · ciclu din 10.02.2026</span>'
        acts = f'''<button type="button" class="ip-btn ip-btn--secondary">{icon('transfer', 16)}Mută păsări</button>
<button type="button" class="ip-btn ip-btn--secondary">{icon('sale', 16)}Vânzare</button>
<button type="button" class="ip-btn ip-btn--ghost" aria-haspopup="menu" aria-expanded="false" aria-label="Acțiuni ciclu">{icon('cycle', 16)}Ciclu{icon('chev', 14)}</button>'''
    return f'''<header class="ip-pagehead">
<div>
{CRUMBS}
<h1>Hîrbovăț</h1>
<div class="ip-row s4-meta">{meta}</div>
</div>
<div class="ip-row">
{acts}
</div>
</header>'''

# ---------- cycle selector + farm track ----------
TW = 800
DOM_LIVE = (dt.date(2025, 3, 1), dt.date(2027, 1, 31), [('2025-04-01', '04.25'), ('2025-07-01', '07.25'), ('2025-10-01', '10.25'), ('2026-01-01', '01.26'),
            ('2026-04-01', '04.26'), ('2026-07-01', '07.26'), ('2026-10-01', '10.26'), ('2027-01-01', '01.27')])
# empty farm on 30.01.2026: the axis ends near "azi" (01.2025 – 03.2026), no future months after a cycle that doesn't exist yet
DOM_EMPTY = (dt.date(2025, 1, 1), dt.date(2026, 3, 31), [('2025-01-01', '01.25'), ('2025-04-01', '04.25'), ('2025-07-01', '07.25'), ('2025-10-01', '10.25'),
             ('2026-01-01', '01.26'), ('2026-03-01', '03.26')])
def track(today_iso, sel='2026', dom=DOM_LIVE):
    """sel: the cycle shown on the page ('2026', '2025') or None (empty farm: nothing current). Seria 2026 exists only when today ≥ its start."""
    d0, d1, tks = dom
    ppd = TW / (d1 - d0).days
    def tx(iso): return round(((dt.date.fromisoformat(iso) - d0).days) * ppd, 1)
    with26 = today_iso >= c26['start_date']
    x25a, x25b = tx(c25['start_date']), tx(c25['end_date'])
    xt = tx(today_iso)
    ticks = ''
    for iso, lbl in tks:
        x = tx(iso); anc = 'start' if x < 8 else 'middle'
        ticks += f'<line class="tick" x1="{x}" x2="{x}" y1="40" y2="45"/><text class="month" x="{x}" y="56" text-anchor="{anc}">{lbl}</text>'
    act25 = 'Ciclul afișat.' if sel == '2025' else 'Deschide ciclul.'
    act26 = 'Ciclul afișat.' if sel == '2026' else 'Deschide ciclul.'
    assert c25['start_inferred'] and c25['start_date'] == '2025-03-07' and PREV_END == '2026-01-28'  # D54: no start event, ended by emptying (no sale)
    lab25 = f'Seria 2025, ouat, început dedus {ddate(c25["start_date"])} · încheiat prin golire {ddate(PREV_END)}, {PREV_WEEKS} săptămâni, închis. {act25}'
    lab26 = f'Seria 2026, ouat, din 10.02.2026, deschis, azi SAPT {WK}. Început prin transfer. {act26}'
    w25 = round(x25b - x25a, 1)
    cur25 = ' aria-current="true"' if sel == '2025' else ''
    sel25 = f'<rect class="sel" x="{x25a}" y="37" width="{w25}" height="2"/>' if sel == '2025' else ''
    b25 = f'<a href="#ciclu-seria-2025"{cur25} aria-label="{lab25}"><title>{lab25}</title><rect class="ring" x="{round(x25a-4,1)}" y="12" width="{round(w25+8,1)}" height="30" rx="3"/><rect class="cyc" x="{x25a}" y="16" width="{w25}" height="18" rx="2"/><text class="cyctxt" x="{round(x25a+7,1)}" y="29">Seria 2025 · ouat · {PREV_WEEKS} săpt.</text><line class="ded" x1="{round(x25a-3,1)}" x2="{round(x25a-3,1)}" y1="14" y2="40"/>{sel25}</a>'
    tl = ddate(today_iso)[:5]
    if with26:
        x26a = tx(c26['start_date']); w26 = round(xt - x26a, 1)
        cur26 = ' aria-current="true"' if sel == '2026' else ''
        sel26 = f'<rect class="sel" x="{x26a}" y="37" width="{w26}" height="2"/>' if sel == '2026' else ''
        b26 = f'<a href="#ciclu-seria-2026"{cur26} aria-label="{lab26}"><title>{lab26}</title><rect class="ring" x="{round(x26a-4,1)}" y="12" width="{round(w26+8,1)}" height="30" rx="3"/><rect class="cyc open" x="{x26a}" y="16" width="{w26}" height="18" rx="2"/><text class="cyctxt open" x="{round(x26a+7,1)}" y="29">Seria 2026 · ouat · din 10.02.2026</text><rect class="xfer" x="{round(x26a-1,1)}" y="16" width="3" height="18"/>{sel26}</a>'
        ro = f'<rect class="ro" x="{round(xt+4,1)}" y="17" width="48" height="16" rx="2"/><text class="rotxt" x="{round(xt+28,1)}" y="28.5" text-anchor="middle">SAPT {WK}</text>'
        shown = {'2025': ' Afișat: Seria 2025.', '2026': ' Afișat: Seria 2026.'}.get(sel, '')
        desc = f'Hîrbovăț: Seria 2025 ouat, început dedus 07.03.2025 · încheiat prin golire 28.01.2026, {PREV_WEEKS} săpt., închis; Seria 2026 ouat din 10.02.2026 (transfer), deschis, azi 16.11.2026, SAPT {WK}.{shown}'
    else:  # empty farm: no open cycle, so no readout next to "azi" and no current block
        b26 = ''; ro = ''
        desc = f'Hîrbovăț: Seria 2025 ouat, început dedus 07.03.2025 · încheiat prin golire 28.01.2026, {PREV_WEEKS} săpt., închis; ferma e goală din 28.01.2026, azi {ddate(today_iso)}. Niciun ciclu deschis.'
    return f'''<svg class="ip-track ip-track--farm" viewBox="0 0 {TW} 60" role="group" aria-label="Ciclurile fermei Hîrbovăț"><desc>{desc}</desc><rect class="lane" x="0" y="14" width="{TW}" height="26"/>{b25}{b26}<line class="today" x1="{xt}" x2="{xt}" y1="12" y2="42"/><text class="todaylbl" x="{xt}" y="9" text-anchor="middle">azi {tl}</text>{ro}{ticks}</svg>'''
def legend(xfer=True):  # names only the marks drawn: Seria 2026 starts by transfer; Seria 2025's start is inferred (D54), no sale
    x = '<span><i class="ip-rec ip-rec--transfer"></i>transfer</span>' if xfer else ''
    return f'<div class="ip-tracklegend" aria-hidden="true">{x}<span><svg class="ded" viewBox="0 0 3 12"><line x1="1.5" x2="1.5" y1="0" y2="12"/></svg>început dedus</span><span><i class="today"></i>azi</span></div>'
LEGEND = legend()
def opt26(selected):
    return f'''<li class="ip-menu__item ip-series__opt{" is-active" if selected else ""}" role="option" id="cyc-2026" aria-selected="{"true" if selected else "false"}"><span class="chk">{icon('check', 16) if selected else ""}</span><span class="txt"><span class="top"><b>Seria 2026 · ouat</b>{lamp('ok', 'deschis')}</span><span class="ip-num">din 10.02.2026 · SAPT {WK}</span></span></li>'''
def opt25(selected):
    return f'''<li class="ip-menu__item ip-series__opt{" is-active" if selected else ""}" role="option" id="cyc-2025" aria-selected="{"true" if selected else "false"}"><span class="chk">{icon('check', 16) if selected else ""}</span><span class="txt"><span class="top"><b>Seria 2025 · ouat</b>{lamp('off', 'închis')}</span><span class="ip-num">07.03.2025 – {ddate(PREV_END)} · {PREV_WEEKS} săpt.</span></span></li>'''
GRP = '<li class="ip-menu__group" role="presentation">Cicluri la Hîrbovăț</li>'
CYC_ITEMS = GRP + opt26(True) + opt25(False)
def hidden_list(items):  # collapsed combobox: its listbox exists (aria-controls) but is hidden, as the sidebar series selector
    return f'<ul class="ip-menu s4s-cycmenu s4s-list" role="listbox" id="cyc-list" aria-labelledby="cyc-lbl" tabindex="-1" hidden="hidden">{items}</ul>'
def combo(text, lmp, opened=False, active=''):
    ad = f' aria-activedescendant="{active}"' if opened else ''
    return f'<div class="ip-input s4-cycsel{" is-focus" if opened else ""}" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="{"true" if opened else "false"}" aria-controls="cyc-list" aria-labelledby="cyc-lbl"{ad}><span class="s4-grow">{text}</span>{lmp}{icon("chev", 16)}</div>'
def cycle(mode='default'):
    if mode == 'empty':  # D51: no open cycle; Seria 2025 only as an option and in the hint; no block current on the track
        sel = combo('Niciun ciclu deschis', lamp('off', 'goală'))
        hint = f'<span class="ip-hint">ultimul ciclu · Seria 2025 · 07.03.2025 – {ddate(PREV_END)}</span>'
        return f'''<section class="s4-cycle" aria-label="Ciclu">
<div class="ip-field s4s-anchor"><span class="ip-label" id="cyc-lbl">Ciclu</span>{sel}{hint}{hidden_list(GRP + opt25(False))}</div>
<div class="ip-stack s4-tight">{track('2026-01-30', sel=None, dom=DOM_EMPTY)}{legend(xfer=False)}</div>
</section>'''
    if mode == 'past':  # D51 "ciclu ales": the page rescoped to the closed Seria 2025 cycle
        sel = combo(f'Seria 2025 · {PREV_WEEKS} săpt.', lamp('off', 'închis'))
        hint = f'<span class="ip-hint">început dedus {ddate(c25["start_date"])} · date din {ddate(c25["data_start_date"])}</span>'
        return f'''<section class="s4-cycle" aria-label="Ciclu">
<div class="ip-field s4s-anchor"><span class="ip-label" id="cyc-lbl">Ciclu</span>{sel}{hint}{hidden_list(GRP + opt26(False) + opt25(True))}</div>
<div class="ip-stack s4-tight">{track(TODAY, sel='2025')}{LEGEND}</div>
</section>'''
    opened = mode == 'menu'
    sel = combo('Seria 2026 · din 10.02.2026', lamp('ok', 'deschis'), opened, 'cyc-2026')
    if opened:  # CycleSelector open: cycles (open first), then the manager+ actions (S4 is the owner's view)
        menu = f'''<div class="ip-menu s4s-cycmenu"><ul class="s4s-list" role="listbox" id="cyc-list" aria-labelledby="cyc-lbl">{CYC_ITEMS}</ul><div class="ip-menu__group" id="cyc-act">Ciclu</div><div role="group" aria-labelledby="cyc-act"><button type="button" class="ip-menu__item s4s-mitem">{icon('cycle', 16)}<span class="s4-grow">Începe ciclu nou</span></button><button type="button" class="ip-menu__item s4s-mitem"><span class="s4s-gap" aria-hidden="true"></span><span class="s4-grow">Schimbă ghidul ciclului</span></button></div></div>'''
    else:
        menu = hidden_list(CYC_ITEMS)
    return f'''<section class="s4-cycle" aria-label="Ciclu">
<div class="ip-field s4s-anchor"><span class="ip-label" id="cyc-lbl">Ciclu</span>{sel}<span class="ip-hint">2 cicluri la Hîrbovăț · deschis din 10.02.2026</span>{menu}</div>
<div class="ip-stack s4-tight">{track(TODAY)}{LEGEND}</div>
</section>'''

# ---------- filter bar + tabs ----------
def seg(label, items, sel):
    its = ''.join(f'<span class="ip-seg__item{" is-selected" if i == sel else ""}" role="radio" aria-checked="{"true" if i == sel else "false"}" tabindex="{0 if i == sel else -1}">{t}</span>' for i, t in enumerate(items))
    return f'<span class="ip-seg" role="radiogroup" aria-label="{label}">{its}</span>'
def chip(k, word, on=True):
    return f'<button type="button" class="ip-chip ip-rec ip-rec--{REC[k]}{"" if on else " ip-chip--off"}" role="checkbox" aria-checked="{"true" if on else "false"}">{icon(k, 16)}{word}</button>'
chip_sexerr = f'<button type="button" class="ip-chip ip-rec ip-chip--split" role="checkbox" aria-checked="true"><span class="ip-tile ip-tile--split" aria-hidden="true">{icon("sexerr", 16)}</span>Eroare sexare</button>'
assert S['sexing_errors'] and all(e['date'] >= '2026-02-10' for e in S['sexing_errors'])
CHIPS = f'<div class="s4-chips" role="group" aria-label="Tipuri de înregistrări">{chip("mortality", "Mortalitate")}{chip("cull", "Sacrificare")}{chip_sexerr}{chip("transfer", "Transfer")}{chip("feed", "Hrană")}{chip("weighing", "Cântărire")}{chip("eggs", "Ouă")}</div>'
EXPORT = f'<button type="button" class="ip-btn ip-btn--ghost ip-filterbar__end">{icon("download", 16)}Exportă .xlsx</button>'
PERIODS = ['Ciclul curent', '4 săpt.', '13 săpt.', 'Interval']
def datef(i, label, val):
    return f'<span class="ip-datef"><label class="ip-label" for="{i}">{label}</label><span class="ip-input"><input id="{i}" type="text" inputmode="numeric" value="{val}" placeholder="zz.ll.aaaa" aria-describedby="{i}-fmt"/><button type="button" class="ip-iconbtn" aria-label="Alege data „{label}” din calendar" aria-haspopup="dialog" aria-expanded="false">{icon("cal", 16)}</button></span></span><span class="ip-sr" id="{i}-fmt">Format zz.ll.aaaa, cel târziu 16.11.2026</span>'
def filterbar(mode='default'):
    if mode == 'period':  # Interval 01.11–15.11.2025: no record of the cycle in it, so no type chips (D45) and nothing to export
        itv = f'<span class="ip-interval" role="group" aria-label="Interval">{datef("iv-a", "de la", "01.11.2025")}{datef("iv-b", "până la", "15.11.2025")}</span>'
        exp = f'<button type="button" class="ip-btn ip-btn--ghost ip-filterbar__end is-disabled" disabled="disabled">{icon("download", 16)}Exportă .xlsx</button>'
        return f'<div class="ip-filterbar s4-filterbar" role="group" aria-label="Filtre">{seg("Perioadă", PERIODS, 3)}{itv}{seg("Pas", ["Zi", "Săpt."], 1)}{exp}</div>'
    if mode == 'past':  # closed cycle: whole-cycle period; only the record types the previous flock has (eggs, mortality)
        chips = f'<div class="s4-chips" role="group" aria-label="Tipuri de înregistrări">{chip("mortality", "Mortalitate")}{chip("eggs", "Ouă")}</div>'
        return f'<div class="ip-filterbar s4-filterbar" role="group" aria-label="Filtre">{seg("Perioadă", ["Tot ciclul"] + PERIODS[1:], 0)}{seg("Pas", ["Zi", "Săpt."], 1)}{EXPORT}{chips}</div>'
    return f'<div class="ip-filterbar s4-filterbar" role="group" aria-label="Filtre">{seg("Perioadă", PERIODS, 0)}{seg("Pas", ["Zi", "Săpt."], 1)}{EXPORT}{CHIPS}</div>'
TABS = [('rezumat', 'Rezumat'), ('pasari', 'Păsări'), ('hrana', 'Hrană'), ('greutate', 'Greutate'), ('productie', 'Producție')]
tabs = '<div class="ip-tabs" role="tablist" aria-label="Vederi Hîrbovăț">' + ''.join(
    f'<a class="ip-tab{" is-selected" if i == 0 else ""}" role="tab" id="tab-{k}" aria-controls="tp-ferma" aria-selected="{"true" if i == 0 else "false"}" tabindex="{0 if i == 0 else -1}" href="#{k}">{t}</a>'
    for i, (k, t) in enumerate(TABS)) + '</div>'

# ---------- readout + guide chart ----------
CW, CH = 540, 216  # S4P: drawn 1:1 in the Readout's 540px content box (no scaling); Readout = the 3 Stat rows (364)
X0, X1, YT, YB = 40, CW - 8, 10, CH - 22
WMAX = 45
def cx(w): return round(X0 + (w - 1) * (X1 - X0) / (WMAX - 1), 1)
def cy(p): return round(YB - p / 100 * (YB - YT), 1)
gweeks = [w for w in range(1, WMAX + 1) if guide[(w, 'F')]['lay_pct'] is not None]
gv = {w: guide[(w, 'F')]['lay_pct'] for w in gweeks}
aweeks = sorted(w for w in lay_all if w <= WK)
a40 = lay_all[WK]['lay_pct']; dev = rnd(a40 - gv[WK], 1)
SWATCH = {k: f'<svg viewBox="0 0 24 8" aria-hidden="true">{v}</svg>' for k, v in {
    'act': '<line class="act" x1="0" x2="24" y1="4" y2="4"/>', 'guide': '<line class="guide" x1="0" x2="24" y1="4" y2="4"/>',
    'band': '<rect class="band" x="0" y="0" width="24" height="8"/>'}.items()}
def chart_legend(items, note=''):  # SYS06 fix 1: .ip-chart__legend swatches; ip-readout__sub keeps only context
    return '<div class="ip-chart__legend">' + ''.join(f'<span>{SWATCH[k]}{t}</span>' for k, t in items) + (f'<span>{note}</span>' if note else '') + '</div>'
def sr_table(caption, cols, rows):  # hidden data table of a chart (SYS06): thead th[scope=col], th[scope=row] per SAPT
    return (f'<table class="ip-sr"><caption>{caption}</caption><thead><tr>' + ''.join(f'<th scope="col">{c}</th>' for c in cols) + '</tr></thead><tbody>'
            + ''.join('<tr><th scope="row">' + str(r[0]) + '</th>' + ''.join(f'<td>{c}</td>' for c in r[1:]) + '</tr>' for r in rows) + '</tbody></table>')
def readout(ng=False):
    band = ' '.join(f'{cx(w)},{cy(min(gv[w]+3,100))}' for w in gweeks) + ' ' + ' '.join(f'{cx(w)},{cy(gv[w]-3)}' for w in reversed(gweeks))
    gline = ' '.join(f'{cx(w)},{cy(gv[w])}' for w in gweeks)
    aline = ' '.join(f'{cx(w)},{cy(lay_all[w]["lay_pct"])}' for w in aweeks)
    grid = ''
    for p in (0, 50, 100):
        grid += f'<line class="{"axis" if p == 0 else "grid"}" x1="{X0}" x2="{X1}" y1="{cy(p)}" y2="{cy(p)}"/><text class="ax" x="{X0-4}" y="{cy(p)+3}" text-anchor="end">{p}%</text>'
    xl = ''.join(f'<text class="ax" x="{cx(w)}" y="{CH-6}" text-anchor="middle">{w}</text>' for w in [1, 10, 20, 30, 40, 45])
    gpart = '' if ng else f'<polygon class="band" points="{band}"/><polyline class="guide" points="{gline}"/>'
    lbl = f'Ouat Hîrbovăț pe săptămânile ciclului, SAPT 3–{WK}; ciclul nu are ghid' if ng else f'Ouat Hîrbovăț față de ghid, pe săptămânile ciclului, SAPT 3–{WK}; ghid SAPT 7–45'
    chart = f'''<svg class="ip-chart" viewBox="0 0 {CW} {CH}" role="group" aria-roledescription="grafic" tabindex="0" aria-label="{lbl}" aria-describedby="lay-help lay-read"><g aria-hidden="true">{grid}{gpart}<polyline class="act" points="{aline}"/><line class="today" x1="{cx(WK)}" x2="{cx(WK)}" y1="{YT-2}" y2="{YB}"/><line class="cur" x1="{cx(WK)}" x2="{cx(WK)}" y1="{YT-2}" y2="{YB}"/><circle class="pt" cx="{cx(WK)}" cy="{cy(a40)}" r="3.5"/>{xl}</g></svg>'''
    if ng:
        rows = [(w, pct(lay_all[w]['lay_pct']), '– fără ghid') for w in aweeks]
        rl = '<span class="ip-rlamp ip-rlamp--off">FĂRĂ GHID</span>'
        sub = 'ciclul nu are ghid, nu se compară · 10–16.11.2026'
        read = f'SAPT {WK} · efectiv {pct(a40)} · ghid – fără ghid'
        leg = chart_legend([('act', 'efectiv')], 'axa: săpt. ciclului')
    else:
        rows = [(w, pct(lay_all[w]['lay_pct']), pct(gv[w]) if w in gv else '– fără date în ghid') for w in aweeks]
        rl = '<span class="ip-rlamp ip-rlamp--ok">ÎN NORMĂ</span>'
        sub = f'ghid {pct(gv[WK])} · {sgn(dev)} pct · 10–16.11.2026'
        read = f'SAPT {WK} · efectiv {pct(a40)} · ghid {pct(gv[WK])} · {sgn(dev)} pct'
        leg = chart_legend([('act', 'efectiv'), ('guide', 'ghid'), ('band', '±3 pct')], 'axa: săpt. ciclului')  # = Producție
    return f'''<div class="ip-readout s4-readout">
<div class="ip-readout__cap"><span>OUAT · HÎRBOVĂȚ · SAPT {WK}</span>{rl}</div>
<div class="ip-readout__val">{n(a40,1)}<small>%</small></div>
<div class="ip-readout__sub">{sub}</div>
{chart}
<span class="ip-sr" id="lay-help">Săgețile stânga și dreapta mută săptămâna; Home și End: prima și ultima săptămână.</span>
<div class="ip-chart__read" id="lay-read" role="status" aria-live="polite">{read}</div>
{sr_table(f'Ouat Hîrbovăț pe săptămâni{"" if ng else ", efectiv față de ghid"}; înregistrări din 01.03.2026 (SAPT 3)', ['SAPT', 'Efectiv', 'Ghid'], rows)}
{leg}
</div>'''

# ---------- stats ----------
def gline(line):  # S4TP: "ghid ♀158 · ♂155 ● în normă" → <span>ghid ♀158 · ♂155</span> + lamp (no 8px gaps between words)
    i = line.find('<span class="ip-lamp')
    head, tail = (line, '') if i < 0 else (line[:i].rstrip(), line[i:])
    if head and not (head.startswith('<span') and head.endswith('</span>') and head.count('<span') == 1):
        head = f'<span>{head}</span>'
    return head + tail
def guides(html):  # apply gline() inside every Stat guide line of a block of Stat markup
    out, i, OPEN = [], 0, '<span class="ip-stat__guide'
    while True:
        j = html.find(OPEN, i)
        if j < 0: out.append(html[i:]); return ''.join(out)
        k = html.index('>', j) + 1; depth, m = 1, k
        while depth:
            a, b = html.find('<span', m), html.find('</span>', m)
            if a != -1 and a < b: depth += 1; m = a + 5
            else: depth -= 1; m = b + 7
        out.append(html[i:k] + gline(html[k:m - 7]) + '</span>'); i = m
def cap(k, text):
    return f'<span class="ip-cap s4-capi">{ico(k) if k else ""}{text}</span>'
s_b2 = stocks['ft-breeder2']; s_co = stocks['ft-cocosi']
NOG = 'ghid: – fără ghid'
def stats(ng=False):
    g_feed = NOG if ng else f'ghid {pair(gf_g, gm_g)} {lamp("ok", "în normă")}'
    g_lay = NOG if ng else f'ghid {pct(lay_guide)} · {sgn(rnd(lay_today-lay_guide,1))} pct {lamp("ok", "în normă")}'
    g_none = NOG if ng else 'ghid: – fără date în ghid'
    return guides(f'''<div class="ip-stats s4-stats">
<div class="ip-stat">{cap(None, 'În viață')}<span class="ip-stat__val">{pair(n(tot['F']), n(tot['M']))}</span><span class="ip-stat__guide ip-num">total {n(tot['all'])}</span><span class="ip-stat__guide">la sosire {pair(n(cum['start_counts']['F']), n(cum['start_counts']['M']))}</span></div>
<div class="ip-stat">{cap('mortality', 'Mortalitate cumulată')}<span class="ip-stat__val">{pair(n(cum['mortality_pct']['F'],2), n(cum['mortality_pct']['M'],2))}<small>%</small></span><span class="ip-stat__guide">{g_none}</span><span class="ip-stat__guide">din 10.02.2026 {pair(n(-cum['removals']['F']['mortality']), n(-cum['removals']['M']['mortality']))}</span></div>
<div class="ip-stat">{cap('weighing', f'Greutate {F} / {M}')}<span class="ip-stat__val">{pair(n(w_f), n(w_m))}<small>g</small></span><span class="ip-stat__guide">{g_none}</span><span class="ip-stat__guide ip-num">cântărit 16.11 · 6\u00a0secțiuni</span></div>
<div class="ip-stat">{cap('feed', 'Hrană azi pe pasăre')}<span class="ip-stat__val">{pair(g_f, g_m)}<small>g</small></span><span class="ip-stat__guide">{g_feed}</span><span class="ip-stat__guide"><span>fără {F} H4.2 · {M} H3.2 (lipsă azi)</span></span></div>
<div class="ip-stat">{cap('feed', 'Stoc hrană · zile rămase')}<span class="ip-stat__val">{n(s_co['days_left'],1)}<small>zile · Cocoși</small></span><span class="ip-stat__guide"><span class="ip-num">până pe {ddate(s_co['run_out_date'])}</span>{lamp('warn', 'abatere')}</span><span class="ip-stat__guide"><span>Breeder 2 · <span class="ip-num">{n(s_b2['days_left'],1)} zile · {n(s_b2['stock_kg'])} kg</span></span></span></div>
<div class="ip-stat">{cap('eggs', 'Ouat azi')}<span class="ip-stat__val">{n(lay_today,1)}<small>%</small></span><span class="ip-stat__guide">{g_lay}</span><span class="ip-stat__guide ip-num">{n(sum(e['total_eggs'] for e in eggs_today.values()))} ouă · 3 hale</span></div>
</div>''')

# ---------- section tiles + totals ----------
def tstate(sec, ng=False):
    hall = sec.split('.')[0]; out = []
    if g_today[(sec, 'F')] is None or g_today[(sec, 'M')] is None: out.append(('warn', 'lipsă'))
    if ng: return out  # no guide: no deviation and no "în normă" (nothing to compare with); "lipsă" is not guide-based
    feed_dev = any(g_today[(sec, sx)] is not None and abs(g_today[(sec, sx)] / guide[(WK, sx)]['feed_g'] - 1) > 0.10 for sx in 'FM')
    lay_dev = abs(eggs_today[hall]['lay_pct'] - eggs_today[hall]['guide_lay_pct']) > 3
    if feed_dev or lay_dev: out.append(('warn', 'abatere'))
    return out or [('ok', 'în normă')]
def gpair(sec):
    f = g_today[(sec, 'F')]; m = g_today[(sec, 'M')]
    if f is None or m is None:
        miss = '<span class="s4-miss">– lipsă</span>'
        fs = F + miss if f is None else hen(f) + '<span class="s4-unit">g</span>'
        ms = M + miss if m is None else roo(m) + '<span class="s4-unit">g</span>'
        return f'<span class="ip-pair s4-kvpair">{fs}<span class="sep">·</span>{ms}</span>'
    return f'<span class="ip-pair s4-kvpair">{hen(f)}<span class="sep">·</span>{roo(m)}<span class="s4-unit">g</span></span>'
def stile(sec, ng=False):
    hall = sec.split('.')[0]; f, m = alive[(sec, 'F')], alive[(sec, 'M')]
    pf = round(f / (f + m) * 100)
    missing = g_today[(sec, 'F')] is None or g_today[(sec, 'M')] is None
    lamps = ''.join(lamp(k, w) for k, w in tstate(sec, ng))
    okey = f'Ouat {hall}' if ng else f'Ouat {hall} · ghid {pct(eggs_today[hall]["guide_lay_pct"])}'
    return f'''<a href="#sectiune-{sec}" class="ip-stile{" is-missing" if missing else ""}"><div class="ip-stile__head"><span class="ip-stile__code">{sec}</span><span class="s4-lamps">{lamps}</span></div>{pair(n(f), n(m))}<div class="ip-bar" aria-hidden="true"><span style="width:{pf}%"></span><span style="width:{100-pf}%"></span></div><div class="ip-stile__kv"><span class="s4-capi">{ico('mortality')}Mort. săpt.</span><b>{n(-mort_wk[sec])}</b></div><div class="ip-stile__kv"><span class="s4-capi">{ico('feed')}Hrană azi</span>{gpair(sec)}</div><div class="ip-stile__kv"><span class="s4-capi">{ico('weighing')}Greutate {F}</span><span class="ip-pair s4-kvpair">{hen_v(n(wt[(sec,'F')]))}<span class="s4-unit">g</span></span></div><div class="ip-stile__kv"><span class="s4-capi">{ico('eggs')}{okey}</span><b>{pct(eggs_today[hall]['lay_pct'])}</b></div></a>'''
def halls(ng=False):
    out = ''
    for hall in ['H1', 'H3', 'H4']:
        secs = [s for s in SECTIONS if s.startswith(hall + '.')]
        hf = sum(alive[(s, 'F')] for s in secs); hm = sum(alive[(s, 'M')] for s in secs)
        out += f'<div class="s4-hall" role="group" aria-label="Hala {hall}"><div class="s4-hallhead"><span class="ip-cap">Hala {hall}</span>{pair(n(hf), n(hm))}</div><div class="ip-tiles">{"".join(stile(s, ng) for s in secs)}</div></div>'
    return out
def totals(ng=False):
    tf = f'<span class="ip-stotals__sub">{NOG}</span>' if ng else f'<span class="ip-stotals__sub s4-nowrap"><span>ghid <span class="ip-num">{gf_g}/{gm_g} g</span></span>{lamp("ok", "în normă")}</span>'
    tw = NOG if ng else 'ghid: – fără date în ghid'
    tl = NOG if ng else f'ghid {pct(lay_guide)} {lamp("ok", "în normă")}'
    return f'''<div class="ip-stotals" role="group" aria-label="Total fermă Hîrbovăț, 16.11.2026"><div><span class="ip-cap">Total fermă</span><span class="ip-stotals__sub">6 secțiuni</span><span class="ip-stotals__sub ip-num">16.11.2026 · SAPT {WK}</span></div><div><span class="ip-stotals__key">În viață</span><span class="ip-stotals__val">{pair(n(tot['F']), n(tot['M']))}</span><span class="ip-stotals__sub ip-num">total {n(tot['all'])}</span></div><div><span class="ip-stotals__key">{ico('mortality')}Mort. săpt.</span><span class="ip-stotals__val">{n(-(mort_wk_sex['F']+mort_wk_sex['M']))}</span><span class="ip-stotals__sub">{pair(n(-mort_wk_sex['F']), n(-mort_wk_sex['M']))}</span></div><div><span class="ip-stotals__key">{ico('feed')}Hrană azi</span><span class="ip-stotals__val">{pair(g_f, g_m)}<small>g</small></span>{tf}<span class="ip-stotals__sub s4-nowrap"><span>fără {F} H4.2 · {M} H3.2</span></span></div><div><span class="ip-stotals__key">{ico('weighing')}Greutate {F}</span><span class="ip-stotals__val">{hen_v(n(w_f))}<small>g</small></span><span class="ip-stotals__sub">{tw}</span><span class="ip-stotals__sub">medie pe găini</span></div><div><span class="ip-stotals__key">{ico('eggs')}Ouat</span><span class="ip-stotals__val">{n(lay_today,1)}<small>%</small></span><span class="ip-stotals__sub">{tl}</span></div></div>'''

# ---------- alerts ----------
h3 = st['lay_pct_week40_by_hall']['H3']; fd = st['feed_deviation_today']
h3_days = sorted((r for r in S['eggs_daily'] if r['hall'] == 'H3' and r['date'] <= TODAY and r['guide_lay_pct'] is not None), key=lambda r: r['date'])
h3_since = None
for r in reversed(h3_days):
    if r['lay_pct'] - r['guide_lay_pct'] < -3: h3_since = r['date']
    else: break
assert h3_since == '2026-11-09', h3_since
def alert(k, title, meta, status, act=None):
    a = f'<a class="ip-link ip-ev__act" href="{act[1]}">{act[0]}</a>' if act else ''
    return f'<div class="ip-ev ip-ev--alert">{tile(k)}<div class="ip-ev__main">{title}<div class="ip-ev__meta ip-num">{meta}</div></div>{a}<span class="ip-ev__status">{status}</span></div>'
def alerts(ng=False):
    a_h3 = alert('eggs', f'H3 · ouat <span class="ip-num">{pct(h3["lay_pct"])}</span> · ghid <span class="ip-num">{pct(lay_guide)}</span> · <span class="ip-num">{n(h3["diff_pts"],1)} pct</span>',
           f'16.11.2026 · SAPT {WK} · sub ghid din {ddate(h3_since)}', lamp('warn', 'abatere'), ('Vezi producția', '#productie'))
    a_h12 = alert('feed', f'H1.2 · hrană {roo(fd["g"], " g")} · ghid {roo(fd["guide_g"], " g")} · <span class="ip-num">+{pct(fd["deviation_pct"], 0)}</span>',
           f'16.11.2026 · SAPT {WK} · Dr. Ceban · ieri {roo(alloc[("H1.2","M")]["yesterday_g"], " g")}', lamp('warn', 'abatere'), ('Deschide secțiunea', '#sectiune-H1.2'))
    g32 = '' if ng else f' · ghid {roo(gm_g, " g")}'
    g42 = '' if ng else f' · ghid {hen(gf_g, " g")}'
    a_32 = alert('feed', f'H3.2 · hrană {M} azi – lipsă{g32}',
           f'16.11.2026 · SAPT {WK} · de la 12:00 · ieri {roo(alloc[("H3.2","M")]["yesterday_g"], " g")}', lamp('warn', 'lipsă'), ('Deschide secțiunea', '#sectiune-H3.2'))
    a_42 = alert('feed', f'H4.2 · hrană {F} azi – lipsă{g42}',
           f'16.11.2026 · SAPT {WK} · de la 12:00 · ieri {hen(alloc[("H4.2","F")]["yesterday_g"], " g")}', lamp('warn', 'lipsă'), ('Deschide secțiunea', '#sectiune-H4.2'))
    a_st = alert('delivery', f'Cocoși · stoc <span class="ip-num">{n(s_co["stock_kg"])} kg</span> · <span class="ip-num">{n(s_co["days_left"],1)} zile</span> · limită <span class="ip-num">5 zile</span>',
           f'16.11.2026 · se termină pe {ddate(s_co["run_out_date"])} · {s_co["avg_daily_kg_last3"]} kg/zi', lamp('warn', 'abatere'), ('Deschide stocul', '#hrana-stoc'))
    a_pr = alert('delivery', f'Hrană · {len(unpriced)} livrări fără preț · <span class="ip-num">{n(unpriced_kg)} kg</span>',
           f'{ddate(unpriced[0]["date"])[:5]}–{ddate(unpriced[-1]["date"])} · Ion Rusu · așteaptă prețul', lamp('info', 'informație'), ('Adaugă prețurile', '#preturi-hrana'))
    # no guide: the guide-based alerts (lay vs guide, feed ±10 %) go; missing rows, stock (D19 threshold) and prices stay
    return [a_32, a_42, a_st, a_pr] if ng else [a_h3, a_h12, a_32, a_42, a_st, a_pr]

# ---------- latest records ----------
# S4TP: a Mortalitate row counts deaths only (D81, as the Păsări Jurnal); culls are their own Sacrificare rows. The Zilnic
# save summary in event_log adds that day's culls (15.11: ♂ −4 = deaths −2 + Sacrificare H3.1 −2), asserted here.
def deaths(d):
    return {sx: sum(r['mortality'] for r in S['daily'] if r['date'] == d and r['sex'] == sx) for sx in 'FM'}
for _d, _e in ((e['at'][:10], e) for e in S['event_log'] if e['record'] == 'mortality' and e['action'] == 'create' and e['at'] >= '2026-11-15'):
    _c = {sx: sum(c['qty'] for c in S['culls'] if c['date'] == _d and c['sex'] == sx) for sx in 'FM'}
    assert f"♀ {n(-(deaths(_d)['F'] + _c['F']))} · ♂ {n(-(deaths(_d)['M'] + _c['M']))}" in _e['summary'], (_d, _e['summary'])
D15, D16 = deaths('2026-11-15'), deaths('2026-11-16')
assert (D15, D16) == ({'F': 7, 'M': 2}, {'F': 8, 'M': 1})
def ev(k, title, meta, status=''):
    return f'<a class="ip-ev" href="#inregistrare">{tile(k)}<div class="ip-ev__main">{title}<div class="ip-ev__meta ip-num">{meta}</div></div>{status}</a>'
events = [
 ev('weighing', f'Cântărire · 6 secțiuni · {pair(n(w_f), n(w_m), " g")}', f'16.11.2026 · 10:10 · SAPT {WK} · Dr. Ceban'),
 ev('feed', f'Hrană zilnică · 10 din 12 rânduri · <span class="ip-num">3{NB}691 kg</span>', f'16.11.2026 · 09:47 · SAPT {WK} · Dr. Ceban', lamp('warn', 'lipsă 2')),
 ev('eggs', f'Ouă · H1, H3, H4 · <span class="ip-num">18{NB}688</span> ouă', f'16.11.2026 · 08:35 · SAPT {WK} · Dr. Ceban'),
 ev('mortality', f'Mortalitate · 6 secțiuni · {hen(n(-D16["F"]))} · {roo(n(-D16["M"]))}', f'16.11.2026 · 08:34 · SAPT {WK} · Dr. Ceban'),
 ev('mortality', f'Corecție mortalitate · H4.1 · {F} · <span class="ip-num">14.11</span> · {hen_v(n(-3))} → {hen_v(n(-5))}', f'15.11.2026 · 10:12 · SAPT {WK} · Ion Rusu', '<span class="ip-tag ip-tag--neutral">corectat</span>'),
 ev('feed', f'Hrană zilnică · 12 din 12 rânduri · <span class="ip-num">4{NB}393 kg</span>', f'15.11.2026 · 09:47 · SAPT {WK} · Dr. Ceban'),
 ev('eggs', f'Ouă · H1, H3, H4 · <span class="ip-num">18{NB}679</span> ouă', f'15.11.2026 · 08:21 · SAPT {WK} · Dr. Ceban'),
 ev('cull', f'Sacrificare · H3.1 · {roo(n(-2))} · boală', f'15.11.2026 · 08:20 · SAPT {WK} · Dr. Ceban'),
 ev('mortality', f'Mortalitate · 6 secțiuni · {hen(n(-D15["F"]))} · {roo(n(-D15["M"]))}', f'15.11.2026 · 08:20 · SAPT {WK} · Dr. Ceban'),
]
for at in ['2026-11-16T10:10', '2026-11-16T09:47', '2026-11-16T08:35', '2026-11-16T08:34', '2026-11-15T10:12', '2026-11-15T09:47', '2026-11-15T08:21', '2026-11-15T08:20']:
    assert any(e['at'] == at for e in S['event_log']), at

def panel_default(ng=False, banner=''):
    al = alerts(ng)
    return f'''<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-rezumat">
{banner}<section class="s4-top" aria-labelledby="h-key"><h2 class="ip-sr" id="h-key">Indicatori{"" if ng else " față de ghid"}</h2>{readout(ng)}{stats(ng)}</section>
<section class="s4-block" aria-labelledby="h-sec"><div class="s4-head"><h2 class="ip-t-heading" id="h-sec">Secțiuni</h2><span class="ip-note ip-num">6 secțiuni · 3 hale · 16.11.2026</span></div><div class="s4-halls">{halls(ng)}</div>{totals(ng)}</section>
<div class="s4-split">
<section class="s4-block" aria-labelledby="h-chk"><div class="s4-head"><h2 class="ip-t-heading" id="h-chk">De verificat</h2><span class="ip-note"><span class="ip-num">{len(al)}</span> alerte · <span class="ip-num">16.11.2026</span></span></div><div class="ip-list">{"".join(al)}</div></section>
<section class="s4-block" aria-labelledby="h-log"><div class="s4-head"><h2 class="ip-t-heading" id="h-log">Ultimele înregistrări</h2><a class="ip-link ip-note" href="#pasari">Tot jurnalul</a></div><div class="ip-list">{"".join(events)}</div></section>
</div>
</div>'''

# ---------- 1. loading: static skeletons (no shimmer) where data arrives ----------
def sk(cls=''):
    return f'<span class="ip-skel{(" " + cls) if cls else ""}"></span>'
def panel_loading():
    stat = '<div class="ip-stat s4s-skstat">' + sk('s4s-w30') + sk('s4s-w60 s4s-h20') + sk('s4s-w80') + sk('s4s-w50') + '</div>'
    tile_ = '<div class="ip-stile s4s-sktile">' + '<div class="ip-stile__head">' + sk('s4s-w20 s4s-h16') + sk('s4s-w20') + '</div>' + sk('s4s-w50 s4s-h16') + sk('s4s-w100 s4s-h4') + ''.join('<div class="ip-stile__kv">' + sk('s4s-w40') + sk('s4s-w20') + '</div>' for _ in range(4)) + '</div>'
    hall = f'<div class="s4-hall"><div class="s4-hallhead">{sk("s4s-w20")}{sk("s4s-w30")}</div><div class="ip-tiles">{tile_}{tile_}</div></div>'
    row = f'<div class="ip-ev s4s-skrow">{sk("s4s-sq")}<div class="ip-ev__main s4s-skmain">{sk("s4s-w70")}{sk("s4s-w45")}</div>{sk("s4s-w15")}</div>'
    return f'''<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-rezumat" aria-busy="true">
<p class="ip-sr" role="status">Se încarcă rezumatul Hîrbovăț pentru ciclul curent.</p>
<section class="s4-top" aria-hidden="true"><div class="s4s-skreadout">{sk("s4s-w40")}{sk("s4s-w25 s4s-h28")}{sk("s4s-w50")}{sk("s4s-chart")}{sk("s4s-w60")}</div><div class="ip-stats s4-stats">{stat * 6}</div></section>
<section class="s4-block" aria-hidden="true"><div class="s4-head"><h2 class="ip-t-heading">Secțiuni</h2></div><div class="s4-halls">{hall * 3}</div><div class="s4s-sktotals">{sk("s4s-w100 s4s-h16")}{sk("s4s-w70")}{sk("s4s-w50")}</div></section>
<div class="s4-split" aria-hidden="true">
<section class="s4-block"><div class="s4-head"><h2 class="ip-t-heading">De verificat</h2></div><div class="ip-list">{row * 4}</div></section>
<section class="s4-block"><div class="s4-head"><h2 class="ip-t-heading">Ultimele înregistrări</h2></div><div class="ip-list">{row * 4}</div></section>
</div>
</div>'''

def empty_box(word, lines, step, aria='Rezumat'):
    body = ''.join(f'<p class="s4s-empty-p">{l}</p>' for l in lines)
    return f'<div class="s4s-emptybox" role="region" aria-label="{aria}"><div class="ip-empty">{lamp("off", word)}{body}{step}</div></div>'

# ---------- 2. no data in the period ----------
# check: no Seria 2026 record at Hîrbovăț in 01.11–15.11.2025 (cycle from 10.02.2026, arrivals 22–28.02.2026, sample from 01.03.2026)
TR = json.load(open(EX / 'transfers-seria2026.json', encoding='utf-8'))
assert all(a['date'] >= '2026-02-22' for a in TR['arrivals']), 'arrival before the cycle'
for k in ('daily', 'eggs_daily', 'weighings', 'deliveries', 'culls', 'sexing_errors', 'feed_by_type_daily'):
    assert all(r['date'] >= '2026-03-01' for r in S[k]), k
def panel_period():
    step = f'<button type="button" class="ip-btn ip-btn--secondary">Alege ciclul curent</button>'
    return f'''<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-rezumat">
{empty_box("fără date", [f"Ciclul Seria 2026 nu are înregistrări între 01.11.2025 și 15.11.2025; a început pe {ddate(c26['start_date'])}."], step, "Rezumat, interval 01.11.2025 – 15.11.2025")}
</div>'''

# ---------- 3. no guide ----------
BANNER_NG = f'<div class="ip-banner ip-banner--info s4s-banner" role="note">{icon("info", 16)}<div class="s4-grow">Ciclul nu are ghid. Graficele arată doar valorile efective.</div><a class="ip-link" href="#ciclu-ghid">Alege ghidul</a></div>'

# ---------- 4. no access ----------
def main_noaccess():
    step = f'<a class="ip-btn ip-btn--secondary" href="#acasa">{icon("home", 16)}Înapoi la Acasă</a>'
    return f'''<main class="s4-main">
{header('noaccess')}
{empty_box("fără acces", ["Nu ai acces la această fermă."], step, "Acces")}
</main>'''

# ---------- 5. empty farm (30.01.2026, after the halls were emptied on 28.01.2026) ----------
def main_empty():
    step = '<a class="ip-link" href="#ciclu-seria-2025">Vezi ultimul ciclu</a>'
    lines = [f'Ferma e goală din {ddate(PREV_END)}. Ultimul ciclu: Seria 2025, {PREV_WEEKS} săpt.', 'Ciclul pornește la prima populare sau transfer.']
    return f'''<main class="s4-main">
{header('empty')}
{cycle('empty')}
{empty_box("goală", lines, step, "Rezumat")}
</main>'''

# ---------- 6b. chosen past cycle: Hîrbovăț · Seria 2025 (D51), real previous-flock data only ----------
# eggs per hall and week (D18 totals, Jos apart), egg weight, hatching eggs; hens_alive and % lay are null (no flock count),
# no grams, no guide for eggs/week → anything else is "–" with its reason. Depopulation removals are not shown as mortality.
PD = json.load(open(EX / 'eggs-hirbovat-prev-daily.json', encoding='utf-8'))
PW = json.load(open(EX / 'eggs-hirbovat-prev-weekly.json', encoding='utf-8'))
assert PD['meta']['cycle_id'] == PW['meta']['cycle_id'] == c25['id']
assert all(r['hens_alive'] is None and r['lay_pct'] is None for r in PW['rows'])
PT = PD['meta']['totals']; PTH = PD['meta']['totals_by_hall']; PLAST = PD['meta']['hall_last_egg_day']
assert PT['total_eggs'] == 4606935 == sum(r['total_eggs'] for r in PD['rows'])
pall = {r['cycle_week']: r for r in PW['rows'] if r['hall'] == 'all'}
phall = {(r['hall'], r['cycle_week']): r for r in PW['rows'] if r['hall'] != 'all'}
PWEEKS = sorted(pall); PFIN = PWEEKS[-1]
assert PFIN == PREV_WEEKS == 47 and PWEEKS[0] == c25['data_start_cycle_week'] == 5
fin = pall[PFIN]
assert fin['partial'] and fin['days'] == 6 and fin['to'] == PREV_END
PDAYS = len({r['date'] for r in PD['rows']})
assert PDAYS == 300
fin_halls = [h for h in ('H1', 'H3', 'H4') if (h, PFIN) in phall]
assert fin_halls == ['H4'] and fin['total_eggs'] == phall[('H4', PFIN)]['total_eggs']
PGL = 'ghid: – fără date în ghid'
# D52/D54: mortality = Σ daily ♀/♂ deaths outside the emptying days (depopulation flag, first 15.01.2026); no flock count, so no %
PMORT = {sx: sum(r[f'mortality_{sx}'] for r in PD['rows'] if not r['depopulation']) for sx in 'FM'}
assert PMORT == {'F': 6948, 'M': 1505}, PMORT
PDEP = min(PD['meta']['depopulation_start'].values())
assert PDEP == '2026-01-15' == min(r['date'] for r in PD['rows'] if r['depopulation'])
NOFLOCK = 'fără efectiv în registru'

def readout_past():
    W0, W1 = 1, PFIN; PX0 = 56; YMAX = 160000
    def px(w): return round(PX0 + (w - W0) * (X1 - PX0) / (W1 - W0), 1)
    def py(v): return round(YB - v / YMAX * (YB - YT), 1)
    assert max(r['total_eggs'] for r in pall.values()) <= YMAX
    grid = ''
    for v in (0, 80000, 160000):
        grid += f'<line class="{"axis" if v == 0 else "grid"}" x1="{PX0}" x2="{X1}" y1="{py(v)}" y2="{py(v)}"/><text class="ax" x="{PX0-4}" y="{py(v)+3}" text-anchor="end">{n(v)}</text>'
    xl = ''.join(f'<text class="ax" x="{px(w)}" y="{CH-6}" text-anchor="middle">{w}</text>' for w in [1, 10, 20, 30, 40, PFIN])
    aline = ' '.join(f'{px(w)},{py(pall[w]["total_eggs"])}' for w in PWEEKS)
    lbl = f'Ouă pe săptămână la Hîrbovăț, Seria 2025, SAPT {PWEEKS[0]}–{PFIN}; fără date în ghid pentru ouă pe săptămână'
    chart = f'<svg class="ip-chart" viewBox="0 0 {CW} {CH}" role="group" aria-roledescription="grafic" tabindex="0" aria-label="{lbl}" aria-describedby="egg-help egg-read"><g aria-hidden="true">{grid}<polyline class="act" points="{aline}"/><line class="cur" x1="{px(PFIN)}" x2="{px(PFIN)}" y1="{YT-2}" y2="{YB}"/><circle class="pt" cx="{px(PFIN)}" cy="{py(fin["total_eggs"])}" r="3.5"/>{xl}</g></svg>'
    rows = [(w, n(pall[w]['total_eggs']), pall[w]['days'], '– fără date în ghid') for w in PWEEKS]
    per = f'{ddate(fin["from"])[:5]}–{ddate(fin["to"])}'
    return f'''<div class="ip-readout s4-readout">
<div class="ip-readout__cap"><span>OUĂ / SĂPT. · HÎRBOVĂȚ · SAPT {PFIN}</span><span class="ip-rlamp ip-rlamp--off">FĂRĂ DATE ÎN GHID</span></div>
<div class="ip-readout__val">{n(fin["total_eggs"])} <small>ouă</small></div>
<div class="ip-readout__sub">{per} · {fin["days"]} zile · doar H4 (H1, H3 fără ouă după {ddate(PLAST["H1"])[:5]})</div>
{chart}
<span class="ip-sr" id="egg-help">Săgețile stânga și dreapta mută săptămâna; Home și End: prima și ultima săptămână.</span>
<div class="ip-chart__read" id="egg-read" role="status" aria-live="polite">SAPT {PFIN} · {n(fin["total_eggs"])} ouă · {fin["days"]} zile · ghid – fără date în ghid</div>
{sr_table(f'Ouă pe săptămână la Hîrbovăț, Seria 2025; înregistrări din {ddate(c25["data_start_date"])} (SAPT {PWEEKS[0]})', ['SAPT', 'Ouă', 'Zile', 'Ghid'], rows)}
{chart_legend([('act', 'ouă totale (Cat. 1–3 + incubabile)')], 'axa: săpt. ciclului, 1–4 fără date')}
</div>'''

def stats_past():
    inc = PT['hatching_eggs'] / PT['total_eggs'] * 100
    jos = PT['floor_eggs'] / PT['total_eggs'] * 100
    w5 = pall[PWEEKS[0]]['avg_egg_weight_g']; wf = fin['avg_egg_weight_g']
    return guides(f'''<div class="ip-stats s4-stats">
<div class="ip-stat">{cap('eggs', 'Ouă în ciclu')}<span class="ip-stat__val">{n(PT['total_eggs'])}</span><span class="ip-stat__guide">{PGL}</span><span class="ip-stat__guide ip-num">{ddate(c25['data_start_date'])} – {ddate(PREV_END)} · {PDAYS} zile</span></div>
<div class="ip-stat">{cap('eggs', 'Incubabile')}<span class="ip-stat__val">{n(inc,1)}<small>%</small></span><span class="ip-stat__guide">{PGL}</span><span class="ip-stat__guide ip-num">{n(PT['hatching_eggs'])} din {n(PT['total_eggs'])} ouă</span></div>
<div class="ip-stat">{cap('eggs', f'Masă medie ou · SAPT {PFIN}')}<span class="ip-stat__val">{n(wf,1)}<small>g</small></span><span class="ip-stat__guide">{PGL}</span><span class="ip-stat__guide ip-num">SAPT {PWEEKS[0]}: {n(w5,1)} g</span></div>
<div class="ip-stat">{cap('eggs', 'Ouă jos')}<span class="ip-stat__val">{n(PT['floor_eggs'])}</span><span class="ip-stat__guide">{PGL}</span><span class="ip-stat__guide">nu intră în total · <span class="ip-num">{pct(jos)}</span> din total</span></div>
<div class="ip-stat">{cap('eggs', 'Ouat')}<span class="ip-stat__val">–</span><span class="ip-stat__guide">{NOFLOCK}, nu se calculează</span><span class="ip-stat__guide">ouă pe găină: – fără efectiv</span></div>
<div class="ip-stat">{cap('mortality', 'Mortalitate cumulată')}<span class="ip-stat__val">{pair(n(PMORT['F']), n(PMORT['M']))}</span><span class="ip-stat__guide">% – {NOFLOCK}</span><span class="ip-stat__guide"><span>fără golirea din <span class="ip-num">{ddate(PDEP)}</span></span></span></div>
</div>''')

ICOSLOT = '<span class="s4-ico" aria-hidden="true"><svg class="ip-icon ip-icon--14" viewBox="0 0 24 24"></svg></span>'  # empty slot, same width as ico()
def htile_past(h):
    last = max(w for (hh, w) in phall if hh == h)
    t = PTH[h]; inc = t['hatching_eggs'] / t['total_eggs'] * 100
    fv = f'<b>{n(phall[(h, PFIN)]["total_eggs"])}</b>' if (h, PFIN) in phall else f'<span class="s4-miss">– fără ouă după {ddate(PLAST[h])[:5]}</span>'
    prev = PFIN - 1
    return f'''<a href="#hala-{h}" class="ip-stile"><div class="ip-stile__head"><span class="ip-stile__code">{h}</span><span class="ip-note ip-num">ouă până pe {ddate(PLAST[h])}</span></div><div class="ip-stile__kv"><span class="s4-capi">{ICOSLOT}În viață</span><span class="s4-miss">– {NOFLOCK}</span></div><div class="ip-stile__kv"><span class="s4-capi">{ico('eggs')}Ouă SAPT {PFIN}</span>{fv}</div><div class="ip-stile__kv"><span class="s4-capi">{ico('eggs')}Ouă SAPT {prev}</span><b>{n(phall[(h, prev)]["total_eggs"])}</b></div><div class="ip-stile__kv"><span class="s4-capi">{ico('eggs')}Ouă în ciclu</span><b>{n(t["total_eggs"])}</b></div><div class="ip-stile__kv"><span class="s4-capi">{ico('eggs')}Incubabile</span><b>{pct(inc)}</b></div></a>'''

def events_past():
    by = {}
    for r in PD['rows']: by.setdefault(r['date'], []).append(r)
    out = []
    for d in sorted(by, reverse=True)[:8]:
        rs = sorted(by[d], key=lambda r: r['hall'])
        out.append(ev('eggs', f'Ouă · {", ".join(r["hall"] for r in rs)} · <span class="ip-num">{n(sum(r["total_eggs"] for r in rs))}</span> ouă', f'{ddate(d)} · SAPT {rs[0]["cycle_week"]} · din registru'))
    return out

def panel_past():
    none = f'<div class="ip-list"><div class="ip-empty s4s-none">{lamp("off", "nimic de verificat")}<span>Ciclul Seria 2025 la Hîrbovăț e închis din {ddate(PREV_END)}; nu are alerte.</span></div></div>'
    return f'''<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-rezumat">
<section class="s4-top" aria-labelledby="h-key"><h2 class="ip-sr" id="h-key">Indicatori, Seria 2025</h2>{readout_past()}{stats_past()}</section>
<section class="s4-block" aria-labelledby="h-sec"><div class="s4-head"><h2 class="ip-t-heading" id="h-sec">Hale</h2><span class="ip-note ip-num">3 hale · ouă pe hală din registru · {ddate(c25['data_start_date'])} – {ddate(PREV_END)}</span></div><div class="s4-halls">{"".join(htile_past(h) for h in ("H1", "H3", "H4"))}</div></section>
<div class="s4-split">
<section class="s4-block" aria-labelledby="h-chk"><div class="s4-head"><h2 class="ip-t-heading" id="h-chk">De verificat</h2><span class="ip-note">ciclu închis</span></div>{none}</section>
<section class="s4-block" aria-labelledby="h-log"><div class="s4-head"><h2 class="ip-t-heading" id="h-log">Ultimele înregistrări</h2><a class="ip-link ip-note" href="#pasari">Tot jurnalul</a></div><div class="ip-list">{"".join(events_past())}</div><p class="ip-note s4s-m0">Ultimele zile ale ciclului. Ieșirile de păsări din 15.01.2026 (golirea halelor) nu apar ca mortalitate.</p></section>
</div>
</div>'''

# ---------- 7. new cycle dialog ----------
DLG = f'''<div class="ip-scrim s4s-scrim"><div class="ip-dialog s4s-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-note">
<h2 id="dlg-t">Începe ciclu nou la Hîrbovăț</h2>
<p class="s4s-dlgnote" id="dlg-note">Ciclul curent (sapt. {WK}) se închide pe 16.11.2026. Săptămâna ghidului pornește de la 1.</p>
<div class="ip-field"><label class="ip-label" for="nc-date">Data de început</label><div class="ip-row s4s-daterow"><span class="ip-seg" role="radiogroup" aria-label="Zi rapidă"><span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Azi</span><span class="ip-seg__item" role="radio" aria-checked="false" tabindex="-1">Ieri</span></span><span class="ip-input s4s-date"><input id="nc-date" type="text" inputmode="numeric" value="16.11.2026" placeholder="zz.ll.aaaa" aria-describedby="nc-date-h"/><button type="button" class="ip-iconbtn" aria-label="Alege data de început din calendar" aria-haspopup="dialog" aria-expanded="false">{icon("cal", 16)}</button></span></div><span class="ip-hint ip-num" id="nc-date-h">SAPT 1 · ciclu nou din 16.11.2026</span></div>
<div class="ip-field"><label class="ip-label" for="nc-guide">Ghid</label><span class="ip-input"><select id="nc-guide" aria-describedby="nc-guide-h"><option selected="selected">Ross 308 ouat</option></select>{icon("chev", 16)}</span><span class="ip-hint" id="nc-guide-h">ghidul fermei · ouat</span></div>
<div class="ip-field"><label class="ip-label" for="nc-why">Motiv</label><span class="ip-input ip-input--area"><textarea id="nc-why" rows="3" required="required" aria-required="true" aria-describedby="nc-why-h"></textarea></span><span class="ip-hint" id="nc-why-h">Obligatoriu. Apare în jurnalul de modificări.</span></div>
<div class="ip-row s4s-dlgfoot"><button type="button" class="ip-btn ip-btn--secondary ip-focus" autofocus="autofocus">Anulează</button><button type="button" class="ip-btn ip-btn--primary">{icon("cycle", 16)}Începe ciclu nou</button></div>
</div></div>'''

def main_page(mode='default', panel=None, ng=False):
    fb = filterbar(mode if mode in ('period', 'past') else 'default')
    return f'''<main class="s4-main">
{header('past' if mode == 'past' else 'noguide' if ng else 'default')}
{cycle(mode if mode in ('menu', 'past') else 'default')}
<div class="s4-tight">{fb}{tabs}</div>
{panel}
</main>'''

CSS = '''body{margin:0}
.s4-shell{display:flex;align-items:stretch}
.s4-side{flex:none;min-height:0}
.s4-grow{flex:1}
.s4-account{display:flex;justify-content:space-between;align-items:center;padding:var(--space-2);border-top:var(--ip-hair) solid var(--line)}
.s4-account b{color:var(--ink);font-weight:500}
.s4-account .ip-link{display:inline-flex;align-items:center;justify-content:flex-end;min-width:var(--space-6);min-height:var(--space-6)} /* S3 fix 1: desktop text link keeps a ≥24px target (D46) */
.s4-main{flex:1;min-width:0;padding:var(--space-6);display:flex;flex-direction:column;gap:var(--space-6)}
.s4-tight{display:flex;flex-direction:column;gap:var(--space-2)}
.s4-panel{display:flex;flex-direction:column;gap:var(--space-6)}
.s4-lamps{display:flex;gap:var(--space-3)}
.s4-meta{gap:var(--space-2) var(--space-3)}
.s4-cycle{display:grid;grid-template-columns:calc(var(--space-12) * 7) minmax(0,1fr);gap:var(--space-6);align-items:start}
.s4-cycle > .s4-tight{padding-top:var(--space-2)} /* S4P: the track lane centres on the CycleSelector field */
.s4-cycsel{cursor:pointer}
.s4-top{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:var(--space-4);align-items:stretch}
.s4-readout{justify-content:flex-start}
.s4-stats{grid-template-columns:repeat(2,minmax(0,1fr))}
.s4-block{display:flex;flex-direction:column;gap:var(--space-3);min-width:0}
.s4-head{display:flex;justify-content:space-between;align-items:baseline;gap:var(--space-3)}
.s4-head h2{margin:0}
.s4-halls{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:var(--space-4)}
.s4-hall{display:flex;flex-direction:column;gap:var(--space-2);min-width:0}
.s4-hallhead{display:flex;justify-content:space-between;align-items:baseline;gap:var(--space-2);font-size:var(--ip-fs-small);line-height:var(--ip-lh-small)}
.s4-split{display:grid;grid-template-columns:minmax(0,7fr) minmax(0,5fr);gap:var(--space-4);align-items:start}
.s4-capi{display:flex;gap:var(--space-1);align-items:center}
.s4-ico{color:var(--c);display:inline-flex}
.s4-kvpair{gap:var(--space-1);font-size:inherit;line-height:inherit}
.s4-kvpair .ip-sex b{font:inherit}
.s4-unit{color:var(--ink-muted);font-family:var(--font-sans)}
.s4-miss{color:var(--ink-muted);font-family:var(--font-sans);margin-left:var(--space-1)}
.s4-nowrap{flex-wrap:wrap}
.s4-nowrap > *{white-space:nowrap} /* S4P: guide + lamp stay one line at 1440, the lamp drops under it when the cell is narrower (no overlap at 1280) */
.ip-ev__main > span.ip-num,.ip-ev__main > .ip-sex,.ip-ev__main > .ip-pair{white-space:nowrap} /* a figure never splits from its unit */
.s4-filterbar{row-gap:var(--space-2)}
.s4-chips{flex-basis:100%;display:flex;flex-wrap:wrap;gap:var(--space-2);align-items:center}
.s4s-rel{position:relative}
.s4s-rel > .s4-shell{min-height:inherit}'''
# S4 boards only (default, states, tabs): S3 and S11 reuse CSS/CSS_STATES and must stay byte-identical
CSS_S4 = '''.s4-head > .ip-link{display:inline-flex;align-items:center;min-height:var(--space-6);white-space:nowrap} /* S4TP: "Tot jurnalul" keeps a ≥24px target (D46), as the tab foot links */
.s4s-banner .ip-link{display:inline-flex;align-items:center;min-height:var(--space-6)} /* S4TP: "Alege ghidul" ≥24px (D46) */'''
# state boards only
CSS_STATES = '''.s4s-emptybox{background:var(--panel);border:var(--ip-hair) solid var(--line);border-radius:var(--radius-md)}
.s4s-empty-p{margin:0;color:var(--ink)}
.s4s-empty-p + .s4s-empty-p{color:var(--ink-secondary)}
.ip-empty > .ip-btn,.ip-empty > .ip-link{margin-top:var(--space-2)}
.s4s-banner{align-items:center}
.s4s-banner .ip-link{white-space:nowrap}
.s4s-skreadout{background:var(--panel);border:var(--ip-hair) solid var(--line);border-radius:var(--radius-md);padding:var(--space-4);display:flex;flex-direction:column;gap:var(--space-3)}
.s4s-skstat{gap:var(--space-2)}
.s4s-sktile{gap:var(--space-2)}
.s4s-sktotals{background:var(--panel);border:var(--ip-hair) solid var(--line);border-top-color:var(--line-strong);border-radius:var(--radius-md);padding:var(--space-3) var(--space-4);display:flex;flex-direction:column;gap:var(--space-2)}
.s4s-skrow{gap:var(--space-3);align-items:center}
.s4s-skmain{display:flex;flex-direction:column;gap:var(--space-2);flex:1}
.ip-skel{display:block}
.s4s-sq{width:var(--record-tile);height:var(--record-tile);flex:none;border-radius:var(--radius-sm)}
.s4s-chart{width:100%;height:calc(var(--space-12) * 4);border-radius:var(--radius-sm)}
.s4s-h4{height:var(--space-1)}.s4s-h16{height:var(--space-4)}.s4s-h20{height:var(--space-5)}.s4s-h28{height:calc(var(--space-6) + var(--space-1))}
.s4s-w15{width:15%}.s4s-w20{width:20%}.s4s-w25{width:25%}.s4s-w30{width:30%}.s4s-w40{width:40%}.s4s-w45{width:45%}.s4s-w50{width:50%}.s4s-w60{width:60%}.s4s-w70{width:70%}.s4s-w80{width:80%}.s4s-w100{width:100%}
.s4s-anchor{position:relative}
.s4s-cycmenu{position:absolute;top:calc(var(--space-5) + var(--control-h) + var(--space-1));left:0;right:0;z-index:3;min-width:0}
.s4s-list{list-style:none;margin:0;padding:0}
.s4s-mitem{width:100%;border:0;background:transparent;font:inherit;color:var(--ink);text-align:left;cursor:pointer;justify-content:flex-start;gap:var(--space-2)}
.s4s-gap{width:var(--space-4);flex:none}
.s4s-scrim{position:absolute;inset:0;border-radius:0;align-items:flex-start;padding-top:calc(var(--space-12) * 4);z-index:5}
.s4s-dialog{gap:var(--space-4)}
.s4s-dlgnote{margin:0;color:var(--ink-secondary)}
.s4s-daterow{gap:var(--space-2);flex-wrap:nowrap}
.s4s-date{flex:1;gap:var(--space-1);padding-right:var(--space-1)}
.s4s-date > input{font-family:var(--font-mono);font-variant-numeric:tabular-nums}
.s4s-date .ip-iconbtn{width:var(--space-6);height:var(--space-6)}
.ip-input > select + .ip-icon{color:var(--ink-secondary);flex:none}
.s4s-dlgfoot{justify-content:flex-end;gap:var(--space-2)}
.s4s-cycmenu[hidden]{display:none}
.s4s-none{padding:var(--space-6) var(--space-4)}
.s4s-m0{margin:0}'''

FONTS = '<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500&amp;family=IBM+Plex+Mono:wght@400;500&amp;display=swap">'
SYS = ROOT / 'design/system'
HEIGHTS = json.loads(pathlib.Path(__file__).with_name('s4s_heights.json').read_text()) if pathlib.Path(__file__).with_name('s4s_heights.json').exists() else {}
def write(stem, title, body, overlay='', H=None, css=None):
    H = H or HEIGHTS.get(stem, 1874)
    css = css if css is not None else CSS + '\n' + CSS_STATES + '\n' + CSS_S4
    root = f'<div class="ip-root s4s-rel" style="padding:0; min-height:{H}px">{body}{overlay}</div>'
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
{FONTS}
<style>{css}</style>
</helmet>
{root}
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":1440,"height":{H}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
    out = ROOT / f'design/canvas/project/{stem}.dc.html'
    out.write_text(dc, encoding='utf-8')
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{FONTS}<style>{css}</style></head><body>{root}</body></html>'''
    pathlib.Path('/tmp/s4s').mkdir(exist_ok=True)
    pathlib.Path(f'/tmp/s4s/{stem}.html').write_text(wrap, encoding='utf-8')
    print('ok', stem, H, len(dc))

def shell(main, viewer=False, inert=False, series='2026'):
    sb = sidebar('Observator', viewer=True) if viewer else sidebar(series=series)
    inner = f'{sb}{main}'
    return f'<div class="s4-shell"{" inert=\"inert\" aria-hidden=\"true\"" if inert else ""}>{inner}</div>'

B = 'S4-ferma-rezumat--'
BOARDS = [
 (B + 'se-incarca', 'S4 · Fermă · Rezumat · se încarcă', shell(main_page(panel=panel_loading())), ''),
 (B + 'fara-date-in-perioada', 'S4 · Fermă · Rezumat · fără date în perioadă', shell(main_page('period', panel=panel_period())), ''),
 (B + 'fara-ghid', 'S4 · Fermă · Rezumat · fără ghid', shell(main_page(panel=panel_default(True, BANNER_NG), ng=True)), ''),
 (B + 'fara-acces', 'S4 · Fermă · Rezumat · fără acces', shell(main_noaccess(), viewer=True), ''),
 (B + 'ferma-goala', 'S4 · Fermă · Rezumat · fermă goală', shell(main_empty()), ''),
 (B + 'ciclu-meniu', 'S4 · Fermă · Rezumat · ciclu · meniu', shell(main_page('menu', panel=panel_default())), ''),
 (B + 'ciclu-ales', 'S4 · Fermă · Rezumat · ciclu ales', shell(main_page('past', panel=panel_past()), series='2025'), ''),
 (B + 'ciclu-nou', 'S4 · Fermă · Rezumat · ciclu nou', shell(main_page(panel=panel_default()), inert=True), DLG),
]
if __name__ == '__main__':
    for stem, title, body, ov in BOARDS:
        write(stem, title, body, ov)
