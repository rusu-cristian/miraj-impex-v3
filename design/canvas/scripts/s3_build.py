"""Build S3 · Seria 2026 (series scope page) on real data, table first (S3N: spec D1 › Series page, D154, D160, D166, D167,
D171), today 18.11.2026 = SAPT 59: desktop 1440 (owner), se încarcă, seria închisă (Seria 2025), fără acces (Observator,
Hîrbovăț only), telefon 390.
python3 -I s3_build.py <repo_root> [heights.json] [--canvas]
Order (D174): PageHeader (Exportă) → CycleTrack full (both farms, transfer link, azi + SAPT readout) → Tabs Păsări · Hrană ·
Greutate · Producție; the Păsări tab = its PeriodTable (no FilterBar, D63: every column group of the period; no Detaliu; the
last 4 weeks SAPT 59–56 all open with their days, D166; spine Data · Vârstă (zile) · SAPT · Fermă; "Total serie": În viață
first, Vânzare last, D167; the table foot "Încarcă încă 4 săptămâni" + count, SYS18/D190, then the notes), then its Readouts (Mortalitate cumulată, În viață
over SAPT 1–59), then "Ferme" (FarmSummary ×2). Each bird counted once in the transfer weeks (Telița until its batch arrives).
Boards: default, 8 săptămâni (after one press, SYS18), începutul seriei (every week loaded to SAPT 1, the region scrolled to the
transfer weeks SAPT 22–19; replaces "toate săptămânile"), se încarcă, seria închisă (content-width table), fără acces, telefon
(single-choice tags, D174). The period is the whole series (SAPT 1–59, D190): its record types set the groups (D45).
Reuses by import: s4t_build (→ s4s_build shell, sidebar, icons, formats, writer, Seria 2025 figures; Păsări groups, Readout
+ GuideChart; sys14_build PeriodTable builders as `P`), s2_build for the phone shell (phone, write_phone, CSS_P; importing
it rewrites the S2/S7 default boards byte-identical, as s4ns). Figures only from realdata (DATA03 extracts, D151–D159); the
cull of 11.11 is the illustrative record of illustrative-events-hirbovat.json (D152).
Writes design/canvas/project/S3-seria-2026[--<state>].dc.html, /tmp/s4s/<stem>.html (desktops), /tmp/s2/<stem>.html (phone);
heights from s3_heights.json (s3_render.py --heights); --canvas refreshes the S3 row (entries keep x/y, the phone board
goes after the desktop ones)."""
import sys, re, json, pathlib, importlib, importlib.util, datetime as dt

HERE = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(HERE)); import canvas_rows as CR
ROOT, HP, CANVAS = CR.args('s3_heights.json')  # heights default: s3_heights.json
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}
sys.argv = ['s4t_build.py', str(ROOT)]
_sp = importlib.util.spec_from_file_location('s4t_build', HERE / 's4t_build.py')
T = importlib.util.module_from_spec(_sp); _sp.loader.exec_module(T)   # writes nothing on import
s4, P, R = T.s4, T.P, T.R
n, pct, sgn, rnd, ddate, icon, lamp, pair, sg, F, M = R.n, R.pct, R.sgn, R.rnd, s4.ddate, s4.icon, s4.lamp, s4.pair, s4.sg, s4.F, s4.M
DASH = '–'
def num(v): return f'<span class="ip-num">{v}</span>'

# ---------- figures (realdata; D151 age weeks, D155 deaths vs sales, D156 lay, D157 weighing vs the week just ended) ----------
cT, cH, cP = R.CTEL, R.C26, R.C25
TODAY, WK = R.TODAY_ISO, R.WK
assert (cT['start_date'], cT['end_date'], cT['sapt_start'], cT['sapt_end']) == ('2025-10-07', '2026-02-28', 1, 21)
assert (cH['start_date'], cH['sapt_start'], cH['status'], WK) == ('2026-02-22', 20, 'open', 59)
TEL_SPAN, HIR_SPAN = f'SAPT {cT["sapt_start"]}–{cT["sapt_end"]}', f'SAPT {cH["sapt_start"]}–{WK}'
P25 = s4.P25                                                    # Seria 2025: SAPT 19–65 (D151)
PLC = R.PLACEMENT_ROWS['totals']; assert PLC == {'F': 28080, 'M': 3510}
PL_DATE = R.PLACEMENT_ROWS['meta']['date']; assert PL_DATE == cT['start_date']
ARR = R.ARRIVED; assert ARR == {'F': 26883, 'M': 3024}           # D28: transfers counted at arrival
XD = sorted({a['date'] for a in R.TR['arrivals']} | {o['date'] for o in R.TR['transfers_out']})
X0D, X1D = XD[0], XD[-1]; assert (X0D, X1D) == ('2026-02-22', '2026-02-28')
XSPAN = f'{R.ddm(X0D)}–{ddate(X1D)}'
ALIVE = {sx: R.alive_farm(sx) for sx in 'FM'}; ALIVE['all'] = ALIVE['F'] + ALIVE['M']
SER = R._CL['series'][0]
assert (SER['F'], SER['M'], SER['by_farm']['telita']) == (ALIVE['F'], ALIVE['M'], 0) == (12561, 794, 0)
# Telița (closed, real): deaths, sexing errors, the register difference at its close (D25/D63), weight at transfer, feed
MT = R.MORT_TELITA
TM = {'F': MT['meta']['totals']['mortality_F'], 'M': MT['meta']['totals']['mortality_M']}; assert TM == {'F': 704, 'M': 239}
assert TM == {sx: sum(r['mortality'] for r in MT['rows'] if r['sex'] == sx) for sx in 'FM'}
TSE = {'F': 0, 'M': 0}
for r in MT['rows']:
    if r.get('sexing_error'): TSE[r['sex']] += -r['sexing_error']   # signed as typed (−1 = one bird removed)
assert TSE == {'F': 0, 'M': 1}, TSE
TMP = {sx: TM[sx] / PLC[sx] * 100 for sx in 'FM'}
TDIFF = {sx: -(PLC[sx] - TM[sx] - TSE[sx] - ARR[sx]) for sx in 'FM'}   # D63: a register difference, never a Corecție
outF = {o['from_section']: o['qty'] for o in R.TR['transfers_out'] if o['sex'] == 'F'}
WSAPT = cT['sapt_end']                                                 # D157: 24.02.2026 = age 140 = SAPT 21, vs guide week 20
wl = {r['section']: r for r in R.WEIGHINGS_TELITA if r['sapt'] == WSAPT and r['sex'] == 'F' and r['section'] in outF}
assert set(wl) == set(outF) and {r['date'] for r in wl.values()} == {'2026-02-24'}
WDATE = '2026-02-24'
WT = rnd(sum(wl[s]['avg_weight_g'] * outF[s] for s in outF) / sum(outF.values()))
WG = R.weigh_guide(WSAPT, 'F'); WDEV = (WT - WG) / WG * 100
assert (WT, WG, rnd(WDEV, 1)) == (2531, 2370, 6.8)
FDAYS = R.FEED_TELITA
FS = {sx: rnd(sum(d[f'g_per_bird_{sx}'] or 0 for d in FDAYS)) for sx in 'FM'}
FG = {sx: sum(R._GG[(d['sapt'], sx)]['feed_g'] for d in FDAYS) for sx in 'FM'}   # the growth guide by age week (Telița's guide)
FKG = {sx: rnd(sum(d[f'kg_{sx}'] or 0 for d in FDAYS)) for sx in 'FM'}
FDEV = {sx: (FS[sx] - FG[sx]) / FG[sx] * 100 for sx in 'FM'}
assert (FS, FG, len(FDAYS)) == ({'F': 8401, 'M': 10142}, {'F': 9065, 'M': 10835}, 143) and all(abs(v) <= 10 for v in FDEV.values())
# Hîrbovăț (open, real) as S4 shows it (s4s figures)
LAY, LAYG = s4.lay_today, s4.lay_guide; LAYD = rnd(LAY - LAYG, 1)
MH, MORT_G = s4.cum['mortality_pct'], s4.MORT_G
GF, GM, GFG, GMG = s4.g_f, s4.g_m, s4.gf_g, s4.gm_g
FEED_OK = abs(GF / GFG - 1) <= 0.10 and abs(GM / GMG - 1) <= 0.10
STK = s4.stocks; SLOW = s4.ST_LOW
# series mortality (D155 real deaths; Telița + Hîrbovăț) from placement (S3N, Claude, to review: the "Total serie" % cumulat
# counts from the series' placement, as the series' own start; a farm table counts from its own arrival, D162)
ARR_DAY = P.ARR_DAY
TEL_DEATHS = sorted((R.D(r['date']), r['sex'], r['mortality']) for r in MT['rows'] if r['mortality'])
SEXERR_CARVED = [(R.D(c['date']), c['sex'], c['carved_from_mortality_of_day'] - c['mortality_left']) for c in R.ILL['sexing_errors']]
def ser_dead(sx, e):   # deaths (culls included, as the register); a sexing error carved from a day's deaths is not a death (19.05 ♀ 1)
    return (sum(q for d, s, q in TEL_DEATHS if s == sx and d <= e) + (R.deaths(ARR_DAY, e, sx=sx) if e >= ARR_DAY else 0)
            - sum(q for d, s, q in SEXERR_CARVED if s == sx and d <= e))
