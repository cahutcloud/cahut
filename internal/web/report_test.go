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

// The moderators withdraw a Window that a report names, or a public
// Window, with a reason: it keeps everything, and serves nothing to its
// readers, who learn that it was withdrawn when they would open it; its
// members still preview it in the console, and change nothing in it until
// it is reinstated, which only the administrators of the server do for a
// Window that they withdrew. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.
func TestWithdrawAWindow(t *testing.T) {
	f := setup(t)
	console := f.ts
	token, olivier := f.consoleSession(t, "olivier")
	publishPage(t, f, "Phish", "public")
	publishPage(t, f, "Draft", "room")
	publishPage(t, f, "Review", "key")
	publishPage(t, f, "Notes", "room")
	link := f.share(t, olivier, "olivier/draft", "")
	uid := f.call(t, "localhost", olivier, "get_window", `{"window": "olivier/phish"}`).(window.Detail).ID
	for _, address := range []string{link.Link, "https://localhost/olivier/review/"} {
		if status, _ := f.sendReport(t, address, "phishing", "It asks for the password of my bank.", ""); status != http.StatusOK {
			t.Fatalf("report %s = %d", address, status)
		}
	}
	ana := activate(t, f, console, "ana", true)
	root := serverAdmin(t, f, "root")
	moderate := func(token, op, window, reason string) apiResponse {
		t.Helper()
		body := `{"window": "` + window + `"}`
		if reason != "" {
			body = `{"window": "` + window + `", "reason": "` + reason + `"}`
		}
		return call(t, f, console, http.MethodPost, "localhost", "/console/api/"+op, token, body)
	}

	// The moderators withdraw a Window that a report names, or a public
	// one, and not another, nor without a reason.
	if r := moderate(ana, "withdraw_window", "olivier/notes", "Spam"); r.status != http.StatusForbidden {
		t.Errorf("withdraw a Window open to its Room, unreported: %d %s", r.status, r.raw)
	}
	if r := moderate(ana, "withdraw_window", "olivier/draft", " "); r.status != http.StatusBadRequest || r.body["code"] != "withdrawal_reason_invalid" {
		t.Errorf("withdraw without a reason: %d %s", r.status, r.raw)
	}
	r := moderate(ana, "withdraw_window", "olivier/draft", "It imitates the sign-in page of a bank.")
	if decided, _ := r.body["decided"].([]any); r.status != http.StatusOK || len(decided) != 1 || r.body["withdrawn"] == nil {
		t.Fatalf("withdraw_window: %d %s", r.status, r.raw)
	}
	for _, w := range []string{"olivier/review", uid} {
		if r := moderate(ana, "withdraw_window", w, "Phishing."); r.status != http.StatusOK {
			t.Fatalf("withdraw_window %s: %d %s", w, r.status, r.raw)
		}
	}
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/list_reports", ana, `{}`)
	if reports, _ := r.body["reports"].([]any); r.body["waiting"] != float64(0) || len(reports) != 2 ||
		reports[0].(map[string]any)["decision"] != moderation.Withdrawn || reports[0].(map[string]any)["window"].(map[string]any)["withdrawn"] != true {
		t.Errorf("list_reports after the withdrawals: %s", r.raw)
	}

	// Readers who would open them learn that they were withdrawn, at every
	// address: the address, the permanent address, a share link, the page
	// of a Key and the form of its Key. Others meet what is not found.
	reader := f.signIn(t, "localhost", "olivier")
	for _, tt := range []struct {
		method, path string
		cookies      []*http.Cookie
		form         url.Values
		want         int
	}{
		{http.MethodGet, "/olivier/phish/", nil, nil, http.StatusGone},
		{http.MethodGet, "/olivier/phish/style.css", nil, nil, http.StatusGone},
		{http.MethodGet, "/-/w/" + uid, nil, nil, http.StatusGone},
		{http.MethodGet, strings.TrimPrefix(link.Link, "https://localhost"), nil, nil, http.StatusGone},
		{http.MethodPost, strings.TrimPrefix(link.Link, "https://localhost"), nil, nil, http.StatusGone},
		{http.MethodGet, "/olivier/review/", nil, nil, http.StatusGone},
		{http.MethodPost, "/-/key", nil, url.Values{"window": {"olivier/review"}, "key": {"anything"}}, http.StatusGone},
		{http.MethodGet, "/olivier/draft/", nil, nil, http.StatusNotFound},
		{http.MethodGet, "/olivier/draft/", []*http.Cookie{reader}, nil, http.StatusGone},
	} {
		resp, body := f.do(t, "localhost", tt.method, tt.path, tt.cookies, tt.form)
		if resp.StatusCode != tt.want || resp.StatusCode == http.StatusGone && !strings.Contains(body, "This Window was withdrawn") ||
			strings.Contains(body, "imitates") || strings.Contains(body, "<h1>") {
			t.Errorf("%s %s = %d %q", tt.method, tt.path, resp.StatusCode, body)
		}
	}
	// The page of the Room no longer lists them.
	if _, body := f.do(t, "localhost", http.MethodGet, "/olivier/", []*http.Cookie{reader}, nil); !strings.Contains(body, "Notes") ||
		strings.Contains(body, "Phish") || strings.Contains(body, "Draft") {
		t.Errorf("the page of the Room lists withdrawn Windows: %q", body)
	}

	// Its members read why, change nothing in it, and still preview it.
	d := f.call(t, "localhost", olivier, "get_window", `{"window": "olivier/draft"}`).(window.Detail)
	if d.Withdrawn == nil || d.Withdrawn.Reason != "It imitates the sign-in page of a bank." || d.Withdrawn.ByServer {
		t.Errorf("get_window gives %+v", d.Withdrawn)
	}
	for op, args := range map[string]string{
		"update_window":     `{"window": "olivier/draft", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "<p>new</p>"}]}`,
		"set_window_access": `{"window": "olivier/draft", "access": "public"}`,
		"create_share_link": `{"window": "olivier/draft"}`,
		"retitle_window":    `{"window": "olivier/draft", "title": "Other"}`,
		"rename_window":     `{"window": "olivier/draft", "name": "other"}`,
		"rollback_window":   `{"window": "olivier/draft"}`,
		"set_room_home":     `{"room": "olivier", "window": "draft"}`,
	} {
		_, err := f.open(t, "localhost").Registry.Invoke(context.Background(), op, olivier, []byte(args))
		if !errors.Is(err, operation.ErrForbidden) || !strings.Contains(err.Error(), "olivier/draft was withdrawn by the moderators of this Cahut: It imitates") {
			t.Errorf("%s of a withdrawn Window: %v", op, err)
		}
	}
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/update_window", token,
		`{"window": "olivier/draft", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "x"}]}`)
	if r.status != http.StatusForbidden || r.body["code"] != "window_withdrawn" {
		t.Errorf("update_window in the console: %d %s", r.status, r.raw)
	}
	previews := call(t, f, console, http.MethodPost, "localhost", "/console/api/preview_windows", token, `{"windows": ["olivier/draft"]}`)
	list, _ := previews.body["previews"].([]any)
	if len(list) != 1 {
		t.Fatalf("preview_windows: %s", previews.raw)
	}
	if resp, body := f.do(t, "localhost", http.MethodGet, list[0].(map[string]any)["frame"].(string), nil, nil); resp.StatusCode != http.StatusOK || !strings.Contains(body, "<h1>Draft</h1>") {
		t.Errorf("the preview of a withdrawn Window = %d %q", resp.StatusCode, body)
	}
	// It goes to the trash and comes back, still withdrawn.
	f.call(t, "localhost", olivier, "trash_window", `{"window": "olivier/draft"}`)
	f.call(t, "localhost", olivier, "restore_window", `{"window": "olivier/draft"}`)
	if resp, _ := f.do(t, "localhost", http.MethodGet, strings.TrimPrefix(link.Link, "https://localhost"), nil, nil); resp.StatusCode != http.StatusGone {
		t.Errorf("a withdrawn Window back from the trash = %d", resp.StatusCode)
	}

	// Reinstated, it opens as it did, its share link with it.
	if r := moderate(ana, "reinstate_window", "olivier/draft", ""); r.status != http.StatusOK || r.body["withdrawn"] != nil {
		t.Fatalf("reinstate_window: %d %s", r.status, r.raw)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, strings.TrimPrefix(link.Link, "https://localhost"), nil, nil); resp.StatusCode != http.StatusOK {
		t.Errorf("the share link of a reinstated Window = %d", resp.StatusCode)
	}

	// A Window that the administrators of the server withdrew is theirs to
	// reinstate, or to withdraw again with another reason.
	if r := moderate(root, "withdraw_window", "olivier/phish", "Phishing, reported by the bank."); r.status != http.StatusOK ||
		r.body["withdrawn"].(map[string]any)["by_server"] != true {
		t.Fatalf("withdraw_window again as an administrator of the server: %d %s", r.status, r.raw)
	}
	for op, reason := range map[string]string{"reinstate_window": "", "withdraw_window": "Not phishing."} {
		r := moderate(ana, op, "olivier/phish", reason)
		if r.status != http.StatusForbidden || r.body["code"] != "window_withdrawn_by_host" {
			t.Errorf("%s as an administrator of the Cahut: %d %s", op, r.status, r.raw)
		}
	}
	if r := moderate(root, "reinstate_window", "olivier/phish", ""); r.status != http.StatusOK {
		t.Errorf("reinstate_window as an administrator of the server: %d %s", r.status, r.raw)
	}
	if resp, body := f.do(t, "localhost", http.MethodGet, "/olivier/phish/", nil, nil); resp.StatusCode != http.StatusOK || !strings.Contains(body, "<h1>Phish</h1>") {
		t.Errorf("a reinstated public Window = %d %q", resp.StatusCode, body)
	}

	// Olivier reads in the Logbook of his Room who withdrew and why.
	var read []string
	for _, e := range f.call(t, "localhost", olivier, "list_logbook", `{}`).(cahut.LogbookPage).Entries {
		if (e.Operation == "withdraw_window" || e.Operation == "reinstate_window") && e.Outcome == logbook.OK {
			read = append(read, e.UserHandle+" "+e.Operation+" "+e.Target+": "+e.Detail)
		}
	}
	if !slices.Contains(read, "ana withdraw_window olivier/draft: reports 1: It imitates the sign-in page of a bank.") ||
		!slices.Contains(read, "root withdraw_window olivier/phish: Phishing, reported by the bank.") {
		t.Errorf("olivier reads %q", read)
	}

	// No AI is offered them.
	for _, name := range []string{"withdraw_window", "reinstate_window", "preview_reported_window"} {
		info, _ := f.open(t, "localhost").Registry.Lookup(name)
		if info.ForAIs(true, false) || info.Services {
			t.Errorf("%s is offered to AIs or services", name)
		}
	}
}

