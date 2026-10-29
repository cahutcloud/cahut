package web

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	charte "cahut.cloud/docs/charte-graphique"
	"cahut.cloud/internal/address"
	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/window"
)

// The pages that a Cahut serves to readers on the port of its Windows: its
// home, the page of a Room, the frame of a protected Window, the page of a
// Key and the page of what is not found. They carry the brand of Cahut.
// Their foot links the legal pages of the Cahut, when it has some; the
// frame of a Window does not. See
// docs/adr/0019-window-versions-access-and-declaration.md,
// docs/adr/0032-the-page-of-a-room.md,
// docs/adr/0042-the-legal-pages-of-a-cahut.md and
// docs/adr/0049-no-house-style.md.

//go:embed pages
var pageFiles embed.FS

var pages = template.Must(template.New("pages").Funcs(template.FuncMap{"icon": icon, "brand": brandSVG}).
	ParseFS(pageFiles, "pages/pages.html"))

// pagePolicy lets a page of the Cahut at base load its style sheets, the
// logos and the fonts of the brand and its two scripts, and frame a Window of its own
// origin, and nothing else. The scripts are named: the Windows at the same
// origin hold scripts too, and none may run here.
func pagePolicy(base string) string {
	return "default-src 'none'; style-src 'self'; script-src " + base + "-/frame.js " + base + "-/pages.js; " +
		"img-src 'self' data:; font-src 'self'; frame-src 'self'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'"
}

// page is what every page shows at its top, and where to sign in.
type page struct {
	Title  string // of the page, as its tab shows it
	Domain string
	Reader string      // the handle of the reader signed in; empty when there is none
	SignIn string      // where to sign in, to come back to this page
	Legal  []legalLink // the legal pages of the Cahut, which its foot links
}

// legalLink is a legal page of the Cahut, as the foot of a page links it.
type legalLink struct {
	Name, URL string
}

// legalNames name the legal pages at the foot of the pages, which speak
// English. See docs/adr/0042-the-legal-pages-of-a-cahut.md.
var legalNames = map[string]string{
	"legal_notice": "Legal notice",
	"terms":        "Terms",
	"privacy":      "Privacy",
	"support":      "Support",
}

func (s *Server) page(r *http.Request, reader, title string) page {
	c := siteOf(r).cahut
	p := page{Title: title, Domain: c.Domain, Reader: reader, SignIn: "/-/sign-in?to=" + url.QueryEscape(r.URL.RequestURI())}
	legal, err := settings.Legal(r.Context(), c.DB.Read)
	if err != nil {
		s.log.Error("read the legal pages", "cahut", c.Domain, "error", err)
	}
	for _, l := range legal.In("") {
		p.Legal = append(p.Legal, legalLink{Name: legalNames[l.Page], URL: l.URL})
	}
	return p
}

// render writes the page called name with data, and the headers of every
// page of Cahut. The pages depend on who reads them: no cache keeps them.
func (s *Server) render(w http.ResponseWriter, r *http.Request, status int, name string, data any) {
	var b bytes.Buffer
	if err := pages.ExecuteTemplate(&b, name, data); err != nil {
		s.log.Error("render a page", "page", name, "error", err)
		http.Error(w, "Internal error.", http.StatusInternalServerError)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Content-Security-Policy", pagePolicy(s.cfg.BaseURL(siteOf(r).cahut.Domain)))
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "same-origin")
	// Another site that opens a page of Cahut cannot count its frames,
	// which would tell whether its visitor may open a Window.
	h.Set("Cross-Origin-Opener-Policy", "same-origin")
	h.Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(b.Bytes())
}

// serveAsset serves the style sheet and the scripts of the pages of Cahut.
func serveAsset(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Cache-Control", "no-cache")
		http.ServeFileFS(w, r, pageFiles, "pages/"+name)
	}
}

// BrandPath is where every Cahut serves the files of the brand of Cahut
// that its pages use, under the paths that they have in the brand
// guidelines.
const BrandPath = "/-/brand/"

// brandTypes are the types of the files of the brand, by extension.
var brandTypes = map[string]string{
	".css":   "text/css; charset=utf-8",
	".svg":   "image/svg+xml",
	".png":   "image/png",
	".ico":   "image/x-icon",
	".woff2": "font/woff2",
	".txt":   "text/plain; charset=utf-8",
}

// brandAssets returns the files of the brand by path, with their ETags.
var brandAssets = sync.OnceValue(func() map[string]servedFile {
	assets := map[string]servedFile{}
	err := fs.WalkDir(charte.Files, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(charte.Files, p)
		if err != nil {
			return err
		}
		assets[p] = newServedFile(brandTypes[path.Ext(p)], b)
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("read the files of the brand: %v", err)) // embedded: a build that runs reads them
	}
	return assets
})

// serveBrand serves a file of the brand of Cahut. Its folders list
// nothing.
func serveBrand(w http.ResponseWriter, r *http.Request) {
	a, ok := brandAssets()[r.PathValue("file")]
	if !ok {
		http.NotFound(w, r)
		return
	}
	serveFile(w, r, a)
}

