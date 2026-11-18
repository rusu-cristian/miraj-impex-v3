"""S22 · Utilizatori (spec A2 Users and access, owner; screens.md S22; nav map: Setări › Utilizatori). Desktop-led like
S18–S21 (a Setări page): desktops first, then the phone. Today = 16.11.2026 (D23/D24).
The list is the SYS12 UserRow (additions.css §30) and the passwords are the SYS12 OneTimePassword: their markup helpers
are read from design/system/scripts/sys12_build.py (the part above its preview blocks, so nothing is written there).
D78 (departs from spec A2's field list): the app generates the temporary password and shows it once after the user is
added (or after Resetează parola); the add form has no password field. D118: the S22 screen calls (logged in decisions).
People: Proprietar (you, the owner; no name in the kit, D96/D104), Ion Rusu (manager), Dr. Ceban (vet) as on the earlier
boards (craft §7, illustrative); Petrea (spec A2's dialog sample) and Observator exemplu are D39 state examples (SYS12),
named in the board titles, as are every sign-in time and both passwords. Farms real (farms-sections.json).
Boards (row below S21: y 38052 + 1024 + 120 = 39196; x step w + 80; desktops first, then the phone):
  S22-utilizatori          list: PageHeader (Setări › Utilizatori, "Adaugă utilizator"), UserRow list under an h2 + count
  --ion-rusu               list beside the non-modal user SidePanel: Nume (focus), Login read-only, Rol Select + what the
                           role does, Ferme checkboxes, Cont block (Resetează parola, Dezactivează contul)
  --utilizator-nou         empty panel: Nume (focus), Login, Rol unset, Ferme unticked, the generated-password note
  --erori                  D39: Adaugă utilizatorul with Nume empty, Login with a space and diacritics, no Rol, no farm
  --login-folosit          D39: Petrea back, Petrea's deactivated account still holds „petrea” (A2 duplicate login)
  --parola-cont-nou        D39: Observator exemplu added: the list with the new row under the OneTimePassword Dialog
  --salvat                 D39: Dr. Ceban given Telița too: panel closed, row updated, Toast + Anulează
  --eroare-la-salvare      D39: the same change, save failed: fault Banner + Reîncearcă (focus), draft kept
  --contul-tau             your own account: Rol read-only (the only active owner), Ferme "toate", Dezactivează disabled
  --dezactiveaza           D39: Dezactivează contul lui Petrea Dialog (A2 sentence), focus on Anulează
  --dezactivat             D39: Petrea listed last, off lamp, Toast + Anulează, focus back on his key
  --parola-resetata        D39: Resetează parola from Ion Rusu's panel: the new password, shown once
  --se-incarca             SYS12 loading rows (aria-busy), header without counts
  --eroare-la-incarcare    SYS12 load failed (fault Banner + Reîncearcă, no table)
  --fara-acces             Observator (a manager or a vet: same picture, own nav and role sentence)
  --doar-tu                D39 first use: your row + "Doar tu ai acces. Adaugă primul utilizator." (A2 Empty)
  --telefon                phone, read-only list ("îi adaugi și îi editezi pe calculator", as S19–S21)
The list region is a size container: under 760px (beside the panel: 558 at 1280, 718 at 1440) Login, Ferme and Ultima
intrare fold under the name so Rol, Stare and the Acțiuni key stay in view (S19P rule).
S22P (D120): while a panel is open the list-head key is secondary (one primary per view), the user open in the panel is
the DataTable current row (aria-current on the row's Acțiuni key, additions.css §45), the Cont note says each thing once,
and the phone list adds the last sign-in line (Mono).
FIN07 (D127): the page follows the S19–S21 list page: "Adaugă utilizator" is the page head's secondary key (pressed and
expanded while its own add panel is open), the list head is an h2 "Utilizatori" with the count on the right (ip-sechead),
the panel key reads "Salvează utilizatorul", and loading draws the §42 DataTable skeleton rows (SYS12 SK).
python3 -I s22_build.py <repo_root> [s22_heights.json] [--canvas]
Imports s19_build (rewrites the S19 boards and, through s18_build, S18/S17/S14S/S2/S7, byte-identical). Writes
design/canvas/project/S22-utilizatori[--<state>].dc.html, /tmp/s22/<stem>.html, /tmp/s22/boards.json; --canvas adds the
row to canvas.json (nothing else changes). Renders: s22_render.py [--w 1280] [stems]."""
import sys, re, json, pathlib, importlib.util, io, contextlib

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve(); HERE = pathlib.Path(__file__).resolve().parent
HP = pathlib.Path(ARGS[1]) if len(ARGS) > 1 else HERE / 's22_heights.json'
HEIGHTS = json.loads(HP.read_text()) if HP.exists() else {}
sys.path.insert(0, str(HERE))
import skel as SKL
sys.argv = ['s19_build.py', str(ROOT)]
sp = importlib.util.spec_from_file_location('s19_build', HERE / 's19_build.py')
M = importlib.util.module_from_spec(sp)
with contextlib.redirect_stdout(io.StringIO()):
    sp.loader.exec_module(M)   # rewrites the S19 boards (and through s18_build S18/S17/S14S/S2/S7) byte-identical
