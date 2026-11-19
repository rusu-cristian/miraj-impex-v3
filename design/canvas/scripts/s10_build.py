"""S10 · Vânzare (manager, spec C4): Fermă, then Data; one SexCountRow per section with ♀/♂ counts and Toate; optional
Cumpărător and Notă; the EffectPreview (farm before → after, what remains, the cycle that closes); Înregistrează vânzarea;
a confirmation Dialog when the sale empties the farm (C4 › Confirmation).
Phone boards reuse the S7 phone form shell by import (s7s_build → s7_build: FormHeader, DateField, icons, SaveBar shell +
2-line summary, viewport/Toast/Banner CSS from sstate_css), S13's record-icon/EffectPreview/"(opțional)" rules and S9's
EffectPreview line rules + modal anchor (sstate_css, s13-/s9- names kept); the rows are the system SexCountRow (SYS07,
stacked phone layout), the cycle marks the CycleTrack vocabulary (`ip-cmark--end sale`, SYS06), the Dialog the system
Dialog in the phone scrim (additions.css section 35, ip-scrim--phone, keys stacked; FIN04). The desktop board reuses
the S4 Hîrbovăț farm page (s4s_build shell) with the system SidePanel beside it (non-modal, D73), "Vânzare" pressed.
Content (RF05, realdata; D151–D158): today 18.11.2026, SAPT 59. The default board is the real 18.11 partial sale (D155,
D158b): H6.1 ♀ 676 · ♂ 201 and H6.2 ♀ 1 155 · ♂ 88 typed, H5.1/H5.2 untouched, the farm goes from ♀ 14 392 · ♂ 1 083
(alive before the sale: end of day + sold that day) to the end-of-day ♀ 12 561 · ♂ 794, the cycle continues; salvat and
eroare la salvare are its states. A compartment row lists the sections with birds; an empty hall (H7, since 14.11) is one
muted "goală din 14.11.2026 · H7.1, H7.2" line (D158d, as S7/S6/S4). D39 examples on the real counts: emptying the whole
farm on the 18.11 end counts (Toate secțiunile, ♀ 12 561 · ♂ 794 → 0, cycle closed in sapt. 59, confirmation Dialog) and
the refusal (the 18.11 sale with H6.2 ♂ 98 typed over the 93 alive). Parțială replays the real 14.11.2026 hall-emptying
sale (Toate on H7.1 + H7.2, ♀ 6 915 · ♂ 639, SAPT 58; "today" = 14.11, as S9/S13 replays, D73), the cycle continues.
The desktop board types nothing (end-of-day counts). S10P: the ghost "Toate secțiunile" key (additions.css §34, D84).
python3 -I s10_build.py <repo_root> [heights.json] [--canvas]
Writes design/canvas/project/S10-vanzare[--<state>].dc.html, /tmp/s10/<stem>.html and /tmp/s4s/<stem>.html (renders);
--canvas also sets the S10 boards in design/canvas/project/canvas.json (boards + order)."""
import sys, json, pathlib, importlib, datetime as dt

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
ROOT, HP, CANVAS = CR.args('s10_heights.json')  # heights default: s10_heights.json
HERE = pathlib.Path(__file__).resolve().parent

# phone shell: s7s_build imports s7_build (both rewrite their S7 boards, unchanged)
sys.argv = ['s7s_build.py', str(ROOT)]
sys.path.insert(0, str(HERE))
P = importlib.import_module('s7s_build'); B = P.B
import sstate_css as SC
import sexsign as SXS  # SX01: ♀/♂ outside a figure span = the §72 hook (sexsign.py)
import realdata as R  # RF05: every figure, date, week and code from realdata (DATA03 extracts)
# desktop shell: the S4 farm page (s4s_build writes nothing on import)
sys.argv = ['s4s_build.py', str(ROOT)]
s4 = importlib.import_module('s4s_build')

n, dd, lamp, NB, MINUS = B.n, B.dd, B.lamp, B.NB, B.MINUS
B.IC.update({'sale': s4.IC['sale'], 'x': '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>'})
icon = B.icon

# ---------- data (realdata, RF05) ----------
DT, WK, HSTART = R.TODAY, R.WK, R.CYCLE_START  # 18.11.2026, SAPT 59, cycle from 22.02.2026
assert DT == B.TODAY and WK == B.WEEK == 59 and R.C26['status'] == 'open' and R.COUNTS_TELITA['totals']['all'] == 0  # Telița: disabled "goală"

