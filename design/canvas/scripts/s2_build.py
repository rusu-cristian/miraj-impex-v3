"""S2 · Acasă (spec D2): the role's start page on 16.11.2026 (D23), Seria 2026.
- vet · telefon: header (farm, today, SAPT) · Lipsesc date (MissingRow, C9) · Înregistrează (EntryKeys, 3) · Introdus azi
  (EventRow actions, C8, one row per record type saved today).
- manager · telefon / desktop: header · De verificat (AlertRow) · zile de hrană (StockReadout per farm × type) ·
  Înregistrează (EntryKeys, 6) · Lipsesc date (farm named first, several farms).
- proprietar · desktop: compact CycleTrack for Seria 2026 · De verificat across farms · Livrări fără preț (UnpricedCount) ·
  stoc hrană (StockReadout) · ferme (FarmSummary --compact).
Shells are reused by import: the phone page uses the S7 phone screen/main classes (s7s_build → s7_build helpers and CSS
lines) with the system BottomTabBar; the desktop uses the S4 shell (s4s_build: Sidebar with Acasă current, PageHeader
type, AlertRows from s4s_build.alerts(), board writer). Blocks are the SYS08 components (MissingRow, EntryKeys,
UnpricedCount), SYS01 AlertRow, SYS03 FarmSummary compact, SYS04 StockReadout, SYS09 EventRow actions.
Every figure comes from the extracts (Hîrbovăț after 28.02.2026 is sample, D27); the vet is Dr. Ceban, the manager
Ion Rusu (illustrative). The page is drawn after 12:00 (MissingRow lists today's rows only from 12:00, C9).
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
def nw(v): return f'<span class="s2-nw">{v}</span>'  # S2P: a meta group that never splits (lines break at the · separators)

# ---------- data ----------
EX = ROOT / 'design/data/extracts'
def J(name): return json.load(open(EX / name, encoding='utf-8'))
S, CL, CY, FT, FS = s4.S, s4.CL, J('cycles.json'), J('feed-types.json'), J('farms-sections.json')
TODAY = dt.date(2026, 11, 16); YDAY = TODAY - dt.timedelta(days=1); D = dd(TODAY)
cyc = {c['id']: c for c in CY['cycles']}
cT, cH = cyc['cyc-telita-2026'], cyc['cyc-hirbovat-2026']
WK = (TODAY - dt.date.fromisoformat(cH['start_date'])).days // 7 + 1
assert WK == 40 == cH['week_today'] and cT['status'] == 'closed' and cT['end_date'] == '2026-02-28' and cT['weeks'] == 21
farms = {f['farm_id']: f for f in CL['farms']}
HB, TL = farms['hirbovat'], farms['telita']
ALIVE = HB['totals']; assert (ALIVE['F'], ALIVE['M'], ALIVE['all']) == (25252, 2685, 27937) and TL['totals']['all'] == 0
alive = {(x['section'], sx): x[sx] for x in HB['sections'] for sx in 'FM'}
FTN = {r['id']: r['name'] for r in FT['rows']}
GUIDES = {g['id']: g['name'] for g in FS['guides']}
TR = J('transfers-seria2026.json'); PL = J('placement-seria2026.json')
XD = sorted({a['date'] for a in TR['arrivals']} | {o['date'] for o in TR['transfers_out']})
X0D, X1D = XD[0], XD[-1]; assert (X0D, X1D) == ('2026-02-22', '2026-02-28')

# C9 · Lipsesc date: sections with birds and no Hrană zilnică / Zilnic row, today (from 12:00) and the 7 days before
alloc = {(a['section'], a['sex']): a for a in S['allocations_today']}
MISS = [(r['date'], r['section'], r['sex']) for r in S['daily']
        if (TODAY - dt.timedelta(days=7)).isoformat() <= r['date'] <= TODAY.isoformat() and r['g_per_bird'] is None]
assert MISS == [('2026-11-16', 'H3.2', 'M'), ('2026-11-16', 'H4.2', 'F')], MISS
z_days = {e['at'][:10] for e in S['event_log'] if e['record'] == 'mortality' and e['action'] == 'create'}
assert {(TODAY - dt.timedelta(days=k)).isoformat() for k in range(7)} <= z_days  # Zilnic saved on 10–16.11 (the log starts 10.11)
assert all(r['mortality'] is not None for r in S['daily'] if r['date'] >= '2026-11-09')

# C8 · Introdus azi (Dr. Ceban, 16.11.2026): one row per record type saved today
LOG = {e['record']: e for e in S['event_log'] if e['at'].startswith(TODAY.isoformat()) and e['user'] == 'Dr. Ceban'}
assert sorted(LOG) == ['eggs', 'feed', 'mortality', 'weighing']
D16 = {sx: sum(r['mortality'] for r in S['daily'] if r['date'] == TODAY.isoformat() and r['sex'] == sx) for sx in 'FM'}
C16 = [c for c in S['culls'] if c['date'] == TODAY.isoformat()]
assert D16 == {'F': 8, 'M': 1} and not C16 and f'♀ −{D16["F"]} · ♂ −{D16["M"]}' in LOG['mortality']['summary']
EGGS16 = [r for r in S['eggs_daily'] if r['date'] == TODAY.isoformat()]
EGG_T = sum(r['total_eggs'] for r in EGGS16); assert EGG_T == 18688 and n(EGG_T) in LOG['eggs']['summary']
FEED_ROWS = [a for a in S['allocations_today'] if a['today_g'] is not None]
FEED_KG = round(sum(round(a['today_g'] * a['alive'] / 1000, 1) for a in FEED_ROWS), 1)
assert (len(FEED_ROWS), n(FEED_KG, 1)) == (10, f'3{NB}691,3')

# stock (SYS04) and unpriced deliveries (SYS08)
STK = {s['feed_type_id']: s for s in S['stock_today']}
B2, CO = STK['ft-breeder2'], STK['ft-cocosi']
assert (B2['stock_kg'], B2['days_left'], CO['stock_kg'], CO['days_left']) == (28252, 7.1, 1650, 3.9)
UNP = [d for d in S['deliveries'] if not d['priced']]
UNP_KG = sum(d['qty_kg'] for d in UNP); assert (len(UNP), UNP_KG) == (4, 76107) == (S['summary_today']['unpriced_deliveries_all_farms'], S['summary_today']['unpriced_kg'])
LAY, LAYG = s4.lay_today, s4.lay_guide; MH = S['summary_today']['cumulative_mortality_pct']
assert (LAY, LAYG, MH['F'], MH['M']) == (74.0, 74.1, 5.54, 11.93)

# ---------- blocks ----------
def head(cap, note, hid):
    return f'<div class="ip-list__head"><h2 class="ip-cap" id="{hid}">{cap}</h2><p class="ip-note">{note}</p></div>'

def day_row(day, sec, sx):  # one section × sex × day of S['daily'] (alive at the end of that day, grams of the day before)
    r = {(x['date'], x['section'], x['sex']): x for x in S['daily']}
    prev = (dt.date.fromisoformat(day) - dt.timedelta(days=1)).isoformat()
    return r[(day, sec, sx)], r[(prev, sec, sx)]
def missing(manager=False, u='', miss=None, past=(), locked=False, when=None, empty=None, scope=None):
    """MissingRow (SYS08): one row per section × entry × day, newest first; manager rows name the farm first.
    S2S: past = extra past-day rows (D39 state examples), locked = a vet's past day (D07: Tag "blocat", no link),
    when = the scope note's time part (before 12:00: the clock time), empty = the sentence of the empty state."""
    rows = []
    for day, sec, sx in sorted(list(MISS if miss is None else miss) + list(past), key=lambda m: (dt.date.fromisoformat(m[0]).toordinal() * -1, m[1])):
        c, sign, word = ('hen', '♀', 'găini') if sx == 'F' else ('rooster', '♂', 'cocoși')
        d0 = dt.date.fromisoformat(day); wk = (d0 - dt.date.fromisoformat(cH['start_date'])).days // 7 + 1
        if day == TODAY.isoformat():
            alv, gy = alive[(sec, sx)], alloc[(sec, sx)]['yesterday_g']
        else:
            rd, ry = day_row(day, sec, sx); alv, gy = rd['alive_end'], ry['g_per_bird']
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
INTRO = ['mortality', 'eggs', 'feed', 'weighing']  # the vet's four saves of 16.11.2026, in time order
def introdus(menu=None, off=False, edited=None):
    """menu / edited: {record: 'open' | 'hh:mm'} for one row (S2S); off: the 23:59 window closed (every row)."""
    menu, edited = menu or {}, edited or {}
    spec = {
        'mortality': (f'Mortalitate · {num(6)} secțiuni · {hen(n(-D16["F"]))} · {roo(n(-D16["M"]))}',
                      f'{num(t_of(LOG["mortality"]))} · Dr. Ceban · Zilnic, fără sacrificări', 'Mortalitate de azi'),
        'eggs': (f'Ouă · H1, H3, H4 · {num(n(EGG_T))} ouă', f'{num(t_of(LOG["eggs"]))} · Dr. Ceban · ouat {num(s4.pct(LAY))}', 'Ouă de azi'),
        'feed': (f'Hrană zilnică · {num(len(FEED_ROWS))} din {num(12)} rânduri · {num(n(FEED_KG, 1) + NB + "kg")}',
                 f'{num(t_of(LOG["feed"]))} · Dr. Ceban · fără {M} H3.2, {F} H4.2', 'Hrană zilnică de azi'),
        'weighing': (f'Cântărire · {num(6)} secțiuni · {pair(n(s4.w_f), n(s4.w_m), NB + "g")}',
                     f'{num(t_of(LOG["weighing"]))} · Dr. Ceban · medii pe fermă', 'Cântărire de azi'),
    }
    rows = [act_row(k, i, *spec[k], menu=menu.get(k, 'closed'), off=off, edited=edited.get(k)) for i, k in enumerate(INTRO, 1)]
    note = (f"{num(D)} · timpul de modificare s-a încheiat la {num('23:59')}" if off else f"{num(D)} · poți modifica sau șterge până la {num('23:59')}")
    return (f'<section class="s2-block" aria-labelledby="ia">{head("Introdus azi", note, "ia")}'
            f'<ul class="ip-list ip-list--act" aria-labelledby="ia">{"".join(rows)}</ul>'
            f'<a class="ip-link s2-more" href="#introdus-azi">Vezi pe secțiuni</a></section>')

