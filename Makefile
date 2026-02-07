# Pinned dependency versions
MERMAID_VERSION := 11.12.2
GITHUB_MD_CSS_VERSION := v5.9.0

MERMAID_URL := https://cdn.jsdelivr.net/npm/mermaid@$(MERMAID_VERSION)/dist/mermaid.min.js
GITHUB_MD_CSS_BASE := https://raw.githubusercontent.com/sindresorhus/github-markdown-css/$(GITHUB_MD_CSS_VERSION)

# SHA-256 checksums for integrity verification (F-08).
# Update these when upgrading dependency versions.
MERMAID_SHA256 := d0830a6c05546e9edb8fe20a8f545f3e0dc7c4c3134d584bad9c13a99d7a71e0
GITHUB_MD_LIGHT_SHA256 := de2d14b5290b8cf2af74c95e92560d9c00642ae72de0b856cece3e4eddb2d885
GITHUB_MD_DARK_SHA256 := b45ead2db01f5856c4eb378f21f47da63f6b0ecf3be5d06385472164b7283df6

# Version info injected at build time
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Docker
IMAGE   := cooked
SHA_TAG := sha-$(COMMIT)

# Tool paths (support tools installed in GOPATH/bin)
GOPATH  := $(shell go env GOPATH)
GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null || echo $(GOPATH)/bin/golangci-lint)

# Cross-compilation output directory
DIST := dist

.PHONY: deps build test test-race docker docker-amd64 docker-arm64 docker-multi clean lint lint-go help
.PHONY: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-all

## deps: Download mermaid.js and github-markdown-css into embed/ (with SHA-256 verification)
deps:
	@mkdir -p embed
	curl -fsSL -o embed/mermaid.min.js "$(MERMAID_URL)"
	curl -fsSL -o embed/github-markdown-light.css "$(GITHUB_MD_CSS_BASE)/github-markdown-light.css"
	curl -fsSL -o embed/github-markdown-dark.css "$(GITHUB_MD_CSS_BASE)/github-markdown-dark.css"
	@echo "$(MERMAID_SHA256)  embed/mermaid.min.js" | shasum -a 256 -c
	@echo "$(GITHUB_MD_LIGHT_SHA256)  embed/github-markdown-light.css" | shasum -a 256 -c
	@echo "$(GITHUB_MD_DARK_SHA256)  embed/github-markdown-dark.css" | shasum -a 256 -c
	@echo "deps: downloaded and verified mermaid@$(MERMAID_VERSION), github-markdown-css@$(GITHUB_MD_CSS_VERSION)"

## build: Build the cooked binary (native)
build:
	go build -ldflags "$(LDFLAGS)" -o cooked ./cmd/cooked

## build-linux-amd64: Cross-compile for Linux x86_64
build-linux-amd64:
	@mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/cooked-linux-amd64 ./cmd/cooked

## build-linux-arm64: Cross-compile for Linux ARM64
build-linux-arm64:
	@mkdir -p $(DIST)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/cooked-linux-arm64 ./cmd/cooked

## build-darwin-amd64: Cross-compile for macOS Intel
build-darwin-amd64:
	@mkdir -p $(DIST)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/cooked-darwin-amd64 ./cmd/cooked

## build-darwin-arm64: Cross-compile for macOS ARM64
build-darwin-arm64:
	@mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/cooked-darwin-arm64 ./cmd/cooked

## build-all: Cross-compile for all platforms
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64

## test: Run all tests
test:
	go test ./...

## test-race: Run all tests with race detector
test-race:
	go test -race ./...

## docker: Build Docker image for current architecture
docker:
	docker build --build-arg LDFLAGS="$(LDFLAGS)" \
		-t $(IMAGE):$(VERSION) \
		-t $(IMAGE):$(SHA_TAG) \
		-t $(IMAGE):latest .

## docker-amd64: Build Docker image for linux/amd64
docker-amd64:
	docker buildx build --platform linux/amd64 --build-arg LDFLAGS="$(LDFLAGS)" \
		-t $(IMAGE):$(VERSION)-amd64 --load .

## docker-arm64: Build Docker image for linux/arm64
docker-arm64:
	docker buildx build --platform linux/arm64 --build-arg LDFLAGS="$(LDFLAGS)" \
		-t $(IMAGE):$(VERSION)-arm64 --load .

## docker-multi: Build and push multi-arch Docker image (requires registry)
docker-multi:
	docker buildx build --platform linux/amd64,linux/arm64 --build-arg LDFLAGS="$(LDFLAGS)" \
		-t $(IMAGE):$(VERSION) \
		-t $(IMAGE):$(SHA_TAG) \
		-t $(IMAGE):latest --push .

## clean: Remove binary, dist/, and downloaded assets
clean:
	rm -f cooked
	rm -rf $(DIST)
	rm -f embed/mermaid.min.js
	rm -f embed/github-markdown-light.css
	rm -f embed/github-markdown-dark.css

## lint: Run golangci-lint and gitleaks
lint: lint-go
	gitleaks detect --source . --no-git -v

## lint-go: Run golangci-lint
lint-go:
	$(GOLANGCI_LINT) run --timeout=5m ./...

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
