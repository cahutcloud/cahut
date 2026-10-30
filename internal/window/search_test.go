package window

import (
	"context"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
)

func TestTextOf(t *testing.T) {
	got := textOf(pagesOf([]decoded{
		{"b.html", []byte("<p>Second page</p>")},
		{"css/deck.css", []byte("ignored")},
		{"a.HTM", []byte("<p>First   page</p>")},
		{"index.html", []byte(`<!doctype html><html><head><title>Q3 review</title><style>h1 { color: red }</style>
			<script>const secret = "nothing";</script></head><body><h1>Été&nbsp;en <b>Bretagne</b></h1>
			<img src="a.png" alt="A map of sales"><template><p>Not shown</p></template><template/><noscript>Turn on scripts</noscript>
			<svg><text>Drawn</text></svg><p>The end &amp; more</p></body></html>`)},
	}))
	want := "Q3 review Été en Bretagne A map of sales Drawn The end & more First page Second page"
	if got != want {
		t.Errorf("textOf =\n%q, want\n%q", got, want)
	}

	// The text stops at MaxText, between two words.
	long := strings.Repeat("word ", MaxText/4)
	if got := textOf([]decoded{{"index.html", []byte(long)}, {"more.html", []byte("after")}}); len(got) > MaxText || strings.Contains(got, "after") ||
		!strings.HasSuffix(got, "word") {
		t.Errorf("textOf of a long page: %d bytes, ends with %q", len(got), got[max(len(got)-10, 0):])
	}
	if got := textOf([]decoded{{"index.html", []byte("caf\xe9 au lait")}}); got != "caf� au lait" {
		t.Errorf("textOf of a page that is not UTF-8 = %q", got)
	}
	if got := textOf(nil); got != "" {
		t.Errorf("textOf of no page = %q", got)
	}
}

func TestMatchOf(t *testing.T) {
	for s, want := range map[string]string{
		"Q3 sales":                `"Q3"* "sales"*`,
		`"quoted" OR a*`:          `"quoted"* "OR"* "a"*`,
		"été, en Bretagne!":       `"été"* "en"* "Bretagne"*`,
		"  ?! ":                   "",
		"a b c d e f g h i j k l": `"a"* "b"* "c"* "d"* "e"* "f"* "g"* "h"* "i"* "j"*`,
	} {
		if got := matchOf(s); got != want {
			t.Errorf("matchOf(%q) = %q, want %q", s, got, want)
		}
	}
}

