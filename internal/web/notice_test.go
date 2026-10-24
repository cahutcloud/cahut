package web

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/mail/mailtest"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
	"cahut.cloud/internal/settings"
)

// subjects returns the subjects of the messages that the relay received,
// with their recipient, once at least n arrived.
func subjects(t *testing.T, relay *mailtest.Relay, n int) []string {
	t.Helper()
	var out []string
	for _, m := range relay.Wait(t, n) {
		b := body(m)
		_, rest, _ := strings.Cut(b, "\r\nSubject: ")
		subject, _, _ := strings.Cut(rest, "\r\n")
		out = append(out, m.To[0]+": "+subject)
	}
	return out
}

func TestNotices(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	relay := f.relay(t)
	// The disk of the machine that runs the test is its own business.
	f.web.disk = func(string) (uint64, uint64, error) { return 100 << 30, 50 << 30, nil }
	olivier := f.person(t, "localhost", "olivier")
	camille := f.person(t, "localhost", "camille")
	c := f.open(t, "localhost")
	for _, q := range []string{
		"UPDATE users SET email = 'olivier@example.org', name = 'Olivier', role = 'admin' WHERE handle = 'olivier'",
		"UPDATE users SET email = 'camille@example.org', name = 'Camille' WHERE handle = 'camille'",
	} {
		if _, err := c.DB.Write.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	publishDeck(t, f, olivier)

	// A new Knock of an AI reaches its person; the same call again makes no
	// new Knock, and no new message.
	ai := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: olivier.UserID, Handle: "olivier", TokenID: 1, Client: "Claude Code"}
	for range 2 {
		_, err := c.Registry.Invoke(ctx, "set_window_access", ai, []byte(`{"window": "olivier/q3-review", "access": "public"}`))
		if !errors.Is(err, operation.ErrApprovalRequired) {
			t.Fatalf("set_window_access public by an AI: %v", err)
		}
	}
	m := relay.Wait(t, 1)[0]
	if b := body(m); m.To[0] != "olivier@example.org" || !strings.Contains(b, "Subject: Claude Code is waiting for your answer") ||
		!strings.Contains(b, "It asks to set who opens a Window: olivier/q3-review.") || !strings.Contains(b, "https://localhost/console/knocks/1") {
		t.Errorf("the message of a Knock:\n%s", b)
	}

	// Being added to a Room is a notice that a person may turn off.
	f.call(t, "localhost", olivier, "create_room", `{"title": "Marketing"}`)
	f.call(t, "localhost", olivier, "add_room_member", `{"room": "marketing", "user": "camille"}`)
	if got := subjects(t, relay, 2)[1]; got != "camille@example.org: You were added to Marketing" {
		t.Errorf("the notice of a new member: %q", got)
	}
	f.call(t, "localhost", camille, "set_mail_notices", `{"off": true}`)
	f.call(t, "localhost", olivier, "create_room", `{"title": "Sales"}`)
	f.call(t, "localhost", olivier, "add_room_member", `{"room": "sales", "user": "camille"}`)

	// The access of a Window that ends within three days reaches the
	// person who set it, once.
	ends := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "internal", "ends_at": "`+ends+`"}`)
	f.web.SendNotices(ctx, time.Now())
	f.web.SendNotices(ctx, time.Now())

	// A limit reached reaches the administrators of the Cahut, and those
	// of the server, once a day.
	if _, err := operation.Do(ctx, f.srv.Registry, server.CreateAdmin, admin, server.CreateAdminInput{Handle: "root", Name: "Root", Email: "root@example.org"}); err != nil {
		t.Fatal(err)
	}
	one := int64(1)
	if _, err := operation.Do(ctx, c.Registry, settings.Set, admin, settings.SetCahutSettingsInput{Limits: settings.Limits{Windows: &one}}); err != nil {
		t.Fatal(err)
	}
	f.web.SendNotices(ctx, time.Now())
	f.web.SendNotices(ctx, time.Now())

	got := subjects(t, relay, 6)
	want := []string{
		"olivier@example.org: Claude Code is waiting for your answer",
		"camille@example.org: You were added to Marketing",
		"olivier@example.org: Q3 review closes to its readers on " + time.Now().UTC().Add(48*time.Hour).Format("2 January"),
		"root@example.org: Choose your password for the Cahut server",
		"olivier@example.org: localhost reached its limit of Windows",
		"root@example.org: localhost reached its limit of Windows",
	}
	time.Sleep(200 * time.Millisecond)
	if n := len(relay.Messages()); n != len(want) || !slices.Equal(slices.Sorted(slices.Values(got)), slices.Sorted(slices.Values(want))) {
		t.Errorf("the messages (%d):\n%s\nwant\n%s", n, strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