T, SC, Q, s4, A = M.T, M.SC, M.Q, M.s4, M.A
icon, lamp, num, NB = M.icon, M.lamp, M.num, M.NB

# SYS12 helpers (UserRow + OneTimePassword): the code above the preview blocks only (no file is written)
SRC = (ROOT / 'design/system/scripts/sys12_build.py').read_text(encoding='utf-8')
CUT = '# ------------------------------------------------------------------ blocks'
assert SRC.count(CUT) == 1
U = {'__name__': 'sys12_parts', '__file__': str(ROOT / 'design/system/scripts/sys12_build.py')}  # sys12 finds the repo from its own path
exec(compile(SRC.split(CUT)[0], 'sys12_build.py', 'exec'), U)
# the loading rows and the empty row are SYS12 preview blocks (below the cut): read the same markup from its source
BLK = SRC.split(CUT)[1]
m = re.search(r"\nEMPTY = \((.*?)\)\n", BLK, re.S); U['EMPTY'] = eval('(' + m.group(1) + ')')
m = re.search(r"\nSK = (''\.join\(.*?for _ in range\(2\)\))\n", BLK, re.S); U['SK'] = eval(m.group(1), {})
assert 'Doar tu ai acces. Adaugă primul utilizator.' in U['EMPTY'] and U['SK'].count('class="ip-table__skel"') == 2

# ================================================================= people (kit names; D39 examples marked in the titles)
FS = M.FS
assert [f['name'] for f in FS['farms']] == ['Telița', 'Hîrbovăț']   # real farms; listed in this order everywhere
assert U['RUSU']['farms'] == 'Telița, Hîrbovăț' and U['CEBAN']['farms'] == 'Hîrbovăț'
OWN = {**U['OWNER'], 'name': 'Proprietar'}   # the owner account as on every earlier board (sidebar "Proprietar", D104)
RUSU, CEBAN = dict(U['RUSU']), dict(U['CEBAN'])
PETREA = {**U['PETREA'], 'active': True}     # D39: Petrea (spec A2 sample) before his account is deactivated
PETREA_OFF = dict(U['PETREA'])
OBS = dict(U['VIEW'])                        # D39: the new user, Observator exemplu (SYS12)
CEBAN2 = {**CEBAN, 'farms': 'Telița, Hîrbovăț'}   # D39: the edit saved on --salvat
ROLE_ORDER = ['Proprietar', 'Manager', 'Veterinar', 'Observator']
def order(us):   # SYS12: active rows by role then name, deactivated rows last
    return sorted(us, key=lambda u: (not u['active'], ROLE_ORDER.index(u['role']), u['name']))
BASE = order([OWN, RUSU, CEBAN, PETREA])
assert [u['name'] for u in BASE] == ['Proprietar', 'Ion Rusu', 'Dr. Ceban', 'Petrea']
OFFL = order([OWN, RUSU, CEBAN, PETREA_OFF])
assert OFFL[-1]['name'] == 'Petrea' and not OFFL[-1]['active']
# D78 passwords: 12 characters, a–z without i l o + 2–9; plainly sample values (alphabet runs), D39
PW_NEW, PW_RESET = 'abcd2345efgh', 'mnpq6789rstu'
ALPHA = set('abcdefghjkmnpqrstuvwxyz23456789')
assert all(len(p) == 12 and set(p) <= ALPHA for p in (PW_NEW, PW_RESET))
assert U['CTX_NEW'].startswith('Cont nou · Observator · Hîrbovăț')

# ================================================================= list (desktop): SYS12 UserRow + the S19P fold
CRUMBS = ('<nav class="ip-crumbs" aria-label="Cale"><a href="#setari">Setări</a><span aria-hidden="true">›</span>'
          '<span aria-current="page">Utilizatori</span></nav>')
META = '<div class="ip-row ip-pagehead__meta"><span class="ip-note">Cine intră în aplicație, cu ce rol și la ce ferme</span></div>'
def newkey(pressed=False):
    """FIN07 (D127): the page head's add key, secondary as S18 Adaugă serie, S19 Adaugă fermă, S21 Adaugă tip; pressed and
    expanded while its own add panel is open (D120/D125), so the panel's key is the view's one primary."""
    p = ' is-pressed" aria-expanded="true" aria-controls="up' if pressed else ''
    return f'<button type="button" class="ip-btn ip-btn--secondary{p}">{icon("plus", 16)}Adaugă utilizator</button>'
def header(pressed=False):
    return f'<header class="ip-pagehead">\n<div>\n{CRUMBS}\n<h1>Utilizatori</h1>\n{META}\n</div>\n<div class="ip-row">{newkey(pressed)}</div>\n</header>'
def sidebar():
    sb = M.sidebar_owner()
    for a, z in (('class="ip-nav ip-nav--sub is-current" href="#ferme" aria-current="page"', 'class="ip-nav ip-nav--sub" href="#ferme"'),
                 ('class="ip-nav ip-nav--sub" href="#utilizatori"', 'class="ip-nav ip-nav--sub is-current" href="#utilizatori" aria-current="page"')):
        assert sb.count(a) == 1, a; sb = sb.replace(a, z)
    assert sb.count('is-current') == 1
    return sb
def sub(s, a, z, n=1):
    assert s.count(a) == n, (a, s.count(a)); return s.replace(a, z)
