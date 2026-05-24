package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"

	"github.com/nghia/dockerai/backend/internal/domain"
	dockerc "github.com/nghia/dockerai/backend/internal/pkg/docker"
	"github.com/nghia/dockerai/backend/internal/pkg/logger"
)

type dockerService struct {
	cli *dockerc.Client
	log *logger.Logger
}

func NewDockerService(host string, log *logger.Logger) (DockerService, error) {
	cli, err := dockerc.New(host)
	if err != nil {
		return nil, err
	}
	return &dockerService{cli: cli, log: log}, nil
}

func (s *dockerService) ListContainers(ctx context.Context) ([]domain.Container, error) {
	containers, err := s.cli.ListContainers(ctx)
	if err != nil {
		return nil, err
	}
	var res []domain.Container
	for _, c := range containers {
		status := c.Status
		name := strings.TrimPrefix(c.Names[0], "/")

		// Calculate uptime from container state, not created time
		uptimeSec := int64(0)
		lowerStatus := strings.ToLower(status)
		if strings.Contains(lowerStatus, "up") {
			// Container is running - estimate uptime from status string or Created time
			uptimeSec = int64(time.Since(time.Unix(c.Created, 0)).Seconds())
		}
		// For exited/stopped containers, uptime stays 0

		res = append(res, domain.Container{
			ID:           c.ID,
			Name:         name,
			Image:        c.Image,
			Status:       status,
			CreatedAt:    time.Unix(c.Created, 0),
			Labels:       c.Labels,
			UptimeSec:    uptimeSec,
			RestartCount: 0, // Not available from List API; use GetContainer for accurate count
		})
	}
	return res, nil
}

func (s *dockerService) GetContainer(ctx context.Context, id string) (*domain.Container, error) {
	info, err := s.cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}
	name := strings.TrimPrefix(info.Name, "/")
	created, _ := time.Parse(time.RFC3339Nano, info.Created)

	// Build a human-readable status consistent with Docker's list API
	status := info.State.Status // "running", "exited", etc.
	uptimeSec := int64(0)
	if info.State.Running {
		started, _ := time.Parse(time.RFC3339Nano, info.State.StartedAt)
		uptime := time.Since(started)
		uptimeSec = int64(uptime.Seconds())
		status = "Up " + formatDuration(uptime)
		if info.State.Health != nil {
			status += " (" + strings.ToLower(info.State.Health.Status) + ")"
		}
	} else if info.State.ExitCode != 0 {
		status = fmt.Sprintf("Exited (%d)", info.State.ExitCode)
	}

	return &domain.Container{
		ID:           info.ID,
		Name:         name,
		Image:        info.Config.Image,
		Status:       status,
		CreatedAt:    created,
		Labels:       info.Config.Labels,
		UptimeSec:    uptimeSec,
		RestartCount: info.RestartCount,
	}, nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		if m > 0 {
			return fmt.Sprintf("%d hours %d minutes", h, m)
		}
		return fmt.Sprintf("%d hours", h)
	}
	days := int(d.Hours()) / 24
	return fmt.Sprintf("%d days", days)
}

func (s *dockerService) GetContainerStats(ctx context.Context, id string) (*domain.ContainerStats, error) {
	stats, err := s.cli.ContainerStats(ctx, id)
	if err != nil {
		return nil, err
	}
	cpuPercent := calculateCPUPercentUnix(stats)
	return &domain.ContainerStats{
		CPUPercent:       cpuPercent,
		MemoryUsageBytes: stats.MemoryStats.Usage,
		MemoryLimitBytes: stats.MemoryStats.Limit,
		NetworkRxBytes:   networkRx(stats.Networks),
		NetworkTxBytes:   networkTx(stats.Networks),
		DiskReadBytes:    blkRead(stats.BlkioStats.IoServiceBytesRecursive),
		DiskWriteBytes:   blkWrite(stats.BlkioStats.IoServiceBytesRecursive),
		HealthStatus:     healthStatus(stats),
	}, nil
}

