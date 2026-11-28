package golden

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/ledger"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// M5 rows (plan M5-D11), read-only so Verify can run them inside the import
// transaction. Both read the golden farm's Zilnic values of 17.11.2026 as
// the Zilnic form state shows them saved (entries.BuildDailyForm): the dev
// seed holds them in Dr. Ceban's corrected 17.11 save, the bare import in
// the farm's migration entry, and the form reads the day's live rows
// whichever entry holds them.
//
// Golden 33 runs Z6 on those values (entries.MortalitySpikes, the function
// the save runs). Golden 37 is the correction of H5.2 ♀ 1 → 3: the diff of
// the values with the old mortality and the stored ones (entries.DiffDaily,
// the correction's diff), the §6.1 check of H5.2 ♀ from 17.11 to AsOf
// (ledger.CheckCounts) and H5.2's birds at the end of 17.11 and 18.11
// (calc.SectionsAliveEnd). The write path itself is proved by the seeded
// correction chain (internal/entries, correction_seeded_test.go).

// spikeDay is the day of goldens 33 and 37.
var spikeDay = day(2026, time.November, 17)

// Spike is golden 33: one Z6 warning of a Zilnic save, its section × sex,
// the typed deaths, the 7-day mean at the message's one decimal and the
// warning's message.
type Spike struct {
	Section string
	Sex     string
	Deaths  int
	Mean7   string // 1 dp
	Message string
}

// Correction is golden 37: a correction of one Zilnic mortality, the
// section × sex's birds at the end of the entry's day and of AsOf after it,
// how many section × sex pairs §6.1 finds below 0 over those days, and the
// audit change the correction records.
type Correction struct {
	Section    string
	Sex        string
	AliveEnd17 int
	AliveEnd18 int
	Negative   int
	Change     ValueChange
}

// ValueChange is one audit change of a count: its §4.10 key, old and new.
type ValueChange struct {
	Key      string
	Old, New int
}

// The correction golden 37 describes (§11.7, S16 boards).
const (
	correctedSection = "H5.2"
	// illustrativeOld is the mortality before the correction (§11.7
	// "1 → 3 (illustrative)"; illustrative-events-hirbovat.json
	// corrections[0].old_values): the workbook holds only the corrected 3,
	// so a bare import, whose 17.11 values are the register's, has no audit
	// row to read it from.
	illustrativeOld = 1
)

// savedDay is a farm's Zilnic values of one day as the form state shows them
// saved: the create shape, each section's code and the entries holding the
// values.
type savedDay struct {
	Values   entries.DailyValues
	Codes    map[uuid.UUID]string
	Existing []entries.DailyFormEntry
}

// readSavedDay reads the golden farm's saved Zilnic values of d through the
// form state (as the owner, as of AsOf and Now).
func readSavedDay(ctx context.Context, db dbq.DBTX, farm uuid.UUID, d timex.Day) (savedDay, error) {
	form, err := entries.BuildDailyForm(ctx, db, entries.DailyFormInput{
		Principal: ownerAt, FarmID: farm, Date: d, Today: AsOf, Now: Now,
	})
	if err != nil {
		return savedDay{}, err
	}
	out := savedDay{
		Values:   entries.DailyValues{FarmID: farm, Date: d, Sections: []entries.DailySection{}, Eggs: []entries.DailyEggs{}},
		Codes:    map[uuid.UUID]string{},
		Existing: form.Existing,
	}
	for _, h := range form.Halls {
		for _, s := range h.Sections {
			out.Codes[s.SectionID] = s.Code
			if s.Saved == nil {
				continue
			}
			out.Values.Sections = append(out.Values.Sections, entries.DailySection{SectionID: s.SectionID, F: s.Saved.F, M: s.Saved.M})
		}
	}
	return out, nil
}

