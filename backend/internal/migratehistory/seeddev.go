package migratehistory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/textkey"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// devUser is one --seed-dev account (M1-D7).
type devUser struct {
	login, name string
	role        dbq.UserRole
	farms       []string // farm keys; none for the owner (owners see every farm)
}

// devUsers are the §11.9 dev accounts. The names "Ion Rusu" and "Dr. Ceban"
// are the authors the illustrative events name.
var devUsers = []devUser{
	{"proprietar", "Proprietar", dbq.UserRoleOwner, nil},
	{"ion.rusu", "Ion Rusu", dbq.UserRoleManager, []string{FarmTelita, FarmHirbovat}},
	{"dr.ceban", "Dr. Ceban", dbq.UserRoleVet, []string{FarmHirbovat}},
	{"observator", "Observator exemplu", dbq.UserRoleViewer, []string{FarmTelita, FarmHirbovat}},
}

// Credential is a seeded login and its one-time password.
type Credential struct {
	Login, Name string
	Role        dbq.UserRole
	Password    string
}

// writeCredentials prints each login and password once (§11.9: shown once,
// never logged).
func writeCredentials(w io.Writer, creds []Credential) error {
	if w == nil || len(creds) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, "Dev accounts (login, password, name; the passwords are shown only now):"); err != nil {
		return err
	}
	for _, c := range creds {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s (%s)\n", c.Login, c.Password, c.Name, c.Role); err != nil {
			return err
		}
	}
	return nil
}

// illustrative-events-hirbovat.json, the parts D7 seeds.
type eventsFile struct {
	Culls []struct {
		Date      timex.Day `json:"date"`
		Section   string    `json:"section"`
		Sex       dbq.Sex   `json:"sex"`
		Qty       int32     `json:"qty"`
		Reason    string    `json:"reason"`
		Note      *string   `json:"note"`
		Carved    int32     `json:"carved_from_mortality_of_day"`
		Left      int32     `json:"mortality_left"`
		CreatedBy string    `json:"created_by"`
		CreatedAt string    `json:"created_at"`
	} `json:"culls"`
	Weighings []struct {
		Date          timex.Day           `json:"date"`
		Section       string              `json:"section"`
		Sex           dbq.Sex             `json:"sex"`
		AvgWeightG    decimal.Decimal     `json:"avg_weight_g"`
		CvPct         decimal.NullDecimal `json:"cv_pct"`
		UniformityPct decimal.NullDecimal `json:"uniformity_pct"`
		SampleSize    *int32              `json:"sample_size"`
		CreatedBy     string              `json:"created_by"`
		CreatedAt     string              `json:"created_at"`
	} `json:"weighings"`
	Corrections []struct {
		At     string `json:"at"`
		User   string `json:"user"`
		Record string `json:"record"`
		Action string `json:"action"`
		Row    struct {
			Date    timex.Day `json:"date"`
			Section string    `json:"section"`
			Sex     dbq.Sex   `json:"sex"`
		} `json:"row"`
		EntryAt   string `json:"entry_at"`
		EntryUser string `json:"entry_user"`
		Old       struct {
			Qty int32 `json:"qty"`
		} `json:"old_values"`
		New struct {
			Qty int32 `json:"qty"`
		} `json:"new_values"`
		Reason string `json:"reason"`
	} `json:"corrections"`
	EventLog []struct {
		At       string `json:"at"`
		User     string `json:"user"`
		Record   string `json:"record"`
		Action   string `json:"action"`
		Sections []struct {
			Section string `json:"section"`
		} `json:"sections"`
	} `json:"event_log"`
}

const localLayout = "2006-01-02T15:04"

func localTime(s string) (time.Time, error) {
	t, err := time.ParseInLocation(localLayout, s, timex.Chisinau)
	if err != nil {
		return time.Time{}, fmt.Errorf("migratehistory: illustrative events: time %q: %w", s, err)
	}
	return t, nil
}

