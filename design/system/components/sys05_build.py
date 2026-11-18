"""SYS05: build EntryCardPlus and HallCard preview.html fragments.
Figures: design/data/extracts/sample-hirbovat-2026.json (16.11.2026 and 09.11–15.11.2026, sample D27)."""
import pathlib
ROOT = pathlib.Path('/home/claude/miraj-impex-v3/design/system/components')
N = ' '  # thousands
M = '−'  # true minus
def g(s):
    if isinstance(s, int): return f'{s:,}'.replace(',', N)
    return str(s).replace(' ', N)

SKULL = '<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 3a7 7 0 0 0-7 7v4l2 2v3h10v-3l2-2v-4a7 7 0 0 0-7-7z"/><circle cx="9.5" cy="11" r="1.2"/><circle cx="14.5" cy="11" r="1.2"/><path d="M11 19v-2M13 19v-2"/></svg>'
BAN = '<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M5.7 5.7l12.6 12.6"/></svg>'
CHEV = '<svg class="ip-icon ip-icon--16" viewBox="0 0 24 24" aria-hidden="true"><path d="m6 9 6 6 6-6"/></svg>'
WARN = '<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg>'
ERR = '<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/></svg>'
EGG = '<svg class="ip-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 22c4 0 7-3 7-7.5C19 9 16 2 12 2S5 9 5 14.5C5 19 8 22 12 22z"/></svg>'
EGG16 = EGG.replace('class="ip-icon"', 'class="ip-icon ip-icon--16"')
CHECK = '<svg class="ip-icon ip-icon--14" viewBox="0 0 24 24" aria-hidden="true"><path d="M20 6 9 17l-5-5"/></svg>'

SEXW = {'hen': ('♀', 'găini'), 'rooster': ('♂', 'cocoși')}
def sx(sex, txt=None):
    sign = SEXW[sex][0]
    return f'<span class="ip-sex ip-sex--{sex}">{sign}{(" <b>" + g(txt) + "</b>") if txt is not None else ""}</span>'
def sxl(sex):  # sign in a label: hidden from AT, the word is read instead
    return f'<span class="ip-sex ip-sex--{sex}" aria-hidden="true">{SEXW[sex][0]}</span><span class="ip-sr">{SEXW[sex][1]}</span>'
def pair(f, m): return f'<span class="ip-pair">{sx("hen", f)}<span class="sep">·</span>{sx("rooster", m)}</span>'

# ---- data (sample 16.11.2026; hints from 15.11 and the 7-day mean 09.11–15.11) ----
SEC = {
 'H1.1': dict(alive=(4281, 429), ieri=(3, 1), avg=('1,0', '0,3'), mort=(1, 0), cull=(0, 0)),
 'H1.2': dict(alive=(4275, 437), ieri=(1, 0), avg=('0,6', '0,1'), mort=(1, 0), cull=(0, 0)),
 'H3.1': dict(alive=(4270, 437), ieri=(0, 0), avg=('0,9', '0,0'), mort=(0, 0), cull=(0, 0)),
 'H3.2': dict(alive=(4205, 427), ieri=(3, 0), avg=('1,4', '0,7'), mort=(1, 1), cull=(0, 0)),
}
EGGS = {
 'H1': dict(secs=('H1.1', 'H1.2'), alive=(8556, 866), jos=345, c1=23, c2=131, c3=70, inc=6264, w='68,3', total=6488, lay='75,8', d=('+', '1,7'), ok=True),
 'H3': dict(secs=('H3.1', 'H3.2'), alive=(8475, 864), jos=115, c1=16, c2=85, c3=47, inc=5792, w='70,1', total=5940, lay='70,1', d=(M, '4,0'), ok=False),
}
GUIDE = '74,1'

def hint(id_, sex, ieri, avg):
    return (f'<span class="ip-entry__hint" id="{id_}"><span>ieri {sx(sex)} <span class="ip-num">{ieri}</span></span>'
            f'<span aria-hidden="true">·</span><span>media 7 zile <span class="ip-num">{avg}</span></span></span>')

def pat(v):
    """In the product the input value is digits only (pattern [0-9]*) and the U+202F grouping is applied on blur.
    The static preview shows the grouped text, so grouped values get a demo pattern that accepts the separator."""
    return ' pattern="[0-9 &#8239;]*"' if N in v else ' pattern="[0-9]*"'

