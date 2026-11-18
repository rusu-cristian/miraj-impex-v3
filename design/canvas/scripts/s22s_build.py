"""S22S · Utilizatori · state boards (spec A2; D78, D118, D119). Imports s22_build (which rewrites the 17 S22 boards, and
through s19_build the S19/S18/S17/S14S/S2/S7 boards, byte-identical) and draws the leftover S22 states on the same row
(y 39220): the new desktops follow the 16 S22 desktops; the S22 phone moves right (x only). Today = 16.11.2026.
Markup is the SYS12 UserRow / OneTimePassword / Dialog (additions.css §30) through the S22 helpers; no new screen CSS.
S22P: SYS12 dialog() is role=alertdialog and otp() leaves the status region empty on a failed copy (the screen-side
overrides are gone); panel boards mark the current row (§45). FIN07: the S22 list page (page-head add key, h2 list head).
People as S22: Proprietar (you), Ion Rusu, Dr. Ceban (kit, illustrative); Petrea (A2 sample), Observator exemplu, every
sign-in time, both passwords and every failure are D39 state examples, said in the board titles.
Boards (desktops, 1440×1024)
  --meniu-deschis         Ion Rusu's row menu open, keyboard on Resetează parola (SYS12 "Meniu deschis")
  --meniu-contul-tau      your own row's menu: Dezactivează focusable but inert, the reason under the items
  --reseteaza-parola      Resetează parola lui Ion Rusu, from his row menu: confirm Dialog, focus on Anulează
  --resetare-esuata       D39: the same Dialog after the reset failed: fault Banner, Anulează + Reîncearcă (focus)
  --parola-copiata        D39: Observator exemplu's password after Copiază parola: ok Lamp "copiată", focus kept on the key
  --copiere-esuata        D39: Ion Rusu's new password, copy failed: focus in the field (the app selects it), Ctrl+C line
  --meniu-dezactivat      D39: Petrea's row menu (deactivated): Resetează parola inert + reason, keyboard on Reactivează;
                          opens downward (room below the list; upward it would cover the list head, SYS12 FIN nit, D119)
  --reactivat             D39: Petrea active again, no Dialog (D78): Toast + Anulează, focus back on his Acțiuni key
  --editeaza-dezactivat   D39: Petrea's panel while deactivated: fields editable (they apply when he is reactivated), Cont
                          block with the off lamp, Reactivează contul, Resetează parola aria-disabled + reason
python3 -I s22s_build.py <repo_root> [--canvas]   after `s22_build --canvas` (canvas_rows.py)
Renders: s22_render.py [--w 1280] <stem suffixes> (the boards are appended to /tmp/s22/boards.json: shots S22-17 … S22-25)."""
import sys, json, pathlib, importlib.util, io, contextlib

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve(); HERE = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
sys.argv = ['s22_build.py', str(ROOT)]
sp = importlib.util.spec_from_file_location('s22_build', HERE / 's22_build.py')
S = importlib.util.module_from_spec(sp)
with contextlib.redirect_stdout(io.StringIO()):
    sp.loader.exec_module(S)   # rewrites the 17 S22 boards byte-identical (no --canvas)
U, sub, icon, lamp = S.U, S.sub, S.icon, S.lamp
S22_OUT = [{k: b[k] for k in ('stem', 'w', 'h')} for b in S.out]
STEM, TT, D39 = S.STEM, S.TT, S.D39
assert len(S22_OUT) == 17 and S22_OUT[-1]['stem'] == STEM + '--telefon'
RUSU, OWN, PETREA, PETREA_OFF, OBS = S.RUSU, S.OWN, S.PETREA, S.PETREA_OFF, S.OBS
BASE, OFFL = S.BASE, S.OFFL
BTN = U['btn']

# ================================================================= list with one row menu open (SYS12 row, S22 fold)
def ulist_open(us, key, item, up=False, p='u'):
    """S22's list with the row `key` menu open and the keyboard on `item` (SYS12 menu: APG menu button; the key shows
    aria-expanded, focus is on the item). `up` as SYS12 (last rows open upward) unless the screen has room below."""
    orig = U['row']
    def one(p_, u_, focus=False):
        if u_['key'] != key: return orig(p_, u_, focus=focus)
        return orig(p_, u_, open_=True, active_item=item, focus=focus, up=up)
    U['row'] = one
    try:
        body = ''.join(S.row(p, u) for u in us)
    finally:
        U['row'] = orig
    assert body.count('aria-expanded="true"') == 1 and body.count(' is-active') == 1
    return S.users(p, body, U["counts"](us), S.NOTE)
