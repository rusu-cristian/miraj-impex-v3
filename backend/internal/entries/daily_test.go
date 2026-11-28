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
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/pgerr"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/ledger"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// dailyFixture is the feed fixture (a laying farm "Hîrbovăț" with halls H5
// {H5.1, H5.2}, H6 {H6.1}, H7 {H7.1 sold out on 14.11}, the retired H8, and
// the growth farm "Telița" in the closed Seria 2025 until 10.11) plus H6.2
// under H6 and 273 more roosters in H5.2. Birds at the start of 18.11:
//
//	H5.1 ♀ 1 000 · ♂ 100   H5.2 ♀ 500 · ♂ 323   H6.1 ♀ 200   H6.2 ♀ 100 · ♂ 10
//
// so the farm ends 18.11 with ♀ 1 800 · ♂ 433 before any Zilnic.
type dailyFixture struct {
	*feedFixture
	misc uuid.UUID // a migration entry holding the extra rows
}

func newDailyFixture(t *testing.T) *dailyFixture {
	t.Helper()
	f := &dailyFixture{feedFixture: newFeedFixture(t)}
	f.sec["H6.2"] = f.id(`insert into section (farm_id, code, parent_id, sort_order) values ($1, 'H6.2', $2, 1) returning id`,
		f.farmA, f.sec["H6"])
	f.misc = f.id(`insert into entry (kind, farm_id, date, source) values ('migration', $1, $2, 'migration') returning id`, f.farmA, nov1)
	f.bird(f.misc, "H6.2", nov1, "F", "placement", 100)
	f.bird(f.misc, "H6.2", nov1, "M", "placement", 10)
	f.bird(f.misc, "H5.2", nov1, "M", "placement", 273)
	return f
}

func dnum(v int64) *num.Num { x := num.FromInt(v); return &x }

func (f *dailyFixture) values(d timex.Day, sections []DailySection, eggs ...DailyEggs) DailyValues {
	if sections == nil {
		sections = []DailySection{}
	}
	return DailyValues{FarmID: f.farmA, Date: d, Sections: sections, Eggs: eggs}
}

func (f *dailyFixture) saveDaily(p auth.Principal, v DailyValues, meta httpx.WriteMeta) (Result, error) {
	f.t.Helper()
	return f.svc.CreateDaily(context.Background(), p, DailyCreateRequest{WriteMeta: meta, DailyValues: v})
}

// toastSections is the §6.2 example on the fixture: 4 sections, ♀ −3 · ♂ 0.
func (f *dailyFixture) toastSections() []DailySection {
	return []DailySection{
		{SectionID: f.sec["H5.1"], F: &DailySex{Mortality: dnum(1)}, M: &DailySex{Mortality: dnum(0)}},
		{SectionID: f.sec["H5.2"], NoLosses: true},
		{SectionID: f.sec["H6.1"], NoLosses: true, F: &DailySex{SexingError: nil}},
		{SectionID: f.sec["H6.2"], F: &DailySex{Mortality: dnum(2)}, M: &DailySex{SexingError: dnum(0)}},
	}
}

// toastSpikes are the Z6 warnings of toastSections: no deaths in the 7 days
// before, so a 0,0 mean warns on any death (D53).
const toastSpikes = `[{"code":"MORTALITY_SPIKE","sectionCode":"H5.1","sex":"F","message":"H5.1: 1 găină moartă, de peste 3 ori media pe 7 zile (0,0). Verifică numărătoarea; poți salva."},{"code":"MORTALITY_SPIKE","sectionCode":"H6.2","sex":"F","message":"H6.2: 2 găini moarte, de peste 3 ori media pe 7 zile (0,0). Verifică numărătoarea; poți salva."}]`

func dailyDecode[T any](t *testing.T, b []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("%s: %v", b, err)
	}
	return v
}

