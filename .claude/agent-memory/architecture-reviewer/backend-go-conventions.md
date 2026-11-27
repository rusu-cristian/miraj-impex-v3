---
name: backend-go-conventions
description: Inferred layering conventions of the Go backend (M0 baseline) to compare later milestones against
metadata:
  type: project
---

Inferred from M0 review (backend/, SPEC §12.3 is the source):
- Modules expose NewService / NewHandler / (*Handler).Routes(httpx.Router, httpx.Chain). Global middleware wraps the mux in cmd/api/wiring.go; per-route (session, role) goes through the chain passed to Routes.
- Spec rule: dbq types stay inside services, DTOs built in each module's dto.go. M0 auth leaks dbq.UserRole into Principal, RequireRole and the DTOs (spec sketch uses auth.Vet / p.Role.In). Check later modules do not spread it.
- platform/httpx imports internal/db (for MapError). db imports platform/config and timex. Watch for cycles / dbq creeping into platform.
- Auth session middleware does 2 queries per request (session join, then user_farm list); optimise if later milestones add per-request lookups.
- Health, route recorder and readiness SQL live in cmd/api (package main), not internal.

**How to apply:** when reviewing M1+ compare new modules to internal/auth (Service/Handler/dto split) and flag per-request extra queries or dbq in handlers.
