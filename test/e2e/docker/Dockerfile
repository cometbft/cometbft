# ---------- Stage 0: base (Go toolchain + DB build deps + RocksDB) ----------
ARG GO_VERSION=1.24
ARG ROCKSDB=8.11.3

FROM golang:${GO_VERSION}-bookworm AS base
ARG ROCKSDB
# Optional checksum: set via --build-arg ROCKSDB_SHA256=<sha>. Leave blank to skip.
ARG ROCKSDB_SHA256=3b51d1d907ea13fab430bf052618610994f08cd8ed0b1341c3e89fe02e199f8e
# Build portable RocksDB
ENV ROCKSDB_PORTABLE=1

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      ca-certificates curl wget make tar build-essential git pkg-config \
      libbz2-dev libgflags-dev libsnappy-dev libzstd-dev zlib1g-dev liblz4-dev \
      libleveldb-dev \
 && rm -rf /var/lib/apt/lists/*

RUN set -eux; \
  url="https://github.com/facebook/rocksdb/archive/refs/tags/v${ROCKSDB}.tar.gz"; \
  wget -q "$url" -O rocksdb.tgz; \
  if [ -n "${ROCKSDB_SHA256:-}" ] && [ "${ROCKSDB_SHA256}" != "skip" ]; then \
    echo "${ROCKSDB_SHA256}  rocksdb.tgz" | sha256sum -c -; \
  else \
    echo "WARNING: skipping RocksDB SHA256 verification"; \
  fi; \
  tar -zxf rocksdb.tgz; \
  cd "rocksdb-${ROCKSDB}"; \
  PORTABLE="${ROCKSDB_PORTABLE}" DEBUG_LEVEL=0 CXXFLAGS="-fPIC" make -j"$(nproc)" shared_lib; \
  make install-shared; \
  ldconfig; \
  cd ..; \
  rm -rf rocksdb.tgz "rocksdb-${ROCKSDB}"

# ---------- Stage 1: build (compile cometbft + e2e node) ----------
FROM base AS build

# Set up build directory /src/cometbft
WORKDIR /src/cometbft

# cache-friendly Go deps
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG COMETBFT_BUILD_OPTIONS=badgerdb,cleveldb,rocksdb,pebbledb,clock_skew,bls12381
ENV COMETBFT_BUILD_OPTIONS=${COMETBFT_BUILD_OPTIONS}
ENV CGO_ENABLED=1

# smaller binaries (safe even if Makefile sets its own flags)
RUN make build BUILD_FLAGS='-trimpath -ldflags="-s -w"'
RUN cd test/e2e && make node BUILD_FLAGS='-trimpath -ldflags="-s -w"'

# ---------- Stage 2: runtime (minimal) ----------
FROM debian:bookworm-slim AS runtime

RUN apt-get -qq update \
 && apt-get -qq upgrade -y \
 && apt-get -qq install -y --no-install-recommends \
      iputils-ping iproute2 \
      libstdc++6 \
      libbz2-1.0 libzstd1 liblz4-1 zlib1g \
      libleveldb1d libsnappy1v5 libgflags2.2 \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /cometbft
VOLUME /cometbft
ENV CMTHOME=/cometbft
ENV GORACE="halt_on_error=1"

# bring in RocksDB shared libs and binaries
COPY --from=build /usr/local/lib/librocksdb.so* /usr/local/lib/
RUN ldconfig

COPY --from=build /src/cometbft/test/e2e/docker/entrypoint* /usr/bin/
COPY --from=build /src/cometbft/build/cometbft /usr/bin/cometbft
COPY --from=build /src/cometbft/test/e2e/build/node /usr/bin/app

EXPOSE 26656 26657 26660 6060
ENTRYPOINT ["/usr/bin/entrypoint"]
CMD ["node"]
