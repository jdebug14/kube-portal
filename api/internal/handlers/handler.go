package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"

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
	h.logger.Error("request error", "code", code, "cause", cause)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(errorResponse{Message: message, Code: code}); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

var namespaceRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

func validateNamespaceName(name string) error {
	if !namespaceRegex.MatchString(name) {
		return fmt.Errorf("kubernetes namespace names must be 1-63 characters, contain only lowercase alphanumeric and hyphens, and must start and end with an alphanumeric character")
	}
	return nil
}

var resourceRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9.\-]{0,251}[a-z0-9])?$`)

func validateResourceName(name string) error {
	if !resourceRegex.MatchString(name) {
		return fmt.Errorf("kubernetes resource names must be 1-232 characters, contain only lowercase alphanumeric, hyphens and dots, and must start and end with an alphanumeric character")
	}
	return nil
}
