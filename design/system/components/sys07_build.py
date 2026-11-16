"""Build SexCountRow/preview.html (SYS07): python3 sys07_build.py"""
import pathlib
OUT = pathlib.Path('/home/claude/miraj-impex-v3/design/system/components/SexCountRow/preview.html')
NB = ' '
def fmt(n): return f'{n:,}'.replace(',', NB)
def num(n): return f'<span class="ip-num">{fmt(n) if isinstance(n,int) else n}</span>'
def sex(s, n=None):
    c = 'hen' if s == 'f' else 'rooster'; g = '♀' if s == 'f' else '♂'
    return f'<span class="ip-sex ip-sex--{c}">{g}' + (f' <b>{fmt(n) if isinstance(n,int) else n}</b>' if n is not None else '') + '</span>'
def pair(f, m, cls=''):
    return f'<span class="ip-pair{cls}">{sex("f",f)}<span class="sep">·</span>{sex("m",m)}</span>'
ICON = lambda d, sz='': f'<svg class="ip-icon{sz}" viewBox="0 0 24 24" aria-hidden="true">{d}</svg>'
X = '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>'
PLUS = '<path d="M5 12h14"/><path d="M12 5v14"/>'
DOWN = '<path d="m6 9 6 6 6-6"/>'; UP = '<path d="m18 15-6-6-6 6"/>'
CHECK = '<path d="M20 6 9 17l-5-5"/>'
ERR = '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>'
TELITA = ['H1', 'H2', 'H3', 'H6', 'H5-1', 'H5-2']

def field(uid, s, name, val, touch, ro=False, err=False, desc=''):
    c = 'hen' if s == 'f' else 'rooster'; word = 'găini' if s == 'f' else 'cocoși'
    iid = f'{s}-{uid}'
    cls = f'ip-input ip-count ip-count--{c}' + ('' if touch else ' ip-count--md') + (' is-error' if err else '')
    v = '' if val is None else f' value="{fmt(val)}"'
    pat = '[0-9 &#8239;]*' if (val is not None and val >= 1000) else '[0-9]*'
    attrs = f' id="{iid}" inputmode="numeric" pattern="{pat}" placeholder="–" autocomplete="off"{v}'
    if ro: attrs += ' readonly'
    if err: attrs += ' aria-invalid="true"'
    if desc: attrs += f' aria-describedby="{desc}"'
    return (f'<div class="ip-field is-{c}"><label class="ip-label ip-sexrow__lbl" for="{iid}"><span class="ip-sr">{name}, </span>'
            f'<span class="ip-sex ip-sex--{c}" aria-hidden="true">{"♀" if s=="f" else "♂"}</span><span class="ip-sr">{word}</span></label>'
            f'<span class="{cls}"><input{attrs}></span></div>')

def select(uid, val, touch, expanded=False, err=False, used=(), active=None, label='Secțiune', opts=None):
    """opts: (group caption, [(code, note_html), ...]) for a menu other than the Telița placement one."""
    val_code = (val or '').split(' › ')[-1]
    lid, cid, lst = f'sl-{uid}', f'sc-{uid}', f'sm-{uid}'
    cls = 'ip-input ip-sexrow__btn' + (' ip-input--touch' if touch else '') + (' is-error' if err else '')
    a = f' aria-activedescendant="so-{uid}-{active}"' if (expanded and active) else ''
    e = f' aria-invalid="true" aria-describedby="msg-{uid}"' if err else ''
    name = f'<span class="name">{val}</span>' if val else '<span class="name is-empty">Alege</span>'
    out = (f'<div class="ip-field ip-sexrow__sel"><span class="ip-label ip-sexrow__lbl" id="{lid}">{label}</span>'
           f'<div class="{cls}" id="{cid}" role="combobox" tabindex="0" aria-haspopup="listbox" aria-expanded="{"true" if expanded else "false"}" aria-controls="{lst}" aria-labelledby="{lid}"{a}{e}>{name}{ICON(UP if expanded else DOWN, " ip-icon--16")}</div>')
    if expanded and opts:
        grp, lst_ = opts
        items = ''.join(f'<li class="ip-menu__item{" is-active" if c == active else ""}" role="option" id="so-{uid}-{c}" aria-selected="{"true" if c == val_code else "false"}">{c} {note}</li>' for c, note in lst_)
        out += f'<ul class="ip-menu ip-sexrow__menu" role="listbox" id="{lst}" aria-labelledby="{lid}" tabindex="-1"><li class="ip-menu__group" role="presentation">{grp}</li>{items}</ul>'
    elif expanded:
        items = ''.join(
            (f'<li class="ip-menu__item" role="option" id="so-{uid}-{c}" aria-selected="false" aria-disabled="true">{c} <span class="ip-note">în listă</span></li>' if c in used else
             f'<li class="ip-menu__item{" is-active" if c == active else ""}" role="option" id="so-{uid}-{c}" aria-selected="false">{c} <span class="ip-note">goală</span></li>')
            for c in TELITA)
        out += f'<ul class="ip-menu ip-sexrow__menu" role="listbox" id="{lst}" aria-labelledby="{lid}" tabindex="-1"><li class="ip-menu__group" role="presentation">Telița · creștere</li>{items}</ul>'
    else:
        out += f'<ul class="ip-menu ip-sexrow__menu" role="listbox" id="{lst}" aria-labelledby="{lid}" hidden></ul>'
    return out + '</div>'

