package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&algorithm, "algorithm", "round_robin", "Choose the algorithm")
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
		serverPool.AlgorithmAssigner(config.Algorithm)

		for _, backendCfg := range config.Backends {

			addBackend(backendCfg)
		}

	} else {

		if len(serverList) == 0 {
			log.Fatal("Please provide one or more backends")
		}

		tokens := strings.Split(serverList, ",")
		serverPool.AlgorithmAssigner(algorithm)

		for _, tok := range tokens {

			addBackend(BackendConfig{
				Url:    tok,
				Weight: 1,
			})
		}

	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", lb)
	mux.Handle("/metrics", promhttp.Handler())

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go healthCheck(healthInterval)

	log.Printf("Load Balancer started at :%d\n", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
