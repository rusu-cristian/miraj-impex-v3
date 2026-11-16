"""Build SYS09 previews (EventRowActions, ChangeRow): python3 sys09_build.py"""
import pathlib
ROOT = pathlib.Path('/home/claude/miraj-impex-v3/design/system/components')
N = ' '  # thin no-break space for thousands

def svg(paths, cls='ip-icon'):
    return f'<svg class="{cls}" viewBox="0 0 24 24" aria-hidden="true">{paths}</svg>'

IC = {
    'mortality': '<path d="M12 3a7 7 0 0 0-7 7v4l2 2v3h10v-3l2-2v-4a7 7 0 0 0-7-7z"/><circle cx="9.5" cy="11" r="1.2"/><circle cx="14.5" cy="11" r="1.2"/><path d="M11 19v-2M13 19v-2"/>',
    'cull': '<circle cx="12" cy="12" r="9"/><path d="M5.7 5.7l12.6 12.6"/>',
    'eggs': '<path d="M12 22c4 0 7-3 7-7.5C19 9 16 2 12 2S5 9 5 14.5C5 19 8 22 12 22z"/>',
    'wheat': '<path d="M12 21V9M12 9c-3 0-4-2-4-5 3 0 4 2 4 5zM12 9c3 0 4-2 4-5-3 0-4 2-4 5zM12 15c-3 0-4-2-4-5 3 0 4 2 4 5zM12 15c3 0 4-2 4-5-3 0-4 2-4 5z"/>',
    'package': '<path d="M12 3 3 7.5v9L12 21l9-4.5v-9zM3 7.5 12 12l9-4.5M12 12v9"/>',
    'weighing': '<path d="M12 3v18"/><path d="M7 21h10"/><path d="M3 7h18"/><path d="m6 7-3 7a3 3 0 0 0 6 0z"/><path d="m18 7-3 7a3 3 0 0 0 6 0z"/>',
    'chev': '<path d="m9 18 6-6-6-6"/>',
    'more': '<circle cx="12" cy="5" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="12" cy="19" r="1"/>',  # Lucide ellipsis-vertical
}

def tile(rec, icon):
    return f'<span class="ip-tile ip-rec ip-rec--{rec}" aria-hidden="true">{svg(IC[icon])}</span>'

def sex(s, n=None):
    c = 'hen' if s == 'F' else 'rooster'
    g = '♀' if s == 'F' else '♂'
    return f'<span class="ip-sex ip-sex--{c}">{g}' + (f' <b>{n}</b>' if n is not None else '') + '</span>'

def cap(text):
    return f'<span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">{text}</span>'

def num(t):
    return f'<span class="ip-num">{t}</span>'

