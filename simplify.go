package nacre

import "math"

// Simplify reduces a polyline with the Ramer–Douglas–Peucker algorithm,
// dropping points that deviate less than tolerance from the line
// between their neighbours. It returns a new slice and never removes
// the endpoints.
func Simplify(pts []Point, tolerance float64) []Point {
	if len(pts) < 3 || tolerance <= 0 {
		return append([]Point(nil), pts...)
	}
	keep := make([]bool, len(pts))
	keep[0], keep[len(pts)-1] = true, true
	rdp(pts, 0, len(pts)-1, tolerance*tolerance, keep)
	out := make([]Point, 0, len(pts)/2+2)
	for i, k := range keep {
		if k {
			out = append(out, pts[i])
		}
	}
	return out
}

func rdp(pts []Point, lo, hi int, tol2 float64, keep []bool) {
	if hi-lo < 2 {
		return
	}
	a, b := pts[lo], pts[hi]
	best, bestD := -1, tol2
	for i := lo + 1; i < hi; i++ {
		if d := segDist2(pts[i], a, b); d > bestD {
			best, bestD = i, d
		}
	}
	if best < 0 {
		return
	}
	keep[best] = true
	rdp(pts, lo, best, tol2, keep)
	rdp(pts, best, hi, tol2, keep)
}

// segDist2 is the squared distance from p to segment ab.
func segDist2(p, a, b Point) float64 {
	ab := b.Sub(a)
	l2 := ab.X*ab.X + ab.Y*ab.Y
	if l2 == 0 {
		d := p.Sub(a)
		return d.X*d.X + d.Y*d.Y
	}
	t := ((p.X-a.X)*ab.X + (p.Y-a.Y)*ab.Y) / l2
	t = math.Max(0, math.Min(1, t))
	q := a.Add(ab.Scale(t))
	d := p.Sub(q)
	return d.X*d.X + d.Y*d.Y
}
