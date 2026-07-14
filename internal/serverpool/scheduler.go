package serverpool

import (
	"load_balancer/internal/backend"
	"load_balancer/internal/constants"
	"net/http"
)

type Scheduler func(*ServerPool, *http.Request) *backend.Backend

func (s *ServerPool) GetNextPeer(r *http.Request) *backend.Backend {
	if s.Scheduler == nil {
		return nil
	}

	return s.Scheduler(s, r)
}

func (s *ServerPool) AlgorithmAssigner(algorithm string) {
	switch algorithm {

	case string(constants.RoundRobinAlgo):
		s.Scheduler = RoundRobinScheduler

	case string(constants.WeightedRoundRobinAlgo):
		s.Scheduler = WeightedRoundRobinScheduler

	case string(constants.LeastConnectionsAlgo):
		s.Scheduler = LeastConnectionsScheduler

	case string(constants.RandomWeightAlgo):
		s.Scheduler = RandomWeightedScheduler

	case string(constants.IPHashAlgo):
		s.Scheduler = IPHashScheduler

	default:
		s.Scheduler = WeightedRoundRobinScheduler
	}

}