class Sale:
    """One Vânzare form. rows = [(code, ♀, ♂ alive before the sale)] for the compartments with birds; sold = {code: (♀, ♂)};
    toate = codes filled with Toate (all alive), the other sold rows typed; empty = {hall: since} for halls without birds."""
    def __init__(s, day, rows, sold, toate=(), empty=None):
        s.day, s.rows, s.sold, s.toate, s.empty = day, rows, sold, set(toate), empty or {}
        s.before = (sum(f for _, f, _ in rows), sum(m for _, _, m in rows))
        s.tot = (sum(v[0] for v in sold.values()), sum(v[1] for v in sold.values()))
        s.after = (s.before[0] - s.tot[0], s.before[1] - s.tot[1])
        s.wk = R.sapt(day)
        assert all(sold[c] == (f, m) for c, f, m in rows if c in s.toate)

def alive_rows(d, plus=lambda c, sx: 0):
    return [(c, R.alive(c, 'F', d) + plus(c, 'F'), R.alive(c, 'M', d) + plus(c, 'M')) for c in R.HSECS
            if R.alive(c, 'F', d) + plus(c, 'F') + R.alive(c, 'M', d) + plus(c, 'M')]
EMPTY = {h: R.empty_since(h) for h in R.HHALLS if R.empty_since(h)}; assert list(EMPTY) == ['H7']  # since 14.11.2026
# 1 · the real 18.11.2026 sale (D155/D158b): typed counts, alive before it = end of day + sold that day (the 3 deaths came first)
SOLD18 = {r['section']: (r['F'], r['M']) for r in R.sales(DT)}
REAL = Sale(DT, alive_rows(DT, lambda c, sx: R.sold(DT, sec=c, sx=sx)), SOLD18, empty=EMPTY)
assert SOLD18 == {'H6.1': (676, 201), 'H6.2': (1155, 88)} and REAL.tot == (1831, 289)
assert REAL.before == (14392, 1083) and REAL.after == (R.alive_farm('F'), R.alive_farm('M')) == (12561, 794)
# 2 · emptying the whole farm on the 18.11 end counts (D39, D158b): Toate secțiunile, the farm reaches 0, the cycle closes
END = alive_rows(DT)
WHOLE = Sale(DT, END, {c: (f, m) for c, f, m in END}, toate=[c for c, _, _ in END], empty=EMPTY)
assert WHOLE.tot == REAL.after and WHOLE.after == (0, 0)
# 3 · refusal (D39): the 18.11 sale with H6.2 ♂ 98 typed over the 93 alive; the refused value stays out of the totals (D77)
RS, RTRY = 'H6.2', 98
RF, RM = next((f, m) for c, f, m in REAL.rows if c == RS); ROVER = RM - RTRY
assert (RM, ROVER) == (93, -5)
RTOT = (REAL.tot[0], REAL.tot[1] - SOLD18[RS][1])
# 4 · parțială: the real 14.11.2026 sale that emptied hall H7 (Toate on H7.1 + H7.2; H6.2's own sale that day is another
#   record, already in its alive), "today" = 14.11 (replay, D73); the farm keeps birds, the cycle continues (no Dialog)
D14 = EMPTY['H7']; PART = tuple(R.HALL_SECS['H7'])
SOLD14 = {r['section']: (r['F'], r['M']) for r in R.sales(D14) if r['section'] in PART}
HALL = Sale(D14, alive_rows(D14, lambda c, sx: SOLD14.get(c, (0, 0))['FM'.index(sx)]), SOLD14, toate=PART)
assert HALL.tot == (6915, 639) and HALL.wk == 58 and len(HALL.rows) == 6 and HALL.after == (R.alive_farm('F', D14), R.alive_farm('M', D14))
# desktop: 18.11 end-of-day counts, nothing typed
DESKS = Sale(DT, END, {}, empty=EMPTY)

def pair(f, m): return B.pair(n(f) if isinstance(f, int) else f, n(m) if isinstance(m, int) else m)
def roo(v): return f'<span class="ip-sex ip-sex--rooster">♂ <b>{n(v)}</b></span>'
def rec(k='sale', size=14): return f'<span class="ip-rec ip-rec--{k} s13-rec">{icon(k, size)}</span>'
def cmark(kind, cls, text):  # SYS06 cycle mark: bar (record that caused it) + mini block (open/closed)
    bar = '<i class="ip-cmark__bar"></i>' if cls else ''
    return f'<span class="ip-cmark ip-cmark--{kind}{(" " + cls) if cls else ""}"><span class="ip-cmark__g" aria-hidden="true">{bar}<i class="ip-cmark__blk"></i></span>{text}</span>'
CHK = '<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true"><path d="M20 6 9 17l-5-5"/></svg>'
def num(v): return f'<span class="ip-num">{v}</span>'

