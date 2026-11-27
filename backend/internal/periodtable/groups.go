package periodtable

import (
	"slices"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
)

// Record-type chips (§7.4 `types`), in the order the FilterBar lists them.
const (
	ChipMortality   = "mortalitate"
	ChipCull        = "sacrificare"
	ChipSexingError = "eroare_sexare"
	ChipTransfer    = "transfer"
	ChipPlacement   = "populare"
	ChipSale        = "vanzare"
	ChipFeed        = "hrana"
	ChipWeighing    = "cantarire"
	ChipEggs        = "oua"
)

// Chips is every chip in §7.4 order.
var Chips = []string{ChipMortality, ChipCull, ChipSexingError, ChipTransfer, ChipPlacement, ChipSale, ChipFeed, ChipWeighing, ChipEggs}

// Rezumat group ids (§7.4), in their order.
const (
	GroupAlive  = "vii"
	GroupDeaths = "morti"
	GroupLay    = "ouat"
	GroupFeedKg = "hranaKg"
	GroupSale   = "vanzare"
	GroupFeedG  = "gPasare"   // growth: g/pasăre ♀ · ♂ (efectiv · ghid)
	GroupWeight = "greutateF" // growth: Greutate ♀ (weighing weeks)
)

// Leaf kinds (§7.4).
const (
	KindCount = "count"
	KindLevel = "level"
	KindRate  = "rate"
	KindGuide = "guide"
	// KindStatus is a lamp column the client fills from the row's status and
	// missing (Consum "stare", M3 Q2); it has no cell.
	KindStatus = "status"
)

// Cell keys of the Rezumat farm group.
const (
	KeyAliveF   = "alive.F"
	KeyAliveM   = "alive.M"
	KeyDeathsF  = "deaths.F"
	KeyDeathsM  = "deaths.M"
	KeyEggs     = "egg.total"
	KeyLayPct   = "egg.layPct"
	KeyLayGuide = "egg.guide"
	KeyInc      = "egg.inc"
	KeyIncPct   = "egg.incPct"
	KeyJos      = "egg.jos"
	KeyJosPct   = "egg.josPct"
	KeyFeedKgF  = "feed.kg.F"
	KeyFeedKgM  = "feed.kg.M"
	KeySoldF    = "sold.F"
	KeySoldM    = "sold.M"

	KeyFeedGF       = "feed.g.F"
	KeyFeedGGuideF  = "feed.g.guide.F"
	KeyFeedGM       = "feed.g.M"
	KeyFeedGGuideM  = "feed.g.guide.M"
	KeyWeightF      = "weight.F"
	KeyWeightGuideF = "weight.guide.F"
	KeyWeightDiffF  = "weight.diffPct.F"

	// Consum (view=report-feed, M3 Q2 (b)): the g/pasăre difference from the
	// guide in %, and the "stare" leaf (no cell).
	KeyFeedGDiffF = "feed.g.diffPct.F"
	KeyFeedGDiffM = "feed.g.diffPct.M"
	KeyFeedState  = "feed.state"
)

// Reasons of a null cell (§7.4).
const (
	ReasonMissing    = "missing"
	ReasonNoGuide    = "no_guide"
	ReasonNoWeighing = "no_weighing"
	ReasonEmptying   = "emptying" // "– golire, nu mortalitate" (M2 Q6 e, proposed)

	// ReasonEmptySince prefixes the day an emptied hall's Consum line is
	// empty since: "empty_since:2026-11-14" (M3 Q2 (a)), under ReasonAllCells.
	ReasonEmptySince = "empty_since:"
	// ReasonAllCells is the reasons key of a row whose every cell is "–".
	ReasonAllCells = "*"
)

// scopeFarm is Group.Scope of the farm group.
const scopeFarm = "farm"

// groupChips are the chips that show each Rezumat record group (M2 D3):
// mortalitate and sacrificare → morti, oua → ouat, hrana → hranaKg,
// vanzare → vanzare. vii has none: it is always shown.
var groupChips = map[string][]string{
	GroupDeaths: {ChipMortality, ChipCull},
	GroupLay:    {ChipEggs},
	GroupFeedKg: {ChipFeed},
	GroupSale:   {ChipSale},
	GroupFeedG:  {ChipFeed},
	GroupWeight: {ChipWeighing},
}

