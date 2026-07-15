package backend

import (
	"sync"
	"sync/atomic"
	"time"
)

type MonitoringStats struct {
	RequestsServed   uint64
	FailedRequests   uint64
	ActiveConnection int64

	TotalLatency time.Duration
	mux          sync.Mutex
}

func (b *Backend) AverageLatency() time.Duration {
	requests := atomic.LoadUint64(&b.Stats.RequestsServed)

	if requests == 0 {
		return 0
	}

	b.Stats.mux.Lock()
	total := b.Stats.TotalLatency
	b.Stats.mux.Unlock()

	return total / time.Duration(requests)
}

func (b *Backend) AddLatency(latency time.Duration) {
	b.Stats.mux.Lock()
	defer b.Stats.mux.Unlock()

	b.Stats.TotalLatency += latency
}
