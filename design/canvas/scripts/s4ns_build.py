"""S4NS · S4 Fermă states on real data, table first (spec D1 › States, Filters, phone; PeriodTable phone/states; D160, D166, D167).
Boards appended to the S4 row after its existing boards (desktops, then phones, D74), owner, today 18.11.2026 = SAPT 59,
Hîrbovăț · Seria 2026, period "Ciclul curent" (D191; the tables open on its last 4 weeks, SAPT 56–59, D190):
- Rezumat · etichete Mortalitate + Vânzare (desktop): chips Mortalitate and Vânzare on, the others off (D45 types of the
  period); the table keeps În viață (always on Rezumat) + Morți + Vânzare; the latest records list only those two types.
- Rezumat · pas Săpt. (ales în filtre) (desktop): grain Săpt. = week rows only, every week closed (each key can open its week
  alone), captioned as the reader's choice (D171; the "one week closed" board is dropped, D167/D171).
- Rezumat · telefon, Păsări · telefon (390, whole tall pages, BottomTabBar with Serie current): FilterBar = one key opening a
  BottomSheet; the table shows one column group per tag (D160/D171: single-choice per tab, default = the tab's own type,
  Păsări Mortalitate, Hrană Hrană, Greutate Cântărire, Producție Ouă; Rezumat Mortalitate; În viață stays), spine folded into
  the sticky first column ("Ma 17.11 · 406", week dates on a second line, D171), 48 px rows (ip-ptable--phone), every week
  open (D166); the compact one-farm CycleTrack (358) under the selector (D171); Păsări has the Detaliu SegmentedControl at
  touch size and "% cumulat" with the cull note (D171); Readouts without their GuideChart (as S14 phone).
- Păsări · telefon · filtre (390×844): the FilterBar BottomSheet open: Perioadă and Pas at touch size, the record-type tags
  as a radiogroup of touch chips (one on), Exportă (the whole period, D191).
- MO03 (D197): Rezumat · rând nou (desktop): the 13:16 sale arriving live (the list row rolls in with "Nou", the figures it
  moves tinted, "Actualizat 13:16"), frozen at the tint's start.
- SYS18 (D190): Rezumat · 8 săptămâni (desktop) and Păsări · telefon · 8 săptămâni: the table after one press of "Încarcă încă
  4 săptămâni": SAPT 59–52 open with their days, the foot "Total SAPT 52–59", the key focused, the live region's sentence.
Reuses by import: s4t_build (→ s4s_build: shell, header, FilterBar parts, Rezumat page and PeriodTable wrapper; Păsări groups,
panel, Readouts), sys14_build through s4s (`P`: ptable, block, region), s2_build (phone screen, BottomTabBar, phone CSS,
write_phone; importing it rewrites the S2/S7 default boards byte-identical). Figures only from realdata.
python3 -I s4ns_build.py <repo_root> [--canvas]
Writes design/canvas/project/<stem>.dc.html, /tmp/s4s/<stem>.html (desktops), /tmp/s2/<stem>.html (phones); heights from
s4ns_heights.json (s4ns_render.py --heights); --canvas places the 5 boards after the S4 row's last board (x step w + 80)."""
import sys, re, json, pathlib, importlib, importlib.util

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]); HERE = pathlib.Path(__file__).resolve().parent
HP = HERE / 's4ns_heights.json'
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}
sys.path.insert(0, str(HERE))

sys.argv = ['s4t_build.py', str(ROOT)]
_sp = importlib.util.spec_from_file_location('s4t_build', HERE / 's4t_build.py')
T = importlib.util.module_from_spec(_sp); _sp.loader.exec_module(T)   # writes nothing on import
s4, P, R = T.s4, T.P, T.R
sys.argv = ['s2_build.py', str(ROOT), str(HERE / 's2_heights.json')]
A = importlib.import_module('s2_build')                               # phone shell; rewrites the S2/S7 default boards unchanged
icon, lamp, num = s4.icon, s4.lamp, T.num
WK, PTYPES, CHIPW = s4.WK, s4.CTYPES, s4.CHIPW   # D191: the period's (the cycle's) types
assert P.WEEKS == [WK, WK - 1, WK - 2, WK - 3] == [59, 58, 57, 56]

