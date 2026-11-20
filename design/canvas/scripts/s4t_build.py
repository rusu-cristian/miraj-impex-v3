"""Build the S4 · Fermă tab boards (S4N, D154 table first): Păsări, Hrană, Greutate, Producție, one desktop board per tab
(owner, today 18.11.2026 = SAPT 59, Hîrbovăț · Seria 2026, period "4 săpt." = SAPT 56–59, D159; grain "Zi": every week open
with its days, newest first, D166), and the Rezumat table for the S4 default board (s4_build.py imports tb_rezumat / main_rezumat from here).
python3 -I s4t_build.py <repo_root> [--canvas]
Every tab opens on its PeriodTable (SYS14, additions.css §71): the row and column builders are imported from
design/system/scripts/sys14_build.py (ptable, block, farm_pasari, hall_pasari, productie, hrana, greutate), never copied;
this file adds the farm-page groups the preview lacks (Sacrificare, Livrări · Lei · Stoc). Then the tab's charts (Readouts
with GuideCharts), then on Păsări the event log. Figures come only from realdata.py (DATA03 extracts, D151–D159); the cull,
the weighing and the log's authors/times are the illustrative records of illustrative-events-hirbovat.json (D152).
--canvas refreshes the 4 tab entries' h/title in place (hand-placed row: x and y kept). Writes design/canvas/project/
S4-ferma--<tab>.dc.html and /tmp/s4s/S4-ferma--<tab>.html (plain wrappers for s4t_render.py)."""
import sys, json, pathlib, importlib.util, datetime as dt

_spec = importlib.util.spec_from_file_location('s4s_build', pathlib.Path(__file__).resolve().with_name('s4s_build.py'))
s4 = importlib.util.module_from_spec(_spec)
DROP = globals().get('DROP')   # S16P fix 1: a deleted Mortalitate record (S16 after-delete board only); s4s refuses it until RF08
if DROP: s4.DROP = DROP
_spec.loader.exec_module(s4)  # reads sys.argv[1] (repo root)
R = s4.R
P = s4.P   # PeriodTable builders (design/system/scripts/sys14_build.py, loaded once by s4s_build; no file written on import)
assert P.R is R
def __getattr__(name):  # PEP 562: `S` = s4s's lazy legacy alias (the superseded sample) for rows not yet ported (s14, s16); S4 boards never read it
    if name == 'S': return s4.S
    raise AttributeError(name)

n, sgn, pct, rnd = R.n, R.sgn, R.pct, R.rnd
ddate, icon, tile, ico, hen, roo, pair, lamp, sg = s4.ddate, s4.icon, s4.tile, s4.ico, s4.hen, s4.roo, s4.pair, s4.lamp, s4.sg
F, M, NB = s4.F, s4.M, s4.NB
import sexsign as SXS   # SX01: ♀/♂ outside a figure span = the §72 hook (glyph aria-hidden + word), sexsign.py
sexify, sign = SXS.sexify, SXS.sign
GL, TR = s4.GL, s4.TR
TODAY, WK, SECTIONS, HALLS = s4.TODAY, s4.WK, R.HSECS, R.HHALLS
def ddm(d): return R.ddm(d)
def num(v): return f'<span class="ip-num">{v}</span>'
DASH = '–'
NBSP = ' '  # no-break space: a ♀/♂ sign, its figure and unit stay on one line
NOGD = 'ghid: – fără date în ghid'
HERE = pathlib.Path(__file__).resolve().parent
HEIGHTS = json.loads((HERE / 's4t_heights.json').read_text()) if (HERE / 's4t_heights.json').exists() else {}

# ---------------------------------------------------------------- period (D159: the last 4 age weeks, newest first)
# S4NG: the period, its record types and the Rezumat table live in s4s_build (the shared Rezumat page draws its table)
WEEKS, P_FROM, P_TO, SER = s4.PWEEKS4, s4.P_FROM, s4.P_TO, s4.SER
inp, PCULLS, PSEXERR, PMOVES, PSALES, PWEIGH, PTYPES = s4.inp, s4.PCULLS, s4.PSEXERR, s4.PMOVES, s4.PSALES, s4.PWEIGH, s4.PTYPES
# D191 (SYS18): the period is the cycle ("Ciclul curent"); its record types set the chips and the tables' groups (D45)
incy, CCULLS, CSEXERR, CMOVES, CWEIGH, CTYPES, CYC_FROM, CYC_WEEKS, W8 = s4.incy, s4.CCULLS, s4.CSEXERR, s4.CMOVES, s4.CWEIGH, s4.CTYPES, s4.CYC_FROM, s4.CYC_WEEKS, s4.W8

# ---------------------------------------------------------------- shell: tabs + filter bar
TABS = [('rezumat', 'Rezumat'), ('pasari', 'Păsări'), ('hrana', 'Hrană'), ('greutate', 'Greutate'), ('productie', 'Producție')]
def tabs(sel):  # same markup as S4 (roving tabindex, one tabpanel #tp-ferma that every tab swaps); only the selection moves
    return '<div class="ip-tabs" role="tablist" aria-label="Vederi Hîrbovăț">' + ''.join(
        f'<a class="ip-tab{" is-selected" if k == sel else ""}" role="tab" id="tab-{k}" aria-controls="tp-ferma" aria-selected="{"true" if k == sel else "false"}" tabindex="{0 if k == sel else -1}" href="#{k}">{t}</a>'
        for k, t in TABS) + '</div>'
BIRD = {'mortality', 'cull', 'sexerr', 'transfer', 'sale'}
CHIPW, chips = s4.CHIPW, s4.chips4
def filterbar(tab): return s4.filterbar4()   # D159/D166/D165: one FilterBar for every tab (shared tag state)
def main(tab, panel):
    return f'''<main class="ip-shell__main s4-main">
{s4.header('default')}
{s4.cycle('default')}
<div class="s4-tight">{filterbar(tab)}{tabs(tab)}</div>
{panel}
</main>'''
def page(tab, panel):
    return s4.shell(main(tab, f'<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-{tab}">\n{panel}\n</div>'))

