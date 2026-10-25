package account

import (
	"context"
	"errors"
	"strings"
	"testing"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// personalTitle returns the title of the personal Room named room.
func (f *fixture) personalTitle(t *testing.T, room string) string {
	t.Helper()
	var title string
	if err := f.db.Read.QueryRowContext(context.Background(), "SELECT title FROM rooms WHERE name = ? AND kind = 'personal'", room).Scan(&title); err != nil {
		t.Fatalf("personal Room %s: %v", room, err)
	}
	return title
}

func TestChangeName(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	me := f.active(t, "olivier", false)

	for _, name := range []string{"  ", strings.Repeat("é", 101)} {
		if _, err := operation.Do(ctx, f.r, ChangeName, me, ChangeNameInput{Name: name}); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("change_name(%q): err = %v, want ErrInvalid", name, err)
		}
	}
	u, err := operation.Do(ctx, f.r, ChangeName, me, ChangeNameInput{Name: "  Olivier Durand "})
	if err != nil {
		t.Fatalf("change_name: %v", err)
	}
	if u.Name != "Olivier Durand" {
		t.Errorf("name = %q, want it trimmed", u.Name)
	}
	a, err := operation.Do(ctx, f.r, GetAccount, me, GetAccountInput{})
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != "Olivier Durand" {
		t.Errorf("get_account: name = %q", a.Name)
	}
	if got := f.personalTitle(t, "olivier"); got != "Olivier Durand" {
		t.Errorf("the personal Room: title = %q, want the new name", got)
	}

	// A personal Room with a title of its own keeps it.
	if _, err := f.db.Write.ExecContext(ctx, "UPDATE rooms SET title = 'Drafts' WHERE name = 'olivier'"); err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Do(ctx, f.r, ChangeName, me, ChangeNameInput{Name: "Olivier"}); err != nil {
		t.Fatalf("change_name: %v", err)
	}
	if got := f.personalTitle(t, "olivier"); got != "Drafts" {
		t.Errorf("the personal Room: title = %q, want the one it was given", got)
	}

	if _, err := operation.Do(ctx, f.r, ChangeName, admin, ChangeNameInput{Name: "Root"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("change_name by the command line: err = %v, want ErrInvalid", err)
	}
	if _, err := operation.Do(ctx, f.r, CreateUser, admin, CreateUserInput{Handle: "camille", Name: strings.Repeat("a", 201)}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("create_user with a name too long: err = %v, want ErrInvalid", err)
	}
}

func TestChangeHandle(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	_, code := f.person(t, "olivier", false)
	s, err := operation.Do(ctx, f.r, ActivateAccount, visitor, ActivateInput{Code: code, Password: password})
	if err != nil {
		t.Fatal(err)
	}
	me := f.signedIn(t, s.Token)
	f.person(t, "camille", false)
	if _, err := operation.Do(ctx, f.r, CreateToken, me, CreateTokenInput{Name: "Claude"}); err != nil {
		t.Fatal(err)
	}

	for _, handle := range []string{"Olivier", "o", "camille", "console"} {
		if _, err := operation.Do(ctx, f.r, ChangeHandle, me, ChangeHandleInput{Handle: handle}); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("change_handle(%q): err = %v, want ErrInvalid", handle, err)
		}
	}

	// A Knock that waits names the addresses of the personal Room.
	now := store.FormatTime(f.now)
	_, err = f.db.Write.ExecContext(ctx, `INSERT INTO knocks (user_id, token_id, client, operation, input, created_at, expires_at)
		SELECT user_id, id, 'Claude', 'trash_window', '{}', ?, ? FROM tokens`, now, store.FormatTime(f.now.Add(1e12)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Do(ctx, f.r, ChangeHandle, me, ChangeHandleInput{Handle: "odurand"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("change_handle with a Knock that waits: err = %v, want ErrInvalid", err)
	}
	if _, err := f.db.Write.ExecContext(ctx, "UPDATE knocks SET state = 'denied', answered_at = ?", now); err != nil {
		t.Fatal(err)
	}

	u, err := operation.Do(ctx, f.r, ChangeHandle, me, ChangeHandleInput{Handle: "odurand"})
	if err != nil {
		t.Fatalf("change_handle: %v", err)
	}
	if u.Handle != "odurand" || u.PersonalRoom != "odurand" {
		t.Errorf("change_handle = %+v, want the new handle and personal Room", u)
	}
	f.personalTitle(t, "odurand")

	// The session goes on, under the new handle; signing in takes it, and
	// the old one is free.
	me = f.signedIn(t, s.Token)
	if me.Handle != "odurand" {
		t.Errorf("the session: handle = %q, want odurand", me.Handle)
	}
	if again := f.signedIn(t, mustSignIn(t, f, "odurand")); again.UserID != me.UserID {
		t.Errorf("signed in as %+v, want the same person", again)
	}
	a, err := operation.Do(ctx, f.r, GetAccount, me, GetAccountInput{})
	if err != nil {
		t.Fatalf("the session that changed the handle: %v", err)
	}
	if a.Handle != "odurand" || len(a.Rooms) != 1 || a.Rooms[0].Room != "odurand" {
		t.Errorf("get_account = %+v, want the new handle and its personal Room", a)
	}
	if _, err := operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: "olivier", Password: password}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("sign_in with the old handle: err = %v, want ErrForbidden", err)
	}
	if _, err := operation.Do(ctx, f.r, CreateUser, admin, CreateUserInput{Handle: "olivier", Name: "Another Olivier"}); err != nil {
		t.Errorf("create_user with the old handle: %v", err)
	}

	if _, err := operation.Do(ctx, f.r, ChangeHandle, admin, ChangeHandleInput{Handle: "root"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("change_handle by the command line: err = %v, want ErrInvalid", err)
	}
}

// In an individual Cahut, the Rooms of a person start with their handle,
// and take the new one.
func TestChangeHandleInAnIndividualCahut(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	me := f.active(t, "olivier", false)
	now := store.FormatTime(f.now)
	for _, q := range []string{
		"UPDATE settings SET kind = 'individual' WHERE id = 1",
		"INSERT INTO rooms (name, kind, title, created_at) VALUES ('olivier-notes', 'team', 'Notes', ?), ('olivier-plans', 'team', 'Plans', ?), ('jo-plans', 'team', 'Plans', ?)",
		"INSERT INTO room_members (room_id, user_id, role) SELECT r.id, u.id, 'owner' FROM rooms r, users u WHERE r.name LIKE 'olivier-%' AND u.handle = 'olivier'",
	} {
		args := []any{}
		for range strings.Count(q, "?") {
			args = append(args, now)
		}
		if _, err := f.db.Write.ExecContext(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := operation.Do(ctx, f.r, ChangeHandle, me, ChangeHandleInput{Handle: "jo"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("change_handle to a handle that another Room starts with: err = %v, want ErrInvalid", err)
	}
	if _, err := operation.Do(ctx, f.r, ChangeHandle, me, ChangeHandleInput{Handle: "odurand"}); err != nil {
		t.Fatalf("change_handle: %v", err)
	}
	rows, err := f.db.Read.QueryContext(ctx, "SELECT name FROM rooms ORDER BY name")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatal(err)
		}
		names = append(names, n)
	}
	if got, want := strings.Join(names, " "), "jo-plans odurand odurand-notes odurand-plans"; got != want {
		t.Errorf("Rooms = %s, want %s", got, want)
	}
}