CX = ' s22-cx'   # columns that fold under the name beside the panel
def fold(u):
    fm = u['farms'] or 'toate fermele'
    seen = f'<span class="ip-num ip-nowrap">{u["seen"]}</span>' if u['seen'] else '<span class="ip-nowrap">– încă n-a intrat</span>'
    return f'<span class="s22-fold"><span class="ip-num">{u["login"]}</span> · {fm} · {seen}</span>'
def row(p, u, focus=False, cur=False):
    """The SYS12 row, unchanged except: the name cell carries the fold line, three cells the fold class; the row whose
    user is open in the panel marks its Acțiuni key aria-current (additions.css §45, the DataTable current row)."""
    r = U['row'](p, u, focus=focus)
    if cur:
        r = sub(r, ' aria-haspopup="menu"', ' aria-current="true" aria-haspopup="menu"')
    r = sub(r, '<td><span class="ip-user__name">', '<td class="s22-name"><span class="ip-user__name">')
    r = re.sub(r'(<span class="ip-user__name">.*?</span>)</td>', lambda m: m.group(1) + fold(u) + '</td>', r, count=1)
    for c in ('ip-user__login', 'ip-user__farms', 'ip-user__seen'):
        r = sub(r, f'<td class="{c}">', f'<td class="{c}{CX}">')
    return r
HEAD = U['HEAD']
for h in ('Login', 'Ferme', 'Ultima intrare'):
    HEAD = sub(HEAD, f'<th scope="col">{h}</th>', f'<th scope="col" class="s22-cx">{h}</th>')
# FIN11G (§54): the desktop list's column shares at 1440 (as content-sized before) on the head the loading skeleton reuses;
# None = the column that takes the rest
HEAD = SKL.widths(HEAD, [None, 15.4, 12.8, 17.6, 23.6, 9.9, 3.5], [89.8, 78.28, 61.08, 92.48, 133.06, 41.75, 32])  # + loaded min-content px (fix 1)
def users(p, body, sumline, note='', **kw):
    """SYS12 users() as an S19–S21 list block (FIN07): section.ip-block with the ip-sechead (h2 "Utilizatori" under the h1, the
    count on the right; none while loading or failed), the key moved to the page head; the list box is the fold container."""
    s = U['users'](p, body, sumline, **kw)
    k = re.search(r'<button type="button" class="ip-btn ip-btn--primary" aria-haspopup="dialog">.*?Adaugă utilizator</button>', s).group(0)
    cnt = f'<span class="ip-note" id="{p}-sum">{sumline}</span>' if sumline else ''
    s = sub(s, f'<section class="ip-users" aria-labelledby="{p}-h"><div class="ip-list__head"><h3 class="ip-cap" id="{p}-h">Utilizatori</h3>'
               f'<p class="ip-note" id="{p}-sum">{sumline}</p>{k}</div>',
            f'<section class="ip-block" aria-labelledby="{p}-h"><div class="ip-sechead"><h2 class="ip-t-heading" id="{p}-h">Utilizatori</h2>{cnt}</div>')
    if kw.get('table', True):
        if not sumline:   # no count line to describe the table
            s = sub(s, f' aria-describedby="{p}-sum"', '')
        s = sub(s, U['HEAD'], HEAD)
        s = sub(s, '<div class="ip-users__box">', '<div class="ip-users__box s22-region">')
    return s[:-len('</section>')] + note + '</section>'
NOTE = ('<p class="ip-note ip-m0">Un cont dezactivat nu mai poate intra; ce a introdus rămâne în rapoarte și în Istoric modificări. '
        'Conturile nu se șterg.</p>')
def ulist(us=BASE, focus=None, p='u', cur=None):
    """cur = the user open in the panel: their row is the current row (§45)."""
    body = ''.join(row(p, u, focus=(u['key'] == focus), cur=(u['key'] == cur)) for u in us)
    return users(p, body, U["counts"](us), NOTE)
def shell(main, sb=None, inert=False, pressed=False):
    i = ' inert="inert"' if inert else ''
    return f'<div class="ip-shell"{i}>{sb or sidebar()}<main class="ip-shell__main">\n{header(pressed)}\n{main}\n</main></div>'
def vp(main, inert=False, pressed=False):
    return f'<div class="ip-dockpage s18-vp">{shell(main, inert=inert, pressed=pressed)}</div>'

# ================================================================= user panel (SidePanel beside the page, as S19/S21)
ERRI = icon('alert', 14)
def err(i, text): return f'<p class="ip-err ip-fieldmsg" id="{i}">{ERRI}<span>{text}</span></p>'
def with_err(f, iid, eid, text):   # TextField / Select in error: fault edge, aria-invalid, the field message names the fix
    f = re.sub(r'class="ip-input( [^"]*)?"', lambda m: f'class="ip-input{(m.group(1) or "").replace(" is-error", "")} is-error"', f, count=1)
    f = re.sub(rf'(<(?:input|select) id="{iid}")', rf'\1 aria-invalid="true"', f, count=1)
    f = re.sub(r' aria-describedby="[^"]*"', '', f)
    f = sub(f, f'id="{iid}" aria-invalid="true"', f'id="{iid}" aria-invalid="true" aria-describedby="{eid}"')
    f = re.sub(r'<span class="ip-hint"[^>]*>.*?</span>', '', f)
    return f[:-6] + err(eid, text) + '</div>'