def ser_cum(sx, e): return ser_dead(sx, e) / PLC[sx] * 100
assert all(ser_dead(sx, R.TODAY) == TM[sx] + s4.cum['dead'][sx] - sum(q for d, s, q in SEXERR_CARVED if s == sx) for sx in 'FM')
assert (ser_dead('F', R.TODAY), ser_dead('M', R.TODAY)) == (2224, 1287)   # = the table's Mortalitate + Sacrificare totals (S3N)
# where the birds were (spine "Fermă"): Telița 07.10.2025–28.02.2026, Hîrbovăț from 22.02.2026
def farms_on(a, b):
    return [f for f, lo, hi in (('Telița', R.D(cT['start_date']), R.D(cT['end_date'])), ('Hîrbovăț', R.D(cH['start_date']), R.TODAY))
            if lo <= b and a <= hi]
def ferma(kind, key):
    a, b = T.span(kind, key); return ' · '.join(farms_on(a, b))
assert [ferma('w', k) for k in (19, 20, 21, 22)] == ['Telița', 'Telița · Hîrbovăț', 'Telița · Hîrbovăț', 'Hîrbovăț']

# ---------- small helpers ----------
PLACE = '<path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><path d="m10 17 5-5-5-5"/><path d="M15 12H3"/>'
s4.IC.update({'placement': PLACE}); s4.REC.update({'placement': 'placement'})
def ico(k): return f'<span class="ip-rec ip-rec--{s4.REC[k]} s4-ico">{icon(k, 14)}</span>'
def fig(k, key, val, sub, status=''):
    st_ = f'<dd class="ip-fsum__status">{status}</dd>' if status else ''
    return f'<div class="ip-fsum__fig"><dt>{ico(k)}{key}</dt><dd class="ip-fsum__val">{val}</dd><dd class="ip-fsum__sub">{sub}</dd>{st_}</div>'
PCT = '<small class="ip-fsum__pct">%</small>'
GN = 'ghid: – fără date în ghid'
def sk(cls=''): return f'<span class="ip-skel{(" " + cls) if cls else ""}"></span>'

# ---------- 1. page header ----------
EXPORT = f'<button type="button" class="ip-btn ip-btn--secondary">{icon("download", 16)}Exportă .xlsx</button>'
def header(mode='default', phone=False):
    title = 'Seria 2025' if mode == 'closed' else 'Seria 2026'
    if mode == 'closed':
        meta = (f'{lamp("off", "închisă")}<span class="ip-note">început dedus <span class="ip-num">{ddate(cP["start_date"])}</span> la Hîrbovăț · '
                f'încheiată prin golire <span class="ip-num">{ddate(cP["end_date"])} · {P25}</span></span>'
                f'<span class="ip-note">efectiv – fără date în registru · doar citire</span>')
    elif mode == 'loading':
        meta = f'<span class="s3-skmeta" aria-hidden="true">{sk("s3-w8")}{sk("s3-w14")}{sk("s3-w24")}</span>'
    elif mode == 'viewer':  # only the farms this viewer is assigned to (Hîrbovăț); nothing about Telița leaks
        meta = (f'{lamp("ok", "deschisă")}{pair(n(ALIVE["F"]), n(ALIVE["M"]))}'
                f'<span class="ip-note">în viață la fermele tale · <span class="ip-num">{ddate(TODAY)} · SAPT {WK}</span></span>')
    else:
        meta = (f'{lamp("ok", "deschisă")}{pair(n(SER["F"]), n(SER["M"]))}'
                f'<span class="ip-note">în viață · <span class="ip-num">{ddate(TODAY)} · SAPT {WK}</span></span>'
                f'<span class="ip-note{"" if phone else " s3-sep"}">populată <span class="ip-num">{ddate(PL_DATE)}</span> la Telița</span>'
                f'{pair(n(PLC["F"]), n(PLC["M"]))}')
    exp = (f'<button type="button" class="ip-btn ip-btn--secondary is-disabled" disabled="disabled">{icon("download", 16)}Exportă .xlsx</button>'
           if mode == 'loading' else EXPORT)
    if phone: exp = exp.replace('class="ip-btn ip-btn--secondary', 'class="ip-btn ip-btn--secondary ip-btn--touch', 1)
    cls = ' s2-phead s4ns-phead' if phone else ''
    return f'''<header class="ip-pagehead{cls}">
<div>
<nav class="ip-crumbs" aria-label="Cale"><span aria-current="page">{title}</span></nav>
<h1>{title}</h1>
<div class="ip-row ip-pagehead__meta">{meta}</div>
</div>
<div class="ip-row">{exp}</div>
</header>'''

# ---------- 2. CycleTrack · full ----------
TWS = (1160, 1000)  # S3P: drawn 1:1 for the 1160px content box at 1440 and the 1000px one at 1280 (container query, §65)
TW_P = 358          # S3N phone: the full track drawn 1:1 for the 358 px box (11px text, short labels, 44 px block targets)
LBL = {'telita': 'TELIȚA', 'hirbovat': 'HÎRBOVĂȚ'}
def months(d0, d1):
    out = []; y, m = d0.year, d0.month
    while dt.date(y, m, 1) < d1:
        out.append((dt.date(y, m, 1), f'{m:02d}.{str(y)[2:]}')); y, m = (y + 1, 1) if m == 12 else (y, m + 1)
    return out
def track(mode='default', phone=False):
    """Full CycleTrack: one lane per farm, Seria cycles as link blocks, transfer link 22–28.02.2026, azi + SAPT readout.
    Desktop: two drawings (1160 / 1000) in a container-query box; phone: one 358 drawing."""
    svgs = track_svg(mode, TW_P) if phone else ''.join(track_svg(mode, tw) for tw in TWS)
    if mode == 'closed':
        leg = '<span><svg class="ded" viewBox="0 0 3 12"><line x1="1.5" x2="1.5" y1="0" y2="12"/></svg>început dedus</span>'
    elif mode == 'viewer':
        leg = '<span><i class="today"></i>azi</span>'
    else:
        leg = (f'<span><i class="ip-rec ip-rec--placement"></i>populare</span><span><i class="ip-rec ip-rec--transfer"></i>transfer '
               f'<span class="ip-num">{XSPAN}</span></span><span><i class="today"></i>azi</span>')
    legend = f'<div class="ip-tracklegend" aria-hidden="true">{leg}</div>'
    if mode == 'closed': sub = f'<span class="ip-note ip-num">02.2025–02.2026 · 2 ferme · 1 ciclu</span>'
    elif mode == 'viewer': sub = '<span class="ip-note">doar fermele tale · Telița: nu ai acces</span>'
    else: sub = f'<span class="ip-note ip-num">10.2025–12.2026 · 2 ferme · 2 cicluri</span>'
    return f'''<section class="ip-block" aria-labelledby="h-track">
<div class="ip-sechead"><h2 class="ip-t-heading" id="h-track">Cicluri pe ferme</h2>{sub}</div>
<div class="s4-tight"><div class="ip-trackbox">{svgs}</div>{legend}</div>
</section>'''

