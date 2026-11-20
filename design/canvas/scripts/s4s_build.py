"""Build the S4 state artboards (S4S) from the extracts. The shared shell/header/sections are functions with a no-guide
flag; % is written without a space (D46). Since S4P the default board (s4_build.py) is drawn by these same functions,
so the shell is identical on all 9 boards; the inferred-start mark and the split chip are system classes (additions.css).
python3 -I s4s_build.py <repo_root>
Writes design/canvas/project/S4-ferma-rezumat--<state>.dc.html and /tmp/s4s/<state>.html (plain wrappers)."""
import sys, re, json, pathlib, datetime as dt
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import skel as SKL  # FIN11G: §54 skeleton markup

ROOT = pathlib.Path(sys.argv[1])
EX = ROOT / 'design/data/extracts'
# RF01: every figure comes from realdata (DATA03 extracts, today 18.11.2026, D151–D158); formats live there too.
import realdata as R
from realdata import NB, MINUS, rnd, n, sgn, pct
CL, GL, CY, TR = R._CL, R.GL, R.CY, R.TR

def __getattr__(name):  # PEP 562: `S` is a legacy alias for rows not yet ported (RF00 §3); the S4 Rezumat boards never read it
    if name == 'S':
        globals()['S'] = json.load(open(EX / 'sample-hirbovat-2026.json', encoding='utf-8'))  # literal-ok: superseded file, legacy alias only
        return globals()['S']
    raise AttributeError(name)

# S16P fix 1: DROP = (date, section, sex) of a deleted Mortalitate record, set on the module before it runs (only the S16
# after-delete board does, through s4t_build). The sample it edited is superseded; the real-data version comes with the
# shared illustrative correction (D158a, RF08).
DROP = globals().get('DROP')
if DROP: raise NotImplementedError('s4s DROP: port the S16 delete onto realdata with the D158a correction (RF08)')
DROPQ = 0

def ddate(iso):
    return R.dd(iso)

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
 'chevup': '<path d="m18 15-6-6-6 6"/>',  # FIN10: an open Select (section 53)
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

# ---------- data (RF01: realdata, 18.11.2026) ----------
TODAY = R.TODAY_ISO; WK = R.WK                        # 2026-11-18, SAPT 59 (D151)
SECTIONS, HALLS = R.HSECS, R.HHALLS                   # H5.1 … H7.2, H5–H7 (D152)
alive = {(s, sx): R.alive(s, sx) for s in SECTIONS for sx in 'FM'}
tot = {'F': R.alive_farm('F'), 'M': R.alive_farm('M')}; tot['all'] = tot['F'] + tot['M']
W0 = R.week_dates(WK)[0]                              # week so far: 17.11–18.11.2026 (Tue–Mon weeks)
WSPAN = R.drange(W0, TODAY)
mort_wk = {s: R.deaths(W0, TODAY, sec=s) for s in SECTIONS}            # real deaths only (D155)
mort_wk_sex = {sx: R.deaths(W0, TODAY, sx=sx) for sx in 'FM'}
FULL = [s for s in SECTIONS if R.has_birds(s)]        # compartments with birds today
EMPTY_HALLS = {h: R.empty_since(h) for h in HALLS if R.empty_since(h)}  # H7: 14.11.2026
assert list(EMPTY_HALLS) == ['H7'] and FULL == ['H5.1', 'H5.2', 'H6.1', 'H6.2']
g_today = {(s, sx): R.feed_g(sec=s, sx=sx) for s in SECTIONS for sx in 'FM'}  # None: nothing fed (empty compartment)
g_f, g_m = rnd(R.feed_g(sx='F')), rnd(R.feed_g(sx='M'))
guide = {(r['age_week'], r['sex']): r for r in GL['rows']}           # by age week (D151)
gf_g, gm_g = R.guide_feed(WK, 'F'), R.guide_feed(WK, 'M')
def feed_dev(sec, sx):
    g = g_today[(sec, sx)]; return g is not None and abs(g / R.guide_feed(WK, sx) - 1) > 0.10
assert not any(feed_dev(s, sx) for s in FULL for sx in 'FM')  # 18.11: ♀ 160 g vs 157, ♂ 141–144 g vs 156 (−9,6 % at most)
# weighings: none at Hîrbovăț in the workbook; the illustrative H5.1 ones of 10.11.2026 (D152) are the latest
WDAY = max(w['date'] for w in R.WEIGHINGS_HIR)
wt = {(w['section'], w['sex']): w['avg_weight_g'] for w in R.WEIGHINGS_HIR if w['date'] == WDAY}
WSECS = sorted({s for s, _ in wt})
def wavg(d, sx):
    keys = [k for k in d if k[1] == sx and d[k] is not None]
    return sum(d[k] * alive[k] for k in keys) / sum(alive[k] for k in keys)
w_f = rnd(wavg(wt, 'F')); w_m = rnd(wavg(wt, 'M'))
assert (WDAY, WSECS, w_f, w_m) == ('2026-11-10', ['H5.1'], 3960, 4870)
lay_today = R.lay_day(); lay_guide = R.guide_lay(WK)  # D156: eggs ÷ hens at the start of the day
EGG_HALLS = [h for h in HALLS if R.eggs(TODAY, h)]
eggs_today = {h: {'total_eggs': R.eggs(TODAY, h), 'lay_pct': R.lay_day(TODAY, h), 'guide_lay_pct': lay_guide} for h in EGG_HALLS}
assert EGG_HALLS == ['H5', 'H6'] and sum(e['total_eggs'] for e in eggs_today.values()) == R.eggs() == 5914
lay_all = {k: {'lay_pct': R.lay_week(k)} for k in sorted(R.WEEKS) if R.lay_week(k) is not None}  # D156: eggs ÷ hen-days
def lay_dev(hall): return eggs_today[hall]['lay_pct'] - lay_guide < -3   # the ±3 pct band of the chart
stocks = {}
for _sx, _ft in R.FEED_TYPE.items():
    _left, _avg, _out = R.stock_days(_sx)
    stocks[_ft] = {'stock_kg': R.stock()[_sx], 'days_left': _left, 'avg_daily_kg_last3': _avg, 'run_out_date': _out.isoformat()}
STOCK_LIMIT = R.STOCK_LIMIT                            # D19: days of feed left (5)
unpriced = R.UNPRICED
unpriced_kg = sum(d['qty_kg'] for d in unpriced)
cum = {'start_counts': R.ARRIVED, 'deaths': R.COUNTS['cycle']['removals'], 'mortality_pct': {}}
for _sx in 'FM':
    _dead = R.deaths(R.CYCLE_START, TODAY, sx=_sx); _sold = R.sold(R.CYCLE_START, TODAY, sx=_sx)
    cum['mortality_pct'][_sx] = _dead / R.ARRIVED[_sx] * 100
    cum.setdefault('dead', {})[_sx] = _dead; cum.setdefault('sold', {})[_sx] = _sold
    assert R.ARRIVED[_sx] - _dead - _sold == tot[_sx]
assert cum['dead'] == {'F': 1521, 'M': 1048} and cum['sold'] == {'F': 12801, 'M': 1182}
c25, c26 = R.C25, R.C26
C26_START = c26['start_date']                         # 22.02.2026
PREV_END = c25['end_date']; PREV_WEEKS = c25['weeks']  # 2026-01-28; 47 weeks = SAPT 19–65
P25 = f'SAPT {c25["sapt_start"]}–{c25["sapt_end"]}'   # Seria 2025 by age week (D151)
assert (PREV_END, P25, C26_START) == ('2026-01-28', 'SAPT 19–65', '2026-02-22')

# ---------- sidebar ----------
def sidebar(role='Proprietar', viewer=False, series='2026'):
    s26 = series == '2026'
    cur = f'<span class="ip-select__val"><b>Seria 2026</b></span>{lamp("ok", "deschisă")}' if s26 else f'<span class="ip-select__val"><b>Seria 2025</b></span>{lamp("off", "închisă")}'
    nav =f'<a class="ip-nav" href="#acasa">{icon("home")}Acasă</a>'
    if not viewer: nav += f'<a class="ip-nav" href="#inregistreaza">{icon("plus")}Înregistrează</a>'
    nav += f'<a class="ip-nav is-current" href="#serie" aria-current="page">{icon("layers")}Serie</a><a class="ip-nav" href="#rapoarte">{icon("chart")}Rapoarte</a>'
    if not viewer: nav += f'<a class="ip-nav" href="#setari">{icon("settings")}Setări</a>'
    return f'''<nav class="ip-sidebar ip-shell__side" aria-label="Navigare">
<div class="ip-wordmark">Impex Line</div>
<div class="ip-select ip-series"><span class="ip-sr" id="ser-lbl">Serie</span><div class="ip-input ip-select__btn" id="ser-combo" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="false" aria-controls="ser-list" aria-labelledby="ser-lbl">{cur}{icon('chev', 16)}</div>
<ul class="ip-menu ip-select__menu" role="listbox" id="ser-list" aria-labelledby="ser-lbl" tabindex="-1" hidden="hidden"><li class="ip-menu__group" role="presentation">Serii</li><li class="ip-menu__item ip-select__opt" role="option" id="ser-2026" aria-selected="{'true' if s26 else 'false'}"><span class="ip-select__chk" aria-hidden="true">{icon('check', 16) if s26 else ''}</span><span class="ip-select__txt"><span class="ip-select__top"><b>Seria 2026</b>{lamp('ok', 'deschisă')}</span><span class="ip-num">din 07.10.2025</span></span></li><li class="ip-menu__item ip-select__opt" role="option" id="ser-2025" aria-selected="{'false' if s26 else 'true'}"><span class="ip-select__chk" aria-hidden="true">{'' if s26 else icon('check', 16)}</span><span class="ip-select__txt"><span class="ip-select__top"><b>Seria 2025</b>{lamp('off', 'închisă')}</span><span class="ip-num">până la 28.01.2026</span></span></li></ul></div>
{nav}
<span class="ip-grow"></span>
<div class="ip-note ip-sidebar__account"><b>{role}</b><a class="ip-link" href="#iesi">Ieși</a></div>
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
                f'<span class="ip-note">Început dedus <span class="ip-num">{ddate(c25["start_date"])}</span> · încheiat prin golire <span class="ip-num">{ddate(PREV_END)}</span> · <span class="ip-num">{P25}</span></span>'
                f'<span class="ip-note">efectiv – fără date în registru</span>')
        acts = '<a class="ip-btn ip-btn--secondary" href="#ciclu-seria-2026">Înapoi la ciclul curent</a>'
        return f'''<header class="ip-pagehead">
<div>
{crumbs('2025')}
<h1>Hîrbovăț</h1>
<div class="ip-row ip-pagehead__meta">{meta}</div>
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
        meta = f'<span class="ip-tag ip-tag--neutral">Ouat</span>{gtag}{pair(n(tot["F"]), n(tot["M"]))}<span class="ip-note ip-num">SAPT {WK} · ciclu din {ddate(C26_START)}</span>'
        acts = f'''<button type="button" class="ip-btn ip-btn--secondary">{icon('transfer', 16)}Mută păsări</button>
<button type="button" class="ip-btn ip-btn--secondary">{icon('sale', 16)}Vânzare</button>
<button type="button" class="ip-btn ip-btn--ghost" aria-haspopup="menu" aria-expanded="false" aria-label="Acțiuni ciclu">{icon('cycle', 16)}Ciclu{icon('chev', 14)}</button>'''
    return f'''<header class="ip-pagehead">
<div>
{CRUMBS}
<h1>Hîrbovăț</h1>
<div class="ip-row ip-pagehead__meta">{meta}</div>
</div>
<div class="ip-row">
{acts}
</div>
</header>'''

# ---------- cycle selector + farm track ----------
# FIN16: the farm track is drawn 1:1 at each width its box can have (additions.css §65 ip-trackbox, one svg per viewBox
# width): beside the CycleSelector 720 (1440) / 560 (1280), alone in the narrow page beside a SidePanel 720 / 560, and
# 1160 / 1000 for a page wider than 1440; so its text stays 11px and its blocks 30px tall (was one 800 drawing, scaled
# to 640 / 560 at 1280: 8.8 / 7.7px text, 24 / 21px blocks).
TWS = (1160, 1000, 720, 560)
DOM_LIVE = (dt.date(2025, 3, 1), dt.date(2027, 1, 31), [('2025-04-01', '04.25'), ('2025-07-01', '07.25'), ('2025-10-01', '10.25'), ('2026-01-01', '01.26'),
            ('2026-04-01', '04.26'), ('2026-07-01', '07.26'), ('2026-10-01', '10.26'), ('2027-01-01', '01.27')])
# empty farm on 30.01.2026: the axis ends near "azi" (01.2025 – 03.2026), no future months after a cycle that doesn't exist yet
DOM_EMPTY = (dt.date(2025, 1, 1), dt.date(2026, 3, 31), [('2025-01-01', '01.25'), ('2025-04-01', '04.25'), ('2025-07-01', '07.25'), ('2025-10-01', '10.25'),
             ('2026-01-01', '01.26'), ('2026-03-01', '03.26')])
