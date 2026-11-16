// Package skill keeps the skills of a Cahut: instructions that its people
// write for some tasks, in the format of Agent Skills, which AIs read over
// MCP. See docs/adr/0064-the-skills-of-a-cahut.md.
package skill

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"

	"cahut.cloud/internal/operation"
)

// Limits of a skill. See docs/adr/0064-the-skills-of-a-cahut.md.
const (
	MaxText            = 64 << 10 // a SKILL.md
	MaxEveryWindowText = 16 << 10 // the SKILL.md of a skill for every Window, which a refusal of the gate carries
	MaxFile            = 2 << 20
	MaxSize            = 4 << 20 // all the files of a version, SKILL.md included
	MaxFiles           = 100
	MaxPerScope        = 50
	MaxEveryWindow     = 3
	// MaxDescription is the longest description, in characters, as Agent
	// Skills have it.
	MaxDescription = 1024
	// BuiltinName is the name of the skill of the plugin for Claude, which
	// no other skill takes.
	BuiltinName = "cahut"
)

// Scope is where a skill lives: in the Cahut, with a person, or in the
// binary.
type Scope string

// Scopes of skills.
const (
	ScopeCahut    Scope = "cahut"
	ScopePersonal Scope = "personal"
	ScopeBuiltin  Scope = "builtin"
)

// File is a file of a skill, other than its SKILL.md.
type File struct {
	Path    string
	Type    string // its media type, from its extension
	Content []byte
}

// Header is what Cahut reads of the front matter of a SKILL.md.
type Header struct{ Name, Description string }

var namePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// CheckName accepts the names of Agent Skills: 1 to 64 of a-z, 0-9 and -,
// without a hyphen at either end or two in a row.
func CheckName(name string) error {
	if len(name) > 64 || !namePattern.MatchString(name) {
		return errNameInvalid(BadName{Name: operation.Quoted(name)})
	}
	return nil
}

// ParseText reads the name and the description from the front matter of a
// SKILL.md, with go.yaml.in/yaml/v3. A value that holds a colon and a space
// unquoted, which strict YAML refuses and Claude accepts, is read as text.
func ParseText(text string) (Header, error) {
	t, err := split(text)
	if err != nil {
		return Header{}, err
	}
	return parseHead(t.head)
}

// RenderText returns the SKILL.md text with name and description in double
// quotes, and the rest of its header and its body as they are: Mistral Vibe
// drops a skill whose description holds a colon and a space unquoted.
func RenderText(text string) (string, error) {
	t, err := split(text)
	if err != nil {
		return "", err
	}
	h, err := parseHead(t.head)
	if err != nil {
		return "", err
	}
	lines := strings.Split(t.head, "\n")
	out := make([]string, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		key := topKey(lines[i])
		if key != "name" && key != "description" {
			out = append(out, lines[i])
			continue
		}
		end := valueEnd(lines, i)
		value := h.Name
		if key == "description" {
			value = h.Description
		}
		cr := ""
		if strings.HasSuffix(lines[i], "\r") {
			cr = "\r"
		}
		out = append(out, key+": "+quote(value)+cr)
		i = end - 1
	}
	return t.open + strings.Join(out, "\n") + t.close + t.body, nil
}

// text is a SKILL.md cut at its header: the line that opens it, the header,
// the line that closes it, and the body.
type text struct{ open, head, close, body string }

func split(s string) (text, error) {
	var t text
	switch {
	case strings.HasPrefix(s, "---\n"):
		t.open = "---\n"
	case strings.HasPrefix(s, "---\r\n"):
		t.open = "---\r\n"
	default:
		return text{}, errHeaderInvalid(HeaderProblem{Problem: "is missing: the text does not open with a line ---"})
	}
	rest := s[len(t.open):]
	for pos := 0; ; {
		line, after, found := strings.Cut(rest[pos:], "\n")
		if strings.TrimSuffix(line, "\r") == "---" {
			t.head, t.close, t.body = rest[:pos], line, after
			if found {
				t.close += "\n"
			}
			return t, nil
		}
		if !found {
			return text{}, errHeaderInvalid(HeaderProblem{Problem: "never ends: no line --- closes it"})
		}
		pos += len(line) + 1
	}
}

