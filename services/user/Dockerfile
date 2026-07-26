# =============================================================================
# STAGE 1: Builder
# =============================================================================
# Dùng golang:1.25-alpine thay vì golang:1.25 (debian) vì:
#   - alpine nhỏ hơn ~700MB so với ~900MB
#   - đủ tool để build Go (gcc, musl-libc)
FROM golang:1.25-alpine AS builder

# Cài git vì `go mod download` có thể cần clone một số dependency từ VCS.
# ca-certificates cần cho HTTPS khi download module.
RUN apk add --no-cache git ca-certificates

# WORKDIR đặt thư mục làm việc bên trong container.
# Mọi lệnh COPY/RUN phía sau đều tương đối với /workspace.
WORKDIR /workspace

# -----------------------------------------------------------------------------
# Bước 1: Copy file module của ROOT repo trước
# -----------------------------------------------------------------------------
# Lý do copy go.mod/go.sum ROOT trước khi copy code:
#   Docker build có cache theo từng layer. Nếu go.mod/go.sum không đổi,
#   Docker sẽ dùng cache → bỏ qua `go mod download` → build nhanh hơn rất nhiều.
#   Chỉ khi thêm/xóa dependency thì layer này mới bị invalidate.
COPY go.mod go.sum ./

# Copy thư mục pkg/ (shared package dùng chung giữa các service).
# user-service import package này qua `replace` directive trong go.mod của nó:
#   replace github.com/hmquannnnn/e-commerce => ../../
# Tức là Go compiler sẽ tìm pkg/ tại /workspace/pkg/ → phải copy vào đây.
COPY pkg/ ./pkg/

# -----------------------------------------------------------------------------
# Bước 2: Copy file module của user-service
# -----------------------------------------------------------------------------
# Copy go.mod/go.sum của service vào đúng vị trí tương đối (services/user/)
# để `replace ../../` trỏ đúng về /workspace/ ở trên.
COPY services/user/go.mod services/user/go.sum ./services/user/

# Download tất cả dependencies được khai báo trong go.mod của user-service.
# Chạy từ thư mục services/user/ để Go đọc đúng go.mod của service này.
# Layer này được cache riêng: chỉ re-run khi go.mod/go.sum thay đổi.
WORKDIR /workspace/services/user
RUN go mod download

# -----------------------------------------------------------------------------
# Bước 3: Copy source code và build
# -----------------------------------------------------------------------------
# Quay về /workspace để copy toàn bộ source của service.
# Copy sau bước download để tận dụng cache: thay đổi code không làm
# mất cache layer `go mod download` ở trên.
WORKDIR /workspace
COPY services/user/ ./services/user/

WORKDIR /workspace/services/user

# Build binary với các flag tối ưu cho production:
#   CGO_ENABLED=0  → build pure Go, không link C library → binary chạy được
#                    trên alpine (musl libc) mà không cần runtime C
#   -ldflags="-w -s" → bỏ debug info (-w) và symbol table (-s)
#                       → giảm kích thước binary ~30%
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/user-service main.go

# =============================================================================
# STAGE 2: Runtime
# =============================================================================
# alpine:3.20 là base image tối thiểu (~5MB), chỉ chứa OS, không có Go toolchain.
# Kết quả: image cuối cùng ~15-20MB thay vì ~300MB nếu dùng golang image.
FROM alpine:3.20

# ca-certificates: cần để service có thể gọi HTTPS ra ngoài (ví dụ: external API).
# tzdata: cần nếu code dùng time.LoadLocation() để xử lý timezone.
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Chỉ copy binary đã được build từ stage 1.
# Không copy source code, không copy Go toolchain → image sạch và nhỏ.
COPY --from=builder /bin/user-service .

# Document port mà service lắng nghe (lấy từ config.go: APP_PORT default "8081").
# EXPOSE chỉ có tác dụng document, không tự động mở port.
# Port thực tế được cấu hình qua env var APP_PORT khi chạy container.
EXPOSE 8081

CMD ["./user-service"]