# ------------------------------------------------------------------ records (sample-hirbovat-2026.json, D27)
R = {
    'm11': dict(tile=('mortality', 'mortality'), name='Mortalitate H1.1',
                title=f'Mortalitate · H1.1 · {sex("F", "−1")} · {sex("M", "0")}', meta=f'{num("08:34")} · Dr. Ceban'),
    'm42': dict(tile=('mortality', 'mortality'), name='Mortalitate H4.2',
                title=f'Mortalitate · H4.2 · {sex("F", "−3")} · {sex("M", "0")}', meta=f'{num("08:34")} · Dr. Ceban'),
    'egg1': dict(tile=('eggs', 'eggs'), name='Ouă H1',
                 title=f'Ouă · H1 · {num(f"6{N}488")} ouă · Jos {num("345")}', meta=f'{num("08:35")} · Dr. Ceban · ouat {num("75,8%")}'),
    'feed': dict(tile=('feed', 'wheat'), name='Hrană zilnică Hîrbovăț',
                 title=f'Hrană zilnică · Hîrbovăț · {num(f"3{N}691,3 kg")}', meta=f'{num("09:47")} · Dr. Ceban · {num("10")} din {num("12")} rânduri'),
    'weigh': dict(tile=('weighing', 'weighing'), name='Cântărire Hîrbovăț',
                  title=f'Cântărire · Hîrbovăț · {num("6")} secțiuni · <span class="ip-ev__sexes">{sex("F")} și {sex("M")}</span>', meta=f'{num("10:10")} · Dr. Ceban · {num("12")} rânduri'),
    'cull31': dict(tile=('cull', 'cull'), name='Sacrificare H3.1, 15.11.2026',
                   title=f'Sacrificare · H3.1 · {sex("M", "−2")} · boală', meta=f'{num("15.11.2026 · SAPT 40")} · Dr. Ceban'),
    'm41': dict(tile=('mortality', 'mortality'), name='Mortalitate H4.1, 14.11.2026',
                title=f'Mortalitate · H4.1 · {sex("F", "−5")} · {sex("M", "0")}', meta=f'{num("14.11.2026 · SAPT 40")} · Dr. Ceban'),
    'del14': dict(tile=('feed', 'package'), name='Livrare hrană Breeder 2, 14.11.2026',
                  title=f'Livrare hrană · Breeder&nbsp;2 · {num(f"24{N}000 kg")}', meta=f'{num("14.11.2026")} · Hîrbovăț · Ion Rusu · fără preț'),
}
MOD41 = (f'<span class="ip-ev__mod"><span class="ip-tag ip-tag--neutral">corectat</span>'
         f'<a class="ip-link" href="#">modificat de Ion Rusu · {num("15.11.2026, 10:12")}<span class="ip-sr">, vezi istoricul</span></a></span>')

def acts(rec, kind, state=('', ''), disabled=False, desc=''):
    """kind: 'edit' (author, same day) or 'fix' (manager/owner, past day)."""
    k1 = 'Editează' if kind == 'edit' else 'Corectează'
    sr = f'<span class="ip-sr"> {rec["name"]}</span>'
    d = f' aria-disabled="true" aria-describedby="{desc}"' if disabled else ''
    s1 = f' {state[0]}' if state[0] else ''
    s2 = f' {state[1]}' if state[1] else ''
    a1 = (f'<a class="ip-btn ip-btn--ghost{s1}" role="link" tabindex="0"{d}>{k1}{sr}</a>' if disabled
          else f'<a class="ip-btn ip-btn--ghost{s1}" href="#">{k1}{sr}</a>')
    b2 = f'<button type="button" class="ip-btn ip-btn--ghost{s2}" aria-haspopup="dialog"{d}>Șterge{sr}</button>'
    return f'<span class="ip-ev__acts">{a1}{b2}</span>'

def more(rec, kind, mid, open_=False, disabled=False, desc='', kstate=''):
    k1 = 'Editează' if kind == 'edit' else 'Corectează'
    exp = 'true' if open_ else 'false'
    key = (f'<button type="button" class="ip-iconbtn ip-iconbtn--touch{" " + kstate if kstate else ""}" aria-label="Acțiuni, {rec["name"]}" '
           f'aria-haspopup="menu" aria-expanded="{exp}" aria-controls="{mid}">{svg(IC["more"])}</button>')
    nid = f'{mid}-n'
    note = ('Timpul de modificare s-a încheiat la <span class="ip-num">23:59</span>. Managerul poate corecta.' if disabled
            else ('Poți modifica sau șterge până la <span class="ip-num">23:59</span>.' if kind == 'edit'
                  else 'Corecția și ștergerea cer un motiv și intră în Istoric modificări.'))
    d = f' aria-disabled="true" aria-describedby="{nid}"' if disabled else ''
    a1 = (f'<a class="ip-menu__item" role="menuitem" tabindex="-1"{d}>{k1}</a>' if disabled
          else f'<a class="ip-menu__item{" is-active" if open_ else ""}" role="menuitem" tabindex="-1" href="#">{k1}</a>')
    menu = (f'<div class="ip-menu ip-ev__menu"{"" if open_ else " hidden"}>'
            f'<div class="ip-ev__items" id="{mid}" role="menu" aria-label="Acțiuni, {rec["name"]}" aria-describedby="{nid}">'
            f'{a1}<button type="button" class="ip-menu__item" role="menuitem" tabindex="-1"{d}>Șterge</button></div>'
            f'<p class="ip-note" id="{nid}">{note}</p></div>')
    return f'<span class="ip-ev__more">{key}{menu}</span>'

