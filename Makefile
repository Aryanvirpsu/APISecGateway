up:
	docker compose up -d

down:
	docker compose down

run-gateway:
	cd gateway && go run cmd/gateway/main.go

run-api:
	cd demo-api && go run cmd/api/main.go
