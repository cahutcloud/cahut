// Package housestyle keeps the house style of a Cahut: the name of its
// organization, its colors, its fonts, its logo and guidance for AIs. Cahut
// serves it under /-/ to the Windows that link it and to its own pages, and
// gives it to AIs with get_house_style. A Cahut keeps its current house
// style and the one before, which rollback_house_style restores. See
// docs/adr/0022-house-style-of-each-cahut.md.
package housestyle

import (
	"context"
	"fmt"
	"math"
	"slices"
	"time"

	"cahut.cloud/internal/operation"
)

// The addresses of the house style, from the root of the domain of its
// Cahut.
const (
	SheetPath       = "/-/house-style.css"
	LogoPath        = "/-/house-style/logo"
	LogoInversePath = "/-/house-style/logo-inverse"
	FontsPath       = "/-/house-style/fonts/"
)

// HouseStyle is the house style of a Cahut, as get_house_style gives it,
// with the colors and the fonts that Cahut derives.
type HouseStyle struct {
	Name           string     `json:"name" jsonschema:"the name of the organization; the domain of the Cahut until an administrator sets one"`
	Neutral        bool       `json:"neutral" jsonschema:"whether the Cahut has the neutral house style, which no administrator has changed"`
	Stylesheet     string     `json:"stylesheet" jsonschema:"the style sheet of the house style, at this path from the root of the domain: link it with <link rel=\"stylesheet\" href=\"/-/house-style.css\">. The Window declares nothing for it"`
	Logo           string     `json:"logo" jsonschema:"the path of the logo, from the root of the domain; without a logo, it shows the name of the organization"`
	LogoInverse    string     `json:"logo_inverse" jsonschema:"the path of the logo for dark backgrounds, such as the primary color; without such a version, it shows the name of the organization in light letters"`
	HasLogo        bool       `json:"has_logo" jsonschema:"whether the house style has a logo"`
	HasLogoInverse bool       `json:"has_logo_inverse" jsonschema:"whether it has a version of its logo for dark backgrounds"`
	DeckEngine     string     `json:"deck_engine" jsonschema:"the path of the deck engine, from the root of the domain: <script src=\"/-/deck.js\" defer></script> in the head of a deck, which then declares that it runs scripts"`
	Colors         Palette    `json:"colors"`
	ChartColors    []string   `json:"chart_colors" jsonschema:"the colors of the series of charts, in their order, as --hs-chart-1 to --hs-chart-8 in the style sheet: give them to series in this order, never cycled"`
	Fonts          Fonts      `json:"fonts"`
	Guidance       string     `json:"guidance,omitempty" jsonschema:"what the organization asks of AIs, such as its tone of voice or rules for charts: follow it"`
	SetAt          *time.Time `json:"set_at,omitempty" jsonschema:"when the house style was set; absent for the neutral house style"`
	SetBy          string     `json:"set_by,omitempty" jsonschema:"the handle of the person who set it, or for whom an AI set it"`
	Previous       *Previous  `json:"previous,omitempty" jsonschema:"the previous house style, which rollback_house_style restores; absent when there is none"`
	Warnings       []string   `json:"warnings,omitempty" jsonschema:"what may not read well, and what to change"`
	Guide          string     `json:"guide,omitempty" jsonschema:"how to use the style sheet, its variables and its classes, in Markdown"`
	Template       *Template  `json:"template,omitempty" jsonschema:"the template of the kind asked for"`
}

// Palette is the colors of a house style, as #rrggbb, and the CSS variables
// that hold them.
type Palette struct {
	Text        string `json:"text" jsonschema:"body text, --hs-text"`
	Background  string `json:"background" jsonschema:"the background of pages, --hs-background"`
	Primary     string `json:"primary" jsonschema:"the main color of the brand, for headings, links, buttons and slides, --hs-primary"`
	OnPrimary   string `json:"on_primary" jsonschema:"text on the primary color, --hs-on-primary"`
	Secondary   string `json:"secondary" jsonschema:"the second color of the brand, for accents, --hs-secondary"`
	OnSecondary string `json:"on_secondary" jsonschema:"text on the secondary color, --hs-on-secondary"`
	Muted       string `json:"muted" jsonschema:"secondary text, such as captions, --hs-muted"`
	Line        string `json:"line" jsonschema:"borders and rules, --hs-line"`
	Surface     string `json:"surface" jsonschema:"the background of cards and panels, --hs-surface"`
}

