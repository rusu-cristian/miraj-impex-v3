# Toast

A save confirmation in `control` (graphite), with an optional undo. Not used for errors.

Copy (FX04, D143): object first, then the verb, then the figures: "Zilnic salvat · …", "Hrană zilnică salvată", "Cântărire salvată", "Transfer salvat", "Vânzare înregistrată", "Livrare adăugată", "Corecție salvată", "Mortalitate ștearsă · …" (never "Salvat · …", "Șters · …"); the Setări toasts already read this way ("Contul lui Petrea e dezactivat."). The preview's save Toast is the S6 Zilnic one.

Placement and the Toast with Anulează: `components.md` › Additions › FIN02 (additions.css §48): phone `ip-toastwrap` inset `space-4` above the SaveBar or BottomTabBar; desktop `ip-toastwrap--page` bottom right at `space-8`, at most `panel-w` wide; `ip-toast__txt` + `ip-toast__undo` (Anulează) on every save Toast.