def name_field(val='', focus=False, e=None):
    f = T.tfield('up-name', 'Nume', val, hint='Cum apare în listă și în Istoric modificări; cel mult 60 de caractere.', touch=False, focus=focus)
    return with_err(f, 'up-name', 'up-name-e', e) if e else f
LOGIN_H = 'Cu el intră în aplicație: litere mici fără diacritice, cifre și punct, 3–40 caractere, de exemplu ion.rusu.'
def login_field(val='', e=None, ro=False, focus=False):
    if ro:   # D118: the login is fixed once the account exists (A2 step 3 changes role, farms, password, state)
        return (f'<div class="ip-field"><label class="ip-label" for="up-login">Login</label><span class="ip-input is-readonly ip-num">'
                f'<input id="up-login" type="text" readonly="readonly" value="{val}" aria-describedby="up-login-h"/></span>'
                f'<span class="ip-hint" id="up-login-h">Loginul nu se schimbă după ce contul e creat.</span></div>')
    f = T.tfield('up-login', 'Login', val, hint=LOGIN_H, touch=False, mono=True, focus=focus)
    f = sub(f, '<input id="up-login" type="text"', '<input id="up-login" type="text" spellcheck="false" autocapitalize="off"')
    return with_err(f, 'up-login', 'up-login-e', e) if e else f
ROLE_DO = {   # what each role does (spec §1 roles table, D03/D06); the hint under Rol
    'Veterinar': 'Introduce mortalitatea, hrana zilnică, cântăririle și ouăle la fermele bifate; nu vede prețuri.',
    'Observator': 'Vede paginile și rapoartele fermelor bifate; prețurile doar le vede; nu introduce nimic.',
    'Manager': 'Introduce tot ce introduce veterinarul, plus livrări, transferuri, vânzări și cicluri noi la fermele bifate; '
               'corectează cu motiv. Prețurile doar le vede.',
    'Proprietar': 'Vede și modifică tot, la toate fermele: prețuri, utilizatori, ferme și ghiduri.'}
ROLES = ['Veterinar', 'Observator', 'Manager', 'Proprietar']   # spec order vet / viewer / manager / owner
def role_field(sel=None, e=None, ro=None):
    if ro:   # your own role while you are the only active owner (A2 rule; D78 "must stop")
        return (f'<div class="ip-field"><label class="ip-label" for="up-role">Rol</label><span class="ip-input is-readonly">'
                f'<input id="up-role" type="text" readonly="readonly" value="{sel}" aria-describedby="up-role-h"/></span>'
                f'<span class="ip-hint" id="up-role-h">{ro}</span></div>')
    opts = ('' if sel else '<option value="" selected="selected">Alege rolul</option>') + ''.join(
        f'<option{" selected=\"selected\"" if r == sel else ""}>{r}</option>' for r in ROLES)
    hint = ROLE_DO[sel] if sel else 'De rol depinde ce vede și ce introduce.'
    f = (f'<div class="ip-field"><label class="ip-label" for="up-role">Rol</label><span class="ip-input">'
         f'<select id="up-role" required="required" aria-describedby="up-role-h">{opts}</select>{icon("chev", 16)}</span>'
         f'<span class="ip-hint" id="up-role-h">{hint}</span></div>')
    return with_err(f, 'up-role', 'up-role-e', e) if e else f
def farms_field(on=(), e=None, owner=False):
    if owner:
        return ('<div class="ip-field"><label class="ip-label" for="up-farms">Ferme</label><span class="ip-input is-readonly">'
                '<input id="up-farms" type="text" readonly="readonly" value="Toate fermele" aria-describedby="up-farms-h"/></span>'
                '<span class="ip-hint" id="up-farms-h">Proprietarul vede toate fermele, și pe cele adăugate mai târziu.</span></div>')
    d = 'up-farms-e' if e else 'up-farms-h'
    inv = ' aria-invalid="true"' if e else ''
    boxes = ''
    for f in FS['farms']:
        c = f['name'] in on
        boxes += (f'<label class="ip-check{" is-checked" if c else ""}"><input type="checkbox" class="ip-sr" id="up-f-{f["id"]}"'
                  f'{" checked=\"checked\"" if c else ""} aria-describedby="{d}"{inv}/>'
                  f'<span class="ip-check__box">{icon("check", 14) if c else ""}</span>{f["name"]}</label>')
    msg = err('up-farms-e', e) if e else '<span class="ip-hint" id="up-farms-h">Vede și introduce doar la fermele bifate; cel puțin una.</span>'
    return (f'<div class="ip-field" role="group" aria-labelledby="up-farms-l"><span class="ip-label" id="up-farms-l">Ferme</span>'
            f'<div class="s22-checks">{boxes}</div>{msg}</div>')
