package nacre

import (
	"math"
	"sort"
)

// Palette is a curated set of colours a style paints with.
type Palette struct {
	Name string
	// Background is the paper the piece is drawn on.
	Background Color
	// Ink is a high-contrast foreground for line work, chosen to read
	// clearly against Background.
	Ink Color
	// Colors are the palette's hues, ordered so that adjacent entries
	// sit well together; At walks through them as a gradient.
	Colors []Color
}

// At returns a colour along the palette's gradient for t in [0, 1],
// blending smoothly between neighbouring entries.
func (p Palette) At(t float64) Color {
	n := len(p.Colors)
	if n == 0 {
		return p.Ink
	}
	if n == 1 {
		return p.Colors[0]
	}
	t = clamp01(t) * float64(n-1)
	i := int(math.Floor(t))
	if i >= n-1 {
		return p.Colors[n-1]
	}
	return p.Colors[i].Mix(p.Colors[i+1], t-float64(i))
}

// Pick returns a random colour from the palette.
func (p Palette) Pick(r *Rand) Color { return Pick(r, p.Colors) }

// IsDark reports whether the palette's background is dark.
func (p Palette) IsDark() bool { return p.Background.IsDark() }

func hexes(ss ...string) []Color {
	cs := make([]Color, len(ss))
	for i, s := range ss {
		cs[i] = Hex(s)
	}
	return cs
}

var palettes = map[string]Palette{
	"nacre": {
		Name:       "nacre",
		Background: Hex("#f5f1ea"),
		Ink:        Hex("#2a2733"),
		Colors:     hexes("#9f8fc0", "#7fa6c4", "#e39a93", "#c6b88a", "#79a397", "#cf8fb2"),
	},
	"dusk": {
		Name:       "dusk",
		Background: Hex("#1b1a2e"),
		Ink:        Hex("#f4e9d4"),
		Colors:     hexes("#f6a55e", "#e86a58", "#b84f8a", "#6d4a9d", "#3b5ba9"),
	},
	"ember": {
		Name:       "ember",
		Background: Hex("#0f0d0c"),
		Ink:        Hex("#f2e0bf"),
		Colors:     hexes("#8c2f1a", "#d94f30", "#ff5a1f", "#ff8c42", "#ffb347"),
	},
	"moss": {
		Name:       "moss",
		Background: Hex("#f0eee3"),
		Ink:        Hex("#1f2a22"),
		Colors:     hexes("#2c4a3a", "#3b6e4a", "#5d8c5a", "#8fb07a", "#c8c47e"),
	},
	"ink": {
		Name:       "ink",
		Background: Hex("#f7f4ee"),
		Ink:        Hex("#141414"),
		Colors:     hexes("#141414", "#2e2e2e", "#4d4d4d", "#7a7a7a", "#a8a8a8"),
	},
	"sakura": {
		Name:       "sakura",
		Background: Hex("#fff5f6"),
		Ink:        Hex("#3a2a34"),
		Colors:     hexes("#f7b5c9", "#f28cae", "#e26a95", "#a05a7a", "#6f8fa3"),
	},
	"tundra": {
		Name:       "tundra",
		Background: Hex("#e8edf1"),
		Ink:        Hex("#20303c"),
		Colors:     hexes("#3a4f63", "#5f7d95", "#90a9bd", "#bccbd7", "#d9a066"),
	},
	"solar": {
		Name:       "solar",
		Background: Hex("#fdf6e3"),
		Ink:        Hex("#073642"),
		Colors:     hexes("#b58900", "#cb4b16", "#dc322f", "#6c71c4", "#268bd2", "#2aa198", "#859900"),
	},
	"noir": {
		Name:       "noir",
		Background: Hex("#0b0b0e"),
		Ink:        Hex("#f1f1f5"),
		Colors:     hexes("#f1f1f5", "#c3c3cc", "#8b8b97", "#5b5b67", "#37373f"),
	},
	"reef": {
		Name:       "reef",
		Background: Hex("#0e2a47"),
		Ink:        Hex("#eaf6ff"),
		Colors:     hexes("#1ec8c8", "#4fd1a1", "#f7d060", "#ff7b54", "#ff4f81"),
	},
}

// Palettes returns every built-in palette, sorted by name.
func Palettes() []Palette {
	names := PaletteNames()
	out := make([]Palette, len(names))
	for i, n := range names {
		out[i] = palettes[n]
	}
	return out
}

// PaletteNames returns the names of every built-in palette, sorted.
func PaletteNames() []string {
	names := make([]string, 0, len(palettes))
	for n := range palettes {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// LookupPalette finds a palette by name.
func LookupPalette(name string) (Palette, bool) {
	p, ok := palettes[name]
	return p, ok
}

// RegisterPalette adds (or replaces) a palette so that styles, the CLI
// and the playground can use it by name.
func RegisterPalette(p Palette) {
	if p.Name == "" {
		panic("nacre: RegisterPalette needs a name")
	}
	palettes[p.Name] = p
}
