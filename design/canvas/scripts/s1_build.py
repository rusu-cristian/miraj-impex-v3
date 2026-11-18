"""S1 · Autentificare (spec A1 Sign in; screens.md S1: Wordmark, Login, Parolă, Intră; D80 sign-in defaults, D121 S1 calls).
A signed-out page: no Sidebar, no BottomTabBar, no PageHeader. The card is the SYS13 LoginCard (additions.css §31) and its
h1 the SYS13 Wordmark, unchanged: their markup helper card() is read from design/system/scripts/sys13_build.py (the part
above its preview blocks, so nothing is written there). Screen CSS is layout only (.s1-root: the page fills the board).
Copy: spec A1 word for word (title "Impex Line", "Login", "Parolă", "Intră", the four refusals); D80 proposed copy
("Scrie loginul." / "Scrie parola.", "Ai uitat parola? Vorbește cu administratorul.").
People as on S22 (D118/D119): Dr. Ceban (vet, login ceban) signs in on the phone; Ion Rusu (manager, ion.rusu) with the
password the owner reset for him on S22 `--parola-resetata` (mnpq6789rstu, D39 example), shown only where the show key is
on; Petrea (petrea), deactivated on S22 `--dezactivat`. Every other password is masked and is a D39 sample value.
Boards (no row was reserved: below S22, y 39345 + 1024 + 120 = 40489; desktops first, then the phones, x step w + 80):
  S1-autentificare                      desktop: card at the top of the centred column, focus on Login (D80)
  --gresit-desktop                      desktop: ion.rusu, wrong credentials (A1), Parolă emptied, focus on Parolă
  --telefon                             phone default, focus on Login
  --parola-vizibila                     phone: ion.rusu + the S22 reset password, show key on (aria-pressed), focus on it
  --se-intra                            phone: ceban, signing in (spinner in Intră, label kept, form aria-busy)
  --campuri-goale                       phone: Intră with both fields empty, focus on Login (D80 copy)
  --gresit                              phone: wrong credentials (A1), Parolă emptied, focus on Parolă
  --prea-multe-incercari                phone: locked 5 minutes (A1), Intră aria-disabled, focus on the Banner
  --cont-dezactivat                     phone: petrea, deactivated (A1), focus on the Banner
  --fara-conexiune                      phone: offline (A1), values kept, focus on the Banner
Success goes to the role's start page (S2 Acasă); A1 has no first-sign-in step (no forced password change, D78), so none.
python3 -I s1_build.py <repo_root> [--canvas]
Writes design/canvas/project/S1-autentificare[--<state>].dc.html, /tmp/s1/<stem>.html, /tmp/s1/boards.json; --canvas adds
the S1 row to canvas.json (nothing else changes). Renders: s1_render.py [stems]."""
import sys, json, pathlib

ARGS = sys.argv[1:]
CANVAS = '--canvas' in ARGS; ARGS = [a for a in ARGS if a != '--canvas']
ROOT = pathlib.Path(ARGS[0]).resolve()

# SYS13 helpers (LoginCard + Wordmark): the code above the preview blocks only (no file is written)
SRC = (ROOT / 'design/system/scripts/sys13_build.py').read_text(encoding='utf-8')
CUT = '# ------------------------------------------------------------------ desktop card states'
assert SRC.count(CUT) == 1
L = {'__name__': 'sys13_parts', '__file__': str(ROOT / 'design/system/scripts/sys13_build.py')}  # sys13 finds the repo from its own path
exec(compile(SRC.split(CUT)[0], 'sys13_build.py', 'exec'), L)
card = L['card']
assert L['MSG']['wrong'] == '<b>Login sau parolă greșită.</b>'   # A1 word for word, through SYS13

PW_RUSU = 'mnpq6789rstu'   # D39: Ion Rusu's password after the S22 reset (s22_build.PW_RESET), shown with the show key on
PW_X = 'exemplu.parola'    # D39: a masked sample value (never shown in clear)
P = 's1'                   # one card per board: ids s1-h, s1-login, s1-pw, s1-msg …

CSS = '''.s1-root{display:flex;flex-direction:column}
.s1-root > .ip-login-page{flex:1 0 auto;min-height:0}'''

FONTS = ('<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500&amp;'
         'family=IBM+Plex+Mono:wght@400;500&amp;display=swap">')
SYS = ROOT / 'design/system'
pathlib.Path('/tmp/s1').mkdir(exist_ok=True)
STEM, TT = 'S1-autentificare', 'S1 · Autentificare'
D39 = 'exemplu de stare (D39)'
out = []

