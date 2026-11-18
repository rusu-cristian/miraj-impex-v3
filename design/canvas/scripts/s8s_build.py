"""S8S · Cântărire state boards (spec C7 + §1 shared rules), appended to the S8 row after its 7 boards (step 470).
python3 -I s8s_build.py <repo_root> [--canvas]
Imports s8_build (which rewrites the 7 S8 boards byte-identical, and S7 through s7s_build; never canvas.json) and reuses its
FormHeader, DateField, SectionSelect combobox, Sex, Masa medie, Readout (one div.ip-readout__live, additions §38: lamp word + comparison, value and
caption aria-hidden), values grid, SaveBar summary and screen CSS. Viewport/Banner/skeleton CSS from sstate_css; the phone
SectionSelect BottomSheet is additions.css §33 (S9P), its scrim anchored as S9 (SC.MODAL). System classes only.
Content (real only, no D39 example needed):
- se încarcă: Hîrbovăț 16.11.2026 (farm, day and week come from the route); Secțiune, Sex, Masa medie, Readout and the values
  load: static sunken skeletons (fields = §54 field skeletons, the Readout slot = the §54 skeleton of the comparison Readout), SaveBar skeleton + Salvează disabled.
- alege secțiunea: the gol form with the Secțiune BottomSheet open (390x844, sheet fixed to the screen): Hîrbovăț sections with
  ♀ · ♂ alive on 16.11.2026 (counts-latest, sample D27), H1.1 chosen and active; Telița sections "goală" and disabled (cycle
  closed 28.02.2026, real) — an empty section can't be weighed.
- zi trecută (medic): replay "today" = 24.02.2026 (D95); the vet opens last week's real weighing, Telița H1 ♀ 17.02.2026
  2 429 g · CV 8,8% · uniformitate 75% (weighings-weekly), read-only, info Banner with the C1 line, Salvează disabled. The
  Readout still shows the comparison (ghid SAPT 20 2 370 g); a saved day is not re-checked (no warn line, D77).
- fermă fără păsări: Telița 16.11.2026, 0 birds, cycle closed 28.02.2026 (counts-latest, real): EmptyState as S6S.
- deja introdus: Hîrbovăț 16.11.2026 H1.1 ♀ already weighed today (3 967 g, Dr. Ceban 10:10 per s8_build): values read-only,
  Secțiune and Sex stay live (pick another to weigh), info Banner, one form-level Editează in the SaveBar (D77/D60).
- gol · Telița 24.02.2026: the empty form on a growth farm; the comparison waits: lamp "FĂRĂ DATE", "ghid SAPT 21: 2 525 g ·
  scrie masa medie" (s8_build readout() path with no value), Salvează disabled.
Writes design/canvas/project/S8-cantarire--<state>.dc.html and /tmp/s8/<stem>.html (wrappers for s8_render.py).
--canvas: adds/updates only these S8 entries in canvas.json (boards + order), y = the S8 row, x = 3290 + 470·i (canvas_rows.py)."""
import sys, json, pathlib, importlib, io, contextlib, datetime as dt

ROOT = pathlib.Path(sys.argv[1]).resolve()
CANVAS = '--canvas' in sys.argv[2:]
HERE = pathlib.Path(__file__).resolve().parent
HP = HERE / 's8_heights.json'
sys.argv = ['s8_build.py', str(ROOT), str(HP)]
sys.path.insert(0, str(HERE))
import skel as SKL
with contextlib.redirect_stdout(io.StringIO()):
    E = importlib.import_module('s8_build')  # rewrites S7 + the 7 S8 boards, byte-identical
B, P, SC = E.B, E.P, E.SC
n, dd, icon, lamp, NB = E.n, E.dd, E.icon, E.lamp, E.NB
CLS, SIGN = E.CLS, E.SIGN
HS = json.loads(HP.read_text()) if HP.exists() else {}
for k, v in {  # Lucide paths, only where s7/s8 have none (info = the Banner preview icon, as S6S)
    'info': '<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>',
    'x': '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',
    'chevup': '<path d="m18 15-6-6-6 6"/>'}.items():
    B.IC.setdefault(k, v)

