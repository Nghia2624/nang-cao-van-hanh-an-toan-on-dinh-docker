package service

import (
	"context"

	"github.com/nghia/dockerai/backend/internal/domain"
)

// DockerService defines Docker integration points.
type DockerService interface {
	ListContainers(ctx context.Context) ([]domain.Container, error)
	GetContainer(ctx context.Context, id string) (*domain.Container, error)
	GetContainerStats(ctx context.Context, id string) (*domain.ContainerStats, error)
	StreamLogs(ctx context.Context, id string, follow bool, tail int) (<-chan domain.LogEntry, error)
	MonitorEvents(ctx context.Context) (<-chan domain.ContainerEvent, error)
}
