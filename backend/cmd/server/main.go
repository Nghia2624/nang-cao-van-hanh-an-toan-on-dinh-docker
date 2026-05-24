package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/nghia/dockerai/backend/internal/config"
	"github.com/nghia/dockerai/backend/internal/handler/httpapi"
	"github.com/nghia/dockerai/backend/internal/pkg/logger"
	"github.com/nghia/dockerai/backend/internal/repository/mongodb"
	promrepo "github.com/nghia/dockerai/backend/internal/repository/prometheus"
	"github.com/nghia/dockerai/backend/internal/service"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level)
	defer log.Sync() //nolint:errcheck

	mongoClient, err := mongodb.New(cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect mongo: %v\n", err)
		os.Exit(1)
	}
	if err := mongodb.EnsureIndexes(ctx, mongoClient.Database()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to ensure mongo indexes: %v\n", err)
		os.Exit(1)
	}

	promRepo, err := promrepo.New(cfg.Prometheus.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init prometheus client: %v\n", err)
		os.Exit(1)
	}

	dockerSvc, err := service.NewDockerService(cfg.Docker.Host, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init docker service: %v\n", err)
		os.Exit(1)
	}

	alertRepo := mongodb.NewAlertRepo(mongoClient)
	aiRepo := mongodb.NewAIRepo(mongoClient)
	aiChatRepo := mongodb.NewAIChatRepo(mongoClient)
	logRepo := mongodb.NewLogRepo(mongoClient)
	predictionFeedbackRepo := mongodb.NewPredictionFeedbackRepo(mongoClient)

	alertEngine := service.NewAlertEngine(alertRepo, log)
	aiAnalyzer := service.NewAIAnalyzer(cfg.AI.Endpoint, cfg.AI.APIKeys, cfg.AI.Model, aiRepo, log)
	metricsSvc := service.NewMetricsService(promRepo, dockerSvc)
	sseHub := httpapi.NewSSEHub()
	logProcessor := service.NewLogProcessor(logRepo, aiAnalyzer, alertEngine, log, sseHub)

	handler := &httpapi.Handler{
		Docker:                 dockerSvc,
		Metrics:                metricsSvc,
		Logs:                   logProcessor,
		AI:                     aiAnalyzer,
		Alerts:                 alertEngine,
		SSE:                    sseHub,
		Log:                    log,
		AlertRepo:              alertRepo,
		LogRepo:                logRepo,
		AIRepo:                 aiRepo,
		AIChatRepo:             aiChatRepo,
		PredictionFeedbackRepo: predictionFeedbackRepo,
	}

	r := chi.NewRouter()
	httpapi.RegisterRoutes(r, handler, log, cfg.App.APIKey)

	server := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second, // Must be > SSE/AI chat timeout (90s)
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Infow("starting http server", "port", cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalw("http server error", "error", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()

	log.Info("shutting down http server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Errorw("graceful shutdown failed", "error", err)
	}

	// Gracefully stop log processor queue (flush pending batch)
	logProcessor.Stop()

	log.Info("server stopped")
}
