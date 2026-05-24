package service

import (
	"context"

	"github.com/nghia/dockerai/backend/internal/domain"
)

type AIAnalyzer interface {
	AnalyzeLogs(ctx context.Context, logs []*domain.LogEntry) (*domain.AIAnalysis, error)
	AnalyzeError(ctx context.Context, log *domain.LogEntry, ctxSummary *domain.SystemMetrics) (*domain.AIAnalysis, error)
	Chat(ctx context.Context, message string, logs []*domain.LogEntry) (answer string, evidence []*domain.LogEntry, insight *domain.AIAnalysis, err error)
	ProcessQueue(ctx context.Context) error
}
