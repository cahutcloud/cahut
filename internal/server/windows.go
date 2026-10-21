package server

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/window"
)

// serveWindow serves the files of the Window called name. Until Doors
// exist, anyone who has a Window's address can read it.
func (s *Server) serveWindow(w http.ResponseWriter, r *http.Request, name string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Windows are read-only.", http.StatusMethodNotAllowed)
		return
	}
	version, err := window.CurrentVersion(r.Context(), s.home.DB, name)
	if errors.Is(err, operation.ErrNotFound) {
		http.Error(w, "There is no Window at this address.", http.StatusNotFound)
		return
	}
	if err != nil {
		s.log.Error("find a Window", "window", name, "error", err)
		http.Error(w, "Internal error.", http.StatusInternalServerError)
		return
	}
	root, err := os.OpenRoot(s.home.Windows.Dir(name, version))
	if err != nil {
		s.log.Error("open the files of a Window", "window", name, "error", err)
		http.Error(w, "Internal error.", http.StatusInternalServerError)
		return
	}
	defer root.Close()
	files := root.FS()

	// No directory listings: a folder without index.html does not exist.
	p := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
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
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	h.Set("Cache-Control", "no-cache")
	http.FileServerFS(files).ServeHTTP(w, r)
}
