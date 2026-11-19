"""Build SYS09 previews (EventRowActions, ChangeRow): python3 sys09_build.py"""
import pathlib, datetime as dt
ROOT = pathlib.Path(__file__).resolve().parents[1] / 'components'  # from this script's place (design/system/scripts), so a copy writes into itself
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

# ------------------------------------------------------------------ records (RF09: realdata + the illustrative events, today 18.11.2026)
import sys
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[2] / 'canvas/scripts'))
import sexsign as SXS  # SX01: every ♀/♂ of the previews = the §72 hook
import realdata as D                                   # the design's one data module (pure data, no HTML)
n, dd, T = D.n, D.dd, D.TODAY
FARM = D.FARMS['hirbovat']['name']
def ev(rec, day): return next(e for e in D.EVENTS if e['record'] == rec and D.D(e['at']) == D.D(day))
def hm(e): return e['at'][11:16]
def secs(day): return [x for x in D.HSECS if D.sd(day, x, 'F') or D.sd(day, x, 'M')]
def dsapt(day): return num(f'{dd(day)} · SAPT {D.sapt(day)}')
# today's three saves of the vet (as S2/S15 Introdus azi): Zilnic, Ouă, Hrană zilnică (D168/D172: one save = one record)
EM, EE, EF = ev('mortality', T), ev('eggs', T), ev('feed', T)
assert {e['user'] for e in (EM, EE, EF)} == {'Dr. Ceban'}
DF, DM = D.deaths(T, sx='F'), D.deaths(T, sx='M')
NOCULL = not [c for c in D.ILL['culls'] if D.D(c['date']) == T]
EH = [h for h in D.HHALLS if D.eggs(T, h)]
KG = {sx: D.feed_kg(T, sx) for sx in 'FM'}; ROWS = [(x, sx) for x in secs(T) for sx in 'FM']; FED = [k for k in ROWS if D.sd(T, *k)['kg']]
# a past-day cull (illustrative, carved out of real deaths) and the corrected save (D158a)
CU = max(D.ILL['culls'], key=lambda c: c['date']); CU_SX = {'F': 'găini', 'M': 'cocoși'}[CU['sex']]
COR = D.ILL['corrections'][0]; CD = COR['row']['date']; CE = next(e for e in D.EVENTS if e['at'] == COR['entry_at'])
CDF, CDM = D.deaths(CD, sx='F'), D.deaths(CD, sx='M')
# the last delivery with no price (real)
DV = max(D.UNPRICED, key=lambda r: r['date']); DV_E = ev('feed_delivery', DV['date']); DV_T = D.FEED_TYPES[DV['feed_type_id']]['name'].replace(' ', '&nbsp;')
assert 'fără preț' in DV_E['summary'] and f'{DV["qty_kg"]} kg' in DV_E['summary']
R = {
    'mz': dict(tile=('mortality', 'mortality'), name='Mortalitate de azi',
               title=f'Mortalitate · {num(len(secs(T)))} secțiuni · {sex("F", n(-DF))} · {sex("M", n(-DM))}',
               meta=f'{num(hm(EM))} · {EM["user"]}' + (' · Zilnic, fără sacrificări' if NOCULL else '')),
    'egg': dict(tile=('eggs', 'eggs'), name='Ouă de azi',
                title=f'Ouă · {", ".join(EH)} · {num(n(D.eggs(T)))} ouă', meta=f'{num(hm(EE))} · {EE["user"]} · ouat {num(D.pct(D.lay_day(T)))}'),
    'feed': dict(tile=('feed', 'wheat'), name='Hrană zilnică de azi',
                 title=f'Hrană zilnică · {num(len(FED))} din {num(len(ROWS))} rânduri · {num(n(KG["F"] + KG["M"], 1) + " kg")}',
                 meta=f'{num(hm(EF))} · {EF["user"]} · {sex("F", n(KG["F"], 1) + " kg")} · {sex("M", n(KG["M"], 1) + " kg")}'),
    'cull': dict(tile=('cull', 'cull'), name=f'Sacrificare {CU["section"]}, {dd(CU["date"])}',
                 title=f'Sacrificare · {CU["section"]} · {sex(CU["sex"], n(-CU["qty"]))} · {CU["reason"]}', meta=f'{dsapt(CU["date"])} · {CU["created_by"]}'),
    'mcor': dict(tile=('mortality', 'mortality'), name=f'Mortalitate, {dd(CD)}',
                 title=f'Mortalitate · {num(len(secs(CD)))} secțiuni · {sex("F", n(-CDF))} · {sex("M", n(-CDM))}', meta=f'{dsapt(CD)} · {CE["user"]}'),
    'dlv': dict(tile=('feed', 'package'), name=f'Livrare hrană {D.FEED_TYPES[DV["feed_type_id"]]["name"]}, {dd(DV["date"])}',
                title=f'Livrare hrană · {DV_T} · {num(n(DV["qty_kg"]) + " kg")}', meta=f'{num(dd(DV["date"]))} · {FARM} · {DV_E["user"]} · fără preț'),
}
assert f'mortalitate ♀ {CDF} · ♂ {CDM}' in CE['summary'] and f'mortalitate ♀ {DF} · ♂ {DM}' in EM['summary'] and f'{D.eggs(T)} total' in EE['summary']
MODC = (f'<span class="ip-ev__mod"><span class="ip-tag ip-tag--neutral">corectat</span>'
        f'<a class="ip-link" href="#">modificat de {COR["user"]} · {num(D.ddt(COR["at"]))}<span class="ip-sr">, vezi istoricul</span></a></span>')
