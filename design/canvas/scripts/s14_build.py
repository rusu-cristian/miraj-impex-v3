"""S14 · Rapoarte · Producție (desktop 1440): spec D3 weekly table per farm › section (hall) by guide week, vs the guide,
with Export. Owner view on 16.11.2026 (Hîrbovăț · Seria 2026, period "Ciclul curent"), plus state boards.
python3 -I s14_build.py <repo_root> [heights.json] [--canvas]   (--canvas also writes the S14 board entries into canvas.json)
Reuses: s11_build (Rapoarte shell: sidebar with Rapoarte current, section heads, Export in the FilterBar end), s4t_build
(S4 Producție tab: Readout + GuideChart, Stat, framed DataTable region with §32 row headers, table foot, formats .5 up),
s4s_build (shell, skeletons, empty box, no-guide banner/readout, Seria 2025 readout + stats, write()).
Eggs are recorded per hall (spec C1: Ouă on the hall's card), so the table's sections are the halls H1, H3, H4; their
mortality and feed sum / weight the compartments (D38). Every figure comes from the extracts; nothing is invented.
Writes design/canvas/project/S14-rapoarte-productie[--<state>].dc.html and /tmp/s4s/<stem>.html (plain wrappers)."""
import sys, json, pathlib, importlib.util, collections as C, datetime as dt

HERE = pathlib.Path(__file__).resolve().parent
def _load(name):
    sp = importlib.util.spec_from_file_location(name, HERE / f'{name}.py')
    m = importlib.util.module_from_spec(sp); sp.loader.exec_module(m); return m
t4 = _load('s4t_build')   # reads sys.argv[1] (repo root)
s11 = _load('s11_build')  # reads sys.argv[2] as its heights only if given (harmless: it writes nothing on import)
s4 = t4.s4
sys.path.insert(0, str(HERE)); import canvas_rows as CR
import skel as SKL
ROOT, HP, CANVAS = CR.args('s14_heights.json')  # heights default: s14_heights.json
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}

n, sgn, pct, num, ddate, ddm, lamp, icon = t4.n, t4.sgn, t4.pct, t4.num, s4.ddate, t4.ddm, s4.lamp, s4.icon
S, GL, WK, TODAY, HALLS, SECTIONS = t4.S, t4.GL, t4.WK, t4.TODAY, t4.HALLS, t4.SECTIONS
F, M, DASH, NOGD = s4.F, s4.M, t4.DASH, t4.NOGD
NOG = s4.NOG  # 'ghid: – fără ghid'
EW = t4.EW
NBSP = ' '

# ================================================================= data · Seria 2026 at Hîrbovăț (sample, D27)
GF = {(r['cycle_week'], r['sex']): r['feed_g'] for r in GL['rows']}
WEEKS_ALL = sorted({w for (h, w) in EW if h == 'all'})
assert WEEKS_ALL[0] == 3 and WEEKS_ALL[-1] == WK and len(WEEKS_ALL) == 38
SHOW = list(range(WK, WK - 8, -1))  # SAPT 40 … 33, newest first
LASTD = {}
for r in S['daily']: LASTD[r['cycle_week']] = max(LASTD.get(r['cycle_week'], ''), r['date'])
HW = C.defaultdict(C.Counter)  # (hall|'all', week) → hens at week end, deaths ♀/♂, ♀ feed kg and hen-days with grams, missing ♀ rows
MISS = []
for r in S['daily']:
    h = r['section'].split('.')[0]; w = r['cycle_week']
    for k in (h, 'all'):
        x = HW[(k, w)]
        x['mort' + r['sex']] += r['mortality']
        if r['date'] == LASTD[w] and r['sex'] == 'F': x['hens'] += r['alive_end']
        if r['sex'] == 'F':
            if r['g_per_bird'] is None: x['missF'] += 1
            else: x['kgF'] += r['kg']; x['hdF'] += r['alive_end']
    if r['g_per_bird'] is None: MISS.append(r)
assert [(m['date'], m['section'], m['sex']) for m in MISS] == [('2026-11-16', 'H3.2', 'M'), ('2026-11-16', 'H4.2', 'F')]
assert HW[('all', WK)]['hens'] == s4.tot['F'] == 25252 and LASTD[WK] == TODAY
assert all(HW[('all', w)]['mortF'] == t4.mw[w]['F'] and HW[('all', w)]['mortM'] == t4.mw[w]['M'] for w in t4.mw)
def gfem(k, w): x = HW[(k, w)]; return x['kgF'] * 1000 / x['hdF']  # ♀ g/pasăre weighted by hens (D38), days with rows (D46)
for w in (37, 38, 39, 40):  # = S11 Consum week rows (♀ by hens)
    assert n(gfem('all', w), 1) == n(s11.g_avg(s11.agg[(w, 'all', 'F')], True), 1), w
