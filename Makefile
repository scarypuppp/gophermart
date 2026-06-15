env:
	cp .env.example .env

build:
	go build -o ./cmd/gophermart/gophermart ./cmd/gophermart/

docker-build:
	docker compose build

run:
	docker compose up

stop:
	docker compose stop

run-dev:
	go run ./cmd/gophermart/main.go

run-mock-accrual:
	python3 -m scripts.accrual_mock

generate-orders:
	python3 ./scripts/gen_orders.py --login=user --password=password

generate-swagger:
	swag init -g cmd/gophermart/main.go -o docs
