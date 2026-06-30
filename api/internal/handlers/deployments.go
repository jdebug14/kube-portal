package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "ns")
	if err := validateNamespaceName(namespace); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid namespace: "+err.Error(), err)
		return
	}

	deployments, err := h.client.ListDeployments(
		r.Context(),
		namespace,
	)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to fetch deployments", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(deployments); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
