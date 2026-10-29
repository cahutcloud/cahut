package web

import (
	"bytes"
	"context"
	"html"
	"image/png"
	"io"
	"net/http"
	"strings"
	"testing"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/card"
	"cahut.cloud/internal/window"
)

// User-Agents of robots of previews, and of others.
const (
	slackbot = "Slackbot-LinkExpanding 1.0 (+https://api.slack.com/robots)"
	iMessage = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_1) AppleWebKit/601.2.4 (KHTML, like Gecko) Version/9.0.1 Safari/601.2.4 facebookexternalhit/1.1 Facebot Twitterbot/1.0"
	chrome   = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36"
)

// fetch gets path on host as the client called ua, with headers.
func (f *fixture) fetch(t *testing.T, host, path, ua string, cookies []*http.Cookie, header http.Header) (response, []byte) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, f.ts.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range header {
		req.Header[k] = v
	}
	req.Header.Set("User-Agent", ua)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := f.client(host, "").Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return response{StatusCode: resp.StatusCode, Header: resp.Header}, b
}

func TestPreviewRobots(t *testing.T) {
	for ua, want := range map[string]bool{
		slackbot: true,
		iMessage: true,
		"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)":             true,
		"LinkedInBot/1.0 (compatible; Mozilla/5.0; Apache-HttpClient +http://www.linkedin.com)": true,
		"Mozilla/5.0 (compatible; Discordbot/2.0; +https://discordapp.com)":                     true,
		"WhatsApp/2.23.20.0 A": true,
		"WhatsApp/2":           true, // Signal
		"Mozilla/5.0 (Windows NT 6.1; WOW64) SkypeUriPreview Preview/0.5 skype-url-preview@microsoft.com": true,
		"TelegramBot (like TwitterBot)":                             true,
		"http.rb/5.1.1 (Mastodon/4.3.0; +https://mastodon.social/)": true,
		"Bluesky Cardyb/1.1":                                        true,
		chrome:                                                      false,
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)": false,
		"curl/8.7.1": false,
		"":           false,
	} {
		r, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		r.Header.Set("User-Agent", ua)
		if got := previewRobot(r); got != want {
			t.Errorf("previewRobot(%q) = %v, want %v", ua, got, want)
		}
	}
}

