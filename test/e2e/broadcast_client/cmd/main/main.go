package main

import (
	"cometbft-client-experiment/internal"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cometbft/cometbft/test/loadtime/payload"
	"github.com/cometbft/cometbft/types"
	"github.com/google/uuid"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var (
	outputFile        = flag.String("output", "log.txt", "Output file to write the logs")
	rpcAddress        = flag.String("rpc", "http://localhost:5716", "Address of the RPC server")
	webSocket         = flag.String("ws", "ws://localhost:5716/websocket", "Address of the WebSocket server")
	txBatches         = flag.Int("txBatches", 1, "Total number of batches to send")
	totalConcurrentTx = flag.Int("totalConcurrentTx", 1, "Total number of concurrent transactions to send")
	broadcastType     = flag.String("broadcastType", "commit", "Broadcast type (commit, async, and sync)")
	delay             = flag.Int("delay", 1000, "Value in milliseconds to wait between batches")
	debug             = flag.Bool("debug", false, "Display progress")
)

var (
	transactionsWaitGroup = sync.WaitGroup{}

	currentBlockHeight = atomic.Int64{}

	running    = atomic.Bool{}
	loggerChan = make(chan string, 50)
)

func main() {
	flag.Parse()

	f, err := os.Create(*outputFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	running.Store(true)
	listenNewBlocks()
	currentBlockHeight.Store(getNodeHeight())
	go startWorkers(*txBatches, *broadcastType, *totalConcurrentTx, *delay)

	for log := range loggerChan {
		_, err := f.WriteString(log + "\n")
		if err != nil {
			panic(err)
		}
	}
}

func getNodeHeight() int64 {
	resp, err := http.Get(fmt.Sprintf("%s/status", *rpcAddress))
	if err != nil {
		panic(err)
	}

	var NodeStatus internal.NodeStatus
	err = json.NewDecoder(resp.Body).Decode(&NodeStatus)
	if err != nil {
		panic(err)
	}

	height, err := strconv.ParseInt(NodeStatus.Result.SyncInfo.LatestBlockHeight, 10, 64)
	if err != nil {
		panic(err)
	}
	return height
}

func logInFile(logStr string) {
	if running.Load() {
		loggerChan <- logStr
	}
}

// Subscribes to RPC and records new blocks. It's single threaded and has to parse each transaction to record them in the logs.
// Should be fine for comet's default 1s between blocks as long as the number of transactions is reasonable for the processing machine.
// For more demanding tests it would be better to write the block with a timestamp to a file to then process after the experiment is done.
func listenNewBlocks() {
	subscribeRequest := `{"jsonrpc":"2.0","method":"subscribe","params":["tm.event='NewBlock'"],"id":"1"}`
	c, _, err := websocket.DefaultDialer.Dial(*webSocket, nil)
	if err != nil {
		panic(err)
	}

	go func() {
		defer c.Close()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			var NewBlock internal.NewBlock
			err = json.Unmarshal(message, &NewBlock)
			if err != nil {
				log.Println("read:", err)
				continue
			}
			if NewBlock.Result.Data.Value.Block.Header.Height != "" {
				newHeight, err := strconv.ParseInt(NewBlock.Result.Data.Value.Block.Header.Height, 10, 64)
				if err != nil {
					log.Println(err)
					continue
				}

				currentBlock := currentBlockHeight.Load()
				if newHeight > currentBlock {
					currentBlockHeight.CompareAndSwap(currentBlock, newHeight)
				}

				totalTxs := len(NewBlock.Result.Data.Value.Block.Data.Txs)
				receivedTxs := ""
				for i := 0; i < totalTxs; i++ {
					str, err := url.QueryUnescape(NewBlock.Result.Data.Value.Block.Data.Txs[i])
					if err != nil {
						log.Println(err)
						continue
					}
					data, err := base64.StdEncoding.DecodeString(str)
					if err != nil {
						log.Println(err)
						continue
					}
					out, err := payload.FromBytes(data)
					if err != nil {
						//Prefix error means it encountered a tx that wasn't submitted by this client and thus can be ignored
						if !strings.Contains(err.Error(), "key prefix") {
							log.Println(err)
						}
						continue
					}
					receivedTxs += string(out.Time.AsTime().UTC().Format(time.RFC3339Nano)) + ","
				}

				timeInTransaction := NewBlock.Result.Data.Value.Block.Header.Time
				timeNow := time.Now().UTC().Format(time.RFC3339Nano)
				logInFile(fmt.Sprintf("NewBlock;%d;%d;%s;%s;%s", newHeight, totalTxs, timeInTransaction, timeNow, receivedTxs))
			}
		}
	}()

	err = c.WriteMessage(websocket.TextMessage, []byte(subscribeRequest))
	if err != nil {
		panic(err)
	}
}

// run txBatches batches of totalConcurrentTx simultaneous requests, waiting delay ms between each batch
func startWorkers(txBatches int, broadcastType string, totalConcurrentTx int, delay int) {
	id := [16]byte(uuid.New())

	d := time.Duration(delay) * time.Millisecond
	//Goes through a loop to start the broadcast processes
	//Debug only adds a print but the whole function is cloned for performance reasons
	if *debug {
		for i := 0; i < txBatches; i++ {
			fmt.Println(fmt.Sprintf("Sending batch %d of %d with %d transactions", i+1, txBatches, totalConcurrentTx))
			for i := 0; i < totalConcurrentTx; i++ {
				go sendTransaction(broadcastType, id)
			}
			time.Sleep(d)
		}
	} else {
		for i := 0; i < txBatches; i++ {
			for i := 0; i < totalConcurrentTx; i++ {
				go sendTransaction(broadcastType, id)
			}
			time.Sleep(d)
		}
	}

	if *debug {
		fmt.Println("Waiting for replies")
	}
	//wait until all broadcasts receive a reply
	transactionsWaitGroup.Wait()
	if *debug {
		fmt.Println("Replies received")
	}

	//wait for a commit transaction to give more time for sync and async transactions to be commited
	transaction, _ := generateTransaction(1024, id[:])
	tx := url.QueryEscape(string(transaction))
	http.Get(fmt.Sprintf("%s/broadcast_tx_commit?tx=\"%s\"", *rpcAddress, tx))

	//close
	running.Store(false)
	close(loggerChan)
}

// Sends transaction of type broadcastType and waits for a reply
// commit -> tx committed
// sync -> tx is in mempool
// async -> node received tx
func sendTransaction(broadcastType string, id [16]byte) {
	transactionsWaitGroup.Add(1)
	defer transactionsWaitGroup.Done()

	transaction, start := generateTransaction(1024, id[:])
	tx := url.QueryEscape(string(transaction))

	startBlockHeight := currentBlockHeight.Load()

	resp, err := http.Get(fmt.Sprintf("%s/broadcast_tx_%s?tx=\"%s\"", *rpcAddress, broadcastType, tx))
	if err != nil {
		panic(err)
	}

	elapsed := time.Since(start)

	timeStart := start.UTC().Format(time.RFC3339Nano)
	timeNow := time.Now().UTC().Format(time.RFC3339Nano)
	logInFile(fmt.Sprintf("transaction;%s;%d;%s;%s;%s", resp.Status, startBlockHeight, elapsed, timeStart, timeNow))

	err = resp.Body.Close()
	if err != nil {
		panic(err)
	}
}

// Generate random transaction data with timestamp
func generateTransaction(size int, id []byte) (types.Tx, time.Time) {
	p := payload.Payload{
		Id:          id,
		Size:        uint64(size),
		Rate:        uint64(1),
		Connections: uint64(1),
	}
	tx, err := payload.NewBytes(&p)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate tx: %v", err))
	}
	time := p.Time.AsTime()
	return tx, time
}
