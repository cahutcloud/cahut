package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/window"
)

// The tests of this file come from the security review of batch 8. See
// docs/revue-securite-v0.md.

// keyPass types key for the Window id and returns the pass cookie.
func (f *fixture) keyPass(t *testing.T, id int64, key string) *http.Cookie {
	t.Helper()
	form := url.Values{"window": {"olivier/q3-review"}, "key": {key}}
	resp, _ := f.do(t, "localhost", http.MethodPost, "/-/key", nil, form)
	for _, c := range resp.Cookies() {
		if c.Name == "__Host-cahut-key-"+strconv.FormatInt(id, 10) && c.MaxAge > 0 {
			return &http.Cookie{Name: c.Name, Value: c.Value} //nolint:gosec // a cookie that the test sends back
		}
	}
	t.Fatalf("no pass: %d %v", resp.StatusCode, resp.Header)
	return nil
}

// TestNewKeyEndsOldFrames checks that a frame opened with an old Key stays
// closed once the next reader gets a pass: passes never reuse an id.
func TestNewKeyEndsOldFrames(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	w := publishDeck(t, f, olivier)
	id := f.windowID(t, w.Name)
	const path = "/olivier/q3-review/"
	oldKey := f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key"}`).(window.AccessSet).Key

	// Mallory types the old Key and keeps the address of her frame.
	mallory := f.frame(t, path, f.keyPass(t, id, oldKey))

	// The owner draws a new Key to shut her out: her frame closes.
	newKey := f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key", "new_key": true}`).(window.AccessSet).Key
	if resp, _ := f.do(t, "localhost", http.MethodGet, mallory, nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("old frame after a new Key = %d, want 404", resp.StatusCode)
	}

	// Alice, given the new Key, opens the Window, and Mallory's frame stays
	// closed.
	f.frame(t, path, f.keyPass(t, id, newKey))
	if resp, _ := f.do(t, "localhost", http.MethodGet, mallory, nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("old frame after the next pass = %d, want 404", resp.StatusCode)
	}
}

// TestEndedSessionsKeepTheirFramesClosed checks that the frame of a reader
// who lost access stays closed when someone else signs in: sessions never
// reuse an id.
func TestEndedSessionsKeepTheirFramesClosed(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	publishDeck(t, f, olivier)
	f.person(t, "localhost", "camille")
	const path = "/olivier/q3-review/"
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "internal"}`)

	// Camille, not a member of the Room, opens the internal Window.
	src := f.frame(t, path, f.signIn(t, "localhost", "camille"))

	// The Window goes back to its Room only, and Camille is disabled: her
	// frame closes.
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "room"}`)
	if _, err := operation.Do(context.Background(), f.open(t, "localhost").Registry, account.DisableUser, admin, account.UserInput{User: "camille"}); err != nil {
		t.Fatal(err)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, src, nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("frame of a disabled reader = %d, want 404", resp.StatusCode)
	}

	// Olivier, a member, signs in to read: his session takes a new id.
	f.signIn(t, "localhost", "olivier")
	if resp, body := f.do(t, "localhost", http.MethodGet, src, nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("frame of a disabled reader after the next sign-in = %d %q, want 404", resp.StatusCode, body)
	}
}

// TestRightKeyOpensAWindowUnderAttack checks that wrong Keys from many
// addresses do not shut out a reader who comes with the right Key, while
// each address that guessed wrong waits.
func TestRightKeyOpensAWindowUnderAttack(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	publishDeck(t, f, olivier)
	key := f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key"}`).(window.AccessSet).Key
	post := func(addr, k string) int {
		form := url.Values{"window": {"olivier/q3-review"}, "key": {k}}
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "http://localhost/-/key", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = addr
		rec := httptest.NewRecorder()
		f.web.ServeHTTP(rec, req)
		return rec.Code
	}

	// One IPv6 /64 counts as one address.
	for i := range KeyBurst + 5 {
		post("[2001:db8::"+strconv.Itoa(i+1)+"]:1234", "wrong wrong")
	}
	if code := post("[2001:db8::99]:1234", key); code != http.StatusTooManyRequests {
		t.Errorf("the right Key from a /64 that spent its tries = %d, want 429", code)
	}

	// Many /64s spend the Window's budget of failures.
	for i := range KeyWindowBurst + 1 {
		post("[2001:db8:"+strconv.FormatInt(int64(i+1), 16)+"::1]:1234", "wrong wrong")
	}
	if code := post("192.0.2.7:1234", key); code != http.StatusSeeOther {
		t.Errorf("the right Key from a fresh address, the Window under attack = %d, want 303", code)
	}
	if code := post("[2001:db8:1::1]:1234", key); code != http.StatusTooManyRequests {
		t.Errorf("a Key from an address that guessed wrong, the Window under attack = %d, want 429", code)
	}
}

