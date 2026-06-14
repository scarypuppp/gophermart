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

run-accrual:
	./cmd/accrual/accrual_darwin_arm64 -a :8081

run-mock-accrual:
	python3 -m scripts.accrual_mock

generate-orders:
	python3 ./scripts/gen_orders.py --login=user --password=password

swagger-generate:
	swag init -g cmd/gophermart/main.go -o docs