def row(key, kind=None, *, mod='', meta_extra='', locked=False, state=('', ''), disabled=False, desc='',
        mid='', open_=False, phone_menu_disabled=False, rowstate=''):
    r = R[key]
    meta = r['meta'] + meta_extra
    main = (f'<span class="ip-ev__main"><span class="ip-ev__title">{r["title"]}</span>'
            f'<span class="ip-ev__meta">{meta}</span>{mod}</span>')
    slot = ''
    if locked:
        slot = '<span class="ip-ev__lock"><span class="ip-tag ip-tag--neutral">blocat</span></span>'
    elif kind:
        slot = acts(r, kind, state, disabled, desc) + more(r, kind, mid or f'mn-{key}', open_, phone_menu_disabled or disabled, '', state[0] or state[1])
    cls = 'ip-ev ip-ev--act' + (f' {rowstate}' if rowstate else '')
    return f'<li class="{cls}">{tile(*r["tile"])}{main}{slot}</li>'

def deleted_row():
    t = f'Sacrificare · H3.1 · <del>{sex("M", "−2")} · boală</del>'
    meta = f'{num("15.11.2026 · SAPT 40")} · Dr. Ceban'
    mod = (f'<span class="ip-ev__mod"><span class="ip-tag ip-tag--neutral">șters</span>'
           f'<a class="ip-link" href="#">șters de Ion Rusu · {num("16.11.2026, 11:05")}<span class="ip-sr">, vezi istoricul</span></a></span>')
    return (f'<li class="ip-ev ip-ev--act ip-ev--deleted">{tile("cull", "cull")}<span class="ip-ev__main">'
            f'<span class="ip-ev__title">{t}<span class="ip-sr"> (șters)</span></span><span class="ip-ev__meta">{meta}</span>{mod}</span></li>')

def lst(rows, label, extra=''):
    return f'<ul class="ip-list ip-list--act" aria-label="{label}"{extra}>{"".join(rows)}</ul>'

def head(title, note):
    return f'<div class="ip-list__head"><h2 class="ip-cap">{title}</h2><p class="ip-note">{note}</p></div>'

def act_blocks(s):
    out = []
    out.append(cap('Veterinar · Introdus azi · propriile intrări, până la 23:59: Editează · Șterge') +
               head('Introdus azi', f'{num("16.11.2026")} · poți modifica sau șterge până la {num("23:59")}') +
               lst([row('m11', 'edit', mid=f'a1{s}'), row('m42', 'edit', mid=f'a2{s}'), row('egg1', 'edit', mid=f'a3{s}'),
                    row('feed', 'edit', mid=f'a4{s}'), row('weigh', 'edit', mid=f'a5{s}')], 'Introdus azi'))
    out.append(cap('Manager · jurnal, zile trecute: Corectează · Șterge (motiv cerut) · rând corectat: Tag + „modificat de”') +
               lst([row('cull31', 'fix', mid=f'b1{s}'), row('m41', 'fix', mod=MOD41, mid=f'b2{s}'), row('del14', 'fix', mid=f'b3{s}')],
                   'Jurnal Hîrbovăț'))
    out.append(cap('Veterinar · zile trecute: propriile intrări blocat, cu motivul · intrarea altcuiva: fără acțiuni (D71) · Observator: fără acțiuni') +
               lst([row('cull31', locked=True, meta_extra=' · zi trecută · o corectează managerul'),
                    row('m41', locked=True, mod=MOD41, meta_extra=' · zi trecută · o corectează managerul'), row('del14')], 'Jurnal, veterinar') +
               '<div style="height:var(--space-3)"></div>' +
               lst([row('m41', mod=MOD41), row('del14')], 'Jurnal, observator'))
    out.append(cap('Stări: tastă hover · apăsată · focus · dezactivată cu motiv (23:59 a trecut cu pagina deschisă) · șters') +
               lst([row('m11', 'edit', state=('is-hover', ''), mid=f'c1{s}'),
                    row('m11', 'edit', state=('', 'is-pressed'), mid=f'c2{s}'),
                    row('m11', 'edit', state=('ip-focus', ''), mid=f'c3{s}'),
                    row('m42', 'edit', disabled=True, desc=f'why{s}', mid=f'c4{s}',
                        meta_extra=f' · <span class="ip-ev__why" id="why{s}">timpul de modificare s-a încheiat la {num("23:59")} · o corectează managerul</span>'),
                    deleted_row()], 'Stări'))
    return out

