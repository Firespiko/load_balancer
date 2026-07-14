package serverpool

import "load_balancer/internal/config"

func (s *ServerPool) ListBackends() []config.BackendConfig {
	s.mux.RLock()
	defer s.mux.RUnlock()

	configs := make([]config.BackendConfig, 0, len(s.backends))

	for _, backend := range s.backends {
		configs = append(configs, config.BackendConfig{
			Url:        backend.URL.String(),
			Weight:     backend.Weight,
			HealthPath: "/health",
		})
	}
	return configs

}
