"""literal_scan.py · the RF check (D158g): no superseded "today" literals on the boards or in their generators (RF00 §3).
python3 -I literal_scan.py <repo_root>               every row: lists the hits of PORTED rows (exit 1 if any), counts the rest
python3 -I literal_scan.py <repo_root> --row S4R …   only these rows; any hit fails (exit 1, file:line per hit)
python3 -I literal_scan.py <repo_root> FILE …        only these files (boards or scripts); any hit fails
A row = its boards (design/canvas/project/<prefix>*.dc.html) + its generators (ROWS). Old literals: the Hîrbovăț codes before
D152 (H1.x/H3.x/H4.x, hall H1/H3/H4 next to "Hîrbovăț"), 16.11.2026 as today, SAPT 40 / sapt. 40 as the current week, the
cycle start 10.02.2026, the superseded sample's figures RF00 lists (25 252, 25 257, 2 685, 26 765, 29 825, 28 252, 3 967, 18 688, 3 691,
31 450, 4 121/4 119/4 130, 8 475),
the sample file, cycle_week, "age − 18" and "47 săpt." (Seria 2025 is SAPT 19–65 now, D151); RF14: the RealGhid lay guide
(53,3 / 54,5 / 55,7 / 56,9 %, deviations −12,2 / −8,0 / −18,4 pct, "sub ghid din 02.11", the paste typo "605"; D184/D185);
S6C: the D164 Cat. 1 placeholder on the Zilnic Ouă group (457 / 149 / 133, and `'ec1': e['eggs'] -`; D183).
Allowed: Telița codes (H1/H3 with "Telița" in reach), Seria 2025 weeks (SAPT 40/41 with "Seria 2025" in reach), a script line
marked `literal-ok` (say why on the line), and the ALLOW entries below (board stem prefix, pattern id, context regex) for
what RF00 keeps (D39 examples, Telița replays). When a row's RF job lands, add it to PORTED."""
import sys, re, pathlib

