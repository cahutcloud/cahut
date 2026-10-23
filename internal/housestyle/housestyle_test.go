package housestyle

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/store"
)

type fixture struct {
	r        *operation.Registry
	db       *store.DB
	admin    operation.Actor // Olivier, an administrator of the Cahut
	member   operation.Actor // Camille, a member
	ai       operation.Actor // Claude Code, for Olivier
	memberAI operation.Actor
}

func setup(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "cahut.db"), store.CahutSchema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	r := operation.NewRegistry(db)
	account.Register(r)
	room.Register(r)
	Register(r, "cahut.test")
	f := &fixture{r: r, db: db}
	cli := operation.Actor{Kind: operation.Admin, Via: "cli"}
	for _, p := range []struct {
		handle string
		admin  bool
	}{{"olivier", true}, {"camille", false}} {
		u, err := operation.Do(ctx, r, account.CreateUser, cli, account.CreateUserInput{Handle: p.handle, Name: p.handle, Admin: p.admin})
		if err != nil {
			t.Fatal(err)
		}
		tok, err := operation.Do(ctx, r, account.CreateToken, cli, account.CreateTokenInput{User: p.handle, Name: "Claude Code"})
		if err != nil {
			t.Fatal(err)
		}
		person := operation.Actor{Kind: operation.Person, Via: "api", UserID: u.ID, Handle: p.handle, Session: 1}
		ai := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: u.ID, Handle: p.handle, TokenID: tok.ID, Client: "Claude Code"}
		if p.admin {
			f.admin, f.ai = person, ai
		} else {
			f.member, f.memberAI = person, ai
		}
	}
	return f
}

func (f *fixture) call(t *testing.T, actor operation.Actor, name, args string) (HouseStyle, error) {
	t.Helper()
	out, err := f.r.Invoke(context.Background(), name, actor, []byte(args))
	if err != nil {
		return HouseStyle{}, err
	}
	return out.(HouseStyle), nil
}

func (f *fixture) must(t *testing.T, actor operation.Actor, name, args string) HouseStyle {
	t.Helper()
	h, err := f.call(t, actor, name, args)
	if err != nil {
		t.Fatalf("%s %s: %v", name, args, err)
	}
	return h
}

func b64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

const svg = `<?xml version="1.0"?><!-- the logo --><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><rect width="10" height="10"/></svg>`

func TestNeutral(t *testing.T) {
	f := setup(t)
	h := f.must(t, f.member, "get_house_style", ``)
	if !h.Neutral || h.Name != "cahut.test" || h.HasLogo || h.Previous != nil || h.SetAt != nil {
		t.Fatalf("neutral house style = %+v", h)
	}
	if h.Colors.Text != neutral.text || h.Colors.Background != neutral.background || h.Colors.Primary != neutral.primary {
		t.Errorf("colors = %+v", h.Colors)
	}
	if !slices.Equal(h.ChartColors, neutral.chart) {
		t.Errorf("chart colors = %v, want the neutral palette", h.ChartColors)
	}
	if h.Fonts.Body.Family != "" || h.Fonts.Body.Stack != stacks["sans-serif"] || h.Fonts.Heading.Stack != stacks["sans-serif"] {
		t.Errorf("fonts = %+v", h.Fonts)
	}
	if h.Stylesheet != "/-/house-style.css" || h.Logo != "/-/house-style/logo" || len(h.Warnings) != 0 {
		t.Errorf("house style = %+v", h)
	}
	if r := contrast(mustColor(h.Colors.Muted), mustColor(h.Colors.Background)); r < MinContrast {
		t.Errorf("muted %s contrasts %.2f:1", h.Colors.Muted, r)
	}
	if r := contrast(mustColor(h.Colors.OnPrimary), mustColor(h.Colors.Primary)); r < MinContrast {
		t.Errorf("on primary %s contrasts %.2f:1", h.Colors.OnPrimary, r)
	}
}

