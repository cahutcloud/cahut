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
	for _, domain := range []string{"cahut.example.com", "*.example.com", "203.0.113.7", ""} {
		if _, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{Domain: domain}); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("add_domain %q: err = %v, want ErrInvalid", domain, err)
		}
	}

	list, err := operation.Do(ctx, s.Registry, ListDomains, admin, ListDomainsInput{})
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

	entries, err := logbook.List(ctx, s.DB.Read, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 6 || entries[5].Operation != "add_domain" || entries[5].Target != "cahut.example.com" {
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
