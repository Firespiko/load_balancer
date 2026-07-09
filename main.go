package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

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

			addBackend(backendCfg)
		}

	} else {

		if len(serverList) == 0 {
			log.Fatal("Please provide one or more backends")
		}

		tokens := strings.Split(serverList, ",")

		for _, tok := range tokens {

			addBackend(BackendConfig{
				Url:    tok,
				Weight: 1,
			})
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