# ---------- form parts ----------
assert 'href="#inregistreaza" aria-label="Înapoi la Înregistrează"' in B.HEAD  # FX07/D142: the S7 shell's back link
def head(day):
    h = B.HEAD.replace('ip-rec--feed', 'ip-rec--sale').replace(icon('feed'), icon('sale')).replace('Hrană zilnică', 'Vânzare') \
        .replace(f'{B.FARM} · {dd(B.TODAY)} · SAPT {B.WEEK}', f'SERIA 2026 · {dd(day)}')
    assert h.count('Vânzare') == 1 and 'ip-rec--sale' in h and 'SAPT' not in h and dd(day) in h
    return h
HEAD, HEAD14 = head(DT), head(D14)

NSEC = len(R.HSECS)  # the farm's compartments (6), with or without birds
assert NSEC == 6 and len(REAL.rows) == len(END) == 4 and len(HALL.rows) == 6  # D168 hint: "6 secțiuni · 4 cu păsări" (14.11: toate)
def farm(sc, state='set', u='', touch=True):
    t = ' ip-input--touch' if touch else ''
    # saved: the counts are from before the sale (as the rows)
    # D168 (DATA04): the farm's compartments and how many hold birds before the sale (18.11: 6 · 4; 14.11 replay: all 6)
    withb = 'toate' if len(sc.rows) == NSEC else num(len(sc.rows))
    hint = f'<span class="ip-note" id="{u}farm-h">ouat · {num(NSEC)} secțiuni · {withb} cu păsări · în viață{" înainte" if state == "ro" else ""} {pair(*sc.before)}</span>'
    if state == 'ro':
        return f'<div class="ip-field"><label class="ip-label" for="{u}farm">Fermă</label><span class="ip-input{t} is-readonly"><input id="{u}farm" type="text" value="Hîrbovăț" readonly aria-describedby="{u}farm-h"/></span>{hint}</div>'
    # FIN10: options carry state (alive ♀ · ♂, Telița disabled "goală"), so Fermă is the section 53 Select: phone opens
    # the BottomSheet (aria-haspopup dialog, no list in the page), desktop its listbox (hidden while closed)
    lbl = f'<span class="ip-label" id="{u}farm-l">Fermă</span>'
    lab = f'aria-labelledby="{u}farm-l" aria-describedby="{u}farm-h"'
    if touch:
        return (f'<div class="ip-field">{lbl}<div class="ip-select"><div class="ip-input ip-select__btn{t}" id="{u}farm" role="combobox" tabindex="0" aria-haspopup="dialog" aria-expanded="false" {lab}>'
                f'<span class="ip-select__val">Hîrbovăț</span>{icon("chev", 16)}</div></div>{hint}</div>')
    O = '<span class="ip-select__txt"><span class="ip-select__top">'
    opts = (f'<li class="ip-menu__item ip-select__opt" role="option" id="{u}farm-o1" aria-selected="true"><span class="ip-select__chk" aria-hidden="true">{icon("check", 16)}</span>{O}Hîrbovăț {pair(*sc.before)}</span></span></li>'
            f'<li class="ip-menu__item ip-select__opt" role="option" id="{u}farm-o2" aria-selected="false" aria-disabled="true"><span class="ip-select__chk" aria-hidden="true"></span>{O}Telița <span class="ip-note">goală</span></span></span></li>')
    return (f'<div class="ip-field">{lbl}<div class="ip-select"><div class="ip-input ip-select__btn{t}" id="{u}farm" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="false" aria-controls="{u}farm-list" {lab}>'
            f'<span class="ip-select__val">Hîrbovăț</span>{icon("chev", 16)}</div>'
            f'<ul class="ip-menu ip-select__menu" role="listbox" id="{u}farm-list" aria-labelledby="{u}farm-l" tabindex="-1" hidden="hidden">{opts}</ul></div>{hint}</div>')

WKHINT = f'SAPT {WK} · ciclu din {dd(HSTART)}'
assert WKHINT in B.DATE and dd(DT) in B.DATE
DATE = B.DATE
DATE14 = B.DATE.replace(dd(DT), dd(D14)).replace(WKHINT, f'SAPT {HALL.wk} · ciclu din {dd(HSTART)}'); assert DATE14.count(dd(D14)) == 1
DATE_RO = f'''<div class="ip-field"><label class="ip-label" for="d-val">Data</label><span class="ip-input ip-input--touch ip-num is-readonly"><input id="d-val" type="text" value="{dd(DT)}" readonly aria-describedby="d-wk"/></span><span class="ip-note ip-num" id="d-wk">{WKHINT}</span></div>'''

