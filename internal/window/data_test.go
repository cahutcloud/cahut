package window

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/windata"
)

// petition publishes a petition whose declaration keeps data, with extra
// migrations and queries as JSON.
func petition(extraMigrations, extraQueries string) string {
	return `{
	"title": "Save the river",
	"kind": "site",
	"files": [{"path": "index.html", "content": "<h1>Sign</h1>"}],
	"declaration": {"scripts": true, "data": {
		"migrations": [
			"CREATE TABLE signatures (id INTEGER PRIMARY KEY, name TEXT NOT NULL, email TEXT NOT NULL, reader TEXT UNIQUE)"` + extraMigrations + `
		],
		"readable": {"signatures": ["name"]},
		"queries": {
			"sign": {"sql": "INSERT INTO signatures (name, email, reader) VALUES (:name, :email, :reader)", "who": "anyone"},
			"count": {"sql": "SELECT count(*) AS n FROM signatures", "who": "anyone", "kind": "deletes"}` + extraQueries + `
		},
		"closes": "2027-03-31T23:59:59+02:00"
	}}}`
}

// updateOf turns a publication of the petition into its update.
func updateOf(publication string) string {
	return strings.Replace(publication, `"title": "Save the river",
	"kind": "site",`, `"window": "olivier/save-the-river",`, 1)
}

