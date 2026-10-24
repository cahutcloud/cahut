package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/mail/mailtest"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/store"
	"cahut.cloud/internal/window"
)

// relay gives the server of f a relay in memory, on this machine and
// without TLS, and returns it.
func (f *fixture) relay(t *testing.T) *mailtest.Relay {
	t.Helper()
	relay := mailtest.Start(t, mailtest.Options{NoStartTLS: true})
	in := fmt.Sprintf(`{"host": "127.0.0.1", "port": %d, "security": "none", "from": "cahut@example.com"}`, relay.Port)
	if _, err := f.srv.Registry.Invoke(context.Background(), "set_mail_relay", admin, []byte(in)); err != nil {
		t.Fatal(err)
	}
	return relay
}

var invitationLink = regexp.MustCompile(`https://localhost/console/invitation#([A-Za-z0-9_-]+)`)

// body returns the data of the message m, with the soft line breaks of
// quoted-printable undone.
func body(m mailtest.Message) string {
	return strings.NewReplacer("=\r\n", "", "=3D", "=").Replace(string(m.Data))
}

func TestInvitations(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	relay := f.relay(t)
	olivier := f.person(t, "localhost", "olivier")
	camille := f.person(t, "localhost", "camille")
	f.person(t, "localhost", "bob") // without an address
	c := f.open(t, "localhost")
	if _, err := c.DB.Write.Exec("UPDATE users SET email = 'camille@example.org', name = 'Camille' WHERE handle = 'camille'"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.DB.Write.Exec("UPDATE users SET name = 'Olivier Durand' WHERE handle = 'olivier'"); err != nil {
		t.Fatal(err)
	}
	publishDeck(t, f, olivier)
	const path = "/olivier/q3-review/"
	invite := func(as operation.Actor, args string) (window.Invited, error) {
		out, err := c.Registry.Invoke(ctx, "invite_to_window", as, []byte(args))
		if err != nil {
			return window.Invited{}, err
		}
		return out.(window.Invited), nil
	}

	// A Window open to its Room only goes to people of the Cahut, by handle.
	if _, err := invite(olivier, `{"window": "olivier/q3-review", "to": ["ana@example.org"]}`); !errors.Is(err, operation.ErrInvalid) ||
		!strings.Contains(err.Error(), "set its access to key") {
		t.Errorf("invite_to_window an address to a Window of its Room: %v", err)
	}
	if _, err := invite(olivier, `{"window": "olivier/q3-review", "to": ["camille"]}`); !errors.Is(err, operation.ErrInvalid) ||
		!strings.Contains(err.Error(), "not in the Room") {
		t.Errorf("invite_to_window someone outside the Room: %v", err)
	}
	if _, err := invite(camille, `{"window": "olivier/q3-review", "to": ["ana@example.org"]}`); err == nil {
		t.Error("invite_to_window by someone outside the Room: no error")
	}

	key := f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key"}`).(window.AccessSet).Key
	out, err := invite(olivier, `{"window": "https://localhost/olivier/q3-review/", "to": ["Ana@Example.org", "camille", "bob", "ana@example.org"], "note": "Have a look\nbefore Friday."}`)
	if err != nil {
		t.Fatalf("invite_to_window: %v", err)
	}
	if len(out.Invitations) != 2 || out.Invitations[0].To != "Ana@example.org" || !out.Invitations[0].Link || out.Invitations[1].To != "camille" ||
		!out.Invitations[1].Link || len(out.Unreachable) != 1 || out.Unreachable[0] != "bob" {
		t.Fatalf("invite_to_window = %+v", out)
	}
	msgs := relay.Wait(t, 2)
	codes := map[string]string{}
	for _, m := range msgs {
		b := body(m)
		link := invitationLink.FindStringSubmatch(b)
		if link == nil || !strings.Contains(b, "Subject: Olivier Durand invites you to read Q3 review") ||
			!strings.Contains(b, "> before Friday.") || strings.Contains(b, key) {
			t.Errorf("the invitation to %v:\n%s", m.To, b)
			continue
		}
		codes[m.To[0]] = link[1]
	}
	if len(codes) != 2 || !strings.Contains(body(msgs[1])+body(msgs[0]), "Hello Camille,") {
		t.Fatalf("the invitations went to %v", codes)
	}

	// The link of the e-mail opens a page of the console, which opens the
	// Window on a click: a code for /-/invitation, which gives a pass.
	visitor := operation.Actor{Kind: operation.Visitor, Via: "api", Domain: "localhost"}
	page, err := c.Registry.Invoke(ctx, "get_invitation", visitor, []byte(`{"code": "`+codes["Ana@example.org"]+`"}`))
	if p, _ := page.(window.InvitationPage); err != nil || p.Title != "Q3 review" || p.By != "Olivier Durand" || p.Note != "Have a look\nbefore Friday." {
		t.Fatalf("get_invitation = %+v %v", page, err)
	}
	pass := func(code string) (*http.Cookie, response) {
		t.Helper()
		o, err := c.Registry.Invoke(ctx, "open_invitation", visitor, []byte(`{"code": "`+code+`"}`))
		if err != nil {
			t.Fatalf("open_invitation: %v", err)
		}
		opening := o.(window.InvitationOpening)
		resp, _ := f.do(t, "localhost", http.MethodGet, "/-/invitation?code="+opening.Code, nil, nil)
		for _, ck := range resp.Cookies() {
			if strings.HasPrefix(ck.Name, "__Host-cahut-key-") {
				return &http.Cookie{Name: ck.Name, Value: ck.Value}, resp //nolint:gosec // a cookie that the test sends back
			}
		}
		return nil, resp
	}
	ana, resp := pass(codes["Ana@example.org"])
	if ana == nil || resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != path {
		t.Fatalf("GET /-/invitation = %d %v", resp.StatusCode, resp.Header)
	}
	f.frame(t, path, ana)
	// The link works again on another device; the code of a step, once.
	second, _ := pass(codes["Ana@example.org"])
	if second == nil {
		t.Fatal("the personal link did not open the Window a second time")
	}
	if resp, body := f.do(t, "localhost", http.MethodGet, "/-/invitation?code=nothing", nil, nil); resp.StatusCode != http.StatusForbidden ||
		!strings.Contains(body, "Open the invitation again") {
		t.Errorf("GET /-/invitation with an unknown code = %d %q", resp.StatusCode, body)
	}

	// Revoking it closes the Window to Ana alone: the passes it gave end.
	list := f.call(t, "localhost", olivier, "list_invitations", `{"window": "olivier/q3-review"}`).(window.InvitationList)
	if len(list.Invitations) != 2 || list.Invitations[1].To != "Ana@example.org" || list.Invitations[1].By != "olivier" {
		t.Fatalf("list_invitations = %+v", list)
	}
	f.call(t, "localhost", olivier, "revoke_invitation", `{"invitation": `+strconv.FormatInt(list.Invitations[1].ID, 10)+`}`)
	for _, ck := range []*http.Cookie{ana, second} {
		if resp, body := f.do(t, "localhost", http.MethodGet, path, []*http.Cookie{ck}, nil); resp.StatusCode != http.StatusUnauthorized || frameSrc.MatchString(body) {
			t.Errorf("GET %s with the pass of a revoked invitation = %d", path, resp.StatusCode)
		}
	}
	if _, err := c.Registry.Invoke(ctx, "get_invitation", visitor, []byte(`{"code": "`+codes["Ana@example.org"]+`"}`)); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_invitation once revoked: %v", err)
	}
	camillePass, _ := pass(codes["camille@example.org"])
	f.frame(t, path, camillePass)

	// A new Key ends the invitations that stood for the old one.
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key", "new_key": true}`)
	if _, err := c.Registry.Invoke(ctx, "open_invitation", visitor, []byte(`{"code": "`+codes["camille@example.org"]+`"}`)); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("open_invitation after a new Key: %v", err)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, path, []*http.Cookie{camillePass}, nil); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("GET %s with a pass of the old Key = %d", path, resp.StatusCode)
	}

	// An AI invites without knocking, and the message says it sent it.
	ai := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: olivier.UserID, Handle: "olivier", TokenID: 1, Client: "Claude Code"}
	if _, err := invite(ai, `{"window": "olivier/q3-review", "to": ["joe@example.org"], "note": "From the deck we made."}`); err != nil {
		t.Fatalf("invite_to_window by an AI: %v", err)
	}
	if b := body(relay.Wait(t, 3)[2]); !strings.Contains(b, "Claude Code sent this invitation for Olivier Durand.") ||
		!strings.Contains(b, "Claude Code, for Olivier Durand wrote:") {
		t.Errorf("the invitation of an AI:\n%s", b)
	}
	entries, err := logbook.List(ctx, c.DB.Read, 50)
	if err != nil {
		t.Fatal(err)
	}
	var logged []string
	for _, e := range entries {
		if e.Operation == "invite_to_window" && e.Outcome == logbook.OK {
			logged = append(logged, e.Actor+" "+e.Target+" "+e.Detail)
		}
	}
	if strings.Join(logged, "; ") != "ai olivier/q3-review to joe@example.org; person olivier/q3-review to Ana@Example.org, camille, bob" {
		t.Errorf("the Logbook of invitations: %q", logged)
	}

	// In an individual Cahut, invitations go by address only.
	if _, err := operation.Do(ctx, c.Registry, settings.Set, admin, settings.SetCahutSettingsInput{Kind: settings.Individual}); err != nil {
		t.Fatal(err)
	}
	if _, err := invite(olivier, `{"window": "olivier/q3-review", "to": ["camille"]}`); !errors.Is(err, operation.ErrInvalid) ||
		!strings.Contains(err.Error(), "by e-mail address") {
		t.Errorf("invite_to_window by handle in an individual Cahut: %v", err)
	}

	// 20 people a call, and 100 invitations a day.
	many := make([]string, 21)
	for i := range many {
		many[i] = "p" + strconv.Itoa(i) + "@example.org"
	}
	args, _ := json.Marshal(window.InviteInput{Window: "olivier/q3-review", To: many})
	if _, err := invite(olivier, string(args)); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("invite_to_window 21 people: %v", err)
	}
	for range 97 {
		_, err := c.DB.Write.Exec("INSERT INTO invitations (window_id, invited_by, email, created_at) SELECT id, ?, 'x@example.org', ? FROM windows",
			olivier.UserID, store.FormatTime(time.Now()))
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := invite(olivier, `{"window": "olivier/q3-review", "to": ["one@example.org", "two@example.org"]}`); !errors.Is(err, operation.ErrTooMany) {
		t.Errorf("invite_to_window beyond 100 a day: %v", err)
	}

}
