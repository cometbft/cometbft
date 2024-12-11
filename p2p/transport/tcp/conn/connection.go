package conn

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"reflect"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/cosmos/gogoproto/proto"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/config"
	flow "github.com/cometbft/cometbft/internal/flowrate"
	"github.com/cometbft/cometbft/internal/timer"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/protoio"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p/transport"
)

const (
	defaultMaxPacketMsgPayloadSize = 1024

	numBatchPacketMsgs = 10
	minReadBufferSize  = 1024
	minWriteBufferSize = 65536
	updateStats        = 2 * time.Second

	// some of these defaults are written in the user config
	// flushThrottle, sendRate, recvRate
	// TODO: remove values present in config.
	defaultFlushThrottle = 10 * time.Millisecond

	defaultSendRate     = int64(512000) // 500KB/s
	defaultRecvRate     = int64(512000) // 500KB/s
	defaultPingInterval = 60 * time.Second
	defaultPongTimeout  = 45 * time.Second

	// Capacity of the receive channel for each stream.
	maxRecvChanCap = 100
)

// OnReceiveFn is a callback func, which is called by the MConnection when a
// new message is received.
type OnReceiveFn = func(byte, []byte)

// MConnection is a multiplexed connection.
//
// __multiplex__ *noun* a system or signal involving simultaneous transmission
// of several messages along a single channel of communication.
//
// Each connection handles message transmission on multiple abstract
// communication streams. Each stream has a globally unique byte id. The byte
// id and the relative priorities of each stream are configured upon
// initialization of the connection.
//
// To open a stream, call OpenStream with the stream id. Remember that the
// stream id must be globally unique.
//
// Connection errors are communicated through the ErrorCh channel.
//
// Connection can be closed either by calling Close or FlushAndClose. If you
// need to flush all pending messages before closing the connection, call
// FlushAndClose. Otherwise, call Close.
type MConnection struct {
	service.BaseService

	conn          net.Conn
	bufConnReader *bufio.Reader
	bufConnWriter *bufio.Writer
	sendMonitor   *flow.Monitor
	recvMonitor   *flow.Monitor
	send          chan struct{}
	pong          chan struct{}
	errorCh       chan error
	config        MConnConfig

	// Closing quitSendRoutine will cause the sendRoutine to eventually quit.
	// doneSendRoutine is closed when the sendRoutine actually quits.
	quitSendRoutine chan struct{}
	doneSendRoutine chan struct{}

	// Closing quitRecvRouting will cause the recvRouting to eventually quit.
	quitRecvRoutine chan struct{}

	flushTimer *timer.ThrottleTimer // flush writes as necessary but throttled.
	pingTimer  *time.Ticker         // send pings periodically

	// close conn if pong is not received in pongTimeout
	pongTimer     *time.Timer
	pongTimeoutCh chan bool // true - timeout, false - peer sent pong

	chStatsTimer *time.Ticker // update channel stats periodically

	created time.Time // time of creation

	_maxPacketMsgSize int

	// streamID -> channel
	channelsIdx map[byte]*stream

	// A map which stores the received messages. Used in tests.
	msgsByStreamIDMap map[byte]chan []byte

	onReceiveFn OnReceiveFn
}

var _ transport.Conn = (*MConnection)(nil)

// MConnConfig is a MConnection configuration.
type MConnConfig struct {
	SendRate int64 `mapstructure:"send_rate"`
	RecvRate int64 `mapstructure:"recv_rate"`

	// Maximum payload size
	MaxPacketMsgPayloadSize int `mapstructure:"max_packet_msg_payload_size"`

	// Interval to flush writes (throttled)
	FlushThrottle time.Duration `mapstructure:"flush_throttle"`

	// Interval to send pings
	PingInterval time.Duration `mapstructure:"ping_interval"`

	// Maximum wait time for pongs
	PongTimeout time.Duration `mapstructure:"pong_timeout"`

	// Fuzz connection
	TestFuzz       bool                   `mapstructure:"test_fuzz"`
	TestFuzzConfig *config.FuzzConnConfig `mapstructure:"test_fuzz_config"`
}

