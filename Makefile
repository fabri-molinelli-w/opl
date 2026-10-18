# Herramientas requeridas en el PATH:
#   oapi-codegen v2.8.0  go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0
#   sqlc v1.31.1         go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
#   golangci-lint v2.x   https://golangci-lint.run/welcome/install/

.PHONY: generate verify run-backend

generate:
	cd backend && oapi-codegen -config oapi-codegen.yaml ../api/openapi.yaml
	cd backend && sqlc generate
	cd backend && go mod tidy
	# Frontend (Orval) se agrega en el paso 0.6

verify:
	cd backend && go vet ./...
	cd backend && golangci-lint run
	cd backend && go test ./...

# Requiere la base local: docker compose up -d postgres
# Exportá HTTP_PORT / DATABASE_URL para sobrescribir los defaults.
run-backend:
	@test -f .env || { echo "Falta .env: copiá .env.example y completalo"; exit 1; }
	@set -a; . ./.env; set +a; \
	cd backend && \
	DATABASE_URL="postgres://$${POSTGRES_USER}:$${POSTGRES_PASSWORD}@127.0.0.1:5432/$${POSTGRES_DB}?sslmode=disable" \
	HTTP_PORT="$${HTTP_PORT:-8080}" \
	go run ./cmd/opl
