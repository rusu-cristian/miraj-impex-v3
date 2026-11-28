package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// The auth rows of the SPEC §3.3 permission matrix (§11.8, plan D6): every
// route `api serve` registers, called by every kind of caller through the
// full handler chain. A route registered without a row here fails
// TestAuthMatrix with the route named; later milestones add their rows.

// caller is who makes a request.
type caller string

const (
	anonymous   caller = "anonymous"   // no cookie; on login it signs in with the vet's credentials
	vet         caller = "vet"         // open session; farm Hîrbovăț
	viewer      caller = "viewer"      // open session; farm Telița
	manager     caller = "manager"     // open session; farms Hîrbovăț and Telița
	owner       caller = "owner"       // open session; no user_farm rows, sees every farm
	revoked     caller = "revoked"     // a revoked session of an active vet
	expired     caller = "expired"     // an expired session of an active vet
	deactivated caller = "deactivated" // an open session of a deactivated vet
)

var callers = []caller{anonymous, vet, viewer, manager, owner, revoked, expired, deactivated}

// matrixRow is one route and the status each caller gets.
type matrixRow struct {
	route string // the pattern exactly as registered ("POST /api/auth/login")
	// path builds the concrete path (wildcards filled, query added) from the
	// farms the test creates at run time; nil = route's path.
	path func(matrixFarms) string
	// signIn: the body is the caller's own credentials (login, password).
	signIn bool
	// body builds the request body from the run-time fixture; nil = none.
	body func(matrixFarms) string
	// account: a 200 body is the caller's account (user, farms).
	account bool
	// stockFarms: a 200 body's stock[].farm are exactly the caller's farms, in
	// name order (GET /api/reports/feed: a viewer sees only their farms).
	stockFarms bool
	want       map[caller]int
	// code is the error.code a caller's refusal must carry, where the status
	// alone can't tell (403 FORBIDDEN vs FARM_FORBIDDEN).
	code map[caller]string
	// also are the same route on other targets (another farm, an unknown
	// id), each with its own statuses. They belong to the row, so deleting
	// the row still leaves its route without one.
	also []matrixCase
}

// matrixCase is a row's route on another target.
type matrixCase struct {
	name string // the target, in the subtest name ("Telița")
	path func(matrixFarms) string
	body func(matrixFarms) string // nil = none
	want map[caller]int
	code map[caller]string
}

// matrixFarms are the farms TestAuthMatrix creates; their ids exist only at
// run time, so farm-scoped rows build their paths and bodies from them.
// section (a compartment of Hîrbovăț, no birds) and feedType (active, ♀) let
// a Hrană zilnică or a Livrare hrană body pass level 1 and reach the service
// (M3 T8, M4 T7). series has a closed cycle on Telița (the viewer's farm),
// cojusnaSeries one on Cojușna only (the owner's), for the series page.
//
// M5: birds is a section of Hîrbovăț with birds (♀ 100 · ♂ 10 from 1.11),
// for the Zilnic dry runs. entry is a Zilnic entry of 17.11 by another vet
// (Dr. Autor: ♀ 1 death there, its create in audit_log) at entryVersion;
// telEntry a priced delivery of Telița without an audit row; deleted a
// Zilnic entry of 16.11 Dr. Autor deleted today (his create and delete in
// audit_log).
type matrixFarms struct {
	hirb, tel, cojusna       uuid.UUID
	section, feedType        uuid.UUID
	series, cojusnaSeries    uuid.UUID
	birds                    uuid.UUID
	entry, telEntry, deleted uuid.UUID
	entryVersion             string
}

// unknownFarm is a farm id that no farm has.
var unknownFarm = uuid.MustParse("0b9e4b8e-6f0a-4c51-9d3e-2f7a1c5d8e90")

func everyone(status int) map[caller]int {
	m := make(map[caller]int, len(callers))
	for _, c := range callers {
		m[c] = status
	}
	return m
}

