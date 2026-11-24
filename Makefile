.PHONY: all build test up down logs clean help

help:
	@echo "Available commands:"
	@echo "  make up       - Launch the complete production stack with Docker Compose"
	@echo "  make down     - Stop all containers and teardown networks"
	@echo "  make build    - Build backend and frontend artifacts locally"
	@echo "  make test     - Run backend domain invariant and integration unit tests"
	@echo "  make logs     - Follow aggregated logs from all services"
	@echo "  make clean    - Remove local build artifacts and databases"

build:
	@echo "==> Building Go backend..."
	cd backend && go build -v -o todo-api ./cmd/api
	@echo "==> Building Next.js frontend..."
	cd frontend && npm run build

test:
	@echo "==> Running Go backend unit & invariant tests..."
	cd backend && go test -v -race ./...

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	rm -f backend/todo-api
	rm -rf backend/data
	rm -rf frontend/.next
