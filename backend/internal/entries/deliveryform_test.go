package entries

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func (f *formFixture) buildDelivery(farm uuid.UUID) DeliveryFormResponse {
	f.t.Helper()
	res, err := BuildDeliveryForm(f.t.Context(), f.d.App, DeliveryFormInput{FarmID: farm, Today: testdb.Today, Now: testdb.Now()})
	if err != nil {
		f.t.Fatal(err)
	}
	return res
}

// T5 criterion 1: the Q14 shape on hand-built data. Hîrbovăț: the open
// Seria 2026 cycle (SAPT 46 today), the three active types in name order
// (Vechi is retired), and the feed form's stock rows (one reader: the same
// rows as BuildFeedForm's). Telița: no open cycle, so no stock.
func TestBuildDeliveryForm(t *testing.T) {
	f := newFormFixture(t)
	res := f.buildDelivery(f.hirbovat)
	if res.Farm != (FormFarm{ID: f.hirbovat, Name: "Hîrbovăț", Type: "laying"}) || res.Today != testdb.Today {
		t.Errorf("farm %+v today %s", res.Farm, res.Today)
	}
	wantCycle := DeliveryFormCycle{ID: f.cycle, Series: "Seria 2026", StartDate: fday(1, 1), SaptToday: 46}
	if res.Cycle == nil || *res.Cycle != wantCycle {
		t.Errorf("cycle = %+v, want %+v", res.Cycle, wantCycle)
	}
	var types []string
	for _, ft := range res.FeedTypes {
		sex := "ambele"
		if ft.ForSex != nil {
			sex = *ft.ForSex
		}
		types = append(types, ft.Name+" "+sex)
	}
	if !slices.Equal(types, []string{"Breeder 2 F", "Cocoși M", "Universal ambele"}) {
		t.Errorf("feedTypes = %v", types)
	}
	feed := f.build(f.mgr, f.hirbovat, testdb.Today)
	if len(res.Stock) != 3 || !slices.EqualFunc(res.Stock, feed.Stock, func(a, b FormStock) bool {
		return a.FeedTypeID == b.FeedTypeID && a.Kg.String() == b.Kg.String() && a.AvgDaily.String() == b.AvgDaily.String() &&
			numStr(a.DaysLeft) == numStr(b.DaysLeft) && a.Window == b.Window && a.Lamp == b.Lamp && a.Reason == b.Reason
	}) {
		t.Errorf("stock = %+v,\nfeed form %+v", res.Stock, feed.Stock)
	}
	if res.AsOf.String() != "2026-11-18T12:00:00+02:00" {
		t.Errorf("asOf = %s", res.AsOf)
	}

	empty := f.buildDelivery(f.telita)
	b, err := json.Marshal(empty)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"cycle":null`) || !strings.Contains(string(b), `"stock":[]`) || len(empty.FeedTypes) != 3 {
		t.Errorf("Telița = %s", b)
	}
	if _, err := BuildDeliveryForm(t.Context(), f.d.App, DeliveryFormInput{FarmID: uuid.New(), Today: testdb.Today}); err != ErrFormFarmNotFound {
		t.Errorf("unknown farm: err = %v, want ErrFormFarmNotFound", err)
	}
}

// T5 criterion 1: vet and viewer 403 FORBIDDEN; a manager's other farm 403
// FARM_FORBIDDEN; an owner's unknown farm 404.
func TestDeliveryFormService(t *testing.T) {
	f := newFormFixture(t)
	svc := NewService(db.NewStore(f.d.App), f.d.Clock)
	ctx := t.Context()
	viewer := auth.Principal{UserID: uuid.New(), Role: auth.Viewer, FarmIDs: []uuid.UUID{f.hirbovat}}
	mgrT := auth.Principal{UserID: uuid.New(), Role: auth.Manager, FarmIDs: []uuid.UUID{f.telita}}
	owner := auth.Principal{UserID: uuid.New(), Role: auth.Owner}

	for _, p := range []auth.Principal{f.vet, viewer} {
		_, err := svc.DeliveryForm(ctx, p, f.hirbovat)
		if body := wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden); body.Message != "Nu ai acces la această pagină." {
			t.Errorf("%s message = %q", p.Role, body.Message)
		}
	}
	_, err := svc.DeliveryForm(ctx, mgrT, f.hirbovat)
	wantCode(t, err, http.StatusForbidden, httpx.CodeFarmForbidden)
	_, err = svc.DeliveryForm(ctx, mgrT, uuid.Nil)
	wantCode(t, err, http.StatusForbidden, httpx.CodeFarmForbidden)
	_, err = svc.DeliveryForm(ctx, owner, uuid.New())
	wantCode(t, err, http.StatusNotFound, httpx.CodeNotFound)

	for _, p := range []auth.Principal{f.mgr, owner} {
		if res, err := svc.DeliveryForm(ctx, p, f.hirbovat); err != nil || len(res.Stock) != 3 {
			t.Errorf("%s: %d stock rows, %v", p.Role, len(res.Stock), err)
		}
	}
}

// T5 criterion 1: the query parameter.
func TestParseDeliveryFormQuery(t *testing.T) {
	id := uuid.New()
	const invalid = "Datele trimise nu sunt valide."
	for _, tc := range []struct {
		name  string
		query string
		want  []httpx.Issue // nil = ok
	}{
		{"ok", "farmId=" + id.String(), nil},
		{"missing", "", []httpx.Issue{{Path: "farmId", Code: httpx.IssueRequired, Message: invalid}}},
		{"malformed", "farmId=x", []httpx.Issue{{Path: "farmId", Code: httpx.IssueBadType, Message: invalid}}},
		{"repeated and unknown", "farmId=" + id.String() + "&farmId=" + id.String() + "&date=2026-11-18", []httpx.Issue{
			{Path: "farmId", Code: httpx.IssueBadType, Message: invalid},
			{Path: "date", Code: httpx.IssueUnknownField, Message: invalid},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/api/forms/delivery?"+tc.query, nil)
			farm, err := parseDeliveryFormQuery(r.URL.Query())
			if tc.want == nil {
				if err != nil || farm != id {
					t.Errorf("= %s, %v", farm, err)
				}
				return
			}
			body := wantCode(t, err, http.StatusBadRequest, httpx.CodeValidation)
			if !slices.Equal(body.Issues, tc.want) {
				t.Errorf("issues = %+v,\nwant %+v", body.Issues, tc.want)
			}
		})
	}
}

// Every read in one InReadTx: the head, then the feed form's stock reads
// only for a farm with an open cycle (at most 5).
func TestDeliveryFormStatements(t *testing.T) {
	f := newFormFixture(t)
	var st formStatements
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
	pc.ConnConfig.Tracer = &st
	pool, err := pgxpool.NewWithConfig(t.Context(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
	st.take()
	svc := NewService(db.NewStore(pool), f.d.Clock)

	run := func(farm uuid.UUID, want ...string) {
		t.Helper()
		if _, err := svc.DeliveryForm(t.Context(), f.mgr, farm); err != nil {
			t.Fatal(err)
		}
		got := st.take()
		if len(got) < 2 || !strings.HasPrefix(got[0], "begin") || !strings.Contains(got[0], "repeatable read") ||
			!strings.Contains(got[0], "read only") || got[len(got)-1] != "commit" {
			t.Fatalf("statements %q: want one repeatable read, read only transaction", got)
		}
		if reads := got[1 : len(got)-1]; !slices.Equal(reads, want) {
			t.Errorf("reads %q,\nwant %q", reads, want)
		}
	}
	run(f.hirbovat, "FormsDeliveryHead", "CalcSettings", "CalcStockEnd", "CalcFeedDayComplete", "CalcConsumption")
	run(f.telita, "FormsDeliveryHead")
}

// The handler: query → service → JSON, errors through the envelope.
func TestDeliveryFormHandler(t *testing.T) {
	f := newFormFixture(t)
	mux := http.NewServeMux()
	NewHandler(NewService(db.NewStore(f.d.App), f.d.Clock)).Routes(mux, httpx.NewChain())
	serve := func(p *auth.Principal, query string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/api/forms/delivery?"+query, nil)
		if p != nil {
			r = r.WithContext(auth.WithPrincipal(r.Context(), *p))
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	q := "farmId=" + f.hirbovat.String()
	w := serve(&f.mgr, q)
	var res DeliveryFormResponse
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &res) != nil || len(res.Stock) != 3 || res.Cycle == nil ||
		w.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("200: %d %s", w.Code, w.Body)
	}
	if w := serve(&f.mgr, q+"&date=2026-11-18"); w.Code != http.StatusBadRequest ||
		!strings.Contains(w.Body.String(), `"path":"date","code":"unknown_field"`) {
		t.Errorf("date: %d %s", w.Code, w.Body)
	}
	if w := serve(&f.vet, q); w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), `"code":"FORBIDDEN"`) {
		t.Errorf("vet: %d %s", w.Code, w.Body)
	}
	if w := serve(nil, q); w.Code != http.StatusUnauthorized {
		t.Errorf("no session: %d %s", w.Code, w.Body)
	}
}
