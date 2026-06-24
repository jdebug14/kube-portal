package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jdebug14/kube-portal/internal/handlers"
	"github.com/jdebug14/kube-portal/internal/k8s"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	kubeclient, err := k8s.NewClient()
	if err != nil {
		logger.Error("failed to create kubernetes client", "error", err)
		os.Exit(1)
	}

	port := os.Getenv("KUBEPORTAL_PORT")
	if port == "" {
		port = "8080"
	}

	router := chi.NewRouter()
	handler := handlers.NewHandler(kubeclient, logger)

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	router.Get("/api/v1/namespaces", handler.ListNamespaces)
	router.Get("/api/v1/namespaces/{ns}/deployments", handler.ListDeployments)
	router.Get("/api/v1/namespaces/{ns}/pods", handler.ListPods)
	router.Get("/api/v1/namespaces/{ns}/pods/{pn}", handler.GetPodDetail)
	router.Get("/api/v1/namespaces/{ns}/events", handler.ListNamespaceEvents)

	logger.Info("server starting", "port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
