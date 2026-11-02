package window

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	netmail "net/mail"
	"reflect"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/mail/mailtest"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/store"
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

// leaver gives Camille the Rooms and the Windows of a person who leaves:
// Rooms that go, one that stops her, one that stays, and drafts.
func (f *fixture) leaver(t *testing.T) {
	t.Helper()
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
}

func TestCheckAccountDeletion(t *testing.T) {
	f := setup(t)
	cli := operation.Actor{Kind: operation.Admin, Via: "cli"}
	olivier, camille := person(f.olivier), person(f.camille)
	f.leaver(t)

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

// mailer gives the Cahut of f the e-mail of its server, and a relay in
// memory as its relay, as the tests of account do. It returns the relay.
// Before it, the Cahut sends no e-mail.
func (f *fixture) mailer(t *testing.T) *mailtest.Relay {
	t.Helper()
	ctx := context.Background()
	svc := mail.New(mail.Config{DataDir: t.TempDir()})
	t.Cleanup(func() { svc.Close(ctx) })
	relay := mailtest.Start(t, mailtest.Options{})
	svc.UseRoots(relay.Roots)
	svc.DialAnywhere()
	_, err := f.db.Write.ExecContext(ctx, `INSERT INTO mail_relay (id, host, port, security, sender, set_at)
		VALUES (1, '127.0.0.1', ?, 'starttls', 'cahut@example.com', ?)`, relay.Port, store.FormatTime(time.Now()))
	if err != nil {
		t.Fatal(err)
	}
	mail.Provide(f.r, &mail.Outbox{Service: svc, DB: f.db, Domain: "example.com", Console: "https://example.com/console/"})
	return relay
}

func TestDeleteAccount(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	cli := operation.Actor{Kind: operation.Admin, Via: "cli"}
	f.leaver(t)
	const password, wrong = "a password of Camille's own", "not the password of Camille"
	hash, err := credential.HashPassword(ctx, password)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Write.ExecContext(ctx, "UPDATE users SET password_hash = ?, email = 'camille@example.org' WHERE handle = 'camille'", hash); err != nil {
		t.Fatal(err)
	}
	var id int64
	if err := f.db.Read.QueryRowContext(ctx, "SELECT id FROM users WHERE handle = 'camille'").Scan(&id); err != nil {
		t.Fatal(err)
	}
	// Camille signs in to the console: this session goes with her account.
	visitor := operation.Actor{Kind: operation.Visitor, Via: "api"}
	if _, err := f.invoke(t, visitor, "sign_in", `{"handle": "camille", "password": "`+password+`"}`); err != nil {
		t.Fatal(err)
	}
	camille := person(f.camille)
	if err := f.db.Read.QueryRowContext(ctx, "SELECT id FROM sessions WHERE user_id = ?", id).Scan(&camille.Session); err != nil {
		t.Fatal(err)
	}

	// The Windows that go: those of her personal Room, of marketing, and of
	// launch once it is hers alone, and her draft in design.
	rows, err := f.db.Read.QueryContext(ctx, `
		SELECT w.id FROM windows w JOIN rooms r ON r.id = w.room_id
		WHERE r.name IN ('camille', 'marketing', 'launch') OR r.name = 'design' AND w.name = 'draft'`)
	if err != nil {
		t.Fatal(err)
	}
	var gone []int64
	for rows.Next() {
		var w int64
		if err := rows.Scan(&w); err != nil {
			t.Fatal(err)
		}
		gone = append(gone, w)
	}
	rows.Close()
	if err := rows.Err(); err != nil || len(gone) != 5 {
		t.Fatalf("the Windows that would go: %v, %v", gone, err)
	}

	deleteAccount := func(actor operation.Actor, password, handle string) (any, error) {
		t.Helper()
		args, err := json.Marshal(map[string]string{"password": password, "handle": handle})
		if err != nil {
			t.Fatal(err)
		}
		return f.invoke(t, actor, "delete_account", string(args))
	}
	refusal := func(err error) *operation.Refusal {
		var r *operation.Refusal
		if errors.As(err, &r) {
			return r
		}
		return &operation.Refusal{}
	}
	kept := func(why string) {
		t.Helper()
		if f.count(t, "SELECT count(*) FROM users WHERE id = ?", id) != 1 {
			t.Fatalf("delete_account %s deleted Camille", why)
		}
		for _, w := range gone {
			if !f.files(w) || f.count(t, "SELECT count(*) FROM windows WHERE id = ?", w) != 1 {
				t.Fatalf("delete_account %s deleted Window %d", why, w)
			}
		}
	}

	if _, err := deleteAccount(f.camille, password, "camille"); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("delete_account by an AI: err = %v, want ErrForbidden", err)
	}
	kept("by an AI")
	if _, err := deleteAccount(camille, wrong, "camille"); !errors.Is(err, credential.ErrWrongCurrentPassword) {
		t.Errorf("delete_account with a wrong password: err = %v, want wrong_current_password", err)
	}
	kept("with a wrong password")
	if _, err := deleteAccount(camille, password, "olivier"); refusal(err).Code != "account_handle_mismatch" || !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("delete_account with the handle of another: err = %v, want account_handle_mismatch", err)
	}
	kept("with the handle of another")
	_, err = deleteAccount(camille, password, "camille")
	if r := refusal(err); r.Code != "account_rooms_shared" || !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "launch") {
		t.Errorf("delete_account with launch shared: err = %v, want account_rooms_shared", err)
	} else if values, _ := json.Marshal(r.Values); string(values) != `{"rooms":"launch"}` {
		t.Errorf("account_rooms_shared gives %s", values)
	}
	kept("with launch shared")

	// launch becomes hers alone; then the administrators turn it off.
	if _, err := f.invoke(t, camille, "remove_room_member", `{"room": "launch", "user": "olivier"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Write.ExecContext(ctx, "UPDATE settings SET account_deletion = 0"); err != nil {
		t.Fatal(err)
	}
	if _, err := deleteAccount(camille, password, "camille"); refusal(err).Code != "account_deletion_off" || !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("delete_account when it is off: err = %v, want account_deletion_off", err)
	}
	kept("when it is off")
	if _, err := f.db.Write.ExecContext(ctx, "UPDATE settings SET account_deletion = 1"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.invoke(t, cli, "set_user_role", `{"user": "camille", "role": "admin"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := deleteAccount(camille, password, "camille"); refusal(err).Code != "cahut_needs_an_admin" {
		t.Errorf("delete_account by the only administrator: err = %v, want cahut_needs_an_admin", err)
	}
	kept("by the only administrator")

	if _, err := f.invoke(t, cli, "set_user_role", `{"user": "olivier", "role": "admin"}`); err != nil {
		t.Fatal(err)
	}
	relay := f.mailer(t)
	out, err := deleteAccount(camille, password, "Camille ")
	if err != nil {
		t.Fatalf("delete_account: %v", err)
	}
	want := DeletedUser{User: "camille", Name: "camille", Rooms: []string{"camille", "launch", "marketing"}, Windows: 5}
	if d := out.(DeletedUser); !reflect.DeepEqual(d, want) {
		t.Fatalf("delete_account = %+v, want %+v", d, want)
	}
	for _, q := range []string{"SELECT count(*) FROM users WHERE id = ?", "SELECT count(*) FROM sessions WHERE user_id = ?",
		"SELECT count(*) FROM tokens WHERE user_id = ?", "SELECT count(*) FROM room_members WHERE user_id = ?"} {
		if n := f.count(t, q, id); n != 0 {
			t.Errorf("%s = %d after delete_account", q, n)
		}
	}
	for _, w := range gone {
		if f.files(w) || f.count(t, "SELECT count(*) FROM windows WHERE id = ?", w) != 0 {
			t.Errorf("Window %d stays", w)
		}
	}
	// The Window she published in design stays, without an author.
	got, err := f.invoke(t, person(f.olivier), "get_window", `{"window": "design/q3"}`)
	if err != nil {
		t.Fatalf("get_window: %v", err)
	}
	if d := got.(Detail); d.CreatedBy != "" || len(d.Versions) != 1 || d.Versions[0].PublishedBy != "" {
		t.Errorf("design/q3 = %+v", d)
	}
	// The Logbook keeps the call under her id, and never a password.
	if n := f.count(t, "SELECT count(*) FROM logbook WHERE operation = 'delete_account' AND outcome = 'ok' AND user_id = ?", id); n != 1 {
		t.Errorf("the Logbook keeps %d deletions of Camille's account", n)
	}
	for _, p := range []string{password, wrong} {
		if n := f.count(t, "SELECT count(*) FROM logbook WHERE detail LIKE ? OR error LIKE ? OR target LIKE ?",
			"%"+p+"%", "%"+p+"%", "%"+p+"%"); n != 0 {
			t.Errorf("the Logbook holds a password in %d entries", n)
		}
	}
	// A message tells her, at her address.
	m := relay.Wait(t, 1)[0]
	parsed, err := netmail.ReadMessage(m.Reader())
	if err != nil {
		t.Fatal(err)
	}
	if m.To[0] != "camille@example.org" || parsed.Header.Get("Subject") != "Your account is deleted" {
		t.Errorf("message to %v: %q", m.To, parsed.Header.Get("Subject"))
	}
}
