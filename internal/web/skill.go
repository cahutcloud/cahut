package web

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"cahut.cloud/plugins"
)

// SkillPath is where each Cahut serves the skill for Claude, as a zip,
// for the Claude apps where skills are uploaded.
const SkillPath = "/-/cahut-skill.zip"

// serveSkill serves the skill for Claude, the same as its plugin's. See
// docs/adr/0022-house-style-of-each-cahut.md.
func (s *Server) serveSkill(w http.ResponseWriter, r *http.Request) {
	b, err := plugins.SkillZip()
	if err != nil {
		s.internal(w, "make the zip of the skill", err)
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="cahut-skill.zip"`)
	serveFile(w, r, newServedFile("application/zip", b))
}

// servedFile is a file that Cahut keeps in memory and serves under /-/.
type servedFile struct {
	Type    string
	ETag    string // quoted, as the header carries it
	Content []byte
}

func newServedFile(typ string, content []byte) servedFile {
	sum := sha256.Sum256(content)
	return servedFile{Type: typ, ETag: `"` + hex.EncodeToString(sum[:16]) + `"`, Content: content}
}

// serveFile serves f to anyone. Browsers revalidate it each time, so that a
// new release reaches every reader at once.
func serveFile(w http.ResponseWriter, r *http.Request, f servedFile) {
	h := w.Header()
	h.Set("Content-Type", f.Type)
	h.Set("ETag", f.ETag)
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(f.Content))
}
