package reports

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// fixture is a hand-built set of three farms, today 18.11.2026:
//
//	Alfa: open cycle (Seria 2026, from 01.11.2026), section A1 with ♀ 100 placed 01.11.
//	  01.11 deliveries Breeder 1 40 kg, Breeder 2 100 kg.
//	  16.11, 17.11, 18.11 Hrană zilnică A1 ♀ Breeder 2 100 g → 10 kg a day (the window 16–18.11
//	  ends today: the only row with birds is fed today).
//	  10.11 Hrană zilnică A1 ♀ Starter (ambele sexe) 5 kg in the register, never delivered.
//	  The Breeder 2 delivery is priced by Ion: 800 lei (typed total, 8 lei/kg); Breeder 1 has no price.
//	Beta (growth): one cycle (Seria 2025, 01.03.2025–28.02.2026), closed; section B1 with ♀ 200 · ♂ 20
//	  placed 01.03.2025; 10.01.2026 Starter 500 kg with a register price (4 lei/kg, 2 000 lei).
//	Gama: never a cycle; 05.11.2026 Starter 10 kg, no price.
//
// Alfa's stock at the end of 18.11: Breeder 1 40 kg (no consumption), Breeder 2 100 − 30 = 70 kg (avg 10,
// 7 days → 25.11), Starter −5 kg.
type fixture struct {
	t                *testing.T
	d                *testdb.DB
	user             uuid.UUID
	alfa, beta, gama uuid.UUID
	// The deliveries: Alfa Breeder 1 and Breeder 2, Beta Starter, Gama Starter.
	alfaB1, alfaB2, betaStarter, gamaStarter uuid.UUID
}

func day(m time.Month, d int) timex.Day { return timex.MustDay(2026, m, d) }

