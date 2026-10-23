package housestyle

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"maps"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"cahut.cloud/internal/operation"
)

// SetInput is the input of set_house_style. What it leaves out stays as it
// is.
type SetInput struct {
	Neutral     bool         `json:"neutral,omitempty" jsonschema:"start again from the neutral house style, then apply the other fields"`
	Name        *string      `json:"name,omitempty" jsonschema:"the name of the organization; empty for the domain of the Cahut"`
	Colors      *ColorsInput `json:"colors,omitempty" jsonschema:"the colors to change; those left out stay"`
	ChartColors []string     `json:"chart_colors,omitempty" jsonschema:"up to 8 colors for the series of charts, in order, each as #rrggbb; an empty list goes back to those that Cahut derives from the primary and secondary colors"`
	HeadingFont *FontInput   `json:"heading_font,omitempty" jsonschema:"the font of headings, which replaces the current one with its files; with no family, headings take the body font"`
	BodyFont    *FontInput   `json:"body_font,omitempty" jsonschema:"the font of body text, which replaces the current one with its files; with no family, the font of the system"`
	Logo        *ImageInput  `json:"logo,omitempty" jsonschema:"the logo of the organization, for light backgrounds"`
	LogoInverse *ImageInput  `json:"logo_inverse,omitempty" jsonschema:"the version of the logo for dark backgrounds, such as slides in the primary color"`
	Guidance    *string      `json:"guidance,omitempty" jsonschema:"what the organization asks of AIs, at most 4000 characters: its tone of voice, its rules for titles or charts. Never readers see it"`
}

// ColorsInput is the colors of a house style that set_house_style changes.
type ColorsInput struct {
	Text       string `json:"text,omitempty" jsonschema:"body text, as #rrggbb"`
	Background string `json:"background,omitempty" jsonschema:"the background of pages, as #rrggbb; body text must contrast 4.5:1 at least on it"`
	Primary    string `json:"primary,omitempty" jsonschema:"the main color of the brand, as #rrggbb"`
	Secondary  string `json:"secondary,omitempty" jsonschema:"the second color of the brand, as #rrggbb"`
}

// FontInput is a font of a house style.
type FontInput struct {
	Family   string          `json:"family,omitempty" jsonschema:"the family, such as Inter or Georgia; empty for the default"`
	Fallback string          `json:"fallback,omitempty" jsonschema:"the generic family to fall back on: sans-serif (the default), serif or monospace"`
	Files    []FontFileInput `json:"files,omitempty" jsonschema:"the files of the family, which readers' browsers load; none for a family that their systems have, such as Georgia"`
}

// FontFileInput is a file of a font.
type FontFileInput struct {
	Content   string `json:"content,omitempty" jsonschema:"the file in base64: WOFF2, WOFF, TrueType or OpenType"`
	Keep      string `json:"keep,omitempty" jsonschema:"instead of content, the name of a file of the current house style to keep, as get_house_style lists it"`
	Weight    int    `json:"weight,omitempty" jsonschema:"its weight, from 100 to 900, 400 by default; for a variable font, the lightest weight it covers"`
	MaxWeight int    `json:"max_weight,omitempty" jsonschema:"for a variable font, the heaviest weight it covers"`
	Style     string `json:"style,omitempty" jsonschema:"normal (the default) or italic"`
}

// ImageInput is a logo.
type ImageInput struct {
	Content  string `json:"content" jsonschema:"the image: SVG as text, or PNG or WebP in base64; empty to remove the logo"`
	Encoding string `json:"encoding,omitempty" jsonschema:"utf-8 (the default, for SVG) or base64"`
}

// Limits of a house style.
const (
	MaxImageSize    = 512 << 10
	MaxFontFileSize = 1 << 20
	MaxFontFiles    = 16
	MaxSize         = 8 << 20 // all its files
	MaxNameLength   = 100     // characters
	MaxGuidance     = 4000    // characters
)

