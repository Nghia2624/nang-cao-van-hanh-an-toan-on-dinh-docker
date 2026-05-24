# DockerAI - Intelligent Container Monitoring Dashboard

Production-ready Docker container monitoring system with AI-powered log analysis, built with Go backend and React TypeScript frontend.

## 🎯 Features

- **Real-time Container Monitoring**: Monitor all containers with sub-second latency
- **AI-Powered Log Analysis**: Automatic root cause analysis using Google Gemini AI
- **Intelligent Alerting**: Context-aware alerts with severity scoring
- **Metrics Visualization**: CPU, Memory, Network, Disk I/O metrics with time-series charts
- **Advanced Log Viewer**: Filter, search, highlight with real-time streaming
- **Health Scoring**: Automated container health assessment
- **Pattern Recognition**: Detect repetitive errors, spikes, cascading failures
- **Production Ready**: Health checks, graceful shutdown, resource limits, security headers

## 🏗️ Architecture

```
┌─────────────────┐
│  React Frontend │ (Port 3000)
└────────┬────────┘
         │ REST API / SSE
┌────────▼────────┐
│  Go Backend     │ (Port 8080)
└───┬──────┬──────┘
    │      │
┌───▼──┐ ┌─▼────────┐ ┌─────────┐
│Docker│ │Prometheus│ │ MongoDB │
└───┬──┘ └────┬─────┘ └─────────┘
    │         │
┌───▼─────────▼──┐
│   cAdvisor     │
└────────────────┘
```

## 📋 Prerequisites

- Docker & Docker Compose
- Docker daemon running (for container monitoring)
- Google Gemini API keys (3 keys recommended for rotation)

## 🚀 Quick Start

### 1. Clone Repository

```bash
git clone <repository-url>
cd DockerAI
```

### 2. Configure Environment Variables

Create `.env` file:

```bash
# AI Configuration (REQUIRED)
AI_API_KEYS=key1,key2,key3
# OR single key:
# AI_API_KEY=your-key-here

# Backend API Key (optional, for API authentication)
APP_API_KEY=your-secure-api-key

# Frontend API Configuration
VITE_API=http://localhost:8080
VITE_API_KEY=your-secure-api-key  # Must match APP_API_KEY if set

# Grafana (optional)
GRAFANA_USER=admin
GRAFANA_PASSWORD=admin

# Log Level
LOG_LEVEL=info
```

### 3. Start Services

```bash
docker compose up -d
```

### 4. Access Services

- **Frontend Dashboard**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin)
- **cAdvisor**: http://localhost:8081

## 📖 API Documentation

### Health Check

```bash
GET /healthz
GET /health
```

Returns detailed health status of all components.

### Containers

```bash
GET /api/v1/containers              # List all containers
GET /api/v1/containers/:id         # Get container details
```

### Metrics

```bash
GET /api/v1/metrics/system                    # System overview
GET /api/v1/metrics/container/:id/cpu         # CPU metrics (timeseries)
GET /api/v1/metrics/container/:id/memory     # Memory metrics (timeseries)
```

### Logs

```bash
GET /api/v1/logs?container=:id&level=ERROR&from=:ts&to=:ts&limit=100
GET /api/v1/logs/stream?container=:id        # SSE stream
```

### Alerts

```bash
GET /api/v1/alerts?status=NEW
POST /api/v1/alerts/:id/acknowledged
POST /api/v1/alerts/:id/resolved
```

### AI Insights

```bash
GET /api/v1/ai/analyses
POST /api/v1/ai/analyze
```

## 🔧 Configuration

### Backend Configuration