def act_phone_menus(s):
    out = []
    out.append(cap('Telefon · meniul Acțiuni deschis (veterinar, azi)') +
               lst([row('m11', 'edit', mid=f'p1{s}', open_=True), row('m42', 'edit', mid=f'p2{s}'), row('egg1', 'edit', mid=f'p3{s}')],
                   'Introdus azi, meniu deschis'))
    out.append('<div style="height:calc(var(--space-12) * 2)"></div>' + cap('Telefon · meniul deschis, dezactivat cu motiv · manager: Corectează') +
               lst([row('m42', 'edit', mid=f'p4{s}', open_=True, phone_menu_disabled=True,
                        meta_extra=f' · <span class="ip-ev__why">timpul de modificare s-a încheiat la {num("23:59")}</span>'),
                    row('cull31', 'fix', mid=f'p5{s}'), row('weigh', 'edit', mid=f'p6{s}')], 'Meniu dezactivat'))
    out.append('<div style="height:calc(var(--space-12) * 2)"></div>' + cap('Telefon · manager, meniul Corectează deschis') +
               lst([row('m41', 'fix', mod=MOD41, mid=f'p7{s}', open_=True), row('cull31', 'fix', mid=f'p8{s}'), row('del14', 'fix', mid=f'p9{s}')], 'Meniu manager'))
    return out

def build_actions():
    d = act_blocks('-d'); m = act_blocks('-m') + act_phone_menus('-m')
    desk = ''.join(f'<div>{x}</div>' for x in d)
    phone = ''.join(f'<div>{x}</div>' for x in m)
    html = ('<!-- @dsCard group="Data" height=2900 subtitle="SYS09" -->\n'
        '<div class="ip-root ip-stack" style="gap:var(--space-6)">\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">Desktop · listă de 720 px (taste în rând, ≥ 480 px)</span><div class="ip-stack" style="width:720px;gap:var(--space-5)">{desk}</div></div>\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">390 · telefon (cheia Acțiuni, &lt; 480 px)</span><div class="ip-phone" style="width:390px"><div class="ip-stack" style="padding:var(--space-4);gap:var(--space-5)">{phone}</div></div></div>\n'
        f'<p class="ip-note" style="margin:0">Hîrbovăț, sample (D27): sample-hirbovat-2026.json › daily 16.11.2026 (Mortalitate H1.1 ♀ 1, H4.2 ♀ 3), eggs_daily (H1 6{N}488, Jos 345, 75,8%), allocations_today (10 din 12 rânduri · 3{N}691,3 kg), event_log (08:34, 08:35, 09:47, 10:10 Dr. Ceban; livrare 14.11.2026 24{N}000 kg Ion Rusu), culls (H3.1 ♂ 2 boală, 15.11.2026), audit_log (H4.1 ♀ 14.11.2026 corectat de Ion Rusu pe 15.11.2026, 10:12). Exemple de stare (D39): tastele dezactivate după 23:59 și rândul șters (Ion Rusu, 16.11.2026, 11:05) nu sunt în date.</p>\n'
        '</div>\n')
    (ROOT / 'EventRowActions').mkdir(exist_ok=True)
    (ROOT / 'EventRowActions' / 'preview.html').write_text(html, encoding='utf-8')