# ================================================================= shared: the Rezumat page with another FilterBar / table
FB = s4.filterbar4()
SEG_ZI = s4.seg('Pas', ['Zi', 'Săpt.'], 0)
CHIPS_ALL = s4.chips4(set(PTYPES))
W8 = s4.W8
assert FB.count(SEG_ZI) == 1 and FB.count(CHIPS_ALL) == 1
def rz_page(fb=FB, table=None, events=None):
    panel = s4.panel_default(table=table)
    if events is not None:
        old = ''.join(s4.events); assert panel.count(old) == 1; panel = panel.replace(old, ''.join(events))
    pg = s4.shell(s4.main_page(panel=panel))
    assert pg.count(FB) == 1
    return pg.replace(FB, fb)
def rz_table(groups, note, open_weeks=None):
    return s4.ptable('rz', 'Rezumat', [('Total fermă', '', groups)], bar=False, open_weeks=open_weeks, after=s4.ptnote(note))
def gname(g):  # the record group's visible word ("În viață", "Morți" …) without its head icon
    return re.sub(r'<[^>]+>', '', g['name']).strip()
RZ = s4.rezumat_groups()
assert [gname(g) for g in RZ] == ['În viață', 'Morți', 'Ouat', 'Hrană kg', 'Vânzare']   # D167 order

# ================================================================= 1. tags on/off: only Mortalitate + Vânzare (desktop)
ON = {'mortality', 'sale'}
FB_TAGS = FB.replace(CHIPS_ALL, s4.chips4(ON))
RZ_TAGS = [g for g in RZ if gname(g) in ('În viață', 'Morți', 'Vânzare')]
NOTE_TAGS = ('Morți = morții din registru, sacrificările incluse (Păsări le arată separat). '
             'Etichete: Mortalitate și Vânzare; Ouat și Hrană kg revin cu etichetele Ouă și Hrană.')
EV_TAGS = [e for e in R.EVENTS if e['record'] in ('mortality', 'sale')][:len(s4.EVENTS)]   # the log under the table: only the chosen types
assert len(EV_TAGS) == 9 and {e['record'] for e in EV_TAGS} == ON
def board_tags():
    return rz_page(FB_TAGS, rz_table(RZ_TAGS, NOTE_TAGS), [s4.event_row(e) for e in EV_TAGS])

# ================================================================= 2. grain Săpt.: week rows only (desktop)
FB_WEEK = FB.replace(SEG_ZI, s4.seg('Pas', ['Zi', 'Săpt.'], 1))
NOTE_WEEK = 'Pas Săpt., ales în filtre: doar rândurile de săptămână; fiecare tastă deschide săptămâna ei. Zi (implicit) le arată pe toate cu zilele lor.'
# ================================================================= SYS18: after one press of "Încarcă încă 4 săptămâni" (desktop)
def board_4():   # D192: the period "4 săpt." chosen in the FilterBar: the load-more row shows too
    t = s4.tb_rezumat()
    assert '4</span> din <span class="ip-num">40' in t
    return rz_page(s4.filterbar_weeks(4), t)
def board_8():     # D192: after one press with "4 săpt.": the filter reads "8 săpt.", SAPT 59–52 open, the row keeps focus
    t = s4.tb_rezumat(weeks=W8, added=W8[4:], period='8 săpt.')
    assert t.count('aria-expanded="true"') == 8 and 'Total SAPT <span class="ip-num">52–59</span>' in t and '8</span> din <span class="ip-num">40' in t
    return rz_page(s4.filterbar_weeks(8), t)

# ================================================================= MO03 (D197): the 13:16 sale arrives live (desktop)
# Frozen right after the arrival, the tint at its start (data-live-tint="on", no data-state: nothing moves on the board): the
# 18.11 Vânzare record rolls into "Ultimele înregistrări" with its "Nou" mark; the figures it moves take the sale tint: the
# 18.11 day row's În viață and Vânzare cells, the SAPT 59 week row's and the foot's, the page head's ♀ · ♂; the head meta
# reads "Actualizat 13:16". (The 18.11 day row itself exists since the 08:15 Zilnic, so in the table it is a value change.)
LIVE_EV = R.EVENTS[0]
assert (LIVE_EV['at'], LIVE_EV['record']) == ('2026-11-18T13:16', 'sale')
LIVE_AT = LIVE_EV['at'][11:16]
LV = ' data-live="changed" data-live-rec="sale" data-live-tint="on"'
def live_cells(row):   # the În viață (untyped group) and Vânzare cells of one table row
    def one(m):
        c = m.group(1)
        hit = 'ip-ptable__g--sale' in c or ('ip-ptable__g' in c.split() and 'ip-ptable__g--' not in c)
        return f'<td class="{c}"{LV if hit else ""}>'
    out, k = re.subn(r'<td class="([^"]*)">', one, row)
    assert out.count(LV) == 4, row
    return out
