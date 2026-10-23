package oauth_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/oauth"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
)

const (
	issuer   = "https://cahut.test"
	console  = "https://cahut.test/console"
	resource = "https://cahut.test/mcp"
	redirect = "https://claude.ai/api/mcp/auth_callback"
	verifier = "a-code-verifier-of-at-least-43-characters-made-of-letters"
)

var admin = operation.Actor{Kind: operation.Admin, Via: "cli"}

type fixture struct {
	t       *testing.T
	c       *cahut.Cahut
	now     time.Time
	olivier operation.Actor // signed in to the console; administers the Cahut
	camille operation.Actor
}

// setup opens a Cahut where olivier, who administers it, and camille share
// the team Room marketing.
func setup(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	c, err := cahut.Open(ctx, cahut.Config{Dir: t.TempDir(), Domain: "cahut.test",
		BaseURL: "https://cahut.test/", ConsoleURL: console + "/"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	f := &fixture{t: t, c: c, now: time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)}
	c.OAuth.Now = func() time.Time { return f.now }
	c.Registry.Now = func() time.Time { return f.now }
	people := map[string]*operation.Actor{"olivier": &f.olivier, "camille": &f.camille}
	for _, handle := range []string{"olivier", "camille"} {
		u, err := operation.Do(ctx, c.Registry, account.CreateUser, admin,
			account.CreateUserInput{Handle: handle, Name: handle, Admin: handle == "olivier"})
		if err != nil {
			t.Fatal(err)
		}
		*people[handle] = operation.Actor{Kind: operation.Person, Via: "api", Domain: "cahut.test", UserID: u.ID, Handle: handle, Session: 1}
	}
	if _, err := operation.Do(ctx, c.Registry, room.CreateRoom, f.olivier, room.CreateRoomInput{Title: "Marketing"}); err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Do(ctx, c.Registry, room.AddMember, f.olivier, room.MemberInput{Room: "marketing", User: "camille"}); err != nil {
		t.Fatal(err)
	}
	return f
}

// register registers a client with a redirect URI, and returns its ID.
func (f *fixture) register(redirects ...string) string {
	f.t.Helper()
	body, _ := json.Marshal(map[string]any{"client_name": "Claude", "redirect_uris": redirects, "unknown_field": true})
	rec := httptest.NewRecorder()
	f.c.OAuth.ServeRegister(rec, httptest.NewRequestWithContext(context.Background(), http.MethodPost, issuer+"/oauth/register", strings.NewReader(string(body))))
	if rec.Code != http.StatusCreated {
		f.t.Fatalf("register: %d %s", rec.Code, rec.Body)
	}
	var reg oauth.Registration
	if err := json.Unmarshal(rec.Body.Bytes(), &reg); err != nil {
		f.t.Fatal(err)
	}
	if reg.TokenEndpointAuthMethod != "none" || reg.ClientName != "Claude" {
		f.t.Errorf("registration = %+v", reg)
	}
	return reg.ClientID
}

func challenge(v string) string {
	sum := sha256.Sum256([]byte(v))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (f *fixture) request(client string) oauth.Request {
	return oauth.Request{ResponseType: "code", ClientID: client, RedirectURI: redirect, State: "xyz",
		CodeChallenge: challenge(verifier), CodeChallengeMethod: "S256", Resource: resource}
}

func (f *fixture) invoke(who operation.Actor, name string, in any) (any, error) {
	f.t.Helper()
	b, err := json.Marshal(in)
	if err != nil {
		f.t.Fatal(err)
	}
	return f.c.Registry.Invoke(context.Background(), name, who, b)
}

// approve connects the client for who, and returns the code.
func (f *fixture) approve(who operation.Actor, in oauth.ApproveInput) string {
	f.t.Helper()
	out, err := f.invoke(who, "approve_authorization", in)
	if err != nil {
		f.t.Fatalf("approve_authorization: %v", err)
	}
	u, err := url.Parse(out.(oauth.Redirect).Redirect)
	if err != nil {
		f.t.Fatal(err)
	}
	q := u.Query()
	if q.Get("state") != "xyz" || q.Get("iss") != issuer || q.Get("code") == "" || u.Host != "claude.ai" {
		f.t.Errorf("redirect = %s", u)
	}
	return q.Get("code")
}

// token calls the token endpoint with form, and returns the status and the
// decoded answer.
func (f *fixture) token(form url.Values) (int, map[string]any) {
	f.t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, issuer+"/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	f.c.OAuth.ServeToken(rec, req)
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		f.t.Fatalf("token: %d %s", rec.Code, rec.Body)
	}
	return rec.Code, out
}

func (f *fixture) exchange(client, code string) (int, map[string]any) {
	return f.token(url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {client},
		"redirect_uri": {redirect}, "code_verifier": {verifier}, "resource": {resource}})
}

