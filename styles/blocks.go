package styles

import (
	"math"

	"github.com/devatobuild/nacre"
	"github.com/devatobuild/nacre/noise"
)

func init() {
	nacre.Register(nacre.Style{
		Name: "blocks",
		Info: "Recursive golden-ratio subdivision into blocks, some plain, some hatched or dotted.",
		Draw: blocks,
	})
}

func blocks(c *nacre.Canvas, r *nacre.Rand, p nacre.Params) {
	pal := p.Palette
	u := c.Unit()
	margin := u * r.Range(0.04, 0.08)
	gutter := u * r.Range(0.006, 0.016)
	minSize := u * r.Range(0.05, 0.1) / math.Sqrt(math.Max(0.1, p.Density))
	mode := r.Intn(3) // 0 outlined, 1 soft, 2 detailed
	splitP := r.Range(0.7, 0.92)
	empty := r.Range(0.1, 0.28)
	tintN := noise.New(r.Uint64())
	tintFreq := r.Range(0.6, 1.8) / u

	type rect struct{ x, y, w, h float64 }
	var leaves []rect
	var split func(rc rect, depth int)
	split = func(rc rect, depth int) {
		canV := rc.w > 2*minSize // vertical cut: left and right halves
		canH := rc.h > 2*minSize
		if !canV && !canH || (depth > 2 && !r.Chance(splitP-float64(depth-2)*0.08)) {
			leaves = append(leaves, rc)
			return
		}
		t := nacre.Pick(r, []float64{0.5, 0.382, 0.618, r.Range(0.3, 0.7)})
		if canV && (!canH || rc.w >= rc.h) {
			cut := math.Max(minSize, math.Min(rc.w-minSize, rc.w*t))
			split(rect{rc.x, rc.y, cut - gutter/2, rc.h}, depth+1)
			split(rect{rc.x + cut + gutter/2, rc.y, rc.w - cut - gutter/2, rc.h}, depth+1)
		} else {
			cut := math.Max(minSize, math.Min(rc.h-minSize, rc.h*t))
			split(rect{rc.x, rc.y, rc.w, cut - gutter/2}, depth+1)
			split(rect{rc.x, rc.y + cut + gutter/2, rc.w, rc.h - cut - gutter/2}, depth+1)
		}
	}
	split(rect{margin, margin, c.Width - 2*margin, c.Height - 2*margin}, 0)

	line := u * 0.0025
	for _, rc := range leaves {
		t := (tintN.FBM2((rc.x+rc.w/2)*tintFreq, (rc.y+rc.h/2)*tintFreq, 2, 2, 0.5) + 1) / 2
		col := pal.At(clamp(t + r.Range(-0.2, 0.2)))
		blank := r.Chance(empty)
		if blank {
			col = pal.Background
		}
		pts := rectPoints(rc.x, rc.y, rc.w, rc.h)
		switch mode {
		case 0:
			c.Polygon(pts, col, pal.Ink, line)
		case 1:
			if blank {
				continue
			}
			c.Fill(pts, col.Mix(pal.Background, 0.15))
		case 2:
			if !blank {
				c.Fill(pts, col)
			}
			switch r.Intn(3) {
			case 1:
				hatch(c, rc.x, rc.y, rc.w, rc.h, u*r.Range(0.008, 0.016), pal.Ink.Alpha(0.8), line*0.5)
			case 2:
				dots(c, rc.x, rc.y, rc.w, rc.h, u*r.Range(0.014, 0.024), pal.Ink.Alpha(0.8))
			}
		}
	}
}

// hatch draws 45° lines across a rectangle.
func hatch(c *nacre.Canvas, x, y, w, h, spacing float64, col nacre.Color, width float64) {
	inset := spacing * 0.6
	x0, y0, x1, y1 := x+inset, y+inset, x+w-inset, y+h-inset
	if x1 <= x0 || y1 <= y0 {
		return
	}
	for k := x0 - y1; k <= x1-y0; k += spacing {
		ya := math.Max(y0, x0-k)
		yb := math.Min(y1, x1-k)
		if yb <= ya {
			continue
		}
		c.Stroke([]nacre.Point{{X: ya + k, Y: ya}, {X: yb + k, Y: yb}}, col, width)
	}
}

// dots fills a rectangle with a grid of small discs.
func dots(c *nacre.Canvas, x, y, w, h, spacing float64, col nacre.Color) {
	nx := int(w / spacing)
	ny := int(h / spacing)
	if nx < 1 || ny < 1 {
		return
	}
	ox := x + (w-float64(nx-1)*spacing)/2
	oy := y + (h-float64(ny-1)*spacing)/2
	for j := 0; j < ny; j++ {
		for i := 0; i < nx; i++ {
			c.Circle(nacre.Pt(ox+float64(i)*spacing, oy+float64(j)*spacing), spacing*0.16, col, nacre.None, 0)
		}
	}
}
