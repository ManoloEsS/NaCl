.PHONY: dev dev-full db-start db-stop db-restart db-migrate db-down db-reset db-new db-status db-test-clean db-init build run test clean ci fmt fmt-check

# Load .env file
include .env
export

# Development: run Go app with hot reload
dev:
	air -c .air.toml

# Full dev: init db, start db, migrate, run dev
dev-full: db-start
	sleep 1
	$(MAKE) db-init
	$(MAKE) db-migrate
	$(MAKE) dev

# Database: initialize (create databases if not exist)
db-init:
	@echo "Creating databases if not exist..."
	@docker exec nacl_postgres psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'nacl_dev'" | grep -q 1 || \
		docker exec nacl_postgres psql -U postgres -c "CREATE DATABASE nacl_dev;"
	@docker exec nacl_postgres psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'nacl_test'" | grep -q 1 || \
		docker exec nacl_postgres psql -U postgres -c "CREATE DATABASE nacl_test;"
	@echo "Databases ready."

# Database: start PostgreSQL container
db-start:
	docker start nacl_postgres 2>/dev/null || \
		docker run -d --name nacl_postgres \
			-e POSTGRES_PASSWORD=postgres \
			-v nacl_dev_data:/var/lib/postgresql \
			-p 5432:5432 \
			postgres:18
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 2
	@$(MAKE) db-init

# Database: stop container
db-stop:
	docker stop nacl_postgres 2>/dev/null || true

# Database: restart container
db-restart:
	docker restart nacl_postgres

# Database: run migrations up
db-migrate:
	goose postgres "$(DATABASE_URL)" -dir sql/migrations up

# Database: rollback last migration
db-down:
	goose postgres "$(DATABASE_URL)" -dir sql/migrations down

# Database: reset (stop, remove container, clean volume, start, migrate)
db-reset: db-stop
	docker rm -f nacl_postgres 2>/dev/null || true
	docker volume rm nacl_dev_data
	$(MAKE) db-start
	sleep 3
	$(MAKE) db-migrate

# Database: create new migration
db-new:
	goose -dir sql/migrations create $(name) sql

# Database: check migration status
db-status:
	goose postgres "$(DATABASE_URL)" -dir sql/migrations status

# Database: clean test database
db-test-clean:
	docker stop nacl_postgres_test 2>/dev/null || true
	docker rm -f nacl_postgres_test 2>/dev/null || true
	docker volume rm nacl_test_data 2>/dev/null || true
	docker run -d --name nacl_postgres_test \
		-e POSTGRES_PASSWORD=postgres \
		-v nacl_test_data:/var/lib/postgresql \
		-p 5433:5432 \
		postgres:18
	sleep 3
	docker exec nacl_postgres_test psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'nacl_test'" | grep -q 1 || \
		docker exec nacl_postgres_test psql -U postgres -c "CREATE DATABASE nacl_test;"
	goose postgres "$(DATABASE_URL_TEST)" -dir sql/migrations up

# Build: compile Go binary
build:
	go build -o bin/nacl_backend ./...

# Run: run compiled binary
run: build
	./bin/nacl_backend

# Test: run tests with test database
test: db-test-clean
	sleep 1
	go test ./... -v -json | tparse -all -smallscreen

# Clean: remove build artifacts
clean:
	rm -rf bin/ tmp/

# Lint: run ci linter
lint:
	golangci-lint run

# Format: auto-fix Go formatting
fmt:
	gofmt -w .

# Format check: fail if unformatted files exist (mirrors CI)
fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Unformatted files:" && gofmt -l . && exit 1)

# CI: run the same lint and format checks as CI
ci: fmt-check lint
