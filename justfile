# Porkbun TUI - justfile

# Default recipe: show available commands
default:
    @just --list

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

# Clean build artifacts
clean:
    rm -f porkbun-tui
    go clean

# Run tests
test:
    go test -v ./...

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
