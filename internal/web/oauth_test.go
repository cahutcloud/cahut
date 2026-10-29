package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/oauth"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/server"
)

// oauthFixture serves the Cahut at localhost on a port of its own, as a
// server does on port 443, so that an OAuth client finds its way from the
// MCP endpoint to the authorization server and the console.
type oauthFixture struct {
	base, console string // http://localhost:<port>/
	session       string // camille's session in the console
	web           *Server
}

func listen(t *testing.T) (net.Listener, string) {
	t.Helper()
	l, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(l.Addr().String())
	return l, port
}

func setupOAuth(t *testing.T) *oauthFixture {
	t.Helper()
	ctx := context.Background()
	main, mainPort := listen(t)
	srv, err := server.Open(ctx, server.Config{
		DataDir:    t.TempDir(),
		BaseURL:    func(domain string) string { return "http://" + domain + ":" + mainPort + "/" },
		ConsoleURL: func(domain string) string { return "http://" + domain + ":" + mainPort + "/console/" },
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { srv.Close() })
	for _, domain := range []string{"localhost", "beta.localhost"} {
		if _, err := operation.Do(ctx, srv.Registry, srv.AddDomain, admin, server.AddDomainInput{Domain: domain}); err != nil {
			t.Fatal(err)
		}
	}
	w := New(srv, Config{Dev: true, DevAddr: "localhost:" + mainPort}, slog.New(slog.DiscardHandler))
	t.Cleanup(func() { w.Close() })
	s := &http.Server{Handler: w, ReadHeaderTimeout: 10 * time.Second}
	go s.Serve(main) //nolint:errcheck // ends with Close
	t.Cleanup(func() { s.Close() })

	c, err := w.site(ctx, "localhost")
	if err != nil {
		t.Fatal(err)
	}
	u, err := operation.Do(ctx, c.cahut.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: "camille", Name: "Camille"})
	if err != nil {
		t.Fatal(err)
	}
	visitor := operation.Actor{Kind: operation.Visitor, Via: "api", Domain: "localhost"}
	session, err := operation.Do(ctx, c.cahut.Registry, account.ActivateAccount, visitor,
		account.ActivateInput{Code: u.Activation.Code, Password: password})
	if err != nil {
		t.Fatal(err)
	}
	person := operation.Actor{Kind: operation.Person, Via: "api", Domain: "localhost", UserID: u.ID, Handle: "camille", Session: 1}
	if _, err := operation.Do(ctx, c.cahut.Registry, room.CreateRoom, person, room.CreateRoomInput{Title: "Marketing"}); err != nil {
		t.Fatal(err)
	}
	return &oauthFixture{
		base:    "http://localhost:" + mainPort + "/",
		console: "http://localhost:" + mainPort + "/console/",
		session: session.Token,
		web:     w,
	}
}

// api calls an operation of the console's API with camille's session.
func (f *oauthFixture) api(t *testing.T, name string, in any) map[string]any {
	t.Helper()
	body, _ := json.Marshal(in)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, f.console+"api/"+name, strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+f.session)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("%s: %d %v %v", name, resp.StatusCode, out, err)
	}
	return out
}

// consent plays the console's consent screen for camille: it reads the
// authorization request from its address, and grants the scope and the
// Rooms.
func (f *oauthFixture) consent(t *testing.T, scope string, rooms ...string) auth.AuthorizationCodeFetcher {
	return func(_ context.Context, args *auth.AuthorizationArgs) (*auth.AuthorizationResult, error) {
		u, err := url.Parse(args.URL)
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(args.URL, f.console+"oauth/authorize?") {
			return nil, fmt.Errorf("the client was sent to %s", args.URL)
		}
		q := u.Query()
		request := oauth.Request{
			ResponseType: q.Get("response_type"), ClientID: q.Get("client_id"), RedirectURI: q.Get("redirect_uri"),
			Scope: q.Get("scope"), State: q.Get("state"), CodeChallenge: q.Get("code_challenge"),
			CodeChallengeMethod: q.Get("code_challenge_method"), Resource: q.Get("resource"),
		}
		described := f.api(t, "describe_authorization", oauth.DescribeInput{Request: request})
		if described["redirect"] != nil {
			return nil, fmt.Errorf("the request is invalid: %v", described["error"])
		}
		approved := f.api(t, "approve_authorization", oauth.ApproveInput{Request: request, Scope: scope, Rooms: rooms, AllRooms: len(rooms) == 0})
		back, err := url.Parse(approved["redirect"].(string))
		if err != nil {
			return nil, err
		}
		return &auth.AuthorizationResult{Code: back.Query().Get("code"), State: back.Query().Get("state"), Iss: back.Query().Get("iss")}, nil
	}
}