// DefaultMConnConfig returns the default config.
func DefaultMConnConfig() MConnConfig {
	return MConnConfig{
		SendRate:                defaultSendRate,
		RecvRate:                defaultRecvRate,
		MaxPacketMsgPayloadSize: defaultMaxPacketMsgPayloadSize,
		FlushThrottle:           defaultFlushThrottle,
		PingInterval:            defaultPingInterval,
		PongTimeout:             defaultPongTimeout,
	}
}

// NewMConnection wraps net.Conn and creates multiplex connection.
func NewMConnection(conn net.Conn, config MConnConfig) *MConnection {
	if config.PongTimeout >= config.PingInterval {
		panic("pongTimeout must be less than pingInterval (otherwise, next ping will reset pong timer)")
	}

	mconn := &MConnection{
		conn:              conn,
		bufConnReader:     bufio.NewReaderSize(conn, minReadBufferSize),
		bufConnWriter:     bufio.NewWriterSize(conn, minWriteBufferSize),
		sendMonitor:       flow.New(0, 0),
		recvMonitor:       flow.New(0, 0),
		send:              make(chan struct{}, 1),
		pong:              make(chan struct{}, 1),
		errorCh:           make(chan error, 1),
		config:            config,
		created:           time.Now(),
		channelsIdx:       make(map[byte]*stream),
		msgsByStreamIDMap: make(map[byte]chan []byte),
	}

	mconn.BaseService = *service.NewBaseService(nil, "MConnection", mconn)

	// maxPacketMsgSize() is a bit heavy, so call just once
	mconn._maxPacketMsgSize = mconn.maxPacketMsgSize()

	return mconn
}

// OnReceive sets the callback function to be executed each time we read a message.
func (c *MConnection) OnReceive(fn OnReceiveFn) {
	c.onReceiveFn = fn
}

func (c *MConnection) SetLogger(l log.Logger) {
	c.BaseService.SetLogger(l)
}

// OnStart implements BaseService.
func (c *MConnection) OnStart() error {
	if err := c.BaseService.OnStart(); err != nil {
		return err
	}
	c.flushTimer = timer.NewThrottleTimer("flush", c.config.FlushThrottle)
	c.pingTimer = time.NewTicker(c.config.PingInterval)
	c.pongTimeoutCh = make(chan bool, 1)
	c.chStatsTimer = time.NewTicker(updateStats)
	c.quitSendRoutine = make(chan struct{})
	c.doneSendRoutine = make(chan struct{})
	c.quitRecvRoutine = make(chan struct{})
	go c.sendRoutine()
	go c.recvRoutine()
	return nil
}

func (c *MConnection) Conn() net.Conn {
	return c.conn
}

// stopServices stops the BaseService and timers and closes the quitSendRoutine.
// if the quitSendRoutine was already closed, it returns true, otherwise it returns false.
func (c *MConnection) stopServices() (alreadyStopped bool) {
	select {
	case <-c.quitSendRoutine:
		// already quit
		return true
	default:
	}

	select {
	case <-c.quitRecvRoutine:
		// already quit
		return true
	default:
	}

	c.flushTimer.Stop()
	c.pingTimer.Stop()
	c.chStatsTimer.Stop()

	// inform the recvRouting that we are shutting down
	close(c.quitRecvRoutine)
	close(c.quitSendRoutine)
	return false
}

// ErrorCh returns a channel that will receive errors from the connection.
func (c *MConnection) ErrorCh() <-chan error {
	return c.errorCh
}

