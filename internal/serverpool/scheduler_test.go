package serverpool

import (
	"load_balancer/internal/backend"
	"net/http/httptest"
	"net/url"
	"testing"
)

func createBackend(rawURL string, weight int) *backend.Backend {
	u, _ := url.Parse(rawURL)

	return &backend.Backend{
		URL:    u,
		Weight: weight,
		Alive:  true,
	}
}

func TestRoundRobin(t *testing.T) {
	b1 := createBackend("https://localhost:8001", 1)
	b2 := createBackend("https://localhost:8002", 1)

	pool := ServerPool{
		backends: []*backend.Backend{
			b1,
			b2,
		},
	}

	got := pool.RoundRobin()
	if got != b2 {
		t.Fatalf("Expected Backend 2 Got %v", got.URL)
	}

	got = pool.RoundRobin()
	if got != b1 {
		t.Fatalf("Expected Backend 1 Got %v", got.URL)
	}

	got = pool.RoundRobin()
	if got != b2 {
		t.Fatalf("Expected Backend 2 Got %v", got.URL)
	}

}

func TestWeightedRoundRobin(t *testing.T) {
	b1 := createBackend("https://localhost:8001", 2)
	b2 := createBackend("https://localhost:8002", 1)

	pool := ServerPool{
		backends: []*backend.Backend{
			b1,
			b2,
		},
	}

	got := pool.RoundRobin()
	if got != b1 {
		t.Fatalf("Expected Backend 2 Got %v", got.URL)
	}

	got = pool.RoundRobin()
	if got != b2 {
		t.Fatalf("Expected Backend 1 Got %v", got.URL)
	}

	got = pool.RoundRobin()
	if got != b1 {
		t.Fatalf("Expected Backend 2 Got %v", got.URL)
	}
}

func TestLeastConnections(T *testing.T) {
	b1 := createBackend("https://localhost:8001", 1)
	b2 := createBackend("https://localhost:8002", 1)
	b3 := createBackend("https://localhost:8002", 1)

	b1.Stats.ActiveConnection = 2
	b2.Stats.ActiveConnection = 8
	b3.Stats.ActiveConnection = 5

	pool := ServerPool{
		backends: []*backend.Backend{
			b1, b2, b3,
		},
	}

	got := pool.LeastConnections()
	if got != b1 {
		T.Fatalf("Excepted b1 got %v", b1)
	}

}

func TestRandomWeighted(t *testing.T) {
	b1 := createBackend("http://localhost:8001", 1)
	b2 := createBackend("http://localhost:8002", 1)

	pool := ServerPool{
		backends: []*backend.Backend{
			b1, b2,
		},
	}

	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "192.168.1.10:12345"

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.10:17654"

	first := pool.IpHash(req1)
	second := pool.IpHash(req2)

	if first != second {
		t.Fatal("Same client IP should always point to the same background")
	}

}

func TestIPHashDifferentClients(t *testing.T) {

	b1 := createBackend("http://localhost:8001", 1)
	b2 := createBackend("http://localhost:8002", 1)

	pool := ServerPool{
		backends: []*backend.Backend{
			b1,
			b2,
		},
	}

	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "10.0.0.1:1234"

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.15:4321"

	_ = pool.IpHash(req1)
	_ = pool.IpHash(req2)

}

func TestRandomWeightDistribution(T *testing.T) {
	b1 := createBackend("http://localhost:8001", 3)
	b2 := createBackend("http://localhost:8002", 1)

	pool := ServerPool{
		backends: []*backend.Backend{
			b1,
			b2,
		},
	}

	var c1, c2 int

	for i := 0; i < 100000; i++ {
		if pool.RandomWeighted() == b1 {
			c1++
		} else {
			c2++
		}
	}

	ratio := float64(c1) / float64(c1+c2)

	if ratio < 0.70 || ratio > 0.80 {
		T.Fatal("Unexpected Distribution")
	}
}
