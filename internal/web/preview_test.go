package web

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/store"
)

func TestPreviews(t *testing.T) {
	f := setup(t)
	token := activate(t, f, f.ts, "camille", false)
	publishDeck(t, f, f.person(t, "localhost", "camille"))
	publishDeck(t, f, f.person(t, "localhost", "olivier"))

	// The console gets a frame for each Window of its person's Rooms.
	r := call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/preview_windows", token,
		`{"windows": ["camille/q3-review", "olivier/q3-review"]}`)
	previews, _ := r.body["previews"].([]any)
	if r.status != http.StatusOK || len(previews) != 1 {
		t.Fatalf("preview_windows: %d %s", r.status, r.raw)
	}
	frame := previews[0].(map[string]any)["frame"].(string)
	if !strings.HasPrefix(frame, "/-/w/") || !strings.HasSuffix(frame, "/") {
		t.Fatalf("frame = %q", frame)
	}

	// The frame shows the Window, without cookies, in its sandbox.
	resp, body := f.do(t, "localhost", http.MethodGet, frame, nil, nil)
	if resp.StatusCode != http.StatusOK || body != "<h1>Q3</h1>" {
		t.Fatalf("GET %s = %d %q", frame, resp.StatusCode, body)
	}
	if h := resp.Header; !strings.HasPrefix(h.Get("Content-Security-Policy"), "sandbox ") || h.Get("Referrer-Policy") != "no-referrer" {
		t.Errorf("the preview's file has the headers %v", h)
	}
	if resp, body := f.do(t, "localhost", http.MethodGet, frame+"css/deck.css", nil, nil); resp.StatusCode != http.StatusOK || !strings.Contains(body, "teal") {
		t.Errorf("GET %scss/deck.css = %d %q", frame, resp.StatusCode, body)
	}

	// A preview does not keep the session from going idle.
	db := f.open(t, "localhost").DB
	earlier := store.FormatTime(time.Now().Add(-10 * time.Minute))
	if _, err := db.Write.Exec("UPDATE sessions SET last_used_at = ? WHERE reader = 0", earlier); err != nil {
		t.Fatal(err)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, frame, nil, nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s = %d", frame, resp.StatusCode)
	}
	var used string
	if err := db.Read.QueryRow("SELECT last_used_at FROM sessions WHERE reader = 0").Scan(&used); err != nil || used != earlier {
		t.Errorf("after a preview, the session was last used at %q, want %q (%v)", used, earlier, err)
	}

	// Once the person signs out of the console, the frame shows nothing.
	if r := call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/sign_out", token, ``); r.status != http.StatusOK {
		t.Fatalf("sign_out: %d %s", r.status, r.raw)
	}
	if resp, body := f.do(t, "localhost", http.MethodGet, frame, nil, nil); resp.StatusCode != http.StatusNotFound || strings.Contains(body, "Q3") {
		t.Errorf("GET %s after sign_out = %d %q, want 404", frame, resp.StatusCode, body)
	}
}
