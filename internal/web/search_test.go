package web

import (
	"context"
	"testing"

	"cahut.cloud/internal/window"
)

// The service reads, in every Cahut, the text of the versions of Windows
// published before the search, so that a search finds their words.
func TestReadTextsOfEveryCahut(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	for _, domain := range []string{"localhost", "beta.localhost"} {
		f.call(t, domain, f.person(t, domain, "olivier"), "publish_window", `{"title": "Atlas", "kind": "document",
			"declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "<p>Mountains of `+domain+`</p>"}]}`)
		// As the migration left the Cahuts that were there before.
		for _, q := range []string{"UPDATE window_versions SET text = NULL", "UPDATE window_search SET text = ''"} {
			if _, err := f.open(t, domain).DB.Write.ExecContext(ctx, q); err != nil {
				t.Fatal(err)
			}
		}
	}
	f.web.ReadTexts(ctx)
	for _, domain := range []string{"localhost", "beta.localhost"} {
		out := f.call(t, domain, f.person(t, domain, "olivier"), "list_windows", `{"query": "mountains"}`).(window.List)
		if len(out.Windows) != 1 || out.Windows[0].Excerpt != "Mountains of "+domain {
			t.Errorf("a search in %s after ReadTexts = %+v", domain, out.Windows)
		}
	}
}
