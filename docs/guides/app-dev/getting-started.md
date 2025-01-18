---
order: 2
---

# Getting Started

## First CometBFT App

As a general purpose blockchain engine, CometBFT is agnostic to the
application you want to run. So, to run a complete blockchain that does
something useful, you must start two programs: one is CometBFT,
the other is your application, which can be written in any programming
language. Recall from [the intro to
ABCI](../explanation/introduction/what-is-cometbft.md#abci-overview) that CometBFT
handles all the p2p and consensus stuff, and just forwards transactions to the
application when they need to be validated, or when they're ready to be
executed and committed.

In this guide, we show you some examples of how to run an application
using CometBFT.

### Install

The first apps we will work with are written in Go. To install them, you
need to [install Go](https://golang.org/doc/install), put
`$GOPATH/bin` in your `$PATH` and enable go modules. If you use `bash`,
follow these instructions:

```bash
echo export GOPATH=\"\$HOME/go\" >> ~/.bash_profile
echo export PATH=\"\$PATH:\$GOPATH/bin\" >> ~/.bash_profile
```

Then run

```bash
go get github.com/cometbft/cometbft
cd $GOPATH/src/github.com/cometbft/cometbft
make install_abci
```

Now you should have the `abci-cli` installed; run `abci-cli` to see the list of commands:

```
Usage:
  abci-cli [command]

Available Commands:
  batch            run a batch of abci commands against an application
  check_tx         validate a transaction
  commit           commit the application state and return the Merkle root hash
  completion       Generate the autocompletion script for the specified shell
  console          start an interactive ABCI console for multiple commands
  echo             have the application echo a message
  finalize_block   deliver a block of transactions to the application
  help             Help about any command
  info             get some info about the application
  kvstore          ABCI demo example
  prepare_proposal prepare proposal
  process_proposal process proposal
  query            query the application state
  test             run integration tests
  version          print ABCI console version

Flags:
      --abci string        either socket or grpc (default "socket")
      --address string     address of application socket (default "tcp://0.0.0.0:26658")
  -h, --help               help for abci-cli
      --log_level string   set the logger level (default "debug")
  -v, --verbose            print the command and results as if it were a console session

Use "abci-cli [command] --help" for more information about a command.
```

You'll notice the `kvstore` command, an example application written in Go.

Now, let's run an app!

## KVStore - A First Example

The kvstore app is a [Merkle
tree](https://en.wikipedia.org/wiki/Merkle_tree) that just stores all
transactions. If the transaction contains an `=`, e.g. `key=value`, then
the `value` is stored under the `key` in the Merkle tree. Otherwise, the
full transaction bytes are stored as the key and the value.

Let's start a kvstore application.

```sh
abci-cli kvstore
```

In another terminal, we can start CometBFT. You should already have the
CometBFT binary installed. If not, follow the steps from
[here](../explanation/introduction/install.md). If you have never run CometBFT
before, use:

```sh
cometbft init
cometbft node
```

If you have used CometBFT, you may want to reset the data for a new
blockchain by running `cometbft unsafe-reset-all`. Then you can run
`cometbft node` to start CometBFT, and connect to the app. For more
details, see [the guide on using CometBFT](../../explanation/core/using-cometbft.md).

You should see CometBFT making blocks! We can get the status of our
CometBFT node as follows:

```sh
curl -s localhost:26657/status
```

The `-s` just silences `curl`. For nicer output, pipe the result into a
tool like [jq](https://stedolan.github.io/jq/) or `json_pp`.

Now let's send some transactions to the kvstore.

```sh
curl -s 'localhost:26657/broadcast_tx_commit?tx="name=satoshi"'
```

Now if we query for `name`, we should get `satoshi`, or `c2F0b3NoaQ==`
in base64:

```sh
curl -s 'localhost:26657/abci_query?data="name"'
```

The response should look something like:

```json
{
  "jsonrpc": "2.0",
  "id": -1,
  "result": {
    "response": {
      "code": 0,
      "log": "exists",
      "info": "",
      "index": "0",
      "key": "bmFtZQ==",
      "value": "c2F0b3NoaQ==",
      "proofOps": null,
      "height": "20795",
      "codespace": ""
    }
  }
}
```

Note the value in the result (`c2F0b3NoaQ==`); this is the base64-encoding of the ASCII of `satoshi`. You can verify this through a simple Bash command:

```bash
echo -n 'c2F0b3NoaQ==' | base64 -d
```

Stay tuned for a future release that [makes this output more human-readable](https://github.com/tendermint/tendermint/issues/1794).

Try some other transactions and queries to make sure everything is
working!
