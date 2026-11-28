package golden_test

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/farms"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/golden"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/periodtable"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/prices"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/reports"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/series"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// The JSON golden files (SPEC §11.8, M2 T9, M3 D12, M4 D8): the PeriodTable
// (Rezumat and Consum), farm page, summary, bootstrap, Hrană zilnică and
// Livrare hrană form states, Rapoarte › Hrană (Stoc, Livrări, Cost), the
// Prețuri list and series page responses on the dev seed as of 18.11.2026
// 12:00, built by the builders the handlers call (no HTTP: route wiring is
// cmd/api's), encoded like httpx.JSON and indented. Ids are replaced by
// stable tokens (M2 D9), because the import's ids are random; so are the
// Livrări cursors, which carry an id, and undo tokens, which carry a
// signature.
//
// Regenerate with: go test ./internal/golden -run TestJSONGolden -update

var update = flag.Bool("update", false, "rewrite the JSON golden files under testdata/")

// fixture is the seeded database with the ids the cases need.
type fixture struct {
	db                dbq.DBTX
	hirbovat, telita  uuid.UUID
	seria2025Hirbovat uuid.UUID // Hîrbovăț's Seria 2025 cycle (no flock count)
	seria2026Telita   uuid.UUID // Telița's Seria 2026 cycle (growth)
	seria2025         uuid.UUID // the closed series
	seria2026         uuid.UUID // the active series
	tokens            map[uuid.UUID]string
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	ctx := t.Context()
	f := fixture{db: testdb.NewSeeded(t).App}
	f.hirbovat = farmID(t, f.db, "hirbovat")
	f.telita = farmID(t, f.db, "telita")
	f.seria2025Hirbovat = cycleID(t, f.db, f.hirbovat, "Seria 2025")
	f.seria2026Telita = cycleID(t, f.db, f.telita, "Seria 2026")
	f.seria2025 = seriesID(t, f.db, "seria2025")
	f.seria2026 = seriesID(t, f.db, "seria2026")
	f.tokens = idTokens(t, ctx, f.db)
	return f
}

func farmID(t *testing.T, db dbq.DBTX, key string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := db.QueryRow(t.Context(), `select id from farm where name_key = $1`, key).Scan(&id); err != nil {
		t.Fatalf("farm %q: %v", key, err)
	}
	return id
}

func seriesID(t *testing.T, db dbq.DBTX, key string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := db.QueryRow(t.Context(), `select id from series where name_key = $1`, key).Scan(&id); err != nil {
		t.Fatalf("series %q: %v", key, err)
	}
	return id
}

func cycleID(t *testing.T, db dbq.DBTX, farm uuid.UUID, series string) uuid.UUID {
	t.Helper()
	cycles, err := calc.FarmCycles(t.Context(), db, farm)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cycles {
		if c.SeriesName == series {
			return c.ID
		}
	}
	t.Fatalf("farm %s has no %s cycle", farm, series)
	return uuid.Nil
}

// idTokens maps every id a response can carry to its stable token (M2 D9):
// <farm:hirbovat>, <series:seria2026>, <cycle:hirbovat/2026>,
// <section:hirbovat/H5.1>, <guide:…>, <feedType:…>, an entry that is
// the only one of its farm, kind and day: <entry:hirbovat/feed_daily/2026-11-18>
// and a delivery that is the only one of its farm, day and feed type:
// <delivery:hirbovat/2026-11-13/breeder2> (an entry or a delivery sharing
// them gets no token: a response naming it fails, as any id without a
// token).
func idTokens(t *testing.T, ctx context.Context, db dbq.DBTX) map[uuid.UUID]string {
	t.Helper()
	rows, err := db.Query(ctx, `
select id, 'farm:' || name_key from farm
union all select id, 'series:' || name_key from series
union all select c.id, 'cycle:' || f.name_key || '/' || s.year
  from cycle c join farm f on f.id = c.farm_id join series s on s.id = c.series_id
union all select s.id, 'section:' || f.name_key || '/' || s.code from section s join farm f on f.id = s.farm_id
union all select id, 'guide:' || name_key from guide
union all select id, 'feedType:' || name_key from feed_type
union all select id, token from (
  select e.id, 'entry:' || f.name_key || '/' || e.kind::text || '/' || to_char(e.date, 'YYYY-MM-DD') as token,
         count(*) over (partition by e.farm_id, e.kind, e.date) as n
  from entry e join farm f on f.id = e.farm_id) e where n = 1
union all select id, token from (
  select d.id, 'delivery:' || f.name_key || '/' || to_char(d.date, 'YYYY-MM-DD') || '/' || ft.name_key as token,
         count(*) over (partition by d.farm_id, d.date, d.feed_type_id) as n
  from feed_delivery d join farm f on f.id = d.farm_id join feed_type ft on ft.id = d.feed_type_id) d where n = 1`)
	if err != nil {
		t.Fatal(err)
	}
	type idToken struct {
		ID    uuid.UUID
		Token string
	}
	all, err := pgx.CollectRows(rows, pgx.RowToStructByPos[idToken])
	if err != nil {
		t.Fatal(err)
	}
	out := make(map[uuid.UUID]string, len(all))
	seen := map[string]bool{}
	for _, r := range all {
		token := "<" + r.Token + ">"
		if seen[token] {
			t.Fatalf("token %s names two ids: the tokens must be unique", token)
		}
		seen[token] = true
		out[r.ID] = token
	}
	return out
}

