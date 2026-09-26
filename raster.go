package nacre

import (
	"image"
	"image/png"
	"io"
	"math"
	"slices"
)

// Image rasterizes the canvas at scale pixels per canvas unit and
// returns an opaque, anti-aliased image. Strokes and circles are drawn
// from exact signed-distance coverage; polygons are filled by a
// scanline rasterizer with vertical supersampling and exact horizontal
// coverage. There are no dependencies beyond the standard library.
func (c *Canvas) Image(scale float64) *image.RGBA {
	if scale <= 0 {
		scale = 1
	}
	w := int(math.Ceil(c.Width * scale))
	h := int(math.Ceil(c.Height * scale))
	r := newRaster(w, h, c.Background)
	for _, s := range c.Shapes {
		switch s := s.(type) {
		case Path:
			r.path(s, scale)
		case Circle:
			r.circle(s, scale)
		}
	}
	return r.image()
}

// WritePNG encodes the rasterized canvas as PNG.
func (c *Canvas) WritePNG(w io.Writer, scale float64) error {
	return png.Encode(w, c.Image(scale))
}

// raster accumulates one shape at a time into a coverage mask, then
// composites the mask with the shape's colour. Overlapping parts of a
// single shape (a polyline's segments, for instance) take the maximum
// coverage, so a stroke never double-blends where it joins itself.
type raster struct {
	w, h    int
	pix     []float32 // RGB, three per pixel
	mask    []float32
	touched []int32
	row     []float32
	cross   []crossing
	tmp     []Point
}

type crossing struct {
	x   float64
	dir int
}

func newRaster(w, h int, bg Color) *raster {
	r := &raster{
		w: w, h: h,
		pix:  make([]float32, w*h*3),
		mask: make([]float32, w*h),
		row:  make([]float32, w+2),
	}
	br, bgc, bb := float32(bg.R), float32(bg.G), float32(bg.B)
	for i := 0; i < w*h; i++ {
		r.pix[i*3], r.pix[i*3+1], r.pix[i*3+2] = br, bgc, bb
	}
	return r
}

func (r *raster) mark(idx int, cov float32) {
	if cov <= 0 {
		return
	}
	if cov > 1 {
		cov = 1
	}
	m := r.mask[idx]
	if m == 0 {
		r.touched = append(r.touched, int32(idx))
	}
	if cov > m {
		r.mask[idx] = cov
	}
}

func (r *raster) flush(col Color) {
	cr, cg, cb, ca := float32(col.R), float32(col.G), float32(col.B), float32(col.A)
	for _, idx := range r.touched {
		a := r.mask[idx] * ca
		p := r.pix[idx*3 : idx*3+3 : idx*3+3]
		p[0] += (cr - p[0]) * a
		p[1] += (cg - p[1]) * a
		p[2] += (cb - p[2]) * a
		r.mask[idx] = 0
	}
	r.touched = r.touched[:0]
}

func (r *raster) bbox(minx, miny, maxx, maxy float64) (x0, y0, x1, y1 int, ok bool) {
	x0 = int(math.Floor(minx))
	y0 = int(math.Floor(miny))
	x1 = int(math.Ceil(maxx))
	y1 = int(math.Ceil(maxy))
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > r.w-1 {
		x1 = r.w - 1
	}
	if y1 > r.h-1 {
		y1 = r.h - 1
	}
	return x0, y0, x1, y1, x0 <= x1 && y0 <= y1
}

// capsule marks coverage for a segment from a to b with half-width hw.
func (r *raster) capsule(a, b Point, hw float64) {
	x0, y0, x1, y1, ok := r.bbox(
		math.Min(a.X, b.X)-hw-1, math.Min(a.Y, b.Y)-hw-1,
		math.Max(a.X, b.X)+hw+1, math.Max(a.Y, b.Y)+hw+1)
	if !ok {
		return
	}
	ab := b.Sub(a)
	l2 := ab.X*ab.X + ab.Y*ab.Y
	for py := y0; py <= y1; py++ {
		cy := float64(py) + 0.5
		for px := x0; px <= x1; px++ {
			cx := float64(px) + 0.5
			t := 0.0
			if l2 > 0 {
				t = ((cx-a.X)*ab.X + (cy-a.Y)*ab.Y) / l2
				if t < 0 {
					t = 0
				} else if t > 1 {
					t = 1
				}
			}
			dx, dy := cx-(a.X+ab.X*t), cy-(a.Y+ab.Y*t)
			cov := hw + 0.5 - math.Sqrt(dx*dx+dy*dy)
			if cov > 0 {
				r.mark(py*r.w+px, float32(cov))
			}
		}
	}
}

// disc marks coverage for a filled circle.
func (r *raster) disc(c Point, rad float64) {
	x0, y0, x1, y1, ok := r.bbox(c.X-rad-1, c.Y-rad-1, c.X+rad+1, c.Y+rad+1)
	if !ok {
		return
	}
	for py := y0; py <= y1; py++ {
		dy := float64(py) + 0.5 - c.Y
		for px := x0; px <= x1; px++ {
			dx := float64(px) + 0.5 - c.X
			cov := rad + 0.5 - math.Sqrt(dx*dx+dy*dy)
			if cov > 0 {
				r.mark(py*r.w+px, float32(cov))
			}
		}
	}
}

