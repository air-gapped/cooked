# Pinned dependency versions
MERMAID_VERSION := 11.12.2
GITHUB_MD_CSS_VERSION := v5.9.0

MERMAID_URL := https://cdn.jsdelivr.net/npm/mermaid@$(MERMAID_VERSION)/dist/mermaid.min.js
GITHUB_MD_CSS_BASE := https://raw.githubusercontent.com/sindresorhus/github-markdown-css/$(GITHUB_MD_CSS_VERSION)

# Version info injected at build time
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Docker
IMAGE   := cooked
SHA_TAG := sha-$(COMMIT)

# Cross-compilation output directory
DIST := dist

.PHONY: deps build test test-race docker docker-amd64 docker-arm64 docker-multi clean lint help
.PHONY: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-all

## deps: Download mermaid.js and github-markdown-css into embed/
deps:
	@mkdir -p embed
	curl -fsSL -o embed/mermaid.min.js "$(MERMAID_URL)"
	curl -fsSL -o embed/github-markdown-light.css "$(GITHUB_MD_CSS_BASE)/github-markdown-light.css"
	curl -fsSL -o embed/github-markdown-dark.css "$(GITHUB_MD_CSS_BASE)/github-markdown-dark.css"
	@echo "deps: downloaded mermaid@$(MERMAID_VERSION), github-markdown-css@$(GITHUB_MD_CSS_VERSION)"

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

## lint: Run gitleaks
lint:
	gitleaks detect --source . --no-git -v

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
