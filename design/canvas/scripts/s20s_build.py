"""S20S · Ghiduri · state boards (spec B2; D76, D113, D114). Imports s20_build (which rewrites the 9 S20 boards, and through
s19_build the S19/S18/S17/S14S/S2/S7 boards, byte-identical) and draws the S20 states on the same row (y 35576).
Desktops first (a Setări page, as S17S–S19S): the new desktops follow the eight S20 desktops; the S20 phone moves right
(x only). Today = 16.11.2026 (D23/D24). Real: the guide figures (guide-growth.json, guide-laying.json), which cycles use
each guide; guide names illustrative (D22). Every file, name or edit not in the extracts is a D39 example, said in the
board title. The import states are SYS11's GuideImportPreview (additions.css §29), composed from its helpers.
Boards
  --se-incarca            header kept (no guide names in the meta yet), framed table skeleton (§42 rows), hidden status
  --eroare-la-incarcare   header kept, fault Banner + Reîncearcă in place of the table; Importă ghid stays
  --fara-acces            Observator Sidebar, EmptyState "Nu ai acces la Ghiduri." + "Înapoi la Acasă"
                          (a manager or a vet: same picture with their own nav and role sentence; one board, as S17S–S19S)
  --niciun-ghid           D39 first use: no guide (and no series), EmptyState with the one next step "Importă primul ghid"
  --se-citeste            D39 file: Înlocuiește Ross 308 creștere, the .xlsx being read (SYS11 §42 rows sized like the columns)
  --format-gresit         D39 file: Ross 308 creștere.pdf → fault Banner + FileDrop
  --necitibil             D39 file: Ross 308 creștere.csv without the SAPT/Sex header
  --gol                   D39 file: Ross 308 creștere.csv with the header only (EmptyState "fără date", its one key)
  --refuzate-filtru       D39 rows 7/10/12: filter Refuzate, focus on row 7's "#REF!" cell (reached from the Banner link)
  --nume-gol              Importă ghid, Salvează ghidul pressed with nothing filled: Nume (focus) + Tip errors (§41)
  --nume-folosit          D39: "ross 308 ouat" typed: the name is taken by Ross 308 ouat (case and diacritics ignored, D114)
  --ghid-nou              D39 name + file: "Ross 308 ouat 2027" (Tip Ouat), the real 94 laying rows read, all valid
  --eroare-la-salvare     D39: the same new guide, save failed: draft kept, fault Banner + Reîncearcă (focus) above the keys
  --sterge                D39: Ross 308 ouat 2027 saved and used by no cycle (s20_build detail_page(name=, cycles=[])):
                          compact Folosit de line, Șterge ghidul Dialog, focus on Anulează
No phone detail (D114). No screen CSS of its own (the load-error Banner with Reîncearcă is additions.css §49, FIN03).
python3 -I s20s_build.py <repo_root> [--canvas]   after `s20_build --canvas` (canvas_rows.py)
Renders: s20_render.py <stems> (the new boards are appended to /tmp/s20/boards.json: shots S20-9 … S20-22)."""
import sys, json, pathlib, importlib.util, io, contextlib

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve(); HERE = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
sys.argv = ['s20_build.py', str(ROOT)]
sp = importlib.util.spec_from_file_location('s20_build', HERE / 's20_build.py')
S = importlib.util.module_from_spec(sp)
with contextlib.redirect_stdout(io.StringIO()):
    sp.loader.exec_module(S)   # rewrites the 9 S20 boards byte-identical (no --canvas)
M, T, G, s4, Q = S.M, S.T, S.G, S.s4, S.M.Q
icon, lamp, num, NB = S.icon, S.lamp, S.num, S.NB
GUIDE, CYC = S.GUIDE, S.CYC
S20_OUT = [{k: b[k] for k in ('stem', 'w', 'h')} for b in S.out]
STEM, TT, D39 = S.STEM, S.TT, S.D39
assert [b['stem'] for b in S20_OUT][-1] == STEM + '--telefon' and len(S20_OUT) == 9

