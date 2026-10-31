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

// policy lets the console load its own files, call its own API and frame
// the previews of Windows, and nothing else: no inline script, no other
// origin, no frame around it. Its scripts come from its assets only, which
// Vite builds for /console/: the Windows that AIs publish at the same
// origin hold scripts too, and none of them may run in the console, should
// a flaw ever let one in. The console of the server frames the reported
// Windows of the other Cahuts too, at their origins. See
// docs/adr/0027-previews-of-windows-in-the-console.md and
// docs/adr/0058-a-console-for-the-server.md.
func policy(r *http.Request, frames []string) string {
	origin := "http://" + r.Host
	if r.TLS != nil {
		origin = "https://" + r.Host
	}
	frameSrc := strings.Join(append([]string{"'self'"}, frames...), " ")
	return "default-src 'self'; script-src " + origin + "/console/assets/; style-src 'self'; img-src 'self' data:; font-src 'self'; " +
		"connect-src 'self'; frame-src " + frameSrc + "; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"
}

// notice is the page of a binary built without the console.
const notice = `<!doctype html>
<html lang="en">
<meta charset="utf-8">
<title>Cahut</title>
<p>This Cahut was built without its console. Its API answers at <code>/console/api/</code>.</p>
<p>To build the console into the binary, run <code>make build</code>, which needs Node.js.</p>
`

// Handler returns the handler of the pages of the console of a Cahut.
func Handler() http.Handler {
	return newHandler(app(), "index.html", nil)
}

// ServerHandler returns the handler of the pages of the console of the
// server, which frames the Windows of the origins that frames gives. See
// docs/adr/0058-a-console-for-the-server.md.
func ServerHandler(frames func(*http.Request) []string) http.Handler {
	return newHandler(app(), "server.html", frames)
}

// newHandler serves the console built into files, whose every page is the
// app of page, or the notice when files is nil. frames, when set, gives the
// other origins whose frames the pages show.
func newHandler(files fs.FS, page string, frames func(*http.Request) []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		var others []string
		if frames != nil {
			others = frames(r)
		}
		h.Set("Content-Security-Policy", policy(r, others))
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
		if name != "" && !strings.HasSuffix(name, "/") && name != "index.html" && name != "server.html" {
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
			// A missing file of the build, under assets/ or at the root, is
			// a 404; any other address is a page of the app, which routes it
			// itself, such as /cahuts/<domain> in the console of the server.
			if path.Ext(name) != "" && (strings.HasPrefix(name, "assets/") || !strings.Contains(name, "/")) {
				http.NotFound(w, r)
				return
			}
		}
		h.Set("Cache-Control", "no-cache")
		serve(w, r, files, page)
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
