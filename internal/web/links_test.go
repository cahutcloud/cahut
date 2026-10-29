package web

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
	"cahut.cloud/internal/window"
)

// share makes a share link to the Window ref for actor, with the input
// args besides the Window, and returns it.
func (f *fixture) share(t *testing.T, actor operation.Actor, ref, args string) window.ShareLinkMade {
	t.Helper()
	if args != "" {
		args = ", " + args
	}
	return f.call(t, "localhost", actor, "create_share_link", `{"window": "`+ref+`"`+args+`}`).(window.ShareLinkMade)
}

// openShared clicks the button of the page of the share link at path, as
// a reader with cookies, and returns the response.
func (f *fixture) openShared(t *testing.T, path string, cookies ...*http.Cookie) response {
	t.Helper()
	resp, _ := f.do(t, "localhost", http.MethodPost, path, cookies, nil)
	return resp
}

// pass returns the pass that a response sets for the Window id.
func pass(t *testing.T, resp response, id int64) *http.Cookie {
	t.Helper()
	for _, c := range resp.Cookies() {
		if c.Name == "__Host-cahut-key-"+strconv.FormatInt(id, 10) {
			return &http.Cookie{Name: c.Name, Value: c.Value} //nolint:gosec // a cookie that the test sends back
		}
	}
	t.Fatalf("no pass in %v", resp.Header)
	return nil
}

// links returns the share links of the Window ref, as get_window gives them
// to actor, and how many stand.
func (f *fixture) links(t *testing.T, actor operation.Actor, ref string) ([]window.ShareLink, int) {
	t.Helper()
	d := f.call(t, "localhost", actor, "get_window", `{"window": "`+ref+`"}`).(window.Detail)
	return d.Links, d.ShareLinks
}

