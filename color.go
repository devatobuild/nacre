package nacre

import (
	"fmt"
	"math"
	"strings"
)

// Color is an sRGB colour with straight (non-premultiplied) alpha.
// All channels are in [0, 1]. The zero Color is fully transparent and
// is what nacre uses to mean "no fill" or "no stroke".
type Color struct{ R, G, B, A float64 }

// None is the transparent colour: no fill, no stroke.
var None = Color{}

// Hex parses "#rgb", "#rrggbb" or "#rrggbbaa" and panics on malformed
// input. It exists for colour literals in palettes and styles; use
// ParseHex for untrusted input.
func Hex(s string) Color {
	c, err := ParseHex(s)
	if err != nil {
		panic(err)
	}
	return c
}

// ParseHex parses "#rgb", "#rrggbb" or "#rrggbbaa" (the '#' is optional).
func ParseHex(s string) (Color, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	var r, g, b, a uint8 = 0, 0, 0, 255
	var err error
	switch len(s) {
	case 3:
		_, err = fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b)
		r, g, b = r*17, g*17, b*17
	case 6:
		_, err = fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
	case 8:
		_, err = fmt.Sscanf(s, "%02x%02x%02x%02x", &r, &g, &b, &a)
	default:
		err = fmt.Errorf("nacre: bad colour %q", s)
	}
	if err != nil {
		return None, fmt.Errorf("nacre: bad colour %q", s)
	}
	return Color{float64(r) / 255, float64(g) / 255, float64(b) / 255, float64(a) / 255}, nil
}

// RGB builds an opaque colour from 8-bit channels.
func RGB(r, g, b uint8) Color { return Color{float64(r) / 255, float64(g) / 255, float64(b) / 255, 1} }

// IsNone reports whether the colour is fully transparent.
func (c Color) IsNone() bool { return c.A <= 0 }

// Hex formats the colour as "#rrggbb", dropping alpha.
func (c Color) Hex() string {
	r, g, b, _ := c.RGBA8()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// RGBA8 returns the colour quantised to 8 bits per channel.
func (c Color) RGBA8() (r, g, b, a uint8) {
	q := func(v float64) uint8 { return uint8(math.Round(clamp01(v) * 255)) }
	return q(c.R), q(c.G), q(c.B), q(c.A)
}

// Alpha returns the colour with its alpha replaced by a.
func (c Color) Alpha(a float64) Color { c.A = clamp01(a); return c }

// Mix linearly interpolates from c to o by t in [0, 1].
func (c Color) Mix(o Color, t float64) Color {
	t = clamp01(t)
	return Color{
		c.R + (o.R-c.R)*t,
		c.G + (o.G-c.G)*t,
		c.B + (o.B-c.B)*t,
		c.A + (o.A-c.A)*t,
	}
}

// Lighten mixes the colour toward white by t.
func (c Color) Lighten(t float64) Color { return c.Mix(Color{1, 1, 1, c.A}, t) }

// Darken mixes the colour toward black by t.
func (c Color) Darken(t float64) Color { return c.Mix(Color{0, 0, 0, c.A}, t) }

// Luminance returns the relative luminance in [0, 1] (WCAG definition).
func (c Color) Luminance() float64 {
	lin := func(v float64) float64 {
		if v <= 0.03928 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
}

// IsDark reports whether the colour reads as dark against white text.
func (c Color) IsDark() bool { return c.Luminance() < 0.18 }

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