# ================================================================= shared parts
ERRI = M.ERRI
def err(i, text): return f'<p class="ip-err ip-fieldmsg" id="{i}">{ERRI}<span>{text}</span></p>'
def one(s, a, z):
    assert s.count(a) == 1, a
    return s.replace(a, z)
META0 = 'Ghidurile de creștere și de ouat, pe săptămâni și sexe'   # before the list is in, the meta names no guide
assert S.LIST_META.startswith(META0)
HDR0 = S.header(S.crumbs('Ghiduri'), 'Ghiduri', META0, S.list_key())
def lshell(main, sidebar=None, hdr=HDR0):
    return f'<div class="ip-shell">{sidebar or S.sidebar()}<main class="ip-shell__main">\n{hdr}\n{main}\n</main></div>'

# ---- loading: the Ghiduri block with a framed table skeleton (the real columns, two guide rows; additions §42)
def skel(w, end=False): return f'<span class="ip-skel ip-skel--{w}{" ip-skel--end" if end else ""}"></span>'
SK_R = (f'<tr class="ip-table__skel"><th scope="row">{skel("w2")}</th>' + ''.join(f'<td>{skel(w)}</td>' for w in ('w3', 'w3', 'w1', 'w1'))
        + f'<td class="s19-act">{skel("w3", True)}</td></tr>')
LOADING = (f'<p class="ip-sr" role="status">Se încarcă ghidurile, cu tipul, săptămânile și ciclurile care le folosesc.</p>'
           f'<section class="ip-block" aria-labelledby="h-gh" aria-busy="true"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-gh">Ghiduri</h2></div>'
           f'<div class="ip-tablescroll ip-tablescroll--framed" aria-hidden="true"><table class="ip-table ip-table--rowheads s20-table">'
           f'{S.LCOLS}<tbody>{SK_R}{SK_R}</tbody></table></div></section>')
# ---- load error: the Banner in place of the table; Importă ghid stays (the name check runs on save)
LOADERR = (f'<section class="ip-block" aria-labelledby="h-gh"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-gh">Ghiduri</h2></div>'
           f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{icon("alert", 16)}'
           f'<div class="ip-banner__body"><p class="ip-m0">Ghidurile nu s-au încărcat.</p>'
           '<p class="ip-m0">Verifică conexiunea la internet, apoi apasă Reîncearcă. Poți importa un ghid nou și acum.</p></div>'
           '<button type="button" class="ip-btn ip-btn--secondary ip-banner__key">Reîncearcă</button></div></section>')
# ---- no access (Observator; a manager or a vet: same picture, their own nav and role sentence)
def noaccess():
    step = f'<a class="ip-btn ip-btn--secondary" href="#acasa">{icon("home", 16)}Înapoi la Acasă</a>'
    box = s4.empty_box('fără acces', ['Nu ai acces la Ghiduri.',
                                      'Ghidurile le importă și le înlocuiește proprietarul. Rolul tău, Observator, vede comparația '
                                      'cu ghidul pe paginile fermelor și în rapoarte.'], step, 'Acces')
    return f'<div class="ip-shell">{Q.sidebar_viewer()}<main class="ip-shell__main">\n<header class="ip-pagehead"><div><h1>Ghiduri</h1></div></header>\n{box}\n</main></div>'
# ---- no guide yet (D39 first use, as S19 --fara-ferme): no series either; the EmptyState holds the one next step
def sidebar_first():
    sb = M.sidebar_owner(no_series=True)
    sb = one(sb, 'class="ip-nav ip-nav--sub is-current" href="#ferme" aria-current="page"', 'class="ip-nav ip-nav--sub" href="#ferme"')
    sb = one(sb, 'class="ip-nav ip-nav--sub" href="#ghiduri"', 'class="ip-nav ip-nav--sub is-current" href="#ghiduri" aria-current="page"')
    return sb
NOGUIDE = s4.empty_box('fără ghiduri', ['Nu există niciun ghid.',
                                        'Importă ghidul rasei, de creștere sau de ouat, din .xlsx, .csv sau tabel lipit. '
                                        'Fermele îl aleg apoi în Ferme și secțiuni.'],
                       f'<a class="ip-btn ip-btn--primary" href="#importa-ghid">{icon("upload", 16)}Importă primul ghid</a>', 'Ghiduri')
