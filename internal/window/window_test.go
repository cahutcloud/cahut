package window

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/address"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/store"
	"cahut.cloud/internal/windata"
)

type fixture struct {
	r       *operation.Registry
	db      *store.DB
	s       *Service
	olivier operation.Actor // Claude Code, acting for olivier
	camille operation.Actor
}

func setup(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	db, err := store.Open(ctx, filepath.Join(dir, "cahut.db"), store.CahutSchema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	r := operation.NewRegistry(db)
	account.Register(r)
	room.Register(r)
	data := windata.NewPool(windata.Options{})
	t.Cleanup(data.Close)
	s := NewService(filepath.Join(dir, "windows"), "https://cahut.test/", frameKey, data)
	s.Register(r)

	f := &fixture{r: r, db: db, s: s}
	admin := operation.Actor{Kind: operation.Admin, Via: "cli"}
	for _, handle := range []string{"olivier", "camille"} {
		u, err := operation.Do(ctx, r, account.CreateUser, admin, account.CreateUserInput{Handle: handle, Name: handle})
		if err != nil {
			t.Fatal(err)
		}
		tok, err := operation.Do(ctx, r, account.CreateToken, admin, account.CreateTokenInput{User: handle, Name: "Claude Code"})
		if err != nil {
			t.Fatal(err)
		}
		a := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: u.ID, Handle: handle, TokenID: tok.ID, Client: "Claude Code"}
		if handle == "olivier" {
			f.olivier = a
		} else {
			f.camille = a
		}
	}
	return f
}

func (f *fixture) invoke(t *testing.T, actor operation.Actor, name string, args string) (any, error) {
	t.Helper()
	return f.r.Invoke(context.Background(), name, actor, []byte(args))
}

const deck = `{
	"title": "Revue du 3e trimestre",
	"kind": "deck",
	"declaration": {"scripts": false}, "files": [
		{"path": "index.html", "content": "<h1>Q3</h1>"},
		{"path": "img/logo.png", "content": "iVBORw0KGgo=", "encoding": "base64"}
	]
}`

func TestPublishAndGet(t *testing.T) {
	f := setup(t)
	out, err := f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil {
		t.Fatalf("publish_window: %v", err)
	}
	w := out.(Window)
	if w.Name != "revue-du-3e-trimestre" || w.URL != "https://cahut.test/olivier/revue-du-3e-trimestre/" ||
		w.Kind != Deck || w.Version != 1 || w.FileCount != 2 ||
		w.Size != int64(len("<h1>Q3</h1>")+8) || w.Room != "olivier" || w.CreatedBy != "olivier" {
		t.Errorf("window = %+v", w)
	}
	id, version, err := current(context.Background(), f.db, "olivier", w.Name)
	if err != nil || version != 1 {
		t.Fatalf("Current = %d, %d, %v", id, version, err)
	}
	html, err := os.ReadFile(filepath.Join(f.s.Dir(id, 1), "index.html"))
	if err != nil || string(html) != "<h1>Q3</h1>" {
		t.Errorf("index.html on disk = %q, %v", html, err)
	}

	// The same title in the same Room gets the next free name.
	out, err = f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil || out.(Window).Name != "revue-du-3e-trimestre-2" {
		t.Errorf("second publication = %+v, %v", out, err)
	}
	if _, err := f.invoke(t, f.olivier, "get_window", `{"window": "olivier/revue-du-3e-trimestre-2"}`); err != nil {
		t.Errorf("get_window by room/name: %v", err)
	}

	out, err = f.invoke(t, f.olivier, "get_window", `{"window": "`+w.URL+`index.html", "include_content": true}`)
	if err != nil {
		t.Fatalf("get_window: %v", err)
	}
	d := out.(Detail)
	if d.Name != w.Name || len(d.Files) != 2 {
		t.Fatalf("detail = %+v", d)
	}
	logo, index := d.Files[0], d.Files[1]
	if index.Path != "index.html" || index.Type != "text/html; charset=utf-8" || index.Content != "<h1>Q3</h1>" || index.Encoding != "utf-8" {
		t.Errorf("index.html = %+v", index)
	}
	png, _ := base64.StdEncoding.DecodeString(logo.Content)
	if logo.Path != "img/logo.png" || logo.Encoding != "base64" || len(png) != 8 {
		t.Errorf("img/logo.png = %+v", logo)
	}

	entries, err := logbook.List(context.Background(), f.db.Read, 10)
	if err != nil {
		t.Fatal(err)
	}
	if entries[3].Operation != "publish_window" || entries[3].Target != "olivier/revue-du-3e-trimestre" ||
		entries[3].Outcome != logbook.OK || !strings.Contains(entries[3].Detail, "2 files") ||
		entries[0].Operation != "get_window" || entries[0].Target != "olivier/revue-du-3e-trimestre" {
		t.Errorf("logbook = %+v", entries[:4])
	}
}