func (f *fixture) refresh(client, token string) (int, map[string]any) {
	return f.token(url.Values{"grant_type": {"refresh_token"}, "refresh_token": {token}, "client_id": {client}})
}

func (f *fixture) roomID(name string) int64 {
	var id int64
	if err := f.c.DB.Read.QueryRow("SELECT id FROM rooms WHERE name = ?", name).Scan(&id); err != nil {
		f.t.Fatal(err)
	}
	return id
}

func (f *fixture) logbook() []logbook.Entry {
	f.t.Helper()
	page, err := operation.Do(context.Background(), f.c.Registry, cahut.ListLogbook, admin, cahut.ListLogbookInput{Limit: 100})
	if err != nil {
		f.t.Fatal(err)
	}
	return page.Entries
}

func TestMetadata(t *testing.T) {
	f := setup(t)
	rec := httptest.NewRecorder()
	f.c.OAuth.ServeMetadata(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, issuer+"/.well-known/oauth-authorization-server", nil))
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]any{
		"issuer":                                         issuer,
		"authorization_endpoint":                         console + "/oauth/authorize",
		"token_endpoint":                                 issuer + "/oauth/token",
		"registration_endpoint":                          issuer + "/oauth/register",
		"revocation_endpoint":                            issuer + "/oauth/revoke",
		"client_id_metadata_document_supported":          true,
		"authorization_response_iss_parameter_supported": true,
	} {
		if m[key] != want {
			t.Errorf("%s = %v, want %v", key, m[key], want)
		}
	}
	if fmt.Sprint(m["code_challenge_methods_supported"]) != "[S256]" || fmt.Sprint(m["token_endpoint_auth_methods_supported"]) != "[none]" ||
		fmt.Sprint(m["scopes_supported"]) != "[read write offline_access]" {
		t.Errorf("metadata = %v", m)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("the metadata answer no CORS request")
	}
}

// TestConnect follows an AI from its registration to its tokens, limited to
// the Room marketing and to reading.
func TestConnect(t *testing.T) {
	f := setup(t)
	client := f.register(redirect)

	out, err := f.invoke(f.camille, "describe_authorization", oauth.DescribeInput{Request: f.request(client)})
	if err != nil {
		t.Fatal(err)
	}
	a := out.(oauth.Authorization)
	if a.Client != "Claude" || a.RedirectHost != "claude.ai" || a.Scope != oauth.Write || a.Redirect != "" || len(a.Rooms) != 2 || a.Rooms[0].Name != "camille" {
		t.Errorf("describe_authorization = %+v", a)
	}

	code := f.approve(f.camille, oauth.ApproveInput{Request: f.request(client), Scope: oauth.Read, Rooms: []string{"marketing"}})
	status, tokens := f.exchange(client, code)
	if status != http.StatusOK || tokens["token_type"] != "Bearer" || tokens["scope"] != "read offline_access" || tokens["expires_in"] != 3600.0 {
		t.Fatalf("exchange = %d %v", status, tokens)
	}
	access, refresh := tokens["access_token"].(string), tokens["refresh_token"].(string)
	if !strings.HasPrefix(access, oauth.AccessPrefix) || !strings.HasPrefix(refresh, oauth.RefreshPrefix) {
		t.Errorf("tokens = %v", tokens)
	}
	owner, expires, err := f.c.OAuth.AccessOwner(context.Background(), access)
	if err != nil {
		t.Fatal(err)
	}
	if owner.Handle != "camille" || !owner.ReadOnly || len(owner.Rooms) != 1 || owner.Rooms[0] != f.roomID("marketing") ||
		owner.TokenName != "Claude" || owner.Connection == 0 || !expires.Equal(f.now.Add(time.Hour)) {
		t.Errorf("owner = %+v, expires %v", owner, expires)
	}

	// A code works once; the second time ends the connection.
	if status, out := f.exchange(client, code); status != http.StatusBadRequest || out["error"] != "invalid_grant" {
		t.Errorf("a code used twice = %d %v", status, out)
	}
	if _, _, err := f.c.OAuth.AccessOwner(context.Background(), access); !errors.Is(err, account.ErrInvalidToken) {
		t.Errorf("the access token of a connection whose code came back: %v", err)
	}
	if e := f.logbook()[0]; e.Operation != "end_connection" || e.Actor != "ai" || e.Connection != owner.Connection ||
		e.Detail != "an authorization code was used twice" {
		t.Errorf("last entry = %+v", e)
	}
	if status, _ := f.refresh(client, refresh); status != http.StatusBadRequest {
		t.Errorf("refresh on an ended connection = %d", status)
	}
}

