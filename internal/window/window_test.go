package window

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
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
	db, err := store.Open(ctx, filepath.Join(dir, "cahut.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	r := operation.NewRegistry(db)
	account.Register(r)
	s := NewService(filepath.Join(dir, "windows"), func(name string) string { return "https://" + name + ".home.test/" })
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
	"files": [
		{"path": "index.html", "content": "<h1>Q3</h1>"},
		{"path": "img/logo.png", "content": "iVBORw0KGgo=", "encoding": "base64"}
	]
}`

func TestPublishAndGet(t *testing.T) {
	f := setup(t)
	published := ""
	f.s.Published = func(name string) { published = name }

	out, err := f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil {
		t.Fatalf("publish_window: %v", err)
	}
	w := out.(Window)
	if !regexp.MustCompile(`^revue-du-3e-trimestre-[a-z2-7]{8}$`).MatchString(w.Name) {
		t.Errorf("name = %q", w.Name)
	}
	if w.URL != "https://"+w.Name+".home.test/" || w.Kind != Deck || w.Version != 1 || w.FileCount != 2 ||
		w.Size != int64(len("<h1>Q3</h1>")+8) || w.Room != "olivier" || w.CreatedBy != "olivier" {
		t.Errorf("window = %+v", w)
	}
	if published != w.Name {
		t.Errorf("Published called with %q, want %q", published, w.Name)
	}
	html, err := os.ReadFile(filepath.Join(f.s.Dir(w.Name, 1), "index.html"))
	if err != nil || string(html) != "<h1>Q3</h1>" {
		t.Errorf("index.html on disk = %q, %v", html, err)
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
	if entries[1].Operation != "publish_window" || entries[1].Target != w.Name || entries[1].Outcome != logbook.OK ||
		!strings.Contains(entries[1].Detail, "2 files") || entries[0].Operation != "get_window" {
		t.Errorf("logbook = %+v", entries[:2])
	}
}

func TestWindowsStayInTheirRoom(t *testing.T) {
	f := setup(t)
	out, err := f.invoke(t, f.olivier, "publish_window", deck)
	if err != nil {
		t.Fatal(err)
	}
	name := out.(Window).Name

	if _, err := f.invoke(t, f.camille, "get_window", `{"window": "`+name+`"}`); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_window by another person: err = %v, want ErrNotFound", err)
	}
	out, err = f.invoke(t, f.camille, "list_windows", `{}`)
	if err != nil || len(out.(List).Windows) != 0 {
		t.Errorf("list_windows by another person = %+v, %v", out, err)
	}
	out, err = f.invoke(t, f.olivier, "list_windows", ``)
	if err != nil || len(out.(List).Windows) != 1 {
		t.Errorf("list_windows by the owner = %+v, %v", out, err)
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
			_, err := f.invoke(t, f.olivier, "publish_window", `{"title": "T", "kind": "site", "files": `+files+`}`)
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
		`{"title": " ", "kind": "deck", "files": [{"path": "index.html", "content": "x"}]}`,
		`{"title": "T", "kind": "slides", "files": [{"path": "index.html", "content": "x"}]}`,
		`{"title": "T", "kind": "deck", "files": [{"path": "index.html", "content": "x"}], "room": "team"}`,
	} {
		if _, err := f.invoke(t, f.olivier, "publish_window", args); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", args, err)
		}
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

func TestSlug(t *testing.T) {
	tests := map[string]string{
		"Revue du 3e trimestre":          "revue-du-3e-trimestre",
		"  Présentation : Œuvre ":        "presentation-oeuvre",
		"Straße & Ölmühle":               "strasse-olmuhle",
		"日本語":                            "window",
		"---":                            "window",
		strings.Repeat("abcdefghij ", 6): "abcdefghij-abcdefghij-abcdefghij-abcdefg",
	}
	for title, want := range tests {
		if got := slug(title); got != want {
			t.Errorf("slug(%q) = %q, want %q", title, got, want)
		}
	}
}

func TestNameFrom(t *testing.T) {
	tests := map[string]string{
		"q3-abcdefgh":                               "q3-abcdefgh",
		"https://q3-abcdefgh.home.example.com/":     "q3-abcdefgh",
		"https://q3-abcdefgh.home.example.com/a.js": "q3-abcdefgh",
		"q3-abcdefgh.home.example.com":              "q3-abcdefgh",
		"http://q3-abcdefgh.localhost:8080/":        "q3-abcdefgh",
		"Q3 Review":                                 "Q3 Review",
	}
	for in, want := range tests {
		if got := nameFrom(in); got != want {
			t.Errorf("nameFrom(%q) = %q, want %q", in, got, want)
		}
	}
}
