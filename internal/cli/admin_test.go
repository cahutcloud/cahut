package cli

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

// run runs the command line with a data directory and returns its exit
// code and outputs.
func run(t *testing.T, data string, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(append(args, "--data", data), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestAdminCommands(t *testing.T) {
	data := t.TempDir()

	code, _, errOut := run(t, data, "user", "create", "--handle", "olivier", "--name", "Olivier Durand")
	if code != exitError || !strings.Contains(errOut, "no Cahut yet") {
		t.Errorf("user create without a Cahut: %d %q", code, errOut)
	}
	code, out, errOut := run(t, data, "domain", "add", "cahut.example.com")
	if code != exitOK || !strings.Contains(out, "Declared cahut.example.com") {
		t.Fatalf("domain add: %d %q %q", code, out, errOut)
	}

	code, out, errOut = run(t, data, "user", "create", "--handle", "olivier", "--name", "Olivier Durand", "--admin")
	if code != exitOK || !strings.Contains(out, "Created olivier (Olivier Durand), admin of cahut.example.com, with a personal Room at cahut.example.com/olivier.") {
		t.Fatalf("user create: %d %q %q", code, out, errOut)
	}
	code, _, errOut = run(t, data, "user", "create", "--handle", "olivier", "--name", "Again")
	if code != exitError || !strings.Contains(errOut, `invalid input: handle "olivier" is taken`) {
		t.Errorf("duplicate user create: %d %q", code, errOut)
	}

	code, out, errOut = run(t, data, "token", "create", "--user", "olivier", "--name", "Claude Code")
	if code != exitOK {
		t.Fatalf("token create: %d %q", code, errOut)
	}
	if !regexp.MustCompile(`(?m)^cahut_pat_[a-z2-7]{52}$`).MatchString(out) {
		t.Errorf("token create printed no token: %q", out)
	}

	code, out, _ = run(t, data, "token", "list")
	if code != exitOK || !regexp.MustCompile(`1\s+olivier\s+Claude Code\s+\S+\s+never\s+active`).MatchString(out) {
		t.Errorf("token list: %d %q", code, out)
	}
	code, out, _ = run(t, data, "token", "revoke", "--id", "1")
	if code != exitOK || !strings.Contains(out, "Revoked token 1 of olivier") {
		t.Errorf("token revoke: %d %q", code, out)
	}

	code, out, _ = run(t, data, "logbook")
	if code != exitOK {
		t.Fatalf("logbook: %d", code)
	}
	for _, want := range []string{
		"revoke_token", "create_token", "create_user",
		"failed: invalid input: handle \"olivier\" is taken",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("logbook lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "cahut_pat_") {
		t.Errorf("logbook shows a token:\n%s", out)
	}
}

func TestSeveralCahuts(t *testing.T) {
	data := t.TempDir()
	for _, domain := range []string{"a.example.com", "b.example.com"} {
		if code, _, errOut := run(t, data, "domain", "add", domain); code != exitOK {
			t.Fatalf("domain add %s: %d %q", domain, code, errOut)
		}
	}
	if code, _, errOut := run(t, data, "domain", "add", "a.example.com"); code != exitError || !strings.Contains(errOut, "already the domain of a Cahut") {
		t.Errorf("duplicate domain add: %d %q", code, errOut)
	}
	if code, _, errOut := run(t, data, "domain", "add"); code != exitUsage || !strings.Contains(errOut, "Usage: cahut domain add") {
		t.Errorf("domain add without a domain: %d %q", code, errOut)
	}
	code, out, _ := run(t, data, "domain", "list")
	if code != exitOK || !regexp.MustCompile(`(?s)1\s+a\.example\.com.*2\s+b\.example\.com`).MatchString(out) {
		t.Errorf("domain list: %d %q", code, out)
	}

	code, _, errOut := run(t, data, "user", "create", "--handle", "olivier", "--name", "Olivier")
	if code != exitError || !strings.Contains(errOut, "choose one with --domain") {
		t.Errorf("user create without --domain: %d %q", code, errOut)
	}
	if code, _, errOut := run(t, data, "user", "create", "--domain", "B.example.com", "--handle", "olivier", "--name", "Olivier"); code != exitOK {
		t.Errorf("user create in b: %d %q", code, errOut)
	}
	if code, _, errOut := run(t, data, "token", "create", "--domain", "a.example.com", "--user", "olivier", "--name", "x"); code != exitError || !strings.Contains(errOut, `no user "olivier"`) {
		t.Errorf("token in a for a person of b: %d %q", code, errOut)
	}
	if code, _, errOut := run(t, data, "token", "list", "--domain", "c.example.com"); code != exitError || !strings.Contains(errOut, "no Cahut at c.example.com") {
		t.Errorf("token list in an unknown Cahut: %d %q", code, errOut)
	}

	code, out, _ = run(t, data, "logbook", "--server")
	if code != exitOK || strings.Count(out, "add_domain") != 3 {
		t.Errorf("server logbook: %d\n%s", code, out)
	}
}

func TestGroupUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"token"}, &stdout, &stderr); code != exitUsage || !strings.Contains(stderr.String(), "revoke") {
		t.Errorf("token without a command: %d %q", code, stderr.String())
	}
	stderr.Reset()
	if code := Run([]string{"token", "burn"}, &stdout, &stderr); code != exitUsage || !strings.Contains(stderr.String(), `unknown command "burn"`) {
		t.Errorf("token burn: %d %q", code, stderr.String())
	}
	if code := Run([]string{"user", "help"}, &stdout, &stderr); code != exitOK || !strings.Contains(stdout.String(), "create") {
		t.Errorf("user help: %d %q", code, stdout.String())
	}
}
