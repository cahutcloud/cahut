package housestyle

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"math"
	"unicode/utf8"

	"cahut.cloud/internal/store"
)

// Asset is a file of the house style, as Cahut serves it.
type Asset struct {
	Type    string
	ETag    string // quoted, as the header carries it
	Content []byte
}

func asset(typ string, content []byte) Asset {
	sum := sha256.Sum256(content)
	return Asset{Type: typ, ETag: `"` + hex.EncodeToString(sum[:16]) + `"`, Content: content}
}

// Brand is what the pages of Cahut show of the house style at their top.
type Brand struct {
	Name string // the name of the organization, or the domain of the Cahut
	Logo bool   // whether the house style has a logo
	// NeutralLook reports whether the house style keeps the look of the
	// neutral one: its pages then carry the brand of Cahut. See
	// docs/adr/0033-the-brand-of-cahut-until-the-house-style-has-a-look.md.
	NeutralLook bool
}

// LoadBrand returns the brand of the Cahut at domain, whose database is db.
func LoadBrand(ctx context.Context, db *store.DB, domain string) (Brand, error) {
	cur, err := loadState(ctx, db.Read, "current", false)
	if err != nil {
		return Brand{}, err
	}
	b := Brand{Name: domain, NeutralLook: cur == nil || cur.neutralLook()}
	if cur != nil {
		if cur.settings.Name != "" {
			b.Name = cur.settings.Name
		}
		_, b.Logo = cur.files[logoFile]
	}
	return b, nil
}

// Sheet returns the style sheet of the house style of the Cahut at domain:
// its variables and the @font-face rules of its fonts, then the classes of
// the kit.
func Sheet(ctx context.Context, db *store.DB, domain string) (Asset, error) {
	cur, err := loadState(ctx, db.Read, "current", false)
	if err != nil {
		return Asset{}, err
	}
	return asset("text/css; charset=utf-8", sheet(describe(cur, nil, domain))), nil
}

// sheet writes the style sheet of h.
func sheet(h HouseStyle) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "/* The house style of %s, served by Cahut. Link it from the root of the domain:\n", cssComment(h.Name))
	fmt.Fprintf(&b, "   <link rel=\"stylesheet\" href=\"%s\">. get_house_style describes its variables and classes. */\n\n", SheetPath)
	seen := map[string]bool{}
	for _, f := range []Font{h.Fonts.Heading, h.Fonts.Body} {
		for _, ff := range f.Files {
			if seen[ff.Name] {
				continue
			}
			seen[ff.Name] = true
			fmt.Fprintf(&b, "@font-face {\n\tfont-family: %q;\n\tsrc: url(%q) format(%q);\n\tfont-weight: %s;\n\tfont-style: %s;\n\tfont-display: swap;\n}\n\n",
				f.Family, ff.URL, formatOf(ff.Name), ff.weights(), ff.Style)
		}
	}
	p := h.Colors
	scheme := "light"
	if mustColor(p.Background).dark() {
		scheme = "dark"
	}
	b.WriteString(":root {\n")
	for _, v := range []struct{ name, value string }{
		{"text", p.Text}, {"background", p.Background},
		{"primary", p.Primary}, {"on-primary", p.OnPrimary},
		{"secondary", p.Secondary}, {"on-secondary", p.OnSecondary},
		{"muted", p.Muted}, {"line", p.Line}, {"surface", p.Surface},
	} {
		fmt.Fprintf(&b, "\t--hs-%s: %s;\n", v.name, v.value)
	}
	for i, c := range h.ChartColors {
		fmt.Fprintf(&b, "\t--hs-chart-%d: %s;\n", i+1, c)
	}
	fmt.Fprintf(&b, "\t--hs-font-heading: %s;\n\t--hs-font-body: %s;\n\t--hs-font-mono: %s;\n", h.Fonts.Heading.Stack, h.Fonts.Body.Stack, h.Fonts.Mono)
	fmt.Fprintf(&b, "\t--hs-logo: url(%q);\n\t--hs-logo-inverse: url(%q);\n\t--hs-scheme: %s;\n}\n", LogoPath, LogoInversePath, scheme)
	b.WriteString("\n")
	b.Write(kitCSS)
	return b.Bytes()
}

// formatOf returns the format of a font file, as @font-face names it, from
// the extension that its name has.
func formatOf(name string) string {
	for _, f := range []struct{ ext, format string }{{".woff2", "woff2"}, {".woff", "woff"}, {".ttf", "truetype"}, {".otf", "opentype"}} {
		if len(name) > len(f.ext) && name[len(name)-len(f.ext):] == f.ext {
			return f.format
		}
	}
	return "woff2"
}

// cssComment makes s safe in a CSS comment.
func cssComment(s string) string {
	return string(bytes.ReplaceAll([]byte(s), []byte("*/"), []byte("* /")))
}

// Logo returns the logo of the house style of the Cahut at domain, or its
// version for dark backgrounds when inverse is set. Without one, it returns
// the name of the organization set in SVG, in light letters for dark
// backgrounds: a logo made for light backgrounds would not read there.
func Logo(ctx context.Context, db *store.DB, domain string, inverse bool) (Asset, error) {
	name := logoFile
	if inverse {
		name = logoInverseFile
	}
	f, ok, err := loadFile(ctx, db.Read, name)
	if err != nil {
		return Asset{}, err
	}
	if ok {
		return Asset{Type: f.typ, ETag: `"` + f.sha256[:32] + `"`, Content: f.content}, nil
	}
	cur, err := loadState(ctx, db.Read, "current", false)
	if err != nil {
		return Asset{}, err
	}
	h := describe(cur, nil, domain)
	fill := h.Colors.Text
	if inverse {
		// The lighter of the text and the background.
		if mustColor(h.Colors.Background).luminance() > mustColor(h.Colors.Text).luminance() {
			fill = h.Colors.Background
		}
	}
	return asset("image/svg+xml", wordmark(h.Name, h.Fonts.Heading.Stack, fill)), nil
}

// wordmark sets name in SVG, in the font of stack and the color fill: the
// logo of a house style that has none. An image in SVG loads no web font,
// so it takes the font from the reader's system when it has it.
func wordmark(name, stack, fill string) []byte {
	const size = 40
	n := utf8.RuneCountInString(name)
	width := int(math.Ceil(float64(n)*0.6*size)) + 4
	var esc bytes.Buffer
	_ = xml.EscapeText(&esc, []byte(name))
	var family bytes.Buffer
	_ = xml.EscapeText(&family, []byte(stack))
	return fmt.Appendf(nil, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d 56" width="%d" height="56" role="img" aria-label="%s">`+
		`<text x="2" y="41" font-family="%s" font-size="%d" font-weight="700" fill="%s" textLength="%d" lengthAdjust="spacing">%s</text></svg>`,
		width, width, esc.String(), family.String(), size, fill, width-4, esc.String())
}

// LoadFont returns the font file called name of the house style, or false
// when it has none.
func LoadFont(ctx context.Context, db *store.DB, name string) (Asset, bool, error) {
	f, ok, err := loadFile(ctx, db.Read, fontsDir+name)
	if err != nil || !ok {
		return Asset{}, false, err
	}
	return Asset{Type: f.typ, ETag: `"` + f.sha256[:32] + `"`, Content: f.content}, true, nil
}
