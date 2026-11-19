"""Build the SYS10 preview (SectionTree, S19 Ferme și secțiuni): python3 sys10_build.py
RF11: names, codes, section states and dates come from design/canvas/scripts/realdata.py (farms-sections, counts, cycles;
today 18.11.2026, D152); guide names are illustrative (D22). Every edit shown on a real code is a D39 state example.
@dsCard height = the content (.ip-root) height at 1440 + 32 (FIN13B); re-measure after a change (H_CARD)."""
import sys, pathlib
REPO = pathlib.Path(__file__).resolve().parents[3]  # repo root from this script's place (design/system/scripts), so a copy writes into itself
ROOT = REPO / 'design/system/components'
sys.path.insert(0, str(REPO / 'design/canvas/scripts'))
import realdata as RD   # s19_build execs this part after s18_build imported it: the same module
DATA = RD.FS

def svg(paths, cls='ip-icon ip-icon--16'):
    return f'<svg class="{cls}" viewBox="0 0 24 24" aria-hidden="true">{paths}</svg>'

IC = {
    'more': '<circle cx="12" cy="5" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="12" cy="19" r="1"/>',  # Lucide ellipsis-vertical
    'plus': '<path d="M5 12h14"/><path d="M12 5v14"/>',  # Lucide plus
    'alert': '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>',  # Lucide circle-alert
}

def cap(text):
    return f'<span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">{text}</span>'

# ------------------------------------------------------------------ data
GUIDES = {g['id']: g['name'] for g in DATA['guides']}
TYPE = {'growth': 'creștere', 'laying': 'ouat'}
FARMS = {f['id']: f for f in DATA['farms']}

def farm_tree(fid, order=None):
    secs = [s for s in DATA['sections'] if s['farm_id'] == fid and s['active']]
    tops = [s for s in secs if s['parent_id'] is None]
    tree = [(t['code'], [c['code'] for c in secs if c['parent_id'] == t['id']]) for t in tops]
    if order:  # state example: siblings after a Mută în sus / jos
        tree.sort(key=lambda t: order.index(t[0]))
    return tree

# Section state today (RD.TODAY, 18.11.2026): Hîrbovăț's laying cycle is open (from RD.CYCLE_START), its halls hold birds
# except the one emptied by sale (H7 since 14.11, D155/D158d); Telița's growth cycle closed when its birds moved to
# Hîrbovăț (D25), so every Telița section is empty. Codes are unique per farm only (Telița H6 ≠ Hîrbovăț H6): look up by farm.
assert RD.COUNTS_TELITA['totals']['all'] == 0
def birds(fid, code):
    """the section (or every compartment of a hall) holds birds at the end of today."""
    return fid == 'hirbovat' and RD.has_birds(code)
def kids_of(fid, code):
    return next((k for c, k in farm_tree(fid) if c == code), [])

def plural(n, one, many):
    return f'<span class="ip-num">{n}</span> {one if n == 1 else many}'

def farm_note(fid, retired=0):
    tree = farm_tree(fid)
    halls = [t for t in tree if t[1]]
    t = TYPE[FARMS[fid]['type']]
    if halls:
        comp = sum(len(c) for _, c in halls)
        return f'{t} · {plural(len(halls), "hală", "hale")} · {plural(comp, "compartiment", "compartimente")}'
    n = len(tree) - retired
    s = f'{t} · {plural(n, "secțiune", "secțiuni")}'
    return s + (f' · {plural(retired, "retrasă", "retrase")}' if retired else '')

# ------------------------------------------------------------------ parts
LAMP = {True: '<span class="ip-lamp ip-lamp--ok">cu păsări</span>', False: '<span class="ip-lamp ip-lamp--off">goală</span>'}
def state_cell(fid, code, kind, b):
    """Stare: the lamp; an empty hall of a farm whose cycle is still open adds its muted date (D158d: "goală din 14.11")."""
    es = RD.empty_since(code) if (fid == 'hirbovat' and kind == 'hall' and not b) else None
    return LAMP[b] + (f' <span class="ip-note ip-nowrap">din <span class="ip-num">{RD.ddm(es)}</span></span>' if es else '')

