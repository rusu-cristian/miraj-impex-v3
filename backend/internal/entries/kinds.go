package entries

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Kind adapters (plan M5 T3, Q1, ARC-1): the entry kinds whose saves M5
// reads, edits, corrects, deletes, undoes and restores as a whole save
// (D172). Each served kind is one entry of the kinds table, its adapter
// (kind_daily.go, kind_feed_daily.go, kind_delivery.go): how its live child
// rows load as its create shape (the `values` of GET /api/entries/{id} and
// of PUT, Q6) and as its audit snapshot (D7), its PUT, the undo of its
// update, its delete, undo and restore texts and the holder of a key a
// restored row meets. Transfer, sale (M7), weighing (M9), placement (M10)
// and the per-farm migration entry are not in the table yet: their routes
// answer 404, and nothing falls back to another kind. Adding a kind = one
// table entry + its adapter file.

// ErrKindNotServed: the entry's kind has no adapter yet (Q1); its routes
// answer 404.
var ErrKindNotServed = errors.New("entries: entry kind not served")

// kindAdapter is a served kind's part of the entry routes (ARC-1). The
// runner and the plans call it only for an entry of its kind.
type kindAdapter interface {
	// load reads the live child rows of entry s.ID into s: Rows, and the
	// Sections and FeedTypes they name (FeedTypes left nil by a kind whose
	// rows name no feed type: the entry read leaves `feedTypes` out).
	load(ctx context.Context, q *dbq.Queries, s *StoredEntry) error
	// values is s as its create shape (Q6): DailyValues, FeedDailyValues, …
	values(s StoredEntry) any
	// update decodes, checks (level 1) and runs a PUT of entry t (§7.8).
	update(h *Handler, w http.ResponseWriter, r *http.Request, p auth.Principal, t updateTarget) (Result, error)
	// undoUpdate is the plan of the undo of an update: the entry put back as
	// before is the update's old_values (Q8). Saved.Restored carries the
	// rows it un-deletes, with the error when that write fails (ARC-5).
	undoUpdate(ctx context.Context, tx *ChangeTx, before updateSnapshot) (Saved, error)
	// createUndoStale reports whether the create of an entry whose live rows
	// are rows can't be undone any more although it is still the entry's
	// latest audited action: a change no audit row of the entry records
	// (COR-2: a delivery priced in Prețuri since) → 409 STALE.
	createUndoStale(rows Snapshot) bool
	// section is the rows of a per-section delete of section id (D98), or
	// 400 VALIDATION on `sectionId` when it names none or the kind has none.
	section(stored StoredEntry, id uuid.UUID) (Snapshot, error)
	// removalEffect is a delete's Dialog sentence (dry-run effects[0],
	// §6.15, S15, Q9).
	removalEffect(c rowsText) string
	// removalToast is the Toast of a delete or of an undo of a create (Q15).
	removalToast(c rowsText) string
	// restoreToast is the Toast of a restore (Q15).
	restoreToast(c rowsText) string
	// summary is the audit summary of the rows a delete, an undo of a create
	// or a restore takes (the Jurnal name, Q5, Q15).
	summary(c rowsText) string
	// holderMessage is the 409 ALREADY_ENTERED sentence of a restored row
	// whose key the live entry h holds now, created by a user (restoreHolder;
	// an imported holder gets EntryAlreadySaved).
	holderMessage(h dbq.EntryRestoreHolderRow, day timex.Day) string
}

// rowsText is what a kind's delete, undo and restore texts name
// (kindAdapter).
type rowsText struct {
	Entry     LockedEntry
	Rows      Snapshot   // the rows taken away or put back
	SectionID *uuid.UUID // a per-section delete, or the restore of one
	Undo      bool       // an undo of a create (removalToast)
	Figures   rowFigures
	Stock     stockAfter
	Info      dbq.EntryDeleteContextRow // the farm's and the author's names, the latest correction
}

// kinds is every served kind's adapter (Q1, ARC-1).
var kinds = map[dbq.EntryKind]kindAdapter{
	dbq.EntryKindDaily:     dailyKind{},
	dbq.EntryKindFeedDaily: feedDailyKind{},
	dbq.EntryKindDelivery:  deliveryKind{},
}

// adapterOf is kind's adapter; ErrKindNotServed for a kind without one.
func adapterOf(kind dbq.EntryKind) (kindAdapter, error) {
	a, ok := kinds[kind]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrKindNotServed, kind)
	}
	return a, nil
}

// Served reports whether the change and read routes serve entries of kind
// (Q1): whether the kinds table has its adapter.
func Served(kind dbq.EntryKind) bool {
	_, ok := kinds[kind]
	return ok
}

