FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

COPY go.mod go.sum ./
COPY services/order/go.mod services/order/go.sum ./services/order/

WORKDIR /workspace/services/order
RUN go mod download

WORKDIR /workspace
COPY services/order/ ./services/order/

WORKDIR /workspace/services/order

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/order-service main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/order-service .

EXPOSE 8085

CMD ["./order-service"]
