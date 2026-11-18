"""S10 · Vânzare (manager, spec C4): Fermă, then Data; one SexCountRow per section with ♀/♂ counts and Toate; optional
Cumpărător and Notă; the EffectPreview (farm before → after, what remains, the cycle that closes); Înregistrează vânzarea;
a confirmation Dialog when the sale empties the farm (C4 › Confirmation).
Phone boards reuse the S7 phone form shell by import (s7s_build → s7_build: FormHeader, DateField, icons, SaveBar shell +
2-line summary, viewport/Toast/Banner CSS from sstate_css), S13's record-icon/EffectPreview/"(opțional)" rules and S9's
EffectPreview line rules + modal anchor (sstate_css, s13-/s9- names kept); the rows are the system SexCountRow (SYS07,
stacked phone layout), the cycle marks the CycleTrack vocabulary (`ip-cmark--end sale`, SYS06), the Dialog the system
Dialog in the phone scrim (additions.css section 35, ip-scrim--phone, keys stacked; FIN04). The desktop board reuses
the S4 Hîrbovăț farm page (s4s_build shell) with the system SidePanel beside it (non-modal, D73), "Vânzare" pressed.
Content: the extracts hold no sale. Hîrbovăț alive on 16.11.2026 (D23) is the sample (counts-latest, D27): ♀ 25 252 · ♂ 2 685
in 6 sections. The phone boards show a D39 state example consistent with it: the whole flock sold on 16.11.2026 (Toate on
every row), so the farm reaches 0 and its Seria 2026 cycle (from 10.02.2026, D25) closes in sapt. 40; refusal = H4.2 ♂ 500
over the 476 alive (the D39 AlertRow example). The desktop board types nothing (alive per section from the sample only).
S10P: the SexCountRow head gets the ghost "Toate secțiunile" key (additions.css §34, D84); a partial sale board (Toate on
H4.1 and H4.2 only, all figures from counts-latest: the cycle continues); the refused/saved SaveBars fit two lines.
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
# desktop shell: the S4 farm page (s4s_build writes nothing on import)
sys.argv = ['s4s_build.py', str(ROOT)]
s4 = importlib.import_module('s4s_build')

n, dd, lamp, NB, MINUS = B.n, B.dd, B.lamp, B.NB, B.MINUS
B.IC.update({'sale': s4.IC['sale'], 'x': '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>'})
icon = B.icon

# ---------- data ----------
EX = ROOT / 'design/data/extracts'
CY = json.load(open(EX / 'cycles.json', encoding='utf-8'))
CL = json.load(open(EX / 'counts-latest.json', encoding='utf-8'))
FS = json.load(open(EX / 'farms-sections.json', encoding='utf-8'))
farms = {f['farm_id']: f for f in CL['farms']}
HB = farms['hirbovat']
SEC = [(s['section'], s['F'], s['M']) for s in HB['sections']]
TOT = (HB['totals']['F'], HB['totals']['M'])
assert [s[0] for s in SEC] == ['H1.1', 'H1.2', 'H3.1', 'H3.2', 'H4.1', 'H4.2']
assert TOT == (25252, 2685) == (sum(s[1] for s in SEC), sum(s[2] for s in SEC))
assert farms['telita']['totals']['all'] == 0  # Telița is empty (cycle closed 28.02.2026): the farm Select lists it disabled
HIR = next(c for c in CY['cycles'] if c['id'] == 'cyc-hirbovat-2026')
HSTART = dt.date.fromisoformat(HIR['start_date'])
DT = B.TODAY; WK = (DT - HSTART).days // 7 + 1
assert dd(DT) == '16.11.2026' and dd(HSTART) == '10.02.2026' and WK == 40 == HB['cycle_week'] and HIR['status'] == 'open'
# refusal (D39 state example, as the AlertRow "refuzat" preview): H4.2 ♂ 500 typed over the 476 alive
RS, RTRY = 'H4.2', 500
RF, RM = next((f, m) for s, f, m in SEC if s == RS); ROVER = RM - RTRY
assert (RM, ROVER) == (476, -24)
RTOT = (TOT[0], TOT[1] - RM)  # the refused value is left out of the totals (D77), the totals line says so
# partial sale (D39 state example, S10P): Toate on hall H4 only (H4.1 + H4.2 alive from counts-latest), the other rows empty;
#   the farm keeps birds, so the cycle continues (no Dialog)
PART = ('H4.1', 'H4.2')
PSOLD = (sum(f for s, f, m in SEC if s in PART), sum(m for s, f, m in SEC if s in PART))
PLEFT = (TOT[0] - PSOLD[0], TOT[1] - PSOLD[1])
assert PSOLD == (8224, 956) and PLEFT == (17028, 1729)

def pair(f, m): return B.pair(n(f) if isinstance(f, int) else f, n(m) if isinstance(m, int) else m)
def roo(v): return f'<span class="ip-sex ip-sex--rooster">♂ <b>{n(v)}</b></span>'
def rec(k='sale', size=14): return f'<span class="ip-rec ip-rec--{k} s13-rec">{icon(k, size)}</span>'
def cmark(kind, cls, text):  # SYS06 cycle mark: bar (record that caused it) + mini block (open/closed)
    bar = '<i class="ip-cmark__bar"></i>' if cls else ''
    return f'<span class="ip-cmark ip-cmark--{kind}{(" " + cls) if cls else ""}"><span class="ip-cmark__g" aria-hidden="true">{bar}<i class="ip-cmark__blk"></i></span>{text}</span>'
CHK = '<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true"><path d="M20 6 9 17l-5-5"/></svg>'
def num(v): return f'<span class="ip-num">{v}</span>'

# ---------- form parts ----------
HEAD = B.HEAD.replace('href="#acasa" aria-label="Înapoi la Acasă"', 'href="#inregistreaza" aria-label="Înapoi la Înregistrează"') \
    .replace('ip-rec--feed', 'ip-rec--sale').replace(icon('feed'), icon('sale')).replace('Hrană zilnică', 'Vânzare') \
    .replace(f'{B.FARM} · {dd(B.TODAY)} · SAPT {B.WEEK}', f'SERIA 2026 · {dd(DT)}')
assert HEAD.count('Vânzare') == 1 and 'ip-rec--sale' in HEAD and 'SAPT' not in HEAD

def farm(state='set', u='', touch=True):
    t = ' ip-input--touch' if touch else ''
    # saved: the farm is empty after the sale, so the hint says the counts are from before it (as the rows)
    hint = f'<span class="ip-note" id="{u}farm-h">ouat · {num(len(SEC))} secțiuni · în viață{" înainte" if state == "ro" else ""} {pair(*TOT)}</span>'
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
    opts = (f'<li class="ip-menu__item ip-select__opt" role="option" id="{u}farm-o1" aria-selected="true"><span class="ip-select__chk" aria-hidden="true">{icon("check", 16)}</span>{O}Hîrbovăț {pair(*TOT)}</span></span></li>'
            f'<li class="ip-menu__item ip-select__opt" role="option" id="{u}farm-o2" aria-selected="false" aria-disabled="true"><span class="ip-select__chk" aria-hidden="true"></span>{O}Telița <span class="ip-note">goală</span></span></span></li>')
    return (f'<div class="ip-field">{lbl}<div class="ip-select"><div class="ip-input ip-select__btn{t}" id="{u}farm" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="false" aria-controls="{u}farm-list" {lab}>'
            f'<span class="ip-select__val">Hîrbovăț</span>{icon("chev", 16)}</div>'
            f'<ul class="ip-menu ip-select__menu" role="listbox" id="{u}farm-list" aria-labelledby="{u}farm-l" tabindex="-1" hidden="hidden">{opts}</ul></div>{hint}</div>')

DATE = B.DATE
DATE_RO = f'''<div class="ip-field"><label class="ip-label" for="d-val">Data</label><span class="ip-input ip-input--touch ip-num is-readonly"><input id="d-val" type="text" value="{dd(DT)}" readonly aria-describedby="d-wk"/></span><span class="ip-note ip-num" id="d-wk">SAPT {WK} · ciclu din {dd(HSTART)}</span></div>'''
assert f'SAPT {WK} · ciclu din {dd(HSTART)}' in DATE and dd(DT) in DATE

def srow(code, f, m, st, u='', touch=True):
    """One SexCountRow. st: all (Toate, read-only counts) · empty · ro (saved) · refuz (♂ over alive, Toate off)."""
    k = u + code.replace('.', '-'); md = '' if touch else ' ip-count--md'
    alive = f'<span class="ip-sexrow__alive" id="a-{k}"><span class="ip-sexrow__k">în viață{" înainte" if st == "ro" else ""}</span> {pair(f, m)}</span>'
    def lbl(sx):
        c, s, w = ('hen', '♀', 'găini') if sx == 'F' else ('rooster', '♂', 'cocoși')
        return f'<label class="ip-label ip-sexrow__lbl" for="{"f" if sx == "F" else "m"}-{k}"><span class="ip-sr">{code} </span><span class="ip-sex ip-sex--{c}" aria-hidden="true">{s}</span><span class="ip-sr">{w}</span></label>'
    def inp(sx, val=None, ro=False, cls='', extra=''):
        c = 'hen' if sx == 'F' else 'rooster'; i = f'{"f" if sx == "F" else "m"}-{k}'
        v = f' value="{n(val)}"' if val is not None else ''
        return f'<span class="ip-input ip-count ip-count--{c}{md}{cls}"><input id="{i}" inputmode="numeric" placeholder="–" autocomplete="off"{v}{" readonly" if ro else ""}{extra}/></span>'
    msg = ''
    if st == 'all':
        d = f' aria-describedby="a-{k} all-{k}"'
        fi, mi = inp('F', f, True, extra=d), inp('M', m, True, extra=d)
    elif st == 'ro':
        d = f' aria-describedby="a-{k}"'
        fi, mi = inp('F', f, True, ' is-readonly', d), inp('M', m, True, ' is-readonly', d)
    elif st == 'refuz':
        fi = inp('F', f, extra=f' aria-describedby="a-{k}"')
        mi = inp('M', RTRY, cls=' is-error', extra=f' aria-invalid="true" aria-describedby="a-{k} e-{k}"')
        msg = (f'<p class="ip-err ip-sexrow__msg" id="e-{k}">{icon("alert", 14)}<span>{code}: {num(n(RTRY))} depășește cei {num(n(m))} cocoși în viață '
               f'pe {num(dd(DT))}. Introdu cel mult {num(n(m))}.</span></p>')
    else:
        d = f' aria-describedby="a-{k}"'
        fi, mi = inp('F', extra=d), inp('M', extra=d)
    if st == 'ro': allk = ''
    elif st == 'all': allk = f'<label class="ip-check ip-sexrow__all is-checked" id="all-{k}"><input type="checkbox" class="ip-sr" checked="checked"/><span class="ip-check__box">{CHK}</span><span class="ip-sr">{code} </span>Toate</label>'
    else: allk = f'<label class="ip-check ip-sexrow__all" id="all-{k}"><input type="checkbox" class="ip-sr"/><span class="ip-check__box"></span><span class="ip-sr">{code} </span>Toate</label>'
    return (f'<li class="ip-sexrow"><div class="ip-sexrow__head"><span class="ip-sexrow__code" id="c-{k}">{code}</span>{alive}</div>'
            f'<div class="ip-field is-hen">{lbl("F")}{fi}</div><div class="ip-field is-rooster">{lbl("M")}{mi}</div>{msg}{allk}</li>')

MINUS_I = '<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12h14"/></svg>'
def allkey(state, u='', touch=True):
    """additions.css §34: the ghost "Toate secțiunile" key under the SexCountRow head (D84). aria-pressed true = every row has
    Toate on, mixed = some rows, false = none; the box is the system Checkbox box (check / minus glyph)."""
    glyph = {'true': CHK, 'mixed': MINUS_I, 'false': ''}[state]
    t = ' ip-btn--touch' if touch else ''
    return (f'<button type="button" class="ip-btn ip-btn--ghost{t} ip-sexrows__allkey" id="{u}allk" aria-pressed="{state}">'
            f'<span class="ip-check__box" aria-hidden="true">{glyph}</span>Toate secțiunile<span class="ip-sr"> din Hîrbovăț</span></button>')

def rows(mode='all', u='', touch=True):
    """mode: all (whole farm, Toate on every row) · empty · ro · refuz (all rows Toate, except H4.2 typed with ♂ 500) ·
    part (Toate on H4.1 and H4.2 only, the other rows empty)."""
    items = []
    for code, f, m in SEC:
        st = {'refuz': 'refuz' if code == RS else 'all', 'part': 'all' if code in PART else 'empty'}.get(mode, mode)
        items.append(srow(code, f, m, st, u, touch))
    key = '' if mode == 'ro' else allkey({'all': 'true', 'empty': 'false'}.get(mode, 'mixed'), u, touch)
    if mode == 'empty':
        filled, tot, sub = 0, ('–', '–'), f'în viață {pair(*TOT)}'
    elif mode == 'refuz':
        filled, tot, sub = len(SEC), RTOT, f'fără {RS} ♂ refuzat'
    elif mode == 'part':
        filled, tot, sub = len(PART), PSOLD, f'în viață {pair(*TOT)}'
    else:
        filled, tot, sub = len(SEC), TOT, f'în viață {"înainte " if mode == "ro" else ""}{pair(*TOT)}'
    lg = f'<span class="ip-pair ip-pair--lg">{B.sexv("F", n(tot[0]) if isinstance(tot[0], int) else tot[0])}<span class="sep">·</span>{B.sexv("M", n(tot[1]) if isinstance(tot[1], int) else tot[1])}</span>'
    total = (f'<div class="ip-sexrows__total"><div class="ip-sexrows__tk"><span class="ip-cap">Total vânzare</span><span class="ip-sexrows__sub">'
             f'<span>{num(filled)} din {num(len(SEC))} secțiuni</span> · {sub}</span></div>{lg}</div>')
    return (f'<section class="ip-panel ip-sexrows" aria-labelledby="{u}t-px"><div class="ip-sexrows__head"><h2 class="ip-cap" id="{u}t-px">Păsări vândute</h2>'
            f'<span class="ip-note">Hîrbovăț · {num(dd(DT))}</span></div>{key}<ol class="ip-sexrows__list">{"".join(items)}</ol>{total}</section>')

def buyer(ro=False, u='', touch=True):
    if ro: return '<div class="ip-field"><span class="ip-label">Cumpărător</span><span class="ip-note">– fără cumpărător</span></div>'
    t = ' ip-input--touch' if touch else ''
    return f'<div class="ip-field"><label class="ip-label" for="{u}buyer">Cumpărător <span class="s13-opt">(opțional)</span></label><span class="ip-input{t}"><input id="{u}buyer" type="text" autocomplete="off"/></span></div>'
def note(ro=False, u=''):
    if ro: return '<div class="ip-field"><span class="ip-label">Notă</span><span class="ip-note">– fără notă</span></div>'
    return f'<div class="ip-field"><label class="ip-label" for="{u}note">Notă <span class="s13-opt">(opțional)</span></label><span class="ip-input ip-input--area"><textarea id="{u}note" rows="2"></textarea></span></div>'

ARROW = '<span class="s9-to" aria-hidden="true">→</span><span class="ip-sr">după vânzare</span>'
CLOSE = f'Ciclu închis la Hîrbovăț, sapt. {WK}'
def effect(state='all', u=''):
    if state == 'empty':
        return (f'<div class="ip-effect s13-effect" id="{u}fx" role="status">{rec("sale", 16)}<div class="s9-fxbody"><span class="ip-cap">Efect</span>'
                '<span>Scrie câte păsări vinzi din fiecare secțiune sau alege Toate secțiunile. Aici vezi ce rămâne în fermă și dacă ciclul se închide.</span></div></div>')
    if state == 'refuz':  # EffectPreview: refused becomes a fault Banner
        return (f'<div class="ip-banner ip-banner--fault s7s-banner" role="alert" id="{u}fx">{icon("alert", 16)}<div class="s7s-banner__body"><p class="ip-m0">'
                f'Vânzare refuzată: Hîrbovăț › {RS} ar ajunge la {roo(ROVER)} pe {num(dd(DT))}. Nicio secțiune nu poate scădea sub{NB}0.</p></div></div>')
    if state == 'part':  # the farm keeps birds: the cycle continues (open mini block, no record bar, as S9 D83)
        after, sub = pair(*PLEFT), f'{num(len(PART))} din {num(len(SEC))} secțiuni · {", ".join(PART)}'
        cyc = f'{cmark("start", "", f"Ciclul la Hîrbovăț continuă, sapt. {WK}")}<span class="s9-fxsub">rămân {pair(*PLEFT)}</span>'
    else:
        after, sub = pair(0, 0), f'{num(len(SEC))} din {num(len(SEC))} secțiuni'
        cyc = f'{cmark("end", "sale", CLOSE)}<span class="s9-fxsub">ferma ajunge la 0 păsări</span>'
    return f'''<div class="ip-effect s13-effect" id="{u}fx" role="status">{rec("sale", 16)}<div class="s9-fxbody"><span class="ip-cap">Efect</span>
<span class="s9-fxline"><b class="s9-strong">Hîrbovăț</b><span class="s9-fxval">{pair(*TOT)}{ARROW}{after}</span></span>
<span class="s9-fxsub">{sub}</span>
<span class="s9-fxcyc">{cyc}</span>
</div></div>'''

def sum1(tot=TOT, lead=''):
    # no truck icon and no farm name: with the long key (Înregistrează vânzarea) the summary has ~130 px, the pair alone fits
    # one line; the sale colour stays on line 2 (the cycle mark's sale bar); the farm is in the form and the header
    return f'<span>{(lead + " · ") if lead else ""}{pair(*tot)}</span>'
SUM2 = f'<span>{cmark("end", "sale", "ciclul se închide")}</span>'
SELL = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" aria-haspopup="dialog">Înregistrează vânzarea</button>'
SELL_OFF = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="sb-sum">Înregistrează vânzarea</button>'

def screen(parts, save, overlay='', view='', modal=''):
    """One 390×844 viewport (D60). view: ' ip-screen__view--end' scrolls the form to its end. modal: a Dialog over the screen;
    the header, form and SaveBar behind it are inert (out of the tab order and the a11y tree)."""
    h, ine = HEAD, ''
    if modal:
        h = h.replace('<header class="ip-formhead">', '<header class="ip-formhead" inert>', 1)
        save = save.replace('<div class="ip-savebar', '<div inert class="ip-savebar', 1); ine = ' inert'
    return f'''<div class="ip-screen ip-screen--fixed{" ip-scrimhost" if modal else ""}">
{h}
<div class="ip-screen__view{view}"{ine}>
<main class="ip-screen__main s13-main">
{"".join(parts)}
</main>
{overlay}
</div>
{save}{modal}
</div>'''

def tall(parts, save):  # the default board: the whole form, one tall phone board (canvas.md), SaveBar sticky at the foot
    return f'''<div class="ip-screen">
{HEAD}
<main class="ip-screen__main s13-main">
{"".join(parts)}
</main>
{save}
</div>'''

FORM = lambda: [farm(), DATE, rows('all'), buyer(), note(), effect()]
boards = []
# 1 · telefon — the whole flock sold with Toate on all six rows, before Înregistrează vânzarea (whole form, tall board)
boards.append(('', 'telefon · toată ferma · exemplu de stare (D39)', tall(FORM(), B.savebar(sum1(), SUM2, SELL)), 'tall'))
# 2 · confirmare — the sale empties the farm, so Înregistrează vânzarea opens the confirmation Dialog (C4); focus on Anulează
DLG = f'''
<!-- Dialog (modal alertdialog): focus starts on Anulează (craft §5); Esc or Anulează closes it and focus returns to Înregistrează
     vânzarea. The form behind is inert. -->
<div class="ip-scrim ip-scrim--phone"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-note dlg-fx">
<h2 id="dlg-t">Înregistrează vânzarea și închide ciclul la Hîrbovăț</h2>
<p class="ip-dialog__note" id="dlg-note">Vinzi toate păsările din Hîrbovăț pe {num(dd(DT))}: {pair(*TOT)}. Ferma ajunge la 0 păsări și ciclul ei se închide.</p>
<div class="ip-effect s13-effect" id="dlg-fx">{rec("sale", 16)}<div class="s9-fxbody"><span class="s9-fxcyc">{cmark("end", "sale", CLOSE)}<span class="s9-fxsub">Seria 2026 · {num(dd(HSTART)[:5] if HSTART.year == DT.year else dd(HSTART))}–{num(dd(DT))}</span></span></div></div>
<div class="ip-dialog__keys ip-dialog__keys--stack"><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-focus" autofocus="autofocus">Anulează</button><button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Înregistrează vânzarea</button></div>
</div></div>'''
boards.append(('confirmare', 'confirmare · golește ferma · exemplu de stare (D39)', screen(FORM(), B.savebar(sum1(), SUM2, SELL), view=' ip-screen__view--end', modal=DLG), 'vp'))
# 3 · refuzat — H4.2 ♂ 500 typed over the 476 alive (D39 state example): row error, preview = fault Banner, key disabled;
#   the refused value is left out of the totals and the SaveBar (D77), line 2 names it
# S10P: lamp + section + fix took 3 lines beside the long key (~130 px); while refused the bar carries only the reason, in two
#   lines (the key is disabled, its aria-describedby reads it); the totals stay in the form's total line, minus the refused value
ref_save = B.savebar(f'{lamp("fault", "refuzat")}<span>{RS}</span>', f'<span>cel mult {roo(RM)}</span>', SELL_OFF)
boards.append(('refuzat', 'refuzat · peste cei în viață · exemplu de stare (D39)', screen([farm(), DATE, rows('refuz'), buyer(), note(), effect('refuz')], ref_save, view=' ip-screen__view--end'), 'vp'))
# 4 · salvat — read-only form scrolled to its end, Toast with Anulează (undo the same day), Vânzare nouă
toast = (f'<div class="ip-toastwrap"><div class="ip-toast" role="status">{icon("check", 16)}<span class="ip-toast__txt">Vândut · '
         f'♀ {num(n(TOT[0]))} · ♂ {num(n(TOT[1]))} · Hîrbovăț · ciclul închis</span><button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
saved = [farm('ro'), DATE_RO, rows('ro'), buyer(True), note(True), effect()]
boards.append(('salvat', 'salvat · exemplu de stare (D39)', screen(saved, B.savebar(sum1(lead='vândut'), f'<span>poți anula până la {num("23:59")}</span>',
              '<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Vânzare nouă</button>'), overlay=toast, view=' ip-screen__view--end'), 'vp'))
# 5 · eroare la salvare — online only (spec §1): fault Banner + Reîncearcă right above the EffectPreview, form scrolled to its
#   end (as S9P), typed values kept as a draft (D60), key stays enabled
ban = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{icon("alert", 16)}<div class="ip-banner__body"><p class="ip-m0">Vânzarea nu s-a salvat: '
       f'{pair(*TOT)}, Hîrbovăț, {num(dd(DT))}. Verifică conexiunea și apasă Reîncearcă.</p>'
       f'<p class="ip-m0">Ce ai scris rămâne păstrat în browser până la salvare.</p></div><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-banner__key">Reîncearcă</button></div>')
boards.append(('eroare-la-salvare', 'eroare la salvare · exemplu de stare (D39)', screen([farm(), DATE, rows('all'), buyer(), note(), ban, effect()], B.savebar(sum1(), SUM2, SELL), view=' ip-screen__view--end'), 'vp'))
# 6 · parțială — Toate on H4.1 and H4.2 only (D39 example from counts-latest): the farm keeps ♀ 17 028 · ♂ 1 729, the cycle
#   continues, so Înregistrează vânzarea saves directly (no Dialog); form scrolled to its end (the preview)
SUM2_PART = f'<span>{cmark("start", "", "ciclul continuă")}</span>'
SELL_PART = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Înregistrează vânzarea</button>'
boards.append(('partiala', 'parțială · ciclul continuă · exemplu de stare (D39)', screen([farm(), DATE, rows('part'), buyer(), note(), effect('part')],
              B.savebar(sum1(PSOLD), SUM2_PART, SELL_PART), view=' ip-screen__view--end'), 'vp'))

# the phone Dialog is additions.css section 35 (ip-scrim--phone, FIN04); the select chevron is additions.css (FIN10 fix 1)
CSS = P.CSS + SC.MAIN + SC.REC + SC.OPT + SC.EFFECT + SC.FX + SC.MODAL

# ---------- desktop: SidePanel beside S4 · Hîrbovăț (manager), 16.11.2026, nothing typed ----------
DDATE = f'''<div class="ip-field"><label class="ip-label" for="p-date">Data</label><div class="ip-row s4s-daterow"><span class="ip-seg" role="radiogroup" aria-label="Zi rapidă"><span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Azi</span><span class="ip-seg__item" role="radio" aria-checked="false" tabindex="-1">Ieri</span></span><span class="ip-input s4s-date"><input id="p-date" type="text" inputmode="numeric" value="{dd(DT)}" placeholder="zz.ll.aaaa" aria-describedby="p-date-h"/><button type="button" class="ip-iconbtn" aria-label="Alege data vânzării din calendar" aria-haspopup="dialog" aria-expanded="false">{icon("cal", 16)}</button></span></div><span class="ip-hint ip-num" id="p-date-h">SAPT {WK} · ciclu din {dd(HSTART)}</span></div>'''
PANEL = f'''<aside class="ip-sidepanel ip-sidepanel--dock" id="vp" aria-labelledby="vp-t">
<div class="ip-sidepanel__head"><div class="ip-row ip-sidepanel__title"><span class="ip-tile ip-rec ip-rec--sale" aria-hidden="true">{icon("sale")}</span><div><h2 class="ip-t-heading ip-m0" id="vp-t">Vânzare</h2><p class="ip-note ip-num ip-m0">SERIA 2026 · {dd(DT)}</p></div></div><button type="button" class="ip-iconbtn" aria-label="Închide Vânzare">{icon("x")}</button></div>
<div class="ip-sidepanel__body ip-sidepanel__body--form">
{farm(u='p', touch=False)}
{DDATE}
{rows('empty', u='p', touch=False)}
{buyer(u='p', touch=False)}
{note(u='p')}
{effect('empty', u='p')}
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
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{B.FONTS}<style>{CSS}</style></head><body>{root}</body></html>'''
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
