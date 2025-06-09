# hks-func CLI Makefile

# Variables
BINARY_NAME := hks-func
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILT_BY := $(shell whoami)

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOVET := $(GOCMD) vet
GOFMT := gofmt

# Build flags
LDFLAGS := -ldflags "\
	-X 'github.com/hexabase/hexabase-ai/cli/hks-func/cmd.version=$(VERSION)' \
	-X 'github.com/hexabase/hexabase-ai/cli/hks-func/cmd.commit=$(COMMIT)' \
	-X 'github.com/hexabase/hexabase-ai/cli/hks-func/cmd.date=$(DATE)' \
	-X 'github.com/hexabase/hexabase-ai/cli/hks-func/cmd.builtBy=$(BUILT_BY)'"

# Platforms
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

# Default target
.DEFAULT_GOAL := build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# Install the binary
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -rf dist/

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
.PHONY: test-coverage
test-coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w .

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	$(GOVET) ./...

# Lint code (requires golangci-lint)
.PHONY: lint
lint:
	@echo "Linting code..."
	@golangci-lint run

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Build for all platforms
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOBUILD) $(LDFLAGS) \
		-o dist/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}$(if $(findstring windows,$${platform%/*}),.exe,) .; \
		echo "Built: dist/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}"; \
	done

# Create release archives
.PHONY: release
release: build-all
	@echo "Creating release archives..."
	@cd dist && \
	for file in $(BINARY_NAME)-*; do \
		if [[ "$$file" == *.exe ]]; then \
			zip "$${file%.exe}.zip" "$$file" ../README.md ../LICENSE; \
		else \
			tar czf "$$file.tar.gz" "$$file" -C .. README.md LICENSE; \
		fi; \
		rm "$$file"; \
	done
	@echo "Release archives created in dist/"

# Run the CLI
.PHONY: run
run: build
	./$(BINARY_NAME) $(ARGS)

# Development build (faster, no optimization)
.PHONY: dev
dev:
	@echo "Building development version..."
	$(GOBUILD) -o $(BINARY_NAME) .

# Check code quality
.PHONY: check
check: fmt vet lint test
	@echo "All checks passed!"

# Show help
.PHONY: help
help:
	@echo "hks-func CLI Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build         Build the binary for current platform"
	@echo "  install       Build and install to GOPATH/bin"
	@echo "  clean         Remove build artifacts"
	@echo "  test          Run tests"
	@echo "  test-coverage Generate test coverage report"
	@echo "  fmt           Format code"
	@echo "  vet           Vet code"
	@echo "  lint          Lint code (requires golangci-lint)"
	@echo "  deps          Download dependencies"
	@echo "  deps-update   Update dependencies"
	@echo "  build-all     Build for all platforms"
	@echo "  release       Create release archives"
	@echo "  run           Build and run with ARGS"
	@echo "  dev           Fast development build"
	@echo "  check         Run all quality checks"
	@echo "  help          Show this help message"

# Print version info
.PHONY: version
version:
	@echo "Version:  $(VERSION)"
	@echo "Commit:   $(COMMIT)"
	@echo "Date:     $(DATE)"
	@echo "Built by: $(BUILT_BY)"