All backend configuration via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_PORT` | `8080` | Backend HTTP port |
| `APP_API_KEY` | - | API key for authentication |
| `MONGO_URI` | `mongodb://mongo:27017` | MongoDB connection string |
| `MONGO_DB` | `dockerai` | MongoDB database name |
| `PROMETHEUS_URL` | `http://prometheus:9090` | Prometheus URL |
| `DOCKER_HOST` | `unix:///var/run/docker.sock` | Docker daemon socket |
| `AI_ENDPOINT` | Gemini API endpoint | Gemini API endpoint |
| `AI_MODEL` | `gemini-1.5-flash` | Gemini model name |
| `AI_API_KEYS` | - | Comma-separated API keys (recommended) |
| `AI_API_KEY` | - | Single API key (fallback) |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |

### Frontend Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API` | `http://localhost:8080` | Backend API URL |
| `VITE_API_KEY` | - | API key (must match APP_API_KEY) |

## 🏭 Production Deployment

### 1. Security Checklist

- [ ] Set strong `APP_API_KEY` and `VITE_API_KEY`
- [ ] Configure MongoDB authentication
- [ ] Use HTTPS (reverse proxy with nginx/traefik)
- [ ] Set resource limits in docker-compose.yml
- [ ] Enable Grafana authentication
- [ ] Rotate AI API keys regularly
- [ ] Monitor resource usage

### 2. Resource Requirements

Minimum:
- CPU: 4 cores
- RAM: 8GB
- Disk: 100GB (for logs and metrics)

Recommended:
- CPU: 8 cores
- RAM: 16GB
- Disk: 500GB SSD

### 3. Scaling

- **Backend**: Scale horizontally (multiple instances behind load balancer)
- **MongoDB**: Use replica set for high availability
- **Prometheus**: Consider remote storage for long retention
- **Frontend**: CDN for static assets

### 4. Monitoring

The system monitors itself:
- Backend health: `/healthz`
- Container metrics via cAdvisor
- Application logs in MongoDB

### 5. Backup Strategy

```bash
# MongoDB backup
docker exec dockerai-mongo mongodump --out /backup

# Prometheus data (if using volumes)
docker exec dockerai-prometheus tar czf /backup/prometheus.tar.gz /prometheus
```

## 🛠️ Development

### Backend Development

```bash
cd backend
go mod download
go run cmd/server/main.go
```

### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

### Running Tests

```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
npm run type-check
```

## 📊 Metrics & Observability

### Prometheus Metrics

- Container CPU usage
- Container memory usage
- Network I/O
- Disk I/O
- Container restarts
- Log error rates

### Grafana Dashboards

Pre-configured dashboards available at:
- http://localhost:3001/dashboards

### Alert Rules

Prometheus alert rules configured in `prometheus/alert_rules.yml`:
- High CPU usage (>80% for 5min)
- High memory usage (>90% for 5min)
- Container restart loops
- Container down

## 🔒 Security

- API key authentication
- CORS configuration
- Rate limiting (5 req/s, burst 20)
- Input validation
- Security headers (XSS protection, frame options)
- Read-only Docker socket mount

## 🐛 Troubleshooting

### Backend won't start

1. Check Docker socket: `ls -la /var/run/docker.sock`
2. Verify MongoDB connection
3. Check AI API keys are set
4. Review logs: `docker compose logs backend`

### Frontend can't connect

1. Verify `VITE_API` matches backend URL
2. Check CORS settings
3. Verify API key matches backend

### No metrics in Prometheus

1. Check cAdvisor is running: `curl http://localhost:8081/healthz`
2. Verify Prometheus targets: http://localhost:9090/targets
3. Check scrape config in `prometheus/prometheus.yml`

### AI analysis not working

1. Verify API keys are valid
2. Check rate limits (100 req/min)
3. Review logs for AI errors
4. Ensure sufficient quota on Gemini API

## 📝 License

[Your License Here]

## 🤝 Contributing

[Contributing Guidelines]

## 📧 Support

[Support Contact]

---

**Built with ❤️ for DevOps teams**
# nang-cao-van-hanh-an-toan-on-dinh-docker
# nang-cao-van-hanh-an-toan-on-dinh-docker
