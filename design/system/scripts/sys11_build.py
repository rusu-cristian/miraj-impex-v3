"""Build the SYS11 preview (GuideImportPreview, S20 Ghiduri import, spec B2): python3 sys11_build.py
Guide figures come from design/data/extracts/guide-growth.json and guide-laying.json (real workbook guide values);
guide and file names are illustrative (D22). The three refused rows of the error state are state examples (D39)."""
import json, pathlib
REPO = pathlib.Path(__file__).resolve().parents[3]  # repo root from this script's place (design/system/scripts), so a copy writes into itself
ROOT = REPO / 'design/system/components'
EXT = REPO / 'design/data/extracts'
GROW = json.loads((EXT / 'guide-growth.json').read_text(encoding='utf-8'))['rows']
LAY = json.loads((EXT / 'guide-laying.json').read_text(encoding='utf-8'))['rows']
NNBSP = ' '

def svg(paths, cls='ip-icon ip-icon--16'):
    return f'<svg class="{cls}" viewBox="0 0 24 24" aria-hidden="true">{paths}</svg>'

IC = {  # Banner and FileDrop previews (Lucide)
    'fault': '<circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>',
    'info': '<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>',
    'ok': '<path d="M20 6 9 17l-5-5"/>',
    'upload': '<path d="M12 15V3"/><path d="m7 8 5-5 5 5"/><path d="M5 21h14"/>',
}

def cap(text):
    return f'<span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">{text}</span>'

def fmt(v, unit=''):
    """Thin no-break thousands, true minus, decimal comma; a percentage keeps one decimal (89,0, as the guide detail)."""
    if v is None:
        return None
    if unit == '%':
        return f'{float(v):.1f}'.replace('.', ',')
    if isinstance(v, float) and not v.is_integer():
        return f'{v:.1f}'.replace('.', ',')
    n = int(v)
    s = f'{abs(n):,}'.replace(',', NNBSP)
    return ('−' if n < 0 else '') + s

def num(n):
    return f'<span class="ip-num">{n}</span>'

SEX = {'F': ('♀', 'hen', 'găini'), 'M': ('♂', 'rooster', 'cocoși')}
COLS = {'growth': [('feed_g', 'Hrană g'), ('body_weight_g', 'Greutate g')],
        'laying': [('feed_g', 'Hrană g'), ('lay_pct', 'Ouat %')]}
NAME = {'feed_g': 'hrana', 'body_weight_g': 'greutatea', 'lay_pct': 'ouat %'}
UNIT = {'feed_g': 'g', 'body_weight_g': 'g', 'lay_pct': '%'}

# ------------------------------------------------------------------ rows (as read from the file; line 1 = header)
def growth_rows():
    rows = [dict(week=r['week'], age=None, sex=r['sex'], v={'feed_g': r['feed_g'], 'body_weight_g': r['body_weight_g']})
            for r in GROW]
    return rows

def laying_rows():
    return [dict(week=r['cycle_week'], age=r['age_week'], sex=r['sex'], v={'feed_g': r['feed_g'], 'lay_pct': r['lay_pct']})
            for r in LAY]

def number(rows):
    for i, r in enumerate(rows):
        r['line'] = i + 2
    return rows

def error_rows():
    """State examples (D39) on the real growth guide: SAPT 3 ♂ weight cell holding the Excel error "#REF!" (real 550);
    SAPT 4 ♀ pasted twice (copy of its own real row); SAPT 5 ♂ with the sex cell left empty."""
    rows = growth_rows()
    out = []
    for r in rows:
        if r['week'] == 3 and r['sex'] == 'M':
            r = dict(r, bad={'body_weight_g': '#REF!'}, kind='nan')
        if r['week'] == 5 and r['sex'] == 'M':
            r = dict(r, sex=None, bad={'sex': ''}, kind='nosex')
        out.append(r)
        if r['week'] == 4 and r['sex'] == 'M':
            src = next(x for x in rows if x['week'] == 4 and x['sex'] == 'F')
            out.append(dict(src, bad={'week': '4', 'sex': 'F'}, kind='dup'))
    number(out)
    first = next(x for x in out if x['week'] == 4 and x['sex'] == 'F' and 'kind' not in x)
    for r in out:
        if r.get('kind') == 'dup':
            r['dup_of'] = first['line']
    return out

def sexg(code, sr=None):
    """A ♀/♂ in running text and link names (D128): the coloured glyph aria-hidden, the word for screen readers."""
    g, tone, word = SEX[code]
    return f'<span class="ip-sex ip-sex--{tone}" aria-hidden="true">{g}</span><span class="ip-sr">{sr or word}</span>'

