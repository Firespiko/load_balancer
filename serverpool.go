package main

import (
	"log"
	"net/url"
	"sync/atomic"
)

type ServerPool struct {
	backends         []*Backend
	weightedBackends []*Backend
	current          uint64
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)

	for i := 0; i < backend.Weight; i++ {
		s.weightedBackends = append(s.weightedBackends, backend)
	}
}

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			b.setAlive(alive)
			break
		}
	}

}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.setAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
		avgLatency := b.AverageLatency()
		log.Printf(
			"Requests=%d Failed=%d Active=%d AvgLatency=%v\n",
			atomic.LoadUint64(&b.Stats.RequestsServed),
			atomic.LoadUint64(&b.Stats.FailedRequests),
			atomic.LoadInt64(&b.Stats.ActiveConnection),
			avgLatency)
	}
}
