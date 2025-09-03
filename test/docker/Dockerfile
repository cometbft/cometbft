FROM golang:1.22

# Avoid interactive prompts during apt operations
ENV DEBIAN_FRONTEND=noninteractive

# Grab deps (jq, hexdump, xxd, killall)
# - hexdump lives in bsdextrautils on bookworm
# - xxd is available as its own package (no need for vim-common)
# - netcat is virtual; pick netcat-openbsd
RUN apt-get update && \
  apt-get install -y --no-install-recommends \
    jq \
    bsdextrautils \
    xxd \
    psmisc \
    netcat-openbsd \
    curl \
  && rm -rf /var/lib/apt/lists/*

# Setup CometBFT repo
ENV REPO=$GOPATH/src/github.com/cometbft/cometbft
ENV GOBIN=$GOPATH/bin
WORKDIR $REPO

# (Optional cache boost) copy mod files first for better go mod caching
# Comment these two lines out if they don't exist in your tree
COPY go.mod go.sum ./
RUN go mod download || true

# Copy in the code
COPY . $REPO

# install ABCI CLI
RUN make install_abci

# install CometBFT
RUN make install

RUN cometbft testnet \
  --config $REPO/test/docker/config-template.toml \
  --node-dir-prefix="mach" \
  --v=4 \
  --populate-persistent-peers=false \
  --o=$REPO/test/p2p/data

# Now copy in the code
# NOTE: this will overwrite whatever is in vendor/
COPY . $REPO

# expose the volume for debugging
VOLUME $REPO

EXPOSE 26656
EXPOSE 26657
