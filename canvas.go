package nacre

import "math"

// Point is a position in canvas units.
type Point struct{ X, Y float64 }

// Pt is shorthand for Point{x, y}.
func Pt(x, y float64) Point { return Point{x, y} }

// Add returns p + q.
func (p Point) Add(q Point) Point { return Point{p.X + q.X, p.Y + q.Y} }

// Sub returns p - q.
func (p Point) Sub(q Point) Point { return Point{p.X - q.X, p.Y - q.Y} }

// Scale returns p scaled by s.
func (p Point) Scale(s float64) Point { return Point{p.X * s, p.Y * s} }

// Len returns the distance from the origin.
func (p Point) Len() float64 { return math.Hypot(p.X, p.Y) }

// Dist returns the distance between p and q.
func (p Point) Dist(q Point) float64 { return math.Hypot(p.X-q.X, p.Y-q.Y) }

// Lerp interpolates from p to q by t.
func (p Point) Lerp(q Point, t float64) Point {
	return Point{p.X + (q.X-p.X)*t, p.Y + (q.Y-p.Y)*t}
}

// Polar returns the point at distance r and angle a (radians) from p.
func (p Point) Polar(r, a float64) Point {
	return Point{p.X + r*math.Cos(a), p.Y + r*math.Sin(a)}
}

// Shape is anything a Canvas can hold. The concrete types are Path and
// Circle; both the SVG writer and the rasterizer understand them.
type Shape interface{ shape() }

// Path is a polyline that can be stroked, filled, or both. A zero
// Stroke or Fill colour means "none".
type Path struct {
	Points []Point
	Closed bool
	Fill   Color
	Stroke Color
	Width  float64
}

// Circle is a disc that can be filled, outlined, or both.
type Circle struct {
	Center Point
	Radius float64
	Fill   Color
	Stroke Color
	Width  float64
}

func (Path) shape()   {}
func (Circle) shape() {}

// Length returns the total length of the polyline (including the closing
// segment when Closed).
func (p Path) Length() float64 {
	l := 0.0
	for i := 1; i < len(p.Points); i++ {
		l += p.Points[i].Dist(p.Points[i-1])
	}
	if p.Closed && len(p.Points) > 2 {
		l += p.Points[0].Dist(p.Points[len(p.Points)-1])
	}
	return l
}

// Canvas is a resolution-independent drawing: a background and an
// ordered list of shapes in canvas units. Styles draw onto a Canvas;
// SVG and PNG are produced from it afterwards.
type Canvas struct {
	Width, Height float64
	Background    Color
	// Title is emitted as the SVG <title> for accessibility.
	Title  string
	Shapes []Shape
}

// NewCanvas returns an empty canvas of the given size.
func NewCanvas(width, height float64) *Canvas {
	return &Canvas{Width: width, Height: height, Background: Color{1, 1, 1, 1}}
}

// Add appends a shape. The canvas keeps a reference to any point slice
// it is given; callers should not modify it afterwards.
func (c *Canvas) Add(s Shape) { c.Shapes = append(c.Shapes, s) }

// Stroke adds an open polyline.
func (c *Canvas) Stroke(pts []Point, col Color, width float64) {
	if len(pts) < 2 || col.IsNone() || width <= 0 {
		return
	}
	c.Add(Path{Points: pts, Stroke: col, Width: width})
}

// Fill adds a filled polygon with no outline.
func (c *Canvas) Fill(pts []Point, col Color) {
	if len(pts) < 3 || col.IsNone() {
		return
	}
	c.Add(Path{Points: pts, Closed: true, Fill: col})
}

// Polygon adds a closed polygon with both fill and outline.
func (c *Canvas) Polygon(pts []Point, fill, stroke Color, width float64) {
	if len(pts) < 3 {
		return
	}
	c.Add(Path{Points: pts, Closed: true, Fill: fill, Stroke: stroke, Width: width})
}

// Rect adds an axis-aligned filled rectangle.
func (c *Canvas) Rect(x, y, w, h float64, fill Color) {
	c.Fill([]Point{{x, y}, {x + w, y}, {x + w, y + h}, {x, y + h}}, fill)
}

// Circle adds a disc.
func (c *Canvas) Circle(center Point, radius float64, fill, stroke Color, width float64) {
	if radius <= 0 || (fill.IsNone() && (stroke.IsNone() || width <= 0)) {
		return
	}
	c.Add(Circle{Center: center, Radius: radius, Fill: fill, Stroke: stroke, Width: width})
}

// Unit returns the shorter side of the canvas. Styles size everything
// relative to it so a piece looks the same at any resolution.
func (c *Canvas) Unit() float64 { return math.Min(c.Width, c.Height) }

// Center returns the middle of the canvas.
func (c *Canvas) Center() Point { return Point{c.Width / 2, c.Height / 2} }

// Contains reports whether p lies inside the canvas, inset by margin.
func (c *Canvas) Contains(p Point, margin float64) bool {
	return p.X >= margin && p.Y >= margin && p.X <= c.Width-margin && p.Y <= c.Height-margin
}
