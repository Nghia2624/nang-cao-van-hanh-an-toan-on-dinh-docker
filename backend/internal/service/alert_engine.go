package service

import (
	"context"

	"github.com/nghia/dockerai/backend/internal/domain"
)

type AlertEngine interface {
	EvaluateMetrics(ctx context.Context, metrics *domain.HealthMetrics, container domain.Container) (*domain.Alert, error)
	EvaluateLog(ctx context.Context, log *domain.LogEntry) (*domain.Alert, error)
	EvaluateAI(ctx context.Context, analysis *domain.AIAnalysis) (*domain.Alert, error)
}