def board_live():
    t = s4.tb_rezumat()
    day = re.search(r'<tr class="ip-ptable__day" id="rz-1118">.*?</tr>', t, re.S).group(0)
    wk = re.search(r'<tr class="is-total ip-ptable__wk">(?:(?!</tr>).)*?SAPT <span class="ip-num">59</span>.*?</tr>', t, re.S).group(0)
    ft = re.search(r'<tfoot>.*?</tfoot>', t, re.S).group(0)
    for part in (day, wk, ft):
        assert t.count(part) == 1; t = t.replace(part, live_cells(part))
    ev0 = s4.events[0]; assert ev0 == s4.event_row(LIVE_EV) and ev0.startswith('<li><a class="ip-ev"') and ev0.endswith('</a></li>')
    ev0 = ('<li data-live="new" data-live-rec="sale" data-live-tint="on">' + ev0[len('<li>'):-len('</a></li>')]
           + '<span class="ip-tag ip-tag--info ip-live__mark">Nou</span></a></li>')
    pg = rz_page(FB, t, [ev0] + s4.events[1:])
    h0 = pg.index('<header class="ip-pagehead">'); h1 = pg.index('</header>', h0); head = pg[h0:h1]
    pair = '<span class="ip-pair">'; meta = f'<span class="ip-note ip-num">SAPT {WK} · ciclu din'
    assert head.count(pair) == 1 and head.count(meta) == 1
    head = head.replace(pair, f'<span class="ip-pair"{LV}>', 1)
    head = head.replace(meta, f'<span class="ip-note ip-live__at">Actualizat <span class="ip-num">{LIVE_AT}</span></span>{meta}', 1)
    pg = pg[:h0] + head + pg[h1:]
    sr = f'<span class="ip-sr" role="status" aria-live="polite">O înregistrare nouă: Vânzare, {R.sale_codes(LIVE_EV)}, {R.ddt(LIVE_EV["at"])}, {LIVE_EV["user"]}.</span>'
    k = '<h1>Hîrbovăț</h1>'; assert pg.count(k) == 1
    return pg.replace(k, k + sr, 1)

def board_week():
    t = s4.ptable('rz', 'Rezumat', [('Total fermă', '', RZ)], bar=False, open_weeks=(), after=s4.ptnote(NOTE_WEEK, s4.RZ_NOTE))
    assert t.count('aria-expanded="false"') == 4 and 'aria-expanded="true"' not in t
    return rz_page(FB_WEEK, t)

# ================================================================= phones (390): owner, Serie current in the BottomTabBar
SLIDERS = ('<path d="M21 4h-7"/><path d="M10 4H3"/><path d="M21 12h-9"/><path d="M8 12H3"/><path d="M21 20h-5"/><path d="M12 20H3"/>'
           '<path d="M14 2v4"/><path d="M8 10v4"/><path d="M16 18v4"/>')   # Lucide sliders-horizontal (as S14 phone)
# D160/D171: one tag at a time, per tab; each tab defaults to its own record type (Rezumat: Mortalitate, its first group after În viață)
TAB_TAG = {'rezumat': 'mortality', 'pasari': 'mortality', 'hrana': 'feed', 'greutate': 'weighing', 'productie': 'eggs'}
assert all(t in PTYPES for t in TAB_TAG.values())
TAG = TAB_TAG['pasari']
SUM = f'Ciclul curent · Zi · {CHIPW[TAG]}'   # D191
def p_filter(open_=False):
    st = 'aria-expanded="true" aria-controls="pf-sheet"' if open_ else 'aria-expanded="false"'
    return ('<div class="ip-filterbar s4ns-pfilter" role="group" aria-label="Filtre">'
            f'<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch" aria-haspopup="dialog" {st} aria-describedby="pf-sum">'
            f'<svg class="ip-icon ip-icon--16" viewBox="0 0 24 24" aria-hidden="true">{SLIDERS}</svg>Filtrează sau exportă</button>'
            f'<span class="ip-note" id="pf-sum">{SUM}</span></div>')
def p_header():  # PageHeader on a phone: touch crumbs (section 52), the meta line, the actions wrapping under it
    h = s4.header('default').replace('<header class="ip-pagehead">', '<header class="ip-pagehead s2-phead s4ns-phead">', 1)
    a = '<nav class="ip-crumbs" aria-label="Cale"><a href="#seria-2026">Seria 2026</a>'
    assert h.count(a) == 1
    return h.replace(a, '<nav class="ip-crumbs ip-crumbs--touch" aria-label="Cale"><a href="#seria-2026"><span class="ip-crumbs__txt">Seria 2026</span></a>', 1)