// Fonts is the fonts of a house style.
type Fonts struct {
	Heading Font   `json:"heading" jsonschema:"the font of headings, --hs-font-heading"`
	Body    Font   `json:"body" jsonschema:"the font of body text, --hs-font-body"`
	Mono    string `json:"mono" jsonschema:"the CSS stack of code and figures in columns, --hs-font-mono"`
}

// Font is a font of a house style.
type Font struct {
	Family   string     `json:"family,omitempty" jsonschema:"its family; absent for the font of the system"`
	Fallback string     `json:"fallback" jsonschema:"the generic family it falls back on: sans-serif, serif or monospace"`
	Stack    string     `json:"stack" jsonschema:"the CSS font-family of the font, its family first"`
	Files    []FontFile `json:"files,omitempty" jsonschema:"the files that readers' browsers load; none for a family that their systems have"`
}

// FontFile is a file of a font.
type FontFile struct {
	Name      string `json:"name" jsonschema:"the name of the file, which set_house_style keeps when given in keep"`
	URL       string `json:"url" jsonschema:"its path, from the root of the domain"`
	Weight    int    `json:"weight"`
	MaxWeight int    `json:"max_weight,omitempty" jsonschema:"for a variable font, the heaviest weight it covers"`
	Style     string `json:"style" jsonschema:"normal or italic"`
	Size      int64  `json:"size" jsonschema:"in bytes"`
}

// Previous is the house style that rollback_house_style restores.
type Previous struct {
	Neutral bool      `json:"neutral" jsonschema:"whether it is the neutral house style"`
	SetAt   time.Time `json:"set_at"`
	SetBy   string    `json:"set_by,omitempty"`
}

// GetInput is the input of get_house_style.
type GetInput struct {
	Kind string `json:"kind,omitempty" jsonschema:"the kind of Window you make, to get its template: document, deck, dashboard or site"`
}

// Register adds the operations of the house style of the Cahut at domain
// to r.
func Register(r *operation.Registry, domain string) {
	operation.Register(r, &operation.Op[GetInput, HouseStyle]{
		Name: "get_house_style",
		Description: "Get the house style of this Cahut before making a Window: the name of the organization, its colors, fonts and logo, " +
			"what it asks of AIs, the path of its style sheet, which a Window links from the root of the domain, a guide to its classes, " +
			"and the template of the kind of Window you make.",
		ReadOnly: true,
		Run: func(ctx context.Context, c *operation.Call, in GetInput) (HouseStyle, error) {
			c.Log("house style", in.Kind)
			if _, ok := kinds[in.Kind]; in.Kind != "" && !ok {
				return HouseStyle{}, operation.Invalidf("kind %q: use document, deck, dashboard or site", in.Kind)
			}
			cur, prev, err := load(ctx, c.Tx, false)
			if err != nil {
				return HouseStyle{}, err
			}
			h := describe(cur, prev, domain)
			h.Guide = guide
			h.Template, err = render(in.Kind, h.Name, c.Now)
			return h, err
		},
	})
	operation.Register(r, &operation.Op[SetInput, HouseStyle]{
		Name: "set_house_style",
		Description: "Change the house style of this Cahut: the name of the organization, its colors, its fonts and their files, its logo, " +
			"and what it asks of AIs. Give only what changes: the rest stays. Every Window that links the style sheet takes the change at once. " +
			"The house style before becomes the previous one, which rollback_house_style restores. For the administrators of the Cahut.",
		Level: operation.CahutAdmin,
		Run: func(ctx context.Context, c *operation.Call, in SetInput) (HouseStyle, error) {
			return set(ctx, c, in, domain)
		},
	})
	operation.Register(r, &operation.Op[RollbackInput, HouseStyle]{
		Name: "rollback_house_style",
		Description: "Go back to the previous house style of this Cahut. The house style you leave becomes the previous one, " +
			"so calling rollback_house_style again returns to it. For the administrators of the Cahut.",
		Level: operation.CahutAdmin,
		Run: func(ctx context.Context, c *operation.Call, _ RollbackInput) (HouseStyle, error) {
			return rollback(ctx, c, domain)
		},
	})
}