// connect connects an MCP client that authorizes over OAuth, registering
// dynamically.
func (f *oauthFixture) connect(t *testing.T, fetch auth.AuthorizationCodeFetcher) (*mcp.ClientSession, error) {
	t.Helper()
	handler, err := auth.NewAuthorizationCodeHandler(&auth.AuthorizationCodeHandlerConfig{
		DynamicClientRegistrationConfig: &auth.DynamicClientRegistrationConfig{
			Metadata: &oauthex.ClientRegistrationMetadata{
				ClientName:   "Test client",
				RedirectURIs: []string{"http://127.0.0.1:1/callback"},
				GrantTypes:   []string{"authorization_code", "refresh_token"},
			},
		},
		RedirectURL:              "http://127.0.0.1:1/callback",
		AuthorizationCodeFetcher: fetch,
	})
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint:             f.base + "mcp",
		OAuthHandler:         handler,
		DisableStandaloneSSE: true,
		MaxRetries:           -1,
	}, nil)
	if err == nil {
		t.Cleanup(func() { session.Close() })
	}
	return session, err
}

// TestOAuthWithTheSDKClient connects the OAuth client of the MCP Go SDK,
// as an AI client would: from the MCP endpoint to the metadata, the
// registration, the consent and the tokens.
func TestOAuthWithTheSDKClient(t *testing.T) {
	f := setupOAuth(t)
	ctx := context.Background()

	// An AI limited to marketing and to reading.
	session, err := f.connect(t, f.consent(t, oauth.Read, "marketing"))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tool := range tools.Tools {
		names = append(names, tool.Name)
	}
	want := []string{"get_account", "get_cahut", "get_knock", "get_room", "get_window", "list_invitations", "list_knocks", "list_logbook", "list_rooms", "list_users", "list_windows", "set_ai_reach"}
	if !slices.Equal(names, want) {
		t.Errorf("tools of an AI that may only read = %v, want %v", names, want)
	}
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_rooms", Arguments: map[string]any{}})
	if err != nil || res.IsError {
		t.Fatalf("list_rooms: %v %v", res, err)
	}
	if rooms := res.StructuredContent.(map[string]any)["rooms"].([]any); len(rooms) != 1 {
		t.Errorf("an AI limited to marketing sees %v", rooms)
	}

	// An AI with write access to every Room publishes. It is let in once:
	// asked again, the person refuses.
	consented := false
	once := func(ctx context.Context, args *auth.AuthorizationArgs) (*auth.AuthorizationResult, error) {
		if consented {
			return nil, errors.New("the person refuses")
		}
		consented = true
		return f.consent(t, oauth.Write)(ctx, args)
	}
	writer, err := f.connect(t, once)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	res, err = writer.CallTool(ctx, &mcp.CallToolParams{Name: "publish_window", Arguments: map[string]any{
		"title": "Q3", "kind": "deck", "room": "marketing", "declaration": map[string]any{"scripts": false}, "files": []any{map[string]any{"path": "index.html", "content": "<h1>Q3</h1>"}},
	}})
	if err != nil || res.IsError {
		t.Fatalf("publish_window: %v %v", res, err)
	}

	// The Logbook names the connection and its client.
	logbook := f.api(t, "list_logbook", map[string]any{"limit": 1})["entries"].([]any)[0].(map[string]any)
	if logbook["operation"] != "publish_window" || logbook["connection_id"] == nil ||
		!strings.HasPrefix(logbook["client"].(string), "Test client (test-client 1.0)") {
		t.Errorf("last entry = %v", logbook)
	}

	// Once disconnected, the AI is refused, and must ask the person again.
	connections := f.api(t, "list_connected_ais", map[string]any{})["connections"].([]any)
	if len(connections) != 2 {
		t.Fatalf("connected AIs = %v", connections)
	}
	f.api(t, "disconnect_ai", map[string]any{"id": connections[0].(map[string]any)["id"]})
	if _, err := writer.ListTools(ctx, nil); err == nil || !strings.Contains(err.Error(), "the person refuses") {
		t.Errorf("a disconnected AI lists its tools: %v", err)
	}
}

