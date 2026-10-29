package web

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/window"
)

func TestPermanentAddresses(t *testing.T) {
	f := setup(t)
	relay := f.relay(t)
	olivier := f.person(t, "localhost", "olivier")
	f.person(t, "localhost", "camille")
	w := publishDeck(t, f, olivier)
	perm := window.PermanentPrefix + w.ID
	if !address.ValidID(w.ID) || w.PermanentURL != "https://localhost"+perm {
		t.Fatalf("publish_window = %+v", w)
	}
	readOlivier, readCamille := f.signIn(t, "localhost", "olivier"), f.signIn(t, "localhost", "camille")
	redirects := func(path string, cookies []*http.Cookie, to string) {
		t.Helper()
		resp, _ := f.do(t, "localhost", http.MethodGet, path, cookies, nil)
		if resp.StatusCode != http.StatusTemporaryRedirect || resp.Header.Get("Location") != to {
			t.Errorf("GET %s = %d %q, want a redirect to %s", path, resp.StatusCode, resp.Header.Get("Location"), to)
		}
	}

	// Room only: its members go to its address, with the rest of the path and
	// the query; nobody else learns the address.
	redirects(perm, []*http.Cookie{readOlivier}, "/olivier/q3-review/")
	redirects(perm+"/", []*http.Cookie{readOlivier}, "/olivier/q3-review/")
	redirects(perm+"/css/deck.css?x=1", []*http.Cookie{readOlivier}, "/olivier/q3-review/css/deck.css?x=1")
	for who, cookies := range map[string][]*http.Cookie{"nobody": nil, "camille": {readCamille}} {
		resp, body := f.do(t, "localhost", http.MethodGet, perm, cookies, nil)
		if resp.StatusCode != http.StatusNotFound || strings.Contains(body, "q3-review") || strings.Contains(body, "Q3") {
			t.Errorf("GET %s by %s = %d %q, want 404", perm, who, resp.StatusCode, body)
		}
		if signIn := strings.Contains(body, `href="/-/sign-in?to=`+url.QueryEscape(perm)+`"`); signIn != (who == "nobody") {
			t.Errorf("the page of %s offers to sign in back to the permanent address: %v", who, signIn)
		}
	}

	// Public: anyone goes, and invitations give the permanent address.
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "`+w.ID+`", "access": "public"}`)
	redirects(perm, nil, "/olivier/q3-review/")
	f.call(t, "localhost", olivier, "invite_to_window", `{"window": "`+w.ID+`", "to": ["ana@example.org"]}`)
	if b := body(relay.Wait(t, 1)[0]); !strings.Contains(b, w.PermanentURL) || strings.Contains(b, "/olivier/q3-review/") {
		t.Errorf("the invitation to a public Window:\n%s", b)
	}

	// Key: the page of its Key names the Window by its id, not its address.
	key := f.call(t, "localhost", olivier, "set_window_access", `{"window": "`+w.ID+`", "access": "key"}`).(window.AccessSet).Key
	resp, page := f.do(t, "localhost", http.MethodGet, perm, nil, nil)
	if resp.StatusCode != http.StatusUnauthorized || !strings.Contains(page, `action="/-/key"`) ||
		!strings.Contains(page, `name="window" value="`+w.ID+`"`) || strings.Contains(page, "q3-review") {
		t.Fatalf("GET %s = %d %q, want the page of its Key", perm, resp.StatusCode, page)
	}
	form := url.Values{"window": {w.ID}, "key": {"not the key"}}
	if resp, body := f.do(t, "localhost", http.MethodPost, "/-/key", nil, form); resp.StatusCode != http.StatusForbidden ||
		!strings.Contains(body, "does not open") || strings.Contains(body, "q3-review") {
		t.Errorf("a wrong Key at the permanent address = %d %q", resp.StatusCode, body)
	}
	form.Set("key", key)
	resp, _ = f.do(t, "localhost", http.MethodPost, "/-/key", nil, form)
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/olivier/q3-review/" || len(resp.Cookies()) != 1 {
		t.Fatalf("the right Key at the permanent address = %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	pass := &http.Cookie{Name: resp.Cookies()[0].Name, Value: resp.Cookies()[0].Value} //nolint:gosec // a cookie that the test sends back
	redirects(perm, []*http.Cookie{pass}, "/olivier/q3-review/")

	// In the trash, a Window is not found at its permanent address, nor in
	// another Cahut, nor at an id that names no Window.
	f.call(t, "localhost", olivier, "trash_window", `{"window": "`+w.ID+`"}`)
	for _, p := range []string{perm, perm + "/index.html"} {
		if resp, _ := f.do(t, "localhost", http.MethodGet, p, []*http.Cookie{readOlivier}, nil); resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s in the trash = %d, want 404", p, resp.StatusCode)
		}
	}
	f.call(t, "localhost", olivier, "restore_window", `{"window": "`+w.ID+`"}`)
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "`+w.ID+`", "access": "public"}`)
	for host, p := range map[string]string{
		"beta.localhost": perm,
		"localhost":      window.PermanentPrefix + address.NewID(),
	} {
		if resp, _ := f.do(t, host, http.MethodGet, p, nil, nil); resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s on %s = %d, want 404", p, host, resp.StatusCode)
		}
	}
	for _, p := range []string{"/-/w/garbage", "/-/w/" + strings.ToUpper(w.ID), "/-/w/"} {
		if resp, _ := f.do(t, "localhost", http.MethodGet, p, nil, nil); resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", p, resp.StatusCode)
		}
	}
}

