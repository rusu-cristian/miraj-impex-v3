"""S18 · Serii (+ Serie nouă) (spec B3, screens.md S18; nav map: Setări › Serii, owner). Desktop-led like S17 (a Setări
page): desktops first, then the phones. RF10: every figure, date, code and week from realdata (D151–D159; today 18.11.2026).
Content (realdata: cycles, placement, farms-sections, counts-latest, transfers; arrival days through s3_build):
- Seria 2026 (year 2026, open, placed 07.10.2025 at age 0, D151): Telița creștere 07.10.2025–28.02.2026 (SAPT 1–21,
  closed, started by the placement ♀ 28 080 · ♂ 3 510), the move to Hîrbovăț 22–28.02.2026 (♀ 26 883 · ♂ 3 024 at arrival,
  D28), Hîrbovăț ouat from 22.02.2026 (SAPT 20–59, open, emptying by sale since 12.11, D155; alive ♀ 12 561 · ♂ 794).
  Seria 2025 (year 2025, closed, illustrative name): Hîrbovăț ouat, inferred start 07.03.2025 – emptied 28.01.2026
  (SAPT 19–65 from the inferred placement 01.11.2024, D151), no flock count, no Telița data. Each series row names its SAPT
  span (open: + today's age, spec B3 round 2); each cycle row its SAPT span in the Săpt. column.
- Serie nouă (B3 flow): Nume, An, Ferma de creștere (growth farms only: Telița), Data populării (DateField, SAPT 1 of
  the new cycle at the default age 0), Vârstă la populare (zile) (S18N), Populare pe secțiuni (SexCountRow `ip-sexrows--add`, SYS07, as its preview),
  EffectPreview (the new cycle, CycleTrack mark vocabulary), Creează seria. S18N: the age field = CountInput + suffix
  "zile", 0–140, spec hint + live line "La populare: vârstă 0 zile · SAPT 1."; its SAPT also drives the head and EffectPreview.
- Refusal (B3 rule) "Telița are încă … păsări din Seria 2026": Telița is empty in counts-latest (closed 28.02.2026), so
  the board is a D39 state example, labelled in its title: as if the move had not been entered, Telița would hold the
  birds that arrived at Hîrbovăț (♀ 26 883 · ♂ 3 024 = 29 907, D28, RF00). Spec sample "Seria 2027 · 2027".
Boards (S18 row; x step w + 80):
  S18-serii                       desktop, owner, the list: PageHeader (Setări › Serii, "Adaugă serie"), CycleTrack full
                                  with both series, DataTable of series (group rows) and their cycles, notes
  S18-serii--serie-noua           desktop, the list reflowed beside the non-modal SidePanel "Serie nouă" (D73), nothing typed,
                                  Telița and Azi preset, focus on Nume
  S18-serii--telefon              phone, owner, the list (track at 358 px, framed table scrolling sideways), tab bar Mai mult
  S18-serii--telefon-serie-noua   phone sheet (`ip-sidepanel--sheet`, §39), replay of 07.10.2025 (as D73/D83): Seria 2026 ·
                                  2026 · Telița · 07.10.2025 · the real placement rows, Creează seria live
  S18-serii--telefon-refuzat      phone sheet, today, D39: the fault Banner with the B3 sentence and its three ways out,
                                  Creează seria disabled
Reuses by import: s17_build (Setări sidebar subs, phone header and FilterBar-free phone shell; it imports s14s_build →
s14_build, s2_build, s4s_build and rewrites the S17/S14S/S2/S7 boards byte-identical), s3_build (series figures; no
writes on import), sstate_css (SidePanel beside the page, sheet head/body, EffectPreview, Banner). No new system class.
python3 -I s18_build.py <repo_root> [s18_heights.json] [--canvas]
Writes design/canvas/project/S18-serii[--<state>].dc.html, /tmp/s18/<stem>.html, /tmp/s18/boards.json; --canvas adds the
5 boards to canvas.json (new row, nothing else changes)."""
import sys, re, json, pathlib, importlib.util, datetime as dt

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve(); HERE = pathlib.Path(__file__).resolve().parent
HP = pathlib.Path(ARGS[1]) if len(ARGS) > 1 else HERE / 's18_heights.json'
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}
sys.path.insert(0, str(HERE))
import vkey as V   # VO02: the voice key on every signed-in board (spec §8, D199; vkey.add in the writer)
import skel as SKL
import canvas_rows as CR   # FV02: board_widths.json (wide boards, D167)
def load(name, argv):
    sys.argv = argv
    sp = importlib.util.spec_from_file_location(name, HERE / f'{name}.py'); m = importlib.util.module_from_spec(sp); sp.loader.exec_module(m)
    return m
Q = load('s17_build', ['s17_build.py', str(ROOT)])      # Setări shell; rewrites S17 (+ S14S/S2/S7) byte-identical
T3 = load('s3_build', ['s3_build.py', str(ROOT)])       # series figures (asserted there); writes nothing on import
import sstate_css as SC
Z, X, A, s4 = Q.Z, Q.X, Q.A, Q.s4
B7 = A.B                                                # s7_build (DateField, CSS)
n, icon, lamp, NB = s4.n, s4.icon, s4.lamp, s4.NB
def dd(x): return (x if isinstance(x, dt.date) else dt.date.fromisoformat(x)).strftime('%d.%m.%Y')
def num(v): return f'<span class="ip-num">{v}</span>'
def pair(f, m): return s4.pair(f if isinstance(f, str) else n(f), m if isinstance(m, str) else n(m))
PLACE = '<path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><path d="m10 17 5-5-5-5"/><path d="M15 12H3"/>'  # RecordTile placement (log-in)
s4.IC.update({'placement': PLACE, 'x': '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',
              'alert': '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>'})
s4.REC.update({'placement': 'placement'})

# ================================================================= data (RF10: realdata only, D151–D159)
import realdata as R
FS = R.FS
TODAY = R.TODAY                                         # 18.11.2026 (D152)
SER, CYC = R.SERIES, R.CYCLES
cT, cH, cP = R.CTEL, R.C26, R.C25
assert (SER['seria-2026']['year'], SER['seria-2026']['status'], SER['seria-2025']['year'], SER['seria-2025']['status']) == (2026, 'active', 2025, 'closed')
assert (cT, cH, cP) == (T3.cT, T3.cH, T3.cP)
GUIDE = {g: v['name'] for g, v in R.GUIDES.items()}
FARM = R.FARMS
TYPE_RO = {'growth': 'Creștere', 'laying': 'Ouat'}
GROWTH = [f for f in FS['farms'] if f['type'] == 'growth']; assert [f['name'] for f in GROWTH] == ['Telița']
TEL_SECS = [s['code'] for s in FS['sections'] if s['farm_id'] == 'telita' and s['active']]
assert TEL_SECS == R.SECTIONS['telita']
PL = R.PLACEMENT_ROWS
PLC = PL['totals']; PROWS = [(r['section'], r['sex'], r['qty']) for r in PL['rows']]
assert PL['meta']['date'] == cT['start_date'] == R.PLACEMENT.isoformat() and cT['started_by'] == 'auto'
assert (sum(q for _, x, q in PROWS if x == 'F'), sum(q for _, x, q in PROWS if x == 'M')) == (PLC['F'], PLC['M'])
ARR, X0D, X1D = T3.ARR, T3.X0D, T3.X1D
TEL, HIR = R.COUNTS_TELITA, R.COUNTS
assert TEL['totals']['all'] == 0 and TEL['cycle_end'] == cT['end_date']   # Telița is empty today (real)
assert (HIR['cycle']['start_counts']['F'], HIR['cycle']['start_counts']['M']) == (ARR['F'], ARR['M']) == (R.ARRIVED['F'], R.ARRIVED['M'])
WK, AGE = R.WK, R.age_days(TODAY)                       # SAPT 59, 407 days (D151)
ALIVE = {x: R.alive_farm(x) for x in 'FM'}; assert ALIVE == {x: HIR['totals'][x] for x in 'FM'}   # ♀ 12 561 · ♂ 794
# cycle and series SAPT spans by age week (D151): a cycle's own span; an open one runs to today's week
for c in (cT, cH, cP):
    assert c['sapt_start'] == R.sapt(c['start_date'], c['series_id']) and (c['end_date'] is None or c['sapt_end'] == R.sapt(c['end_date'], c['series_id']))
