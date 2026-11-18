"""Screen CSS shared by the phone state boards (S7S, S6S; S13 gets it through s7s_build.CSS): the
keyless Banner body (s7s-banner; a Banner with an action key is additions.css §49 since FIN03), and the skeleton block (s7s-skblock/skrow; the field and text-line skeletons are additions.css §54 since FIN11G) (the Toast with Anulează and its placement moved to additions.css §48, FIN02:
.ip-toastwrap[--page], .ip-toast__txt, .ip-toast__undo). Class names stay `s7s-`
(the screen and its main are additions.css §57 ip-screen/__main since FIN15; the 390×844 viewport, its view and the
sticky SaveBar are §63 ip-screen--fixed/__view/__savebar since FIN15B). Strings are concatenated in this order by s7s_build (its boards are byte-identical to before).
MAIN ('' since FX03: every phone form main keeps ip-screen__main's space-6, additions.css §57). SIDE / M0 (both '' since FIN15/FIN15B, additions.css §58/§62) / REC / EFFECT: the desktop entry form's record icon
and EffectPreview line-up, shared by s13_build and s9_build. Class names stay `s13-` (as S6 reuses s7s-); s13_build
concatenates them in its own order (its boards are byte-identical to before).
FIN03: the Banner-action layout moved to additions.css §49 (.ip-banner--act, __body, __key)."""

VIEW = ''  # FIN15B: the 390x844 viewport and its clipped view are additions.css §63 ip-screen--fixed, ip-screen__view(--end)

BANNER = '''
.s7s-banner .ip-icon{flex:none;margin-top:var(--ip-hair)}
.s7s-banner__body{display:flex;flex-direction:column;align-items:flex-start;gap:var(--space-2)}'''

SKEL = '''
.s7s-skblock{display:flex;flex-direction:column;gap:var(--space-6)}
.s7s-skrow{align-items:start}'''


# the page beside the panel: board height 1024 (desktop board h; the board root's min-height gives it, FIN15B)
SIDE_H = 1024
SIDE = ''  # FIN15B: the page beside the docked panel (ip-dockpage) and the form body (ip-sidepanel__body--form) are additions.css §62

MAIN = ''  # FX03: the phone form main gap is ip-screen__main's space-6 (additions.css §57), as S6-S8; was .s13-main{gap:space-5}

M0 = ''  # FIN15: .ip-m0 and the docked SidePanel (ip-sidepanel--dock, __title, __keys) are additions.css §58

OPT = '''
.s13-opt{font-weight:400;color:var(--ink-muted)}'''  # "(opțional)" after a field label (craft §5), shared by s13_build and s10_build

REC = '''
.s13-rec{color:var(--c);display:inline-flex;flex:none}'''

EFFECT = '''
.s13-effect{align-items:flex-start}
.s13-effect > .s13-rec{margin-top:var(--ip-hair)}'''

# EffectPreview body: one line per scope (name left, before → after right), cycle-mark lines with a sub-line; and the phone
# screen as the anchor of a modal over it (BottomSheet scrim, Dialog scrim). Shared by s9_build and s10_build; class names
# stay `s9-` (s9_build concatenates them where its own rules stood, so its boards are byte-identical to before).
FX = '''
.s9-fxbody{display:flex;flex-direction:column;gap:var(--space-1);min-width:0;flex:1}
.s9-fxline{display:flex;flex-wrap:wrap;justify-content:space-between;gap:0 var(--space-2)}
.s9-fxval{display:inline-flex;align-items:baseline;gap:var(--space-1);white-space:nowrap}
.s9-fxval .ip-pair{gap:var(--space-1)}
.s9-to{color:var(--ink-secondary)}
.s9-strong{font-weight:500;color:var(--ink)}
.s9-fxcyc{display:flex;flex-wrap:wrap;align-items:center;gap:0 var(--space-2);margin-top:var(--space-1)}
.s9-fxsub{color:var(--ink-secondary);font-size:var(--ip-fs-small);line-height:var(--ip-lh-small)}'''

MODAL = ''  # FIN15: the phone screen as a modal's anchor is additions.css §60 .ip-scrimhost
