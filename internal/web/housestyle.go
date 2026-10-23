package web

import (
	"bytes"
	"net/http"
	"time"

	"cahut.cloud/internal/housestyle"
)

// The files of the house style of a Cahut: its style sheet, its logos and
// its fonts, at stable addresses under /-/. They are public, and the pages
// of Windows, whose origin is opaque, fetch the fonts in CORS mode. Browsers
// revalidate them each time, so that a change reaches every reader at once.
// They keep the sandbox of every response: a logo in SVG opened on its own
// runs nothing. See docs/adr/0022-house-style-of-each-cahut.md.

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

func serveHouseAsset(w http.ResponseWriter, r *http.Request, a housestyle.Asset) {
	h := w.Header()
	h.Set("Content-Type", a.Type)
	h.Set("ETag", a.ETag)
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(a.Content))
}
