"""S2 · Acasă (spec D2): the role's start page on today = 18.11.2026 (D152), Seria 2026 (RF06: realdata only).
- vet · telefon: header (farm, today, SAPT) · Lipsesc date (MissingRow, C9) · Înregistrează (EntryKeys, 3) · Introdus azi
  (EventRow actions, C8, one row per record type saved today).
- manager · telefon / desktop: header · De verificat (AlertRow) · stoc hrană (StockReadout per farm × type) ·
  Înregistrează (EntryKeys, 6) · Lipsesc date (farm named first, several farms).
- proprietar · desktop: compact CycleTrack for Seria 2026 · De verificat across farms · Livrări fără preț (UnpricedCount) ·
  stoc hrană (StockReadout) · ferme (FarmSummary --compact).
Shells are reused by import: the phone page uses the S7 phone screen/main classes (s7s_build → s7_build helpers and CSS
lines) with the system BottomTabBar; the desktop uses the S4 shell (s4s_build: Sidebar with Acasă current, PageHeader
type, AlertRows from s4s_build.alerts(), board writer). Blocks are the SYS08 components (MissingRow, EntryKeys,
UnpricedCount), SYS01 AlertRow, SYS03 FarmSummary compact, SYS04 StockReadout, SYS09 EventRow actions.
Every figure comes from realdata.py (the workbook to 18.11.2026; event times/authors from the illustrative log): the vet is
Dr. Ceban, the manager Ion Rusu (illustrative). The page is drawn after 12:00 (MissingRow lists today's rows only from
12:00, C9) and after the manager's 13:16 sale (Introdus azi).
python3 -I s2_build.py <repo_root> [heights.json] [--canvas]
Writes design/canvas/project/S2-acasa[--<role>].dc.html, /tmp/s2/<stem>.html (phones) and /tmp/s4s/<stem>.html (desktops);
--canvas also sets the S2 boards in design/canvas/project/canvas.json (new row below the lowest one, phones first, D74)."""
import sys, re, json, pathlib, importlib, datetime as dt

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
ROOT, HP, CANVAS = CR.args('s2_heights.json')  # heights default: s2_heights.json
HERE = pathlib.Path(__file__).resolve().parent

sys.path.insert(0, str(HERE))
sys.argv = ['s7s_build.py', str(ROOT)]
P = importlib.import_module('s7s_build'); B = P.B          # phone helpers (rewrites the S7 boards unchanged)
sys.argv = ['s4s_build.py', str(ROOT)]
s4 = importlib.import_module('s4s_build')                    # desktop shell (writes nothing on import)

n, dd, NB, MINUS = s4.n, B.dd, s4.NB, s4.MINUS
NB2 = '\u00a0'  # no-break space before a · separator
lamp, pair, hen, roo, F, M = s4.lamp, s4.pair, s4.hen, s4.roo, s4.F, s4.M
IC = dict(s4.IC)
IC.update({
 'go': '<path d="m9 18 6-6-6-6"/>',                                                                   # MissingRow chevron-right
 'dots': '<circle cx="12" cy="5" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="12" cy="19" r="1"/>',  # Acțiuni (ellipsis-vertical)
 'more': '<circle cx="5" cy="12" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/>',  # BottomTabBar Mai mult
 'user': '<circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/>',                               # BottomTabBar Cont
 'placement': '<path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><path d="m10 17 5-5-5-5"/><path d="M15 12H3"/>',
})
REC = dict(s4.REC); REC.update({'placement': 'placement'})
def icon(k, size=None):
    return f'<svg class="ip-icon{f" ip-icon--{size}" if size else ""}" viewBox="0 0 24 24" aria-hidden="true">{IC[k]}</svg>'
def tile(k): return f'<span class="ip-tile ip-rec ip-rec--{REC[k]}" aria-hidden="true">{icon(k)}</span>'
def ico(k): return f'<span class="ip-rec ip-rec--{REC[k]}">{icon(k, 14)}</span>'
def num(v): return f'<span class="ip-num">{v}</span>'
def nw(v): return f'<span class="ip-nowrap">{v}</span>'  # S2P: a meta group that never splits (lines break at the · separators)

# ---------- data (realdata only, RF06; today 18.11.2026, D152) ----------
import realdata as R
dd = R.dd
TODAY = R.TODAY; YDAY = TODAY - dt.timedelta(days=1); D = dd(TODAY)
cT, cH = R.CTEL, R.C26                                   # Telița 07.10.2025–28.02.2026 (closed) · Hîrbovăț from 22.02.2026 (open)
WK = R.WK                                                # bird age week (D151)
assert WK == R.sapt(TODAY) == cH['sapt_today'] and cT['status'] == 'closed' and (cT['sapt_start'], cT['sapt_end']) == (1, 21)
TL = R.COUNTS_TELITA
ALIVE = {sx: R.alive_farm(sx) for sx in 'FM'}; ALIVE['all'] = ALIVE['F'] + ALIVE['M']
assert (ALIVE['F'], ALIVE['M']) == (R.COUNTS['totals']['F'], R.COUNTS['totals']['M']) and TL['totals']['all'] == 0
alive = {(s, sx): R.alive(s, sx) for s in R.HSECS for sx in 'FM'}
FTN = {k: v['name'] for k, v in R.FEED_TYPES.items()}
GUIDES = {k: v['name'] for k, v in R.GUIDES.items()}
TR = R.TR; PL = R.PLACEMENT_ROWS
XD = sorted({a['date'] for a in TR['arrivals']} | {o['date'] for o in TR['transfers_out']})
X0D, X1D = XD[0], XD[-1]; assert X0D == cH['start_date'] and X1D == cT['end_date']

# C9 · Lipsesc date: compartments with birds and no Hrană zilnică / Zilnic row, today (from 12:00) and the 7 days before.
# A day "has birds" when the compartment holds birds at the start and at the end of it: the sale day that empties it is no
# lipsă (as S11 Consum, RF03). The workbook has no kg for H7 ♀ on 13.11 (H7.1, H7.2): two past-day rows (D65: the vet sees
# them locked, the manager opens the form). Zilnic: every day has its mortality and egg row in the workbook.
W7 = [TODAY - dt.timedelta(days=k) for k in range(7, -1, -1)]
def fed(d, sec, sx): return R.fed(d, sec, sx)
MISS = R.missing('hirbovat', TODAY)   # D181: the one C9 window (realdata), also the owner's/viewer's and S4 Rezumat's lipsă alerts
assert [(R.ddm(d), s, sx) for d, s, sx in MISS] == [('13.11', 'H7.1', 'F'), ('13.11', 'H7.2', 'F')], MISS

