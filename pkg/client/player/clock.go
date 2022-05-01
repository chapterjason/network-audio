package player

import (
	"sync"
	"time"

	"network-audio/pkg/messages"
	"network-audio/pkg/timex"
)

type Clock struct {
	lock    *sync.RWMutex
	latency time.Duration
}

func NewClock(latency time.Duration) *Clock {
	return &Clock{
		latency: latency,
		lock:    &sync.RWMutex{},
	}
}

func (c *Clock) GetLatency() time.Duration {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.latency
}

func (c *Clock) Now() time.Time {
	return time.Now().Add(-c.GetLatency())
}

func (c *Clock) UpdateLatency(m *messages.Latency) time.Duration {
	c.lock.Lock()
	defer c.lock.Unlock()

	sent := timex.ToTime(m.Time)
	received := time.Now()

	c.latency = (received.Sub(sent) + time.Duration(m.Latency)) / 2

	return c.latency
}
