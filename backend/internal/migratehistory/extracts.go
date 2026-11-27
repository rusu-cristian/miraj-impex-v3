package migratehistory

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Extracts is the content of design/data/extracts that the history migration
// reads (SPEC §10.3), decoded but not yet interpreted. Only the fields the
// plan uses are kept; numbers that are not counts stay exact decimals.
type Extracts struct {
	farms       farmsFile
	placement   placementFile
	cycles      cyclesFile
	feedTypes   feedTypesFile
	transfers   transfersFile
	telitaDaily telitaMortalityFile
	telitaFeed  telitaFeedFile
	deliveries  deliveriesFile
	weighings   weighingsFile
	guideGrowth guideGrowthFile
	guideLaying guideLayingFile
	hirbovat    hirbovatDaily
	prevEggs    prevEggsFile
}

// ReadExtracts reads and decodes the extract files in dir. It touches no
// database, so a wrong directory fails before anything is written.
func ReadExtracts(dir string) (*Extracts, error) {
	if fi, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("migratehistory: extracts directory: %w", err)
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("migratehistory: extracts directory: %s is not a directory", dir)
	}
	x := &Extracts{}
	var hd hirbovatDailyFile
	for _, f := range []struct {
		name string
		dst  any
	}{
		{"farms-sections.json", &x.farms},
		{"placement-seria2026.json", &x.placement},
		{"cycles.json", &x.cycles},
		{"feed-types.json", &x.feedTypes},
		{"transfers-seria2026.json", &x.transfers},
		{"mortality-telita-daily.json", &x.telitaDaily},
		{"feed-daily-telita.json", &x.telitaFeed},
		{"feed-deliveries.json", &x.deliveries},
		{"weighings-weekly.json", &x.weighings},
		{"guide-growth.json", &x.guideGrowth},
		{"guide-laying.json", &x.guideLaying},
		{"hirbovat-daily.json", &hd},
		{"eggs-hirbovat-prev-daily.json", &x.prevEggs},
	} {
		if err := readJSON(filepath.Join(dir, f.name), f.dst); err != nil {
			return nil, err
		}
	}
	var err error
	if x.hirbovat.sectionDay, err = decodeTable[hirbovatSectionDay](hd.SectionDay); err != nil {
		return nil, fmt.Errorf("migratehistory: hirbovat-daily.json section_day: %w", err)
	}
	if x.hirbovat.sales, err = decodeTable[hirbovatSale](hd.Sales); err != nil {
		return nil, fmt.Errorf("migratehistory: hirbovat-daily.json sales: %w", err)
	}
	if x.hirbovat.hallDay, err = decodeTable[hirbovatHallDay](hd.HallDay); err != nil {
		return nil, fmt.Errorf("migratehistory: hirbovat-daily.json hall_day: %w", err)
	}
	return x, nil
}

func readJSON(path string, dst any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("migratehistory: %w", err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return fmt.Errorf("migratehistory: %s: %w", filepath.Base(path), err)
	}
	return nil
}

// table is a block of hirbovat-daily.json: column names plus rows as arrays.
type table struct {
	Cols []string            `json:"cols"`
	Rows [][]json.RawMessage `json:"rows"`
}

// decodeTable turns each array row into an object keyed by the column names
// and decodes it into T (JSON tags = column names).
func decodeTable[T any](t table) ([]T, error) {
	if len(t.Cols) == 0 {
		return nil, errors.New("no cols")
	}
	out := make([]T, 0, len(t.Rows))
	for i, row := range t.Rows {
		if len(row) != len(t.Cols) {
			return nil, fmt.Errorf("row %d has %d cells, want %d", i, len(row), len(t.Cols))
		}
		var obj bytes.Buffer
		obj.WriteByte('{')
		for j, col := range t.Cols {
			if j > 0 {
				obj.WriteByte(',')
			}
			name, _ := json.Marshal(col)
			obj.Write(name)
			obj.WriteByte(':')
			obj.Write(row[j])
		}
		obj.WriteByte('}')
		var v T
		if err := json.Unmarshal(obj.Bytes(), &v); err != nil {
			return nil, fmt.Errorf("row %d: %w", i, err)
		}
		out = append(out, v)
	}
	return out, nil
}

// farms-sections.json
type farmsFile struct {
	Meta struct {
		CodeMapOldToNew map[string]string `json:"code_map_old_to_new"`
	} `json:"meta"`
	Guides []struct {
		ID   string       `json:"id"`
		Name string       `json:"name"`
		Type dbq.FarmType `json:"type"`
	} `json:"guides"`
	Farms []struct {
		ID      string       `json:"id"`
		Name    string       `json:"name"`
		Type    dbq.FarmType `json:"type"`
		GuideID *string      `json:"guide_id"`
	} `json:"farms"`
	Sections []struct {
		ID       string  `json:"id"`
		FarmID   string  `json:"farm_id"`
		Code     string  `json:"code"`
		ParentID *string `json:"parent_id"`
	} `json:"sections"`
}

