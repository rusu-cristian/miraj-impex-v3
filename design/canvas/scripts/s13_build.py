"""S13 · Livrare hrană (manager, spec C5): farm, date, feed type, kg, optional note, the stock-after line, Adaugă. No price.
Phone boards (390×844) reuse the S7 phone form shell by import (s7_build / s7s_build: FormHeader, DateField, icons, Lamp,
SaveBar shell + compact two-line summary, viewport/Toast/Banner CSS; the SidePanel/rec/effect CSS lives in sstate_css). The desktop board reuses the S11 manager page by
import (s11_build → s4s_build shell, sidebar, board writer) with the system SidePanel over it.
Content: the phone boards replay the 10.11.2026 Breeder 2 delivery (sample-hirbovat-2026.json, D27; the spec C5 example):
24 000 kg, stock 7 450 kg (end of 09.11) → 31 450 kg, entered by Ion Rusu at 11:52 (event log). The desktop board is a new
delivery on 16.11.2026 with the kg still empty, over the S11 page of that day (stock 28 252 kg, as its Readout).
python3 -I s13_build.py <repo_root> [heights.json, default s13_heights.json]
Writes design/canvas/project/S13-livrare-hrana[--<state>].dc.html, /tmp/s13/<stem>.html and /tmp/s4s/<stem>.html (renders)."""
import sys, re, json, pathlib, importlib, importlib.util, datetime as dt

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
ROOT, HP, CANVAS = CR.args('s13_heights.json')  # heights default: s13_heights.json
HERE = pathlib.Path(__file__).resolve().parent

# phone shell: s7s_build imports s7_build (both rewrite their S7 boards, unchanged)
sys.argv = ['s7s_build.py', str(ROOT)]
sys.path.insert(0, str(HERE))
P = importlib.import_module('s7s_build'); B = P.B
import sstate_css as SC  # SidePanel-beside-the-page, record icon, EffectPreview CSS (s13- names), shared with s9_build
# desktop shell: S11 page (manager) through s11_build → s4s_build
sys.argv = ['s11_build.py', str(ROOT)]
spec = importlib.util.spec_from_file_location('s11_build', HERE / 's11_build.py')
s11 = importlib.util.module_from_spec(spec); spec.loader.exec_module(s11)
s4 = s11.s4

n, dd, lamp, NB = B.n, B.dd, B.lamp, B.NB
S, START, FTN, FTS = B.S, B.START, B.FTN, B.FTS
B.IC.update({'delivery': s4.IC['delivery'], 'coins': s11.COINS, 'x': '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>'})
icon = B.icon

# ---------- data ----------
DEL = next(d for d in S['deliveries'] if d['id'] == 'del-hir-2026-11-10-b2')
EDAY = dt.date.fromisoformat(DEL['date']); FT = DEL['feed_type_id']; KG = DEL['qty_kg']
EWEEK = (EDAY - START).days // 7 + 1
fbt = {(f['date'], f['feed_type_id']): f for f in S['feed_by_type_daily']}
BEFORE = round(fbt[((EDAY - dt.timedelta(days=1)).isoformat(), FT)]['stock_end_kg']); AFTER = BEFORE + KG
win = [EDAY - dt.timedelta(days=k) for k in (3, 2, 1)]
AVG3 = sum(fbt[(d.isoformat(), FT)]['consumption_kg'] for d in win) / 3; DAYS = AFTER / AVG3
EV = next(e for e in S['event_log'] if e['record'] == 'feed_delivery' and e['at'].startswith(DEL['date']))
assert (EWEEK, KG, BEFORE, AFTER, FTN[FT], n(DAYS, 1)) == (40, 24000, 7450, 31450, 'Breeder 2', '7,9')
assert AFTER == S['summary_today']['spec_examples']['C5_toast_stock_after_10_11_delivery_kg'] and not DEL['priced']
assert (EV['user'], EV['at'][11:]) == ('Ion Rusu', '11:52')
NOW = s11.stocks[FT]['stock_kg']; assert NOW == 28252  # desktop: 16.11.2026, as the S11 Readout
HB = [t for t in B.HB_TYPES]  # Hîrbovăț active types: Breeder 2 (♀), Cocoși (♂)

def sexsign(t):
    sx = FTS[t]; c = 'hen' if sx == 'F' else 'rooster'
    return f'<span class="ip-sex ip-sex--{c}">{"♀" if sx == "F" else "♂"}</span> pentru {"găini" if sx == "F" else "cocoși"}'
