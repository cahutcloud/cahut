package cli

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/mail/mailtest"
)

// reach lets the commands reach relay, on this machine, with its
// certificate, as a relay of a Cahut on the internet.
func reach(t *testing.T, relay *mailtest.Relay) {
	t.Helper()
	mailForTests = func(s *mail.Service) {
		s.UseRoots(relay.Roots)
		s.DialAnywhere()
	}
	t.Cleanup(func() { mailForTests = nil })
}

func TestMailCommands(t *testing.T) {
	data := t.TempDir()
	if code, _, errOut := run(t, data, "domain", "add", "cahut.example.com"); code != exitOK {
		t.Fatalf("domain add: %d %q", code, errOut)
	}
	relay := mailtest.Start(t, mailtest.Options{User: "cahut", Password: "s3cret pass"})
	port := strconv.Itoa(relay.Port)

	code, out, _ := run(t, data, "mail", "show")
	if code != exitOK || !strings.Contains(out, "cahut.example.com has no relay: it sends no e-mail") {
		t.Errorf("mail show: %d %q", code, out)
	}
	if code, _, errOut := run(t, data, "mail", "test", "--to", "jane@example.org"); code != exitError || !strings.Contains(errOut, "sends no e-mail") {
		t.Errorf("mail test without a relay: %d %q", code, errOut)
	}

	// The relay of a Cahut is on the internet.
	stdin = strings.NewReader("s3cret pass")
	t.Cleanup(func() { stdin = strings.NewReader("") })
	code, _, errOut := run(t, data, "mail", "set", "--host", "127.0.0.1", "--port", port,
		"--user", "cahut", "--password-stdin", "--from", "cahut@example.com")
	if code != exitError || !strings.Contains(errOut, "not a public address") {
		t.Errorf("mail set on this machine: %d %q", code, errOut)
	}

	reach(t, relay)
	stdin = strings.NewReader("s3cret pass\n")
	code, out, errOut = run(t, data, "mail", "set", "--domain", "cahut.example.com", "--host", "127.0.0.1", "--port", port,
		"--user", "cahut", "--password-stdin", "--from", "cahut@example.com", "--name", "Example")
	if code != exitOK || !strings.Contains(out, "sends the e-mail of cahut.example.com") ||
		!regexp.MustCompile(`host\s+127.0.0.1:`+port+`, starttls`).MatchString(out) || !regexp.MustCompile(`from\s+Example <cahut@example.com>`).MatchString(out) {
		t.Fatalf("mail set: %d %q %q", code, out, errOut)
	}
	if strings.Contains(out, "s3cret") {
		t.Errorf("mail set prints the password: %q", out)
	}

	code, out, errOut = run(t, data, "mail", "test", "--to", "jane@example.org")
	if code != exitOK || !strings.Contains(out, "Sent a test from cahut.example.com to jane@example.org, through its relay") {
		t.Errorf("mail test: %d %q %q", code, out, errOut)
	}
	if msgs := relay.Wait(t, 1); msgs[0].To[0] != "jane@example.org" {
		t.Errorf("the test went to %v", msgs[0].To)
	}

	code, out, _ = run(t, data, "mail", "remove")
	if code != exitOK || !strings.Contains(out, "Removed the relay of cahut.example.com") || !strings.Contains(out, "sends no e-mail") {
		t.Errorf("mail remove: %d %q", code, out)
	}
	if code, _, errOut := run(t, data, "mail", "set", "--from", "cahut@example.com"); code != exitUsage || !strings.Contains(errOut, "give --host") {
		t.Errorf("mail set without --host: %d %q", code, errOut)
	}
}
