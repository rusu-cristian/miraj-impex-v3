"""Build the S6 · Zilnic state artboards (S6S) from realdata (RF04, today 18.11.2026), reusing s6_build.py (as s7s_build reuses s7_build).
python3 -I s6s_build.py <repo_root> [s6s_heights.json] [--canvas]   (--canvas: refresh the 9 entries' h/title in place, x and y kept)
Imports s6_build (it rewrites S7 at its default height and S6 at its s6_heights.json height, both byte-identical) and reuses its
entry()/eggs()/halls() (state keywords d / ov / saved / ro, default output unchanged), header, date field, SaveBar and CSS;
screen CSS (viewport, Banner, Toast, skeleton) from sstate_css.py, shared with s7s_build (class names s7s-).
Every state board is one 390×844 phone viewport (D60): FormHeader pinned at the top, SaveBar at the foot, the form
scrolled to the card the state concerns. Scroll offsets come from s6s_heights.json ("scroll", measured at 390 by
s6s_render.py --measure; line heights are fixed px tokens, so they do not depend on the font). States: spec C1, D53, D07, D60.
Real stories on the 18.11 figures: se încarcă, deja introdus (the 08:15 save), zi trecută (17.11, saved 08:07), eroare la salvare,
fermă fără păsări (Telița, closed 28.02.2026), salvat. No real 18.11 entry triggers the 3× warning, a refusal or a cull, so
abatere 3×, refuzat and motiv lipsă are D39 state examples typed on the real counts (title "· exemplu de stare (D39)").
Writes design/canvas/project/S6-zilnic--<state>.dc.html and /tmp/s6s/<stem>.html (plain wrappers for renders)."""
import sys, json, pathlib, importlib, datetime as dt

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
import vkey as V   # VO02: the voice key on every signed-in board (spec §8, D199; vkey.add in the writer)
ROOT, HP, CANVAS = CR.args('s6s_heights.json'); ROOT = ROOT.resolve()
HERE = pathlib.Path(__file__).resolve().parent
SCROLL = json.loads(HP.read_text()).get('scroll', {}) if HP.exists() else {}
sys.argv = ['s6_build.py', str(ROOT), str(HERE / 's6_heights.json')]
sys.path.insert(0, str(HERE))
B = importlib.import_module('s6_build')
S7 = B.B
R = B.R
import skel as SKL
import sstate_css as SC  # viewport / Toast-undo / Banner-action / skeleton CSS, shared with s7s_build
import sexsign as SXS  # SX01: ♀/♂ outside a figure span = the §72 hook (sexsign.py)
n, dd, sexv, lamp, icon, NB, MINUS = B.n, B.dd, B.sexv, B.lamp, B.icon, B.NB, B.MINUS
TODAY, YDAY, START = B.TODAY, B.YDAY, B.START
S7.IC.update({'info': '<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>'})  # Banner info preview
VH = 844
D39 = ' · exemplu de stare (D39)'

# ---------- halls for a day and a state ----------
def halls(d=TODAY, ov=None, saved=None, ro=False, kw=None):  # kw (S6M): {section: entry() keywords, e.g. reasons / more}
    if not (ov or saved or d != TODAY or kw): return B.HALLS_HTML, B.ALL, B.EGGS
    return B.halls(d, lambda s: B.entry(s, d, (ov or {}).get(s), saved, **(kw or {}).get(s, {})), ro)

def removals(ALL, skip=()):  # Mortalitate + Sacrificare + Eroare sexare per sex; skip = refused (section, sex), left out (as S7S)
    return tuple(sum(c[1][sx]['mortality'] + c[1][sx]['culls'] + c[1][sx]['sexing_errors'] for s, c in ALL if (s, sx) not in skip) for sx in 'FM')

def rec14(k): return f'<span class="ip-rec ip-rec--{k}">{icon(k, 14)}</span>'
def line1(ALL, rem, lead=''):
    return (f'{rec14("mortality")}<span>{lead}<span class="ip-num">{len(ALL)}</span> secțiuni ·</span>'
            f'<span>{sexv("F", B.signed(rem[0]))} · {sexv("M", B.signed(rem[1]))}</span>')
def summary_ok(ev, ALL, rem):  # the event log's save of these values (illustrative author/time, real figures)
    return ev['summary'] == f'Zilnic salvat · {len(ALL)} secțiuni · mortalitate ♀ {rem[0]} · ♂ {rem[1]}'
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
    return f'''<div class="ip-screen ip-screen--fixed">
{hd or B.HEAD}
<div class="ip-screen__view{(" " + view_cls) if view_cls else ""}">
<main class="ip-screen__main"{style}>
{"".join(main_parts)}
</main>
{overlay}
</div>
{save}
</div>'''

