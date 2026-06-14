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

make run-accrual:
	./cmd/accrual/accrual_darwin_arm64 -a localhost:8081