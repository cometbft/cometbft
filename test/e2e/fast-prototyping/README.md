## Fast Prototyping

This directory contains a set of scripts to run experiments for fast-prototyping cometbft reactors.
Please note that (to date) these scripts are mostly oriented toward prototyping mempool reactors.

For the impatient

    cd path/to/cometbft/
    E2E_DIR=$(pwd)/test/e2e
    (cd ${E2E_DIR} && make node-fast generator runner docker-fast)
    ${E2E_DIR}/fast-prototyping/experiments.sh ${E2E_DIR}/fast-prototyping/render/sample.csv # takes around two minutes
    cp ${E2E_DIR}/fast-prototyping/render/results.csv.tmpl ${E2E_DIR}/fast-prototyping/render/results.csv
    echo "sample;A sample result.;sample.csv" >>  ${E2E_DIR}/fast-prototyping/render/results.csv
    ${E2E_DIR}/fast-prototyping/render.sh

## Injecting a custom reactor

A new reactor should be added under `fast-prototyping/reactors`.
Then, the reactor has to be referenced in `e2e/node/main.go`, using the `registry` variable.
This variable holds a mapping between a name and a reactor.
(Notice here that there is no reflexive API in golang.)

Once this is done, we can inject the reactor in the TOML configuration file of the test network.
An injection is using the mapping (json-encoded) `experimental_custom_reactors`.
For instance, the mapping `CONSENSUS = p2p.mock.reactor` indicates that we mock the consensus reactor.

## Running a (custom) benchmark

The entry point for benchmarking a reactor is the `custom` benchmark (`runner/main.go`).
This benchmark is accessible through the command line of the runner.
It is also possible to use the standard sequence `start`, `load`, `stop` then `cleanup`.

The `custom` benchmark reports several metrics of interest, in particular to assess the bandwidth consumption of the mempool.
Part of these metrics are also provided using the (new) `stats` command.

## Running an experiment

We may run an experiment using the help of `experiments.sh`.

## Plotting

To plot an experiment, we use `render.sh`.
This script starts a web server that displays the experimental results using Apache Echarts.
Each result comes as a button at the top of the screen.
To be displayed, a result must be referenced under `fast-prototyping/render/result.csv`.
A reference is of the form `name;description;location`.
For instance, `sample;A sample result.;/tmp/sample.csv` indicates that the result called `sample` is stored under `/tmp/sample.csv`.
The tooltip `A sample result` is displayed when the mouse is over the button attached to the result.