def menu_html(mid, code, items, note, open_=False):
    """items: (label, disabled, sep_before, state). The note is about Retrage (its effect, or why it is disabled), so only
    the Retrage item points to it: read once, when Retrage gets focus (FIN13; the menu no longer repeats it). The key's
    aria-controls names the role=menu element (id = mid), as SYS09."""
    nid = f'{mid}-n'
    out = []
    for label, dis, sep, st in items:
        if sep:
            out.append('<div class="ip-stree__sep" role="separator"></div>')
        d = ' aria-disabled="true"' if dis else ''
        if label == 'Retrage':
            d += f' aria-describedby="{nid}"'
        hp = ' aria-haspopup="dialog"' if label == 'Retrage' else ''  # opens the Retrage Dialog, as SYS09's Șterge
        s = f' {st}' if st and open_ else ''
        out.append(f'<button type="button" class="ip-menu__item{s}" role="menuitem" tabindex="-1"{hp}{d}>{label}</button>')
    return (f'<div class="ip-menu ip-stree__menu ip-t-body"{"" if open_ else " hidden"}>'
            f'<div class="ip-stree__items" id="{mid}" role="menu" aria-label="Acțiuni, {code}">{"".join(out)}</div>'
            f'<p class="ip-note" id="{nid}">{note}</p></div>')

def act_cell(mid, code, items, note, open_=False, kstate=''):
    exp = 'true' if open_ else 'false'
    k = (f'<button type="button" class="ip-iconbtn ip-stree__key{" " + kstate if kstate else ""}" aria-label="Acțiuni, {code}" '
         f'aria-haspopup="menu" aria-expanded="{exp}" aria-controls="{mid}">{svg(IC["more"])}</button>')
    return f'<td class="ip-stree__act">{k}{menu_html(mid, code, items, note, open_)}</td>'

def items_for(kind, first, last, birds, state_on=None):
    st = (lambda i: state_on.get(i, '')) if state_on else (lambda i: '')
    it = [('Redenumește', False, False, st(0)),
          ('Mută în sus', first, False, st(1)),
          ('Mută în jos', last, False, st(2))]
    if kind != 'comp':
        it.append(('Adaugă compartiment', False, False, st(3)))
    it.append(('Retrage', 'why' if birds else False, True, st(4)))
    return it

def note_for(kind, code, birds, kids=()):
    if kind == 'hall' and birds:
        return f'{code} are compartimente cu păsări. O poți retrage după ce le retragi pe ele.'
    if kind == 'hall' and kids:   # components.md SYS10 Rules: a hall with active compartments is blocked as one with birds
        return f'{code} are compartimente active ({", ".join(kids)}). O poți retrage după ce le retragi pe ele.'
    if not birds:
        return f'Retrage scoate {code} din formulare și de pe pagina fermei; înregistrările rămân.'
    return f'{code} are păsări. O poți retrage când e goală.'

def farm_row(fid, name=None, tag=None, note=None):
    f = FARMS.get(fid, {})
    name = name or f['name']
    tag = tag or f'<span class="ip-tag ip-tag--neutral">{GUIDES[f["guide_id"]]}</span>'
    note = note or farm_note(fid)
    return (f'<tr class="is-group ip-stree__farm"><td colspan="3"><span class="ip-stree__farmname"><b>{name}</b>{tag}</span>'
            f'<span class="ip-note">{note}</span></td></tr>')

def sec_row(p, fid, code, kind, first, last, birds, open_=False, kstate='', parent=None, is_last=False, note='',
            retired=False, istate=None):
    lvl = 'ip-stree--l3' if kind == 'comp' else 'ip-stree--l2'
    cls = f'ip-stree__row {lvl}' + (' is-last' if is_last else '') + (' is-retired' if retired else '')
    sr = (f'<span class="ip-sr">, compartiment în {parent}</span>' if kind == 'comp'
          else '<span class="ip-sr">, hală</span>' if kind == 'hall' else '')
    nt = f'<span class="ip-note">{note}</span>' if note else ''
    cell = f'<td><span class="ip-stree__cell"><span class="ip-stree__code">{code}{sr}</span>{nt}</span></td>'
    if retired:
        return f'<tr class="{cls}">{cell}<td><span class="ip-tag ip-tag--neutral">retrasă</span></td><td class="ip-stree__act"></td></tr>'
    mid = f'{p}-m-{code}'.replace('.', '-').replace('/', '-')
    kids = kids_of(fid, code) if kind == 'hall' else ()
    act = act_cell(mid, code, items_for(kind, first, last, birds or bool(kids), istate), note_for(kind, code, birds, kids), open_, kstate)
    return f'<tr class="{cls}">{cell}<td>{state_cell(fid, code, kind, birds)}</td>{act}</tr>'