# C8 · Introdus azi (Dr. Ceban, today): one row per record type saved today (illustrative event log on the real figures)
LOG = {e['record']: e for e in R.EVENTS if e['at'][:10] == TODAY.isoformat() and e['user'] == 'Dr. Ceban'}
assert sorted(LOG) == ['eggs', 'feed', 'mortality']                         # no weighing at Hîrbovăț today (S8, RF07)
DTH = {sx: R.deaths(TODAY, sx=sx) for sx in 'FM'}                            # real deaths (D155); the sale is a Vânzare record
ZSEC = [s for s in R.HSECS if R.sd(TODAY, s, 'F') or R.sd(TODAY, s, 'M')]    # compartments in today's Zilnic
CULLS = [c for c in R.ILL['culls'] if c['date'] == TODAY.isoformat()]
assert f'{len(ZSEC)} secțiuni · mortalitate ♀ {DTH["F"]} · ♂ {DTH["M"]}' in LOG['mortality']['summary'] and not CULLS
EGG_HALLS = [h for h in R.HHALLS if R.eggs(TODAY, h)]
EGG_T = R.eggs(); assert LOG['eggs']['summary'].startswith(f'Ouă · {EGG_T} total') and LOG['eggs']['scope'].endswith(', '.join(EGG_HALLS))
DUE = [(s, sx) for s in R.HSECS for sx in 'FM' if R.alive_start(s, sx)]
FEED_ROWS = [(s, sx) for s, sx in DUE if fed(TODAY, s, sx)]
FEED_KGS = {sx: R.feed_kg(TODAY, sx) for sx in 'FM'}; FEED_KG = round(FEED_KGS['F'] + FEED_KGS['M'], 1)
assert LOG['feed']['summary'] == f'Hrană · ♀ {FEED_KGS["F"]:g} kg · ♂ {FEED_KGS["M"]:g} kg'

# stock (SYS04: days left on the mean of the last 3 days, realdata.stock_days) and unpriced deliveries (SYS08)
W3 = (TODAY - dt.timedelta(days=2), TODAY)
B2, CO = s4.stocks['ft-breeder2'], s4.stocks['ft-cocosi']
assert all(abs(st['days_left'] - R.stock_days(sx)[0]) < 1e-9 for st, sx in ((B2, 'F'), (CO, 'M')))
UNP = R.UNPRICED
UNP_KG = sum(d['qty_kg'] for d in UNP); assert (len(UNP), UNP_KG) == (len(s4.unpriced), s4.unpriced_kg)
LAY, LAYG = s4.lay_today, s4.lay_guide; MH = s4.cum['mortality_pct']; MORT_G = s4.MORT_G   # D156 lay · D155 deaths · D161 guide ♀
assert abs(LAY - R.lay_day()) < 1e-9 and LAYG == R.guide_lay(WK)

# ---------- blocks ----------
def head(cap, note, hid):
    return f'<div class="ip-list__head"><h2 class="ip-cap" id="{hid}">{cap}</h2><p class="ip-note">{note}</p></div>'

# FX02: on desktop a page block is the system's section.ip-block > div.ip-sechead > h2.ip-t-heading (16/22) + its note, as
# every other desktop page (S3, S4, S9–S22); phones keep the caption heads above. The blocks are drawn once (phone markup)
# and desk_heads() rewrites them for the desktop page: s2-block blocks, MissingRow and EntryKeys. UnpricedCount keeps its
# caption: that head is inside its panel (RecordTile + caption), as a Readout caption, not a page block head.
HEAD_RE = re.compile(r'<div class="(?:ip-list__head|ip-missing__head)"><h2 class="ip-cap"( id="[^"]+")>(.*?)</h2><p class="ip-note">(.*?)</p></div>')
EK_RE = re.compile(r'<section class="ip-entrykeys"( aria-labelledby="[^"]+")><h2 class="ip-cap"( id="[^"]+")>(.*?)</h2>')
def desk_heads(html):
    html = HEAD_RE.sub(r'<div class="ip-sechead"><h2 class="ip-t-heading"\1>\2</h2><span class="ip-note">\3</span></div>', html)
    html = EK_RE.sub(r'<section class="ip-entrykeys ip-block"\1><div class="ip-sechead"><h2 class="ip-t-heading"\2>\3</h2></div>', html)
    html = html.replace('<section class="s2-block"', '<section class="ip-block"').replace('<section class="ip-missing"', '<section class="ip-missing ip-block"')
    assert not re.search(r'ip-list__head|ip-missing__head|s2-block', html), 'a caption block head left on desktop'
    assert re.findall(r'<h2 class="ip-cap"[^>]*>', html) in ([], ['<h2 class="ip-cap" id="up">']), 'only the UnpricedCount panel keeps a caption'
    return html

