package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) ListNamespaces(w http.ResponseWriter, r *http.Request) {
	namespaces, err := h.client.ListNamespaces(r.Context())
	if err != nil {
		http.Error(w, "failed to list namespaces", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(namespaces); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
