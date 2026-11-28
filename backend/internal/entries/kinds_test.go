package entries

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// T3 criterion 4 (Q1): daily, feed_daily and delivery are served; the other
// kinds are not yet.
func TestServed(t *testing.T) {
	for _, k := range []dbq.EntryKind{
		dbq.EntryKindDaily, dbq.EntryKindFeedDaily, dbq.EntryKindDelivery, dbq.EntryKindWeighing, dbq.EntryKindTransfer,
		dbq.EntryKindSale, dbq.EntryKindPlacement, dbq.EntryKindMigration,
	} {
		want := k == dbq.EntryKindDaily || k == dbq.EntryKindFeedDaily || k == dbq.EntryKindDelivery
		if Served(k) != want {
			t.Errorf("Served(%s) = %v", k, !want)
		}
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// T3 criterion 4: a daily entry's live rows as its create shape (sections in
// tree order, a sex only with a value, mortality and cull as positive
// counts, the sexing error signed, the reason as stored; the eggs) and as its
// snapshot (every row with its id, under its table). A soft-deleted row is
// not loaded.
func TestLoadStoredDaily(t *testing.T) {
	f := newFeedFixture(t)
	ctx := t.Context()
	e := f.entry(dbq.EntryKindDaily, f.farmA, nov18, f.vet, at0815)
	f.bird(e, "H5.2", nov18, "F", "mortality", 0)
	f.bird(e, "H5.1", nov18, "M", "sexing_error", -1)
	f.bird(e, "H5.1", nov18, "F", "mortality", -2)
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, cull_reason)
		values ($1, $2, $3, $4, 'F', 'cull', -1, 'boala')`, e, f.farmA, f.sec["H5.1"], nov18)
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, deleted_at)
		values ($1, $2, $3, $4, 'M', 'mortality', -5, $5)`, e, f.farmA, f.sec["H5.2"], nov18, testdb.Now())
	f.exec(`insert into egg_production (entry_id, farm_id, section_id, date, floor_eggs, cat1, cat3, hatching_eggs, avg_egg_weight_g)
		values ($1, $2, $3, $4, 98, 13, 27, 3415, 67.04)`, e, f.farmA, f.sec["H5"], nov18)

	s, err := LoadStored(ctx, f.d.App, e, dbq.EntryKindDaily, f.farmA, nov18)
	if err != nil {
		t.Fatal(err)
	}
	id := func(code string) string { return f.sec[code].String() }
	want := `{"farmId":"` + f.farmA.String() + `","date":"2026-11-18","sections":[` +
		`{"sectionId":"` + id("H5.1") + `","noLosses":false,"F":{"mortality":2,"cull":1,"cullReason":"boala","sexingError":null},` +
		`"M":{"mortality":null,"cull":null,"cullReason":null,"sexingError":-1}},` +
		`{"sectionId":"` + id("H5.2") + `","noLosses":false,"F":{"mortality":0,"cull":null,"cullReason":null,"sexingError":null},"M":null}],` +
		`"eggs":[{"sectionId":"` + id("H5") + `","floor":98,"cat1":13,"cat2":null,"cat3":27,"hatching":3415,"avgWeightG":67.04}]}`
	if got := mustJSON(t, s.Values()); got != want {
		t.Errorf("values\n got %s\nwant %s", got, want)
	}
	if len(s.Rows.BirdEvents) != 4 || len(s.Rows.EggProduction) != 1 || s.Rows.FeedAllocations != nil || !s.Live() {
		t.Fatalf("rows = %+v", s.Rows)
	}
	b := s.Rows.BirdEvents[1] // H5.1 ♀: mortality, then cull (the enum's order, the form's)
	if b.Section != "H5.1" || b.Type != dbq.BirdEventTypeCull || b.Qty != -1 || b.CullReason == nil || *b.CullReason != "boala" {
		t.Errorf("first bird row = %+v", b)
	}
	snap := mustJSON(t, s.Rows)
	wantEggs := `"eggProduction":[{"id":"` + s.Rows.EggProduction[0].ID.String() + `","sectionId":"` + id("H5") +
		`","section":"H5","floor":98,"cat1":13,"cat2":null,"cat3":27,"hatching":3415,"avgWeightG":67.04}]}`
	m := s.Rows.BirdEvents[0]
	if want := `{"birdEvents":[{"id":"` + m.ID.String() + `","sectionId":"` + id("H5.1") +
		`","section":"H5.1","sex":"F","type":"mortality","qty":-2},{"id":"` + b.ID.String() + `","sectionId":"` + id("H5.1") +
		`","section":"H5.1","sex":"F","type":"cull","qty":-1,"cullReason":"boala"},`; snap[:len(want)] != want ||
		snap[len(snap)-len(wantEggs):] != wantEggs {
		t.Errorf("snapshot = %s", snap)
	}
	codes := []string{}
	for _, sec := range s.Sections {
		codes = append(codes, sec.Code+"/"+sec.HallCode)
	}
	if mustJSON(t, codes) != `["H5.1/H5","H5.2/H5","H5/H5"]` || s.Code(f.sec["H5.2"]) != "H5.2" || len(s.FeedTypes) != 0 {
		t.Errorf("sections = %v, feed types %v", codes, s.FeedTypes)
	}
}

