package migratehistory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ErrReplaceNeedsOwner is returned by Load with Options.Replace when the
// connection's role can't empty the tables: il_app can't touch audit_log
// (append-only), so --replace runs with the owner DSN (user answer Q4).
var ErrReplaceNeedsOwner = errors.New("migratehistory: --replace needs the owner role (MIGRATE_DATABASE_URL)")

// keptTables survive --replace: the migration history and the D19 settings.
var keptTables = []string{"goose_db_version", "app_setting"}

// replace empties every other table of the public schema in the import
// transaction (user answer Q4: dev only; it also removes users, sessions and
// everything entered since). The table list is read from the catalog, so a
// table added by a later migration is included.
func replace(ctx context.Context, tx pgx.Tx) error {
	rows, err := tx.Query(ctx, `
		select tablename, has_table_privilege(format('%I.%I', schemaname, tablename), 'TRUNCATE')
		from pg_tables
		where schemaname = 'public' and not (tablename = any ($1))
		order by tablename`, keptTables)
	if err != nil {
		return fmt.Errorf("migratehistory: list tables: %w", err)
	}
	type table struct {
		name    string
		canDrop bool
	}
	tables, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (table, error) {
		var t table
		err := r.Scan(&t.name, &t.canDrop)
		return t, err
	})
	if err != nil {
		return fmt.Errorf("migratehistory: list tables: %w", err)
	}
	names := make([]string, 0, len(tables))
	for _, t := range tables {
		if !t.canDrop {
			return ErrReplaceNeedsOwner
		}
		names = append(names, pgx.Identifier{"public", t.name}.Sanitize())
	}
	if len(names) == 0 {
		return nil
	}
	if _, err := tx.Exec(ctx, "truncate table "+strings.Join(names, ", ")+" restart identity"); err != nil {
		return fmt.Errorf("migratehistory: empty the tables: %w", err)
	}
	return nil
}
