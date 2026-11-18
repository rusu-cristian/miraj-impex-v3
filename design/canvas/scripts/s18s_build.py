"""S18S · Serii · state boards (spec B3; D107). Imports s18_build (which rewrites the 5 S18 boards, and through s17_build
the S17/S14S/S2/S7 boards, byte-identical) and draws the S18 states on the same row (y 32813). Desktops first (a Setări
page, as S17): the new desktops follow the two S18 desktops; the S18 phones move right (x only, as S16S/S17S); the new
phones follow them. Today = 16.11.2026 (D23/D24) unless the board is a replay of the real creation (07.10.2025, D73/D83).
Boards
  --se-incarca                  header kept, CycleTrack skeleton (§54 ip-trackskel, track height) + framed table skeleton, hidden status
  --eroare-la-incarcare         header kept (Adaugă serie stays), fault Banner + Reîncearcă in place of track and table
  --fara-acces                  Observator Sidebar, title only, EmptyState "Nu ai acces la Serii." + "Înapoi la Acasă"
                                (a manager or a vet: the same picture with their own nav and role sentence; not drawn, one board)
  --desktop-scris               replay 07.10.2025: the list as it was before the creation (Seria 2025 only, Hîrbovăț open
                                SAPT 31) beside the SidePanel with the real creation typed (H1 ♀ 14 080 · H2 ♀ 14 000 ·
                                H5-1 ♂ 3 510), Creează seria live
  --desktop-erori               16.11.2026, D39 values: Nume lipsă (focus), An 2026 taken by Seria 2026, Data 17.11.2026 in
                                the future, two rows H1 with no count (H1 de două ori + niciun rând completat): one form
  --desktop-refuzat             16.11.2026, D39 (as the phone refusal): Telița still holds the 29 825 that arrived at
                                Hîrbovăț, so the list beside it is drawn on the same premise (the move not entered: Telița
                                open SAPT 58, no Hîrbovăț cycle in Seria 2026); fault Banner, Creează seria disabled
  --salvat                      replay 07.10.2025 after Creează seria: panel closed, Seria 2026 in the list (Telița open,
                                SAPT 1, populare), Toast in B4 words + Anulează, focus back on Adaugă serie
  --inchide-seria               16.11.2026, D39: Seria 2025 as if not closed yet (its last cycle ended 28.01.2026), its
                                Închide seria pressed, Dialog "Închide Seria 2025" over the inert page, focus on Anulează
  --telefon-eroare-la-salvare   replay 07.10.2025: the typed sheet kept, fault Banner + Reîncearcă right above the
                                EffectPreview (S9P), Creează seria enabled
  --telefon-alege-sectiunea     replay 07.10.2025: H1 and H2 filled, a third row's SectionSelect open as the §33 sheet:
                                H1/H2 "în listă" disabled, H3 … H5-2 "goală" enabled, H3 active (over the inert form)
No new system class: the skeletons are additions.css §54 `ip-trackskel` (full, 120 as the drawn track; FIN11G) and §42 `tr.ip-table__skel` under the §54 column shares (FIN11; were `.s18s-trackskel`/`.s18s-skt`); screen CSS: one panel-body rule (the Dialog scrim is additions.css §50, FIN04; the Toast is additions.css §48, FIN02).
python3 -I s18s_build.py <repo_root> [--canvas]   (heights from s18_heights.json; renders: s18_render.py <stems>)"""
import sys, re, json, pathlib, importlib.util, datetime as dt

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve(); HERE = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
import skel as SKL
sys.argv = ['s18_build.py', str(ROOT)]
sp = importlib.util.spec_from_file_location('s18_build', HERE / 's18_build.py')
T = importlib.util.module_from_spec(sp); sp.loader.exec_module(T)   # rewrites the 5 S18 boards byte-identical
HEIGHTS = T.HEIGHTS
SC, Q, s4 = T.SC, T.Q, T.s4
n, icon, lamp, num, dd, pair, NB = T.n, T.icon, T.lamp, T.num, T.dd, T.pair, T.NB
cT, cH, cP, SER, GUIDE, FARM = T.cT, T.cH, T.cP, T.SER, T.GUIDE, T.FARM
TODAY, REPLAY, PLC, ARR = T.TODAY, T.REPLAY, T.PLC, T.ARR
D = dt.date.fromisoformat
def wk(start, day): return (day - D(start)).days // 7 + 1
WK_P = wk(cP['start_date'], REPLAY); WK_T0 = wk(cT['start_date'], REPLAY); WK_T39 = wk(cT['start_date'], TODAY)
assert (WK_P, WK_T0, WK_T39) == (31, 1, 58)
assert D(cP['start_date']) < REPLAY < D(cP['end_date'])       # on 07.10.2025 the Seria 2025 cycle at Hîrbovăț was still open
assert REPLAY < D(cH['start_date'])                             # and Seria 2026 had no Hîrbovăț cycle yet
ROWS = [(c, q if sx == 'F' else None, q if sx == 'M' else None) for c, sx, q in T.PROWS]
s26 = SER['seria-2026']; NAME26, YEAR26 = s26['name'], str(s26['year'])

