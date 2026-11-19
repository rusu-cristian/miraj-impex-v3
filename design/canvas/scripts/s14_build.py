"""S14 · Rapoarte · Producție (desktop 1440, S14N): spec D3 / screens S14 round 2 on real data (realdata.py, DATA03 extracts,
today 18.11.2026 = SAPT 59, D151–D159). Owner view, Hîrbovăț · Seria 2026, period "4 săpt." (SAPT 56–59, D159), grain Zi.
python3 -I s14_build.py <repo_root> [heights.json] [--canvas]   (--canvas also writes the S14 board entries into canvas.json)
Page order (screens S14): filters (Perioadă, Pas Zi / Săpt., Fermă, Exportă) → the PeriodTable (SYS14: every week open with its
days, D166; per week the farm week row → its daily farm rows → the hall week rows H5 · H6 · H7, spec D3 rowGroups; foot
"Total perioadă") → the % ouat Readout and the cycle Stats (D92, after the table).
Columns (spec D3, D167 order): Găini (hens at the start of the day / week, D162) · Ouat (ouă, % ouat, ghid, față de ghid, stare, % Cat. 1–3
of the row's eggs (D183, S4C), jos, % jos; D156/D159 start-of-day basis; guide = Productivitate col I per age week, D184) · Incubabile · Masă ou · Mortalitate (real deaths, D155) · Hrană ♀ g/pasăre ·
Vânzare (the emptying sales, last group, D155/D167). Builders imported from design/system/scripts/sys14_build.py (through
s4t_build → s4s_build): productie (farm and hall), farm_pasari / hall_pasari, hrana, ptable(rowgroups=…), block, never copied.
Seria 2025 (closed): the previous flock's last 4 age weeks SAPT 62–65 on its own calendar (sys14 Cal, as S4 ciclu ales,
D170) + "Arată toate săptămânile"; eggs per hall and deaths only (no flock count, no feed, D52).
Reuses: s11_build (Rapoarte sidebar, Export key), s4t_build (Stat, top, foot), s4s_build (shell, filters, Readouts, past-cycle
helpers, empty box, no-guide banner, write()). Writes design/canvas/project/S14-rapoarte-productie[--<state>].dc.html and
/tmp/s4s/<stem>.html (plain wrappers for s14_render.py)."""
import sys, json, pathlib, importlib.util, datetime as dt

HERE = pathlib.Path(__file__).resolve().parent
def _load(name):
    sp = importlib.util.spec_from_file_location(name, HERE / f'{name}.py')
    m = importlib.util.module_from_spec(sp); sp.loader.exec_module(m); return m
t4 = _load('s4t_build')   # reads sys.argv[1] (repo root)
s11 = _load('s11_build')  # reads sys.argv[2] as its heights only if given (harmless: it writes nothing on import)
s4, P, R = t4.s4, t4.P, t4.R
sys.path.insert(0, str(HERE)); import canvas_rows as CR
ROOT, HP, CANVAS = CR.args('s14_heights.json')  # heights default: s14_heights.json
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}

n, sgn, pct, num, ddate, lamp, icon = R.n, R.sgn, R.pct, t4.num, s4.ddate, s4.lamp, s4.icon
WK, TODAY, HALLS = R.WK, R.TODAY, R.HHALLS
F, M, DASH, NOGD = s4.F, s4.M, t4.DASH, t4.NOGD
NOG = s4.NOG  # 'ghid: – fără ghid'
NBSP = ' '
WEEKS = P.WEEKS                                   # SAPT 59 … 56 (D159: newest first, the last 4 weeks)
assert WEEKS == [59, 58, 57, 56] == [WK - i for i in range(4)]
P_FROM, P_TO = P.P_FROM, P.P_TO                   # 27.10 – 18.11.2026
H7_EMPTY = R.empty_since('H7')                    # 14.11.2026 (D158)
assert H7_EMPTY == dt.date(2026, 11, 14) and [h for h in HALLS if R.empty_since(h)] == ['H7']

