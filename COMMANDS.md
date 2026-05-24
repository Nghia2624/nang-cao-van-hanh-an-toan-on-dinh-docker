# 🐳 DockerAI - Hướng Dẫn Chạy & Quản Lý Dự Án

## 📋 Tổng Quan Dự Án

**DockerAI** là hệ thống giám sát Docker container tích hợp AI, bao gồm:

| Service | Mô tả | Port | URL |
|---------|--------|------|-----|
| **Frontend** | React Dashboard (Vite + TypeScript + Nginx) | `3000` | http://localhost:3000 |
| **Backend** | Go API Server (Chi router + MongoDB) | `8080` | http://localhost:8080 |
| **MongoDB** | Database chính | `27018` | `mongodb://localhost:27018` |
| **Prometheus** | Metrics collector | `9090` | http://localhost:9090 |
| **Grafana** | Dashboard monitoring | `3001` | http://localhost:3001 |
| **cAdvisor** | Container metrics exporter | `8081` | http://localhost:8081 |

---

## 🚀 Khởi Chạy Dự Án

### Yêu Cầu Trước Khi Chạy
- Docker Engine ≥ 24.x
- Docker Compose V2 (plugin)
- File `.env` đã được cấu hình (copy từ `.env.example`)

### Chạy Lần Đầu (Build + Start)
```bash
# Build tất cả images và khởi chạy containers
docker compose up -d --build
```

### Chạy Bình Thường (Không Rebuild)
```bash
# Khởi chạy tất cả containers
docker compose up -d
```

### Chạy Với Log Hiển Thị (Foreground)
```bash
# Chạy và xem logs trực tiếp (Ctrl+C để dừng)
docker compose up --build
```

---

## ⏹️ Dừng Dự Án

### Dừng Tất Cả Containers
```bash
# Dừng containers (giữ data volumes)
docker compose down
```

### Dừng Và Xóa Toàn Bộ Data
```bash
# ⚠️ CẢNH BÁO: Xóa tất cả data (MongoDB, Prometheus, Grafana)
docker compose down -v
```

### Dừng Và Xóa Kèm Images
```bash
# Xóa containers, volumes, và images đã build
docker compose down -v --rmi local
```

---

## 🔧 Lệnh Docker Compose Thường Dùng

### Quản Lý Containers

```bash
# Xem trạng thái containers
docker compose ps

# Xem logs tất cả services
docker compose logs

# Xem logs 1 service cụ thể (ví dụ: backend)
docker compose logs backend

# Xem logs realtime (follow)
docker compose logs -f

# Xem logs backend realtime (50 dòng cuối)
docker compose logs -f --tail=50 backend

# Restart 1 service
docker compose restart backend

# Rebuild và restart 1 service cụ thể
docker compose up -d --build backend

# Rebuild và restart frontend
docker compose up -d --build frontend

# Dừng 1 service
docker compose stop backend

# Khởi động 1 service đã dừng
docker compose start backend
```

### Xem Tài Nguyên

```bash
# Xem containers đang chạy với ports
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Xem tài nguyên CPU/Memory realtime
docker stats --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"

# Xem dung lượng images
docker compose images

# Xem dung lượng volumes
docker volume ls --filter "label=com.dockerai.volume"
```

---

## 🏭 Production Mode

### Chạy Production
```bash
# Sử dụng file docker-compose.prod.yml
docker compose -f docker-compose.prod.yml up -d
```

### Dừng Production
```bash
docker compose -f docker-compose.prod.yml down
```

---

## 🔍 Kiểm Tra Health

### Kiểm Tra Backend Health
```bash
# Health check đầy đủ (Docker, MongoDB, Prometheus)
curl -s http://localhost:8080/healthz | python3 -m json.tool
```

### Kiểm Tra API Hoạt Động
```bash
# Liệt kê containers (cần API key)
curl -s -H "X-API-Key: 18mryrm2yW6v8vraI17+dzv3JU1LozkVAZmjjc90x1s=" \
  http://localhost:8080/api/v1/containers | python3 -m json.tool

# Xem system metrics
curl -s -H "X-API-Key: 18mryrm2yW6v8vraI17+dzv3JU1LozkVAZmjjc90x1s=" \
  http://localhost:8080/api/v1/metrics/system | python3 -m json.tool
```

