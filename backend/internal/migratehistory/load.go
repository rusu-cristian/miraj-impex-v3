package migratehistory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// ErrAlreadyMigrated is returned by Load when the database already holds a
// `migration` entry and Options.Replace is not set (§10.3: the import runs
// once).
var ErrAlreadyMigrated = errors.New("migratehistory: the database already holds the history migration")

// lockKey serialises concurrent imports (pg_advisory_xact_lock), so two runs
// can't both pass the "no migration entry yet" check.
const lockKey int64 = 0x696c5f6869737472 // "il_histr"

// Options configure Load.
type Options struct {
	// Extracts is the extracts directory (design/data/extracts).
	Extracts string
	// Replace empties every domain table before loading (user answer Q4):
	// dev only, so the caller allows it only when APP_ENV=dev, and it
	// needs a connection as the owner role (il_app can't empty audit_log).
	Replace bool
	// SeedDev adds the dev seed (§11.9, M1-D7) in the same transaction: four
	// users and the illustrative events. Dev only: the caller allows it only
	// when APP_ENV=dev.
	SeedDev bool
	// DevPassword, when set, is every seeded user's password (tests);
	// otherwise each gets a generated one.
	DevPassword string
	// SeedOut receives each seeded login and its password, once, after the
	// commit (the command passes stdout). Nil discards them. Passwords are
	// never logged.
	SeedOut io.Writer

	// extraChecks are added to Verify's checks (tests inject a failing one).
	extraChecks []Check
}

// Beginner starts a transaction: a *pgxpool.Pool or a *pgx.Conn.
type Beginner interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
}

// TableCount is how many rows the import wrote to a table.
type TableCount struct {
	Table string
	Rows  int64
}

// Load runs the history migration (SPEC §10.3): it reads the extracts and
// builds the plan before touching the database, then in one transaction
// writes every row with COPY, one audit_log row per farm (M1-D3), runs Verify
// (M1-D4) and commits only when every check passes. Any failure rolls the
// whole import back. Timestamps come from clock (M1-D2).
func Load(ctx context.Context, db Beginner, clock timex.Clock, opts Options) ([]TableCount, error) {
	x, err := ReadExtracts(opts.Extracts)
	if err != nil {
		return nil, err
	}
	now := clock.Now()
	p, err := BuildPlan(x, now)
	if err != nil {
		return nil, err
	}
	var creds []Credential
	if opts.SeedDev {
		if creds, err = seedDev(ctx, p, opts.Extracts, now, opts.DevPassword); err != nil {
			return nil, err
		}
	}
	var counts []TableCount
	err = pgx.BeginTxFunc(ctx, db, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		q := dbq.New(tx)
		if _, err := tx.Exec(ctx, "select pg_advisory_xact_lock($1)", lockKey); err != nil {
			return fmt.Errorf("migratehistory: lock: %w", err)
		}
		if opts.Replace {
			if err := replace(ctx, tx); err != nil {
				return err
			}
		} else {
			done, err := q.MigrationEntryExists(ctx)
			if err != nil {
				return fmt.Errorf("migratehistory: look for an earlier import: %w", err)
			}
			if done {
				return ErrAlreadyMigrated
			}
		}
		if counts, err = write(ctx, q, p); err != nil {
			return err
		}
		n, err := writeAudit(ctx, q, p, now)
		if err != nil {
			return err
		}
		counts = append(counts, TableCount{"audit_log", n})
		return Verify(ctx, tx, opts.extraChecks...)
	})
	if err != nil {
		return nil, err
	}
	if err := writeCredentials(opts.SeedOut, creds); err != nil {
		return counts, fmt.Errorf("%w (%w)", ErrCredentialsLost, err)
	}
	return counts, nil
}

// ErrCredentialsLost is returned by Load (wrapping the write error) when the
// import committed but the dev logins and passwords could not be written to
// Options.SeedOut. Only their hashes are stored, so the passwords are lost;
// the way back is another run with --replace (dev only, owner DSN).
var ErrCredentialsLost = errors.New("migratehistory: the import was committed, but writing the dev passwords failed, " +
	"so they are lost (only their hashes are stored); run again with --replace --seed-dev " +
	"(APP_ENV=dev, the owner DSN in MIGRATE_DATABASE_URL or --db) to get new ones")

