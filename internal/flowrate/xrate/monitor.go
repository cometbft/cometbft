package xrate

import (
	"time"

	"golang.org/x/time/rate"

	legacyimpl "github.com/cometbft/cometbft/internal/flowrate/legacy"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

type Monitor struct {
	limiter *rate.Limiter

	// rate.Limiter configuration
	burst int
	rate  int64

	// Tokens reserved but not yet used
	reserved int

	statusMtx cmtsync.Mutex
	status    legacyimpl.Status
}

func New(_, _ time.Duration) *Monitor {
	return &Monitor{
		limiter: rate.NewLimiter(rate.Inf, 0),
	}
}

func (m *Monitor) Limit(want int, maxRate int64, block bool) (n int) {
	// Update rate limiter parameters, if needed
	// XXX: this should never change after the first call
	if maxRate != m.rate {
		m.limiter.SetLimit(rate.Limit(maxRate))
		m.rate = maxRate
	}
	if want > m.burst {
		m.limiter.SetBurst(want)
		m.burst = want
	}

	var sleepTime time.Duration
	for want > m.reserved && block {
		tokens := want - m.reserved
		// We cannot reserve more than m.limiter.Burst()
		if tokens > m.burst {
			tokens = m.burst
		}

		res := m.limiter.ReserveN(time.Now(), tokens)
		if !res.OK() {
			panic("ReserveN returned false")
		}

		// Requested tokens are available after res.Delay()
		delay := res.Delay()
		if delay > 0 {
			start := time.Now()
			time.Sleep(delay)
			sleepTime += time.Since(start)
		}

		// Now we have the tokens
		m.reserved += tokens
	}

	if sleepTime > 0 {
		m.addSleepTime(sleepTime)
	}

	return m.reserved
}

func (m *Monitor) Update(n int) int {
	// We consumed part of reserved tokens
	m.reserved -= n
	return n
}

func (*Monitor) SetREMA(_ float64) {
	panic("SetREMA unimplemented by flowrate/xrate package")
}

func (m *Monitor) addSleepTime(duration time.Duration) {
	m.statusMtx.Lock()
	m.status.SleepTime += duration
	m.statusMtx.Unlock()
}

func (m *Monitor) Status() legacyimpl.Status {
	m.statusMtx.Lock()
	status := m.status // copy
	m.status.SleepTime = 0
	m.statusMtx.Unlock()
	return status
}
