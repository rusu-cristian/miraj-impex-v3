package reports_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/farms"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/reports"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb/seed"
)

// TestMain serves the whole test binary: the in-package tests take
// testdb.New, the seeded ones here testdb.NewSeeded (seed imports farms and
// golden, so only an external test package may import it).
func TestMain(m *testing.M) { seed.Main(m) }

func farmID(t *testing.T, db dbq.DBTX, key string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := db.QueryRow(t.Context(), `select id from farm where name_key = $1`, key).Scan(&id); err != nil {
		t.Fatalf("farm %q: %v", key, err)
	}
	return id
}

// row is a Stoc row with its numbers at the precision the test states.
type row struct {
	Farm, Type, Kg, DaysLeft, AvgDaily, RunOut, Lamp, Reason, Sub string
	Window                                                        [2]string
}

func rows(t *testing.T, res reports.FeedResponse) []row {
	t.Helper()
	out := make([]row, 0, len(res.Stock))
	fixed := func(n *num.Num, places int32) string {
		if n == nil {
			return "null"
		}
		return n.Decimal().StringFixed(places)
	}
	for _, r := range res.Stock {
		switch r := r.(type) {
		case reports.TypeStock:
			x := row{
				Farm: r.Farm, Type: r.Type, Kg: fixed(&r.Kg, 1), DaysLeft: fixed(r.DaysLeft, 3),
				AvgDaily: fixed(&r.AvgDaily, 2), Lamp: string(r.Lamp), Reason: r.Reason,
				Window: [2]string{r.Window[0].String(), r.Window[1].String()},
			}
			if r.RunOut != nil {
				x.RunOut = r.RunOut.String()
			}
			out = append(out, x)
		case reports.FarmStock:
			out = append(out, row{Farm: r.Farm, Lamp: string(r.Lamp), Reason: r.Reason, Sub: r.Sub})
		default:
			t.Fatalf("stock row %T", r)
		}
	}
	return out
}

var (
	window = [2]string{"2026-11-16", "2026-11-18"}
	// Goldens 27 and 28 (§5.8): 31 664,9 ÷ 2 392,2 = 13,237; 1 781,0 ÷ 170,93 = 10,419.
	breeder2 = row{Farm: "Hîrbovăț", Type: "Breeder 2", Kg: "31664.9", DaysLeft: "13.237", AvgDaily: "2392.20",
		RunOut: "2026-12-01", Lamp: "in_norma", Window: window}
	cocosi = row{Farm: "Hîrbovăț", Type: "Cocoși", Kg: "1781.0", DaysLeft: "10.419", AvgDaily: "170.93",
		RunOut: "2026-11-28", Lamp: "in_norma", Window: window}
	telita = row{Farm: "Telița", Lamp: "oprit", Reason: "fără ciclu deschis", Sub: "ultimul ciclu: Seria 2026, închis pe 28.02.2026"}
)

// TestBuildFeedSeeded: the Stoc report on the dev seed as of 18.11.2026
// (§7.12, M3 T6): Hîrbovăț's two feed types, then Telița's single "fără
// ciclu deschis" row; a viewer of Telița only sees Telița.
func TestBuildFeedSeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	telitaID := farmID(t, d.App, "telita")
	for _, tc := range []struct {
		name string
		p    auth.Principal
		want []row
	}{
		{"owner, every farm", auth.Principal{UserID: uuid.New(), Role: auth.Owner}, []row{breeder2, cocosi, telita}},
		{"viewer of Telița", auth.Principal{UserID: uuid.New(), Role: auth.Viewer, FarmIDs: []uuid.UUID{telitaID}}, []row{telita}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res, err := reports.BuildFeed(t.Context(), d.App, reports.Input{Principal: tc.p, Today: testdb.Today, Now: testdb.Now()})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, rows(t, res)); diff != "" {
				t.Errorf("stock (-want +got):\n%s", diff)
			}
		})
	}
}

// TestFeedMatchesSummary: the report and the summary's feedDaysLeft give the
// same figures for Hîrbovăț (M3 D12), both from calc.ReadStockDays.
func TestFeedMatchesSummary(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	hirbovat := farmID(t, d.App, "hirbovat")
	owner := auth.Principal{UserID: uuid.New(), Role: auth.Owner}
	report, err := reports.BuildFeed(ctx, d.App, reports.Input{Principal: owner, FarmID: &hirbovat, Today: testdb.Today, Now: testdb.Now()})
	if err != nil {
		t.Fatal(err)
	}
	summary, err := farms.BuildSummary(ctx, d.App, farms.Input{Principal: owner, FarmID: hirbovat, Today: testdb.Today, Now: testdb.Now()})
	if err != nil {
		t.Fatal(err)
	}

	// Both encoded, keeping the fields they share.
	type shared struct {
		Type   string   `json:"type"`
		ForSex *string  `json:"forSex"`
		Kg     num.Num  `json:"kg"`
		Days   *num.Num `json:"days"`
		RunOut *string  `json:"runOut"`
		Lamp   string   `json:"lamp"`
		Reason string   `json:"reason"`
	}
	var fromReport, fromSummary []shared
	for _, r := range report.Stock {
		ts, ok := r.(reports.TypeStock)
		if !ok {
			t.Fatalf("row %T on a farm with an open cycle", r)
		}
		var runOut *string
		if ts.RunOut != nil {
			s := ts.RunOut.String()
			runOut = &s
		}
		fromReport = append(fromReport, shared{ts.Type, ts.ForSex, ts.Kg, ts.DaysLeft, runOut, string(ts.Lamp), ts.Reason})
	}
	for _, st := range summary.Stats {
		if st, ok := st.(farms.FeedDaysLeftStat); ok {
			b, err := json.Marshal(st.Items)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(b, &fromSummary); err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(fromReport) != 2 {
		t.Fatalf("report rows = %d, want Breeder 2 and Cocoși", len(fromReport))
	}
	encode := func(v any) string {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	if got, want := encode(fromSummary), encode(fromReport); got != want {
		t.Errorf("summary feedDaysLeft = %s,\nreport stock = %s", got, want)
	}
}