// write copies the plan table by table, parents first.
func write(ctx context.Context, q *dbq.Queries, p *Plan) ([]TableCount, error) {
	var counts []TableCount
	copyTable := func(table string, n int64, err error) error {
		if err != nil {
			return fmt.Errorf("migratehistory: copy %s: %w", table, err)
		}
		counts = append(counts, TableCount{table, n})
		return nil
	}
	steps := []func() error{
		func() error {
			if len(p.Users) == 0 {
				return nil
			}
			n, err := q.CopyUsers(ctx, conv(p.Users, func(r dbq.AppUser) dbq.CopyUsersParams {
				return dbq.CopyUsersParams{ID: r.ID, Name: r.Name, Login: r.Login, PasswordHash: r.PasswordHash,
					Role: r.Role, Active: r.Active, PasswordGeneratedAt: r.PasswordGeneratedAt,
					CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
			}))
			return copyTable("app_user", n, err)
		},
		func() error {
			n, err := q.CopyGuides(ctx, conv(p.Guides, func(r dbq.Guide) dbq.CopyGuidesParams {
				return dbq.CopyGuidesParams{ID: r.ID, Name: r.Name, NameKey: r.NameKey, Type: r.Type, CreatedAt: r.CreatedAt}
			}))
			return copyTable("guide", n, err)
		},
		func() error {
			n, err := q.CopyGuideWeeks(ctx, conv(p.GuideWeeks, func(r dbq.GuideWeek) dbq.CopyGuideWeeksParams { return dbq.CopyGuideWeeksParams(r) }))
			return copyTable("guide_week", n, err)
		},
		func() error {
			n, err := q.CopyFarms(ctx, conv(p.Farms, func(r dbq.Farm) dbq.CopyFarmsParams {
				return dbq.CopyFarmsParams{ID: r.ID, Name: r.Name, NameKey: r.NameKey, Type: r.Type, GuideID: r.GuideID,
					CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
			}))
			return copyTable("farm", n, err)
		},
		func() error {
			n, err := q.CopySections(ctx, conv(p.Sections, func(r dbq.Section) dbq.CopySectionsParams {
				return dbq.CopySectionsParams{ID: r.ID, FarmID: r.FarmID, Code: r.Code, ParentID: r.ParentID,
					SortOrder: r.SortOrder, CreatedAt: r.CreatedAt}
			}))
			return copyTable("section", n, err)
		},
		func() error {
			if len(p.UserFarms) == 0 {
				return nil
			}
			n, err := q.CopyUserFarms(ctx, conv(p.UserFarms, func(r dbq.UserFarm) dbq.CopyUserFarmsParams {
				return dbq.CopyUserFarmsParams(r)
			}))
			return copyTable("user_farm", n, err)
		},
		func() error {
			n, err := q.CopyFeedTypes(ctx, conv(p.FeedTypes, func(r dbq.FeedType) dbq.CopyFeedTypesParams {
				return dbq.CopyFeedTypesParams{ID: r.ID, Name: r.Name, NameKey: r.NameKey, ForSex: r.ForSex,
					Active: r.Active, CreatedAt: r.CreatedAt}
			}))
			return copyTable("feed_type", n, err)
		},
		func() error {
			n, err := q.CopySeries(ctx, conv(p.Series, func(r dbq.Series) dbq.CopySeriesParams {
				return dbq.CopySeriesParams{ID: r.ID, Name: r.Name, NameKey: r.NameKey, Year: r.Year, Status: r.Status,
					PlacementDate: r.PlacementDate, PlacementFarmID: r.PlacementFarmID,
					PlacementInferred: r.PlacementInferred, AgeAtPlacementDays: r.AgeAtPlacementDays,
					ClosedAt: r.ClosedAt, CreatedAt: r.CreatedAt}
			}))
			return copyTable("series", n, err)
		},
		func() error {
			n, err := q.CopyCycles(ctx, conv(p.Cycles, func(r dbq.Cycle) dbq.CopyCyclesParams {
				return dbq.CopyCyclesParams{ID: r.ID, SeriesID: r.SeriesID, FarmID: r.FarmID, GuideID: r.GuideID,
					StartDate: r.StartDate, EndDate: r.EndDate, StartedBy: r.StartedBy,
					StartInferred: r.StartInferred, EndedBy: r.EndedBy, CreatedAt: r.CreatedAt}
			}))
			return copyTable("cycle", n, err)
		},
		func() error {
			n, err := q.CopyEntries(ctx, conv(p.Entries, func(r dbq.Entry) dbq.CopyEntriesParams {
				return dbq.CopyEntriesParams{ID: r.ID, Kind: r.Kind, FarmID: r.FarmID, PeerFarmID: r.PeerFarmID,
					Date: r.Date, SeriesID: r.SeriesID, Buyer: r.Buyer, Note: r.Note, Source: r.Source,
					CreatedBy: r.CreatedBy, CreatedAt: r.CreatedAt, UpdatedBy: r.UpdatedBy, UpdatedAt: r.UpdatedAt,
					Corrected: r.Corrected}
			}))
			return copyTable("entry", n, err)
		},
		func() error {
			n, err := q.CopyBirdEvents(ctx, conv(p.BirdEvents, func(r dbq.BirdEvent) dbq.CopyBirdEventsParams {
				return dbq.CopyBirdEventsParams{ID: r.ID, EntryID: r.EntryID, FarmID: r.FarmID, SectionID: r.SectionID,
					Date: r.Date, Sex: r.Sex, Type: r.Type, Qty: r.Qty, CullReason: r.CullReason,
					PeerSectionID: r.PeerSectionID, Note: r.Note}
			}))
			return copyTable("bird_event", n, err)
		},
		func() error {
			n, err := q.CopyEggProductions(ctx, conv(p.EggProductions, func(r dbq.EggProduction) dbq.CopyEggProductionsParams {
				return dbq.CopyEggProductionsParams{ID: r.ID, EntryID: r.EntryID, FarmID: r.FarmID,
					SectionID: r.SectionID, Date: r.Date, FloorEggs: r.FloorEggs, Cat1: r.Cat1, Cat2: r.Cat2,
					Cat3: r.Cat3, HatchingEggs: r.HatchingEggs, AvgEggWeightG: r.AvgEggWeightG}
			}))
			return copyTable("egg_production", n, err)
		},
		func() error {
			n, err := q.CopyFeedAllocations(ctx, conv(p.FeedAllocations, func(r dbq.FeedAllocation) dbq.CopyFeedAllocationsParams {
				return dbq.CopyFeedAllocationsParams{ID: r.ID, EntryID: r.EntryID, FarmID: r.FarmID,
					SectionID: r.SectionID, Date: r.Date, Sex: r.Sex, FeedTypeID: r.FeedTypeID, GPerBird: r.GPerBird,
					KgRecorded: r.KgRecorded}
			}))
			return copyTable("feed_allocation", n, err)
		},
		func() error {
			n, err := q.CopyFeedDeliveries(ctx, conv(p.FeedDeliveries, func(r dbq.FeedDelivery) dbq.CopyFeedDeliveriesParams {
				return dbq.CopyFeedDeliveriesParams{ID: r.ID, EntryID: r.EntryID, FarmID: r.FarmID,
					FeedTypeID: r.FeedTypeID, Date: r.Date, QtyKg: r.QtyKg, Note: r.Note, PricePerKg: r.PricePerKg,
					TotalLei: r.TotalLei, PriceInput: r.PriceInput, PricedBy: r.PricedBy, PricedAt: r.PricedAt,
					PriceFromRegister: r.PriceFromRegister}
			}))
			return copyTable("feed_delivery", n, err)
		},
		func() error {
			n, err := q.CopyFeedStockAdjustments(ctx, conv(p.FeedStockAdjustments, func(r dbq.FeedStockAdjustment) dbq.CopyFeedStockAdjustmentsParams {
				return dbq.CopyFeedStockAdjustmentsParams(r)
			}))
			return copyTable("feed_stock_adjustment", n, err)
		},
		func() error {
			n, err := q.CopyWeighings(ctx, conv(p.Weighings, func(r dbq.Weighing) dbq.CopyWeighingsParams {
				return dbq.CopyWeighingsParams{ID: r.ID, EntryID: r.EntryID, FarmID: r.FarmID, SectionID: r.SectionID,
					Date: r.Date, Sex: r.Sex, AvgWeightG: r.AvgWeightG, CvPct: r.CvPct,
					UniformityPct: r.UniformityPct, SampleSize: r.SampleSize}
			}))
			return copyTable("weighing", n, err)
		},
		func() error {
			n, err := q.CopyLegacyMortality(ctx, conv(p.LegacyMortality, func(r dbq.LegacyMortality) dbq.CopyLegacyMortalityParams { return dbq.CopyLegacyMortalityParams(r) }))
			return copyTable("legacy_mortality", n, err)
		},
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return nil, err
		}
	}
	return counts, nil
}

func conv[T, U any](in []T, f func(T) U) []U {
	out := make([]U, len(in))
	for i, r := range in {
		out[i] = f(r)
	}
	return out
}

// farmCounts are the new_values of a farm's migration audit row (M1-D3): the
// rows the import wrote for that farm, per table, whichever entry owns them
// (user answer Q3), and the entries that touch the farm, per kind.
type farmCounts struct {
	Farm    string                `json:"farm"`
	Rows    map[string]int        `json:"rows"`
	Entries map[dbq.EntryKind]int `json:"entries"`
}

// writeAudit writes one audit_log row per farm: create of the farm's
// migration entry, no author, no summary (M1-D3); then the seed's rows.
func writeAudit(ctx context.Context, q *dbq.Queries, p *Plan, now time.Time) (int64, error) {
	var n int64
	for _, f := range p.Farms {
		c := farmCounts{Farm: f.Name, Rows: map[string]int{}, Entries: map[dbq.EntryKind]int{}}
		count := func(table string, farm uuid.UUID) {
			if farm == f.ID {
				c.Rows[table]++
			}
		}
		for _, r := range p.Sections {
			count("section", r.FarmID)
		}
		for _, r := range p.Cycles {
			count("cycle", r.FarmID)
		}
		for _, r := range p.Entries {
			if r.FarmID == f.ID || (r.PeerFarmID != nil && *r.PeerFarmID == f.ID) {
				c.Rows["entry"]++
				c.Entries[r.Kind]++
			}
		}
		for _, r := range p.BirdEvents {
			count("bird_event", r.FarmID)
		}
		for _, r := range p.EggProductions {
			count("egg_production", r.FarmID)
		}
		for _, r := range p.FeedAllocations {
			count("feed_allocation", r.FarmID)
		}
		for _, r := range p.FeedDeliveries {
			count("feed_delivery", r.FarmID)
		}
		for _, r := range p.FeedStockAdjustments {
			count("feed_stock_adjustment", r.FarmID)
		}
		for _, r := range p.Weighings {
			count("weighing", r.FarmID)
		}
		for _, r := range p.LegacyMortality {
			count("legacy_mortality", r.FarmID)
		}
		values, err := json.Marshal(c)
		if err != nil {
			return 0, err
		}
		key := p.farmKey(f.ID)
		entry := p.migrationEntry[key]
		var day *timex.Day
		for _, e := range p.Entries {
			if e.ID == entry {
				d := e.Date
				day = &d
			}
		}
		kind := dbq.EntryKindMigration
		err = q.InsertEntryAudit(ctx, dbq.InsertEntryAuditParams{
			At: now, Action: dbq.AuditActionCreate, EntryID: entry, EntryKind: &kind,
			RecordTypes: []string{}, FarmIds: []uuid.UUID{f.ID}, RecordDate: day,
			NewValues: values, Source: dbq.RecordSourceMigration,
		})
		if err != nil {
			return 0, fmt.Errorf("migratehistory: audit %s: %w", f.Name, err)
		}
		n++
	}
	for _, a := range p.Audit {
		if err := q.InsertEntryAudit(ctx, a); err != nil {
			return 0, fmt.Errorf("migratehistory: audit: %w", err)
		}
		n++
	}
	return n, nil
}
