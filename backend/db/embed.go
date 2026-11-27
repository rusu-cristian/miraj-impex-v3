// Package db embeds the goose SQL migrations (SPEC §11.9) so the binaries carry
// their schema and need no files at run time.
package db

import "embed"

// Migrations holds db/migrations/*.sql: 0001_baseline.sql (the former
// schema.sql, never edited) and every later numbered migration.
//
//go:embed migrations/*.sql
var Migrations embed.FS
