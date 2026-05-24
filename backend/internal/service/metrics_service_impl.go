package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nghia/dockerai/backend/internal/domain"
	"github.com/nghia/dockerai/backend/internal/repository"
)

type metricsService struct {
	repo   repository.MetricsRepository
	docker DockerService
}

func NewMetricsService(repo repository.MetricsRepository, docker DockerService) MetricsService {
	return &metricsService{repo: repo, docker: docker}
}

// containerIDFilter builds a PromQL regex matching both full and short container IDs.
func containerIDFilter(containerID string) string {
	shortID := containerID
	if len(containerID) > 12 {
		shortID = containerID[:12]
	}
	return fmt.Sprintf(`/docker/%s|/docker/%s|%s|%s`, containerID, shortID, containerID, shortID)
}

func (s *metricsService) GetContainerCPU(ctx context.Context, containerID string, from, to time.Time) ([]domain.MetricPoint, error) {
	idFilter := containerIDFilter(containerID)
	promql := fmt.Sprintf(`rate(container_cpu_usage_seconds_total{id=~"%s"}[5m])*100`, idFilter)
	points, err := s.repo.QueryRange(ctx, promql, from, to, 15*time.Second)
	if err == nil && len(points) > 0 {
		return points, nil
	}

	// Fallback: use Docker Stats API directly
	if s.docker != nil {
		stats, statsErr := s.docker.GetContainerStats(ctx, containerID)
		if statsErr == nil && stats != nil {
			now := time.Now()
			return []domain.MetricPoint{
				{Timestamp: now, Value: stats.CPUPercent},
			}, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return points, nil
}

func (s *metricsService) GetContainerMemory(ctx context.Context, containerID string, from, to time.Time) ([]domain.MetricPoint, error) {
	idFilter := containerIDFilter(containerID)
	promql := fmt.Sprintf(`(container_memory_usage_bytes{id=~"%s"} / container_spec_memory_limit_bytes{id=~"%s"})*100`, idFilter, idFilter)
	points, err := s.repo.QueryRange(ctx, promql, from, to, 15*time.Second)
	if err == nil && len(points) > 0 {
		return points, nil
	}

	// Fallback: use Docker Stats API directly
	if s.docker != nil {
		stats, statsErr := s.docker.GetContainerStats(ctx, containerID)
		if statsErr == nil && stats != nil && stats.MemoryLimitBytes > 0 {
			now := time.Now()
			memPercent := float64(stats.MemoryUsageBytes) / float64(stats.MemoryLimitBytes) * 100
			return []domain.MetricPoint{
				{Timestamp: now, Value: memPercent},
			}, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return points, nil
}

func (s *metricsService) GetContainerNetwork(ctx context.Context, containerID string, duration time.Duration) (*domain.NetworkMetrics, error) {
	from := time.Now().Add(-duration)
	to := time.Now()
	idFilter := containerIDFilter(containerID)
	rx, err := s.repo.QueryRange(ctx, fmt.Sprintf(`rate(container_network_receive_bytes_total{id=~"%s"}[5m])`, idFilter), from, to, 30*time.Second)
	if err != nil {
		return nil, err
	}
	tx, err := s.repo.QueryRange(ctx, fmt.Sprintf(`rate(container_network_transmit_bytes_total{id=~"%s"}[5m])`, idFilter), from, to, 30*time.Second)
	if err != nil {
		return nil, err
	}
	rxErr, _ := s.repo.QueryRange(ctx, fmt.Sprintf(`rate(container_network_receive_errors_total{id=~"%s"}[5m])`, idFilter), from, to, 30*time.Second)
	txErr, _ := s.repo.QueryRange(ctx, fmt.Sprintf(`rate(container_network_transmit_errors_total{id=~"%s"}[5m])`, idFilter), from, to, 30*time.Second)
	return &domain.NetworkMetrics{
		RxBytesRate: lastValue(rx),
		TxBytesRate: lastValue(tx),
		RxErrors:    lastValue(rxErr),
		TxErrors:    lastValue(txErr),
	}, nil
}

func (s *metricsService) GetSystemOverview(ctx context.Context) (*domain.SystemMetrics, error) {
	cpu, _ := s.repo.QueryInstant(ctx, `avg(rate(container_cpu_usage_seconds_total[5m]))*100`, time.Now())
	mem, _ := s.repo.QueryInstant(ctx, `(sum(container_memory_usage_bytes)/sum(container_spec_memory_limit_bytes))*100`, time.Now())

	runningCount := 0
	totalCount := 0
	if s.docker != nil {
		containers, err := s.docker.ListContainers(ctx)
		if err == nil {
			totalCount = len(containers)
			for _, c := range containers {
				if strings.Contains(strings.ToLower(c.Status), "up") {
					runningCount++
				}
			}
		}
	}

	return &domain.SystemMetrics{
		CPUPercent:       cpu,
		MemoryPercent:    mem,
		ContainerRunning: runningCount,
		ContainerTotal:   totalCount,
	}, nil
}

func (s *metricsService) GetContainerHealth(ctx context.Context, containerID string) (*domain.HealthMetrics, error) {
	idFilter := containerIDFilter(containerID)
	now := time.Now()
	from := now.Add(-15 * time.Minute)
	cpu := lastValueOrZero(s.repo.QueryRange(ctx, fmt.Sprintf(`rate(container_cpu_usage_seconds_total{id=~"%s"}[5m])*100`, idFilter), from, now, 15*time.Second))
	mem := lastValueOrZero(s.repo.QueryRange(ctx, fmt.Sprintf(`(container_memory_usage_bytes{id=~"%s"} / container_spec_memory_limit_bytes{id=~"%s"})*100`, idFilter, idFilter), from, now, 15*time.Second))

	score := 100
	var factors []string
	if cpu > 80 {
		score -= 20
		factors = append(factors, "CPU > 80%")
	}
	if mem > 90 {
		score -= 25
		factors = append(factors, "Memory > 90%")
	}
	status := "HEALTHY"
	switch {
	case score < 50:
		status = "CRITICAL"
	case score < 70:
		status = "UNHEALTHY"
	case score < 90:
		status = "DEGRADED"
	}
	recs := []string{}
	if cpu > 80 {
		recs = append(recs, "Kiểm tra hotspots CPU, scale hoặc giới hạn tài nguyên.")
	}
	if mem > 90 {
		recs = append(recs, "Kiểm tra memory leak, tăng limit hoặc tối ưu bộ nhớ.")
	}

	return &domain.HealthMetrics{
		Score:           score,
		Status:          status,
		Summary:         fmt.Sprintf("CPU %.1f%%, MEM %.1f%%", cpu, mem),
		Factors:         factors,
		Recommendations: recs,
	}, nil
}

func lastValue(points []domain.MetricPoint) float64 {
	if len(points) == 0 {
		return 0
	}
	return points[len(points)-1].Value
}

func lastValueOrZero(points []domain.MetricPoint, err error) float64 {
	if err != nil {
		return 0
	}
	return lastValue(points)
}

