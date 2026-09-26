package noise

import (
	"math"
	"testing"
)

func TestRangeAndDeterminism(t *testing.T) {
	a, b := New(1), New(1)
	for y := -5.0; y < 5; y += 0.37 {
		for x := -5.0; x < 5; x += 0.41 {
			v := a.Eval2(x, y)
			if v < -1 || v > 1 || math.IsNaN(v) {
				t.Fatalf("Eval2(%v,%v) = %v", x, y, v)
			}
			if v != b.Eval2(x, y) {
				t.Fatal("same seed, different noise")
			}
			w := a.Eval3(x, y, 0.5)
			if w < -1 || w > 1 || math.IsNaN(w) {
				t.Fatalf("Eval3 = %v", w)
			}
			if f := a.FBM2(x, y, 4, 2, 0.5); f < -1 || f > 1 {
				t.Fatalf("FBM2 = %v", f)
			}
			if rd := a.Ridged2(x, y, 3, 2, 0.5); rd < 0 || rd > 1 {
				t.Fatalf("Ridged2 = %v", rd)
			}
		}
	}
	if New(1).Eval2(0.3, 0.7) == New(2).Eval2(0.3, 0.7) {
		t.Fatal("different seeds should differ")
	}
}

func TestContinuity(t *testing.T) {
	n := New(9)
	for i := 0; i < 1000; i++ {
		x, y := float64(i)*0.013, float64(i)*0.007
		if d := math.Abs(n.Eval2(x, y) - n.Eval2(x+1e-4, y)); d > 1e-2 {
			t.Fatalf("discontinuity at %v,%v: %v", x, y, d)
		}
	}
}

func TestNotFlat(t *testing.T) {
	n := New(4)
	lo, hi := 1.0, -1.0
	for i := 0; i < 2000; i++ {
		v := n.Eval2(float64(i)*0.05, float64(i)*0.031)
		lo, hi = math.Min(lo, v), math.Max(hi, v)
	}
	if hi-lo < 1 {
		t.Fatalf("noise range too narrow: %v..%v", lo, hi)
	}
}