XIN_LIVE = [a for a in R.TR['arrivals'] if a['date'] == '2026-02-28']   # MO03 (D39): the last real Telița → Hîrbovăț day, shown as arriving today
assert [a['section'] for a in XIN_LIVE] == ['H7.1', 'H7.2']
def track_svg(mode, TW):
    phone = TW == TW_P
    LX = 72 if phone else 88            # lane-label column
    RO_W = 52 if phone else 56          # "SAPT 59" readout at 11px Mono with its sides
    if mode == 'closed':
        d0, d1 = dt.date(2025, 2, 1), dt.date(2026, 3, 1); farms = ['telita', 'hirbovat']
    else:
        d0, d1 = dt.date(2025, 10, 1), (dt.date(2027, 4, 1) if phone else dt.date(2027, 1, 1))   # phone: room for the readout right of azi
        farms = ['hirbovat'] if mode == 'viewer' else ['telita', 'hirbovat']
    ppd = (TW - LX) / (d1 - d0).days
    def tx(iso): return round(LX + (R.D(iso) - d0).days * ppd, 1)
    y0 = 6 if mode == 'closed' else 22                       # room for the "azi" label only when azi is on the axis
    ly = {f: y0 + 48 * i for i, f in enumerate(farms)}       # lane y; block y = lane + 3, h 20
    H = ly[farms[-1]] + 26
    g = ''
    for f in farms:
        g += f'<text class="lanelbl" x="0" y="{ly[f] + 17}">{LBL[f]}</text><rect class="lane" x="{LX}" y="{ly[f]}" width="{TW - LX}" height="26"/>'
    def block(f, a_iso, b_iso, label, aria, href, open_=False, marks=''):
        xa, xb = tx(a_iso), tx(b_iso); w = round(xb - xa, 1); y = ly[f] + 3
        o = ' open' if open_ else ''
        hit = f'<rect x="{xa}" y="{y - 12}" width="{w}" height="44" fill="none" pointer-events="all"/>' if phone else ''   # 44 px touch target
        return (f'<a href="{href}" aria-label="{aria}"><title>{aria}</title>{hit}'
                f'<rect class="ring" x="{round(xa - 4, 1)}" y="{y - 4}" width="{round(w + 8, 1)}" height="28" rx="3"/>'
                f'<rect class="cyc{o}" x="{xa}" y="{y}" width="{w}" height="20" rx="2"/>'
                f'<text class="cyctxt{o}" x="{round(xa + 8, 1)}" y="{y + 14}">{label}</text>{marks}</a>')
    if mode == 'closed':
        xa = tx(cP['start_date'])
        ded = f'<line class="ded" x1="{round(xa - 3, 1)}" x2="{round(xa - 3, 1)}" y1="{ly["hirbovat"]}" y2="{ly["hirbovat"] + 26}"/>'
        aria = (f'Hîrbovăț, Seria 2025, ouat, început dedus {ddate(cP["start_date"])} · încheiat prin golire {ddate(cP["end_date"])}, '
                f'{P25.replace("SAPT", "sapt.")}, închis. Deschide ferma la acest ciclu.')
        g += f'<text class="cyctxt" x="{LX + 12}" y="{ly["telita"] + 17}">– fără date pentru Seria 2025 la Telița</text>'
        g += block('hirbovat', cP['start_date'], cP['end_date'], f'Seria 2025 · ouat · {P25}', aria, '#/ferme/hirbovat?ciclu=cyc-hirbovat-2025', marks=ded)
        desc = (f'Seria 2025, închisă. Telița: fără date pentru Seria 2025. Hîrbovăț: ouat, început dedus {ddate(cP["start_date"])}, '
                f'încheiat prin golire {ddate(cP["end_date"])}, {P25.replace("SAPT", "sapt.")}. Azi, {ddate(TODAY)}, e după sfârșitul seriei.')
    else:
        xt = tx(TODAY)
        if 'telita' in farms:
            xa = tx(cT['start_date'])
            place = f'<rect class="place" x="{round(xa - 1, 1)}" y="{ly["telita"] + 3}" width="3" height="20"/>'
            aria = (f'Telița, Seria 2026, creștere, {ddate(cT["start_date"])}–{ddate(cT["end_date"])}, {TEL_SPAN.replace("SAPT", "sapt.")}, închis. '
                    f'Început prin populare (găini {n(PLC["F"])}, cocoși {n(PLC["M"])}), încheiat prin transfer la Hîrbovăț {XSPAN}. Deschide ferma la acest ciclu.')
            g += block('telita', cT['start_date'], cT['end_date'], 'creștere' if phone else f'Seria 2026 · creștere · {TEL_SPAN}', aria,
                       '#/ferme/telita?ciclu=cyc-telita-2026', marks=place)
        src = f' Păsări sosite prin transfer de la Telița {XSPAN} (găini {n(ARR["F"])}, cocoși {n(ARR["M"])}).' if 'telita' in farms else ''
        aria = f'Hîrbovăț, Seria 2026, ouat, din {ddate(cH["start_date"])}, deschis, azi sapt. {WK}.{src} Deschide ferma la acest ciclu.'
        g += block('hirbovat', cH['start_date'], TODAY, f'ouat · din {R.ddm(cH["start_date"])}' if phone else f'Seria 2026 · ouat · din {ddate(cH["start_date"])} · {HIR_SPAN}',
                   aria, '#/ferme/hirbovat?ciclu=cyc-hirbovat-2026', open_=True)
        if 'telita' in farms:  # transfer link: the real movement days 22–28.02.2026, from Telița's block to Hîrbovăț's
            yt, yb = ly['telita'] + 23, ly['hirbovat'] + 3; ym = (yt + yb) / 2
            xa, xb = tx(X0D), tx(X1D)
            g += (f'<g aria-hidden="true"><line class="link" x1="{xa}" y1="{yt}" x2="{xa}" y2="{yb}"/><line class="link" x1="{xb}" y1="{yt}" x2="{xb}" y2="{yb}"/>'
                  f'<line class="link" x1="{xa}" y1="{ym}" x2="{xb}" y2="{ym}"/><circle class="linkdot" cx="{xa}" cy="{ym}" r="3"/><circle class="linkdot" cx="{xb}" cy="{ym}" r="3"/></g>')
        top = ly[farms[0]] - 8; yh = ly['hirbovat'] + 5
        assert xt + 4 + RO_W <= TW
        g += (f'<line class="today" x1="{xt}" x2="{xt}" y1="{top}" y2="{H + 4}"/><text class="todaylbl" x="{xt}" y="{top - 3}" text-anchor="middle">azi {R.ddm(TODAY)}</text>'
              f'<rect class="ro" x="{round(xt + 4, 1)}" y="{yh}" width="{RO_W}" height="16" rx="2"/><text class="rotxt" x="{round(xt + 4 + RO_W / 2, 1)}" y="{yh + 12}" text-anchor="middle">SAPT {WK}</text>')
        if mode == 'xfer':   # MO03 (D197), D39: a transfer arriving today, drawn with the 28.02.2026 H2 → H7.1, H7.2 figures (7 001 găini)
            xq = sum(a['qty'] for a in XIN_LIVE); xs = ', '.join(a['section'] for a in XIN_LIVE)
            yt2, yb2 = ly['telita'] + 23, ly['hirbovat'] + 3; ym2 = (yt2 + yb2) / 2
            g += (f'<g class="ip-track__live" data-live-rec="transfer"><line class="link" x1="{xt}" y1="{yt2}" x2="{xt}" y2="{yb2}"/>'
                  f'<circle class="linkdot" cx="{xt}" cy="{ym2}" r="3"/><text class="xferlbl" x="{round(xt - 8, 1)}" y="{ly["telita"] + 17}" text-anchor="end">'
                  f'Transfer azi · Telița H2 → {xs} · {n(xq)} găini</text></g>')
        if mode == 'viewer':
            desc = f'Seria 2026, doar fermele la care ai acces. Hîrbovăț: ouat din {ddate(cH["start_date"])}, deschis, azi {ddate(TODAY)}, sapt. {WK}.'
        else:
            desc = (f'Seria 2026, 2 ferme. Telița: creștere {ddate(cT["start_date"])}–{ddate(cT["end_date"])}, {TEL_SPAN.replace("SAPT", "sapt.")}, închis, început prin populare. '
                    f'Transfer Telița → Hîrbovăț {XSPAN}: găini {n(ARR["F"])}, cocoși {n(ARR["M"])}. '
                    f'Hîrbovăț: ouat din {ddate(cH["start_date"])}, deschis, azi {ddate(TODAY)}, sapt. {WK}.')
            if mode == 'xfer':
                desc += f' Transfer azi: Telița H2 → {", ".join(a["section"] for a in XIN_LIVE)}, găini {n(sum(a["qty"] for a in XIN_LIVE))} (exemplu).'
    tk = ''
    for d, lbl in months(d0, d1):
        x = tx(d.isoformat()); anc = 'start' if x - LX < 8 else 'middle'
        tk += f'<line class="tick" x1="{x}" x2="{x}" y1="{H}" y2="{H + 6}"/>'
        if not phone or ((d.month - 1) % 3 == 0 and x - LX >= 16): tk += f'<text class="month" x="{x}" y="{H + 20}" text-anchor="{anc}">{lbl}</text>'   # phone: quarter months, none on the axis start
    cls = 'ip-track s3-track' + ('' if phone else f' ip-track--{TW}')
    return (f'<svg class="{cls}" viewBox="0 0 {TW} {H + 24}" role="group" aria-label="Ciclurile seriei {"2025" if mode == "closed" else "2026"} pe ferme">'
            f'<desc>{desc}</desc>{g}{tk}</svg>')

def track_loading():  # same box as the live track (azi room 22 · lane 26 · gap 22 · lane 26 · axis 24, legend 16): no shift when it loads
    lane = f'<div class="s3-sklane">{sk("s3-sklbl")}{sk("s3-skbar")}</div>'
    return f'''<section class="ip-block" aria-hidden="true">
<div class="ip-sechead"><h2 class="ip-t-heading">Cicluri pe ferme</h2></div>
<div class="s4-tight"><div class="s3-sktrack">{lane}{lane}<div class="s3-skaxis"></div></div><div class="s3-sklegend">{sk("s3-w24")}</div></div>
</section>'''

