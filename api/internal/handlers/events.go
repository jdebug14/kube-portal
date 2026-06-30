package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "ns")
	if err := validateNamespaceName(namespace); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid namespace: "+err.Error(), err)
		return
	}
	involvedObject := r.URL.Query().Get("involvedObjectName")
	if involvedObject != "" {
		if err := validateResourceName(involvedObject); err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid object filter: "+err.Error(), err)
			return
		}
	}

	events, err := h.client.ListEvents(
		r.Context(),
		namespace,
		involvedObject,
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
