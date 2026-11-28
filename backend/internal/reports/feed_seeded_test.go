package reports_test

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/reports"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// seeded is the dev seed as of 18.11.2026 with its two farms.
type seeded struct {
	d                  *testdb.DB
	hirbovat, telita   uuid.UUID
	owner, manager     auth.Principal
	viewer, managerHrb auth.Principal
}

func newSeeded(t *testing.T) seeded {
	t.Helper()
	d := testdb.NewSeeded(t)
	s := seeded{d: d, hirbovat: farmID(t, d.App, "hirbovat"), telita: farmID(t, d.App, "telita")}
	both := []uuid.UUID{s.hirbovat, s.telita}
	s.owner = auth.Principal{UserID: uuid.New(), Role: auth.Owner}
	s.manager = auth.Principal{UserID: uuid.New(), Role: auth.Manager, FarmIDs: both}
	s.viewer = auth.Principal{UserID: uuid.New(), Role: auth.Viewer, FarmIDs: both}
	s.managerHrb = auth.Principal{UserID: uuid.New(), Role: auth.Manager, FarmIDs: []uuid.UUID{s.hirbovat}}
	return s
}

func (s seeded) input(p auth.Principal) reports.Input {
	return reports.Input{Principal: p, Today: testdb.Today, Now: testdb.Now()}
}

// rowLine is a Livrări row at the precision the test states.
type rowLine struct {
	Date, Farm, Type, Kg, PerKg, Total, PricedBy, Tag string
	Set, Change                                       bool
}

