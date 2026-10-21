package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string // substring expected on stdout, empty for none
		wantStderr string // substring expected on stderr, empty for none
	}{
		{name: "no command", args: nil, wantCode: exitUsage, wantStderr: "Usage:"},
		{name: "help", args: []string{"help"}, wantCode: exitOK, wantStdout: "Commands:"},
		{name: "help flag", args: []string{"--help"}, wantCode: exitOK, wantStdout: "Commands:"},
		{name: "unknown command", args: []string{"frobnicate"}, wantCode: exitUsage, wantStderr: `unknown command "frobnicate"`},
		{name: "version", args: []string{"version"}, wantCode: exitOK, wantStdout: "cahut "},
		{name: "version flag", args: []string{"--version"}, wantCode: exitOK, wantStdout: "cahut "},
		{name: "version help", args: []string{"version", "-h"}, wantCode: exitOK, wantStderr: "Usage of cahut version"},
		{name: "version bad flag", args: []string{"version", "-x"}, wantCode: exitUsage, wantStderr: "flag provided but not defined"},
		{name: "version extra argument", args: []string{"version", "now"}, wantCode: exitUsage, wantStderr: "unexpected arguments"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(tt.args, &stdout, &stderr)
			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			check(t, "stdout", stdout.String(), tt.wantStdout)
			check(t, "stderr", stderr.String(), tt.wantStderr)
		})
	}
}

func check(t *testing.T, stream, got, want string) {
	t.Helper()
	if want == "" {
		if got != "" {
			t.Errorf("%s = %q, want nothing", stream, got)
		}
		return
	}
	if !strings.Contains(got, want) {
		t.Errorf("%s = %q, want it to contain %q", stream, got, want)
	}
}
