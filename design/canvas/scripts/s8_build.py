"""S8 · Cântărire (vet, spec C7): section, sex, date, Masa medie (g), CV %, Uniformitate %, Mărime probă, the live guide
comparison, Salvează. Phone boards only (screens.md S8 lists no desktop layout; D74 phones first).
Reuses the S7 phone form shell by import (s7s_build → s7_build: FormHeader, DateField, icons, Lamp, SexCount, SaveBar shell +
compact two-line summary; viewport/Toast/Banner CSS from sstate_css, "(opțional)" from SC.OPT). System classes only: FormHeader,
DateField, SectionSelect phone combobox (closed), SegmentedControl (sex), CountInput (Masa medie, ♀/♂ bar), TextField,
Readout (+ ip-readout__kv, off lamp D59), Lamp, Banner, Toast, SaveBar. Screen CSS (s8-*) is layout with tokens.
Content (real only):
- Hîrbovăț boards (D23, today 16.11.2026, SAPT 40): the latest weighing in the extracts, H1.1 ♀ 3 967 g · CV 7,3% · uniformitate
  81,9% · 78 păsări (sample-hirbovat-2026.json, D27; Dr. Ceban 10:10 in the event log); previous week 09.11.2026 4 006 g.
  The laying guide has no body weight (D29): Readout off lamp "FĂRĂ DATE ÎN GHID" (D59).
- Guide comparison boards replay the last real growth weighings at Telița (weighings-weekly.json + guide-growth.json), "today" =
  the weighing's date (as D73/D83): H1 ♀ 24.02.2026 2 553 g vs ghid SAPT 21 2 525 g (+1,1%, în normă) and H5-2 ♂ 17.02.2026
  3 372 g vs ghid SAPT 20 3 040 g (+10,9%, abatere; ±5% D19). Guide week by the D02 formula from placement 07.10.2025. The
  workbook has no sample size, so Mărime probă stays empty there (optional field).
- D39 state example: uniformitate typed 819 (for 81,9) on the Hîrbovăț weighing → refused, Salvează disabled.
python3 -I s8_build.py <repo_root> [heights.json]
Writes design/canvas/project/S8-cantarire[--<state>].dc.html and /tmp/s8/<stem>.html (plain wrappers for s8_render.py)."""
import sys, json, pathlib, importlib, datetime as dt

ROOT = pathlib.Path(sys.argv[1])
HP = pathlib.Path(sys.argv[2]) if len(sys.argv) > 2 else None
HERE = pathlib.Path(__file__).resolve().parent
HS = json.loads(HP.read_text()) if HP and HP.exists() else {}

# phone shell: s7s_build imports s7_build (both rewrite their S7 boards, unchanged)
sys.argv = ['s7s_build.py', str(ROOT)]
sys.path.insert(0, str(HERE))
P = importlib.import_module('s7s_build'); B = P.B
import sstate_css as SC
n, dd, icon, sexv, lamp, NB, MINUS = B.n, B.dd, B.icon, B.sexv, B.lamp, B.NB, B.MINUS
B.IC.update({  # weighing = RecordTile preview (scale); alert/check already added by s7s_build
 'weighing': '<path d="M12 3v18"/><path d="M7 21h10"/><path d="M3 7h18"/><path d="m6 7-3 7a3 3 0 0 0 6 0z"/><path d="m18 7-3 7a3 3 0 0 0 6 0z"/>',
})

EX = ROOT / 'design/data/extracts'
S = json.load(open(EX / 'sample-hirbovat-2026.json', encoding='utf-8'))
WW = json.load(open(EX / 'weighings-weekly.json', encoding='utf-8'))
GG = json.load(open(EX / 'guide-growth.json', encoding='utf-8'))
GW = {(r['week'], r['sex']): r['body_weight_g'] for r in GG['rows']}
TSTART = dt.date(2025, 10, 7)  # Telița placement (craft §7); weighings-weekly dates = placement + 7 × age week
assert '2025-10-07' in WW['meta']['conventions']
HSTART = B.START; assert HSTART == dt.date(2026, 2, 10)
def week(d, start): return (d - start).days // 7 + 1  # D02

