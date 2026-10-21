package server

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
	"cahut.cloud/internal/home"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/window"
)

type fixture struct {
	home  *home.Home
	ts    *httptest.Server
	token string
}

// setup serves a home for olivier, who has a token for Claude Code. The
// home's domain is localhost, so its Windows are at <name>.localhost.
func setup(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	cfg := Config{Domain: "localhost"}
	h, err := home.Open(ctx, home.Config{DataDir: t.TempDir(), WindowURL: cfg.WindowURL()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { h.Close() })
	admin := operation.Actor{Kind: operation.Admin, Via: "cli"}
	if _, err := operation.Do(ctx, h.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: "olivier", Name: "Olivier"}); err != nil {
		t.Fatal(err)
	}
	tok, err := operation.Do(ctx, h.Registry, account.CreateToken, admin, account.CreateTokenInput{User: "olivier", Name: "Claude Code"})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(New(h, cfg, slog.New(slog.DiscardHandler)))
	t.Cleanup(ts.Close)
	return &fixture{home: h, ts: ts, token: tok.Token}
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
	return &http.Client{Transport: rewrite{host: host, token: token, next: f.ts.Client().Transport}}
}

func (f *fixture) get(t *testing.T, host, method, path string) (int, string) {
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
	return resp.StatusCode, string(body)
}

func (f *fixture) connect(t *testing.T, token string) *mcp.ClientSession {
	t.Helper()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint:             f.ts.URL + "/mcp",
		HTTPClient:           f.client("localhost", token),
		DisableStandaloneSSE: true,
		MaxRetries:           -1,
	}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func TestPublishOverMCP(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	session := f.connect(t, f.token)

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
	name := w["name"].(string)
	if w["url"] != "https://"+name+".localhost/" {
		t.Errorf("url = %v", w["url"])
	}

	host := name + ".localhost"
	if code, body := f.get(t, host, http.MethodGet, "/"); code != http.StatusOK || body != "<h1>Q3</h1>" {
		t.Errorf("GET / = %d %q", code, body)
	}
	if code, body := f.get(t, host, http.MethodGet, "/css/deck.css"); code != http.StatusOK || !strings.Contains(body, "teal") {
		t.Errorf("GET /css/deck.css = %d %q", code, body)
	}
	for path, want := range map[string]int{"/css/": http.StatusNotFound, "/nope.js": http.StatusNotFound} {
		if code, _ := f.get(t, host, http.MethodGet, path); code != want {
			t.Errorf("GET %s = %d, want %d", path, code, want)
		}
	}
	if code, _ := f.get(t, host, http.MethodPost, "/"); code != http.StatusMethodNotAllowed {
		t.Errorf("POST / = %d, want 405", code)
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

	entries, err := logbook.List(ctx, f.home.DB.Read, 10)
	if err != nil {
		t.Fatal(err)
	}
	ok = false
	for _, e := range entries {
		if e.Operation == "publish_window" && e.Outcome == logbook.OK {
			ok = e.Target == name && e.Client == "Claude Code (test-client 1.0)" && e.Via == "mcp" && e.UserHandle == "olivier"
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
	post := func(token string) int {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.ts.URL+"/mcp", strings.NewReader(`{}`))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := f.client("localhost", token).Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if code := post(""); code != http.StatusUnauthorized {
		t.Errorf("no token: %d, want 401", code)
	}
	if code := post("cahut_pat_wrong"); code != http.StatusUnauthorized {
		t.Errorf("wrong token: %d, want 401", code)
	}

	admin := operation.Actor{Kind: operation.Admin, Via: "cli"}
	if _, err := operation.Do(ctx, f.home.Registry, account.RevokeToken, admin, account.RevokeTokenInput{ID: 1}); err != nil {
		t.Fatal(err)
	}
	if code := post(f.token); code != http.StatusUnauthorized {
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
		{"localhost", "/healthz", http.StatusOK},
		{"localhost", "/console", http.StatusNotFound},
		{"nope-abcdefgh.localhost", "/", http.StatusNotFound},
		{"example.org", "/", http.StatusMisdirectedRequest},
	}
	for _, tt := range tests {
		if code, _ := f.get(t, tt.host, http.MethodGet, tt.path); code != tt.want {
			t.Errorf("GET %s%s = %d, want %d", tt.host, tt.path, code, tt.want)
		}
	}
}

func TestWindowURL(t *testing.T) {
	if got := (Config{Domain: "home.example.com"}).WindowURL()("q3"); got != "https://q3.home.example.com/" {
		t.Errorf("WindowURL = %q", got)
	}
	if got := (Config{Dev: true, DevAddr: "localhost:8080"}).WindowURL()("q3"); got != "http://q3.localhost:8080/" {
		t.Errorf("dev WindowURL = %q", got)
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
	session := f.connect(t, f.token)
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
	rd := redirectToHTTPS{domain: "home.example.com"}
	tests := []struct {
		host, want string
		code       int
	}{
		{"home.example.com", "https://home.example.com/a?b=c", http.StatusPermanentRedirect},
		{"q3-abcdefgh.home.example.com:80", "https://q3-abcdefgh.home.example.com/a?b=c", http.StatusPermanentRedirect},
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
