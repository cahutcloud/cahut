package web

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/window"
)

// sandbox gives every page of a Window an opaque origin, so that its scripts
// can read neither the other Windows of the Cahut nor the reader's session.
// See docs/adr/0010-windows-in-a-csp-sandbox.md.
const sandbox = "sandbox allow-scripts allow-forms allow-modals allow-popups allow-popups-to-escape-sandbox allow-downloads"

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

// serveWindow serves the files of a Window. Until access levels exist,
// anyone who has a Window's address can read it.
func (s *Server) serveWindow(w http.ResponseWriter, r *http.Request) {
	room, name := r.PathValue("room"), r.PathValue("window")
	if !address.ValidName(room) || !address.ValidName(name) {
		http.NotFound(w, r)
		return
	}
	c := siteOf(r).cahut
	id, version, err := window.Current(r.Context(), c.DB, room, name)
	if errors.Is(err, operation.ErrNotFound) {
		http.Error(w, "There is no Window at this address.", http.StatusNotFound)
		return
	}
	if err != nil {
		s.log.Error("find a Window", "cahut", c.Domain, "window", room+"/"+name, "error", err)
		http.Error(w, "Internal error.", http.StatusInternalServerError)
		return
	}
	root, err := os.OpenRoot(c.Windows.Dir(id, version))
	if err != nil {
		s.log.Error("open the files of a Window", "cahut", c.Domain, "window", room+"/"+name, "error", err)
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

	h := w.Header()
	h.Set("Content-Security-Policy", sandbox)
	// The sandbox makes the page's module scripts, fetch calls and fonts
	// cross-origin requests; they carry no credentials.
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	h.Set("Cache-Control", "no-cache")
	http.StripPrefix("/"+room+"/"+name, http.FileServerFS(files)).ServeHTTP(w, r)
}