def srow(sc, code, f, m, st, u='', touch=True):
    """One SexCountRow. st: all (Toate, read-only counts) · typed (sold counts typed, Toate off) · empty · ro (saved: the
    sold counts, '–' where none) · refuz (♂ typed over alive, Toate off)."""
    k = u + code.replace('.', '-'); md = '' if touch else ' ip-count--md'
    alive = f'<span class="ip-sexrow__alive" id="a-{k}"><span class="ip-sexrow__k">în viață{" înainte" if st == "ro" else ""}</span> {pair(f, m)}</span>'
    def lbl(sx):
        c, s, w = ('hen', '♀', 'găini') if sx == 'F' else ('rooster', '♂', 'cocoși')
        return f'<label class="ip-label ip-sexrow__lbl" for="{"f" if sx == "F" else "m"}-{k}"><span class="ip-sr">{code} </span><span class="ip-sex ip-sex--{c}" aria-hidden="true">{s}</span><span class="ip-sr">{w}</span></label>'
    def inp(sx, val=None, ro=False, cls='', extra=''):
        c = 'hen' if sx == 'F' else 'rooster'; i = f'{"f" if sx == "F" else "m"}-{k}'
        v = f' value="{n(val)}"' if val is not None else ''
        return f'<span class="ip-input ip-count ip-count--{c}{md}{cls}"><input id="{i}" inputmode="numeric" placeholder="–" autocomplete="off"{v}{" readonly" if ro else ""}{extra}/></span>'
    q = sc.sold.get(code)
    msg = ''
    if st == 'all':
        d = f' aria-describedby="a-{k} all-{k}"'
        fi, mi = inp('F', f, True, extra=d), inp('M', m, True, extra=d)
    elif st == 'typed':
        d = f' aria-describedby="a-{k}"'
        fi, mi = inp('F', q[0], extra=d), inp('M', q[1], extra=d)
    elif st == 'ro':
        d = f' aria-describedby="a-{k}"'
        fi, mi = inp('F', q[0] if q else None, True, ' is-readonly', d), inp('M', q[1] if q else None, True, ' is-readonly', d)
    elif st == 'refuz':
        fi = inp('F', q[0], extra=f' aria-describedby="a-{k}"')
        mi = inp('M', RTRY, cls=' is-error', extra=f' aria-invalid="true" aria-describedby="a-{k} e-{k}"')
        msg = (f'<p class="ip-err ip-sexrow__msg" id="e-{k}">{icon("alert", 14)}<span>{code}: {num(n(RTRY))} depășește cei {num(n(m))} cocoși în viață '
               f'pe {num(dd(sc.day))}. Introdu cel mult {num(n(m))}.</span></p>')
    else:
        d = f' aria-describedby="a-{k}"'
        fi, mi = inp('F', extra=d), inp('M', extra=d)
    if st == 'ro': allk = ''
    elif st == 'all': allk = f'<label class="ip-check ip-sexrow__all is-checked" id="all-{k}"><input type="checkbox" class="ip-sr" checked="checked"/><span class="ip-check__box">{CHK}</span><span class="ip-sr">{code} </span>Toate</label>'
    else: allk = f'<label class="ip-check ip-sexrow__all" id="all-{k}"><input type="checkbox" class="ip-sr"/><span class="ip-check__box"></span><span class="ip-sr">{code} </span>Toate</label>'
    return (f'<li class="ip-sexrow"><div class="ip-sexrow__head"><span class="ip-sexrow__code" id="c-{k}">{code}</span>{alive}</div>'
            f'<div class="ip-field is-hen">{lbl("F")}{fi}</div><div class="ip-field is-rooster">{lbl("M")}{mi}</div>{msg}{allk}</li>')

def erow(hall, since):  # D158d: an empty hall in a form = its code + one muted line, no inputs
    return (f'<li class="ip-sexrow"><div class="ip-sexrow__head"><span class="ip-sexrow__code">{hall}</span><span class="ip-sexrow__alive">'
            f'{lamp("off", "goală")}<span>din {num(dd(since))} · {", ".join(R.HALL_SECS[hall])}</span></span></div></li>')

MINUS_I = '<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12h14"/></svg>'
def allkey(state, u='', touch=True):
    """additions.css §34: the ghost "Toate secțiunile" key under the SexCountRow head (D84). aria-pressed true = every row has
    Toate on, mixed = some rows, false = none; the box is the system Checkbox box (check / minus glyph)."""
    glyph = {'true': CHK, 'mixed': MINUS_I, 'false': ''}[state]
    t = ' ip-btn--touch' if touch else ''
    return (f'<button type="button" class="ip-btn ip-btn--ghost{t} ip-sexrows__allkey" id="{u}allk" aria-pressed="{state}">'
            f'<span class="ip-check__box" aria-hidden="true">{glyph}</span>Toate secțiunile<span class="ip-sr"> din Hîrbovăț</span></button>')

