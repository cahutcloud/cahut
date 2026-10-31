package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// petitionWindow is a Window that keeps data: anyone signs, people see
// their own signature, and members count them.
const petitionWindow = `{"title": "Save the river", "kind": "site", "files": [{"path": "index.html", "content": "<h1>Sign</h1>"}],
	"declaration": {"scripts": true, "data": {
		"migrations": ["CREATE TABLE signatures (id INTEGER PRIMARY KEY, name TEXT NOT NULL, email TEXT, reader TEXT)"],
		"readable": {"signatures": ["name"]},
		"queries": {
			"sign": {"sql": "INSERT INTO signatures (name, email, reader) VALUES (:name, :email, :reader)", "who": "anyone"},
			"mine": {"sql": "SELECT name, email, reader FROM signatures WHERE reader = :reader", "who": "people"},
			"count": {"sql": "SELECT count(*) AS n FROM signatures", "who": "members"}
		}%s
	}}}`

// post sends body to path at localhost, as the page of a Window does, and
// returns the status and the JSON of the response.
func (f *fixture) post(t *testing.T, path, body string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, f.ts.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "null")
	resp, err := f.client("localhost", "").Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("POST %s: Access-Control-Allow-Origin = %q", path, got)
	}
	b, _ := io.ReadAll(resp.Body)
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("POST %s = %d %s: %v", path, resp.StatusCode, b, err)
	}
	return resp.StatusCode, out
}

func errorCode(out map[string]any) string {
	e, _ := out["error"].(map[string]any)
	code, _ := e["code"].(string)
	return code
}

func TestPagesCallTheirData(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	f.call(t, "localhost", olivier, "publish_window", strings.Replace(petitionWindow, "%s", "", 1))
	base := "/olivier/save-the-river/.data/"

	// A Window of its Room keeps its data to the frames of its readers.
	if status, out := f.post(t, base+"sql", `{"sql": "SELECT name FROM signatures"}`); status != http.StatusNotFound || errorCode(out) != "not_found" {
		t.Errorf("the data of a Window of its Room, at its address: %d %v", status, out)
	}

	f.setAccess(t, "localhost", "save-the-river", "public")
	status, out := f.post(t, base+"queries/sign", `{"params": {"name": "Ana", "email": "ana@example.com"}}`)
	if status != http.StatusOK || out["changes"] != float64(1) || out["id"] != float64(1) {
		t.Fatalf("sign = %d %v", status, out)
	}
	status, out = f.post(t, base+"sql", `{"sql": "SELECT name, email, reader FROM signatures"}`)
	if got, _ := json.Marshal(out["rows"]); status != http.StatusOK || string(got) != `[["Ana",null,null]]` {
		t.Errorf("rows = %d %s", status, got)
	}
	for _, tc := range []struct {
		path, body string
		status     int
		code       string
	}{
		{base + "queries/sign", `{"params": {"email": "x@example.com"}}`, http.StatusConflict, "constraint"},
		{base + "queries/sign", `{"params": {"name": "Ben", "reader": "someone"}}`, http.StatusBadRequest, "invalid"},
		{base + "queries/mine", `{}`, http.StatusForbidden, "forbidden"},
		{base + "queries/count", `{}`, http.StatusForbidden, "forbidden"},
		{base + "queries/nothing", `{}`, http.StatusNotFound, "not_found"},
		{base + "sql", `{"sql": "DELETE FROM signatures"}`, http.StatusBadRequest, "invalid"},
		{base + "sql", `{"sql": "WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n) SELECT count(*) FROM n"}`, http.StatusBadRequest, "timeout"},
		{base + "sql", `not json`, http.StatusBadRequest, "invalid"},
		{base + "sql", `{"sql": "` + strings.Repeat("x", 70<<10) + `"}`, http.StatusRequestEntityTooLarge, "invalid"},
		{base + "other", `{}`, http.StatusNotFound, "not_found"},
		{"/olivier/nothing/.data/sql", `{"sql": "SELECT 1"}`, http.StatusNotFound, "not_found"},
	} {
		if status, out := f.post(t, tc.path, tc.body); status != tc.status || errorCode(out) != tc.code {
			t.Errorf("POST %s %.40s = %d %v, want %d %s", tc.path, tc.body, status, out, tc.status, tc.code)
		}
	}

	// The preflight of a sandbox.
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodOptions, f.ts.URL+base+"sql", nil)
	req.Header.Set("Origin", "null")
	req.Header.Set("Access-Control-Request-Method", "POST")
	resp, err := f.client("localhost", "").Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent || resp.Header.Get("Access-Control-Allow-Origin") != "*" ||
		resp.Header.Get("Access-Control-Allow-Headers") != "Content-Type" {
		t.Errorf("preflight = %d %v", resp.StatusCode, resp.Header)
	}

	// Writes from one address are limited, reads go on.
	var limited bool
	for range DataWriteBurst + 1 {
		if status, out := f.post(t, base+"queries/sign", `{"params": {"name": "Robot"}}`); status == http.StatusTooManyRequests && errorCode(out) == "too_many" {
			limited = true
		}
	}
	if !limited {
		t.Error("writes from one address are not limited")
	}
	if status, _ := f.post(t, base+"sql", `{"sql": "SELECT count(*) FROM signatures"}`); status != http.StatusOK {
		t.Errorf("a read after the writes were limited: %d", status)
	}
}

