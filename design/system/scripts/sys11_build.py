"""Build the SYS11 preview (GuideImportPreview, S20 Ghiduri import, spec B2): python3 sys11_build.py
Guide figures come from design/data/extracts/guide-growth.json and guide-laying.json (real workbook guide values, both by
the birds' age week, D151: SAPT = age week, no offset); guide and file names are illustrative (D22). The laying rows are
the guide the app uses: lay % ♀ from Productivitate col I (SAPT 23–62, D184), cumulative mortality % ♀ from RealGhid, feed
g ♀/♂ (SAPT 19–65). Pasted, they carry one illustrative typed error (D185): week 53's lay % typed "645" (guide 64,5), refused
(D180). The three refused rows of the growth error state are state examples (D39)."""
import json, pathlib
import sys as _sys; _sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[2] / 'canvas/scripts')); import sexsign as SXS  # SX01: every ♀/♂ of the preview = the §72 hook
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
    'warn': '<path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"/>',
    'upload': '<path d="M12 15V3"/><path d="m7 8 5-5 5 5"/><path d="M5 21h14"/>',
}

def cap(text):
    return f'<span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">{text}</span>'

def fmt(v, unit='', dec=1):
    """Thin no-break thousands, true minus, decimal comma; a percentage keeps one decimal (89,0, as the guide detail),
    a cumulative mortality % two (7,25, as the guide's RealGhid column and S4)."""
    if v is None:
        return None
    if unit == '%':
        return f'{float(v):.{dec}f}'.replace('.', ',')
    if isinstance(v, float) and not v.is_integer():
        return f'{v:.1f}'.replace('.', ',')
    n = int(v)
    s = f'{abs(n):,}'.replace(',', NNBSP)
    return ('−' if n < 0 else '') + s

def num(n):
    return f'<span class="ip-num">{n}</span>'

SEX = {'F': ('♀', 'hen', 'găini'), 'M': ('♂', 'rooster', 'cocoși')}
COLS = {'growth': [('feed_g', 'Hrană g'), ('body_weight_g', 'Greutate g')],
        'laying': [('feed_g', 'Hrană g'), ('mortality_cum_pct', 'Mortalitate %'), ('lay_pct', 'Ouat %')]}
NAME = {'feed_g': 'hrana', 'body_weight_g': 'greutatea', 'mortality_cum_pct': 'mortalitatea', 'lay_pct': 'ouat %'}
UNIT = {'feed_g': 'g', 'body_weight_g': 'g', 'mortality_cum_pct': '%', 'lay_pct': '%'}
DEC = {'mortality_cum_pct': 2}           # the cumulative mortality (RealGhid) keeps its two decimals (7,25)
HEN_ONLY = {'mortality_cum_pct', 'lay_pct'}   # guide columns given on the ♀ rows only
FILE_COLS = [('feed_g', 'Hrană g'), ('body_weight_g', 'Greutate g'), ('mortality_cum_pct', 'Mortalitate %'), ('lay_pct', 'Ouat %'),
             ('hatching_pct', 'Incubabile %'), ('egg_weight_g', 'Masă ou g')]   # the guide columns a file can carry (B2), in DROP_NOTE order
def fmtk(key, v): return fmt(v, UNIT[key], DEC.get(key, 1))

# ------------------------------------------------------------------ rows (as read from the file; line 1 = header)
def growth_rows():
    rows = [dict(week=r['week'], age=None, sex=r['sex'], v={'feed_g': r['feed_g'], 'body_weight_g': r['body_weight_g']})
            for r in GROW]
    return rows

def laying_rows():
    """By age week (D151), as the guide stores them: SAPT 19–65, feed ♀/♂, lay % (Productivitate col I, D184) and cumulative
    mortality % (RealGhid) on ♀."""
    return [dict(week=r['age_week'], sex=r['sex'], v={k: r[k] for k, _ in COLS['laying']}) for r in LAY]

def laying_typed_rows():
    """The laying rows pasted with one illustrative typed error (D185): week 53's lay % typed "645" for the guide's 64,5, so
    that cell is refused when the table is pasted (D180)."""
    out = number(laying_rows())
    for r in out:
        if r['sex'] == 'F' and r['v']['lay_pct'] == TYPO_READ and r['week'] == TYPO_WEEK:
            r.update(bad={'lay_pct': TYPO_RAW}, kind='range')
    assert sum(1 for r in out if r.get('kind')) == 1
    return out
