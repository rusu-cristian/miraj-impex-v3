"""S2S · Acasă state boards (spec D2, C8, C9; design/notes/S2.md state list), built on s2_build.py's blocks.
python3 -I s2s_build.py <repo_root> [--canvas]
Imports s2_build (it rewrites the 4 S2 default boards and, through s7s_build, the S7 boards byte-identical) and reuses
its blocks (missing, entrykeys, introdus, alerts, stock, unpriced, track, farms_block), shells (phone tab bar, Sidebar)
and CSS; phone viewport / Banner / modal anchor CSS from sstate_css.py; empty-box classes from s4s_build. S2P: phone Dialog,
tab bar key off, StockReadout and CycleTrack skeletons are system classes (additions.css section 35).
Phone state boards are 390×844 viewports (D60) with the BottomTabBar at the foot; desktop boards are pages (1440).
Every figure comes from the extracts; boards whose example is not in the sample say "exemplu de stare (D39)".
Writes design/canvas/project/S2-acasa--<state>.dc.html, /tmp/s2/<stem>.html (phones), /tmp/s4s/<stem>.html (desktops),
/tmp/s2/states.json (for s2_render.py); --canvas sets every S2 board in canvas.json: row y 20922, the 4 default boards
first, then the state boards, phones before desktops (D87), 80 px gaps, then the 2 Observator boards (S2V, D89).
S2V · Acasă for the viewer (D89): the owner layout without Livrări fără preț (no prices, spec §1 roles), no EntryKeys, no
Introdus azi, no Înregistrează: CycleTrack · De verificat (the lipsă rows read-only: their link opens the section page) ·
Stoc hrană (kg) · Ferme. Phone = a whole tall page (as the default phones), tab bar Acasă · Serie · Rapoarte · Mai mult;
desktop Sidebar without Înregistrează and Setări (every Setări item is manager+ or owner). Writes /tmp/s2/viewer.json."""
import sys, json, pathlib, importlib, datetime as dt

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve()
HERE = pathlib.Path(__file__).resolve().parent
HP = HERE / 's2_heights.json'
sys.path.insert(0, str(HERE))
sys.argv = ['s2_build.py', str(ROOT), str(HP)]
A = importlib.import_module('s2_build')        # writes the S2 default boards (and S7, unchanged)
import sstate_css as SC
s4, B = A.s4, A.B
n, dd, num, lamp, icon, tile, NB, D = A.n, A.dd, A.num, A.lamp, A.icon, A.tile, A.NB, A.D
TODAY, S, WK = A.TODAY, A.S, A.WK
A.IC.update({'fault': '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>'})  # Banner fault (system preview)
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}
VH = 844

# ---------- checks on the data the states rest on ----------
D7 = [(TODAY - dt.timedelta(days=k)) for k in range(7, 0, -1)]            # 09.11 – 15.11
DAYS = {r['date'] for r in S['daily']}
assert all(d.isoformat() in DAYS for d in D7)
assert all(r['g_per_bird'] is not None and r['mortality'] is not None for r in S['daily'] if D7[0].isoformat() <= r['date'] < TODAY.isoformat())
AM = '10:40'  # before 12:00: every save of the vet's day is in (last at 10:10), today's missing rows are not listed yet (C9)
assert max(e['at'][11:16] for e in A.LOG.values()) < AM < '12:00'
PAST = ('2026-11-15', 'H4.2', 'F')  # D39: a past day missing (the sample has 09.11–15.11 complete); real counts of that day
pr, py = A.day_row(*PAST); assert (pr['alive_end'], py['g_per_bird']) == (4110, 156) and pr['g_per_bird'] is not None
TL = A.TL; T_END = dd(dt.date.fromisoformat(A.cT['end_date'])); assert TL['totals']['all'] == 0 and T_END == '28.02.2026'
B2 = A.B2; NEG = -1240  # D39 / D40: the kit's C2 "stoc sub 0" figure (spec example), drawn on Breeder 2 as the SYS04 preview
assert B2['avg_daily_kg_last3'] == 3976

