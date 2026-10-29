package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"

	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/oauth"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
)

// The acceptance of the V0 replays four of the five criteria of the README
// on a Cahut of a test server, through what people and AIs use: the
// console's API, MCP with OAuth, and the addresses of Windows. The first
// criterion, the installation, is the Install job of CI. See
// docs/recette-v0.md.

// undo names, for each operation offered to AIs that changes something
// and does not knock first, the operation that undoes it. An AI that calls
// an operation knocks first, or a person can take it back.
var undo = map[string]string{
	"create_room":       "delete_room",
	"publish_window":    "trash_window",
	"update_window":     "rollback_window",
	"retitle_window":    "retitle_window",
	"rename_window":     "rename_window",
	"rollback_window":   "rollback_window",
	"trash_window":      "restore_window",
	"restore_window":    "trash_window",
	"set_window_access": "set_window_access", // making a Window public knocks first
	"update_room":       "update_room",
	"set_room_home":     "set_room_home",
	"set_ai_reach":      "set_ai_reach", // widening the reach knocks first
	// The e-mail stays sent, and the link it carries opens nothing more.
	"invite_to_window":  "revoke_invitation",
	"revoke_invitation": "invite_to_window",
	// The link given stays where it went, and opens nothing more.
	"create_share_link": "revoke_share_link",
	"revoke_share_link": "create_share_link",
}

// TestEveryAIOperationIsReversibleOrApproved checks the rule of
// governance on the registry: each operation offered to AIs changes
// nothing, knocks first, or has its undo. An operation that is in none of
// the three fails it.
func TestEveryAIOperationIsReversibleOrApproved(t *testing.T) {
	f := setup(t)
	ops := map[string]operation.Info{}
	for _, op := range f.open(t, "localhost").Registry.Ops() {
		ops[op.Name] = op
	}
	for _, op := range ops {
		if !op.ForAIs(true, false) || op.ReadOnly || op.KnockFirst {
			continue
		}
		back, ok := undo[op.Name]
		if !ok {
			t.Errorf("%s changes things for an AI, without knocking first and without an undo: knock first, or name its undo", op.Name)
			continue
		}
		if _, ok := ops[back]; !ok {
			t.Errorf("%s is undone by %s, which is not an operation", op.Name, back)
		}
	}
	for name := range undo {
		if op, ok := ops[name]; !ok || op.ReadOnly || op.KnockFirst || !op.ForAIs(true, false) {
			t.Errorf("undo names %s, which is not an operation of AIs that changes things without knocking first", name)
		}
	}
}

// acceptance is a test server in development mode, at localhost, with one
// Cahut that has nobody yet.
type acceptance struct {
	srv     *server.Server
	web     *Server
	base    string // the address of the Cahut, with its port
	console string
	session string // the console session of the first person
}

func setupAcceptance(t *testing.T) *acceptance {
	t.Helper()
	ctx := context.Background()
	l, port := listen(t)
	srv, err := server.Open(ctx, server.Config{
		DataDir:    t.TempDir(),
		BaseURL:    func(domain string) string { return "http://" + domain + ":" + port + "/" },
		ConsoleURL: func(domain string) string { return "http://" + domain + ":" + port + "/console/" },
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { srv.Close() })
	if _, err := operation.Do(ctx, srv.Registry, srv.AddDomain, admin, server.AddDomainInput{Domain: "localhost"}); err != nil {
		t.Fatal(err)
	}
	w := New(srv, Config{Dev: true, DevAddr: "localhost:" + port}, slog.New(slog.DiscardHandler))
	t.Cleanup(func() { w.Close() })
	hs := &http.Server{Handler: w, ReadHeaderTimeout: 10 * time.Second}
	go hs.Serve(l) //nolint:errcheck // ends with Close
	t.Cleanup(func() { hs.Close() })
	return &acceptance{srv: srv, web: w, base: "http://localhost:" + port + "/", console: "http://localhost:" + port + "/console/"}
}

// api calls an operation of the console's API, with the first person's
// session once there is one, and decodes its answer into out.
func (a *acceptance) api(t *testing.T, route string, in, out any) int {
	t.Helper()
	body, _ := json.Marshal(in)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, a.console+"api/"+route, strings.NewReader(string(body)))
	if a.session != "" {
		req.Header.Set("Authorization", "Bearer "+a.session)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if out != nil && resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(raw, out); err != nil {
			t.Fatalf("%s: %v in %s", route, err, raw)
		}
	}
	if resp.StatusCode != http.StatusOK {
		t.Logf("%s: %d %s", route, resp.StatusCode, raw)
	}
	return resp.StatusCode
}