TYPO_WEEK, TYPO_RAW, TYPO_READ = 53, '645', 64.5   # D185: the guide's 64,5 typed without its comma (illustrative)
assert next(r['lay_pct'] for r in LAY if r['age_week'] == TYPO_WEEK and r['sex'] == 'F') == TYPO_READ   # Productivitate col I (D184)

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
    if k == 'range':
        return (f'SAPT {r["week"]} {sexg("F")}', f'ouat % „{r["bad"]["lay_pct"]}” e peste 100',
                f'Ouat % „{r["bad"]["lay_pct"]}” e peste 100. Scrie procentul cu virgulă, între 0 și 100.', 'lay_pct')
    if k == 'nan':
        return (f'SAPT {r["week"]} {sexg("M")}', 'greutatea „#REF!” nu e număr', 'Greutatea „#REF!” nu e număr. Scrie greutatea în grame.', 'body_weight_g')
    if k == 'dup':
        return (f'SAPT {r["week"]} {sexg("F")}', f'e deja pe rândul {r["dup_of"]}',
                f'Rândul {num(r["dup_of"])} are deja sapt. {num(r["week"])} {sexg("F")}. Păstrează un singur rând.', 'week')  # FW03: figures in Mono in the detail's error line
    if k == 'nosex':
        return (f'SAPT {r["week"]}', 'lipsește sexul', f'Lipsește sexul. Scrie G ({sexg("F")}) sau C ({sexg("M")}).', 'sex')
    return None

# ------------------------------------------------------------------ table
def head(kind):
    th = '<th scope="col" class="num">Rând</th>'
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
        v = fmtk(key, r['v'][key])
        if v is None:
            if key in HEN_ONLY and r['sex'] == 'M':   # the guide gives these for hens only: no note on every ♂ row
                why = 'nu se aplică la cocoși' if key == 'lay_pct' else 'ghidul nu o dă la cocoși'
                cells.append(f'<td class="num ip-gimp__none">–<span class="ip-sr"> {why}</span></td>')
            else:
                cells.append('<td class="num ip-gimp__none">–<span class="ip-sr"> fără date în ghid</span></td>')
                gaps.append(NAME[key])
        else:
            cells.append(f'<td class="num{tone}">{v}</td>')
    rs = reason(r)
    detail = ''
    if rs:  # FW03 (D150): the row keeps its height with the fault Lamp in Verificare; the reason sits in the detail row under it
        chk = '<span class="ip-gimp__why"><span class="ip-lamp ip-lamp--fault">refuzat</span></span>'
        detail = (f'<tr class="ip-table__detail"><td colspan="{len(cells) + 1}"><div class="ip-table__detailbody">'
                  f'<p class="ip-err ip-fieldmsg" id="{p}-r{r["line"]}-why">{svg(IC["fault"], "ip-icon ip-icon--14")}<span>{rs[2]}</span></p></div></td></tr>')
    elif gaps:
        chk = f'<span class="ip-gimp__gap">fără date în ghid: {", ".join(gaps)}</span>'
    else:
        chk = ''
    cls = ' class="is-refused"' if rs else ''
    return f'<tr{cls}>{"".join(cells)}<td class="ip-gimp__chk">{chk}</td></tr>{detail}'

# the laying table has three guide columns: its figure heads may wrap between words (additions §32, FIN14), so Verificare keeps
# the room for "fără date în ghid: mortalitatea, ouat %" on one line at 672 (one row height, FW04)
WRAP = {'growth': '', 'laying': ' ip-table--wraphead'}
def table(p, kind, rows, label, focus=None, busy=False, body=None, sum_=True):
    b = ' aria-busy="true"' if busy else ''
    d = f' aria-describedby="{p}-sum"' if sum_ else ''
    tb = body if body is not None else ''.join(row_html(p, kind, r, focus.get(r['line']) if focus else None) for r in rows)
    return (f'<div class="ip-gimp__box" tabindex="0" role="region" aria-label="Rânduri citite, {label}">'
            f'<table class="ip-table ip-gimp__t{WRAP[kind]}"{d}{b}>{head(kind)}<tbody>{tb}</tbody></table></div>')

