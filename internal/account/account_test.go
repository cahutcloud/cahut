package account

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

var admin = operation.Actor{Kind: operation.Admin, Via: "cli"}

const password = "correct horse battery staple"

// fixture is a Cahut's registry with a clock that tests move.
type fixture struct {
	r   *operation.Registry
	db  *store.DB
	now time.Time
}

func setup(t *testing.T) *fixture {
	t.Helper()
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "cahut.db"), store.CahutSchema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	f := &fixture{r: operation.NewRegistry(db), db: db, now: time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)}
	f.r.Now = func() time.Time { return f.now }
	Register(f.r)
	return f
}

// person creates a person and returns their activation code.
func (f *fixture) person(t *testing.T, handle string, isAdmin bool) (User, string) {
	t.Helper()
	u, err := operation.Do(context.Background(), f.r, CreateUser, admin, CreateUserInput{Handle: handle, Name: strings.ToUpper(handle), Admin: isAdmin})
	if err != nil {
		t.Fatalf("create_user %s: %v", handle, err)
	}
	return u.User, u.Activation.Code
}

// active creates a person who has chosen their password, and returns them
// as an actor in the console.
func (f *fixture) active(t *testing.T, handle string, isAdmin bool) operation.Actor {
	t.Helper()
	_, code := f.person(t, handle, isAdmin)
	s, err := operation.Do(context.Background(), f.r, ActivateAccount, visitor, ActivateInput{Code: code, Password: password})
	if err != nil {
		t.Fatalf("activate_account %s: %v", handle, err)
	}
	return f.signedIn(t, s.Token)
}

// signedIn returns the actor of the person who holds a session token.
func (f *fixture) signedIn(t *testing.T, token string) operation.Actor {
	t.Helper()
	o, err := VerifySession(context.Background(), f.db, token, f.now)
	if err != nil {
		t.Fatalf("VerifySession: %v", err)
	}
	return operation.Actor{Kind: operation.Person, Via: "api", UserID: o.UserID, Handle: o.Handle, Session: o.SessionID}
}

func (f *fixture) entries(t *testing.T) []logbook.Entry {
	t.Helper()
	e, err := logbook.List(context.Background(), f.db.Read, 1000)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

var visitor = operation.Actor{Kind: operation.Visitor, Via: "api"}

func TestCreateUser(t *testing.T) {
	f := setup(t)
	ctx := context.Background()

	u, err := operation.Do(ctx, f.r, CreateUser, admin, CreateUserInput{Handle: "olivier", Name: " Olivier Durand ", Admin: true})
	if err != nil {
		t.Fatalf("create_user: %v", err)
	}
	if u.ID == 0 || u.Name != "Olivier Durand" || u.Role != Admin || u.Status != Pending || u.PersonalRoom != "olivier" {
		t.Errorf("user = %+v", u)
	}
	if u.Activation == nil || len(u.Activation.Code) != 52 || !u.Activation.ExpiresAt.Equal(f.now.Add(credential.ActivationTTL)) {
		t.Errorf("activation = %+v", u.Activation)
	}
	var role, room, title string
	err = f.db.Read.QueryRowContext(ctx, `
		SELECT m.role, r.name, r.title FROM room_members m JOIN rooms r ON r.id = m.room_id
		WHERE m.user_id = ? AND r.kind = 'personal'`, u.ID).Scan(&role, &room, &title)
	if err != nil || role != "owner" || room != "olivier" || title != "Olivier Durand" {
		t.Errorf("personal Room: %q %q, role %q, err %v", room, title, role, err)
	}

	if _, err := f.db.Write.ExecContext(ctx, "INSERT INTO rooms (name, kind, title, created_at) VALUES ('marketing', 'team', 'Marketing', 'now')"); err != nil {
		t.Fatal(err)
	}
	for _, in := range []CreateUserInput{
		{Handle: "olivier", Name: "Someone else"},
		{Handle: "marketing", Name: "Taken by a Room"},
		{Handle: "mcp", Name: "Reserved"},
		{Handle: "Olivier", Name: "Capital"},
		{Handle: "x", Name: "Too short"},
		{Handle: "camille-", Name: "Ends with a hyphen"},
		{Handle: "camille", Name: "  "},
	} {
		if _, err := operation.Do(ctx, f.r, CreateUser, admin, in); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("create_user(%+v): err = %v, want ErrInvalid", in, err)
		}
	}

	// A person who is not an administrator cannot create people, and the AI
	// of an administrator must knock first.
	member := f.active(t, "camille", false)
	if _, err := operation.Do(ctx, f.r, CreateUser, member, CreateUserInput{Handle: "bruno", Name: "Bruno"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("create_user by a member: err = %v, want ErrForbidden", err)
	}
	ai := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: u.ID, Handle: "olivier", TokenID: 1, Client: "Claude"}
	if _, err := operation.Do(ctx, f.r, CreateUser, ai, CreateUserInput{Handle: "bruno", Name: "Bruno"}); !errors.Is(err, operation.ErrApprovalRequired) {
		t.Errorf("create_user by an AI: err = %v, want ErrApprovalRequired", err)
	}
}

