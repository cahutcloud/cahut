package web

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"cahut.cloud/internal/room"
)

// publishPage publishes a page called title in the personal Room of
// olivier, with its access.
func publishPage(t *testing.T, f *fixture, title, access string) {
	t.Helper()
	olivier := f.person(t, "localhost", "olivier")
	f.call(t, "localhost", olivier, "publish_window", `{"title": "`+title+`", "kind": "document", "declaration": {"scripts": false},
		"files": [{"path": "index.html", "content": "<h1>`+title+`</h1>"}]}`)
	if access != "room" {
		name := strings.ToLower(title)
		f.call(t, "localhost", olivier, "set_window_access", `{"window": "olivier/`+name+`", "access": "`+access+`"}`)
	}
}

var previewSrc = regexp.MustCompile(`<iframe class="card__preview" data-src="([^"]+)"`)

// previews returns the addresses of the previews on the page at path, for
// a reader with cookies.
func (f *fixture) previews(t *testing.T, path string, cookies ...*http.Cookie) (string, []string) {
	t.Helper()
	resp, body := f.do(t, "localhost", http.MethodGet, path, cookies, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s = %d %q", path, resp.StatusCode, body)
	}
	var srcs []string
	for _, m := range previewSrc.FindAllStringSubmatch(body, -1) {
		srcs = append(srcs, strings.ReplaceAll(m[1], "&amp;", "&"))
	}
	return body, srcs
}

// The page of a Room lists, to each reader, the Windows that they open,
// each with a preview that opens it alone. See
// docs/adr/0032-the-page-of-a-room.md.
func TestRoomPage(t *testing.T) {
	f := setup(t)
	f.person(t, "localhost", "camille")
	publishPage(t, f, "Notes", "room")
	publishPage(t, f, "Roadmap", "internal")
	publishPage(t, f, "Launch", "public")
	readOlivier, readCamille := f.signIn(t, "localhost", "olivier"), f.signIn(t, "localhost", "camille")

	// Anyone: the public Window, previewed at its own address.
	body, srcs := f.previews(t, "/olivier/")
	if !strings.Contains(body, `href="/olivier/launch/"`) || strings.Contains(body, "Roadmap") || strings.Contains(body, "Notes") ||
		!strings.Contains(body, "1 Window") || !strings.Contains(body, `c-badge--public`) || len(srcs) != 1 || srcs[0] != "/olivier/launch/" {
		t.Errorf("the Room for anyone = %q, previews %q", body, srcs)
	}
	if !strings.Contains(body, `href="/-/sign-in?to=%2Folivier%2F"`) {
		t.Errorf("the Room does not offer to sign in: %q", body)
	}

	// Everyone signed in: the internal Window too, in a frame of its own.
	body, srcs = f.previews(t, "/olivier/", readCamille)
	if !strings.Contains(body, "Roadmap") || strings.Contains(body, "Notes") || !strings.Contains(body, "2 Windows") || len(srcs) != 2 {
		t.Fatalf("the Room for camille = %q, previews %q", body, srcs)
	}
	for _, src := range srcs {
		resp, page := f.do(t, "localhost", http.MethodGet, src, nil, nil)
		if resp.StatusCode != http.StatusOK || !strings.HasPrefix(page, "<h1>") {
			t.Errorf("the preview %s = %d %q", src, resp.StatusCode, page)
		}
	}

	// Its members: every Window, Room only included.
	body, srcs = f.previews(t, "/olivier/", readOlivier)
	if !strings.Contains(body, "Notes") || !strings.Contains(body, "3 Windows") || !strings.Contains(body, "c-badge--room") || len(srcs) != 3 {
		t.Errorf("the Room for olivier = %q, previews %q", body, srcs)
	}
	if !strings.Contains(body, `<script src="/-/pages.js" defer></script>`) {
		t.Errorf("the Room loads no script for its previews: %q", body)
	}
	if resp, js := f.get(t, "localhost", http.MethodGet, "/-/pages.js"); resp.StatusCode != http.StatusOK || !strings.Contains(js, "card__preview") {
		t.Errorf("GET /-/pages.js = %d %q", resp.StatusCode, js)
	}

	// A Room without Windows says so to its members, and to nobody else.
	f.call(t, "localhost", f.person(t, "localhost", "olivier"), "create_room", `{"title": "Empty"}`)
	if body, _ := f.previews(t, "/empty/", readOlivier); !strings.Contains(body, "No Window here yet") {
		t.Errorf("an empty Room = %q", body)
	}
	if resp, _ := f.do(t, "localhost", http.MethodGet, "/empty/", []*http.Cookie{readCamille}, nil); resp.StatusCode != http.StatusNotFound {
		t.Errorf("an empty Room for camille = %d, want 404", resp.StatusCode)
	}
}

