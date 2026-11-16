package web

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/skill"
)

// toolsWithoutSession lists the tools of the Cahut at host for the AI that
// holds token, in a request of the protocol of 2026-07-28, which opens no
// session.
func (f *fixture) toolsWithoutSession(t *testing.T, host, token string) map[string]*mcp.Tool {
	t.Helper()
	meta := `"_meta": {"io.modelcontextprotocol/protocolVersion": "2026-07-28", "io.modelcontextprotocol/clientInfo": {"name": "test", "version": "1"}, "io.modelcontextprotocol/clientCapabilities": {}}`
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, f.ts.URL+"/mcp",
		strings.NewReader(`{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {`+meta+`}}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Mcp-Protocol-Version", "2026-07-28")
	req.Header.Set("Mcp-Method", "tools/list")
	resp, err := f.client(host, token).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body []byte
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		sc := bufio.NewScanner(resp.Body)
		sc.Buffer(nil, 1<<20)
		for sc.Scan() {
			if data, ok := strings.CutPrefix(sc.Text(), "data: "); ok {
				body = []byte(data)
				break
			}
		}
	} else if err := json.NewDecoder(resp.Body).Decode((*json.RawMessage)(&body)); err != nil {
		t.Fatal(err)
	}
	var out struct {
		Result mcp.ListToolsResult `json:"result"`
	}
	if err := json.Unmarshal(body, &out); err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("tools/list without a session = %d %s: %v", resp.StatusCode, body, err)
	}
	tools := map[string]*mcp.Tool{}
	for _, tool := range out.Result.Tools {
		tools[tool.Name] = tool
	}
	return tools
}

// servers returns the keys of the MCP servers that the Cahut at domain
// keeps.
func (f *fixture) servers(t *testing.T, domain string) []mcpKey {
	t.Helper()
	f.web.mu.Lock()
	st := f.web.sites[domain]
	f.web.mu.Unlock()
	if st == nil {
		t.Fatalf("no site at %s", domain)
	}
	st.servers.mu.Lock()
	defer st.servers.mu.Unlock()
	var keys []mcpKey
	for k := range st.servers.m {
		keys = append(keys, k)
	}
	return keys
}

func writeSkill(t *testing.T, f *fixture, who operation.Actor, scope skill.Scope, name, description string) {
	t.Helper()
	text := "---\nname: " + name + "\ndescription: \"" + description + "\"\n---\n# " + name + "\n"
	f.call(t, "localhost", who, "write_skill", fmt.Sprintf(`{"scope": %q, "skill": %q, "text": %q}`, scope, name, text))
}

func TestTheCatalogFollowsTheSkills(t *testing.T) {
	f := setup(t)
	camille := f.person(t, "localhost", "camille")
	c := f.open(t, "localhost")
	if _, err := operation.Do(context.Background(), c.Registry, account.SetUserRole, admin, account.SetUserRoleInput{User: "camille", Role: account.Admin}); err != nil {
		t.Fatal(err)
	}
	description := func() string {
		return f.toolsWithoutSession(t, "localhost", f.token)["find_skills"].Description
	}
	pointed := func() bool {
		return strings.Contains(f.connect(t, "localhost", f.token).InitializeResult().Instructions, "skills of the user's team")
	}
	if pointed() {
		t.Error("the instructions of a Cahut without skills send AIs to find_skills before each task")
	}
	if d := description(); strings.Contains(d, "today:") {
		t.Errorf("find_skills without skills:\n%s", d)
	}
	// The client of the tests speaks the protocol of 2026-07-28, which has
	// no sessions: a session needs an earlier one.
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	before, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint: f.ts.URL + "/mcp", HTTPClient: f.client("localhost", f.token), DisableStandaloneSSE: true, MaxRetries: -1,
	}, &mcp.ClientSessionOptions{ProtocolVersion: "2025-11-25"})
	if err != nil {
		t.Fatal(err)
	}
	defer before.Close()
	if before.ID() == "" {
		t.Fatal("no session with the protocol of 2025-11-25")
	}
	tools, err := before.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	seen := false
	for _, tool := range tools.Tools {
		seen = seen || tool.Name == "find_skills" && !strings.Contains(tool.Description, "today:")
	}
	if !seen {
		t.Fatal("the session does not see find_skills without a catalog")
	}

	writeSkill(t, f, camille, skill.ScopeCahut, "relance-facture", "L'e-mail de relance d'une facture impayée.")
	if d := description(); !strings.HasSuffix(d, "\nThe skills of this Cahut today:\n- relance-facture: L'e-mail de relance d'une facture impayée.") {
		t.Errorf("find_skills after a write:\n%s", d)
	}
	if !pointed() {
		t.Error("the instructions of a Cahut with a skill do not send AIs to find_skills")
	}
	// A session opened before keeps the list it had.
	tools, err = before.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools.Tools {
		if tool.Name == "find_skills" && strings.Contains(tool.Description, "relance-facture") {
			t.Errorf("a session opened before the write sees the new catalog")
		}
	}

	f.call(t, "localhost", camille, "delete_skill", `{"skill": "relance-facture"}`)
	if d := description(); strings.Contains(d, "relance-facture") || strings.Contains(d, "today:") {
		t.Errorf("find_skills after a deletion:\n%s", d)
	}
	if pointed() {
		t.Error("the instructions still send AIs to find_skills once the last skill is deleted")
	}
}

func TestAPersonalSkillKeepsTheServer(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	f.toolsWithoutSession(t, "localhost", f.token)
	keys := f.servers(t, "localhost")
	writeSkill(t, f, olivier, skill.ScopePersonal, "notes", "My notes.")
	f.toolsWithoutSession(t, "localhost", f.token)
	if after := f.servers(t, "localhost"); len(after) != 1 || len(keys) != 1 || after[0] != keys[0] {
		t.Errorf("servers after a personal write = %v, before %v", after, keys)
	}
	if d := f.toolsWithoutSession(t, "localhost", f.token)["find_skills"].Description; strings.Contains(d, "notes") {
		t.Errorf("the catalog shows a personal skill:\n%s", d)
	}
}

func TestOldServersAreForgotten(t *testing.T) {
	f := setup(t)
	camille := f.person(t, "localhost", "camille")
	c := f.open(t, "localhost")
	if _, err := operation.Do(context.Background(), c.Registry, account.SetUserRole, admin, account.SetUserRoleInput{User: "camille", Role: account.Admin}); err != nil {
		t.Fatal(err)
	}
	tok, err := operation.Do(context.Background(), c.Registry, account.CreateToken, admin, account.CreateTokenInput{User: "camille", Name: "Mistral Vibe"})
	if err != nil {
		t.Fatal(err)
	}
	for i := range 10 {
		writeSkill(t, f, camille, skill.ScopeCahut, fmt.Sprintf("skill-%d", i), "A skill.")
		f.toolsWithoutSession(t, "localhost", f.token)
		f.toolsWithoutSession(t, "localhost", tok.Token)
	}
	revision, err := skill.Revision(context.Background(), c.DB.Read)
	if err != nil {
		t.Fatal(err)
	}
	keys := f.servers(t, "localhost")
	if len(keys) != 2 {
		t.Errorf("the Cahut keeps %d servers, want 2: %v", len(keys), keys)
	}
	for _, k := range keys {
		if k.revision != revision {
			t.Errorf("a server of revision %d stays, the catalog is at %d", k.revision, revision)
		}
	}
}

// TestToolListWithAFullCatalog checks that the list of tools of an
// administrator who may write stays under 40 KB with fifty skills of the
// Cahut, each with the longest description, and 700 bytes to spare for
// follows and from_skill (ADR 57, ADR 64).
func TestToolListWithAFullCatalog(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	c := f.open(t, "localhost")
	camille := f.person(t, "localhost", "camille")
	if _, err := operation.Do(ctx, c.Registry, account.SetUserRole, admin, account.SetUserRoleInput{User: "camille", Role: account.Admin}); err != nil {
		t.Fatal(err)
	}
	for i := range 50 {
		description := string([]rune(strings.Repeat(fmt.Sprintf("Le skill %02d dit «comment» <faire> & quoi. ", i), 25))[:1024])
		writeSkill(t, f, camille, skill.ScopeCahut, fmt.Sprintf("skill-%02d", i), strings.TrimSpace(description))
	}
	tok, err := operation.Do(ctx, c.Registry, account.CreateToken, admin, account.CreateTokenInput{User: "camille", Name: "Claude Code"})
	if err != nil {
		t.Fatal(err)
	}
	tools, err := f.connect(t, "localhost", tok.Token).ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var find string
	for _, tool := range tools.Tools {
		if tool.Name == "find_skills" {
			find = tool.Description
		}
	}
	if !strings.Contains(find, "\n- skill-00: ") || !strings.Contains(find, "more skills: find_skills lists them") {
		t.Errorf("find_skills does not carry a full catalog:\n%s", find)
	}
	list, _ := json.Marshal(tools)
	if len(list) >= 40<<10-700 {
		t.Errorf("the list of tools weighs %d bytes with a full catalog, %d or more", len(list), 40<<10-700)
	}
	t.Logf("the list of tools weighs %d bytes with a full catalog", len(list))
}

// The console imports a skill of 4 MiB, which takes 6 MiB in base64, and
// every other call stays under 1 MiB. See
// docs/adr/0064-the-skills-of-a-cahut.md.
func TestImportSkillTakesALargeCall(t *testing.T) {
	f := setup(t)
	console := f.ts
	ada := activate(t, f, console, "ada", true)

	// 3 MiB of bytes that do not compress, in two files.
	files := []skill.File{{Path: "assets/a.bin", Content: make([]byte, 3<<19)}, {Path: "assets/b.bin", Content: make([]byte, 3<<19)}}
	for _, f := range files {
		_, _ = rand.Read(f.Content) // crypto/rand never fails
	}
	text := "---\nname: charte-acme\ndescription: The brand of Acme.\n---\n# Acme\n"
	zip, err := skill.WriteZip("charte-acme", text, files)
	if err != nil {
		t.Fatal(err)
	}
	body := `{"scope": "cahut", "zip": "` + base64.StdEncoding.EncodeToString(zip) + `"}`
	if len(body) < 4<<20 {
		t.Fatalf("the call holds %d bytes, want more than 4 MiB", len(body))
	}
	r := call(t, f, console, http.MethodPost, "localhost", "/console/api/import_skill", ada, body)
	if r.status != http.StatusOK || r.body["skill"] != "charte-acme" {
		t.Fatalf("import_skill of %d bytes: %d %.200s", len(body), r.status, r.raw)
	}

	for _, c := range []struct {
		path, body string
	}{
		{"/console/api/write_skill", `{"scope": "cahut", "skill": "x", "text": "` + strings.Repeat("x", 2<<20) + `"}`},
		{"/console/api/import_skill", `{"scope": "cahut", "zip": "` + strings.Repeat("A", 6<<20) + `"}`},
	} {
		if r := call(t, f, console, http.MethodPost, "localhost", c.path, ada, c.body); r.status != http.StatusRequestEntityTooLarge {
			t.Errorf("%s with %d bytes: %d, want 413", c.path, len(c.body), r.status)
		}
	}
}