var authMatrix = []matrixRow{
	// Public, outside /api (SPEC §11.12): 200 whatever the cookie.
	{route: "GET /healthz", want: everyone(http.StatusOK)},
	{route: "GET /readyz", want: everyone(http.StatusOK)},

	// The only /api route without a session (§7.0). Every active account signs
	// in, whatever state its old session is in; a deactivated one is refused
	// after its correct password (§6.16).
	{route: "POST /api/auth/login", signIn: true, account: true, want: map[caller]int{
		anonymous: 200, vet: 200, viewer: 200, manager: 200, owner: 200,
		revoked: 200, expired: 200, deactivated: http.StatusForbidden,
	}},
	{route: "POST /api/auth/logout", want: map[caller]int{
		anonymous: 401, vet: 204, viewer: 204, manager: 204, owner: 204,
		revoked: 401, expired: 401, deactivated: 401,
	}},
	{route: "GET /api/auth/me", account: true, want: map[caller]int{
		anonymous: 401, vet: 200, viewer: 200, manager: 200, owner: 200,
		revoked: 401, expired: 401, deactivated: 401,
	}},

	// The farm scope pages (M2 D4): viewer, manager and owner, each on the
	// farms they may see. Vets never reach them (403 FORBIDDEN); the viewer
	// of Telița is refused Hîrbovăț and Cojușna (403 FARM_FORBIDDEN), and so
	// is anyone but the owner on an unknown id, which the owner gets as 404.
	// GET /api/table carries every view's cases in its one row: Rezumat
	// (view=farm) and, from M3, Consum (view=report-feed), with the same
	// statuses (M3 T8).
	withCases(farmRow("GET /api/table", func(id uuid.UUID) string { return "/api/table?view=farm&farmId=" + id.String() }),
		"report-feed", farmRow("GET /api/table", func(id uuid.UUID) string {
			return "/api/table?view=report-feed&farmId=" + id.String()
		})),
	farmRow("GET /api/farms/{id}/page", func(id uuid.UUID) string { return "/api/farms/" + id.String() + "/page" }),
	farmRow("GET /api/farms/{id}/summary", func(id uuid.UUID) string { return "/api/farms/" + id.String() + "/summary" }),

	// The shell bootstrap (M2 D7): every active session, scoped to its farms.
	{route: "GET /api/bootstrap", want: map[caller]int{
		anonymous: 401, vet: 200, viewer: 200, manager: 200, owner: 200,
		revoked: 401, expired: 401, deactivated: 401,
	}},

	// Hrană zilnică (§7.8, M3 D5, D8): vet, manager and owner on their farms;
	// the viewer never (403 FORBIDDEN). The form state of the matrix farm (no
	// cycle, no birds) is 200 with rows []; the save passes level 1 and is
	// refused by the service, 422 REFUSED (H3: the section has no birds).
	writerRow("GET /api/forms/feed-daily", http.StatusOK, http.StatusOK,
		func(id uuid.UUID) string { return "/api/forms/feed-daily?farmId=" + id.String() + "&date=2026-11-18" }, nil),
	writerRow("POST /api/entries/feed-daily", http.StatusUnprocessableEntity, http.StatusBadRequest,
		func(uuid.UUID) string { return "/api/entries/feed-daily" }, feedDailyBody),

	// Rapoarte › Hrană (§7.12, M3 D9): viewer, manager and owner; vets never
	// reach Rapoarte. farmId=all (the default) is the caller's farms only.
	{route: "GET /api/reports/feed", stockFarms: true, want: map[caller]int{
		anonymous: 401, vet: 403, viewer: 200, manager: 200, owner: 200,
		revoked: 401, expired: 401, deactivated: 401,
	}, code: map[caller]string{vet: httpx.CodeForbidden},
		also: []matrixCase{{name: "Hîrbovăț",
			path: func(f matrixFarms) string { return "/api/reports/feed?farmId=" + f.hirb.String() },
			want: map[caller]int{
				anonymous: 401, vet: 403, viewer: 403, manager: 200, owner: 200,
				revoked: 401, expired: 401, deactivated: 401,
			},
			code: map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeFarmForbidden}}}},

	// The series page (§7.3, M4 Q4): viewer, manager and owner, each on a
	// series with a cycle on one of their farms; vets never reach it (403
	// FORBIDDEN). A series without a cycle on the caller's farms, or an
	// unknown one, is 404 NOT_FOUND for every other role (no existence leak).
	{route: "GET /api/series/{id}/page",
		path: func(f matrixFarms) string { return seriesPagePath(f.series) },
		want: map[caller]int{
			anonymous: 401, vet: 403, viewer: 200, manager: 200, owner: 200,
			revoked: 401, expired: 401, deactivated: 401,
		},
		code: map[caller]string{vet: httpx.CodeForbidden},
		also: []matrixCase{
			{name: "Cojușna's series", path: func(f matrixFarms) string { return seriesPagePath(f.cojusnaSeries) },
				want: map[caller]int{
					anonymous: 401, vet: 403, viewer: 404, manager: 404, owner: 200,
					revoked: 401, expired: 401, deactivated: 401,
				},
				code: map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeNotFound, manager: httpx.CodeNotFound}},
			{name: "unknown series", path: func(matrixFarms) string { return seriesPagePath(unknownFarm) },
				want: map[caller]int{
					anonymous: 401, vet: 403, viewer: 404, manager: 404, owner: 404,
					revoked: 401, expired: 401, deactivated: 401,
				},
				code: map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeNotFound,
					manager: httpx.CodeNotFound, owner: httpx.CodeNotFound}},
		}},

	// Livrare hrană (§6.7, §7.8, M4 D5, Q14): manager and owner on their
	// farms; vets and viewers never (403 FORBIDDEN). The save is a dry run
	// that passes level 1 and answers 200 (nothing written). A manager's body
	// with a price field is 403 PRICE_FORBIDDEN, the owner's 400 (M4 Q9).
	deliveryRow("GET /api/forms/delivery", func(id uuid.UUID) string { return "/api/forms/delivery?farmId=" + id.String() }, nil),
	withPriceField(deliveryRow("POST /api/entries/delivery", func(uuid.UUID) string { return "/api/entries/delivery" },
		func(f matrixFarms, farm uuid.UUID) string { return deliveryDryRun(f, farm, "") })),

	// Prețuri hrană (§3.3, §7.13, M4 Q10): the owner's only. The S12 list is
	// 403 FORBIDDEN for every other role; the writes are 403 PRICE_FORBIDDEN
	// for a manager ("○ sees, cannot set") and 403 FORBIDDEN for the vet and
	// the viewer. The owner reaches the service, which refuses the matrix's
	// requests: an unknown delivery in a batch and a forged undo token → 400,
	// a change of an unknown delivery → 404.
	{route: "GET /api/deliveries", path: func(matrixFarms) string { return "/api/deliveries?priced=false" },
		want: map[caller]int{
			anonymous: 401, vet: 403, viewer: 403, manager: 403, owner: 200,
			revoked: 401, expired: 401, deactivated: 401,
		},
		code: map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeForbidden, manager: httpx.CodeForbidden}},
	priceWriteRow("POST /api/deliveries/prices", "/api/deliveries/prices",
		`{"items":[{"id":"`+unknownFarm.String()+`","pricePerKg":8}]}`, http.StatusBadRequest, httpx.CodeValidation),
	priceWriteRow("POST /api/deliveries/prices/undo", "/api/deliveries/prices/undo",
		`{"undoToken":"nu-e-un-token"}`, http.StatusBadRequest, httpx.CodeValidation),
	priceWriteRow("PATCH /api/deliveries/{id}/price", "/api/deliveries/"+unknownFarm.String()+"/price",
		`{"pricePerKg":7.74,"reason":"Factura corectată"}`, http.StatusNotFound, httpx.CodeNotFound),

	// Zilnic (§7.8, M5-D13): vet, manager and owner on their farms, as
	// Hrană zilnică; the viewer never (403 FORBIDDEN). The form state of the
	// matrix farm is 200; the save is a dry run of today on the section with
	// birds (200, nothing written). A vet's save of another day is 403
	// PAST_DAY_VET (Z1); a manager's or the owner's is a dry run like today's.
	writerRow("GET /api/forms/daily", http.StatusOK, http.StatusOK,
		func(id uuid.UUID) string { return "/api/forms/daily?farmId=" + id.String() + "&date=2026-11-18" }, nil),
	withPastDay(writerRow("POST /api/entries/daily", http.StatusOK, http.StatusBadRequest,
		func(uuid.UUID) string { return dailyPath }, func(f matrixFarms, farm uuid.UUID) string {
			return dailyDryRun(f, farm, "2026-11-18")
		})),

	// One saved entry (§7.8, M5 Q10). The read: any role that sees every farm
	// of the entry, the viewer included; another farm's entry is 403
	// FARM_FORBIDDEN, an unknown one 404 for everyone.
	entryRow("GET /api/entries/{id}", func(f matrixFarms) uuid.UUID { return f.entry }, nil,
		map[caller]int{vet: 200, viewer: 403, manager: 200, owner: 200},
		map[caller]string{viewer: httpx.CodeFarmForbidden},
		matrixCase{name: "Telița's entry", path: entryPath(func(f matrixFarms) uuid.UUID { return f.telEntry }, ""),
			want: withSessions(map[caller]int{vet: 403, viewer: 200, manager: 200, owner: 200}),
			code: map[caller]string{vet: httpx.CodeFarmForbidden}}),
	// The writes on the entry of 17.11 by another vet (Dr. Autor): the viewer
	// is refused (403), the vet too (another's entry, 403 FORBIDDEN); the
	// manager and the owner reach the service. PUT and DELETE are corrections
	// with a Motiv in dry run (200, nothing written), PUT with an old
	// ifVersion 409 STALE.
	entryRow("PUT /api/entries/{id}", func(f matrixFarms) uuid.UUID { return f.entry },
		func(f matrixFarms) string { return correctionDryRun(f, f.entryVersion) },
		map[caller]int{vet: 403, viewer: 403, manager: 200, owner: 200},
		map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeForbidden},
		matrixCase{name: "old ifVersion", path: entryPath(func(f matrixFarms) uuid.UUID { return f.entry }, ""),
			body: func(f matrixFarms) string { return correctionDryRun(f, "2026-11-16T06:00:00.000000Z") },
			want: withSessions(map[caller]int{vet: 403, viewer: 403, manager: 409, owner: 409}),
			code: map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeForbidden,
				manager: httpx.CodeStale, owner: httpx.CodeStale}}),
	entryRow("DELETE /api/entries/{id}", func(f matrixFarms) uuid.UUID { return f.entry },
		func(matrixFarms) string { return `{"reason":"Introdusă de două ori","dryRun":true}` },
		map[caller]int{vet: 403, viewer: 403, manager: 200, owner: 200},
		map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeFarmForbidden}),
	// Undo and restore act only on the caller's own latest action of today
	// (Q8): on Dr. Autor's create (the entry of 17.11) and his delete of
	// today (the deleted entry) every writer is refused, 403 FORBIDDEN; on
	// an entry without an action of that kind (Telița's delivery for undo,
	// the live entry for restore) the manager and the owner get 409 STALE.
	entryRow("POST /api/entries/{id}/undo", func(f matrixFarms) uuid.UUID { return f.entry },
		func(matrixFarms) string { return `{}` },
		map[caller]int{vet: 403, viewer: 403, manager: 403, owner: 403},
		map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeFarmForbidden,
			manager: httpx.CodeForbidden, owner: httpx.CodeForbidden},
		matrixCase{name: "Telița's entry", path: entryPath(func(f matrixFarms) uuid.UUID { return f.telEntry }, "/undo"),
			body: func(matrixFarms) string { return `{}` },
			want: withSessions(map[caller]int{vet: 403, viewer: 403, manager: 409, owner: 409}),
			code: map[caller]string{vet: httpx.CodeFarmForbidden, viewer: httpx.CodeForbidden,
				manager: httpx.CodeStale, owner: httpx.CodeStale}}),
	entryRow("POST /api/entries/{id}/restore", func(f matrixFarms) uuid.UUID { return f.deleted },
		func(matrixFarms) string { return `{}` },
		map[caller]int{vet: 403, viewer: 403, manager: 403, owner: 403},
		map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeFarmForbidden,
			manager: httpx.CodeForbidden, owner: httpx.CodeForbidden},
		matrixCase{name: "live entry", path: entryPath(func(f matrixFarms) uuid.UUID { return f.entry }, "/restore"),
			body: func(matrixFarms) string { return `{}` },
			want: withSessions(map[caller]int{vet: 409, viewer: 403, manager: 409, owner: 409}),
			code: map[caller]string{vet: httpx.CodeStale, viewer: httpx.CodeFarmForbidden,
				manager: httpx.CodeStale, owner: httpx.CodeStale}}),
}

