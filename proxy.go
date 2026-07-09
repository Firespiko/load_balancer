package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

func addBackend(cfg BackendConfig) {
	serverURL, err := url.Parse(cfg.Url)
	if err != nil {
		log.Fatal(err)
	}
	rp := httputil.NewSingleHostReverseProxy(serverURL)
	if cfg.Weight <= 0 {
		cfg.Weight = 1
	}

	backend := &Backend{
		URL:          serverURL,
		Weight:       cfg.Weight,
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
