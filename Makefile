GO ?= go

# Get git information for versioning
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo "unknown")
GIT_DIRTY := $(shell test -n "$$(git status --porcelain 2>/dev/null)" && echo "true" || echo "false")

# Set ldflags for version information
LDFLAGS = -s -w \
	-X 'git-sync/internal/version.Version=$(GIT_TAG)' \
	-X 'git-sync/internal/version.Commit=$(GIT_COMMIT)' \
	-X 'git-sync/internal/version.Date=$(BUILD_DATE)' \
	-X 'git-sync/internal/version.Dirty=$(GIT_DIRTY)'

# Docker image name and tag
IMAGE_NAME ?= git-sync
IMAGE_TAG ?= $(GIT_TAG)

# Default target
.PHONY: all
all: build

# Build the application using Docker (default)
.PHONY: build
build:
	docker build \
		--build-arg VERSION=$(GIT_TAG) \
		--build-arg COMMIT=$(GIT_COMMIT) \
		--build-arg DATE=$(BUILD_DATE) \
		--build-arg DIRTY=$(GIT_DIRTY) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) .

# Build the application locally (for development)
.PHONY: build-local
build-local:
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o bin/git-sync ./cmd

# Build for Windows with .exe extension
.PHONY: build-windows
build-windows:
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o bin/git-sync.exe ./cmd

# Build for multiple platforms
.PHONY: build-all
build-all: build-local build-windows
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o bin/git-sync-linux ./cmd
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o bin/git-sync-darwin ./cmd

# Install the application
.PHONY: install
install:
	$(GO) install -trimpath -ldflags "$(LDFLAGS)" ./cmd

# Run tests
.PHONY: test
test:
	$(GO) test ./...

# Run tests with coverage
.PHONY: cover
cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf bin/
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true

# Create a git tag for release
.PHONY: tag
tag:            ## Create a git tag for release. Usage: make tag v=1.0.0
	git tag -a v$(v) -m "release: v$(v)"
	git push origin v$(v)

# Run the application in Docker
.PHONY: run
run: build
	docker run --rm -it $(IMAGE_NAME):$(IMAGE_TAG)

# Help target
.PHONY: help
help: ## Display this help message
	@echo ""
	@echo "Usage:"
	@echo "  make ^<target^>"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build the application using Docker (default)"
	@echo "  build-local    Build the application locally (for development)"
	@echo "  build-windows  Build Windows executable with .exe extension"
	@echo "  build-all      Build for multiple platforms"
	@echo "  install        Install the application"
	@echo "  test           Run tests"
	@echo "  cover          Run tests with coverage"
	@echo "  clean          Clean build artifacts"
	@echo "  tag            Create a git tag for release. Usage: make tag v=1.0.0"
	@echo "  run            Run the application in Docker"
	@echo "  help           Display this help message"
	@echo ""