// Package noise implements simplex noise in two and three dimensions,
// plus fractional Brownian motion over it. It is self-contained: seed
// it with any 64-bit value and it will always produce the same field.
package noise

import "math"

// Noise is a seeded simplex noise field.
type Noise struct {
	perm      [512]int
	permMod12 [512]int
}

// New returns a noise field determined by seed.
func New(seed uint64) *Noise {
	var p [256]int
	for i := range p {
		p[i] = i
	}
	x := seed
	next := func() uint64 {
		x += 0x9e3779b97f4a7c15
		z := x
		z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
		z = (z ^ (z >> 27)) * 0x94d049bb133111eb
		return z ^ (z >> 31)
	}
	for i := 255; i > 0; i-- {
		j := int(next() % uint64(i+1))
		p[i], p[j] = p[j], p[i]
	}
	n := &Noise{}
	for i := 0; i < 512; i++ {
		n.perm[i] = p[i&255]
		n.permMod12[i] = n.perm[i] % 12
	}
	return n
}

var grad3 = [12][3]float64{
	{1, 1, 0}, {-1, 1, 0}, {1, -1, 0}, {-1, -1, 0},
	{1, 0, 1}, {-1, 0, 1}, {1, 0, -1}, {-1, 0, -1},
	{0, 1, 1}, {0, -1, 1}, {0, 1, -1}, {0, -1, -1},
}

var (
	skew2   = 0.5 * (math.Sqrt(3) - 1)
	unskew2 = (3 - math.Sqrt(3)) / 6
)

const (
	skew3   = 1.0 / 3
	unskew3 = 1.0 / 6
)

func floor(x float64) int {
	i := int(x)
	if x < float64(i) {
		i--
	}
	return i
}

// Eval2 returns 2-D simplex noise at (x, y), in roughly [-1, 1].
func (n *Noise) Eval2(x, y float64) float64 {
	s := (x + y) * skew2
	i, j := floor(x+s), floor(y+s)
	t := float64(i+j) * unskew2
	x0 := x - (float64(i) - t)
	y0 := y - (float64(j) - t)
	var i1, j1 int
	if x0 > y0 {
		i1, j1 = 1, 0
	} else {
		i1, j1 = 0, 1
	}
	x1 := x0 - float64(i1) + unskew2
	y1 := y0 - float64(j1) + unskew2
	x2 := x0 - 1 + 2*unskew2
	y2 := y0 - 1 + 2*unskew2
	ii, jj := i&255, j&255
	gi0 := n.permMod12[ii+n.perm[jj]]
	gi1 := n.permMod12[ii+i1+n.perm[jj+j1]]
	gi2 := n.permMod12[ii+1+n.perm[jj+1]]
	var n0, n1, n2 float64
	if t0 := 0.5 - x0*x0 - y0*y0; t0 > 0 {
		t0 *= t0
		n0 = t0 * t0 * (grad3[gi0][0]*x0 + grad3[gi0][1]*y0)
	}
	if t1 := 0.5 - x1*x1 - y1*y1; t1 > 0 {
		t1 *= t1
		n1 = t1 * t1 * (grad3[gi1][0]*x1 + grad3[gi1][1]*y1)
	}
	if t2 := 0.5 - x2*x2 - y2*y2; t2 > 0 {
		t2 *= t2
		n2 = t2 * t2 * (grad3[gi2][0]*x2 + grad3[gi2][1]*y2)
	}
	return 70 * (n0 + n1 + n2)
}