# ---------------------------------------------------------------- PeriodTable blocks (SYS14)
note, PART, FIT, ptable = s4.ptnote, s4.PART, s4.FIT, s4.ptable   # S4NG: moved to s4s_build (the shared Rezumat page)
W0, W1 = s4.PW0, s4.PW1
def thico(k): return f'<span class="ip-rec ip-rec--{s4.REC.get(k, k)} ip-table__ico">{icon(k, 14)}</span>'  # §32 head icon (FIN14B)
def span(kind, key):
    if kind == 'w': a, b = R.week_dates(key); return a, min(b, R.TODAY)
    return (key, key) if kind == 'd' else P.fspan(key)   # 'f': the loaded weeks (SYS18: the foot adds up the weeks on screen)
def in_scope(r, h): return h is None or r['section'].startswith(h + '.')
def carved(kind, key, s, h):
    """illustrative culls / sexing errors are carved out of that compartment-day's real deaths (DATA03): Mortalitate shows
    what is left, the carved birds sit in their own group, so the groups add up to the register's deaths."""
    a, b = span(kind, key)
    return sum(r['carved_from_mortality_of_day'] - r['mortality_left'] for r in CCULLS + CSEXERR
               if r['sex'] == s and in_scope(r, h) and a <= R.D(r['date']) <= b)

def pasari(h=None):
    """Păsări groups of one scope (farm or hall), chips Mortalitate · Sacrificare · Vânzare on (D45 types of the period):
    În viață (end of day / week; first, D167) · Mortalitate (deaths less the carved culls) · % cumulat (D162: real deaths
    since arrival, culls included) · Sacrificare · Vânzare (D155; always the last group, D167)."""
    base = P.farm_pasari() if h is None else P.hall_pasari(h)   # [În viață, Mortalitate, % cumulat, Vânzare] (D167)
    bm, bc, bs, bv = (P.group(base, w) for w in ('Mortalitate', '% cumulat', 'Vânzare', 'În viață'))
    def mort(s, fn):
        def f(kind, key):
            v = fn(kind, key); c = carved(kind, key, s, h)
            if isinstance(v, tuple) or not c: return v
            a, b = span(kind, key); return R.n(R.deaths(a, b, sx=s, hall=h) - c)
        return f
    def cull(s, guard):
        def f(kind, key):
            g = guard(kind, key)
            if isinstance(g, tuple): return g          # the hall is empty (goală din 14.11)
            a, b = span(kind, key)
            return R.n(sum(c['qty'] for c in CCULLS if c['sex'] == s and in_scope(c, h) and a <= R.D(c['date']) <= b))
        return f
    def sexerr(s, guard):   # signed as typed (−1 = one bird out of this sex), illustrative (19.05 H6.2 ♀ −1)
        def f(kind, key):
            g = guard(kind, key)
            if isinstance(g, tuple): return g
            a, b = span(kind, key)
            return R.n(sum(e['qty'] for e in CSEXERR if e['sex'] == s and in_scope(e, h) and a <= R.D(e['date']) <= b))
        return f
    def xfer(s, guard):   # signed: + in, − out (arrivals 22–28.02; the H7.2 → H7.1 move of 22.10 nets 0 in the farm and the hall)
        def f(kind, key):
            g = guard(kind, key)
            if isinstance(g, tuple): return g
            a, b = span(kind, key); v = 0
            for m in CMOVES:
                if m['sex'] != s or not a <= R.D(m['date']) <= b: continue
                if 'section' in m: v += m['qty'] if in_scope(m, h) else 0                                        # arrival
                else: v += (m['qty'] if h is None or m['to'].startswith(h + '.') else 0) - (m['qty'] if h is None or m['from'].startswith(h + '.') else 0)
            return '+' + R.n(v) if v > 0 else R.n(v)
        return f
    m = {'name': bm['name'], 'leaves': [(lab, c, mort(s, fn)) for (lab, c, fn), s in zip(bm['leaves'], 'FM')]}
    cu = {'name': f'{thico("cull")}Sacrificare', 'leaves': [(P.sx(s), c, cull(s, fn)) for (_, c, fn), s in zip(bs['leaves'], 'FM')]}
    # D191/D45: the cycle's bird types add Eroare sexare and Transfer (0 in SAPT 56–59), so the columns stay put while weeks load
    se = {'name': f'<span class="ip-table__ico">{icon("sexerr", 14)}</span>Eroare sexare', 'leaves': [(P.sx(s), c, sexerr(s, fn)) for (_, c, fn), s in zip(bs['leaves'], 'FM')]}
    tr = {'name': f'{thico("transfer")}Transfer', 'leaves': [(P.sx(s), c, xfer(s, fn)) for (_, c, fn), s in zip(bs['leaves'], 'FM')]}
    return [bv, m, bc, cu, se, tr, bs]   # D167: În viață first, Vânzare last; RecordCode order between
assert R.week(WK - 1)['mort_F'] == R.deaths(*R.week_dates(WK - 1), sx='F') == 36 and R.week(WK - 1)['mort_M'] == 8  # spec D1 sample

kg0, rezumat_groups = s4.kg0, s4.rezumat_groups

