FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

COPY go.mod go.sum ./

COPY pkg/ ./pkg/

COPY services/file/go.mod services/file/go.sum ./services/file/

WORKDIR /workspace/services/file
RUN go mod download

WORKDIR /workspace
COPY services/file/ ./services/file/

WORKDIR /workspace/services/file

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/file-service main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/file-service .

EXPOSE 8082

CMD ["./file-service"]