// mortalitySpikes reads golden 33: Z6 on the golden farm's saved Zilnic
// values of 17.11.2026.
func mortalitySpikes(ctx context.Context, env Env) (any, error) {
	farm, err := FarmID(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	saved, err := readSavedDay(ctx, env.DB, farm, spikeDay)
	if err != nil {
		return nil, err
	}
	spikes, err := entries.MortalitySpikes(ctx, env.DB, farm, spikeDay, saved.Values)
	if err != nil {
		return nil, err
	}
	out := make([]Spike, 0, len(spikes))
	for _, s := range spikes {
		out = append(out, Spike{Section: s.Section, Sex: string(s.Sex), Deaths: s.Deaths, Mean7: fixed(&s.Mean7, 1), Message: s.Message})
	}
	return out, nil
}

// correction reads golden 37 on the golden farm's saved Zilnic values of
// 17.11.2026, whose H5.2 ♀ mortality is the corrected one (new). The old
// value is the `old` of that key in the latest audit row of the entry
// holding H5.2's values, which must be an update (the seed's Corectează);
// when the register holds the value (the import's migration entry, no
// author) it is the illustrative old value of §11.7.
func correction(ctx context.Context, env Env) (any, error) {
	farm, err := FarmID(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	saved, err := readSavedDay(ctx, env.DB, farm, spikeDay)
	if err != nil {
		return nil, err
	}
	var section uuid.UUID
	for id, code := range saved.Codes {
		if code == correctedSection {
			section = id
		}
	}
	i := slices.IndexFunc(saved.Values.Sections, func(s entries.DailySection) bool { return s.SectionID == section })
	if section == uuid.Nil || i < 0 || saved.Values.Sections[i].F == nil || saved.Values.Sections[i].F.Mortality == nil {
		return nil, fmt.Errorf("%s ♀ has no mortality saved on %s", correctedSection, spikeDay)
	}
	key := messages.AuditKeyDaily(correctedSection, dbq.SexF, messages.FieldMortality)
	old, err := oldMortality(ctx, env.DB, saved.Existing, key)
	if err != nil {
		return nil, err
	}

	// The values before the correction: the stored ones with the old mortality.
	before := saved.Values
	before.Sections = slices.Clone(saved.Values.Sections)
	f := *before.Sections[i].F
	n := num.FromInt(int64(old))
	f.Mortality = &n
	before.Sections[i].F = &f
	changes := entries.DiffDaily(before, saved.Values, saved.Codes)
	if len(changes) != 1 {
		return nil, fmt.Errorf("the correction diff has %d changes, want 1: %+v", len(changes), changes)
	}
	change, err := countChange(changes[0])
	if err != nil {
		return nil, err
	}

	negative, err := ledger.CheckCounts(ctx, env.DB, []ledger.SectionSex{{SectionID: section, Sex: dbq.SexF}}, spikeDay, AsOf)
	if err != nil {
		return nil, err
	}
	out := Correction{Section: correctedSection, Sex: string(dbq.SexF), Negative: len(negative), Change: change}
	for _, d := range []struct {
		day   timex.Day
		alive *int
	}{{spikeDay, &out.AliveEnd17}, {AsOf, &out.AliveEnd18}} {
		secs, err := calc.SectionsAliveEnd(ctx, env.DB, farm, d.day)
		if err != nil {
			return nil, err
		}
		j := slices.IndexFunc(secs, func(s calc.SectionAlive) bool { return s.SectionID == section })
		if j < 0 {
			return nil, fmt.Errorf("calc.SectionsAliveEnd has no %s on %s", correctedSection, d.day)
		}
		*d.alive = secs[j].AliveEnd.F
	}
	return out, nil
}

// oldMortality is the value of key before the correction: from the latest
// audit row of the entry holding the corrected section's values (an update
// with that key in its changes), or illustrativeOld when the register holds
// them (an entry without an author: the import).
func oldMortality(ctx context.Context, db dbq.DBTX, existing []entries.DailyFormEntry, key string) (int, error) {
	i := slices.IndexFunc(existing, func(e entries.DailyFormEntry) bool { return slices.Contains(e.Sections, correctedSection) })
	if i < 0 {
		return 0, fmt.Errorf("no entry holds the values of %s on %s", correctedSection, spikeDay)
	}
	holder := existing[i]
	if holder.Author == nil {
		return illustrativeOld, nil
	}
	a, err := dbq.New(db).EntryLatestAudit(ctx, holder.EntryID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("the entry %s holding %s on %s has no audit row", holder.EntryID, correctedSection, spikeDay)
	}
	if err != nil {
		return 0, err
	}
	if a.Action != dbq.AuditActionUpdate {
		return 0, fmt.Errorf("the latest audit row of the entry holding %s on %s is %s, want update", correctedSection, spikeDay, a.Action)
	}
	var changes []audit.Change
	if err := json.Unmarshal(a.Changes, &changes); err != nil {
		return 0, fmt.Errorf("the changes of audit row %d: %w", a.ID, err)
	}
	for _, c := range changes {
		if c.Key == key {
			ch, err := countChange(c)
			return ch.Old, err
		}
	}
	return 0, fmt.Errorf("audit row %d has no change %q", a.ID, key)
}

// countChange is an audit change of a count as a ValueChange; a text or null
// side, or a fraction, is an error.
func countChange(c audit.Change) (ValueChange, error) {
	side := func(v audit.Value) (int, error) {
		n, ok := v.Num()
		if !ok || !n.Decimal().IsInteger() {
			return 0, fmt.Errorf("change %q: a side is not a count", c.Key)
		}
		return int(n.Decimal().IntPart()), nil
	}
	old, err := side(c.Old)
	if err != nil {
		return ValueChange{}, err
	}
	new, err := side(c.New)
	if err != nil {
		return ValueChange{}, err
	}
	return ValueChange{Key: c.Key, Old: old, New: new}, nil
}
