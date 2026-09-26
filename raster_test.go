package nacre

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRasterFillRect(t *testing.T) {
	c := NewCanvas(40, 40)
	c.Background = Hex("#000")
	c.Rect(10, 10, 20, 20, Hex("#fff"))
	img := c.Image(1)
	at := func(x, y int) uint8 { return img.Pix[(y*40+x)*4] }
	if at(20, 20) != 255 {
		t.Fatalf("inside = %d", at(20, 20))
	}
	if at(5, 5) != 0 || at(35, 35) != 0 {
		t.Fatal("outside should be background")
	}
	if at(9, 20) != 0 || at(10, 20) != 255 || at(29, 20) != 255 || at(30, 20) != 0 {
		t.Fatalf("edges: %d %d %d %d", at(9, 20), at(10, 20), at(29, 20), at(30, 20))
	}
	// A half-pixel offset should give half coverage.
	c2 := NewCanvas(40, 40)
	c2.Background = Hex("#000")
	c2.Rect(10.5, 10, 20, 20, Hex("#fff"))
	img2 := c2.Image(1)
	if v := img2.Pix[(20*40+10)*4]; v < 120 || v > 135 {
		t.Fatalf("half coverage = %d", v)
	}
}

func TestRasterCircleAndStroke(t *testing.T) {
	c := NewCanvas(40, 40)
	c.Background = Hex("#000")
	c.Circle(Pt(20, 20), 8, Hex("#fff"), None, 0)
	c.Stroke([]Point{{2, 36}, {38, 36}}, Hex("#fff"), 2)
	img := c.Image(1)
	at := func(x, y int) uint8 { return img.Pix[(y*40+x)*4] }
	if at(20, 20) != 255 || at(20, 30) != 0 {
		t.Fatalf("circle: centre %d, outside %d", at(20, 20), at(20, 30))
	}
	if at(20, 35) != 255 || at(20, 32) != 0 {
		t.Fatalf("stroke: on %d, off %d", at(20, 35), at(20, 32))
	}
	if at(0, 35) != 0 {
		t.Fatal("round cap should not reach x=0")
	}
	var buf bytes.Buffer
	if err := c.WritePNG(&buf, 2); err != nil {
		t.Fatal(err)
	}
	decoded, err := png.Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if b := decoded.Bounds(); b.Dx() != 80 || b.Dy() != 80 {
		t.Fatalf("scaled bounds %v", b)
	}
}

func TestRasterUnionOfSegments(t *testing.T) {
	// Two overlapping segments of one stroke with 50% alpha must not
	// double-blend where they overlap.
	c := NewCanvas(40, 40)
	c.Background = Hex("#000")
	c.Stroke([]Point{{5, 20}, {20, 20}, {35, 20}}, Hex("#fff").Alpha(0.5), 4)
	img := c.Image(1)
	mid := img.Pix[(20*40+20)*4]
	side := img.Pix[(20*40+12)*4]
	if mid != side {
		t.Fatalf("join %d differs from body %d", mid, side)
	}
}
