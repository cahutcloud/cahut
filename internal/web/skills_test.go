package web

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	"cahut.cloud/internal/skill"
)

// The console imports a skill of 4 MiB, which takes 6 MiB in base64, and
// every other call stays under 1 MiB. See
// docs/adr/0064-the-skills-of-a-cahut.md.
func TestImportSkillTakesALargeCall(t *testing.T) {
	f := setup(t)
	console := f.ts
	ada := activate(t, f, console, "ada", true)

	// 3 MiB of bytes that do not compress, in two files.
	files := []skill.File{{Path: "assets/a.bin", Content: make([]byte, 3<<19)}, {Path: "assets/b.bin", Content: make([]byte, 3<<19)}}
	for _, f := range files {
		_, _ = rand.Read(f.Content) // crypto/rand never fails
	}
	text := "---\nname: charte-acme\ndescription: The brand of Acme.\n---\n# Acme\n"
	zip, err := skill.WriteZip("charte-acme", text, files)
	if err != nil {
		t.Fatal(err)
	}
	body := `{"scope": "cahut", "zip": "` + base64.StdEncoding.EncodeToString(zip) + `"}`
	if len(body) < 4<<20 {
		t.Fatalf("the call holds %d bytes, want more than 4 MiB", len(body))
	}
	r := call(t, f, console, http.MethodPost, "localhost", "/console/api/import_skill", ada, body)
	if r.status != http.StatusOK || r.body["skill"] != "charte-acme" {
		t.Fatalf("import_skill of %d bytes: %d %.200s", len(body), r.status, r.raw)
	}

	for _, c := range []struct {
		path, body string
	}{
		{"/console/api/write_skill", `{"scope": "cahut", "skill": "x", "text": "` + strings.Repeat("x", 2<<20) + `"}`},
		{"/console/api/import_skill", `{"scope": "cahut", "zip": "` + strings.Repeat("A", 6<<20) + `"}`},
	} {
		if r := call(t, f, console, http.MethodPost, "localhost", c.path, ada, c.body); r.status != http.StatusRequestEntityTooLarge {
			t.Errorf("%s with %d bytes: %d, want 413", c.path, len(c.body), r.status)
		}
	}
}