# C8 · Introdus azi for the manager (D87): Ion Rusu's records of 16.11.2026. The sample has none (his last saves are the
# 14.11 delivery and the 15.11 correction), so the block shows its empty state; nothing is invented.
MGR_LOG = [e for e in S['event_log'] if e['at'].startswith(TODAY.isoformat()) and e['user'] == 'Ion Rusu']
assert not MGR_LOG and any(e['user'] == 'Ion Rusu' for e in S['event_log'])
def introdus_empty(line, u=''):
    """Introdus azi with nothing saved today: the list frame keeps its place, off Lamp + word + one sentence (SYS09 li.ip-list__empty)."""
    return (f'<section class="s2-block" aria-labelledby="{u}ia">{head("Introdus azi", num(D), u + "ia")}'
            f'<ul class="ip-list ip-list--act" aria-labelledby="{u}ia"><li class="ip-list__empty"><span class="ip-lamp ip-lamp--off">nimic introdus</span>'
            f'<span>{line}</span></li></ul></section>')
INTRO_MGR = f'Nu ai salvat nimic azi. Ce salvezi azi apare aici, pe tipuri; poți modifica sau șterge până la {num("23:59")}.'  # S15P: the S15 gol sentence (D98)

FARM_P = '<div class="ip-ev__main">Hîrbovăț · '
def alerts(owner=False, u=''):
    a_h3, a_h12, a_32, a_42, a_st, a_pr = s4.alerts()
    # the missing rows have their own block (Lipsesc date) for the manager; the owner has none, so they stay alerts.
    # "fără preț" has its own block for the owner (UnpricedCount); the manager can't add prices (D05), so it is not his alert.
    items = [a_h3, a_h12, a_32, a_42, a_st] if owner else [a_h3, a_h12, a_st]
    items = [a.replace('<div class="ip-ev__main">', FARM_P, 1) for a in items]
    LIM = 'limită <span class="ip-num">5 zile</span>'
    items = [a.replace(' · ' + LIM, ' · ' + nw(LIM), 1) for a in items]
    assert sum(LIM in a for a in items) == 1
    # the meta breaks only at its · separators (each group stays whole, as the MissingRow meta)
    items = [re.sub(r'(<div class="ip-ev__meta ip-num">)(.*?)(</div>)', lambda m: m[1] + ' · '.join(nw(g) for g in m[2].split(' · ')) + m[3], a, count=1) for a in items]
    assert all(a.count(FARM_P) == 1 for a in items)
    note = f'{num(len(items))} alerte · {"toate fermele" if owner else "fermele tale"} · {num(D)}'
    return f'<section class="s2-block" aria-labelledby="{u}dv">{head("De verificat", note, u + "dv")}<div class="ip-list ip-list--alerts">{"".join(items)}</div></section>'