# ---------- 3. farm summaries (FarmSummary: --wide on desktop, --touch on the phone) ----------
TAG = {'telita': ('Creștere', 'Ross 308 creștere'), 'hirbovat': ('Ouat', 'Ross 308 ouat')}
def fs_cls(phone): return 'ip-fsum ip-fsum--touch' if phone else 'ip-fsum ip-fsum--wide'
def fs_head(fid, name, href, lmp, tags=True):  # tags: True = type + guide, 'type' = type only (no cycle → no guide), False = none
    nm = f'<a href="{href}">{name}</a>' if href else name
    tg = ''.join(f'<span class="ip-tag ip-tag--neutral">{t}</span>' for t in (TAG[fid][:1] if tags == 'type' else TAG[fid] if tags else ()))
    return f'<div class="ip-fsum__head"><div class="ip-fsum__id"><h3 class="ip-fsum__name" id="fs-{fid}">{nm}</h3>{tg}</div>{lmp}</div>'
def alive_line(cap, f, m):
    return (f'<div class="ip-fsum__alive"><span class="ip-cap">{cap}</span><div><span class="ip-pair ip-pair--lg"><span class="ip-sex ip-sex--hen">♀ <b>{n(f)}</b></span>'
            f'<span class="sep">·</span><span class="ip-sex ip-sex--rooster">♂ <b>{n(m)}</b></span></span><span class="ip-note">total {n(f + m)}</span></div></div>')
def fs_telita(phone=False):
    figs = (fig('weighing', f'Greutate {F} la transfer', f'{n(WT)}<small>g</small>',
                f'<span class="ip-num">ghid {n(WG)} g · {sgn(WDEV)}% ·</span><span class="ip-num">SAPT {WSAPT} · cântărit {ddate(WDATE)}</span>',
                lamp('warn' if abs(WDEV) > 5 else 'ok', 'abatere' if abs(WDEV) > 5 else 'în normă'))
            + fig('mortality', 'Mortalitate cumulată', f'{pair(n(TMP["F"], 2), n(TMP["M"], 2))}{PCT}', GN)
            + fig('feed', 'Hrană totală', f'{pair(n(FS["F"]), n(FS["M"]))}<small>g</small>',
                  f'<span class="ip-num">ghid {n(FG["F"])} / {n(FG["M"])} g ·</span><span class="ip-num">{sgn(FDEV["F"])}% / {sgn(FDEV["M"])}% · {len(FDAYS)} zile ·</span>'
                  f'<span class="ip-num">{T.SXS.figure("F", n(FKG["F"]))} · {T.SXS.figure("M", n(FKG["M"]))} kg</span>',   # SX01: §72 signs (was one text span per sex, a ~5 px glyph)
                  lamp('ok', 'în normă'))
            + fig('transfer', 'Transferate', pair(n(ARR['F']), n(ARR['M'])), f'<span>sosite la Hîrbovăț ·</span><span class="ip-num">{XSPAN}</span>'))
    return f'''<article class="{fs_cls(phone)}" aria-labelledby="fs-telita" aria-describedby="fs-telita-c">{fs_head('telita', 'Telița', '#/ferme/telita?ciclu=cyc-telita-2026', lamp('off', 'închis'))}
<p class="ip-fsum__cycle" id="fs-telita-c">{R.drange(cT['start_date'], cT['end_date'])} · {TEL_SPAN}</p>
{alive_line(f'Populat · {ddate(PL_DATE)}', PLC['F'], PLC['M'])}
<dl class="ip-fsum__figs">{figs}</dl></article>'''
def fs_hirbovat(phone=False):
    lay_ok = LAYD >= -3
    lo = STK[SLOW]; st_ok = lo['days_left'] >= s4.STOCK_LIMIT
    figs = (fig('eggs', 'Ouat azi', f'{n(LAY, 1)}{PCT}', f'<span class="ip-num">ghid {pct(LAYG)} · {sgn(LAYD)} pct · SAPT {WK}</span>',
                lamp('ok' if lay_ok else 'warn', 'în normă' if lay_ok else 'abatere'))
            + fig('mortality', 'Mortalitate cumulată', f'{pair(n(MH["F"], 2), n(MH["M"], 2))}{PCT}',
                  f'<span class="ip-num">ghid {s4.hen(pct(MORT_G, 2))} · din sosire</span>', lamp('ok', 'în normă') if MH['F'] <= MORT_G else lamp('warn', 'abatere'))
            + fig('feed', 'Hrană azi', f'{pair(n(GF), n(GM))}<small>g</small>',
                  f'<span class="ip-num">ghid {GFG} / {GMG} g ·</span><span>{len(s4.FULL)} secțiuni · {s4.EMPTY_NOTE}</span>',
                  lamp('ok', 'în normă') if FEED_OK else lamp('warn', 'abatere'))
            + fig('delivery', 'Zile stoc', f'{pair(n(STK["ft-breeder2"]["days_left"], 1), n(STK["ft-cocosi"]["days_left"], 1))}<small>zile</small>',
                  f'<span>{s4.FT_NAME[SLOW]} până pe <span class="ip-num">{ddate(lo["run_out_date"])}</span>{"" if st_ok else " · sub 5 zile"}</span>',
                  lamp('ok', 'în normă') if st_ok else lamp('warn', 'abatere')))
    return f'''<article class="{fs_cls(phone)}" aria-labelledby="fs-hirbovat" aria-describedby="fs-hirbovat-c">{fs_head('hirbovat', 'Hîrbovăț', '#/ferme/hirbovat?ciclu=cyc-hirbovat-2026', lamp('ok', 'deschis'))}
<p class="ip-fsum__cycle" id="fs-hirbovat-c">ciclu din {ddate(cH['start_date'])} · {HIR_SPAN}</p>
{alive_line(f'În viață · {ddate(TODAY)}', ALIVE['F'], ALIVE['M'])}
<dl class="ip-fsum__figs">{figs}</dl></article>'''
def fs_empty(fid, name, href, word, lines, step, tags=True):
    body = ''.join(f'<p class="ip-m0">{l}</p>' for l in lines)
    return f'''<article class="ip-fsum ip-fsum--wide" aria-labelledby="fs-{fid}" aria-describedby="fs-{fid}-c">{fs_head(fid, name, href, lamp('off', word), tags)}
<div class="ip-fsum__empty" id="fs-{fid}-c">{body}{step}</div></article>'''
def fs_prev():  # Hîrbovăț · Seria 2025 (real previous flock; no flock count, no guide for these figures)
    PT = s4.PT; inc = PT['hatching_eggs'] / PT['total_eggs'] * 100; fin = s4.fin
    figs = (fig('eggs', 'Ouă în ciclu', n(PT['total_eggs']), f'<span class="ip-num">{R.drange(cP["data_start_date"], cP["end_date"])} · {s4.PDAYS} zile ·</span><span>{GN}</span>')
            + fig('eggs', 'Incubabile', f'{n(inc, 1)}{PCT}', f'<span class="ip-num">{n(PT["hatching_eggs"])} din {n(PT["total_eggs"])} ouă ·</span><span>{GN}</span>')
            + fig('mortality', 'Mortalitate cumulată', pair(n(s4.PMORT['F']), n(s4.PMORT['M'])),
                  f'<span>% – {s4.NOFLOCK} ·</span><span>fără golirea din <span class="ip-num">{ddate(s4.PDEP)}</span></span>')
            + fig('eggs', f'Masă medie ou · SAPT {s4.PFIN}', f'{n(fin["avg_egg_weight_g"], 1)}<small>g</small>', GN))
    return f'''<article class="ip-fsum ip-fsum--wide" aria-labelledby="fs-hirbovat" aria-describedby="fs-hirbovat-c">{fs_head('hirbovat', 'Hîrbovăț', '#/ferme/hirbovat?ciclu=cyc-hirbovat-2025', lamp('off', 'închis'))}
<p class="ip-fsum__cycle" id="fs-hirbovat-c">{R.drange(cP['start_date'], cP['end_date'])} · {P25} · început dedus</p>
<div class="ip-fsum__alive"><span class="ip-cap">Populat · început dedus {ddate(cP['start_date'])}</span><div><span class="ip-pair ip-pair--lg"><span class="ip-sex"><b>–</b></span></span><span class="ip-note">{s4.NOFLOCK}</span></div></div>
<dl class="ip-fsum__figs">{figs}</dl></article>'''
NOACC = ['Nu ai acces la această fermă.', 'Cere acces proprietarului.']
def summaries(mode='default', phone=False):
    if mode == 'closed':
        tel = fs_empty('telita', 'Telița', '#/ferme/telita', 'fără date', ['Registrul nu are date pentru Seria 2025 la Telița.'],
                       '<a class="ip-link" href="#/serii/seria-2026">Vezi Seria 2026</a>', tags='type')
        cards, note = tel + fs_prev(), '2 ferme · Seria 2025'
    elif mode == 'viewer':
        cards, note = fs_empty('telita', 'Telița', '', 'fără acces', NOACC, '', tags=False) + fs_hirbovat(), 'fermele tale: Hîrbovăț'
    else:
        cards, note = fs_telita(phone) + fs_hirbovat(phone), f'2 ferme · {ddate(TODAY)}'
    return f'''<section class="ip-block" aria-labelledby="h-farms">
<div class="ip-sechead"><h2 class="ip-t-heading" id="h-farms">Ferme</h2><span class="ip-note{" ip-num" if any(c.isdigit() for c in note) else ""}">{note}</span></div>
<div class="ip-fsums{"" if phone else " ip-fsums--stretch"}">{cards}</div>
</section>'''
def summaries_loading():
    card = (f'<div class="ip-fsum ip-fsum--wide s3-skcard"><div class="ip-fsum__head">{sk("s3-w30 s3-h16")}{sk("s3-w14")}</div>{sk("s3-w40")}'
            f'{sk("s3-w50 s3-h20")}' + ''.join(f'<div class="s3-skfig"><div class="s3-skrow">{sk("s3-w40")}{sk("s3-w20")}</div>{sk("s3-w60")}</div>' for _ in range(4)) + '</div>')
    return f'''<section class="ip-block" aria-hidden="true">
<div class="ip-sechead"><h2 class="ip-t-heading">Ferme</h2></div>
<div class="ip-fsums ip-fsums--stretch">{card}{card}</div>
</section>'''