def rec(k, size=14): return f'<span class="ip-rec ip-rec--feed s13-rec">{icon(k, size)}</span>'
FARMS = ['Hîrbovăț', 'Telița']
CYC = json.loads((ROOT / 'design/data/extracts/cycles.json').read_text(encoding='utf-8'))
assert [c['status'] for c in CYC['cycles'] if c['farm_id'] == 'telita'] == ['closed'] and \
    max(c['end_date'] for c in CYC['cycles'] if c['farm_id'] == 'telita') == '2026-02-28'  # no open cycle at Telița on 10.11 or 16.11.2026
TEL_WHY = 'fără ciclu deschis'
def kgs(v): return f'<span class="ip-num">{n(v)}{NB}kg</span>'

# ---------- phone form (S7 shell) ----------
HEAD = B.HEAD.replace('href="#acasa" aria-label="Înapoi la Acasă"', 'href="#inregistreaza" aria-label="Înapoi la Înregistrează"') \
    .replace(icon('feed'), icon('delivery')).replace('Hrană zilnică', 'Livrare hrană').replace(dd(B.TODAY), dd(EDAY))
assert HEAD.count('Livrare hrană') == 1 and dd(EDAY) in HEAD and icon('delivery') in HEAD

def farm(ro=False, touch=True):
    t = ' ip-input--touch' if touch else ''
    if ro: return f'<div class="ip-field"><label class="ip-label" for="f-farm">Fermă</label><span class="ip-input{t} is-readonly"><input id="f-farm" type="text" value="Hîrbovăț" readonly/></span></div>'
    # FIN12 (D130): Telița has no open cycle (closed 28.02.2026), so it is listed disabled with its reason; options with
    # state make Fermă the section 53 Select: phone opens the BottomSheet (aria-haspopup dialog, no list in the page),
    # desktop its listbox (hidden while closed)
    lbl = '<span class="ip-label" id="f-farm-l">Fermă</span>'
    if touch:
        return (f'<div class="ip-field">{lbl}<div class="ip-select"><div class="ip-input ip-select__btn{t}" id="f-farm" role="combobox" tabindex="0" aria-haspopup="dialog" aria-expanded="false" aria-labelledby="f-farm-l">'
                f'<span class="ip-select__val">Hîrbovăț</span>{icon("chev", 16)}</div></div></div>')
    O = '<span class="ip-select__txt"><span class="ip-select__top">'
    opts = (f'<li class="ip-menu__item ip-select__opt" role="option" id="f-farm-o1" aria-selected="true"><span class="ip-select__chk" aria-hidden="true">{icon("check", 16)}</span>{O}{FARMS[0]}</span></span></li>'
            f'<li class="ip-menu__item ip-select__opt" role="option" id="f-farm-o2" aria-selected="false" aria-disabled="true"><span class="ip-select__chk" aria-hidden="true"></span>{O}{FARMS[1]} <span class="ip-note">{TEL_WHY}</span></span></span></li>')
    return (f'<div class="ip-field">{lbl}<div class="ip-select"><div class="ip-input ip-select__btn{t}" id="f-farm" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="false" aria-controls="f-farm-list" aria-labelledby="f-farm-l">'
            f'<span class="ip-select__val">Hîrbovăț</span>{icon("chev", 16)}</div>'
            f'<ul class="ip-menu ip-select__menu" role="listbox" id="f-farm-list" aria-labelledby="f-farm-l" tabindex="-1" hidden="hidden">{opts}</ul></div></div>')

DATE = B.DATE.replace(dd(B.TODAY), dd(EDAY))
assert DATE.count(dd(EDAY)) == 1 and f'SAPT {EWEEK}' in DATE
DATE_RO = f'''<div class="ip-field"><label class="ip-label" for="d-val">Data</label><span class="ip-input ip-input--touch ip-num is-readonly"><input id="d-val" type="text" value="{dd(EDAY)}" readonly aria-describedby="d-wk"/></span><span class="ip-note ip-num" id="d-wk">SAPT {EWEEK} · ciclu din {dd(START)}</span></div>'''