# ---------- phone pieces ----------
def pscreen(parts, vet=True, overlay='', modal=False, end=False, key_off=None):
    """end: the viewport shows the foot of the page (scrolled to the bottom, as S6S "salvat"; sstate_css .s7s-view--end).
    key_off: id of the reason; the tab bar's Înregistrează key is aria-disabled with it (as the EntryKeys on that board)."""
    inner = f'<div class="s7s-view{" s7s-view--end" if end else ""}"><main class="s7-main s2-main">\n{"".join(parts)}\n</main></div>\n{A.tabbar(vet, key_off)}'
    if modal:  # the page under the Dialog is inert; the scrim covers the whole screen, tab bar included
        inner = f'<div class="s2s-under" inert="inert" aria-hidden="true">{inner}</div>'
    return f'<div class="s7-screen s7s-vp s2-phone{" s9-modal" if modal else ""}">\n{inner}\n{overlay}\n</div>'
PH = lambda ctx: A.pagehead(ctx, True)

def skrow(): return ('<div class="ip-ev ip-ev--skel" aria-hidden="true"><span class="ip-skel ip-skel--tile"></span><span class="ip-ev__main">'
                     '<span class="ip-skel ip-skel--w1"></span><span class="ip-skel ip-skel--w2"></span></span><span class="ip-skel ip-skel--w3"></span></div>')
def sklist(k): return f'<div class="ip-list">{skrow() * k}</div>'
def skmissing(scope):
    return (f'<section class="ip-missing" aria-labelledby="ms" aria-busy="true"><div class="ip-missing__head"><h2 class="ip-cap" id="ms">Lipsesc date</h2>'
            f'<p class="ip-note">{scope} · azi de la {num("12:00")} · ultimele {num(7)} zile</p></div>{sklist(2)}</section>')
def skblock(cap, note, hid, body):
    return f'<section class="s2-block" aria-labelledby="{hid}" aria-busy="true">{A.head(cap, note, hid)}{body}</section>'
STATUS = lambda what: f'<p class="ip-sr" role="status">Se încarcă {what} pentru <span class="ip-num">{D}</span>…</p>'

def ek_off(note_html, u='x'):
    """EntryKeys disabled (SYS08): links without href, aria-disabled, the one reason under the grid."""
    li = ''.join(f'<li><a class="ip-btn ip-btn--secondary ip-entrykey" role="link" aria-disabled="true" aria-describedby="{u}ek-n">{tile(k)}'
                 f'<span class="ip-entrykey__txt"><span>{lb}</span><span class="ip-note">{hint}</span></span></a></li>' for k, lb, hint, _ in A.KEYS[:3])
    return (f'<section class="ip-entrykeys" aria-labelledby="{u}ek"><h2 class="ip-cap" id="{u}ek">Înregistrează</h2><ul class="ip-entrykeys__list">{li}</ul>'
            f'<p class="ip-note" id="{u}ek-n">{note_html}</p></section>')

boards = []  # (slug, state title, body, kind) kind: 'p' phone / 'm' manager desktop / 'o' owner desktop

# 1 · vet · se încarcă — header and keys are known from the route; the blocks are static skeleton rows (no shimmer)
boards.append(('se-incarca', 'vet · se încarcă', 'p', pscreen([
    PH(A.CTX_VET), STATUS('Acasă'), skmissing('Hîrbovăț'), A.entrykeys(),
    skblock('Introdus azi', f'{num(D)} · poți modifica sau șterge până la {num("23:59")}', 'ia', sklist(3))])))

# 2 · vet · înainte de 12:00 (10:40, real sample): Lipsesc date empty, before-12:00 sentence; Acțiuni menu open on the
# Hrană zilnică row (the first step to Șterge, board 4)
E_AM = (f'Zilele <span class="ip-num">{dd(D7[0])}–{dd(D7[-1])}</span> sunt complete. Lipsurile de azi apar de la {num("12:00")}.')
boards.append(('inainte-de-12', 'vet · înainte de 12:00 · Acțiuni deschis', 'p', pscreen([
    PH(A.CTX_VET), A.missing(when=AM, empty=E_AM), A.entrykeys(), A.introdus(menu={'feed': 'open'})])))

