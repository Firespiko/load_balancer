package main

import (
	"context"
	"flag"
	"os"
	"strings"

	// "flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	Attempts int = iota
	Retry
)

type MonitoringStats struct {
	RequestsServed   uint64
	FailedRequests   uint64
	ActiveConnection int64

	TotalLatency time.Duration
	mux          sync.Mutex
}

type Backend struct {
	URL          *url.URL
	Weight       int
	Alive        bool
	mux          sync.RWMutex
	reverseproxy *httputil.ReverseProxy
	Stats        MonitoringStats
}

type ServerPool struct {
	backends         []*Backend
	weightedBackends []*Backend
	current          uint64
}

type BackendConfig struct {
	Url        string `yaml:"url"`
	Weight     uint64 `yaml:"weight"`
	HealthPath string `yaml:"health_path"`
}

type Config struct {
	Port           int64           `yaml:"port"`
	HealthInterval time.Duration   `yaml:"health_interval"`
	Algorithm      string          `yaml:"algorithm"`
	Backends       []BackendConfig `yaml:"backends"`
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			b.setAlive(alive)
			break
		}
	}

}

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) GetNextPeer() *Backend {
	next := s.NextIndex()
	l := len(s.backends) + next
	// this essentially takes the current index lets say 1 out of 4 and adds the total length 4 which makes it 5 so we have to do 4 iterations

	for i := next; i < l; i++ {
		idx := i % (len(s.backends))
		// the above idx stores the current backend it is modded since if we start from 2 out of 4 and complete till 4 and i becomes 5 then we need it to point to first backend
		if s.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}

	return nil

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

func (b *Backend) setAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return alive
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", config)
	return &config, err
}

func isBackendAlive(u *url.URL) bool {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	healthURL := *u
	healthURL.Path = "/health"
	resp, err := client.Get(healthURL.String())

	if err != nil {
		log.Println("Site Unreachable err: ", err)
		return false
	}

	if resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()
	return true
}

func (b *Backend) AverageLatency() time.Duration {
	requests := atomic.LoadUint64(&b.Stats.RequestsServed)

	if requests == 0 {
		return 0
	}

	b.Stats.mux.Lock()
	total := b.Stats.TotalLatency
	b.Stats.mux.Unlock()

	return total / time.Duration(requests)
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.setAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
		avgLatency := b.AverageLatency()
		log.Printf(
			"Requests=%d Failed=%d Active=%d AvgLatency=%v\n",
			atomic.LoadUint64(&b.Stats.RequestsServed),
			atomic.LoadUint64(&b.Stats.FailedRequests),
			atomic.LoadInt64(&b.Stats.ActiveConnection),
			avgLatency)
	}
}

func healthCheck(healthCheckInterval time.Duration) {
	t := time.NewTicker(healthCheckInterval)
	for {
		select {
		case <-t.C:
			log.Println("Starting Health Check.... \n")
			serverPool.HealthCheck()
			log.Println("Completed Health Check!!!\n")
		}
	}

}

func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%(%s) Manimu number of Attempts Exceeded\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not Available", http.StatusServiceUnavailable)
		return
	}

	nextPeer := serverPool.GetNextPeer()

	if nextPeer != nil {
		// increasing request served
		atomic.AddUint64(&nextPeer.Stats.RequestsServed, 1)
		// increasing active connection
		atomic.AddInt64(&nextPeer.Stats.ActiveConnection, 1)
		// defer pushes the code into the stack and executes at the last
		defer atomic.AddInt64(&nextPeer.Stats.ActiveConnection, -1)
		start := time.Now()
		nextPeer.reverseproxy.ServeHTTP(w, r)
		latency := time.Since(start)
		nextPeer.Stats.mux.Lock()
		nextPeer.Stats.TotalLatency = nextPeer.Stats.TotalLatency + latency
		nextPeer.Stats.mux.Unlock()
		return
	}
	http.Error(w, "Service is not available at the moment", http.StatusServiceUnavailable)

}

func addBackend(serverURL *url.URL) {
	rp := httputil.NewSingleHostReverseProxy(serverURL)

	backend := &Backend{
		URL:          serverURL,
		Alive:        true,
		reverseproxy: rp,
	}

	rp.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		atomic.AddUint64(&backend.Stats.FailedRequests, 1)

		log.Printf("[%s] %s\n", serverURL.Host, e.Error())

		retries := GetRetryFromContext(request)
		if retries < 3 {
			select {
			case <-time.After(10 * time.Millisecond):
				ctx := context.WithValue(request.Context(), Retry, retries+1)
				rp.ServeHTTP(writer, request.WithContext(ctx))
			}
			return
		}

		serverPool.MarkBackendStatus(serverURL, false)

		attempts := GetAttemptsFromContext(request)
		log.Printf("%s(%s) Attempting Retry %d\n",
			request.RemoteAddr,
			request.URL.Path,
			attempts,
		)

		ctx := context.WithValue(request.Context(), Attempts, attempts+1)
		lb(writer, request.WithContext(ctx))
	}

	serverPool.AddBackend(backend)

	log.Printf("Configured server: %s\n", serverURL)
}

var serverPool ServerPool

func main() {
	var configPath string
	var serverList string
	var port int
	var healthInterval time.Duration

	flag.StringVar(&configPath, "config", "", "Path to config file")

	flag.StringVar(&serverList, "backends", "", "Load balanced backends")
	flag.IntVar(&port, "port", 3030, "Port to serve")
	flag.DurationVar(
		&healthInterval,
		"health-check-interval",
		20*time.Second,
		"Backend health check interval",
	)

	flag.Parse()

	if configPath != "" {

		config, err := LoadConfig(configPath)
		if err != nil {
			log.Fatal(err)
		}

		port = int(config.Port)
		healthInterval = config.HealthInterval

		for _, backendCfg := range config.Backends {

			serverURL, err := url.Parse(backendCfg.Url)
			if err != nil {
				log.Fatal(err)
			}

			addBackend(serverURL)
		}

	} else {

		if len(serverList) == 0 {
			log.Fatal("Please provide one or more backends")
		}

		tokens := strings.Split(serverList, ",")

		for _, tok := range tokens {

			serverURL, err := url.Parse(tok)
			if err != nil {
				log.Fatal(err)
			}

			addBackend(serverURL)
		}
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	go healthCheck(healthInterval)

	log.Printf("Load Balancer started at :%d\n", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
