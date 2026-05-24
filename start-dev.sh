#!/bin/bash
set -euo pipefail

# ============================================================
# start-dev.sh — Entorno de desarrollo local PRODE
#
# Levanta PostgreSQL via Docker y corre la API con base local,
# SIN tocar el .env de produccion.
#
# Uso:   ./start-dev.sh
# ============================================================

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

echo "==> Levantando PostgreSQL local..."
docker compose up -d db

echo "==> Esperando a que PostgreSQL este lista..."
until docker compose exec db pg_isready -U powermix -d powermix -q 2>/dev/null; do
	sleep 1
done
echo "==> PostgreSQL lista!"

# Exportar variables locales.
# godotenv.Load() NO pisa variables ya definidas en el entorno,
# asi que estas prevalecen sobre lo que este en .env (productivo).
export DSN="postgresql://powermix:devpass@localhost:5432/powermix?sslmode=disable"

export PRODE_ENABLED=true
export PRODE_MAINTENANCE_ENABLED=true
export PRODE_ADMIN_API_KEY="dev-key-123"
export PRODE_ADMIN_EMAILS="admin@test.com"

echo "==> DSN local: ${DSN}"
echo "==> PRODE_ENABLED: ${PRODE_ENABLED}"
echo "==> PRODE_ADMIN_API_KEY: ${PRODE_ADMIN_API_KEY}"
echo ""
echo "==> Iniciando API..."
go run ./cmd/api
