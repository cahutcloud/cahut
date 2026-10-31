package address

import (
	"strings"
	"testing"
)

func TestValidName(t *testing.T) {
	for _, s := range []string{"q3", "q3-review", "olivier", "a", strings.Repeat("a", 63)} {
		if !ValidName(s) {
			t.Errorf("ValidName(%q) = false", s)
		}
	}
	for _, s := range []string{"", "-", "-q3", "q3-", "Q3", "q3_review", "q3.review", "é", strings.Repeat("a", 64)} {
		if ValidName(s) {
			t.Errorf("ValidName(%q) = true", s)
		}
	}
}

func TestReserved(t *testing.T) {
	if !Reserved("mcp") || Reserved("marketing") {
		t.Error("Reserved is wrong about mcp or marketing")
	}
	// The console of the server lives under /server/ of every domain. See
	// docs/adr/0058-a-console-for-the-server.md.
	if !Reserved("server") {
		t.Error("Reserved is wrong about server")
	}
}

func TestValidDomain(t *testing.T) {
	for _, s := range []string{"localhost", "beta.localhost", "cahut.example.com", "example.org", "xn--caf-dma.fr"} {
		if !ValidDomain(s) {
			t.Errorf("ValidDomain(%q) = false", s)
		}
	}
	for _, s := range []string{"", "Example.org", "example.org.", ".example.org", "*.example.org", "203.0.113.7",
		"example..org", "-example.org", "example.org:443", strings.Repeat("a.", 127) + "org"} {
		if ValidDomain(s) {
			t.Errorf("ValidDomain(%q) = true", s)
		}
	}
}

func TestSlug(t *testing.T) {
	tests := map[string]string{
		"Revue du 3e trimestre":          "revue-du-3e-trimestre",
		"  Présentation : Œuvre ":        "presentation-oeuvre",
		"Straße & Ölmühle":               "strasse-olmuhle",
		"日本語":                            "window",
		"---":                            "window",
		strings.Repeat("abcdefghij ", 6): "abcdefghij-abcdefghij-abcdefghij-abcdefg",
	}
	for title, want := range tests {
		if got := Slug(title, "window"); got != want {
			t.Errorf("Slug(%q) = %q, want %q", title, got, want)
		}
	}
}
