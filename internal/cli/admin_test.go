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

	code, out, errOut := run(t, data, "user", "create", "--handle", "olivier", "--name", "Olivier Durand", "--admin")
	if code != exitOK || !strings.Contains(out, "Created olivier (Olivier Durand), admin") {
		t.Fatalf("user create: %d %q %q", code, out, errOut)
	}
	code, _, errOut = run(t, data, "user", "create", "--handle", "olivier", "--name", "Again")
	if code != exitError || !strings.Contains(errOut, `create_user: invalid input: handle "olivier" is taken`) {
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