func (c *MConnection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *MConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// OpenStream opens a new stream on the connection. Remember that the
// stream id must be globally unique.
//
// Panics if the connection is already running (i.e., all streams
// must be registered in advance).
func (c *MConnection) OpenStream(streamID byte, desc any) (transport.Stream, error) {
	if c.IsRunning() {
		panic("MConnection is already running. Please register all streams in advance")
	}

	c.Logger.Debug("Opening stream", "streamID", streamID, "desc", desc)

	if _, ok := c.channelsIdx[streamID]; ok {
		return nil, fmt.Errorf("stream %X already exists", streamID)
	}

	d := StreamDescriptor{
		ID:       streamID,
		Priority: 1,
	}
	if desc, ok := desc.(StreamDescriptor); ok {
		d = desc
	}
	c.channelsIdx[streamID] = newChannel(c, d)
	c.channelsIdx[streamID].SetLogger(c.Logger.With("streamID", streamID))
	c.msgsByStreamIDMap[streamID] = make(chan []byte, maxRecvChanCap)

	return &MConnectionStream{conn: c, streamID: streamID}, nil
}

// HandshakeStream returns the underlying net.Conn connection.
func (c *MConnection) HandshakeStream() transport.HandshakeStream {
	return c.conn
}

// Close closes the connection. It flushes all pending writes before closing.
func (c *MConnection) Close(reason string) error {
	if err := c.Stop(); err != nil {
		// If the connection was not fully started (an error occurred before the
		// peer was started), close the underlying connection.
		if errors.Is(err, service.ErrNotStarted) {
			return c.conn.Close()
		}
		return err
	}

	if c.stopServices() {
		return nil
	}

	// inform the error channel that we are shutting down.
	select {
	case c.errorCh <- errors.New(reason):
	default:
	}

	return c.conn.Close()
}

func (c *MConnection) FlushAndClose(reason string) error {
	if err := c.Stop(); err != nil {
		// If the connection was not fully started (an error occurred before the
		// peer was started), close the underlying connection.
		if errors.Is(err, service.ErrNotStarted) {
			return c.conn.Close()
		}
		return err
	}

	if c.stopServices() {
		return nil
	}

	// inform the error channel that we are shutting down.
	select {
	case c.errorCh <- errors.New(reason):
	default:
	}

	// flush all pending writes
	{
		// wait until the sendRoutine exits
		// so we dont race on calling sendSomePacketMsgs
		<-c.doneSendRoutine
		// Send and flush all pending msgs.
		// Since sendRoutine has exited, we can call this
		// safely
		w := protoio.NewDelimitedWriter(c.bufConnWriter)
		eof := c.sendBatchPacketMsgs(w, numBatchPacketMsgs)
		for !eof {
			eof = c.sendBatchPacketMsgs(w, numBatchPacketMsgs)
		}
		_ = c.flush()
	}

	return c.conn.Close()
}

func (c *MConnection) ConnState() (state transport.ConnState) {
	state.ConnectedFor = time.Since(c.created)
	state.SendRateLimiterDelay = c.sendMonitor.Status().SleepTime
	state.RecvRateLimiterDelay = c.recvMonitor.Status().SleepTime
	state.StreamStates = make(map[byte]transport.StreamState)

	for streamID, channel := range c.channelsIdx {
		state.StreamStates[streamID] = transport.StreamState{
			SendQueueSize:     channel.loadSendQueueSize(),
			SendQueueCapacity: cap(channel.sendQueue),
		}
	}

	return state
}

func (c *MConnection) String() string {
	return fmt.Sprintf("MConn{%v}", c.conn.RemoteAddr())
}

func (c *MConnection) flush() error {
	return c.bufConnWriter.Flush()
}

// Catch panics, usually caused by remote disconnects.
func (c *MConnection) _recover() {
	if r := recover(); r != nil {
		c.Logger.Error("MConnection panicked", "err", r, "stack", string(debug.Stack()))
		c.Close(fmt.Sprintf("recovered from panic: %v", r))
	}
}

// thread-safe.
func (c *MConnection) sendBytes(chID byte, msgBytes []byte, blocking bool) error {
	if !c.IsRunning() {
		return nil
	}

	// c.Logger.Debug("Send",
	// 	"streamID", chID,
	// 	"msgBytes", log.NewLazySprintf("%X", msgBytes),
	// 	"timeout", timeout)

	channel, ok := c.channelsIdx[chID]
	if !ok {
		panic(fmt.Sprintf("Unknown channel %X. Forgot to register?", chID))
	}
	if err := channel.sendBytes(msgBytes, blocking); err != nil {
		// c.Logger.Error("Send failed", "err", err)
		return err
	}

	// Wake up sendRoutine if necessary
	select {
	case c.send <- struct{}{}:
	default:
	}
	return nil
}

// CanSend returns true if you can send more data onto the chID, false
// otherwise. Use only as a heuristic.
//
// thread-safe.
func (c *MConnection) CanSend(chID byte) bool {
	if !c.IsRunning() {
		return false
	}

	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.Logger.Error(fmt.Sprintf("Unknown channel %X", chID))
		return false
	}
	return channel.canSend()
}

