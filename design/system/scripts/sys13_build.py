"""Build the SYS13 previews (LoginCard + Wordmark, S1 Autentificare, spec A1): python3 sys13_build.py
Copy is spec A1's (title "Impex Line", fields "Login" / "Parolă", key "Intră", the four refusal sentences).
Login `petrea` is A1's sample; `ion.rusu` and the password `k7mq4tzw9bhe` are SYS12 state examples (D39):
Ion Rusu signing in with the temporary password from his reset. Brand: PRODUCT.md (name, no logo)."""
import pathlib
REPO = pathlib.Path('/home/claude/miraj-impex-v3')
ROOT = REPO / 'design/system/components'

def svg(paths, cls='ip-icon ip-icon--16'):
    return f'<svg class="{cls}" viewBox="0 0 24 24" aria-hidden="true">{paths}</svg>'

IC = {  # Lucide
    'eye': '<path d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0"/>'
           '<circle cx="12" cy="12" r="3"/>',
    'eyeoff': '<path d="M10.733 5.076a10.744 10.744 0 0 1 11.205 6.575 1 1 0 0 1 0 .696 10.747 10.747 0 0 1-1.444 2.49"/>'
              '<path d="M14.084 14.158a3 3 0 0 1-4.242-4.242"/>'
              '<path d="M17.479 17.499a10.75 10.75 0 0 1-15.417-5.151 1 1 0 0 1 0-.696 10.75 10.75 0 0 1 4.446-5.143"/>'
              '<path d="m2 2 20 20"/>',
    'fault': '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>',
}

def cap(text):
    return f'<span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">{text}</span>'

PW = 'k7mq4tzw9bhe'   # state example (D39, SYS12): Ion Rusu's temporary password

# A1 › States, word for word. Bold = the first sentence (the Banner pattern of SYS12).
MSG = {
    'wrong': '<b>Login sau parolă greșită.</b>',
    'locked': '<b>Prea multe încercări.</b> Încearcă din nou peste <span class="ip-num">5</span> minute.',
    'off': '<b>Contul tău este dezactivat.</b> Vorbește cu administratorul.',
    'offline': '<b>Nu există conexiune.</b> Verifică internetul și încearcă din nou.',
}