HERE = pathlib.Path(__file__).resolve().parent
SP = '[   ]'
PATTERNS = [  # id, regex, applies to ('dc' board, 'py' script)
    ('old-code', re.compile(r'(?<![\w-])H[134]\.[12](?!\d)'), 'dc py'),
    ('old-hall', re.compile(r'(?<![\w.\-/#])H[134](?![\w.\-])'), 'dc py'),
    ('today-16.11', re.compile(r'16\.11\.2026|\bazi 16\.11\b|2026-11-16'), 'dc py'),
    ('sapt-40', re.compile(rf'(?i)\bsapt\.?{SP}*40\b'), 'dc py'),
    ('sapt-41', re.compile(rf'(?i)\bsapt\.?{SP}*41\b'), 'dc py'),  # D161 (RF02): the old sample's next week
    ('start-10.02', re.compile(r'10\.02\.2026|2026-02-10'), 'dc py'),
    ('sample-figure', re.compile(rf'(?<![\d  ])(?:25{SP}?252|25{SP}?257|2{SP}685|26{SP}?765|29{SP}?825|28{SP}?252|3{SP}967|18{SP}?688|3{SP}691|31{SP}?450|4{SP}121|4{SP}119|4{SP}130|8{SP}475)(?![\d])'), 'dc py'),
    ('sample-file', re.compile(r'sample-hirbovat|sample_hirbovat'), 'dc py'),
    ('cycle-week', re.compile(r'cycle_week|data_start_cycle_week|week_today'), 'py'),
    ('age-18', re.compile(r'(?i)(?:age|vârst\w*)\s?[−-]\s?18\b'), 'dc py'),
    ('weeks-47', re.compile(rf'\b47{SP}?(?:săpt|săptămâni)'), 'dc py'),
    # RF14 (D184): the RealGhid lay guide before Productivitate col I: SAPT 56–59 guide 56,9 / 55,7 / 54,5 / 53,3 %, the 18.11
    # deviations farm −12,2 · H5 −8,0 · H6 −18,4 pts (SAPT 59's old −12,9 is a real hall figure now), the H5 alert "sub ghid din
    # 02.11", the S20/SYS11 guide cells and the RealGhid paste typo "605" (D185: "645")
    ('guide-lay-old', re.compile(rf'(?<![\d,\u202f])(?:53,3|54,5|55,7|56,9){SP}?%'), 'dc py'),
    ('guide-cell-old', re.compile(r'(?:is-hen|ip-sex--hen)">(?:53,3|54,5|55,7|56,9|60,5)<'), 'dc'),
    ('lay-dev-old', re.compile(rf'[−-](?:12,2|8,0|18,4)(?:</span>)?{SP}?(?:pct|puncte)'), 'dc py'),
    ('below-since-old', re.compile(r'sub ghid din (?:<[^>]*>)?02\.11'), 'dc py'),
    ('typo-605', re.compile(r'„605”|"605"'), 'dc py'),
    # S6C (D183): the D164 Zilnic placeholder Cat. 1 = eggs − incubabile (18.11 H5 457 · H6 149; 17.11 H5 133 · H6 149) in the
    # Ouă group's Cat. 1 input, and the rule that computed it; Cat. 1–3 are the day's real counts now
    ('cat1-placeholder', re.compile(r'id="ec1-H\d"[^>]*value="(?:457|149|133)"'), 'dc'),
    ('cat1-rule-d164', re.compile(r"'ec1':\s*e\['eggs'\]\s*-"), 'py'),
]
NEAR = 160  # chars around a board hit searched for context
CTX_OK = {'old-hall': re.compile(r'Teli[țţt]a'), 'sapt-40': re.compile(r'Seria 2025'), 'sapt-41': re.compile(r'Seria 2025')}
CTX_NEED = {'old-hall': re.compile(r'Hîrbovăț|Hirbovat|Hala H[134]\b|H1, H3|H3, H4')}  # boards: a bare H1/H3/H4 only counts at Hîrbovăț
ALLOW = [  # (board stem prefix, pattern id, regex the hit's context must match) — what RF00 keeps; add with the reason
    # RF03 → S14N: the end of a past age week (SAPT 58 = 10.11–16.11.2026) in the PeriodTable week keys' hidden names: S11 Consum
    # (also under S13's SidePanel) and S14 Producție
    ('S11-', 'today-16.11', r'SAPT 58, \d\d\.\d\d–16\.11\.2026'),
    ('S13-livrare-hrana--desktop', 'today-16.11', r'SAPT 58, \d\d\.\d\d–16\.11\.2026'),
    ('S14-', 'today-16.11', r'SAPT 58, \d\d\.\d\d–16\.11\.2026'),
    # S14N: the real Hrană ♀ 3 691,7 kg of 02.11.2026 (S11 Consum day row, as on S4 Hrană)
    ('S11-', 'sample-figure', r'is-hen(?: ip-ptable__[\w-]+)*">3\u202f691,7<'),   # SYS15/SYS16: + the column group classes (__g, __g--<type>, __gs)
    ('S13-livrare-hrana--desktop', 'sample-figure', r'is-hen(?: ip-ptable__[\w-]+)*">3\u202f691,7<'),
    # RF03: Telița's real last delivery (10.02.2026), on the S11 "fără livrări în perioadă" board
    ('S11-rapoarte-hrana--fara-livrari', 'start-10.02', r'Ultima livrare'),
    # S4N: 16.11.2026 is a past day of SAPT 58 (10.11–16.11.2026) in the PeriodTable week keys and the Păsări log (real 16.11 records)
    ('S4-ferma', 'today-16.11', r'SAPT 58, \d\d\.\d\d–16\.11\.2026'),
    ('S4-ferma--pasari', 'today-16.11', r'16\.11\.2026, \d\d:\d\d · SAPT 58|(?:Mortalitate|Vânzare H\d\.\d), 16\.11\.2026'),
    # DATA04 (D168): with one record per sale form, the 9 "Ultimele înregistrări" rows reach the real 16.11 sale (SAPT 58, past day)
    ('S4-ferma', 'today-16.11', r'16\.11\.2026, \d\d:\d\d · SAPT 58 · '),
    ('S9-transfer', 'today-16.11', r'16\.11\.2026, \d\d:\d\d · SAPT 58 · '),
    ('S10-vanzare', 'today-16.11', r'16\.11\.2026, \d\d:\d\d · SAPT 58 · '),
    # S4NG: the S9/S10 desktop backdrop is the Rezumat page with its table: SAPT 58 ends 16.11.2026 (the week's hidden name)
    ('S9-transfer--desktop', 'today-16.11', r'SAPT 58, \d\d\.\d\d–16\.11\.2026'),
    ('S10-vanzare--desktop', 'today-16.11', r'SAPT 58, \d\d\.\d\d–16\.11\.2026'),
    # S4N: real figures that equal old sample ones: Hrană ♀ 3 691,7 kg on 02.11.2026 (day row), Hala 6 ♀ 8 475 alive on a SAPT 57 day (Păsări pe hale)
    ('S4-ferma', 'sample-figure', r'is-hen(?: ip-ptable__[\w-]+)*">3\u202f691,7<'),   # SYS15/SYS16: + the column group classes (__g, __g--<type>, __gs)
    ('S4-ferma--pasari', 'sample-figure', r'is-hen(?: ip-ptable__[\w-]+)*">8\u202f475<'),   # S4NG: În viață is a section's first column (D167); SYS15: + group classes
    # RF08: the S16 desktops show S4 · Păsări (table-first page) beside / under the panel: 16.11.2026 is a past day of SAPT 58 there
    ('S16-corecteaza--desktop', 'today-16.11', r'SAPT 58, \d\d\.\d\d–16\.11\.2026|16\.11\.2026, \d\d:\d\d · SAPT 58|(?:Mortalitate|Vânzare H\d\.\d), 16\.11\.2026'),
    # RF09: the S17 observer board is S4 · Păsări (table-first page): 16.11.2026 is a past day of SAPT 58 there
    ('S17-istoric-modificari--observator-marcaj-text', 'today-16.11', r'SAPT 58, \d\d\.\d\d–16\.11\.2026|16\.11\.2026, \d\d:\d\d · SAPT 58|(?:Mortalitate|Vânzare H\d\.\d), 16\.11\.2026'),
    # S3N: the series table's week keys: SAPT 58 ends 16.11.2026 (a past age week; the hidden week name)
    ('S3-seria-2026', 'today-16.11', r'SAPT 58, \d\d\.\d\d–16\.11\.2026'),
    # RF07: Telița's real weighing of 10.02.2026 (H2 ♀, SAPT 19) is last week's on the S8 zi trecută replay (17.02.2026)
    ('S8-cantarire--zi-trecuta', 'start-10.02', r'Săpt\. trecută · 10\.02\.2026'),
    # S20N: the growth guide's real ♀ weight at SAPT 22 (2 685 g) in the guide tables, and a small-curve point ("9.4 130.7")
    ('S20-', 'sample-figure', r'(?:is-hen|ip-sex--hen)">2\u202f685<'),
    ('S20-', 'sample-figure', r'\d,\d+\.4 130\.\d,\d'),
    # RF14: system previews scanned too; they keep what their boards keep (S20 growth guide ♀ 2 685 g; S4 PeriodTable's past
    # SAPT 58 week key, Hrană ♀ 3 691,7 kg on 02.11.2026 and Hala 6 ♀ 8 475 on a SAPT 57 day)
    ('GuideImportPreview/', 'sample-figure', r'is-hen">2\u202f685<'),
    ('PeriodTable/', 'today-16.11', r'SAPT 58, \d\d\.\d\d–16\.11\.2026'),
    # MO01 (D196): the motion demos' PeriodTable is sys14's (same week-key hidden names: SAPT 58 ends on 16.11.2026)
    ('M1-miscare', 'today-16.11', r'SAPT 58, \d\d\.\d\d–16\.11\.2026'),
    ('Motion/', 'today-16.11', r'SAPT 58, \d\d\.\d\d–16\.11\.2026'),
    ('PeriodTable/', 'sample-figure', r'is-hen(?: ip-ptable__[\w-]+)*">(?:3\u202f691,7|8\u202f475)<'),
]
# README order. (id, board stem prefixes, generator scripts)
ROWS = [
    ('S4R', ['S4-ferma-rezumat'], ['realdata.py', 's4s_build.py', 's4_build.py', 's4s_render.py', 'canvas_rows.py', 's4ns_build.py', 's4ns_render.py']),  # S4NS: its boards are S4-ferma-rezumat--* and S4-ferma--pasari--telefon*
    ('S4T', ['S4-ferma--'], ['s4t_build.py', 's4t_render.py']),
    ('S7', ['S7-'], ['s7_build.py', 's7s_build.py']),
    ('S3', ['S3-'], ['s3_build.py', 's3_render.py']),
    ('S11', ['S11-'], ['s11_build.py', 's11_render.py']),
    ('S12', ['S12-'], ['s12_build.py', 's12_render.py']),
    ('S13', ['S13-'], ['s13_build.py', 's13_render.py']),
    ('S6', ['S6-'], ['s6_build.py', 's6s_build.py', 's6_render.py', 's6s_render.py']),
    ('S9', ['S9-'], ['s9_build.py', 's9_render.py']),
    ('S10', ['S10-'], ['s10_build.py', 's10_render.py']),
    ('S2', ['S2-'], ['s2_build.py', 's2s_build.py', 's2_render.py']),
    ('S14', ['S14-'], ['s14_build.py', 's14s_build.py', 's14_render.py']),
    ('S8', ['S8-'], ['s8_build.py', 's8s_build.py', 's8_render.py']),
    ('S15', ['S15-'], ['s15_build.py', 's15s_build.py', 's15_render.py', 's15s_render.py']),
    ('S16', ['S16-'], ['s16_build.py', 's16s_build.py', 's16_render.py', 's16s_render.py']),
    ('S17', ['S17-'], ['s17_build.py', 's17s_build.py', 's17_render.py']),
    ('S18', ['S18-'], ['s18_build.py', 's18s_build.py', 's18_render.py']),
    ('S19', ['S19-'], ['s19_build.py', 's19s_build.py', 's19_render.py']),
    ('S20', ['S20-'], ['s20_build.py', 's20s_build.py', 's20_render.py']),
    ('S21', ['S21-'], ['s21_build.py', 's21_render.py']),
    ('S22', ['S22-'], ['s22_build.py', 's22s_build.py', 's22_render.py']),
    ('S1', ['S1-'], ['s1_build.py', 's1_render.py']),
    ('SH', ['Main'], ['sstate_css.py', 'skel.py']),
    # RF11: system previews/READMEs a row's job ports (no boards); paths from this folder
    ('SYS10', [], ['../../system/scripts/sys10_build.py', '../../system/components/SectionTree/README.md',
                   '../../system/components/SectionTree/preview.html']),
    # S20N: the GuideImportPreview builder, README and preview (its helpers are exec'd by s20_build)
    ('SYS11', [], ['../../system/scripts/sys11_build.py', '../../system/scripts/sys11_render.py',
                   '../../system/components/GuideImportPreview/README.md', '../../system/components/GuideImportPreview/preview.html']),
    # RF13: the UserRow/OneTimePassword (S22) and LoginCard (S1) builders and previews (their tops are exec'd by s22/s1)
    ('SYS12', [], ['../../system/scripts/sys12_build.py', '../../system/components/UserRow/preview.html',
                   '../../system/components/OneTimePassword/preview.html']),
    ('SYS13', [], ['../../system/scripts/sys13_build.py', '../../system/components/LoginCard/preview.html']),
    # RF14: the PeriodTable builder and preview (S4/S3/S14 Producție guide columns, D184)
    ('SYS14', [], ['../../system/scripts/sys14_build.py', '../../system/components/PeriodTable/preview.html']),
]
PORTED = {'S4R', 'S4T', 'S7', 'S11', 'S12', 'S13', 'S6', 'S9', 'S10', 'S8', 'S2', 'S15', 'S16', 'S3', 'S14', 'S17', 'S18', 'S19', 'SYS10', 'S21', 'S20', 'SYS11', 'S22', 'S1', 'SYS12', 'SYS13', 'SYS14'}  # rows whose RF job landed: their hits fail the default run


