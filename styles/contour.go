package styles

import (
	"math"

	"github.com/devatobuild/nacre"
	"github.com/devatobuild/nacre/noise"
)

func init() {
	nacre.Register(nacre.Style{
		Name: "contour",
		Info: "Topographic isolines of a domain-warped noise landscape.",
		Draw: contour,
	})
}

func contour(c *nacre.Canvas, r *nacre.Rand, p nacre.Params) {
	pal := p.Palette
	u := c.Unit()
	land := noise.New(r.Uint64())
	warpN := noise.New(r.Uint64())

	cells := int(math.Round(110 * math.Max(0.4, math.Min(2.5, p.Density))))
	gw := cells
	gh := int(math.Round(float64(cells) * c.Height / c.Width))
	if gh < 2 {
		gh = 2
	}
	margin := u * r.Range(0.05, 0.1)
	x0, y0 := margin, margin
	ww, hh := c.Width-2*margin, c.Height-2*margin

	freq := r.Range(0.8, 2.2) / u
	octaves := r.IntRange(3, 6)
	warp := r.Range(0, 0.9)
	field := make([]float64, (gw+1)*(gh+1))
	lo, hi := math.Inf(1), math.Inf(-1)
	for j := 0; j <= gh; j++ {
		for i := 0; i <= gw; i++ {
			x := x0 + ww*float64(i)/float64(gw)
			y := y0 + hh*float64(j)/float64(gh)
			qx := warpN.FBM2(x*freq*0.7, y*freq*0.7, 2, 2, 0.5)
			qy := warpN.FBM2(x*freq*0.7+5.2, y*freq*0.7+1.3, 2, 2, 0.5)
			v := land.FBM2((x+warp*qx/freq)*freq, (y+warp*qy/freq)*freq, octaves, 2, 0.5)
			field[j*(gw+1)+i] = v
			lo, hi = math.Min(lo, v), math.Max(hi, v)
		}
	}

	levels := r.IntRange(14, 34)
	mode := r.Intn(2) // 0 gradient, 1 ink
	thin := u * r.Range(0.001, 0.0018)
	every := r.IntRange(4, 6)
	for l := 0; l < levels; l++ {
		t := (float64(l) + 0.5) / float64(levels)
		iso := lo + (hi-lo)*t
		col := pal.Ink
		if mode == 0 {
			col = pal.At(t)
		}
		w := thin
		if l%every == 0 {
			w = thin * 2.2
		}
		for _, line := range marchingSquares(field, gw, gh, iso) {
			if len(line) < 3 {
				continue
			}
			pts := make([]nacre.Point, len(line))
			for i, g := range line {
				pts[i] = nacre.Pt(x0+g.X/float64(gw)*ww, y0+g.Y/float64(gh)*hh)
			}
			c.Stroke(pts, col, w)
		}
	}
}

// marchingSquares extracts the iso-contour of a (gw+1)×(gh+1) grid as
// polylines in grid coordinates, joining cell segments end to end.
func marchingSquares(f []float64, gw, gh int, iso float64) [][]nacre.Point {
	stride := gw + 1
	val := func(i, j int) float64 { return f[j*stride+i] }
	// Edge ids: each grid vertex owns the horizontal edge to its right
	// (2k) and the vertical edge below it (2k+1).
	hID := func(i, j int) int { return (j*stride + i) * 2 }
	vID := func(i, j int) int { return (j*stride+i)*2 + 1 }
	hPt := func(i, j int) nacre.Point {
		a, b := val(i, j), val(i+1, j)
		return nacre.Pt(float64(i)+(iso-a)/(b-a), float64(j))
	}
	vPt := func(i, j int) nacre.Point {
		a, b := val(i, j), val(i, j+1)
		return nacre.Pt(float64(i), float64(j)+(iso-a)/(b-a))
	}

	type seg struct {
		ea, eb int
		pa, pb nacre.Point
	}
	var segs []seg
	adj := make([]int32, 2*2*stride*(gh+1))
	for i := range adj {
		adj[i] = -1
	}
	link := func(e, s int) {
		if adj[2*e] < 0 {
			adj[2*e] = int32(s)
		} else {
			adj[2*e+1] = int32(s)
		}
	}
	emit := func(ea int, pa nacre.Point, eb int, pb nacre.Point) {
		segs = append(segs, seg{ea, eb, pa, pb})
		link(ea, len(segs)-1)
		link(eb, len(segs)-1)
	}

	for j := 0; j < gh; j++ {
		for i := 0; i < gw; i++ {
			idx := 0
			if val(i, j) >= iso {
				idx |= 8
			}
			if val(i+1, j) >= iso {
				idx |= 4
			}
			if val(i+1, j+1) >= iso {
				idx |= 2
			}
			if val(i, j+1) >= iso {
				idx |= 1
			}
			if idx == 0 || idx == 15 {
				continue
			}
			top, right, bottom, left := hID(i, j), vID(i+1, j), hID(i, j+1), vID(i, j)
			tp := func() nacre.Point { return hPt(i, j) }
			rp := func() nacre.Point { return vPt(i+1, j) }
			bp := func() nacre.Point { return hPt(i, j+1) }
			lp := func() nacre.Point { return vPt(i, j) }
			switch idx {
			case 1, 14:
				emit(left, lp(), bottom, bp())
			case 2, 13:
				emit(bottom, bp(), right, rp())
			case 3, 12:
				emit(left, lp(), right, rp())
			case 4, 11:
				emit(top, tp(), right, rp())
			case 6, 9:
				emit(top, tp(), bottom, bp())
			case 7, 8:
				emit(top, tp(), left, lp())
			case 5, 10:
				center := (val(i, j) + val(i+1, j) + val(i+1, j+1) + val(i, j+1)) / 4
				inside := center >= iso
				if (idx == 5) == inside {
					emit(top, tp(), left, lp())
					emit(bottom, bp(), right, rp())
				} else {
					emit(top, tp(), right, rp())
					emit(left, lp(), bottom, bp())
				}
			}
		}
	}

	used := make([]bool, len(segs))
	next := func(edge, from int) int {
		for _, s := range adj[2*edge : 2*edge+2] {
			if s >= 0 && int(s) != from && !used[s] {
				return int(s)
			}
		}
		return -1
	}
	var lines [][]nacre.Point
	for s := range segs {
		if used[s] {
			continue
		}
		used[s] = true
		line := []nacre.Point{segs[s].pa, segs[s].pb}
		// Walk forward from the end, then backward from the start.
		cur, edge := s, segs[s].eb
		for {
			n := next(edge, cur)
			if n < 0 {
				break
			}
			used[n] = true
			if segs[n].ea == edge {
				line = append(line, segs[n].pb)
				edge = segs[n].eb
			} else {
				line = append(line, segs[n].pa)
				edge = segs[n].ea
			}
			cur = n
		}
		closed := edge == segs[s].ea
		if !closed {
			var back []nacre.Point
			cur, edge = s, segs[s].ea
			for {
				n := next(edge, cur)
				if n < 0 {
					break
				}
				used[n] = true
				if segs[n].eb == edge {
					back = append(back, segs[n].pa)
					edge = segs[n].ea
				} else {
					back = append(back, segs[n].pb)
					edge = segs[n].eb
				}
				cur = n
			}
			for i, j := 0, len(back)-1; i < j; i, j = i+1, j-1 {
				back[i], back[j] = back[j], back[i]
			}
			line = append(back, line...)
		}
		lines = append(lines, line)
	}
	return lines
}
