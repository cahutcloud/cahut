package windata

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// The test binary is also the worker of the tests: started with
// CAHUT_TEST_WORKER set to the folder of a database, it runs the worker,
// confinement included. CAHUT_TEST_HANG makes a worker that never
// answers; CAHUT_TEST_PROBE confines the process to CAHUT_TEST_WORKER, then
// tries to read the file it names.
func TestMain(m *testing.M) {
	if dir := os.Getenv("CAHUT_TEST_WORKER"); dir != "" {
		switch {
		case os.Getenv("CAHUT_TEST_HANG") != "":
			_, _ = io.Copy(io.Discard, os.Stdin)
			os.Exit(0)
		case os.Getenv("CAHUT_TEST_PROBE") != "":
			os.Exit(probe(dir, os.Getenv("CAHUT_TEST_PROBE")))
		}
		os.Exit(RunWorker([]string{dir}, os.Stderr))
	}
	os.Exit(m.Run())
}

func testCommand(env ...string) func(string) (*exec.Cmd, error) {
	return func(dir string) (*exec.Cmd, error) {
		cmd := exec.Command(os.Args[0], "-test.run=^$") //nolint:gosec,noctx // the test binary itself, which the pool stops
		cmd.Env = append([]string{"CAHUT_TEST_WORKER=" + dir}, env...)
		return cmd, nil
	}
}

func newPool(t *testing.T, o Options) *Pool {
	t.Helper()
	if o.Command == nil && !o.Inline {
		o.Command = testCommand()
	}
	p := NewPool(o)
	t.Cleanup(p.Close)
	return p
}

