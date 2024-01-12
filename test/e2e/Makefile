COMETBFT_BUILD_OPTIONS += badgerdb,boltdb,cleveldb,rocksdb,pebbledb

include ../../common.mk

all: docker generator runner

fast: docker-fast generator runner

# This will set up a container with all required dependencies for compiling, copy all the source
# code to the container, and compile the binary inside it.
docker:
	@echo "Building E2E Docker image"
	@docker build \
		--tag cometbft/e2e-node:local-version \
		-f docker/Dockerfile ../..

docker-debug:
	@echo "Building E2E Docker image for debugging"
	@docker build \
		--tag cometbft/e2e-node:local-version \
		-f docker/Dockerfile.debug ../..

# This will compile a binary to be executed in a container, set up a slim container, and copy only
# the binary to it.
docker-fast:
	@echo "Compiling binary for slim E2E Docker image"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o build/node ./node
	@echo "Building slim E2E Docker image"
	@docker build \
		--tag cometbft/e2e-node:local-version \
		-f docker/Dockerfile.fast .

# We need to build support for database backends into the app in
# order to build a binary with a CometBFT node in it (for built-in
# ABCI testing).
node:
	go build -race $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o build/node ./node

generator:
	go build -o build/generator ./generator

runner:
	go build -o build/runner ./runner

lint:
	@echo "--> Running linter for E2E"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

grammar-gen:
	go run github.com/goccmack/gogll/v3@latest -o pkg/grammar/grammar-auto pkg/grammar/abci_grammar.md

.PHONY: all slow fast node docker docker-debug docker-fast generator runner lint grammar-gen
