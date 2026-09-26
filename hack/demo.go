//go:build ignore

// demo.go writes docs/demo.svg: an animated terminal session showing
// the CLI, built from real command output so it never goes stale.
//
//	go run ./hack/demo.go > docs/demo.svg
package main

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type line struct {
	text   string
	prompt bool
}

func run(args ...string) []string {
	cmd := exec.Command("go", append([]string{"run", "./cmd/nacre"}, args...)...)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	var lines []string
	timing := regexp.MustCompile(`, \d+m?s\)`)
	for _, l := range strings.Split(strings.TrimRight(out.String(), "\n"), "\n") {
		lines = append(lines, timing.ReplaceAllString(l, ")"))
	}
	return lines
}

func main() {
	tmp, _ := os.MkdirTemp("", "nacre-demo")
	defer os.RemoveAll(tmp)
	var script []line
	add := func(cmd string, args ...string) {
		script = append(script, line{"nacre " + cmd, true})
		for _, l := range run(args...) {
			script = append(script, line{strings.ReplaceAll(l, tmp+"/", ""), false})
		}
	}
	add("styles", "styles")
	add("render -style flow -seed 0x3f1a -palette dusk -o flow.png",
		"render", "-style", "flow", "-seed", "0x3f1a", "-palette", "dusk", "-o", tmp+"/flow.png")
	add(`render -seed "sea glass" -animate`,
		"render", "-seed", "sea glass", "-animate", "-o", tmp+"/x.svg")
	add("gallery -seed 0x5eed -out gallery", "gallery", "-seed", "0x5eed", "-out", tmp+"/gallery")

	// Fix the file name printed by the "sea glass" render.
	for i := range script {
		script[i].text = regexp.MustCompile(`→ x\.svg`).ReplaceAllString(script[i].text, "→ "+pieceName(script[i].text))
		script[i].text = strings.ReplaceAll(script[i].text, "gallery from seed 0x5eed → "+tmp+"/gallery", "gallery from seed 0x5eed → gallery")
	}

	const (
		w, pad, lh = 860, 22, 21
		top        = 44
		charW      = 7.8
		typeSpeed  = 0.035 // seconds per character
		pause      = 0.5
	)
	h := top + pad + lh*len(script) + pad/2
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="ui-monospace,SFMono-Regular,Menlo,Consolas,monospace" font-size="13">`+"\n", w, h, w, h)
	b.WriteString("<title>nacre command line demo</title>\n<style>\n")
	b.WriteString(".l{opacity:0;animation:show 0s linear forwards}\n")
	b.WriteString(".t{clip-path:inset(0 100% 0 0);animation:type steps(var(--n)) forwards}\n")
	b.WriteString("@keyframes show{to{opacity:1}}\n@keyframes type{to{clip-path:inset(0 0 0 0)}}\n")
	b.WriteString("@media (prefers-reduced-motion:reduce){.l,.t{animation:none;opacity:1;clip-path:none}}\n")
	b.WriteString("</style>\n")
	fmt.Fprintf(&b, `<rect width="%d" height="%d" rx="10" fill="#1b1a24"/>`+"\n", w, h)
	for i, c := range []string{"#ff5f57", "#febc2e", "#28c840"} {
		fmt.Fprintf(&b, `<circle cx="%d" cy="20" r="6" fill="%s"/>`+"\n", 22+i*20, c)
	}
	fmt.Fprintf(&b, `<text x="%d" y="24" fill="#8a86a0" text-anchor="middle" font-size="12">nacre</text>`+"\n", w/2)

	t := 0.4
	for i, l := range script {
		y := top + pad + i*lh
		text := html.EscapeString(l.text)
		if l.prompt {
			n := len([]rune(l.text))
			d := float64(n) * typeSpeed
			fmt.Fprintf(&b, `<g class="l" style="animation-delay:%.2fs"><text x="%d" y="%d" fill="#b7a6e0">$</text>`, t, pad, y)
			fmt.Fprintf(&b, `<text class="t" x="%d" y="%d" fill="#ecebf1" style="--n:%d;animation-duration:%.2fs;animation-delay:%.2fs">%s</text></g>`+"\n",
				pad+16, y, n, d, t, text)
			t += d + pause
			continue
		}
		fill := "#a9a4b8"
		if strings.HasPrefix(l.text, "✦") {
			fill = "#ecebf1"
			text = `<tspan fill="#f0a3c0">✦</tspan>` + strings.TrimPrefix(text, "✦")
		}
		fmt.Fprintf(&b, `<text class="l" x="%d" y="%d" fill="%s" xml:space="preserve" style="animation-delay:%.2fs">%s</text>`+"\n", pad+16, y, fill, t, text)
		t += 0.12
	}
	b.WriteString("</svg>\n")
	os.Stdout.WriteString(b.String())
}

func pieceName(s string) string {
	f := strings.Fields(s)
	if len(f) < 3 {
		return "x.svg"
	}
	return f[1] + "-" + f[2] + ".svg"
}
