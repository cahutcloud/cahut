package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/address"
	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/window"
)

// Readers open Windows on port 443. A person signs in there through the
// console, which hands a single-use code to /-/session; a reader who types
// the Key of a Window gets a pass for it. Both live in cookies that only
// Cahut sets. See docs/adr/0011-console-on-its-own-port.md and
// docs/adr/0019-window-versions-access-and-declaration.md.

// The cookies of readers, without the __Host- prefix that they take over
// HTTPS. A pass takes the id of its Window after its name.
const (
	readerCookie = "cahut-reader"
	passCookie   = "cahut-key-" //nolint:gosec // a prefix, not a credential
)

// Attempts at a Key are limited by address, and their failures by Window.
const (
	KeyBurst       = 10
	KeyEvery       = 6 * time.Second
	KeyWindowBurst = 20
	KeyWindowEvery = 3 * time.Second
)

// cookieName returns the name of a cookie of Cahut: with the __Host- prefix,
// which binds it to the host over HTTPS, but for development in plain HTTP.
func (s *Server) cookieName(name string) string {
	if s.cfg.Dev {
		return name
	}
	return "__Host-" + name
}

// setCookie sets a cookie of Cahut until expires, or removes it when
// expires is zero.
func (s *Server) setCookie(w http.ResponseWriter, name, value string, expires time.Time) {
	c := &http.Cookie{ //nolint:gosec // Secure, but for development in plain HTTP
		Name:     s.cookieName(name),
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   !s.cfg.Dev,
		SameSite: http.SameSiteLaxMode,
	}
	if expires.IsZero() {
		c.MaxAge = -1
	} else {
		c.MaxAge = max(int(time.Until(expires).Seconds()), 1)
	}
	http.SetCookie(w, c)
}

// readerOf returns the reader session of the request, or none.
func (s *Server) readerOf(r *http.Request, now time.Time) (account.SessionOwner, bool) {
	cookie, err := r.Cookie(s.cookieName(readerCookie))
	if err != nil {
		return account.SessionOwner{}, false
	}
	c := siteOf(r).cahut
	o, err := account.VerifyReader(r.Context(), c.DB, cookie.Value, now)
	if err != nil {
		if !errors.Is(err, account.ErrInvalidSession) {
			s.log.Error("verify a reader session", "cahut", c.Domain, "error", err)
		}
		return account.SessionOwner{}, false
	}
	return o, true
}

// passes returns the passes of the request, by the id of their Window.
func (s *Server) passes(r *http.Request) map[int64]string {
	prefix := s.cookieName(passCookie)
	passes := map[int64]string{}
	for _, c := range r.Cookies() {
		if id, ok := strings.CutPrefix(c.Name, prefix); ok {
			if n, err := strconv.ParseInt(id, 10, 64); err == nil {
				passes[n] = c.Value
			}
		}
	}
	return passes
}

// passOf returns the id of the request's pass for the Window id, or zero.
func (s *Server) passOf(r *http.Request, id int64, now time.Time) int64 {
	cookie, err := r.Cookie(s.cookieName(passCookie + strconv.FormatInt(id, 10)))
	if err != nil {
		return 0
	}
	c := siteOf(r).cahut
	pass, err := window.Pass(r.Context(), c.DB, id, cookie.Value, now)
	if err != nil {
		s.log.Error("verify a pass", "cahut", c.Domain, "error", err)
		return 0
	}
	return pass
}