func TestFramesCallTheirData(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	f.call(t, "localhost", olivier, "publish_window", strings.Replace(petitionWindow, "%s", "", 1))
	frame := f.frame(t, "/olivier/save-the-river/", f.signIn(t, "localhost", "olivier"))
	base := frame + ".data/"
	if status, out := f.post(t, base+"queries/sign", `{"params": {"name": "Olivier"}}`); status != http.StatusOK {
		t.Fatalf("sign in a frame = %d %v", status, out)
	}
	status, out := f.post(t, base+"queries/mine", `{}`)
	rows, _ := out["rows"].([]any)
	if status != http.StatusOK || len(rows) != 1 {
		t.Fatalf("mine = %d %v", status, out)
	}
	reader, _ := rows[0].([]any)[2].(string)
	if !strings.HasPrefix(reader, "r") || len(reader) != 25 {
		t.Errorf(":reader = %q", reader)
	}
	if status, out := f.post(t, base+"queries/count", `{}`); status != http.StatusOK {
		t.Errorf("count by a member = %d %v", status, out)
	}
	// A frame that does not verify opens nothing.
	if status, out := f.post(t, "/-/frame/forged/.data/sql", `{"sql": "SELECT 1"}`); status != http.StatusNotFound || errorCode(out) != "not_found" {
		t.Errorf("a forged frame = %d %v", status, out)
	}
}