func TestRefresh(t *testing.T) {
	f := setup(t)
	client := f.register(redirect)
	code := f.approve(f.olivier, oauth.ApproveInput{Request: f.request(client), Scope: oauth.Write, AllRooms: true})
	_, tokens := f.exchange(client, code)
	first := tokens["refresh_token"].(string)
	// Write includes read, and every connection refreshes: clients compare
	// these scopes with those they asked for.
	if tokens["scope"] != "read write offline_access" {
		t.Errorf("scope of a connection that may write = %v", tokens["scope"])
	}

	f.now = f.now.Add(2 * time.Hour)
	if _, _, err := f.c.OAuth.AccessOwner(context.Background(), tokens["access_token"].(string)); !errors.Is(err, account.ErrInvalidToken) {
		t.Errorf("an access token after an hour: %v", err)
	}
	status, tokens := f.refresh(client, first)
	if status != http.StatusOK {
		t.Fatalf("refresh = %d %v", status, tokens)
	}
	owner, _, err := f.c.OAuth.AccessOwner(context.Background(), tokens["access_token"].(string))
	if err != nil || owner.Rooms != nil || owner.ReadOnly {
		t.Errorf("owner of a connection to every Room = %+v, %v", owner, err)
	}
	if status, _ := f.refresh("another-client", tokens["refresh_token"].(string)); status != http.StatusBadRequest {
		t.Errorf("refresh by another client = %d", status)
	}

	// Two refreshes sent at once: the late one is refused, and the
	// connection stays.
	f.now = f.now.Add(10 * time.Second)
	if status, out := f.refresh(client, first); status != http.StatusBadRequest || !strings.Contains(out["error_description"].(string), "just replaced") {
		t.Errorf("a refresh token replaced 10 s ago = %d %v", status, out)
	}
	if _, _, err := f.c.OAuth.AccessOwner(context.Background(), tokens["access_token"].(string)); err != nil {
		t.Errorf("the connection ended within the grace period: %v", err)
	}
	// Later, a replaced refresh token is a replay: the connection ends.
	f.now = f.now.Add(time.Minute)
	if status, _ := f.refresh(client, first); status != http.StatusBadRequest {
		t.Errorf("a replaced refresh token = %d", status)
	}
	if _, _, err := f.c.OAuth.AccessOwner(context.Background(), tokens["access_token"].(string)); !errors.Is(err, account.ErrInvalidToken) {
		t.Errorf("the connection outlived a replayed refresh token: %v", err)
	}

	// A refresh token lives 30 days.
	f.now = time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)
	code = f.approve(f.olivier, oauth.ApproveInput{Request: f.request(client), Scope: oauth.Write, AllRooms: true})
	_, tokens = f.exchange(client, code)
	f.now = f.now.Add(31 * 24 * time.Hour)
	if status, _ := f.refresh(client, tokens["refresh_token"].(string)); status != http.StatusBadRequest {
		t.Errorf("a refresh token after 31 days = %d", status)
	}
}

