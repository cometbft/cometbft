---
order: 1
parent:
  title: Method
  order: 1
---

# Method

This document provides a detailed description of the QA process.
It is intended to be used by engineers reproducing the experimental setup for future tests of CometBFT.

The (first iteration of the) QA process as described [in the RELEASES.md document][releases]
was applied to version v0.34.x in order to have a set of results acting as benchmarking baseline.
This baseline is then compared with results obtained in later versions.

Out of the testnet-based test cases described in [the releases document][releases] we focused on two of them:
_200 Node Test_, and _Rotating Nodes Test_.

[releases]: https://github.com/cometbft/cometbft/blob/v0.38.x/RELEASES.md#large-scale-testnets

## Software Dependencies

### Infrastructure Requirements to Run the Tests

* An account at Digital Ocean (DO), with a high droplet limit (>202)
* The machine to orchestrate the tests should have the following installed:
    * A clone of the [testnet repository][testnet-repo]
        * This repository contains all the scripts mentioned in the remainder of this section
    * [Digital Ocean CLI][doctl]
    * [Terraform CLI][Terraform]
    * [Ansible CLI][Ansible]

[testnet-repo]: https://github.com/cometbft/qa-infra
[Ansible]: https://docs.ansible.com/ansible/latest/index.html
[Terraform]: https://www.terraform.io/docs
[doctl]: https://docs.digitalocean.com/reference/doctl/how-to/install/

### Requirements for Result Extraction

* [Prometheus DB][prometheus] to collect metrics from nodes
* Prometheus DB to process queries (may be different node from the previous)
* blockstore DB of one of the full nodes in the testnet


[prometheus]: https://prometheus.io/

## 200 Node Testnet

### Running the test

This section explains how the tests were carried out for reproducibility purposes.

