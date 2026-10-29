package web

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/card"
	"cahut.cloud/internal/window"
)

// CardPath is where a Cahut serves the card of each Window, which the
// previews of its links show: CardPath + <id> + ".png".
const CardPath = "/-/card/"

// previewRobots are words of the User-Agent of the robots that fetch a link
// to draw its preview, in lowercase. Search engines are not among them:
// they follow the permanent address of a Window as people do. See
// docs/adr/0051-sharing-a-window-and-its-preview.md.
var previewRobots = []string{
	"slackbot",            // Slack: Slackbot-LinkExpanding
	"twitterbot",          // X, and iMessage
	"facebookexternalhit", // Facebook, Messenger and Instagram, and iMessage
	"facebot",
	"linkedinbot",
	"discordbot",
	"telegrambot",
	"whatsapp",         // WhatsApp, and Signal
	"skypeuripreview",  // Teams and Skype
	"microsoftpreview", // Teams and Outlook
	"mastodon",
	"cardyb", // Bluesky
	"pinterestbot",
	"redditbot",
	"embedly",
	"iframely",
	"zulipurlpreview",
	"mattermost",
}

// previewRobot says whether r comes from a robot that draws the preview of
// a link.
func previewRobot(r *http.Request) bool {
	ua := strings.ToLower(r.UserAgent())
	for _, word := range previewRobots {
		if strings.Contains(ua, word) {
			return true
		}
	}
	return false
}

// cardTitle returns the title that the card and the preview of win show:
// the title of a public Window or of a Window with a Key, whose page of a
// Key already shows it, and none for the others.
func cardTitle(win window.Served) string {
	if win.Access == window.AccessPublic || win.Access == window.AccessKey {
		return win.Title
	}
	return ""
}

// previewPage is the page that a robot of previews gets at the permanent
// address of a Window.
type previewPage struct {
	page
	Description string
	URL         string // the permanent address
	Image       string // the address of the card
	ImageAlt    string
}

// servePreview answers a robot that previews the permanent address of win
// with a page of Cahut, whose tags say what that address says to a
// stranger, without a session or a Key: the title and the description of a
// public Window, the title of a Window with a Key, and of the others only
// the Cahut they are in. It opens nothing.
func (s *Server) servePreview(w http.ResponseWriter, r *http.Request, c *cahut.Cahut, win window.Served) {
	base := strings.TrimSuffix(s.cfg.BaseURL(c.Domain), "/")
	title := cardTitle(win)
	p := previewPage{URL: base + window.PermanentPrefix + win.UID}
	switch win.Access {
	case window.AccessPublic:
		p.Description = win.Description
		p.ImageAlt = title + ", a Window at " + c.Domain
	case window.AccessKey:
		p.Description = "Opens with its Key."
		p.ImageAlt = title + ", a Window at " + c.Domain
	default:
		title = "A Window at " + c.Domain
		p.Description = "Sign in to open it."
		p.ImageAlt = title
	}
	p.page = s.page(r, "", title)
	p.Image = base + CardPath + win.UID + ".png?v=" + card.Tag(cardTitle(win), c.Domain)
	s.render(w, r, http.StatusOK, "preview", p)
}

// serveCard serves the card of a Window, which the previews of its links
// show, with the title that they show. Anyone who has its id fetches it,
// as the robots do.
func (s *Server) serveCard(w http.ResponseWriter, r *http.Request) {
	c := siteOf(r).cahut
	id, ok := strings.CutSuffix(r.PathValue("file"), ".png")
	if !ok || !address.ValidID(id) {
		http.NotFound(w, r)
		return
	}
	win, err := window.LookupUID(r.Context(), c.DB, id, time.Now())
	if err != nil && !isNotFound(err) {
		s.internal(w, "find a Window", err)
		return
	}
	if err != nil {
		http.NotFound(w, r)
		return
	}
	b, tag, err := s.cards.Get(cardTitle(win), c.Domain)
	if err != nil {
		s.internal(w, "draw the card of a Window", err)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "image/png")
	h.Set("ETag", `"`+tag+`"`)
	// A new title gives a new address to the card in its previews; an hour
	// keeps an old one where the address stays, as after a change of access.
	h.Set("Cache-Control", "public, max-age=3600")
	h.Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(b))
}