var (
	uuidText = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	// cursorText and undoTokenText find a Livrări cursor (reports.Cursor) and
	// a price undo token in the indented JSON.
	cursorText    = regexp.MustCompile(`"nextCursor": "([^"]*)"`)
	undoTokenText = regexp.MustCompile(`"undoToken": "([^"]*)"`)
)

// normalise encodes v like httpx.JSON, indents it and replaces every id with
// its token. A Livrări cursor becomes <cursor:day/farm/type/id> with its id
// tokenised, and an undo token <undoToken>. An id without a token, a cursor
// that doesn't decode or an empty undo token fails the test.
func normalise(t *testing.T, v any, tokens map[uuid.UUID]string) []byte {
	t.Helper()
	b, err := httpx.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := json.Indent(&out, b, "", "  "); err != nil {
		t.Fatal(err)
	}
	var bad []string
	text := cursorText.ReplaceAllStringFunc(out.String(), func(s string) string {
		raw := cursorText.FindStringSubmatch(s)[1]
		c, err := reports.ParseCursor(raw)
		if err != nil {
			bad = append(bad, "cursor "+raw)
			return s
		}
		return fmt.Sprintf(`"nextCursor": "<cursor:%s/%s/%s/%s>"`, c.Date, c.FarmKey, c.TypeKey, c.ID)
	})
	text = undoTokenText.ReplaceAllStringFunc(text, func(s string) string {
		if undoTokenText.FindStringSubmatch(s)[1] == "" {
			bad = append(bad, "an empty undoToken")
			return s
		}
		return `"undoToken": "<undoToken>"`
	})
	if len(bad) > 0 {
		t.Fatalf("not normalisable: %q", bad)
	}
	var unknown []string
	text = uuidText.ReplaceAllStringFunc(text, func(s string) string {
		if tok, ok := tokens[uuid.MustParse(s)]; ok {
			return tok
		}
		unknown = append(unknown, s)
		return s
	})
	if len(unknown) > 0 {
		t.Fatalf("ids without a token (add their table to idTokens): %v", unknown)
	}
	return []byte(text)
}

// compareGolden compares got with testdata/<name>, or rewrites the file with
// -update.
func compareGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if diff := cmp.Diff(string(want), string(got)); diff != "" {
		t.Errorf("%s differs (-want +got; -update rewrites it):\n%s", path, diff)
	}
}

// tableQuery parses a GET /api/table query string the way the handler does.
func tableQuery(t *testing.T, raw string) periodtable.TableQuery {
	t.Helper()
	values, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatal(err)
	}
	q, err := periodtable.ParseQuery(values)
	if err != nil {
		t.Fatalf("query %q: %v", raw, err)
	}
	var v httpx.Validator
	q.Validate(&v)
	if err := v.Err(); err != nil {
		t.Fatalf("query %q: %v", raw, err)
	}
	return q
}

func (f fixture) table(t *testing.T, raw string) periodtable.TableResponse {
	t.Helper()
	res, err := periodtable.Build(t.Context(), f.db, periodtable.Input{Query: tableQuery(t, raw), Today: golden.AsOf, Now: golden.Now})
	if err != nil {
		t.Fatalf("Build(%s): %v", raw, err)
	}
	return res
}

func (f fixture) input(p auth.Principal, farm uuid.UUID) farms.Input {
	return farms.Input{Principal: p, FarmID: farm, Today: golden.AsOf, Now: golden.Now}
}

var (
	owner = auth.Principal{Name: "Proprietar", Role: auth.Owner}
)

// viewer sees Hîrbovăț only.
func (f fixture) viewer() auth.Principal {
	return auth.Principal{Name: "Observator", Role: auth.Viewer, FarmIDs: []uuid.UUID{f.hirbovat}}
}

// manager manages both farms, as the seed's Ion Rusu.
func (f fixture) manager() auth.Principal {
	return auth.Principal{Name: "Ion Rusu", Role: auth.Manager, FarmIDs: []uuid.UUID{f.hirbovat, f.telita}}
}

// tableCases are the PeriodTable golden files: query string → file.
func (f fixture) tableCases() []struct{ file, query string } {
	h := "view=farm&farmId=" + f.hirbovat.String()
	return []struct{ file, query string }{
		{"table_farm_rezumat_cycle.json", h},
		{"table_farm_rezumat_before56.json", h + "&before=56"},
		{"table_farm_rezumat_before24.json", h + "&before=24"},
		{"table_farm_rezumat_weeks4.json", h + "&period=weeks:4"},
		{"table_farm_rezumat_grain_week.json", h + "&grain=week"},
		{"table_farm_rezumat_telita_growth.json", "view=farm&farmId=" + f.telita.String() + "&cycleId=" + f.seria2026Telita.String() + "&period=cycle"},
		{"table_farm_rezumat_seria2025.json", h + "&cycleId=" + f.seria2025Hirbovat.String()},
		{"table_report_feed_cycle.json", "view=report-feed&farmId=" + f.hirbovat.String()},
		{"table_report_feed_before56.json", "view=report-feed&farmId=" + f.hirbovat.String() + "&before=56"},
	}
}

