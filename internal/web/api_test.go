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
	"cahut.cloud/internal/cahut"
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

// serverAdminAt activates an administrator of the server at the console of
// the server of host, and returns their session.
func serverAdminAt(t *testing.T, f *fixture, host, handle string) string {
	t.Helper()
	a, err := operation.Do(context.Background(), f.srv.Registry, server.CreateAdmin, admin, server.CreateAdminInput{Handle: handle, Name: handle})
	if err != nil {
		t.Fatal(err)
	}
	r := call(t, f, f.ts, http.MethodPost, host, "/server/api/activate_account", "",
		`{"code": "`+a.Activation.Code+`", "password": "`+password+`"}`)
	if r.status != http.StatusOK {
		t.Fatalf("activate_account of the server: %d %s", r.status, r.raw)
	}
	return r.body["token"].(string)
}

// From the domain where they signed in, the administrators of the server
// call the operations of the server, and in every Cahut those that serve
// the server. See docs/adr/0058-a-console-for-the-server.md.
func TestServerAPI(t *testing.T) {
	f := setup(t)
	root := serverAdminAt(t, f, "localhost", "root")
	post := func(host, path, token, body string) apiResponse {
		t.Helper()
		return call(t, f, f.ts, http.MethodPost, host, path, token, body)
	}

	if r := post("localhost", "/server/api/list_domains", root, `{}`); r.status != http.StatusOK || len(r.body["cahuts"].([]any)) != 2 {
		t.Errorf("list_domains: %d %s", r.status, r.raw)
	}
	for _, domain := range []string{"beta.localhost", "BETA.localhost."} {
		if r := post("localhost", "/server/api/cahuts/"+domain+"/get_cahut", root, `{}`); r.status != http.StatusOK || r.body["domain"] != "beta.localhost" {
			t.Errorf("get_cahut of %s: %d %s", domain, r.status, r.raw)
		}
	}
	// A Cahut that the server does not hold, or an operation that does not
	// serve the server, is not found.
	for _, path := range []string{
		"/server/api/cahuts/nope.localhost/get_cahut",
		"/server/api/cahuts/..%2Fetc/get_cahut",
		"/server/api/cahuts/beta.localhost/list_users",
		"/server/api/cahuts/beta.localhost/nope",
	} {
		if r := post("localhost", path, root, `{}`); r.status != http.StatusNotFound || r.body["error"] == nil {
			t.Errorf("POST %s: %d %s", path, r.status, r.raw)
		}
	}

	// The console of the server says in its language that a Cahut is not
	// there, such as one that moved to another domain.
	if r := post("localhost", "/server/api/cahuts/nope.localhost/get_cahut", root, `{}`); r.body["code"] != "cahut_not_on_server" ||
		r.body["values"].(map[string]any)["domain"] != "nope.localhost" {
		t.Errorf("get_cahut of a Cahut that the server does not hold: %d %s", r.status, r.raw)
	}

	// A suspended Cahut still answers the administrators of the server.
	if r := post("localhost", "/server/api/cahuts/beta.localhost/suspend_cahut", root, `{"reason": "Unpaid."}`); r.status != http.StatusOK {
		t.Fatalf("suspend_cahut: %d %s", r.status, r.raw)
	}
	if r := post("localhost", "/server/api/cahuts/beta.localhost/get_cahut", root, `{}`); r.status != http.StatusOK || r.body["suspension"] == nil {
		t.Errorf("get_cahut of a suspended Cahut: %d %s", r.status, r.raw)
	}
	if r := post("localhost", "/server/api/cahuts/beta.localhost/resume_cahut", root, `{}`); r.status != http.StatusOK {
		t.Errorf("resume_cahut: %d %s", r.status, r.raw)
	}
	// What they do in a Cahut goes to its Logbook, under their handle.
	var logged bool
	for _, e := range f.call(t, "beta.localhost", admin, "list_logbook", `{}`).(cahut.LogbookPage).Entries {
		logged = logged || e.Operation == "suspend_cahut" && e.UserHandle == "root"
	}
	if !logged {
		t.Error("the Logbook of beta.localhost does not say that root suspended it")
	}

	// Another kind of session, or a session of another domain, is refused.
	camille := activate(t, f, f.ts, "camille", false)
	for who, tc := range map[string]struct{ host, token string }{
		"a person":                    {"localhost", camille},
		"an AI":                       {"localhost", f.token},
		"nobody":                      {"localhost", ""},
		"a session of beta.localhost": {"beta.localhost", root},
	} {
		for _, path := range []string{"/server/api/list_domains", "/server/api/cahuts/localhost/get_cahut"} {
			if r := post(tc.host, path, tc.token, `{}`); r.status != http.StatusUnauthorized {
				t.Errorf("POST %s as %s: %d %s", path, who, r.status, r.raw)
			}
		}
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
	if r.status != http.StatusTooManyRequests || !strings.Contains(r.raw, "this address") || r.body["code"] != "too_many_attempts_from_address" {
		t.Errorf("the 11th sign-in from an address: %d %s", r.status, r.raw)
	}
}

// TestAPIRefusalCodes checks that the refusals that people meet give their
// code and their values next to their message, which the console says in
// the language of its person. See
// docs/adr/0037-refusals-with-a-code-for-the-console.md.
func TestAPIRefusalCodes(t *testing.T) {
	f := setup(t)
	console := f.ts
	activate(t, f, console, "camille", false)
	c := f.open(t, "localhost")
	u, err := operation.Do(context.Background(), c.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: "bruno", Name: "Bruno"})
	if err != nil {
		t.Fatal(err)
	}

	checkRefusals(t, f, console, "", []refusalCase{
		{
			name: "a wrong password", path: "/console/api/sign_in", body: `{"handle": "camille", "password": "not the password at all"}`,
			status: 403, error: "not allowed: wrong handle or password", code: "wrong_handle_or_password",
		},
		{
			name: "a short password", path: "/console/api/activate_account", body: `{"code": "` + u.Activation.Code + `", "password": "short"}`,
			status: 400, error: "invalid input: the password has 5 characters: use at least 15", code: "password_too_short",
			values: map[string]any{"length": 5.0, "limit": 15.0},
		},
		{
			name: "a used link", path: "/console/api/activate_account", body: `{"code": "used", "password": "` + password + `"}`,
			status: 403, code: "activation_link_gone",
			error: "not allowed: this activation link is unknown, used or expired: ask an administrator of the Cahut for a new one",
		},
		{
			name: "a link of the server", path: "/console/api/server/activate_account", body: `{"code": "used", "password": "` + password + `"}`,
			status: 403, code: "server_activation_link_gone",
		},
		{
			name: "no session", path: "/console/api/get_account",
			status: 401, error: "invalid or ended session: sign in first, then send the session token in the Authorization header", code: "session_ended",
		},
		{
			name: "a refusal without a code", path: "/console/api/get_invitation", body: `{"x": 1}`,
			status: 400, error: `invalid input: arguments: json: unknown field "x"`,
		},
	})
}