def count(id_, kind, sex, val=None, state=None, ro=False, hint_html='', desc=(), md=False):
    icon = f'<span class="ip-rec ip-rec--{"mortality" if kind == "Mortalitate" else "cull"}">{SKULL if kind == "Mortalitate" else BAN}</span>'
    cls = f'ip-input ip-count ip-count--{sex}' + (' ip-count--md' if md else '') + (f' is-{state}' if state else '')
    v = '' if val is None else f' value="{g(val)}"'
    p = pat(v)
    extra = ' readonly' if ro else ''
    if state == 'error': extra += ' aria-invalid="true"'
    d = ' '.join(desc)
    dd = f' aria-describedby="{d}"' if d else ''
    return (f'<div class="ip-field is-{sex}"><label class="ip-label" for="{id_}">{icon}{kind} {sxl(sex)}</label>'
            f'<span class="{cls}"><input id="{id_}" inputmode="numeric"{p} placeholder="–" autocomplete="off"{v}{extra}{dd}></span>{hint_html}</div>')

def motiv(id_, sex, val=None, error_id=None, md=False):
    opts = ''.join(f'<option{" selected" if val == o else ""}>{o}</option>' for o in ('boală', 'șchiopătare', 'altul'))
    ph = f'<option value=""{"" if val else " selected"}>Alege</option>'
    cls = 'ip-input' + ('' if md else ' ip-input--touch') + (' is-error' if error_id else '')
    inv = f' aria-invalid="true" aria-describedby="{error_id}"' if error_id else ''
    return (f'<div class="ip-field is-{sex}"><label class="ip-label" for="{id_}">Motiv {sxl(sex)}</label>'
            f'<span class="{cls}"><select id="{id_}" required{inv}>{ph}{opts}</select>{CHEV}</span></div>')

def entry(code, sfx, mode='empty', md=False, h='h4', over=None):
    """mode: empty | filled | warn | error | cullmissing | nolosses | saved"""
    s = SEC[code]; uid = f'{code.replace(".", "")}-{sfx}'
    f_alive, m_alive = s['alive']
    mort = list(s['mort']); cull = list(s['cull'])
    tag = ''
    if mode == 'empty': mort = [None, None]; cull = [None, None]
    if over:
        for k, v in over.items():
            if k.startswith('mort'): mort[0 if k.endswith('f') else 1] = v
            if k.startswith('cull'): cull[0 if k.endswith('f') else 1] = v
    ro = mode in ('saved', 'nolosses')
    if mode == 'nolosses': mort = [0, 0]; cull = [0, 0]; tag = '<span class="ip-tag ip-tag--neutral">fără pierderi</span>'
    if mode == 'saved': tag = '<span class="ip-tag ip-tag--neutral">salvat</span>'
    if mode == 'warn': tag = '<span class="ip-tag ip-tag--warn">abatere</span>'
    if mode in ('error', 'cullmissing'): tag = '<span class="ip-tag ip-tag--fault">refuzat</span>'
    cells = []; msgs_after_mort = ''; msgs_after_cull = ''
    st = [None, None]; desc = [[f'hf-{uid}'], [f'hm-{uid}']]
    if mode == 'warn':
        st[0] = 'warn'; desc[0].append(f'w-{uid}')
        msgs_after_mort = (f'<p class="ip-warnline ip-entry__msg" id="w-{uid}" role="status" style="margin:0">{WARN}<span>{code}: '
                           f'<span class="ip-num">{g(mort[0])}</span> găini moarte, de peste 3 ori media pe 7 zile (<span class="ip-num">{s["avg"][0]}</span>). Verifică numărătoarea; poți salva.</span></p>')
    if mode == 'error':
        st[1] = 'error'; desc[1].append(f'e-{uid}')
        msgs_after_mort = (f'<p class="ip-err ip-entry__msg" id="e-{uid}" style="margin:0">{ERR}<span>{code}: <span class="ip-num">{g(mort[1])}</span> depășește cei '
                           f'<span class="ip-num">{g(m_alive)}</span> cocoși în viață pe <span class="ip-num">16.11.2026</span>. Introdu cel mult <span class="ip-num">{g(m_alive)}</span>.</span></p>')
    for i, sex in enumerate(('hen', 'rooster')):
        cells.append(count(f'm{i}-{uid}', 'Mortalitate', sex, mort[i], st[i], ro, hint(desc[i][0], sex, s['ieri'][i], s['avg'][i]), desc[i], md))
    for i, sex in enumerate(('hen', 'rooster')):
        cells.append(count(f'c{i}-{uid}', 'Sacrificare', sex, cull[i], None, ro, '', (), md))
    for i, sex in enumerate(('hen', 'rooster')):
        if cull[i]:
            eid = f'r{i}-err-{uid}' if mode == 'cullmissing' else None
            cells.append(motiv(f'r{i}-{uid}', sex, None if mode == 'cullmissing' else 'șchiopătare', eid, md))
            if eid:
                noun = 'găini sacrificate' if sex == 'hen' else 'cocoși sacrificați'
                msgs_after_cull += f'<p class="ip-err ip-entry__msg" id="{eid}" style="margin:0">{ERR}<span>{code}: alege motivul pentru <span class="ip-num">{cull[i]}</span> {noun}.</span></p>'
    grid = cells[0] + cells[1] + msgs_after_mort + ''.join(cells[2:]) + msgs_after_cull
    btn = 'ip-btn' if md else 'ip-btn ip-btn--touch'
    if mode == 'saved':
        foot = (f'<div class="ip-entry__foot"><div class="ip-entry__saved"><span>Salvat la <span class="ip-num">08:34</span> · Dr. Ceban</span>'
                f'<span>poți modifica până la <span class="ip-num">23:59</span></span></div>'
                f'<button type="button" class="{btn} ip-btn--secondary">Editează<span class="ip-sr"> {code}</span></button></div>')
        more = ''
    else:
        chk = ' is-checked' if mode == 'nolosses' else ''
        box = CHECK if mode == 'nolosses' else ''
        checked = ' checked' if mode == 'nolosses' else ''
        more = f'<button type="button" class="{btn} ip-btn--ghost ip-entry__more" aria-expanded="false" aria-controls="more-{uid}">{CHEV}Mai mult · Eroare sexare</button>'
        more += f'<div class="ip-entry__panel" id="more-{uid}" hidden></div>'
        foot = f'<div class="ip-entry__foot"><label class="ip-check{chk}"><input type="checkbox" class="ip-sr"{checked}><span class="ip-check__box">{box}</span>Fără pierderi azi</label></div>'
    return (f'<section class="ip-panel ip-entry" aria-labelledby="t-{uid}"><div class="ip-entry__head"><{h} class="ip-entry__code" id="t-{uid}">{code}</{h}>{tag}'
            f'<span class="ip-entry__alive">în viață {pair(f_alive, m_alive)}</span></div>'
            f'<div class="ip-entry__grid">{grid}</div>{more}{foot}</section>')

