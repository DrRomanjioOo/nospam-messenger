.PHONY: up down test test-unit test-integration

up:
	docker compose up --build -d

down:
	docker compose down

test-unit:
	go test ./... -count=1

test-integration:
	docker compose up -d postgres redis
	docker compose run --rm test-integration

test:
	docker compose up -d postgres redis
	docker compose run --rm test
