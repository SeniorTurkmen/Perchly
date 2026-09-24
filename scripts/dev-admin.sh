#!/usr/bin/env bash
# Brings up everything the admin dashboard needs, in order: backend
# infra (colima + Postgres/pgvector), migrations, a first admin account
# if none exists yet, the Go API, and the Next.js admin app — then runs
# both servers until Ctrl+C. Safe to re-run: every step is idempotent.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
ADMIN_DIR="$ROOT_DIR/admin"
BACKEND_PORT="${SERVER_PORT:-8080}"
BACKEND_LOG="/tmp/perchly-backend.log"

log() { printf '\033[1;34m[dev-admin]\033[0m %s\n' "$1"; }
die() { printf '\033[1;31m[dev-admin]\033[0m %s\n' "$1" >&2; exit 1; }

command -v colima >/dev/null 2>&1 || die "colima bulunamadı. 'brew install colima docker docker-compose' ile kurun."
command -v docker >/dev/null 2>&1 || die "docker CLI bulunamadı. 'brew install colima docker docker-compose' ile kurun."
command -v npm >/dev/null 2>&1 || die "npm bulunamadı. Node.js kurun (https://nodejs.org)."

# 1. Ensure colima (Docker runtime) is running.
if colima status >/dev/null 2>&1; then
  log "colima zaten çalışıyor."
else
  log "colima başlatılıyor..."
  colima start
fi

# 2. Wait for the Docker daemon.
log "Docker daemon bekleniyor..."
for _ in $(seq 1 30); do
  docker info >/dev/null 2>&1 && break
  sleep 1
done
docker info >/dev/null 2>&1 || die "Docker daemon 30 saniye içinde hazır olmadı."

# 3. backend/.env — required by the Makefile (it does `include .env`).
if [ ! -f "$BACKEND_DIR/.env" ]; then
  log "backend/.env yok, .env.example'dan oluşturuluyor (LLM sağlayıcı anahtarını sonradan girebilirsin)..."
  cp "$BACKEND_DIR/.env.example" "$BACKEND_DIR/.env"
fi

# 4. Start Postgres (pgvector).
cd "$BACKEND_DIR"
log "Postgres (pgvector) konteyneri başlatılıyor..."
docker compose up -d

# 5. Wait for Postgres to report healthy.
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
log "Postgres hazır."

# 6. Apply migrations (creates admin_users/admin_sessions/admin_audit_log too).
log "Migration'lar uygulanıyor..."
make migrate-up

# 7. Make sure at least one admin account exists.
admin_count="$(docker compose exec -T postgres psql -U perchly -d perchly -tAc 'SELECT count(*) FROM admin_users' 2>/dev/null || echo 0)"
if [ "${admin_count:-0}" -eq 0 ]; then
  log "Henüz admin hesabı yok, bir tane oluşturuluyor."
  admin_email="${ADMIN_EMAIL:-}"
  [ -n "$admin_email" ] || read -rp "Admin e-posta: " admin_email
  admin_password="${ADMIN_PASSWORD:-}"
  if [ -z "$admin_password" ]; then
    read -rsp "Admin şifre: " admin_password
    echo
  fi
  make admin-create email="$admin_email" password="$admin_password"
else
  log "Admin hesabı zaten mevcut ($admin_count), atlanıyor (yeni biri için: make admin-create email=... password=...)."
fi

# 8. admin/ dependencies + local env.
cd "$ADMIN_DIR"
if [ ! -d node_modules ]; then
  log "admin/ bağımlılıkları kuruluyor (npm install)..."
  npm install
fi
if [ ! -f .env.local ]; then
  log "admin/.env.local yok, .env.example'dan oluşturuluyor..."
  cp .env.example .env.local
fi

# 9. Run the backend API in the background; the admin dashboard runs in
# the foreground below, so Ctrl+C stops both together. Built to a real
# binary and run directly (not `make run` / `go run`, which interpose a
# process that doesn't reliably relay a kill to the compiled binary it
# spawns, leaving it orphaned on port 8080 after this script exits).
cd "$BACKEND_DIR"
log "Backend derleniyor..."
make build
log "Backend API arka planda başlatılıyor (log: $BACKEND_LOG)..."
./bin/api >"$BACKEND_LOG" 2>&1 &

cleanup() {
  trap - EXIT INT TERM # avoid re-entering this from the kill below
  log "Kapatılıyor..."
  # Signal the whole process group, not just the backend — `npm run
  # dev` below spawns next dev -> next-server as further descendants
  # that killing a single pid doesn't reach, leaving them orphaned on
  # :3000 (and bin/api orphaned on :8080) after this script exits.
  # $$ is this script's own pid, NOT its process group id (this script
  # isn't the group leader — whatever invoked it is), so the group id
  # has to be looked up rather than assumed to equal $$.
  pgid="$(ps -o pgid= -p $$ | tr -d ' ')"
  kill -TERM -- "-$pgid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

log "Backend'in ayağa kalkması bekleniyor..."
backend_ready=""
for _ in $(seq 1 30); do
  if curl -s -o /dev/null "http://localhost:$BACKEND_PORT/health"; then
    backend_ready=1
    break
  fi
  sleep 1
done
[ -n "$backend_ready" ] || die "Backend 30 saniye içinde ayağa kalkmadı, bkz: $BACKEND_LOG"
log "Backend hazır (:$BACKEND_PORT)."

cd "$ADMIN_DIR"
log "Admin panel başlatılıyor (http://localhost:3000) — durdurmak için Ctrl+C (backend de birlikte kapanır)."
# Backgrounded + waited on explicitly, not run as a plain foreground
# command: bash only runs a trap right away while it's blocked in the
# `wait` builtin. While blocked waiting on a synchronous foreground
# command instead, it defers the trap until that command exits — which
# `npm run dev` never does on its own, so Ctrl+C would appear to do
# nothing.
npm run dev &
wait $!
