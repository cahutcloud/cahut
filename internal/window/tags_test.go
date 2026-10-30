package window

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
)

func TestCheckTag(t *testing.T) {
	for given, want := range map[string]string{
		"q3":                    "q3",
		"  #Q3 Review ":         "q3-review",
		"sales_team":            "sales-team",
		"a - b":                 "a-b",
		"-edge-":                "edge",
		"Été":                   "été",
		"été":                 "été", // the accents apart, as some keyboards type them
		"日本":                    "日本",
		"":                      "",
		" # ":                   "",
		strings.Repeat("a", 32): strings.Repeat("a", 32),
	} {
		if got, err := CheckTag(given); err != nil || got != want {
			t.Errorf("CheckTag(%q) = %q, %v, want %q", given, got, err, want)
		}
	}
	for _, given := range []string{"a/b", "q3!", "##q3", "a.b", strings.Repeat("é", 33)} {
		_, err := CheckTag(given)
		var r *operation.Refusal
		if !errors.As(err, &r) || r.Code != "tag_invalid" || !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("CheckTag(%q): err = %v, want the refusal tag_invalid", given, err)
		}
	}
}

func TestCheckTags(t *testing.T) {
	got, err := CheckTags([]string{"Sales", "q3", "", "sales", "#Q3"})
	if err != nil || !slices.Equal(got, []string{"q3", "sales"}) {
		t.Errorf("CheckTags = %q, %v", got, err)
	}
	if got, err := CheckTags(nil); err != nil || got == nil || len(got) != 0 {
		t.Errorf("CheckTags(nil) = %#v, %v, want an empty list", got, err)
	}
	// Ten tags, given twice, are ten.
	var ten []string
	for _, c := range "abcdefghij" {
		ten = append(ten, string(c), strings.ToUpper(string(c)))
	}
	if got, err := CheckTags(ten); err != nil || len(got) != 10 {
		t.Errorf("CheckTags of ten tags twice = %q, %v", got, err)
	}
	_, err = CheckTags(append(ten, "k"))
	var r *operation.Refusal
	if !errors.As(err, &r) || r.Code != "tags_too_many" || r.Values != (TagCount{Count: 11, Limit: 10}) {
		t.Errorf("CheckTags of eleven tags: err = %#v, want the refusal tags_too_many", err)
	}
}

