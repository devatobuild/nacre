package styles

import (
	"bytes"
	"math"
	"testing"

	"github.com/devatobuild/nacre"
)

func TestEveryStyleRendersDeterministically(t *testing.T) {
	for _, name := range nacre.StyleNames() {
		for _, seed := range []uint64{0, 1, 0x3f1a, math.MaxUint64} {
			o := nacre.Options{Style: name, Seed: seed, Width: 300, Height: 200}
			a, err := nacre.Render(o)
			if err != nil {
				t.Fatalf("%s/%x: %v", name, seed, err)
			}
			if len(a.Canvas.Shapes) == 0 {
				t.Fatalf("%s/%x drew nothing", name, seed)
			}
			b, _ := nacre.Render(o)
			if !bytes.Equal(a.SVG(nacre.SVGOptions{}), b.SVG(nacre.SVGOptions{})) {
				t.Fatalf("%s/%x is not deterministic", name, seed)
			}
			if a.Style.Name != name {
				t.Fatalf("style %s rendered as %s", name, a.Style.Name)
			}
		}
	}
}

func TestSeedPicksStyleAndPalette(t *testing.T) {
	a, err := nacre.Render(nacre.Options{Seed: 12, Width: 100})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := nacre.Render(nacre.Options{Seed: 12, Width: 100})
	if a.Style.Name != b.Style.Name || a.Palette.Name != b.Palette.Name {
		t.Fatal("seed should pick the same style and palette")
	}
	if a.Canvas.Height != 100 {
		t.Fatalf("height should default to width, got %v", a.Canvas.Height)
	}
	if _, err := nacre.Render(nacre.Options{Style: "nope"}); err == nil {
		t.Fatal("unknown style should error")
	}
	if _, err := nacre.Render(nacre.Options{Palette: "nope"}); err == nil {
		t.Fatal("unknown palette should error")
	}
}

func TestPaletteDoesNotChangeGeometry(t *testing.T) {
	a, _ := nacre.Render(nacre.Options{Style: "circles", Seed: 5, Width: 200, Palette: "dusk"})
	b, _ := nacre.Render(nacre.Options{Style: "circles", Seed: 5, Width: 200, Palette: "moss"})
	if len(a.Canvas.Shapes) != len(b.Canvas.Shapes) {
		t.Fatal("palette changed the drawing")
	}
	ca, cb := a.Canvas.Shapes[0].(nacre.Circle), b.Canvas.Shapes[0].(nacre.Circle)
	if ca.Center != cb.Center || ca.Radius != cb.Radius {
		t.Fatal("palette moved a circle")
	}
}

func TestEveryPaletteWorks(t *testing.T) {
	for _, p := range nacre.PaletteNames() {
		if _, err := nacre.Render(nacre.Options{Style: "blocks", Seed: 2, Width: 120, Palette: p}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDensityScales(t *testing.T) {
	sparse, _ := nacre.Render(nacre.Options{Style: "flow", Seed: 8, Width: 300, Density: 0.3})
	dense, _ := nacre.Render(nacre.Options{Style: "flow", Seed: 8, Width: 300, Density: 2})
	if len(sparse.Canvas.Shapes) >= len(dense.Canvas.Shapes) {
		t.Fatalf("density: sparse %d dense %d", len(sparse.Canvas.Shapes), len(dense.Canvas.Shapes))
	}
}

func TestMarchingSquaresClosesLoops(t *testing.T) {
	const gw, gh = 20, 20
	f := make([]float64, (gw+1)*(gh+1))
	for j := 0; j <= gh; j++ {
		for i := 0; i <= gw; i++ {
			dx, dy := float64(i)-10, float64(j)-10
			f[j*(gw+1)+i] = -math.Hypot(dx, dy) // a single peak in the middle
		}
	}
	lines := marchingSquares(f, gw, gh, -5)
	if len(lines) != 1 {
		t.Fatalf("expected one contour, got %d", len(lines))
	}
	line := lines[0]
	if len(line) < 20 {
		t.Fatalf("contour too short: %d points", len(line))
	}
	if d := line[0].Dist(line[len(line)-1]); d > 1.01 {
		t.Fatalf("contour should close, ends are %v apart", d)
	}
	for _, p := range line {
		if r := math.Hypot(p.X-10, p.Y-10); math.Abs(r-5) > 0.6 {
			t.Fatalf("point %v is not on the circle (r=%v)", p, r)
		}
	}
}

func TestParseSeed(t *testing.T) {
	cases := map[string]uint64{"0x3f1a": 0x3f1a, "0X10": 16, "42": 42}
	for in, want := range cases {
		got, err := nacre.ParseSeed(in)
		if err != nil || got != want {
			t.Fatalf("ParseSeed(%q) = %v, %v", in, got, err)
		}
	}
	a, _ := nacre.ParseSeed("sea glass")
	b, _ := nacre.ParseSeed("sea glass")
	c, _ := nacre.ParseSeed("sea gloss")
	if a != b || a == c {
		t.Fatal("text seeds should hash stably")
	}
	if _, err := nacre.ParseSeed(""); err == nil {
		t.Fatal("empty seed should fail")
	}
	if _, err := nacre.ParseSeed("0xzz"); err == nil {
		t.Fatal("bad hex should fail")
	}
	if s := nacre.FormatSeed(0x3f1a); s != "0x3f1a" {
		t.Fatalf("FormatSeed = %s", s)
	}
}

func BenchmarkFlowSVG(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p, _ := nacre.Render(nacre.Options{Style: "flow", Seed: 1})
		_ = p.SVG(nacre.SVGOptions{Simplify: 0.35})
	}
}

func BenchmarkFlowPNG(b *testing.B) {
	p, _ := nacre.Render(nacre.Options{Style: "flow", Seed: 1})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Canvas.Image(1)
	}
}

func TestBlocksNeverNearlyEmpty(t *testing.T) {
	for seed := uint64(0); seed < 200; seed++ {
		p, _ := nacre.Render(nacre.Options{Style: "blocks", Seed: seed, Width: 300})
		if n := len(p.Canvas.Shapes); n < 8 {
			t.Fatalf("seed %d drew only %d shapes", seed, n)
		}
	}
}
