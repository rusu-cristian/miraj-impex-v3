package entries_test

import (
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// T6 on the dev seed (golden 37, S16 boards, §11.8 "Rules"): the owner puts
// H5.2 ♀ of Dr. Ceban's corrected 17.11.2026 Zilnic back to 1; Ion Rusu's
// dry run 1 → 3 answers the §7.8 preview (H5.2 ♀ 4 285 → 4 283 on 17.11,
// 4 283 → 4 281 today, the farm ♀ 12 563 → 12 561, golden 33's warning
// only); the save's Toast names H5.2 ♀ 4 281; typing 4 292 is refused with
// the bound 4 284.
func TestUpdateDailySeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	svc := entries.NewService(db.NewStore(d.App), d.Clock)
	ion := seededUser(t, d.App, "ion.rusu")
	var (
		id    uuid.UUID
		owner auth.Principal
	)
	if err := d.App.QueryRow(ctx, `select id from entry where kind = 'daily' and date = '2026-11-17' and corrected`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if err := d.App.QueryRow(ctx, `select id, name from app_user where role = 'owner' order by created_at limit 1`).
		Scan(&owner.UserID, &owner.Name); err != nil {
		t.Fatal(err)
	}
	owner.Role = auth.Owner
	put := func(p auth.Principal, mortality int64, reason string, dryRun bool) (entries.Result, error) {
		t.Helper()
		var farm uuid.UUID
		var at time.Time
		if err := d.App.QueryRow(ctx, `select farm_id, coalesce(updated_at, created_at) from entry where id = $1`, id).Scan(&farm, &at); err != nil {
			t.Fatal(err)
		}
		s, err := entries.LoadStored(ctx, d.App, id, dbq.EntryKindDaily, farm, testdb.Today.AddDays(-1))
		if err != nil {
			t.Fatal(err)
		}
		v := s.Daily()
		h52 := slices.IndexFunc(s.Sections, func(x entries.EntrySection) bool { return x.Code == "H5.2" })
		for i, sec := range v.Sections {
			if sec.SectionID == s.Sections[h52].ID {
				n := num.FromInt(mortality)
				f := *sec.F
				f.Mortality = &n
				v.Sections[i].F = &f
			}
		}
		return svc.UpdateDaily(ctx, p, id, entries.DailyUpdateRequest{
			IfVersion: at.UTC().Format("2006-01-02T15:04:05.000000Z07:00"), Values: v, Reason: &reason, DryRun: dryRun,
		})
	}

	if _, err := put(owner, 1, "Înapoi la registru.", false); err != nil {
		t.Fatal(err)
	}
	res, err := put(ion, 3, "", true)
	p := decodeResult[entries.CorrectionPreview](t, res, err)
	wantEffects := []string{
		"H5.2 · găini în viață de pe 17.11.2026 se schimbă cu −2.", "Mortalitate · morți ♀ −5 → −7",
		"Hîrbovăț · ♀ azi 12 563 → 12 561",
	}
	wantAlive := []entries.PreviewAlive{
		{Section: "H5.2", Sex: dbq.SexF, Date: testdb.Today.AddDays(-1), Before: 4285, After: 4283},
		{Section: "H5.2", Sex: dbq.SexF, Date: testdb.Today, Before: 4283, After: 4281},
	}
	if len(p.Changes) != 1 || p.Changes[0].Key != "H5.2 · găini moarte" || !slices.Equal(p.Effects, wantEffects) ||
		!slices.Equal(p.Alive, wantAlive) || p.CheckedDays != 2 || len(p.Refusals) != 0 || len(p.Warnings) != 1 ||
		p.Warnings[0].Message != "H6.1: 4 găini moarte, de peste 3 ori media pe 7 zile (0,9). Verifică numărătoarea; poți salva." {
		t.Errorf("preview %s", res.Body)
	}

	res, err = put(ion, 3, "Numărătoare refăcută: 2 găini găsite moarte sub hrănitoare.", false)
	w := decodeResult[entries.EntryWriteResponse](t, res, err)
	if w.Toast != "Corecție salvată · H5.2 ♀ 1 → 3 · în viață ♀ 4 281" {
		t.Errorf("save %s", res.Body)
	}

	_, err = put(ion, 4292, "Test.", false)
	he, ok := err.(httpx.HTTPError)
	if !ok {
		t.Fatalf("err = %v", err)
	}
	status, body := he.HTTPError()
	if status != http.StatusUnprocessableEntity || len(body.Refusals) != 1 ||
		body.Message != "Corecția ar lăsa H5.2 cu −6 găini pe 17.11.2026 și cu −8 găini azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0." ||
		body.Refusals[0].Message != "H5.2 ♀ 4 292: ar lăsa −8 găini azi, 18.11.2026. Scrie cel mult 4 284." ||
		*body.Refusals[0].Max != 4284 || *body.Refusals[0].MinAliveEnd != -8 || *body.Refusals[0].FirstNegativeDate != testdb.Today.AddDays(-1) {
		t.Errorf("refusal %d %+v", status, body)
	}
}