MODT = (f'<span class="ip-ev__mod"><span class="ip-tag ip-tag--neutral">corectat</span>'   # D104/D105 (D178): plain text for roles without Istoric (vet, observer)
        f'<span>modificat de {COR["user"]} · {num(D.ddt(COR["at"]))}</span></span>')
DEL_BY, DEL_AT = 'Ion Rusu', f'{T.isoformat()}T11:05'   # D39: the deletion of a real record (who and when are the example)
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
    t = f'Sacrificare · {CU["section"]} · <del>{sex(CU["sex"], n(-CU["qty"]))} · {CU["reason"]}</del>'
    meta = R['cull']['meta']
    mod = (f'<span class="ip-ev__mod"><span class="ip-tag ip-tag--neutral">șters</span>'
           f'<a class="ip-link" href="#">șters de {DEL_BY} · {num(D.ddt(DEL_AT))}<span class="ip-sr">, vezi istoricul</span></a></span>')
    return (f'<li class="ip-ev ip-ev--act ip-ev--deleted">{tile("cull", "cull")}<span class="ip-ev__main">'
            f'<span class="ip-ev__title">{t}<span class="ip-sr"> (șters)</span></span><span class="ip-ev__meta">{meta}</span>{mod}</span></li>')

def lst(rows, label, extra=''):
    return f'<ul class="ip-list ip-list--act" aria-label="{label}"{extra}>{"".join(rows)}</ul>'

def head(title, note):
    return f'<div class="ip-list__head"><h2 class="ip-cap">{title}</h2><p class="ip-note">{note}</p></div>'

def act_blocks(s):
    out = []
    out.append(cap('Veterinar · Introdus azi · propriile intrări, până la 23:59: Editează · Șterge') +
               head('Introdus azi', f'{num(dd(T))} · poți modifica sau șterge până la {num("23:59")}') +
               lst([row('mz', 'edit', mid=f'a1{s}'), row('egg', 'edit', mid=f'a2{s}'), row('feed', 'edit', mid=f'a3{s}')], 'Introdus azi'))
    out.append(cap('Manager · jurnal, zile trecute: Corectează · Șterge (motiv cerut) · rând corectat: Tag + „modificat de”') +
               lst([row('mcor', 'fix', mod=MODC, mid=f'b1{s}'), row('dlv', 'fix', mid=f'b2{s}'), row('cull', 'fix', mid=f'b3{s}')],
                   f'Jurnal {FARM}'))
    out.append(cap('Veterinar · zile trecute: propriile intrări blocat, cu motivul · intrarea altcuiva: fără acțiuni (D71) · Observator: fără acțiuni · „modificat de” ca text (fără Istoric, D104)') +
               lst([row('mcor', locked=True, mod=MODT, meta_extra=' · zi trecută · o corectează managerul'),
                    row('dlv'), row('cull', locked=True, meta_extra=' · zi trecută · o corectează managerul')], 'Jurnal, veterinar') +
               '<div style="height:var(--space-3)"></div>' +
               lst([row('mcor', mod=MODT), row('dlv')], 'Jurnal, observator'))
    out.append(cap('Stări: tastă hover · apăsată · focus · dezactivată cu motiv (23:59 a trecut cu pagina deschisă) · șters') +
               lst([row('mz', 'edit', state=('is-hover', ''), mid=f'c1{s}'),
                    row('mz', 'edit', state=('', 'is-pressed'), mid=f'c2{s}'),
                    row('mz', 'edit', state=('ip-focus', ''), mid=f'c3{s}'),
                    row('egg', 'edit', disabled=True, desc=f'why{s}', mid=f'c4{s}',
                        meta_extra=f' · <span class="ip-ev__why" id="why{s}">timpul de modificare s-a încheiat la {num("23:59")} · o corectează managerul</span>'),
                    deleted_row()], 'Stări'))
    return out

