package main

import (
	"hash/fnv"
	"net"
	"net/http"
)

func (s *ServerPool) IpHash(r *http.Request) *Backend {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	hasher := fnv.New32a()
	hasher.Write([]byte(host))

	start := int(hasher.Sum32()) % len(s.backends)
	l := len(s.backends) + start

	for i := start; i < l; i++ {

		idx := i % len(s.backends)

		if s.backends[idx].IsAlive() {
			return s.backends[idx]
		}
	}
	return nil
}

func IPHashScheduler(s *ServerPool, r *http.Request) *Backend {
	return s.IpHash(r)
}
