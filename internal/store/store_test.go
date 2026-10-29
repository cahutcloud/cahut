package store

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
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

// TestRemadeTablesKeepTheirRows fills a Cahut's database of migration 9,
// then opens it: migration 10 makes rooms and windows again, with
// AUTOINCREMENT, and the foreign keys off. The rows that refer to them
// stay, and a deleted id is not given again.
func TestRemadeTablesKeepTheirRows(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "cahut.db")
	db, err := Open(ctx, path, schemaBefore(t, "0010"))
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		"INSERT INTO users (id, handle, name, role, created_at) VALUES (1, 'olivier', 'Olivier', 'admin', 'now')",
		"INSERT INTO rooms (id, name, kind, title, created_at) VALUES (1, 'olivier', 'personal', 'Olivier', 'now')",
		"INSERT INTO rooms (id, name, kind, title, created_at) VALUES (2, 'marketing', 'team', 'Marketing', 'now')",
		"INSERT INTO room_members (room_id, user_id, role) VALUES (1, 1, 'owner'), (2, 1, 'owner')",
		`INSERT INTO windows (id, room_id, name, title, kind, version, size, file_count, created_by, created_at, updated_at, access)
			VALUES (1, 2, 'q3', 'Q3', 'deck', 1, 10, 1, 1, 'now', 'now', 'key')`,
		"INSERT INTO window_versions (window_id, number, size, file_count, published_by, published_at) VALUES (1, 1, 10, 1, 1, 'now')",
		"INSERT INTO window_passes (window_id, hash, created_at, expires_at) VALUES (1, x'01', 'now', 'later')",
	} {
		if _, err := db.Write.ExecContext(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db, err = Open(ctx, path, CahutSchema)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	for q, want := range map[string]int{
		"SELECT count(*) FROM room_members":                                                    2,
		"SELECT count(*) FROM window_versions":                                                 1,
		"SELECT count(*) FROM window_passes":                                                   1,
		"SELECT count(*) FROM windows WHERE access = 'key'":                                    1,
		"SELECT count(*) FROM rooms WHERE name = 'marketing'":                                  1,
		"SELECT count(*) FROM sqlite_sequence WHERE seq >= 1 AND name IN ('rooms', 'windows')": 2,
		// Migration 13 makes sessions and passes again, the same way.
		"SELECT count(*) FROM sqlite_sequence WHERE seq >= 1 AND name = 'window_passes'":    1,
		"SELECT count(*) FROM pragma_table_info('sessions') WHERE name = 'opened_by'":       1,
		"SELECT count(*) FROM pragma_foreign_key_list('windows') WHERE \"table\" = 'rooms'": 1,
	} {
		var got int
		if err := db.Read.QueryRowContext(ctx, q).Scan(&got); err != nil || got != want {
			t.Errorf("%s = %d, %v; want %d", q, got, err, want)
		}
	}

	// The foreign keys hold again, and delete what refers to a Room.
	if _, err := db.Write.ExecContext(ctx, "DELETE FROM windows WHERE id = 1"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Write.ExecContext(ctx, "DELETE FROM rooms WHERE id = 2"); err != nil {
		t.Fatal(err)
	}
	var members int
	if err := db.Read.QueryRowContext(ctx, "SELECT count(*) FROM room_members").Scan(&members); err != nil || members != 1 {
		t.Errorf("room_members = %d, %v, want 1", members, err)
	}
	var id int64
	err = db.Write.QueryRowContext(ctx, "INSERT INTO rooms (name, kind, title, created_at) VALUES ('sales', 'team', 'Sales', 'now') RETURNING id").Scan(&id)
	if err != nil || id != 3 {
		t.Errorf("a new Room takes id %d, %v, want 3", id, err)
	}
}

// schemaBefore returns the migrations of a Cahut whose numbers come before
// number.
func schemaBefore(t *testing.T, number string) fstest.MapFS {
	t.Helper()
	before := fstest.MapFS{}
	names, err := fs.Glob(CahutSchema, "*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if name < number {
			data, err := fs.ReadFile(CahutSchema, name)
			if err != nil {
				t.Fatal(err)
			}
			before[name] = &fstest.MapFile{Data: data}
		}
	}
	return before
}

func TestWindowsGetTheirIDs(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "cahut.db")
	db, err := Open(ctx, path, schemaBefore(t, "0026"))
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		"INSERT INTO users (id, handle, name, role, created_at) VALUES (1, 'olivier', 'Olivier', 'admin', 'now')",
		"INSERT INTO rooms (id, name, kind, title, created_at) VALUES (1, 'olivier', 'personal', 'Olivier', 'now')",
		`INSERT INTO windows (room_id, name, title, kind, version, size, file_count, created_by, created_at, updated_at)
			VALUES (1, 'q3', 'Q3', 'deck', 1, 10, 1, 1, 'now', 'now'), (1, 'q4', 'Q4', 'deck', 1, 10, 1, 1, 'now', 'now'),
			(1, 'q1', 'Q1', 'deck', 1, 10, 1, 1, 'now', 'now')`,
	} {
		if _, err := db.Write.ExecContext(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db, err = Open(ctx, path, CahutSchema)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	rows, err := db.Read.QueryContext(ctx, "SELECT uid FROM windows")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	// As address.ValidID has it: 25 characters of 5 bits, then one of 3.
	anID := regexp.MustCompile(`^[0-9a-hjkmnp-tv-z]{25}[048cgmrw]$`)
	seen := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		if !anID.MatchString(id) || seen[id] {
			t.Errorf("a Window has the id %q, after %v", id, seen)
		}
		seen[id] = true
	}
	if err := rows.Err(); err != nil || len(seen) != 3 {
		t.Errorf("%d Windows have an id, %v; want 3", len(seen), err)
	}
	if _, err := db.Write.ExecContext(ctx, "UPDATE windows SET uid = (SELECT uid FROM windows WHERE name = 'q3') WHERE name = 'q4'"); err == nil {
		t.Error("two Windows took the same id")
	}
}

func TestNewerDatabaseDoesNotOpen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "cahut.db")
	newer := fstest.MapFS{
		"0001_one.sql": {Data: []byte("CREATE TABLE one (id INTEGER PRIMARY KEY);")},
		"0002_two.sql": {Data: []byte("CREATE TABLE two (id INTEGER PRIMARY KEY);")},
	}
	db, err := Open(ctx, path, newer)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	older := fstest.MapFS{"0001_one.sql": newer["0001_one.sql"]}
	_, err = Open(ctx, path, older)
	if !errors.Is(err, ErrNewer) {
		t.Fatalf("Open with an older schema: err = %v, want ErrNewer", err)
	}
	if want := "schema is at version 2, and this binary knows versions up to 1"; !strings.Contains(err.Error(), want) {
		t.Errorf("error %q does not say %q", err, want)
	}
	if _, err := Open(ctx, path, newer); err != nil {
		t.Errorf("Open with the schema that made it: %v", err)
	}
}