W = R.week
UNP_P = [u for u in R.UNPRICED if inp(u['date'])]
assert [(u['date'], u['qty_kg']) for u in UNP_P] == [('2026-11-12', 9000)]
def hrana_farm():
    """Hrană · Total fermă: g/pasăre efectiv · ghid and kg (SYS14 hrana) + Livrări kg ♀ · ♂ feed, Lei (owner, D64) and Stoc kg
    at the end of the day / week; week rows sum kg, deliveries and lei, stock = the week's end."""
    def deliv(s):
        def f(kind, key):
            if kind == 'w': return R.n(W(key)[f'deliv_{s}_kg'])
            if kind == 'd': return R.n(R.day(key)[f'deliv_{s}_kg'] or 0)
            return R.n(sum(W(k)[f'deliv_{s}_kg'] for k in P.fweeks(key)))
        return f
    def lei(kind, key):
        if kind == 'd':
            if any(u['date'] == R.iso(key) for u in UNP_P): return '<span class="ip-muted">fără preț</span>'
            return R.n(R.day(key)['lei'] or 0)
        return R.n(W(key)['lei'] if kind == 'w' else sum(W(k)['lei'] for k in P.fweeks(key)))
    def stock(s):
        def f(kind, key):
            if kind == 'w': return R.n(W(key)[f'stock_{s}_end'], 1)
            return R.n(R.day(key if kind == 'd' else P_TO)[f'stock_{s}_kg'], 1)
        return f
    for k in WEEKS:  # week lei = the priced deliveries of the week (unpriced ones are named in the note)
        a, b = span('w', k)
        assert W(k)['lei'] == sum(d['total_lei'] for d in R.DELIVERIES if d['priced'] and a <= R.D(d['date']) <= b)
        assert W(k)['deliv_F_kg'] + W(k)['deliv_M_kg'] == sum(d['qty_kg'] for d in R.DELIVERIES if a <= R.D(d['date']) <= b)
    # a feed with no delivery in the period has no column (D45 per record): D191, the cycle has ♀ and ♂ feed deliveries
    # (SAPT 56–59 ♀ only: the ♂ column is 0 there)
    dsx = [s for s in 'FM' if any(d['for_sex'] == s and incy(d['date']) for d in R.DELIVERIES)]
    assert dsx == ['F', 'M']
    dl = ({'name': f'{thico("delivery")}Livrări kg', 'single': True, 'sr': '', 'leaves': [(P.sx('F'), 'is-hen', deliv('F'))]}
          if dsx == ['F'] else {'name': f'{thico("delivery")}Livrări kg', 'leaves': [(P.sx(s), P.SX[s][0], deliv(s)) for s in dsx]})
    return P.hrana() + [dl,
        {'name': f'{thico("price")}Lei', 'single': True, 'sr': ', livrări', 'leaves': [('', '', lei)]},
        {'name': f'{thico("delivery")}Stoc kg', 'leaves': [(P.sx('F'), 'is-hen', stock('F')), (P.sx('M'), 'is-rooster', stock('M'))]}]

tb_rezumat = s4.tb_rezumat
def tb_pasari(hale=False):
    """D165: the default board shows Detaliu Fermă (fits at 1440); Detaliu Hale is its own state board (scrolls sideways)."""
    secs = [('Total fermă', '', pasari())] + ([(f'Hala {h[1:]}', P.GOALA if h == 'H7' else '', pasari(h)) for h in HALLS] if hale else [])
    c = PCULLS[0]
    return ptable('pa', 'Păsări', secs, scope='pe hale' if hale else '', sel='Hale' if hale else 'Fermă',
                  after=note(f'Golirea din {ddm(R.EMPTYING_SINCE)} e Vânzare; Mortalitate = doar morți reali. % cumulat = morți din sosire ÷ păsări sosite. '
                             f'Sacrificarea din {ddm(c["date"])} ({c["section"]} {M} {c["qty"]}) e ilustrativă și e scăzută din morții zilei. '
                             f'Eroare sexare și Transfer au înregistrări în ciclu ({num(ddm(CSEXERR[0]["date"]))}; {num(R.drange(CYC_FROM, R.D("2026-02-28"), year_end=False))}, {num(ddm(CMOVES[0]["date"]))}), nu în aceste săptămâni.'))
# D163/D165: the workbook's ♀ kg cells far from g × birds at the start of the day (460 kg/compartiment, 02–06.11; extracts
# README anomaly), kept as typed: the Hrană note names them per week, as S11's Consum foot
ODD = {}
for _k in WEEKS:
    for _d in P.days_of(_k):
        for _s in SECTIONS:
            for _x in 'FM':
                _r, _st = R.sd(_d, _s, _x), R.alive_start(_s, _x, _d)
                if _r and _r['kg'] and _st and not _r['sold'] and abs(_r['kg'] / (_r['g'] * _st / 1000) - 1) > 0.25:  # a sale day feeds fewer birds
                    ODD.setdefault(_k, []).append((_d, _x, _r['kg'], _r['g'] * _st / 1000))
def odd_note():
    cells = [c for k in ODD for c in ODD[k]]
    kgs = {rnd(c[2]) for c in cells}; sxs = {c[1] for c in cells}
    assert kgs == {460} and sxs == {'F'}, (kgs, sxs)
    exp = rnd(sum(c[3] for c in cells) / len(cells) / 10) * 10
    def days(k): ds = sorted({c[0] for c in ODD[k]}); return R.drange(ds[0], ds[-1], year_end=False) if len(ds) > 1 else R.ddm(ds[0])
    return (f'{F} {num(n(460) + NBSP + "kg")}/compartiment în registru (g × păsări ≈ {num(n(exp) + NBSP + "kg")}): '
            + '; '.join(f'{num(R.SAPT(k))}: {num(days(k))}' for k in sorted(ODD, reverse=True)) + '; păstrate cum sunt scrise.')
def tb_hrana():
    return ptable('hr', 'Hrană', [('Total fermă', '', hrana_farm())],
                  after=note('g/pasăre: zi = g din registru (media pe păsări a secțiunilor hrănite); săptămână și perioadă = kg ÷ zile-pasăre hrănite. '
                             f'Stoc la sfârșitul zilei, al săptămânii. Lei fără livrarea fără preț din {ddm(UNP_P[0]["date"])} ({num(n(UNP_P[0]["qty_kg"]) + " kg")}).',
                             odd_note()))
def tb_greutate():
    w = PWEIGH[0]
    g = P.greutate(); NB2 = {'medie g': 'medie\u00a0g', 'față de ghid': 'față\u00a0de ghid', 'CV %': 'CV\u00a0%', 'unif. %': 'unif.\u00a0%'}
    def nb(lab):
        for a, b in NB2.items(): lab = lab.replace(f' {a}', f' {b}')
        return lab
    g = [{**r, 'leaves': [(nb(lab), c, fn) for lab, c, fn in r['leaves']]} for r in g]
    return ptable('gr', 'Greutate', [('Total fermă', '', g)], weighings=True,
                  after=note(f'{DASH} fără date în ghid: ghidul de creștere are greutăți până la SAPT 25. Total fermă = secțiunile cântărite '
                             f'({w["section"]}, {num(ddate(w["date"]))}); cântărire ilustrativă, Hîrbovăț nu are cântăriri în registru.'))
