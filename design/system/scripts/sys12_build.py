"""Build the SYS12 previews (UserRow list + OneTimePassword, S22 Utilizatori, spec A2): python3 sys12_build.py
Names from the kit: Ion Rusu, Dr. Ceban (craft.md §7, illustrative) and petrea (spec A1/A2 sample login, A2 dialog).
State examples (D39): the owner row "Proprietar exemplu", the viewer "Observator exemplu", every login except petrea,
every last sign-in time, Petrea's deactivated state and both passwords. Farm names are real (farms-sections.json)."""
import pathlib
REPO = pathlib.Path(__file__).resolve().parents[3]  # repo root from this script's place (design/system/scripts), so a copy writes into itself
ROOT = REPO / 'design/system/components'

def svg(paths, cls='ip-icon ip-icon--16'):
    return f'<svg class="{cls}" viewBox="0 0 24 24" aria-hidden="true">{paths}</svg>'

IC = {  # Lucide
    'more': '<circle cx="12" cy="5" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="12" cy="19" r="1"/>',
    'copy': '<rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/>',
    'fault': '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>',
    'info': '<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>',
    'plus': '<path d="M5 12h14"/><path d="M12 5v14"/>',
}

def cap(text):
    return f'<span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">{text}</span>'

def num(n):
    return f'<span class="ip-num">{n}</span>'

# ------------------------------------------------------------------ users (order: active by role, then deactivated)
OWNER = dict(key='own', name='Proprietar exemplu', login='proprietar', role='Proprietar', farms=None,
             seen='16.11.2026, 08:02', active=True, me=True)
RUSU = dict(key='rusu', name='Ion Rusu', login='ion.rusu', role='Manager', farms='Telița, Hîrbovăț',
            seen='16.11.2026, 17:40', active=True)
CEBAN = dict(key='ceban', name='Dr. Ceban', login='ceban', role='Veterinar', farms='Hîrbovăț',
             seen='16.11.2026, 08:34', active=True)
VIEW = dict(key='obs', name='Observator exemplu', login='observator', role='Observator', farms='Hîrbovăț',
            seen=None, active=True)
PETREA = dict(key='petrea', name='Petrea', login='petrea', role='Veterinar', farms='Telița',
              seen='28.02.2026, 07:15', active=False)
ALL = [OWNER, RUSU, CEBAN, VIEW, PETREA]

def items(p, u):
    """Row menu: Editează · Resetează parola · Dezactivează / Reactivează, disabled items with a reason note."""
    mid = f'{p}-{u["key"]}'
    note = None
    rows = [('Editează', '', False)]
    if u['active']:
        rows.append(('Resetează parola', ' aria-haspopup="dialog"', False))
        if u.get('me'):
            rows.append(('Dezactivează', ' aria-haspopup="dialog"', True))
            note = 'Nu îți poți dezactiva propriul cont.'
        else:
            rows.append(('Dezactivează', ' aria-haspopup="dialog"', False))
    else:
        rows.append(('Resetează parola', ' aria-haspopup="dialog"', True))
        rows.append(('Reactivează', '', False))
        note = 'Contul e dezactivat. Reactivează-l ca să-i dai o parolă nouă.'
    return rows, note, mid

def menu(p, u, active_item=None, open_=True, up=False):
    rows, note, mid = items(p, u)
    out = []
    for i, (label, extra, dis) in enumerate(rows):
        if i == 2:
            out.append('<div class="ip-users__sep" role="separator"></div>')
        cls = 'ip-menu__item' + (' is-active' if label == active_item else '')
        d = f' aria-disabled="true" aria-describedby="{mid}-why"' if dis else ''
        out.append(f'<button type="button" role="menuitem" tabindex="-1" class="{cls}"{extra}{d}>{label}</button>')
    n = f'<p class="ip-note" id="{mid}-why">{note}</p>' if note else ''
    db = f' aria-describedby="{mid}-why"' if note else ''
    h = '' if open_ else ' hidden'
    c = ' ip-users__menu--up' if up else ''
    return (f'<div class="ip-menu ip-users__menu{c}" id="{mid}-menu"{h}>'
            f'<div class="ip-users__items" role="menu" aria-label="Acțiuni pentru {u["name"]}"{db}>{"".join(out)}</div>{n}</div>')

