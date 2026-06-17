package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jdebug14/kube-portal/internal/k8s"
)

type Handler struct {
	client *k8s.Client
	logger *slog.Logger
}

func NewHandler(c *k8s.Client, l *slog.Logger) *Handler {
	return &Handler{client: c, logger: l}
}

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