// seeder layers the dev seed on a plan.
type seeder struct {
	p     *Plan
	farm  uuid.UUID // Hîrbovăț: every illustrative event is there
	users map[string]uuid.UUID
	codes map[uuid.UUID]string
}

// seedDev adds the --seed-dev content of M1-D7 to the plan: the dev users,
// and from illustrative-events-hirbovat.json the culls, the two weighings,
// the event-log records as their own authored entries and the 17.11
// correction with its two audit rows. Counts and goldens stay the same. The
// users' passwords are devPassword when set (tests), else generated.
//
// The file's one sexing error (H6.2 ♀ 19.05.2026, carved out of that day's
// deaths) is not seeded: a sexing error is not a death (§5.5), so carving it
// would take golden 10 to ♀ 1 520 · 5,65 % (user answer Q-T7-1, 27.11.2026).
func seedDev(ctx context.Context, p *Plan, dir string, now time.Time, devPassword string) ([]Credential, error) {
	var ev eventsFile
	if err := readJSON(filepath.Join(dir, "illustrative-events-hirbovat.json"), &ev); err != nil {
		return nil, err
	}
	farm, ok := p.FarmID(FarmHirbovat)
	if !ok {
		return nil, errors.New("migratehistory: seed: no Hîrbovăț farm")
	}
	s := &seeder{p: p, farm: farm, users: map[string]uuid.UUID{}, codes: map[uuid.UUID]string{}}
	for _, sec := range p.Sections {
		s.codes[sec.ID] = sec.Code
	}
	creds, err := s.addUsers(ctx, now, devPassword)
	if err != nil {
		return nil, err
	}
	for _, step := range []func(*eventsFile) error{s.culls, s.weighings, s.eventLog, s.correction} {
		if err := step(&ev); err != nil {
			return nil, err
		}
	}
	if err := p.validate(); err != nil {
		return nil, err
	}
	return creds, nil
}

func (s *seeder) addUsers(ctx context.Context, now time.Time, devPassword string) ([]Credential, error) {
	hasher := auth.NewHasher()
	var creds []Credential
	for _, u := range devUsers {
		password := devPassword
		if password == "" {
			var err error
			if password, err = auth.GeneratePassword(); err != nil {
				return nil, err
			}
		}
		hash, err := hasher.Hash(ctx, password)
		if err != nil {
			return nil, err
		}
		id := uuid.New()
		s.p.Users = append(s.p.Users, dbq.AppUser{
			ID: id, Name: u.name, Login: u.login, PasswordHash: hash, Role: u.role, Active: true,
			PasswordGeneratedAt: now, CreatedAt: now, UpdatedAt: now,
		})
		for _, f := range u.farms {
			fid, ok := s.p.FarmID(f)
			if !ok {
				return nil, fmt.Errorf("migratehistory: seed: unknown farm %q", f)
			}
			s.p.UserFarms = append(s.p.UserFarms, dbq.UserFarm{UserID: id, FarmID: fid})
		}
		s.users[u.name] = id
		creds = append(creds, Credential{Login: u.login, Name: u.name, Role: u.role, Password: password})
	}
	return creds, nil
}

func (s *seeder) user(name string) (*uuid.UUID, error) {
	id, ok := s.users[name]
	if !ok {
		return nil, fmt.Errorf("migratehistory: seed: unknown author %q", name)
	}
	return &id, nil
}

// entry adds an authored entry at Hîrbovăț (source manual: a form save).
func (s *seeder) entry(kind dbq.EntryKind, day timex.Day, author, at string) (uuid.UUID, error) {
	by, err := s.user(author)
	if err != nil {
		return uuid.Nil, err
	}
	t, err := localTime(at)
	if err != nil {
		return uuid.Nil, err
	}
	e := dbq.Entry{ID: uuid.New(), Kind: kind, FarmID: s.farm, Date: day, Source: dbq.RecordSourceManual, CreatedBy: by, CreatedAt: t}
	s.p.Entries = append(s.p.Entries, e)
	return e.ID, nil
}

