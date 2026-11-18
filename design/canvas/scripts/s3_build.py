"""Build S3 · Seria 2026 (series scope page, desktop 1440) and its state boards from the extracts.
python3 -I s3_build.py <repo_root> [heights.json] [--canvas]
Reuses the S4 shell (s4s_build.py: sidebar, icons, SexCount/Lamp helpers, number formats, board writer), so the
sidebar and formats can't drift between scope pages. % without space (D46). Writes
design/canvas/project/S3-seria-2026[--<state>].dc.html (and plain wrappers in /tmp/s4s/)."""
import sys, json, pathlib, importlib.util, datetime as dt

spec = importlib.util.spec_from_file_location('s4s_build', pathlib.Path(__file__).resolve().with_name('s4s_build.py'))
s4 = importlib.util.module_from_spec(spec); spec.loader.exec_module(s4)  # reads sys.argv[1] (repo root)
EX = s4.EX
n, pct, ddate, icon, lamp, pair, sg, F, M = s4.n, s4.pct, s4.ddate, s4.icon, s4.lamp, s4.pair, s4.sg, s4.F, s4.M
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
ROOT, HP, CANVAS = CR.args('s3_heights.json')  # heights default: s3_heights.json
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}

def J(name): return json.load(open(EX / name, encoding='utf-8'))
CY, PL, TR, CL, MT, WW, GG, FD = (J('cycles.json'), J('placement-seria2026.json'), J('transfers-seria2026.json'), J('counts-latest.json'),
                                  J('mortality-telita-daily.json'), J('weighings-weekly.json'), J('guide-growth.json'), J('feed-daily-telita.json'))
S = s4.S
cyc = {c['id']: c for c in CY['cycles']}
cT, cH, cP = cyc['cyc-telita-2026'], cyc['cyc-hirbovat-2026'], cyc['cyc-hirbovat-2025']
TODAY = '2026-11-16'; WK = cH['week_today']
assert WK == 40 and cT['weeks'] == 21 and cT['end_date'] == '2026-02-28' and cH['start_date'] == '2026-02-10'

# ---------- figures (every one from the extracts) ----------
PLC = PL['totals']; assert PLC == {'F': 28080, 'M': 3510}
ARR = TR['arrival_totals']; assert ARR == {'F': 26765, 'M': 3060}           # D28: transfers counted at arrival
XD = sorted({a['date'] for a in TR['arrivals']} | {o['date'] for o in TR['transfers_out']})
X0D, X1D = XD[0], XD[-1]; assert (X0D, X1D) == ('2026-02-22', '2026-02-28')  # D25: real transfer days
hz = next(f for f in CL['farms'] if f['farm_id'] == 'hirbovat'); tz = next(f for f in CL['farms'] if f['farm_id'] == 'telita')
ALIVE = hz['totals']; assert (ALIVE['F'], ALIVE['M']) == (25252, 2685) and tz['totals']['all'] == 0
SER = CL['series'][0]; assert (SER['F'], SER['M']) == (ALIVE['F'], ALIVE['M'])
HC = hz['cycle']; HR = HC['removals']
TM = {'F': MT['meta']['totals']['mortality_F'], 'M': MT['meta']['totals']['mortality_M']}; assert TM == {'F': 704, 'M': 239}
TSE = {'F': 0, 'M': 0}
for r in MT['rows']:
    if r.get('sexing_error'):  # signed as typed (−1 = one bird removed); sex by section
        TSE[r['sex'] if 'sex' in r else ('M' if r['section'].startswith('H5') else 'F')] += -r['sexing_error']
assert TSE == {'F': 0, 'M': 1}, TSE
TMP = {sx: round(TM[sx] / PLC[sx] * 100, 2) for sx in 'FM'}; assert TMP == tz['cycle']['mortality_pct']
# Telița closes at 0 (D25): placement − deaths − sexing errors − transfers (at arrival, D28) doesn't reach 0 in the workbook.
# D63: the gap is a register difference, shown in a "Diferență registru" column with the info lamp — never as a Corecție
# (no correction was recorded) and never as leftover birds.
TDIFF = {sx: -(PLC[sx] - TM[sx] - TSE[sx] - ARR[sx]) for sx in 'FM'}; assert TDIFF == {'F': -611, 'M': -210}, TDIFF
for sx in 'FM':
    assert ARR[sx] - HR[sx]['mortality'] - HR[sx]['culls'] - HR[sx]['sexing_errors'] == ALIVE[sx]
# Telița summary (FarmSummary README): weight ♀ at transfer = SAPT 20 weighings weighted by hens sent out
outF = {o['from_section']: o['qty'] for o in TR['transfers_out'] if o['sex'] == 'F'}
w20 = {r['section']: r for r in WW['rows'] if r['age_week'] == 20 and r['sex'] == 'F' and r['section'] in outF}
WT = round(sum(w20[s]['avg_weight_g'] * outF[s] for s in outF) / sum(outF.values()))
WDATE = {r['date'] for r in w20.values()}; assert WDATE == {'2026-02-24'}
gg = {(r['week'], r['sex']): r for r in GG['rows']}
WG = gg[(20, 'F')]['body_weight_g']; WDEV = round((WT - WG) / WG * 100, 1)
assert (WT, WG, WDEV) == (2531, 2370, 6.8)
days = FD['days']
FS = {sx: round(sum(d[f'g_per_bird_{sx}'] or 0 for d in days)) for sx in 'FM'}
FG = {sx: sum(gg[(d['age_week'], sx)]['feed_g'] for d in days) for sx in 'FM'}
FKG = {sx: round(sum(d[f'kg_{sx}'] or 0 for d in days)) for sx in 'FM'}
FDEV = {sx: round((FS[sx] - FG[sx]) / FG[sx] * 100, 1) for sx in 'FM'}
assert (FS, FG, len(days)) == ({'F': 8401, 'M': 10142}, {'F': 9065, 'M': 10835}, 143), (FS, FG, len(days))
assert all(abs(v) <= 10 for v in FDEV.values())
# Hîrbovăț summary (sample, D27) from the S4 module (same figures as S4)
st = S['summary_today']; MH = st['cumulative_mortality_pct']
stk = {s['feed_type_id']: s for s in st['stocks']}; B2, CO = stk['ft-breeder2'], stk['ft-cocosi']
assert (B2['days_left'], CO['days_left'], CO['run_out_date']) == (7.1, 3.9, '2026-11-19')
assert abs(s4.lay_today - s4.lay_guide) <= 3

