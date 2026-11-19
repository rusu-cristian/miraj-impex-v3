"""S19S · Ferme și secțiuni · state boards (spec B1; D110, D75, D111). Imports s19_build (which rewrites the 10 S19 boards,
and through s18_build the S18/S17/S14S/S2/S7 boards, byte-identical) and draws the S19 states on the same row (y 34408).
Desktops first (a Setări page, as S17S/S18S): the new desktops follow the nine S19 desktops; the S19 phone moves right
(x only). RF11: realdata through s19_build (today 18.11.2026, D152; age weeks, D151). Real: farms, types, codes, halls,
Hîrbovăț open in today's age week (H7 empty since 14.11), Telița empty since its cycle closed; guide names illustrative (D22).
Every edit not in the extracts is a D39 example on the real codes (SYS10 top: ADD_CODE, DUP_CODE, TEL_*), said in the title.
Boards
  --se-incarca            header kept (no farm names in the meta yet), framed table skeleton, hidden status
  --eroare-la-incarcare   header kept, fault Banner + Reîncearcă in place of the table
  --fara-acces            Observator Sidebar, EmptyState "Nu ai acces la Ferme și secțiuni." + "Înapoi la Acasă"
                          (a manager or a vet: same picture with their own nav and role sentence; one board, as S17S/S18S)
  --salvat                D39: Telița's guide chosen from "fără ghid": panel closed, the list as it is in the extracts,
                          Toast + Anulează, focus back on Telița's Editează
  --eroare-la-salvare     D39: the same edit, save failed: draft kept (Ross 308 creștere), fault Banner + Reîncearcă (focus)
                          under the fields it covers, the list still says "fără ghid"
  --nume-gol              D39: Fermă nouă after Adaugă ferma with nothing filled: Nume (focus) + Tip errors, Ghid waits
  --nume-folosit          D39: Fermă nouă "Hirbovat" (ouat): the name is taken by Hîrbovăț (compared without case and
                          diacritics, D111)
  --ghid-lista            D39: Telița without a guide, the Ghid list open (Fără ghid · Ghiduri de creștere)
  --adauga-compartiment   D39: Hîrbovăț, Adaugă compartiment in the first hall, its next code (ADD_CODE) typed, save
                          failed: tree Banner + Reîncearcă (focus, S16P; FX04, was the row's Salvează), code kept
  --cod-duplicat          D39: Hîrbovăț, that hall's last code (DUP_CODE) typed in the new-compartment row: error, focus
  --redenumeste-cod-gol   D39: Telița, Redenumește its last section with the code cleared: error, focus in the field
  --retrasa-mutata        D39: Telița after the Retrage Dialog (its last section retrasă, listed last, no key) and then
                          Mută în sus on its third (a 2-a din 5; focus stays on that key; live region)
No screen CSS of its own (the Banner with Reîncearcă is additions.css §49, FIN03; the Toast is §48, FIN02); S19P: the skeleton rows
are the system `tr.ip-table__skel` (additions.css §42), the Tip error uses §41, the tree error line gets the alert icon. Fix 1: every
multi-line Banner puts its icon on the first line (s7s-banner, §49 with a key); Ghid on Fermă nouă is not required (s19_build).
python3 -I s19s_build.py <repo_root> [--canvas]   after `s19_build --canvas` (canvas_rows.py)
Renders: s19_render.py <stems> (the new boards are appended to /tmp/s19/boards.json: shots S19-10 … S19-21)."""
import sys, json, pathlib, importlib.util, io, contextlib

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve(); HERE = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
sys.argv = ['s19_build.py', str(ROOT)]
sp = importlib.util.spec_from_file_location('s19_build', HERE / 's19_build.py')
M = importlib.util.module_from_spec(sp)
with contextlib.redirect_stdout(io.StringIO()):
    sp.loader.exec_module(M)   # rewrites the 10 S19 boards byte-identical (no --canvas)
