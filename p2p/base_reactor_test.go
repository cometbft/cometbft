package p2p

import (
	"net"
	"sync"
	"testing"
	"time"

	mpproto "github.com/cometbft/cometbft/api/cometbft/mempool/v1"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p/conn"
	cmtconn "github.com/cometbft/cometbft/p2p/conn"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

// TestBaseReactorProcessor tests the BaseReactor's message processing by
// queueing encoded messages and adding artificial delay to the first message.
// Depending on the processors used, the ordering of the sender could be lost.
func TestBaseReactorProcessor(t *testing.T) {
	// a reactor that is using the default proessor should be able to queue
	// messages and they get processed in order.
	or := NewOrderedReactor(false)

	msgs := []string{"msg1", "msg2", "msg3"}
	or.fillQueue(t, msgs...)

	time.Sleep(300 * time.Millisecond) // wait plenty of time for the processing to finish

	require.Equal(t, len(msgs), len(or.received))
	require.Equal(t, msgs, or.received)

	// since the orderedReactor adds a delay to the first received message, we
	// expect the parallel processor to not be in the original send order.
	pr := NewOrderedReactor(true)

	pr.fillQueue(t, msgs...)
	time.Sleep(300 * time.Millisecond)
	require.NotEqual(t, msgs, pr.received)
}

var _ Reactor = &orderedReactor{}

// orderedReactor is used for testing. It saves each envelope in the order it
// receives it.
type orderedReactor struct {
	BaseReactor

	mtx           *sync.RWMutex
	received      []string
	receivedFirst bool
}

func NewOrderedReactor(parallel bool) *orderedReactor {
	r := &orderedReactor{mtx: &sync.RWMutex{}}
	procOpt := WithProcessor(DefaultProcessor(r))
	if parallel {
		procOpt = WithProcessor(ParallelProcessor(r, 2))
	}
	r.BaseReactor = *NewBaseReactor("Ordered Rector", r, procOpt, WithIncomingQueueSize(10))
	return r
}

func (r *orderedReactor) GetChannels() []*conn.ChannelDescriptor {
	return []*conn.ChannelDescriptor{
		{
			ID:                  0x99,
			Priority:            1,
			RecvMessageCapacity: 10,
			MessageType:         &mpproto.Txs{},
		},
	}

}

// Receive adds a delay to the first processed envelope to test ordering.
func (r *orderedReactor) Receive(e Envelope) {
	r.mtx.Lock()
	f := r.receivedFirst
	if !f {
		r.receivedFirst = true
		r.mtx.Unlock()
		time.Sleep(100 * time.Millisecond)
	} else {
		r.mtx.Unlock()
	}
	r.mtx.Lock()
	defer r.mtx.Unlock()

	envMsg := e.Message.(*mpproto.Txs)
	r.received = append(r.received, string(envMsg.Txs[0]))
}

func (r *orderedReactor) fillQueue(t *testing.T, msgs ...string) {
	peer := &imaginaryPeer{}
	for _, msg := range msgs {
		s, err := proto.Marshal(&mpproto.Txs{Txs: [][]byte{[]byte(msg)}})
		require.NoError(t, err)
		r.QueueUnprocessedEnvelope(UnprocessedEnvelope{
			Src:       peer,
			Message:   s,
			ChannelID: 0x99,
		})
	}
}

var _ IntrospectivePeer = &imaginaryPeer{}

type imaginaryPeer struct {
	service.BaseService
}

func (ip *imaginaryPeer) FlushStop()                         {}
func (ip *imaginaryPeer) ID() ID                             { return "" }
func (ip *imaginaryPeer) RemoteIP() net.IP                   { return []byte{} }
func (ip *imaginaryPeer) RemoteAddr() net.Addr               { return nil }
func (ip *imaginaryPeer) IsOutbound() bool                   { return true }
func (ip *imaginaryPeer) CloseConn() error                   { return nil }
func (ip *imaginaryPeer) IsPersistent() bool                 { return false }
func (ip *imaginaryPeer) NodeInfo() NodeInfo                 { return nil }
func (ip *imaginaryPeer) Status() cmtconn.ConnectionStatus   { return cmtconn.ConnectionStatus{} }
func (ip *imaginaryPeer) SocketAddr() *NetAddress            { return nil }
func (ip *imaginaryPeer) Send(e Envelope) bool               { return true }
func (ip *imaginaryPeer) TrySend(e Envelope) bool            { return true }
func (ip *imaginaryPeer) Set(key string, value any)          {}
func (ip *imaginaryPeer) Get(key string) any                 { return nil }
func (ip *imaginaryPeer) SetRemovalFailed()                  {}
func (ip *imaginaryPeer) GetRemovalFailed() bool             { return false }
func (ip *imaginaryPeer) Metrics() *Metrics                  { return NoopMetrics() }
func (ip *imaginaryPeer) ChIDToMetricLabel(chID byte) string { return "" }
func (ip *imaginaryPeer) ValueToMetricLabel(i any) string    { return "" }