def track(today_iso, sel='2026', dom=DOM_LIVE, TW=TWS[0]):
    """sel: the cycle shown on the page ('2026', '2025') or None (empty farm: nothing current). Seria 2026 exists only when today ≥ its start.
    TW: the viewBox width of this drawing (one of TWS)."""
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
    lab25 = f'Seria 2025, ouat, început dedus {ddate(c25["start_date"])} · încheiat prin golire {ddate(PREV_END)}, sapt. {c25["sapt_start"]}–{c25["sapt_end"]}, închis. {act25}'
    lab26 = f'Seria 2026, ouat, din {ddate(C26_START)}, deschis, azi sapt. {WK}. Început prin transfer. {act26}'
    w25 = round(x25b - x25a, 1)
    cur25 = ' aria-current="true"' if sel == '2025' else ''
    sel25 = f'<rect class="sel" x="{x25a}" y="37" width="{w25}" height="2"/>' if sel == '2025' else ''
    b25 = f'<a href="#ciclu-seria-2025"{cur25} aria-label="{lab25}"><title>{lab25}</title><rect class="ring" x="{round(x25a-4,1)}" y="12" width="{round(w25+8,1)}" height="30" rx="3"/><rect class="cyc" x="{x25a}" y="16" width="{w25}" height="18" rx="2"/><text class="cyctxt" x="{round(x25a+7,1)}" y="29">Seria 2025 · ouat · {P25}</text><line class="ded" x1="{round(x25a-3,1)}" x2="{round(x25a-3,1)}" y1="14" y2="40"/>{sel25}</a>'
    tl = ddate(today_iso)[:5]
    if with26:
        x26a = tx(c26['start_date']); w26 = round(xt - x26a, 1)
        cur26 = ' aria-current="true"' if sel == '2026' else ''
        sel26 = f'<rect class="sel" x="{x26a}" y="37" width="{w26}" height="2"/>' if sel == '2026' else ''
        b26 = f'<a href="#ciclu-seria-2026"{cur26} aria-label="{lab26}"><title>{lab26}</title><rect class="ring" x="{round(x26a-4,1)}" y="12" width="{round(w26+8,1)}" height="30" rx="3"/><rect class="cyc open" x="{x26a}" y="16" width="{w26}" height="18" rx="2"/><text class="cyctxt open" x="{round(x26a+7,1)}" y="29">Seria 2026 · ouat · din {ddate(C26_START)}</text><rect class="xfer" x="{round(x26a-1,1)}" y="16" width="3" height="18"/>{sel26}</a>'
        ro = f'<rect class="ro" x="{round(xt+4,1)}" y="17" width="56" height="16" rx="2"/><text class="rotxt" x="{round(xt+32,1)}" y="28.5" text-anchor="middle">SAPT {WK}</text>'  # 56: "SAPT nn" at 11px Mono (FIN13), as S3
        shown = {'2025': ' Afișat: Seria 2025.', '2026': ' Afișat: Seria 2026.'}.get(sel, '')
        desc = f'Hîrbovăț: Seria 2025 ouat, început dedus {ddate(c25["start_date"])} · încheiat prin golire {ddate(PREV_END)}, sapt. {c25["sapt_start"]}–{c25["sapt_end"]}, închis; Seria 2026 ouat din {ddate(C26_START)} (transfer), deschis, vânzări din {ddate(R.EMPTYING_SINCE)}, azi {ddate(today_iso)}, sapt. {WK}.{shown}'
    else:  # empty farm: no open cycle, so no readout next to "azi" and no current block
        b26 = ''; ro = ''
        desc = f'Hîrbovăț: Seria 2025 ouat, început dedus {ddate(c25["start_date"])} · încheiat prin golire {ddate(PREV_END)}, sapt. {c25["sapt_start"]}–{c25["sapt_end"]}, închis; ferma e goală din {ddate(PREV_END)}, azi {ddate(today_iso)}. Niciun ciclu deschis.'
    return f'''<svg class="ip-track ip-track--farm ip-track--{TW}" viewBox="0 0 {TW} 60" role="group" aria-label="Ciclurile fermei Hîrbovăț"><desc>{desc}</desc><rect class="lane" x="0" y="14" width="{TW}" height="26"/>{b25}{b26}<line class="today" x1="{xt}" x2="{xt}" y1="12" y2="42"/><text class="todaylbl" x="{xt}" y="9" text-anchor="middle">azi {tl}</text>{ro}{ticks}</svg>'''
def trackbox(today_iso, sel='2026', dom=DOM_LIVE):  # FIN16: §65 width box, the widest drawing that fits is shown
    return '<div class="ip-trackbox">' + ''.join(track(today_iso, sel, dom, w) for w in TWS) + '</div>'
def legend(xfer=True):  # names only the marks drawn: Seria 2026 starts by transfer; Seria 2025's start is inferred (D54), no sale
    x = '<span><i class="ip-rec ip-rec--transfer"></i>transfer</span>' if xfer else ''
    return f'<div class="ip-tracklegend" aria-hidden="true">{x}<span><svg class="ded" viewBox="0 0 3 12"><line x1="1.5" x2="1.5" y1="0" y2="12"/></svg>început dedus</span><span><i class="today"></i>azi</span></div>'
LEGEND = legend()
def opt26(selected, active=False):  # is-active only in an open list (section 53)
    return f'''<li class="ip-menu__item ip-select__opt{" is-active" if active else ""}" role="option" id="cyc-2026" aria-selected="{"true" if selected else "false"}"><span class="ip-select__chk" aria-hidden="true">{icon('check', 16) if selected else ""}</span><span class="ip-select__txt"><span class="ip-select__top"><b>Seria 2026 · ouat</b>{lamp('ok', 'deschis')}</span><span class="ip-num">din {ddate(C26_START)} · SAPT {WK}</span></span></li>'''
def opt25(selected, active=False):
    return f'''<li class="ip-menu__item ip-select__opt{" is-active" if active else ""}" role="option" id="cyc-2025" aria-selected="{"true" if selected else "false"}"><span class="ip-select__chk" aria-hidden="true">{icon('check', 16) if selected else ""}</span><span class="ip-select__txt"><span class="ip-select__top"><b>Seria 2025 · ouat</b>{lamp('off', 'închis')}</span><span class="ip-num">{R.drange(c25["start_date"], PREV_END)} · {P25}</span></span></li>'''
GRP = '<li class="ip-menu__group" role="presentation">Cicluri la Hîrbovăț</li>'
CYC_ITEMS = GRP + opt26(True) + opt25(False)
CYC_OPEN = GRP + opt26(True, True) + opt25(False)   # open: the chosen cycle is the active option (aria-activedescendant)
def hidden_list(items):  # collapsed combobox: its listbox exists (aria-controls) but is hidden, as the sidebar series selector
    return f'<ul class="ip-menu ip-select__menu" role="listbox" id="cyc-list" aria-labelledby="cyc-lbl" tabindex="-1" hidden="hidden">{items}</ul>'
def combo(text, lmp, opened=False, active=''):
    ad = f' aria-activedescendant="{active}"' if opened else ''
    return f'<div class="ip-input ip-select__btn{" is-focus" if opened else ""}" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="{"true" if opened else "false"}" aria-controls="cyc-list" aria-labelledby="cyc-lbl"{ad}><span class="ip-select__val">{text}</span>{lmp}{icon("chevup" if opened else "chev", 16)}</div>'
def cycle(mode='default'):
    if mode == 'empty':  # D51: no open cycle; Seria 2025 only as an option and in the hint; no block current on the track
        sel = combo('Niciun ciclu deschis', lamp('off', 'goală'))
        hint = f'<span class="ip-hint">ultimul ciclu · Seria 2025 · {R.drange(c25["start_date"], PREV_END)}</span>'
        return f'''<section class="s4-cycle" aria-label="Ciclu">
<div class="ip-field"><span class="ip-label" id="cyc-lbl">Ciclu</span><div class="ip-select">{sel}{hidden_list(GRP + opt25(False))}</div>{hint}</div>
<div class="ip-stack s4-tight">{trackbox('2026-01-30', sel=None, dom=DOM_EMPTY)}{legend(xfer=False)}</div>
</section>'''
    if mode == 'past':  # D51 "ciclu ales": the page rescoped to the closed Seria 2025 cycle
        sel = combo(f'Seria 2025 · {P25}', lamp('off', 'închis'))
        hint = f'<span class="ip-hint">început dedus {ddate(c25["start_date"])} · date din {ddate(c25["data_start_date"])}</span>'
        return f'''<section class="s4-cycle" aria-label="Ciclu">
<div class="ip-field"><span class="ip-label" id="cyc-lbl">Ciclu</span><div class="ip-select">{sel}{hidden_list(GRP + opt26(False) + opt25(True))}</div>{hint}</div>
<div class="ip-stack s4-tight">{trackbox(TODAY, sel='2025')}{LEGEND}</div>
</section>'''
    opened = mode == 'menu'
    sel = combo(f'Seria 2026 · din {ddate(C26_START)}', lamp('ok', 'deschis'), opened, 'cyc-2026')
    if opened:  # CycleSelector open: cycles (open first), then the manager+ actions (S4 is the owner's view)
        menu = f'''<div class="ip-menu ip-select__menu"><ul class="ip-select__list" role="listbox" id="cyc-list" aria-labelledby="cyc-lbl" tabindex="-1">{CYC_OPEN}</ul><div class="ip-menu__group" id="cyc-act">Ciclu</div><div role="group" aria-labelledby="cyc-act"><button type="button" class="ip-menu__item ip-select__act">{icon('cycle', 16)}<span class="ip-grow">Începe ciclu nou</span></button><button type="button" class="ip-menu__item ip-select__act"><span class="ip-select__chk" aria-hidden="true"></span><span class="ip-grow">Schimbă ghidul ciclului</span></button></div></div>'''
    else:
        menu = hidden_list(CYC_ITEMS)
    return f'''<section class="s4-cycle" aria-label="Ciclu">
<div class="ip-field"><span class="ip-label" id="cyc-lbl">Ciclu</span><div class="ip-select">{sel}{menu}</div><span class="ip-hint">2 cicluri la Hîrbovăț · deschis din {ddate(C26_START)}</span></div>
<div class="ip-stack s4-tight">{trackbox(TODAY)}{LEGEND}</div>
</section>'''

# ---------- filter bar + tabs ----------
def seg(label, items, sel):
    its = ''.join(f'<span class="ip-seg__item{" is-selected" if i == sel else ""}" role="radio" aria-checked="{"true" if i == sel else "false"}" tabindex="{0 if i == sel else -1}">{t}</span>' for i, t in enumerate(items))
    return f'<span class="ip-seg" role="radiogroup" aria-label="{label}">{its}</span>'
def chip(k, word, on=True):
    return f'<button type="button" class="ip-chip ip-rec ip-rec--{REC[k]}{"" if on else " ip-chip--off"}" role="checkbox" aria-checked="{"true" if on else "false"}">{icon(k, 16)}{word}</button>'
chip_sexerr = f'<button type="button" class="ip-chip ip-rec ip-chip--split" role="checkbox" aria-checked="true"><span class="ip-tile ip-tile--split" aria-hidden="true">{icon("sexerr", 16)}</span>Eroare sexare</button>'
# one chip per record type the cycle holds (D45): culls, a sexing error and weighings are illustrative (D152); transfers are the
# arrivals 22–28.02 and the 22.10 move; Vânzare = the real sales from 12.11 (D155)
for _k in ('culls', 'sexing_errors', 'weighings', 'sales'):
    assert R.ILL[_k] and all(C26_START <= e['date'] <= TODAY for e in R.ILL[_k]), _k
assert R.TR['internal_moves_hirbovat'] and R.sales()
CHIPS = f'<div class="s4-chips" role="group" aria-label="Tipuri de înregistrări">{chip("mortality", "Mortalitate")}{chip("cull", "Sacrificare")}{chip_sexerr}{chip("transfer", "Transfer")}{chip("sale", "Vânzare")}{chip("feed", "Hrană")}{chip("weighing", "Cântărire")}{chip("eggs", "Ouă")}</div>'
EXPORT = f'<button type="button" class="ip-btn ip-btn--ghost ip-filterbar__end">{icon("download", 16)}Exportă .xlsx</button>'
PERIODS = ['Ciclul curent', '4 săpt.', '13 săpt.', 'Interval']
def datef(i, label, val):
    return f'<span class="ip-datef"><label class="ip-label" for="{i}">{label}</label><span class="ip-input"><input id="{i}" type="text" inputmode="numeric" value="{val}" placeholder="zz.ll.aaaa" aria-describedby="{i}-fmt"/><button type="button" class="ip-iconbtn" aria-label="Alege data „{label}” din calendar" aria-haspopup="dialog" aria-expanded="false">{icon("cal", 16)}</button></span></span><span class="ip-sr" id="{i}-fmt">Format zz.ll.aaaa, cel târziu {ddate(TODAY)}</span>'
def filterbar(mode='default'):
    if mode == 'period':  # Interval 01.11–15.11.2025: no record of the cycle in it, so no type chips (D45) and nothing to export
        itv = f'<span class="ip-interval" role="group" aria-label="Interval">{datef("iv-a", "de la", "01.11.2025")}{datef("iv-b", "până la", "15.11.2025")}</span>'
        exp = f'<button type="button" class="ip-btn ip-btn--ghost ip-filterbar__end is-disabled" disabled="disabled">{icon("download", 16)}Exportă .xlsx</button>'
        return f'<div class="ip-filterbar s4-filterbar" role="group" aria-label="Filtre">{seg("Perioadă", PERIODS, 3)}{itv}{seg("Pas", ["Zi", "Săpt."], 0)}{exp}</div>'  # D166: grain Zi by default
    if mode == 'past':  # closed cycle: whole-cycle period; only the record types the previous flock has (eggs, mortality)
        chips = f'<div class="s4-chips" role="group" aria-label="Tipuri de înregistrări">{chip("mortality", "Mortalitate")}{chip("eggs", "Ouă")}</div>'
        return f'<div class="ip-filterbar s4-filterbar" role="group" aria-label="Filtre">{seg("Perioadă", ["Tot ciclul"] + PERIODS[1:], 0)}{seg("Pas", ["Zi", "Săpt."], 0)}{EXPORT}{chips}</div>'  # D166/D170: grain Zi
    return f'<div class="ip-filterbar s4-filterbar" role="group" aria-label="Filtre">{seg("Perioadă", PERIODS, 0)}{seg("Pas", ["Zi", "Săpt."], 1)}{EXPORT}{CHIPS}</div>'