func TestWindowsStayInTheirRoom(t *testing.T) {
	f := setup(t)
	out, err := f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil {
		t.Fatal(err)
	}
	url := out.(Window).URL

	if _, err := f.invoke(t, f.camille, "get_window", `{"window": "`+url+`"}`); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_window by another person: err = %v, want ErrNotFound", err)
	}
	if _, err := f.invoke(t, f.camille, "publish_window", strings.Replace(deck, `"kind"`, `"room": "olivier", "kind"`, 1)); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("publish_window in another person's Room: err = %v, want ErrNotFound", err)
	}
	out, err = f.invoke(t, f.camille, "list_windows", `{}`)
	if err != nil || len(out.(List).Windows) != 0 {
		t.Errorf("list_windows by another person = %+v, %v", out, err)
	}
	out, err = f.invoke(t, f.olivier, "list_windows", ``)
	if err != nil || len(out.(List).Windows) != 1 {
		t.Errorf("list_windows by the owner = %+v, %v", out, err)
	}

	// An administrator of the server manages people, not content; the
	// command line reaches everything.
	serverAdmin := operation.Actor{Kind: operation.Admin, Via: "api", Domain: "cahut.example.com", Handle: "root", Session: 1}
	out, err = f.invoke(t, serverAdmin, "list_windows", ``)
	if err != nil || len(out.(List).Windows) != 0 {
		t.Errorf("list_windows by an administrator of the server = %+v, %v", out, err)
	}
	if _, err := f.invoke(t, serverAdmin, "get_window", `{"window": "`+url+`"}`); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_window by an administrator of the server: err = %v, want ErrNotFound", err)
	}
	out, err = f.invoke(t, operation.Actor{Kind: operation.Admin, Via: "cli"}, "list_windows", ``)
	if err != nil || len(out.(List).Windows) != 1 {
		t.Errorf("list_windows by the command line = %+v, %v", out, err)
	}
}

func TestTeamRooms(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	if _, err := f.invoke(t, f.olivier, "create_room", `{"title": "Marketing"}`); err != nil {
		t.Fatalf("create_room: %v", err)
	}
	inMarketing := strings.Replace(deck, `"kind"`, `"room": "marketing", "kind"`, 1)
	if _, err := f.invoke(t, f.camille, "publish_window", inMarketing); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("publish_window in a Room one is not in: err = %v, want ErrNotFound", err)
	}

	// olivier, in the console, adds camille, who publishes there.
	owner := operation.Actor{Kind: operation.Person, Via: "api", UserID: f.olivier.UserID, Handle: "olivier", Session: 1}
	if _, err := operation.Do(ctx, f.r, room.AddMember, owner, room.MemberInput{Room: "marketing", User: "camille"}); err != nil {
		t.Fatal(err)
	}
	out, err := f.invoke(t, f.camille, "publish_window", inMarketing)
	if err != nil {
		t.Fatalf("publish_window in a team Room: %v", err)
	}
	w := out.(Window)
	if w.Room != "marketing" || w.URL != "https://cahut.test/marketing/revue-du-3e-trimestre/" || w.CreatedBy != "camille" {
		t.Errorf("window = %+v", w)
	}
	if _, err := f.invoke(t, f.olivier, "publish_window", deck); err != nil {
		t.Fatal(err)
	}
	for _, who := range []operation.Actor{f.olivier, f.camille} {
		want := 1
		out, err := f.invoke(t, who, "list_windows", `{"room": "marketing"}`)
		if err != nil || len(out.(List).Windows) != want {
			t.Errorf("list_windows of marketing by %s = %+v, %v; want %d", who.Handle, out, err, want)
		}
	}
	out, err = f.invoke(t, f.olivier, "list_windows", `{}`)
	if err != nil || len(out.(List).Windows) != 2 {
		t.Errorf("list_windows of all of olivier's Rooms = %+v, %v", out, err)
	}

	// Once out of the Room, camille no longer reaches its Windows.
	if _, err := operation.Do(ctx, f.r, room.RemoveMember, owner, room.RemoveMemberInput{Room: "marketing", User: "camille"}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.invoke(t, f.camille, "list_windows", `{"room": "marketing"}`); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("list_windows of a Room one left: err = %v, want ErrNotFound", err)
	}
	if _, err := f.invoke(t, f.camille, "get_window", `{"window": "`+w.URL+`"}`); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_window in a Room one left: err = %v, want ErrNotFound", err)
	}

	// The entries of the Room name it, for its members.
	entries, err := logbook.List(ctx, f.db.Read, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Operation == "publish_window" && e.Outcome == logbook.OK && e.Target == "marketing/revue-du-3e-trimestre" && e.RoomID == 0 {
			t.Errorf("a publication in marketing names no Room: %+v", e)
		}
	}
}

