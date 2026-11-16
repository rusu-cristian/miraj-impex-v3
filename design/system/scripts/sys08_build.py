"""Build SYS08 previews (MissingRow, EntryKeys, UnpricedCount): python3 sys08_build.py"""
import pathlib
ROOT = pathlib.Path('/home/claude/miraj-impex-v3/design/system/components')
N = ' '  # thin no-break space for thousands

def svg(paths, cls='ip-icon'):
    return f'<svg class="{cls}" viewBox="0 0 24 24" aria-hidden="true">{paths}</svg>'

IC = {
    'mortality': '<path d="M12 3a7 7 0 0 0-7 7v4l2 2v3h10v-3l2-2v-4a7 7 0 0 0-7-7z"/><circle cx="9.5" cy="11" r="1.2"/><circle cx="14.5" cy="11" r="1.2"/><path d="M11 19v-2M13 19v-2"/>',
    'wheat': '<path d="M12 21V9M12 9c-3 0-4-2-4-5 3 0 4 2 4 5zM12 9c3 0 4-2 4-5-3 0-4 2-4 5zM12 15c-3 0-4-2-4-5 3 0 4 2 4 5zM12 15c3 0 4-2 4-5-3 0-4 2-4 5z"/>',
    'weighing': '<path d="M12 3v18"/><path d="M7 21h10"/><path d="M3 7h18"/><path d="m6 7-3 7a3 3 0 0 0 6 0z"/><path d="m18 7-3 7a3 3 0 0 0 6 0z"/>',
    'transfer': '<path d="m16 3 4 4-4 4"/><path d="M20 7H4"/><path d="m8 21-4-4 4-4"/><path d="M4 17h16"/>',
    'sale': '<path d="M3 7h11v9H3zM14 10h4l3 3v3h-7"/><circle cx="7" cy="18" r="2"/><circle cx="17" cy="18" r="2"/>',
    'package': '<path d="M12 3 3 7.5v9L12 21l9-4.5v-9zM3 7.5 12 12l9-4.5M12 12v9"/>',
    'coins': '<ellipse cx="9" cy="7" rx="6" ry="3"/><path d="M3 7v4c0 1.7 2.7 3 6 3s6-1.3 6-3V7"/><path d="M9 14v3c0 1.7 2.7 3 6 3s6-1.3 6-3v-4c0-1.7-2.7-3-6-3"/>',
    'chev': '<path d="m9 18 6-6-6-6"/>',
}

def tile(rec, icon):
    return f'<span class="ip-tile ip-rec ip-rec--{rec}" aria-hidden="true">{svg(IC[icon])}</span>'

def sex(s, n=None):
    c = 'hen' if s == 'F' else 'rooster'
    g = '♀' if s == 'F' else '♂'
    return f'<span class="ip-sex ip-sex--{c}">{g}' + (f' <b>{n}</b>' if n else '') + '</span>'

SRSEX = {'F': 'găini', 'M': 'cocoși'}

def cap(text, sub=None):
    s = f'<span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">{text}</span>'
    return s

# ---------------------------------------------------------------- MissingRow
def mrow(entry, section, sexes, date, meta, farm=None, link=True, locked=False, state='', go=True):
    rec, icon, name = ('mortality', 'mortality', 'Zilnic') if entry == 'z' else ('feed', 'wheat', 'Hrană zilnică')
    signs = ''
    if sexes:
        signs = ' ' + ' · '.join(sex(s) for s in sexes) + '<span class="ip-sr"> ' + ' și '.join(SRSEX[s] for s in sexes) + '</span>'
    pre = f'{farm} · ' if farm else ''
    lead = '<span class="ip-sr">Completează: </span>' if link and not locked else ''
    btag = '<span class="ip-tag ip-tag--neutral">blocat</span>' if locked else ''
    title = f'<span class="ip-ev__title">{lead}{pre}{section} · <span class="ip-ev__type">{name}{signs}</span>{btag}</span>'
    body = f'{tile(rec, icon)}<span class="ip-ev__main">{title}<span class="ip-ev__meta ip-num">{date} · SAPT 40 · {meta}</span></span>'
    tag = ''
    lamp = '<span class="ip-ev__status"><span class="ip-lamp ip-lamp--warn">lipsă</span></span>'
    gos = f'<span class="ip-ev__go">{svg(IC["chev"], "ip-icon ip-icon--16") if (link and not locked) else ""}</span>'
    cls = 'ip-ev ip-ev--missing' + (' is-locked' if locked else '') + (f' {state}' if state else '')
    if link and not locked:
        return f'<a class="{cls}" href="#">{body}{tag}{lamp}{gos}</a>'
    return f'<div class="{cls}">{body}{tag}{lamp}{gos}</div>'