# ---------- small helpers ----------
PLACE = '<path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><path d="m10 17 5-5-5-5"/><path d="M15 12H3"/>'
s4.IC.update({'placement': PLACE})
s4.REC.update({'placement': 'placement'})
def ico(k): return f'<span class="ip-rec ip-rec--{s4.REC[k]} s4-ico">{icon(k, 14)}</span>'
def sexerr_ico(): return f'<span class="s4-ico">{icon("sexerr", 14)}</span>'
def hico(k):  # head icon: additions.css §32 span.ip-table__ico (FIN14B; was s4-ico + an S3 copy of the rule)
    return f'<span class="ip-table__ico">{icon("sexerr", 14)}</span>' if k == 'sexerr' else f'<span class="ip-rec ip-rec--{s4.REC[k]} ip-table__ico">{icon(k, 14)}</span>'
def fig(k, key, val, sub, status=''):
    st_ = f'<dd class="ip-fsum__status">{status}</dd>' if status else ''
    return f'<div class="ip-fsum__fig"><dt>{ico(k) if k != "sexerr" else sexerr_ico()}{key}</dt><dd class="ip-fsum__val">{val}</dd><dd class="ip-fsum__sub">{sub}</dd>{st_}</div>'
P = '<small class="ip-fsum__pct">%</small>'
GN = 'ghid: – fără date în ghid'
def sk(cls=''): return f'<span class="ip-skel{(" " + cls) if cls else ""}"></span>'

# ---------- 1. page header ----------
EXPORT = f'<button type="button" class="ip-btn ip-btn--secondary">{icon("download", 16)}Exportă .xlsx</button>'
def header(mode='default'):
    if mode == 'closed':
        crumb, title = 'Seria 2025', 'Seria 2025'
        meta = (f'{lamp("off", "închisă")}<span class="ip-note">început dedus <span class="ip-num">{ddate(cP["start_date"])}</span> la Hîrbovăț · '
                f'încheiată prin golire <span class="ip-num">{ddate(cP["end_date"])}</span> · <span class="ip-num">{cP["weeks"]} săpt.</span></span>'
                f'<span class="ip-note">efectiv – fără date în registru · doar citire</span>')
    elif mode == 'loading':
        crumb, title = 'Seria 2026', 'Seria 2026'
        meta = f'<span class="s3-skmeta" aria-hidden="true">{sk("s3-w8")}{sk("s3-w14")}{sk("s3-w24")}</span>'
    else:
        crumb, title = 'Seria 2026', 'Seria 2026'
        if mode == 'viewer':  # only the farms this viewer is assigned to (Hîrbovăț); nothing about Telița leaks
            meta = (f'{lamp("ok", "deschisă")}{pair(n(ALIVE["F"]), n(ALIVE["M"]))}'
                    f'<span class="ip-note">în viață la fermele tale · <span class="ip-num">{ddate(TODAY)}</span></span>')
        else:
            meta = (f'{lamp("ok", "deschisă")}{pair(n(SER["F"]), n(SER["M"]))}'
                    f'<span class="ip-note">în viață în ambele ferme · <span class="ip-num">{ddate(TODAY)}</span></span>'
                    f'<span class="ip-note s3-sep">populată <span class="ip-num">{ddate(PL["meta"]["date"])}</span> la Telița</span>'
                    f'{pair(n(PLC["F"]), n(PLC["M"]))}')
    exp = (f'<button type="button" class="ip-btn ip-btn--secondary is-disabled" disabled="disabled">{icon("download", 16)}Exportă .xlsx</button>'
           if mode == 'loading' else EXPORT)
    return f'''<header class="ip-pagehead">
<div>
<nav class="ip-crumbs" aria-label="Cale"><span aria-current="page">{crumb}</span></nav>
<h1>{title}</h1>
<div class="ip-row s4-meta">{meta}</div>
</div>
<div class="ip-row">{exp}</div>
</header>'''

# ---------- 2. CycleTrack · full ----------
LX = 88  # lane-label column
TWS = (1160, 1000)  # S3P: drawn 1:1 for the 1160px content box at 1440 and the 1000px one at 1280 (1440/1280 − sidebar 232 − 2 × 24);
                    # a container query shows the one that fits, so track text never scales below its 11px caption size
LBL = {'telita': 'TELIȚA', 'hirbovat': 'HÎRBOVĂȚ'}
RO_W = 56  # "SAPT 40" readout at 11px Mono (7 × 6,6px) with 4px sides
def months(d0, d1):
    out = []; y, m = d0.year, d0.month
    while dt.date(y, m, 1) < d1:
        out.append((dt.date(y, m, 1), f'{m:02d}.{str(y)[2:]}')); y, m = (y + 1, 1) if m == 12 else (y, m + 1)
    return out