func TestSetChangesWhatItNames(t *testing.T) {
	f := setup(t)
	h := f.must(t, f.admin, "set_house_style", `{"name": " Example Co ", "colors": {"primary": "#0B5FFF", "secondary": "#f70"}, "guidance": "Sentence case for titles."}`)
	if h.Neutral || h.Name != "Example Co" || h.Colors.Primary != "#0b5fff" || h.Colors.Secondary != "#ff7700" ||
		h.Colors.Text != neutral.text || h.Guidance != "Sentence case for titles." {
		t.Fatalf("house style = %+v", h)
	}
	if h.SetBy != "olivier" || h.SetAt == nil || h.Previous == nil || !h.Previous.Neutral {
		t.Errorf("set by %q at %v, previous %+v", h.SetBy, h.SetAt, h.Previous)
	}
	if h.ChartColors[0] != "#0b5fff" || h.ChartColors[1] != "#ff7700" || len(h.ChartColors) != MaxChartColors {
		t.Errorf("chart colors = %v, want the primary and the secondary colors first", h.ChartColors)
	}
	first := h.ChartColors[:5]
	for i, a := range first {
		for _, b := range first[i+1:] {
			if distance(mustColor(a), mustColor(b)) < distinct {
				t.Errorf("chart colors %s and %s are too close", a, b)
			}
		}
	}

	// A change of one color keeps the rest.
	h = f.must(t, f.admin, "set_house_style", `{"colors": {"background": "#fbfaf7"}}`)
	if h.Name != "Example Co" || h.Colors.Primary != "#0b5fff" || h.Colors.Background != "#fbfaf7" || h.Previous.Neutral {
		t.Fatalf("after a change of background, house style = %+v", h)
	}
	h = f.must(t, f.member, "get_house_style", ``)
	if h.Colors.Background != "#fbfaf7" || h.Guidance == "" {
		t.Errorf("get_house_style = %+v", h)
	}

	// The same house style again makes no new version.
	h = f.must(t, f.admin, "set_house_style", `{"colors": {"background": "#fbfaf7"}}`)
	if h.Previous == nil || h.Previous.Neutral {
		t.Errorf("previous after setting the same house style = %+v", h.Previous)
	}
	if n := f.count(t); n != 2 {
		t.Errorf("%d house styles kept, want 2", n)
	}

	// Neutral starts again from the neutral house style.
	h = f.must(t, f.admin, "set_house_style", `{"neutral": true, "name": "Example Co"}`)
	if h.Name != "Example Co" || h.Colors.Primary != neutral.primary || h.Guidance != "" || h.Neutral {
		t.Errorf("after neutral, house style = %+v", h)
	}
	h = f.must(t, f.admin, "set_house_style", `{"name": ""}`)
	if h.Name != "cahut.test" {
		t.Errorf("an empty name gives %q, want the domain", h.Name)
	}

	h = f.must(t, f.admin, "set_house_style", `{"chart_colors": ["#111111", "#ee0000"]}`)
	if h.ChartColors[0] != "#111111" || h.ChartColors[1] != "#ee0000" || len(h.ChartColors) != MaxChartColors {
		t.Errorf("chart colors = %v", h.ChartColors)
	}
	h = f.must(t, f.admin, "set_house_style", `{"chart_colors": []}`)
	if !slices.Equal(h.ChartColors, neutral.chart) {
		t.Errorf("an empty list of chart colors gives %v, want those that Cahut derives", h.ChartColors)
	}
}