def card(p, login='', pw='', phone=False, focus=None, shown=False, busy=False, msg=None, empty_err=(),
         locked=False, live=''):
    """LoginCard. focus: 'login' | 'pw' | 'eye' | 'go' | 'msg' (static ip-focus / is-focus demo classes).
    msg: a key of MSG -> fault Banner (summary) at the top of the form. empty_err: fields left empty on Intră."""
    t = ' ip-input--touch' if phone else ''
    tb = ' ip-btn--touch' if phone else ''
    ti = ' ip-iconbtn--touch' if phone else ''
    m = f'{p}-msg'
    banner = ''
    if msg:
        fb = ' ip-focus' if focus == 'msg' else ''
        banner = (f'<div class="ip-banner ip-banner--fault ip-login__msg{fb}" id="{m}" tabindex="-1">'
                  f'{svg(IC["fault"])}<p class="ip-login__bn">{MSG[msg]}</p></div>')
    # login field
    lerr = 'login' in empty_err
    lbad = lerr or msg == 'wrong'
    ldesc = ' '.join(x for x in ((f'{p}-login-err' if lerr else ''), (m if msg == 'wrong' else '')) if x)
    lcls = 'ip-input' + t + (' is-error' if lbad else '') + (' is-focus' if focus == 'login' else '')
    la = (' aria-invalid="true"' if lbad else '') + (f' aria-describedby="{ldesc}"' if ldesc else '')
    lv = f' value="{login}"' if login else ''
    f_login = (f'<div class="ip-field"><label class="ip-label" for="{p}-login">Login</label>'
               f'<div class="{lcls}"><input id="{p}-login" name="login" type="text" autocomplete="username" '
               f'autocapitalize="none" autocorrect="off" spellcheck="false" maxlength="40" required{lv}{la}></div>'
               + (f'<p class="ip-err" id="{p}-login-err">Scrie loginul.</p>' if lerr else '') + '</div>')
    # password field (+ show/hide key inside the field)
    perr = 'pw' in empty_err
    pbad = perr or msg == 'wrong'
    pdesc = ' '.join(x for x in ((f'{p}-pw-err' if perr else ''), (m if msg == 'wrong' else '')) if x)
    pcls = ('ip-input ip-login__pw' + t + (' is-error' if pbad else '')
            + (' is-focus' if focus == 'pw' else '') + (' is-eye' if focus == 'eye' else ''))
    pa = (' aria-invalid="true"' if pbad else '') + (f' aria-describedby="{pdesc}"' if pdesc else '')
    pv = f' value="{pw}"' if pw else ''
    ptype = 'text' if shown else 'password'
    fe = ' ip-focus' if focus == 'eye' else ''
    eye = (f'<button type="button" class="ip-iconbtn{ti} ip-login__eye{fe}" aria-label="Arată parola" '
           f'aria-pressed="{"true" if shown else "false"}" aria-controls="{p}-pw">'
           f'{svg(IC["eyeoff" if shown else "eye"])}</button>')
    f_pw = (f'<div class="ip-field"><label class="ip-label" for="{p}-pw">Parolă</label>'
            f'<div class="{pcls}"><input id="{p}-pw" name="password" type="{ptype}" autocomplete="current-password" '
            f'autocapitalize="none" autocorrect="off" spellcheck="false" required{pv}{pa}>{eye}</div>'
            + (f'<p class="ip-err" id="{p}-pw-err">Scrie parola.</p>' if perr else '') + '</div>')
    # Intră
    gcls = 'ip-btn ip-btn--primary' + tb + ' ip-login__go' + (' ip-focus' if focus == 'go' else '')
    ga = (f' aria-disabled="true" aria-describedby="{m}"' if locked else '')
    spin = '<span class="ip-spin" aria-hidden="true"></span>' if busy else ''
    go = f'<button type="submit" class="{gcls}"{ga}>{spin}Intră</button>'
    hint = '<p class="ip-hint ip-login__hint">Ai uitat parola? Vorbește cu administratorul.</p>'
    b = ' aria-busy="true"' if busy else ''
    return (f'<section class="ip-panel ip-login" aria-labelledby="{p}-h">'
            f'<h1 class="ip-wordmark ip-wordmark--login" id="{p}-h">Impex Line</h1>'
            f'<form class="ip-login__form" method="post" action="/intra" novalidate aria-labelledby="{p}-h"{b}>'
            f'{banner}{f_login}{f_pw}{go}{hint}'
            f'<p class="ip-sr" role="status" aria-live="polite">{live}</p></form></section>')

def page(inner, phone=False):
    return f'<div class="ip-login-page{" ip-login-page--phone" if phone else ""}">{inner}</div>'

# ------------------------------------------------------------------ desktop card states (A1 › States + form states)
S = [
    ('Implicit · focus pe Login la deschidere', card('a', focus='login')),
    ('Completat · focus pe Intră', card('b', login='petrea', pw=PW, focus='go')),
    ('Parolă vizibilă · tasta ochi apăsată (aria-pressed)', card('c', login='ion.rusu', pw=PW, shown=True, focus='eye')),
    ('Se intră · spinner în Intră, eticheta rămâne', card('d', login='petrea', pw=PW, busy=True)),
    ('Câmpuri goale la Intră · focus pe primul câmp cu eroare', card('e', empty_err=('login', 'pw'), focus='login')),
    ('Login sau parolă greșită · parola golită, focus pe Parolă', card('f', login='petrea', msg='wrong', focus='pw')),
    ('Prea multe încercări · Intră indisponibil, focus pe mesaj',
     card('g', login='petrea', pw=PW, msg='locked', locked=True, focus='msg')),
    ('Cont dezactivat · focus pe mesaj', card('h', login='petrea', pw=PW, msg='off', focus='msg')),
    ('Fără conexiune · valorile rămân, focus pe mesaj', card('i', login='petrea', pw=PW, msg='offline', focus='msg')),
]
PH = [
    ('Telefon · implicit', card('j', phone=True, focus='login')),
    ('Telefon · se intră', card('k', login='petrea', pw=PW, phone=True, busy=True)),
    ('Telefon · greșit (Enter în Parolă: anunț prin status)',
     card('l', login='petrea', phone=True, msg='wrong', focus='pw', live='Login sau parolă greșită.')),
]