func TestShareLinks(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	c := f.open(t, "localhost")
	olivier, camille := f.person(t, "localhost", "olivier"), f.person(t, "localhost", "camille")
	deck := publishDeck(t, f, olivier)
	const path = "/olivier/q3-review/"
	var windowID int64
	if err := c.DB.Read.QueryRow("SELECT id FROM windows WHERE name = 'q3-review'").Scan(&windowID); err != nil {
		t.Fatal(err)
	}

	// An AI makes a link to a Window of its Room, without knocking first,
	// and gets it once.
	ai := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: olivier.UserID, Handle: "olivier", TokenID: 1, Client: "Claude Code"}
	made := f.share(t, ai, "olivier/q3-review", `"max_openings": 2, "label": "For Acme"`)
	token, ok := strings.CutPrefix(made.Link, "https://localhost/-/s/")
	if !ok || len(token) != 52 || made.Window != "olivier/q3-review" || made.AI != "Claude Code" || made.By != "olivier" ||
		made.MaxOpenings != 2 || !made.Stands || !made.ExpiresAt.Equal(made.CreatedAt.Add(window.LinkTTL)) {
		t.Fatalf("create_share_link = %+v", made)
	}
	link := "/-/s/" + token
	list, standing := f.links(t, olivier, "olivier/q3-review")
	if standing != 1 || len(list) != 1 || list[0].ID != made.ID || list[0].Label != "For Acme" || list[0].Openings != 0 {
		t.Fatalf("get_window gives the links %+v, %d standing", list, standing)
	}
	if _, err := c.Registry.Invoke(ctx, "create_share_link", camille, []byte(`{"window": "olivier/q3-review"}`)); err == nil {
		t.Error("create_share_link by someone outside the Room: no error")
	}

	// Every visit gets the page of the link, with the tags of a preview:
	// it opens nothing and counts nothing.
	for _, ua := range []string{chrome, slackbot, "curl/8.7.1"} {
		resp, b := f.fetch(t, "localhost", link, ua, nil, nil)
		body := string(b)
		if resp.StatusCode != http.StatusOK || !strings.Contains(body, `<meta property="og:title" content="Q3 review">`) ||
			!strings.Contains(body, `<meta property="og:url" content="https://localhost`+link+`">`) ||
			!strings.Contains(body, `content="Opens with this link until `) || !strings.Contains(body, "Open Q3 review") ||
			!strings.Contains(body, "for 2 more readers") || !strings.Contains(body, `<form method="post">`) {
			t.Errorf("GET the link as %q = %d %s", ua, resp.StatusCode, body)
		}
		// The card is that of the Window, which shows no title while it opens
		// to its Room.
		if !strings.Contains(body, `<meta property="og:image" content="https://localhost/-/card/`+deck.ID+`.png?v=`) {
			t.Errorf("the page of the link names no card: %s", body)
		}
		if strings.Contains(body, "For Acme") || strings.Contains(body, "Sign in") {
			t.Errorf("the page of the link shows its label or a sign-in: %s", body)
		}
		if resp.Header.Get("Referrer-Policy") != "no-referrer" || len(resp.Cookies()) != 0 {
			t.Errorf("GET the link sets %v", resp.Header)
		}
	}
	if list, _ := f.links(t, olivier, "olivier/q3-review"); list[0].Openings != 0 {
		t.Errorf("visits counted %d openings", list[0].Openings)
	}

	// The click spends an opening, and gives a pass that opens the Window,
	// its frame and its files.
	resp := f.openShared(t, link)
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != path {
		t.Fatalf("POST the link = %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	ana := pass(t, resp, windowID)
	src := f.frame(t, path, ana)
	if resp, body := f.do(t, "localhost", http.MethodGet, src+"css/deck.css", nil, nil); resp.StatusCode != http.StatusOK || !strings.Contains(body, "teal") {
		t.Errorf("GET %scss/deck.css with the pass of a link = %d %q", src, resp.StatusCode, body)
	}
	if list, standing := f.links(t, olivier, "olivier/q3-review"); list[0].Openings != 1 || list[0].OpenedAt == nil || standing != 1 {
		t.Errorf("after an opening, get_window gives %+v, %d standing", list[0], standing)
	}
	// It opens the Window, not the page of its Room.
	if resp, _ := f.do(t, "localhost", http.MethodGet, "/olivier/", []*http.Cookie{ana}, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("the Room for the reader of a link = %d, want 404", resp.StatusCode)
	}

	// A reader who opens the Window already goes straight to it: the reader
	// of the link, and a member of its Room.
	for who, cookie := range map[string]*http.Cookie{"ana": ana, "olivier": f.signIn(t, "localhost", "olivier")} {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			resp, _ := f.do(t, "localhost", method, link, []*http.Cookie{cookie}, nil)
			if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != path || len(resp.Cookies()) != 0 {
				t.Errorf("%s %s by %s = %d %q", method, link, who, resp.StatusCode, resp.Header.Get("Location"))
			}
		}
	}

	// The second reader spends the last opening; the next one gets nothing,
	// and those who opened it keep reading.
	bob := pass(t, f.openShared(t, link), windowID)
	if resp, body := f.fetch(t, "localhost", link, chrome, nil, nil); resp.StatusCode != http.StatusNotFound ||
		!strings.Contains(string(body), "This link opens nothing") || strings.Contains(string(body), "Q3") {
		t.Errorf("GET a spent link = %d %s", resp.StatusCode, body)
	}
	if resp := f.openShared(t, link); resp.StatusCode != http.StatusNotFound || len(resp.Cookies()) != 0 {
		t.Errorf("POST a spent link = %d", resp.StatusCode)
	}
	f.frame(t, path, ana)
	f.frame(t, path, bob)
	if _, standing := f.links(t, olivier, "olivier/q3-review"); standing != 0 {
		t.Errorf("a spent link stands: %d", standing)
	}

	// The Logbook records each opening in the Room, without the reader.
	entries, err := logbook.List(ctx, c.DB.Read, 50)
	if err != nil {
		t.Fatal(err)
	}
	var openings []string
	for _, e := range entries {
		switch e.Operation {
		case "open_share_link":
			if e.Outcome == logbook.OK {
				openings = append(openings, e.Detail)
			}
			if e.UserID != 0 || e.Address != "" || e.RoomID == 0 || e.Target != "olivier/q3-review" || e.Via != "web" {
				t.Errorf("the Logbook keeps the opening %+v", e)
			}
		case "create_share_link":
			if e.Outcome == logbook.OK && e.Detail != "link 1, until "+made.ExpiresAt.Format(time.RFC3339)+", 2 openings at most" || strings.Contains(e.Detail, token) {
				t.Errorf("the Logbook keeps the link as %q", e.Detail)
			}
		}
	}
	if strings.Join(openings, "; ") != "link 1, opening 2 of 2; link 1, opening 1 of 2" {
		t.Errorf("the Logbook keeps the openings %q", openings)
	}

	// A new Key ends the passes of the Key, not those of links.
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key"}`)
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "key", "new_key": true}`)
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "room"}`)
	f.frame(t, path, ana)

	// Taken back, the link closes the Window to its readers, at the next
	// file too.
	revoked := f.call(t, "localhost", ai, "revoke_share_link", `{"link": `+strconv.FormatInt(made.ID, 10)+`}`).(window.ShareLink)
	if revoked.RevokedAt == nil || revoked.Stands {
		t.Errorf("revoke_share_link = %+v", revoked)
	}
	for _, p := range []string{path, src} {
		if resp, _ := f.do(t, "localhost", http.MethodGet, p, []*http.Cookie{ana}, nil); resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s after the link is taken back = %d, want 404", p, resp.StatusCode)
		}
	}
	if _, err := c.Registry.Invoke(ctx, "revoke_share_link", olivier, []byte(`{"link": 1}`)); !refused(err, "share_link_ended") {
		t.Errorf("revoke_share_link twice: %v", err)
	}
	if _, err := c.Registry.Invoke(ctx, "revoke_share_link", camille, []byte(`{"link": 1}`)); err == nil {
		t.Error("revoke_share_link by someone outside the Room: no error")
	}

	// A link that ends closes the Window to its readers.
	later := f.share(t, olivier, "olivier/q3-review", "")
	carl := pass(t, f.openShared(t, strings.TrimPrefix(later.Link, "https://localhost")), windowID)
	past := store.FormatTime(time.Now().Add(-time.Minute))
	for _, q := range []string{"UPDATE share_links SET expires_at = ? WHERE id = ?", "UPDATE window_passes SET expires_at = ? WHERE link_id = ?"} {
		if _, err := c.DB.Write.Exec(q, past, later.ID); err != nil {
			t.Fatal(err)
		}
	}
	if resp, _ := f.fetch(t, "localhost", strings.TrimPrefix(later.Link, "https://localhost"), chrome, nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET an ended link = %d, want 404", resp.StatusCode)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, path, []*http.Cookie{carl}, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("the Window for the reader of an ended link = %d, want 404", resp.StatusCode)
	}

	// In the trash, a link opens nothing; restored, it opens again.
	again := strings.TrimPrefix(f.share(t, olivier, "olivier/q3-review", "").Link, "https://localhost")
	f.call(t, "localhost", olivier, "trash_window", `{"window": "olivier/q3-review"}`)
	if resp, _ := f.fetch(t, "localhost", again, chrome, nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET the link of a Window in the trash = %d, want 404", resp.StatusCode)
	}
	if _, err := c.Registry.Invoke(ctx, "create_share_link", olivier, []byte(`{"window": "olivier/q3-review"}`)); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("create_share_link to a Window in the trash: %v", err)
	}
	f.call(t, "localhost", olivier, "restore_window", `{"window": "olivier/q3-review"}`)
	if resp, _ := f.fetch(t, "localhost", again, chrome, nil, nil); resp.StatusCode != http.StatusOK {
		t.Errorf("GET the link of a Window restored = %d, want 200", resp.StatusCode)
	}
	if resp, _ := f.fetch(t, "localhost", "/-/s/nothing", chrome, nil, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET a link that names nothing = %d, want 404", resp.StatusCode)
	}

	// Purged, the Window takes its links and their passes with it.
	f.call(t, "localhost", olivier, "trash_window", `{"window": "olivier/q3-review"}`)
	f.call(t, "localhost", olivier, "purge_window", `{"window": "olivier/q3-review"}`)
	var left int
	if err := c.DB.Read.QueryRow("SELECT (SELECT count(*) FROM share_links) + (SELECT count(*) FROM window_passes)").Scan(&left); err != nil || left != 0 {
		t.Errorf("after a purge, %d links and passes are left: %v", left, err)
	}
}

