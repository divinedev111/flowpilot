#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
PID_DIR="$ROOT_DIR/.pids"

mkdir -p "$PID_DIR"

# Source .env for port variables
if [ -f "$ROOT_DIR/.env" ]; then
    set -a
    . "$ROOT_DIR/.env"
    set +a
fi

BACKEND_PORT="${PORT:-8080}"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"

start_backend() {
    if [ -f "$PID_DIR/backend.pid" ] && kill -0 "$(cat "$PID_DIR/backend.pid")" 2>/dev/null; then
        echo "Backend already running (PID $(cat "$PID_DIR/backend.pid"))"
        return
    fi
    echo "Building backend..."
    cd "$ROOT_DIR/backend"
    go build -o "$ROOT_DIR/backend/backend" ./cmd/server
    echo "Starting backend..."
    cd "$ROOT_DIR"
    "$ROOT_DIR/backend/backend" > "$ROOT_DIR/backend/backend.log" 2>&1 &
    echo $! > "$PID_DIR/backend.pid"
    echo "Backend started (PID $!) → http://localhost:${BACKEND_PORT}"
}

start_frontend() {
    if [ -f "$PID_DIR/frontend.pid" ] && kill -0 "$(cat "$PID_DIR/frontend.pid")" 2>/dev/null; then
        echo "Frontend already running (PID $(cat "$PID_DIR/frontend.pid"))"
        return
    fi
    echo "Starting frontend..."
    cd "$ROOT_DIR/frontend"
    npx vite --port "$FRONTEND_PORT" > "$ROOT_DIR/frontend/frontend.log" 2>&1 &
    echo $! > "$PID_DIR/frontend.pid"
    echo "Frontend started (PID $!) → http://localhost:${FRONTEND_PORT}"
}

stop_service() {
    local name=$1
    local pidfile="$PID_DIR/${name}.pid"
    if [ -f "$pidfile" ]; then
        local pid
        pid=$(cat "$pidfile")
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid"
            echo "Stopped $name (PID $pid)"
        else
            echo "$name not running (stale PID)"
        fi
        rm -f "$pidfile"
    else
        echo "$name not running"
    fi
}

status() {
    for svc in backend frontend; do
        local pidfile="$PID_DIR/${svc}.pid"
        if [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile")" 2>/dev/null; then
            echo "$svc: running (PID $(cat "$pidfile"))"
        else
            echo "$svc: stopped"
        fi
    done
}

logs() {
    local name=${1:-}
    if [ -z "$name" ]; then
        echo "Usage: $0 logs <backend|frontend>"
        exit 1
    fi
    case "$name" in
        backend)  tail -f "$ROOT_DIR/backend/backend.log" ;;
        frontend) tail -f "$ROOT_DIR/frontend/frontend.log" ;;
        *)        echo "Unknown service: $name"; exit 1 ;;
    esac
}

case "${1:-}" in
    start)
        start_backend
        start_frontend
        echo ""
        echo "All services started. Use '$0 logs <backend|frontend>' to view logs."
        ;;
    stop)
        stop_service frontend
        stop_service backend
        echo "All services stopped."
        ;;
    restart)
        stop_service frontend
        stop_service backend
        sleep 1
        start_backend
        start_frontend
        echo ""
        echo "All services restarted."
        ;;
    status)
        status
        ;;
    logs)
        logs "${2:-}"
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs <backend|frontend>}"
        exit 1
        ;;
esac