// T5 criterion 1: every level-1 issue with its path and sentence.
func TestDailyValidate(t *testing.T) {
	h51, h5, other := uuid.New(), uuid.New(), uuid.New()
	refs := DailyRefs{Sections: map[uuid.UUID]string{h51: "H5.1", h5: "H5"}}
	farm := uuid.New()
	day := timex.MustDay(2026, time.November, 18)
	reason := func(s string) *string { return &s }
	dec := func(s string) *num.Num { x := num.MustParse(s); return &x }
	base := func(secs []DailySection, eggs ...DailyEggs) DailyCreateRequest {
		return DailyCreateRequest{DailyValues: DailyValues{FarmID: farm, Date: day, Sections: secs, Eggs: eggs}}
	}
	for _, tc := range []struct {
		name string
		req  DailyCreateRequest
		want []httpx.Issue
	}{
		{"valid", base([]DailySection{{SectionID: h51, F: &DailySex{Mortality: dnum(1), Cull: dnum(2), CullReason: reason("boala"),
			SexingError: dnum(-3)}}}, DailyEggs{SectionID: h5, Cat1: dnum(13), AvgWeightG: dec("67.04")}), nil},
		{"missing farm, date and values", DailyCreateRequest{}, []httpx.Issue{
			{Path: "farmId", Code: "required", Message: messages.InvalidData()},
			{Path: "date", Code: "required", Message: messages.InvalidData()},
			{Path: "sections", Code: "required", Message: messages.InvalidData()},
		}},
		{"only empty rows", base([]DailySection{{SectionID: h51, F: &DailySex{}}}, DailyEggs{SectionID: h5}), []httpx.Issue{
			{Path: "sections", Code: "required", Message: messages.InvalidData()},
		}},
		{"unknown, missing and repeated sections", base([]DailySection{
			{SectionID: other, NoLosses: true}, {NoLosses: true}, {SectionID: h51, NoLosses: true}, {SectionID: h51, NoLosses: true},
		}), []httpx.Issue{
			{Path: "sections[0].sectionId", Code: "out_of_range", Message: messages.InvalidData()},
			{Path: "sections[1].sectionId", Code: "required", Message: messages.InvalidData()},
			{Path: "sections[3]", Code: "out_of_range", Message: messages.InvalidData()},
		}},
		{"Z2", base([]DailySection{{SectionID: h51, F: &DailySex{Mortality: dec("1.5"), Cull: dnum(-1), CullReason: reason("boala"),
			SexingError: dec("0.5")}, M: &DailySex{Mortality: dnum(3_000_000_000)}}}), []httpx.Issue{
			{Path: "sections[0].F.mortality", Code: "not_integer", Message: "H5.1 ♀: scrie un număr întreg."},
			{Path: "sections[0].F.cull", Code: "out_of_range", Message: "H5.1 ♀: scrie un număr întreg."},
			{Path: "sections[0].F.sexingError", Code: "not_integer", Message: "H5.1 ♀: scrie un număr întreg."},
			{Path: "sections[0].M.mortality", Code: "out_of_range", Message: messages.InvalidData()},
		}},
		{"cull reason", base([]DailySection{{SectionID: h51, F: &DailySex{Mortality: dnum(1), CullReason: reason("boala")},
			M: &DailySex{Cull: dnum(1), CullReason: reason("boală")}}}), []httpx.Issue{
			{Path: "sections[0].F.cullReason", Code: "out_of_range", Message: messages.InvalidData()},
			{Path: "sections[0].M.cullReason", Code: "out_of_range", Message: messages.InvalidData()},
		}},
		{"noLosses with a typed removal", base([]DailySection{{SectionID: h51, NoLosses: true, M: &DailySex{Cull: dnum(0)}}}), []httpx.Issue{
			{Path: "sections[0].noLosses", Code: "out_of_range", Message: messages.InvalidData()},
		}},
		{"noLosses with a sexing error", base([]DailySection{{SectionID: h51, NoLosses: true, M: &DailySex{SexingError: dnum(2)}}}), nil},
		{"Z8", base(nil, DailyEggs{SectionID: h5, Floor: dnum(-1), Cat1: dec("2.5"), AvgWeightG: dec("130")},
			DailyEggs{SectionID: h51, AvgWeightG: dec("67.045")}, DailyEggs{SectionID: h5, Cat2: dnum(1)}), []httpx.Issue{
			{Path: "eggs[0].floor", Code: "out_of_range", Message: "H5: scrie un număr întreg."},
			{Path: "eggs[0].cat1", Code: "not_integer", Message: "H5: scrie un număr întreg."},
			{Path: "eggs[0].avgWeightG", Code: "out_of_range",
				Message: "H5: masa medie 130 g e în afara limitei. Scrie o valoare între 20 și 120 g."},
			{Path: "eggs[1].avgWeightG", Code: "out_of_range", Message: messages.InvalidData()},
			{Path: "eggs[2]", Code: "out_of_range", Message: messages.InvalidData()},
		}},
		{"bad source", DailyCreateRequest{WriteMeta: httpx.WriteMeta{Source: "radio"},
			DailyValues: DailyValues{FarmID: farm, Date: day, Sections: []DailySection{{SectionID: h51, NoLosses: true}}}},
			[]httpx.Issue{{Path: "source", Code: "out_of_range", Message: messages.InvalidData()}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var v httpx.Validator
			tc.req.Validate(&v, refs)
			if got := v.Issues(); !slices.Equal(got, tc.want) {
				t.Errorf("issues = %+v\nwant %+v", got, tc.want)
			}
		})
	}
	// PUT validates the same values under "values.".
	var v httpx.Validator
	DailyValues{FarmID: farm, Date: day, Sections: []DailySection{{SectionID: other, NoLosses: true}}}.ValidateAt(&v, refs, "values.")
	if got := v.Issues(); len(got) != 1 || got[0].Path != "values.sections[0].sectionId" {
		t.Errorf("prefixed issues = %+v", got)
	}
}

