package service

import (
	"context"
	"time"

	"github.com/nghia/dockerai/backend/internal/domain"
)

type MetricsService interface {
	GetContainerCPU(ctx context.Context, containerID string, from, to time.Time) ([]domain.MetricPoint, error)
	GetContainerMemory(ctx context.Context, containerID string, from, to time.Time) ([]domain.MetricPoint, error)
	GetContainerNetwork(ctx context.Context, containerID string, duration time.Duration) (*domain.NetworkMetrics, error)
	GetSystemOverview(ctx context.Context) (*domain.SystemMetrics, error)
	GetContainerHealth(ctx context.Context, containerID string) (*domain.HealthMetrics, error)
}
