# File Service

File service for handling file uploads and storage operations using MinIO.

## Features

- Generate presigned URLs for file uploads
- Support multiple file types (avatars, images, products, documents)
- MinIO integration for object storage
- RESTful API

## Prerequisites

- Go 1.23 or higher
- MinIO server running
- Environment variables configured

## Configuration

Create a `.env` file in the `services/file/` directory:

```env
# Application
APP_NAME=file-service
APP_ENV=development
APP_PORT=8082
LOG_LEVEL=info

# Storage (MinIO)
MINIO_ENDPOINT=http://localhost:9000
MINIO_ACCESS_KEY=minio
MINIO_SECRET_KEY=minio123
MINIO_USE_SSL=false
MINIO_REGION=us-east-1
STORAGE_BUCKET_NAME=app-files

# JWT (for authentication)
JWT_SECRET=your-jwt-secret-here
```

## Installation

```bash
# Install dependencies
make install

# Or manually
go mod download
```

## Running the Service

```bash
# Run the service
make run

# Or manually
go run main.go
```

The service will start on port 8082 (or the port specified in `.env`).

## API Endpoints

### Health Check
- `GET /health` - Check service health
- `GET /api/health` - Check service health

### File Operations
- `POST /api/files/presigned-url` - Generate presigned upload URL

#### Example Request:
```json
POST /api/files/presigned-url
Content-Type: application/json
Authorization: Bearer <token>

{
  "file_type": "avatar",
  "content_type": "image/jpeg"
}
```

#### Example Response:
```json
{
  "success": true,
  "message": "Presigned URL generated successfully",
  "data": {
    "presigned_url": "http://localhost:9000/app-files/avatars/uuid.jpg?...",
    "file_path": "app-files/avatars/uuid.jpg",
    "expires_in": 900
  }
}
```

### Supported File Types
- `avatar` - User avatars
- `image` - General images
- `product` - Product images
- `document` - Documents

### Supported Content Types
- `image/jpeg`
- `image/png`
- `image/gif`
- `image/webp`
- `image/svg+xml`
- `application/pdf`
- `video/mp4`
- `video/quicktime`

## Development

```bash
# Run with hot reload (if air is installed)
make dev

# Build binary
make build

# Run tests
make test

# Clean build artifacts
make clean
```

## Architecture

```
services/file/
├── config/          # Configuration management
├── handler/         # HTTP handlers
├── routes/          # Route definitions and middlewares
├── service/         # Business logic
├── main.go          # Application entry point
├── go.mod           # Go module definition
└── Makefile         # Build commands
```

## Notes

- Authentication is handled by the API Gateway
- This service assumes requests are already authenticated
- Files are stored in MinIO with organized prefixes
- Presigned URLs expire after 15 minutes
