package cli

import (
	"regexp"
	"strings"
	"testing"
)

func TestSettingsCommands(t *testing.T) {
	data := t.TempDir()
	if code, _, errOut := run(t, data, "domain", "add", "cahut.example.com"); code != exitOK {
		t.Fatalf("domain add: %d %q", code, errOut)
	}
	if code, _, errOut := run(t, data, "user", "create", "--handle", "olivier", "--name", "Olivier", "--admin"); code != exitOK {
		t.Fatalf("user create: %d %q", code, errOut)
	}

	code, out, errOut := run(t, data, "settings", "show")
	if code != exitOK || !strings.Contains(out, "Settings of cahut.example.com:") ||
		!regexp.MustCompile(`people\s+1\s+none`).MatchString(out) || !regexp.MustCompile(`disk\s+0 bytes\s+none`).MatchString(out) {
		t.Errorf("settings show: %d %q %q", code, out, errOut)
	}

	code, out, errOut = run(t, data, "settings", "set", "--people", "20", "--disk", "10GB")
	if code != exitOK || !regexp.MustCompile(`people\s+1\s+20`).MatchString(out) || !regexp.MustCompile(`disk\s+0 bytes\s+10 GB`).MatchString(out) {
		t.Errorf("settings set: %d %q %q", code, out, errOut)
	}
	// The limits left out stay; none lifts one.
	code, out, _ = run(t, data, "settings", "set", "--rooms", "3", "--disk", "none")
	if code != exitOK || !regexp.MustCompile(`people\s+1\s+20`).MatchString(out) ||
		!regexp.MustCompile(`team Rooms\s+0\s+3`).MatchString(out) || !regexp.MustCompile(`disk\s+0 bytes\s+none`).MatchString(out) {
		t.Errorf("settings set again: %d %q", code, out)
	}

	// The kind, and the limits per person that go with an individual Cahut.
	code, out, errOut = run(t, data, "settings", "set", "--kind", "individual", "--windows-per-person", "50", "--disk-per-person", "1GB")
	if code != exitOK || !strings.Contains(out, "Kind: individual") || !regexp.MustCompile(`Windows per person\s+50`).MatchString(out) ||
		!regexp.MustCompile(`disk per person\s+1 GB`).MatchString(out) || !regexp.MustCompile(`people\s+1\s+20`).MatchString(out) {
		t.Errorf("settings set --kind individual: %d %q %q", code, out, errOut)
	}
	code, out, errOut = run(t, data, "settings", "set", "--kind", "collaborative")
	if code != exitOK || !strings.Contains(out, "Kind: collaborative") || strings.Contains(out, "per person") {
		t.Errorf("settings set --kind collaborative: %d %q %q", code, out, errOut)
	}
	if code, _, errOut := run(t, data, "settings", "set", "--windows-per-person", "5"); code != exitError || !strings.Contains(errOut, "make the Cahut individual") {
		t.Errorf("settings set --windows-per-person in a collaborative Cahut: %d %q", code, errOut)
	}

	if code, _, errOut := run(t, data, "settings", "set", "--people", "0"); code != exitError || !strings.Contains(errOut, "people: give a limit of 1 or more") {
		t.Errorf("settings set --people 0: %d %q", code, errOut)
	}
	if code, _, errOut := run(t, data, "settings", "set", "--windows", "many"); code != exitUsage || !strings.Contains(errOut, "give a number, or none") {
		t.Errorf("settings set --windows many: %d %q", code, errOut)
	}
}

func TestParseSize(t *testing.T) {
	for s, want := range map[string]int64{"10GB": 10e9, "1.5 GB": 1.5e9, "500mb": 500e6, "2T": 2e12, "750k": 750e3, "1234": 1234, "0": 0, "12B": 12} {
		if got, err := parseSize(s); err != nil || got != want {
			t.Errorf("parseSize(%q) = %d, %v; want %d", s, got, err, want)
		}
	}
	for _, s := range []string{"", "GB", "-1GB", "ten", "1e30GB", "NaN"} {
		if _, err := parseSize(s); err == nil {
			t.Errorf("parseSize(%q) succeeded", s)
		}
	}
}

func TestSettingsLinks(t *testing.T) {
	data := t.TempDir()
	if code, _, errOut := run(t, data, "domain", "add", "cahut.example.com"); code != exitOK {
		t.Fatalf("domain add: %d %q", code, errOut)
	}

	code, out, errOut := run(t, data, "settings", "links")
	if code != exitOK || !strings.Contains(out, "Legal pages of cahut.example.com:") || !regexp.MustCompile(`terms\s+none\n`).MatchString(out) {
		t.Errorf("settings links: %d %q %q", code, out, errOut)
	}

	code, out, errOut = run(t, data, "settings", "links", "--terms", "https://host.example/terms", "--terms-fr", "https://host.example/fr/conditions",
		"--privacy", "https://host.example/privacy")
	if code != exitOK || !strings.Contains(out, "Set the legal pages of cahut.example.com:") ||
		!regexp.MustCompile(`terms\s+https://host.example/terms\s+https://host.example/fr/conditions\n`).MatchString(out) ||
		!regexp.MustCompile(`privacy\s+https://host.example/privacy\s*\n`).MatchString(out) {
		t.Errorf("settings links --terms: %d %q %q", code, out, errOut)
	}
	// The links left out stay; none removes one, with its address in French.
	code, out, _ = run(t, data, "settings", "links", "--support", "https://help.host.example/", "--terms", "none")
	if code != exitOK || !regexp.MustCompile(`terms\s+none\s*\n`).MatchString(out) || strings.Contains(out, "conditions") ||
		!regexp.MustCompile(`privacy\s+https://host.example/privacy`).MatchString(out) || !regexp.MustCompile(`support\s+https://help.host.example/`).MatchString(out) {
		t.Errorf("settings links again: %d %q", code, out)
	}
	if code, out, _ := run(t, data, "settings", "show"); code != exitOK || !strings.Contains(out, "Legal pages:") ||
		!regexp.MustCompile(`privacy\s+https://host.example/privacy`).MatchString(out) {
		t.Errorf("settings show: %d %q", code, out)
	}

	if code, _, errOut := run(t, data, "settings", "links", "--privacy", "http://host.example/privacy"); code != exitError ||
		!strings.Contains(errOut, "privacy: give the address of a page, which starts with https://") {
		t.Errorf("settings links --privacy http://: %d %q", code, errOut)
	}
	if code, _, errOut := run(t, data, "settings", "links", "--legal-notice-fr", "https://host.example/fr/mentions-legales"); code != exitError ||
		!strings.Contains(errOut, "fr.legal_notice: give legal_notice too") {
		t.Errorf("settings links --legal-notice-fr alone: %d %q", code, errOut)
	}
}
