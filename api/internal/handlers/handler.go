package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jdebug14/kube-portal/internal/k8s"
)

type errorResponse struct {
	Message string `json:"error"`
	Code    int    `json:"code"`
}

type Handler struct {
	client *k8s.Client
	logger *slog.Logger
}

func NewHandler(c *k8s.Client, l *slog.Logger) *Handler {
	return &Handler{client: c, logger: l}
}

func (h *Handler) writeError(w http.ResponseWriter, code int, message string, cause error) {
	h.logger.Error("request error", "code", code, "error", message, "cause", cause)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(errorResponse{Message: message, Code: code}); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