func (f *fixture) count(t *testing.T) int {
	t.Helper()
	var n int
	if err := f.db.Read.QueryRowContext(context.Background(), "SELECT count(*) FROM house_styles").Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestColorsMustRead(t *testing.T) {
	f := setup(t)
	for _, tc := range []struct{ args, want string }{
		{`{"colors": {"text": "#9a9a9a"}}`, "contrasts 2.8:1"},
		{`{"colors": {"primary": "blue"}}`, `colors.primary "blue": give a color as #rrggbb`},
		{`{"colors": {}}`, "colors: give text, background, primary or secondary"},
		{`{"chart_colors": ["#000", "red"]}`, `chart_colors[1] "red"`},
		{`{"chart_colors": ["#000", "#111", "#222", "#333", "#444", "#555", "#666", "#777", "#888"]}`, "more than the 8 allowed"},
		{`{}`, "give what changes"},
		{`{"name": "Line\nbreak"}`, "on one line"},
		{`{"guidance": "` + strings.Repeat("a", MaxGuidance+1) + `"}`, "more than the 4000 allowed"},
		{`{"nom": "Example"}`, "unknown field"},
	} {
		_, err := f.call(t, f.admin, "set_house_style", tc.args)
		if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("set_house_style %.60s: error %v, want %q", tc.args, err, tc.want)
		}
	}
	if n := f.count(t); n != 0 {
		t.Errorf("%d house styles after refused calls, want none", n)
	}

	h := f.must(t, f.admin, "set_house_style", `{"colors": {"primary": "#ffe14d"}}`)
	if len(h.Warnings) != 1 || !strings.Contains(h.Warnings[0], "colors.primary #ffe14d contrasts 1.3:1 on the background") {
		t.Errorf("warnings for a yellow primary color = %q", h.Warnings)
	}
	if h.Colors.OnPrimary != neutral.text {
		t.Errorf("text on a yellow primary color = %s, want the text color", h.Colors.OnPrimary)
	}

	// A dark house style.
	h = f.must(t, f.admin, "set_house_style", `{"colors": {"text": "#f2f2f2", "background": "#121417", "primary": "#7cb1ff"}}`)
	if h.Colors.OnPrimary != "#121417" || len(h.Warnings) != 0 && !strings.Contains(strings.Join(h.Warnings, " "), "secondary") {
		t.Errorf("dark house style = %+v", h)
	}
}