TABS = [('rezumat', 'Rezumat'), ('pasari', 'Păsări'), ('hrana', 'Hrană'), ('greutate', 'Greutate'), ('productie', 'Producție')]
tabs = '<div class="ip-tabs" role="tablist" aria-label="Vederi Hîrbovăț">' + ''.join(
    f'<a class="ip-tab{" is-selected" if i == 0 else ""}" role="tab" id="tab-{k}" aria-controls="tp-ferma" aria-selected="{"true" if i == 0 else "false"}" tabindex="{0 if i == 0 else -1}" href="#{k}">{t}</a>'
    for i, (k, t) in enumerate(TABS)) + '</div>'

# ---------- readout + guide chart ----------
CW, CH = 540, 216  # S4P: drawn 1:1 in the Readout's 540px content box (no scaling); Readout = the 3 Stat rows (364)
X0, X1, YT, YB = 40, CW - 8, 10, CH - 22
WMIN, WMAX = c26['sapt_start'], max(R.GUIDE_LAY_WEEKS[-1], WK) + 3   # age weeks (D151): cycle start SAPT 20 … 65
assert (WMIN, WMAX) == (20, 65)
def cx(w): return round(X0 + (w - WMIN) * (X1 - X0) / (WMAX - WMIN), 1)
def cy(p): return round(YB - p / 100 * (YB - YT), 1)
gweeks = [w for w in R.GUIDE_LAY_WEEKS if WMIN <= w <= WMAX]                 # guide lay % SAPT 23–62
gv = {w: R.guide_lay(w) for w in gweeks}
aweeks = sorted(w for w in lay_all if w <= WK)                                # SAPT 21–59 (SAPT 59 = 17–18.11)
a40 = lay_all[WK]['lay_pct']; dev = rnd(a40 - gv[WK], 1)                     # current week so far (D156: eggs ÷ hen-days)
LAY_DEV = dev < -3                                                            # outside the ±3 pct band
XTICKS = [WMIN, 30, 40, 50, WK, WMAX]
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
    xl = ''.join(f'<text class="ax" x="{cx(w)}" y="{CH-6}" text-anchor="middle">{w}</text>' for w in XTICKS)
    gpart = '' if ng else f'<polygon class="band" points="{band}"/><polyline class="guide" points="{gline}"/>'
    lbl = f'Ouat Hîrbovăț pe săptămânile de vârstă, sapt. {aweeks[0]}–{WK}; ciclul nu are ghid' if ng else f'Ouat Hîrbovăț față de ghid, pe săptămânile de vârstă, sapt. {aweeks[0]}–{WK}; ghid sapt. {gweeks[0]}–{gweeks[-1]}'  # hidden running text: sapt. (D115, FIN13)
    chart = f'''<svg class="ip-chart" viewBox="0 0 {CW} {CH}" role="group" aria-roledescription="grafic" tabindex="0" aria-label="{lbl}" aria-describedby="lay-help lay-read"><g aria-hidden="true">{grid}{gpart}<polyline class="act" points="{aline}"/><line class="today" x1="{cx(WK)}" x2="{cx(WK)}" y1="{YT-2}" y2="{YB}"/><line class="cur" x1="{cx(WK)}" x2="{cx(WK)}" y1="{YT-2}" y2="{YB}"/><circle class="pt" cx="{cx(WK)}" cy="{cy(a40)}" r="3.5"/>{xl}</g></svg>'''
    if ng:
        rows = [(w, pct(lay_all[w]['lay_pct']), '– fără ghid') for w in aweeks]
        rl = '<span class="ip-rlamp ip-rlamp--off">FĂRĂ GHID</span>'
        sub = f'ciclul nu are ghid, nu se compară · {WSPAN}'
        read = f'SAPT {WK} · efectiv {pct(a40)} · ghid – fără ghid'
        leg = chart_legend([('act', 'efectiv')], 'axa: SAPT de vârstă')
    else:
        rows = [(w, pct(lay_all[w]['lay_pct']), pct(gv[w]) if w in gv else '– fără date în ghid') for w in aweeks]
        rl = '<span class="ip-rlamp ip-rlamp--warn">ABATERE</span>' if LAY_DEV else '<span class="ip-rlamp ip-rlamp--ok">ÎN NORMĂ</span>'
        sub = f'ghid {pct(gv[WK])} · {sgn(dev)} pct · {WSPAN}'
        read = f'SAPT {WK} · efectiv {pct(a40)} · ghid {pct(gv[WK])} · {sgn(dev)} pct'
        leg = chart_legend([('act', 'efectiv'), ('guide', 'ghid'), ('band', '±3 pct')], 'axa: SAPT de vârstă')  # = Producție
    return f'''<div class="ip-readout s4-readout">
<div class="ip-readout__cap"><span>OUAT · HÎRBOVĂȚ · SAPT {WK}</span>{rl}</div>
<div class="ip-readout__val">{n(a40,1)}<small>%</small></div>
<div class="ip-readout__sub">{sub}</div>
{chart}
<span class="ip-sr" id="lay-help">Săgețile stânga și dreapta mută săptămâna; Home și End: prima și ultima săptămână.</span>
<div class="ip-chart__read" id="lay-read" role="status" aria-live="polite">{read}</div>
{sr_table(f'Ouat Hîrbovăț pe săptămâni{"" if ng else ", efectiv față de ghid"}; ciclul din {ddate(C26_START)} (SAPT {WMIN})', ['SAPT', 'Efectiv', 'Ghid'], rows)}
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
FT_NAME = {k: R.FEED_TYPES[k]['name'] for k in stocks}                      # Breeder 2 / Cocoși (illustrative, D30)
ST_LOW, ST_HIGH = sorted(stocks, key=lambda k: stocks[k]['days_left'])       # the stock that runs out first leads
assert ST_LOW == 'ft-cocosi'
def st_lamp(k): return lamp('warn', 'abatere') if stocks[k]['days_left'] < STOCK_LIMIT else lamp('ok', 'în normă')
def lay_lamp(dev_pts): return lamp('warn', 'abatere') if dev_pts < -3 else lamp('ok', 'în normă')
MORT_G = R.guide_mort(WK)                                                     # ♀ cumulative mortality, RealGhid (7,25 % at SAPT 59)
assert MORT_G is not None and R.guide_row(WK, 'M')['mortality_cum_pct'] is None
assert R.weigh_guide(R.sapt(WDAY), 'F') is None and R.guide_row(R.sapt(WDAY) - 1)['body_weight_g'] is None  # D157: no weight guide at SAPT 57
EMPTY_NOTE = ' · '.join(f'{h} goală din {ddate(d)[:5]}' for h, d in EMPTY_HALLS.items())  # "H7 goală din 14.11"
NOG = 'ghid: – fără ghid'
def stats(ng=False):
    g_feed = NOG if ng else f'ghid {pair(gf_g, gm_g)} {lamp("ok", "în normă")}'
    g_lay = NOG if ng else f'ghid {pct(lay_guide)} · {sgn(rnd(lay_today-lay_guide,1))} pct {lay_lamp(lay_today - lay_guide)}'
    # D161 (S4N): the guide's cumulative mortality counts from lay start, ours from arrival; the line says so
    g_mort = NOG if ng else f'ghid {hen(pct(MORT_G, 2))} {lamp("ok", "în normă") if cum["mortality_pct"]["F"] <= MORT_G else lamp("warn", "abatere")}'
    g_none = NOG if ng else 'ghid: – fără date în ghid'
    lo, hi = stocks[ST_LOW], stocks[ST_HIGH]
    return guides(f'''<div class="ip-stats s4-stats">
