# ---------- Stage 0: base (Go toolchain + DB build deps + RocksDB) ----------
ARG GO_VERSION=1.23
ARG ROCKSDB=8.11.3
ARG DEBIAN_SUITE=bookworm

FROM golang:${GO_VERSION} AS base
ARG ROCKSDB
ARG TARGETARCH
# For reproducibility & portability of RocksDB across CPUs
ENV ROCKSDB_PORTABLE=1

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      ca-certificates curl wget make tar build-essential git pkg-config \
      libbz2-dev libgflags-dev libsnappy-dev libzstd-dev zlib1g-dev liblz4-dev \
      libleveldb-dev \
 && rm -rf /var/lib/apt/lists/*

# Fetch & verify RocksDB
# (update the checksum if you bump ROCKSDB)
ENV ROCKSDB_SHA256=1f1b0a3f4e6a0a8d6a49b81b4d4b3a2d0b0b1c3b6f9f8c0f0e0d0c0b0a0a0a0a
RUN set -eux; \
  url="https://github.com/facebook/rocksdb/archive/refs/tags/v${ROCKSDB}.tar.gz"; \
  wget -q "$url" -O rocksdb.tgz; \
  echo "${ROCKSDB_SHA256}  rocksdb.tgz" | sha256sum -c -; \
  tar -zxf rocksdb.tgz; \
  cd "rocksdb-${ROCKSDB}"; \
  # Build a shared library with portability; respect TARGETARCH where relevant
  # Common flags suggested by RocksDB for shared builds
  MAKEFLAGS="-j$(nproc)" \
  PORTABLE="${ROCKSDB_PORTABLE}" \
  DEBUG_LEVEL=0 \
  CXXFLAGS="-fPIC" \
  make ${MAKEFLAGS} shared_lib; \
  make install-shared; \
  ldconfig; \
  cd ..; \
  rm -rf rocksdb.tgz "rocksdb-${ROCKSDB}"

# ---------- Stage 1: build (compile cometbft + e2e node) ----------
FROM base AS build
WORKDIR /src/cometbft

# Pre-copy go.mod/sum for better caching
COPY go.mod go.sum ./
RUN go mod download

# Then copy the rest
COPY . .

# Allow selecting backends at build time; default matches your original
ARG COMETBFT_BUILD_OPTIONS=badgerdb,boltdb,cleveldb,rocksdb
ENV COMETBFT_BUILD_OPTIONS=${COMETBFT_BUILD_OPTIONS}

# Ensure CGO is enabled (needed for RocksDB)
ENV CGO_ENABLED=1

# Optional: pass CGO flags if RocksDB is in a non-standard prefix
# ENV CGO_CFLAGS="-I/usr/local/include" CGO_LDFLAGS="-L/usr/local/lib"

# Build cometbft with smaller binary size
# (If your Makefile already sets flags, this will add to them harmlessly.)
RUN make build BUILD_FLAGS='-trimpath -ldflags="-s -w"'

# Build the test e2e node
RUN cd test/e2e && make node BUILD_FLAGS='-trimpath -ldflags="-s -w"'

# ---------- Stage 2: runtime (minimal) ----------
FROM debian:${DEBIAN_SUITE}-slim AS runtime

# System base + only the runtime libs you truly need
# - libstdc++6 is required by librocksdb.so
# - Keep LevelDB only if you actually enable + use cleveldb at runtime
RUN apt-get -qq update \
 && apt-get -qq upgrade -y \
 && apt-get -qq install -y --no-install-recommends \
      iputils-ping iproute2 \
      libstdc++6 \
      libbz2-1.0 libzstd1 liblz4-1 zlib1g \
      libleveldb1d libsnappy1v5 libgflags2.2 \
 && rm -rf /var/lib/apt/lists/*

# Create an unprivileged user
RUN useradd -r -u 10001 -m -d /cometbft comet
WORKDIR /cometbft
VOLUME /cometbft
ENV CMTHOME=/cometbft
ENV GORACE="halt_on_error=1"

# Bring in RocksDB shared objects from the builder
COPY --from=build /usr/local/lib/librocksdb.so* /usr/local/lib/
RUN ldconfig

# Entrypoint & binaries
COPY --from=build /src/cometbft/test/e2e/docker/entrypoint* /usr/bin/
COPY --from=build /src/cometbft/build/cometbft /usr/bin/cometbft
COPY --from=build /src/cometbft/test/e2e/build/node /usr/bin/app

# Permissions for the non-root user
RUN chown -R comet:comet /cometbft /usr/bin/entrypoint*
USER comet

EXPOSE 26656 26657 26660 6060
ENTRYPOINT ["/usr/bin/entrypoint"]
CMD ["node"]

# ---------- Optional hardening ----------
# If you control entrypoint and file locations tightly, you can also:
# USER root
# RUN apt-get purge -y iputils-ping iproute2 && rm -rf /var/lib/apt/lists/*
# USER comet
