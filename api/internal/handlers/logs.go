package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetPodLogs(w http.ResponseWriter, r *http.Request) {
	tailLines, err := parseTailLines(r.URL.Query().Get("tailLines"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "tailLines must be a positive integer", err)
		return
	}
	podLogs, err := h.client.GetPodLogs(
		r.Context(),
		chi.URLParam(r, "ns"),
		chi.URLParam(r, "pn"),
		r.URL.Query().Get("container"),
		tailLines,
	)
	if err != nil {
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
	if tailLines < 100 {
		return 0, fmt.Errorf("tailLines cannot be negative")
	}
	if tailLines > 1000 {
		tailLines = 1000
	}
	return tailLines, nil
}