def mhead(hid, scope):
    return f'<div class="ip-missing__head"><h2 class="ip-cap" id="{hid}">Lipsesc date</h2><p class="ip-note">{scope}</p></div>'

def missing(sfx):
    today = f'16.11.2026'
    real = [
        mrow('f', 'H3.2', 'M', today, f'{sex("M", "426")} în viață · ieri 157{N}g'),
        mrow('f', 'H4.2', 'F', today, f'{sex("F", f"4{N}107")} în viață · ieri 156{N}g'),
    ]
    scope = 'Hîrbovăț · azi de la <span class="ip-num">12:00</span> · ultimele <span class="ip-num">7</span> zile'
    early = 'Hîrbovăț · <span class="ip-num">10:40</span> · ultimele <span class="ip-num">7</span> zile'
    out = []
    # 1 default (vet, real)
    out.append(f'<section class="ip-missing" aria-labelledby="ms-a{sfx}">{mhead("ms-a"+sfx, scope)}<div class="ip-list">{"".join(real)}</div></section>')
    # 2 state examples (D39)
    ex = [
        mrow('z', 'H1.1', '', today, f'{sex("F", f"4{N}281")} · {sex("M", "429")} în viață'),
        mrow('f', 'H4.2', 'F', '15.11.2026', f'zi trecută · o completează managerul', locked=True),
    ]
    out.append(cap('Exemple de stare (D39): Zilnic lipsă · zi trecută, vet (blocat)') +
               f'<section class="ip-missing" aria-labelledby="ms-b{sfx}">{mhead("ms-b"+sfx, scope)}<div class="ip-list">{"".join(ex)}</div></section>')
    # 3 manager, farm named
    mg = [
        mrow('f', 'H4.2', 'F', '15.11.2026', f'{sex("F", f"4{N}110")} în viață', farm='Hîrbovăț'),
        mrow('f', 'H3.2', 'M', today, f'{sex("M", "426")} în viață · ieri 157{N}g', farm='Hîrbovăț'),
    ]
    out.append(cap('Manager · mai multe ferme: ferma în titlu · zi trecută = link (D07)') +
               f'<section class="ip-missing" aria-labelledby="ms-c{sfx}">{mhead("ms-c"+sfx, "fermele tale · azi de la <span class=\"ip-num\">12:00</span> · ultimele <span class=\"ip-num\">7</span> zile")}<div class="ip-list">{"".join(mg)}</div></section>')
    # 3b before 12:00, past days missing: past-day rows only, today not listed yet
    pe = [mrow('f', 'H4.2', 'F', '15.11.2026', f'{sex("F", f"4{N}110")} în viață')]
    out.append(cap('Înainte de 12:00 cu zile trecute lipsă (manager): doar zilele trecute, azi nu apare încă') +
               f'<section class="ip-missing" aria-labelledby="ms-g{sfx}">{mhead("ms-g"+sfx, early)}<div class="ip-list">{"".join(pe)}</div></section>')
    # 4 interaction
    st = [
        mrow('f', 'H3.2', 'M', today, f'{sex("M", "426")} în viață · ieri 157{N}g', state='is-hover'),
        mrow('f', 'H3.2', 'M', today, f'{sex("M", "426")} în viață · ieri 157{N}g', state='is-pressed'),
        mrow('f', 'H3.2', 'M', today, f'{sex("M", "426")} în viață · ieri 157{N}g', state='ip-focus'),
    ]
    out.append(cap('Stări: hover · apăsat · focus') + f'<div class="ip-list">{"".join(st)}</div>')
    # 5 empty after 12:00, empty before 12:00, loading
    out.append(cap('Nimic lipsă (după 12:00) · înainte de 12:00 · se încarcă') +
        f'<div class="ip-stack" style="gap:var(--space-4)">'
        f'<section class="ip-missing" aria-labelledby="ms-d{sfx}">{mhead("ms-d"+sfx, scope)}<div class="ip-list"><p class="ip-missing__empty"><span class="ip-lamp ip-lamp--off">nimic lipsă</span><span>Toate secțiunile cu păsări au Zilnic și Hrană zilnică în ultimele <span class="ip-num">7</span> zile.</span></p></div></section>'
        f'<section class="ip-missing" aria-labelledby="ms-e{sfx}">{mhead("ms-e"+sfx, early)}<div class="ip-list"><p class="ip-missing__empty"><span class="ip-lamp ip-lamp--off">nimic lipsă</span><span>Zilele <span class="ip-num">09.11.2026–15.11.2026</span> sunt complete. Lipsurile de azi apar de la <span class="ip-num">12:00</span>.</span></p></div></section>'
        f'<section class="ip-missing" aria-labelledby="ms-f{sfx}" aria-busy="true">{mhead("ms-f"+sfx, scope)}<div class="ip-list">'
        + ''.join('<div class="ip-ev ip-ev--skel" aria-hidden="true"><span class="ip-skel ip-skel--tile"></span><span class="ip-ev__main"><span class="ip-skel ip-skel--w1"></span><span class="ip-skel ip-skel--w2"></span></span><span class="ip-skel ip-skel--w3"></span></div>' for _ in range(2))
        + '<span class="ip-sr">Se încarcă</span></div></section></div>')
    return out

