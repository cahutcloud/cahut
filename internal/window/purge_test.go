package window

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"cahut.cloud/internal/account"
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

func TestDeleteUser(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	cli := operation.Actor{Kind: operation.Admin, Via: "cli"}
	if _, err := f.invoke(t, cli, "set_user_role", `{"user": "olivier", "role": "admin"}`); err != nil {
		t.Fatal(err)
	}
	olivier, camille := person(f.olivier), person(f.camille)
	// Camille is the only owner of marketing, and an owner of design with
	// Olivier: marketing goes with her, design stays.
	for _, step := range []struct {
		actor      operation.Actor
		name, args string
	}{
		{camille, "create_room", `{"title": "Marketing"}`},
		{camille, "add_room_member", `{"room": "marketing", "user": "olivier"}`},
		{olivier, "create_room", `{"title": "Design"}`},
		{olivier, "add_room_member", `{"room": "design", "user": "camille", "role": "owner"}`},
	} {
		if _, err := f.invoke(t, step.actor, step.name, step.args); err != nil {
			t.Fatalf("%s: %v", step.name, err)
		}
	}
	var gone []int64
	for _, room := range []string{"camille", "marketing", "design"} {
		out, err := f.invoke(t, f.camille, "publish_window", `{"title": "Q3", "kind": "deck", "room": "`+room+`",
			"declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "<h1>Q3</h1>"}]}`)
		if err != nil {
			t.Fatal(err)
		}
		id, _, err := current(ctx, f.db, room, out.(Window).Name)
		if err != nil {
			t.Fatal(err)
		}
		if room != "design" {
			gone = append(gone, id)
		}
	}
	if _, err := f.invoke(t, f.camille, "trash_window", `{"window": "camille/q3"}`); err != nil {
		t.Fatal(err)
	}
	var id int64
	if err := f.db.Read.QueryRowContext(ctx, "SELECT id FROM users WHERE handle = 'camille'").Scan(&id); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name  string
		actor operation.Actor
		user  string
		want  error
	}{
		{name: "themself", actor: olivier, user: "olivier", want: operation.ErrInvalid},
		{name: "the last administrator", actor: cli, user: "olivier", want: operation.ErrInvalid},
		{name: "by a member", actor: camille, user: "olivier", want: operation.ErrForbidden},
		{name: "by an AI", actor: f.olivier, user: "camille", want: operation.ErrApprovalRequired},
		{name: "no such person", actor: olivier, user: "ana", want: operation.ErrNotFound},
	} {
		if _, err := f.invoke(t, tt.actor, "delete_user", `{"user": "`+tt.user+`"}`); !errors.Is(err, tt.want) {
			t.Errorf("%s: err = %v, want %v", tt.name, err, tt.want)
		}
	}
	if f.count(t, "SELECT count(*) FROM users WHERE id = ?", id) != 1 || !f.files(gone[0]) {
		t.Fatal("a refused delete_user deleted something")
	}

	out, err := f.invoke(t, olivier, "delete_user", `{"user": "camille"}`)
	if err != nil {
		t.Fatalf("delete_user: %v", err)
	}
	if d := out.(DeletedUser); d.User != "camille" || len(d.Rooms) != 2 || d.Rooms[0] != "camille" || d.Rooms[1] != "marketing" || d.Windows != 2 {
		t.Fatalf("delete_user = %+v", d)
	}
	for _, q := range []string{"SELECT count(*) FROM users WHERE id = ?", "SELECT count(*) FROM tokens WHERE user_id = ?",
		"SELECT count(*) FROM room_members WHERE user_id = ?"} {
		if n := f.count(t, q, id); n != 0 {
			t.Errorf("%s = %d after delete_user", q, n)
		}
	}
	if n := f.count(t, "SELECT count(*) FROM rooms WHERE name IN ('camille', 'marketing')"); n != 0 {
		t.Errorf("%d Rooms of Camille stay", n)
	}
	for _, w := range gone {
		if f.files(w) || f.count(t, "SELECT count(*) FROM windows WHERE id = ?", w) != 0 {
			t.Errorf("Window %d stays", w)
		}
	}
	// The Window she published in design stays, without an author.
	got, err := f.invoke(t, olivier, "get_window", `{"window": "design/q3"}`)
	if err != nil {
		t.Fatalf("get_window: %v", err)
	}
	if d := got.(Detail); d.CreatedBy != "" || len(d.Versions) != 1 || d.Versions[0].PublishedBy != "" {
		t.Errorf("design/q3 = %+v", d)
	}
	// The Logbook keeps her entries; her handle is free, and her id is not
	// given again.
	if n := f.count(t, "SELECT count(*) FROM logbook WHERE user_id = ?", id); n < 4 {
		t.Errorf("the Logbook keeps %d entries of Camille", n)
	}
	u, err := operation.Do(ctx, f.r, account.CreateUser, cli, account.CreateUserInput{Handle: "camille", Name: "Camille"})
	if err != nil || u.ID <= id {
		t.Errorf("the new Camille has id %d, after %d: %v", u.ID, id, err)
	}
}