def eggfield(id_, label, val, md, ro, weight=False, empty=False):
    v = '' if empty else f' value="{g(val)}"'
    r = ' readonly' if ro else ''
    if weight:
        cls = 'ip-input ip-eggs__weight' + ('' if md else ' ip-input--touch')
        return (f'<div class="ip-field"><label class="ip-label" for="{id_}">{label}</label><span class="{cls}">'
                f'<input id="{id_}" inputmode="decimal" placeholder="–" autocomplete="off"{v}{r}><span class="suffix">g</span></span></div>')
    cls = 'ip-input ip-count' + (' ip-count--md' if md else '')
    return (f'<div class="ip-field"><label class="ip-label" for="{id_}">{label}</label><span class="{cls}">'
            f'<input id="{id_}" inputmode="numeric"{pat(v)} placeholder="–" autocomplete="off"{v}{r}></span></div>')

def eggs(hall, sfx, mode='filled', md=False):
    e = EGGS[hall]; uid = f'{hall}-{sfx}'; empty = mode == 'empty'; ro = mode == 'saved'
    tag = '<span class="ip-tag ip-tag--neutral">salvat</span>' if ro else ''
    fields = ''.join(eggfield(f'{k}-{uid}', lab, e[k], md, ro, empty=empty) for k, lab in
                     (('jos', 'Jos'), ('c1', 'Cat. 1'), ('c2', 'Cat. 2'), ('c3', 'Cat. 3'), ('inc', 'Incubabile')))
    fields += eggfield(f'w-{uid}', 'Masa medie<span class="ip-sr"> grame</span>', e['w'], md, ro, weight=True, empty=empty)
    f_alive = g(e['alive'][0])
    if empty:
        total = ('<div><dt>Total ouă</dt><dd class="ip-eggs__val">–</dd><dd class="ip-eggs__sub">= Cat. 1 + 2 + 3 + Incubabile · Jos separat · '
                 'completează cele 4 câmpuri</dd></div>')
        lay = (f'<div><dt>Ouat azi</dt><dd class="ip-eggs__val">–</dd><dd class="ip-eggs__sub">ghid <span class="ip-num">{GUIDE}%</span> · '
               f'după ce introduci ouăle · din <span class="ip-sex ip-sex--hen">♀ {f_alive}</span></dd></div>')
    else:
        total = (f'<div><dt>Total ouă</dt><dd class="ip-eggs__val">{g(e["total"])}</dd><dd class="ip-eggs__sub">= Cat. 1 + 2 + 3 + Incubabile · '
                 f'Jos separat (<span class="ip-num">{e["jos"]}</span>)</dd></div>')
        lamp = '<span class="ip-lamp ip-lamp--ok">în normă</span>' if e['ok'] else '<span class="ip-lamp ip-lamp--warn">abatere</span>'
        thr = '' if e['ok'] else ' (prag ±3)'
        lay = (f'<div><dt>Ouat azi</dt><dd class="ip-eggs__val">{e["lay"]}%</dd><dd class="ip-eggs__sub">ghid <span class="ip-num">{GUIDE}%</span> · '
               f'<span class="ip-num">{e["d"][0]}{e["d"][1]}</span> puncte{thr} · din <span class="ip-sex ip-sex--hen">♀ {f_alive}</span></dd>'
               f'<dd class="ip-eggs__lamp">{lamp}</dd></div>')
    btn = 'ip-btn' if md else 'ip-btn ip-btn--touch'
    foot = ''
    if ro:
        foot = (f'<div class="ip-entry__foot"><div class="ip-entry__saved"><span>Salvat la <span class="ip-num">08:35</span> · Dr. Ceban</span>'
                f'<span>poți modifica până la <span class="ip-num">23:59</span></span></div>'
                f'<button type="button" class="{btn} ip-btn--secondary">Editează<span class="ip-sr"> ouăle {hall}</span></button></div>')
    return (f'<section class="ip-panel ip-eggs" aria-labelledby="t-eggs-{uid}"><div class="ip-eggs__head">'
            f'<span class="ip-tile ip-rec ip-rec--eggs" aria-hidden="true">{EGG}</span><h4 class="ip-entry__code" id="t-eggs-{uid}">Ouă · {hall}</h4>'
            f'<span class="ip-note">pe hală, pentru {" + ".join(e["secs"])}</span>{tag}</div>'
            f'<div class="ip-eggs__grid">{fields}</div>'
            f'<dl class="ip-eggs__calc" aria-label="Calculat din câmpuri, doar citire">{total}{lay}</dl>{foot}</section>')