func TestJSONGolden(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	for _, tc := range f.tableCases() {
		t.Run(tc.file, func(t *testing.T) {
			res := f.table(t, tc.query)
			if p := nullContainers(res); len(p) > 0 {
				t.Errorf("null lists at %q", p)
			}
			compareGolden(t, filepath.Join("periodtable", tc.file), normalise(t, res, f.tokens))
		})
	}

	farmCases := []struct {
		file  string
		build func() (any, error)
	}{
		{"page_hirbovat.json", func() (any, error) { return farms.BuildPage(ctx, f.db, f.input(owner, f.hirbovat)) }},
		{"page_telita_empty.json", func() (any, error) { return farms.BuildPage(ctx, f.db, f.input(owner, f.telita)) }},
		{"summary_hirbovat.json", func() (any, error) { return farms.BuildSummary(ctx, f.db, f.input(owner, f.hirbovat)) }},
		{"summary_hirbovat_seria2025.json", func() (any, error) {
			in := f.input(owner, f.hirbovat)
			in.CycleID = &f.seria2025Hirbovat
			return farms.BuildSummary(ctx, f.db, in)
		}},
		{"bootstrap_owner.json", func() (any, error) { return farms.BuildBootstrap(ctx, f.db, f.input(owner, uuid.Nil)) }},
	}
	for _, tc := range farmCases {
		t.Run(tc.file, func(t *testing.T) {
			res, err := tc.build()
			if err != nil {
				t.Fatal(err)
			}
			if p := nullContainers(res); len(p) > 0 {
				t.Errorf("null lists at %q", p)
			}
			compareGolden(t, filepath.Join("farms", tc.file), normalise(t, res, f.tokens))
		})
	}

	// M3 (D12): the Hrană zilnică form state and the Stoc report; M4 (D8):
	// Rapoarte › Hrană with Livrări and Cost per role, the Livrare hrană form
	// state, the Prețuri list and the series page.
	feedCases := []struct {
		file   string
		build  func() (any, error)
		viewer bool // a viewer's response: no price field either (D64)
	}{
		{"forms/feed_daily_hirbovat_18.json", func() (any, error) {
			return entries.BuildFeedForm(ctx, f.db, entries.FeedFormInput{
				Principal: owner, FarmID: f.hirbovat, Date: golden.AsOf, Today: golden.AsOf, Now: golden.Now,
			})
		}, false},
		{"reports/feed_owner.json", func() (any, error) { return f.reportsFeedPriced(ctx, owner, nil, calc.Period{}) }, false},
		{"reports/feed_manager.json", func() (any, error) { return f.reportsFeedPriced(ctx, f.manager(), nil, calc.Period{}) }, false},
		{"reports/feed_viewer.json", func() (any, error) { return f.reportsFeed(ctx, f.viewer()) }, true},
		{"reports/feed_owner_deliveries_page2.json", func() (any, error) { return f.reportsFeedPage2(t, owner) }, false},
		{"reports/feed_telita_interval.json", func() (any, error) {
			return f.reportsFeedPriced(ctx, owner, &f.telita, calc.Period{Kind: calc.PeriodInterval,
				From: day(2026, time.March, 1), To: golden.AsOf})
		}, false},
		{"prices/unpriced_owner.json", func() (any, error) { return prices.BuildUnpriced(ctx, f.db, golden.Now) }, false},
		{"forms/delivery_hirbovat.json", func() (any, error) {
			return entries.BuildDeliveryForm(ctx, f.db, entries.DeliveryFormInput{FarmID: f.hirbovat, Today: golden.AsOf, Now: golden.Now})
		}, false},
		{"series/page_seria2026_owner.json", func() (any, error) { return f.seriesPage(ctx, owner, f.seria2026) }, false},
		{"series/page_seria2026_viewer_hirbovat.json", func() (any, error) { return f.seriesPage(ctx, f.viewer(), f.seria2026) }, true},
		{"series/page_seria2025_owner.json", func() (any, error) { return f.seriesPage(ctx, owner, f.seria2025) }, false},
	}
	for _, tc := range feedCases {
		t.Run(tc.file, func(t *testing.T) {
			res, err := tc.build()
			if err != nil {
				t.Fatal(err)
			}
			if p := nullContainers(res); len(p) > 0 {
				t.Errorf("null lists at %q", p)
			}
			got := normalise(t, res, f.tokens)
			if tc.viewer {
				found, err := priceFields(got)
				if err != nil {
					t.Fatal(err)
				}
				if len(found) > 0 {
					t.Errorf("a viewer response carries prices %q", found)
				}
			}
			compareGolden(t, tc.file, got)
		})
	}
}

// reportsFeed is GET /api/reports/feed with farmId=all (every farm of p) as
// a viewer gets it (no prices).
func (f fixture) reportsFeed(ctx context.Context, p auth.Principal) (reports.FeedResponse, error) {
	return reports.BuildFeed(ctx, f.db, reports.Input{Principal: p, Today: golden.AsOf, Now: golden.Now})
}

