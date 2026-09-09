.PHONY: run build dev sqlc migrate-up migrate-down migrate-version migrate-force docs test tidy clean

# Konfigurasi
BINARY_NAME = news-api
DB_URL ?= postgres://news:news_secret_2026@localhost:5433/news_db?sslmode=disable
MIGRATE ?= $(shell which migrate)
SWAG ?= $(shell which swag)

## Jalankan server (dengan .env)
run:
	go run ./cmd/server

## Build binary
build:
	go build -o bin/$(BINARY_NAME) ./cmd/server

## Jalankan server dari binary hasil build
dev: build
	./bin/$(BINARY_NAME)

## Regenerate sqlc code setelah edit db/queries/
sqlc:
	cd db && sqlc generate

## Regenerate Swagger docs setelah edit handler
docs:
	@if [ -z "$(SWAG)" ]; then \
		echo "swag belum terinstall. Install: go install github.com/swaggo/swag/cmd/swag@latest"; \
		exit 1; \
	fi
	swag init -g cmd/server/main.go -o docs

## Jalankan semua migrations
migrate-up:
	@if [ -z "$(MIGRATE)" ]; then \
		echo "golang-migrate belum terinstall. Install: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	migrate -path db/migrations -database "$(DB_URL)" up

## Rollback 1 migration
migrate-down:
	@if [ -z "$(MIGRATE)" ]; then \
		echo "golang-migrate belum terinstall. Install: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	migrate -path db/migrations -database "$(DB_URL)" down 1

## Cek versi migration saat ini
migrate-version:
	@if [ -z "$(MIGRATE)" ]; then \
		echo "golang-migrate belum terinstall. Install: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	migrate -path db/migrations -database "$(DB_URL)" version

## Paksa set versi (gunakan hati-hati!)
migrate-force:
	@if [ -z "$(MIGRATE)" ]; then \
		echo "golang-migrate belum terinstall. Install: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	migrate -path db/migrations -database "$(DB_URL)" force $(VERSION)

## Test
test:
	go test ./... -v

## Bersihkan dependency yang tidak terpakai
tidy:
	go mod tidy

## Bersihkan binary
clean:
	rm -rf bin/
