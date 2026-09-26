package nacre_test

import (
	"fmt"
	"math"

	"github.com/devatobuild/nacre"
)

// A style is just a function. Register it and it can be rendered by
// name, from Go, the CLI or the playground.
func Example_customStyle() {
	nacre.Register(nacre.Style{
		Name: "orbits",
		Info: "Concentric rings with a little wobble.",
		Draw: func(c *nacre.Canvas, r *nacre.Rand, p nacre.Params) {
			u := c.Unit()
			for i := 0; i < int(30*p.Density); i++ {
				pts := make([]nacre.Point, 0, 120)
				radius := u * (0.05 + 0.4*float64(i)/30)
				wobble := u * r.Range(0, 0.01)
				for k := 0; k < 120; k++ {
					a := 2 * math.Pi * float64(k) / 119
					pts = append(pts, c.Center().Polar(radius+wobble*math.Sin(6*a), a))
				}
				c.Stroke(pts, p.Palette.At(float64(i)/30), u*0.002)
			}
		},
	})
	piece, err := nacre.Render(nacre.Options{Style: "orbits", Seed: 7, Width: 400, Palette: "dusk"})
	if err != nil {
		panic(err)
	}
	fmt.Println(piece.Style.Name, len(piece.Canvas.Shapes), "rings")
	// Output: orbits 30 rings
}
