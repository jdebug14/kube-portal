package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListNamespaceEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.client.ListNamespaceEvents(
		r.Context(),
		chi.URLParam(r, "ns"),
		r.URL.Query().Get("involvedObjectName"),
	)
	if err != nil {
		http.Error(w, "failed to fetch namespace events", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
