package oracle

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	// cfg "github.com/cometbft/cometbft/config"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/oracle/service/adapters"
	"github.com/cometbft/cometbft/oracle/service/runner"
	oracletypes "github.com/cometbft/cometbft/oracle/service/types"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/redis"
	"github.com/cometbft/cometbft/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// Reactor handles mempool tx broadcasting amongst peers.
// It maintains a map from peer ID to counter, to prevent gossiping txs to the
// peers you received it from.
type Reactor struct {
	p2p.BaseReactor
	OracleInfo  *oracletypes.OracleInfo
	grpcAddress string
	// config  *cfg.MempoolConfig
	// mempool *CListMempool
	// ids     *mempoolIDs
}

// NewReactor returns a new Reactor with the given config and mempool.
func NewReactor(configPath string, grpcAddress string) *Reactor {
	// load oracle.json config if present
	jsonFile, openErr := os.Open(configPath)
	if openErr != nil {
		logrus.Warnf("[oracle] error opening oracle.json config file: %v", openErr)
	}

	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		logrus.Warnf("[oracle] error reading oracle.json config file: %v", err)
	}

	var config oracletypes.Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		logrus.Warnf("[oracle] error parsing oracle.json config file: %v", err)
	}

	oracleInfo := oracletypes.OracleInfo{
		Oracles: nil,
		Config:  config,
	}

	jsonFile.Close()

	memR := &Reactor{
		OracleInfo:  &oracleInfo,
		grpcAddress: grpcAddress,
	}
	memR.BaseReactor = *p2p.NewBaseReactor("Oracle", memR)

	return memR
}

// InitPeer implements Reactor by creating a state for the peer.
// func (memR *Reactor) InitPeer(peer p2p.Peer) p2p.Peer {
// 	memR.ids.ReserveForPeer(peer)
// 	return peer
// }

// SetLogger sets the Logger on the reactor and the underlying mempool.
func (memR *Reactor) SetLogger(l log.Logger) {
	memR.Logger = l
	memR.BaseService.SetLogger(l)
}

// OnStart implements p2p.BaseReactor.
func (memR *Reactor) OnStart() error {
	memR.OracleInfo.Redis = redis.NewService(0)

	grpcMaxRetryCount := 12
	retryCount := 0
	sleepTime := time.Second
	var client *grpc.ClientConn

	for {
		logrus.Infof("[oracle] trying to connect to grpc with address %s : %d", memR.grpcAddress, retryCount)
		if retryCount == grpcMaxRetryCount {
			panic("failed to connect to grpc:grpcClient after 12 tries")
		}
		time.Sleep(sleepTime)

		// reinit otherwise connection will be idle, in idle we can't tell if it's really ready
		var err error
		client, err = grpc.Dial(
			memR.grpcAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			panic(err)
		}
		// give it some time to connect after dailing, but not too long as connection can become idle
		time.Sleep(time.Duration(retryCount*int(time.Second) + 1))

		if client.GetState() == connectivity.Ready {
			memR.OracleInfo.GrpcClient = client
			break
		}
		client.Close()
		retryCount++
		sleepTime *= 2
	}

	memR.OracleInfo.AdapterMap = adapters.GetAdapterMap(memR.OracleInfo.GrpcClient, &memR.OracleInfo.Redis)
	logrus.Info("[oracle] running oracle service...")
	runner.Run(memR.OracleInfo)

	return nil
}

// GetChannels implements Reactor by returning the list of channels for this
// reactor.
// func (memR *Reactor) GetChannels() []*p2p.ChannelDescriptor {
// 	largestTx := make([]byte, memR.config.MaxTxBytes)
// 	batchMsg := protomem.Message{
// 		Sum: &protomem.Message_Txs{
// 			Txs: &protomem.Txs{Txs: [][]byte{largestTx}},
// 		},
// 	}

// 	return []*p2p.ChannelDescriptor{
// 		{
// 			ID:                  MempoolChannel,
// 			Priority:            5,
// 			RecvMessageCapacity: batchMsg.Size(),
// 			MessageType:         &protomem.Message{},
// 		},
// 	}
// }

// AddPeer implements Reactor.
// It starts a broadcast routine ensuring all txs are forwarded to the given peer.
// func (memR *Reactor) AddPeer(peer p2p.Peer) {
// 	if memR.config.Broadcast {
// 		go memR.broadcastTxRoutine(peer)
// 	}
// }

// RemovePeer implements Reactor.
// func (memR *Reactor) RemovePeer(peer p2p.Peer, _ interface{}) {
// 	memR.ids.Reclaim(peer)
// 	// broadcast routine checks if peer is gone and returns
// }