func TestPublishRefusesBadFiles(t *testing.T) {
	tests := map[string]string{
		"no files":        `[]`,
		"no index":        `[{"path": "page.html", "content": "x"}]`,
		"absolute path":   `[{"path": "index.html", "content": "x"}, {"path": "/etc/passwd", "content": "x"}]`,
		"parent path":     `[{"path": "index.html", "content": "x"}, {"path": "../x.html", "content": "x"}]`,
		"unclean path":    `[{"path": "./index.html", "content": "x"}]`,
		"hidden file":     `[{"path": "index.html", "content": "x"}, {"path": ".htaccess", "content": "x"}]`,
		"backslash":       `[{"path": "index.html", "content": "x"}, {"path": "img\\a.png", "content": "x"}]`,
		"duplicate":       `[{"path": "index.html", "content": "x"}, {"path": "INDEX.html", "content": "x"}]`,
		"file and dir":    `[{"path": "index.html", "content": "x"}, {"path": "img", "content": "x"}, {"path": "img/a.png", "content": "x"}]`,
		"bad base64":      `[{"path": "index.html", "content": "not base64!", "encoding": "base64"}]`,
		"bad encoding":    `[{"path": "index.html", "content": "x", "encoding": "latin-1"}]`,
		"control in path": `[{"path": "index.html", "content": "x"}, {"path": "a\u0007.html", "content": "x"}]`,
	}
	for name, files := range tests {
		t.Run(name, func(t *testing.T) {
			f := setup(t)
			_, err := f.invoke(t, f.olivier, "publish_window", `{"title": "T", "kind": "site", "declaration": {"scripts": false}, "files": `+files+`}`)
			if !errors.Is(err, operation.ErrInvalid) {
				t.Errorf("err = %v, want ErrInvalid", err)
			}
			if entries, _ := os.ReadDir(f.s.dir); len(entries) != 0 {
				t.Errorf("files left on disk: %v", entries)
			}
		})
	}
}

