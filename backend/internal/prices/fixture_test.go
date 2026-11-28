package prices

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/tokens"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// testSecret is the fixed APP_SECRET of the tests (M4 D4).
var testSecret = bytes.Repeat([]byte{7}, 32)

func undoKey(t testing.TB) tokens.Key {
	t.Helper()
	k, err := tokens.DeriveKey(testSecret, tokens.PurposeUndo)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func day(m time.Month, d int) timex.Day { return timex.MustDay(2026, m, d) }

// fixture: Hîrbovăț (farm A) in the open Seria T from 01.10.2026 with
//   - b2Nov1: Breeder 2, 01.11, 2 000 kg, unpriced;
//   - coNov2: Cocoși, 02.11, 1 000 kg, unpriced;
//   - b2Nov3: Breeder 2, 03.11, 25 640 kg, register price 7,7613 lei/kg ·
//     199 000 lei ("din registru");
//   - deleted: Breeder 2, 04.11, 500 kg, unpriced, its entry deleted;
//
// Telița (farm B) in the closed Seria 2025 from 01.09 to 30.09.2026 with
//   - closed: Breeder 2, 15.09, 3 000 kg, unpriced.
//
// Users: owner and owner2, ion (manager of both farms), the vet and viewer.
type fixture struct {
	t                                       *testing.T
	d                                       *testdb.DB
	svc                                     *Service
	farmA, farmB, breeder2, cocosi          uuid.UUID
	b2Nov1, coNov2, b2Nov3, deleted, closed uuid.UUID
	owner, owner2, ion, vet, viewer         auth.Principal
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{t: t, d: testdb.New(t)}
	f.svc = NewService(db.NewStore(f.d.App), f.d.Clock, undoKey(t))
	f.farmA = f.id(`insert into farm (name, name_key, type) values ('Hîrbovăț', 'hirbovat', 'laying') returning id`)
	f.farmB = f.id(`insert into farm (name, name_key, type) values ('Telița', 'telita', 'growth') returning id`)
	open := f.id(`insert into series (name, name_key, year, placement_date) values ('Seria T', 'seria t', 2030, $1) returning id`, day(time.October, 1))
	f.exec(`insert into cycle (series_id, farm_id, start_date, started_by) values ($1, $2, $3, 'auto')`, open, f.farmA, day(time.October, 1))
	closed := f.id(`insert into series (name, name_key, year, placement_date, status, closed_at)
		values ('Seria 2025', 'seria 2025', 2025, $1, 'closed', $2) returning id`, day(time.September, 1), testdb.Now())
	f.exec(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by)
		values ($1, $2, $3, $4, 'auto', 'emptied')`, closed, f.farmB, day(time.September, 1), day(time.September, 30))
	f.breeder2 = f.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder 2', 'F') returning id`)
	f.cocosi = f.id(`insert into feed_type (name, name_key, for_sex) values ('Cocoși', 'cocosi', 'M') returning id`)

	f.b2Nov1 = f.deliver(f.farmA, f.breeder2, day(time.November, 1), "2000")
	f.coNov2 = f.deliver(f.farmA, f.cocosi, day(time.November, 2), "1000")
	f.b2Nov3 = f.deliver(f.farmA, f.breeder2, day(time.November, 3), "25640")
	f.exec(`update feed_delivery set price_per_kg = 7.7613, total_lei = 199000, price_from_register = true where id = $1`, f.b2Nov3)
	f.deleted = f.deliver(f.farmA, f.breeder2, day(time.November, 4), "500")
	f.closed = f.deliver(f.farmB, f.breeder2, day(time.September, 15), "3000")

	user := func(name, login string, role auth.Role, farms ...uuid.UUID) auth.Principal {
		id := f.id(`insert into app_user (name, login, password_hash, role) values ($1, $2, 'x', $3) returning id`, name, login, role)
		if farms == nil && role != auth.Owner {
			farms = []uuid.UUID{}
		}
		return auth.Principal{UserID: id, Name: name, Role: role, FarmIDs: farms}
	}
	f.owner = user("Proprietar", "owner", auth.Owner)
	f.owner2 = user("Proprietar 2", "owner2", auth.Owner)
	f.ion = user("Ion Rusu", "ion", auth.Manager, f.farmA, f.farmB)
	f.vet = user("Dr. Ceban", "ceban", auth.Vet, f.farmA)
	f.viewer = user("Ana", "ana", auth.Viewer, f.farmA)
	f.exec(`update feed_delivery set deleted_at = $2 where id = $1`, f.deleted, testdb.Now())
	f.exec(`update entry set deleted_at = $2, deleted_by = $3 where id = (select entry_id from feed_delivery where id = $1)`,
		f.deleted, testdb.Now(), f.ion.UserID)
	return f
}

