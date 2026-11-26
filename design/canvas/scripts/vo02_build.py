"""VO02 · the 13 voice boards VO-1 … VO-13 (screens.md › Voice › VO02 boards; spec §8 F1, D199), row "VO · Comenzi vocale"
right under V1 · Voce. Each board is a real screen of 18.11.2026 (Hîrbovăț, Seria 2026, SAPT 59) drawn by its own generator
(imported: s2_build S2 vet page, s6_build Zilnic cards, s7_build Hrană zilnică rows, s4t/s4ns S4 parts and the phone shell,
s14_build Producție, s14s phone Producție table, s13_build Livrare hrană over the S11 manager page), with VO01's VoiceButton /
VoicePanel / VoiceMark helpers (design/system/scripts/vo01_build.py through vkey, imported). Spoken figures not in the
workbook (♀ 3 · ♂ 1, „2 găini”) and the state examples are marked D39 in the title.
python3 -I vo02_build.py <repo_root> [--canvas]
Writes design/canvas/project/VO-*.dc.html and /tmp/vo02/<stem>.html (render wrappers); heights / scroll from vo02_heights.json
(vo02_render.py --measure); --canvas adds or refreshes the row under V1 (every later row moves down by its height + 120)."""
import sys, re, json, pathlib, importlib, importlib.util, datetime as dt

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve(); HERE = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
import vkey as V
VO = V.VO
HJ = HERE / 'vo02_heights.json'
HS = json.loads(HJ.read_text()) if HJ.exists() else {}


def load(name):   # the generators read their repo root from argv; none of them gets --canvas here
    sys.argv = [f'{name}.py', str(ROOT)]
    if name in sys.modules: return sys.modules[name]
    sp = importlib.util.spec_from_file_location(name, HERE / f'{name}.py')
    m = importlib.util.module_from_spec(sp); sys.modules[name] = m; sp.loader.exec_module(m)
    return m


A = load('s2_build')        # S2 phone page (vet), tab bar, phone CSS (rewrites the S2 / S7 default boards unchanged)
S6 = load('s6_build')       # Zilnic: entry(), halls(), HEAD, DATE (rewrites S6 / S7 unchanged)
B = S6.B                    # s7_build: Hrană zilnică rows, SaveBar, FormHeader
NS = load('s4ns_build')     # S4 phone shell parts (rewrites its boards unchanged)
T, s4, P = NS.T, NS.s4, NS.P
X = load('s14_build')       # Producție page (writes nothing on import)
Q = load('s13_build')       # Livrare hrană desktop over the S11 manager page (rewrites S13 / S7 unchanged)
R = s4.R
SX = VO.SXS
n, dd = R.n, R.dd
TODAY = R.TODAY
assert R.WK == 59 and dd(TODAY) == '18.11.2026'

OUT = []                    # (stem, title, w, h)


