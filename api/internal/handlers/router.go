package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *Handler) *chi.Mux {
	router := chi.NewRouter()

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	router.Get("/api/v1/namespaces", h.ListNamespaces)
	router.Get("/api/v1/namespaces/{ns}/deployments", h.ListDeployments)
	router.Get("/api/v1/namespaces/{ns}/deployments/{name}", h.GetDeploymentDetails)
	router.Get("/api/v1/namespaces/{ns}/pods", h.ListPods)
	router.Get("/api/v1/namespaces/{ns}/pods/{name}", h.GetPodDetails)
	router.Get("/api/v1/namespaces/{ns}/events", h.ListEvents)
	router.Get("/api/v1/namespaces/{ns}/pods/{name}/logs", h.GetPodLogs)

	return router
}
