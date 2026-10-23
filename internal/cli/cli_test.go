package cli

import (
	"bytes"
	"strings"
	"testing"

	"cahut.cloud/internal/doctor"
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
		{name: "version extra argument", args: []string{"version", "now"}, wantCode: exitUsage, wantStderr: `unexpected argument "now"`},
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

func TestDoctorReport(t *testing.T) {
	for _, tt := range []struct {
		results []doctor.Result
		code    int
		summary string
	}{
		{[]doctor.Result{{Status: doctor.OK, Topic: "Clock", Text: "Synchronized."}}, exitOK, "No problem.\n"},
		{[]doctor.Result{{Status: doctor.Warn, Topic: "Data", Text: "3.0 GiB free."}}, exitOK, "No problem, 1 warning.\n"},
		{[]doctor.Result{
			{Status: doctor.Fail, Topic: "Clock", Text: "Not synchronized."},
			{Status: doctor.Fail, Topic: "Ports", Text: "Nothing listens."},
			{Status: doctor.Warn, Topic: "Data", Text: "3.0 GiB free."},
		}, exitError, "2 problems and 1 warning: fix what fails above, then run cahut doctor again.\n"},
	} {
		var out bytes.Buffer
		if code := report(&out, tt.results); code != tt.code || !strings.HasSuffix(out.String(), tt.summary) {
			t.Errorf("report = %d\n%s\nwant %d, ending with %q", code, out.String(), tt.code, tt.summary)
		}
	}
	var out bytes.Buffer
	report(&out, []doctor.Result{{Status: doctor.OK, Topic: "Clock", Text: "Synchronized."}, {Status: doctor.Fail, Topic: "Let's Encrypt", Text: "Cannot be reached."}})
	if want := "ok    Clock          Synchronized.\nfail  Let's Encrypt  Cannot be reached.\n"; !strings.HasPrefix(out.String(), want) {
		t.Errorf("report printed\n%s\nwant it to start with\n%s", out.String(), want)
	}
}