func TestActivateAndSignIn(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	u, code := f.person(t, "olivier", true)

	if _, err := operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: "olivier", Password: password}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("sign_in before activation: err = %v, want ErrForbidden", err)
	}
	if _, err := operation.Do(ctx, f.r, ActivateAccount, visitor, ActivateInput{Code: code, Password: "too short"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("activate_account with a short password: err = %v, want ErrInvalid", err)
	}
	s, err := operation.Do(ctx, f.r, ActivateAccount, visitor, ActivateInput{Code: code, Password: password})
	if err != nil {
		t.Fatalf("activate_account: %v", err)
	}
	if !strings.HasPrefix(s.Token, SessionPrefix) || s.User.Handle != "olivier" || s.User.Status != Active ||
		!s.ExpiresAt.Equal(f.now.Add(credential.SessionMax)) {
		t.Errorf("session = %+v", s)
	}
	if _, err := operation.Do(ctx, f.r, ActivateAccount, visitor, ActivateInput{Code: code, Password: password}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("a used link: err = %v, want ErrForbidden", err)
	}

	for _, in := range []SignInInput{
		{Handle: "olivier", Password: "wrong horse battery staple"},
		{Handle: "nobody", Password: password},
	} {
		_, err := operation.Do(ctx, f.r, SignIn, visitor, in)
		if !errors.Is(err, operation.ErrForbidden) || !strings.Contains(err.Error(), "wrong handle or password") {
			t.Errorf("sign_in(%q): err = %v", in.Handle, err)
		}
	}
	s2, err := operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: " Olivier ", Password: password})
	if err != nil {
		t.Fatalf("sign_in: %v", err)
	}
	me := f.signedIn(t, s2.Token)
	if me.UserID != u.ID {
		t.Errorf("session of %+v, want %d", me, u.ID)
	}
	if _, err := operation.Do(ctx, f.r, SignOut, me, SignOutInput{}); err != nil {
		t.Fatalf("sign_out: %v", err)
	}
	if _, err := VerifySession(ctx, f.db, s2.Token, f.now); !errors.Is(err, ErrInvalidSession) {
		t.Errorf("a session after sign_out: err = %v", err)
	}
	f.signedIn(t, s.Token) // the first session is still open

	// The Logbook names the person who signed in, but not an unknown
	// handle, and never a password, a code or a token.
	var signIns []string
	for _, e := range f.entries(t) {
		all := e.Target + e.Detail + e.Error
		for _, secret := range []string{password, "wrong horse", code, s.Token, s2.Token, "nobody"} {
			if strings.Contains(all, secret) {
				t.Errorf("the Logbook holds %q: %+v", secret, e)
			}
		}
		if e.Operation == "sign_in" {
			signIns = append(signIns, string(e.Outcome)+" "+e.Actor+" "+e.UserHandle)
		}
	}
	want := []string{"ok person olivier", "refused person ", "refused person olivier", "refused person olivier"}
	if !slices.Equal(signIns, want) {
		t.Errorf("sign_in entries = %q, want %q", signIns, want)
	}
}