// T3 criterion 4: a feed_daily entry as its rows and types; a priced
// delivery as its kg, type and note (no price in the values, Q6) with the
// price kept in its snapshot (Q9); another kind is not served.
func TestLoadStoredFeedAndDelivery(t *testing.T) {
	f := newFeedFixture(t)
	ctx := t.Context()
	feed := f.entry(dbq.EntryKindFeedDaily, f.farmA, nov18, f.vet, at0815)
	f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird) values
		($1, $2, $3, $5, 'M', $7, 156), ($1, $2, $3, $5, 'F', $6, 160.5), ($1, $2, $4, $5, 'F', $6, 158)`,
		feed, f.farmA, f.sec["H5.1"], f.sec["H5.2"], nov18, f.breeder2, f.cocosi)
	s, err := LoadStored(ctx, f.d.App, feed, dbq.EntryKindFeedDaily, f.farmA, nov18)
	if err != nil {
		t.Fatal(err)
	}
	row := func(code, sex, ft, g string) string {
		return `{"sectionId":"` + f.sec[code].String() + `","sex":"` + sex + `","feedTypeId":"` + ft + `","g":` + g + `}`
	}
	want := `{"farmId":"` + f.farmA.String() + `","date":"2026-11-18","rows":[` +
		row("H5.1", "F", f.breeder2.String(), "160.5") + "," + row("H5.1", "M", f.cocosi.String(), "156") + "," +
		row("H5.2", "F", f.breeder2.String(), "158") + `]}`
	if got := mustJSON(t, s.Values()); got != want {
		t.Errorf("feed values\n got %s\nwant %s", got, want)
	}
	if got := mustJSON(t, s.FeedTypes); got != `[{"id":"`+f.breeder2.String()+`","name":"Breeder 2","forSex":"F","active":true},`+
		`{"id":"`+f.cocosi.String()+`","name":"Cocoși","forSex":"M","active":true}]` {
		t.Errorf("feed types = %s", got)
	}

	delivery := f.entry(dbq.EntryKindDelivery, f.farmA, nov17, f.ion, at0815)
	f.exec(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg, note, price_per_kg, total_lei,
		price_input, priced_by, priced_at) values ($1, $2, $3, $4, 9000, 'factura 12', 7.74, 69660, 'per_kg', $5, $6)`,
		delivery, f.farmA, f.breeder2, nov17, f.owner.UserID, testdb.Now())
	s, err = LoadStored(ctx, f.d.App, delivery, dbq.EntryKindDelivery, f.farmA, nov17)
	if err != nil {
		t.Fatal(err)
	}
	if got := mustJSON(t, s.Values()); got != `{"farmId":"`+f.farmA.String()+`","date":"2026-11-17","feedTypeId":"`+
		f.breeder2.String()+`","qtyKg":9000,"note":"factura 12"}` {
		t.Errorf("delivery values = %s", got)
	}
	d := s.Rows.FeedDeliveries[0]
	if got := mustJSON(t, s.Rows); got != `{"feedDeliveries":[{"id":"`+d.ID.String()+`","feedTypeId":"`+f.breeder2.String()+
		`","type":"Breeder 2","kg":9000,"note":"factura 12","pricePerKg":7.74,"totalLei":69660,"priceInput":"per_kg",`+
		`"pricedBy":"`+f.owner.UserID.String()+`","pricedAt":"2026-11-18T12:00:00+02:00"}]}` {
		t.Errorf("delivery snapshot = %s", got)
	}
	if len(s.Sections) != 0 || len(s.FeedTypes) != 1 {
		t.Errorf("sections %v, types %v", s.Sections, s.FeedTypes)
	}

	sale := f.entry(dbq.EntryKindSale, f.farmA, nov18, f.ion, at0815)
	if _, err := LoadStored(ctx, f.d.App, sale, dbq.EntryKindSale, f.farmA, nov18); !errors.Is(err, ErrKindNotServed) {
		t.Errorf("sale: %v", err)
	}
}
