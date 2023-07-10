## Fast Prototyping

This directory contains a set of scripts to run experiments for fast-prototyping cometbft reactors.
Please note that (to date) these scripts are mostly oriented toward prototyping mempool reactors.

For the impatient

	cd path/to/cometbft/
	E2E_DIR=$(pwd)/test/e2e
	(cd ${E2E_DIR} && make node generator runner docker-fast)
	${E2E_DIR}/fast-prototyping/experiments.sh ${E2E_DIR}/fast-prototyping/render/results.csv # takes around a minute
	${E2E_DIR}/fast-prototyping/render.sh

## Injecting a custom reactor

A new reactor is to be added under `fast-prototyping/reactors`.
Then, the reactor may be referenced in `e2e/node/main.go`.
More precisely, it has to be added in the `registry` variable.
This variable holds a mapping between a name and a reactor (because there is no reflexive API in golang).

Once this is done, we can inject the reactor in the TOML configuration file of the test network.
An injection is using the mapping (json-encoded) `experimental_custom_reactors`.
For instance, the mapping `CONSENSUS = p2p.mock.reactor` indicates that we mock the consensus reactor.

## Running a (custom) benchmark

The entry point for benchmarking a reactor is the `custom` benchmark.
This benchmark is accessible through the command line of the runner.
It is also possible to use the usual sequence, that is `start`, `load`, `stop` then `cleanup`.
The `custom` benchmark reports several metrics of interest, in particular to assess the bandwidth consumption of the system.
Part of these metrics are also output using the (new) `stats` command.

## Running an experiment

We may run an experiment in full with the help of `experiments.sh`.

## Plotting 

To plot the result of an experiment, we use `render.sh`.
This starts a web server to display the experimental results using Apache Echarts.
