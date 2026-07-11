package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var serverPool ServerPool

func main() {
	var configPath string
	var serverList string
	var port int
	var healthInterval time.Duration
	var algorithm string
	var requestTimeout time.Duration

	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&algorithm, "algorithm", "round_robin", "Choose the algorithm")
	flag.StringVar(&serverList, "backends", "", "Load balanced backends")
	flag.DurationVar(&requestTimeout, "timeout", 5*time.Second, "Request for Timeout")
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
		serverPool.AlgorithmAssigner(config.Algorithm)

		for _, backendCfg := range config.Backends {

			serverPool.AddBackend(backendCfg)
		}

	} else {

		if len(serverList) == 0 {
			log.Fatal("Please provide one or more backends")
		}

		tokens := strings.Split(serverList, ",")
		serverPool.AlgorithmAssigner(algorithm)

		for _, tok := range tokens {

			serverPool.AddBackend(BackendConfig{
				Url:    tok,
				Weight: 1,
			})
		}

	}
	mux := http.NewServeMux()
	mux.Handle("/", LoggingMiddleware(http.HandlerFunc(lb)))
	mux.Handle("/metrics", LoggingMiddleware(promhttp.Handler()))
	mux.Handle("/backends", LoggingMiddleware(http.HandlerFunc(BackendsHandler)))
	mux.Handle("/backends/maintenance", LoggingMiddleware(http.HandlerFunc(MaintenanceHandler)))

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		os.Interrupt,
		syscall.SIGTERM,
	)

	go healthCheck(healthInterval)

	log.Printf("Load Balancer started at :%d\n", port)

	go func() {

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			log.Fatal(err)

		}

	}()

	<-quit
	log.Println("Shutting down server....")
	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)

	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server exited gracefully.")
}
