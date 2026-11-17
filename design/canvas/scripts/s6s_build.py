"""Build the S6 · Zilnic state artboards (S6S) from the extracts, reusing s6_build.py (as s7s_build reuses s7_build).
python3 -I s6s_build.py <repo_root> [s6s_heights.json]
Imports s6_build (it rewrites S7 at 3162 and S6 at its s6_heights.json height, both byte-identical) and reuses its
entry()/eggs() (state keywords d / ov / saved / ro, default output unchanged), header, date field, SaveBar and CSS;
screen CSS (viewport, Banner, Toast, skeleton) from sstate_css.py, shared with s7s_build (class names s7s-).
Every state board is one 390×844 phone viewport (D60): FormHeader pinned at the top, SaveBar at the foot, the form
scrolled to the card the state concerns. Scroll offsets come from s6s_heights.json ("scroll", measured at 390 by
s6s_render.py; line heights are fixed px tokens, so they do not depend on the font). States: spec C1, D53, D07, D60.
Writes design/canvas/project/S6-zilnic--<state>.dc.html and /tmp/s6s/<stem>.html (plain wrappers for renders)."""
import sys, json, pathlib, importlib, datetime as dt

ROOT = pathlib.Path(sys.argv[1]).resolve()
HP = pathlib.Path(sys.argv[2]) if len(sys.argv) > 2 else pathlib.Path(__file__).resolve().parent / 's6s_heights.json'
HERE = pathlib.Path(__file__).resolve().parent
SCROLL = json.loads(HP.read_text()).get('scroll', {}) if HP.exists() else {}
sys.argv = ['s6_build.py', str(ROOT), str(HERE / 's6_heights.json')]
sys.path.insert(0, str(HERE))
B = importlib.import_module('s6_build')
S7 = B.B
import sstate_css as SC  # viewport / Toast-undo / Banner-action / skeleton CSS, shared with s7s_build
n, dd, sexv, lamp, icon, NB, MINUS = B.n, B.dd, B.sexv, B.lamp, B.icon, B.NB, B.MINUS
TODAY, YDAY, START, S, FS = B.TODAY, B.YDAY, B.START, B.S, B.FS
S7.IC.update({'info': '<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>'})  # Banner info preview
VH = 844
EV = {e['at'][:10]: e for e in S['event_log'] if e['record'] == 'mortality' and e['action'] == 'create'}

# ---------- halls for a day and a state ----------
def halls(d=TODAY, ov=None, saved=None, ro=False):
    parts, ALL, EG = [], [], []
    for hl in B.HALLS:
        h = hl['code']; secs = [s['code'] for s in B.SECS[hl['id']]]
        cards = [B.entry(s, d, (ov or {}).get(s), saved) if (ov or saved or d != TODAY) else B.entry(s) for s in secs]
        ALL += list(zip(secs, cards))
        af = sum(c[1]['F']['alive_start'] for c in cards); am = sum(c[1]['M']['alive_start'] for c in cards)
        eh, tot, lay, ok = B.eggs(h, secs, af, d, ro); EG.append((h, tot, lay, ok))
        parts.append(f'''<section class="ip-hall" aria-labelledby="h-{h}"><header class="ip-hall__head"><span class="ip-cap">Hală</span><h2 class="ip-hall__code" id="h-{h}">{h}</h2><span class="ip-note">secțiuni <span class="ip-num">{" · ".join(secs)}</span></span><span class="ip-pair"><span class="ip-sr">în viață </span>{sexv("F", n(af))}<span class="sep">·</span>{sexv("M", n(am))}</span></header>
<div class="ip-hall__sections">{"".join(c[0] for c in cards)}</div>
{eh}</section>''')
    return '<div class="ip-halls">' + ''.join(parts) + '</div>', ALL, EG

def removals(ALL, skip=()):  # Mortalitate + Sacrificare + Eroare sexare per sex; skip = refused (section, sex), left out (as S7S)
    return tuple(sum(c[1][sx]['mortality'] + c[1][sx]['culls'] + c[1][sx]['sexing_errors'] for s, c in ALL if (s, sx) not in skip) for sx in 'FM')

def rec14(k): return f'<span class="ip-rec ip-rec--{k}">{icon(k, 14)}</span>'
def line1(ALL, rem, lead=''):
    return (f'{rec14("mortality")}<span>{lead}<span class="ip-num">{len(ALL)}</span> secțiuni ·</span>'
            f'<span>{sexv("F", f"{MINUS}{rem[0]}" if rem[0] else "0")} · {sexv("M", f"{MINUS}{rem[1]}" if rem[1] else "0")}</span>')