# farm weeks vs guide (±3 pct) and the hall-weeks in abatere over the whole cycle
GWEEKS = [w for w in WEEKS_ALL if EW[('all', w)]['guide_lay_pct'] is not None]
assert (GWEEKS[0], GWEEKS[-1], len(GWEEKS)) == (7, 40, 34)
FARM_DEV = [w for w in GWEEKS if abs(EW[('all', w)]['diff_pts']) > 3]
HALL_DEV = [(h, w) for w in GWEEKS for h in HALLS if abs(EW[(h, w)]['diff_pts']) > 3]
assert FARM_DEV == [] and HALL_DEV == [('H3', 40)]
eggs_cyc = sum(r['total_eggs'] for r in S['eggs_daily']); hatch_cyc = sum(r['hatching_eggs'] for r in S['eggs_daily'])
assert eggs_cyc == t4.eggs_cycle == 5301053
# ♀ feed over the cycle vs the guide over the same rows (D46 like with like)
fr = [r for r in S['daily'] if r['sex'] == 'F' and r['g_per_bird'] is not None]
FEED_F = sum(r['kg'] for r in fr) * 1000 / sum(r['alive_end'] for r in fr)
FEED_GF = sum(GF[(r['cycle_week'], 'F')] * r['alive_end'] for r in fr) / sum(r['alive_end'] for r in fr)
FEED_DEV = (FEED_F / FEED_GF - 1) * 100
# S14P: egg weight over the cycle, weighted by eggs (daily hall rows; the 6 rows without a weight have no eggs)
_ew = [r for r in S['eggs_daily'] if r['total_eggs']]
assert all(r['avg_egg_weight_g'] is not None for r in _ew) and sum(r['total_eggs'] for r in _ew) == eggs_cyc
EGG_W = sum(r['avg_egg_weight_g'] * r['total_eggs'] for r in _ew) / eggs_cyc
EGG_WK = sorted({r['cycle_week'] for r in _ew}); assert (EGG_WK[0], EGG_WK[-1]) == (4, WK)
# S14P: peak lay over the cycle (farm weeks), for the no-guide board where the Readout already gives SAPT 40
PEAK = max((w for w in WEEKS_ALL if EW[('all', w)]['total_eggs']), key=lambda w: EW[('all', w)]['lay_pct'])
assert (PEAK, EW[('all', PEAK)]['lay_pct']) == (18, 90.6)
cyc = s4.cum

# ================================================================= shell (Rapoarte current, as S11)
def sidebar(series='2026'):
    sb = s4.sidebar(series=series)
    sb = sb.replace('class="ip-nav is-current" href="#serie" aria-current="page"', 'class="ip-nav" href="#serie"')
    sb = sb.replace('class="ip-nav" href="#rapoarte"', 'class="ip-nav is-current" href="#rapoarte" aria-current="page"')
    assert sb.count('is-current') == 1
    return sb
assert sidebar() == s11.sidebar('Proprietar')

def header(mode='default'):
    crumbs = '<nav class="ip-crumbs" aria-label="Cale"><a href="#rapoarte">Rapoarte</a><span aria-hidden="true">›</span><span aria-current="page">Producție</span></nav>'
    acts = ''
    if mode == 'past':
        c25 = s4.c25
        meta = (f'{lamp("off", "închis")}<span class="ip-note">Seria 2025 · Hîrbovăț · început dedus {num(ddate(c25["start_date"]))} · încheiat prin golire {num(ddate(s4.PREV_END))} · '
                f'{num(f"{s4.PREV_WEEKS} săpt.")} · date din {num(ddate(c25["data_start_date"]))}</span>')
        acts = '<div class="ip-row"><a class="ip-btn ip-btn--secondary" href="#ciclu-seria-2026">Înapoi la ciclul curent</a></div>'
    else:
        gtag = '<span class="ip-tag ip-tag--neutral">fără ghid</span>' if mode == 'noguide' else '<span class="ip-tag ip-tag--neutral">Ross 308 ouat</span>'
        meta = (f'<span class="ip-tag ip-tag--neutral">Ouat</span>{gtag}<span class="ip-note">Seria 2026 · Hîrbovăț {num(f"SAPT {WK}")}, ciclu din {num("10.02.2026")} · '
                f'Telița: creștere, fără producție</span>')
    return f'''<header class="ip-pagehead">
<div>
{crumbs}
<h1>Producție</h1>
<div class="ip-row ip-pagehead__meta">{meta}</div>
</div>
{acts}
</header>'''