def rows(sc, mode='set', u='', touch=True):
    """mode: set (each row as the sale fills it: Toate rows, typed rows, the rest empty) · empty (nothing typed) · ro (saved) ·
    refuz (the sale with RS ♂ typed over alive)."""
    def st(code):
        if mode in ('empty', 'ro'): return mode
        if mode == 'refuz' and code == RS: return 'refuz'
        return 'all' if code in sc.toate else 'typed' if code in sc.sold else 'empty'
    items = [srow(sc, c, f, m, st(c), u, touch) for c, f, m in sc.rows] + [erow(h, d) for h, d in sc.empty.items()]
    nt = len(sc.toate) if mode not in ('empty', 'ro') else 0
    key = '' if mode == 'ro' else allkey('true' if nt == len(sc.rows) else 'mixed' if nt else 'false', u, touch)
    if mode == 'empty':
        filled, tot, sub = 0, ('–', '–'), f'în viață {pair(*sc.before)}'
    elif mode == 'refuz':
        filled, tot, sub = len(sc.sold), RTOT, f'fără {RS} {SXS.sign("M")} refuzat'
    else:
        filled, tot, sub = len(sc.sold), sc.tot, f'în viață {"înainte " if mode == "ro" else ""}{pair(*sc.before)}'
    lg = f'<span class="ip-pair ip-pair--lg">{B.sexv("F", n(tot[0]) if isinstance(tot[0], int) else tot[0])}<span class="sep">·</span>{B.sexv("M", n(tot[1]) if isinstance(tot[1], int) else tot[1])}</span>'
    total = (f'<div class="ip-sexrows__total"><div class="ip-sexrows__tk"><span class="ip-cap">Total vânzare</span><span class="ip-sexrows__sub">'
             f'<span>{num(filled)} din {num(len(sc.rows))} secțiuni</span> · {sub}</span></div>{lg}</div>')
    return (f'<section class="ip-panel ip-sexrows" aria-labelledby="{u}t-px"><div class="ip-sexrows__head"><h2 class="ip-cap" id="{u}t-px">Păsări vândute</h2>'
            f'<span class="ip-note">Hîrbovăț · {num(dd(sc.day))}</span></div>{key}<ol class="ip-sexrows__list">{"".join(items)}</ol>{total}</section>')

def buyer(ro=False, u='', touch=True):
    if ro: return '<div class="ip-field"><span class="ip-label">Cumpărător</span><span class="ip-note">– fără cumpărător</span></div>'
    t = ' ip-input--touch' if touch else ''
    return f'<div class="ip-field"><label class="ip-label" for="{u}buyer">Cumpărător <span class="s13-opt">(opțional)</span></label><span class="ip-input{t}"><input id="{u}buyer" type="text" autocomplete="off"/></span></div>'
def note(ro=False, u=''):
    if ro: return '<div class="ip-field"><span class="ip-label">Notă</span><span class="ip-note">– fără notă</span></div>'
    return f'<div class="ip-field"><label class="ip-label" for="{u}note">Notă <span class="s13-opt">(opțional)</span></label><span class="ip-input ip-input--area"><textarea id="{u}note" rows="2"></textarea></span></div>'

ARROW = '<span class="s9-to" aria-hidden="true">→</span><span class="ip-sr">după vânzare</span>'
def close_txt(sc): return f'Ciclu închis la Hîrbovăț, sapt. {sc.wk}'
def effect(sc, state='set', u=''):
    if state == 'empty':
        return (f'<div class="ip-effect s13-effect" id="{u}fx" role="status">{rec("sale", 16)}<div class="s9-fxbody"><span class="ip-cap">Efect</span>'
                '<span>Scrie câte păsări vinzi din fiecare secțiune sau alege Toate secțiunile. Aici vezi ce rămâne în fermă și dacă ciclul se închide.</span></div></div>')
    if state == 'refuz':  # EffectPreview: refused becomes a fault Banner
        return (f'<div class="ip-banner ip-banner--fault s7s-banner" role="alert" id="{u}fx">{icon("alert", 16)}<div class="s7s-banner__body"><p class="ip-m0">'
                f'Vânzare refuzată: Hîrbovăț › {RS} ar ajunge la {roo(ROVER)} pe {num(dd(sc.day))}. Nicio secțiune nu poate scădea sub{NB}0.</p></div></div>')
    codes = [c for c, _, _ in sc.rows if c in sc.sold]
    if sum(sc.after):  # the farm keeps birds: the cycle continues (open mini block, no record bar, as S9 D83)
        sub = f'{num(len(codes))} din {num(len(sc.rows))} secțiuni · {", ".join(codes)}'
        cyc = f'{cmark("start", "", f"Ciclul la Hîrbovăț continuă, sapt. {sc.wk}")}<span class="s9-fxsub">rămân {pair(*sc.after)}</span>'
    else:
        sub = f'{num(len(sc.rows))} din {num(len(sc.rows))} secțiuni'
        cyc = f'{cmark("end", "sale", close_txt(sc))}<span class="s9-fxsub">ferma ajunge la 0 păsări</span>'
    return f'''<div class="ip-effect s13-effect" id="{u}fx" role="status">{rec("sale", 16)}<div class="s9-fxbody"><span class="ip-cap">Efect</span>
<span class="s9-fxline"><b class="s9-strong">Hîrbovăț</b><span class="s9-fxval">{pair(*sc.before)}{ARROW}{pair(*sc.after)}</span></span>
<span class="s9-fxsub">{sub}</span>
<span class="s9-fxcyc">{cyc}</span>
</div></div>'''

