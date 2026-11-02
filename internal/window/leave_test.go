package window

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
)

// publishIn publishes a Window titled title in the Room name, for the
// person the AI actor acts for, with the access given, or the default one
// when it is empty.
func (f *fixture) publishIn(t *testing.T, actor operation.Actor, name, title, access string) {
	t.Helper()
	args := `{"title": "` + title + `", "kind": "deck", "room": "` + name + `", "access": "` + access + `",
		"declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "<h1>` + title + `</h1>"}]}`
	if _, err := f.invoke(t, actor, "publish_window", args); err != nil {
		t.Fatalf("publish_window %s in %s: %v", title, name, err)
	}
}

func TestCheckAccountDeletion(t *testing.T) {
	f := setup(t)
	cli := operation.Actor{Kind: operation.Admin, Via: "cli"}
	olivier, camille := person(f.olivier), person(f.camille)
	for _, step := range []struct {
		actor      operation.Actor
		name, args string
	}{
		// Camille is the only owner and member of marketing, the only owner
		// of launch with Olivier in it, and an owner of design with him.
		{camille, "create_room", `{"title": "Marketing"}`},
		{camille, "create_room", `{"title": "Launch"}`},
		{camille, "add_room_member", `{"room": "launch", "user": "olivier"}`},
		{olivier, "create_room", `{"title": "Design"}`},
		{olivier, "add_room_member", `{"room": "design", "user": "camille", "role": "owner"}`},
	} {
		if _, err := f.invoke(t, step.actor, step.name, step.args); err != nil {
			t.Fatalf("%s: %v", step.name, err)
		}
	}
	for _, name := range []string{"camille", "marketing", "launch", "design"} {
		f.publishIn(t, f.camille, name, "Q3", "")
	}
	f.publishIn(t, f.camille, "design", "Draft", "author")
	f.publishIn(t, f.camille, "marketing", "Notes", "author")
	// The trash counts: a Window in it goes with its Room.
	if _, err := f.invoke(t, f.camille, "trash_window", `{"window": "camille/q3"}`); err != nil {
		t.Fatal(err)
	}

	check := func(actor operation.Actor) (AccountDeletion, error) {
		t.Helper()
		out, err := f.invoke(t, actor, "check_account_deletion", `{}`)
		if err != nil {
			return AccountDeletion{}, err
		}
		return out.(AccountDeletion), nil
	}

	got, err := check(camille)
	if err != nil {
		t.Fatalf("check_account_deletion: %v", err)
	}
	want := AccountDeletion{
		Allowed: true,
		Rooms: []LeavingRoom{
			{Room: "camille", Title: "camille", Kind: "personal", Windows: 1},
			{Room: "marketing", Title: "Marketing", Kind: "team", Windows: 2},
		},
		Shared: []LeavingRoom{{Room: "launch", Title: "Launch", Kind: "team", Windows: 1, Members: 1}},
		Left:   []LeavingRoom{{Room: "design", Title: "Design", Kind: "team", Windows: 2}},
		// The draft of design is the only one that stays to go: the draft
		// of marketing goes with its Room, and counts once, as a Window of
		// that Room.
		Drafts: 1, Published: 1, Windows: 4,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("check_account_deletion = %+v\nwant %+v", got, want)
	}

	// Olivier shares design with Camille, another owner, and is a member of
	// launch: both stay without him.
	got, err = check(olivier)
	if err != nil {
		t.Fatalf("check_account_deletion by Olivier: %v", err)
	}
	if len(got.Rooms) != 1 || got.Rooms[0].Room != "olivier" || len(got.Shared) != 0 || len(got.Left) != 2 ||
		got.Left[0].Room != "design" || got.Left[1].Room != "launch" || got.LastAdmin {
		t.Errorf("check_account_deletion by Olivier = %+v", got)
	}

	// An AI, and the administrator on the command line, are not people.
	if _, err := check(f.camille); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("check_account_deletion by an AI: err = %v, want ErrForbidden", err)
	}
	if _, err := check(cli); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("check_account_deletion by the administrator: err = %v, want ErrInvalid", err)
	}

	// The only administrator cannot leave.
	if _, err := f.invoke(t, cli, "set_user_role", `{"user": "olivier", "role": "admin"}`); err != nil {
		t.Fatal(err)
	}
	if got, err := check(olivier); err != nil || !got.LastAdmin || !got.Allowed {
		t.Errorf("check_account_deletion by the only administrator = %+v, %v", got, err)
	}
	if got, err := check(camille); err != nil || got.LastAdmin {
		t.Errorf("check_account_deletion by a member = %+v, %v", got, err)
	}

	// Turned off, it says so and lists nothing, as empty arrays.
	if _, err := f.db.Write.ExecContext(context.Background(), "UPDATE settings SET account_deletion = 0"); err != nil {
		t.Fatal(err)
	}
	got, err = check(olivier)
	if err != nil {
		t.Fatal(err)
	}
	off := AccountDeletion{Rooms: []LeavingRoom{}, Shared: []LeavingRoom{}, Left: []LeavingRoom{}}
	if !reflect.DeepEqual(got, off) {
		t.Errorf("check_account_deletion when it is off = %+v, want %+v", got, off)
	}
}