// Create shapes without WriteMeta (Q6 `values`): what POST sends and PUT
// sends back. GET /api/entries/{id} returns them filled from the live rows,
// so a client can send them back unchanged.

// DailyValues is a Zilnic save (§7.8 DailyCreateRequest without WriteMeta):
// the bird values per section × sex and the eggs per egg unit.
type DailyValues struct {
	FarmID   uuid.UUID      `json:"farmId"`
	Date     timex.Day      `json:"date"`
	Sections []DailySection `json:"sections"`
	Eggs     []DailyEggs    `json:"eggs"`
}

// DailySection is one section of a Zilnic save.
type DailySection struct {
	SectionID uuid.UUID `json:"sectionId"`
	NoLosses  bool      `json:"noLosses"` // "Fără pierderi azi": 0 rows for mortality and cull of the sexes present
	F         *DailySex `json:"F"`
	M         *DailySex `json:"M"`
}

// DailySex is one sex of a Zilnic section; nil = field left empty → no row
// (empty is null, never 0, §4.7).
type DailySex struct {
	Mortality   *num.Num `json:"mortality"`
	Cull        *num.Num `json:"cull"`
	CullReason  *string  `json:"cullReason"`  // boala | schiopatare | altul
	SexingError *num.Num `json:"sexingError"` // signed
}

// DailyEggs is one egg unit of a Zilnic save: the hall, or a section without
// parent and children.
type DailyEggs struct {
	SectionID  uuid.UUID `json:"sectionId"`
	Floor      *num.Num  `json:"floor"`
	Cat1       *num.Num  `json:"cat1"`
	Cat2       *num.Num  `json:"cat2"`
	Cat3       *num.Num  `json:"cat3"`
	Hatching   *num.Num  `json:"hatching"`
	AvgWeightG *num.Num  `json:"avgWeightG"`
}

// FeedDailyValues is a Hrană zilnică save (FeedDailyCreateRequest without
// WriteMeta).
type FeedDailyValues struct {
	FarmID uuid.UUID      `json:"farmId"`
	Date   timex.Day      `json:"date"`
	Rows   []FeedDailyRow `json:"rows"`
}

// DeliveryValues is a Livrare hrană save (DeliveryCreateRequest without
// WriteMeta and without the price, which only Prețuri sets, Q6).
type DeliveryValues struct {
	FarmID     uuid.UUID `json:"farmId"`
	Date       timex.Day `json:"date"`
	FeedTypeID uuid.UUID `json:"feedTypeId"`
	QtyKg      *num.Num  `json:"qtyKg"`
	Note       *string   `json:"note"`
}

// Snapshot is an entry's live child rows as audit_log.old_values /
// new_values store them (§4.10, plan M5-D7): every row with its id, under
// the name of its table, with its section code and feed type name. A
// restore un-deletes exactly the rows a delete's snapshot names; an undone
// update writes back the rows its old_values hold. Tables without rows are
// left out.
type Snapshot struct {
	BirdEvents      []SnapshotBird     `json:"birdEvents,omitempty"`
	EggProduction   []SnapshotEggs     `json:"eggProduction,omitempty"`
	FeedAllocations []SnapshotFeed     `json:"feedAllocations,omitempty"`
	FeedDeliveries  []SnapshotDelivery `json:"feedDeliveries,omitempty"`
}

// SnapshotBird is one bird_event row. Qty is the stored signed delta
// (mortality and cull ≤ 0, a sexing error as recorded).
type SnapshotBird struct {
	ID         uuid.UUID         `json:"id"`
	SectionID  uuid.UUID         `json:"sectionId"`
	Section    string            `json:"section"`
	Sex        dbq.Sex           `json:"sex"`
	Type       dbq.BirdEventType `json:"type"`
	Qty        int32             `json:"qty"`
	CullReason *dbq.CullReason   `json:"cullReason,omitzero"`
}

// SnapshotEggs is one egg_production row (the egg unit's values; total_eggs
// is generated and never stored here).
type SnapshotEggs struct {
	ID         uuid.UUID `json:"id"`
	SectionID  uuid.UUID `json:"sectionId"`
	Section    string    `json:"section"`
	Floor      *int32    `json:"floor"`
	Cat1       *int32    `json:"cat1"`
	Cat2       *int32    `json:"cat2"`
	Cat3       *int32    `json:"cat3"`
	Hatching   *int32    `json:"hatching"`
	AvgWeightG *num.Num  `json:"avgWeightG"`
}