def build_missing():
    d = missing('-d'); m = missing('-m')
    desk = ''.join(f'<div>{x}</div>' for x in d)
    phone = ''.join(f'<div>{x}</div>' for x in m)
    html = ('<!-- @dsCard group="Data" height=2050 subtitle="SYS08" -->\n'
        '<div class="ip-root ip-stack" style="gap:var(--space-6)">\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">S2 · Acasă · Lipsesc date · desktop (coloană de 560 px)</span><div class="ip-stack" style="width:560px;gap:var(--space-5)">{desk}</div></div>\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">390 · telefon</span><div class="ip-phone" style="width:390px"><div class="ip-stack" style="padding:var(--space-4);gap:var(--space-5)">{phone}</div></div></div>\n'
        f'<p class="ip-note" style="margin:0">Hîrbovăț 16.11.2026: sample (D27), sample-hirbovat-2026.json › allocations_today (H3.2 ♂ și H4.2 ♀ fără grame; ieri 157{N}g și 156{N}g), daily (în viață). Zilnic H1.1 lipsă, H4.2 ♀ lipsă pe 15.11 și listele „nimic lipsă” sunt exemple de stare (D39): pe 16.11 Zilnic e salvat la 08:34 și 15.11 e complet. Telița e goală din 28.02.2026, deci nu apare (doar secțiuni cu păsări).</p>\n'
        '</div>\n')
    (ROOT / 'MissingRow').mkdir(exist_ok=True)
    (ROOT / 'MissingRow' / 'preview.html').write_text(html, encoding='utf-8')

# ---------------------------------------------------------------- EntryKeys
KEYS = [
    ('mortality', 'mortality', 'Zilnic', 'Mortalitate, sacrificare, ouă'),
    ('feed', 'wheat', 'Hrană zilnică', 'Grame pe găină și cocoș'),
    ('weighing', 'weighing', 'Cântărire', 'Greutate, CV, uniformitate'),
    ('transfer', 'transfer', 'Transfer', 'Între secțiuni sau ferme'),
    ('sale', 'sale', 'Vânzare', 'Păsări vândute pe secțiuni'),
    ('feed', 'package', 'Livrare hrană', 'Kg pe tip de hrană'),
]

