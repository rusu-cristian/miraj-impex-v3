package entries_test

import (
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// T8 on the dev seed: Ion Rusu corrects H5.1 ♀ of Dr. Ceban's Hrană zilnică
// of 18.11.2026 (2 457,9 kg, M3) by −2 g: the Motiv sentence and the Toast
// name the g change, the save's kg drops by 2 g × 4 264 hens at the start of
// the day (H5.1 ♀, M5 T4) to 2 449,4 kg; his undo puts the save back.
func TestUpdateFeedDailySeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	svc := entries.NewService(db.NewStore(d.App), d.Clock)
	ion := seededUser(t, d.App, "ion.rusu")
	var (
		id   uuid.UUID
		farm uuid.UUID
	)
	if err := d.App.QueryRow(ctx, `select id, farm_id from entry where kind = 'feed_daily' and date = '2026-11-18'`).Scan(&id, &farm); err != nil {
		t.Fatal(err)
	}
	s, err := entries.LoadStored(ctx, d.App, id, dbq.EntryKindFeedDaily, farm, testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	v := s.Feed()
	h51 := slices.IndexFunc(s.Sections, func(x entries.EntrySection) bool { return x.Code == "H5.1" })
	i := slices.IndexFunc(v.Rows, func(r entries.FeedDailyRow) bool { return r.SectionID == s.Sections[h51].ID && r.Sex == dbq.SexF })
	old := *v.Rows[i].G
	g := num.Of(old.Decimal().Sub(num.FromInt(2).Decimal()))
	v.Rows[i].G = &g
	put := func(reason string) (entries.Result, error) {
		t.Helper()
		var at time.Time
		if err := d.App.QueryRow(ctx, `select coalesce(updated_at, created_at) from entry where id = $1`, id).Scan(&at); err != nil {
			t.Fatal(err)
		}
		req := entries.FeedDailyUpdateRequest{IfVersion: at.UTC().Format("2006-01-02T15:04:05.000000Z07:00"), Values: v}
		if reason != "" {
			req.Reason = &reason
		}
		return svc.UpdateFeedDaily(ctx, ion, id, req)
	}
	label := messages.FeedLabel("H5.1", dbq.SexF, messages.FieldG)
	from, to := messages.ShownDec(old.Decimal(), 1), messages.ShownDec(g.Decimal(), 1)

	_, err = put("")
	var he httpx.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("err %v, want a 400", err)
	}
	if status, body := he.HTTPError(); status != http.StatusBadRequest || len(body.Issues) != 1 ||
		body.Issues[0].Message != messages.ReasonMissingChange(label, from, to) {
		t.Errorf("%d %+v", status, body)
	}

	res, err := put("Cântar refăcut.")
	resp := decodeResult[entries.EntryWriteResponse](t, res, err)
	if resp.Toast != messages.CorrectionSaved(label, from, to) {
		t.Errorf("toast %q", resp.Toast)
	}
	summary := func() string {
		var s string
		if err := d.App.QueryRow(ctx, `select summary from audit_log where entry_id = $1 order by at desc, id desc limit 1`, id).Scan(&s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	if got := summary(); got != "Hrană zilnică · 8 din 8 rânduri · 2\u202f449,4 kg" {
		t.Errorf("summary %q", got)
	}
	res, err = svc.UndoEntry(ctx, ion, id)
	resp = decodeResult[entries.EntryWriteResponse](t, res, err)
	if resp.Toast != messages.CorrectionUndone(label, to, from) {
		t.Errorf("undo toast %q", resp.Toast)
	}
	if got := summary(); got != "Hrană zilnică · 8 din 8 rânduri · 2\u202f457,9 kg" {
		t.Errorf("undo summary %q", got)
	}
}