def tb_productie():
    return ptable('pr', 'Producție', [('Total fermă', '', P.productie())],
                  after=note('% ouat: zi = ouă ÷ găini la început de zi; săptămână și perioadă = ouă ÷ zile-găină. Găini = la început de zi (săptămâna: la începutul ei). Stare: ±3 pct față de ghid. % Cat. 1–3 = ouăle categoriei ÷ ouă.'))

# ---------------------------------------------------------------- Readout + GuideChart (S4 geometry: 540×216, drawn 1:1)
CW, CH = s4.CW, s4.CH
YT, YB = 10, CH - 22
WMIN = R.C26['sapt_start']                       # SAPT 20: the cycle's first age week at Hîrbovăț
def gchart(cid, label, wmax, series, ylo, yhi, yticks, yfmt, x0=40, guide=None, band=None, xticks=(WMIN, 30, 40, 50, WK)):
    """series: [{'cls': 'act'|'s2', 'pt': 'pt'|'pt2', 'pts': {week: value}}]; guide {week: v}; band (lo, hi) dicts. x = age weeks."""
    X1 = CW - 8
    def X(w): return round(x0 + (w - WMIN) * (X1 - x0) / (wmax - WMIN), 1)
    def Y(v): return round(YB - (v - ylo) / (yhi - ylo) * (YB - YT), 1)
    g = ''
    for i, t in enumerate(yticks):
        g += f'<line class="{"axis" if i == 0 else "grid"}" x1="{x0}" x2="{X1}" y1="{Y(t)}" y2="{Y(t)}"/><text class="ax" x="{x0 - 4}" y="{Y(t) + 3}" text-anchor="end">{yfmt(t)}</text>'
    if band:
        lo, hi = band; ws = sorted(lo)
        g += '<polygon class="band" points="' + ' '.join(f'{X(w)},{Y(hi[w])}' for w in ws) + ' ' + ' '.join(f'{X(w)},{Y(lo[w])}' for w in reversed(ws)) + '"/>'
    if guide:
        g += '<polyline class="guide" points="' + ' '.join(f'{X(w)},{Y(v)}' for w, v in sorted(guide.items())) + '"/>'
    for s in series:
        g += f'<polyline class="{s["cls"]}" points="' + ' '.join(f'{X(w)},{Y(v)}' for w, v in sorted(s['pts'].items())) + '"/>'
    g += f'<line class="today" x1="{X(WK)}" x2="{X(WK)}" y1="{YT - 2}" y2="{YB}"/><line class="cur" x1="{X(WK)}" x2="{X(WK)}" y1="{YT - 2}" y2="{YB}"/>'
    for s in series:
        g += f'<circle class="{s.get("pt", "pt")}" cx="{X(WK)}" cy="{Y(s["pts"][WK])}" r="3.5"/>'
    g += ''.join(f'<text class="ax" x="{X(w)}" y="{CH - 6}" text-anchor="middle">{w}</text>' for w in xticks)
    return (f'<svg class="ip-chart" viewBox="0 0 {CW} {CH}" role="group" aria-roledescription="grafic" tabindex="0" aria-label="{label}" '
            f'aria-describedby="{cid}-help {cid}-read"><g aria-hidden="true">{g}</g></svg>')
SW = {'act': '<svg viewBox="0 0 24 8" aria-hidden="true"><line class="act" x1="0" x2="24" y1="4" y2="4"/></svg>',
      's2': '<svg viewBox="0 0 24 8" aria-hidden="true"><line class="s2" x1="0" x2="24" y1="4" y2="4"/></svg>',
      'guide': '<svg viewBox="0 0 24 8" aria-hidden="true"><line class="guide" x1="0" x2="24" y1="4" y2="4"/></svg>',
      'band': '<svg viewBox="0 0 24 8" aria-hidden="true"><rect class="band" x="0" y="0" width="24" height="8"/></svg>'}
def legend(items, note=''):
    return '<div class="ip-chart__legend">' + ''.join(f'<span>{SW[k]}{t}</span>' for k, t in items) + (f'<span>{note}</span>' if note else '') + '</div>'
RLAMP = {'ok': '<span class="ip-rlamp ip-rlamp--ok">ÎN NORMĂ</span>', 'warn': '<span class="ip-rlamp ip-rlamp--warn">ABATERE</span>',
         'off': '<span class="ip-rlamp ip-rlamp--off">FĂRĂ DATE ÎN GHID</span>', None: ''}
def readout(cid, cap, rl, val, sub, chart, read, cols, rows, caption, leg):
    """chart None: no GuideChart (nothing to draw); the Readout keeps its sub-lines only."""
    tb = ('<table class="ip-sr"><caption>' + caption + '</caption><thead><tr>' + ''.join(f'<th scope="col">{c}</th>' for c in cols) + '</tr></thead><tbody>'
          + ''.join('<tr><th scope="row">' + r[0] + '</th>' + ''.join(f'<td>{c}</td>' for c in r[1:]) + '</tr>' for r in rows) + '</tbody></table>')
    ch = (f'{chart}\n<span class="ip-sr" id="{cid}-help">Săgețile stânga și dreapta mută săptămâna; Home și End: prima și ultima săptămână.</span>\n'
          f'<div class="ip-chart__read" id="{cid}-read" role="status" aria-live="polite">{read}</div>\n{tb}\n{leg}') if chart else read
    return sexify(f'''<div class="ip-readout s4-readout">
<div class="ip-readout__cap"><span>{cap}</span>{RLAMP[rl]}</div>
<div class="ip-readout__val">{val}</div>
<div class="ip-readout__sub">{sub}</div>
{ch}
</div>''')   # SX01: the value, sub-line, read line and legend signs = §72 icons (the hidden data table keeps its text)
HB, RB = '♀', '♂'   # SX01: legend keys; readout() draws them as §72 icons in readout-ink (was <b>♀</b>, a ~5 px fallback glyph)
def rpair(f, m, unit=''):  # ♀ · ♂ figures on the dark Readout: signs in readout-ink (rec-hen/rooster are <3:1 there, D58); readout() hooks them (SX01)
    return f'♀ {f}<small>{unit}</small> · ♂ {m}<small>{unit}</small>'

