"""sex_scan.py · the SX01 check (D167 5, D170, additions.css §72): every ♀/♂ on the boards and the system previews is drawn as
the §72 icon, none as a plain font glyph (no font in the stack has the signs; the fallback drew them ~5 px).
python3 -I sex_scan.py <repo_root>            boards (design/canvas/project/*.dc.html) + previews (design/system/components/*/preview.html)
python3 -I sex_scan.py <repo_root> FILE …     only these files
Lists file:line · context of every sign outside the hook, and every standalone hook span (no child element, not in td.num)
holding other text than one sign (§72 draws it at font-size 0, so that text would vanish); exit 1 if any. A sign is in the hook when its text node is
- the only text of span.ip-sex / span.is-hen / span.is-rooster (no child element; the standalone 16/14 px icon), or
- the text before the <b> figure of span.ip-sex ("♀ <b>12 561</b>", the sign with its figure), or
- the only text of td.ip-gimp__sex (GuideImportPreview),
and the span is not in a context §72 leaves as text (td.num > *; a figure span after a signed .s4-capi / .ip-stotals__key). Text that is not drawn is skipped: <script>/<style>/<title>,
visually hidden .ip-sr (chart data tables, hidden words), [hidden], and attribute values (aria-label …)."""
import sys, re, pathlib
from bs4 import BeautifulSoup, NavigableString, Comment

SIGN = re.compile('[♀♂]')
SKIP = {'script', 'style', 'title', 'template'}
HOOK = {'ip-sex', 'is-hen', 'is-rooster'}


def cls(el):
    return set(el.get('class') or []) if el is not None and getattr(el, 'name', None) else set()


def excluded(span):
    """§72's :not(...) contexts, where a hook's sign stays text: td.num > * (figures under a signed head), and for the sign
    with its figure also the value after a signed .s4-capi / .ip-stotals__key (SX01 dropped the cost foot, eggs base and
    FarmSummary sub-line exclusions: those lines now write the standalone sign)"""
    p = span.parent
    if p is None:
        return False
    if p.name == 'td' and 'num' in cls(p):
        return True
    if span.find('b', recursive=False) is not None:   # a figure under a signed key: no sign drawn
        prev = p.find_previous_sibling()
        if prev is not None and cls(prev) & {'s4-capi', 'ip-stotals__key'} and prev.find(class_='ip-sex', recursive=False):
            return True
    return False


def hooked(node):
    el = node.parent
    c = cls(el)
    txt = str(node)
    if el.name == 'td' and 'ip-gimp__sex' in c:
        return SIGN.fullmatch(txt.strip()) is not None
    if el.name != 'span' or not (c & HOOK):
        return False
    if excluded(el):
        return False
    kids = [k for k in el.children if getattr(k, 'name', None)]
    if not kids:
        return SIGN.fullmatch(txt.strip()) is not None and sum(1 for k in el.children if isinstance(k, NavigableString) and k.strip()) == 1
    if 'ip-sex' in c and [k.name for k in kids] == ['b'] and node.next_sibling is kids[0]:
        return SIGN.fullmatch(txt.strip()) is not None and node.previous_sibling is None
    return False


def scan(path):
    text = path.read_text(encoding='utf-8')
    if not SIGN.search(text) and not re.search(r'ip-sex|is-hen|is-rooster', text):
        return []
    soup = BeautifulSoup(text, 'html.parser')
    hits = []
    for el in soup.find_all('span'):
        txt = el.get_text()
        if cls(el) & HOOK and not el.find(True) and txt.strip() and not SIGN.fullmatch(txt.strip()) and not excluded(el):   # §72 sets it at font-size 0
            hits.append(f"{path}:{el.sourceline} · hidden text in a standalone hook {' '.join(txt.split())[:60]!r}")
    for node in soup.find_all(string=SIGN):
        if isinstance(node, Comment) or any(p.name in SKIP or 'ip-sr' in cls(p) or p.has_attr('hidden') for p in node.parents if p.name != '[document]'):
            continue
        if hooked(node):
            continue
        el = node.parent
        line = (getattr(node, 'sourceline', None) or getattr(el, 'sourceline', None) or 0)
        where = el.name + ''.join('.' + x for x in (el.get('class') or []))
        hits.append(f"{path}:{line} · <{where}> {' '.join(str(node).split())[:90]!r}")
    return hits


def main(argv):
    root = pathlib.Path(argv[1])
    files = [pathlib.Path(f) for f in argv[2:]] or (
        sorted((root / 'design/canvas/project').glob('*.dc.html')) + sorted((root / 'design/system/components').glob('*/preview.html')))
    hits = [h for f in files for h in scan(f)]
    for h in hits:
        print(h)
    print(f'sex_scan: {len(files)} files, {len(hits)} plain sign(s) outside §72', file=sys.stderr)
    return 1 if hits else 0


if __name__ == '__main__':
    sys.exit(main(sys.argv))