// A robot of previews reads at the permanent address of a Window what that
// address tells a stranger, and opens nothing; people get the Window as
// before. See docs/adr/0051-sharing-a-window-and-its-preview.md.
func TestPreviewsOfLinks(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	w := publishDeck(t, f, olivier)
	f.call(t, "localhost", olivier, "retitle_window", `{"window": "`+w.ID+`", "description": "Revenue & costs, by region."}`)
	perm := window.PermanentPrefix + w.ID
	preview := func(ua string) string {
		t.Helper()
		resp, b := f.fetch(t, "localhost", perm, ua, nil, nil)
		if resp.StatusCode != http.StatusOK || len(resp.Cookies()) != 0 || !strings.Contains(resp.Header.Get("Vary"), "User-Agent") {
			t.Fatalf("GET %s by %q = %d, cookies %v, Vary %q", perm, ua, resp.StatusCode, resp.Cookies(), resp.Header.Get("Vary"))
		}
		return string(b)
	}
	has := func(page string, tags ...string) {
		t.Helper()
		for _, tag := range tags {
			if !strings.Contains(page, tag) {
				t.Errorf("the preview has no %s:\n%s", tag, page)
			}
		}
	}
	common := []string{
		`<meta property="og:type" content="website">`,
		`<meta property="og:site_name" content="localhost">`,
		`<meta property="og:url" content="https://localhost` + perm + `">`,
		`<meta property="og:image:width" content="1200">`,
		`<meta name="twitter:card" content="summary_large_image">`,
	}
	image := func(title string) string {
		return `<meta property="og:image" content="https://localhost/-/card/` + w.ID + `.png?v=` + card.Tag(title, "localhost") + `">`
	}

	// Room only: the Cahut, and nothing of the Window.
	page := preview(slackbot)
	has(page, append(common, `<meta property="og:title" content="A Window at localhost">`,
		`<meta property="og:description" content="Sign in to open it.">`, image(""))...)
	if strings.Contains(page, "Q3") || strings.Contains(page, "Revenue") || strings.Contains(page, "q3-review") || strings.Contains(page, "/-/frame/") {
		t.Errorf("the preview of a Window open to its Room tells of it:\n%s", page)
	}
	if page := preview(iMessage); !strings.Contains(page, "A Window at localhost") {
		t.Errorf("the preview for iMessage = %q", page)
	}
	// People, and the files under the permanent address, as before.
	for _, c := range []struct{ path, ua string }{{perm, chrome}, {perm, ""}, {perm + "/css/deck.css", slackbot}} {
		if resp, _ := f.fetch(t, "localhost", c.path, c.ua, nil, nil); resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s by %q = %d, want 404", c.path, c.ua, resp.StatusCode)
		}
	}

	// Key: its title, which the page of its Key shows, and no pass.
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "`+w.ID+`", "access": "key", "key": "correct horse"}`)
	page = preview(slackbot)
	has(page, append(common, `<meta property="og:title" content="Q3 review">`,
		`<meta property="og:description" content="Opens with its Key.">`, image("Q3 review"),
		`<meta property="og:image:alt" content="Q3 review, a Window at localhost">`)...)
	if strings.Contains(page, "Revenue") || strings.Contains(page, "q3-review") {
		t.Errorf("the preview of a Window with a Key tells more than its page of a Key:\n%s", page)
	}

	// Public: its title and its description.
	f.call(t, "localhost", olivier, "set_window_access", `{"window": "`+w.ID+`", "access": "public"}`)
	page = preview(slackbot)
	has(page, append(common, `<meta property="og:title" content="Q3 review">`, image("Q3 review"),
		`<meta property="og:description" content="`+html.EscapeString("Revenue & costs, by region.")+`">`)...)
	if resp, _ := f.fetch(t, "localhost", perm, chrome, nil, nil); resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("GET %s by a browser = %d, want a redirect", perm, resp.StatusCode)
	}
	if resp, b := f.fetch(t, "localhost", "/olivier/q3-review/", slackbot, nil, nil); resp.StatusCode != http.StatusOK || string(b) != "<h1>Q3</h1>" {
		t.Errorf("the address of the moment by a robot = %d %q, want the files of the author", resp.StatusCode, b)
	}

	// In the trash, or at an id that names no Window: no preview.
	f.call(t, "localhost", olivier, "trash_window", `{"window": "`+w.ID+`"}`)
	for _, p := range []string{perm, window.PermanentPrefix + address.NewID()} {
		if resp, _ := f.fetch(t, "localhost", p, slackbot, nil, nil); resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s by a robot = %d, want 404", p, resp.StatusCode)
		}
	}
}

// The card of a Window shows the title that its preview shows, and is
// found where its Window is.
func TestCards(t *testing.T) {
	f := setup(t)
	olivier := f.person(t, "localhost", "olivier")
	w := publishDeck(t, f, olivier)
	path := CardPath + w.ID + ".png"
	get := func(header http.Header) (response, []byte) {
		t.Helper()
		return f.fetch(t, "localhost", path, slackbot, nil, header)
	}

	resp, b := get(nil)
	if resp.StatusCode != http.StatusOK || resp.Header.Get("Content-Type") != "image/png" ||
		resp.Header.Get("ETag") != `"`+card.Tag("", "localhost")+`"` || resp.Header.Get("Cache-Control") != "public, max-age=3600" {
		t.Fatalf("GET %s = %d %v", path, resp.StatusCode, resp.Header)
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil || img.Bounds().Dx() != card.Width || img.Bounds().Dy() != card.Height {
		t.Fatalf("the card = %v, %v", img.Bounds(), err)
	}
	if resp, _ := get(http.Header{"If-None-Match": {resp.Header.Get("ETag")}}); resp.StatusCode != http.StatusNotModified {
		t.Errorf("GET %s again = %d, want 304", path, resp.StatusCode)
	}

	f.call(t, "localhost", olivier, "set_window_access", `{"window": "`+w.ID+`", "access": "public"}`)
	resp, titled := get(nil)
	if resp.StatusCode != http.StatusOK || resp.Header.Get("ETag") != `"`+card.Tag("Q3 review", "localhost")+`"` || bytes.Equal(titled, b) {
		t.Errorf("the card of a public Window = %d %q", resp.StatusCode, resp.Header.Get("ETag"))
	}

	f.call(t, "localhost", olivier, "trash_window", `{"window": "`+w.ID+`"}`)
	for host, p := range map[string]string{
		"localhost":      path,
		"beta.localhost": CardPath + address.NewID() + ".png",
	} {
		if resp, _ := f.fetch(t, host, p, slackbot, nil, nil); resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s on %s = %d, want 404", p, host, resp.StatusCode)
		}
	}
	f.call(t, "localhost", olivier, "restore_window", `{"window": "`+w.ID+`"}`)
	for _, p := range []string{CardPath + w.ID, CardPath + "garbage.png", CardPath + strings.ToUpper(w.ID) + ".png", CardPath + w.ID + ".png/x"} {
		if resp, _ := f.fetch(t, "localhost", p, slackbot, nil, nil); resp.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", p, resp.StatusCode)
		}
	}
}
