package nacre

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// SVGOptions controls how a Canvas is written as SVG.
type SVGOptions struct {
	// Precision is the number of decimal places kept for coordinates.
	// Zero means one decimal; use -1 for integers.
	Precision int
	// Simplify drops polyline points that deviate less than this many
	// canvas units from a straight line (Ramer–Douglas–Peucker). It can
	// shrink files by several times with no visible change. Zero keeps
	// every point.
	Simplify float64
	// Animate makes strokes draw themselves in and fills fade in when
	// the SVG is opened. It uses only CSS, so it works inside <img>.
	Animate bool
	// Duration is how long the whole animation takes, in seconds.
	// Zero means six seconds.
	Duration float64
}

// SVG renders the canvas as an SVG document.
func (c *Canvas) SVG(o SVGOptions) []byte {
	var b bytes.Buffer
	_ = c.WriteSVG(&b, o)
	return b.Bytes()
}

// WriteSVG writes the canvas as an SVG document to w.
func (c *Canvas) WriteSVG(w io.Writer, o SVGOptions) error {
	prec := o.Precision
	switch {
	case prec == 0:
		prec = 1
	case prec < 0:
		prec = 0
	}
	dur := o.Duration
	if dur <= 0 {
		dur = 6
	}
	sw := &svgWriter{buf: make([]byte, 0, 1<<16), prec: prec}

	sw.s(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 `)
	sw.n(c.Width).sp().n(c.Height).s(`" width="`).n(c.Width).s(`" height="`).n(c.Height).s("\">\n")
	if c.Title != "" {
		sw.s("<title>").s(escapeXML(c.Title)).s("</title>\n")
	}
	if o.Animate {
		sw.s(animationCSS)
	}
	sw.s(`<rect width="`).n(c.Width).s(`" height="`).n(c.Height).s(`" fill="`).s(c.Background.Hex()).s("\"/>\n")
	sw.s("<g fill=\"none\" stroke-linecap=\"round\" stroke-linejoin=\"round\">\n")

	// Animation timing: each shape gets a start offset spread across the
	// duration and a per-shape draw time, so the piece builds up in the
	// order it was drawn.
	n := len(c.Shapes)
	draw := math.Max(0.8, dur*0.3)
	if draw > dur {
		draw = dur
	}
	spread := dur - draw
	anim := func(i int, class string, length float64) {
		start := 0.0
		if n > 1 {
			start = spread * float64(i) / float64(n-1)
		}
		sw.s(` class="`).s(class).s(`" style="--s:`).sec(start).s("s")
		if class == "d" {
			sw.s(";--t:").sec(draw).s("s;--l:").n(length)
		}
		sw.s(`"`)
	}

	for i, s := range c.Shapes {
		switch s := s.(type) {
		case Path:
			pts := s.Points
			if o.Simplify > 0 {
				pts = Simplify(pts, o.Simplify)
			}
			if len(pts) < 2 {
				continue
			}
			sw.s(`<path d="M`)
			for j, p := range pts {
				if j > 0 {
					sw.sp()
				}
				sw.n(p.X).sp().n(p.Y)
			}
			if s.Closed {
				sw.s("Z")
			}
			sw.s(`"`)
			sw.paint(s.Fill, s.Stroke, s.Width)
			if o.Animate {
				if s.Fill.IsNone() && !s.Stroke.IsNone() {
					anim(i, "d", Path{Points: pts, Closed: s.Closed}.Length()+1)
				} else {
					anim(i, "f", 0)
				}
			}
			sw.s("/>\n")
		case Circle:
			sw.s(`<circle cx="`).n(s.Center.X).s(`" cy="`).n(s.Center.Y).s(`" r="`).n(s.Radius).s(`"`)
			sw.paint(s.Fill, s.Stroke, s.Width)
			if o.Animate {
				if s.Fill.IsNone() && !s.Stroke.IsNone() {
					anim(i, "d", 2*math.Pi*s.Radius+1)
				} else {
					anim(i, "f", 0)
				}
			}
			sw.s("/>\n")
		}
	}
	sw.s("</g>\n</svg>\n")
	_, err := w.Write(sw.buf)
	return err
}

const animationCSS = `<style>
@keyframes nd{to{stroke-dashoffset:0}}
@keyframes nf{to{opacity:1}}
.d{stroke-dasharray:var(--l);stroke-dashoffset:var(--l);animation:nd var(--t) cubic-bezier(.4,.05,.25,1) var(--s) forwards}
.f{opacity:0;animation:nf .7s ease-out var(--s) forwards}
@media (prefers-reduced-motion:reduce){.d,.f{animation:none;stroke-dashoffset:0;opacity:1}}
</style>
`

type svgWriter struct {
	buf  []byte
	prec int
}

func (w *svgWriter) s(s string) *svgWriter { w.buf = append(w.buf, s...); return w }
func (w *svgWriter) sp() *svgWriter        { w.buf = append(w.buf, ' '); return w }

// n appends a number with the configured precision, trimming trailing
// zeros so "12.0" becomes "12" and "-0" becomes "0".
func (w *svgWriter) n(v float64) *svgWriter {
	w.buf = appendNum(w.buf, v, w.prec)
	return w
}

// sec appends a duration in seconds with two decimals.
func (w *svgWriter) sec(v float64) *svgWriter {
	w.buf = appendNum(w.buf, v, 2)
	return w
}

func appendNum(buf []byte, v float64, prec int) []byte {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		v = 0
	}
	start := len(buf)
	buf = strconv.AppendFloat(buf, v, 'f', prec, 64)
	if bytes.IndexByte(buf[start:], '.') >= 0 {
		end := len(buf)
		for end > start && buf[end-1] == '0' {
			end--
		}
		if end > start && buf[end-1] == '.' {
			end--
		}
		buf = buf[:end]
	}
	if string(buf[start:]) == "-0" {
		buf = append(buf[:start], '0')
	}
	return buf
}

func (w *svgWriter) paint(fill, stroke Color, width float64) {
	if !fill.IsNone() {
		w.s(` fill="`).s(fill.Hex()).s(`"`)
		if fill.A < 1 {
			w.s(` fill-opacity="`).sec(fill.A).s(`"`)
		}
	}
	if !stroke.IsNone() && width > 0 {
		w.s(` stroke="`).s(stroke.Hex()).s(`" stroke-width="`)
		w.buf = appendNum(w.buf, width, 2)
		w.s(`"`)
		if stroke.A < 1 {
			w.s(` stroke-opacity="`).sec(stroke.A).s(`"`)
		}
	}
}

func escapeXML(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}

// String implements fmt.Stringer with a short description of the canvas.
func (c *Canvas) String() string {
	return fmt.Sprintf("canvas %gx%g, %d shapes", c.Width, c.Height, len(c.Shapes))
}