// reportsFeedPriced is GET /api/reports/feed as a manager or the owner gets
// it, for farm (nil: farmId=all) and period (zero: cycle).
func (f fixture) reportsFeedPriced(ctx context.Context, p auth.Principal, farm *uuid.UUID, period calc.Period) (reports.FeedResponsePriced, error) {
	return reports.BuildFeedPriced(ctx, f.db, reports.Input{Principal: p, FarmID: farm, Period: period, Today: golden.AsOf, Now: golden.Now})
}

// reportsFeedPage2 is the second page of Livrări (farmId=all, period=cycle):
// GET /api/reports/feed?cursor= with the first page's nextCursor.
func (f fixture) reportsFeedPage2(t *testing.T, p auth.Principal) (reports.FeedPagePriced, error) {
	t.Helper()
	first, err := f.reportsFeedPriced(t.Context(), p, nil, calc.Period{})
	if err != nil {
		return reports.FeedPagePriced{}, err
	}
	if first.Deliveries.NextCursor == nil {
		t.Fatal("the first Livrări page has no nextCursor")
	}
	c, err := reports.ParseCursor(*first.Deliveries.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	return reports.BuildFeedPagePriced(t.Context(), f.db, reports.Input{Principal: p, Cursor: &c, Today: golden.AsOf, Now: golden.Now})
}

// seriesPage is GET /api/series/{id}/page for p.
func (f fixture) seriesPage(ctx context.Context, p auth.Principal, id uuid.UUID) (series.PageResponse, error) {
	return series.BuildPage(ctx, f.db, series.Input{Principal: p, SeriesID: id, Today: golden.AsOf, Now: golden.Now})
}

// TestViewerResponsesHaveNoPricesOrNullLists builds every M2–M4 response a
// viewer can get, as a viewer of Hîrbovăț sees it, and walks it: no lei,
// price or cost field and no "fără preț" tag (D2, D64, D200, §3.3) and no
// null list (§7.0a).
func TestViewerResponsesHaveNoPricesOrNullLists(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	viewer := f.viewer()
	responses := map[string]func(t *testing.T) (any, error){
		"page":         func(*testing.T) (any, error) { return farms.BuildPage(ctx, f.db, f.input(viewer, f.hirbovat)) },
		"summary":      func(*testing.T) (any, error) { return farms.BuildSummary(ctx, f.db, f.input(viewer, f.hirbovat)) },
		"bootstrap":    func(*testing.T) (any, error) { return farms.BuildBootstrap(ctx, f.db, f.input(viewer, uuid.Nil)) },
		"reports/feed": func(*testing.T) (any, error) { return f.reportsFeed(ctx, viewer) },
		"reports/feed cursor page": func(t *testing.T) (any, error) {
			first, err := f.reportsFeed(ctx, viewer)
			if err != nil || first.Deliveries.NextCursor == nil {
				t.Fatalf("first page: nextCursor %v, err %v", first.Deliveries.NextCursor, err)
			}
			c, err := reports.ParseCursor(*first.Deliveries.NextCursor)
			if err != nil {
				t.Fatal(err)
			}
			return reports.BuildFeedPage(ctx, f.db, reports.Input{Principal: viewer, Cursor: &c, Today: golden.AsOf, Now: golden.Now})
		},
		"series page Seria 2026": func(*testing.T) (any, error) { return f.seriesPage(ctx, viewer, f.seria2026) },
		"series page Seria 2025": func(*testing.T) (any, error) { return f.seriesPage(ctx, viewer, f.seria2025) },
	}
	for _, tc := range f.tableCases() {
		if strings.Contains(tc.query, f.telita.String()) {
			continue // the viewer sees Hîrbovăț only
		}
		responses["table "+tc.file] = func(t *testing.T) (any, error) { return f.table(t, tc.query), nil }
	}
	for name, build := range responses {
		t.Run(name, func(t *testing.T) {
			res, err := build(t)
			if err != nil {
				t.Fatal(err)
			}
			if p := nullContainers(res); len(p) > 0 {
				t.Errorf("null lists at %q", p)
			}
			b, err := httpx.Marshal(res)
			if err != nil {
				t.Fatal(err)
			}
			found, err := priceFields(b)
			if err != nil {
				t.Fatal(err)
			}
			if len(found) > 0 {
				t.Errorf("a viewer response carries prices %q", found)
			}
		})
	}
}

// exampleCell is a cell of the §7.4 example response, written with the
// decimals the example shows.
type exampleCell = map[string]string

// TestTableFirstPageMatchesSpecExample compares the first page with the §7.4
// example response: every cell the example lists, rounded once (half away
// from zero) to the decimals the example writes (the wire keeps ≤ 4, §7.4),
// and the paging fields.
func TestTableFirstPageMatchesSpecExample(t *testing.T) {
	f := newFixture(t)
	res := f.table(t, "view=farm&farmId="+f.hirbovat.String())

	week59 := exampleCell{"alive.F": "12561", "alive.M": "794", "deaths.F": "10", "deaths.M": "0", "egg.total": "11806",
		"egg.layPct": "40.40", "egg.guide": "57.6", "egg.inc": "10918", "egg.incPct": "92.48", "egg.jos": "311", "egg.josPct": "2.63",
		"feed.kg.F": "4675.3", "feed.kg.M": "329.7", "sold.F": "2255", "sold.M": "433"}
	day17 := exampleCell{"alive.F": "14395", "alive.M": "1083", "deaths.F": "7", "deaths.M": "0", "egg.total": "5892",
		"egg.layPct": "39.74", "egg.guide": "57.6", "egg.inc": "5610", "egg.incPct": "95.21", "egg.jos": "153", "egg.josPct": "2.60",
		"feed.kg.F": "2372.1", "feed.kg.M": "175.0", "sold.F": "424", "sold.M": "144"}
	day18 := exampleCell{"alive.F": "12561", "alive.M": "794", "deaths.F": "3", "deaths.M": "0", "egg.total": "5914",
		"egg.layPct": "41.08", "egg.guide": "57.6", "egg.inc": "5308", "egg.incPct": "89.75", "egg.jos": "158", "egg.josPct": "2.67",
		"feed.kg.F": "2303.2", "feed.kg.M": "154.7", "sold.F": "1831", "sold.M": "289"}
	week58 := exampleCell{"alive.F": "14826", "alive.M": "1227", "deaths.F": "36", "deaths.M": "8", "egg.total": "47512",
		"egg.layPct": "30.38", "egg.guide": "58.8", "egg.inc": "45745", "egg.incPct": "96.28", "egg.jos": "1221", "egg.josPct": "2.57",
		"feed.kg.F": "22230.9", "feed.kg.M": "1682.7", "sold.F": "10546", "sold.M": "749"}
	totals := exampleCell{"alive.F": "12561", "alive.M": "794", "deaths.F": "534", "deaths.M": "77", "egg.total": "226472",
		"egg.layPct": "41.55", "egg.guide": "59.9", "sold.F": "12801", "sold.M": "1182", "feed.kg.F": "78898.6", "feed.kg.M": "5996.6"}

	if got := weekSapts(res); !slices.Equal(got, []int{59, 58, 57, 56}) {
		t.Fatalf("weeks %v, want SAPT 59, 58, 57, 56", got)
	}
	w59, w58 := res.Weeks[0], res.Weeks[1]
	checkCells(t, "SAPT 59", w59.Cells, week59)
	checkCells(t, "SAPT 58", w58.Cells, week58)
	if len(w59.Days) != 2 {
		t.Fatalf("SAPT 59 has %d days, want 17.11 and 18.11", len(w59.Days))
	}
	checkCells(t, "17.11", w59.Days[0].Cells, day17)
	checkCells(t, "18.11", w59.Days[1].Cells, day18)
	if res.Totals == nil {
		t.Fatal("no totals")
	}
	checkCells(t, "totals", res.Totals.Cells, totals)

	type period struct {
		Kind, Label      string
		SaptFrom, SaptTo int
		From, To         string
		WeeksTotal       int
	}
	type head struct {
		Period      period
		W59, W58    string
		Days59      [2]string
		Status59    map[string]string
		Loaded      periodtable.Loaded
		More        periodtable.More
		Start       *periodtable.StartNote
		TotalsLabel string
		Notes       []string
	}
	got := head{
		Period: period{Kind: res.Period.Kind, Label: res.Period.Label, SaptFrom: res.Period.SaptFrom, SaptTo: res.Period.SaptTo,
			From: res.Period.From.String(), To: res.Period.To.String(), WeeksTotal: res.Period.WeeksTotal},
		W59: w59.Label, W58: w58.Label,
		Days59:   [2]string{w59.Days[0].Weekday, w59.Days[1].Weekday},
		Status59: w59.Status, Start: res.Start, TotalsLabel: res.Totals.Label, Notes: res.Notes,
	}
	if res.Loaded != nil {
		got.Loaded = *res.Loaded
	}
	if res.More != nil {
		got.More = *res.More
	}
	want := head{
		Period: period{Kind: "cycle", Label: "Ciclul curent", SaptFrom: 20, SaptTo: 59,
			From: "2026-02-22", To: "2026-11-18", WeeksTotal: 40},
		W59: "SAPT 59 · 17.11–18.11 · 2 din 7 zile", W58: "SAPT 58 · 10.11–16.11",
		Days59:      [2]string{"Ma", "Mi"},
		Status59:    map[string]string{"egg.layPct": "abatere"},
		Loaded:      periodtable.Loaded{SaptFrom: 56, SaptTo: 59, Weeks: 4},
		More:        periodtable.More{Before: 56, Remaining: 36, Label: "Încarcă încă 4 săptămâni · SAPT 56–59 · 4 din 40"},
		TotalsLabel: "Total SAPT 56–59",
		Notes:       []string{"Morți include sacrificările, ca în registru."},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("first page (-§7.4 example +got):\n%s", diff)
	}
}

func weekSapts(res periodtable.TableResponse) []int {
	out := make([]int, 0, len(res.Weeks))
	for _, w := range res.Weeks {
		out = append(out, w.Sapt)
	}
	return out
}

// checkCells compares the example's cells with got, each rounded to the
// decimals the example writes.
func checkCells(t *testing.T, row string, got periodtable.Cells, want exampleCell) {
	t.Helper()
	gotRounded := map[string]string{}
	for key, w := range want {
		places := int32(0)
		if i := strings.IndexByte(w, '.'); i >= 0 {
			places = int32(len(w) - i - 1)
		}
		v, ok := got[key]
		switch {
		case !ok:
			gotRounded[key] = "(absent)"
		case v == nil:
			gotRounded[key] = "null"
		default:
			gotRounded[key] = num.Round(v.Decimal(), places).StringFixed(places)
		}
	}
	if diff := cmp.Diff(want, gotRounded); diff != "" {
		t.Errorf("%s cells (-§7.4 example +got):\n%s", row, diff)
	}
}

// nullContainers returns the JSON path of every slice or map in v that the
// encoder writes as null (§7.0a: lists are [] when empty, never null). It
// walks v by reflection, following the json tags: a field left out by
// omitzero/omitempty is not written, an embedded struct's fields are the
// parent's, and a type with its own JSON or text encoding is a leaf. A nil
// pointer is a JSON null value, not a list, so it is allowed.
func nullContainers(v any) []string {
	var out []string
	var walk func(v reflect.Value, path string)
	walk = func(v reflect.Value, path string) {
		if !v.IsValid() {
			return
		}
		if leaf(v.Type()) {
			return
		}
		switch v.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !v.IsNil() {
				walk(v.Elem(), path)
			}
		case reflect.Struct:
			for i := range v.NumField() {
				sf := v.Type().Field(i)
				if !sf.IsExported() {
					continue
				}
				name, opts, _ := strings.Cut(sf.Tag.Get("json"), ",")
				if name == "-" && opts == "" {
					continue
				}
				fv := v.Field(i)
				if omitted(fv, opts) {
					continue
				}
				if sf.Anonymous && name == "" {
					walk(fv, path)
					continue
				}
				if name == "" {
					name = sf.Name
				}
				walk(fv, joinPath(path, name))
			}
		case reflect.Slice:
			if v.IsNil() {
				out = append(out, path)
				return
			}
			fallthrough
		case reflect.Array:
			for i := range v.Len() {
				walk(v.Index(i), fmt.Sprintf("%s[%d]", path, i))
			}
		case reflect.Map:
			if v.IsNil() {
				out = append(out, path)
				return
			}
			keys := v.MapKeys()
			slices.SortFunc(keys, func(a, b reflect.Value) int { return strings.Compare(fmt.Sprint(a), fmt.Sprint(b)) })
			for _, k := range keys {
				walk(v.MapIndex(k), joinPath(path, fmt.Sprint(k)))
			}
		}
	}
	walk(reflect.ValueOf(v), "")
	slices.Sort(out)
	return out
}

