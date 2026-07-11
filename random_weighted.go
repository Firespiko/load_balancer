package main

import (
	"math/rand"
	"net/http"
)

func (s *ServerPool) RandomWeighted() *Backend {

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

func RandomWeightedScheduler(s *ServerPool, _ *http.Request) *Backend {
	return s.RandomWeighted()
}
