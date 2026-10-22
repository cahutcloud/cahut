package server

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
)

const password = "correct horse battery staple"

// visitor comes to the console of the Cahut at cahut.example.com, signed out.
var visitor = operation.Actor{Kind: operation.Visitor, Via: "api", Domain: "cahut.example.com"}

// clock makes the registry of s read its time from now, which tests move.
func clock(s *Server) *time.Time {
	now := time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)
	s.Registry.Now = func() time.Time { return now }
	return &now
}

// signedIn returns the actor of the administrator who holds a session token,
// at domain.
func signedIn(t *testing.T, s *Server, token, domain string, now time.Time) operation.Actor {
	t.Helper()
	a, err := s.VerifySession(context.Background(), token, domain, now)
	if err != nil {
		t.Fatalf("VerifySession: %v", err)
	}
	return operation.Actor{Kind: operation.Admin, Via: "api", Domain: domain, Handle: a.Handle, Session: a.SessionID}
}

func TestServerAdmins(t *testing.T) {
	s := setup(t)
	now := clock(s)
	ctx := context.Background()

	a, err := operation.Do(ctx, s.Registry, CreateAdmin, admin, CreateAdminInput{Handle: "olivier", Name: " Olivier Durand "})
	if err != nil {
		t.Fatalf("create_admin: %v", err)
	}
	if a.Handle != "olivier" || a.Name != "Olivier Durand" || a.Status != account.Pending || len(a.Activation.Code) != 52 {
		t.Errorf("admin = %+v", a)
	}
	for _, in := range []CreateAdminInput{{Handle: "olivier", Name: "Again"}, {Handle: "Olivier", Name: "Capital"}, {Handle: "camille", Name: " "}} {
		if _, err := operation.Do(ctx, s.Registry, CreateAdmin, admin, in); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("create_admin(%+v): err = %v, want ErrInvalid", in, err)
		}
	}

	// Activating opens a session at the domain of the call, and only there.
	if _, err := operation.Do(ctx, s.Registry, ActivateAccount, operation.Actor{Kind: operation.Visitor, Via: "api"},
		account.ActivateInput{Code: a.Activation.Code, Password: password}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("activate_account without a domain: err = %v, want ErrInvalid", err)
	}
	sess, err := operation.Do(ctx, s.Registry, ActivateAccount, visitor, account.ActivateInput{Code: a.Activation.Code, Password: password})
	if err != nil {
		t.Fatalf("activate_account: %v", err)
	}
	if !strings.HasPrefix(sess.Token, SessionPrefix) || sess.Domain != "cahut.example.com" || sess.Admin.Status != account.Active {
		t.Errorf("session = %+v", sess)
	}
	me := signedIn(t, s, sess.Token, "cahut.example.com", *now)
	if _, err := s.VerifySession(ctx, sess.Token, "beta.example.org", *now); !errors.Is(err, account.ErrInvalidSession) {
		t.Errorf("a session at another domain: err = %v, want ErrInvalidSession", err)
	}
	if _, err := s.VerifySession(ctx, "cahut_ses_"+strings.TrimPrefix(sess.Token, SessionPrefix), "cahut.example.com", *now); !errors.Is(err, account.ErrInvalidSession) {
		t.Errorf("a session token of a person: err = %v, want ErrInvalidSession", err)
	}
	if _, err := s.VerifySession(ctx, sess.Token, "cahut.example.com", now.Add(credential.SessionIdle)); !errors.Is(err, account.ErrInvalidSession) {
		t.Errorf("a session after 7 days without use: err = %v", err)
	}

	// The administrator who signed in acts on the server.
	if _, err := operation.Do(ctx, s.Registry, CreateAdmin, me, CreateAdminInput{Handle: "camille", Name: "Camille"}); err != nil {
		t.Errorf("create_admin by an administrator who signed in: %v", err)
	}
	list, err := operation.Do(ctx, s.Registry, ListAdmins, me, ListAdminsInput{})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, a := range list.Admins {
		got = append(got, a.Handle+" "+a.Status)
	}
	if want := []string{"camille pending", "olivier active"}; !slices.Equal(got, want) {
		t.Errorf("list_admins = %q, want %q", got, want)
	}

	// Sign in, at another domain, then out.
	beta := visitor
	beta.Domain = "beta.example.org"
	for _, in := range []account.SignInInput{{Handle: "olivier", Password: "wrong"}, {Handle: "nobody", Password: password}, {Handle: "camille", Password: password}} {
		_, err := operation.Do(ctx, s.Registry, SignIn, beta, in)
		if !errors.Is(err, operation.ErrForbidden) || !strings.Contains(err.Error(), "wrong handle or password") {
			t.Errorf("sign_in(%q): err = %v", in.Handle, err)
		}
	}
	other, err := operation.Do(ctx, s.Registry, SignIn, beta, account.SignInInput{Handle: " Olivier ", Password: password})
	if err != nil || other.Domain != "beta.example.org" {
		t.Fatalf("sign_in = %+v, %v", other, err)
	}
	atBeta := signedIn(t, s, other.Token, "beta.example.org", *now)
	if _, err := operation.Do(ctx, s.Registry, SignOut, atBeta, account.SignOutInput{}); err != nil {
		t.Fatalf("sign_out: %v", err)
	}
	if _, err := s.VerifySession(ctx, other.Token, "beta.example.org", *now); !errors.Is(err, account.ErrInvalidSession) {
		t.Errorf("a session after sign_out: err = %v", err)
	}

	// Changing the password ends the other sessions.
	other, err = operation.Do(ctx, s.Registry, SignIn, beta, account.SignInInput{Handle: "olivier", Password: password})
	if err != nil {
		t.Fatal(err)
	}
	const newPassword = "a longer and newer passphrase"
	if _, err := operation.Do(ctx, s.Registry, ChangePassword, me, account.ChangePasswordInput{Password: "wrong", NewPassword: newPassword}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("change_password with a wrong password: err = %v, want ErrForbidden", err)
	}
	if _, err := operation.Do(ctx, s.Registry, ChangePassword, me, account.ChangePasswordInput{Password: password, NewPassword: newPassword}); err != nil {
		t.Fatalf("change_password: %v", err)
	}
	if _, err := s.VerifySession(ctx, other.Token, "beta.example.org", *now); !errors.Is(err, account.ErrInvalidSession) {
		t.Errorf("the other session after change_password: err = %v", err)
	}
	signedIn(t, s, sess.Token, "cahut.example.com", *now)
	if _, err := operation.Do(ctx, s.Registry, ChangePassword, admin, account.ChangePasswordInput{Password: password, NewPassword: newPassword}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("change_password on the command line: err = %v, want ErrInvalid", err)
	}

	// The Logbook names the administrator who signed in, as an admin, but not
	// an unknown handle, and never a password, a code or a token.
	entries, err := logbook.List(ctx, s.DB.Read, 100)
	if err != nil {
		t.Fatal(err)
	}
	var signIns []string
	for _, e := range entries {
		for _, secret := range []string{password, newPassword, a.Activation.Code, sess.Token, other.Token, "nobody"} {
			if strings.Contains(e.Target+e.Detail+e.Error, secret) {
				t.Errorf("the Logbook holds %q: %+v", secret, e)
			}
		}
		if e.Operation == "sign_in" {
			signIns = append(signIns, string(e.Outcome)+" "+e.Actor+" "+e.UserHandle)
		}
	}
	want := []string{"ok admin olivier", "ok admin olivier", "refused admin camille", "refused admin ", "refused admin olivier"}
	if !slices.Equal(signIns, want) {
		t.Errorf("sign_in entries = %q, want %q", signIns, want)
	}
}