def acct(u, pressed=None):
    """Cont: state + last sign-in, then Resetează parola and Dezactivează contul (both open a Dialog); on your own account
    Dezactivează is aria-disabled with the SYS12 reason (A2: you can't deactivate yourself)."""
    seen = f'ultima intrare <span class="ip-num">{u["seen"]}</span>' if u['seen'] else '– încă n-a intrat'
    pr = ' is-pressed' if pressed == 'reset' else ''
    keys = f'<button type="button" class="ip-btn ip-btn--secondary{pr}" aria-haspopup="dialog">Resetează parola</button>'
    why = ''
    note = ('Resetează parola dacă a uitat-o: cea veche nu mai merge, iar cea nouă se arată o singură dată. '
            'Un cont dezactivat se reactivează din listă.')
    if u.get('me'):
        note = 'Resetează-ți parola dacă ai uitat-o: cea veche nu mai merge, iar cea nouă se arată o singură dată.'
        keys += '<button type="button" class="ip-btn ip-btn--destructive" aria-disabled="true" aria-describedby="up-off-why">Dezactivează contul</button>'
        why = '<p class="ip-note ip-m0" id="up-off-why">Nu îți poți dezactiva propriul cont.</p>'
    else:
        keys += '<button type="button" class="ip-btn ip-btn--destructive" aria-haspopup="dialog">Dezactivează contul</button>'
    return (f'<section class="s22-acct" aria-labelledby="up-acct-h"><h3 class="ip-cap ip-m0" id="up-acct-h">Cont</h3>'
            f'<p class="ip-m0 s22-state">{lamp("ok", "activ")}<span>{seen}</span></p>'
            f'<p class="ip-note ip-m0">{note}</p>'
            f'<div class="s22-keys">{keys}</div>{why}</section>')
NEWPW = ('<section class="s22-acct" aria-labelledby="up-acct-h"><h3 class="ip-cap ip-m0" id="up-acct-h">Parolă</h3>'
         '<p class="ip-m0">Parola temporară se generează la salvare și se arată o singură dată.</p></section>')
SAVE = '<button type="button" class="ip-btn ip-btn--primary">Salvează utilizatorul</button>'   # FIN07: as Salvează ferma / Salvează tipul
ADD = '<button type="button" class="ip-btn ip-btn--primary">Adaugă utilizatorul</button>'
def ctx(u):
    fm = (u['farms'] or 'toate fermele').upper()
    return f'{u["role"].upper()} · {"ACTIV" if u["active"] else "DEZACTIVAT"} · {fm}'
def panel(title, note, body, key=SAVE, inert=False):
    i = ' inert="inert"' if inert else ''
    nt = f'<p class="ip-note ip-num ip-m0">{note}</p>' if note else ''
    return (f'<aside class="ip-sidepanel ip-sidepanel--dock" id="up" aria-labelledby="up-t"{i}>\n<div class="ip-sidepanel__head"><div>'
            f'<h2 class="ip-t-heading ip-m0" id="up-t">{title}</h2>{nt}</div>'
            f'<button type="button" class="ip-iconbtn" aria-label="Închide panoul utilizatorului">{icon("x")}</button></div>\n'
            f'<div class="ip-sidepanel__body ip-sidepanel__body--form s19-pbody">\n{"".join(body)}\n</div>\n'
            f'<div class="ip-sidepanel__foot s18-foot"><div class="ip-row ip-sidepanel__keys"><button type="button" class="ip-btn ip-btn--secondary">Anulează</button>{key}</div></div>\n</aside>')
def upanel(u, farms_on=None, focus=False, extra='', inert=False, pressed=None, title=None):
    fo = (u['farms'] or '').split(', ') if farms_on is None else farms_on
    return panel(title or u['name'] + (' (tu)' if u.get('me') else ''), ctx(u),
                 [name_field(u['name'], focus), login_field(u['login'], ro=True), role_field(u['role']), farms_field(fo), extra, acct(u, pressed)],
                 inert=inert)
def save_banner(*paras):   # D147: the lead, then the retry line + what is kept
    body = ''.join(f'<p class="ip-m0">{t}</p>' for t in paras)
    return (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert" id="up-err">{icon("alert", 16)}<div class="ip-banner__body">'
            f'{body}</div><button type="button" class="ip-btn ip-btn--secondary ip-banner__key ip-focus">Reîncearcă</button></div>')

# ================================================================= dialogs (SYS12), Toast, states
def scrim(dlg): return f'<div class="ip-scrim ip-scrim--page">{dlg}</div>'
BTN = U['btn']
def dlg_off(u):   # A2: the sentence, then SYS12's undo line; focus on Anulează
    body = f'<p id="do-n">{u["name"]} nu va mai putea intra. Datele introduse de el rămân. Poți reactiva contul oricând.</p>'
    return U['dialog']('do', f'Dezactivează contul lui {u["name"]}', body,
                       BTN('Anulează', focus=True, extra=' autofocus="autofocus"') + BTN('Dezactivează', 'destructive'), desc='do-n')
def dlg_pw(p, title, ctx_, pw, name):   # OneTimePassword in its Dialog; focus on Copiază parola (SYS12)
    return U['dialog'](p, title, U['otp'](p, ctx_, pw, name, focus='copy'), BTN('Închide', 'primary'))
FOCUS_NOTE = ('<!-- Focus after {what}: {where}; the Toast (role=status) announces it and keeps Anulează in the tab order. -->')
def toast(text, what, where):
    return (FOCUS_NOTE.format(what=what, where=where) + f'<div class="ip-toastwrap ip-toastwrap--page"><div class="ip-toast" role="status">{icon("check", 16)}'
            f'<span class="ip-toast__txt">{text}</span><button type="button" class="ip-link ip-toast__undo">Anulează</button></div></div>')
