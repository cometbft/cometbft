# We need to build in a Linux environment to support C libraries.
FROM golang:1.23

RUN apt-get -qq update -y && apt-get -qq upgrade -y >/dev/null

# For latency emulation, install iproute2, which includes tc.
RUN apt-get -qq install -y iputils-ping iproute2 >/dev/null

# Set up build directory /src/cometbft
ENV COMETBFT_BUILD_OPTIONS clock_skew,bls12381,secp256k1eth
WORKDIR /src/cometbft

# Fetch dependencies separately (for layer caching)
COPY go.mod go.sum api/go.mod api/go.sum ./
RUN go mod download

# Build CometBFT and install into /usr/bin/cometbft
COPY . .
RUN make build && cp build/cometbft /usr/bin/cometbft
COPY test/e2e/docker/entrypoint* /usr/bin/
RUN cd test/e2e && make node && cp build/node /usr/bin/app

# Set up runtime directory.
WORKDIR /cometbft
VOLUME /cometbft
ENV CMTHOME=/cometbft
ENV GORACE="halt_on_error=1"

EXPOSE 26656 26657 26660 6060
ENTRYPOINT ["/usr/bin/entrypoint"]
CMD ["node"]
STOPSIGNAL SIGTERM