// pricedAt is when Ion priced Alfa's Breeder 2 delivery.
var pricedAt = time.Date(2026, time.November, 2, 9, 30, 0, 0, timex.Chisinau)

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{t: t, d: testdb.New(t)}
	f.user = f.id(`insert into app_user (name, login, password_hash, role) values ('Ion', 'ion', 'x', 'owner')`)
	s26 := f.id(`insert into series (name, name_key, year, placement_date) values ('Seria 2026', 'seria 2026', 2026, '2026-11-01')`)
	s25 := f.id(`insert into series (name, name_key, year, placement_date, status, closed_at)
		values ('Seria 2025', 'seria 2025', 2025, '2025-03-01', 'closed', $1)`, testdb.Now())
	// Inserted out of name order: the report orders farms by name.
	f.gama = f.id(`insert into farm (name, name_key, type) values ('Gama', 'gama', 'laying')`)
	f.beta = f.id(`insert into farm (name, name_key, type) values ('Beta', 'beta', 'growth')`)
	f.alfa = f.id(`insert into farm (name, name_key, type) values ('Alfa', 'alfa', 'laying')`)
	a1 := f.id(`insert into section (farm_id, code, sort_order) values ($1, 'A1', 1)`, f.alfa)
	f.id(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by)
		values ($1, $2, '2025-03-01', '2026-02-28', 'auto', 'emptied')`, s25, f.beta)
	f.id(`insert into cycle (series_id, farm_id, start_date, started_by) values ($1, $2, '2026-11-01', 'auto')`, s26, f.alfa)

	breeder1 := f.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder 1', 'breeder 1', 'F')`)
	breeder2 := f.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder 2', 'F')`)
	starter := f.id(`insert into feed_type (name, name_key) values ('Starter', 'starter')`)

	placed := day(time.November, 1)
	e := f.id(`insert into entry (kind, farm_id, date, series_id, created_by) values ('placement', $1, $2, $3, $4)`,
		f.alfa, placed, s26, f.user)
	f.id(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
		values ($1, $2, $3, $4, 'F', 'placement', 100)`, e, f.alfa, a1, placed)
	e = f.entry("delivery", placed)
	f.alfaB1 = f.id(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, 40)`,
		e, f.alfa, breeder1, placed)
	f.alfaB2 = f.id(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg, price_per_kg, total_lei,
		price_input, priced_by, priced_at) values ($1, $2, $3, $4, 100, 8, 800, 'total', $5, $6)`,
		e, f.alfa, breeder2, placed, f.user, pricedAt)

	b1 := f.id(`insert into section (farm_id, code, sort_order) values ($1, 'B1', 1)`, f.beta)
	betaPlaced := timex.MustDay(2025, time.March, 1)
	e = f.id(`insert into entry (kind, farm_id, date, series_id, created_by) values ('placement', $1, $2, $3, $4)`,
		f.beta, betaPlaced, s25, f.user)
	for _, sex := range []struct {
		sex string
		qty int
	}{{"F", 200}, {"M", 20}} {
		f.id(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
			values ($1, $2, $3, $4, $5, 'placement', $6)`, e, f.beta, b1, betaPlaced, sex.sex, sex.qty)
	}
	betaDelivered := day(time.January, 10)
	e = f.id(`insert into entry (kind, farm_id, date, created_by) values ('delivery', $1, $2, $3)`, f.beta, betaDelivered, f.user)
	f.betaStarter = f.id(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg, price_per_kg, total_lei,
		price_from_register) values ($1, $2, $3, $4, 500, 4, 2000, true)`, e, f.beta, starter, betaDelivered)
	gamaDelivered := day(time.November, 5)
	e = f.id(`insert into entry (kind, farm_id, date, created_by) values ('delivery', $1, $2, $3)`, f.gama, gamaDelivered, f.user)
	f.gamaStarter = f.id(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, 10)`,
		e, f.gama, starter, gamaDelivered)
	for _, d := range []int{16, 17, 18} {
		date := day(time.November, d)
		f.id(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
			values ($1, $2, $3, $4, 'F', $5, 100)`, f.entry("feed_daily", date), f.alfa, a1, date, breeder2)
	}
	date := day(time.November, 10)
	f.id(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, kg_recorded)
		values ($1, $2, $3, $4, 'F', $5, 5)`, f.entry("feed_daily", date), f.alfa, a1, date, starter)
	return f
}

func (f *fixture) id(sql string, args ...any) uuid.UUID {
	f.t.Helper()
	var id uuid.UUID
	if err := f.d.App.QueryRow(context.Background(), sql+" returning id", args...).Scan(&id); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
	return id
}

func (f *fixture) entry(kind string, d timex.Day) uuid.UUID {
	f.t.Helper()
	return f.id(`insert into entry (kind, farm_id, date, created_by) values ($1, $2, $3, $4)`, kind, f.alfa, d, f.user)
}

func (f *fixture) input(p auth.Principal, farm *uuid.UUID) Input {
	return Input{Principal: p, FarmID: farm, Today: testdb.Today, Now: testdb.Now()}
}

func owner() auth.Principal { return auth.Principal{UserID: uuid.New(), Role: auth.Owner} }

func principal(role auth.Role, farms ...uuid.UUID) auth.Principal {
	if farms == nil {
		farms = []uuid.UUID{}
	}
	return auth.Principal{UserID: uuid.New(), Role: role, FarmIDs: farms}
}

