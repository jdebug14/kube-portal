package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
		h.writeError(w, http.StatusInternalServerError, "failed to retrieve deployments", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(deployments); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) GetDeploymentDetails(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "ns")
	if err := validateNamespaceName(namespace); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid namespace: "+err.Error(), err)
		return
	}
	deploymentName := chi.URLParam(r, "name")
	if err := validateResourceName(deploymentName); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid deployment: "+err.Error(), err)
		return
	}

	deploymentDetails, err := h.client.GetDeploymentDetails(r.Context(), namespace, deploymentName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			h.writeError(w, http.StatusNotFound, "deployment not found", err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, "failed to retrieve deployment details", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(deploymentDetails); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
