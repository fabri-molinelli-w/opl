# Herramientas requeridas en el PATH:
#   oapi-codegen v2.8.0  go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0
#   sqlc v1.31.1         go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
#   golangci-lint v2.x   https://golangci-lint.run/welcome/install/
#   node >= 24           Vite 8 / Orval 8 / Vitest 5

.PHONY: generate verify verify-ai setup-ai run-ai run-backend seed test-integration

generate:
	cd backend && oapi-codegen -config oapi-codegen.yaml ../api/openapi.yaml
	cd backend && sqlc generate
	cd backend && go mod tidy
	cd frontend && npm run generate

verify:
	cd backend && go vet ./...
	cd backend && golangci-lint run
	cd backend && go test ./...
	cd frontend && npm run lint
	cd frontend && npm run typecheck
	cd frontend && npm run test -- --run
	@if [ -x ai-service/.venv/bin/python ]; then $(MAKE) verify-ai; else echo "ai-service: correr make setup-ai para verificarlo"; fi

# --- OPSBot (ai-service, Python >= 3.12) ---
setup-ai:
	cd ai-service && python3 -m venv .venv && .venv/bin/pip install -q --upgrade pip && .venv/bin/pip install -q -e '.[dev]'

verify-ai:
	cd ai-service && .venv/bin/ruff check . && .venv/bin/ruff format --check . && .venv/bin/mypy . && .venv/bin/pytest -q

# Requiere make setup-ai y las variables LLM_* y AI_INTERNAL_TOKEN en .env.
# Escucha solo en 127.0.0.1: el ai-service nunca se expone (lo llama el backend).
run-ai:
	@test -f .env || { echo "Falta .env: copiá .env.example y completalo"; exit 1; }
	@set -a; . ./.env; set +a; \
	cd ai-service && .venv/bin/uvicorn app.main:app --host 127.0.0.1 --port 8000

# Requiere la base local: docker compose up -d postgres
# Exportá HTTP_PORT / DATABASE_URL para sobrescribir los defaults.
run-backend:
	@test -f .env || { echo "Falta .env: copiá .env.example y completalo"; exit 1; }
	@set -a; . ./.env; set +a; \
	cd backend && \
	DATABASE_URL="postgres://$${POSTGRES_USER}:$${POSTGRES_PASSWORD}@127.0.0.1:5432/$${POSTGRES_DB}?sslmode=disable" \
	HTTP_PORT="$${HTTP_PORT:-8080}" \
	go run ./cmd/opl

# Datos ficticios de inventario. Solo para desarrollo local.
# Requiere Postgres local y el backend arrancado al menos una vez (migraciones).
seed:
	@test -f .env || { echo "Falta .env: copiá .env.example y completalo"; exit 1; }
	@set -a; . ./.env; set +a; \
	cd backend && \
	DATABASE_URL="postgres://$${POSTGRES_USER}:$${POSTGRES_PASSWORD}@127.0.0.1:5432/$${POSTGRES_DB}?sslmode=disable" \
	go run ./cmd/seed

# Tests de integración del inventario contra un Postgres real (CA-4, RN-6).
# Crea la base opl_test si no existe. Requiere docker compose up -d postgres.
test-integration:
	@test -f .env || { echo "Falta .env: copiá .env.example y completalo"; exit 1; }
	@set -a; . ./.env; set +a; \
	TEST_DB="$${TEST_DB:-opl_test}"; \
	docker compose exec -T postgres createdb -U "$${POSTGRES_USER}" "$$TEST_DB" 2>/dev/null || true; \
	cd backend && \
	TEST_DATABASE_URL="postgres://$${POSTGRES_USER}:$${POSTGRES_PASSWORD}@127.0.0.1:5432/$$TEST_DB?sslmode=disable" \
	go test -tags integration ./internal/repository/...

# Requiere el backend en http://localhost:8080 (make run-backend)
.PHONY: run-frontend
run-frontend:
	cd frontend && npm run dev

# Verifica que el código generado (oapi-codegen, sqlc, Orval) esté sincronizado
# con api/openapi.yaml. Mismo chequeo que el job "contract" del CI.
# Requiere Go y Node >= 24 (con las deps del frontend instaladas: cd frontend && npm ci).
.PHONY: check-generated
check-generated:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
	PATH="$$PATH:$$(go env GOPATH)/bin" $(MAKE) generate
	git diff --exit-code -- backend/internal/api/gen backend/internal/db/gen frontend/src/api/gen backend/go.mod backend/go.sum
