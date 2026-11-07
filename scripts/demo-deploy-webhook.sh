#!/bin/sh
# Demo CU-07: simula al pipeline de CI/CD avisando un deploy a OPL.
# Uso: ./scripts/demo-deploy-webhook.sh [servicio] [environment] [versión] [resultado]
#   ej: ./scripts/demo-deploy-webhook.sh lab-nginx dev v1.1.0 success
# Lee DEPLOY_WEBHOOK_TOKEN de .env sin mostrarlo. OPL_URL por defecto: http://localhost:8080
set -eu
SERVICE="${1:-lab-nginx}"
ENVIRONMENT="${2:-dev}"
VERSION="${3:-v1.1.0}"
RESULT="${4:-success}"
OPL_URL="${OPL_URL:-http://localhost:8080}"
TOKEN="${DEPLOY_WEBHOOK_TOKEN:-$(grep -E '^DEPLOY_WEBHOOK_TOKEN=' .env | cut -d= -f2-)}"

curl -sS -X POST "$OPL_URL/api/deploys/webhook" \
  -H "Content-Type: application/json" \
  -H "X-Deploy-Token: $TOKEN" \
  -d "{\"serviceName\":\"$SERVICE\",\"environment\":\"$ENVIRONMENT\",\"version\":\"$VERSION\",\"result\":\"$RESULT\",\"author\":\"pipeline-demo\",\"notes\":\"Enviado por scripts/demo-deploy-webhook.sh\"}"
echo
