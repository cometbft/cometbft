# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# Build parameters
BUILD_DIR = build
SOURCE_DIR = cmd/main/main.go

BINARY_AMD64_NAME = client
LINUX_AMD64_BINARY = $(BUILD_DIR)/$(BINARY_AMD64_NAME)

BINARY_MAC_ARM64_NAME = client_mac_arm64
MAC_BINARY = $(BUILD_DIR)/$(BINARY_MAC_ARM64_NAME)

# Build targets
build: linux_amd64

linux_amd64:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(LINUX_AMD64_BINARY) $(SOURCE_DIR)

mac_arm64:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(MAC_BINARY) $(SOURCE_DIR)


.PHONY: build linux_arm64
