.PHONY: up down tidy build run-api run-gateway test demo

up:
	docker compose up -d

down:
	docker compose down

tidy:
	cd gateway && go mod tidy
	cd demo-api && go mod tidy

build:
	cd gateway && go build ./...
	cd demo-api && go build ./...

run-api:
	cd demo-api && go run ./cmd/api

run-gateway:
	cd gateway && go run ./cmd/gateway

test:
	bash tests/smoke.sh
	bash tests/idor.sh
	bash tests/rate_abuse.sh

demo:
	powershell -ExecutionPolicy Bypass -File scripts/demo.ps1
