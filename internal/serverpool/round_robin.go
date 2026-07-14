package serverpool

import (
	"load_balancer/internal/backend"
	"net/http"
	"sync/atomic"
)

func (s *ServerPool) RoundRobin() *backend.Backend {
	next := int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))

	l := len(s.backends) + next

	for i := next; i < l; i++ {
		idx := i % len(s.backends)

		if s.backends[idx].IsAvailable() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]

		}
	}
	return nil
}

func RoundRobinScheduler(s *ServerPool, _ *http.Request) *backend.Backend {
	return s.RoundRobin()
}