def key(k, state='', disabled=False, desc=''):
    rec, icon, label, hint = k
    hint = hint.replace('♀ · ♂', f'{sex("F")} · {sex("M")}')
    inner = f'{tile(rec, icon)}<span class="ip-entrykey__txt"><span>{label}</span><span class="ip-note">{hint}</span></span>'
    cls = 'ip-btn ip-btn--secondary ip-entrykey' + (f' {state}' if state else '')
    if disabled:
        return f'<li><a class="{cls}" role="link" aria-disabled="true"{f' aria-describedby="{desc}"' if desc else ''}>{inner}</a></li>'
    return f'<li><a class="{cls}" href="#">{inner}</a></li>'

def keys_block(hid, ks, note='', disabled=False, states=None):
    states = states or [''] * len(ks)
    lis = ''.join(key(k, s, disabled, f'{hid}-n' if note else '') for k, s in zip(ks, states))
    n = f'<p class="ip-note" id="{hid}-n">{note}</p>' if note else ''
    return f'<section class="ip-entrykeys" aria-labelledby="{hid}"><h2 class="ip-cap" id="{hid}">Înregistrează</h2><ul class="ip-entrykeys__list">{lis}</ul>{n}</section>'

def build_keys():
    tel = 'Telița e goală din <span class="ip-num">28.02.2026</span>. Intrările se deschid cu ciclul nou.'
    d = [
        ('Manager · 6 chei · 1136 px, un rând', keys_block('ek-m-d', KEYS)),
        ('Vet · 3 chei · aceeași lățime de cheie', keys_block('ek-v-d', KEYS[:3])),
        ('Stări: hover · apăsat · focus', keys_block('ek-s-d', KEYS[:3], states=['is-hover', 'is-pressed', 'ip-focus'])),
        ('Dezactivat: vet doar la Telița (fermă goală)', keys_block('ek-x-d', KEYS[:3], note=tel, disabled=True)),
    ]
    m = [
        ('Vet · 3 chei (Cântărire pe tot rândul)', keys_block('ek-v-m', KEYS[:3])),
        ('Manager · 6 chei, 2 coloane', keys_block('ek-m-m', KEYS)),
        ('Dezactivat', keys_block('ek-x-m', KEYS[:3], note=tel, disabled=True)),
    ]
    dd = ''.join(f'<div>{cap(t)}{b}</div>' for t, b in d)
    mm = ''.join(f'<div>{cap(t)}{b}</div>' for t, b in m)
    html = ('<!-- @dsCard group="Actions" height=1250 subtitle="SYS08" -->\n'
        '<div class="ip-root ip-stack" style="gap:var(--space-6)">\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">S2 · Acasă · chei de introducere · 1440 (conținut 1136 px)</span><div class="ip-stack" style="width:1136px;gap:var(--space-5)">{dd}</div></div>\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">390 · telefon</span><div class="ip-phone" style="width:390px"><div class="ip-stack" style="padding:var(--space-4);gap:var(--space-5)">{mm}</div></div></div>\n'
        '<p class="ip-note" style="margin:0">Telița: real, ciclu închis 28.02.2026 (cycles.json). Ordinea cheilor: screens.md S2 (vet: Zilnic, Hrană zilnică, Cântărire; managerul adaugă Transfer, Vânzare, Livrare hrană).</p>\n'
        '</div>\n')
    (ROOT / 'EntryKeys').mkdir(exist_ok=True)
    (ROOT / 'EntryKeys' / 'preview.html').write_text(html, encoding='utf-8')