SIGN = {'F': '♀', 'M': '♂'}; CLS = {'F': 'hen', 'M': 'rooster'}; WORD = {'F': 'găini', 'M': 'cocoși'}; WORDC = {'F': 'Găini', 'M': 'Cocoși'}
def dec(v):  # as the workbook: 8,95 / 7,3 / 68
    if v is None: return None
    return n(v, 0 if v == int(v) else (2 if round(v * 100) % 10 else 1))

# ---------- the three real weighings ----------
def ctx_hirbovat():
    d = dt.date(2026, 11, 16)
    cur = next(w for w in S['weighings'] if w['date'] == d.isoformat() and w['section'] == 'H1.1' and w['sex'] == 'F')
    prev = next(w for w in S['weighings'] if w['date'] == (d - dt.timedelta(days=7)).isoformat() and w['section'] == 'H1.1' and w['sex'] == 'F')
    assert d == B.TODAY and max(w['date'] for w in S['weighings']) == d.isoformat() and cur['guide_weight_g'] is None
    return dict(farm='Hîrbovăț', day=d, start=HSTART, wk=week(d, HSTART), sec='H1.1', sx='F', one_sex=False, guide=None,
                w=cur['avg_weight_g'], cv=cur['cv_pct'], uni=cur['uniformity_pct'], size=cur['sample_size'],
                prev=(dt.date.fromisoformat(prev['date']), prev['cycle_week'], prev['avg_weight_g']))
def ctx_telita(sec, sx, date):
    d = dt.date.fromisoformat(date)
    cur = next(r for r in WW['rows'] if r['section'] == sec and r['sex'] == sx and r['date'] == date)
    prev = next(r for r in WW['rows'] if r['section'] == sec and r['sex'] == sx and r['age_week'] == cur['age_week'] - 1)
    assert {r['sex'] for r in WW['rows'] if r['section'] == sec} == {sx} and cur['sample_size'] is None
    wk = week(d, TSTART); pd = dt.date.fromisoformat(prev['date'])
    return dict(farm='Telița', day=d, start=TSTART, wk=wk, sec=sec, sx=sx, one_sex=True, guide=GW[(wk, sx)],
                w=cur['avg_weight_g'], cv=cur['cv_pct'], uni=cur['uniformity_pct'], size=None,
                prev=(pd, week(pd, TSTART), prev['avg_weight_g']))
HB = ctx_hirbovat()
assert (HB['w'], HB['cv'], HB['uni'], HB['size'], HB['wk'], HB['prev'][2]) == (3967, 7.3, 81.9, 78, 40, 4006)
OK = ctx_telita('H1', 'F', '2026-02-24'); WARN = ctx_telita('H5-2', 'M', '2026-02-17')
def dev(c): return (c['w'] - c['guide']) / c['guide'] * 100 if c['guide'] else None
assert (OK['wk'], OK['guide'], n(dev(OK), 1)) == (21, 2525, '1,1') and abs(dev(OK)) <= 5
assert (WARN['wk'], WARN['guide'], n(dev(WARN), 1)) == (20, 3040, '10,9') and dev(WARN) > 5
def pct(v): return f'{"+" if v > 0 else (MINUS if v < 0 else "")}{n(abs(v), 1)}%'

# ---------- form parts ----------
def head(c):
    h = B.HEAD.replace(f'<span class="ip-tile ip-rec ip-rec--feed" aria-hidden="true">{icon("feed")}</span>',
                       f'<span class="ip-tile ip-rec ip-rec--weighing" aria-hidden="true">{icon("weighing")}</span>').replace('>Hrană zilnică<', '>Cântărire<')
    h = h.replace(f'HÎRBOVĂȚ · {dd(B.TODAY)} · SAPT {B.WEEK}', f'{c["farm"].upper()} · {dd(c["day"])} · SAPT {c["wk"]}')
    assert '>Cântărire<' in h and 'ip-rec--weighing' in h and dd(c['day']) in h
    return h
