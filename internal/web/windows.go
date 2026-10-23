package web

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/address"
	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/window"
)

func validName(s string) bool {
	return address.ValidName(s)
}

func isNotFound(err error) bool {
	return errors.Is(err, operation.ErrNotFound)
}

// redirectToWindow sends <room>/<window> to <room>/<window>/, so that the
// relative links of the Window's pages resolve under it.
func redirectToWindow(w http.ResponseWriter, r *http.Request) {
	if !address.ValidName(r.PathValue("room")) || !address.ValidName(r.PathValue("window")) {
		http.NotFound(w, r)
		return
	}
	target := "/" + r.PathValue("room") + "/" + r.PathValue("window") + "/"
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusPermanentRedirect) //nolint:gosec // a path of this host, made of two checked names
}

// serveWindow serves a Window at its address: the files of a public
// Window; for a protected Window, the page of Cahut that frames it, or the
// page that asks for its Key; and otherwise the page of what is not found.
func (s *Server) serveWindow(w http.ResponseWriter, r *http.Request) {
	room, name := r.PathValue("room"), r.PathValue("window")
	if !address.ValidName(room) || !address.ValidName(name) {
		http.NotFound(w, r)
		return
	}
	c := siteOf(r).cahut
	now := time.Now()
	win, err := window.Lookup(r.Context(), c.DB, room, name, now)
	if err != nil && !isNotFound(err) {
		s.internal(w, "find a Window", err)
		return
	}
	if err == nil && win.Access == window.AccessPublic {
		s.serveFiles(w, r, c, win, "/"+room+"/"+name, "strict-origin-when-cross-origin")
		return
	}
	reader, _ := s.readerOf(r, now)
	if err != nil {
		s.notFound(w, r, reader.Handle)
		return
	}
	how, err := win.Opens(r.Context(), c.DB, reader.UserID, s.passOf(r, win.ID, now))
	if err != nil {
		s.internal(w, "check the access of a Window", err)
		return
	}
	var frame window.Frame
	switch how {
	case window.BySession:
		frame = window.Frame{Window: win.ID, Kind: window.FrameBySession, ID: reader.SessionID}
	case window.ByPass:
		frame = window.Frame{Window: win.ID, Kind: window.FrameByPass, ID: s.passOf(r, win.ID, now)}
	default:
		if win.Access == window.AccessKey {
			s.askKey(w, r, win, reader.Handle, "", http.StatusUnauthorized)
		} else {
			s.notFound(w, r, reader.Handle)
		}
		return
	}
	frame.Expires = now.Add(window.FrameTTL).Unix()
	src := window.FramePrefix + frame.Sign(c.FrameKey) + "/" + strings.TrimPrefix(r.URL.EscapedPath(), "/"+room+"/"+name+"/")
	if r.URL.RawQuery != "" {
		src += "?" + r.URL.RawQuery
	}
	// The frame has the sandbox of the Window's files.
	s.render(w, r, http.StatusOK, "frame", struct{ Title, Src, Sandbox string }{win.Title, src, win.Declaration.Sandbox()})
}

// serveFrame serves the files of a Window in its frame, under the
// capability that the page of a protected Window gave, or that the console
// got for its preview, while its reader may still open it.
func (s *Server) serveFrame(w http.ResponseWriter, r *http.Request) {
	c := siteOf(r).cahut
	now := time.Now()
	gone := func() {
		http.Error(w, "This frame no longer opens its Window: open the Window at its address again.", http.StatusNotFound)
	}
	signed := r.PathValue("capability")
	frame, ok := window.VerifyFrame(c.FrameKey, signed, now)
	if !ok {
		gone()
		return
	}
	win, err := window.LookupID(r.Context(), c.DB, frame.Window, now)
	if err != nil && !isNotFound(err) {
		s.internal(w, "find a Window", err)
		return
	}
	if err != nil {
		gone()
		return
	}
	var userID, pass int64
	switch frame.Kind {
	case window.FrameBySession:
		o, err := account.Reader(r.Context(), c.DB, frame.ID, now)
		if err != nil && !errors.Is(err, account.ErrInvalidSession) {
			s.internal(w, "verify a reader session", err)
			return
		}
		userID = o.UserID
	case window.FrameByConsole:
		o, err := account.Console(r.Context(), c.DB, frame.ID, now)
		if err != nil && !errors.Is(err, account.ErrInvalidSession) {
			s.internal(w, "verify a session", err)
			return
		}
		userID = o.UserID
	case window.FrameByPass:
		lasts, err := window.PassLasts(r.Context(), c.DB, win.ID, frame.ID, now)
		if err != nil {
			s.internal(w, "verify a pass", err)
			return
		}
		if lasts {
			pass = frame.ID
		}
	}
	how, err := win.Opens(r.Context(), c.DB, userID, pass)
	if err != nil {
		s.internal(w, "check the access of a Window", err)
		return
	}
	if how == window.Closed {
		gone()
		return
	}
	s.serveFiles(w, r, c, win, window.FramePrefix+signed, "no-referrer")
}

// serveFiles serves the file of the Window win that the request names
// under prefix, in the sandbox.
func (s *Server) serveFiles(w http.ResponseWriter, r *http.Request, c *cahut.Cahut, win window.Served, prefix, referrer string) {
	root, err := os.OpenRoot(c.Windows.Dir(win.ID, win.Version))
	if err != nil {
		s.log.Error("open the files of a Window", "cahut", c.Domain, "window", win.Room+"/"+win.Name, "error", err)
		http.Error(w, "Internal error.", http.StatusInternalServerError)
		return
	}
	defer root.Close()
	files := root.FS()

	// No directory listings: a folder without index.html does not exist.
	p := path.Clean("/" + r.PathValue("file"))[1:]
	if p == "" {
		p = "."
	}
	if info, err := fs.Stat(files, p); err == nil && info.IsDir() {
		if _, err := fs.Stat(files, path.Join(p, "index.html")); err != nil {
			http.NotFound(w, r)
			return
		}
	}

	// Every file runs in a sandbox, which gives its page an opaque origin, so
	// that its scripts can read neither the other Windows of the Cahut nor
	// the reader's session, and loads and calls only what its version
	// declares. See docs/adr/0010-windows-in-a-csp-sandbox.md.
	h := w.Header()
	h.Set("Content-Security-Policy", win.Declaration.Policy(strings.TrimSuffix(s.cfg.BaseURL(c.Domain), "/")))
	// The sandbox makes the page's module scripts, fetch calls and fonts
	// cross-origin requests; they carry no credentials.
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", referrer)
	h.Set("Cache-Control", "no-cache")
	http.StripPrefix(prefix, http.FileServerFS(files)).ServeHTTP(w, r)
}
