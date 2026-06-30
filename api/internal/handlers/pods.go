package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (h *Handler) ListPods(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "ns")
	if err := validateNamespaceName(namespace); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid namespace: "+err.Error(), err)
		return
	}

	pods, err := h.client.ListPods(
		r.Context(),
		namespace,
	)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to fetch pods", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pods); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) GetPodDetail(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "ns")
	if err := validateNamespaceName(namespace); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid namespace: "+err.Error(), err)
		return
	}
	podName := chi.URLParam(r, "pn")
	if err := validateResourceName(podName); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid pod: "+err.Error(), err)
		return
	}
	podDetails, err := h.client.GetPodDetails(
		r.Context(),
		namespace,
		podName,
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			h.writeError(w, http.StatusNotFound, "pod not found", err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, "failed to retrieve pod details", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(podDetails); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
