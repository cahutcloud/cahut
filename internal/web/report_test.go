package web

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"

	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/moderation"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
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

// serverAdmin activates an administrator of the server, and returns their
// session in the console of localhost.
func serverAdmin(t *testing.T, f *fixture, handle string) string {
	t.Helper()
	a, err := operation.Do(context.Background(), f.srv.Registry, server.CreateAdmin, admin, server.CreateAdminInput{Handle: handle, Name: handle})
	if err != nil {
		t.Fatal(err)
	}
	r := call(t, f, f.ts, http.MethodPost, "localhost", "/console/api/server/activate_account", "",
		`{"code": "`+a.Activation.Code+`", "password": "`+password+`"}`)
	if r.status != http.StatusOK {
		t.Fatalf("activate_account of the server: %d %s", r.status, r.raw)
	}
	return r.body["token"].(string)
}

// A moderator opens the Window that a report names, whatever its access,
// in a frame bound to their session, while the report waits; the members
// of its Room read each opening in the Logbook. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.
func TestOpenAReportedWindow(t *testing.T) {
	f := setup(t)
	console := f.ts
	olivier := f.person(t, "localhost", "olivier")
	publishPage(t, f, "Draft", "room")
	publishPage(t, f, "Other", "room")
	if status, _ := f.sendReport(t, "https://localhost/olivier/draft/", "illegal", "It shows what the law forbids.", ""); status != http.StatusOK {
		t.Fatalf("report = %d", status)
	}
	ana := activate(t, f, console, "ana", true)
	camille := activate(t, f, console, "camille", false)
	root := serverAdmin(t, f, "root")

	// A member cannot open it, nor can the command line, which has no
	// console.
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/preview_reported_window", camille, `{"report": 1}`); r.status != http.StatusForbidden {
		t.Errorf("preview_reported_window as a member: %d %s", r.status, r.raw)
	}
	if _, err := f.open(t, "localhost").Registry.Invoke(context.Background(), "preview_reported_window", admin, []byte(`{"report": 1}`)); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("preview_reported_window from the command line: %v", err)
	}

	// An administrator of the Cahut who is not in its Room opens it, and so
	// does an administrator of the server, each in a frame of their own.
	frames := map[string]string{}
	for who, token := range map[string]string{"ana": ana, "root": root} {
		r := call(t, f, console, http.MethodPost, "localhost", "/console/api/preview_reported_window", token, `{"report": 1}`)
		frame, _ := r.body["frame"].(string)
		if r.status != http.StatusOK || r.body["window"] != "olivier/draft" || !strings.HasPrefix(frame, "/-/frame/") {
			t.Fatalf("preview_reported_window as %s: %d %s", who, r.status, r.raw)
		}
		if resp, body := f.do(t, "localhost", http.MethodGet, frame, nil, nil); resp.StatusCode != http.StatusOK || !strings.Contains(body, "<h1>Draft</h1>") {
			t.Errorf("the frame of %s = %d %q", who, resp.StatusCode, body)
		}
		frames[who] = frame
	}
	// The frame opens that Window only.
	other := strings.Replace(frames["ana"], "/-/frame/", "/-/frame/x", 1)
	if resp, _ := f.do(t, "localhost", http.MethodGet, other, nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("a forged frame = %d", resp.StatusCode)
	}

	// Olivier reads in the Logbook of his Room that both opened it.
	var opened []string
	for _, e := range f.call(t, "localhost", olivier, "list_logbook", `{}`).(cahut.LogbookPage).Entries {
		if e.Operation == "preview_reported_window" && e.Target == "olivier/draft" && e.Outcome == logbook.OK {
			opened = append(opened, e.UserHandle+" "+e.Detail)
		}
	}
	if len(opened) != 2 || !slices.Contains(opened, "ana report 1") || !slices.Contains(opened, "root report 1") {
		t.Errorf("olivier reads the openings %q", opened)
	}

	// A moderator who no longer administers the Cahut loses the frame.
	activate(t, f, console, "bob", true)
	f.call(t, "localhost", admin, "set_user_role", `{"user": "ana", "role": "member"}`)
	if resp, _ := f.do(t, "localhost", http.MethodGet, frames["ana"], nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("the frame of a former administrator = %d", resp.StatusCode)
	}

	// Once the report is decided, the frames open nothing, and no new one
	// is given.
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/dismiss_report", root, `{"report": 1}`); r.status != http.StatusOK || r.body["by_server"] != true {
		t.Fatalf("dismiss_report as an administrator of the server: %d %s", r.status, r.raw)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, frames["root"], nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("the frame of a decided report = %d", resp.StatusCode)
	}
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/preview_reported_window", root, `{"report": 1}`); r.status != http.StatusBadRequest || r.body["code"] != "report_decided" {
		t.Errorf("preview_reported_window of a decided report: %d %s", r.status, r.raw)
	}
}