def saved_meta(at, who, line2):
    return f'<div class="ip-entry__saved"><span>Salvat la <span class="ip-num">{at}</span> · {who}</span><span>{line2}</span></div>'

boards, TARGET = [], {}

# 1 · se încarcă — static Skeleton (no shimmer); header and date are known from the route. Geometry = the loaded geometry,
# so nothing shifts when the cards arrive (as S7P; measured at 390 by the S6P audit: hall head 35, card 393, cards 12 apart,
# SaveBar 73): head = code 22 + alive 20; grid = label 16 + CountInput 56 + hint 16, then label 16 + CountInput 56; Mai mult
# key 48; foot = the Checkbox row itself (an aria-hidden `ip-check` box holding the bar, so it takes the system row height).
def skel(w, h): return f'<div class="ip-skel" style="width:{w};height:{h}"></div>'
SK16, SK20, SK22 = 'var(--ip-lh-small)', 'var(--ip-lh-figure)', 'var(--ip-lh-heading)'
# FIN11G: fields are the §54 field skeleton (label line, CountInput box, hint line), the head the card head's own lines
def skfield(hint=True): return SKL.field('count', 1 if hint else 0)
skcard = (f'<section class="ip-panel ip-entry" aria-hidden="true"><div class="ip-entry__head">{SKL.line("ip-skel--w3", cls="ip-entry__code")}'
          f'<span class="ip-entry__alive">{SKL.line("ip-skel--w2", cls="ip-pair")}</span></div>'
          f'<div class="ip-entry__grid">{skfield()}{skfield()}{skfield(False)}{skfield(False)}</div>'
          f'{skel("50%", "var(--control-h-touch)")}'
          f'<div class="ip-entry__foot"><span class="ip-check s6s-skcheck">{skel("calc(var(--space-12) * 3)", SK20)}</span></div></section>')
skhall = f'<section class="ip-hall" aria-hidden="true"><div class="ip-hall__head">{skel("40%", SK22)}</div><div class="ip-hall__sections">{skcard * 2}</div></section>'
sk = f'''<p class="ip-sr" role="status">Se încarcă secțiunile pentru <span class="ip-num">{dd(TODAY)}</span>…</p>
<div class="s7s-skblock" aria-busy="true"><div class="ip-halls">{skhall}</div></div>'''
sv1 = f'''<div class="ip-savebar ip-screen__savebar"><span class="ip-savebar__sum ip-savebar__sum--2l" aria-hidden="true">{SKL.line("ip-skel--w1")}{SKL.line("ip-skel--w2")}</span><button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled>Salvează</button></div>'''
boards.append(('se-incarca', 'se încarcă', screen([B.DATE, sk], sv1, 'se-incarca')))

# 2 · deja introdus — 18.11 opened again after the 08:15 save: read-only cards, Tag "salvat", one form-level Editează (D60)
ev = B.EV; at = ev['at'][11:]
h2, A2, E2 = halls(saved=saved_meta(at, ev['user'], f'poți modifica până la <span class="ip-num">23:59</span>'), ro=True)
r2 = removals(A2); assert summary_ok(ev, A2, r2)
SAVED_BAR = S7.savebar(line1(A2, r2, 'salvat · '), eggseg(E2) + f'<span>{ev["user"]} · <span class="ip-num">{at}</span></span>', EDIT_KEY)
boards.append(('deja-introdus', 'deja introdus', screen([B.DATE, h2], SAVED_BAR, 'deja-introdus')))

# 3 · zi trecută (medic) — Ieri 17.11: saved by Dr. Ceban at 08:07 (real deaths), read-only for a vet (D07), info Banner with the spec line
evy = B.ev(YDAY, 'mortality'); aty = evy['at'][11:]
wky = R.sapt(YDAY)
h3, A3, E3 = halls(d=YDAY, saved=saved_meta(aty, evy['user'], 'o poate modifica managerul'), ro=True)
r3 = removals(A3); assert summary_ok(evy, A3, r3) and r3 == (R.deaths(YDAY, sx='F'), R.deaths(YDAY, sx='M'))
ban3 = f'''<div class="ip-banner ip-banner--info s7s-banner" id="past-why">{icon("info", 16)}<div class="s7s-banner__body"><p class="ip-m0">Pentru zilele trecute cere managerului.</p><p class="ip-m0">Ziua <span class="ip-num">{dd(YDAY)}</span> e salvată; o vezi doar în citire.</p></div></div>'''
sv3 = S7.savebar(line1(A3, r3, f'<span class="ip-num">{dd(YDAY)}</span> · '), eggseg(E3) + '<span>doar citire</span>', save_off('past-why'))
boards.append(('zi-trecuta', 'zi trecută (medic)', screen([date(YDAY, ieri=True), ban3, h3], sv3, 'zi-trecuta',
               hd=head(f'HÎRBOVĂȚ · {dd(YDAY)} · SAPT {wky}'))))

