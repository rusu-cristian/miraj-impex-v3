package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Login throttling (SPEC §11.4, A1, user answer Q3): 5 failures for a login,
// or 20 from one IP, within 15 minutes, the latest of them less than 5
// minutes ago, make the next attempt wait.
const (
	throttleWindow   = 15 * time.Minute
	throttleLock     = 5 * time.Minute
	maxLoginFailures = 5
	maxIPFailures    = 20
)

// loginPattern is the shape of every login (A2, app_user check constraint).
var loginPattern = regexp.MustCompile(`^[a-z0-9.]{3,40}$`)

// Stored text is capped so an absurd value (the body may be 64 KiB) can't
// overflow the login_attempt btree index (a row over ~2.7 kB is an error),
// and sanitised (storedText).
const (
	maxStoredLogin     = 200
	maxStoredUserAgent = 512
)

// Service is the auth module's service.
type Service struct {
	store *db.Store
	clock timex.Clock
	hash  *Hasher
}

// NewService builds the auth service on the app pool's store.
func NewService(store *db.Store, clock timex.Clock) *Service {
	return &Service{store: store, clock: clock, hash: NewHasher()}
}

// HashPassword returns the argon2id PHC string of password through the
// service's Hasher (user creation and resets), which it shares with login so
// the maxHashers bound covers both.
func (s *Service) HashPassword(ctx context.Context, password string) (string, error) {
	return s.hash.Hash(ctx, password)
}

// User is the signed-in account as /api/auth/me and login show it.
type User struct {
	ID    uuid.UUID
	Name  string
	Login string
	Role  Role
}

// Farm is a farm the user may see.
type Farm struct {
	ID   uuid.UUID
	Name string
	Type FarmType
}

// LoginInput is a sign-in attempt. Login and Password are as typed; IP is
// the client address (zero when unknown).
type LoginInput struct {
	Login     string
	Password  string
	IP        netip.Addr
	UserAgent string
}

// LoginResult is a successful sign-in: the cookie token (sent once, never
// logged), the user, their farms and the business day.
type LoginResult struct {
	Token string
	User  User
	Farms []Farm
	Today timex.Day
}

// NormalizeLogin trims and lower-cases a typed login (user answer Q4).
func NormalizeLogin(login string) string { return strings.ToLower(strings.TrimSpace(login)) }

// ValidLogin reports whether login has the shape every login must have
// (SPEC §6.10 U2: [a-z0-9.]{3,40}). Normalize it first.
func ValidLogin(login string) bool { return loginPattern.MatchString(login) }

