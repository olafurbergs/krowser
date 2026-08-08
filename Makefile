BINARY := krowser
PKG := ./...
GOFLAGS ?=

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: all build install run dist test lint fmt tidy clean

all: build

build:
	CGO_ENABLED=0 go build $(GOFLAGS) -o bin/$(BINARY) ./cmd/$(BINARY)

install:
	go install $(GOFLAGS) ./cmd/$(BINARY)

run:
	go run $(GOFLAGS) ./cmd/$(BINARY)

# dist cross-compiles the binary for every platform and writes checksums.
dist:
	@rm -rf dist
	@mkdir -p dist
	@set -e; for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		name="$(BINARY)_$${os}_$${arch}$${ext}"; \
		echo "building $$name"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(GOFLAGS) -o dist/$$name ./cmd/$(BINARY); \
	done
	@cd dist && { shasum -a 256 $(BINARY)_* 2>/dev/null || sha256sum $(BINARY)_*; } > SHA256SUMS
	@echo "artifacts in dist/"

test:
	go test $(GOFLAGS) -race ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

tidy:
	go mod tidy

clean:
	rm -rf bin dist