import datetime as dt
TW_P = 358                                                       # D171: the phone's one-farm CycleTrack, drawn 1:1 for the 358 px box
DOM_P = (dt.date(2025, 3, 1), dt.date(2027, 4, 1))               # to 04.2027 so the SAPT readout fits right of the azi line (as S2 phone)
def track_phone():
    """s4s.track at 358 (11px text): short block labels (as S18 telefon), 44 px hit rects on the block links (as S2 phone), month
    labels on quarter starts, readout 52 wide; same ARIA as the desktop drawings (labels, desc, aria-current)."""
    d0, d1 = DOM_P; ppd = TW_P / (d1 - d0).days
    def tx(d): return round(((dt.date.fromisoformat(str(d)) - d0).days) * ppd, 1)
    c25, c26, today = s4.c25, s4.c26, s4.TODAY
    x25a, x25b, x26a, xt = tx(c25['start_date']), tx(s4.PREV_END), tx(c26['start_date']), tx(today)
    w25, w26 = round(x25b - x25a, 1), round(xt - x26a, 1)
    full = s4.track(today, '2026', s4.DOM_LIVE, s4.TWS[0])                   # the desktop drawing: its block labels and desc
    lab25, lab26 = re.findall(r'<a href="#ciclu-seria-20\d\d"[^>]*aria-label="([^"]+)"', full)
    desc = re.search(r'<desc>(.*?)</desc>', full).group(1)
    def hit(xa, w): return f'<rect x="{xa}" y="3" width="{w}" height="44" fill="none" pointer-events="all"/>'
    b25 = (f'<a href="#ciclu-seria-2025" aria-label="{lab25}"><title>{lab25}</title>{hit(x25a, w25)}<rect class="ring" x="{round(x25a-4,1)}" y="12" width="{round(w25+8,1)}" height="30" rx="3"/>'
           f'<rect class="cyc" x="{x25a}" y="16" width="{w25}" height="18" rx="2"/><text class="cyctxt" x="{round(x25a+7,1)}" y="29">2025 · ouat</text>'
           f'<line class="ded" x1="{round(x25a-3,1)}" x2="{round(x25a-3,1)}" y1="14" y2="40"/></a>')
    b26 = (f'<a href="#ciclu-seria-2026" aria-current="true" aria-label="{lab26}"><title>{lab26}</title>{hit(x26a, w26)}<rect class="ring" x="{round(x26a-4,1)}" y="12" width="{round(w26+8,1)}" height="30" rx="3"/>'
           f'<rect class="cyc open" x="{x26a}" y="16" width="{w26}" height="18" rx="2"/><text class="cyctxt open" x="{round(x26a+7,1)}" y="29">2026 · ouat</text>'
           f'<rect class="xfer" x="{round(x26a-1,1)}" y="16" width="3" height="18"/><rect class="sel" x="{x26a}" y="37" width="{w26}" height="2"/></a>')
    ro = f'<rect class="ro" x="{round(xt+4,1)}" y="17" width="52" height="16" rx="2"/><text class="rotxt" x="{round(xt+30,1)}" y="28.5" text-anchor="middle">SAPT {WK}</text>'
    ticks = ''
    y, m = d0.year, d0.month
    while dt.date(y, m, 1) < d1:
        x = tx(dt.date(y, m, 1)); ticks += f'<line class="tick" x1="{x}" x2="{x}" y1="40" y2="45"/>'
        if (m - 1) % 3 == 0 and dt.date(y, m, 1) > d0: ticks += f'<text class="month" x="{x}" y="56" text-anchor="middle">{m:02d}.{str(y)[2:]}</text>'
        y, m = (y + 1, 1) if m == 12 else (y, m + 1)
    assert xt + 56 <= TW_P
    return (f'<svg class="ip-track ip-track--farm s4ns-track" viewBox="0 0 {TW_P} 60" role="group" aria-label="Ciclurile fermei Hîrbovăț"><desc>{desc}</desc>'
            f'<rect class="lane" x="0" y="14" width="{TW_P}" height="26"/>{b25}{b26}<line class="today" x1="{xt}" x2="{xt}" y1="12" y2="42"/>'
            f'<text class="todaylbl" x="{xt}" y="9" text-anchor="middle">azi {s4.ddate(today)[:5]}</text>{ro}{ticks}</svg>')
def p_cycle():  # the CycleSelector field (touch), then the compact one-farm CycleTrack at 358 (D171) and its legend
    c = s4.cycle('default')
    c2, k = re.subn(r'<div class="ip-trackbox">.*?</div>(<div class="ip-tracklegend")', lambda m: f'<div class="ip-trackbox">{track_phone()}</div>' + m.group(1), c, count=1, flags=re.S); assert k == 1
    k = 'class="ip-input ip-select__btn"'; assert c2.count(k) == 1
    return c2.replace(k, 'class="ip-input ip-select__btn ip-input--touch"', 1)
