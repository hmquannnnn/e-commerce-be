#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

SERVICES=(
  "user-service:services/user"
  "file-service:services/file"
  "product-service:services/product"
  "order-service:services/order"
  "payment-service:services/payment"
)

GATEWAY="api-gateway:api-gateway"
GATEWAY_DELAY_SECONDS="${GATEWAY_DELAY_SECONDS:-3}"

PIDS=()
NAMES=()

log() {
  printf '[run-all] %s\n' "$*"
}

require_go() {
  if ! command -v go >/dev/null 2>&1; then
    log "go is not available in PATH"
    exit 1
  fi
}

start_process() {
  local name="$1"
  local rel_dir="$2"
  local abs_dir="$ROOT_DIR/$rel_dir"

  if [[ ! -f "$abs_dir/main.go" ]]; then
    log "missing main.go for $name at $abs_dir"
    exit 1
  fi

  log "starting $name from $rel_dir"
  (
    cd "$abs_dir"
    exec go run .
  ) &

  PIDS+=("$!")
  NAMES+=("$name")
}

stop_all() {
  local status="${1:-0}"

  if [[ "${#PIDS[@]}" -gt 0 ]]; then
    log "stopping services"
    for pid in "${PIDS[@]}"; do
      if kill -0 "$pid" >/dev/null 2>&1; then
        kill "$pid" >/dev/null 2>&1 || true
      fi
    done
  fi

  exit "$status"
}

on_signal() {
  log "received stop signal"
  stop_all 130
}

trap on_signal INT TERM

require_go

for entry in "${SERVICES[@]}"; do
  IFS=':' read -r name rel_dir <<<"$entry"
  start_process "$name" "$rel_dir"
done

log "waiting ${GATEWAY_DELAY_SECONDS}s before starting api-gateway"
sleep "$GATEWAY_DELAY_SECONDS"

IFS=':' read -r gateway_name gateway_dir <<<"$GATEWAY"
start_process "$gateway_name" "$gateway_dir"

log "all backend services are starting"
log "api-gateway: http://localhost:8080"
log "product-service direct: http://localhost:8083"
log "press Ctrl+C to stop all services"

while true; do
  for i in "${!PIDS[@]}"; do
    pid="${PIDS[$i]}"
    name="${NAMES[$i]}"

    if ! kill -0 "$pid" >/dev/null 2>&1; then
      set +e
      wait "$pid"
      status="$?"
      set -e
      log "$name exited with status $status"
      stop_all "$status"
    fi
  done

  sleep 1
done
