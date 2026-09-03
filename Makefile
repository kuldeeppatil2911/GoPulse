.PHONY: run test build lint docker-up docker-down migrate-up migrate-down gen-graphql

run:
	go run cmd/server/main.go

test:
	go test -v ./...

build:
	go build -o bin/server cmd/server/main.go

lint:
	go vet ./...
	go fmt ./...

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

migrate-up:
	migrate -path migrations -database "${DATABASE_URL}" up

migrate-down:
	migrate -path migrations -database "${DATABASE_URL}" down

gen-graphql:
	go run github.com/99designs/gqlgen generate