EX = ROOT / 'design/data/extracts'
CL = json.load(open(EX / 'counts-latest.json', encoding='utf-8'))
FS = json.load(open(EX / 'farms-sections.json', encoding='utf-8'))
FARMS = {f['farm_id']: f for f in CL['farms']}
HIR, TEL = FARMS['hirbovat'], FARMS['telita']
assert CL['meta']['as_of'] == E.HB['day'].isoformat() and TEL['cycle_status'] == 'closed' and TEL['totals']['all'] == 0
TEND = dt.date.fromisoformat(TEL['cycle_end'])
HB, OK = E.HB, E.OK
PAST = E.ctx_telita('H1', 'F', '2026-02-17')  # last week's real weighing, seen from the 24.02.2026 replay (D95)
assert (PAST['wk'], PAST['guide'], PAST['w'], PAST['cv'], PAST['uni']) == (20, 2370, 2429, 8.8, 75)
assert (PAST['day'] + dt.timedelta(days=7)) == OK['day']
SAVED_AT, SAVED_BY = '10:10', 'Dr. Ceban'  # the 16.11.2026 H1.1 ♀ weighing in the event log (s8_build docstring, S6 boards)

SAVE_OFF = E.SAVE_OFF
def save_off(desc): return f'<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="{desc}">Salvează</button>'
EDIT_KEY = '<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Editează</button>'  # D77 / D60, as S6S
OLD_CTX = f'{HB["farm"].upper()} · {dd(HB["day"])} · SAPT {HB["wk"]}'

def screen(parts, save, c=HB, hd=None, sheet=''):
    """s8_build.screen with a header override and an optional modal sheet: the sheet board is one 390x844 phone screen (the
    sheet is fixed to it); the header, form and SaveBar behind it are inert (S9P)."""
    h = hd or E.head(c)
    if sheet:
        h = h.replace('<header class="ip-formhead">', '<header class="ip-formhead" inert>', 1)
        save = save.replace('<div class="ip-savebar', '<div inert class="ip-savebar', 1)
    return f'''<div class="s7-screen s8-screen{" s7s-vp s9-modal" if sheet else ""}">
{h}
<div class="s7s-view"{" inert" if sheet else ""}>
<main class="s7-main">
{"".join(parts)}
</main>
</div>
{save}{sheet}
</div>'''

def banner(bid, *lines):
    body = ''.join(f'<p class="s7-m0">{l}</p>' for l in lines)
    return f'<div class="ip-banner ip-banner--info s7s-banner" id="{bid}">{icon("info", 16)}<div class="s7s-banner__body">{body}</div></div>'
def sexs(sx): return f'<span class="ip-sex ip-sex--{CLS[sx]}">{SIGN[sx]}</span>'

boards = []  # (slug, state, body)

# 1 · se încarcă — header and date from the route; the rest static sunken skeletons (no shimmer), SaveBar summary skeleton
def skel(w, h): return f'<div class="ip-skel" style="width:{w};height:{h}"></div>'
SK16 = 'var(--ip-lh-small)'
def skfield(w, box): return SKL.field(box)  # §54 field skeleton (FIN11G; was s7s-skcol): label line + the loaded box
TOUCH, COUNT = 'touch', 'count'
RO_SKEL = SKL.readout(None, kv=2, live=True)  # §54 Readout skeleton: the comparison Readout's own lines (FIN11G)
sk = (f'<p class="ip-sr" role="status">Se încarcă secțiunile și ghidul pentru <span class="ip-num">{dd(HB["day"])}</span>…</p>'
      f'<div class="s7s-skblock" aria-busy="true">{skfield("30%", TOUCH)}{skfield("15%", TOUCH)}{skfield("40%", COUNT)}{RO_SKEL}'
      f'<div class="s8-pair">{skfield("50%", TOUCH)}{skfield("70%", TOUCH)}{skfield("80%", TOUCH)}</div></div>')
sv1 = (f'<div class="ip-savebar s7-savebar"><span class="ip-savebar__sum ip-ftot-sum" aria-hidden="true">{SKL.line("ip-skel--w1")}{SKL.line("ip-skel--w2")}</span>'
       f'<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled>Salvează</button></div>')
boards.append(('se-incarca', 'se încarcă', screen([E.date(HB), sk], sv1)))

# 2 · alege secțiunea — the gol form, Secțiune combobox expanded into the §33 BottomSheet (focus on its listbox, H1.1 chosen)
def oid(farm, code): return f'so-{"t" if farm == "Telița" else ""}{code.replace(".", "-")}'
def opt(farm, code, body, dis=False, sel=False):
    a = ' aria-disabled="true"' if dis else ''
    return (f'<div class="ip-menu__item{" is-active" if sel else ""}" role="option" id="{oid(farm, code)}" aria-selected="{"true" if sel else "false"}"{a}>'
            f'<span class="ip-sheet__code"><span class="ip-sr">{farm} </span>{code}</span>{body}</div>')
