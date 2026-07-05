package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
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
	current uint64
}


func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current,uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) GetNextPeer() *Backend {
	next := s.NextIndex()
	l := len(s.backends) + next
	// this essentially takes the current index lets say 1 out of 4 and adds the total length 4 which makes it 5 so we have to do 4 iterations
	
	for i := next; i < l; i++ {
		idx := i % (len(s.backends))
		// the above idx stores the current backend it is modded since if we start from 2 out of 4 and complete till 4 and i becomes 5 then we need it to point to first backend
		if s.backends[idx].IsAlive() {
			if i != next{
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}

	return nil

}

func GetRetryFromContext(r *http.Request) int {
	if retry,ok := r.Context.Value(Retry).(int); ok {
		return retry
	}
	return 0
}

func GetAttemptsFromContext(r *http.Request) int {
	if attempts,ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	} 
	return 0
}

func (b *Backend) setAlive(alive bool) {
	b.mux.Lock()
	b.Alive := alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive := b.Alive
	b.mux.RUnlock()
}

func isBackendAlive(u *url.URL) bool {
	duration := 2 * time.Second
	conn,err := net.DialTimeout("tcp",u.Host,timeout)
	if err != nil {
		log.println("Site Unreachable err: ",err)
		return false
	}
	_ = conn.Close()
	return true
}

func (s *ServerPool) HealthCheck() {
	for _,b := range s.backends {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.setAlive(alive)
		if !alive {
			status := "down"
		}
		log.Printf("%s [%s]\n", b.URL,status)
	}
}


func healthCheck() {
	t := time.NewTicket(time.Second * 20)
	for {
		select {
			case <- t.C:
				log.println("Starting Health Check.... \n")
				serverPool.HealthCheck()
				log.Println("Completed Health Check!!!\n")
		}
	}

}

func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttempsFromContext(request)
	if attempts > 3 {
		log.printf("%(%s) Manimu number of Attempts Exceeded\n",r.RemoteAddr,r.URL.path)
		http.Error(w, "Service not Available", http.StatusServiceUnavailable)
		return
	}

	nextPeer := serverPool.GetNextPeer()

	if nextPeer != nil {
		nextPeer.reverseproxy.ServeHTTP(w,r)
		return
	}
	http.Error(w, "Service is not available at the moment", http.StatusServiceUnavailable)

}


func main() {
	u, _ := url.Parse("http://localhost:8080")
  rp := httputil.NewSingleHostReverseProxy(u)

	server := http.Server {
		Addr: fmt.SprintF(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	rp.ErrorHandler = func(writer http.ResponseWriter, request *http.Response, e error) {
		log.printf("[%s] %s\n", u.host,e.Error())
		retries := GetRetryFromContext(request)
		if retries < 3 {
			select {
				case <- time.after(10 * time.MilliSecond):
					ctx := context.WithValue(request.Context(),Retry,retries + 1)
					rp.serverHTTP(writer, request.WithContext(ctx))
			}
			return
		}

		server.MarkBackendStatus(u,false)

		attempts := GetAttemptsFromContext(request)
		log.printf("%s(%s) Attempting Retry %d \n",request.RemoteAddr,request.URL.path,attempts)
		ctx := context.WithValue(request.Context(), Attempts, attempts + 1)
		lb(writer, request.WithContext(ctx))
	} 


}
