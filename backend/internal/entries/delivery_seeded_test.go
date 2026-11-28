package entries_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// T5 criterion 1: the Livrare hrană form state on the dev seed. Hîrbovăț:
// the open Seria 2026 cycle and the M3 stock figures (goldens 27–28);
// Telița: no open cycle, no stock (§5.8: its leftover is never shown).
func TestBuildDeliveryFormSeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	build := func(farm string) entries.DeliveryFormResponse {
		t.Helper()
		res, err := entries.BuildDeliveryForm(t.Context(), d.App, entries.DeliveryFormInput{
			FarmID: seededFarm(t, d.App, farm), Today: testdb.Today, Now: testdb.Now(),
		})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	nov := func(day int) timex.Day { return timex.MustDay(2026, time.November, day) }

	res := build("hirbovat")
	if res.Cycle == nil || res.Cycle.Series != "Seria 2026" || res.Cycle.SaptToday != 59 ||
		res.Cycle.StartDate != timex.MustDay(2026, time.February, 22) {
		t.Errorf("cycle = %+v, want Seria 2026 from 22.02.2026, SAPT 59", res.Cycle)
	}
	if len(res.FeedTypes) == 0 {
		t.Error("feedTypes: none")
	}
	w := [2]timex.Day{nov(16), nov(18)}
	want := []formStock{
		{"Breeder 2", "31664.9", "2392.2", "13.237", "2026-12-01", "in_norma", w},
		{"Cocoși", "1781.0", "170.9", "10.419", "2026-11-28", "in_norma", w},
	}
	if got := deliveryStocks(res); !slices.Equal(got, want) {
		t.Errorf("stock\n got %+v\nwant %+v", got, want)
	}

	telita := build("telita")
	if telita.Cycle != nil || telita.Stock == nil || len(telita.Stock) != 0 || len(telita.FeedTypes) != len(res.FeedTypes) {
		t.Errorf("Telița: cycle %+v, stock %+v, %d types", telita.Cycle, telita.Stock, len(telita.FeedTypes))
	}
}

func deliveryStocks(res entries.DeliveryFormResponse) []formStock {
	return formStocks(entries.FeedFormResponse{Stock: res.Stock})
}

// T5 criterion 3: on the dev seed, the 12.11.2026 delivery to Hîrbovăț
// (9 000 kg Breeder 2, unpriced) is soft-deleted and posted again by Ion
// Rusu through the handler. stockAfter is the Breeder 2 stock at the end of
// 12.11 after the save (Q13: calc.StockAt, ≈ 21 119,3 kg); the stock today
// is golden 27's 31 664,9 kg again; one audit row and the two live events.
func TestCreateDeliverySeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	farm := seededFarm(t, d.App, "hirbovat")
	nov12 := timex.MustDay(2026, time.November, 12)

	var ion auth.Principal
	if err := d.App.QueryRow(ctx, `select id, name from app_user where login = 'ion.rusu'`).Scan(&ion.UserID, &ion.Name); err != nil {
		t.Fatal(err)
	}
	ion.Role = auth.Manager
	ion.FarmIDs = []uuid.UUID{farm, seededFarm(t, d.App, "telita")}

	var old, breeder2 uuid.UUID
	var qty string
	if err := d.App.QueryRow(ctx, `select fd.entry_id, fd.feed_type_id, fd.qty_kg::text
		from feed_delivery fd join feed_type ft on ft.id = fd.feed_type_id
		where fd.farm_id = $1 and fd.date = $2 and ft.name = 'Breeder 2' and fd.deleted_at is null and fd.total_lei is null`,
		farm, nov12).Scan(&old, &breeder2, &qty); err != nil {
		t.Fatal(err)
	}
	var rowsOfOld int
	if err := d.App.QueryRow(ctx, `select count(*) from feed_delivery where entry_id = $1`, old).Scan(&rowsOfOld); err != nil {
		t.Fatal(err)
	}
	if qty != "9000.0" || rowsOfOld != 1 {
		t.Fatalf("seeded 12.11 delivery = %s kg, its entry holds %d rows; want 9000.0 alone", qty, rowsOfOld)
	}
	if _, err := d.App.Exec(ctx, `update feed_delivery set deleted_at = $2 where entry_id = $1`, old, testdb.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := d.App.Exec(ctx, `update entry set deleted_at = $2, deleted_by = $3 where id = $1`, old, testdb.Now(), ion.UserID); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	entries.NewHandler(entries.NewService(db.NewStore(d.App), d.Clock)).Routes(mux, httpx.NewChain())
	payload, err := json.Marshal(map[string]any{
		"farmId": farm, "date": nov12, "feedTypeId": breeder2, "qtyKg": 9000, "clientRequestId": uuid.New(),
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequestWithContext(auth.WithPrincipal(ctx, ion), http.MethodPost, "/api/entries/delivery", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST = %d %s", w.Code, w.Body)
	}
	var resp entries.DeliveryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	stockAt := func(day timex.Day) string {
		t.Helper()
		stock, err := calc.StockAt(ctx, d.App, farm, day)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range stock {
			if s.FeedTypeID == breeder2 {
				return s.Kg.StringFixed(1)
			}
		}
		t.Fatalf("no Breeder 2 stock on %s", day)
		return ""
	}
	if oracle := stockAt(nov12); resp.StockAfter.Decimal().StringFixed(1) != oracle || oracle != "21119.3" {
		t.Errorf("stockAfter = %s, calc.StockAt(12.11) = %s, planner 21119.3", resp.StockAfter, oracle)
	}
	if want := "Livrare adăugată · 9 000 kg Breeder 2 · stoc Hîrbovăț: 21 119 kg · fără preț"; resp.Toast != want {
		t.Errorf("toast = %q, want %q", resp.Toast, want)
	}
	if got := stockAt(testdb.Today); got != "31664.9" { // golden 27
		t.Errorf("stock today = %s, want 31664.9", got)
	}

	var (
		n                int
		action, summary  string
		recordTypes      []string
		events           []string
		deliveriesOf1211 int
	)
	if err := d.App.QueryRow(ctx, `select count(*), min(action::text), min(summary), min(record_types::text)::text[]
		from audit_log where entry_id = $1`, resp.Entry.ID).Scan(&n, &action, &summary, &recordTypes); err != nil {
		t.Fatal(err)
	}
	if n != 1 || action != "create" || summary != "Livrare · 9 000 kg Breeder 2" || !slices.Equal(recordTypes, []string{"livrare"}) {
		t.Errorf("audit = %d %s %q %v", n, action, summary, recordTypes)
	}
	if err := d.App.QueryRow(ctx, `select array_agg(type order by id) from live_event where payload->>'entryId' = $1::text`,
		resp.Entry.ID).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(events, []string{"entry.created", "delivery.created"}) {
		t.Errorf("live events = %v", events)
	}
	if err := d.App.QueryRow(ctx, `select count(*) from feed_delivery fd join entry n on n.id = fd.entry_id and n.deleted_at is null
		where fd.farm_id = $1 and fd.date = $2 and fd.deleted_at is null and fd.total_lei is null`, farm, nov12).
		Scan(&deliveriesOf1211); err != nil || deliveriesOf1211 != 1 {
		t.Errorf("live unpriced 12.11 deliveries = %d (%v), want 1", deliveriesOf1211, err)
	}
}
