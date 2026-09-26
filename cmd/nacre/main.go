// Command nacre renders generative art from a seed.
//
//	nacre render                       a surprise, saved as <style>-<seed>.svg
//	nacre render -style flow -seed 0x3f1a -palette dusk -o flow.png
//	nacre gallery -out docs/gallery    one piece per style
//	nacre styles                       list styles
//	nacre palettes                     list palettes with swatches
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/devatobuild/nacre"
	_ "github.com/devatobuild/nacre/styles"
)

var version = "dev"

const usageText = `nacre renders generative art from a seed.

Usage:
  nacre render   [flags]    draw one piece as SVG or PNG
  nacre gallery  [flags]    draw one piece per style into a folder
  nacre styles              list the styles
  nacre palettes            list the palettes
  nacre version             print the version

Run "nacre render -h" or "nacre gallery -h" for flags.
A seed can be hex (0x3f1a), decimal (16154) or any words ("sea glass").
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "render":
		err = cmdRender(os.Args[2:])
	case "gallery":
		err = cmdGallery(os.Args[2:])
	case "styles":
		cmdStyles()
	case "palettes":
		cmdPalettes()
	case "version", "-v", "--version":
		fmt.Println("nacre", version)
	case "help", "-h", "--help":
		fmt.Print(usageText)
	default:
		fmt.Fprintf(os.Stderr, "nacre: unknown command %q\n\n%s", os.Args[1], usageText)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "nacre:", strings.TrimPrefix(err.Error(), "nacre: "))
		os.Exit(1)
	}
}

// renderFlags are shared by render and gallery.
type renderFlags struct {
	style, palette string
	size, w, h     float64
	density        float64
	animate        bool
	duration       float64
	scale          float64
	simplify       float64
	precision      int
	quiet          bool
}

func (f *renderFlags) bind(fs *flag.FlagSet) {
	fs.StringVar(&f.style, "style", "", "style name, or empty to let the seed choose")
	fs.StringVar(&f.palette, "palette", "", "palette name, or empty to let the seed choose")
	fs.Float64Var(&f.size, "size", 1200, "canvas size in units (square)")
	fs.Float64Var(&f.w, "width", 0, "canvas width, overrides -size")
	fs.Float64Var(&f.h, "height", 0, "canvas height, overrides -size")
	fs.Float64Var(&f.density, "density", 1, "how much the style draws; 0.5 is sparse, 2 is dense")
	fs.BoolVar(&f.animate, "animate", false, "SVG only: strokes draw themselves in when opened")
	fs.Float64Var(&f.duration, "duration", 6, "SVG only: animation length in seconds")
	fs.Float64Var(&f.scale, "scale", 1, "PNG only: pixels per canvas unit")
	fs.Float64Var(&f.simplify, "simplify", 0.35, "SVG only: drop points closer than this to a straight line (0 keeps all)")
	fs.IntVar(&f.precision, "precision", 1, "SVG only: decimal places for coordinates")
	fs.BoolVar(&f.quiet, "quiet", false, "print nothing but errors")
}

func (f *renderFlags) options(seed uint64) nacre.Options {
	w, h := f.size, f.size
	if f.w > 0 {
		w = f.w
	}
	if f.h > 0 {
		h = f.h
	}
	return nacre.Options{Style: f.style, Seed: seed, Width: w, Height: h, Palette: f.palette, Density: f.density}
}

func (f *renderFlags) svgOptions() nacre.SVGOptions {
	return nacre.SVGOptions{Precision: f.precision, Simplify: f.simplify, Animate: f.animate, Duration: f.duration}
}

func cmdRender(args []string) error {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	var f renderFlags
	f.bind(fs)
	seedText := fs.String("seed", "", "seed: hex (0x3f1a), decimal, or any words; empty picks one at random")
	out := fs.String("o", "", "output file; .svg or .png decides the format; \"-\" writes SVG to stdout")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: nacre render [flags]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	seed, err := seedFrom(*seedText)
	if err != nil {
		return err
	}
	start := time.Now()
	piece, err := nacre.Render(f.options(seed))
	if err != nil {
		return err
	}
	path := *out
	if path == "" {
		path = piece.Filename(".svg")
	}
	size, err := writePiece(piece, path, &f)
	if err != nil {
		return err
	}
	if !f.quiet {
		report(piece, path, size, time.Since(start))
	}
	return nil
}

func cmdGallery(args []string) error {
	fs := flag.NewFlagSet("gallery", flag.ExitOnError)
	var f renderFlags
	f.bind(fs)
	seedText := fs.String("seed", "", "base seed the gallery's seeds are derived from; empty picks one at random")
	out := fs.String("out", "gallery", "output folder")
	per := fs.Int("per", 1, "pieces per style")
	format := fs.String("format", "svg", "svg or png")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: nacre gallery [flags]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	base, err := seedFrom(*seedText)
	if err != nil {
		return err
	}
	if *format != "svg" && *format != "png" {
		return fmt.Errorf("format must be svg or png, not %q", *format)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	styles := nacre.StyleNames()
	if f.style != "" {
		styles = []string{f.style}
	}
	seeds := nacre.NewRand(base)
	if !f.quiet {
		fmt.Printf("%s gallery from seed %s → %s\n", accent("✦"), nacre.FormatSeed(base), *out)
	}
	for _, name := range styles {
		for k := 0; k < *per; k++ {
			start := time.Now()
			o := f.options(seeds.Uint64())
			o.Style = name
			piece, err := nacre.Render(o)
			if err != nil {
				return err
			}
			path := filepath.Join(*out, piece.Filename("."+*format))
			size, err := writePiece(piece, path, &f)
			if err != nil {
				return err
			}
			if !f.quiet {
				report(piece, path, size, time.Since(start))
			}
		}
	}
	return nil
}

func cmdStyles() {
	w := 0
	for _, s := range nacre.Styles() {
		w = max(w, len(s.Name))
	}
	for _, s := range nacre.Styles() {
		fmt.Printf("  %s%s  %s\n", bold(s.Name), strings.Repeat(" ", w-len(s.Name)), dim(s.Info))
	}
}

func cmdPalettes() {
	w := 0
	for _, p := range nacre.Palettes() {
		w = max(w, len(p.Name))
	}
	for _, p := range nacre.Palettes() {
		fmt.Printf("  %s%s  %s", bold(p.Name), strings.Repeat(" ", w-len(p.Name)), swatch(p.Background))
		for _, c := range p.Colors {
			fmt.Print(swatch(c))
		}
		fmt.Printf("  %s", dim(p.Background.Hex()))
		for _, c := range p.Colors {
			fmt.Printf(" %s", dim(c.Hex()))
		}
		fmt.Println()
	}
}

func seedFrom(text string) (uint64, error) {
	if text == "" {
		return nacre.RandomSeed(), nil
	}
	return nacre.ParseSeed(text)
}

// writePiece saves the piece as SVG or PNG depending on the extension
// and returns the number of bytes written.
func writePiece(piece *nacre.Piece, path string, f *renderFlags) (int64, error) {
	if path == "-" {
		b := piece.SVG(f.svgOptions())
		_, err := os.Stdout.Write(b)
		return int64(len(b)), err
	}
	file, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	bw := bufio.NewWriterSize(file, 1<<20)
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		err = piece.Canvas.WritePNG(bw, f.scale)
	case ".svg", "":
		err = piece.Canvas.WriteSVG(bw, f.svgOptions())
	default:
		return 0, fmt.Errorf("output must end in .svg or .png, not %q", filepath.Ext(path))
	}
	if err != nil {
		return 0, err
	}
	if err := bw.Flush(); err != nil {
		return 0, err
	}
	info, err := file.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func report(p *nacre.Piece, path string, size int64, took time.Duration) {
	fmt.Printf("%s %s  %s  %s  %s shapes  → %s %s\n",
		accent("✦"), bold(p.Style.Name), nacre.FormatSeed(p.Seed), p.Palette.Name,
		group(len(p.Canvas.Shapes)), path, dim(fmt.Sprintf("(%s, %s)", human(size), took.Round(time.Millisecond))))
}

func human(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.0f KB", float64(n)/(1<<10))
	}
	return fmt.Sprintf("%d B", n)
}

func group(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

// Terminal colour, only when stdout is a terminal and NO_COLOR is unset.
var colour = func() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	info, err := os.Stdout.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}()

func wrap(code, s string) string {
	if !colour {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func bold(s string) string   { return wrap("1", s) }
func dim(s string) string    { return wrap("2", s) }
func accent(s string) string { return wrap("35", s) }

func swatch(c nacre.Color) string {
	if !colour {
		return "■"
	}
	r, g, b, _ := c.RGBA8()
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm  \x1b[0m", r, g, b)
}