// SnapshotFeed is one feed_allocation row. G is null on an imported kg-only
// row; KgRecorded is set only on imported rows (D163).
type SnapshotFeed struct {
	ID         uuid.UUID `json:"id"`
	SectionID  uuid.UUID `json:"sectionId"`
	Section    string    `json:"section"`
	Sex        dbq.Sex   `json:"sex"`
	FeedTypeID uuid.UUID `json:"feedTypeId"`
	Type       string    `json:"type"`
	G          *num.Num  `json:"g"`
	KgRecorded *num.Num  `json:"kgRecorded,omitzero"`
}

// SnapshotDelivery is one feed_delivery row, its price included (a deleted
// priced delivery keeps its lei in the audit, Q9).
type SnapshotDelivery struct {
	ID                uuid.UUID        `json:"id"`
	FeedTypeID        uuid.UUID        `json:"feedTypeId"`
	Type              string           `json:"type"`
	Kg                num.Num          `json:"kg"`
	Note              *string          `json:"note"`
	PricePerKg        *num.Num         `json:"pricePerKg,omitzero"`
	TotalLei          *num.Num         `json:"totalLei,omitzero"`
	PriceInput        *string          `json:"priceInput,omitzero"`
	PricedBy          *uuid.UUID       `json:"pricedBy,omitzero"`
	PricedAt          *timex.LocalTime `json:"pricedAt,omitzero"`
	PriceFromRegister bool             `json:"priceFromRegister,omitzero"`
}

// StoredEntry is a served entry's live child rows, loaded by LoadStored.
type StoredEntry struct {
	ID     uuid.UUID
	Kind   dbq.EntryKind
	FarmID uuid.UUID
	Date   timex.Day
	Rows   Snapshot
	// Sections are the sections the rows name (bird sections in tree order,
	// then the egg units), for the codes in messages and the entry read.
	Sections []EntrySection
	// FeedTypes are the feed types the rows name (feed_daily, delivery), in
	// the rows' order; nil for a daily entry.
	FeedTypes []EntryFeedType
}

// LoadStored reads the live child rows of entry id of a served kind on db
// (a change's transaction, or the entry read's snapshot) with the kind's
// adapter: one statement for feed_daily and delivery, two for daily (birds,
// eggs). Another kind → ErrKindNotServed.
func LoadStored(ctx context.Context, db dbq.DBTX, id uuid.UUID, kind dbq.EntryKind, farmID uuid.UUID, date timex.Day) (StoredEntry, error) {
	a, err := adapterOf(kind)
	if err != nil {
		return StoredEntry{}, err
	}
	s := StoredEntry{ID: id, Kind: kind, FarmID: farmID, Date: date, Sections: []EntrySection{}}
	if err := a.load(ctx, dbq.New(db), &s); err != nil {
		return StoredEntry{}, err
	}
	return s, nil
}

func (s *StoredEntry) addSection(sec EntrySection) {
	if !slices.ContainsFunc(s.Sections, func(e EntrySection) bool { return e.ID == sec.ID }) {
		s.Sections = append(s.Sections, sec)
	}
}

func (s *StoredEntry) addFeedType(ft EntryFeedType) {
	if !slices.ContainsFunc(s.FeedTypes, func(e EntryFeedType) bool { return e.ID == ft.ID }) {
		s.FeedTypes = append(s.FeedTypes, ft)
	}
}

// Code is the code of a section the rows name, "" for another one.
func (s StoredEntry) Code(sectionID uuid.UUID) string {
	for _, sec := range s.Sections {
		if sec.ID == sectionID {
			return sec.Code
		}
	}
	return ""
}

// Live reports whether the entry still has a live child row.
func (s StoredEntry) Live() bool {
	r := s.Rows
	return len(r.BirdEvents)+len(r.EggProduction)+len(r.FeedAllocations)+len(r.FeedDeliveries) > 0
}

// Values is the entry's create shape (Q6), its adapter's: DailyValues,
// FeedDailyValues or DeliveryValues; nil for a kind not served.
func (s StoredEntry) Values() any {
	a, err := adapterOf(s.Kind)
	if err != nil {
		return nil
	}
	return a.values(s)
}

func intNum(i int64) *num.Num {
	n := num.FromInt(i)
	return &n
}

func int32Num(i *int32) *num.Num {
	if i == nil {
		return nil
	}
	return intNum(int64(*i))
}

// localTimePtr is t as a LocalTime, nil for nil.
func localTimePtr(t *time.Time) *timex.LocalTime {
	if t == nil {
		return nil
	}
	at := timex.At(*t)
	return &at
}
