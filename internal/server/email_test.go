package server

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"testing"

	"cahut.cloud/internal/account"
	cmail "cahut.cloud/internal/mail"
	"cahut.cloud/internal/mail/mailtest"
	"cahut.cloud/internal/operation"
)

// mailed is a server whose relay is a relay in memory, on this machine and
// without TLS, with the Cahut cahut.example.com.
func mailed(t *testing.T) (*Server, *mailtest.Relay) {
	t.Helper()
	ctx := context.Background()
	s, err := Open(ctx, Config{
		DataDir:    t.TempDir(),
		BaseURL:    func(domain string) string { return "https://" + domain + "/" },
		ConsoleURL: func(domain string) string { return "https://" + domain + "/console/" },
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if _, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{Domain: "cahut.example.com"}); err != nil {
		t.Fatal(err)
	}
	relay := mailtest.Start(t, mailtest.Options{NoStartTLS: true})
	in := cmail.SetMailRelayInput{Host: "127.0.0.1", Port: relay.Port, Security: cmail.Plain, From: "cahut@example.com"}
	if _, err := do[cmail.SetMailRelayInput, cmail.Delivery](ctx, s, "set_mail_relay", in); err != nil {
		t.Fatal(err)
	}
	return s, relay
}

// message waits until the relay holds n messages, and returns the address
// and the subject of the last one, and the link to the console it holds.
func message(t *testing.T, relay *mailtest.Relay, n int) (to, subject, link string) {
	t.Helper()
	m := relay.Wait(t, n)[n-1]
	parsed, err := mail.ReadMessage(m.Reader())
	if err != nil {
		t.Fatal(err)
	}
	body := strings.NewReplacer("=\r\n", "", "=3D", "=").Replace(string(m.Data))
	link = regexp.MustCompile(`https://cahut\.example\.com/console/[^\s"<]+`).FindString(body)
	return m.To[0], parsed.Header.Get("Subject"), link
}

func TestAdminEmail(t *testing.T) {
	s, relay := mailed(t)
	now := clock(s)
	ctx := context.Background()

	a, err := operation.Do(ctx, s.Registry, CreateAdmin, admin, CreateAdminInput{Handle: "olivier", Name: "Olivier", Email: "olivier@example.org"})
	if err != nil || !a.Activation.Emailed || a.Email != "olivier@example.org" {
		t.Fatalf("create_admin: %+v %v", a, err)
	}
	to, subject, link := message(t, relay, 1)
	if to != "olivier@example.org" || subject != "Choose your password for the Cahut server" ||
		link != "https://cahut.example.com/console/server/activate#"+a.Activation.Code {
		t.Errorf("the activation of an administrator: %s %q %s", to, subject, link)
	}
	if _, err := operation.Do(ctx, s.Registry, CreateAdmin, admin, CreateAdminInput{Handle: "other", Name: "Other", Email: "OLIVIER@example.org"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("create_admin with a taken address: %v", err)
	}
	sess, err := operation.Do(ctx, s.Registry, ActivateAccount, visitor, account.ActivateInput{Code: a.Activation.Code, Password: password})
	if err != nil {
		t.Fatal(err)
	}
	olivier := signedIn(t, s, sess.Token, "cahut.example.com", *now)
	if list, err := operation.Do(ctx, s.Registry, ListAdmins, olivier, ListAdminsInput{}); err != nil || list.Admins[0].Email != "olivier@example.org" {
		t.Errorf("list_admins: %+v %v", list, err)
	}

	// A forgotten password, by address.
	out, err := do[account.RequestPasswordLinkInput, account.PasswordLinkRequested](ctx, s, "request_password_link",
		account.RequestPasswordLinkInput{Who: "Olivier@Example.org"}, visitor)
	if err != nil || out.Message == "" {
		t.Fatalf("request_password_link: %+v %v", out, err)
	}
	if to, subject, link := message(t, relay, 2); to != "olivier@example.org" || subject != "Choose a new password for the Cahut server" ||
		!strings.HasPrefix(link, "https://cahut.example.com/console/server/activate?reset#") {
		t.Errorf("the link of a forgotten password: %s %q %s", to, subject, link)
	}
	if _, err := do[account.RequestPasswordLinkInput, account.PasswordLinkRequested](ctx, s, "request_password_link",
		account.RequestPasswordLinkInput{Who: "nobody"}, visitor); err != nil {
		t.Errorf("request_password_link of nobody: %v", err)
	}

	// A new address, confirmed by its link; the old one learns it.
	changed, err := operation.Do(ctx, s.Registry, ChangeEmail, olivier, account.ChangeEmailInput{Email: "olivier@new.example"})
	if err != nil || changed.Waiting == nil {
		t.Fatalf("change_email: %+v %v", changed, err)
	}
	to, _, link = message(t, relay, 3)
	if to != "olivier@new.example" || !strings.HasPrefix(link, "https://cahut.example.com/console/server/confirm-email#") {
		t.Fatalf("the confirmation: %s %s", to, link)
	}
	code := link[strings.Index(link, "#")+1:]
	if done, err := operation.Do(ctx, s.Registry, ConfirmEmail, visitor, account.ConfirmEmailInput{Code: code}); err != nil || done.Email != "olivier@new.example" {
		t.Fatalf("confirm_email: %+v %v", done, err)
	}
	if to, subject, _ := message(t, relay, 4); to != "olivier@example.org" || subject != "Your e-mail address changed" {
		t.Errorf("the notice to the old address: %s %q", to, subject)
	}
	if _, err := operation.Do(ctx, s.Registry, ChangeEmail, admin, account.ChangeEmailInput{Email: "x@example.org"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("change_email on the command line: %v", err)
	}
}

func TestFirstStartWithEmail(t *testing.T) {
	s, relay := mailed(t)
	ctx := context.Background()
	link, err := operation.Do(ctx, s.Registry, s.CreateSetupLink, admin, CreateSetupLinkInput{Domain: "cahut.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	done, err := do[SetUpInput, SetUp](ctx, s, "set_up", SetUpInput{Code: link.Code, Handle: "olivier", Name: "Olivier",
		Password: password, Email: "olivier@example.org"}, visitorAt("cahut.example.com"))
	if err != nil || !done.ServerAdmin {
		t.Fatalf("set_up: %+v %v", done, err)
	}
	// The person chose their password there: no link goes to them.
	if n := len(relay.Messages()); n != 0 {
		t.Errorf("set_up sent %d messages", n)
	}
	list, err := operation.Do(ctx, s.Registry, ListAdmins, admin, ListAdminsInput{})
	if err != nil || len(list.Admins) != 1 || list.Admins[0].Email != "olivier@example.org" {
		t.Errorf("the administrator of the server: %+v %v", list, err)
	}
	c, err := s.OpenCahut(ctx, "cahut.example.com")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	people, err := operation.Do(ctx, c.Registry, account.ListUsers, admin, account.ListUsersInput{})
	if err != nil || len(people.Users) != 1 || people.Users[0].Email != "olivier@example.org" {
		t.Errorf("the first person: %+v %v", people, err)
	}
}