<div class="ip-stat">{cap(None, 'În viață')}<span class="ip-stat__val">{pair(n(tot['F']), n(tot['M']))}</span><span class="ip-stat__guide">la sosire {pair(n(cum['start_counts']['F']), n(cum['start_counts']['M']))}</span><span class="ip-stat__guide">vânzări {pair(n(-cum['sold']['F']), n(-cum['sold']['M']))}</span></div>
<div class="ip-stat">{cap('mortality', 'Mortalitate cumulată')}<span class="ip-stat__val">{pair(n(cum['mortality_pct']['F'],2), n(cum['mortality_pct']['M'],2))}<small>%</small></span><span class="ip-stat__guide">{g_mort}</span>{"" if ng else '<span class="ip-stat__guide">(ghid din intrarea în ouat)</span>'}<span class="ip-stat__guide">din {ddate(C26_START)} {pair(n(-cum['dead']['F']), n(-cum['dead']['M']))}</span></div>
<div class="ip-stat">{cap('weighing', f'Greutate {F} / {M}')}<span class="ip-stat__val">{pair(n(w_f), n(w_m))}<small>g</small></span><span class="ip-stat__guide">{g_none}</span><span class="ip-stat__guide ip-num">cântărit {ddate(WDAY)[:5]} · {", ".join(WSECS)}</span></div>
<div class="ip-stat">{cap('feed', 'Hrană azi pe pasăre')}<span class="ip-stat__val">{pair(g_f, g_m)}<small>g</small></span><span class="ip-stat__guide">{g_feed}</span><span class="ip-stat__guide"><span>{len(FULL)} secțiuni · {EMPTY_NOTE}</span></span></div>
<div class="ip-stat">{cap('feed', 'Stoc hrană · zile rămase')}<span class="ip-stat__val">{n(lo['days_left'],1)}<small>zile · {FT_NAME[ST_LOW]}</small></span><span class="ip-stat__guide"><span class="ip-num">până pe {ddate(lo['run_out_date'])}</span>{st_lamp(ST_LOW)}</span><span class="ip-stat__guide"><span>{FT_NAME[ST_HIGH]} · <span class="ip-num">{n(hi['days_left'],1)} zile · {n(hi['stock_kg'])} kg</span></span></span></div>
<div class="ip-stat">{cap('eggs', 'Ouat azi')}<span class="ip-stat__val">{n(lay_today,1)}<small>%</small></span><span class="ip-stat__guide">{g_lay}</span><span class="ip-stat__guide ip-num">{n(sum(e['total_eggs'] for e in eggs_today.values()))} ouă · {len(EGG_HALLS)} hale</span></div>
</div>''')

# ---------- section tiles + totals ----------
def tstate(sec, ng=False):
    hall = sec.split('.')[0]; out = []
    if g_today[(sec, 'F')] is None or g_today[(sec, 'M')] is None: out.append(('warn', 'lipsă'))
    if ng: return out  # no guide: no deviation and no "în normă" (nothing to compare with); "lipsă" is not guide-based
    if any(feed_dev(sec, sx) for sx in 'FM') or abs(eggs_today[hall]['lay_pct'] - eggs_today[hall]['guide_lay_pct']) > 3: out.append(('warn', 'abatere'))
    return out or [('ok', 'în normă')]
def gpair(sec):
    f = g_today[(sec, 'F')]; m = g_today[(sec, 'M')]
    if f is None or m is None:
        miss = '<span class="s4-miss">– lipsă</span>'
        fs = F + miss if f is None else hen(f) + '<span class="s4-unit">g</span>'
        ms = M + miss if m is None else roo(m) + '<span class="s4-unit">g</span>'
        return f'<span class="ip-pair s4-kvpair">{fs}<span class="sep">·</span>{ms}</span>'
    return f'<span class="ip-pair s4-kvpair">{hen(f)}<span class="sep">·</span>{roo(m)}<span class="s4-unit">g</span></span>'
def wkv(sec):  # the latest weighing of the cycle (10.11, H5.1 only); the other compartments: – with its reason
    if (sec, 'F') in wt: return f'<span class="ip-pair s4-kvpair">{hen_v(n(wt[(sec, "F")]))}<span class="s4-unit">g · {ddate(WDAY)[:5]}</span></span>'
    return '<span class="s4-miss">– necântărit</span>'
def stile(sec, ng=False):
    hall = sec.split('.')[0]; f, m = alive[(sec, 'F')], alive[(sec, 'M')]
    pf = round(f / (f + m) * 100)
    missing = g_today[(sec, 'F')] is None or g_today[(sec, 'M')] is None
    lamps = ''.join(lamp(k, w) for k, w in tstate(sec, ng))
    okey = f'Ouat {hall}' if ng else f'Ouat {hall} · ghid {pct(eggs_today[hall]["guide_lay_pct"])}'
    return f'''<a href="#sectiune-{sec}" class="ip-stile{" is-missing" if missing else ""}"><div class="ip-stile__head"><span class="ip-stile__code">{sec}</span><span class="s4-lamps">{lamps}</span></div>{pair(n(f), n(m))}<div class="ip-bar" aria-hidden="true"><span style="width:{pf}%"></span><span style="width:{100-pf}%"></span></div><div class="ip-stile__kv"><span class="s4-capi">{ico('mortality')}Mort. săpt.</span><b>{n(-mort_wk[sec])}</b></div><div class="ip-stile__kv"><span class="s4-capi">{ico('feed')}Hrană azi</span>{gpair(sec)}</div><div class="ip-stile__kv"><span class="s4-capi">{ico('weighing')}Greutate {F}</span>{wkv(sec)}</div><div class="ip-stile__kv"><span class="s4-capi">{ico('eggs')}{okey}</span><b>{pct(eggs_today[hall]['lay_pct'])}</b></div></a>'''
def halls(ng=False):
    out = ''
    for hall in HALLS:
        secs = R.HALL_SECS[hall]
        hf = sum(alive[(s, 'F')] for s in secs); hm = sum(alive[(s, 'M')] for s in secs)
        if hall in EMPTY_HALLS:  # D158(d): an empty hall shows one muted line, not hidden and not as tiles of zeros
            body = f'<p class="ip-note ip-m0">{lamp("off", "goală")} din {ddate(EMPTY_HALLS[hall])} · {", ".join(secs)}</p>'
        else:
            body = f'<div class="ip-tiles">{"".join(stile(s, ng) for s in secs)}</div>'
        out += f'<div class="s4-hall" role="group" aria-label="Hala {hall}"><div class="s4-hallhead"><span class="ip-cap">Hala {hall}</span>{pair(n(hf), n(hm))}</div>{body}</div>'
    return out
def totals(ng=False):
    tf = f'<span class="ip-stotals__sub">{NOG}</span>' if ng else f'<span class="ip-stotals__sub s4-nowrap"><span>ghid <span class="ip-num">{gf_g}/{gm_g} g</span></span>{lamp("ok", "în normă")}</span>'
    assert not any(feed_dev(s, sx) for s in FULL for sx in 'FM')
    tw = NOG if ng else 'ghid: – fără date în ghid'
    tl = NOG if ng else f'ghid {pct(lay_guide)} {lay_lamp(lay_today - lay_guide)}'
    return f'''<div class="ip-stotals" role="group" aria-label="Total fermă Hîrbovăț, {ddate(TODAY)}"><div><span class="ip-cap">Total fermă</span><span class="ip-stotals__sub">{len(FULL)} secțiuni cu păsări</span><span class="ip-stotals__sub ip-num">{ddate(TODAY)} · SAPT {WK}</span></div><div><span class="ip-stotals__key">În viață</span><span class="ip-stotals__val">{pair(n(tot['F']), n(tot['M']))}</span><span class="ip-stotals__sub ip-num">total {n(tot['all'])}</span></div><div><span class="ip-stotals__key">{ico('mortality')}Mort. săpt.</span><span class="ip-stotals__val">{n(-(mort_wk_sex['F']+mort_wk_sex['M']))}</span><span class="ip-stotals__sub">{pair(n(-mort_wk_sex['F']), n(-mort_wk_sex['M']))}</span></div><div><span class="ip-stotals__key">{ico('feed')}Hrană azi</span><span class="ip-stotals__val">{pair(g_f, g_m)}<small>g</small></span>{tf}<span class="ip-stotals__sub s4-nowrap"><span>{EMPTY_NOTE}</span></span></div><div><span class="ip-stotals__key">{ico('weighing')}Greutate {F}</span><span class="ip-stotals__val">{hen_v(n(w_f))}<small>g</small></span><span class="ip-stotals__sub">{tw}</span><span class="ip-stotals__sub ip-num">{", ".join(WSECS)} · {ddate(WDAY)[:5]}</span></div><div><span class="ip-stotals__key">{ico('eggs')}Ouat</span><span class="ip-stotals__val">{n(lay_today,1)}<small>%</small></span><span class="ip-stotals__sub">{tl}</span></div></div>'''

# ---------- alerts: the farm's one alert set (realdata.alerts, D181), drawn here for S4 Rezumat "De verificat" and for S2
# (s2_build takes its rows through alert_rows). 18.11.2026: H5, H6 lay below the guide band; 13.11 H7.1/H7.2 ♀ feed lipsă
# (C9 window); no feed deviation, no stock under 5 days; 4 Hîrbovăț deliveries without a price. ----------
ALERTS = R.alerts('hirbovat', TODAY)
assert [(a['kind'], a.get('hall') or a.get('section')) for a in ALERTS] == [('lay', 'H5'), ('lay', 'H6'), ('missing', 'H7.1'), ('missing', 'H7.2'), ('unpriced', None)]
def below_since(hall): return R.lay_below_since(hall, TODAY)
LAY_ALERTS = [a['hall'] for a in ALERTS if a['kind'] == 'lay']
assert all(lay_dev(h) for h in LAY_ALERTS) and (below_since('H5'), below_since('H6')) == (dt.date(2026, 11, 1), dt.date(2026, 11, 1))  # D184 (RF14): H5 from 01.11 on the Productivitate guide
LAY_SINCE = {h: below_since(h) for h in LAY_ALERTS}
h3_since = LAY_SINCE['H6']  # literal-ok: s4t's old name for this hall (= H6, D152)
_YDAY = R.D(TODAY) - dt.timedelta(days=1)
alloc = {(s, sx): {'section': s, 'sex': sx, 'today_g': g_today[(s, sx)], 'yesterday_g': R.feed_g(_YDAY, sx, s), 'guide_g': R.guide_feed(WK, sx)}
         for s in FULL for sx in 'FM'}  # the day's norm g/bird per compartment (was the sample's allocations_today)
MISSING = [(a['section'], a['sex']) for a in ALERTS if a['kind'] == 'missing' and a['date'] == TODAY]   # today's lipsă rows (none on 18.11)
assert not MISSING and STOCK_LIMIT == R.STOCK_LIMIT
def alert(k, title, meta, status, act=None):
    a = f'<a class="ip-link ip-ev__act" href="{act[1]}">{act[0]}</a>' if act else ''
    return f'<li class="ip-ev ip-ev--alert">{tile(k)}<div class="ip-ev__main">{title}<div class="ip-ev__meta ip-num">{meta}</div></div>{a}<span class="ip-ev__status">{status}</span></li>'  # FIN16: a row of ul.ip-list
def alert_row(a, role='owner'):
    """one AlertRow of realdata.alerts, the same wording on S2 and S4. role changes only the link: the vet's past-day lipsă row
    is locked ("blocat · o completează managerul", D65); every other row and role opens its target."""
    k, d = a['kind'], a['date']
    if k == 'lay':
        return alert('eggs', f'{a["hall"]} · ouat <span class="ip-num">{pct(a["lay"])}</span> · ghid <span class="ip-num">{pct(a["guide_pct"])}</span> · <span class="ip-num">{n(a["lay"] - a["guide_pct"], 1)} pct</span>',
                     f'{ddate(d)} · SAPT {R.sapt(d)} · sub ghid din {ddate(a["since"])}', lamp('warn', 'abatere'), ('Vezi producția', '#productie'))
    if k == 'feed':
        sxf = hen if a['sex'] == 'F' else roo
        return alert('feed', f'{a["section"]} · hrană {sxf(n(a["g"]), " g")} · ghid {sxf(a["guide_g"], " g")} · <span class="ip-num">{sgn((a["g"] / a["guide_g"] - 1) * 100, 0)}%</span>',
                     f'{ddate(d)} · SAPT {R.sapt(d)}', lamp('warn', 'abatere'), ('Deschide secțiunea', f'#sectiune-{a["section"]}'))
    if k == 'missing':
        azi = ' azi' if d == TODAY else ''
        row = alert('feed', f'{a["section"]} · hrană {sg(a["sex"])}{azi} – lipsă', f'{ddate(d)} · SAPT {R.sapt(d)}', lamp('warn', 'lipsă'), ('Deschide secțiunea', f'#sectiune-{a["section"]}'))
        if role == 'vet' and d != TODAY:
            row = re.sub(r'<a class="ip-link ip-ev__act"[^>]*>.*?</a>', '<span class="ip-note ip-ev__act">blocat · o completează managerul</span>', row, count=1)
        return row
    if k == 'stock':
        ft = R.FEED_TYPE[a['sex']]
        return alert('delivery', f'{FT_NAME[ft]} · stoc <span class="ip-num">{n(a["kg"])} kg</span> · <span class="ip-num">{n(a["days"],1)} zile</span> · limită <span class="ip-num">{STOCK_LIMIT} zile</span>',
                     f'{ddate(d)} · se termină pe {ddate(a["run_out"])} · {n(a["avg"])} kg/zi', lamp('warn', 'abatere'), ('Deschide stocul', '#hrana-stoc'))
    if k == 'unpriced':
        return alert('delivery', f'Hrană · {a["n"]} livrări fără preț · <span class="ip-num">{n(a["kg"])} kg</span>',
                     f'{R.drange(a["first"], a["last"])} · așteaptă prețul', lamp('info', 'informație'), ('Adaugă prețurile', '#preturi-hrana'))
    raise ValueError(a)
def alert_rows(role='owner', ng=False, kinds=None):
    """the farm's AlertRows for a role: no guide-based rows without a guide; "fără preț" only for roles that see prices
    (owner, manager: D64); kinds limits the set (S2: the manager's lipsă rows sit in Lipsesc date, the owner's fără preț in its block)."""
    return [alert_row(a, role) for a in ALERTS if not (ng and a['guide']) and not (a['kind'] == 'unpriced' and role in ('viewer', 'vet'))
            and (kinds is None or a['kind'] in kinds)]
def alerts(ng=False):  # S4 Rezumat "De verificat" (owner): the whole farm set, 5 rows on 18.11
    return alert_rows('owner', ng)
assert len(unpriced) == 4 and unpriced_kg == 18160

# ---------- latest records: the illustrative event log on the real daily figures (D152, D155) ----------
# A Mortalitate row counts real deaths only (D81, D155); sales are their own Vânzare rows (real quantities, illustrative
# author/time). Each row's figures are recomputed from realdata and checked against the log's summary.
def ev(k, title, meta, status=''):  # FIN16: the row link in its own li (ul.ip-list > li > a.ip-ev, additions.css section 1), as the tabs' lists
    return f'<li><a class="ip-ev" href="#inregistrare">{tile(k)}<div class="ip-ev__main">{title}<div class="ip-ev__meta ip-num">{meta}</div></div>{status}</a></li>'
def sexpair(f, m):  # removals with true minus; a sex with nothing removed is left out of a sale
    return ' · '.join(x for x in (hen(n(-f)) if f else '', roo(n(-m)) if m else '') if x)
def event_row(e):
    d = e['at'][:10]; meta = f'{R.ddt(e["at"])} · SAPT {R.sapt(d)} · {e["user"]}'
    if e['record'] == 'mortality':
        f, m = R.deaths(d, sx='F'), R.deaths(d, sx='M'); k = sum(1 for s in SECTIONS if R.sd(d, s, 'F') or R.sd(d, s, 'M'))
        assert f'mortalitate ♀ {f} · ♂ {m}' in e['summary'] and f'{k} secțiuni' in e['summary'], e
        return ev('mortality', f'Mortalitate · {k} secțiuni · {hen(n(-f))} · {roo(n(-m))}', meta)
    if e['record'] == 'eggs':
        hs = [h for h in HALLS if R.eggs(d, h)]; t = R.eggs(d)
        assert e['summary'].startswith(f'Ouă · {t} total'), e
        return ev('eggs', f'Ouă · {", ".join(hs)} · <span class="ip-num">{n(t)}</span> ouă', meta)
    if e['record'] == 'feed':
        rows = [(s, sx) for s in SECTIONS for sx in 'FM' if R.sd(d, s, sx) and R.sd(d, s, sx)['kg']]
        due = [(s, sx) for s in SECTIONS for sx in 'FM' if R.alive_start(s, sx, d)]
        kg = R.feed_kg(d, 'F') + R.feed_kg(d, 'M')
        assert e['summary'] == f'Hrană · ♀ {R.feed_kg(d, "F"):g} kg · ♂ {R.feed_kg(d, "M"):g} kg', e
        miss = '' if len(rows) == len(due) else lamp('warn', f'lipsă {len(due) - len(rows)}')
        return ev('feed', f'Hrană zilnică · {len(rows)} din {len(due)} rânduri · <span class="ip-num">{n(kg, 1)} kg</span>', meta, miss)  # D181: the record's kg, one decimal (as S2/S15)
    if e['record'] == 'sale':
        return ev('sale', f'Vânzare · {R.sale_codes(e)} · {sexpair(e["F"], e["M"])}', meta)  # DATA04 (D168): one record per sale form
    raise ValueError(e)
EVENTS = R.EVENTS[:9]
assert [((R.D(e['at']) - R.TODAY).days, e['at'][11:]) for e in EVENTS][::4] == [(0, '13:16'), (-1, '11:42'), (-2, '13:25')]  # DATA04: 18.11 sale = one record
events = [event_row(e) for e in EVENTS]

# ---------- Rezumat PeriodTable (S4N, D154/D165; S4NG, D167): the shared Rezumat page draws its table here, so every board
# that shows the page (S4 default and states, the S9/S10 desktop backdrops) gets the same table; s4t_build re-exports these.
import importlib.util as _ilu
_spec14 = _ilu.spec_from_file_location('sys14_build', ROOT / 'design/system/scripts/sys14_build.py')
P = _ilu.module_from_spec(_spec14); _spec14.loader.exec_module(P)   # PeriodTable builders (no file written on import)
assert P.R is R
PWEEKS4, P_FROM, P_TO, SER = P.WEEKS, P.P_FROM, P.P_TO, P.SER
assert PWEEKS4 == [WK, WK - 1, WK - 2, WK - 3] and P_FROM == R.week_dates(WK - 3)[0] and P_TO == R.TODAY
def inp(d): return P_FROM <= R.D(d) <= P_TO
PCULLS = [c for c in R.ILL['culls'] if inp(c['date'])]                 # illustrative (D152): 11.11 H5.2 ♂ 1
PSEXERR = [e for e in R.ILL['sexing_errors'] if inp(e['date'])]       # none in SAPT 56–59
PMOVES = [m for m in R.TR['internal_moves_hirbovat'] if inp(m['date'])] + [a for a in R.TR['arrivals'] if inp(a['date'])]
PSALES = R.sales(P_FROM, P_TO)
PWEIGH = [w for w in R.WEIGHINGS_HIR if inp(w['date'])]
assert [(c['date'], c['section'], c['sex'], c['qty']) for c in PCULLS] == [('2026-11-11', 'H5.2', 'M', 1)]
assert not PSEXERR and not PMOVES and PSALES and PWEIGH  # Eroare sexare / Transfer / Populare: no records → no chip, no columns (D45)
# D45: the FilterBar lists the record types with records in the period (Populare: Hîrbovăț is populated by transfer, never placed)
PTYPES = [k for k, has in (('mortality', R.deaths(P_FROM, P_TO) > 0), ('cull', PCULLS), ('sexerr', PSEXERR), ('transfer', PMOVES),
                           ('sale', PSALES), ('feed', R.feed_kg(P_TO, 'F') > 0), ('weighing', PWEIGH), ('eggs', R.eggs() > 0)) if has]
assert PTYPES == ['mortality', 'cull', 'sale', 'feed', 'weighing', 'eggs']   # the record types of the 4 weeks on screen
# D191 (SYS18): the FilterBar period defaults to "Ciclul curent" (22.02–18.11.2026, SAPT 20–59); the table opens on its last 4
# weeks with "Încarcă încă 4 săptămâni" (D190). D45: the chips (and the tables' record groups) are the types with records in
# the chosen period, so the columns stay put while weeks load: the cycle adds Eroare sexare (19.05, illustrative) and Transfer
# (the arrivals 22–28.02, the H7.2 → H7.1 move of 22.10)
CYC_FROM = R.D(C26_START)
def incy(d): return CYC_FROM <= R.D(d) <= P_TO
CCULLS = [c for c in R.ILL['culls'] if incy(c['date'])]
CSEXERR = [e for e in R.ILL['sexing_errors'] if incy(e['date'])]
CMOVES = [m for m in R.TR['internal_moves_hirbovat'] if incy(m['date'])] + [a for a in R.TR['arrivals'] if incy(a['date'])]
CWEIGH = [w for w in R.WEIGHINGS_HIR if incy(w['date'])]
CTYPES = [k for k, has in (('mortality', R.deaths(CYC_FROM, P_TO) > 0), ('cull', CCULLS), ('sexerr', CSEXERR), ('transfer', CMOVES),
                           ('sale', R.sales(CYC_FROM, P_TO)), ('feed', R.feed_kg(P_TO, 'F') > 0), ('weighing', CWEIGH), ('eggs', R.eggs() > 0)) if has]
assert CTYPES == ['mortality', 'cull', 'sexerr', 'transfer', 'sale', 'feed', 'weighing', 'eggs']
CYC_WEEKS = list(range(WK, c26['sapt_start'] - 1, -1))   # SAPT 59 … 20: 40 weeks (D190 count)
assert len(CYC_WEEKS) == 40 and CYC_WEEKS[-1] == P.CYCLE_FIRST == 20
W8 = [WK - i for i in range(8)]                             # after one press: SAPT 59–52 (SYS18 state boards)
CHIPW = {'mortality': 'Mortalitate', 'cull': 'Sacrificare', 'sexerr': 'Eroare sexare', 'transfer': 'Transfer', 'sale': 'Vânzare',
         'feed': 'Hrană', 'weighing': 'Cântărire', 'eggs': 'Ouă'}
def chips4(on, types=None):
    out = ''
    for k in (types or CTYPES):   # D191: the period's types (D45), the cycle by default
        if k == 'sexerr':
            out += (f'<button type="button" class="ip-chip ip-rec ip-chip--split{"" if k in on else " ip-chip--off"}" role="checkbox" aria-checked="{"true" if k in on else "false"}">'
                    f'<span class="ip-tile ip-tile--split" aria-hidden="true">{icon("sexerr", 16)}</span>Eroare sexare</button>')
        else:
            out += chip(k, CHIPW[k], k in on)
    return f'<div class="s4-chips" role="group" aria-label="Tipuri de înregistrări">{out}</div>'
def filterbar4():
    """D191 (SYS18): period "Ciclul curent" (the table opens on its last 4 weeks, SAPT 56–59, and loads 4 more per press, D190;
    supersedes D159's "4 săpt." default); grain Zi (D166: every week open, each key can close its week; Săpt. = week rows
    only); chips = types of the period (D45). D165: the tag state is shared across tabs, all on by default."""
    return f'<div class="ip-filterbar s4-filterbar" role="group" aria-label="Filtre">{seg("Perioadă", PERIODS, 0)}{seg("Pas", ["Zi", "Săpt."], 0)}{EXPORT}{chips4(set(CTYPES))}</div>'
# D192 (SYS18): with a period filter set the load-more row shows too; a press widens the filter 4 weeks back, so "4 săpt." reads
# "8 săpt." after one press (the item keeps its place in the SegmentedControl). The chips follow the period's types (D45):
# SAPT 52–59 adds Transfer (the H7.2 → H7.1 move of 22.10, SAPT 55)
def types_of(a, b):
    def inn(d): return a <= R.D(d) <= b
    return [k for k, has in (('mortality', R.deaths(a, b) > 0), ('cull', [c for c in CCULLS if inn(c['date'])]), ('sexerr', [e for e in CSEXERR if inn(e['date'])]),
                             ('transfer', [m for m in CMOVES if inn(m['date'])]), ('sale', R.sales(a, b)), ('feed', R.feed_kg(b, 'F') > 0),
                             ('weighing', [w for w in CWEIGH if inn(w['date'])]), ('eggs', R.eggs() > 0)) if has]
assert types_of(P_FROM, P_TO) == PTYPES
T8 = types_of(R.week_dates(W8[-1])[0], P_TO); assert T8 == ['mortality', 'cull', 'transfer', 'sale', 'feed', 'weighing', 'eggs']
def filterbar_weeks(k):
    """the FilterBar with the period "k săpt." (4: as chosen; 8: after one press of the load-more row, D192)"""
    types = PTYPES if k == 4 else T8
    per = ['Ciclul curent', f'{k} săpt.', '13 săpt.', 'Interval']
    return f'<div class="ip-filterbar s4-filterbar" role="group" aria-label="Filtre">{seg("Perioadă", per, 1)}{seg("Pas", ["Zi", "Săpt."], 0)}{EXPORT}{chips4(set(types), types)}</div>'
def ptnote(*lines):  # below the region (ip-ptable__after): what a column means or leaves out
    return '<div class="ip-ptable__after">' + ''.join(f'<p class="ip-note">{l}</p>' for l in lines) + '</div>'
PW0, PW1 = R.week_dates(WK)[0], R.TODAY
PART = (f'<span class="ip-num">{R.ddm(PW0)}–{R.ddm(PW1)}</span> · {(PW1 - PW0).days + 1} din 7 zile',      # SYS14 label
        f'<span class="ip-num">{PW0.day:02d}–{R.ddm(PW1)}</span> · {(PW1 - PW0).days + 1} din 7 zile')    # D165: "17–18.11"
assert PW0.month == PW1.month
FIT = {'rz': (True, False), 'hr': (False, False), 'gr': (True, False), 'pr': (True, False)}  # uid → (dense, wrap), S4N fix 1, S4NG; SYS17 (D189): no wrapping heads, the board grows wider instead
def ptable(uid, tab, sections, scope='', bar=True, sel='Fermă', open_weeks=None, weighings=False, after='', dense=False, wrap=False, skel=None, colw=None,
           weeks=None, added=None, period=None):
    """dense = ip-ptable--dense (SYS14: figure cells one spacing step narrower, same row height). SYS17 (D189, supersedes D165's
    "fits at 1440"): heads stay on one line, wrap is never set (the FIN14 .ip-table--wraphead is retired); a table wider than
    its page scrolls under the sticky column, on the canvas the board grows wider instead (D167). open_weeks None = every week of the period open (D166, SYS14 default).
    skel = the loading state (§42/§71): the same table, every week open, bars of skel[i] ch per leaf (D132)."""
    weeks = weeks or PWEEKS4   # SYS18: the weeks on screen (W8 after one press, "added" = the 4 it appended: focus kept, live region)
    cap = f'{tab} · Hîrbovăț · Seria 2026 · SAPT {weeks[-1]}–{weeks[0]}' + (f' · {scope}' if scope else '')
    dense, wrap = FIT.get(uid, (dense, wrap))
    more = P.more(added=added, focus=bool(added), period=period)   # SYS18 (D190, D192): the load-more row before the foot total
    t, tid = P.ptable(uid, cap, sections, open_weeks=open_weeks, weighings=weighings, skel=skel or False, colw=colw, weeks=weeks, more=more)
    if not skel: assert t.count(PART[0]) == 1; t = t.replace(PART[0], PART[1])   # the partial week's visible dates; its hidden name keeps the full range
    cls = ' '.join(c for c, on in (('ip-table--wraphead', wrap),) if on)
    if cls or dense:
        t = t.replace('class="ip-table ip-table--rowheads ip-ptable"', 'class="ip-table ip-table--rowheads' + (f' {cls}' if cls else '') + ' ip-ptable' + (' ip-ptable--dense' if dense else '') + '"', 1)
    if skel:   # FIN11G: the loading table (region aria-hidden, block aria-busy, hidden status line), no Detaliu bar on Rezumat
        return (f'<section aria-hidden="true"><p class="ip-cap ip-ptable__cap">{cap}</p>{P.region(t, cap, hidden=True)}</section>')
    return (f'<section aria-labelledby="{uid}-h"><h2 class="ip-sr" id="{uid}-h">{tab}: tabel pe săptămâni și zile</h2>'
            + P.block(uid, cap, t, bar=bar, sel=sel, after=after) + '</section>')
def kg0(s):  # Rezumat: feed in whole kg (the Hrană tab keeps 0,1 kg), so the Total fermă table stays narrow (D165)
    def f(kind, key):
        if kind == 'w': return R.n(R.week(key)[f'kg_{s}'])
        return R.n(R.feed_kg(key, s) if kind == 'd' else sum(R.week(k)[f'kg_{s}'] for k in P.fweeks(key)))
    return f
def rezumat_groups(ng=False):
    """spec D1 › Rezumat (D165, D167): Total fermă only, no Detaliu, groups by record type: În viață first (D167) · Morți ♀ · ♂
    (the register's deaths, culls included; Păsări splits them off) · Ouat (ouă, % ouat, ghid, incubabile, % incubabile, jos,
    % jos; D167 adds the incubabile count and % jos; no "stare": the Readout carries it) · Hrană kg ♀ · ♂ · Vânzare last
    (D167; only when the period has sales). ng: no guide, the ghid column drops out (D92)."""
    pa, pr, hr = P.farm_pasari(), P.productie(), P.hrana()
    ou, inc, jos = P.group(pr, 'Ouat'), P.group(pr, 'Incubabile'), P.group(pr, 'Jos')
    lv = {lab: (lab, c, fn) for lab, c, fn in ou['leaves']}
    def short(vis, full): return f'<span aria-hidden="true">{vis}</span><span class="ip-sr">{full}</span>'   # short heads: the table stays narrow (D165)
    egg = [lv['ouă'], lv['% ouat']] + ([] if ng else [lv['ghid']]) + [
        (short('incub.', 'ouă incubabile'), '', inc['leaves'][0][2]), (short('% incub.', '% incubabile'), '', inc['leaves'][1][2]),
        (short('jos', 'ouă jos'), '', jos['leaves'][0][2]), (short('% jos', '% ouă jos'), '', jos['leaves'][1][2])]
    out = [P.group(pa, 'În viață'), {'name': f'{P.ico("mortality")}Morți', 'leaves': P.group(pa, 'Mortalitate')['leaves']},
           {'name': ou['name'], 'leaves': egg},
           {'name': f'{P.ico("feed")}Hrană kg', 'leaves': [(P.sx(s), c, kg0(s)) for (_, c, _), s in zip(P.group(hr, 'Hrană · kg')['leaves'], 'FM')]}]   # S4NG: unit in the group head, plain ♀ · ♂ leaves (spec D1 "Hrană kg ♀ · ♂")
    if PSALES: out.append(P.group(pa, 'Vânzare'))
    return out
RZ_NOTE = 'Morți = morții din registru, sacrificările incluse (Păsări le arată separat). % ouat: zi = ouă ÷ găini la început de zi; săptămână și perioadă = ouă ÷ zile-găină.'
# loading bars per leaf (ch of the usual figure): În viață 6 · 5, Morți 2 · 1, ouă 6, % ouat 5, ghid 5, incub. 6, % 5, jos 3, % 4, Hrană 5 · 4, Vânzare 6 · 3
RZ_SKEL = [4, 4, 2, 1, 4, 4, 4, 4, 4, 3, 4, 4, 4, 4, 3]   # ch1–ch4 exist (§42); the columns keep the loaded widths through RZ_COLW
# §54 --ip-colw on the loading table's heads (Vârstă, SAPT, then the 15 leaves: În viață ♀ ♂, Morți ♀ ♂, ouă … % jos,
# Hrană ♀ ♂, Vânzare ♀ ♂): the loaded columns' px (S4NG, measured) as shares of the 1158 px region at 1440, so every figure
# column keeps its loaded width; the loaded table is wider than the page (D167; 173 px since SYS17's one-line heads, D189:
# Vârstă (zile), % ouat, % incub., % jos re-measured), the skeleton fits it, the Data column takes the difference
RZ_COLW = [round(x / 1158 * 100, 2) for x in (110, 56, 64, 55, 44, 40, 71, 72, 55, 71, 81, 55, 63, 63, 55, 63, 55)]
def tb_rezumat(ng=False, skel=False, weeks=None, added=None, period=None):
    g = rezumat_groups(ng)
    if skel: assert not ng and len(RZ_SKEL) == sum(len(r['leaves']) for r in g)
    return ptable('rz', 'Rezumat', [('Total fermă', '', g)], bar=False, after='' if skel else ptnote(RZ_NOTE), skel=RZ_SKEL if skel else None,
                  colw=RZ_COLW if skel else None, weeks=weeks, added=added, period=period)

def panel_default(ng=False, banner='', table=None):  # S4N (D154): the Rezumat PeriodTable first in the tab (S4NG: drawn here by default)
    if table is None: table = tb_rezumat(ng)
    al = alerts(ng)
    return f'''<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-rezumat">
{banner}{table}<section class="s4-top" aria-labelledby="h-key"><h2 class="ip-sr" id="h-key">Indicatori{"" if ng else " față de ghid"}</h2>{readout(ng)}{stats(ng)}</section>
<section class="ip-block" aria-labelledby="h-sec"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-sec">Secțiuni</h2><span class="ip-note ip-num">{len(FULL)} secțiuni cu păsări · {len(HALLS)} hale · {ddate(TODAY)}</span></div><div class="s4-halls">{halls(ng)}</div>{totals(ng)}</section>
<div class="s4-split">
<section class="ip-block" aria-labelledby="h-chk"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-chk">De verificat</h2><span class="ip-note"><span class="ip-num">{len(al)}</span> {"alertă" if len(al) == 1 else "alerte"} · <span class="ip-num">{ddate(TODAY)}</span></span></div><ul class="ip-list" aria-labelledby="h-chk">{"".join(al)}</ul></section>
<section class="ip-block" aria-labelledby="h-log"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-log">Ultimele înregistrări</h2><a class="ip-link ip-note" href="#pasari">Tot jurnalul</a></div><ul class="ip-list" aria-labelledby="h-log">{"".join(events)}</ul></section>
</div>
</div>'''

# ---------- 1. loading: static skeletons (no shimmer) where data arrives ----------
def sk(cls=''):
    return f'<span class="ip-skel{(" " + cls) if cls else ""}"></span>'
# FIN11G fix 1: the six loaded Stats' guide-line items, one-line widths in px (measured at 1440, fallback fonts)
STAT_SKEL = [[[86.7], [167.91]], [[132.89], [197.59]], [[132.89], [195.08]], [[109.41, 64.19], [168.94]], [[130.05, 57.64], [210.44]], [[114.83, 64.19], [137.28]]]
def panel_loading():
    stat = ''.join(SKL.stat(g) for g in STAT_SKEL)  # FIN11G fix 1: §54 Stat skeletons, guide lines as wide as the loaded ones
    tile_ = '<div class="ip-stile s4s-sktile">' + '<div class="ip-stile__head">' + sk('s4s-w20 s4s-h16') + sk('s4s-w20') + '</div>' + sk('s4s-w50 s4s-h16') + sk('s4s-w100 s4s-h4') + ''.join('<div class="ip-stile__kv">' + sk('s4s-w40') + sk('s4s-w20') + '</div>' for _ in range(4)) + '</div>'
    hall = f'<div class="s4-hall"><div class="s4-hallhead">{sk("s4s-w20")}{sk("s4s-w30")}</div><div class="ip-tiles">{tile_}{tile_}</div></div>'
    row = f'<div class="ip-ev s4s-skrow">{sk("s4s-sq")}<div class="ip-ev__main s4s-skmain">{sk("s4s-w70")}{sk("s4s-w45")}</div>{sk("s4s-w15")}</div>'
    return f'''<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-rezumat" aria-busy="true">
