package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
)

const password = "correct horse battery staple"

// apiResponse is what call returns of a response of the API.
type apiResponse struct {
	status int
	header http.Header
	body   map[string]any
	raw    string
}

// call sends a request to the server, as the console of the Cahut at host.
func call(t *testing.T, f *fixture, console *httptest.Server, method, host, path, token, body string) apiResponse {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, console.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := f.client(host, token).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	r := apiResponse{status: resp.StatusCode, header: resp.Header, raw: string(raw)}
	_ = json.Unmarshal(raw, &r.body)
	return r
}

// activate creates a person of the Cahut at localhost and activates their
// account over the API. It returns their session token.
func activate(t *testing.T, f *fixture, console *httptest.Server, handle string, isAdmin bool) string {
	t.Helper()
	c := f.open(t, "localhost")
	u, err := operation.Do(context.Background(), c.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: handle, Name: handle, Admin: isAdmin})
	if err != nil {
		t.Fatal(err)
	}
	r := call(t, f, console, http.MethodPost, "localhost", "/console/api/activate_account", "",
		`{"code": "`+u.Activation.Code+`", "password": "`+password+`"}`)
	if r.status != http.StatusOK {
		t.Fatalf("activate_account: %d %s", r.status, r.raw)
	}
	return r.body["token"].(string)
}

func TestAPI(t *testing.T) {
	f := setup(t)
	console := f.ts
	camille := activate(t, f, console, "camille", false)

	r := call(t, f, console, http.MethodPost, "localhost", "/console/api/get_account", camille, ``)
	if r.status != http.StatusOK || r.body["handle"] != "camille" || r.body["role"] != "member" {
		t.Fatalf("get_account: %d %s", r.status, r.raw)
	}
	for name, want := range map[string]string{
		"Content-Type":            "application/json",
		"Cache-Control":           "no-store",
		"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'",
		"X-Content-Type-Options":  "nosniff",
	} {
		if got := r.header.Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}

	for _, tt := range []struct {
		name, method, host, path, token, body string
		want                                  int
		says                                  string
	}{
		{name: "no session", method: http.MethodPost, host: "localhost", path: "/console/api/get_account", want: 401, says: "sign in first"},
		{name: "a session at another domain", method: http.MethodPost, host: "beta.localhost", path: "/console/api/get_account", token: camille, want: 401},
		{name: "a personal token", method: http.MethodPost, host: "localhost", path: "/console/api/get_account", token: f.token, want: 401, says: "AIs use MCP"},
		{name: "a GET", method: http.MethodGet, host: "localhost", path: "/console/api/get_account", token: camille, want: 405},
		{name: "a preflight", method: http.MethodOptions, host: "localhost", path: "/console/api/get_account", want: 405},
		{name: "no operation", method: http.MethodPost, host: "localhost", path: "/console/api/delete_everything", token: camille, want: 404},
		{name: "no route", method: http.MethodPost, host: "localhost", path: "/console/api/a/b", token: camille, want: 404},
		{name: "a bad input", method: http.MethodPost, host: "localhost", path: "/console/api/create_room", token: camille, body: `{"title": " "}`, want: 400},
		{name: "an unknown field", method: http.MethodPost, host: "localhost", path: "/console/api/get_account", token: camille, body: `{"x": 1}`, want: 400},
		{name: "a refusal", method: http.MethodPost, host: "localhost", path: "/console/api/create_user", token: camille, body: `{"handle": "bruno", "name": "Bruno"}`, want: 403},
		{name: "a missing record", method: http.MethodPost, host: "localhost", path: "/console/api/get_room", token: camille, body: `{"room": "nope"}`, want: 404},
		{name: "a large body", method: http.MethodPost, host: "localhost", path: "/console/api/get_account", token: camille, body: `{"x": "` + strings.Repeat("x", MaxAPIRequestSize) + `"}`, want: 413},
		{name: "the server, as a person", method: http.MethodPost, host: "localhost", path: "/console/api/server/list_domains", token: camille, want: 401},
		{name: "a domain of no Cahut", method: http.MethodPost, host: "nope.localhost", path: "/console/api/get_account", token: camille, want: 421},
	} {
		r := call(t, f, console, tt.method, tt.host, tt.path, tt.token, tt.body)
		if r.status != tt.want || !strings.Contains(r.raw, tt.says) {
			t.Errorf("%s: %d %s, want %d saying %q", tt.name, r.status, r.raw, tt.want, tt.says)
		}
		if r.header.Get("Access-Control-Allow-Origin") != "" {
			t.Errorf("%s: Access-Control-Allow-Origin = %q", tt.name, r.header.Get("Access-Control-Allow-Origin"))
		}
	}
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/get_account", "", ``); r.header.Get("WWW-Authenticate") != "Bearer" {
		t.Errorf("WWW-Authenticate = %q", r.header.Get("WWW-Authenticate"))
	}
	if r := call(t, f, console, http.MethodGet, "localhost", "/console/api/get_account", "", ``); r.header.Get("Allow") != "POST" {
		t.Errorf("Allow = %q", r.header.Get("Allow"))
	}

	// A person creates a Room, and the API returns what the operation does.
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/create_room", camille, `{"title": "Marketing"}`)
	if r.status != http.StatusOK || r.body["name"] != "marketing" || r.body["role"] != "owner" {
		t.Errorf("create_room: %d %s", r.status, r.raw)
	}
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/sign_out", camille, ``)
	if r.status != http.StatusOK {
		t.Errorf("sign_out: %d %s", r.status, r.raw)
	}
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/get_account", camille, ``); r.status != http.StatusUnauthorized {
		t.Errorf("after sign_out: %d %s", r.status, r.raw)
	}

	// The tests build no console: its address answers a notice.
	r = call(t, f, console, http.MethodGet, "localhost", "/console/", "", ``)
	if r.status != http.StatusNotFound || !strings.Contains(r.raw, "built without its console") {
		t.Errorf("GET /console/: %d %s", r.status, r.raw)
	}
}

