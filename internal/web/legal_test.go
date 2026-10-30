package web

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
)

// legalFoot is the foot of a page of the Cahut that links its legal pages.
const legalFoot = `aria-label="Legal pages"><a href="https://`

// reportFoot is the foot of a page of a Cahut without legal pages, which
// links the page where anyone reports a Window.
const reportFoot = `<footer class="legal">
<nav class="legal__links" aria-label="Legal pages"><a href="/-/report">Report a Window</a>
</nav>
</footer>`

// The administrators of the server link the legal pages of a Cahut from its
// console, and the pages that readers see link them at their foot, in
// English, and the frame of a Window does not. See
// docs/adr/0042-the-legal-pages-of-a-cahut.md.
func TestLegalLinks(t *testing.T) {
	f := setup(t)
	console := f.ts
	olivier := f.person(t, "localhost", "olivier")
	publishPage(t, f, "Launch", "public")
	publishPage(t, f, "Review", "key")

	pages := map[string]string{"the page of a Room": "/olivier/", "what is not found": "/nobody/", "the page of a Key": "/olivier/review/"}
	for what, path := range pages {
		if _, body := f.do(t, "localhost", http.MethodGet, path, nil, nil); strings.Contains(body, legalFoot) {
			t.Errorf("%s of a Cahut without legal pages links some: %q", what, body)
		}
	}
	// Their foot links the page of a report all the same, which the page of
	// a Key fills in with its Window. See
	// docs/adr/0054-reporting-and-withdrawing-a-window.md.
	if _, body := f.do(t, "localhost", http.MethodGet, "/olivier/", nil, nil); !strings.Contains(body, reportFoot) {
		t.Errorf("the page of a Room of a Cahut without legal pages does not link the page of a report: %q", body)
	}
	if _, body := f.do(t, "localhost", http.MethodGet, "/olivier/review/", nil, nil); !strings.Contains(body, `<a href="/-/report?window=https%3A%2F%2Flocalhost%2Folivier%2Freview%2F">Report a Window</a>`) {
		t.Errorf("the page of a Key does not fill in its Window in the link of a report: %q", body)
	}

	// An administrator of the server sets them over the API, in the console
	// of the Cahut.
	a, err := operation.Do(context.Background(), f.srv.Registry, server.CreateAdmin, admin, server.CreateAdminInput{Handle: "root", Name: "Root"})
	if err != nil {
		t.Fatal(err)
	}
	r := call(t, f, console, http.MethodPost, "localhost", "/console/api/server/activate_account", "",
		`{"code": "`+a.Activation.Code+`", "password": "`+password+`"}`)
	if r.status != http.StatusOK {
		t.Fatalf("activate_account of the server: %d %s", r.status, r.raw)
	}
	root := r.body["token"].(string)
	checkRefusals(t, f, console, root, []refusalCase{
		{
			name: "an address in HTTP", path: "/console/api/set_legal_links", body: `{"terms": "http://host.example/terms"}`,
			status: 400, code: "legal_link_invalid", values: map[string]any{"link": "terms"},
			error: "invalid input: terms: give the address of a page, which starts with https://, such as https://example.org/terms, or leave it out to remove the link",
		},
		{
			name: "an address in French alone", path: "/console/api/set_legal_links", body: `{"fr": {"privacy": "https://host.example/fr/confidentialite"}}`,
			status: 400, code: "legal_link_in_french_only", values: map[string]any{"link": "fr.privacy", "page": "privacy"},
			error: "invalid input: fr.privacy: give privacy too, the address of the page for every other language; the address in French takes its place in French only",
		},
	})
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/set_legal_links", root, `{
		"legal_notice": "https://host.example/legal", "terms": "https://host.example/terms?from=cahut&lang=en", "privacy": "https://host.example/privacy",
		"fr": {"legal_notice": "https://host.example/fr/mentions-legales"}}`)
	if r.status != http.StatusOK || r.body["terms"] != "https://host.example/terms?from=cahut&lang=en" {
		t.Fatalf("set_legal_links: %d %s", r.status, r.raw)
	}
	// Those of the Cahut cannot.
	camille := activate(t, f, console, "camille", true)
	if r := call(t, f, console, http.MethodPost, "localhost", "/console/api/set_legal_links", camille, `{}`); r.status != http.StatusForbidden {
		t.Errorf("set_legal_links as an administrator of the Cahut: %d %s", r.status, r.raw)
	}

	// The console reads them before sign-in, and once signed in.
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/describe_sign_in", "", `{}`)
	if legal, _ := r.body["legal"].(map[string]any); r.status != http.StatusOK || legal["privacy"] != "https://host.example/privacy" || legal["fr"] == nil {
		t.Errorf("describe_sign_in: %d %s", r.status, r.raw)
	}
	r = call(t, f, console, http.MethodPost, "localhost", "/console/api/get_cahut", camille, `{}`)
	if legal, _ := r.body["legal"].(map[string]any); r.status != http.StatusOK || legal["legal_notice"] != "https://host.example/legal" {
		t.Errorf("get_cahut: %d %s", r.status, r.raw)
	}

	// The pages for readers speak English, and link the addresses that are
	// not in French, in the order of the pages.
	want := `<footer class="legal">
<nav class="legal__links" aria-label="Legal pages"><a href="https://host.example/legal" target="_blank" rel="noopener">Legal notice</a>
<a href="https://host.example/terms?from=cahut&amp;lang=en" target="_blank" rel="noopener">Terms</a>
<a href="https://host.example/privacy" target="_blank" rel="noopener">Privacy</a>
<a href="/-/report`
	for what, path := range pages {
		if _, body := f.do(t, "localhost", http.MethodGet, path, nil, nil); !strings.Contains(body, want) || strings.Contains(body, "mentions-legales") {
			t.Errorf("%s does not link the legal pages: %q", what, body)
		}
	}

	// The Window that a Room shows at its address fills the page, without
	// them.
	f.call(t, "localhost", olivier, "set_room_home", `{"room": "olivier", "window": "launch"}`)
	if _, body := f.do(t, "localhost", http.MethodGet, "/olivier/", nil, nil); !strings.Contains(body, `<iframe src="/olivier/launch/"`) || strings.Contains(body, `<footer`) {
		t.Errorf("the home of a Room = %q", body)
	}
	// Another Cahut of the server links none.
	if _, body := f.do(t, "beta.localhost", http.MethodGet, "/nobody/", nil, nil); strings.Contains(body, legalFoot) {
		t.Errorf("another Cahut links the legal pages of the first: %q", body)
	}
}