TEL_WHY = 'creștere, fără ouă'  # why Telița can't be chosen in Producție
def farm_select():  # laying farms only; Telița (growth) is listed but disabled with its reason
    # FIN12 (D130): an option disabled with a reason makes Fermă the section 53 Select (listbox hidden while closed)
    O = '<span class="ip-select__txt"><span class="ip-select__top">'
    opts = (f'<li class="ip-menu__item ip-select__opt" role="option" id="f-farm-o1" aria-selected="true"><span class="ip-select__chk" aria-hidden="true">{icon("check", 16)}</span>{O}Hîrbovăț</span></span></li>'
            f'<li class="ip-menu__item ip-select__opt" role="option" id="f-farm-o2" aria-selected="false" aria-disabled="true"><span class="ip-select__chk" aria-hidden="true"></span>{O}Telița <span class="ip-note">{TEL_WHY}</span></span></span></li>')
    return ('<div class="s11-farm"><span class="ip-label" id="f-farm-l">Fermă</span><div class="ip-select s11-select">'
            '<div class="ip-input ip-select__btn" id="f-farm" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="false" aria-controls="f-farm-list" aria-labelledby="f-farm-l">'
            f'<span class="ip-select__val">Hîrbovăț</span>{icon("chev", 16)}</div>'
            f'<ul class="ip-menu ip-select__menu" role="listbox" id="f-farm-list" aria-labelledby="f-farm-l" tabindex="-1" hidden="hidden">{opts}</ul></div></div>')
def filterbar(mode='default'):
    if mode == 'empty':
        itv = f'<span class="ip-interval" role="group" aria-label="Interval">{s4.datef("iv-a", "de la", "10.02.2026")}{s4.datef("iv-b", "până la", "28.02.2026")}</span>'
        return f'<div class="ip-filterbar s11-filterbar" role="group" aria-label="Filtre">{s4.seg("Perioadă", s4.PERIODS, 3)}{itv}{farm_select()}{s11.export(True)}</div>'
    per = ['Tot ciclul'] + s4.PERIODS[1:] if mode == 'past' else s4.PERIODS
    return f'<div class="ip-filterbar s11-filterbar" role="group" aria-label="Filtre">{s4.seg("Perioadă", per, 0)}{farm_select()}{s11.export()}</div>'

# ================================================================= top: Readout (S4 Producție GuideChart) + period Stats
def stats_cur(ng=False):
    st, sx = t4.stat, t4.stats
    # S14P fix 2: at 1280 the lamp wrapped alone ("● abatere" as the last line); it now runs in the text, kept with "−4,0 pct"
    # (Chrome breaks before an inline-flex lamp even after a no-break space, so the pair sits in a nowrap span)
    dev_line = (f'hale: H3 {num(f"SAPT {WK}")} · <span class="ip-nowrap">{num(sgn(EW[("H3", WK)]["diff_pts"]) + " pct")} {lamp("warn", "abatere")}</span>')
    pk = EW[('all', PEAK)]
    s_dev = (st('eggs', 'Vârf de ouat', f'{n(pk["lay_pct"], 1)}<small>%</small>', NOG, f'{num(f"SAPT {PEAK} · " + span(pk["from"], pk["to"]))} · fermă') if ng else
             st('eggs', 'Săptămâni în abatere', f'{len(FARM_DEV)} din {len(GWEEKS)}<small>săpt. · fermă</small>',
                f'ghid {num("±3 pct")} · {num(f"SAPT{NBSP}{GWEEKS[0]}–{WK}")}', 'S14DEV').replace('<span>S14DEV</span>', f'<span>{dev_line}</span>', 1))
    assert ng or s_dev.count(dev_line) == 1
    fd = abs(FEED_DEV) <= 10
    g_feed = NOG if ng else f'ghid {num(n(FEED_GF, 1) + " g")} · {num(sgn(FEED_DEV, 1) + "%")} {lamp("ok" if fd else "warn", "în normă" if fd else "abatere")}'
    first = t4.first_egg
    return sx(
        st('eggs', 'Ouă în ciclu', n(eggs_cyc), f'{num("01.03–16.11.2026")} · Cat.{NBSP}1–⁠3{NBSP}+{NBSP}incubabile', f'primul ou {num(ddate(first))} · {num(f"SAPT {t4.cw(first)}")}'),
        s_dev,
        st('eggs', 'Incubabile în ciclu', f'{n(hatch_cyc / eggs_cyc * 100, 1)}<small>%</small>', NOG if ng else NOGD, f'{num(n(hatch_cyc))} din {num(n(eggs_cyc))} ouă'),
        st('eggs', 'Masă medie ou în ciclu', f'{n(EGG_W, 1)}<small>g</small>', NOG if ng else NOGD, f'medie pe ouă · {num(f"SAPT {EGG_WK[0]}–{WK}")}'),
        st('mortality', 'Mortalitate cumulată', f'{s4.pair(n(cyc["mortality_pct"]["F"], 2), n(cyc["mortality_pct"]["M"], 2))}<small>%</small>', NOG if ng else NOGD,
           f'din 10.02.2026 {s4.pair(n(-cyc["removals"]["F"]["mortality"]), n(-cyc["removals"]["M"]["mortality"]))}'),
        st('feed', f'Hrană {F} în ciclu', f'{n(FEED_F, 1)}<small>g/pasăre/zi</small>', g_feed, f'medie pe găini · fără {F} H4.2 {num("16.11")}{NBSP}(lipsă)'))  # S14P fix 2: "(lipsă)" no longer alone at 1280

