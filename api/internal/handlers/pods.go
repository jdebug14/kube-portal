package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListPods(w http.ResponseWriter, r *http.Request) {
	pods, err := h.client.ListPods(r.Context(), chi.URLParam(r, "ns"))
	if err != nil {
		http.Error(w, "failed to list pods", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pods); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) GetPodDetail(w http.ResponseWriter, r *http.Request) {
	podDetails, err := h.client.GetPodDetail(r.Context(), chi.URLParam(r, "ns"), chi.URLParam(r, "pn"))
	if err != nil {
		http.Error(w, "failed to get pod", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(podDetails); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
