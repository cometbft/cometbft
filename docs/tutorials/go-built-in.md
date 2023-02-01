---
order: 2
---

# Creating a built-in application in Go

## Guide Assumptions

This guide is designed for beginners who want to get started with a CometBFT
application from scratch. It does not assume that you have any prior
experience with CometBFT.

<<<<<<< HEAD
Tendermint Core is Byzantine Fault Tolerant (BFT) middleware that takes a state
transition machine - written in any programming language - and securely
replicates it on many machines.

Although Tendermint Core is written in the Golang programming language, prior
knowledge of it is not required for this guide. You can learn it as we go due
to it's simplicity. However, you may want to go through [Learn X in Y minutes
Where X=Go](https://learnxinyminutes.com/docs/go/) first to familiarize
yourself with the syntax.

By following along with this guide, you'll create a Tendermint Core project
called kvstore, a (very) simple distributed BFT key-value store.

## Built-in app vs external app

Running your application inside the same process as Tendermint Core will give
you the best possible performance.
=======
CometBFT is a service that provides a Byzantine Fault Tolerant consensus engine
for state-machine replication. The replicated state-machine, or "application", can be written
in any language that can send and receive protocol buffer messages in a client-server model.
Applications written in Go can also use CometBFT as a library and run the service in the same
process as the application.

By following along this tutorial you will create a CometBFT application called kvstore,
a (very) simple distributed BFT key-value store.
The application will be written in Go and  
some understanding of the Go programming language is expected.
If you have never written Go, you may want to go through [Learn X in Y minutes
Where X=Go](https://learnxinyminutes.com/docs/go/) first, to familiarize
yourself with the syntax.

Note: Please use the latest released version of this guide and of CometBFT.
We strongly advise against using unreleased commits for your development.

### Built-in app vs external app

On the one hand, to get maximum performance you can run your application in
the same process as the CometBFT, as long as your application is written in Go.
[Cosmos SDK](https://github.com/cosmos/cosmos-sdk) is written
this way.
This is the approach followed in this tutorial.

On the other hand, having a separate application might give you better security
guarantees as two processes would be communicating via established binary protocol.
CometBFT will not have access to application's state.
If that is the way you wish to proceed, use the [Creating an application in Go](./go.md) guide instead of this one.
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

For other languages, your application have to communicate with Tendermint Core
through a TCP, Unix domain socket or gRPC.

## 1.1 Installing Go

Please refer to [the official guide for installing
Go](https://golang.org/doc/install).

Verify that you have the latest version of Go installed:

```bash
$ go version
go version go1.13.1 darwin/amd64
```

Make sure you have `$GOPATH` environment variable set:

```bash
$ echo $GOPATH
/Users/melekes/go
```

## 1.2 Creating a new Go project

We'll start by creating a new Go project.

```bash
mkdir kvstore
cd kvstore
```

Inside the example directory create a `main.go` file with the following content:

```go
package main

import (
	"fmt"
)

func main() {
<<<<<<< HEAD
	fmt.Println("Hello, Tendermint Core")
=======
    fmt.Println("Hello, CometBFT")
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))
}
```

When run, this should print "Hello, CometBFT" to the standard output.

```bash
$ go run main.go
Hello, CometBFT
```

<<<<<<< HEAD
## 1.3 Writing a Tendermint Core application

Tendermint Core communicates with the application through the Application
BlockChain Interface (ABCI). All message types are defined in the [protobuf
file](https://github.com/tendermint/tendermint/blob/v0.34.x/proto/tendermint/abci/types.proto).
This allows Tendermint Core to run applications written in any programming
language.
=======
We are going to use [Go modules](https://github.com/golang/go/wiki/Modules) for
dependency management, so let's start by including a dependency on the latest version of
CometBFT, `v0.37.0` in this example.

```bash
go mod init kvstore
go get github.com/cometbft/cometbft@v0.37.0
```

After running the above commands you will see two generated files, `go.mod` and `go.sum`.
The go.mod file should look similar to:

```go
module github.com/me/example

go 1.19

require (
	github.com/cometbft/cometbft v0.37.0
)
```

As you write the kvstore application, you can rebuild the binary by
pulling any new dependencies and recompiling it.

```sh
go get
go build
```

## 1.3 Writing a CometBFT application

CometBFT communicates with the application through the Application
BlockChain Interface (ABCI). The messages exchanged through the interface are
defined in the ABCI [protobuf
file](https://github.com/cometbft/cometbft/blob/v0.37.x/proto/tendermint/abci/types.proto).
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

Create a file called `app.go` with the following content:

```go
package main

import (
	abcitypes "github.com/cometbft/cometbft/abci/types"
)

type KVStoreApplication struct {}

var _ abcitypes.Application = (*KVStoreApplication)(nil)

func NewKVStoreApplication() *KVStoreApplication {
	return &KVStoreApplication{}
}

func (KVStoreApplication) Info(req abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{}
}

func (KVStoreApplication) SetOption(req abcitypes.RequestSetOption) abcitypes.ResponseSetOption {
	return abcitypes.ResponseSetOption{}
}

func (KVStoreApplication) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	return abcitypes.ResponseDeliverTx{Code: 0}
}

func (KVStoreApplication) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	return abcitypes.ResponseCheckTx{Code: 0}
}

func (KVStoreApplication) Commit() abcitypes.ResponseCommit {
	return abcitypes.ResponseCommit{}
}

func (KVStoreApplication) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	return abcitypes.ResponseQuery{Code: 0}
}

func (KVStoreApplication) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	return abcitypes.ResponseInitChain{}
}

func (KVStoreApplication) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	return abcitypes.ResponseBeginBlock{}
}

func (KVStoreApplication) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{}
}

func (KVStoreApplication) ListSnapshots(abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

func (KVStoreApplication) OfferSnapshot(abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

func (KVStoreApplication) LoadSnapshotChunk(abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

func (KVStoreApplication) ApplySnapshotChunk(abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}
```

<<<<<<< HEAD
Now I will go through each method explaining when it's called and adding
required business logic.

### 1.3.1 CheckTx
=======
The types used here are defined in the CometBFT library and were added as a dependency
to the project when you ran `go get`. If your IDE is not recognizing the types, go ahead and run the command again.

```bash
go get github.com/cometbft/cometbft@v0.37.0
```
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

When a new transaction is added to the Tendermint Core, it will ask the
application to check it (validate the format, signatures, etc.).

```go
<<<<<<< HEAD
import "bytes"

func (app *KVStoreApplication) isValid(tx []byte) (code uint32) {
=======
func main() {
    fmt.Println("Hello, CometBFT")

    _ = NewKVStoreApplication()
}
```

You can recompile and run the application now by running `go get` and `go build`, but it does
not do anything.
So let's revisit the code adding the logic needed to implement our minimal key/value store
and to start it along with the CometBFT Service.


### 1.3.1 Add a persistent data store

Our application will need to write its state out to persistent storage so that it
can stop and start without losing all of its data.

For this tutorial, we will use [BadgerDB](https://github.com/dgraph-io/badger), a
a fast embedded key-value store.

First, add Badger as a dependency of your go module using the `go get` command:

`go get github.com/dgraph-io/badger/v3`

Next, let's update the application and its constructor to receive a handle to the database, as follows:

```go
type KVStoreApplication struct {
	db           *badger.DB
	onGoingBlock *badger.Txn
}

var _ abcitypes.Application = (*KVStoreApplication)(nil)

func NewKVStoreApplication(db *badger.DB) *KVStoreApplication {
	return &KVStoreApplication{db: db}
}
```

The `onGoingBlock` keeps track of the Badger transaction that will update the application's state when a block
is completed. Don't worry about it for now, we'll get to that later.

Next, update the `import` stanza at the top to include the Badger library:

```go
import(
	"github.com/dgraph-io/badger/v3"
	abcitypes "github.com/cometbft/cometbft/abci/types"
)
```

Finally, update the `main.go` file to invoke the updated constructor:

```go
	_ = NewKVStoreApplication(nil)
```

### 1.3.2 CheckTx

When CometBFT receives a new transaction from a client, or from another full node,
CometBFT asks the application if the transaction is acceptable, using the `CheckTx` method.
Invalid transactions will not be shared with other nodes and will not become part of any blocks and, therefore, will not be executed by the application.

In our application, a transaction is a string with the form `key=value`, indicating a key and value to write to the store.

The most basic validation check we can perform is to check if the transaction conforms to the `key=value` pattern.
For that, let's add the following helper method to app.go:

```go
func (app *KVStoreApplication) isValid(tx []byte) uint32 {
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))
	// check format
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return 1
	}

	key, value := parts[0], parts[1]

	// check if the same key=value already exists
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == nil {
			return item.Value(func(val []byte) error {
				if bytes.Equal(val, value) {
					code = 2
				}
				return nil
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return code
}

func (app *KVStoreApplication) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	code := app.isValid(req.Tx)
	return abcitypes.ResponseCheckTx{Code: code, GasWanted: 1}
}
```

Don't worry if this does not compile yet.

<<<<<<< HEAD
If the transaction does not have a form of `{bytes}={bytes}`, we return `1`
code. When the same key=value already exist (same key and value), we return `2`
code. For others, we return a zero code indicating that they are valid.
=======
Depending on the checks and on the conditions violated, the function may return
different values, but any response with a non-zero code will be considered invalid
by CometBFT. Our `CheckTx` logic returns 0 to CometBFT when a transaction passes
its validation checks. The specific value of the code is meaningless to CometBFT.
Non-zero codes are logged by CometBFT so applications can provide more specific
information on why the transaction was rejected.
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

Note that anything with non-zero code will be considered invalid (`-1`, `100`,
etc.) by Tendermint Core.

Valid transactions will eventually be committed given they are not too big and
have enough gas. To learn more about gas, check out ["the
specification"](https://github.com/tendermint/tendermint/blob/v0.34.x/spec/abci/apps.md#gas).

For the underlying key-value store we'll use
[badger](https://github.com/dgraph-io/badger), which is an embeddable,
persistent and fast key-value (KV) database.

```go
import "github.com/dgraph-io/badger"

<<<<<<< HEAD
type KVStoreApplication struct {
	db           *badger.DB
	currentBatch *badger.Txn
}

func NewKVStoreApplication(db *badger.DB) *KVStoreApplication {
	return &KVStoreApplication{
		db: db,
	}
}
=======
	"github.com/dgraph-io/badger/v3"
	abcitypes "github.com/cometbft/cometbft/abci/types"
)
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))
```

### 1.3.2 BeginBlock -> DeliverTx -> EndBlock -> Commit

<<<<<<< HEAD
When Tendermint Core has decided on the block, it's transfered to the
application in 3 parts: `BeginBlock`, one `DeliverTx` per transaction and
`EndBlock` in the end. DeliverTx are being transfered asynchronously, but the
responses are expected to come in order.
=======
### 1.3.3 BeginBlock -> DeliverTx -> EndBlock -> Commit

When the CometBFT consensus engine has decided on the block, the block is transferred to the
application over three ABCI method calls: `BeginBlock`, `DeliverTx`, and `EndBlock`.

- `BeginBlock` is called once to indicate to the application that it is about to
receive a block.
- `DeliverTx` is called repeatedly, once for each application transaction that was included in the block.
- `EndBlock` is called once to indicate to the application that no more transactions
will be delivered to the application in within this block.

Note that, to implement these calls in our application we're going to make use of Badger's
transaction mechanism. We will always refer to these as Badger transactions, not to
confuse them with the transactions included in the blocks delivered by CometBFT,
the _application transactions_.

First, let's create a new Badger transaction during `BeginBlock`. All application transactions in the
current block will be executed within this Badger transaction.
Then, return informing CometBFT that the application is ready to receive application transactions:
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

```go
func (app *KVStoreApplication) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	app.currentBatch = app.db.NewTransaction(true)
	return abcitypes.ResponseBeginBlock{}
}

```

Here we create a batch, which will store block's transactions.

```go
func (app *KVStoreApplication) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	code := app.isValid(req.Tx)
	if code != 0 {
		return abcitypes.ResponseDeliverTx{Code: code}
	}

	parts := bytes.Split(req.Tx, []byte("="))
	key, value := parts[0], parts[1]

<<<<<<< HEAD
	err := app.currentBatch.Set(key, value)
	if err != nil {
		panic(err)
=======
	if err := app.onGoingBlock.Set(key, value); err != nil {
		log.Panicf("Error writing to database, unable to execute tx: %v", err)
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))
	}

	return abcitypes.ResponseDeliverTx{Code: 0}
}
```

If the transaction is badly formatted or the same key=value already exist, we
again return the non-zero code. Otherwise, we add it to the current batch.

In the current design, a block can include incorrect transactions (those who
passed CheckTx, but failed DeliverTx or transactions included by the proposer
directly). This is done for performance reasons.

Note we can't commit transactions inside the `DeliverTx` because in such case
`Query`, which may be called in parallel, will return inconsistent data (i.e.
it will report that some value already exist even when the actual block was not
yet committed).

`Commit` instructs the application to persist the new state.

```go
func (app *KVStoreApplication) Commit() abcitypes.ResponseCommit {
	app.currentBatch.Commit()
	return abcitypes.ResponseCommit{Data: []byte{}}
}
```

### 1.3.3 Query

Now, when the client wants to know whenever a particular key/value exist, it
will call Tendermint Core RPC `/abci_query` endpoint, which in turn will call
the application's `Query` method.

Applications are free to provide their own APIs. But by using Tendermint Core
as a proxy, clients (including [light client
package](https://godoc.org/github.com/tendermint/tendermint/light)) can leverage
the unified API across different applications. Plus they won't have to call the
otherwise separate Tendermint Core API for additional proofs.

Note we don't include a proof here.

```go
<<<<<<< HEAD
func (app *KVStoreApplication) Query(reqQuery abcitypes.RequestQuery) (resQuery abcitypes.ResponseQuery) {
	resQuery.Key = reqQuery.Data
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(reqQuery.Data)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
=======
import (
	"bytes"
	"log"

	"github.com/dgraph-io/badger/v3"
	abcitypes "github.com/cometbft/cometbft/abci/types"
)
```

You may have noticed that the application we are writing will crash if it receives
an unexpected error from the Badger database during the `DeliverTx` or `Commit` methods.
This is not an accident. If the application received an error from the database, there
is no deterministic way for it to make progress so the only safe option is to terminate.

### 1.3.4 Query

When a client tries to read some information from the `kvstore`, the request will be
handled in the `Query` method. To do this, let's rewrite the `Query` method in `app.go`:

```go
func (app *KVStoreApplication) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	resp := abcitypes.ResponseQuery{Key: req.Data}

	dbErr := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(req.Data)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			resp.Log = "key does not exist"
			return nil
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))
		}
		if err == badger.ErrKeyNotFound {
			resQuery.Log = "does not exist"
		} else {
			return item.Value(func(val []byte) error {
				resQuery.Log = "exists"
				resQuery.Value = val
				return nil
			})
		}
		return nil
	})
<<<<<<< HEAD
	if err != nil {
		panic(err)
=======
	if dbErr != nil {
		log.Panicf("Error reading database, unable to execute query: %v", dbErr)
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))
	}
	return
}
```

<<<<<<< HEAD
The complete specification can be found
[here](https://github.com/tendermint/tendermint/tree/v0.34.x/spec/abci/).
=======
Since it reads only committed data from the store, transactions that are part of a block
that is being processed are not reflected in the query result.

### 1.3.5 PrepareProposal and ProcessProposal

`PrepareProposal` and `ProcessProposal` are methods introduced in CometBFT v0.37.0
to give the application more control over the construction and processing of transaction blocks.

When CometBFT sees that valid transactions (validated through `CheckTx`) are available to be
included in blocks, it groups some of these transactions and then gives the application a chance
to modify the group by invoking `PrepareProposal`.

The application is free to modify the group before returning from the call, as long as the resulting set
does not use more bytes than `RequestPrepareProposal.max_tx_bytes'
For example, the application may reorder, add, or even remove transactions from the group to improve the
execution of the block once accepted.
In the following code, the application simply returns the unmodified group of transactions:

```go
func (app *KVStoreApplication) PrepareProposal(proposal abcitypes.RequestPrepareProposal) abcitypes.ResponsePrepareProposal {
	return abcitypes.ResponsePrepareProposal{Txs: proposal.Txs}
}
```

Once a proposed block is received by a node, the proposal is passed to the application to give
its blessing before voting to accept the proposal.

This mechanism may be used for different reasons, for example to deal with blocks manipulated
by malicious nodes, in which case the block should not be considered valid.
The following code simply accepts all proposals:

```go
func (app *KVStoreApplication) ProcessProposal(proposal abcitypes.RequestProcessProposal) abcitypes.ResponseProcessProposal {
	return abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_ACCEPT}
}
```
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

## 1.4 Starting an application and a CometBFT instance in the same process

<<<<<<< HEAD
Put the following code into the "main.go" file:
=======
Now that we have the basic functionality of our application in place, let's put it all together inside of our main.go file.

Change the contents of your `main.go` file to the following.
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

```go
package main

import (
<<<<<<< HEAD
 "flag"
 "fmt"
 "os"
 "os/signal"
 "path/filepath"
 "syscall"

 "github.com/dgraph-io/badger"
 "github.com/spf13/viper"

 abci "github.com/tendermint/tendermint/abci/types"
 cfg "github.com/tendermint/tendermint/config"
 cmtflags "github.com/tendermint/tendermint/libs/cli/flags"
 "github.com/tendermint/tendermint/libs/log"
 nm "github.com/tendermint/tendermint/node"
 "github.com/tendermint/tendermint/p2p"
 "github.com/tendermint/tendermint/privval"
 "github.com/tendermint/tendermint/proxy"
=======
	"flag"
	"fmt"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
	cfg "github.com/cometbft/cometbft/config"
	cmtflags "github.com/cometbft/cometbft/libs/cli/flags"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	nm "github.com/cometbft/cometbft/node"
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))
)

var configFile string

func init() {
<<<<<<< HEAD
	flag.StringVar(&configFile, "config", "$HOME/.tendermint/config/config.toml", "Path to config.toml")
}

func main() {
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
=======
	flag.StringVar(&homeDir, "cmt-home", "", "Path to the CometBFT config directory (if empty, uses $HOME/.cometbft)")
}

func main() {
	flag.Parse()
	if homeDir == "" {
		homeDir = os.ExpandEnv("$HOME/.cometbft")
	}
	config := cfg.DefaultConfig()

	config.SetRoot(homeDir)

	viper.SetConfigFile(fmt.Sprintf("%s/%s", homeDir, "config/config.toml"))
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Reading config: %v", err)
	}
	if err := viper.Unmarshal(config); err != nil {
		log.Fatalf("Decoding config: %v", err)
	}
	if err := config.ValidateBasic(); err != nil {
		log.Fatalf("Invalid configuration data: %v", err)
	}

	dbPath := filepath.Join(homeDir, "badger")
	db, err := badger.Open(badger.DefaultOptions(dbPath))
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open badger db: %v", err)
		os.Exit(1)
	}
	defer db.Close()
	app := NewKVStoreApplication(db)

	flag.Parse()

	node, err := newTendermint(app, configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(2)
	}

	node.Start()
	defer func() {
		node.Stop()
		node.Wait()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	os.Exit(0)
}

func newTendermint(app abci.Application, configFile string) (*nm.Node, error) {
 // read config
 config := cfg.DefaultConfig()
 config.RootDir = filepath.Dir(filepath.Dir(configFile))
 viper.SetConfigFile(configFile)
 if err := viper.ReadInConfig(); err != nil {
  return nil, fmt.Errorf("viper failed to read config file: %w", err)
 }
 if err := viper.Unmarshal(config); err != nil {
  return nil, fmt.Errorf("viper failed to unmarshal config: %w", err)
 }
 if err := config.ValidateBasic(); err != nil {
  return nil, fmt.Errorf("config is invalid: %w", err)
 }

 // create logger
 logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
 var err error
 logger, err = cmtflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel)
 if err != nil {
  return nil, fmt.Errorf("failed to parse log level: %w", err)
 }

 // read private validator
 pv := privval.LoadFilePV(
  config.PrivValidatorKeyFile(),
  config.PrivValidatorStateFile(),
 )

 // read node key
 nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
 if err != nil {
  return nil, fmt.Errorf("failed to load node's key: %w", err)
 }

 // create node
 node, err := nm.NewNode(
  config,
  pv,
  nodeKey,
  proxy.NewLocalClientCreator(app),
  nm.DefaultGenesisDocProviderFunc(config),
  nm.DefaultDBProvider,
  nm.DefaultMetricsProvider(config.Instrumentation),
  logger)
 if err != nil {
  return nil, fmt.Errorf("failed to create new Tendermint node: %w", err)
 }

 return node, nil
}
```

This is a huge blob of code, so let's break it down into pieces.

<<<<<<< HEAD
First, we initialize the Badger database and create an app instance:
=======
First, we use [viper](https://github.com/spf13/viper) to load the CometBFT configuration files, which we will generate later:

>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

```go
db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
if err != nil {
	fmt.Fprintf(os.Stderr, "failed to open badger db: %v", err)
	os.Exit(1)
}
defer db.Close()
app := NewKVStoreApplication(db)
```

For **Windows** users, restarting this app will make badger throw an error as it requires value log to be truncated. For more information on this, visit [here](https://github.com/dgraph-io/badger/issues/744).
This can be avoided by setting the truncate option to true, like this:

```go
db, err := badger.Open(badger.DefaultOptions("/tmp/badger").WithTruncate(true))
```

Then we use it to create a Tendermint Core `Node` instance:

```go
flag.Parse()

node, err := newTendermint(app, configFile)
if err != nil {
	fmt.Fprintf(os.Stderr, "%v", err)
	os.Exit(2)
}

...

// create node
node, err := nm.NewNode(
	config,
	pv,
	nodeKey,
	proxy.NewLocalClientCreator(app),
	nm.DefaultGenesisDocProviderFunc(config),
	nm.DefaultDBProvider,
	nm.DefaultMetricsProvider(config.Instrumentation),
	logger)
if err != nil {
	return nil, fmt.Errorf("failed to create new Tendermint node: %w", err)
}
```

`NewNode` requires a few things including a configuration file, a private
validator, a node key and a few others in order to construct the full node.

Note we use `proxy.NewLocalClientCreator` here to create a local client instead
of one communicating through a socket or gRPC.

[viper](https://github.com/spf13/viper) is being used for reading the config,
which we will generate later using the `tendermint init` command.

```go
config := cfg.DefaultConfig()
config.RootDir = filepath.Dir(filepath.Dir(configFile))
viper.SetConfigFile(configFile)
if err := viper.ReadInConfig(); err != nil {
	return nil, fmt.Errorf("viper failed to read config file: %w", err)
}
if err := viper.Unmarshal(config); err != nil {
	return nil, fmt.Errorf("viper failed to unmarshal config: %w", err)
}
if err := config.ValidateBasic(); err != nil {
	return nil, fmt.Errorf("config is invalid: %w", err)
}
```

We use `FilePV`, which is a private validator (i.e. thing which signs consensus
messages). Normally, you would use `SignerRemote` to connect to an external
[HSM](https://kb.certus.one/hsm.html).

```go
pv := privval.LoadFilePV(
	config.PrivValidatorKeyFile(),
	config.PrivValidatorStateFile(),
)

```

`nodeKey` is needed to identify the node in a p2p network.

```go
nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
if err != nil {
	return nil, fmt.Errorf("failed to load node's key: %w", err)
}
```

<<<<<<< HEAD
As for the logger, we use the build-in library, which provides a nice
abstraction over [go-kit's
logger](https://github.com/go-kit/kit/tree/master/log).
=======
Now we have everything set up to run the CometBFT node. We construct
a node by passing it the configuration, the logger, a handle to our application and
the genesis information:
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

```go
logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
var err error
logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel())
if err != nil {
	return nil, fmt.Errorf("failed to parse log level: %w", err)
}
```

<<<<<<< HEAD
Finally, we start the node and add some signal handling to gracefully stop it
upon receiving SIGTERM or Ctrl-C.
=======
Finally, we start the node, i.e., the CometBFT service inside our application:
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

```go
node.Start()
defer func() {
	node.Stop()
	node.Wait()
}()

c := make(chan os.Signal, 1)
signal.Notify(c, os.Interrupt, syscall.SIGTERM)
<-c
os.Exit(0)
```

## 1.5 Getting Up and Running

We are going to use [Go modules](https://github.com/golang/go/wiki/Modules) for
dependency management.

```bash
go mod init github.com/me/example
go get github.com/tendermint/tendermint/@v0.34.0
```

After running the above commands you will see two generated files, go.mod and go.sum. The go.mod file should look similar to:

```go
module github.com/me/example

go 1.15

require (
	github.com/dgraph-io/badger v1.6.2
	github.com/tendermint/tendermint v0.34.0
)
```

Finally, we will build our binary:

<<<<<<< HEAD
```sh
go build
```
=======
Our application is almost ready to run, but first we'll need to populate the CometBFT configuration files.
The following command will create a `cometbft-home` directory in your project and add a basic set of configuration files in `cometbft-home/config/`.
For more information on what these files contain see [the configuration documentation](https://github.com/cometbft/cometbft/blob/v0.37.x/docs/core/configuration.md).
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

To create a default configuration, nodeKey and private validator files, let's
execute `tendermint init`. But before we do that, we will need to install
Tendermint Core. Please refer to [the official
guide](https://docs.tendermint.com/v0.34/introduction/install.html). If you're
installing from source, don't forget to checkout the latest release (`git checkout vX.Y.Z`).

```bash
<<<<<<< HEAD
$ rm -rf /tmp/example
$ TMHOME="/tmp/example" tendermint init

I[2019-07-16|18:40:36.480] Generated private validator                  module=main keyFile=/tmp/example/config/priv_validator_key.json stateFile=/tmp/example2/data/priv_validator_state.json
I[2019-07-16|18:40:36.481] Generated node key                           module=main path=/tmp/example/config/node_key.json
I[2019-07-16|18:40:36.482] Generated genesis file                       module=main path=/tmp/example/config/genesis.json
=======
go run github.com/cometbft/cometbft/cmd/cometbft@v0.37.0 init --home /tmp/cometbft-home
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))
```

We are ready to start our application:

```bash
<<<<<<< HEAD
$ ./example -config "/tmp/example/config/config.toml"

badger 2019/07/16 18:42:25 INFO: All 0 tables opened in 0s
badger 2019/07/16 18:42:25 INFO: Replaying file id: 0 at offset: 0
badger 2019/07/16 18:42:25 INFO: Replay took: 695.227s
E[2019-07-16|18:42:25.818] Couldn't connect to any seeds                module=p2p
I[2019-07-16|18:42:26.853] Executed block                               module=state height=1 validTxs=0 invalidTxs=0
I[2019-07-16|18:42:26.865] Committed state                              module=state height=1 txs=0 appHash=
=======
I[2022-11-09|09:06:34.444] Generated private validator                  module=main keyFile=/tmp/cometbft-home/config/priv_validator_key.json stateFile=/tmp/cometbft-home/data/priv_validator_state.json
I[2022-11-09|09:06:34.444] Generated node key                           module=main path=/tmp/cometbft-home/config/node_key.json
I[2022-11-09|09:06:34.444] Generated genesis file                       module=main path=/tmp/cometbft-home/config/genesis.json
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))
```

Now open another tab in your terminal and try sending a transaction:

```bash
$ curl -s 'localhost:26657/broadcast_tx_commit?tx="tendermint=rocks"'
{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "check_tx": {
      "gasWanted": "1"
    },
    "deliver_tx": {},
    "hash": "1B3C5A1093DB952C331B1749A21DCCBB0F6C7F4E0055CD04D16346472FC60EC6",
    "height": "128"
  }
}
```

Response should contain the height where this transaction was committed.

<<<<<<< HEAD
Now let's check if the given key now exists and its value:
=======
```bash
./kvstore -cmt-home /tmp/cometbft-home
```

The application will start and you should see a continuous output starting with:

```bash
badger 2022/11/09 09:08:50 INFO: All 0 tables opened in 0s
badger 2022/11/09 09:08:50 INFO: Discard stats nextEmptySlot: 0
badger 2022/11/09 09:08:50 INFO: Set nextTxnTs to 0
I[2022-11-09|09:08:50.085] service start                                module=proxy msg="Starting multiAppConn service" impl=multiAppConn
I[2022-11-09|09:08:50.085] service start                                module=abci-client connection=query msg="Starting localClient service" impl=localClient
I[2022-11-09|09:08:50.085] service start                                module=abci-client connection=snapshot msg="Starting localClient service" impl=localClient
...
```

More importantly, the application using CometBFT is producing blocks  🎉🎉 and you can see this reflected in the log output in lines like this:

```bash
I[2022-11-09|09:08:52.147] received proposal                            module=consensus proposal="Proposal{2/0 (F518444C0E348270436A73FD0F0B9DFEA758286BEB29482F1E3BEA75330E825C:1:C73D3D1273F2, -1) AD19AE292A45 @ 2022-11-09T12:08:52.143393Z}"
I[2022-11-09|09:08:52.152] received complete proposal block             module=consensus height=2 hash=F518444C0E348270436A73FD0F0B9DFEA758286BEB29482F1E3BEA75330E825C
I[2022-11-09|09:08:52.160] finalizing commit of block                   module=consensus height=2 hash=F518444C0E348270436A73FD0F0B9DFEA758286BEB29482F1E3BEA75330E825C root= num_txs=0
I[2022-11-09|09:08:52.167] executed block                               module=state height=2 num_valid_txs=0 num_invalid_txs=0
I[2022-11-09|09:08:52.171] committed state                              module=state height=2 num_txs=0 app_hash=
```

The blocks, as you can see from the `num_valid_txs=0` part, are empty, but let's remedy that next.

## 1.6 Using the application

Let's try submitting a transaction to our new application.
Open another terminal window and run the following curl command:


```bash
curl -s 'localhost:26657/broadcast_tx_commit?tx="cometbft=rocks"'
```

If everything went well, you should see a response indicating which height the
transaction was included in the blockchain.

Finally, let's make sure that transaction really was persisted by the application.
Run the following command:

```bash
curl -s 'localhost:26657/abci_query?data="cometbft"'
```

Let's examine the response object that this request returns.
The request returns a `json` object with a `key` and `value` field set.
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

```json
$ curl -s 'localhost:26657/abci_query?data="tendermint"'
{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "response": {
      "log": "exists",
      "key": "dGVuZGVybWludA==",
      "value": "cm9ja3M="
    }
  }
}
```

<<<<<<< HEAD
"dGVuZGVybWludA==" and "cm9ja3M=" are the base64-encoding of the ASCII of
"tendermint" and "rocks" accordingly.
=======
Those values don't look like the `key` and `value` we sent to CometBFT.
What's going on here?

The response contains a `base64` encoded representation of the data we submitted.
To get the original value out of this data, we can use the `base64` command line utility:

```bash
echo cm9ja3M=" | base64 -d
```
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

## Outro

I hope everything went smoothly and your first, but hopefully not the last,
<<<<<<< HEAD
Tendermint Core application is up and running. If not, please [open an issue on
Github](https://github.com/tendermint/tendermint/issues/new/choose). To dig
deeper, read [the docs](https://docs.tendermint.com/v0.34/).
=======
CometBFT application is up and running. If not, please [open an issue on
Github](https://github.com/cometbft/cometbft/issues/new/choose). To dig
deeper, read [the docs](https://docs.cometbft.com/main/).
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))