var (
	jsonMarshaler = reflect.TypeFor[json.Marshaler]()
	textMarshaler = reflect.TypeFor[encoding.TextMarshaler]()
	zeroer        = reflect.TypeFor[interface{ IsZero() bool }]()
)

// leaf reports a type encoded by its own MarshalJSON or MarshalText (num.Num,
// timex.Day, uuid.UUID, …).
func leaf(t reflect.Type) bool {
	if t.Kind() == reflect.Interface {
		return false
	}
	return t.Implements(jsonMarshaler) || t.Implements(textMarshaler) ||
		reflect.PointerTo(t).Implements(jsonMarshaler) || reflect.PointerTo(t).Implements(textMarshaler)
}

// omitted reports a field encoding/json leaves out under its tag options.
func omitted(v reflect.Value, opts string) bool {
	for _, o := range strings.Split(opts, ",") {
		switch o {
		case "omitzero":
			if v.Type().Implements(zeroer) && (v.Kind() != reflect.Pointer || !v.IsNil()) {
				if v.Interface().(interface{ IsZero() bool }).IsZero() {
					return true
				}
			} else if v.IsZero() {
				return true
			}
		case "omitempty":
			switch v.Kind() {
			case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
				if v.Len() == 0 {
					return true
				}
			case reflect.Interface, reflect.Pointer:
				if v.IsNil() {
					return true
				}
			default:
				if v.IsZero() {
					return true
				}
			}
		}
	}
	return false
}

