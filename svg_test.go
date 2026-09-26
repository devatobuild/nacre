package nacre

import (
	"strings"
	"testing"
)

func TestAppendNum(t *testing.T) {
	cases := []struct {
		v    float64
		prec int
		want string
	}{
		{12.0, 1, "12"}, {12.34, 1, "12.3"}, {-0.04, 1, "0"}, {0.5, 0, "0"}, {1.25, 2, "1.25"}, {100, 2, "100"},
	}
	for _, c := range cases {
		if got := string(appendNum(nil, c.v, c.prec)); got != c.want {
			t.Errorf("appendNum(%v,%d) = %q want %q", c.v, c.prec, got, c.want)
		}
	}
}

func TestWriteSVG(t *testing.T) {
	c := NewCanvas(100, 50)
	c.Title = `a <b> & "c"`
	c.Rect(10, 10, 20, 20, Hex("#ff0000"))
	c.Stroke([]Point{{0, 0}, {50, 25}, {100, 50}}, Hex("#00ff00").Alpha(0.5), 2)
	c.Circle(Pt(50, 25), 5, None, Hex("#0000ff"), 1)
	svg := string(c.SVG(SVGOptions{}))
	for _, want := range []string{
		`viewBox="0 0 100 50"`,
		`<title>a &lt;b&gt; &amp; &quot;c&quot;</title>`,
		`<path d="M10 10 30 10 30 30 10 30Z" fill="#ff0000"/>`,
		`<path d="M0 0 50 25 100 50" stroke="#00ff00" stroke-width="2" stroke-opacity="0.5"/>`,
		`<circle cx="50" cy="25" r="5" stroke="#0000ff" stroke-width="1"/>`,
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("missing %s in:\n%s", want, svg)
		}
	}
	if strings.Contains(svg, "<style>") {
		t.Error("unexpected animation css")
	}
	anim := string(c.SVG(SVGOptions{Animate: true, Duration: 3}))
	if !strings.Contains(anim, "<style>") || !strings.Contains(anim, `class="d"`) || !strings.Contains(anim, `class="f"`) {
		t.Errorf("animation markup missing:\n%s", anim)
	}
}

func TestSimplify(t *testing.T) {
	line := []Point{{0, 0}, {1, 0.01}, {2, -0.01}, {3, 0}, {4, 0}}
	if got := Simplify(line, 0.1); len(got) != 2 || got[0] != line[0] || got[1] != line[4] {
		t.Fatalf("straight line simplified to %v", got)
	}
	bent := []Point{{0, 0}, {5, 5}, {10, 0}}
	if got := Simplify(bent, 0.1); len(got) != 3 {
		t.Fatalf("corner lost: %v", got)
	}
	if got := Simplify(bent, 0); len(got) != 3 {
		t.Fatalf("zero tolerance should keep all: %v", got)
	}
}
