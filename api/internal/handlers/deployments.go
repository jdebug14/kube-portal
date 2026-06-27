package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	deployments, err := h.client.ListDeployments(
		r.Context(),
		chi.URLParam(r, "ns"),
	)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to list deployments", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(deployments); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