def row(uid, touch, code=None, sel=None, alive=None, f=None, m=None, toate=None, rm=None, err=None, msg=None):
    """err: 'f' | 'm' | 'sel'; rm: aria-label text."""
    name = code or (sel or {}).get('rowname') or sel.get('val')
    h = '<div class="ip-sexrow__head">'
    h += f'<span class="ip-sexrow__code" id="c-{uid}">{code}</span>' if code else select(uid, sel.get('val'), touch, sel.get('expanded', False), err == 'sel', sel.get('used', ()), sel.get('active'), sel.get('label', 'Secțiune'), sel.get('opts'))
    desc = ''
    if alive:
        h += f'<span class="ip-sexrow__alive" id="a-{uid}"><span class="ip-sexrow__k">în viață</span> {pair(*alive)}</span>'
        desc = f'a-{uid}'
    if rm:
        h += f'<button type="button" class="ip-iconbtn{" ip-iconbtn--touch" if touch else ""} ip-sexrow__rm" aria-label="{rm}">{ICON(X)}</button>'
    h += '</div>'
    ro = bool(toate)
    # read-only via Toate: the checked "Toate din H1.1" label describes why the inputs can't be typed in (native readonly, no aria-readonly)
    d = lambda s: ' '.join(x for x in (desc, f'all-{uid}' if ro else '', f'msg-{uid}' if err == s else '') if x)
    h += field(uid, 'f', name, f, touch, ro, err == 'f', d('f'))
    h += field(uid, 'm', name, m, touch, ro, err == 'm', d('m'))
    if msg:
        h += f'<p class="ip-err ip-sexrow__msg" id="msg-{uid}">{ICON(ERR, " ip-icon--14")}<span>{msg}</span></p>'
    if toate is not None:
        h += (f'<label class="ip-check ip-sexrow__all{" is-checked" if toate else ""}" id="all-{uid}"><input type="checkbox" class="ip-sr"{" checked" if toate else ""}>'
              f'<span class="ip-check__box">{ICON(CHECK, " ip-icon--14") if toate else ""}</span>Toate<span class="ip-sr"> din {name}</span></label>')
    return f'<li class="ip-sexrow">{h}</li>'

def panel(uid, title, ctx, rows, total, sub, wide=False, add=False, cols=None, pick=False):
    cls = 'ip-panel ip-sexrows' + (' ip-sexrows--wide' if wide else '') + (' ip-sexrows--add' if add else '') + (' ip-sexrows--pick' if pick else '')
    o = f'<section class="{cls}" aria-labelledby="t-{uid}"><div class="ip-sexrows__head"><h4 class="ip-cap" id="t-{uid}">{title}</h4><span class="ip-note">{ctx}</span></div>'
    if wide and cols:
        o += '<div class="ip-sexrows__cols" aria-hidden="true">' + ''.join(f'<span class="ip-cap">{c}</span>' for c in cols) + '</div>'
    o += f'<ol class="ip-sexrows__list">{"".join(rows)}</ol>'
    if add:
        o += f'<button type="button" class="ip-btn ip-btn--ghost{"" if wide else " ip-btn--touch"} ip-sexrows__add">{ICON(PLUS, " ip-icon--16")}Adaugă secțiune</button>'
    if total is None: return o + '</section>'
    o += f'<div class="ip-sexrows__total"><div class="ip-sexrows__tk"><span class="ip-cap">{total[0]}</span><span class="ip-sexrows__sub">{sub}</span></div>{pair(total[1], total[2], " ip-pair--lg")}</div></section>'
    return o