def top(mode='default'):
    if mode == 'noguide': return t4.top('h-key', 'Producție, ciclul curent', s4.readout(True), stats_cur(True))
    if mode == 'past': return t4.top('h-key', 'Producție, Seria 2025', readout_past(), stats_past())
    return t4.top('h-key', 'Producție față de ghid, ciclul curent', t4.ro_productie(), stats_cur())

# ================================================================= the weekly table (spec D3), Seria 2026
# FIN11G (§54): the desktop weekly table's 14 column shares at 1440 (as content-sized before), declared on the head the
# loading skeleton reuses (desktop only: the phone table scrolls at its content width)
M_WEEKS = [137.75, 46.97, 54.8, 50.42, 39.14, 82.13, 64.19, 54.8, 39.14, 56.8, 39.5, 39.5, 69.36, 43.81]  # fix 1: loaded min-content (px)
W_WEEKS = [None, 6.1, 6.8, 6.4, 5.5, 9.2, 7.6, 6.8, 5.5, 7.0, 5.5, 5.5, 8.1, 5.9]  # None: Săptămâna · hala takes the rest
def th_head(ng=False, desk=False):
    eg, mo, fe = t4.thico('eggs'), t4.thico('mortality'), t4.thico('feed')
    sx = lambda s: f'<th scope="col" class="num">{t4.thsex(s)}</th>'
    if ng:
        h1 = (f'<tr><th scope="col" rowspan="2">Săptămâna · hala</th><th scope="col" rowspan="2" class="num">Găini</th><th scope="col" rowspan="2" class="num">{eg}Ouă</th>'
              '<th scope="col" rowspan="2" class="num">Ouat</th><th scope="colgroup" colspan="2">Incubabile</th><th scope="col" rowspan="2" class="num">Masă ou</th>'
              f'<th scope="colgroup" colspan="2">{mo}Mortalitate</th><th scope="col" rowspan="2" class="num">{fe}Hrană {t4.thsex("F")} g/pasăre</th></tr>')
        h2 = f'<tr><th scope="col" class="num">Ouă</th><th scope="col" class="num">%</th>{sx("F")}{sx("M")}</tr>'
    else:
        h1 = (f'<tr><th scope="col" rowspan="2">Săptămâna · hala</th><th scope="col" rowspan="2" class="num">Găini</th><th scope="col" rowspan="2" class="num">{eg}Ouă</th>'
              '<th scope="colgroup" colspan="4">Ouat</th><th scope="colgroup" colspan="2">Incubabile</th><th scope="col" rowspan="2" class="num">Masă ou</th>'
              f'<th scope="colgroup" colspan="2">{mo}Mortalitate</th><th scope="colgroup" colspan="2">{fe}Hrană {t4.thsex("F")} g/pasăre</th></tr>')
        h2 = ('<tr><th scope="col" class="num">Efectiv</th><th scope="col" class="num">Ghid</th><th scope="col" class="num">Față de ghid</th><th scope="col">Stare</th>'
              f'<th scope="col" class="num">Ouă</th><th scope="col" class="num">%</th>{sx("F")}{sx("M")}<th scope="col" class="num">Efectiv</th><th scope="col" class="num">Ghid</th></tr>')
    return SKL.widths(f'<thead>{h1}{h2}</thead>', W_WEEKS, M_WEEKS) if desk and not ng else f'<thead>{h1}{h2}</thead>'