// RollbackInput is the input of rollback_house_style.
type RollbackInput struct{}

// The neutral house style, that of a Cahut until an administrator sets one.
// It is not the brand of Cahut.
var neutral = struct {
	text, background, primary, secondary string
	chart                                []string
}{
	text: "#1d1f23", background: "#ffffff", primary: "#2a5bd7", secondary: "#0f766e",
	// A categorical palette checked for readers with color vision
	// deficiencies, in its order.
	chart: []string{"#2a78d6", "#eb6834", "#1baf7a", "#eda100", "#e87ba4", "#008300", "#6250d6", "#e34948"},
}

// Generic families and their stacks.
var stacks = map[string]string{
	"sans-serif": `system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif`,
	"serif":      `Georgia, "Times New Roman", Times, serif`,
	"monospace":  `ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace`,
}

// MaxChartColors is how many series colors a house style has.
const MaxChartColors = 8

// describe returns the house style cur, or the neutral house style when cur
// is nil, with prev before it, as get_house_style gives it.
func describe(cur, prev *version, domain string) HouseStyle {
	var s settings
	files := map[string]file{}
	if cur != nil {
		s, files = cur.settings, cur.files
	}
	h := HouseStyle{
		Name:       s.Name,
		Neutral:    cur == nil || cur.neutral(),
		Stylesheet: SheetPath, Logo: LogoPath, LogoInverse: LogoInversePath, DeckEngine: DeckPath,
		Guidance: s.Guidance,
	}
	if h.Name == "" {
		h.Name = domain
	}
	_, h.HasLogo = files[logoFile]
	_, h.HasLogoInverse = files[logoInverseFile]
	h.Colors, h.ChartColors = palette(s)
	h.Fonts = Fonts{Body: fontOf(s.Body, files), Mono: stacks["monospace"]}
	h.Fonts.Heading = h.Fonts.Body
	if s.Heading != nil {
		h.Fonts.Heading = fontOf(s.Heading, files)
	}
	h.Warnings = warnings(h.Colors)
	if h.HasLogo && !h.HasLogoInverse {
		h.Warnings = append(h.Warnings, "logo_inverse is missing: on dark backgrounds, such as the slides that open a part, "+
			"the name of the organization shows in place of the logo; give a version of the logo for dark backgrounds")
	}
	if cur != nil && !h.Neutral {
		at := cur.setAt
		h.SetAt, h.SetBy = &at, cur.setBy
	}
	if prev != nil {
		h.Previous = &Previous{Neutral: prev.neutral(), SetAt: prev.setAt, SetBy: prev.setBy}
	}
	return h
}

// palette returns the colors of s, with those that Cahut derives, and the
// colors of the series of charts.
func palette(s settings) (Palette, []string) {
	pick := func(set, def string) rgb {
		if set != "" {
			return mustColor(set)
		}
		return mustColor(def)
	}
	text, background := pick(s.Colors.Text, neutral.text), pick(s.Colors.Background, neutral.background)
	primary, secondary := pick(s.Colors.Primary, neutral.primary), pick(s.Colors.Secondary, neutral.secondary)
	p := Palette{
		Text: text.String(), Background: background.String(),
		Primary: primary.String(), OnPrimary: on(primary, text, background).String(),
		Secondary: secondary.String(), OnSecondary: on(secondary, text, background).String(),
		Muted:   muted(text, background).String(),
		Line:    mix(text, background, 0.16).String(),
		Surface: mix(text, background, 0.04).String(),
	}
	return p, chart(s, primary, secondary)
}

// on returns the color of text on the color c: the text color or the
// background color, whichever contrasts more with it.
func on(c, text, background rgb) rgb {
	if contrast(c, text) >= contrast(c, background) {
		return text
	}
	return background
}