// placement-seria2026.json
type placementFile struct {
	Meta struct {
		Series string `json:"series"`
	} `json:"meta"`
	Rows []struct {
		Date    timex.Day `json:"date"`
		FarmID  string    `json:"farm_id"`
		Section string    `json:"section"`
		Sex     dbq.Sex   `json:"sex"`
		Qty     int32     `json:"qty"`
	} `json:"rows"`
}

// cycles.json
type cyclesFile struct {
	Series []struct {
		ID                 string           `json:"id"`
		Name               string           `json:"name"`
		Year               int16            `json:"year"`
		Status             dbq.SeriesStatus `json:"status"`
		PlacementDate      timex.Day        `json:"placement_date"`
		PlacementInferred  bool             `json:"placement_inferred"`
		AgeAtPlacementDays int16            `json:"age_at_placement_days"`
	} `json:"series"`
	Cycles []struct {
		ID            string     `json:"id"`
		SeriesID      string     `json:"series_id"`
		FarmID        string     `json:"farm_id"`
		GuideID       *string    `json:"guide_id"`
		StartDate     timex.Day  `json:"start_date"`
		EndDate       *timex.Day `json:"end_date"`
		StartInferred bool       `json:"start_inferred"`
	} `json:"cycles"`
}

// feed-types.json
type feedTypesFile struct {
	Rows []struct {
		ID     string   `json:"id"`
		Name   string   `json:"name"`
		ForSex *dbq.Sex `json:"for_sex"`
		FarmID string   `json:"farm_id"`
		Active bool     `json:"active"`
	} `json:"rows"`
}

// transfers-seria2026.json
type transfersFile struct {
	TransfersOut []struct {
		Date        timex.Day `json:"date"`
		FromFarm    string    `json:"from_farm"`
		FromSection string    `json:"from_section"`
		Sex         dbq.Sex   `json:"sex"`
		Qty         int32     `json:"qty"`
	} `json:"transfers_out"`
	Arrivals []struct {
		Date    timex.Day `json:"date"`
		FarmID  string    `json:"farm_id"`
		Section string    `json:"section"`
		Sex     dbq.Sex   `json:"sex"`
		Qty     int32     `json:"qty"`
	} `json:"arrivals"`
	Links []struct {
		OutDate     timex.Day        `json:"out_date"`
		From        []string         `json:"from"`
		Sex         dbq.Sex          `json:"sex"`
		QtyOut      int32            `json:"qty_out"`
		ArrivalDate timex.Day        `json:"arrival_date"`
		To          map[string]int32 `json:"to"`
		QtyIn       int32            `json:"qty_in"`
	} `json:"links_inferred_by_date"`
	MovesHirbovat []struct {
		Date timex.Day `json:"date"`
		From string    `json:"from"`
		To   string    `json:"to"`
		Sex  dbq.Sex   `json:"sex"`
		Qty  int32     `json:"qty"`
	} `json:"internal_moves_hirbovat"`
	MovesTelita []struct {
		Date    timex.Day `json:"date"`
		From    string    `json:"from"`
		To      string    `json:"to"`
		SexFrom dbq.Sex   `json:"sex_from"`
		SexTo   dbq.Sex   `json:"sex_to"`
		Qty     int32     `json:"qty"`
	} `json:"internal_moves_telita"`
	OtherCorrectionsTelita []struct {
		Date    timex.Day `json:"date"`
		Section string    `json:"section"`
		Sex     dbq.Sex   `json:"sex"`
		Qty     int32     `json:"qty"`
	} `json:"other_corrections_telita"`
}

// mortality-telita-daily.json
type telitaMortalityFile struct {
	Rows []struct {
		Date        timex.Day `json:"date"`
		Section     string    `json:"section"`
		Sex         dbq.Sex   `json:"sex"`
		Mortality   int32     `json:"mortality"`
		SexingError int32     `json:"sexing_error"`
		Correction  int32     `json:"correction"`
	} `json:"rows"`
}

// feed-daily-telita.json
type telitaFeedFile struct {
	Days []struct {
		Date     timex.Day `json:"date"`
		Sections map[string]struct {
			Sex      dbq.Sex             `json:"sex"`
			GPerBird decimal.NullDecimal `json:"g_per_bird"`
			Kg       decimal.NullDecimal `json:"kg"`
		} `json:"sections"`
	} `json:"days"`
}