# ---------------------------------------------------------------- UnpricedCount
def unpriced(hid, kind='default', touch=False, btn_state=''):
    t = ' ip-btn--touch' if touch else ''
    head = f'<div class="ip-unpriced__head">{tile("feed", "coins")}<h2 class="ip-cap" id="{hid}">Livrări fără preț</h2></div>'
    if kind == 'loading':
        return (f'<section class="ip-unpriced" aria-labelledby="{hid}" aria-busy="true">{head}'
                '<div aria-hidden="true" class="ip-unpriced__skel"><span class="ip-skel ip-skel--val"></span><span class="ip-skel ip-skel--w1"></span><span class="ip-skel ip-skel--w2"></span></div>'
                '<span class="ip-sr">Se încarcă</span></section>')
    if kind == 'none':
        return (f'<section class="ip-unpriced" aria-labelledby="{hid}">{head}'
                '<p class="ip-unpriced__fig"><span class="ip-unpriced__val">0<small> livrări</small></span></p>'
                '<p class="ip-note">Toate livrările au preț.</p>'
                f'<a class="ip-btn ip-btn--secondary{t}" href="#">Vezi livrările</a></section>')
    return (f'<section class="ip-unpriced" aria-labelledby="{hid}">{head}'
            f'<p class="ip-unpriced__fig"><span class="ip-unpriced__val">4<small> livrări</small></span><span class="ip-unpriced__kg">76{N}107 kg</span></p>'
            '<ul class="ip-unpriced__list">'
            f'<li><span>Hîrbovăț · Breeder&nbsp;2</span><span><span class="ip-num">3</span> livrări · <span class="ip-num">72{N}068 kg</span></span></li>'
            f'<li><span>Hîrbovăț · Cocoși</span><span><span class="ip-num">1</span> livrare · <span class="ip-num">4{N}039 kg</span></span></li>'
            '</ul>'
            '<p class="ip-note">Cea mai veche din <span class="ip-num">04.11.2026</span>. Kg fără preț nu intră în cost.</p>'
            f'<a class="ip-btn ip-btn--secondary{t}{" " + btn_state if btn_state else ""}" href="#">Adaugă prețurile</a></section>')

def build_unpriced():
    d = [
        ('Implicit · 4 livrări fără preț', unpriced('up-a-d')),
        ('Toate au preț (exemplu de stare, D39)', unpriced('up-b-d', 'none')),
        ('Se încarcă', unpriced('up-c-d', 'loading')),
        ('Cheia: focus', unpriced('up-e-d', btn_state='ip-focus')),
    ]
    m = [
        ('Implicit', unpriced('up-a-m', touch=True)),
        ('Toate au preț', unpriced('up-b-m', 'none', touch=True)),
    ]
    dd = ''.join(f'<div style="width:360px">{cap(t)}{b}</div>' for t, b in d)
    mm = ''.join(f'<div>{cap(t)}{b}</div>' for t, b in m)
    html = ('<!-- @dsCard group="Data" height=900 subtitle="SYS08" -->\n'
        '<div class="ip-root ip-stack" style="gap:var(--space-6)">\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">S2 · Acasă · proprietar · desktop (coloană de 360 px)</span><div class="ip-row" style="align-items:flex-start;gap:var(--space-5)">{dd}</div></div>\n'
        f'<div><span class="ip-cap" style="display:block;margin:0 0 var(--space-2)">390 · telefon</span><div class="ip-phone" style="width:390px"><div class="ip-stack" style="padding:var(--space-4);gap:var(--space-5)">{mm}</div></div></div>\n'
        f'<p class="ip-note" style="margin:0">Sample (D27): sample-hirbovat-2026.json › deliveries: 4 livrări fără preț la Hîrbovăț, 04.11–14.11.2026 (Breeder 2 24{N}068 + 24{N}000 + 24{N}000 kg; Cocoși 4{N}039 kg) = 76{N}107 kg. Toate livrările Telița au preț. „0 livrări” e un exemplu de stare (D39).</p>\n'
        '</div>\n')
    (ROOT / 'UnpricedCount').mkdir(exist_ok=True)
    (ROOT / 'UnpricedCount' / 'preview.html').write_text(html, encoding='utf-8')

build_missing(); build_keys(); build_unpriced()
print('ok')
