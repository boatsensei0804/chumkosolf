# Makefile — คำสั่งช่วยพัฒนา backend
# DATABASE_URL ใช้ host.docker.internal ให้ migrate (ใน container) ต่อ postgres บนเครื่อง
DATABASE_URL ?= postgres://chumko:chumko_dev_pwd@host.docker.internal:5432/chumko?sslmode=disable
MIGRATE_IMAGE := migrate/migrate:v4.17.1

.PHONY: help test vet lint build run tidy seed-admin migrate-up migrate-down migrate-create

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: ## รัน unit/integration test
	go test ./...

vet: ## go vet
	go vet ./...

lint: ## golangci-lint
	golangci-lint run

build: ## build ทุก binary
	go build ./...

run: ## รัน API server
	go run ./cmd/api

tidy: ## จัดการ go.mod
	go mod tidy

seed-admin: ## สร้าง super admin (อ่าน SEED_ADMIN_PASSWORD/USERNAME จาก env)
	go run ./cmd/seed

migrate-up: ## รัน migration ขึ้นทั้งหมด
	docker run --rm -v $(CURDIR)/migrations:/migrations $(MIGRATE_IMAGE) \
		-path=/migrations -database "$(DATABASE_URL)" up

migrate-down: ## ถอย migration ลง 1 ขั้น
	docker run --rm -v $(CURDIR)/migrations:/migrations $(MIGRATE_IMAGE) \
		-path=/migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## สร้างไฟล์ migration ใหม่: make migrate-create name=create_xxx
	docker run --rm -v $(CURDIR)/migrations:/migrations $(MIGRATE_IMAGE) \
		create -ext sql -dir /migrations -seq $(name)
