#!/usr/bin/env bash
# Brings up local dev infra (colima + Postgres/pgvector), applies migrations,
# then runs the API in the foreground. Safe to re-run: every step is
# idempotent (colima/docker compose skip work that's already done).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

log() { printf '\033[1;34m[dev]\033[0m %s\n' "$1"; }
die() { printf '\033[1;31m[dev]\033[0m %s\n' "$1" >&2; exit 1; }

command -v colima >/dev/null 2>&1 || die "colima bulunamadı. 'brew install colima docker docker-compose' ile kurun."
command -v docker >/dev/null 2>&1 || die "docker CLI bulunamadı. 'brew install colima docker docker-compose' ile kurun."

# 1. Ensure colima (Docker runtime) is running.
if colima status >/dev/null 2>&1; then
  log "colima zaten çalışıyor."
else
  log "colima başlatılıyor..."
  colima start
fi

# 2. Wait for the Docker daemon to accept connections.
log "Docker daemon bekleniyor..."
for _ in $(seq 1 30); do
  docker info >/dev/null 2>&1 && break
  sleep 1
done
docker info >/dev/null 2>&1 || die "Docker daemon 30 saniye içinde hazır olmadı."

# 3. Start Postgres (pgvector).
cd "$BACKEND_DIR"
log "Postgres (pgvector) konteyneri başlatılıyor..."
docker compose up -d

# 4. Wait for Postgres to report healthy.
log "Postgres sağlık kontrolü bekleniyor..."
status=""
for _ in $(seq 1 30); do
  status="$(docker compose ps --format '{{.Health}}' postgres 2>/dev/null || true)"
  [ "$status" = "healthy" ] && break
  sleep 1
done
if [ "$status" != "healthy" ]; then
  docker compose logs postgres | tail -50
  die "Postgres 'healthy' durumuna geçmedi (son durum: ${status:-bilinmiyor})."
fi
log "Postgres hazır (port ${DB_PORT:-5433})."

# 5. Apply migrations.
log "Migration'lar uygulanıyor..."
make migrate-up

# 6. Run the API in the foreground (Ctrl+C stops it; infra keeps running).
log "Backend başlatılıyor — durdurmak için Ctrl+C (Postgres konteyneri arka planda kalır)."
exec make run
