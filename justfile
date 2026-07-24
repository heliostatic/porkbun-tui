# Porkbun TUI - justfile

# Default recipe: show available commands
default:
    @just --list

# One-time setup: verify Go, fetch deps, install dev tools, prove the build
bootstrap:
    #!/usr/bin/env bash
    set -euo pipefail
    command -v go >/dev/null 2>&1 || { echo "Error: Go 1.21+ is required — https://go.dev/dl/"; exit 1; }
    echo "==> Downloading module dependencies"
    go mod download
    echo "==> Installing gofumpt (used by 'just fmt')"
    go install mvdan.cc/gofumpt@latest
    if ! command -v golangci-lint >/dev/null 2>&1; then
        if command -v brew >/dev/null 2>&1; then
            echo "==> Installing golangci-lint (used by 'just lint')"
            brew install golangci-lint
        else
            echo "Note: golangci-lint not found — install it for 'just lint': https://golangci-lint.run/usage/install/"
        fi
    fi
    if ! command -v vhs >/dev/null 2>&1; then
        echo "Note: vhs not installed (only needed for 'just screenshots') — brew install vhs"
    fi
    echo "==> Verifying build and tests"
    go build ./...
    go test ./... >/dev/null
    echo ""
    echo "Ready. Set PORKBUN_API_KEY / PORKBUN_SECRET_KEY (see 'just check-creds'),"
    echo "then run 'just run' — or 'just demo' to try it without credentials."

# Build the binary with version from git
build:
    go build -ldflags="-X main.version=0.0.1-$(git rev-parse --short=8 HEAD 2>/dev/null || echo unknown)" -o porkbun-tui ./cmd/porkbun-tui

# Build release with custom version
build-release version:
    go build -ldflags="-s -w -X main.version={{version}}" -o porkbun-tui ./cmd/porkbun-tui

# Run the TUI
run: build
    ./porkbun-tui

# Run without building (go run)
dev:
    go run ./cmd/porkbun-tui

# Run in demo mode (no credentials needed)
demo:
    go run ./cmd/porkbun-tui --demo

# Clean build artifacts
clean:
    rm -f porkbun-tui
    go clean

# Run tests
test:
    go test -v ./...

# Run tests with the race detector, uncached
test-race:
    go test -race -count=1 ./...

# Vet for suspicious constructs
vet:
    go vet ./...

# CI-style gate: formatting, vet, race tests
check:
    #!/usr/bin/env bash
    set -euo pipefail
    unformatted=$(gofmt -l .)
    if [ -n "$unformatted" ]; then
        echo "gofmt needed on:"; echo "$unformatted"; exit 1
    fi
    go vet ./...
    go test -race -count=1 ./...
    echo "check passed"

# Run tests with coverage
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report: coverage.html"

# Format code
fmt:
    go fmt ./...
    gofumpt -w . 2>/dev/null || true

# Lint code
lint:
    golangci-lint run ./...

# Tidy dependencies
tidy:
    go mod tidy

# Update dependencies
update:
    go get -u ./...
    go mod tidy

# Install binary to GOPATH/bin
install:
    go install ./cmd/porkbun-tui

# Uninstall binary from GOPATH/bin
uninstall:
    rm -f $(go env GOPATH)/bin/porkbun-tui

# Check if credentials are set
check-creds:
    @if [ -z "${PORKBUN_API_KEY:-}" ]; then echo "PORKBUN_API_KEY not set"; exit 1; fi
    @if [ -z "${PORKBUN_SECRET_KEY:-}" ]; then echo "PORKBUN_SECRET_KEY not set"; exit 1; fi
    @echo "Credentials are set"

# Build for multiple platforms
build-all:
    #!/usr/bin/env bash
    VERSION="0.0.1-$(git rev-parse --short=8 HEAD 2>/dev/null || echo 'unknown')"
    LDFLAGS="-s -w -X main.version=$VERSION"
    mkdir -p dist
    GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o dist/porkbun-tui-darwin-amd64 ./cmd/porkbun-tui
    GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o dist/porkbun-tui-darwin-arm64 ./cmd/porkbun-tui
    GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o dist/porkbun-tui-linux-amd64 ./cmd/porkbun-tui
    GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o dist/porkbun-tui-linux-arm64 ./cmd/porkbun-tui
    GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o dist/porkbun-tui-windows-amd64.exe ./cmd/porkbun-tui
    echo "Binaries built in dist/ with version $VERSION"

# Clean dist folder
clean-dist:
    rm -rf dist/

# Full clean
clean-all: clean clean-dist
    rm -f coverage.out coverage.html

# Generate screenshots using VHS (requires: brew install vhs)
screenshots: build
    vhs screenshots/demo.tape
