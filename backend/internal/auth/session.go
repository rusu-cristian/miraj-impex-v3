package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
)

// Session settings (SPEC §11.4, A1).
const (
	CookieName      = "il_session"
	SessionLifetime = 30 * 24 * time.Hour // sliding
	// SlideEvery: expires_at moves forward at most once per hour; the cookie
	// is re-sent with a fresh Max-Age when it does (user answer Q2).
	SlideEvery = time.Hour
	tokenLen   = 32
)

// newToken returns a cookie token (32 random bytes, base64url without
// padding) and the sha256 the database stores.
func newToken() (token string, hash []byte, err error) {
	b := make([]byte, tokenLen)
	if _, err := rand.Read(b); err != nil {
		return "", nil, fmt.Errorf("auth: session token: %w", err)
	}
	sum := sha256.Sum256(b)
	return base64.RawURLEncoding.EncodeToString(b), sum[:], nil
}

// tokenHash decodes a cookie token and returns its sha256; ok is false when
// the value is not a token we could have issued.
func tokenHash(token string) ([]byte, bool) {
	b, err := base64.RawURLEncoding.Strict().DecodeString(token)
	if err != nil || len(b) != tokenLen {
		return nil, false
	}
	sum := sha256.Sum256(b)
	return sum[:], true
}

// sessionCookie is il_session=<token>; Path=/; Max-Age=2592000; HttpOnly;
// Secure; SameSite=Lax.
func sessionCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name: CookieName, Value: token, Path: "/", MaxAge: int(SessionLifetime / time.Second),
		HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode,
	}
}

// clearCookie removes the cookie (Max-Age=0).
func clearCookie() *http.Cookie {
	return &http.Cookie{
		Name: CookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode,
	}
}

// RequireSession is the session middleware (SPEC §11.4). It looks up the
// il_session cookie among unrevoked, unexpired sessions of active users (one
// query, the user's farms included; a second only when the session slides),
// puts the Principal in the context and names the user on the access-log line.
// Anything else is 401 UNAUTHENTICATED. Once an hour of use it slides
// expires_at to now + 30 days and re-sends the cookie with a fresh Max-Age.
func (s *Service) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(CookieName)
		if err != nil {
			httpx.WriteError(w, httpx.Unauthenticated())
			return
		}
		hash, ok := tokenHash(c.Value)
		if !ok {
			httpx.WriteError(w, httpx.Unauthenticated())
			return
		}
		ctx := r.Context()
		now := s.clock.Now()
		q := s.store.Q()
		row, err := q.GetActiveSession(ctx, dbq.GetActiveSessionParams{TokenHash: hash, Now: now})
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, httpx.Unauthenticated())
			return
		}
		if err != nil {
			httpx.WriteError(w, fmt.Errorf("auth: load session: %w", err))
			return
		}
		if row.LastSeenAt.Before(now.Add(-SlideEvery)) {
			n, err := q.SlideSession(ctx, dbq.SlideSessionParams{
				Now: now, ExpiresAt: now.Add(SessionLifetime), ID: row.SessionID, SeenBefore: now.Add(-SlideEvery),
			})
			if err != nil {
				httpx.WriteError(w, fmt.Errorf("auth: slide session: %w", err))
				return
			}
			if n == 1 {
				http.SetCookie(w, sessionCookie(c.Value))
			}
		}
		role := Role(row.Role)
		p := Principal{
			UserID: row.UserID, Name: row.Name, Role: role, FarmIDs: principalFarms(role, row.FarmIds),
			sessionID: row.SessionID, login: row.Login,
		}
		httpx.SetUserID(w, p.UserID)
		next.ServeHTTP(w, r.WithContext(WithPrincipal(ctx, p)))
	})
}

// principalFarms is nil for the owner, who sees every farm, and the user's
// user_farm farms (GetActiveSession, name_key order; never nil) for others.
func principalFarms(role Role, ids []uuid.UUID) []uuid.UUID {
	if role == Owner {
		return nil
	}
	if ids == nil {
		return []uuid.UUID{}
	}
	return ids
}