def track(mode='default'):
    """Full CycleTrack: one lane per farm, Seria cycles as link blocks, transfer link 22–28.02.2026, azi + SAPT readout.
    Two drawings (1160 / 1000 wide) in a container-query box; the hidden one is display:none (out of tab order and a11y tree)."""
    svgs = ''.join(track_svg(mode, tw) for tw in TWS)
    if mode == 'closed':
        leg = '<span><svg class="ded" viewBox="0 0 3 12"><line x1="1.5" x2="1.5" y1="0" y2="12"/></svg>început dedus</span>'
    elif mode == 'viewer':
        leg = '<span><i class="today"></i>azi</span>'
    else:
        leg = (f'<span><i class="ip-rec ip-rec--placement"></i>populare</span><span><i class="ip-rec ip-rec--transfer"></i>transfer '
               f'<span class="ip-num">{ddate(X0D)[:5]}–{ddate(X1D)}</span></span><span><i class="today"></i>azi</span>')
    legend = f'<div class="ip-tracklegend" aria-hidden="true">{leg}</div>'
    if mode == 'closed':
        sub = f'<span class="ip-note ip-num">02.2025–02.2026 · 2 ferme · 1 ciclu</span>'
    elif mode == 'viewer':
        sub = '<span class="ip-note">doar fermele tale · Telița: nu ai acces</span>'
    else:
        sub = f'<span class="ip-note ip-num">10.2025–12.2026 · 2 ferme · 2 cicluri</span>'
    return f'''<section class="s4-block" aria-labelledby="h-track">
<div class="s4-head"><h2 class="ip-t-heading" id="h-track">Cicluri pe ferme</h2>{sub}</div>
<div class="s4-tight"><div class="s3-trackbox">{svgs}</div>{legend}</div>
</section>'''

def track_svg(mode, TW):
    if mode == 'closed':
        d0, d1 = dt.date(2025, 2, 1), dt.date(2026, 3, 1)
        farms = ['telita', 'hirbovat']
    else:
        d0, d1 = dt.date(2025, 10, 1), dt.date(2027, 1, 1)
        farms = ['hirbovat'] if mode == 'viewer' else ['telita', 'hirbovat']
    ppd = (TW - LX) / (d1 - d0).days
    def tx(iso): return round(LX + (dt.date.fromisoformat(iso) - d0).days * ppd, 1)
    y0 = 6 if mode == 'closed' else 22                          # room for the "azi" label only when azi is on the axis
    ly = {f: y0 + 48 * i for i, f in enumerate(farms)}       # lane y; block y = lane + 3, h 20
    H = ly[farms[-1]] + 26                                    # bottom of the last lane
    g = ''
    for f in farms:
        g += f'<text class="lanelbl" x="0" y="{ly[f] + 17}">{LBL[f]}</text><rect class="lane" x="{LX}" y="{ly[f]}" width="{TW - LX}" height="26"/>'
    def block(f, a_iso, b_iso, label, aria, href, open_=False, marks=''):
        xa, xb = tx(a_iso), tx(b_iso); w = round(xb - xa, 1); y = ly[f] + 3
        o = ' open' if open_ else ''
        return (f'<a href="{href}" aria-label="{aria}"><title>{aria}</title>'
                f'<rect class="ring" x="{round(xa - 4, 1)}" y="{y - 4}" width="{round(w + 8, 1)}" height="28" rx="3"/>'
                f'<rect class="cyc{o}" x="{xa}" y="{y}" width="{w}" height="20" rx="2"/>'
                f'<text class="cyctxt{o}" x="{round(xa + 8, 1)}" y="{y + 14}">{label}</text>{marks}</a>')
    desc = ''
    if mode == 'closed':
        xa = tx(cP['start_date'])
        ded = f'<line class="ded" x1="{round(xa - 3, 1)}" x2="{round(xa - 3, 1)}" y1="{ly["hirbovat"]}" y2="{ly["hirbovat"] + 26}"/>'
        aria = (f'Hîrbovăț, Seria 2025, ouat, început dedus {ddate(cP["start_date"])} · încheiat prin golire {ddate(cP["end_date"])}, '
                f'{cP["weeks"]} săptămâni, închis. Deschide ferma la acest ciclu.')
        g += f'<text class="cyctxt" x="{LX + 12}" y="{ly["telita"] + 17}">– fără date pentru Seria 2025 la Telița</text>'
        g += block('hirbovat', cP['start_date'], cP['end_date'], f'Seria 2025 · ouat · {cP["weeks"]} săpt.', aria, '#/ferme/hirbovat?ciclu=cyc-hirbovat-2025', marks=ded)
        desc = (f'Seria 2025, închisă. Telița: fără date pentru Seria 2025. Hîrbovăț: ouat, început dedus {ddate(cP["start_date"])}, '
                f'încheiat prin golire {ddate(cP["end_date"])}, {cP["weeks"]} săpt. Azi, {ddate(TODAY)}, e după sfârșitul seriei.')
    else:
        xt = tx(TODAY)
        if 'telita' in farms:
            xa = tx(cT['start_date'])
            place = f'<rect class="place" x="{round(xa - 1, 1)}" y="{ly["telita"] + 3}" width="3" height="20"/>'
            aria = (f'Telița, Seria 2026, creștere, {ddate(cT["start_date"])}–{ddate(cT["end_date"])}, {cT["weeks"]} săptămâni, închis. '
                    f'Început prin populare (găini {n(PLC["F"])}, cocoși {n(PLC["M"])}), încheiat prin transfer la Hîrbovăț '
                    f'{ddate(X0D)[:5]}–{ddate(X1D)}. Deschide ferma la acest ciclu.')
            g += block('telita', cT['start_date'], cT['end_date'], f'Seria 2026 · creștere · {cT["weeks"]} săpt.', aria, '#/ferme/telita?ciclu=cyc-telita-2026', marks=place)
        src = (f' Păsări sosite prin transfer de la Telița {ddate(X0D)[:5]}–{ddate(X1D)} (găini {n(ARR["F"])}, cocoși {n(ARR["M"])}).'
               if 'telita' in farms else '')
        aria = f'Hîrbovăț, Seria 2026, ouat, din {ddate(cH["start_date"])}, deschis, azi sapt. {WK}.{src} Deschide ferma la acest ciclu.'
        g += block('hirbovat', cH['start_date'], TODAY, f'Seria 2026 · ouat · din {ddate(cH["start_date"])}', aria, '#/ferme/hirbovat?ciclu=cyc-hirbovat-2026', open_=True)
        if 'telita' in farms:  # transfer link: the real movement days 22–28.02.2026 (D25), from Telița's block to Hîrbovăț's
            yt, yb = ly['telita'] + 23, ly['hirbovat'] + 3; ym = (yt + yb) / 2
            xa, xb = tx(X0D), tx(X1D)
            g += (f'<g aria-hidden="true"><line class="link" x1="{xa}" y1="{yt}" x2="{xa}" y2="{yb}"/><line class="link" x1="{xb}" y1="{yt}" x2="{xb}" y2="{yb}"/>'
                  f'<line class="link" x1="{xa}" y1="{ym}" x2="{xb}" y2="{ym}"/><circle class="linkdot" cx="{xa}" cy="{ym}" r="3"/><circle class="linkdot" cx="{xb}" cy="{ym}" r="3"/></g>')
        top = ly[farms[0]] - 8
        yh = ly['hirbovat'] + 5
        g += (f'<line class="today" x1="{xt}" x2="{xt}" y1="{top}" y2="{H + 4}"/><text class="todaylbl" x="{xt}" y="{top - 3}" text-anchor="middle">azi {ddate(TODAY)[:5]}</text>'
              f'<rect class="ro" x="{round(xt + 4, 1)}" y="{yh}" width="{RO_W}" height="16" rx="2"/><text class="rotxt" x="{round(xt + 4 + RO_W / 2, 1)}" y="{yh + 12}" text-anchor="middle">SAPT {WK}</text>')
        if mode == 'viewer':
            desc = f'Seria 2026, doar fermele la care ai acces. Hîrbovăț: ouat din {ddate(cH["start_date"])}, deschis, azi {ddate(TODAY)}, sapt. {WK}.'
        else:
            desc = (f'Seria 2026, 2 ferme. Telița: creștere {ddate(cT["start_date"])}–{ddate(cT["end_date"])}, {cT["weeks"]} săpt., închis, început prin populare. '
                    f'Transfer Telița → Hîrbovăț {ddate(X0D)[:5]}–{ddate(X1D)}: găini {n(ARR["F"])}, cocoși {n(ARR["M"])}. '
                    f'Hîrbovăț: ouat din {ddate(cH["start_date"])}, deschis, azi {ddate(TODAY)}, sapt. {WK}.')
    tk = ''
    for d, lbl in months(d0, d1):
        x = tx(d.isoformat()); anc = 'start' if x - LX < 8 else 'middle'
        tk += f'<line class="tick" x1="{x}" x2="{x}" y1="{H}" y2="{H + 6}"/><text class="month" x="{x}" y="{H + 20}" text-anchor="{anc}">{lbl}</text>'
    VH = H + 24
    return (f'<svg class="ip-track s3-track s3-track--{TW}" viewBox="0 0 {TW} {VH}" role="group" aria-label="Ciclurile seriei {"2025" if mode == "closed" else "2026"} pe ferme">'
            f'<desc>{desc}</desc>{g}{tk}</svg>')

