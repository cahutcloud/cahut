package account

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/operation"
)

func TestServiceTokens(t *testing.T) {
	f := setup(t)
	ctx := context.Background()

	st, err := operation.Do(ctx, f.r, CreateServiceToken, admin, CreateServiceTokenInput{Name: " account "})
	if err != nil {
		t.Fatalf("create_service_token: %v", err)
	}
	if st.Name != "account" || !strings.HasPrefix(st.Token, ServiceTokenPrefix) || len(st.Token) != len(ServiceTokenPrefix)+52 {
		t.Errorf("service token = %+v", st)
	}
	if _, err := operation.Do(ctx, f.r, CreateServiceToken, admin, CreateServiceTokenInput{Name: " "}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("a service token without a name: err = %v, want ErrInvalid", err)
	}
	// The administrators of the Cahut do not create them: those of the
	// server do.
	olivier := f.active(t, "olivier", true)
	if _, err := operation.Do(ctx, f.r, CreateServiceToken, olivier, CreateServiceTokenInput{Name: "x"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("create_service_token by an administrator of the Cahut: err = %v, want ErrForbidden", err)
	}
	for _, e := range f.entries(t) {
		if strings.Contains(e.Target+e.Detail+e.Error, ServiceTokenPrefix) {
			t.Errorf("the Logbook holds a token: %+v", e)
		}
	}

	o, err := VerifyService(ctx, f.db, st.Token, f.now)
	if err != nil || o.TokenID != st.ID || o.Name != "account" {
		t.Fatalf("VerifyService = %+v, %v", o, err)
	}
	pat, err := operation.Do(ctx, f.r, CreateToken, admin, CreateTokenInput{User: "olivier", Name: "Claude"})
	if err != nil {
		t.Fatal(err)
	}
	// Neither kind of token opens the door of the other.
	for _, bad := range []string{"", "cahut_svc_nope", st.Token + "x", strings.TrimPrefix(st.Token, ServiceTokenPrefix), pat.Token} {
		if _, err := VerifyService(ctx, f.db, bad, f.now); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("VerifyService(%q): err = %v, want ErrInvalidToken", bad, err)
		}
	}
	if _, err := Verify(ctx, f.db, st.Token, f.now); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Verify of a service token: err = %v, want ErrInvalidToken", err)
	}

	// The last use is recorded at most once a minute.
	used := f.now
	f.now = f.now.Add(30 * time.Second)
	if _, err := VerifyService(ctx, f.db, st.Token, f.now); err != nil {
		t.Fatal(err)
	}
	list, err := operation.Do(ctx, f.r, ListServiceTokens, admin, ListServiceTokensInput{})
	if err != nil || len(list.Tokens) != 1 || list.Tokens[0].LastUsedAt == nil || !list.Tokens[0].LastUsedAt.Equal(used) {
		t.Errorf("list_service_tokens = %+v, %v", list, err)
	}

	revoked, err := operation.Do(ctx, f.r, RevokeServiceToken, admin, RevokeServiceTokenInput{ID: st.ID})
	if err != nil || revoked.RevokedAt == nil || revoked.Name != "account" {
		t.Fatalf("revoke_service_token = %+v, %v", revoked, err)
	}
	if _, err := VerifyService(ctx, f.db, st.Token, f.now); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("VerifyService after revoke: err = %v, want ErrInvalidToken", err)
	}
	if _, err := operation.Do(ctx, f.r, RevokeServiceToken, admin, RevokeServiceTokenInput{ID: st.ID}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("revoke twice: err = %v, want ErrInvalid", err)
	}
	if _, err := operation.Do(ctx, f.r, RevokeServiceToken, admin, RevokeServiceTokenInput{ID: 99}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("revoke an unknown token: err = %v, want ErrNotFound", err)
	}
}