def hall(code, sfx, mode='filled', md=False, smode=None):
    e = EGGS[code]; uid = f'{code}-{sfx}'
    secs = ''.join(entry(s, sfx, smode or ('saved' if mode == 'saved' else mode), md=md) for s in e['secs'])
    return (f'<section class="ip-hall{" ip-hall--wide" if md else ""}" aria-labelledby="t-hall-{uid}"><header class="ip-hall__head">'
            f'<span class="ip-cap">Hală</span><h3 class="ip-hall__code" id="t-hall-{uid}">{code}</h3>'
            f'<span class="ip-note">secțiuni <span class="ip-num">{" · ".join(e["secs"])}</span></span>'
            f'<span class="ip-pair"><span class="ip-sr">în viață </span>{sx("hen", e["alive"][0])}<span class="sep">·</span>{sx("rooster", e["alive"][1])}</span></header>'
            f'<div class="ip-hall__sections">{secs}</div>{eggs(code, sfx, mode, md)}</section>')

def savebar(lamp, word, what, disabled):
    dis = ' disabled' if disabled else ''
    return (f'<div class="ip-savebar"><span class="ip-savebar__sum"><span class="ip-entry-sum"><span class="ip-lamp ip-lamp--{lamp}">{word}</span>'
            f'<span>{what}</span></span></span><button type="button" class="ip-btn ip-btn--primary ip-btn--touch"{dis}>Salvează</button></div>')

def phone(cap, body, bar=''):
    return (f'<div class="ip-stack" style="width:390px;gap:var(--space-2)"><span class="ip-cap">{cap}</span>'
            f'<div class="ip-phone"><div style="padding:var(--space-4)">{body}</div>{bar}</div></div>')

DEMO = ('<p class="ip-note" style="margin:0;max-width:1234px">În produs valoarea câmpului e doar cifre (<code>pattern="[0-9]*"</code>); '
        'gruparea cu spațiu îngust se aplică la ieșirea din câmp. Previzualizarea statică arată textul grupat, deci câmpurile cu valori grupate '
        'folosesc aici, doar pentru demonstrație, <code>pattern="[0-9 ]*"</code> (cifre, spațiu și spațiu îngust).</p>')