def missing(manager=False, u='', miss=None, past=(), locked=None, when=None, empty=None, scope=None):
    """MissingRow (SYS08): one row per section × entry × day, newest first; manager rows name the farm first.
    locked: a past day is a locked row for the vet (D07/D65: Tag "blocat", no link); default = not manager.
    S2S: past = extra past-day rows, when = the scope note's time part (before 12:00: the clock time; today's rows are
    not listed yet, C9), empty = the sentence of the empty state."""
    locked = (not manager) if locked is None else locked
    rows = []
    src = [m for m in list(MISS if miss is None else miss) + list(past) if not (when and m[0] == TODAY.isoformat())]
    for day, sec, sx in sorted(src, key=lambda m: (dt.date.fromisoformat(m[0]).toordinal() * -1, m[1])):
        c, sign, word = ('hen', '♀', 'găini') if sx == 'F' else ('rooster', '♂', 'cocoși')
        d0 = dt.date.fromisoformat(day); wk = R.sapt(d0)
        alv, gy = R.alive(sec, sx, d0), R.feed_g(d0 - dt.timedelta(days=1), sx, sec)   # alive at the end of that day, grams of the day before
        farm = 'Hîrbovăț · ' if manager else ''
        sexl = f'<span class="ip-sex ip-sex--{c}">{sign} <b>{n(alv)}</b></span>'
        tl = f'<span class="ip-tile ip-rec ip-rec--feed" aria-hidden="true">{icon("feed")}</span>'
        typ = f'<span class="ip-ev__type">Hrană zilnică <span class="ip-sex ip-sex--{c}">{sign}</span><span class="ip-sr"> {word}</span></span>'
        if locked and day != TODAY.isoformat():
            rows.append(f'<div class="ip-ev ip-ev--missing is-locked">{tl}<span class="ip-ev__main"><span class="ip-ev__title">{farm}{sec} · {typ}<span class="ip-tag ip-tag--neutral">blocat</span></span>'
                        f'<span class="ip-ev__meta ip-num">{dd(d0)} · {nw("SAPT " + str(wk))} · {nw("zi trecută")} · {nw("o completează managerul")}</span></span>'
                        f'<span class="ip-ev__status">{lamp("warn", "lipsă")}</span><span class="ip-ev__go"></span></div>')
            continue
        rows.append(f'<a class="ip-ev ip-ev--missing" href="#hrana-zilnica-{sec}-{sx}{"" if day == TODAY.isoformat() else "-" + day}">{tl}'
                    f'<span class="ip-ev__main"><span class="ip-ev__title"><span class="ip-sr">Completează: </span>{farm}{sec} · {typ}</span>'
                    f'<span class="ip-ev__meta ip-num">{dd(d0)} · {nw("SAPT " + str(wk))} · {nw(sexl + " în viață")} · {nw(f"ieri {gy}{NB}g")}</span></span>'
                    f'<span class="ip-ev__status">{lamp("warn", "lipsă")}</span><span class="ip-ev__go">{icon("go", 16)}</span></a>')
    if empty: rows = [f'<p class="ip-missing__empty"><span class="ip-lamp ip-lamp--off">nimic lipsă</span><span>{empty}</span></p>']
    scope = scope or ('fermele tale' if manager else 'Hîrbovăț')
    w = num(when) if when else f'azi de la {num("12:00")}'
    return (f'<section class="ip-missing" aria-labelledby="{u}ms"><div class="ip-missing__head"><h2 class="ip-cap" id="{u}ms">Lipsesc date</h2>'
            f'<p class="ip-note">{scope} · {w} · ultimele {num(7)} zile</p></div><div class="ip-list">{"".join(rows)}</div></section>')

KEYS = [('mortality', 'Zilnic', 'Mortalitate, sacrificare, ouă', '#zilnic'), ('feed', 'Hrană zilnică', 'Grame pe găină și cocoș', '#hrana-zilnica'),
        ('weighing', 'Cântărire', 'Greutate, CV, uniformitate', '#cantarire'), ('transfer', 'Transfer', 'Între secțiuni sau ferme', '#transfer'),
        ('sale', 'Vânzare', 'Păsări vândute pe secțiuni', '#vanzare'), ('delivery', 'Livrare hrană', 'Kg pe tip de hrană', '#livrare-hrana')]
def entrykeys(manager=False, u=''):
    ks = KEYS if manager else KEYS[:3]
    li = ''.join(f'<li><a class="ip-btn ip-btn--secondary ip-entrykey" href="{h}">{tile(k)}<span class="ip-entrykey__txt"><span>{lb}</span><span class="ip-note">{hint}</span></span></a></li>'
                 for k, lb, hint, h in ks)
    return f'<section class="ip-entrykeys" aria-labelledby="{u}ek"><h2 class="ip-cap" id="{u}ek">Înregistrează</h2><ul class="ip-entrykeys__list">{li}</ul></section>'

def act_row(k, i, title, meta, name, menu='closed', off=False, edited=None):
    """SYS09 EventRow with actions: own entry today → Editează (the form of today) + Șterge (Dialog); phone = Acțiuni menu.
    S2S: menu='open' draws the phone menu open (first item active and focused, ip-focus); off = the 23:59 window closed while the page was open
    (aria-disabled keys and items, the reason in the menu note); edited = the author's same-day edit time (SYS09 A)."""
    mid = f'a{i}-m'
    dis = f' aria-disabled="true" aria-describedby="{mid}-n"' if off else ''
    if edited: meta += f'{NB2}· <span class="ip-ev__edited">modificat la {num(edited)}</span>'
    ed_k = f'<a class="ip-btn ip-btn--ghost" role="link" tabindex="0"{dis}>' if off else f'<a class="ip-btn ip-btn--ghost" href="#editeaza-{k}">'
    fo = ' ip-focus' if menu == 'open' else ''  # the open menu shows its focused first item (SYS09 overrides .is-active's fill)
    ed_i = f'<a class="ip-menu__item{fo}" role="menuitem" tabindex="-1"{dis}>' if off else \
           f'<a class="ip-menu__item{" is-active" if menu == "open" else ""}{fo}" role="menuitem" tabindex="-1" href="#editeaza-{k}">'
    note = (f'Timpul de modificare s-a încheiat la {num("23:59")}. Managerul poate corecta.' if off else
            f'Poți modifica sau șterge până la {num("23:59")}.')
    shut = menu == 'closed'
    return (f'<li class="ip-ev ip-ev--act">{tile(k)}<span class="ip-ev__main"><span class="ip-ev__title">{title}</span><span class="ip-ev__meta">{meta}</span></span>'
            f'<span class="ip-ev__acts">{ed_k}Editează<span class="ip-sr"> {name}</span></a>'
            f'<button type="button" class="ip-btn ip-btn--ghost" aria-haspopup="dialog"{dis}>Șterge<span class="ip-sr"> {name}</span></button></span>'
            f'<span class="ip-ev__more"><button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Acțiuni, {name}" aria-haspopup="menu" aria-expanded="{"false" if shut else "true"}" aria-controls="{mid}">{icon("dots")}</button>'
            f'<div class="ip-menu ip-ev__menu"{" hidden=\"hidden\"" if shut else ""}><div class="ip-ev__items" id="{mid}" role="menu" aria-label="Acțiuni, {name}" aria-describedby="{mid}-n">'
            f'{ed_i}Editează</a><button type="button" class="ip-menu__item" role="menuitem" tabindex="-1"{dis}>Șterge</button></div>'
            f'<p class="ip-note" id="{mid}-n">{note}</p></div></span></li>')
