package entries_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// T5 criterion 5 on the dev seed: Dr. Ceban's 18.11.2026 Zilnic mortality
// entry is soft-deleted with its rows and the same values are saved again by
// Dr. Ceban: the §6.2 Toast "Zilnic salvat · 4 secțiuni · ♀ −3 · ♂ 0" and
// the farm's birds at the end of today ♀ 12 561 · ♂ 794 (golden 5).
func TestCreateDailySeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	farm := seededFarm(t, d.App, "hirbovat")
	vet := seededUser(t, d.App, "dr.ceban")

	var old uuid.UUID
	if err := d.App.QueryRow(ctx, `select distinct e.id from entry e join bird_event b on b.entry_id = e.id
		where e.farm_id = $1 and e.kind = 'daily' and e.date = $2 and e.created_by = $3 and e.deleted_at is null
		  and b.type = 'mortality' and b.deleted_at is null`, farm, testdb.Today, vet.UserID).Scan(&old); err != nil {
		t.Fatal(err)
	}
	rs, err := d.App.Query(ctx, `select b.section_id, b.sex::text, -b.qty from bird_event b join section s on s.id = b.section_id
		where b.entry_id = $1 and b.deleted_at is null and b.type = 'mortality' order by s.code, b.sex`, old)
	if err != nil {
		t.Fatal(err)
	}
	var sections []entries.DailySection
	for rs.Next() {
		var (
			id   uuid.UUID
			sex  string
			dead int64
		)
		if err := rs.Scan(&id, &sex, &dead); err != nil {
			t.Fatal(err)
		}
		if len(sections) == 0 || sections[len(sections)-1].SectionID != id {
			sections = append(sections, entries.DailySection{SectionID: id})
		}
		v := num.FromInt(dead)
		s := &sections[len(sections)-1]
		if sex == "F" {
			s.F = &entries.DailySex{Mortality: &v}
		} else {
			s.M = &entries.DailySex{Mortality: &v}
		}
	}
	if err := rs.Err(); err != nil {
		t.Fatal(err)
	}
	if len(sections) != 4 {
		t.Fatalf("seeded 18.11 mortality covers %d sections, want 4", len(sections))
	}
	if _, err := d.App.Exec(ctx, `update bird_event set deleted_at = $2 where entry_id = $1`, old, testdb.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := d.App.Exec(ctx, `update entry set deleted_at = $2, deleted_by = $3 where id = $1`, old, testdb.Now(), vet.UserID); err != nil {
		t.Fatal(err)
	}

	svc := entries.NewService(db.NewStore(d.App), d.Clock)
	key := uuid.New()
	req := entries.DailyCreateRequest{
		WriteMeta:   httpx.WriteMeta{ClientRequestID: &key},
		DailyValues: entries.DailyValues{FarmID: farm, Date: testdb.Today, Sections: sections, Eggs: []entries.DailyEggs{}},
	}
	refs, err := svc.DailyRefs(ctx, farm)
	if err != nil {
		t.Fatal(err)
	}
	var v httpx.Validator
	if req.Validate(&v, refs); v.Err() != nil {
		t.Fatal(v.Err())
	}
	res, err := svc.CreateDaily(ctx, vet, req)
	if err != nil {
		t.Fatal(err)
	}
	var resp entries.EntryWriteResponse
	if err := json.Unmarshal(res.Body, &resp); err != nil {
		t.Fatal(err)
	}
	if res.Status != http.StatusCreated || resp.Toast != "Zilnic salvat · 4 secțiuni · ♀ −3 · ♂ 0" ||
		resp.Alive == nil || *resp.Alive != (httpx.SexPair[int]{F: 12561, M: 794}) {
		t.Errorf("got %d %s", res.Status, res.Body)
	}
	t.Logf("warnings: %+v", resp.Warnings)
}