def stock(u='', cap='Zile de hrană'):
    def ro(fid, s, sign, word):
        st = 'ok' if s['days_left'] >= 5 else 'warn'
        w0, w1 = s['avg_window']
        sub = '' if st == 'ok' else '<p class="ip-readout__sub s2-m0">sub pragul de 5 zile</p>'
        rid = f'{u}sr-{fid}'
        return (f'<li><div class="ip-readout ip-readout--stock" role="group" aria-labelledby="{rid}"><div class="ip-readout__cap"><span id="{rid}">Hîrbovăț · {FTN[fid]}'
                f'<span class="ip-sr">, pentru {word}</span></span><span class="ip-rlamp ip-rlamp--{st}">{"în normă" if st == "ok" else "abatere"}</span></div>'
                f'<div class="ip-readout__val">{n(s["stock_kg"])}<small> kg</small></div><dl class="ip-readout__kv"><div><dt>zile rămase</dt><dd>{n(s["days_left"], 1)}<small> zile</small></dd></div>'
                f'<div><dt>medie/zi {dd(dt.date.fromisoformat(w0))[:5]}–{dd(dt.date.fromisoformat(w1))[:5]}</dt><dd>{n(s["avg_daily_kg_last3"])}<small> kg</small></dd></div>'
                f'<div><dt>se termină</dt><dd>{dd(dt.date.fromisoformat(s["run_out_date"]))}</dd></div></dl>{sub}</div></li>')
    tel = (f'<li><div class="ip-readout ip-readout--stock ip-readout--off" role="group" aria-labelledby="{u}sr-te"><div class="ip-readout__cap"><span id="{u}sr-te">Telița</span>'
           f'<span class="ip-rlamp ip-rlamp--off">oprit</span></div><div class="ip-readout__val">–</div><p class="ip-readout__sub s2-m0">fără ciclu deschis</p>'
           f'<p class="ip-readout__sub s2-m0">ultimul ciclu: Seria 2026, închis pe <span class="ip-num s2-nw">{dd(dt.date.fromisoformat(cT["end_date"]))}</span></p></div></li>')
    w0, w1 = B2['avg_window']
    return (f'<section class="s2-block" aria-labelledby="{u}st">{head(cap, f"stoc și zile rămase · {num(D)}", u + "st")}'
            f'<ul class="ip-readouts" aria-labelledby="{u}st">{ro("ft-breeder2", B2, "♀", "găini")}{ro("ft-cocosi", CO, "♂", "cocoși")}{tel}</ul>'
            f'<p class="ip-note s2-m0">Zile rămase = stoc ÷ consumul mediu din ultimele 3 zile încheiate ({num(dd(dt.date.fromisoformat(w0))[:5] + "–" + dd(dt.date.fromisoformat(w1)))}); sub 5 zile = abatere.</p></section>')