func TestSearch(t *testing.T) {
	f := setup(t)
	publish := func(as operation.Actor, title, extra, body string) Window {
		t.Helper()
		out, err := f.invoke(t, as, "publish_window", `{"title": "`+title+`", "kind": "document", `+extra+`
			"declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "`+body+`"}]}`)
		if err != nil {
			t.Fatalf("publish_window of %s: %v", title, err)
		}
		return out.(Window)
	}
	search := func(as operation.Actor, args string) []Window {
		t.Helper()
		out, err := f.invoke(t, as, "list_windows", args)
		if err != nil {
			t.Fatalf("list_windows %s: %v", args, err)
		}
		return out.(List).Windows
	}
	names := func(ws []Window) []string {
		var n []string
		for _, w := range ws {
			n = append(n, w.Name)
		}
		return n
	}

	publish(f.olivier, "Summer", `"tags": ["holidays"],`, `<h1>Été en Bretagne</h1><p>We sold umbrellas in Brittany and Normandy.</p>`)
	publish(f.olivier, "Bretagne", `"description": "Our shop in Rennes.",`, `<p>Opening hours.</p>`)
	publish(f.olivier, "Plan", `"access": "author",`, `<p>Umbrellas for next year.</p>`)
	publish(f.camille, "Hers", "", `<p>Umbrellas everywhere, in Bretagne too.</p>`)

	// Words match at their start, whatever their case and accents, in the
	// title, the description, the tags and the text; the title weighs most.
	got := search(f.olivier, `{"query": "bretagne"}`)
	if !slices.Equal(names(got), []string{"bretagne", "summer"}) {
		t.Errorf("a search of bretagne = %q", names(got))
	}
	if got[1].Excerpt != "Été en Bretagne We sold umbrellas in Brittany and Normandy." || got[0].Excerpt != "" {
		t.Errorf("the excerpts = %q, %q", got[0].Excerpt, got[1].Excerpt)
	}
	for query, want := range map[string][]string{
		"ETE bret":      {"summer"},
		"umbrel":        {"plan", "summer"},
		"rennes":        {"bretagne"},
		"holiday":       {"summer"},
		"umbrella shop": nil,
	} {
		if got := names(search(f.olivier, `{"query": "`+query+`"}`)); !slices.Equal(got, want) {
			t.Errorf("a search of %q = %q, want %q", query, got, want)
		}
	}
	// It finds Windows of one's own Rooms only, but those that others keep
	// for themselves.
	if got := names(search(f.camille, `{"query": "umbrellas"}`)); !slices.Equal(got, []string{"hers"}) {
		t.Errorf("a search of camille = %q", got)
	}
	// It goes with a tag, a Room and the trash.
	if got := names(search(f.olivier, `{"query": "bretagne", "tag": "holidays"}`)); !slices.Equal(got, []string{"summer"}) {
		t.Errorf("a search with a tag = %q", got)
	}
	if got := names(search(f.olivier, `{"query": "bretagne", "room": "olivier"}`)); len(got) != 2 {
		t.Errorf("a search in a Room = %q", got)
	}
	if _, err := f.invoke(t, f.olivier, "list_windows", `{"query": "?!"}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("a search without a word: err = %v, want ErrInvalid", err)
	}

	// A new title, new tags, a new version and a rollback change what it
	// finds.
	if _, err := f.invoke(t, f.olivier, "retitle_window", `{"window": "olivier/summer", "title": "Holidays in Brittany", "tags": ["sea"]}`); err != nil {
		t.Fatal(err)
	}
	if got := names(search(f.olivier, `{"query": "sea brittany"}`)); !slices.Equal(got, []string{"summer"}) {
		t.Errorf("a search after retitle_window = %q", got)
	}
	if _, err := f.invoke(t, f.olivier, "update_window", `{"window": "olivier/summer", "declaration": {"scripts": false},
		"files": [{"path": "index.html", "content": "<p>Raincoats in Lyon.</p>"}]}`); err != nil {
		t.Fatal(err)
	}
	if got := names(search(f.olivier, `{"query": "umbrellas"}`)); !slices.Equal(got, []string{"plan"}) {
		t.Errorf("a search of the text of the old version = %q", got)
	}
	if got := names(search(f.olivier, `{"query": "raincoats"}`)); !slices.Equal(got, []string{"summer"}) {
		t.Errorf("a search of the text of the new version = %q", got)
	}
	if _, err := f.invoke(t, f.olivier, "rollback_window", `{"window": "olivier/summer"}`); err != nil {
		t.Fatal(err)
	}
	if got := names(search(f.olivier, `{"query": "umbrellas"}`)); !slices.Equal(got, []string{"plan", "summer"}) {
		t.Errorf("a search after rollback_window = %q", got)
	}

	// The trash keeps a Window out, but with trashed; a purge takes its row.
	if _, err := f.invoke(t, f.olivier, "trash_window", `{"window": "olivier/bretagne"}`); err != nil {
		t.Fatal(err)
	}
	if got := names(search(f.olivier, `{"query": "rennes"}`)); len(got) != 0 {
		t.Errorf("a search finds a Window in the trash: %q", got)
	}
	if got := names(search(f.olivier, `{"query": "rennes", "trashed": true}`)); !slices.Equal(got, []string{"bretagne"}) {
		t.Errorf("a search of the trash = %q", got)
	}
	if _, err := f.invoke(t, person(f.olivier), "purge_window", `{"window": "olivier/bretagne"}`); err != nil {
		t.Fatal(err)
	}
	if n := f.count(t, "SELECT count(*) FROM window_search WHERE window_search MATCH 'rennes'"); n != 0 {
		t.Errorf("a purged Window leaves %d rows in the index", n)
	}

	entries, err := logbook.List(context.Background(), f.db.Read, 1)
	if err != nil {
		t.Fatal(err)
	}
	if e := entries[0]; e.Operation != "purge_window" {
		t.Fatalf("the last entry is %s", e.Operation)
	}
	search(f.olivier, `{"query": "umbrellas", "tag": "sea"}`)
	entries, _ = logbook.List(context.Background(), f.db.Read, 1)
	if e := entries[0]; e.Operation != "list_windows" || e.Detail != `tag sea; query "umbrellas"; 1 Windows` {
		t.Errorf("the Logbook of a search = %q", e.Detail)
	}
}

func TestReadTexts(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	publish := func(title, body string) Window {
		t.Helper()
		out, err := f.invoke(t, f.olivier, "publish_window", `{"title": "`+title+`", "kind": "document", "declaration": {"scripts": false},
			"files": [{"path": "index.html", "content": "`+body+`"}, {"path": "notes/more.html", "content": "<p>`+title+` notes</p>"}]}`)
		if err != nil {
			t.Fatal(err)
		}
		return out.(Window)
	}
	publish("Atlas", "<p>Mountains of Morocco</p>")
	lost := publish("Lost", "<p>Gone</p>")
	if _, err := f.invoke(t, f.olivier, "update_window", `{"window": "olivier/atlas", "declaration": {"scripts": false},
		"files": [{"path": "index.html", "content": "<p>Rivers of Morocco</p>"}]}`); err != nil {
		t.Fatal(err)
	}

	// The Cahut was there before the search: its versions have no text yet,
	// and its index holds their titles alone, as the migration left them.
	for _, q := range []string{"UPDATE window_versions SET text = NULL", "UPDATE window_search SET text = ''"} {
		if _, err := f.db.Write.ExecContext(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	id, version, err := current(ctx, f.db, "olivier", lost.Name)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(f.s.Dir(id, version)); err != nil {
		t.Fatal(err)
	}
	if n := f.count(t, "SELECT count(*) FROM window_search WHERE window_search MATCH 'morocco'"); n != 0 {
		t.Fatalf("the index finds %d Windows before the texts are read", n)
	}

	read, err := f.s.ReadTexts(ctx, f.db)
	if err != nil || read != 3 {
		t.Errorf("ReadTexts = %d, %v, want the 3 versions", read, err)
	}
	var texts []string
	rows, err := f.db.Read.QueryContext(ctx, "SELECT text FROM window_versions ORDER BY window_id, number")
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		texts = append(texts, s)
	}
	rows.Close()
	if want := []string{"Mountains of Morocco Atlas notes", "Rivers of Morocco", ""}; !slices.Equal(texts, want) {
		t.Errorf("the texts read = %q, want %q", texts, want)
	}
	out, err := f.invoke(t, f.olivier, "list_windows", `{"query": "rivers"}`)
	if err != nil || len(out.(List).Windows) != 1 || out.(List).Windows[0].Excerpt != "Rivers of Morocco" {
		t.Errorf("a search after ReadTexts = %+v, %v", out, err)
	}
	if n := f.count(t, "SELECT count(*) FROM window_search WHERE window_search MATCH 'mountains'"); n != 0 {
		t.Errorf("the index holds the text of a previous version")
	}
	if read, err := f.s.ReadTexts(ctx, f.db); err != nil || read != 0 {
		t.Errorf("ReadTexts again = %d, %v, want nothing to read", read, err)
	}
}