// muted returns the lightest mix of text and background that still reads
// as body text on the background.
func muted(text, background rgb) rgb {
	for t := 0.55; t < 1; t += 0.05 {
		if m := mix(text, background, t); contrast(m, background) >= MinContrast {
			return m
		}
	}
	return text
}

// chart returns the colors of the series of charts: those that the house
// style sets, or its primary and secondary colors, completed to eight with
// the neutral palette, leaving out the colors too close to those already
// there.
func chart(s settings, primary, secondary rgb) []string {
	var chosen []rgb
	switch {
	case len(s.ChartColors) > 0:
		for _, c := range s.ChartColors {
			chosen = append(chosen, mustColor(c))
		}
	case s.Colors.Primary != "" || s.Colors.Secondary != "":
		chosen = []rgb{primary}
		if distance(primary, secondary) >= distinct {
			chosen = append(chosen, secondary)
		}
	default:
		return slices.Clone(neutral.chart)
	}
	// nearest returns the distance of c to the closest color chosen.
	nearest := func(c rgb) float64 {
		d := math.Inf(1)
		for _, o := range chosen {
			d = min(d, distance(c, o))
		}
		return d
	}
	var left []rgb
	for _, n := range neutral.chart {
		if c := mustColor(n); nearest(c) >= distinct && len(chosen) < MaxChartColors {
			chosen = append(chosen, c)
		} else {
			left = append(left, c)
		}
	}
	// When too few colors are far enough apart, the farthest of those left
	// complete the series.
	for len(chosen) < MaxChartColors && len(left) > 0 {
		far := 0
		for i, c := range left {
			if nearest(c) > nearest(left[far]) {
				far = i
			}
		}
		chosen = append(chosen, left[far])
		left = slices.Delete(left, far, far+1)
	}
	out := make([]string, len(chosen))
	for i, c := range chosen {
		out[i] = c.String()
	}
	return out
}

// distinct is the distance below which two colors of series are hard to
// tell apart.
const distinct = 15

// Contrasts that colors should keep, as WCAG AA asks: MinContrast for text,
// and MinUIContrast for large text and the parts of an interface.
const (
	MinContrast   = 4.5
	MinUIContrast = 3
)

// warnings says which colors of p may not read well.
func warnings(p Palette) []string {
	var w []string
	background := mustColor(p.Background)
	for _, c := range []struct{ name, color, on string }{
		{"primary", p.Primary, p.OnPrimary},
		{"secondary", p.Secondary, p.OnSecondary},
	} {
		color := mustColor(c.color)
		if r := contrast(color, background); r < MinUIContrast {
			w = append(w, fmt.Sprintf("colors.%s %s contrasts %.1f:1 on the background: links, buttons and headings in it are hard to see; "+
				"%.0f:1 at least reads well", c.name, c.color, r, float64(MinUIContrast)))
		}
		if r := contrast(color, mustColor(c.on)); r < MinContrast {
			w = append(w, fmt.Sprintf("text on colors.%s %s contrasts %.1f:1 at best: titles on slides in that color are hard to read; "+
				"a darker or a lighter %s color reads better", c.name, c.color, r, c.name))
		}
	}
	return w
}

// fontOf returns the font f with the files that it names, or the font of
// the system when f is nil.
func fontOf(f *font, files map[string]file) Font {
	if f == nil {
		f = &font{}
	}
	out := Font{Family: f.Family, Fallback: f.Fallback}
	if out.Fallback == "" {
		out.Fallback = "sans-serif"
	}
	out.Stack = stacks[out.Fallback]
	if f.Family != "" {
		out.Stack = `"` + f.Family + `", ` + out.Stack
	}
	for _, ff := range f.Files {
		out.Files = append(out.Files, FontFile{Name: ff.Name, URL: FontsPath + ff.Name, Weight: ff.Weight, MaxWeight: ff.MaxWeight,
			Style: ff.Style, Size: files[fontsDir+ff.Name].size})
	}
	return out
}

// weights returns the weight of a font file as CSS writes it: one weight,
// or the range of a variable font.
func (f FontFile) weights() string {
	if f.MaxWeight > 0 {
		return fmt.Sprintf("%d %d", f.Weight, f.MaxWeight)
	}
	return fmt.Sprint(f.Weight)
}
