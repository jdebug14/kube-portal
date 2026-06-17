package handlers

import (
	"log/slog"

	"github.com/jdebug14/kube-portal/internal/k8s"
)

type Handler struct {
	client *k8s.Client
	logger *slog.Logger
}

func NewHandler(c *k8s.Client, l *slog.Logger) *Handler {
	return &Handler{client: c, logger: l}
}