// set runs set_house_style.
func set(ctx context.Context, c *operation.Call, in SetInput, domain string) (HouseStyle, error) {
	c.Log("house style", "")
	if c.Actor.Limited() {
		return HouseStyle{}, operation.Forbiddenf("set_house_style changes the look of the whole Cahut, and this AI reaches some Rooms only: "+
			"%s can grant it every Room in the console of the Cahut, or change the house style there", c.Actor.Handle)
	}
	cur, prev, err := load(ctx, c.Tx, true)
	if err != nil {
		return HouseStyle{}, err
	}
	next := &version{files: map[string]file{}}
	if cur != nil && !in.Neutral {
		next.settings = cur.settings
		maps.Copy(next.files, cur.files)
	}
	kept := map[string]file{} // the files that a font may keep
	if cur != nil {
		kept = cur.files
	}
	var changes []string
	if in.Neutral {
		changes = append(changes, "neutral")
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if utf8.RuneCountInString(name) > MaxNameLength || strings.ContainsFunc(name, unicode.IsControl) {
			return HouseStyle{}, operation.Invalidf("name: give at most %d characters, on one line", MaxNameLength)
		}
		next.settings.Name = name
		changes = append(changes, fmt.Sprintf("name %q", name))
	}
	if in.Colors != nil {
		changed, err := setColors(&next.settings.Colors, in.Colors)
		if err != nil {
			return HouseStyle{}, err
		}
		changes = append(changes, "colors "+strings.Join(changed, ", "))
	}
	if in.ChartColors != nil {
		list, err := chartColors(in.ChartColors)
		if err != nil {
			return HouseStyle{}, err
		}
		next.settings.ChartColors = list
		changes = append(changes, fmt.Sprintf("%d chart colors", len(list)))
	}
	for _, f := range []struct {
		field string
		in    *FontInput
		role  string
		to    **font
	}{
		{"heading_font", in.HeadingFont, "heading", &next.settings.Heading},
		{"body_font", in.BodyFont, "body", &next.settings.Body},
	} {
		if f.in == nil {
			continue
		}
		set, files, err := setFont(f.field, f.role, f.in, kept)
		if err != nil {
			return HouseStyle{}, err
		}
		for name := range next.files {
			if strings.HasPrefix(name, fontsDir+f.role+"-") {
				delete(next.files, name)
			}
		}
		maps.Copy(next.files, files)
		*f.to = set
		what := f.role + " font by default"
		if set != nil && set.Family != "" {
			what = fmt.Sprintf("%s font %s, %d files", f.role, set.Family, len(set.Files))
		} else if set != nil {
			what = f.role + " font " + set.Fallback
		}
		changes = append(changes, what)
	}
	for _, l := range []struct {
		field, name string
		in          *ImageInput
	}{
		{"logo", logoFile, in.Logo},
		{"logo_inverse", logoInverseFile, in.LogoInverse},
	} {
		if l.in == nil {
			continue
		}
		f, err := image(l.field, l.in)
		if err != nil {
			return HouseStyle{}, err
		}
		if f == nil {
			delete(next.files, l.name)
			changes = append(changes, l.field+" removed")
			continue
		}
		next.files[l.name] = *f
		changes = append(changes, fmt.Sprintf("%s %s, %d bytes", l.field, f.typ, f.size))
	}
	if in.Guidance != nil {
		g := strings.TrimSpace(*in.Guidance)
		if utf8.RuneCountInString(g) > MaxGuidance {
			return HouseStyle{}, operation.Invalidf("guidance has %d characters, more than the %d allowed", utf8.RuneCountInString(g), MaxGuidance)
		}
		if strings.ContainsFunc(g, func(r rune) bool { return unicode.IsControl(r) && r != '\n' && r != '\t' }) {
			return HouseStyle{}, operation.Invalidf("guidance holds control characters: give plain text")
		}
		next.settings.Guidance = g
		changes = append(changes, fmt.Sprintf("guidance, %d characters", utf8.RuneCountInString(g)))
	}
	if len(changes) == 0 {
		return HouseStyle{}, operation.Invalidf("give what changes: name, colors, chart_colors, heading_font, body_font, logo, logo_inverse or guidance, or neutral")
	}
	if err := check(next); err != nil {
		return HouseStyle{}, err
	}
	c.Log("house style", strings.Join(changes, "; "))
	if cur != nil && same(cur, next) {
		return describe(cur, prev, domain), nil
	}
	if err := save(ctx, c.Tx, next, c.Actor.UserID, c.Now); err != nil {
		return HouseStyle{}, err
	}
	next.setAt, next.setBy = c.Now, c.Actor.Handle
	if cur == nil {
		cur = &version{setAt: c.Now}
	}
	return describe(next, cur, domain), nil
}

