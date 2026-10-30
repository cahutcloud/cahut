package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/card"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/window"
)

// Share links open a Window to whoever holds them. Every visit to the
// address of a link gets its page, which opens nothing and counts nothing,
// and carries the tags that the robots of previews read; a click on its
// button spends an opening and gives a pass, as the right Key does. The
// page sends no referrer, and the Window never sees the link. See
// docs/adr/0052-share-links.md.

// Visits to the addresses of links are limited by address, as sign-in is,
// with a budget of their own: a token cannot be guessed, but each visit
// reads the database, and the people of an office behind one address may
// open the same link together.
const (
	LinkBurst = 60
	LinkEvery = time.Second
)

// linkPage is the page of a share link.
type linkPage struct {
	page
	Description string // what og:description says
	URL         string // the link itself
	Image       string // the address of the card of the Window
	ImageAlt    string
	Until       string // the end of the link, in words
	Limited     bool   // it opens to a number of readers only
	Left        int    // how many, from now on
	Once        bool   // it opens once, for a session of reading
	Hours       int    // how long that session lasts
}

// serveLink answers a visit to the address of a share link with its page,
// or sends a reader who opens its Window already straight to it. A link
// that no longer opens gives the page of what it no longer opens.
func (s *Server) serveLink(w http.ResponseWriter, r *http.Request) {
	l, ok := s.findLink(w, r)
	if !ok {
		return
	}
	c := siteOf(r).cahut
	base := strings.TrimSuffix(s.cfg.BaseURL(c.Domain), "/")
	title := cardTitle(l.Window)
	p := linkPage{
		Description: "Opens with this link until " + mail.Day(l.ExpiresAt) + ".",
		URL:         base + window.LinkPrefix + r.PathValue("token"),
		Image:       base + CardPath + l.Window.UID + ".png?v=" + card.Tag(title, c.Domain),
		ImageAlt:    "A Window at " + c.Domain,
		Until:       mail.Day(l.ExpiresAt),
		Limited:     l.Limited,
		Left:        l.Left,
		Once:        l.Once,
		Hours:       int(window.OnceTTL / time.Hour),
	}
	if l.Once {
		p.Description = "Opens once, until " + mail.Day(l.ExpiresAt) + "."
	}
	if title != "" {
		p.ImageAlt = title + ", a Window at " + c.Domain
	}
	p.page = s.page(r, "", l.Window.Title)
	// The link stays out of the address of the console.
	p.SignIn = ""
	s.render(w, r, http.StatusOK, "link", p)
}

// openLink takes the click on the page of a share link: it spends an
// opening, gives the reader a pass in the cookie of the passes of its
// Window, and sends them to the Window.
func (s *Server) openLink(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	l, ok := s.findLink(w, r)
	if !ok {
		return
	}
	c := siteOf(r).cahut
	in, err := json.Marshal(window.OpenShareLinkInput{Token: r.PathValue("token")})
	if err != nil {
		s.internal(w, "open a share link", err)
		return
	}
	actor := operation.Actor{Kind: operation.Visitor, Via: "web", Domain: c.Domain, Address: clientKey(r)}
	out, err := c.Registry.Invoke(r.Context(), "open_share_link", actor, in)
	if errors.Is(err, operation.ErrNotFound) {
		s.linkGone(w, r)
		return
	}
	if err != nil {
		s.internal(w, "open a share link", err)
		return
	}
	o := out.(window.ShareLinkOpening)
	if o.Once {
		// A session of reading: the pass ends on the server 4 hours after
		// the click, and the cookie with the browser.
		s.setSessionCookie(w, passCookie+strconv.FormatInt(o.WindowID, 10), o.Pass)
	} else {
		s.setCookie(w, passCookie+strconv.FormatInt(o.WindowID, 10), o.Pass, o.ExpiresAt)
	}
	http.Redirect(w, r, "/"+l.Window.Room+"/"+l.Window.Name+"/", http.StatusSeeOther)
}

// findLink returns the share link of the request, while it opens its Window
// to someone new. A reader who opens the Window already goes to it, without
// spending an opening; anyone else gets the page of a link that opens
// nothing when it no longer opens. It answers the request itself, and
// returns false, when it does not return a link.
func (s *Server) findLink(w http.ResponseWriter, r *http.Request) (window.Link, bool) {
	w.Header().Set("Referrer-Policy", "no-referrer")
	if !s.linkVisits.allow(clientKey(r)) {
		http.Error(w, "Too many requests: wait a minute, then try again.", http.StatusTooManyRequests)
		return window.Link{}, false
	}
	c := siteOf(r).cahut
	now := time.Now()
	l, err := window.FindLink(r.Context(), c.DB, r.PathValue("token"), now)
	if s.withdrawn(w, r, err, account.SessionOwner{}, true) {
		return window.Link{}, false
	}
	if errors.Is(err, window.ErrNoLink) {
		s.linkGone(w, r)
		return window.Link{}, false
	}
	if err != nil {
		s.internal(w, "find a share link", err)
		return window.Link{}, false
	}
	reader, _ := s.readerOf(r, now)
	how, err := l.Window.Opens(r.Context(), c.DB, reader.UserID, s.passOf(r, l.Window.ID, now))
	if err != nil {
		s.internal(w, "check the access of a Window", err)
		return window.Link{}, false
	}
	if how != window.Closed {
		http.Redirect(w, r, "/"+l.Window.Room+"/"+l.Window.Name+"/", http.StatusSeeOther)
		return window.Link{}, false
	}
	if !l.Stands {
		s.linkGone(w, r)
		return window.Link{}, false
	}
	return l, true
}

// linkGone answers with the page of a share link that opens nothing,
// without saying why.
func (s *Server) linkGone(w http.ResponseWriter, r *http.Request) {
	p := s.page(r, "", "Share link")
	p.SignIn = ""
	s.render(w, r, http.StatusNotFound, "linkgone", p)
}