def date(c, ro=False):
    wl = f'SAPT {c["wk"]} · ciclu din {dd(c["start"])}'
    if ro:  # saved: read-only date (S9 DATE_RO pattern)
        return f'<div class="ip-field"><label class="ip-label" for="d-val">Data</label><span class="ip-input ip-input--touch ip-num is-readonly"><input id="d-val" type="text" value="{dd(c["day"])}" readonly aria-describedby="d-wk"/></span><span class="ip-note ip-num" id="d-wk">{wl}</span></div>'
    out = B.DATE.replace(dd(B.TODAY), dd(c['day'])).replace(f'SAPT {B.WEEK} · ciclu din {dd(B.START)}', wl)
    assert wl in out; return out

def section(c, ro=False):
    if ro:
        return f'<div class="ip-field"><label class="ip-label" for="sec">Secțiune</label><span class="ip-input ip-input--touch is-readonly"><input id="sec" type="text" value="{c["sec"]}" readonly/></span></div>'
    # SectionSelect, phone: a combobox that opens the BottomSheet (S9P §33); closed here
    return (f'<div class="ip-field"><span class="ip-label" id="sec-l">Secțiune</span>'
            f'<div class="ip-input ip-input--touch" id="sec" role="combobox" tabindex="0" aria-haspopup="dialog" aria-expanded="false" aria-labelledby="sec-l sec-v">'
            f'<span class="s8-grow" id="sec-v">{c["sec"]}</span>{icon("chev", 16)}</div></div>')

def sex(c, ro=False):
    sx = c['sx']
    if ro or c['one_sex']:  # one-sex section (Telița) or saved: read-only TextField, sign in its record colour
        hint = f'<span class="ip-hint" id="sx-h">{c["sec"]} are doar {WORD[sx]}.</span>' if c['one_sex'] and not ro else ''
        desc = ' aria-describedby="sx-h"' if hint else ''
        return (f'<div class="ip-field"><label class="ip-label" for="sx">Sex</label><span class="ip-input ip-input--touch is-readonly">'
                f'<span class="ip-sex ip-sex--{CLS[sx]}" aria-hidden="true">{SIGN[sx]}</span><input id="sx" type="text" value="{WORDC[sx]}" readonly{desc}/></span>{hint}</div>')
    items = ''.join(f'<span class="ip-seg__item{" is-selected" if s == sx else ""}" role="radio" aria-checked="{"true" if s == sx else "false"}" tabindex="{0 if s == sx else -1}">'
                    f'<span class="ip-sex ip-sex--{CLS[s]}" aria-hidden="true">{SIGN[s]}</span>{WORDC[s]}</span>' for s in 'FM')
    return f'<div class="ip-field"><span class="ip-label" id="sx-l">Sex</span><span class="ip-seg ip-seg--touch s8-seg" role="radiogroup" aria-labelledby="sx-l">{items}</span></div>'

def msg(kind, mid, text):
    cls, ic = ('ip-warnline', 'warn') if kind == 'warn' else ('ip-err', 'alert')
    return f'<p class="{cls} s8-msg" id="{mid}">{icon(ic, 14)}<span>{text}</span></p>'

def weight(c, v, ro=False, st=None):
    sx = c['sx']; mid = 'w-msg' if st else None
    desc = ' '.join(['ro-sub'] + ([mid] if mid else []))
    m = ''
    if st == 'warn':
        d = dev(c)
        m = msg('warn', mid, f'{c["sec"]} <span class="ip-sex ip-sex--{CLS[sx]}">{SIGN[sx]}</span>: <span class="ip-num">{n(v)}{NB}g</span>, cu <span class="ip-num">{n(abs(d), 1)}%</span> {"peste" if d > 0 else "sub"} ghidul de <span class="ip-num">{n(c["guide"])}{NB}g</span> (prag <span class="ip-num">±5%</span>). Verifică proba; poți salva.')
    val = n(v) if v is not None else ''
    return (f'<div class="ip-field"><label class="ip-label" for="w">Masa medie <span class="ip-sex ip-sex--{CLS[sx]}" aria-hidden="true">{SIGN[sx]}</span><span class="ip-sr">{WORD[sx]}</span></label>'
            f'<span class="ip-input ip-count ip-count--{CLS[sx]}{" is-" + st if st else ""}"><input id="w" inputmode="numeric" pattern="[0-9 &#8239;]*" placeholder="–" autocomplete="off" value="{val}" aria-describedby="{desc}"{" readonly" if ro else ""}/><span class="suffix">g</span></span>{m}</div>')