// T5 criterion 4: the §6.2 Toast example. 201 with entry, Toast, alive and
// asOf; "Fără pierderi azi" writes mortality 0 and cull 0 of the sexes present
// only (H6.1 has no roosters); one audit row, entry.created and
// counts.changed.
func TestCreateDailySaves(t *testing.T) {
	ctx := context.Background()
	f := newDailyFixture(t)
	res, err := f.saveDaily(f.ion, f.values(nov18, f.toastSections()), httpx.WriteMeta{})
	if err != nil {
		t.Fatal(err)
	}
	resp := dailyDecode[EntryWriteResponse](t, res.Body)
	want := `{"entry":{"id":"` + resp.Entry.ID.String() + `","kind":"daily","version":"2026-11-18T10:00:00.000000Z"},` +
		`"warnings":` + toastSpikes + `,"toast":"Zilnic salvat · 4 secțiuni · ♀ −3 · ♂ 0","alive":{"F":1797,"M":433},` +
		`"asOf":"2026-11-18T12:00:00+02:00"}`
	if res.Status != http.StatusCreated || strings.TrimSuffix(string(res.Body), "\n") != want {
		t.Errorf("got %d %s\nwant %s", res.Status, res.Body, want)
	}

	rs, err := f.d.App.Query(ctx, `select s.code || ' ' || b.sex || ' ' || b.type || ' ' || b.qty || coalesce(' ' || b.cull_reason, '')
		from bird_event b join section s on s.id = b.section_id where b.entry_id = $1 order by 1`, resp.Entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := pgx.CollectRows(rs, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	wantRows := []string{
		"H5.1 F mortality -1", "H5.1 M mortality 0",
		"H5.2 F cull 0", "H5.2 F mortality 0", "H5.2 M cull 0", "H5.2 M mortality 0",
		"H6.1 F cull 0", "H6.1 F mortality 0",
		"H6.2 F mortality -2", "H6.2 M sexing_error 0",
	}
	if !slices.Equal(rows, wantRows) {
		t.Errorf("rows = %q\nwant %q", rows, wantRows)
	}

	var (
		action, summary  string
		types            []string
		newValues        []byte
		userID           uuid.UUID
		changes          string
		entryID, farmIDs = uuid.UUID{}, []uuid.UUID{}
	)
	if err := f.d.App.QueryRow(ctx, `select action::text, summary, record_types, new_values, user_id, changes::text, entry_id, farm_ids
		from audit_log`).Scan(&action, &summary, &types, &newValues, &userID, &changes, &entryID, &farmIDs); err != nil {
		t.Fatal(err)
	}
	if action != "create" || summary != "Mortalitate · 4 secțiuni" || !slices.Equal(types, []string{"mortalitate", "sacrificare", "eroare_sexare"}) ||
		userID != f.ion.UserID || changes != "[]" || entryID != resp.Entry.ID || !slices.Equal(farmIDs, []uuid.UUID{f.farmA}) {
		t.Errorf("audit = %s %q %v %s %s", action, summary, types, userID, changes)
	}
	nv := dailyDecode[dailyAuditValues](t, newValues)
	if len(nv.Sections) != 4 || nv.Sections[2].Section != "H6.1" || nv.Sections[2].M != nil || nv.Sections[2].F == nil ||
		nv.Sections[2].F.Mortality == nil || nv.Sections[2].F.Mortality.Decimal().Sign() != 0 || nv.FarmID != f.farmA {
		t.Errorf("new_values = %s", newValues)
	}

	rs, err = f.d.App.Query(ctx, `select type || ' ' || payload::text from live_event order by id`)
	if err != nil {
		t.Fatal(err)
	}
	events, err := pgx.CollectRows(rs, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || !strings.HasPrefix(events[0], "entry.created ") ||
		!strings.Contains(events[0], `"sections": ["H5.1", "H5.2", "H6.1", "H6.2"]`) ||
		!strings.Contains(events[0], `"recordTypes": ["mortalitate", "sacrificare", "eroare_sexare"]`) ||
		!strings.Contains(events[0], `"summary": "Mortalitate · 4 secțiuni"`) ||
		!strings.HasPrefix(events[1], "counts.changed ") || !strings.Contains(events[1], `"alive": {"F": 1797, "M": 433}`) ||
		!strings.Contains(events[1], `"fromDate": "2026-11-18"`) {
		t.Errorf("events = %q", events)
	}
}

// T5 criterion 4: the same clientRequestId twice → one entry, the same bytes
// with 200; a dry run writes nothing.
func TestCreateDailyReplayAndDryRun(t *testing.T) {
	f := newDailyFixture(t)
	key := uuid.New()
	v := f.values(nov18, f.toastSections())
	dry, err := f.saveDaily(f.ion, v, httpx.WriteMeta{ClientRequestID: &key, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if dry.Status != http.StatusOK || strings.TrimSuffix(string(dry.Body), "\n") != `{"effects":[],"warnings":`+toastSpikes+`,"refusals":[]}` {
		t.Errorf("dry run = %d %s", dry.Status, dry.Body)
	}
	if got := [3]int{f.count("entry where kind = 'daily'"), f.count("audit_log"), f.count("bird_event where type <> 'placement' and type <> 'sale'")}; got != [3]int{} {
		t.Errorf("a dry run wrote %v", got)
	}
	first, err := f.saveDaily(f.ion, v, httpx.WriteMeta{ClientRequestID: &key})
	if err != nil {
		t.Fatal(err)
	}
	again, err := f.saveDaily(f.ion, v, httpx.WriteMeta{ClientRequestID: &key})
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != http.StatusCreated || again.Status != http.StatusOK || !bytes.Equal(first.Body, again.Body) {
		t.Errorf("save %d, replay %d:\n%s\n%s", first.Status, again.Status, first.Body, again.Body)
	}
	if got := f.count("entry where kind = 'daily'"); got != 1 {
		t.Errorf("%d daily entries, want 1", got)
	}
}

// T5 criterion 4 (Q5): an egg-only save has its own Toast and summary, the
// chip oua, total_eggs generated, and no counts.changed.
func TestCreateDailyEggsOnly(t *testing.T) {
	ctx := context.Background()
	f := newDailyFixture(t)
	w := num.MustParse("67.04")
	v := f.values(nov18, nil,
		DailyEggs{SectionID: f.sec["H5"], Floor: dnum(98), Cat1: dnum(13), Cat2: dnum(417), Cat3: dnum(27), Hatching: dnum(3415), AvgWeightG: &w},
		DailyEggs{SectionID: f.sec["H6"], Cat1: dnum(10), Cat2: dnum(2000), Cat3: dnum(20), Hatching: dnum(12)},
		DailyEggs{SectionID: f.sec["H7"]}, // empty: not written, not checked
	)
	res, err := f.saveDaily(f.ion, v, httpx.WriteMeta{})
	if err != nil {
		t.Fatal(err)
	}
	resp := dailyDecode[EntryWriteResponse](t, res.Body)
	if res.Status != http.StatusCreated || resp.Toast != "Zilnic salvat · 2 hale · 5\u202f914 ouă" ||
		resp.Alive == nil || *resp.Alive != (httpx.SexPair[int]{F: 1800, M: 433}) {
		t.Errorf("got %d %s", res.Status, res.Body)
	}
	var total int
	var weight string
	if err := f.d.App.QueryRow(ctx, `select total_eggs, avg_egg_weight_g::text from egg_production where section_id = $1`,
		f.sec["H5"]).Scan(&total, &weight); err != nil || total != 3872 || weight != "67.04" {
		t.Errorf("H5 eggs: total %d, weight %s, %v", total, weight, err)
	}
	var summary string
	var types []string
	if err := f.d.App.QueryRow(ctx, `select summary, record_types from audit_log`).Scan(&summary, &types); err != nil ||
		summary != "Ouă · 2 hale" || !slices.Equal(types, []string{"oua"}) {
		t.Errorf("audit %q %v %v", summary, types, err)
	}
	if got := f.count("live_event where type = 'counts.changed'"); got != 0 || f.count("live_event") != 1 {
		t.Errorf("counts.changed = %d, events %d", got, f.count("live_event"))
	}
}

// T5 criterion 1: Z1, a future date, a closed series.
func TestCreateDailyChecks(t *testing.T) {
	f := newDailyFixture(t)
	one := []DailySection{{SectionID: f.sec["H5.1"], F: &DailySex{Mortality: dnum(1)}}}

	_, err := f.saveDaily(f.vet, f.values(nov17, one), httpx.WriteMeta{})
	if body := wantCode(t, err, http.StatusForbidden, httpx.CodePastDayVet); body.Message != "Pentru zilele trecute cere managerului." {
		t.Errorf("Z1 = %q", body.Message)
	}
	_, err = f.saveDaily(f.ion, f.values(nov19, one), httpx.WriteMeta{})
	body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused)
	if want := "19.11.2026 e în viitor. Alege o dată până la 18.11.2026."; body.Message != want ||
		len(body.Refusals) != 1 || body.Refusals[0].Field != "date" {
		t.Errorf("future = %+v", body)
	}
	_, err = f.saveDaily(f.vet, f.values(nov19, one), httpx.WriteMeta{}) // the future first, for the vet too
	wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused)

	closed := DailyValues{FarmID: f.farmB, Date: timex.MustDay(2026, time.November, 5),
		Sections: []DailySection{{SectionID: f.sec["T1"], NoLosses: true}}}
	_, err = f.saveDaily(f.ion, closed, httpx.WriteMeta{})
	if body := wantCode(t, err, http.StatusConflict, httpx.CodeSeriesClosed); !strings.HasPrefix(body.Message, "Seria 2025 e închisă") {
		t.Errorf("closed = %q", body.Message)
	}
	_, err = f.saveDaily(f.viewer, f.values(nov18, one), httpx.WriteMeta{})
	wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden)
	_, err = f.saveDaily(f.outsider, f.values(nov18, one), httpx.WriteMeta{})
	wantCode(t, err, http.StatusForbidden, httpx.CodeFarmForbidden)
	if res, err := f.saveDaily(f.vet, f.values(nov18, one), httpx.WriteMeta{}); err != nil || res.Status != http.StatusCreated {
		t.Errorf("vet today: %d %v", res.Status, err)
	}
	if f.count("entry where kind = 'daily'") != 1 {
		t.Errorf("daily entries = %d", f.count("entry where kind = 'daily'"))
	}
}

// T5 criterion 1: Z3 (no birds, a hall, retired, "Fără pierderi azi" without
// birds), Z4, Z5 (the §6.1 sentence), Z7 (a compartment), eggs without hens,
// all together in document order; the first refusal's text is the message
// (Z5's SaveBar text when it comes first); a dry run answers them in a 200.
func TestCreateDailyRefusals(t *testing.T) {
	f := newDailyFixture(t)
	reason := "boala"
	v := f.values(nov18, []DailySection{
		{SectionID: f.sec["H5.1"], F: &DailySex{Mortality: dnum(1)}},                                                            // fine
		{SectionID: f.sec["H7.1"], F: &DailySex{Mortality: dnum(0)}},                                                            // Z3: sold out on 14.11
		{SectionID: f.sec["H5"], M: &DailySex{Mortality: dnum(1)}},                                                              // Z3: a hall
		{SectionID: f.sec["H8"], NoLosses: true},                                                                                // Z3: retired
		{SectionID: f.sec["H6.1"], M: &DailySex{Cull: dnum(1), CullReason: &reason}, F: &DailySex{Cull: dnum(2)}},               // Z4 ♀, Z3 ♂
		{SectionID: f.sec["H5.2"], M: &DailySex{Mortality: dnum(3230)}},                                                         // Z5
		{SectionID: f.sec["H6.2"], M: &DailySex{Mortality: dnum(6), Cull: dnum(6), CullReason: &reason, SexingError: dnum(-5)}}, // Z5: no field alone
	},
		DailyEggs{SectionID: f.sec["H5.1"], Cat1: dnum(1)}, // Z7
		DailyEggs{SectionID: f.sec["H7"], Cat1: dnum(1)},   // no hens since 14.11
		DailyEggs{SectionID: f.sec["H5"], Cat1: dnum(1)},   // fine
	)
	threeTwoThree, zero := 323, 0
	want := []httpx.Refusal{
		{Field: "sections[1].F", SectionCode: "H7.1", Sex: "F", Message: "H7.1 nu are găini pe 18.11.2026."},
		{Field: "sections[2].M", SectionCode: "H5", Sex: "M", Message: "H5 nu are cocoși pe 18.11.2026."},
		{Field: "sections[3].noLosses", SectionCode: "H8", Message: "H8 nu are găini pe 18.11.2026."},
		{Field: "sections[4].F.cullReason", SectionCode: "H6.1", Sex: "F",
			Message: "H6.1 ♀: Motiv lipsă pentru 2 găini sacrificate pe 18.11.2026. Alege motivul: boală, șchiopătare sau altul."},
		{Field: "sections[4].M", SectionCode: "H6.1", Sex: "M", Message: "H6.1 nu are cocoși pe 18.11.2026."},
		{Field: "sections[5].M.mortality", SectionCode: "H5.2", Sex: "M", Max: &threeTwoThree,
			Message: "H5.2 ♂: Mortalitate + Sacrificare + Eroare sexare = 3\u202f230, peste cei 323 cocoși în viață pe 18.11.2026. La Mortalitate scrie cel mult 323."},
		{Field: "sections[6].M.mortality", SectionCode: "H6.2", Sex: "M", Max: &zero,
			Message: "H6.2 ♂: Mortalitate + Sacrificare + Eroare sexare = 17, peste cei 10 cocoși în viață pe 18.11.2026. La Mortalitate scrie cel mult 0."},
		{Field: "eggs[0].sectionId", SectionCode: "H5.1", Message: "Ouăle se scriu pe hală: H5."},
		{Field: "eggs[1].sectionId", SectionCode: "H7", Message: "H7 nu are găini pe 18.11.2026."},
	}
	before := f.rows()
	_, err := f.saveDaily(f.ion, v, httpx.WriteMeta{})
	body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused)
	if body.Message != want[0].Message || !slices.EqualFunc(body.Refusals, want, dailyRefusalEqual) {
		t.Errorf("message %q, refusals:\n%s\nwant\n%s", body.Message, dailyJSON(t, body.Refusals), dailyJSON(t, want))
	}
	if f.rows() != before {
		t.Errorf("rows %v → %v", before, f.rows())
	}

	v.Sections = v.Sections[5:6] // Z5 alone: the SaveBar text is the message
	_, err = f.saveDaily(f.ion, v, httpx.WriteMeta{})
	if body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused); body.Message != "refuzat · H5.2 ♂ 3\u202f230 · cel mult 323" {
		t.Errorf("Z5 message = %q", body.Message)
	}

	// Z7 at a growth farm: one refusal for the egg list.
	g := DailyValues{FarmID: f.farmB, Date: nov18, Eggs: []DailyEggs{{SectionID: f.sec["T1"], Cat1: dnum(1)}, {SectionID: f.sec["T1"], Cat2: dnum(1)}}}
	_, err = f.saveDaily(f.ion, g, httpx.WriteMeta{})
	body = wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused)
	if want := []httpx.Refusal{{Field: "eggs", Message: "Telița e fermă de creștere; ouăle se scriu doar la fermele de ouat."}}; !slices.EqualFunc(body.Refusals, want, dailyRefusalEqual) {
		t.Errorf("growth = %+v", body.Refusals)
	}

	// The dry run: the same refusals in a 200.
	v = f.values(nov18, []DailySection{{SectionID: f.sec["H5.1"]}, {SectionID: f.sec["H7.1"], F: &DailySex{Mortality: dnum(0)}}})
	res, err := f.saveDaily(f.ion, v, httpx.WriteMeta{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	dry := dailyDecode[DryRunResponse](t, res.Body)
	if res.Status != http.StatusOK || !slices.EqualFunc(dry.Refusals, want[:1], dailyRefusalEqual) {
		t.Errorf("dry run = %d %s", res.Status, res.Body)
	}
}

func dailyRefusalEqual(a, b httpx.Refusal) bool { return dailyJSONString(a) == dailyJSONString(b) }

func dailyJSONString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func dailyJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// T5 criterion 1, Z9: a second save of a live key → 409 ALREADY_ENTERED with
// the holder and its time; an imported holder gets the generic sentence; a
// different type of the same section × sex is a different key; eggs too; the
// dry run lists the collision as a refusal.
func TestCreateDailyAlreadyEntered(t *testing.T) {
	f := newDailyFixture(t)
	first, err := f.saveDaily(f.vet, f.values(nov18, []DailySection{{SectionID: f.sec["H5.1"], F: &DailySex{Mortality: dnum(1)}}},
		DailyEggs{SectionID: f.sec["H5"], Cat1: dnum(5)}), httpx.WriteMeta{})
	if err != nil {
		t.Fatal(err)
	}
	holder := dailyDecode[EntryWriteResponse](t, first.Body).Entry.ID

	reason := "altul"
	v := f.values(nov18, []DailySection{
		{SectionID: f.sec["H5.2"], F: &DailySex{Mortality: dnum(1)}},
		{SectionID: f.sec["H5.1"], F: &DailySex{Cull: dnum(1), CullReason: &reason, Mortality: dnum(2)}},
	})
	const msg = "H5.1 ♀ are deja valori introduse pe 18.11.2026 la 12:00. Apasă Editează ca să schimbi valorile."
	before := f.rows()
	_, err = f.saveDaily(f.ion, v, httpx.WriteMeta{})
	body := wantCode(t, err, http.StatusConflict, httpx.CodeAlreadyEntered)
	if body.Message != msg || body.EntryID == nil || *body.EntryID != holder {
		t.Errorf("Z9 = %q %v, want %q %s", body.Message, body.EntryID, msg, holder)
	}
	if f.rows() != before {
		t.Errorf("rows %v → %v", before, f.rows())
	}
	res, err := f.saveDaily(f.ion, v, httpx.WriteMeta{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	dry := dailyDecode[DryRunResponse](t, res.Body)
	if want := []httpx.Refusal{{Field: "sections[1].F", SectionCode: "H5.1", Sex: "F", Message: msg}}; !slices.EqualFunc(dry.Refusals, want, dailyRefusalEqual) {
		t.Errorf("dry run = %s", res.Body)
	}

	// Another key of the same section × sex (a cull) is free.
	if res, err := f.saveDaily(f.ion, f.values(nov18, []DailySection{{SectionID: f.sec["H5.1"], F: &DailySex{Cull: dnum(1), CullReason: &reason}}}),
		httpx.WriteMeta{}); err != nil || res.Status != http.StatusCreated {
		t.Errorf("cull: %d %v", res.Status, err)
	}
	// Eggs.
	_, err = f.saveDaily(f.ion, f.values(nov18, nil, DailyEggs{SectionID: f.sec["H5"], Cat2: dnum(5)}), httpx.WriteMeta{})
	body = wantCode(t, err, http.StatusConflict, httpx.CodeAlreadyEntered)
	if body.Message != "H5 are deja ouăle introduse pe 18.11.2026 la 12:00. Apasă Editează ca să schimbi valorile." || *body.EntryID != holder {
		t.Errorf("eggs Z9 = %q", body.Message)
	}
	// An imported holder.
	f.bird(f.misc, "H6.2", nov18, "M", "mortality", 0)
	_, err = f.saveDaily(f.ion, f.values(nov18, []DailySection{{SectionID: f.sec["H6.2"], NoLosses: true}}), httpx.WriteMeta{})
	body = wantCode(t, err, http.StatusConflict, httpx.CodeAlreadyEntered)
	if body.Message != messages.EntryAlreadySaved() || *body.EntryID != f.misc {
		t.Errorf("imported Z9 = %q %v", body.Message, body.EntryID)
	}
}

// The OnConflict path (M3 D2): after a bird_event_daily_uq violation the
// holder is named on the pool; nothing held → nil.
func TestDailyHolderError(t *testing.T) {
	f := newDailyFixture(t)
	q := dbq.New(f.d.App)
	v := f.values(nov18, []DailySection{{SectionID: f.sec["H5.2"], M: &DailySex{Mortality: dnum(1)}}})
	if err := dailyHolderError(context.Background(), q, v); err != nil {
		t.Errorf("free key: %v", err)
	}
	first, err := f.saveDaily(f.maria, v, httpx.WriteMeta{})
	if err != nil {
		t.Fatal(err)
	}
	holder := dailyDecode[EntryWriteResponse](t, first.Body).Entry.ID
	status, body := errBody(t, dailyHolderError(context.Background(), q, v))
	if status != http.StatusConflict || body.EntryID == nil || *body.EntryID != holder ||
		body.Message != "H5.2 ♂ are deja valori introduse pe 18.11.2026 la 12:00. Apasă Editează ca să schimbi valorile." {
		t.Errorf("holder = %d %+v", status, body)
	}
}

// T5 criterion 2, Z6: 4 vs a mean of 6 ÷ 7 = 0,857… warns, 3 vs 8 ÷ 7 =
// 1,143… doesn't, a 0 mean warns on any death (D53), culls count in the mean
// (Q3); the save returns the same warnings, and saves.
func TestMortalitySpikes(t *testing.T) {
	ctx := context.Background()
	f := newDailyFixture(t)
	for i, d := range []int{10, 11, 12, 13, 14, 15, 16} {
		day := timex.MustDay(2026, time.November, d)
		e := f.id(`insert into entry (kind, farm_id, date) values ('daily', $1, $2) returning id`, f.farmA, day)
		if i < 5 {
			f.bird(e, "H6.1", day, "F", "mortality", -1) // 5 deaths …
		}
		f.bird(e, "H5.2", day, "F", "mortality", -1) // 7 deaths …
		if i == 6 {
			f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, cull_reason)
				values ($1, $2, $3, $4, 'F', 'cull', -1, 'boala'), ($1, $2, $5, $4, 'F', 'cull', -1, 'altul')`,
				e, f.farmA, f.sec["H6.1"], day, f.sec["H5.2"]) // … + 1 cull each: 6 and 8
		}
	}
	v := f.values(nov17, []DailySection{
		{SectionID: f.sec["H5.1"], F: &DailySex{Mortality: dnum(1)}, M: &DailySex{Mortality: dnum(0)}},
		{SectionID: f.sec["H5.2"], F: &DailySex{Mortality: dnum(3)}},
		{SectionID: f.sec["H6.1"], F: &DailySex{Mortality: dnum(4)}},
	})
	spikes, err := MortalitySpikes(ctx, f.d.App, f.farmA, nov17, v)
	if err != nil {
		t.Fatal(err)
	}
	const h61 = "H6.1: 4 găini moarte, de peste 3 ori media pe 7 zile (0,9). Verifică numărătoarea; poți salva."
	const h51 = "H5.1: 1 găină moartă, de peste 3 ori media pe 7 zile (0,0). Verifică numărătoarea; poți salva."
	if len(spikes) != 2 || spikes[0].Section != "H5.1" || spikes[0].Message != h51 || spikes[0].Deaths != 1 || !spikes[0].Mean7.IsZero() ||
		spikes[1].Section != "H6.1" || spikes[1].Sex != dbq.SexF || spikes[1].Deaths != 4 || spikes[1].Message != h61 ||
		num.Round(spikes[1].Mean7, 3).String() != "0.857" {
		t.Errorf("spikes = %+v", spikes)
	}

	res, err := f.saveDaily(f.ion, v, httpx.WriteMeta{})
	if err != nil {
		t.Fatal(err)
	}
	resp := dailyDecode[EntryWriteResponse](t, res.Body)
	want := []httpx.Warning{
		{Code: "MORTALITY_SPIKE", SectionCode: "H5.1", Sex: "F", Message: h51},
		{Code: "MORTALITY_SPIKE", SectionCode: "H6.1", Sex: "F", Message: h61},
	}
	if res.Status != http.StatusCreated || !slices.Equal(resp.Warnings, want) {
		t.Errorf("save = %d %s", res.Status, res.Body)
	}
}

// spikesOf compares at full precision: 3 vs 3 × 1 = 3 doesn't warn, 4 does.
func TestSpikesOfFullPrecision(t *testing.T) {
	id := uuid.New()
	deaths := []calc.SectionDeaths{{SectionID: id, Code: "H1", Week: calc.PerSex[int]{F: 7, M: 7}}}
	v := DailyValues{Sections: []DailySection{{SectionID: id, F: &DailySex{Mortality: dnum(3)}, M: &DailySex{Mortality: dnum(4)}}}}
	got := spikesOf(deaths, decimal.NewFromInt(3), v)
	if len(got) != 1 || got[0].Sex != dbq.SexM || got[0].Message != "H1: 4 cocoși morți, de peste 3 ori media pe 7 zile (1,0). Verifică numărătoarea; poți salva." {
		t.Errorf("spikes = %+v", got)
	}
}

// T5 criterion 3, §6.1 after the insert (Q4): a cull on 13.11 in H7.1, sold
// out on 14.11, passes Z5 (300 at the start of 13.11) but takes 14.11 to
// today below 0 → 422 with the field bound and the "Salvarea ar lăsa …"
// banner; the dry run answers the same refusal in a 200.
func TestCreateDailyCountsBelowZero(t *testing.T) {
	f := newDailyFixture(t)
	reason := "boala"
	nov13 := timex.MustDay(2026, time.November, 13)
	v := f.values(nov13, []DailySection{{SectionID: f.sec["H7.1"], F: &DailySex{Cull: dnum(5), CullReason: &reason}}})
	nov14, zero, low := nov14, 0, -5
	want := []httpx.Refusal{{
		Field: "sections[0].F.cull", SectionCode: "H7.1", Sex: "F", Max: &zero, FirstNegativeDate: &nov14,
		LastNegativeDate: &nov18, MinAliveEnd: &low, Message: "H7.1 ♀ 5: ar lăsa −5 găini azi, 18.11.2026. Scrie cel mult 0.",
	}}
	const banner = "Salvarea ar lăsa H7.1 cu −5 găini pe 14.11.2026 și cu −5 găini azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0."
	before := f.rows()
	_, err := f.saveDaily(f.ion, v, httpx.WriteMeta{})
	body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused)
	if body.Message != banner || !slices.EqualFunc(body.Refusals, want, dailyRefusalEqual) {
		t.Errorf("got %q %s", body.Message, dailyJSON(t, body.Refusals))
	}
	if f.rows() != before || f.count("bird_event where type = 'cull'") != 0 {
		t.Errorf("rows %v → %v", before, f.rows())
	}
	res, err := f.saveDaily(f.ion, v, httpx.WriteMeta{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	dry := dailyDecode[DryRunResponse](t, res.Body)
	if res.Status != http.StatusOK || !slices.EqualFunc(dry.Refusals, want, dailyRefusalEqual) || f.count("bird_event where type = 'cull'") != 0 {
		t.Errorf("dry run = %d %s", res.Status, res.Body)
	}
}

// countRefusals names the first removal that can fix the pair alone, else
// the first with 0; a pair without a removal gets the banner on its sex.
func TestCountRefusals(t *testing.T) {
	sec := uuid.New()
	key := ledger.SectionSex{SectionID: sec, Sex: dbq.SexM}
	today := timex.MustDay(2026, time.November, 18)
	neg := ledger.Negative{SectionID: sec, Code: "H6.2", Sex: dbq.SexM, FirstDate: today, FirstAliveEnd: -3, LastDate: today, MinAliveEnd: -3}
	three := -3
	neg.TodayAliveEnd = &three
	for _, tc := range []struct {
		name     string
		removals []removal
		field    string
		max      *int
		message  string
	}{
		{"sexing error, as a removal", []removal{{messages.FieldSexingError, 4}}, "sections[0].M.sexingError", dailyMax(1),
			"H6.2 ♂ 4: ar lăsa −3 cocoși azi, 18.11.2026. Scrie cel mult 1."},
		{"mortality too small, cull can", []removal{{messages.FieldMortality, 2}, {messages.FieldCull, 4}}, "sections[0].M.cull", dailyMax(1),
			"H6.2 ♂ 4: ar lăsa −3 cocoși azi, 18.11.2026. Scrie cel mult 1."},
		{"no field alone", []removal{{messages.FieldMortality, 1}, {messages.FieldCull, 2}}, "sections[0].M.mortality", dailyMax(0),
			"H6.2 ♂ 1: ar lăsa −3 cocoși azi, 18.11.2026. Scrie cel mult 0."},
		{"no removal", nil, "sections[0].M", nil,
			"Salvarea ar lăsa H6.2 cu −3 cocoși azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := dailyPlan{Pairs: map[ledger.SectionSex]dailyPair{key: {Path: "sections[0].M", Code: "H6.2", Sex: dbq.SexM, Removals: tc.removals}}}
			got, banner := countRefusals(p, []ledger.Negative{neg}, today, messages.ActionSave)
			if len(got) != 1 || got[0].Field != tc.field || got[0].Message != tc.message || dailyJSONString(got[0].Max) != dailyJSONString(tc.max) ||
				banner != "Salvarea ar lăsa H6.2 cu −3 cocoși azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0." {
				t.Errorf("got %+v %q", got, banner)
			}
		})
	}
}

func dailyMax(i int) *int { return &i }

// T5 criterion 6: a save with a key, bird values and eggs is at most 14
// statements besides begin and commit.
func TestCreateDailyStatements(t *testing.T) {
	f := newDailyFixture(t)
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
	v := f.values(nov18, f.toastSections(), DailyEggs{SectionID: f.sec["H5"], Cat1: dnum(13)}, DailyEggs{SectionID: f.sec["H6"], Cat1: dnum(3)})
	if _, err := svc.CreateDaily(context.Background(), f.ion, DailyCreateRequest{WriteMeta: httpx.WriteMeta{ClientRequestID: &key}, DailyValues: v}); err != nil {
		t.Fatal(err)
	}
	got := slices.DeleteFunc(log.take(), func(n string) bool { return n == "?" })
	want := []string{
		"GetReplay", "LockFarm", "EntryIDByClientRequest", "DailySaveSlots", "CalcDeathsWindow", "InsertEntry",
		"DailyInsertBirdEvents", "DailyInsertEggProduction", "NegativeCounts", "CalcFarmSexDay", "InsertAuditLog",
		"EmitLiveEvent", "EmitLiveEvent", "InsertReplay",
	}
	if !slices.Equal(got, want) || len(got) > 14 {
		t.Errorf("statements %q,\nwant %q", got, want)
	}
}

// The handler's order (plan M5-D13): decode → role → farm → codes →
// Validate → service, each answering its own envelope; the save's answer is
// sent byte for byte.
func TestDailyHandler(t *testing.T) {
	f := newDailyFixture(t)
	post := func(p auth.Principal, body string) *httptest.ResponseRecorder {
		t.Helper()
		mux := http.NewServeMux()
		NewHandler(f.svc).dailySaveRoutes(mux, httpx.NewChain())
		req := httptest.NewRequestWithContext(auth.WithPrincipal(context.Background(), p), http.MethodPost,
			"/api/entries/daily", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		return w
	}
	body := func(farm uuid.UUID, sections ...map[string]any) string {
		b, err := json.Marshal(map[string]any{"farmId": farm, "date": "2026-11-18", "sections": sections})
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	h51 := map[string]any{"sectionId": f.sec["H5.1"], "F": map[string]any{"mortality": 1}}
	for _, tc := range []struct {
		name   string
		p      auth.Principal
		body   string
		status int
		code   string
		issue  string
		msg    string
	}{
		{"unknown field", f.ion, `{"sections":[{"F":{"dead":1}}]}`, 400, "VALIDATION", "sections[0].F.dead", ""},
		{"string count", f.ion, `{"sections":[{"F":{"mortality":"1"}}]}`, 400, "VALIDATION", "sections[0].F.mortality", ""},
		{"viewer", f.viewer, body(f.farmA, h51), 403, "FORBIDDEN", "", ""},
		{"another farm", f.outsider, body(f.farmA, h51), 403, "FARM_FORBIDDEN", "", ""},
		{"owner, unknown farm", f.owner, body(uuid.New(), h51), 404, "NOT_FOUND", "", ""},
		{"a section of another farm", f.ion, body(f.farmA, map[string]any{"sectionId": f.sec["T1"], "noLosses": true}),
			400, "VALIDATION", "sections[0].sectionId", ""},
		{"Z2 names the section", f.ion, body(f.farmA, map[string]any{"sectionId": f.sec["H5.1"], "F": map[string]any{"mortality": 1.5}}),
			400, "VALIDATION", "sections[0].F.mortality", "H5.1 ♀: scrie un număr întreg."},
		{"level 2 (Z3)", f.ion, body(f.farmA, map[string]any{"sectionId": f.sec["H7.1"], "F": map[string]any{"mortality": 1}}),
			422, "REFUSED", "", "H7.1 nu are găini pe 18.11.2026."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := post(tc.p, tc.body)
			e := dailyDecode[httpx.ErrorBody](t, w.Body.Bytes()).Error
			if w.Code != tc.status || e.Code != tc.code {
				t.Fatalf("got %d %s %q, want %d %s", w.Code, e.Code, e.Message, tc.status, tc.code)
			}
			if tc.issue != "" && (len(e.Issues) == 0 || e.Issues[0].Path != tc.issue) {
				t.Errorf("issues = %+v, want the first on %s", e.Issues, tc.issue)
			}
			if tc.msg != "" && e.Message != tc.msg {
				t.Errorf("message = %q, want %q", e.Message, tc.msg)
			}
		})
	}
	key := uuid.New()
	b, err := json.Marshal(map[string]any{"farmId": f.farmA, "date": "2026-11-18", "clientRequestId": key,
		"sections": []any{h51}, "eggs": []any{map[string]any{"sectionId": f.sec["H5"], "cat1": 13, "avgWeightG": 67.04}}})
	if err != nil {
		t.Fatal(err)
	}
	first, again := post(f.ion, string(b)), post(f.ion, string(b))
	if first.Code != http.StatusCreated || again.Code != http.StatusOK || !bytes.Equal(first.Body.Bytes(), again.Body.Bytes()) ||
		first.Header().Get("Cache-Control") != "no-store" ||
		!strings.Contains(first.Body.String(), `"toast":"Zilnic salvat · 1 secțiune · ♀ −1 · ♂ 0"`) {
		t.Errorf("save %d, replay %d:\n%s\n%s", first.Code, again.Code, first.Body, again.Body)
	}
}

// A batch insert that takes a live key fails with the key's constraint name,
// which CreateDaily's OnConflict maps after the transaction (M3 D2).
func TestDailyInsertConstraintNames(t *testing.T) {
	ctx := context.Background()
	f := newDailyFixture(t)
	q := dbq.New(f.d.App)
	e := f.id(`insert into entry (kind, farm_id, date) values ('daily', $1, $2) returning id`, f.farmA, nov18)
	f.bird(e, "H5.1", nov18, "F", "mortality", -1)
	f.exec(`insert into egg_production (entry_id, farm_id, section_id, date, cat1) values ($1, $2, $3, $4, 1)`, e, f.farmA, f.sec["H5"], nov18)

	birds, err := jsonRows([]dailyBird{
		{SectionID: f.sec["H5.2"], Sex: dbq.SexF, Type: dbq.BirdEventTypeMortality, Qty: -1},
		{SectionID: f.sec["H5.1"], Sex: dbq.SexF, Type: dbq.BirdEventTypeMortality, Qty: -2},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = q.DailyInsertBirdEvents(ctx, dbq.DailyInsertBirdEventsParams{EntryID: e, FarmID: f.farmA, Date: nov18, Rows: birds})
	if ce, ok := pgerr.AsConstraint(err, pgerr.Unique); !ok || ce.Constraint != constraintBirdDaily {
		t.Errorf("birds: %v", err)
	}
	eggs, err := jsonRows([]dailyEggRow{{SectionID: f.sec["H5"]}})
	if err != nil {
		t.Fatal(err)
	}
	err = q.DailyInsertEggProduction(ctx, dbq.DailyInsertEggProductionParams{EntryID: e, FarmID: f.farmA, Date: nov18, Rows: eggs})
	if ce, ok := pgerr.AsConstraint(err, pgerr.Unique); !ok || ce.Constraint != constraintEggs {
		t.Errorf("eggs: %v", err)
	}
}