# ================================================================= data · the cycle (Stats; every figure from the extracts)
WMIN = R.C26['sapt_start']                        # SAPT 20 (22.02.2026, the first arrival)
CWEEKS = [k for k in range(WMIN, WK + 1) if R.week(k)]
W = R.week
eggs_cyc = sum(W(k)['eggs'] or 0 for k in CWEEKS); hatch_cyc = sum(W(k)['inc'] or 0 for k in CWEEKS)
assert eggs_cyc == sum(r['eggs'] or 0 for r in R.FARM_DAY) == 5073155
FIRST = next(r for r in R.FARM_DAY if r['eggs'])  # first eggs at Hîrbovăț: 10.03.2026, SAPT 23
assert (FIRST['date'], FIRST['sapt']) == ('2026-03-10', 23)
_ew = [r for r in R.FARM_DAY if r['eggs'] and r['egg_g']]   # the first 5 laying days (10–14.03.2026) have no egg weight
EGG_NOW = eggs_cyc - sum(r['eggs'] for r in _ew); EGG_D0 = _ew[0]['date']
assert EGG_NOW == 5119 and EGG_D0 == '2026-03-15'
EGG_W = sum(r['egg_g'] * r['eggs'] for r in _ew) / (eggs_cyc - EGG_NOW)   # egg-weighted mean over the farm days (D46: medie pe ouă)
LAY = {k: W(k)['eggs'] / W(k)['hen_days'] * 100 for k in CWEEKS if W(k)['eggs']}
PEAK = max(LAY, key=LAY.get)
assert (PEAK, R.rnd(LAY[PEAK], 1)) == (30, 90.3)  # key-figures: peak SAPT 30, 90,3 %
_fw = [k for k in CWEEKS if W(k)['g_F']]
_bd = {k: W(k)['kg_F'] * 1000 / W(k)['g_F'] for k in _fw}   # ♀ fed bird-days (kg ÷ g, D38)
FEED_F = sum(W(k)['kg_F'] for k in _fw) * 1000 / sum(_bd.values())
FEED_GF = sum(W(k)['guide']['feed_g_F'] * _bd[k] for k in _fw) / sum(_bd.values())   # guide weighted by bird-days (D162)
FEED_DEV = (FEED_F / FEED_GF - 1) * 100
cyc = s4.cum
MORT_G = R.guide_mort(WK)

# egg weight per week / day / period from the extract (D176/DATA05): hall and farm weekly egg_g = Σ(eggs × g) ÷ Σ eggs over
# the days with eggs (empty days excluded; the sheet's Masa Total row, which averaged in H7's empty days, is egg_g_sheet);
# the period = the weeks' egg_g weighted by their eggs (as sys14 productie, so S4 Producție and S14 agree)
def egg_w(h=None):
    def f(kind, key):
        if kind == 'w': r = W(key, h); g = r and r['egg_g']
        elif kind == 'd': r = R.day(key, h); g = r and r['eggs'] and r['egg_g']
        else:
            ws = [W(k, h) for k in WEEKS]; ws = [r for r in ws if r and r['eggs'] and r['egg_g']]
            g = sum(r['egg_g'] * r['eggs'] for r in ws) / sum(r['eggs'] for r in ws) if ws else None
        return n(g, 1) if g else (DASH, f'goală din {R.ddm(H7_EMPTY)}')
    return f
assert egg_w('H7')('w', 58) == '66,3' and egg_w()('w', 58) == '66,6' and W(58)['egg_g_sheet'] == 57.16 and W(58, 'H7')['egg_g_sheet'] == 37.88