func TestCheckAccountDeletionInAnIndividualCahut(t *testing.T) {
	f := setup(t)
	if _, err := f.db.Write.ExecContext(context.Background(), "UPDATE settings SET kind = 'individual'"); err != nil {
		t.Fatal(err)
	}
	camille := person(f.camille)
	out, err := f.invoke(t, camille, "create_room", `{"title": "Notes"}`)
	if err != nil {
		t.Fatal(err)
	}
	notes := out.(room.Room).Name // started with her handle
	f.publishIn(t, f.camille, notes, "Plan", "")

	out, err = f.invoke(t, camille, "check_account_deletion", `{}`)
	if err != nil {
		t.Fatalf("check_account_deletion: %v", err)
	}
	got := out.(AccountDeletion)
	want := AccountDeletion{
		Allowed: true,
		Rooms: []LeavingRoom{
			{Room: "camille", Title: "camille", Kind: "personal"},
			{Room: notes, Title: "Notes", Kind: "team", Windows: 1},
		},
		Shared: []LeavingRoom{}, Left: []LeavingRoom{},
		Windows: 1,
	}
	if notes != "camille-notes" || !reflect.DeepEqual(got, want) {
		t.Errorf("check_account_deletion in an individual Cahut = %+v\nwant %+v (Room %q)", got, want, notes)
	}
}

// TestLeaving checks the Rooms that go with a person: the very ones that
// delete_user deletes with them, but for those that others are in.
func TestLeaving(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	camille := person(f.camille)
	for _, title := range []string{"Marketing", "Launch"} {
		if _, err := f.invoke(t, camille, "create_room", `{"title": "`+title+`"}`); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.invoke(t, camille, "add_room_member", `{"room": "launch", "user": "olivier"}`); err != nil {
		t.Fatal(err)
	}
	var id int64
	if err := f.db.Read.QueryRowContext(ctx, "SELECT id FROM users WHERE handle = 'camille'").Scan(&id); err != nil {
		t.Fatal(err)
	}
	leave := func() (AccountDeletion, []int64, *sql.Tx) {
		t.Helper()
		tx, err := f.db.Read.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = tx.Rollback() }) // a read changes nothing
		d, rooms, err := f.s.leaving(ctx, tx, id)
		if err != nil {
			t.Fatal(err)
		}
		return d, rooms, tx
	}
	d, rooms, tx := leave()
	var names []string
	for _, r := range d.Rooms {
		names = append(names, r.Room)
	}
	if !reflect.DeepEqual(names, []string{"camille", "marketing"}) || len(rooms) != 2 {
		t.Fatalf("leaving = %+v, %v", d, rooms)
	}
	// The ids are those of the Rooms, in the same order.
	for i, name := range names {
		var got string
		if err := tx.QueryRowContext(ctx, "SELECT name FROM rooms WHERE id = ?", rooms[i]).Scan(&got); err != nil || got != name {
			t.Errorf("Room %d = %q, %v, want %q", rooms[i], got, err, name)
		}
	}
	// Turned off, no Room goes.
	if _, err := f.db.Write.ExecContext(ctx, "UPDATE settings SET account_deletion = 0"); err != nil {
		t.Fatal(err)
	}
	if d, rooms, _ := leave(); d.Allowed || len(rooms) != 0 || len(d.Rooms) != 0 {
		t.Errorf("leaving when it is off = %+v, %v", d, rooms)
	}
}