// // Receive implements Reactor.
// // It adds any received transactions to the mempool.
// func (memR *Reactor) Receive(e p2p.Envelope) {
// 	memR.Logger.Debug("Receive", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
// 	switch msg := e.Message.(type) {
// 	case *protomem.Txs:
// 		protoTxs := msg.GetTxs()
// 		if len(protoTxs) == 0 {
// 			memR.Logger.Error("received empty txs from peer", "src", e.Src)
// 			return
// 		}
// 		txInfo := TxInfo{SenderID: memR.ids.GetForPeer(e.Src)}
// 		if e.Src != nil {
// 			txInfo.SenderP2PID = e.Src.ID()
// 		}

// 		var err error
// 		for _, tx := range protoTxs {
// 			ntx := types.Tx(tx)
// 			err = memR.mempool.CheckTx(ntx, nil, txInfo)
// 			if errors.Is(err, ErrTxInCache) {
// 				memR.Logger.Debug("Tx already exists in cache", "tx", ntx.String())
// 			} else if err != nil {
// 				memR.Logger.Info("Could not check tx", "tx", ntx.String(), "err", err)
// 			}
// 		}
// 	default:
// 		memR.Logger.Error("unknown message type", "src", e.Src, "chId", e.ChannelID, "msg", e.Message)
// 		memR.Switch.StopPeerForError(e.Src, fmt.Errorf("mempool cannot handle message of type: %T", e.Message))
// 		return
// 	}

// 	// broadcasting happens from go routines per peer
// }

// PeerState describes the state of a peer.
type PeerState interface {
	GetHeight() int64
}

// // Send new mempool txs to peer.
// func (memR *Reactor) broadcastTxRoutine(peer p2p.Peer) {
// 	peerID := memR.ids.GetForPeer(peer)
// 	var next *clist.CElement

// 	for {
// 		// In case of both next.NextWaitChan() and peer.Quit() are variable at the same time
// 		if !memR.IsRunning() || !peer.IsRunning() {
// 			return
// 		}
// 		// This happens because the CElement we were looking at got garbage
// 		// collected (removed). That is, .NextWait() returned nil. Go ahead and
// 		// start from the beginning.
// 		if next == nil {
// 			select {
// 			case <-memR.mempool.TxsWaitChan(): // Wait until a tx is available
// 				if next = memR.mempool.TxsFront(); next == nil {
// 					continue
// 				}
// 			case <-peer.Quit():
// 				return
// 			case <-memR.Quit():
// 				return
// 			}
// 		}

// 		// Make sure the peer is up to date.
// 		peerState, ok := peer.Get(types.PeerStateKey).(PeerState)
// 		if !ok {
// 			// Peer does not have a state yet. We set it in the consensus reactor, but
// 			// when we add peer in Switch, the order we call reactors#AddPeer is
// 			// different every time due to us using a map. Sometimes other reactors
// 			// will be initialized before the consensus reactor. We should wait a few
// 			// milliseconds and retry.
// 			time.Sleep(PeerCatchupSleepIntervalMS * time.Millisecond)
// 			continue
// 		}

// 		// Allow for a lag of 1 block.
// 		memTx := next.Value.(*mempoolTx)
// 		if peerState.GetHeight() < memTx.Height()-1 {
// 			time.Sleep(PeerCatchupSleepIntervalMS * time.Millisecond)
// 			continue
// 		}

// 		// NOTE: Transaction batching was disabled due to
// 		// https://github.com/tendermint/tendermint/issues/5796

// 		if !memTx.isSender(peerID) {
// 			success := peer.Send(p2p.Envelope{
// 				ChannelID: MempoolChannel,
// 				Message:   &protomem.Txs{Txs: [][]byte{memTx.tx}},
// 			})
// 			if !success {
// 				time.Sleep(PeerCatchupSleepIntervalMS * time.Millisecond)
// 				continue
// 			}
// 		}

// 		select {
// 		case <-next.NextWaitChan():
// 			// see the start of the for loop for nil check
// 			next = next.Next()
// 		case <-peer.Quit():
// 			return
// 		case <-memR.Quit():
// 			return
// 		}
// 	}
// }

// TxsMessage is a Message containing transactions.
type TxsMessage struct {
	Txs []types.Tx
}

// String returns a string representation of the TxsMessage.
func (m *TxsMessage) String() string {
	return fmt.Sprintf("[TxsMessage %v]", m.Txs)
}
