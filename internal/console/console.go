// Package console serves the web console of every Cahut: a Preact app,
// built by Vite from ui/ and embedded in the binary. It calls the API of the
// console and nothing else. A binary built without the console, as a plain
// go build is, serves a notice instead. See
// docs/adr/0005-preact-console-embedded.md and
// docs/adr/0020-console-under-a-path-of-port-443.md.
package console

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

// policy lets the console load its own files and call its own API, and
// nothing else: no inline script, no other origin, no frame around it.
const policy = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; " +
	"connect-src 'self'; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

// notice is the page of a binary built without the console.
const notice = `<!doctype html>
<html lang="en">
<meta charset="utf-8">
<title>Cahut</title>
<p>This Cahut was built without its console. Its API answers at <code>/console/api/</code>.</p>
<p>To build the console into the binary, run <code>make build</code>, which needs Node.js.</p>
`

// Handler returns the handler of the console's pages.
func Handler() http.Handler {
	return newHandler(app())
}

// newHandler serves the console built into files, or the notice when files
// is nil.
func newHandler(files fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", policy)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			h.Set("Allow", "GET, HEAD")
			http.Error(w, "The console's pages answer GET only.", http.StatusMethodNotAllowed)
			return
		}
		if files == nil {
			h.Set("Content-Type", "text/html; charset=utf-8")
			h.Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, notice)
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name != "" && !strings.HasSuffix(name, "/") && name != "index.html" {
			if info, err := fs.Stat(files, name); err == nil && !info.IsDir() {
				// Vite names the files of assets/ after their contents.
				if strings.HasPrefix(name, "assets/") {
					h.Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					h.Set("Cache-Control", "no-cache")
				}
				serve(w, r, files, name)
				return
			}
			// A missing file is a 404; any other address is a page of the
			// app, which routes it itself.
			if path.Ext(name) != "" {
				http.NotFound(w, r)
				return
			}
		}
		h.Set("Cache-Control", "no-cache")
		serve(w, r, files, "index.html")
	})
}

// serve writes the file name of files.
func serve(w http.ResponseWriter, r *http.Request, files fs.FS, name string) {
	b, err := fs.ReadFile(files, name)
	if errors.Is(err, fs.ErrNotExist) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "Internal error.", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(b))
}