# ================================================================= the PeriodTable groups (spec D3, D167 order)
def short(vis, full): return f'<span aria-hidden="true">{vis}</span><span class="ip-sr">{full}</span>'   # short head, full words hidden (as Rezumat)
def groups(h=None, guide=True, tag=None):
    """tag None: every group (desktop). tag = one group's word (phone, D176: single-choice group tags, one group fits 358):
    Găini stays (as În viață on the Fermă phone), then that group; Ouat on the phone = ouă · % ouat with the ghid as the
    cell's second small line (sys14 stacked)."""
    pr = P.productie(guide, h)
    pa = P.farm_pasari() if h is None else P.hall_pasari(h)
    gh = P.group(P.hrana(h), 'Hrană · g/pasăre')
    feed = {'name': gh['name'], 'leaves': [lv for lv in gh['leaves'] if 'is-hen' in lv[1]][:2 if guide else 1]}   # ♀ efectiv · ghid
    egg = P.group(pr, 'Masă ou'); ou = P.group(pr, 'Ouat'); jos = P.group(pr, 'Jos')
    jl = [(short('jos', 'ouă jos'), '', jos['leaves'][0][2]), (short('% jos', '% ouă jos'), '', jos['leaves'][1][2])]
    out = [P.group(pr, 'Găini'), {'name': ou['name'], 'leaves': ou['leaves'] + jl}, P.group(pr, 'Incubabile'),   # D176: Jos (count + %) in Ouat
           {**egg, 'leaves': [(egg['leaves'][0][0], '', egg_w(h))]},
           P.group(pa, 'Mortalitate'), feed, P.group(pa, 'Vânzare')]   # Vânzare last (D167)
    if tag is None: return out
    if tag == 'Ouat':
        lv = {lab: (lab, c, fn) for lab, c, fn in ou['leaves']}
        st = P.stacked({'name': ou['name'], 'leaves': [lv['% ouat']]}, {'leaves': [lv['ghid']]}, 'ghid') if guide else {'name': ou['name'], 'leaves': [lv['% ouat']]}
        return [out[0], {'name': st['name'], 'leaves': [lv['ouă']] + st['leaves']}]
    if tag == TAG_INC:   # D186/D188: Cat. 1–3 are egg characteristics like Incubabile, so this tag shows both; Găini stays and the
        # table scrolls sideways under the sticky first column (395 px in the 356 region, §36: the ~30–40 px D188 measured); % incub.
        # rides under the count as the cell's second line (the phone's companion rate, sys14 stacked), Cat. heads wrap on the phone
        inc = out[2]; cats = [(lab.replace('\u00a0', ' '), c, fn) for lab, c, fn in ou['leaves'] if lab in P.CAT_HEAD.values()]
        st = P.stacked({'name': inc['name'].replace('Incubabile', TAG_INC), 'leaves': inc['leaves'][:1]}, {'leaves': inc['leaves'][1:]}, '% incub.')
        return [out[0], {'name': st['name'], 'leaves': st['leaves'] + cats}]
    return [out[0], P.group(out[2:], tag)]
TAG_INC = 'Incubabile · Cat.'   # D188 (Claude, to review): the phone tag that shows Incubabile and % Cat. 1–3
GOALA_H = f'goală din {R.dd(H7_EMPTY)} · {", ".join(R.HALL_SECS["H7"])}'
def hall_rows(guide=True, tag=None):
    return [{'label': h, 'sr': '', 'sections': [('', '', groups(h, guide, tag))],
             'empty': (lambda k, h=h: GOALA_H if h == 'H7' and not W(k, 'H7')['alive_F_start'] else None)} for h in HALLS]
SER = P.SER                                        # 'Hîrbovăț · Seria 2026 · SAPT 56–59'
PHONE_TAG = 'Ouat'   # D176: the phone's default group (the report's own: eggs)
def ptable(guide=True, phone=False, skel=None):
    tag = PHONE_TAG if phone else None
    cap = f'Producție · {SER} · pe zile și hale' + ('' if guide else ' · fără ghid') + (f' · {tag}' if tag else '')
    t, tid = P.ptable('pr', cap, [('Total fermă', '', groups(None, guide, tag))], spine='phone' if phone else 'farm',
                      skel=skel or False, rowgroups=hall_rows(guide, tag), colw=COLW if skel else None)
    if not skel and not phone: assert t.count(s4.PART[0]) == 1; t = t.replace(s4.PART[0], s4.PART[1])   # "17–18.11 · 2 din 7 zile" (D165)
    if skel:   # the Data column at its loaded px too (the loaded partial week's label "SAPT 59 · 17–18.11 · 2 din 7 zile")
        a = '<th scope="col" class="ip-ptable__data" rowspan="2">'; assert t.count(a) == 1
        t = t.replace(a, a[:-1] + f' style="--ip-colw:{DATA_W}px">')
    assert t.count('aria-expanded="true"') == (0 if skel else 4)   # D166: every week open
    return cap, t
NOTE_LAY = ('% ouat: zi = ouă ÷ găini la început de zi; săptămână și perioadă = ouă ÷ zile-găină. Găini = la început de zi (săptămâna: la începutul ei). '
            f'Stare: {num("±3 pct")} față de ghidul Ross 308 ouat al săptămânii de vârstă.')
NOTE_CAT = ' % Cat. 1–3 = ouăle categoriei ÷ ouă.'   # D183 (S4C): desktop only (the phone's Ouat tag has no category columns)
NOTE_ROWS = (f'Sub fiecare săptămână: zilele fermei, apoi halele pe săptămână (ouăle se numără pe hală; zilele pe hale sunt în Fermă › Producție). '
             f'Mortalitate = morți reali; golirea din {num(R.ddm(R.EMPTYING_SINCE))} e Vânzare. Hrană {F}: zi = g din registru, săptămână = kg ÷ zile-găină hrănite. Masă ou = medie pe ouă.')