const dailyPath = "/api/entries/daily"

// withSessions completes want with 401 for every caller without an open
// session of an active account.
func withSessions(want map[caller]int) map[caller]int {
	m := map[caller]int{anonymous: 401, revoked: 401, expired: 401, deactivated: 401}
	maps.Copy(m, want)
	return m
}

// entryPath is the path of an entry route on the entry pick names, with
// suffix ("/undo") after the id.
func entryPath(pick func(matrixFarms) uuid.UUID, suffix string) func(matrixFarms) string {
	return func(f matrixFarms) string { return "/api/entries/" + pick(f).String() + suffix }
}

// entryRow is the row of a route on one saved entry (M5): the main target is
// the entry pick names, with body (nil = none) and the statuses and codes of
// the four session callers; extra are other targets. Every row also gets an
// unknown entry: 404 NOT_FOUND for everyone with a session, except a viewer
// on PUT, refused by role before the entry is read.
func entryRow(route string, pick func(matrixFarms) uuid.UUID, body func(matrixFarms) string,
	want map[caller]int, code map[caller]string, extra ...matrixCase) matrixRow {
	_, pattern, _ := strings.Cut(route, " ")
	suffix := strings.TrimPrefix(pattern, "/api/entries/{id}")
	unknown := map[caller]int{vet: 404, viewer: 404, manager: 404, owner: 404}
	unknownCode := map[caller]string{vet: httpx.CodeNotFound, viewer: httpx.CodeNotFound,
		manager: httpx.CodeNotFound, owner: httpx.CodeNotFound}
	if strings.HasPrefix(route, "PUT ") {
		unknown[viewer], unknownCode[viewer] = 403, httpx.CodeForbidden
	}
	return matrixRow{
		route: route, path: entryPath(pick, suffix), body: body,
		want: withSessions(want), code: code,
		also: append(extra, matrixCase{name: "unknown entry",
			path: entryPath(func(matrixFarms) uuid.UUID { return unknownFarm }, suffix), body: body,
			want: withSessions(unknown), code: unknownCode}),
	}
}

