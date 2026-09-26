package styles

import (
	"math"

	"github.com/devatobuild/nacre"
	"github.com/devatobuild/nacre/noise"
)

func init() {
	nacre.Register(nacre.Style{
		Name: "ridges",
		Info: "Stacked noise horizons that hide one another, the way mountain ranges do.",
		Draw: ridges,
	})
}

func ridges(c *nacre.Canvas, r *nacre.Rand, p nacre.Params) {
	pal := p.Palette
	u := c.Unit()
	n := noise.New(r.Uint64())
	lines := int(p.Density * r.Range(36, 72))
	if lines < 4 {
		lines = 4
	}
	mx := c.Width * r.Range(0.12, 0.2)
	my := c.Height * r.Range(0.12, 0.2)
	const samples = 260
	amp := (c.Height - 2*my) * r.Range(0.1, 0.25)
	envCenter := c.Width/2 + r.Range(-0.1, 0.1)*c.Width
	envWidth := c.Width * r.Range(0.16, 0.3)
	freq := r.Range(2, 6) / c.Width
	lineFreq := r.Range(0.05, 0.25)
	spiky := r.Range(1, 2.6)
	mode := r.Intn(2) // 0 ink, 1 gradient
	width := u * r.Range(0.0012, 0.0025)

	for i := 0; i < lines; i++ {
		yBase := my + (c.Height-2*my)*float64(i)/float64(lines-1)
		pts := make([]nacre.Point, samples+1)
		for s := 0; s <= samples; s++ {
			x := mx + (c.Width-2*mx)*float64(s)/float64(samples)
			env := math.Exp(-math.Pow((x-envCenter)/envWidth, 2))
			v := (n.FBM2(x*freq, float64(i)*lineFreq, 3, 2, 0.5) + 1) / 2
			v = math.Pow(v, spiky)
			v += 0.04 * n.Eval2(x*freq*6, float64(i)*lineFreq*3+10)
			pts[s] = nacre.Pt(x, yBase-amp*env*(v+0.06))
		}
		poly := make([]nacre.Point, 0, samples+3)
		poly = append(poly, pts...)
		poly = append(poly, nacre.Pt(pts[samples].X, c.Height), nacre.Pt(pts[0].X, c.Height))
		c.Fill(poly, pal.Background)
		col := pal.Ink
		if mode == 1 {
			col = pal.At(float64(i) / float64(lines-1))
		}
		c.Stroke(pts, col, width)
	}
}
