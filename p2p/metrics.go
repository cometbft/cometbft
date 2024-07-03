package p2p

import (
	"fmt"
	"reflect"
	"regexp"
	"sync"

	"github.com/go-kit/kit/metrics"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this
	// package.
	MetricsSubsystem = "p2p"
)

// valueToLabelRegexp is used to find the golang package name and type name
// so that the name can be turned into a prometheus label where the characters
// in the label do not include prometheus special characters such as '*' and '.'.
var valueToLabelRegexp = regexp.MustCompile(`\*?(\w+)\.(.*)`)

//go:generate go run ../scripts/metricsgen -struct=Metrics

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// Number of peers.
	Peers metrics.Gauge
	// Pending bytes to be sent to a given peer.
	PeerPendingSendBytes metrics.Gauge `metrics_labels:"peer_id"`
	// Number of transactions submitted by each peer.
	NumTxs metrics.Gauge `metrics_labels:"peer_id"`
	// Number of bytes of each message type received.
	MessageReceiveBytesTotal metrics.Counter `metrics_labels:"message_type"`
	// Number of bytes of each message type sent.
	MessageSendBytesTotal metrics.Counter `metrics_labels:"message_type"`
}

type metricsLabelCache struct {
	mtx               *sync.RWMutex
	messageLabelNames map[reflect.Type]string
}

type peerPendingMetricsCache struct {
	mtx             *sync.Mutex
	perMessageCache map[reflect.Type]peerPendingMetricsCacheEntry
}

type peerPendingMetricsCacheEntry struct {
	label            string
	pendingSendBytes int
	pendingRecvBytes int
}

func peerPendingMetricsCacheFromMlc(mlc *metricsLabelCache) *peerPendingMetricsCache {
	pendingCache := &peerPendingMetricsCache{
		mtx:             &sync.Mutex{},
		perMessageCache: make(map[reflect.Type]peerPendingMetricsCacheEntry),
	}
	if mlc != nil {
		mlc.mtx.RLock()
		for k, v := range mlc.messageLabelNames {
			pendingCache.perMessageCache[k] = peerPendingMetricsCacheEntry{label: v}
		}
		mlc.mtx.RUnlock()
	}
	return pendingCache
}

func (c *peerPendingMetricsCache) AddPendingSendBytes(msgType reflect.Type, addBytes int) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if entry, ok := c.perMessageCache[msgType]; ok {
		entry.pendingSendBytes += addBytes
		c.perMessageCache[msgType] = entry
	} else {
		c.perMessageCache[msgType] = peerPendingMetricsCacheEntry{
			label:            buildLabel(msgType),
			pendingSendBytes: addBytes,
		}
	}
}

func (c *peerPendingMetricsCache) AddPendingRecvBytes(msgType reflect.Type, addBytes int) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if entry, ok := c.perMessageCache[msgType]; ok {
		entry.pendingRecvBytes += addBytes
		c.perMessageCache[msgType] = entry
	} else {
		c.perMessageCache[msgType] = peerPendingMetricsCacheEntry{
			label:            buildLabel(msgType),
			pendingRecvBytes: addBytes,
		}
	}
}

func getMsgType(i any) reflect.Type {
	return reflect.TypeOf(i)
}

// ValueToMetricLabel is a method that is used to produce a prometheus label value of the golang
// type that is passed in.
// This method uses a map on the Metrics struct so that each label name only needs
// to be produced once to prevent expensive string operations.
func (m *metricsLabelCache) ValueToMetricLabel(i any) string {
	t := getMsgType(i)
	m.mtx.RLock()

	if s, ok := m.messageLabelNames[t]; ok {
		m.mtx.RUnlock()
		return s
	}
	m.mtx.RUnlock()

	l := buildLabel(t)
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.messageLabelNames[t] = l
	return l
}

func buildLabel(msgType reflect.Type) string {
	s := msgType.String()
	ss := valueToLabelRegexp.FindStringSubmatch(s)
	return fmt.Sprintf("%s_%s", ss[1], ss[2])
}

func newMetricsLabelCache() *metricsLabelCache {
	return &metricsLabelCache{
		mtx:               &sync.RWMutex{},
		messageLabelNames: map[reflect.Type]string{},
	}
}