func TestFonts(t *testing.T) {
	f := setup(t)
	woff2 := b64("wOF2 heading")
	h := f.must(t, f.admin, "set_house_style", fmt.Sprintf(`{
		"heading_font": {"family": "Fraunces", "fallback": "serif", "files": [
			{"content": %q, "weight": 600}, {"content": %q, "weight": 600, "style": "italic"}]},
		"body_font": {"family": "Inter", "files": [{"content": %q, "weight": 100, "max_weight": 900}]}}`,
		woff2, b64("wOFF heading italic"), b64("\x00\x01\x00\x00 body")))
	heading, body := h.Fonts.Heading, h.Fonts.Body
	if heading.Family != "Fraunces" || heading.Stack != `"Fraunces", `+stacks["serif"] || len(heading.Files) != 2 {
		t.Fatalf("heading font = %+v", heading)
	}
	if heading.Files[0].Name != "heading-600.woff2" || heading.Files[0].URL != "/-/house-style/fonts/heading-600.woff2" ||
		heading.Files[1].Name != "heading-600-italic.woff" || heading.Files[0].Size != 12 {
		t.Errorf("heading files = %+v", heading.Files)
	}
	if body.Family != "Inter" || body.Fallback != "sans-serif" || len(body.Files) != 1 || body.Files[0].Name != "body-100-900.ttf" {
		t.Errorf("body font = %+v", body)
	}

	css := string(sheet(h))
	for _, want := range []string{
		"@font-face {\n\tfont-family: \"Fraunces\";\n\tsrc: url(\"/-/house-style/fonts/heading-600.woff2\") format(\"woff2\");\n\tfont-weight: 600;\n\tfont-style: normal;",
		"src: url(\"/-/house-style/fonts/heading-600-italic.woff\") format(\"woff\");",
		"font-weight: 100 900;",
		"--hs-font-heading: \"Fraunces\", Georgia",
		"--hs-font-body: \"Inter\", system-ui",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("the sheet lacks %q:\n%s", want, css)
		}
	}

	// A font file kept by its name, and the others of its role dropped.
	h = f.must(t, f.admin, "set_house_style", `{"heading_font": {"family": "Fraunces", "files": [{"keep": "heading-600.woff2", "weight": 700}]}}`)
	if files := h.Fonts.Heading.Files; len(files) != 1 || files[0].Name != "heading-700.woff2" || files[0].Size != 12 {
		t.Errorf("kept heading files = %+v", files)
	}
	if len(h.Fonts.Body.Files) != 1 {
		t.Errorf("the body font lost its files: %+v", h.Fonts.Body)
	}
	a, ok, err := LoadFont(context.Background(), f.db, "heading-700.woff2")
	if err != nil || !ok || string(a.Content) != "wOF2 heading" || a.Type != "font/woff2" {
		t.Errorf("LoadFont = %q, %v, %v", a.Content, ok, err)
	}
	if _, ok, _ := LoadFont(context.Background(), f.db, "heading-600-italic.woff"); ok {
		t.Error("a font file left out is still served")
	}

	// No family: headings take the body font.
	h = f.must(t, f.admin, "set_house_style", `{"heading_font": {}}`)
	if h.Fonts.Heading.Family != "Inter" {
		t.Errorf("heading font without a family = %+v, want the body font", h.Fonts.Heading)
	}

	for _, tc := range []struct{ args, want string }{
		{`{"body_font": {"family": "Inter\"; } body { color: red"}}`, "give the name of a family"},
		{`{"body_font": {"family": "serif"}}`, "give the name of a family"},
		{`{"body_font": {"files": [{"content": "` + woff2 + `"}]}}`, "give the family of these files"},
		{`{"body_font": {"family": "Inter", "fallback": "cursive"}}`, "use sans-serif, serif or monospace"},
		{`{"body_font": {"family": "Inter", "files": [{"content": "` + b64("<html>") + `"}]}}`, "is not a font"},
		{`{"body_font": {"family": "Inter", "files": [{"content": "` + woff2 + `"}, {"content": "` + woff2 + `", "weight": 400}]}}`, "same weight and style"},
		{`{"body_font": {"family": "Inter", "files": [{"content": "` + woff2 + `", "weight": 950}]}}`, "use 100 to 900"},
		{`{"body_font": {"family": "Inter", "files": [{"content": "` + woff2 + `", "style": "oblique"}]}}`, "use normal or italic"},
		{`{"body_font": {"family": "Inter", "files": [{"keep": "body-200.woff2"}]}}`, "no such font file"},
		{`{"body_font": {"family": "Inter", "files": [{}]}}`, "give its content in base64"},
		{`{"body_font": {"family": "Inter", "files": [{"content": "` + b64("wOF2"+strings.Repeat("x", MaxFontFileSize)) + `"}]}}`, "more than the 1048576 allowed"},
	} {
		_, err := f.call(t, f.admin, "set_house_style", tc.args)
		if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("set_house_style %.70s: error %v, want %q", tc.args, err, tc.want)
		}
	}
}