def act_phone_menus(s):
    out = []
    out.append(cap('Telefon · meniul Acțiuni deschis (veterinar, azi)') +
               lst([row('mz', 'edit', mid=f'p1{s}', open_=True), row('egg', 'edit', mid=f'p2{s}'), row('feed', 'edit', mid=f'p3{s}')],
                   'Introdus azi, meniu deschis'))
    out.append('<div style="height:calc(var(--space-12) * 2)"></div>' + cap('Telefon · meniul deschis, dezactivat cu motiv · manager: Corectează') +
               lst([row('egg', 'edit', mid=f'p4{s}', open_=True, phone_menu_disabled=True,
                        meta_extra=f' · <span class="ip-ev__why">timpul de modificare s-a încheiat la {num("23:59")}</span>'),
                    row('cull', 'fix', mid=f'p5{s}'), row('feed', 'edit', mid=f'p6{s}')], 'Meniu dezactivat'))
    out.append('<div style="height:calc(var(--space-12) * 2)"></div>' + cap('Telefon · manager, meniul Corectează deschis') +
               lst([row('mcor', 'fix', mod=MODC, mid=f'p7{s}', open_=True), row('dlv', 'fix', mid=f'p8{s}'), row('cull', 'fix', mid=f'p9{s}')], 'Meniu manager'))
    return out

def build_actions():
    d = act_blocks('-d'); m = act_blocks('-m') + act_phone_menus('-m')
    desk = ''.join(f'<div>{x}</div>' for x in d)
    phone = ''.join(f'<div>{x}</div>' for x in m)
    html = (f'<!-- @dsCard group="Data" height={H_ACT} subtitle="SYS09" -->\n'
        '<div class="ip-root ip-stack" style="gap:var(--space-6)">\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">Desktop · listă de 720 px (taste în rând, ≥ 480 px)</span><div class="ip-stack" style="width:720px;gap:var(--space-5)">{desk}</div></div>\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">390 · telefon (cheia Acțiuni, &lt; 480 px)</span><div class="ip-phone" style="width:390px"><div class="ip-stack" style="padding:var(--space-4);gap:var(--space-5)">{phone}</div></div></div>\n'
        f'<p class="ip-note" style="margin:0">{NOTE_ACT}</p>\n'
        '</div>\n')
    (ROOT / 'EventRowActions').mkdir(exist_ok=True)
    (ROOT / 'EventRowActions' / 'preview.html').write_text(SXS.sexify(html), encoding='utf-8')

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