// author marks an existing entry as saved by a user at a time.
func (s *seeder) author(id uuid.UUID, name, at string) error {
	by, err := s.user(name)
	if err != nil {
		return err
	}
	t, err := localTime(at)
	if err != nil {
		return err
	}
	for i := range s.p.Entries {
		if s.p.Entries[i].ID == id {
			s.p.Entries[i].CreatedBy, s.p.Entries[i].CreatedAt, s.p.Entries[i].Source = by, t, dbq.RecordSourceManual
			return nil
		}
	}
	return fmt.Errorf("migratehistory: seed: entry %s not planned", id)
}

// mortality returns the index of the Hîrbovăț mortality row of a
// compartment-day.
func (s *seeder) mortality(code string, day timex.Day, sex dbq.Sex) (int, error) {
	for i, e := range s.p.BirdEvents {
		if e.FarmID == s.farm && e.Type == dbq.BirdEventTypeMortality && e.Date == day && e.Sex == sex && s.codes[e.SectionID] == code {
			return i, nil
		}
	}
	return 0, fmt.Errorf("migratehistory: seed: no mortality row %s %s %s", code, day, sex)
}

// culls carves each illustrative cull out of its compartment-day's deaths:
// the mortality row keeps mortality_left and the cull is its own daily entry
// by its author (M1-D7). Deaths + culls stay the register's deaths.
func (s *seeder) culls(ev *eventsFile) error {
	for _, c := range ev.Culls {
		i, err := s.mortality(c.Section, c.Date, c.Sex)
		if err != nil {
			return err
		}
		row := &s.p.BirdEvents[i]
		if -row.Qty != c.Carved || c.Carved-c.Qty != c.Left {
			return fmt.Errorf("migratehistory: seed: cull %s %s %s: deaths %d, carved %d − %d ≠ left %d", c.Section, c.Date, c.Sex, -row.Qty, c.Carved, c.Qty, c.Left)
		}
		reason := dbq.CullReason(textkey.Key(c.Reason))
		if !slices.Contains([]dbq.CullReason{dbq.CullReasonBoala, dbq.CullReasonSchiopatare, dbq.CullReasonAltul}, reason) {
			return fmt.Errorf("migratehistory: seed: cull reason %q", c.Reason)
		}
		row.Qty = -c.Left
		entry, err := s.entry(dbq.EntryKindDaily, c.Date, c.CreatedBy, c.CreatedAt)
		if err != nil {
			return err
		}
		s.p.BirdEvents = append(s.p.BirdEvents, dbq.BirdEvent{
			ID: uuid.New(), EntryID: entry, FarmID: s.farm, SectionID: row.SectionID, Date: c.Date, Sex: c.Sex,
			Type: dbq.BirdEventTypeCull, Qty: -c.Qty, CullReason: &reason, Note: c.Note,
		})
	}
	return nil
}

// weighings adds each illustrative weighing as its own weighing entry.
func (s *seeder) weighings(ev *eventsFile) error {
	for _, w := range ev.Weighings {
		sec, ok := s.p.SectionID(FarmHirbovat, w.Section)
		if !ok {
			return fmt.Errorf("migratehistory: seed: unknown section %q", w.Section)
		}
		entry, err := s.entry(dbq.EntryKindWeighing, w.Date, w.CreatedBy, w.CreatedAt)
		if err != nil {
			return err
		}
		s.p.Weighings = append(s.p.Weighings, dbq.Weighing{
			ID: uuid.New(), EntryID: entry, FarmID: s.farm, SectionID: sec, Date: w.Date, Sex: w.Sex,
			AvgWeightG: w.AvgWeightG, CvPct: w.CvPct, UniformityPct: w.UniformityPct, SampleSize: w.SampleSize,
		})
	}
	return nil
}