def eggseg(EG): return f'{rec14("eggs")}<span><span class="ip-num">{n(sum(x[1] for x in EG))}</span> ouă ·</span>'
SAVE_KEY = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Salvează</button>'
def save_off(desc='sb-sum'): return f'<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="{desc}">Salvează</button>'
EDIT_KEY = '<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Editează</button>'

def head(meta):  # S6 FormHeader with another context line (farm · day · week)
    old = f'HÎRBOVĂȚ · {dd(TODAY)} · SAPT {B.WEEK}'; assert old in B.HEAD
    return B.HEAD.replace(old, meta)
def date(d, week_line=None, ieri=False):
    t = B.DATE
    if ieri:
        a = '<span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Azi</span><span class="ip-seg__item" role="radio" aria-checked="false" tabindex="-1">Ieri</span>'
        assert a in t
        t = t.replace(a, '<span class="ip-seg__item" role="radio" aria-checked="false" tabindex="-1">Azi</span><span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Ieri</span>')
    t = t.replace(f'value="{dd(TODAY)}"', f'value="{dd(d)}"')
    if week_line:
        old = f'SAPT {B.WEEK} · ciclu din {dd(START)}'; assert old in t
        t = t.replace(old, week_line)
    return t

def screen(main_parts, save, slug, hd=None, view_cls='', overlay=''):
    sc = SCROLL.get(slug, 0)
    style = f' style="margin-top:calc(-1 * {sc}px)"' if sc else ''
    return f'''<div class="s7-screen s7s-vp">
{hd or B.HEAD}
<div class="s7s-view{(" " + view_cls) if view_cls else ""}">
<main class="s7-main"{style}>
{"".join(main_parts)}
</main>
{overlay}
</div>
{save}
</div>'''

def saved_meta(at, who, line2):
    return f'<div class="ip-entry__saved"><span>Salvat la <span class="ip-num">{at}</span> · {who}</span><span>{line2}</span></div>'

boards, TARGET = [], {}

# 1 · se încarcă — static Skeleton (no shimmer); header and date are known from the route. Geometry = the loaded card
# (head: code 22 + alive 20; grid: label 16 + CountInput 56 + hint 16, twice; Mai mult key 48; foot = Checkbox row).
def skel(w, h): return f'<div class="ip-skel" style="width:{w};height:{h}"></div>'
SK16, SK20, SK22 = 'var(--ip-lh-small)', 'var(--ip-lh-figure)', 'var(--ip-lh-heading)'
def skfield(hint=True): return f'<div class="s7s-skcol">{skel("70%", SK16)}{skel("100%", "var(--count-input-h)")}{skel("80%", SK16) if hint else ""}</div>'
skcard = (f'<section class="ip-panel ip-entry" aria-hidden="true"><div class="s7s-skcol">{skel("20%", SK22)}{skel("55%", SK20)}</div>'
          f'<div class="ip-entry__grid">{skfield()}{skfield()}{skfield(False)}{skfield(False)}</div>'
          f'<div class="s7s-skcol">{skel("50%", "var(--control-h-touch)")}</div>'
          f'<div class="ip-entry__foot">{skel("45%", SK20)}</div></section>')
skhall = f'<section class="ip-hall" aria-hidden="true"><div class="ip-hall__head">{skel("40%", SK22)}</div><div class="ip-hall__sections">{skcard * 2}</div></section>'
sk = f'''<p class="ip-sr" role="status">Se încarcă secțiunile pentru <span class="ip-num">{dd(TODAY)}</span>…</p>
<div class="s7s-skblock" aria-busy="true"><div class="ip-halls">{skhall}</div></div>'''
sv1 = f'''<div class="ip-savebar s7-savebar"><span class="ip-savebar__sum s7s-skcol" aria-hidden="true">{skel("70%", SK16)}{skel("45%", SK16)}</span><button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled>Salvează</button></div>'''
boards.append(('se-incarca', 'se încarcă', screen([B.DATE, sk], sv1, 'se-incarca')))

