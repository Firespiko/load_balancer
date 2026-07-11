package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

func BackendsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		configs := serverPool.ListBackends()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(configs)

	case http.MethodPost:
		var cfg BackendConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusBadRequest,
			)
		}
		serverPool.AddBackend(cfg)
		w.WriteHeader(http.StatusCreated)

	case http.MethodDelete:

		url := r.URL.Query().Get("url")

		if url == "" {

			http.Error(
				w,
				"Missing url parameter",
				http.StatusBadRequest,
			)

			return
		}

		serverPool.RemoveBackend(url)

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)
	}

}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 0
}

func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%(%s) Manimu number of Attempts Exceeded\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not Available", http.StatusServiceUnavailable)
		return
	}

	nextPeer := serverPool.GetNextPeer(r)

	if nextPeer != nil {
		// increasing request served
		atomic.AddUint64(&nextPeer.Stats.RequestsServed, 1)
		RequestsTotal.Inc()
		// increasing active connection
		atomic.AddInt64(&nextPeer.Stats.ActiveConnection, 1)
		ActiveConnections.Inc()
		// defer pushes the code into the stack and executes at the last
		defer atomic.AddInt64(&nextPeer.Stats.ActiveConnection, -1)
		defer ActiveConnections.Dec()
		start := time.Now()
		nextPeer.reverseproxy.ServeHTTP(w, r)
		latency := time.Since(start)
		RequestLatency.Observe(latency.Seconds())
		nextPeer.Stats.mux.Lock()
		nextPeer.Stats.TotalLatency = nextPeer.Stats.TotalLatency + latency
		nextPeer.Stats.mux.Unlock()
		return
	}
	http.Error(w, "Service is not available at the moment", http.StatusServiceUnavailable)

}