// sendRoutine polls for packets to send from channels.
func (c *MConnection) sendRoutine() {
	defer c._recover()

	protoWriter := protoio.NewDelimitedWriter(c.bufConnWriter)

FOR_LOOP:
	for {
		var _n int
		var err error
	SELECTION:
		select {
		case <-c.flushTimer.Ch:
			// NOTE: flushTimer.Set() must be called every time
			// something is written to .bufConnWriter.
			if fErr := c.flush(); fErr != nil {
				c.Logger.Error("Failed to flush", "err", fErr)
			}
		case <-c.chStatsTimer.C:
			for _, channel := range c.channelsIdx {
				channel.updateStats()
			}
		case <-c.pingTimer.C:
			c.Logger.Debug("Send Ping")
			_n, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
			if err != nil {
				c.Logger.Error("Failed to send PacketPing", "err", err)
				break SELECTION
			}
			c.sendMonitor.Update(_n)
			c.Logger.Debug("Starting pong timer", "dur", c.config.PongTimeout)
			c.pongTimer = time.AfterFunc(c.config.PongTimeout, func() {
				select {
				case c.pongTimeoutCh <- true:
				default:
				}
			})
			if fErr := c.flush(); fErr != nil {
				c.Logger.Error("Failed to flush", "err", fErr)
			}
		case timeout := <-c.pongTimeoutCh:
			if timeout {
				c.Logger.Debug("Pong timeout")
				err = errors.New("pong timeout")
			} else {
				c.stopPongTimer()
			}
		case <-c.pong:
			c.Logger.Debug("Send Pong")
			_n, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
			if err != nil {
				c.Logger.Error("Failed to send PacketPong", "err", err)
				break SELECTION
			}
			c.sendMonitor.Update(_n)
			if fErr := c.flush(); fErr != nil {
				c.Logger.Error("Failed to flush", "err", fErr)
			}
		case <-c.quitSendRoutine:
			break FOR_LOOP
		case <-c.send:
			// Send some PacketMsgs
			eof := c.sendSomePacketMsgs(protoWriter)
			if !eof {
				// Keep sendRoutine awake.
				select {
				case c.send <- struct{}{}:
				default:
				}
			}
		}

		if !c.IsRunning() {
			break FOR_LOOP
		}
		if err != nil {
			c.Logger.Error("Connection failed @ sendRoutine", "err", err)
			c.Close(err.Error())
			break FOR_LOOP
		}
	}

	// Cleanup
	c.stopPongTimer()
	close(c.doneSendRoutine)
}

// Returns true if messages from channels were exhausted.
// Blocks in accordance to .sendMonitor throttling.
func (c *MConnection) sendSomePacketMsgs(w protoio.Writer) bool {
	// Block until .sendMonitor says we can write.
	// Once we're ready we send more than we asked for,
	// but amortized it should even out.
	c.sendMonitor.Limit(c._maxPacketMsgSize, c.config.SendRate, true)

	// Now send some PacketMsgs.
	return c.sendBatchPacketMsgs(w, numBatchPacketMsgs)
}

// Returns true if messages from channels were exhausted.
func (c *MConnection) sendBatchPacketMsgs(w protoio.Writer, batchSize int) bool {
	// Send a batch of PacketMsgs.
	totalBytesWritten := 0
	defer func() {
		if totalBytesWritten > 0 {
			c.sendMonitor.Update(totalBytesWritten)
		}
	}()
	for i := 0; i < batchSize; i++ {
		channel := c.selectChannel()
		// nothing to send across any channel.
		if channel == nil {
			return true
		}
		bytesWritten, err := c.sendPacketMsgOnChannel(w, channel)
		if err {
			return true
		}
		totalBytesWritten += bytesWritten
	}
	return false
}