# ================================================================= CycleTrack for a scene (as s18_build.track_svg)
# a scene: the day, the blocks per farm lane, the transfer link (or none), one SAPT readout per open block
TXT = {  # labels per drawing width (11px track text)
    'h25open': {1160: 'Seria 2025 · ouat · din 07.03.2025', 1000: 'Seria 2025 · ouat · din 07.03.2025', 720: 'Seria 2025 · ouat', 560: 'Seria 2025 · ouat'},
    'tel0': {1160: 'Seria 2026 · creștere · din 07.10.2025', 1000: 'Seria 2026 · creștere · din 07.10.2025', 720: 'Seria 2026 · din 07.10.2025', 560: 'Seria 2026'},
    'tel39': {1160: 'Seria 2026 · creștere · din 07.10.2025', 1000: 'Seria 2026 · creștere · din 07.10.2025', 720: 'Seria 2026 · creștere', 560: 'Seria 2026'},
}
assert dd(cP['start_date']) == '07.03.2025' and dd(cT['start_date']) == '07.10.2025'
def scene_svg(TW, sc):
    LX, ro_w, d0 = 88, 56, dt.date(2025, 2, 1)
    day = sc['day']; d1 = sc['d1']
    while LX + (day - d0).days * (TW - LX) / (d1 - d0).days + 4 + ro_w > TW:
        d1 = dt.date(d1.year + d1.month // 12, d1.month % 12 + 1, 1)
    ppd = (TW - LX) / (d1 - d0).days
    def tx(iso): return round(LX + (D(iso) - d0).days * ppd, 1)
    LH, BH = 26, 20
    ly = {'telita': 22, 'hirbovat': 70}; H = ly['hirbovat'] + LH
    t = T.TXT[TW]
    g = ''.join(f'<text class="lanelbl" x="0" y="{ly[f] + 17}">{T.LBL[f]}</text><rect class="lane" x="{LX}" y="{ly[f]}" width="{TW - LX}" height="{LH}"/>' for f in ly)
    if sc.get('no25'): g += f'<text class="cyctxt" x="{LX + 12}" y="{ly["telita"] + 17}">{t["no25"]}</text>'
    xt = tx(day.isoformat())
    for b in sc['blocks']:
        f = b['farm']; y = ly[f] + 3; o = ' open' if b['open'] else ''
        xa = tx(b['a']); xb = tx(b['b']) if b['b'] else round(xt + ppd, 1)   # an open cycle runs through today (its day counts)
        w = round(xb - xa, 1)
        lbl = b['label'][TW] if isinstance(b['label'], dict) else t[b['label']]
        left = b.get('left')   # a one-day block: its label sits left of it, right-aligned
        txt = (f'<text class="cyctxt" x="{round(xa - 8, 1)}" y="{y + 14}" text-anchor="end">{lbl}</text>' if left else
               f'<text class="cyctxt{o}" x="{round(xa + 8, 1)}" y="{y + 14}">{lbl}</text>')
        marks = ''
        if b.get('place'): marks += f'<rect class="place" x="{round(xa - 1, 1)}" y="{y}" width="3" height="{BH}"/>'
        if b.get('ded'): marks += f'<line class="ded" x1="{round(xa - 3, 1)}" x2="{round(xa - 3, 1)}" y1="{ly[f]}" y2="{ly[f] + LH}"/>'
        g += (f'<a href="{b["href"]}" aria-label="{b["aria"]}"><title>{b["aria"]}</title><rect class="ring" x="{round(xa - 4, 1)}" y="{y - 4}" width="{round(w + 8, 1)}" height="{BH + 8}" rx="3"/>'
              f'<rect class="cyc{o}" x="{xa}" y="{y}" width="{w}" height="{BH}" rx="2"/>{txt}{marks}</a>')
    top = ly['telita'] - 8
    assert xt + 4 + ro_w <= TW, (TW, xt)
    g += f'<line class="today" x1="{xt}" x2="{xt}" y1="{top}" y2="{H + 4}"/><text class="todaylbl" x="{xt}" y="{top - 3}" text-anchor="middle">azi {dd(day)[:5]}</text>'
    for f, w_ in sc['ro']:
        yh = ly[f] + 5
        g += (f'<rect class="ro" x="{round(xt + 4, 1)}" y="{yh}" width="{ro_w}" height="16" rx="2"/>'
              f'<text class="rotxt" x="{round(xt + 4 + ro_w / 2, 1)}" y="{yh + 12}" text-anchor="middle">SAPT {w_}</text>')
    tk = T.month_axis(d0, d1, tx, LX, H, 3 if TW <= 720 else 2)
    return (f'<svg class="ip-track s18-track s18-track--{TW}" viewBox="0 0 {TW} {H + 24}" role="group" aria-label="Ciclurile seriilor pe ferme">'
            f'<desc>{sc["desc"]}</desc>{g}{tk}</svg>')
LEG_NOTR = re.sub(r'<span><i class="ip-rec ip-rec--transfer"></i>transfer <span class="ip-num">[^<]*</span></span>', '', T.LEG)
assert LEG_NOTR != T.LEG and 'transfer' not in LEG_NOTR
def scene_track(sc):
    svgs = ''.join(scene_svg(w, sc) for w in T.TWS)
    return (f'<section class="s4-block" aria-labelledby="h-track"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-track">Cicluri pe ferme</h2>'
            f'<span class="ip-note">{sc["sub"]}</span></div>'
            f'<div class="s4-tight"><div class="s18-trackbox">{svgs}</div><div class="ip-tracklegend" aria-hidden="true">{sc.get("leg", LEG_NOTR)}</div></div></section>')

B25_OPEN = dict(farm='hirbovat', a=cP['start_date'], b=None, open=True, label=TXT['h25open'], ded=True, href='#/ferme/hirbovat?ciclu=cyc-hirbovat-2025',
                aria=f'Hîrbovăț, Seria 2025, ouat, început dedus {dd(cP["start_date"])}, deschis, azi sapt. {WK_P}. Deschide ferma la acest ciclu.')
B25_CLOSED = dict(farm='hirbovat', a=cP['start_date'], b=cP['end_date'], open=False, label='h25', ded=True, href='#/ferme/hirbovat?ciclu=cyc-hirbovat-2025',
                  aria=f'Hîrbovăț, Seria 2025, ouat, început dedus {dd(cP["start_date"])}, încheiat prin golire {dd(cP["end_date"])}, {cP["weeks"]} săptămâni, închis. Deschide ferma la acest ciclu.')
PLACE_TXT = f'găini {n(PLC["F"])}, cocoși {n(PLC["M"])}'
def b_tel(day, label, left=False):
    return dict(farm='telita', a=cT['start_date'], b=None, open=True, label=label, place=True, left=left, href='#/ferme/telita?ciclu=cyc-telita-2026',
                aria=f'Telița, Seria 2026, creștere, din {dd(cT["start_date"])}, deschis, azi sapt. {wk(cT["start_date"], day)}. Început prin populare ({PLACE_TXT}). Deschide ferma la acest ciclu.')
D1_REPLAY = dt.date(2026, 1, 1)   # the replay axis runs to the end of 2025 (then on by whole months until the readout fits)
LEG_BEFORE = re.sub(r'<span><i class="ip-rec ip-rec--placement"></i>populare</span>', '', LEG_NOTR); assert LEG_BEFORE != LEG_NOTR   # S18P: no placement mark yet
SC_BEFORE = dict(day=REPLAY, d1=D1_REPLAY, no25=True, blocks=[B25_OPEN], ro=[('hirbovat', WK_P)], leg=LEG_BEFORE,
                 sub=f'{num(1)} serie · {num(2)} ferme · {num(1)} ciclu',
                 desc=(f'1 serie, 2 ferme, 1 ciclu. Seria 2025, deschisă: Hîrbovăț ouat, început dedus {dd(cP["start_date"])}, deschis, azi {dd(REPLAY)}, '
                       f'sapt. {WK_P}; la Telița fără date.'))
SC_AFTER = dict(SC_BEFORE, leg=LEG_NOTR, blocks=[B25_OPEN, b_tel(REPLAY, TXT['tel0'], left=True)], ro=[('telita', WK_T0), ('hirbovat', WK_P)],
                sub=f'{num(2)} serii · {num(2)} ferme · {num(2)} cicluri',
                desc=(f'2 serii, 2 ferme, 2 cicluri. Seria 2026, deschisă: Telița creștere din {dd(REPLAY)}, început prin populare ({PLACE_TXT}), azi sapt. {WK_T0}. '
                      f'Seria 2025, deschisă: Hîrbovăț ouat, început dedus {dd(cP["start_date"])}, azi sapt. {WK_P}; la Telița fără date.'))
SC_D39 = dict(day=TODAY, d1=dt.date(2027, 1, 1), no25=True, blocks=[B25_CLOSED, b_tel(TODAY, TXT['tel39'])], ro=[('telita', WK_T39)],
              sub=f'{num(2)} serii · {num(2)} ferme · {num(2)} cicluri',
              desc=(f'2 serii, 2 ferme, 2 cicluri. Seria 2026, deschisă: Telița creștere din {dd(cT["start_date"])}, '
                    f'început prin populare ({PLACE_TXT}), deschis, azi {dd(TODAY)}, sapt. {WK_T39}. Seria 2025, închisă: Hîrbovăț ouat, început dedus '
                    f'{dd(cP["start_date"])}, încheiat prin golire {dd(cP["end_date"])}, {cP["weeks"]} săpt.; la Telița fără date.'))

# ================================================================= the series table for a scene (s18_build.crow / grow / acts)
def crow_at(fid, cyc, start, week, dedus=False):
    """s18_build.crow for a cycle as it stood on the scene's day: open cycles read SAPT <week>; an open inferred start
    keeps its "dedus" word."""
    keep = T.WK; T.WK = week
    try: r = T.crow(fid, cyc, start, dedus=dedus)
    finally: T.WK = keep
    if dedus and not cyc['end_date']:
        a = f'din {num(dd(cyc["start_date"]))}'; assert r.count(a) == 1
        r = r.replace(a, a + '<span class="ip-note"> dedus</span>')
    return r
def open_(c): return dict(c, end_date=None)
def why(sid, farm, week):
    return f'<p class="ip-note s4s-m0" id="why-{sid}">{SER[sid]["name"]} se poate închide după ce se încheie ciclul de la {farm} (deschis, sapt. {week}).</p>'
def gro(sid, st, note, acts):
    return (f'<tr class="is-group"><th scope="rowgroup" colspan="5" id="g-{sid}"><span class="s18-g"><span class="s18-gname">{SER[sid]["name"]}</span>{st}'
            f'<span class="ip-note">{note}</span></span></th><td colspan="2" class="s18-gact"><span class="s18-acts">{acts}</span></td></tr>')
def acts(sid, close=''):
    a = f'<a class="ip-btn ip-btn--ghost" href="#/serii/{sid}">Deschide seria<span class="ip-sr"> {SER[sid]["name"]}</span></a>'
    if close == 'off':
        a += f'<button type="button" class="ip-btn ip-btn--secondary" disabled="disabled" aria-describedby="why-{sid}">Închide seria<span class="ip-sr"> {SER[sid]["name"]}</span></button>'
    elif close == 'open':   # the Dialog is open: the key that opened it shows pressed
        a += f'<button type="button" class="ip-btn ip-btn--secondary is-pressed" aria-haspopup="dialog" aria-expanded="true">Închide seria<span class="ip-sr"> {SER[sid]["name"]}</span></button>'
    return a
assert gro('seria-2026', lamp('ok', 'deschisă'), f'an {num(2026)} · populată {num(dd(REPLAY))} la Telița', acts('seria-2026', 'off')) == T.grow('seria-2026')
assert gro('seria-2025', lamp('off', 'închisă'), f'an {num(2025)} · doar citire · ultimul ciclu încheiat {num(dd(cP["end_date"]))}', acts('seria-2025')) == T.grow('seria-2025')
def stable(groups, notes, count):
    body = ''.join(f'<tbody>{g}</tbody>' for g in groups)
    return (f'<section class="s4-block" aria-labelledby="h-ser"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-ser">Serii</h2>'
            f'<span class="ip-note">{count}</span></div>'
            f'<div class="ip-tablescroll ip-tablescroll--framed" tabindex="0" role="region" aria-labelledby="h-ser">'
            f'<table class="ip-table ip-table--rowheads s18-table"><caption class="ip-sr">Seriile și ciclurile lor pe ferme</caption>{T.DCOLS}{body}</table></div>'
            f'<div class="s18-notes">{"".join(notes)}</div></section>')
NOTE25_OPEN = f'an {num(2025)} · început dedus {num(dd(cP["start_date"]))} la Hîrbovăț'
G25_OPEN = (gro('seria-2025', lamp('ok', 'deschisă'), NOTE25_OPEN, acts('seria-2025', 'off'))
            + crow_at('hirbovat', open_(cP), T.START_P, WK_P, dedus=True) + T.crow('telita'))
G26_NEW = (gro('seria-2026', lamp('ok', 'deschisă'), f'an {num(2026)} · populată {num(dd(REPLAY))} la Telița', acts('seria-2026', 'off'))
           + crow_at('telita', open_(cT), T.START_T, WK_T0))
TBL_BEFORE = stable([G25_OPEN], [why('seria-2025', 'Hîrbovăț', WK_P)], f'{num(1)} serie')   # S18P: no closed series, no read-only note
TBL_AFTER = stable([G26_NEW, G25_OPEN], [why('seria-2026', 'Telița', WK_T0), why('seria-2025', 'Hîrbovăț', WK_P)], f'{num(2)} serii · cele mai noi întâi')
G26_D39 = (gro('seria-2026', lamp('ok', 'deschisă'), f'an {num(2026)} · populată {num(dd(REPLAY))} la Telița', acts('seria-2026', 'off'))
           + crow_at('telita', open_(cT), T.START_T, WK_T39))
G25_REAL = T.grow('seria-2025') + T.crow('hirbovat', cP, T.START_P, dedus=True) + T.crow('telita')
TBL_D39 = stable([G26_D39, G25_REAL], [why('seria-2026', 'Telița', WK_T39), T.RO_NOTE], f'{num(2)} serii · cele mai noi întâi')
# Închide seria (D39): the real list with Seria 2025 not closed yet; its last cycle ended 28.01.2026, so the key is live
G25_CLOSABLE = gro('seria-2025', lamp('ok', 'deschisă'), f'an {num(2025)} · ultimul ciclu încheiat {num(dd(cP["end_date"]))} · se poate închide', acts('seria-2025', 'open'))
TBL_REAL = T.table(); assert TBL_REAL.count(T.grow('seria-2025')) == 1
TBL_CLOSE = TBL_REAL.replace(T.grow('seria-2025'), G25_CLOSABLE)
NEWNAME_TXT = 'O serie închisă e doar citire.'; assert NEWNAME_TXT in T.RO_NOTE

O25 = f'<b>Seria 2025</b><span class="ip-lamp ip-lamp--off">închisă</span></span><span class="ip-num">până la {dd(cP["end_date"])}</span>'
O25_OPEN = f'<b>Seria 2025</b><span class="ip-lamp ip-lamp--ok">deschisă</span></span><span class="ip-num">din {dd(cP["start_date"])}</span>'
O26 = re.compile(r'<li class="ip-menu__item ip-select__opt" role="option" id="ser-2026" aria-selected="true">.*?</li>')
CHK = f'<span class="ip-select__chk" aria-hidden="true">{icon("check", 16)}</span>'

def sidebar(when='today'):
    """The owner Sidebar; its SeriesSelector follows the board's day: on 07.10.2025 Seria 2025 was still open, and before
    Creează seria only Seria 2025 existed (selected); 'close' = the D39 Seria 2025 not closed yet."""
    sb = T.sidebar_owner()
    if when == 'today': return sb
    assert sb.count(O25) == 1
    if when == 'close': return sb.replace(O25, O25.replace('ip-lamp--off">închisă', 'ip-lamp--ok">deschisă'))
    sb = sb.replace(O25, O25_OPEN)
    if when == 'before':
        v26, c25 = '<span class="ip-select__val"><b>Seria 2026</b></span>', 'id="ser-2025" aria-selected="false"><span class="ip-select__chk" aria-hidden="true"></span>'
        assert len(O26.findall(sb)) == 1 and sb.count(v26) == 1 and sb.count(c25) == 1
        sb = O26.sub('', sb).replace(c25, 'id="ser-2025" aria-selected="true">' + CHK).replace(v26, '<span class="ip-select__val"><b>Seria 2025</b></span>')
    return sb
def page(track, table, pressed=False, inert=False, when='today'):
    i = ' inert="inert"' if inert else ''
    return f'<div class="ip-shell"{i}>{sidebar(when)}<main class="ip-shell__main">\n{T.header(pressed)}\n{track}\n{table}\n</main></div>'
assert page(T.track(), T.table()) == T.page()

# ================================================================= loading / load error / no access
def skel(w): return f'<span class="ip-skel ip-skel--{w}"></span>'
SK_G = f'<tr class="is-group ip-table__skel"><th scope="rowgroup" colspan="5">{skel("w2")}</th><td colspan="2">{skel("w3")}</td></tr>'
SK_R = '<tr class="ip-table__skel">' + ''.join(f'<td>{skel(w)}</td>' for w in ('w3', 'w3', 'w3', 'w2', 'w3', 'w2', 'w3')) + '</tr>'
LOADING = (f'<p class="ip-sr" role="status">Se încarcă seriile și ciclurile lor pe ferme.</p>'
           f'<section class="s4-block" aria-labelledby="h-track" aria-busy="true"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-track">Cicluri pe ferme</h2></div>'
           f'<div class="s4-tight">{SKL.track()}{SKL.legend()}</div></section>\n'  # fix 1: + the legend line, as loaded
           f'<section class="s4-block" aria-labelledby="h-ser" aria-busy="true"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-ser">Serii</h2></div>'
           f'<div class="ip-tablescroll ip-tablescroll--framed" aria-hidden="true"><table class="ip-table ip-table--rowheads s18-table">{T.DCOLS}'
           f'<tbody>{SK_G}{SK_R}{SK_R}</tbody><tbody>{SK_G}{SK_R}{SK_R}</tbody></table></div></section>')
FAULT = f'<svg class="ip-icon ip-icon--16" viewBox="0 0 24 24" aria-hidden="true">{Q.Z.FAULT}</svg>'
LOADERR = (f'<section class="s4-block" aria-labelledby="h-ser"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-ser">Serii</h2></div>'
           f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{FAULT}'
           f'<div class="ip-banner__body"><p class="s4s-m0">Seriile și ciclurile lor pe ferme nu s-au încărcat.</p>'
           '<p class="s4s-m0">Verifică conexiunea la internet, apoi reîncearcă. Poți crea o serie nouă și acum.</p></div>'
           '<button type="button" class="ip-btn ip-btn--secondary ip-banner__key">Reîncearcă</button></div></section>')
def noaccess():
    step = f'<a class="ip-btn ip-btn--secondary" href="#acasa">{icon("home", 16)}Înapoi la Acasă</a>'   # fix 1: works for Observator, manager and vet (S17)
    box = s4.empty_box('fără acces', ['Nu ai acces la Serii.',
                                      'Seriile le creează și le închide proprietarul. Rolul tău, Observator, vede paginile seriilor și ale fermelor, și rapoartele.'],
                       step, 'Acces')
    return f'<div class="ip-shell">{Q.sidebar_viewer()}<main class="ip-shell__main">\n<header class="ip-pagehead"><div><h1>Serii</h1></div></header>\n{box}\n</main></div>'

# ================================================================= Serie nouă form states
s4.IC.setdefault('chevup', '<path d="m18 15-6-6-6 6"/>')   # Lucide chevron-up (open SectionSelect, as S9)
ERRI = icon('alert', 14)
def err(i, text): return f'<p class="ip-err ip-fieldmsg" id="{i}">{ERRI}<span>{text}</span></p>'
def name_err():
    f = T.tfield('sn-name', 'Nume', '', 'Apare în selectorul de serii, pe pagini și în rapoarte.', False, focus=True)
    a = '<span class="ip-input is-focus"><input id="sn-name" type="text" autocomplete="off" aria-describedby="sn-name-h" autofocus="autofocus"/></span><span class="ip-hint" id="sn-name-h">Apare în selectorul de serii, pe pagini și în rapoarte.</span>'
    assert f.count(a) == 1, f
    return f.replace(a, '<span class="ip-input is-focus is-error"><input id="sn-name" type="text" autocomplete="off" aria-invalid="true" aria-describedby="sn-name-e" autofocus="autofocus"/></span>'
                     + err('sn-name-e', 'Scrie numele seriei, de exemplu Seria 2027.'))
def year_err(year):
    f = T.tfield('sn-year', 'An', year, '', False, mono=True, numeric=True, cls=' s18-year')
    a = f'<span class="ip-input ip-num"><input id="sn-year" type="text" inputmode="numeric" autocomplete="off" value="{year}"/></span>'
    assert f.count(a) == 1, f
    return f.replace(a, f'<span class="ip-input ip-num is-error"><input id="sn-year" type="text" inputmode="numeric" autocomplete="off" value="{year}" aria-invalid="true" aria-describedby="sn-year-e"/></span>'
                     + err('sn-year-e', f'Anul {num(year)} e deja folosit de {NAME26}. Scrie anul seriei noi, de exemplu {num(2027)}.'))
def date_err(typed):
    f = T.date_field(TODAY, False)
    a = '<span class="ip-seg__item is-selected" role="radio" aria-checked="true" tabindex="0">Azi</span>'; assert f.count(a) == 1
    f = f.replace(a, '<span class="ip-seg__item" role="radio" aria-checked="false" tabindex="0">Azi</span>')
    a = f'<span class="ip-input ip-num">{icon("cal", 16)}<input id="d-val" type="text" inputmode="numeric" value="{dd(TODAY)}" aria-labelledby="d-lbl" aria-describedby="d-wk"/></span>'
    assert f.count(a) == 1, f
    f = f.replace(a, f'<span class="ip-input ip-num is-error">{icon("cal", 16)}<input id="d-val" type="text" inputmode="numeric" value="{dd(typed)}" aria-labelledby="d-lbl" aria-invalid="true" aria-describedby="d-err"/></span>')
    a = f'<span class="ip-note ip-num" id="d-wk">SAPT 1 · ciclu nou din {dd(TODAY)}</span>'; assert f.count(a) == 1
    return f.replace(a, err('d-err', f'{num(dd(typed))} e în viitor. Alege Azi, {num(dd(TODAY))}, sau o zi trecută.'))
def rows_err():
    """Two rows, both H1, no count: the second row's duplicate error (SYS07 words) and the list's 'niciun rând completat'."""
    s = T.prows('d', TODAY, [('H1', None, None), ('H1', None, None)], touch=False)
    a = 'aria-expanded="false" aria-controls="sm-d1" aria-labelledby="sl-d1"'; assert s.count(a) == 1
    a2 = 'class="ip-input ip-select__btn" id="sc-d1"'; assert s.count(a2) == 1
    s = s.replace(a2, 'class="ip-input ip-select__btn is-error" id="sc-d1"')
    s = s.replace(a, a + ' aria-invalid="true" aria-describedby="msg-d1"')
    a = 'aria-label="Elimină H1">'; assert s.count(a) == 2
    s = s.replace(a, 'aria-label="Elimină H1, rândul 1">', 1).replace(a, 'aria-label="Elimină H1, rândul 2">', 1)
    msg = f'<p class="ip-err ip-sexrow__msg" id="msg-d1">{ERRI}<span>H1 e deja în listă. Alege altă secțiune sau elimină rândul.</span></p>'
    lis = s.split('</li>'); assert len(lis) == 3
    s = lis[0] + '</li>' + lis[1] + msg + '</li>' + lis[2]
    a = '<span class="ip-sexrows__sub"><span>niciun rând completat</span></span>'; assert s.count(a) == 1
    s = s.replace(a, '<span class="ip-sexrows__sub"><span>niciun rând completat</span> · H1 de două ori</span>')
    for k in ('f-d0', 'm-d0', 'f-d1', 'm-d1'):   # fix 1: the list error names every empty count field
        a = f'<input id="{k}" inputmode="numeric"'; assert s.count(a) == 1, k
        s = s.replace(a, f'<input id="{k}" aria-invalid="true" aria-describedby="rows-e" inputmode="numeric"')
    assert s.endswith('</div></section>')
    return s[:-len('</section>')] + err('rows-e', 'Niciun rând completat. Scrie câte găini sau cocoși intră în cel puțin o secțiune a Teliței.') + '</section>'
def form_errors():
    return [name_err(), year_err(YEAR26), T.farm_field(TODAY, False), date_err(TODAY + dt.timedelta(days=1)), rows_err(), T.effect('empty', TODAY)]
SAVE_ERR = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert" id="sn-err">{icon("alert", 16)}<div class="ip-banner__body">'
            f'<p class="ip-m0">{NAME26} nu s-a creat: conexiunea s-a întrerupt. Ce ai scris rămâne aici; verifică internetul, apoi reîncearcă.</p>'
            '</div><button type="button" class="ip-btn ip-btn--secondary ip-btn--touch ip-banner__key ip-focus">Reîncearcă</button></div>')