def add_row(fid, label='Adaugă secțiune'):
    name = FARMS[fid]['name'] if fid in FARMS else 'Fermă exemplu'
    return (f'<tr class="ip-stree__add ip-stree--l2"><td colspan="3"><button type="button" class="ip-btn ip-btn--ghost">'
            f'{svg(IC["plus"])}{label}<span class="ip-sr"> la {name}</span></button></td></tr>')

def edit_row(p, level, label, value, err=None, focus=False, is_last=True, placeholder=''):
    """Inline edit row (add or rename), one row height since FW03 (D150): the row holds the code field at the code's place
    and Salvează + Anulează on the row line, on the edit fill (tr.ip-table__edit); under it one detail row (additions.css §69, indented as the row's level)
    holds the field's label ("Compartiment nou în H5", "Cod nou pentru H5-2": keeps the old code in view) and its error."""
    lvl = 'ip-stree--l3' if level == 3 else 'ip-stree--l2'
    iid = f'{p}-in'
    eid = f'{p}-err'
    st = ' is-error' if err else (' is-focus' if focus else '')
    inv = f' aria-invalid="true" aria-errormessage="{eid}" aria-describedby="{eid}"' if err else ''
    ph = f' placeholder="{placeholder}"' if placeholder else ''
    e = f'<p class="ip-err ip-fieldmsg" id="{eid}">{svg(IC["alert"], "ip-icon ip-icon--14")}<span>{err}</span></p>' if err else ''  # §38 field message with the alert icon (S19P, FIN13)
    return (f'<tr class="ip-table__edit ip-stree__edit {lvl}{" is-last" if is_last else ""}"><td colspan="3"><span class="ip-stree__form">'
            f'<span class="ip-input ip-stree__in{st}"><input id="{iid}" type="text" value="{value}" autocomplete="off"{ph}{inv}></span>'
            f'<button type="button" class="ip-btn ip-btn--secondary">Salvează</button>'
            f'<button type="button" class="ip-btn ip-btn--ghost">Anulează</button></span></td></tr>'
            f'<tr class="ip-table__detail ip-stree__detail"><td colspan="3"><div class="ip-table__detailbody">'
            f'<label class="ip-label" for="{iid}">{label}</label>{e}</div></td></tr>')

def section_rows(p, fid, open_code=None, kstates=None, after=None, replace=None, retired=(), istate=None, order=None):
    """after: {code: extra_row_html} inserted after that row; replace: {code: html}."""
    kstates = kstates or {}
    after = after or {}
    replace = replace or {}
    rows = []
    tree = farm_tree(fid, order)
    active = [t for t in tree if t[0] not in retired]
    for i, (code, kids) in enumerate(active):
        kind = 'hall' if kids else 'sec'
        note = plural(len(kids), 'compartiment', 'compartimente') if kids else ''
        if code in replace:
            rows.append(replace[code])
        else:
            rows.append(sec_row(p, fid, code, kind, i == 0, i == len(active) - 1, birds(fid, code), open_code == code,
                                kstates.get(code, ''), note=note, istate=istate if open_code == code else None))
        for j, c in enumerate(kids):
            last_kid = j == len(kids) - 1 and code not in after
            rows.append(sec_row(p, fid, c, 'comp', j == 0, j == len(kids) - 1, birds(fid, c), open_code == c,
                                kstates.get(c, ''), parent=code, is_last=last_kid,
                                istate=istate if open_code == c else None))
        if code in after:
            rows.append(after[code])
    for code in retired:
        rows.append(sec_row(p, fid, code, 'sec', False, False, False, retired=True))
    return rows

HEAD_NOTE = 'Secțiunile cu păsări nu se pot retrage.'

