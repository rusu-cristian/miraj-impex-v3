package entries_test

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/golden"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// The seeded correction chain (plan M5 T10, golden 37, S16 boards, §11.8
// "Rules"), through entries.Service as the handlers call it: the owner puts
// H5.2 ♀ of Dr. Ceban's corrected 17.11.2026 Zilnic back to 1, then Ion Rusu
// corrects it 1 → 3 again: the dry run, the save, and a typed 4 292 that
// §6.1 refuses. Each step reads the entry as the client does (GET, then PUT
// its values with its version).

// correctionChain is the seeded database, the service and the 17.11 save.
type correctionChain struct {
	d          *testdb.DB
	svc        *entries.Service
	entry      uuid.UUID
	farm, h52  uuid.UUID
	owner, ion auth.Principal
}

var day17 = testdb.Today.AddDays(-1)

func newCorrectionChain(t *testing.T) correctionChain {
	t.Helper()
	d := testdb.NewSeeded(t)
	c := correctionChain{
		d: d, svc: entries.NewService(db.NewStore(d.App), d.Clock), farm: seededFarm(t, d.App, "hirbovat"),
		owner: seededUser(t, d.App, "proprietar"), ion: seededUser(t, d.App, "ion.rusu"),
	}
	if err := d.App.QueryRow(t.Context(), `select id from entry where kind = 'daily' and date = $1 and corrected
		and deleted_at is null`, day17).Scan(&c.entry); err != nil {
		t.Fatal(err)
	}
	if err := d.App.QueryRow(t.Context(), `select id from section where farm_id = $1 and code = 'H5.2'`, c.farm).Scan(&c.h52); err != nil {
		t.Fatal(err)
	}
	return c
}

// put reads the entry as p and sends its values back with H5.2 ♀ mortality
// set to mortality (PUT /api/entries/{id} with ifVersion = the read version).
func (c correctionChain) put(t *testing.T, p auth.Principal, mortality int64, reason string, dryRun bool) (entries.Result, error) {
	t.Helper()
	read, err := c.svc.ReadEntry(t.Context(), p, c.entry)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := read.Values.(entries.DailyValues)
	if !ok {
		t.Fatalf("entry values %T, want entries.DailyValues", read.Values)
	}
	v.Sections = slices.Clone(v.Sections)
	i := slices.IndexFunc(v.Sections, func(s entries.DailySection) bool { return s.SectionID == c.h52 })
	if i < 0 || v.Sections[i].F == nil {
		t.Fatalf("the 17.11 save has no H5.2 ♀ values: %+v", v.Sections)
	}
	f := *v.Sections[i].F
	n := num.FromInt(mortality)
	f.Mortality = &n
	v.Sections[i].F = &f
	return c.svc.UpdateDaily(t.Context(), p, c.entry, entries.DailyUpdateRequest{
		IfVersion: read.Entry.Version, Values: v, Reason: &reason, DryRun: dryRun,
	})
}

// h52F is H5.2 ♀ at the end of 17.11 and of 18.11.
func (c correctionChain) h52F(t *testing.T) [2]int {
	t.Helper()
	var out [2]int
	for i, day := range [2]timex.Day{day17, testdb.Today} {
		secs, err := calc.SectionsAliveEnd(t.Context(), c.d.App, c.farm, day)
		if err != nil {
			t.Fatal(err)
		}
		j := slices.IndexFunc(secs, func(s calc.SectionAlive) bool { return s.SectionID == c.h52 })
		if j < 0 {
			t.Fatalf("no H5.2 on %s", day)
		}
		out[i] = secs[j].AliveEnd.F
	}
	return out
}

// auditRow is the entry's latest audit row.
type auditRow struct {
	Action        dbq.AuditAction
	UserID        uuid.UUID
	Reason        *string
	SameDayAuthor bool
	Changes       json.RawMessage
}

func (c correctionChain) latestAudit(t *testing.T) (auditRow, int) {
	t.Helper()
	var (
		a auditRow
		n int
	)
	if err := c.d.App.QueryRow(t.Context(), `select action, user_id, reason, same_day_author, changes,
		(select count(*) from audit_log where entry_id = $1)
		from audit_log where entry_id = $1 order by at desc, id desc limit 1`, c.entry).
		Scan(&a.Action, &a.UserID, &a.Reason, &a.SameDayAuthor, &a.Changes, &n); err != nil {
		t.Fatal(err)
	}
	return a, n
}

// wantChanges is the one change of the 1 → 3 correction (§4.10, Q12).
const wantChanges = `[{"key":"H5.2 · găini moarte","section":"H5.2","sex":"F","field":"mortality","old":1,"new":3}]`