HDR_FIRST = S.header(S.crumbs('Ghiduri'), 'Ghiduri', META0)

# ---- Înlocuiește Ross 308 creștere: the file states (Tip fixed by the cycles, so the file is read as creștere)
g = 'g-growth'; GF = S.GF; P = 'gi'
def repl(gimp, rows=len(S.gr)):
    return S.import_page(S.crumbs(S.GH, (S.slug(g), GUIDE[g]), 'Înlocuiește'), f'Înlocuiește {GUIDE[g]}', S.detail_meta(g),
                         S.fields(GUIDE[g], S.tip_ro(g)), S.repl_note(g, rows), gimp)
def wrap(inner): return G['wrap'](P, inner, 'Import ghid')
# SYS11's loading rows (six §42 skeleton rows, aria-busy table): each bar as wide as the column's figures, at their edge
READING = wrap(G['src_line'](P, GF, 'fișier .xlsx · tip creștere', '<span class="ip-gimp__wait">Se citește fișierul.</span>', 'Anulează citirea')
               + G['bar'](P) + G['table'](P, 'growth', [], GF, busy=True, body=G['skel_body']('growth'))
               + G['foot'](P, False, 'Se citește fișierul.') + G['live'](f'Se citește {GF}.'))
def fault(bid, html):   # a file fault: the Banner is announced and describes the FileDrop that takes the next file
    b = one(G['banner']('fault', html, bid), f'id="{bid}"', f'id="{bid}" role="alert"')
    d = one(S.DROP_P, 'role="group"', f'role="group" aria-describedby="{bid}"')
    return b + d
GPDF, GCSV = 'Ross 308 creștere.pdf', 'Ross 308 creștere.csv'
WRONG = wrap(fault('gi-err', f'<p><b>{GPDF} nu se poate citi.</b> Încarcă un fișier .xlsx sau .csv, sau lipește tabelul.</p>')
             + G['foot'](P, False, 'Alege un fișier .xlsx sau .csv.'))
NOHEAD = wrap(fault('gi-err', G['NOHEAD'].format(f=GCSV))
              + G['foot'](P, False, 'Alege un fișier cu antet.'))
EMPTYF = wrap(G['src_line'](P, GCSV, 'fișier .csv · tip creștere', G['counts'](0, 0, 0), '')   # one "Alege alt fișier": the EmptyState's
              + G['empty_file'](GCSV) + G['foot'](P, False, G['EMPTY_NOTE']) + G['live']('Fișier citit: niciun rând.'))
er, nbad = S.er, S.nbad
assert [r['line'] for r in er if G['reason'](r)] == [7, 10, 12]
FILTER = wrap(G['src_line'](P, GF, 'fișier .xlsx · tip creștere', G['counts'](len(er), len(er) - nbad, nbad))
              + G['err_banner'](P, er) + G['info_banner'](S.INFO_E) + G['seg'](P, len(er), nbad, 'bad')
              + G['table'](P, 'growth', [r for r in er if G['reason'](r)], GF, focus={7: 'body_weight_g'})
              + G['foot'](P, False, f'{nbad} rânduri refuzate: corectează fișierul și încarcă-l din nou.')
              + G['live'](f'Se arată doar rândurile refuzate: {nbad} din {len(er)}.'))
assert 'id="gi-r7-body_weight_g"' in FILTER and 'href="#gi-r7-body_weight_g"' in FILTER

# ---- Importă ghid (new): field errors, the new name, the failed save
def name_err(val, e):
    f = T.tfield('gi-name', 'Nume', val, touch=False, focus=True)
    f = one(f, 'class="ip-input is-focus"><input id="gi-name" type="text" autocomplete="off"',
            'class="ip-input is-focus is-error"><input id="gi-name" type="text" autocomplete="off" aria-invalid="true" aria-describedby="gi-name-e"')
    assert f.endswith('</div>')
    return f[:-6] + err('gi-name-e', e) + '</div>'