assert cH['end_date'] is None and cH['sapt_today'] == WK
def sapt_span(a, b): return f'SAPT {a}' if a == b else f'SAPT {a}–{b}'
def cyc_span(c, week=None):
    """the cycle's SAPT span; open (end_date None) → its start to `week` (default today's)."""
    return sapt_span(c['sapt_start'], c['sapt_end'] if c['end_date'] else (week or WK))
def ser_span(sid, day=None):
    """a series' SAPT span: from its placement week (Seria 2025: its first cycle's, the placement being inferred) to the
    last closed week or the day's week."""
    cs = [c for c in (cT, cH, cP) if c['series_id'] == sid]
    a = min(c['sapt_start'] for c in cs) if SER[sid].get('placement_inferred') else R.sapt(SER[sid]['placement_date'], sid)
    if day is not None: return sapt_span(a, R.sapt(day, sid))
    return sapt_span(a, WK if any(c['end_date'] is None for c in cs) else max(c['sapt_end'] for c in cs))
AGE0 = SER['seria-2026']['age_at_placement_days']       # default 0 (D151: day-old chicks); S18N: the "Vârstă la populare (zile)" field
AGE_MAX = 140                                            # spec B3 round 2 (KIT02): 0–140 whole days (the age Seria 2026 left the growth farm)
def age_wk(age): return age // 7 + 1                     # D151: the placement day's week (13 days → SAPT 2, 14 → SAPT 3)
assert (age_wk(0), age_wk(13), age_wk(14), age_wk(21)) == (1, 2, 3, 4)
NEW_WK = age_wk(AGE0)                                    # a new series placed at age 0 starts in SAPT 1
def zile(k):
    """'0 zile', '14 zile', '21 de zile', '160 de zile' (Romanian: 'de' after 20–99 in the last two digits, and 100, 200…)."""
    r = k % 100
    return f'{num(k)} {"de " if r >= 20 or (k >= 100 and r == 0) else ""}zile'
# D39 refusal: Telița as if the move had not been entered = the birds counted at arrival (D28, RF00)
R39 = {'F': ARR['F'], 'M': ARR['M']}; R39ALL = R39['F'] + R39['M']
REPLAY = R.PLACEMENT                                    # phone replay of the real creation: "today" = 07.10.2025
assert R.sapt(REPLAY) == NEW_WK and R.age_days(REPLAY) == AGE0
NEWNAME, NEWYEAR = f'Seria {TODAY.year + 1}', str(TODAY.year + 1)   # spec B3 sample "Seria 2027 · 2027" (refusal board only)

# ================================================================= CycleTrack · full, both series
def months(d0, d1):
    out = []; y, m = d0.year, d0.month
    while dt.date(y, m, 1) < d1:
        out.append((dt.date(y, m, 1), f'{m:02d}.{str(y)[2:]}')); y, m = (y + 1, 1) if m == 12 else (y, m + 1)
    return out
LBL = {'telita': 'TELIȚA', 'hirbovat': 'HÎRBOVĂȚ'}
MONTH_W, MONTH_GAP = 6.6, 8   # 11px Plex Mono: 0.6em a figure; at least space-2 between two month labels
def month_axis(d0, d1, tx, LX, H, step):
    """Ticks every month, labels every <step> months (none on the axis end). S18P: when the first label (start-anchored at
    the axis start) would sit closer than space-2 to the second (560 drawing: 02.25 | 05.25), the first label is dropped;
    its tick stays."""
    out, lbls = [], []
    for d, lbl in months(d0, d1):
        x = tx(d.isoformat()); anc = 'start' if x - LX < 8 else 'middle'
        out.append(f'<line class="tick" x1="{x}" x2="{x}" y1="{H}" y2="{H + 6}"/>')
        if (d.month - 2) % step == 0 and (d.year, d.month) != (d1.year, d1.month):
            w = len(lbl) * MONTH_W; a = x if anc == 'start' else x - w / 2
            lbls.append((a, a + w, len(out))); out.append(f'<text class="month" x="{x}" y="{H + 20}" text-anchor="{anc}">{lbl}</text>')
    if len(lbls) > 1 and lbls[1][0] - lbls[0][1] < MONTH_GAP:
        out[lbls[0][2]] = ''; lbls = lbls[1:]
    assert all(b[0] - a[1] >= MONTH_GAP for a, b in zip(lbls, lbls[1:])), lbls
    return ''.join(out)