func line(r reports.DeliveryRowPriced) rowLine {
	opt := func(n *num.Num) string {
		if n == nil {
			return "null"
		}
		return n.Decimal().String()
	}
	l := rowLine{
		Date: r.Date.String(), Farm: r.Farm, Type: r.Type, Kg: r.Kg.Decimal().String(), PerKg: opt(r.PricePerKg),
		Total: opt(r.TotalLei), Tag: r.Tag, Set: r.Actions.SetPrice, Change: r.Actions.ChangePrice, PricedBy: "null",
	}
	if r.PricedBy != nil {
		l.PricedBy = *r.PricedBy
	}
	return l
}

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func encode(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// The 12 newest deliveries of Seria 2026 (S11 board "cele mai noi 12 din 82"), as the owner sees them.
var ownerFirstPage = []rowLine{
	{"2026-11-13", "Hîrbovăț", "Breeder 2", "25640", "7.7613", "199000", "din registru", "", false, true},
	{"2026-11-12", "Hîrbovăț", "Breeder 2", "9000", "null", "null", "null", "fără preț", true, false},
	{"2026-11-06", "Hîrbovăț", "Breeder 2", "17020", "7.7615", "132100", "din registru", "", false, true},
	{"2026-10-30", "Hîrbovăț", "Breeder 2", "17440", "7.7408", "135000", "din registru", "", false, true},
	{"2026-10-25", "Hîrbovăț", "Breeder 2", "7140", "7.7401", "55264", "din registru", "", false, true},
	{"2026-10-24", "Hîrbovăț", "Breeder 2", "18600", "7.74", "143964", "din registru", "", false, true},
	{"2026-10-23", "Hîrbovăț", "Cocoși", "7000", "7.8", "54600", "din registru", "", false, true},
	{"2026-10-19", "Hîrbovăț", "Breeder 2", "25280", "7.73", "195415", "din registru", "", false, true},
	{"2026-10-10", "Hîrbovăț", "Breeder 2", "24680", "7.73", "190777", "din registru", "", false, true},
	{"2026-10-06", "Hîrbovăț", "Breeder 2", "25160", "7.85", "197506", "din registru", "", false, true},
	{"2026-10-02", "Hîrbovăț", "Breeder 2", "17160", "7.9855", "137031.91", "din registru", "", false, true},
	{"2026-10-02", "Hîrbovăț", "Cocoși", "6980", "7.9855", "55739.09", "din registru", "", false, true},
}

// TestFeedDeliveriesSeeded: Livrări for the owner, every farm, period=cycle
// (§7.12, M4 Q6–Q8): 82 deliveries of Seria 2026 (Hîrbovăț 69 · Telița 13),
// 4 unpriced, the 12 newest rows and a next cursor.
func TestFeedDeliveriesSeeded(t *testing.T) {
	s := newSeeded(t)
	in := s.input(s.owner)
	in.Period = calc.Period{Kind: calc.PeriodCycle}
	res, err := reports.BuildFeedPriced(t.Context(), s.d.App, in)
	if err != nil {
		t.Fatal(err)
	}
	d := res.Deliveries
	if d.Total != 82 {
		t.Errorf("total = %d, want 82", d.Total)
	}
	if want := []reports.FarmCount{{Farm: "Hîrbovăț", Count: 69}, {Farm: "Telița", Count: 13}}; !cmp.Equal(d.ByFarm, want) {
		t.Errorf("byFarm = %+v, want %+v", d.ByFarm, want)
	}
	if got := encode(t, d.Unpriced); got != `{"count":4,"kg":18160,"dates":["2026-02-17","2026-03-14","2026-11-12"]}` {
		t.Errorf("unpriced = %s", got)
	}
	got := make([]rowLine, 0, len(d.Rows))
	for _, r := range d.Rows {
		got = append(got, line(r))
	}
	if diff := cmp.Diff(ownerFirstPage, got); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
	if d.NextCursor == nil {
		t.Error("nextCursor = null, want a cursor")
	}
	if d.LastBefore != nil {
		t.Errorf("lastBefore = %+v, want null with deliveries in the range", d.LastBefore)
	}
	if d.Rows[0].PricedAt != nil {
		t.Errorf("a register price has pricedAt %v, want null", d.Rows[0].PricedAt)
	}
}

// TestFeedPagesSeeded: the cursor pages (M4 Q7) walk every one of the 82
// deliveries once, newest first, and the last page has nextCursor null; the
// viewer's pages hold the same rows.
func TestFeedPagesSeeded(t *testing.T) {
	s := newSeeded(t)
	ctx := t.Context()
	first, err := reports.BuildFeedPriced(ctx, s.d.App, s.input(s.owner))
	if err != nil {
		t.Fatal(err)
	}
	ids := []uuid.UUID{}
	for _, r := range first.Deliveries.Rows {
		ids = append(ids, r.ID)
	}
	next, pages := first.Deliveries.NextCursor, 1
	for next != nil {
		c, err := reports.ParseCursor(*next)
		if err != nil {
			t.Fatalf("page %d cursor: %v", pages, err)
		}
		in := s.input(s.owner)
		in.Cursor = &c
		pg, err := reports.BuildFeedPagePriced(ctx, s.d.App, in)
		if err != nil {
			t.Fatal(err)
		}
		if pages == 1 {
			if b := encode(t, pg); strings.Contains(b, `"stock"`) || strings.Contains(b, `"cost"`) || strings.Contains(b, `"total"`) {
				t.Errorf("a cursor page carries more than its rows: %.200s", b)
			}
			if len(pg.Deliveries.Rows) != 12 {
				t.Errorf("page 2 has %d rows, want 12", len(pg.Deliveries.Rows))
			}
		}
		vin := s.input(s.viewer)
		vin.Cursor = &c
		view, err := reports.BuildFeedPage(ctx, s.d.App, vin)
		if err != nil {
			t.Fatal(err)
		}
		if len(view.Deliveries.Rows) != len(pg.Deliveries.Rows) {
			t.Fatalf("viewer page %d has %d rows, owner %d", pages+1, len(view.Deliveries.Rows), len(pg.Deliveries.Rows))
		}
		for i, r := range pg.Deliveries.Rows {
			if encode(t, view.Deliveries.Rows[i]) != encode(t, r.DeliveryRow) {
				t.Errorf("page %d row %d: viewer %+v, owner %+v", pages+1, i, view.Deliveries.Rows[i], r.DeliveryRow)
			}
			ids = append(ids, r.ID)
		}
		next = pg.Deliveries.NextCursor
		if (view.Deliveries.NextCursor == nil) != (next == nil) {
			t.Errorf("page %d: viewer and owner disagree on the next cursor", pages+1)
		}
		pages++
	}
	if len(ids) != 82 || pages != 7 {
		t.Errorf("%d rows over %d pages, want 82 over 7", len(ids), pages)
	}
	seen := map[uuid.UUID]bool{}
	for _, id := range ids {
		if seen[id] {
			t.Errorf("row %s on two pages", id)
		}
		seen[id] = true
	}
}

// TestFeedCostSeeded: Cost for the owner and a manager of both farms
// (golden 31, §5.9) at full precision, and the same figures from BuildCost.
func TestFeedCostSeeded(t *testing.T) {
	s := newSeeded(t)
	ctx := t.Context()
	div := func(lei string, n int64) string { return num.Of(dec(lei).Div(decimal.NewFromInt(n))).String() }
	// Golden 31: Telița 2 031 768 ÷ 31 590; Hîrbovăț 9 562 437 ÷ 29 907 and ÷ 5 073 155 eggs; the series
	// 11 594 205 ÷ 31 590 and ÷ 5 073 155.
	want := `{"asOf":"2026-11-18","cycles":[` +
		`{"label":"Telița · Seria 2026","span":"07.10.2025–28.02.2026","pricedCount":13,"kg":263740,"lei":2031768,` +
		`"birds":{"F":28080,"M":3510},"eggs":null,"leiPerBird":64.3168,"leiPerEgg":null,"eggReason":"creștere, fără ouă","unpriced":null},` +
		`{"label":"Hîrbovăț · Seria 2026","span":"din 22.02.2026 · SAPT 59","pricedCount":65,"kg":1174480,"lei":9562437,` +
		`"birds":{"F":26883,"M":3024},"eggs":5073155,"leiPerBird":319.7391,"leiPerEgg":1.8849,` +
		`"unpriced":{"count":4,"kg":18160,"span":"17.02–12.11.2026"}}],` +
		`"total":{"label":"Total Seria 2026","kg":1438220,"lei":11594205,"birds":{"F":28080,"M":3510},"eggs":5073155,` +
		`"leiPerBird":367.0214,"leiPerEgg":2.2854,"excludedKg":18160}}`
	for _, p := range []auth.Principal{s.owner, s.manager} {
		res, err := reports.BuildFeedPriced(ctx, s.d.App, s.input(p))
		if err != nil {
			t.Fatal(err)
		}
		if got := encode(t, res.Cost); got != want {
			t.Errorf("%s cost\n got %s\nwant %s", p.Role, got, want)
		}
		cost, err := reports.BuildCost(ctx, s.d.App, s.input(p))
		if err != nil {
			t.Fatal(err)
		}
		if got := encode(t, cost); got != want {
			t.Errorf("%s BuildCost\n got %s", p.Role, got)
		}
		// Full precision: the wire rounds to 4 decimals, the values are the exact quotients.
		h := cost.Cycles[1]
		if got, w := h.LeiPerBird.Decimal(), dec("9562437").Div(decimal.NewFromInt(29907)); !got.Equal(w) {
			t.Errorf("Hîrbovăț lei/pasăre = %s, want %s", got, w)
		}
		if got := h.LeiPerBird.Decimal().StringFixed(2) + " " + h.LeiPerEgg.Decimal().StringFixed(2) + " " +
			cost.Cycles[0].LeiPerBird.Decimal().StringFixed(2) + " " + cost.Total.LeiPerBird.Decimal().StringFixed(2) + " " +
			cost.Total.LeiPerEgg.Decimal().StringFixed(2); got != "319.74 1.88 64.32 367.02 2.29" {
			t.Errorf("golden 31 at 2 decimals = %s", got)
		}
		if got := num.Of(cost.Total.LeiPerEgg.Decimal()).String(); got != div("11594205", 5073155) {
			t.Errorf("total lei/ou = %s", got)
		}
	}
}

// TestFeedManagerSeeded: a manager of both farms sees the owner's rows and
// prices with every action false; a manager of Hîrbovăț only gets its cost
// row and no total (Telița, the series' other farm, is not his, M4 Q6).
func TestFeedManagerSeeded(t *testing.T) {
	s := newSeeded(t)
	ctx := t.Context()
	res, err := reports.BuildFeedPriced(ctx, s.d.App, s.input(s.manager))
	if err != nil {
		t.Fatal(err)
	}
	want := slices.Clone(ownerFirstPage)
	for i := range want {
		want[i].Set, want[i].Change = false, false
	}
	got := make([]rowLine, 0, len(res.Deliveries.Rows))
	for _, r := range res.Deliveries.Rows {
		got = append(got, line(r))
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("manager rows (-want +got):\n%s", diff)
	}

	one, err := reports.BuildFeedPriced(ctx, s.d.App, s.input(s.managerHrb))
	if err != nil {
		t.Fatal(err)
	}
	if len(one.Cost.Cycles) != 1 || one.Cost.Cycles[0].Label != "Hîrbovăț · Seria 2026" || one.Cost.Total != nil {
		t.Errorf("manager of Hîrbovăț: cost = %s, want the Hîrbovăț row and no total", encode(t, one.Cost))
	}
	if b := encode(t, one.Cost); strings.Contains(b, `"total"`) {
		t.Errorf("cost has a total key: %s", b)
	}
	if one.Deliveries.Total != 69 {
		t.Errorf("manager of Hîrbovăț: total = %d, want 69", one.Deliveries.Total)
	}
}

// priceKeys are the keys a viewer's JSON never holds (D2, D64, D200).
func isPriceKey(k string) bool {
	return k == "pricePerKg" || k == "totalLei" || k == "cost" || k == "pricedBy" || k == "pricedAt" ||
		strings.HasPrefix(k, "lei") || strings.HasPrefix(k, "price")
}

// walkNoPrices fails on any price key or "fără preț" value in v (a decoded
// JSON value) and on any null array.
func walkNoPrices(t *testing.T, path string, v any) {
	t.Helper()
	switch v := v.(type) {
	case map[string]any:
		for k, x := range v {
			if isPriceKey(k) {
				t.Errorf("%s.%s: a price key in the viewer's JSON", path, k)
			}
			walkNoPrices(t, path+"."+k, x)
		}
	case []any:
		for i, x := range v {
			walkNoPrices(t, path+"["+strconv.Itoa(i)+"]", x)
		}
	case string:
		if strings.Contains(v, "fără preț") {
			t.Errorf("%s = %q: a price tag in the viewer's JSON", path, v)
		}
	}
}

// TestFeedViewerSeeded: the viewer gets Stoc and Livrări without a single
// price key, the unpriced count only and no Cost (§7.0a, D64, D200).
func TestFeedViewerSeeded(t *testing.T) {
	s := newSeeded(t)
	ctx := t.Context()
	res, err := reports.BuildFeed(ctx, s.d.App, s.input(s.viewer))
	if err != nil {
		t.Fatal(err)
	}
	b := encode(t, res)
	var doc any
	if err := json.Unmarshal([]byte(b), &doc); err != nil {
		t.Fatal(err)
	}
	walkNoPrices(t, "$", doc)
	if got := encode(t, res.Deliveries.Unpriced); got != `{"count":4}` {
		t.Errorf("unpriced = %s, want the count only", got)
	}
	if got := encode(t, res.Deliveries.Rows[1]); !strings.HasPrefix(got, `{"id":"`) ||
		!strings.HasSuffix(got, `"date":"2026-11-12","farm":"Hîrbovăț","type":"Breeder 2","forSex":"F","kg":9000}`) {
		t.Errorf("row = %s", got)
	}
	if res.Deliveries.Total != 82 || len(res.Deliveries.Rows) != 12 {
		t.Errorf("total %d, rows %d; want 82, 12", res.Deliveries.Total, len(res.Deliveries.Rows))
	}
	c, err := reports.ParseCursor(*res.Deliveries.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	in := s.input(s.viewer)
	in.Cursor = &c
	pg, err := reports.BuildFeedPage(ctx, s.d.App, in)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(encode(t, pg)), &doc); err != nil {
		t.Fatal(err)
	}
	walkNoPrices(t, "$page", doc)

	for name, build := range map[string]func() error{
		"BuildFeedPriced":     func() error { _, err := reports.BuildFeedPriced(ctx, s.d.App, s.input(s.viewer)); return err },
		"BuildFeedPagePriced": func() error { _, err := reports.BuildFeedPagePriced(ctx, s.d.App, in); return err },
		"BuildCost":           func() error { _, err := reports.BuildCost(ctx, s.d.App, s.input(s.viewer)); return err },
	} {
		if err := build(); !errors.Is(err, reports.ErrPricesHidden) {
			t.Errorf("%s for a viewer: err = %v, want ErrPricesHidden", name, err)
		}
	}
}

// TestFeedEmptyIntervalSeeded: Telița from 01.03 to 18.11.2026 has no
// delivery (S11 "fără livrări în perioadă"): rows [], the last delivery
// before the interval (10.02.2026 · 21 640 kg Pre-ouat) and no cost row.
func TestFeedEmptyIntervalSeeded(t *testing.T) {
	s := newSeeded(t)
	in := s.input(s.owner)
	in.FarmID = &s.telita
	in.Period = calc.Period{Kind: calc.PeriodInterval, From: timex.MustDay(2026, time.March, 1), To: testdb.Today}
	res, err := reports.BuildFeedPriced(t.Context(), s.d.App, in)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"total":0,"byFarm":[],"unpriced":{"count":0,"kg":0,"dates":[]},"rows":[],"nextCursor":null,` +
		`"lastBefore":{"date":"2026-02-10","farm":"Telița","type":"Pre-ouat","kg":21640}}`
	if got := encode(t, res.Deliveries); got != want {
		t.Errorf("deliveries\n got %s\nwant %s", got, want)
	}
	if got := encode(t, res.Cost); got != `{"asOf":"2026-11-18","cycles":[]}` {
		t.Errorf("cost = %s", got)
	}
}

// TestCostInputsMatchCalc (M4 ARC-1): the batched cost bases
// (ReportsCostInputs) equal calc.FarmBase over the cycle, calc.EggsBetween
// over the range and calc.SeriesPlacement, the readers they batch, on the
// seed and on a hand-built series whose rows the seed lacks.
func TestCostInputsMatchCalc(t *testing.T) {
	t.Run("seed", func(t *testing.T) {
		s := newSeeded(t)
		var series uuid.UUID
		if err := s.d.App.QueryRow(t.Context(), `select id from series where name = 'Seria 2026'`).Scan(&series); err != nil {
			t.Fatal(err)
		}
		cost := checkCostInputs(t, s.d.App, series, s.telita, s.hirbovat)
		// Golden 31's bases.
		if b := cost.Total.Birds; b.F+b.M != 31590 || cost.Cycles[1].Birds.F+cost.Cycles[1].Birds.M != 29907 {
			t.Errorf("total birds %+v, Hîrbovăț birds %+v; want 31 590 and 29 907", b, cost.Cycles[1].Birds)
		}
	})

	// Seria H, today 18.11.2026:
	//	Gama (growth): cycle H 01.01–31.03.2026 (emptied); placement 01.01 ♀ 100 · ♂ 10; 05.01 delivery
	//	  priced 500 lei.
	//	Lima (laying): cycle H from 01.03.2026, open; 01.03 transfer in from Gama ♀ 60 · ♂ 6; 02.03
	//	  placement ♀ 5; a placement 03.03 ♀ 1 000 whose bird row is deleted and one 04.03 ♂ 700 whose entry
	//	  is deleted (neither counts); eggs 10.03 total 50, 11.03 no Cat. 1 (null total, 999 incubabile:
	//	  not counted), 12.03 total 40; 05.03 delivery priced 1 000 lei.
	// So Gama's base ♀ 100 · ♂ 10, Lima's ♀ 65 · ♂ 6 with 90 eggs; the series placement ♀ 105 · ♂ 10.
	t.Run("hand-built", func(t *testing.T) {
		d := testdb.New(t)
		ctx := t.Context()
		id := func(sql string, args ...any) uuid.UUID {
			t.Helper()
			var id uuid.UUID
			if err := d.App.QueryRow(ctx, sql+" returning id", args...).Scan(&id); err != nil {
				t.Fatalf("%s: %v", sql, err)
			}
			return id
		}
		day := func(m time.Month, dd int) timex.Day { return timex.MustDay(2026, m, dd) }
		user := id(`insert into app_user (name, login, password_hash, role) values ('Ion', 'ion', 'x', 'owner')`)
		series := id(`insert into series (name, name_key, year, placement_date) values ('Seria H', 'seria h', 2026, $1)`, day(time.January, 1))
		gama := id(`insert into farm (name, name_key, type) values ('Gama', 'gama', 'growth')`)
		lima := id(`insert into farm (name, name_key, type) values ('Lima', 'lima', 'laying')`)
		g1 := id(`insert into section (farm_id, code, sort_order) values ($1, 'G1', 1)`, gama)
		l1 := id(`insert into section (farm_id, code, sort_order) values ($1, 'L1', 1)`, lima)
		id(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by)
			values ($1, $2, $3, $4, 'auto', 'emptied')`, series, gama, day(time.January, 1), day(time.March, 31))
		id(`insert into cycle (series_id, farm_id, start_date, started_by) values ($1, $2, $3, 'auto')`, series, lima, day(time.March, 1))

		entry := func(kind string, farm uuid.UUID, d timex.Day) uuid.UUID {
			t.Helper()
			var s *uuid.UUID
			if kind == "placement" {
				s = &series
			}
			return id(`insert into entry (kind, farm_id, date, series_id, created_by) values ($1, $2, $3, $4, $5)`, kind, farm, d, s, user)
		}
		bird := `insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, peer_section_id)
			values ($1, $2, $3, $4, $5, $6, $7, $8)`
		e := entry("placement", gama, day(time.January, 1))
		id(bird, e, gama, g1, day(time.January, 1), "F", "placement", 100, nil)
		id(bird, e, gama, g1, day(time.January, 1), "M", "placement", 10, nil)
		e = id(`insert into entry (kind, farm_id, peer_farm_id, date, created_by) values ('transfer', $1, $2, $3, $4)`,
			gama, lima, day(time.March, 1), user)
		for _, x := range []struct {
			sex string
			qty int
		}{{"F", 60}, {"M", 6}} {
			id(bird, e, gama, g1, day(time.March, 1), x.sex, "transfer", -x.qty, l1)
			id(bird, e, lima, l1, day(time.March, 1), x.sex, "transfer", x.qty, g1)
		}
		e = entry("placement", lima, day(time.March, 2))
		id(bird, e, lima, l1, day(time.March, 2), "F", "placement", 5, nil)
		e = entry("placement", lima, day(time.March, 3))
		deleted := id(bird, e, lima, l1, day(time.March, 3), "F", "placement", 1000, nil)
		if _, err := d.App.Exec(ctx, `update bird_event set deleted_at = $2 where id = $1`, deleted, testdb.Now()); err != nil {
			t.Fatal(err)
		}
		e = entry("placement", lima, day(time.March, 4))
		id(bird, e, lima, l1, day(time.March, 4), "M", "placement", 700, nil)
		if _, err := d.App.Exec(ctx, `update entry set deleted_at = $2, deleted_by = $3 where id = $1`, e, testdb.Now(), user); err != nil {
			t.Fatal(err)
		}

		eggs := `insert into egg_production (entry_id, farm_id, section_id, date, cat1, cat2, cat3, hatching_eggs)
			values ($1, $2, $3, $4, $5, $6, $7, $8)`
		for _, x := range []struct {
			d                    timex.Day
			cat1                 *int
			cat2, cat3, hatching int
		}{{day(time.March, 10), ptr(10), 10, 10, 20}, {day(time.March, 11), nil, 0, 0, 999}, {day(time.March, 12), ptr(0), 0, 0, 40}} {
			id(eggs, entry("daily", lima, x.d), lima, l1, x.d, x.cat1, x.cat2, x.cat3, x.hatching)
		}

		starter := id(`insert into feed_type (name, name_key) values ('Starter', 'starter')`)
		delivery := `insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg, price_per_kg, total_lei,
			price_from_register) values ($1, $2, $3, $4, 100, $5, $6, true)`
		id(delivery, entry("delivery", gama, day(time.January, 5)), gama, starter, day(time.January, 5), 5, 500)
		id(delivery, entry("delivery", lima, day(time.March, 5)), lima, starter, day(time.March, 5), 10, 1000)

		cost := checkCostInputs(t, d.App, series, gama, lima)
		if g, l := cost.Cycles[0], cost.Cycles[1]; g.Birds.F != 100 || g.Birds.M != 10 || l.Birds.F != 65 || l.Birds.M != 6 ||
			l.Eggs == nil || *l.Eggs != 90 {
			t.Errorf("Gama birds %+v, Lima birds %+v eggs %v; want 100 · 10, 65 · 6 and 90", g.Birds, l.Birds, l.Eggs)
		}
		if b := cost.Total.Birds; b.F != 105 || b.M != 10 {
			t.Errorf("total birds = %+v, want 105 · 10", b)
		}
	})
}

