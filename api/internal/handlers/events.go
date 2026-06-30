package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.client.ListEvents(
		r.Context(),
		chi.URLParam(r, "ns"),
		r.URL.Query().Get("involvedObjectName"),
	)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to fetch events", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