def ftype(sel=FT, ro=False, touch=True):
    t = ' ip-input--touch' if touch else ''
    hint = f'<span class="ip-hint" id="f-type-h">{sexsign(sel)}</span>'
    if ro: return f'<div class="ip-field"><label class="ip-label" for="f-type">Tip hrană</label><span class="ip-input{t} is-readonly"><input id="f-type" type="text" value="{FTN[sel]}" readonly aria-describedby="f-type-h"/></span>{hint}</div>'
    opts = ''.join(f'<option{" selected" if x == sel else ""}>{FTN[x]}</option>' for x in HB)
    return f'<div class="ip-field"><label class="ip-label" for="f-type">Tip hrană</label><span class="ip-input{t}"><select id="f-type" aria-describedby="f-type-h">{opts}</select>{icon("chev", 16)}</span>{hint}</div>'

ERR_KG = f'Scrie câte kg de {FTN[FT]} s-au livrat la Hîrbovăț.'
def qty(v=KG, state='', touch=True):
    t = ' ip-input--touch' if touch else ''
    val = f'{n(v)}' if v is not None else ''
    if state == 'ro':
        return f'<div class="ip-field"><label class="ip-label" for="f-kg">Cantitate</label><span class="ip-input{t} ip-num is-readonly"><input id="f-kg" type="text" value="{val}" readonly/><span class="suffix">kg</span></span></div>'
    if state == 'error':
        return (f'<div class="ip-field"><label class="ip-label" for="f-kg">Cantitate</label><span class="ip-input{t} ip-num is-error"><input id="f-kg" type="text" inputmode="decimal" value="" required aria-required="true" aria-invalid="true" aria-describedby="f-kg-e f-stock"/><span class="suffix">kg</span></span>'
                f'<p class="ip-err s13-err" id="f-kg-e">{icon("alert", 14)}<span>{ERR_KG}</span></p></div>')
    foc = ' is-focus' if state == 'focus' else ''
    return f'<div class="ip-field"><label class="ip-label" for="f-kg">Cantitate</label><span class="ip-input{t} ip-num{foc}"><input id="f-kg" type="text" inputmode="decimal" value="{val}" required aria-required="true" aria-describedby="f-stock"/><span class="suffix">kg</span></span></div>'

def effect(before, after, avg=None, days=None, when='acum'):
    if after is None:
        aft = '<span>după livrare: <span class="ip-num">–</span>, scrie cantitatea</span>'
    else:
        aft = f'<span>după livrare: <b class="s13-strong">{kgs(after)}</b></span>'
    sub = f'<span class="s13-effsub">ajunge <span class="ip-num">{n(days, 1)}</span> zile la media <span class="ip-num ip-nowrap">{win[0].strftime("%d.%m")}–{win[-1].strftime("%d.%m")}</span> (<span class="ip-num ip-nowrap">{n(avg)}{NB}kg/zi</span>)</span>' if days else ''
    return f'''<div class="ip-effect s13-effect" id="f-stock">{rec("delivery", 16)}<div class="s13-effbody"><span>Stoc {FTN[FT]} la Hîrbovăț {when}: {kgs(before)}</span>{aft}{sub}</div></div>'''

def note(ro=False):
    if ro: return '<div class="ip-field"><span class="ip-label">Notă</span><span class="ip-note">– fără notă</span></div>'
    return '<div class="ip-field"><label class="ip-label" for="f-note">Notă <span class="s13-opt">(opțional)</span></label><span class="ip-input ip-input--area"><textarea id="f-note" rows="2" placeholder="de ex. numărul avizului"></textarea></span></div>'
NOPRICE = f'<p class="ip-note s13-noprice">{rec("coins")}<span>Fără preț: îl adaugă proprietarul, din factură.</span></p>'

PKG14 = rec('delivery')
def sum1(kg=KG, lead=''):
    k = f'<span class="ip-num">{n(kg)}</span>' if kg is not None else '<span class="ip-num">–</span>'
    if lead: return f'{PKG14}<span>{lead}{k} kg {FTN[FT]}</span>'  # saved: the farm is in the header; one line at 390
    return f'{PKG14}<span>{k} kg {FTN[FT]} ·</span><span>Hîrbovăț</span>'
SUM2 = f'<span><span class="ip-num">{dd(EDAY)}</span> ·</span><span>fără preț</span>'
ADD = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Adaugă</button>'
ADD_OFF = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="sb-sum">Adaugă</button>'

def screen(parts, save, overlay='', view=''):  # s7s_build.screen with this form's header; view ' ip-screen__view--end': scrolled to its end
    return f'''<div class="ip-screen ip-screen--fixed">
{HEAD}
<div class="ip-screen__view{view}">
<main class="ip-screen__main">
{"".join(parts)}
</main>
{overlay}
</div>
{save}
</div>'''

