package server

import (
	"net/http"

	"mini-imageflux-go/internal/metrics"
)

func HandleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(metrics.Render())); err != nil {
		return
	}
}