func TestExchangeChecks(t *testing.T) {
	f := setup(t)
	client := f.register(redirect, "http://127.0.0.1/callback")
	tests := map[string]struct {
		form url.Values
		want string
	}{
		"wrong verifier":   {url.Values{"code_verifier": {strings.Repeat("x", 43)}}, "invalid_grant"},
		"no verifier":      {url.Values{"code_verifier": {""}}, "invalid_grant"},
		"another client":   {url.Values{"client_id": {"someone-else"}}, "invalid_grant"},
		"another redirect": {url.Values{"redirect_uri": {"https://claude.ai/other"}}, "invalid_grant"},
		"another resource": {url.Values{"resource": {"https://other.test/mcp"}}, "invalid_target"},
		"unknown code":     {url.Values{"code": {"nope"}}, "invalid_grant"},
		"no grant type":    {url.Values{"grant_type": {""}}, "invalid_request"},
		"password grant":   {url.Values{"grant_type": {"password"}}, "unsupported_grant_type"},
	}
	for name, tt := range tests {
		code := f.approve(f.olivier, oauth.ApproveInput{Request: f.request(client), Scope: oauth.Write, AllRooms: true})
		form := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {client},
			"redirect_uri": {redirect}, "code_verifier": {verifier}}
		for k, v := range tt.form {
			form[k] = v
		}
		if status, out := f.token(form); status != http.StatusBadRequest || out["error"] != tt.want {
			t.Errorf("%s: %d %v, want %s", name, status, out, tt.want)
		}
	}

	code := f.approve(f.olivier, oauth.ApproveInput{Request: f.request(client), Scope: oauth.Write, AllRooms: true})
	f.now = f.now.Add(6 * time.Minute)
	if status, out := f.exchange(client, code); status != http.StatusBadRequest || !strings.Contains(out["error_description"].(string), "expired") {
		t.Errorf("an expired code = %d %v", status, out)
	}

	// A loopback redirect URI matches whatever its port.
	req := f.request(client)
	req.RedirectURI = "http://127.0.0.1:53682/callback"
	if _, err := f.invoke(f.olivier, "describe_authorization", oauth.DescribeInput{Request: req}); err != nil {
		t.Errorf("a loopback redirect URI on another port: %v", err)
	}
	req.RedirectURI = "http://127.0.0.1:53682/other"
	if _, err := f.invoke(f.olivier, "describe_authorization", oauth.DescribeInput{Request: req}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("a loopback redirect URI with another path: %v", err)
	}
}

func TestDescribeChecks(t *testing.T) {
	f := setup(t)
	client := f.register(redirect)
	// Without a valid client and redirect URI, the request does not go
	// back to the client.
	for name, change := range map[string]func(*oauth.Request){
		"no client":         func(r *oauth.Request) { r.ClientID = "" },
		"unknown client":    func(r *oauth.Request) { r.ClientID = "nope" },
		"another redirect":  func(r *oauth.Request) { r.RedirectURI = "https://evil.test/callback" },
		"redirect by query": func(r *oauth.Request) { r.RedirectURI = redirect + "?x=1" },
	} {
		req := f.request(client)
		change(&req)
		if _, err := f.invoke(f.olivier, "describe_authorization", oauth.DescribeInput{Request: req}); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("%s: %v, want ErrInvalid", name, err)
		}
	}
	// Otherwise, it goes back with an error.
	for name, tt := range map[string]struct {
		change func(*oauth.Request)
		want   string
	}{
		"implicit grant":   {func(r *oauth.Request) { r.ResponseType = "token" }, "unsupported_response_type"},
		"no PKCE":          {func(r *oauth.Request) { r.CodeChallenge, r.CodeChallengeMethod = "", "" }, "invalid_request"},
		"plain PKCE":       {func(r *oauth.Request) { r.CodeChallengeMethod = "plain" }, "invalid_request"},
		"another resource": {func(r *oauth.Request) { r.Resource = "https://other.test/mcp" }, "invalid_target"},
	} {
		req := f.request(client)
		tt.change(&req)
		out, err := f.invoke(f.olivier, "describe_authorization", oauth.DescribeInput{Request: req})
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		u, _ := url.Parse(out.(oauth.Authorization).Redirect)
		if u == nil || u.Query().Get("error") != tt.want || u.Query().Get("state") != "xyz" || u.Query().Get("iss") != issuer {
			t.Errorf("%s: redirect %v, want error %s", name, u, tt.want)
		}
		if _, err := f.invoke(f.olivier, "approve_authorization", oauth.ApproveInput{Request: req, Scope: oauth.Read, AllRooms: true}); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("approve %s: %v, want ErrInvalid", name, err)
		}
	}
	// The resource may differ by a final slash, or be absent.
	for _, resource := range []string{"https://CAHUT.test/mcp/", ""} {
		req := f.request(client)
		req.Resource = resource
		out, err := f.invoke(f.olivier, "describe_authorization", oauth.DescribeInput{Request: req})
		if err != nil || out.(oauth.Authorization).Redirect != "" {
			t.Errorf("resource %q: %+v, %v", resource, out, err)
		}
	}
}