def sum1(tot, lead=''):
    # no truck icon and no farm name: with the long key (Înregistrează vânzarea) the summary has ~130 px, the pair alone fits
    # one line; the sale colour stays on line 2 (the cycle mark's sale bar); the farm is in the form and the header
    return f'<span>{(lead + " · ") if lead else ""}{pair(*tot)}</span>'
SUM2 = f'<span>{cmark("end", "sale", "ciclul se închide")}</span>'
SUM2_CONT = f'<span>{cmark("start", "", "ciclul continuă")}</span>'
SELL_DLG = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" aria-haspopup="dialog">Înregistrează vânzarea</button>'
SELL = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Înregistrează vânzarea</button>'
SELL_OFF = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="sb-sum">Înregistrează vânzarea</button>'

def screen(parts, save, overlay='', view='', modal='', head=None):
    """One 390×844 viewport (D60). view: ' ip-screen__view--end' scrolls the form to its end. modal: a Dialog over the screen;
    the header, form and SaveBar behind it are inert (out of the tab order and the a11y tree)."""
    h, ine = head or HEAD, ''
    if modal:
        h = h.replace('<header class="ip-formhead">', '<header class="ip-formhead" inert>', 1)
        save = save.replace('<div class="ip-savebar', '<div inert class="ip-savebar', 1); ine = ' inert'
    return f'''<div class="ip-screen ip-screen--fixed{" ip-scrimhost" if modal else ""}">
{h}
<div class="ip-screen__view{view}"{ine}>
<main class="ip-screen__main">
{"".join(parts)}
</main>
{overlay}
</div>
{save}{modal}
</div>'''

def tall(parts, save):  # the default board: the whole form, one tall phone board (canvas.md), SaveBar sticky at the foot
    return f'''<div class="ip-screen">
{HEAD}
<main class="ip-screen__main">
{"".join(parts)}
</main>
{save}
</div>'''

