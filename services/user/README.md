# User Service

User Service là service cốt lõi của hệ thống e-commerce, chịu trách nhiệm authentication, authorization và quản lý user.

## Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL 15+
- **Storage**: MinIO (S3-compatible)
- **Framework**: Gin (HTTP)
- **Migration**: golang-migrate

## Prerequisites

- Go 1.21+
- PostgreSQL 15+
- MinIO (hoặc S3-compatible storage)
- golang-migrate (sẽ được cài tự động)

## Cài Đặt

### 1. Clone Repository

```bash
git clone <repository-url>
cd e-commerce-be/services/user
```

### 2. Cài Đặt Dependencies

```bash
go mod download
```

### 3. Cài Đặt Tools

```bash
make install-tools
```

## Cấu Hình

### 1. Tạo File `.env`

Tạo file `.env` trong thư mục `services/user/`:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=user_db
DB_SSL_MODE=disable

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minio
MINIO_SECRET_KEY=minio123
MINIO_USE_SSL=false
MINIO_REGION=us-east-1

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_ACCESS_TOKEN_EXPIRE=15m
JWT_REFRESH_TOKEN_EXPIRE=168h

# App
APP_NAME=user-service
APP_ENV=development
APP_PORT=8080
LOG_LEVEL=info
```

### 2. Khởi Động Infrastructure

Từ thư mục `deployment/`:

```bash
cd ../../deployment
docker-compose up -d
```

Hoặc nếu bạn đã có PostgreSQL và MinIO chạy sẵn, có thể bỏ qua bước này.

## Chạy Dự Án

### 1. Chạy Migrations

```bash
# Quay lại thư mục service
cd ../services/user

# Chạy migrations
make migrate-up

# Kiểm tra version
make migrate-version
```

### 2. Seed Data (Tùy chọn)

```bash
# Seed dữ liệu địa điểm Việt Nam
make seed-locations
```

### 3. Chạy Service

```bash
make run
```

Service sẽ chạy tại: `http://localhost:8080`

### 4. Kiểm Tra Health

```bash
curl http://localhost:8080/health
```

## Các Lệnh Hữu Ích

```bash
# Xem tất cả lệnh
make help

# Migrations
make migrate-up        # Chạy migrations
make migrate-down      # Rollback migrations
make migrate-version   # Xem version hiện tại
make migrate-create NAME=name  # Tạo migration mới

# Seeds
make seed-locations    # Seed dữ liệu địa điểm

# Build & Run
make build            # Build binary
make run              # Chạy service
make clean            # Xóa build artifacts

# Dependencies
make deps             # Download dependencies
```

## API Endpoints

### Public Endpoints

- `POST /api/auth/register` - Đăng ký user mới
- `POST /api/auth/login` - Đăng nhập
- `POST /api/auth/refresh` - Refresh access token

### Protected Endpoints (Cần JWT token)

- `GET /api/auth/me` - Lấy thông tin user hiện tại
- `POST /api/auth/logout` - Đăng xuất
- `PATCH /api/users/profile` - Cập nhật profile
- `POST /api/files/presigned-url` - Lấy presigned URL để upload file

## Cấu Trúc Dự Án

```
services/user/
├── config/          # Configuration
├── handler/         # HTTP handlers
├── service/         # Business logic
├── repository/      # Data access layer
├── model/           # Domain models
├── migrations/      # Database migrations
├── internal/        # Internal utilities
│   ├── db/         # Database connection
│   └── util/       # Utilities (JWT, password, etc.)
└── routes/         # Route setup & middleware
```

## License

This project is part of a graduation thesis.