// rollback runs rollback_house_style.
func rollback(ctx context.Context, c *operation.Call, domain string) (HouseStyle, error) {
	c.Log("house style", "")
	if c.Actor.Limited() {
		return HouseStyle{}, operation.Forbiddenf("rollback_house_style changes the look of the whole Cahut, and this AI reaches some Rooms only")
	}
	cur, prev, err := load(ctx, c.Tx, false)
	if err != nil {
		return HouseStyle{}, err
	}
	if cur == nil || prev == nil {
		return HouseStyle{}, operation.Invalidf("this Cahut has no previous house style to go back to")
	}
	if err := swap(ctx, c.Tx, cur, prev); err != nil {
		return HouseStyle{}, err
	}
	back := "back to the neutral house style"
	if !prev.neutral() {
		back = "back to the house style set at " + prev.setAt.Format("2006-01-02 15:04 UTC")
		if prev.setBy != "" {
			back += " for " + prev.setBy
		}
	}
	c.Log("house style", back)
	return describe(prev, cur, domain), nil
}

// same reports whether two versions hold the same house style.
func same(a, b *version) bool {
	ja, _ := json.Marshal(a.settings)
	jb, _ := json.Marshal(b.settings)
	if !bytes.Equal(ja, jb) || len(a.files) != len(b.files) {
		return false
	}
	for name, f := range a.files {
		if g, ok := b.files[name]; !ok || g.sha256 != f.sha256 {
			return false
		}
	}
	return true
}

// setColors changes the colors that in gives, and returns their names.
func setColors(to *colors, in *ColorsInput) ([]string, error) {
	var changed []string
	for _, c := range []struct {
		name, value string
		to          *string
	}{
		{"text", in.Text, &to.Text},
		{"background", in.Background, &to.Background},
		{"primary", in.Primary, &to.Primary},
		{"secondary", in.Secondary, &to.Secondary},
	} {
		if c.value == "" {
			continue
		}
		v, ok := parseColor(c.value)
		if !ok {
			return nil, operation.Invalidf("colors.%s %q: give a color as #rrggbb, such as #0b5fff", c.name, c.value)
		}
		*c.to = v.String()
		changed = append(changed, c.name+" "+v.String())
	}
	if len(changed) == 0 {
		return nil, operation.Invalidf("colors: give text, background, primary or secondary, as #rrggbb")
	}
	return changed, nil
}

func chartColors(in []string) ([]string, error) {
	if len(in) > MaxChartColors {
		return nil, operation.Invalidf("chart_colors lists %d colors, more than the %d allowed", len(in), MaxChartColors)
	}
	out := make([]string, 0, len(in))
	for i, s := range in {
		v, ok := parseColor(s)
		if !ok {
			return nil, operation.Invalidf("chart_colors[%d] %q: give a color as #rrggbb, such as #0b5fff", i, s)
		}
		out = append(out, v.String())
	}
	return out, nil
}

