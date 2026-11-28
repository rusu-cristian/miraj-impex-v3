package entries_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// seededUser is the dev seed's user with login as a principal.
func seededUser(t *testing.T, d dbq.DBTX, login string) auth.Principal {
	t.Helper()
	var (
		p    auth.Principal
		role string
	)
	err := d.QueryRow(t.Context(), `select u.id, u.name, u.role,
		coalesce(array(select uf.farm_id from user_farm uf where uf.user_id = u.id), '{}')
		from app_user u where u.login = $1`, login).Scan(&p.UserID, &p.Name, &role, &p.FarmIDs)
	if err != nil {
		t.Fatalf("user %q: %v", login, err)
	}
	p.Role = auth.Role(role)
	return p
}

// T3 criterion 5 on the dev seed: Dr. Ceban's Zilnic mortality of
// 17.11.2026, corrected by Ion Rusu on 18.11 at 10:12 (H5.2 ♀ 1 → 3).
func TestReadEntrySeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	svc := entries.NewService(db.NewStore(d.App), d.Clock)
	var id uuid.UUID
	if err := d.App.QueryRow(ctx, `select id from entry where kind = 'daily' and date = '2026-11-17' and corrected`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	hirbovat, telita := seededFarm(t, d.App, "hirbovat"), seededFarm(t, d.App, "telita")
	owner := seededUser(t, d.App, "proprietar")
	vet := seededUser(t, d.App, "dr.ceban")
	viewer := seededUser(t, d.App, "observator")

	res, err := svc.ReadEntry(ctx, owner, id)
	if err != nil {
		t.Fatal(err)
	}
	h := res.Entry
	if h.Kind != dbq.EntryKindDaily || h.Farm.ID != hirbovat || h.Farm.Name != "Hîrbovăț" || h.Date.String() != "2026-11-17" ||
		h.Author == nil || h.Author.Name != "Dr. Ceban" || h.Corrected == nil || h.Corrected.By != "Ion Rusu" ||
		h.Corrected.At.String() != "2026-11-18T10:12:00+02:00" || h.SAPT == nil || *h.SAPT != 59 || h.AgeDays == nil || *h.AgeDays != 406 {
		t.Errorf("head = %+v (author %+v, corrected %+v)", h, h.Author, h.Corrected)
	}
	if res.Actions != (entries.EntryReadActions{Correct: true, Delete: true}) {
		t.Errorf("owner actions = %+v", res.Actions)
	}

	// The values: 4 sections, H5.2 ♀ 3 and H6.1 ♀ 4 dead.
	values, ok := res.Values.(entries.DailyValues)
	if !ok {
		t.Fatalf("values = %T", res.Values)
	}
	codes := map[uuid.UUID]string{}
	for _, s := range res.Sections {
		codes[s.ID] = s.Code
	}
	dead := map[string]int64{}
	for _, s := range values.Sections {
		if s.F != nil && s.F.Mortality != nil {
			dead[codes[s.SectionID]+" F"] = s.F.Mortality.Decimal().IntPart()
		}
		if s.M != nil && s.M.Mortality != nil {
			dead[codes[s.SectionID]+" M"] = s.M.Mortality.Decimal().IntPart()
		}
	}
	if len(values.Sections) != 4 || len(values.Eggs) != 0 || dead["H5.2 F"] != 3 || dead["H6.1 F"] != 4 || values.FarmID != hirbovat {
		t.Errorf("values: %d sections, %d eggs, deaths %v", len(values.Sections), len(values.Eggs), dead)
	}
	// Sent back unchanged, values decode as the create shape (no read-only field).
	b, err := json.Marshal(res.Values)
	if err != nil {
		t.Fatal(err)
	}
	var back entries.DailyValues
	if err := json.Unmarshal(b, &back); err != nil || len(back.Sections) != 4 {
		t.Errorf("round trip: %v", err)
	}

	for _, p := range []auth.Principal{vet, viewer} {
		res, err := svc.ReadEntry(ctx, p, id)
		if err != nil {
			t.Fatalf("%s: %v", p.Name, err)
		}
		if res.Actions != (entries.EntryReadActions{}) {
			t.Errorf("%s: actions %+v", p.Name, res.Actions)
		}
	}

	// Ion Rusu's correction of today is the entry's latest action, but its
	// audit row is imported without an old_values snapshot: POST …/undo
	// answers 409 STALE, so the read offers no undo either (criterion 11).
	ion := seededUser(t, d.App, "ion.rusu")
	res, err = svc.ReadEntry(ctx, ion, id)
	if err != nil {
		t.Fatal(err)
	}
	if res.Actions != (entries.EntryReadActions{Correct: true, Delete: true}) {
		t.Errorf("Ion Rusu: actions %+v, want no undo", res.Actions)
	}
	_, err = svc.UndoEntry(ctx, ion, id)
	wantHTTP(t, err, http.StatusConflict, httpx.CodeStale)

	telitaManager := auth.Principal{UserID: uuid.New(), Name: "Vasile", Role: auth.Manager, FarmIDs: []uuid.UUID{telita}}
	_, err = svc.ReadEntry(ctx, telitaManager, id)
	wantHTTP(t, err, http.StatusForbidden, httpx.CodeFarmForbidden)
	_, err = svc.ReadEntry(ctx, owner, uuid.New())
	wantHTTP(t, err, http.StatusNotFound, httpx.CodeNotFound)
	var migration uuid.UUID
	if err := d.App.QueryRow(ctx, `select id from entry where kind = 'migration' and farm_id = $1`, hirbovat).Scan(&migration); err != nil {
		t.Fatal(err)
	}
	_, err = svc.ReadEntry(ctx, owner, migration)
	wantHTTP(t, err, http.StatusNotFound, httpx.CodeNotFound)
	if _, err := d.App.Exec(ctx, `update entry set deleted_at = $2, deleted_by = $3 where id = $1`, id, testdb.Now(), owner.UserID); err != nil {
		t.Fatal(err)
	}
	_, err = svc.ReadEntry(ctx, owner, id)
	wantHTTP(t, err, http.StatusNotFound, httpx.CodeNotFound)
}

func wantHTTP(t *testing.T, err error, status int, code string) {
	t.Helper()
	var he httpx.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("err = %v, want %d %s", err, status, code)
	}
	if st, body := he.HTTPError(); st != status || body.Code != code {
		t.Fatalf("err = %d %s, want %d %s", st, body.Code, status, code)
	}
}