func TestApproveChecks(t *testing.T) {
	f := setup(t)
	client := f.register(redirect)
	read := f.request(client)
	read.Scope = "read offline_access"
	for name, tt := range map[string]struct {
		who  operation.Actor
		in   oauth.ApproveInput
		want error
	}{
		"more than asked":     {f.olivier, oauth.ApproveInput{Request: read, Scope: oauth.Write, AllRooms: true}, operation.ErrInvalid},
		"unknown scope":       {f.olivier, oauth.ApproveInput{Request: read, Scope: "admin", AllRooms: true}, operation.ErrInvalid},
		"no Room":             {f.olivier, oauth.ApproveInput{Request: read, Scope: oauth.Read}, operation.ErrInvalid},
		"a Room of another":   {f.olivier, oauth.ApproveInput{Request: read, Scope: oauth.Read, Rooms: []string{"camille"}}, operation.ErrNotFound},
		"by the command line": {admin, oauth.ApproveInput{Request: read, Scope: oauth.Read, AllRooms: true}, operation.ErrForbidden},
		"by an AI": {operation.Actor{Kind: operation.AI, Via: "mcp", UserID: f.olivier.UserID, Handle: "olivier", TokenID: 1},
			oauth.ApproveInput{Request: read, Scope: oauth.Read, AllRooms: true}, operation.ErrForbidden},
	} {
		if _, err := f.invoke(tt.who, "approve_authorization", tt.in); !errors.Is(err, tt.want) {
			t.Errorf("%s: %v, want %v", name, err, tt.want)
		}
	}

	out, err := f.invoke(f.olivier, "deny_authorization", oauth.DenyInput{Request: f.request(client)})
	if err != nil {
		t.Fatal(err)
	}
	if u, _ := url.Parse(out.(oauth.Redirect).Redirect); u.Query().Get("error") != "access_denied" {
		t.Errorf("deny_authorization = %v", u)
	}
	if e := f.logbook()[0]; e.Operation != "deny_authorization" || e.UserHandle != "olivier" || e.Target != "Claude" {
		t.Errorf("last entry = %+v", e)
	}
}

func TestRegister(t *testing.T) {
	f := setup(t)
	for _, uris := range [][]string{
		{},
		{"http://claude.ai/callback"},
		{"javascript:alert(1)"},
		{"https://claude.ai/callback#fragment"},
		{"/relative"},
	} {
		body, _ := json.Marshal(map[string]any{"redirect_uris": uris})
		rec := httptest.NewRecorder()
		f.c.OAuth.ServeRegister(rec, httptest.NewRequestWithContext(context.Background(), http.MethodPost, issuer+"/oauth/register", strings.NewReader(string(body))))
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid_redirect_uri") {
			t.Errorf("register %q = %d %s", uris, rec.Code, rec.Body)
		}
	}
	// Applications register their own schemes, and terminals a loopback
	// address.
	f.register("cursor://anysphere.cursor-retrieval/oauth/callback", "http://localhost:33418/callback")

	// A client may list grants that Cahut does not give, as claude.ai does,
	// but not do without the authorization code.
	register := func(metadata string) (int, string) {
		rec := httptest.NewRecorder()
		f.c.OAuth.ServeRegister(rec, httptest.NewRequestWithContext(context.Background(), http.MethodPost, issuer+"/oauth/register", strings.NewReader(metadata)))
		return rec.Code, rec.Body.String()
	}
	status, body := register(`{"client_name": "Claude", "redirect_uris": ["https://claude.ai/api/mcp/auth_callback"],
		"grant_types": ["authorization_code", "refresh_token", "urn:ietf:params:oauth:grant-type:jwt-bearer"],
		"response_types": ["code"], "token_endpoint_auth_method": "none", "scope": "read write"}`)
	if status != http.StatusCreated || !strings.Contains(body, `"grant_types":["authorization_code","refresh_token"]`) {
		t.Errorf("register as claude.ai = %d %s", status, body)
	}
	for _, metadata := range []string{
		`{"redirect_uris": ["https://claude.ai/api/mcp/auth_callback"], "grant_types": ["client_credentials"]}`,
		`{"redirect_uris": ["https://claude.ai/api/mcp/auth_callback"], "response_types": ["token"]}`,
	} {
		if status, body := register(metadata); status != http.StatusBadRequest || !strings.Contains(body, "invalid_client_metadata") {
			t.Errorf("register %s = %d %s", metadata, status, body)
		}
	}

	// A client that receives no connection within 7 days is deleted at the
	// next registration; a connected one stays.
	unused, used := f.register(redirect), f.register(redirect)
	f.approve(f.olivier, oauth.ApproveInput{Request: f.request(used), Scope: oauth.Read, AllRooms: true})
	f.now = f.now.Add(8 * 24 * time.Hour)
	f.register(redirect)
	for id, want := range map[string]bool{unused: false, used: true} {
		var n int
		if err := f.c.DB.Read.QueryRow("SELECT count(*) FROM oauth_clients WHERE id = ?", id).Scan(&n); err != nil || (n == 1) != want {
			t.Errorf("client kept = %v, want %v (%v)", n == 1, want, err)
		}
	}
}