def t_of(e): return e['at'][11:16]
INTRO = [k for k in ('mortality', 'eggs', 'feed', 'weighing') if k in LOG]  # the vet's saves of today, in time order
assert [LOG[k]['at'] for k in INTRO] == sorted(LOG[k]['at'] for k in INTRO)
def kgp(f, m): return pair(n(f, 1), n(m, 1), NB + 'kg')
def intro_spec(e):
    """(title, meta, accessible name) of one of today's records (C8), figures from realdata."""
    k, who = e['record'], e['user']
    if k == 'mortality':
        why = 'Zilnic, fără sacrificări' if not CULLS else f'Zilnic, {num(len(CULLS))} sacrificări'
        return (f'Mortalitate · {num(len(ZSEC))} secțiuni · {hen(n(-DTH["F"]))} · {roo(n(-DTH["M"]))}', f'{num(t_of(e))} · {who} · {why}', 'Mortalitate de azi')
    if k == 'eggs':
        return (f'Ouă · {", ".join(EGG_HALLS)} · {num(n(EGG_T))} ouă', f'{num(t_of(e))} · {who} · ouat {num(s4.pct(LAY))}', 'Ouă de azi')
    if k == 'feed':
        lack = [f'{(F if sx == "F" else M)} {s}' for s, sx in DUE if (s, sx) not in FEED_ROWS]
        return (f'Hrană zilnică · {num(len(FEED_ROWS))} din {num(len(DUE))} rânduri · {num(n(FEED_KG, 1) + NB + "kg")}',
                f'{num(t_of(e))} · {who} · ' + (f'fără {", ".join(lack)}' if lack else kgp(FEED_KGS['F'], FEED_KGS['M'])), 'Hrană zilnică de azi')
    if k == 'sale':  # DATA04 (D168): one sale form = one record; the farm's birds after it
        return (f'Vânzare · {R.sale_codes(e)} · {s4.sexpair(e["F"], e["M"])}', f'{num(t_of(e))} · {who} · în viață {pair(n(ALIVE["F"]), n(ALIVE["M"]))}', 'Vânzarea de azi')
    raise ValueError(e)
def introdus(menu=None, off=False, edited=None, recs=None):
    """menu / edited: {record: 'open' | 'hh:mm'} for one row (S2S); off: the 23:59 window closed (every row);
    recs: the records listed (default: the vet's of today)."""
    menu, edited = menu or {}, edited or {}
    recs = recs if recs is not None else [LOG[k] for k in INTRO]
    rows = [act_row(e['record'], i, *intro_spec(e), menu=menu.get(e['record'], 'closed'), off=off, edited=edited.get(e['record']))
            for i, e in enumerate(recs, 1)]
    note = (f"{num(D)} · timpul de modificare s-a încheiat la {num('23:59')}" if off else f"{num(D)} · poți modifica sau șterge până la {num('23:59')}")
    return (f'<section class="s2-block" aria-labelledby="ia">{head("Introdus azi", note, "ia")}'
            f'<ul class="ip-list ip-list--act" aria-labelledby="ia">{"".join(rows)}</ul>'
            f'<a class="ip-link s2-more" href="#introdus-azi">Vezi pe secțiuni</a></section>')

# C8 · Introdus azi for the manager (D87): Ion Rusu's records of today. On 18.11 he saved the sale of H6.1 + H6.2 at 13:16
# (one form = one record, D168; real figures, illustrative time), so the block lists it (the page is drawn after it).
MGR_LOG = sorted((e for e in R.EVENTS if e['at'][:10] == TODAY.isoformat() and e['user'] == 'Ion Rusu'), key=lambda e: e['at'])
assert [e['record'] for e in MGR_LOG] == ['sale'] and (MGR_LOG[0]['F'], MGR_LOG[0]['M']) == (R.sold(TODAY, sx='F'), R.sold(TODAY, sx='M'))
def introdus_mgr(**kw): return introdus(recs=MGR_LOG, **kw)
def introdus_empty(line, u=''):
    """Introdus azi with nothing saved today: the list frame keeps its place, off Lamp + word + one sentence (SYS09 li.ip-list__empty)."""
    return (f'<section class="s2-block" aria-labelledby="{u}ia">{head("Introdus azi", num(D), u + "ia")}'
            f'<ul class="ip-list ip-list--act" aria-labelledby="{u}ia"><li class="ip-list__empty"><span class="ip-lamp ip-lamp--off">nimic introdus</span>'
            f'<span>{line}</span></li></ul></section>')
INTRO_MGR = f'Nu ai salvat nimic azi. Ce salvezi azi apare aici, pe tipuri; poți modifica sau șterge până la {num("23:59")}.'  # S15P: the S15 gol sentence (D98)

FARM_P = '<div class="ip-ev__main">Hîrbovăț · '
LIPSA = lamp('warn', 'lipsă')
def alerts(owner=False, u='', extra=()):
    """De verificat = the farms' one alert set (realdata.alerts → s4s_build.alert_rows, D181: the S4 Rezumat rows, same
    wording), farm named first. The manager's lipsă rows have their own block (Lipsesc date); the owner/viewer have none, so
    they are alerts (every day of the C9 window, newest first). "fără preț" has its own block for the owner (UnpricedCount);
    the manager can't add prices (D05), so it is not his alert. extra: alerts appended (S2S state examples)."""
    items = s4.alert_rows('owner', kinds=('lay', 'feed', 'missing', 'stock') if owner else ('lay', 'feed', 'stock'))
    stk = [a for a in items if 'Deschide stocul' in a]
    items += list(extra)
    items = [a.replace('<div class="ip-ev__main">', FARM_P, 1) for a in items]
    LIM = 'limită <span class="ip-num">5 zile</span>'
    items = [a.replace(' · ' + LIM, ' · ' + nw(LIM), 1) for a in items]
    assert sum(LIM in a for a in items) == len(stk)
    # the meta breaks only at its · separators (each group stays whole, as the MissingRow meta)
    items = [re.sub(r'(<div class="ip-ev__meta ip-num">)(.*?)(</div>)', lambda m: m[1] + ' · '.join(nw(g) for g in m[2].split(' · ')) + m[3], a, count=1) for a in items]
    assert all(a.count(FARM_P) == 1 for a in items)
    note = f'{num(len(items))} {"alertă" if len(items) == 1 else "alerte"} · {"toate fermele" if owner else "fermele tale"} · {num(D)}'
    return f'<section class="s2-block" aria-labelledby="{u}dv">{head("De verificat", note, u + "dv")}<ul class="ip-list ip-list--alerts" aria-labelledby="{u}dv">{"".join(items)}</ul></section>'  # FIN16: ul/li, as the S4 Rezumat