func TestSessionsEnd(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	_, code := f.person(t, "olivier", true)
	s, err := operation.Do(ctx, f.r, ActivateAccount, visitor, ActivateInput{Code: code, Password: password})
	if err != nil {
		t.Fatal(err)
	}
	start := f.now
	// Used every 6 days, a session lasts 30 days.
	for days := 6; days < 30; days += 6 {
		if _, err := VerifySession(ctx, f.db, s.Token, start.Add(time.Duration(days)*24*time.Hour)); err != nil {
			t.Fatalf("day %d: %v", days, err)
		}
	}
	if _, err := VerifySession(ctx, f.db, s.Token, start.Add(credential.SessionMax)); !errors.Is(err, ErrInvalidSession) {
		t.Errorf("after 30 days: err = %v", err)
	}

	s2, err := operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: "olivier", Password: password})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySession(ctx, f.db, s2.Token, f.now.Add(credential.SessionIdle)); !errors.Is(err, ErrInvalidSession) {
		t.Errorf("after 7 days without use: err = %v", err)
	}
	for _, bad := range []string{"", "cahut_ses_nope", s2.Token + "x", strings.TrimPrefix(s2.Token, SessionPrefix)} {
		if _, err := VerifySession(ctx, f.db, bad, f.now); !errors.Is(err, ErrInvalidSession) {
			t.Errorf("VerifySession(%q): err = %v", bad, err)
		}
	}
}

func TestSignInLimit(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.active(t, "olivier", true)
	for range credential.MaxRefusals {
		if _, err := operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: "olivier", Password: "wrong"}); !errors.Is(err, operation.ErrForbidden) {
			t.Fatalf("wrong password: err = %v", err)
		}
	}
	_, err := operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: "olivier", Password: password})
	if !errors.Is(err, operation.ErrTooMany) || !operation.Public(err) {
		t.Errorf("sign_in after %d refusals: err = %v, want ErrTooMany", credential.MaxRefusals, err)
	}
	f.now = f.now.Add(credential.RefusalWindow + time.Minute)
	if _, err := operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: "olivier", Password: password}); err != nil {
		t.Errorf("sign_in once the refusals are old: %v", err)
	}
}

// TestSignInLimitByAddress checks that wrong passwords from one address
// make that address wait, not the person, until they come from everywhere;
// and that an unknown handle answers as a known one does.
func TestSignInLimitByAddress(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.active(t, "olivier", true)
	from := func(address string) operation.Actor {
		a := visitor
		a.Address = address
		return a
	}
	signIn := func(address, handle, password string) error {
		_, err := operation.Do(ctx, f.r, SignIn, from(address), SignInInput{Handle: handle, Password: password})
		return err
	}
	for range credential.MaxRefusals {
		if err := signIn("192.0.2.1", "olivier", "wrong"); !errors.Is(err, operation.ErrForbidden) {
			t.Fatalf("wrong password: err = %v", err)
		}
		if err := signIn("192.0.2.1", "nobody", "wrong"); !errors.Is(err, operation.ErrForbidden) {
			t.Fatalf("unknown handle: err = %v", err)
		}
	}
	for _, handle := range []string{"olivier", "nobody"} {
		if err := signIn("192.0.2.1", handle, password); !errors.Is(err, operation.ErrTooMany) {
			t.Errorf("sign_in as %s from the address that guessed wrong: err = %v, want ErrTooMany", handle, err)
		}
	}
	if err := signIn("192.0.2.2", "olivier", password); err != nil {
		t.Errorf("sign_in from another address: %v", err)
	}

	// From many addresses, the handle waits everywhere.
	for i := range credential.MaxHandleRefusals {
		_ = signIn(fmt.Sprintf("198.51.100.%d", i), "olivier", "wrong")
	}
	if err := signIn("192.0.2.3", "olivier", password); !errors.Is(err, operation.ErrTooMany) {
		t.Errorf("sign_in after %d refusals from everywhere: err = %v, want ErrTooMany", credential.MaxHandleRefusals, err)
	}
	if entries, err := logbook.List(ctx, f.db.Read, 1); err != nil || len(entries) != 1 || entries[0].Address != "192.0.2.3" {
		t.Errorf("the Logbook does not keep the address of a sign-in: %+v, %v", entries, err)
	}
}