func TestPublishRefusesBadInput(t *testing.T) {
	f := setup(t)
	for _, args := range []string{
		`{"title": " ", "kind": "deck", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "x"}]}`,
		`{"title": "T", "kind": "slides", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "x"}]}`,
		`{"title": "T", "kind": "deck", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "x"}], "version": 2}`,
	} {
		if _, err := f.invoke(t, f.olivier, "publish_window", args); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", args, err)
		}
	}
	if _, err := f.invoke(t, f.olivier, "publish_window", `{"title": "T", "kind": "deck", "room": "team", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "x"}]}`); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("publish in an unknown Room: err = %v, want ErrNotFound", err)
	}
	if _, err := f.invoke(t, f.olivier, "get_window", `{"window": "Q3 review"}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("get_window without an address: err = %v, want ErrInvalid", err)
	}
	admin := operation.Actor{Kind: operation.Admin, Via: "cli"}
	if _, err := f.invoke(t, admin, "publish_window", deck); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("publish by the administrator: err = %v, want ErrInvalid", err)
	}
}

func TestPublishRemovesFilesWhenItFails(t *testing.T) {
	f := setup(t)
	// A duplicate title can never fail, so make the Logbook write fail
	// instead, after the files are written.
	if _, err := f.db.Write.ExecContext(context.Background(), "DROP TABLE logbook"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.invoke(t, f.olivier, "publish_window", deck); err == nil {
		t.Fatal("publish_window succeeded without a Logbook")
	}
	if entries, _ := os.ReadDir(f.s.dir); len(entries) != 0 {
		t.Errorf("files left on disk: %v", entries)
	}
	var n int
	if err := f.db.Read.QueryRowContext(context.Background(), "SELECT count(*) FROM windows").Scan(&n); err != nil || n != 0 {
		t.Errorf("windows = %d, %v; want 0", n, err)
	}
}

func TestParseRef(t *testing.T) {
	const id = "3f9k2m7q8x4r1t6v0b5n8c2d4g"
	tests := map[string]reference{
		"olivier/q3-review":                                     {room: "olivier", name: "q3-review"},
		"/olivier/q3-review/":                                   {room: "olivier", name: "q3-review"},
		"https://cahut.example.com/olivier/q3-review/":          {room: "olivier", name: "q3-review"},
		"https://cahut.example.com/olivier/q3-review/css/a.css": {room: "olivier", name: "q3-review"},
		"http://localhost:8080/marketing/q3-review":             {room: "marketing", name: "q3-review"},
		" marketing/q3-review ":                                 {room: "marketing", name: "q3-review"},
		id:                                                      {id: id},
		" 3F9K2M7Q8X4R1T6V0B5N8C2D4G ":                          {id: id},
		"/-/w/" + id:                                            {id: id},
		"https://cahut.example.com/-/w/" + id:                   {id: id},
		"https://cahut.example.com/-/w/" + id + "/img/logo.svg": {id: id},
	}
	for in, want := range tests {
		if got, ok := parseRef(in); !ok || got != want {
			t.Errorf("parseRef(%q) = %+v, %v; want %+v", in, got, ok, want)
		}
	}
	for _, in := range []string{"q3-review", "Q3 Review", "https://cahut.example.com/", "olivier/Q3", "-/healthz",
		"/-/w/q3-review", "https://cahut.example.com/-/w/" + id[:25], "/-/frame/" + id + "/"} {
		if got, ok := parseRef(in); ok {
			t.Errorf("parseRef(%q) = %+v, true; want false", in, got)
		}
	}
}

func TestWindowsHaveIDs(t *testing.T) {
	f := setup(t)
	out, err := f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil {
		t.Fatalf("publish_window: %v", err)
	}
	w := out.(Window)
	if !address.ValidID(w.ID) || w.PermanentURL != "https://cahut.test/-/w/"+w.ID {
		t.Errorf("window = %+v", w)
	}
	again, err := f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil || again.(Window).ID == w.ID {
		t.Errorf("a second Window takes the id %+v, %v", again, err)
	}

	// The id and the permanent address name the Window, as its address does.
	for _, ref := range []string{w.ID, w.PermanentURL, w.PermanentURL + "/index.html"} {
		out, err := f.invoke(t, f.olivier, "get_window", `{"window": "`+ref+`"}`)
		if err != nil || out.(Detail).ID != w.ID || out.(Detail).Name != w.Name {
			t.Errorf("get_window %s = %+v, %v", ref, out, err)
		}
	}
	if _, err := f.invoke(t, f.olivier, "trash_window", `{"window": "`+w.ID+`"}`); err != nil {
		t.Errorf("trash_window by id: %v", err)
	}
	out, err = f.invoke(t, f.olivier, "list_windows", `{"trashed": true}`)
	if err != nil || len(out.(List).Windows) != 1 || out.(List).Windows[0].ID != w.ID {
		t.Errorf("list_windows in the trash = %+v, %v", out, err)
	}

	// Outside the Rooms of the actor, an id finds nothing, and says nothing
	// of the Window.
	_, err = f.invoke(t, f.camille, "get_window", `{"window": "`+w.ID+`"}`)
	if !errors.Is(err, operation.ErrNotFound) || strings.Contains(err.Error(), "olivier") {
		t.Errorf("get_window of another's Window by id: %v", err)
	}
	_, err = f.invoke(t, f.camille, "rollback_window", `{"window": "`+w.ID+`"}`)
	if !errors.Is(err, operation.ErrNotFound) || strings.Contains(err.Error(), "olivier") {
		t.Errorf("rollback_window of another's Window by id: %v", err)
	}
	if _, err := f.invoke(t, f.olivier, "get_window", `{"window": "`+address.NewID()+`"}`); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_window of an id that names no Window: %v", err)
	}

	entries, err := logbook.List(context.Background(), f.db.Read, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		switch {
		case e.Operation == "publish_window" && e.Target == "olivier/"+w.Name && !strings.Contains(e.Detail, "id "+w.ID):
			t.Errorf("the Logbook does not keep the id of the Window: %+v", e)
		case e.Operation == "trash_window" && e.Target != "olivier/"+w.Name:
			t.Errorf("the Logbook names a Window by its id, not its address: %+v", e)
		}
	}
}

func TestUpdateAndRollback(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	out, err := f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil {
		t.Fatal(err)
	}
	w := out.(Window)
	index := func() string {
		t.Helper()
		id, version, err := current(ctx, f.db, "olivier", w.Name)
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(f.s.Dir(id, version), "index.html"))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	update := func(html string) Window {
		t.Helper()
		out, err := f.invoke(t, f.olivier, "update_window", `{"window": "olivier/`+w.Name+`", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "`+html+`"}]}`)
		if err != nil {
			t.Fatalf("update_window: %v", err)
		}
		return out.(Window)
	}

	if _, err := f.invoke(t, f.olivier, "rollback_window", `{"window": "olivier/`+w.Name+`"}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("rollback_window with a single version: err = %v, want ErrInvalid", err)
	}
	u := update("v2")
	if u.Version != 2 || u.Previous != 1 || u.FileCount != 1 || u.Size != 2 || u.URL != w.URL || u.Title != w.Title || index() != "v2" {
		t.Errorf("after an update: %+v, index.html = %q", u, index())
	}

	// A rollback swaps the versions, and a second one swaps them back.
	out, err = f.invoke(t, f.olivier, "rollback_window", `{"window": "`+w.URL+`"}`)
	if err != nil {
		t.Fatalf("rollback_window: %v", err)
	}
	if r := out.(Window); r.Version != 1 || r.Previous != 2 || r.FileCount != 2 || index() != "<h1>Q3</h1>" {
		t.Errorf("after a rollback: %+v, index.html = %q", r, index())
	}
	out, err = f.invoke(t, f.olivier, "rollback_window", `{"window": "`+w.URL+`"}`)
	if err != nil || out.(Window).Version != 2 || index() != "v2" {
		t.Errorf("after a second rollback: %+v, %v", out, err)
	}

	// After a rollback to 1, the next update is 3, and 2 goes.
	if _, err := f.invoke(t, f.olivier, "rollback_window", `{"window": "`+w.URL+`"}`); err != nil {
		t.Fatal(err)
	}
	u = update("v3")
	if u.Version != 3 || u.Previous != 1 || index() != "v3" {
		t.Errorf("after an update from a rollback: %+v", u)
	}
	id, _, _ := current(ctx, f.db, "olivier", w.Name)
	if _, err := os.Stat(f.s.Dir(id, 2)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the files of version 2 stay: %v", err)
	}
	out, err = f.invoke(t, f.olivier, "get_window", `{"window": "`+w.URL+`"}`)
	if err != nil {
		t.Fatal(err)
	}
	d := out.(Detail)
	if len(d.Versions) != 2 || d.Versions[0].Number != 3 || !d.Versions[0].Current ||
		d.Versions[1].Number != 1 || d.Versions[1].Current || d.Versions[1].FileCount != 2 || d.Versions[0].PublishedBy != "olivier" {
		t.Errorf("versions = %+v", d.Versions)
	}

	// A new title keeps the address.
	out, err = f.invoke(t, f.olivier, "update_window", `{"window": "`+w.URL+`", "title": "Q3", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "v4"}]}`)
	if err != nil || out.(Window).Title != "Q3" || out.(Window).URL != w.URL {
		t.Errorf("update with a title = %+v, %v", out, err)
	}

	// Others cannot update it, and a bad publication changes nothing.
	if _, err := f.invoke(t, f.camille, "update_window", `{"window": "`+w.URL+`", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "x"}]}`); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("update_window by another person: err = %v, want ErrNotFound", err)
	}
	if _, err := f.invoke(t, f.olivier, "update_window", `{"window": "`+w.URL+`", "declaration": {"scripts": false}, "files": [{"path": "page.html", "content": "x"}]}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("update_window without index.html: err = %v, want ErrInvalid", err)
	}
	if index() != "v4" {
		t.Errorf("index.html = %q after a refused update", index())
	}

	entries, err := logbook.List(ctx, f.db.Read, 3)
	if err != nil {
		t.Fatal(err)
	}
	if e := entries[2]; e.Operation != "update_window" || e.Outcome != logbook.OK || !strings.HasPrefix(e.Detail, "version 4") || e.RoomID == 0 {
		t.Errorf("logbook = %+v", e)
	}
}

func TestUpdateRemovesFilesWhenItFails(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	out, err := f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil {
		t.Fatal(err)
	}
	w := out.(Window)
	if _, err := f.db.Write.ExecContext(ctx, "DROP TABLE logbook"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.invoke(t, f.olivier, "update_window", `{"window": "`+w.URL+`", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "v2"}]}`); err == nil {
		t.Fatal("update_window succeeded without a Logbook")
	}
	id, version, err := current(ctx, f.db, "olivier", w.Name)
	if err != nil || version != 1 {
		t.Fatalf("Current = %d, %v", version, err)
	}
	if _, err := os.Stat(f.s.Dir(id, 2)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the files of the failed version stay: %v", err)
	}
}

func TestTrash(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	out, err := f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil {
		t.Fatal(err)
	}
	w := out.(Window)
	ref := `{"window": "` + w.URL + `"}`
	out, err = f.invoke(t, f.olivier, "trash_window", ref)
	if err != nil || out.(Window).TrashedAt == nil {
		t.Fatalf("trash_window = %+v, %v", out, err)
	}
	if _, _, err := current(ctx, f.db, "olivier", w.Name); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("a Window in the trash is served: %v", err)
	}
	if out, _ := f.invoke(t, f.olivier, "list_windows", `{}`); len(out.(List).Windows) != 0 {
		t.Errorf("list_windows shows the trash: %+v", out)
	}
	if out, _ := f.invoke(t, f.olivier, "list_windows", `{"trashed": true}`); len(out.(List).Windows) != 1 {
		t.Errorf("list_windows of the trash = %+v", out)
	}
	if out, _ := f.invoke(t, f.olivier, "get_room", `{"room": "olivier"}`); out.(room.Detail).Windows != 0 {
		t.Errorf("the Room counts the trash: %+v", out)
	}
	for _, op := range []string{"trash_window", "rollback_window"} {
		if _, err := f.invoke(t, f.olivier, op, ref); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("%s in the trash: err = %v, want ErrInvalid", op, err)
		}
	}
	if _, err := f.invoke(t, f.olivier, "update_window", `{"window": "`+w.URL+`", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "v2"}]}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("update_window in the trash: err = %v, want ErrInvalid", err)
	}
	// Its name stays taken.
	if out, err := f.invoke(t, f.olivier, "publish_window", deck); err != nil || out.(Window).Name != w.Name+"-2" {
		t.Errorf("publish_window over a Window in the trash = %+v, %v", out, err)
	}
	if _, err := f.invoke(t, f.camille, "restore_window", ref); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("restore_window by another person: err = %v, want ErrNotFound", err)
	}
	out, err = f.invoke(t, f.olivier, "restore_window", ref)
	if err != nil || out.(Window).TrashedAt != nil {
		t.Fatalf("restore_window = %+v, %v", out, err)
	}
	if _, _, err := current(ctx, f.db, "olivier", w.Name); err != nil {
		t.Errorf("a restored Window is not served: %v", err)
	}
	if _, err := f.invoke(t, f.olivier, "restore_window", ref); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("restore_window out of the trash: err = %v, want ErrInvalid", err)
	}
}