T, ST, SC, Q, s4 = M.T, M.ST, M.SC, M.Q, M.s4
icon, lamp, num, dd, NB = M.icon, M.lamp, M.num, M.dd, M.NB
FARM, GUIDE, WK, cH, cT = M.FARM, M.GUIDE, M.WK, M.cH, M.cT
ADD_HALL, ADD_CODE, DUP_CODE, TEL_RET, TEL_UP, TEL_UP_ORDER = M.ADD_HALL, M.ADD_CODE, M.DUP_CODE, M.TEL_RET, M.TEL_UP, M.TEL_UP_ORDER
S19_OUT = [{k: b[k] for k in ('stem', 'w', 'h')} for b in M.out]
STEM, TT = M.STEM, M.TT
assert [b['stem'] for b in S19_OUT][-1] == STEM + '--telefon' and len(S19_OUT) == 10

# ================================================================= shared parts
ERRI = M.ERRI
def err(i, text): return f'<p class="ip-err ip-fieldmsg" id="{i}">{ERRI}<span>{text}</span></p>'
HDR = M.header()
assert HDR.count(M.meta(True)) == 1
HDR_NOLIST = HDR.replace(M.meta(True), M.meta(False))   # before the list is in, the meta names no farm
def shell(main, sidebar=None, hdr=HDR):
    return f'<div class="ip-shell">{sidebar or M.sidebar_owner()}<main class="ip-shell__main">\n{hdr}\n{main}\n</main></div>'

# ---- loading: the Ferme block with a framed table skeleton (the real columns, two farm rows)
def skel(w): return f'<span class="ip-skel ip-skel--{w}"></span>'
SK_R = ('<tr class="ip-table__skel"><th scope="row">' + skel('w3') + '</th>' + ''.join(f'<td>{skel(w)}</td>' for w in ('w3', 'w2', 'w2', 'w2'))
        + f'<td class="s19-act"><span class="ip-skel ip-skel--w3 ip-skel--end"></span></td></tr>')
LOADING = (f'<p class="ip-sr" role="status">Se încarcă fermele, cu tipul, ghidul și secțiunile lor.</p>'
           f'<section class="ip-block" aria-labelledby="h-fer" aria-busy="true"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-fer">Ferme</h2></div>'
           f'<div class="ip-tablescroll ip-tablescroll--framed" aria-hidden="true"><table class="ip-table ip-table--rowheads s19-table">'
           f'{M.COLS}<tbody>{SK_R}{SK_R}</tbody></table></div></section>')
# ---- load error: the Banner in place of the table; Adaugă fermă stays (the name check runs on save)
LOADERR = (f'<section class="ip-block" aria-labelledby="h-fer"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-fer">Ferme</h2></div>'
           f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{icon("alert", 16)}'
           f'<div class="ip-banner__body"><p class="ip-m0">Fermele și secțiunile lor nu s-au încărcat.</p>'
           '<p class="ip-m0">Verifică conexiunea la internet, apoi apasă Reîncearcă. Poți adăuga o fermă nouă și acum.</p></div>'
           '<button type="button" class="ip-btn ip-btn--secondary ip-banner__key">Reîncearcă</button></div></section>')
# ---- no access (Observator; a manager or a vet: same picture, their own nav and role sentence)
def noaccess():
    step = f'<a class="ip-btn ip-btn--secondary" href="#acasa">{icon("home", 16)}Înapoi la Acasă</a>'
    box = s4.empty_box('fără acces', ['Nu ai acces la Ferme și secțiuni.',
                                      'Fermele, tipul, ghidul și secțiunile lor le setează proprietarul. Rolul tău, Observator, '
                                      'vede paginile seriilor și ale fermelor, și rapoartele.'], step, 'Acces')
    return f'<div class="ip-shell">{Q.sidebar_viewer()}<main class="ip-shell__main">\n<header class="ip-pagehead"><div><h1>Ferme și secțiuni</h1></div></header>\n{box}\n</main></div>'