// selects a channel to gossip our next message on.
// TODO: Make "batchChannelToGossipOn", so we can do our proto marshaling overheads in parallel,
// and we can avoid re-checking for `isSendPending`.
// We can easily mock the recentlySent differences for the batch choosing.
func (c *MConnection) selectChannel() *stream {
	// Choose a channel to create a PacketMsg from.
	// The chosen channel will be the one whose recentlySent/priority is the least.
	var leastRatio float32 = math.MaxFloat32
	var leastChannel *stream
	for _, channel := range c.channelsIdx {
		// If nothing to send, skip this channel
		// TODO: Skip continually looking for isSendPending on channels we've already skipped in this batch-send.
		if !channel.isSendPending() {
			continue
		}
		// Get ratio, and keep track of lowest ratio.
		// TODO: RecentlySent right now is bytes. This should be refactored to num messages to fix
		// gossip prioritization bugs.
		ratio := float32(channel.recentlySent) / float32(channel.desc.Priority)
		if ratio < leastRatio {
			leastRatio = ratio
			leastChannel = channel
		}
	}
	return leastChannel
}

// returns (num_bytes_written, error_occurred).
func (c *MConnection) sendPacketMsgOnChannel(w protoio.Writer, sendChannel *stream) (int, bool) {
	// Make & send a PacketMsg from this channel
	n, err := sendChannel.writePacketMsgTo(w)
	if err != nil {
		c.Logger.Error("Failed to write PacketMsg", "err", err)
		c.Close(err.Error())
		return n, true
	}
	// TODO: Change this to only add flush signals at the start and end of the batch.
	c.flushTimer.Set()
	return n, false
}

// recvRoutine reads PacketMsgs and reconstructs the message using the
// channels' "recving" buffer. After a whole message has been assembled, it's
// pushed to an internal queue, which is accessible via Read. Blocks depending
// on how the connection is throttled. Otherwise, it never blocks.
func (c *MConnection) recvRoutine() {
	defer c._recover()

	protoReader := protoio.NewDelimitedReader(c.bufConnReader, c._maxPacketMsgSize)

FOR_LOOP:
	for {
		// Block until .recvMonitor says we can read.
		c.recvMonitor.Limit(c._maxPacketMsgSize, atomic.LoadInt64(&c.config.RecvRate), true)

		// Peek into bufConnReader for debugging
		/*
			if numBytes := c.bufConnReader.Buffered(); numBytes > 0 {
				bz, err := c.bufConnReader.Peek(cmtmath.MinInt(numBytes, 100))
				if err == nil {
					// return
				} else {
					c.Logger.Debug("Error peeking connection buffer", "err", err)
					// return nil
				}
				c.Logger.Info("Peek connection buffer", "numBytes", numBytes, "bz", bz)
			}
		*/

		// Read packet type
		var packet tmp2p.Packet

		_n, err := protoReader.ReadMsg(&packet)
		c.recvMonitor.Update(_n)
		if err != nil {
			// stopServices was invoked and we are shutting down
			// receiving is expected to fail since we will close the connection
			select {
			case <-c.quitRecvRoutine:
				break FOR_LOOP
			default:
			}

			if c.IsRunning() {
				if errors.Is(err, io.EOF) {
					c.Logger.Info("Connection is closed @ recvRoutine (likely by the other side)")
				} else {
					c.Logger.Debug("Connection failed @ recvRoutine (reading byte)", "err", err)
				}
				c.Close(err.Error())
			}
			break FOR_LOOP
		}

		// Read more depending on packet type.
		switch pkt := packet.Sum.(type) {
		case *tmp2p.Packet_PacketPing:
			// TODO: prevent abuse, as they cause flush()'s.
			// https://github.com/tendermint/tendermint/issues/1190
			c.Logger.Debug("Receive Ping")
			select {
			case c.pong <- struct{}{}:
			default:
				// never block
			}
		case *tmp2p.Packet_PacketPong:
			c.Logger.Debug("Receive Pong")
			select {
			case c.pongTimeoutCh <- false:
			default:
				// never block
			}
		case *tmp2p.Packet_PacketMsg:
			channelID := byte(pkt.PacketMsg.ChannelID)
			channel, ok := c.channelsIdx[channelID]
			if !ok || pkt.PacketMsg.ChannelID < 0 || pkt.PacketMsg.ChannelID > math.MaxUint8 {
				err := fmt.Errorf("unknown channel %X", pkt.PacketMsg.ChannelID)
				c.Logger.Error("Connection failed @ recvRoutine", "err", err)
				c.Close(err.Error())
				break FOR_LOOP
			}

			msgBytes, err := channel.recvPacketMsg(*pkt.PacketMsg)
			if err != nil {
				c.Logger.Error("Connection failed @ recvRoutine", "err", err)
				c.Close(err.Error())
				break FOR_LOOP
			}
			if msgBytes != nil {
				// c.Logger.Debug("Received", "streamID", channelID, "msgBytes", log.NewLazySprintf("%X", msgBytes))
				if c.onReceiveFn != nil {
					c.onReceiveFn(channelID, msgBytes)
				} else {
					bz := make([]byte, len(msgBytes))
					copy(bz, msgBytes)
					c.msgsByStreamIDMap[channelID] <- bz
				}
			}
		default:
			err := fmt.Errorf("unknown message type %v", reflect.TypeOf(packet))
			c.Logger.Error("Connection failed @ recvRoutine", "err", err)
			c.Close(err.Error())
			break FOR_LOOP
		}
	}

	// Cleanup
	close(c.pong)
}