# 2 · deja introdus — 16.11 opened again after the 08:34 save: read-only cards, Tag "salvat", one form-level Editează (D60)
ev = EV[TODAY.isoformat()]; at = ev['at'][11:]
h2, A2, E2 = halls(saved=saved_meta(at, ev['user'], f'poți modifica până la <span class="ip-num">23:59</span>'), ro=True)
r2 = removals(A2); assert ev['summary'] == f'Zilnic salvat · {len(A2)} secțiuni · ♀ {MINUS}{r2[0]} · ♂ {MINUS}{r2[1]}'
SAVED_BAR = S7.savebar(line1(A2, r2, 'salvat · '), eggseg(E2) + f'<span>{ev["user"]} · <span class="ip-num">{at}</span></span>', EDIT_KEY)
boards.append(('deja-introdus', 'deja introdus', screen([B.DATE, h2], SAVED_BAR, 'deja-introdus')))

# 3 · zi trecută (medic) — Ieri 15.11: saved by Dr. Ceban at 08:20, read-only for a vet (D07), info Banner with the spec line
evy = EV[YDAY.isoformat()]; aty = evy['at'][11:]
wky = (YDAY - START).days // 7 + 1
h3, A3, E3 = halls(d=YDAY, saved=saved_meta(aty, evy['user'], 'o poate modifica managerul'), ro=True)
r3 = removals(A3); assert evy['summary'] == f'Zilnic salvat · {len(A3)} secțiuni · ♀ {MINUS}{r3[0]} · ♂ {MINUS}{r3[1]}' and wky == 40
ban3 = f'''<div class="ip-banner ip-banner--info s7s-banner" id="past-why">{icon("info", 16)}<div class="s7s-banner__body"><p class="s7-m0">Pentru zilele trecute cere managerului.</p><p class="s7-m0">Ziua <span class="ip-num">{dd(YDAY)}</span> e salvată; o vezi doar în citire.</p></div></div>'''
sv3 = S7.savebar(line1(A3, r3, f'<span class="ip-num">{dd(YDAY)}</span> · '), eggseg(E3) + '<span>doar citire</span>', save_off('past-why'))
boards.append(('zi-trecuta', 'zi trecută (medic)', screen([date(YDAY, ieri=True), ban3, h3], sv3, 'zi-trecuta',
               hd=head(f'HÎRBOVĂȚ · {dd(YDAY)} · SAPT {wky}'))))

# 4 · abatere 3× — D39 state example (SYS05): H1.1 ♀ 4 typed vs the 7-day mean 1,0 (literal rule, D53); saveable
OV4 = {'H1.1': {('mortality', 'F'): 4}}
assert B.hint('H1.1', 'F')[1] == 1.0 and B.day('H1.1', 'F')['mortality'] != 4
h4, A4, E4 = halls(ov=OV4)
r4 = removals(A4)
bad4 = [h for h, _, _, ok in E4 if not ok]
sv4 = S7.savebar(line1(A4, r4), eggseg(E4) + lamp('warn', 'abatere') + '<span>H1.1 <span class="ip-sex ip-sex--hen">♀</span>, ' + ', '.join(bad4) + '</span>', SAVE_KEY)
boards.append(('abatere', 'abatere 3× · exemplu de stare (D39)', screen([B.DATE, h4], sv4, 'abatere')))
TARGET['abatere'] = 't-H11'

# 5 · refuzat — D39 state example (SYS05): H1.2 ♂ 4 370 typed (437 alive at the start of 16.11): save disabled
SEC5, SX5 = 'H1.2', 'M'
OV5 = {SEC5: {('mortality', SX5): 4370}}
assert B.day(SEC5, SX5)['alive_start'] == 437
h5, A5, E5 = halls(ov=OV5)
t5 = dict(A5)[SEC5][1][SX5]; assert B.removed(t5) > t5['alive_start']  # the card's own figures (as its error line)
r5 = removals(A5, skip={(SEC5, SX5)})
sv5 = S7.savebar(line1(A5, r5), f'{lamp("fault", "refuzat")}<span>{SEC5} <span class="ip-sex ip-sex--{B.CLS[SX5]}">{B.SIGN[SX5]}</span> <span class="ip-num">{n(t5["mortality"])}</span> ·</span><span>cel mult <span class="ip-num">{n(B.most(t5))}</span></span>', save_off())
boards.append(('refuzat', 'refuzat · peste cei în viață · exemplu de stare (D39)', screen([B.DATE, h5], sv5, 'refuzat')))
TARGET['refuzat'] = 't-H12'

