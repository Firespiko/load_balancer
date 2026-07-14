package serverpool

import (
	"load_balancer/internal/backend"
	"math/rand"
	"net/http"
)

func (s *ServerPool) RandomWeighted() *backend.Backend {

	totalWeight := 0

	for _, backend := range s.backends {
		if backend.IsAvailable() {
			totalWeight += backend.Weight
		}
	}

	if totalWeight == 0 {
		return nil
	}

	r := rand.Intn(totalWeight)

	sum := 0

	for _, backend := range s.backends {
		if !backend.IsAvailable() {
			continue
		}

		sum += backend.Weight

		if r < sum {
			return backend
		}
	}
	return nil
}

func RandomWeightedScheduler(s *ServerPool, _ *http.Request) *backend.Backend {
	return s.RandomWeighted()
}