# 4 · abatere 3× — D39 state example (SYS05): no real 18.11 death is above 3 × the 7-day mean, so H5.1 ♀ 4 is typed (the real
# count was 1) vs the mean of 11–17.11 (D53); saveable
SEC4, SX4, Q4 = 'H5.1', 'F', 4
OV4 = {SEC4: {('mortality', SX4): Q4}}
assert Q4 > 3 * B.hint(SEC4, SX4)[1] and B.day(SEC4, SX4)['mortality'] != Q4
h4, A4, E4 = halls(ov=OV4)
r4 = removals(A4)
bad4 = [h for h, _, _, ok in E4 if not ok]
sv4 = S7.savebar(line1(A4, r4), eggseg(E4) + lamp('warn', 'abatere') + f'<span>{SEC4} <span class="ip-sex ip-sex--{B.CLS[SX4]}">{B.SIGN[SX4]}</span>, ' + ', '.join(bad4) + '</span>', SAVE_KEY)
boards.append(('abatere', 'abatere 3×' + D39, screen([B.DATE, h4], sv4, 'abatere')))
TARGET['abatere'] = 't-' + SEC4.replace('.', '')

# 5 · refuzat — D39 state example (SYS05): H5.2 ♂ typed with an extra 0 (ten times the cocoși alive at the start of 18.11): save disabled
SEC5, SX5 = 'H5.2', 'M'
OV5 = {SEC5: {('mortality', SX5): 10 * B.day(SEC5, SX5)['alive_start']}}
h5, A5, E5 = halls(ov=OV5)
t5 = dict(A5)[SEC5][1][SX5]; assert B.removed(t5) > t5['alive_start']  # the card's own figures (as its error line)
r5 = removals(A5, skip={(SEC5, SX5)})
sv5 = S7.savebar(line1(A5, r5), f'{lamp("fault", "refuzat")}<span>{SEC5} <span class="ip-sex ip-sex--{B.CLS[SX5]}">{B.SIGN[SX5]}</span> <span class="ip-num">{n(t5["mortality"])}</span> ·</span><span>cel mult <span class="ip-num">{n(B.most(t5))}</span></span>', save_off())
boards.append(('refuzat', 'refuzat · peste cei în viață' + D39, screen([B.DATE, h5], sv5, 'refuzat')))
TARGET['refuzat'] = 't-' + SEC5.replace('.', '')

# 6 · eroare la salvare — connectivity (spec §1): fault Banner + Reîncearcă (focus, S16P), typed values kept (draft), Salvează enabled (D60);
#   the Banner ends the form, scrolled to its end, right above the SaveBar (D141, as S9/S10)
r6 = removals(B.ALL); eg6 = sum(x[1] for x in B.EGGS)
assert (r6, eg6) == ((B.remF, B.remM), R.eggs())
ban6 = f'''<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{icon("alert", 16)}<div class="ip-banner__body"><p class="ip-m0">Zilnic nu s-a salvat: <span class="ip-num">{len(B.ALL)}</span> secțiuni · {sexv("F", B.signed(r6[0]))} · {sexv("M", B.signed(r6[1]))} · <span class="ip-num">{n(eg6)}</span> ouă, <span class="ip-num">{dd(TODAY)}</span>.</p><p class="ip-m0">Verifică conexiunea la internet, apoi apasă Reîncearcă. Ce ai scris rămâne păstrat în browser până la salvare.</p></div><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-banner__key ip-focus" autofocus="autofocus">Reîncearcă</button></div>'''
boards.append(('eroare-la-salvare', 'eroare la salvare', screen([B.DATE, B.HALLS_HTML, ban6], B.SAVE, 'eroare-la-salvare', view_cls='ip-screen__view--end')))

