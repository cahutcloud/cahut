package skill

import (
	"errors"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	"cahut.cloud/internal/operation"
)

func TestCheckName(t *testing.T) {
	for name, ok := range map[string]bool{
		"charte-acme":           true,
		"a":                     true,
		"relance-facture-2026":  true,
		strings.Repeat("a", 64): true,
		strings.Repeat("a", 65): false,
		"":                      false,
		"-a":                    false,
		"a-":                    false,
		"a--b":                  false,
		"A":                     false,
		"é":                     false,
		"a_b":                   false,
		"a b":                   false,
	} {
		err := CheckName(name)
		if ok && err != nil {
			t.Errorf("CheckName(%q) = %v, want nil", name, err)
		}
		if !ok && !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("CheckName(%q) = %v, want ErrInvalid", name, err)
		}
	}
}

func TestParseText(t *testing.T) {
	for _, c := range []struct {
		text string
		want Header
	}{
		{"---\nname: charte-acme\ndescription: The brand of Acme, for every deck.\n---\n# Acme\n",
			Header{"charte-acme", "The brand of Acme, for every deck."}},
		// A colon and a space, unquoted, as Claude accepts it.
		{"---\nname: relance-facture\ndescription: Use it for: an e-mail that reminds a client of an invoice\n---\nBody\n",
			Header{"relance-facture", "Use it for: an e-mail that reminds a client of an invoice"}},
		{"---\r\nname: crlf\r\ndescription: Lines end with CR LF.\r\n---\r\nBody\r\n",
			Header{"crlf", "Lines end with CR LF."}},
		{"---\nname: literal\ndescription: |\n  Two lines,\n  kept.\nlicense: MIT\n---\n",
			Header{"literal", "Two lines,\nkept.\n"}},
		{"---\nname: folded\ndescription: >\n  One line,\n  folded.\n---\n",
			Header{"folded", "One line, folded.\n"}},
		{"---\nname: with-metadata\ndescription: \"Quoted: already.\"\nmetadata:\n  author: acme\n  version: \"1.0\"\n---\n",
			Header{"with-metadata", "Quoted: already."}},
	} {
		got, err := ParseText(c.text)
		if err != nil || got != c.want {
			t.Errorf("ParseText(%q) = %+v, %v; want %+v", c.text, got, err, c.want)
		}
	}

	for _, text := range []string{
		"# No header\n",
		"---\nname: no-end\ndescription: The header never ends.\n",
		"---\ndescription: No name.\n---\n",
		"---\nname: no-description\n---\n",
		"---\nname: empty\ndescription: \"\"\n---\n",
		"---\nname: long\ndescription: " + strings.Repeat("a", 1025) + "\n---\n",
		"---\nname: Not_A_Name\ndescription: The name is not one.\n---\n",
		"---\nname: [a, b]\ndescription: The name is a list.\n---\n",
		"---\nname: twice\nname: twice\ndescription: Two names.\n---\n",
		"---\n- a list\n---\n",
	} {
		if h, err := ParseText(text); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("ParseText(%q) = %+v, %v; want ErrInvalid", text, h, err)
		}
	}
	// A description of 1,024 characters, accents counted as one.
	if _, err := ParseText("---\nname: long\ndescription: " + strings.Repeat("é", 1024) + "\n---\n"); err != nil {
		t.Errorf("a description of 1,024 characters: %v", err)
	}
}

