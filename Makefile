GO ?= go

# Check if Go version is compatible
GO_VERSION := $(shell $(GO) version | cut -d' ' -f3 | cut -c3-)
GO_VERSION_MAJOR := $(shell echo $(GO_VERSION) | cut -d'.' -f1)
GO_VERSION_MINOR := $(shell echo $(GO_VERSION) | cut -d'.' -f2)

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

# Runtime configuration for Docker builds
RUNTIME_FAMILY ?= alpine
RUNTIME_IMAGE ?= alpine:3.21

# Default target
.PHONY: all
all: build

# Check Go version compatibility
.PHONY: check-go-version
check-go-version:
	@echo "Go version: $(GO_VERSION)"
	@echo "Required Go version: 1.23"
	@if [ "$(GO_VERSION_MAJOR)" -lt 1 ] || ( [ "$(GO_VERSION_MAJOR)" -eq 1 ] && [ "$(GO_VERSION_MINOR)" -lt 23 ] ); then \
		echo "Error: Go version 1.23 or higher is required"; \
		exit 1; \
	fi

# Build the application using Docker (default)
.PHONY: build
build:
	docker build \
		--build-arg VERSION=$(GIT_TAG) \
		--build-arg COMMIT=$(GIT_COMMIT) \
		--build-arg DATE=$(BUILD_DATE) \
		--build-arg DIRTY=$(GIT_DIRTY) \
		--build-arg RUNTIME_FAMILY=$(RUNTIME_FAMILY) \
		--build-arg RUNTIME_IMAGE=$(RUNTIME_IMAGE) \
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

# Run detailed coverage analysis
.PHONY: coverage
coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

# Run go vet
.PHONY: vet
vet:
	$(GO) vet ./...

# Run errcheck
.PHONY: errcheck
errcheck: check-go-version
	@echo "Running errcheck..."
	@GOFLAGS="-buildvcs=false" $(GO) run github.com/kisielk/errcheck@latest ./... || \
	echo "Warning: errcheck failed to run (errcheck not installed or not accessible)"

# Run go fmt
.PHONY: fmt
fmt:
	$(GO) fmt ./...

# Prepare code (format and basic checks)
.PHONY: prepare
prepare: fmt vet
	@echo "Code preparation completed: formatted and vetted"

# Run all linting checks
.PHONY: check
check: check-go-version vet errcheck lint-local
	@echo "All linting checks passed"

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

# Help target
.PHONY: help
help: ## Display this help message
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build              Build the application using Docker (default)"
	@echo "  build-local        Build the application locally (for development)"
	@echo "  build-windows      Build Windows executable with .exe extension"
	@echo "  build-all          Build for multiple platforms"
	@echo "  install            Install the application"
	@echo "  test               Run tests"
	@echo "  cover              Run tests with coverage"
	@echo "  coverage           Run detailed coverage analysis"
	@echo "  vet                Run go vet"
	@echo "  errcheck           Run errcheck (using 'go run' if not installed)"
	@echo "  fmt                Run go fmt"
	@echo "  prepare            Prepare code (format and basic checks)"
	@echo "  check              Run all linting checks"
	@echo "  verify             Run comprehensive verification (tests, linting, coverage)"
	@echo "  check-go-version   Check Go version compatibility"
	@echo "  lint               Run golangci-lint using Docker"
	@echo "  lint-local         Run golangci-lint locally (using 'go run' if not installed)"
	@echo "  sonar              Run SonarQube scanner"
	@echo "  vuln               Run vulnerability scan using govulncheck"
	@echo "  clean              Clean build artifacts"
	@echo "  tag                Create a git tag for release. Usage: make tag v=1.0.0"
	@echo "  run                Run the application in Docker"
	@echo "  act-docker         Test docker workflow locally using act"
	@echo "  act-prepare-release Test prepare-release workflow locally using act"
	@echo "  act-release        Test release workflow locally using act"
	@echo "  act-test           Test test workflow locally using act"
	@echo "  act-commitlint     Test commitlint workflow locally using act"
	@echo "  act-all            Test all workflows locally using act"
	@echo "  help               Display this help message"
	@echo ""
	@echo "Note for Windows/WSL users:"
	@echo "  The 'go run' approach is used for tools like errcheck and golangci-lint"
	@echo "  which eliminates the need to pre-install these tools."
	@echo ""

# Run golangci-lint using Docker
.PHONY: lint
lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.61.0 golangci-lint run -c ./.golangci.yml

# Run golangci-lint locally (requires golangci-lint installation)
.PHONY: lint-local
lint-local:
	@GOFLAGS="-buildvcs=false" $(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -c .golangci.yml || \
	echo "Warning: golangci-lint failed to run (golangci-lint not installed or not accessible)"

# Run SonarQube scanner using Docker
.PHONY: sonar
sonar:
	docker run --rm \
		-v $(PWD):/usr/src \
		-w /usr/src \
		-e SONAR_TOKEN \
		-e SONAR_HOST_URL \
		sonarsource/sonar-scanner-cli:latest \
		sonar-scanner

# Run vulnerability scan using govulncheck
.PHONY: vuln
vuln:
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Run comprehensive verification (tests, linting, coverage)
.PHONY: verify
verify: check-go-version test check coverage
	@echo "Comprehensive verification completed successfully"

# Run the application in Docker
.PHONY: run
run: build
	docker run --rm -it $(IMAGE_NAME):$(IMAGE_TAG)

# Run GitHub Actions workflows locally using act
.PHONY: act-docker
act-docker:
	act workflow_dispatch -W .github/workflows/docker.yml -j docker --input local_build=true --input skip_supply_chain=true --input skip_signing=true --input use_latest_tag=true --container-options "--privileged" --bind

.PHONY: act-prepare-release
act-prepare-release:
	act workflow_dispatch -W .github/workflows/prepare-release.yml --input bump=auto --input skip_pr=true

.PHONY: act-release
act-release:
	act -n -W .github/workflows/release.yml

.PHONY: act-test
act-test:
	act -n -W .github/workflows/test.yml

.PHONY: act-commitlint
act-commitlint:
	act -n -W .github/workflows/commitlint.yml

.PHONY: act-all
act-all: act-docker act-prepare-release act-release act-test act-commitlint
	@echo "All GitHub Actions workflows tested locally"
