package serverpool

import (
	"load_balancer/internal/backend"
	"log"
	"net/url"
	"sync/atomic"
)

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}

}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := backend.IsBackendAlive(b.URL)
		b.SetAlive(alive)
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
