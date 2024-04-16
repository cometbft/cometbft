---
order: 3
---

# Install CometBFT

## From Go Package

Install the latest version of CometBFT's Go package:

```sh
go install github.com/cometbft/cometbft/cmd/cometbft@latest
```

Install a specific version of CometBFT's Go package:

```sh
go install github.com/cometbft/cometbft/cmd/cometbft@v0.38
```

## From Binary

To download pre-built binaries, see the [releases page](https://github.com/cometbft/cometbft/releases).

## From Source

You'll need `go` [installed](https://golang.org/doc/install) and the required
environment variables set, which can be done with the following commands:

```sh
echo export GOPATH=\"\$HOME/go\" >> ~/.bash_profile
echo export PATH=\"\$PATH:\$GOPATH/bin\" >> ~/.bash_profile
```

### Get Source Code

```sh
git clone https://github.com/cometbft/cometbft.git
cd cometbft
```

### Compile

```sh
make install
```

to put the binary in `$GOPATH/bin` or use:

```sh
make build
```

to put the binary in `./build`.

_DISCLAIMER_ The binary of CometBFT is build/installed without the DWARF
symbol table. If you would like to build/install CometBFT with the DWARF
symbol and debug information, remove `-s -w` from `BUILD_FLAGS` in the make
file.

The latest CometBFT is now installed. You can verify the installation by
running:

```sh
cometbft version
```

## Reinstall

If you already have CometBFT installed, and you make updates, simply

```sh
make install
```

To upgrade, run

```sh
git pull origin main
make install
```
