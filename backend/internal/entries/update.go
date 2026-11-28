package entries

import (
	"context"
	"net/http"
	"slices"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/ledger"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// routeUpdateEntry is the edit or Corectează of a saved entry, decided by the
// server (§7.8, plan M5-D5).
const routeUpdateEntry = "PUT /api/entries/{id}"

// updateRoutes mounts PUT /api/entries/{id} behind authed (Routes calls it,
// plan M5-D1).
func (h *Handler) updateRoutes(mux httpx.Router, authed httpx.Chain) {
	mux.Handle(routeUpdateEntry, authed.ThenFunc(h.updateEntry))
}

// updateEntry is PUT /api/entries/{id} (§7.8): the path id → role (viewer
// 403 FORBIDDEN, before any read) → the stored entry (loadUpdateTarget: 404,
// 403 FARM_FORBIDDEN) → the kind's own decode of the body (`values` in the
// create shape of the kind, unknown members refused with their path, 64 KiB)
// → level-1 Validate → service. The answer is sent as the runner encoded it.
// Each served kind (Q1: daily, feed_daily, delivery) has its adapter
// (kindAdapter.update).
func (h *Handler) updateEntry(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	id, err := parseEntryPath(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := checkWriter(p, uuid.Nil); err != nil {
		httpx.WriteError(w, err)
		return
	}
	t, err := h.svc.loadUpdateTarget(r.Context(), p, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := t.adapter.update(h, w, r, p, t)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteRaw(w, res.Status, res.Body)
}

// updateTarget is the stored entry a PUT names, read before its body is
// decoded: the kind picks the shape of `values` (its adapter), the farm and
// date are what `values` must repeat (D101).
type updateTarget struct {
	ID      uuid.UUID
	Kind    dbq.EntryKind
	FarmID  uuid.UUID
	Date    timex.Day
	adapter kindAdapter
}

// loadUpdateTarget reads entry id for a PUT by p with checkEntryHead (ARC-4):
// no entry → 404; a farm of it outside p's farms → 403 FARM_FORBIDDEN; a
// kind M5 doesn't serve (Q1), or deleted → 404. The change runner checks
// everything again under the locks.
func (s *Service) loadUpdateTarget(ctx context.Context, p auth.Principal, id uuid.UUID) (updateTarget, error) {
	h, err := checkEntryHead(ctx, s.store.Q(), p, id, false)
	if err != nil {
		return updateTarget{}, err
	}
	return updateTarget{ID: id, Kind: h.Kind, FarmID: h.FarmID, Date: h.Date, adapter: h.adapter}, nil
}

// allowUpdate is who may PUT (plan M5-D5, Q10): the author on the creation
// day (Editable: an edit, no Motiv) or a manager on own farms and the owner
// (Correctable: a correction, the Motiv checked by the plan, whose sentence
// names the change); anyone else 403 (changeForbidden: a vet's own entry of
// an earlier day gets the S16 sentence).
func allowUpdate(p auth.Principal) func(EntryMeta, EntryActions) error {
	return func(e EntryMeta, a EntryActions) error {
		if a.Editable || a.Correctable {
			return nil
		}
		return changeForbidden(p, e)
	}
}

// checkUpdateHead is the part of an update's plan every kind runs first,
// under the locks: `values` names the stored farm and date (D101; the
// handler checked them on its pre-read, a service caller may not have) and
// ifVersion is still the entry's version (§7.0 → 409 STALE; UpdateHeader
// guards the write again).
func checkUpdateHead(tx *ChangeTx, farmID uuid.UUID, date timex.Day, ifVersion string) error {
	var v httpx.Validator
	v.Check(farmID == tx.Entry.FarmID, "values.farmId", httpx.IssueOutOfRange, messages.InvalidData())
	v.Check(date == tx.Entry.Date, "values.date", httpx.IssueOutOfRange, messages.InvalidData())
	if err := v.Err(); err != nil {
		return err
	}
	at, err := ParseVersion(ifVersion)
	if err != nil {
		return httpx.Validation(httpx.Issue{Path: "ifVersion", Code: httpx.IssueBadType, Message: messages.InvalidData()})
	}
	if !at.Equal(tx.Entry.VersionAt) {
		return httpx.Stale()
	}
	return nil
}

// birdKey is one live Zilnic bird value of a save: a section × sex × type
// (bird_event_daily_uq without the date, which a save fixes).
type birdKey struct {
	SectionID uuid.UUID
	Sex       dbq.Sex
	Type      dbq.BirdEventType
}

// moveOf is the bird move of a change from the rows before to the rows
// after: per section × sex whose Σ qty differs, after − before, in the order
// the pairs first appear in after, then in before. sections and the farm's
// move follow; qty and sexes are left empty (the delete Dialog's figures).
func moveOf(before, after []SnapshotBird) birdMove {
	m := birdMove{codes: map[uuid.UUID]string{}, delta: map[ledger.SectionSex]int{}}
	var order []ledger.SectionSex
	add := func(rows []SnapshotBird, sign int) {
		for _, b := range rows {
			pair := ledger.SectionSex{SectionID: b.SectionID, Sex: b.Sex}
			if _, ok := m.delta[pair]; !ok {
				order = append(order, pair)
				m.delta[pair] = 0
			}
			m.delta[pair] += sign * int(b.Qty)
			m.codes[b.SectionID] = b.Section
		}
	}
	add(after, +1)
	add(before, -1)
	for _, p := range order {
		d := m.delta[p]
		if d == 0 {
			delete(m.delta, p)
			continue
		}
		m.pairs = append(m.pairs, p)
		if code := m.codes[p.SectionID]; !slices.Contains(m.sections, code) {
			m.sections = append(m.sections, code)
		}
		if p.Sex == dbq.SexF {
			m.move.F += d
		} else {
			m.move.M += d
		}
	}
	return m
}

// readAlive fills f's counts before a change's writes (rowFigures): the
// farm and every section at the end of today and, in a dry run on an
// earlier day, every section at the end of the entry date (the preview's
// alive rows). birds = whether any bird value changes (a Toast names a
// section's birds even when the count doesn't move).
func readAlive(ctx context.Context, tx *ChangeTx, f *rowFigures, birds bool) error {
	if !birds {
		return nil
	}
	e := tx.Entry
	day, err := calc.ReadFarmDay(ctx, tx.Tx, e.FarmID, tx.Today)
	if err != nil {
		return err
	}
	f.farm = day.AliveEnd
	if f.today, err = sectionsAliveAt(ctx, tx, tx.Today); err != nil {
		return err
	}
	f.atDate = f.today
	if tx.DryRun && e.Date != tx.Today {
		if f.atDate, err = sectionsAliveAt(ctx, tx, e.Date); err != nil {
			return err
		}
	}
	return nil
}

// signedSums is Σ signed qty per sex of rows ("morți ♀ −5", §7.8).
func signedSums(rows []SnapshotBird) calc.PerSex[int] {
	var s calc.PerSex[int]
	for _, b := range rows {
		if b.Sex == dbq.SexF {
			s.F += int(b.Qty)
		} else {
			s.M += int(b.Qty)
		}
	}
	return s
}

// moveEffects are the preview's effect lines of a bird move (§7.8, §6.15,
// Q7): per moved section × sex "H5.2 · găini în viață de pe 17.11.2026 se
// schimbă cu −2.", then per moved sex the save's Σ signed qty before and
// after "Mortalitate · morți ♀ −5 → −7", then per moved sex the farm at the
// end of today "Hîrbovăț · ♀ azi 12 563 → 12 561".
func moveEffects(f rowFigures, date timex.Day, farm string, before, after calc.PerSex[int]) []string {
	out := []string{}
	m := f.birds
	for _, p := range m.pairs {
		out = append(out, messages.EffectAliveFrom(m.codes[p.SectionID], p.Sex, date, m.delta[p]))
	}
	sexes := []struct {
		sex          dbq.Sex
		move, b, a   int
		farm, farmTo int
	}{
		{dbq.SexF, m.move.F, before.F, after.F, f.farm.F, f.farm.F + m.move.F},
		{dbq.SexM, m.move.M, before.M, after.M, f.farm.M, f.farm.M + m.move.M},
	}
	for _, s := range sexes {
		if s.move != 0 {
			out = append(out, messages.EffectDeaths(s.sex, s.b, s.a))
		}
	}
	for _, s := range sexes {
		if s.move != 0 {
			out = append(out, messages.EffectFarmAliveToday(farm, s.sex, s.farm, s.farmTo))
		}
	}
	return out
}
