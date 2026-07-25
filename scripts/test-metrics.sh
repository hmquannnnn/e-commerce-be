#!/usr/bin/env bash
set -uo pipefail

declare -A SERVICES=(
  [api-gateway]=8080
  [user-service]=8081
  [file-service]=8082
  [product-service]=8083
  [order-service]=8085
  [payment-service]=8086
)

HOST="http://localhost"
up_count=0
down_count=0

check_service() {
  local name="$1"
  local port="$2"
  local base="$HOST:$port"

  code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 2 "$base/metrics")

  if [[ "$code" == "200" ]]; then
    echo "Service $name is UP (HTTP $code)"
    ((up_count++))
  else
    echo "Service $name is DOWN (HTTP $code)"
    ((down_count++))
  fi
  echo "---------------------DONE-------------------"
}

for name in "${!SERVICES[@]}"; do
  check_service "$name" "${SERVICES[$name]}"
done

echo "Tổng kết: $up_count UP, $down_count DOWN"