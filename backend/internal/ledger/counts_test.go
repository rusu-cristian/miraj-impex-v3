package ledger

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func jan(d int) timex.Day { return timex.MustDay(2026, time.January, d) }

func ip(v int) *int { return &v }

// countsFixture is a hand-built farm with four bird-holding sections. today is
// 20.01.2026 (passed explicitly; the pool's app.today is not read).
//
//	S1 ♀: +10 (01.01), mortality −12 (05.01) → −2, adjustment +5 (06.01) → 3: negative on 05.01 only.
//	S1 ♂: +5 (01.01), sale −8 (10.01) → −3 until today.
//	S2 ♀: +5 (01.01), mortality −7 (02.01) → −2, +10 (03.01) → 8, mortality −1 (15.01) → 7: negative before 03.01 only.
//	S2 ♂: +5 (01.01); a soft-deleted mortality −100 (04.01) and a mortality −100 (05.01) in a deleted entry.
//	S3 ♀: +5 (01.01), −10 (02.01) → −5, +10 (04.01) → 5, −8 (08.01) → −3, +3 (09.01) → 0: two negative stretches.
type countsFixture struct {
	t   *testing.T
	d   *testdb.DB
	sec map[string]uuid.UUID
}

func (f *countsFixture) id(sql string, args ...any) uuid.UUID {
	f.t.Helper()
	var id uuid.UUID
	if err := f.d.App.QueryRow(context.Background(), sql+" returning id", args...).Scan(&id); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
	return id
}

