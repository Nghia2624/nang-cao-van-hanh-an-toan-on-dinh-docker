package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	App struct {
		Port   string
		APIKey string
	}
	Mongo struct {
		URI      string
		Database string
	}
	Prometheus struct {
		URL string
	}
	Docker struct {
		Host string
	}
	AI struct {
		Endpoint string
		APIKeys  []string
		Model    string
	}
	Log struct {
		Level string
	}
}

func Load() (*Config, error) {
	cfg := &Config{}
	cfg.App.Port = env("APP_PORT", "8080")
	cfg.App.APIKey = env("APP_API_KEY", "")
	cfg.Mongo.URI = env("MONGO_URI", "mongodb://mongo:27017")
	cfg.Mongo.Database = env("MONGO_DB", "dockerai")
	cfg.Prometheus.URL = env("PROMETHEUS_URL", "http://prometheus:9090")
	cfg.Docker.Host = env("DOCKER_HOST", "unix:///var/run/docker.sock")
	cfg.AI.Endpoint = env("AI_ENDPOINT", "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent")
	rawKeys := env("AI_API_KEYS", "")
	if rawKeys != "" {
		parts := strings.Split(rawKeys, ",")
		for _, p := range parts {
			if v := strings.TrimSpace(p); v != "" {
				cfg.AI.APIKeys = append(cfg.AI.APIKeys, v)
			}
		}
	}
	// Backward compatibility: single key env
	if len(cfg.AI.APIKeys) == 0 {
		if single := env("AI_API_KEY", ""); single != "" {
			cfg.AI.APIKeys = append(cfg.AI.APIKeys, single)
		}
	}
	cfg.AI.Model = env("AI_MODEL", "gemini-2.5-flash")
	cfg.Log.Level = env("LOG_LEVEL", "info")

	if len(cfg.AI.APIKeys) == 0 {
		return nil, fmt.Errorf("AI_API_KEYS or AI_API_KEY is required")
	}

	return cfg, nil
}

func env(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
