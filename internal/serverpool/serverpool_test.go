package serverpool

import (
	"load_balancer/internal/config"
	"testing"
)

func TestAddBackend(t *testing.T) {
	pool := ServerPool{}

	cfg := config.BackendConfig{
		Url:    "http://localhost:8001",
		Weight: 1,
	}

	pool.AddBackend(cfg)

	if len(pool.backends) != 1 {
		t.Fatalf("Expected 1 Backend got %d", len(pool.backends))
	}

	if len(pool.weightedBackends) != 2 {
		t.Fatalf("Expected Weighted Backend Length 2, got %d", len(pool.backends))
	}
}

func TestRemoveBackend(t *testing.T) {
	pool := ServerPool{}

	pool.AddBackend(config.BackendConfig{
		Url:    "http://localhost:8001",
		Weight: 1,
	})

	pool.AddBackend(config.BackendConfig{
		Url:    "http://localhost:8002",
		Weight: 1,
	})

	pool.RemoveBackend("http://localhost:8001")

	if len(pool.backends) != 1 {
		t.Fatalf("Expected 1 backend got %d", len(pool.backends))
	}
	if pool.backends[0].URL.String() != "http://localhost:8002" {
		t.Fatal("Wrong Backend Removed")
	}
}

func TestListBackends(T *testing.T) {

}