// localPath returns to when it is a path of this host, with its query, and
// / otherwise, so that no address sends a reader to another site.
func localPath(to string) string {
	if !strings.HasPrefix(to, "/") || strings.HasPrefix(to, "//") || strings.HasPrefix(to, `/\`) {
		return "/"
	}
	u, err := url.Parse(to)
	if err != nil || u.Scheme != "" || u.Host != "" || u.User != nil {
		return "/"
	}
	return u.RequestURI()
}

// serveSignIn sends a reader to the console, which signs them in, then
// sends them back to the path they came from.
func (s *Server) serveSignIn(w http.ResponseWriter, r *http.Request) {
	to := localPath(r.URL.Query().Get("to"))
	target := s.cfg.ConsoleURL(siteOf(r).cahut.Domain) + "read?to=" + url.QueryEscape(to)
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// serveSession exchanges the code that the console handed over for a
// reader session, then sends the reader where they were going.
func (s *Server) serveSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Referrer-Policy", "no-referrer")
	c := siteOf(r).cahut
	to := localPath(r.URL.Query().Get("to"))
	rs, err := account.OpenReaderSession(r.Context(), c.DB, r.URL.Query().Get("code"), time.Now())
	if errors.Is(err, account.ErrInvalidSession) {
		p := s.page(r, "")
		p.SignIn = "/-/sign-in?to=" + url.QueryEscape(to)
		s.render(w, http.StatusForbidden, "expired", p)
		return
	}
	if err != nil {
		s.internal(w, "open a reader session", err)
		return
	}
	s.setCookie(w, readerCookie, rs.Token, rs.ExpiresAt)
	http.Redirect(w, r, to, http.StatusSeeOther) //nolint:gosec // localPath keeps it on this host
}

// serveSignOut ends the reader session of the request.
func (s *Server) serveSignOut(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(s.cookieName(readerCookie)); err == nil {
		if err := account.EndReaderSession(r.Context(), siteOf(r).cahut.DB, cookie.Value); err != nil {
			s.internal(w, "end a reader session", err)
			return
		}
	}
	s.setCookie(w, readerCookie, "", time.Time{})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// keyPage is the page that asks for the Key of a Window.
type keyPage struct {
	page
	Title, Window, Error string
}

func (s *Server) askKey(w http.ResponseWriter, r *http.Request, win window.Served, reader, msg string, status int) {
	p := s.page(r, reader)
	p.SignIn = "/-/sign-in?to=" + url.QueryEscape("/"+win.Room+"/"+win.Name+"/")
	s.render(w, status, "key", keyPage{page: p, Title: win.Title, Window: win.Room + "/" + win.Name, Error: msg})
}

// serveKey checks the Key that a reader typed, and gives them a pass for
// its Window.
func (s *Server) serveKey(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "The form could not be read.", http.StatusBadRequest)
		return
	}
	c := siteOf(r).cahut
	now := time.Now()
	reader, _ := s.readerOf(r, now)
	roomName, name, _ := strings.Cut(r.PostForm.Get("window"), "/")
	if !address.ValidName(roomName) || !address.ValidName(name) {
		s.notFound(w, r, reader.Handle)
		return
	}
	win, err := window.Lookup(r.Context(), c.DB, roomName, name, now)
	if err != nil && !isNotFound(err) {
		s.internal(w, "find a Window", err)
		return
	}
	if err != nil || win.Access != window.AccessKey {
		s.notFound(w, r, reader.Handle)
		return
	}
	byWindow := fmt.Sprintf("%s/%d", c.Domain, win.ID)
	if !s.keyTries.allow(clientIP(r)) || s.keyFailures.spent(byWindow) {
		s.askKey(w, r, win, reader.Handle, "Too many tries: wait a minute, then try again.", http.StatusTooManyRequests)
		return
	}
	key := r.PostForm.Get("key")
	if len(key) > credential.MaxKey*4 {
		key = ""
	}
	token, expires, err := window.OpenWithKey(r.Context(), c.DB, win, key, now)
	if errors.Is(err, window.ErrWrongKey) {
		s.keyFailures.allow(byWindow)
		s.askKey(w, r, win, reader.Handle, "This Key does not open this Window.", http.StatusForbidden)
		return
	}
	if err != nil {
		s.internal(w, "open a Window with its Key", err)
		return
	}
	s.setCookie(w, passCookie+strconv.FormatInt(win.ID, 10), token, expires)
	http.Redirect(w, r, "/"+win.Room+"/"+win.Name+"/", http.StatusSeeOther)
}