HSEC = [(s['section'], s['F'], s['M']) for s in HIR['sections']]
TSEC = [s['code'] for s in FS['sections'] if s['farm_id'] == 'telita']
assert [s['section'] for s in TEL['sections']] == TSEC and all(s['F'] == s['M'] == 0 for s in TEL['sections']) and HSEC[0][0] == HB['sec']
ACT = oid('Hîrbovăț', HB['sec'])
hg = ''.join(opt('Hîrbovăț', s, B.pair(n(f), n(m)), sel=(s == HB['sec'])) for s, f, m in HSEC)
tg = ''.join(opt('Telița', s, '<span class="ip-note">goală</span>', dis=True) for s in TSEC)
SHEET = f'''
<!-- BottomSheet (modal, §33): focus moves to the listbox on open, the chosen option active (aria-activedescendant; arrow keys
     move it, Enter or a tap picks it and closes the sheet), Tab cycles Închide <-> listbox (focus trap), Esc or Închide closes
     it and focus returns to the Secțiune combobox. Header, form and SaveBar behind are inert. -->
<div class="ip-scrim ip-scrim--sheet">
<div class="ip-sheet ip-sheet--select" role="dialog" aria-modal="true" id="sh-sec" aria-labelledby="sh-sec-t" aria-describedby="sh-sec-d">
<div class="ip-sheet__handle" aria-hidden="true"></div>
<div class="ip-sheet__head"><div><h2 class="ip-t-heading ip-sheet__title" id="sh-sec-t">Alege secțiunea</h2><p class="ip-note ip-sheet__sub" id="sh-sec-d">În viață pe <span class="ip-num">{dd(HB["day"])}</span>. O secțiune goală nu se cântărește.</p></div><button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Închide lista">{icon("x")}</button></div>
<div class="ip-sheet__list ip-focus" role="listbox" id="sh-sec-lb" aria-labelledby="sh-sec-t" tabindex="0" aria-activedescendant="{ACT}">
<div role="group" aria-labelledby="sg-h"><div class="ip-menu__group" id="sg-h">Hîrbovăț · ouat</div>{hg}</div>
<div role="group" aria-labelledby="sg-t"><div class="ip-menu__group" id="sg-t">Telița · creștere</div>{tg}</div>
</div>
</div>
</div>'''
sec_closed = E.section(HB)
sec_open = sec_closed.replace('aria-expanded="false"', 'aria-expanded="true" aria-controls="sh-sec"').replace(icon('chev', 16), icon('chevup', 16))
assert sec_open != sec_closed and 'aria-controls="sh-sec"' in sec_open
gol = E.form(HB, None, None, None, None); gol[1] = sec_open
boards.append(('alege-sectiunea', 'alege secțiunea', screen(gol, B.savebar(E.rec14() + E.what(HB, None), '<span>de completat · Masa medie</span>', SAVE_OFF), sheet=SHEET)))

# 3 · zi trecută (medic) — Telița H1 ♀ 17.02.2026 opened from the 24.02.2026 replay: read-only, Banner, Salvează disabled
C3 = PAST
AZI_ON = '<span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Azi</span>'
d3 = E.date(C3); assert AZI_ON in d3
d3 = d3.replace(AZI_ON, '<span class="ip-seg__item" role="radio" aria-checked="false" tabindex="0">Azi</span>')  # a calendar day: neither Azi nor Ieri
ban3 = banner('past-why', 'Pentru zilele trecute cere managerului.',
              f'Cântărirea {C3["sec"]} {sexs(C3["sx"])} din <span class="ip-num">{dd(C3["day"])}</span> e salvată; o vezi doar în citire.')
p3 = [d3, ban3, E.section(C3), E.sex(C3), E.weight(C3, C3['w'], ro=True), E.readout(C3, C3['w']), E.values(C3, E.dec(C3['cv']), E.dec(C3['uni']), None, ro=True, size_why='Nescrisă la cântărire.')]  # S8P: – with its reason (true for imported and in-app records; fix round 1)
sv3 = B.savebar(E.rec14() + f'<span><span class="ip-num">{dd(C3["day"])}</span> ·</span>' + E.what(C3, C3['w']), '<span>doar citire</span>', save_off('past-why'))
boards.append(('zi-trecuta', f'zi trecută (medic) · Telița {dd(C3["day"])}', screen(p3, sv3, c=C3)))

