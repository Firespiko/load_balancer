package handlers

import (
	"encoding/json"
	"net/http"
)

type MaintenanceRequest struct {
	URL         string `json:"url"`
	Maintenance bool   `json:"maintenance"`
}

func (h *Handler) MaintenanceHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPut {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	var req MaintenanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	h.Pool.SetMaintenance(
		req.URL,
		req.Maintenance,
	)
	w.WriteHeader(http.StatusOK)
}
