package web

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
	"cahut.cloud/internal/window"
)

var admin = operation.Actor{Kind: operation.Admin, Via: "cli"}

type fixture struct {
	srv   *server.Server
	web   *Server
	ts    *httptest.Server
	token string // olivier's token in the Cahut at localhost
}

// setup serves two Cahuts, at localhost and beta.localhost. In the first,
// olivier has a token for Claude Code.
func setup(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	cfg := Config{}
	srv, err := server.Open(ctx, server.Config{DataDir: t.TempDir(), BaseURL: cfg.BaseURL})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { srv.Close() })
	for _, domain := range []string{"localhost", "beta.localhost"} {
		if _, err := operation.Do(ctx, srv.Registry, srv.AddDomain, admin, server.AddDomainInput{Domain: domain}); err != nil {
			t.Fatal(err)
		}
	}
	c, err := srv.OpenCahut(ctx, "localhost")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, err := operation.Do(ctx, c.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: "olivier", Name: "Olivier"}); err != nil {
		t.Fatal(err)
	}
	tok, err := operation.Do(ctx, c.Registry, account.CreateToken, admin, account.CreateTokenInput{User: "olivier", Name: "Claude Code"})
	if err != nil {
		t.Fatal(err)
	}
	w := New(srv, cfg, slog.New(slog.DiscardHandler))
	t.Cleanup(func() { w.Close() })
	ts := httptest.NewServer(w)
	t.Cleanup(ts.Close)
	return &fixture{srv: srv, web: w, ts: ts, token: tok.Token}
}

// rewrite sends every request to the test server under another host name,
// with a bearer token when one is set.
type rewrite struct {
	host, token string
	next        http.RoundTripper
}

func (rt rewrite) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Host = rt.host
	if rt.token != "" {
		r.Header.Set("Authorization", "Bearer "+rt.token)
	}
	return rt.next.RoundTrip(r)
}

