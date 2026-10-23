package server

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
)

func visitorAt(domain string) operation.Actor {
	return operation.Actor{Kind: operation.Visitor, Via: "api", Domain: domain}
}

func TestFirstStart(t *testing.T) {
	s := setup(t)
	ctx := context.Background()
	for _, d := range []string{"cahut.example.com", "beta.example.org"} {
		if _, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{Domain: d}); err != nil {
			t.Fatal(err)
		}
	}
	old, err := operation.Do(ctx, s.Registry, s.CreateSetupLink, admin, CreateSetupLinkInput{Domain: "cahut.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	link, err := operation.Do(ctx, s.Registry, s.CreateSetupLink, admin, CreateSetupLinkInput{Domain: "Cahut.Example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if link.Domain != "cahut.example.com" || link.Code == old.Code || time.Until(link.ExpiresAt) < 23*time.Hour {
		t.Errorf("create_setup_link = %+v", link)
	}

	get := func(domain, code string) (Setup, error) {
		return do[SetupInput, Setup](ctx, s, "get_setup", SetupInput{Code: code}, visitorAt(domain))
	}
	if _, err := get("cahut.example.com", old.Code); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("the replaced link: err = %v", err)
	}
	if _, err := get("beta.example.org", link.Code); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("the link at another Cahut: err = %v", err)
	}
	got, err := get("cahut.example.com", link.Code)
	if err != nil || !got.ServerAdmin || got.Domain != "cahut.example.com" {
		t.Errorf("get_setup = %+v, %v", got, err)
	}

	setUp := func(in SetUpInput) (SetUp, error) {
		return do[SetUpInput, SetUp](ctx, s, "set_up", in, visitorAt("cahut.example.com"))
	}
	if _, err := setUp(SetUpInput{Code: link.Code, Handle: "olivier", Name: "Olivier", Password: "short"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("a short password: err = %v", err)
	}
	if _, err := setUp(SetUpInput{Code: link.Code, Handle: "Not a handle", Name: "Olivier", Password: password}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("a bad handle: err = %v", err)
	}
	done, err := setUp(SetUpInput{Code: link.Code, Handle: "Olivier", Name: "Olivier Durand", Password: password})
	if err != nil {
		t.Fatalf("set_up: %v", err)
	}
	if !done.ServerAdmin || done.Session.Token == "" || done.Session.User.Handle != "olivier" || done.Session.User.Role != account.Admin {
		t.Errorf("set_up = %+v", done)
	}
	if _, err := setUp(SetUpInput{Code: link.Code, Handle: "other", Name: "Other", Password: password}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("the link used again: err = %v", err)
	}

	// The person signs in at the Cahut, and the administrator at the server,
	// with the same password.
	c, err := s.OpenCahut(ctx, "cahut.example.com")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, err := account.VerifySession(ctx, c.DB, done.Session.Token, time.Now()); err != nil {
		t.Errorf("the session of the person: %v", err)
	}
	if _, err := operation.Do(ctx, s.Registry, SignIn, visitorAt("cahut.example.com"),
		account.SignInInput{Handle: "olivier", Password: password}); err != nil {
		t.Errorf("the administrator signs in: %v", err)
	}
	if _, err := operation.Do(ctx, s.Registry, s.CreateSetupLink, admin, CreateSetupLinkInput{Domain: "cahut.example.com"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("a link for a Cahut that has people: err = %v", err)
	}

	// The second Cahut gets its first person, who does not administer the
	// server.
	second, err := operation.Do(ctx, s.Registry, s.CreateSetupLink, admin, CreateSetupLinkInput{Domain: "beta.example.org"})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := get("beta.example.org", second.Code); err != nil || got.ServerAdmin {
		t.Errorf("get_setup of the second Cahut = %+v, %v", got, err)
	}
	b, err := do[SetUpInput, SetUp](ctx, s, "set_up", SetUpInput{Code: second.Code, Handle: "ana", Name: "Ana", Password: password}, visitorAt("beta.example.org"))
	if err != nil || b.ServerAdmin {
		t.Errorf("set_up of the second Cahut = %+v, %v", b, err)
	}
	admins, err := operation.Do(ctx, s.Registry, ListAdmins, admin, ListAdminsInput{})
	if err != nil || len(admins.Admins) != 1 || admins.Admins[0].Handle != "olivier" || admins.Admins[0].Status != account.Active {
		t.Errorf("list_admins = %+v, %v", admins, err)
	}

	entries, err := logbook.List(ctx, s.DB.Read, 20)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, e := range entries {
		if e.Operation == "set_up" && e.Outcome == logbook.OK {
			lines = append(lines, e.Actor+" "+e.UserHandle+" "+e.Target+" "+e.Detail+" "+string(e.Outcome))
		}
		if strings.Contains(e.Detail+e.Target+e.Error, password) || strings.Contains(e.Detail+e.Target, link.Code) {
			t.Errorf("an entry holds a secret: %+v", e)
		}
	}
	want := []string{"person ana ana  ok", "admin olivier olivier with the administrator of the server ok"}
	if len(lines) != 2 || lines[0] != want[0] || lines[1] != want[1] {
		t.Errorf("set_up entries = %q, want %q first", lines, want)
	}
}

func TestSetupLinkExpires(t *testing.T) {
	s := setup(t)
	ctx := context.Background()
	if _, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{Domain: "cahut.example.com"}); err != nil {
		t.Fatal(err)
	}
	link, err := operation.Do(ctx, s.Registry, s.CreateSetupLink, admin, CreateSetupLinkInput{Domain: "cahut.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB.Write.ExecContext(ctx, "UPDATE setup_links SET expires_at = '2026-01-01T00:00:00.000Z'"); err != nil {
		t.Fatal(err)
	}
	if _, err := do[SetupInput, Setup](ctx, s, "get_setup", SetupInput{Code: link.Code}, visitorAt("cahut.example.com")); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("an expired link: err = %v", err)
	}
	if _, err := operation.Do(ctx, s.Registry, s.CreateSetupLink, admin, CreateSetupLinkInput{Domain: "gone.example.org"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("a link for no Cahut: err = %v", err)
	}
}
