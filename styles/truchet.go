package styles

import (
	"math"

	"github.com/devatobuild/nacre"
	"github.com/devatobuild/nacre/noise"
)

func init() {
	nacre.Register(nacre.Style{
		Name: "truchet",
		Info: "Quarter-circle tiles at several scales that join into endless winding paths.",
		Draw: truchet,
	})
}

func truchet(c *nacre.Canvas, r *nacre.Rand, p nacre.Params) {
	pal := p.Palette
	cols := int(math.Round(p.Density * float64(r.IntRange(6, 13))))
	if cols < 2 {
		cols = 2
	}
	tile := c.Width / float64(cols)
	rows := int(math.Ceil(c.Height / tile))
	mode := r.Intn(3) // 0 tinted tiles + ink arcs, 1 coloured arcs, 2 fine double lines
	subdivide := r.Range(0.12, 0.4)
	widthFrac := r.Range(0.13, 0.22)
	if mode == 2 {
		widthFrac = 0.03
	}
	tintN := noise.New(r.Uint64())
	tintFreq := r.Range(0.6, 1.4) / c.Unit()

	type cell struct{ x, y, size float64 }
	var leaves []cell
	var split func(x, y, size float64, depth int)
	split = func(x, y, size float64, depth int) {
		if depth < 2 && r.Chance(subdivide/float64(depth+1)) {
			h := size / 2
			split(x, y, h, depth+1)
			split(x+h, y, h, depth+1)
			split(x, y+h, h, depth+1)
			split(x+h, y+h, h, depth+1)
			return
		}
		leaves = append(leaves, cell{x, y, size})
	}
	for j := 0; j < rows; j++ {
		for i := 0; i < cols; i++ {
			split(float64(i)*tile, float64(j)*tile, tile, 0)
		}
	}

	// Tile fills go down first so no arc is ever clipped by a neighbour.
	if mode == 0 {
		for _, l := range leaves {
			if !r.Chance(0.55) {
				continue
			}
			t := (tintN.FBM2((l.x+l.size/2)*tintFreq, (l.y+l.size/2)*tintFreq, 2, 2, 0.5) + 1) / 2
			c.Rect(l.x, l.y, l.size, l.size, pal.At(t).Mix(pal.Background, 0.55))
		}
	}
	for _, l := range leaves {
		cx, cy := l.x+l.size/2, l.y+l.size/2
		col := pal.Ink
		if mode == 1 {
			t := (tintN.FBM2(cx*tintFreq, cy*tintFreq, 2, 2, 0.5) + 1) / 2
			col = pal.At(clamp(t + r.Range(-0.15, 0.15)))
		}
		w := l.size * widthFrac
		rad := l.size / 2
		var a, b nacre.Point
		var aa, ba float64
		if r.Chance(0.5) {
			a, aa = nacre.Pt(l.x, l.y), 0
			b, ba = nacre.Pt(l.x+l.size, l.y+l.size), math.Pi
		} else {
			a, aa = nacre.Pt(l.x+l.size, l.y), math.Pi/2
			b, ba = nacre.Pt(l.x, l.y+l.size), 3*math.Pi/2
		}
		n := 10
		if mode == 2 {
			g := l.size * 0.09
			for _, off := range []float64{-g, g} {
				c.Stroke(arc(a, rad+off, aa, aa+math.Pi/2, n), col, w)
				c.Stroke(arc(b, rad+off, ba, ba+math.Pi/2, n), col, w)
			}
			continue
		}
		c.Stroke(arc(a, rad, aa, aa+math.Pi/2, n), col, w)
		c.Stroke(arc(b, rad, ba, ba+math.Pi/2, n), col, w)
	}
}
