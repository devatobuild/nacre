//go:build js && wasm

// Command nacre-wasm exposes the engine to the browser playground as
// window.nacre with render, styles, palettes and seed functions.
package main

import (
	"syscall/js"
	"time"

	"github.com/devatobuild/nacre"
	_ "github.com/devatobuild/nacre/styles"
)

func main() {
	js.Global().Set("nacre", js.ValueOf(map[string]any{
		"render":   js.FuncOf(render),
		"styles":   js.FuncOf(styles),
		"palettes": js.FuncOf(palettes),
		"seed":     js.FuncOf(func(js.Value, []js.Value) any { return nacre.FormatSeed(nacre.RandomSeed()) }),
		"parseSeed": js.FuncOf(func(_ js.Value, args []js.Value) any {
			if len(args) == 0 {
				return nil
			}
			s, err := nacre.ParseSeed(args[0].String())
			if err != nil {
				return nil
			}
			return nacre.FormatSeed(s)
		}),
	}))
	js.Global().Call("dispatchEvent", js.Global().Get("Event").New("nacre-ready"))
	select {}
}

// render({style, seed, palette, size, density, animate}) → {svg, style, palette, seed, shapes, ms} or {error}.
func render(_ js.Value, args []js.Value) any {
	opts := nacre.Options{}
	svgo := nacre.SVGOptions{Simplify: 0.4}
	if len(args) > 0 && args[0].Type() == js.TypeObject {
		a := args[0]
		if v := a.Get("style"); v.Type() == js.TypeString {
			opts.Style = v.String()
		}
		if v := a.Get("palette"); v.Type() == js.TypeString {
			opts.Palette = v.String()
		}
		if v := a.Get("size"); v.Type() == js.TypeNumber {
			opts.Width = v.Float()
		}
		if v := a.Get("density"); v.Type() == js.TypeNumber {
			opts.Density = v.Float()
		}
		if v := a.Get("animate"); v.Type() == js.TypeBoolean {
			svgo.Animate = v.Bool()
		}
		if v := a.Get("seed"); v.Type() == js.TypeString && v.String() != "" {
			s, err := nacre.ParseSeed(v.String())
			if err != nil {
				return map[string]any{"error": err.Error()}
			}
			opts.Seed = s
		} else {
			opts.Seed = nacre.RandomSeed()
		}
	}
	start := time.Now()
	piece, err := nacre.Render(opts)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	svg := piece.SVG(svgo)
	return map[string]any{
		"svg":     string(svg),
		"style":   piece.Style.Name,
		"palette": piece.Palette.Name,
		"seed":    nacre.FormatSeed(piece.Seed),
		"shapes":  len(piece.Canvas.Shapes),
		"ms":      time.Since(start).Milliseconds(),
	}
}

func styles(js.Value, []js.Value) any {
	out := make([]any, 0)
	for _, s := range nacre.Styles() {
		out = append(out, map[string]any{"name": s.Name, "info": s.Info})
	}
	return out
}

func palettes(js.Value, []js.Value) any {
	out := make([]any, 0)
	for _, p := range nacre.Palettes() {
		cols := make([]any, len(p.Colors))
		for i, c := range p.Colors {
			cols[i] = c.Hex()
		}
		out = append(out, map[string]any{
			"name": p.Name, "background": p.Background.Hex(), "ink": p.Ink.Hex(), "colors": cols,
		})
	}
	return out
}