// Used in tests.
func (c *MConnection) readBytes(streamID byte, b []byte, timeout time.Duration) (n int, err error) {
	select {
	case msgBytes := <-c.msgsByStreamIDMap[streamID]:
		n = copy(b, msgBytes)
		if n < len(msgBytes) {
			err = errors.New("short buffer")
			return 0, err
		}
		return n, nil
	case <-time.After(timeout):
		return 0, errors.New("read timeout")
	}
}

// not goroutine-safe.
func (c *MConnection) stopPongTimer() {
	if c.pongTimer != nil {
		_ = c.pongTimer.Stop()
		c.pongTimer = nil
	}
}

// maxPacketMsgSize returns a maximum size of PacketMsg.
func (c *MConnection) maxPacketMsgSize() int {
	bz, err := proto.Marshal(mustWrapPacket(&tmp2p.PacketMsg{
		ChannelID: 0x01,
		EOF:       true,
		Data:      make([]byte, c.config.MaxPacketMsgPayloadSize),
	}))
	if err != nil {
		panic(err)
	}
	return len(bz)
}

// -----------------------------------------------------------------------------

// NOTE: not goroutine-safe.
type stream struct {
	conn          *MConnection
	desc          StreamDescriptor
	sendQueue     chan []byte
	sendQueueSize int32 // atomic.
	recving       []byte
	sending       []byte
	recentlySent  int64 // exponential moving average

	nextPacketMsg           *tmp2p.PacketMsg
	nextP2pWrapperPacketMsg *tmp2p.Packet_PacketMsg
	nextPacket              *tmp2p.Packet

	maxPacketMsgPayloadSize int

	Logger log.Logger
}

func newChannel(conn *MConnection, desc StreamDescriptor) *stream {
	desc = desc.FillDefaults()
	if desc.Priority <= 0 {
		panic("Channel default priority must be a positive integer")
	}
	return &stream{
		conn:                    conn,
		desc:                    desc,
		sendQueue:               make(chan []byte, desc.SendQueueCapacity),
		recving:                 make([]byte, 0, desc.RecvBufferCapacity),
		nextPacketMsg:           &tmp2p.PacketMsg{ChannelID: int32(desc.ID)},
		nextP2pWrapperPacketMsg: &tmp2p.Packet_PacketMsg{},
		nextPacket:              &tmp2p.Packet{},
		maxPacketMsgPayloadSize: conn.config.MaxPacketMsgPayloadSize,
	}
}

func (ch *stream) SetLogger(l log.Logger) {
	ch.Logger = l
}

// Queues message to send to this channel. Blocks if blocking is true.
// thread-safe.
func (ch *stream) sendBytes(bytes []byte, blocking bool) error {
	if blocking {
		select {
		case ch.sendQueue <- bytes:
			atomic.AddInt32(&ch.sendQueueSize, 1)
			return nil
		case <-ch.conn.Quit():
			return nil
		}
	}

	select {
	case ch.sendQueue <- bytes:
		atomic.AddInt32(&ch.sendQueueSize, 1)
		return nil
	default:
		return ErrWriteQueueFull{}
	case <-ch.conn.Quit():
		return nil
	}
}

// Goroutine-safe.
func (ch *stream) loadSendQueueSize() (size int) {
	return int(atomic.LoadInt32(&ch.sendQueueSize))
}

