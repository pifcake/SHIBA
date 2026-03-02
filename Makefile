.PHONY: all build run test migrate-up migrate-down docker-up docker-down docker-build dev-db

# Build the backend
build:
	go build -o bin/server ./cmd/api/main.go

# Run locally (requires .env loaded)
run:
	go run ./cmd/api/main.go

# Run tests
test:
	go test ./...

# Lint
lint:
	golangci-lint run ./...

# Format
fmt:
	gofmt -w .

# Docker
docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f api

docker-restart:
	docker compose restart api

# Full reset (WARNING: destroys data)
docker-reset:
	docker compose down -v
	docker compose up -d

# Database migrations (run outside Docker)
migrate-up:
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DB_URL)" down 1

# Start only DB + MinIO for local development
dev-infra:
	docker compose up -d postgres minio minio-init

# Generate admin user (run after first migration)
create-admin:
	@echo "Use POST /api/v1/auth/register with role=admin to create admin"
	@echo "Then manually set status='active' in DB:"
	@echo "  UPDATE users SET status='active' WHERE email='admin@example.com';"

# Dependency management
deps:
	go mod tidy
	go mod verify

# Go module download
download:
	go mod download

# Build Flutter web
flutter-web:
	cd frontend && flutter build web --release

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build backend binary"
	@echo "  run           - Run backend locally"
	@echo "  test          - Run tests"
	@echo "  docker-up     - Start all services"
	@echo "  docker-down   - Stop all services"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-reset  - Full reset (destroys data)"
	@echo "  dev-infra     - Start only DB+MinIO"
	@echo "  deps          - Tidy and verify dependencies"
	@echo "  flutter-web   - Build Flutter web"
