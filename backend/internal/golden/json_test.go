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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/farms"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/golden"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/periodtable"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// The JSON golden files (SPEC §11.8, M2 T9): the PeriodTable, farm page,
// summary and bootstrap responses on the dev seed as of 18.11.2026 12:00,
// built by the builders the handlers call (no HTTP: route wiring is
// cmd/api's), encoded like httpx.JSON and indented. Ids are replaced by
// stable tokens (M2 D9), because the import's ids are random.
//
// Regenerate with: go test ./internal/golden -run TestJSONGolden -update

var update = flag.Bool("update", false, "rewrite the JSON golden files under testdata/")

// fixture is the seeded database with the ids the cases need.
type fixture struct {
	db                dbq.DBTX
	hirbovat, telita  uuid.UUID
	seria2025Hirbovat uuid.UUID // Hîrbovăț's Seria 2025 cycle (no flock count)
	seria2026Telita   uuid.UUID // Telița's Seria 2026 cycle (growth)
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
// <section:hirbovat/H5.1>, <guide:…>, <feedType:…>.
func idTokens(t *testing.T, ctx context.Context, db dbq.DBTX) map[uuid.UUID]string {
	t.Helper()
	rows, err := db.Query(ctx, `
select id, 'farm:' || name_key from farm
union all select id, 'series:' || name_key from series
union all select c.id, 'cycle:' || f.name_key || '/' || s.year
  from cycle c join farm f on f.id = c.farm_id join series s on s.id = c.series_id
union all select s.id, 'section:' || f.name_key || '/' || s.code from section s join farm f on f.id = s.farm_id
union all select id, 'guide:' || name_key from guide
union all select id, 'feedType:' || name_key from feed_type`)
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

var uuidText = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

// normalise encodes v like httpx.JSON, indents it and replaces every id with
// its token. An id without a token fails the test.
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
	var unknown []string
	text := uuidText.ReplaceAllStringFunc(out.String(), func(s string) string {
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
}

// TestViewerResponsesHaveNoPricesOrNullLists builds every M2 response as a
// viewer of Hîrbovăț sees it and walks it: no lei, price or cost field (D64,
// §3.3) and no null list (§7.0a).
func TestViewerResponsesHaveNoPricesOrNullLists(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	viewer := f.viewer()
	responses := map[string]func(t *testing.T) (any, error){
		"page":      func(*testing.T) (any, error) { return farms.BuildPage(ctx, f.db, f.input(viewer, f.hirbovat)) },
		"summary":   func(*testing.T) (any, error) { return farms.BuildSummary(ctx, f.db, f.input(viewer, f.hirbovat)) },
		"bootstrap": func(*testing.T) (any, error) { return farms.BuildBootstrap(ctx, f.db, f.input(viewer, uuid.Nil)) },
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
			keys, err := priceKeys(b)
			if err != nil {
				t.Fatal(err)
			}
			if len(keys) > 0 {
				t.Errorf("a viewer response carries price fields %q", keys)
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

// priceKey matches a field that carries money: lei…, price…, …cost… (D64:
// never sent to a viewer).
var priceKey = regexp.MustCompile(`(?i)^(lei|price)|cost`)

// priceKeys returns the JSON path of every object key in body that names a
// price.
func priceKeys(body []byte) ([]string, error) {
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

// TestPriceKeysCatchesPrices shows the walker finds price fields at any depth
// and ignores the rest.
func TestPriceKeysCatchesPrices(t *testing.T) {
	body := `{"lei": 1, "feed": {"leiPerKg": 2, "kg": 3}, "rows": [{"priceKg": 4}, {"totalCost": 5}], "deleted": 6, "label": "lei"}`
	got, err := priceKeys([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"feed.leiPerKg", "lei", "rows[0].priceKg", "rows[1].totalCost"}
	if !slices.Equal(got, want) {
		t.Errorf("priceKeys = %q, want %q", got, want)
	}
}

// TestNormaliseReplacesIDs shows every id becomes its token, in the indented
// httpx.JSON encoding.
func TestNormaliseReplacesIDs(t *testing.T) {
	known := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	got := normalise(t, map[string]any{"farmId": known, "list": []uuid.UUID{known}}, map[uuid.UUID]string{known: "<farm:hirbovat>"})
	want := "{\n  \"farmId\": \"<farm:hirbovat>\",\n  \"list\": [\n    \"<farm:hirbovat>\"\n  ]\n}\n"
	if string(got) != want {
		t.Errorf("normalise:\n%s\nwant\n%s", got, want)
	}
}
