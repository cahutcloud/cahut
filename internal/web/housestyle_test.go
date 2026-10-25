package web

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"

	"cahut.cloud/internal/window"
)

func TestHouseStyleServed(t *testing.T) {
	f := setup(t)
	revalidate := func(path, etag string) int {
		t.Helper()
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, f.ts.URL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("If-None-Match", etag)
		resp, err := f.client("localhost", "").Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return resp.StatusCode
	}
	headers := func(path string, resp response, typ string) {
		t.Helper()
		for name, want := range map[string]string{
			"Content-Type": typ, "Access-Control-Allow-Origin": "*", "X-Content-Type-Options": "nosniff",
			"Cache-Control": "no-cache", "Content-Security-Policy": defaultPolicy,
		} {
			if got := resp.Header.Get(name); got != want {
				t.Errorf("GET %s: %s = %q, want %q", path, name, got, want)
			}
		}
		if resp.Header.Get("ETag") == "" {
			t.Errorf("GET %s has no ETag", path)
		}
	}

	resp, css := f.get(t, "localhost", http.MethodGet, "/-/house-style.css")
	if resp.StatusCode != http.StatusOK || !strings.Contains(css, "--hs-primary: #2a5bd7;") {
		t.Fatalf("GET /-/house-style.css = %d %q", resp.StatusCode, css)
	}
	headers("/-/house-style.css", resp, "text/css; charset=utf-8")
	neutral := resp.Header.Get("ETag")
	if code := revalidate("/-/house-style.css", neutral); code != http.StatusNotModified {
		t.Errorf("revalidating the sheet = %d, want 304", code)
	}
	resp, logo := f.get(t, "localhost", http.MethodGet, "/-/house-style/logo")
	if resp.StatusCode != http.StatusOK || !strings.Contains(logo, ">localhost</text>") {
		t.Errorf("logo of the neutral house style = %d %q", resp.StatusCode, logo)
	}
	headers("/-/house-style/logo", resp, "image/svg+xml")

	const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><circle r="5"/></svg>`
	woff2 := base64.StdEncoding.EncodeToString([]byte("wOF2 a font"))
	f.call(t, "localhost", admin, "set_house_style", `{"name": "Example Co", "colors": {"primary": "#0b5fff"},
		"logo": {"content": "`+strings.ReplaceAll(svg, `"`, `\"`)+`"}, "body_font": {"family": "Inter", "files": [{"content": "`+woff2+`"}]}}`)

	resp, css = f.get(t, "localhost", http.MethodGet, "/-/house-style.css")
	if !strings.Contains(css, "--hs-primary: #0b5fff;") || !strings.Contains(css, `url("/-/house-style/fonts/body-400.woff2") format("woff2")`) {
		t.Errorf("the sheet after a change = %q", css)
	}
	if code := revalidate("/-/house-style.css", neutral); code != http.StatusOK || resp.Header.Get("ETag") == neutral {
		t.Errorf("revalidating the sheet after a change = %d, want 200", code)
	}
	resp, logo = f.get(t, "localhost", http.MethodGet, "/-/house-style/logo")
	if logo != svg {
		t.Errorf("logo = %q", logo)
	}
	headers("/-/house-style/logo", resp, "image/svg+xml")
	resp, inverse := f.get(t, "localhost", http.MethodGet, "/-/house-style/logo-inverse")
	if resp.StatusCode != http.StatusOK || !strings.Contains(inverse, ">Example Co</text>") {
		t.Errorf("logo for dark backgrounds, without one = %d %q", resp.StatusCode, inverse)
	}
	resp, font := f.get(t, "localhost", http.MethodGet, "/-/house-style/fonts/body-400.woff2")
	if resp.StatusCode != http.StatusOK || font != "wOF2 a font" {
		t.Errorf("font = %d %q", resp.StatusCode, font)
	}
	headers("/-/house-style/fonts/body-400.woff2", resp, "font/woff2")
	if resp, _ := f.get(t, "localhost", http.MethodGet, "/-/house-style/fonts/body-700.woff2"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("a font that the house style lacks = %d", resp.StatusCode)
	}

	// Each Cahut has its own.
	_, css = f.get(t, "beta.localhost", http.MethodGet, "/-/house-style.css")
	_, logo = f.get(t, "beta.localhost", http.MethodGet, "/-/house-style/logo")
	if !strings.Contains(css, "--hs-primary: #2a5bd7;") || strings.Contains(css, "Inter") || !strings.Contains(logo, ">beta.localhost</text>") {
		t.Errorf("the house style of beta.localhost = %q, %q", css, logo)
	}
	if resp, _ := f.get(t, "beta.localhost", http.MethodGet, "/-/house-style/fonts/body-400.woff2"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("a font of localhost at beta.localhost = %d", resp.StatusCode)
	}
}

// The pages of a Cahut carry the brand of Cahut while its house style is
// neutral, and its house style once it has one. See
// docs/adr/0032-the-page-of-a-room.md.
func TestPagesInTheHouseStyle(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	reader := f.signIn(t, "localhost", "olivier")
	homePage := func() (response, string) {
		return f.do(t, "localhost", http.MethodGet, "/", []*http.Cookie{reader}, nil)
	}
	resp, home := homePage()
	if !strings.Contains(home, `<img class="top__cahut" src="/-/brand/logos/cahut-logo-horizontal.svg" alt="Cahut"`) ||
		!strings.Contains(home, `<span class="top__domain">localhost</span>`) ||
		!strings.Contains(home, `<link rel="stylesheet" href="/-/brand/tokens/cahut-fonts.css">`) ||
		strings.Contains(home, "/-/house-style.css") || strings.Contains(home, "data-house-style") {
		t.Errorf("home of the neutral house style = %q", home)
	}
	if csp := resp.Header.Get("Content-Security-Policy"); csp != pagePolicy("https://localhost/") || !strings.Contains(csp, "font-src 'self'") {
		t.Errorf("the policy of the home = %q", csp)
	}
	if resp, css := f.get(t, "localhost", http.MethodGet, "/-/pages.css"); resp.StatusCode != http.StatusOK || !strings.Contains(css, "var(--hs-primary)") {
		t.Errorf("GET /-/pages.css = %d %q", resp.StatusCode, css)
	}

	f.call(t, "localhost", admin, "set_house_style", `{"name": "Example & Co", "logo": {"content": "<svg xmlns=\"http://www.w3.org/2000/svg\"/>"}}`)
	_, home = homePage()
	if !strings.Contains(home, `<img class="top__logo" src="/-/house-style/logo" alt="Example &amp; Co">`) ||
		!strings.Contains(home, `<h1 class="hero__title">Example &amp; Co</h1>`) ||
		!strings.Contains(home, `<html lang="en" data-theme="light" data-house-style>`) ||
		!strings.Contains(home, `<link rel="stylesheet" href="/-/house-style.css">`) ||
		strings.Contains(home, "cahut-logo") || strings.Contains(home, "cahut-fonts.css") {
		t.Errorf("home with a logo = %q", home)
	}
	publishDeck(t, f, olivier)
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key", "key": "correct horse"}`)
	resp, page := f.get(t, "localhost", http.MethodGet, "/olivier/q3-review/")
	if resp.StatusCode != http.StatusUnauthorized || !strings.Contains(page, `alt="Example &amp; Co"`) {
		t.Errorf("the page of a Key = %d %q", resp.StatusCode, page)
	}
	if _, page := f.get(t, "beta.localhost", http.MethodGet, "/nope/"); !strings.Contains(page, `<span class="top__domain">beta.localhost</span>`) {
		t.Errorf("a page of beta.localhost = %q", page)
	}

	// Back to the neutral house style, the pages carry the brand of Cahut again.
	f.call(t, "localhost", admin, "set_house_style", `{"neutral": true}`)
	if _, home := homePage(); !strings.Contains(home, "cahut-logo-horizontal.svg") {
		t.Errorf("home of the neutral house style again = %q", home)
	}

	// What the house style says to AIs, and whether they use it, leave the
	// brand of Cahut; a name, colors, fonts or a logo take it off. See
	// docs/adr/0033-the-brand-of-cahut-until-the-house-style-has-a-look.md.
	for _, tt := range []struct {
		change string
		cahut  bool
	}{
		{`"guidance": "Write short sentences.", "active": false`, true},
		{`"name": "Example"`, false},
		{`"colors": {"primary": "#c2410c"}`, false},
		{`"chart_colors": ["#c2410c", "#0f766e"]`, false},
		{`"body_font": {"family": "Georgia", "fallback": "serif"}`, false},
		{`"logo": {"content": "<svg xmlns=\"http://www.w3.org/2000/svg\"/>"}`, false},
	} {
		f.call(t, "localhost", admin, "set_house_style", `{"neutral": true, `+tt.change+`}`)
		_, home := homePage()
		if cahut := strings.Contains(home, "cahut-logo-horizontal.svg") && !strings.Contains(home, "data-house-style"); cahut != tt.cahut {
			t.Errorf("with %s, the pages carry the brand of Cahut: %v, want %v", tt.change, cahut, tt.cahut)
		}
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

// A Window links the house style of its Cahut from the root of the domain,
// and declares nothing for it: its policy admits its Cahut's origin.
func TestWindowsLoadTheHouseStyle(t *testing.T) {
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

func TestDeckEngineAndTemplates(t *testing.T) {
	f := setup(t)
	resp, js := f.get(t, "localhost", http.MethodGet, "/-/deck.js")
	if resp.StatusCode != http.StatusOK || !strings.Contains(js, `document.querySelector(".hs-deck")`) ||
		resp.Header.Get("Content-Type") != "text/javascript; charset=utf-8" || resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("GET /-/deck.js = %d %v", resp.StatusCode, resp.Header)
	}
	f.call(t, "localhost", admin, "set_house_style", `{"name": "Example Co"}`)
	for _, kind := range []string{"document", "deck", "dashboard"} {
		resp, page := f.get(t, "localhost", http.MethodGet, "/-/templates/"+kind)
		csp := resp.Header.Get("Content-Security-Policy")
		if resp.StatusCode != http.StatusOK || !strings.Contains(page, `class="hs-`+kind+`"`) || !strings.Contains(page, "Example Co") ||
			!strings.HasPrefix(csp, "sandbox allow-scripts ") || !strings.Contains(csp, "connect-src https://localhost;") {
			t.Errorf("GET /-/templates/%s = %d, policy %q", kind, resp.StatusCode, csp)
		}
	}
	for _, kind := range []string{"site", "poster"} {
		if resp, _ := f.get(t, "localhost", http.MethodGet, "/-/templates/"+kind); resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET /-/templates/%s = %d, want 404", kind, resp.StatusCode)
		}
	}
	if !strings.Contains(instructions, "call get_house_style") || !strings.Contains(instructions, "/-/house-style.css") {
		t.Errorf("the instructions of the MCP server do not send AIs to the house style: %q", instructions)
	}
}

func TestSkillServed(t *testing.T) {
	f := setup(t)
	resp, body := f.get(t, "localhost", http.MethodGet, SkillPath)
	if resp.StatusCode != http.StatusOK || resp.Header.Get("Content-Type") != "application/zip" ||
		resp.Header.Get("Content-Disposition") != `attachment; filename="cahut-skill.zip"` || !strings.HasPrefix(body, "PK") {
		t.Errorf("GET %s = %d %v", SkillPath, resp.StatusCode, resp.Header)
	}
}