def reason(r):
    k = r.get('kind')
    if k == 'nan':
        return (f'SAPT {r["week"]} {sexg("M")}', 'greutatea „#REF!” nu e număr', 'Greutatea „#REF!” nu e număr. Scrie greutatea în grame.', 'body_weight_g')
    if k == 'dup':
        return (f'SAPT {r["week"]} {sexg("F")}', f'e deja pe rândul {r["dup_of"]}',
                f'Rândul {r["dup_of"]} are deja sapt. {r["week"]} {sexg("F")}. Păstrează un singur rând.', 'week')
    if k == 'nosex':
        return (f'SAPT {r["week"]}', 'lipsește sexul', f'Lipsește sexul. Scrie G ({sexg("F")}) sau C ({sexg("M")}).', 'sex')
    return None

# ------------------------------------------------------------------ table
def head(kind):
    th = '<th scope="col" class="num">Rând</th>'
    if kind == 'laying':
        th += '<th scope="col" class="num">Vârstă</th>'
    th += '<th scope="col" class="num">SAPT</th><th scope="col">Sex</th>'
    th += ''.join(f'<th scope="col" class="num">{lbl}</th>' for _, lbl in COLS[kind])
    th += '<th scope="col" class="ip-gimp__chkh">Verificare</th>'
    return f'<thead><tr>{th}</tr></thead>'

def bad_td(p, r, key, raw, cls='num', focus=False):
    cid = f'{p}-r{r["line"]}-{key}'
    why = f'{p}-r{r["line"]}-why'
    f = ' ip-focus' if focus else ''
    shown = raw if raw else '<span class="ip-sr">gol</span>'
    return (f'<td class="{cls} ip-gimp__bad{f}" id="{cid}" tabindex="-1" aria-invalid="true" '
            f'aria-describedby="{why}">{shown}</td>')

def row_html(p, kind, r, focus_cell=None):
    bad = r.get('bad', {})
    sx = SEX.get(r['sex'])
    tone = f' is-{sx[1]}' if sx else ''
    cells = [f'<td class="num ip-gimp__ln">{r["line"]}</td>']
    if kind == 'laying':
        cells.append(f'<td class="num">{r["age"]}</td>')
    fc = lambda k: focus_cell == k
    cells.append(bad_td(p, r, 'week', bad['week'], focus=fc('week')) if 'week' in bad else f'<td class="num">{r["week"]}</td>')
    if 'sex' in bad:
        shown = {'F': '♀', 'M': '♂'}.get(bad['sex'], bad['sex'])
        cells.append(bad_td(p, r, 'sex', shown, 'ip-gimp__sex', focus=fc('sex')))
    else:
        cells.append(f'<td class="ip-gimp__sex{tone}">{sx[0]}<span class="ip-sr"> {sx[2]}</span></td>')
    gaps = []
    for key, lbl in COLS[kind]:
        if key in bad:
            cells.append(bad_td(p, r, key, bad[key], focus=fc(key)))
            continue
        v = fmt(r['v'][key], UNIT[key])
        if v is None:
            if key == 'lay_pct' and r['sex'] == 'M':
                cells.append('<td class="num ip-gimp__none">–<span class="ip-sr"> nu se aplică la cocoși</span></td>')
            else:
                cells.append('<td class="num ip-gimp__none">–<span class="ip-sr"> fără date în ghid</span></td>')
                gaps.append(NAME[key])
        else:
            cells.append(f'<td class="num{tone}">{v}</td>')
    rs = reason(r)
    if rs:
        chk = (f'<span class="ip-gimp__why"><span class="ip-lamp ip-lamp--fault">refuzat</span>'
               f'<span id="{p}-r{r["line"]}-why">{rs[2]}</span></span>')
    elif gaps:
        chk = f'<span class="ip-gimp__gap">fără date în ghid: {", ".join(gaps)}</span>'
    else:
        chk = ''
    cls = ' class="is-refused"' if rs else ''
    return f'<tr{cls}>{"".join(cells)}<td class="ip-gimp__chk">{chk}</td></tr>'

def table(p, kind, rows, label, focus=None, busy=False, body=None, sum_=True):
    b = ' aria-busy="true"' if busy else ''
    d = f' aria-describedby="{p}-sum"' if sum_ else ''
    tb = body if body is not None else ''.join(row_html(p, kind, r, focus.get(r['line']) if focus else None) for r in rows)
    return (f'<div class="ip-gimp__box" tabindex="0" role="region" aria-label="Rânduri citite, {label}">'
            f'<table class="ip-table ip-gimp__t"{d}{b}>{head(kind)}<tbody>{tb}</tbody></table></div>')

