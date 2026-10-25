package settings

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"cahut.cloud/internal/operation"
)

// LegalPages are the addresses of the legal pages of a Cahut, which its host
// publishes elsewhere. Each is optional.
type LegalPages struct {
	LegalNotice string `json:"legal_notice,omitempty" jsonschema:"the address of the legal notice: who publishes the service and who hosts it"`
	Terms       string `json:"terms,omitempty" jsonschema:"the address of the terms of use"`
	Privacy     string `json:"privacy,omitempty" jsonschema:"the address of the privacy policy"`
	Support     string `json:"support,omitempty" jsonschema:"the address of the page where people get help"`
}

// LegalLinks are the links to the legal pages of a Cahut, which the
// console and the pages for readers show at their foot, with their
// addresses in French where they differ. See
// docs/adr/0042-the-legal-pages-of-a-cahut.md.
type LegalLinks struct {
	LegalPages
	FR *LegalPages `json:"fr,omitempty" jsonschema:"the addresses of the same pages in French, where they differ: the console links them when it speaks French. Each needs the address of its page for every other language"`
}

// link is a link of LegalLinks: its page, its language, empty for every
// language without an address of its own, and its address.
type link struct {
	page, lang string
	url        *string
}

// name is the name of the link in the input of set_legal_links, such as
// terms or fr.terms.
func (k link) name() string {
	if k.lang == "" {
		return k.page
	}
	return k.lang + "." + k.page
}

func (p *LegalPages) links(lang string) []link {
	return []link{
		{"legal_notice", lang, &p.LegalNotice},
		{"terms", lang, &p.Terms},
		{"privacy", lang, &p.Privacy},
		{"support", lang, &p.Support},
	}
}

// links returns every link of l, those in French included, in the order
// that pages show them. It gives l an FR to point into.
func (l *LegalLinks) links() []link {
	if l.FR == nil {
		l.FR = &LegalPages{}
	}
	return append(l.LegalPages.links(""), l.FR.links("fr")...)
}

// lister runs a query that returns rows, in a transaction or not.
type lister interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Legal returns the links to the legal pages of the Cahut.
func Legal(ctx context.Context, q lister) (LegalLinks, error) {
	rows, err := q.QueryContext(ctx, "SELECT page, lang, url FROM legal_links")
	if err != nil {
		return LegalLinks{}, err
	}
	defer rows.Close()
	var l LegalLinks
	all := l.links()
	for rows.Next() {
		var page, lang, address string
		if err := rows.Scan(&page, &lang, &address); err != nil {
			return LegalLinks{}, err
		}
		for _, k := range all {
			if k.page == page && k.lang == lang {
				*k.url = address
			}
		}
	}
	if *l.FR == (LegalPages{}) {
		l.FR = nil
	}
	return l, rows.Err()
}

// In returns the address of each page in the language lang, its own or
// else the address for every other language, in the order that pages show
// them. A page without an address is left out.
func (l LegalLinks) In(lang string) []Link {
	var own LegalPages // the addresses of the pages in lang
	if lang == "fr" && l.FR != nil {
		own = *l.FR
	}
	owned := own.links(lang)
	var out []Link
	for i, k := range l.LegalPages.links("") {
		if address := cmp.Or(*owned[i].url, *k.url); address != "" {
			out = append(out, Link{Page: k.page, URL: address})
		}
	}
	return out
}

// Link is the address of a legal page, in one language.
type Link struct {
	Page string // legal_notice, terms, privacy or support
	URL  string
}

// SetLegalLinksInput is the input of set_legal_links.
type SetLegalLinksInput struct {
	LegalLinks
}

// LegalLink names a link of set_legal_links, as its input names it, such
// as terms or fr.terms.
type LegalLink struct {
	Link string `json:"link"`
}

// LinkInFrenchOnly names an address in French whose page has no address
// for the other languages.
type LinkInFrenchOnly struct {
	Link string `json:"link"` // such as fr.terms
	Page string `json:"page"` // such as terms
}

var (
	badLink = operation.NewRefusalOf[LegalLink]("legal_link_invalid", operation.ErrInvalid,
		"{link}: give the address of a page, which starts with https://, such as https://example.org/terms, or leave it out to remove the link")
	frenchOnly = operation.NewRefusalOf[LinkInFrenchOnly]("legal_link_in_french_only", operation.ErrInvalid,
		"{link}: give {page} too, the address of the page for every other language; the address in French takes its place in French only")
)

// maxLink is the longest address of a legal page, in bytes.
const maxLink = 2000

// SetLegalLinks sets the links to the legal pages of the Cahut, for the
// administrators of the server.
var SetLegalLinks = &operation.Op[SetLegalLinksInput, LegalLinks]{
	Name:  "set_legal_links",
	Title: "Set the links to the legal pages",
	Description: "Set the links to the legal pages of this Cahut, which its host publishes elsewhere: its legal notice, terms, " +
		"privacy policy and support. Each is the address of a page, which starts with https://; fr gives another address for a page in French, " +
		"where it differs, which the console links when it speaks French. The console and the pages for readers link them at their foot. " +
		"Give every link: a link left out is removed. For the administrators of the server.",
	Level: operation.ServerAdmin,
	Run: func(ctx context.Context, c *operation.Call, in SetLegalLinksInput) (LegalLinks, error) {
		c.Log("legal links", "")
		l := in.LegalLinks
		all := l.links()
		for _, k := range all {
			*k.url = strings.TrimSpace(*k.url)
			if *k.url != "" && !validLink(*k.url) {
				return LegalLinks{}, badLink(LegalLink{Link: k.name()})
			}
		}
		others := l.LegalPages.links("")
		for i, k := range l.FR.links("fr") {
			if *k.url != "" && *others[i].url == "" {
				return LegalLinks{}, frenchOnly(LinkInFrenchOnly{Link: k.name(), Page: k.page})
			}
		}
		before, err := Legal(ctx, c.Tx)
		if err != nil {
			return LegalLinks{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM legal_links"); err != nil {
			return LegalLinks{}, err
		}
		for _, k := range all {
			if *k.url == "" {
				continue
			}
			_, err := c.Tx.ExecContext(ctx, "INSERT INTO legal_links (page, lang, url) VALUES (?, ?, ?)", k.page, k.lang, *k.url)
			if err != nil {
				return LegalLinks{}, err
			}
		}
		c.Log("legal links", linkChanges(before, l))
		return Legal(ctx, c.Tx)
	},
}

// validLink reports whether s is the address of a page that a Cahut may
// link: an absolute address in HTTPS, with a host and without a password.
func validLink(s string) bool {
	if len(s) > maxLink || strings.ContainsFunc(s, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
		return false
	}
	u, err := url.Parse(s)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil
}

// linkChanges describes, for the Logbook, how the links went from before to
// after, such as "terms none → https://example.org/terms".
func linkChanges(before, after LegalLinks) string {
	b, a := before.links(), after.links()
	var out []string
	for i := range b {
		was, is := cmp.Or(*b[i].url, "none"), cmp.Or(*a[i].url, "none")
		if was != is {
			out = append(out, fmt.Sprintf("%s %s → %s", b[i].name(), was, is))
		}
	}
	if len(out) == 0 {
		return "unchanged"
	}
	return strings.Join(out, ", ")
}
