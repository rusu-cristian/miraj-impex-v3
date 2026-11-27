package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
)

// readyTimeout is the budget of each readiness check (SPEC §11.12: the ping
// must answer within 1 s).
const readyTimeout = time.Second

// The readiness check names /readyz reports when they fail.
const (
	checkShutdown   = "shutdown"   // the process is shutting down
	checkDatabase   = "database"   // pool.Ping failed or took over 1 s
	checkMigrations = "migrations" // goose version ≠ newest embedded migration
)

// goose's own query for the database version (pressly/goose/v3
// internal/dialects/postgres.go, GetLatestVersion); il_app may only read the
// table (0002_roles.sql).
const dbVersionSQL = "select coalesce(max(version_id), 0) from goose_db_version"

// database is what the readiness checks need from the app pool.
type database interface {
	Ping(ctx context.Context) error
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// health serves /healthz and /readyz (SPEC §11.12, plan D8: at M0 readiness
// is the pool ping and the migration version; the LISTEN check comes at M11).
type health struct {
	db       database
	latest   int64 // newest embedded migration
	log      *slog.Logger
	draining atomic.Bool

	mu   sync.Mutex
	last string // the failing checks of the previous /readyz ("unchecked" before the first), to log changes only
}

func newHealth(d database, latest int64, log *slog.Logger) *health {
	return &health{db: d, latest: latest, log: log, last: "unchecked"}
}

// drain makes /readyz answer 503 from now on (the first step of shutdown).
func (h *health) drain() { h.draining.Store(true) }

// live is GET /healthz: 200 "ok" while the process serves, without touching
// the database.
func (h *health) live(w http.ResponseWriter, _ *http.Request) {
	writeText(w, http.StatusOK, "ok")
}

// ready is GET /readyz: 200 "ok", or 503 with the failing check names
// separated by commas (e.g. "migrations").
func (h *health) ready(w http.ResponseWriter, r *http.Request) {
	failing := h.check(r.Context())
	if len(failing) == 0 {
		writeText(w, http.StatusOK, "ok")
		return
	}
	writeText(w, http.StatusServiceUnavailable, strings.Join(failing, ","))
}

// check runs the readiness checks and returns the names of the failing ones.
// While shutting down it reports only "shutdown"; when the ping fails the
// version is not read (it would fail for the same reason).
func (h *health) check(ctx context.Context) []string {
	if h.draining.Load() {
		return []string{checkShutdown}
	}
	var failing []string
	var attrs []slog.Attr
	if err := h.ping(ctx); err != nil {
		failing = append(failing, checkDatabase)
		attrs = append(attrs, slog.String("error", err.Error()))
	} else if v, err := h.version(ctx); err != nil {
		failing = append(failing, checkMigrations)
		attrs = append(attrs, slog.String("error", err.Error()))
	} else if v != h.latest {
		failing = append(failing, checkMigrations)
		attrs = append(attrs, slog.Int64("db_version", v), slog.Int64("want_version", h.latest))
	}
	h.logChange(failing, attrs...)
	return failing
}

func (h *health) ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, readyTimeout)
	defer cancel()
	return h.db.Ping(ctx)
}

func (h *health) version(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, readyTimeout)
	defer cancel()
	var v int64
	err := h.db.QueryRow(ctx, dbVersionSQL).Scan(&v)
	return v, err
}

// logChange logs readiness only when the set of failing checks changes, so a
// probe every few seconds against a down database doesn't flood the log.
func (h *health) logChange(failing []string, attrs ...slog.Attr) {
	now := strings.Join(failing, ",")
	h.mu.Lock()
	changed := now != h.last
	h.last = now
	h.mu.Unlock()
	if !changed {
		return
	}
	if now == "" {
		h.log.Info("ready")
		return
	}
	attrs = append(attrs, slog.String("failing", now))
	h.log.LogAttrs(context.Background(), slog.LevelWarn, "not ready", attrs...)
}

func writeText(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}