def with_banner(parts, banner):
    """S9P: the fault Banner sits right above the EffectPreview (the last part)."""
    assert parts[-1].startswith('<div class="ip-effect')
    return parts[:-1] + [banner, parts[-1]]

# ---- phone: the section sheet (§33, S9P markup) over the inert form; the third row's SectionSelect is open
TSECS = T.TEL_SECS
def sheet_select(k, used):
    def oid(c): return f'so-{k}-{c.replace(".", "-")}'
    first = next(c for c in TSECS if c not in used)
    opts = ''.join(f'<div class="ip-menu__item{" is-active" if c == first else ""}" role="option" id="{oid(c)}" aria-selected="false"'
                   f'{" aria-disabled=\"true\"" if c in used else ""}><span class="ip-sheet__code"><span class="ip-sr">Telița </span>{c}</span>'
                   f'<span class="ip-note">{"în listă" if c in used else "goală"}</span></div>' for c in TSECS)
    return f'''
<!-- BottomSheet (modal, §33): focus on the listbox, the first enabled option active (aria-activedescendant); arrows move it, Enter or a tap picks it and closes
     the sheet; Tab cycles Închide ↔ listbox; Esc or Închide closes it and focus returns to the row's combobox. The Serie nouă sheet behind is inert. -->
<div class="ip-scrim ip-scrim--sheet">
<div class="ip-sheet ip-sheet--select" role="dialog" aria-modal="true" id="sh-{k}" aria-labelledby="sh-{k}-t" aria-describedby="sh-{k}-d">
<div class="ip-sheet__handle" aria-hidden="true"></div>
<div class="ip-sheet__head"><div><h2 class="ip-t-heading ip-sheet__title" id="sh-{k}-t">Rândul 3: alege secțiunea</h2><p class="ip-note ip-sheet__sub" id="sh-{k}-d">Telița e goală; H1 și H2 sunt deja în listă.</p></div><button type="button" class="ip-iconbtn ip-iconbtn--touch" aria-label="Închide lista">{icon("x")}</button></div>
<div class="ip-sheet__list ip-focus" role="listbox" id="sl-{k}-lb" aria-labelledby="sh-{k}-t" tabindex="0" aria-activedescendant="{oid(first)}">
<div role="group" aria-labelledby="sg-{k}-t"><div class="ip-menu__group" id="sg-{k}-t">Telița · creștere</div>{opts}</div>
</div>
</div>
</div>'''
def rows_picking():
    """H1 and H2 typed (the real figures), a third row added and its SectionSelect open (phone: aria-haspopup=dialog,
    aria-controls on the open sheet, no hidden listbox: S9P)."""
    s = T.prows('p', REPLAY, [ROWS[0], ROWS[1], (None, None, None)], touch=True)
    a = ('<div class="ip-input ip-select__btn ip-input--touch" id="sc-p2" role="combobox" tabindex="0" aria-haspopup="dialog" aria-expanded="false" '
         'aria-labelledby="sl-p2"><span class="ip-select__val is-empty">Alege</span>' + T.CHEV16 + '</div>')
    assert s.count(a) == 1, s
    up = icon('chevup', 16)
    s = s.replace(a, '<div class="ip-input ip-select__btn ip-input--touch" id="sc-p2" role="combobox" tabindex="0" aria-haspopup="dialog" aria-expanded="true" '
                     f'aria-controls="sh-p2" aria-labelledby="sl-p2"><span class="ip-select__val is-empty">Alege</span>{up}</div>')
    return s   # the other rows' comboboxes open the sheet too (phone: aria-haspopup=dialog, no listbox; s18_build since FIN10)