// Rows of the fixture's report, as encoded.
const (
	alfaRows = `{"farm":"Alfa","type":"Breeder 1","forSex":"F","kg":40,"daysLeft":null,"avgDaily":0,` +
		`"window":["2026-11-16","2026-11-18"],"runOut":null,"lamp":"oprit","reason":"– fără consum în ultimele 3 zile"},` +
		`{"farm":"Alfa","type":"Breeder 2","forSex":"F","kg":70,"daysLeft":7,"avgDaily":10,` +
		`"window":["2026-11-16","2026-11-18"],"runOut":"2026-11-25","lamp":"in_norma"},` +
		`{"farm":"Alfa","type":"Starter","forSex":null,"kg":-5,"daysLeft":0,"avgDaily":0,` +
		`"window":["2026-11-16","2026-11-18"],"runOut":"2026-11-18","lamp":"abatere","reason":"stoc sub 0 · verifică livrările"}`
	betaRow = `{"farm":"Beta","lamp":"oprit","reason":"fără ciclu deschis","sub":"ultimul ciclu: Seria 2025, închis pe 28.02.2026"}`
	gamaRow = `{"farm":"Gama","lamp":"oprit","reason":"fără ciclu deschis"}`
	asOf    = `"asOf":"2026-11-18T12:00:00+02:00"`
)