# FIN07: as S19S/S21 se încarcă: a hidden status line, the block aria-busy, the box aria-hidden with the real head and the
#   two §42 skeleton rows (SYS12 SK); the list head names no count
LOADING = ('<p class="ip-sr" role="status">Se încarcă utilizatorii, cu rolul și fermele lor.</p>'
           + sub(sub(users("u", U["SK"], ""), '<section class="ip-block" aria-labelledby="u-h">',
                     '<section class="ip-block" aria-labelledby="u-h" aria-busy="true">'),
                 '<div class="ip-users__box s22-region">', '<div class="ip-users__box s22-region" aria-hidden="true">'))
# FIN03 (D123): the Banner with action (additions.css §49), role=alert and the S19–S21 copy; the alert announces it, so the
#   list's status line stays empty (it repeated the Banner's first sentence)
FAIL = (f'<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">{U["svg"](U["IC"]["fault"])}<div class="ip-banner__body">'
        '<p>Utilizatorii nu s-au încărcat.</p><p>Verifică conexiunea la internet, apoi apasă Reîncearcă. Poți adăuga un utilizator nou și acum.</p></div>'
        + sub(BTN("Reîncearcă", "secondary"), 'class="ip-btn ip-btn--secondary"', 'class="ip-btn ip-btn--secondary ip-banner__key"') + '</div>')
LOADERR = users("u", "", "", extra=FAIL, table=False)   # FIN07: as S19S/S21, the list head names no count; the alert says it
def noaccess():
    step = f'<a class="ip-btn ip-btn--secondary" href="#acasa">{icon("home", 16)}Înapoi la Acasă</a>'
    box = s4.empty_box('fără acces', ['Nu ai acces la Utilizatori.',
                                      'Conturile și accesul la ferme le setează proprietarul. Rolul tău, Observator, vede paginile seriilor '
                                      'și ale fermelor, și rapoartele.'], step, 'Acces')
    return f'<div class="ip-shell">{Q.sidebar_viewer()}<main class="ip-shell__main">\n<header class="ip-pagehead"><div><h1>Utilizatori</h1></div></header>\n{box}\n</main></div>'
EMPTY = users("u", row("u", OWN) + U["EMPTY"], U["counts"]([OWN]))

# ================================================================= phone list (owner, read-only, as S19–S21)
def phone_list():
    rows = ''.join(f'<tr{" class=\"s22-poff\"" if not u["active"] else ""}><th scope="row" class="s19-fname"><span class="s22-pn">{u["name"]}'
                   f'{" <span class=\"ip-note\">(tu)</span>" if u.get("me") else ""}</span>'
                   f'<span class="ip-note s22-pm">{u["role"]} · {u["farms"] or "toate fermele"}</span>'
                   + (f'<span class="ip-note s22-pm">ultima intrare <span class="ip-num">{u["seen"]}</span></span>' if u['seen'] else
                      '<span class="ip-note s22-pm">– încă n-a intrat</span>') + '</th>'
                   f'<td class="ip-nowrap">{lamp("ok", "activ") if u["active"] else lamp("off", "dezactivat")}</td></tr>' for u in BASE)
    blk = (f'<section class="ip-block" aria-labelledby="h-usr"><div class="ip-sechead"><h2 class="ip-t-heading" id="h-usr">Utilizatori</h2>'
           f'<span class="ip-note">{U["counts"](BASE)}</span></div>'
           f'<div class="ip-tablescroll ip-tablescroll--framed" tabindex="0" role="region" aria-labelledby="h-usr">'
           f'<table class="ip-table ip-table--rowheads s19-table s19-ptable s22-ptable"><caption class="ip-sr">Utilizatorii, cu rolul, fermele, ultima intrare și starea lor</caption>'
           f'<thead><tr><th scope="col">Nume · rol · ferme · intrare</th><th scope="col">Stare</th></tr></thead><tbody>{rows}</tbody></table></div>'
           f'<p class="ip-note ip-m0">Utilizatorii îi adaugi și îi editezi pe calculator.</p></section>')
    cr = CRUMBS.replace('<a href="#setari">Setări</a>', '<a href="#setari"><span class="ip-crumbs__txt">Setări</span></a>').replace('<nav class="ip-crumbs"', '<nav class="ip-crumbs ip-crumbs--touch"', 1)  # FIN06: system touch crumb (section 52)
    body = A.phone([f'<header class="ip-pagehead s2-phead s14-phead"><div>{cr}<h1>Utilizatori</h1></div></header>', blk])
    for a, z in (('class="ip-tabbar__item is-current" href="#acasa" aria-current="page"', 'class="ip-tabbar__item" href="#acasa"'),
                 ('class="ip-tabbar__item" href="#mai-mult"', 'class="ip-tabbar__item is-current" href="#mai-mult" aria-current="page"')):
        body = sub(body, a, z)
    assert body.count('is-current') == 1
    return body