def unpriced():
    by = {}
    for d in UNP: by.setdefault(d['feed_type_id'], []).append(d)
    li = ''.join(f'<li><span>Hîrbovăț · {FTN[t].replace(" ", "&nbsp;")}</span><span>{num(len(ds))} {"livrare" if len(ds) == 1 else "livrări"} · {num(n(sum(d["qty_kg"] for d in ds)) + " kg")}</span></li>'
                 for t, ds in by.items())
    assert [(t, sum(d['qty_kg'] for d in ds)) for t, ds in by.items()] == [('ft-breeder2', 72068), ('ft-cocosi', 4039)]
    old = min(d['date'] for d in UNP)
    return (f'<section class="ip-unpriced" aria-labelledby="up"><div class="ip-unpriced__head">{tile("price")}<h2 class="ip-cap" id="up">Livrări fără preț</h2></div>'
            f'<p class="ip-unpriced__fig"><span class="ip-unpriced__val">{len(UNP)}<small> livrări</small></span><span class="ip-unpriced__kg">{n(UNP_KG)} kg</span></p>'
            f'<ul class="ip-unpriced__list">{li}</ul><p class="ip-note">Cea mai veche din {num(dd(dt.date.fromisoformat(old)))}. Kg fără preț nu intră în cost.</p>'
            f'<a class="ip-btn ip-btn--secondary" href="#rapoarte-hrana-livrari-fara-pret">Adaugă prețurile</a></section>')

