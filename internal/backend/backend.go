package backend

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type CircuitState int

const (
	Closed CircuitState = iota
	Open
	HalfOpen
)

type Backend struct {
	URL             *url.URL
	Weight          int
	Alive           bool
	mux             sync.RWMutex
	Reverseproxy    *httputil.ReverseProxy
	Stats           MonitoringStats
	FailureCount    int
	CircuitState    CircuitState
	LastFailureTime time.Time
	Maintenance     bool
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) SetMaintenance(maintenance bool) {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.Maintenance = maintenance
}

func (b *Backend) IsAvailable() bool {

	b.mux.RLock()
	defer b.mux.RUnlock()

	return b.Alive && !b.Maintenance
}