CHART_PARTS = (r'<svg class="ip-chart".*?</svg>\n?', r'<span class="ip-sr" id="[\w-]+-help">.*?</span>\n?', r'<div class="ip-chart__read"[^>]*>.*?</div>\n?',
               r'<table class="ip-sr">.*?</table>\n?', r'<div class="ip-chart__legend">.*?</span></div>\n?')
def nochart(html, k):  # Readouts without their GuideChart (the 540 drawing would set axis text under 11 px at 358, as S14 phone)
    for pat in CHART_PARTS:
        html, m = re.subn(pat, '', html, flags=re.S); assert m == k, (pat, m)
    assert 'ip-chart' not in html
    return html
def phone(parts, cls=''):
    body = A.phone(parts)
    for a, z in (('class="ip-tabbar__item is-current" href="#acasa" aria-current="page"', 'class="ip-tabbar__item" href="#acasa"'),
                 ('class="ip-tabbar__item" href="#serie"', 'class="ip-tabbar__item is-current" href="#serie" aria-current="page"'),
                 ('<main class="ip-screen__main s2-main">', '<main class="ip-screen__main s2-main s4-main">'),
                 ('<div class="ip-screen">', f'<div class="ip-screen s4ns-phone{cls}">')):
        assert body.count(a) == 1, a; body = body.replace(a, z, 1)
    assert body.count('is-current') == 1
    return body
def p_table(uid, tab, groups, bar, note, tag, weeks=None, added=None):
    weeks = weeks or P.WEEKS
    cap = f'{tab} · Hîrbovăț · Seria 2026 · SAPT {weeks[-1]}–{weeks[0]} · {CHIPW[tag]}'
    t, tid = P.ptable(uid, cap, [('Total fermă', '', groups)], spine='phone', weeks=weeks, more=P.more(added=added, focus=bool(added)))   # SYS18 (D192): the 48 px load-more row
    assert t.count('aria-expanded="true"') == len(weeks)                          # D166: every week open
    return (f'<section aria-labelledby="{uid}-h"><h2 class="ip-sr" id="{uid}-h">{tab}: tabel pe săptămâni și zile</h2>'
            + P.block(uid, cap, t, bar=bar, touch=True, after=s4.ptnote(note)) + '</section>')

# Rezumat · telefon: În viață + Morți (tag Mortalitate)
RZ_P = [g for g in RZ if gname(g) in ('În viață', 'Morți')]
NB_USER = lambda html, e: html.replace(e['user'], e['user'].replace(' ', '\u00a0'))   # "Dr. Ceban" never splits (phone meta)
EV_P = [e for e in R.EVENTS if e['record'] == TAB_TAG['rezumat']][:len(s4.EVENTS)]    # the latest records of the one tag (D160 + spec D1 › Columns)
def alerts_narrow(panel):  # the AlertRow list sets its rows from its own width (additions.css §35: link and lamp on a second line under 480)
    a = '<ul class="ip-list" aria-labelledby="h-chk">'; assert panel.count(a) == 1
    return panel.replace(a, '<ul class="ip-list ip-list--alerts" aria-labelledby="h-chk">')
def phone_rezumat():
    tb = p_table('rz', 'Rezumat', RZ_P, False, 'Morți = morții din registru, sacrificările incluse. O etichetă odată: alege alta din Filtrează.', TAB_TAG['rezumat'])
    panel = s4.panel_default(table=tb); old = ''.join(s4.events); assert panel.count(old) == 1
    panel = alerts_narrow(nochart(panel.replace(old, ''.join(NB_USER(s4.event_row(e), e) for e in EV_P)), 1))
    return phone([p_header(), p_cycle(), f'<div class="s4-tight">{p_filter()}{s4.tabs}</div>', panel])