# ---------------------------------------------------------------------------------------------------------- writers
def phone_doc(stem, title, inner, css, w=390, h=844, whole=False):
    """a phone board: inner = one fixed 390×844 screen, or several frames side by side (w wider); whole = a tall page drawn whole"""
    hs = f'min-height:{h}px' if whole else f'height:{h}px; overflow:hidden'
    root = f'<div class="ip-root" style="padding:0; width:{w}px; {hs}">{SX.sexify(inner)}</div>'
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
<style>{css}</style>
</helmet>
{root}
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":{w},"height":{h}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
    f = ROOT / f'design/canvas/project/{stem}.dc.html'
    if not f.exists() or f.read_text(encoding='utf-8') != dc: f.write_text(dc, encoding='utf-8')
    SYS = ROOT / 'design/system'
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{B.FONTS}<style>{css}</style></head><body>{root.replace(f"min-height:{h}px", "min-height:844px", 1) if whole else root}</body></html>'''
    pathlib.Path('/tmp/vo02').mkdir(exist_ok=True)
    pathlib.Path(f'/tmp/vo02/{stem}.html').write_text(wrap, encoding='utf-8')
    OUT.append((stem, title, w, h)); print('ok', stem, w, h)


def desk_doc(stem, title, body, overlay='', css='', h=1024):
    s4.write(stem, title, SX.sexify(body), SX.sexify(overlay), H=h, css=css)   # the S4 desktop writer (wide boards from board_widths.json)
    import canvas_rows as CR
    OUT.append((stem, title, CR.width(stem, 1440), h))


# ---------------------------------------------------------------------------------------------------------- shared parts
VO_CSS = '''
.vo-pair{display:flex;gap:var(--space-10);align-items:flex-start}
.vo-pair > .vo-frame{flex:none;width:390px;height:844px;position:relative;overflow:hidden}
.vo-strip{display:flex;flex-direction:column;gap:var(--space-6);padding:var(--space-6) 0}
.vo-strip__cell{display:flex;align-items:center;gap:var(--space-4)}
.vo-strip__cell > .ip-note{margin:0}
.vo-vp{position:absolute;inset:0;overflow:hidden}
.vo-popstack{position:absolute;top:calc(100% + var(--space-2));left:0;z-index:5;display:flex;flex-direction:column;gap:var(--space-6);width:var(--panel-w)}
.vo-popstack > .ip-voicepanel--pop{position:relative;top:auto;right:auto;transform-origin:top left} /* the page in its first 1024 px viewport (in the app it scrolls) */'''
STATUS_PREFILL = lambda k: f'<p class="ip-sr" role="status" aria-live="polite">Am completat {k} câmpuri din ce ai spus. Verifică și salvează.</p>'


def strip_pop(section):   # VO01's popover drawing carries a preview-only static style; on a page it is placed by §78
    a = ' style="position:static;margin:var(--space-2) 0 0 auto"'
    assert section.count(a) == 1
    return section.replace(a, '', 1)


def with_anchor(html, inner):
    """the page's header key (vkey) with its popover: the anchor holds the key in its state and the panel under it"""
    html = V.add(html)
    a = f'<span class="ip-voiceanchor">{V.VO.key("head", "rest", V.UID)}</span>'
    assert html.count(a) == 1
    return html.replace(a, f'<span class="ip-voiceanchor">{inner}</span>', 1)


def describe(html, ctl_id, extra):
    m = re.search(rf'<(input|textarea|select|div)\b[^>]*\bid="{re.escape(ctl_id)}"[^>]*>', html)
    tag = m.group(0)
    if 'aria-describedby="' in tag:
        new = re.sub(r'aria-describedby="([^"]*)"', lambda k: f'aria-describedby="{k.group(1)} {extra}"', tag, count=1)
    else:
        new = tag[:-1] + f' aria-describedby="{extra}">'
    return html[:m.start()] + new + html[m.end():]


# S2 Acasă · vet (Dr. Ceban, 3-item tab bar), as the S2 board
def s2_fixed(scrim=''):
    b = A.VET
    for a, z in (('<div class="ip-screen">', '<div class="ip-screen ip-screen--fixed ip-scrimhost">'),
                 ('<main class="ip-screen__main s2-main">', '<div class="ip-screen__view"><main class="ip-screen__main s2-main">'),
                 ('</main>', '</main></div>')):
        assert b.count(a) == 1, a; b = b.replace(a, z, 1)
    assert b.endswith('</div>')
    return b[:-len('</div>')] + (f'<div class="ip-scrim ip-scrim--sheet">{scrim}</div>' if scrim else '') + '</div>'


# S6 Zilnic at 08:20, before the day's saves: the form opens new (every field empty: empty is null, not 0)
OV3 = {('mortality', 'F'): 3, ('mortality', 'M'): 1}     # the spoken figures (sample, D39)
def s6_halls(prefill=False):
    card = (lambda s: S6.entry(s, ov=OV3) if s == 'H5.1' else S6.entry(s)) if prefill else None
    h, ALL, EG = S6.halls(card=card)
    keep = {'m0-H51', 'm1-H51'} if prefill else set()
    h = re.sub(r'(<input id="([\w-]+)"[^>]*?) value="[^"]*"', lambda m: m.group(0) if m.group(2) in keep else f'{m.group(1)} value=""', h)
    h = h.replace(' readonly', '')
    h = h.replace('<label class="ip-check is-checked">', '<label class="ip-check">').replace(' checked>', '>')
    h = h.replace(f'<span class="ip-check__box">{S6.icon("check", 14)}</span>', '<span class="ip-check__box"></span>')
    h = h.replace('<span class="ip-tag ip-tag--neutral">fără pierderi</span>', '')
    blank = (f'<dl class="ip-eggs__calc" aria-label="Calculat din câmpuri, doar citire"><div><dt>Total ouă</dt><dd class="ip-eggs__val">–</dd>'
             f'<dd class="ip-eggs__sub">= Cat. 1 + 2 + 3 + Incubabile · Jos separat</dd></div><div><dt>Ouat azi</dt><dd class="ip-eggs__val">–</dd>'
             f'<dd class="ip-eggs__sub">ghid <span class="ip-num">{n(S6.LAYG, 1)}%</span> · scrie ouăle</dd></div></dl>')
    h, k = re.subn(r'<dl class="ip-eggs__calc".*?</dl>', blank, h, flags=re.S); assert k == 2
    if prefill:
        for fid, sx in (('m0-H51', 'găini'), ('m1-H51', 'cocoși')):
            lab = f'id="l-{fid}"'
            i = h.index(lab); j = h.index('</label>', i)
            h = h[:j] + VO.mark(f'vm-{fid}') + h[j:]
            h = describe(h, fid, f'vm-{fid}')
        a = '<span class="ip-input ip-count ip-count--hen is-warn"><input id="m0-H51"'
        assert h.count(a) == 1; h = h.replace(a, a.replace('is-warn"', 'is-warn is-focus"'), 1)   # focus on the first filled field
    return h


def s6_save(prefill):
    rec = S6.rec14
    if prefill:
        l1 = f'{rec("mortality")}<span><span class="ip-num">1</span> secțiune ·</span><span>{B.sexv("F", "−3")} · {B.sexv("M", "−1")}</span>'
        l2 = (f'{rec("eggs")}<span><span class="ip-num">–</span> ouă ·</span>' + B.lamp('warn', 'abatere')
              + '<span>H5.1 <span class="ip-sex ip-sex--hen">♀</span> <span class="ip-sex ip-sex--rooster">♂</span></span>')
        key = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Salvează</button>'
    else:
        l1 = f'{rec("mortality")}<span><span class="ip-num">0</span> secțiuni ·</span><span>nimic scris</span>'
        l2 = f'{rec("eggs")}<span><span class="ip-num">–</span> ouă</span>'
        key = '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch" disabled aria-describedby="sb-sum">Salvează</button>'
    return B.savebar(l1, l2, key)


def s6_fixed(main_top='', prefill=False, scrim='', key=True):
    body = (f'<div class="ip-screen ip-screen--fixed{" ip-scrimhost" if scrim else ""}">\n{S6.HEAD}\n<div class="ip-screen__view"><main class="ip-screen__main">\n'
            f'{main_top}{S6.DATE}\n{s6_halls(prefill)}\n</main></div>\n{s6_save(prefill)}\n')
    if scrim:
        return body + f'<div class="ip-scrim ip-scrim--sheet">{scrim}</div></div>'
    return body + (VO.key('fab', 'rest', 'vk', float_=True) if key else '') + '</div>'


def prefill_banner(line2, touch=True, chip=''):
    return VO.banner('info', ['Am completat din ce ai spus. Verifică și salvează.', line2] + ([chip] if chip else []),
                     key_='Anulează completarea', role='', touch=touch)


# ---------------------------------------------------------------------------------------------------------- VO-1 … VO-13
CSS_S2 = A.CSS_P + VO_CSS
CSS_FORM = B.CSS + VO_CSS


def vo1():   # S2 Acasă · vet · idle: the whole page, the key above the 3-item tab bar, Introdus azi ending 88 above it
    h = A.HEIGHTS['S2-acasa']
    phone_doc('VO-1-acasa-repaus', 'VO-1 · Voce · S2 Acasă · vet · telefon · tasta în repaus, 88 px liberi sub Introdus azi',
              V.add(A.VET), CSS_S2, h=h, whole=True)


def vo2():   # listening: the sheet, Lamp Ascult…, meter, 0:04, live transcript, the key latched in the sheet foot
    pnl = VO.panel('v2', True, state=('ok', 'Ascult…'), time='0:04', level=8, hint=True, said=VO.SAY_P01, partial=True, key_st='listen')
    phone_doc('VO-2-acasa-asculta', 'VO-2 · Voce · S2 Acasă · vet · telefon · ascultă (transcriere live; comanda e exemplu, D39)',
              s2_fixed(pnl), CSS_S2)


def vo3():   # Zilnic prefilled from „la H5.1 au murit trei găini și un cocoș”
    bnr = prefill_banner('2 câmpuri: <span class="ip-num">H5.1</span> · ♀ Mortalitate <span class="ip-num">3</span> · ♂ Mortalitate <span class="ip-num">1</span>')
    body = s6_fixed(bnr + '\n', prefill=True)
    body = body[:-len('</div>')] + STATUS_PREFILL(2) + '</div>'
    phone_doc('VO-3-zilnic-precompletat', 'VO-3 · Voce · S6 Zilnic · vet · telefon · formular precompletat la 08:20 (♀ 3 · ♂ 1 spuse: exemplu, D39)',
              body, CSS_FORM)


def vo4():   # clarification: which compartment of H5
    pnl = VO.panel('v4', True, said='pune două găini moarte la hala cinci', sr='La care secțiune din H5?',
                   result=VO.chips('v4', 'La care secțiune din H5?', ['H5.1', 'H5.2'], touch=True, fill=True)
                   + f'<div class="ip-voicepanel__keys">{VO.btn("Spune din nou", "secondary", touch=True)}</div>')
    phone_doc('VO-4-zilnic-clarificare', 'VO-4 · Voce · S6 Zilnic · vet · telefon · clarificare „La care secțiune din H5?” (exemplu, D39)',
              s6_fixed(scrim=pnl), CSS_FORM)


def vo5():   # Hrană zilnică prefilled: 160 g on the 4 ♀ rows, the ♂ rows left as yesterday's suggestions
    rows = [dict(r, today_g=None) if r['sex'] == 'M' else r for r in B.ROWS]
    fem = [r for r in rows if r['sex'] == 'F']
    assert len(fem) == 4 and {r['today_g'] for r in fem} == {160} and {r['guide_g'] for r in fem} == {157}
    built = [B.row(r) for r in rows]
    kgF = R.rnd(sum(b[2] for b in built if b[1]), 1)
    assert kgF == R.feed_kg(TODAY, 'F') and n(kgF, 1) == '2\u202f303,2'
    pairs = []
    first = True
    for r, b in zip(rows, built):
        html = b[0]
        if r['sex'] == 'F':
            gid = f'{r["section"].replace(".", "-")}-F'
            lab = f'g/pasăre<span class="ip-sr">, {r["section"]} găini</span>'
            assert html.count(lab) == 1
            html = html.replace(lab, lab + VO.mark(f'vm-{gid}'), 1)
            html = describe(html, f'g-{gid}', f'vm-{gid}')
            if first:
                a = f'<span class="ip-input ip-input--touch ip-num"><input id="g-{gid}"'; assert html.count(a) == 1
                html = html.replace(a, a.replace('ip-num"', 'ip-num is-focus"'), 1); first = False
        pairs.append((r, html))
    halls = B.halls_html(pairs)
    secs = ' · '.join(f'<span class="ip-num">{r["section"]}</span>' for r in fem)
    bnr = prefill_banner(f'4 câmpuri: g/pasăre ♀ <span class="ip-num">160 g</span> · {secs}')
    save = B.savebar(B.sum_rows(4, len(rows), kgF), B.slamp, '<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">Salvează</button>')
    sc = HS.get('VO-5-hrana-zilnica-precompletat@scroll', 0)   # spec §8: scrolled to the first filled row (vo02_render --measure)
    st = f' style="margin-top:calc(-1 * {sc}px)"' if sc else ''
    body = (f'<div class="ip-screen ip-screen--fixed">\n{B.HEAD}\n<div class="ip-screen__view"><main class="ip-screen__main"{st}>\n{bnr}\n{B.DATE}\n{B.COPY}\n{halls}\n</main></div>\n'
            f'{save}\n{VO.key("fab", "rest", "vk", float_=True)}{STATUS_PREFILL(4)}</div>')
    phone_doc('VO-5-hrana-zilnica-precompletat', 'VO-5 · Voce · S7 Hrană zilnică · vet · telefon · precompletat la 08:20: „dă 160 de grame la toate găinile” (♀ 160 g, ghid 157 g)',
              body, CSS_FORM)


def vo6():   # nepermis: a past day for a vet
    pnl = VO.panel('v6', True, said='pune mortalitatea de ieri la ha șase punct unu',
                   result=VO.banner('info', ['Pentru zilele trecute cere managerului.'])
                   + '<div class="ip-chips ip-voicechips" role="group" aria-label="Unde poți merge">'
                     '<button type="button" class="ip-chip ip-voicechip ip-chip--touch">Deschide Introdus azi</button></div>')
    phone_doc('VO-6-nepermis', 'VO-6 · Voce · S2 Acasă · vet · telefon · nepermis: ziua de ieri (exemplu, D39)', s2_fixed(pnl), CSS_S2)


# S5 · Hîrbovăț › H5 (hall page): the S4 page parts scoped to H5
def s5_page(fb, table):
    crumbs = s4.CRUMBS.replace('<span aria-current="page">Hîrbovăț</span>',
                               '<a href="#hirbovat">Hîrbovăț</a><span aria-hidden="true">›</span><span aria-current="page">H5</span>')
    f, m = R.alive_hall('H5', 'F', TODAY), R.alive_hall('H5', 'M', TODAY)
    meta = (f'<span class="ip-tag ip-tag--neutral">Hală</span><span class="ip-tag ip-tag--neutral">Ross 308 ouat</span>{s4.pair(n(f), n(m))}'
            f'<span class="ip-note ip-num">secțiuni H5.1 · H5.2 · SAPT {R.WK} · ciclu din {s4.ddate(s4.C26_START)}</span>')
    acts = (f'<button type="button" class="ip-btn ip-btn--secondary">{s4.icon("transfer", 16)}Mută păsări</button>\n'
            f'<button type="button" class="ip-btn ip-btn--secondary">{s4.icon("sale", 16)}Vânzare</button>')
    head = f'<header class="ip-pagehead">\n<div>\n{crumbs}\n<h1>H5</h1>\n<div class="ip-row ip-pagehead__meta">{meta}</div>\n</div>\n<div class="ip-row">\n{acts}\n</div>\n</header>'
    tabs = T.tabs('pasari').replace('aria-label="Vederi Hîrbovăț"', 'aria-label="Vederi H5"')
    main = (f'<main class="ip-shell__main s4-main">\n{head}\n<div class="s4-tight">{fb}{tabs}</div>\n'
            f'<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-pasari">\n{table}\n</div>\n</main>')
    return f'<div class="ip-shell">{s4.sidebar("Manager")}{main}</div>'


def vo7():   # executat: „arată-mi mortalitatea la H5 pe ultimele 8 săptămâni”
    W8 = s4.W8
    assert W8[0] == 59 and W8[-1] == 52
    groups = T.pasari('H5')
    assert [NS.gname(g) for g in groups[:3]] == ['În viață', 'Mortalitate', '% cumulat']
    fb = s4.filterbar_weeks(8)
    types = s4.T8
    fb2 = fb.replace(s4.chips4(set(types), types), s4.chips4({'mortality'}, types)); assert fb2 != fb
    table = s4.ptable('pa', 'Păsări', [('Hala 5', '', groups[:3])], scope='H5', weeks=W8, period='8 săpt.',
                      after=s4.ptnote('Eticheta Mortalitate: În viață rămâne. % cumulat = morți din sosire ÷ păsări sosite în H5. '
                                      f'Golirea din {R.ddm(R.EMPTYING_SINCE)} e Vânzare; Mortalitate = doar morți reali.'))
    page = with_anchor(s5_page(fb2, table), VO.key('head', 'rest', 'vk'))
    sr = '<span class="ip-sr" role="status" aria-live="polite"></span>'
    toast = ('<div class="ip-toastwrap ip-toastwrap--page"><div class="ip-toast" role="status">' + VO.svg('mic', 16)
             + '<span class="ip-toast__txt">Am deschis Hîrbovăț › H5 · Păsări: Mortalitate, 8 săpt.</span>'
               '<button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
    desk_doc('VO-7-h5-pasari-executat', 'VO-7 · Voce · S5 Hîrbovăț › H5 · Păsări · manager · desktop · executat: Mortalitate, 8 săpt. (SAPT 52–59), Toast cu Anulează',
             f'<div class="vo-vp">{page}</div>', toast, css=T.CSS + VO_CSS)


def vo8():   # răspuns on S14: „care e procentul de ouat săptămâna asta față de ghid?”
    lay, gl = R.lay_week(R.WK), R.guide_lay(R.WK)
    a, b = R.week_dates(R.WK); days = (TODAY - a).days + 1
    diff = R.rnd(gl - lay, 1)
    assert (n(lay, 1), n(gl, 1), n(diff, 1), days) == ('40,4', '57,6', '17,2', 2)
    ans = (f'<p>Sapt. <span class="ip-num">{R.WK}</span> (<span class="ip-num">{days}</span> din <span class="ip-num">7</span> zile): ouat '
           f'<span class="ip-num">{n(lay, 1)}%</span>, ghid <span class="ip-num">{n(gl, 1)}%</span>, sub ghid cu <span class="ip-num">{n(diff, 1)}</span> puncte.</p>'
           + VO.lamp('warn', 'abatere') +
           '<p class="ip-voicepanel__opened">Am deschis Rapoarte · Producție<button type="button" class="ip-link">Anulează</button></p>')
    pnl = strip_pop(VO.panel('v8', False, said='care e procentul de ouat săptămâna asta față de ghid?', result=ans,
                             sr=f'Sapt. {R.WK} ({days} din 7 zile): ouat {n(lay, 1)}%, ghid {n(gl, 1)}%, sub ghid cu {n(diff, 1)} puncte. Abatere.'))
    page = with_anchor(X.page(), VO.key('head', 'open', 'vk') + pnl)
    desk_doc('VO-8-productie-raspuns', 'VO-8 · Voce · S14 Rapoarte · Producție · proprietar · desktop · răspuns: sapt. 59, ouat 40,4% față de ghid 57,6%',
             page, css=X.CSS, h=HS.get('VO-8-productie-raspuns', 2171))


def vo9():   # se gândește + transcript corectat: the pair on one board (right: before the tap; left: after it)
    before = strip_pop(VO.panel('v9a', False, state=('info', 'Mă gândesc…'), said=VO.SAY_P12, said_btn=True, cmd=VO.CMD_P12, key_st='think'))
    after = strip_pop(VO.panel('v9b', False, state=('off', 'oprit'), edit=VO.SAY_P12))
    # beside the docked SidePanel the PageHeader actions wrap under the title, the key on the page's left: a popover right-aligned
    # to it would run over the Sidebar, so here it opens left-aligned under the key (to review: §78 flip); the pair stacks
    pair = f'<div class="vo-popstack">{before}{after}</div>'
    page = with_anchor(Q.DESK, VO.key('head', 'think', 'vk') + pair)
    desk_doc('VO-9-livrare-se-gandeste', 'VO-9 · Voce · S13 Livrare hrană · manager · desktop · se gândește (sus) și, după apăsarea pe text, transcrierea gata de corectat (jos) · pereche de stări (D39)',
             page, Q.PANEL, css=Q.CSS_D + VO_CSS)


def vo10():  # Livrare hrană prefilled: Hîrbovăț, 13.11.2026, Breeder 2, 25 640 kg (the real 13.11 delivery)
    d13 = R.D('2026-11-13'); dl = next(x for x in R.DELIVERIES if x['date'] == '2026-11-13' and x['for_sex'] == 'F')
    assert dl['qty_kg'] == 25640 and dl['feed_type_id'] == Q.FT
    wk = R.sapt(d13)
    before = R.stock(d13 - dt.timedelta(days=1))['F']; after = before + dl['qty_kg']
    pnl = Q.PANEL
    # the date: 13.11.2026 typed (neither Azi nor Ieri), its age week
    a = '<span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Azi</span>'
    assert pnl.count(a) == 1
    pnl = pnl.replace(a, '<span class="ip-seg__item" role="radio" aria-checked="false" tabindex="0">Azi</span>', 1)
    for x, y in ((f'value="{dd(TODAY)}"', f'value="{dd(d13)}"'), (f'SAPT {R.WK} · ciclu din', f'SAPT {wk} · ciclu din')):
        assert pnl.count(x) >= 1; pnl = pnl.replace(x, y)
    pnl = pnl.replace(f'HÎRBOVĂȚ · {dd(d13)} · SAPT {wk}', f'HÎRBOVĂȚ · {dd(TODAY)} · SAPT {R.WK}')   # the panel head keeps today
    # kg filled, the stock line with it (stock at the end of 12.11 + the delivery)
    k = re.search(r'<span class="ip-input ip-num is-focus"><input id="p-kg" type="text" inputmode="decimal" value=""', pnl)
    pnl = pnl.replace(k.group(0), k.group(0).replace(' is-focus', '').replace('value=""', f'value="{n(dl["qty_kg"])}"'), 1)
    eff = re.search(r'<div class="ip-effect s13-effect" id="p-stock">.*?</div></div>', pnl, re.S).group(0)
    new = (f'<div class="ip-effect s13-effect" id="p-stock">{Q.rec("delivery", 16)}<div class="s13-effbody">'
           f'<span>Stoc {Q.FTN[Q.FT]} la Hîrbovăț la <span class="ip-num">{dd(d13 - dt.timedelta(days=1))}</span>: {Q.kgs(before)}</span>'
           f'<span>după livrare: <b class="s13-strong">{Q.kgs(after)}</b></span></div></div>')
    pnl = pnl.replace(eff, new, 1)
    # the keys: Adaugă enabled (every required field filled)
    a = '<span class="ip-note" id="lp-why">Scrie cantitatea în kg</span>'; assert pnl.count(a) == 1; pnl = pnl.replace(a, '<span></span>', 1)
    a = '<button type="button" class="ip-btn ip-btn--primary" disabled aria-describedby="lp-why">Adaugă</button>'; assert pnl.count(a) == 1
    pnl = pnl.replace(a, '<button type="button" class="ip-btn ip-btn--primary">Adaugă</button>', 1)
    # VoiceMarks on the 4 filled fields, focus on the first (Fermă)
    for lab, ctl, uid in (('<span class="ip-label" id="p-farm-l">Fermă</span>', 'p-farm', 'vm-farm'),
                          ('<label class="ip-label" for="p-date">Data</label>', 'p-date', 'vm-date'),
                          ('<label class="ip-label" for="p-type">Tip hrană</label>', 'p-type', 'vm-type'),
                          ('<label class="ip-label" for="p-kg">Cantitate</label>', 'p-kg', 'vm-kg')):
        assert pnl.count(lab) == 1, lab
        pnl = pnl.replace(lab, lab[:lab.rindex('</')] + VO.mark(uid) + lab[lab.rindex('</'):], 1)
        pnl = describe(pnl, ctl, uid)
    a = '<div class="ip-input ip-select__btn" id="p-farm"'; assert pnl.count(a) == 1
    pnl = pnl.replace(a, '<div class="ip-input ip-select__btn is-focus" id="p-farm"', 1)
    bnr = prefill_banner(f'4 câmpuri: Hîrbovăț · <span class="ip-num">{dd(d13)}</span> · {Q.FTN[Q.FT]} · <span class="ip-num ip-nowrap">{n(dl["qty_kg"])} kg</span>', touch=False)
    a = '<div class="ip-sidepanel__body ip-sidepanel__body--form">\n'; assert pnl.count(a) == 1
    pnl = pnl.replace(a, a + bnr + '\n', 1)
    desk_doc('VO-10-livrare-precompletat', 'VO-10 · Voce · S13 Livrare hrană · manager · desktop · formular precompletat: Hîrbovăț, 13.11.2026, Breeder 2, 25 640 kg',
             Q.DESK, pnl + STATUS_PREFILL(4), css=Q.CSS_D)


def vo11():  # fără microfon (S6, phone Chrome) + neînțeles (S2), two phones
    left = VO.panel('v11a', True, result=VO.banner('warn', ['Microfonul e blocat pentru Impex Line.',
                                                            'Apasă lacătul de lângă adresă › Permisiuni › Microfon › Permite.', 'Poți scrie comanda mai jos.']),
                    field_focus=True)
    neo = ('<div role="alert" class="ip-stack" style="gap:var(--space-1)"><p>Nu am înțeles „pune trei la hala”.</p>'
           '<p>Încearcă așa: „pune 3 găini moarte la H5.1” sau „ce lipsește azi?”.</p></div>')
    right = VO.panel('v11b', True, said='pune trei la hala', result=neo + '<div class="ip-voicepanel__keys">'
                     + VO.btn('Spune din nou', 'primary', touch=True) + VO.btn('Scrie comanda', touch=True) + '</div>')
    inner = f'<div class="vo-pair"><div class="vo-frame">{s6_fixed(scrim=left)}</div><div class="vo-frame">{s2_fixed(right)}</div></div>'
    phone_doc('VO-11-fara-microfon-neinteles', 'VO-11 · Voce · telefon · fără microfon (stânga, S6 Zilnic, Chrome) · neînțeles (dreapta, S2 Acasă) · exemple de stare (D39)',
              inner, CSS_FORM + '\n' + A.CSS_P, w=390 * 2 + 40)


def vo12():  # fără conexiune + the key's states 1:1
    pnl = VO.panel('v12', True, result=VO.banner('fault', ['Comanda nu s-a trimis: „pune 3 găini moarte la H5.1”.',
                                                           'Verifică conexiunea la internet, apoi apasă Reîncearcă. Ce ai spus rămâne scris mai jos.'],
                                                 key_='Reîncearcă', touch=True, focus_key=True), field='pune 3 găini moarte la H5.1')
    sts = [('rest', 'Repaus'), ('listen', 'Ascultă · aria-pressed'), ('think', 'Se gândește · aria-busy'), ('off', 'Dezactivat · „Fără conexiune”')]
    strip = '<div class="vo-strip">' + ''.join(f'<div class="vo-strip__cell">{VO.key("fab", s, f"v12{s}")}<p class="ip-note">{w}</p></div>' for s, w in sts) + '</div>'
    inner = f'<div class="vo-pair"><div class="vo-frame">{s2_fixed(pnl)}</div>{strip}</div>'
    phone_doc('VO-12-fara-conexiune', 'VO-12 · Voce · S2 Acasă · vet · telefon · fără conexiune (exemplu, D39) · stările tastei 1:1',
              inner, CSS_S2, w=390 + 40 + 280)


def vo13():  # S4 phone Producție after „deschide Hîrbovăț la Producție”: Toast space-2 above the key, the table clear of both
    pf = NS.p_filter().replace(f'Ciclul curent · Zi · {s4.CHIPW["mortality"]}', f'Ciclul curent · Zi · {s4.CHIPW["eggs"]}')
    assert s4.CHIPW['eggs'] in pf
    tb = X.table(phone=True)
    panel = f'<div class="s4-panel" id="tp-ferma" role="tabpanel" aria-labelledby="tab-productie">\n{tb}\n</div>'
    # D195 (SYS20): the selected "Producție" with its dot runs 2 px past its 71.6 px fifth of 358: the short word, full word hidden
    tabs = T.tabs('productie'); a = 'href="#productie">Producție</a>'; assert tabs.count(a) == 1
    tabs = tabs.replace(a, 'href="#productie"><span aria-hidden="true">Prod.</span><span class="ip-sr">Producție</span></a>', 1)
    body = NS.phone([NS.p_header(), NS.p_cycle(), f'<div class="s4-tight">{pf}{tabs}</div>', panel])
    # the whole page (as the S4 phone boards): header, Producție tab, the table; the Toast is placed in the main, which ends at
    # the tab bar, so it stacks space-2 above the floating key (§78) and the table's last row ends clear of both
    a = '<main class="ip-screen__main s2-main s4-main">'; assert body.count(a) == 1
    body = body.replace(a, '<main class="ip-screen__main s2-main s4-main" style="position:relative">', 1)
    toast = ('<div class="ip-toastwrap"><div class="ip-toast" role="status">' + VO.svg('mic', 16)
             + '<span class="ip-toast__txt">Am deschis Hîrbovăț · Producție</span><button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
    a = '</main>'; assert body.count(a) == 1; body = body.replace(a, toast + '</main>', 1)
    body = V.add(body)
    phone_doc('VO-13-ferma-productie-toast', 'VO-13 · Voce · S4 Hîrbovăț · Producție · manager · telefon · executat: pagina întreagă, Toast space-2 deasupra tastei, ultimul rând al tabelului liber de amândouă',
              body, X.CSS + '\n' + NS.CSS_PHONE + VO_CSS, h=HS.get('VO-13-ferma-productie-toast', 3400), whole=True)


for f in (vo1, vo2, vo3, vo4, vo5, vo6, vo7, vo8, vo9, vo10, vo11, vo12, vo13):
    f()
json.dump([{'stem': s, 'w': w, 'h': h} for s, _, w, h in OUT], open('/tmp/vo02/boards.json', 'w'))

if CANVAS:
    import canvas_rows as CR
    c = CR.load(ROOT); Bd = c['boards']
    v1 = Bd['V1-voce.dc.html']; y = v1['y'] + v1['h'] + CR.GAP_Y
    rowh = max(h for _, _, _, h in OUT)
    mine = {f'{s}.dc.html' for s, _, _, _ in OUT}
    have = [k for k in mine if k in Bd]
    if not have:   # a new row under V1: every later row moves down by the row + 120
        for k, v in Bd.items():
            if v['y'] >= y: v['y'] += rowh + CR.GAP_Y
    x = 0
    for s, t, w, h in OUT:
        e = {'x': x, 'y': y, 'w': w, 'h': h}
        if w >= 1440: e['expand'] = 'fill'
        e['title'] = t
        Bd[f'{s}.dc.html'] = e
        if f'{s}.dc.html' not in c['order']: c['order'].append(f'{s}.dc.html')
        x += w + CR.GAP_X
    for mv in CR.relayout(c): print('relayout', *mv)
    CR.sort_order(c); CR.save(ROOT, c)
    print('canvas: VO row y', y, 'h', rowh)