def row(p, u, open_=False, active_item=None, focus=False, up=False, cur=False):
    cls = 'ip-user' + ('' if u['active'] else ' is-off') + (' is-self' if u.get('me') else '') + (' is-current' if cur else '')
    me = '<span class="ip-note">(tu)</span>' if u.get('me') else ''
    farms = u['farms'] if u['farms'] else '<span class="ip-user__all">toate fermele</span>'
    seen = (f'<td class="ip-user__seen">{u["seen"]}</td>' if u['seen'] else
            '<td class="ip-user__seen"><span class="ip-user__never">– încă n-a intrat</span></td>')
    lamp = ('<span class="ip-lamp ip-lamp--ok">activ</span>' if u['active'] else
            '<span class="ip-lamp ip-lamp--off">dezactivat</span>')
    mid = f'{p}-{u["key"]}'
    exp = 'true' if open_ else 'false'
    f = ' ip-focus' if focus else ''
    key = (f'<button type="button" class="ip-iconbtn{f}" aria-label="Acțiuni pentru {u["name"]}" aria-haspopup="menu" '
           f'aria-expanded="{exp}" aria-controls="{mid}-menu">{svg(IC["more"])}</button>')
    m = menu(p, u, active_item if open_ else None, open_, up)
    return (f'<tr class="{cls}"><td><span class="ip-user__name"><b>{u["name"]}</b>{me}</span></td>'
            f'<td class="ip-user__login">{u["login"]}</td><td>{u["role"]}</td><td class="ip-user__farms">{farms}</td>'
            f'{seen}<td>{lamp}</td><td class="ip-users__act">{key}{m}</td></tr>')

HEAD = ('<thead><tr><th scope="col">Nume</th><th scope="col">Login</th><th scope="col">Rol</th><th scope="col">Ferme</th>'
        '<th scope="col">Ultima intrare</th><th scope="col">Stare</th>'
        '<th scope="col" class="ip-users__acth"><span class="ip-sr">Acțiuni</span></th></tr></thead>')

def counts(users):
    a = sum(1 for u in users if u['active'])
    off = len(users) - a
    s = f'{num(len(users))} {"utilizator" if len(users) == 1 else "utilizatori"} · {num(a)} {"activ" if a == 1 else "activi"}'
    return s + (f' · {num(off)} {"dezactivat" if off == 1 else "dezactivați"}' if off else '')

def users(p, body, sumline, busy=False, extra='', table=True, live=''):
    b = ' aria-busy="true"' if busy else ''
    return (f'<section class="ip-users" aria-labelledby="{p}-h">'
            f'<div class="ip-list__head"><h3 class="ip-cap" id="{p}-h">Utilizatori</h3>'
            f'<p class="ip-note" id="{p}-sum">{sumline}</p>'
            f'<button type="button" class="ip-btn ip-btn--primary" aria-haspopup="dialog">{svg(IC["plus"])}Adaugă utilizator</button></div>'
            f'{extra}' + (f'<div class="ip-users__box"><table class="ip-table ip-users__t" aria-labelledby="{p}-h" aria-describedby="{p}-sum"{b}>'
            f'{HEAD}<tbody>{body}</tbody></table></div>' if table else '') +
            f'<p class="ip-sr" role="status" aria-live="polite">{live}</p></section>')

def list_rows(p, open_key=None, active_item=None, focus_key=None, data=ALL, up=False, cur_key=None):
    """The row menu opens downward; `up` (ip-users__menu--up) only where the room below the key, to the viewport or
    scroll container's bottom, is less than the menu's height and the room above is larger (S22P, D120). The app
    measures it on open; a still board or preview picks the one that fits."""
    return ''.join(row(p, u, open_=(u['key'] == open_key), active_item=active_item, focus=(u['key'] == focus_key),
                       up=up and u['key'] == open_key, cur=(u['key'] == cur_key)) for u in data)

# ------------------------------------------------------------------ dialogs
def banner(kind, body, id_=None):
    i = f' id="{id_}"' if id_ else ''
    return f'<div class="ip-banner ip-banner--{kind}"{i}>{svg(IC[kind])}<p class="ip-users__bn">{body}</p></div>'

