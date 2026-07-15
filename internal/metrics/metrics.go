package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var RequestsTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "requests_total",
		Help: "Total Number of Requests Served",
	},
)

var FailedRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "failed_requests_total",
		Help: "Total Number of Failed Requests",
	},
)

var ActiveConnections = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "active_connections",
		Help: "Current active backend connections.",
	},
)

var RequestLatency = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "request_latency_seconds",
		Help:    "Latency of backend requests.",
		Buckets: prometheus.DefBuckets,
	},
)

func init() {

	prometheus.MustRegister(RequestsTotal)
	prometheus.MustRegister(FailedRequests)
	prometheus.MustRegister(ActiveConnections)
	prometheus.MustRegister(RequestLatency)

}