# ---------------------------------------------------------------- Stat + section helpers (also used by s14)
def stat(k, capt, val, *lines):
    c = s4.cap(k, capt) if k != 'sexerr' else f'<span class="ip-cap s4-capi"><span class="s4-ico" aria-hidden="true">{icon("sexerr", 14)}</span>{capt}</span>'
    return f'<div class="ip-stat">{c}<span class="ip-stat__val">{val}</span>' + ''.join(f'<span class="ip-stat__guide">{s4.gline(l)}</span>' for l in lines) + '</div>'
def stats(*items): return '<div class="ip-stats s4-stats">' + ''.join(items) + '</div>'
def top(hid, title, ro, st, start=False):  # start: the Readout keeps its content height (D165: a Readout without a chart)
    return f'<section class="s4-top{" s4t-top--start" if start else ""}" aria-labelledby="{hid}"><h2 class="ip-sr" id="{hid}">{title}</h2>{ro}{st}</section>'
def charts(*parts):  # one Readout per page section (craft): each chart its own section, two side by side
    return '<div class="s4t-split">' + ''.join(f'<section aria-labelledby="{h}"><h2 class="ip-sr" id="{h}">{t}</h2>{ro}</section>' for h, t, ro in parts) + '</div>'
def block(hid, title, note_, body):
    return f'<section class="ip-block" aria-labelledby="{hid}"><div class="ip-sechead"><h2 class="ip-t-heading" id="{hid}">{title}</h2><span class="ip-note">{note_}</span></div>{body}</section>'
def tbox(cap, inner, cls=''):  # DataTable in the system scroll region (SYS06 .ip-tablescroll, framed) with system row headers (§32)
    return (f'<div class="ip-tablescroll ip-tablescroll--framed" role="region" tabindex="0" aria-label="{cap}"><table class="ip-table ip-table--rowheads{(" " + cls) if cls else ""}">'
            f'<caption class="ip-sr">{cap}</caption>{inner}</table></div>')
def foot(note_, link=None):
    lk = f'<a class="ip-link s4t-lk" href="{link[1]}">{link[0]}</a>' if link else ''
    return f'<div class="s4t-foot"><span class="ip-note">{note_}</span>{lk}</div>'
def thsex(sx):
    return (f'<span class="ip-sex ip-sex--hen">♀</span><span class="ip-sr"> găini</span>' if sx == 'F'
            else f'<span class="ip-sex ip-sex--rooster">♂</span><span class="ip-sr"> cocoși</span>')
SXC = {'F': 's4t-hen', 'M': 's4t-roo'}
def wend(k): return min(R.week_dates(k)[1], R.TODAY)
CWEEKS = [k for k in range(WMIN, WK + 1)]

# ================================================================= 1. PĂSĂRI: table · charts (mortality, alive) · event log
ARR_DAY = P.ARR_DAY
mF = {k: R.deaths(ARR_DAY, wend(k), sx='F') / R.ARRIVED['F'] * 100 for k in CWEEKS}
mM = {k: R.deaths(ARR_DAY, wend(k), sx='M') / R.ARRIVED['M'] * 100 for k in CWEEKS}
assert (rnd(mF[WK], 2), rnd(mM[WK], 2)) == (5.66, 34.66)
MG = {k: R.guide_mort(k) for k in CWEEKS if R.guide_mort(k) is not None}   # ♀ only, counted from lay start (D161)
MORT_OK = mF[WK] <= MG[WK]
def ro_mort():
    rows = [[str(k), pct(mF[k], 2), pct(mM[k], 2), pct(MG[k], 2) if k in MG else DASH + ' fără date în ghid'] for k in CWEEKS]
    lbl = f'Mortalitate cumulată găini și cocoși din sosire, Hîrbovăț Seria 2026, sapt. {WMIN}–{WK}; ghid găini din intrarea în ouat, sapt. {min(MG)}–{max(MG)}'
    ch = gchart('mort', lbl, WK, [{'cls': 'act', 'pts': mF}, {'cls': 's2', 'pt': 'pt2', 'pts': mM}], 0, 40, [0, 10, 20, 30, 40], lambda t: f'{t}%', guide=MG)
    return readout('mort', f'MORTALITATE CUMULATĂ · HÎRBOVĂȚ · SAPT {WK}', 'ok' if MORT_OK else 'warn', rpair(n(mF[WK], 2), n(mM[WK], 2), '%'),
                   f'ghid ♀ {pct(MG[WK], 2)} (ghid din intrarea în ouat) · morți din sosire',
                   ch, f'SAPT {WK} · ♀ {pct(mF[WK], 2)} · ♂ {pct(mM[WK], 2)} · ghid ♀ {pct(MG[WK], 2)}',
                   ['SAPT', 'Găini ♀', 'Cocoși ♂', 'Ghid ♀'], rows, f'Mortalitate cumulată pe săptămânile de vârstă, % din păsările sosite (♀ {n(R.ARRIVED["F"])} · ♂ {n(R.ARRIVED["M"])}); morți reali, fără vânzări',
                   legend([('act', f'{HB} găini'), ('s2', f'{RB} cocoși'), ('guide', f'ghid {HB} din intrarea în ouat')], 'axa: SAPT de vârstă'))