def scan(path):
    """[(line, pattern id, snippet)] for one board or script."""
    kind = 'py' if path.suffix == '.py' else 'dc'   # RF14: system previews and READMEs are scanned as markup too
    stem = (path.name[:-8] if path.name.endswith('.dc.html') else path.name if kind == 'py'
            else f'{path.parent.name}/{path.name}')   # a system file: 'PeriodTable/preview.html' (ALLOW by component)
    if path.name == 'literal_scan.py': return []
    hits = []
    for i, line in enumerate(path.read_text(encoding='utf-8').splitlines(), 1):
        if kind == 'py' and 'literal-ok' in line: continue
        for pid, rx, kinds in PATTERNS:
            if kind not in kinds: continue
            for m in rx.finditer(line):
                a, b = max(0, m.start() - NEAR), m.end() + NEAR
                ctx = line[a:b] if kind == 'dc' else line
                if pid in CTX_OK and CTX_OK[pid].search(ctx): continue
                if kind == 'dc' and pid in CTX_NEED and not CTX_NEED[pid].search(ctx): continue
                if kind == 'py' and pid == 'old-hall' and not re.search(r'["\']', line.split('#')[0]): continue  # code, not prose
                if any(stem.startswith(p) and q == pid and re.search(r, ctx) for p, q, r in ALLOW): continue
                snip = line[max(0, m.start() - 40):m.end() + 40].replace(' ', ' ').strip()
                hits.append((i, pid, snip))
    return hits


