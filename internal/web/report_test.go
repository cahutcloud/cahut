package web

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/moderation"
	"cahut.cloud/internal/window"
)

// sendReport posts the form of a report, and returns the status and the
// page that answers it.
func (f *fixture) sendReport(t *testing.T, window, category, explanation, email string) (int, string) {
	t.Helper()
	form := url.Values{"window": {window}, "category": {category}, "explanation": {explanation}, "good_faith": {"yes"}}
	if email != "" {
		form.Set("email", email)
		form.Set("name", "Jane Reader")
	}
	resp, body := f.do(t, "localhost", http.MethodPost, ReportPath, nil, form)
	return resp.StatusCode, body
}

// Anyone reports a Window, without an account, by any address that its
// readers meet; its moderators read the report in the console and decide.
// See docs/adr/0054-reporting-and-withdrawing-a-window.md.
func TestReportAWindow(t *testing.T) {
	f := setup(t)
	console := f.ts
	olivier := f.person(t, "localhost", "olivier")
	publishPage(t, f, "Launch", "public")
	publishPage(t, f, "Review", "key")
	publishPage(t, f, "Draft", "room")
	link := f.share(t, olivier, "olivier/draft", "")
	uid := f.call(t, "localhost", olivier, "get_window", `{"window": "olivier/review"}`).(window.Detail).ID

	// The page asks for the address, what is wrong, an explanation, and
	// whether the report is made in good faith, and sends no referrer.
	resp, body := f.do(t, "localhost", http.MethodGet, ReportPath+"?window="+url.QueryEscape("https://localhost/olivier/launch/"), nil, nil)
	if resp.StatusCode != http.StatusOK || resp.Header.Get("Referrer-Policy") != "no-referrer" ||
		!strings.Contains(body, `value="https://localhost/olivier/launch/"`) || !strings.Contains(body, `value="phishing"`) ||
		!strings.Contains(body, `name="good_faith"`) || strings.Contains(body, "Report a Window</a>") {
		t.Fatalf("GET %s = %d %v %q", ReportPath, resp.StatusCode, resp.Header, body)
	}

	// Each address of a Window names it: its address and a file under it,
	// its permanent address, a share link, and a Room that shows it.
	f.call(t, "localhost", olivier, "set_room_home", `{"room": "olivier", "window": "launch"}`)
	for _, address := range []string{
		"https://localhost/olivier/launch/", "/olivier/launch/img/logo.png", "olivier/launch", "https://localhost/-/w/" + uid,
		link.Link, "https://localhost/olivier/",
	} {
		status, body := f.sendReport(t, address, "phishing", "It asks for the password of my bank.", "")
		if status != http.StatusOK || !strings.Contains(body, "Report received") || strings.Contains(body, "Launch") || strings.Contains(body, "Draft") {
			t.Errorf("report %s = %d %q", address, status, body)
		}
	}

	// A report is refused, with what to fix and what was typed, when it
	// names no Window of the Cahut, or lacks what it needs.
	for _, tt := range []struct{ window, category, explanation, want string }{
		{"https://localhost/olivier/nothing/", "phishing", "It asks for my password.", "This address names no Window of this Cahut"},
		{"https://beta.localhost/olivier/launch/", "phishing", "It asks for my password.", "This address names no Window of this Cahut"},
		{"https://localhost/-/s/nolink", "phishing", "It asks for my password.", "This address names no Window of this Cahut"},
		{"https://localhost/olivier/launch/", "", "It asks for my password.", "Say what is wrong with the Window"},
		{"https://localhost/olivier/launch/", "phishing", "Bad", "Explain what is wrong with the Window in 10 to 2000 characters"},
	} {
		status, body := f.sendReport(t, tt.window, tt.category, tt.explanation, "")
		if status != http.StatusUnprocessableEntity || !strings.Contains(body, tt.want) || !strings.Contains(body, "It asks for my password.") && tt.explanation != "Bad" {
			t.Errorf("report %+v = %d %q", tt, status, body)
		}
	}
	if status, body := f.sendReport(t, "https://localhost/olivier/launch/", "malware", "It downloads a program.", "not an address"); status != http.StatusUnprocessableEntity ||
		!strings.Contains(body, "Give one e-mail address alone") {
		t.Errorf("report with a wrong address = %d %q", status, body)
	}
	if status, _ := f.sendReport(t, "https://localhost/olivier/launch/", "malware", "It downloads a program.", "reader@example.org"); status != http.StatusOK {
		t.Errorf("report with an e-mail address = %d", status)
	}

	// The Logbook names neither whoever reports, nor their address, nor what
	// they wrote, nor the share link they gave.
	c := f.open(t, "localhost")
	entries, err := logbook.List(context.Background(), c.DB.Read, 100)
	if err != nil {
		t.Fatal(err)
	}
	reported := 0
	for _, e := range entries {
		if e.Operation != "report_window" {
			continue
		}
		if e.Outcome == logbook.OK {
			reported++
		}
		raw := e.Target + " " + e.Detail + " " + e.Error + " " + e.Address + " " + e.UserHandle
		for _, secret := range []string{"password", "Jane Reader", "reader@example.org", "not an address", strings.TrimPrefix(link.Link, "https://localhost/-/s/"), "127.0.0.1"} {
			if strings.Contains(raw, secret) {
				t.Errorf("the Logbook keeps %q: %+v", secret, e)
			}
		}
		if e.RoomID != 0 {
			t.Errorf("a report is logged in a Room: %+v", e)
		}
	}
	if reported != 7 {
		t.Errorf("%d reports logged, want 7", reported)
	}

	// The moderators read the reports in the console: a member cannot.
	ana := activate(t, f, console, "ana", true)
	camille := activate(t, f, console, "camille", false)
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/list_reports", camille, `{}`); r.status != http.StatusForbidden {
		t.Errorf("list_reports as a member: %d %s", r.status, r.raw)
	}
	r := call(t, f, console, http.MethodPost, "localhost", "/console/api/list_reports", ana, `{}`)
	reports, _ := r.body["reports"].([]any)
	if r.status != http.StatusOK || r.body["waiting"] != float64(7) || len(reports) != 7 {
		t.Fatalf("list_reports: %d %s", r.status, r.raw)
	}
	last := reports[6].(map[string]any)
	if w := last["window"].(map[string]any); w["room"] != "olivier" || w["name"] != "launch" || last["email"] != "reader@example.org" ||
		last["name"] != "Jane Reader" || last["category"] != "malware" {
		t.Errorf("the last report = %v", last)
	}
	if w := reports[4].(map[string]any)["window"].(map[string]any); w["name"] != "draft" || w["access"] != "room" {
		t.Errorf("the report of a share link names %v", w)
	}

	// Dismissing a report decides it, and forgets who made it.
	id := strconv.FormatInt(int64(last["id"].(float64)), 10)
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/dismiss_report", ana, `{"report": `+id+`, "note": "A real installer."}`)
	if r.status != http.StatusOK || r.body["decision"] != moderation.Dismissed || r.body["decided_by"] != "ana" || r.body["email"] != nil || r.body["name"] != nil ||
		r.body["note"] != "A real installer." {
		t.Errorf("dismiss_report: %d %s", r.status, r.raw)
	}
	checkRefusals(t, f, console, ana, []refusalCase{{
		name: "a report decided already", path: "/console/api/dismiss_report", body: `{"report": ` + id + `}`,
		status: 400, code: "report_decided", values: map[string]any{"report": float64(last["id"].(float64))},
		error: "invalid input: report " + id + " is decided already: list_reports gives those that wait",
	}})
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/list_reports", ana, `{}`)
	if reports, _ := r.body["reports"].([]any); r.body["waiting"] != float64(6) || len(reports) != 7 || reports[6].(map[string]any)["decision"] != "dismissed" {
		t.Errorf("list_reports after a decision: %s", r.raw)
	}

	// No AI is offered the operations of moderation.
	for _, name := range []string{"report_window", "list_reports", "dismiss_report"} {
		info, _ := c.Registry.Lookup(name)
		if info.ForAIs(true, false) || info.Services {
			t.Errorf("%s is offered to AIs or services", name)
		}
	}

	// Reports are limited by address.
	limited := false
	for range ReportBurst + 1 {
		if status, _ := f.sendReport(t, "https://localhost/olivier/launch/", "other", "Something else is wrong.", ""); status == http.StatusTooManyRequests {
			limited = true
		}
	}
	if !limited {
		t.Error("reports are not limited by address")
	}
}