def farms_block():
    P_ = '<small class="ip-fsum__pct">%</small>'
    lay_ok = abs(LAY - LAYG) <= 3
    hir = (f'<article class="ip-fsum ip-fsum--compact" aria-labelledby="fs-hir" aria-describedby="fs-hir-c"><div class="ip-fsum__head"><div class="ip-fsum__id">'
           f'<h3 class="ip-fsum__name" id="fs-hir"><a href="#/ferme/hirbovat?ciclu=cyc-hirbovat-2026">Hîrbovăț</a></h3><span class="ip-tag ip-tag--neutral">Ouat</span>'
           f'<span class="ip-tag ip-tag--neutral">{GUIDES["g-laying"]}</span></div>{lamp("ok", "deschis")}</div>'
           f'<p class="ip-fsum__cycle" id="fs-hir-c">ciclu din {dd(dt.date.fromisoformat(cH["start_date"]))} · SAPT {WK}</p>'
           f'<p class="ip-fsum__alive"><span class="ip-cap">În viață</span>{pair(n(ALIVE["F"]), n(ALIVE["M"]))}</p><dl class="ip-fsum__figs">'
           f'<div class="ip-fsum__fig"><dt>{ico("eggs")}Ouat azi</dt><dd class="ip-fsum__val">{n(LAY, 1)}{P_}</dd><dd class="ip-fsum__sub"><span class="ip-num">ghid {s4.pct(LAYG)} · SAPT {WK}</span></dd>'
           f'<dd class="ip-fsum__status">{lamp("ok" if lay_ok else "warn", "în normă" if lay_ok else "abatere")}</dd></div>'
           f'<div class="ip-fsum__fig"><dt>{ico("mortality")}Mort. cumulată</dt><dd class="ip-fsum__val">{pair(n(MH["F"], 2), n(MH["M"], 2))}{P_}</dd>'
           f'<dd class="ip-fsum__sub">ghid: – fără date în ghid</dd></div></dl></article>')
    tel = (f'<article class="ip-fsum ip-fsum--compact" aria-labelledby="fs-tel" aria-describedby="fs-tel-c"><div class="ip-fsum__head"><div class="ip-fsum__id">'
           f'<h3 class="ip-fsum__name" id="fs-tel"><a href="#/ferme/telita">Telița</a></h3><span class="ip-tag ip-tag--neutral">Creștere</span>'
           f'<span class="ip-tag ip-tag--neutral">{GUIDES["g-growth"]}</span></div>{lamp("off", "goală")}</div>'
           f'<div class="ip-fsum__empty" id="fs-tel-c"><p class="s2-m0">Ferma e goală din {num(dd(dt.date.fromisoformat(cT["end_date"])))}. Ultimul ciclu: Seria 2026, {num(cT["weeks"])} săpt.</p>'
           f'<a class="ip-link" href="#/ferme/telita?ciclu=cyc-telita-2026">Vezi ultimul ciclu</a></div></article>')
    return (f'<section class="s2-block" aria-labelledby="fm">{head("Ferme", f"{num(2)} ferme · {num(D)}", "fm")}'
            f'<div class="ip-fsums s2-fsums">{hir}{tel}</div></section>')