func TestService(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	st, err := operation.Do(ctx, f.r, CreateServiceToken, admin, CreateServiceTokenInput{Name: "account"})
	if err != nil {
		t.Fatal(err)
	}
	o, err := VerifyService(ctx, f.db, st.Token, f.now)
	if err != nil {
		t.Fatal(err)
	}
	svc := o.Actor("example.com")
	no := false

	// check_handle says why a handle cannot go to a new person.
	if _, err := f.db.Write.ExecContext(ctx, "INSERT INTO rooms (name, kind, title, created_at) VALUES ('marketing', 'team', 'Marketing', 'now')"); err != nil {
		t.Fatal(err)
	}
	f.person(t, "olivier", true)
	for handle, want := range map[string]string{
		"ana": "", "Ana": HandleInvalid, "x": HandleInvalid, "mcp": HandleReserved, "server": HandleReserved, "olivier": HandleTaken, "marketing": HandleTaken,
	} {
		check, err := operation.Do(ctx, f.r, CheckHandle, svc, CheckHandleInput{Handle: handle})
		if err != nil || check.Reason != want || check.Available != (want == "") || check.Handle != handle {
			t.Errorf("check_handle %q = %+v, %v; want reason %q", handle, check, err, want)
		}
	}

	// A service creates a person without a Knock, and gets their link,
	// which stays away from their address when it says so.
	relay := f.mailer(t)
	u, err := operation.Do(ctx, f.r, CreateUser, svc, CreateUserInput{Handle: "ana", Name: "Ana", Email: "ana@example.org", SendActivationEmail: &no})
	if err != nil || u.Activation == nil || u.Activation.Emailed || u.Email != "ana@example.org" {
		t.Fatalf("create_user by a service = %+v, %v", u, err)
	}
	a, err := operation.Do(ctx, f.r, CreateActivationLink, svc, ActivationLinkInput{User: "ana", SendActivationEmail: &no})
	if err != nil || a.Emailed || a.User != "ana" || a.Code == u.Activation.Code {
		t.Fatalf("create_activation_link by a service = %+v, %v", a, err)
	}
	// The first message the relay receives is that of the next person: the
	// two links of ana went nowhere.
	if _, err := operation.Do(ctx, f.r, CreateUser, svc, CreateUserInput{Handle: "bob", Name: "Bob", Email: "bob@example.org"}); err != nil {
		t.Fatal(err)
	}
	if m := read(t, relay, 1); m.to != "bob@example.org" {
		t.Errorf("the first message went to %s", m.to)
	}

	got, err := operation.Do(ctx, f.r, GetUser, svc, UserInput{User: "ana"})
	if err != nil || got.Handle != "ana" || got.Email != "ana@example.org" || got.Status != Pending || got.Role != Member {
		t.Errorf("get_user = %+v, %v", got, err)
	}
	if _, err := operation.Do(ctx, f.r, GetUser, svc, UserInput{User: "nobody"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_user of nobody: err = %v, want ErrNotFound", err)
	}
	if _, err := operation.Do(ctx, f.r, DisableUser, svc, UserInput{User: "ana"}); err != nil {
		t.Errorf("disable_user by a service: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, EnableUser, svc, UserInput{User: "ana"}); err != nil {
		t.Errorf("enable_user by a service: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, SetUserEmail, svc, SetUserEmailInput{User: "ana", Email: "ana@example.net"}); err != nil {
		t.Errorf("set_user_email by a service: %v", err)
	}

	// A service makes no administrator, and reaches no other operation.
	if _, err := operation.Do(ctx, f.r, CreateUser, svc, CreateUserInput{Handle: "eve", Name: "Eve", Admin: true}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("an administrator made by a service: err = %v, want ErrForbidden", err)
	}
	for name, call := range map[string]func() error{
		"set_user_role": func() error {
			_, err := operation.Do(ctx, f.r, SetUserRole, svc, SetUserRoleInput{User: "ana", Role: Admin})
			return err
		},
		"list_users": func() error {
			_, err := operation.Do(ctx, f.r, ListUsers, svc, ListUsersInput{})
			return err
		},
		"create_token": func() error {
			_, err := operation.Do(ctx, f.r, CreateToken, svc, CreateTokenInput{User: "ana", Name: "x"})
			return err
		},
		"create_service_token": func() error {
			_, err := operation.Do(ctx, f.r, CreateServiceToken, svc, CreateServiceTokenInput{Name: "x"})
			return err
		},
	} {
		if err := call(); !errors.Is(err, operation.ErrForbidden) {
			t.Errorf("%s by a service: err = %v, want ErrForbidden", name, err)
		}
	}

	// check_handle and get_user serve people, and not AIs; get_user serves
	// the administrators of the Cahut only.
	olivier := f.active(t, "chloe", true)
	ai := olivier
	ai.Kind, ai.Session, ai.TokenID, ai.Client = operation.AI, 0, 1, "Claude"
	member := f.active(t, "dan", false)
	if _, err := operation.Do(ctx, f.r, GetUser, olivier, UserInput{User: "ana"}); err != nil {
		t.Errorf("get_user by an administrator: %v", err)
	}
	for who, as := range map[string]operation.Actor{"AI of an administrator": ai, "member": member} {
		if _, err := operation.Do(ctx, f.r, GetUser, as, UserInput{User: "ana"}); !errors.Is(err, operation.ErrForbidden) {
			t.Errorf("get_user by the %s: err = %v, want ErrForbidden", who, err)
		}
		if _, err := operation.Do(ctx, f.r, CheckHandle, as, CheckHandleInput{Handle: "zoe"}); !errors.Is(err, operation.ErrForbidden) {
			t.Errorf("check_handle by the %s: err = %v, want ErrForbidden", who, err)
		}
	}

	// The Logbook records the changes of the service as those of an
	// administrator, through admin-api, under the name of its token.
	var changes int
	for _, e := range f.entries(t) {
		if e.Via != "admin-api" {
			continue
		}
		changes++
		if e.Actor != "admin" || e.Client != "service token account" {
			t.Errorf("entry of the service = %+v", e)
		}
	}
	if changes == 0 {
		t.Error("the Logbook holds no call of the service")
	}
}
