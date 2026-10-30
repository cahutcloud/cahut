package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"cahut.cloud/internal/moderation"
	"cahut.cloud/internal/operation"
)

// Anyone reports a Window to the moderators of its Cahut, without an
// account, on a page of Cahut that runs no script, which the foot of every
// other page links. Cahut draws nothing in or around the Windows: the
// reader of a public Window types or pastes its address. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.

// ReportPath is the address of the page where anyone reports a Window.
const ReportPath = "/-/report"

// Reports are limited by address, as sign-in is, with a budget of their
// own: a burst of ReportBurst, then one every ReportEvery.
const (
	ReportBurst = 20
	ReportEvery = 3 * time.Minute
)

// reportPage is the page of a report: its form, with what the reader typed
// when the report is refused.
type reportPage struct {
	page
	Window, Category, Explanation, Name, Email string
	GoodFaith                                  bool
	Categories                                 []moderation.Category
	Min, Max                                   int
	Error                                      string
}

// reportLink returns the address of the page of a report, which fills in
// the address of a Window when one is given.
func reportLink(window string) string {
	if window == "" {
		return ReportPath
	}
	return ReportPath + "?window=" + url.QueryEscape(window)
}

// serveReport answers with the form of a report.
func (s *Server) serveReport(w http.ResponseWriter, r *http.Request) {
	s.askReport(w, r, reportPage{Window: r.URL.Query().Get("window")}, http.StatusOK)
}

func (s *Server) askReport(w http.ResponseWriter, r *http.Request, p reportPage, status int) {
	p.page = s.page(r, "", "Report a Window")
	// The page of a share link sends no referrer: neither does this one,
	// whose form may hold one.
	w.Header().Set("Referrer-Policy", "no-referrer")
	p.SignIn, p.Report = "", ""
	p.Categories = moderation.Categories
	p.Min, p.Max = moderation.MinExplanation, moderation.MaxExplanation
	s.render(w, r, status, "report", p)
}

// sendReport takes the form of a report, keeps the report, and says that
// it was received, and nothing about the Window.
func (s *Server) sendReport(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "The form could not be read.", http.StatusBadRequest)
		return
	}
	f := r.PostForm
	p := reportPage{
		Window:      f.Get("window"),
		Category:    f.Get("category"),
		Explanation: f.Get("explanation"),
		Name:        f.Get("name"),
		Email:       f.Get("email"),
		GoodFaith:   f.Get("good_faith") == "yes",
	}
	if !s.reports.allow(clientKey(r)) {
		p.Error = "Too many reports from your network: wait a few minutes, then try again."
		s.askReport(w, r, p, http.StatusTooManyRequests)
		return
	}
	c := siteOf(r).cahut
	in, err := json.Marshal(moderation.ReportInput{
		Window: p.Window, Category: p.Category, Explanation: p.Explanation, Name: p.Name, Email: p.Email, GoodFaith: p.GoodFaith,
	})
	if err != nil {
		s.internal(w, "report a Window", err)
		return
	}
	actor := operation.Actor{Kind: operation.Visitor, Via: "web", Domain: c.Domain, Address: clientKey(r)}
	_, err = c.Registry.Invoke(r.Context(), "report_window", actor, in)
	if errors.Is(err, operation.ErrInvalid) || errors.Is(err, operation.ErrNotFound) {
		p.Error = sentence(err)
		s.askReport(w, r, p, http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		s.internal(w, "report a Window", err)
		return
	}
	pg := s.page(r, "", "Report received")
	pg.SignIn, pg.Report = "", ""
	w.Header().Set("Referrer-Policy", "no-referrer")
	s.render(w, r, http.StatusOK, "reported", pg)
}

// sentence returns the message of a refusal as a sentence for a page:
// without the kind of error before it, with a capital and a full stop.
func sentence(err error) string {
	msg := err.Error()
	for _, kind := range []error{operation.ErrInvalid, operation.ErrNotFound} {
		msg = strings.TrimPrefix(msg, kind.Error()+": ")
	}
	if msg == "" {
		return msg
	}
	runes := []rune(msg)
	runes[0] = unicode.ToUpper(runes[0])
	msg = string(runes)
	if !strings.HasSuffix(msg, ".") {
		msg += "."
	}
	return msg
}