func newCountsFixture(t *testing.T) *countsFixture {
	t.Helper()
	f := &countsFixture{t: t, d: testdb.New(t), sec: map[string]uuid.UUID{}}
	farm := f.id(`insert into farm (name, name_key, type) values ('Ferma L', 'ferma l', 'laying')`)
	user := f.id(`insert into app_user (name, login, password_hash, role) values ('Test', 'test.ledger', 'x', 'owner')`)
	h := f.id(`insert into section (farm_id, code) values ($1, 'H1')`, farm)
	for _, code := range []string{"S1", "S2", "S3"} {
		f.sec[code] = f.id(`insert into section (farm_id, code, parent_id) values ($1, $2, $3)`, farm, code, h)
	}
	e := f.id(`insert into entry (kind, farm_id, date, source) values ('migration', $1, $2, 'migration')`, farm, jan(1))
	dead := f.id(`insert into entry (kind, farm_id, date, deleted_at, deleted_by) values ('daily', $1, $2, $3, $4)`,
		farm, jan(5), testdb.Now(), user)
	bird := func(entry uuid.UUID, code string, d int, sex, typ string, qty int, deleted bool) {
		t.Helper()
		var at *time.Time
		if deleted {
			now := testdb.Now()
			at = &now
		}
		f.id(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, deleted_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8)`, entry, farm, f.sec[code], jan(d), sex, typ, qty, at)
	}
	bird(e, "S1", 1, "F", "placement", 10, false)
	bird(e, "S1", 5, "F", "mortality", -12, false)
	bird(e, "S1", 6, "F", "adjustment", 5, false)
	bird(e, "S1", 1, "M", "placement", 5, false)
	bird(e, "S1", 10, "M", "sale", -8, false)
	bird(e, "S2", 1, "F", "placement", 5, false)
	bird(e, "S2", 2, "F", "mortality", -7, false)
	bird(e, "S2", 3, "F", "adjustment", 10, false)
	bird(e, "S2", 15, "F", "mortality", -1, false)
	bird(e, "S2", 1, "M", "placement", 5, false)
	bird(e, "S2", 4, "M", "mortality", -100, true)
	bird(dead, "S2", 5, "M", "mortality", -100, false)
	bird(e, "S3", 1, "F", "placement", 5, false)
	bird(e, "S3", 2, "F", "adjustment", -10, false)
	bird(e, "S3", 4, "F", "adjustment", 10, false)
	bird(e, "S3", 8, "F", "adjustment", -8, false)
	bird(e, "S3", 9, "F", "adjustment", 3, false)
	return f
}

func (f *countsFixture) pair(code string, sex dbq.Sex) SectionSex {
	return SectionSex{SectionID: f.sec[code], Sex: sex}
}

func (f *countsFixture) neg(code string, sex dbq.Sex, first, firstAlive, last, min int, today *int) Negative {
	return Negative{
		SectionID: f.sec[code], Code: code, Sex: sex,
		FirstDate: jan(first), FirstAliveEnd: firstAlive, LastDate: jan(last), MinAliveEnd: min, TodayAliveEnd: today,
	}
}

func TestCheckCounts(t *testing.T) {
	f := newCountsFixture(t)
	F, M := dbq.SexF, dbq.SexM
	today := jan(20)
	tests := []struct {
		name     string
		affected []SectionSex
		from     timex.Day
		want     []Negative
	}{
		{"one day negative, then recovers (first = last)",
			[]SectionSex{f.pair("S1", F)}, jan(5),
			[]Negative{f.neg("S1", F, 5, -2, 5, -2, nil)}},
		{"negative until today",
			[]SectionSex{f.pair("S1", M)}, jan(10),
			[]Negative{f.neg("S1", M, 10, -3, 20, -3, ip(-3))}},
		{"from inside a negative stretch without an event that day",
			[]SectionSex{f.pair("S1", M)}, jan(12),
			[]Negative{f.neg("S1", M, 12, -3, 20, -3, ip(-3))}},
		{"two pairs in affected order, a duplicate counted once",
			[]SectionSex{f.pair("S1", M), f.pair("S1", F), f.pair("S1", M)}, jan(1),
			[]Negative{f.neg("S1", M, 10, -3, 20, -3, ip(-3)), f.neg("S1", F, 5, -2, 5, -2, nil)}},
		{"a pair negative only before from",
			[]SectionSex{f.pair("S2", F)}, jan(3), []Negative{}},
		{"the same pair from its first event",
			[]SectionSex{f.pair("S2", F)}, jan(1),
			[]Negative{f.neg("S2", F, 2, -2, 2, -2, nil)}},
		{"soft-deleted rows and rows of deleted entries ignored",
			[]SectionSex{f.pair("S2", M)}, jan(1), []Negative{}},
		{"from before the first event",
			[]SectionSex{f.pair("S1", F)}, timex.MustDay(2025, time.December, 1),
			[]Negative{f.neg("S1", F, 5, -2, 5, -2, nil)}},
		{"two negative stretches: the first, the last and the lowest",
			[]SectionSex{f.pair("S3", F)}, jan(1),
			[]Negative{f.neg("S3", F, 2, -5, 8, -5, nil)}},
		{"from in the gap between the stretches",
			[]SectionSex{f.pair("S3", F)}, jan(5),
			[]Negative{f.neg("S3", F, 8, -3, 8, -3, nil)}},
		{"from after the last negative day",
			[]SectionSex{f.pair("S1", F), f.pair("S3", F)}, jan(9), []Negative{}},
		{"a pair without events", []SectionSex{f.pair("S3", M)}, jan(1), []Negative{}},
		{"nothing affected", nil, jan(1), []Negative{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CheckCounts(t.Context(), f.d.App, tc.affected, tc.from, today)
			if err != nil {
				t.Fatal(err)
			}
			if got == nil {
				t.Fatal("CheckCounts returned a nil slice")
			}
			if diff := cmp.Diff(tc.want, got, cmp.AllowUnexported(timex.Day{})); diff != "" {
				t.Errorf("CheckCounts (-want +got):\n%s", diff)
			}
		})
	}
}

// TestCheckCountsInTx: inside a transaction CheckCounts sees the transaction's
// own uncommitted change (it runs after the write, before the commit).
func TestCheckCountsInTx(t *testing.T) {
	f := newCountsFixture(t)
	ctx := t.Context()
	tx, err := f.d.App.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `update bird_event set qty = -20 where section_id = $1 and type = 'mortality' and date = $2`,
		f.sec["S2"], jan(15)); err != nil {
		t.Fatal(err)
	}
	got, err := CheckCounts(ctx, tx, []SectionSex{f.pair("S2", dbq.SexF)}, jan(15), jan(20))
	if err != nil {
		t.Fatal(err)
	}
	want := []Negative{f.neg("S2", dbq.SexF, 15, -12, 20, -12, ip(-12))}
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(timex.Day{})); diff != "" {
		t.Errorf("CheckCounts (-want +got):\n%s", diff)
	}
	if got[0].RemovalMax(20) != 8 {
		t.Errorf("RemovalMax(20) = %d, want 8", got[0].RemovalMax(20))
	}
}

// TestCheckCountsError: a failing read is returned wrapped, never as "no
// negative day".
func TestCheckCountsError(t *testing.T) {
	f := newCountsFixture(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	got, err := CheckCounts(ctx, f.d.App, []SectionSex{f.pair("S1", dbq.SexF)}, jan(1), jan(20))
	if err == nil || got != nil {
		t.Fatalf("CheckCounts on a cancelled context = %v, %v; want nil and an error", got, err)
	}
}

// TestBounds: the §6.1 step 3 examples.
func TestBounds(t *testing.T) {
	if got := RemovalMax(4292, -8); got != 4284 {
		t.Errorf("RemovalMax(4 292, −8) = %d, want 4 284", got)
	}
	if got := InflowMin(4381, -2); got != 4383 {
		t.Errorf("InflowMin(4 381, −2) = %d, want 4 383", got)
	}
	n := Negative{MinAliveEnd: -8}
	if got := n.RemovalMax(4292); got != 4284 {
		t.Errorf("Negative.RemovalMax(4 292) = %d, want 4 284", got)
	}
	if got := (Negative{MinAliveEnd: -2}).InflowMin(4381); got != 4383 {
		t.Errorf("Negative.InflowMin(4 381) = %d, want 4 383", got)
	}
	if got := RemovalMax(3, -8); got != -5 {
		t.Errorf("RemovalMax(3, −8) = %d, want −5 (the field alone can't fix it)", got)
	}
}
