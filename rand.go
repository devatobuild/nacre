package nacre

import "math"

// Rand is a small, fast, deterministic pseudo-random number generator.
// It is xoshiro256** seeded through SplitMix64, so a single 64-bit seed
// fully determines every number it will ever produce. Every piece nacre
// renders is derived from one of these, which is what makes a seed a
// complete recipe for a picture.
type Rand struct {
	s     [4]uint64
	spare float64
	has   bool
}

func splitmix64(x *uint64) uint64 {
	*x += 0x9e3779b97f4a7c15
	z := *x
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// NewRand returns a generator whose entire output is determined by seed.
func NewRand(seed uint64) *Rand {
	r := &Rand{}
	x := seed
	for i := range r.s {
		r.s[i] = splitmix64(&x)
	}
	return r
}

func rotl(x uint64, k uint) uint64 { return (x << k) | (x >> (64 - k)) }

// Uint64 returns the next 64 random bits.
func (r *Rand) Uint64() uint64 {
	s := &r.s
	result := rotl(s[1]*5, 7) * 9
	t := s[1] << 17
	s[2] ^= s[0]
	s[3] ^= s[1]
	s[1] ^= s[2]
	s[0] ^= s[3]
	s[2] ^= t
	s[3] = rotl(s[3], 45)
	return result
}

// Float returns a uniform float64 in [0, 1).
func (r *Rand) Float() float64 { return float64(r.Uint64()>>11) / (1 << 53) }

// Range returns a uniform float64 in [lo, hi).
func (r *Rand) Range(lo, hi float64) float64 { return lo + (hi-lo)*r.Float() }

// Intn returns a uniform int in [0, n). It panics if n <= 0.
func (r *Rand) Intn(n int) int {
	if n <= 0 {
		panic("nacre: Intn called with n <= 0")
	}
	return int(r.Float() * float64(n))
}

// IntRange returns a uniform int in [lo, hi).
func (r *Rand) IntRange(lo, hi int) int { return lo + r.Intn(hi-lo) }

// Chance reports true with probability p.
func (r *Rand) Chance(p float64) bool { return r.Float() < p }

// Sign returns -1 or +1 with equal probability.
func (r *Rand) Sign() float64 {
	if r.Chance(0.5) {
		return 1
	}
	return -1
}

// Gauss returns a normally distributed float64 with mean 0 and
// standard deviation 1, using the Box–Muller transform.
func (r *Rand) Gauss() float64 {
	if r.has {
		r.has = false
		return r.spare
	}
	var u, v, s float64
	for {
		u = r.Range(-1, 1)
		v = r.Range(-1, 1)
		s = u*u + v*v
		if s > 0 && s < 1 {
			break
		}
	}
	m := math.Sqrt(-2 * math.Log(s) / s)
	r.spare, r.has = v*m, true
	return u * m
}

// Fork derives an independent generator from this one. Forking lets a
// piece hand sub-streams to different concerns (palette choice, layout,
// texture) so that a change in one does not reshuffle the others.
func (r *Rand) Fork() *Rand { return NewRand(r.Uint64()) }

// Shuffle performs a Fisher–Yates shuffle over n elements using swap.
func (r *Rand) Shuffle(n int, swap func(i, j int)) {
	for i := n - 1; i > 0; i-- {
		swap(i, r.Intn(i+1))
	}
}

// Weighted returns an index in [0, len(weights)) chosen with probability
// proportional to its weight.
func (r *Rand) Weighted(weights []float64) int {
	total := 0.0
	for _, w := range weights {
		total += w
	}
	x := r.Float() * total
	for i, w := range weights {
		x -= w
		if x < 0 {
			return i
		}
	}
	return len(weights) - 1
}

// Pick returns a uniformly chosen element of xs. It panics if xs is empty.
func Pick[T any](r *Rand, xs []T) T { return xs[r.Intn(len(xs))] }