def dialog(p, title, body, foot, desc=None):
    """Every SYS12 Dialog is role=alertdialog (S22P, D120): the confirms (reset, deactivate) as every confirm in the
    project, and the password step too, because the reset confirm turns into it in place (one role, never switched)
    and it holds a once-only warning the owner must act on (Închide).
    FY01 (FIN04, D124): the Dialog of every other screen: a bare body paragraph is the note `p.ip-dialog__note`
    (ink-secondary, §35; still the aria-describedby target), the keys sit in `div.ip-dialog__keys`, Anulează secondary."""
    if desc is None and f'id="{p}-once"' in body:   # the password step: the once rule is the alert (S22P fix 1)
        desc = f'{p}-once'
    if body.startswith('<p id="'):   # the confirm sentence
        body = '<p class="ip-dialog__note" id="' + body[len('<p id="'):]
    d = f' aria-describedby="{desc}"' if desc else ''
    return (f'<div class="ip-dialog" role="alertdialog" aria-modal="true" aria-labelledby="{p}-t"{d}>'
            f'<h2 id="{p}-t">{title}</h2>{body}<div class="ip-dialog__keys">{foot}</div></div>')

def btn(label, kind='secondary', focus=False, extra=''):   # FY01: Anulează is secondary, as in every other Dialog
    f = ' ip-focus' if focus else ''
    return f'<button type="button" class="ip-btn ip-btn--{kind}{f}"{extra}>{label}</button>'

def otp(p, ctx, pw, name, copied=False, failed=False, focus=None):
    """OneTimePassword: context line, label, the password in a real read-only TextField (input, Mono), copy key + done
    lamp, the once rule. Copy failed: the app calls input.focus() + input.select(); the preview marks it is-selected."""
    fpw = ' ip-focus is-selected' if focus == 'pw' else ''
    fcp = ' ip-focus' if focus == 'copy' else ''
    done = '<span class="ip-lamp ip-lamp--ok">copiată</span>' if copied else ''
    err = (f'<p class="ip-err ip-fieldmsg" id="{p}-err">{svg(IC["fault"], "ip-icon ip-icon--14")}'
           f'<span>Copierea nu a mers. Parola e selectată: apasă Ctrl+C.</span></p>' if failed else '')
    dpw = f'{p}-once' + (f' {p}-err' if failed else '')
    # copy failed (SYS13 rule, D119): focus moves to the field, which reads the error line; the status region stays empty
    live = 'Parola a fost copiată.' if copied else ''
    return (f'<div class="ip-otp">'
            f'<p class="ip-otp__ctx">{ctx}</p>'
            f'<label class="ip-label" for="{p}-pw">Parolă temporară</label>'
            f'<input type="text" readonly class="ip-input is-readonly ip-otp__pw{fpw}" id="{p}-pw" value="{pw}" '
            f'autocomplete="off" spellcheck="false" autocapitalize="off" aria-describedby="{dpw}">'
            f'<div class="ip-otp__keys"><button type="button" class="ip-btn ip-btn--secondary ip-otp__copy{fcp}">'
            f'{svg(IC["copy"])}Copiază parola</button>{done}</div>{err}'
            f'<div class="ip-banner ip-banner--info" id="{p}-once">{svg(IC["info"])}<p class="ip-users__bn">'
            f'<b>Se arată o singură dată.</b> Dă-i-o lui {name} acum. După ce închizi n-o mai poți vedea; '
            f'o poți doar reseta din nou.</p></div>'
            f'<p class="ip-sr" role="status" aria-live="polite">{live}</p></div>')

PW_RESET = 'k7mq4tzw9bhe'   # state example (D39): generated, a-z without i l o + 2-9, 12 characters
PW_NEW = 'r3xu8pdn6vka'     # state example (D39)

CTX_RESET = f'Parolă resetată · login <span class="ip-num">ion.rusu</span>'
CTX_NEW = f'Cont nou · Observator · Hîrbovăț · login <span class="ip-num">observator</span>'

# ------------------------------------------------------------------ blocks
L, O = [], []

