package card

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"golang.org/x/image/math/fixed"
)

func decode(t *testing.T, b []byte) image.Image {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds() != image.Rect(0, 0, Width, Height) {
		t.Fatalf("a card of %v, want %dx%d", img.Bounds(), Width, Height)
	}
	return img
}

// inked counts the pixels of the column of the title that are darker than
// the paper: the letters of the title.
func inked(img image.Image) int {
	n := 0
	for y := logoTop + logoHeight + 24; y < domainBaseline-domainSize-24; y++ {
		for x := margin; x < margin+column; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if (r+g+b)/3 < 0x8000 {
				n++
			}
		}
	}
	return n
}

func TestDraw(t *testing.T) {
	a, err := load()
	if err != nil {
		t.Fatal(err)
	}
	if a.paper != (color.RGBA{0xF7, 0xF5, 0xF0, 0xFF}) {
		t.Errorf("paper = %v, want the token of the brand", a.paper)
	}
	cards := map[string]image.Image{}
	for name, title := range map[string]string{
		"short":    "Q3 review",
		"long":     strings.Repeat("Le plan de lancement du produit pour le marché européen, ", 4),
		"word":     strings.Repeat("Donaudampfschifffahrtsgesellschaft", 6),
		"emoji":    "📊 Ventes du 2ᵉ trimestre",
		"japanese": "第3四半期の売上レビュー",
		"none":     "",
	} {
		b, err := Draw(title, "cahut.example.com")
		if err != nil {
			t.Fatalf("Draw(%q) = %v", title, err)
		}
		cards[name] = decode(t, b)
	}
	for _, name := range []string{"short", "long", "word", "emoji"} {
		if inked(cards[name]) == 0 {
			t.Errorf("the card %s has no title", name)
		}
	}
	// A title that the font cannot write leaves the column empty.
	for _, name := range []string{"japanese", "none"} {
		if n := inked(cards[name]); n != 0 {
			t.Errorf("the card %s has %d inked pixels in the column of the title, want none", name, n)
		}
	}
	if inked(cards["long"]) <= inked(cards["short"]) {
		t.Error("a long title writes no more than a short one")
	}
}

func TestWritable(t *testing.T) {
	a, err := load()
	if err != nil {
		t.Fatal(err)
	}
	for in, want := range map[string]string{
		"Q3 review":                  "Q3 review",
		"  Q3\treview\n":             "Q3 review",
		"📊 Ventes du 2ᵉ trimestre":   "Ventes du 2e trimestre",
		"Báo cáo doanh thu quý ba":   "Báo cáo doanh thu quý ba",
		"Œuvres d'été — « vues »":    "Œuvres d'été — « vues »",
		"ﬁnal":                       "final",
		"第3四半期":                      "",
		"Годовой отчёт":              "",
		"Report: Q3 for Ελλάδα team": "",
	} {
		got, ok := writable(a.title, in)
		if ok != (want != "") || got != want {
			t.Errorf("writable(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
}

func TestWrap(t *testing.T) {
	a, err := load()
	if err != nil {
		t.Fatal(err)
	}
	w := newWriter(a.title, 64, -0.02)
	lines, fits := w.wrap("Q3 review", column, 3)
	if !fits || len(lines) != 1 || lines[0] != "Q3 review" {
		t.Errorf("wrap(Q3 review) = %q, %v", lines, fits)
	}
	long := strings.Repeat("Donaudampfschifffahrtsgesellschaft", 3)
	lines, fits = w.wrap("A "+long+" plan", column, 9)
	if !fits || strings.Join(lines, "") != "A"+long+" plan" {
		t.Errorf("wrap of a long word = %q, %v", lines, fits)
	}
	for _, l := range lines {
		if w.width(l) > fixed.I(column) {
			t.Errorf("the line %q is wider than the column", l)
		}
	}
	lines, fits = w.wrap(strings.Repeat("word ", 60), column, 3)
	if fits || len(lines) != 3 {
		t.Errorf("wrap of a long title in 3 lines = %d lines, %v", len(lines), fits)
	}
	if e := w.ellipsize(lines[2], column); !strings.HasSuffix(e, "…") || w.width(e) > fixed.I(column) {
		t.Errorf("ellipsize = %q", e)
	}
}

func TestTag(t *testing.T) {
	a, b := Tag("Q3 review", "cahut.example.com"), Tag("Q3 review", "cahut.example.com")
	if a != b || len(a) != 12 {
		t.Errorf("Tag = %q, %q", a, b)
	}
	if a == Tag("Q4 review", "cahut.example.com") || a == Tag("Q3 review", "beta.example.com") || a == Tag("", "cahut.example.com") {
		t.Error("two cards share a tag")
	}
}

func TestCache(t *testing.T) {
	c := NewCache(2)
	first, tag, err := c.Get("One", "localhost")
	if err != nil || tag != Tag("One", "localhost") {
		t.Fatalf("Get = %q, %v", tag, err)
	}
	again, _, _ := c.Get("One", "localhost")
	if &again[0] != &first[0] {
		t.Error("the cache drew a card it holds")
	}
	for _, title := range []string{"Two", "Three"} {
		if _, _, err := c.Get(title, "localhost"); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok := c.cards[tag]; ok || c.order.Len() != 2 {
		t.Errorf("the cache holds %d cards, the first one included: %v", c.order.Len(), ok)
	}
}