func TestLogos(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	a, err := Logo(ctx, f.db, "cahut.test", false)
	if err != nil || a.Type != "image/svg+xml" || !strings.Contains(string(a.Content), ">cahut.test</text>") {
		t.Fatalf("logo of the neutral house style = %s, %v", a.Content, err)
	}

	f.must(t, f.admin, "set_house_style", `{"name": "<A & B>"}`)
	a, _ = Logo(ctx, f.db, "cahut.test", false)
	if !strings.Contains(string(a.Content), `aria-label="&lt;A &amp; B&gt;"`) || !isSVG(a.Content) {
		t.Errorf("wordmark of a name to escape = %s", a.Content)
	}

	png := "\x89PNG\r\n\x1a\n rest of the image"
	h := f.must(t, f.admin, "set_house_style", fmt.Sprintf(`{"logo": {"content": %q}, "logo_inverse": {"content": %q, "encoding": "base64"}}`, svg, b64(png)))
	if !h.HasLogo || !h.HasLogoInverse {
		t.Fatalf("house style with logos = %+v", h)
	}
	a, _ = Logo(ctx, f.db, "cahut.test", false)
	b, _ := Logo(ctx, f.db, "cahut.test", true)
	if string(a.Content) != svg || a.Type != "image/svg+xml" || string(b.Content) != png || b.Type != "image/png" || a.ETag == b.ETag {
		t.Errorf("logos = %s (%s), %s (%s)", a.Content, a.Type, b.Content, b.Type)
	}
	brand, err := LoadBrand(ctx, f.db, "cahut.test")
	if err != nil || brand.Name != "<A & B>" || !brand.Logo {
		t.Errorf("brand = %+v, %v", brand, err)
	}

	h = f.must(t, f.admin, "set_house_style", `{"logo_inverse": {"content": ""}}`)
	b, _ = Logo(ctx, f.db, "cahut.test", true)
	if h.HasLogoInverse || !h.HasLogo || string(b.Content) != svg {
		t.Errorf("without a logo for dark backgrounds, it serves %s", b.Content)
	}

	for _, tc := range []struct{ args, want string }{
		{`{"logo": {"content": "<html><body/></html>"}}`, "neither SVG, PNG nor WebP"},
		{`{"logo": {"content": "<svg><g></svg>"}}`, "neither SVG, PNG nor WebP"},
		{`{"logo": {"content": "` + b64("GIF89a") + `", "encoding": "base64"}}`, "neither SVG, PNG nor WebP"},
		{`{"logo": {"content": "%%%", "encoding": "base64"}}`, "not valid base64"},
		{`{"logo": {"content": "x", "encoding": "latin1"}}`, "use utf-8 or base64"},
		{`{"logo": {"content": "<svg>` + strings.Repeat(" ", MaxImageSize) + `</svg>"}}`, "more than the 524288 allowed"},
	} {
		_, err := f.call(t, f.admin, "set_house_style", tc.args)
		if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("set_house_style %.60s: error %v, want %q", tc.args, err, tc.want)
		}
	}
}

func TestRollback(t *testing.T) {
	f := setup(t)
	if _, err := f.call(t, f.admin, "rollback_house_style", ``); !errors.Is(err, operation.ErrInvalid) {
		t.Fatalf("rollback of the neutral house style: %v", err)
	}
	f.must(t, f.admin, "set_house_style", fmt.Sprintf(`{"name": "First", "logo": {"content": %q}}`, svg))
	h := f.must(t, f.admin, "rollback_house_style", ``)
	if !h.Neutral || h.Name != "cahut.test" || h.Previous == nil || h.Previous.Neutral {
		t.Fatalf("back to neutral = %+v", h)
	}
	h = f.must(t, f.admin, "rollback_house_style", ``)
	if h.Name != "First" || !h.HasLogo || !h.Previous.Neutral {
		t.Fatalf("back to the first = %+v", h)
	}

	f.must(t, f.admin, "set_house_style", `{"name": "Second"}`)
	f.must(t, f.admin, "set_house_style", `{"name": "Third"}`)
	if n := f.count(t); n != 2 {
		t.Errorf("%d house styles kept, want 2", n)
	}
	var files int
	if err := f.db.Read.QueryRowContext(context.Background(), "SELECT count(*) FROM house_style_files").Scan(&files); err != nil || files != 2 {
		t.Errorf("%d files kept, want the logo of the second and of the third: %v", files, err)
	}
	h = f.must(t, f.admin, "rollback_house_style", ``)
	if h.Name != "Second" || h.Previous.SetBy != "olivier" {
		t.Errorf("back to the second = %+v", h)
	}
}