aF = {k: R.alive_farm('F', wend(k)) for k in CWEEKS}; aM = {k: R.alive_farm('M', wend(k)) for k in CWEEKS}
assert (aF[WK], aM[WK]) == (12561, 794) and max(aF.values()) <= 30000
SOLD = {s: R.sold(R.CYCLE_START, R.TODAY, sx=s) for s in 'FM'}
def ro_alive():
    rows = [[str(k), n(aF[k]), n(aM[k])] for k in CWEEKS]
    lbl = f'Păsări în viață la sfârșitul săptămânii, găini și cocoși, Hîrbovăț Seria 2026, sapt. {WMIN}–{WK}; vânzări din {ddate(R.EMPTYING_SINCE)}'
    ch = gchart('alive', lbl, WK, [{'cls': 'act', 'pts': aF}, {'cls': 's2', 'pt': 'pt2', 'pts': aM}], 0, 30000, [0, 10000, 20000, 30000], lambda t: n(t), x0=52)
    return readout('alive', f'ÎN VIAȚĂ · HÎRBOVĂȚ · SAPT {WK}', None, rpair(n(aF[WK]), n(aM[WK])),
                   f'la sosire ♀ {n(R.ARRIVED["F"])} · ♂ {n(R.ARRIVED["M"])} · vânzări ♀ {n(-SOLD["F"])} · ♂ {n(-SOLD["M"])}',
                   ch, f'SAPT {WK} · ♀ {n(aF[WK])} · ♂ {n(aM[WK])}',
                   ['SAPT', 'Găini ♀', 'Cocoși ♂'], rows, 'Păsări în viață la sfârșitul fiecărei săptămâni de vârstă (SAPT 59: 18.11.2026)',
                   legend([('act', f'{HB} găini'), ('s2', f'{RB} cocoși')], 'axa: SAPT de vârstă'))

# --- event log (D1: the bird records of the period, the type chips filter it; SYS09 owner keys on every record, D70 B) ---
EVID = [0]
def acts(name):
    EVID[0] += 1; i = f'ev{EVID[0]}'
    dots = '<svg class="ip-icon" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="5" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="12" cy="19" r="1"/></svg>'
    return (f'<span class="ip-ev__acts"><a class="ip-btn ip-btn--ghost" href="#corecteaza-{i}">Corectează<span class="ip-sr"> {name}</span></a>'
            f'<button type="button" class="ip-btn ip-btn--ghost" aria-haspopup="dialog">Șterge<span class="ip-sr"> {name}</span></button></span>'
            f'<span class="ip-ev__more"><button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Acțiuni, {name}" aria-haspopup="menu" aria-expanded="false" aria-controls="{i}-m">{dots}</button>'
            f'<div class="ip-menu ip-ev__menu" hidden="hidden"><div class="ip-ev__items" id="{i}-m" role="menu" aria-label="Acțiuni, {name}" aria-describedby="{i}-n">'
            f'<a class="ip-menu__item" role="menuitem" tabindex="-1" href="#corecteaza-{i}">Corectează</a><button type="button" class="ip-menu__item" role="menuitem" tabindex="-1">Șterge</button></div>'
            f'<p class="ip-note" id="{i}-n">Corectarea și ștergerea cer un motiv.</p></div></span>')
def evrow(tl, title, meta, name, mod='', slot=True):
    return (f'<li class="ip-ev ip-ev--act">{tl}<span class="ip-ev__main"><span class="ip-ev__title">{title}</span><span class="ip-ev__meta">{meta}</span>{mod}</span>'
            f'{acts(name) if slot else ""}</li>')
LOG_FROM = R.TODAY - dt.timedelta(days=2)   # page 1 = the last three days (the illustrative log's Zilnic saves start 16.11)
LOGE = [e for e in R.EVENTS if e['record'] in ('mortality', 'sale') and R.D(e['at']) >= LOG_FROM]
assert not [c for c in PCULLS + PSEXERR if R.D(c['date']) >= LOG_FROM] and not [m for m in PMOVES if R.D(m['date']) >= LOG_FROM]
def log_row(e):
    d = R.D(e['at']); meta = f'<span class="ip-num">{R.ddt(e["at"])} · SAPT {R.sapt(d)}</span> · {e["user"]}'
    if e['record'] == 'mortality':
        f, m = R.deaths(d, sx='F'), R.deaths(d, sx='M'); k = sum(1 for s in SECTIONS if R.sd(d, s, 'F') or R.sd(d, s, 'M'))
        assert f'mortalitate ♀ {f} · ♂ {m}' in e['summary'] and f'{k} secțiuni' in e['summary'], e
        return evrow(tile('mortality'), f'Mortalitate · {k} secțiuni · {hen(n(-f))} · {roo(n(-m))}', meta, f'Mortalitate, {ddate(d)}')
    sec = R.sale_codes(e)  # DATA04 (D168): one record per sale form
    return evrow(tile('sale'), f'Vânzare · {sec} · {s4.sexpair(e["F"], e["M"])}', meta, f'Vânzare {sec}, {ddate(d)}')
def panel_pasari(hale=False):
    rows = [log_row(e) for e in LOGE]
    log = block('h-jurnal', 'Jurnal păsări', f'Mortalitate · Sacrificare · Vânzare · cele mai noi',
                f'<ul class="ip-list ip-list--act" aria-label="Jurnal păsări Hîrbovăț, {R.drange(LOG_FROM, R.TODAY)}">{"".join(rows)}</ul>'
                + foot(f'<span class="ip-num">{len(rows)} înregistrări · {R.drange(LOG_FROM, R.TODAY)}</span> · Zilnic: o înregistrare pe zi pentru secțiunile fermei',
                       ('Mai vechi', '#jurnal-pagina-2')))
    return (tb_pasari(hale) + '\n' + charts(('h-mort', 'Mortalitate cumulată', ro_mort()), ('h-alive', 'În viață', ro_alive())) + '\n' + log)