func joinPath(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}

// priceKey matches a field that carries money: lei…, …Lei (totalLei),
// price… (pricePerKg, pricedBy), …cost… (D2, D64: never sent to a viewer).
var priceKey = regexp.MustCompile(`(?i)^(lei|price)|lei$|cost`)

// priceFields returns the JSON path of every object key in body that names a
// price, and of every string value that holds the "fără preț" tag (D64: a
// viewer sees the unpriced count only).
func priceFields(body []byte) ([]string, error) {
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, err
	}
	var out []string
	var walk func(v any, path string)
	walk = func(v any, path string) {
		switch v := v.(type) {
		case map[string]any:
			for k, e := range v {
				p := joinPath(path, k)
				if priceKey.MatchString(k) {
					out = append(out, p)
				}
				walk(e, p)
			}
		case []any:
			for i, e := range v {
				walk(e, fmt.Sprintf("%s[%d]", path, i))
			}
		case string:
			if strings.Contains(v, messages.Unpriced()) {
				out = append(out, path+" = "+strconv.Quote(v))
			}
		}
	}
	walk(v, "")
	slices.Sort(out)
	return out, nil
}

// TestNullContainersCatchesNullLists shows the walker reports each null list
// the encoder would write and nothing else.
func TestNullContainersCatchesNullLists(t *testing.T) {
	type inner struct {
		Items []string `json:"items"`
	}
	type Embedded struct {
		Codes []string `json:"codes"`
	}
	type doc struct {
		Embedded
		List      []int               `json:"list"`
		Empty     []int               `json:"empty"`
		Omitted   []int               `json:"omitted,omitzero"`
		OmitEmpty []int               `json:"omitEmpty,omitempty"`
		Map       map[string]*num.Num `json:"map"`
		NilMap    map[string]string   `json:"nilMap"`
		Ptr       *inner              `json:"ptr"`
		Any       []any               `json:"any"`
		Num       num.Num             `json:"num"`
		ID        uuid.UUID           `json:"id"`
		Skipped   []int               `json:"-"`
	}
	v := doc{
		Empty: []int{},
		Map:   map[string]*num.Num{"a": nil},
		Any:   []any{inner{}, inner{Items: []string{}}, farms.HallTile[farms.TileSection]{Hall: "H5"}},
	}
	want := []string{"any[0].items", "any[2].sections", "codes", "list", "nilMap"}
	if got := nullContainers(v); !slices.Equal(got, want) {
		t.Errorf("nullContainers = %q, want %q", got, want)
	}
}