func (f *fixture) id(sql string, args ...any) uuid.UUID {
	f.t.Helper()
	var v uuid.UUID
	if err := f.d.App.QueryRow(context.Background(), sql, args...).Scan(&v); err != nil {
		f.t.Fatal(err)
	}
	return v
}

func (f *fixture) exec(sql string, args ...any) {
	f.t.Helper()
	if _, err := f.d.App.Exec(context.Background(), sql, args...); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) count(sql string, args ...any) int {
	f.t.Helper()
	var n int
	if err := f.d.App.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		f.t.Fatal(err)
	}
	return n
}

func (f *fixture) deliver(farm, feedType uuid.UUID, d timex.Day, kg string) uuid.UUID {
	f.t.Helper()
	e := f.id(`insert into entry (kind, farm_id, date) values ('delivery', $1, $2) returning id`, farm, d)
	return f.id(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, $5) returning id`,
		e, farm, feedType, d, kg)
}

// price is a delivery's stored price columns as text ("" for null).
type price struct {
	PerKg, Total, Input, By, At string
	FromRegister                bool
}

func (f *fixture) price(id uuid.UUID) price {
	f.t.Helper()
	var p price
	if err := f.d.App.QueryRow(context.Background(), `select coalesce(price_per_kg::text, ''), coalesce(total_lei::text, ''),
		coalesce(price_input, ''), coalesce(priced_by::text, ''), coalesce(priced_at::text, ''), price_from_register
		from feed_delivery where id = $1`, id).Scan(&p.PerKg, &p.Total, &p.Input, &p.By, &p.At, &p.FromRegister); err != nil {
		f.t.Fatal(err)
	}
	return p
}

// do sends method path with body (JSON, "" = none) as p through Routes.
func (f *fixture) do(p auth.Principal, method, path, body string) *httptest.ResponseRecorder {
	f.t.Helper()
	return serve(f.svc, p, method, path, body)
}

func serve(svc *Service, p auth.Principal, method, path, body string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	NewHandler(svc).Routes(mux, httpx.NewChain())
	req := httptest.NewRequestWithContext(auth.WithPrincipal(context.Background(), p), method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func envelope(t *testing.T, w *httptest.ResponseRecorder) httpx.APIError {
	t.Helper()
	var b httpx.ErrorBody
	if err := json.Unmarshal(w.Body.Bytes(), &b); err != nil {
		t.Fatalf("%d %s: %v", w.Code, w.Body, err)
	}
	return b.Error
}

func decode[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("%d %s: %v", w.Code, w.Body, err)
	}
	return v
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// items is a set body: each pair is an id and a member {"pricePerKg": x} or
// {"totalLei": x}.
func items(t *testing.T, its ...map[string]any) string {
	return mustJSON(t, map[string]any{"items": its})
}

func perKg(id uuid.UUID, v any) map[string]any { return map[string]any{"id": id, "pricePerKg": v} }
func total(id uuid.UUID, v any) map[string]any { return map[string]any{"id": id, "totalLei": v} }