# ------------------------------------------------------------------ ChangeRow
def val(k, old, new, deleted=False):
    newv = '–' if deleted else new
    sr_new = 'șters' if deleted else newv
    return (f'<li class="ip-chg__val"><span class="ip-chg__k">{k}</span>'
            f'<del><span class="ip-sr">înainte </span>{old}</del><span class="ip-chg__arrow" aria-hidden="true">→</span>'
            f'<ins><span class="ip-sr">, după </span>{newv if not deleted else "<span aria-hidden=" + chr(34) + "true" + chr(34) + ">–</span><span class=" + chr(34) + "ip-sr" + chr(34) + ">șters</span>"}</ins></li>')

def chg(rec_tile, link, sr_link, meta, vals, why, when, who, role, tag, state=''):
    whyh = (f'<p class="ip-chg__why"><span class="ip-chg__whyk">Motiv: </span>{why}</p>' if why
            else '<p class="ip-chg__why is-none"><span class="ip-chg__whyk">Motiv: </span>– în aceeași zi, fără motiv</p>')
    d, t = when
    cls = 'ip-ev ip-ev--change' + (f' {state}' if state else '')
    return (f'<li class="{cls}">{tile(*rec_tile)}'
            f'<span class="ip-ev__main"><a class="ip-chg__link" href="#">{link}<span class="ip-sr">{sr_link}</span></a>'
            f'<span class="ip-ev__meta">{meta}</span></span>'
            f'<ul class="ip-chg__vals">{"".join(vals)}</ul>{whyh}'
            f'<span class="ip-chg__meta"><span class="ip-chg__when">{d}<span class="ip-chg__dot">, </span><span>{t}</span></span>'
            f'<span class="ip-chg__dot" aria-hidden="true">·</span>'
            f'<span class="ip-chg__who"><span>{who}</span><span class="ip-chg__dot" aria-hidden="true">·</span><span class="ip-note">{role}</span></span></span>'
            f'<span class="ip-tag ip-tag--neutral">{tag}</span>'
            f'<span class="ip-ev__go">{svg(IC["chev"], "ip-icon ip-icon--16")}</span></li>')

def changes(state=''):
    a = chg(('mortality', 'mortality'), f'Mortalitate · H4.1 {sex("F")}', ' găini, ziua 14.11.2026',
            f'ziua {num("14.11.2026 · SAPT 40")} · Hîrbovăț',
            [val('găini moarte', '3', '5')], 'Numărătoare refăcută: 2 găini găsite moarte sub hrănitoare.',
            ('15.11.2026', '10:12'), 'Ion Rusu', 'manager', 'corectat', state)
    b = chg(('cull', 'cull'), f'Sacrificare · H3.1 {sex("M")}', ' cocoși, ziua 15.11.2026',
            f'ziua {num("15.11.2026 · SAPT 40")} · Hîrbovăț',
            [val('cocoși sacrificați', '2', '', True), val('motiv sacrificare', 'boală', '', True)],
            'Sacrificare înregistrată din greșeală.', ('16.11.2026', '11:05'), 'Ion Rusu', 'manager', 'șters')
    c = chg(('feed', 'wheat'), f'Hrană zilnică · H1.2 {sex("M")}', ' cocoși, ziua 16.11.2026',
            f'ziua {num("16.11.2026 · SAPT 40")} · Hîrbovăț · Cocoși',
            [val('g/pasăre', f'155{N}g', f'172{N}g'), val('kg', f'67,7{N}kg', f'75,2{N}kg')], '',
            ('16.11.2026', '09:52'), 'Dr. Ceban', 'veterinar', 'editat')
    d = chg(('weighing', 'weighing'), f'Cântărire · H1.1 {sex("F")}', ' găini, ziua 16.11.2026',
            f'ziua {num("16.11.2026 · SAPT 40")} · Hîrbovăț',
            [val('greutate medie', f'3{N}967{N}g', '', True), val('CV', '7,3%', '', True), val('uniformitate', '81,9%', '', True)], '',
            ('16.11.2026', '10:14'), 'Dr. Ceban', 'veterinar', 'șters')
    return [a, b, c, d]

