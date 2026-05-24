package domain

import "time"

type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type NetworkMetrics struct {
	RxBytesRate float64 `json:"rx_bytes_rate"`
	TxBytesRate float64 `json:"tx_bytes_rate"`
	RxErrors    float64 `json:"rx_errors"`
	TxErrors    float64 `json:"tx_errors"`
}

type SystemMetrics struct {
	CPUPercent       float64 `json:"cpu_percent"`
	MemoryPercent    float64 `json:"memory_percent"`
	ContainerRunning int     `json:"container_running"`
	ContainerTotal   int     `json:"container_total"`
}

type HealthMetrics struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Summary         string   `json:"summary"`
	Factors         []string `json:"factors"`
	Recommendations []string `json:"recommendations"`
}