// validFamily matches the name of a font family: letters, digits, spaces
// and hyphens, which CSS reads safely between quotes.
var validFamily = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9 -]{0,62}[A-Za-z0-9])?$`)

// generic are the names of the generic families of CSS, and its keywords,
// which no family takes.
var generic = []string{"serif", "sans-serif", "monospace", "cursive", "fantasy", "system-ui", "math", "emoji", "fangsong",
	"ui-serif", "ui-sans-serif", "ui-monospace", "ui-rounded", "inherit", "initial", "unset", "revert", "default"}

// setFont checks the font in, under the field it was given in, and returns
// it with its files, named after role. A font without a family is the
// default: nil.
func setFont(field, role string, in *FontInput, current map[string]file) (*font, map[string]file, error) {
	family := strings.Join(strings.Fields(in.Family), " ")
	switch in.Fallback {
	case "", "sans-serif", "serif", "monospace":
	default:
		return nil, nil, operation.Invalidf("%s.fallback %q: use sans-serif, serif or monospace", field, in.Fallback)
	}
	if family == "" {
		if len(in.Files) > 0 {
			return nil, nil, operation.Invalidf("%s.files: give the family of these files", field)
		}
		if in.Fallback == "" || in.Fallback == "sans-serif" {
			return nil, nil, nil
		}
		return &font{Fallback: in.Fallback}, nil, nil
	}
	if !validFamily.MatchString(family) || slices.Contains(generic, strings.ToLower(family)) {
		return nil, nil, operation.Invalidf("%s.family %q: give the name of a family, such as Inter, in letters, digits, spaces and hyphens", field, in.Family)
	}
	f := &font{Family: family, Fallback: in.Fallback}
	files := map[string]file{}
	seen := map[string]bool{}
	for i, fi := range in.Files {
		at := fmt.Sprintf("%s.files[%d]", field, i)
		ff := fontFile{Weight: fi.Weight, MaxWeight: fi.MaxWeight, Style: fi.Style}
		if ff.Weight == 0 {
			ff.Weight = 400
		}
		if ff.Style == "" {
			ff.Style = "normal"
		}
		switch {
		case ff.Weight < 100 || ff.Weight > 900:
			return nil, nil, operation.Invalidf("%s.weight %d: use 100 to 900", at, ff.Weight)
		case ff.MaxWeight != 0 && (ff.MaxWeight <= ff.Weight || ff.MaxWeight > 900):
			return nil, nil, operation.Invalidf("%s.max_weight %d: use a weight above weight, up to 900", at, ff.MaxWeight)
		case ff.Style != "normal" && ff.Style != "italic":
			return nil, nil, operation.Invalidf("%s.style %q: use normal or italic", at, ff.Style)
		}
		var content []byte
		switch {
		case fi.Content != "" && fi.Keep != "":
			return nil, nil, operation.Invalidf("%s: give content or keep, not both", at)
		case fi.Keep != "":
			kept, ok := current[fontsDir+fi.Keep]
			if !ok {
				return nil, nil, operation.Invalidf("%s.keep %q: the house style has no such font file; get_house_style lists them", at, fi.Keep)
			}
			content = kept.content
		case fi.Content != "":
			var err error
			if content, err = decode64(fi.Content); err != nil {
				return nil, nil, operation.Invalidf("%s.content: give the file in base64", at)
			}
		default:
			return nil, nil, operation.Invalidf("%s: give its content in base64, or the name of a file to keep", at)
		}
		if len(content) > MaxFontFileSize {
			return nil, nil, operation.Invalidf("%s is %d bytes, more than the %d allowed", at, len(content), MaxFontFileSize)
		}
		format, typ, ext := fontFormat(content)
		if format == "" {
			return nil, nil, operation.Invalidf("%s is not a font: give a WOFF2, WOFF, TrueType or OpenType file", at)
		}
		ff.Format = format
		ff.Name = fmt.Sprintf("%s-%s", role, strings.ReplaceAll(FontFile{Weight: ff.Weight, MaxWeight: ff.MaxWeight}.weights(), " ", "-"))
		if ff.Style == "italic" {
			ff.Name += "-italic"
		}
		if seen[ff.Name] {
			return nil, nil, operation.Invalidf("%s: another file of the family has the same weight and style", at)
		}
		seen[ff.Name] = true
		ff.Name += ext
		f.Files = append(f.Files, ff)
		files[fontsDir+ff.Name] = newFile(typ, content)
	}
	return f, files, nil
}

// fontFormat recognizes a font file by its first bytes, and returns its
// format as @font-face names it, its media type and its extension.
func fontFormat(b []byte) (format, typ, ext string) {
	switch {
	case bytes.HasPrefix(b, []byte("wOF2")):
		return "woff2", "font/woff2", ".woff2"
	case bytes.HasPrefix(b, []byte("wOFF")):
		return "woff", "font/woff", ".woff"
	case bytes.HasPrefix(b, []byte{0, 1, 0, 0}), bytes.HasPrefix(b, []byte("true")):
		return "truetype", "font/ttf", ".ttf"
	case bytes.HasPrefix(b, []byte("OTTO")):
		return "opentype", "font/otf", ".otf"
	}
	return "", "", ""
}

// image checks a logo given under field, and returns it as a file, or nil
// when it is given empty, to remove it.
func image(field string, in *ImageInput) (*file, error) {
	var content []byte
	switch in.Encoding {
	case "", "utf-8":
		content = []byte(in.Content)
	case "base64":
		var err error
		if content, err = decode64(in.Content); err != nil {
			return nil, operation.Invalidf("%s.content: the content is not valid base64", field)
		}
	default:
		return nil, operation.Invalidf("%s.encoding %q: use utf-8 or base64", field, in.Encoding)
	}
	if len(content) == 0 {
		return nil, nil
	}
	if len(content) > MaxImageSize {
		return nil, operation.Invalidf("%s is %d bytes, more than the %d allowed", field, len(content), MaxImageSize)
	}
	typ := imageType(content)
	if typ == "" {
		return nil, operation.Invalidf("%s is neither SVG, PNG nor WebP: give SVG as text, or PNG or WebP in base64 with encoding base64", field)
	}
	f := newFile(typ, content)
	return &f, nil
}

// imageType recognizes a logo by its content, and returns its media type.
func imageType(b []byte) string {
	switch {
	case bytes.HasPrefix(b, []byte("\x89PNG\r\n\x1a\n")):
		return "image/png"
	case len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP")):
		return "image/webp"
	case utf8.Valid(b) && isSVG(b):
		return "image/svg+xml"
	}
	return ""
}

// isSVG reports whether b is a well-formed XML document whose root is an
// svg element.
func isSVG(b []byte) bool {
	d := xml.NewDecoder(bytes.NewReader(b))
	root := false
	for {
		t, err := d.Token()
		if errors.Is(err, io.EOF) {
			return root
		}
		if err != nil {
			return false
		}
		if s, ok := t.(xml.StartElement); ok && !root {
			if s.Name.Local != "svg" {
				return false
			}
			root = true
		}
	}
}

func decode64(s string) ([]byte, error) {
	s = strings.Join(strings.Fields(s), "")
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		b, err = base64.RawStdEncoding.DecodeString(strings.TrimRight(s, "="))
	}
	return b, err
}

func newFile(typ string, content []byte) file {
	sum := sha256.Sum256(content)
	return file{typ: typ, size: int64(len(content)), sha256: hex.EncodeToString(sum[:]), content: content}
}

// check refuses a house style whose body text does not read on its
// background, or whose files are too many or too large.
func check(v *version) error {
	p, _ := palette(v.settings)
	text, background := mustColor(p.Text), mustColor(p.Background)
	if r := contrast(text, background); r < MinContrast {
		return operation.Invalidf("colors.text %s contrasts %.1f:1 on colors.background %s; body text needs %.1f:1 at least: "+
			"choose a text and a background further apart", p.Text, r, p.Background, MinContrast)
	}
	var fonts int
	var size int64
	for name, f := range v.files {
		if strings.HasPrefix(name, fontsDir) {
			fonts++
		}
		size += f.size
	}
	if fonts > MaxFontFiles {
		return operation.Invalidf("the fonts have %d files, more than the %d allowed", fonts, MaxFontFiles)
	}
	if size > MaxSize {
		return operation.Invalidf("the files of the house style make %d bytes, more than the %d allowed", size, MaxSize)
	}
	return nil
}