// Eval3 returns 3-D simplex noise at (x, y, z), in roughly [-1, 1].
// The third coordinate is handy as time.
func (n *Noise) Eval3(x, y, z float64) float64 {
	s := (x + y + z) * skew3
	i, j, k := floor(x+s), floor(y+s), floor(z+s)
	t := float64(i+j+k) * unskew3
	x0 := x - (float64(i) - t)
	y0 := y - (float64(j) - t)
	z0 := z - (float64(k) - t)
	var i1, j1, k1, i2, j2, k2 int
	if x0 >= y0 {
		switch {
		case y0 >= z0:
			i1, j1, k1, i2, j2, k2 = 1, 0, 0, 1, 1, 0
		case x0 >= z0:
			i1, j1, k1, i2, j2, k2 = 1, 0, 0, 1, 0, 1
		default:
			i1, j1, k1, i2, j2, k2 = 0, 0, 1, 1, 0, 1
		}
	} else {
		switch {
		case y0 < z0:
			i1, j1, k1, i2, j2, k2 = 0, 0, 1, 0, 1, 1
		case x0 < z0:
			i1, j1, k1, i2, j2, k2 = 0, 1, 0, 0, 1, 1
		default:
			i1, j1, k1, i2, j2, k2 = 0, 1, 0, 1, 1, 0
		}
	}
	x1, y1, z1 := x0-float64(i1)+unskew3, y0-float64(j1)+unskew3, z0-float64(k1)+unskew3
	x2, y2, z2 := x0-float64(i2)+2*unskew3, y0-float64(j2)+2*unskew3, z0-float64(k2)+2*unskew3
	x3, y3, z3 := x0-1+3*unskew3, y0-1+3*unskew3, z0-1+3*unskew3
	ii, jj, kk := i&255, j&255, k&255
	gi0 := n.permMod12[ii+n.perm[jj+n.perm[kk]]]
	gi1 := n.permMod12[ii+i1+n.perm[jj+j1+n.perm[kk+k1]]]
	gi2 := n.permMod12[ii+i2+n.perm[jj+j2+n.perm[kk+k2]]]
	gi3 := n.permMod12[ii+1+n.perm[jj+1+n.perm[kk+1]]]
	var n0, n1, n2, n3 float64
	if t0 := 0.6 - x0*x0 - y0*y0 - z0*z0; t0 > 0 {
		t0 *= t0
		n0 = t0 * t0 * dot3(grad3[gi0], x0, y0, z0)
	}
	if t1 := 0.6 - x1*x1 - y1*y1 - z1*z1; t1 > 0 {
		t1 *= t1
		n1 = t1 * t1 * dot3(grad3[gi1], x1, y1, z1)
	}
	if t2 := 0.6 - x2*x2 - y2*y2 - z2*z2; t2 > 0 {
		t2 *= t2
		n2 = t2 * t2 * dot3(grad3[gi2], x2, y2, z2)
	}
	if t3 := 0.6 - x3*x3 - y3*y3 - z3*z3; t3 > 0 {
		t3 *= t3
		n3 = t3 * t3 * dot3(grad3[gi3], x3, y3, z3)
	}
	return 32 * (n0 + n1 + n2 + n3)
}

func dot3(g [3]float64, x, y, z float64) float64 { return g[0]*x + g[1]*y + g[2]*z }

// FBM2 layers octaves of Eval2, each at lacunarity times the frequency
// and gain times the amplitude of the last, normalised to [-1, 1].
func (n *Noise) FBM2(x, y float64, octaves int, lacunarity, gain float64) float64 {
	sum, amp, freq, norm := 0.0, 1.0, 1.0, 0.0
	for o := 0; o < octaves; o++ {
		sum += amp * n.Eval2(x*freq, y*freq)
		norm += amp
		amp *= gain
		freq *= lacunarity
	}
	return sum / norm
}

// FBM3 is FBM2 in three dimensions.
func (n *Noise) FBM3(x, y, z float64, octaves int, lacunarity, gain float64) float64 {
	sum, amp, freq, norm := 0.0, 1.0, 1.0, 0.0
	for o := 0; o < octaves; o++ {
		sum += amp * n.Eval3(x*freq, y*freq, z*freq)
		norm += amp
		amp *= gain
		freq *= lacunarity
	}
	return sum / norm
}

// Ridged2 returns ridged multifractal noise in [0, 1]: sharp creases
// where the underlying noise crosses zero.
func (n *Noise) Ridged2(x, y float64, octaves int, lacunarity, gain float64) float64 {
	sum, amp, freq, norm := 0.0, 1.0, 1.0, 0.0
	for o := 0; o < octaves; o++ {
		v := 1 - math.Abs(n.Eval2(x*freq, y*freq))
		sum += amp * v * v
		norm += amp
		amp *= gain
		freq *= lacunarity
	}
	return sum / norm
}
