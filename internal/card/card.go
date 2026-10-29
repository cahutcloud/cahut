// Package card draws the card of a Window: the picture that the previews of
// a link to it show, in the brand of Cahut, with the title of the Window and
// the domain of its Cahut. The server draws it in Go, without a browser,
// from the fonts and the pictures of the brand guidelines. See
// docs/adr/0051-sharing-a-window-and-its-preview.md.
package card

import (
	"bytes"
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"math"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"golang.org/x/text/unicode/norm"

	charte "cahut.cloud/docs/charte-graphique"
)

// Width and Height are the size of a card, in pixels: the size that the
// platforms ask of the picture of a preview.
const (
	Width  = 1200
	Height = 630
)

// The layout follows the picture of sharing of the brand guidelines,
// docs/charte-graphique/applications/gabarits/og-image.html: the logo at the
// top left, the title where its signature stands, the domain at the bottom
// left and the open house on the right.
const (
	margin     = 72
	logoTop    = 64
	logoHeight = 46
	// The title is no wider than its column, left of the open house.
	column = 600
	// The baseline of the domain, in IBM Plex Mono.
	domainBaseline = Height - 62
	domainSize     = 22
	// The open house, faded at its edges.
	artWidth, artHeight = 600, 400
	artLeft, artTop     = Width + 70 - artWidth, Height - artHeight - 40
)

// sizes are the sizes of the title, in pixels, from the largest, and the
// lines it may take at each: a long title shrinks, then stops with an
// ellipsis.
var sizes = []struct {
	px    float64
	lines int
}{{64, 3}, {56, 4}, {48, 5}, {40, 5}}

// version changes with the drawing, so that the tags of the cards change
// with it.
const version = "1"

// Tag returns a short hash of what the card of title in the Cahut at domain
// shows. Its address carries it, so that the platforms that keep pictures
// by address take the new card after a new title.
func Tag(title, domain string) string {
	sum := sha256.Sum256([]byte(version + "\x00" + domain + "\x00" + title))
	return hex.EncodeToString(sum[:6])
}

// Draw returns the card of the Window called title, in the Cahut at domain,
// as a PNG. An empty title, or one that the font of the brand cannot write,
// gives the card without a title.
func Draw(title, domain string) ([]byte, error) {
	a, err := load()
	if err != nil {
		return nil, err
	}
	img := image.NewRGBA(image.Rect(0, 0, Width, Height))
	draw.Draw(img, img.Bounds(), image.NewUniform(a.paper), image.Point{}, draw.Src)
	draw.Draw(img, a.art.Bounds().Add(image.Pt(artLeft, artTop)), a.art, image.Point{}, draw.Over)
	draw.Draw(img, a.logo.Bounds().Add(image.Pt(margin, logoTop)), a.logo, image.Point{}, draw.Over)

	if t, ok := writable(a.title, title); ok && t != "" {
		var w *writer
		var lines []string
		for i, s := range sizes {
			w = newWriter(a.title, s.px, -0.02)
			var fits bool
			if lines, fits = w.wrap(t, column, s.lines); fits {
				break
			}
			if i == len(sizes)-1 {
				lines[len(lines)-1] = w.ellipsize(lines[len(lines)-1], column)
			}
		}
		// The title sits in the middle of the space between the logo and
		// the domain, on lines 1.12 times as high as its letters.
		lineHeight := int(math.Round(w.px * 1.12))
		top, bottom := logoTop+logoHeight+48, domainBaseline-domainSize-48
		ascent := int(math.Round(w.px * 0.92))
		y := top + max(0, (bottom-top-lineHeight*len(lines))/2) + ascent
		for _, line := range lines {
			w.draw(img, line, margin, y, a.ink)
			y += lineHeight
		}
	}

	m := newWriter(a.mono, domainSize, 0)
	d := domain
	if m.width(d) > fixed.I(column) {
		d = m.ellipsize(d, column)
	}
	m.draw(img, d, margin, domainBaseline, a.sky)

	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// assets are what every card is drawn with, read and scaled once.
type assets struct {
	title, mono     *sfnt.Font
	logo, art       image.Image
	paper, ink, sky color.Color
}

var load = sync.OnceValues(func() (*assets, error) {
	var a assets
	var err error
	if a.title, err = parseFont("typographie/polices/ttf/PlusJakartaSans-ExtraBold.ttf"); err != nil {
		return nil, err
	}
	if a.mono, err = parseFont("typographie/polices/ttf/IBMPlexMono-Regular.ttf"); err != nil {
		return nil, err
	}
	logo, err := decodePNG("logos/png/cahut-logo-horizontal-1600px.png")
	if err != nil {
		return nil, err
	}
	a.logo = scale(logo, logo.Bounds().Dx()*logoHeight/logo.Bounds().Dy(), logoHeight)
	art, err := decodePNG("illustrations/maison-ouverte.png")
	if err != nil {
		return nil, err
	}
	a.art = fade(scale(art, artWidth, artHeight))
	tokens, err := fs.ReadFile(charte.Files, "tokens/cahut-tokens.css")
	if err != nil {
		return nil, err
	}
	if a.paper, err = token(tokens, "cahut-paper"); err != nil {
		return nil, err
	}
	if a.ink, err = token(tokens, "cahut-ink"); err != nil {
		return nil, err
	}
	if a.sky, err = token(tokens, "cahut-sky-700"); err != nil {
		return nil, err
	}
	return &a, nil
})

func parseFont(p string) (*sfnt.Font, error) {
	b, err := fs.ReadFile(charte.CardFiles, p)
	if err != nil {
		return nil, err
	}
	return opentype.Parse(b)
}

func decodePNG(p string) (image.Image, error) {
	b, err := fs.ReadFile(charte.CardFiles, p)
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(b))
}