# ---------- 4. tabs + the Păsări tab: PeriodTable (Total serie) + its notes, then its charts (D174: right under the track) ----------
TABS = [('pasari', 'Păsări'), ('hrana', 'Hrană'), ('greutate', 'Greutate'), ('productie', 'Producție')]
def tabs(series='2026'):
    return f'<div class="ip-tabs" role="tablist" aria-label="Vederi Seria {series}">' + ''.join(
        f'<a class="ip-tab{" is-selected" if i == 0 else ""}" role="tab" id="tab-{k}" aria-controls="tp-serie" aria-selected="{"true" if i == 0 else "false"}" tabindex="{0 if i == 0 else -1}" href="#{k}">{t}</a>'
        for i, (k, t) in enumerate(TABS)) + '</div>'

WEEKS = P.WEEKS; assert WEEKS == [59, 58, 57, 56]
SER_CAP = f'Seria 2026 · Total serie · SAPT {WEEKS[-1]}–{WEEKS[0]}'
def gname(g): return re.sub(r'<[^>]+>', '', g['name']).strip()
def wend(k): return min(R.week_dates(k)[1], R.TODAY)

# --- series counts, each bird once (D174): Telița holds a batch until it arrives at Hîrbovăț (transfers 22–28.02.2026) ---
_TS = {}                                                 # (date, section) -> (sex, alive_sheet) of Telița's register
for r in MT['rows']: _TS[(R.D(r['date']), r['section'])] = (r['sex'], r['alive_sheet'])
TSECS = sorted({r['section'] for r in MT['rows']})
LINK = {}                                                # Telița section -> (out day, arrival day, qty out)
for l in R.TR['links_inferred_by_date']:
    for s in l['from']:
        LINK[s] = (R.D(l['out_date']), R.D(l['arrival_date']), next(o['qty'] for o in R.TR['transfers_out'] if o['from_section'] == s and o['sex'] == l['sex']))
assert set(LINK) == set(TSECS)
SEX = {s: next(r['sex'] for r in MT['rows'] if r['section'] == s) for s in TSECS}
def tel_pending(sx, d):
    """Telița's birds of sex sx at the end of day d not yet arrived at Hîrbovăț: a section counts its register until its
    batch leaves, the batch in transit until it arrives (H2: out 27.02, in 28.02); what stays behind after the batch left
    (residuals, the 277 hens moved into H5-1 on 23.02) is the register difference at the close (D25/D63), not counted."""
    if d > R.D(cT['end_date']): return 0
    out = 0
    for s in TSECS:
        if SEX[s] != sx: continue
        o, a, q = LINK[s]
        if d < o: out += max(_TS.get((d, s), (sx, 0))[1], 0)
        elif d < a: out += q
    return out
def ser_alive(sx, d): return tel_pending(sx, d) + R.alive_farm(sx, d)
MOVED = next(r for r in MT['rows'] if r['date'] == '2026-02-23' and r['section'] == 'H5-1')['correction']
assert MOVED == 277 and next(r for r in MT['rows'] if r['date'] == '2026-02-23' and r['section'] == 'H6')['correction'] == -277
assert (ser_alive('F', R.D('2026-02-22')), ser_alive('M', R.D('2026-02-23'))) == (27381, 3024) and (ser_alive('F', R.TODAY), ser_alive('M', R.TODAY)) == (12561, 794)
assert all(ser_alive(sx, d) <= PLC[sx] for sx in 'FM' for d in (R.D('2026-02-2' + str(i)) for i in range(2, 9)))

# --- Păsări groups at series level, by span (kind 'w' week, 'd' day, 'f' the period [FOOT]) ---
CARVE = R.ILL['culls'] + R.ILL['sexing_errors']           # illustrative records carved from that day's real deaths (DATA03)
TEL_SE = [(R.D(r['date']), r['sex'], r['sexing_error']) for r in MT['rows'] if r['sexing_error']]
def sp(kind, key): return T.span(kind, key)              # 'f': the loaded weeks (SYS18: the foot adds up the weeks on screen)
def inn(d, a, b): return a <= R.D(d) <= b
def g_alive(s): return lambda kind, key: n(ser_alive(s, sp(kind, key)[1]))
def g_mort(s):
    def f(kind, key):
        a, b = sp(kind, key)
        tel = sum(q for d, x, q in TEL_DEATHS if x == s and a <= d <= b)
        hir = R.deaths(max(a, ARR_DAY), b, sx=s) if b >= ARR_DAY else 0
        cv = sum(c['carved_from_mortality_of_day'] - c['mortality_left'] for c in CARVE if c['sex'] == s and inn(c['date'], a, b))
        return n(tel + hir - cv)
    return f
def g_cum(s): return lambda kind, key: pct(ser_cum(s, sp(kind, key)[1]))
def g_cull(s): return lambda kind, key: n(sum(c['qty'] for c in R.ILL['culls'] if c['sex'] == s and inn(c['date'], *sp(kind, key))))
def g_sexerr(s):   # signed as typed (−1 = one bird out of this sex)
    def f(kind, key):
        a, b = sp(kind, key)
        return n(sum(q for d, x, q in TEL_SE if x == s and a <= d <= b) + sum(c['qty'] for c in R.ILL['sexing_errors'] if c['sex'] == s and inn(c['date'], a, b)))
    return f
def g_place(s): return lambda kind, key: n(PLC[s] if inn(PL_DATE, *sp(kind, key)) else 0)
def g_sale(s): return lambda kind, key: n(R.sold(*sp(kind, key), sx=s))
def grp(name, fn, icon_=None):
    return {'name': (P.ico(icon_) if icon_ in P.IC else (T.thico(icon_) if icon_ else '')) + name,
            'leaves': [(P.sx(s), P.SX[s][0], fn(s)) for s in 'FM']}
SEXERR_ICO = f'<span class="ip-table__ico">{icon("sexerr", 14)}</span>'
def ser_groups():
    """spec D1 › Series page (D63/D45): every group the period has records for, D167 order. SYS18 (D190): the period is the
    whole series (SAPT 1–59) and the table loads it 4 weeks at a time, so the groups are the series' and stay put while weeks
    load: În viață · Mortalitate · % cumulat · Sacrificare · Eroare sexare · Populare · Vânzare. A transfer between the series'
    farms leaves the series total unchanged (each bird counted once, D174): no Transfer group."""
    return [grp('În viață', g_alive), grp('Mortalitate', g_mort, 'mortality'), grp('% cumulat', g_cum, 'mortality'), grp('Sacrificare', g_cull, 'cull'),
            {'name': SEXERR_ICO + 'Eroare sexare', 'leaves': grp('', g_sexerr)['leaves']}, grp('Populare', g_place, 'placement'), grp('Vânzare', g_sale, 'sale')]
_tp = T.pasari()   # the farm table's figures for SAPT 56–59 equal the series' (Telița closed): one check per group
for _i, _g in zip((0, 1, 3, 6), (g_alive, g_mort, g_cull, g_sale)):
    assert all(_tp[_i]['leaves'][j][2]('w', k) == _g(s)('w', k) for j, s in enumerate('FM') for k in WEEKS)
