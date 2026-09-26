VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  = -s -w -X main.version=$(VERSION)
GOROOT  := $(shell go env GOROOT)

.PHONY: build test lint bench wasm serve gallery hero demo clean

build:            ## build the CLI into bin/nacre
	go build -ldflags "$(LDFLAGS)" -o bin/nacre ./cmd/nacre

test:             ## run the tests
	go test ./...

lint:             ## gofmt, go vet and staticcheck
	@test -z "$$(gofmt -l .)" || { gofmt -l .; echo "gofmt: files need formatting"; exit 1; }
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

bench:            ## run the benchmarks
	go test ./styles -bench . -benchmem -run ^$$

wasm:             ## build the browser playground into dist/
	rm -rf dist && mkdir -p dist
	GOOS=js GOARCH=wasm go build -ldflags "$(LDFLAGS)" -o dist/nacre.wasm ./cmd/nacre-wasm
	cp "$(GOROOT)/lib/wasm/wasm_exec.js" dist/
	cp web/* dist/

serve: wasm       ## serve the playground on http://localhost:8080
	python3 -m http.server 8080 -d dist

gallery: build    ## re-render the README gallery
	./bin/nacre gallery -out docs/gallery -seed nacre -size 900 -quiet
	@ls docs/gallery

hero: build       ## re-render the animated README hero
	./bin/nacre render -style flow -seed 0x3f1a -palette dusk -width 1600 -height 640 -density 0.45 -simplify 0.8 -animate -duration 7 -o docs/hero.svg

demo:             ## rebuild the terminal demo animation
	go run ./hack/demo.go > docs/demo.svg

clean:
	rm -rf bin dist out gallery
