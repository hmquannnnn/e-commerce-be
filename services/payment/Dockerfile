FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY services/payment/go.mod services/payment/go.sum ./services/payment/

WORKDIR /workspace/services/payment
RUN go mod download

WORKDIR /workspace
COPY services/payment/ ./services/payment/

WORKDIR /workspace/services/payment
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/payment-service main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/payment-service .

EXPOSE 8086

CMD ["./payment-service"]
