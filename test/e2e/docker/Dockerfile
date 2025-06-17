FROM cometbft/cometbft-db-testing:v1.0.4 AS build

RUN apt-get -qq update -y && apt-get -qq upgrade -y >/dev/null
RUN apt-get -qq install -y iputils-ping iproute2 >/dev/null

WORKDIR /src/cometbft

COPY go.mod go.sum api/go.mod api/go.sum ./
RUN go mod download

COPY . .

ENV COMETBFT_BUILD_OPTIONS=badgerdb,rocksdb,clock_skew,bls12381,secp256k1eth
RUN make build
RUN cd test/e2e && make node

FROM debian:bookworm-slim

WORKDIR /cometbft
VOLUME /cometbft
ENV CMTHOME=/cometbft
ENV GORACE="halt_on_error=1"

COPY --from=build /usr/local/lib/librocksdb.so* /lib/

COPY --from=build /src/cometbft/test/e2e/docker/entrypoint* /usr/bin/
COPY --from=build /src/cometbft/build/cometbft /usr/bin/cometbft
COPY --from=build /src/cometbft/test/e2e/build/node /usr/bin/app

EXPOSE 26656 26657 26660 6060

ENTRYPOINT ["/usr/bin/entrypoint"]
CMD ["node"]
