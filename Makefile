.PHONY: dev-backend dev-frontend dev install test lint

install:
	cd frontend && npm install
	cd backend && go mod tidy

dev-backend:
	cd backend && go run main.go

dev-frontend:
	cd frontend && npm run dev

dev:
	@echo "Starting backend and frontend..."
	@trap 'kill 0' EXIT; \
	cd backend && go run main.go & \
	cd frontend && npm run dev & \
	wait

test:
	cd backend && go test ./... -v

lint:
	cd backend && golangci-lint run ./...
	cd frontend && npm run lint