// withPastDay adds the Zilnic dry run of 16.11 on Hîrbovăț: the vet is
// refused (403 PAST_DAY_VET, Z1), the manager and the owner reach the dry run
// (200); the viewer is refused by role first.
func withPastDay(row matrixRow) matrixRow {
	row.also = append(row.also, matrixCase{name: "16.11",
		path: func(matrixFarms) string { return dailyPath },
		body: func(f matrixFarms) string { return dailyDryRun(f, f.hirb, "2026-11-16") },
		want: withSessions(map[caller]int{vet: 403, viewer: 403, manager: 200, owner: 200}),
		code: map[caller]string{vet: httpx.CodePastDayVet, viewer: httpx.CodeForbidden}})
	return row
}

// dailyDryRun is a Zilnic dry run on farm and day that passes level 1: one
// ♀ death in the matrix section with birds.
func dailyDryRun(f matrixFarms, farm uuid.UUID, day string) string {
	return `{"farmId":"` + farm.String() + `","date":"` + day + `","sections":[{"sectionId":"` + f.birds.String() +
		`","F":{"mortality":1}}],"dryRun":true}`
}

// correctionDryRun is a Corectează dry run of the matrix entry (♀ deaths
// 1 → 2 on 17.11) at version ifVersion, with a Motiv.
func correctionDryRun(f matrixFarms, ifVersion string) string {
	return `{"ifVersion":"` + ifVersion + `","values":{"farmId":"` + f.hirb.String() + `","date":"2026-11-17","sections":[` +
		`{"sectionId":"` + f.birds.String() + `","F":{"mortality":2}}]},"reason":"Numărătoare refăcută","dryRun":true}`
}

func seriesPagePath(id uuid.UUID) string { return "/api/series/" + id.String() + "/page" }

// deliveryRow is the row of a Livrare hrană route (form state or dry-run
// save): path and body build the request for a farm id. Managers and the
// owner reach the service on their farms: Hîrbovăț (main) for both, Cojușna
// for the owner only (a manager gets 403 FARM_FORBIDDEN), and an unknown id
// is 404 for the owner, 403 FARM_FORBIDDEN for a manager. Vets and viewers
// are refused by role first, everywhere (403 FORBIDDEN).
func deliveryRow(route string, path func(uuid.UUID) string, body func(matrixFarms, uuid.UUID) string) matrixRow {
	at := func(pick func(matrixFarms) uuid.UUID) (func(matrixFarms) string, func(matrixFarms) string) {
		p := func(f matrixFarms) string { return path(pick(f)) }
		if body == nil {
			return p, nil
		}
		return p, func(f matrixFarms) string { return body(f, pick(f)) }
	}
	scoped := func(managerStatus, ownerStatus int) map[caller]int {
		return map[caller]int{
			anonymous: 401, vet: 403, viewer: 403, manager: managerStatus, owner: ownerStatus,
			revoked: 401, expired: 401, deactivated: 401,
		}
	}
	byRole := map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeForbidden}
	with := func(extra map[caller]string) map[caller]string {
		m := maps.Clone(byRole)
		maps.Copy(m, extra)
		return m
	}
	hirbPath, hirbBody := at(func(f matrixFarms) uuid.UUID { return f.hirb })
	cojPath, cojBody := at(func(f matrixFarms) uuid.UUID { return f.cojusna })
	unknownPath, unknownBody := at(func(matrixFarms) uuid.UUID { return unknownFarm })
	return matrixRow{
		route: route, path: hirbPath, body: hirbBody,
		want: scoped(200, 200),
		code: byRole,
		also: []matrixCase{
			{name: "Cojușna", path: cojPath, body: cojBody, want: scoped(403, 200),
				code: with(map[caller]string{manager: httpx.CodeFarmForbidden})},
			{name: "unknown farm", path: unknownPath, body: unknownBody, want: scoped(403, 404),
				code: with(map[caller]string{manager: httpx.CodeFarmForbidden, owner: httpx.CodeNotFound})},
		},
	}
}