N_AL = {o: alerts(o).count('<li class="ip-ev ip-ev--alert">') for o in (False, True)}   # manager 2 (lay H5, H6) · owner 4 (+ 13.11 H7 ♀ lipsă)

def stock(u='', cap='Stoc hrană'):  # D181: one title for every role (was 'Zile de hrană' for the manager)
    def ro(fid, s, sign, word):
        st = 'ok' if s['days_left'] >= 5 else 'warn'
        w0, w1 = W3
        sub = '' if st == 'ok' else '<p class="ip-readout__sub ip-m0">sub pragul de 5 zile</p>'
        rid = f'{u}sr-{fid}'
        return (f'<li><div class="ip-readout ip-readout--stock" role="group" aria-labelledby="{rid}"><div class="ip-readout__cap"><span id="{rid}">Hîrbovăț · {FTN[fid]}'
                f'<span class="ip-sr">, pentru {word}</span></span><span class="ip-rlamp ip-rlamp--{st}">{"în normă" if st == "ok" else "abatere"}</span></div>'
                f'<div class="ip-readout__val">{n(s["stock_kg"])}<small> kg</small></div><dl class="ip-readout__kv"><div><dt>zile rămase</dt><dd>{n(s["days_left"], 1)}<small> zile</small></dd></div>'
                f'<div><dt>medie/zi {R.drange(w0, w1, False)}</dt><dd>{n(s["avg_daily_kg_last3"])}<small> kg</small></dd></div>'
                f'<div><dt>se termină</dt><dd>{dd(s["run_out_date"])}</dd></div></dl>{sub}</div></li>')
    tel = (f'<li><div class="ip-readout ip-readout--stock ip-readout--off" role="group" aria-labelledby="{u}sr-te"><div class="ip-readout__cap"><span id="{u}sr-te">Telița</span>'
           f'<span class="ip-rlamp ip-rlamp--off">oprit</span></div><div class="ip-readout__val">–</div><p class="ip-readout__sub ip-m0">fără ciclu deschis</p>'
           f'<p class="ip-readout__sub ip-m0">ultimul ciclu: Seria 2026, închis pe <span class="ip-num ip-nowrap">{dd(cT["end_date"])}</span></p></div></li>')
    w0, w1 = W3
    return (f'<section class="s2-block" aria-labelledby="{u}st">{head(cap, f"stoc și zile rămase · {num(D)}", u + "st")}'
            f'<ul class="ip-readouts" aria-labelledby="{u}st">{ro("ft-breeder2", B2, "♀", "găini")}{ro("ft-cocosi", CO, "♂", "cocoși")}{tel}</ul>'
            f'<p class="ip-note ip-m0">Zile rămase = stoc ÷ consumul mediu din ultimele {num(3)} zile ({num(R.drange(w0, w1))}); sub 5 zile = abatere.</p></section>')

def unpriced():
    by = {}
    for d in UNP: by.setdefault(d['feed_type_id'], []).append(d)
    li = ''.join(f'<li><span>Hîrbovăț · {FTN[t].replace(" ", "&nbsp;")}</span><span>{num(len(ds))} {"livrare" if len(ds) == 1 else "livrări"} · {num(n(sum(d["qty_kg"] for d in ds)) + " kg")}</span></li>'
                 for t, ds in by.items())
    assert list(by) == ['ft-breeder2', 'ft-cocosi'] and sum(d['qty_kg'] for ds in by.values() for d in ds) == UNP_KG  # ♀ 11 000 · ♂ 7 160 kg
    old = min(d['date'] for d in UNP)
    return (f'<section class="ip-unpriced" aria-labelledby="up"><div class="ip-unpriced__head">{tile("price")}<h2 class="ip-cap" id="up">Livrări fără preț</h2></div>'
            f'<p class="ip-unpriced__fig"><span class="ip-unpriced__val">{len(UNP)}<small> livrări</small></span><span class="ip-unpriced__kg">{n(UNP_KG)} kg</span></p>'
            f'<ul class="ip-unpriced__list">{li}</ul><p class="ip-note">Cea mai veche din {num(dd(old))}. Kg fără preț nu intră în cost.</p>'
            f'<a class="ip-btn ip-btn--secondary" href="#rapoarte-hrana-livrari-fara-pret">Adaugă prețurile</a></section>')