def tree(p, label, body_rows, banner='', busy=False, head=True, live=''):
    thead = ('<thead><tr><th scope="col">Secțiune</th><th scope="col">Stare</th>'
             '<th scope="col" class="ip-stree__acth"><span class="ip-sr">Acțiuni</span></th></tr></thead>')
    b = ' aria-busy="true"' if busy else ''
    h = (f'<div class="ip-list__head"><h3 class="ip-cap">Secțiuni</h3><p class="ip-note">{HEAD_NOTE}</p></div>' if head else '')
    return (f'<div class="ip-stree">{h}{banner}<div class="ip-stree__box"><table class="ip-table ip-stree__t" aria-label="{label}"{b}>'
            f'{thead}<tbody>{"".join(body_rows)}</tbody></table></div>'
            f'<p class="ip-sr" role="status" aria-live="polite">{live}</p></div>')  # announces Mută în sus / jos

def farm(p, fid, **kw):
    retired = kw.get('retired', ())
    fr = farm_row(fid, note=farm_note(fid, len(retired)) if retired else None)
    live = kw.pop('live', '')
    return tree(p, f'Secțiuni, {FARMS[fid]["name"]}', [fr] + section_rows(p, fid, **kw) + [add_row(fid)], live=live)

def banner(kind, paras, focus=False):
    """A failure Banner as the S19 boards draw it (FX05): the Banner with action (additions.css section 49, FIN03), role=alert,
    16px icon on the first line, one p per sentence group, then Reîncearcă (D143: every failure Banner carries it; after a
    failed save it takes the focus, D141). Copy D143: the object leads, then the one retry line."""
    body = ''.join(f'<p class="ip-m0">{t}</p>' for t in paras)
    f = ' ip-focus' if focus else ''
    return (f'<div class="ip-banner ip-banner--{kind} ip-banner--act" role="alert">{svg(IC["alert"])}<div class="ip-banner__body">{body}</div>'
            f'<button type="button" class="ip-btn ip-btn--secondary ip-banner__key{f}">Reîncearcă</button></div>')

# ------------------------------------------------------------------ state examples on the real tree (RF11; S19/S19S read these)
# Hîrbovăț: Adaugă compartiment in the first hall = its next free code (H5 → H5.3); the duplicate = that hall's last code
# (H5.2); Retrage blocked on the first compartment with birds (H5.1); the empty-code example in the last hall (H7). Telița
# (D39, no section was ever retired): Retrage / Redenumește on its last section (H5-2), Mută în sus on its third (H3).
TREE_H = dict(farm_tree('hirbovat'))
ADD_HALL = RD.HHALLS[0]
ADD_CODE = f'{ADD_HALL}.{len(TREE_H[ADD_HALL]) + 1}'
DUP_CODE = TREE_H[ADD_HALL][-1]
BLOCKED = next(c for c in RD.HSECS if birds('hirbovat', c))
LAST_HALL = RD.HHALLS[-1]
assert ADD_CODE not in RD.HSECS and DUP_CODE in RD.HSECS and BLOCKED == RD.HSECS[0]
TEL = [c for c, _ in farm_tree('telita')]; assert TEL == RD.SECTIONS['telita']
TEL_OPEN, TEL_RET, TEL_UP = TEL[1], TEL[-1], TEL[2]
TEL_UP_ORDER = TEL[:1] + [TEL_UP] + [c for c in TEL[1:] if c != TEL_UP]   # TEL_UP after Mută în sus: 2nd
TEL_RENAME = TEL_RET.replace('-', '/')   # the register's own label for that section (farms-sections workbook_labels)
assert TEL_RENAME in next(x for x in DATA['sections'] if x['farm_id'] == 'telita' and x['code'] == TEL_RET)['workbook_labels']
ORD = {2: 'a 2-a'}
DUP_MSG = f'{DUP_CODE} există deja în {ADD_HALL}. Scrie alt cod, de exemplu {ADD_CODE}.'   # one wording, preview and S19 board
def moved(n_active): return f'{TEL_UP} mutată în sus: {ORD[TEL_UP_ORDER.index(TEL_UP) + 1]} din {n_active} secțiuni la Telița.'

# ------------------------------------------------------------------ blocks
B = []
B.append(('Hîrbovăț · implicit (ciclu deschis; ' + LAST_HALL + ' goală din ' + RD.ddm(RD.empty_since(LAST_HALL)) + ')', farm('a', 'hirbovat')))
B.append((f'Hîrbovăț · meniul {BLOCKED} deschis: Retrage blocat, cu motivul', farm('b', 'hirbovat', open_code=BLOCKED, istate={0: 'is-active'})))
B.append((f'Telița · meniul {TEL_OPEN} deschis: toate acțiunile (ciclu închis, secțiuni goale)', farm('c', 'telita', open_code=TEL_OPEN)))
B.append((f'Cheia Acțiuni: hover ({TEL[0]}) · apăsat ({TEL[1]}) · focus ({TEL[2]})',
          farm('d', 'telita', kstates={TEL[0]: 'is-hover', TEL[1]: 'is-pressed', TEL[2]: 'ip-focus'})))