# ================================================================= CSS (screen only, tokens only)
CSS18S = '''
.ip-sidepanel--dock > .ip-sidepanel__body{position:relative} /* fix 1: hidden text in the panel body is placed inside it, so the board stays 1024 */'''
CSS_D = T.CSS_DESK + CSS18S
CSS_SH = T.CSS_SH + SC.MODAL

# ================================================================= write
STEM, TT = T.STEM, T.T
out = []
def desk(stem, title, root, h=1024):
    out.append(T.write(STEM + stem, f'{TT} · {title}', f'<div class="ip-root s4s-rel" style="padding:0; min-height:{h}px">{root}</div>', CSS_D, 1440, h))
def hd(stem, dflt=1024): return HEIGHTS.get(STEM + stem, dflt)
DKEY = f'<button type="button" class="ip-btn ip-btn--primary">{T.CREATE}</button>'
def vp(track, table, when='today'): return f'<div class="s13-vp s18-vp">{page(track, table, pressed=True, when=when)}</div>'

# 1 · se încarcă
desk('--se-incarca', 'se încarcă', page(LOADING, ''), hd('--se-incarca'))
# 2 · eroare la încărcare
desk('--eroare-la-incarcare', 'eroare la încărcare', page(LOADERR, ''), hd('--eroare-la-incarcare'))
# 3 · fără acces (Observator; a manager or a vet sees the same picture)
desk('--fara-acces', 'observator · fără acces', noaccess(), hd('--fara-acces'))
# 4 · desktop, values typed: replay 07.10.2025, the list as it stood before the creation
body = T.form(REPLAY, False, NAME26, YEAR26, ROWS, state='set')
desk('--desktop-scris', f'Serie nouă · desktop · scris · reluare {dd(REPLAY)}',
     vp(scene_track(SC_BEFORE), TBL_BEFORE, 'before') + T.panel(body, REPLAY, DKEY), SC.SIDE_H)