def farms_block(phone=False):  # FY03: phone = ip-fsum--touch (48px links, additions.css §9)
    P_ = '<small class="ip-fsum__pct">%</small>'; T_ = ' ip-fsum--touch' if phone else ''
    lay_ok = abs(LAY - LAYG) <= 3
    hir = (f'<article class="ip-fsum ip-fsum--compact{T_}" aria-labelledby="fs-hir" aria-describedby="fs-hir-c"><div class="ip-fsum__head"><div class="ip-fsum__id">'
           f'<h3 class="ip-fsum__name" id="fs-hir"><a href="#/ferme/hirbovat?ciclu=cyc-hirbovat-2026">Hîrbovăț</a></h3><span class="ip-tag ip-tag--neutral">Ouat</span>'
           f'<span class="ip-tag ip-tag--neutral">{GUIDES["g-laying"]}</span></div>{lamp("ok", "deschis")}</div>'
           f'<p class="ip-fsum__cycle" id="fs-hir-c">ciclu din {dd(cH["start_date"])} · SAPT {WK}</p>'
           f'<p class="ip-fsum__alive"><span class="ip-cap">În viață</span>{pair(n(ALIVE["F"]), n(ALIVE["M"]))}</p><dl class="ip-fsum__figs">'
           f'<div class="ip-fsum__fig"><dt>{ico("eggs")}Ouat azi</dt><dd class="ip-fsum__val">{n(LAY, 1)}{P_}</dd><dd class="ip-fsum__sub"><span class="ip-num">ghid {s4.pct(LAYG)} · SAPT {WK}</span></dd>'
           f'<dd class="ip-fsum__status">{lamp("ok" if lay_ok else "warn", "în normă" if lay_ok else "abatere")}</dd></div>'
           f'<div class="ip-fsum__fig"><dt>{ico("mortality")}Mort. cumulată</dt><dd class="ip-fsum__val">{pair(n(MH["F"], 2), n(MH["M"], 2))}{P_}</dd>'
           f'<dd class="ip-fsum__sub"><span class="ip-num">ghid {hen(s4.pct(MORT_G, 2))}</span></dd>'   # D161: ♀ vs the guide's cumulative %, deaths only (D155)
           f'<dd class="ip-fsum__status">{lamp("ok", "în normă") if MH["F"] <= MORT_G else lamp("warn", "abatere")}</dd></div></dl></article>')
    tel = (f'<article class="ip-fsum ip-fsum--compact{T_}" aria-labelledby="fs-tel" aria-describedby="fs-tel-c"><div class="ip-fsum__head"><div class="ip-fsum__id">'
           f'<h3 class="ip-fsum__name" id="fs-tel"><a href="#/ferme/telita">Telița</a></h3><span class="ip-tag ip-tag--neutral">Creștere</span>'
           f'<span class="ip-tag ip-tag--neutral">{GUIDES["g-growth"]}</span></div>{lamp("off", "goală")}</div>'
           f'<div class="ip-fsum__empty" id="fs-tel-c"><p class="ip-m0">Ferma e goală din {num(dd(cT["end_date"]))}. Ultimul ciclu: Seria 2026, {num(cT["weeks"])} săpt.</p>'
           f'<a class="ip-link" href="#/ferme/telita?ciclu=cyc-telita-2026">Vezi ultimul ciclu</a></div></article>')
    return (f'<section class="s2-block" aria-labelledby="fm">{head("Ferme", f"{num(2)} ferme · {num(D)}", "fm")}'
            f'<div class="ip-fsums s2-fsums">{hir}{tel}</div></section>')

# ---------- compact CycleTrack (owner): Seria 2026, one lane per farm ----------
DOM0 = R.PLACEMENT.replace(day=1)                     # axis from the series' placement month (10.2025) …
DOM1 = dt.date(TODAY.year + 1, 1, 1); DOM1_P = dt.date(TODAY.year + 1, 3, 1)   # … to the end of this year (phone: + 2 months, S2V)
LX, RO_W, TWS = 72, 56, (1160, 1000)  # drawn 1:1 for the 1160 px content box at 1440 and the 1000 px one at 1280 (as S3P)
def months(d0, d1):
    out = []; y, m = d0.year, d0.month
    while dt.date(y, m, 1) < d1:
        out.append((dt.date(y, m, 1), f'{m:02d}.{str(y)[2:]}')); y, m = (y + 1, 1) if m == 12 else (y, m + 1)
    return out
