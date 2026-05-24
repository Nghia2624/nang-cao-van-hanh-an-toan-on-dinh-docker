package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/nghia/dockerai/backend/internal/middleware"
	"github.com/nghia/dockerai/backend/internal/pkg/logger"
)

// RegisterRoutes wires HTTP routes with handlers and middleware.
func RegisterRoutes(r *chi.Mux, h *Handler, log *logger.Logger, apiKey string) {
	r.Use(middleware.CORS)
	r.Use(middleware.RequestLogger(log))

	// Health checks - no auth required
	healthHandler := &HealthHandler{
		Docker:    h.Docker,
		Metrics:   h.Metrics,
		LogRepo:   h.LogRepo,
		AlertRepo: h.AlertRepo,
		Log:       log,
	}
	r.Get("/healthz", healthHandler.HealthCheck)
	r.Get("/health", healthHandler.HealthCheck)

	// API routes - require auth and rate limiting
	apiMiddleware := []func(http.Handler) http.Handler{
		middleware.RateLimit(20, 50), // 20 req/s, burst 50
	}
	if apiKey != "" {
		apiMiddleware = append(apiMiddleware, middleware.APIKeyAuth(apiKey))
	}

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Use(apiMiddleware...)

		v1.Get("/containers", h.listContainers)
		v1.Get("/containers/{id}", h.getContainer)
		v1.Get("/metrics/system", h.systemMetrics)
		v1.Get("/metrics/container/{id}/cpu", h.getContainerCPU)
		v1.Get("/metrics/container/{id}/memory", h.getContainerMemory)
		v1.Get("/logs", h.getLogs)
		v1.Get("/logs/stream", h.streamLogsSSE)
		v1.Get("/alerts", h.listAlerts)
		v1.Post("/alerts/{id}/{action}", h.updateAlertStatus)
		v1.Get("/ai/analyses", h.listAIAnalyses)
		v1.Post("/ai/analyze", h.aiAnalyze)

		v1.Post("/ai/chat/sessions", h.chatCreateSession)
		v1.Get("/ai/chat/sessions", h.chatListSessions)
		v1.Get("/ai/chat/sessions/{id}", h.chatGetSession)
		v1.Post("/ai/chat/sessions/{id}/messages", h.chatSendMessage)

		v1.Get("/containers/{id}/prediction", h.getContainerPrediction)
		v1.Post("/containers/{id}/prediction/feedback", h.submitPredictionFeedback)
	})

	// Events endpoint
	r.Group(func(events chi.Router) {
		events.Use(middleware.RateLimit(10, 30))
		if apiKey != "" {
			events.Use(middleware.APIKeyAuth(apiKey))
		}
		events.Handle("/events", h.SSE)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Warnw("not found", "path", r.URL.Path)
		http.NotFound(w, r)
	})
}
