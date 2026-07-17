package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestMetricsEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	rec := httptest.NewRecorder()

	handler := promhttp.Handler()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected %d got %d", http.StatusOK, rec.Code)
	}
}

func TestRequestsMetric(t *testing.T) {

	RequestsTotal.Inc()

	req := httptest.NewRequest(
		http.MethodGet,
		"/metrics",
		nil,
	)

	rec := httptest.NewRecorder()

	promhttp.Handler().ServeHTTP(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "requests_total") {
		t.Fatal("requests_total metric missing")
	}
}

func TestFailedMetric(t *testing.T) {

	FailedRequests.Inc()

	req := httptest.NewRequest(
		http.MethodGet,
		"/metrics",
		nil,
	)

	rec := httptest.NewRecorder()

	promhttp.Handler().ServeHTTP(rec, req)

	if !strings.Contains(
		rec.Body.String(),
		"failed_requests",
	) {
		t.Fatal("failed_requests metric missing")
	}
}

func TestActiveConnectionsMetric(t *testing.T) {

	ActiveConnections.Inc()

	req := httptest.NewRequest(
		http.MethodGet,
		"/metrics",
		nil,
	)

	rec := httptest.NewRecorder()

	promhttp.Handler().ServeHTTP(rec, req)

	if !strings.Contains(
		rec.Body.String(),
		"active_connections",
	) {
		t.Fatal("active_connections metric missing")
	}
}

func TestLatencyMetric(t *testing.T) {

	RequestLatency.Observe(0.25)

	req := httptest.NewRequest(
		http.MethodGet,
		"/metrics",
		nil,
	)

	rec := httptest.NewRecorder()

	promhttp.Handler().ServeHTTP(rec, req)

	if !strings.Contains(
		rec.Body.String(),
		"request_latency",
	) {
		t.Fatal("latency metric missing")
	}
}
