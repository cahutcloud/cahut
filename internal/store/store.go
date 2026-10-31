// Package store opens the SQLite databases of a server and of its Cahuts,
// and keeps their schemas up to date. See docs/adr/0004-sqlite-in-pure-go.md
// and docs/adr/0008-several-cahuts-per-server.md.
package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

//go:embed migrations
var migrations embed.FS

// The schemas of the two kinds of database: the migrations of a Cahut's
// database, and those of the server's.
var (
	CahutSchema  = schema("cahut")
	ServerSchema = schema("server")
)

func schema(name string) fs.FS {
	sub, err := fs.Sub(migrations, path.Join("migrations", name))
	if err != nil {
		panic(err) // a constant path, so a programming error
	}
	return sub
}

// DB is a database, the server's or a Cahut's. Writes go through a single
// connection, so they never wait on each other's locks; reads use a pool of
// read-only connections.
type DB struct {
	Write *sql.DB
	Read  *sql.DB
}

// ReadConnections is how many connections read a database at once. A call
// that holds one in its transaction reads through it, never through the pool
// beside it: calls that each waited for a second connection would wait on
// one another.
const ReadConnections = 4

// Open opens the database file at path, creating it if needed, and applies
// the migrations of schema that it lacks.
func Open(ctx context.Context, path string, schema fs.FS) (*DB, error) {
	const common = "_busy_timeout=5000&_foreign_keys=1"
	write, err := sql.Open("sqlite", path+"?"+common+"&_journal_mode=WAL&_synchronous=NORMAL&_txlock=immediate")
	if err != nil {
		return nil, err
	}
	write.SetMaxOpenConns(1)
	if err := migrate(ctx, write, schema); err != nil {
		write.Close()
		return nil, err
	}
	read, err := sql.Open("sqlite", path+"?"+common+"&_query_only=1")
	if err != nil {
		write.Close()
		return nil, err
	}
	read.SetMaxOpenConns(ReadConnections)
	return &DB{Write: write, Read: read}, nil
}

// Close closes both connection pools.
func (db *DB) Close() error {
	return errors.Join(db.Read.Close(), db.Write.Close())
}

// ErrNewer is the error of a database that a newer version of Cahut made.
var ErrNewer = errors.New("the database was made by a newer version of Cahut")

// migrate applies, in order and each in its own transaction, the migration
// files of schema numbered above the database's user_version. It refuses a
// database whose user_version is above every migration of schema.
func migrate(ctx context.Context, db *sql.DB, schema fs.FS) error {
	var current int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	names, err := fs.Glob(schema, "*.sql")
	if err != nil {
		return err
	}
	versions := make([]int, len(names))
	for i, name := range names { // fs.Glob returns names in lexical order
		if versions[i], err = strconv.Atoi(strings.SplitN(name, "_", 2)[0]); err != nil {
			return fmt.Errorf("migration %s: name must start with its number", name)
		}
	}
	// A binary put back after an update would run on a schema it does not
	// know. See docs/adr/0024-signed-releases-and-updates.md.
	if last := slices.Max(append(versions, 0)); current > last {
		return fmt.Errorf("%w: its schema is at version %d, and this binary knows versions up to %d; "+
			"install the version of Cahut that made it, or a newer one", ErrNewer, current, last)
	}
	for i, name := range names {
		version := versions[i]
		if version <= current {
			continue
		}
		script, err := fs.ReadFile(schema, name)
		if err != nil {
			return err
		}
		if err := apply(ctx, db, version, string(script)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}
	return nil
}

// keysOff starts a migration that makes again a table that others refer
// to. SQLite asks for the foreign keys to be off while the old table is
// dropped, lest its rows take those of other tables with them: the
// migration runs with them off, and checks them before it commits.
const keysOff = "-- foreign keys off\n"

func apply(ctx context.Context, db *sql.DB, version int, script string) error {
	// The pragma does nothing in a transaction, and db has one connection.
	off := strings.HasPrefix(script, keysOff)
	if off {
		if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
			return err
		}
		defer db.ExecContext(ctx, "PRAGMA foreign_keys = ON") //nolint:errcheck // a later query would fail on a closed connection anyway
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	if _, err := tx.ExecContext(ctx, script); err != nil {
		return err
	}
	if off {
		var table string
		err := tx.QueryRowContext(ctx, "PRAGMA foreign_key_check").Scan(&table, new(any), new(any), new(any))
		if err == nil {
			return fmt.Errorf("a row of %s refers to a row that is gone", table)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
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
