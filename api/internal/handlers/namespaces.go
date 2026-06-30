package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) ListNamespaces(w http.ResponseWriter, r *http.Request) {
	namespaces, err := h.client.ListNamespaces(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to fetch namespaces", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(namespaces); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