# D163 (as S11 Consum and the Fermă Hrană tab): the register's ♀ kg cells of 460 kg/compartiment keep SAPT 57's week figure
# (kg ÷ hen-days) under its days' register grams
G57 = sorted(R.feed_g(d, 'F') for d in P.days_of(57))
assert (n(W(57)['g_F'], 1), n(G57[0]), n(G57[-1])) == ('130,5', '143', '161')
NOTE_ODD = (f'Hrană {F} {num(R.SAPT(57))}: {num(n(W(57)["g_F"], 1) + " g")} pe săptămână (kg ÷ zile-găină) față de {num(n(G57[0]) + "–" + n(G57[-1]) + " g")} în zile (g din registru): '
            + t4.odd_note())
def table(guide=True, phone=False):
    cap, t = ptable(guide, phone)
    notes = [(NOTE_LAY if guide else NOTE_LAY.split(' Stare:')[0] + ' Ciclul nu are ghid: coloanele ghid, față de ghid și stare lipsesc.') + NOTE_CAT, NOTE_ROWS, NOTE_ODD]
    if phone: notes = [notes[0][:-len(NOTE_CAT)], NOTE_ROWS.split(' Mortalitate')[0], f'Pe telefon tabelul arată o grupă odată ({PHONE_TAG}; ghidul pe al doilea rând al celulei); alege alta din Filtrează. Găini rămâne.']
    return (f'<section aria-labelledby="pr-h"><h2 class="ip-sr" id="pr-h">Producție pe săptămâni, zile și hale</h2>'
            + P.block('pr', cap, t, bar=False, touch=phone, after=s4.ptnote(*notes)) + '</section>')
# loading: bars per leaf (ch of the usual figure): Găini 4 · ouă 4 · % 4 · ghid 4 · față 4 · stare 4 · % Cat. 1–3 4 · 4 · 4 (S4C, D183: "6,2%") · jos 3 · % 4 · inc 4 · % 4 · g 3 · ♀ 2 · ♂ 2 · g 4 · ghid 3 · ♀ 4 · ♂ 3
SKEL = [4, 4, 4, 4, 4, 4, 4, 4, 4, 3, 4, 4, 4, 3, 2, 2, 4, 3, 4, 3]
# §54 on the loading table's heads (Vârstă, SAPT, then the 20 leaves): the loaded columns' px (S14N, measured on the
# wide board). The loaded table is wider than the page (D167), so shares of the 1158 region cannot hold it: the heads declare the
# loaded px and the loading board's body cells drop their 1% (.s14-skel, screen CSS; Chrome's auto layout lets a cell's %
# beat a head's px), so every column sits where the loaded one does and the board is as wide as the loaded one
DATA_W = 259.17
COLW = [f'{x}px' for x in (110.11, 63.14, 78.36, 78.8, 72.23, 63.14, 106.13, 88.19, 78.06, 78.06, 78.06, 63.14, 63.05, 78.8, 63.14, 113.3, 67.41, 59.59, 89.14, 68.72, 70.97, 63.14)]   # as drawn on the wide board (1954; S4C: + % Cat. 1–3, 78.06 each)

# ================================================================= shell (Rapoarte current, as S11)
def sidebar(series='2026'):
    sb = s4.sidebar(series=series)
    sb = sb.replace('class="ip-nav is-current" href="#serie" aria-current="page"', 'class="ip-nav" href="#serie"')
    sb = sb.replace('class="ip-nav" href="#rapoarte"', 'class="ip-nav is-current" href="#rapoarte" aria-current="page"')
    assert sb.count('is-current') == 1
    return sb
assert sidebar() == s11.sidebar('Proprietar')