func ptr[T any](v T) *T { return &v }

// checkCostInputs builds the owner's Cost (period=cycle) on db, which must
// have a row for growth and laying (in that order, each the farm's shown
// cycle) and the total of series, and checks each base against the calc
// reader that ReportsCostInputs batches.
func checkCostInputs(t *testing.T, db *pgxpool.Pool, series, growth, laying uuid.UUID) reports.Cost {
	t.Helper()
	ctx := t.Context()
	cost, err := reports.BuildCost(ctx, db, reports.Input{
		Principal: auth.Principal{UserID: uuid.New(), Role: auth.Owner}, Today: testdb.Today, Now: testdb.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(cost.Cycles) != 2 || cost.Total == nil {
		t.Fatalf("cost = %+v, want two rows and a total", cost)
	}
	type farmCase struct {
		farm   uuid.UUID
		row    reports.CostCycle
		laying bool
	}
	for _, c := range []farmCase{{growth, cost.Cycles[0], false}, {laying, cost.Cycles[1], true}} {
		cycles, err := calc.FarmCycles(ctx, db, c.farm)
		if err != nil {
			t.Fatal(err)
		}
		shown, _ := calc.DefaultCycle(cycles)
		to := testdb.Today
		if shown.EndDate != nil {
			to = *shown.EndDate
		}
		base, err := calc.FarmBase(ctx, db, c.farm, shown.StartDate, to)
		if err != nil {
			t.Fatal(err)
		}
		if c.row.Birds.F != base.F || c.row.Birds.M != base.M {
			t.Errorf("%s birds = %+v, calc.FarmBase = %+v", c.row.Label, c.row.Birds, base)
		}
		r, err := calc.DeliveryRange(cycles, &shown, calc.Period{Kind: calc.PeriodCycle}, testdb.Today)
		if err != nil {
			t.Fatal(err)
		}
		eggs, err := calc.EggsBetween(ctx, db, c.farm, r)
		if err != nil {
			t.Fatal(err)
		}
		if c.laying && (c.row.Eggs == nil || *c.row.Eggs != eggs) {
			t.Errorf("%s eggs = %v, calc.EggsBetween = %d", c.row.Label, c.row.Eggs, eggs)
		}
	}
	placed, err := calc.SeriesPlacement(ctx, db, series, calc.AllFarms)
	if err != nil {
		t.Fatal(err)
	}
	if cost.Total.Birds.F != placed.F || cost.Total.Birds.M != placed.M {
		t.Errorf("total birds = %+v, calc.SeriesPlacement = %+v", cost.Total.Birds, placed)
	}
	return cost
}

// statements records the statements a pool runs, by their "-- name: X"
// header ("?" without one).
type statements struct {
	mu    sync.Mutex
	names []string
}

func (s *statements) TraceQueryStart(ctx context.Context, _ *pgx.Conn, d pgx.TraceQueryStartData) context.Context {
	name := "?"
	if rest, ok := strings.CutPrefix(d.SQL, "-- name: "); ok {
		name = strings.Fields(rest)[0]
	}
	s.mu.Lock()
	s.names = append(s.names, name)
	s.mu.Unlock()
	return ctx
}

func (s *statements) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (s *statements) take() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.names
	s.names = nil
	return out
}

// tracedPool is an il_app pool on d's database whose statements are recorded.
func tracedPool(t *testing.T, d *testdb.DB) (*pgxpool.Pool, *statements) {
	t.Helper()
	var st statements
	pc, err := db.PoolConfig(testdb.DSN(d.Name), 1, &testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	prepare := pc.AfterConnect
	pc.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if _, err := conn.Exec(ctx, "set role "+testdb.RoleApp); err != nil {
			return err
		}
		return prepare(ctx, conn)
	}
	pc.ConnConfig.Tracer = &st
	pool, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	st.take()
	return pool, &st
}

// TestFeedStatementsSeeded: the owner's first page is at most 8 statements,
// a cursor page 2 (M4 T3).
func TestFeedStatementsSeeded(t *testing.T) {
	s := newSeeded(t)
	pool, st := tracedPool(t, s.d)
	var q dbq.DBTX = pool
	res, err := reports.BuildFeedPriced(t.Context(), q, s.input(s.owner))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"ReportsFeedCycles", "CalcSettings", "CalcStockEnd", "CalcFeedDayComplete", "CalcConsumption",
		"CalcDeliveries", "ReportsDeliveryRows", "ReportsCostInputs",
	}
	if diff := cmp.Diff(want, st.take()); diff != "" {
		t.Errorf("first page statements (-want +got):\n%s", diff)
	}
	c, err := reports.ParseCursor(*res.Deliveries.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	in := s.input(s.owner)
	in.Cursor = &c
	if _, err := reports.BuildFeedPagePriced(t.Context(), q, in); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"ReportsFeedCycles", "ReportsDeliveryRows"}, st.take()); diff != "" {
		t.Errorf("cursor page statements (-want +got):\n%s", diff)
	}
}
