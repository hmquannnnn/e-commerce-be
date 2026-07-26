FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

COPY go.mod go.sum ./
COPY pkg/ ./pkg/

COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
WORKDIR /workspace/api-gateway
RUN go mod download

WORKDIR /workspace
COPY api-gateway/ ./api-gateway/
WORKDIR /workspace/api-gateway
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/api-gateway main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