// withPriceField adds the delivery save with a price field on Hîrbovăț: 403
// PRICE_FORBIDDEN for the manager, 400 VALIDATION (unknown_field) for the
// owner (M4 Q9); the vet and the viewer are refused by role first.
func withPriceField(row matrixRow) matrixRow {
	row.also = append(row.also, matrixCase{name: "with pricePerKg",
		path: func(matrixFarms) string { return "/api/entries/delivery" },
		body: func(f matrixFarms) string { return deliveryDryRun(f, f.hirb, `"pricePerKg":7.74,`) },
		want: map[caller]int{
			anonymous: 401, vet: 403, viewer: 403, manager: 403, owner: 400,
			revoked: 401, expired: 401, deactivated: 401,
		},
		code: map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeForbidden,
			manager: httpx.CodePriceForbidden, owner: httpx.CodeValidation}})
	return row
}

// deliveryDryRun is a Livrare hrană dry run on farm that passes level 1 (M4
// D5): the matrix's active type, 9 000 kg today; extra is spliced in as
// members before the others (`"pricePerKg":7.74,`).
func deliveryDryRun(f matrixFarms, farm uuid.UUID, extra string) string {
	return `{` + extra + `"farmId":"` + farm.String() + `","date":"2026-11-18","feedTypeId":"` + f.feedType.String() +
		`","qtyKg":9000,"dryRun":true}`
}

// priceWriteRow is the row of a Prețuri write: the owner reaches the service
// (ownerStatus, ownerCode); a manager gets 403 PRICE_FORBIDDEN, the vet and
// the viewer 403 FORBIDDEN, before the body is read.
func priceWriteRow(route, path, body string, ownerStatus int, ownerCode string) matrixRow {
	return matrixRow{
		route: route,
		path:  func(matrixFarms) string { return path },
		body:  func(matrixFarms) string { return body },
		want: map[caller]int{
			anonymous: 401, vet: 403, viewer: 403, manager: 403, owner: ownerStatus,
			revoked: 401, expired: 401, deactivated: 401,
		},
		code: map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeForbidden,
			manager: httpx.CodePriceForbidden, owner: ownerCode},
	}
}

// withCases adds other's targets (its main one and its also) to row's also,
// their names prefixed: one route, several query shapes, one row.
func withCases(row matrixRow, prefix string, other matrixRow) matrixRow {
	row.also = append(row.also, matrixCase{name: prefix + " Hîrbovăț", path: other.path, body: other.body, want: other.want, code: other.code})
	for _, a := range other.also {
		a.name = prefix + " " + a.name
		row.also = append(row.also, a)
	}
	return row
}

// writerRow is the row of a Hrană zilnică route (form state or save): path
// and body build the request for a farm id. Vets, managers and owners reach
// the service on their farms: reached on Hîrbovăț (the vet's farm), onTelita
// for the manager and the owner on Telița, which has no section (the
// Hîrbovăț section in the save's body is not one of its sections: 400). The
// viewer is refused by role first, everywhere (403 FORBIDDEN); the vet on
// Telița and anyone but the owner on an unknown id get 403 FARM_FORBIDDEN,
// the owner 404.
func writerRow(route string, reached, onTelita int, path func(uuid.UUID) string, body func(matrixFarms, uuid.UUID) string) matrixRow {
	at := func(pick func(matrixFarms) uuid.UUID) (func(matrixFarms) string, func(matrixFarms) string) {
		p := func(f matrixFarms) string { return path(pick(f)) }
		if body == nil {
			return p, nil
		}
		return p, func(f matrixFarms) string { return body(f, pick(f)) }
	}
	scoped := func(vetStatus, managerStatus, ownerStatus int) map[caller]int {
		return map[caller]int{
			anonymous: 401, vet: vetStatus, viewer: 403, manager: managerStatus, owner: ownerStatus,
			revoked: 401, expired: 401, deactivated: 401,
		}
	}
	hirbPath, hirbBody := at(func(f matrixFarms) uuid.UUID { return f.hirb })
	telPath, telBody := at(func(f matrixFarms) uuid.UUID { return f.tel })
	unknownPath, unknownBody := at(func(matrixFarms) uuid.UUID { return unknownFarm })
	reachedCode := map[caller]string{viewer: httpx.CodeForbidden}
	if reached == http.StatusUnprocessableEntity {
		for _, c := range []caller{vet, manager, owner} {
			reachedCode[c] = httpx.CodeRefused
		}
	}
	return matrixRow{
		route: route, path: hirbPath, body: hirbBody,
		want: scoped(reached, reached, reached),
		code: reachedCode,
		also: []matrixCase{
			{name: "Telița", path: telPath, body: telBody,
				want: scoped(403, onTelita, onTelita),
				code: map[caller]string{vet: httpx.CodeFarmForbidden, viewer: httpx.CodeForbidden}},
			{name: "unknown farm", path: unknownPath, body: unknownBody,
				want: scoped(403, 403, 404),
				code: map[caller]string{vet: httpx.CodeFarmForbidden, viewer: httpx.CodeForbidden,
					manager: httpx.CodeFarmForbidden, owner: httpx.CodeNotFound}},
		},
	}
}