func TestRights(t *testing.T) {
	f := setup(t)
	set := `{"name": "Example Co"}`
	for _, a := range []operation.Actor{f.member, f.memberAI} {
		if _, err := f.call(t, a, "set_house_style", set); !errors.Is(err, operation.ErrForbidden) {
			t.Errorf("set_house_style by %s %s: %v", a.Kind, a.Handle, err)
		}
		if _, err := f.call(t, a, "rollback_house_style", ``); !errors.Is(err, operation.ErrForbidden) {
			t.Errorf("rollback_house_style by %s %s: %v", a.Kind, a.Handle, err)
		}
	}
	limited := f.ai
	limited.Rooms = []int64{1}
	if _, err := f.call(t, limited, "set_house_style", set); !errors.Is(err, operation.ErrForbidden) || !strings.Contains(err.Error(), "reaches some Rooms only") {
		t.Errorf("set_house_style by an AI limited to some Rooms: %v", err)
	}
	readOnly := f.ai
	readOnly.ReadOnly = true
	if _, err := f.call(t, readOnly, "set_house_style", set); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("set_house_style by an AI that may only read: %v", err)
	}
	for _, a := range []operation.Actor{limited, readOnly, f.memberAI} {
		if _, err := f.call(t, a, "get_house_style", ``); err != nil {
			t.Errorf("get_house_style by %+v: %v", a, err)
		}
	}

	// The AI of an administrator changes it at once: it can be undone.
	h := f.must(t, f.ai, "set_house_style", set+``)
	if h.Name != "Example Co" || h.SetBy != "olivier" {
		t.Errorf("set by the AI of an administrator = %+v", h)
	}
	cli := operation.Actor{Kind: operation.Admin, Via: "cli"}
	if h := f.must(t, cli, "set_house_style", `{"name": "From the command line"}`); h.SetBy != "" || h.Previous.SetBy != "olivier" {
		t.Errorf("set from the command line = %+v", h)
	}
}

func TestLogbook(t *testing.T) {
	f := setup(t)
	logo := strings.Replace(svg, "<rect", "<rect id=\"secret-content\"", 1)
	f.must(t, f.ai, "set_house_style", fmt.Sprintf(`{"name": "Example Co", "colors": {"primary": "#0b5fff"},
		"body_font": {"family": "Inter", "files": [{"content": %q}]}, "logo": {"content": %q}, "guidance": "Keep it short."}`, b64("wOF2 secret-content"), logo))
	f.must(t, f.ai, "get_house_style", ``)
	f.must(t, f.admin, "rollback_house_style", ``)
	entries, err := logbook.List(context.Background(), f.db.Read, 10)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(entries)
	if strings.Contains(string(b), "secret-content") || strings.Contains(string(b), "Keep it short") {
		t.Errorf("the Logbook holds the contents of files or the guidance: %s", b)
	}
	if entries[1].Operation != "get_house_style" {
		t.Errorf("entry of get_house_style = %+v", entries[1])
	}
	if e := entries[2]; e.Operation != "set_house_style" || e.Target != "house style" ||
		e.Detail != fmt.Sprintf(`name "Example Co"; colors primary #0b5fff; body font Inter, 1 files; logo image/svg+xml, %d bytes; guidance, 14 characters`, len(logo)) {
		t.Errorf("entry of set_house_style = %+v", e)
	}
	if e := entries[0]; e.Operation != "rollback_house_style" || e.Detail != "back to the neutral house style" {
		t.Errorf("entry of rollback_house_style = %+v", e)
	}
}

func TestSheetFollowsTheHouseStyle(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	a, err := Sheet(ctx, f.db, "cahut.test")
	if err != nil {
		t.Fatal(err)
	}
	css := string(a.Content)
	for _, want := range []string{"--hs-text: #1d1f23;", "--hs-primary: #2a5bd7;", "--hs-on-primary: #ffffff;", "--hs-chart-8: #e34948;",
		"--hs-font-body: system-ui", "--hs-logo: url(\"/-/house-style/logo\");", "--hs-scheme: light;"} {
		if !strings.Contains(css, want) {
			t.Errorf("the neutral sheet lacks %q:\n%s", want, css)
		}
	}
	if strings.Contains(css, "@font-face") || a.Type != "text/css; charset=utf-8" {
		t.Errorf("neutral sheet = %s", css)
	}
	f.must(t, f.admin, "set_house_style", `{"name": "End */ of comment", "colors": {"text": "#eeeeee", "background": "#101010", "primary": "#80b0ff"}}`)
	b, _ := Sheet(ctx, f.db, "cahut.test")
	if b.ETag == a.ETag || !strings.Contains(string(b.Content), "--hs-primary: #80b0ff;") || !strings.Contains(string(b.Content), "--hs-scheme: dark;") ||
		strings.Count(string(b.Content), "*/") != 1 {
		t.Errorf("sheet after a change = %s", b.Content)
	}
}