def write(stem, title, inner, w, h, note):
    phone = w == 390
    size = f'width:390px; min-height:{h}px' if phone else f'min-height:{h}px'
    main = f'<main class="ip-login-page{" ip-login-page--phone" if phone else ""}">{inner}</main>'
    root = f'<!-- {note} -->\n<div class="ip-root s1-root" style="padding:0; {size}">{main}</div>'
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
{FONTS}
<style>body{{margin:0}}
{CSS}</style>
</helmet>
{root}
</x-dc>
<script type="text/x-dc" data-dc-script data-props='{{"$preview":{{"width":{w},"height":{h}}}}}'>
class Component extends DCLogic {{ renderVals() {{ return {{}}; }} }}
</script>
</body>
</html>
'''
    (ROOT / f'design/canvas/project/{stem}.dc.html').write_text(dc, encoding='utf-8')
    wrap = f'''<!doctype html><html lang="ro"><head><meta charset="utf-8"><title>{title}</title>
<link rel="stylesheet" href="file://{SYS}/tokens.css"><link rel="stylesheet" href="file://{SYS}/components/bundle.css"><link rel="stylesheet" href="file://{SYS}/components/additions.css">
{FONTS}<style>body{{margin:0}}
{CSS}</style></head><body>{root}</body></html>'''
    pathlib.Path(f'/tmp/s1/{stem}.html').write_text(wrap, encoding='utf-8')
    out.append({'stem': stem, 'title': title, 'w': w, 'h': h})

def desk(stem, title, note, **kw): write(STEM + stem, f'{TT}{" · " + title if title else ""}', card(P, **kw), 1440, 1024, note)
def phone(stem, title, note, **kw): write(STEM + stem, f'{TT} · telefon{" · " + title if title else ""}', card(P, phone=True, **kw), 390, 844, note)

# ---------------------------------------------------------------- desktop (owner and managers sign in here too)
desk('', '', 'The page opens with focus on Login (D80). The real page title element reads "Autentificare · Impex Line" (SYS13).',
     focus='login')
desk('--gresit-desktop', f'login sau parolă greșită · ion.rusu: {D39}',
     'Intră refused (A1 wrong credentials): both fields aria-invalid + described by the Banner, Parolă emptied, focus on '
     'Parolă (D80); the Banner itself has no role, so the message is read once through the field.',
     login='ion.rusu', msg='wrong', focus='pw')
# ---------------------------------------------------------------- phone (A1 job: the vet gets in quickly on the phone)
phone('--telefon', '', 'Phone, as opened from the home-screen shortcut: focus on Login (D80).', focus='login')
phone('--parola-vizibila', f'parola vizibilă · ion.rusu cu parola resetată în S22: {D39}',
      'Arată parola pressed: aria-pressed=true, eye-off icon, the field is type=text; focus stays on the key. The field '
      'goes back to type=password before the form submits (password managers).',
      login='ion.rusu', pw=PW_RUSU, shown=True, focus='eye')
phone('--se-intra', f'se intră · ceban: {D39}',
      'Intră pressed: spinner in the key, label kept, form aria-busy; more presses and Enter are ignored. On success the '
      "role's start page opens (S2 Acasă); the session lasts 30 days on this device (A1).",
      login='ceban', pw=PW_X, busy=True)
phone('--campuri-goale', 'câmpuri goale',
      'Intră with both fields empty: the app checks only emptiness (D80); focus on the first empty field, which reads '
      'its error line through aria-describedby. Length rules stay on the server.',
      empty_err=('login', 'pw'), focus='login')
phone('--gresit', f'login sau parolă greșită · ceban: {D39}',
      'Intră refused (A1 wrong credentials): Parolă emptied, focus on Parolă, the Banner read through its aria-describedby.',
      login='ceban', msg='wrong', focus='pw')
phone('--prea-multe-incercari', f'prea multe încercări · ceban: {D39}',
      'A1: after 5 failed attempts sign-in waits 5 minutes. Focus on the Banner (tabindex=-1, nothing in a field fixes it); '
      'Intră aria-disabled and described by the Banner until the 5 minutes are over.',
      login='ceban', pw=PW_X, msg='locked', locked=True, focus='msg')
phone('--cont-dezactivat', f'cont dezactivat · petrea, dezactivat în S22: {D39}',
      'A1 deactivated user: focus on the Banner; values kept. The owner reactivates the account in S22 (D119).',
      login='petrea', pw=PW_X, msg='off', focus='msg')
phone('--fara-conexiune', f'fără conexiune · ceban: {D39}',
      'A1 offline: when the device knows it is offline the Banner shows on Intră without a request; both values kept, '
      'focus on the Banner; Intră again retries (A1 online only, D80).',
      login='ceban', pw=PW_X, msg='offline', focus='msg')
pathlib.Path('/tmp/s1/boards.json').write_text(json.dumps([{k: b[k] for k in ('stem', 'w', 'h')} for b in out]))
for b in out: print('ok', b['stem'], b['w'], b['h'])

if CANVAS:  # S1 row, entries keep x/y, a new row goes below the lowest (canvas_rows.py); 80 px gaps
    sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent)); import canvas_rows as CR
    c = CR.load(ROOT)
    x = CR.place(c, STEM, [(b['stem'], b['w'], b['h'], b['title']) for b in out])
    CR.save(ROOT, c)
    print('canvas: S1 row y', CR.row_y(c, STEM), 'ends at x', x - 80)
