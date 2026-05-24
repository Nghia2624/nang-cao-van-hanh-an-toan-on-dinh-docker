package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nghia/dockerai/backend/internal/pkg/logger"
	"github.com/nghia/dockerai/backend/internal/repository"
	"github.com/nghia/dockerai/backend/internal/service"
)

type HealthHandler struct {
	Docker    service.DockerService
	Metrics   service.MetricsService
	LogRepo   repository.LogRepository
	AlertRepo repository.AlertRepository
	Log       *logger.Logger
}

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
	Version   string            `json:"version"`
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    make(map[string]string),
		Version:   "1.0.0",
	}

	// Check Docker connection
	_, err := h.Docker.ListContainers(ctx)
	if err != nil {
		status.Status = "degraded"
		status.Checks["docker"] = "unhealthy: " + err.Error()
	} else {
		status.Checks["docker"] = "healthy"
	}

	// Check MongoDB
	_, err = h.LogRepo.Query(ctx, "", time.Now().Add(-time.Minute), time.Now(), "", 1)
	if err != nil {
		status.Status = "degraded"
		status.Checks["mongodb"] = "unhealthy: " + err.Error()
	} else {
		status.Checks["mongodb"] = "healthy"
	}

	// Check Prometheus
	_, err = h.Metrics.GetSystemOverview(ctx)
	if err != nil {
		status.Status = "degraded"
		status.Checks["prometheus"] = "unhealthy: " + err.Error()
	} else {
		status.Checks["prometheus"] = "healthy"
	}

	// Always return 200 for liveness; report component health in payload.
	// Docker access may be intentionally restricted in some deployments.
	code := http.StatusOK

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(status)
}