DENSE = False  # S3N: measured: the series table with its Fermă column fits the 1440 page without the dense cells (D165: only where needed)
def notes(note):   # SYS18: the notes under the table foot (the load-more key and its count come first, P.more)
    return '<div class="ip-ptable__after">' + ''.join(f'<p class="ip-note">{x}</p>' for x in note) + '</div>'
CULL = T.PCULLS[0]
NOTE_SER = [f'Mai vechi: Telița în {TEL_SPAN}, ambele ferme în săptămânile transferului ({num(XSPAN)}). Eroare sexare și Populare au înregistrări în serie ({num(ddate(PL_DATE))}, {num("23.02")}, {num("19.05")}).',
            f'% cumulat = morți din populare ({num(ddate(PL_DATE))}) ÷ păsări populate {pair(n(PLC["F"]), n(PLC["M"]))}. '
            f'Golirea din {num(R.ddm(R.EMPTYING_SINCE))} e Vânzare; Mortalitate = doar morți reali.',
            f'Sacrificarea din {num(R.ddm(CULL["date"]))} ({CULL["section"]} {M} {num(CULL["qty"])}) e ilustrativă și e scăzută din morții zilei. '
            f'Telița se închide la 0 pe {num(ddate(cT["end_date"]))}: diferența din registru {pair(n(TDIFF["F"]), n(TDIFF["M"]))} nu e o corecție înregistrată.']
NOTE_ALL = [f'Transfer {num(XSPAN)}: fiecare pasăre e numărată o singură dată, la Telița până ajunge lotul ei la Hîrbovăț; '
            f'transferul dintre fermele seriei nu schimbă totalul, deci nu are coloană aici (îl arată paginile fermelor).',
            f'{num(n(MOVED))} de găini mutate pe {num("23.02")} din H6 în H5-1 (secția cocoșilor) sunt trecute la cocoși în registrul Teliței și nu ajung '
            f'la Hîrbovăț: intră în diferența din registru la închidere {pair(n(TDIFF["F"]), n(TDIFF["M"]))}, nu în Mortalitate.',
            f'% cumulat = morți din populare ÷ păsări populate. La Telița sacrificările sunt în Mortalitate (registrul nu le ține separat); '
            f'eroarea de sexare din {num("23.02")} e a Teliței, cea din {num("19.05")} e ilustrativă.']
def tb_block(uid, cap, t, after, cls=''):
    t = t.replace(s4.PART[0], s4.PART[1]) if s4.PART[0] in t else t   # D165: the partial week's visible dates "17–18.11"; its hidden name keeps the range
    return (f'<section{f" class={chr(34)}{cls}{chr(34)}" if cls else ""} aria-labelledby="{uid}-h"><h2 class="ip-sr" id="{uid}-h">Păsări: tabel pe săptămâni și zile</h2>'
            + P.block(uid, cap, t, bar=False, after=after) + '</section>')
def dense(t):
    k = 'class="ip-table ip-table--rowheads ip-ptable"'; assert t.count(k) == 1
    return t.replace(k, 'class="ip-table ip-table--rowheads ip-ptable ip-ptable--dense"', 1)
SER_FIRST = 1                                            # the series' first age week (placement 07.10.2025): SAPT 1–59 = 59 weeks
W8 = [WK - i for i in range(8)]                          # after one press: SAPT 59–52
def tb_serie(weeks=None, added=None):
    weeks = weeks or WEEKS
    cap = f'Păsări · Seria 2026 · Total serie · SAPT {weeks[-1]}–{weeks[0]}'
    more = P.more(first=SER_FIRST, word='seriei', added=added, focus=bool(added))   # SYS18 (D190, D192): 4 din 59
    t, tid = P.ptable('ps', cap, [('Total serie', '', ser_groups())], spine='s3', farm=ferma, weeks=weeks, more=more)
    assert t.count('aria-expanded="true"') == len(weeks)                                     # D166 every week open
    if weeks == WEEKS: assert t.count('<td>Hîrbovăț</td>') == 4 + 2 + 21 + 1               # SAPT 56–59 at Hîrbovăț
    return tb_block('ps', cap, dense(t) if DENSE else t, notes(NOTE_SER))
AWEEKS = [22, 21, 20, 19]   # D174: the start board, the region scrolled to the transfer weeks
SALL = list(range(WK, SER_FIRST - 1, -1))                 # SAPT 59 … 1: every week loaded (14 presses: 4 × 14 = 56, the last adds SAPT 3–1)
def tb_all():
    """SYS18 (D190): every week loaded (SAPT 59 → 1) and open (D166), the series' first week reached, so the key gave way to
    the note "Începutul seriei · SAPT 1" (focus moved to it, the live region said so); the region shown scrolled to SAPT 22–19
    (Fermă Telița → Hîrbovăț): the weeks above and below sit behind one muted line each; the foot adds up the 59 weeks on
    screen (0–407, SAPT 1–59, Telița · Hîrbovăț)."""
    cap = f'Păsări · Seria 2026 · Total serie · SAPT 1–{WK}'
    t, tid = P.ptable('ps', cap, [('Total serie', '', ser_groups())], spine='s3', farm=ferma, weeks=AWEEKS, fweeks=SALL,
                      more=P.more(first=SER_FIRST, word='seriei', added=SALL[-3:]))
    assert f'<td class="num">0{DASH}{R.age_days(R.TODAY)}</td><td class="num">1{DASH}{WK}</td><td>Telița · Hîrbovăț</td>' in t   # the foot's spine: the whole series
    ncol = 4 + sum(len(g['leaves']) for g in ser_groups())
    def gap(txt): return f'<tbody class="s3-gap"><tr class="ip-ptable__day ip-ptable__none"><td colspan="{ncol}"><span>{txt}</span></td></tr></tbody>'
    above, below = WK - AWEEKS[0], AWEEKS[-1] - 1
    i = t.index('<tbody class="ip-ptable__week">'); j = t.index('<tbody class="ip-ptable__morebody">')   # SYS18: the weeks below sit above the start note
    t = t[:i] + gap(f'SAPT {WK}–{AWEEKS[0] + 1} mai sus · {above} săptămâni, fiecare deschisă cu zilele ei') + t[i:j] + gap(f'SAPT {AWEEKS[-1] - 1}–1 mai jos · {below} săptămâni') + t[j:]
    assert t.count('aria-expanded="true"') == 4 and '<td>Telița · Hîrbovăț</td>' in t
    return tb_block('ps', cap, t, notes(NOTE_ALL))
def tb_viewer():   # Observator (Hîrbovăț only): the farms it can see; % cumulat on the farm's own basis (from arrival, D162)
    cap = f'Păsări · Seria 2026 · fermele tale (Hîrbovăț) · SAPT {WEEKS[-1]}–{WEEKS[0]}'
    t, tid = P.ptable('ps', cap, [('Fermele tale', '', T.pasari())], spine='s3', farm=ferma, more=P.more())   # 4 din 40 (Hîrbovăț SAPT 20–59)
    note = [f'Doar fermele la care ai acces: Hîrbovăț, de la sosirea păsărilor (SAPT {P.CYCLE_FIRST}).',
            f'% cumulat = morți din sosire ÷ păsări sosite. Golirea din {num(R.ddm(R.EMPTYING_SINCE))} e Vânzare; Mortalitate = doar morți reali.']
    return tb_block('ps', cap, dense(t) if DENSE else t, notes(note))
def tb_closed():   # Seria 2025: the register has deaths and eggs, no flock count, no culls or sales (D52, D54); D174: content width
    cap = f'Păsări · Seria 2025 · Total serie · SAPT {s4.PW4[-1]}–{s4.PW4[0]}'
    groups = [{'name': f'{P.ico("mortality")}Mortalitate', 'leaves': [(P.sx('F'), 'is-hen', s4.p_mort('F')), (P.sx('M'), 'is-rooster', s4.p_mort('M'))]}]
    t, tid = P.ptable('ps', cap, [('Total serie', '', groups)], spine='s3', weeks=s4.PW4, cal=s4.PCAL, more=P.more(first=s4.PWEEKS[0], word='registrului'))   # 4 din 43
    note = [f'Registrul are date din {num(ddate(cP["data_start_date"]))} (SAPT {s4.PWEEKS[0]}); începutul ciclului e dedus ({num(ddate(cP["start_date"]))}, SAPT {cP["sapt_start"]}).',
            f'Fără efectiv în registru: fără În viață și % cumulat, deci fără graficele lor. Golirea din {num(ddate(s4.PDEP))} nu apare ca mortalitate; nicio sacrificare sau vânzare în registru.']
    return tb_block('ps', cap, t, notes(note), cls='s3-fit')