COLS = ('<li class="ip-chg__cols" aria-hidden="true"><span class="ip-cap">Când</span><span class="ip-cap">Înregistrare</span>'
        '<span class="ip-cap">Valori · vechi → nou</span><span class="ip-cap">Motiv</span><span class="ip-cap">Cine</span><span class="ip-cap">Tip</span></li>')

def clist(rows, label, cols=True):
    return f'<ul class="ip-list ip-list--change" aria-label="{label}">{COLS if cols else ""}{"".join(rows)}</ul>'

def change_blocks(s):
    rows = changes()
    out = []
    out.append(cap('Tipuri: corectat (Corectează, real) · șters cu motiv (manager) · editat în aceeași zi · șters în aceeași zi (mai multe câmpuri)') +
               clist(rows, 'Istoric modificări'))
    a = changes()[0]
    out.append(cap('Stări: hover · apăsat · focus (rândul întreg deschide înregistrarea)') +
               clist([a.replace('ip-ev ip-ev--change', 'ip-ev ip-ev--change is-hover', 1),
                      a.replace('ip-ev ip-ev--change', 'ip-ev ip-ev--change is-pressed', 1),
                      a.replace('ip-ev ip-ev--change', 'ip-ev ip-ev--change ip-focus', 1)], 'Stări'))
    out.append(cap('Nicio modificare în perioadă · se încarcă') +
               '<div class="ip-stack" style="gap:var(--space-4)">'
               f'<ul class="ip-list ip-list--change" aria-label="Istoric modificări"><li class="ip-list__empty"><span class="ip-lamp ip-lamp--off">fără modificări</span>'
               f'<span>Nicio modificare la Hîrbovăț între {num("09.11.2026")} și {num("16.11.2026")}.</span><span>Schimbă perioada sau filtrele.</span></li></ul>'
               '<ul class="ip-list ip-list--change" aria-label="Istoric modificări" aria-busy="true">'
               + ''.join('<li class="ip-ev ip-ev--skel" aria-hidden="true"><span class="ip-skel ip-skel--tile"></span><span class="ip-ev__main"><span class="ip-skel ip-skel--w1"></span><span class="ip-skel ip-skel--w2"></span></span><span class="ip-skel ip-skel--w3"></span></li>' for _ in range(2))
               + '<li class="ip-sr">Se încarcă</li></ul></div>')
    return out

def build_change():
    d = change_blocks('-d'); m = change_blocks('-m')
    desk = ''.join(f'<div>{x}</div>' for x in d)
    phone = ''.join(f'<div>{x}</div>' for x in m)
    html = ('<!-- @dsCard group="Data" height=1900 subtitle="SYS09" -->\n'
        '<div class="ip-root ip-stack" style="gap:var(--space-6)">\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">S17 · Istoric modificări · desktop (listă de 1136 px, un rând pe modificare, ≥ 720 px)</span><div class="ip-stack" style="width:1136px;gap:var(--space-5)">{desk}</div></div>\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">390 · telefon (stivuit)</span><div class="ip-phone" style="width:390px"><div class="ip-stack" style="padding:var(--space-4);gap:var(--space-5)">{phone}</div></div></div>\n'
        f'<p class="ip-note" style="margin:0">Real în sample (D27): sample-hirbovat-2026.json › audit_log, singura modificare: Mortalitate H4.1 ♀ din 14.11.2026, 3 → 5, Ion Rusu, 15.11.2026, 10:12, cu motivul de mai sus. Exemple de stare (D39), construite din cifre reale: ștergerea Sacrificării H3.1 ♂ 2 · boală (culls 15.11.2026), editarea H1.2 ♂ 155 → 172{N}g (allocations_today: ieri 155, azi 172; kg = g × 437 cocoși) și ștergerea Cântăririi H1.1 ♀ (weighings 16.11.2026); orele 11:05, 09:52, 10:14 și motivul ștergerii nu sunt în date.</p>\n'
        '</div>\n')
    (ROOT / 'ChangeRow').mkdir(exist_ok=True)
    (ROOT / 'ChangeRow' / 'preview.html').write_text(html, encoding='utf-8')

build_actions(); build_change()
print('ok')
