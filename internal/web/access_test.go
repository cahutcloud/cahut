package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/window"
)

// do sends a request with cookies, and a form when one is given, and
// returns the response and its body.
func (f *fixture) do(t *testing.T, host, method, path string, cookies []*http.Cookie, form url.Values) (response, string) {
	t.Helper()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(context.Background(), method, f.ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := f.client(host, "").Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return response{StatusCode: resp.StatusCode, Header: resp.Header}, string(b)
}

// Cookies returns the cookies that the response sets.
func (r response) Cookies() []*http.Cookie {
	return (&http.Response{Header: r.Header}).Cookies()
}

// person returns the actor of the person handle in the console of the
// Cahut at domain, creating them if needed.
func (f *fixture) person(t *testing.T, domain, handle string) operation.Actor {
	t.Helper()
	c := f.open(t, domain)
	var id int64
	err := c.DB.Read.QueryRow("SELECT id FROM users WHERE handle = ?", handle).Scan(&id)
	if err != nil {
		u, err := operation.Do(context.Background(), c.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: handle, Name: handle})
		if err != nil {
			t.Fatal(err)
		}
		id = u.ID
	}
	return operation.Actor{Kind: operation.Person, Via: "api", UserID: id, Handle: handle, Session: 1}
}

// call runs an operation of the Cahut at domain for actor.
func (f *fixture) call(t *testing.T, domain string, actor operation.Actor, op, args string) any {
	t.Helper()
	out, err := f.open(t, domain).Registry.Invoke(context.Background(), op, actor, []byte(args))
	if err != nil {
		t.Fatalf("%s: %v", op, err)
	}
	return out
}

// signIn signs the person handle in to read the Windows of the Cahut at
// domain, as the console does, and returns their reader cookie.
func (f *fixture) signIn(t *testing.T, domain, handle string) *http.Cookie {
	t.Helper()
	code := f.call(t, domain, f.person(t, domain, handle), "sign_in_to_windows", `{}`).(account.ReaderCode).Code
	resp, _ := f.do(t, domain, http.MethodGet, "/-/session?code="+code+"&to=%2Folivier%2F", nil, nil)
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/olivier/" {
		t.Fatalf("GET /-/session = %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	for _, c := range resp.Cookies() {
		if c.Name == "__Host-cahut-reader" {
			if !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode || c.Path != "/" {
				t.Errorf("reader cookie = %+v", c)
			}
			return &http.Cookie{Name: c.Name, Value: c.Value} //nolint:gosec // a cookie that the test sends back
		}
	}
	t.Fatalf("no reader cookie in %v", resp.Header)
	return nil
}

var frameSrc = regexp.MustCompile(`<iframe src="(/-/w/[^"]+)"`)

// frame opens the Window at path as a reader with cookies, and returns the
// address of its frame, or fails.
func (f *fixture) frame(t *testing.T, path string, cookies ...*http.Cookie) string {
	t.Helper()
	resp, body := f.do(t, "localhost", http.MethodGet, path, cookies, nil)
	m := frameSrc.FindStringSubmatch(body)
	if resp.StatusCode != http.StatusOK || m == nil {
		t.Fatalf("GET %s = %d %q, want a frame", path, resp.StatusCode, body)
	}
	if csp := resp.Header.Get("Content-Security-Policy"); csp != pagePolicy {
		t.Errorf("the frame of %s has the policy %q", path, csp)
	}
	return strings.ReplaceAll(m[1], "&amp;", "&")
}

func publishDeck(t *testing.T, f *fixture, actor operation.Actor) window.Window {
	t.Helper()
	return f.call(t, "localhost", actor, "publish_window", `{"title": "Q3 review", "kind": "deck", "declaration": {"scripts": false}, "files": [
		{"path": "index.html", "content": "<h1>Q3</h1>"}, {"path": "css/deck.css", "content": "h1 { color: teal }"}]}`).(window.Window)
}