<p class="ip-sr" role="status">Se încarcă rezumatul Hîrbovăț, SAPT {WK - 3}–{WK}.</p>
{tb_rezumat(skel=True)}
<section class="s4-top" aria-hidden="true">{SKL.chart_readout("s4-readout", f"0 0 {CW} {CH}")}<div class="ip-stats s4-stats">{stat}</div></section>
<section class="ip-block" aria-hidden="true"><div class="ip-sechead"><h2 class="ip-t-heading">Secțiuni</h2></div><div class="s4-halls">{hall * 3}</div><div class="s4s-sktotals">{sk("s4s-w100 s4s-h16")}{sk("s4s-w70")}{sk("s4s-w50")}</div></section>
<div class="s4-split" aria-hidden="true">
<section class="ip-block"><div class="ip-sechead"><h2 class="ip-t-heading">De verificat</h2></div><div class="ip-list">{row * 4}</div></section>
<section class="ip-block"><div class="ip-sechead"><h2 class="ip-t-heading">Ultimele înregistrări</h2></div><div class="ip-list">{row * 4}</div></section>
</div>
</div>'''

def empty_box(word, lines, step, aria='Rezumat'):
    body = ''.join(f'<p class="ip-empty__line">{l}</p>' for l in lines)
    return f'<div class="ip-emptybox" role="region" aria-label="{aria}"><div class="ip-empty">{lamp("off", word)}{body}{step}</div></div>'

# ---------- 2. no data in the period ----------
# check: no Seria 2026 record at Hîrbovăț in 01.11–15.11.2025 (cycle from 22.02.2026; the sheets start 17.02.2026)
assert all(a['date'] >= C26_START for a in TR['arrivals']), 'arrival before the cycle'
assert min(r['date'] for r in R.SECTION_DAY + R.HALL_DAY + R.FARM_DAY + R.DELIVERIES) >= '2026-02-17'
assert all(e['date'] >= C26_START for k in ('culls', 'sexing_errors', 'weighings', 'sales') for e in R.ILL[k])
def panel_period():
    step = f'<button type="button" class="ip-btn ip-btn--secondary">Alege ciclul curent</button>'
    return f'''<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-rezumat">
{empty_box("fără date", [f"Ciclul Seria 2026 nu are înregistrări între 01.11.2025 și 15.11.2025; a început pe {ddate(c26['start_date'])}."], step, "Rezumat, interval 01.11–15.11.2025")}
</div>'''

# ---------- 3. no guide ----------
BANNER_NG = f'<div class="ip-banner ip-banner--info ip-banner--act" role="note">{icon("info", 16)}<div class="ip-banner__body">Ciclul nu are ghid. Graficele arată doar valorile efective.</div><a class="ip-link ip-banner__key" href="#ciclu-ghid">Alege ghidul</a></div>'

# ---------- 4. no access ----------
def main_noaccess():
    step = f'<a class="ip-btn ip-btn--secondary" href="#acasa">{icon("home", 16)}Înapoi la Acasă</a>'
    return f'''<main class="ip-shell__main s4-main">
{header('noaccess')}
{empty_box("fără acces", ["Nu ai acces la această fermă."], step, "Acces")}
</main>'''

# ---------- 5. empty farm between cycles: Telița on 18.11.2026 (spec D1 round 2, S4NS), real cycles.json / counts ----------
# Telița's only cycle: Seria 2026, creștere, 07.10.2025–28.02.2026 (placement, then transfers out 22–28.02 to Hîrbovăț), SAPT 1–21,
# closed at 0 per D25; Seria 2025 has no Telița data (S3). Today the farm holds no birds and no cycle is open (D48: no filter bar,
# tabs or table). The Hîrbovăț 30.01.2026 drawing (header/cycle 'empty', DOM_EMPTY) is kept for importers, unused here.
CT = R.CTEL
assert (CT['status'], CT['start_date'], CT['end_date'], CT['sapt_start'], CT['sapt_end']) == ('closed', '2025-10-07', '2026-02-28', 1, 21)
assert R.COUNTS_TELITA['totals']['all'] == 0 and R.FARMS['telita']['type'] == 'growth'
TEL_XFER = sorted({o['date'] for o in TR['transfers_out']} | {a['date'] for a in TR['arrivals']})   # transfer days Telița → Hîrbovăț (as S2/S3)
assert (TEL_XFER[0], TEL_XFER[-1]) == ('2026-02-22', '2026-02-28')
TEL_SPAN = f'SAPT {CT["sapt_start"]}–{CT["sapt_end"]}'
TEL_GUIDE = R.GUIDES[R.FARMS['telita']['guide_id']]['name']           # Ross 308 creștere
DOM_TEL = (dt.date(2025, 8, 1), dt.date(2026, 12, 31), [('2025-10-01', '10.25'), ('2026-01-01', '01.26'), ('2026-04-01', '04.26'),
           ('2026-07-01', '07.26'), ('2026-10-01', '10.26')])
def track_telita(TW):
    d0, d1, tks = DOM_TEL; ppd = TW / (d1 - d0).days
    def tx(iso): return round(((dt.date.fromisoformat(iso) - d0).days) * ppd, 1)
    xa, xb, xt = tx(CT['start_date']), tx(CT['end_date']), tx(TODAY); w = round(xb - xa, 1)
    xo = tx(TEL_XFER[0]); wo = round(xb - xo, 1)        # transfers out 22–28.02 (the cycle's end)
    ticks = ''.join(f'<line class="tick" x1="{tx(i)}" x2="{tx(i)}" y1="40" y2="45"/><text class="month" x="{tx(i)}" y="56" text-anchor="middle">{l}</text>' for i, l in tks)
    lab = (f'Seria 2026, creștere, {ddate(CT["start_date"])}–{ddate(CT["end_date"])}, {TEL_SPAN.lower().replace("sapt", "sapt.")}, închis; '
           f'început prin populare, încheiat prin transfer la Hîrbovăț {R.ddm(TEL_XFER[0])}–{ddate(TEL_XFER[-1])}. Deschide ciclul.')
    blk = (f'<a href="#ciclu-telita-seria-2026" aria-label="{lab}"><title>{lab}</title><rect class="ring" x="{round(xa-4,1)}" y="12" width="{round(w+8,1)}" height="30" rx="3"/>'
           f'<rect class="cyc" x="{xa}" y="16" width="{w}" height="18" rx="2"/><text class="cyctxt" x="{round(xa+7,1)}" y="29">Seria 2026 · creștere · {TEL_SPAN}</text>'
           f'<rect class="xfer" x="{xo}" y="16" width="{wo}" height="18"/></a>')
    desc = (f'Telița: Seria 2026 creștere {ddate(CT["start_date"])}–{ddate(CT["end_date"])}, {TEL_SPAN}, închis (transfer la Hîrbovăț '
            f'{R.ddm(TEL_XFER[0])}–{ddate(TEL_XFER[-1])}); ferma e goală din {ddate(CT["end_date"])}, azi {ddate(TODAY)}. Niciun ciclu deschis.')
    return (f'<svg class="ip-track ip-track--farm ip-track--{TW}" viewBox="0 0 {TW} 60" role="group" aria-label="Ciclurile fermei Telița"><desc>{desc}</desc>'
            f'<rect class="lane" x="0" y="14" width="{TW}" height="26"/>{blk}<line class="today" x1="{xt}" x2="{xt}" y1="12" y2="42"/>'
            f'<text class="todaylbl" x="{xt}" y="9" text-anchor="middle">azi {ddate(TODAY)[:5]}</text>{ticks}</svg>')
def main_empty():
    crumbs_t = '<nav class="ip-crumbs" aria-label="Cale"><a href="#seria-2026">Seria 2026</a><span aria-hidden="true">›</span><span aria-current="page">Telița</span></nav>'
    meta = (f'<span class="ip-tag ip-tag--neutral">Creștere</span><span class="ip-tag ip-tag--neutral">{TEL_GUIDE}</span>{lamp("off", "goală")}'
            f'<span class="ip-note ip-num">din {ddate(CT["end_date"])} · niciun ciclu deschis</span>')
    head = (f'<header class="ip-pagehead">\n<div>\n{crumbs_t}\n<h1>Telița</h1>\n<div class="ip-row ip-pagehead__meta">{meta}</div>\n</div>\n'
            f'<div class="ip-row">\n<button type="button" class="ip-btn ip-btn--secondary">{icon("transfer", 16)}Mută păsări</button>\n</div>\n</header>')
    opt = (f'<li class="ip-menu__item ip-select__opt" role="option" id="cyc-tel-2026" aria-selected="false"><span class="ip-select__chk" aria-hidden="true"></span>'
           f'<span class="ip-select__txt"><span class="ip-select__top"><b>Seria 2026 · creștere</b>{lamp("off", "închis")}</span>'
           f'<span class="ip-num">{R.drange(CT["start_date"], CT["end_date"])} · {TEL_SPAN}</span></span></li>')
    sel = combo('Niciun ciclu deschis', lamp('off', 'goală'))
    hint = f'<span class="ip-hint">ultimul ciclu · Seria 2026 · {R.drange(CT["start_date"], CT["end_date"])}</span>'
    leg = '<div class="ip-tracklegend" aria-hidden="true"><span><i class="ip-rec ip-rec--transfer"></i>transfer</span><span><i class="today"></i>azi</span></div>'
    cyc = (f'<section class="s4-cycle" aria-label="Ciclu">\n<div class="ip-field"><span class="ip-label" id="cyc-lbl">Ciclu</span><div class="ip-select">{sel}'
           f'{hidden_list(GRP.replace("Hîrbovăț", "Telița") + opt)}</div>{hint}</div>\n'
           f'<div class="ip-stack s4-tight"><div class="ip-trackbox">{"".join(track_telita(w) for w in TWS)}</div>{leg}</div>\n</section>')
    step = '<a class="ip-link" href="#ciclu-telita-seria-2026">Vezi ultimul ciclu</a>'
    lines = [f'Ferma e goală din {ddate(CT["end_date"])}. Ultimul ciclu: Seria 2026, {TEL_SPAN}.', 'Ciclul pornește la prima populare sau transfer.']
    return f'''<main class="ip-shell__main s4-main">
{head}
{cyc}
{empty_box("goală", lines, step, "Rezumat")}
</main>'''

# ---------- 6b. chosen past cycle: Hîrbovăț · Seria 2025 (D51), real previous-flock data only ----------
# eggs per hall and week (D18 totals, Jos apart), egg weight, hatching eggs; hens_alive and % lay are null (no flock count),
# no grams, no guide for eggs/week → anything else is "–" with its reason. Depopulation removals are not shown as mortality.
# RF01: realdata maps the previous-flock halls to H5–H7 (D158c); the boards index by age week (D151): data SAPT 23–65,
# cycle SAPT 19–65.
PD = {'rows': R.PREV_DAILY, 'meta': R.PREV_META}; PW = {'rows': R.PREV_WEEKLY}
assert all(r['hens_alive'] is None and r['lay_pct'] is None for r in PW['rows'])
PT = PD['meta']['totals']; PTH = PD['meta']['totals_by_hall']; PLAST = PD['meta']['hall_last_egg_day']
assert PT['total_eggs'] == 4606935 == sum(r['total_eggs'] for r in PD['rows'])
pall = {r['age_week']: r for r in PW['rows'] if r['hall'] == 'all'}
phall = {(r['hall'], r['age_week']): r for r in PW['rows'] if r['hall'] != 'all'}
PWEEKS = sorted(pall); PFIN = PWEEKS[-1]
assert PFIN == c25['sapt_end'] == 65 and PWEEKS[0] == c25['data_start_sapt'] == 23 and PFIN - c25['sapt_start'] + 1 == PREV_WEEKS
fin = pall[PFIN]
assert fin['partial'] and fin['days'] == 6 and fin['to'] == PREV_END
PDAYS = len({r['date'] for r in PD['rows']})
assert PDAYS == 300
PHALLS = HALLS                                        # H5–H7 = the old H1/H3/H4 (same halls, D152)
fin_halls = [h for h in PHALLS if (h, PFIN) in phall]
assert fin_halls == ['H7'] and fin['total_eggs'] == phall[('H7', PFIN)]['total_eggs']
PGL = 'ghid: – fără date în ghid'
assert PLAST['H5'] == PLAST['H6'] == '2026-01-22'  # FIN16: the Readout sub-line names one last day for H5 and H6 (one line at 1280)
# D52/D54: mortality = Σ daily ♀/♂ deaths outside the emptying days (depopulation flag, first 15.01.2026); no flock count, so no %
PMORT = {sx: sum(r[f'mortality_{sx}'] for r in PD['rows'] if not r['depopulation']) for sx in 'FM'}
assert PMORT == {'F': 6948, 'M': 1505}, PMORT
PDEP = min(PD['meta']['depopulation_start'].values())
assert PDEP == '2026-01-15' == min(r['date'] for r in PD['rows'] if r['depopulation'])
NOFLOCK = 'fără efectiv în registru'

def readout_past():
    W0, W1 = c25['sapt_start'], PFIN; PX0 = 56; YMAX = 160000
    def px(w): return round(PX0 + (w - W0) * (X1 - PX0) / (W1 - W0), 1)
    def py(v): return round(YB - v / YMAX * (YB - YT), 1)
    assert max(r['total_eggs'] for r in pall.values()) <= YMAX
    grid = ''
    for v in (0, 80000, 160000):
        grid += f'<line class="{"axis" if v == 0 else "grid"}" x1="{PX0}" x2="{X1}" y1="{py(v)}" y2="{py(v)}"/><text class="ax" x="{PX0-4}" y="{py(v)+3}" text-anchor="end">{n(v)}</text>'
    xl = ''.join(f'<text class="ax" x="{px(w)}" y="{CH-6}" text-anchor="middle">{w}</text>' for w in [W0, 30, 40, 50, 60, PFIN])
    aline = ' '.join(f'{px(w)},{py(pall[w]["total_eggs"])}' for w in PWEEKS)
    lbl = f'Ouă pe săptămână la Hîrbovăț, Seria 2025, sapt. {PWEEKS[0]}–{PFIN}; fără date în ghid pentru ouă pe săptămână'
    chart = f'<svg class="ip-chart" viewBox="0 0 {CW} {CH}" role="group" aria-roledescription="grafic" tabindex="0" aria-label="{lbl}" aria-describedby="egg-help egg-read"><g aria-hidden="true">{grid}<polyline class="act" points="{aline}"/><line class="cur" x1="{px(PFIN)}" x2="{px(PFIN)}" y1="{YT-2}" y2="{YB}"/><circle class="pt" cx="{px(PFIN)}" cy="{py(fin["total_eggs"])}" r="3.5"/>{xl}</g></svg>'
    rows = [(w, n(pall[w]['total_eggs']), pall[w]['days'], '– fără date în ghid') for w in PWEEKS]
    per = R.drange(fin['from'], fin['to'])
    return f'''<div class="ip-readout s4-readout">