FORM = lambda kgstate='', kgv=KG: [farm(), DATE, ftype(), qty(kgv, kgstate), effect(BEFORE, AFTER if kgv else None, AVG3 if kgv else None, DAYS if kgv else None), note(), NOPRICE]
boards = []
# 1 · telefon — the 10.11.2026 delivery typed, before Adaugă
boards.append(('', 'telefon', screen(FORM(), B.savebar(sum1(), SUM2, ADD))))
# 2 · salvat — read-only form, C5 Toast (with undo), Livrare nouă
toast = (f'<div class="ip-toastwrap"><div class="ip-toast" role="status">{icon("check", 16)}<span class="ip-toast__txt">Livrare adăugată · '
         f'<span class="ip-num">{n(KG)}</span> kg {FTN[FT]} · stoc Hîrbovăț: <span class="ip-num">{n(AFTER)}</span> kg · preț de adăugat de proprietar</span>'
         f'<button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
saved = [farm(ro=True), DATE_RO, ftype(ro=True), qty(state='ro'), effect(BEFORE, AFTER, AVG3, DAYS, when='înainte'), note(ro=True), NOPRICE]
boards.append(('salvat', 'salvat', screen(saved, B.savebar(sum1(lead='adăugat · '), f'<span>{EV["user"]} ·</span><span><span class="ip-num">{EV["at"][11:]}</span> · fără preț</span>', '<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Livrare nouă</button>'), overlay=toast)))
# 3 · cantitate lipsă — kg left empty (validated on blur): field error naming the fix, Adaugă disabled
boards.append(('cantitate-lipsa', 'cantitate lipsă', screen(FORM('error', None), B.savebar(sum1(None), f'{lamp("fault", "refuzat")}<span>scrie cantitatea în kg</span>', ADD_OFF))))
# 4 · eroare la salvare — online only (spec §1): fault Banner + Reîncearcă (focus, S16P), typed values kept as a draft; the Banner
#   ends the form, scrolled to its end, right above the SaveBar (D141, as S9/S10)
ban = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{icon("alert", 16)}<div class="ip-banner__body"><p class="ip-m0">Livrarea nu s-a salvat: '
       f'<span class="ip-num">{n(KG)}</span> kg {FTN[FT]}, Hîrbovăț, <span class="ip-num">{dd(EDAY)}</span>. Verifică conexiunea la internet, apoi apasă Reîncearcă.</p>'
       f'<p class="ip-m0">Ce ai scris rămâne păstrat în browser până la salvare.</p></div><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-banner__key ip-focus" autofocus="autofocus">Reîncearcă</button></div>')
boards.append(('eroare-la-salvare', 'eroare la salvare', screen(FORM() + [ban], B.savebar(sum1(), SUM2, ADD), view=' ip-screen__view--end')))

CSS = P.CSS + SC.MAIN + SC.REC + SC.OPT + SC.M0 + '''
.s13-err{margin:0}
.s13-err > .ip-icon{flex:none;margin-top:var(--ip-hair)}''' + SC.EFFECT + '''
.s13-effbody{display:flex;flex-direction:column}
.s13-effbody > span{white-space:normal}
.s13-strong{font-weight:500;color:var(--ink)}
.s13-effsub{color:var(--ink-secondary);font-size:var(--ip-fs-small);line-height:var(--ip-lh-small);margin-top:var(--space-1)}
.s13-noprice{display:flex;gap:var(--space-2);align-items:flex-start;margin:0}
.s13-noprice > .s13-rec{margin-top:var(--ip-hair)}
.ip-input--area textarea::placeholder{color:var(--ink-muted)}'''

