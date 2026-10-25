package cahut

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
)

func (l *limits) setLegal(in settings.LegalLinks) (settings.LegalLinks, error) {
	l.t.Helper()
	return operation.Do(l.ctx, l.c.Registry, settings.SetLegalLinks, l.admin, settings.SetLegalLinksInput{LegalLinks: in})
}

func (l *limits) signInPage() account.SignInPage {
	l.t.Helper()
	p, err := operation.Do(l.ctx, l.c.Registry, account.DescribeSignIn, operation.Actor{Kind: operation.Visitor, Via: "api"}, account.DescribeSignInInput{})
	if err != nil {
		l.t.Fatalf("describe_sign_in: %v", err)
	}
	return p
}

func (l *limits) lastEntry() logbook.Entry {
	l.t.Helper()
	entries, err := logbook.List(l.ctx, l.c.DB.Read, 1)
	if err != nil {
		l.t.Fatal(err)
	}
	return entries[0]
}

// The administrators of the server link the legal pages of a Cahut, which
// get_cahut gives to people and AIs, and describe_sign_in to whoever has not
// signed in. See docs/adr/0042-the-legal-pages-of-a-cahut.md.
func TestSetLegalLinks(t *testing.T) {
	l := openLimits(t)
	olivier := l.person("olivier", true)

	// A Cahut links nothing until they say so.
	if a := l.about(olivier); a.Legal != (settings.LegalLinks{}) {
		t.Errorf("the links of a new Cahut = %+v", a.Legal)
	}
	if p := l.signInPage(); p.Legal != (settings.LegalLinks{}) {
		t.Errorf("describe_sign_in of a new Cahut = %+v", p.Legal)
	}

	set := settings.LegalLinks{
		LegalPages: settings.LegalPages{
			LegalNotice: " https://host.example/legal ",
			Terms:       "https://host.example/terms",
			Privacy:     "https://host.example/privacy",
			Support:     "https://help.host.example/",
		},
		FR: &settings.LegalPages{LegalNotice: "https://host.example/fr/mentions-legales", Terms: "https://host.example/fr/conditions"},
	}
	out, err := l.setLegal(set)
	if err != nil {
		t.Fatalf("set_legal_links: %v", err)
	}
	if out.LegalNotice != "https://host.example/legal" || out.Support != "https://help.host.example/" || out.FR == nil ||
		out.FR.Terms != "https://host.example/fr/conditions" || out.FR.Privacy != "" {
		t.Errorf("set_legal_links = %+v %+v", out, out.FR)
	}
	if a := l.about(olivier); a.Legal.Terms != "https://host.example/terms" || a.Legal.FR == nil || a.Legal.FR.LegalNotice == "" {
		t.Errorf("get_cahut after set_legal_links = %+v", a.Legal)
	}
	if p := l.signInPage(); p.Legal.Privacy != "https://host.example/privacy" || p.Legal.FR == nil {
		t.Errorf("describe_sign_in after set_legal_links = %+v", p.Legal)
	}
	if e := l.lastEntry(); e.Operation != "set_legal_links" || e.Target != "legal links" ||
		!strings.HasPrefix(e.Detail, "legal_notice none → https://host.example/legal, terms none → https://host.example/terms, ") ||
		!strings.HasSuffix(e.Detail, ", fr.terms none → https://host.example/fr/conditions") {
		t.Errorf("logbook = %+v", e)
	}

	// Each page in a language: its own address, or else the other one.
	var got []string
	for _, k := range out.In("fr") {
		got = append(got, k.Page+" "+k.URL)
	}
	if !slices.Equal(got, []string{
		"legal_notice https://host.example/fr/mentions-legales", "terms https://host.example/fr/conditions",
		"privacy https://host.example/privacy", "support https://help.host.example/",
	}) {
		t.Errorf("the links in French = %q", got)
	}
	if got := out.In(""); len(got) != 4 || got[0].URL != "https://host.example/legal" || got[1].URL != "https://host.example/terms" {
		t.Errorf("the links in English = %+v", got)
	}

	// Setting them again replaces them: a link left out is removed.
	out, err = l.setLegal(settings.LegalLinks{LegalPages: settings.LegalPages{Terms: "https://host.example/terms"}})
	if err != nil || out != (settings.LegalLinks{LegalPages: settings.LegalPages{Terms: "https://host.example/terms"}}) {
		t.Errorf("set_legal_links again = %+v %v", out, err)
	}
	if got := out.In("fr"); len(got) != 1 || got[0].Page != "terms" || got[0].URL != "https://host.example/terms" {
		t.Errorf("the links in French, without French addresses = %+v", got)
	}
	if e := l.lastEntry(); !strings.Contains(e.Detail, "privacy https://host.example/privacy → none") ||
		strings.Contains(e.Detail, "terms https://host.example/terms") {
		t.Errorf("logbook = %+v", e)
	}
	if _, err := l.setLegal(settings.LegalLinks{LegalPages: settings.LegalPages{Terms: "https://host.example/terms"}}); err != nil {
		t.Fatal(err)
	}
	if e := l.lastEntry(); e.Detail != "unchanged" {
		t.Errorf("logbook of the same links = %+v", e)
	}

	// Only the address of a page in HTTPS, and a page in French only with
	// its address for the other languages.
	for _, bad := range []struct {
		links settings.LegalLinks
		code  string
		link  string
	}{
		{settings.LegalLinks{LegalPages: settings.LegalPages{Terms: "http://host.example/terms"}}, "legal_link_invalid", "terms"},
		{settings.LegalLinks{LegalPages: settings.LegalPages{Support: "javascript:alert(1)"}}, "legal_link_invalid", "support"},
		{settings.LegalLinks{LegalPages: settings.LegalPages{Privacy: "https://ana@host.example/"}}, "legal_link_invalid", "privacy"},
		{settings.LegalLinks{LegalPages: settings.LegalPages{Privacy: "https:///privacy"}}, "legal_link_invalid", "privacy"},
		{settings.LegalLinks{LegalPages: settings.LegalPages{Privacy: "https://host.example/a page"}}, "legal_link_invalid", "privacy"},
		{settings.LegalLinks{LegalPages: settings.LegalPages{Privacy: "//host.example/privacy"}}, "legal_link_invalid", "privacy"},
		{settings.LegalLinks{LegalPages: settings.LegalPages{Terms: "https://host.example/" + strings.Repeat("x", 2000)}}, "legal_link_invalid", "terms"},
		{settings.LegalLinks{FR: &settings.LegalPages{Terms: "ftp://host.example/fr"}}, "legal_link_invalid", "fr.terms"},
		{settings.LegalLinks{FR: &settings.LegalPages{Privacy: "https://host.example/fr/confidentialite"}}, "legal_link_in_french_only", "fr.privacy"},
	} {
		_, err := l.setLegal(bad.links)
		var r *operation.Refusal
		if !errors.Is(err, operation.ErrInvalid) || !errors.As(err, &r) || r.Code != bad.code || !strings.HasPrefix(err.Error(), "invalid input: "+bad.link+": ") {
			t.Errorf("set_legal_links %+v: err = %v, want %s for %s", bad.links, err, bad.code, bad.link)
		}
	}
	if a := l.about(olivier); a.Legal.Terms != "https://host.example/terms" || a.Legal.Privacy != "" {
		t.Errorf("the links after refused calls = %+v", a.Legal)
	}

	// Only the administrators of the server set them: not those of the
	// Cahut, and no AI, which is never offered the operation.
	_, err = operation.Do(l.ctx, l.c.Registry, settings.SetLegalLinks, olivier, settings.SetLegalLinksInput{})
	l.refused(err, "administrators of the server")
	info, _ := l.c.Registry.Lookup("set_legal_links")
	if info.ForAIs(true, false) {
		t.Error("set_legal_links is offered to the AIs of an administrator of the Cahut")
	}
}