// TestOAuthDiscovery checks what a client reads before it authorizes.
func TestOAuthDiscovery(t *testing.T) {
	f := setupOAuth(t)
	get := func(method, address string) (response, map[string]any) {
		t.Helper()
		req, _ := http.NewRequestWithContext(context.Background(), method, address, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var out map[string]any
		body, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &out)
		return response{StatusCode: resp.StatusCode, Header: resp.Header}, out
	}
	resp, _ := get(http.MethodPost, f.base+"mcp")
	wantMeta := f.base + ".well-known/oauth-protected-resource/mcp"
	if resp.StatusCode != http.StatusUnauthorized || !strings.Contains(resp.Header.Get("WWW-Authenticate"), `resource_metadata="`+wantMeta+`"`) {
		t.Errorf("MCP without a token: %d %q", resp.StatusCode, resp.Header.Get("WWW-Authenticate"))
	}
	// The servers of AI clients reach everything on the usual port, and the
	// consent screen is a page of the console. See
	// docs/adr/0017-oauth-endpoints-on-port-443.md and
	// docs/adr/0020-console-under-a-path-of-port-443.md.
	issuer := strings.TrimSuffix(f.base, "/")
	for _, address := range []string{wantMeta, f.base + ".well-known/oauth-protected-resource"} {
		resp, meta := get(http.MethodGet, address)
		if resp.StatusCode != http.StatusOK || meta["resource"] != f.base+"mcp" || fmt.Sprint(meta["authorization_servers"]) != "["+issuer+"]" {
			t.Errorf("%s: %d %v", address, resp.StatusCode, meta)
		}
	}
	resp, meta := get(http.MethodGet, f.base+".well-known/oauth-authorization-server")
	if resp.StatusCode != http.StatusOK || meta["issuer"] != issuer || resp.Header.Get("Access-Control-Allow-Origin") != "*" ||
		meta["authorization_endpoint"] != f.console+"oauth/authorize" || meta["token_endpoint"] != f.base+"oauth/token" {
		t.Errorf("authorization server metadata: %d %v", resp.StatusCode, meta)
	}
	// Cahut is no OpenID provider: a client that probes for one reads a 404,
	// not a page.
	if resp, _ := get(http.MethodGet, f.base+".well-known/openid-configuration"); resp.StatusCode != http.StatusNotFound ||
		resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("OpenID Connect discovery: %d %v", resp.StatusCode, resp.Header)
	}
	if resp, _ := get(http.MethodOptions, f.base+"oauth/token"); resp.StatusCode != http.StatusNoContent ||
		resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("preflight of the token endpoint: %d %v", resp.StatusCode, resp.Header)
	}
	// The consent screen is a page of the console, which no other page may
	// frame. Tests build no console: the page is its notice.
	if resp, _ := get(http.MethodGet, f.console+"oauth/authorize?client_id=x"); !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") ||
		!strings.Contains(resp.Header.Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Errorf("the consent screen: %d %v", resp.StatusCode, resp.Header)
	}
	// Each Cahut protects its own MCP endpoint, so that a token of one
	// opens nothing in another.
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet,
		strings.Replace(f.base, "localhost", "127.0.0.1", 1)+".well-known/oauth-protected-resource/mcp", nil)
	req.Host = strings.TrimSuffix(strings.TrimPrefix(strings.Replace(f.base, "localhost", "beta.localhost", 1), "http://"), "/")
	betaResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer betaResp.Body.Close()
	_ = json.NewDecoder(betaResp.Body).Decode(&meta)
	if want := "http://" + req.Host + "/mcp"; meta["resource"] != want {
		t.Errorf("the resource of beta.localhost: %v, want %s", meta["resource"], want)
	}
}

// TestRegistrationsAreLimited checks that registration is limited by
// address.
func TestRegistrationsAreLimited(t *testing.T) {
	f := setupOAuth(t)
	status := 0
	for range RegisterBurst + 1 {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, f.base+"oauth/register",
			strings.NewReader(`{"redirect_uris": ["http://127.0.0.1/callback"]}`))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		status = resp.StatusCode
	}
	if status != http.StatusTooManyRequests {
		t.Errorf("registration %d = %d, want 429", RegisterBurst+1, status)
	}
}

// TestMCPLog checks what the journal says of a refused request to the MCP
// endpoint: its JSON-RPC method, the answer, and the kind of credential,
// never the credential.
func TestMCPLog(t *testing.T) {
	f := setup(t)
	var journal bytes.Buffer
	f.web.log = slog.New(slog.NewTextHandler(&journal, nil))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, f.ts.URL+"/mcp",
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))
	resp, err := f.client("localhost", "cahut_oat_secretpart").Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	line := journal.String()
	for _, want := range []string{"msg=mcp", "rpc=initialize", "status=401", `credential="Bearer cahut_oat_"`, "answer="} {
		if !strings.Contains(line, want) {
			t.Errorf("the journal says %q, without %q", line, want)
		}
	}
	if strings.Contains(line, "secretpart") {
		t.Errorf("the journal holds the token: %q", line)
	}
	for _, tt := range []struct{ header, want string }{
		{"", "none"},
		{"Bearer " + account.TokenPrefix + "abc", "Bearer " + account.TokenPrefix},
		{"DPoP eyJhbGciOiJSUzI1", "DPoP, 16 characters"},
	} {
		if got := credentialKind(tt.header); got != tt.want {
			t.Errorf("credentialKind(%q) = %q, want %q", tt.header, got, tt.want)
		}
	}
}