# 5 · desktop, field errors after Creează seria (16.11.2026, D39 values), focus on the first error (Nume)
desk('--desktop-erori', 'Serie nouă · desktop · erori în câmpuri · exemplu de stare (D39)',
     vp(T.track(), T.table()) + T.panel(form_errors(), TODAY, DKEY), SC.SIDE_H)
# 6 · desktop refusal (D39, as the phone board): the list drawn on the same premise
RKEY = f'<button type="button" class="ip-btn ip-btn--primary" disabled="disabled" aria-describedby="sn-ref">{T.CREATE}</button>'
body = T.form(TODAY, False, T.NEWNAME, T.NEWYEAR, state='refuz')
desk('--desktop-refuzat', 'Serie nouă · desktop · refuzat: Telița are încă păsări · exemplu de stare (D39): transferul la Hîrbovăț neintrodus',
     vp(scene_track(SC_D39), TBL_D39) + T.panel(body, TODAY, RKEY), SC.SIDE_H)
# 7 · salvat: replay 07.10.2025, panel closed, the new series in the list, Toast
TOAST = (f'<div class="ip-toastwrap ip-toastwrap--page"><div class="ip-toast" role="status">{icon("check", 16)}<span class="ip-toast__txt">{NAME26} creată · '
         f'ciclu nou la Telița · sapt. {num(WK_T0)} de la {num(dd(REPLAY))} · ghid {GUIDE["g-growth"]}</span>'
         f'<button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')   # fix 1: undo until 23:59 (owner), as S9/S10; S18P: B4 words
