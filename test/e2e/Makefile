include ../../common.mk

all: node generator runner docker

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

docker-fast:
	@echo "Building fast-prototyping E2E Docker image"
	@docker build \
		--tag cometbft/e2e-node:local-version \
		-f docker/Dockerfile.fast .

# We need to build support for database backends into the app in
# order to build a binary with a CometBFT node in it (for built-in
# ABCI testing).
node:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/node ./node

generator:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/generator ./generator

runner:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/runner ./runner

.PHONY: all node docker generator runner