<div class="ip-readout__cap"><span>OUĂ / SĂPT. · HÎRBOVĂȚ · SAPT {PFIN}</span><span class="ip-rlamp ip-rlamp--off">FĂRĂ DATE ÎN GHID</span></div>
<div class="ip-readout__val">{n(fin["total_eggs"])} <small>ouă</small></div>
<div class="ip-readout__sub">{per} · {fin["days"]} zile · doar {fin_halls[0]} ({", ".join(h for h in PHALLS if h not in fin_halls)} până pe {ddate(PLAST["H5"])[:5]})</div>
{chart}
<span class="ip-sr" id="egg-help">Săgețile stânga și dreapta mută săptămâna; Home și End: prima și ultima săptămână.</span>
<div class="ip-chart__read" id="egg-read" role="status" aria-live="polite">SAPT {PFIN} · {n(fin["total_eggs"])} ouă · {fin["days"]} zile · ghid – fără date în ghid</div>
{sr_table(f'Ouă pe săptămână la Hîrbovăț, Seria 2025; înregistrări din {ddate(c25["data_start_date"])} (SAPT {PWEEKS[0]})', ['SAPT', 'Ouă', 'Zile', 'Ghid'], rows)}
{chart_legend([('act', 'ouă totale (Cat. 1–3 + incubabile)')], f'axa: SAPT; {W0}–{PWEEKS[0] - 1} fără date')}
</div>'''

# FIN16: the peak week of the cycle (eggs per day, whole weeks only): the top of the Readout's curve as a figure, in place
# of "Masă medie ou · SAPT 47", which repeated one row of the weekly table (the final week's egg weight)
PPEAK = max((r for r in pall.values() if not r['partial']), key=lambda r: r['eggs_per_day'])
assert (PPEAK['age_week'], PPEAK['eggs_per_day'], PPEAK['total_eggs']) == (27, 22739, 159175) and PPEAK['eggs_per_day'] == rnd(PPEAK['total_eggs'] / PPEAK['days'])
assert PPEAK['total_eggs'] == max(r['total_eggs'] for r in pall.values())  # also the highest point of the weekly chart
def stats_past():
    inc = PT['hatching_eggs'] / PT['total_eggs'] * 100
    jos = PT['floor_eggs'] / PT['total_eggs'] * 100
    # FIN16 (1280): lines that wrapped in the 206px Stat are shortened (the end date and the reason are in the header)
    return guides(f'''<div class="ip-stats s4-stats">
