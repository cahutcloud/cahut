package web

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/store"
	"cahut.cloud/internal/window"
)

// publishLetter publishes, for actor, a Window in the team Room studio
// for its author alone, as a draft or a Window that opens once.
func publishLetter(t *testing.T, f *fixture, actor operation.Actor) window.Window {
	t.Helper()
	return f.call(t, "localhost", actor, "publish_window", `{"title": "Offer", "kind": "document", "room": "studio", "access": "author",
		"declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "<h1>Offer</h1>"}]}`).(window.Window)
}

// A Window may open to its author alone, in a team Room too, and a share
// link that opens once gives it a single reader, for a session of reading.
// See docs/adr/0053-windows-for-their-author-alone.md.
func TestWindowForItsAuthorAlone(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	c := f.open(t, "localhost")
	olivier, camille := f.person(t, "localhost", "olivier"), f.person(t, "localhost", "camille")
	f.call(t, "localhost", olivier, "create_room", `{"title": "Studio"}`)
	f.call(t, "localhost", olivier, "add_room_member", `{"room": "studio", "user": "camille", "role": "owner"}`)
	ai := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: olivier.UserID, Handle: "olivier", TokenID: 1, Client: "Claude Code"}
	camilleAI := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: camille.UserID, Handle: "camille", TokenID: 2, Client: "Claude Code"}
	shown := f.call(t, "localhost", camille, "publish_window", `{"title": "Plan", "kind": "document", "room": "studio",
		"declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "<h1>Plan</h1>"}]}`).(window.Window)

	// The AI of the author publishes it for them alone, without knocking.
	letter := publishLetter(t, f, ai)
	if letter.Access != window.AccessAuthor || letter.CreatedBy != "olivier" {
		t.Fatalf("publish_window for its author = %+v", letter)
	}
	const path = "/studio/offer/"
	served, err := window.Lookup(ctx, c.DB, "studio", "offer", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	id := served.ID

	// Its author sees it as any Window of theirs.
	if l := f.call(t, "localhost", ai, "list_windows", `{"room": "studio"}`).(window.List); len(l.Windows) != 2 {
		t.Errorf("list_windows for the author = %+v", l.Windows)
	}
	if d := f.call(t, "localhost", olivier, "get_room", `{"room": "studio"}`).(room.Detail); d.Windows != 2 {
		t.Errorf("get_room for the author counts %d Windows", d.Windows)
	}
	readOlivier, readCamille := f.signIn(t, "localhost", "olivier"), f.signIn(t, "localhost", "camille")
	f.frame(t, path, readOlivier)
	if _, srcs := f.previews(t, "/studio/", readOlivier); len(srcs) != 2 {
		t.Errorf("the page of the Room for the author shows %d Windows", len(srcs))
	}

	// The other members of its Room, its owners included, and their AIs see
	// nothing of it.
	for _, who := range []operation.Actor{camille, camilleAI} {
		if l := f.call(t, "localhost", who, "list_windows", `{}`).(window.List); len(l.Windows) != 1 || l.Windows[0].ID != shown.ID {
			t.Errorf("list_windows for %s = %+v", who.Kind, l.Windows)
		}
		for _, ref := range []string{"studio/offer", letter.ID, letter.PermanentURL} {
			if _, err := c.Registry.Invoke(ctx, "get_window", who, []byte(`{"window": "`+ref+`"}`)); !isNotFound(err) {
				t.Errorf("get_window %s for %s: %v", ref, who.Kind, err)
			}
		}
		for op, args := range map[string]string{
			"trash_window":      `{"window": "studio/offer"}`,
			"create_share_link": `{"window": "studio/offer"}`,
			"retitle_window":    `{"window": "studio/offer", "title": "Mine"}`,
			"update_window":     `{"window": "studio/offer", "files": [{"path": "index.html", "content": "Mine"}], "declaration": {"scripts": false}}`,
		} {
			if _, err := c.Registry.Invoke(ctx, op, who, []byte(args)); !isNotFound(err) {
				t.Errorf("%s for %s: %v", op, who.Kind, err)
			}
		}
	}
	if p := f.call(t, "localhost", camille, "preview_windows", `{"windows": ["studio/offer", "studio/plan"]}`).(window.Previews); len(p.Previews) != 1 {
		t.Errorf("preview_windows for camille = %+v", p.Previews)
	}
	if d := f.call(t, "localhost", camille, "get_room", `{"room": "studio"}`).(room.Detail); d.Windows != 1 {
		t.Errorf("get_room for camille counts %d Windows", d.Windows)
	}
	if _, err := c.Registry.Invoke(ctx, "set_room_home", camille, []byte(`{"room": "studio", "window": "offer"}`)); !isNotFound(err) {
		t.Errorf("set_room_home to the Window of another, for themself: %v", err)
	}
	for _, p := range []string{path, window.PermanentPrefix + letter.ID} {
		if resp, body := f.do(t, "localhost", http.MethodGet, p, []*http.Cookie{readCamille}, nil); resp.StatusCode != http.StatusNotFound || strings.Contains(body, "Offer") {
			t.Errorf("GET %s by camille = %d", p, resp.StatusCode)
		}
	}
	if body, srcs := f.previews(t, "/studio/", readCamille); len(srcs) != 1 || strings.Contains(body, "Offer") {
		t.Errorf("the page of the Room for camille = %s", body)
	}

	// Only its author keeps a Window for themself; nobody is invited to it.
	if _, err := c.Registry.Invoke(ctx, "set_window_access", olivier, []byte(`{"window": "studio/plan", "access": "author"}`)); !refused(err, "window_author_only") {
		t.Errorf("set_window_access author by someone else than its author: %v", err)
	}
	if _, err := c.Registry.Invoke(ctx, "set_window_access", olivier, []byte(`{"window": "studio/offer", "access": "author", "ends_at": "2099-01-01T00:00:00Z"}`)); err == nil {
		t.Error("set_window_access author with an end: no error")
	}
	if _, err := c.Registry.Invoke(ctx, "invite_to_window", olivier, []byte(`{"window": "studio/offer", "to": ["camille"]}`)); !refused(err, "invitation_author_only") {
		t.Errorf("invite_to_window to a Window for its author alone: %v", err)
	}

	// A link that opens once: its page says so, and visits spend nothing.
	made := f.share(t, ai, "studio/offer", `"max_openings": 1`)
	link := strings.TrimPrefix(made.Link, "https://localhost")
	for _, ua := range []string{slackbot, chrome} {
		resp, b := f.fetch(t, "localhost", link, ua, nil, nil)
		if body := string(b); resp.StatusCode != http.StatusOK || !strings.Contains(body, `content="Opens once, until `) ||
			!strings.Contains(body, "This link opens it once") || !strings.Contains(body, "for 4 hours") {
			t.Errorf("GET the link as %q = %d %s", ua, resp.StatusCode, body)
		}
	}
	// The robots of previews see what a Window open to its Room shows at its
	// permanent address.
	if _, b := f.fetch(t, "localhost", window.PermanentPrefix+letter.ID, slackbot, nil, nil); !strings.Contains(string(b), `content="A Window at localhost"`) {
		t.Errorf("the preview of its permanent address = %s", b)
	}

	// The click gives a pass for a session of reading: a cookie without an
	// end, and 4 hours on the server.
	resp := f.openShared(t, link)
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != path {
		t.Fatalf("POST the link = %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	for _, cookie := range resp.Cookies() {
		if strings.HasPrefix(cookie.Name, "__Host-cahut-key-") && (cookie.MaxAge != 0 || !cookie.Expires.IsZero()) {
			t.Errorf("the pass of a link that opens once ends with %+v", cookie)
		}
	}
	ana := pass(t, resp, id)
	src := f.frame(t, path, ana)
	if resp, body := f.do(t, "localhost", http.MethodGet, src, nil, nil); resp.StatusCode != http.StatusOK || body != "<h1>Offer</h1>" {
		t.Errorf("the frame for the reader of the link = %d %q", resp.StatusCode, body)
	}
	var expires string
	if err := c.DB.Read.QueryRow("SELECT expires_at FROM window_passes WHERE link_id = ?", made.ID).Scan(&expires); err != nil {
		t.Fatal(err)
	}
	if at, err := store.ParseTime(expires); err != nil || time.Until(at) > window.OnceTTL || time.Until(at) < window.OnceTTL-time.Minute {
		t.Errorf("the pass ends at %s: %v", expires, err)
	}
	// Its reader comes back to the link and goes to the Window; nobody else
	// opens it.
	if resp, _ := f.do(t, "localhost", http.MethodGet, link, []*http.Cookie{ana}, nil); resp.StatusCode != http.StatusSeeOther {
		t.Errorf("GET the link by its reader = %d", resp.StatusCode)
	}
	if resp := f.openShared(t, link); resp.StatusCode != http.StatusNotFound || len(resp.Cookies()) != 0 {
		t.Errorf("POST a link that opened once = %d", resp.StatusCode)
	}
	// Once the pass ends, its reader gets nothing more.
	if _, err := c.DB.Write.Exec("UPDATE window_passes SET expires_at = ? WHERE link_id = ?", store.FormatTime(time.Now().Add(-time.Second)), made.ID); err != nil {
		t.Fatal(err)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, path, []*http.Cookie{ana}, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("the Window after the session of reading = %d, want 404", resp.StatusCode)
	}

	// The Window stays with its author, and the link shows opened.
	list, standing := f.links(t, olivier, "studio/offer")
	if standing != 0 || len(list) != 1 || list[0].Openings != 1 || list[0].OpenedAt == nil {
		t.Errorf("after the opening, get_window gives %+v, %d standing", list, standing)
	}
	f.frame(t, path, readOlivier)

	// The Logbook gives what happens to the Window to its author, not to the
	// Room; the administrators read it all. Camille reads her own calls,
	// which found no Window.
	about := func(entries []logbook.Entry) []string {
		var ops []string
		for _, e := range entries {
			if e.Target == "studio/offer" && e.UserID != camille.UserID {
				ops = append(ops, e.Operation)
			}
		}
		return ops
	}
	mine, err := logbook.ListOf(ctx, c.DB.Read, olivier.UserID, 100)
	if err != nil {
		t.Fatal(err)
	}
	if ops := strings.Join(about(mine), " "); !strings.Contains(ops, "open_share_link") || !strings.Contains(ops, "create_share_link") || !strings.Contains(ops, "publish_window") {
		t.Errorf("the author reads %q", ops)
	}
	theirs, err := logbook.ListOf(ctx, c.DB.Read, camille.UserID, 100)
	if err != nil {
		t.Fatal(err)
	}
	if ops := about(theirs); len(ops) != 0 {
		t.Errorf("camille reads %q about a Window she does not see", ops)
	}
	if all, err := logbook.List(ctx, c.DB.Read, 100); err != nil || !strings.Contains(strings.Join(about(all), " "), "open_share_link") {
		t.Errorf("the administrators read %q: %v", about(all), err)
	}

	// Its author shares it with the Room; the Room reads that entry.
	f.call(t, "localhost", ai, "set_window_access", `{"window": "studio/offer", "access": "room"}`)
	if l := f.call(t, "localhost", camille, "list_windows", `{"room": "studio"}`).(window.List); len(l.Windows) != 2 {
		t.Errorf("list_windows for camille once it opens to the Room = %d Windows", len(l.Windows))
	}
	theirs, _ = logbook.ListOf(ctx, c.DB.Read, camille.UserID, 100)
	if ops := about(theirs); len(ops) != 1 || ops[0] != "set_window_access" {
		t.Errorf("camille reads %q", ops)
	}
	f.call(t, "localhost", ai, "set_window_access", `{"window": "studio/offer", "access": "author"}`)
	theirs, _ = logbook.ListOf(ctx, c.DB.Read, camille.UserID, 100)
	if ops := about(theirs); len(ops) != 2 {
		t.Errorf("camille does not read that it was taken from the Room: %q", ops)
	}

	// An author who leaves the Room gives it their Windows back.
	f.call(t, "localhost", camille, "remove_room_member", `{"room": "studio", "user": "olivier"}`)
	if l := f.call(t, "localhost", camille, "list_windows", `{"room": "studio"}`).(window.List); len(l.Windows) != 2 || l.Windows[0].Access != window.AccessRoom && l.Windows[1].Access != window.AccessRoom {
		t.Errorf("list_windows once the author left = %+v", l.Windows)
	}
}

// A person deleted gives the Windows they kept for themselves back to
// their Room, without an author. See
// docs/adr/0053-windows-for-their-author-alone.md.
func TestDeletedAuthor(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	c := f.open(t, "localhost")
	olivier, camille := f.person(t, "localhost", "olivier"), f.person(t, "localhost", "camille")
	f.call(t, "localhost", camille, "create_room", `{"title": "Studio"}`)
	f.call(t, "localhost", camille, "add_room_member", `{"room": "studio", "user": "olivier"}`)
	publishLetter(t, f, olivier)

	f.call(t, "localhost", admin, "delete_user", `{"user": "olivier"}`)
	l := f.call(t, "localhost", camille, "list_windows", `{"room": "studio"}`).(window.List)
	if len(l.Windows) != 1 || l.Windows[0].Access != window.AccessRoom || l.Windows[0].CreatedBy != "" {
		t.Fatalf("list_windows once the author is deleted = %+v", l.Windows)
	}
	if _, err := c.Registry.Invoke(ctx, "set_window_access", camille, []byte(`{"window": "studio/offer", "access": "author"}`)); !refused(err, "window_without_author") {
		t.Errorf("set_window_access author without an author: %v", err)
	}
}