TWS = (1160, 1000, 720, 560)   # 1:1 for the content box at 1440, at 1280, and beside the SidePanel at 1440 (1160 − panel-w 440) and 1280 (1000 − 440)
# block labels per drawing width (11px track text): long where the block is wide enough, short where not
# RF10: closed blocks name their SAPT span (D151: age weeks), as the S4 farm track ("Seria 2026 · creștere · SAPT 1–21")
TXT = {
    1160: dict(tel=f'Seria 2026 · creștere · {cyc_span(cT)}', h25=f'Seria 2025 · ouat · {cyc_span(cP)}', h26=f'Seria 2026 · ouat · din {dd(cH["start_date"])}', no25='– fără date pentru Seria 2025 la Telița'),
    1000: dict(tel='Seria 2026 · creștere', h25=f'Seria 2025 · ouat · {cyc_span(cP)}', h26=f'Seria 2026 · ouat · din {dd(cH["start_date"])}', no25='– fără date pentru Seria 2025'),
    720: dict(tel='Seria 2026', h25=f'Seria 2025 · ouat · {cyc_span(cP)}', h26='Seria 2026 · ouat', no25='– fără date Seria 2025'),
    560: dict(tel='2026', h25='Seria 2025 · ouat', h26='Seria 2026 · ouat', no25='– fără date 2025'),
    358: dict(tel='2026', h25='2025 · ouat', h26='2026 · ouat', no25=''),
}
def track_svg(TW):
    phone = TW == 358
    LX = 72 if phone else 88
    d0 = dt.date(2025, 2, 1)
    ro_w = 52 if phone else 56
    d1 = dt.date(2027, 1, 1)   # the axis runs on by whole months until the SAPT readout fits right of the azi line (as D91)
    while LX + (TODAY - d0).days * (TW - LX) / (d1 - d0).days + 4 + ro_w > TW:
        d1 = dt.date(d1.year + d1.month // 12, d1.month % 12 + 1, 1)
    ppd = (TW - LX) / (d1 - d0).days
    def tx(iso): return round(LX + (dt.date.fromisoformat(iso) - d0).days * ppd, 1)
    LH, BH = 26, 20
    ly = {'telita': 22, 'hirbovat': 70}; H = ly['hirbovat'] + LH
    t = TXT[TW]
    g = ''.join(f'<text class="lanelbl" x="0" y="{ly[f] + 17}">{LBL[f]}</text><rect class="lane" x="{LX}" y="{ly[f]}" width="{TW - LX}" height="{LH}"/>' for f in ly)
    def block(f, a, b, label, aria, href, open_=False, marks=''):
        xa, xb = tx(a), tx(b); w = round(xb - xa, 1); y = ly[f] + 3; o = ' open' if open_ else ''
        hit = f'<rect x="{xa}" y="{y + BH // 2 - 22}" width="{w}" height="44" fill="none" pointer-events="all"/>' if phone else ''
        return (f'<a href="{href}" aria-label="{aria}"><title>{aria}</title>{hit}<rect class="ring" x="{round(xa - 4, 1)}" y="{y - 4}" width="{round(w + 8, 1)}" height="{BH + 8}" rx="3"/>'
                f'<rect class="cyc{o}" x="{xa}" y="{y}" width="{w}" height="{BH}" rx="2"/><text class="cyctxt{o}" x="{round(xa + 8, 1)}" y="{y + 14}">{label}</text>{marks}</a>')
    if t['no25']:
        g += f'<text class="cyctxt" x="{LX + 12}" y="{ly["telita"] + 17}">{t["no25"]}</text>'
    xa = tx(cP['start_date'])
    ded = f'<line class="ded" x1="{round(xa - 3, 1)}" x2="{round(xa - 3, 1)}" y1="{ly["hirbovat"]}" y2="{ly["hirbovat"] + LH}"/>'
    g += block('hirbovat', cP['start_date'], cP['end_date'], t['h25'],
               f'Hîrbovăț, Seria 2025, ouat, început dedus {dd(cP["start_date"])}, încheiat prin golire {dd(cP["end_date"])}, {cyc_span(cP).lower().replace("sapt", "sapt.")}, închis. Deschide ferma la acest ciclu.',
               '#/ferme/hirbovat?ciclu=cyc-hirbovat-2025', marks=ded)
    xa = tx(cT['start_date'])
    place = f'<rect class="place" x="{round(xa - 1, 1)}" y="{ly["telita"] + 3}" width="3" height="{BH}"/>'
    g += block('telita', cT['start_date'], cT['end_date'], t['tel'],
               f'Telița, Seria 2026, creștere, {R.drange(cT["start_date"], cT["end_date"])}, {cyc_span(cT).lower().replace("sapt", "sapt.")}, închis. Început prin populare '
               f'(găini {n(PLC["F"])}, cocoși {n(PLC["M"])}), încheiat prin transfer la Hîrbovăț {dd(X0D)[:5]}–{dd(X1D)}. Deschide ferma la acest ciclu.',
               '#/ferme/telita?ciclu=cyc-telita-2026', marks=place)
    g += block('hirbovat', cH['start_date'], TODAY.isoformat(), t['h26'],
               f'Hîrbovăț, Seria 2026, ouat, din {dd(cH["start_date"])}, sapt. {cH["sapt_start"]}–{WK}, deschis, azi sapt. {WK}, în viață găini {n(ALIVE["F"])}, cocoși {n(ALIVE["M"])}. Păsări sosite prin transfer de la Telița '
               f'{dd(X0D)[:5]}–{dd(X1D)} (găini {n(ARR["F"])}, cocoși {n(ARR["M"])}). Deschide ferma la acest ciclu.',
               '#/ferme/hirbovat?ciclu=cyc-hirbovat-2026', open_=True)
    yt, yb = ly['telita'] + 3 + BH, ly['hirbovat'] + 3; ym = (yt + yb) / 2; xa, xb = tx(X0D), tx(X1D)
    g += (f'<g aria-hidden="true"><line class="link" x1="{xa}" y1="{yt}" x2="{xa}" y2="{yb}"/><line class="link" x1="{xb}" y1="{yt}" x2="{xb}" y2="{yb}"/>'
          f'<line class="link" x1="{xa}" y1="{ym}" x2="{xb}" y2="{ym}"/><circle class="linkdot" cx="{xa}" cy="{ym}" r="3"/><circle class="linkdot" cx="{xb}" cy="{ym}" r="3"/></g>')
    xt = tx(TODAY.isoformat()); top = ly['telita'] - 8; yh = ly['hirbovat'] + 5
    assert xt + 4 + ro_w <= TW, (TW, xt)
    g += (f'<line class="today" x1="{xt}" x2="{xt}" y1="{top}" y2="{H + 4}"/><text class="todaylbl" x="{xt}" y="{top - 3}" text-anchor="middle">azi {dd(TODAY)[:5]}</text>'
          f'<rect class="ro" x="{round(xt + 4, 1)}" y="{yh}" width="{ro_w}" height="16" rx="2"/><text class="rotxt" x="{round(xt + 4 + ro_w / 2, 1)}" y="{yh + 12}" text-anchor="middle">SAPT {WK}</text>')
    tk = month_axis(d0, d1, tx, LX, H, 6 if phone else 3 if TW <= 720 else 2)   # every 2nd month, quarterly at 720 and 560, half-yearly on phone
    desc = (f'2 serii, 2 ferme, 3 cicluri. Seria 2025, închisă: Hîrbovăț ouat, început dedus {dd(cP["start_date"])}, încheiat prin golire {dd(cP["end_date"])}, '
            f'{cyc_span(cP).lower().replace("sapt", "sapt.")}; la Telița fără date. Seria 2026, deschisă: Telița creștere {R.drange(cT["start_date"], cT["end_date"])}, {cyc_span(cT).lower().replace("sapt", "sapt.")}, '
            f'început prin populare; transfer Telița → Hîrbovăț {dd(X0D)[:5]}–{dd(X1D)}: găini {n(ARR["F"])}, cocoși {n(ARR["M"])}; '
            f'Hîrbovăț ouat din {dd(cH["start_date"])}, deschis, azi {dd(TODAY)}, sapt. {WK}, în viață găini {n(ALIVE["F"])}, cocoși {n(ALIVE["M"])}.')
    return (f'<svg class="ip-track s18-track {"ip" if TW in (1160, 1000, 720, 560) else "s18"}-track--{TW}" viewBox="0 0 {TW} {H + 24}" role="group" aria-label="Ciclurile seriilor pe ferme">'
            f'<desc>{desc}</desc>{g}{tk}</svg>')
LEG = (f'<span><i class="ip-rec ip-rec--placement"></i>populare</span><span><i class="ip-rec ip-rec--transfer"></i>transfer '
       f'<span class="ip-num">{dd(X0D)[:5]}–{dd(X1D)}</span></span>'
       '<span><svg class="ded" viewBox="0 0 3 12"><line x1="1.5" x2="1.5" y1="0" y2="12"/></svg>început dedus</span><span><i class="today"></i>azi</span>')
def track(phone=False):
    svgs = track_svg(358) if phone else ''.join(track_svg(w) for w in TWS)
    sub = f'<span class="ip-note">{num(2)} serii · {num(2)} ferme · {num(3)} cicluri</span>'
    return (f'<section class="ip-block" aria-labelledby="h-track"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-track">Cicluri pe ferme</h2>{sub}</div>'
            f'<div class="s4-tight"><div class="ip-trackbox">{svgs}</div><div class="ip-tracklegend" aria-hidden="true">{LEG}</div></div></section>')

# ================================================================= DataTable: series (group rows) and their cycles
def ico(k):
    return f'<span class="ip-rec ip-rec--{k} s4-ico">{icon(k, 14)}</span>'
def crow(fid, cyc=None, start='', st=None, dedus=False, week=None, agecol=True):
    """one cycle row; Săpt. = the cycle's SAPT span (D151, spec B3 round 2), an open cycle's up to `week` (default today's)."""
    f = FARM[fid]
    if cyc is None:  # a farm the series has no cycle at
        return (f'<tr><th scope="row" class="ip-table__child">{f["name"]}</th>'
                f'<td colspan="{7 if agecol else 6}" class="ip-muted">– fără date pentru Seria 2025 la {f["name"]} (registrul începe cu Seria 2026)</td></tr>')
    href = f'#/ferme/{fid}?ciclu={cyc["id"]}'
    end = dd(cyc['end_date']) if cyc['end_date'] else None
    # FIN09/FY02: "start–end" unspaced. FW04 (D150): every cell on one line, so "· început dedus" moves from after the range
    # (as S3/S4/S20) to the Început column, which says how the cycle began (the range alone fits; at 1280 both on one
    # line made the table 44px wider than its box)
    per = (f'<span class="s18-cyc"><span>{num(dd(cyc["start_date"]) + "–")}</span><span>{num(end)}</span></span>' if end
           else f'din {num(dd(cyc["start_date"]))}')
    if dedus:
        start += '<span class="ip-note"> · început dedus</span>'
    wk = num(cyc_span(cyc, week))
    stt = lamp('off', 'închis') if end else lamp('ok', 'deschis')
    return (f'<tr><th scope="row" class="ip-table__child"><a class="s18-rlink" href="{href}" aria-label="{f["name"]}, ciclul {"din " + dd(cyc["start_date"]) if not end else dd(cyc["start_date"]) + "–" + end}">{f["name"]}</a></th>'
            f'<td>{TYPE_RO[f["type"]]}</td><td>{GUIDE[cyc["guide_id"]]}</td><td>{per}</td><td class="num">{wk}</td>{'<td class="num s18-agecell"></td>' if agecol else ''}<td>{start}</td><td>{stt}</td></tr>')   # S18N (D179): the age column is the series row's
START_T = f'<span class="s18-start">{ico("placement")}<span>populare</span>{pair(PLC["F"], PLC["M"])}</span>'
START_H = f'<span class="s18-start">{ico("transfer")}<span>transfer de la Telița</span>{pair(ARR["F"], ARR["M"])}</span>'
START_P = f'<span class="ip-muted">– {s4.NOFLOCK}</span>'
def gnote(sid, day=None, lead=None):
    """the series row's note (spec B3 round 2): year, how it began, its SAPT span, and for an open series today's age
    ("SAPT 1–59 · azi 407 zile"); a closed one "doar citire" and its last cycle's end. day: a replay's "today" (open)."""
    s = SER[sid]; y = f'an {num(s["year"])}'
    if lead is None:
        lead = (f'populată {num(dd(s["placement_date"]))} la Telița' if not s.get('placement_inferred')
                else f'început dedus {num(dd(min(c["start_date"] for c in (cT, cH, cP) if c["series_id"] == sid)))}')   # its one farm is the row below
    if s['status'] == 'active' or day is not None:
        d = day or TODAY
        if s.get('placement_inferred'):   # Seria 2025 (replay, open): its placement is inferred, so no age; the span only
            return f'{y} · {lead} · {num(ser_span(sid, d))}'
        return f'{y} · {lead} · {num(ser_span(sid, d))} · azi {num(R.age_days(d, sid))} zile'
    last = max(c['end_date'] for c in (cT, cH, cP) if c['series_id'] == sid)
    return f'{y} · {num(ser_span(sid))} · doar citire · ultimul ciclu încheiat {num(dd(last))}'
def agecell(sid):
    """S18N (D179, spec B3 round 2): the series' age at placement in the "Vârstă la populare" column, on its series row
    ("0 zile"; an inferred placement adds "· dedus")."""
    s = SER[sid]; d = '<span class="ip-note"> · dedus</span>' if s.get('placement_inferred') else ''
    return f'<td class="num s18-agecell">{zile(s["age_at_placement_days"])}{d}</td>'
def agefold(sid):
    """the same value for the series note (phones; desktop when the region is narrower than the table, D179 note)."""
    s = SER[sid]
    return f' · vârstă la populare {zile(s["age_at_placement_days"])}{" · dedus" if s.get("placement_inferred") else ""}'
def grow(sid, touch=False, keys=True):  # keys=False (phone): the series keys sit under the table, outside the scroll area
    s = SER[sid]
    if s['status'] == 'active':
        st, note = lamp('ok', 'deschisă'), gnote(sid)
    else:
        st, note = lamp('off', 'închisă'), gnote(sid)
    if not keys:   # phone (D179): no column room, so the age folds into the series row's note, which spans the row as before
        return (f'<tr class="is-group"><th scope="rowgroup" colspan="7" id="g-{sid}"><span class="s18-g"><span class="s18-gname">{s["name"]}</span>{st}'
                f'<span class="ip-note">{note}{agefold(sid)}</span></span></th></tr>')
    head = (f'<tr class="is-group"><th scope="rowgroup" colspan="5" id="g-{sid}"><span class="s18-g"><span class="s18-gname">{s["name"]}</span>{st}'
            f'<span class="ip-note">{note}</span></span></th>{agecell(sid)}')
    return head + f'<td colspan="2" class="s18-gact"><span class="s18-acts">{acts(sid, touch)}</span></td></tr>'
COLS = ('<thead><tr><th scope="col">Serie · fermă</th><th scope="col">Tip</th><th scope="col">Ghid</th><th scope="col">Ciclu</th>'
        '<th scope="col" class="num">Săpt.</th><th scope="col" class="num s18-agehead"><span>Vârstă la populare</span></th><th scope="col">Început</th><th scope="col">Stare</th></tr></thead>')
COLS_P = COLS.replace('<th scope="col" class="num s18-agehead"><span>Vârstă la populare</span></th>', '')   # phone: the age is in the series note (D179)
assert COLS_P != COLS
# FIN11G (§54): the desktop list's column shares at 1440 (as content-sized before) on the head the loading skeleton reuses;
# None = the column that takes the rest
DCOLS = SKL.widths(COLS, [11.0, 7.5, 13.4, 16.5, 7.8, 9.0, None, 8.1], [87.36, 52.58, 111.8, 164.38, 78.28, 105.0, 273.69, 58.56])  # + loaded min-content px (fix 1; FW04: one-line cells, Ciclu without "început dedus"; RF10: Săpt. = the cycle's SAPT span, re-measured)   # desktop (the phone tables scroll at content width)
def acts(sid, touch):
    t = ' ip-btn--touch' if touch else ''
    a = f'<a class="ip-btn ip-btn--ghost{t}" href="#/serii/{sid}">Deschide seria<span class="ip-sr"> {SER[sid]["name"]}</span></a>'
    if SER[sid]['status'] == 'active':
        a += f'<button type="button" class="ip-btn ip-btn--secondary{t}" disabled="disabled" aria-describedby="why-{sid}">Închide seria<span class="ip-sr"> {SER[sid]["name"]}</span></button>'
    return a
# RF10: the reason with today's real figures: the Hîrbovăț cycle closes when the sales take the farm to 0 (B4, D155)
WHY26 = (f'<p class="ip-note ip-m0" id="why-seria-2026">Seria 2026 se poate închide după ce se încheie ciclul de la Hîrbovăț '
         f'(deschis, sapt. {num(WK)}, încă {pair(ALIVE["F"], ALIVE["M"])} în viață).</p>')
RO_NOTE = '<p class="ip-note ip-m0">O serie închisă e doar citire.</p>'   # S18P: the cycle-start sentence cut (B4 rule, said by the EffectPreview)
def ptable():
    """Phone (fix 1): one framed scroll table per series (group row + its cycles); the series keys sit under it, outside the
    scroll area, with Închide seria's reason right under its key."""
    out = ''
    for sid, rows in (('seria-2026', crow('telita', cT, START_T, agecol=False) + crow('hirbovat', cH, START_H, agecol=False)),
                      ('seria-2025', crow('hirbovat', cP, START_P, dedus=True, agecol=False) + crow('telita', agecol=False))):
        g = grow(sid, keys=False)
        out += (f'<div class="s18-pser"><div class="ip-tablescroll ip-tablescroll--framed" tabindex="0" role="region" aria-labelledby="g-{sid}">'
                f'<table class="ip-table ip-table--rowheads s18-table"><caption class="ip-sr">{SER[sid]["name"]}: ciclurile pe ferme</caption>{COLS_P}<tbody>{g}{rows}</tbody></table></div>'
                f'<div class="s18-pacts">{acts(sid, True)}</div>{WHY26 if sid == "seria-2026" else ""}</div>')
    return (f'<section class="ip-block" aria-labelledby="h-ser"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-ser">Serii</h2>'
            f'<span class="ip-note">{num(2)} serii · cele mai noi întâi</span></div>{out}<div class="s18-notes">{RO_NOTE}</div></section>')
def table(touch=False):
    if touch: return ptable()
    body = (f'<tbody>{grow("seria-2026", touch)}{crow("telita", cT, START_T)}{crow("hirbovat", cH, START_H)}</tbody>'
            f'<tbody>{grow("seria-2025", touch)}{crow("hirbovat", cP, START_P, dedus=True)}{crow("telita")}</tbody>')
    notes = [WHY26, RO_NOTE]
    return (f'<section class="ip-block" aria-labelledby="h-ser"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-ser">Serii</h2>'
            f'<span class="ip-note">{num(2)} serii · cele mai noi întâi</span></div>'
            f'<div class="ip-tablescroll ip-tablescroll--framed" tabindex="0" role="region" aria-labelledby="h-ser">'
            f'<table class="ip-table ip-table--rowheads s18-table"><caption class="ip-sr">Seriile și ciclurile lor pe ferme</caption>{DCOLS}{body}</table></div>'
            f'<div class="s18-notes">{"".join(notes)}</div></section>')

# ================================================================= page (desktop) + header + sidebar
CRUMBS = '<nav class="ip-crumbs" aria-label="Cale"><a href="#setari">Setări</a><span aria-hidden="true">›</span><span aria-current="page">Serii</span></nav>'
META = f'<div class="ip-row ip-pagehead__meta"><span class="ip-note">Seriile anuale, populările și ciclurile lor pe ferme{NB}·{NB}Hîrbovăț, Telița</span></div>'
def newkey(pressed=False, touch=False):
    t = ' ip-btn--touch' if touch else ''
    p = ' is-pressed" aria-expanded="true" aria-controls="sn' if pressed else ''   # the panel exists only while open
    return f'<button type="button" class="ip-btn ip-btn--secondary{t}{p}">{icon("plus", 16)}Adaugă serie</button>'   # FIN07 (D127): keys are verbs (craft §6); the panel keeps its title
def header(pressed=False):
    return (f'<header class="ip-pagehead">\n<div>\n{CRUMBS}\n<h1>Serii</h1>\n{META}\n</div>\n'
            f'<div class="ip-row">{newkey(pressed)}</div>\n</header>')
def sidebar_owner():
    sb = s4.sidebar('Proprietar').replace('class="ip-nav is-current" href="#serie" aria-current="page"', 'class="ip-nav" href="#serie"')
    par = f'<a class="ip-nav" href="#setari">{icon("settings")}Setări</a>'; assert sb.count(par) == 1
    subs = ''.join(f'<a class="ip-nav ip-nav--sub{" is-current" if h == "#serii" else ""}" href="{h}"'
                   f'{" aria-current=\"page\"" if h == "#serii" else ""}>{t}</a>' for h, t in Q.OWNER_SUBS)
    sb = sb.replace(par, par + subs, 1)
    assert sb.count('is-current') == 1 and 'href="#serii" aria-current="page"' in sb
    return sb
def page(pressed=False):
    return f'<div class="ip-shell">{sidebar_owner()}<main class="ip-shell__main">\n{header(pressed)}\n{track()}\n{table()}\n</main></div>'

# ================================================================= Serie nouă form (SidePanel desktop, sheet on phone)
TILE = f'<span class="ip-tile ip-rec ip-rec--placement" aria-hidden="true">{icon("placement")}</span>'
def p_head(ctx, touch):
    t = ' ip-iconbtn--touch' if touch else ''
    return (f'<div class="ip-sidepanel__head"><div class="ip-row ip-sidepanel__title">{TILE}<div><h2 class="ip-t-heading ip-m0" id="sn-t">Serie nouă</h2>'
            f'<p class="ip-note ip-num ip-m0">{ctx}</p></div></div><button type="button" class="ip-iconbtn{t}" aria-label="Închide Serie nouă">{icon("x")}</button></div>')
def tfield(i, label, val='', hint='', touch=True, mono=False, focus=False, numeric=False, cls=''):
    t = ' ip-input--touch' if touch else ''
    v = f' value="{val}"' if val else ''
    st = ' is-focus' if focus else ''
    af = ' autofocus="autofocus"' if focus else ''
    im = ' inputmode="numeric"' if numeric else ''
    h = f'<span class="ip-hint" id="{i}-h">{hint}</span>' if hint else ''
    d = f' aria-describedby="{i}-h"' if hint else ''
    return (f'<div class="ip-field{cls}"><label class="ip-label" for="{i}">{label}</label><span class="ip-input{t}{" ip-num" if mono else ""}{st}">'
            f'<input id="{i}" type="text"{im} autocomplete="off"{v}{d}{af}/></span>{h}</div>')
def farm_field(day, touch=True, refused=False):
    """FIN10: the growth farms' options carry state (empty since / sections / birds alive), so Ferma de creștere is the
    section 53 Select: phone opens the BottomSheet (aria-haspopup dialog, no list in the page), desktop its listbox."""
    t = ' ip-input--touch' if touch else ''
    if refused:
        hint = f'<span class="ip-hint" id="sn-farm-h">Doar fermele de creștere · ghid {GUIDE["g-growth"]}</span>'
        note, st, inv, desc = pair(R39['F'], R39['M']), ' is-error', ' aria-invalid="true"', 'sn-ref sn-farm-h'
    else:
        goala = day > dt.date.fromisoformat(cT['end_date'])
        state = (f' · goală din {num(dd(cT["end_date"]))}' if goala else f' · {num(len(TEL_SECS))} secțiuni')
        hint = f'<span class="ip-hint" id="sn-farm-h">Doar fermele de creștere · ghid {GUIDE["g-growth"]}{state}</span>'
        note = (f'<span class="ip-note">goală din {num(dd(cT["end_date"]))}</span>' if goala else f'<span class="ip-note">{num(len(TEL_SECS))} secțiuni</span>')
        st, inv, desc = '', '', 'sn-farm-h'
    lbl = '<span class="ip-label" id="sn-farm-l">Ferma de creștere</span>'
    val = f'<span class="ip-select__val">Telița</span>{icon("chev", 16)}</div>'
    if touch:
        return (f'<div class="ip-field">{lbl}<div class="ip-select"><div class="ip-input ip-select__btn{t}{st}" id="sn-farm" role="combobox" tabindex="0" aria-haspopup="dialog" '
                f'aria-expanded="false" aria-labelledby="sn-farm-l" aria-describedby="{desc}"{inv}>{val}</div>{hint}</div>')
    opt = (f'<li class="ip-menu__item ip-select__opt" role="option" id="sn-farm-o1" aria-selected="true"><span class="ip-select__chk" aria-hidden="true">{icon("check", 16)}</span>'
           f'<span class="ip-select__txt"><span class="ip-select__top">Telița {note}</span></span></li>')
    return (f'<div class="ip-field">{lbl}<div class="ip-select"><div class="ip-input ip-select__btn{t}{st}" id="sn-farm" role="combobox" tabindex="0" aria-haspopup="listbox" '
            f'aria-expanded="false" aria-controls="sn-farm-list" aria-labelledby="sn-farm-l" aria-describedby="{desc}"{inv}>{val}'
            f'<ul class="ip-menu ip-select__menu" role="listbox" id="sn-farm-list" aria-labelledby="sn-farm-l" tabindex="-1" hidden="hidden">{opt}</ul></div>{hint}</div>')
def date_field(day, touch=True):
    t = ' ip-seg--touch' if touch else ''; ti = ' ip-input--touch' if touch else ''
    return (f'<div class="ip-field"><span class="ip-label" id="d-lbl">Data populării</span><div class="s7-date">'
            f'<span class="ip-seg{t}" role="radiogroup" aria-labelledby="d-lbl"><span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Azi</span>'
            f'<span class="ip-seg__item" role="radio" aria-checked="false" tabindex="-1">Ieri</span></span>'
            f'<span class="ip-input{ti} ip-num">{icon("cal", 16)}<input id="d-val" type="text" inputmode="numeric" value="{dd(day)}" aria-labelledby="d-lbl"/></span></div></div>')
    # S18N: the date's own "SAPT 1 · ciclu nou din …" note is gone: the age field right under it says the SAPT from both
assert 'class="s7-date"' in B7.DATE and 'ip-seg ip-seg--touch' in B7.DATE   # the S7 DateField markup, relabelled
AGE_HINT = '0 pentru pui de o zi. Săptămâna păsărilor (SAPT) se numără de la această vârstă.'   # spec B3 round 2, word for word
def age_err_txt(v):
    """spec B3 round 2 errors (on blur and on save; each a refusal): emptied, out of range, not a whole number."""
    if v == '': return 'Scrie vârsta la populare în zile, 0 pentru pui de o zi.'
    if not v.isdigit(): return f'Scrie un număr întreg de zile, de exemplu {num(0)}.'
    return f'Vârsta la populare e {zile(int(v))}. Scrie o valoare între {num(0)} și {num(AGE_MAX)}.'
assert re.sub('<[^>]+>', '', age_err_txt('160')) == 'Vârsta la populare e 160 de zile. Scrie o valoare între 0 și 140.'   # the spec's example
def age_ok(v): return v.isdigit() and int(v) <= AGE_MAX
def age_field(v=str(AGE0), touch=True, focus=False):
    """S18N · Vârstă la populare (zile) (spec B3 round 2, D151): a CountInput (whole days; desktop md, phone large, as the
    placement rows; no sex bar) with the unit suffix "zile", 0 by default; the spec hint, then the live line from the date
    and the age (Mono figures, polite live region): "La populare: vârstă 0 zile · SAPT 1.". Out of range / empty / not
    whole: fault border + the spec's error line in place of the live line (the hint stays), aria-invalid + errormessage."""
    md = '' if touch else ' ip-count--md'
    ok = age_ok(v); st = (' is-focus' if focus else '') + ('' if ok else ' is-error')
    af = ' autofocus="autofocus"' if focus else ''
    # review (b): ONE persistent polite live region right under the field; its content switches between the live line and
    # the error, so a change either way is announced (the hint stays under it)
    if ok:
        k = int(v)
        msg = f'<span class="ip-hint" id="sn-age-l">La populare: vârstă {zile(k)} · {num(f"SAPT {age_wk(k)}")}.</span>'
        aria = ' aria-describedby="sn-age-l sn-age-h"'
    else:
        msg = f'<p class="ip-err ip-fieldmsg" id="sn-age-e">{icon("alert", 14)}<span>{age_err_txt(v)}</span></p>'
        aria = ' aria-invalid="true" aria-errormessage="sn-age-e" aria-describedby="sn-age-e sn-age-h"'
    return (f'<div class="ip-field s18-age"><label class="ip-label" for="sn-age">Vârstă la populare (zile)</label>'
            f'<span class="ip-input ip-count{md}{st}"><input id="sn-age" inputmode="numeric" pattern="[0-9]*" autocomplete="off" value="{v}"{aria}{af}/>'
            f'<span class="suffix">zile</span></span><div class="s18-agelive" aria-live="polite" aria-atomic="true">{msg}</div>'
            f'<span class="ip-hint" id="sn-age-h">{AGE_HINT}</span></div>')
def ctx_line(day, v=str(AGE0)):
    """the panel/sheet head's Mono context line: farm · date · the new series' SAPT (from the age; none while the age is invalid)."""
    return f'TELIȚA · {dd(day)}' + (f' · SAPT {age_wk(int(v))}' if age_ok(v) else '')

CHEV16 = icon('chev', 16)
def prow(u, i, code, f=None, m=None, touch=True):
    """One user-added placement row, as the SYS07 preview (Serie nouă): SectionSelect combobox, remove key, ♀ / ♂ CountInputs."""
    k = f'{u}{i}'; who = code or f'rândul {i + 1}'
    ti = ' ip-input--touch' if touch else ''; tb = ' ip-iconbtn--touch' if touch else ''; md = '' if touch else ' ip-count--md'
    name = f'<span class="ip-select__val"><b>{code}</b></span>' if code else '<span class="ip-select__val is-empty">Alege</span>'
    def cnt(sx, v):
        c, s, w = ('hen', '♀', 'găini') if sx == 'F' else ('rooster', '♂', 'cocoși'); i_ = f'{"f" if sx == "F" else "m"}-{k}'
        val = f' value="{n(v)}"' if v is not None else ''
        return (f'<div class="ip-field is-{c}"><label class="ip-label ip-sexrow__lbl" for="{i_}"><span class="ip-sr">{who}, </span><span class="ip-sex ip-sex--{c}" aria-hidden="true">{s}</span>'
                f'<span class="ip-sr">{w}</span></label><span class="ip-input ip-count ip-count--{c}{md}"><input id="{i_}" inputmode="numeric" pattern="[0-9 &#8239;]*" placeholder="–" autocomplete="off"{val}/></span></div>')
    if touch:   # FIN10 (section 53, S9P): on phone the SectionSelect opens the BottomSheet, no listbox left in the page
        sel = (f'<div class="ip-select"><div class="ip-input ip-select__btn{ti}" id="sc-{k}" role="combobox" tabindex="0" aria-haspopup="dialog" aria-expanded="false" aria-labelledby="sl-{k}">{name}{CHEV16}</div></div>')
    else:
        sel = (f'<div class="ip-select"><div class="ip-input ip-select__btn{ti}" id="sc-{k}" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="false" aria-controls="sm-{k}" aria-labelledby="sl-{k}">{name}{CHEV16}</div>'
               f'<ul class="ip-menu ip-select__menu" role="listbox" id="sm-{k}" aria-labelledby="sl-{k}" tabindex="-1" hidden="hidden"></ul></div>')
    return (f'<li class="ip-sexrow"><div class="ip-sexrow__head"><div class="ip-field ip-sexrow__sel"><span class="ip-label ip-sexrow__lbl" id="sl-{k}">Secțiune</span>'
            f'{sel}</div>'
            f'<button type="button" class="ip-iconbtn{tb} ip-sexrow__rm" aria-label="Elimină {who}">{icon("x")}</button></div>{cnt("F", f)}{cnt("M", m)}</li>')
def prows(u, day, rows, touch=True):
    t = ' ip-btn--touch' if touch else ''
    items = ''.join(prow(u, i, c, f, m, touch) for i, (c, f, m) in enumerate(rows))
    filled = [r for r in rows if r[0] and (r[1] is not None or r[2] is not None)]
    sf, sm = sum(r[1] or 0 for r in filled), sum(r[2] or 0 for r in filled)
    if filled:
        sub, tot = f'<span>{num(len(filled))} secțiuni</span>', pair(sf, sm)
    else:
        sub, tot = '<span>niciun rând completat</span>', pair('–', '–')
    tot = tot.replace('class="ip-pair"', 'class="ip-pair ip-pair--lg"', 1)
    return (f'<section class="ip-panel ip-sexrows ip-sexrows--add" aria-labelledby="t-{u}"><div class="ip-sexrows__head"><h3 class="ip-cap" id="t-{u}">Populare pe secțiuni</h3>'
            f'<span class="ip-note">Telița · {num(dd(day))}</span></div><ol class="ip-sexrows__list">{items}</ol>'
            f'<button type="button" class="ip-btn ip-btn--ghost{t} ip-sexrows__add">{icon("plus", 16)}Adaugă secțiune</button>'
            f'<div class="ip-sexrows__total"><div class="ip-sexrows__tk"><span class="ip-cap">Total populare</span><span class="ip-sexrows__sub">{sub}</span></div>{tot}</div></section>')
def cmark(kind, cls, text):  # SYS06 cycle mark (as s10_build): record bar + mini block
    bar = '<i class="ip-cmark__bar"></i>' if cls else ''
    return f'<span class="ip-cmark ip-cmark--{kind}{(" " + cls) if cls else ""}"><span class="ip-cmark__g" aria-hidden="true">{bar}<i class="ip-cmark__blk"></i></span>{text}</span>'
REC = f'<span class="ip-rec ip-rec--placement s13-rec">{icon("placement", 16)}</span>'
ARROW = '<span class="s9-to" aria-hidden="true">→</span><span class="ip-sr">după populare</span>'
def effect(state, day, name='', year='', rows=(), wk=NEW_WK):   # wk None: no valid age, no SAPT (as the head)
    if state == 'empty':
        return (f'<div class="ip-effect s13-effect" id="fx" role="status">{REC}<div class="s9-fxbody"><span class="ip-cap">Efect</span>'
                f'<span>Alege secțiunile Teliței și scrie câte găini și cocoși intră în fiecare. Aici vezi seria nouă și ciclul nou de la Telița.</span></div></div>')
    if state == 'refuz':
        return (f'<div class="ip-banner ip-banner--fault s7s-banner" role="alert" id="sn-ref">{icon("alert", 16)}<div class="s7s-banner__body">'
                f'<p class="ip-m0">Telița are încă {n(R39ALL)} păsări din Seria 2026. Mută-le sau vinde-le înainte, sau începe un ciclu nou manual.</p>'
                f'<p class="ip-m0 s9-fxsub">în viață la Telița pe {num(dd(TODAY))}: {pair(R39["F"], R39["M"])}</p>'
                f'<ul class="s18-ways" aria-label="Ce poți face">'
                f'<li><a class="ip-link" href="#/inregistreaza/transfer?din=telita">Mută păsările de la Telița</a></li>'
                f'<li><a class="ip-link" href="#/inregistreaza/vanzare?ferma=telita">Vinde păsările de la Telița</a></li>'
                f'<li><a class="ip-link" href="#/ferme/telita?actiune=ciclu-nou">Începe ciclu nou la Telița</a></li></ul></div></div>')
    sf, sm = sum(r[1] or 0 for r in rows), sum(r[2] or 0 for r in rows)
    codes = ', '.join(r[0] for r in rows)
    return (f'<div class="ip-effect s13-effect" id="fx" role="status">{REC}<div class="s9-fxbody"><span class="ip-cap">Efect</span>'
            f'<span class="s9-fxline"><b class="s9-strong">{name} · an {year}</b><span>serie nouă, deschisă</span></span>'
            f'<span class="s9-fxline"><b class="s9-strong">Telița · {dd(day)}</b><span class="s9-fxval">{pair(0, 0)}{ARROW}{pair(sf, sm)}</span></span>'
            f'<span class="s9-fxsub">{num(len(rows))} secțiuni · {codes}</span>'
            f'<span class="s9-fxcyc">{cmark("start", "place", "Ciclu nou la Telița" + (f", sapt. {wk}" if wk else ""))}</span>'
            f'<span class="s9-fxsub">ghid {GUIDE["g-growth"]}</span></div></div>')   # S18P: the guide line (it names the type); mark wording as C3/S9

def form(day, touch, name='', year='', rows=((None, None, None),), state='empty', focus=False, age=str(AGE0), age_focus=False):
    u = 'p' if touch else 'd'
    parts = [tfield('sn-name', 'Nume', name, 'Apare în selectorul de serii, pe pagini și în rapoarte.', touch, focus=focus),
             tfield('sn-year', 'An', year, '', touch, mono=True, numeric=True, cls=' s18-year'),
             farm_field(day, touch, refused=state == 'refuz'), date_field(day, touch), age_field(age, touch, age_focus)]   # S18N: the age after the date (spec B3 flow 2)
    if state == 'refuz':
        parts.append(effect('refuz', day))
    parts.append(prows(u, day, list(rows), touch))
    if state != 'refuz':
        parts.append(effect(state, day, name, year, [r for r in rows if r[0]], age_wk(int(age)) if age_ok(age) else None))
    return parts
CREATE = 'Creează seria'
def sheet(body, day, key, age=str(AGE0)):
    return (f'<aside class="ip-sidepanel ip-sidepanel--sheet" role="dialog" aria-modal="true" aria-labelledby="sn-t">{p_head(ctx_line(day, age), True)}'
            f'<div class="ip-sidepanel__body ip-sidepanel__body--form">{"".join(body)}</div><div class="ip-sidepanel__foot">{key}</div></aside>')
def panel(body, day, key, age=str(AGE0)):
    return (f'<aside class="ip-sidepanel ip-sidepanel--dock" id="sn" aria-labelledby="sn-t">\n{p_head(ctx_line(day, age), False)}\n'
            f'<div class="ip-sidepanel__body ip-sidepanel__body--form">\n{"".join(body)}\n</div>\n'
            f'<div class="ip-sidepanel__foot s18-foot"><div class="ip-row ip-sidepanel__keys"><button type="button" class="ip-btn ip-btn--secondary">Anulează</button>{key}</div></div>\n</aside>')

# ================================================================= phone list (owner)
def p_header():
    h = CRUMBS.replace('<a href="#setari">Setări</a>', '<a href="#setari"><span class="ip-crumbs__txt">Setări</span></a>').replace('<nav class="ip-crumbs"', '<nav class="ip-crumbs ip-crumbs--touch"', 1)  # FIN06: system touch crumb (section 52)
    return f'<header class="ip-pagehead s2-phead s14-phead"><div>{h}<h1>Serii</h1></div></header>'
def phone_list():
    key = f'<div class="s18-pkey">{newkey(touch=True)}</div>'
    body = A.phone([p_header(), key, track(phone=True), table(touch=True)])
    for a, z in (('class="ip-tabbar__item is-current" href="#acasa" aria-current="page"', 'class="ip-tabbar__item" href="#acasa"'),
                 ('class="ip-tabbar__item" href="#mai-mult"', 'class="ip-tabbar__item is-current" href="#mai-mult" aria-current="page"')):
        assert body.count(a) == 1, a; body = body.replace(a, z)
    assert body.count('is-current') == 1
    return body

# ================================================================= CSS (screen-only; tokens only; system classes untouched)
def rule(css, sel):
    r = re.findall(r'[^}\n]*' + re.escape(sel) + r'[^}]*}', css); assert r, sel
    return '\n' + '\n'.join(r)
DATE_CSS = rule(B7.CSS, '.s7-date')
# FIN15B: the CycleTrack width box (1160/1000/720/560; fix 1: beside the panel at 1280 the 560 drawing, 1:1) is additions.css §65
CSS18 = '''
.s18-table{width:100%}
.s18-table thead th{white-space:nowrap}
.s18-table td{white-space:nowrap} /* FW04 (D150): every cell on one line, so every cycle row is the row height (Ciclu, Început wrapped at 1440/1280: 51 beside 34); Început takes the rest */
.s18-start .ip-pair,.s18-cyc > span{white-space:nowrap}
.s18-table tbody > tr > :is(th,td){height:calc(var(--row-h) + var(--ip-hair))} /* FW04 fix 1 (D150): one height for every row: a cycle row's 24 px farm link + space-1 above and below + its rule (33), so the series rows and the "fără date" row match (were 32 beside 33 / 33.5) */
.s18-table .s18-start,.s18-table .s18-rlink{vertical-align:middle} /* the icon/link line box no longer adds half a pixel */
.s18-g{display:inline-flex;align-items:center;flex-wrap:wrap;gap:var(--space-1) var(--space-3)}
.s18-g > .ip-note{font-weight:400}
.s18-gname{font-weight:500}
.s18-gact{text-align:right}
.s18-table td.s18-gact{padding-block:0;vertical-align:middle}
.s18-acts{vertical-align:middle}
.s18-gact .ip-btn{height:calc(var(--row-h) - var(--space-1))} /* FW04 (D150): the series keys in the group row are row-h less space-1 (28, as §69 in-row keys), so the series row is the row height (was 45) */
.s18-acts{display:inline-flex;gap:var(--space-2);justify-content:flex-end;align-items:center}
.s18-rlink{color:var(--ink);text-decoration:underline;text-decoration-color:var(--line-strong);text-underline-offset:calc(3 * var(--ip-hair));display:inline-flex;align-items:center;min-height:var(--space-6)}
.s18-rlink:hover{text-decoration-color:var(--ink)}
.s18-start{display:inline-flex;align-items:center;gap:0 var(--space-2)}
.s18-notes{display:flex;flex-direction:column;gap:var(--space-1)}
.s18-year .ip-input{width:calc(var(--space-12) * 2 + var(--space-6))} /* a 4-digit year: 120 px */
.s18-ways{list-style:none;margin:0;padding:0;display:flex;flex-direction:column;gap:var(--space-1)}
.s18-ways .ip-link{display:inline-flex;align-items:center;min-height:var(--space-6)}
.s18-vp > .ip-shell{min-height:100%} /* the clipped page beside the panel keeps its sidebar full height */
.s18-vp .s18-table th{width:auto} .s18-vp .s18-table th::after{content:none} /* FIN11G: beside the panel the columns stay content-sized (the §54 shares and minimums are the list alone's) */'''
X_CSS = '\n'.join(l for l in X.CSS.split('\n') if not l.startswith('.s11-select{'))   # FIN10 fix 1: no native select on S18–S22 desktops
assert X_CSS != X.CSS
CSS_DESK = X_CSS + SC.SIDE + SC.M0 + SC.REC + SC.EFFECT + SC.FX + SC.BANNER + DATE_CSS + CSS18 + '''
.s18-foot{justify-content:flex-end}'''
# S18N: the age field's width; kept out of CSS18/CSS_DESK/CSS_PL, which S19–S22 reuse (their boards stay byte-identical)
CSS_AGE = '''
.s18-age .ip-input{width:calc(var(--space-12) * 2 + var(--space-6))} /* "140" + "zile": 120 px, as the year */
.s18-agehead > span{white-space:nowrap} /* SYS17 (D189): the head on one line (was two, D179); the table runs past 1280 under its sticky column (D179 note 2) */'''
CSS_D18 = CSS_DESK + CSS_AGE   # the S18 desktops (s18s too)
Z_CSS = '\n'.join(l for l in Z.CSS_PHONE.split('\n') if '.s11-select' not in l)   # FIN10 fix 1: no native select on S18–S22 phones
CSS_PL = Z_CSS + CSS18 + '''
.s18-pkey{display:flex;flex-direction:column}
.s18-pser{display:flex;flex-direction:column;gap:var(--space-2)}
.s18-pser + .s18-pser{margin-top:var(--space-2)}
.s18-pacts{display:flex;flex-wrap:wrap;gap:var(--space-2)}
.s18-gact .ip-btn{white-space:nowrap}
.s18-ways .ip-link{min-height:var(--control-h-touch)}
.s18-rlink{min-height:var(--control-h-touch);min-width:var(--control-h-touch)} /* touch: the farm link in a table row is a 48 px target */
.s18-table tbody > tr:not(.is-group) > :is(th,td){height:calc(var(--control-h-touch) + var(--space-2))} /* FW04 (D150): every cycle row as tall as a row with its 48 px farm link (the "fără date" row was 32 beside 57) */'''
CSS_PL18 = CSS_PL + CSS_AGE    # S18N: the S18 phone list (CSS_PL itself is S19–S22's too)
SHEET = ''  # FIN15B: the sheet's scrolling form body is additions.css §62 ip-sidepanel__body--form
CSS_SH = ('body{margin:0}\n.s18-screen{height:100%;display:flex}' + SHEET + SC.M0 + SC.REC + SC.EFFECT + SC.FX + SC.BANNER + DATE_CSS + CSS18 + CSS_AGE + '''
.s18-ways .ip-link{min-height:var(--control-h-touch)} /* touch: each way out is a 48 px target */''')

# ================================================================= write
FONTS = s4.FONTS; SYS = ROOT / 'design/system'
pathlib.Path('/tmp/s18').mkdir(exist_ok=True)
ROOT_D, ROOT_P = '<div class="ip-root s4s-rel" style="padding:0; ', '<div class="ip-root" style="padding:0; width:390px; '
CSS_PWIDE = '''
.s4ns-wide > .ip-screen{width:390px}
.s4ns-wide .ip-tablescroll{width:min-content;min-width:100%;max-width:none;overflow:visible}'''   # as s4ns_build.CSS_PWIDE
def wide(stem, root, css, w):
    """D167 4 / D181 (FV02): a board whose Serii table is wider than its page region (beside the docked SidePanel; the phone's
    358) is drawn wider (board_widths.json): the page keeps its layout (1440 less the panel / 390) and the table runs past its
    right edge in full (s4s_build.CSS_WIDE / the s4ns phone pattern, canvas only); boards at their base width are unchanged."""
    W = CR.width(stem, w)
    if W == w: return root, css, w
    if w == 1440:
        assert root.count(ROOT_D) == 1, stem
        return root.replace(ROOT_D, '<div class="ip-root s4s-rel s4s-wide" style="padding:0; --board-page-w:1440px; ', 1), css + '\n' + s4.CSS_WIDE, W
    assert w == 390 and root.count(ROOT_P) == 1, stem
    return root.replace(ROOT_P, f'<div class="ip-root s4ns-wide" style="padding:0; width:{W}px; ', 1), css + CSS_PWIDE, W
def write(stem, title, root, css, w, h):
    root = V.add(root)   # VO02
    root, css, W = wide(stem, root, css, w)   # out keeps the base w (phone/desktop); canvas_rows.place widens it again
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
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":{W},"height":{h}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
    (ROOT / f'design/canvas/project/{stem}.dc.html').write_text(dc, encoding='utf-8')
    # RF10 (as RF03's s4s_build.write): the render wrapper keeps only the viewport floor (1024 / 844), not the stored h, so
    # s18_render.py --measure reads the content's natural height instead of the stored one
    root = re.sub(r'(?<![\w-])(min-height|height):\d+px', lambda m: f'{m.group(1)}:{1024 if w == 1440 else 844}px', root, count=1)
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{FONTS}<style>{css}</style></head><body>{root}</body></html>'''
    pathlib.Path(f'/tmp/s18/{stem}.html').write_text(wrap, encoding='utf-8')
    print('ok', stem, W, h)
    return {'stem': stem, 'title': title, 'w': w, 'h': h}

STEM, T = 'S18-serii', 'S18 · Serii'
out = []
# 1 · desktop list
h = HEIGHTS.get(STEM, 1024)
out.append(write(STEM, T, f'<div class="ip-root s4s-rel" style="padding:0; min-height:{h}px">{page()}</div>', CSS_D18, 1440, h))
# 2 · desktop, the SidePanel open beside the list: nothing typed, Telița + Azi preset, focus on Nume (fields validate on save)
DKEY = f'<button type="button" class="ip-btn ip-btn--primary">{CREATE}</button>'
PAN = panel(form(TODAY, False, focus=True), TODAY, DKEY)
VP = f'<div class="ip-dockpage s18-vp">{page(pressed=True)}</div>'
h = HEIGHTS.get(STEM + '--serie-noua', SC.SIDE_H)   # S18N: natural height (the panel body never scrolls on the board)
out.append(write(STEM + '--serie-noua', f'{T} · Serie nouă · desktop', f'<div class="ip-root s4s-rel" style="padding:0; min-height:{h}px">{VP}{PAN}</div>', CSS_D18, 1440, h))
# 3 · phone list
h = HEIGHTS.get(STEM + '--telefon', 844)
out.append(write(STEM + '--telefon', f'{T} · telefon', f'<div class="ip-root" style="padding:0; width:390px; min-height:{h}px">{phone_list()}</div>', CSS_PL18, 390, h))
# 4 · phone sheet, replay 07.10.2025: the real creation of Seria 2026
ROWS = [(c, q if sx == 'F' else None, q if sx == 'M' else None) for c, sx, q in PROWS]
assert sum(r[1] or 0 for r in ROWS) == PLC['F'] and sum(r[2] or 0 for r in ROWS) == PLC['M']
s26 = SER['seria-2026']
PKEY = f'<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">{CREATE}</button>'
body = form(REPLAY, True, s26['name'], str(s26['year']), ROWS, state='set', focus=True)   # fix 1: focus on the first field (S16 rule)
h = HEIGHTS.get(STEM + '--telefon-serie-noua', 844)
out.append(write(STEM + '--telefon-serie-noua', f'{T} · Serie nouă · telefon · reluare {dd(REPLAY)}',
                 f'<div class="ip-root" style="padding:0; width:390px; height:{h}px; overflow:hidden"><div class="s18-screen">{sheet(body, REPLAY, PKEY)}</div></div>', CSS_SH, 390, h))
# 5 · phone sheet, refusal (D39): Telița still holding Seria 2026 birds
RKEY = f'<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled="disabled" aria-describedby="sn-ref">{CREATE}</button>'
body = form(TODAY, True, NEWNAME, NEWYEAR, state='refuz', focus=True)
h = HEIGHTS.get(STEM + '--telefon-refuzat', 844)
out.append(write(STEM + '--telefon-refuzat', f'{T} · Serie nouă · telefon · refuzat: Telița are încă păsări · exemplu de stare (D39)',
                 f'<div class="ip-root" style="padding:0; width:390px; height:{h}px; overflow:hidden"><div class="s18-screen">{sheet(body, TODAY, RKEY)}</div></div>', CSS_SH, 390, h))
pathlib.Path('/tmp/s18/boards.json').write_text(json.dumps([{k: b[k] for k in ('stem', 'w', 'h')} for b in out]))

if CANVAS:  # S18 row, entries keep x/y, a new row goes below the lowest (canvas_rows.py); 80 px gaps
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    x = CR.place(c, STEM, [(b['stem'], b['w'], b['h'], b['title']) for b in out])
    CR.save(ROOT, c)
    print('canvas: S18 row y', CR.row_y(c, STEM), 'ends at x', x - 80)
