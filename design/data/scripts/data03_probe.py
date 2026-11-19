"""DATA03 probe: print a bounded window of a sheet (for building data03_build.py).
Usage: python3 -I data03_probe.py <xlsx> <sheet> <row_from> <row_to> [col_from] [col_to]
Prints only non-empty cells, values truncated to 30 chars."""
import sys, datetime
import openpyxl

def main():
    path, sheet = sys.argv[1], sys.argv[2]
    r0, r1 = int(sys.argv[3]), int(sys.argv[4])
    c0 = int(sys.argv[5]) if len(sys.argv) > 5 else 1
    c1 = int(sys.argv[6]) if len(sys.argv) > 6 else 60
    wb = openpyxl.load_workbook(path, read_only=True, data_only=True)
    if sheet == '?':
        for ws in wb.worksheets:
            print(ws.title, ws.max_row, ws.max_column)
        return
    ws = wb[sheet]
    for i, row in enumerate(ws.iter_rows(min_row=r0, max_row=r1, min_col=c0, max_col=c1, values_only=True), start=r0):
        cells = []
        for j, v in enumerate(row, start=c0):
            if v is None or v == '':
                continue
            if isinstance(v, datetime.datetime):
                v = v.date().isoformat()
            s = str(v)
            if isinstance(v, float):
                s = f'{v:.4g}'
            cells.append(f'{openpyxl.utils.get_column_letter(j)}={s[:30]}')
        if cells:
            print(i, ' '.join(cells))

main()