# Păsări · telefon: Detaliu (touch) + În viață + Mortalitate (tag Mortalitate; the carved culls sit under Sacrificare)
PA = T.pasari()
assert [gname(g) for g in PA] == ['În viață', 'Mortalitate', '% cumulat', 'Sacrificare', 'Eroare sexare', 'Transfer', 'Vânzare']
PA_P = [PA[0], P.stacked(PA[1], PA[2])]          # D171 + S4NS fix 2: În viață + Mortalitate with % cumulat as its cells' second line (4 figure columns)
def phone_pasari(open_=False, weeks=None, added=None):
    T.EVID[0] = 0
    c = T.PCULLS[0]
    tb = p_table('pa', 'Păsări', PA_P, True, f'Golirea din {R.ddm(R.EMPTYING_SINCE)} e Vânzare; Mortalitate = doar morți reali. % cumulat = morți din sosire ÷ păsări sosite. '
                 f'Sacrificarea din {R.ddm(c["date"])} ({c["section"]} {s4.M} {c["qty"]}) e ilustrativă și e scăzută din morții zilei. O etichetă odată: alege alta din Filtrează.', TAG,
                 weeks=weeks, added=added)
    rows = [NB_USER(T.log_row(e), e) for e in T.LOGE if e['record'] == TAG]           # the event log under the table: the one tag's rows (spec D1 › Columns)
    assert len(rows) == 3
    span = R.drange(T.LOG_FROM, R.TODAY)
    log = T.block('h-jurnal', 'Jurnal păsări', f'{CHIPW[TAG]} · cele mai noi',
                  f'<ul class="ip-list ip-list--act" aria-label="Jurnal păsări Hîrbovăț, {CHIPW[TAG]}, {span}">{"".join(rows)}</ul>'
                  + T.foot(f'<span class="ip-num">{len(rows)} înregistrări · {span}</span> · Zilnic: o înregistrare pe zi pentru secțiunile fermei', ('Mai vechi', '#jurnal-pagina-2')))
    panel = nochart(tb + '\n' + T.charts(('h-mort', 'Mortalitate cumulată', T.ro_mort()), ('h-alive', 'În viață', T.ro_alive())) + '\n' + log, 2)
    panel = f'<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-pasari">\n{panel}\n</div>'
    return phone([p_header(), p_cycle(), f'<div class="s4-tight">{p_filter(open_)}{T.tabs("pasari")}</div>', panel])

# Păsări · telefon · filtre: the FilterBar BottomSheet (single-choice tags, D160)
X_ICON = '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>'
def touch_seg(seg, lid, focus=False):
    a = seg[:seg.index('>') + 1]; assert a.startswith('<span class="ip-seg" role="radiogroup"')
    out = f'<span class="ip-seg ip-seg--touch ip-seg--fill" role="radiogroup" aria-labelledby="{lid}">' + seg[len(a):].replace(
        '>Ciclul curent<', '><span aria-hidden="true">Ciclu</span><span class="ip-sr">Ciclul curent</span><')   # SYS20 (D195): equal segments in 358, short word
    return out.replace('<span class="ip-seg__item is-selected"', '<span class="ip-seg__item is-selected ip-focus"', 1) if focus else out
PA_TYPES = [k for k in PTYPES if k in T.BIRD]      # the Păsări tab's types of the period (D45, D191: the cycle): no Hrană / Cântărire / Ouă
assert PA_TYPES == ['mortality', 'cull', 'sexerr', 'transfer', 'sale']
def radio_chips():
    out = ''
    for k in PA_TYPES:
        on = k == TAG
        a = f'role="radio" aria-checked="{"true" if on else "false"}" tabindex="{"0" if on else "-1"}"'
        if k == 'sexerr':   # the neutral split chip (D47), as the desktop FilterBar
            out += (f'<button type="button" class="ip-chip ip-chip--touch ip-rec ip-chip--split{"" if on else " ip-chip--off"}" {a}>'
                    f'<span class="ip-tile ip-tile--split" aria-hidden="true">{icon("sexerr", 16)}</span>{CHIPW[k]}</button>')
            continue
        out += (f'<button type="button" class="ip-chip ip-chip--touch ip-rec ip-rec--{s4.REC[k]}{"" if on else " ip-chip--off"}" {a}>{icon(k, 16)}{CHIPW[k]}</button>')
    return f'<div class="ip-chips ip-chips--wrap ip-chips--fill" role="radiogroup" aria-labelledby="pf-tags" aria-describedby="pf-tags-h">{out}</div>'   # SYS20: 2 equal columns