// feedDailyBody is a Hrană zilnică save of farm that passes level 1 (M3 D5):
// one row of the matrix section, an existing feed type, g within 1–300.
func feedDailyBody(f matrixFarms, farm uuid.UUID) string {
	return `{"farmId":"` + farm.String() + `","date":"2026-11-18","rows":[{"sectionId":"` + f.section.String() +
		`","sex":"F","feedTypeId":"` + f.feedType.String() + `","g":160}]}`
}

// farmRow is the row of a farm-scoped read: path builds the request for a
// farm id. The main target is Hîrbovăț (the viewer's other farm); also are
// Telița (the viewer's farm), Cojușna (the owner's only) and an unknown id.
func farmRow(route string, path func(uuid.UUID) string) matrixRow {
	at := func(pick func(matrixFarms) uuid.UUID) func(matrixFarms) string {
		return func(f matrixFarms) string { return path(pick(f)) }
	}
	// scoped are the statuses with the session check and RequireRole(viewer,
	// manager, owner) in front of the farm check.
	scoped := func(viewerStatus, managerStatus, ownerStatus int) map[caller]int {
		return map[caller]int{
			anonymous: 401, vet: 403, viewer: viewerStatus, manager: managerStatus, owner: ownerStatus,
			revoked: 401, expired: 401, deactivated: 401,
		}
	}
	return matrixRow{
		route: route,
		path:  at(func(f matrixFarms) uuid.UUID { return f.hirb }),
		want:  scoped(403, 200, 200),
		code:  map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeFarmForbidden},
		also: []matrixCase{
			{name: "Telița", path: at(func(f matrixFarms) uuid.UUID { return f.tel }),
				want: scoped(200, 200, 200),
				code: map[caller]string{vet: httpx.CodeForbidden}},
			{name: "Cojușna", path: at(func(f matrixFarms) uuid.UUID { return f.cojusna }),
				want: scoped(403, 403, 200),
				code: map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeFarmForbidden, manager: httpx.CodeFarmForbidden}},
			{name: "unknown farm", path: func(matrixFarms) string { return path(unknownFarm) },
				want: scoped(403, 403, 404),
				code: map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeFarmForbidden,
					manager: httpx.CodeFarmForbidden, owner: httpx.CodeNotFound}},
		},
	}
}

// matrixGaps compares the registered routes with the table: missing are
// routes without a row, stale are rows naming no registered route (or a
// route a previous row already has: one row per route, so deleting a row
// always leaves its route missing), partial are rows or cases that lack a
// caller.
func matrixGaps(routes []string, rows []matrixRow) (missing, stale, partial []string) {
	have := make(map[string]bool, len(rows))
	for _, r := range rows {
		if have[r.route] || !slices.Contains(routes, r.route) {
			stale = append(stale, r.route)
		}
		have[r.route] = true
		for _, c := range callers {
			if _, ok := r.want[c]; !ok {
				partial = append(partial, r.route+" × "+string(c))
			}
			for _, a := range r.also {
				if _, ok := a.want[c]; !ok {
					partial = append(partial, r.route+" ("+a.name+") × "+string(c))
				}
			}
		}
	}
	for _, route := range routes {
		if !have[route] {
			missing = append(missing, route)
		}
	}
	return missing, stale, partial
}

// account is what login and me must return for a caller.
type account struct {
	id    uuid.UUID
	login string
	role  string
	farms []string // names, in name_key order
}

type accountBody struct {
	User struct {
		ID    uuid.UUID `json:"id"`
		Login string    `json:"login"`
		Role  string    `json:"role"`
	} `json:"user"`
	Farms []struct {
		Name string `json:"name"`
	} `json:"farms"`
	Today string `json:"today"`
}