assert WK_T0 == 1 and GUIDE[cT['guide_id']] == GUIDE['g-growth'] == 'Ross 308 creștere'
NEWKEY = T.newkey(); assert NEWKEY.count('class="ip-btn ip-btn--secondary"') == 1
FOCUS_NOTE = ('<!-- S18P focus after Creează seria: the panel closes and focus returns to Adaugă serie, the key that opened it; '
              'the Toast (role=status) announces the save and keeps Anulează in the tab order. -->')
SAVED = page(scene_track(SC_AFTER), TBL_AFTER, when='after'); assert SAVED.count(NEWKEY) == 1
SAVED = SAVED.replace(NEWKEY, FOCUS_NOTE + NEWKEY.replace('class="ip-btn ip-btn--secondary"', 'class="ip-btn ip-btn--secondary ip-focus"'))
desk('--salvat', f'salvat · reluare {dd(REPLAY)}', SAVED + TOAST, hd('--salvat'))
# 8 · Închide seria Dialog (D39: Seria 2025 not closed yet), over the inert page, focus on Anulează
DLG = (f'<div class="ip-scrim ip-scrim--page"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-n">'
       f'<h2 id="dlg-t">Închide Seria 2025</h2>'
       f'<p class="ip-dialog__note" id="dlg-n">Ultimul ei ciclu, Hîrbovăț ouat, s-a încheiat prin golire pe {num(dd(cP["end_date"]))}, după {num(cP["weeks"])} săpt. '
       f'Închisă, seria rămâne doar citire: paginile și rapoartele o arată mai departe, dar nu mai primește înregistrări sau corecții.</p>'
       f'<div class="ip-dialog__keys"><button type="button" class="ip-btn ip-btn--secondary ip-focus" autofocus="autofocus">Anulează</button>'
       f'<button type="button" class="ip-btn ip-btn--primary">Închide seria</button></div></div></div>')