def tip_err():
    f = one(M.tip_field(), 'aria-required="true">', 'aria-required="true" aria-invalid="true" aria-describedby="gi-tip-e">')
    return f[:-6] + err('gi-tip-e', 'Alege tipul ghidului: Creștere sau Ouat. De el depinde cum se citește tabelul.') + '</div>'
def flds(*f): return f'<div class="s20-fields">{"".join(f)}</div>'
NEW_FOOT = 'Scrie numele, alege tipul și adaugă tabelul ghidului.'
def new_page(fl, gimp, note=''):
    return S.import_page(S.crumbs(S.GH, 'Importă ghid'), 'Importă ghid', 'Un ghid nou, din fișier .xlsx sau .csv sau din tabelul lipit',
                         fl, note, gimp)
EMPTY_IMPORT = wrap(S.DROP_P + G['foot'](P, False, NEW_FOOT))
TAKEN = GUIDE['g-laying']
NEWN = 'Ross 308 ouat 2027'; NEWF = f'{NEWN}.xlsx'
lr = S.lr; NL = len(lr)
INFO_LF = one(S.INFO_L, 'iar tabelul nu are', 'iar fișierul nu are')
NEW_NOTE = 'Un ghid nou nu e folosit de nicio fermă: după salvare îl alegi la o fermă de ouat, în Ferme și secțiuni.'
NEW_FLDS = flds(T.tfield('gi-name', 'Nume', NEWN, touch=False), S.tip_seg(sel='laying'))
def new_read(extra=''):
    return wrap(G['src_line'](P, NEWF, 'fișier .xlsx · tip ouat', G['counts'](NL, NL, 0))
                + G['banner']('ok', f'<p><b>Toate cele {NL} de rânduri sunt valide.</b> Verifică cifrele și salvează ghidul.</p>')
                + G['info_banner'](INFO_LF) + G['bar'](P) + G['table'](P, 'laying', lr, NEWF) + extra
                + G['foot'](P, True, f'Se salvează {NL} de rânduri în ghidul nou {NEWN}.')
                + G['live'](f'Fișier citit: {NL} de rânduri, toate valide.'))
SAVE_ERR = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert" id="gi-save-err">{icon("alert", 16)}<div class="ip-banner__body">'
            f'<p class="ip-m0">{NEWN} nu s-a salvat: conexiunea s-a întrerupt. Verifică conexiunea la internet, apoi apasă Reîncearcă. '
            f'Numele, tipul și cele {NL} de rânduri citite rămân aici.</p>'
            '</div><button type="button" class="ip-btn ip-btn--secondary ip-banner__key ip-focus">Reîncearcă</button></div>')
SAVE_FOCUS = ('<!-- Focus after a failed Salvează ghidul: Reîncearcă in the fault Banner (role=alert) above the keys; the name, '
              'the type and the rows read stay as typed and read. -->')

# ---- Șterge ghidul (D39): the new guide, saved and used by no cycle; the Dialog over the inert detail page
def unused_detail():   # the detail of a laying guide with the new name and no cycle (s20_build parameters)
    pg = S.detail_page('g-laying', name=NEWN, cycles=[], pressed=True)
    assert 'azi</span>' not in pg and 'e azi în sapt.' not in pg and 'Ross 308 ouat<' not in pg
    return one(pg, '<div class="ip-shell">', '<div class="ip-shell" inert="inert">')
DLG = (f'<div class="ip-scrim ip-scrim--page"><div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="dlg-t" aria-describedby="dlg-n">'
       f'<h2 id="dlg-t">Șterge ghidul {NEWN}</h2>'
       f'<p class="ip-dialog__note" id="dlg-n">Ghidul și cele {NL} de rânduri ale lui se șterg. Niciun ciclu nu îl folosește, deci nicio comparație '
       f'nu se schimbă. Dacă îl vrei înapoi, îl imporți din nou din {NEWF}.</p>'
       f'<div class="ip-dialog__keys"><button type="button" class="ip-btn ip-btn--secondary ip-focus" autofocus="autofocus">Anulează</button>'
       f'<button type="button" class="ip-btn ip-btn--destructive">Șterge ghidul</button></div></div></div>')