func TestCorrectionChainSeeded(t *testing.T) {
	c := newCorrectionChain(t)
	if got := c.h52F(t); got != [2]int{4283, 4281} {
		t.Fatalf("seed: H5.2 ♀ ends 17.11 / 18.11 = %v, want 4 283 / 4 281", got)
	}

	// The owner puts the illustrative old value back: H5.2 ♀ 3 → 1.
	if _, err := c.put(t, c.owner, 1, "Înapoi la valoarea inițială.", false); err != nil {
		t.Fatal(err)
	}
	if got := c.h52F(t); got != [2]int{4285, 4283} {
		t.Fatalf("after 3 → 1: H5.2 ♀ ends 17.11 / 18.11 = %v, want 4 285 / 4 283", got)
	}
	_, auditsBefore := c.latestAudit(t)

	t.Run("Ion Rusu's dry run 1 → 3 previews the correction and writes nothing", func(t *testing.T) {
		res, err := c.put(t, c.ion, 3, "", true)
		p := decodeResult[entries.CorrectionPreview](t, res, err)
		if res.Status != http.StatusOK {
			t.Errorf("status %d, want 200", res.Status)
		}
		changes, err := json.Marshal(p.Changes)
		if err != nil {
			t.Fatal(err)
		}
		if string(changes) != wantChanges {
			t.Errorf("changes %s, want %s", changes, wantChanges)
		}
		wantEffects := []string{
			"H5.2 · găini în viață de pe 17.11.2026 se schimbă cu −2.",
			"Mortalitate · morți ♀ −5 → −7",
			"Hîrbovăț · ♀ azi 12\u202f563 → 12\u202f561",
		}
		if diff := cmp.Diff(wantEffects, p.Effects); diff != "" {
			t.Errorf("effects (-want +got):\n%s", diff)
		}
		wantAlive := []entries.PreviewAlive{
			{Section: "H5.2", Sex: dbq.SexF, Date: day17, Before: 4285, After: 4283},
			{Section: "H5.2", Sex: dbq.SexF, Date: testdb.Today, Before: 4283, After: 4281},
		}
		if !slices.Equal(wantAlive, p.Alive) {
			t.Errorf("alive rows %+v, want %+v", p.Alive, wantAlive)
		}
		if p.CheckedDays != 2 || len(p.Refusals) != 0 {
			t.Errorf("checkedDays %d, refusals %+v; want 2 days and no refusal", p.CheckedDays, p.Refusals)
		}
		wantWarnings := []httpx.Warning{{Code: httpx.WarnMortalitySpike, SectionCode: "H6.1", Sex: "F",
			Message: "H6.1: 4 găini moarte, de peste 3 ori media pe 7 zile (0,9). Verifică numărătoarea; poți salva."}}
		if diff := cmp.Diff(wantWarnings, p.Warnings); diff != "" {
			t.Errorf("warnings (-want +got; golden 33's only):\n%s", diff)
		}
		if got := c.h52F(t); got != [2]int{4285, 4283} {
			t.Errorf("after the dry run H5.2 ♀ ends = %v, want 4 285 / 4 283 unchanged", got)
		}
		if _, n := c.latestAudit(t); n != auditsBefore {
			t.Errorf("the dry run wrote %d audit rows", n-auditsBefore)
		}
	})

	t.Run("typing 4 292 is refused with the bound 4 284", func(t *testing.T) {
		_, err := c.put(t, c.ion, 4292, "Test.", false)
		he, ok := err.(httpx.HTTPError)
		if !ok {
			t.Fatalf("err = %v, want a 422", err)
		}
		status, body := he.HTTPError()
		if status != http.StatusUnprocessableEntity || body.Code != "REFUSED" || len(body.Refusals) != 1 {
			t.Fatalf("got %d %+v, want 422 REFUSED with one refusal", status, body)
		}
		if want := "Corecția ar lăsa H5.2 cu −6 găini pe 17.11.2026 și cu −8 găini azi, 18.11.2026. " +
			"Nicio secțiune nu poate scădea sub 0."; body.Message != want {
			t.Errorf("banner %q, want %q", body.Message, want)
		}
		r := body.Refusals[0]
		if want := "H5.2 ♀ 4\u202f292: ar lăsa −8 găini azi, 18.11.2026. Scrie cel mult 4\u202f284."; r.Message != want {
			t.Errorf("field message %q, want %q", r.Message, want)
		}
		if r.Max == nil || *r.Max != 4284 || r.FirstNegativeDate == nil || *r.FirstNegativeDate != day17 ||
			r.MinAliveEnd == nil || *r.MinAliveEnd != -8 || r.SectionCode != "H5.2" || r.Sex != "F" {
			t.Errorf("refusal %+v, want H5.2 ♀ max 4 284, firstNegativeDate 2026-11-17, minAliveEnd −8", r)
		}
		if got := c.h52F(t); got != [2]int{4285, 4283} {
			t.Errorf("after the refusal H5.2 ♀ ends = %v, want 4 285 / 4 283 unchanged", got)
		}
	})

	t.Run("the save 1 → 3 leaves 4 283 / 4 281 and an update audit row with the Motiv", func(t *testing.T) {
		reason := "Numărătoare refăcută: 2 găini găsite moarte sub hrănitoare."
		res, err := c.put(t, c.ion, 3, reason, false)
		w := decodeResult[entries.EntryWriteResponse](t, res, err)
		if res.Status != http.StatusOK { // a change answers 200 (Runner.Change)
			t.Errorf("status %d, want 200", res.Status)
		}
		if want := "Corecție salvată · H5.2 ♀ 1 → 3 · în viață ♀ 4\u202f281"; w.Toast != want {
			t.Errorf("toast %q, want %q", w.Toast, want)
		}
		if got := c.h52F(t); got != [2]int{4283, 4281} {
			t.Errorf("H5.2 ♀ ends 17.11 / 18.11 = %v, want 4 283 / 4 281", got)
		}
		a, n := c.latestAudit(t)
		if n != auditsBefore+1 || a.Action != dbq.AuditActionUpdate || a.UserID != c.ion.UserID || a.SameDayAuthor ||
			a.Reason == nil || *a.Reason != reason {
			t.Errorf("audit row %+v (%d rows, want %d): want Ion Rusu's update with the Motiv", a, n, auditsBefore+1)
		}
		var got, want any
		if err := json.Unmarshal(a.Changes, &got); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(wantChanges), &want); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("audit changes (-want +got):\n%s", diff)
		}
	})

	// The chain leaves the state goldens 33 and 37 describe, written through
	// the API this time: both still hold.
	for _, id := range []int{33, 37} {
		r := golden.Rows()[id-1]
		if err := r.Check(t.Context(), golden.Env{DB: c.d.App}); err != nil {
			t.Errorf("after the chain: %v", err)
		}
	}
}
