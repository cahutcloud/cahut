package web

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"cahut.cloud/internal/window"
)

// The pages of a Cahut carry the brand of Cahut, light or dark as the
// reader's system is. See docs/adr/0032-the-page-of-a-room.md and
// docs/adr/0049-no-house-style.md.
func TestPagesInTheBrandOfCahut(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	reader := f.signIn(t, "localhost", "olivier")
	resp, body := f.do(t, "localhost", http.MethodGet, "/olivier/", []*http.Cookie{reader}, nil)
	if !strings.Contains(body, `<img class="top__cahut" src="/-/brand/logos/cahut-logo-horizontal.svg" alt="Cahut"`) ||
		!strings.Contains(body, `<span class="top__domain">localhost</span>`) ||
		!strings.Contains(body, `<link rel="stylesheet" href="/-/brand/tokens/cahut-fonts.css">`) ||
		!strings.Contains(body, `<link rel="icon" href="/-/brand/logos/cahut-favicon.svg" type="image/svg+xml">`) {
		t.Errorf("the page of a Room = %q", body)
	}
	if csp := resp.Header.Get("Content-Security-Policy"); csp != pagePolicy("https://localhost/") || !strings.Contains(csp, "font-src 'self'") {
		t.Errorf("the policy of the page of a Room = %q", csp)
	}
	publishDeck(t, f, olivier)
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key", "key": "correct horse"}`)
	resp, page := f.get(t, "localhost", http.MethodGet, "/olivier/q3-review/")
	if resp.StatusCode != http.StatusUnauthorized || !strings.Contains(page, "cahut-logo-horizontal.svg") {
		t.Errorf("the page of a Key = %d %q", resp.StatusCode, page)
	}
	if _, page := f.get(t, "beta.localhost", http.MethodGet, "/nope/"); !strings.Contains(page, `<span class="top__domain">beta.localhost</span>`) {
		t.Errorf("a page of beta.localhost = %q", page)
	}
}

// Cahut serves no style sheet, deck engine or template for Windows, and
// offers AIs no tool for them. See docs/adr/0049-no-house-style.md.
func TestNoHouseStyle(t *testing.T) {
	f := setup(t)
	for _, p := range []string{"/-/house-style.css", "/-/house-style/logo", "/-/house-style/logo-inverse",
		"/-/house-style/fonts/body-400.woff2", "/-/deck.js", "/-/templates/deck"} {
		if resp, _ := f.get(t, "localhost", http.MethodGet, p); resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", p, resp.StatusCode)
		}
	}
	st, err := f.web.site(context.Background(), "localhost")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"get_house_style", "set_house_style", "rollback_house_style"} {
		if _, ok := st.cahut.Registry.Lookup(name); ok {
			t.Errorf("the registry still has %s", name)
		}
	}
	if strings.Contains(instructions, "house style") || strings.Contains(instructions, "/-/") {
		t.Errorf("the instructions of the MCP server send AIs under /-/: %q", instructions)
	}
}