// A refusalCase is a call of the console to the Cahut at localhost, and
// the refusal it gets: its status, its message, its code and its values.
type refusalCase struct {
	name, path, body string
	status           int
	error, code      string
	values           map[string]any
}

// checkRefusals makes the calls of cases with the session token, and checks
// the refusal of each.
func checkRefusals(t *testing.T, f *fixture, console *httptest.Server, token string, cases []refusalCase) {
	t.Helper()
	for _, tt := range cases {
		r := call(t, f, console, http.MethodPost, "localhost", tt.path, token, tt.body)
		code, _ := r.body["code"].(string)
		values, _ := r.body["values"].(map[string]any)
		if r.status != tt.status || tt.error != "" && r.body["error"] != tt.error || code != tt.code || len(values) != len(tt.values) {
			t.Errorf("%s: %d %s", tt.name, r.status, r.raw)
		}
		for k, v := range tt.values {
			if values[k] != v {
				t.Errorf("%s: values[%s] = %v, want %v", tt.name, k, values[k], v)
			}
		}
	}
}

// TestAPIRefusalCodesOfScreens checks the codes of the refusals that people
// meet on the screens of the console: People, Rooms, Account, the page of a
// Window and the first start.
func TestAPIRefusalCodesOfScreens(t *testing.T) {
	f := setup(t)
	console := f.ts
	camille := activate(t, f, console, "camille", true)
	publishDeck(t, f, f.person(t, "localhost", "camille"))
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/create_room", camille, `{"title": "Marketing"}`); r.status != http.StatusOK {
		t.Fatalf("create_room: %d %s", r.status, r.raw)
	}

	checkRefusals(t, f, console, camille, []refusalCase{
		{
			name: "a handle taken", path: "/console/api/create_user", body: `{"handle": "olivier", "name": "Olivier"}`,
			status: 400, error: `invalid input: handle "olivier" is taken, by a person or a Room`, code: "handle_taken",
			values: map[string]any{"handle": "olivier"},
		},
		{
			name: "a handle with capitals", path: "/console/api/create_user", body: `{"handle": "Ana", "name": "Ana"}`,
			status: 400, code: "handle_invalid", values: map[string]any{"handle": "Ana"},
			error: `invalid input: handle "Ana": use 2 to 32 lowercase letters, digits and hyphens, starting with a letter and ending with a letter or a digit`,
		},
		{
			name: "the last administrator", path: "/console/api/set_user_role", body: `{"user": "camille", "role": "member"}`,
			status: 400, error: "invalid input: the Cahut needs an administrator: make another person one first", code: "cahut_needs_an_admin",
		},
		{
			name: "the name of a Room taken", path: "/console/api/create_room", body: `{"title": "Olivier", "name": "olivier"}`,
			status: 400, error: `invalid input: name "olivier" is taken, by a Room or a person: choose another`, code: "room_name_taken",
			values: map[string]any{"name": "olivier"},
		},
		{
			name: "a reserved name", path: "/console/api/create_room", body: `{"title": "Console"}`,
			status: 400, code: "room_name_reserved", values: map[string]any{"name": "console"},
		},
		{
			name: "the last owner", path: "/console/api/remove_room_member", body: `{"room": "marketing", "user": "camille"}`,
			status: 400, error: "invalid input: the Room marketing needs an owner: make another person one first", code: "room_needs_an_owner",
			values: map[string]any{"room": "marketing"},
		},
		{
			name: "a short Key", path: "/console/api/set_window_access", body: `{"window": "camille/q3-review", "access": "key", "key": "short"}`,
			status: 400, error: "invalid input: the Key has 5 characters: give it 10 at least, or let Cahut draw one", code: "key_too_short",
			values: map[string]any{"length": 5.0, "limit": 10.0},
		},
		{
			name: "a Key with spaces", path: "/console/api/set_window_access", body: `{"window": "camille/q3-review", "access": "key", "key": " a long enough Key"}`,
			status: 400, error: "invalid input: the Key starts or ends with a space", code: "key_has_spaces",
		},
		{
			name: "an invitation without a relay", path: "/console/api/invite_to_window", body: `{"window": "camille/q3-review", "to": ["olivier"]}`,
			status: 400, code: "no_mail_to_invite",
			error: "invalid input: this Cahut sends no e-mail: give them the address of the Window yourself, with its Key if it has one",
		},
		{
			name: "a new address without a relay", path: "/console/api/change_email", body: `{"email": "camille@example.org"}`,
			status: 400, code: "no_mail",
			error: "invalid input: this Cahut sends no e-mail: an administrator of the Cahut sets its relay first, on the page E-mail of the console",
		},
		{
			name: "a used link to the first start", path: "/console/api/server/get_setup", body: `{"code": "used"}`,
			status: 404, code: "setup_link_gone",
		},
	})

	f.relay(t)
	if _, err := f.open(t, "localhost").DB.Write.Exec("UPDATE users SET email = 'camille@example.org' WHERE handle = 'camille'"); err != nil {
		t.Fatal(err)
	}
	checkRefusals(t, f, console, camille, []refusalCase{
		{
			name: "an address to a Window of its Room", path: "/console/api/invite_to_window", body: `{"window": "camille/q3-review", "to": ["ana@example.org"]}`,
			status: 400, code: "invite_address_room_only", values: map[string]any{"room": "camille", "window": "q3-review"},
			error: "invalid input: camille/q3-review opens to its Room only: invite people of the Cahut by handle, " +
				"or set its access to key first, so that each address gets a personal link",
		},
		{
			name: "someone outside the Room", path: "/console/api/invite_to_window", body: `{"window": "camille/q3-review", "to": ["olivier"]}`,
			status: 400, code: "invite_not_in_room", values: map[string]any{"handle": "olivier", "room": "camille", "window": "q3-review"},
			error: "invalid input: olivier is not in the Room camille, and camille/q3-review opens to its Room only: " +
				"add them to the Room, or set its access to internal or key first",
		},
		{
			name: "nobody of this handle", path: "/console/api/invite_to_window", body: `{"window": "camille/q3-review", "to": ["nobody"]}`,
			status: 404, error: `not found: no person "nobody" in this Cahut: invite someone else by e-mail address`, code: "invite_no_person",
			values: map[string]any{"to": "nobody"},
		},
		{
			name: "half an address", path: "/console/api/invite_to_window", body: `{"window": "camille/q3-review", "to": ["joe@"]}`,
			status: 400, error: `invalid input: to: "joe@" is not an e-mail address: give one address alone, such as jane@example.com`,
			code: "invite_bad_address", values: map[string]any{"to": "joe@"},
		},
		{
			name: "the same address", path: "/console/api/change_email", body: `{"email": "camille@example.org"}`,
			status: 400, error: "invalid input: camille@example.org is already your e-mail address", code: "email_already_yours",
			values: map[string]any{"email": "camille@example.org"},
		},
	})
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