# --- the tab's charts: Mortalitate cumulată and În viață over the series' age weeks (Telița → Hîrbovăț), as S4's Readouts ---
SWEEKS = list(range(1, WK + 1))
aS = {sx: {k: ser_alive(sx, wend(k)) for k in SWEEKS} for sx in 'FM'}
mS = {sx: {k: ser_cum(sx, wend(k)) for k in SWEEKS} for sx in 'FM'}
assert (aS['F'][WK], aS['M'][WK]) == (12561, 794) and all(aS[s][k] <= PLC[s] for s in 'FM' for k in SWEEKS) and max(mS['M'].values()) <= 40
XT = (1, 10, 20, 30, 40, 50, WK)
def schart(*a, **k):   # s4t.gchart with the series' first age week (its x axis starts at T.WMIN)
    old = T.WMIN; T.WMIN = 1
    try: return T.gchart(*a, **k)
    finally: T.WMIN = old
HB, RB = T.HB, T.RB
NBS = ' '   # a sign and its figure never split (phone Readout sub-lines)
def ro_ser_alive(chart=True):
    rows = [[str(k), n(aS['F'][k]), n(aS['M'][k]), ferma('w', k)] for k in SWEEKS]
    lbl = f'Păsări în viață la sfârșitul săptămânii, găini și cocoși, Seria 2026, sapt. 1–{WK}; Telița până în sapt. {cT["sapt_end"]}, Hîrbovăț din sapt. {cH["sapt_start"]}'
    ch = schart('salive', lbl, WK, [{'cls': 'act', 'pts': aS['F']}, {'cls': 's2', 'pt': 'pt2', 'pts': aS['M']}], 0, 30000, [0, 10000, 20000, 30000], lambda t: n(t), x0=52, xticks=XT) if chart else None
    return T.readout('salive', f'ÎN VIAȚĂ · SERIA 2026 · SAPT {WK}', None, T.rpair(n(aS['F'][WK]), n(aS['M'][WK])),
                     f'populate ♀{NBS}{n(PLC["F"])} · ♂{NBS}{n(PLC["M"])} · vânzări ♀{NBS}{n(-T.SOLD["F"])} · ♂{NBS}{n(-T.SOLD["M"])}',
                     ch, f'SAPT {WK} · ♀ {n(aS["F"][WK])} · ♂ {n(aS["M"][WK])}' if chart else '',
                     ['SAPT', 'Găini ♀', 'Cocoși ♂', 'Fermă'], rows, 'Păsări în viață la sfârșitul fiecărei săptămâni de vârstă, fiecare pasăre o dată (în transfer: la Telița până ajunge la Hîrbovăț); SAPT 59: 18.11.2026',
                     T.legend([('act', f'{HB} găini'), ('s2', f'{RB} cocoși')], 'axa: SAPT de vârstă · Telița 1–21 · Hîrbovăț din 20'))
def ro_ser_mort(chart=True):
    rows = [[str(k), pct(mS['F'][k], 2), pct(mS['M'][k], 2)] for k in SWEEKS]
    lbl = f'Mortalitate cumulată găini și cocoși din populare, Seria 2026, sapt. 1–{WK}; morți reali la Telița și Hîrbovăț'
    ch = schart('smort', lbl, WK, [{'cls': 'act', 'pts': mS['F']}, {'cls': 's2', 'pt': 'pt2', 'pts': mS['M']}], 0, 40, [0, 10, 20, 30, 40], lambda t: f'{t}%', xticks=XT) if chart else None
    return T.readout('smort', f'MORTALITATE CUMULATĂ · SERIA 2026 · SAPT {WK}', None, T.rpair(n(mS['F'][WK], 2), n(mS['M'][WK], 2), '%'),
                     f'morți din populare ♀{NBS}{n(ser_dead("F", R.TODAY))} · ♂{NBS}{n(ser_dead("M", R.TODAY))} · ghidul se compară pe fermă',
                     ch, f'SAPT {WK} · ♀ {pct(mS["F"][WK], 2)} · ♂ {pct(mS["M"][WK], 2)}' if chart else '',
                     ['SAPT', 'Găini ♀', 'Cocoși ♂'], rows, f'Mortalitate cumulată pe săptămânile de vârstă, % din păsările populate (♀ {n(PLC["F"])} · ♂ {n(PLC["M"])}); morți reali, fără vânzări',
                     T.legend([('act', f'{HB} găini'), ('s2', f'{RB} cocoși')], 'axa: SAPT de vârstă'))
def ser_charts(chart=True): return T.charts(('h-smort', 'Mortalitate cumulată', ro_ser_mort(chart)), ('h-salive', 'În viață', ro_ser_alive(chart)))   # S4 Păsări order
def panel(mode='default'):
    if mode == 'closed': body = tb_closed()
    elif mode == 'viewer': body = tb_viewer() + '\n' + T.charts(('h-mort', 'Mortalitate cumulată', T.ro_mort()), ('h-alive', 'În viață', T.ro_alive()))
    elif mode == 'all': body = tb_all() + '\n' + ser_charts()
    elif mode == '8': body = tb_serie(W8, added=W8[4:]) + '\n' + ser_charts()
    else: body = tb_serie() + '\n' + ser_charts()
    return f'<div class="s4-panel s3-panel" id="tp-serie" role="tabpanel" aria-labelledby="tab-pasari">\n{body}\n</div>'
# loading: the same table, every week open, bars at the loaded column widths (D132, D167 d)
SKEL = [4, 4, 2, 1, 4, 4, 1, 1, 1, 1, 1, 1, 4, 3]   # ch of each leaf's usual figure (ch1–ch4, §42): În viață ♀ ♂, Mortalitate, % cumulat, Sacrificare, Eroare sexare, Populare, Vânzare
# §54 --ip-colw on the loading table's heads (Vârstă, SAPT, then the 10 leaves): the loaded columns' px at 1440 (s3_render.py, S3N) as shares of
# the 1158 px region, so every column keeps its loaded width when the data arrives; the Data column takes the rest
# SYS18: the series' groups (Eroare sexare, Populare) make the loaded table 1397 px, wider than the 1158 region: the loading
# board stays 1440 (D167 d, as S4's), every figure column keeps its loaded px and the Data column takes the difference
COLW = [round(x / 1158 * 100, 2) for x in (110.1, 63.1, 72, 63.1, 67.4, 59.6, 55.3, 63.1, 59.8, 63, 69.3, 72.5, 52.3, 55.5, 71, 63.1)]
def panel_loading():
    cap = f'Păsări · {SER_CAP}'
    g = ser_groups(); assert len(SKEL) == sum(len(r['leaves']) for r in g)
    t, _ = P.ptable('ps', cap, [('Total serie', '', g)], spine='s3', farm=ferma, skel=SKEL, colw=COLW, more=P.more(first=SER_FIRST, word='seriei'))
    if DENSE: t = dense(t)
    ro = s4.SKL.chart_readout('s4-readout', f'0 0 {T.CW} {T.CH}')
    return f'''<div class="s4-panel s3-panel" id="tp-serie" role="tabpanel" aria-labelledby="tab-pasari" aria-busy="true">
<p class="ip-sr" role="status">Se încarcă Seria 2026: ciclurile, tabelul Păsări (SAPT {WEEKS[-1]}–{WEEKS[0]}) și fermele.</p>
<section aria-hidden="true"><p class="ip-cap ip-ptable__cap">{cap}</p>{P.region(t, cap, hidden=True)}</section>
<div class="s4t-split" aria-hidden="true"><section>{ro}</section><section>{ro}</section></div>
</div>'''

# ---------- page (D174: header → CycleTrack → tabs → table + notes → Readouts → Ferme) ----------
def page(mode='default'):
    series = '2025' if mode == 'closed' else '2026'
    hmode = 'default' if mode in ('all', '8', 'xfer') else mode
    if mode == 'loading':
        top, pn, farms = track_loading(), panel_loading(), summaries_loading()
    else:
        top, pn, farms = track('xfer' if mode == 'xfer' else hmode), panel(mode), summaries(hmode)
    main = f'''<main class="ip-shell__main s4-main">
{header(hmode)}
{top}
<section class="ip-block" aria-label="Date pe săptămâni">
{tabs(series)}
{pn}
</section>
{farms}
</main>'''
    return s4.shell(main, viewer=(mode == 'viewer'), series=series)