<div class="ip-stat">{cap('eggs', 'Ouă în ciclu')}<span class="ip-stat__val">{n(PT['total_eggs'])}</span><span class="ip-stat__guide">{PGL}</span><span class="ip-stat__guide ip-num">{PDAYS} zile · din {ddate(c25['data_start_date'])}</span></div>
<div class="ip-stat">{cap('eggs', 'Incubabile')}<span class="ip-stat__val">{n(inc,1)}<small>%</small></span><span class="ip-stat__guide">{PGL}</span><span class="ip-stat__guide ip-num">{n(PT['hatching_eggs'])} din {n(PT['total_eggs'])} ouă</span></div>
<div class="ip-stat">{cap('eggs', 'Vârf · ouă pe zi')}<span class="ip-stat__val">{n(PPEAK['eggs_per_day'])}</span><span class="ip-stat__guide">{PGL}</span><span class="ip-stat__guide ip-num">SAPT {PPEAK['age_week']} · {R.drange(PPEAK['from'], PPEAK['to'])}</span></div>
<div class="ip-stat">{cap('eggs', 'Ouă jos')}<span class="ip-stat__val">{n(PT['floor_eggs'])}</span><span class="ip-stat__guide">{PGL}</span><span class="ip-stat__guide">nu intră în total · <span class="ip-num">{pct(jos)}</span> din total</span></div>
<div class="ip-stat">{cap('eggs', 'Ouat')}<span class="ip-stat__val">–</span><span class="ip-stat__guide">{NOFLOCK}</span><span class="ip-stat__guide">ouă pe găină: – fără efectiv</span></div>
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
        out.append(ev('eggs', f'Ouă · {", ".join(r["hall"] for r in rs)} · <span class="ip-num">{n(sum(r["total_eggs"] for r in rs))}</span> ouă', f'{ddate(d)} · SAPT {rs[0]["age_week"]} · din registru'))
    return out

# S4NG (D170): the past cycle's Rezumat table: its last 4 age weeks (SAPT 62–65) all open, "Încarcă încă 4 săptămâni" under it
# (SYS18, D190: 4 din 43, the register's weeks SAPT 23–65); the register of the previous flock has eggs per hall and deaths, no flock count and no feed (D52): no În viață,
# no % ouat or guide, no Hrană; emptying days (from 15.01.2026) are not mortality (D54)
PCAL = P.Cal(placed=dt.date.fromisoformat(R.D(c25['start_date']).isoformat()) - dt.timedelta(days=c25['age_days_start']), last=R.D(PREV_END))
assert PCAL.sapt(R.D(PREV_END)) == PFIN == 65 and PCAL.age(R.D(c25['start_date'])) == c25['age_days_start']
PW4 = [PFIN, PFIN - 1, PFIN - 2, PFIN - 3]
PBY = {}
for _r in PD['rows']: PBY.setdefault(R.D(_r['date']), []).append(_r)
for _k in PW4: assert all(d in PBY for d in PCAL.days(_k)) and sum(r['total_eggs'] for d in PCAL.days(_k) for r in PBY[d]) == pall[_k]['total_eggs']
def _pdays(kind, key): return [key] if kind == 'd' else PCAL.days(key) if kind == 'w' else [d for k in (key or PW4) for d in PCAL.days(k)]   # 'f': the loaded weeks (SYS18)
def _prows(kind, key, h=None): return [r for d in _pdays(kind, key) for r in PBY[d] if h is None or r['hall'] == h]   # S14N: h = one hall's rows (S14 hall week rows)
def p_mort(sx, h=None):
    def f(kind, key):
        rows = _prows(kind, key, h)
        live = [r for r in rows if not r['depopulation']]
        return (DASH_, 'golire, nu mortalitate') if not live else n(sum(r[f'mortality_{sx}'] for r in live))
    return f
def p_sum(field, h=None):
    return lambda kind, key: n(sum(r[field] for r in _prows(kind, key, h)))
def p_share(field, h=None):
    def f(kind, key):
        rows = _prows(kind, key, h)
        return pct(sum(r[field] for r in rows) / sum(r['total_eggs'] for r in rows) * 100)
    return f
DASH_ = '–'
def tb_past():
    def short(vis, full): return f'<span aria-hidden="true">{vis}</span><span class="ip-sr">{full}</span>'
    groups = [{'name': f'{P.ico("mortality")}Morți', 'leaves': [(P.sx('F'), 'is-hen', p_mort('F')), (P.sx('M'), 'is-rooster', p_mort('M'))]},
              {'name': f'{P.ico("eggs")}Ouat', 'leaves': [('ouă', '', p_sum('total_eggs')), (short('incub.', 'ouă incubabile'), '', p_sum('hatching_eggs')),
                                                         (short('% incub.', '% incubabile'), '', p_share('hatching_eggs')), (short('jos', 'ouă jos'), '', p_sum('floor_eggs')),
                                                         (short('% jos', '% ouă jos'), '', p_share('floor_eggs'))]}]
    cap = f'Rezumat · Hîrbovăț · Seria 2025 · SAPT {PW4[-1]}–{PW4[0]}, ultimele 4 din SAPT {PWEEKS[0]}–{PFIN}'
    t, tid = P.ptable('rz', cap, [('Total fermă', '', groups)], weeks=PW4, cal=PCAL, more=P.more(first=PWEEKS[0], word='registrului'))   # SYS18 (D190, D192): 4 din 43 (SAPT 23–65)
    after = (f'<div class="ip-ptable__after">'
             f'<p class="ip-note">Tabelul începe la SAPT {PWEEKS[0]}, banda ciclurilor la SAPT {c25["sapt_start"]}: '
             f'începutul ciclului e dedus ({ddate(c25["start_date"])}), registrul are date din {ddate(c25["data_start_date"])}. '
             f'Fără efectiv în registru: fără În viață și % ouat; golirea din {ddate(PDEP)} nu apare ca mortalitate.</p></div>')
    assert PWEEKS[0] == c25['data_start_sapt'] == 23 and c25['sapt_start'] == 19   # S4NS: why the table span differs from the CycleTrack
    return (f'<section aria-labelledby="rz-h"><h2 class="ip-sr" id="rz-h">Rezumat: tabel pe săptămâni și zile</h2>'
            + P.block('rz', cap, t, bar=False, after=after) + '</section>')