// With a relay, each report goes to the administrators of the Cahut who
// have an address and to its address of moderation, whoever reports gets a
// receipt and the decision, and the members of the Room hear that their
// Window was withdrawn, and why, and that it was reinstated. No e-mail
// holds what whoever reported wrote. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.
func TestModerationMail(t *testing.T) {
	f := setup(t)
	console := f.ts
	relay := f.relay(t)
	publishPage(t, f, "Phish", "public")
	c := f.open(t, "localhost")
	ana := activate(t, f, console, "ana", true)
	activate(t, f, console, "bob", true) // without an address
	for _, q := range []string{
		"UPDATE users SET email = 'olivier@example.org', name = 'Olivier' WHERE handle = 'olivier'",
		"UPDATE users SET email = 'ana@example.org', name = 'Ana' WHERE handle = 'ana'",
	} {
		if _, err := c.DB.Write.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	root := serverAdmin(t, f, "root")

	// Only the administrators of the server set the address of moderation.
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/set_moderation_address", ana, `{"email": "abuse@host.example"}`); r.status != http.StatusForbidden {
		t.Errorf("set_moderation_address as an administrator of the Cahut: %d %s", r.status, r.raw)
	}
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/set_moderation_address", root, `{"email": "not an address"}`); r.status != http.StatusBadRequest {
		t.Errorf("set_moderation_address with a wrong address: %d %s", r.status, r.raw)
	}
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/set_moderation_address", root, `{"email": "abuse@host.example"}`); r.status != http.StatusOK {
		t.Fatalf("set_moderation_address: %d %s", r.status, r.raw)
	}
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/get_cahut", ana, `{}`); r.body["moderation_email"] != "abuse@host.example" {
		t.Errorf("get_cahut for an administrator: %s", r.raw)
	}

	// A report goes to Ana and to the address of moderation, and its receipt
	// to whoever reported.
	if status, _ := f.sendReport(t, "https://localhost/olivier/phish/extra-words-of-a-spammer", "phishing", "Buy cheap watches at spam.example now.", "reader@example.org"); status != http.StatusOK {
		t.Fatalf("report = %d", status)
	}
	msgs := relay.Wait(t, 3)
	to := map[string]string{}
	for _, m := range msgs {
		to[m.To[0]] = body(m)
	}
	for _, who := range []string{"ana@example.org", "abuse@host.example"} {
		if b := to[who]; !strings.Contains(b, "Report 1 about a Window of localhost") || !strings.Contains(b, "https://localhost/olivier/phish/") ||
			!strings.Contains(b, "phishing or a scam") || !strings.Contains(b, "https://localhost/console/moderation") {
			t.Errorf("the report to %s = %q", who, b)
		}
	}
	if b := to["reader@example.org"]; !strings.Contains(b, "received your report 1 about the Window https://localhost/olivier/phish/") {
		t.Errorf("the receipt = %q", b)
	}
	for who, b := range to {
		if strings.Contains(b, "spam.example") || strings.Contains(b, "Jane Reader") || strings.Contains(b, "extra-words") {
			t.Errorf("the e-mail to %s holds what whoever reported wrote: %q", who, b)
		}
	}

	// Withdrawing tells the decision to whoever reported, and the reason to
	// the members of the Room.
	r := call(t, f, console, http.MethodPost, "localhost", "/console/api/withdraw_window", ana, `{"window": "olivier/phish", "reason": "It imitates a bank."}`)
	if r.status != http.StatusOK {
		t.Fatalf("withdraw_window: %d %s", r.status, r.raw)
	}
	msgs = relay.Wait(t, 5)[3:]
	to = map[string]string{}
	for _, m := range msgs {
		to[m.To[0]] = body(m)
	}
	if b := to["reader@example.org"]; !strings.Contains(b, "and withdrew the Window: it no longer opens") {
		t.Errorf("the decision = %q", b)
	}
	if b := to["olivier@example.org"]; !strings.Contains(b, "olivier/phish was withdrawn") || !strings.Contains(b, "It imitates a bank.") ||
		!strings.Contains(b, "The moderators of the Cahut at localhost withdrew the Window https://localhost/olivier/phish/") {
		t.Errorf("the withdrawal to olivier = %q", b)
	}

	// Reinstating tells them that it opens again.
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/reinstate_window", ana, `{"window": "olivier/phish"}`); r.status != http.StatusOK {
		t.Fatalf("reinstate_window: %d %s", r.status, r.raw)
	}
	if m := relay.Wait(t, 6)[5]; m.To[0] != "olivier@example.org" || !strings.Contains(body(m), "olivier/phish is open again") {
		t.Errorf("the reinstatement = %v %q", m.To, body(m))
	}
}