L.append(('Listă · implicit (proprietarul vede toți utilizatorii)', users('a', list_rows('a'), counts(ALL))))
# current row (additions.css section 45): the user open in a panel beside the list; on a screen the row's Acțiuni key
# carries aria-current="true" (and the head key steps down to secondary while the panel is open); the preview uses tr.is-current
L.append(('Rând curent · Ion Rusu deschis în panou (§45; pe ecran aria-current pe cheia Acțiuni, demo tr.is-current)',
          users('n', list_rows('n', cur_key='rusu'), counts(ALL))))
L.append(('Meniu deschis · Ion Rusu (tastatura pe „Resetează parola”)',
          users('b', list_rows('b', 'rusu', 'Resetează parola'), counts(ALL))))
L.append(('Rândul tău · Dezactivează indisponibil, cu motiv',
          users('c', list_rows('c', 'own', 'Dezactivează'), counts(ALL))))
L.append(('Rând dezactivat · Petrea: Reactivează; Resetează parola indisponibil',
          users('d', list_rows('d', 'petrea', 'Reactivează'), counts(ALL))))
EMPTY = ('<tr class="ip-users__empty"><td colspan="7"><span class="ip-lamp ip-lamp--off">doar tu</span>'
         '<span>Doar tu ai acces. Adaugă primul utilizator.</span></td></tr>')
L.append(('Gol · doar proprietarul (A2 Empty)', users('e', row('e', OWNER) + EMPTY, counts([OWNER]))))
# FIN07: the DataTable skeleton row (additions.css §42): two tr.ip-table__skel rows, one bar per cell, the action bar at the end
# FIN11: the Acțiuni bar is an IconButton wide (§42 --key, control-h), so the column keeps its loaded 40px (a --w3 bar made it 80)
SK = ''.join('<tr aria-hidden="true" class="ip-table__skel">' + ''.join(
    f'<td><span class="ip-skel ip-skel--w{w}"></span></td>' for w in (2, 3, 3, 3, 2, 3))
    + '<td class="ip-users__act"><span class="ip-skel ip-skel--key ip-skel--end"></span></td></tr>' for _ in range(2))
L.append(('Se încarcă (skeleton, aria-busy)', users('f', SK + '<tr><td colspan="7" class="ip-users__srcell"><span class="ip-sr">Se încarcă utilizatorii</span></td></tr>',
                                                   '<span class="ip-users__wait">Se încarcă.</span>', busy=True)))
# FIN13: the Banner with action (additions.css §49, FIN03): role=alert, the S22 copy in two sentences, the key at the end;
# the list's status region stays empty (the alert announces it once), as on S22 eroare la încărcare; FX05: the D143 retry line
FAIL = ('<div class="ip-banner ip-banner--fault ip-banner--act" role="alert">' + svg(IC['fault']) + '<div class="ip-banner__body">'
        '<p>Utilizatorii nu s-au încărcat.</p><p>Verifică conexiunea la internet, apoi apasă Reîncearcă. Poți adăuga un utilizator nou și acum.</p></div>'
        + btn('Reîncearcă', 'secondary').replace('ip-btn--secondary"', 'ip-btn--secondary ip-banner__key"') + '</div>')
# FX06: the list head names no count while the list failed (as S22 since FIN07): the alert says it
L.append(('Eroare la încărcare (fără tabel)', users('g', '', '', extra=FAIL, table=False).replace('<p class="ip-note" id="g-sum"></p>', '', 1)))

# dialogs (the row menu's actions)
O.append(('Resetează parola · confirmare (focus pe Anulează)', dialog('h', 'Resetează parola lui Ion Rusu',
    '<p id="h-n">Parola de acum nu va mai merge. Noua parolă temporară se arată o singură dată, aici.</p>',
    btn('Anulează', focus=True) + btn('Resetează parola', 'primary'), desc='h-n')))
O.append(('Parola temporară · după resetare (focus pe Copiază parola)', dialog('i', 'Parola nouă pentru Ion Rusu',
    otp('i', CTX_RESET, PW_RESET, 'Ion Rusu', focus='copy'), btn('Închide', 'primary'))))