// consent plays the consent screen of the console, as the first person,
// and grants write access to every Room.
func (a *acceptance) consent(t *testing.T) auth.AuthorizationCodeFetcher {
	return func(_ context.Context, args *auth.AuthorizationArgs) (*auth.AuthorizationResult, error) {
		u, err := url.Parse(args.URL)
		if err != nil || !strings.HasPrefix(args.URL, a.console+"oauth/authorize?") {
			return nil, fmt.Errorf("the client was sent to %s", args.URL)
		}
		q := u.Query()
		request := oauth.Request{
			ResponseType: q.Get("response_type"), ClientID: q.Get("client_id"), RedirectURI: q.Get("redirect_uri"),
			Scope: q.Get("scope"), State: q.Get("state"), CodeChallenge: q.Get("code_challenge"),
			CodeChallengeMethod: q.Get("code_challenge_method"), Resource: q.Get("resource"),
		}
		var described oauth.Authorization
		if status := a.api(t, "describe_authorization", oauth.DescribeInput{Request: request}, &described); status != http.StatusOK || described.Redirect != "" {
			return nil, fmt.Errorf("describe_authorization: %d %s", status, described.Error)
		}
		var approved oauth.Redirect
		if status := a.api(t, "approve_authorization", oauth.ApproveInput{Request: request, Scope: oauth.Write, AllRooms: true}, &approved); status != http.StatusOK {
			return nil, fmt.Errorf("approve_authorization: %d", status)
		}
		back, err := url.Parse(approved.Redirect)
		if err != nil {
			return nil, err
		}
		return &auth.AuthorizationResult{Code: back.Query().Get("code"), State: back.Query().Get("state"), Iss: back.Query().Get("iss")}, nil
	}
}