// TestPriceFieldsCatchesPrices shows the walker finds price fields and the
// "fără preț" tag at any depth and ignores the rest.
func TestPriceFieldsCatchesPrices(t *testing.T) {
	body := `{"lei": 1, "feed": {"leiPerKg": 2, "kg": 3}, "rows": [{"priceKg": 4}, {"totalCost": 5},
		{"pricePerKg": null, "totalLei": null, "pricedBy": "din registru", "tag": "fără preț"}],
		"cost": {"cycles": []}, "deleted": 6, "label": "lei", "unpriced": {"count": 4}, "notes": ["3 livrări fără preț"]}`
	got, err := priceFields([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"cost", "feed.leiPerKg", "lei", `notes[0] = "3 livrări fără preț"`, "rows[0].priceKg", "rows[1].totalCost",
		"rows[2].pricePerKg", "rows[2].pricedBy", `rows[2].tag = "fără preț"`, "rows[2].totalLei"}
	if !slices.Equal(got, want) {
		t.Errorf("priceFields = %q, want %q", got, want)
	}
}

// TestNormaliseReplacesIDs shows every id becomes its token, in the indented
// httpx.JSON encoding, a Livrări cursor its decoded fields with the id
// tokenised, and an undo token a fixed placeholder.
func TestNormaliseReplacesIDs(t *testing.T) {
	known := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	delivery := uuid.MustParse("66666666-7777-8888-9999-000000000000")
	cursor := reports.Cursor{Date: day(2026, time.October, 2), FarmKey: "hirbovat", TypeKey: "cocosi", ID: delivery}.String()
	got := normalise(t, map[string]any{
		"farmId": known, "list": []uuid.UUID{known},
		"deliveries": map[string]any{"nextCursor": cursor}, "undoToken": "c2lnbmVk.dG9rZW4",
	}, map[uuid.UUID]string{known: "<farm:hirbovat>", delivery: "<delivery:hirbovat/2026-10-02/cocosi>"})
	want := `{
  "deliveries": {
    "nextCursor": "<cursor:2026-10-02/hirbovat/cocosi/<delivery:hirbovat/2026-10-02/cocosi>>"
  },
  "farmId": "<farm:hirbovat>",
  "list": [
    "<farm:hirbovat>"
  ],
  "undoToken": "<undoToken>"
}
`
	if string(got) != want {
		t.Errorf("normalise:\n%s\nwant\n%s", got, want)
	}
}

// consumFileRow is a week, day, section week or totals row of a committed
// table golden file, as the client reads it.
type consumFileRow struct {
	Code     string               `json:"code"`
	Sapt     int                  `json:"sapt"`
	Date     string               `json:"date"`
	Label    string               `json:"label"`
	Cells    map[string]*num.Num  `json:"cells"`
	Status   map[string]string    `json:"status"`
	Reasons  map[string]string    `json:"reasons"`
	Missing  *periodtable.Missing `json:"missing"`
	Days     []consumFileRow      `json:"days"`
	Sections []consumFileRow      `json:"sections"`
}

type consumFile struct {
	Loaded periodtable.Loaded `json:"loaded"`
	Weeks  []consumFileRow    `json:"weeks"`
	Totals consumFileRow      `json:"totals"`
	Notes  []string           `json:"notes"`
}

func readTableFile(t *testing.T, name string) consumFile {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "periodtable", name))
	if err != nil {
		t.Fatal(err)
	}
	var f consumFile
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return f
}

