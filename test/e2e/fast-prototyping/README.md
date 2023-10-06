# Fast Prototyping

This directory contains a set of scripts to run experiments for fast-prototyping cometbft reactors.
Please note that (to date) these scripts are mostly oriented toward prototyping mempool reactors.

## Fast docker image

To use the fast-prototyping feature and be able to run dozens of nodes in containers, you need a very lean docker image.
To build it:
- Have an updated docker installed
- From the CometBFT clone, build
   ```bash
   E2E_DIR=$(pwd)/test/e2e
   cd ${E2E_DIR}
   make node-fast generator runner docker-fast
   ```
  
- At this point you can run some tests using the fast docker image, but be aware that it is a very lean image.
   ```
   ./run-multiple.sh networks/long.toml
   ```