// eventLog gives each event-log record its author and time: the 22.10
// transfer and the 7 sales keep their entries; the 12.11 and 13.11
// deliveries and the Zilnic mortality, Zilnic eggs and Hrană zilnică of
// 16–18.11 move off the migration entry into their own entries.
func (s *seeder) eventLog(ev *eventsFile) error {
	for _, r := range ev.EventLog {
		if r.Action != "create" {
			return fmt.Errorf("migratehistory: seed: event %s: action %q", r.At, r.Action)
		}
		at, err := localTime(r.At)
		if err != nil {
			return err
		}
		day := timex.DayOf(at)
		switch r.Record {
		case "transfer", "sale":
			kind := dbq.EntryKind(r.Record)
			var sections []string
			for _, sec := range r.Sections {
				sections = append(sections, sec.Section)
			}
			id, err := s.findEntry(kind, day, sections)
			if err != nil {
				return err
			}
			if err := s.author(id, r.User, r.At); err != nil {
				return err
			}
		case "feed_delivery":
			e, err := s.entry(dbq.EntryKindDelivery, day, r.User, r.At)
			if err != nil {
				return err
			}
			if n := s.move(e, func(t string, farm uuid.UUID, d timex.Day) bool {
				return t == "feed_delivery" && farm == s.farm && d == day
			}); n != 1 {
				return fmt.Errorf("migratehistory: seed: %d Hîrbovăț deliveries on %s, want 1", n, day)
			}
		case "mortality", "eggs", "feed":
			kind, table := dbq.EntryKindDaily, map[string]string{"mortality": "mortality", "eggs": "egg_production", "feed": "feed_allocation"}[r.Record]
			if r.Record == "feed" {
				kind = dbq.EntryKindFeedDaily
			}
			e, err := s.entry(kind, day, r.User, r.At)
			if err != nil {
				return err
			}
			if n := s.move(e, func(t string, farm uuid.UUID, d timex.Day) bool {
				return t == table && farm == s.farm && d == day
			}); n == 0 {
				return fmt.Errorf("migratehistory: seed: no Hîrbovăț %s rows on %s", r.Record, day)
			}
		default:
			return fmt.Errorf("migratehistory: seed: event %s: record %q", r.At, r.Record)
		}
	}
	return nil
}

// findEntry finds the planned Hîrbovăț entry of kind on day whose bird rows
// are on exactly the given sections (any section for an empty list).
func (s *seeder) findEntry(kind dbq.EntryKind, day timex.Day, sections []string) (uuid.UUID, error) {
	slices.Sort(sections)
	for _, e := range s.p.Entries {
		if e.Kind != kind || e.FarmID != s.farm || e.Date != day {
			continue
		}
		var got []string
		for _, b := range s.p.BirdEvents {
			if b.EntryID == e.ID && b.Qty <= 0 && !slices.Contains(got, s.codes[b.SectionID]) {
				got = append(got, s.codes[b.SectionID])
			}
		}
		slices.Sort(got)
		if len(sections) == 0 || slices.Equal(got, sections) {
			return e.ID, nil
		}
	}
	return uuid.Nil, fmt.Errorf("migratehistory: seed: no %s entry on %s for %s", kind, day, strings.Join(sections, ", "))
}

// move re-points the migration entry's rows that match onto entry and
// returns how many moved. Only the mortality rows of bird_event move.
func (s *seeder) move(entry uuid.UUID, match func(table string, farm uuid.UUID, day timex.Day) bool) int {
	from := s.p.migrationEntry[FarmHirbovat]
	n := 0
	for i := range s.p.BirdEvents {
		r := &s.p.BirdEvents[i]
		if r.EntryID == from && r.Type == dbq.BirdEventTypeMortality && match("mortality", r.FarmID, r.Date) {
			r.EntryID = entry
			n++
		}
	}
	for i := range s.p.EggProductions {
		r := &s.p.EggProductions[i]
		if r.EntryID == from && match("egg_production", r.FarmID, r.Date) {
			r.EntryID = entry
			n++
		}
	}
	for i := range s.p.FeedAllocations {
		r := &s.p.FeedAllocations[i]
		if r.EntryID == from && match("feed_allocation", r.FarmID, r.Date) {
			r.EntryID = entry
			n++
		}
	}
	for i := range s.p.FeedDeliveries {
		r := &s.p.FeedDeliveries[i]
		if r.EntryID == from && match("feed_delivery", r.FarmID, r.Date) {
			r.EntryID = entry
			n++
		}
	}
	return n
}