# 7 · fermă fără păsări — Telița: cycle closed 28.02.2026, 0 birds (counts-latest, real): EmptyState with the next step
tl = R.COUNTS_TELITA
assert tl['cycle_status'] == 'closed' and tl['totals']['all'] == 0
end = R.D(tl['cycle_end']); assert end == R.D(R.CTEL['end_date'])
empty = f'''<div class="ip-empty"><span class="ip-lamp ip-lamp--off">ciclu închis</span><div>Ferma e goală din <span class="ip-num">{dd(end)}</span>. Ciclul pornește la prima populare sau transfer.</div><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Alege altă fermă</button></div>'''
sv7 = f'''<div class="ip-savebar ip-screen__savebar"><span class="ip-savebar__sum ip-savebar__sum--2l" id="sb-sum"><span class="ip-savebar__line">{rec14("mortality")}<span><span class="ip-num">0</span> secțiuni cu păsări</span></span></span>{save_off()}</div>'''
boards.append(('ferma-fara-pasari', 'fermă fără păsări', screen([date(TODAY, f'fără ciclu deschis · ultimul închis pe {dd(end)}'), empty], sv7, 'ferma-fara-pasari',
               hd=head(f'TELIȚA · {dd(TODAY)}'))))

# 8 · salvat — right after Salvează at the foot of the form: Toast (spec C1 copy, object first: D143) + Anulează, read-only, Editează (D60)
toast = SXS.sexify(f'''<div class="ip-toastwrap"><div class="ip-toast" role="status">{icon("check", 16)}<span class="ip-toast__txt">Zilnic salvat · <span class="ip-num">{len(A2)}</span> secțiuni · ♀&#160;<span class="ip-num">{B.signed(r2[0])}</span> · ♂&#160;<span class="ip-num">{B.signed(r2[1])}</span></span><button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>''')  # SX01: §72 signs
boards.append(('salvat', 'salvat', screen([B.DATE, h2], SAVED_BAR, 'salvat', view_cls='ip-screen__view--end', overlay=toast)))

# 9 · motiv lipsă (S6M, D79) — D39 state example (no illustrative cull on 18.11): H5.1 ♀ Sacrificare 2 typed, Motiv left on
# "Alege" (refusal, D53: blocks save) and Mai mult open on that card with the Eroare sexare fields (18.11 values, 0). The 2 culls
# without a reason are left out of the SaveBar totals (refused value, as S7S / D77); Salvează disabled, described by the SaveBar line.
SEC9, SX9, Q9 = 'H5.1', 'F', 2
assert B.day(SEC9, SX9)['culls'] == 0 and all(B.day(SEC9, x)['sexing_errors'] == 0 for x in 'FM')
h9, A9, E9 = halls(ov={SEC9: {('culls', SX9): Q9}}, kw={SEC9: {'reasons': {SX9: ''}, 'more': True}})
t9 = dict(A9)[SEC9][1][SX9]; assert t9['culls'] == Q9 and B.removed(t9) <= t9['alive_start']
r9 = removals(A9); r9 = (r9[0] - Q9, r9[1]); assert r9 == removals(B.ALL)  # = the default board's removals
sv9 = S7.savebar(line1(A9, r9), f'{lamp("fault", "refuzat")}<span>{SEC9} <span class="ip-sex ip-sex--{B.CLS[SX9]}">{B.SIGN[SX9]}</span> · Motiv lipsă</span>', save_off())
boards.append(('motiv-lipsa', 'motiv lipsă · Mai mult deschis' + D39, screen([B.DATE, h9], sv9, 'motiv-lipsa')))
TARGET['motiv-lipsa'] = 't-' + SEC9.replace('.', '')

CSS = B.CSS + SC.VIEW + SC.BANNER  # S7S's screen CSS (sstate_css), class names kept s7s- as S6 keeps ip-screen/ip-screen__main
CSS_EXTRA = {
 'se-incarca': SC.SKEL + '''
.s6s-skcheck{cursor:default}''',
}  # skeleton rules only on the board that uses them (the Toast is additions.css §48, FIN02)

STEM0, TITLE0 = 'S6-zilnic', 'S6 · Zilnic'
SYS = ROOT / 'design/system'
pathlib.Path('/tmp/s6s').mkdir(exist_ok=True)
meta = []
for slug, state, body in boards:
    stem, title = f'{STEM0}--{slug}', f'{TITLE0} · {state}'
    root = f'<div class="ip-root" style="padding:0; width:390px; height:{VH}px; overflow:hidden">{V.add(body)}</div>'
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
if CANVAS and __name__ == '__main__':  # RF04: refresh the 9 hand-placed entries' h/title in place (x, y kept; canvas_rows.place)
    c = CR.load(ROOT); CR.place(c, STEM0, [(m['stem'], 390, VH, m['title']) for m in meta]); CR.save(ROOT, c)
print('removals', r2, r3, r4, r5, r9)