func TestTags(t *testing.T) {
	f := setup(t)
	publish := func(as operation.Actor, title, tags string) Window {
		t.Helper()
		out, err := f.invoke(t, as, "publish_window", `{"title": "`+title+`", "kind": "document", "tags": `+tags+`,
			"declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "<h1>`+title+`</h1>"}]}`)
		if err != nil {
			t.Fatalf("publish_window of %s: %v", title, err)
		}
		return out.(Window)
	}
	retitle := func(args string) (Window, error) {
		t.Helper()
		out, err := f.invoke(t, f.olivier, "retitle_window", args)
		if err != nil {
			return Window{}, err
		}
		return out.(Window), nil
	}
	list := func(as operation.Actor, args string) []string {
		t.Helper()
		out, err := f.invoke(t, as, "list_windows", args)
		if err != nil {
			t.Fatalf("list_windows %s: %v", args, err)
		}
		var names []string
		for _, w := range out.(List).Windows {
			names = append(names, w.Room+"/"+w.Name)
		}
		return names
	}

	q3 := publish(f.olivier, "Q3", `["Sales", "#Q3", "sales"]`)
	if !slices.Equal(q3.Tags, []string{"q3", "sales"}) {
		t.Errorf("publish_window with tags = %q", q3.Tags)
	}
	plan := publish(f.olivier, "Plan", `[]`)
	if plan.Tags == nil || len(plan.Tags) != 0 {
		t.Errorf("publish_window without tags = %#v, want an empty list", plan.Tags)
	}
	publish(f.camille, "Hers", `["sales"]`)
	for name, tags := range map[string]string{
		"a bad tag":     `["a/b"]`,
		"too many tags": `["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"]`,
	} {
		_, err := f.invoke(t, f.olivier, "publish_window", `{"title": "Bad", "kind": "document", "tags": `+tags+`,
			"declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "x"}]}`)
		if !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("publish_window with %s: err = %v, want ErrInvalid", name, err)
		}
	}

	// The tags are found by list_windows, in the Rooms of the person only.
	if got := list(f.olivier, `{"tag": "Sales"}`); !slices.Equal(got, []string{"olivier/q3"}) {
		t.Errorf("list_windows with the tag sales = %q", got)
	}
	if got := list(f.olivier, `{"tag": "nothing"}`); len(got) != 0 {
		t.Errorf("list_windows with a tag that no Window has = %q", got)
	}
	if got := list(f.olivier, `{"tag": ""}`); len(got) != 2 {
		t.Errorf("list_windows with an empty tag = %q, want every Window", got)
	}
	if _, err := f.invoke(t, f.olivier, "list_windows", `{"tag": "a/b"}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("list_windows with a bad tag: err = %v, want ErrInvalid", err)
	}

	// retitle_window changes the tags in place, and keeps them when it is not
	// given any.
	got, err := retitle(`{"window": "olivier/q3", "tags": ["Review", "q3"]}`)
	if err != nil || !slices.Equal(got.Tags, []string{"q3", "review"}) || got.Title != "Q3" || got.Version != 1 {
		t.Errorf("retitle_window of its tags = %+v, %v", got, err)
	}
	if got, err := retitle(`{"window": "olivier/q3", "title": "Q3 review"}`); err != nil || !slices.Equal(got.Tags, []string{"q3", "review"}) {
		t.Errorf("retitle_window of its title = %+v, %v", got, err)
	}
	if _, err := retitle(`{"window": "olivier/q3", "tags": ["a/b"]}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("retitle_window with a bad tag: err = %v, want ErrInvalid", err)
	}
	if got := list(f.olivier, `{"tag": "sales"}`); len(got) != 0 {
		t.Errorf("list_windows with the tag sales, gone = %q", got)
	}

	// A new version and a rollback keep them.
	out, err := f.invoke(t, f.olivier, "update_window", `{"window": "olivier/q3", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "v2"}]}`)
	if err != nil || !slices.Equal(out.(Window).Tags, []string{"q3", "review"}) {
		t.Errorf("update_window = %+v, %v", out, err)
	}
	out, err = f.invoke(t, f.olivier, "rollback_window", `{"window": "olivier/q3"}`)
	if err != nil || !slices.Equal(out.(Window).Tags, []string{"q3", "review"}) {
		t.Errorf("rollback_window = %+v, %v", out, err)
	}
	out, err = f.invoke(t, f.olivier, "get_window", `{"window": "`+q3.ID+`"}`)
	if err != nil || !slices.Equal(out.(Detail).Tags, []string{"q3", "review"}) {
		t.Errorf("get_window = %+v, %v", out, err)
	}

	// An empty list removes them.
	if got, err := retitle(`{"window": "olivier/q3", "tags": []}`); err != nil || got.Tags == nil || len(got.Tags) != 0 {
		t.Errorf("retitle_window with no tags = %#v, %v", got.Tags, err)
	}
	if n := f.count(t, "SELECT count(*) FROM window_tags WHERE window_id = (SELECT id FROM windows WHERE uid = ?)", q3.ID); n != 0 {
		t.Errorf("the Window keeps %d tags, want none", n)
	}

	// A Window that goes for good takes its tags with it.
	if _, err := retitle(`{"window": "olivier/plan", "tags": ["draft"]}`); err != nil {
		t.Fatal(err)
	}
	if _, err := f.invoke(t, f.olivier, "trash_window", `{"window": "olivier/plan"}`); err != nil {
		t.Fatal(err)
	}
	if got := list(f.olivier, `{"tag": "draft"}`); len(got) != 0 {
		t.Errorf("list_windows with a tag of a Window in the trash = %q", got)
	}
	if got := list(f.olivier, `{"tag": "draft", "trashed": true}`); !slices.Equal(got, []string{"olivier/plan"}) {
		t.Errorf("list_windows of the trash with a tag = %q", got)
	}
	if _, err := f.invoke(t, person(f.olivier), "purge_window", `{"window": "olivier/plan"}`); err != nil {
		t.Fatal(err)
	}
	if n := f.count(t, "SELECT count(*) FROM window_tags WHERE tag = 'draft'"); n != 0 {
		t.Errorf("a purged Window leaves %d tags", n)
	}

	entries, err := logbook.List(context.Background(), f.db.Read, 50)
	if err != nil {
		t.Fatal(err)
	}
	var logged []string
	for _, e := range entries {
		if e.Operation == "retitle_window" && e.Outcome == logbook.OK || e.Operation == "publish_window" && e.Target == "olivier/q3" {
			logged = append(logged, e.Target+" "+e.Detail[strings.LastIndex(e.Detail, ";")+1:])
		}
	}
	if want := `olivier/plan tags draft; olivier/q3 no tags; olivier/q3 title "Q3 review"; olivier/q3 tags q3, review; olivier/q3  tags q3, sales`; strings.Join(logged, "; ") != want {
		t.Errorf("the Logbook = %q, want %q", strings.Join(logged, "; "), want)
	}
}