def form(sc, mode='set', date=DATE, fx=None): return [farm(sc), date, rows(sc, mode), buyer(), note(), fx or effect(sc)]
boards = []
# 1 · telefon — the real 18.11.2026 sale typed (H6.1, H6.2), before Înregistrează vânzarea: the farm keeps birds, the cycle
#   continues, so the key saves directly (no Dialog) (whole form, tall board)
boards.append(('', f'telefon · vânzarea din {dd(DT)}', tall(form(REAL), B.savebar(sum1(REAL.tot), SUM2_CONT, SELL)), 'tall'))
# 2 · confirmare — emptying the whole farm on the 18.11 end counts (D39, D158b): Toate secțiunile, the sale empties the farm,
#   so Înregistrează vânzarea opens the confirmation Dialog (C4); focus on Anulează
DLG = f'''
<!-- Dialog (modal alertdialog): focus starts on Anulează (craft §5); Esc or Anulează closes it and focus returns to Înregistrează
     vânzarea. The form behind is inert. -->
<div class="ip-scrim ip-scrim--phone"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-note dlg-fx">
<h2 id="dlg-t">Înregistrează vânzarea și închide ciclul la Hîrbovăț</h2>
<p class="ip-dialog__note" id="dlg-note">Vinzi toate păsările din Hîrbovăț pe {num(dd(DT))}: {pair(*WHOLE.tot)}. Ferma ajunge la 0 păsări și ciclul ei se închide.</p>
<div class="ip-effect s13-effect" id="dlg-fx">{rec("sale", 16)}<div class="s9-fxbody"><span class="s9-fxcyc">{cmark("end", "sale", close_txt(WHOLE))}<span class="s9-fxsub">Seria 2026 · {num(R.drange(HSTART, DT))}</span></span></div></div>
<div class="ip-dialog__keys ip-dialog__keys--stack"><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-focus" autofocus="autofocus">Anulează</button><button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Înregistrează vânzarea</button></div>
</div></div>'''
boards.append(('confirmare', 'confirmare · golește ferma · exemplu de stare (D39)', screen(form(WHOLE), B.savebar(sum1(WHOLE.tot), SUM2, SELL_DLG), view=' ip-screen__view--end', modal=DLG), 'vp'))
# 3 · refuzat — the 18.11 sale with H6.2 ♂ 98 typed over the 93 alive (D39 state example): row error, preview = fault Banner,
#   key disabled; the refused value is left out of the totals and the SaveBar (D77), line 2 names it
# S10P: while refused the bar carries only the reason, in two lines (the key is disabled, its aria-describedby reads it);
#   the totals stay in the form's total line, minus the refused value
ref_save = B.savebar(f'{lamp("fault", "refuzat")}<span>{RS}</span>', f'<span>cel mult {roo(RM)}</span>', SELL_OFF)
boards.append(('refuzat', 'refuzat · peste cei în viață · exemplu de stare (D39)', screen(form(REAL, 'refuz', fx=effect(REAL, 'refuz')), ref_save, view=' ip-screen__view--end'), 'vp'))
# 4 · salvat — the real sale saved: read-only form scrolled to its end, Toast with Anulează (undo the same day), Vânzare nouă
toast = SXS.sexify(f'<div class="ip-toastwrap"><div class="ip-toast" role="status">{icon("check", 16)}<span class="ip-toast__txt">Vânzare înregistrată · '
         f'<span class="ip-nowrap">♀ {num(n(REAL.tot[0]))}</span> · <span class="ip-nowrap">♂ {num(n(REAL.tot[1]))}</span> · Hîrbovăț · ciclul continuă</span><button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
saved = [farm(REAL, 'ro'), DATE_RO, rows(REAL, 'ro'), buyer(True), note(True), effect(REAL)]
boards.append(('salvat', 'salvat', screen(saved, B.savebar(sum1(REAL.tot, lead='vândut'), f'<span>poți anula până la {num("23:59")}</span>',
              '<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Vânzare nouă</button>'), overlay=toast, view=' ip-screen__view--end'), 'vp'))
# 5 · eroare la salvare — the real sale; online only (spec §1): fault Banner + Reîncearcă as the form's last part, right above
#   the SaveBar (D141/D181, as S6/S7/S8/S13/S16; was above the EffectPreview), form scrolled to its end (as S9), typed values
#   kept as a draft (D60), key stays enabled
ban = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{icon("alert", 16)}<div class="ip-banner__body"><p class="ip-m0">Vânzarea nu s-a salvat: '
       f'{pair(*REAL.tot)}, Hîrbovăț, {num(dd(DT))}.</p>'
       f'<p class="ip-m0">Verifică conexiunea la internet, apoi apasă Reîncearcă. Ce ai scris rămâne păstrat în browser până la salvare.</p></div><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-banner__key ip-focus" autofocus="autofocus">Reîncearcă</button></div>')
boards.append(('eroare-la-salvare', 'eroare la salvare', screen(form(REAL)[:5] + [effect(REAL), ban], B.savebar(sum1(REAL.tot), SUM2_CONT, SELL), view=' ip-screen__view--end'), 'vp'))
# 6 · parțială — the real 14.11.2026 sale that emptied hall H7 (Toate on H7.1 and H7.2), replayed with "today" = 14.11: the farm
#   keeps birds, the cycle continues, so Înregistrează vânzarea saves directly (no Dialog); form scrolled to its end (the preview)
boards.append(('partiala', f'parțială · golește hala H7 · reluare {dd(D14)}', screen(form(HALL, date=DATE14),
              B.savebar(sum1(HALL.tot), SUM2_CONT, SELL), view=' ip-screen__view--end', head=HEAD14), 'vp'))

# the phone Dialog is additions.css section 35 (ip-scrim--phone, FIN04); the select chevron is additions.css (FIN10 fix 1)
CSS = P.CSS + SC.MAIN + SC.REC + SC.OPT + SC.EFFECT + SC.FX + SC.MODAL