// ring marks coverage for a circle outline of half-width hw.
func (r *raster) ring(c Point, rad, hw float64) {
	x0, y0, x1, y1, ok := r.bbox(c.X-rad-hw-1, c.Y-rad-hw-1, c.X+rad+hw+1, c.Y+rad+hw+1)
	if !ok {
		return
	}
	for py := y0; py <= y1; py++ {
		dy := float64(py) + 0.5 - c.Y
		for px := x0; px <= x1; px++ {
			dx := float64(px) + 0.5 - c.X
			cov := hw + 0.5 - math.Abs(math.Sqrt(dx*dx+dy*dy)-rad)
			if cov > 0 {
				r.mark(py*r.w+px, float32(cov))
			}
		}
	}
}

// polygon marks coverage for a filled polygon using the non-zero
// winding rule, four sub-scanlines per pixel row and exact horizontal
// span coverage.
func (r *raster) polygon(pts []Point) {
	n := len(pts)
	if n < 3 {
		return
	}
	type edge struct {
		x0, y0, x1, y1 float64
		dir            int
	}
	edges := make([]edge, 0, n)
	minx, miny := math.Inf(1), math.Inf(1)
	maxx, maxy := math.Inf(-1), math.Inf(-1)
	for i := range pts {
		p, q := pts[i], pts[(i+1)%n]
		minx, maxx = math.Min(minx, p.X), math.Max(maxx, p.X)
		miny, maxy = math.Min(miny, p.Y), math.Max(maxy, p.Y)
		if p.Y == q.Y {
			continue
		}
		dir := 1
		if p.Y > q.Y {
			p, q, dir = q, p, -1
		}
		edges = append(edges, edge{p.X, p.Y, q.X, q.Y, dir})
	}
	if len(edges) < 2 {
		return
	}
	x0, y0, x1, y1, ok := r.bbox(minx, miny, maxx, maxy)
	if !ok {
		return
	}
	const sub = 4
	const inc = 1.0 / sub
	wf := float64(r.w)
	for py := y0; py <= y1; py++ {
		row := r.row[x0 : x1+2]
		clear(row)
		hit := false
		for s := 0; s < sub; s++ {
			y := float64(py) + (float64(s)+0.5)*inc
			cross := r.cross[:0]
			for i := range edges {
				e := &edges[i]
				if y < e.y0 || y >= e.y1 {
					continue
				}
				x := e.x0 + (y-e.y0)*(e.x1-e.x0)/(e.y1-e.y0)
				cross = append(cross, crossing{x, e.dir})
			}
			r.cross = cross
			if len(cross) < 2 {
				continue
			}
			slices.SortFunc(cross, func(a, b crossing) int {
				if a.x < b.x {
					return -1
				}
				if a.x > b.x {
					return 1
				}
				return 0
			})
			wind := 0
			for k := 0; k < len(cross)-1; k++ {
				wind += cross[k].dir
				if wind == 0 {
					continue
				}
				xa, xb := cross[k].x, cross[k+1].x
				if xa < 0 {
					xa = 0
				}
				if xb > wf {
					xb = wf
				}
				if xb <= xa {
					continue
				}
				hit = true
				ia, ib := int(xa), int(xb)
				if ia == ib {
					r.row[ia] += float32((xb - xa) * inc)
					continue
				}
				r.row[ia] += float32((float64(ia+1) - xa) * inc)
				for i := ia + 1; i < ib; i++ {
					r.row[i] += inc
				}
				if ib < r.w {
					r.row[ib] += float32((xb - float64(ib)) * inc)
				}
			}
		}
		if !hit {
			continue
		}
		base := py * r.w
		for px := x0; px <= x1; px++ {
			if v := r.row[px]; v > 0 {
				r.mark(base+px, v)
			}
		}
	}
}

func (r *raster) scaled(pts []Point, scale float64) []Point {
	r.tmp = r.tmp[:0]
	for _, p := range pts {
		r.tmp = append(r.tmp, Point{p.X * scale, p.Y * scale})
	}
	return r.tmp
}

func (r *raster) path(p Path, scale float64) {
	pts := r.scaled(p.Points, scale)
	if !p.Fill.IsNone() && len(pts) >= 3 {
		r.polygon(pts)
		r.flush(p.Fill)
	}
	if !p.Stroke.IsNone() && p.Width > 0 && len(pts) >= 2 {
		hw := p.Width * scale / 2
		for i := 1; i < len(pts); i++ {
			r.capsule(pts[i-1], pts[i], hw)
		}
		if p.Closed {
			r.capsule(pts[len(pts)-1], pts[0], hw)
		}
		r.flush(p.Stroke)
	}
}

func (r *raster) circle(c Circle, scale float64) {
	center := c.Center.Scale(scale)
	rad := c.Radius * scale
	if !c.Fill.IsNone() {
		r.disc(center, rad)
		r.flush(c.Fill)
	}
	if !c.Stroke.IsNone() && c.Width > 0 {
		r.ring(center, rad, c.Width*scale/2)
		r.flush(c.Stroke)
	}
}

func (r *raster) image() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, r.w, r.h))
	for i := 0; i < r.w*r.h; i++ {
		img.Pix[i*4] = q8(r.pix[i*3])
		img.Pix[i*4+1] = q8(r.pix[i*3+1])
		img.Pix[i*4+2] = q8(r.pix[i*3+2])
		img.Pix[i*4+3] = 255
	}
	return img
}

func q8(v float32) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}