# ================================================================= CSS (screen-only; tokens only; system classes untouched)
CSS22 = '''
.s22-region{container:s22-users / inline-size}
.ip-users__t td.s22-name{white-space:normal} /* S22P: beats bundle .ip-table td nowrap, so the fold line can wrap beside the panel at 1280 (it ran 4px past the list) */
.s22-fold{display:none}
/* S19P rule: beside the panel the list is under 760px (558 at 1280, 718 at 1440), so Login, Ferme and Ultima intrare fold
 * under the name and Rol, Stare and the Acțiuni key stay in view; the list alone keeps all seven columns. */
@container s22-users (max-width: 759.98px){
  .ip-users__t th{width:auto} .ip-users__t th::after{content:none} /* FIN11G: beside the panel the columns are content-sized again (the §54 shares are the list alone's) */
  .s22-cx{display:none}
  .s22-fold{display:block;color:var(--ink-secondary)}
}
.s22-checks{display:flex;flex-wrap:wrap;gap:0 var(--space-6);margin-block:calc(-1 * var(--space-3))} /* two farm boxes side by side: 44px targets, 20px text line in the form rhythm (as §44) */
.s22-acct{display:flex;flex-direction:column;align-items:flex-start;gap:var(--space-2);padding-top:var(--space-4);border-top:var(--ip-hair) solid var(--rule)}
.s22-state{display:flex;flex-wrap:wrap;align-items:center;gap:0 var(--space-3)}
.s22-keys{display:flex;flex-wrap:wrap;gap:var(--space-2);margin-top:var(--space-1)}'''
CSS_DESK = M.CSS_DESK + SC.BANNER + CSS22
CSS_PL = M.CSS_PL + CSS22 + '''
.s22-pn{display:block}
.s22-pm{display:block;font-weight:400}
.s22-poff .s22-pn{color:var(--ink-secondary)}'''

# ================================================================= write
pathlib.Path('/tmp/s22').mkdir(exist_ok=True)
def write(stem, title, root, css, w, h):
    r = M.write(stem, title, root, css, w, h)   # S19's writer: the .dc.html + /tmp/s19/<stem>.html; moved into /tmp/s22
    src = pathlib.Path(f'/tmp/s19/{stem}.html'); (pathlib.Path('/tmp/s22') / src.name).write_text(src.read_text(encoding='utf-8'), encoding='utf-8'); src.unlink()
    return r
STEM, TT = 'S22-utilizatori', 'S22 · Utilizatori'
D39 = 'exemplu de stare (D39)'
PH = SC.SIDE_H
out = []
def desk(stem, title, root, h=None):
    h = h or HEIGHTS.get(STEM + stem, 1024)
    out.append(write(STEM + stem, f'{TT}{" · " + title if title else ""}', f'<div class="ip-root s4s-rel" style="padding:0; min-height:{h}px">{root}</div>', CSS_DESK, 1440, h))

# 1 · the list
desk('', f'Petrea: {D39}', shell(ulist()))
# 2 · Ion Rusu's panel (Editează from his row menu): Nume (focus), Login fixed, Rol + what it does, Ferme, Cont
desk('--ion-rusu', f'Ion Rusu · panoul utilizatorului · Petrea: {D39}', vp(ulist(cur='rusu')) + upanel(RUSU, focus=True), PH)
# 3 · Adaugă utilizator: empty panel, the generated-password note (D78)
NEWBODY = lambda **e: [name_field('', True, e.get('name')), login_field(e.get('login_v', ''), e.get('login')), role_field(e.get('role_v'), e.get('role')),
                       farms_field(e.get('farms_v', ()), e.get('farms')), NEWPW]
desk('--utilizator-nou', f'Utilizator nou · Petrea: {D39}', vp(ulist(), pressed=True) + panel('Utilizator nou', '', NEWBODY(), ADD), PH)
# 4 · field errors (D39): Adaugă utilizatorul pressed with Nume empty, a login with a space and diacritics, no role, no farm
E = dict(name='Scrie numele persoanei, cum vrei să apară în listă, de exemplu Ion Rusu.',
         login_v='observator hîrbovăț',
         login='Loginul are un spațiu și diacritice. Folosește doar litere mici fără diacritice, cifre și punct, de exemplu observator.hirbovat.',
         role='Alege rolul: Veterinar, Observator, Manager sau Proprietar. De el depinde ce vede și ce introduce.',
         farms='Bifează cel puțin o fermă, de exemplu Hîrbovăț. Doar proprietarul vede toate fermele fără bifă.')
desk('--erori', f'Utilizator nou · nume gol, login greșit, fără rol, fără fermă · {D39}', vp(ulist(), pressed=True) + panel('Utilizator nou', '', NEWBODY(**E), ADD), PH)
# 5 · duplicate login (A2 State, D39): Petrea comes back; his deactivated account still holds „petrea”
DUP = ('Loginul „petrea” este deja folosit, de contul dezactivat al lui Petrea. Reactivează-l din listă sau alege alt login, '
       'de exemplu petrea.2.')
desk('--login-folosit', f'Utilizator nou · login deja folosit · {D39}: Petrea, cont dezactivat',
     vp(ulist(OFFL), pressed=True) + panel('Utilizator nou', '', [name_field('Petrea'), login_field('petrea', DUP, focus=True), role_field('Veterinar'),
                                                               farms_field(('Telița',)), NEWPW], ADD), PH)
# 6 · after Adaugă utilizatorul (D39 Observator exemplu): the panel closed, the new row listed, the password Dialog over it
WITH_OBS = order(BASE + [OBS])
PW_NOTE = ('<!-- After Adaugă utilizatorul: the panel closes, the new row is listed and the password Dialog opens with focus on '
           'Copiază parola; Închide discards the password (no Toast, D78) and focus goes to the new row\'s Acțiuni key. -->')
