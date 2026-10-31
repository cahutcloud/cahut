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
	if s.withdrawn(w, r, err, reader, false) {
		return
	}
	if err != nil {
		s.notFound(w, r, reader.Handle)
		return
	}
	pass := s.passOf(r, win.ID, now)
	how, err := win.Opens(r.Context(), c.DB, reader.UserID, pass)
	if err != nil {
		s.internal(w, "check the access of a Window", err)
		return
	}
	if how != window.BySession && how != window.ByPass {
		if win.Access == window.AccessKey {
			s.askKey(w, r, win, room+"/"+name, reader.Handle, "", http.StatusUnauthorized)
		} else {
			s.notFound(w, r, reader.Handle)
		}
		return
	}
	src := s.frameSrc(c, window.Opened{Served: win, How: how, Pass: pass}, reader.SessionID, now) +
		strings.TrimPrefix(r.URL.EscapedPath(), "/"+room+"/"+name+"/")
	if r.URL.RawQuery != "" {
		src += "?" + r.URL.RawQuery
	}
	// The frame has the sandbox of the Window's files.
	s.render(w, r, http.StatusOK, "frame", struct{ Title, Src, Sandbox string }{win.Title, src, win.Declaration.Sandbox()})
}

// servePermanent sends the reader of the permanent address of a Window,
// /-/w/<id>, to its address of the moment, with what follows the id and
// the query, when they may open it there. A reader who may not gets what
// that address would give them, without learning it: the page of what is
// not found, or the page of its Key, which names the Window by its id. The
// robots that draw the preview of a link to it get a page of Cahut that
// tells them what its address tells a stranger. See
// docs/adr/0050-window-ids-permanent-addresses-and-names.md and
// docs/adr/0051-sharing-a-window-and-its-preview.md.
func (s *Server) servePermanent(w http.ResponseWriter, r *http.Request) {
	c := siteOf(r).cahut
	now := time.Now()
	id := r.PathValue("id")
	win, err := window.LookupUID(r.Context(), c.DB, id, now)
	if err != nil && !isNotFound(err) {
		s.internal(w, "find a Window", err)
		return
	}
	// The robots of previews get a page of their own, people the Window.
	w.Header().Add("Vary", "User-Agent")
	if _, ok := errors.AsType[*window.Withdrawn](err); ok {
		reader, _ := s.readerOf(r, now)
		s.withdrawn(w, r, err, reader, false)
		return
	}
	if err == nil && r.PathValue("file") == "" && previewRobot(r) {
		s.servePreview(w, r, c, win)
		return
	}
	if err != nil || win.Access != window.AccessPublic {
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
		if how == window.Closed {
			if win.Access == window.AccessKey {
				s.askKey(w, r, win, win.UID, reader.Handle, "", http.StatusUnauthorized)
			} else {
				s.notFound(w, r, reader.Handle)
			}
			return
		}
	}
	target := "/" + win.Room + "/" + win.Name + "/" +
		strings.TrimPrefix(strings.TrimPrefix(r.URL.EscapedPath(), window.PermanentPrefix+id), "/")
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	// Not a permanent redirect: the name of the Window may change.
	http.Redirect(w, r, target, http.StatusTemporaryRedirect) //nolint:gosec // a path of this host, from the names of a Window
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
	if wd, ok := errors.AsType[*window.Withdrawn](err); ok {
		// The members of its Room still preview it in the console, and its
		// moderators open it while a report about it waits; its readers
		// learn that it was withdrawn. See
		// docs/adr/0054-reporting-and-withdrawing-a-window.md.
		switch frame.Kind {
		case window.FrameByConsole, window.FrameByModerator, window.FrameByHost:
			win, err = wd.Window, nil
		default:
			s.withdrawn(w, r, err, account.SessionOwner{}, true)
			return
		}
	}
	if err != nil && !isNotFound(err) {
		s.internal(w, "find a Window", err)
		return
	}
	if err != nil {
		gone()
		return
	}
	if frame.Kind == window.FrameByModerator || frame.Kind == window.FrameByHost {
		opens, err := c.Windows.ModeratorOpens(r.Context(), frame, now)
		if err != nil {
			s.internal(w, "verify the frame of a moderator", err)
			return
		}
		if !opens {
			gone()
			return
		}
		s.serveFiles(w, r, c, win, window.FramePrefix+signed, "no-referrer")
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

// withdrawn answers a reader who reaches a Window that the moderators of
// its Cahut withdrew, when err says so, and reports whether it did. A
// reader who would open the Window otherwise, who holds its Key or one of
// its links, or who reaches a public Window, gets the page that says it was
// withdrawn, with the status 410; anyone else gets the page of what is not
// found, and does not learn that it was there. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.
func (s *Server) withdrawn(w http.ResponseWriter, r *http.Request, err error, reader account.SessionOwner, held bool) bool {
	wd, ok := errors.AsType[*window.Withdrawn](err)
	if !ok {
		return false
	}
	win := wd.Window
	tell := held || win.Access == window.AccessPublic || win.Access == window.AccessKey
	if !tell {
		how, err := win.Opens(r.Context(), siteOf(r).cahut.DB, reader.UserID, s.passOf(r, win.ID, time.Now()))
		if err != nil {
			s.internal(w, "check the access of a Window", err)
			return true
		}
		tell = how != window.Closed
	}
	if !tell {
		s.notFound(w, r, reader.Handle)
		return true
	}
	p := s.page(r, reader.Handle, "Withdrawn")
	p.Report = ""
	s.render(w, r, http.StatusGone, "withdrawn", p)
	return true
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