func (s *dockerService) StreamLogs(ctx context.Context, id string, follow bool, tail int) (<-chan domain.LogEntry, error) {
	tailStr := "100"
	if tail > 0 {
		tailStr = strconv.Itoa(tail)
	}
	logStream, err := s.cli.Logs(ctx, id, follow, tailStr)
	if err != nil {
		return nil, err
	}

	out := make(chan domain.LogEntry, 1000)
	go func() {
		defer close(out)
		defer logStream.Close() // Properly release the stream
		if logStream.Reader == nil {
			s.log.Errorw("log stream reader is nil", "container", id)
			return
		}
		reader := logStream.Reader
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					s.log.Warnw("log stream error", "err", err)
				}
				return
			}
			entry := parseDockerLogLine(line, id)
			select {
			case out <- entry:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func (s *dockerService) MonitorEvents(ctx context.Context) (<-chan domain.ContainerEvent, error) {
	evChan, errChan := s.cli.Events(ctx)
	out := make(chan domain.ContainerEvent, 100)
	go func() {
		defer close(out)
		for {
			select {
			case ev := <-evChan:
				out <- mapEvent(ev)
			case err := <-errChan:
				if err != nil {
					s.log.Warnw("docker events error", "err", err)
				}
				return
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func mapEvent(ev events.Message) domain.ContainerEvent {
	return domain.ContainerEvent{
		Type: ev.Action,
		Container: domain.Container{
			ID: ev.ID,
		},
		Timestamp: time.Unix(ev.Time, 0),
	}
}

func networkRx(stats map[string]types.NetworkStats) uint64 {
	var total uint64
	for _, s := range stats {
		total += s.RxBytes
	}
	return total
}

func networkTx(stats map[string]types.NetworkStats) uint64 {
	var total uint64
	for _, s := range stats {
		total += s.TxBytes
	}
	return total
}

func blkRead(entries []types.BlkioStatEntry) uint64 {
	var total uint64
	for _, e := range entries {
		if strings.ToLower(e.Op) == "read" {
			total += e.Value
		}
	}
	return total
}

func blkWrite(entries []types.BlkioStatEntry) uint64 {
	var total uint64
	for _, e := range entries {
		if strings.ToLower(e.Op) == "write" {
			total += e.Value
		}
	}
	return total
}

func healthStatus(stats types.StatsJSON) string {
	// StatsJSON doesn't carry health info directly; inspected state needed.
	// Return "running" as default when stats are available (container is alive).
	return "running"
}

// calculateCPUPercentUnix is adapted from moby/moby helper.
func calculateCPUPercentUnix(v types.StatsJSON) float64 {
	var (
		cpuPercent  = 0.0
		cpuDelta    = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(v.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(v.PreCPUStats.SystemUsage)
	)
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

func parseDockerLogLine(line string, containerID string) domain.LogEntry {
	// Docker logs have a header: [8 bytes: stream type (1 byte) + padding (3 bytes) + size (4 bytes)]
	// Remove Docker log header bytes if present
	cleaned := line
	if len(line) >= 8 {
		// Check if first byte is 1 (stdout) or 2 (stderr) - Docker log header indicator
		if line[0] == 1 || line[0] == 2 {
			// Skip the 8-byte header
			if len(line) > 8 {
				cleaned = line[8:]
			} else {
				cleaned = ""
			}
		}
	}
	
	// Remove null bytes and other control characters except newlines and tabs
	cleaned = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || r >= 32 {
			return r
		}
		return -1 // Remove character
	}, cleaned)
	
	trimmed := strings.TrimSpace(cleaned)
	
	// Parse timestamp if present (Docker format: 2026-01-19T16:37:31.348494479Z)
	ts := time.Now().UTC()
	if len(trimmed) > 20 {
		// Find the first space after timestamp to separate timestamp from message
		spaceIdx := strings.IndexByte(trimmed, ' ')
		if spaceIdx > 10 && spaceIdx < 40 {
			timePart := trimmed[:spaceIdx]
			if parsed, err := time.Parse(time.RFC3339Nano, timePart); err == nil {
				ts = parsed
				trimmed = strings.TrimSpace(trimmed[spaceIdx+1:])
			}
		}
	}
	
	level := detectLevel(trimmed)
	
	return domain.LogEntry{
		ContainerID: containerID,
		Level:       level,
		Message:     trimmed,
		Timestamp:   ts,
	}
}

func detectLevel(msg string) domain.LogLevel {
	// 1. Try to parse as structured JSON and extract "level" field
	if len(msg) > 2 && msg[0] == '{' {
		var obj struct {
			Level string `json:"level"`
		}
		if json.Unmarshal([]byte(msg), &obj) == nil && obj.Level != "" {
			switch strings.ToLower(obj.Level) {
			case "fatal", "panic":
				return domain.LogLevelFatal
			case "error", "err":
				return domain.LogLevelError
			case "warn", "warning":
				return domain.LogLevelWarn
			case "debug", "trace":
				return domain.LogLevelDebug
			default:
				return domain.LogLevelInfo
			}
		}
	}

	// 2. Fallback: keyword detection on plain text logs
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "panic"), strings.Contains(lower, "fatal"):
		return domain.LogLevelFatal
	case strings.Contains(lower, "[error]"), strings.Contains(lower, "\"level\":\"error\""),
		strings.Contains(lower, "exception"), strings.Contains(lower, " error:"):
		return domain.LogLevelError
	case strings.Contains(lower, "[warn"), strings.Contains(lower, "\"level\":\"warn\""),
		strings.Contains(lower, "warning:"):
		return domain.LogLevelWarn
	default:
		return domain.LogLevelInfo
	}
}
