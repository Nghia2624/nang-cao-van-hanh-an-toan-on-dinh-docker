package service

import (
	"context"

	"github.com/nghia/dockerai/backend/internal/domain"
)

type LogProcessor interface {
	ProcessLog(ctx context.Context, log *domain.LogEntry) error
	ProcessBatch(ctx context.Context, logs []*domain.LogEntry) error
	ShouldAnalyzeWithAI(log *domain.LogEntry) bool
	QueueForAI(log *domain.LogEntry) error
	// Stop gracefully shuts down the log processor queue, flushing pending items.
	Stop()
}