// showGroup reports whether a record group is shown: one of its chips has
// records in the period (D45) and, when the query names types, is among them.
func showGroup(id string, periodChips, types []string) bool {
	for _, c := range groupChips[id] {
		if slices.Contains(periodChips, c) && (types == nil || slices.Contains(types, c)) {
			return true
		}
	}
	return false
}

// layingRezumatGroups are the Rezumat groups of a laying cycle (§7.4):
// vii · morti · ouat · hranaKg · vanzare, each record group only when it is
// shown; without a guide the ouat group has no guide leaf.
func layingRezumatGroups(periodChips, types []string, noGuide bool) []Group {
	groups := []Group{{ID: GroupAlive, Record: GroupAlive, Label: messages.GroupAlive(), Scope: scopeFarm, Leaves: sexed("alive", KindLevel)}}
	if showGroup(GroupDeaths, periodChips, types) {
		groups = append(groups, Group{ID: GroupDeaths, Record: ChipMortality, Label: messages.GroupDeaths(), Scope: scopeFarm, Leaves: sexed("deaths", KindCount)})
	}
	if showGroup(GroupLay, periodChips, types) {
		leaves := []Leaf{
			{ID: KeyEggs, Label: messages.LeafEggs(), Kind: KindCount},
			{ID: KeyLayPct, Label: messages.LeafLayPct(), Kind: KindRate},
			{ID: KeyLayGuide, Label: messages.LeafGuide(), Kind: KindGuide},
			{ID: KeyInc, Label: messages.LeafHatching(), Kind: KindCount},
			{ID: KeyIncPct, Label: messages.LeafHatchingPct(), Kind: KindRate},
			{ID: KeyJos, Label: messages.LeafFloor(), Kind: KindCount},
			{ID: KeyJosPct, Label: messages.LeafFloorPct(), Kind: KindRate},
		}
		if noGuide {
			leaves = slices.DeleteFunc(leaves, func(l Leaf) bool { return l.ID == KeyLayGuide })
		}
		groups = append(groups, Group{ID: GroupLay, Record: ChipEggs, Label: messages.GroupLay(), Scope: scopeFarm, Leaves: leaves})
	}
	if showGroup(GroupFeedKg, periodChips, types) {
		groups = append(groups, Group{ID: GroupFeedKg, Record: ChipFeed, Label: messages.GroupFeedKg(), Scope: scopeFarm, Leaves: sexed("feed.kg", KindCount)})
	}
	if showGroup(GroupSale, periodChips, types) {
		groups = append(groups, Group{ID: GroupSale, Record: ChipSale, Label: messages.GroupSale(), Scope: scopeFarm, Leaves: sexed("sold", KindCount)})
	}
	return groups
}

// sexed is the ♀ and ♂ leaves "{prefix}.F", "{prefix}.M".
func sexed(prefix, kind string) []Leaf {
	return []Leaf{
		{ID: prefix + ".F", Label: messages.LeafHens(), Kind: kind},
		{ID: prefix + ".M", Label: messages.LeafRoosters(), Kind: kind},
	}
}