func TestServerAdminLinksAndLimits(t *testing.T) {
	s := setup(t)
	now := clock(s)
	ctx := context.Background()
	a, err := operation.Do(ctx, s.Registry, CreateAdmin, admin, CreateAdminInput{Handle: "olivier", Name: "Olivier"})
	if err != nil {
		t.Fatal(err)
	}
	link, err := operation.Do(ctx, s.Registry, CreateActivationLink, admin, AdminInput{Admin: "olivier"})
	if err != nil {
		t.Fatalf("create_activation_link: %v", err)
	}
	if _, err := operation.Do(ctx, s.Registry, ActivateAccount, visitor, account.ActivateInput{Code: a.Activation.Code, Password: password}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("a replaced link: err = %v, want ErrForbidden", err)
	}
	if _, err := operation.Do(ctx, s.Registry, ActivateAccount, visitor, account.ActivateInput{Code: link.Code, Password: password}); err != nil {
		t.Fatalf("activate_account: %v", err)
	}
	if _, err := operation.Do(ctx, s.Registry, CreateActivationLink, admin, AdminInput{Admin: "nobody"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("create_activation_link for nobody: err = %v, want ErrNotFound", err)
	}

	for range credential.MaxRefusals {
		if _, err := operation.Do(ctx, s.Registry, SignIn, visitor, account.SignInInput{Handle: "olivier", Password: "wrong"}); !errors.Is(err, operation.ErrForbidden) {
			t.Fatalf("wrong password: err = %v", err)
		}
	}
	if _, err := operation.Do(ctx, s.Registry, SignIn, visitor, account.SignInInput{Handle: "olivier", Password: password}); !errors.Is(err, operation.ErrTooMany) {
		t.Errorf("sign_in after %d refusals: err = %v, want ErrTooMany", credential.MaxRefusals, err)
	}
	*now = now.Add(credential.RefusalWindow + time.Minute)
	if _, err := operation.Do(ctx, s.Registry, SignIn, visitor, account.SignInInput{Handle: "olivier", Password: password}); err != nil {
		t.Errorf("sign_in once the refusals are old: %v", err)
	}

	// People and AIs never reach the operations of the server.
	person := operation.Actor{Kind: operation.Person, Via: "api", UserID: 1, Handle: "olivier", Session: 1}
	if _, err := operation.Do(ctx, s.Registry, ListAdmins, person, ListAdminsInput{}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("list_admins by a person: err = %v, want ErrForbidden", err)
	}
	if _, err := operation.Do(ctx, s.Registry, SignOut, visitor, account.SignOutInput{}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("sign_out by a visitor: err = %v, want ErrForbidden", err)
	}
}

func TestAddDomainWithAdmin(t *testing.T) {
	s := setup(t)
	ctx := context.Background()
	n, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{
		Domain: "cahut.example.com", Admin: &FirstAdmin{Handle: "ana", Name: "Ana Lima"},
	})
	if err != nil {
		t.Fatalf("add_domain: %v", err)
	}
	if n.Domain != "cahut.example.com" || n.Admin == nil || n.Admin.Role != account.Admin || n.Admin.Activation == nil {
		t.Fatalf("add_domain = %+v", n)
	}
	c, err := s.OpenCahut(ctx, "cahut.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Do(ctx, c.Registry, account.ActivateAccount, visitor, account.ActivateInput{Code: n.Admin.Activation.Code, Password: password}); err != nil {
		t.Errorf("activate the first administrator: %v", err)
	}
	c.Close()

	// A first administrator that cannot be created leaves no Cahut.
	_, err = operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{
		Domain: "beta.example.org", Admin: &FirstAdmin{Handle: "Ana", Name: "Ana"},
	})
	if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "first administrator") {
		t.Errorf("add_domain with a bad first administrator: err = %v", err)
	}
	if ok, _ := s.Declared(ctx, "beta.example.org"); ok {
		t.Error("the Cahut was declared all the same")
	}
	dirs, err := os.ReadDir(filepath.Join(s.cfg.DataDir, "cahuts"))
	if err != nil || len(dirs) != 1 {
		t.Errorf("Cahut directories: %v, %v; want only the first", dirs, err)
	}
}
