package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	for _, name := range environment {
		t.Setenv(name, "")
	}
	dir, again := t.TempDir(), t.TempDir()
	stale := filepath.Join(dir, "removed.md")
	if err := os.WriteFile(stale, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run(t.Context(), dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("stat %s: %v, want it removed", stale, err)
	}

	want := map[string][]string{
		"mcp-tools.md":  {"## `publish_window`", "| `files[].path` | string | yes |", "Every call is recorded in the Logbook"},
		"cli.md":        {"## `cahut serve`", "### `cahut token create`", "| `--data string` | `/var/lib/cahut` |"},
		"operations.md": {"[`publish_window`](/reference/mcp-tools/#publish_window)", "| `add_domain` |"},
		"http-api.md": {"## Sessions", "| `POST /server/api/add_domain` | Administrators of the server | yes |", "| 401 Unauthorized |",
			"| `POST /server/api/cahuts/<domain>/suspend_cahut` |", "| `POST /console/api/list_rooms` | Any person | yes |",
			"| `password_too_short` | `length`, `limit` | The password has {length} characters: use at least {limit} |", "| `session_ended` |  | Invalid or ended session |"},
	}
	for file, parts := range want {
		got, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.HasPrefix(got, []byte("---\ntitle: ")) {
			t.Errorf("%s starts with %.40q, want front matter", file, got)
		}
		for _, part := range parts {
			if !bytes.Contains(got, []byte(part)) {
				t.Errorf("%s lacks %q", file, part)
			}
		}
	}

	// The same code gives the same pages.
	if err := run(t.Context(), again); err != nil {
		t.Fatal(err)
	}
	for file := range want {
		a, _ := os.ReadFile(filepath.Join(dir, file))
		b, _ := os.ReadFile(filepath.Join(again, file))
		if !bytes.Equal(a, b) {
			t.Errorf("%s differs from one run to the next", file)
		}
	}
}

func TestParseFlags(t *testing.T) {
	help := strings.Join([]string{
		"Usage of cahut logbook:",
		"  -data string",
		"    \tdirectory of the data (env CAHUT_DATA) (default \"/var/lib/cahut\")",
		"  -limit int",
		"    \thow many entries (default 20)",
		"  -server",
		"    \tprint the Logbook of the server,",
		"    \tnot of a Cahut",
		"  -x\ta short flag",
		"",
	}, "\n")
	synopsis, flags, err := parseFlags("cahut logbook", help)
	if err != nil {
		t.Fatal(err)
	}
	if synopsis != "cahut logbook [flags]" {
		t.Errorf("synopsis = %q", synopsis)
	}
	want := []flagHelp{
		{name: "--data", value: "string", def: "/var/lib/cahut", usage: "directory of the data (env CAHUT_DATA)"},
		{name: "--limit", value: "int", def: "20", usage: "how many entries"},
		{name: "--server", usage: "print the Logbook of the server, not of a Cahut"},
		{name: "--x", usage: "a short flag"},
	}
	if !reflect.DeepEqual(flags, want) {
		t.Errorf("flags = %+v\nwant %+v", flags, want)
	}

	if _, _, err := parseFlags("cahut logbook", "something else\n"); err == nil {
		t.Error("parseFlags accepted a help it does not know")
	}
}

func TestTypeName(t *testing.T) {
	tests := map[string]string{
		`{"type": "string"}`:                                        "string",
		`{"type": ["null", "string"]}`:                              "string or null",
		`{"type": "array", "items": {"type": "object"}}`:            "array of objects",
		`{"type": ["null", "array"], "items": {"type": "integer"}}`: "array of integers or null",
		`{}`: "any",
	}
	for in, want := range tests {
		var s schema
		if err := json.Unmarshal([]byte(in), &s); err != nil {
			t.Fatal(err)
		}
		if got := s.typeName(); got != want {
			t.Errorf("typeName(%s) = %q, want %q", in, got, want)
		}
	}
}

func TestText(t *testing.T) {
	tests := map[string]string{
		"serve plain HTTP on --dev-addr, for *.localhost": "serve plain HTTP on `--dev-addr`, for \\*.localhost",
		"address for --dev": "address for `--dev`",
		"a | b <c>":         "a \\| b &lt;c&gt;",
		"(see --data)":      "(see `--data`)",
	}
	for in, want := range tests {
		if got := text(in); got != want {
			t.Errorf("text(%q) = %q, want %q", in, got, want)
		}
	}
}
