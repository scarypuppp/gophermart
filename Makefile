env:
	cp .env.example .env

build:
	go build -o ./cmd/gophermart/server ./cmd/gophermart/

run:
	docker compose up

