// Package auth signs users in and out, keeps their sessions, and answers who
// the caller is and what they may see (SPEC §4.3, §6.16, §7.1, §11.4):
// argon2id passwords, login throttling, the il_session cookie, the session
// middleware that puts a Principal in the request context, RequireRole and
// CanSeeFarm.
package auth

import (
	"context"
	"net/http"
	"slices"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
)

// Principal is the signed-in caller.
type Principal struct {
	UserID uuid.UUID
	Name   string
	Role   Role
	// FarmIDs are the farms of a vet, viewer or manager (user_farm; non-nil,
	// possibly empty). nil for the owner, who sees every farm.
	FarmIDs []uuid.UUID

	sessionID uuid.UUID
	login     string
}

// CanSeeFarm reports whether the caller may see farm id: every farm for the
// owner, only the listed farms for everyone else.
func (p Principal) CanSeeFarm(id uuid.UUID) bool {
	if p.Role == Owner {
		return true
	}
	return slices.Contains(p.FarmIDs, id)
}

type ctxKey struct{}

// WithPrincipal returns ctx carrying p.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

// PrincipalFrom is the caller the session middleware put in ctx.
func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(ctxKey{}).(Principal)
	return p, ok
}

// RequireRole lets a request through only when the caller has one of roles:
// 401 UNAUTHENTICATED without a principal (the session middleware didn't
// run), 403 FORBIDDEN "Nu ai acces la această pagină." for any other role
// (a route whose screen has its own sentence, e.g. "Nu ai acces la Serii.",
// answers with that one instead).
func RequireRole(roles ...Role) httpx.Middleware {
	allowed := slices.Clone(roles)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := PrincipalFrom(r.Context())
			switch {
			case !ok:
				httpx.WriteError(w, httpx.Unauthenticated())
			case !p.Role.In(allowed...):
				httpx.WriteError(w, httpx.NewError(http.StatusForbidden, httpx.CodeForbidden, messages.PageForbidden()))
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}