def span(d0, d1):  # "10.11–16.11", "27.10–02.11" (year in the section note; FIN09 range form)
    return f'{ddm(d0)}–{ddm(d1)}'
def mortc(v, sx): return f'<td class="num {t4.SXC[sx]}">{n(-v) if v else "0"}</td>'
def cells(k, w, ng=False):
    e = EW[(k, w)]; x = HW[(k, w)]
    o = f'<td class="num s4t-hen">{n(x["hens"])}</td><td class="num">{n(e["total_eggs"])}</td><td class="num">{pct(e["lay_pct"])}</td>'
    if not ng:
        stt = lamp('warn', 'abatere') if abs(e['diff_pts']) > 3 else lamp('ok', 'în normă')
        o += f'<td class="num ip-muted">{pct(e["guide_lay_pct"])}</td><td class="num">{sgn(e["diff_pts"])} pct</td><td>{stt}</td>'
    o += (f'<td class="num">{n(e["hatching_eggs"])}</td><td class="num">{pct(e["hatching_pct"])}</td><td class="num">{n(e["avg_egg_weight_g"], 1)} g</td>'
          f'{mortc(x["mortF"], "F")}{mortc(x["mortM"], "M")}<td class="num s4t-hen">{n(gfem(k, w), 1)}</td>')
    if not ng: o += f'<td class="num ip-muted">{GF[(w, "F")]}</td>'
    return o

def tb_weeks(ng=False, phone=False):
    body = ''
    for w in SHOW:
        a = EW[('all', w)]
        body += (f'<tbody><tr class="is-group"><th scope="rowgroup"><span class="ip-num">SAPT {w}</span> <span class="s4t-sub">{span(a["from"], a["to"])}</span>'
                 f'<span class="ip-sr">, Hîrbovăț</span></th>{cells("all", w, ng)}</tr>')
        for h in HALLS:
            body += f'<tr><th scope="row" class="ip-table__child">{h}<span class="ip-sr">, SAPT {w}</span></th>{cells(h, w, ng)}</tr>'
        body += '</tbody>'
    cap = f'Producție Hîrbovăț pe săptămâni și hale, SAPT {SHOW[-1]}–{WK}' + ('' if ng else ', față de ghid')
    t = t4.tbox(cap, th_head(ng, desk=not phone) + body)
    note = (f'Hîrbovăț · {num(f"SAPT {SHOW[-1]}–{WK}")} ({num(ddm(EW[("all", SHOW[-1])]["from"]) + "–" + ddate(TODAY))}) din {num(f"SAPT {WEEKS_ALL[0]}–{WK}")} · cele mai noi întâi · '
            + ('ciclul nu are ghid, nu se compară' if ng else f'ghid Ross 308 ouat · toleranță ouat {num("±3 pct")}'))
    fnote = (f'Rândul săptămânii = ferma; halele dedesubt (ouăle se înregistrează pe hală). {num(f"Sapt. {WK}")} include azi, {num(ddate(TODAY))}. Găini = în viață la sfârșitul săptămânii. '
             f'Ouat = ouă (Cat.\u00a01–\u20603 + incubabile, fără ouăle de jos) ÷ găini la începutul zilei. Mortalitate = morți; sacrificările sunt în Fermă › Păsări. '
             f'Hrană {F} = medie pe găini, din zilele cu rânduri; {num(f"sapt. {WK}")} fără H4.2 {F} {num("16.11")} (lipsă).'
             + ('' if ng else f' Incubabile, masa oului și mortalitatea: – fără date în{NBSP}ghid.'))  # S14P fix 2: "ghid." no longer alone at 1280
    return t4.block('h-psapt', 'Pe săptămâni și hale', note, t + t4.foot(fnote, (f'Vezi toate săptămânile ({len(WEEKS_ALL)})', '#productie-saptamani')))

# ================================================================= Seria 2025 (real previous-flock data, D51/D52/D54)
PD, PW = s4.PD, s4.PW
pw = {(r['hall'], r['cycle_week']): r for r in PW['rows']}
PSHOW = list(range(s4.PFIN, s4.PFIN - 8, -1))  # SAPT 47 … 40
DEP = PD['meta']['depopulation_start']
pm = C.defaultdict(C.Counter)  # deaths outside the emptying days (D52/D54), and how many such days the week has
for r in PD['rows']:
    if r['depopulation']: continue
    for k in (r['hall'], 'all'):
        pm[(k, r['cycle_week'])]['F'] += r['mortality_F']; pm[(k, r['cycle_week'])]['M'] += r['mortality_M']; pm[(k, r['cycle_week'])]['days'] += 1