# loading: six section 42 skeleton rows (tr.ip-table__skel, aria-hidden) under the real head; each bar is as wide as the
# column's figures will be (ch of the Mono cell: Rând, SAPT, Vârstă 2; Sex 1; hrană 3; greutate, ouat % 4) and sits where
# they will sit (numbers at the right edge, sex at the left); Verificare stays empty (a valid row has no text there).
CH = {'line': 2, 'week': 2, 'sex': 1, 'feed_g': 3, 'body_weight_g': 4, 'mortality_cum_pct': 4, 'lay_pct': 4}
def skel_body(kind, n=6):
    keys = ['line', 'week', 'sex'] + [k for k, _ in COLS[kind]]
    bar = lambda k: f'<span class="ip-skel ip-skel--ch{CH[k]}{"" if k == "sex" else " ip-skel--end"}"></span>'  # §42 ch widths (FIN11)
    tds = ''.join(f'<td class="{"ip-gimp__sex" if k == "sex" else "num"}">{bar(k)}</td>' for k in keys)
    return ''.join(f'<tr class="ip-table__skel" aria-hidden="true">{tds}<td class="ip-gimp__chk"></td></tr>' for _ in range(n))

# ------------------------------------------------------------------ parts
p_src = {}   # prefix → the source named on its line (a pasted table's fault Banner says "lipește-l din nou")
def src_line(p, name, note, counts, key='Alege alt fișier'):
    """key='' leaves the key out (the empty file: the EmptyState holds the one "Alege alt fișier")."""
    p_src[p] = name
    c = f'<p class="ip-gimp__counts" id="{p}-sum">{counts}</p>' if counts else f'<p class="ip-gimp__counts" id="{p}-sum"></p>'
    k = f'<button type="button" class="ip-btn ip-btn--ghost">{key}</button>' if key else ''
    return (f'<div class="ip-gimp__src"><span class="ip-gimp__file"><b>{name}</b><span class="ip-note">{note}</span></span>'
            f'{c}{k}</div>')

def refused_words(n): return f'{num(n)} {"rând refuzat" if n == 1 else "rânduri refuzate"}'
def counts(read, valid, refused):
    s = f'{num(read)} rânduri citite · {num(valid)} valide'
    return s + (f' · {num(refused)} {"refuzat" if refused == 1 else "refuzate"}' if refused else '')

def banner(kind, body, bid=''):
    i = f' id="{bid}"' if bid else ''
    return f'<div class="ip-banner ip-banner--{kind}"{i}>{svg(IC[kind])}<div class="ip-gimp__bn">{body}</div></div>'

def err_banner(p, rows):
    bad = [r for r in rows if reason(r)]
    li = ''.join(f'<li><a class="ip-link" href="#{p}-r{r["line"]}-{reason(r)[3]}">Rândul {r["line"]} · {reason(r)[0]}</a>: '
                 f'{reason(r)[1]}.</li>' for r in bad)
    n = len(bad)
    fix = 'Corectează-l' if n == 1 else 'Corectează-le'
    src = 'în tabel și lipește-l' if 'Tabel lipit' in p_src.get(p, '') else 'în fișier și încarcă-l'
    return banner('fault', f'<p><b>{refused_words(n)}.</b> {fix} {src} din nou; '
                           f'până atunci ghidul nu se poate salva.</p><ul class="ip-gimp__errs">{li}</ul>', f'{p}-errs')

def info_banner(text):
    return banner('info', f'<p><b>Lipsurile nu blochează salvarea.</b> {text}</p>')

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

def gap_weeks(ws):
    """[24, 25] → '24–25'; runs joined with ' și '."""
    out = []
    for w in sorted(ws):
        if out and out[-1][1] == w - 1: out[-1][1] = w
        else: out.append([w, w])
    return ' și '.join(f'{a}–{b}' if a != b else f'{a}' for a, b in out)