// TestSignOutEndsReaderSessions checks that signing out of the console
// signs out of the Windows too, in that browser.
func TestSignOutEndsReaderSessions(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	c := f.open(t, "localhost")
	olivier := f.person(t, "localhost", "olivier")
	publishDeck(t, f, olivier)

	link, err := operation.Do(ctx, c.Registry, account.CreateActivationLink, admin, account.UserInput{User: "olivier"})
	if err != nil {
		t.Fatal(err)
	}
	visitor := operation.Actor{Kind: operation.Visitor, Via: "api", Domain: "localhost"}
	s, err := operation.Do(ctx, c.Registry, account.ActivateAccount, visitor, account.ActivateInput{Code: link.Code, Password: "correct horse battery staple"})
	if err != nil {
		t.Fatal(err)
	}
	owner, err := account.VerifySession(ctx, c.DB, s.Token, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	olivier.Session = owner.SessionID

	code := f.call(t, "localhost", olivier, "sign_in_to_windows", `{}`).(account.ReaderCode).Code
	resp, _ := f.do(t, "localhost", http.MethodGet, "/-/session?code="+code+"&to=%2Folivier%2F", nil, nil)
	var reader *http.Cookie
	for _, ck := range resp.Cookies() {
		if ck.Name == "__Host-cahut-reader" {
			reader = &http.Cookie{Name: ck.Name, Value: ck.Value} //nolint:gosec // a cookie that the test sends back
		}
	}
	if reader == nil {
		t.Fatalf("no reader cookie: %d %v", resp.StatusCode, resp.Header)
	}
	f.frame(t, "/olivier/q3-review/", reader)

	f.call(t, "localhost", olivier, "sign_out", `{}`)
	if _, err := account.VerifyReader(ctx, c.DB, reader.Value, time.Now()); err == nil {
		t.Error("the reader session opened by a console session lasts after it signed out")
	}
}

func TestClientKey(t *testing.T) {
	for addr, want := range map[string]string{
		"192.0.2.7:443":             "192.0.2.7",
		"[2001:db8:1:2:3::4]:443":   "2001:db8:1:2::/64",
		"[2001:db8:1:2:ffff::]:443": "2001:db8:1:2::/64",
		"[::ffff:192.0.2.7]:443":    "192.0.2.7",
	} {
		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost/", nil)
		r.RemoteAddr = addr
		if got := clientKey(r); got != want {
			t.Errorf("clientKey(%s) = %s, want %s", addr, got, want)
		}
	}
}

// TestMCPSessionsPerToken checks that a token opens a burst of MCP
// sessions, then waits, while the sessions it has keep working.
func TestMCPSessionsPerToken(t *testing.T) {
	f := setup(t)
	client := f.client("localhost", f.token)
	initialize := func() (int, string) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, f.ts.URL+"/mcp", strings.NewReader(
			`{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2025-06-18", "capabilities": {}, "clientInfo": {"name": "test", "version": "1"}}}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		return resp.StatusCode, resp.Header.Get("Mcp-Session-Id")
	}
	var session string
	for range MCPSessionBurst {
		status, id := initialize()
		if status != http.StatusOK {
			t.Fatalf("initialize = %d", status)
		}
		session = id
	}
	if status, _ := initialize(); status != http.StatusTooManyRequests {
		t.Errorf("initialize past the burst = %d, want 429", status)
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, f.ts.URL+"/mcp", strings.NewReader(
		`{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Mcp-Session-Id", session)
	req.Header.Set("Mcp-Protocol-Version", "2025-06-18")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("tools/list in an open session = %d, want 200", resp.StatusCode)
	}
}

// TestTokenEndpointLimit checks that one address calls the token endpoint
// a burst of times, then waits.
func TestTokenEndpointLimit(t *testing.T) {
	f := setup(t)
	post := func() int {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "http://localhost/oauth/token",
			strings.NewReader("grant_type=authorization_code&code=nope"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "192.0.2.9:1234"
		rec := httptest.NewRecorder()
		f.web.ServeHTTP(rec, req)
		return rec.Code
	}
	for range TokenBurst {
		if code := post(); code == http.StatusTooManyRequests {
			t.Fatal("the token endpoint refused a call within its burst")
		}
	}
	if code := post(); code != http.StatusTooManyRequests {
		t.Errorf("the token endpoint past its burst = %d, want 429", code)
	}
}
