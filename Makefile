env:
	cp .env.example .env

build:
	go build -o ./cmd/gophermart/gophermart ./cmd/gophermart/


docker-build:
	docker compose build
docker-run:
	docker compose up

run-dev:
	go run ./cmd/gophermart/main.go
