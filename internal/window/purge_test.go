package window

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
)

// person returns the person whom the AI a acts for, in the console.
func person(a operation.Actor) operation.Actor {
	a.Kind, a.Via, a.TokenID, a.Client, a.Session = operation.Person, "api", 0, "", 1
	return a
}

func (f *fixture) count(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	if err := f.db.Read.QueryRowContext(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func (f *fixture) files(id int64) bool {
	_, err := os.Stat(filepath.Join(f.s.dir, strconv.FormatInt(id, 10)))
	return err == nil
}

func TestPurge(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	out, err := f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil {
		t.Fatal(err)
	}
	w := out.(Window)
	ref := `{"window": "` + w.URL + `"}`
	if _, err := f.invoke(t, f.olivier, "update_window", `{"window": "`+w.URL+`", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "v2"}]}`); err != nil {
		t.Fatal(err)
	}
	id, _, err := current(ctx, f.db, "olivier", w.Name)
	if err != nil {
		t.Fatal(err)
	}
	olivier := person(f.olivier)

	if _, err := f.invoke(t, olivier, "purge_window", ref); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("purge_window out of the trash: err = %v, want ErrInvalid", err)
	}
	if _, err := f.invoke(t, f.olivier, "trash_window", ref); err != nil {
		t.Fatal(err)
	}
	// An AI knocks first: nothing goes.
	if _, err := f.invoke(t, f.olivier, "purge_window", ref); !errors.Is(err, operation.ErrApprovalRequired) || !f.files(id) {
		t.Errorf("purge_window by an AI: err = %v, files kept = %v", err, f.files(id))
	}
	if _, err := f.invoke(t, person(f.camille), "purge_window", ref); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("purge_window by another person: err = %v, want ErrNotFound", err)
	}

	out, err = f.invoke(t, olivier, "purge_window", ref)
	if err != nil || out.(Window).Name != w.Name {
		t.Fatalf("purge_window = %+v, %v", out, err)
	}
	if n := f.count(t, "SELECT count(*) FROM windows WHERE id = ?", id) + f.count(t, "SELECT count(*) FROM window_versions WHERE window_id = ?", id); n != 0 || f.files(id) {
		t.Errorf("after the purge: %d rows, files kept = %v", n, f.files(id))
	}
	entries, err := logbook.List(ctx, f.db.Read, 1)
	if err != nil || entries[0].Operation != "purge_window" || entries[0].Target != "olivier/"+w.Name ||
		entries[0].Detail != `deck "Revue du 3e trimestre", versions 2 and 1` {
		t.Errorf("Logbook = %+v, %v", entries, err)
	}

	// Its name is free, and its id is not given again.
	out, err = f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil || out.(Window).Name != w.Name {
		t.Fatalf("publish_window after the purge = %+v, %v", out, err)
	}
	if again, _, err := current(ctx, f.db, "olivier", w.Name); err != nil || again <= id {
		t.Errorf("the new Window has id %d, after %d: %v", again, id, err)
	}
}

func TestDeleteRoom(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	olivier, camille := person(f.olivier), person(f.camille)
	if _, err := f.invoke(t, olivier, "create_room", `{"title": "Marketing"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := f.invoke(t, olivier, "add_room_member", `{"room": "marketing", "user": "camille"}`); err != nil {
		t.Fatal(err)
	}
	var ids []int64
	for _, title := range []string{"Q3", "Q4"} {
		out, err := f.invoke(t, f.camille, "publish_window", `{"title": "`+title+`", "kind": "deck", "room": "marketing",
			"declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "<h1>`+title+`</h1>"}]}`)
		if err != nil {
			t.Fatal(err)
		}
		id, _, err := current(ctx, f.db, "marketing", out.(Window).Name)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if _, err := f.invoke(t, f.camille, "trash_window", `{"window": "marketing/q4"}`); err != nil {
		t.Fatal(err)
	}
	var room int64
	if err := f.db.Read.QueryRowContext(ctx, "SELECT id FROM rooms WHERE name = 'marketing'").Scan(&room); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name  string
		actor operation.Actor
		room  string
		want  error
	}{
		{name: "a personal Room", actor: olivier, room: "olivier", want: operation.ErrInvalid},
		{name: "by a member", actor: camille, room: "marketing", want: operation.ErrForbidden},
		{name: "by an AI", actor: f.olivier, room: "marketing", want: operation.ErrApprovalRequired},
		{name: "no such Room", actor: olivier, room: "sales", want: operation.ErrNotFound},
	} {
		if _, err := f.invoke(t, tt.actor, "delete_room", `{"room": "`+tt.room+`"}`); !errors.Is(err, tt.want) {
			t.Errorf("%s: err = %v, want %v", tt.name, err, tt.want)
		}
	}
	if f.count(t, "SELECT count(*) FROM windows WHERE room_id = ?", room) != 2 || !f.files(ids[0]) {
		t.Fatal("a refused delete_room deleted something")
	}

	out, err := f.invoke(t, olivier, "delete_room", `{"room": "marketing"}`)
	if d := out.(DeletedRoom); err != nil || d.Room != "marketing" || d.Title != "Marketing" || d.Windows != 2 {
		t.Fatalf("delete_room = %+v, %v", out, err)
	}
	for _, q := range []string{"SELECT count(*) FROM rooms WHERE id = ?", "SELECT count(*) FROM room_members WHERE room_id = ?",
		"SELECT count(*) FROM windows WHERE room_id = ?"} {
		if n := f.count(t, q, room); n != 0 {
			t.Errorf("%s = %d after delete_room", q, n)
		}
	}
	for _, id := range ids {
		if f.files(id) || f.count(t, "SELECT count(*) FROM window_versions WHERE window_id = ?", id) != 0 {
			t.Errorf("Window %d stays", id)
		}
	}
	// The Logbook keeps what happened in the Room; the name is free, and
	// the id is not given again.
	if n := f.count(t, "SELECT count(*) FROM logbook WHERE room_id = ?", room); n < 4 {
		t.Errorf("the Logbook keeps %d entries of the Room", n)
	}
	if _, err := f.invoke(t, olivier, "create_room", `{"title": "Marketing"}`); err != nil {
		t.Fatalf("create_room after delete_room: %v", err)
	}
	var again int64
	if err := f.db.Read.QueryRowContext(ctx, "SELECT id FROM rooms WHERE name = 'marketing'").Scan(&again); err != nil || again <= room {
		t.Errorf("the new Room has id %d, after %d: %v", again, room, err)
	}
}
