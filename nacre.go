// Package nacre is a zero-dependency generative art engine.
//
// A piece is fully determined by a 64-bit seed: the seed picks a style
// and a palette (unless you choose them), and the style draws onto a
// resolution-independent Canvas which can then be written as SVG or
// rasterized to PNG. The same seed always reproduces the same piece.
//
//	import (
//		"os"
//
//		"github.com/devatobuild/nacre"
//		_ "github.com/devatobuild/nacre/styles" // registers the built-in styles
//	)
//
//	piece, err := nacre.Render(nacre.Options{Style: "flow", Seed: 0x3f1a})
//	if err != nil {
//		panic(err)
//	}
//	os.WriteFile("flow.svg", piece.SVG(nacre.SVGOptions{Simplify: 0.4}), 0o644)
//
// Styles are plain functions. Register your own with Register and it
// becomes available everywhere, including the CLI and the playground.
package nacre

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"time"
)

// Options describe the piece to render. Zero values pick sensible
// defaults: a 1200×1200 canvas, density 1, and a style and palette
// chosen by the seed.
type Options struct {
	// Style is a registered style name. Empty or "random" lets the seed choose.
	Style string
	// Seed determines everything else.
	Seed uint64
	// Width and Height are in canvas units (pixels at scale 1).
	Width, Height float64
	// Palette is a palette name. Empty or "random" lets the seed choose.
	Palette string
	// Density scales how much the style draws; 1 is the default.
	Density float64
}

// Piece is a rendered canvas together with the choices that produced it.
type Piece struct {
	Canvas  *Canvas
	Style   Style
	Palette Palette
	Seed    uint64
	Options Options
}

// Render draws a piece. It returns an error for an unknown style or
// palette, or if no styles are registered.
func Render(o Options) (*Piece, error) {
	if o.Width <= 0 {
		o.Width = 1200
	}
	if o.Height <= 0 {
		o.Height = o.Width
	}
	if o.Density <= 0 {
		o.Density = 1
	}
	styles := Styles()
	if len(styles) == 0 {
		return nil, errors.New(`nacre: no styles registered (import _ "github.com/devatobuild/nacre/styles")`)
	}

	// Three independent streams so that fixing the style or palette by
	// hand never reshuffles the drawing itself.
	r := NewRand(o.Seed)
	styleR, palR, drawR := r.Fork(), r.Fork(), r.Fork()

	var st Style
	if o.Style == "" || o.Style == "random" {
		st = styles[styleR.Intn(len(styles))]
	} else {
		var ok bool
		if st, ok = LookupStyle(o.Style); !ok {
			return nil, fmt.Errorf("nacre: unknown style %q (have %s)", o.Style, strings.Join(StyleNames(), ", "))
		}
	}

	var pal Palette
	if o.Palette == "" || o.Palette == "random" {
		names := PaletteNames()
		pal = palettes[names[palR.Intn(len(names))]]
	} else {
		var ok bool
		if pal, ok = LookupPalette(o.Palette); !ok {
			return nil, fmt.Errorf("nacre: unknown palette %q (have %s)", o.Palette, strings.Join(PaletteNames(), ", "))
		}
	}

	c := NewCanvas(o.Width, o.Height)
	c.Background = pal.Background
	c.Title = fmt.Sprintf("nacre %s %s %s", st.Name, FormatSeed(o.Seed), pal.Name)
	st.Draw(c, drawR, Params{Palette: pal, Density: o.Density})
	return &Piece{Canvas: c, Style: st, Palette: pal, Seed: o.Seed, Options: o}, nil
}

// SVG writes the piece as SVG.
func (p *Piece) SVG(o SVGOptions) []byte { return p.Canvas.SVG(o) }

// Filename returns a descriptive file name such as "flow-0x3f1a.svg".
func (p *Piece) Filename(ext string) string {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return fmt.Sprintf("%s-%s%s", p.Style.Name, FormatSeed(p.Seed), ext)
}

// RandomSeed returns a fresh seed from the operating system's entropy
// source, falling back to the clock.
func RandomSeed() uint64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return uint64(time.Now().UnixNano())
	}
	return binary.LittleEndian.Uint64(b[:])
}

// ParseSeed turns text into a seed. "0x…" is read as hexadecimal, a
// run of digits as decimal, and any other text ("sea glass", say) is
// hashed so words make perfectly good seeds too.
func ParseSeed(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("nacre: empty seed")
	}
	if len(s) > 2 && (s[:2] == "0x" || s[:2] == "0X") {
		v, err := strconv.ParseUint(s[2:], 16, 64)
		if err != nil {
			return 0, fmt.Errorf("nacre: bad hex seed %q", s)
		}
		return v, nil
	}
	if v, err := strconv.ParseUint(s, 10, 64); err == nil {
		return v, nil
	}
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64(), nil
}

// FormatSeed renders a seed the way ParseSeed likes to read it back.
func FormatSeed(seed uint64) string { return "0x" + strconv.FormatUint(seed, 16) }
