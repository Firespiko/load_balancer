package main

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
	reverseproxy    *httputil.ReverseProxy
	Stats           MonitoringStats
	FailureCount    int
	CircuitState    CircuitState
	LastFailureTime time.Time
}

func (b *Backend) setAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return alive
}
