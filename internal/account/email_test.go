package account

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"testing"
	"time"

	cmail "cahut.cloud/internal/mail"
	"cahut.cloud/internal/mail/mailtest"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

const console = "https://example.com/console/"

// mailer gives the Cahut of f the e-mail of its server, and a relay in
// memory as its relay, which the server reaches on this machine as tests
// may. It returns the relay. Before it, the Cahut sends no e-mail.
func (f *fixture) mailer(t *testing.T) *mailtest.Relay {
	t.Helper()
	ctx := context.Background()
	svc := cmail.New(cmail.Config{DataDir: t.TempDir()})
	t.Cleanup(func() { svc.Close(ctx) })
	relay := mailtest.Start(t, mailtest.Options{})
	svc.UseRoots(relay.Roots)
	svc.DialAnywhere()
	_, err := f.db.Write.ExecContext(ctx, `INSERT INTO mail_relay (id, host, port, security, sender, set_at)
		VALUES (1, '127.0.0.1', ?, 'starttls', 'cahut@example.com', ?)`, relay.Port, store.FormatTime(f.now))
	if err != nil {
		t.Fatal(err)
	}
	cmail.Provide(f.r, &cmail.Outbox{
		Service: svc, DB: f.db, Domain: "example.com", Console: console,
	})
	return relay
}

// received is a message the relay received, as a person reads it.
type received struct {
	to, subject, body string
}

// link returns the link to the console that the message holds.
func (m received) link(t *testing.T) string {
	t.Helper()
	l := regexp.MustCompile(regexp.QuoteMeta(console) + `[^\s"<]+`).FindString(m.body)
	if l == "" {
		t.Fatalf("no link in %q", m.subject)
	}
	return l
}

// read waits until the relay holds n messages, and returns the last one.
func read(t *testing.T, relay *mailtest.Relay, n int) received {
	t.Helper()
	msgs := relay.Wait(t, n)
	m := msgs[n-1]
	parsed, err := mail.ReadMessage(m.Reader())
	if err != nil {
		t.Fatal(err)
	}
	return received{to: m.To[0], subject: parsed.Header.Get("Subject"), body: decodeQP(string(m.Data))}
}

// decodeQP undoes the soft line breaks and the escapes of quoted-printable
// that the tests look through, enough to find links and words.
func decodeQP(s string) string {
	s = strings.ReplaceAll(s, "=\r\n", "")
	return strings.NewReplacer("=3D", "=", "=20", " ").Replace(s)
}

