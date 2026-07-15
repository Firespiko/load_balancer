package handlers

import (
	"bytes"
	"load_balancer/internal/config"
	"load_balancer/internal/serverpool"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestHandler() *Handler {
	pool := &serverpool.ServerPool{}

	return &Handler{
		Pool: pool,
	}
}

func TestGetBackends(t *testing.T) {
	h := newTestHandler()

	h.Pool.AddBackend(config.BackendConfig{
		Url:    "http://localhost:8001",
		Weight: 1,
	})

	req := httptest.NewRequest(http.MethodGet, "/backends", nil)

	rec := httptest.NewRecorder()

	h.BackendsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatal("Expected Status OK")
	}

}

func TestPostBackend(t *testing.T) {
	h := newTestHandler()

	body := bytes.NewBufferString(`{
		"url":"http://localhost:8001",
		"weight":2
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/backends",
		body,
	)

	rec := httptest.NewRecorder()

	h.BackendsHandler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	backends := h.Pool.ListBackends()

	if len(backends) != 1 {
		t.Fatalf(
			"expected 1 backend got %d",
			len(backends),
		)
	}

	if backends[0].Url != "http://localhost:8001" {
		t.Fatal("incorrect backend url")
	}

	if backends[0].Weight != 2 {
		t.Fatal("incorrect backend weight")
	}
}

func TestDeleteBackend(t *testing.T) {

	h := newTestHandler()

	h.Pool.AddBackend(config.BackendConfig{
		Url:    "http://localhost:8001",
		Weight: 1,
	})

	req := httptest.NewRequest(
		http.MethodDelete,
		"/backends?url=http://localhost:8001",
		nil,
	)

	rec := httptest.NewRecorder()

	h.BackendsHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf(
			"expected %d got %d",
			http.StatusNoContent,
			rec.Code,
		)
	}

	if len(h.Pool.ListBackends()) != 0 {
		t.Fatal("backend was not deleted")
	}
}

func TestMaintenance(t *testing.T) {

	h := newTestHandler()

	h.Pool.AddBackend(config.BackendConfig{
		Url:    "http://localhost:8001",
		Weight: 1,
	})

	body := bytes.NewBufferString(`{
		"url":"http://localhost:8001",
		"maintenance":true
	}`)

	req := httptest.NewRequest(
		http.MethodPut,
		"/backends/maintenance",
		body,
	)

	rec := httptest.NewRecorder()

	h.MaintenanceHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	backends := h.Pool.ListBackends()

	if len(backends) != 1 {
		t.Fatal("expected one backend")
	}

	backends = h.Pool.ListBackends()
	if !backends[0].Maintenance {
		t.Fatal("maintenance mode not enabled")
	}
}

func TestInvalidMethod(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(
		http.MethodPatch,
		"/backends",
		nil,
	)

	rec := httptest.NewRecorder()

	h.BackendsHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected %d got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestInvalidJSON(t *testing.T) {

	h := newTestHandler()

	body := bytes.NewBufferString(`{
		"url":
	`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/backends",
		body,
	)

	rec := httptest.NewRecorder()

	h.BackendsHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestDeleteMissingURL(t *testing.T) {

	h := newTestHandler()

	req := httptest.NewRequest(
		http.MethodDelete,
		"/backends",
		nil,
	)

	rec := httptest.NewRecorder()

	h.BackendsHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestGetEmptyBackends(t *testing.T) {

	h := newTestHandler()

	req := httptest.NewRequest(
		http.MethodGet,
		"/backends",
		nil,
	)

	rec := httptest.NewRecorder()

	h.BackendsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	expected := "[]\n"

	if rec.Body.String() != expected {
		t.Fatalf(
			"expected %q got %q",
			expected,
			rec.Body.String(),
		)
	}
}
