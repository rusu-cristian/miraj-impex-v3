package entries

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// deliveryKind is the Livrare hrană kind's adapter (kindAdapter, ARC-1):
// one feed_delivery row, its price included. Its PUT and the undo of its
// update are delivery_update.go.
type deliveryKind struct{}

// load reads the live delivery row of the entry with its price.
func (deliveryKind) load(ctx context.Context, q *dbq.Queries, s *StoredEntry) error {
	rows, err := q.EntryDeliveryRows(ctx, s.ID)
	if err != nil {
		return fmt.Errorf("entries: delivery rows of %s: %w", s.ID, err)
	}
	s.FeedTypes = []EntryFeedType{}
	for _, r := range rows {
		s.Rows.FeedDeliveries = append(s.Rows.FeedDeliveries, SnapshotDelivery{
			ID: r.ID, FeedTypeID: r.FeedTypeID, Type: r.FeedTypeName, Kg: num.Of(r.QtyKg), Note: r.Note,
			PricePerKg: nullNum(r.PricePerKg), TotalLei: nullNum(r.TotalLei), PriceInput: r.PriceInput,
			PricedBy: r.PricedBy, PricedAt: localTimePtr(r.PricedAt), PriceFromRegister: r.PriceFromRegister,
		})
		s.addFeedType(EntryFeedType{ID: r.FeedTypeID, Name: r.FeedTypeName, ForSex: r.FeedTypeForSex, Active: r.FeedTypeActive})
	}
	return nil
}

func (deliveryKind) values(s StoredEntry) any { return s.Delivery() }

func (deliveryKind) update(h *Handler, w http.ResponseWriter, r *http.Request, p auth.Principal, t updateTarget) (Result, error) {
	return h.updateDelivery(w, r, p, t)
}

func (deliveryKind) undoUpdate(ctx context.Context, tx *ChangeTx, before updateSnapshot) (Saved, error) {
	return planUndoDeliveryUpdate(ctx, tx, before)
}

// createUndoStale (COR-2): a create never carries a price (§6.7: only
// Prețuri sets it, and its audit rows don't name the entry), so a priced
// delivery was priced after its create, which an undo of the create would
// silently throw away.
func (deliveryKind) createUndoStale(rows Snapshot) bool {
	for _, d := range rows.FeedDeliveries {
		if d.TotalLei != nil || d.PricePerKg != nil {
			return true
		}
	}
	return false
}

// section: a delivery is deleted whole.
func (deliveryKind) section(StoredEntry, uuid.UUID) (Snapshot, error) {
	return Snapshot{}, badSection()
}

func (deliveryKind) removalEffect(c rowsText) string {
	d := c.Rows.FeedDeliveries[0]
	return messages.DeliveryDeleteEffect(authorOf(c.Info), c.Entry.CreatedAt, correctedBy(c.Entry, c.Info), d.Kg.Decimal(), d.Type)
}

func (deliveryKind) removalToast(c rowsText) string {
	d := c.Rows.FeedDeliveries[0]
	if c.Undo {
		return messages.DeliveryUndone(d.Kg.Decimal(), d.Type)
	}
	return messages.DeliveryDeleted(d.Kg.Decimal(), d.Type, c.Info.FarmName, c.Stock.kg)
}

func (deliveryKind) restoreToast(c rowsText) string {
	d := c.Rows.FeedDeliveries[0]
	return messages.DeliveryRestored(d.Kg.Decimal(), d.Type, c.Info.FarmName, c.Stock.kg)
}

// summary: "Livrare · 9 000 kg Breeder 2".
func (deliveryKind) summary(c rowsText) string {
	d := c.Rows.FeedDeliveries[0]
	return messages.DeliverySummary(d.Kg.Decimal(), d.Type)
}

// holderMessage: a delivery row has no unique key (EntryRestoreHolder never
// names one); the generic sentence.
func (deliveryKind) holderMessage(dbq.EntryRestoreHolderRow, timex.Day) string {
	return messages.EntryAlreadySaved()
}

// Delivery is a delivery entry's row as DeliveryValues (zero kg fields when
// its row is gone).
func (s StoredEntry) Delivery() DeliveryValues {
	v := DeliveryValues{FarmID: s.FarmID, Date: s.Date}
	if len(s.Rows.FeedDeliveries) > 0 {
		d := s.Rows.FeedDeliveries[0]
		kg := d.Kg
		v.FeedTypeID, v.QtyKg, v.Note = d.FeedTypeID, &kg, d.Note
	}
	return v
}
