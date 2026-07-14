package serverpool

import (
	"load_balancer/internal/backend"
	"sync"
	"time"
)

type ServerPool struct {
	mux              sync.RWMutex
	backends         []*backend.Backend
	weightedBackends []*backend.Backend
	current          uint64
	Scheduler        Scheduler
	RequestTimeout   time.Duration
}