// petition is the database of a petition, as an AI would declare it.
var petition = []string{
	`CREATE TABLE signatures (id INTEGER PRIMARY KEY, name TEXT NOT NULL, email TEXT NOT NULL, city TEXT,
		reader TEXT UNIQUE, signed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
	`CREATE VIEW cities AS SELECT city, email FROM signatures`,
	`INSERT INTO signatures (name, email, city, reader) VALUES ('Ana', 'ana@example.com', 'Lyon', 'r1'), ('Ben', 'ben@example.com', 'Nantes', 'r2')`,
}

var readable = map[string][]string{"signatures": {"name", "city", "signed_at"}}

func setup(t *testing.T, p *Pool) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "data")
	if n, err := p.Migrate(context.Background(), dir, AILimits, petition); err != nil || n != len(petition) {
		t.Fatalf("Migrate = %d, %v", n, err)
	}
	return dir
}

func params(kv ...string) map[string]json.RawMessage {
	m := map[string]json.RawMessage{}
	for i := 0; i+1 < len(kv); i += 2 {
		b, _ := json.Marshal(kv[i+1])
		m[kv[i]] = b
	}
	return m
}

func values(r Rows) [][]string {
	out := [][]string{}
	for _, row := range r.Rows {
		var line []string
		for _, v := range row {
			line = append(line, string(v))
		}
		out = append(out, line)
	}
	return out
}

func code(err error) string {
	if err == nil {
		return ""
	}
	return CodeOf(err)
}

// TestPrivateColumns checks that a reader learns nothing of a column that
// the declaration does not make readable, by any way of reading it.
func TestPrivateColumns(t *testing.T) {
	p := newPool(t, Options{})
	dir := setup(t, p)
	ctx := context.Background()
	for _, tc := range []struct {
		sql  string
		want string
	}{
		{"SELECT name, email FROM signatures ORDER BY id", `[["\"Ana\"","null"],["\"Ben\"","null"]]`},
		{"SELECT count(*) FROM signatures WHERE email LIKE 'ana%'", `[["0"]]`},
		{"SELECT name FROM signatures ORDER BY email DESC, id", `[["\"Ana\""],["\"Ben\""]]`},
		{"SELECT a.name FROM signatures a JOIN signatures b ON a.email = b.email", `[]`},
		{"SELECT (SELECT email FROM signatures LIMIT 1)", `[["null"]]`},
		{"SELECT json_object('e', email) FROM signatures WHERE name = 'Ana'", `[["\"{\\\"e\\\":null}\""]]`},
		{"WITH x AS (SELECT email AS e FROM signatures) SELECT count(e) FROM x", `[["0"]]`},
		{"SELECT length(email), typeof(email) FROM signatures WHERE name = 'Ana'", `[["null","\"null\""]]`},
		{"SELECT count(*) FROM signatures", `[["2"]]`},
	} {
		rows, err := p.Select(ctx, dir, ReaderLimits, tc.sql, nil, readable)
		if err != nil {
			t.Errorf("%s: %v", tc.sql, err)
			continue
		}
		if got, _ := json.Marshal(values(rows)); string(got) != tc.want {
			t.Errorf("%s = %s, want %s", tc.sql, got, tc.want)
		}
	}
	for _, sql := range []string{
		"SELECT * FROM cities",
		"SELECT * FROM cahut_migrations",
		"SELECT * FROM sqlite_schema",
		"SELECT * FROM pragma_table_info('signatures')",
		"SELECT load_extension('x')",
		"SELECT 1; SELECT 2",
		"DELETE FROM signatures",
		"PRAGMA table_info(signatures)",
		"ATTACH 'x.db' AS x",
	} {
		if _, err := p.Select(ctx, dir, ReaderLimits, sql, nil, readable); code(err) != CodeInvalid {
			t.Errorf("%s: error %v, want %s", sql, err, CodeInvalid)
		}
	}
	// The view reads the e-mails: listing it does not make them readable.
	if _, err := p.Check(ctx, dir, Check{Migrations: petition, Readable: map[string][]string{"cities": {"city"}}}); code(err) != CodeInvalid {
		t.Errorf("Check with a view in readable: %v, want refused", err)
	}
}

func TestQueriesOfTheDeclaration(t *testing.T) {
	p := newPool(t, Options{})
	dir := setup(t, p)
	ctx := context.Background()
	queries := map[string]string{
		"sign":   "INSERT INTO signatures (name, email, city, reader) VALUES (:name, :email, :city, :reader)",
		"mine":   "SELECT name, email FROM signatures WHERE reader = :reader",
		"rename": "UPDATE signatures SET name = :name WHERE reader = :reader",
		"forget": "DELETE FROM signatures WHERE reader = :reader",
		"count":  "SELECT count(*) AS n FROM signatures",
	}
	checked, err := p.Check(ctx, dir, Check{Migrations: petition, Queries: queries, Readable: readable})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"sign": Adds, "mine": Reads, "rename": Changes, "forget": Deletes, "count": Reads}
	for name, k := range want {
		if checked.Kinds[name] != k {
			t.Errorf("kind of %s = %q, want %q", name, checked.Kinds[name], k)
		}
	}
	if checked.New != 0 {
		t.Errorf("Check found %d new migrations, want 0", checked.New)
	}
	r, err := p.Call(ctx, dir, ReaderLimits, queries["sign"], params("name", "Chloé", "email", "c@example.com", "city", "Brest", "reader", "r3"), true)
	if err != nil || r.Changes != 1 || r.LastID != 3 {
		t.Fatalf("sign = %+v, %v", r, err)
	}
	if _, err := p.Call(ctx, dir, ReaderLimits, queries["sign"], params("name", "Chloé", "email", "c@example.com", "reader", "r3"), true); code(err) != CodeConstraint {
		t.Errorf("signing twice: %v, want %s", err, CodeConstraint)
	}
	r, err = p.Call(ctx, dir, ReaderLimits, queries["mine"], params("reader", "r3"), false)
	if got, _ := json.Marshal(values(r.Rows)); err != nil || string(got) != `[["\"Chloé\"","\"c@example.com\""]]` {
		t.Errorf("mine = %s, %v", got, err)
	}
	// A parameter that the page does not give is NULL.
	if _, err := p.Call(ctx, dir, ReaderLimits, queries["sign"], params("name", "Dan", "reader", "r4"), true); code(err) != CodeConstraint {
		t.Errorf("signing without an e-mail: %v, want %s", err, CodeConstraint)
	}
	// A query that reads does not write, even when it says so.
	if _, err := p.Call(ctx, dir, ReaderLimits, queries["forget"], params("reader", "r3"), false); err == nil {
		t.Error("a query that deletes ran on a connection that reads")
	}
	for _, sql := range []string{"DROP TABLE signatures", "CREATE TABLE x (a)", "SELECT * FROM cahut_migrations", "BEGIN"} {
		if _, err := p.Call(ctx, dir, ReaderLimits, sql, nil, true); code(err) != CodeInvalid {
			t.Errorf("call %s: %v, want %s", sql, err, CodeInvalid)
		}
	}
}

func TestMigrations(t *testing.T) {
	p := newPool(t, Options{})
	dir := setup(t, p)
	ctx := context.Background()
	more := append(slices.Clone(petition), "ALTER TABLE signatures ADD COLUMN country TEXT", "CREATE INDEX by_city ON signatures (city)")
	if n, err := p.Migrate(ctx, dir, AILimits, more); err != nil || n != 2 {
		t.Fatalf("Migrate = %d, %v, want 2", n, err)
	}
	if n, err := p.Migrate(ctx, dir, AILimits, more); err != nil || n != 0 {
		t.Errorf("Migrate again = %d, %v, want 0", n, err)
	}
	// A previous version, with fewer migrations, runs nothing.
	if n, err := p.Migrate(ctx, dir, AILimits, petition[:1]); err != nil || n != 0 {
		t.Errorf("Migrate fewer = %d, %v, want 0", n, err)
	}
	for name, migrations := range map[string][]string{
		"changed":     append([]string{"CREATE TABLE signatures (id INTEGER PRIMARY KEY)"}, more[1:]...),
		"drop column": append(slices.Clone(more), "ALTER TABLE signatures DROP COLUMN country"),
		"rename":      append(slices.Clone(more), "ALTER TABLE signatures RENAME COLUMN city TO town"),
		"drop table":  append(slices.Clone(more), "DROP TABLE signatures"),
		"update":      append(slices.Clone(more), "UPDATE signatures SET city = 'Paris'"),
		"delete":      append(slices.Clone(more), "DELETE FROM signatures"),
		"pragma":      append(slices.Clone(more), "PRAGMA journal_mode = DELETE"),
		"vacuum":      append(slices.Clone(more), "VACUUM"),
		"cahut":       append(slices.Clone(more), "CREATE TABLE cahut_x (a)"),
		"parameter":   append(slices.Clone(more), "INSERT INTO signatures (name, email) VALUES (:a, :b)"),
	} {
		if _, err := p.Migrate(ctx, dir, AILimits, migrations); code(err) != CodeInvalid {
			t.Errorf("%s: %v, want refused", name, err)
		}
		if _, err := p.Check(ctx, dir, Check{Migrations: migrations}); code(err) != CodeInvalid {
			t.Errorf("check %s: %v, want refused", name, err)
		}
	}
	s, err := p.Schema(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.Migrations != 5 || len(s.Tables) != 1 || s.Tables[0].Rows != 2 || len(s.Tables[0].Columns) != 7 ||
		!slices.Equal(s.Views, []string{"cities"}) || !slices.Equal(s.Indexes, []string{"by_city"}) || s.Size == 0 {
		t.Errorf("Schema = %+v", s)
	}
}

func TestExecAndCopies(t *testing.T) {
	p := newPool(t, Options{})
	dir := setup(t, p)
	ctx := context.Background()
	c, err := p.Exec(ctx, dir, AILimits, "DROP VIEW cities; ALTER TABLE signatures DROP COLUMN city", nil, true)
	if err != nil || !slices.Equal(c.Removes, []string{"column signatures.city", "view cities"}) {
		t.Fatalf("dry Exec = %+v, %v", c, err)
	}
	if s, _ := p.Schema(ctx, dir); len(s.Tables[0].Columns) != 6 {
		t.Errorf("a dry run changed the schema: %+v", s)
	}
	first, err := p.Copy(ctx, dir, 3)
	if err != nil {
		t.Fatal(err)
	}
	c, err = p.Exec(ctx, dir, AILimits, "DELETE FROM signatures WHERE name = :name; UPDATE signatures SET city = 'Paris'", params("name", "Ben"), false)
	if err != nil || c.Changes != 2 || c.Statements != 2 || len(c.Removes) != 0 {
		t.Fatalf("Exec = %+v, %v", c, err)
	}
	count := func() string {
		r, err := p.Query(ctx, dir, AILimits, "SELECT group_concat(name || ' ' || city, ', ') FROM signatures", nil)
		if err != nil {
			t.Fatal(err)
		}
		return string(r.Rows[0][0])
	}
	if got := count(); got != `"Ana Paris"` {
		t.Fatalf("after Exec: %s", got)
	}
	back, err := p.Restore(ctx, dir, first.Name, 3)
	if err != nil {
		t.Fatal(err)
	}
	if got := count(); got != `"Ana Lyon, Ben Nantes"` {
		t.Errorf("after Restore: %s", got)
	}
	if _, err := p.Restore(ctx, dir, back.Name, 3); err != nil {
		t.Fatal(err)
	}
	if got := count(); got != `"Ana Paris"` {
		t.Errorf("after a second Restore: %s", got)
	}
	// The migrations travel with the copies.
	if s, _ := p.Schema(ctx, dir); s.Migrations != len(petition) {
		t.Errorf("migrations after Restore = %d", s.Migrations)
	}
	for range 4 {
		if _, err := p.Copy(ctx, dir, 3); err != nil {
			t.Fatal(err)
		}
	}
	if copies, _ := Copies(dir); len(copies) != 3 {
		t.Errorf("copies = %+v, want 3", copies)
	}
	for _, sql := range []string{"ATTACH 'x.db' AS x", "VACUUM", "PRAGMA writable_schema = 1", "DELETE FROM cahut_migrations", "BEGIN; COMMIT"} {
		if _, err := p.Exec(ctx, dir, AILimits, sql, nil, false); code(err) != CodeInvalid {
			t.Errorf("Exec %s: %v, want %s", sql, err, CodeInvalid)
		}
	}
	if _, err := p.Restore(ctx, dir, "../data.db", 3); code(err) != CodeInvalid {
		t.Errorf("Restore of a path: %v", err)
	}
}

// TestHostileSQL checks that the statements that a stranger could send
// end in time and within their memory, and that the worker answers next.
func TestHostileSQL(t *testing.T) {
	p := newPool(t, Options{})
	dir := setup(t, p)
	ctx := context.Background()
	open := map[string][]string{"signatures": {"name", "city", "signed_at", "id"}}
	for _, tc := range []struct {
		sql  string
		code string
	}{
		{"WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n) SELECT count(*) FROM n", CodeTimeout},
		{"SELECT count(*) FROM signatures a, signatures b, signatures c, signatures d, signatures e, signatures f, signatures g, signatures h, signatures i, signatures j, signatures k, signatures l, signatures m, signatures n, signatures o, signatures p, signatures q, signatures r, signatures s, signatures t, signatures u, signatures v, signatures w, signatures x, signatures y, signatures z, signatures aa, signatures ab, signatures ac, signatures ad, signatures ae, signatures af, signatures ag, signatures ah, signatures ai, signatures aj, signatures ak, signatures al", CodeTimeout},
		{"SELECT printf('%.*c', 1000000000, 'x')", CodeInvalid},
		{"SELECT zeroblob(1000000000)", CodeInvalid},
		{"SELECT randomblob(1000000000)", CodeInvalid},
		{"SELECT length(replace(hex(zeroblob(400000)), '0', '00000000'))", CodeInvalid},
		{"SELECT " + strings.Repeat("abs(", 200) + "1" + strings.Repeat(")", 200), CodeInvalid},
		{"SELECT " + strings.Repeat("1 + ", 5000) + "1", CodeInvalid},
		{"SELECT 1" + strings.Repeat(" UNION SELECT 1", 100), CodeInvalid},
		{"SELECT '" + strings.Repeat("x", 200_000) + "'", CodeInvalid},
	} {
		start := time.Now()
		_, err := p.Select(ctx, dir, ReaderLimits, tc.sql, nil, open)
		if code(err) != tc.code {
			t.Errorf("%.60s: %v, want %s", tc.sql, err, tc.code)
		}
		if d := time.Since(start); d > 2*time.Second {
			t.Errorf("%.60s took %s", tc.sql, d)
		}
		if r, err := p.Select(ctx, dir, ReaderLimits, "SELECT count(*) FROM signatures", nil, open); err != nil || string(r.Rows[0][0]) != "2" {
			t.Fatalf("after %.60s: %+v, %v", tc.sql, r, err)
		}
	}
	// Rows beyond the limits are cut, and say so.
	r, err := p.Select(ctx, dir, Limits{Millis: 1000, Rows: 10, Bytes: 1 << 20}, "WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n LIMIT 100) SELECT i FROM n", nil, open)
	if err != nil || len(r.Rows) != 10 || !r.Cut {
		t.Errorf("rows = %d, cut %v, %v", len(r.Rows), r.Cut, err)
	}
	r, err = p.Select(ctx, dir, Limits{Millis: 1000, Rows: 1000, Bytes: 100}, "WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n LIMIT 100) SELECT i FROM n", nil, open)
	if err != nil || len(r.Rows) == 0 || len(r.Rows) >= 100 || !r.Cut {
		t.Errorf("rows within 100 bytes = %d, cut %v, %v", len(r.Rows), r.Cut, err)
	}
}

// TestFull checks that a database stops growing at its size.
func TestFull(t *testing.T) {
	p := newPool(t, Options{})
	dir := setup(t, p)
	lim := ReaderLimits
	lim.MaxSize = 64 << 10
	var err error
	for i := 0; i < 100 && err == nil; i++ {
		_, err = p.Call(context.Background(), dir, lim, "INSERT INTO signatures (name, email) VALUES (:name, 'x')",
			params("name", strings.Repeat("x", 4000)), true)
	}
	if code(err) != CodeFull {
		t.Errorf("filling the database: %v, want %s", err, CodeFull)
	}
}

// TestFunctions checks that every function of SQLite is allowed or
// denied by name, so that a new version of the driver decides nothing
// alone.
func TestFunctions(t *testing.T) {
	p := newPool(t, Options{Inline: true})
	dir := filepath.Join(t.TempDir(), "data")
	ctx := context.Background()
	if _, err := p.Migrate(ctx, dir, AILimits, nil); err != nil {
		t.Fatal(err)
	}
	c, err := p.Exec(ctx, dir, AILimits, "CREATE TABLE f AS SELECT DISTINCT name FROM pragma_function_list", nil, false)
	if err == nil {
		t.Fatalf("pragma_function_list is open to AIs: %+v", c)
	}
	// The worker reads the list for itself, outside any policy.
	d, err := openDatabase(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer d.close()
	cn := <-d.readers
	defer func() { d.readers <- cn }()
	rows, err := cn.c.QueryContext(ctx, "SELECT DISTINCT name FROM pragma_function_list")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		if !functions[name] && !deniedFunctions[name] {
			t.Errorf("the function %s is neither allowed nor denied", name)
		}
	}
	for name := range functions {
		if deniedFunctions[name] {
			t.Errorf("the function %s is both allowed and denied", name)
		}
	}
}

func TestValues(t *testing.T) {
	p := newPool(t, Options{Inline: true})
	dir := filepath.Join(t.TempDir(), "data")
	ctx := context.Background()
	if _, err := p.Migrate(ctx, dir, AILimits, []string{"CREATE TABLE v (a, d DATETIME)", "INSERT INTO v (d) VALUES (CURRENT_TIMESTAMP)"}); err != nil {
		t.Fatal(err)
	}
	in := map[string]json.RawMessage{
		"n": json.RawMessage(`9007199254740993`), "f": json.RawMessage(`1.5`), "t": json.RawMessage(`true`),
		"s": json.RawMessage(`"é"`), "b": json.RawMessage(`{"blob":"AAE="}`), "o": json.RawMessage(`{"a":[1]}`), "z": json.RawMessage(`null`),
	}
	r, err := p.Query(ctx, dir, AILimits, "SELECT :n, typeof(:n), :f, :t, :s, :b, typeof(:b), :o, :z", in)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(r.Rows[0])
	if want := `[9007199254740993,"integer",1.5,1,"é",{"blob":"AAE="},"blob","{\"a\":[1]}",null]`; string(got) != want {
		t.Errorf("values = %s, want %s", got, want)
	}
	r, err = p.Query(ctx, dir, AILimits, "SELECT d, d = CURRENT_TIMESTAMP OR d < CURRENT_TIMESTAMP FROM v", nil)
	if err != nil || len(r.Rows) != 1 || len(string(r.Rows[0][0])) != len(`"2006-01-02 15:04:05"`) || string(r.Rows[0][1]) != "1" {
		t.Errorf("a DATETIME = %s, %v", r.Rows, err)
	}
}

// TestPool checks the life of workers: they start on demand, stop when
// idle, give way to another database, and are stopped when they hang.
func TestPool(t *testing.T) {
	ctx := context.Background()
	t.Run("idle", func(t *testing.T) {
		p := newPool(t, Options{Idle: time.Second})
		setup(t, p)
		deadline := time.Now().Add(10 * time.Second)
		for {
			p.mu.Lock()
			n := len(p.workers)
			p.mu.Unlock()
			if n == 0 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("an idle worker did not stop")
			}
			time.Sleep(100 * time.Millisecond)
		}
	})
	t.Run("one at a time", func(t *testing.T) {
		p := newPool(t, Options{MaxWorkers: 1})
		a, b := setup(t, p), setup(t, p)
		for _, dir := range []string{a, b, a} {
			if _, err := p.Schema(ctx, dir); err != nil {
				t.Fatal(err)
			}
		}
	})
	t.Run("hang", func(t *testing.T) {
		p := newPool(t, Options{Command: testCommand("CAHUT_TEST_HANG=1")})
		dir := t.TempDir()
		start := time.Now()
		_, err := p.Select(ctx, dir, Limits{Millis: 100}, "SELECT 1", nil, nil)
		if code(err) != CodeTimeout || time.Since(start) > grace+5*time.Second {
			t.Errorf("a worker that hangs: %v after %s", err, time.Since(start))
		}
	})
	t.Run("stop", func(t *testing.T) {
		p := newPool(t, Options{})
		dir := setup(t, p)
		p.Stop(dir)
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
		p.Close()
		if _, err := p.Schema(ctx, dir); code(err) != CodeUnavailable {
			t.Errorf("a call after Close: %v", err)
		}
	})
	t.Run("inline", func(t *testing.T) {
		p := newPool(t, Options{Inline: true})
		dir := setup(t, p)
		if r, err := p.Select(ctx, dir, ReaderLimits, "SELECT name FROM signatures ORDER BY id", nil, readable); err != nil || len(r.Rows) != 2 {
			t.Errorf("inline = %+v, %v", r, err)
		}
	})
}

func TestRunWorkerRefuses(t *testing.T) {
	var out strings.Builder
	if RunWorker([]string{"relative"}, &out) == 0 || !strings.Contains(out.String(), "absolute") {
		t.Errorf("RunWorker with a relative folder: %q", out.String())
	}
}

// probe confines the process to dir, then checks that it writes there
// and reads nothing elsewhere. It returns 0 when it is so.
func probe(dir, outside string) int {
	if err := confine(dir); err != nil {
		os.Stderr.WriteString(err.Error())
		return 1
	}
	if err := os.WriteFile(filepath.Join(dir, "inside"), []byte("x"), 0o600); err != nil { //nolint:gosec // a folder of the test
		os.Stderr.WriteString("write inside: " + err.Error())
		return 2
	}
	if _, err := os.ReadFile(outside); !errors.Is(err, os.ErrPermission) { //nolint:gosec // a file of the test
		os.Stderr.WriteString("read outside: " + errString(err))
		return 3
	}
	return 0
}

func errString(err error) string {
	if err == nil {
		return "no error"
	}
	return err.Error()
}

func TestConfinement(t *testing.T) {
	if LandlockABI() < 1 {
		t.Skip("no Landlock on this system")
	}
	if !restrictsThreads {
		t.Skip("built with cgo, for the race detector: make test runs this test again without it")
	}
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(outside, []byte("key"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^$") //nolint:gosec // the test binary itself
	cmd.Env = []string{"CAHUT_TEST_WORKER=" + dir, "CAHUT_TEST_PROBE=" + outside}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("the confined worker: %v: %s", err, out)
	}
}
