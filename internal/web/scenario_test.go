package web

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
)

// TestBatch2Scenario plays the demonstration of batch 2 of the V0 plan
// (docs/plan-lot-2.md) through the console's API and MCP: two Cahuts, an
// administrator of the server, people who sign in, a team Room, AIs that
// publish, and the Logbook.
func TestBatch2Scenario(t *testing.T) {
	f := setup(t)
	console := f.ts
	ctx := context.Background()
	must := func(step string, r apiResponse, status int) map[string]any {
		t.Helper()
		if r.status != status {
			t.Fatalf("%s: %d %s", step, r.status, r.raw)
		}
		return r.body
	}
	post := func(host, path, token string, body any) apiResponse {
		t.Helper()
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		return call(t, f, console, http.MethodPost, host, path, token, string(b))
	}
	code := func(m map[string]any, keys ...string) string {
		for _, k := range keys[:len(keys)-1] {
			m = m[k].(map[string]any)
		}
		return m[keys[len(keys)-1]].(string)
	}

	// 1. The administrator of the server activates their account at the
	// console of the first Cahut.
	a, err := operation.Do(ctx, f.srv.Registry, server.CreateAdmin, admin, server.CreateAdminInput{Handle: "root", Name: "Root"})
	if err != nil {
		t.Fatal(err)
	}
	root := code(must("1. activate the administrator of the server",
		post("localhost", "/server/api/activate_account", "", map[string]string{"code": a.Activation.Code, "password": password}), 200), "token")

	// 2. They declare a third Cahut, with its first administrator, who
	// activates their account there.
	n := must("2. declare a Cahut", post("localhost", "/server/api/add_domain", root,
		map[string]any{"domain": "gamma.localhost", "admin": map[string]string{"handle": "zoe", "name": "Zoé"}}), 200)
	zoe := code(must("2. activate its first administrator", post("gamma.localhost", "/console/api/activate_account", "",
		map[string]string{"code": code(n, "admin", "activation", "code"), "password": password}), 200), "token")

	// 3. In the first Cahut, its administrator, whom the command line made,
	// creates two people, who activate their accounts. The administrator of
	// the server manages no person there. See
	// docs/adr/0058-a-console-for-the-server.md.
	must("3. create alice as the administrator of the server", post("localhost", "/server/api/cahuts/localhost/create_user", root,
		map[string]string{"handle": "alice", "name": "alice"}), 404)
	carla := activate(t, f, f.ts, "carla", true)
	sessions := map[string]string{}
	for _, handle := range []string{"alice", "bruno"} {
		u := must("3. create "+handle, post("localhost", "/console/api/create_user", carla, map[string]string{"handle": handle, "name": handle}), 200)
		sessions[handle] = code(must("3. activate "+handle, post("localhost", "/console/api/activate_account", "",
			map[string]string{"code": code(u, "activation", "code"), "password": password}), 200), "token")
	}

	// 4. Alice creates the team Room marketing and adds Bruno.
	must("4. create marketing", post("localhost", "/console/api/create_room", sessions["alice"], map[string]string{"title": "Marketing"}), 200)
	must("4. add bruno", post("localhost", "/console/api/add_room_member", sessions["alice"], map[string]string{"room": "marketing", "user": "bruno"}), 200)

	// 5. Each creates a token and connects an AI, which publishes.
	ais := map[string]*mcp.ClientSession{}
	for _, handle := range []string{"alice", "bruno"} {
		tok := code(must("5. create a token for "+handle, post("localhost", "/console/api/create_token", sessions[handle], map[string]string{"name": "Claude Code"}), 200), "token")
		ais[handle] = f.connect(t, "localhost", tok)
	}
	publish := func(handle, title, room string) *mcp.CallToolResult {
		t.Helper()
		res, err := ais[handle].CallTool(ctx, &mcp.CallToolParams{Name: "publish_window", Arguments: map[string]any{
			"title": title, "kind": "document", "room": room, "declaration": map[string]any{"scripts": false}, "files": []map[string]string{{"path": "index.html", "content": "<h1>" + title + "</h1>"}},
		}})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	for _, p := range [][3]string{{"alice", "Q3 review", "marketing"}, {"alice", "Notes", ""}, {"bruno", "Draft", ""}} {
		if res := publish(p[0], p[1], p[2]); res.IsError {
			t.Fatalf("5. the AI of %s publishes %s: %v", p[0], p[1], res.Content)
		}
	}

	// 6. Bruno sees the Window of marketing, not the personal Room of
	// Alice; the third Cahut sees nothing of the first.
	ws := must("6. bruno's Windows", post("localhost", "/console/api/list_windows", sessions["bruno"], map[string]string{}), 200)["windows"].([]any)
	var names []string
	for _, w := range ws {
		m := w.(map[string]any)
		names = append(names, m["room"].(string)+"/"+m["name"].(string))
	}
	slices.Sort(names)
	if want := []string{"bruno/draft", "marketing/q3-review"}; !slices.Equal(names, want) {
		t.Errorf("6. bruno sees %q, want %q", names, want)
	}
	res, err := ais["bruno"].CallTool(ctx, &mcp.CallToolParams{Name: "get_window", Arguments: map[string]any{"window": "alice/notes"}})
	if err != nil || !res.IsError {
		t.Errorf("6. the AI of bruno opens the personal Room of alice: %v %v", res, err)
	}
	if ws := must("6. zoe's Windows", post("gamma.localhost", "/console/api/list_windows", zoe, map[string]string{}), 200)["windows"].([]any); len(ws) != 0 {
		t.Errorf("6. zoe sees %v", ws)
	}
	must("6. zoe's session at localhost", post("localhost", "/console/api/list_windows", zoe, map[string]string{}), 401)

	// 7. The Logbook shows the publications, the accounts and the Room, with
	// the AI and the person.
	entries := must("7. the Logbook", post("localhost", "/console/api/list_logbook", carla, map[string]int{"limit": 200}), 200)["entries"].([]any)
	seen := map[string]int{}
	for _, e := range entries {
		m := e.(map[string]any)
		if m["outcome"] != "ok" {
			continue
		}
		op := m["operation"].(string)
		if op == "publish_window" && !strings.HasPrefix(m["client"].(string), "Claude Code") {
			t.Errorf("7. a publication names no AI: %v", m)
		}
		seen[op]++
	}
	// setup created olivier in the first Cahut already, and the command line
	// carla.
	for op, want := range map[string]int{"publish_window": 3, "create_user": 4, "activate_account": 3, "create_room": 1, "add_room_member": 1} {
		if seen[op] != want {
			t.Errorf("7. the Logbook holds %d %s, want %d", seen[op], op, want)
		}
	}

	// 8. The AI of Bruno tries to add someone to marketing, which only its
	// owners do: the call is checked before it could knock, and refused.
	res, err = ais["bruno"].CallTool(ctx, &mcp.CallToolParams{Name: "add_room_member", Arguments: map[string]any{"room": "marketing", "user": "alice"}})
	if err != nil || !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "only the owners of the Room marketing") {
		t.Errorf("8. the AI of bruno changes the people of marketing: %v %v", res, err)
	}
}