def tf(fid, label, v, suffix, mode='decimal', opt=True, ro=False, err=None):
    inv = f' aria-invalid="true" aria-describedby="{fid}-msg"' if err else ''
    o = ' <span class="s13-opt">(opțional)</span>' if opt else ''
    return (f'<div class="ip-field"><label class="ip-label" for="{fid}">{label}{o}</label>'
            f'<span class="ip-input ip-input--touch ip-num{" is-readonly" if ro else ""}{" is-error" if err else ""}"><input id="{fid}" type="text" inputmode="{mode}" placeholder="–" autocomplete="off" value="{v if v is not None else ""}"{" readonly" if ro else ""}{inv}/><span class="suffix">{suffix}</span></span></div>')

def readout(c, v, refused=False):
    sx = c['sx']; pdte, pwk, pw = c['prev']
    if c['guide'] is None:
        lmp = '<span class="ip-rlamp ip-rlamp--off">FĂRĂ DATE ÎN GHID</span>'
        # fix 2: the lamp word already says it; the live region hears "FĂRĂ DATE ÎN GHID SAPT n" (no repeat, no bare dash)
        sub = f'<span aria-hidden="true">ghid </span>SAPT {c["wk"]}<span aria-hidden="true">: – fără date în ghid</span>'
    elif v is None:
        lmp = '<span class="ip-rlamp ip-rlamp--off">FĂRĂ DATE</span>'
        sub = f'ghid SAPT {c["wk"]}: {n(c["guide"])} g · scrie masa medie'
    else:
        d = dev(c); ok = abs(d) <= 5  # D19 weight ±5%
        lmp = f'<span class="ip-rlamp ip-rlamp--{"ok" if ok else "warn"}">{"ÎN NORMĂ" if ok else "ABATERE"}</span>'
        sub = f'ghid SAPT {c["wk"]}: {n(c["guide"])} g · {n(v)} g ({pct(d)})'  # spec C7: "ghid sapt. 41: 3 460 g · 3 412 g (−1,4%)"
    val = f'{n(v)}<small> g</small>' if v is not None else '–'
    diff = f'{"+" if v - pw > 0 else MINUS if v - pw < 0 else ""}{n(abs(v - pw))}<small> g</small>' if v is not None else '–'
    return f'''<section class="ip-readout" aria-labelledby="ro-h">
<h2 class="ip-sr" id="ro-h">Față de ghid</h2>
<div class="s8-live" aria-live="polite" aria-atomic="true">
<div class="ip-readout__cap"><span aria-hidden="true">{c["sec"]} {WORD[sx]} · SAPT {c["wk"]}</span>{lmp}</div>
<div class="ip-readout__val" aria-hidden="true">{val}</div>
<div class="ip-readout__sub" id="ro-sub">{sub}</div>
</div>
<dl class="ip-readout__kv"><div><dt>Săpt. trecută · {dd(pdte)}</dt><dd>{n(pw)}<small> g</small></dd></div><div><dt>Diferență</dt><dd>{diff}</dd></div></dl>
</section>'''

def values(c, cv, uni, size, ro=False, uni_err=None):
    m = msg('error', 'uni-msg', uni_err) if uni_err else ''
    return (f'<div class="s8-pair">{tf("cv", "CV", cv, "%", ro=ro)}{tf("uni", "Uniformitate", uni, "%", ro=ro, err=uni_err)}{m}'
            f'{tf("size", "Mărime probă", size, "păsări", mode="numeric", ro=ro)}</div>')

