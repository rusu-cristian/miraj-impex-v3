package auth

import "slices"

// Role is a user's role (SPEC §3). The values are the PostgreSQL user_role
// enum; the service converts from dbq.UserRole at the edge, so callers never
// import dbq for it.
type Role string

// The roles of SPEC §3.
const (
	Owner   Role = "owner"   // Proprietar
	Manager Role = "manager" // Manager
	Vet     Role = "vet"     // Veterinar
	Viewer  Role = "viewer"  // Observator
)

// In reports whether r is one of roles (p.Role.In(auth.Manager, auth.Owner)).
func (r Role) In(roles ...Role) bool { return slices.Contains(roles, r) }

// FarmType is a farm's type: "laying" or "growth" (the farm_type enum).
type FarmType string

// The farm types.
const (
	FarmLaying FarmType = "laying"
	FarmGrowth FarmType = "growth"
)
