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
	if resp.StatusCode != http.StatusOK || inverse != svg {
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

func TestPagesInTheHouseStyle(t *testing.T) {
	f := setup(t)
	resp, home := f.get(t, "localhost", http.MethodGet, "/")
	if !strings.Contains(home, `<a class="top__home" href="/">localhost</a>`) ||
		!strings.Contains(home, `<link rel="stylesheet" href="/-/house-style.css">`+"\n"+`<link rel="stylesheet" href="/-/pages.css">`) {
		t.Errorf("home of the neutral house style = %q", home)
	}
	if csp := resp.Header.Get("Content-Security-Policy"); csp != pagePolicy || !strings.Contains(csp, "font-src 'self'") {
		t.Errorf("the policy of the home = %q", csp)
	}
	if resp, css := f.get(t, "localhost", http.MethodGet, "/-/pages.css"); resp.StatusCode != http.StatusOK || !strings.Contains(css, "var(--hs-primary)") {
		t.Errorf("GET /-/pages.css = %d %q", resp.StatusCode, css)
	}

	f.call(t, "localhost", admin, "set_house_style", `{"name": "Example & Co", "logo": {"content": "<svg xmlns=\"http://www.w3.org/2000/svg\"/>"}}`)
	_, home = f.get(t, "localhost", http.MethodGet, "/")
	if !strings.Contains(home, `<img class="top__logo" src="/-/house-style/logo" alt="Example &amp; Co">`) || !strings.Contains(home, "<h1>Example &amp; Co</h1>") {
		t.Errorf("home with a logo = %q", home)
	}
	olivier := f.person(t, "localhost", "olivier")
	publishDeck(t, f, olivier)
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key", "key": "correct horse"}`)
	resp, page := f.get(t, "localhost", http.MethodGet, "/olivier/q3-review/")
	if resp.StatusCode != http.StatusUnauthorized || !strings.Contains(page, `alt="Example &amp; Co"`) {
		t.Errorf("the page of a Key = %d %q", resp.StatusCode, page)
	}
	if _, home := f.get(t, "beta.localhost", http.MethodGet, "/"); !strings.Contains(home, `href="/">beta.localhost</a>`) {
		t.Errorf("home of beta.localhost = %q", home)
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