func TestAPIForServerAdmins(t *testing.T) {
	f := setup(t)
	console := f.ts
	a, err := operation.Do(context.Background(), f.srv.Registry, server.CreateAdmin, admin, server.CreateAdminInput{Handle: "root", Name: "Root"})
	if err != nil {
		t.Fatal(err)
	}
	r := call(t, f, console, http.MethodPost, "localhost", "/console/api/server/activate_account", "",
		`{"code": "`+a.Activation.Code+`", "password": "`+password+`"}`)
	if r.status != http.StatusOK || r.body["domain"] != "localhost" {
		t.Fatalf("activate_account of the server: %d %s", r.status, r.raw)
	}
	root := r.body["token"].(string)

	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/server/list_domains", root, ``)
	if r.status != http.StatusOK || len(r.body["cahuts"].([]any)) != 2 {
		t.Errorf("list_domains: %d %s", r.status, r.raw)
	}
	// In the Cahut of the domain, the administrator manages people, but
	// sees no Window.
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/list_users", root, ``)
	if r.status != http.StatusOK || len(r.body["users"].([]any)) != 1 {
		t.Errorf("list_users as an administrator of the server: %d %s", r.status, r.raw)
	}
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/list_windows", root, ``)
	if r.status != http.StatusOK || len(r.body["windows"].([]any)) != 0 {
		t.Errorf("list_windows as an administrator of the server: %d %s", r.status, r.raw)
	}
	// The session holds at its domain only.
	if r := call(t, f, console, http.MethodPost, "beta.localhost", "/console/api/server/list_domains", root, ``); r.status != http.StatusUnauthorized {
		t.Errorf("list_domains at another domain: %d %s", r.status, r.raw)
	}
	if r := call(t, f, console, http.MethodPost, "beta.localhost", "/console/api/list_users", root, ``); r.status != http.StatusUnauthorized {
		t.Errorf("list_users at another domain: %d %s", r.status, r.raw)
	}
}

func TestAPISignInLimit(t *testing.T) {
	f := setup(t)
	console := f.ts
	body := `{"handle": "nobody", "password": "` + password + `"}`
	for range 10 {
		if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/sign_in", "", body); r.status != http.StatusForbidden {
			t.Fatalf("sign_in of nobody: %d %s", r.status, r.raw)
		}
	}
	r := call(t, f, console, http.MethodPost, "beta.localhost", "/console/api/sign_in", "", body)
	if r.status != http.StatusTooManyRequests || !strings.Contains(r.raw, "this address") {
		t.Errorf("the 11th sign-in from an address: %d %s", r.status, r.raw)
	}
}