func parseHead(head string) (Header, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(head), &doc); err != nil {
		lenient, changed := quotePlain(head)
		if !changed || yaml.Unmarshal([]byte(lenient), &doc) != nil {
			return Header{}, errHeaderInvalid(HeaderProblem{Problem: "is not valid YAML"})
		}
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return Header{}, errHeaderInvalid(HeaderProblem{Problem: "is not a list of keys and values"})
	}
	m := doc.Content[0]
	var h Header
	seen := map[string]bool{}
	for i := 0; i+1 < len(m.Content); i += 2 {
		k, v := m.Content[i], m.Content[i+1]
		if k.Kind != yaml.ScalarNode {
			return Header{}, errHeaderInvalid(HeaderProblem{Problem: "has a key that is not a word"})
		}
		if seen[k.Value] {
			return Header{}, errHeaderInvalid(HeaderProblem{Problem: fmt.Sprintf("gives %s twice", k.Value)})
		}
		seen[k.Value] = true
		if k.Value != "name" && k.Value != "description" {
			continue
		}
		if v.Kind != yaml.ScalarNode || v.Tag == "!!null" {
			return Header{}, errHeaderInvalid(HeaderProblem{Problem: fmt.Sprintf("gives %s as something other than text", k.Value)})
		}
		if k.Value == "name" {
			h.Name = v.Value
		} else {
			h.Description = v.Value
		}
	}
	switch {
	case !seen["name"]:
		return Header{}, errHeaderInvalid(HeaderProblem{Problem: "has no name"})
	case !seen["description"] || strings.TrimSpace(h.Description) == "":
		return Header{}, errHeaderInvalid(HeaderProblem{Problem: "has no description"})
	case utf8.RuneCountInString(h.Description) > MaxDescription:
		return Header{}, errHeaderInvalid(HeaderProblem{Problem: fmt.Sprintf("has a description longer than %d characters", MaxDescription)})
	}
	if err := CheckName(h.Name); err != nil {
		return Header{}, err
	}
	// RenderText quotes name and description where they stand: each must
	// open a line of its own.
	var keys []string
	for _, line := range strings.Split(head, "\n") {
		if k := topKey(line); k == "name" || k == "description" {
			keys = append(keys, k)
		}
	}
	if slices.Sort(keys); !slices.Equal(keys, []string{"description", "name"}) {
		return Header{}, errHeaderInvalid(HeaderProblem{Problem: "does not give name and description each on a line of its own"})
	}
	return h, nil
}

var keyLine = regexp.MustCompile(`^([A-Za-z0-9_][A-Za-z0-9_.-]*)[ \t]*:([ \t]|$)`)

// topKey returns the key that a line of a header opens at its first
// column, or "".
func topKey(line string) string {
	if m := keyLine.FindStringSubmatch(strings.TrimSuffix(line, "\r")); m != nil {
		return m[1]
	}
	return ""
}

// valueEnd returns the index of the first line after the value of the key
// on line i: the lines that follow it indented belong to it, and so do the
// blank lines between them.
func valueEnd(lines []string, i int) int {
	end := i + 1
	for j := i + 1; j < len(lines); j++ {
		line := strings.TrimSuffix(lines[j], "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if line[0] != ' ' && line[0] != '\t' {
			break
		}
		end = j + 1
	}
	return end
}

var plainValue = regexp.MustCompile(`^([A-Za-z0-9_][A-Za-z0-9_.-]*)[ \t]*:[ \t]+(.*)$`)

// quotePlain quotes the plain values of the top keys of head that hold a
// colon and a space, as Claude reads them, with the indented lines that
// continue them. It reports whether it changed anything.
func quotePlain(head string) (string, bool) {
	lines := strings.Split(head, "\n")
	out := make([]string, 0, len(lines))
	changed := false
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSuffix(lines[i], "\r")
		m := plainValue.FindStringSubmatch(line)
		if m == nil || m[2] == "" || strings.ContainsRune(`"'|>[{&*!%@`+"`", rune(m[2][0])) ||
			!strings.Contains(m[2], ": ") && !strings.Contains(m[2], ":\t") {
			out = append(out, lines[i])
			continue
		}
		end := valueEnd(lines, i)
		parts := []string{strings.TrimSpace(m[2])}
		for _, l := range lines[i+1 : end] {
			if l = strings.TrimSpace(l); l != "" {
				parts = append(parts, l)
			}
		}
		out = append(out, m[1]+": "+quote(strings.Join(parts, " ")))
		changed = true
		i = end - 1
	}
	return strings.Join(out, "\n"), changed
}

// quote writes s as a double-quoted scalar: a string of JSON, which YAML
// reads.
func quote(s string) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s) // a string always encodes
	return strings.TrimSuffix(b.String(), "\n")
}

// CheckPath accepts the path of a file of a skill: clean, relative, with /
// as separator and no hidden part, neither SKILL.md nor under scripts/.
func CheckPath(p string) error {
	switch {
	case p == "":
		return operation.Invalidf("a file has an empty path")
	case len(p) > 255:
		return operation.Invalidf("path %.40q... is longer than 255 bytes", p)
	case !utf8.ValidString(p) || strings.ContainsFunc(p, unicode.IsControl) || strings.Contains(p, `\`):
		return operation.Invalidf("path %q: use printable characters and / as separator", p)
	case path.IsAbs(p) || path.Clean(p) != p:
		return operation.Invalidf("path %q: use a clean relative path, such as references/deck.md", p)
	}
	for _, part := range strings.Split(p, "/") {
		if strings.HasPrefix(part, ".") {
			return operation.Invalidf("path %q: no part may start with a dot", p)
		}
	}
	lower := strings.ToLower(p)
	switch {
	case lower == "scripts" || strings.HasPrefix(lower, "scripts/"):
		return errHasScripts(ScriptPath{Path: operation.Quoted(p)})
	case lower == "skill.md":
		return operation.Invalidf("SKILL.md is the text of the skill, given apart from its files")
	}
	return nil
}

// Digest returns the first 8 hexadecimal digits of the SHA-256 of the text
// of a skill and of its files, by path.
func Digest(text string, files []File) string {
	sorted := slices.SortedFunc(slices.Values(files), func(a, b File) int { return cmp.Compare(a.Path, b.Path) })
	h := sha256.New()
	write := func(b []byte) {
		_ = binary.Write(h, binary.BigEndian, uint64(len(b))) // a hash never fails to write
		h.Write(b)
	}
	write([]byte(text))
	for _, f := range sorted {
		write([]byte(f.Path))
		write(f.Content)
	}
	return hex.EncodeToString(h.Sum(nil))[:8]
}