def p_sheet():
    per = touch_seg(s4.seg('Perioadă', s4.PERIODS, 0), 'pf-per', focus=True)   # D191: Ciclul curent
    pas = touch_seg(SEG_ZI, 'pf-pas')
    tags = (f'<div class="ip-field" role="group" aria-labelledby="pf-tags"><span class="ip-label" id="pf-tags">Tipuri de înregistrări</span>{radio_chips()}'
            f'<span class="ip-hint" id="pf-tags-h">Pe telefon tabelul arată un tip odată, ales pe fiecare filă; În viață rămâne.</span></div>')
    exp = ('<div class="s4ns-fsheet__exp"><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch" aria-describedby="pf-exp">'
           f'{icon("download", 16)}Exportă .xlsx</button>'
           f'<p class="ip-hint" id="pf-exp">Toată perioada aleasă: Hîrbovăț, Seria 2026, ciclul curent (<span class="ip-num">SAPT {s4.CYC_WEEKS[-1]}–{WK}</span>), pe zile, {CHIPW[TAG]}.</p></div>')   # D191: Exportă = the whole period
    return ('<div class="ip-scrim ip-scrim--sheet">'
            '<div class="ip-sheet s4ns-fsheet" role="dialog" aria-modal="true" id="pf-sheet" aria-labelledby="pf-t" aria-describedby="pf-d">'
            '<div class="ip-sheet__handle" aria-hidden="true"></div>'
            '<div class="ip-sheet__head"><div><h2 class="ip-t-heading ip-sheet__title" id="pf-t">Filtrează sau exportă</h2>'
            '<p class="ip-note ip-sheet__sub" id="pf-d">Tabelul se schimbă la fiecare alegere.</p></div>'
            f'<button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Închide filtrele"><svg class="ip-icon" viewBox="0 0 24 24" aria-hidden="true">{X_ICON}</svg></button></div>'
            f'<div class="s4ns-fsheet__body"><div class="ip-field"><span class="ip-label" id="pf-per">Perioadă</span>{per}</div>'
            f'<div class="ip-field"><span class="ip-label" id="pf-pas">Pas</span>{pas}</div>{tags}{exp}</div>'
            '</div></div>')
def phone_sheet():  # 390×844: the page (at its top) and tab bar inert under the scrim, the sheet at the foot
    b = phone_pasari(open_=True)
    for a, z in (('<div class="ip-screen s4ns-phone">', '<div class="ip-screen s4ns-phone ip-screen--fixed ip-scrimhost">'),
                 ('<main class="ip-screen__main s2-main s4-main">', '<div class="ip-screen__view" inert="inert"><main class="ip-screen__main s2-main s4-main">'),
                 ('</main>', '</main></div>'), ('<nav class="ip-tabbar', '<nav inert="inert" class="ip-tabbar')):
        assert b.count(a) == 1, a; b = b.replace(a, z)
    assert b.endswith('</div>')
    return b[:-len('</div>')] + p_sheet() + '</div>'

# ================================================================= CSS (screen only, tokens only)
CSS_D = s4.CSS + '\n' + s4.CSS_STATES
CSS_PHONE = A.CSS_P + '\n' + T.CSS + '''
.s4ns-phone .s4-stats{grid-template-columns:minmax(0,1fr)} /* phone: the Stats one per line (as S14 phone) */
.s4ns-phone .ip-sechead{flex-direction:column;align-items:flex-start;gap:var(--space-1)}
.s4ns-phone .s4t-foot{flex-direction:column;align-items:flex-start;gap:var(--space-2)}
.s4ns-phone .s4t-lk,.s4ns-phone .ip-sechead > .ip-link{min-height:var(--control-h-touch);display:inline-flex;align-items:center} /* touch: 48px target */
.s4ns-phead{flex-wrap:wrap;gap:var(--space-3)}
.s4ns-phead .ip-pagehead__meta{margin-top:var(--space-2)}
.s4ns-phead > .ip-row{flex-wrap:wrap}
.s4ns-phead > .ip-row > .ip-btn{min-height:var(--control-h-touch)}
.s4ns-pfilter{gap:var(--space-2) var(--space-3)}
.s4ns-phone .ip-tab{height:var(--control-h-touch)} /* touch: 48px tabs */
.s4ns-phone .ip-readout__cap{flex-wrap:wrap;row-gap:var(--space-1)}
.s4ns-phone .ip-rlamp{white-space:nowrap} /* the lamp drops under the caption whole, never split */'''
# D167 on a phone board (canvas only): the screen keeps its 390 layout, the table region runs past its right edge in full
CSS_PWIDE = '''
.s4ns-wide > .ip-screen{width:390px}
.s4ns-wide .ip-tablescroll{width:min-content;min-width:100%;max-width:none;overflow:visible}'''
CSS_SHEET = CSS_PHONE + '''
.s4ns-fsheet{display:flex;flex-direction:column;min-height:0;max-height:100%;padding-bottom:0} /* phone FilterBar sheet, as S14 phone */
.s4ns-fsheet__body{flex:1;min-height:0;overflow:auto;display:flex;flex-direction:column;gap:var(--space-4);padding:var(--space-4);border-top:var(--ip-hair) solid var(--rule)}
.s4ns-fsheet__exp{display:flex;flex-direction:column;gap:var(--space-2);padding-top:var(--space-4);border-top:var(--ip-hair) solid var(--rule)}
.s4ns-fsheet__exp > .ip-btn{width:100%}
.s4ns-fsheet__exp > .ip-hint{margin:0}'''

