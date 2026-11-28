package prices

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/tokens"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

const (
	pathList   = "/api/deliveries?priced=false"
	pathSet    = "/api/deliveries/prices"
	pathUndo   = "/api/deliveries/prices/undo"
	asOfNoon   = "2026-11-18T12:00:00+02:00"
	staleMsg   = "Înregistrarea s-a schimbat între timp. Reîncarcă și încearcă din nou."
	invalidMsg = "Datele trimise nu sunt valide."
)

func pathChange(id any) string { return "/api/deliveries/" + fmt.Sprint(id) + "/price" }

// T4 criterion 1: every route is the owner's, checked before the body is
// read: vet and viewer 403 FORBIDDEN everywhere; a manager 403
// PRICE_FORBIDDEN on the writes and FORBIDDEN on the list.
func TestRoles(t *testing.T) {
	f := newFixture(t)
	routes := []struct{ method, path string }{
		{http.MethodPost, pathSet}, {http.MethodPost, pathUndo}, {http.MethodPatch, pathChange(f.b2Nov3)},
	}
	for _, p := range []auth.Principal{f.vet, f.viewer, f.ion} {
		for _, r := range routes {
			w := f.do(p, r.method, r.path, `{"not":"read"}`)
			code := httpx.CodeForbidden
			if p.Role == auth.Manager {
				code = httpx.CodePriceForbidden
			}
			if e := envelope(t, w); w.Code != http.StatusForbidden || e.Code != code {
				t.Errorf("%s %s %s = %d %s, want 403 %s", p.Role, r.method, r.path, w.Code, e.Code, code)
			}
		}
		w := f.do(p, http.MethodGet, pathList, "")
		if e := envelope(t, w); w.Code != http.StatusForbidden || e.Code != httpx.CodeForbidden {
			t.Errorf("%s list = %d %s, want 403 FORBIDDEN", p.Role, w.Code, e.Code)
		}
	}
	if e := envelope(t, f.do(f.ion, http.MethodPost, pathSet, `{}`)); e.Message != "Prețul îl adaugă proprietarul în Prețuri hrană." {
		t.Errorf("manager message = %q", e.Message)
	}
	if n := f.count(`select count(*) from audit_log`); n != 0 {
		t.Errorf("%d audit rows", n)
	}
}

// T4 criterion 2 (hand-built): the S12 list holds the live unpriced
// deliveries oldest first; deleted ones, priced ones and those inside a
// cycle of a closed series are left out. priced absent or not "false", or
// another parameter → 400.
func TestUnpriced(t *testing.T) {
	f := newFixture(t)
	w := f.do(f.owner, http.MethodGet, pathList, "")
	if w.Code != http.StatusOK {
		t.Fatalf("GET = %d %s", w.Code, w.Body)
	}
	want := `{"items":[` +
		`{"id":"` + f.b2Nov1.String() + `","date":"2026-11-01","farm":"Hîrbovăț","type":"Breeder 2","forSex":"F","kg":2000},` +
		`{"id":"` + f.coNov2.String() + `","date":"2026-11-02","farm":"Hîrbovăț","type":"Cocoși","forSex":"M","kg":1000}],` +
		`"count":2,"kg":3000,"farms":["Hîrbovăț"],"asOf":"` + asOfNoon + `"}`
	if got := strings.TrimSpace(w.Body.String()); got != want {
		t.Errorf("body\n got %s\nwant %s", got, want)
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Error("no Cache-Control: no-store")
	}

	for _, c := range []struct{ query, path, code string }{
		{"", "priced", httpx.IssueRequired},
		{"?priced=true", "priced", httpx.IssueOutOfRange},
		{"?priced=false&priced=false", "priced", httpx.IssueOutOfRange},
		{"?priced=false&farmId=x", "farmId", httpx.IssueUnknownField},
	} {
		w := f.do(f.owner, http.MethodGet, "/api/deliveries"+c.query, "")
		e := envelope(t, w)
		if w.Code != http.StatusBadRequest || len(e.Issues) != 1 || e.Issues[0].Path != c.path || e.Issues[0].Code != c.code {
			t.Errorf("%q = %d %+v, want 400 %s %s", c.query, w.Code, e.Issues, c.path, c.code)
		}
	}

	// Nothing to price: an empty list, never null.
	f.exec(`update feed_delivery set price_per_kg = 1, total_lei = 1, price_from_register = true where total_lei is null`)
	if got := strings.TrimSpace(f.do(f.owner, http.MethodGet, pathList, "").Body.String()); got !=
		`{"items":[],"count":0,"kg":0,"farms":[],"asOf":"`+asOfNoon+`"}` {
		t.Errorf("empty list = %s", got)
	}
}

