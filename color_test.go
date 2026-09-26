package nacre

import "testing"

func TestParseHex(t *testing.T) {
	cases := map[string]string{"#fff": "#ffffff", "1a2b3c": "#1a2b3c", "#1A2B3C80": "#1a2b3c"}
	for in, want := range cases {
		c, err := ParseHex(in)
		if err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if got := c.Hex(); got != want {
			t.Fatalf("%s: got %s want %s", in, got, want)
		}
	}
	if c, _ := ParseHex("#1A2B3C80"); c.A < 0.5 || c.A > 0.51 {
		t.Fatalf("alpha = %v", c.A)
	}
	for _, bad := range []string{"", "#12", "#ggg", "#12345"} {
		if _, err := ParseHex(bad); err == nil {
			t.Fatalf("%q should fail", bad)
		}
	}
}

func TestMixAndPalette(t *testing.T) {
	black, white := Hex("#000"), Hex("#fff")
	if got := black.Mix(white, 0.5).Hex(); got != "#808080" {
		t.Fatalf("mix = %s", got)
	}
	p := Palette{Colors: []Color{black, white}}
	if p.At(0) != black || p.At(1) != white {
		t.Fatal("palette endpoints")
	}
	if p.At(2) != white || p.At(-1) != black {
		t.Fatal("palette clamps")
	}
	if !Hex("#000").IsDark() || Hex("#fff").IsDark() {
		t.Fatal("IsDark")
	}
	for _, name := range PaletteNames() {
		pal, ok := LookupPalette(name)
		if !ok || pal.Name != name || len(pal.Colors) == 0 {
			t.Fatalf("palette %q malformed", name)
		}
	}
}
