package serverpool

import (
	"load_balancer/internal/backend"
	"load_balancer/internal/config"
	"log"
	"net/http/httputil"
	"net/url"
)

func (s *ServerPool) registerBackend(backend *backend.Backend) {
	s.backends = append(s.backends, backend)

	for i := 0; i < backend.Weight; i++ {
		s.weightedBackends = append(s.weightedBackends, backend)
	}
}

func (s *ServerPool) AddBackend(cfg config.BackendConfig) {
	s.mux.Lock()
	defer s.mux.Unlock()

	serverURL, err := url.Parse(cfg.Url)
	if err != nil {
		log.Fatal(err)
	}
	rp := httputil.NewSingleHostReverseProxy(serverURL)
	if cfg.Weight <= 0 {
		cfg.Weight = 1
	}

	backend := &backend.Backend{
		URL:          serverURL,
		Weight:       cfg.Weight,
		Alive:        true,
		Reverseproxy: rp,
	}

	s.registerBackend(backend)

	log.Printf("Configured server: %s\n", serverURL)
}