# loading: six section 42 skeleton rows (tr.ip-table__skel, aria-hidden) under the real head; each bar is as wide as the
# column's figures will be (ch of the Mono cell: Rând, SAPT, Vârstă 2; Sex 1; hrană 3; greutate, ouat % 4) and sits where
# they will sit (numbers at the right edge, sex at the left); Verificare stays empty (a valid row has no text there).
CH = {'line': 2, 'age': 2, 'week': 2, 'sex': 1, 'feed_g': 3, 'body_weight_g': 4, 'lay_pct': 4}
def skel_body(kind, n=6):
    keys = ['line'] + (['age'] if kind == 'laying' else []) + ['week', 'sex'] + [k for k, _ in COLS[kind]]
    bar = lambda k: f'<span class="ip-skel ip-skel--ch{CH[k]}{"" if k == "sex" else " ip-skel--end"}"></span>'  # §42 ch widths (FIN11)
    tds = ''.join(f'<td class="{"ip-gimp__sex" if k == "sex" else "num"}">{bar(k)}</td>' for k in keys)
    return ''.join(f'<tr class="ip-table__skel" aria-hidden="true">{tds}<td class="ip-gimp__chk"></td></tr>' for _ in range(n))

# ------------------------------------------------------------------ parts
def src_line(p, name, note, counts, key='Alege alt fișier'):
    """key='' leaves the key out (the empty file: the EmptyState holds the one "Alege alt fișier")."""
    c = f'<p class="ip-gimp__counts" id="{p}-sum">{counts}</p>' if counts else f'<p class="ip-gimp__counts" id="{p}-sum"></p>'
    k = f'<button type="button" class="ip-btn ip-btn--ghost">{key}</button>' if key else ''
    return (f'<div class="ip-gimp__src"><span class="ip-gimp__file"><b>{name}</b><span class="ip-note">{note}</span></span>'
            f'{c}{k}</div>')

def counts(read, valid, refused):
    s = f'{num(read)} rânduri citite · {num(valid)} valide'
    return s + (f' · {num(refused)} refuzate' if refused else '')

def banner(kind, body, bid=''):
    i = f' id="{bid}"' if bid else ''
    return f'<div class="ip-banner ip-banner--{kind}"{i}>{svg(IC[kind])}<div class="ip-gimp__bn">{body}</div></div>'

def err_banner(p, rows):
    bad = [r for r in rows if reason(r)]
    li = ''.join(f'<li><a class="ip-link" href="#{p}-r{r["line"]}-{reason(r)[3]}">Rândul {r["line"]} · {reason(r)[0]}</a>: '
                 f'{reason(r)[1]}.</li>' for r in bad)
    n = len(bad)
    return banner('fault', f'<p><b>{n} rânduri refuzate.</b> Corectează-le în fișier și încarcă-l din nou; '
                           f'până atunci ghidul nu se poate salva.</p><ul class="ip-gimp__errs">{li}</ul>', f'{p}-errs')

def info_banner(text):
    return banner('info', f'<p><b>Nu blochează salvarea.</b> {text}</p>')

def seg(p, total, refused, on='all'):
    a = ('is-selected', 'true') if on == 'all' else ('', 'false')
    b = ('is-selected', 'true') if on == 'bad' else ('', 'false')
    return (f'<div class="ip-gimp__bar"><h4 class="ip-cap">Rânduri</h4>'
            f'<div class="ip-seg" role="group" aria-label="Rânduri afișate">'
            f'<button type="button" class="ip-seg__item {a[0]}" aria-pressed="{a[1]}">Toate {num(total)}</button>'
            f'<button type="button" class="ip-seg__item {b[0]}" aria-pressed="{b[1]}">Refuzate {num(refused)}</button></div></div>')

def bar(p):
    return f'<div class="ip-gimp__bar"><h4 class="ip-cap">Rânduri</h4></div>'

def foot(p, ok, note):
    k = ('<button type="button" class="ip-btn ip-btn--primary">Salvează ghidul</button>' if ok else
         f'<button type="button" class="ip-btn ip-btn--primary is-disabled" aria-disabled="true" aria-describedby="{p}-foot">Salvează ghidul</button>')
    return (f'<div class="ip-gimp__foot">{k}<button type="button" class="ip-btn ip-btn--ghost">Anulează</button>'
            f'<p class="ip-note" id="{p}-foot">{note}</p></div>')

def live(text):
    return f'<p class="ip-sr" role="status" aria-live="polite">{text}</p>'

def wrap(p, inner, label):
    return f'<section class="ip-gimp" aria-label="{label}">{inner}</section>'