// A new name ends the old address at once, and what holds the Window by its
// id follows it: its permanent address, the passes of its Key and the home
// of its Room.
func TestRenameKeepsWhatHoldsTheWindow(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	w := publishDeck(t, f, olivier)
	key := f.call(t, "localhost", olivier, "set_window_access", `{"window": "`+w.ID+`", "access": "key"}`).(window.AccessSet).Key
	f.call(t, "localhost", olivier, "set_room_home", `{"room": "olivier", "window": "q3-review"}`)
	resp, _ := f.do(t, "localhost", http.MethodPost, "/-/key", nil, url.Values{"window": {"olivier/q3-review"}, "key": {key}})
	if resp.StatusCode != http.StatusSeeOther || len(resp.Cookies()) != 1 {
		t.Fatalf("the right Key = %d", resp.StatusCode)
	}
	pass := &http.Cookie{Name: resp.Cookies()[0].Name, Value: resp.Cookies()[0].Value} //nolint:gosec // a cookie that the test sends back

	renamed := f.call(t, "localhost", olivier, "rename_window", `{"window": "olivier/q3-review", "name": "q3"}`).(window.Window)
	if renamed.URL != "https://localhost/olivier/q3/" || renamed.PermanentURL != w.PermanentURL {
		t.Fatalf("rename_window = %+v", renamed)
	}
	if resp, body := f.do(t, "localhost", http.MethodGet, "/olivier/q3-review/", []*http.Cookie{pass}, nil); resp.StatusCode != http.StatusNotFound ||
		strings.Contains(body, "Q3") {
		t.Errorf("the old address = %d %q, want 404", resp.StatusCode, body)
	}
	f.frame(t, "/olivier/q3/", pass)
	resp, _ = f.do(t, "localhost", http.MethodGet, window.PermanentPrefix+w.ID+"/css/deck.css", []*http.Cookie{pass}, nil)
	if resp.StatusCode != http.StatusTemporaryRedirect || resp.Header.Get("Location") != "/olivier/q3/css/deck.css" {
		t.Errorf("the permanent address after a new name = %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	if d := f.call(t, "localhost", olivier, "get_room", `{"room": "olivier"}`).(room.Detail); d.Home != "q3" {
		t.Errorf("the home of the Room after a new name = %q", d.Home)
	}
	// The address of the Room still shows its home, in a frame.
	f.frame(t, "/olivier/", pass)
}
