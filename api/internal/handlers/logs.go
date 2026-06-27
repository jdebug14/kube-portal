package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetPodLogs(w http.ResponseWriter, r *http.Request) {
	var tailLines int64
	var err error
	if tailLinesReq := r.URL.Query().Get("tailLines"); tailLinesReq != "" {
		tailLines, err = strconv.ParseInt(tailLinesReq, 10, 64)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "tailLines was not a valid integer", err)
			return
		}
		if tailLines > 1000 {
			h.logger.Warn("Pod log tail lines are restricted to max of 1000", "tailLinesRequested", tailLines)
			tailLines = 1000
		}
	} else {
		tailLines = 100
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