// connect connects an AI client over OAuth, named software, which
// identifies itself with the metadata document at document when it is set,
// as ChatGPT and Claude do, or registers dynamically, as Le Chat does.
func (a *acceptance) connect(t *testing.T, software, document string) *mcp.ClientSession {
	t.Helper()
	const redirect = "http://127.0.0.1:1/callback"
	cfg := &auth.AuthorizationCodeHandlerConfig{RedirectURL: redirect, AuthorizationCodeFetcher: a.consent(t), RequestRefreshToken: true}
	if document != "" {
		cfg.ClientIDMetadataDocumentConfig = &auth.ClientIDMetadataDocumentConfig{URL: document}
	} else {
		cfg.DynamicClientRegistrationConfig = &auth.DynamicClientRegistrationConfig{Metadata: &oauthex.ClientRegistrationMetadata{
			ClientName: "Le Chat", RedirectURIs: []string{redirect}, GrantTypes: []string{"authorization_code", "refresh_token"},
		}}
	}
	handler, err := auth.NewAuthorizationCodeHandler(cfg)
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: software, Version: "1.0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint: a.base + "mcp", OAuthHandler: handler, DisableStandaloneSSE: true, MaxRetries: -1,
	}, nil)
	if err != nil {
		t.Fatalf("connect %s: %v", software, err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

// useTool calls a tool and returns its structured result, or its error.
func useTool(t *testing.T, s *mcp.ClientSession, tool string, args map[string]any) (map[string]any, string) {
	t.Helper()
	res, err := s.CallTool(context.Background(), &mcp.CallToolParams{Name: tool, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", tool, err)
	}
	if res.IsError {
		return nil, res.Content[0].(*mcp.TextContent).Text
	}
	out, _ := res.StructuredContent.(map[string]any)
	return out, ""
}

var deckHTML = `<!doctype html><html lang="en"><meta charset="utf-8"><title>Why keep our pages</title>
<link rel="stylesheet" href="css/deck.css">
<section class="slide"><h1>Why keep our pages</h1></section>
<section class="slide"><h2>On our own server</h2></section><script src="js/deck.js"></script>`

var acceptanceFrame = regexp.MustCompile(`<iframe src="(/-/frame/[^"]+)"`)

// TestV0Acceptance replays the criteria of the V0 that a test can: the
// first start without a terminal, the first publication by an AI, sharing
// with a Key from the console, two AI clients on one Cahut, and a Logbook
// where every call of an AI either can be undone or was approved.
func TestV0Acceptance(t *testing.T) {
	a := setupAcceptance(t)
	ctx := context.Background()

	// The installation ends with a link to the first start.
	link, err := operation.Do(ctx, a.srv.Registry, a.srv.CreateSetupLink, admin, server.CreateSetupLinkInput{Domain: "localhost"})
	if err != nil {
		t.Fatal(err)
	}

	// Criterion 2. In the browser, the first person chooses their account
	// and connects an AI: nobody touches the server.
	var up server.SetUp
	if status := a.api(t, "server/set_up", server.SetUpInput{Code: link.Code, Handle: "ana", Name: "Ana Lima", Password: password}, &up); status != http.StatusOK {
		t.Fatalf("set_up = %d", status)
	}
	if !up.ServerAdmin {
		t.Error("the first person of the first Cahut is not the administrator of the server")
	}
	a.session = up.Session.Token

	// ChatGPT identifies itself with its metadata document, which prefers
	// private_key_jwt and offers none.
	docs := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{ //nolint:gosec // a client metadata document, which holds no secret
			"client_id": "https://" + r.Host + r.URL.Path, "client_name": "ChatGPT", "client_uri": "https://chatgpt.com/",
			"redirect_uris": []string{"http://127.0.0.1:1/callback"}, "grant_types": []string{"authorization_code", "refresh_token"},
			"token_endpoint_auth_method": "private_key_jwt", "token_endpoint_auth_methods_supported": []string{"none", "private_key_jwt"},
		})
	}))
	defer docs.Close()
	st, err := a.web.site(ctx, "localhost")
	if err != nil {
		t.Fatal(err)
	}
	st.cahut.OAuth.SetDocumentTransport(docs.Client().Transport)
	chatgpt := a.connect(t, "openai-mcp", docs.URL+"/oauth/client.json")

	w, msg := useTool(t, chatgpt, "publish_window", map[string]any{
		"title": "Why keep our pages", "kind": "deck",
		"files":       []map[string]string{{"path": "index.html", "content": deckHTML}},
		"declaration": map[string]any{"scripts": true},
	})
	if msg != "" {
		t.Fatalf("publish_window: %s", msg)
	}
	if w["url"] != a.base+"ana/why-keep-our-pages/" {
		t.Errorf("the address of the Window = %v", w["url"])
	}

	// Criterion 3. Without a terminal, the person gives the deck a Key, and
	// the link opens with it in a private window.
	var set struct{ Key string }
	if status := a.api(t, "set_window_access", map[string]any{"window": "ana/why-keep-our-pages", "access": "key"}, &set); status != http.StatusOK || set.Key == "" {
		t.Fatalf("set_window_access key = %d %q", status, set.Key)
	}
	private := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	read := func(path string, cookies ...*http.Cookie) (int, string) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, a.base+strings.TrimPrefix(path, "/"), nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		resp, err := private.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(body)
	}
	if status, body := read("/ana/why-keep-our-pages/"); status != http.StatusUnauthorized && !strings.Contains(body, "Key") {
		t.Errorf("the Window without its Key = %d, want the page that asks for it", status)
	}
	form := url.Values{"window": {"ana/why-keep-our-pages"}, "key": {set.Key}}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, a.base+"-/key", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := private.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther || len(resp.Cookies()) == 0 {
		t.Fatalf("the Key = %d", resp.StatusCode)
	}
	_, page := read("/ana/why-keep-our-pages/", resp.Cookies()...)
	src := acceptanceFrame.FindStringSubmatch(page)
	if src == nil {
		t.Fatalf("no frame in %s", page)
	}
	if status, body := read(src[1]); status != http.StatusOK || !strings.Contains(body, "On our own server") {
		t.Errorf("the deck with its Key = %d %q", status, body)
	}

	// Criterion 4. A second client, which registers dynamically as Le Chat
	// does, works on the same Window: it updates it, then goes back.
	lechat := a.connect(t, "mistral-le-chat", "")
	if _, msg := useTool(t, lechat, "update_window", map[string]any{
		"window":      "ana/why-keep-our-pages",
		"files":       []map[string]string{{"path": "index.html", "content": strings.Replace(deckHTML, "On our own server", "In its own style", 1)}},
		"declaration": map[string]any{"scripts": true},
	}); msg != "" {
		t.Fatalf("update_window: %s", msg)
	}
	if _, body := read(src[1]); !strings.Contains(body, "In its own style") {
		t.Errorf("the frame does not show the new version: %q", body)
	}
	if _, msg := useTool(t, lechat, "rollback_window", map[string]any{"window": "ana/why-keep-our-pages"}); msg != "" {
		t.Fatalf("rollback_window: %s", msg)
	}

	// Criterion 5. Making the deck public knocks first: the person
	// approves it in the console, and the call runs.
	if _, msg := useTool(t, lechat, "set_window_access", map[string]any{"window": "ana/why-keep-our-pages", "access": "public"}); !strings.Contains(msg, "/console/knocks/") {
		t.Fatalf("set_window_access public by an AI: %q, want a Knock", msg)
	}
	if status, _ := read("/ana/why-keep-our-pages/"); status == http.StatusOK {
		t.Error("the Window is public before the person approved")
	}
	var knocks struct {
		Knocks []struct {
			ID    int64  `json:"id"`
			State string `json:"state"`
		} `json:"knocks"`
	}
	if status := a.api(t, "list_knocks", map[string]any{}, &knocks); status != http.StatusOK || len(knocks.Knocks) != 1 || knocks.Knocks[0].State != "waiting" {
		t.Fatalf("list_knocks = %d %+v", status, knocks)
	}
	if status := a.api(t, "approve_knock", map[string]any{"knock": knocks.Knocks[0].ID}, nil); status != http.StatusOK {
		t.Fatalf("approve_knock = %d", status)
	}
	if status, body := read("/ana/why-keep-our-pages/"); status != http.StatusOK || !strings.Contains(body, "On our own server") {
		t.Errorf("the public Window = %d", status)
	}

	// Every call of the two AIs is in the Logbook, under its client and
	// its person, and each either changes nothing, can be undone, or was
	// approved.
	var lb cahut.LogbookPage
	if status := a.api(t, "list_logbook", cahut.ListLogbookInput{Limit: 1000}, &lb); status != http.StatusOK {
		t.Fatalf("list_logbook = %d", status)
	}
	ops := map[string]operation.Info{}
	for _, op := range st.cahut.Registry.Ops() {
		ops[op.Name] = op
	}
	clients := map[string]int{}
	for _, e := range lb.Entries {
		if e.Actor != "ai" {
			continue
		}
		if e.UserHandle != "ana" {
			t.Errorf("an AI call for %q, not for ana: %+v", e.UserHandle, e)
		}
		switch name, _, _ := strings.Cut(e.Client, " ("); name {
		case "ChatGPT", "Le Chat":
			clients[name]++
		default:
			t.Errorf("an AI call by %q", e.Client)
		}
		_, undone := undo[e.Operation]
		switch {
		case e.Outcome == logbook.Knocked, ops[e.Operation].ReadOnly:
		case e.Outcome == logbook.OK && (undone || e.KnockID != 0):
		default:
			t.Errorf("an AI call that cannot be undone and was not approved: %+v", e)
		}
	}
	if clients["ChatGPT"] != 1 || clients["Le Chat"] != 4 {
		t.Errorf("calls in the Logbook by client = %v", clients)
	}
}
