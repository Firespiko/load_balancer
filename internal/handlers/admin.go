package handlers

import (
	"encoding/json"
	"load_balancer/internal/config"
	"net/http"
)

func (h *Handler) BackendsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		configs := h.Pool.ListBackends()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(configs)

	case http.MethodPost:
		var cfg config.BackendConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusBadRequest,
			)
		}
		h.Pool.AddBackend(cfg)
		w.WriteHeader(http.StatusCreated)

	case http.MethodDelete:

		url := r.URL.Query().Get("url")

		if url == "" {

			http.Error(
				w,
				"Missing url parameter",
				http.StatusBadRequest,
			)

			return
		}

		h.Pool.RemoveBackend(url)

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(
			w,
			"Method Not Allowed",
			http.StatusMethodNotAllowed,
		)
	}

}