# ---------- desktop: SidePanel beside S4 · Hîrbovăț (manager), 18.11.2026, nothing typed (end-of-day counts) ----------
DDATE = f'''<div class="ip-field"><label class="ip-label" for="p-date">Data</label><div class="ip-row s4s-daterow"><span class="ip-seg" role="radiogroup" aria-label="Zi rapidă"><span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Azi</span><span class="ip-seg__item" role="radio" aria-checked="false" tabindex="-1">Ieri</span></span><span class="ip-input s4s-date"><input id="p-date" type="text" inputmode="numeric" value="{dd(DT)}" placeholder="zz.ll.aaaa" aria-describedby="p-date-h"/><button type="button" class="ip-iconbtn" aria-label="Alege data vânzării din calendar" aria-haspopup="dialog" aria-expanded="false">{icon("cal", 16)}</button></span></div><span class="ip-hint ip-num" id="p-date-h">SAPT {WK} · ciclu din {dd(HSTART)}</span></div>'''
PANEL = f'''<aside class="ip-sidepanel ip-sidepanel--dock" id="vp" aria-labelledby="vp-t">
<div class="ip-sidepanel__head"><div class="ip-row ip-sidepanel__title"><span class="ip-tile ip-rec ip-rec--sale" aria-hidden="true">{icon("sale")}</span><div><h2 class="ip-t-heading ip-m0" id="vp-t">Vânzare</h2><p class="ip-note ip-num ip-m0">SERIA 2026 · {dd(DT)}</p></div></div><button type="button" class="ip-iconbtn" aria-label="Închide Vânzare">{icon("x")}</button></div>
<div class="ip-sidepanel__body ip-sidepanel__body--form">
{farm(DESKS, u='p', touch=False)}
{DDATE}
{rows(DESKS, 'empty', u='p', touch=False)}
{buyer(u='p', touch=False)}
{note(u='p')}
{effect(DESKS, 'empty', u='p')}
</div>
<div class="ip-sidepanel__foot"><span class="ip-note" id="vp-why">Scrie câte vinzi</span><div class="ip-row ip-sidepanel__keys"><button type="button" class="ip-btn ip-btn--secondary">Anulează</button><button type="button" class="ip-btn ip-btn--primary" disabled aria-describedby="vp-why">Înregistrează vânzarea</button></div></div>
</aside>'''
KEY = f'<button type="button" class="ip-btn ip-btn--secondary">{s4.icon("sale", 16)}Vânzare</button>'
SPG = s4.shell(s4.main_page(panel=s4.panel_default())); assert SPG.count(KEY) == 1
SPG = SPG.replace(KEY, KEY.replace('ip-btn--secondary"', 'ip-btn--secondary is-pressed" aria-expanded="true" aria-controls="vp"'))
DESK = f'<div class="ip-dockpage">{SPG}</div>'
CSS_D = s4.CSS + '\n' + s4.CSS_STATES + SC.SIDE + SC.M0 + SC.REC + SC.OPT + SC.EFFECT + SC.FX

# ---------- write ----------
STEM0, TITLE0 = 'S10-vanzare', 'S10 · Vânzare'
HEIGHTS = json.loads(HP.read_text()) if HP and HP.exists() else {}
VH = P.VH
def write_phone(slug, state, body, kind):
    stem = STEM0 + (f'--{slug}' if slug else ''); title = f'{TITLE0} · {state}'
    h = HEIGHTS.get(stem, 2400) if kind == 'tall' else VH
    root = (f'<div class="ip-root" style="padding:0; width:390px; min-height:{h}px">{body}</div>' if kind == 'tall'
            else f'<div class="ip-root" style="padding:0; width:390px; height:{VH}px; overflow:hidden">{body}</div>')
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
<style>{CSS}</style>
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
    # RF05: the render wrapper of the tall board is not pinned to the stored h (keeps the 844 phone floor), so
    # s10_render.py --heights measures the form's natural height (as s4s_build.write, RF03)
    wroot = root.replace(f'min-height:{h}px', f'min-height:{VH}px', 1) if kind == 'tall' else root
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{B.FONTS}<style>{CSS}</style></head><body>{wroot}</body></html>'''
    pathlib.Path('/tmp/s10').mkdir(exist_ok=True)
    pathlib.Path(f'/tmp/s10/{stem}.html').write_text(wrap, encoding='utf-8')
    print('ok', stem, title, h)
    return stem, title, h, 390

out = [write_phone(*b) for b in boards]
DH = SC.SIDE_H
s4.write(STEM0 + '--desktop', TITLE0 + ' · desktop', DESK, PANEL, H=DH, css=CSS_D)
out.append((STEM0 + '--desktop', TITLE0 + ' · desktop', DH, 1440))
if HP:
    HP.write_text(json.dumps({s: h for s, _, h, _ in out}, indent=1))

if CANVAS:  # S10 row, entries keep x/y, a new row goes below the lowest (canvas_rows.py); 80 px gaps
    c = CR.load(ROOT)
    x = CR.place(c, STEM0, [(s, w, h, t) for s, t, h, w in out])
    # keep the S10 entries of `order` in board order (S10P inserted the partial board before the desktop one)
    mine = [f'{s}.dc.html' for s, _, _, _ in out]; pos = [i for i, k in enumerate(c['order']) if k in mine]
    for i, k in zip(pos, mine): c['order'][i] = k
    CR.save(ROOT, c)
    print('canvas: S10 row y', CR.row_y(c, STEM0), 'ends at x', x - 80)