func TestRenderTextQuotes(t *testing.T) {
	text := "---\n" +
		"# The brand of Acme\n" +
		"name: charte-acme\n" +
		"description: Use it for: every deck,\n" +
		"  and every report.\n" +
		"license: Proprietary. See LICENSE.txt\n" +
		"metadata:\n" +
		"  author: acme\n" +
		"---\n" +
		"# Acme\n\nBody: kept, *byte for byte*.\n"
	got, err := RenderText(text)
	if err != nil {
		t.Fatal(err)
	}
	head, body, ok := strings.Cut(strings.TrimPrefix(got, "---\n"), "\n---\n")
	if !ok {
		t.Fatalf("rendered = %q, without its header", got)
	}
	if body != "# Acme\n\nBody: kept, *byte for byte*.\n" {
		t.Errorf("body = %q", body)
	}
	var m map[string]any
	if err := yaml.Unmarshal([]byte(head), &m); err != nil {
		t.Fatalf("the rendered header does not read as YAML: %v\n%s", err, head)
	}
	if m["name"] != "charte-acme" || m["description"] != "Use it for: every deck, and every report." {
		t.Errorf("name, description = %q, %q", m["name"], m["description"])
	}
	for _, line := range []string{
		"# The brand of Acme",
		`name: "charte-acme"`,
		`description: "Use it for: every deck, and every report."`,
		"license: Proprietary. See LICENSE.txt",
		"metadata:\n  author: acme",
	} {
		if !strings.Contains(head, line) {
			t.Errorf("rendered header lacks %q:\n%s", line, head)
		}
	}
	if again, err := RenderText(got); err != nil || again != got {
		t.Errorf("rendering twice = %q, %v; want %q", again, err, got)
	}

	// A description with quotes, a backslash and a line of its own.
	got, err = RenderText("---\nname: odd\ndescription: |\n  Say \"hi\" \\ then\n  go.\n---\n")
	if err != nil {
		t.Fatal(err)
	}
	if want := "---\nname: \"odd\"\ndescription: \"Say \\\"hi\\\" \\\\ then\\ngo.\\n\"\n---\n"; got != want {
		t.Errorf("rendered = %q, want %q", got, want)
	}
	if h, err := ParseText(got); err != nil || h.Description != "Say \"hi\" \\ then\ngo.\n" {
		t.Errorf("read back = %+v, %v", h, err)
	}
}

func TestCheckPath(t *testing.T) {
	for _, p := range []string{"references/deck.md", "assets/fonts/inter.woff2", "LICENSE.txt", "template.html"} {
		if err := CheckPath(p); err != nil {
			t.Errorf("CheckPath(%q) = %v, want nil", p, err)
		}
	}
	for _, p := range []string{"", "../x", "/x", "a//b", "a/./b", "a\\b", ".hidden", "a/.git/config", "SKILL.md", "skill.md", strings.Repeat("a", 256)} {
		if err := CheckPath(p); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("CheckPath(%q) = %v, want ErrInvalid", p, err)
		}
	}
	for _, p := range []string{"scripts/run.sh", "scripts", "Scripts/x.py", "scripts/a/b.js"} {
		var r *operation.Refusal
		if err := CheckPath(p); !errors.As(err, &r) || r.Code != "skill_has_scripts" {
			t.Errorf("CheckPath(%q) = %v, want the refusal skill_has_scripts", p, err)
		}
	}
}

func TestDigest(t *testing.T) {
	text := "---\nname: a\ndescription: A.\n---\n"
	files := []File{
		{Path: "references/deck.md", Content: []byte("# Deck")},
		{Path: "assets/logo.svg", Content: []byte("<svg/>")},
	}
	d := Digest(text, files)
	if len(d) != 8 || strings.Trim(d, "0123456789abcdef") != "" {
		t.Fatalf("Digest = %q, want 8 hexadecimal digits", d)
	}
	if got := Digest(text, []File{files[1], files[0]}); got != d {
		t.Errorf("the same files in another order: %q, want %q", got, d)
	}
	for _, other := range []struct {
		text  string
		files []File
	}{
		{text + " ", files},
		{text, []File{files[0], {Path: "assets/logo.svg", Content: []byte("<svg />")}}},
		{text, []File{files[0], {Path: "assets/logo2.svg", Content: []byte("<svg/>")}}},
		{text, files[:1]},
		// Bytes that move from a path to a content do not give the same digest.
		{text, []File{files[0], {Path: "assets/logo.sv", Content: []byte("g<svg/>")}}},
	} {
		if got := Digest(other.text, other.files); got == d {
			t.Errorf("Digest(%q, %v) = %q, the same as the original", other.text, other.files, got)
		}
	}
}