# 3 · vet · zi trecută lipsă, blocat (D07) — before 12:00, past-day rows only (D39: 15.11 H4.2 ♀ missing)
boards.append(('zi-trecuta-blocat', 'vet · zi trecută lipsă · blocat · exemplu de stare (D39)', 'p', pscreen([
    PH(A.CTX_VET), A.missing(miss=[], past=[PAST], locked=True, when=AM), A.entrykeys(), A.introdus()])))

# 4 · vet · Șterge · confirmare — Dialog over the default page (after 12:00); no reason for the author's same-day delete
# (D70 A); the whole save goes (D86); focus starts on Anulează
dlg = (f'<div class="ip-scrim ip-scrim--phone"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-n">'
       f'<h2 id="dlg-t">Șterge Hrană zilnică din <span class="ip-num">{D}</span></h2>'
       f'<p class="ip-dialog__note" id="dlg-n">Se șterg toate cele {num(len(A.FEED_ROWS))} rânduri salvate la {num(A.t_of(A.LOG["feed"]))} '
       f'({num(n(A.FEED_KG, 1) + NB + "kg")}). Secțiunile lor apar din nou în Lipsesc date. Ștergerea rămâne în Istoric modificări.</p>'
       f'<div class="ip-dialog__keys ip-dialog__keys--stack"><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-focus" autofocus="autofocus">Anulează</button>'
       f'<button type="button" class="ip-btn ip-btn--destructive ip-btn--touch">Șterge</button></div></div></div>')
boards.append(('sterge-confirmare', 'vet · Șterge · confirmare', 'p', pscreen(
    [PH(A.CTX_VET), A.missing(), A.entrykeys(), A.introdus()], overlay=dlg, modal=True)))

# 5 · vet · după 23:59 — the page of 16.11 still open at 00:05: keys aria-disabled, menu open with the reason;
# the Mortalitate row edited by its author the same day (D39: no edit in the sample) shows "modificat la 09:52";
# the menu opens on the Ouă row (it covers neither that mark nor the tab bar), the view scrolled to the foot of the page
boards.append(('dupa-2359', 'vet · după 23:59 · modificat la · exemplu de stare (D39)', 'p', pscreen([
    PH(A.CTX_VET), A.missing(), A.entrykeys(), A.introdus(menu={'eggs': 'open'}, off=True, edited={'mortality': '09:52'})], end=True)))

# 6 · vet · fermă goală — a vet at Telița (D39: the assignment), farm empty since 28.02.2026 (real): nothing to enter
boards.append(('ferma-goala', 'vet · fermă goală (Telița) · exemplu de stare (D39)', 'p', pscreen([
    PH(f'TELIȚA · {D}'),
    A.missing(miss=[], scope='Telița', empty=f'Telița e goală din {num(T_END)}. Aici apar doar secțiunile cu păsări.'),
    ek_off(f'Telița e goală din {num(T_END)}. Intrările se deschid cu ciclul nou.'),
    A.introdus_empty('Nu ai salvat nimic azi.')], key_off='xek-n')))

# 7 · vet · fără acces — no farm assigned: one EmptyState, no blocks (they all act on a farm)
box = s4.empty_box('fără acces', ['Nu ai acces la nicio fermă.', 'Cere proprietarului să te adauge la ferma la care lucrezi.'], '', 'Acces')
box = box.replace('<p class="s4s-empty-p">Nu ai acces', '<p class="s4s-empty-p" id="fa-n">Nu ai acces', 1); assert 'id="fa-n"' in box
boards.append(('fara-acces', 'vet · fără acces', 'p', pscreen([PH(D), box], key_off='fa-n')))  # the tab bar key reads the EmptyState reason

# 8 · vet · eroare la încărcare — online only (spec §1): fault Banner naming what failed + Reîncearcă; keys stay
ban = (f'<div class="ip-banner ip-banner--fault s7s-banner" role="alert">{icon("fault", 16)}<div class="s7s-banner__body">'
       f'<p class="s7-m0">Lipsesc date și Introdus azi nu s-au încărcat. Verifică conexiunea și reîncearcă.</p>'
       f'<button type="button" class="ip-btn ip-btn--secondary ip-btn--touch">Reîncearcă</button></div></div>')