func (f *fixture) client(host, token string) *http.Client {
	return &http.Client{
		Transport:     rewrite{host: host, token: token, next: f.ts.Client().Transport},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// response is what get returns of a response.
type response struct {
	StatusCode int
	Header     http.Header
}

func (f *fixture) get(t *testing.T, host, method, path string) (response, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, f.ts.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := f.client(host, "").Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return response{StatusCode: resp.StatusCode, Header: resp.Header}, string(body)
}

func (f *fixture) connect(t *testing.T, host, token string) *mcp.ClientSession {
	t.Helper()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint:             f.ts.URL + "/mcp",
		HTTPClient:           f.client(host, token),
		DisableStandaloneSSE: true,
		MaxRetries:           -1,
	}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

// open returns the Cahut at domain, as the web server opened it.
func (f *fixture) open(t *testing.T, domain string) *cahut.Cahut {
	t.Helper()
	st, err := f.web.site(context.Background(), domain)
	if err != nil {
		t.Fatal(err)
	}
	return st.cahut
}

func TestPublishOverMCP(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	session := f.connect(t, "localhost", f.token)

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tool := range tools.Tools {
		names = append(names, tool.Name)
	}
	if !slices.Equal(names, []string{"get_window", "list_windows", "publish_window"}) {
		t.Errorf("tools = %v, want the Window operations only", names)
	}

	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "publish_window", Arguments: map[string]any{
		"title": "Q3 review",
		"kind":  "deck",
		"files": []map[string]string{
			{"path": "index.html", "content": "<h1>Q3</h1>"},
			{"path": "css/deck.css", "content": "h1 { color: teal }"},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("publish_window failed: %v", res.Content[0].(*mcp.TextContent).Text)
	}
	w, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content = %#v", res.StructuredContent)
	}
	if w["name"] != "q3-review" || w["room"] != "olivier" || w["url"] != "https://localhost/olivier/q3-review/" {
		t.Errorf("window = %v", w)
	}

	resp, body := f.get(t, "localhost", http.MethodGet, "/olivier/q3-review/")
	if resp.StatusCode != http.StatusOK || body != "<h1>Q3</h1>" {
		t.Errorf("GET /olivier/q3-review/ = %d %q", resp.StatusCode, body)
	}
	if csp := resp.Header.Get("Content-Security-Policy"); !strings.HasPrefix(csp, "sandbox ") || strings.Contains(csp, "allow-same-origin") {
		t.Errorf("Content-Security-Policy = %q, want a sandbox without allow-same-origin", csp)
	}
	if acao := resp.Header.Get("Access-Control-Allow-Origin"); acao != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want *", acao)
	}
	resp, body = f.get(t, "localhost", http.MethodGet, "/olivier/q3-review/css/deck.css")
	if resp.StatusCode != http.StatusOK || !strings.Contains(body, "teal") || !strings.HasPrefix(resp.Header.Get("Content-Security-Policy"), "sandbox ") {
		t.Errorf("GET /olivier/q3-review/css/deck.css = %d %q %v", resp.StatusCode, body, resp.Header)
	}
	resp, _ = f.get(t, "localhost", http.MethodGet, "/olivier/q3-review?x=1")
	if resp.StatusCode != http.StatusPermanentRedirect || resp.Header.Get("Location") != "/olivier/q3-review/?x=1" {
		t.Errorf("GET /olivier/q3-review = %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	for path, want := range map[string]int{
		"/olivier/q3-review/css/":    http.StatusNotFound,
		"/olivier/q3-review/nope.js": http.StatusNotFound,
		"/olivier/nope/":             http.StatusNotFound,
		"/olivier/":                  http.StatusNotFound,
		"/Olivier/q3-review/":        http.StatusNotFound,
	} {
		if resp, _ := f.get(t, "localhost", http.MethodGet, path); resp.StatusCode != want {
			t.Errorf("GET %s = %d, want %d", path, resp.StatusCode, want)
		}
	}
	if resp, _ := f.get(t, "localhost", http.MethodPost, "/olivier/q3-review/"); resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("POST /olivier/q3-review/ = %d, want 405", resp.StatusCode)
	}
	// The other Cahut does not serve the Windows of the first.
	if resp, _ := f.get(t, "beta.localhost", http.MethodGet, "/olivier/q3-review/"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET beta.localhost/olivier/q3-review/ = %d, want 404", resp.StatusCode)
	}

	res, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "publish_window", Arguments: map[string]any{
		"title": "Broken", "kind": "deck", "files": []map[string]string{{"path": "../x", "content": ""}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "invalid input") {
		t.Errorf("bad publication: %+v", res)
	}

	entries, err := logbook.List(ctx, f.open(t, "localhost").DB.Read, 10)
	if err != nil {
		t.Fatal(err)
	}
	ok = false
	for _, e := range entries {
		if e.Operation == "publish_window" && e.Outcome == logbook.OK {
			ok = e.Target == "olivier/q3-review" && e.Client == "Claude Code (test-client 1.0)" && e.Via == "mcp" && e.UserHandle == "olivier"
			if !ok {
				t.Errorf("logbook entry = %+v", e)
			}
		}
	}
	if !ok {
		t.Errorf("no publish_window entry in the Logbook: %+v", entries)
	}
}

func TestMCPNeedsAValidToken(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	post := func(host, token string) int {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.ts.URL+"/mcp", strings.NewReader(`{}`))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := f.client(host, token).Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if code := post("localhost", ""); code != http.StatusUnauthorized {
		t.Errorf("no token: %d, want 401", code)
	}
	if code := post("localhost", "cahut_pat_wrong"); code != http.StatusUnauthorized {
		t.Errorf("wrong token: %d, want 401", code)
	}
	if code := post("beta.localhost", f.token); code != http.StatusUnauthorized {
		t.Errorf("token of another Cahut: %d, want 401", code)
	}

	c := f.open(t, "localhost")
	if _, err := operation.Do(ctx, c.Registry, account.RevokeToken, admin, account.RevokeTokenInput{ID: 1}); err != nil {
		t.Fatal(err)
	}
	if code := post("localhost", f.token); code != http.StatusUnauthorized {
		t.Errorf("revoked token: %d, want 401", code)
	}
}

func TestRouting(t *testing.T) {
	f := setup(t)
	tests := []struct {
		host, path string
		want       int
	}{
		{"localhost", "/", http.StatusOK},
		{"beta.localhost", "/", http.StatusOK},
		{"localhost", "/-/healthz", http.StatusOK},
		{"localhost", "/-/nope", http.StatusNotFound},
		{"localhost", "/console", http.StatusNotFound},
		{"nope.localhost", "/", http.StatusMisdirectedRequest},
		{"example.org", "/", http.StatusMisdirectedRequest},
	}
	for _, tt := range tests {
		if resp, _ := f.get(t, tt.host, http.MethodGet, tt.path); resp.StatusCode != tt.want {
			t.Errorf("GET %s%s = %d, want %d", tt.host, tt.path, resp.StatusCode, tt.want)
		}
	}
	if _, body := f.get(t, "beta.localhost", http.MethodGet, "/"); !strings.Contains(body, "beta.localhost, a Cahut") {
		t.Errorf("GET beta.localhost/ = %q", body)
	}
}

func TestBaseURL(t *testing.T) {
	if got := (Config{}).BaseURL("cahut.example.com"); got != "https://cahut.example.com/" {
		t.Errorf("BaseURL = %q", got)
	}
	if got := (Config{Dev: true, DevAddr: "localhost:8080"}).BaseURL("beta.localhost"); got != "http://beta.localhost:8080/" {
		t.Errorf("dev BaseURL = %q", got)
	}
}

// TestLargePublication publishes, over MCP, a Window close to the size
// limit, then one above it. What an AI client can send in one tool call is
// smaller; this checks that the server is not the bottleneck.
func TestLargePublication(t *testing.T) {
	if testing.Short() {
		t.Skip("sends tens of megabytes")
	}
	f := setup(t)
	ctx := context.Background()
	session := f.connect(t, "localhost", f.token)
	publish := func(size int) *mcp.CallToolResult {
		t.Helper()
		image := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0xa5}, size))
		res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "publish_window", Arguments: map[string]any{
			"title": "Large", "kind": "site", "files": []map[string]string{
				{"path": "index.html", "content": `<img src="big.bin">`},
				{"path": "big.bin", "content": image, "encoding": "base64"},
			},
		}})
		if err != nil {
			t.Fatalf("publish %d bytes: %v", size, err)
		}
		return res
	}

	start := time.Now()
	if res := publish(window.MaxSize - 1024); res.IsError {
		t.Fatalf("publication under the limit refused: %v", res.Content[0].(*mcp.TextContent).Text)
	}
	t.Logf("%d MiB published in %v", window.MaxSize>>20, time.Since(start).Round(time.Millisecond))

	res := publish(window.MaxSize + 1)
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "more than 32 MiB") {
		t.Errorf("publication over the limit: %+v", res)
	}
}

func TestRedirectToHTTPS(t *testing.T) {
	f := setup(t)
	rd := redirectToHTTPS{srv: f.srv}
	tests := []struct {
		host, want string
		code       int
	}{
		{"localhost", "https://localhost/a?b=c", http.StatusPermanentRedirect},
		{"beta.localhost:80", "https://beta.localhost/a?b=c", http.StatusPermanentRedirect},
		{"evil.example.org", "", http.StatusMisdirectedRequest},
	}
	for _, tt := range tests {
		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+tt.host+"/a?b=c", nil)
		w := httptest.NewRecorder()
		rd.ServeHTTP(w, r)
		if w.Code != tt.code || w.Header().Get("Location") != tt.want {
			t.Errorf("%s: %d %q, want %d %q", tt.host, w.Code, w.Header().Get("Location"), tt.code, tt.want)
		}
	}
}