func TestAuthMatrix(t *testing.T) {
	t.Parallel()
	s := newStack(t)

	missing, stale, partial := matrixGaps(s.app.routes, authMatrix)
	for _, r := range missing {
		t.Errorf("route %q is registered but has no row in authMatrix (SPEC §3.3)", r)
	}
	for _, r := range stale {
		t.Errorf("authMatrix row %q names no registered route", r)
	}
	for _, r := range partial {
		t.Errorf("authMatrix has no expected status for %s", r)
	}

	// Farms and accounts. Cojușna has no user_farm rows: only the owner sees it.
	hirb := s.farm(t, "Hîrbovăț", "hirbovat")
	tel := s.farm(t, "Telița", "telita")
	ids := matrixFarms{hirb: hirb, tel: tel, cojusna: s.farm(t, "Cojușna", "cojusna")}
	if err := s.d.App.QueryRow(t.Context(), `insert into section (farm_id, code) values ($1, 'H1') returning id`, hirb).
		Scan(&ids.section); err != nil {
		t.Fatal(err)
	}
	if err := s.d.App.QueryRow(t.Context(), `insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder 2', 'F')
		returning id`).Scan(&ids.feedType); err != nil {
		t.Fatal(err)
	}
	// Series for the series page: one with a closed cycle on Telița, one with
	// a closed cycle on Cojușna only.
	cycleSeries := func(name string, year int, farm uuid.UUID, from, to string) uuid.UUID {
		var id uuid.UUID
		if err := s.d.App.QueryRow(t.Context(), `insert into series (name, name_key, year, placement_date)
			values ($1, lower($1), $2, $3) returning id`, name, year, from).Scan(&id); err != nil {
			t.Fatal(err)
		}
		s.exec(t, `insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by)
			values ($1, $2, $3, $4, 'auto', 'emptied')`, id, farm, from, to)
		return id
	}
	ids.series = cycleSeries("Seria M", 2026, tel, "2026-01-05", "2026-02-28")
	ids.cojusnaSeries = cycleSeries("Seria C", 2027, ids.cojusna, "2026-03-02", "2026-04-30")
	m5MatrixFixture(t, s, &ids)
	accounts := map[caller]account{}
	add := func(c caller, login, name string, role dbq.UserRole, active bool, farms []uuid.UUID, names ...string) {
		accounts[c] = account{id: s.user(t, login, name, role, active, farms...), login: login, role: string(role), farms: names}
	}
	add(vet, "ceban", "Dr. Ceban", dbq.UserRoleVet, true, []uuid.UUID{hirb}, "Hîrbovăț")
	add(viewer, "observator", "Ana Observator", dbq.UserRoleViewer, true, []uuid.UUID{tel}, "Telița")
	add(manager, "ionescu", "Ion Ionescu", dbq.UserRoleManager, true, []uuid.UUID{tel, hirb}, "Hîrbovăț", "Telița")
	add(owner, "proprietar", "Petru Proprietar", dbq.UserRoleOwner, true, nil, "Cojușna", "Hîrbovăț", "Telița")
	add(revoked, "revocat", "Vet Revocat", dbq.UserRoleVet, true, []uuid.UUID{hirb}, "Hîrbovăț")
	add(expired, "expirat", "Vet Expirat", dbq.UserRoleVet, true, []uuid.UUID{hirb}, "Hîrbovăț")
	add(deactivated, "dezactivat", "Vet Dezactivat", dbq.UserRoleVet, false, []uuid.UUID{hirb}, "Hîrbovăț")
	accounts[anonymous] = accounts[vet]

	// Each cell gets its own session, so cells don't depend on each other
	// (a logout cell revokes only its own).
	newCookie := func(t *testing.T, c caller) string {
		switch c {
		case anonymous:
			return ""
		case revoked:
			return s.session(t, accounts[c].id, sessionRevoked)
		case expired:
			return s.session(t, accounts[c].id, sessionExpired)
		default:
			return s.session(t, accounts[c].id, sessionOpen)
		}
	}

	for _, row := range authMatrix {
		method, path, _ := strings.Cut(row.route, " ")
		if row.path != nil {
			path = row.path(ids)
		}
		type target struct {
			name, path string
			body       func(matrixFarms) string
			want       map[caller]int
			code       map[caller]string
		}
		cases := []target{{path: path, body: row.body, want: row.want, code: row.code}}
		for _, a := range row.also {
			cases = append(cases, target{name: a.name, path: a.path(ids), body: a.body, want: a.want, code: a.code})
		}
		for _, mc := range cases {
			name := row.route
			if mc.name != "" {
				name += " (" + mc.name + ")"
			}
			for _, c := range callers {
				want, ok := mc.want[c]
				if !ok {
					continue // reported above
				}
				t.Run(name+"/"+string(c), func(t *testing.T) {
					t.Parallel()
					req := call{method: method, path: mc.path, cookie: newCookie(t, c)}
					if row.signIn {
						req.body = loginBody(accounts[c].login, testPassword)
					}
					if mc.body != nil {
						req.body = mc.body(ids)
					}
					r := s.do(t, req)
					if r.status != want {
						t.Fatalf("%s as %s: %s, want %d", name, c, r, want)
					}
					if code, ok := mc.code[c]; ok {
						if got := errorCode(t, r); got != code {
							t.Fatalf("%s as %s: %s, want error.code %s", name, c, r, code)
						}
					}
					if row.account && r.status == http.StatusOK {
						checkAccount(t, accounts[c], r)
					}
					if row.stockFarms && mc.name == "" && r.status == http.StatusOK {
						checkStockFarms(t, accounts[c], r)
					}
					if (c == viewer || c == vet) && r.status == http.StatusOK {
						noPrices(t, name+" as "+string(c), r) // D2: never sent
					}
				})
			}
		}
	}
}

