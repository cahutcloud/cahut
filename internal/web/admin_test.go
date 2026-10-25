package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
)

// serviceToken creates a service token in the Cahut at domain.
func serviceToken(t *testing.T, f *fixture, domain string) string {
	t.Helper()
	st, err := operation.Do(context.Background(), f.open(t, domain).Registry, account.CreateServiceToken, admin,
		account.CreateServiceTokenInput{Name: "account"})
	if err != nil {
		t.Fatal(err)
	}
	return st.Token
}

func TestAdminAPI(t *testing.T) {
	f := setup(t)
	api := httptest.NewServer(f.web.AdminHandler())
	t.Cleanup(api.Close)
	svc := serviceToken(t, f, "localhost")
	beta := serviceToken(t, f, "beta.localhost")
	camille := activate(t, f, f.ts, "camille", true)

	r := call(t, f, api, http.MethodPost, "localhost", "/api/check_handle", svc, `{"handle": "ana"}`)
	if r.status != http.StatusOK || r.body["available"] != true {
		t.Fatalf("check_handle: %d %s", r.status, r.raw)
	}
	if r.header.Get("Content-Type") != "application/json" || r.header.Get("Cache-Control") != "no-store" {
		t.Errorf("headers = %v", r.header)
	}
	// A service creates a person at once, and gets their link.
	r = call(t, f, api, http.MethodPost, "localhost", "/api/create_user", svc,
		`{"handle": "ana", "name": "Ana", "email": "ana@example.org", "send_activation_email": false}`)
	activation, _ := r.body["activation"].(map[string]any)
	if r.status != http.StatusOK || activation["code"] == nil || activation["emailed"] != nil {
		t.Fatalf("create_user: %d %s", r.status, r.raw)
	}
	r = call(t, f, api, http.MethodPost, "localhost", "/api/get_user", svc, `{"user": "ana"}`)
	if r.status != http.StatusOK || r.body["email"] != "ana@example.org" || r.body["status"] != "pending" {
		t.Errorf("get_user: %d %s", r.status, r.raw)
	}
	r = call(t, f, api, http.MethodPost, "localhost", "/api/check_handle", svc, `{"handle": "ana"}`)
	if r.status != http.StatusOK || r.body["available"] != false || r.body["reason"] != "taken" {
		t.Errorf("check_handle of a taken handle: %d %s", r.status, r.raw)
	}

	for _, tt := range []struct {
		name, method, host, path, token, body string
		want                                  int
		says                                  string
	}{
		{name: "no token", method: http.MethodPost, host: "localhost", path: "/api/get_cahut", want: 401, says: "service tokens"},
		{name: "a personal token", method: http.MethodPost, host: "localhost", path: "/api/get_cahut", token: f.token, want: 401},
		{name: "a session", method: http.MethodPost, host: "localhost", path: "/api/get_cahut", token: camille, want: 401},
		{name: "the token of another Cahut", method: http.MethodPost, host: "localhost", path: "/api/get_cahut", token: beta, want: 401},
		{name: "an operation closed to services", method: http.MethodPost, host: "localhost", path: "/api/list_users", token: svc, want: 403,
			says: "not open to service tokens"},
		{name: "an operation of the server", method: http.MethodPost, host: "localhost", path: "/api/create_service_token", token: svc,
			body: `{"name": "x"}`, want: 403},
		{name: "an administrator", method: http.MethodPost, host: "localhost", path: "/api/create_user", token: svc,
			body: `{"handle": "eve", "name": "Eve", "admin": true}`, want: 403},
		{name: "no operation", method: http.MethodPost, host: "localhost", path: "/api/delete_everything", token: svc, want: 404},
		{name: "the console's route", method: http.MethodPost, host: "localhost", path: "/console/api/get_cahut", token: svc, want: 404},
		{name: "a GET", method: http.MethodGet, host: "localhost", path: "/api/get_cahut", token: svc, want: 405},
		{name: "a domain of no Cahut", method: http.MethodPost, host: "nope.localhost", path: "/api/get_cahut", token: svc, want: 421,
			says: "the host of the request names the Cahut"},
	} {
		r := call(t, f, api, tt.method, tt.host, tt.path, tt.token, tt.body)
		if r.status != tt.want || !strings.Contains(r.raw, tt.says) {
			t.Errorf("%s: %d %s, want %d saying %q", tt.name, r.status, r.raw, tt.want, tt.says)
		}
	}
	if r := call(t, f, api, http.MethodPost, "localhost", "/api/get_cahut", "", ``); r.header.Get("WWW-Authenticate") != "Bearer" {
		t.Errorf("WWW-Authenticate = %q", r.header.Get("WWW-Authenticate"))
	}

	// A service token opens neither the console's API nor MCP.
	if r := call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/get_cahut", svc, ``); r.status != http.StatusUnauthorized {
		t.Errorf("the console's API with a service token: %d %s", r.status, r.raw)
	}
	if r := call(t, f, f.ts, http.MethodPost, "localhost", "/mcp", svc, `{}`); r.status != http.StatusUnauthorized {
		t.Errorf("MCP with a service token: %d %s", r.status, r.raw)
	}
	// Nor does the admin API answer a client off the machine.
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "http://localhost/api/get_cahut", strings.NewReader(`{}`))
	req.RemoteAddr = "203.0.113.7:4242"
	req.Header.Set("Authorization", "Bearer "+svc)
	rec := httptest.NewRecorder()
	f.web.AdminHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("a client off the machine: %d %s", rec.Code, rec.Body)
	}
}

func TestCheckAdminAPIAddr(t *testing.T) {
	for addr, ok := range map[string]bool{
		"127.0.0.1:8090": true, "[::1]:8090": true, "localhost:8090": true, "127.0.0.2:8090": true,
		":8090": false, "0.0.0.0:8090": false, "[::]:8090": false, "203.0.113.7:8090": false, "127.0.0.1": false,
		"cahut.example.com:8090": false,
	} {
		if err := CheckAdminAPIAddr(addr); (err == nil) != ok {
			t.Errorf("CheckAdminAPIAddr(%q) = %v, want ok %v", addr, err, ok)
		}
	}
}