def row_files(root, row):
    _, prefixes, scripts = row
    proj = root / 'design/canvas/project'
    boards = sorted(p for p in proj.glob('*.dc.html') if any(p.name.startswith(x) for x in prefixes))
    return boards + [(HERE / s).resolve() for s in scripts if (HERE / s).exists()]


def report(root, files, show=True):
    n = 0
    for f in files:
        for i, pid, snip in scan(f):
            n += 1
            if show: print(f'{f.relative_to(root) if f.is_relative_to(root) else f}:{i}: [{pid}] …{snip}…')
    return n


def main(argv):
    root = pathlib.Path(argv[1]).resolve(); rest = argv[2:]
    if rest and rest[0] != '--row':  # explicit files
        n = report(root, [pathlib.Path(x).resolve() for x in rest])
        print(f'{n} hits'); return 1 if n else 0
    want = [rest[i + 1] for i, x in enumerate(rest) if x == '--row']
    rows = [r for r in ROWS if not want or r[0] in want]
    assert not want or len(rows) == len(want), f'unknown row in {want}; rows: {[r[0] for r in ROWS]}'
    bad = 0
    for r in rows:
        fails = bool(want) or r[0] in PORTED
        files = row_files(root, r)
        nb = report(root, [f for f in files if f.suffix != '.py'], show=fails)   # boards (+ system previews/READMEs, RF14)
        ns = report(root, [f for f in files if f.suffix == '.py'], show=fails)
        state = 'PORTED' if r[0] in PORTED else 'not ported'
        print(f'{r[0]:4} {state:10} boards {nb:4} · scripts {ns:4}' + ('  FAIL' if fails and nb + ns else ''))
        if fails: bad += nb + ns
    return 1 if bad else 0


if __name__ == '__main__':
    sys.exit(main(sys.argv))
