package entries_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func decodeResult[T any](t *testing.T, res entries.Result, err error) T {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	var v T
	if err := json.Unmarshal(res.Body, &v); err != nil {
		t.Fatalf("%d %s: %v", res.Status, res.Body, err)
	}
	return v
}

// T7 criterion 2 on the dev seed (S16 boards): Ion Rusu deletes Dr. Ceban's
// corrected Zilnic of 17.11.2026 with a Motiv (♀ −7 · ♂ 0, farm ♀ today
// 12 561 → 12 568), then restores it.
func TestDeleteSeededWholeSave(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	svc := entries.NewService(db.NewStore(d.App), d.Clock)
	ion := seededUser(t, d.App, "ion.rusu")
	var id uuid.UUID
	if err := d.App.QueryRow(ctx, `select id from entry where kind = 'daily' and date = '2026-11-17' and corrected`).Scan(&id); err != nil {
		t.Fatal(err)
	}

	res, err := svc.DeleteEntry(ctx, ion, id, entries.DeleteRequest{DryRun: true})
	p := decodeResult[entries.CorrectionPreview](t, res, err)
	want := "Se șterge salvarea Zilnic introdusă de Dr. Ceban la 08:07 și corectată de Ion Rusu pe 18.11.2026: 4 secțiuni · " +
		"♀ −7 · ♂ 0. Găini în viață azi în Hîrbovăț: ♀ 12 561 → ♀ 12 568. Ștergerea rămâne în Istoric modificări."
	if len(p.Effects) != 1 || p.Effects[0] != want || p.CheckedDays != 2 || len(p.Refusals) != 0 {
		t.Errorf("preview %s", res.Body)
	}

	reason := "Salvare dublă."
	res, err = svc.DeleteEntry(ctx, ion, id, entries.DeleteRequest{Reason: &reason})
	w := decodeResult[entries.EntryWriteResponse](t, res, err)
	if w.Toast != "Mortalitate ștearsă · 17.11.2026 · 4 secțiuni · în viață azi ♀ 12 568" ||
		w.Alive == nil || *w.Alive != (httpx.SexPair[int]{F: 12568, M: 794}) {
		t.Errorf("delete %s", res.Body)
	}

	res, err = svc.RestoreEntry(ctx, ion, id)
	w = decodeResult[entries.EntryWriteResponse](t, res, err)
	if w.Toast != "Mortalitate restaurată · 17.11.2026 · 4 secțiuni · în viață azi ♀ 12 561" ||
		*w.Alive != (httpx.SexPair[int]{F: 12561, M: 794}) {
		t.Errorf("restore %s", res.Body)
	}
}

// T7 criterion 2 on the dev seed (S15 boards): Dr. Ceban deletes H5.2 of his
// Zilnic of 18.11 (saved at 08:15, ♀ −2), H5.2 ♀ 4 281 → 4 283; the restore
// brings it back to 4 281 with an audit `restore` and entry.restored.
func TestDeleteSeededSection(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	svc := entries.NewService(db.NewStore(d.App), d.Clock)
	vet := seededUser(t, d.App, "dr.ceban")
	hirbovat := seededFarm(t, d.App, "hirbovat")
	var id, h52 uuid.UUID
	if err := d.App.QueryRow(ctx, `select e.id from entry e where e.kind = 'daily' and e.date = '2026-11-18'
		and exists (select 1 from bird_event b where b.entry_id = e.id and b.deleted_at is null)`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if err := d.App.QueryRow(ctx, `select id from section where farm_id = $1 and code = 'H5.2' and retired_at is null`,
		hirbovat).Scan(&h52); err != nil {
		t.Fatal(err)
	}
	h52Alive := func() int {
		t.Helper()
		rows, err := calc.SectionsAliveEnd(ctx, d.App, hirbovat, testdb.Today)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range rows {
			if r.SectionID == h52 {
				return r.AliveEnd.F
			}
		}
		t.Fatal("no H5.2")
		return 0
	}
	if got := h52Alive(); got != 4281 {
		t.Fatalf("H5.2 ♀ before = %d", got)
	}

	res, err := svc.DeleteEntry(ctx, vet, id, entries.DeleteRequest{SectionID: &h52, DryRun: true})
	p := decodeResult[entries.CorrectionPreview](t, res, err)
	want := "Se șterge rândul salvat la 08:15: ♀ −2 · ♂ 0. În viață în H5.2: ♀ 4 281 → 4 283. Ștergerea rămâne în " +
		"Istoric modificări."
	if len(p.Effects) != 1 || p.Effects[0] != want {
		t.Errorf("preview %s", res.Body)
	}
	if want := []entries.PreviewAlive{{Section: "H5.2", Sex: "F", Date: testdb.Today, Before: 4281, After: 4283}}; !slices.Equal(p.Alive, want) {
		t.Errorf("alive %+v", p.Alive)
	}

	res, err = svc.DeleteEntry(ctx, vet, id, entries.DeleteRequest{SectionID: &h52})
	w := decodeResult[entries.EntryWriteResponse](t, res, err)
	if w.Toast != "Mortalitate ștearsă · H5.2 · în viață ♀ 4 283" || h52Alive() != 4283 {
		t.Errorf("delete %s", res.Body)
	}

	res, err = svc.RestoreEntry(ctx, vet, id)
	w = decodeResult[entries.EntryWriteResponse](t, res, err)
	if w.Toast != "Mortalitate restaurată · H5.2 · în viață ♀ 4 281" || h52Alive() != 4281 {
		t.Errorf("restore %s", res.Body)
	}
	var action, liveType string
	if err := d.App.QueryRow(ctx, `select (select action::text from audit_log where entry_id = $1 order by at desc, id desc limit 1),
		(select type from live_event where type <> 'counts.changed' order by id desc limit 1)`, id).Scan(&action, &liveType); err != nil {
		t.Fatal(err)
	}
	if action != "restore" || liveType != "entry.restored" {
		t.Errorf("audit %s, live %s", action, liveType)
	}
}
