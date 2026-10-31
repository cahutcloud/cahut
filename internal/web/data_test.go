package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
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