func TestProtectedWindows(t *testing.T) {
	f := setup(t)
	olivier, camille := f.person(t, "localhost", "olivier"), f.person(t, "localhost", "camille")
	publishDeck(t, f, olivier)
	const path = "/olivier/q3-review/"
	readOlivier, readCamille := f.signIn(t, "localhost", "olivier"), f.signIn(t, "localhost", "camille")

	// Room only: its members open it, in a frame; nobody else finds it.
	for who, cookies := range map[string][]*http.Cookie{"nobody": nil, "camille": {readCamille}} {
		resp, body := f.do(t, "localhost", http.MethodGet, path, cookies, nil)
		if resp.StatusCode != http.StatusNotFound || strings.Contains(body, "Q3") {
			t.Errorf("GET %s by %s = %d %q, want 404", path, who, resp.StatusCode, body)
		}
		if signIn := strings.Contains(body, `href="/-/sign-in?to=%2Folivier%2Fq3-review%2F"`); signIn != (who == "nobody") {
			t.Errorf("the page of %s offers to sign in: %v", who, signIn)
		}
	}
	src := f.frame(t, path, readOlivier)
	resp, body := f.do(t, "localhost", http.MethodGet, src, nil, nil)
	if resp.StatusCode != http.StatusOK || body != "<h1>Q3</h1>" {
		t.Fatalf("GET %s = %d %q", src, resp.StatusCode, body)
	}
	if h := resp.Header; !strings.HasPrefix(h.Get("Content-Security-Policy"), "sandbox ") || h.Get("Referrer-Policy") != "no-referrer" ||
		h.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("the frame's file has the headers %v", h)
	}
	if resp, body := f.do(t, "localhost", http.MethodGet, src+"css/deck.css", nil, nil); resp.StatusCode != http.StatusOK || !strings.Contains(body, "teal") {
		t.Errorf("GET %scss/deck.css = %d %q", src, resp.StatusCode, body)
	}
	// A deep address keeps its path and its query in the frame.
	if deep := f.frame(t, path+"css/deck.css?x=1", readOlivier); !strings.HasSuffix(deep, "/css/deck.css?x=1") {
		t.Errorf("the frame of a deep address is %q", deep)
	}

	// Internal: everyone signed in to the Cahut.
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "internal"}`)
	srcCamille := f.frame(t, path, readCamille)
	if resp, _ := f.do(t, "localhost", http.MethodGet, path, nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("an internal Window opens to nobody: %d", resp.StatusCode)
	}
	// Once its reader is disabled, the frame no longer opens.
	if _, err := operation.Do(context.Background(), f.open(t, "localhost").Registry, account.DisableUser, admin, account.UserInput{User: "camille"}); err != nil {
		t.Fatal(err)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, srcCamille, nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("the frame of a disabled reader = %d, want 404", resp.StatusCode)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, path, []*http.Cookie{readCamille}, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("the Window for a disabled reader = %d, want 404", resp.StatusCode)
	}
	_ = camille

	// The capability is signed, and bound to its Window and its Cahut.
	c := f.open(t, "localhost")
	signed := strings.Split(src, "/")[3]
	tampered := []byte(signed)
	tampered[5] ^= 1
	expired := capability{Window: 1, Kind: bySession, ID: 1, Expires: time.Now().Add(-time.Minute).Unix()}.sign(c.FrameKey)
	other := capability{Window: 2, Kind: bySession, ID: 1, Expires: time.Now().Add(time.Hour).Unix()}.sign(c.FrameKey)
	for name, capa := range map[string]string{"tampered": string(tampered), "expired": expired, "of a missing Window": other, "garbage": "abc"} {
		if resp, _ := f.do(t, "localhost", http.MethodGet, "/-/w/"+capa+"/", nil, nil); resp.StatusCode != http.StatusNotFound {
			t.Errorf("a capability %s = %d, want 404", name, resp.StatusCode)
		}
	}
	if resp, _ := f.do(t, "beta.localhost", http.MethodGet, src, nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("a capability of another Cahut = %d, want 404", resp.StatusCode)
	}

	// In the trash, the Window and its frames close.
	f.call(t, "localhost", olivier, "trash_window", `{"window": "olivier/q3-review"}`)
	for _, p := range []string{path, src} {
		if resp, _ := f.do(t, "localhost", http.MethodGet, p, []*http.Cookie{readOlivier}, nil); resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s in the trash = %d, want 404", p, resp.StatusCode)
		}
	}
}

func TestKeys(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	w := publishDeck(t, f, olivier)
	const path = "/olivier/q3-review/"
	key := f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key"}`).(window.AccessSet).Key

	resp, body := f.do(t, "localhost", http.MethodGet, path, nil, nil)
	if resp.StatusCode != http.StatusUnauthorized || !strings.Contains(body, `action="/-/key"`) || !strings.Contains(body, "Q3 review") {
		t.Fatalf("GET %s = %d %q, want the page of its Key", path, resp.StatusCode, body)
	}
	form := url.Values{"window": {"olivier/q3-review"}, "key": {"not the key"}}
	if resp, body := f.do(t, "localhost", http.MethodPost, "/-/key", nil, form); resp.StatusCode != http.StatusForbidden || !strings.Contains(body, "does not open") {
		t.Errorf("a wrong Key = %d %q", resp.StatusCode, body)
	}
	form.Set("key", key)
	resp, _ = f.do(t, "localhost", http.MethodPost, "/-/key", nil, form)
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != path {
		t.Fatalf("the right Key = %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	var pass *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "__Host-cahut-key-"+strconv.FormatInt(f.windowID(t, w.Name), 10) && c.HttpOnly && c.Secure && c.MaxAge > 0 {
			pass = &http.Cookie{Name: c.Name, Value: c.Value} //nolint:gosec // a cookie that the test sends back
		}
	}
	if pass == nil {
		t.Fatalf("no pass in %v", resp.Header)
	}
	src := f.frame(t, path, pass)
	if resp, body := f.do(t, "localhost", http.MethodGet, src, nil, nil); resp.StatusCode != http.StatusOK || body != "<h1>Q3</h1>" {
		t.Errorf("GET %s = %d %q", src, resp.StatusCode, body)
	}
	// The Room lists what its reader opens.
	if resp, body := f.do(t, "localhost", http.MethodGet, "/olivier/", []*http.Cookie{pass}, nil); resp.StatusCode != http.StatusOK || !strings.Contains(body, `href="/olivier/q3-review/"`) {
		t.Errorf("GET /olivier/ with a pass = %d %q", resp.StatusCode, body)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, "/olivier/", nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET /olivier/ without a pass = %d, want 404", resp.StatusCode)
	}

	// A new Key ends the pass, and the frames it opened.
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key", "new_key": true}`)
	if resp, _ := f.do(t, "localhost", http.MethodGet, path, []*http.Cookie{pass}, nil); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("an old pass = %d, want the page of the Key", resp.StatusCode)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, src, nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("a frame of an old pass = %d, want 404", resp.StatusCode)
	}

	// A Window without a Key takes none, and Keys are limited by address.
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "room"}`)
	if resp, _ := f.do(t, "localhost", http.MethodPost, "/-/key", nil, form); resp.StatusCode != http.StatusNotFound {
		t.Errorf("a Key for a Window without one = %d, want 404", resp.StatusCode)
	}
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key", "key": "correct horse battery"}`)
	form.Set("key", "wrong")
	var status int
	for range KeyBurst + 1 {
		resp, _ := f.do(t, "localhost", http.MethodPost, "/-/key", nil, form)
		status = resp.StatusCode
	}
	if status != http.StatusTooManyRequests {
		t.Errorf("after %d tries, a Key = %d, want 429", KeyBurst+1, status)
	}
}