// token returns the color of the brand called name in the tokens of the
// brand guidelines, which hold every color of Cahut.
func token(css []byte, name string) (color.Color, error) {
	m := regexp.MustCompile(`--` + regexp.QuoteMeta(name) + `:\s*#([0-9A-Fa-f]{6});`).FindSubmatch(css)
	if m == nil {
		return nil, fmt.Errorf("no color --%s in the tokens of the brand", name)
	}
	v, err := hex.DecodeString(string(m[1]))
	if err != nil {
		return nil, err
	}
	return color.RGBA{v[0], v[1], v[2], 0xFF}, nil
}

func scale(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
	return dst
}

// fade fades the edges of the open house into the paper, as the mask of
// the picture of sharing of the brand guidelines does: opaque within 62% of
// an ellipse of 72% by 70% of the picture, centered at 55% and 46%, and
// clear at its edge.
func fade(img *image.RGBA) image.Image {
	b := img.Bounds()
	mask := image.NewAlpha(b)
	cx, cy := 0.55*float64(b.Dx()), 0.46*float64(b.Dy())
	rx, ry := 0.72*float64(b.Dx()), 0.70*float64(b.Dy())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			d := math.Hypot((float64(x)-cx)/rx, (float64(y)-cy)/ry)
			mask.SetAlpha(x, y, color.Alpha{A: uint8(255 * math.Max(0, math.Min(1, (1-d)/0.38)))})
		}
	}
	out := image.NewRGBA(b)
	draw.DrawMask(out, b, img, b.Min, mask, b.Min, draw.Src)
	return out
}

// writable returns title as a card writes it: in the compatible form of
// Unicode, so that 2ᵉ reads 2e, on one line, without the symbols that the
// font lacks, such as emoji. It returns false when the font lacks a letter
// or a digit of it: the card then shows no title, and the preview gives it
// in text, which the platforms write in their own fonts.
func writable(f *sfnt.Font, title string) (string, bool) {
	var buf sfnt.Buffer
	var b strings.Builder
	for _, r := range norm.NFKC.String(title) {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			b.WriteByte(' ')
			continue
		}
		if i, err := f.GlyphIndex(&buf, r); err == nil && i != 0 {
			b.WriteRune(r)
			continue
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r) {
			return "", false
		}
	}
	return strings.Join(strings.Fields(b.String()), " "), true
}