def form(c, w, cv, uni, size, ro=False, st=None, uni_err=None):
    return [date(c, ro), section(c, ro), sex(c, ro), weight(c, w, ro, st), readout(c, w), values(c, cv, uni, size, ro, uni_err)]

def rec14(): return f'<span class="ip-rec ip-rec--weighing">{icon("weighing", 14)}</span>'
def what(c, w):
    if w is None: return f'<span>{c["sec"]} <span class="ip-sex ip-sex--{CLS[c["sx"]]}">{SIGN[c["sx"]]}</span> ·</span><span>–{NB}g</span>'
    return f'<span>{c["sec"]} {sexv(c["sx"], n(w) + NB + "g")}</span>'
SAVE_KEY, SAVE_OFF = P.SAVE_KEY, P.SAVE_OFF
def status(c, w):
    if c['guide'] is None: return '<span>ghid: – fără date în ghid</span>'
    d = dev(c); ok = abs(d) <= 5
    return lamp('ok' if ok else 'warn', 'în normă' if ok else 'abatere') + f'<span><span class="ip-num">{pct(d)}</span> față de ghid</span>'

def screen(parts, save, overlay='', c=HB):
    return f'''<div class="s7-screen s8-screen">
{head(c)}
<div class="s7s-view">
<main class="s7-main{" s8-main--toast" if overlay else ""}">
{"".join(parts)}
</main>
{overlay}
</div>
{save}
</div>'''

boards = []  # (slug, state, body)
C = HB; dv = lambda v: dec(v)
# 0 · default: Hîrbovăț 16.11.2026, H1.1 ♀ typed, before Salvează; no weight in the laying guide (D29)
boards.append(('', None, screen(form(C, C['w'], dv(C['cv']), dv(C['uni']), C['size']),
               B.savebar(rec14() + what(C, C['w']), status(C, C['w']), SAVE_KEY))))
# 1 · gol: the form opened, H1.1 ♀ preselected, nothing typed; the comparison waits; Salvează disabled until Masa medie
boards.append(('gol', 'gol', screen(form(C, None, None, None, None),
               B.savebar(rec14() + what(C, None), '<span>de completat · Masa medie</span>', SAVE_OFF))))
# 2 · în normă: Telița H1 ♀ 24.02.2026 (real), ghid SAPT 21 2 525 g, +1,1%
for slug, state, c, st in (('in-norma', f'în normă · Telița {dd(OK["day"])}', OK, None), ('abatere', f'abatere · Telița {dd(WARN["day"])}', WARN, 'warn')):
    boards.append((slug, state, screen(form(c, c['w'], dv(c['cv']), dv(c['uni']), None, st=st),
                   B.savebar(rec14() + what(c, c['w']), status(c, c['w']), SAVE_KEY), c=c)))
# 4 · valoare refuzată (D39 example): uniformitate 819 typed for 81,9 → outside 0–100%, Salvează disabled, focus to the field
UNI_BAD = 819
err = (f'{C["sec"]} <span class="ip-sex ip-sex--{CLS[C["sx"]]}">{SIGN[C["sx"]]}</span>: uniformitate <span class="ip-num">{UNI_BAD}%</span> e peste limită. '
       f'Scrie o valoare între <span class="ip-num">0</span> și <span class="ip-num">100%</span>.')
boards.append(('valoare-refuzata', 'valoare refuzată · exemplu de stare (D39)', screen(form(C, C['w'], dv(C['cv']), str(UNI_BAD), C['size'], uni_err=err),
               B.savebar(rec14() + what(C, C['w']), lamp('fault', 'refuzat') + f'<span>Uniformitate <span class="ip-num">{UNI_BAD}%</span> ·</span><span>scrie <span class="ip-num">0–100%</span></span>', SAVE_OFF))))
# 5 · salvat: read-only, Toast with Anulează (same day, D07), next step "Cântărire nouă" (as S13 "Livrare nouă", D73)
toast = (f'<div class="s7s-toastwrap"><div class="ip-toast s7s-toast" role="status">{icon("check", 16)}<span class="s7s-toast__txt">Salvat · {C["sec"]} '
         f'{SIGN[C["sx"]]} · <span class="ip-num">{n(C["w"])}</span> g</span><button type="button" class="ip-link s7s-undo">Anulează</button></div></div>')