// rootIcons are the icons that browsers, and the services that draw the
// icon of a site for AI clients, ask for at the root of a domain, by the
// path of the file of the brand that each one serves.
var rootIcons = map[string]string{
	"/favicon.ico":          "logos/favicon/favicon.ico",
	"/apple-touch-icon.png": "logos/favicon/apple-touch-icon.png",
}

// serveRootIcon serves the icon of Cahut at the root of the Cahut, as its
// pages carry its brand: the tab of every page of the domain shows it,
// Windows that name no icon included. The console names its icons under
// BrandPath. See docs/adr/0040-the-icon-of-a-cahut.md and
// docs/adr/0049-no-house-style.md.
func serveRootIcon(w http.ResponseWriter, r *http.Request) {
	serveFile(w, r, brandAssets()[rootIcons[r.URL.Path]])
}

// brandSVG writes inline the picture of the brand at p, so that its
// strokes take the color of the text, in light and dark themes alike.
func brandSVG(p string) template.HTML {
	return template.HTML(brandAssets()[p].Content) //nolint:gosec // a file of the brand, embedded
}

// notFound answers what does not exist, and what its reader may not open,
// alike, with a link to sign in.
func (s *Server) notFound(w http.ResponseWriter, r *http.Request, reader string) {
	s.render(w, r, http.StatusNotFound, "notfound", s.page(r, reader, "Not found"))
}

// serveRoot sends the root of the Cahut to its console, which opens on its
// overview for whoever has signed in, and on its page of sign-in for the
// others: only the console knows its session, which lives in the browser.
// See docs/adr/0038-the-root-of-a-cahut-opens-its-console.md.
func (s *Server) serveRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.cfg.ConsoleURL(siteOf(r).cahut.Domain), http.StatusSeeOther)
}

// roomPage is the page that lists the Windows of a Room.
type roomPage struct {
	page
	Name    string
	Count   string
	Member  bool
	MCP     string
	Windows []windowCard
}

// windowCard is a Window on the page of its Room, with the address of its
// preview.
type windowCard struct {
	URL, Title         string
	Kind, KindIcon     string
	Access, AccessIcon string
	AccessClass        string
	Updated            string
	Preview            string // empty when the Window has no preview
}

// serveRoom serves the page of a Room: the Window that the Room shows at
// its address, when its reader opens it, or else the Windows that its
// reader opens. A reader outside the Room who opens none of its Windows
// finds nothing. See docs/adr/0032-the-page-of-a-room.md.
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
	if v.Home != nil {
		src := s.frameSrc(c, *v.Home, reader.SessionID, now)
		if r.URL.RawQuery != "" {
			src += "?" + r.URL.RawQuery
		}
		s.render(w, r, http.StatusOK, "frame", struct{ Title, Src, Sandbox string }{v.Home.Title, src, v.Home.Declaration.Sandbox()})
		return
	}
	data := roomPage{
		page:   s.page(r, reader.Handle, v.Title),
		Name:   v.Name,
		Count:  plural(len(v.Windows), "Window", "Windows"),
		Member: v.Member,
		MCP:    s.cfg.BaseURL(c.Domain) + "mcp",
	}
	for _, o := range v.Windows {
		kind, ok := kinds[o.Kind]
		if !ok {
			kind = kinds["site"]
		}
		access := accesses[o.Access]
		data.Windows = append(data.Windows, windowCard{
			URL: "/" + o.Room + "/" + o.Name + "/", Title: o.Title,
			Kind: kind.Name, KindIcon: kind.Icon,
			Access: access.Name, AccessIcon: access.Icon, AccessClass: o.Access,
			Updated: ago(o.UpdatedAt, now),
			Preview: s.frameSrc(c, o, reader.SessionID, now),
		})
	}
	s.render(w, r, http.StatusOK, "room", data)
}

// frameSrc returns the address that opens the Window o in a frame of a page
// of Cahut, for the reader of the session sessionID: its own address when
// it is public, and otherwise a frame that names it alone. See
// docs/adr/0019-window-versions-access-and-declaration.md.
func (s *Server) frameSrc(c *cahut.Cahut, o window.Opened, sessionID int64, now time.Time) string {
	frame := window.Frame{Window: o.ID, Expires: now.Add(window.FrameTTL).Unix()}
	switch o.How {
	case window.BySession:
		frame.Kind, frame.ID = window.FrameBySession, sessionID
	case window.ByPass:
		frame.Kind, frame.ID = window.FrameByPass, o.Pass
	default:
		return "/" + o.Room + "/" + o.Name + "/"
	}
	return window.FramePrefix + frame.Sign(c.FrameKey) + "/"
}

// ago says how long ago a time was, as the console does: "2 min ago",
// "Yesterday" or "Oct 12".
func ago(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "Just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d h ago", int(d.Hours()))
	case d < 48*time.Hour:
		return "Yesterday"
	case t.Year() != now.Year():
		return t.Format("Jan 2, 2006")
	}
	return t.Format("Jan 2")
}

// plural says a count with its noun, as in "1 Window" or "3 Windows".
func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
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
