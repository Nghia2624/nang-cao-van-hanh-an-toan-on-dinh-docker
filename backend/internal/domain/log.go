package domain

import "time"

type LogLevel string

const (
	LogLevelFatal LogLevel = "FATAL"
	LogLevelError LogLevel = "ERROR"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelDebug LogLevel = "DEBUG"
)

type LogEntry struct {
	ID           string            `json:"id,omitempty"`
	ContainerID  string            `json:"container_id"`
	Container    string            `json:"container"`
	Image        string            `json:"image"`
	Level        LogLevel          `json:"level"`
	Message      string            `json:"message"`
	Timestamp    time.Time         `json:"timestamp"`
	Stream       string            `json:"stream,omitempty"`
	Fingerprint  string            `json:"fingerprint,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	RequestID    string            `json:"request_id,omitempty"`
	DeploymentID string            `json:"deployment_id,omitempty"`
}
