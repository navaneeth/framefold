.PHONY: all clean darwin linux raspberry windows

BINARY_NAME=framefold
VERSION=1.0.0
BUILD_DIR=build

all: clean darwin linux raspberry

clean:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)

# Build for macOS (both AMD64 and ARM64 for M1/M2 Macs)
darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64

# Build for Linux (AMD64)
linux:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64

# Build for Raspberry Pi (ARM)
raspberry:
	GOOS=linux GOARCH=arm GOARM=6 go build -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-raspberry-armv6
	GOOS=linux GOARCH=arm GOARM=7 go build -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-raspberry-armv7
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-raspberry-arm64

# Install locally (uses host OS/ARCH)
install:
	go install

# Run tests
test:
	go test -v ./...