O.append(('Parola temporară · cont nou, după Copiază: lampă „copiată”', dialog('j', 'Parola temporară pentru Observator exemplu',
    otp('j', CTX_NEW, PW_NEW, 'Observator exemplu', copied=True), btn('Închide', 'primary'))))
O.append(('Copierea nu a mers (focus pe parolă, selectată)', dialog('k', 'Parola nouă pentru Ion Rusu',
    otp('k', CTX_RESET, PW_RESET, 'Ion Rusu', failed=True, focus='pw'), btn('Închide', 'primary'))))
O.append(('Resetarea nu a mers (focus pe Reîncearcă, Esc = Anulează)', dialog('l', 'Resetează parola lui Ion Rusu',
    banner('fault', '<b>Parola lui Ion Rusu nu s-a resetat;</b> cea veche merge în continuare. '
                    'Verifică conexiunea la internet, apoi apasă Reîncearcă.',
           'l-why'),
    btn('Anulează') + btn('Reîncearcă', 'primary', focus=True, extra=' aria-describedby="l-why"'), desc='l-why')))
O.append(('Dezactivează · confirmare (A2), focus pe Anulează; apoi Toast', dialog('m', 'Dezactivează contul lui Petrea',
    '<p id="m-n">Petrea nu va mai putea intra. Datele introduse de el rămân. Poți reactiva contul oricând.</p>',
    btn('Anulează', focus=True) + btn('Dezactivează', 'destructive'), desc='m-n')
    + '<div class="ip-toast" role="status" style="margin-top:var(--space-4)">Contul lui Petrea e dezactivat.</div>'))
O.append(('Reactivează · fără dialog, doar Toast', '<div class="ip-toast" role="status">Contul lui Petrea e activ din nou.</div>'))

W = 'calc(var(--space-12) * 22)'   # 1056px: the S22 content width at 1440 (sidebar + gutters)
D = 'var(--panel-w)'
ROOM = {2: 'var(--space-6)', 3: 'var(--space-6)', 4: 'calc(var(--space-12) * 4)'}   # menus open below
lst = ''.join(f'<div style="width:{W}' + (f';padding-bottom:{ROOM[i]}' if i in ROOM else '') + f'">{cap(t)}{h}</div>'
              for i, (t, h) in enumerate(L))
dlg = ''.join(f'<div style="width:{D}">{cap(t)}{h}</div>' for t, h in O)
FOOT = ('<p class="ip-note" style="margin:0;max-width:{w}">Din kit: numele Ion Rusu și Dr. Ceban (ilustrative, craft.md §7), '
        'Petrea și loginul petrea (spec A1/A2); fermele Telița și Hîrbovăț (reale). Exemple de stare (D39): rândurile '
        '„Proprietar exemplu” și „Observator exemplu”, celelalte loginuri, orele ultimei intrări, contul dezactivat al lui '
        'Petrea și cele două parole.</p>')

ur = ('<!-- @dsCard group="Data" height=2368 subtitle="SYS12" -->\n'
      '<div class="ip-root ip-stack" style="gap:var(--space-8)">\n'
      f'<div>{cap("S22 · Utilizatori · listă · desktop (lățime 1056 px)")}'
      f'<div class="ip-stack" style="gap:var(--space-8)">{lst}</div></div>\n'
      + FOOT.format(w=W) + '\n</div>\n')
ot = ('<!-- @dsCard group="Data" height=1380 subtitle="SYS12" -->\n'
      '<div class="ip-root ip-stack" style="gap:var(--space-8)">\n'
      f'<div>{cap("S22 · Utilizatori · dialoguri · parola temporară (Dialog, lățime 440 px)")}'
      f'<div style="display:grid;grid-template-columns:repeat(2,{D});gap:var(--space-8) var(--space-6);align-items:start">{dlg}</div></div>\n'
      + FOOT.format(w=f'calc({D} * 2 + var(--space-6))') + '\n</div>\n')
for comp, html in (('UserRow', ur), ('OneTimePassword', ot)):
    (ROOT / comp).mkdir(exist_ok=True)
    (ROOT / comp / 'preview.html').write_text(html, encoding='utf-8')
print('ok', len(L), len(O))
