package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/jdebug14/kube-portal/internal/handlers"
	"github.com/jdebug14/kube-portal/internal/k8s"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	kubeclient, err := k8s.NewClient(logger)
	if err != nil {
		logger.Error("failed to create kubernetes client", "error", err)
		os.Exit(1)
	}

	port := os.Getenv("KUBEPORTAL_PORT")
	if port == "" {
		port = "8080"
	}

	handler := handlers.NewHandler(kubeclient, logger)
	router := handlers.NewRouter(handler)

	logger.Info("server starting", "port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