MOVED = moved(len(TEL))
B.append((f'Mută în sus {TEL_UP} (exemplu de stare): focus rămâne pe cheia {TEL_UP} · anunț live „{MOVED}”',
          farm('d2', 'telita', order=TEL_UP_ORDER, kstates={TEL_UP: 'ip-focus'}, live=MOVED)))
B.append((f'Adaugă compartiment în {ADD_HALL} · cod scris (exemplu de stare)',
          farm('e', 'hirbovat', after={ADD_HALL: edit_row('e', 3, f'Compartiment nou în {ADD_HALL}', ADD_CODE, focus=True)})))
B.append(('Cod deja folosit',
          farm('f', 'hirbovat', after={ADD_HALL: edit_row('f', 3, f'Compartiment nou în {ADD_HALL}', DUP_CODE,
               err=DUP_MSG)})))
B.append(('Cod gol (Salvează fără cod)',
          farm('g', 'hirbovat', after={LAST_HALL: edit_row('g', 3, f'Compartiment nou în {LAST_HALL}', '',
               err='Scrie codul compartimentului.')})))
B.append((f'Redenumește {TEL_RET} → {TEL_RENAME} (eticheta din registru, exemplu de stare)',
          farm('h', 'telita', replace={TEL_RET: edit_row('h', 2, f'Cod nou pentru {TEL_RET}', TEL_RENAME, focus=True)})))
B.append(('Salvarea a eșuat: rândul rămâne deschis, cu codul scris (focus pe Reîncearcă)',
          tree('i', 'Secțiuni, Hîrbovăț',
               [farm_row('hirbovat')] + section_rows('i', 'hirbovat', after={ADD_HALL: edit_row('i', 3, f'Compartiment nou în {ADD_HALL}', ADD_CODE)})
               + [add_row('hirbovat')],
               banner=banner('fault', [f'Compartimentul {ADD_CODE} nu s-a salvat: conexiunea s-a întrerupt.',
                                       'Verifică conexiunea la internet, apoi apasă Reîncearcă. Codul scris rămâne în rând.'], focus=True))))
DIALOG = ('<div class="ip-scrim"><div class="ip-dialog" role="dialog" aria-modal="true" aria-labelledby="j-t" aria-describedby="j-d">'
          f'<h2 id="j-t">Retrage secțiunea {TEL_RET} de la Telița</h2>'
          f'<p id="j-d" style="margin:0">{TEL_RET} nu mai apare în formulare și pe pagina fermei. Înregistrările ei din Seria 2026 rămân în rapoarte.</p>'
          '<div class="ip-row" style="justify-content:flex-end;gap:var(--space-2)">'
          '<button type="button" class="ip-btn ip-btn--secondary ip-focus">Anulează</button>'
          '<button type="button" class="ip-btn ip-btn--destructive">Retrage secțiunea</button></div></div></div>')
B.append((f'Retrage {TEL_RET} · Dialog (exemplu de stare, focus pe Anulează)', DIALOG))
B.append((f'După retragere: {TEL_RET} retrasă (exemplu de stare)', farm('k', 'telita', retired=(TEL_RET,))))
EX = ('<tr class="is-group ip-stree__farm"><td colspan="3"><span class="ip-stree__farmname"><b>Fermă exemplu</b>'
      '<span class="ip-tag ip-tag--neutral">fără ghid</span></span><span class="ip-note">creștere · fără secțiuni</span></td></tr>'
      '<tr class="ip-stree__empty ip-stree--l2"><td colspan="3"><span class="ip-stree__cell"><span class="ip-lamp ip-lamp--off">fără secțiuni</span>'
      '<span>Fermă exemplu nu are încă secțiuni.</span></span></td></tr>')
