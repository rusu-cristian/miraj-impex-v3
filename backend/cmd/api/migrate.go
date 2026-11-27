package main

import (
	"context"
	"fmt"
	"io"
	"path"
	"text/tabwriter"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
)

const migrateUsage = "usage: api migrate up|status"

// runMigrate applies (up) or lists (status) the embedded migrations, using
// MIGRATE_DATABASE_URL (the owner role) and nothing else (SPEC §11.9). goose
// holds a PostgreSQL session advisory lock while it works, so a second
// concurrent run waits and then finds nothing to do.
func runMigrate(ctx context.Context, args []string, lookup config.LookupFunc, stdout, stderr io.Writer) int {
	if len(args) != 1 || (args[0] != "up" && args[0] != "status") {
		fmt.Fprintln(stderr, migrateUsage)
		return exitUsage
	}
	cfg, ok := loadConfig(lookup, config.Migrate, stderr)
	if !ok {
		return exitError
	}
	log := newLogger(cfg, stdout)

	sqlDB, err := db.OpenSQL(cfg.MigrateDatabaseURL, nil)
	if err != nil {
		log.Error("migrate: open database", "err", err)
		return exitError
	}
	p, err := db.NewMigrator(sqlDB) // closes sqlDB with p.Close
	if err != nil {
		_ = sqlDB.Close()
		log.Error("migrate: goose", "err", err)
		return exitError
	}
	defer p.Close()

	if args[0] == "status" {
		statuses, err := p.Status(ctx)
		if err != nil {
			log.Error("migrate status", "err", err)
			return exitError
		}
		tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "VERSION\tSTATE\tAPPLIED AT\tFILE")
		for _, s := range statuses {
			applied := "-"
			if !s.AppliedAt.IsZero() {
				applied = s.AppliedAt.UTC().Format("2006-01-02 15:04:05Z")
			}
			fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", s.Source.Version, s.State, applied, path.Base(s.Source.Path))
		}
		if err := tw.Flush(); err != nil {
			return exitError
		}
		return exitOK
	}

	results, err := p.Up(ctx)
	for _, r := range results {
		log.Info("migration applied", "version", r.Source.Version, "file", path.Base(r.Source.Path),
			"duration_ms", r.Duration.Milliseconds())
	}
	if err != nil {
		log.Error("migrate up", "err", err)
		return exitError
	}
	version, err := p.GetDBVersion(ctx)
	if err != nil {
		log.Error("migrate up: read version", "err", err)
		return exitError
	}
	log.Info("migrate up done", "applied", len(results), "version", version)
	return exitOK
}
