# Multi-stage build Dockerfile for CometBFT end-to-end testing
# Stage 1: Build environment - compiles the application and test node
FROM cometbft/cometbft-db-testing:v0.14.2 AS build

WORKDIR /src/cometbft

# Copy Go module files and download dependencies
# This is done before copying the rest of the code to leverage Docker's build cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire codebase into the container
COPY . .

# Set build options to include specific features and databases
ENV COMETBFT_BUILD_OPTIONS=badgerdb,boltdb,cleveldb,rocksdb
RUN make build
RUN cd test/e2e && make node

# Stage 2: Final image - minimal runtime environment
FROM debian:bookworm-slim AS runtime

# Update system packages and install network utilities
RUN apt-get -qq update -y && apt-get -qq upgrade -y >/dev/null
RUN apt-get -qq install -y iputils-ping iproute2 >/dev/null
RUN apt install libsnappy-dev libgflags-dev libleveldb-dev -y >/dev/null

WORKDIR /cometbft
VOLUME /cometbft
ENV CMTHOME=/cometbft

# Configure Go race detector to halt on error
ENV GORACE="halt_on_error=1"

# Copy RocksDB shared libraries from the build stage
COPY --from=build /usr/local/lib/librocksdb.so* /lib/

# Copy executables from the build stage
# - entrypoint script for container initialization
# - cometbft binary for the blockchain node
# - app binary for the test application
COPY --from=build /src/cometbft/test/e2e/docker/entrypoint* /usr/bin/
COPY --from=build /src/cometbft/build/cometbft /usr/bin/cometbft
COPY --from=build /src/cometbft/test/e2e/build/node /usr/bin/app

# Expose ports:
# - 26656: P2P communication between nodes
# - 26657: RPC server for API requests
# - 26660: ABCI server for application communication
# - 6060: Prometheus metrics endpoint
EXPOSE 26656 26657 26660 6060

# Set the entrypoint script to initialize the container
ENTRYPOINT ["/usr/bin/entrypoint"]
# Default command to run when container starts (can be overridden)
CMD ["node"]