// feed-deliveries.json
type deliveriesFile struct {
	Meta struct {
		HirbovatAdjustments []struct {
			Date   timex.Day       `json:"date"`
			ForSex dbq.Sex         `json:"for_sex"`
			QtyKg  decimal.Decimal `json:"qty_kg"`
		} `json:"hirbovat_adjustments"`
	} `json:"meta"`
	Rows []struct {
		Date         timex.Day           `json:"date"`
		FarmID       string              `json:"farm_id"`
		QtyKg        decimal.Decimal     `json:"qty_kg"`
		WorkbookNote *string             `json:"workbook_note"`
		FeedTypeID   string              `json:"feed_type_id"`
		Priced       bool                `json:"priced"`
		PricePerKg   decimal.NullDecimal `json:"price_per_kg"`
		TotalLei     decimal.NullDecimal `json:"total_lei"`
	} `json:"rows"`
}

// weighings-weekly.json
type weighingsFile struct {
	Rows []struct {
		Section       string              `json:"section"`
		Sex           dbq.Sex             `json:"sex"`
		Date          timex.Day           `json:"date"`
		AvgWeightG    decimal.Decimal     `json:"avg_weight_g"`
		CvPct         decimal.NullDecimal `json:"cv_pct"`
		UniformityPct decimal.NullDecimal `json:"uniformity_pct"`
		SampleSize    *int32              `json:"sample_size"`
	} `json:"rows"`
}

// guide-growth.json
type guideGrowthFile struct {
	Rows []struct {
		GuideID     string              `json:"guide_id"`
		Week        int16               `json:"week"`
		Sex         dbq.Sex             `json:"sex"`
		FeedG       decimal.NullDecimal `json:"feed_g"`
		BodyWeightG decimal.NullDecimal `json:"body_weight_g"`
	} `json:"rows"`
}

// guide-laying.json
type guideLayingFile struct {
	Rows []struct {
		GuideID         string              `json:"guide_id"`
		AgeWeek         int16               `json:"age_week"`
		Sex             dbq.Sex             `json:"sex"`
		FeedG           decimal.NullDecimal `json:"feed_g"`
		LayPct          decimal.NullDecimal `json:"lay_pct"`
		MortalityCumPct decimal.NullDecimal `json:"mortality_cum_pct"`
	} `json:"rows"`
}

// hirbovat-daily.json, as stored: blocks of array rows.
type hirbovatDailyFile struct {
	SectionDay table `json:"section_day"`
	Sales      table `json:"sales"`
	HallDay    table `json:"hall_day"`
}

// hirbovatDaily is hirbovat-daily.json with its blocks decoded.
type hirbovatDaily struct {
	sectionDay []hirbovatSectionDay
	sales      []hirbovatSale
	hallDay    []hirbovatHallDay
}

type hirbovatSectionDay struct {
	Date    timex.Day           `json:"date"`
	Section string              `json:"section"`
	Sex     dbq.Sex             `json:"sex"`
	Mort    int32               `json:"mort"`
	Sold    int32               `json:"sold"`
	Arr     int32               `json:"arr"`
	Mv      int32               `json:"mv"`
	Alive   int32               `json:"alive"`
	G       decimal.NullDecimal `json:"g"`
	Kg      decimal.NullDecimal `json:"kg"`
}

type hirbovatSale struct {
	Date    timex.Day `json:"date"`
	Section string    `json:"section"`
	F       int32     `json:"F"`
	M       int32     `json:"M"`
}

type hirbovatHallDay struct {
	Date timex.Day           `json:"date"`
	Hall string              `json:"hall"`
	Eggs *int32              `json:"eggs"`
	Inc  *int32              `json:"inc"`
	Jos  *int32              `json:"jos"`
	EggG decimal.NullDecimal `json:"egg_g"`
	Cat1 *int32              `json:"cat1"`
	Cat2 *int32              `json:"cat2"`
	Cat3 *int32              `json:"cat3"`
}

// eggs-hirbovat-prev-daily.json (Seria 2025, old hall codes).
type prevEggsFile struct {
	Rows []struct {
		Date          timex.Day           `json:"date"`
		Hall          string              `json:"hall"`
		FloorEggs     *int32              `json:"floor_eggs"`
		Cat1          *int32              `json:"cat1"`
		Cat2          *int32              `json:"cat2"`
		Cat3          *int32              `json:"cat3"`
		HatchingEggs  *int32              `json:"hatching_eggs"`
		TotalEggs     *int32              `json:"total_eggs"`
		AvgEggWeightG decimal.NullDecimal `json:"avg_egg_weight_g"`
		Mortality     map[string]struct {
			F int32 `json:"F"`
			M int32 `json:"M"`
		} `json:"mortality"`
		Depopulation bool `json:"depopulation"`
	} `json:"rows"`
}