func TestCreateUserWithEmail(t *testing.T) {
	f := setup(t)
	ctx := context.Background()

	// Without a relay, the link is only given.
	u, err := operation.Do(ctx, f.r, CreateUser, admin, CreateUserInput{Handle: "ana", Name: "Ana", Email: "ana@example.org"})
	if err != nil || u.Activation.Emailed || u.Email != "ana@example.org" {
		t.Fatalf("create_user without a relay: %+v %v", u, err)
	}

	relay := f.mailer(t)
	u, err = operation.Do(ctx, f.r, CreateUser, admin, CreateUserInput{Handle: "olivier", Name: "Olivier Durand", Admin: true, Email: " Olivier@Example.ORG "})
	if err != nil || !u.Activation.Emailed || u.Email != "Olivier@example.org" {
		t.Fatalf("create_user: %+v %v", u, err)
	}
	m := read(t, relay, 1)
	if m.to != "Olivier@example.org" || m.subject != "Choose your password for example.com" ||
		m.link(t) != console+"activate#"+u.Activation.Code || !strings.Contains(m.body, "Your handle is olivier.") {
		t.Errorf("the activation: %+v", m)
	}

	// An address is unique in the Cahut, whatever its case.
	if _, err := operation.Do(ctx, f.r, CreateUser, admin, CreateUserInput{Handle: "other", Name: "Other", Email: "olivier@EXAMPLE.org"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("create_user with a taken address: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, CreateUser, admin, CreateUserInput{Handle: "other", Name: "Other", Email: "Other <other@example.org>"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("create_user with a name in the address: %v", err)
	}

	// A new link goes to the address too.
	a, err := operation.Do(ctx, f.r, CreateActivationLink, admin, ActivationLinkInput{User: "olivier"})
	if err != nil || !a.Emailed {
		t.Fatalf("create_activation_link: %+v %v", a, err)
	}
	if m := read(t, relay, 2); m.link(t) != console+"activate#"+a.Code {
		t.Errorf("the new link: %+v", m)
	}

	// The administrators read the addresses; the people and the AIs do not.
	member := f.active(t, "bob", false)
	olivier := f.active(t, "chloe", true)
	ai := olivier
	ai.Kind, ai.Session, ai.TokenID, ai.Client = operation.AI, 0, 1, "Claude"
	for _, tt := range []struct {
		who  string
		as   operation.Actor
		want bool
	}{{"administrator", olivier, true}, {"member", member, false}, {"AI of an administrator", ai, false}} {
		list, err := operation.Do(ctx, f.r, ListUsers, tt.as, ListUsersInput{})
		if err != nil {
			t.Fatal(err)
		}
		shown := false
		for _, u := range list.Users {
			shown = shown || u.Email != ""
		}
		if shown != tt.want {
			t.Errorf("list_users for the %s shows addresses: %v", tt.who, shown)
		}
	}
	for _, e := range f.entries(t) {
		if strings.Contains(e.Target+e.Detail, "@") {
			t.Errorf("an address in the Logbook: %+v", e)
		}
	}
}

func TestChangeEmail(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ana := f.active(t, "ana", false)
	// Without a relay, a person has no address to change.
	if _, err := operation.Do(ctx, f.r, ChangeEmail, ana, ChangeEmailInput{Email: "ana@new.example"}); !errors.Is(err, operation.ErrInvalid) ||
		!strings.Contains(err.Error(), "sends no e-mail") {
		t.Errorf("change_email without a relay: %v", err)
	}
	relay := f.mailer(t)
	if _, err := operation.Do(ctx, f.r, SetUserEmail, admin, SetUserEmailInput{User: "ana", Email: "ana@example.org"}); err != nil {
		t.Fatal(err)
	}

	// An AI does not change the address of its person.
	ai := ana
	ai.Kind, ai.Session, ai.TokenID, ai.Client = operation.AI, 0, 1, "Claude"
	if _, err := operation.Do(ctx, f.r, ChangeEmail, ai, ChangeEmailInput{Email: "evil@example.net"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("change_email by an AI: %v", err)
	}
	if acc, err := operation.Do(ctx, f.r, GetAccount, ai, GetAccountInput{}); err != nil || acc.Email != "" {
		t.Errorf("get_account by an AI gives the address: %+v %v", acc, err)
	}

	out, err := operation.Do(ctx, f.r, ChangeEmail, ana, ChangeEmailInput{Email: "ana@new.example"})
	if err != nil || out.Email != "ana@example.org" || out.Waiting == nil || out.Waiting.Email != "ana@new.example" {
		t.Fatalf("change_email: %+v %v", out, err)
	}
	acc, err := operation.Do(ctx, f.r, GetAccount, ana, GetAccountInput{})
	if err != nil || acc.Email != "ana@example.org" || acc.EmailChange == nil || acc.EmailChange.Email != "ana@new.example" {
		t.Errorf("get_account while a change waits: %+v %v", acc, err)
	}
	m := read(t, relay, 1)
	if m.to != "ana@new.example" || m.subject != "Confirm your e-mail address" {
		t.Fatalf("the confirmation: %+v", m)
	}
	code := strings.TrimPrefix(m.link(t), console+"confirm-email#")

	// The link confirms it, without signing in; the old address learns it.
	done, err := operation.Do(ctx, f.r, ConfirmEmail, visitor, ConfirmEmailInput{Code: code})
	if err != nil || done.Email != "ana@new.example" {
		t.Fatalf("confirm_email: %+v %v", done, err)
	}
	if m := read(t, relay, 2); m.to != "ana@example.org" || m.subject != "Your e-mail address changed" || !strings.Contains(m.body, "is now ana@new.example") {
		t.Errorf("the notice to the old address: %+v", m)
	}
	if _, err := operation.Do(ctx, f.r, ConfirmEmail, visitor, ConfirmEmailInput{Code: code}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("confirm_email twice: %v", err)
	}

	// An address taken meanwhile is refused when the link is used, which
	// only the owner of the address can do.
	if _, err := operation.Do(ctx, f.r, ChangeEmail, ana, ChangeEmailInput{Email: "bob@example.org"}); err != nil {
		t.Fatal(err)
	}
	code = strings.TrimPrefix(read(t, relay, 3).link(t), console+"confirm-email#")
	f.person(t, "bob", false)
	if _, err := operation.Do(ctx, f.r, SetUserEmail, admin, SetUserEmailInput{User: "bob", Email: "BOB@example.org"}); err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Do(ctx, f.r, ConfirmEmail, visitor, ConfirmEmailInput{Code: code}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("confirm_email of a taken address: %v", err)
	}

	// The link expires after 24 hours.
	if _, err := operation.Do(ctx, f.r, ChangeEmail, ana, ChangeEmailInput{Email: "ana@later.example"}); err != nil {
		t.Fatal(err)
	}
	code = strings.TrimPrefix(read(t, relay, 4).link(t), console+"confirm-email#")
	f.now = f.now.Add(25 * time.Hour)
	if _, err := operation.Do(ctx, f.r, ConfirmEmail, visitor, ConfirmEmailInput{Code: code}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("confirm_email after 25 hours: %v", err)
	}

	// An empty address removes it at once.
	if _, err := operation.Do(ctx, f.r, ChangeEmail, ana, ChangeEmailInput{}); err != nil {
		t.Fatal(err)
	}
	if m := read(t, relay, 5); m.to != "ana@new.example" || !strings.Contains(m.body, "was removed") {
		t.Errorf("the notice of a removed address: %+v", m)
	}
	if acc, _ := operation.Do(ctx, f.r, GetAccount, ana, GetAccountInput{}); acc.Email != "" || acc.EmailChange != nil {
		t.Errorf("get_account after removing the address: %+v", acc)
	}
}

func TestRequestPasswordLink(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ask := func(who string) {
		t.Helper()
		out, err := operation.Do(ctx, f.r, RequestPasswordLink, visitor, RequestPasswordLinkInput{Who: who})
		if err != nil || out.Message != passwordLinkAnswer {
			t.Fatalf("request_password_link %q: %+v %v", who, out, err)
		}
	}
	describe := func() bool {
		t.Helper()
		p, err := operation.Do(ctx, f.r, DescribeSignIn, visitor, DescribeSignInInput{})
		if err != nil {
			t.Fatalf("describe_sign_in: %v", err)
		}
		return p.SendsMail
	}
	f.active(t, "olivier", false)
	f.active(t, "bob", false) // without an address
	f.active(t, "chloe", true)
	if _, err := operation.Do(ctx, f.r, SetUserEmail, admin, SetUserEmailInput{User: "olivier", Email: "olivier@example.org"}); err != nil {
		t.Fatal(err)
	}
	// Without a relay, the page of sign-in offers no link, and a visitor who
	// asks all the same hears that the Cahut sends no e-mail.
	if describe() {
		t.Error("describe_sign_in says a Cahut without a relay sends e-mail")
	}
	if _, err := operation.Do(ctx, f.r, RequestPasswordLink, visitor, RequestPasswordLinkInput{Who: "olivier"}); !errors.Is(err, operation.ErrInvalid) ||
		!strings.Contains(err.Error(), "sends no e-mail") {
		t.Errorf("request_password_link without a relay: %v", err)
	}

	relay := f.mailer(t)
	if !describe() {
		t.Error("describe_sign_in says a Cahut with a relay sends no e-mail")
	}
	ask("OLIVIER@example.org")
	m := read(t, relay, 1)
	if m.to != "olivier@example.org" || m.subject != "Choose a new password for example.com" || !strings.HasPrefix(m.link(t), console+"activate?reset#") {
		t.Fatalf("the link of a forgotten password: %+v", m)
	}

	// Not again within 5 minutes, and never for an unknown account, one
	// without an address, or a disabled one: the answer stays the same.
	ask("olivier")
	ask("nobody")
	ask("nobody@example.org")
	ask("bob")
	chloe := f.signedIn(t, mustSignIn(t, f, "chloe"))
	if _, err := operation.Do(ctx, f.r, DisableUser, chloe, UserInput{User: "olivier"}); err != nil {
		t.Fatal(err)
	}
	f.now = f.now.Add(10 * time.Minute)
	ask("olivier")
	if _, err := operation.Do(ctx, f.r, EnableUser, chloe, UserInput{User: "olivier"}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	if n := len(relay.Messages()); n != 1 {
		t.Fatalf("the relay received %d messages, want 1", n)
	}

	// The link of disabling is gone; a new one works for an hour, and the
	// password works until it is used.
	ask("olivier")
	code := strings.TrimPrefix(read(t, relay, 2).link(t), console+"activate?reset#")
	if _, err := operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: "olivier", Password: password}); err != nil {
		t.Errorf("sign_in with the old password before the link is used: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, ActivateAccount, visitor, ActivateInput{Code: code, Password: "a brand new password here"}); err != nil {
		t.Fatalf("activate_account with the link: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, SignIn, visitor, SignInInput{Handle: "olivier", Password: password}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("sign_in with the old password after: %v", err)
	}
	// The person hears that their password changed.
	if m := read(t, relay, 3); m.to != "olivier@example.org" || m.subject != "Your password changed" {
		t.Errorf("the notice of a new password: %+v", m)
	}
	f.now = f.now.Add(10 * time.Minute)
	ask("olivier")
	code = strings.TrimPrefix(read(t, relay, 4).link(t), console+"activate?reset#")
	f.now = f.now.Add(61 * time.Minute)
	if _, err := operation.Do(ctx, f.r, ActivateAccount, visitor, ActivateInput{Code: code, Password: "yet another new password"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("activate_account with a link older than an hour: %v", err)
	}

	// The Logbook says what happened, under the person, without what the
	// visitor typed.
	var details []string
	for _, e := range f.entries(t) {
		if e.Operation == "request_password_link" {
			if e.Error != "" {
				details = append(details, "refused: "+e.Error)
				continue
			}
			details = append(details, e.UserHandle+": "+e.Detail)
			if strings.Contains(e.Target+e.Detail, "@") {
				t.Errorf("an address in the Logbook: %+v", e)
			}
		}
	}
	want := []string{"olivier: link sent", "olivier: link sent", "olivier: not sent: disabled", "bob: not sent: no address",
		": unknown account", ": unknown account", "olivier: not sent: a link went out less than 5 minutes ago",
		"olivier: link sent", "refused: invalid input: this Cahut sends no e-mail: an administrator of the Cahut can give you a new link"}
	if strings.Join(details, "; ") != strings.Join(want, "; ") {
		t.Errorf("the Logbook, newest first:\n%q\nwant\n%q", details, want)
	}
}

// mustSignIn signs in the person handle, and returns their session token.
func mustSignIn(t *testing.T, f *fixture, handle string) string {
	t.Helper()
	s, err := operation.Do(context.Background(), f.r, SignIn, visitor, SignInInput{Handle: handle, Password: password})
	if err != nil {
		t.Fatal(err)
	}
	return s.Token
}

func TestSecurityNotices(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	relay := f.mailer(t)
	ana := f.active(t, "ana", false)
	chloe := f.active(t, "chloe", true)
	if _, err := operation.Do(ctx, f.r, SetUserEmail, admin, SetUserEmailInput{User: "ana", Email: "ana@example.org"}); err != nil {
		t.Fatal(err)
	}

	// A new password, a personal token and a new role are about the
	// account: they reach the person even with notices off.
	if _, err := operation.Do(ctx, f.r, SetMailNotices, ana, SetMailNoticesInput{Off: true}); err != nil {
		t.Fatal(err)
	}
	if acc, err := operation.Do(ctx, f.r, GetAccount, ana, GetAccountInput{}); err != nil || !acc.NoticesOff {
		t.Errorf("get_account with notices off: %+v %v", acc, err)
	}
	if _, err := operation.Do(ctx, f.r, ChangePassword, ana, ChangePasswordInput{Password: password, NewPassword: "another long password"}); err != nil {
		t.Fatal(err)
	}
	if m := read(t, relay, 1); m.to != "ana@example.org" || m.subject != "Your password changed" {
		t.Errorf("the notice of a new password: %+v", m)
	}
	if _, err := operation.Do(ctx, f.r, CreateToken, ana, CreateTokenInput{Name: "Claude Code"}); err != nil {
		t.Fatal(err)
	}
	if m := read(t, relay, 2); m.subject != "A personal token was created" || !strings.Contains(m.body, "named Claude Code") {
		t.Errorf("the notice of a token: %+v", m)
	}
	if _, err := operation.Do(ctx, f.r, SetUserRole, chloe, SetUserRoleInput{User: "ana", Role: Admin}); err != nil {
		t.Fatal(err)
	}
	if m := read(t, relay, 3); m.subject != "You are now an administrator of example.com" {
		t.Errorf("the notice of a role: %+v", m)
	}
	if _, err := operation.Do(ctx, f.r, SetUserRole, chloe, SetUserRoleInput{User: "ana", Role: Member}); err != nil {
		t.Fatal(err)
	}
	if m := read(t, relay, 4); m.subject != "You are no longer an administrator of example.com" {
		t.Errorf("the notice of a role taken back: %+v", m)
	}
	// An AI does not turn notices off.
	ai := ana
	ai.Kind, ai.Session, ai.TokenID, ai.Client = operation.AI, 0, 1, "Claude"
	if _, err := operation.Do(ctx, f.r, SetMailNotices, ai, SetMailNoticesInput{Off: true}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("set_mail_notices by an AI: %v", err)
	}
}
