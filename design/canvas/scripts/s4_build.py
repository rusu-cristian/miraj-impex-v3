"""Build S4 · Fermă · Rezumat (the default board) from realdata.
python3 -I s4_build.py <repo_root> [<height>] [--canvas]
Since S4P the default board is drawn by the same functions as its 8 state boards (s4s_build.py), so the shell, the figures
and the formats can't drift between them. S4N (D154): the Rezumat tab opens on its PeriodTable ("Total fermă", SAPT 56–59,
period 4 săpt., D159), drawn by s4t_build.py (tb_rezumat / main_rezumat, SYS14 builders), then the S4S panel (Readout, Stats,
Secțiuni, De verificat, Ultimele înregistrări). Writes design/canvas/project/S4-ferma-rezumat.dc.html and
/tmp/s4s/S4-ferma-rezumat.html (plain wrapper for s4s_render.py)."""
import sys, pathlib, importlib.util

A = [a for a in sys.argv[1:] if a != '--canvas']  # --canvas: refresh the entry's h/title in place (x, y kept; RF01)
H = int(A[1]) if len(A) > 1 else 2944  # S4NF (D166): every week open, 2272 → 2944
spec = importlib.util.spec_from_file_location('s4t_build', pathlib.Path(__file__).resolve().with_name('s4t_build.py'))
T = importlib.util.module_from_spec(spec); spec.loader.exec_module(T)  # loads s4s_build (reads sys.argv[1], the repo root)
s4s = T.s4

s4s.write('S4-ferma-rezumat', 'S4 · Fermă · Rezumat', T.main_rezumat(), H=H, css=T.CSS)
if '--canvas' in sys.argv[2:]:
    import canvas_rows as CR
    c = CR.load(s4s.ROOT); CR.place(c, 'S4-ferma-rezumat', [('S4-ferma-rezumat', 1440, H, 'S4 · Fermă · Rezumat')]); CR.save(s4s.ROOT, c)