# a week made only of emptying days: "pierit" there holds the birds taken out, so deaths are unknown ("–"), not 0
PUNK = sorted((k, w) for (k, w) in pw if w in range(PSHOW[-1], s4.PFIN + 1) and not pm[(k, w)]['days'])
assert PUNK == [('H1', 46), ('H3', 46), ('H4', 47), ('all', 47)], PUNK
def pmortc(k, w, sx):
    if not pm[(k, w)]['days']:
        return f'<td class="num {t4.SXC[sx]}"><span class="ip-muted">{DASH}</span><span class="ip-sr"> necunoscut, golire</span></td>'
    return mortc(pm[(k, w)][sx], sx)
assert sum(v['F'] for (k, _), v in pm.items() if k == 'all') == s4.PMORT['F']
# S14P fix 1: the Seria 2025 Stat "Masă medie ou · SAPT 47" repeated its table row; it becomes the cycle's egg-weighted
# mean egg weight (daily hall rows with a weight; the 15 rows of 04–08.04.2025 have none: "Masa" 0 in the sheet = null)
_pw = [r for r in PD['rows'] if r['avg_egg_weight_g'] is not None]
PEGG_NOW = sum(r['total_eggs'] for r in PD['rows'] if r['avg_egg_weight_g'] is None)
PEGG_W = sum(r['avg_egg_weight_g'] * r['total_eggs'] for r in _pw) / sum(r['total_eggs'] for r in _pw)
PEGG_WK = sorted({r['cycle_week'] for r in _pw if r['total_eggs']})
assert PEGG_NOW == 9314 and sum(r['total_eggs'] for r in _pw) + PEGG_NOW == PD['meta']['totals']['total_eggs'] == 4606935
assert (PEGG_WK[0], PEGG_WK[-1]) == (5, s4.PFIN) == (5, 47) and n(PEGG_W, 1) == '62,0' and abs(PEGG_W - 61.9677) < 1e-4
OLD_EGGW = '<span class="ip-stat__val">69,9<small>g</small></span>'
def stats_past():
    sp = s4.stats_past()
    old = [d for d in sp.split('\n') if 'Masă medie ou · SAPT' in d]
    assert len(old) == 1 and OLD_EGGW in old[0], old
    new = t4.stat('eggs', 'Masă medie ou în ciclu', f'{n(PEGG_W, 1)}<small>g</small>', NOGD,
                  f'medie pe ouă · {num(f"SAPT {PEGG_WK[0]}–{PEGG_WK[-1]}")}', f'fără {num(n(PEGG_NOW))} ouă fără masă')  # two lines: one wrapped line left "fără masă" alone
    out = sp.replace(old[0], new); assert out.count('Masă medie ou') == 1 and 'SAPT 5: 47,0 g' not in out
    # S14P fix 2: no word alone on a Stat's last line at 1280 (S4's own stats_past untouched)
    for a, z in ((f' · {s4.PDAYS} zile</span>', f' · {s4.PDAYS}{NBSP}zile</span>'), (', nu se calculează</span>', f', nu se{NBSP}calculează</span>')):
        assert out.count(a) == 1, a; out = out.replace(a, z)
    return out
def readout_past():  # S14P fix 2: "după 22.01)" stays together on the Readout's sub line (S4's readout_past untouched)
    ro = s4.readout_past(); a = f' fără ouă după {s4.ddate(s4.PLAST["H1"])[:5]})</div>'
    assert ro.count(a) == 1 and a.startswith(' fără ouă după 22.01)'), a
    return ro.replace(a, a.replace(' după ', f' după{NBSP}'))
def dep_week(h): return (dt.date.fromisoformat(DEP[h]) - dt.date.fromisoformat(s4.c25['start_date'])).days // 7 + 1
assert {h: dep_week(h) for h in HALLS} == {'H1': 45, 'H3': 46, 'H4': 46}
def pnote(k, w):
    r = pw[(k, w)]; out = []
    if k != 'all' and w >= dep_week(k): out.append(f'golire din {num(ddm(DEP[k]))}')
    if k == 'all' and any(w >= dep_week(h) for h in HALLS if (h, w) in pw): out.append('golire, vezi halele')
    if r['partial']: out.append(f'{num(r["days"])} zile')
    return '<span class="ip-muted">' + ' · '.join(out) + '</span>' if out else ''
