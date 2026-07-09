package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (h *Handler) GetPodLogs(w http.ResponseWriter, r *http.Request) {
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
	container := r.URL.Query().Get("container")
	if err := validateResourceName(container); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid container: "+err.Error(), err)
		return
	}
	tailLines, err := parseTailLines(r.URL.Query().Get("tailLines"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "tailLines must be a positive integer", err)
		return
	}

	podLogs, err := h.client.GetPodLogs(
		r.Context(),
		namespace,
		podName,
		container,
		tailLines,
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			h.writeError(w, http.StatusNotFound, "pod not found", err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, "failed to retrieve pod logs", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, podLogs)
}

func parseTailLines(raw string) (int64, error) {
	if raw == "" {
		return 100, nil
	}
	tailLines, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	if tailLines <= 0 {
		return 0, fmt.Errorf("tailLines must be a positive integer")
	}
	if tailLines > 1000 {
		tailLines = 1000
	}
	return tailLines, nil
}
