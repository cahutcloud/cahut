package web

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/window"
)

// The pages that a Cahut serves to readers on the port of its Windows: its
// home, the page of a Room, the frame of a protected Window, the page of a
// Key and the page of what is not found. See
// docs/adr/0019-window-versions-access-and-declaration.md.

//go:embed pages
var pageFiles embed.FS

var pages = template.Must(template.ParseFS(pageFiles, "pages/pages.html"))

// pagePolicy lets a page of Cahut load its style sheet and its script, and
// frame a Window of its own origin, and nothing else.
const pagePolicy = "default-src 'none'; style-src 'self'; script-src 'self'; img-src 'self' data:; frame-src 'self'; " +
	"form-action 'self'; base-uri 'none'; frame-ancestors 'none'"

// page is what every page shows at its top, and where to sign in.
type page struct {
	Domain string
	Reader string // the handle of the reader signed in; empty when there is none
	SignIn string // where to sign in, to come back to this page
}

func (s *Server) page(r *http.Request, reader string) page {
	return page{Domain: siteOf(r).cahut.Domain, Reader: reader, SignIn: "/-/sign-in?to=" + url.QueryEscape(r.URL.RequestURI())}
}

// render writes the page called name with data, and the headers of every
// page of Cahut. The pages depend on who reads them: no cache keeps them.
func (s *Server) render(w http.ResponseWriter, status int, name string, data any) {
	var b bytes.Buffer
	if err := pages.ExecuteTemplate(&b, name, data); err != nil {
		s.log.Error("render a page", "page", name, "error", err)
		http.Error(w, "Internal error.", http.StatusInternalServerError)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Content-Security-Policy", pagePolicy)
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "same-origin")
	h.Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(b.Bytes())
}

// serveAsset serves the style sheet and the script of the pages of Cahut.
func serveAsset(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Cache-Control", "no-cache")
		http.ServeFileFS(w, r, pageFiles, "pages/"+name)
	}
}

// notFound answers what does not exist, and what its reader may not open,
// alike, with a link to sign in.
func (s *Server) notFound(w http.ResponseWriter, r *http.Request, reader string) {
	s.render(w, http.StatusNotFound, "notfound", s.page(r, reader))
}

// serveRoot serves the home of the Cahut: the Rooms of a reader signed in,
// or a link to sign in.
func (s *Server) serveRoot(w http.ResponseWriter, r *http.Request) {
	c := siteOf(r).cahut
	reader, _ := s.readerOf(r, time.Now())
	data := struct {
		page
		Rooms []room.Access
		MCP   string
	}{page: s.page(r, reader.Handle), MCP: s.cfg.BaseURL(c.Domain) + "mcp"}
	if reader.UserID != 0 {
		var err error
		if data.Rooms, err = room.Of(r.Context(), c.DB, reader.UserID); err != nil {
			s.internal(w, "list the Rooms of a reader", err)
			return
		}
	}
	s.render(w, http.StatusOK, "home", data)
}

// serveRoom serves the page of a Room: the Windows that its reader opens.
// A reader outside the Room who opens none of its Windows finds nothing.
func (s *Server) serveRoom(w http.ResponseWriter, r *http.Request) {
	c := siteOf(r).cahut
	now := time.Now()
	reader, _ := s.readerOf(r, now)
	name := r.PathValue("room")
	v, err := window.ViewRoom(r.Context(), c.DB, name, reader.UserID, s.passes(r), now)
	if err != nil && !isNotFound(err) {
		s.internal(w, "view a Room", err)
		return
	}
	if err != nil || !v.Member && len(v.Windows) == 0 {
		s.notFound(w, r, reader.Handle)
		return
	}
	type item struct{ URL, Title, Kind, Access string }
	data := struct {
		page
		Title   string
		Windows []item
	}{page: s.page(r, reader.Handle), Title: v.Title}
	for _, w := range v.Windows {
		data.Windows = append(data.Windows, item{URL: "/" + w.Room + "/" + w.Name + "/", Title: w.Title, Kind: w.Kind, Access: accessNames[w.Access]})
	}
	s.render(w, http.StatusOK, "room", data)
}

// accessNames says who opens a Window, on the page of its Room.
var accessNames = map[string]string{
	window.AccessRoom:     "Room only",
	window.AccessInternal: "Internal",
	window.AccessKey:      "With its Key",
	window.AccessPublic:   "Public",
}

// redirectToRoom sends <room> to <room>/, the page of the Room.
func (s *Server) redirectToRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "This address answers GET only.", http.StatusMethodNotAllowed)
		return
	}
	if name := r.PathValue("room"); !validName(name) || address.Reserved(name) {
		reader, _ := s.readerOf(r, time.Now())
		s.notFound(w, r, reader.Handle)
		return
	}
	http.Redirect(w, r, "/"+r.PathValue("room")+"/", http.StatusPermanentRedirect)
}

func (s *Server) internal(w http.ResponseWriter, what string, err error) {
	s.log.Error(what, "error", err)
	http.Error(w, "Internal error.", http.StatusInternalServerError)
}
