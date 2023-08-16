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
## Run prototype

The `fast-prototyping/experiments.sh` script exemplifies how to run a set of experiments on prototype reactors and the lean image built on the previous step.
The experiment uses the new `custom` command in the e2e framework.

As is, if ran as `./fast-prototyping/experiments.sh output.csv` the script runs for about two minutes and output a series of metrics to the `output.csv` file for the configurations specified.
You may wish to change the script to adjust the following configurations:

- propagation rate (`PROP_RATE`), or the rate of txs that will not be skipped during gossip
- number of nodes simulated (`NUM_NODES`)

You can also specify if 0, 1, or all nodes will be validators. 
Any non-validator is a full node.

The script creates a temporary `custom.toml` manifest file under the `/tmp` folder while running.