def info_text(kind, rows, src='fișierul'):
    """The info Banner's text (doesn't block saving), from the rows read: the weeks each column leaves blank (♀ only for the
    hen-only columns), the file rows they sit on (growth: one run) and the guide columns the file lacks."""
    parts = {}   # weeks text → [column names] (columns blank on the same weeks share one clause)
    for key, lbl in COLS[kind]:
        sexes = 'F' if key in HEN_ONLY else 'FM'
        ws = sorted({r['week'] for r in rows if r['sex'] and r['sex'] in sexes and r['v'][key] is None and key not in r.get('bad', {})})
        if not ws: continue
        lines = [r['line'] for r in rows if r['sex'] and r['sex'] in sexes and r['v'][key] is None]
        where = f' (rândurile {lines[0]}–{lines[-1]})' if lines[-1] - lines[0] == len(lines) - 1 else ''
        parts.setdefault(f'sapt. {gap_weeks(ws)}{where}', []).append(NAME[key])
    parts = [f'{" și ".join(ns)} {"lipsește" if len(ns) == 1 else "lipsesc"} la {w}' for w, ns in parts.items()]
    absent = [lbl for k, lbl in FILE_COLS if k not in dict(COLS[kind])]
    t = (', '.join(parts[:-1]) + ' și ' + parts[-1] if len(parts) > 1 else ''.join(parts))
    t = t[0].upper() + t[1:] if t else ''
    nu = f'{src} nu are coloanele {", ".join(absent[:-1])} și {absent[-1]}'
    s = (f'{t}, iar {nu}.' if t else f'{nu[0].upper()}{nu[1:]}.') + ' În ghid apar ca „–” fără date în ghid.'
    if kind == 'laying':
        s += (' Mortalitate % și Ouat % se citesc doar pe rândurile <span class="ip-sex ip-sex--hen" aria-hidden="true">♀</span>'
              '<span class="ip-sr">de găini</span>.')
    return s

def from_transfer_rows():
    """D180 state example (D39 file): the real laying rows in a file that counts its weeks from the transfer, so its first row
    is SAPT 1 (the guide's first age week becomes 1). Same figures, other week numbers."""
    k0 = min(r['age_week'] for r in LAY) - 1
    out = [dict(r, week=r['week'] - k0) for r in laying_rows()]
    assert min(r['week'] for r in out) == 1
    return number(out)
WEEK1 = ('<p><b>Ghidul de ouat începe la SAPT 1.</b> Săptămâna 1 e prima săptămână de viață. Ghidurile de ouat încep de obicei la '
         'SAPT 19–23; verifică dacă fișierul numără săptămânile de la transfer.</p>')   # spec B2 round 2; D180: file level, Save allowed
def week1_banner(p):
    """D180: a laying file whose first row is week 1: one file-level warn Banner (role=status); the guide can still be saved."""
    return banner('warn', WEEK1, f'{p}-wk1').replace(f'id="{p}-wk1"', f'id="{p}-wk1" role="status"', 1)

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
INFO_G = info_text('growth', gr)
GW = sorted({r['week'] for r in gr}); GWT = [r['week'] for r in gr if r['v']['body_weight_g'] is None]
B.append((f'Toate valide · creștere (real, {len(gr)} de rânduri; cutia derulează)', wrap('b',
    src_line('b', GF, 'fișier .xlsx · tip creștere', counts(len(gr), len(gr), 0))
    + banner('ok', f'<p><b>Toate cele {len(gr)} de rânduri sunt valide.</b> Verifică cifrele și salvează ghidul.</p>')
    + info_banner(INFO_G)
    + bar('b') + table('b', 'growth', gr, GF)
    + foot('b', True, f'Se salvează {len(gr)} de rânduri în ghidul Ross 308 creștere.')
    + live(f'Fișier citit: {len(gr)} de rânduri, toate valide.'), 'Import ghid')))

# 2b all valid, end of the box (the partial weeks)
B.append((f'Toate valide · capătul tabelului (derulat): SAPT {gap_weeks(GWT)} fără greutate', wrap('b2',
    bar('b2') + table('b2', 'growth', [r for r in gr if r['week'] >= GW[-1] - 5], GF, sum_=False), 'Import ghid')))

# 3 refused rows
er = error_rows()
nbad = sum(1 for r in er if reason(r))
INFO_E = info_text('growth', er)
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

# 5 laying, pasted: the guide the app uses (D184) by age week (D151), week 53 lay % typed "645" (D185, illustrative) refused
lr = number(laying_rows())        # the rows as the guide stores them (64,5): S20's new-guide file, all valid
lt = laying_typed_rows()          # the same rows pasted with week 53 typed "645": one cell refused
tb = next(r for r in lt if r.get('kind'))
LF = 'Ross 308 ouat'
LW = sorted({r['week'] for r in lr})
INFO_L = info_text('laying', lt, 'tabelul')
NOTE_L = f'{refused_words(1)}: corectează tabelul și lipește-l din nou.'
B.append((f'Ouat · tabel lipit (ghidul din aplicație, real, {len(lt)} de rânduri, SAPT {LW[0]}–{LW[-1]}): rândul {tb["line"]} refuzat, '
          f'ouat % „{TYPO_RAW}” la SAPT {TYPO_WEEK} ♀ (greșeală de tastare ilustrativă, D185)', wrap('e',
    src_line('e', 'Tabel lipit', 'din clipboard · tip ouat', counts(len(lt), len(lt) - 1, 1), 'Lipește alt tabel')
    + err_banner('e', lt) + info_banner(INFO_L)
    + seg('e', len(lt), 1, 'all') + table('e', 'laying', lt, 'tabel lipit')
    + foot('e', False, NOTE_L)
    + live(f'Tabel citit: {len(lt)} de rânduri, 1 refuzat.'), 'Import ghid')))

