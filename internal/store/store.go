// Package store opens the SQLite database of a home and keeps its schema up
// to date. See docs/adr/0004-sqlite-in-pure-go.md.
package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

//go:embed migrations/*.sql
var migrations embed.FS

// DB is the database of a home. Writes go through a single connection, so
// they never wait on each other's locks; reads use a pool of read-only
// connections.
type DB struct {
	Write *sql.DB
	Read  *sql.DB
}

// Open opens the database file at path, creating it if needed, and applies
// the migrations it lacks.
func Open(ctx context.Context, path string) (*DB, error) {
	const common = "_busy_timeout=5000&_foreign_keys=1"
	write, err := sql.Open("sqlite", path+"?"+common+"&_journal_mode=WAL&_synchronous=NORMAL&_txlock=immediate")
	if err != nil {
		return nil, err
	}
	write.SetMaxOpenConns(1)
	if err := migrate(ctx, write); err != nil {
		write.Close()
		return nil, err
	}
	read, err := sql.Open("sqlite", path+"?"+common+"&_query_only=1")
	if err != nil {
		write.Close()
		return nil, err
	}
	read.SetMaxOpenConns(4)
	return &DB{Write: write, Read: read}, nil
}

// Close closes both connection pools.
func (db *DB) Close() error {
	return errors.Join(db.Read.Close(), db.Write.Close())
}

// migrate applies, in order and each in its own transaction, the migration
// files numbered above the database's user_version.
func migrate(ctx context.Context, db *sql.DB) error {
	var current int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	names, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		return err
	}
	for _, name := range names { // fs.Glob returns names in lexical order
		version, err := strconv.Atoi(strings.SplitN(path.Base(name), "_", 2)[0])
		if err != nil {
			return fmt.Errorf("migration %s: name must start with its number", name)
		}
		if version <= current {
			continue
		}
		script, err := migrations.ReadFile(name)
		if err != nil {
			return err
		}
		if err := apply(ctx, db, version, string(script)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}
	return nil
}

func apply(ctx context.Context, db *sql.DB, version int, script string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	if _, err := tx.ExecContext(ctx, script); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", version)); err != nil {
		return err
	}
	return tx.Commit()
}

// timeLayout sorts in time order as text, which SQLite indexes rely on.
const timeLayout = "2006-01-02T15:04:05.000Z"

// FormatTime formats t in UTC for a TEXT column.
func FormatTime(t time.Time) string {
	return t.UTC().Format(timeLayout)
}

// ParseTime parses a time written by FormatTime.
func ParseTime(s string) (time.Time, error) {
	return time.Parse(timeLayout, s)
}
