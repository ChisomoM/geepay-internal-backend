.PHONY: help setup db-up db-down run build clean test lint new-module

help:
	@echo "Available commands:"
	@echo "  make setup        - Install dependencies"
	@echo "  make db-up        - Start PostgreSQL"
	@echo "  make db-down      - Stop PostgreSQL"
	@echo "  make run          - Run the server"
	@echo "  make build        - Build binary"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run linter"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make new-module   - Create a new module (e.g., make new-module NAME=projects)"

setup:
	go mod download
	go mod tidy

db-up:
	docker-compose up -d postgres

db-down:
	docker-compose down

run: db-up
	go run main.go

build:
	go build -o bin/server main.go

test:
	go test -v ./...

lint:
	go fmt ./...
	go vet ./...

clean:
	rm -rf bin/
	rm -f *.log

new-module:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make new-module NAME=modulename"; \
		exit 1; \
	fi
	@mkdir -p modules/$(NAME)
	@cp -r modules/_template/* modules/$(NAME)/
	@find modules/$(NAME)/ -type f -exec sed -i 's/MODULENAME/$(NAME)/g' {} +
	@echo "Created new module: modules/$(NAME)/"
	@echo "Next steps:"
	@echo "  1. Edit modules/$(NAME)/service.go - implement business logic"
	@echo "  2. Edit modules/$(NAME)/handler.go - implement HTTP handlers"
	@echo "  3. Register routes in routes/routes.go"
	@echo "  4. Add models in models/$(NAME).go if needed"