B.append(('Fermă nouă: fără secțiuni, fără ghid (Fermă exemplu)', tree('l', 'Secțiuni, Fermă exemplu', [EX, add_row('ex')])))
NOFARM = ('<div class="ip-stree__box"><div class="ip-empty"><span class="ip-lamp ip-lamp--off">fără ferme</span>'
          '<span>Nu există nicio fermă.</span><button type="button" class="ip-btn ip-btn--primary">Adaugă prima fermă</button></div></div>')
B.append(('Nicio fermă (lista din S19, B1)', NOFARM))
SK = ('<tr class="is-group ip-stree__farm ip-stree__skel" aria-hidden="true"><td colspan="3"><span class="ip-skel ip-skel--w2"></span></td></tr>'
      + ''.join(f'<tr class="ip-stree__skel ip-stree--l2" aria-hidden="true"><td><span class="ip-skel ip-skel--w3"></span></td>'
                f'<td><span class="ip-skel ip-skel--w3"></span></td><td class="ip-stree__act"></td></tr>' for _ in range(4))
      + '<tr><td colspan="3" class="ip-sr">Se încarcă</td></tr>')
B.append(('Se încarcă', tree('n', 'Secțiuni', [SK], busy=True)))
B.append(('Nu s-a încărcat',
          '<div class="ip-stree"><div class="ip-list__head"><h3 class="ip-cap">Secțiuni</h3></div>'
          + banner('fault', ['Secțiunile fermei Hîrbovăț nu s-au încărcat.', 'Verifică conexiunea la internet, apoi apasă Reîncearcă.'])
          + '</div>'))
WIDE = tree('p', 'Secțiuni, toate fermele',
            [farm_row('telita')] + section_rows('p', 'telita') + [add_row('telita'),
             farm_row('hirbovat')] + section_rows('p2', 'hirbovat') + [add_row('hirbovat')])

OPEN = (f'Hîrbovăț · meniul {BLOCKED}', f'Telița · meniul {TEL_OPEN}')  # the open menu hangs below its table: reserve the room
grid = ''.join(f'<div style="width:var(--panel-w){";padding-bottom:calc(var(--space-12) * 2)" if t.startswith(OPEN) else ""}">{cap(t)}{h}</div>' for t, h in B)
def codes(fid):
    return ', '.join(f'{c} ({", ".join(k)})' if k else c for c, k in farm_tree(fid))
H_CARD = 3678   # @dsCard height = content (.ip-root) at 1440 + 32 (FIN13B/FX05 rule; RF11: 3646 + 32, was 3797)
html = (f'<!-- @dsCard group="Data" height={H_CARD} subtitle="SYS10" -->\n'
        '<div class="ip-root ip-stack" style="gap:var(--space-8)">\n'
        f'<div>{cap("S19 · panoul fermei · desktop (lățime panou 440 px)")}'
        f'<div style="display:grid;grid-template-columns:repeat(3,var(--panel-w));gap:var(--space-8) var(--space-6);align-items:start">{grid}</div></div>\n'
        f'<div style="width:calc(var(--panel-w) * 2 + var(--space-6))">{cap("Toate fermele · aceeași componentă, lățime de pagină")}{WIDE}</div>\n'
        '<p class="ip-note" style="margin:0;max-width:calc(var(--panel-w) * 2)">Reale (farms-sections.json): '
        f'Telița {codes("telita")}; Hîrbovăț {codes("hirbovat")}; tipurile fermelor. Ilustrative (D22): numele ghidurilor. '
        f'Stări de azi ({RD.dd(RD.TODAY)}): ciclul de ouat Hîrbovăț e deschis din {RD.dd(RD.CYCLE_START)}, azi {RD.SAPT(RD.WK)}; '
        f'{LAST_HALL} goală din {RD.dd(RD.empty_since(LAST_HALL))} (vânzare); Telița goală din {RD.dd(RD.CTEL["end_date"])} (D25). '
        f'Exemple de stare (D39): codul scris {ADD_CODE}, codul dublat {DUP_CODE}, redenumirea {TEL_RET} → {TEL_RENAME} (eticheta din registru), '
        f'{TEL_RET} retrasă, {TEL_UP} mutată, salvarea eșuată; ferma nouă e „Fermă exemplu” (D38).</p>\n'
        '</div>\n')
(ROOT / 'SectionTree').mkdir(exist_ok=True)
(ROOT / 'SectionTree' / 'preview.html').write_text(html, encoding='utf-8')
print('ok')