func TestConnectedAIs(t *testing.T) {
	f := setup(t)
	client := f.register(redirect)
	code := f.approve(f.camille, oauth.ApproveInput{Request: f.request(client), Scope: oauth.Write, Rooms: []string{"marketing", "camille"}})
	_, tokens := f.exchange(client, code)

	out, err := f.invoke(f.camille, "list_connected_ais", oauth.ListConnectedAIsInput{})
	if err != nil {
		t.Fatal(err)
	}
	list := out.(oauth.ConnectedAIs).Connections
	if len(list) != 1 || list[0].User != "camille" || list[0].Client != "Claude" || list[0].Status != oauth.Active ||
		fmt.Sprint(list[0].Rooms) != "[camille marketing]" || list[0].LastUsedAt == nil {
		t.Errorf("list_connected_ais = %+v", list)
	}
	id := list[0].ID
	if _, err := f.invoke(f.olivier, "list_connected_ais", oauth.ListConnectedAIsInput{User: "camille"}); err != nil {
		t.Errorf("an administrator lists the AIs of others: %v", err)
	}
	if _, err := f.invoke(f.camille, "list_connected_ais", oauth.ListConnectedAIsInput{User: "olivier"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("a member lists the AIs of others: %v", err)
	}

	// A disabled person's AIs are suspended.
	if _, err := operation.Do(context.Background(), f.c.Registry, account.DisableUser, admin, account.UserInput{User: "camille"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := f.c.OAuth.AccessOwner(context.Background(), tokens["access_token"].(string)); !errors.Is(err, account.ErrInvalidToken) {
		t.Errorf("the AI of a disabled person: %v", err)
	}
	if status, _ := f.refresh(client, tokens["refresh_token"].(string)); status != http.StatusBadRequest {
		t.Errorf("refresh for a disabled person = %d", status)
	}
	if _, err := operation.Do(context.Background(), f.c.Registry, account.EnableUser, admin, account.UserInput{User: "camille"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := f.c.OAuth.AccessOwner(context.Background(), tokens["access_token"].(string)); err != nil {
		t.Errorf("the AI of a person enabled again: %v", err)
	}

	// An administrator disconnects it; its tokens stop at once.
	out, err = f.invoke(f.olivier, "disconnect_ai", oauth.DisconnectAIInput{ID: id})
	if err != nil || out.(oauth.Connection).Status != oauth.Ended || out.(oauth.Connection).EndReason != "disconnected by olivier" {
		t.Errorf("disconnect_ai = %+v, %v", out, err)
	}
	if _, _, err := f.c.OAuth.AccessOwner(context.Background(), tokens["access_token"].(string)); !errors.Is(err, account.ErrInvalidToken) {
		t.Errorf("the access token of a disconnected AI: %v", err)
	}
	if _, err := f.invoke(f.olivier, "disconnect_ai", oauth.DisconnectAIInput{ID: id}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("disconnect_ai twice: %v", err)
	}
	// The AIs of others do not exist for a member.
	code = f.approve(f.olivier, oauth.ApproveInput{Request: f.request(client), Scope: oauth.Write, AllRooms: true})
	f.exchange(client, code)
	out, _ = f.invoke(f.olivier, "list_connected_ais", oauth.ListConnectedAIsInput{User: "olivier"})
	theirs := out.(oauth.ConnectedAIs).Connections[0].ID
	if _, err := f.invoke(f.camille, "disconnect_ai", oauth.DisconnectAIInput{ID: theirs}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("a member disconnects the AI of another: %v", err)
	}
}

func TestRevoke(t *testing.T) {
	f := setup(t)
	client := f.register(redirect)
	connect := func() map[string]any {
		code := f.approve(f.olivier, oauth.ApproveInput{Request: f.request(client), Scope: oauth.Write, AllRooms: true})
		_, tokens := f.exchange(client, code)
		return tokens
	}
	revoke := func(token string) {
		t.Helper()
		form := url.Values{"token": {token}, "client_id": {client}}
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, issuer+"/oauth/revoke", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		f.c.OAuth.ServeRevoke(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("revoke = %d %s", rec.Code, rec.Body)
		}
	}
	valid := func(token string) bool {
		_, _, err := f.c.OAuth.AccessOwner(context.Background(), token)
		return err == nil
	}

	tokens := connect()
	revoke(tokens["access_token"].(string))
	if valid(tokens["access_token"].(string)) {
		t.Error("a revoked access token still works")
	}
	if status, _ := f.refresh(client, tokens["refresh_token"].(string)); status != http.StatusOK {
		t.Errorf("revoking an access token ended its connection: %d", status)
	}

	tokens = connect()
	revoke(tokens["refresh_token"].(string))
	if valid(tokens["access_token"].(string)) {
		t.Error("revoking a refresh token left its connection")
	}
	if e := f.logbook()[0]; e.Operation != "end_connection" || e.Detail != "revoked by its client" {
		t.Errorf("last entry = %+v", e)
	}
	revoke("unknown") // answers 200 all the same
}

func TestMetadataDocuments(t *testing.T) {
	f := setup(t)
	var doc map[string]any
	var redirectTo string
	var gone atomic.Bool
	var agent atomic.Value
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agent.Store(r.UserAgent())
		if gone.Load() {
			http.NotFound(w, r)
			return
		}
		if redirectTo != "" {
			http.Redirect(w, r, redirectTo, http.StatusFound)
			return
		}
		_ = json.NewEncoder(w).Encode(doc)
	}))
	defer srv.Close()
	id := srv.URL + "/oauth/client.json"

	// The guard refuses a loopback address.
	if _, err := oauth.FetchDocument(f.c.OAuth, id); err == nil {
		t.Error("a client metadata document on a loopback address was fetched")
	}

	oauth.UseDocumentTransport(f.c.OAuth, srv.Client().Transport)
	doc = map[string]any{"client_id": id, "client_name": "Claude", "redirect_uris": []string{redirect}, "token_endpoint_auth_method": "none"}
	c, err := oauth.FetchDocument(f.c.OAuth, id)
	if err != nil || c.Name != "Claude" || !c.Document {
		t.Errorf("fetch = %+v, %v", c, err)
	}
	// ChatGPT prefers to authenticate with a key, and can do without: it
	// connects as a public client.
	doc = map[string]any{"client_id": id, "client_name": "ChatGPT", "redirect_uris": []string{"https://chatgpt.com/connector_platform_oauth_redirect"}}
	doc["token_endpoint_auth_method"] = "private_key_jwt"
	doc["token_endpoint_auth_methods_supported"] = []string{"none", "private_key_jwt"}
	doc["jwks_uri"] = "https://chatgpt.com/oauth/jwks.json"
	if c, err := oauth.FetchDocument(f.c.OAuth, id); err != nil || c.Name != "ChatGPT" {
		t.Errorf("fetch ChatGPT's document = %+v, %v", c, err)
	}
	for name, change := range map[string]func(){
		"another client ID": func() { doc["client_id"] = srv.URL + "/other.json" },
		"a secret method":   func() { doc["token_endpoint_auth_method"] = "client_secret_basic" },
		"a secret":          func() { doc["client_secret"] = "s3cret" },
		"a key only": func() {
			doc["token_endpoint_auth_method"], doc["token_endpoint_auth_methods_supported"] = "private_key_jwt", []string{"private_key_jwt"}
		},
		"no redirect URI": func() { delete(doc, "redirect_uris") },
		"too large":       func() { doc["padding"] = strings.Repeat("x", 17<<10) },
		"a redirect":      func() { redirectTo = srv.URL + "/elsewhere" },
	} {
		doc = map[string]any{"client_id": id, "client_name": "Claude", "redirect_uris": []string{redirect}}
		redirectTo = ""
		change()
		if _, err := oauth.FetchDocument(f.c.OAuth, id); err == nil {
			t.Errorf("a document with %s was accepted", name)
		}
	}

	// A connection through a metadata document shows its host.
	doc, redirectTo = map[string]any{"client_id": id, "client_name": "Claude", "redirect_uris": []string{redirect}}, ""
	req := f.request(id)
	out, err := f.invoke(f.olivier, "describe_authorization", oauth.DescribeInput{Request: req})
	if err != nil {
		t.Fatal(err)
	}
	if a := out.(oauth.Authorization); a.ClientHost != strings.TrimPrefix(srv.URL, "https://") {
		t.Errorf("describe_authorization = %+v", a)
	}
	code := f.approve(f.olivier, oauth.ApproveInput{Request: req, Scope: oauth.Read, AllRooms: true})
	if status, out := f.exchange(id, code); status != http.StatusOK {
		t.Errorf("exchange for a client known by its document = %d %v", status, out)
	}
	// A client that authenticates with a key names itself in its assertion,
	// without client_id.
	code = f.approve(f.olivier, oauth.ApproveInput{Request: req, Scope: oauth.Read, AllRooms: true})
	claims := base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"` + id + `","sub":"` + id + `","aud":"` + issuer + `/oauth/token"}`))
	status, out := f.token(url.Values{"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {redirect},
		"code_verifier": {verifier}, "client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion": {"eyJhbGciOiJSUzI1NiJ9." + claims + ".c2lnbmF0dXJl"}})
	if status != http.StatusOK {
		t.Errorf("exchange with a client assertion = %d %v", status, out)
	}
	if ua, _ := agent.Load().(string); !strings.HasPrefix(ua, "Cahut/") {
		t.Errorf("documents are fetched as %q", ua)
	}

	// When the document no longer answers, as after a restart while its
	// host is down, its last valid copy stands in, for 30 days.
	describe := func() error {
		oauth.ForgetDocuments(f.c.OAuth)
		_, err := f.invoke(f.olivier, "describe_authorization", oauth.DescribeInput{Request: req})
		return err
	}
	gone.Store(true)
	if err := describe(); err != nil {
		t.Errorf("describe_authorization while the document answers 404: %v", err)
	}
	f.now = f.now.Add(31 * 24 * time.Hour)
	if err := describe(); !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "404") {
		t.Errorf("describe_authorization with a copy older than 30 days: %v", err)
	}
	// A document that answers but is invalid is refused, whatever its copy.
	gone.Store(false)
	if err := describe(); err != nil {
		t.Fatalf("describe_authorization once the document is back: %v", err)
	}
	doc["client_id"] = srv.URL + "/other.json"
	if err := describe(); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("describe_authorization with an invalid document and a valid copy: %v", err)
	}
}

func TestPublicAddresses(t *testing.T) {
	for ip, want := range map[string]bool{
		"160.79.104.10":   true,
		"2606:4700::1111": true,
		"127.0.0.1":       false,
		"10.0.0.1":        false,
		"172.16.0.1":      false,
		"192.168.1.1":     false,
		"169.254.169.254": false, // cloud metadata
		"100.64.0.1":      false,
		"0.0.0.0":         false,
		"::1":             false,
		"fe80::1":         false,
		"fc00::1":         false,
		"::ffff:10.0.0.1": false,
		"64:ff9b::a00:1":  false,
		"224.0.0.1":       false,
	} {
		if got := oauth.Public(ip); got != want {
			t.Errorf("public(%s) = %v, want %v", ip, got, want)
		}
	}
}
