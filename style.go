package nacre

import (
	"fmt"
	"sort"
)

// Params carries what a style needs beyond the canvas and its random
// stream.
type Params struct {
	// Palette is the colour set to paint with.
	Palette Palette
	// Density scales how much a style draws (strokes, tiles, circles);
	// 1 is the style's own default, 0.5 is sparse, 2 is dense.
	Density float64
}

// DrawFunc draws one piece onto c. It must take all randomness from r
// so that the same seed always produces the same piece.
type DrawFunc func(c *Canvas, r *Rand, p Params)

// Style is a named way of drawing.
type Style struct {
	Name string
	// Info is a one-line description shown by the CLI and playground.
	Info string
	Draw DrawFunc
}

var styleRegistry = map[string]Style{}

// Register adds a style so it can be rendered by name. The built-in
// styles register themselves when the styles package is imported.
func Register(s Style) {
	if s.Name == "" || s.Draw == nil {
		panic("nacre: Register needs a name and a Draw function")
	}
	styleRegistry[s.Name] = s
}

// Styles returns every registered style, sorted by name.
func Styles() []Style {
	names := StyleNames()
	out := make([]Style, len(names))
	for i, n := range names {
		out[i] = styleRegistry[n]
	}
	return out
}

// StyleNames returns the names of every registered style, sorted.
func StyleNames() []string {
	names := make([]string, 0, len(styleRegistry))
	for n := range styleRegistry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// LookupStyle finds a registered style by name.
func LookupStyle(name string) (Style, bool) {
	s, ok := styleRegistry[name]
	return s, ok
}

func (s Style) String() string { return fmt.Sprintf("%s: %s", s.Name, s.Info) }