// current returns the id and the version that readers get of a Window.
func current(ctx context.Context, db *store.DB, room, name string) (int64, int, error) {
	w, err := Lookup(ctx, db, room, name, time.Now())
	return w.ID, w.Version, err
}

// TestFullDiskRefusesPublications checks that a Cahut refuses to publish
// rather than fill the disk that every Cahut of the server shares.
func TestFullDiskRefusesPublications(t *testing.T) {
	f := setup(t)
	free := uint64(MinFreeSpace + 1<<20)
	defer func(old func(string) (uint64, error)) { freeSpace = old }(freeSpace)
	freeSpace = func(string) (uint64, error) { return free, nil }
	if _, err := f.invoke(t, f.olivier, "publish_window", deck); err != nil {
		t.Fatalf("publish with room on the disk: %v", err)
	}
	free = MinFreeSpace
	if _, err := f.invoke(t, f.olivier, "publish_window", strings.Replace(deck, "Revue", "Autre revue", 1)); !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "almost full") {
		t.Errorf("publish on a full disk: err = %v", err)
	}
}

func TestRetitleAndDescribe(t *testing.T) {
	f := setup(t)
	out, err := f.invoke(t, f.olivier, "publish_window", `{"title": "Q3", "description": "  The figures\n of the third quarter. ", "kind": "deck",
		"declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "<h1>Q3</h1>"}]}`)
	if err != nil {
		t.Fatalf("publish_window: %v", err)
	}
	w := out.(Window)
	if w.Description != "The figures of the third quarter." {
		t.Errorf("publish_window with a description = %+v", w)
	}
	retitle := func(as operation.Actor, args string) (Window, error) {
		t.Helper()
		out, err := f.invoke(t, as, "retitle_window", args)
		if err != nil {
			return Window{}, err
		}
		return out.(Window), nil
	}

	// The title changes in place: same name, same version, same time of its
	// last version, and the description stays.
	got, err := retitle(f.olivier, `{"window": "`+w.ID+`", "title": " Q3 review "}`)
	if err != nil || got.Title != "Q3 review" || got.Name != "q3" || got.Version != 1 || !got.UpdatedAt.Equal(w.UpdatedAt) ||
		got.Description != w.Description || got.URL != w.URL {
		t.Errorf("retitle_window of its title = %+v, %v", got, err)
	}
	got, err = retitle(f.olivier, `{"window": "olivier/q3", "description": "What we sold."}`)
	if err != nil || got.Title != "Q3 review" || got.Description != "What we sold." {
		t.Errorf("retitle_window of its description = %+v, %v", got, err)
	}
	if out, err := f.invoke(t, f.olivier, "list_windows", `{}`); err != nil || out.(List).Windows[0].Description != "What we sold." {
		t.Errorf("list_windows = %+v, %v", out, err)
	}
	if got, err := retitle(f.olivier, `{"window": "olivier/q3", "description": ""}`); err != nil || got.Description != "" {
		t.Errorf("retitle_window with an empty description = %+v, %v", got, err)
	}

	for name, args := range map[string]string{
		"nothing":      `{"window": "olivier/q3"}`,
		"spaces":       `{"window": "olivier/q3", "title": "  "}`,
		"a long title": `{"window": "olivier/q3", "title": "` + strings.Repeat("a", 201) + `"}`,
		"too long":     `{"window": "olivier/q3", "description": "` + strings.Repeat("é", 301) + `"}`,
		"a control":    `{"window": "olivier/q3", "description": "A bell\u0007."}`,
	} {
		if _, err := retitle(f.olivier, args); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("retitle_window with %s: err = %v, want ErrInvalid", name, err)
		}
	}
	if _, err := retitle(f.camille, `{"window": "`+w.ID+`", "title": "Ours"}`); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("retitle_window by someone outside the Room: err = %v, want ErrNotFound", err)
	}

	// update_window keeps the description unless it gives one.
	update := func(extra string) Window {
		t.Helper()
		out, err := f.invoke(t, f.olivier, "update_window", `{"window": "olivier/q3", `+extra+`"declaration": {"scripts": false},
			"files": [{"path": "index.html", "content": "<h1>Q3</h1>"}]}`)
		if err != nil {
			t.Fatalf("update_window: %v", err)
		}
		return out.(Window)
	}
	if got := update(`"description": "Version two.", `); got.Description != "Version two." {
		t.Errorf("update_window with a description = %+v", got)
	}
	if got := update(""); got.Description != "Version two." || got.Title != "Q3 review" {
		t.Errorf("update_window without a description = %+v", got)
	}

	// In the trash, a Window keeps its texts.
	if _, err := f.invoke(t, f.olivier, "trash_window", `{"window": "olivier/q3"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := retitle(f.olivier, `{"window": "olivier/q3", "title": "Gone"}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("retitle_window in the trash: err = %v, want ErrInvalid", err)
	}

	entries, err := logbook.List(context.Background(), f.db.Read, 20)
	if err != nil {
		t.Fatal(err)
	}
	var logged []string
	for _, e := range entries {
		if e.Operation == "retitle_window" && e.Outcome == logbook.OK {
			logged = append(logged, e.Target+" "+e.Detail)
		}
	}
	if want := `olivier/q3 description ""; olivier/q3 description "What we sold."; olivier/q3 title "Q3 review"`; strings.Join(logged, "; ") != want {
		t.Errorf("the Logbook of retitle_window = %q, want %q", logged, want)
	}
}

