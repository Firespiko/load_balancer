package constants

const (
	Attempts int = iota
	Retry
)

type contextKey string

const BackendKey contextKey = "backend"

type Algorithm string

const (
	RoundRobinAlgo         Algorithm = "round_robin"
	WeightedRoundRobinAlgo Algorithm = "weighted_round_robin"
	LeastConnectionsAlgo   Algorithm = "least_connections"
	RandomWeightAlgo       Algorithm = "random_weight"
	IPHashAlgo             Algorithm = "ip_hash"
)
