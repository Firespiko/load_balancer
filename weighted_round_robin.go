package main

import (
	"net/http"
	"sync/atomic"
)

func (s *ServerPool) NextIndex() int {
	if len(s.weightedBackends) == 0 {
		return -1
	}

	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.weightedBackends)))
}

func (s *ServerPool) WeightedRoundRobin() *Backend {
	next := s.NextIndex()
	if next == -1 {
		return nil
	}
	l := len(s.weightedBackends) + next
	// this essentially takes the current index lets say 1 out of 4 and adds the total length 4 which makes it 5 so we have to do 4 iterations

	for i := next; i < l; i++ {
		idx := i % (len(s.weightedBackends))
		// the above idx stores the current backend it is modded since if we start from 2 out of 4 and complete till 4 and i becomes 5 then we need it to point to first backend
		if s.weightedBackends[idx].IsAvailable() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.weightedBackends[idx]
		}
	}

	return nil

}

func WeightedRoundRobinScheduler(s *ServerPool, _ *http.Request) *Backend {
	return s.WeightedRoundRobin()
}