boards.append(('salvat', 'salvat', screen(form(C, C['w'], dv(C['cv']), dv(C['uni']), C['size'], ro=True),
               B.savebar(rec14() + f'<span>salvat ·</span>{what(C, C["w"])}', '<span>poți modifica până la <span class="ip-num">23:59</span></span>',
                         '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Cântărire nouă</button>'), overlay=toast)))
# 6 · eroare la salvare (Connectivity: online only, draft kept, retry)
ban = (f'<div class="ip-banner ip-banner--fault s7s-banner" role="alert">{icon("alert", 16)}<div class="s7s-banner__body"><p class="s7-m0">Nu s-a salvat cântărirea '
       f'{C["sec"]} <span class="ip-sex ip-sex--{CLS[C["sx"]]}">{SIGN[C["sx"]]}</span> din <span class="ip-num">{dd(C["day"])}</span> (<span class="ip-num">{n(C["w"])}</span> g). Verifică conexiunea și apasă Reîncearcă.</p>'
       f'<p class="s7-m0">Ce ai scris rămâne păstrat în browser până la salvare.</p><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Reîncearcă</button></div></div>')
boards.append(('eroare-la-salvare', 'eroare la salvare', screen([ban] + form(C, C['w'], dv(C['cv']), dv(C['uni']), C['size']),
               B.savebar(rec14() + what(C, C['w']), status(C, C['w']), SAVE_KEY))))

CSS = P.CSS + SC.OPT + '''
.s8-screen{min-height:inherit}
.s8-grow{flex:1;min-width:0}
.s8-seg{display:flex}
.s8-seg > .ip-seg__item{flex:1;justify-content:center}
.s8-live{display:flex;flex-direction:column;gap:inherit;align-items:inherit} /* the live region (lamp word + comparison, fix 1): lays out as the Readout's own column */
.s8-pair{display:grid;grid-template-columns:1fr 1fr;gap:var(--space-4) var(--space-3)}
.s8-pair > .s8-msg{grid-column:1 / -1}
.s8-msg{margin:0;align-items:flex-start}
.s8-msg .ip-icon{flex:none;margin-top:var(--ip-hair)}
.s8-main--toast{padding-bottom:calc(var(--control-h-touch) + var(--space-6) + var(--space-4) + var(--space-8))} /* the last field scrolls clear of the Toast (48 key + 2x12 padding + 16 from the edge) */'''

STEM0, TITLE0 = 'S8-cantarire', 'S8 · Cântărire'
SYS = ROOT / 'design/system'
pathlib.Path('/tmp/s8').mkdir(exist_ok=True)
for slug, state, body in boards:
    stem = STEM0 + (f'--{slug}' if slug else ''); title = TITLE0 + (f' · {state}' if state else '')
    H = HS.get(stem, 844)
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

# ---------- canvas.json: one S8 row below the S14 row (phones only, step 470) ----------
CJ = ROOT / 'design/canvas/project/canvas.json'
cj = json.loads(CJ.read_text(encoding='utf-8'))
s14 = [b for k, b in cj['boards'].items() if k.startswith('S14-')]
Y = max(b['y'] for b in s14) + max(b['h'] for b in s14 if b['y'] == max(x['y'] for x in s14)) + 120
Y = max(Y, 26043, max(b['y'] + b['h'] + 120 for k, b in cj['boards'].items() if not k.startswith('S8-')))
for i, (slug, state, _) in enumerate(boards):
    stem = STEM0 + (f'--{slug}' if slug else ''); f = stem + '.dc.html'
    title = TITLE0 + (f' · {state}' if state else '')
    cj['boards'][f] = {'x': i * 470, 'y': Y, 'w': 390, 'h': HS.get(stem, 844), 'title': title}
    if f not in cj['order']: cj['order'].append(f)
CJ.write_text(json.dumps(cj, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')  # same format as before (round-trips byte-identical)
print('row y', Y)
