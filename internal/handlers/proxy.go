package handlers

import (
	"context"
	"load_balancer/internal/constants"
	"load_balancer/internal/metrics"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

func (h *Handler) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%(%s) Manimu number of Attempts Exceeded\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not Available", http.StatusServiceUnavailable)
		return
	}

	nextPeer := h.Pool.GetNextPeer(r)

	if nextPeer == nil {
		http.Error(
			w,
			"Service is not available at the moment",
			http.StatusServiceUnavailable,
		)
		return
	}

	nextPeer.Reverseproxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		atomic.AddUint64(&nextPeer.Stats.FailedRequests, 1)
		metrics.FailedRequests.Inc()

		log.Printf("[%s] %s\n", nextPeer.URL.Host, e.Error())

		retries := GetRetryFromContext(request)
		if retries < 3 {

			delay := 10 * time.Millisecond * time.Duration(1<<retries)
			select {
			case <-time.After(delay):
				ctx := context.WithValue(request.Context(), constants.Retry, retries+1)
				nextPeer.Reverseproxy.ServeHTTP(writer, request.WithContext(ctx))
			}
			return
		}

		h.Pool.MarkBackendStatus(nextPeer.URL, false)

		attempts := GetAttemptsFromContext(request)
		log.Printf("%s(%s) Attempting Retry %d\n",
			request.RemoteAddr,
			request.URL.Path,
			attempts,
		)

		ctx := context.WithValue(request.Context(), constants.Attempts, attempts+1)
		h.ProxyHandler(writer, request.WithContext(ctx))
	}

	if nextPeer != nil {
		// increasing request served
		atomic.AddUint64(&nextPeer.Stats.RequestsServed, 1)
		metrics.RequestsTotal.Inc()
		// increasing active connection
		atomic.AddInt64(&nextPeer.Stats.ActiveConnection, 1)
		metrics.ActiveConnections.Inc()
		// defer pushes the code into the stack and executes at the last
		defer atomic.AddInt64(&nextPeer.Stats.ActiveConnection, -1)
		defer metrics.ActiveConnections.Dec()
		ctx, cancel := context.WithTimeout(
			r.Context(),
			h.Pool.RequestTimeout,
		)

		defer cancel()
		start := time.Now()
		nextPeer.Reverseproxy.ServeHTTP(w, r.WithContext(ctx))
		latency := time.Since(start)
		metrics.RequestLatency.Observe(latency.Seconds())
		nextPeer.AddLatency(latency)
		return
	}
	http.Error(w, "Service is not available at the moment", http.StatusServiceUnavailable)

}
