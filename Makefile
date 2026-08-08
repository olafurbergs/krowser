BINARY := krowser
PKG := ./...
GOFLAGS ?=

.PHONY: all build install run test lint fmt tidy clean

all: build

build:
	CGO_ENABLED=0 go build $(GOFLAGS) -o bin/$(BINARY) ./cmd/$(BINARY)

install:
	go install $(GOFLAGS) ./cmd/$(BINARY)

run:
	go run $(GOFLAGS) ./cmd/$(BINARY)

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
