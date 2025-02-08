package mconn

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/cometbft/cometbft/internal/timer"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p/internal/fuzz"
	"github.com/cometbft/cometbft/p2p/transport"
)

const (
	defaultSendQueueCapacity   = 1
	defaultRecvBufferCapacity  = 4096
	defaultRecvMessageCapacity = 22020096      // 21MB
	defaultSendRate            = int64(512000) // 500KB/s
	defaultRecvRate            = int64(512000) // 500KB/s
	flushThrottleMS            = 100
)

// MConnConfig is the configuration for an MConnection
type MConnConfig struct {
	SendRate            int64
	RecvRate            int64
	SendQueueCapacity   int
	RecvBufferCapacity  int
	RecvMessageCapacity int

	// Fuzz params - for testing
	TestFuzz       bool
	TestFuzzConfig *fuzz.FuzzConnConfig
}

// DefaultMConnConfig returns the default config
func DefaultMConnConfig() MConnConfig {
	return MConnConfig{
		SendRate:            defaultSendRate,
		RecvRate:            defaultRecvRate,
		SendQueueCapacity:   defaultSendQueueCapacity,
		RecvBufferCapacity:  defaultRecvBufferCapacity,
		RecvMessageCapacity: defaultRecvMessageCapacity,
	}
}

// MConnection handles multiplexing of connections
type MConnection struct {
	service.BaseService

	conn        net.Conn
	bufReader   *bufio.Reader
	bufWriter   *bufio.Writer
	sendMonitor *timer.ThrottleTimer
	recvMonitor *timer.ThrottleTimer

	send     chan struct{}
	pong     chan struct{}
	channels map[byte]*Channel

	quit        chan struct{}
	flushTimer  *time.Timer
	sendRoutine sync.WaitGroup
	recvRoutine sync.WaitGroup

	config  MConnConfig
	logger  log.Logger
	metrics *Metrics
}

// NewMConnection creates a new connection with the given config
func NewMConnection(conn net.Conn, config MConnConfig) *MConnection {
	mconn := &MConnection{
		conn:        conn,
		bufReader:   bufio.NewReaderSize(conn, config.RecvBufferCapacity),
		bufWriter:   bufio.NewWriterSize(conn, config.RecvBufferCapacity),
		sendMonitor: timer.NewThrottleTimer("send", time.Duration(config.SendRate)),
		recvMonitor: timer.NewThrottleTimer("recv", time.Duration(config.RecvRate)),
		send:        make(chan struct{}, 1),
		pong:        make(chan struct{}, 1),
		channels:    make(map[byte]*Channel),
		quit:        make(chan struct{}),
		config:      config,
		logger:      log.NewNopLogger(),
		metrics:     NopMetrics(),
	}

	mconn.BaseService = *service.NewBaseService(nil, "MConnection", mconn)
	return mconn
}

// Start implements BaseService
func (c *MConnection) OnStart() error {
	c.flushTimer = time.NewTimer(flushThrottleMS * time.Millisecond)
	c.sendRoutine.Add(1)
	go c.sendRoutineHandler()
	c.recvRoutine.Add(1)
	go c.recvRoutineHandler()
	return nil
}

// Stop implements BaseService
func (c *MConnection) OnStop() {
	c.flushTimer.Stop()
	if err := c.conn.Close(); err != nil {
		c.logger.Error("Error closing connection", "err", err)
	}
	close(c.quit)
	c.sendRoutine.Wait()
	c.recvRoutine.Wait()
}

func (c *MConnection) sendRoutineHandler() {
	defer c.sendRoutine.Done()
	for {
		select {
		case <-c.flushTimer.C:
			// Send pending messages and flush
			c.flush()
		case <-c.send:
			// Send pending messages and flush
			c.flush()
		case <-c.quit:
			// Quit
			return
		}
	}
}

func (c *MConnection) recvRoutineHandler() {
	defer c.recvRoutine.Done()
	for {
		select {
		case <-c.quit:
			return
		default:
			// Receive and handle messages
			if err := c.receiveNext(); err != nil {
				c.logger.Error("Error receiving message", "err", err)
				return
			}
		}
	}
}

func (c *MConnection) flush() {
	// Flush the buffered writer
	if err := c.bufWriter.Flush(); err != nil {
		c.logger.Error("Error flushing buffer", "err", err)
	}
}

func (c *MConnection) receiveNext() error {
	packet, err := readPacket(c.bufReader)
	if err != nil {
		return err
	}

	channel, ok := c.channels[packet.ChID]
	if !ok {
		return fmt.Errorf("unknown channel %X", packet.ChID)
	}

	switch packet.Type {
	case packetTypeMsg:
		return c.handleMessage(channel, packet.Data)
	case packetTypePing:
		c.handlePing()
		return nil
	case packetTypePong:
		c.handlePong()
		return nil
	default:
		return fmt.Errorf("unknown packet type %X", packet.Type)
	}
}

func (c *MConnection) handleMessage(channel *Channel, data []byte) error {
	channel.mtx.Lock()
	defer channel.mtx.Unlock()

	channel.recving = data
	return nil
}

func (c *MConnection) handlePing() {
	select {
	case c.pong <- struct{}{}:
	default:
	}
}

func (c *MConnection) handlePong() {
	// Reset ping timer
}

// Send queues a message to be sent to a channel
func (c *MConnection) Send(chID byte, msg []byte) error {
	if !c.IsRunning() {
		return fmt.Errorf("cannot send message - connection is not running")
	}

	channel, ok := c.channels[chID]
	if !ok {
		return fmt.Errorf("cannot send message - unknown channel %X", chID)
	}

	select {
	case channel.sendQueue <- msg:
		// Wake up sendRoutine
		select {
		case c.send <- struct{}{}:
		default:
		}
		return nil
	default:
		return fmt.Errorf("send queue full")
	}
}

// OpenStream creates a new channel for the given ID
func (c *MConnection) OpenStream(chID byte, desc any) (transport.Stream, error) {
	if !c.IsRunning() {
		return nil, fmt.Errorf("cannot open stream - connection is not running")
	}

	if _, ok := c.channels[chID]; ok {
		return nil, fmt.Errorf("channel %X already exists", chID)
	}

	channel := newChannel(chID, c.config.RecvMessageCapacity)
	c.channels[chID] = channel

	return &MStream{
		conn: c,
		chID: chID,
	}, nil
}

// TrySend attempts to send a message without blocking
func (c *MConnection) TrySend(chID byte, msg []byte) error {
	if !c.IsRunning() {
		return fmt.Errorf("cannot send message - connection is not running")
	}

	channel, ok := c.channels[chID]
	if !ok {
		return fmt.Errorf("cannot send message - unknown channel %X", chID)
	}

	select {
	case channel.sendQueue <- msg:
		// Wake up sendRoutine
		select {
		case c.send <- struct{}{}:
		default:
		}
		return nil
	default:
		return fmt.Errorf("send queue full")
	}
}

// FlushClose flushes pending messages and closes the connection
func (c *MConnection) FlushClose() error {
	c.flush()
	return c.Stop()
}

// CloseConn implements the Conn interface
func (c *MConnection) CloseConn(reason string) error {
	c.logger.Info("Connection closing", "reason", reason)
	return c.Stop()
}

// ... implement remaining methods for sending/receiving data