// The files of the brand of Cahut that its pages use are served as they
// are, and nothing else.
func TestBrandFiles(t *testing.T) {
	f := setup(t)
	tests := []struct {
		path, typ string
		status    int
	}{
		{"/-/brand/tokens/cahut-tokens.css", "text/css; charset=utf-8", http.StatusOK},
		{"/-/brand/tokens/cahut-fonts.css", "text/css; charset=utf-8", http.StatusOK},
		{"/-/brand/typographie/polices/web/plus-jakarta-sans-latin-variable.woff2", "font/woff2", http.StatusOK},
		{"/-/brand/typographie/polices/OFL-PlusJakartaSans.txt", "text/plain; charset=utf-8", http.StatusOK},
		{"/-/brand/logos/cahut-logo-horizontal.svg", "image/svg+xml", http.StatusOK},
		{"/-/brand/logos/cahut-icone-app.svg", "image/svg+xml", http.StatusOK},
		{"/-/brand/logos/favicon/favicon.ico", "image/x-icon", http.StatusOK},
		{"/-/brand/logos/favicon/apple-touch-icon.png", "image/png", http.StatusOK},
		{"/-/brand/logos/favicon/icon-512.png", "image/png", http.StatusOK},
		{"/-/brand/tokens/", "", http.StatusNotFound},
		{"/-/brand/", "", http.StatusNotFound},
		{"/-/brand/README.md", "", http.StatusNotFound},
		{"/-/brand/charte.go", "", http.StatusNotFound},
	}
	for _, tt := range tests {
		resp, body := f.get(t, "localhost", http.MethodGet, tt.path)
		if resp.StatusCode != tt.status {
			t.Errorf("GET %s = %d, want %d", tt.path, resp.StatusCode, tt.status)
			continue
		}
		if tt.status != http.StatusOK {
			continue
		}
		if got := resp.Header.Get("Content-Type"); got != tt.typ || body == "" || resp.Header.Get("ETag") == "" {
			t.Errorf("GET %s = %q, %d bytes, ETag %q", tt.path, got, len(body), resp.Header.Get("ETag"))
		}
	}
	// The fonts sheet finds the fonts beside it.
	_, css := f.get(t, "localhost", http.MethodGet, "/-/brand/tokens/cahut-fonts.css")
	if !strings.Contains(css, `url("../typographie/polices/web/plus-jakarta-sans-latin-variable.woff2")`) {
		t.Errorf("the fonts sheet = %q", css)
	}
}

// The root of a Cahut serves the icon of Cahut, where browsers and the
// services that draw the icon of a site look for it, as its pages carry
// its brand. The console names its icons under /-/brand/. See
// docs/adr/0040-the-icon-of-a-cahut.md.
func TestRootIcons(t *testing.T) {
	f := setup(t)
	for p, file := range rootIcons {
		resp, body := f.get(t, "localhost", http.MethodGet, p)
		if resp.StatusCode != http.StatusOK || body != string(brandAssets()[file].Content) {
			t.Errorf("GET %s = %d, %d bytes, not those of %s", p, resp.StatusCode, len(body), file)
		}
		if resp, _ := f.get(t, "localhost", http.MethodGet, BrandPath+file); resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s%s = %d", BrandPath, file, resp.StatusCode)
		}
	}
	if resp, _ := f.get(t, "localhost", http.MethodGet, "/favicon.ico"); resp.Header.Get("Content-Type") != "image/x-icon" {
		t.Errorf("GET /favicon.ico is a %q", resp.Header.Get("Content-Type"))
	}
}

// A Window loads the files of its Cahut from the origin of its Cahut, and
// declares nothing for them: its policy admits that origin.
func TestWindowsLoadFromTheirCahut(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	publishDeck(t, f, olivier)
	f.setAccess(t, "localhost", "q3-review", window.AccessPublic)
	resp, _ := f.get(t, "localhost", http.MethodGet, "/olivier/q3-review/")
	csp := resp.Header.Get("Content-Security-Policy")
	for _, want := range []string{"style-src https://localhost ", "font-src https://localhost ", "img-src https://localhost "} {
		if !strings.Contains(csp, want) {
			t.Errorf("the policy of a Window lacks %q: %q", want, csp)
		}
	}
}

func TestSkillServed(t *testing.T) {
	f := setup(t)
	resp, body := f.get(t, "localhost", http.MethodGet, SkillPath)
	if resp.StatusCode != http.StatusOK || resp.Header.Get("Content-Type") != "application/zip" ||
		resp.Header.Get("Content-Disposition") != `attachment; filename="cahut-skill.zip"` || !strings.HasPrefix(body, "PK") ||
		resp.Header.Get("Access-Control-Allow-Origin") != "*" || resp.Header.Get("Cache-Control") != "no-cache" || resp.Header.Get("ETag") == "" {
		t.Errorf("GET %s = %d %v", SkillPath, resp.StatusCode, resp.Header)
	}
}