MENU_NOTE = ('<!-- Row menu (SYS12, APG menu button): focus is on the item marked is-active; ↑/↓, Home/End, type-ahead move; '
             'Esc or Tab closes it and focus goes back to the Acțiuni key; disabled items stay focusable and inert, their reason '
             'tied by aria-describedby. -->')

# ================================================================= dialogs (SYS12)
RP_BODY = '<p id="rp-n">Parola de acum nu va mai merge. Noua parolă temporară se arată o singură dată, aici.</p>'
def dlg_reset():   # SYS12 "Resetează parola · confirmare": focus on Anulează
    return (U['dialog']('rp', 'Resetează parola lui Ion Rusu', RP_BODY,
                       BTN('Anulează', focus=True, extra=' autofocus="autofocus"') + BTN('Resetează parola', 'primary'), desc='rp-n'))
def dlg_reset_failed():   # SYS12 "Resetarea nu a mers": fault Banner (no role) described by the dialog and Reîncearcă (focus)
    why = U['banner']('fault', '<b>Parola lui Ion Rusu nu s-a resetat;</b> cea veche merge în continuare. '
                               'Verifică conexiunea la internet, apoi apasă Reîncearcă.', 'rf-why')
    return (U['dialog']('rf', 'Resetează parola lui Ion Rusu', why,
                       BTN('Anulează') + BTN('Reîncearcă', 'primary', focus=True, extra=' autofocus="autofocus" aria-describedby="rf-why"'),
                       desc='rf-why'))
def dlg_pw(p, title, ctx_, pw, name, **kw):
    o = U['otp'](p, ctx_, pw, name, **kw)
    if kw.get('focus') == 'pw':   # is-selected is the SYS12 preview stand-in only: a screen board draws the focus ring
        o = sub(o, ' ip-focus is-selected"', ' ip-focus"')
    if kw.get('failed'):   # D119 (SYS13 rule), in SYS12 otp() since S22P: the field carries the message, the status region stays empty
        assert '<p class="ip-sr" role="status" aria-live="polite"></p>' in o
    return U['dialog'](p, title, o, BTN('Închide', 'primary'))
RESET_NOTE = ("<!-- Opened from Ion Rusu's row menu (Resetează parola). Anulează or Esc: focus back on his Acțiuni key. "
              "Resetează parola: the same Dialog turns into the password step (as --parola-resetata, focus on Copiază parola) "
              "or, if it fails, shows the fault Banner (--resetare-esuata). -->")

# ================================================================= the deactivated user's panel (D119)
def acct_off(u):
    """Cont for a deactivated account: off lamp + last sign-in, Reactivează contul (no Dialog, Toast, D78) in place of
    Dezactivează contul, Resetează parola aria-disabled with the SYS12 reason (a deactivated user gets no new password)."""
    keys = ('<button type="button" class="ip-btn ip-btn--secondary">Reactivează contul</button>'
            '<button type="button" class="ip-btn ip-btn--secondary" aria-disabled="true" aria-describedby="up-pw-why">Resetează parola</button>')
    return (f'<section class="s22-acct" aria-labelledby="up-acct-h"><h3 class="ip-cap ip-m0" id="up-acct-h">Cont</h3>'
            f'<p class="ip-m0 s22-state">{lamp("off", "dezactivat")}<span>ultima intrare <span class="ip-num">{u["seen"]}</span></span></p>'
            f'<p class="ip-note ip-m0">Rolul și fermele schimbate acum se aplică '
            f'după ce îl reactivezi; reactivat, intră cu parola lui de dinainte.</p>'
            f'<div class="s22-keys">{keys}</div>'
            f'<p class="ip-note ip-m0" id="up-pw-why">Contul e dezactivat. Reactivează-l ca să-i dai o parolă nouă.</p></section>')
def panel_off(u):
    p = S.upanel(u, focus=True)
    a = S.acct(u)
    assert p.count(a) == 1
    return p.replace(a, acct_off(u))

# ================================================================= write
out = []
def desk(stem, title, root, h=1024):
    out.append(S.write(STEM + stem, f'{TT} · {title}', f'<div class="ip-root s4s-rel" style="padding:0; min-height:{h}px">{root}</div>',
                       S.CSS_DESK, 1440, h))
PH = S.PH