def track_svg(TW, phone=False):
    """phone (S2V, 358 px content box): full lanes 26 / blocks 20 spaced 48 apart, each block link with a 44 px hit rect
    (touch target), short block labels, the axis to 03.2027 so the SAPT readout fits right of the azi line (52 px),
    month labels on quarter starts after the first month (ticks every month)."""
    d0, d1 = DOM0, (DOM1_P if phone else DOM1)
    ppd = (TW - LX) / (d1 - d0).days
    def tx(iso): return round(LX + (dt.date.fromisoformat(iso) - d0).days * ppd, 1)
    LH, BH = (26, 20) if phone else (22, 16)          # compact: lane 22, block 16 (full: 26 / 20); with the focus ring the link is 24 px (D46)
    ly = {'telita': 18, 'hirbovat': 66 if phone else 50}; H = ly['hirbovat'] + LH
    ro_w = 52 if phone else RO_W
    g = ''.join(f'<text class="lanelbl" x="0" y="{ly[f] + 15}">{l}</text><rect class="lane" x="{LX}" y="{ly[f]}" width="{TW - LX}" height="{LH}"/>'
                for f, l in (('telita', 'TELIȚA'), ('hirbovat', 'HÎRBOVĂȚ')))
    def block(f, a, b, label, aria, href, open_=False, marks=''):
        xa, xb = tx(a), tx(b); w = round(xb - xa, 1); y = ly[f] + 3; o = ' open' if open_ else ''
        hit = f'<rect x="{xa}" y="{y + BH // 2 - 22}" width="{w}" height="44" fill="none" pointer-events="all"/>' if phone else ''
        return (f'<a href="{href}" aria-label="{aria}"><title>{aria}</title>{hit}<rect class="ring" x="{round(xa - 4, 1)}" y="{y - 4}" width="{round(w + 8, 1)}" height="{BH + 8}" rx="3"/>'
                f'<rect class="cyc{o}" x="{xa}" y="{y}" width="{w}" height="{BH}" rx="2"/><text class="cyctxt{o}" x="{round(xa + 8, 1)}" y="{y + BH - 4}">{label}</text>{marks}</a>')
    PLC = PL['totals']; ARR = TR['arrival_totals']
    xa = tx(cT['start_date'])
    place = f'<rect class="place" x="{round(xa - 1, 1)}" y="{ly["telita"] + 3}" width="3" height="{BH}"/>'
    g += block('telita', cT['start_date'], cT['end_date'], 'creștere' if phone else f'Seria 2026 · creștere · {cT["weeks"]} săpt.',
               f'Telița, Seria 2026, creștere, {R.drange(cT["start_date"], cT["end_date"])}, {cT["weeks"]} săptămâni, închis. '
               f'Început prin populare (găini {n(PLC["F"])}, cocoși {n(PLC["M"])}), încheiat prin transfer la Hîrbovăț. Deschide ferma la acest ciclu.',
               '#/ferme/telita?ciclu=cyc-telita-2026', marks=place)
    g += block('hirbovat', cH['start_date'], TODAY.isoformat(), f'{"" if phone else "Seria 2026 · "}ouat · din {dd(cH["start_date"])}',
               f'Hîrbovăț, Seria 2026, ouat, din {dd(cH["start_date"])}, deschis, vânzări din {dd(R.EMPTYING_SINCE)}, azi sapt. {WK}. Păsări sosite prin transfer de la Telița '
               f'{dd(dt.date.fromisoformat(X0D))[:5]}–{dd(dt.date.fromisoformat(X1D))} (găini {n(ARR["F"])}, cocoși {n(ARR["M"])}). Deschide ferma la acest ciclu.',
               '#/ferme/hirbovat?ciclu=cyc-hirbovat-2026', open_=True)
    yt, yb = ly['telita'] + 3 + BH, ly['hirbovat'] + 3; ym = (yt + yb) / 2; xa, xb = tx(X0D), tx(X1D)
    g += (f'<g aria-hidden="true"><line class="link" x1="{xa}" y1="{yt}" x2="{xa}" y2="{yb}"/><line class="link" x1="{xb}" y1="{yt}" x2="{xb}" y2="{yb}"/>'
          f'<line class="link" x1="{xa}" y1="{ym}" x2="{xb}" y2="{ym}"/><circle class="linkdot" cx="{xa}" cy="{ym}" r="3"/><circle class="linkdot" cx="{xb}" cy="{ym}" r="3"/></g>')
    xt = tx(TODAY.isoformat()); top = ly['telita'] - 6; yh = ly['hirbovat'] + 3
    g += (f'<line class="today" x1="{xt}" x2="{xt}" y1="{top}" y2="{H + 3}"/><text class="todaylbl" x="{xt}" y="{top - 3}" text-anchor="middle">azi {D[:5]}</text>'
          f'<rect class="ro" x="{round(xt + 4, 1)}" y="{yh}" width="{ro_w}" height="{BH}" rx="2"/><text class="rotxt" x="{round(xt + 4 + ro_w / 2, 1)}" y="{yh + BH - 4}" text-anchor="middle">SAPT {WK}</text>')
    tk = ''
    for d, lbl in months(d0, d1):
        x = tx(d.isoformat()); anc = 'start' if x - LX < 8 else 'middle'
        tk += f'<line class="tick" x1="{x}" x2="{x}" y1="{H}" y2="{H + 4}"/>'
        if not phone or ((d.month - 1) % 3 == 0 and d > d0): tk += f'<text class="month" x="{x}" y="{H + 16}" text-anchor="{anc}">{lbl}</text>'
    desc = (f'Seria 2026, 2 ferme. Telița: creștere {R.drange(cT["start_date"], cT["end_date"])}, {cT["weeks"]} săpt., închis, '
            f'început prin populare. Transfer Telița → Hîrbovăț {dd(dt.date.fromisoformat(X0D))[:5]}–{dd(dt.date.fromisoformat(X1D))}: găini {n(ARR["F"])}, cocoși {n(ARR["M"])}. '
            f'Hîrbovăț: ouat din {dd(cH["start_date"])}, deschis, vânzări din {dd(R.EMPTYING_SINCE)}, azi {D}, sapt. {WK}.')
    return (f'<svg class="ip-track s2-track {"ip" if TW in (1160, 1000, 720, 560) else "s2"}-track--{TW}" viewBox="0 0 {TW} {H + 18}" role="group" aria-label="Ciclurile seriei 2026 pe ferme">'
            f'<desc>{desc}</desc>{g}{tk}</svg>')
def track(phone=False):
    """phone (S2V): one 358 px drawing, "Deschide seria" as a 48 px link under the legend instead of inside the note."""
    leg = (f'<span><i class="ip-rec ip-rec--placement"></i>populare</span><span><i class="ip-rec ip-rec--transfer"></i>transfer '
           f'<span class="ip-num">{dd(dt.date.fromisoformat(X0D))[:5]}–{dd(dt.date.fromisoformat(X1D))}</span></span><span><i class="today"></i>azi</span>')
    lnk = '<a class="ip-link s2-lnk" href="#/serii/seria-2026">Deschide seria</a>'
    note = f'{num(f"{DOM0:%m.%Y}–{DOM1 - dt.timedelta(days=1):%m.%Y}")} · {num(2)} ferme · {num(2)} cicluri' + ('' if phone else f' · {lnk}')
    svgs = track_svg(358, phone=True) if phone else "".join(track_svg(tw) for tw in TWS)
    more = lnk.replace('s2-lnk', 's2-more') if phone else ''
    return (f'<section class="s2-block" aria-labelledby="tk">{head("Seria 2026 · cicluri pe ferme", note, "tk")}'
            f'<div class="ip-trackbox">{svgs}</div><div class="ip-tracklegend" aria-hidden="true">{leg}</div>{more}</section>')

# ---------- page parts ----------
def pagehead(ctx, phone=False):
    return f'<header class="ip-pagehead{" s2-phead" if phone else ""}"><div><h1>Acasă</h1><p class="ip-note ip-num ip-m0">{ctx}</p></div></header>'
CTX_VET = f'HÎRBOVĂȚ · {D} · SAPT {WK}'
CTX_MGR = f'SERIA 2026 · {D} · HÎRBOVĂȚ SAPT {WK}'
CTX_OWN = f'SERIA 2026 · {D} · 2 FERME'  # uppercase Mono context line, as S6/S7

def tabbar(vet=False, key_off=None, viewer=False):
    """key_off (S2S): id of the reason text; the Înregistrează key is then aria-disabled (no href, focusable, as SYS09 Editează)."""
    it = lambda k, lbl, h, cur=False: (f'<a class="ip-tabbar__item{" is-current" if cur else ""}" href="{h}"{" aria-current=\"page\"" if cur else ""}>{icon(k, 22)}{lbl}</a>')
    key = f'<a class="ip-tabbar__item" href="#inregistreaza"><span class="ip-tabbar__key">{icon("plus", 22)}</span>Înregistrează</a>'
    if key_off: key = key.replace('href="#inregistreaza"', f'role="link" tabindex="0" aria-disabled="true" aria-describedby="{key_off}"', 1)
    if viewer:  # S2V (D89): the viewer enters nothing, so no Înregistrează key: 4 items
        return (f'<nav class="ip-tabbar ip-screen__tabbar" aria-label="Navigare">{it("home", "Acasă", "#acasa", True)}{it("layers", "Serie", "#serie")}'
                f'{it("chart", "Rapoarte", "#rapoarte")}{it("more", "Mai mult", "#mai-mult")}</nav>')
    items = (it('home', 'Acasă', '#acasa', True) + key + it('user', 'Cont', '#cont')) if vet else \
            (it('home', 'Acasă', '#acasa', True) + it('layers', 'Serie', '#serie') + key + it('chart', 'Rapoarte', '#rapoarte') + it('more', 'Mai mult', '#mai-mult'))
    return f'<nav class="ip-tabbar ip-screen__tabbar" aria-label="Navigare">{items}</nav>'

