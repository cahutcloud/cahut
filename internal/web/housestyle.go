package web

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"cahut.cloud/internal/housestyle"
	"cahut.cloud/internal/window"
	"cahut.cloud/plugins"
)

// The files of the house style of a Cahut: its style sheet, its logos, its
// fonts and the deck engine, at stable addresses under /-/. They are public,
// and the pages of Windows, whose origin is opaque, fetch the fonts in CORS
// mode. Browsers revalidate them each time, so that a change reaches every
// reader at once. They keep the sandbox of every response: a logo in SVG
// opened on its own runs nothing. See
// docs/adr/0022-house-style-of-each-cahut.md.

func (s *Server) serveHouseStyle(w http.ResponseWriter, r *http.Request) {
	c := siteOf(r).cahut
	a, err := housestyle.Sheet(r.Context(), c.DB, c.Domain)
	if err != nil {
		s.internal(w, "make the house style sheet", err)
		return
	}
	serveHouseAsset(w, r, a)
}

func (s *Server) serveLogo(inverse bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := siteOf(r).cahut
		a, err := housestyle.Logo(r.Context(), c.DB, c.Domain, inverse)
		if err != nil {
			s.internal(w, "read the logo of the house style", err)
			return
		}
		serveHouseAsset(w, r, a)
	}
}

func (s *Server) serveFont(w http.ResponseWriter, r *http.Request) {
	a, ok, err := housestyle.LoadFont(r.Context(), siteOf(r).cahut.DB, r.PathValue("file"))
	if err != nil {
		s.internal(w, "read a font of the house style", err)
		return
	}
	if !ok {
		http.NotFound(w, r)
		return
	}
	serveHouseAsset(w, r, a)
}

func serveDeckEngine(w http.ResponseWriter, r *http.Request) {
	serveHouseAsset(w, r, housestyle.DeckEngine())
}

// serveTemplate serves a template of the kit, filled with sample content,
// in the house style of the Cahut, for people to look at. It runs in the
// sandbox of a Window that runs scripts and calls nothing.
func (s *Server) serveTemplate(w http.ResponseWriter, r *http.Request) {
	c := siteOf(r).cahut
	page, ok, err := housestyle.TemplatePage(r.Context(), c.DB, c.Domain, r.PathValue("kind"), time.Now())
	if err != nil {
		s.internal(w, "render a template of the house style", err)
		return
	}
	if !ok {
		reader, _ := s.readerOf(r, time.Now())
		s.notFound(w, r, reader.Handle)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Content-Security-Policy", (&window.Declaration{Scripts: true}).Policy(strings.TrimSuffix(s.cfg.BaseURL(c.Domain), "/")))
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	h.Set("Cache-Control", "no-cache")
	_, _ = w.Write(page)
}

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
	sum := sha256.Sum256(b)
	w.Header().Set("Content-Disposition", `attachment; filename="cahut-skill.zip"`)
	serveHouseAsset(w, r, housestyle.Asset{Type: "application/zip", ETag: `"` + hex.EncodeToString(sum[:16]) + `"`, Content: b})
}

func serveHouseAsset(w http.ResponseWriter, r *http.Request, a housestyle.Asset) {
	h := w.Header()
	h.Set("Content-Type", a.Type)
	h.Set("ETag", a.ETag)
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(a.Content))
}