boards.append(('eroare-la-incarcare', 'vet · eroare la încărcare', 'p', pscreen([PH(A.CTX_VET), ban, A.entrykeys()])))

# ---------- desktop pieces ----------
sk = s4.sk
def skreadouts(k, hid):
    ro = f'<li><div class="ip-readouts__skel" aria-hidden="true">{sk() * 5}</div></li>'
    return f'<ul class="ip-readouts" aria-labelledby="{hid}">{ro * k}</ul>'

# 9 · manager · desktop · se încarcă
boards.append(('manager-desktop--se-incarca', 'manager · desktop · se încarcă', 'm', A.desktop('Manager', [
    A.pagehead(A.CTX_MGR), STATUS('Acasă'),
    f'<div class="s2-cols">{skblock("De verificat", f"fermele tale · {num(D)}", "dv", sklist(3))}{skmissing("fermele tale")}</div>',
    skblock('Zile de hrană', f'stoc și zile rămase · {num(D)}', 'st', skreadouts(3, 'st')),
    A.entrykeys(True), skblock('Introdus azi', num(D), 'ia', sklist(2))])))

# 10 · manager · desktop · zi trecută lipsă (link, D07: managers enter past days) + stoc sub 0 (D50) — D39 both
a_neg = s4.alert('delivery', f'Breeder 2 · stoc <span class="ip-num">{n(NEG)} kg</span> · sub 0',
                 f'{D} · consumul depășește livrările', lamp('warn', 'abatere'), ('Deschide stocul', '#hrana-stoc'))
al = A.alerts()
first = '<div class="ip-ev ip-ev--alert">'
assert al.count(first) == 3 and f'{num(3)} alerte' in al
al = al.replace(f'{num(3)} alerte', f'{num(4)} alerte')
i3 = al.rindex(first)  # the stock alert (Cocoși) is the last; the Breeder 2 one goes just before it
al = al[:i3] + a_neg.replace('<div class="ip-ev__main">', A.FARM_P, 1) + al[i3:]
st = A.stock()
ro_b2 = st[st.index('<li><div class="ip-readout ip-readout--stock" role="group" aria-labelledby="sr-ft-breeder2">'):st.index('</li>') + 5]
w0, w1 = B2['avg_window']
ro_neg = (f'<li><div class="ip-readout ip-readout--stock" role="group" aria-labelledby="sr-ft-breeder2"><div class="ip-readout__cap"><span id="sr-ft-breeder2">Hîrbovăț · Breeder 2'
          f'<span class="ip-sr">, pentru găini</span></span><span class="ip-rlamp ip-rlamp--warn">abatere</span></div>'
          f'<div class="ip-readout__val">{n(NEG)}<small> kg</small></div><dl class="ip-readout__kv"><div><dt>zile rămase</dt><dd>{n(0, 1)}<small> zile</small></dd></div>'
          f'<div><dt>medie/zi {w0[8:10]}–{dd(dt.date.fromisoformat(w1))[:5]}</dt><dd>{n(B2["avg_daily_kg_last3"])}<small> kg</small></dd></div>'
          f'<div><dt>epuizat pe</dt><dd>{D}</dd></div></dl><p class="ip-readout__sub s2-m0">stoc sub 0 · verifică livrările</p></div></li>')
assert st.count(ro_b2) == 1
boards.append(('manager-desktop--zi-trecuta-stoc-sub-0', 'manager · desktop · zi trecută lipsă · stoc sub 0 · exemplu de stare (D39)', 'm', A.desktop('Manager', [
    A.pagehead(A.CTX_MGR), f'<div class="s2-cols">{al}{A.missing(True, past=[PAST])}</div>', st.replace(ro_b2, ro_neg),
    A.entrykeys(True), A.introdus_empty(A.INTRO_MGR)])))

# 11 · proprietar · desktop · se încarcă — track, alerts, UnpricedCount (its own skeleton), stock, farm summaries
sktrack = f'<div class="ip-trackskel" aria-hidden="true">{sk() * 3}</div>'  # 2 lanes + the month axis
skup = (f'<section class="ip-unpriced" aria-labelledby="up" aria-busy="true"><div class="ip-unpriced__head">{tile("price")}<h2 class="ip-cap" id="up">Livrări fără preț</h2></div>'
        f'<div aria-hidden="true" class="ip-unpriced__skel"><span class="ip-skel ip-skel--val"></span><span class="ip-skel ip-skel--w1"></span><span class="ip-skel ip-skel--w2"></span></div></section>')