// T4 criterion 3 (hand-built): a batch prices every item (the typed value
// and the derived one), priced_by = the owner, priced_at = the audit row's
// at; one price_set audit row with the Q10 columns and changes; one
// price.set event without lei.
func TestSet(t *testing.T) {
	f := newFixture(t)
	// A clock with nanoseconds: priced_at and audit_log.at must still match (D4).
	f.d.Clock.Set(testdb.Now().Add(123456789 * time.Nanosecond))
	w := f.do(f.owner, http.MethodPost, pathSet, items(t, total(f.b2Nov1, 16000), perKg(f.coNov2, 8.5)))
	if w.Code != http.StatusOK {
		t.Fatalf("POST = %d %s", w.Code, w.Body)
	}
	res := decode[SetResponse](t, w)
	want := `{"deliveries":[{"id":"` + f.b2Nov1.String() + `","pricePerKg":8,"totalLei":16000},` +
		`{"id":"` + f.coNov2.String() + `","pricePerKg":8.5,"totalLei":8500}],"count":2,"kg":3000,` +
		"\"toast\":\"Prețuri salvate · 2 livrări · 3\u202f000 kg intră în cost\",\"undoToken\":\"" + res.UndoToken + `",` +
		`"asOf":"2026-11-18T12:00:00+02:00"}`
	if got := strings.TrimSpace(w.Body.String()); got != want || res.UndoToken == "" {
		t.Errorf("body\n got %s\nwant %s", got, want)
	}
	if got, want := f.price(f.b2Nov1), (price{"8.0000", "16000.00", "total", f.owner.UserID.String(), f.price(f.b2Nov1).At, false}); got != want || got.At == "" {
		t.Errorf("Breeder 2 01.11 = %+v", got)
	}
	if got := f.price(f.coNov2); got.PerKg != "8.5000" || got.Total != "8500.00" || got.Input != "per_kg" || got.By != f.owner.UserID.String() {
		t.Errorf("Cocoși 02.11 = %+v", got)
	}
	if n := f.count(`select count(*) from feed_delivery fd join audit_log a on a.at = fd.priced_at
		where a.action = 'price_set' and fd.id = any($1)`, []uuid.UUID{f.b2Nov1, f.coNov2}); n != 2 {
		t.Errorf("%d deliveries with priced_at = the audit row's at, want 2", n)
	}

	var (
		entity, summary string
		entityID        uuid.UUID
		recordTypes     []string
		farms           []uuid.UUID
		recordDate      timex.Day
		changes         []auditChange
		newValues       []priceValues
		userID          uuid.UUID
	)
	if err := f.d.App.QueryRow(t.Context(), `select entity, entity_id, record_types, farm_ids, record_date, summary,
		changes, new_values, user_id from audit_log where action = 'price_set'`).
		Scan(&entity, &entityID, &recordTypes, &farms, &recordDate, &summary, &changes, &newValues, &userID); err != nil {
		t.Fatal(err)
	}
	if entity != "feed_delivery" || entityID != f.b2Nov1 || !slices.Equal(recordTypes, []string{"pret"}) ||
		!slices.Equal(farms, []uuid.UUID{f.farmA}) || recordDate != day(time.November, 1) ||
		summary != "Prețuri · 2 livrări · 3\u202f000 kg" || userID != f.owner.UserID {
		t.Errorf("audit row: %s %s %v %v %s %q %s", entity, entityID, recordTypes, farms, recordDate, summary, userID)
	}
	if len(changes) != 4 || changes[0].Key != "Breeder 2 · 01.11.2026 · lei/kg" || changes[0].Field != "pricePerKg" ||
		changes[0].Old != nil || changes[0].New == nil || changes[0].New.String() != "8" ||
		changes[3].Key != "Cocoși · 02.11.2026 · total lei" || changes[3].New.String() != "8500" {
		t.Errorf("changes = %+v", changes)
	}
	if len(newValues) != 2 || newValues[0].ID != f.b2Nov1 || newValues[1].ID != f.coNov2 {
		t.Errorf("new_values = %+v", newValues)
	}

	var typ string
	var payload map[string]any
	var evFarms []uuid.UUID
	if err := f.d.App.QueryRow(t.Context(), `select type, payload, farm_ids from live_event`).Scan(&typ, &payload, &evFarms); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(payload))
	for k := range payload {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	if typ != "price.set" || !slices.Equal(keys, []string{"count", "deliveryIds", "kg"}) || payload["count"] != 2.0 ||
		payload["kg"] != 3000.0 || !slices.Equal(evFarms, []uuid.UUID{f.farmA}) {
		t.Errorf("live event %s %v %v", typ, payload, evFarms)
	}

	// The undo of a batch saved at a nanosecond instant still matches (D4).
	if w := f.do(f.owner, http.MethodPost, pathUndo, mustJSON(t, map[string]string{"undoToken": res.UndoToken})); w.Code != http.StatusOK {
		t.Errorf("undo = %d %s", w.Code, w.Body)
	}
}