ROW = '<div class="ip-row" style="align-items:flex-start;gap:var(--space-8)">'

# ---------------- EntryCardPlus ----------------
ec = ['<!-- @dsCard group="Entry" height=2300 subtitle="SYS05" -->', '<div class="ip-root ip-stack" style="gap:var(--space-8)">']
ec.append(ROW
    + phone('Gol · 16.11.2026 · toate „–”', entry('H1.1', 'a', 'empty'))
    + phone('Completat · 16.11.2026 (eșantion)', entry('H1.1', 'b', 'filled'))
    + phone('Salvat · citire, Editează până la 23:59', entry('H1.1', 'c', 'saved'))
    + '</div>')
ec.append(ROW
    + phone('Abatere · peste 3× media 7 zile · se poate salva', entry('H1.1', 'd', 'warn', over={'mortf': 4}),
            savebar('warn', 'abatere', f'H1.1 {sx("hen")} <span class="ip-num">4</span> · poți salva', False))
    + phone('Refuzat · peste păsările în viață', entry('H1.2', 'e', 'error', over={'mortm': 4370}),
            savebar('fault', 'refuzat', f'H1.2 {sx("rooster")} <span class="ip-num">{g("4 370")}</span> · corectează ca să salvezi', True))
    + phone('Motiv lipsă · după Salvează', entry('H1.1', 'f', 'cullmissing', over={'mortf': 1, 'mortm': 0, 'cullf': 2}),
            savebar('fault', 'refuzat', f'H1.1 {sx("hen")} · alege motivul', True))
    + '</div>')
ec.append(ROW + phone('Fără pierderi azi · bifat', entry('H1.2', 'g', 'nolosses')) + '</div>')
ec.append('<p class="ip-note" style="margin:0;max-width:1234px">Cifre din eșantionul Hîrbovăț (D27): în viață la începutul zilei <span class="ip-num">16.11.2026</span>, '
          'ieri = <span class="ip-num">15.11.2026</span>, media 7 zile = <span class="ip-num">09.11–15.11.2026</span>. Exemple de stare, nu din extras (D39): '
          f'H1.1 {sx("hen")} mortalitate <span class="ip-num">4</span> (abatere), H1.2 {sx("rooster")} <span class="ip-num">{g("4 370")}</span> (refuzat), '
          f'H1.1 {sx("hen")} sacrificare <span class="ip-num">2</span> fără motiv.</p>')
ec.append(DEMO)
ec.append('</div>')
(ROOT / 'EntryCardPlus').mkdir(exist_ok=True)
(ROOT / 'EntryCardPlus' / 'preview.html').write_text('\n'.join(ec) + '\n', encoding='utf-8')

# ---------------- HallCard ----------------
hc = ['<!-- @dsCard group="Entry" height=3900 subtitle="SYS05" -->', '<div class="ip-root ip-stack" style="gap:var(--space-8)">']
hc.append(ROW
    + phone('Hală · gol · 16.11.2026', hall('H1', 'a', 'empty'))
    + phone('Hală · completat · 16.11.2026 (eșantion)', hall('H1', 'b', 'filled'))
    + phone('Hală · salvat', hall('H1', 'c', 'saved'))
    + '</div>')
hc.append('<div class="ip-stack" style="max-width:1136px;gap:var(--space-2)"><span class="ip-cap">Desktop · 2 coloane · 16.11.2026 (eșantion) · H3 ouat în abatere</span>'
          '<div class="ip-halls">' + hall('H1', 'd', 'filled', md=True) + hall('H3', 'd', 'filled', md=True) + '</div></div>')
hc.append('<p class="ip-note" style="margin:0;max-width:1136px">Cifre din eșantionul Hîrbovăț pentru <span class="ip-num">16.11.2026</span> (D27). '
          'Total ouă = Cat. 1 + Cat. 2 + Cat. 3 + Incubabile, Jos separat (D18). Ouat = total ouă ÷ găinile în viață la începutul zilei; '
          'ghid <span class="ip-num">SAPT 40</span> = <span class="ip-num">74,1%</span>, abatere peste ±3 puncte (D19).</p>')
hc.append(DEMO.replace('1234px', '1136px'))
hc.append('</div>')
(ROOT / 'HallCard').mkdir(exist_ok=True)
(ROOT / 'HallCard' / 'preview.html').write_text('\n'.join(hc) + '\n', encoding='utf-8')
print('ok')
