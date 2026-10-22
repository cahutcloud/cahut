package store

import (
	"context"
	"io/fs"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenMigrates(t *testing.T) {
	ctx := context.Background()
	for _, schema := range []fs.FS{CahutSchema, ServerSchema} {
		testOpenMigrates(ctx, t, schema)
	}
}

func testOpenMigrates(ctx context.Context, t *testing.T, schema fs.FS) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cahut.db")
	migrations, err := fs.Glob(schema, "*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for range 2 { // the second open finds the schema up to date
		db, err := Open(ctx, path, schema)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		var version int
		if err := db.Read.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
			t.Fatal(err)
		}
		if version != len(migrations) {
			t.Errorf("user_version = %d, want %d, the number of the last migration", version, len(migrations))
		}
		var mode string
		if err := db.Read.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&mode); err != nil {
			t.Fatal(err)
		}
		if mode != "wal" {
			t.Errorf("journal_mode = %q, want wal", mode)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReadPoolIsReadOnly(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "cahut.db"), CahutSchema)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Read.ExecContext(ctx, "INSERT INTO rooms (name, kind, title, created_at) VALUES ('x', 'team', 'X', 'now')")
	if err == nil {
		t.Error("insert through the read pool succeeded, want an error")
	}
}

func TestForeignKeys(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "cahut.db"), CahutSchema)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Write.ExecContext(ctx, "INSERT INTO room_members (room_id, user_id, role) VALUES (41, 42, 'owner')")
	if err == nil {
		t.Error("insert with dangling references succeeded, want a foreign key error")
	}
}

func TestTimeRoundTrip(t *testing.T) {
	at := time.Date(2026, 10, 21, 8, 30, 15, 123_000_000, time.FixedZone("CEST", 2*60*60))
	s := FormatTime(at)
	if s != "2026-10-21T06:30:15.123Z" {
		t.Errorf("FormatTime = %q", s)
	}
	back, err := ParseTime(s)
	if err != nil {
		t.Fatal(err)
	}
	if !back.Equal(at) {
		t.Errorf("ParseTime(FormatTime(t)) = %v, want %v", back, at)
	}
}