# ================================================================= CSS (screen only, tokens only)
CSS20S = ''   # FIN03: the Banner action key is additions.css §49 (.ip-banner__key); no screen CSS left
CSS_D = S.CSS_DESK + CSS20S

# ================================================================= write
out = []
def desk(stem, title, root, h=None):
    h = h or S.HEIGHTS.get(STEM + stem, 1024)
    out.append(S.write(STEM + stem, f'{TT} · {title}', f'<div class="ip-root s4s-rel" style="padding:0; min-height:{h}px">{root}</div>', CSS_D, 1440, h))

desk('--se-incarca', 'se încarcă', lshell(LOADING))
desk('--eroare-la-incarcare', 'eroare la încărcare', lshell(LOADERR))
desk('--fara-acces', 'observator · fără acces', noaccess())
desk('--niciun-ghid', f'niciun ghid · {D39}: prima folosire', lshell(NOGUIDE, sidebar_first(), HDR_FIRST))
desk('--se-citeste', f'Înlocuiește Ross 308 creștere · se citește fișierul · {D39}: fișierul', repl(READING))
desk('--format-gresit', f'Înlocuiește Ross 308 creștere · format greșit · {D39}: „{GPDF}”', repl(WRONG))
desk('--necitibil', f'Înlocuiește Ross 308 creștere · fișier necitibil, fără antet · {D39}: „{GCSV}”', repl(NOHEAD))
desk('--gol', f'Înlocuiește Ross 308 creștere · fișier gol, doar antetul · {D39}: „{GCSV}”', repl(EMPTYF))
desk('--refuzate-filtru', f'Înlocuiește Ross 308 creștere · filtru Refuzate, focus pe celula „#REF!” · {D39}: rândurile 7, 10, 12',
     repl(FILTER, len(er)))
desk('--nume-gol', 'Importă ghid · nume gol și tip lipsă',
     new_page(flds(name_err('', f'Scrie numele ghidului, de exemplu {NEWN}.'), tip_err()), EMPTY_IMPORT))
desk('--nume-folosit', f'Importă ghid · nume deja folosit · {D39}: „ross 308 ouat”',
     new_page(flds(name_err('ross 308 ouat', f'Numele e deja folosit de ghidul {TAKEN} (literele mari și diacriticele nu contează). '
                                             f'Scrie alt nume, de exemplu cu anul ediției, sau deschide {TAKEN} și apasă Înlocuiește ghidul.'),
                   S.tip_seg(sel='laying')), EMPTY_IMPORT))
desk('--ghid-nou', f'Importă ghid · ghid nou, fișier citit, toate valide · {D39}: numele „{NEWN}” și fișierul', new_page(NEW_FLDS, new_read(), NEW_NOTE))
desk('--eroare-la-salvare', f'Importă ghid · eroare la salvare · {D39}: ghidul nou „{NEWN}”',
     new_page(NEW_FLDS, new_read(SAVE_FOCUS + SAVE_ERR), NEW_NOTE))
desk('--sterge', f'Șterge ghidul nefolosit · {D39}: „{NEWN}”, salvat și nefolosit',
     f'<div class="s20-clip">{unused_detail()}</div>' + DLG, M.SC.SIDE_H)

p = pathlib.Path('/tmp/s20/boards.json')
p.write_text(json.dumps(S20_OUT + [{k: b[k] for k in ('stem', 'w', 'h')} for b in out]))

if CANVAS:  # S20 row: the S20 desktops, these desktops, then the S20 phone
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    old = [b['stem'] for b in S20_OUT]
    x = CR.place(c, STEM, old[:8] + [(b['stem'], b['w'], b['h'], b['title']) for b in out] + old[8:], 0)
    CR.save(ROOT, c)
    print('canvas: S20 row y', CR.row_y(c, STEM), 'ends at x', x - 80)
