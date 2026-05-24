.PHONY: build check clean install test vet wallpaper

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf 1.0.0)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.sourceRoot=$(CURDIR)
GO_CACHE ?= $(shell if [ -n "$$GOCACHE" ]; then printf '%s' "$$GOCACHE"; elif mkdir -p "$${XDG_CACHE_HOME:-$$HOME/.cache}/go-build" 2>/dev/null && [ -w "$${XDG_CACHE_HOME:-$$HOME/.cache}/go-build" ]; then printf '%s' "$${XDG_CACHE_HOME:-$$HOME/.cache}/go-build"; else printf '%s' "$${TMPDIR:-/tmp}/hyprglass-go-cache"; fi)

build:
	mkdir -p build
	mkdir -p "$(GO_CACHE)"
	env GOCACHE="$(GO_CACHE)" go build -buildvcs=false -ldflags "$(LDFLAGS)" -o build/hyprglass ./cmd/hyprglass

check:
	./scripts/check.sh

clean:
	rm -rf build

install:
	./install.sh --yes

test:
	mkdir -p "$(GO_CACHE)"
	env GOCACHE="$(GO_CACHE)" go test ./...

vet:
	mkdir -p "$(GO_CACHE)"
	env GOCACHE="$(GO_CACHE)" go vet ./...

wallpaper:
	python3 ./scripts/generate-wallpaper.py