// Goroutine-safe
// Use only as a heuristic.
func (ch *stream) canSend() bool {
	return ch.loadSendQueueSize() < defaultSendQueueCapacity
}

// Returns true if any PacketMsgs are pending to be sent.
// Call before calling updateNextPacket
// Goroutine-safe.
func (ch *stream) isSendPending() bool {
	if len(ch.sending) == 0 {
		if len(ch.sendQueue) == 0 {
			return false
		}
		ch.sending = <-ch.sendQueue
	}
	return true
}

// Updates the nextPacket proto message for us to send.
// Not goroutine-safe.
func (ch *stream) updateNextPacket() {
	maxSize := ch.maxPacketMsgPayloadSize
	if len(ch.sending) <= maxSize {
		ch.nextPacketMsg.Data = ch.sending
		ch.nextPacketMsg.EOF = true
		ch.sending = nil
		atomic.AddInt32(&ch.sendQueueSize, -1) // decrement sendQueueSize
	} else {
		ch.nextPacketMsg.Data = ch.sending[:maxSize]
		ch.nextPacketMsg.EOF = false
		ch.sending = ch.sending[maxSize:]
	}

	ch.nextP2pWrapperPacketMsg.PacketMsg = ch.nextPacketMsg
	ch.nextPacket.Sum = ch.nextP2pWrapperPacketMsg
}

// Writes next PacketMsg to w and updates c.recentlySent.
// Not goroutine-safe.
func (ch *stream) writePacketMsgTo(w protoio.Writer) (n int, err error) {
	ch.updateNextPacket()
	n, err = w.WriteMsg(ch.nextPacket)
	if err != nil {
		err = ErrPacketWrite{Source: err}
	}

	atomic.AddInt64(&ch.recentlySent, int64(n))
	return n, err
}

// Handles incoming PacketMsgs. It returns a message bytes if message is
// complete. NOTE message bytes may change on next call to recvPacketMsg.
// Not goroutine-safe.
func (ch *stream) recvPacketMsg(packet tmp2p.PacketMsg) ([]byte, error) {
	recvCap, recvReceived := ch.desc.RecvMessageCapacity, len(ch.recving)+len(packet.Data)
	if recvCap < recvReceived {
		return nil, ErrPacketTooBig{Max: recvCap, Received: recvReceived}
	}

	ch.recving = append(ch.recving, packet.Data...)
	if packet.EOF {
		msgBytes := ch.recving

		// clear the slice without re-allocating.
		// http://stackoverflow.com/questions/16971741/how-do-you-clear-a-slice-in-go
		//   suggests this could be a memory leak, but we might as well keep the memory for the channel until it closes,
		//	at which point the recving slice stops being used and should be garbage collected
		ch.recving = ch.recving[:0] // make([]byte, 0, ch.desc.RecvBufferCapacity)
		return msgBytes, nil
	}
	return nil, nil
}

// Call this periodically to update stats for throttling purposes.
// thread-safe.
func (ch *stream) updateStats() {
	// Exponential decay of stats.
	// TODO: optimize.
	atomic.StoreInt64(&ch.recentlySent, int64(float64(atomic.LoadInt64(&ch.recentlySent))*0.8))
}

// ----------------------------------------
// Packet

// mustWrapPacket takes a packet kind (oneof) and wraps it in a tmp2p.Packet message.
func mustWrapPacket(pb proto.Message) *tmp2p.Packet {
	msg := &tmp2p.Packet{}
	mustWrapPacketInto(pb, msg)
	return msg
}

func mustWrapPacketInto(pb proto.Message, dst *tmp2p.Packet) {
	switch pb := pb.(type) {
	case *tmp2p.PacketPing:
		dst.Sum = &tmp2p.Packet_PacketPing{
			PacketPing: pb,
		}
	case *tmp2p.PacketPong:
		dst.Sum = &tmp2p.Packet_PacketPong{
			PacketPong: pb,
		}
	case *tmp2p.PacketMsg:
		dst.Sum = &tmp2p.Packet_PacketMsg{
			PacketMsg: pb,
		}
	default:
		panic(fmt.Errorf("unknown packet type %T", pb))
	}
}
