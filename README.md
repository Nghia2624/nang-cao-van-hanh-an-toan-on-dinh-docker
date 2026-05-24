<h2 align="center">
    🎓 Khoa Công nghệ Thông tin (Đại học Đại Nam)
</h2>

<h2 align="center">
    ĐỒ ÁN TỐT NGHIỆP
</h2>
<h3 align="center">
    ĐỀ TÀI: NGHIÊN CỨU NÂNG CAO MỨC ĐỘ AN TOÀN TRONG VẬN HÀNH DOCKER
</h3>
<div align="center">
    <p align="center">
        <img src="anh_dulieu_cauhinh/dnu_logo.png" alt="DaiNam University Logo" width="200"/>
    </p>

[![Khoa Công nghệ Thông tin](https://img.shields.io/badge/Khoa%20Công%20nghệ%20Thông%20tin-blue?style=for-the-badge)](https://dainam.edu.vn/vi/khoa-cong-nghe-thong-tin)
[![Đại học Đại Nam](https://img.shields.io/badge/Đại%20học%20Đại%20Nam-orange?style=for-the-badge)](https://dainam.edu.vn)

</div>

## 📖 1. Giới thiệu
**Dashboard** là một hệ thống giám sát container trực quan, hỗ trợ người quản trị trong việc theo dõi, đánh giá và đảm bảo an toàn cho hệ thống Docker. Dashboard được phát triển như một phần của đồ án tốt nghiệp với đề tài **"Nghiên cứu nâng cao mức độ an toàn trong vận hành Docker"**, hướng đến mục tiêu tối ưu hóa khả năng quản lý và bảo mật trong môi trường container hóa. Hệ thống tích hợp phân tích log bằng AI (Google Gemini), đưa ra các cảnh báo an toàn và đánh giá tình trạng container theo thời gian thực, từ đó giúp người quản trị nhanh chóng phát hiện các bất thường và nâng cao tính ổn định cho hệ thống.

## 🔧 2. Các công nghệ được sử dụng
<div align="center">

### Hệ điều hành
[![Ubuntu](https://img.shields.io/badge/Ubuntu-E95420?style=for-the-badge&logo=ubuntu&logoColor=white)](https://ubuntu.com/)
### Backend & Frontend
[![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-007ACC?style=for-the-badge&logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
### Dữ liệu & Giám sát
[![MongoDB](https://img.shields.io/badge/MongoDB-4EA94B?style=for-the-badge&logo=mongodb&logoColor=white)](https://www.mongodb.com/)
[![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=for-the-badge&logo=prometheus&logoColor=white)](https://prometheus.io/)
[![Grafana](https://img.shields.io/badge/Grafana-F46800?style=for-the-badge&logo=grafana&logoColor=white)](https://grafana.com/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
</div>

## 🎯 3. Các tính năng chính

- **Giám sát Container theo thời gian thực**: Theo dõi toàn bộ trạng thái container, bao gồm CPU, Memory, Network và Disk I/O.
- **Phân tích Log bằng AI**: Tự động phân tích nguyên nhân của lỗi bằng Google Gemini.
- **Cảnh báo thông minh**: Hệ thống đánh giá các sự kiện sinh ra và cảnh báo khi có rủi ro.
- **Theo dõi tình trạng**: Đánh giá tình trạng của container.
- **Nhận diện bất thường**: Phát hiện lỗi lặp lại, lỗi đột biến tài nguyên hoặc lỗi dây chuyền.

## ⚙️ 4. Cài đặt

### 4.1. Yêu cầu hệ thống
- Môi trường đã cài đặt Docker & Docker Compose.
- Docker daemon đang chạy trên môi trường Host.
- Khóa API của Google Gemini (Sử dụng nhiều khóa API để thay phiên, giảm thiểu giới hạn request).

### 4.2. Tải mã nguồn
```bash
git clone https://github.com/Nghia2624/nang-cao-van-hanh-an-toan-on-dinh-docker.git
cd nang-cao-van-hanh-an-toan-on-dinh-docker
```

### 4.3. Cấu hình biến môi trường
Tạo tệp **.env** tại thư mục gốc với nội dung:
```bash
# Cấu hình AI (Bắt buộc)
AI_API_KEYS=key1,key2,key3
# HOẶC sử dụng 1 key duy nhất:
# AI_API_KEY=your-key-here

# Cấu hình API Backend (Tùy chọn)
APP_API_KEY=your-secure-api-key

# Cấu hình Frontend
VITE_API=http://localhost:8080
VITE_API_KEY=your-secure-api-key

# Grafana (Tùy chọn)
GRAFANA_USER=admin
GRAFANA_PASSWORD=admin

LOG_LEVEL=info
```

### 4.4. Khởi chạy hệ thống
Thực thi lệnh sau:
```bash
docker compose up -d
```
Sau khi chạy, bạn có thể truy cập các dịch vụ qua:
- **Frontend Dashboard**: `http://localhost:3000`
- **Backend API**: `http://localhost:8080`
- **Prometheus**: `http://localhost:9090`
- **Grafana**: `http://localhost:3001`
- **cAdvisor**: `http://localhost:8081`

## 📝 5. Công bố Khoa học

Đề tài đã được công bố thành bài báo khoa học:
- **Tên bài báo**: Giải pháp nâng cao an toàn, ổn định vận hành Docker, tích hợp LLM trong giám sát, phát hiện dấu hiệu bất thường
- **Nơi đăng**: [Kỷ yếu Hội thảo Khoa học Quốc tế về Công nghệ và Sức khỏe số 2026 (INCOTEH)](https://ebook365.vn/ky-yeu-hoi-thao-khoa-hoc-quoc-te-ve-cong-nghe-va-suc-khoe-so-2026-international-scientific-conference-on-technology-and-digital-health-2026-rGXXZW.html)
- **Trang**: 95-107

## 👨‍💻 6. Thông tin Phát triển

| Trường thông tin | Nội dung |
| --- | --- |
| 🏛️ **Trường** | Đại học Đại Nam (DaiNam University) |
| 💻 **Khoa** | Công nghệ Thông tin |
| 🎓 **Loại đồ án** | Đồ án tốt nghiệp |
| 👤 **Sinh viên** | Đỗ Ngọc Nghĩa |
| 📧 **Email** | dnghia9119@gmail.com |
| 🌐 **Website cá nhân** | [dnnghia.vercel.app](https://dnnghia.vercel.app) |
| 🏫 **Lớp** | CNTT 16-03 |
| 📅 **Năm học** | 2025-2026 |

<div align="center">
© 2026 Faculty of Information Technology, DaiNam University. All rights reserved.
</div>
