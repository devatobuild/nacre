// Package styles holds nacre's built-in styles. Importing it (usually
// for side effects) registers them:
//
//	import _ "github.com/devatobuild/nacre/styles"
//
// Every style is a plain function of a canvas, a random stream and a
// palette; see nacre.Register to add your own.
package styles

import (
	"math"

	"github.com/devatobuild/nacre"
)

func clamp(v float64) float64 { return math.Max(0, math.Min(1, v)) }

// arc returns n+1 points along a circular arc from angle a0 to a1.
func arc(center nacre.Point, radius, a0, a1 float64, n int) []nacre.Point {
	pts := make([]nacre.Point, n+1)
	for i := 0; i <= n; i++ {
		pts[i] = center.Polar(radius, a0+(a1-a0)*float64(i)/float64(n))
	}
	return pts
}

func rectPoints(x, y, w, h float64) []nacre.Point {
	return []nacre.Point{{X: x, Y: y}, {X: x + w, Y: y}, {X: x + w, Y: y + h}, {X: x, Y: y + h}}
}
