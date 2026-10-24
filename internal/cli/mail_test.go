package cli

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"cahut.cloud/internal/mail/mailtest"
)

func TestMailCommands(t *testing.T) {
	data := t.TempDir()
	if code, _, errOut := run(t, data, "domain", "add", "cahut.example.com"); code != exitOK {
		t.Fatalf("domain add: %d %q", code, errOut)
	}
	// A relay on this machine, without TLS: the relay of the server alone
	// may do without it.
	relay := mailtest.Start(t, mailtest.Options{NoStartTLS: true, User: "cahut", Password: "s3cret pass"})
	port := strconv.Itoa(relay.Port)

	code, out, _ := run(t, data, "mail", "show")
	if code != exitOK || !strings.Contains(out, "The server has no relay") {
		t.Errorf("mail show: %d %q", code, out)
	}
	code, out, _ = run(t, data, "mail", "show", "--domain", "cahut.example.com")
	if code != exitOK || !strings.Contains(out, "cahut.example.com has no relay, and the server has none") {
		t.Errorf("mail show --domain: %d %q", code, out)
	}

	stdin = strings.NewReader("s3cret pass\n")
	t.Cleanup(func() { stdin = strings.NewReader("") })
	code, out, errOut := run(t, data, "mail", "set", "--host", "127.0.0.1", "--port", port, "--security", "none",
		"--user", "cahut", "--password-stdin", "--from", "cahut@example.com", "--name", "Example")
	if code != exitOK || !strings.Contains(out, "sends the e-mail of the server") ||
		!regexp.MustCompile(`host\s+127.0.0.1:`+port+`, none`).MatchString(out) || !regexp.MustCompile(`from\s+Example <cahut@example.com>`).MatchString(out) {
		t.Fatalf("mail set: %d %q %q", code, out, errOut)
	}
	if strings.Contains(out, "s3cret") {
		t.Errorf("mail set prints the password: %q", out)
	}

	// The Cahut, without a relay of its own, sends through that of the server.
	code, out, errOut = run(t, data, "mail", "test", "--domain", "cahut.example.com", "--to", "jane@example.org")
	if code != exitOK || !strings.Contains(out, "Sent a test from cahut.example.com to jane@example.org, through the relay of the server") {
		t.Errorf("mail test --domain: %d %q %q", code, out, errOut)
	}
	if msgs := relay.Wait(t, 1); msgs[0].To[0] != "jane@example.org" {
		t.Errorf("the test went to %v", msgs[0].To)
	}

	// The relay of a Cahut is on the internet: a relay here is the server's.
	stdin = strings.NewReader("s3cret pass")
	code, _, errOut = run(t, data, "mail", "set", "--domain", "cahut.example.com", "--host", "127.0.0.1", "--port", port,
		"--user", "cahut", "--password-stdin", "--from", "cahut@example.com")
	if code != exitError || !strings.Contains(errOut, "not a public address") {
		t.Errorf("mail set --domain on this machine: %d %q", code, errOut)
	}

	code, out, _ = run(t, data, "mail", "remove")
	if code != exitOK || !strings.Contains(out, "Removed the relay of the server") {
		t.Errorf("mail remove: %d %q", code, out)
	}
	code, _, errOut = run(t, data, "mail", "test", "--to", "jane@example.org")
	if code != exitError || !strings.Contains(errOut, "no relay") {
		t.Errorf("mail test without a relay: %d %q", code, errOut)
	}
	if code, _, errOut := run(t, data, "mail", "set", "--from", "cahut@example.com"); code != exitUsage || !strings.Contains(errOut, "give --host") {
		t.Errorf("mail set without --host: %d %q", code, errOut)
	}
}