# ---- panel fields
TEL_TIP = M.tip_field(sel='growth')
HIR_TIP = M.tip_field('hirbovat', readonly=True)
def name_field(val, e=None):
    """Nume with an error: focus on it (the first error, craft §5), typed value kept."""
    f = M.name_field(val, focus=True)
    a = 'class="ip-input is-focus"><input id="fp-name" type="text" autocomplete="off"'
    assert f.count(a) == 1, f
    f = f.replace(a, 'class="ip-input is-focus is-error"><input id="fp-name" type="text" autocomplete="off" aria-invalid="true" aria-describedby="fp-name-e"')
    assert f.endswith('</div>')
    return f[:-6] + err('fp-name-e', e) + '</div>'
def tip_err():
    f = M.tip_field()
    a = 'aria-required="true">'; assert f.count(a) == 1
    f = f.replace(a, 'aria-required="true" aria-invalid="true" aria-describedby="fp-tip-e">')
    return f[:-6] + err('fp-tip-e', 'Alege tipul fermei: Creștere sau Ouat. Lista de ghiduri depinde de el.') + '</div>'
NEW_GHID_OUAT = M.gbox(GUIDE['g-laying'], '<span class="ip-hint" id="fp-ghid-h">Doar ghiduri de ouat. Se aplică de la primul ciclu al fermei.</span>', 'laying')
def save_banner(*paras):   # D147: the lead, then the retry line + what is kept
    body = ''.join(f'<p class="ip-m0">{t}</p>' for t in paras)
    return (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert" id="fp-err">{icon("alert", 16)}<div class="ip-banner__body">'
            f'{body}</div><button type="button" class="ip-btn ip-btn--secondary ip-banner__key ip-focus">Reîncearcă</button></div>')

# ---- tree states (SYS10 helpers; one edit row open at a time)
def tree_x(fid, noguide=False, note=None, banner='', live='', **kw):
    fr = ST['farm_row'](fid, tag=M.gtag(fid, True) if noguide else None, note=note)
    return ST['tree']('st', f'Secțiuni, {FARM[fid]["name"]}', [fr] + ST['section_rows']('st', fid, **kw) + [ST['add_row'](fid)],
                      banner=banner, live=live)
def edit(level, label, value, e=None, focus=True, is_last=True):
    r = ST['edit_row']('st-e', level, label, value, err=e, focus=focus and not e, is_last=is_last)
    if e:   # the error row keeps the focus in its field (Enter = Salvează ran the check)
        a = 'class="ip-input ip-stree__in is-error"'; assert r.count(a) == 1
        r = r.replace(a, 'class="ip-input ip-stree__in is-error is-focus"')
        assert r.count(err('st-e-err', e)) == 1   # S19P: alert icon, as the panel field errors (SYS10 draws it since FIN13)
    return r
TREE_BANNER = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{icon("alert", 16)}<div class="ip-banner__body">'
               f'<p class="ip-m0">Compartimentul {ADD_CODE} nu s-a salvat: conexiunea s-a întrerupt.</p>'
               '<p class="ip-m0">Verifică conexiunea la internet, apoi apasă Reîncearcă. Codul scris rămâne în rând.</p></div>'
               '<button type="button" class="ip-btn ip-btn--secondary ip-banner__key ip-focus">Reîncearcă</button></div>')  # FX04 (D143): §49 + Reîncearcă (focus, S16P)
ORDER_UP = TEL_UP_ORDER
assert [c for c, _ in M.tree_of('telita')] == M.R.SECTIONS['telita'] and ORDER_UP.index(TEL_UP) == 1
MOVED = ST['moved'](len(ORDER_UP) - 1)   # TEL_RET retired first: 5 active

# ================================================================= CSS (screen only, tokens only)
CSS19S = ''   # FIN03: the Banner action key is additions.css §49 (.ip-banner__key); no screen CSS left
CSS_D = M.CSS_DESK + CSS19S

# ================================================================= write
out = []
def desk(stem, title, root, h=1024):
    out.append(M.write(STEM + stem, f'{TT} · {title}', f'<div class="ip-root s4s-rel" style="padding:0; min-height:{h}px">{root}</div>', CSS_D, 1440, h))