// The moderators list the public Windows of their Cahut, the latest made
// public first, and the list of the Cahuts of the server counts the
// reports that wait in each. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.
func TestPublicWindowsAndWaitingReports(t *testing.T) {
	f := setup(t)
	console := f.ts
	olivier := f.person(t, "localhost", "olivier")
	publishPage(t, f, "First", "public")
	publishPage(t, f, "Second", "room")
	publishPage(t, f, "Third", "public")
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/second", "access": "public"}`)
	ana := activate(t, f, console, "ana", true)
	root := serverAdmin(t, f, "root")
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/withdraw_window", ana, `{"window": "olivier/third", "reason": "Spam."}`); r.status != http.StatusOK {
		t.Fatalf("withdraw_window: %d %s", r.status, r.raw)
	}
	r := call(t, f, console, http.MethodPost, "localhost", "/console/api/list_public_windows", ana, `{}`)
	var names []string
	for _, w := range r.body["windows"].([]any) {
		names = append(names, w.(map[string]any)["name"].(string)+" by "+w.(map[string]any)["by"].(string))
	}
	if r.status != http.StatusOK || !slices.Equal(names, []string{"second by olivier", "first by olivier"}) {
		t.Errorf("list_public_windows: %d %q %s", r.status, names, r.raw)
	}
	camille := activate(t, f, console, "camille", false)
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/list_public_windows", camille, `{}`); r.status != http.StatusForbidden {
		t.Errorf("list_public_windows as a member: %d %s", r.status, r.raw)
	}

	for range 2 {
		if status, _ := f.sendReport(t, "https://localhost/olivier/first/", "other", "Something is wrong here.", ""); status != http.StatusOK {
			t.Fatalf("report = %d", status)
		}
	}
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/server/list_domains", root, `{}`)
	counts := map[string]float64{}
	for _, c := range r.body["cahuts"].([]any) {
		reports, _ := c.(map[string]any)["reports"].(float64)
		counts[c.(map[string]any)["domain"].(string)] = reports
	}
	if r.status != http.StatusOK || counts["localhost"] != 2 || counts["beta.localhost"] != 0 {
		t.Errorf("list_domains: %d %v %s", r.status, counts, r.raw)
	}
}
