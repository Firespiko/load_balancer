package serverpool

import "load_balancer/internal/backend"

func (s *ServerPool) RemoveBackend(rawurl string) {
	s.mux.Lock()
	defer s.mux.Unlock()

	filtered := make([]*backend.Backend, 0)

	for _, backend := range s.backends {
		if backend.URL.String() == rawurl {
			continue
		}

		filtered = append(filtered, backend)
	}

	s.backends = filtered
	s.rebuildWeightedBackends()
}

func (s *ServerPool) rebuildWeightedBackends() {
	s.weightedBackends = nil

	for _, backend := range s.backends {
		for i := 0; i < backend.Weight; i++ {
			s.weightedBackends = append(s.weightedBackends, backend)
		}
	}

}
