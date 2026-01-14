---
order: 2
---

# Docker Compose

With Docker Compose, you can spin up local testnets with a single command.

## Requirements

1. [Install CometBFT](../guides/install.md)
2. [Install docker](https://docs.docker.com/engine/installation/)
3. [Install docker-compose](https://docs.docker.com/compose/install/)

## Build

Build the `cometbft` binary and, optionally, the `cometbft/localnode`
docker image.

Note the binary will be mounted into the container so it can be updated without
rebuilding the image.

```sh
# Build the linux binary in ./build
make build-linux

# (optionally) Build cometbft/localnode image
make build-docker-localnode
```

## Run a testnet

To start a 4 node testnet run:

```sh
make localnet-start
```

The nodes bind their RPC servers to ports 26657, 26660, 26662, and 26664 on the
host.

This file creates a 4-node network using the localnode image.

The nodes of the network expose their P2P and RPC endpoints to the host machine
on ports 26656-26657, 26659-26660, 26661-26662, and 26663-26664 respectively.

To update the binary, just rebuild it and restart the nodes:

```sh
make build-linux
make localnet-start
```

## Configuration

The `make localnet-start` creates files for a 4-node testnet in `./build` by
calling the `cometbft testnet` command.

The `./build` directory is mounted to the `/cometbft` mount point to attach
the binary and config files to the container.

To inspect or modify how the local testnet is generated, see the `localnet-start` Makefile target [here](../../Makefile):

```makefile
localnet-start: localnet-stop build-docker-localnode
	@if ! [ -f build/node0/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/cometbft:Z cometbft/localnode testnet --config /etc/cometbft/config-template.toml --o . --starting-ip-address 192.167.10.2; fi
	docker compose up -d
```

By default this command generates configuration for a 4-validator testnet. To create a network with a different number of validators or non-validators, pass the appropriate `--v` and `--n` flags to the `testnet` command (either by running it manually before `docker compose up` or by extending the Makefile target above). When increasing the number of nodes, you must also add corresponding services to `docker-compose.yml` so that each generated node directory has a matching container definition.

```yml
  node3: # bump by 1 for every node
    container_name: node3 # bump by 1 for every node
    image: "cometbft/localnode"
    environment:
      - ID=3
      - LOG=${LOG:-cometbft.log}
    ports:
      - "26663-26664:26656-26657" # Bump 26663-26664 by one for every node
    volumes:
      - ./build:/cometbft:Z
    networks:
      localnet:
        ipv4_address: 192.167.10.5 # bump the final digit by 1 for every node
```

Before running it, don't forget to cleanup the old files:

```sh
# Clear the build folder
rm -rf ./build/node*
```

## Configuring ABCI containers

To use your own ABCI applications with 4-node setup edit the [docker-compose.yaml](https://github.com/cometbft/cometbft/blob/v0.38.x/docker-compose.yml) file and add images to your ABCI application.

```yml
 abci0:
    container_name: abci0
    image: "abci-image"
    build:
      context: .
      dockerfile: abci.Dockerfile
    command: <insert command to run your abci application>
    networks:
      localnet:
        ipv4_address: 192.167.10.6

  abci1:
    container_name: abci1
    image: "abci-image"
    build:
      context: .
      dockerfile: abci.Dockerfile
    command: <insert command to run your abci application>
    networks:
      localnet:
        ipv4_address: 192.167.10.7

  abci2:
    container_name: abci2
    image: "abci-image"
    build:
      context: .
      dockerfile: abci.Dockerfile
    command: <insert command to run your abci application>
    networks:
      localnet:
        ipv4_address: 192.167.10.8

  abci3:
    container_name: abci3
    image: "abci-image"
    build:
      context: .
      dockerfile: abci.Dockerfile
    command: <insert command to run your abci application>
    networks:
      localnet:
        ipv4_address: 192.167.10.9

```

Override the [command](https://github.com/cometbft/cometbft/blob/v0.38.x/networks/local/localnode/Dockerfile#L11) in each node to connect to it's ABCI.

```yml
  node0:
    container_name: node0
    image: "cometbft/localnode"
    ports:
      - "26656-26657:26656-26657"
    environment:
      - ID=0
      - LOG=$${LOG:-cometbft.log}
    volumes:
      - ./build:/cometbft:Z
    command: node --proxy_app=tcp://abci0:26658
    networks:
      localnet:
        ipv4_address: 192.167.10.2
```

Similarly do for node1, node2 and node3 then [run testnet](#run-a-testnet).

## Logging

Log is saved under the attached volume, in the `cometbft.log` file. If the
`LOG` environment variable is set to `stdout` at start, the log is not saved,
but printed on the screen.

## Special binaries

If you have multiple binaries with different names, you can specify which one
to run with the `BINARY` environment variable. The path of the binary is relative
to the attached volume.