desk('--inchide-seria', 'Închide seria · exemplu de stare (D39)', page(T.track(), TBL_CLOSE, inert=True, when='close') + DLG, hd('--inchide-seria'))
# phones
PKEY = f'<button type="button" class="ip-btn ip-btn--primary ip-btn--touch">{T.CREATE}</button>'
def phone(stem, title, inner, h):
    out.append(T.write(STEM + stem, f'{TT} · {title}', f'<div class="ip-root" style="padding:0; width:390px; height:{h}px; overflow:hidden">{inner}</div>', CSS_SH, 390, h))
# 9 · save error: replay 07.10.2025, the draft kept, Banner + Reîncearcă above the preview, focus on Reîncearcă
body = with_banner(T.form(REPLAY, True, NAME26, YEAR26, ROWS, state='set'), SAVE_ERR)
phone('--telefon-eroare-la-salvare', f'Serie nouă · telefon · eroare la salvare · reluare {dd(REPLAY)}',
      f'<div class="s18-screen">{T.sheet(body, REPLAY, PKEY)}</div>', hd('--telefon-eroare-la-salvare', 844))
# 10 · section sheet open: replay 07.10.2025, H1 and H2 typed, the third row picking (390×844, as S9P)
body = T.form(REPLAY, True, NAME26, YEAR26, state='set')
assert body[4].startswith('<section class="ip-panel ip-sexrows') and body[5].startswith('<div class="ip-effect')
body[4] = rows_picking(); body[5] = T.effect('set', REPLAY, NAME26, YEAR26, [ROWS[0], ROWS[1]])
SH = T.sheet(body, REPLAY, PKEY).replace('<aside class="ip-sidepanel ip-sidepanel--sheet" role="dialog" aria-modal="true" aria-labelledby="sn-t">',
                                          '<aside class="ip-sidepanel ip-sidepanel--sheet" role="dialog" aria-modal="true" aria-labelledby="sn-t" inert="inert">', 1)