// T4 criterion 5 (level 1): every shape problem is a 400 naming the item's
// field; nothing is saved.
func TestSetValidation(t *testing.T) {
	f := newFixture(t)
	unknown := uuid.New()
	for _, c := range []struct {
		name, body, path, code, msg string
	}{
		{"no items", `{"items":[]}`, "items", httpx.IssueRequired, invalidMsg},
		{"items missing", `{}`, "items", httpx.IssueRequired, invalidMsg},
		{"price 0", items(t, total(f.b2Nov1, 16000), perKg(f.coNov2, 0)), "items[1].pricePerKg", httpx.IssueOutOfRange,
			"Cocoși 02.11.2026: scrie un preț mai mare de 0 lei/kg"},
		{"negative price", items(t, perKg(f.coNov2, -1)), "items[0].pricePerKg", httpx.IssueOutOfRange,
			"Cocoși 02.11.2026: scrie un preț mai mare de 0 lei/kg"},
		{"total 0", items(t, total(f.coNov2, 0)), "items[0].totalLei", httpx.IssueOutOfRange,
			"Cocoși 02.11.2026: scrie un total mai mare de 0 lei"},
		{"both", mustJSON(t, map[string]any{"items": []any{map[string]any{"id": f.coNov2, "pricePerKg": 8, "totalLei": 8000}}}),
			"items[0].pricePerKg", httpx.IssueOutOfRange, "Cocoși 02.11.2026: scrie prețul pe kg sau totalul, nu amândouă."},
		{"neither", mustJSON(t, map[string]any{"items": []any{map[string]any{"id": f.coNov2}}}),
			"items[0].pricePerKg", httpx.IssueRequired, invalidMsg},
		{"null price", mustJSON(t, map[string]any{"items": []any{map[string]any{"id": f.coNov2, "pricePerKg": nil}}}),
			"items[0].pricePerKg", httpx.IssueRequired, invalidMsg},
		{"5 decimals", items(t, perKg(f.coNov2, 8.12345)), "items[0].pricePerKg", httpx.IssueOutOfRange, invalidMsg},
		{"3 decimals total", items(t, total(f.coNov2, 8000.125)), "items[0].totalLei", httpx.IssueOutOfRange, invalidMsg},
		{"price too large", items(t, perKg(f.coNov2, 100000000)), "items[0].pricePerKg", httpx.IssueOutOfRange, invalidMsg},
		{"duplicate", items(t, perKg(f.coNov2, 8), perKg(f.coNov2, 9)), "items[1].id", httpx.IssueOutOfRange, invalidMsg},
		{"unknown id", items(t, perKg(unknown, 8)), "items[0].id", httpx.IssueOutOfRange, invalidMsg},
		{"no id", `{"items":[{"pricePerKg":8}]}`, "items[0].id", httpx.IssueRequired, invalidMsg},
		{"dryRun", `{"items":[],"dryRun":true}`, "dryRun", httpx.IssueUnknownField, invalidMsg},
		{"clientRequestId", `{"clientRequestId":"` + uuid.NewString() + `","items":[]}`, "clientRequestId", httpx.IssueUnknownField, invalidMsg},
		{"string price", `{"items":[{"id":"` + f.coNov2.String() + `","pricePerKg":"8"}]}`, "items[0].pricePerKg", httpx.IssueBadType, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := f.do(f.owner, http.MethodPost, pathSet, c.body)
			e := envelope(t, w)
			if w.Code != http.StatusBadRequest || e.Code != httpx.CodeValidation || len(e.Issues) == 0 ||
				e.Issues[0].Path != c.path || e.Issues[0].Code != c.code || (c.msg != "" && e.Message != c.msg) {
				t.Errorf("= %d %s %q %+v, want 400 %s %s %q", w.Code, e.Code, e.Message, e.Issues, c.path, c.code, c.msg)
			}
		})
	}
	if p := f.price(f.b2Nov1); p.Total != "" || f.count(`select count(*) from audit_log`) != 0 {
		t.Errorf("something was saved: %+v", p)
	}
}

