package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/update"
)

// A suspended Cahut keeps everything, and serves nothing of what it holds
// but to the administrators of its server: readers meet its page, people,
// AIs and services a refusal. Once resumed, everything works as before,
// with the same sessions and connections. See
// docs/adr/0048-suspending-a-cahut.md.
func TestSuspendedCahut(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	olivier := f.person(t, "localhost", "olivier")
	publishDeck(t, f, olivier)
	f.setAccess(t, "localhost", "q3-review", "public")
	camille := activate(t, f, f.ts, "camille", false)
	svc := serviceToken(t, f, "localhost")
	adminAPI := httptest.NewServer(f.web.AdminHandler())
	t.Cleanup(adminAPI.Close)
	ai := f.connect(t, "localhost", f.token)
	a, err := operation.Do(ctx, f.srv.Registry, server.CreateAdmin, admin, server.CreateAdminInput{Handle: "root", Name: "Root"})
	if err != nil {
		t.Fatal(err)
	}
	r := call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/server/activate_account", "",
		`{"code": "`+a.Activation.Code+`", "password": "`+password+`"}`)
	if r.status != http.StatusOK {
		t.Fatalf("activate_account of the server: %d %s", r.status, r.raw)
	}
	root := r.body["token"].(string)

	// The administrators of the server suspend it from the console of the
	// Cahut, with their session.
	r = call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/suspend_cahut", camille, `{"reason": "Unpaid invoice"}`)
	if r.status != http.StatusForbidden {
		t.Errorf("suspend_cahut by a person: %d %s", r.status, r.raw)
	}
	r = call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/suspend_cahut", root, `{"reason": "Unpaid invoice"}`)
	if r.status != http.StatusOK || r.body["suspension"].(map[string]any)["reason"] != "Unpaid invoice" {
		t.Fatalf("suspend_cahut: %d %s", r.status, r.raw)
	}

	// Readers meet the page of a suspended Cahut, which names nothing that
	// it holds, wherever they look.
	for _, path := range []string{"/olivier/q3-review/", "/olivier/q3-review/css/deck.css", "/olivier/", SkillPath, "/-/sign-in"} {
		resp, body := f.get(t, "localhost", http.MethodGet, path)
		if resp.StatusCode != http.StatusServiceUnavailable || !strings.Contains(body, "This Cahut is suspended") ||
			strings.Contains(body, "Q3") || strings.Contains(body, "Unpaid") || strings.Contains(body, "Sign in") {
			t.Errorf("GET %s = %d %q", path, resp.StatusCode, body)
		}
		if csp := resp.Header.Get("Content-Security-Policy"); csp != pagePolicy("https://localhost/") {
			t.Errorf("GET %s has the policy %q", path, csp)
		}
	}
	if resp, _ := f.get(t, "localhost", http.MethodPost, "/-/key"); resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("POST /-/key = %d", resp.StatusCode)
	}
	// What Cahut serves of its own still answers: its health, its brand,
	// its icon, its console and the endpoints of OAuth.
	for _, path := range []string{"/-/healthz", "/-/pages.css", "/-/brand/tokens/cahut-tokens.css", "/favicon.ico",
		"/.well-known/oauth-authorization-server"} {
		if resp, body := f.get(t, "localhost", http.MethodGet, path); resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s = %d %q", path, resp.StatusCode, body)
		}
	}
	// Without the console build tag, the console answers with a notice.
	if resp, body := f.get(t, "localhost", http.MethodGet, "/console/"); strings.Contains(body, "This Cahut is suspended") {
		t.Errorf("GET /console/ = %d %q", resp.StatusCode, body)
	}
	// The other Cahuts of the server serve as before.
	if resp, body := f.get(t, "beta.localhost", http.MethodGet, "/nobody/"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET /nobody/ of another Cahut = %d %q", resp.StatusCode, body)
	}

	// People, services and AIs are refused, with the code of the console.
	refused := func(who string, r apiResponse) {
		t.Helper()
		if r.status != http.StatusForbidden || r.body["code"] != "cahut_suspended" {
			t.Errorf("%s: %d %s", who, r.status, r.raw)
		}
	}
	refused("get_cahut of a person", call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/get_cahut", camille, `{}`))
	refused("sign_in", call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/sign_in", "",
		`{"handle": "camille", "password": "`+password+`"}`))
	refused("get_cahut of a service", call(t, f, adminAPI, http.MethodPost, "localhost", "/api/get_cahut", svc, `{}`))
	res, err := ai.CallTool(ctx, &mcp.CallToolParams{Name: "get_cahut", Arguments: map[string]any{}})
	if err != nil || !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "suspended by the administrators of its server") {
		t.Errorf("get_cahut of an AI = %+v, %v", res, err)
	}
	// The page of sign-in learns that the Cahut is suspended.
	r = call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/describe_sign_in", "", `{}`)
	if r.status != http.StatusOK || r.body["suspended"] != true {
		t.Errorf("describe_sign_in: %d %s", r.status, r.raw)
	}

	// The administrators of the server act in it, and see it suspended in
	// the list of the Cahuts.
	r = call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/list_users", root, `{}`)
	if r.status != http.StatusOK {
		t.Errorf("list_users as an administrator of the server: %d %s", r.status, r.raw)
	}
	r = call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/server/list_domains", root, `{}`)
	cahuts, _ := r.body["cahuts"].([]any)
	if r.status != http.StatusOK || len(cahuts) != 2 || cahuts[0].(map[string]any)["suspension"] == nil ||
		cahuts[1].(map[string]any)["suspension"] != nil {
		t.Errorf("list_domains: %d %s", r.status, r.raw)
	}
	// The health of the server opens it, and says it is suspended.
	health := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1/-/healthz", nil)
	health.RemoteAddr = "127.0.0.1:50000"
	w := httptest.NewRecorder()
	f.web.localHealth(nil).ServeHTTP(w, health)
	var h update.Health
	if err := json.Unmarshal(w.Body.Bytes(), &h); err != nil || !h.OK || !h.Cahuts[0].Suspended || h.Cahuts[1].Suspended {
		t.Errorf("health = %d %s", w.Code, w.Body)
	}

	// Once resumed, everything works as before: the same session of the
	// person, the same AI, the same service token, the same Window.
	if _, err := operation.Do(ctx, f.open(t, "localhost").Registry, settings.Resume, admin, settings.ResumeCahutInput{}); err != nil {
		t.Fatalf("resume_cahut: %v", err)
	}
	if resp, body := f.get(t, "localhost", http.MethodGet, "/olivier/q3-review/"); resp.StatusCode != http.StatusOK || !strings.Contains(body, "Q3") {
		t.Errorf("GET the Window once resumed = %d %q", resp.StatusCode, body)
	}
	if r := call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/get_cahut", camille, `{}`); r.status != http.StatusOK {
		t.Errorf("get_cahut of the person once resumed: %d %s", r.status, r.raw)
	}
	if r := call(t, f, adminAPI, http.MethodPost, "localhost", "/api/get_cahut", svc, `{}`); r.status != http.StatusOK {
		t.Errorf("get_cahut of the service once resumed: %d %s", r.status, r.raw)
	}
	if res, err := ai.CallTool(ctx, &mcp.CallToolParams{Name: "get_cahut", Arguments: map[string]any{}}); err != nil || res.IsError {
		t.Errorf("get_cahut of the AI once resumed = %+v, %v", res, err)
	}
}