def track_loading():  # same box as the live track (azi room 22 · lane 26 · gap 22 · lane 26 · axis 24, legend 16): no shift when it loads
    lane = f'<div class="s3-sklane">{sk("s3-sklbl")}{sk("s3-skbar")}</div>'
    return f'''<section class="s4-block" aria-hidden="true">
<div class="s4-head"><h2 class="ip-t-heading">Cicluri pe ferme</h2></div>
<div class="s4-tight"><div class="s3-sktrack">{lane}{lane}<div class="s3-skaxis"></div></div><div class="s3-sklegend">{sk("s3-w24")}</div></div>
</section>'''

# ---------- 3. farm summaries (FarmSummary --wide) ----------
TAG = {'telita': ('Creștere', 'Ross 308 creștere'), 'hirbovat': ('Ouat', 'Ross 308 ouat')}
def fs_head(fid, name, href, lmp, tags=True):  # tags: True = type + guide, 'type' = type only (no cycle → no guide), False = none
    nm = f'<a href="{href}">{name}</a>' if href else name
    tg = ''.join(f'<span class="ip-tag ip-tag--neutral">{t}</span>' for t in (TAG[fid][:1] if tags == 'type' else TAG[fid] if tags else ()))
    return f'<div class="ip-fsum__head"><div class="ip-fsum__id"><h3 class="ip-fsum__name" id="fs-{fid}">{nm}</h3>{tg}</div>{lmp}</div>'