def pcells(k, w):
    r = pw[(k, w)]
    return (f'<td class="num">{n(r["total_eggs"])}</td><td class="num">{n(r["hatching_eggs"])}</td><td class="num">{pct(r["hatching_pct"])}</td>'
            f'<td class="num">{n(r["avg_egg_weight_g"], 1)} g</td><td class="num ip-muted">{n(r["floor_eggs"])}</td>{pmortc(k, w, "F")}{pmortc(k, w, "M")}<td>{pnote(k, w)}</td>')
def tb_past():
    eg, mo = t4.thico('eggs'), t4.thico('mortality')
    sx = lambda s: f'<th scope="col" class="num">{t4.thsex(s)}</th>'
    head = (f'<thead><tr><th scope="col" rowspan="2">Săptămâna · hala</th><th scope="col" rowspan="2" class="num">{eg}Ouă</th><th scope="colgroup" colspan="2">Incubabile</th>'
            f'<th scope="col" rowspan="2" class="num">Masă ou</th><th scope="col" rowspan="2" class="num">Jos</th><th scope="colgroup" colspan="2">{mo}Mortalitate</th><th scope="col" rowspan="2">Notă</th></tr>'
            f'<tr><th scope="col" class="num">Ouă</th><th scope="col" class="num">%</th>{sx("F")}{sx("M")}</tr></thead>')
    body = ''
    for w in PSHOW:
        a = pw[('all', w)]
        body += (f'<tbody><tr class="is-group"><th scope="rowgroup"><span class="ip-num">SAPT {w}</span> <span class="s4t-sub">{span(a["from"], a["to"])}</span>'
                 f'<span class="ip-sr">, Hîrbovăț</span></th>{pcells("all", w)}</tr>')
        for h in HALLS:
            if (h, w) in pw:
                body += f'<tr><th scope="row" class="ip-table__child">{h}<span class="ip-sr">, SAPT {w}</span></th>{pcells(h, w)}</tr>'
        body += '</tbody>'
    assert ('H1', s4.PFIN) not in pw and ('H4', s4.PFIN) in pw
    t = t4.tbox(f'Producție Hîrbovăț pe săptămâni și hale, Seria 2025, SAPT {PSHOW[-1]}–{s4.PFIN}', head + body)
    wks = sorted(w for (h, w) in pw if h == 'all')
    note = (f'Hîrbovăț · Seria 2025 · {num(f"SAPT {PSHOW[-1]}–{s4.PFIN}")} din {num(f"SAPT {wks[0]}–{wks[-1]}")} · cele mai noi întâi · '
            f'găini, ouat și hrană: – fără date în registru')
    fnote = (f'Seria 2025 nu are efectiv în registru (nicio populare sau transfer), deci găinile și ouatul nu se calculează; grame de hrană nu există pentru acest ciclu. '
             f'Mortalitatea nu include golirea halelor (H1 din {num(ddate(DEP["H1"]))}, H3 din {num(ddate(DEP["H3"]))}, H4 din {num(ddate(DEP["H4"]))}); în zilele de golire registrul trece la morți și păsările scoase, deci o săptămână numai cu zile de golire are {DASH} (morți necunoscuți). '
             f'Jos nu intră în ouă. H1 și H3 fără ouă după {num(ddate(s4.PLAST["H1"]))}.')
    return t4.block('h-psapt', 'Pe săptămâni și hale', note, t + t4.foot(fnote, (f'Vezi toate săptămânile ({len(wks)})', '#productie-saptamani')))

# ================================================================= states
sk = s4.sk
def page_loading():
    # FIN11G fix 1: §54 Stat skeletons with the six loaded Stats' guide-line items (one-line px at 1440), so they wrap at 1280 as loaded
    stat = ''.join(SKL.stat(g) for g in [[[243.36], [182.34]], [[146.7], [229.89]], [[132.89], [178.56]], [[132.89], [152.13]], [[132.89], [197.59]], [[117.8, 64.19], [234.22]]])
    head = th_head(desk=True)
    ncols = 14
    numcols = [True] * ncols; numcols[0] = False; numcols[6] = False  # Săptămâna · hala, Stare
    tr = lambda cls: f'<tr{cls}>' + ''.join(f'<td{" class=\"num\"" if numcols[i] else ""}>{sk("s4s-w70")}</td>' for i in range(ncols)) + '</tr>'
    rows = ''.join('<tbody>' + tr(' class="is-group"') + tr('') * 3 + '</tbody>' for _ in range(2))
    tbl = f'<div class="ip-tablescroll ip-tablescroll--framed"><table class="ip-table ip-table--rowheads">{head}{rows}</table></div>'
    return f'''<div class="s14-page" aria-busy="true">
<p class="ip-sr" role="status">Se încarcă raportul Producție pentru Hîrbovăț, ciclul curent.</p>
<section class="s4-top" aria-hidden="true">{SKL.chart_readout("s4-readout", f"0 0 {s4.CW} {s4.CH}")}<div class="ip-stats s4-stats">{stat}</div></section>
<section class="ip-block" aria-hidden="true"><div class="ip-sechead"><h2 class="ip-t-heading">Pe săptămâni și hale</h2></div>{tbl}</section>
</div>'''

