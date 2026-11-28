package prices_test

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
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/tokens"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/prices"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// seeded is the dev seed (the imported history to 18.11.2026) with the
// owner, the Hîrbovăț farm and the prices routes.
type seeded struct {
	t        *testing.T
	d        *testdb.DB
	mux      *http.ServeMux
	owner    auth.Principal
	hirbovat uuid.UUID
}

func newSeeded(t *testing.T) *seeded {
	t.Helper()
	s := &seeded{t: t, d: testdb.NewSeeded(t)}
	key, err := tokens.DeriveKey(bytes.Repeat([]byte{7}, 32), tokens.PurposeUndo)
	if err != nil {
		t.Fatal(err)
	}
	s.mux = http.NewServeMux()
	prices.NewHandler(prices.NewService(db.NewStore(s.d.App), s.d.Clock, key)).Routes(s.mux, httpx.NewChain())
	if err := s.d.App.QueryRow(t.Context(), `select id, name from app_user where role = 'owner' order by login limit 1`).
		Scan(&s.owner.UserID, &s.owner.Name); err != nil {
		t.Fatal(err)
	}
	s.owner.Role = auth.Owner
	if err := s.d.App.QueryRow(t.Context(), `select id from farm where name_key = 'hirbovat'`).Scan(&s.hirbovat); err != nil {
		t.Fatal(err)
	}
	return s
}