// TestBuildFeed: the Stoc part (M3 D9) is the same for every builder and
// role; it ignores the period.
func TestBuildFeed(t *testing.T) {
	f := newFixture(t)
	for _, tc := range []struct {
		name string
		in   Input
		want string
	}{
		{"owner, every farm", f.input(owner(), nil), `[` + alfaRows + `,` + betaRow + `,` + gamaRow + `]`},
		{"manager, own farms", f.input(principal(auth.Manager, f.gama, f.beta), nil), `[` + betaRow + `,` + gamaRow + `]`},
		{"viewer without farms", f.input(principal(auth.Viewer), nil), `[]`},
		{"one farm", f.input(owner(), &f.alfa), `[` + alfaRows + `]`},
		{"one farm without a cycle", f.input(principal(auth.Viewer, f.gama), &f.gama), `[` + gamaRow + `]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res, err := BuildFeed(t.Context(), f.d.App, tc.in)
			if err != nil {
				t.Fatal(err)
			}
			if got := encode(t, res.Stock); got != tc.want {
				t.Errorf("stock\n got %s\nwant %s", got, tc.want)
			}
			if !seesPrices(tc.in.Principal) {
				return
			}
			in := tc.in
			in.Period = calc.Period{Kind: calc.PeriodWeeks, Weeks: 4}
			priced, err := BuildFeedPriced(t.Context(), f.d.App, in)
			if err != nil {
				t.Fatal(err)
			}
			if got := encode(t, priced.Stock); got != tc.want {
				t.Errorf("priced stock\n got %s\nwant %s", got, tc.want)
			}
		})
	}

	unknown := uuid.New()
	for name, build := range map[string]func(Input) error{
		"BuildFeed":       func(in Input) error { _, err := BuildFeed(t.Context(), f.d.App, in); return err },
		"BuildFeedPriced": func(in Input) error { _, err := BuildFeedPriced(t.Context(), f.d.App, in); return err },
		"BuildCost":       func(in Input) error { _, err := BuildCost(t.Context(), f.d.App, in); return err },
		"BuildFeedPage": func(in Input) error {
			in.Cursor = &Cursor{Date: testdb.Today, ID: uuid.New()}
			_, err := BuildFeedPage(t.Context(), f.d.App, in)
			return err
		},
	} {
		if err := build(f.input(owner(), &unknown)); !errors.Is(err, ErrFarmNotFound) {
			t.Errorf("%s, unknown farm: err = %v, want ErrFarmNotFound", name, err)
		}
	}
}

func encode(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// fixture rows of Livrări, newest first, as the owner sees them: Gama's
// Starter (no cycle: its range is every day up to today), Alfa's Breeder 1
// (unpriced) and Breeder 2 (priced by Ion), Beta's Starter (a register price
// inside the closed Seria 2025: no action).
func (f *fixture) ownerRows() string {
	return `[{"id":"` + f.gamaStarter.String() + `","date":"2026-11-05","farm":"Gama","type":"Starter","forSex":null,"kg":10,` +
		`"pricePerKg":null,"totalLei":null,"pricedBy":null,"pricedAt":null,"tag":"fără preț","actions":{"setPrice":true,"changePrice":false}},` +
		`{"id":"` + f.alfaB1.String() + `","date":"2026-11-01","farm":"Alfa","type":"Breeder 1","forSex":"F","kg":40,` +
		`"pricePerKg":null,"totalLei":null,"pricedBy":null,"pricedAt":null,"tag":"fără preț","actions":{"setPrice":true,"changePrice":false}},` +
		`{"id":"` + f.alfaB2.String() + `","date":"2026-11-01","farm":"Alfa","type":"Breeder 2","forSex":"F","kg":100,` +
		`"pricePerKg":8,"totalLei":800,"pricedBy":"Ion","pricedAt":"2026-11-02T09:30:00+02:00","actions":{"setPrice":false,"changePrice":true}},` +
		`{"id":"` + f.betaStarter.String() + `","date":"2026-01-10","farm":"Beta","type":"Starter","forSex":null,"kg":500,` +
		`"pricePerKg":4,"totalLei":2000,"pricedBy":"din registru","pricedAt":null,"actions":{"setPrice":false,"changePrice":false}}]`
}

// The fixture's Cost for the owner, period=cycle: Beta's growth cycle (2 000 lei ÷ 220 birds, no eggs by
// type), then Alfa's laying cycle without eggs yet (800 lei ÷ 100). Two series: no total.
const (
	betaCost = `{"label":"Beta · Seria 2025","span":"01.03.2025–28.02.2026","pricedCount":1,"kg":500,"lei":2000,` +
		`"birds":{"F":200,"M":20},"eggs":null,"leiPerBird":9.0909,"leiPerEgg":null,"eggReason":"creștere, fără ouă","unpriced":null}`
	alfaCost = `{"label":"Alfa · Seria 2026","span":"din 01.11.2026 · SAPT 3","pricedCount":1,"kg":100,"lei":800,` +
		`"birds":{"F":100,"M":0},"eggs":0,"leiPerBird":8,"leiPerEgg":null,"eggReason":"fără ouă în perioadă",` +
		`"unpriced":{"count":1,"kg":40,"span":"01.11.2026"}}`
	alfaTotal = `"total":{"label":"Total Seria 2026","kg":100,"lei":800,"birds":{"F":100,"M":0},"eggs":0,"leiPerBird":8,` +
		`"leiPerEgg":null,"eggReason":"fără ouă în perioadă","excludedKg":40}`
)

// TestBuildFeedDeliveries: Livrări and Cost on the fixture (§7.12, M4 D2,
// D6, Q6–Q8).
func TestBuildFeedDeliveries(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()

	res, err := BuildFeedPriced(ctx, f.d.App, f.input(owner(), nil))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"total":4,"byFarm":[{"farm":"Alfa","count":2},{"farm":"Beta","count":1},{"farm":"Gama","count":1}],` +
		`"unpriced":{"count":2,"kg":50,"dates":["2026-11-01","2026-11-05"]},"rows":` + f.ownerRows() +
		`,"nextCursor":null,"lastBefore":null}`
	if got := encode(t, res.Deliveries); got != want {
		t.Errorf("owner deliveries\n got %s\nwant %s", got, want)
	}
	if got, want := encode(t, res.Cost), `{"asOf":"2026-11-18","cycles":[`+betaCost+`,`+alfaCost+`]}`; got != want {
		t.Errorf("owner cost\n got %s\nwant %s", got, want)
	}

	// One series with every one of its farms: the total, ÷ the series placement.
	cost, err := BuildCost(ctx, f.d.App, f.input(owner(), &f.alfa))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := encode(t, cost), `{"asOf":"2026-11-18","cycles":[`+alfaCost+`],`+alfaTotal+`}`; got != want {
		t.Errorf("Alfa cost\n got %s\nwant %s", got, want)
	}

	// A manager: the same prices, no action.
	mgr, err := BuildFeedPriced(ctx, f.d.App, f.input(principal(auth.Manager, f.alfa, f.beta), nil))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range mgr.Deliveries.Rows {
		if r.Actions.SetPrice || r.Actions.ChangePrice {
			t.Errorf("manager row %s %s has an action", r.Farm, r.Type)
		}
	}
	if len(mgr.Deliveries.Rows) != 3 || mgr.Deliveries.Rows[1].PricedBy == nil || *mgr.Deliveries.Rows[1].PricedBy != "Ion" {
		t.Errorf("manager rows = %s", encode(t, mgr.Deliveries.Rows))
	}

	// The viewer: rows without prices, the unpriced count only, no Cost.
	view, err := BuildFeed(ctx, f.d.App, f.input(principal(auth.Viewer, f.alfa), nil))
	if err != nil {
		t.Fatal(err)
	}
	want = `{"total":2,"byFarm":[{"farm":"Alfa","count":2}],"unpriced":{"count":1},"rows":[` +
		`{"id":"` + f.alfaB1.String() + `","date":"2026-11-01","farm":"Alfa","type":"Breeder 1","forSex":"F","kg":40},` +
		`{"id":"` + f.alfaB2.String() + `","date":"2026-11-01","farm":"Alfa","type":"Breeder 2","forSex":"F","kg":100}],` +
		`"nextCursor":null,"lastBefore":null}`
	if got := encode(t, view.Deliveries); got != want {
		t.Errorf("viewer deliveries\n got %s\nwant %s", got, want)
	}
	if _, err := BuildFeedPriced(ctx, f.d.App, f.input(principal(auth.Viewer, f.alfa), nil)); !errors.Is(err, ErrPricesHidden) {
		t.Errorf("BuildFeedPriced for a viewer: err = %v, want ErrPricesHidden", err)
	}
	if _, err := BuildCost(ctx, f.d.App, f.input(principal(auth.Vet, f.alfa), nil)); !errors.Is(err, ErrPricesHidden) {
		t.Errorf("BuildCost for a vet: err = %v, want ErrPricesHidden", err)
	}
}

// TestBuildFeedPeriods: the period sets each farm's range (M4 Q6); an empty
// range shows the newest delivery before it, and a range with deliveries but
// none priced has no cost row.
func TestBuildFeedPeriods(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	for _, tc := range []struct {
		name     string
		farm     uuid.UUID
		period   calc.Period
		total    int
		last     string
		costRows int
	}{
		// 02.11–18.11: Alfa's deliveries of 01.11 are before it.
		{"interval after the deliveries", f.alfa, calc.Period{Kind: calc.PeriodInterval, From: day(time.November, 2), To: day(time.November, 18)},
			0, `{"date":"2026-11-01","farm":"Alfa","type":"Breeder 1","kg":40}`, 0},
		// Beta's last 4 weeks of its cycle (SAPT 50–53, 07.02–28.02.2026) hold no delivery; the one of 10.01 is before.
		{"weeks before the delivery", f.beta, calc.Period{Kind: calc.PeriodWeeks, Weeks: 4},
			0, `{"date":"2026-01-10","farm":"Beta","type":"Starter","kg":500}`, 0},
		// An interval before every delivery: nothing before it either.
		{"nothing before", f.alfa, calc.Period{Kind: calc.PeriodInterval, From: day(time.October, 1), To: day(time.October, 31)},
			0, `null`, 0},
		// Gama: no cycle; the interval is as typed. Its delivery has no price: no cost row.
		{"no cycle, unpriced", f.gama, calc.Period{Kind: calc.PeriodInterval, From: day(time.November, 1), To: day(time.November, 18)},
			1, `null`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := f.input(owner(), &tc.farm)
			in.Period = tc.period
			res, err := BuildFeedPriced(ctx, f.d.App, in)
			if err != nil {
				t.Fatal(err)
			}
			d := res.Deliveries
			if d.Total != tc.total || len(d.Rows) != tc.total || encode(t, d.LastBefore) != tc.last || len(res.Cost.Cycles) != tc.costRows {
				t.Errorf("total %d, rows %d, lastBefore %s, cost rows %d; want %d, %d, %s, %d",
					d.Total, len(d.Rows), encode(t, d.LastBefore), len(res.Cost.Cycles), tc.total, tc.total, tc.last, tc.costRows)
			}
		})
	}
}

// TestFeedPagesLimit: pages of one row walk the fixture's deliveries in
// order, and the last one has no next cursor. Two more Gama Starter
// deliveries of 05.11 tie on (date, farm, type): the id orders them, so no
// page boundary drops or repeats one.
func TestFeedPagesLimit(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	e := f.id(`insert into entry (kind, farm_id, date, created_by) values ('delivery', $1, $2, $3)`,
		f.gama, day(time.November, 5), f.user)
	var ties []uuid.UUID
	for range 2 {
		ties = append(ties, f.id(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg)
			select $1, $2, feed_type_id, date, 5 from feed_delivery where id = $3`, e, f.gama, f.gamaStarter))
	}
	ties = append(ties, f.gamaStarter)
	slices.SortFunc(ties, func(a, b uuid.UUID) int { return strings.Compare(a.String(), b.String()) })
	in := f.input(owner(), nil)
	in.Limit = 1
	first, err := BuildFeedPriced(ctx, f.d.App, in)
	if err != nil {
		t.Fatal(err)
	}
	ids := []uuid.UUID{first.Deliveries.Rows[0].ID}
	next := first.Deliveries.NextCursor
	for next != nil {
		c, err := ParseCursor(*next)
		if err != nil {
			t.Fatal(err)
		}
		in.Cursor = &c
		pg, err := BuildFeedPagePriced(ctx, f.d.App, in)
		if err != nil {
			t.Fatal(err)
		}
		if len(pg.Deliveries.Rows) != 1 {
			t.Fatalf("page of %d rows, want 1", len(pg.Deliveries.Rows))
		}
		ids = append(ids, pg.Deliveries.Rows[0].ID)
		next = pg.Deliveries.NextCursor
	}
	if want := append(ties, f.alfaB1, f.alfaB2, f.betaStarter); !slices.Equal(ids, want) {
		t.Errorf("ids = %v, want %v", ids, want)
	}
}

func TestParseCursor(t *testing.T) {
	c := Cursor{Date: day(time.October, 2), FarmKey: "hirbovat", TypeKey: "breeder 2", ID: uuid.New()}
	got, err := ParseCursor(c.String())
	if err != nil || got != c {
		t.Fatalf("round trip: %+v, %v; want %+v", got, err, c)
	}
	b64 := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
	for name, s := range map[string]string{
		"empty":        "",
		"not base64":   "***",
		"padded":       c.String() + "=",
		"not JSON":     b64("x"),
		"three values": b64(`["2026-10-02","a","b"]`),
		"five values":  b64(`["2026-10-02","a","b","` + c.ID.String() + `","x"]`),
		"not strings":  b64(`[1,2,3,4]`),
		"bad day":      b64(`["2026-02-30","a","b","` + c.ID.String() + `"]`),
		"bad id":       b64(`["2026-10-02","a","b","x"]`),
		"an object":    b64(`{"date":"2026-10-02"}`),
		"far too long": strings.Repeat("A", 600),
	} {
		if _, err := ParseCursor(s); !errors.Is(err, ErrCursor) {
			t.Errorf("%s: err = %v, want ErrCursor", name, err)
		}
	}
}

// server serves the reports routes, the caller being p (nil: no session
// middleware ran).
func (f *fixture) server(p *auth.Principal) http.Handler {
	mux := http.NewServeMux()
	withCaller := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if p != nil {
				r = r.WithContext(auth.WithPrincipal(r.Context(), *p))
			}
			next.ServeHTTP(w, r)
		})
	}
	NewHandler(NewService(db.NewStore(f.d.App), f.d.Clock)).Routes(mux, httpx.NewChain(withCaller))
	return mux
}

type errorBody struct {
	Error struct {
		Code    string        `json:"code"`
		Message string        `json:"message"`
		Issues  []httpx.Issue `json:"issues"`
	} `json:"error"`
}

func TestFeedHandler(t *testing.T) {
	f := newFixture(t)
	alfa, beta := f.alfa.String(), f.beta.String()
	vet := principal(auth.Vet, f.alfa)
	viewer := principal(auth.Viewer, f.beta)
	manager := principal(auth.Manager, f.alfa, f.beta)
	own := owner()
	const path = "/api/reports/feed"
	page2 := Cursor{Date: day(time.November, 1), FarmKey: "alfa", TypeKey: "breeder 1", ID: f.alfaB1}.String()

	var (
		viewerKeys = []string{"asOf", "deliveries", "stock"}
		pricedKeys = []string{"asOf", "cost", "deliveries", "stock"}
		pageKeys   = []string{"asOf", "deliveries"}
	)
	for _, tc := range []struct {
		name      string
		caller    *auth.Principal
		query     string
		status    int
		code      string
		issuePath string
		issueCode string
		message   string
		keys      []string // the top-level keys of a 200 body
		stock     string   // its stock, when it has one
		rows      int      // its deliveries.rows
	}{
		{name: "no session", query: "", status: 401, code: "UNAUTHENTICATED"},
		{name: "vet", caller: &vet, status: 403, code: "FORBIDDEN"},
		{name: "viewer on another farm", caller: &viewer, query: "?farmId=" + alfa, status: 403, code: "FARM_FORBIDDEN",
			message: "Nu ai acces la această fermă."},
		{name: "viewer on an unknown farm", caller: &viewer, query: "?farmId=" + uuid.NewString(), status: 403, code: "FARM_FORBIDDEN"},
		{name: "owner on an unknown farm", caller: &own, query: "?farmId=" + uuid.NewString(), status: 404, code: "NOT_FOUND"},
		{name: "owner on an unknown farm, a cursor page", caller: &own, query: "?farmId=" + uuid.NewString() + "&cursor=" + page2,
			status: 404, code: "NOT_FOUND"},
		{name: "malformed farm id", caller: &manager, query: "?farmId=alfa", status: 400, code: "VALIDATION",
			issuePath: "farmId", issueCode: "bad_type"},
		{name: "repeated farm id", caller: &manager, query: "?farmId=all&farmId=all", status: 400, code: "VALIDATION",
			issuePath: "farmId", issueCode: "bad_type"},
		{name: "unknown parameter", caller: &manager, query: "?view=stock", status: 400, code: "VALIDATION",
			issuePath: "view", issueCode: "unknown_field"},
		{name: "bad period", caller: &manager, query: "?period=weeks:x", status: 400, code: "VALIDATION",
			issuePath: "period", issueCode: "bad_type"},
		{name: "unknown period", caller: &manager, query: "?period=month", status: 400, code: "VALIDATION",
			issuePath: "period", issueCode: "out_of_range"},
		{name: "interval in the future", caller: &manager, query: "?period=interval:2026-11-01,2026-11-20", status: 400,
			code: "VALIDATION", issuePath: "period", issueCode: "out_of_range",
			message: "Până la 20.11.2026 e în viitor. Alege o dată până la 18.11.2026."},
		{name: "bad cursor", caller: &manager, query: "?cursor=abc", status: 400, code: "VALIDATION",
			issuePath: "cursor", issueCode: "bad_type", message: "Datele trimise nu sunt valide."},
		{name: "repeated cursor", caller: &manager, query: "?cursor=" + page2 + "&cursor=" + page2, status: 400, code: "VALIDATION",
			issuePath: "cursor", issueCode: "bad_type"},
		{name: "limit 0", caller: &manager, query: "?limit=0", status: 400, code: "VALIDATION",
			issuePath: "limit", issueCode: "out_of_range"},
		{name: "limit 101", caller: &manager, query: "?limit=101", status: 400, code: "VALIDATION",
			issuePath: "limit", issueCode: "out_of_range"},
		{name: "limit not a number", caller: &manager, query: "?limit=12.5", status: 400, code: "VALIDATION",
			issuePath: "limit", issueCode: "bad_type"},
		{name: "viewer, own farms", caller: &viewer, status: 200, keys: viewerKeys, stock: `[` + betaRow + `]`, rows: 1},
		{name: "manager, all, a period", caller: &manager, query: "?farmId=all&period=weeks:4", status: 200, keys: pricedKeys,
			stock: `[` + alfaRows + `,` + betaRow + `]`, rows: 2},
		{name: "manager, one farm", caller: &manager, query: "?farmId=" + beta + "&period=interval:2026-01-01,2026-11-18",
			status: 200, keys: pricedKeys, stock: `[` + betaRow + `]`, rows: 1},
		{name: "owner", caller: &own, query: "?period=cycle", status: 200, keys: pricedKeys,
			stock: `[` + alfaRows + `,` + betaRow + `,` + gamaRow + `]`, rows: 4},
		{name: "owner, a limit", caller: &own, query: "?limit=1", status: 200, keys: pricedKeys,
			stock: `[` + alfaRows + `,` + betaRow + `,` + gamaRow + `]`, rows: 1},
		{name: "owner, a cursor page", caller: &own, query: "?cursor=" + page2 + "&limit=100", status: 200, keys: pageKeys, rows: 2},
		{name: "viewer, a cursor page", caller: &viewer, query: "?cursor=" + page2, status: 200, keys: pageKeys, rows: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			f.server(tc.caller).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path+tc.query, nil))
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.status, rec.Body)
			}
			if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", cc)
			}
			if tc.status == http.StatusOK {
				var body map[string]json.RawMessage
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if keys := slices.Sorted(maps.Keys(body)); !slices.Equal(keys, tc.keys) {
					t.Errorf("keys = %v, want %v", keys, tc.keys)
				}
				if tc.stock != "" && string(body["stock"]) != tc.stock {
					t.Errorf("stock\n got %s\nwant %s", body["stock"], tc.stock)
				}
				var d struct {
					Rows []json.RawMessage `json:"rows"`
				}
				if err := json.Unmarshal(body["deliveries"], &d); err != nil {
					t.Fatal(err)
				}
				if len(d.Rows) != tc.rows {
					t.Errorf("rows = %d, want %d", len(d.Rows), tc.rows)
				}
				if string(body["asOf"]) != `"2026-11-18T12:00:00+02:00"` {
					t.Errorf("asOf = %s", body["asOf"])
				}
				if slices.Equal(tc.keys, viewerKeys) || tc.caller.Role == auth.Viewer {
					for _, k := range []string{`"pricePerKg"`, `"totalLei"`, `"cost"`, `"pricedBy"`, `"fără preț"`} {
						if strings.Contains(rec.Body.String(), k) {
							t.Errorf("the viewer's body holds %s", k)
						}
					}
				}
				return
			}
			var body errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Code != tc.code {
				t.Errorf("code = %s, want %s", body.Error.Code, tc.code)
			}
			if tc.issuePath != "" && (len(body.Error.Issues) != 1 || body.Error.Issues[0].Path != tc.issuePath ||
				body.Error.Issues[0].Code != tc.issueCode) {
				t.Errorf("issues = %+v, want one %s %s", body.Error.Issues, tc.issuePath, tc.issueCode)
			}
			if tc.message != "" && body.Error.Message != tc.message && (len(body.Error.Issues) == 0 || body.Error.Issues[0].Message != tc.message) {
				t.Errorf("message = %q, want %q", body.Error.Message, tc.message)
			}
		})
	}
}
