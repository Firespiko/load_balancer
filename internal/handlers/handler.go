package handlers

import "load_balancer/internal/serverpool"

type Handler struct {
	Pool *serverpool.ServerPool
}