// T4 criterion 5 (level 2): an already priced or deleted delivery is 409
// STALE for the whole batch, one inside a closed series' cycle 409
// SERIES_CLOSED; nothing is saved.
func TestSetRefusals(t *testing.T) {
	f := newFixture(t)
	for _, c := range []struct {
		name   string
		second uuid.UUID
		code   string
		msg    string
	}{
		{"priced", f.b2Nov3, httpx.CodeStale, staleMsg},
		{"deleted", f.deleted, httpx.CodeStale, staleMsg},
		{"closed series", f.closed, httpx.CodeSeriesClosed, "Seria 2025 e închisă: nu mai primește înregistrări sau corecții."},
	} {
		w := f.do(f.owner, http.MethodPost, pathSet, items(t, total(f.b2Nov1, 16000), perKg(c.second, 8)))
		if e := envelope(t, w); w.Code != http.StatusConflict || e.Code != c.code || e.Message != c.msg {
			t.Errorf("%s = %d %s %q, want 409 %s %q", c.name, w.Code, e.Code, e.Message, c.code, c.msg)
		}
	}
	if p := f.price(f.b2Nov1); p.Total != "" || f.count(`select count(*) from audit_log`)+f.count(`select count(*) from live_event`) != 0 {
		t.Errorf("something was saved: %+v", p)
	}
	if p := f.price(f.b2Nov3); p.PerKg != "7.7613" || !p.FromRegister {
		t.Errorf("the register price changed: %+v", p)
	}
}

// setBatch prices Breeder 2 01.11 (total 16 000) and Cocoși 02.11 (8,5
// lei/kg) as the owner and returns the undo token.
func (f *fixture) setBatch() string {
	f.t.Helper()
	w := f.do(f.owner, http.MethodPost, pathSet, items(f.t, total(f.b2Nov1, 16000), perKg(f.coNov2, 8.5)))
	if w.Code != http.StatusOK {
		f.t.Fatalf("set = %d %s", w.Code, w.Body)
	}
	return decode[SetResponse](f.t, w).UndoToken
}

func undoBody(t *testing.T, token string) string {
	return mustJSON(t, map[string]string{"undoToken": token})
}

// T4 criterion 4 (hand-built): the undo returns the batch to "fără preț",
// writes one undo audit row (same_day_author, old_values = the prices) and
// one price.changed event; the same token again is 409 STALE.
func TestUndo(t *testing.T) {
	f := newFixture(t)
	token := f.setBatch()
	w := f.do(f.owner, http.MethodPost, pathUndo, undoBody(t, token))
	if w.Code != http.StatusOK {
		t.Fatalf("undo = %d %s", w.Code, w.Body)
	}
	want := `{"deliveries":["` + f.b2Nov1.String() + `","` + f.coNov2.String() + `"],` +
		`"toast":"Prețuri anulate · 2 livrări înapoi fără preț","asOf":"` + asOfNoon + `"}`
	if got := strings.TrimSpace(w.Body.String()); got != want {
		t.Errorf("body\n got %s\nwant %s", got, want)
	}
	for _, id := range []uuid.UUID{f.b2Nov1, f.coNov2} {
		if p := f.price(id); p != (price{}) {
			t.Errorf("%s still priced: %+v", id, p)
		}
	}
	var sameDay bool
	var reason *string
	var old []priceValues
	var farms []uuid.UUID
	if err := f.d.App.QueryRow(t.Context(), `select same_day_author, reason, old_values, farm_ids from audit_log where action = 'undo'`).
		Scan(&sameDay, &reason, &old, &farms); err != nil {
		t.Fatal(err)
	}
	if !sameDay || reason != nil || len(old) != 2 || old[0].TotalLei == nil || old[0].TotalLei.String() != "16000" ||
		old[1].PricePerKg.String() != "8.5" || !slices.Equal(farms, []uuid.UUID{f.farmA}) {
		t.Errorf("undo audit row: %v %v %+v %v", sameDay, reason, old, farms)
	}
	if n := f.count(`select count(*) from live_event where type = 'price.changed' and payload->>'count' = '2'`); n != 1 {
		t.Errorf("%d price.changed events", n)
	}
	if n := f.count(`select count(*) from live_event where payload::text ~ '(lei|price)'`); n != 0 {
		t.Errorf("%d live events carry a price", n)
	}

	w = f.do(f.owner, http.MethodPost, pathUndo, undoBody(t, token))
	if e := envelope(t, w); w.Code != http.StatusConflict || e.Code != httpx.CodeStale || e.Message != staleMsg {
		t.Errorf("same token again = %d %s %q", w.Code, e.Code, e.Message)
	}
}