# ================================================================= 2. HRANĂ: table · g/pasăre vs guide · kg/zi, stock, deliveries
gF = {k: W(k)['g_F'] for k in CWEEKS if W(k)['g_F']}; gM = {k: W(k)['g_M'] for k in CWEEKS if W(k)['g_M']}
gG = {k: W(k)['guide']['feed_g_F'] for k in CWEEKS if W(k)['guide']['feed_g_F']}
assert min(gF) == WMIN + 1 and min(gM) == WMIN and 80 <= min(gF.values()) and max(gF.values()) <= 200 and max(gM.values()) <= 200
FDEV = (gF[WK] / gG[WK] - 1) * 100
FEED_OK = abs(FDEV) <= 10                                    # ±10 % (spec C2, as S4 Stats and S7)
def ro_hrana():
    band = ({k: v * 0.9 for k, v in gG.items()}, {k: v * 1.1 for k, v in gG.items()})
    rows = [[str(k), n(gF[k], 1) + ' g' if k in gF else DASH + ' fără găini', n(gM[k], 1) + ' g', n(gG[k]) + ' g'] for k in CWEEKS]  # ♀ from SAPT 21 (25–28.02)
    gm = W(WK)['guide']['feed_g_M']
    lbl = f'Hrană pe pasăre pe săptămână, găini față de ghid și cocoși, Hîrbovăț Seria 2026, sapt. {WMIN}–{WK}'
    ch = gchart('feed', lbl, WK, [{'cls': 'act', 'pts': gF}, {'cls': 's2', 'pt': 'pt2', 'pts': gM}], 80, 200, [80, 120, 160, 200], lambda t: f'{t}', guide=gG, band=band)
    return readout('feed', f'HRANĂ G/PASĂRE · HÎRBOVĂȚ · SAPT {WK}', 'ok' if FEED_OK else 'warn', rpair(n(gF[WK]), n(gM[WK]), 'g'),
                   f'ghid ♀ {n(gG[WK])} · ♂ {n(gm)} g · ♀ {sgn(FDEV, 1)}% · {R.drange(*span("w", WK))}',
                   ch, f'SAPT {WK} · ♀ {n(gF[WK], 1)} g · ghid {n(gG[WK])} g · ♂ {n(gM[WK], 1)} g',
                   ['SAPT', 'Găini ♀', 'Cocoși ♂', 'Ghid ♀'], rows, 'Hrană pe pasăre pe săptămâna de vârstă = kg ÷ zile-pasăre; ghid găini Ross 308 ouat',
                   legend([('act', f'{HB} găini'), ('s2', f'{RB} cocoși'), ('guide', f'ghid {HB}'), ('band', '±10%')], 'axa: SAPT de vârstă'))
def st_hrana():
    st = s4.stocks; b2, co = st['ft-breeder2'], st['ft-cocosi']
    kgF, kgM = R.feed_kg(R.TODAY, 'F'), R.feed_kg(R.TODAY, 'M')
    pk = {s: sum(W(k)[f'kg_{s}'] for k in CYC_WEEKS if W(k)) for s in 'FM'}   # D191: the Stats follow the period (the cycle)
    dl = [d for d in R.DELIVERIES if incy(d['date'])]; unp = [d for d in dl if not d['priced']]
    def sl(k): return lamp('warn', 'abatere') if st[k]['days_left'] < s4.STOCK_LIMIT else lamp('ok', 'în normă')
    return stats(
        stat('feed', 'Hrană azi', f'{pair(n(kgF, 1), n(kgM, 1))}<small>kg</small>', f'<span>{len(s4.FULL)} secțiuni · {s4.EMPTY_NOTE}</span>', f'<span class="ip-num">{ddate(R.TODAY)} · SAPT {WK}</span>'),
        stat('feed', 'Hrană pe zi · 3 zile', f'{pair(n(b2["avg_daily_kg_last3"]), n(co["avg_daily_kg_last3"]))}<small>kg</small>', f'<span class="ip-num">{R.drange(R.TODAY - dt.timedelta(days=2), R.TODAY)}</span>', 'baza zilelor rămase'),
        stat('delivery', f'Stoc Breeder 2 {F}', f'{n(b2["days_left"], 1)}<small>zile</small>', f'<span class="ip-num">până pe {ddate(b2["run_out_date"])}</span>{sl("ft-breeder2")}', f'<span class="ip-num">{n(b2["stock_kg"], 1)} kg</span>'),
        stat('delivery', f'Stoc Cocoși {M}', f'{n(co["days_left"], 1)}<small>zile</small>', f'<span class="ip-num">până pe {ddate(co["run_out_date"])}</span>{sl("ft-cocosi")}', f'<span class="ip-num">{n(co["stock_kg"], 1)} kg</span>'),
        stat('feed', 'Consum · ciclu', f'{pair(n(pk["F"]), n(pk["M"]))}<small>kg</small>', f'<span class="ip-num">{R.drange(CYC_FROM, P_TO)}</span>', f'<span class="ip-num">SAPT {CYC_WEEKS[-1]}–{WK}</span>'),
        stat('delivery', 'Livrări · ciclu', f'{n(sum(d["qty_kg"] for d in dl))}<small>kg</small>', f'{num(len(dl))} livrări · {num(len(unp))} fără preț {lamp("info", "informație")}',
             f'<span class="ip-num">{n(sum(d["total_lei"] for d in dl if d["priced"]))} lei</span>'))
def panel_hrana():
    return tb_hrana() + '\n' + top('h-key', 'Hrană față de ghid', ro_hrana(), st_hrana())

# ================================================================= 3. GREUTATE: table (weighings) · Readout (one weighing, no chart)
WG = {w['sex']: w for w in PWEIGH}
assert len(PWEIGH) == 2 and {w['section'] for w in PWEIGH} == {'H5.1'} and R.weigh_guide(R.sapt(PWEIGH[0]['date']), 'F') is None
def ro_greutate():
    w = PWEIGH[0]; k = R.sapt(w['date'])
    return readout('wt', f'GREUTATE · HÎRBOVĂȚ · SAPT {k}', 'off', rpair(n(WG['F']['avg_weight_g']), n(WG['M']['avg_weight_g']), 'g'),
                   f'cântărit {ddate(w["date"])} · {w["section"]} · ilustrativ',
                   None, '<div class="ip-readout__sub">o singură cântărire în ciclu: graficul începe de la a doua</div>'
                   '<div class="ip-readout__sub">ghidul de creștere are greutăți până la SAPT 25</div>', [], [], '', '')