// Login signs a user in (SPEC §6.16, §11.4, answers Q3/Q4). Errors are
// envelope errors: 400 VALIDATION for an empty field, 429 RATE_LIMITED, 401
// UNAUTHENTICATED "Login sau parolă greșită." for an unknown login, a login
// that can't exist or a wrong password, and 403 ACCOUNT_DISABLED only after
// a correct password.
//
// The attempt is reserved before the password is checked: in one short
// transaction, under advisory locks on the login and the IP, the failures are
// counted and, unless that answers 429 (nothing recorded, Q3a), the attempt is
// inserted as a failure. argon2 then runs outside the transaction, and only a
// correct password of an active user turns the row into a success. So a burst
// of concurrent guesses can't all pass the throttle before any is recorded.
func (s *Service) Login(ctx context.Context, in LoginInput) (LoginResult, error) {
	login := NormalizeLogin(in.Login)
	if err := httpx.Validate(loginFields{login: login, password: in.Password}); err != nil {
		return LoginResult{}, err
	}
	now := s.clock.Now()
	attemptID, err := s.reserveAttempt(ctx, storedText(login, maxStoredLogin), in.IP, now)
	if err != nil {
		return LoginResult{}, err
	}
	wrong := httpx.NewError(http.StatusUnauthorized, httpx.CodeUnauthenticated, messages.LoginFailed())

	if !ValidLogin(login) {
		return LoginResult{}, wrong // no such account can exist (control characters included); nothing to time
	}
	q := s.store.Q()
	u, err := q.GetUserByLogin(ctx, login)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := s.hash.burn(ctx, in.Password); err != nil {
			return LoginResult{}, err
		}
		return LoginResult{}, wrong
	}
	if err != nil {
		return LoginResult{}, fmt.Errorf("auth: load user: %w", err)
	}
	ok, rehash, err := s.hash.Verify(ctx, in.Password, u.PasswordHash)
	if err != nil {
		return LoginResult{}, fmt.Errorf("auth: verify password of user %s: %w", u.ID, err)
	}
	if !ok {
		return LoginResult{}, wrong
	}
	if !u.Active {
		// The reserved row stays a failure (Q3b).
		return LoginResult{}, httpx.NewError(http.StatusForbidden, httpx.CodeAccountDisabled, messages.AccountDisabled())
	}

	var newHash string
	if rehash {
		if newHash, err = s.hash.Hash(ctx, in.Password); err != nil {
			return LoginResult{}, err
		}
	}
	token, tokenHash, err := newToken()
	if err != nil {
		return LoginResult{}, err
	}
	role := Role(u.Role)
	var farms []Farm
	err = s.store.InTx(ctx, func(tx db.Tx) error {
		if newHash != "" {
			if err := tx.Q.UpdatePasswordHash(ctx, dbq.UpdatePasswordHashParams{PasswordHash: newHash, ID: u.ID}); err != nil {
				return fmt.Errorf("auth: rehash: %w", err)
			}
		}
		if err := tx.Q.MarkLoginAttemptSucceeded(ctx, attemptID); err != nil {
			return fmt.Errorf("auth: mark login attempt: %w", err)
		}
		if err := tx.Q.SetLastLogin(ctx, dbq.SetLastLoginParams{Now: now, ID: u.ID}); err != nil {
			return fmt.Errorf("auth: last login: %w", err)
		}
		if _, err := tx.Q.InsertSession(ctx, dbq.InsertSessionParams{
			UserID: u.ID, TokenHash: tokenHash, Now: now, ExpiresAt: now.Add(SessionLifetime),
			UserAgent: optional(storedText(in.UserAgent, maxStoredUserAgent)), Ip: addrOrNil(in.IP),
		}); err != nil {
			return fmt.Errorf("auth: insert session: %w", err)
		}
		farms, err = listFarms(ctx, tx.Q, u.ID, role)
		return err
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		Token: token,
		User:  User{ID: u.ID, Name: u.Name, Login: u.Login, Role: role},
		Farms: farms,
		Today: timex.DayOf(now),
	}, nil
}

// loginFields is the level-1 check of a sign-in (answer Q4: empty login or
// password is a 400 with an issue on each empty field).
type loginFields struct{ login, password string }

func (f loginFields) Validate(v *httpx.Validator) {
	v.Check(f.login != "", "login", httpx.IssueRequired, messages.LoginRequired())
	v.Check(f.password != "", "password", httpx.IssueRequired, messages.PasswordRequired())
}

// reserveAttempt is the throttle check and the attempt record in one
// transaction, serialised per login and per IP by advisory locks (login
// first, then IP: one order, no deadlock). It returns 429 RATE_LIMITED
// (answer Q3 e/f: retryAfterSeconds counts to the latest failure + 5 min)
// without recording anything, or the id of the new login_attempt row, written
// as a failure.
func (s *Service) reserveAttempt(ctx context.Context, login string, ip netip.Addr, now time.Time) (int64, error) {
	var id int64
	err := s.store.InTx(ctx, func(tx db.Tx) error {
		if err := tx.Q.LockLoginThrottle(ctx, login); err != nil {
			return fmt.Errorf("auth: lock login throttle: %w", err)
		}
		if ip.IsValid() {
			if err := tx.Q.LockIPThrottle(ctx, ip); err != nil {
				return fmt.Errorf("auth: lock IP throttle: %w", err)
			}
		}
		if err := checkThrottle(ctx, tx.Q, login, ip, now); err != nil {
			return err
		}
		var err error
		id, err = tx.Q.InsertLoginAttempt(ctx, dbq.InsertLoginAttemptParams{
			Login: login, Ip: addrOrNil(ip), At: now, Success: false,
		})
		if err != nil {
			return fmt.Errorf("auth: record login attempt: %w", err)
		}
		return nil
	})
	return id, err
}

