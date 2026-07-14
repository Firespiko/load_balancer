package serverpool

import (
	"load_balancer/internal/backend"
	"net/http"
	"sync/atomic"
)

func (s *ServerPool) LeastConnections() *backend.Backend {
	var selected *backend.Backend
	var minConnections int64

	for _, backend := range s.backends {
		if !backend.IsAvailable() {
			continue
		}

		connections := atomic.LoadInt64(&backend.Stats.ActiveConnection)

		if selected == nil {
			selected = backend
			minConnections = connections
			continue
		}

		if connections < minConnections {
			selected = backend
			minConnections = connections
		}
	}
	return selected
}

func LeastConnectionsScheduler(s *ServerPool, _ *http.Request) *backend.Backend {
	return s.LeastConnections()
}
