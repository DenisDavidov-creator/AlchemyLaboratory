include .env
export 

run:
	go run cmd/main.go 

create-migrate:
	migrate create -ext sql -dir db/migrations -seq create_initial_schema

migrate-up:
	migrate -database ${DB_PATH} -path db/migrations up

migrate-down:
	migrate -database ${DB_PATH} -path db/migrations down

test-web:
	go test -coverprofile=c.out ./...
	go tool cover -html=c.out -o coverage.html

test:
	go test -v ./...

test-cover:
	go test -v ./... -cover