def fs_telita():
    figs = (fig('weighing', f'Greutate {F} la transfer', f'{n(WT)}<small>g</small>',
                f'<span class="ip-num">ghid {n(WG)} g · {s4.sgn(WDEV)}% ·</span><span class="ip-num">SAPT 20 · cântărit {ddate(min(WDATE))}</span>',
                lamp('warn' if abs(WDEV) > 5 else 'ok', 'abatere' if abs(WDEV) > 5 else 'în normă'))
            + fig('mortality', 'Mortalitate cumulată', f'{pair(n(TMP["F"], 2), n(TMP["M"], 2))}{P}', GN)
            + fig('feed', 'Hrană totală', f'{pair(n(FS["F"]), n(FS["M"]))}<small>g</small>',
                  f'<span class="ip-num">ghid {n(FG["F"])} / {n(FG["M"])} g ·</span><span class="ip-num">{s4.sgn(FDEV["F"])}% / {s4.sgn(FDEV["M"])}% · {len(days)} zile ·</span>'
                  f'<span class="ip-num"><span class="ip-sex ip-sex--hen">♀ {n(FKG["F"])}</span> · <span class="ip-sex ip-sex--rooster">♂ {n(FKG["M"])}</span> kg</span>',
                  lamp('ok', 'în normă'))
            + fig('transfer', 'Transferate', pair(n(ARR['F']), n(ARR['M'])), f'<span>sosite la Hîrbovăț ·</span><span class="ip-num">{ddate(X0D)[:5]}–{ddate(X1D)}</span>'))
    return f'''<article class="ip-fsum ip-fsum--wide" aria-labelledby="fs-telita" aria-describedby="fs-telita-c">{fs_head('telita', 'Telița', '#/ferme/telita?ciclu=cyc-telita-2026', lamp('off', 'închis'))}
<p class="ip-fsum__cycle" id="fs-telita-c">{ddate(cT['start_date'])}–{ddate(cT['end_date'])} · {cT['weeks']} săpt.</p>
<div class="ip-fsum__alive"><span class="ip-cap">Populat · {ddate(PL['meta']['date'])}</span><div><span class="ip-pair ip-pair--lg"><span class="ip-sex ip-sex--hen">♀ <b>{n(PLC['F'])}</b></span><span class="sep">·</span><span class="ip-sex ip-sex--rooster">♂ <b>{n(PLC['M'])}</b></span></span><span class="ip-note">total {n(PLC['F'] + PLC['M'])}</span></div></div>
<dl class="ip-fsum__figs">{figs}</dl></article>'''
def fs_hirbovat():
    lay_ok = abs(s4.lay_today - s4.lay_guide) <= 3
    figs = (fig('eggs', 'Ouat azi', f'{n(s4.lay_today, 1)}{P}', f'<span class="ip-num">ghid {pct(s4.lay_guide)} · SAPT {WK}</span>', lamp('ok' if lay_ok else 'warn', 'în normă' if lay_ok else 'abatere'))
            + fig('mortality', 'Mortalitate cumulată', f'{pair(n(MH["F"], 2), n(MH["M"], 2))}{P}', GN)
            + fig('feed', 'Hrană azi', f'{pair(n(s4.g_f), n(s4.g_m))}<small>g</small>',
                  f'<span class="ip-num">ghid {s4.gf_g} / {s4.gm_g} g ·</span><span>medie fără {F} H4.2 · {M} H3.2</span>', lamp('ok', 'în normă'))
            + fig('delivery', 'Zile stoc', f'{pair(n(B2["days_left"], 1), n(CO["days_left"], 1))}<small>zile</small>',
                  f'<span>Cocoși până pe <span class="ip-num">{ddate(CO["run_out_date"])}</span> · sub 5 zile</span>', lamp('warn', 'abatere')))
    return f'''<article class="ip-fsum ip-fsum--wide" aria-labelledby="fs-hirbovat" aria-describedby="fs-hirbovat-c">{fs_head('hirbovat', 'Hîrbovăț', '#/ferme/hirbovat?ciclu=cyc-hirbovat-2026', lamp('ok', 'deschis'))}
<p class="ip-fsum__cycle" id="fs-hirbovat-c">ciclu din {ddate(cH['start_date'])} · SAPT {WK}</p>
<div class="ip-fsum__alive"><span class="ip-cap">În viață · {ddate(TODAY)}</span><div><span class="ip-pair ip-pair--lg"><span class="ip-sex ip-sex--hen">♀ <b>{n(ALIVE['F'])}</b></span><span class="sep">·</span><span class="ip-sex ip-sex--rooster">♂ <b>{n(ALIVE['M'])}</b></span></span><span class="ip-note">total {n(ALIVE['all'])}</span></div></div>
<dl class="ip-fsum__figs">{figs}</dl></article>'''
def fs_empty(fid, name, href, word, lines, step, tags=True):
    body = ''.join(f'<p class="s3-m0">{l}</p>' for l in lines)
    return f'''<article class="ip-fsum ip-fsum--wide" aria-labelledby="fs-{fid}" aria-describedby="fs-{fid}-c">{fs_head(fid, name, href, lamp('off', word), tags)}
<div class="ip-fsum__empty" id="fs-{fid}-c">{body}{step}</div></article>'''
def fs_prev():  # Hîrbovăț · Seria 2025 (real previous flock; no flock count, no guide for these figures)
    PT = s4.PT; inc = PT['hatching_eggs'] / PT['total_eggs'] * 100; fin = s4.fin
    figs = (fig('eggs', 'Ouă în ciclu', n(PT['total_eggs']), f'<span class="ip-num">{ddate(cP["data_start_date"])}–{ddate(cP["end_date"])} · {s4.PDAYS} zile ·</span><span>{GN}</span>')
            + fig('eggs', 'Incubabile', f'{n(inc, 1)}{P}', f'<span class="ip-num">{n(PT["hatching_eggs"])} din {n(PT["total_eggs"])} ouă ·</span><span>{GN}</span>')
            + fig('mortality', 'Mortalitate cumulată', pair(n(s4.PMORT['F']), n(s4.PMORT['M'])),
                  f'<span>% – {s4.NOFLOCK} ·</span><span>fără golirea din <span class="ip-num">{ddate(s4.PDEP)}</span></span>')
            + fig('eggs', f'Masă medie ou · SAPT {s4.PFIN}', f'{n(fin["avg_egg_weight_g"], 1)}<small>g</small>', GN))
    return f'''<article class="ip-fsum ip-fsum--wide" aria-labelledby="fs-hirbovat" aria-describedby="fs-hirbovat-c">{fs_head('hirbovat', 'Hîrbovăț', '#/ferme/hirbovat?ciclu=cyc-hirbovat-2025', lamp('off', 'închis'))}
<p class="ip-fsum__cycle" id="fs-hirbovat-c">{ddate(cP['start_date'])}–{ddate(cP['end_date'])} · {cP['weeks']} săpt. · început dedus</p>
<div class="ip-fsum__alive"><span class="ip-cap">Populat · început dedus {ddate(cP['start_date'])}</span><div><span class="ip-pair ip-pair--lg"><b>–</b></span><span class="ip-note">{s4.NOFLOCK}</span></div></div>
<dl class="ip-fsum__figs">{figs}</dl></article>'''
NOACC = ['Nu ai acces la această fermă.', 'Cere acces proprietarului.']
def summaries(mode='default'):
    if mode == 'closed':
        tel = fs_empty('telita', 'Telița', '#/ferme/telita', 'fără date', ['Registrul nu are date pentru Seria 2025 la Telița.'],
                       '<a class="ip-link" href="#/serii/seria-2026">Vezi Seria 2026</a>', tags='type')
        cards, note = tel + fs_prev(), '2 ferme · Seria 2025'
    elif mode == 'viewer':
        cards, note = fs_empty('telita', 'Telița', '', 'fără acces', NOACC, '', tags=False) + fs_hirbovat(), 'fermele tale: Hîrbovăț'
    else:
        cards, note = fs_telita() + fs_hirbovat(), f'2 ferme · {ddate(TODAY)}'
    return f'''<section class="s4-block" aria-labelledby="h-farms">
<div class="s4-head"><h2 class="ip-t-heading" id="h-farms">Ferme</h2><span class="ip-note{" ip-num" if any(c.isdigit() for c in note) else ""}">{note}</span></div>
<div class="ip-fsums ip-fsums--stretch">{cards}</div>
</section>'''
def summaries_loading():
    card = (f'<div class="ip-fsum ip-fsum--wide s3-skcard"><div class="ip-fsum__head">{sk("s3-w30 s3-h16")}{sk("s3-w14")}</div>{sk("s3-w40")}'
            f'{sk("s3-w50 s3-h20")}' + ''.join(f'<div class="s3-skfig"><div class="s3-skrow">{sk("s3-w40")}{sk("s3-w20")}</div>{sk("s3-w60")}</div>' for _ in range(4)) + '</div>')
    return f'''<section class="s4-block" aria-hidden="true">
<div class="s4-head"><h2 class="ip-t-heading">Ferme</h2></div>
<div class="ip-fsums ip-fsums--stretch">{card}{card}</div>
</section>'''

