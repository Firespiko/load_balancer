package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"golang.org/x/time/rate"
)

type RequestLog struct {
	Timestamp string `json:"timestamp"`
	ClientIP  string `json:"client_ip"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := rate.NewLimiter(10, 20)

		if !limiter.Allow() {
			http.Error(
				w,
				"Too Many Requests",
				http.StatusTooManyRequests,
			)
			return
		}

		start := time.Now()

		lrw := &LoggingResponseWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		next.ServeHTTP(lrw, r)

		latency := time.Since(start)

		entry := RequestLog{
			Timestamp: time.Now().Format(time.RFC3339),
			ClientIP:  r.RemoteAddr,
			Method:    r.Method,
			Path:      r.URL.Path,
			Status:    lrw.StatusCode,
			LatencyMS: latency.Milliseconds(),
		}

		json.NewEncoder(os.Stdout).Encode(entry)

	})
}
