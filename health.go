package main

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

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