def st_greutate():
    w = PWEIGH[0]
    def p(key, dec=1): return pair(n(WG['F'][key], dec), n(WG['M'][key], dec))
    return stats(
        stat('weighing', 'Uniformitate', f'{p("uniformity_pct", 0)}<small>%</small>', NOGD, f'păsări în {num("±10%")} din medie'),
        stat('weighing', 'Coeficient de variație', f'{p("cv_pct")}<small>%</small>', NOGD, f'<span class="ip-num">{w["section"]} · {ddate(w["date"])}</span>'),
        stat('weighing', 'Probă cântărită', p('sample_size', 0), f'<span class="ip-num">1 din {len(s4.FULL)} secțiuni cu păsări</span>',
             f'din păsări {pair(pct(WG["F"]["sample_size"] / R.alive(w["section"], "F", w["date"]) * 100), pct(WG["M"]["sample_size"] / R.alive(w["section"], "M", w["date"]) * 100))}'),
        stat('weighing', 'Cântăriri · ciclu', f'{len({c["date"] for c in CWEIGH})}<small>zi</small>', f'<span class="ip-num">SAPT {R.sapt(w["date"])} · {ddate(w["date"])}</span>', 'singura cântărire din ciclu'))
def panel_greutate():
    return tb_greutate() + '\n' + top('h-key', 'Greutate', ro_greutate(), st_greutate(), start=True)

# ================================================================= 4. PRODUCȚIE: table · % ouat vs guide · egg weight
eW = {k: W(k)['egg_g'] for k in CWEEKS if W(k)['egg_g']}
assert max(eW.values()) < 80 and min(eW) == 23
def ro_egg():
    rows = [[str(k), n(eW[k], 1) + ' g' if k in eW else DASH + ' fără ouă', DASH + ' fără date în ghid'] for k in CWEEKS]
    lbl = f'Masa medie a oului pe săptămână, Hîrbovăț Seria 2026, sapt. {min(eW)}–{WK}; ghidul nu are masa oului'
    ch = gchart('egg', lbl, WK, [{'cls': 'act', 'pts': eW}], 0, 80, [0, 40, 80], lambda t: f'{t} g')
    return readout('egg', f'MASĂ OU · HÎRBOVĂȚ · SAPT {WK}', 'off', f'{n(eW[WK], 1)}<small>g</small>',
                   f'azi {n(R.day(R.TODAY)["egg_g"], 1)} g · SAPT {WK - 1}: {n(eW[WK - 1], 1)} g · medie pe ouă',
                   ch, f'SAPT {WK} · {n(eW[WK], 1)} g · ghid {DASH} fără date în ghid',
                   ['SAPT', 'Masă ou', 'Ghid'], rows, f'Masa medie a oului pe săptămâna de vârstă (medie după ouă); primele ouă în SAPT {min(eW)}',
                   legend([('act', 'efectiv')], f'axa: SAPT de vârstă · ouă din SAPT {min(eW)}'))
def panel_productie():
    return tb_productie() + '\n' + charts(('h-lay', 'Ouat față de ghid', s4.readout()), ('h-egg', 'Masă ou', ro_egg()))

# ================================================================= 0. REZUMAT (the S4 default board, written by s4_build.py)
def main_rezumat():   # S4NG: s4s.panel_default draws the Rezumat table itself (shared with the state boards and S9/S10)
    return s4.shell(main('rezumat', s4.panel_default()))

# ================================================================= CSS (screen only, tokens only) + write
CSS = s4.CSS + '''
.s4t-split{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:var(--space-4);align-items:stretch}
.s4t-split > section{display:flex;flex-direction:column;min-width:0}
.s4t-split > section > .ip-readout{flex:1}
.s4t-top--start{align-items:start}
.s4-panel .ip-table .ip-lamp{white-space:nowrap}
.s4t-hen{color:var(--rec-hen)}
.s4t-roo{color:var(--rec-rooster)}
.s4t-gnote{color:var(--ink-secondary)}
.s4t-sub{color:var(--ink-secondary);font-family:var(--font-mono);font-size:var(--ip-fs-small)}
.s4t-state{display:inline-flex;align-items:center;gap:var(--space-2)}
.s4t-foot{display:flex;justify-content:space-between;align-items:center;gap:var(--space-4)}
.s4t-lk{display:inline-flex;align-items:center;min-height:var(--space-6);white-space:nowrap;flex:none}
@container s4main (max-width: 999.98px){.s4t-split{grid-template-columns:minmax(0,1fr)}} /* FIN16: the narrow page (beside a SidePanel): one column, so the Jurnal keys stay inline (SYS09 >= 480px) */'''

B = 'S4-ferma--'
BOARDS = [
 (B + 'pasari', 'S4 · Fermă · Păsări', lambda: page('pasari', panel_pasari())),
 (B + 'pasari--pe-hale', 'S4 · Fermă · Păsări · pe hale', lambda: page('pasari', panel_pasari(hale=True))),
 (B + 'hrana', 'S4 · Fermă · Hrană', lambda: page('hrana', panel_hrana())),
 (B + 'greutate', 'S4 · Fermă · Greutate', lambda: page('greutate', panel_greutate())),
 (B + 'productie', 'S4 · Fermă · Producție', lambda: page('productie', panel_productie())),
]
if __name__ == '__main__':
    for stem, title, fn in BOARDS:
        EVID[0] = 0
        s4.write(stem, title, fn(), H=HEIGHTS.get(stem, 1874), css=CSS)
    if '--canvas' in sys.argv[2:]:  # S4N: refresh the 4 entries' h/title in place (hand-placed row: x and y kept)
        import canvas_rows as CR
        c = CR.load(s4.ROOT); old = [b for b in BOARDS if f'{b[0]}.dc.html' in c['boards']]; new = [b for b in BOARDS if b not in old]
        CR.place(c, 'S4-ferma-rezumat', [(stem, 1440, HEIGHTS.get(stem, 1874), title) for stem, title, _ in old])  # in place
        for stem, title, _ in new:  # D165: a new state board goes after the row's last board, one step (w + 80) to the right
            row = [v for k, v in c['boards'].items() if CR.in_row(k, 'S4-ferma-rezumat') or k.startswith(B)]
            x0 = max(v['x'] + v['w'] for v in row) + CR.GAP_X
            CR.place(c, 'S4-ferma-rezumat', [(stem, 1440, HEIGHTS.get(stem, 1874), title)], x0=x0)
            f = f'{stem}.dc.html'; c['order'].remove(f); c['order'].insert(c['order'].index(f'{B}productie.dc.html') + 1, f)
        CR.save(s4.ROOT, c)
