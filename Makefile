VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS  = -ldflags "-s -w -X github.com/sota-io/sota-cli/internal/cmd.version=$(VERSION) -X github.com/sota-io/sota-cli/internal/cmd.commit=$(COMMIT)"

.PHONY: build build-all install clean

build:
	go build $(LDFLAGS) -o bin/sota ./cmd/sota/

install:
	go build $(LDFLAGS) -o $(GOPATH)/bin/sota ./cmd/sota/

build-all:
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o bin/sota-linux-amd64   ./cmd/sota/
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o bin/sota-linux-arm64   ./cmd/sota/
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o bin/sota-darwin-amd64  ./cmd/sota/
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o bin/sota-darwin-arm64  ./cmd/sota/

clean:
	rm -rf bin/
