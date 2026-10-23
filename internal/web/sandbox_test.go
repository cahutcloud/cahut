package web

import (
	"net/http"
	"strings"
	"testing"
)

// TestSandboxByDefault walks the routes of port 443: every response is
// sandboxed, but those of the pages of Cahut, which have their own strict
// policy. See docs/adr/0020-console-under-a-path-of-port-443.md.
func TestSandboxByDefault(t *testing.T) {
	f := setup(t)
	publishDeck(t, f, f.person(t, "localhost", "olivier"))
	f.setAccess(t, "localhost", "q3-review", "public")

	pages := []string{
		"/",
		"/olivier/",
		"/nobody/",
		"/olivier/missing/",
		"/-/session?code=nope",
	}
	sandboxed := []struct{ host, method, path string }{
		{"localhost", http.MethodGet, "/olivier/q3-review/"},
		{"localhost", http.MethodGet, "/olivier/q3-review/index.html"},
		{"localhost", http.MethodGet, "/olivier/q3-review/css/deck.css"},
		{"localhost", http.MethodGet, "/olivier/q3-review/missing.html"},
		{"localhost", http.MethodGet, "/olivier/q3-review/css/"},
		{"localhost", http.MethodGet, "/olivier/q3-review"},
		{"localhost", http.MethodGet, "/olivier"},
		{"localhost", http.MethodPost, "/olivier/q3-review/"},
		{"localhost", http.MethodGet, "/-/nope"},
		{"localhost", http.MethodGet, "/-/w/garbage/index.html"},
		{"localhost", http.MethodGet, "/-/sign-in?to=%2F"},
		{"localhost", http.MethodGet, "/-/healthz"},
		{"localhost", http.MethodGet, "/mcp"},
		{"localhost", http.MethodGet, "/.well-known/oauth-authorization-server"},
		{"localhost", http.MethodGet, "/.well-known/openid-configuration"},
		{"localhost", http.MethodGet, "/oauth/token"},
		{"localhost", http.MethodPost, "/oauth/token"},
		{"nope.localhost", http.MethodGet, "/olivier/q3-review/"},
	}

	for _, p := range pages {
		resp, _ := f.do(t, "localhost", http.MethodGet, p, nil, nil)
		if csp := resp.Header.Get("Content-Security-Policy"); csp != pagePolicy {
			t.Errorf("GET %s = %d with the policy %q, want the policy of a page", p, resp.StatusCode, csp)
		}
	}
	for _, tt := range sandboxed {
		resp, _ := f.do(t, tt.host, tt.method, tt.path, nil, nil)
		if csp := resp.Header.Get("Content-Security-Policy"); !strings.HasPrefix(csp, "sandbox") || strings.Contains(csp, "allow-same-origin") {
			t.Errorf("%s %s%s = %d with the policy %q, want a sandbox", tt.method, tt.host, tt.path, resp.StatusCode, csp)
		}
	}
}
