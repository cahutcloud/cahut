package skill

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"cahut.cloud/internal/operation"
)

const acmeText = "---\nname: charte-acme\ndescription: The brand of Acme: for every deck.\n---\n# Acme\n"

type entry struct {
	name    string
	content []byte
	mode    fs.FileMode
}

func makeZip(t *testing.T, entries ...entry) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for _, e := range entries {
		h := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		if e.mode != 0 {
			h.SetMode(e.mode)
		}
		f, err := w.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(e.content); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestReadZipAtRoot(t *testing.T) {
	data := makeZip(t,
		entry{name: "SKILL.md", content: []byte(acmeText)},
		entry{name: "references/"},
		entry{name: "references/deck.md", content: []byte("# Deck\n")},
		entry{name: "assets/logo.svg", content: []byte("<svg/>")},
	)
	text, files, err := ReadZip(data)
	if err != nil {
		t.Fatal(err)
	}
	if text != acmeText {
		t.Errorf("text = %q", text)
	}
	if got := fmt.Sprint(files); got != fmt.Sprint([]File{
		{Path: "assets/logo.svg", Type: "image/svg+xml", Content: []byte("<svg/>")},
		{Path: "references/deck.md", Type: "text/markdown; charset=utf-8", Content: []byte("# Deck\n")},
	}) {
		t.Errorf("files = %v", got)
	}
}

func TestReadZipInAFolder(t *testing.T) {
	// As a client exports a skill, and as the Finder of macOS zips a folder.
	data := makeZip(t,
		entry{name: "charte-acme/"},
		entry{name: "charte-acme/SKILL.md", content: []byte(acmeText)},
		entry{name: "charte-acme/assets/fonts/inter.woff2", content: []byte{0, 1, 2}},
		entry{name: "charte-acme/.DS_Store", content: []byte("finder")},
		entry{name: "__MACOSX/charte-acme/._SKILL.md", content: []byte("finder")},
	)
	text, files, err := ReadZip(data)
	if err != nil {
		t.Fatal(err)
	}
	if text != acmeText || len(files) != 1 || files[0].Path != "assets/fonts/inter.woff2" || !bytes.Equal(files[0].Content, []byte{0, 1, 2}) {
		t.Errorf("text = %q, files = %v", text, files)
	}
}

func TestReadZipRefuses(t *testing.T) {
	skill := entry{name: "SKILL.md", content: []byte(acmeText)}
	many := []entry{skill}
	for i := range MaxFiles + 1 {
		many = append(many, entry{name: fmt.Sprintf("references/%d.md", i), content: []byte("x")})
	}
	for name, c := range map[string]struct {
		entries []entry
		code    string
	}{
		"two folders":       {[]entry{{name: "a/SKILL.md", content: []byte(acmeText)}, {name: "b/x.md", content: []byte("x")}}, "skill_zip_invalid"},
		"a parent":          {[]entry{skill, {name: "../x.md", content: []byte("x")}}, "skill_zip_invalid"},
		"a parent inside":   {[]entry{skill, {name: "references/../../x.md", content: []byte("x")}}, "skill_zip_invalid"},
		"an absolute path":  {[]entry{skill, {name: "/etc/x.md", content: []byte("x")}}, "skill_zip_invalid"},
		"a backslash":       {[]entry{skill, {name: `references\x.md`, content: []byte("x")}}, "skill_zip_invalid"},
		"a link":            {[]entry{skill, {name: "references/x.md", content: []byte("/etc/passwd"), mode: fs.ModeSymlink | 0o777}}, "skill_zip_invalid"},
		"a duplicate":       {[]entry{skill, {name: "references/x.md", content: []byte("x")}, {name: "references/x.md", content: []byte("y")}}, "skill_zip_invalid"},
		"a duplicate case":  {[]entry{skill, {name: "references/x.md", content: []byte("x")}, {name: "references/X.md", content: []byte("y")}}, "skill_zip_invalid"},
		"a file and folder": {[]entry{skill, {name: "references", content: []byte("x")}, {name: "references/x.md", content: []byte("y")}}, "skill_zip_invalid"},
		"two SKILL.md":      {[]entry{skill, skill}, "skill_zip_invalid"},
		"scripts":           {[]entry{skill, {name: "scripts/run.sh", content: []byte("rm -rf ~")}}, "skill_has_scripts"},
		"no SKILL.md":       {[]entry{{name: "references/x.md", content: []byte("x")}}, "skill_zip_invalid"},
		"an invalid header": {[]entry{{name: "SKILL.md", content: []byte("# No header\n")}}, "skill_header_invalid"},
		"101 files":         {many, "skill_too_large"},
		"a large file":      {[]entry{skill, {name: "assets/big.bin", content: make([]byte, MaxFile+1)}}, "skill_too_large"},
		"a large SKILL.md":  {[]entry{{name: "SKILL.md", content: []byte(acmeText + strings.Repeat("x", MaxText))}}, "skill_too_large"},
		"too much in all": {[]entry{skill,
			{name: "assets/a.bin", content: make([]byte, MaxFile)}, {name: "assets/b.bin", content: make([]byte, MaxFile)}}, "skill_too_large"},
		// A zip bomb: 64 MiB of zeros, a few kilobytes compressed.
		"a bomb": {[]entry{skill, {name: "assets/zeros.bin", content: make([]byte, 64<<20)}}, "skill_too_large"},
	} {
		_, _, err := ReadZip(makeZip(t, c.entries...))
		var r *operation.Refusal
		if !errors.As(err, &r) || r.Code != c.code || !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("%s: err = %v, want the refusal %s", name, err, c.code)
		}
	}
	if _, _, err := ReadZip([]byte("not a zip")); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("not a zip: err = %v, want ErrInvalid", err)
	}
}

func TestWriteZipReadsBack(t *testing.T) {
	files := []File{
		{Path: "references/deck.md", Content: []byte("# Deck\n")},
		{Path: "assets/fonts/inter.woff2", Content: []byte{0, 0xff, 1}},
	}
	data, err := WriteZip("charte-acme", acmeText, files)
	if err != nil {
		t.Fatal(err)
	}
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	if got, want := strings.Join(names, " "), "charte-acme/SKILL.md charte-acme/assets/fonts/inter.woff2 charte-acme/references/deck.md"; got != want {
		t.Errorf("entries = %s, want %s", got, want)
	}
	text, back, err := ReadZip(data)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := RenderText(acmeText)
	if err != nil {
		t.Fatal(err)
	}
	if text != rendered || !strings.Contains(text, `description: "The brand of Acme: for every deck."`) {
		t.Errorf("SKILL.md = %q, want %q", text, rendered)
	}
	if len(back) != 2 || back[0].Path != "assets/fonts/inter.woff2" || !bytes.Equal(back[0].Content, files[1].Content) ||
		back[1].Path != "references/deck.md" || !bytes.Equal(back[1].Content, files[0].Content) {
		t.Errorf("files = %v", back)
	}
}