# the illustrative correction (D158a) on the whole save (D172), then D39 examples on real records: the cull deleted with a reason,
# the vet's same-day edit of today's Zilnic (as S15 după 23:59: one compartment, D39 old values and time), a weighing deleted
# the day it was saved (D39)
CS = secs(CD); SXW = {'F': 'găini moarte', 'M': 'cocoși morți'}
ED_AT = f'{T.isoformat()}T09:52'                       # D39 (S15 după 23:59 "modificat la 09:52")
ED_SEC = secs(T)[0]; ED_NOW = {sx: D.sd(T, ED_SEC, sx)['mort'] for sx in 'FM'}; ED_OLD = {'F': ED_NOW['F'] + 1, 'M': ED_NOW['M'] + 1}   # D39 old values
assert hm(EM) < ED_AT[11:] < COR['at'][11:]
WG = min(D.WEIGHINGS_HIR, key=lambda w: w['created_at']); WG_DEL = WG['created_at'][:11] + f"{int(WG['created_at'][11:13]):02d}:{int(WG['created_at'][14:16]) + 4:02d}"   # D39: deleted 4 min after its save
def changes(state=''):
    a = chg(('mortality', 'mortality'), f'Mortalitate · {len(CS)} secțiuni', f', ziua {dd(CD)}',
            f'ziua {dsapt(CD)} · {FARM}',
            [val(f'{COR["row"]["section"]} · {SXW[COR["row"]["sex"]]}', n(COR['old_values']['qty']), n(COR['new_values']['qty']))], COR['reason'],
            (dd(COR['at']), COR['at'][11:16]), COR['user'], 'manager', 'corectat', state)
    b = chg(('cull', 'cull'), f'Sacrificare · {CU["section"]} {sex(CU["sex"])}', f' {CU_SX}, ziua {dd(CU["date"])}',
            f'ziua {dsapt(CU["date"])} · {FARM}',
            [val(f'{CU_SX} sacrificați' if CU['sex'] == 'M' else f'{CU_SX} sacrificate', n(CU['qty']), '', True), val('motiv sacrificare', CU['reason'], '', True)],
            'Sacrificare înregistrată din greșeală.', (dd(DEL_AT), DEL_AT[11:]), DEL_BY, 'manager', 'șters')
    c = chg(('mortality', 'mortality'), f'Mortalitate · {len(secs(T))} secțiuni', f', ziua {dd(T)}',
            f'ziua {dsapt(T)} · {FARM}',
            [val(f'{ED_SEC} · {SXW[sx]}', n(ED_OLD[sx]), n(ED_NOW[sx])) for sx in 'FM'], '',
            (dd(T), ED_AT[11:]), EM['user'], 'veterinar', 'editat')
    d = chg(('weighing', 'weighing'), f'Cântărire · {WG["section"]} {sex(WG["sex"])}', f' {"găini" if WG["sex"] == "F" else "cocoși"}, ziua {dd(WG["date"])}',
            f'ziua {dsapt(WG["date"])} · {FARM}',
            [val('greutate medie', f'{n(WG["avg_weight_g"])}{N}g', '', True), val('CV', D.pct(WG['cv_pct']), '', True), val('uniformitate', D.pct(WG['uniformity_pct']), '', True)], '',
            (dd(WG['date']), WG_DEL[11:]), WG['created_by'], 'veterinar', 'șters')
    return [a, b, c, d]

COLS = ('<li class="ip-chg__cols" aria-hidden="true"><span class="ip-cap">Când</span><span class="ip-cap">Înregistrare</span>'
        '<span class="ip-cap">Valori · vechi → nou</span><span class="ip-cap">Motiv</span><span class="ip-cap">Cine</span><span class="ip-cap">Tip</span></li>')

def clist(rows, label, cols=True):
    return f'<ul class="ip-list ip-list--change" aria-label="{label}">{COLS if cols else ""}{"".join(rows)}</ul>'

def change_blocks(s):
    rows = changes()
    out = []
    out.append(cap('Tipuri: corectat (Corectează, toată salvarea) · șters cu motiv (manager) · editat în aceeași zi · șters în aceeași zi (mai multe câmpuri)') +
               clist(rows, 'Istoric modificări'))
    a = changes()[0]
    out.append(cap('Stări: hover · apăsat · focus (rândul întreg deschide înregistrarea)') +
               clist([a.replace('ip-ev ip-ev--change', 'ip-ev ip-ev--change is-hover', 1),
                      a.replace('ip-ev ip-ev--change', 'ip-ev ip-ev--change is-pressed', 1),
                      a.replace('ip-ev ip-ev--change', 'ip-ev ip-ev--change ip-focus', 1)], 'Stări'))
    out.append(cap('Nicio modificare în perioadă · se încarcă') +
               '<div class="ip-stack" style="gap:var(--space-4)">'
               f'<ul class="ip-list ip-list--change" aria-label="Istoric modificări"><li class="ip-list__empty"><span class="ip-lamp ip-lamp--off">fără modificări</span>'
               f'<span>Nicio modificare la {FARM} între {num(dd(T - dt.timedelta(days=6)))} și {num(dd(T))}.</span><span>Schimbă perioada sau filtrele.</span></li></ul>'
               '<ul class="ip-list ip-list--change" aria-label="Istoric modificări" aria-busy="true">'
               + ''.join('<li class="ip-ev ip-ev--skel" aria-hidden="true"><span class="ip-skel ip-skel--tile"></span><span class="ip-ev__main"><span class="ip-skel ip-skel--w1"></span><span class="ip-skel ip-skel--w2"></span></span><span class="ip-skel ip-skel--w3"></span></li>' for _ in range(2))
               + '<li class="ip-sr">Se încarcă</li></ul></div>')
    return out

