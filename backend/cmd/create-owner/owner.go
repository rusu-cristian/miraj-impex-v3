package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// maxNameLen is U1's limit, in characters (app_user's check counts characters).
const maxNameLen = 60

// ownerLockKey serialises create-owner runs (pg_advisory_xact_lock), so two at
// once can't both pass the "no active owner" check.
const ownerLockKey int64 = 0x696c5f6f776e6572 // "il_owner"

var (
	errOwnerExists = errors.New("an active owner exists")
	errLoginTaken  = errors.New("login already used")
)

// ownerInput is a validated name and login.
type ownerInput struct {
	Name  string
	Login string
}

// newOwnerInput applies SPEC §6.10 U1 (name required, at most 60 characters,
// trimmed) and U2 (login [a-z0-9.]{3,40}, trimmed and lower-cased as at sign-in).
func newOwnerInput(name, login string) (ownerInput, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxNameLen {
		return ownerInput{}, fmt.Errorf("--name is required, at most %d characters (U1)", maxNameLen)
	}
	login = auth.NormalizeLogin(login)
	if !auth.ValidLogin(login) {
		return ownerInput{}, errors.New("--login must be 3–40 characters of a–z, 0–9 and '.', e.g. ion.rusu (U2)")
	}
	return ownerInput{Name: name, Login: login}, nil
}

// createdOwner is the new account and its one-time password.
type createdOwner struct {
	ID       uuid.UUID
	Password string
}

// auditUser is the new_values snapshot of the user_change audit row (no
// password hash).
type auditUser struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Login  string    `json:"login"`
	Role   auth.Role `json:"role"`
	Active bool      `json:"active"`
}

// createOwner generates a password, and in one transaction checks that no
// active owner exists (unless force), inserts the owner and writes one
// audit_log row user_change with user_id null (system, user answer Q6).
func createOwner(ctx context.Context, store *db.Store, clock timex.Clock, in ownerInput, force bool) (createdOwner, error) {
	password, err := auth.GeneratePassword()
	if err != nil {
		return createdOwner{}, err
	}
	hash, err := auth.NewService(store, clock).HashPassword(ctx, password)
	if err != nil {
		return createdOwner{}, err
	}
	now := clock.Now()
	var id uuid.UUID
	err = store.InTx(ctx, func(tx db.Tx) error {
		if _, err := tx.Exec(ctx, "select pg_advisory_xact_lock($1)", ownerLockKey); err != nil {
			return fmt.Errorf("lock: %w", err)
		}
		if !force {
			n, err := tx.Q.CountActiveOwners(ctx)
			if err != nil {
				return fmt.Errorf("count owners: %w", err)
			}
			if n > 0 {
				return errOwnerExists
			}
		}
		u, err := tx.Q.InsertUser(ctx, dbq.InsertUserParams{
			Name: in.Name, Login: in.Login, PasswordHash: hash, Role: dbq.UserRoleOwner, Now: now,
		})
		if ce, ok := db.AsConstraint(err, db.Unique); ok && ce.Constraint == "app_user_login_uq" {
			return errLoginTaken
		}
		if err != nil {
			return fmt.Errorf("insert user: %w", err)
		}
		id = u.ID
		snapshot, err := json.Marshal(auditUser{ID: u.ID, Name: u.Name, Login: u.Login, Role: auth.Role(u.Role), Active: u.Active})
		if err != nil {
			return err
		}
		if err := tx.Q.InsertAudit(ctx, dbq.InsertAuditParams{
			At: now, Action: dbq.AuditActionUserChange, Entity: "app_user", EntityID: u.ID, NewValues: snapshot,
		}); err != nil {
			return fmt.Errorf("audit: %w", err)
		}
		return nil
	})
	if err != nil {
		return createdOwner{}, err
	}
	return createdOwner{ID: id, Password: password}, nil
}