# ================================================================= boards
RZB, PAB = 'S4-ferma-rezumat--', 'S4-ferma--pasari--'
DESK = [(RZB + 'etichete', 'S4 · Fermă · Rezumat · etichete Mortalitate + Vânzare', board_tags),
        (RZB + 'pas-saptamana', 'S4 · Fermă · Rezumat · pas Săpt. (ales în filtre)', board_week),
        (RZB + '4-saptamani', 'S4 · Fermă · Rezumat · perioada 4 săpt. (rândul „Încarcă încă 4 săptămâni”)', board_4),
        (RZB + '8-saptamani', 'S4 · Fermă · Rezumat · 8 săpt. (după o apăsare pe „Încarcă încă 4 săptămâni”)', board_8),
        (RZB + 'rand-nou', 'S4 · Fermă · Rezumat · rând nou (live, D197): vânzarea de la 13:16 sosește, nuanța la început, „Nou”', board_live)]
PHONE = [(RZB + 'telefon', 'S4 · Fermă · Rezumat · telefon', phone_rezumat, False),
         (PAB + 'telefon', 'S4 · Fermă · Păsări · telefon', phone_pasari, False),
         (PAB + 'telefon-filtre', 'S4 · Fermă · Păsări · telefon · filtre', phone_sheet, True),
         (PAB + 'telefon--8-saptamani', 'S4 · Fermă · Păsări · telefon · 8 săptămâni (după „Încarcă încă 4 săptămâni”)', lambda: phone_pasari(weeks=W8, added=W8[4:]), False)]
out = []
for stem, title, fn in DESK:
    h = HEIGHTS.get(stem, 1024)
    s4.write(stem, title, fn(), H=h, css=CSS_D); out.append((stem, title, h, 1440))
import canvas_rows as CR
for stem, title, fn, fixed in PHONE:
    A.HEIGHTS[stem] = h = HEIGHTS.get(stem, 844); W = CR.width(stem, 390)
    o = A.write_phone(stem, title, fn(), css=(CSS_SHEET if fixed else CSS_PHONE) + (CSS_PWIDE if W > 390 else ''), fixed=fixed)
    if W > 390:  # D167: board_widths.json gives the phone board the width its table needs (page 390, table drawn whole)
        a = f'<div class="ip-root" style="padding:0; width:390px; '
        for f in (ROOT / f'design/canvas/project/{stem}.dc.html', pathlib.Path(f'/tmp/s2/{stem}.html')):
            t = f.read_text(encoding='utf-8'); assert t.count(a) == 1
            t = t.replace(a, f'<div class="ip-root s4ns-wide" style="padding:0; width:{W}px; ')
            if f.suffix == '.html' and f.name.endswith('.dc.html'):
                k = f'"$preview":{{"width":390,"height":{h}}}'; assert t.count(k) == 1; t = t.replace(k, f'"$preview":{{"width":{W},"height":{h}}}')
            f.write_text(t, encoding='utf-8')
        o = (stem, title, h, W)
    out.append(o)
json.dump([{'stem': s, 'w': w, 'h': h} for s, _, h, w in out], open('/tmp/s4s/s4ns.json', 'w'))

if CANVAS:  # S4 row: after its last board (x step w + 80; wide boards from board_widths.json), desktops then phones
    c = CR.load(ROOT)
    mine = {f'{s}.dc.html' for s, _, _, _ in out}
    row = [v for k, v in c['boards'].items() if k not in mine and v['y'] == CR.row_y(c, 'S4-ferma-rezumat')]
    x0 = min((c['boards'][f]['x'] for f in mine if f in c['boards']), default=max(v['x'] + v['w'] for v in row) + CR.GAP_X)
    CR.place(c, 'S4-ferma-rezumat', [(s, w, h, t) for s, t, h, w in out], x0)
    for s, _, _, w in out:  # a desktop board drawn wider (D167) is still a PAGE board: "expand":"fill", as the other wide S4 boards
        e = c['boards'][f'{s}.dc.html']
        if w >= 1440 and 'expand' not in e: c['boards'][f'{s}.dc.html'] = {'x': e['x'], 'y': e['y'], 'w': e['w'], 'h': e['h'], 'expand': 'fill', 'title': e['title']}
    CR.save(ROOT, c)
