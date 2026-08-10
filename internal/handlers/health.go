package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/charbelbahry/aws-cloud-platform/internal/database"
)

var startTime = time.Now()

// HealthHandler handles liveness, readiness, and metrics.
type HealthHandler struct {
	db *database.DB
}

// NewHealthHandler returns a new HealthHandler.
func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// RegisterRoutes registers health and metrics endpoints on ServeMux.
func (h *HealthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.Liveness)
	mux.HandleFunc("GET /ready", h.Readiness)
	mux.HandleFunc("GET /metrics", h.Metrics)
}

// Liveness handles GET /health (always returns 200 OK).
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readiness handles GET /ready (pings DB, returns 200 OK or 503 Service Unavailable).
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.db.Ping(ctx); err != nil {
		respondError(w, http.StatusServiceUnavailable, "database connection unavailable")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// Metrics handles GET /metrics (returns basic application metrics).
func (h *HealthHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	stats := map[string]any{
		"uptime_seconds":  int(time.Since(startTime).Seconds()),
		"goroutine_count": runtime.NumGoroutine(),
		"go_version":      runtime.Version(),
	}
	respondJSON(w, http.StatusOK, stats)
}
