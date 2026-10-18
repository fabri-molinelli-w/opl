# Herramientas requeridas en el PATH:
#   oapi-codegen v2.8.0  go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0
#   sqlc v1.31.1         go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
#   golangci-lint v2.x   https://golangci-lint.run/welcome/install/
#   node >= 24           Vite 8 / Orval 8 / Vitest 5

.PHONY: generate verify run-backend

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

# Requiere la base local: docker compose up -d postgres
# Exportá HTTP_PORT / DATABASE_URL para sobrescribir los defaults.
run-backend:
	@test -f .env || { echo "Falta .env: copiá .env.example y completalo"; exit 1; }
	@set -a; . ./.env; set +a; \
	cd backend && \
	DATABASE_URL="postgres://$${POSTGRES_USER}:$${POSTGRES_PASSWORD}@127.0.0.1:5432/$${POSTGRES_DB}?sslmode=disable" \
	HTTP_PORT="$${HTTP_PORT:-8080}" \
	go run ./cmd/opl

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
