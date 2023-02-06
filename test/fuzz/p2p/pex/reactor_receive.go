package pex

import (
	"net"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/pex"
	"github.com/cometbft/cometbft/version"
	"github.com/cosmos/gogoproto/proto"
)

var (
	pexR *pex.Reactor
	peer p2p.Peer
)

func init() {
	addrB := pex.NewAddrBook("./testdata/addrbook1", false)
	pexR := pex.NewReactor(addrB, &pex.ReactorConfig{SeedMode: false})
	if pexR == nil {
		panic("NewReactor returned nil")
	}
	pexR.SetLogger(log.NewNopLogger())
	peer := newFuzzPeer()
	pexR.AddPeer(peer)

}

func Fuzz(data []byte) int {
	// MakeSwitch uses log.TestingLogger which can't be executed in init()
	cfg := config.DefaultP2PConfig()
	cfg.PexReactor = true
	sw := p2p.MakeSwitch(cfg, 0, "127.0.0.1", "123.123.123", func(i int, sw *p2p.Switch) *p2p.Switch {
		return sw
	})
	pexR.SetSwitch(sw)

	var msg proto.Message
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return 0
	}
	pexR.ReceiveEnvelope(p2p.Envelope{
		ChannelID: pex.PexChannel,
		Src:       peer,
		Message:   msg,
	})

	return 1
}

type fuzzPeer struct {
	*service.BaseService
	m map[string]interface{}
}

var _ p2p.Peer = (*fuzzPeer)(nil)

func newFuzzPeer() *fuzzPeer {
	fp := &fuzzPeer{m: make(map[string]interface{})}
	fp.BaseService = service.NewBaseService(nil, "fuzzPeer", fp)
	return fp
}

var privKey = ed25519.GenPrivKey()
var nodeID = p2p.PubKeyToID(privKey.PubKey())
var defaultNodeInfo = p2p.DefaultNodeInfo{
	ProtocolVersion: p2p.NewProtocolVersion(
		version.P2PProtocol,
		version.BlockProtocol,
		0,
	),
	DefaultNodeID: nodeID,
	ListenAddr:    "0.0.0.0:98992",
	Moniker:       "foo1",
}

func (fp *fuzzPeer) FlushStop()       {}
func (fp *fuzzPeer) ID() p2p.ID       { return nodeID }
func (fp *fuzzPeer) RemoteIP() net.IP { return net.IPv4(0, 0, 0, 0) }
func (fp *fuzzPeer) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: fp.RemoteIP(), Port: 98991, Zone: ""}
}
func (fp *fuzzPeer) IsOutbound() bool                    { return false }
func (fp *fuzzPeer) IsPersistent() bool                  { return false }
func (fp *fuzzPeer) CloseConn() error                    { return nil }
func (fp *fuzzPeer) NodeInfo() p2p.NodeInfo              { return defaultNodeInfo }
func (fp *fuzzPeer) Status() p2p.ConnectionStatus        { var cs p2p.ConnectionStatus; return cs }
func (fp *fuzzPeer) SocketAddr() *p2p.NetAddress         { return p2p.NewNetAddress(fp.ID(), fp.RemoteAddr()) }
func (fp *fuzzPeer) SendEnvelope(e p2p.Envelope) bool    { return true }
func (fp *fuzzPeer) TrySendEnvelope(e p2p.Envelope) bool { return true }
func (fp *fuzzPeer) Send(_ byte, _ []byte) bool          { return true }
func (fp *fuzzPeer) TrySend(_ byte, _ []byte) bool       { return true }
func (fp *fuzzPeer) Set(key string, value interface{})   { fp.m[key] = value }
func (fp *fuzzPeer) Get(key string) interface{}          { return fp.m[key] }
func (fp *fuzzPeer) GetRemovalFailed() bool              { return false }
func (fp *fuzzPeer) SetRemovalFailed()                   {}
