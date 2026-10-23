package housestyle

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// rgb is a color of a house style, in sRGB.
type rgb struct{ r, g, b uint8 }

// parseColor reads a color as #rrggbb or #rgb, in any case.
func parseColor(s string) (rgb, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if !strings.HasPrefix(s, "#") {
		return rgb{}, false
	}
	s = s[1:]
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return rgb{}, false
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return rgb{}, false
	}
	return rgb{uint8(v >> 16), uint8(v >> 8 & 0xff), uint8(v & 0xff)}, true //nolint:gosec // six hex digits fit in 24 bits
}

func mustColor(s string) rgb {
	c, ok := parseColor(s)
	if !ok {
		panic("housestyle: invalid color " + s) // a constant: a programming error
	}
	return c
}

// String returns the color as #rrggbb.
func (c rgb) String() string {
	return fmt.Sprintf("#%02x%02x%02x", c.r, c.g, c.b)
}

// linear returns a channel of the color in linear light, from 0 to 1.
func linear(v uint8) float64 {
	c := float64(v) / 255
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// luminance returns the relative luminance of the color, as WCAG defines it.
func (c rgb) luminance() float64 {
	return 0.2126*linear(c.r) + 0.7152*linear(c.g) + 0.0722*linear(c.b)
}

// contrast returns the contrast ratio of two colors, from 1 to 21, as WCAG
// defines it.
func contrast(a, b rgb) float64 {
	la, lb := a.luminance(), b.luminance()
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// mix returns t of a and 1-t of b.
func mix(a, b rgb, t float64) rgb {
	ch := func(x, y uint8) uint8 {
		return uint8(math.Round(t*float64(x) + (1-t)*float64(y)))
	}
	return rgb{ch(a.r, b.r), ch(a.g, b.g), ch(a.b, b.b)}
}

// distance returns how far apart two colors look, as the distance of their
// coordinates in OKLab, times 100: below 15, people tell them apart with
// difficulty.
func distance(a, b rgb) float64 {
	la, aa, ba := a.oklab()
	lb, ab, bb := b.oklab()
	return 100 * math.Sqrt((la-lb)*(la-lb)+(aa-ab)*(aa-ab)+(ba-bb)*(ba-bb))
}

// oklab returns the coordinates of the color in OKLab.
func (c rgb) oklab() (l, a, b float64) {
	r, g, bl := linear(c.r), linear(c.g), linear(c.b)
	lc := math.Cbrt(0.4122214708*r + 0.5363325363*g + 0.0514459929*bl)
	mc := math.Cbrt(0.2119034982*r + 0.6806995451*g + 0.1073969566*bl)
	sc := math.Cbrt(0.0883024619*r + 0.2817188376*g + 0.6299787005*bl)
	return 0.2104542553*lc + 0.7936177850*mc - 0.0040720468*sc,
		1.9779984951*lc - 2.4285922050*mc + 0.4505937099*sc,
		0.0259040371*lc + 0.7827717662*mc - 0.8086757660*sc
}

// dark reports whether text on the color should be light.
func (c rgb) dark() bool {
	return c.luminance() < 0.18
}
