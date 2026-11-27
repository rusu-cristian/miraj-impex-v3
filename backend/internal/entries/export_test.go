package entries

import (
	"context"
	"slices"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// SaveRowKeys is every section × sex a Hrană zilnică save of farmID on day
// accepts (fed over EntriesFeedDaySlots), as "code sex", sorted.
func SaveRowKeys(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) ([]string, error) {
	slots, err := loadFeedSlots(ctx, dbq.New(db), farmID, day)
	if err != nil {
		return nil, err
	}
	out := []string{}
	for _, s := range slots {
		if fed(s) {
			out = append(out, s.Code+" "+string(s.Sex))
		}
	}
	slices.Sort(out)
	return out, nil
}

// FormRowKeys are the rows of a form state as "code sex", sorted.
func FormRowKeys(res FeedFormResponse) []string {
	out := []string{}
	for _, r := range res.Rows {
		out = append(out, r.Code+" "+string(r.Sex))
	}
	slices.Sort(out)
	return out
}