// T4 criterion 4: the undo refusals of Q11, each leaving the prices as they
// are.
func TestUndoRefusals(t *testing.T) {
	f := newFixture(t)
	token := f.setBatch()
	key := undoKey(t)
	setupToken, err := tokens.Sign(key, []string{"setup", "1", f.owner.UserID.String()}, testdb.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	p64, sig, _ := strings.Cut(token, ".")
	tampered := p64[:len(p64)-2] + "AA." + sig
	otherKey, err := tokens.DeriveKey(testSecret, tokens.PurposeVoicePending)
	if err != nil {
		t.Fatal(err)
	}
	otherKind, err := tokens.Sign(otherKey, []string{"price", "1", f.owner.UserID.String()}, testdb.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name  string
		p     auth.Principal
		body  string
		state int
		code  string
		msg   string
	}{
		{"another owner", f.owner2, undoBody(t, token), http.StatusForbidden, httpx.CodeForbidden, "Nu ai acces la această pagină."},
		{"tampered", f.owner, undoBody(t, tampered), http.StatusBadRequest, httpx.CodeValidation, invalidMsg},
		{"setup kind", f.owner, undoBody(t, setupToken), http.StatusBadRequest, httpx.CodeValidation, invalidMsg},
		{"other purpose", f.owner, undoBody(t, otherKind), http.StatusBadRequest, httpx.CodeValidation, invalidMsg},
		{"garbage", f.owner, undoBody(t, "abc"), http.StatusBadRequest, httpx.CodeValidation, invalidMsg},
		{"empty", f.owner, undoBody(t, ""), http.StatusBadRequest, httpx.CodeValidation, invalidMsg},
		{"unknown field", f.owner, `{"undoToken":"x","dryRun":true}`, http.StatusBadRequest, httpx.CodeValidation, invalidMsg},
	} {
		w := f.do(c.p, http.MethodPost, pathUndo, c.body)
		if e := envelope(t, w); w.Code != c.state || e.Code != c.code || e.Message != c.msg {
			t.Errorf("%s = %d %s %q, want %d %s %q", c.name, w.Code, e.Code, e.Message, c.state, c.code, c.msg)
		}
	}
	if w := f.do(f.owner, http.MethodPost, pathUndo, undoBody(t, tampered)); envelope(t, w).Issues[0].Path != "undoToken" {
		t.Errorf("tampered issue = %s", w.Body)
	}

	// After 23:59 of the save day.
	f.d.Clock.Set(time.Date(2026, time.November, 19, 0, 0, 0, 0, timex.Chisinau))
	w := f.do(f.owner, http.MethodPost, pathUndo, undoBody(t, token))
	if e := envelope(t, w); w.Code != http.StatusConflict || e.Code != httpx.CodeStale ||
		e.Message != "Nu se mai poate anula: ziua salvării s-a încheiat." {
		t.Errorf("next day = %d %s %q", w.Code, e.Code, e.Message)
	}
	f.d.Clock.Set(time.Date(2026, time.November, 18, 23, 59, 59, 0, timex.Chisinau))

	// A delivery of the batch priced again (changed with a Motiv) after the save.
	w = f.do(f.owner, http.MethodPatch, pathChange(f.coNov2), `{"pricePerKg":9,"reason":"Factura nouă"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("change = %d %s", w.Code, w.Body)
	}
	w = f.do(f.owner, http.MethodPost, pathUndo, undoBody(t, token))
	if e := envelope(t, w); w.Code != http.StatusConflict || e.Code != httpx.CodeStale {
		t.Errorf("re-priced = %d %s", w.Code, e.Code)
	}
	if p := f.price(f.b2Nov1); p.Total != "16000.00" {
		t.Errorf("Breeder 2 01.11 = %+v, want still priced", p)
	}
	if p := f.price(f.coNov2); p.PerKg != "9.0000" {
		t.Errorf("Cocoși 02.11 = %+v, want the changed price", p)
	}
	if n := f.count(`select count(*) from audit_log where action = 'undo'`); n != 0 {
		t.Errorf("%d undo audit rows", n)
	}
}

// D4: the token is valid until 23:59 of the save day, also across the
// 25.10.2026 DST change in Europe/Chisinau (a 25-hour day).
func TestUndoExpiry(t *testing.T) {
	for _, c := range []struct{ now, want time.Time }{
		{time.Date(2026, 11, 18, 12, 0, 0, 0, timex.Chisinau), time.Date(2026, 11, 19, 0, 0, 0, 0, timex.Chisinau)},
		{time.Date(2026, 11, 18, 23, 59, 59, 0, timex.Chisinau), time.Date(2026, 11, 19, 0, 0, 0, 0, timex.Chisinau)},
		{time.Date(2026, 11, 18, 0, 0, 0, 0, timex.Chisinau), time.Date(2026, 11, 19, 0, 0, 0, 0, timex.Chisinau)},
		{time.Date(2026, 10, 24, 22, 30, 0, 0, time.UTC), time.Date(2026, 10, 26, 0, 0, 0, 0, timex.Chisinau)}, // 25.10 01:30 +03
		{time.Date(2026, 10, 25, 12, 0, 0, 0, timex.Chisinau), time.Date(2026, 10, 26, 0, 0, 0, 0, timex.Chisinau)},
		{time.Date(2026, 3, 29, 1, 0, 0, 0, timex.Chisinau), time.Date(2026, 3, 30, 0, 0, 0, 0, timex.Chisinau)},
	} {
		if got := undoExpiry(c.now); !got.Equal(c.want) {
			t.Errorf("undoExpiry(%s) = %s, want %s", c.now, got, c.want)
		}
	}
	if got := undoExpiry(time.Date(2026, 10, 25, 12, 0, 0, 0, timex.Chisinau)).Sub(
		time.Date(2026, 10, 25, 0, 0, 0, 0, timex.Chisinau)); got != 25*time.Hour {
		t.Errorf("25.10.2026 lasts %s, want 25h", got)
	}
}

// T4 criterion 6 (hand-built): a price change with a Motiv.
func TestChange(t *testing.T) {
	f := newFixture(t)
	w := f.do(f.owner, http.MethodPatch, pathChange(f.b2Nov3), `{"pricePerKg":7.74,"reason":"  Factura corectată  "}`)
	if w.Code != http.StatusOK {
		t.Fatalf("PATCH = %d %s", w.Code, w.Body)
	}
	want := `{"delivery":{"id":"` + f.b2Nov3.String() + `","pricePerKg":7.74,"totalLei":198453.6,"pricedBy":"Proprietar",` +
		`"pricedAt":"` + asOfNoon + `"},` + "\"toast\":\"Preț schimbat · Breeder 2 · 03.11.2026 · 7,74 lei/kg · 198\u202f454 lei\"," +
		`"asOf":"` + asOfNoon + `"}`
	if got := strings.TrimSpace(w.Body.String()); got != want {
		t.Errorf("body\n got %s\nwant %s", got, want)
	}
	if p := f.price(f.b2Nov3); p.PerKg != "7.7400" || p.Total != "198453.60" || p.Input != "per_kg" ||
		p.By != f.owner.UserID.String() || p.FromRegister {
		t.Errorf("row = %+v", p)
	}
	var reason, summary string
	var changes []auditChange
	var old changeOldValues
	var recordTypes []string
	if err := f.d.App.QueryRow(t.Context(), `select reason, summary, changes, old_values, record_types from audit_log where action = 'price_change'`).
		Scan(&reason, &summary, &changes, &old, &recordTypes); err != nil {
		t.Fatal(err)
	}
	if reason != "Factura corectată" || summary != "Preț schimbat · Breeder 2 · 03.11.2026" || len(changes) != 2 ||
		changes[0].Key != "Breeder 2 · 03.11.2026 · lei/kg" || changes[0].Old.String() != "7.7613" || changes[0].New.String() != "7.74" ||
		changes[1].Key != "Breeder 2 · 03.11.2026 · total lei" || changes[1].Old.String() != "199000" || changes[1].New.String() != "198453.6" ||
		!old.PriceFromRegister || old.PricedBy != nil || !slices.Equal(recordTypes, []string{"pret"}) {
		t.Errorf("audit: %q %q %+v %+v %v", reason, summary, changes, old, recordTypes)
	}
	if n := f.count(`select count(*) from live_event where type = 'price.changed' and payload->>'count' = '1'`); n != 1 {
		t.Errorf("%d price.changed events", n)
	}

	// A total typed: the price per kg is derived.
	w = f.do(f.owner, http.MethodPatch, pathChange(f.b2Nov3), `{"totalLei":199000,"reason":"Înapoi"}`)
	if res := decode[ChangeResponse](t, w); w.Code != http.StatusOK || res.Delivery.PricePerKg.String() != "7.7613" {
		t.Errorf("total = %d %s", w.Code, w.Body)
	}
	if p := f.price(f.b2Nov3); p.Input != "total" || p.PerKg != "7.7613" {
		t.Errorf("row = %+v", p)
	}
}

// T4 criterion 6: the price change refusals of Q12.
func TestChangeRefusals(t *testing.T) {
	f := newFixture(t)
	f.exec(`update feed_delivery set price_per_kg = 8, total_lei = 24000, price_from_register = true where id = $1`, f.closed)
	long := strings.Repeat("ă", 212)
	for _, c := range []struct {
		name, path, body string
		state            int
		code, msg, field string
	}{
		{"no reason", pathChange(f.b2Nov3), `{"pricePerKg":7.74}`, http.StatusBadRequest, httpx.CodeValidation,
			"Breeder 2 03.11.2026: scrie motivul ca să schimbi prețul din 7,76 în 7,74 lei/kg", "reason"},
		{"blank reason, total typed", pathChange(f.b2Nov3), `{"totalLei":198453.6,"reason":"   "}`, http.StatusBadRequest,
			httpx.CodeValidation, "Breeder 2 03.11.2026: scrie motivul ca să schimbi prețul din 7,76 în 7,74 lei/kg", "reason"},
		{"long reason", pathChange(f.b2Nov3), `{"pricePerKg":7.74,"reason":"` + long + `"}`, http.StatusBadRequest,
			httpx.CodeValidation, "Motivul are 212 caractere; cel mult 200. Scurtează-l cu 12 ca să poți salva.", "reason"},
		{"price 0", pathChange(f.b2Nov3), `{"pricePerKg":0,"reason":"x"}`, http.StatusBadRequest, httpx.CodeValidation,
			"Breeder 2 03.11.2026: scrie un preț mai mare de 0 lei/kg", "pricePerKg"},
		{"both", pathChange(f.b2Nov3), `{"pricePerKg":7,"totalLei":7,"reason":"x"}`, http.StatusBadRequest, httpx.CodeValidation,
			"Breeder 2 03.11.2026: scrie prețul pe kg sau totalul, nu amândouă.", "pricePerKg"},
		{"same price", pathChange(f.b2Nov3), `{"totalLei":199000,"reason":"x"}`, http.StatusUnprocessableEntity, httpx.CodeRefused,
			"Nu ai schimbat nicio cifră.", ""},
		{"same per kg", pathChange(f.b2Nov3), `{"pricePerKg":7.7613,"reason":"x"}`, http.StatusOK, "", "", ""},
		{"unpriced", pathChange(f.b2Nov1), `{"pricePerKg":8,"reason":"x"}`, http.StatusConflict, httpx.CodeStale, staleMsg, ""},
		{"closed series", pathChange(f.closed), `{"pricePerKg":9,"reason":"x"}`, http.StatusConflict, httpx.CodeSeriesClosed,
			"Seria 2025 e închisă: nu mai primește înregistrări sau corecții.", ""},
		{"unknown", pathChange(uuid.New()), `{"pricePerKg":8,"reason":"x"}`, http.StatusNotFound, httpx.CodeNotFound, "Adresa cerută nu există.", ""},
		{"deleted", pathChange(f.deleted), `{"pricePerKg":8,"reason":"x"}`, http.StatusNotFound, httpx.CodeNotFound, "Adresa cerută nu există.", ""},
		{"bad id", pathChange("abc"), `{"pricePerKg":8,"reason":"x"}`, http.StatusBadRequest, httpx.CodeValidation, invalidMsg, "id"},
		{"unknown field", pathChange(f.b2Nov3), `{"pricePerKg":8,"reason":"x","ifVersion":"1"}`, http.StatusBadRequest,
			httpx.CodeValidation, invalidMsg, "ifVersion"},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := f.do(f.owner, http.MethodPatch, c.path, c.body)
			if c.state == http.StatusOK {
				// 7,7613 lei/kg × 25 640 kg = 198 999,73 lei ≠ 199 000: a change.
				if w.Code != http.StatusOK || decode[ChangeResponse](t, w).Delivery.TotalLei.String() != "198999.73" {
					t.Errorf("= %d %s", w.Code, w.Body)
				}
				return
			}
			e := envelope(t, w)
			if w.Code != c.state || e.Code != c.code || e.Message != c.msg ||
				(c.field != "" && (len(e.Issues) == 0 || e.Issues[0].Path != c.field)) {
				t.Errorf("= %d %s %q %+v, want %d %s %q at %s", w.Code, e.Code, e.Message, e.Issues, c.state, c.code, c.msg, c.field)
			}
		})
	}
}

// T4 criterion 7: two concurrent batches on the same delivery: exactly one
// 200 and one 409 STALE, no 500.
func TestSetConcurrent(t *testing.T) {
	f := newFixture(t)
	const n = 2
	var (
		start sync.WaitGroup
		done  sync.WaitGroup
		mu    sync.Mutex
		codes []int
	)
	start.Add(n)
	done.Add(n)
	for i := range n {
		body := items(t, perKg(f.b2Nov1, 8+i))
		go func() {
			defer done.Done()
			start.Done()
			start.Wait() // barrier: both send together
			w := serve(f.svc, f.owner, http.MethodPost, pathSet, body)
			mu.Lock()
			codes = append(codes, w.Code)
			if w.Code == http.StatusConflict && !strings.Contains(w.Body.String(), `"code":"STALE"`) {
				codes = append(codes, -1)
			}
			mu.Unlock()
		}()
	}
	done.Wait()
	slices.Sort(codes)
	if !slices.Equal(codes, []int{http.StatusOK, http.StatusConflict}) {
		t.Errorf("codes = %v, want [200 409] (the 409 STALE)", codes)
	}
	if n := f.count(`select count(*) from audit_log where action = 'price_set'`); n != 1 {
		t.Errorf("%d price_set rows", n)
	}
}

// statementLog is a pgx QueryTracer that records the "-- name: X" of every
// statement ("?" without one: begin, commit).
type statementLog struct {
	mu    sync.Mutex
	names []string
}

func (s *statementLog) TraceQueryStart(ctx context.Context, _ *pgx.Conn, d pgx.TraceQueryStartData) context.Context {
	name := "?"
	if rest, ok := strings.CutPrefix(d.SQL, "-- name: "); ok {
		name = strings.Fields(rest)[0]
	}
	s.mu.Lock()
	s.names = append(s.names, name)
	s.mu.Unlock()
	return ctx
}

func (s *statementLog) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (s *statementLog) take() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.names
	s.names = nil
	return out
}

// T4 criterion 8: a batch is at most 10 statements (begin and commit
// included); the rows are updated by one statement, whatever their number.
func TestSetStatements(t *testing.T) {
	f := newFixture(t)
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
	pool, err := pgxpool.NewWithConfig(t.Context(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
	log.take()

	svc := NewService(db.NewStore(pool), f.d.Clock, undoKey(t))
	w := serve(svc, f.owner, http.MethodPost, pathSet, items(t, total(f.b2Nov1, 16000), perKg(f.coNov2, 8.5)))
	if w.Code != http.StatusOK {
		t.Fatalf("set = %d %s", w.Code, w.Body)
	}
	got := log.take()
	want := []string{
		"PricesDeliveryRefs", "?", "PricesDeliveryFarms", "LockFarm", "PricesLockDeliveries", "PricesSet",
		"InsertAuditLogID", "EmitLiveEvent", "?",
	}
	if !slices.Equal(got, want) {
		t.Errorf("statements %q,\nwant %q", got, want)
	}
	if len(got) > 10 {
		t.Errorf("%d statements, want at most 10", len(got))
	}
}