func (s *seeded) do(method, path string, body any) *httptest.ResponseRecorder {
	s.t.Helper()
	var r *strings.Reader
	if body == nil {
		r = strings.NewReader("")
	} else {
		b, err := json.Marshal(body)
		if err != nil {
			s.t.Fatal(err)
		}
		r = strings.NewReader(string(b))
	}
	req := httptest.NewRequestWithContext(auth.WithPrincipal(context.Background(), s.owner), method, path, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	return w
}

// hirbovat2026 are Hîrbovăț's deliveries in its Seria 2026 cycle window
// (§5.9: after Seria 2025's end on 28.01.2026, up to today; golden 29).
func hirbovat2026(t *testing.T, pool *pgxpool.Pool, farm uuid.UUID) calc.FarmDeliveries {
	t.Helper()
	from := timex.MustDay(2026, time.January, 29)
	got, err := calc.ReadDeliveries(t.Context(), pool, []calc.FarmRange{{FarmID: farm, DayRange: calc.DayRange{From: &from, To: testdb.Today}}})
	if err != nil || len(got) != 1 {
		t.Fatalf("ReadDeliveries = %v, %v", got, err)
	}
	return got[0]
}

// T4 criterion 2: the four unpriced Hîrbovăț deliveries, oldest first.
func TestUnpricedSeeded(t *testing.T) {
	s := newSeeded(t)
	w := s.do(http.MethodGet, "/api/deliveries?priced=false", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("GET = %d %s", w.Code, w.Body)
	}
	var res prices.UnpricedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	type row struct{ date, farm, typ, kg string }
	var got []row
	for _, it := range res.Items {
		got = append(got, row{it.Date.String(), it.Farm, it.Type, it.Kg.String()})
	}
	want := []row{
		{"2026-02-17", "Hîrbovăț", "Breeder 2", "2000"},
		{"2026-02-17", "Hîrbovăț", "Cocoși", "1000"},
		{"2026-03-14", "Hîrbovăț", "Cocoși", "6160"},
		{"2026-11-12", "Hîrbovăț", "Breeder 2", "9000"},
	}
	if !slices.Equal(got, want) || res.Count != 4 || res.Kg.String() != "18160" || !slices.Equal(res.Farms, []string{"Hîrbovăț"}) {
		t.Errorf("list = %+v, count %d, kg %s, farms %v", got, res.Count, res.Kg, res.Farms)
	}

	// BuildUnpriced (the builder a golden file calls) gives the same answer.
	built, err := prices.BuildUnpriced(t.Context(), s.d.App, testdb.Now())
	if err != nil {
		t.Fatal(err)
	}
	b, err := httpx.Marshal(built)
	if err != nil || !bytes.Equal(b, w.Body.Bytes()) {
		t.Errorf("BuildUnpriced = %s, the route %s (%v)", b, w.Body, err)
	}
}

// T4 criteria 3 and 4 on the seed: the owner prices the four deliveries
// (17.02 Breeder 2 by its total, the others per kg): every Hîrbovăț delivery
// of Seria 2026 is priced (69); the undo returns the four to "fără preț"
// and the cost inputs are golden 31's again.
func TestSetUndoSeeded(t *testing.T) {
	s := newSeeded(t)
	ctx := t.Context()
	before := hirbovat2026(t, s.d.App, s.hirbovat)
	if before.Count != 69 || before.Priced.Count != 65 || before.Priced.Lei.String() != "9562437" || before.Unpriced.Count != 4 {
		t.Fatalf("before: %d deliveries, %d priced, %s lei, %d unpriced", before.Count, before.Priced.Count, before.Priced.Lei, before.Unpriced.Count)
	}
	un := before.Unpriced.Items
	body := map[string]any{"items": []map[string]any{
		{"id": un[0].ID, "totalLei": 16000},
		{"id": un[1].ID, "pricePerKg": 8.5},
		{"id": un[2].ID, "pricePerKg": 8.25},
		{"id": un[3].ID, "pricePerKg": 7.74},
	}}
	w := s.do(http.MethodPost, "/api/deliveries/prices", body)
	if w.Code != http.StatusOK {
		t.Fatalf("set = %d %s", w.Code, w.Body)
	}
	var res prices.SetResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	type pr struct{ perKg, total string }
	var got []pr
	for _, d := range res.Deliveries {
		got = append(got, pr{d.PricePerKg.String(), d.TotalLei.String()})
	}
	if want := []pr{{"8", "16000"}, {"8.5", "8500"}, {"8.25", "50820"}, {"7.74", "69660"}}; !slices.Equal(got, want) ||
		res.Count != 4 || res.Kg.String() != "18160" || res.UndoToken == "" ||
		res.Toast != "Prețuri salvate · 4 livrări · 18 160 kg intră în cost" {
		t.Errorf("set = %+v %+v", got, res)
	}
	var pricedAtOK, byOwner int
	if err := s.d.App.QueryRow(ctx, `select count(*) filter (where fd.priced_at = a.at), count(*) filter (where fd.priced_by = $2)
		from feed_delivery fd, audit_log a where a.action = 'price_set' and fd.id = any($1)`,
		[]uuid.UUID{un[0].ID, un[1].ID, un[2].ID, un[3].ID}, s.owner.UserID).Scan(&pricedAtOK, &byOwner); err != nil {
		t.Fatal(err)
	}
	var recordTypes []string
	var farms []uuid.UUID
	var nChanges int
	if err := s.d.App.QueryRow(ctx, `select record_types, farm_ids, jsonb_array_length(changes) from audit_log where action = 'price_set'`).
		Scan(&recordTypes, &farms, &nChanges); err != nil {
		t.Fatal(err)
	}
	if pricedAtOK != 4 || byOwner != 4 || !slices.Equal(recordTypes, []string{"pret"}) ||
		!slices.Equal(farms, []uuid.UUID{s.hirbovat}) || nChanges != 8 {
		t.Errorf("rows: %d at the audit instant, %d by the owner; audit %v %v %d changes", pricedAtOK, byOwner, recordTypes, farms, nChanges)
	}
	var setEvents, priceKeys int
	if err := s.d.App.QueryRow(ctx, `select count(*) filter (where type = 'price.set'),
		count(*) filter (where payload::text ~ '(lei|price)') from live_event`).Scan(&setEvents, &priceKeys); err != nil {
		t.Fatal(err)
	}
	if setEvents != 1 || priceKeys != 0 {
		t.Errorf("%d price.set events, %d events with a price", setEvents, priceKeys)
	}
	if after := hirbovat2026(t, s.d.App, s.hirbovat); after.Priced.Count != 69 || after.Unpriced.Count != 0 ||
		after.Priced.Lei.String() != "9707417" { // 9 562 437 + 16 000 + 8 500 + 50 820 + 69 660
		t.Errorf("after the set: %d priced, %s lei, %d unpriced", after.Priced.Count, after.Priced.Lei, after.Unpriced.Count)
	}

	w = s.do(http.MethodPost, "/api/deliveries/prices/undo", map[string]string{"undoToken": res.UndoToken})
	if w.Code != http.StatusOK {
		t.Fatalf("undo = %d %s", w.Code, w.Body)
	}
	if !strings.Contains(w.Body.String(), `"toast":"Prețuri anulate · 4 livrări înapoi fără preț"`) {
		t.Errorf("undo body = %s", w.Body)
	}
	again := hirbovat2026(t, s.d.App, s.hirbovat)
	if again.Priced.Count != 65 || again.Priced.Lei.String() != "9562437" || again.Unpriced.Count != 4 ||
		again.Unpriced.Kg.String() != "18160" {
		t.Errorf("after the undo: %d priced, %s lei, %d unpriced %s kg", again.Priced.Count, again.Priced.Lei, again.Unpriced.Count, again.Unpriced.Kg)
	}
	// Golden 31's Hîrbovăț cost again: 9 562 437 lei ÷ 29 907 birds, ÷ 5 073 155 eggs.
	if c := calc.CostOf(again.Priced.Lei, 29907, 5073155, false); c.PerBird.Round(2).String() != "319.74" || c.PerEgg.Round(2).String() != "1.88" {
		t.Errorf("cost = %s / %s", c.PerBird, c.PerEgg)
	}
	var undoRows, changedEvents int
	if err := s.d.App.QueryRow(ctx, `select (select count(*) from audit_log where action = 'undo'),
		(select count(*) from live_event where type = 'price.changed')`).Scan(&undoRows, &changedEvents); err != nil {
		t.Fatal(err)
	}
	if undoRows != 1 || changedEvents != 1 {
		t.Errorf("%d undo audit rows, %d price.changed events", undoRows, changedEvents)
	}
}

// T4 criterion 6 on the seed: Modifică on the 13.11 delivery (25 640 kg
// Breeder 2, 7,7613 lei/kg · 199 000 lei "din registru") to 7,74 lei/kg.
func TestChangeSeeded(t *testing.T) {
	s := newSeeded(t)
	var id uuid.UUID
	if err := s.d.App.QueryRow(t.Context(), `select fd.id from feed_delivery fd join feed_type ft on ft.id = fd.feed_type_id
		where fd.farm_id = $1 and fd.date = '2026-11-13' and ft.name = 'Breeder 2' and fd.deleted_at is null`, s.hirbovat).Scan(&id); err != nil {
		t.Fatal(err)
	}
	path := "/api/deliveries/" + id.String() + "/price"

	w := s.do(http.MethodPatch, path, map[string]any{"pricePerKg": 7.74})
	var e httpx.ErrorBody
	if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil || w.Code != http.StatusBadRequest ||
		e.Error.Message != "Breeder 2 13.11.2026: scrie motivul ca să schimbi prețul din 7,76 în 7,74 lei/kg" {
		t.Errorf("no reason = %d %s", w.Code, w.Body)
	}

	w = s.do(http.MethodPatch, path, map[string]any{"pricePerKg": 7.74, "reason": "Factura corectată de furnizor"})
	if w.Code != http.StatusOK {
		t.Fatalf("PATCH = %d %s", w.Code, w.Body)
	}
	var res prices.ChangeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Delivery.TotalLei.String() != "198453.6" || res.Delivery.PricePerKg.String() != "7.74" || res.Delivery.PricedBy != s.owner.Name ||
		res.Toast != "Preț schimbat · Breeder 2 · 13.11.2026 · 7,74 lei/kg · 198 454 lei" {
		t.Errorf("PATCH = %+v", res)
	}
	var fromRegister bool
	var reason string
	if err := s.d.App.QueryRow(t.Context(), `select fd.price_from_register, a.reason from feed_delivery fd, audit_log a
		where fd.id = $1 and a.action = 'price_change' and a.entity_id = $1`, id).Scan(&fromRegister, &reason); err != nil {
		t.Fatal(err)
	}
	if fromRegister || reason != "Factura corectată de furnizor" {
		t.Errorf("from register %v, reason %q", fromRegister, reason)
	}
}