// m5MatrixFixture adds the M5 rows' data to f (see matrixFarms): the section
// with birds (placed by a migration entry), Dr. Autor (a vet of Hîrbovăț who
// is none of the callers), his entry of 17.11 and his deleted entry of 16.11
// with their audit rows, and Telița's delivery.
func m5MatrixFixture(t *testing.T, s *stack, f *matrixFarms) {
	t.Helper()
	ctx := t.Context()
	id := func(sql string, args ...any) uuid.UUID {
		t.Helper()
		var v uuid.UUID
		if err := s.d.App.QueryRow(ctx, sql, args...).Scan(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	author := s.user(t, "autor", "Dr. Autor", dbq.UserRoleVet, true, f.hirb)
	f.birds = id(`insert into section (farm_id, code) values ($1, 'H2') returning id`, f.hirb)
	placed := id(`insert into entry (kind, farm_id, date, source) values ('migration', $1, '2026-11-01', 'migration') returning id`, f.hirb)
	s.exec(t, `insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
		values ($1, $2, $3, '2026-11-01', 'F', 'placement', 100), ($1, $2, $3, '2026-11-01', 'M', 'placement', 10)`, placed, f.hirb, f.birds)

	zilnic := func(day, at string, sex string) uuid.UUID {
		e := id(`insert into entry (kind, farm_id, date, created_by, created_at) values ('daily', $1, $2, $3, $4) returning id`,
			f.hirb, day, author, at)
		s.exec(t, `insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty) values ($1, $2, $3, $4, $5, 'mortality', -1)`,
			e, f.hirb, f.birds, day, sex)
		s.exec(t, `insert into audit_log (at, user_id, action, entity, entity_id, entry_id, entry_kind, farm_ids, record_date, same_day_author)
			values ($1, $2, 'create', 'entry', $3, $3, 'daily', array[$4::uuid], $5, false)`, at, author, e, f.hirb, day)
		return e
	}
	created := "2026-11-17T08:00:00+02:00"
	f.entry = zilnic("2026-11-17", created, "F")
	at, err := time.Parse(time.RFC3339, created)
	if err != nil {
		t.Fatal(err)
	}
	f.entryVersion = at.UTC().Format("2006-01-02T15:04:05.000000Z07:00") // entries' version format
	f.deleted = zilnic("2026-11-16", "2026-11-16T08:00:00+02:00", "M")
	deletedAt := "2026-11-18T09:00:00+02:00"
	s.exec(t, `update bird_event set deleted_at = $2 where entry_id = $1`, f.deleted, deletedAt)
	s.exec(t, `update entry set deleted_at = $2, deleted_by = $3 where id = $1`, f.deleted, deletedAt, author)
	s.exec(t, `insert into audit_log (at, user_id, action, entity, entity_id, entry_id, entry_kind, farm_ids, record_date, same_day_author)
		values ($1, $2, 'delete', 'entry', $3, $3, 'daily', array[$4::uuid], '2026-11-16', true)`, deletedAt, author, f.deleted, f.hirb)

	f.telEntry = id(`insert into entry (kind, farm_id, date, created_at) values ('delivery', $1, '2026-11-10', '2026-11-10T10:30:00+02:00')
		returning id`, f.tel)
	// Priced in Prețuri (M5 SEC-1): its entry read, which the viewer may
	// make, still carries no price (entries.TestEntryReadPricedDeliveryHasNoPrice).
	s.exec(t, `insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg, price_per_kg, total_lei, price_input,
		priced_at) values ($1, $2, $3, '2026-11-10', 9000, 7.74, 69660, 'per_kg', '2026-11-12T09:00:00+02:00')`,
		f.telEntry, f.tel, f.feedType)
}

// errorCode is the error.code of an error envelope ("" for another body).
func errorCode(t *testing.T, r response) string {
	t.Helper()
	var b struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(r.body), &b); err != nil {
		t.Fatalf("body %q: %v", r.body, err)
	}
	return b.Error.Code
}

// checkAccount fails unless the body names a's user, role and farms.
func checkAccount(t *testing.T, a account, r response) {
	t.Helper()
	var b accountBody
	if err := json.Unmarshal([]byte(r.body), &b); err != nil {
		t.Fatalf("body %q: %v", r.body, err)
	}
	if b.User.ID != a.id || b.User.Login != a.login || b.User.Role != a.role {
		t.Errorf("user %+v, want id %s login %s role %s", b.User, a.id, a.login, a.role)
	}
	var farms []string
	for _, f := range b.Farms {
		farms = append(farms, f.Name)
	}
	if !slices.Equal(farms, a.farms) {
		t.Errorf("farms %q, want %q", farms, a.farms)
	}
	if b.Today != testdb.Today.String() {
		t.Errorf("today %q, want %s", b.Today, testdb.Today)
	}
}

// checkStockFarms fails unless the stock rows of a reports/feed body name
// exactly a's farms, in name order. The matrix farms have no cycle, so each
// has one "fără ciclu deschis" row (M3 D7).
func checkStockFarms(t *testing.T, a account, r response) {
	t.Helper()
	var b struct {
		Stock []struct {
			Farm string `json:"farm"`
		} `json:"stock"`
	}
	if err := json.Unmarshal([]byte(r.body), &b); err != nil {
		t.Fatalf("body %q: %v", r.body, err)
	}
	var farms []string
	for _, s := range b.Stock {
		farms = append(farms, s.Farm)
	}
	if !slices.Equal(farms, a.farms) {
		t.Errorf("stock farms %q, want the caller's %q", farms, a.farms)
	}
}

// Deleting any row from authMatrix makes matrixGaps name its route, against
// the routes the real wiring registers.
func TestAuthMatrixCatchesMissingRow(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	a, err := newApp(config.Config{AppOrigin: testOrigin, AppSecret: testAppSecret}, log, timex.FixedClock(testdb.Now()), &db.Pools{})
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range authMatrix {
		rows := slices.Delete(slices.Clone(authMatrix), i, i+1)
		missing, stale, partial := matrixGaps(a.routes, rows)
		if !slices.Equal(missing, []string{row.route}) || stale != nil || partial != nil {
			t.Errorf("without the row %q: missing %q, stale %q, partial %q; want only %q missing",
				row.route, missing, stale, partial, row.route)
		}
	}

	// A row naming an unregistered route, and a second row for a route (both
	// stale), and a row without every caller.
	extra := append(slices.Clone(authMatrix),
		matrixRow{route: "GET /api/nu-exista", want: everyone(http.StatusOK)},
		matrixRow{route: "GET /healthz", want: map[caller]int{anonymous: 200}})
	_, stale, partial := matrixGaps(a.routes, extra)
	if !slices.Equal(stale, []string{"GET /api/nu-exista", "GET /healthz"}) {
		t.Errorf("stale %q, want the unregistered route and the second healthz row", stale)
	}
	if len(partial) != len(callers)-1 || !strings.HasPrefix(partial[0], "GET /healthz × ") {
		t.Errorf("partial %q, want the healthz row's %d missing callers", partial, len(callers)-1)
	}

	// A case without every caller is partial too, named with its target.
	row := farmRow("GET /api/table", func(uuid.UUID) string { return "/api/table" })
	delete(row.also[0].want, viewer)
	_, _, partial = matrixGaps(a.routes, []matrixRow{row})
	if !slices.Equal(partial, []string{"GET /api/table (Telița) × viewer"}) {
		t.Errorf("partial %q, want the Telița case's viewer", partial)
	}
}