func TestChangePassword(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	me := f.active(t, "olivier", true)
	other, err := operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: "olivier", Password: password})
	if err != nil {
		t.Fatal(err)
	}
	const newPassword = "a longer and newer passphrase"
	for in, want := range map[ChangePasswordInput]error{
		{Password: "wrong horse battery staple", NewPassword: newPassword}: operation.ErrForbidden,
		{Password: password, NewPassword: "short"}:                         operation.ErrInvalid,
	} {
		if _, err := operation.Do(ctx, f.r, ChangePassword, me, in); !errors.Is(err, want) {
			t.Errorf("change_password(%+v): err = %v, want %v", in, err, want)
		}
	}
	if _, err := operation.Do(ctx, f.r, ChangePassword, me, ChangePasswordInput{Password: password, NewPassword: newPassword}); err != nil {
		t.Fatalf("change_password: %v", err)
	}
	if _, err := VerifySession(ctx, f.db, other.Token, f.now); !errors.Is(err, ErrInvalidSession) {
		t.Errorf("the other session: err = %v, want it ended", err)
	}
	if _, err := operation.Do(ctx, f.r, GetAccount, me, GetAccountInput{}); err != nil {
		t.Errorf("the session that changed the password: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: "olivier", Password: newPassword}); err != nil {
		t.Errorf("sign_in with the new password: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, ChangePassword, admin, ChangePasswordInput{Password: password, NewPassword: newPassword}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("change_password by the command line: err = %v, want ErrInvalid", err)
	}
}

func TestActivationLinks(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	me := f.active(t, "olivier", true)
	_, first := f.person(t, "camille", false)

	// A new link replaces the first.
	a, err := operation.Do(ctx, f.r, CreateActivationLink, me, UserInput{User: "camille"})
	if err != nil {
		t.Fatalf("create_activation_link: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, ActivateAccount, visitor, ActivateInput{Code: first, Password: password}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("a replaced link: err = %v, want ErrForbidden", err)
	}
	f.now = f.now.Add(credential.ActivationTTL)
	if _, err := operation.Do(ctx, f.r, ActivateAccount, visitor, ActivateInput{Code: a.Code, Password: password}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("an expired link: err = %v, want ErrForbidden", err)
	}

	// After an oversight, the old password works until the new link is used,
	// and the link ends every session.
	a, err = operation.Do(ctx, f.r, CreateActivationLink, admin, UserInput{User: "olivier"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: "olivier", Password: password}); err != nil {
		t.Errorf("sign_in with the old password: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, ActivateAccount, visitor, ActivateInput{Code: a.Code, Password: "a brand new passphrase"}); err != nil {
		t.Fatalf("activate_account: %v", err)
	}
	var sessions int
	if err := f.db.Read.QueryRowContext(ctx, "SELECT count(*) FROM sessions WHERE user_id = ?", me.UserID).Scan(&sessions); err != nil || sessions != 1 {
		t.Errorf("sessions of olivier after activation: %d, %v; want only the new one", sessions, err)
	}

	member := f.active(t, "bruno", false)
	if _, err := operation.Do(ctx, f.r, CreateActivationLink, member, UserInput{User: "camille"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("create_activation_link by a member: err = %v, want ErrForbidden", err)
	}
	ai := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: me.UserID, Handle: "olivier", TokenID: 1, Client: "Claude"}
	if _, err := operation.Do(ctx, f.r, CreateActivationLink, ai, UserInput{User: "camille"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("create_activation_link by an AI: err = %v, want ErrForbidden", err)
	}
}

func TestRolesAndDisabling(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	me := f.active(t, "olivier", true)
	camille := f.active(t, "camille", false)

	if _, err := operation.Do(ctx, f.r, SetUserRole, me, SetUserRoleInput{User: "olivier", Role: Member}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("demote the last administrator: err = %v, want ErrInvalid", err)
	}
	if _, err := operation.Do(ctx, f.r, SetUserRole, camille, SetUserRoleInput{User: "camille", Role: Admin}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("set_user_role by a member: err = %v, want ErrForbidden", err)
	}
	if _, err := operation.Do(ctx, f.r, SetUserRole, me, SetUserRoleInput{User: "camille", Role: "owner"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("set_user_role to owner: err = %v, want ErrInvalid", err)
	}
	u, err := operation.Do(ctx, f.r, SetUserRole, me, SetUserRoleInput{User: "camille", Role: Admin})
	if err != nil || u.Role != Admin {
		t.Fatalf("set_user_role: %+v, %v", u, err)
	}

	if _, err := operation.Do(ctx, f.r, DisableUser, me, UserInput{User: "olivier"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("disable yourself: err = %v, want ErrInvalid", err)
	}
	tok, err := operation.Do(ctx, f.r, CreateToken, camille, CreateTokenInput{Name: "Claude Code"})
	if err != nil {
		t.Fatal(err)
	}
	u, err = operation.Do(ctx, f.r, DisableUser, me, UserInput{User: "camille"})
	if err != nil || u.Status != Disabled {
		t.Fatalf("disable_user: %+v, %v", u, err)
	}
	if _, err := operation.Do(ctx, f.r, GetAccount, camille, GetAccountInput{}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("a call by a disabled person: err = %v, want ErrForbidden", err)
	}
	if _, err := Verify(ctx, f.db, tok.Token, f.now); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("the token of a disabled person: err = %v, want ErrInvalidToken", err)
	}
	var sessions int
	if err := f.db.Read.QueryRowContext(ctx, "SELECT count(*) FROM sessions WHERE user_id = ?", camille.UserID).Scan(&sessions); err != nil || sessions != 0 {
		t.Errorf("sessions of a disabled person: %d, %v", sessions, err)
	}
	_, err = operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: "camille", Password: password})
	if !errors.Is(err, operation.ErrForbidden) || !strings.Contains(err.Error(), "disabled") {
		t.Errorf("sign_in of a disabled person: err = %v", err)
	}
	// A disabled administrator does not count: olivier is the last one.
	if _, err := operation.Do(ctx, f.r, SetUserRole, me, SetUserRoleInput{User: "olivier", Role: Member}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("demote the last administrator who is not disabled: err = %v, want ErrInvalid", err)
	}

	u, err = operation.Do(ctx, f.r, EnableUser, me, UserInput{User: "camille"})
	if err != nil || u.Status != Active {
		t.Fatalf("enable_user: %+v, %v", u, err)
	}
	if _, err := Verify(ctx, f.db, tok.Token, f.now); err != nil {
		t.Errorf("the token after enable_user: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, EnableUser, me, UserInput{User: "camille"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("enable an enabled person: err = %v, want ErrInvalid", err)
	}
	if _, err := operation.Do(ctx, f.r, DisableUser, me, UserInput{User: "nobody"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("disable nobody: err = %v, want ErrNotFound", err)
	}
}

func TestListUsersAndGetAccount(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	me := f.active(t, "olivier", true)
	f.person(t, "camille", false)

	list, err := operation.Do(ctx, f.r, ListUsers, me, ListUsersInput{})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, u := range list.Users {
		got = append(got, u.Handle+" "+u.Role+" "+u.Status)
	}
	if want := []string{"camille member pending", "olivier admin active"}; !slices.Equal(got, want) {
		t.Errorf("list_users = %q, want %q", got, want)
	}

	if _, err := f.db.Write.ExecContext(ctx, `
		INSERT INTO rooms (name, kind, title, created_at) VALUES ('marketing', 'team', 'Marketing', 'now');
		INSERT INTO room_members (room_id, user_id, role) VALUES (last_insert_rowid(), ?, 'member')`, me.UserID); err != nil {
		t.Fatal(err)
	}
	a, err := operation.Do(ctx, f.r, GetAccount, me, GetAccountInput{})
	if err != nil {
		t.Fatal(err)
	}
	want := []Membership{{Room: "olivier", Title: "OLIVIER", Kind: "personal", Role: "owner"}, {Room: "marketing", Title: "Marketing", Kind: "team", Role: "member"}}
	if a.Handle != "olivier" || a.Role != Admin || !slices.Equal(a.Rooms, want) {
		t.Errorf("get_account = %+v", a)
	}
	if _, err := operation.Do(ctx, f.r, GetAccount, admin, GetAccountInput{}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("get_account by the command line: err = %v, want ErrInvalid", err)
	}
}

func TestTokens(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.person(t, "olivier", true)

	nt, err := operation.Do(ctx, f.r, CreateToken, admin, CreateTokenInput{User: "olivier", Name: "Claude Code"})
	if err != nil {
		t.Fatalf("create_token: %v", err)
	}
	if !strings.HasPrefix(nt.Token, TokenPrefix) || len(nt.Token) != len(TokenPrefix)+52 {
		t.Errorf("token = %q", nt.Token)
	}
	if _, err := operation.Do(ctx, f.r, CreateToken, admin, CreateTokenInput{User: "nobody", Name: "x"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("token for an unknown user: err = %v, want ErrNotFound", err)
	}
	if _, err := operation.Do(ctx, f.r, CreateToken, admin, CreateTokenInput{Name: "x"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("token for nobody named: err = %v, want ErrInvalid", err)
	}
	for _, e := range f.entries(t) {
		if strings.Contains(e.Target+e.Detail+e.Error, TokenPrefix) {
			t.Errorf("the Logbook holds a token: %+v", e)
		}
	}

	o, err := Verify(ctx, f.db, nt.Token, f.now)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if o.TokenID != nt.ID || o.Handle != "olivier" || o.TokenName != "Claude Code" || !o.Admin {
		t.Errorf("owner = %+v", o)
	}
	for _, bad := range []string{"", "cahut_pat_nope", nt.Token + "x", strings.TrimPrefix(nt.Token, TokenPrefix)} {
		if _, err := Verify(ctx, f.db, bad, f.now); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("Verify(%q): err = %v, want ErrInvalidToken", bad, err)
		}
	}
	list, err := operation.Do(ctx, f.r, ListTokens, admin, ListTokensInput{User: "olivier"})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tokens) != 1 || list.Tokens[0].LastUsedAt == nil || !list.Tokens[0].LastUsedAt.Equal(f.now) {
		t.Errorf("list_tokens = %+v", list)
	}

	// A member creates, lists and revokes their own tokens, and nobody
	// else's.
	camille := f.active(t, "camille", false)
	own, err := operation.Do(ctx, f.r, CreateToken, camille, CreateTokenInput{Name: "Claude Code"})
	if err != nil || own.User != "camille" {
		t.Fatalf("create_token by a member: %+v, %v", own, err)
	}
	if _, err := operation.Do(ctx, f.r, CreateToken, camille, CreateTokenInput{User: "olivier", Name: "x"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("a member's token for another: err = %v, want ErrForbidden", err)
	}
	list, err = operation.Do(ctx, f.r, ListTokens, camille, ListTokensInput{})
	if err != nil || len(list.Tokens) != 1 || list.Tokens[0].ID != own.ID {
		t.Errorf("list_tokens by a member = %+v, %v", list, err)
	}
	if _, err := operation.Do(ctx, f.r, ListTokens, camille, ListTokensInput{User: "olivier"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("a member lists another's tokens: err = %v, want ErrForbidden", err)
	}
	if _, err := operation.Do(ctx, f.r, RevokeToken, camille, RevokeTokenInput{ID: nt.ID}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("a member revokes another's token: err = %v, want ErrNotFound", err)
	}
	if _, err := operation.Do(ctx, f.r, RevokeToken, camille, RevokeTokenInput{ID: own.ID}); err != nil {
		t.Errorf("a member revokes their token: %v", err)
	}
	ai := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: camille.UserID, Handle: "camille", TokenID: own.ID, Client: "Claude"}
	if _, err := operation.Do(ctx, f.r, CreateToken, ai, CreateTokenInput{Name: "x"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("create_token by an AI: err = %v, want ErrForbidden", err)
	}

	if _, err := operation.Do(ctx, f.r, RevokeToken, admin, RevokeTokenInput{ID: nt.ID}); err != nil {
		t.Fatalf("revoke_token: %v", err)
	}
	if _, err := Verify(ctx, f.db, nt.Token, f.now); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Verify after revoke: err = %v, want ErrInvalidToken", err)
	}
	if _, err := operation.Do(ctx, f.r, RevokeToken, admin, RevokeTokenInput{ID: nt.ID}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("second revoke: err = %v, want ErrInvalid", err)
	}
}
