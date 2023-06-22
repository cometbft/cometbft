include ../../common.mk

all: node generator runner docker

docker:
	@echo "Building Simulator Docker image"
	@docker build \
		--tag 0track/simulator-node:local-version \
		-f docker/Dockerfile.tiny .

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