// TestKnockAnsweredInTheConsole has an AI knock over MCP, and its person
// approve the Knock in the console. See docs/adr/0021-ais-knock-first.md.
func TestKnockAnsweredInTheConsole(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	console := f.ts
	camille := activate(t, f, console, "camille", true)
	c := f.open(t, "localhost")
	tok, err := operation.Do(ctx, c.Registry, account.CreateToken, admin, account.CreateTokenInput{User: "camille", Name: "Claude Code"})
	if err != nil {
		t.Fatal(err)
	}
	ai := f.connect(t, "localhost", tok.Token)

	res, err := ai.CallTool(ctx, &mcp.CallToolParams{Name: "disable_user", Arguments: map[string]any{"user": "olivier"}})
	if err != nil {
		t.Fatal(err)
	}
	if text := res.Content[0].(*mcp.TextContent).Text; !res.IsError || !strings.Contains(text, "/console/knocks/1.") ||
		res.StructuredContent != nil {
		t.Fatalf("disable_user by an AI: %v %q", res.IsError, text)
	}

	r := call(t, f, console, http.MethodPost, "localhost", "/console/api/list_knocks", camille, `{"waiting": true}`)
	if r.status != http.StatusOK || r.body["waiting"] != float64(1) {
		t.Fatalf("list_knocks: %d %s", r.status, r.raw)
	}
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/approve_knock", camille, `{"knock": 1}`)
	if r.status != http.StatusOK || r.body["knock"].(map[string]any)["state"] != "done" ||
		r.body["output"].(map[string]any)["status"] != "disabled" {
		t.Fatalf("approve_knock: %d %s", r.status, r.raw)
	}

	res, err = ai.CallTool(ctx, &mcp.CallToolParams{Name: "get_knock", Arguments: map[string]any{"knock": 1}})
	if err != nil || res.IsError || res.StructuredContent.(map[string]any)["state"] != "done" {
		t.Errorf("get_knock: %v %+v", err, res)
	}
}

// The first start goes through the API of the server, without a session,
// at the domain of its Cahut only. See
// docs/adr/0025-first-start-in-the-browser.md.
func TestFirstStartOverTheAPI(t *testing.T) {
	f := setup(t)
	console := f.ts
	link, err := operation.Do(context.Background(), f.srv.Registry, f.srv.CreateSetupLink, admin, server.CreateSetupLinkInput{Domain: "beta.localhost"})
	if err != nil {
		t.Fatal(err)
	}
	body := `{"code": "` + link.Code + `"}`
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/server/get_setup", "", body); r.status != http.StatusNotFound {
		t.Errorf("get_setup at another Cahut: %d %s", r.status, r.raw)
	}
	r := call(t, f, console, http.MethodPost, "beta.localhost", "/console/api/server/get_setup", "", body)
	if r.status != http.StatusOK || r.body["server_admin"] != true {
		t.Fatalf("get_setup: %d %s", r.status, r.raw)
	}
	r = call(t, f, console, http.MethodPost, "beta.localhost", "/console/api/server/set_up", "",
		`{"code": "`+link.Code+`", "handle": "ana", "name": "Ana", "password": "`+password+`"}`)
	if r.status != http.StatusOK {
		t.Fatalf("set_up: %d %s", r.status, r.raw)
	}
	token := r.body["session"].(map[string]any)["token"].(string)
	if r := call(t, f, console, http.MethodPost, "beta.localhost", "/console/api/get_account", token, ``); r.status != http.StatusOK || r.body["role"] != "admin" {
		t.Errorf("get_account with the session of the first start: %d %s", r.status, r.raw)
	}
	if r := call(t, f, console, http.MethodPost, "beta.localhost", "/console/api/server/set_up", "",
		`{"code": "`+link.Code+`", "handle": "eve", "name": "Eve", "password": "`+password+`"}`); r.status != http.StatusNotFound {
		t.Errorf("set_up again: %d %s", r.status, r.raw)
	}
}