# 6 · eroare la salvare — connectivity (spec §1): fault Banner + Reîncearcă, typed values kept (draft), Salvează enabled (D60)
r6 = removals(B.ALL); eg6 = sum(x[1] for x in B.EGGS)
assert (r6, eg6) == ((8, 1), 18688)
ban6 = f'''<div class="ip-banner ip-banner--fault s7s-banner" role="alert">{icon("alert", 16)}<div class="s7s-banner__body"><p class="s7-m0">Nu s-a salvat Zilnic din <span class="ip-num">{dd(TODAY)}</span> (<span class="ip-num">{len(B.ALL)}</span> secțiuni · {sexv("F", f"{MINUS}{r6[0]}")} · {sexv("M", f"{MINUS}{r6[1]}")} · <span class="ip-num">{n(eg6)}</span> ouă). Verifică conexiunea și apasă Reîncearcă.</p><p class="s7-m0">Ce ai scris rămâne păstrat în browser până la salvare.</p><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Reîncearcă</button></div></div>'''
boards.append(('eroare-la-salvare', 'eroare la salvare', screen([ban6, B.DATE, B.HALLS_HTML], B.SAVE, 'eroare-la-salvare')))

# 7 · fermă fără păsări — Telița: cycle closed 28.02.2026, 0 birds (counts-latest, real): EmptyState with the next step
C = json.load(open(ROOT / 'design/data/extracts/counts-latest.json', encoding='utf-8'))
tl = next(f for f in C['farms'] if f['farm_id'] == 'telita')
assert tl['cycle_status'] == 'closed' and tl['totals']['all'] == 0
end = dt.date.fromisoformat(tl['cycle_end'])
empty = f'''<div class="ip-empty"><span class="ip-lamp ip-lamp--off">ciclu închis</span><div>Ferma e goală din <span class="ip-num">{dd(end)}</span>. Ciclul pornește la prima populare sau transfer.</div><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Alege altă fermă</button></div>'''
sv7 = f'''<div class="ip-savebar s7-savebar"><span class="ip-savebar__sum ip-ftot-sum" id="sb-sum"><span class="ip-ftot-sum__line">{rec14("mortality")}<span><span class="ip-num">0</span> secțiuni cu păsări</span></span></span>{save_off()}</div>'''
boards.append(('ferma-fara-pasari', 'fermă fără păsări', screen([date(TODAY, f'fără ciclu deschis · ultimul închis pe {dd(end)}'), empty], sv7, 'ferma-fara-pasari',
               hd=head(f'TELIȚA · {dd(TODAY)}'))))

# 8 · salvat — right after Salvează at the foot of the form: Toast (spec C1 copy) + Anulează, read-only, Editează (D60)
toast = f'''<div class="s7s-toastwrap"><div class="ip-toast s7s-toast" role="status">{icon("check", 16)}<span class="s7s-toast__txt">Salvat · <span class="ip-num">{len(A2)}</span> secțiuni · ♀ <span class="ip-num">{MINUS}{r2[0]}</span> · ♂ <span class="ip-num">{MINUS}{r2[1]}</span></span><button type="button" class="ip-link s7s-undo">Anulează</button></div></div>'''
boards.append(('salvat', 'salvat', screen([B.DATE, h2], SAVED_BAR, 'salvat', view_cls='s7s-view--end', overlay=toast)))

CSS = B.CSS + SC.VIEW + SC.BANNER  # S7S's screen CSS (sstate_css), class names kept s7s- as S6 keeps s7-screen/s7-main
CSS_EXTRA = {
 'salvat': SC.TOAST,
 'se-incarca': SC.SKEL + '''
.s7s-skcol > .ip-skel{flex:none}''',
}  # toast / skeleton rules only on the board that uses them

STEM0, TITLE0 = 'S6-zilnic', 'S6 · Zilnic'
SYS = ROOT / 'design/system'
pathlib.Path('/tmp/s6s').mkdir(exist_ok=True)
meta = []
for slug, state, body in boards:
    stem, title = f'{STEM0}--{slug}', f'{TITLE0} · {state}'
    root = f'<div class="ip-root" style="padding:0; width:390px; height:{VH}px; overflow:hidden">{body}</div>'
    css = CSS + CSS_EXTRA.get(slug, '')
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
{B.B.FONTS}
<style>{css}</style>
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
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{B.B.FONTS}<style>{css}</style></head><body>{root}</body></html>'''
    pathlib.Path(f'/tmp/s6s/{stem}.html').write_text(wrap, encoding='utf-8')
    meta.append({'stem': stem, 'slug': slug, 'title': title, 'target': TARGET.get(slug)})
    print('ok', stem, title, len(dc), 'scroll', SCROLL.get(slug, 0))
pathlib.Path('/tmp/s6s/boards.json').write_text(json.dumps(meta, ensure_ascii=False))
print('removals', r2, r3, r4, r5)