desk('--parola-cont-nou', f'Observator exemplu adăugat · parola temporară, o singură dată · {D39}, parola exemplu',
     shell(ulist(WITH_OBS), inert=True) + PW_NOTE + scrim(dlg_pw('pn', 'Parola temporară pentru Observator exemplu', U['CTX_NEW'], PW_NEW, 'Observator exemplu')))
# 7 · salvat (D39): Dr. Ceban given Telița as well: panel closed, row updated, focus on his key, Toast + Anulează
WITH_C2 = order([OWN, RUSU, CEBAN2, PETREA])
desk('--salvat', f'Dr. Ceban salvat · {D39}: Telița adăugată',
     shell(ulist(WITH_C2, focus='ceban')) + toast('Dr. Ceban salvat · Veterinar la Telița și Hîrbovăț', 'Salvează utilizatorul',
                                                   "the panel closes and focus goes back to Dr. Ceban's Acțiuni key"))
# 8 · eroare la salvare (D39): the same change, draft kept, Banner + Reîncearcă (focus) under the fields
SAVE_ERR = save_banner('Modificările lui Dr. Ceban nu s-au salvat: conexiunea s-a întrerupt.',
                       'Verifică conexiunea la internet, apoi apasă Reîncearcă. '
                       'Ce ai bifat rămâne în panou.')
desk('--eroare-la-salvare', f'Dr. Ceban · eroare la salvare · {D39}: Telița bifată',
     vp(ulist(cur='ceban')) + upanel(CEBAN, farms_on=('Telița', 'Hîrbovăț'), extra=SAVE_ERR), PH)
# 9 · your own account: the only active owner keeps the role (A2), farms "toate", Dezactivează aria-disabled with the reason
ONLY_OWNER = ('Ești singurul proprietar activ, așa că rolul tău rămâne Proprietar. Ca să-l schimbi, dă întâi rolul Proprietar '
              'altui utilizator, de exemplu lui Ion Rusu.')
desk('--contul-tau', f'contul tău · singurul proprietar, nu se dezactivează · Petrea: {D39}',
     vp(ulist(cur='own')) + panel('Proprietar (tu)', ctx(OWN), [name_field(OWN['name'], True), login_field(OWN['login'], ro=True),
                                                      role_field('Proprietar', ro=ONLY_OWNER), farms_field(owner=True), acct(OWN)]), PH)
# 10 · Dezactivează (A2 Dialog, D39 Petrea) from his row menu: page inert, focus on Anulează
desk('--dezactiveaza', f'Dezactivează contul lui Petrea · {D39}', shell(ulist(), inert=True) + scrim(dlg_off(PETREA)))
# 11 · dezactivat (D39): Petrea last with the off lamp, Toast + Anulează, focus back on his key
desk('--dezactivat', f'Petrea dezactivat · {D39}',
     shell(ulist(OFFL, focus='petrea')) + toast('Contul lui Petrea e dezactivat · datele introduse de el rămân', 'Dezactivează',
                                                 "the Dialog closes and focus goes back to Petrea's Acțiuni key (his row now listed last)"))
# 12 · Resetează parola from Ion Rusu's panel (after its confirmation): the new password, shown once (D78, D39 value)
desk('--parola-resetata', f'Ion Rusu · parolă resetată, o singură dată · {D39}, parola exemplu',
     vp(ulist(cur='rusu'), inert=True) + upanel(RUSU, inert=True, pressed='reset')
     + '<!-- Închide: focus goes back to Resetează parola in the panel. -->'
     + scrim(dlg_pw('pr', 'Parola nouă pentru Ion Rusu', U['CTX_RESET'], PW_RESET, 'Ion Rusu')), PH)
# 13 · se încarcă (SYS12 loading rows)
desk('--se-incarca', 'se încarcă', shell(LOADING))
# 14 · eroare la încărcare (SYS12 load failed)
desk('--eroare-la-incarcare', 'eroare la încărcare', shell(LOADERR))
# 15 · fără acces (Observator; a manager or a vet: same picture)
desk('--fara-acces', 'observator · fără acces', noaccess())
# 16 · doar tu (A2 Empty, D39 first use)
desk('--doar-tu', f'doar tu · {D39}', shell(EMPTY))
# 17 · phone, read-only list
h = HEIGHTS.get(STEM + '--telefon', 844)
out.append(write(STEM + '--telefon', f'{TT} · telefon · doar citire · Petrea: {D39}', f'<div class="ip-root" style="padding:0; width:390px; min-height:{h}px">{phone_list()}</div>', CSS_PL, 390, h))
pathlib.Path('/tmp/s22/boards.json').write_text(json.dumps([{k: b[k] for k in ('stem', 'w', 'h')} for b in out]))

if CANVAS:  # S22 row, entries keep x/y, a new row goes below the lowest (canvas_rows.py); 80 px gaps
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    x = CR.place(c, STEM, [(b['stem'], b['w'], b['h'], b['title']) for b in out])
    CR.save(ROOT, c)
    print('canvas: S22 row y', CR.row_y(c, STEM), 'ends at x', x - 80)