# 1 · Ion Rusu's row menu open, keyboard on Resetează parola
desk('--meniu-deschis', f'meniul rândului deschis · Ion Rusu, după ↓ tastatura pe „Resetează parola” (deschis, cade pe Editează) · Petrea: {D39}',
     S.shell(MENU_NOTE + ulist_open(BASE, 'rusu', 'Resetează parola')))
# 2 · your own row: Dezactivează focusable but inert, its reason under the items
desk('--meniu-contul-tau', f'meniul rândului tău · Dezactivează indisponibil, cu motivul · Petrea: {D39}',
     S.shell(MENU_NOTE + ulist_open(BASE, 'own', 'Dezactivează')))
# 3 · Resetează parola lui Ion Rusu (from his row menu): confirm, focus on Anulează
desk('--reseteaza-parola', f'Resetează parola lui Ion Rusu · confirmare · Petrea: {D39}',
     S.shell(S.ulist(), inert=True) + RESET_NOTE + S.scrim(dlg_reset()))
# 4 · reset failed (D39): the Dialog stays, fault Banner, Reîncearcă (focus), Esc = Anulează
desk('--resetare-esuata', f'Resetează parola lui Ion Rusu · resetarea nu a mers, Reîncearcă · {D39}',
     S.shell(S.ulist(), inert=True) + '<!-- Focus moves to Reîncearcă (described by the Banner); Esc = Anulează; the old password still works. -->'
     + S.scrim(dlg_reset_failed()))
# 5 · copied (D39 Observator exemplu, as --parola-cont-nou): ok Lamp "copiată", live "Parola a fost copiată.", focus kept on the key
desk('--parola-copiata', f'Observator exemplu · parola temporară copiată · {D39}, parola exemplu',
     S.shell(S.ulist(S.WITH_OBS), inert=True) + S.PW_NOTE
     + S.scrim(dlg_pw('pc', 'Parola temporară pentru Observator exemplu', U['CTX_NEW'], S.PW_NEW, 'Observator exemplu', copied=True, focus='copy')))
# 6 · copy failed (D39, as --parola-resetata): focus in the field, the app selects the value; the line names the fix
desk('--copiere-esuata', f'Ion Rusu · copierea parolei nu a mers, parola selectată · {D39}, parola exemplu',
     S.vp(S.ulist(cur='rusu'), inert=True) + S.upanel(RUSU, inert=True, pressed='reset')
     + '<!-- Copy failed: the app runs input.focus() + input.select() (the value in the ::selection colours, SYS12); a still board '
       'holds no selection, so only the focus ring is drawn. Închide: focus back on Resetează parola in the panel. -->'
     + S.scrim(dlg_pw('cf', 'Parola nouă pentru Ion Rusu', U['CTX_RESET'], S.PW_RESET, 'Ion Rusu', failed=True, focus='pw')), PH)
# 7 · Petrea's row menu (deactivated, D39): Resetează parola inert + reason, keyboard on Reactivează; opens downward (D119)
desk('--meniu-dezactivat', f'meniul unui cont dezactivat · Petrea: Reactivează, Resetează parola indisponibil · {D39}',
     S.shell(MENU_NOTE + ulist_open(OFFL, 'petrea', 'Reactivează', up=False)))
# 8 · reactivated (D39): no Dialog (D78); Petrea back among the vets, Toast + Anulează (D118), focus on his key
desk('--reactivat', f'Petrea reactivat · {D39}',
     S.shell(S.ulist(BASE, focus='petrea')) + S.toast('Contul lui Petrea e activ din nou · intră cu parola lui de dinainte', 'Reactivează',
                                                     "the menu closes and focus goes back to Petrea's Acțiuni key (his row now among the active vets)"))
# 9 · editing a deactivated user (D39 Petrea, Editează from his row menu)
desk('--editeaza-dezactivat', f'Petrea, cont dezactivat · panoul utilizatorului · {D39}',
     S.vp(S.ulist(OFFL, cur='petrea')) + panel_off(PETREA_OFF), PH)

pathlib.Path('/tmp/s22/boards.json').write_text(json.dumps(S22_OUT + [{k: b[k] for k in ('stem', 'w', 'h')} for b in out]))

if CANVAS:  # S22 row: the S22 desktops, these desktops, then the S22 phone
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    old = [b['stem'] for b in S22_OUT]
    x = CR.place(c, STEM, old[:16] + [(b['stem'], b['w'], b['h'], b['title']) for b in out] + old[16:], 0)
    CR.save(ROOT, c)
    print('canvas: S22 row y', CR.row_y(c, STEM), 'ends at x', x - 80)