// auditChange is one `changes` item of an audit row (§4.10).
type auditChange struct {
	Key     string  `json:"key"`
	Section string  `json:"section"`
	Sex     dbq.Sex `json:"sex"`
	Field   string  `json:"field"`
	Old     int32   `json:"old"`
	New     int32   `json:"new"`
}

// correction applies the shared illustrative correction (RF08, D158a): the
// vet's 17.11 Zilnic is corrected by the manager on 18.11, so the entry is
// `corrected` and has two audit rows, its create and the update with the
// reason and the changed value. The row already holds the new (real) value.
func (s *seeder) correction(ev *eventsFile) error {
	for _, c := range ev.Corrections {
		if c.Record != "mortality" || c.Action != "update" || c.Row.Sex != dbq.SexF {
			return fmt.Errorf("migratehistory: seed: correction %s: only a ♀ mortality update is supported", c.At)
		}
		i, err := s.mortality(c.Row.Section, c.Row.Date, c.Row.Sex)
		if err != nil {
			return err
		}
		row := s.p.BirdEvents[i]
		if -row.Qty != c.New.Qty {
			return fmt.Errorf("migratehistory: seed: correction %s %s: row holds %d, the correction's new value is %d", c.Row.Section, c.Row.Date, -row.Qty, c.New.Qty)
		}
		created, err := localTime(c.EntryAt)
		if err != nil {
			return err
		}
		updated, err := localTime(c.At)
		if err != nil {
			return err
		}
		author, err := s.user(c.EntryUser)
		if err != nil {
			return err
		}
		corrector, err := s.user(c.User)
		if err != nil {
			return err
		}
		var entry *dbq.Entry
		for j := range s.p.Entries {
			if s.p.Entries[j].ID == row.EntryID {
				entry = &s.p.Entries[j]
			}
		}
		if entry == nil || entry.Kind != dbq.EntryKindDaily || !entry.CreatedAt.Equal(created) || *entry.CreatedBy != *author {
			return fmt.Errorf("migratehistory: seed: correction %s: the row's entry is not %s's Zilnic of %s", c.At, c.EntryUser, c.EntryAt)
		}
		entry.Corrected, entry.UpdatedBy, entry.UpdatedAt = true, corrector, &updated
		changes, err := json.Marshal([]auditChange{{
			Key: messages.AuditKeyHensDead(c.Row.Section), Section: c.Row.Section, Sex: c.Row.Sex,
			Field: "mortality", Old: c.Old.Qty, New: c.New.Qty,
		}})
		if err != nil {
			return err
		}
		kind, day, reason := dbq.EntryKindDaily, c.Row.Date, c.Reason
		base := dbq.InsertEntryAuditParams{
			EntryID: entry.ID, EntryKind: &kind, RecordTypes: []string{"mortalitate"},
			FarmIds: []uuid.UUID{s.farm}, RecordDate: &day, Source: dbq.RecordSourceManual,
		}
		create, update := base, base
		create.At, create.UserID, create.Action = created, author, dbq.AuditActionCreate
		update.At, update.UserID, update.Action, update.Reason, update.Changes = updated, corrector, dbq.AuditActionUpdate, &reason, changes
		s.p.Audit = append(s.p.Audit, create, update)
	}
	return nil
}