// checkThrottle returns 429 RATE_LIMITED when the login or the IP is locked.
// Failures counts first: Latest is `since` when there are none.
func checkThrottle(ctx context.Context, q *dbq.Queries, login string, ip netip.Addr, now time.Time) error {
	since := now.Add(-throttleWindow)
	var until time.Time
	lf, err := q.LoginFailuresSince(ctx, dbq.LoginFailuresSinceParams{Since: since, Login: login})
	if err != nil {
		return fmt.Errorf("auth: count login failures: %w", err)
	}
	if lf.Failures >= maxLoginFailures {
		until = lf.Latest.Add(throttleLock)
	}
	if ip.IsValid() {
		ipf, err := q.IPFailuresSince(ctx, dbq.IPFailuresSinceParams{Since: since, Ip: ip})
		if err != nil {
			return fmt.Errorf("auth: count IP failures: %w", err)
		}
		if ipf.Failures >= maxIPFailures {
			if u := ipf.Latest.Add(throttleLock); u.After(until) {
				until = u
			}
		}
	}
	if !until.After(now) {
		return nil
	}
	wait := until.Sub(now)
	secs := int((wait + time.Second - 1) / time.Second) // round up
	return httpx.RateLimited(messages.LoginRateLimited(), secs)
}

// Logout revokes the caller's session.
func (s *Service) Logout(ctx context.Context, p Principal) error {
	if _, err := s.store.Q().RevokeSession(ctx, dbq.RevokeSessionParams{Now: s.clock.Now(), ID: p.sessionID}); err != nil {
		return fmt.Errorf("auth: revoke session: %w", err)
	}
	return nil
}

// Me is the data of GET /api/auth/me.
type Me struct {
	User  User
	Farms []Farm
	Now   time.Time
}

// Me returns the caller's account, farms and the clock's now.
func (s *Service) Me(ctx context.Context, p Principal) (Me, error) {
	farms, err := listFarms(ctx, s.store.Q(), p.UserID, p.Role)
	if err != nil {
		return Me{}, err
	}
	return Me{User: User{ID: p.UserID, Name: p.Name, Login: p.login, Role: p.Role}, Farms: farms, Now: s.clock.Now()}, nil
}

// RevokeUserSessions revokes every open session of a user (password reset,
// deactivation, any change of role or farms, SPEC §4.3). Pass the queries of
// the transaction that makes the change. It returns how many were revoked.
func (s *Service) RevokeUserSessions(ctx context.Context, q *dbq.Queries, userID uuid.UUID) (int64, error) {
	n, err := q.RevokeUserSessions(ctx, dbq.RevokeUserSessionsParams{Now: s.clock.Now(), UserID: userID})
	if err != nil {
		return 0, fmt.Errorf("auth: revoke sessions: %w", err)
	}
	return n, nil
}

// listFarms is every farm for the owner and the user_farm farms for others,
// ordered by name_key (answer Q7).
func listFarms(ctx context.Context, q *dbq.Queries, userID uuid.UUID, role Role) ([]Farm, error) {
	out := make([]Farm, 0)
	if role == Owner {
		rows, err := q.ListFarms(ctx)
		if err != nil {
			return nil, fmt.Errorf("auth: list farms: %w", err)
		}
		for _, f := range rows {
			out = append(out, Farm{ID: f.ID, Name: f.Name, Type: FarmType(f.Type)})
		}
		return out, nil
	}
	rows, err := q.ListUserFarms(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth: list user farms: %w", err)
	}
	for _, f := range rows {
		out = append(out, Farm{ID: f.ID, Name: f.Name, Type: FarmType(f.Type)})
	}
	return out, nil
}

func addrOrNil(a netip.Addr) *netip.Addr {
	if !a.IsValid() {
		return nil
	}
	return &a
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// storedText makes typed text safe to store: invalid UTF-8 and control
// characters (PostgreSQL text refuses U+0000) become U+FFFD, and the result is
// cut to at most n bytes without splitting a character.
func storedText(s string, n int) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return utf8.RuneError
		}
		return r
	}, strings.ToValidUTF8(s, string(utf8.RuneError)))
	if len(s) <= n {
		return s
	}
	s = s[:n]
	for len(s) > 0 && !utf8.ValidString(s) {
		s = s[:len(s)-1]
	}
	return s
}