HIR = [('H1.1', 4280, 429), ('H1.2', 4274, 437), ('H3.1', 4270, 437), ('H3.2', 4204, 426), ('H4.1', 4117, 480), ('H4.2', 4107, 476)]
SALE_CTX = f'Hîrbovăț · {num("16.11.2026")}'
SALE_COLS = ['Secțiune', 'În viață', f'{sex("f")} găini', f'{sex("m")} cocoși', '']
PLACE_COLS = ['Secțiune', f'{sex("f")} găini', f'{sex("m")} cocoși', '', '']
ALIVE_SUB = f'în viață {pair(25252, 2685)}'

def sale(uid, touch, wide, refused=False):
    rows = []
    for i, (c, F, M) in enumerate(HIR):
        u = f'{uid}{i}'
        if c == 'H1.1': rows.append(row(u, touch, code=c, alive=(F, M), f=F, m=M, toate=True))
        elif c == 'H1.2': rows.append(row(u, touch, code=c, alive=(F, M), f=2000, toate=False))
        elif c == 'H3.1' and refused:
            rows.append(row(u, touch, code=c, alive=(F, M), m=500, toate=False, err='m',
                msg=f'H3.1: {num(500)} depășește cei {num(437)} cocoși în viață pe {num("16.11.2026")}. Introdu cel mult {num(437)}.'))
        else: rows.append(row(u, touch, code=c, alive=(F, M), toate=False))
    n = 3 if refused else 2
    return panel(uid, 'Păsări vândute', SALE_CTX, rows, ('Total vânzare', 6280, 929 if refused else 429),
                 f'<span>{num(n)} din {num(6)} secțiuni</span> · {ALIVE_SUB}', wide, False, SALE_COLS)

def place(uid, touch, wide, mode=None):
    rows = [row(f'{uid}0', touch, sel={'val': 'H1'}, f=14080, rm='Elimină H1'),
            row(f'{uid}1', touch, sel={'val': 'H2'}, f=14000, rm='Elimină H2'),
            row(f'{uid}2', touch, sel={'val': 'H5-1'}, m=3510, rm='Elimină H5-1')]
    sub = f'<span>{num(3)} secțiuni</span>'
    if mode == 'menu':
        rows.append(row(f'{uid}3', touch, sel={'val': None, 'rowname': 'rândul 4', 'expanded': True, 'used': ('H1', 'H2', 'H5-1'), 'active': 'H3'}, rm='Elimină rândul 4'))
        sub = f'<span>{num(4)} rânduri</span> · rândul 4 fără secțiune'
    if mode == 'dup':
        rows.append(row(f'{uid}3', touch, sel={'val': 'H2'}, rm='Elimină H2, rândul 4', err='sel',
                        msg='H2 e deja în listă. Alege altă secțiune sau elimină rândul.'))
        sub = f'<span>{num(4)} rânduri</span> · H2 de două ori'
    return panel(uid, 'Populare pe secțiuni', f'Telița · {num("07.10.2025")}', rows, ('Total populare', 28080, 3510), sub, wide, True, PLACE_COLS)