1. [If you haven't done it before]
   Follow steps 1-4 of the `README.md` at the top of the testnet repository to configure Terraform, and `doctl`.
2. Copy file `testnets/testnet200.toml` onto `testnet.toml` (do NOT commit this change)
3. Set the variable `VERSION_TAG` in the `Makefile` to the git hash that is to be tested.
   * If you are running the base test, which implies an homogeneous network (all nodes are running the same version),
     then make sure makefile variable `VERSION2_WEIGHT` is set to 0
   * If you are running a mixed network, set the variable `VERSION2_TAG` to the other version you want deployed
     in the network.
     Then adjust the weight variables `VERSION_WEIGHT` and `VERSION2_WEIGHT` to configure the
     desired proportion of nodes running each of the two configured versions.
4. Follow steps 5-10 of the `README.md` to configure and start the 200 node testnet
    * WARNING: Do NOT forget to run `make terraform-destroy` as soon as you are done with the tests (see step 9)
5. As a sanity check, connect to the Prometheus node's web interface (port 9090)
    and check the graph for the `cometbft_consensus_height` metric. All nodes
    should be increasing their heights.

    * You can find the Prometheus node's IP address in `ansible/hosts` under section `[prometheus]`.
    * The following URL will display the metrics `cometbft_consensus_height` and `cometbft_mempool_size`:

      ```
      http://<PROMETHEUS-NODE-IP>:9090/classic/graph?g0.range_input=1h&g0.expr=cometbft_consensus_height&g0.tab=0&g1.range_input=1h&g1.expr=cometbft_mempool_size&g1.tab=0
      ```

6. You now need to start the load runner that will produce transaction load.
    * If you don't know the saturation load of the version you are testing, you need to discover it.
        * Run `make loadrunners-init`. This will copy the loader scripts to the
          `testnet-load-runner` node and install the load tool.
        * Find the IP address of the `testnet-load-runner` node in
          `ansible/hosts` under section `[loadrunners]`.
        * `ssh` into `testnet-load-runner`.
          * Edit the script `/root/200-node-loadscript.sh` in the load runner
            node to provide the IP address of a full node (for example,
            `validator000`). This node will receive all transactions from the
            load runner node.
          * Run `/root/200-node-loadscript.sh` from the load runner node.
            * This script will take about 40 mins to run, so it is suggested to
              first run `tmux` in case the ssh session breaks.
            * It is running 90-seconds-long experiments in a loop with different
              loads.
    * If you already know the saturation load, you can simply run the test (several times) for 90 seconds with a load somewhat
      below saturation:
        * set makefile variables `LOAD_CONNECTIONS`, `LOAD_TX_RATE`, to values that will produce the desired transaction load.
        * set `LOAD_TOTAL_TIME` to 90 (seconds).
        * run "make runload" and wait for it to complete. You may want to run this several times so the data from different runs can be compared.
7. Run `make retrieve-data` to gather all relevant data from the testnet into the orchestrating machine
    * Alternatively, you may want to run `make retrieve-prometheus-data` and `make retrieve-blockstore` separately.
      The end result will be the same.
    * `make retrieve-blockstore` accepts the following values in makefile variable `RETRIEVE_TARGET_HOST`
        * `any`: (which is the default) picks up a full node and retrieves the blockstore from that node only.
        * `all`: retrieves the blockstore from all full nodes; this is extremely slow, and consumes plenty of bandwidth,
           so use it with care.
        * the name of a particular full node (e.g., `validator01`): retrieves the blockstore from that node only.
8. Verify that the data was collected without errors
    * at least one blockstore DB for a CometBFT validator
    * the Prometheus database from the Prometheus node
    * for extra care, you can run `zip -T` on the `prometheus.zip` file and (one of) the `blockstore.db.zip` file(s)
9. **Run `make terraform-destroy`**
    * Don't forget to type `yes`! Otherwise you're in trouble.

### Result Extraction

The method for extracting the results described here is highly manual (and exploratory) at this stage.
The CometBFT team should improve it at every iteration to increase the amount of automation.

#### Steps

1. Unzip the blockstore into a directory
2. To identify saturation points
   1. Extract the latency report for all the experiments.
       * Run these commands from the directory containing the `blockstore.db` folder.
       * It is advisable to adjust the hash in the `go run` command to the latest possible.
       * ```bash
         mkdir results
         go run github.com/cometbft/cometbft/test/loadtime/cmd/report@3003ef7 --database-type goleveldb --data-dir ./ > results/report.txt
         ```
   2. File `report.txt` contains an unordered list of experiments with varying concurrent connections and transaction rate.
      You will need to separate data per experiment.

        * Create files `report01.txt`, `report02.txt`, `report04.txt` and, for each experiment in file `report.txt`,
          copy its related lines to the filename that matches the number of connections, for example

          ```bash
          for cnum in 1 2 4; do echo "$cnum"; grep "Connections: $cnum" results/report.txt -B 2 -A 10 > results/report$cnum.txt;  done
          ```

        * Sort the experiments in `report01.txt` in ascending tx rate order. Likewise for `report02.txt` and `report04.txt`.
        * Otherwise just keep `report.txt`, and skip to the next step.
    4. Generate file `report_tabbed.txt` by showing the contents `report01.txt`, `report02.txt`, `report04.txt` side by side
        * This effectively creates a table where rows are a particular tx rate and columns are a particular number of websocket connections.
        * Combine the column files into a single table file:
           * Replace tabs by spaces in all column files. For example,
             `sed -i.bak 's/\t/    /g' results/report1.txt`.
        * Merge the new column files into one:
           `paste results/report1.txt results/report2.txt results/report4.txt | column -s $'\t' -t > report_tabbed.txt`

3. To generate a latency vs throughput plot, extract the data as a CSV
    * ```bash
       go run github.com/cometbft/cometbft/test/loadtime/cmd/report@3003ef7 --database-type goleveldb --data-dir ./ --csv results/raw.csv
       ```
    * Follow the instructions for the [`latency_throughput.py`] script.
    This plot is useful to visualize the saturation point.
    * Alternatively,  follow the instructions for the [`latency_plotter.py`] script.
    This script generates a series of plots per experiment and configuration that may
    help with visualizing Latency vs Throughput variation.

[`latency_throughput.py`]: https://github.com/cometbft/cometbft/tree/v0.38.x/scripts/qa/reporting#latency-vs-throughput-plotting
[`latency_plotter.py`]: https://github.com/cometbft/cometbft/tree/v0.38.x/scripts/qa/reporting#latency-vs-throughput-plotting-version-2

#### Extracting Prometheus Metrics

1. Stop the prometheus server if it is running as a service (e.g. a `systemd` unit).
2. Unzip the prometheus database retrieved from the testnet, and move it to replace the
   local prometheus database.
3. Start the prometheus server and make sure no error logs appear at start up.
4. Identify the time window you want to plot in your graphs.
5. Execute the [`prometheus_plotter.py`] script for the time window.

[`prometheus_plotter.py`]: https://github.com/cometbft/cometbft/tree/v0.38.x/scripts/qa/reporting#prometheus-metrics

## Rotating Node Testnet

### Running the test

This section explains how the tests were carried out for reproducibility purposes.

1. [If you haven't done it before]
   Follow steps 1-4 of the `README.md` at the top of the testnet repository to configure Terraform, and `doctl`.
2. Copy file `testnet_rotating.toml` onto `testnet.toml` (do NOT commit this change)
3. Set variable `VERSION_TAG` to the git hash that is to be tested.
4. Run `make terraform-apply EPHEMERAL_SIZE=25`
    * WARNING: Do NOT forget to run `make terraform-destroy` as soon as you are done with the tests
5. Follow steps 6-10 of the `README.md` to configure and start the "stable" part of the rotating node testnet
6. As a sanity check, connect to the Prometheus node's web interface and check the graph for the `tendermint_consensus_height` metric.
   All nodes should be increasing their heights.
7. On a different shell,
    * run `make runload LOAD_CONNECTIONS=X LOAD_TX_RATE=Y LOAD_TOTAL_TIME=Z`
    * `X` and `Y` should reflect a load below the saturation point (see, e.g.,
      [this paragraph](./TMCore-QA-34.md#finding-the-saturation-point) for further info)
    * `Z` (in seconds) should be big enough to keep running throughout the test, until we manually stop it in step 9.
      In principle, a good value for `Z` is `7200` (2 hours)
8. Run `make rotate` to start the script that creates the ephemeral nodes, and kills them when they are caught up.
    * WARNING: If you run this command from your laptop, the laptop needs to be up and connected for the full length
      of the experiment.
    * [This](http://<PROMETHEUS-NODE-IP>:9090/classic/graph?g0.range_input=100m&g0.expr=cometbft_consensus_height%7Bjob%3D~%22ephemeral.*%22%7D%20or%20cometbft_blocksync_latest_block_height%7Bjob%3D~%22ephemeral.*%22%7D&g0.tab=0&g1.range_input=100m&g1.expr=cometbft_mempool_size%7Bjob!~%22ephemeral.*%22%7D&g1.tab=0&g2.range_input=100m&g2.expr=cometbft_consensus_num_txs%7Bjob!~%22ephemeral.*%22%7D&g2.tab=0)
      is an example Prometheus URL you can use to monitor the test case's progress
9. When the height of the chain reaches 3000, stop the `make runload` script.
10. When the rotate script has made two iterations (i.e., all ephemeral nodes have caught up twice)
    after height 3000 was reached, stop `make rotate`
11. Run `make stop-network`
12. Run `make retrieve-data` to gather all relevant data from the testnet into the orchestrating machine
13. Verify that the data was collected without errors
    * at least one blockstore DB for a CometBFT validator
    * the Prometheus database from the Prometheus node
    * for extra care, you can run `zip -T` on the `prometheus.zip` file and (one of) the `blockstore.db.zip` file(s)
14. **Run `make terraform-destroy`**

Steps 8 to 10 are highly manual at the moment and will be improved in next iterations.

### Result Extraction

In order to obtain a latency plot, follow the instructions above for the 200 node experiment,
but the `results.txt` file contains only one experiment.

As for prometheus, the same method as for the 200 node experiment can be applied.

## Vote Extensions Testnet

### Running the test

This section explains how the tests were carried out for reproducibility purposes.

1. [If you haven't done it before]
   Follow steps 1-4 of the `README.md` at the top of the testnet repository to configure Terraform, and `doctl`.
2. Copy file `varyVESize.toml` onto `testnet.toml` (do NOT commit this change).
3. Set variable `VERSION_TAG` in the `Makefile` to the git hash that is to be tested.
4. Follow steps 5-10 of the `README.md` to configure and start the testnet
    * WARNING: Do NOT forget to run `make terraform-destroy` as soon as you are done with the tests
5. Configure the load runner to produce the desired transaction load.
    * set makefile variables `ROTATE_CONNECTIONS`, `ROTATE_TX_RATE`, to values that will produce the desired transaction load.
    * set `ROTATE_TOTAL_TIME` to 150 (seconds).
    * set `ITERATIONS` to the number of iterations that each configuration should run for.
6. Execute steps 5-10 of the `README.md` file at the testnet repository.

7. Repeat the following steps for each desired `vote_extension_size`
    1. Update the configuration (you can skip this step if you didn't change the `vote_extension_size`)
        * Update the `vote_extensions_size` in the `testnet.toml` to the desired value.
        * `make configgen`
        * `ANSIBLE_SSH_RETRIES=10 ansible-playbook ./ansible/re-init-testapp.yaml -u root -i ./ansible/hosts --limit=validators -e "testnet_dir=testnet" -f 20`
        * `make restart`
    2. Run the test
        * `make runload`
          This will repeat the tests `ITERATIONS` times every time it is invoked.
    3. Collect your data
        * `make retrieve-data`
          Gathers all relevant data from the testnet into the orchestrating machine, inside folder `experiments`.
          Two subfolders are created, one blockstore DB for a CometBFT validator and one for the Prometheus DB data.
        * Verify that the data was collected without errors with `zip -T` on the `prometheus.zip` file and (one of) the `blockstore.db.zip` file(s).
8. Clean up your setup.
    * `make terraform-destroy`; don't forget that you need to type **yes** for it to complete.


### Result Extraction

In order to obtain a latency plot, follow the instructions above for the 200 node experiment, but:

* The `results.txt` file contains only one experiment
* Therefore, no need for any `for` loops

As for Prometheus, the same method as for the 200 node experiment can be applied.