// A writer writes text in a font, at a size, with a tracking in em.
type writer struct {
	px       float64
	face     font.Face
	tracking fixed.Int26_6
}

func newWriter(f *sfnt.Font, px, tracking float64) *writer {
	// NewFace fails only on options out of range.
	face, _ := opentype.NewFace(f, &opentype.FaceOptions{Size: px, DPI: 72, Hinting: font.HintingNone})
	return &writer{px: px, face: face, tracking: fixed.Int26_6(math.Round(tracking * px * 64))}
}

func (w *writer) width(s string) fixed.Int26_6 {
	var x fixed.Int26_6
	prev := rune(-1)
	for _, r := range s {
		if prev >= 0 {
			x += w.face.Kern(prev, r)
		}
		a, _ := w.face.GlyphAdvance(r)
		x += a + w.tracking
		prev = r
	}
	return x
}

func (w *writer) draw(dst draw.Image, s string, x, y int, c color.Color) {
	d := font.Drawer{Dst: dst, Src: image.NewUniform(c), Face: w.face, Dot: fixed.P(x, y)}
	prev := rune(-1)
	for _, r := range s {
		if prev >= 0 {
			d.Dot.X += w.face.Kern(prev, r)
		}
		d.DrawString(string(r))
		d.Dot.X += w.tracking
		prev = r
	}
}

// wrap breaks s into lines no wider than width, between words, and inside
// a word too wide for a line. It returns at most n lines, and whether s
// fits in them.
func (w *writer) wrap(s string, width, n int) ([]string, bool) {
	limit := fixed.I(width)
	var lines []string
	line := ""
	words := strings.Fields(s)
	for len(words) > 0 {
		word := words[0]
		next := word
		if line != "" {
			next = line + " " + word
		}
		switch {
		case w.width(next) <= limit:
			line, words = next, words[1:]
			continue
		case line != "":
			lines, line = append(lines, line), ""
		default:
			// A word wider than the line: as much of it as fits, a letter
			// at least.
			cut := 0
			for i, r := range word {
				end := i + utf8.RuneLen(r)
				if cut > 0 && w.width(word[:end]) > limit {
					break
				}
				cut = end
			}
			lines, words[0] = append(lines, word[:cut]), word[cut:]
		}
		if len(lines) == n {
			return lines, false
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines, len(lines) <= n
}

// ellipsize ends line with an ellipsis, shortened so that it fits width.
func (w *writer) ellipsize(line string, width int) string {
	limit := fixed.I(width)
	rs := []rune(line)
	for len(rs) > 0 && w.width(string(rs)+"…") > limit {
		rs = rs[:len(rs)-1]
	}
	return strings.TrimRight(string(rs), " ,;:") + "…"
}

// Cache keeps the cards drawn last, by what they show, so that the robots
// that come again cost no drawing.
type Cache struct {
	mu    sync.Mutex
	max   int
	order *list.List // the most recent first
	cards map[string]*list.Element
}

type cached struct {
	tag string
	png []byte
}

// NewCache returns a cache of the last n cards.
func NewCache(n int) *Cache {
	return &Cache{max: n, order: list.New(), cards: map[string]*list.Element{}}
}

// Get returns the card of title in the Cahut at domain, and its tag,
// drawing it if the cache does not hold it.
func (c *Cache) Get(title, domain string) ([]byte, string, error) {
	tag := Tag(title, domain)
	c.mu.Lock()
	if e, ok := c.cards[tag]; ok {
		c.order.MoveToFront(e)
		c.mu.Unlock()
		return e.Value.(*cached).png, tag, nil
	}
	c.mu.Unlock()
	b, err := Draw(title, domain)
	if err != nil {
		return nil, "", err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.cards[tag]; !ok {
		c.cards[tag] = c.order.PushFront(&cached{tag: tag, png: b})
		for c.order.Len() > c.max {
			last := c.order.Back()
			c.order.Remove(last)
			delete(c.cards, last.Value.(*cached).tag)
		}
	}
	return b, tag, nil
}