TRANSFER_COLS = ['Din', 'În viață', f'{sex("f")} găini', f'{sex("m")} cocoși', '']
def transfer(uid, touch, wide, mode='typed'):
    """S9 Transfer Din row: SectionSelect (farm › section, alive in the menu) + alive + ♀/♂ + Toate; no remove, no Adaugă, no totals."""
    F, M = HIR[0][1], HIR[0][2]
    sel = {'val': 'Hîrbovăț › H1.1', 'rowname': 'H1.1', 'label': 'Din'}
    if mode == 'menu':
        sel.update(expanded=True, active='H1.1', opts=('Hîrbovăț · ouat', [(c, pair(a, b)) for c, a, b in HIR]))
        r = row(f'{uid}0', touch, sel=sel, alive=(F, M), toate=False)
    elif mode == 'all':
        r = row(f'{uid}0', touch, sel=sel, alive=(F, M), f=F, m=M, toate=True)
    else:
        r = row(f'{uid}0', touch, sel=sel, alive=(F, M), f=1000, m=100, toate=False)
    return panel(uid, 'Păsări transferate', num('16.11.2026'), [r], None, '', wide, False, TRANSFER_COLS, pick=True)

def phone(cap, body):
    return (f'<div class="ip-stack" style="width:390px;gap:var(--space-2)"><span class="ip-cap">{cap}</span>'
            f'<div class="ip-phone"><div style="padding:var(--space-4)">{body}</div></div></div>')
def desk(cap, body, extra=''):
    return f'<div class="ip-stack" style="max-width:800px;gap:var(--space-2){extra}"><span class="ip-cap">{cap}</span>{body}</div>'

html = '<!-- @dsCard group="Entry" height=5600 subtitle="SYS07" -->\n<div class="ip-root ip-stack" style="gap:var(--space-8)">\n'
html += '<div class="ip-row" style="align-items:flex-start;gap:var(--space-8)">'
html += phone('Vânzare · telefon · Toate pe H1.1, H1.2 parțial', sale('pa', True, False))
html += phone('Vânzare · telefon · refuzat: peste cei în viață', sale('pb', True, False, True))
html += phone('Serie nouă · telefon · populare 07.10.2025', place('pc', True, False))
html += phone('Serie nouă · telefon · secțiune de două ori', place('pd', True, False, 'dup'))
html += phone('Transfer · telefon · Din H1.1, parțial', transfer('pe', True, False))
html += '</div>\n<div class="ip-stack" style="gap:var(--space-8)">'
html += desk('Vânzare · desktop · Toate pe H1.1, H1.2 parțial', sale('da', False, True))
html += desk('Vânzare · desktop · refuzat: peste cei în viață', sale('db', False, True, True))
html += desk('Serie nouă · desktop · rând nou, alegere secțiune (deschis)', place('dc', False, True, 'menu'), ';padding-bottom:calc(var(--space-12) * 4)')
html += desk('Serie nouă · desktop · secțiune de două ori', place('dd', False, True, 'dup'))
html += desk('Transfer · desktop · Din H1.1, Toate', transfer('de', False, True, 'all'))
html += desk('Transfer · desktop · alegere Din (deschis)', transfer('df', False, True, 'menu'), ';padding-bottom:calc(var(--space-12) * 4)')
html += '</div>\n'
html += (f'<p class="ip-note" style="margin:0;max-width:1234px">Cifre: Hîrbovăț în viață pe {num("16.11.2026")} din eșantion (D27, counts-latest); populare Telița {num("07.10.2025")} reală (H1 {sex("f",14080)}, H2 {sex("f",14000)}, H5-1 {sex("m",3510)}). '
         f'Cifre tastate, exemple de stare (D39): H1.2 {sex("f",2000)} (parțial), H3.1 {sex("m",500)} (refuzat), H2 adăugat a doua oară.</p>\n')
html += (f'<p class="ip-note" style="margin:0;max-width:1234px">Transfer din Hîrbovăț › H1.1: exemplu de stare (D39); extrasele au doar transferurile Telița → Hîrbovăț din februarie 2026. '
         f'În viață H1.1 {pair(4280, 429)} din counts-latest (eșantion, D27); tastat {pair(1000, 100)} (parțial) și Toate (tot ce e în viață).</p>\n')
html += ('<p class="ip-note" style="margin:0;max-width:1234px">În produs valoarea câmpului e doar cifre (<code>pattern="[0-9]*"</code>); gruparea cu spațiu îngust se aplică la ieșirea din câmp. '
         'Previzualizarea statică arată textul grupat, deci câmpurile cu valori grupate folosesc aici, doar pentru demonstrație, <code>pattern="[0-9 &#8239;]*"</code>.</p>\n</div>\n')
OUT.write_text(html, encoding='utf-8'); print(len(html))
