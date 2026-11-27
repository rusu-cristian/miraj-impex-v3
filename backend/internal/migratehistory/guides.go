package migratehistory

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/textkey"
)

// guides plans the two guides and their weeks (§10.3): growth SAPT 1–25
// (feed_g, body_weight_g) and laying SAPT 19–65 (feed_g ♀/♂, lay_pct and
// mortality_cum_pct ♀). The other guide-laying fields are not imported (D184).
func (b *builder) guides() error {
	for _, g := range b.x.farms.Guides {
		row := dbq.Guide{ID: uuid.New(), Name: g.Name, NameKey: textkey.Key(g.Name), Type: g.Type, CreatedAt: b.now}
		b.p.guideIDs[g.ID] = row.ID
		b.p.Guides = append(b.p.Guides, row)
	}
	guide := func(key string) (uuid.UUID, error) {
		id, ok := b.p.guideIDs[key]
		if !ok {
			return uuid.Nil, fmt.Errorf("migratehistory: unknown guide %q", key)
		}
		return id, nil
	}
	for _, r := range b.x.guideGrowth.Rows {
		id, err := guide(r.GuideID)
		if err != nil {
			return err
		}
		b.p.GuideWeeks = append(b.p.GuideWeeks, dbq.GuideWeek{
			GuideID: id, Week: r.Week, Sex: r.Sex, FeedG: r.FeedG, BodyWeightG: r.BodyWeightG,
		})
	}
	for _, r := range b.x.guideLaying.Rows {
		id, err := guide(r.GuideID)
		if err != nil {
			return err
		}
		b.p.GuideWeeks = append(b.p.GuideWeeks, dbq.GuideWeek{
			GuideID: id, Week: r.AgeWeek, Sex: r.Sex, FeedG: r.FeedG,
			LayPct: r.LayPct, MortalityCumPct: r.MortalityCumPct,
		})
	}
	return nil
}