# 4 · fermă fără păsări — Telița 16.11.2026: EmptyState (S6S copy), the next step, SaveBar off
hd4 = E.head(HB).replace(OLD_CTX, f'TELIȚA · {dd(HB["day"])}'); assert 'TELIȚA · ' in hd4
WL = f'SAPT {HB["wk"]} · ciclu din {dd(HB["start"])}'
d4 = E.date(HB); assert WL in d4; d4 = d4.replace(WL, f'fără ciclu deschis · ultimul închis pe {dd(TEND)}')
empty = (f'<div class="ip-empty"><span class="ip-lamp ip-lamp--off">ciclu închis</span><div>Ferma e goală din <span class="ip-num">{dd(TEND)}</span>. '
         f'Ciclul pornește la prima populare sau transfer.</div><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Alege altă fermă</button></div>')
sv4 = (f'<div class="ip-savebar s7-savebar"><span class="ip-savebar__sum ip-ftot-sum" id="sb-sum"><span class="ip-ftot-sum__line">{E.rec14()}'
       f'<span><span class="ip-num">0</span> secțiuni cu păsări</span></span></span>{SAVE_OFF}</div>')
boards.append(('ferma-fara-pasari', 'fermă fără păsări', screen([d4, empty], sv4, hd=hd4)))

# 5 · deja introdus — H1.1 ♀ chosen again on 16.11.2026: saved values read-only, Secțiune/Sex live, Banner, Editează
C5 = HB
ban5 = banner('dup-why', f'{C5["sec"]} {sexs(C5["sx"])} e cântărit azi la <span class="ip-num">{SAVED_AT}</span>. '
              'Apasă Editează ca să schimbi valorile sau alege altă secțiune.')
p5 = [E.date(C5), E.section(C5), E.sex(C5), ban5, E.weight(C5, C5['w'], ro=True), E.readout(C5, C5['w']),
      E.values(C5, E.dec(C5['cv']), E.dec(C5['uni']), C5['size'], ro=True)]
sv5 = B.savebar(E.rec14() + '<span>salvat ·</span>' + E.what(C5, C5['w']), f'<span>{SAVED_BY} · <span class="ip-num">{SAVED_AT}</span></span>', EDIT_KEY)
boards.append(('deja-introdus', 'deja introdus', screen(p5, sv5)))

# 6 · gol · Telița 24.02.2026 — growth farm, nothing typed: lamp FĂRĂ DATE, "ghid SAPT 21: 2 525 g · scrie masa medie"
p6 = E.form(OK, None, None, None, None)
assert 'scrie masa medie' in p6[4] and 'FĂRĂ DATE<' in p6[4]
boards.append(('gol-telita', f'gol · Telița {dd(OK["day"])}', screen(p6, B.savebar(E.rec14() + E.what(OK, None), '<span>de completat · Masa medie</span>', SAVE_OFF), c=OK)))

CSS = E.CSS + SC.SKEL + SC.MODAL
STEM0, TITLE0 = 'S8-cantarire', 'S8 · Cântărire'
SYS = ROOT / 'design/system'
pathlib.Path('/tmp/s8').mkdir(exist_ok=True)
for slug, state, body in boards:
    stem, title = f'{STEM0}--{slug}', f'{TITLE0} · {state}'
    H = 844 if slug == 'alege-sectiunea' else HS.get(stem, 844)
    root = f'<div class="ip-root" style="padding:0; width:390px; min-height:{H}px">{body}</div>'
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
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":390,"height":{H}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
    (ROOT / f'design/canvas/project/{stem}.dc.html').write_text(dc, encoding='utf-8')
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{B.FONTS}<style>{CSS}</style></head><body>{root}</body></html>'''
    pathlib.Path(f'/tmp/s8/{stem}.html').write_text(wrap, encoding='utf-8')
    print('ok', stem, H, title)

if CANVAS:  # S8 row: these boards after the other S8 boards, step 470
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    mine = {f'{STEM0}--{s}.dc.html' for s, _, _ in boards}
    X0 = max(b['x'] for k, b in c['boards'].items() if CR.in_row(k, STEM0) and k not in mine) + 470
    x = CR.place(c, STEM0, [(f'{STEM0}--{slug}', 390, 844 if slug == 'alege-sectiunea' else HS.get(f'{STEM0}--{slug}', 844), f'{TITLE0} · {state}') for slug, state, _ in boards], X0)
    CR.save(ROOT, c)
    print('canvas: S8 row y', CR.row_y(c, STEM0), 'ends at x', x - 80)