assert 'inert="inert"' in SH and 'autofocus' not in SH   # focus is in the section sheet
phone('--telefon-alege-sectiunea', f'Serie nouă · telefon · alege secțiunea · reluare {dd(REPLAY)}',
      f'<div class="s18-screen ip-scrimhost">{SH}{sheet_select("p2", ("H1", "H2"))}</div>', 844)

S18_OUT = json.loads(pathlib.Path('/tmp/s18/boards.json').read_text())
assert [b['stem'] for b in S18_OUT] == [STEM, STEM + '--serie-noua', STEM + '--telefon', STEM + '--telefon-serie-noua', STEM + '--telefon-refuzat']
pathlib.Path('/tmp/s18/boards.json').write_text(json.dumps(S18_OUT + [{k: b[k] for k in ('stem', 'w', 'h')} for b in out]))

if CANVAS:  # S18 row: S18 desktops, these desktops, the S18 phones, these phones
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    old = [b['stem'] for b in S18_OUT]
    new = lambda w: [(b['stem'], w, b['h'], b['title']) for b in out if b['w'] == w]
    x = CR.place(c, STEM, old[:2] + new(1440) + old[2:] + new(390), 0)
    CR.save(ROOT, c)
    print('canvas: S18 row y', CR.row_y(c, STEM), 'ends at x', x - 80)