def phone(parts, vet=False, viewer=False):
    return f'''<div class="ip-screen">
<main class="ip-screen__main s2-main">
{"".join(parts)}
</main>
{tabbar(vet, viewer=viewer)}
</div>'''

def sidebar(role):
    sb = s4.sidebar(role)
    sb = sb.replace('class="ip-nav is-current" href="#serie" aria-current="page"', 'class="ip-nav" href="#serie"')
    sb = sb.replace('class="ip-nav" href="#acasa"', 'class="ip-nav is-current" href="#acasa" aria-current="page"')
    assert sb.count('is-current') == 1 and 'href="#acasa" aria-current="page"' in sb
    return sb
def desktop(role, parts):
    return f'<div class="ip-shell">{sidebar(role)}<main class="ip-shell__main">\n{desk_heads("".join(parts))}\n</main></div>'

VET = phone([pagehead(CTX_VET, True), missing(), entrykeys(), introdus()], vet=True)
MGR_P = phone([pagehead(CTX_MGR, True), alerts(), stock(), entrykeys(True), missing(True), introdus_mgr()])  # D87: Introdus azi last
MGR_D = desktop('Manager', [pagehead(CTX_MGR), f'<div class="s2-cols">{alerts()}{missing(True)}</div>', stock(), entrykeys(True), introdus_mgr()])
OWN_D = desktop('Proprietar', [pagehead(CTX_OWN), track(), f'<div class="s2-cols">{alerts(True)}{unpriced()}</div>', stock(), farms_block()])

# ---------- CSS ----------
PICK = ('body{',)  # FIN15: .ip-screen/__main are additions.css §57; FIN15B: .s7-m0 is ip-m0 (§58)
CSS7 = '\n'.join(l for l in B.CSS.splitlines() if l.startswith(PICK)); assert CSS7.count('\n') == len(PICK) - 1
CSS_BLOCK = '''
.s2-block{display:flex;flex-direction:column;gap:var(--space-2);min-width:0}
.s2-block > .ip-list__head{margin:0}'''  # phone blocks only (FX02: desktop blocks are additions.css §61 ip-block)
CSS_LINK = '''
.s2-more,.s2-lnk{display:inline-flex;align-items:center;min-height:var(--space-6);align-self:flex-start} /* text links keep a >=24px target (D46) */'''
CSS_SHARED = CSS_BLOCK + CSS_LINK
CSS_P = CSS7 + CSS_SHARED + '''
.s2-phead h1{margin:0}
.s2-more{min-height:var(--control-h-touch)} /* touch: 48px target */'''
CSS_D = s4.CSS + CSS_LINK + '''
.s2-cols{display:grid;grid-template-columns:minmax(0,7fr) minmax(0,5fr);gap:var(--space-6);align-items:start}
.s2-cols > .ip-unpriced{margin-top:calc(var(--ip-lh-heading) + var(--space-3))} /* the panel's top on the alert list's top (under its section head, FX02) */
.s2-fsums{grid-template-columns:repeat(2,minmax(0,1fr))}'''  # FIN15B: the CycleTrack width box is additions.css §65 ip-trackbox; FX05: the MissingRow meta wrap is additions.css §1

# ---------- write ----------
STEM0, TITLE0 = 'S2-acasa', 'S2 · Acasă'
HEIGHTS = json.loads(HP.read_text()) if HP and HP.exists() else {}
def write_phone(stem, title, body, css=None, fixed=False):
    CSS_P_ = css or CSS_P
    h = HEIGHTS.get(stem, 844)
    hs = f'height:{h}px; overflow:hidden' if fixed else f'min-height:{h}px'  # FIN15B: a fixed screen (ip-screen--fixed) takes the frame's height
    root = f'<div class="ip-root" style="padding:0; width:390px; {hs}">{body}</div>'
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
<style>{CSS_P_}</style>
</helmet>
{root}
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":390,"height":{h}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
    (ROOT / f'design/canvas/project/{stem}.dc.html').write_text(dc, encoding='utf-8')
    SYS = ROOT / 'design/system'
    wroot = root if fixed else root.replace(f'min-height:{h}px', 'min-height:844px', 1)  # RF06: the render keeps only the 844 floor, so s2_render measures the natural height
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{B.FONTS}<style>{CSS_P_}</style></head><body>{wroot}</body></html>'''
    pathlib.Path('/tmp/s2').mkdir(exist_ok=True)
    pathlib.Path(f'/tmp/s2/{stem}.html').write_text(wrap, encoding='utf-8')
    print('ok', stem, title, h)
    return stem, title, h, 390

out = [write_phone(STEM0, f'{TITLE0} · vet · telefon', VET),
       write_phone(STEM0 + '--manager', f'{TITLE0} · manager · telefon', MGR_P)]
for stem, title, body in ((STEM0 + '--manager-desktop', f'{TITLE0} · manager · desktop', MGR_D),
                          (STEM0 + '--proprietar-desktop', f'{TITLE0} · proprietar · desktop', OWN_D)):
    h = HEIGHTS.get(stem, 1024)
    s4.write(stem, title, body, H=h, css=CSS_D)
    out.append((stem, title, h, 1440))
if HP and not HP.exists():
    HP.write_text(json.dumps({s: h for s, _, h, _ in out}, indent=1))

if CANVAS:  # S2 row, entries keep x/y, a new row goes below the lowest (canvas_rows.py); 80 px gaps
    c = CR.load(ROOT)
    x = CR.place(c, STEM0, [(s, w, h, t) for s, t, h, w in out])
    CR.save(ROOT, c)
    print('canvas: S2 row y', CR.row_y(c, STEM0), 'ends at x', x - 80)