CSS = T.CSS + '''
.s3-sep{margin-left:var(--space-2);padding-left:var(--space-3);border-left:var(--ip-hair) solid var(--line)}
.s3-panel{padding-top:var(--space-1)}
.s3-skmeta{display:flex;gap:var(--space-3);align-items:center;min-height:var(--space-5)}
.s3-sklane{display:grid;grid-template-columns:calc(var(--space-12) + var(--space-8) + var(--space-2)) minmax(0,1fr);align-items:center}
.s3-sktrack{padding-top:calc(var(--space-5) + 2 * var(--ip-hair))}
.s3-sklane + .s3-sklane{margin-top:calc(var(--space-5) + 2 * var(--ip-hair))}
.s3-sklbl{width:60%;display:block}
.s3-skbar{height:calc(var(--space-6) + var(--ip-hair) + var(--ip-hair));display:block}
.s3-skaxis{height:var(--space-6)} /* the month axis room, left empty */
.s3-sklegend{height:var(--space-4);display:flex;align-items:center}
.s3-skcard{gap:var(--space-3)}
.s3-skfig{display:flex;flex-direction:column;gap:var(--space-2);padding:var(--space-2) 0;border-top:var(--ip-hair) solid var(--rule)}
.s3-skrow{display:flex;justify-content:space-between}
.s3-h16{height:var(--space-4)}.s3-h20{height:var(--space-5)}
.s3-w8{width:calc(var(--space-12) + var(--space-6))}.s3-w14{width:calc(var(--space-12) * 2)}.s3-w24{width:calc(var(--space-12) * 4)}
.s3-w20{width:20%}.s3-w30{width:30%}.s3-w40{width:40%}.s3-w50{width:50%}.s3-w60{width:60%}
.s3-fit .ip-tablescroll{width:fit-content;max-width:100%} /* D174: a narrow table (Seria 2025: one group) takes its content width, no slack Data column */
.s3-fit .ip-ptable{width:auto}
.s3-gap .ip-ptable__none > td{border-bottom:var(--ip-hair) solid var(--line)} /* the start board: the weeks above / below the scrolled view */'''

# ---------- phone 390 (owner, D174): single-choice record tags as the Fermă phone (default Mortalitate: În viață + Mortalitate
# with % cumulat on its cells' second line); no Fermă column: in the transfer weeks the farms ride on the sticky cell's 2nd line ----------
sys.argv = ['s2_build.py', str(ROOT), str(HERE / 's2_heights.json')]
A = importlib.import_module('s2_build')   # phone shell; rewrites the S2/S7 default boards byte-identical (as s4ns)
PTAGS = [('mortality', 'Mortalitate'), ('cull', 'Sacrificare'), ('sexerr', 'Eroare sexare'), ('placement', 'Populare'), ('sale', 'Vânzare')]   # the series' Păsări types (D45, SYS18: the period is the series)
def radio_tags(on='mortality'):
    def one(k, w):
        a = f'role="radio" aria-checked="{"true" if k == on else "false"}" tabindex="{"0" if k == on else "-1"}"'
        if k == 'sexerr':   # the neutral split chip (D47)
            return (f'<button type="button" class="ip-chip ip-chip--touch ip-rec ip-chip--split{"" if k == on else " ip-chip--off"}" {a}>'
                    f'<span class="ip-tile ip-tile--split" aria-hidden="true">{icon("sexerr", 16)}</span>{w}</button>')
        return f'<button type="button" class="ip-chip ip-chip--touch ip-rec ip-rec--{s4.REC[k]}{"" if k == on else " ip-chip--off"}" {a}>{icon(k, 16)}{w}</button>'
    out = ''.join(one(k, w) for k, w in PTAGS)
    return (f'<div class="ip-field s3-ptags" role="group" aria-labelledby="pt-tags"><span class="ip-label" id="pt-tags">Tipuri de înregistrări</span>'
            f'<div class="ip-chips ip-chips--wrap ip-chips--fill" role="radiogroup" aria-labelledby="pt-tags" aria-describedby="pt-tags-h">{out}</div>'   # SYS20: 2 equal columns
            f'<span class="ip-hint" id="pt-tags-h">Pe telefon tabelul arată un tip odată; În viață rămâne.</span></div>')
def ferma_phone(kind, key):   # only where the farm changes the reading: the transfer days hold both farms
    f = ferma(kind, key); return f if ' · ' in f else None
def phone_page():
    g = ser_groups()
    gp = [g[0], P.stacked(g[1], g[2], '% cumulat')]   # D171/S4NS: % cumulat as the Mortalitate cells' second line
    cap = f'Păsări · {SER_CAP} · Mortalitate'
    t, tid = P.ptable('ps', cap, [('Total serie', '', gp)], spine='phone', farm=ferma_phone, more=P.more(first=SER_FIRST, word='seriei'))
    assert t.count('aria-expanded="true"') == 4 and '<td>Hîrbovăț' not in t
    note = [NOTE_SER[0], f'Toate zilele de aici: Hîrbovăț; în săptămânile transferului ferma apare sub dată. Golirea din {num(R.ddm(R.EMPTYING_SINCE))} e Vânzare.']
    tb = tb_block('ps', cap, t, notes(note))
    pn = f'<div class="s4-panel s3-panel" id="tp-serie" role="tabpanel" aria-labelledby="tab-pasari">\n{radio_tags()}\n{tb}\n{ser_charts(False)}\n</div>'   # Readouts without GuideChart (as S4/S14 phone)
    body = A.phone([header('default', phone=True), track('default', phone=True),
                    f'<section class="ip-block" aria-label="Date pe săptămâni">{tabs()}{pn}</section>', summaries('default', phone=True)])
    for a, z in (('class="ip-tabbar__item is-current" href="#acasa" aria-current="page"', 'class="ip-tabbar__item" href="#acasa"'),
                 ('class="ip-tabbar__item" href="#serie"', 'class="ip-tabbar__item is-current" href="#serie" aria-current="page"'),
                 ('<main class="ip-screen__main s2-main">', '<main class="ip-screen__main s2-main s4-main">'),
                 ('<div class="ip-screen">', '<div class="ip-screen s4ns-phone">')):
        assert body.count(a) == 1, a; body = body.replace(a, z, 1)
    return body
CSS_PHONE = A.CSS_P + '\n' + CSS + """
.s4ns-phone .ip-sechead{flex-direction:column;align-items:flex-start;gap:var(--space-1)}
.s4ns-phone .ip-tab{height:var(--control-h-touch)} /* touch: 48px tabs */
.s4ns-phead{flex-wrap:wrap;gap:var(--space-3)}
.s4ns-phead .ip-pagehead__meta{margin-top:var(--space-2);flex-wrap:wrap}
.s4ns-phead > .ip-row{flex-wrap:wrap}
.s4ns-phone .s4t-split{grid-template-columns:minmax(0,1fr)}
.s4ns-phone .ip-readout__cap{flex-wrap:wrap;row-gap:var(--space-1)}
.s3-ptags{margin:0}"""

B = 'S3-seria-2026'
BOARDS = [
 (B, 'S3 · Seria 2026', 'default'),
 (B + '--se-incarca', 'S3 · Seria 2026 · se încarcă', 'loading'),
 (B + '--seria-inchisa', 'S3 · Seria 2026 · seria închisă (Seria 2025)', 'closed'),
 (B + '--fara-acces', 'S3 · Seria 2026 · fără acces (Observator, doar Hîrbovăț)', 'viewer'),
 (B + '--8-saptamani', 'S3 · Seria 2026 · 8 săptămâni (după „Încarcă încă 4 săptămâni”)', '8'),
 (B + '--inceputul-seriei', 'S3 · Seria 2026 · începutul seriei (încărcat până la SAPT 1, derulat la SAPT 22–19, transfer)', 'all'),
 (B + '--transfer-azi', 'S3 · Seria 2026 · transfer azi (live, D197) · exemplu de stare (D39): transferul Telița H2 → H7.1, H7.2 din 28.02.2026 (7 001 găini) arătat ca sosit azi', 'xfer'),   # MO03
]
PHONE = (B + '--telefon', 'S3 · Seria 2026 · telefon')
if __name__ == '__main__':
    out = []
    for stem, title, mode in BOARDS:
        h = HEIGHTS.get(stem, 1400)
        s4.write(stem, title, page(mode), H=h, css=CSS); out.append((stem, CR.width(stem, 1440), h, title))
    stem, title = PHONE; h = HEIGHTS.get(stem, 844); assert CR.width(stem, 390) == 390   # D174: the phone table fits 358
    A.HEIGHTS[stem] = h
    A.write_phone(stem, title, phone_page(), css=CSS_PHONE)
    out.append((stem, 390, h, title))
    json.dump([{'stem': s, 'w': w, 'h': hh} for s, w, hh, _ in out], open('/tmp/s4s/s3.json', 'w'))
    if CANVAS:  # S3 row: desktops, then the phone (D74); a new board lays the row out again from x 0 (80 px gaps, canvas_rows.place)
        c = CR.load(ROOT)
        CR.place(c, B, [(s, w, hh, t) for s, w, hh, t in out])
        o = c['order']; mine = [f'{s}.dc.html' for s, _, _, _ in out]; i = min(o.index(f) for f in mine if f in o)
        rest = [f for f in o if f not in mine]; c['order'] = rest[:i] + mine + rest[i:]   # the row's boards together, in row order
        if CR.overlaps(c):   # the taller row would overlap the next one: re-stack the rows below first (canvas_rows relayout)
            for m in CR.relayout(c): print('relayout', *m)
        CR.save(ROOT, c)
