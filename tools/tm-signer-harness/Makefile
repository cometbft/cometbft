.PHONY: build install docker-image

COMETBFT_VERSION?=latest
BUILD_TAGS?='cometbft'
VERSION := $(shell git describe --always)
BUILD_FLAGS = -ldflags "-X github.com/cometbft/cometbft/version.TMCoreSemVer=$(VERSION)

.DEFAULT_GOAL := build

build:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -tags $(BUILD_TAGS) -o ../../build/tm-signer-harness main.go

install:
	CGO_ENABLED=0 go install $(BUILD_FLAGS) -tags $(BUILD_TAGS) .

docker-image:
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -tags $(BUILD_TAGS) -o tm-signer-harness main.go
	docker build \
		--build-arg COMETBFT_VERSION=$(COMETBFT_VERSION) \
		-t tendermint/tm-signer-harness:$(COMETBFT_VERSION) .
	rm -rf tm-signer-harness