skfs = f'<article class="ip-fsum ip-fsum--compact" aria-hidden="true">{sk("s4s-w40 s4s-h20")}{sk("s4s-w50")}{sk("s4s-w30 s4s-h20")}{sk("s4s-w100")}{sk("s4s-w80")}</article>'
boards.append(('proprietar-desktop--se-incarca', 'proprietar · desktop · se încarcă', 'o', A.desktop('Proprietar', [
    A.pagehead(A.CTX_OWN), STATUS('Acasă'),
    skblock('Seria 2026 · cicluri pe ferme', f'{num(2)} ferme', 'tk', sktrack),
    f'<div class="s2-cols">{skblock("De verificat", f"toate fermele · {num(D)}", "dv", sklist(4))}{skup}</div>',
    skblock('Stoc hrană', f'stoc și zile rămase · {num(D)}', 'st', skreadouts(3, 'st')),
    skblock('Ferme', f'{num(2)} ferme · {num(D)}', 'fm', f'<div class="ip-fsums s2-fsums">{skfs * 2}</div>')])))

# 12 · proprietar · desktop · toate livrările au preț (UnpricedCount "none", D39: the sample has 4 unpriced)
up0 = (f'<section class="ip-unpriced" aria-labelledby="up"><div class="ip-unpriced__head">{tile("price")}<h2 class="ip-cap" id="up">Livrări fără preț</h2></div>'
       f'<p class="ip-unpriced__fig"><span class="ip-unpriced__val">0<small> livrări</small></span></p><p class="ip-note">Toate livrările au preț.</p>'
       f'<a class="ip-btn ip-btn--secondary" href="#rapoarte-hrana-livrari">Vezi livrările</a></section>')
boards.append(('proprietar-desktop--toate-cu-pret', 'proprietar · desktop · toate livrările au preț · exemplu de stare (D39)', 'o', A.desktop('Proprietar', [
    A.pagehead(A.CTX_OWN), A.track(), f'<div class="s2-cols">{A.alerts(True)}{up0}</div>', A.stock(cap='Stoc hrană'), A.farms_block()])))

# ---------- S2V · Observator (D89) ----------
def v_alerts(u=''):
    al = A.alerts(True, u)  # the owner's 5 alerts: the lipsă rows link to the section page (Deschide secțiunea), never to a form
    assert al.count('toate fermele') == 1 and 'Completează' not in al and 'preț' not in al and al.count('lipsă</span>') == 2
    return al.replace('toate fermele', 'fermele tale')  # the viewer sees his assigned farms (spec §1); here both
def v_sidebar():
    sb = A.sidebar('Observator')
    for h in ('#inregistreaza', '#setari'):
        sb, k = __import__('re').subn(r'<a class="ip-nav" href="' + h + r'">.*?</a>', '', sb, count=1); assert k == 1
    assert sb.count('class="ip-nav') == 3 and '<b>Observator</b>' in sb
    return sb
V_PARTS = lambda phone: [A.pagehead(A.CTX_OWN, phone), A.track(phone), v_alerts(), A.stock(cap='Stoc hrană'), A.farms_block()]
VIEWER_P = A.phone(V_PARTS(True), viewer=True)
VIEWER_D = f'<div class="s4-shell">{v_sidebar()}<main class="s4-main">\n{"".join(V_PARTS(False))}\n</main></div>'
for v in (VIEWER_P, VIEWER_D):
    assert not any(w in v for w in ('Înregistrează', 'Introdus azi', 'preț', 'lei', 'Completează', 'ip-entrykey', 'ip-unpriced', 'Dr. Ceban ·'))
CSS_V = A.CSS_P + '''
.s2-track .month,.s2-track .todaylbl,.s2-track .rotxt{font-size:var(--ip-fs-caption)} /* track text >= 11px, as the desktop boards */
.s2-phone .ip-fsum__name a{display:inline-block;padding:var(--space-3) var(--space-1);margin:calc(-1 * var(--space-3)) calc(-1 * var(--space-1))} /* touch: 48px hit area, line unchanged */
.s2-phone .ip-fsum__empty .ip-link{display:inline-flex;align-items:center;min-height:var(--control-h-touch)}'''

