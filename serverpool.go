package main

import (
	"log"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type ServerPool struct {
	mux              sync.RWMutex
	backends         []*Backend
	weightedBackends []*Backend
	current          uint64
	Scheduler        Scheduler
	RequestTimeout   time.Duration
}

func (s *ServerPool) registerBackend(backend *Backend) {
	s.backends = append(s.backends, backend)

	for i := 0; i < backend.Weight; i++ {
		s.weightedBackends = append(s.weightedBackends, backend)
	}
}

func (s *ServerPool) RemoveBackend(rawurl string) {
	s.mux.Lock()
	defer s.mux.Unlock()

	filtered := make([]*Backend, 0)

	for _, backend := range s.backends {
		if backend.URL.String() == rawurl {
			continue
		}

		filtered = append(filtered, backend)
	}

	s.backends = filtered
	s.rebuildWeightedBackends()
}

func (s *ServerPool) ListBackends() []BackendConfig {
	s.mux.RLock()
	defer s.mux.RUnlock()

	configs := make([]BackendConfig, 0, len(s.backends))

	for _, backend := range s.backends {
		configs = append(configs, BackendConfig{
			Url:        backend.URL.String(),
			Weight:     backend.Weight,
			HealthPath: "/health",
		})
	}
	return configs

}

func (s *ServerPool) rebuildWeightedBackends() {
	s.weightedBackends = nil

	for _, backend := range s.backends {
		for i := 0; i < backend.Weight; i++ {
			s.weightedBackends = append(s.weightedBackends, backend)
		}
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

func (s *ServerPool) SetMaintenance(
	url string,
	maintenance bool,
) {

	s.mux.Lock()
	defer s.mux.Unlock()

	for _, backend := range s.backends {

		if backend.URL.String() == url {

			backend.SetMaintenance(maintenance)

			return
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

func (s *ServerPool) AlgorithmAssigner(algorithm string) {
	switch algorithm {

	case string(RoundRobinAlgo):
		s.Scheduler = RoundRobinScheduler

	case string(WeightedRoundRobinAlgo):
		s.Scheduler = WeightedRoundRobinScheduler

	case string(LeastConnectionsAlgo):
		s.Scheduler = LeastConnectionsScheduler

	case string(RandomWeightAlgo):
		s.Scheduler = RandomWeightedScheduler

	case string(IPHashAlgo):
		s.Scheduler = IPHashScheduler

	default:
		s.Scheduler = WeightedRoundRobinScheduler
	}

}