# ---------- 4. tabs + Păsări aggregated per farm ----------
TABS = [('pasari', 'Păsări'), ('hrana', 'Hrană'), ('greutate', 'Greutate'), ('productie', 'Producție')]
def tabs(series='2026'):
    return f'<div class="ip-tabs" role="tablist" aria-label="Vederi Seria {series}">' + ''.join(
        f'<a class="ip-tab{" is-selected" if i == 0 else ""}" role="tab" id="tab-{k}" aria-controls="tp-serie" aria-selected="{"true" if i == 0 else "false"}" tabindex="{0 if i == 0 else -1}" href="#{k}">{t}</a>'
        for i, (k, t) in enumerate(TABS)) + '</div>'
COLS = [('', 'Început'), ('mortality', 'Mortalitate'), ('', 'Mort.\u00a0%'), ('cull', 'Sacrificare'), ('sexerr', 'Eroare sexare'),
        ('transfer', 'Transfer'), ('sale', 'Vânzare'), ('', 'Diferență registru'), ('', f'În\u00a0viață · {ddate(TODAY)[:5]}')]
COLS_CLOSED = COLS[:-1] + [('', f'La închidere · <span class="s3-nw">{ddate(cP["end_date"])}</span>')]  # a closed series counts at its end, not today
def th(k, w):  # S3P: caption in Sans (DataTable: caption headers); may wrap at 1280, the icon stays with the first word
    i = hico(k) if k else ''
    first, _, rest = w.partition(' ')
    return f'<th scope="col" class="num"><span class="s3-nw">{i}{first}</span>{(" " + rest) if rest else ""}</th>'
def cell(v, sx, sign=False, unit=''):
    if v is None: return '<td class="num s3-dash">–</td>'
    if isinstance(v, str): return f'<td class="num s3-dash">{v}</td>'
    txt = ('−' + n(v)) if (sign and v) else (n(v, 2) + unit if unit else n(v))
    cls = 'hen' if sx == 'F' else 'rooster'
    return f'<td class="num"><span class="ip-sex ip-sex--{cls}"><b>{txt}</b></span></td>'
def diff(v, sx):  # D63: register difference, info lamp + word, figure in its sex colour
    cls = 'hen' if sx == 'F' else 'rooster'
    return f'<td class="num"><span class="s3-diff">{lamp("info", "informație")}<span class="ip-sex ip-sex--{cls}"><b>−{n(-v)}</b></span></span></td>'
def sexrow(sx, vals):
    word = 'găini' if sx == 'F' else 'cocoși'
    return f'<tr><th scope="row" class="ip-table__child">{sg(sx)} {word}</th>{"".join(vals)}</tr>'
