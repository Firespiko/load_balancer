package backend

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

func IsBackendAlive(u *url.URL) bool {
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