// growthRezumatGroups are the Rezumat groups of a growth cycle (§7.4):
// vii · morti · g/pasăre ♀ · ♂ (efectiv · ghid) · Greutate ♀ · hranaKg ·
// vanzare, each record group only when it is shown; without a guide the
// guide leaves are left out.
func growthRezumatGroups(periodChips, types []string, noGuide bool) []Group {
	groups := []Group{{ID: GroupAlive, Record: GroupAlive, Label: messages.GroupAlive(), Scope: scopeFarm, Leaves: sexed("alive", KindLevel)}}
	if showGroup(GroupDeaths, periodChips, types) {
		groups = append(groups, Group{ID: GroupDeaths, Record: ChipMortality, Label: messages.GroupDeaths(), Scope: scopeFarm, Leaves: sexed("deaths", KindCount)})
	}
	if showGroup(GroupFeedG, periodChips, types) {
		leaves := []Leaf{
			{ID: KeyFeedGF, Label: messages.SexLeaf("F", messages.LeafActual()), Kind: KindRate},
			{ID: KeyFeedGGuideF, Label: messages.SexLeaf("F", messages.LeafGuide()), Kind: KindGuide},
			{ID: KeyFeedGM, Label: messages.SexLeaf("M", messages.LeafActual()), Kind: KindRate},
			{ID: KeyFeedGGuideM, Label: messages.SexLeaf("M", messages.LeafGuide()), Kind: KindGuide},
		}
		if noGuide {
			leaves = slices.DeleteFunc(leaves, func(l Leaf) bool { return l.Kind == KindGuide })
		}
		groups = append(groups, Group{ID: GroupFeedG, Record: ChipFeed, Label: messages.GroupFeedG(), Scope: scopeFarm, Leaves: leaves})
	}
	if showGroup(GroupWeight, periodChips, types) {
		leaves := []Leaf{
			{ID: KeyWeightF, Label: messages.LeafWeightMean(), Kind: KindLevel},
			{ID: KeyWeightGuideF, Label: messages.LeafGuide(), Kind: KindGuide},
			{ID: KeyWeightDiffF, Label: messages.LeafVsGuide(), Kind: KindRate},
		}
		if noGuide {
			leaves = leaves[:1]
		}
		groups = append(groups, Group{ID: GroupWeight, Record: ChipWeighing, Label: messages.GroupWeightF(), Scope: scopeFarm, Leaves: leaves})
	}
	if showGroup(GroupFeedKg, periodChips, types) {
		groups = append(groups, Group{ID: GroupFeedKg, Record: ChipFeed, Label: messages.GroupFeedKg(), Scope: scopeFarm, Leaves: sexed("feed.kg", KindCount)})
	}
	if showGroup(GroupSale, periodChips, types) {
		groups = append(groups, Group{ID: GroupSale, Record: ChipSale, Label: messages.GroupSale(), Scope: scopeFarm, Leaves: sexed("sold", KindCount)})
	}
	return groups
}

// legacyRezumatGroups are the Rezumat groups of a past cycle without a flock
// count (D52, D54, M2 Q7): no vii (no counts), morti from the register, ouat
// without % ouat and guide (no hen-days), then hranaKg and vanzare as usual.
func legacyRezumatGroups(periodChips, types []string) []Group {
	groups := layingRezumatGroups(periodChips, types, true)[1:] // drop vii
	for i := range groups {
		if groups[i].ID == GroupLay {
			groups[i].Leaves = slices.DeleteFunc(groups[i].Leaves, func(l Leaf) bool { return l.ID == KeyLayPct })
		}
	}
	return groups
}

// consumGroups are the Consum groups (§7.4 report-feed, S11, M3 Q2):
// "Hrană · kg" ♀ · ♂, then "Hrană · g/pasăre" with per sex efectiv · ghid ·
// față de ghid, and the "stare" lamp. Without a guide the ghid and față de
// ghid leaves are left out.
func consumGroups(noGuide bool) []Group {
	leaves := []Leaf{
		{ID: KeyFeedGF, Label: messages.SexLeaf("F", messages.LeafActual()), Kind: KindRate},
		{ID: KeyFeedGGuideF, Label: messages.SexLeaf("F", messages.LeafGuide()), Kind: KindGuide},
		{ID: KeyFeedGDiffF, Label: messages.SexLeaf("F", messages.LeafVsGuide()), Kind: KindRate},
		{ID: KeyFeedGM, Label: messages.SexLeaf("M", messages.LeafActual()), Kind: KindRate},
		{ID: KeyFeedGGuideM, Label: messages.SexLeaf("M", messages.LeafGuide()), Kind: KindGuide},
		{ID: KeyFeedGDiffM, Label: messages.SexLeaf("M", messages.LeafVsGuide()), Kind: KindRate},
		{ID: KeyFeedState, Label: messages.LeafState(), Kind: KindStatus},
	}
	if noGuide {
		leaves = slices.DeleteFunc(leaves, func(l Leaf) bool {
			return l.Kind == KindGuide || l.ID == KeyFeedGDiffF || l.ID == KeyFeedGDiffM
		})
	}
	return []Group{
		{ID: GroupFeedKg, Record: ChipFeed, Label: messages.GroupConsumKg(), Scope: scopeFarm, Leaves: sexed("feed.kg", KindCount)},
		{ID: GroupFeedG, Record: ChipFeed, Label: messages.GroupConsumG(), Scope: scopeFarm, Leaves: leaves},
	}
}

// leafSet is the set of leaf ids of groups: a row writes only these cells.
func leafSet(groups []Group) map[string]bool {
	set := map[string]bool{}
	for _, g := range groups {
		for _, l := range g.Leaves {
			set[l.ID] = true
		}
	}
	return set
}