func (f *fixture) windowID(t *testing.T, name string) int64 {
	t.Helper()
	w, err := window.Lookup(context.Background(), f.open(t, "localhost").DB, "olivier", name, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return w.ID
}

func TestReaders(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	publishDeck(t, f, olivier)

	// Signing in goes through the console, and comes back to this host only.
	for to, want := range map[string]string{
		"/olivier/q3-review/?x=1": "/olivier/q3-review/?x=1",
		"//evil.example/":         "/",
		"https://evil.example/":   "/",
		`/\evil.example/`:         "/",
		"":                        "/",
	} {
		resp, _ := f.do(t, "localhost", http.MethodGet, "/-/sign-in?to="+url.QueryEscape(to), nil, nil)
		if loc := resp.Header.Get("Location"); resp.StatusCode != http.StatusSeeOther || loc != "https://localhost/console/read?to="+url.QueryEscape(want) {
			t.Errorf("sign in to %q = %d %q", to, resp.StatusCode, loc)
		}
	}
	code := f.call(t, "localhost", olivier, "sign_in_to_windows", `{}`).(account.ReaderCode).Code
	resp, _ := f.do(t, "localhost", http.MethodGet, "/-/session?code="+code+"&to=https://evil.example/", nil, nil)
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/" {
		t.Errorf("a session that goes elsewhere = %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	// A code works once, and only in its Cahut.
	if resp, _ := f.do(t, "localhost", http.MethodGet, "/-/session?code="+code, nil, nil); resp.StatusCode != http.StatusForbidden {
		t.Errorf("a used code = %d, want 403", resp.StatusCode)
	}
	code = f.call(t, "localhost", olivier, "sign_in_to_windows", `{}`).(account.ReaderCode).Code
	if resp, _ := f.do(t, "beta.localhost", http.MethodGet, "/-/session?code="+code, nil, nil); resp.StatusCode != http.StatusForbidden {
		t.Errorf("a code of another Cahut = %d, want 403", resp.StatusCode)
	}
	// The administrators of the server read no Windows.
	if _, err := f.open(t, "localhost").Registry.Invoke(context.Background(), "sign_in_to_windows",
		operation.Actor{Kind: operation.Admin, Via: "api", Handle: "root", Session: 1}, []byte(`{}`)); err == nil {
		t.Error("an administrator of the server signed in to read Windows")
	}
	// A reader session is no session of the console.
	reader := f.signIn(t, "localhost", "olivier")
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "http://localhost/console/api/get_account", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+reader.Value)
	rec := httptest.NewRecorder()
	f.web.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("the API with a reader session = %d, want 401", rec.Code)
	}

	// The home lists the reader's Rooms, and signing out ends the session.
	if _, body := f.do(t, "localhost", http.MethodGet, "/", []*http.Cookie{reader}, nil); !strings.Contains(body, `href="/olivier/"`) || !strings.Contains(body, "Sign out") {
		t.Errorf("GET / signed in = %q", body)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, "/olivier/", []*http.Cookie{reader}, nil); resp.StatusCode != http.StatusOK {
		t.Errorf("GET /olivier/ by its member = %d", resp.StatusCode)
	}
	resp, _ = f.do(t, "localhost", http.MethodPost, "/-/sign-out", []*http.Cookie{reader}, url.Values{})
	if resp.StatusCode != http.StatusSeeOther || len(resp.Cookies()) != 1 || resp.Cookies()[0].MaxAge >= 0 {
		t.Errorf("sign out = %d %v", resp.StatusCode, resp.Header)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, "/olivier/q3-review/", []*http.Cookie{reader}, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("after signing out, the Window = %d, want 404", resp.StatusCode)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, "/olivier", nil, nil); resp.StatusCode != http.StatusPermanentRedirect || resp.Header.Get("Location") != "/olivier/" {
		t.Errorf("GET /olivier = %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
}

func TestDeclaredPolicy(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	publishDeck(t, f, olivier) // declares no scripts
	reader := f.signIn(t, "localhost", "olivier")
	_, body := f.do(t, "localhost", http.MethodGet, "/olivier/q3-review/", []*http.Cookie{reader}, nil)
	if !strings.Contains(body, `sandbox="allow-forms allow-modals`) {
		t.Errorf("the frame of a Window without scripts = %q", body)
	}
	src := f.frame(t, "/olivier/q3-review/", reader)
	resp, _ := f.do(t, "localhost", http.MethodGet, src, nil, nil)
	if csp := resp.Header.Get("Content-Security-Policy"); !strings.HasPrefix(csp, "sandbox allow-forms ") ||
		!strings.Contains(csp, "connect-src https://localhost;") || strings.Contains(csp, "script-src") {
		t.Errorf("the policy of a Window without scripts = %q", csp)
	}

	f.call(t, "localhost", olivier, "publish_window", `{"title": "Chart", "kind": "dashboard",
		"declaration": {"scripts": true, "origins": ["https://cdn.jsdelivr.net"]}, "files": [{"path": "index.html", "content": "x"}]}`)
	f.setAccess(t, "localhost", "chart", window.AccessPublic)
	resp, _ = f.do(t, "localhost", http.MethodGet, "/olivier/chart/", nil, nil)
	if csp := resp.Header.Get("Content-Security-Policy"); !strings.HasPrefix(csp, "sandbox allow-scripts ") ||
		!strings.Contains(csp, "script-src https://localhost 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net;") {
		t.Errorf("the policy of a public Window with scripts = %q", csp)
	}
}
