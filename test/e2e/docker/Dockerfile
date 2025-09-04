# ---------- Stage 0: base (Go toolchain + DB build deps + RocksDB) ----------
ARG GO_VERSION=1.23
FROM golang:${GO_VERSION} AS base
ARG ROCKSDB=8.11.3

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      ca-certificates curl wget make tar build-essential git pkg-config \
      libbz2-dev libgflags-dev libsnappy-dev libzstd-dev zlib1g-dev liblz4-dev \
      libleveldb-dev libleveldb1d \
 && rm -rf /var/lib/apt/lists/*

RUN set -eux; \
  wget -q "https://github.com/facebook/rocksdb/archive/refs/tags/v${ROCKSDB}.tar.gz"; \
  tar -zxf "v${ROCKSDB}.tar.gz"; \
  cd "rocksdb-${ROCKSDB}"; \
  DEBUG_LEVEL=0 make -j"$(nproc)" shared_lib; \
  make install-shared; \
  ldconfig; \
  cd ..; \
  rm -rf "v${ROCKSDB}.tar.gz" "rocksdb-${ROCKSDB}"

# ---------- Stage 1: build (compile cometbft + e2e node) ----------
FROM base AS build

# Set up build directory /src/cometbft
WORKDIR /src/cometbft

COPY go.mod go.sum ./
RUN go mod download
COPY . .

ENV COMETBFT_BUILD_OPTIONS badgerdb,cleveldb,rocksdb,pebbledb,clock_skew,bls12381
RUN make build
RUN cd test/e2e && make node

# ---------- Stage 2: runtime (minimal) ----------
FROM debian:bookworm-slim AS runtime

RUN apt-get -qq update \
 && apt-get -qq upgrade -y \
 && apt-get -qq install -y --no-install-recommends \
      iputils-ping iproute2 \
      libbz2-1.0 libzstd1 liblz4-1 zlib1g \
      libleveldb1d libsnappy1v5 libgflags2.2 \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /cometbft
VOLUME /cometbft
ENV CMTHOME=/cometbft
ENV GORACE="halt_on_error=1"

COPY --from=build /usr/local/lib/librocksdb.so* /usr/local/lib/
RUN ldconfig

COPY --from=build /src/cometbft/test/e2e/docker/entrypoint* /usr/bin/
COPY --from=build /src/cometbft/build/cometbft /usr/bin/cometbft
COPY --from=build /src/cometbft/test/e2e/build/node /usr/bin/app

EXPOSE 26656 26657 26660 6060
ENTRYPOINT ["/usr/bin/entrypoint"]
CMD ["node"]