// boardFigures writes a Consum row at the S11 board's precision, as T7's
// seeded test does: kg and g at 1 decimal, the guide whole, the deviation at
// 1 decimal, the status of g per sex and the missing keys.
func (r consumFileRow) boardFigures() string {
	at := func(key string, places int32) string {
		v, ok := r.Cells[key]
		switch {
		case !ok:
			return "absent"
		case v == nil:
			return "–"
		}
		return num.Round(v.Decimal(), places).StringFixed(places)
	}
	s := fmt.Sprintf("kg %s · %s, g %s · %s, ghid %s · %s, %s%% · %s%%, %s · %s",
		at(periodtable.KeyFeedKgF, 1), at(periodtable.KeyFeedKgM, 1), at(periodtable.KeyFeedGF, 1), at(periodtable.KeyFeedGM, 1),
		at(periodtable.KeyFeedGGuideF, 0), at(periodtable.KeyFeedGGuideM, 0),
		at(periodtable.KeyFeedGDiffF, 1), at(periodtable.KeyFeedGDiffM, 1),
		r.Status[periodtable.KeyFeedGF], r.Status[periodtable.KeyFeedGM])
	if r.Missing != nil {
		var keys []string
		for _, k := range r.Missing.Keys {
			keys = append(keys, k.Code+" "+string(k.Sex)+" "+k.Date.String())
		}
		s += fmt.Sprintf(", lipsă %d: %s", r.Missing.Rows, strings.Join(keys, ", "))
	}
	return s
}

// TestConsumGoldenFileMatchesSeededValues (M3 T9 criterion 3): the committed
// page 1 of Hîrbovăț's Consum (table_report_feed_cycle.json) carries T7's
// seeded values (M3 T7 criterion 4, at the board's precision), and its kg
// agree at full precision with the committed Rezumat page 1
// (table_farm_rezumat_cycle.json): the two views read one §5.7 kg.
func TestConsumGoldenFileMatchesSeededValues(t *testing.T) {
	consum := readTableFile(t, "table_report_feed_cycle.json")
	if consum.Loaded != (periodtable.Loaded{SaptFrom: 56, SaptTo: 59, Weeks: 4}) {
		t.Fatalf("loaded %+v, want SAPT 56–59", consum.Loaded)
	}
	got := map[string]string{"totals": consum.Totals.Label + ": " + consum.Totals.boardFigures()}
	for _, w := range consum.Weeks {
		got[fmt.Sprint("SAPT ", w.Sapt)] = w.boardFigures()
		for _, d := range w.Days {
			got[d.Date] = d.boardFigures()
		}
		for _, s := range w.Sections {
			got[fmt.Sprint(s.Code, " SAPT ", w.Sapt)] = s.boardFigures()
		}
	}
	for key, want := range map[string]string{
		"SAPT 59":      "kg 4675.3 · 329.7, g 160.0 · 142.7, ghid 157 · 156, 1.9% · -8.5%, in_norma · in_norma",
		"2026-11-18":   "kg 2303.2 · 154.7, g 160.0 · 142.8, ghid 157 · 156, 1.9% · -8.4%, in_norma · in_norma",
		"H5.1 SAPT 59": "kg 1364.4 · 98.4, g 160.0 · 143.9, ghid 157 · 156, 1.9% · -7.8%, in_norma · in_norma",
		"SAPT 58": "kg 22230.9 · 1682.7, g 155.9 · 142.2, ghid 158 · 155, -1.3% · -8.3%, in_norma · in_norma, " +
			"lipsă 2: H7.1 F 2026-11-13, H7.2 F 2026-11-13",
		"SAPT 57": "kg 23284.2 · 1970.6, g 130.5 · 140.8, ghid 158 · 154, -17.4% · -8.6%, abatere · in_norma",
		"totals": "Total SAPT 56–59: kg 78898.6 · 5996.6, g 148.5 · 141.3, ghid 158 · 154, -6.0% · -8.3%, in_norma · in_norma, " +
			"lipsă 2: H7.1 F 2026-11-13, H7.2 F 2026-11-13",
	} {
		if got[key] != want {
			t.Errorf("%s:\n got %s\nwant %s", key, got[key], want)
		}
	}

	// kg at full precision, row by row, against the Rezumat file.
	rezumat := readTableFile(t, "table_farm_rezumat_cycle.json")
	kg := func(r consumFileRow) string {
		out := make([]string, 0, 2)
		for _, key := range []string{periodtable.KeyFeedKgF, periodtable.KeyFeedKgM} {
			v := r.Cells[key]
			if v == nil || v.Decimal().IsZero() { // Consum "–" is the Rezumat's 0 kg
				out = append(out, "0")
				continue
			}
			out = append(out, v.Decimal().String())
		}
		return strings.Join(out, " · ")
	}
	figures := func(f consumFile) map[string]string {
		out := map[string]string{"totals": kg(f.Totals)}
		for _, w := range f.Weeks {
			out[fmt.Sprint("SAPT ", w.Sapt)] = kg(w)
			for _, d := range w.Days {
				out[d.Date] = kg(d)
			}
		}
		return out
	}
	consumKg, rezumatKg := figures(consum), figures(rezumat)
	if n := len(consumKg); n != 1+4+7+7+7+2 { // totals, SAPT 56–59, their days (SAPT 59: 17–18.11)
		t.Errorf("Consum page 1 has %d totals, week and day rows, want 28", n)
	}
	if diff := cmp.Diff(rezumatKg, consumKg); diff != "" {
		t.Errorf("kg of the Consum file vs the Rezumat file (-Rezumat +Consum):\n%s", diff)
	}
}