# 5b the same, scrolled to the refused row (focus on its cell, reached from the Banner link)
B.append((f'Ouat · derulat la rândul {tb["line"]} · focus pe celula „{TYPO_RAW}” (după linkul din rezumat)', wrap('e2',
    bar('e2') + table('e2', 'laying', [r for r in lt if TYPO_WEEK - 3 <= r['week'] <= TYPO_WEEK + 2], 'tabel lipit',
                      focus={tb['line']: 'lay_pct'}, sum_=False), 'Import ghid')))

# 5c laying file numbered from the transfer (D39 file, real figures): file-level warn Banner, Save allowed (D180)
wr = from_transfer_rows()
WF = 'Ross 308 ouat.xlsx'
B.append(('Ouat · fișier numărat de la transfer (exemplu de stare D39, cifre reale): începe la SAPT 1, avertisment, se poate salva', wrap('e3',
    src_line('e3', WF, 'fișier .xlsx · tip ouat', counts(len(wr), len(wr), 0))
    + week1_banner('e3') + info_banner(info_text('laying', wr))
    + bar('e3') + table('e3', 'laying', wr, WF)
    + foot('e3', True, f'Se salvează {len(wr)} de rânduri în ghidul {LF}, de la SAPT 1.')
    + live(f'Fișier citit: {len(wr)} de rânduri, toate valide; ghidul începe la SAPT 1.'), 'Import ghid')))

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
html = ('<!-- @dsCard group="Data" height=3876 subtitle="SYS11" -->\n'
        '<div class="ip-root ip-stack" style="gap:var(--space-8)">\n'
        f'<div>{cap("S20 · Ghiduri · import · desktop (lățime 672 px)")}'
        f'<div style="display:grid;grid-template-columns:repeat(2,{col});gap:var(--space-8) var(--space-6);align-items:start">{grid}</div></div>\n'
        f'<p class="ip-note" style="margin:0;max-width:calc({col} * 2)">Reale: valorile ghidurilor din guide-growth.json '
        f'(creștere, sapt. {GW[0]}–{GW[-1]}, hrană g și greutate g ♀/♂, greutatea până la sapt. {min(GWT) - 1}) și guide-laying.json (ouat, '
        f'sapt. {LW[0]}–{LW[-1]}, hrană g ♀/♂; ouat % ♀ din Productivitate col I (D184) și mortalitate cumulată % ♀ din RealGhid, sapt. {min(r["week"] for r in lr if r["v"]["lay_pct"] is not None)}–'
        f'{max(r["week"] for r in lr if r["v"]["lay_pct"] is not None)}), pe săptămâna de vârstă (SAPT, D151). Ilustrative (D22): numele ghidurilor și ale fișierelor. '
        f'Exemplu de stare (D185): în tabelul lipit, ouat % la sapt. {TYPO_WEEK} scris „{TYPO_RAW}” (în ghid {fmt(TYPO_READ, "%")}), refuzat. '
        f'Exemple de stare (D39): rândul 7 cu greutatea „#REF!” (în ghid {fmt(next(r["v"]["body_weight_g"] for r in gr if r["week"] == 3 and r["sex"] == "M"))}), '
        'rândul 10 copia rândului 8 (SAPT 4 ♀), rândul 12 fără sex (SAPT 5 ♂); '
        'fișierul .pdf, cel fără antet și cel gol; fișierul de ouat numărat de la transfer (SAPT 1, D180).</p>\n'
        '</div>\n')
(ROOT / 'GuideImportPreview').mkdir(exist_ok=True)
(ROOT / 'GuideImportPreview' / 'preview.html').write_text(SXS.sexify(html), encoding='utf-8')
print('ok', len(gr), len(er), nbad, len(lr), tb['line'])
