package entries

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
)

// T4 criterion 1: every level-1 issue of a feed save, its path, code and
// exact message (M3 D5, H1, Q9).
func TestFeedDailyValidate(t *testing.T) {
	farm, h11, h52, ft := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	refs := FeedRefs{
		Sections:  map[uuid.UUID]string{h11: "H1.1", h52: "H5.2"},
		FeedTypes: map[uuid.UUID]bool{ft: true},
	}
	g := func(s string) *num.Num { n := num.MustParse(s); return &n }
	ok := FeedDailyRow{SectionID: h11, Sex: dbq.SexF, FeedTypeID: ft, G: g("160")}
	with := func(edit func(r *FeedDailyRow)) []FeedDailyRow {
		r := ok
		edit(&r)
		return []FeedDailyRow{r}
	}
	const invalid = "Datele trimise nu sunt valide."
	issue := func(path, code, msg string) httpx.Issue { return httpx.Issue{Path: path, Code: code, Message: msg} }
	for _, tc := range []struct {
		name string
		req  FeedDailyCreateRequest
		want []httpx.Issue
	}{
		{"valid", FeedDailyCreateRequest{FarmID: farm, Date: nov18, Rows: []FeedDailyRow{ok}}, nil},
		{"valid: one decimal, bounds", FeedDailyCreateRequest{FarmID: farm, Date: nov18, Rows: []FeedDailyRow{
			{SectionID: h11, Sex: dbq.SexF, FeedTypeID: ft, G: g("1")},
			{SectionID: h11, Sex: dbq.SexM, FeedTypeID: ft, G: g("300")},
			{SectionID: h52, Sex: dbq.SexF, FeedTypeID: ft, G: g("160.50")},
		}}, nil},
		{"farmId, date and rows missing", FeedDailyCreateRequest{}, []httpx.Issue{
			issue("farmId", "required", invalid), issue("date", "required", invalid), issue("rows", "required", invalid),
		}},
		{"source", FeedDailyCreateRequest{WriteMeta: httpx.WriteMeta{Source: "sms"}, FarmID: farm, Date: nov18,
			Rows: []FeedDailyRow{ok}}, []httpx.Issue{issue("source", "out_of_range", invalid)}},
		{"empty rows", FeedDailyCreateRequest{FarmID: farm, Date: nov18, Rows: []FeedDailyRow{}},
			[]httpx.Issue{issue("rows", "required", invalid)}},
		{"H1 above", FeedDailyCreateRequest{FarmID: farm, Date: nov18, Rows: with(func(r *FeedDailyRow) { r.G = g("350") })},
			[]httpx.Issue{issue("rows[0].g", "out_of_range", "H1.1 ♀: 350 g e peste limită. Scrie o valoare între 1 și 300 g.")}},
		{"H1 below", FeedDailyCreateRequest{FarmID: farm, Date: nov18, Rows: with(func(r *FeedDailyRow) { r.G = g("0.5") })},
			[]httpx.Issue{issue("rows[0].g", "out_of_range", "H1.1 ♀: 0,5 g e sub limită. Scrie o valoare între 1 și 300 g.")}},
		{"Q9 two decimals", FeedDailyCreateRequest{FarmID: farm, Date: nov18, Rows: with(func(r *FeedDailyRow) { r.G = g("160.25") })},
			[]httpx.Issue{issue("rows[0].g", "out_of_range", invalid)}},
		{"g missing", FeedDailyCreateRequest{FarmID: farm, Date: nov18, Rows: with(func(r *FeedDailyRow) { r.G = nil })},
			[]httpx.Issue{issue("rows[0].g", "required", invalid)}},
		{"sex missing", FeedDailyCreateRequest{FarmID: farm, Date: nov18, Rows: with(func(r *FeedDailyRow) { r.Sex = "" })},
			[]httpx.Issue{issue("rows[0].sex", "required", invalid)}},
		{"sex not F/M; H1 then has no section sentence", FeedDailyCreateRequest{FarmID: farm, Date: nov18,
			Rows: with(func(r *FeedDailyRow) { r.Sex, r.G = "X", g("350") })},
			[]httpx.Issue{issue("rows[0].sex", "out_of_range", invalid), issue("rows[0].g", "out_of_range", invalid)}},
		{"sectionId missing", FeedDailyCreateRequest{FarmID: farm, Date: nov18,
			Rows: with(func(r *FeedDailyRow) { r.SectionID = uuid.Nil })},
			[]httpx.Issue{issue("rows[0].sectionId", "required", invalid)}},
		{"section of another farm", FeedDailyCreateRequest{FarmID: farm, Date: nov18,
			Rows: with(func(r *FeedDailyRow) { r.SectionID = uuid.New() })},
			[]httpx.Issue{issue("rows[0].sectionId", "out_of_range", invalid)}},
		{"feedTypeId missing", FeedDailyCreateRequest{FarmID: farm, Date: nov18,
			Rows: with(func(r *FeedDailyRow) { r.FeedTypeID = uuid.Nil })},
			[]httpx.Issue{issue("rows[0].feedTypeId", "required", invalid)}},
		{"unknown feed type", FeedDailyCreateRequest{FarmID: farm, Date: nov18,
			Rows: with(func(r *FeedDailyRow) { r.FeedTypeID = uuid.New() })},
			[]httpx.Issue{issue("rows[0].feedTypeId", "out_of_range", invalid)}},
		{"a section × sex twice", FeedDailyCreateRequest{FarmID: farm, Date: nov18, Rows: []FeedDailyRow{ok, ok}},
			[]httpx.Issue{issue("rows[1]", "out_of_range", invalid)}},
		{"no farm: the section and type checks wait for it", FeedDailyCreateRequest{Date: nov18,
			Rows: []FeedDailyRow{{SectionID: uuid.New(), Sex: dbq.SexM, FeedTypeID: uuid.New(), G: g("350")}}},
			[]httpx.Issue{issue("farmId", "required", invalid), issue("rows[0].g", "out_of_range", invalid)}},
		{"every issue, in document order", FeedDailyCreateRequest{FarmID: farm, Date: nov18, Rows: []FeedDailyRow{
			{SectionID: h52, Sex: dbq.SexM, FeedTypeID: ft, G: g("301")},
			{SectionID: uuid.New(), Sex: dbq.SexF, FeedTypeID: uuid.New(), G: nil},
		}}, []httpx.Issue{
			issue("rows[0].g", "out_of_range", "H5.2 ♂: 301 g e peste limită. Scrie o valoare între 1 și 300 g."),
			issue("rows[1].sectionId", "out_of_range", invalid),
			issue("rows[1].feedTypeId", "out_of_range", invalid),
			issue("rows[1].g", "required", invalid),
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var v httpx.Validator
			tc.req.Validate(&v, refs)
			if got := v.Issues(); !slices.Equal(got, tc.want) {
				t.Errorf("issues = %+v\nwant %+v", got, tc.want)
			}
			err := v.Err()
			if len(tc.want) == 0 {
				if err != nil {
					t.Errorf("err = %v", err)
				}
				return
			}
			status, body := errBody(t, err)
			if status != 400 || body.Code != "VALIDATION" || body.Message != tc.want[0].Message {
				t.Errorf("err = %d %s %q", status, body.Code, body.Message)
			}
		})
	}
}
