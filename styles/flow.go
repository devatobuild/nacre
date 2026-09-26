package styles

import (
	"math"

	"github.com/devatobuild/nacre"
	"github.com/devatobuild/nacre/noise"
)

func init() {
	nacre.Register(nacre.Style{
		Name: "flow",
		Info: "Particles traced through a layered noise field, as hairlines or ribbons.",
		Draw: flow,
	})
}

func flow(c *nacre.Canvas, r *nacre.Rand, p nacre.Params) {
	u := c.Unit()
	pal := p.Palette
	field := noise.New(r.Uint64())
	tint := noise.New(r.Uint64())

	margin := u * r.Range(0.06, 0.12)
	freq := r.Range(0.6, 1.8) / u // field cycles across the canvas
	turns := r.Range(1, 2.4)      // how far the field twists
	octaves := r.IntRange(1, 4)
	swirl := 0.0
	if r.Chance(0.35) {
		swirl = r.Range(0.2, 0.8)
	}
	eye := c.Center().Add(nacre.Pt(r.Range(-0.2, 0.2)*u, r.Range(-0.2, 0.2)*u))

	ribbon := r.Chance(0.3)
	count := int(p.Density * r.Range(1200, 2600))
	baseWidth := u * r.Range(0.0012, 0.0028)
	alpha := 0.85
	if ribbon {
		count /= 5
		baseWidth *= r.Range(3, 6)
		alpha = 0.92
	}
	step := u * 0.0035
	maxSteps := r.IntRange(60, 220)
	tintFreq := r.Range(0.5, 1.5) / u
	accent := r.Range(0.03, 0.12)

	angleAt := func(pt nacre.Point) float64 {
		a := field.FBM2(pt.X*freq, pt.Y*freq, octaves, 2, 0.5) * math.Pi * turns
		if swirl > 0 {
			d := pt.Sub(eye)
			a = a*(1-swirl) + (math.Atan2(d.Y, d.X)+math.Pi/2)*swirl
		}
		return a
	}

	for i := 0; i < count; i++ {
		pt := nacre.Pt(r.Range(margin, c.Width-margin), r.Range(margin, c.Height-margin))
		steps := int(float64(maxSteps) * r.Range(0.3, 1))
		pts := make([]nacre.Point, 0, steps+1)
		pts = append(pts, pt)
		for s := 0; s < steps; s++ {
			pt = pt.Polar(step, angleAt(pt))
			if !c.Contains(pt, margin) {
				break
			}
			pts = append(pts, pt)
		}
		if len(pts) < 3 {
			continue
		}
		t := (tint.FBM2(pts[0].X*tintFreq, pts[0].Y*tintFreq, 2, 2, 0.5) + 1) / 2
		col := pal.At(clamp(t + r.Range(-0.08, 0.08)))
		w := baseWidth * r.Range(0.5, 1.6)
		if r.Chance(accent) {
			col, w = pal.Ink, baseWidth*0.7
		}
		c.Stroke(pts, col.Alpha(alpha), w)
	}
}
