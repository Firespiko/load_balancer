package main

import "net/http"

type Scheduler func(*ServerPool, *http.Request) *Backend

func (s *ServerPool) GetNextPeer(r *http.Request) *Backend {
	if s.Scheduler == nil {
		return nil
	}

	return s.Scheduler(s, r)
}