# ---------- compact CycleTrack (owner): Seria 2026, one lane per farm ----------
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
    d0, d1 = dt.date(2025, 10, 1), (dt.date(2027, 3, 1) if phone else dt.date(2027, 1, 1))
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
               f'Telița, Seria 2026, creștere, {dd(dt.date.fromisoformat(cT["start_date"]))}–{dd(dt.date.fromisoformat(cT["end_date"]))}, {cT["weeks"]} săptămâni, închis. '
               f'Început prin populare (găini {n(PLC["F"])}, cocoși {n(PLC["M"])}), încheiat prin transfer la Hîrbovăț. Deschide ferma la acest ciclu.',
               '#/ferme/telita?ciclu=cyc-telita-2026', marks=place)
    g += block('hirbovat', cH['start_date'], TODAY.isoformat(), f'{"" if phone else "Seria 2026 · "}ouat · din {dd(dt.date.fromisoformat(cH["start_date"]))}',
               f'Hîrbovăț, Seria 2026, ouat, din {dd(dt.date.fromisoformat(cH["start_date"]))}, deschis, azi sapt. {WK}. Păsări sosite prin transfer de la Telița '
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
    desc = (f'Seria 2026, 2 ferme. Telița: creștere {dd(dt.date.fromisoformat(cT["start_date"]))}–{dd(dt.date.fromisoformat(cT["end_date"]))}, {cT["weeks"]} săpt., închis, '
            f'început prin populare. Transfer Telița → Hîrbovăț {dd(dt.date.fromisoformat(X0D))[:5]}–{dd(dt.date.fromisoformat(X1D))}: găini {n(ARR["F"])}, cocoși {n(ARR["M"])}. '
            f'Hîrbovăț: ouat din {dd(dt.date.fromisoformat(cH["start_date"]))}, deschis, azi {D}, sapt. {WK}.')
    return (f'<svg class="ip-track s2-track s2-track--{TW}" viewBox="0 0 {TW} {H + 18}" role="group" aria-label="Ciclurile seriei 2026 pe ferme">'
            f'<desc>{desc}</desc>{g}{tk}</svg>')
def track(phone=False):
    """phone (S2V): one 358 px drawing, "Deschide seria" as a 48 px link under the legend instead of inside the note."""
    leg = (f'<span><i class="ip-rec ip-rec--placement"></i>populare</span><span><i class="ip-rec ip-rec--transfer"></i>transfer '
           f'<span class="ip-num">{dd(dt.date.fromisoformat(X0D))[:5]}–{dd(dt.date.fromisoformat(X1D))}</span></span><span><i class="today"></i>azi</span>')
    lnk = '<a class="ip-link s2-lnk" href="#/serii/seria-2026">Deschide seria</a>'
    note = f'{num("10.2025–12.2026")} · {num(2)} ferme · {num(2)} cicluri' + ('' if phone else f' · {lnk}')
    svgs = track_svg(358, phone=True) if phone else "".join(track_svg(tw) for tw in TWS)
    more = lnk.replace('s2-lnk', 's2-more') if phone else ''
    return (f'<section class="s2-block" aria-labelledby="tk">{head("Seria 2026 · cicluri pe ferme", note, "tk")}'
            f'<div class="s2-trackbox">{svgs}</div><div class="ip-tracklegend" aria-hidden="true">{leg}</div>{more}</section>')

# ---------- page parts ----------
def pagehead(ctx, phone=False):
    return f'<header class="ip-pagehead{" s2-phead" if phone else ""}"><div><h1>Acasă</h1><p class="ip-note ip-num s2-m0">{ctx}</p></div></header>'
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
    return f'<div class="ip-shell">{sidebar(role)}<main class="ip-shell__main">\n{"".join(parts)}\n</main></div>'

VET = phone([pagehead(CTX_VET, True), missing(), entrykeys(), introdus()], vet=True)
MGR_P = phone([pagehead(CTX_MGR, True), alerts(), stock(), entrykeys(True), missing(True), introdus_empty(INTRO_MGR)])  # D87: Introdus azi last
MGR_D = desktop('Manager', [pagehead(CTX_MGR), f'<div class="s2-cols">{alerts()}{missing(True)}</div>', stock(), entrykeys(True), introdus_empty(INTRO_MGR)])
OWN_D = desktop('Proprietar', [pagehead(CTX_OWN), track(), f'<div class="s2-cols">{alerts(True)}{unpriced()}</div>', stock(cap='Stoc hrană'), farms_block()])

# ---------- CSS ----------
PICK = ('body{', '.s7-m0{')  # FIN15: .ip-screen/__main are additions.css §57
CSS7 = '\n'.join(l for l in B.CSS.splitlines() if l.startswith(PICK)); assert CSS7.count('\n') == len(PICK) - 1
CSS_SHARED = '''
.s2-m0{margin:0}
.s2-nw{white-space:nowrap}
.s2-block{display:flex;flex-direction:column;gap:var(--space-2);min-width:0}
.s2-block > .ip-list__head{margin:0}
.s2-more,.s2-lnk{display:inline-flex;align-items:center;min-height:var(--space-6);align-self:flex-start} /* text links keep a >=24px target (D46) */'''
CSS_P = CSS7 + CSS_SHARED + '''
.s2-phead h1{margin:0}
.s2-more{min-height:var(--control-h-touch)} /* touch: 48px target */'''
CSS_D = s4.CSS + CSS_SHARED + '''
.s2-cols{display:grid;grid-template-columns:minmax(0,7fr) minmax(0,5fr);gap:var(--space-6);align-items:start}
.s2-cols > .ip-unpriced{margin-top:calc(var(--ip-lh-small) + var(--space-2))} /* the panel's top on the alert list's top (under its head) */
.s2-fsums{grid-template-columns:repeat(2,minmax(0,1fr))}
.ip-ev--missing .ip-ev__main > .ip-ev__meta{white-space:normal} /* the S4 shell keeps ip-num children of a row on one line; the MissingRow meta wraps in the 5fr column */
.s2-trackbox{container-type:inline-size}
.s2-track--1000{display:none}
@container (max-width: 1159.98px){.s2-track--1160{display:none}.s2-track--1000{display:block}}'''

# ---------- write ----------
STEM0, TITLE0 = 'S2-acasa', 'S2 · Acasă'
HEIGHTS = json.loads(HP.read_text()) if HP and HP.exists() else {}
def write_phone(stem, title, body, css=None):
    CSS_P_ = css or CSS_P
    h = HEIGHTS.get(stem, 844)
    root = f'<div class="ip-root" style="padding:0; width:390px; min-height:{h}px">{body}</div>'
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
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{B.FONTS}<style>{CSS_P_}</style></head><body>{root}</body></html>'''
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