func TestShareLinkBounds(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	c := f.open(t, "localhost")
	olivier := f.person(t, "localhost", "olivier")
	publishDeck(t, f, olivier)
	create := func(args string) error {
		_, err := c.Registry.Invoke(ctx, "create_share_link", olivier, []byte(`{"window": "olivier/q3-review", `+args+`}`))
		return err
	}
	at := func(d time.Duration) string {
		return `"ends_at": "` + time.Now().Add(d).UTC().Format(time.RFC3339) + `"`
	}
	for args, code := range map[string]string{
		at(-time.Minute):                              "share_link_end_past",
		at(31 * 24 * time.Hour):                       "share_link_too_long",
		`"max_openings": 1001`:                        "share_link_openings",
		`"max_openings": -1`:                          "share_link_openings",
		`"label": "For\nAcme"`:                        "share_link_label",
		`"label": "` + strings.Repeat("a", 101) + `"`: "share_link_label",
	} {
		if err := create(args); !refused(err, code) {
			t.Errorf("create_share_link with %s: %v, want %s", args, err, code)
		}
	}
	if err := create(at(30*24*time.Hour) + `, "max_openings": 1000, "label": "` + strings.Repeat("é", 100) + `"`); err != nil {
		t.Errorf("create_share_link at the bounds: %v", err)
	}

	// A Window holds 20 links that stand.
	for range window.MaxLinks - 1 {
		if err := create(`"label": "more"`); err != nil {
			t.Fatal(err)
		}
	}
	if err := create(`"label": "one more"`); !refused(err, "too_many_share_links") {
		t.Errorf("create_share_link past 20 links: %v", err)
	}

	// A person makes 100 links a day, their AIs included.
	if _, err := c.DB.Write.Exec("UPDATE share_links SET revoked_at = created_at"); err != nil {
		t.Fatal(err)
	}
	for range window.MaxLinksPerDay - window.MaxLinks {
		if _, err := c.DB.Write.Exec(`INSERT INTO share_links (window_id, hash, created_by, created_at, expires_at, revoked_at)
			SELECT id, randomblob(32), created_by, ?, ?, ? FROM windows`, store.FormatTime(time.Now()), store.FormatTime(time.Now().Add(time.Hour)), store.FormatTime(time.Now())); err != nil {
			t.Fatal(err)
		}
	}
	if err := create(`"label": "tomorrow"`); !refused(err, "too_many_share_links_today") {
		t.Errorf("create_share_link past 100 links a day: %v", err)
	}

	// A public Window takes no link: its permanent address opens it.
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/q3-review", "access": "public"}`)
	if err := create(`"label": "public"`); !refused(err, "share_link_public") {
		t.Errorf("create_share_link to a public Window: %v", err)
	}
}

// refused reports whether err is the refusal code.
func refused(err error, code string) bool {
	var r *operation.Refusal
	return errors.As(err, &r) && r.Code == code
}
