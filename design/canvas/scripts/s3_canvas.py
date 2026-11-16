"""python3 s3_canvas.py <canvas.json> : add the S3 boards (row below the S7 row) to canvas.json, keeping every other entry."""
import sys, json
p = sys.argv[1]; c = json.load(open(p, encoding='utf-8'))
H = json.load(open('/home/claude/miraj-impex-v3/design/canvas/scripts/s3_heights.json'))
s7 = [b for k, b in c['boards'].items() if k.startswith('S7-')]
y = max(b['y'] for b in s7); y = y + max(b['h'] for b in s7 if b['y'] == y) + 120
T = [('S3-seria-2026', 'S3 · Seria 2026'), ('S3-seria-2026--se-incarca', 'S3 · Seria 2026 · se încarcă'),
     ('S3-seria-2026--seria-inchisa', 'S3 · Seria 2026 · seria închisă (Seria 2025)'),
     ('S3-seria-2026--fara-acces', 'S3 · Seria 2026 · fără acces (Observator, doar Hîrbovăț)')]
x = 0
for stem, title in T:
    k = stem + '.dc.html'
    c['boards'][k] = {'x': x, 'y': y, 'w': 1440, 'h': H[stem], 'expand': 'fill', 'title': title}
    if k not in c['order']: c['order'].append(k)
    x += 1440 + 80
json.dump(c, open(p, 'w', encoding='utf-8'), ensure_ascii=False, indent=2)
print('row y', y)
