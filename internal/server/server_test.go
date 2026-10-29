package server

import (
	"context"
	"errors"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
)

var admin = operation.Actor{Kind: operation.Admin, Via: "cli"}

func setup(t *testing.T) *Server {
	t.Helper()
	s, err := Open(context.Background(), Config{
		DataDir: t.TempDir(),
		BaseURL: func(domain string) string { return "https://" + domain + "/" },
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestAddDomain(t *testing.T) {
	s := setup(t)
	ctx := context.Background()

	for _, domain := range []string{"cahut.example.com", " Beta.Example.org. "} {
		if _, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{Domain: domain}); err != nil {
			t.Fatalf("add_domain %q: %v", domain, err)
		}
	}
	for _, domain := range []string{"cahut.example.com", "*.example.com", "10.0.0.1", "127.0.0.1", "2001:db8::7", "203.0.113", ""} {
		if _, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{Domain: domain}); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("add_domain %q: err = %v, want ErrInvalid", domain, err)
		}
	}

	list, err := operation.Do(ctx, s.Registry, s.ListDomains, admin, ListDomainsInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Cahuts) != 2 || list.Cahuts[0].Domain != "cahut.example.com" || list.Cahuts[1].Domain != "beta.example.org" {
		t.Errorf("list_domains = %+v", list)
	}
	if ok, err := s.Declared(ctx, "beta.example.org"); !ok || err != nil {
		t.Errorf("Declared(beta.example.org) = %v, %v", ok, err)
	}
	if ok, err := s.Declared(ctx, "example.net"); ok || err != nil {
		t.Errorf("Declared(example.net) = %v, %v", ok, err)
	}

	entries, err := logbook.List(ctx, s.DB.Read, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 9 || entries[8].Operation != "add_domain" || entries[8].Target != "cahut.example.com" {
		t.Errorf("server Logbook = %+v", entries)
	}
}

// TestCahutsAreSealedOff checks that each Cahut has its own people.
func TestCahutsAreSealedOff(t *testing.T) {
	s := setup(t)
	ctx := context.Background()
	for _, domain := range []string{"a.example.com", "b.example.com"} {
		if _, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{Domain: domain}); err != nil {
			t.Fatal(err)
		}
	}

	a, err := s.OpenCahut(ctx, "a.example.com")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if a.Domain != "a.example.com" {
		t.Errorf("Domain = %q", a.Domain)
	}
	if _, err := operation.Do(ctx, a.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: "olivier", Name: "Olivier"}); err != nil {
		t.Fatal(err)
	}

	b, err := s.OpenCahut(ctx, "b.example.com")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if _, err := operation.Do(ctx, b.Registry, account.CreateToken, admin, account.CreateTokenInput{User: "olivier", Name: "x"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("token in b for a person of a: err = %v, want ErrNotFound", err)
	}
	if _, err := operation.Do(ctx, b.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: "olivier", Name: "Another Olivier"}); err != nil {
		t.Errorf("same handle in another Cahut: %v", err)
	}

	if _, err := s.OpenCahut(ctx, "c.example.com"); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("OpenCahut(c.example.com): err = %v, want ErrNotFound", err)
	}
	if _, err := s.OpenCahut(ctx, "../a.example.com"); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("OpenCahut(../a.example.com): err = %v, want ErrNotFound", err)
	}
}

func TestCahutAtAnIPAddress(t *testing.T) {
	s := setup(t)
	ctx := context.Background()
	n, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{Domain: "203.0.113.7"})
	if err != nil {
		t.Fatal(err)
	}
	if n.Domain != "203.0.113.7" {
		t.Errorf("domain = %q", n.Domain)
	}
	c, err := s.OpenCahut(ctx, "203.0.113.7")
	if err != nil {
		t.Fatalf("OpenCahut at the address: %v", err)
	}
	c.Close()
	if _, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{Domain: "203.0.113.7"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("a second Cahut at the address: err = %v", err)
	}
}

func TestRenameDomain(t *testing.T) {
	s := setup(t)
	ctx := context.Background()
	if _, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{
		Domain: "203.0.113.7", Admin: &FirstAdmin{Handle: "olivier", Name: "Olivier"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{Domain: "beta.example.org"}); err != nil {
		t.Fatal(err)
	}
	before, err := s.CahutID(ctx, "203.0.113.7")
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct{ from, to string }{
		{"203.0.113.7", "beta.example.org"}, // taken
		{"203.0.113.7", "192.168.1.2"},
		{"203.0.113.7", "not a domain"},
		{"203.0.113.7", "203.0.113.7"},
	} {
		if _, err := operation.Do(ctx, s.Registry, RenameDomain, admin, RenameDomainInput{Domain: tt.from, To: tt.to}); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("rename %s to %s: err = %v, want ErrInvalid", tt.from, tt.to, err)
		}
	}
	if _, err := operation.Do(ctx, s.Registry, RenameDomain, admin, RenameDomainInput{Domain: "gone.example.org", To: "new.example.org"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("rename a Cahut that is not there: err = %v", err)
	}

	info, err := operation.Do(ctx, s.Registry, RenameDomain, admin, RenameDomainInput{Domain: "203.0.113.7", To: " Cahut.Example.com. "})
	if err != nil {
		t.Fatal(err)
	}
	if info.Domain != "cahut.example.com" || info.ID != before {
		t.Errorf("rename_domain = %+v, want cahut.example.com with id %d", info, before)
	}
	if ok, _ := s.Declared(ctx, "203.0.113.7"); ok {
		t.Error("the old address is still declared")
	}
	c, err := s.OpenCahut(ctx, "cahut.example.com")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	people, err := operation.Do(ctx, c.Registry, account.ListUsers, admin, account.ListUsersInput{})
	if err != nil || len(people.Users) != 1 || people.Users[0].Handle != "olivier" {
		t.Errorf("the people of the moved Cahut = %+v, %v", people, err)
	}
	entries, err := logbook.List(ctx, s.DB.Read, 1)
	if err != nil || len(entries) != 1 || entries[0].Operation != "rename_domain" || entries[0].Target != "203.0.113.7" ||
		entries[0].Detail != "to cahut.example.com" {
		t.Errorf("latest entry = %+v, %v", entries, err)
	}
}
