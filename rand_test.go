package nacre

import (
	"math"
	"testing"
)

func TestRandDeterministic(t *testing.T) {
	a, b := NewRand(42), NewRand(42)
	for i := 0; i < 1000; i++ {
		if a.Uint64() != b.Uint64() {
			t.Fatalf("streams diverged at %d", i)
		}
	}
	if NewRand(1).Uint64() == NewRand(2).Uint64() {
		t.Fatal("different seeds gave the same first value")
	}
}

func TestRandRanges(t *testing.T) {
	r := NewRand(7)
	for i := 0; i < 10000; i++ {
		if f := r.Float(); f < 0 || f >= 1 {
			t.Fatalf("Float out of range: %v", f)
		}
		if n := r.Intn(5); n < 0 || n >= 5 {
			t.Fatalf("Intn out of range: %d", n)
		}
		if v := r.Range(-2, 3); v < -2 || v >= 3 {
			t.Fatalf("Range out of range: %v", v)
		}
		if n := r.IntRange(10, 12); n < 10 || n >= 12 {
			t.Fatalf("IntRange out of range: %d", n)
		}
	}
}

func TestRandGauss(t *testing.T) {
	r := NewRand(99)
	const n = 200000
	sum, sq := 0.0, 0.0
	for i := 0; i < n; i++ {
		g := r.Gauss()
		sum += g
		sq += g * g
	}
	mean := sum / n
	sd := math.Sqrt(sq/n - mean*mean)
	if math.Abs(mean) > 0.02 || math.Abs(sd-1) > 0.02 {
		t.Fatalf("gauss mean %.3f sd %.3f", mean, sd)
	}
}

func TestRandShuffleIsPermutation(t *testing.T) {
	r := NewRand(3)
	xs := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r.Shuffle(len(xs), func(i, j int) { xs[i], xs[j] = xs[j], xs[i] })
	seen := make([]bool, len(xs))
	for _, x := range xs {
		if seen[x] {
			t.Fatalf("duplicate %d", x)
		}
		seen[x] = true
	}
}

func TestRandForkIndependent(t *testing.T) {
	r := NewRand(5)
	a := r.Fork()
	b := r.Fork()
	if a.Uint64() == b.Uint64() {
		t.Fatal("forks should differ")
	}
}

func TestWeightedAndPick(t *testing.T) {
	r := NewRand(11)
	counts := [3]int{}
	for i := 0; i < 3000; i++ {
		counts[r.Weighted([]float64{1, 0, 2})]++
	}
	if counts[1] != 0 || counts[2] < counts[0] {
		t.Fatalf("weighted counts %v", counts)
	}
	if p := Pick(r, []string{"only"}); p != "only" {
		t.Fatalf("Pick = %q", p)
	}
}