# ---------- desktop: SidePanel over S11 · Rapoarte · Hrană (manager), 16.11.2026 ----------
DT = s4.TODAY; DTD = dt.date.fromisoformat(DT); DWEEK = (DTD - START).days // 7 + 1; assert DWEEK == 40
DDATE = f'''<div class="ip-field"><label class="ip-label" for="p-date">Data</label><div class="ip-row s4s-daterow"><span class="ip-seg" role="radiogroup" aria-label="Zi rapidă"><span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Azi</span><span class="ip-seg__item" role="radio" aria-checked="false" tabindex="-1">Ieri</span></span><span class="ip-input s4s-date"><input id="p-date" type="text" inputmode="numeric" value="{dd(DTD)}" placeholder="zz.ll.aaaa" aria-describedby="p-date-h"/><button type="button" class="ip-iconbtn" aria-label="Alege data livrării din calendar" aria-haspopup="dialog" aria-expanded="false">{icon("cal", 16)}</button></span></div><span class="ip-hint ip-num" id="p-date-h">SAPT {DWEEK} · ciclu din {dd(START)}</span></div>'''
PANEL = f'''<aside class="ip-sidepanel ip-sidepanel--dock" id="lp" aria-labelledby="lp-t">
<div class="ip-sidepanel__head"><div class="ip-row ip-sidepanel__title"><span class="ip-tile ip-rec ip-rec--feed" aria-hidden="true">{icon("delivery")}</span><div><h2 class="ip-t-heading ip-m0" id="lp-t">Livrare hrană</h2><p class="ip-note ip-num ip-m0">HÎRBOVĂȚ · {dd(DTD)} · SAPT {DWEEK}</p></div></div><button type="button" class="ip-iconbtn" aria-label="Închide Livrare hrană">{icon("x")}</button></div>
<div class="ip-sidepanel__body ip-sidepanel__body--form">
{farm(touch=False)}
{DDATE}
{ftype(touch=False)}
{qty(None, 'focus', touch=False)}
{effect(NOW, None)}
{note()}
{NOPRICE}
</div>
<div class="ip-sidepanel__foot"><span class="ip-note" id="lp-why">Scrie cantitatea în kg</span><div class="ip-row ip-sidepanel__keys"><button type="button" class="ip-btn ip-btn--secondary">Anulează</button><button type="button" class="ip-btn ip-btn--primary" disabled aria-describedby="lp-why">Adaugă</button></div></div>
</aside>'''
PANEL = re.sub(r'((?:id|for|aria-describedby|aria-labelledby|aria-controls)="[^"]*")', lambda m: re.sub(r'\bf-', 'p-', m.group(1)), PANEL)  # S11 page has its own f-farm
DH = 1024
# the page reflows beside the panel (margin = panel-w, nothing under it, D73); its "Livrare hrană" key shows pressed and
# controls the open panel, as secondary, so the panel's Adaugă stays the one primary key in view
KEY_RE = r'<a class="ip-btn ip-btn--primary" href="#livrare-hrana">(.*?)Livrare hrană</a>'
SPG = s11.page("Manager"); assert len(re.findall(KEY_RE, SPG)) == 1
SPG = re.sub(KEY_RE, r'<button type="button" class="ip-btn ip-btn--secondary is-pressed" aria-expanded="true" aria-controls="lp">\1Livrare hrană</button>', SPG)
# beside the open panel the page is 440 px narrower: at 1280 the S11 tables (711 px) no longer fit their 560 px block; they
# scroll inside their box as s11_build's named, focusable regions (S11X; FIN14: the system .ip-tablescroll, was .s11-tscroll)
assert SPG.count('class="ip-tablescroll ip-tablescroll--framed" role="region" tabindex="0"') == 2
DESK = f'<div class="ip-dockpage">{SPG}</div>'
assert DH == SC.SIDE_H
CSS_D = s11.CSS + SC.SIDE + '''
''' + CSS[len(P.CSS + SC.MAIN):]  # everything after the phone screen CSS and its main gap (MAIN '' since FX03)

# ---------- write ----------
STEM0, TITLE0 = 'S13-livrare-hrana', 'S13 · Livrare hrană'
VH = P.VH
def write_phone(slug, state, body):
    stem = STEM0 + (f'--{slug}' if slug else ''); title = f'{TITLE0} · {state}'
    root = f'<div class="ip-root" style="padding:0; width:390px; height:{VH}px; overflow:hidden">{body}</div>'
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
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":390,"height":{VH}}}}}'>
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
    pathlib.Path('/tmp/s13').mkdir(exist_ok=True)
    pathlib.Path(f'/tmp/s13/{stem}.html').write_text(wrap, encoding='utf-8')
    print('ok', stem, title, VH)
    return stem, title, VH

out = [write_phone(*b) for b in boards]
s4.write(STEM0 + '--desktop', TITLE0 + ' · desktop', DESK, PANEL, H=DH, css=CSS_D)
out.append((STEM0 + '--desktop', TITLE0 + ' · desktop', DH))
if HP:
    HP.write_text(json.dumps({s: h for s, _, h in out}, indent=1))