func TestRename(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	publish := func(args string) (Window, error) {
		t.Helper()
		out, err := f.invoke(t, f.olivier, "publish_window", args)
		if err != nil {
			return Window{}, err
		}
		return out.(Window), nil
	}
	const files = `"kind": "deck", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "<h1>Q3</h1>"}]`

	// A name chosen at publication, rather than from the title; a taken one
	// is refused rather than numbered.
	w, err := publish(`{"title": "Revue du 3e trimestre", "name": "q3", ` + files + `}`)
	if err != nil || w.Name != "q3" || w.URL != "https://cahut.test/olivier/q3/" {
		t.Fatalf("publish_window with a name = %+v, %v", w, err)
	}
	for name, args := range map[string]string{
		"taken":     `{"title": "Another", "name": "q3", ` + files + `}`,
		"malformed": `{"title": "Another", "name": "Q3 Review", ` + files + `}`,
	} {
		if _, err := publish(args); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("publish_window with a name %s: err = %v, want ErrInvalid", name, err)
		}
	}
	other, err := publish(`{"title": "Budget", ` + files + `}`)
	if err != nil {
		t.Fatal(err)
	}

	// The name changes, and nothing else: the id, the title, the version and
	// the files stay, and the old name is free.
	out, err := f.invoke(t, f.olivier, "rename_window", `{"window": "`+w.ID+`", "name": "q3-review"}`)
	if err != nil {
		t.Fatalf("rename_window: %v", err)
	}
	renamed := out.(Window)
	if renamed.Name != "q3-review" || renamed.URL != "https://cahut.test/olivier/q3-review/" || renamed.ID != w.ID ||
		renamed.PermanentURL != w.PermanentURL || renamed.Title != w.Title || renamed.Version != 1 || !renamed.UpdatedAt.Equal(w.UpdatedAt) {
		t.Errorf("rename_window = %+v", renamed)
	}
	if _, err := f.invoke(t, f.olivier, "get_window", `{"window": "olivier/q3"}`); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_window at the old name: err = %v, want ErrNotFound", err)
	}
	if d, err := f.invoke(t, f.olivier, "get_window", `{"window": "olivier/q3-review", "include_content": true}`); err != nil ||
		d.(Detail).Files[0].Content != "<h1>Q3</h1>" {
		t.Errorf("get_window at the new name = %+v, %v", d, err)
	}

	// Undone by itself while the old name is free; refused once it is taken,
	// or not a name, or the same, or by someone outside the Room.
	if out, err := f.invoke(t, f.olivier, "rename_window", `{"window": "olivier/q3-review", "name": "q3"}`); err != nil || out.(Window).Name != "q3" {
		t.Errorf("rename_window back = %+v, %v", out, err)
	}
	for name, args := range map[string]string{
		"taken by another": `{"window": "olivier/q3", "name": "` + other.Name + `"}`,
		"the same":         `{"window": "olivier/q3", "name": "q3"}`,
		"malformed":        `{"window": "olivier/q3", "name": "-q3"}`,
		"too long":         `{"window": "olivier/q3", "name": "` + strings.Repeat("a", 64) + `"}`,
	} {
		if _, err := f.invoke(t, f.olivier, "rename_window", args); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("rename_window %s: err = %v, want ErrInvalid", name, err)
		}
	}
	if _, err := f.invoke(t, f.camille, "rename_window", `{"window": "`+w.ID+`", "name": "ours"}`); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("rename_window by someone outside the Room: err = %v, want ErrNotFound", err)
	}

	// A Window in the trash keeps its name, which no other takes.
	if _, err := f.invoke(t, f.olivier, "trash_window", `{"window": "`+other.ID+`"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := f.invoke(t, f.olivier, "rename_window", `{"window": "`+other.ID+`", "name": "budget-2027"}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("rename_window in the trash: err = %v, want ErrInvalid", err)
	}
	if _, err := f.invoke(t, f.olivier, "rename_window", `{"window": "olivier/q3", "name": "`+other.Name+`"}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("rename_window to the name of a Window in the trash: err = %v, want ErrInvalid", err)
	}

	entries, err := logbook.List(ctx, f.db.Read, 30)
	if err != nil {
		t.Fatal(err)
	}
	var logged []string
	for _, e := range entries {
		if e.Operation == "rename_window" && e.Outcome == logbook.OK {
			logged = append(logged, e.Target+" "+e.Detail)
		}
	}
	if want := "olivier/q3-review to olivier/q3, id " + w.ID + "; olivier/q3 to olivier/q3-review, id " + w.ID; strings.Join(logged, "; ") != want {
		t.Errorf("the Logbook of rename_window = %q, want %q", logged, want)
	}
}
