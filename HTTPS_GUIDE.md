# HTTPS Setup Guide

Hướng dẫn cài đặt HTTPS cho Manager Bot API sử dụng Docker Compose.

## Tổng quan

Setup này sử dụng:
- **Nginx** làm reverse proxy với SSL/TLS termination
- **Self-signed certificates** cho localhost development
- **Docker Compose** để orchestrate các services

## Kiến trúc

```
Browser (HTTPS) --> Nginx (SSL Termination) --> Go App (HTTP)
```

## Cấu trúc files

```
├── nginx/
│   └── nginx.conf              # Nginx configuration
├── certs/                      # SSL certificates (generated)
│   ├── localhost.crt           # Certificate file
│   └── localhost.key           # Private key file
├── generate-ssl.sh             # Script to generate SSL certificates
├── setup-https.sh              # Main setup script
└── docker-compose.yml          # Updated with nginx service
```

## Cách sử dụng

### 1. Quick Setup
```bash
./setup-https.sh
```

### 2. Manual Setup

#### Bước 1: Tạo SSL certificates
```bash
./generate-ssl.sh
```

#### Bước 2: Khởi động services
```bash
docker-compose up --build -d
```

## Services

Sau khi setup thành công, bạn có thể truy cập:

| Service | URL | Description |
|---------|-----|-------------|
| HTTPS API | https://localhost | Main API với HTTPS |
| HTTP API | http://localhost | Redirect tự động tới HTTPS |
| MySQL | localhost:3306 | Database |
| Redis | localhost:6379 | Cache |
| phpMyAdmin | http://localhost:8080 | Database management |

## SSL Certificate Details

Self-signed certificate được tạo với thông tin:
- **Country**: VN
- **State**: HoChiMinh  
- **City**: HoChiMinh
- **Organization**: LocalDev
- **OU**: IT
- **Common Name**: localhost
- **Subject Alternative Names**: 
  - DNS: localhost
  - DNS: *.localhost
  - IP: 127.0.0.1

## Security Headers

Nginx được cấu hình với các security headers:
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=63072000; includeSubDomains; preload`

## Browser Warning

Khi truy cập lần đầu, browser sẽ hiển thị cảnh báo bảo mật vì đây là self-signed certificate:

1. Click **"Advanced"** 
2. Click **"Proceed to localhost (unsafe)"**

## Troubleshooting

### 1. Port đã được sử dụng
```bash
# Kiểm tra process đang sử dụng port 80/443
sudo lsof -i :80
sudo lsof -i :443

# Stop service nếu cần
sudo kill -9 <PID>
```

### 2. Permission denied khi generate certificates
```bash
# Đảm bảo script có quyền execute
chmod +x generate-ssl.sh
chmod +x setup-https.sh
```

### 3. Docker compose fails
```bash
# Stop và clean up
docker-compose down
docker system prune -f

# Rebuild
docker-compose up --build -d
```

### 4. Xem logs
```bash
# Tất cả services
docker-compose logs -f

# Specific service
docker-compose logs -f nginx
docker-compose logs -f app
```

## Commands hữu ích

```bash
# Stop all services
docker-compose down

# Start services
docker-compose up -d

# Rebuild and start
docker-compose up --build -d

# View logs
docker-compose logs -f

# Access container shell
docker-compose exec app sh
docker-compose exec nginx sh

# Recreate SSL certificates
rm -rf certs/*
./generate-ssl.sh
docker-compose restart nginx
```

## Production Notes

⚠️ **Lưu ý**: Setup này chỉ dành cho development. Đối với production:

1. Sử dụng certificates từ trusted CA (Let's Encrypt, etc.)
2. Cấu hình proper domain names
3. Thêm rate limiting và security measures
4. Sử dụng secrets management cho sensitive data