# ---------- CSS ----------
EMPTY_CSS = '\n'.join(l for l in s4.CSS_STATES.splitlines() if l.startswith(('.s4s-emptybox', '.s4s-empty-p', '.ip-empty > .ip-btn')))
CSS_PS = A.CSS_P + SC.VIEW + SC.BANNER + SC.MODAL + '\n' + EMPTY_CSS + '''
.s2s-under{display:flex;flex-direction:column;flex:1;min-height:0} /* the inert page under the Dialog keeps the screen's column */'''
CSS_DS = A.CSS_D + '\n' + s4.CSS_STATES

# ---------- write ----------
STEM0, TITLE0 = A.STEM0, A.TITLE0
SYS = ROOT / 'design/system'
out = []
for slug, state, kind, body in boards:
    stem, title = f'{STEM0}--{slug}', f'{TITLE0} · {state}'
    if kind == 'p':
        css = CSS_PS
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
{B.FONTS}<style>{css}</style></head><body>{root}</body></html>'''
        pathlib.Path('/tmp/s2').mkdir(exist_ok=True)
        pathlib.Path(f'/tmp/s2/{stem}.html').write_text(wrap, encoding='utf-8')
        print('ok', stem, VH)
        out.append((stem, title, VH, 390))
    else:
        h = HEIGHTS.get(stem, 1024)
        s4.write(stem, title, body, H=h, css=CSS_DS)
        out.append((stem, title, h, 1440))
V_OUT = [A.write_phone(f'{STEM0}--observator', f'{TITLE0} · observator · telefon', VIEWER_P, css=CSS_V)]
h = HEIGHTS.get(f'{STEM0}--observator-desktop', 1024)
s4.write(f'{STEM0}--observator-desktop', f'{TITLE0} · observator · desktop', VIEWER_D, H=h, css=A.CSS_D)
V_OUT.append((f'{STEM0}--observator-desktop', f'{TITLE0} · observator · desktop', h, 1440))
pathlib.Path('/tmp/s2/viewer.json').write_text(json.dumps([{'stem': s, 'w': w, 'h': h} for s, _, h, w in V_OUT], ensure_ascii=False))
pathlib.Path('/tmp/s2/states.json').write_text(json.dumps([{'stem': s, 'w': w, 'h': h} for s, _, h, w in out], ensure_ascii=False))

if CANVAS:  # S2 row: y 20922; 4 default boards (phones first, D74), then states: phones, then desktops (D87); 80 px gaps
    cp = ROOT / 'design/canvas/project/canvas.json'
    c = json.loads(cp.read_text(encoding='utf-8'))
    Y = 20922
    defaults = [(STEM0, f'{TITLE0} · vet · telefon', 390), (STEM0 + '--manager', f'{TITLE0} · manager · telefon', 390),
                (STEM0 + '--manager-desktop', f'{TITLE0} · manager · desktop', 1440), (STEM0 + '--proprietar-desktop', f'{TITLE0} · proprietar · desktop', 1440)]
    for s, _, _ in defaults: assert c['boards'][f'{s}.dc.html']['y'] == Y
    seq = [(s, t, HEIGHTS.get(s, 844 if w == 390 else 1024), w) for s, t, w in defaults] + \
          [b for b in out if b[3] == 390] + [b for b in out if b[3] == 1440] + V_OUT  # S2V: viewer phone, then desktop
    x = 0
    for s, t, h, w in seq:
        b = {'x': x, 'y': Y, 'w': w, 'h': h, 'title': t}
        if w == 1440: b = {'x': x, 'y': Y, 'w': w, 'h': h, 'expand': 'fill', 'title': t}
        c['boards'][f'{s}.dc.html'] = b; x += w + 80
        if f'{s}.dc.html' not in c['order']: c['order'].append(f'{s}.dc.html')
    cp.write_text(json.dumps(c, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    print('canvas S2 row y', Y, 'boards', len(seq), 'ends at x', x - 80)