c25 = s4.c25
def header(mode='default'):
    crumbs = '<nav class="ip-crumbs" aria-label="Cale"><a href="#rapoarte">Rapoarte</a><span aria-hidden="true">›</span><span aria-current="page">Producție</span></nav>'
    acts = ''
    if mode == 'past':
        meta = (f'{lamp("off", "închis")}<span class="ip-note">Seria 2025 · Hîrbovăț · început dedus {num(ddate(c25["start_date"]))} · încheiat prin golire {num(ddate(s4.PREV_END))} · '
                f'{num(s4.P25)} · date din {num(ddate(c25["data_start_date"]))}</span>')
        acts = '<div class="ip-row"><a class="ip-btn ip-btn--secondary" href="#ciclu-seria-2026">Înapoi la ciclul curent</a></div>'
    else:
        gtag = '<span class="ip-tag ip-tag--neutral">fără ghid</span>' if mode == 'noguide' else '<span class="ip-tag ip-tag--neutral">Ross 308 ouat</span>'
        meta = (f'<span class="ip-tag ip-tag--neutral">Ouat</span>{gtag}<span class="ip-note">Seria 2026 · Hîrbovăț {num(R.SAPT(WK))}, ciclu din {num(ddate(R.CYCLE_START))} · '
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
def farm_select():  # laying farms only; Telița (growth) is listed but disabled with its reason (FIN12 / D130: section 53 Select)
    O = '<span class="ip-select__txt"><span class="ip-select__top">'
    opts = (f'<li class="ip-menu__item ip-select__opt" role="option" id="f-farm-o1" aria-selected="true"><span class="ip-select__chk" aria-hidden="true">{icon("check", 16)}</span>{O}Hîrbovăț</span></span></li>'
            f'<li class="ip-menu__item ip-select__opt" role="option" id="f-farm-o2" aria-selected="false" aria-disabled="true"><span class="ip-select__chk" aria-hidden="true"></span>{O}Telița <span class="ip-note">{TEL_WHY}</span></span></span></li>')
    return ('<div class="s11-farm"><span class="ip-label" id="f-farm-l">Fermă</span><div class="ip-select s11-select">'
            '<div class="ip-input ip-select__btn" id="f-farm" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="false" aria-controls="f-farm-list" aria-labelledby="f-farm-l">'
            f'<span class="ip-select__val">Hîrbovăț</span>{icon("chev", 16)}</div>'
            f'<ul class="ip-menu ip-select__menu" role="listbox" id="f-farm-list" aria-labelledby="f-farm-l" tabindex="-1" hidden="hidden">{opts}</ul></div></div>')
# the "fără date" interval: before the first bird reached Hîrbovăț (22.02.2026); the Seria 2026 sheets start 17.02 (deliveries only)
IV = (dt.date(2026, 2, 1), R.CYCLE_START - dt.timedelta(days=1))
assert IV[1] == dt.date(2026, 2, 21) and not [r for r in R.SECTION_DAY if R.D(r['date']) <= IV[1] and (r['alive'] or r['mort'] or r['kg'])]
assert not [r for r in R.HALL_DAY if R.D(r['date']) <= IV[1] and r['eggs']]
PAS = lambda: s4.seg('Pas', ['Zi', 'Săpt.'], 0)   # D166: grain Zi (every week open)
def filterbar(mode='default'):
    if mode == 'empty':
        itv = f'<span class="ip-interval" role="group" aria-label="Interval">{s4.datef("iv-a", "de la", R.dd(IV[0]))}{s4.datef("iv-b", "până la", R.dd(IV[1]))}</span>'
        return f'<div class="ip-filterbar s11-filterbar" role="group" aria-label="Filtre">{s4.seg("Perioadă", s4.PERIODS, 3)}{itv}{PAS()}{farm_select()}{s11.export(True)}</div>'
    per = s4.seg('Perioadă', ['Tot ciclul'] + s4.PERIODS[1:], 0) if mode == 'past' else s4.seg('Perioadă', s4.PERIODS, 1)   # D159: 4 săpt.; past as S4 ciclu ales
    return f'<div class="ip-filterbar s11-filterbar" role="group" aria-label="Filtre">{per}{PAS()}{farm_select()}{s11.export()}</div>'

# ================================================================= after the table: Readout (S4 lay chart) + cycle Stats
def stats_cur(ng=False):
    st, sx = t4.stat, t4.stats
    pk = W(PEAK); a, b = R.week_dates(PEAK)
    g_pk = NOG if ng else f'ghid {num(pct(pk["guide"]["lay_pct"]))} · {num(sgn(R.rnd(LAY[PEAK], 1) - R.rnd(pk["guide"]["lay_pct"], 1)) + " pct")}'
    fd = abs(FEED_DEV) <= 10
    g_feed = NOG if ng else f'ghid {num(n(FEED_GF, 1) + " g")} · {num(sgn(FEED_DEV, 1) + "%")} {lamp("ok" if fd else "warn", "în normă" if fd else "abatere")}'
    mok = cyc['mortality_pct']['F'] <= MORT_G
    g_mort = NOG if ng else f'ghid {s4.hen(pct(MORT_G, 2))} {lamp("ok" if mok else "warn", "în normă" if mok else "abatere")}'
    return sx(
        st('eggs', 'Ouă în ciclu', n(eggs_cyc), f'{num(R.drange(FIRST["date"], TODAY))} · Cat.{NBSP}1–⁠3{NBSP}+{NBSP}incubabile', f'primele ouă {num(ddate(FIRST["date"]))} · {num(R.SAPT(FIRST["sapt"]))}'),
        st('eggs', 'Vârf de ouat', f'{n(LAY[PEAK], 1)}<small>%</small>', g_pk, f'{num(R.SAPT(PEAK) + " · " + R.drange(a, b))}'),
        st('eggs', 'Incubabile în ciclu', f'{n(hatch_cyc / eggs_cyc * 100, 1)}<small>%</small>', NOG if ng else NOGD, f'{num(n(hatch_cyc))} din {num(n(eggs_cyc))} ouă'),
        st('eggs', 'Masă medie ou în ciclu', f'{n(EGG_W, 1)}<small>g</small>', NOG if ng else NOGD, f'medie pe ouă · din {num(R.ddm(EGG_D0))}', f'fără {num(n(EGG_NOW))} ouă fără masă'),
        st('mortality', 'Mortalitate cumulată', f'{s4.pair(n(cyc["mortality_pct"]["F"], 2), n(cyc["mortality_pct"]["M"], 2))}<small>%</small>', g_mort,
           *(() if ng else ('(ghid din intrarea în ouat)',)), f'din {num(ddate(R.CYCLE_START))} {s4.pair(n(-cyc["dead"]["F"]), n(-cyc["dead"]["M"]))}'),
        st('feed', f'Hrană {F} în ciclu', f'{n(FEED_F, 1)}<small>g/pasăre/zi</small>', g_feed, 'kg ÷ zile-găină hrănite'))

def top(mode='default'):
    if mode == 'noguide': return t4.top('h-key', 'Producție, ciclul curent', s4.readout(True), stats_cur(True))
    if mode == 'past': return t4.top('h-key', 'Producție, Seria 2025', readout_past(), stats_past())
    return t4.top('h-key', 'Producție față de ghid, ciclul curent', s4.readout(), stats_cur())

# ================================================================= Seria 2025 (real previous-flock data, D51/D52/D54; S4NG calendar)
PW4, PCAL, PLAST, PFIN, PWEEKS = s4.PW4, s4.PCAL, s4.PLAST, s4.PFIN, s4.PWEEKS
DEP = R.PREV_META['depopulation_start']             # H5 15.01 · H6 16.01 · H7 20.01.2026
assert PW4 == [65, 64, 63, 62] and DEP == {'H5': '2026-01-15', 'H6': '2026-01-16', 'H7': '2026-01-20'}
assert [PCAL.label(k)[:2] for k in (65, 62)] == [(dt.date(2026, 1, 23), dt.date(2026, 1, 28)), (dt.date(2026, 1, 2), dt.date(2026, 1, 8))]
def p_egg(h=None):
    def f(kind, key):
        rs = [r for r in s4._prows(kind, key, h) if r['total_eggs'] and r['avg_egg_weight_g']]
        return n(sum(r['avg_egg_weight_g'] * r['total_eggs'] for r in rs) / sum(r['total_eggs'] for r in rs), 1)
    return f
def p_cat(c, h=None):   # D187 (S4C2): Cat. c as % of the eggs, from the previous flock's daily form counts (D18: total = Cat. 1–3 + incubabile)
    def f(kind, key):
        rs = s4._prows(kind, key, h); e = sum(r['total_eggs'] for r in rs)
        return R.pct(sum(r[f'cat{c}'] for r in rs) / e * 100) if e else (DASH, 'fără ouă')
    return f
for _k in PW4:   # built-in check: the farm and hall week cells = the weekly extract's Cat. counts ÷ its eggs (as the current cycle's _cat_check)
    for _h in (None, *HALLS):
        _w = s4.pall[_k] if _h is None else s4.phall.get((_h, _k))
        if _w: assert all(p_cat(c, _h)('w', _k) == R.pct(_w[f'cat{c}'] / _w['total_eggs'] * 100) for c in (1, 2, 3)), (_k, _h)
def past_groups(h=None):
    E = P.ico('eggs')
    return [{'name': f'{E}Ouă', 'leaves': [('ouă', '', s4.p_sum('total_eggs', h))] + [(P.CAT_HEAD[c], '', p_cat(c, h)) for c in (1, 2, 3)]},   # D187: % Cat. 1–3 beside Incubabile (D186)
            {'name': f'{E}Incubabile', 'leaves': [('ouă', '', s4.p_sum('hatching_eggs', h)), ('%', '', s4.p_share('hatching_eggs', h))]},
            {'name': f'{E}Jos', 'leaves': [('ouă', '', s4.p_sum('floor_eggs', h)), ('%', '', s4.p_share('floor_eggs', h))]},
            {'name': f'{E}Masă ou', 'single': True, 'sr': ', grame', 'leaves': [('g', '', p_egg(h))]},
            {'name': f'{P.ico("mortality")}Mortalitate', 'leaves': [(P.sx('F'), 'is-hen', s4.p_mort('F', h)), (P.sx('M'), 'is-rooster', s4.p_mort('M', h))]}]
def past_rows():
    return [{'label': h, 'sr': '', 'sections': [('', '', past_groups(h))],
             'empty': (lambda k, h=h: None if (h, k) in s4.phall else f'fără ouă după {R.ddm(PLAST[h])} · golire din {R.ddm(DEP[h])}')} for h in HALLS]
assert sorted(h for h in HALLS if (h, 65) not in s4.phall) == ['H5', 'H6']
def tb_past():
    cap = f'Producție · Hîrbovăț · Seria 2025 · SAPT {PW4[-1]}–{PW4[0]}, ultimele 4 din SAPT {PWEEKS[0]}–{PFIN} · pe zile și hale'
    t, tid = P.ptable('pr', cap, [('Total fermă', '', past_groups())], weeks=PW4, cal=PCAL, rowgroups=past_rows())
    assert t.count('aria-expanded="true"') == 4
    after = (f'<div class="ip-ptable__after"><button type="button" class="ip-btn ip-btn--secondary" aria-controls="{tid}">Arată toate săptămânile</button>'
             f'<p class="ip-note">apoi SAPT {PFIN} → {PWEEKS[0]} (registrul are date din {ddate(c25["data_start_date"])}).</p></div>'
             + s4.ptnote('Seria 2025 nu are efectiv în registru (nicio populare sau transfer): fără Găini și % ouat; grame de hrană nu există pentru acest ciclu.' + NOTE_CAT,
                         f'Mortalitatea nu include golirea halelor (H5 din {num(R.ddm(DEP["H5"]))}, H6 din {num(R.ddm(DEP["H6"]))}, H7 din {num(R.ddm(DEP["H7"]))}): '
                         f'în zilele de golire registrul trece la morți și păsările scoase, deci o zi sau o săptămână numai cu golire are {DASH}. Jos nu intră în ouă.'))
    return (f'<section aria-labelledby="pr-h"><h2 class="ip-sr" id="pr-h">Producție pe săptămâni, zile și hale, Seria 2025</h2>'
            + P.block('pr', cap, t, bar=False, after=after) + '</section>')
# the Seria 2025 Stats: S4's past Stats with S14's egg-weight Stat in the peak's place (the Readout already draws the peak)
_pw = [r for r in R.PREV_DAILY if r['avg_egg_weight_g'] is not None and r['total_eggs']]
PEGG_NOW = sum(r['total_eggs'] for r in R.PREV_DAILY if r['avg_egg_weight_g'] is None)
PEGG_W = sum(r['avg_egg_weight_g'] * r['total_eggs'] for r in _pw) / sum(r['total_eggs'] for r in _pw)
PEGG_WK = sorted({r['age_week'] for r in _pw})
assert PEGG_NOW == 9314 and sum(r['total_eggs'] for r in _pw) + PEGG_NOW == R.PREV_META['totals']['total_eggs'] == 4606935
assert (PEGG_WK[0], PEGG_WK[-1]) == (23, PFIN) and n(PEGG_W, 1) == '62,0'
OLD_PEAK = f'<span class="ip-stat__val">{n(22739)}</span>'  # FIN16: S4's peak-week Stat
def stats_past():
    sp = s4.stats_past()
    old = [d for d in sp.split('\n') if 'Vârf · ouă pe zi' in d]
    assert len(old) == 1 and OLD_PEAK in old[0], old
    new = t4.stat('eggs', 'Masă medie ou în ciclu', f'{n(PEGG_W, 1)}<small>g</small>', NOGD,
                  f'medie pe ouă · {num(f"SAPT {PEGG_WK[0]}–{PEGG_WK[-1]}")}', f'fără {num(n(PEGG_NOW))} ouă fără masă')
    out = sp.replace(old[0], new); assert out.count('Masă medie ou') == 1 and 'Vârf' not in out
    return out
def readout_past():  # S14P fix 2: "pe 22.01)" stays together on the Readout's sub line
    ro = s4.readout_past(); a = f' până pe {s4.ddate(PLAST["H5"])[:5]})</div>'
    assert ro.count(a) == 1 and a.startswith(' până pe 22.01)'), a
    return ro.replace(a, a.replace(' pe ', f' pe{NBSP}'))

# ================================================================= states
def page_loading():
    import skel as SKL
    stat = ''.join(SKL.stat(g) for g in [[[243.36], [182.34]], [[146.7], [229.89]], [[132.89], [178.56]], [[132.89], [152.13]], [[132.89], [197.59]], [[117.8, 64.19], [234.22]]])
    cap, t = ptable(skel=SKEL)
    return f'''<div class="s14-page s14-skel" aria-busy="true">
<p class="ip-sr" role="status">Se încarcă raportul Producție pentru Hîrbovăț, {R.SAPT(WEEKS[-1])}–{WK}.</p>
<section aria-hidden="true"><p class="ip-cap ip-ptable__cap">{cap}</p>{P.region(t, cap, hidden=True)}</section>
<section class="s4-top" aria-hidden="true">{SKL.chart_readout("s4-readout", f"0 0 {s4.CW} {s4.CH}")}<div class="ip-stats s4-stats">{stat}</div></section>
</div>'''

def page_empty():
    arr = sorted({a['date'] for a in R.TR['arrivals']})
    assert (arr[0], arr[-1]) == ('2026-02-22', '2026-02-28')
    lines = [f'Nicio înregistrare de producție (ouă, mortalitate, hrană) la Hîrbovăț între {num(R.dd(IV[0]))} și {num(R.dd(IV[1]))}.',
             f'Păsările Seriei 2026 au sosit prin transfer pe {num(R.drange(arr[0], arr[-1]))}; primele ouă sunt din {num(ddate(FIRST["date"]))} ({num(R.sapt_txt(FIRST["sapt"]))}).']
    step = '<button type="button" class="ip-btn ip-btn--secondary">Alege ultimele 4 săptămâni</button>'
    return f'<div class="s14-page">{s4.empty_box("fără date", lines, step, "Producție, interval " + R.drange(IV[0], IV[1]))}</div>'

def page(mode='default'):
    hmode = {'noguide': 'noguide', 'past': 'past'}.get(mode, 'default')
    if mode == 'loading': body = page_loading()
    elif mode == 'empty': body = page_empty()
    elif mode == 'noguide': body = f'<div class="s14-page">{s4.BANNER_NG}{table(False)}{top("noguide")}</div>'
    elif mode == 'past': body = f'<div class="s14-page">{tb_past()}{top("past")}</div>'
    else: body = f'<div class="s14-page">{table()}{top()}</div>'
    fb = filterbar('empty' if mode == 'empty' else 'past' if mode == 'past' else 'default')
    return f'<div class="ip-shell">{sidebar("2025" if mode == "past" else "2026")}<main class="ip-shell__main">\n{header(hmode)}\n{fb}\n{body}\n</main></div>'

# ================================================================= CSS: shared rules only (S4 shell + states, S4T tables, S11 farm select)
_t4_base = s4.CSS
assert t4.CSS.startswith(_t4_base)
S11_LINES = [l for l in s11.CSS.split('\n') if l.startswith(('.s11-farm', '.s11-select', '.s11-filterbar:has', '.ip-input > select + .ip-icon'))]
assert len(S11_LINES) == 4, S11_LINES   # FIN10 fix 1: the select chevron rule is additions.css only
CSS = (s4.CSS + '\n' + s4.CSS_STATES + t4.CSS[len(_t4_base):] + '\n' + '\n'.join(l for l in S11_LINES if not l.startswith('.ip-input > select'))
       + '\n.s14-page{display:flex;flex-direction:column;gap:var(--space-6)}'
       + '\n.s14-skel .ip-ptable{width:max-content}.s14-skel .ip-ptable.ip-ptable :is(thead th:not([style*="--ip-colw"]),:is(tbody,tfoot) > tr > td){width:auto} /* S14N: the loading table keeps the loaded px its heads declare (section 54) */')

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
