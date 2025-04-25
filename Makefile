.PHONY: build clean test dist install

# Get the git commit hash (short form)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Get version from version.go
VERSION := $(shell grep 'Version = ' pkg/framefold/version.go | cut -d'"' -f2)

# Build flags
LDFLAGS := -ldflags "-X 'framefold/pkg/framefold.CommitHash=$(GIT_COMMIT)'"

# Build targets for different platforms
TARGETS := \
	dist/framefold-$(VERSION)-darwin-amd64.tar.gz \
	dist/framefold-$(VERSION)-darwin-arm64.tar.gz \
	dist/framefold-$(VERSION)-linux-amd64.tar.gz \
	dist/framefold-$(VERSION)-linux-arm64.tar.gz \
	dist/framefold-$(VERSION)-linux-armv6.tar.gz \
	dist/framefold-$(VERSION)-linux-armv7.tar.gz

# Default target
all: clean build

# Build all targets
build: $(TARGETS)

# Create distribution directory
dist:
	@mkdir -p dist
	@mkdir -p build

# Create tarball with binary and documentation
define make_tarball
	@echo "Creating tarball for $(1)"
	@mkdir -p build/tmp/$(1)
	@cp build/$(1)/framefold build/tmp/$(1)/
	@cp README.md build/tmp/$(1)/
	@cp LICENSE build/tmp/$(1)/ 2>/dev/null || echo "No LICENSE file found"
	@cp config.json build/tmp/$(1)/config.example.json 2>/dev/null || echo "No config.json found"
	@cd build/tmp && tar -czf ../../dist/$(1).tar.gz $(1)
	@rm -rf build/tmp
	@rm -rf build/$(1)
endef

# Platform-specific builds
dist/framefold-$(VERSION)-darwin-amd64.tar.gz: dist
	@mkdir -p build/framefold-$(VERSION)-darwin-amd64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o build/framefold-$(VERSION)-darwin-amd64/framefold .
	$(call make_tarball,framefold-$(VERSION)-darwin-amd64)

dist/framefold-$(VERSION)-darwin-arm64.tar.gz: dist
	@mkdir -p build/framefold-$(VERSION)-darwin-arm64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o build/framefold-$(VERSION)-darwin-arm64/framefold .
	$(call make_tarball,framefold-$(VERSION)-darwin-arm64)

dist/framefold-$(VERSION)-linux-amd64.tar.gz: dist
	@mkdir -p build/framefold-$(VERSION)-linux-amd64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o build/framefold-$(VERSION)-linux-amd64/framefold .
	$(call make_tarball,framefold-$(VERSION)-linux-amd64)

dist/framefold-$(VERSION)-linux-arm64.tar.gz: dist
	@mkdir -p build/framefold-$(VERSION)-linux-arm64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o build/framefold-$(VERSION)-linux-arm64/framefold .
	$(call make_tarball,framefold-$(VERSION)-linux-arm64)

dist/framefold-$(VERSION)-linux-armv6.tar.gz: dist
	@mkdir -p build/framefold-$(VERSION)-linux-armv6
	GOOS=linux GOARCH=arm GOARM=6 go build $(LDFLAGS) -o build/framefold-$(VERSION)-linux-armv6/framefold .
	$(call make_tarball,framefold-$(VERSION)-linux-armv6)

dist/framefold-$(VERSION)-linux-armv7.tar.gz: dist
	@mkdir -p build/framefold-$(VERSION)-linux-armv7
	GOOS=linux GOARCH=arm GOARM=7 go build $(LDFLAGS) -o build/framefold-$(VERSION)-linux-armv7/framefold .
	$(call make_tarball,framefold-$(VERSION)-linux-armv7)

# Build for the current platform
local:
	go build $(LDFLAGS) -o framefold .

# Install locally (uses host OS/ARCH)
install:
	go install $(LDFLAGS)

# Clean build artifacts
clean:
	rm -rf build/
	rm -rf dist/
	rm -f framefold

# Run tests
test:
	go test -v ./...
