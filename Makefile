.PHONY: dev-server dev-web build-server build-web docker-up docker-down

dev-server:
	cd server && go run cmd/server/main.go

dev-web:
	cd web && pnpm dev:h5

build-server:
	cd server && go build -o bin/server cmd/server/main.go

build-web:
	cd web && pnpm build:h5

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