// A Room shows its home Window at its address to the readers who open it,
// and the list of the Windows they open to the others.
func TestRoomHome(t *testing.T) {
	f := setup(t)
	olivier, camille := f.person(t, "localhost", "olivier"), f.person(t, "localhost", "camille")
	publishPage(t, f, "Welcome", "room")
	publishPage(t, f, "Launch", "public")
	readOlivier := f.signIn(t, "localhost", "olivier")

	d := f.call(t, "localhost", olivier, "set_room_home", `{"room": "olivier", "window": "welcome"}`).(room.Detail)
	if d.Home != "welcome" {
		t.Errorf("set_room_home = %+v", d)
	}
	// Its members open it at the address of the Room, in a frame.
	src := f.frame(t, "/olivier/", readOlivier)
	if resp, body := f.do(t, "localhost", http.MethodGet, src, nil, nil); resp.StatusCode != http.StatusOK || body != "<h1>Welcome</h1>" {
		t.Errorf("the home of the Room = %d %q", resp.StatusCode, body)
	}
	// Anyone else gets the Windows they open.
	if body, srcs := f.previews(t, "/olivier/"); strings.Contains(body, "Welcome") || len(srcs) != 1 {
		t.Errorf("the Room for anyone, with a home of the Room only = %q", body)
	}

	// A public home opens to anyone, at its own address, with the query.
	f.call(t, "localhost", olivier, "set_room_home", `{"room": "olivier", "window": "launch"}`)
	resp, body := f.do(t, "localhost", http.MethodGet, "/olivier/?slide=2", nil, nil)
	if resp.StatusCode != http.StatusOK || !strings.Contains(body, `<iframe src="/olivier/launch/?slide=2" title="Launch"`) {
		t.Errorf("a public home = %d %q", resp.StatusCode, body)
	}

	// A home in the trash leaves the list until it is restored.
	f.call(t, "localhost", olivier, "trash_window", `{"window": "olivier/launch"}`)
	if d := f.call(t, "localhost", olivier, "get_room", `{"room": "olivier"}`).(room.Detail); d.Home != "" {
		t.Errorf("the home in the trash = %q", d.Home)
	}
	if body, _ := f.previews(t, "/olivier/", readOlivier); !strings.Contains(body, "Welcome") {
		t.Errorf("the Room with its home in the trash = %q", body)
	}
	f.call(t, "localhost", olivier, "restore_window", `{"window": "olivier/launch"}`)
	if d := f.call(t, "localhost", olivier, "get_room", `{"room": "olivier"}`).(room.Detail); d.Home != "launch" {
		t.Errorf("the home restored = %q", d.Home)
	}

	// The home is a Window of the Room, out of the trash, chosen by its owners.
	f.call(t, "localhost", camille, "publish_window", `{"title": "Elsewhere", "kind": "document", "declaration": {"scripts": false},
		"files": [{"path": "index.html", "content": "x"}]}`)
	for _, tt := range []struct{ who, args string }{
		{"olivier", `{"room": "olivier", "window": "elsewhere"}`},
		{"olivier", `{"room": "olivier", "window": "nope"}`},
		{"camille", `{"room": "olivier", "window": "welcome"}`},
	} {
		actor := olivier
		if tt.who == "camille" {
			actor = camille
		}
		if _, err := f.open(t, "localhost").Registry.Invoke(t.Context(), "set_room_home", actor, []byte(tt.args)); err == nil {
			t.Errorf("set_room_home %s by %s: no error", tt.args, tt.who)
		}
	}

	// Back to the list; and a purged home leaves it too.
	f.call(t, "localhost", olivier, "set_room_home", `{"room": "olivier"}`)
	if body, _ := f.previews(t, "/olivier/", readOlivier); !strings.Contains(body, "Welcome") || !strings.Contains(body, "Launch") {
		t.Errorf("the Room back to its list = %q", body)
	}
	f.call(t, "localhost", olivier, "set_room_home", `{"room": "olivier", "window": "launch"}`)
	f.call(t, "localhost", olivier, "trash_window", `{"window": "olivier/launch"}`)
	f.call(t, "localhost", olivier, "purge_window", `{"window": "olivier/launch"}`)
	var home any
	if err := f.open(t, "localhost").DB.Read.QueryRow("SELECT home_window_id FROM rooms WHERE name = 'olivier'").Scan(&home); err != nil || home != nil {
		t.Errorf("the home of the Room after its purge = %v, %v", home, err)
	}
}

// A new handle renames the personal Room: its Windows answer at the new
// address only.
func TestHandleMovesWindows(t *testing.T) {
	f := setup(t)
	publishPage(t, f, "Notes", "public")
	if resp, body := f.do(t, "localhost", http.MethodGet, "/olivier/notes/", nil, nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /olivier/notes/ = %d %q", resp.StatusCode, body)
	}
	f.call(t, "localhost", f.person(t, "localhost", "olivier"), "change_handle", `{"handle": "odurand"}`)
	for path, want := range map[string]int{"/odurand/notes/": http.StatusOK, "/olivier/notes/": http.StatusNotFound, "/olivier/": http.StatusNotFound} {
		if resp, _ := f.do(t, "localhost", http.MethodGet, path, nil, nil); resp.StatusCode != want {
			t.Errorf("GET %s = %d, want %d", path, resp.StatusCode, want)
		}
	}
}