// TestModeratorFramesReadWhileTheyOpen checks that the pages of a reported
// Window, in the frame of a moderator, read its data only while the frame
// opens the Window: while their session lasts, they moderate the Cahut and
// the report waits, as its files are served. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.
func TestModeratorFramesReadWhileTheyOpen(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	f.call(t, "localhost", olivier, "publish_window", strings.Replace(petitionWindow, "%s", "", 1))
	if status, _ := f.sendReport(t, "https://localhost/olivier/save-the-river/", "illegal", "It shows what the law forbids.", ""); status != http.StatusOK {
		t.Fatalf("report = %d", status)
	}
	read := `{"sql": "SELECT count(*) FROM signatures"}`
	frameOf := func(who, path, token string) string {
		t.Helper()
		r := call(t, f, f.ts, http.MethodPost, "localhost", path, token, `{"report": 1}`)
		frame, _ := r.body["frame"].(string)
		if r.status != http.StatusOK || frame == "" {
			t.Fatalf("preview_reported_window as %s: %d %s", who, r.status, r.raw)
		}
		if status, out := f.post(t, frame+".data/sql", read); status != http.StatusOK {
			t.Fatalf("the data in the frame of %s = %d %v", who, status, out)
		}
		return frame
	}
	closed := func(who, frame string) {
		t.Helper()
		if status, out := f.post(t, frame+".data/sql", read); status != http.StatusNotFound || errorCode(out) != "not_found" {
			t.Errorf("the data in the frame of %s = %d %v, want 404 not_found", who, status, out)
		}
	}

	root := serverAdmin(t, f, "root")
	ana := activate(t, f, f.ts, "ana", true)
	bob := activate(t, f, f.ts, "bob", true)
	rootFrame := frameOf("root", "/server/api/cahuts/localhost/preview_reported_window", root)
	anaFrame := frameOf("ana", "/console/api/preview_reported_window", ana)
	bobFrame := frameOf("bob", "/console/api/preview_reported_window", bob)

	if r := call(t, f, f.ts, http.MethodPost, "localhost", "/server/api/sign_out", root, `{}`); r.status != http.StatusOK {
		t.Fatalf("sign_out of the server: %d %s", r.status, r.raw)
	}
	closed("root once signed out", rootFrame)

	if r := call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/sign_out", ana, `{}`); r.status != http.StatusOK {
		t.Fatalf("sign_out of ana: %d %s", r.status, r.raw)
	}
	closed("ana once signed out", anaFrame)

	// The pages of a frame call at once, more than the reads of the Cahut
	// take together: the check holds no connection of its own.
	statuses := make(chan int, 2*readConnections)
	var wg sync.WaitGroup
	for range 2 * readConnections {
		wg.Go(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, http.MethodPost, f.ts.URL+bobFrame+".data/sql", strings.NewReader(read))
			req.Header.Set("Content-Type", "application/json")
			resp, err := f.client("localhost", "").Do(req)
			if err != nil {
				statuses <- 0
				return
			}
			resp.Body.Close()
			statuses <- resp.StatusCode
		})
	}
	wg.Wait()
	close(statuses)
	for status := range statuses {
		if status != http.StatusOK {
			t.Errorf("a read among many at once = %d", status)
		}
	}

	f.call(t, "localhost", admin, "dismiss_report", `{"report": 1}`)
	closed("bob once the report is decided", bobFrame)
}

// readConnections is how many connections read the database of a Cahut.
const readConnections = 4

func TestWritesClose(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	f.call(t, "localhost", olivier, "publish_window", strings.Replace(petitionWindow, "%s", `, "closes": "2020-01-01T00:00:00Z"`, 1))
	f.setAccess(t, "localhost", "save-the-river", "public")
	if status, out := f.post(t, "/olivier/save-the-river/.data/queries/sign", `{"params": {"name": "Late"}}`); status != http.StatusForbidden || errorCode(out) != "closed" {
		t.Errorf("a write after closes = %d %v", status, out)
	}
	if status, _ := f.post(t, "/olivier/save-the-river/.data/sql", `{"sql": "SELECT count(*) FROM signatures"}`); status != http.StatusOK {
		t.Errorf("a read after closes = %d", status)
	}
}

// TestPublicKnockSaysWhatAnyoneWrites checks that the Knock that makes a
// Window public names the queries that write for anyone.
func TestPublicKnockSaysWhatAnyoneWrites(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	session := f.connect(t, "localhost", f.token)
	var args map[string]any
	if err := json.Unmarshal([]byte(strings.Replace(petitionWindow, "%s", "", 1)), &args); err != nil {
		t.Fatal(err)
	}
	if res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "publish_window", Arguments: args}); err != nil || res.IsError {
		t.Fatalf("publish_window = %+v, %v", res, err)
	}
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "set_window_access", Arguments: map[string]any{"window": "olivier/save-the-river", "access": "public"}})
	if err != nil || !res.IsError {
		t.Fatalf("set_window_access public = %+v, %v, want a Knock", res, err)
	}
	var detail string
	if err := f.open(t, "localhost").DB.Read.QueryRow("SELECT detail FROM knocks WHERE operation = 'set_window_access'").Scan(&detail); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(detail, "anyone writes with sign (adds)") || strings.Contains(detail, "count") {
		t.Errorf("the detail of the Knock = %q", detail)
	}
}