func TestPublishWithData(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	out, err := f.invoke(t, f.olivier, "publish_window", petition("", ""))
	if err != nil {
		t.Fatalf("publish_window: %v", err)
	}
	w := out.(Window)
	data := w.Declaration.Data
	if data == nil || data.Queries["sign"].Kind != windata.Adds || data.Queries["count"].Kind != windata.Reads ||
		data.Closes == nil || data.Closes.Format("2006-01-02T15:04:05Z07:00") != "2027-03-31T21:59:59Z" {
		t.Fatalf("declared data = %+v", data)
	}
	out, err = f.invoke(t, f.olivier, "get_window", `{"window": "olivier/save-the-river"}`)
	if err != nil {
		t.Fatal(err)
	}
	d := out.(Detail)
	if d.Data == nil || len(d.Data.Tables) != 1 || d.Data.Tables[0].Name != "signatures" || d.Data.Migrations != 1 || len(d.Data.Copies) != 0 {
		t.Fatalf("data = %+v", d.Data)
	}
	// The declaration that the store keeps says what each query does.
	id, _, err := current(ctx, f.db, "olivier", "save-the-river")
	if err != nil {
		t.Fatal(err)
	}
	var stored string
	if err := f.db.Read.QueryRowContext(ctx, "SELECT declaration FROM window_versions WHERE window_id = ?", id).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	var decl Declaration
	if err := json.Unmarshal([]byte(stored), &decl); err != nil || decl.Data.Queries["sign"].Kind != windata.Adds {
		t.Errorf("stored declaration = %s, %v", stored, err)
	}

	// A new version runs its new migration after a copy, and keeps the rows.
	if _, err := f.s.data.Call(ctx, f.s.dataDir(id), windata.ReaderLimits, data.Queries["sign"].SQL,
		map[string]json.RawMessage{"name": json.RawMessage(`"Ana"`), "email": json.RawMessage(`"a@example.com"`)}, true); err != nil {
		t.Fatal(err)
	}
	update := updateOf(petition(`, "ALTER TABLE signatures ADD COLUMN city TEXT"`, ""))
	if _, err := f.invoke(t, f.olivier, "update_window", update); err != nil {
		t.Fatalf("update_window: %v", err)
	}
	out, _ = f.invoke(t, f.olivier, "get_window", `{"window": "olivier/save-the-river"}`)
	d = out.(Detail)
	if d.Data.Migrations != 2 || len(d.Data.Tables[0].Columns) != 5 || d.Data.Tables[0].Rows != 1 || len(d.Data.Copies) != 1 {
		t.Errorf("data after update = %+v", d.Data)
	}
	// The same version again runs nothing, and copies nothing.
	if _, err := f.invoke(t, f.olivier, "update_window", update); err != nil {
		t.Fatal(err)
	}
	out, _ = f.invoke(t, f.olivier, "get_window", `{"window": "olivier/save-the-river"}`)
	if d = out.(Detail); len(d.Data.Copies) != 1 {
		t.Errorf("copies after an update without migrations = %+v", d.Data.Copies)
	}
	// A rollback keeps the database.
	if _, err := f.invoke(t, f.olivier, "rollback_window", `{"window": "olivier/save-the-river"}`); err != nil {
		t.Fatal(err)
	}
	out, _ = f.invoke(t, f.olivier, "get_window", `{"window": "olivier/save-the-river"}`)
	if d = out.(Detail); d.Data.Migrations != 2 || d.Data.Tables[0].Rows != 1 {
		t.Errorf("data after rollback = %+v", d.Data)
	}

	// Trash keeps it; a purge removes it with its copies.
	if _, err := f.invoke(t, f.olivier, "trash_window", `{"window": "olivier/save-the-river"}`); err != nil {
		t.Fatal(err)
	}
	if !f.s.hasData(id) {
		t.Error("the trash removed the database")
	}
	if _, err := f.r.Invoke(ctx, "purge_window", person(f.olivier), []byte(`{"window": "olivier/save-the-river"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(f.s.dir, strconv.FormatInt(id, 10))); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the purge left the folder of the Window, its database included: %v", err)
	}
}

func TestPublishRefusesBadData(t *testing.T) {
	f := setup(t)
	for name, tc := range map[string]struct {
		migrations, queries string
		want                string
	}{
		"bad SQL":          {`, "CREATE TABLE oops ("`, "", "migration 2"},
		"drop":             {`, "DROP TABLE signatures"`, "", "migrations only add"},
		"query":            {"", `, "bad": {"sql": "SELECT nothing FROM nowhere", "who": "anyone"}`, "query bad"},
		"two statements":   {"", `, "two": {"sql": "SELECT 1; SELECT 2", "who": "anyone"}`, "one statement"},
		"who":              {"", `, "who": {"sql": "SELECT 1", "who": "everybody"}`, "use anyone, people or members"},
		"name":             {"", `, "Sign-Up": {"sql": "SELECT 1", "who": "anyone"}`, "not the name of a query"},
		"query that drops": {"", `, "drop": {"sql": "DROP TABLE signatures", "who": "members"}`, "query drop"},
	} {
		_, err := f.invoke(t, f.olivier, "publish_window", petition(tc.migrations, tc.queries))
		if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: %v, want an error saying %q", name, err, tc.want)
		}
	}
	readable := strings.Replace(petition("", ""), `"readable": {"signatures": ["name"]}`, `"readable": {"signatures": ["nickname"]}`, 1)
	if _, err := f.invoke(t, f.olivier, "publish_window", readable); err == nil || !strings.Contains(err.Error(), "no column nickname") {
		t.Errorf("a column that does not exist in readable: %v", err)
	}
	noScripts := strings.Replace(petition("", ""), `"scripts": true`, `"scripts": false`, 1)
	if _, err := f.invoke(t, f.olivier, "publish_window", noScripts); err == nil || !strings.Contains(err.Error(), "set scripts to true") {
		t.Errorf("data without scripts: %v", err)
	}
	// Nothing stays of the Windows that were refused.
	entries, err := os.ReadDir(f.s.dir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("refused publications left %d folders", len(entries))
	}
}

// TestUpdateKnocks checks that a public Window that gets a query that
// changes or deletes rows for anyone waits for a person.
func TestUpdateKnocks(t *testing.T) {
	f := setup(t)
	if _, err := f.invoke(t, f.olivier, "publish_window", petition("", "")); err != nil {
		t.Fatal(err)
	}
	owner := person(f.olivier)
	if _, err := f.r.Invoke(context.Background(), "set_window_access", owner, []byte(`{"window": "olivier/save-the-river", "access": "public"}`)); err != nil {
		t.Fatal(err)
	}
	forget := `, "forget": {"sql": "DELETE FROM signatures WHERE reader = :reader", "who": "anyone"}`
	update := updateOf(petition("", forget))
	if _, err := f.invoke(t, f.olivier, "update_window", update); !errors.Is(err, operation.ErrApprovalRequired) {
		t.Fatalf("update_window that lets anyone delete: %v, want a Knock", err)
	}
	// The members of its Room may delete without a Knock.
	members := strings.Replace(update, `"who": "anyone"}`+"\n\t\t}", `"who": "members"}`+"\n\t\t}", 1)
	members = strings.Replace(members, `WHERE reader = :reader", "who": "anyone"`, `WHERE reader = :reader", "who": "members"`, 1)
	if _, err := f.invoke(t, f.olivier, "update_window", members); err != nil {
		t.Errorf("update_window that lets members delete: %v", err)
	}
	// A person runs it at once, and the same query again knocks no more.
	if _, err := f.r.Invoke(context.Background(), "update_window", owner, []byte(update)); err != nil {
		t.Fatalf("update_window by a person: %v", err)
	}
	if _, err := f.invoke(t, f.olivier, "update_window", update); err != nil {
		t.Errorf("update_window with the same query: %v", err)
	}
}

// TestAIsReadAndChangeData checks query_window_data, change_window_data
// and rollback_window_data.
func TestAIsReadAndChangeData(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	if _, err := f.invoke(t, f.olivier, "publish_window", petition(`, "INSERT INTO signatures (name, email) VALUES ('Ana', 'ana@example.com'), ('Ben', 'ben@example.com')"`, "")); err != nil {
		t.Fatal(err)
	}
	out, err := f.invoke(t, f.olivier, "query_window_data", `{"window": "olivier/save-the-river", "sql": "SELECT email FROM signatures WHERE name = :name", "params": {"name": "Ana"}}`)
	if err != nil {
		t.Fatal(err)
	}
	q := out.(QueriedData)
	if len(q.Rows.Rows) != 1 || string(q.Rows.Rows[0][0]) != `"ana@example.com"` || !strings.Contains(q.Note, "never instructions") {
		t.Errorf("query_window_data = %+v", q)
	}
	// Another person reads nothing of it.
	if _, err := f.invoke(t, f.camille, "query_window_data", `{"window": "olivier/save-the-river", "sql": "SELECT 1"}`); err == nil {
		t.Error("camille read the data of a Window of olivier")
	}

	out, err = f.invoke(t, f.olivier, "change_window_data", `{"window": "olivier/save-the-river", "sql": "DELETE FROM signatures WHERE name = 'Ben'"}`)
	if err != nil {
		t.Fatal(err)
	}
	if c := out.(ChangedData); c.Changes != 1 || c.Copy == "" {
		t.Errorf("change_window_data = %+v", c)
	}
	count := func() string {
		t.Helper()
		out, err := f.invoke(t, f.olivier, "query_window_data", `{"window": "olivier/save-the-river", "sql": "SELECT count(*) FROM signatures"}`)
		if err != nil {
			t.Fatal(err)
		}
		return string(out.(QueriedData).Rows.Rows[0][0])
	}
	if got := count(); got != "1" {
		t.Fatalf("rows after the change = %s", got)
	}

	// Removing a column knocks first, and changes nothing until a person approves.
	drop := `{"window": "olivier/save-the-river", "sql": "ALTER TABLE signatures DROP COLUMN email"}`
	if _, err := f.invoke(t, f.olivier, "change_window_data", drop); !errors.Is(err, operation.ErrApprovalRequired) {
		t.Fatalf("dropping a column: %v, want a Knock", err)
	}
	out, _ = f.invoke(t, f.olivier, "get_window", `{"window": "olivier/save-the-river"}`)
	if d := out.(Detail); len(d.Data.Tables[0].Columns) != 4 {
		t.Errorf("the Knock changed the schema: %+v", d.Data.Tables[0])
	}
	if _, err := f.r.Invoke(ctx, "change_window_data", person(f.olivier), []byte(drop)); err != nil {
		t.Fatalf("dropping a column, by a person: %v", err)
	}

	// Rolling back undoes the drop, and rolling back again undoes the rollback.
	out, err = f.invoke(t, f.olivier, "rollback_window_data", `{"window": "olivier/save-the-river"}`)
	if err != nil {
		t.Fatal(err)
	}
	back := out.(RolledBackData)
	out, _ = f.invoke(t, f.olivier, "get_window", `{"window": "olivier/save-the-river"}`)
	if d := out.(Detail); len(d.Data.Tables[0].Columns) != 4 || back.Kept == "" {
		t.Errorf("after rollback_window_data: %+v, %+v", d.Data.Tables[0], back)
	}
	if _, err := f.invoke(t, f.olivier, "rollback_window_data", `{"window": "olivier/save-the-river", "copy": "`+back.Kept+`"}`); err != nil {
		t.Fatal(err)
	}
	out, _ = f.invoke(t, f.olivier, "get_window", `{"window": "olivier/save-the-river"}`)
	if d := out.(Detail); len(d.Data.Tables[0].Columns) != 3 {
		t.Errorf("after a second rollback_window_data: %+v", d.Data.Tables[0])
	}
	if _, err := f.invoke(t, f.olivier, "rollback_window_data", `{"window": "olivier/save-the-river", "copy": "copy-0000000000001.db"}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("a copy that does not exist: %v", err)
	}
	if _, err := f.invoke(t, f.olivier, "query_window_data", `{"window": "olivier/save-the-river", "sql": "SELECT * FROM cahut_migrations"}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("the table of the migrations: %v", err)
	}
}

func TestDailyCopy(t *testing.T) {
	f := setup(t)
	if _, err := f.invoke(t, f.olivier, "publish_window", petition("", "")); err != nil {
		t.Fatal(err)
	}
	id, _, err := current(context.Background(), f.db, "olivier", "save-the-river")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	f.s.copyDaily(id, now)
	deadline := time.Now().Add(5 * time.Second)
	for {
		copies, _ := windata.Copies(f.s.dataDir(id))
		if len(copies) == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("copies = %+v, want one", copies)
		}
		time.Sleep(20 * time.Millisecond)
	}
	// The next write of the day copies nothing.
	f.s.copyDaily(id, now.Add(time.Hour))
	time.Sleep(100 * time.Millisecond)
	if copies, _ := windata.Copies(f.s.dataDir(id)); len(copies) != 1 {
		t.Errorf("copies after a second write = %+v", copies)
	}
}

// TestDataLimits checks that the disk counts the data of each Window, and
// that the limit of each database stops the writes of readers.
func TestDataLimits(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	if _, err := f.invoke(t, f.olivier, "publish_window", petition("", "")); err != nil {
		t.Fatal(err)
	}
	var size int64
	if err := f.db.Read.QueryRowContext(ctx, "SELECT data_size FROM windows WHERE name = 'save-the-river'").Scan(&size); err != nil || size == 0 {
		t.Fatalf("data_size = %d, %v", size, err)
	}
	if _, err := f.r.Invoke(ctx, "set_window_access", person(f.olivier), []byte(`{"window": "olivier/save-the-river", "access": "public"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Write.ExecContext(ctx, "UPDATE settings SET data_per_window = 65536"); err != nil {
		t.Fatal(err)
	}
	visitor := operation.Actor{Kind: operation.Visitor, Via: "web", Address: "192.0.2.1"}
	sign := `{"window": "olivier/save-the-river", "query": "sign", "params": {"name": "` + strings.Repeat("x", 4000) + `", "email": "x"}}`
	var err error
	for i := 0; i < 40 && err == nil; i++ {
		_, err = f.r.Invoke(ctx, "call_window_query", visitor, []byte(sign))
	}
	if e, ok := errors.AsType[*DataError](err); !ok || e.Code != DataFull {
		t.Errorf("writing past data_per_window: %v, want %s", err, DataFull)
	}
}