DROP_NOTE = 'sapt · sex · hrană g · greutate g · mortalitate % · ouat % · incubabile % · masă ou g'
# FileDrop as the paste target (additions section 43): focusable group, Ctrl+V pastes into it; the file button inside
DROP_LABEL = 'Lipește tabelul aici (Ctrl+V) sau alege un fișier .xlsx sau .csv'
DROP = (f'<div class="ip-drop ip-drop--paste" tabindex="0" role="group" aria-label="{DROP_LABEL}">{svg(IC["upload"], "ip-icon ip-icon--22")}'
        '<div>Lipește tabelul aici sau <button type="button" class="ip-link">alege un fișier</button></div>'
        f'<div class="ip-note">{DROP_NOTE}</div></div>')

def empty_file(name):
    """A file with the header only: EmptyState in the table box, its one next step "Alege alt fișier"."""
    return ('<div class="ip-gimp__box"><div class="ip-empty"><span class="ip-lamp ip-lamp--off">fără date</span>'
            f'<span>{name} are doar antetul, niciun rând de ghid.</span>'
            '<button type="button" class="ip-btn ip-btn--secondary">Alege alt fișier</button></div></div>')
EMPTY_NOTE = 'Scrie rândurile ghidului sub antet sau alege alt fișier.'
NOHEAD = ('<p><b>{f} nu are antetul cu coloanele SAPT și Sex.</b> Pune pe primul rând numele coloanelor de sub chenar '
          'și încarcă fișierul din nou.</p>')

# ------------------------------------------------------------------ blocks
B = []
GF = 'Ross 308 creștere.xlsx'

# 1 loading
B.append(('Se citește fișierul (skeleton, aria-busy)', wrap('a',
    src_line('a', GF, 'fișier .xlsx · tip creștere', '<span class="ip-gimp__wait">Se citește fișierul.</span>', 'Anulează citirea')
    + bar('a') + table('a', 'growth', [], GF, busy=True, body=skel_body('growth'))
    + foot('a', False, 'Se citește fișierul.') + live(f'Se citește {GF}.'), 'Import ghid')))

# 2 all valid, growth
gr = number(growth_rows())
gap_lines = [r['line'] for r in gr if r['v']['body_weight_g'] is None]
INFO_G = (f'Greutatea lipsește la sapt. 24–25 (rândurile {gap_lines[0]}–{gap_lines[-1]}), iar fișierul nu are coloanele '
          'Mortalitate %, Ouat %, Incubabile % și Masă ou g. În ghid apar ca „–” fără date în ghid.')
B.append(('Toate valide · creștere (real, 50 de rânduri; cutia derulează)', wrap('b',
    src_line('b', GF, 'fișier .xlsx · tip creștere', counts(len(gr), len(gr), 0))
    + banner('ok', f'<p><b>Toate cele {len(gr)} de rânduri sunt valide.</b> Verifică cifrele și salvează ghidul.</p>')
    + info_banner(INFO_G)
    + bar('b') + table('b', 'growth', gr, GF)
    + foot('b', True, f'Se salvează {len(gr)} de rânduri în ghidul Ross 308 creștere.')
    + live(f'Fișier citit: {len(gr)} de rânduri, toate valide.'), 'Import ghid')))

# 2b all valid, end of the box (the partial weeks)
B.append(('Toate valide · capătul tabelului (derulat): SAPT 24–25 fără greutate', wrap('b2',
    bar('b2') + table('b2', 'growth', [r for r in gr if r['week'] >= 20], GF, sum_=False), 'Import ghid')))

# 3 refused rows
er = error_rows()
nbad = sum(1 for r in er if reason(r))
gl = [r['line'] for r in er if r['v']['body_weight_g'] is None]
INFO_E = (f'Greutatea lipsește la sapt. 24–25 (rândurile {gl[0]}–{gl[-1]}), iar fișierul nu are coloanele '
          'Mortalitate %, Ouat %, Incubabile % și Masă ou g. În ghid apar ca „–” fără date în ghid.')
NOTE_E = f'{nbad} rânduri refuzate: corectează fișierul și încarcă-l din nou.'
B.append((f'Rânduri refuzate · Toate (exemple de stare D39: rândurile 7, 10, 12)', wrap('c',
    src_line('c', GF, 'fișier .xlsx · tip creștere', counts(len(er), len(er) - nbad, nbad))
    + err_banner('c', er) + info_banner(INFO_E)
    + seg('c', len(er), nbad, 'all') + table('c', 'growth', er, GF)
    + foot('c', False, NOTE_E)
    + live(f'Fișier citit: {len(er)} de rânduri, {nbad} refuzate.'), 'Import ghid')))