def panel_past():
    # FIN16: the AlertRow empty state (SYS01: off lamp "nimic de verificat" + one sentence) as the list's own empty row
    # (li.ip-list__empty, as S2's "nimic introdus"), in place of the screen box .s4s-none
    none = f'<ul class="ip-list" aria-labelledby="h-chk"><li class="ip-list__empty">{lamp("off", "nimic de verificat")}<span>Ciclul Seria 2025 la Hîrbovăț e închis din {ddate(PREV_END)}; nu are alerte.</span></li></ul>'
    return f'''<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-rezumat">
{tb_past()}
<section class="s4-top" aria-labelledby="h-key"><h2 class="ip-sr" id="h-key">Indicatori, Seria 2025</h2>{readout_past()}{stats_past()}</section>
<section class="ip-block" aria-labelledby="h-sec"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-sec">Hale</h2><span class="ip-note ip-num">3 hale · ouă pe hală din registru · {R.drange(c25['data_start_date'], PREV_END)}</span></div><div class="s4-halls">{"".join(htile_past(h) for h in PHALLS)}</div></section>
<div class="s4-split">
<section class="ip-block" aria-labelledby="h-chk"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-chk">De verificat</h2><span class="ip-note">ciclu închis</span></div>{none}</section>
<section class="ip-block" aria-labelledby="h-log"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-log">Ultimele înregistrări</h2><a class="ip-link ip-note" href="#pasari">Tot jurnalul</a></div><ul class="ip-list" aria-labelledby="h-log">{"".join(events_past())}</ul><p class="ip-note ip-m0">Ultimele zile ale ciclului. Ieșirile de păsări din 15.01.2026 (golirea halelor) nu apar ca mortalitate.</p></section>
</div>
</div>'''

# ---------- 7. new cycle dialog ----------
DLG = f'''<div class="ip-scrim ip-scrim--page"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-note">
<h2 id="dlg-t">Începe ciclu nou la Hîrbovăț</h2>
<p class="ip-dialog__note" id="dlg-note">Ciclul curent (sapt. {WK}) se închide pe {ddate(TODAY)}. Săptămâna rămâne cea de vârstă a păsărilor.</p>
<div class="ip-field"><label class="ip-label" for="nc-date">Data de început</label><div class="ip-row s4s-daterow"><span class="ip-seg" role="radiogroup" aria-label="Zi rapidă"><span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Azi</span><span class="ip-seg__item" role="radio" aria-checked="false" tabindex="-1">Ieri</span></span><span class="ip-input s4s-date"><input id="nc-date" type="text" inputmode="numeric" value="{ddate(TODAY)}" placeholder="zz.ll.aaaa" aria-describedby="nc-date-h"/><button type="button" class="ip-iconbtn" aria-label="Alege data de început din calendar" aria-haspopup="dialog" aria-expanded="false">{icon("cal", 16)}</button></span></div><span class="ip-hint ip-num" id="nc-date-h">SAPT {WK} · ciclu nou din {ddate(TODAY)}</span></div>
<div class="ip-field"><span class="ip-label" id="nc-guide-l">Ghid</span><div class="ip-select"><div class="ip-input ip-select__btn" id="nc-guide" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="false" aria-controls="nc-glist" aria-labelledby="nc-guide-l" aria-describedby="nc-guide-h"><span class="ip-select__val">Ross 308 ouat</span>{icon("chev", 16)}</div><ul class="ip-menu ip-select__menu" role="listbox" id="nc-glist" aria-labelledby="nc-guide-l" tabindex="-1" hidden="hidden"><li class="ip-menu__item ip-select__opt" role="option" id="nc-g0" aria-selected="false"><span class="ip-select__chk" aria-hidden="true"></span><span class="ip-select__txt"><span class="ip-select__top">Fără ghid</span></span></li><li class="ip-menu__group" role="presentation">Ghiduri de ouat</li><li class="ip-menu__item ip-select__opt" role="option" id="nc-g1" aria-selected="true"><span class="ip-select__chk" aria-hidden="true">{icon("check", 16)}</span><span class="ip-select__txt"><span class="ip-select__top">Ross 308 ouat <span class="ip-note">ghidul fermei</span></span></span></li></ul></div><span class="ip-hint" id="nc-guide-h">ghidul fermei · ouat</span></div>
<div class="ip-field"><label class="ip-label" for="nc-why">Motiv</label><span class="ip-input ip-input--area"><textarea id="nc-why" rows="3" required="required" aria-required="true" aria-describedby="nc-why-h"></textarea></span><span class="ip-hint" id="nc-why-h">Obligatoriu. Apare în jurnalul de modificări.</span></div>
<div class="ip-dialog__keys"><button type="button" class="ip-btn ip-btn--secondary ip-focus" autofocus="autofocus">Anulează</button><button type="button" class="ip-btn ip-btn--primary">{icon("cycle", 16)}Începe ciclu nou</button></div>
</div></div>'''

def main_page(mode='default', panel=None, ng=False):
    # S4NG: the current cycle's Rezumat page (default, menu, no guide, loading; the S9/S10 backdrops) has the default board's
    # FilterBar (4 săpt., Zi, the period's chips), since its table shows SAPT 56–59 (D159, D166)
    fb = filterbar(mode) if mode in ('period', 'past') else filterbar4()
    return f'''<main class="ip-shell__main s4-main">
{header('past' if mode == 'past' else 'noguide' if ng else 'default')}
{cycle(mode if mode in ('menu', 'past') else 'default')}
<div class="s4-tight">{fb}{tabs}</div>
{panel}
</main>'''

CSS = '''body{margin:0}
.s4-tight{display:flex;flex-direction:column;gap:var(--space-2)}
.s4-panel{display:flex;flex-direction:column;gap:var(--space-6)}
.s4-lamps{display:flex;gap:var(--space-3)}
.s4-main{container:s4main / inline-size} /* FIN16: the S4 page lays itself out from its own width (beside a SidePanel 720 / 560) */
.s4-cycle{display:grid;grid-template-columns:calc(var(--space-12) * 8 + var(--space-8)) minmax(0,1fr);gap:var(--space-6);align-items:start} /* FIN16: 416 + 24 leave the track box a CycleTrack step (additions.css §65): 720 at 1440, 560 at 1280 */
.s4-cycle > .s4-tight{padding-top:var(--space-2)} /* S4P: the track lane centres on the CycleSelector field */
.s4-top{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:var(--space-4);align-items:stretch}
.s4-readout{justify-content:flex-start}
.s4-stats{grid-template-columns:repeat(2,minmax(0,1fr))}
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
.s4-filterbar{row-gap:var(--space-2)}
.s4-chips{flex-basis:100%;display:flex;flex-wrap:wrap;gap:var(--space-2);align-items:center}
.s4s-rel{position:relative}
.s4s-rel > .ip-shell{min-height:inherit}
@container s4main (max-width: 999.98px){ /* FIN16: the narrow page (720 / 560 beside a SidePanel): one column; the track takes the page width (its 720 / 560 drawings), the field keeps its width */
.s4-cycle,.s4-top,.s4-halls,.s4-split{grid-template-columns:minmax(0,1fr)}
.s4-cycle > .ip-field{max-width:calc(var(--space-12) * 8 + var(--space-8))}
.s4-cycle > .s4-tight{padding-top:0}
}'''
# S4 boards only (default, states, tabs): S3 and S11 reuse CSS/CSS_STATES and must stay byte-identical
# FIN15: CSS_S4 (the S4TP section-head link rule) is now additions.css §56 `.ip-sechead > .ip-link`
# FX05: the EventRow figure no-wrap (CSS) is additions.css §1, the block skeleton bar (CSS_STATES) §42
# state boards only
CSS_STATES = '''.s4s-sktile{gap:var(--space-2)}
.s4s-sktotals{background:var(--panel);border:var(--ip-hair) solid var(--line);border-top-color:var(--line-strong);border-radius:var(--radius-md);padding:var(--space-3) var(--space-4);display:flex;flex-direction:column;gap:var(--space-2)}
.s4s-skrow{gap:var(--space-3);align-items:center}
.s4s-skmain{display:flex;flex-direction:column;gap:var(--space-2);flex:1}
.s4s-sq{width:var(--record-tile);height:var(--record-tile);flex:none;border-radius:var(--radius-sm)}
.s4s-h4{height:var(--space-1)}.s4s-h16{height:var(--space-4)}.s4s-h20{height:var(--space-5)}.s4s-h28{height:calc(var(--space-6) + var(--space-1))}
.s4s-w15{width:15%}.s4s-w20{width:20%}.s4s-w25{width:25%}.s4s-w30{width:30%}.s4s-w40{width:40%}.s4s-w45{width:45%}.s4s-w50{width:50%}.s4s-w60{width:60%}.s4s-w70{width:70%}.s4s-w80{width:80%}.s4s-w100{width:100%}
.s4s-daterow{gap:var(--space-2);flex-wrap:nowrap}
.s4s-date{flex:1;gap:var(--space-1);padding-right:var(--space-1)}
.s4s-date > input{font-family:var(--font-mono);font-variant-numeric:tabular-nums}
.s4s-date .ip-iconbtn{width:var(--space-6);height:var(--space-6)}'''

# D167 (S4NG, canvas only): a board drawn wider than 1440 (board_widths.json) keeps its page at the 1440 layout (beside a
# docked SidePanel: 1440 less the panel) and lets each table region grow to its table's full width, past the page's right
# edge, instead of a clipped sideways scroll; the product keeps the scroll region with the sticky Data column (D165)
CSS_WIDE = '''.s4s-wide > .ip-shell{width:var(--board-page-w)}
.s4s-wide > .ip-dockpage > .ip-shell{width:calc(var(--board-page-w) - var(--panel-w))}
.s4s-wide .ip-tablescroll{width:min-content;min-width:100%;max-width:none;overflow:visible}'''
FONTS = '<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500&amp;family=IBM+Plex+Mono:wght@400;500&amp;display=swap">'
SYS = ROOT / 'design/system'
HEIGHTS = json.loads(pathlib.Path(__file__).with_name('s4s_heights.json').read_text()) if pathlib.Path(__file__).with_name('s4s_heights.json').exists() else {}
def write(stem, title, body, overlay='', H=None, css=None, vp=None):
    H = H or HEIGHTS.get(stem, 1874)
    import canvas_rows as _CR
    W = _CR.width(stem, 1440)   # D167: a board whose table is wider than its page is drawn wider (board_widths.json)
    css = css if css is not None else CSS + '\n' + CSS_STATES
    vps = f'; --board-vp-h:{vp}px' if vp else ''  # FIN04 fix 1: a tall board's Dialog centres in its first viewport (additions.css §50)
    wide = W > 1440
    if wide:  # D167 (S4NG): the page keeps its 1440 layout, its wide tables extend past the page's right edge in full
        vps += '; --board-page-w:1440px'; css = css + '\n' + CSS_WIDE
    root = f'<div class="ip-root s4s-rel{" s4s-wide" if wide else ""}" style="padding:0; min-height:{H}px{vps}">{body}{overlay}</div>'
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
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":{W},"height":{H}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
    out = ROOT / f'design/canvas/project/{stem}.dc.html'
    out.write_text(dc, encoding='utf-8')
    # RF03 fix: the render wrapper is not pinned to the stored height H (a render would only read H back); it keeps the
    # 1024 board-viewport floor (canvas.md), so sNN_render measures the content's natural height
    wroot = root.replace(f'min-height:{H}px', f'min-height:{min(H, 1024)}px', 1)
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{FONTS}<style>{css}</style></head><body>{wroot}</body></html>'''
    pathlib.Path('/tmp/s4s').mkdir(exist_ok=True)
    pathlib.Path(f'/tmp/s4s/{stem}.html').write_text(wrap, encoding='utf-8')
    print('ok', stem, *((W,) if W != 1440 else ()), H, len(dc))

def shell(main, viewer=False, inert=False, series='2026'):
    sb = sidebar('Observator', viewer=True) if viewer else sidebar(series=series)
    inner = f'{sb}{main}'
    return f'<div class="ip-shell"{" inert=\"inert\" aria-hidden=\"true\"" if inert else ""}>{inner}</div>'

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
        write(stem, title, body, ov, vp=1024 if stem == B + 'ciclu-nou' else None)  # board viewport 1440×1024 (canvas.md)
    if '--canvas' in sys.argv[2:]:  # RF01: refresh the 8 entries' h/title in place (hand-placed row: x and y kept, canvas_rows.place)
        import canvas_rows as CR
        c = CR.load(ROOT); CR.place(c, 'S4-ferma-rezumat', [(stem, 1440, HEIGHTS.get(stem, 1874), title) for stem, title, _, _ in BOARDS])
        CR.save(ROOT, c)