COLW = 'var(--panel-w)'
DESK_W = 'calc(var(--space-12) * 22)'          # 1056: a desktop page band
DESK_H = 'calc(var(--space-12) * 12)'          # 576
PHONE_H = 'calc(var(--space-12) * 17 + var(--space-6) + var(--space-1))'   # 844
FRAME = 'border:var(--ip-hair) solid var(--line);border-radius:var(--radius-md);overflow:hidden'

desk = (f'<div>{cap("S1 · Autentificare · desktop (pagina pe ground, cardul la lățimea panel-w, sus în coloană)")}'
        f'<div style="width:{DESK_W};height:{DESK_H};{FRAME}">'
        + page(card('z', focus='login')) + '</div></div>')
grid = (f'<div>{cap("Stări · desktop (card 440)")}'
        f'<div style="display:grid;grid-template-columns:repeat(3,{COLW});gap:var(--space-8) var(--space-6);align-items:start">'
        + ''.join(f'<div style="width:{COLW}">{cap(t)}{h}</div>' for t, h in S) + '</div></div>')
phones = (f'<div>{cap("S1 · telefon (390 × 844)")}'
          f'<div style="display:flex;gap:var(--space-6);align-items:start">'
          + ''.join(f'<div style="flex:1 1 0;min-width:0">{cap(t)}<div class="ip-phone" style="width:100%;height:{PHONE_H}">'
                    f'{page(h, phone=True)}</div></div>' for t, h in PH) + '</div></div>')
FOOT = ('<p class="ip-note" style="margin:0;max-width:{w}">Din kit: textele din spec A1 (titlul „Impex Line”, câmpurile „Login” '
        'și „Parolă”, tasta „Intră”, cele patru mesaje) și loginul petrea (A1). Exemple de stare (D39, din SYS12): loginul '
        'ion.rusu și parola k7mq4tzw9bhe. „Scrie loginul.”, „Scrie parola.” și „Ai uitat parola? Vorbește cu '
        'administratorul.” sunt texte de interfață propuse (SYS13, de revăzut).</p>')

lc = ('<!-- @dsCard group="App shell" height=2400 subtitle="SYS13" -->\n'
      '<div class="ip-root ip-stack" style="gap:var(--space-8)">\n'
      + desk + '\n' + grid + '\n' + phones + '\n' + FOOT.format(w=DESK_W) + '\n</div>\n')

# ------------------------------------------------------------------ Wordmark
SIDE = (f'<div style="width:var(--sidebar-w);background:var(--panel);border:var(--ip-hair) solid var(--line);'
        f'border-radius:var(--radius-md);padding:var(--space-3)">'
        f'<div class="ip-wordmark">Impex Line</div>'
        f'<a class="ip-nav is-current" href="#">Acasă</a><a class="ip-nav" href="#">Serie</a></div>')
LOGIN = (f'<div style="width:{COLW};background:var(--panel);border:var(--ip-hair) solid var(--line);'
         f'border-radius:var(--radius-md);padding:var(--space-6)">'
         f'<h1 class="ip-wordmark ip-wordmark--login">Impex Line</h1></div>')
wm = ('<!-- @dsCard group="App shell" height=420 subtitle="SYS13" -->\n'
      '<div class="ip-root ip-stack" style="gap:var(--space-8)">\n'
      '<div style="display:flex;gap:var(--space-8);align-items:start">'
      f'<div>{cap("Sidebar · .ip-wordmark (15/20, bundle.css, neschimbat)")}{SIDE}</div>'
      f'<div>{cap("Login · .ip-wordmark--login (title 24/30), h1 pe card · desktop și telefon")}{LOGIN}</div>'
      '</div>\n'
      '<p class="ip-note" style="margin:0;max-width:calc(var(--panel-w) * 2)">Numele produsului din PRODUCT.md '
      '(„Impex Line”, fără logo). În cod textul e „Impex Line”; majusculele și spațierea vin din CSS, așa că '
      'cititorul de ecran spune numele, nu literele.</p>\n</div>\n')

for comp, html in (('LoginCard', lc), ('Wordmark', wm)):
    (ROOT / comp).mkdir(exist_ok=True)
    (ROOT / comp / 'preview.html').write_text(html, encoding='utf-8')
print('ok', len(S), len(PH))
