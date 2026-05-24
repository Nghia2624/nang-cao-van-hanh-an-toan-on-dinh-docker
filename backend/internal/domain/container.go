package domain

import "time"

type Container struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	Labels      map[string]string `json:"labels,omitempty"`
	UptimeSec   int64             `json:"uptime_sec"`
	RestartCount int              `json:"restart_count"`
}

type ContainerStats struct {
	CPUPercent        float64 `json:"cpu_percent"`
	MemoryUsageBytes  uint64  `json:"memory_usage_bytes"`
	MemoryLimitBytes  uint64  `json:"memory_limit_bytes"`
	NetworkRxBytes    uint64  `json:"network_rx_bytes"`
	NetworkTxBytes    uint64  `json:"network_tx_bytes"`
	DiskReadBytes     uint64  `json:"disk_read_bytes"`
	DiskWriteBytes    uint64  `json:"disk_write_bytes"`
	HealthStatus      string  `json:"health_status"`
}

type ContainerEvent struct {
	Type      string    `json:"type"`
	Container Container `json:"container"`
	Timestamp time.Time `json:"timestamp"`
}
