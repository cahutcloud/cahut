package cli

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"cahut.cloud/internal/mail/mailtest"
)

func TestEmailCommands(t *testing.T) {
	data := t.TempDir()
	if code, _, errOut := run(t, data, "domain", "add", "cahut.example.com"); code != exitOK {
		t.Fatalf("domain add: %d %q", code, errOut)
	}
	relay := mailtest.Start(t, mailtest.Options{})
	reach(t, relay)
	if code, _, errOut := run(t, data, "mail", "set", "--host", "127.0.0.1", "--port", strconv.Itoa(relay.Port),
		"--from", "cahut@example.com"); code != exitOK {
		t.Fatalf("mail set: %d %q", code, errOut)
	}

	// The link goes to the address of the person, and shows here too.
	code, out, errOut := run(t, data, "user", "create", "--handle", "olivier", "--name", "Olivier", "--admin", "--email", "olivier@example.org")
	if code != exitOK || !strings.Contains(out, "went to their e-mail address") {
		t.Fatalf("user create --email: %d %q %q", code, out, errOut)
	}
	printed := regexp.MustCompile(`https://cahut\.example\.com/console/activate#\S+`).FindString(out)
	m := relay.Wait(t, 1)[0]
	if printed == "" || m.To[0] != "olivier@example.org" || !strings.Contains(strings.ReplaceAll(string(m.Data), "=\r\n", ""), strings.ReplaceAll(printed, "=", "=3D")) {
		t.Errorf("the link printed %q is not the one sent:\n%s", printed, m.Data)
	}

	code, out, _ = run(t, data, "user", "list")
	if code != exitOK || !regexp.MustCompile(`olivier\s+Olivier\s+admin\s+pending\s+olivier@example.org`).MatchString(out) {
		t.Errorf("user list: %d %q", code, out)
	}
	code, out, _ = run(t, data, "user", "email", "--handle", "olivier", "--email", "Olivier@Example.ORG")
	if code != exitOK || !strings.Contains(out, "is Olivier@example.org now") {
		t.Errorf("user email: %d %q", code, out)
	}
	code, out, _ = run(t, data, "user", "email", "--handle", "olivier", "--email", "")
	if code != exitOK || !strings.Contains(out, "has no e-mail address now") {
		t.Errorf("user email without an address: %d %q", code, out)
	}
	// The old address learns that it was removed; a change of case alone
	// sent nothing.
	if m := relay.Wait(t, 2); m[1].To[0] != "Olivier@example.org" {
		t.Errorf("the notice went to %v", m[1].To)
	}
}
