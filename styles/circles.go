package styles

import (
	"math"
	"sort"

	"github.com/devatobuild/nacre"
	"github.com/devatobuild/nacre/noise"
)

func init() {
	nacre.Register(nacre.Style{
		Name: "circles",
		Info: "Circle packing, coloured by a slow noise tide, as discs or nested rings.",
		Draw: circles,
	})
}

func circles(c *nacre.Canvas, r *nacre.Rand, p nacre.Params) {
	pal := p.Palette
	u := c.Unit()
	margin := u * r.Range(0.05, 0.1)
	minR := u * r.Range(0.003, 0.006)
	maxR := u * r.Range(0.07, 0.18)
	gap := u * r.Range(0.002, 0.006)
	attempts := int(p.Density * r.Range(5000, 9000))
	mode := r.Intn(3) // 0 discs, 1 rings, 2 discs with ink dots
	tintN := noise.New(r.Uint64())
	tintFreq := r.Range(0.5, 1.6) / u

	type circ struct {
		p nacre.Point
		r float64
	}
	var cs []circ
	for i := 0; i < attempts; i++ {
		pt := nacre.Pt(r.Range(margin, c.Width-margin), r.Range(margin, c.Height-margin))
		allowed := math.Min(math.Min(pt.X-margin, c.Width-margin-pt.X), math.Min(pt.Y-margin, c.Height-margin-pt.Y))
		for _, o := range cs {
			if d := pt.Dist(o.p) - o.r - gap; d < allowed {
				allowed = d
				if allowed < minR {
					break
				}
			}
		}
		if allowed < minR {
			continue
		}
		rad := math.Max(minR, math.Min(allowed, maxR)*r.Range(0.6, 1))
		cs = append(cs, circ{pt, rad})
	}
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].r > cs[j].r })

	ringW := u * 0.0012
	for _, ci := range cs {
		t := (tintN.FBM2(ci.p.X*tintFreq, ci.p.Y*tintFreq, 2, 2, 0.5)+1)/2 + r.Range(-0.1, 0.1)
		col := pal.At(clamp(t))
		switch mode {
		case 0:
			c.Circle(ci.p, ci.r, col, pal.Ink.Alpha(0.25), u*0.0008)
		case 1:
			spacing := math.Max(u*0.006, ci.r/float64(r.IntRange(3, 9)))
			for rr := ci.r; rr > spacing*0.6; rr -= spacing {
				c.Circle(ci.p, rr, nacre.None, col, ringW)
			}
		case 2:
			if ci.r > maxR*0.25 {
				c.Circle(ci.p, ci.r, col, nacre.None, 0)
			} else {
				c.Circle(ci.p, ci.r, pal.Ink.Alpha(0.85), nacre.None, 0)
			}
		}
	}
}
