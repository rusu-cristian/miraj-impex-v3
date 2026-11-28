package ledger

// The bounds of §6.1 step 3 for the field a refused write edited. Both take
// the typed value and the pair's MinAliveEnd (< 0, from CheckCounts): the
// field moves the count by the same amount on every later day, so it must
// change by −MinAliveEnd to bring the lowest day back to 0.

// RemovalMax is the most a removal field (mortality, cull, sale, a transfer
// out) may hold: typed + minAliveEnd. Removal 4 292 with a lowest count of −8
// → 4 284. It is below 0 when this field alone can't keep the count ≥ 0.
func RemovalMax(typed, minAliveEnd int) int { return typed + minAliveEnd }

// InflowMin is the least an inflow field being lowered (a placement, a
// transfer in) may hold: typed − minAliveEnd. Inflow 4 381 with a lowest
// count of −2 → 4 383.
func InflowMin(typed, minAliveEnd int) int { return typed - minAliveEnd }

// RemovalMax is the bound of a removal field of this pair (§6.1 step 3).
func (n Negative) RemovalMax(typed int) int { return RemovalMax(typed, n.MinAliveEnd) }

// InflowMin is the bound of a lowered inflow field of this pair (§6.1 step 3).
func (n Negative) InflowMin(typed int) int { return InflowMin(typed, n.MinAliveEnd) }
