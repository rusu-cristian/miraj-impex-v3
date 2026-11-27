package entries

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Service is the entries module's service: the form states and the creates
// (§7.8). Every create runs on the save runner.
type Service struct {
	store  *db.Store
	clock  timex.Clock
	runner *Runner
}

// NewService builds the entries service on the app pool's store.
func NewService(store *db.Store, clock timex.Clock) *Service {
	return &Service{store: store, clock: clock, runner: NewRunner(store, clock)}
}

// checkWriter is the access check of an entry create (§3.3, M3 D5): vet,
// manager and owner may write, anyone else gets 403 FORBIDDEN "Nu ai acces la
// această pagină."; a farm outside the caller's farms is 403 FARM_FORBIDDEN.
// A nil farmID is a level-1 issue of the body, not an access question, so
// only the role is checked then.
func checkWriter(p auth.Principal, farmID uuid.UUID) error {
	if !p.Role.In(auth.Vet, auth.Manager, auth.Owner) {
		return httpx.NewError(http.StatusForbidden, httpx.CodeForbidden, messages.PageForbidden())
	}
	if farmID != uuid.Nil && !p.CanSeeFarm(farmID) {
		return httpx.FarmForbidden()
	}
	return nil
}

// FeedRefs reads what FeedDailyCreateRequest.Validate needs: the codes of
// the farm's sections and which of the request's feed types exist. An
// unknown farm is 404 NOT_FOUND (M2 D4: only an owner gets this far with
// one; anyone else got FARM_FORBIDDEN from checkWriter). A request without
// farmId reads nothing.
func (s *Service) FeedRefs(ctx context.Context, req FeedDailyCreateRequest) (FeedRefs, error) {
	refs := FeedRefs{Sections: map[uuid.UUID]string{}, FeedTypes: map[uuid.UUID]bool{}}
	if req.FarmID == uuid.Nil {
		return refs, nil
	}
	typeIDs := make([]uuid.UUID, 0, len(req.Rows))
	for _, r := range req.Rows {
		typeIDs = append(typeIDs, r.FeedTypeID)
	}
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) error {
		q := dbq.New(tx)
		sections, err := q.EntriesFarmSections(ctx, req.FarmID)
		if err != nil {
			return err
		}
		if len(sections) == 0 {
			return httpx.NotFound()
		}
		for _, sec := range sections {
			if sec.SectionID != nil && sec.Code != nil {
				refs.Sections[*sec.SectionID] = *sec.Code
			}
		}
		types, err := q.EntriesKnownFeedTypes(ctx, typeIDs)
		if err != nil {
			return err
		}
		for _, id := range types {
			refs.FeedTypes[id] = true
		}
		return nil
	})
	return refs, err
}
