FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

COPY go.mod go.sum ./

COPY pkg/ ./pkg/

COPY services/product/go.mod services/product/go.sum ./services/product/

WORKDIR /workspace/services/product
RUN go mod download

WORKDIR /workspace
COPY services/product/ ./services/product/

WORKDIR /workspace/services/product

RUN go build -o /bin/product-service main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/product-service .

EXPOSE 8083 9083

CMD ["./product-service"]