def group(fid, name, href, line):
    nm = f'<a class="s3-glink" href="{href}">{name}</a>' if href else f'<span class="s3-gname">{name}</span>'
    return f'<tr class="is-group"><th scope="rowgroup" colspan="{len(COLS) + 1}" class="s3-group">{nm}<span class="ip-note">{line}</span></th></tr>'
def pasari(mode='default'):
    head = '<thead><tr><th scope="col">Fermă · sex</th>' + ''.join(th(k, w) for k, w in (COLS_CLOSED if mode == 'closed' else COLS)) + '</tr></thead>'
    tel_line = (f'{ico("placement")}creștere · populată <span class="ip-num">{ddate(PL["meta"]["date"])}</span> · închis '
                f'<span class="ip-num">{ddate(cT["end_date"])}</span>')
    hir_line = (f'{ico("transfer")}ouat · sosite de la Telița <span class="ip-num">{ddate(X0D)[:5]}–{ddate(X1D)}</span> · deschis, '
                f'<span class="ip-num">SAPT {WK}</span>')
    notes = []
    if mode == 'closed':
        body = (f'<tbody>{group("telita", "Telița", "", "– fără date pentru Seria 2025 la Telița")}</tbody>'
                f'<tbody>{group("hirbovat", "Hîrbovăț", "#/ferme/hirbovat?ciclu=cyc-hirbovat-2025", f"ouat · început dedus <span class=\"ip-num\">{ddate(cP['start_date'])}</span> · golit <span class=\"ip-num\">{ddate(cP['end_date'])}</span>")}'
                + ''.join(sexrow(sx, [cell(None, sx), cell(s4.PMORT[sx], sx, sign=True), cell(None, sx), cell(None, sx), cell(None, sx), cell(None, sx), cell(None, sx), cell(None, sx), cell(0, sx)]) for sx in 'FM')
                + '</tbody>')
        notes = [f'– la Hîrbovăț: registrul ciclului anterior are doar ouă și mortalitate pe zi, fără efectiv, sacrificări, erori de sexare, transferuri sau vânzări.',
                 f'Mortalitatea nu include golirea halelor din <span class="ip-num">{ddate(s4.PDEP)}</span>; golirea încheie ciclul, nu e vânzare.']
        cap = 'Păsări în Seria 2025, pe ferme'
    else:
        tel = ''
        if mode == 'viewer':
            tel = f'<tbody>{group("telita", "Telița", "", "nu ai acces la această fermă")}</tbody>'
        else:
            tel = (f'<tbody>{group("telita", "Telița", "#/ferme/telita?ciclu=cyc-telita-2026", tel_line)}'
                   + ''.join(sexrow(sx, [cell(PLC[sx], sx), cell(TM[sx], sx, sign=True), cell(TMP[sx], sx, unit='%'), cell(None, sx),
                                         cell(TSE[sx], sx, sign=True), cell(ARR[sx], sx, sign=True), cell(0, sx), diff(TDIFF[sx], sx), cell(0, sx)]) for sx in 'FM')
                   + '</tbody>')
            notes = [f'Început: populare la Telița, sosire prin transfer la Hîrbovăț. Sacrificare – la Telița: registrul nu le ține separat.',
                     'Diferență registru la Telița: diferență din registrul Seria 2026, de verificat la import; nu e o corecție înregistrată.',
                     f'Nicio vânzare în Seria 2026 până pe <span class="ip-num">{ddate(TODAY)}</span>.']
        if mode == 'viewer':  # the arrival is Hîrbovăț's own record; the source farm isn't named for a viewer without Telița
            hir_line = hir_line.replace('sosite de la Telița', 'sosite prin transfer')
        hir = (f'<tbody>{group("hirbovat", "Hîrbovăț", "#/ferme/hirbovat?ciclu=cyc-hirbovat-2026", hir_line)}'
               + ''.join(sexrow(sx, [cell(ARR[sx], sx), cell(HR[sx]['mortality'], sx, sign=True), cell(MH[sx], sx, unit='%'), cell(HR[sx]['culls'], sx, sign=True),
                                     cell(HR[sx]['sexing_errors'], sx, sign=True), cell(0, sx), cell(0, sx), cell(0, sx), cell(ALIVE[sx], sx)]) for sx in 'FM')
               + '</tbody>')
        body = tel + hir
        if mode == 'viewer':
            notes = ['Început: sosire prin transfer la Hîrbovăț.', f'Nicio vânzare în Seria 2026 până pe <span class="ip-num">{ddate(TODAY)}</span>.']
        cap = 'Păsări în Seria 2026, pe ferme'
    note = ''.join(f'<p class="ip-note s3-m0">{x}</p>' for x in notes)
    return f'''<div class="s3-panel" id="tp-serie" role="tabpanel" aria-labelledby="tab-pasari">
<div class="ip-tablescroll ip-tablescroll--framed" role="region" tabindex="0" aria-label="{cap}; mortalitate % din păsările de la începutul ciclului fermei"><table class="ip-table ip-table--rowheads ip-table--wraphead s3-table"><caption class="ip-sr">{cap}; mortalitate % din păsările de la începutul ciclului fermei</caption>{head}{body}</table></div>
<div class="s3-notes">{note}</div>
</div>'''
def pasari_loading():
    row = '<tr>' + f'<td>{sk("s3-w60")}</td>' + ''.join(f'<td class="num">{sk("s3-w60 s3-right")}</td>' for _ in COLS) + '</tr>'
    head = '<thead><tr><th scope="col">Fermă · sex</th>' + ''.join(th(k, w) for k, w in COLS) + '</tr></thead>'
    grp = f'<tr class="is-group"><td colspan="{len(COLS) + 1}">{sk("s3-w14")}</td></tr>'
    return f'''<div class="s3-panel" id="tp-serie" role="tabpanel" aria-labelledby="tab-pasari" aria-busy="true">
<p class="ip-sr" role="status">Se încarcă Seria 2026: ciclurile, fermele și păsările.</p>
<div class="ip-tablescroll ip-tablescroll--framed"><table class="ip-table ip-table--rowheads ip-table--wraphead s3-table" aria-hidden="true">{head}<tbody>{grp}{row}{row}{grp}{row}{row}</tbody></table></div>
</div>'''

