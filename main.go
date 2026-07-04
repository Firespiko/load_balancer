package main

import "fmt"


type Backend Struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	reverseproxy *httputil.ReverseProxy
}

type Serverpool Struct {
	backends []*Backend
	current uint64
}


u,_ := url.Parse("http://localhost:8080")
rp = httputil.NewSingleHostReverseProxy(u)

http.HandlerFunc(rp.serveHTTP)

func (s *ServerPool) NextIndex() int {
	return int(atomic64.AddUint64(&s.current,Uint64(1)) % Uint64(len(s.backends)))
}