def page_empty():
    arr = sorted({a['date'] for a in t4.TR['arrivals']})
    assert (arr[0], arr[-1]) == ('2026-02-22', '2026-02-28')
    for k in ('daily', 'eggs_daily', 'weighings', 'culls', 'sexing_errors'):
        assert all(r['date'] >= '2026-03-01' for r in S[k]), k
    lines = ['Nicio înregistrare de producție (ouă, mortalitate, hrană) la Hîrbovăț între <span class="ip-num">10.02.2026</span> și <span class="ip-num">28.02.2026</span>.',
             f'Găinile au sosit prin transfer pe <span class="ip-num">{ddm(arr[0])}–{ddate(arr[-1])}</span>; primele înregistrări zilnice sunt din <span class="ip-num">01.03.2026</span> (sapt. 3).']
    step = '<button type="button" class="ip-btn ip-btn--secondary">Alege ciclul curent</button>'
    return f'<div class="s14-page">{s4.empty_box("fără date", lines, step, "Producție, interval 10.02–28.02.2026")}</div>'

def page(mode='default'):
    hmode = {'noguide': 'noguide', 'past': 'past'}.get(mode, 'default')
    if mode == 'loading': body = page_loading()
    elif mode == 'empty': body = page_empty()
    elif mode == 'noguide': body = f'<div class="s14-page">{s4.BANNER_NG}{top("noguide")}{tb_weeks(True)}</div>'
    elif mode == 'past': body = f'<div class="s14-page">{top("past")}{tb_past()}</div>'
    else: body = f'<div class="s14-page">{top()}{tb_weeks()}</div>'
    fb = filterbar('empty' if mode == 'empty' else 'past' if mode == 'past' else 'default')
    return f'<div class="ip-shell">{sidebar("2025" if mode == "past" else "2026")}<main class="ip-shell__main">\n{header(hmode)}\n{fb}\n{body}\n</main></div>'

# ================================================================= CSS: shared rules only (S4 shell + states, S4T tables, S11 farm select)
_t4_base = s4.CSS
assert t4.CSS.startswith(_t4_base)
S11_LINES = [l for l in s11.CSS.split('\n') if l.startswith(('.s11-farm', '.s11-select', '.s11-filterbar:has', '.ip-input > select + .ip-icon'))]
assert len(S11_LINES) == 4, S11_LINES   # FIN10 fix 1: the select chevron rule is additions.css only
CSS = (s4.CSS + '\n' + s4.CSS_STATES + t4.CSS[len(_t4_base):] + '\n' + '\n'.join(l for l in S11_LINES if not l.startswith('.ip-input > select'))
       + '\n.s14-page{display:flex;flex-direction:column;gap:var(--space-6)}')  # FIN15B: .s14-nw (S14P fix 2: a figure and its lamp on one line) is additions.css §64 ip-nowrap  # S14P: the scroll region's position:relative is now additions.css section 36

STEM = 'S14-rapoarte-productie'
BOARDS = [
 (STEM, 'S14 · Rapoarte · Producție', lambda: page()),
 (STEM + '--se-incarca', 'S14 · Rapoarte · Producție · se încarcă', lambda: page('loading')),
 (STEM + '--fara-date-in-perioada', 'S14 · Rapoarte · Producție · fără date în perioadă', lambda: page('empty')),
 (STEM + '--fara-ghid', 'S14 · Rapoarte · Producție · fără ghid', lambda: page('noguide')),
 (STEM + '--seria-2025', 'S14 · Rapoarte · Producție · Seria 2025 (ciclu închis)', lambda: page('past')),
]
def canvas_heights():  # --canvas: the S14 boards' entries (h from the heights file; x/y kept, canvas_rows.py)
    c = CR.load(ROOT); CR.place(c, STEM, [(s, 1440, HEIGHTS.get(s, 1024), t) for s, t, _ in BOARDS])
    CR.save(ROOT, c)
if __name__ == '__main__':
    for stem, title, fn in BOARDS:
        s4.write(stem, title, fn(), '', H=HEIGHTS.get(stem, 1024), css=CSS)
    if CANVAS: canvas_heights()
