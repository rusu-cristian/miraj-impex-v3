package entries

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// The delivery tests run on feedFixture: Hîrbovăț (farm A, Seria T open
// from 01.11.2026) has Breeder 2 1 000 kg and Cocoși 100 kg delivered on
// 01.11 and no allocation; Breeder 1 is retired; Telița (farm B) sits in the
// closed Seria 2025 from 01.10 to 10.11.2026. ion and maria manage both
// farms, outsider only Telița.

// postDelivery sends body to POST /api/entries/delivery as p, through Routes.
func (f *feedFixture) postDelivery(p auth.Principal, body string) *httptest.ResponseRecorder {
	f.t.Helper()
	mux := http.NewServeMux()
	NewHandler(f.svc).Routes(mux, httpx.NewChain())
	req := httptest.NewRequestWithContext(auth.WithPrincipal(context.Background(), p), http.MethodPost,
		"/api/entries/delivery", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

// deliveryBody is a delivery of farm A on day as JSON, with extra members
// merged in (a nil value removes the member).
func (f *feedFixture) deliveryBody(day timex.Day, feedType uuid.UUID, qty any, extra map[string]any) string {
	f.t.Helper()
	m := map[string]any{"farmId": f.farmA, "date": day, "feedTypeId": feedType, "qtyKg": qty}
	for k, v := range extra {
		if v == nil {
			delete(m, k)
			continue
		}
		m[k] = v
	}
	b, err := json.Marshal(m)
	if err != nil {
		f.t.Fatal(err)
	}
	return string(b)
}

func (f *feedFixture) delivery(day timex.Day, feedType uuid.UUID, qty string, key *uuid.UUID) DeliveryCreateRequest {
	n := num.MustParse(qty)
	return DeliveryCreateRequest{
		WriteMeta: httpx.WriteMeta{ClientRequestID: key}, FarmID: f.farmA, Date: day, FeedTypeID: feedType, QtyKg: &n,
	}
}

func envelopeOf(t *testing.T, w *httptest.ResponseRecorder) httpx.APIError {
	t.Helper()
	var b httpx.ErrorBody
	if err := json.Unmarshal(w.Body.Bytes(), &b); err != nil {
		t.Fatalf("%d %s: %v", w.Code, w.Body, err)
	}
	return b.Error
}

// T5 criterion 2: 201 {entry, toast, stockAfter, warnings [], asOf}; the
// stock is the type's at the end of the delivery date (Q13), not today's;
// the row has no price; one audit row, entry.created + delivery.created; the
// same key again answers 200 with the same bytes and saves nothing.
func TestCreateDeliverySaves(t *testing.T) {
	ctx := context.Background()
	f := newFeedFixture(t)
	// 18.11 consumes 266,4 kg Breeder 2: the stock at the end of 17.11 doesn't see it.
	if _, err := f.save(f.ion, f.request(nov18, nil, f.allRows()...)); err != nil {
		t.Fatal(err)
	}
	key := uuid.New()
	body := f.deliveryBody(nov17, f.breeder2, 9000, map[string]any{"clientRequestId": key, "note": "  factura 12 "})
	first := f.postDelivery(f.ion, body)
	if first.Code != http.StatusCreated {
		t.Fatalf("POST = %d %s", first.Code, first.Body)
	}
	var resp DeliveryResponse
	if err := json.Unmarshal(first.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	want := `{"entry":{"id":"` + resp.Entry.ID.String() + `","kind":"delivery","version":"2026-11-18T10:00:00.000000Z"},` +
		"\"toast\":\"Livrare adăugată · 9 000 kg Breeder 2 · stoc Hîrbovăț: 10 000 kg · fără preț\"," +
		`"stockAfter":10000,"warnings":[],"asOf":"2026-11-18T12:00:00+02:00"}`
	if got := strings.TrimSpace(first.Body.String()); got != want {
		t.Errorf("body\n got %s\nwant %s", got, want)
	}

	var (
		qty, note                 string
		farm, feedType            uuid.UUID
		date                      timex.Day
		priced, fromRegister, del bool
	)
	if err := f.d.App.QueryRow(ctx, `select farm_id, feed_type_id, date, qty_kg::text, note,
		(price_per_kg is not null or total_lei is not null or price_input is not null or priced_by is not null
		 or priced_at is not null), price_from_register, deleted_at is not null
		from feed_delivery where entry_id = $1`, resp.Entry.ID).
		Scan(&farm, &feedType, &date, &qty, &note, &priced, &fromRegister, &del); err != nil {
		t.Fatal(err)
	}
	if farm != f.farmA || feedType != f.breeder2 || date != nov17 || qty != "9000.0" || note != "factura 12" ||
		priced || fromRegister || del {
		t.Errorf("row = %s %s %s %s %q priced %v register %v deleted %v", farm, feedType, date, qty, note, priced, fromRegister, del)
	}

	var (
		action, entity, summary string
		recordTypes             []string
		farmIDs                 []uuid.UUID
		recordDate              timex.Day
		newValues               string
	)
	if err := f.d.App.QueryRow(ctx, `select action::text, entity, summary, record_types, farm_ids, record_date,
		new_values::text from audit_log where entry_id = $1`, resp.Entry.ID).
		Scan(&action, &entity, &summary, &recordTypes, &farmIDs, &recordDate, &newValues); err != nil {
		t.Fatal(err)
	}
	if action != "create" || entity != "entry" || summary != "Livrare · 9 000 kg Breeder 2" ||
		!slices.Equal(recordTypes, []string{"livrare"}) || !slices.Equal(farmIDs, []uuid.UUID{f.farmA}) || recordDate != nov17 {
		t.Errorf("audit = %s %s %q %v %v %s", action, entity, summary, recordTypes, farmIDs, recordDate)
	}
	var values deliveryAuditValues
	if err := json.Unmarshal([]byte(newValues), &values); err != nil {
		t.Fatal(err)
	}
	if values.FeedTypeID != f.breeder2 || values.Type != "Breeder 2" || values.Date != nov17 || values.Kg.String() != "9000" ||
		values.Note == nil || *values.Note != "factura 12" {
		t.Errorf("new_values = %s", newValues)
	}

	rows, err := f.d.App.Query(ctx, `select type, payload::text from live_event where payload->>'entryId' = $1::text order by id`,
		resp.Entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	type event struct{ Type, Payload string }
	events, err := pgx.CollectRows(rows, pgx.RowToStructByPos[event])
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Type != "entry.created" || events[1].Type != "delivery.created" {
		t.Fatalf("live events = %+v", events)
	}
	var created struct {
		Kind        string   `json:"kind"`
		Sections    []string `json:"sections"`
		RecordTypes []string `json:"recordTypes"`
		Dates       []string `json:"dates"`
		Author      string   `json:"author"`
		Summary     string   `json:"summary"`
	}
	if err := json.Unmarshal([]byte(events[0].Payload), &created); err != nil {
		t.Fatal(err)
	}
	if created.Kind != "delivery" || created.Sections == nil || len(created.Sections) != 0 ||
		!slices.Equal(created.RecordTypes, []string{"livrare"}) || !slices.Equal(created.Dates, []string{"2026-11-17"}) ||
		created.Author != "Ion Rusu" || created.Summary != summary {
		t.Errorf("entry.created = %s", events[0].Payload)
	}
	wantDelivery := `{"entryId":"` + resp.Entry.ID.String() + `","farmId":"` + f.farmA.String() +
		`","feedType":"Breeder 2","kg":9000,"date":"2026-11-17"}`
	if !jsonEqual(t, events[1].Payload, wantDelivery) {
		t.Errorf("delivery.created = %s, want %s", events[1].Payload, wantDelivery)
	}

	before := f.rows()
	again := f.postDelivery(f.ion, body)
	if again.Code != http.StatusOK || !bytes.Equal(again.Body.Bytes(), first.Body.Bytes()) {
		t.Errorf("replay = %d %s, want 200 with the first body", again.Code, again.Body)
	}
	if f.rows() != before || f.count("feed_delivery") != 3 {
		t.Errorf("the replay wrote: %v → %v, %d deliveries", before, f.rows(), f.count("feed_delivery"))
	}
}

// jsonEqual compares two JSON texts as values.
func jsonEqual(t *testing.T, a, b string) bool {
	t.Helper()
	var va, vb any
	if err := json.Unmarshal([]byte(a), &va); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(b), &vb); err != nil {
		t.Fatal(err)
	}
	ja, _ := json.Marshal(va)
	jb, _ := json.Marshal(vb)
	return bytes.Equal(ja, jb)
}

// Two deliveries of one type on one day are legal (no uniqueness key); the
// second's stock includes the first. A blank note is stored as null; a
// decimal kg shows with its decimal.
func TestCreateDeliveryTwiceSameDay(t *testing.T) {
	ctx := context.Background()
	f := newFeedFixture(t)
	if res, err := f.svc.CreateDelivery(ctx, f.ion, f.delivery(nov18, f.cocosi, "50", nil)); err != nil || res.Status != http.StatusCreated {
		t.Fatalf("first: %d %v", res.Status, err)
	}
	req := f.delivery(nov18, f.cocosi, "20.5", nil)
	blank := "   "
	req.Note = &blank
	res, err := f.svc.CreateDelivery(ctx, f.maria, req)
	if err != nil || res.Status != http.StatusCreated {
		t.Fatalf("second: %d %v", res.Status, err)
	}
	var resp DeliveryResponse
	if err := json.Unmarshal(res.Body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.StockAfter.String() != "170.5" || resp.Toast != "Livrare adăugată · 20,5 kg Cocoși · stoc Hîrbovăț: 171 kg · fără preț" {
		t.Errorf("second = %s %q", resp.StockAfter, resp.Toast)
	}
	var notes int
	if err := f.d.App.QueryRow(ctx, `select count(*) from feed_delivery where date = $1 and note is null`, nov18).Scan(&notes); err != nil || notes != 2 {
		t.Errorf("deliveries of 18.11 without a note = %d (%v), want 2", notes, err)
	}
}

// T5 criterion 2: price fields (Q9). A manager's request with pricePerKg or
// totalLei, any value and null too, is 403 PRICE_FORBIDDEN; the owner's is
// 400 unknown_field on each field. The role comes first (a viewer's is
// FORBIDDEN). Nothing is written.
func TestCreateDeliveryPriceFields(t *testing.T) {
	f := newFeedFixture(t)
	before := f.rows()
	for _, tc := range []struct {
		name   string
		p      auth.Principal
		extra  map[string]any
		status int
		code   string
		paths  []string
	}{
		{"manager, pricePerKg", f.ion, map[string]any{"pricePerKg": 7.76}, 403, "PRICE_FORBIDDEN", nil},
		{"manager, null pricePerKg", f.ion, map[string]any{"pricePerKg": json.RawMessage("null")}, 403, "PRICE_FORBIDDEN", nil},
		{"manager, totalLei", f.ion, map[string]any{"totalLei": "199000"}, 403, "PRICE_FORBIDDEN", nil},
		{"manager, invalid qty and a price", f.ion, map[string]any{"pricePerKg": 1, "qtyKg": 0}, 403, "PRICE_FORBIDDEN", nil},
		{"owner, pricePerKg", f.owner, map[string]any{"pricePerKg": 7.76}, 400, "VALIDATION", []string{"pricePerKg"}},
		{"owner, both null", f.owner, map[string]any{"pricePerKg": json.RawMessage("null"), "totalLei": json.RawMessage("null")},
			400, "VALIDATION", []string{"pricePerKg", "totalLei"}},
		{"viewer, pricePerKg", f.viewer, map[string]any{"pricePerKg": 7.76}, 403, "FORBIDDEN", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := f.postDelivery(tc.p, f.deliveryBody(nov18, f.breeder2, 9000, tc.extra))
			e := envelopeOf(t, w)
			if w.Code != tc.status || e.Code != tc.code {
				t.Fatalf("got %d %s %q, want %d %s", w.Code, e.Code, e.Message, tc.status, tc.code)
			}
			if tc.code == "PRICE_FORBIDDEN" && e.Message != "Prețul îl adaugă proprietarul în Prețuri hrană." {
				t.Errorf("message = %q", e.Message)
			}
			var paths []string
			for _, is := range e.Issues {
				paths = append(paths, is.Path)
				if is.Code != httpx.IssueUnknownField || is.Message != "Datele trimise nu sunt valide." {
					t.Errorf("issue = %+v", is)
				}
			}
			if !slices.Equal(paths, tc.paths) {
				t.Errorf("issue paths = %v, want %v", paths, tc.paths)
			}
		})
	}
	// The service checks them too.
	req := f.delivery(nov18, f.breeder2, "9000", nil)
	req.TotalLei = json.RawMessage("1")
	_, err := f.svc.CreateDelivery(context.Background(), f.ion, req)
	wantCode(t, err, http.StatusForbidden, httpx.CodePriceForbidden)
	if f.rows() != before {
		t.Errorf("rows %v → %v", before, f.rows())
	}
}

// T5 criterion 2: the handler's level-1 checks and access, in its order
// (decode → role → farm → price → names → Validate).
func TestCreateDeliveryHandler(t *testing.T) {
	f := newFeedFixture(t)
	const qtyMsg = "Scrie câte kg de Breeder 2 s-au livrat la Hîrbovăț."
	const invalid = "Datele trimise nu sunt valide."
	before := f.rows()
	for _, tc := range []struct {
		name   string
		p      auth.Principal
		body   string
		status int
		code   string
		issues []httpx.Issue
	}{
		{"qtyKg missing", f.ion, f.deliveryBody(nov18, f.breeder2, nil, map[string]any{"qtyKg": nil}), 400, "VALIDATION",
			[]httpx.Issue{{Path: "qtyKg", Code: "required", Message: qtyMsg}}},
		{"qtyKg null", f.ion, f.deliveryBody(nov18, f.breeder2, nil, nil), 400, "VALIDATION",
			[]httpx.Issue{{Path: "qtyKg", Code: "required", Message: qtyMsg}}},
		{"qtyKg 0", f.ion, f.deliveryBody(nov18, f.breeder2, 0, nil), 400, "VALIDATION",
			[]httpx.Issue{{Path: "qtyKg", Code: "out_of_range", Message: qtyMsg}}},
		{"qtyKg negative", f.ion, f.deliveryBody(nov18, f.breeder2, -5, nil), 400, "VALIDATION",
			[]httpx.Issue{{Path: "qtyKg", Code: "out_of_range", Message: qtyMsg}}},
		{"qtyKg 1,25", f.ion, f.deliveryBody(nov18, f.breeder2, json.Number("1.25"), nil), 400, "VALIDATION",
			[]httpx.Issue{{Path: "qtyKg", Code: "out_of_range", Message: invalid}}},
		{"qtyKg beyond numeric(12,1)", f.ion, f.deliveryBody(nov18, f.breeder2, json.Number("99999999999.96"), nil), 400, "VALIDATION",
			[]httpx.Issue{{Path: "qtyKg", Code: "out_of_range", Message: invalid}}},
		{"qtyKg 1e11", f.ion, f.deliveryBody(nov18, f.breeder2, json.Number("100000000000"), nil), 400, "VALIDATION",
			[]httpx.Issue{{Path: "qtyKg", Code: "out_of_range", Message: invalid}}},
		{"qtyKg string", f.ion, f.deliveryBody(nov18, f.breeder2, "9000", nil), 400, "VALIDATION",
			[]httpx.Issue{{Path: "qtyKg", Code: "bad_type", Message: invalid}}},
		{"note too long", f.ion, f.deliveryBody(nov18, f.breeder2, 10, map[string]any{"note": strings.Repeat("ț", 201)}), 400, "VALIDATION",
			[]httpx.Issue{{Path: "note", Code: "too_long", Message: invalid}}},
		{"unknown type, no names for qty", f.ion, f.deliveryBody(nov18, uuid.New(), 0, nil), 400, "VALIDATION",
			[]httpx.Issue{{Path: "feedTypeId", Code: "out_of_range", Message: invalid}, {Path: "qtyKg", Code: "out_of_range", Message: invalid}}},
		{"required ids and date", f.ion, `{"farmId":"` + f.farmA.String() + `","qtyKg":5}`, 400, "VALIDATION",
			[]httpx.Issue{{Path: "date", Code: "required", Message: invalid}, {Path: "feedTypeId", Code: "required", Message: invalid}}},
		{"no farm", f.ion, `{"date":"2026-11-18","feedTypeId":"` + f.breeder2.String() + `","qtyKg":5}`, 400, "VALIDATION",
			[]httpx.Issue{{Path: "farmId", Code: "required", Message: invalid}}},
		{"unknown field", f.ion, f.deliveryBody(nov18, f.breeder2, 10, map[string]any{"price": 1}), 400, "VALIDATION",
			[]httpx.Issue{{Path: "price", Code: "unknown_field", Message: invalid}}},
		{"bad source", f.ion, f.deliveryBody(nov18, f.breeder2, 10, map[string]any{"source": "fax"}), 400, "VALIDATION",
			[]httpx.Issue{{Path: "source", Code: "out_of_range", Message: invalid}}},
		{"vet", f.vet, f.deliveryBody(nov18, f.breeder2, 10, nil), 403, "FORBIDDEN", nil},
		{"viewer", f.viewer, f.deliveryBody(nov18, f.breeder2, 10, nil), 403, "FORBIDDEN", nil},
		{"another farm, before its names are read", f.outsider, f.deliveryBody(nov18, f.breeder2, 0, nil), 403, "FARM_FORBIDDEN", nil},
		{"owner, unknown farm", f.owner, f.deliveryBody(nov18, f.breeder2, 10, map[string]any{"farmId": uuid.New()}), 404, "NOT_FOUND", nil},
		{"level 2 (future day)", f.ion, f.deliveryBody(nov19, f.breeder2, 10, nil), 422, "REFUSED", nil},
		{"over 64 KiB", f.ion, `{"note":"` + strings.Repeat("x", 65<<10) + `"}`, 413, "PAYLOAD_TOO_LARGE", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := f.postDelivery(tc.p, tc.body)
			e := envelopeOf(t, w)
			if w.Code != tc.status || e.Code != tc.code {
				t.Fatalf("got %d %s %q, want %d %s", w.Code, e.Code, e.Message, tc.status, tc.code)
			}
			if tc.issues != nil && (!slices.Equal(e.Issues, tc.issues) || e.Message != tc.issues[0].Message) {
				t.Errorf("issues = %+v (%q),\nwant %+v", e.Issues, e.Message, tc.issues)
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Errorf("Cache-Control = %q", w.Header().Get("Cache-Control"))
			}
		})
	}
	if f.rows() != before {
		t.Errorf("rows %v → %v", before, f.rows())
	}
	// A 200-character note (in characters, not bytes) and a past day: 201.
	if w := f.postDelivery(f.ion, f.deliveryBody(nov1, f.breeder2, 10, map[string]any{"note": strings.Repeat("ț", 200)})); w.Code != http.StatusCreated {
		t.Errorf("200-character note, past day: %d %s", w.Code, w.Body)
	}
}

// T5 criterion 2: the level-2 refusals, each with its own code and nothing
// written; in a dry run the 422s come back as refusals in a 200.
func TestCreateDeliveryRefusals(t *testing.T) {
	ctx := context.Background()
	f := newFeedFixture(t)
	closedDay := timex.MustDay(2026, time.November, 5)
	closed := f.delivery(closedDay, f.breeder2, "100", nil)
	closed.FarmID = f.farmB
	for _, tc := range []struct {
		name    string
		req     DeliveryCreateRequest
		status  int
		code    string
		message string
		field   string
	}{
		{"future day", f.delivery(nov19, f.breeder2, "100", nil), 422, "REFUSED",
			"19.11.2026 e în viitor. Alege o dată până la 18.11.2026.", "date"},
		{"retired type", f.delivery(nov18, f.breeder1, "100", nil), 422, "REFUSED", "Breeder 1 e retras. Alege alt tip.", "feedTypeId"},
		{"closed series", closed, 409, "SERIES_CLOSED", "Seria 2025 e închisă: nu mai primește înregistrări sau corecții.", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := f.rows()
			_, err := f.svc.CreateDelivery(ctx, f.ion, tc.req)
			status, body := errBody(t, err)
			if status != tc.status || body.Code != tc.code || body.Message != tc.message {
				t.Errorf("got %d %s %q, want %d %s %q", status, body.Code, body.Message, tc.status, tc.code, tc.message)
			}
			if tc.field != "" {
				if want := []httpx.Refusal{{Field: tc.field, Message: tc.message}}; !slices.Equal(body.Refusals, want) {
					t.Errorf("refusals = %+v, want %+v", body.Refusals, want)
				}
				dry := tc.req
				dry.DryRun = true
				res, err := f.svc.CreateDelivery(ctx, f.ion, dry)
				want := `{"effects":[],"warnings":[],"refusals":[{"field":"` + tc.field + `","message":"` + tc.message + `"}]}`
				if err != nil || res.Status != http.StatusOK || strings.TrimSpace(string(res.Body)) != want {
					t.Errorf("dry run = %d %s %v, want 200 %s", res.Status, res.Body, err, want)
				}
			}
			if f.rows() != before {
				t.Errorf("rows %v → %v", before, f.rows())
			}
		})
	}
	// A day after the closed cycle's end, at the empty farm: no series, saved.
	after := closed
	after.Date = timex.MustDay(2026, time.November, 11)
	if res, err := f.svc.CreateDelivery(ctx, f.ion, after); err != nil || res.Status != http.StatusCreated {
		t.Errorf("Telița 11.11: %d %v", res.Status, err)
	}
}

// T5 criterion 2: a dry run of a valid body answers 200 with empty lists and
// writes nothing; its key isn't taken, so the real save with it goes through.
func TestCreateDeliveryDryRun(t *testing.T) {
	f := newFeedFixture(t)
	before := [2]int{f.count("feed_delivery"), f.count("entry")}
	key := uuid.New()
	req := f.delivery(nov18, f.breeder2, "9000", &key)
	req.DryRun = true
	res, err := f.svc.CreateDelivery(context.Background(), f.ion, req)
	if err != nil || res.Status != http.StatusOK || strings.TrimSpace(string(res.Body)) != `{"effects":[],"warnings":[],"refusals":[]}` {
		t.Fatalf("dry run = %d %s %v", res.Status, res.Body, err)
	}
	if after := [2]int{f.count("feed_delivery"), f.count("entry")}; after != before ||
		f.count("audit_log")+f.count("live_event")+f.count("request_replay") != 0 {
		t.Errorf("a dry run wrote: %v → %v", before, after)
	}
	req.DryRun = false
	if res, err := f.svc.CreateDelivery(context.Background(), f.ion, req); err != nil || res.Status != http.StatusCreated {
		t.Fatalf("save after the dry run: %d %v", res.Status, err)
	}
}

// T5 criterion 4: one save with a key is at most 12 statements besides
// begin and commit.
func TestCreateDeliveryStatements(t *testing.T) {
	f := newFeedFixture(t)
	var log statementLog
	pc, err := db.PoolConfig(testdb.DSN(f.d.Name), 1, &testdb.Today)
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
	pc.ConnConfig.Tracer = &log
	pool, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	log.take()

	svc := NewService(db.NewStore(pool), f.d.Clock)
	key := uuid.New()
	if _, err := svc.CreateDelivery(context.Background(), f.ion, f.delivery(nov18, f.breeder2, "9000", &key)); err != nil {
		t.Fatal(err)
	}
	got := log.take()
	want := []string{
		"GetReplay", "?", "LockFarm", "EntryIDByClientRequest", "EntriesDeliverySaveContext", "InsertEntry",
		"EntriesInsertFeedDelivery", "CalcStockSeries", "InsertAuditLog", "EmitLiveEvent", "EmitLiveEvent", "InsertReplay", "?",
	}
	if !slices.Equal(got, want) {
		t.Errorf("statements %q,\nwant %q", got, want)
	}
	if n := len(slices.DeleteFunc(slices.Clone(got), func(s string) bool { return s == "?" })); n > 12 {
		t.Errorf("%d statements besides begin and commit, want at most 12", n)
	}
}
