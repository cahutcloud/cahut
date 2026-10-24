package server

import (
	"context"
	"strings"
	"testing"
	"time"

	cmail "cahut.cloud/internal/mail"
	"cahut.cloud/internal/mail/mailtest"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/update"
)

func TestUpdateNotices(t *testing.T) {
	ctx := context.Background()
	state := t.TempDir()
	s, err := Open(ctx, Config{DataDir: t.TempDir(), UpdateDir: state, ConsoleURL: func(d string) string { return "https://" + d + "/console/" }})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if _, err := operation.Do(ctx, s.Registry, s.AddDomain, admin, AddDomainInput{Domain: "cahut.example.com"}); err != nil {
		t.Fatal(err)
	}
	// No administrator has an address yet: nothing is recorded, so that the
	// notice goes once one has.
	at := time.Date(2026, 10, 23, 4, 0, 0, 0, time.UTC)
	writeStatus(t, state, update.Status{Policy: update.Patch, Check: &update.Check{At: at, Latest: "0.7.0", URL: "https://example.com/v0.7.0"}})
	if err := s.noticeReleaseOver(ctx, "0.6.2"); err != nil {
		t.Fatal(err)
	}

	relay := mailtest.Start(t, mailtest.Options{NoStartTLS: true})
	in := cmail.SetMailRelayInput{Host: "127.0.0.1", Port: relay.Port, Security: cmail.Plain, From: "cahut@example.com"}
	if _, err := do[cmail.SetMailRelayInput, cmail.Delivery](ctx, s, "set_mail_relay", in); err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Do(ctx, s.Registry, CreateAdmin, admin, CreateAdminInput{Handle: "root", Name: "Root", Email: "root@example.org"}); err != nil {
		t.Fatal(err)
	}
	relay.Wait(t, 1) // the activation link

	// A release that the policy leaves to a person, once.
	for _, running := range []string{"0.6.2", "0.6.2", "0.7.0"} {
		if err := s.noticeReleaseOver(ctx, running); err != nil {
			t.Fatal(err)
		}
	}
	// A release of the same minor version installs on its own with patch.
	writeStatus(t, state, update.Status{Policy: update.Patch, Check: &update.Check{At: at, Latest: "0.6.3"}})
	if err := s.noticeReleaseOver(ctx, "0.6.2"); err != nil {
		t.Fatal(err)
	}
	// An update that failed, once.
	writeStatus(t, state, update.Status{Last: &update.Run{
		ID: "one", From: "0.6.2", To: "0.6.3", By: update.ByTimer, Result: update.RolledBack, Reason: "the health check failed", Ended: at,
	}})
	for range 2 {
		if err := s.LogUpdates(ctx); err != nil {
			t.Fatal(err)
		}
	}
	msgs := relay.Wait(t, 3)
	time.Sleep(200 * time.Millisecond)
	if n := len(relay.Messages()); n != 3 {
		t.Errorf("the relay received %d messages, want 3", n)
	}
	host := s.Mail.Host()
	for i, want := range []string{"Subject: Cahut 0.7.0 is ready to install on " + host, "Subject: An update failed on " + host} {
		b := strings.ReplaceAll(string(msgs[i+1].Data), "=\r\n", "")
		if msgs[i+1].To[0] != "root@example.org" || !strings.Contains(b, want) {
			t.Errorf("message %d:\n%s\nwant %q", i+2, b, want)
		}
	}
	if !strings.Contains(strings.ReplaceAll(string(msgs[2].Data), "=\r\n", ""), "failed: the health check failed. The server went back to 0.6.2.") {
		t.Errorf("the reason of the failed update is missing:\n%s", msgs[2].Data)
	}
}