# ---------- page ----------
def page(mode='default'):
    series = '2025' if mode == 'closed' else '2026'
    if mode == 'loading':
        mid = track_loading() + summaries_loading(); panel = pasari_loading()
    else:
        mid = track(mode) + summaries(mode); panel = pasari(mode)
    main = f'''<main class="s4-main">
{header(mode)}
{mid}
<section class="s4-block" aria-label="Date pe ferme">
{tabs(series)}
{panel}
</section>
</main>'''
    sb = s4.sidebar('Observator', viewer=True) if mode == 'viewer' else s4.sidebar(series=series)
    return f'<div class="s4-shell">{sb}{main}</div>'

CSS = s4.CSS + '''
.s3-sep{margin-left:var(--space-2);padding-left:var(--space-3);border-left:var(--ip-hair) solid var(--line)}
.s3-m0{margin:0}
.s3-panel{display:flex;flex-direction:column;gap:var(--space-2);padding-top:var(--space-1)}
.s3-group .s4-ico{margin-right:var(--space-1)} /* group line: icon space-1 + the line's gap, as before FIN14B */
.s3-panel{container-type:inline-size}
@container (max-width: 1159.98px){.s3-table .num{padding-left:var(--space-2);padding-right:var(--space-2)}} /* 1280: 8px sides on figure columns, no sideways scroll */
.s3-group > .ip-note{display:inline-flex;align-items:center;gap:var(--space-1);margin-left:var(--space-3);font-weight:400}
.s3-glink{display:inline-flex;align-items:center;min-height:var(--space-6);color:var(--ink);text-decoration:underline;text-decoration-color:var(--line-strong);text-underline-offset:calc(3 * var(--ip-hair))}
.s3-glink:hover{text-decoration-color:var(--ink)}
.s3-gname{display:inline-flex;align-items:center;min-height:var(--space-6)}
.s3-dash{color:var(--ink-muted)}
.s3-nw{white-space:nowrap}
.s3-diff{display:inline-flex;align-items:center;gap:var(--space-2)}
.s3-notes{display:flex;flex-direction:column;gap:var(--space-1)}
.s3-trackbox{container-type:inline-size}
.s3-track--1000{display:none}
@container (max-width: 1159.98px){.s3-track--1160{display:none}.s3-track--1000{display:block}}
.s3-skmeta{display:flex;gap:var(--space-3);align-items:center;min-height:var(--space-5)}
.s3-skmeta .ip-skel{display:block}
.s3-sklane{display:grid;grid-template-columns:calc(var(--space-12) + var(--space-8) + var(--space-2)) minmax(0,1fr);align-items:center}
.s3-sktrack{padding-top:calc(var(--space-5) + 2 * var(--ip-hair))}
.s3-sklane + .s3-sklane{margin-top:calc(var(--space-5) + 2 * var(--ip-hair))}
.s3-sklbl{width:60%;display:block}
.s3-skbar{height:calc(var(--space-6) + var(--ip-hair) + var(--ip-hair));display:block}
.s3-skaxis{height:var(--space-6)} /* the month axis room, left empty */
.s3-sklegend .ip-skel{display:block}
.s3-sklegend{height:var(--space-4);display:flex;align-items:center}
.s3-skcard{gap:var(--space-3)}
.s3-skcard .ip-skel,.s3-panel .ip-skel{display:block}
.s3-skfig{display:flex;flex-direction:column;gap:var(--space-2);padding:var(--space-2) 0;border-top:var(--ip-hair) solid var(--rule)}
.s3-skrow{display:flex;justify-content:space-between}
.s3-right{margin-left:auto}
.s3-h16{height:var(--space-4)}.s3-h20{height:var(--space-5)}
.s3-w8{width:calc(var(--space-12) + var(--space-6))}.s3-w14{width:calc(var(--space-12) * 2)}.s3-w24{width:calc(var(--space-12) * 4)}
.s3-w20{width:20%}.s3-w30{width:30%}.s3-w40{width:40%}.s3-w50{width:50%}.s3-w60{width:60%}.s3-w100{width:100%}'''

B = 'S3-seria-2026'
BOARDS = [
 (B, 'S3 · Seria 2026', 'default'),
 (B + '--se-incarca', 'S3 · Seria 2026 · se încarcă', 'loading'),
 (B + '--seria-inchisa', 'S3 · Seria 2026 · seria închisă (Seria 2025)', 'closed'),
 (B + '--fara-acces', 'S3 · Seria 2026 · fără acces (Observator, doar Hîrbovăț)', 'viewer'),
]
if __name__ == '__main__':
    for stem, title, mode in BOARDS:
        s4.write(stem, title, page(mode), H=HEIGHTS.get(stem, 1400), css=CSS)
    if CANVAS:  # S3 row (was s3_canvas.py): entries keep x/y, a new row goes below the lowest (canvas_rows.py); 80 px gaps
        c = CR.load(ROOT); CR.place(c, B, [(s, 1440, HEIGHTS.get(s, 1400), t) for s, t, _ in BOARDS])
        CR.save(ROOT, c)
