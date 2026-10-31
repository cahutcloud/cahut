package console

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// response is what get returns of a response.
type response struct {
	status int
	header http.Header
}

func get(t *testing.T, h http.Handler, method, path string) (response, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), method, path, nil))
	resp := rec.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return response{status: resp.StatusCode, header: resp.Header}, string(body)
}

func TestHandler(t *testing.T) {
	h := newHandler(fstest.MapFS{
		"index.html":         {Data: []byte("<!doctype html><title>Cahut</title>")},
		"assets/app-1a2b.js": {Data: []byte("console.log(1)")},
		"robots.txt":         {Data: []byte("User-agent: *\nDisallow: /\n")},
	}, "index.html", nil)
	for _, tt := range []struct {
		path, body, cache string
		status            int
	}{
		{path: "/", body: "<title>Cahut</title>", cache: "no-cache", status: 200},
		{path: "/sign-in", body: "<title>Cahut</title>", cache: "no-cache", status: 200},
		{path: "/rooms/marketing", body: "<title>Cahut</title>", cache: "no-cache", status: 200},
		{path: "/index.html", body: "<title>Cahut</title>", cache: "no-cache", status: 200},
		{path: "/assets/app-1a2b.js", body: "console.log(1)", cache: "public, max-age=31536000, immutable", status: 200},
		{path: "/robots.txt", body: "Disallow", cache: "no-cache", status: 200},
		{path: "/assets/missing.js", status: 404},
		{path: "/../../etc/passwd", body: "<title>Cahut</title>", cache: "no-cache", status: 200},
		// A missing file of the build is a 404; an address of the app with a
		// dot is a page.
		{path: "/favicon.ico", status: 404},
		{path: "/rooms/a.b", body: "<title>Cahut</title>", cache: "no-cache", status: 200},
	} {
		resp, body := get(t, h, http.MethodGet, tt.path)
		if resp.status != tt.status || !strings.Contains(body, tt.body) || tt.cache != "" && resp.header.Get("Cache-Control") != tt.cache {
			t.Errorf("GET %s = %d %q, Cache-Control %q", tt.path, resp.status, body, resp.header.Get("Cache-Control"))
		}
		if csp := resp.header.Get("Content-Security-Policy"); !strings.Contains(csp, "script-src http://example.com/console/assets/;") ||
			!strings.Contains(csp, "frame-ancestors 'none'") {
			t.Errorf("GET %s: Content-Security-Policy = %q", tt.path, csp)
		}
	}
	if resp, _ := get(t, h, http.MethodPost, "/"); resp.status != http.StatusMethodNotAllowed {
		t.Errorf("POST / = %d, want 405", resp.status)
	}
}

// The console of the server serves its own page, and frames the Windows of
// every Cahut of the server. See docs/adr/0058-a-console-for-the-server.md.
func TestServerHandler(t *testing.T) {
	h := newHandler(fstest.MapFS{
		"index.html":         {Data: []byte("<!doctype html><title>Cahut</title>")},
		"server.html":        {Data: []byte("<!doctype html><title>Cahut server</title>")},
		"assets/app-1a2b.js": {Data: []byte("console.log(1)")},
	}, "server.html", func(*http.Request) []string { return []string{"http://example.com", "http://beta.example.com"} })
	for _, path := range []string{"/", "/sign-in", "/cahuts/beta.example.com", "/index.html"} {
		resp, body := get(t, h, http.MethodGet, path)
		if resp.status != http.StatusOK || !strings.Contains(body, "<title>Cahut server</title>") {
			t.Errorf("GET %s = %d %q", path, resp.status, body)
		}
		if csp := resp.header.Get("Content-Security-Policy"); !strings.Contains(csp, "frame-src 'self' http://example.com http://beta.example.com;") ||
			!strings.Contains(csp, "script-src http://example.com/console/assets/;") {
			t.Errorf("GET %s: Content-Security-Policy = %q", path, csp)
		}
	}
	if resp, _ := get(t, h, http.MethodGet, "/assets/missing.js"); resp.status != http.StatusNotFound {
		t.Errorf("GET /assets/missing.js = %d, want 404", resp.status)
	}
	if resp, body := get(t, h, http.MethodGet, "/assets/app-1a2b.js"); resp.status != http.StatusOK || body != "console.log(1)" {
		t.Errorf("GET /assets/app-1a2b.js = %d %q", resp.status, body)
	}
}

func TestNotice(t *testing.T) {
	resp, body := get(t, newHandler(nil, "index.html", nil), http.MethodGet, "/sign-in")
	if resp.status != http.StatusNotFound || !strings.Contains(body, "built without its console") ||
		resp.header.Get("Content-Security-Policy") == "" {
		t.Errorf("GET /sign-in without the console = %d %q", resp.status, body)
	}
}
