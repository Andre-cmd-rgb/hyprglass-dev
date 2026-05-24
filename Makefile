.PHONY: build check test wallpaper
build:
	go build -o build/hyprglass ./cmd/hyprglass
check:
	./scripts/check.sh
test:
	go test ./...
wallpaper:
	./scripts/generate-wallpaper.py
