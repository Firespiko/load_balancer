package main

import (
	"context"
	"flag"
	"strings"

	// "flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	// "strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Attempts int = iota
	Retry
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	reverseproxy *httputil.ReverseProxy
}

type ServerPool struct {
	backends []*Backend
	current  uint64
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

func isBackendAlive(u *url.URL) bool {
	duration := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, duration)
	if err != nil {
		log.Println("Site Unreachable err: ", err)
		return false
	}
	_ = conn.Close()
	return true
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
	}
}

func healthCheck() {
	t := time.NewTicker(time.Second * 20)
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
		nextPeer.reverseproxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service is not available at the moment", http.StatusServiceUnavailable)

}

var serverPool ServerPool

func main() {
	var serverList string
	var port int

	flag.StringVar(&serverList, "backends", "", "Load Balanced Backends, use comma to seperate")
	flag.IntVar(&port, "Port", 3030, "Enter the port number,default: 3030")
	flag.Parse()

	if len(serverList) == 0 {
		log.Fatal("Please provide one or more backends to balance")
	}
	tokens := strings.Split(serverList, ",")
	for _, tok := range tokens {
		serverURL, err := url.Parse(tok)
		if err != nil {
			log.Fatal(err)
		}
		rp := httputil.NewSingleHostReverseProxy(serverURL)
		rp.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
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
			log.Printf("%s(%s) Attempting Retry %d \n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			lb(writer, request.WithContext(ctx))
		}

		serverPool.AddBackend(&Backend{
			URL:          serverURL,
			Alive:        true,
			reverseproxy: rp,
		})

		log.Printf("Configured Server: %s\n", port)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	go healthCheck()

	log.Printf("Load Balance Stored at :%d\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
