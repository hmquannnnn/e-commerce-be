FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy root module files for the shared package
COPY go.mod go.sum ./
COPY pkg/ ./pkg/

# Copy service module
COPY services/product/ ./services/product/

WORKDIR /app/services/product

RUN go mod download
RUN go build -o /bin/product-service main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/product-service .

EXPOSE 8082

CMD ["./product-service"]