### Kiểm Tra Từng Service
```bash
# Frontend
curl -s -o /dev/null -w "Frontend: HTTP %{http_code}\n" http://localhost:3000

# Backend
curl -s -o /dev/null -w "Backend: HTTP %{http_code}\n" http://localhost:8080/healthz

# Prometheus
curl -s -o /dev/null -w "Prometheus: HTTP %{http_code}\n" http://localhost:9090/-/healthy

# Grafana
curl -s -o /dev/null -w "Grafana: HTTP %{http_code}\n" http://localhost:3001/api/health

# cAdvisor
curl -s -o /dev/null -w "cAdvisor: HTTP %{http_code}\n" http://localhost:8081/healthz
```

---

## 🐛 Troubleshooting

### Xem Logs Khi Có Lỗi
```bash
# Xem logs tất cả services (50 dòng cuối)
docker compose logs --tail=50

# Xem logs backend chi tiết
docker compose logs --tail=100 backend

# Xem logs mongo khi khởi động
docker compose logs mongo
```

### Kiểm Tra Docker Socket
```bash
# Kiểm tra quyền Docker socket
ls -la /var/run/docker.sock

# Kiểm tra GID docker group (cần match với group_add trong docker-compose.yml)
stat -c '%g' /var/run/docker.sock
```

### Restart Hoàn Toàn
```bash
# Dừng → Xóa → Build lại → Chạy
docker compose down
docker compose up -d --build
```

### Xóa Cache Build
```bash
# Xóa Docker build cache
docker builder prune -f

# Build lại không dùng cache
docker compose build --no-cache
docker compose up -d
```

### Lỗi Port Đã Bị Chiếm
```bash
# Kiểm tra port nào đang bị dùng
ss -tlnp | grep -E ':(3000|3001|8080|8081|9090|27018) '

# Hoặc
lsof -i :8080
```

---

## 📊 Truy Cập Dashboard

| Dashboard | URL | Tài khoản |
|-----------|-----|-----------|
| **DockerAI Frontend** | http://localhost:3000 | Không cần đăng nhập |
| **Grafana** | http://localhost:3001 | `admin` / `admin123456` |
| **Prometheus** | http://localhost:9090 | Không cần đăng nhập |
| **cAdvisor** | http://localhost:8081 | Không cần đăng nhập |

---

## 📁 Cấu Trúc Dự Án

```
DockerAI/
├── .env                          # Biến môi trường (KHÔNG commit)
├── .env.example                  # Template biến môi trường
├── docker-compose.yml            # Docker Compose (development)
├── docker-compose.prod.yml       # Docker Compose (production)
├── COMMANDS.md                   # File này
│
├── backend/                      # Go Backend (API Server)
│   ├── Dockerfile
│   ├── go.mod / go.sum
│   ├── cmd/server/main.go        # Entrypoint
│   ├── internal/
│   │   ├── config/               # Configuration
│   │   ├── domain/               # Domain models
│   │   ├── handler/httpapi/      # HTTP handlers & routes
│   │   ├── middleware/           # Auth, CORS, Rate Limit, Logger
│   │   ├── repository/          # MongoDB & Prometheus repos
│   │   ├── service/             # Business logic
│   │   └── pkg/                 # Internal utilities (Docker client, Logger)
│   └── pkg/api/                 # Public API types
│
├── frontend/                     # React Frontend (Vite + TypeScript)
│   ├── Dockerfile
│   ├── nginx.conf                # Nginx config (production)
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── App.tsx               # Router
│       ├── api/                  # API client
│       ├── components/           # Shared components
│       ├── pages/                # Page components
│       ├── hooks/                # Custom hooks (SSE)
│       ├── types/                # TypeScript types
│       └── utils/                # Utility functions
│
├── prometheus/                   # Prometheus Configuration
│   ├── prometheus.yml            # Scrape config
│   └── alert_rules.yml           # Alert rules
│
└── grafana/                      # Grafana Configuration
    ├── provisioning/
    │   ├── datasources/          # Prometheus datasource
    │   └── dashboards/           # Dashboard provisioner
    └── dashboards/               # JSON dashboard files
```

---

## ⚡ Quick Reference

```bash
# ═══════════════════════════════════════
#  LỆNH NHANH - Copy & Paste
# ═══════════════════════════════════════

# 🟢 CHẠY dự án
docker compose up -d --build

# 🔴 DỪNG dự án
docker compose down

# 🔄 RESTART dự án
docker compose restart

# 📋 XEM trạng thái
docker compose ps

# 📜 XEM logs
docker compose logs -f --tail=50

# ❤️ KIỂM TRA health
curl -s http://localhost:8080/healthz | python3 -m json.tool

# 🧹 DỌN DẸP hoàn toàn
docker compose down -v --rmi local
```
