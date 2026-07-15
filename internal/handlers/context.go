package handlers

import (
	"load_balancer/internal/constants"
	"net/http"
)

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(constants.Retry).(int); ok {
		return retry
	}
	return 0
}

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(constants.Attempts).(int); ok {
		return attempts
	}
	return 0
}