def hd(stem): return M.HEIGHTS.get(STEM + stem, 1024)
PH = M.PH
D39 = 'exemplu de stare (D39)'

# 1 · se încarcă
desk('--se-incarca', 'se încarcă', shell(LOADING, hdr=HDR_NOLIST), hd('--se-incarca'))
# 2 · eroare la încărcare
desk('--eroare-la-incarcare', 'eroare la încărcare', shell(LOADERR, hdr=HDR_NOLIST), hd('--eroare-la-incarcare'))
# 3 · fără acces (Observator; a manager or a vet sees the same picture)
desk('--fara-acces', 'observator · fără acces', noaccess(), hd('--fara-acces'))
# 4 · salvat (D39): Telița's guide chosen from "fără ghid"; the list now equals the extracts; focus back on Telița's Editează
EDK = M.editkey('telita'); assert EDK.count('class="ip-btn ip-btn--ghost"') == 1
FOCUS_NOTE = ('<!-- Focus after Salvează ferma: the panel closes and focus returns to the Editează key that opened it; the Toast '
              '(role=status) announces the save and keeps Anulează in the tab order. -->')
SAVED = shell(M.flist())
assert SAVED.count(EDK) == 1
SAVED = SAVED.replace(EDK, FOCUS_NOTE + EDK.replace('class="ip-btn ip-btn--ghost"', 'class="ip-btn ip-btn--ghost ip-focus"'))
TOAST = (f'<div class="ip-toastwrap ip-toastwrap--page"><div class="ip-toast" role="status">{icon("check", 16)}<span class="ip-toast__txt">Telița salvată · '
         f'ghid {GUIDE["g-growth"]} de la următorul ciclu al fermei</span>'
         '<button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
desk('--salvat', f'salvat · {D39}: ghidul Teliței ales din „fără ghid”', SAVED + TOAST, hd('--salvat'))
# 5 · eroare la salvare (D39, the same edit): draft kept, Banner + Reîncearcă (focus) under the fields it covers
BODY = [M.name_field('Telița'), TEL_TIP, M.ghid_field('telita'),
        save_banner('Telița nu s-a salvat: conexiunea s-a întrerupt.', f'Verifică conexiunea la internet, apoi apasă Reîncearcă. Ghidul ales, {GUIDE["g-growth"]}, rămâne în panou.'),
        M.tree_farm('telita', noguide=True)]
desk('--eroare-la-salvare', f'eroare la salvare · {D39}: ghidul Teliței ales din „fără ghid”',
     M.vp('telita', noguide=('telita',)) + M.panel('Telița', M.ctx('telita'), BODY, M.SAVE), PH)
# 6 · Nume gol (D39): Fermă nouă after Adaugă ferma with nothing filled; focus on Nume (first error)
NEWP = f'<div class="ip-dockpage s18-vp">{M.page(None, pressed_new=True)}</div>'
FREE = 'Telița 2'  # FIN12 (D133): the example names no farm in the extracts (Telița, Hîrbovăț are taken), as S20S's guide name
assert all(FREE.casefold() != f['name'].casefold() for f in json.loads((ROOT / 'design/data/extracts/farms-sections.json').read_text(encoding='utf-8'))['farms'])
desk('--nume-gol', f'Fermă nouă · nume gol și tip lipsă · {D39}',
     NEWP + M.panel('Fermă nouă', '', [name_field('', f'Scrie numele fermei, de exemplu {FREE}.'), tip_err(),
                                       M.ghid_field(None, 'new'), M.NEWTREE], M.ADD), PH)
# 7 · Nume deja folosit (D39, D111): "Hirbovat" equals Hîrbovăț without case and diacritics
TIP_OUAT = M.tip_field(sel='laying')
desk('--nume-folosit', f'Fermă nouă · nume deja folosit · {D39}: „Hirbovat”',
     NEWP + M.panel('Fermă nouă', '', [name_field('Hirbovat', 'Numele e deja folosit de ferma Hîrbovăț (literele mari și diacriticele nu contează). '
                                                  'Alege alt nume.'), TIP_OUAT, NEW_GHID_OUAT, M.NEWTREE], M.ADD), PH)
# 8 · Ghid list open (D39 fără ghid): filtered to the farm's type
desk('--ghid-lista', f'Telița fără ghid · lista Ghid deschisă · {D39}',
     M.vp('telita', noguide=('telita',)) + M.fpanel('telita', TEL_TIP, M.ghid_field('telita', 'open'), M.tree_farm('telita', noguide=True)), PH)
# 9 · Adaugă compartiment ADD_CODE, save failed (D39): Banner + Reîncearcă (focus, S16P; FX04) over the tree, the code kept in the row
r = edit(3, f'Compartiment nou în {ADD_HALL}', ADD_CODE, focus=False)
assert r.count('class="ip-btn ip-btn--secondary">Salvează') == 1
desk('--adauga-compartiment', f'Hîrbovăț · Adaugă compartiment {ADD_CODE} · salvarea a eșuat · {D39}',
     M.vp('hirbovat') + M.fpanel('hirbovat', HIR_TIP, M.ghid_field('hirbovat'), tree_x('hirbovat', banner=TREE_BANNER, after={ADD_HALL: r})), PH)
# 10 · cod duplicat (D39): DUP_CODE typed in the new-compartment row
desk('--cod-duplicat', f'Hîrbovăț · Adaugă compartiment · cod deja folosit · {D39}: „{DUP_CODE}”',
     M.vp('hirbovat') + M.fpanel('hirbovat', HIR_TIP, M.ghid_field('hirbovat'),
                                 tree_x('hirbovat', after={ADD_HALL: edit(3, f'Compartiment nou în {ADD_HALL}', DUP_CODE, ST['DUP_MSG'])})), PH)
# 11 · Redenumește TEL_RET with the code cleared (D39)
desk('--redenumeste-cod-gol', f'Telița · Redenumește {TEL_RET} · cod gol · {D39}',
     M.vp('telita') + M.fpanel('telita', TEL_TIP, M.ghid_field('telita'),
                               tree_x('telita', replace={TEL_RET: edit(2, f'Cod nou pentru {TEL_RET}', '', f'Scrie codul nou al secțiunii {TEL_RET} sau apasă Anulează.', is_last=False)})), PH)
# 12 · after the Retrage TEL_RET Dialog, then Mută în sus on TEL_UP (D39): the list beside counts 5 active sections in the new order
keep = M.tree_of
M.tree_of = lambda fid: [t for t in ST['farm_tree'](fid, ORDER_UP) if t[0] != TEL_RET] if fid == 'telita' else keep(fid)
try:
    VP = M.vp('telita')
finally:
    M.tree_of = keep
assert '<span class="ip-num">5</span> secțiuni' in VP and TEL_RET not in VP.split('<span class="s19-fn">Telița</span>')[1].split('</tr>')[0]
desk('--retrasa-mutata', f'Telița · {TEL_RET} retrasă, apoi {TEL_UP} mutată în sus · {D39} · anunț live „{MOVED}”',
     VP + M.fpanel('telita', TEL_TIP, M.ghid_field('telita'),
                   tree_x('telita', note=ST['farm_note']('telita', 1), live=MOVED, order=ORDER_UP, retired=(TEL_RET,), kstates={TEL_UP: 'ip-focus'})), PH)

pathlib.Path('/tmp/s19/boards.json').write_text(json.dumps(S19_OUT + [{k: b[k] for k in ('stem', 'w', 'h')} for b in out]))

if CANVAS:  # S19 row: the S19 desktops, these desktops, then the S19 phone
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    old = [b['stem'] for b in S19_OUT]
    x = CR.place(c, STEM, old[:9] + [(b['stem'], b['w'], b['h'], b['title']) for b in out] + old[9:], 0)
    CR.save(ROOT, c)
    print('canvas: S19 row y', CR.row_y(c, STEM), 'ends at x', x - 80)