def build_change():
    d = change_blocks('-d'); m = change_blocks('-m')
    desk = ''.join(f'<div>{x}</div>' for x in d)
    phone = ''.join(f'<div>{x}</div>' for x in m)
    html = (f'<!-- @dsCard group="Data" height={H_CHG} subtitle="SYS09" -->\n'
        '<div class="ip-root ip-stack" style="gap:var(--space-6)">\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">S17 · Istoric modificări · desktop (listă de 1136 px, un rând pe modificare, ≥ 720 px)</span><div class="ip-stack" style="width:1136px;gap:var(--space-5)">{desk}</div></div>\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">390 · telefon (stivuit)</span><div class="ip-phone" style="width:390px"><div class="ip-stack" style="padding:var(--space-4);gap:var(--space-5)">{phone}</div></div></div>\n'
        f'<p class="ip-note" style="margin:0">{NOTE_CHG}</p>\n'
        '</div>\n')
    (ROOT / 'ChangeRow').mkdir(exist_ok=True)
    (ROOT / 'ChangeRow' / 'preview.html').write_text(SXS.sexify(html), encoding='utf-8')

SX = {'F': '♀', 'M': '♂'}
NOTE_ACT = (f'{FARM}, azi {num(dd(T))} (realdata, D152). Introdus azi = salvările veterinarului: Zilnic {num(hm(EM))} ({num(len(secs(T)))} secțiuni, ♀ {num(DF)} · ♂ {num(DM)}), '
            f'Ouă {num(hm(EE))} ({num(n(D.eggs(T)))}, ouat {num(D.pct(D.lay_day(T)))}), Hrană zilnică {num(hm(EF))} ({num(len(FED))} din {num(len(ROWS))} rânduri, {num(n(KG["F"] + KG["M"], 1))} kg). '
            f'Jurnal: Zilnicul din {num(dd(CD))} (♀ {num(CDF)}) corectat de {COR["user"]} pe {num(D.ddt(COR["at"]))} (corecția ilustrativă comună, D158a); '
            f'Sacrificare {CU["section"]} {SX[CU["sex"]]} {num(CU["qty"])} · {CU["reason"]} din {num(dd(CU["date"]))} (ilustrativă, din morțile reale ale zilei); '
            f'livrarea din {num(dd(DV["date"]))}, {num(n(DV["qty_kg"]))} kg fără preț (reală). Autorii și orele sunt ilustrative. '
            f'Exemple de stare (D39): tastele dezactivate după {num("23:59")} și ștergerea Sacrificării ({DEL_BY}, {num(D.ddt(DEL_AT))}).')
NOTE_CHG = (f'Corecția ilustrativă comună (D158a, illustrative-events-hirbovat.json › corrections): {COR["user"]}, {num(D.ddt(COR["at"]))}, corectează Zilnicul salvat de '
            f'{CE["user"]} pe {num(D.ddt(CE["at"]))}; înregistrarea e toată salvarea ({num(len(CS))} secțiuni, D172), se schimbă doar {COR["row"]["section"]} {SX[COR["row"]["sex"]]}: '
            f'{num(COR["old_values"]["qty"])} → {num(COR["new_values"]["qty"])} (morțile reale ale zilei). Exemple de stare (D39) pe înregistrări reale: ștergerea Sacrificării '
            f'{CU["section"]} {SX[CU["sex"]]} {num(CU["qty"])} · {CU["reason"]} din {num(dd(CU["date"]))} (ora și motivul sunt exemplu), editarea Zilnicului de azi în aceeași zi '
            f'({ED_SEC}: valorile vechi și ora {num(ED_AT[11:])} sunt exemplu, cele noi reale, ca S15 după 23:59) și ștergerea Cântăririi {WG["section"]} {SX[WG["sex"]]} din '
            f'{num(dd(WG["date"]))} (cântărire ilustrativă, D152; ștergerea e exemplu).')
H_ACT, H_CHG = 4034, 2715   # @dsCard heights = the preview page's height at 1440 (sys09_render.py prints it; re-measure after a change)

build_actions(); build_change()
print('ok')