# 4 filter: refused only, focus on the cell reached from the summary link
B.append(('Filtru Refuzate · focus pe celula „#REF!” (după linkul din rezumat)', wrap('d',
    seg('d', len(er), nbad, 'bad') + table('d', 'growth', [r for r in er if reason(r)], GF, focus={7: 'body_weight_g'}, sum_=False)
    + foot('d', False, NOTE_E), 'Import ghid')))

# 5 laying, pasted: age + SAPT, lay % partial
lr = number(laying_rows())
LF = 'Ross 308 ouat'
INFO_L = ('Ouat % lipsește la vârsta 19–24 și 64–65 (sapt. 1–6 și 46–47), iar tabelul nu are coloanele Greutate g, '
          'Mortalitate %, Incubabile % și Masă ou g. În ghid apar ca „–” fără date în ghid. Ouat % se citește doar pe rândurile <span class="ip-sex ip-sex--hen" aria-hidden="true">♀</span><span class="ip-sr">de găini</span>.')
B.append(('Ouat · tabel lipit (real, 94 de rânduri): vârstă și SAPT (SAPT = vârstă − 18, D26)', wrap('e',
    src_line('e', 'Tabel lipit', 'din clipboard · tip ouat', counts(len(lr), len(lr), 0), 'Lipește alt tabel')
    + banner('ok', f'<p><b>Toate cele {len(lr)} de rânduri sunt valide.</b> Verifică cifrele și salvează ghidul.</p>')
    + info_banner(INFO_L)
    + bar('e') + table('e', 'laying', lr, 'tabel lipit')
    + foot('e', True, f'Se salvează {len(lr)} de rânduri în ghidul {LF}.')
    + live(f'Tabel citit: {len(lr)} de rânduri, toate valide.'), 'Import ghid')))

# 6 wrong format
B.append(('Format greșit (.pdf)', wrap('f',
    banner('fault', '<p><b>Ross 308 creștere.pdf nu se poate citi.</b> Încarcă un fișier .xlsx sau .csv, sau lipește tabelul.</p>', 'f-err')
    + DROP + foot('f', False, 'Alege un fișier .xlsx sau .csv.'), 'Import ghid')))

# 7 unreadable: no header
B.append(('Fișier necitibil: antetul nu e găsit', wrap('g',
    banner('fault', NOHEAD.format(f='Ross 308 creștere.csv'), 'g-err')
    + DROP + foot('g', False, 'Alege un fișier cu antet.'), 'Import ghid')))

# 8 empty
B.append(('Gol: doar antetul', wrap('h',
    src_line('h', 'Ross 308 creștere.csv', 'fișier .csv · tip creștere', counts(0, 0, 0), '')
    + empty_file('Ross 308 creștere.csv') + foot('h', False, EMPTY_NOTE) + live('Fișier citit: niciun rând.'), 'Import ghid')))

col = 'calc(var(--space-12) * 14)'  # 672px: two previews side by side at 1440
grid = ''.join(f'<div style="width:{col}">{cap(t)}{h}</div>' for t, h in B)
html = ('<!-- @dsCard group="Data" height=3032 subtitle="SYS11" -->\n'
        '<div class="ip-root ip-stack" style="gap:var(--space-8)">\n'
        f'<div>{cap("S20 · Ghiduri · import · desktop (lățime 672 px)")}'
        f'<div style="display:grid;grid-template-columns:repeat(2,{col});gap:var(--space-8) var(--space-6);align-items:start">{grid}</div></div>\n'
        f'<p class="ip-note" style="margin:0;max-width:calc({col} * 2)">Reale: valorile ghidurilor din guide-growth.json '
        '(creștere, sapt. 1–25, hrană g și greutate g ♀/♂, greutatea până la sapt. 23) și guide-laying.json (ouat, vârsta 19–65, '
        'hrană g ♀/♂; ouat % ♀ ilustrativ, D29). Ilustrative (D22): numele ghidurilor și ale fișierelor. Exemple de stare (D39): '
        'rândul 7 cu greutatea „#REF!” (în ghid 550), rândul 10 copia rândului 8 (SAPT 4 ♀), rândul 12 fără sex (SAPT 5 ♂); '
        'fișierul .pdf, cel fără antet și cel gol.</p>\n'
        '</div>\n')
(ROOT / 'GuideImportPreview').mkdir(exist_ok=True)
(ROOT / 'GuideImportPreview' / 'preview.html').write_text(html, encoding='utf-8')
print('ok', len(gr), len(er), nbad, len(lr))
