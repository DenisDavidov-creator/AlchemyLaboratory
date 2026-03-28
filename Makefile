include .env
export 


# initialization
init:
	cp .env.example .env
	cp .env.example api-service/.env
	cp .env.example db-service/.env
	cp .env.example worker-service/.env

# go mod tidy
go-mod-tidy:
	cd api-service && go mod tidy
	cd db-service && go mod tidy
	cd worker-service && go mod tidy
	cd shared && go mod tidy

# docker 
up:
	docker compose up -d --build
up-db:
	docker compose up -d --build db-service
up-api:
	docker compose up -d --build api-service
up-worker:
	docker compose up -d --build worker-service

down: 
	docker compose down 

# logs
logs-db:
	docker compose logs db-service
logs-api:
	docker compose logs api-service
logs-worker:
	docker compose logs worker-service

# Migrations
create-migrate:
	migrate create -ext sql -dir db-service/internal/db/migrations -seq create_seeds

migrate-up:
	migrate -database ${DB_PATH} -path db/migrations up

migrate-down:
	migrate -database ${DB_PATH} -path db/migrations down


# mock generation 
mock:
	cd api-service && go generate ./...
	cd db-service && go generate ./...
	cd worker-service && go generate ./...


# testing

test:
	cd api-service && go test ./...
	cd db-service && go test ./...
	cd worker-service && go test ./...

test-cover:
	cd api-service && go test -cover ./...
	cd db-service && go test -cover ./...
	cd worker-service && go test -cover ./...

test-cover-html:
	cd api-service && go test -coverprofile=../coverage-api.out ./... && go tool cover -html=../coverage-api.out -o ../coverage-api.html
	cd db-service && go test -coverprofile=../coverage-db.out ./... && go tool cover -html=../coverage-db.out -o ../coverage-db.html
	cd worker-service && go test -coverprofile=../coverage-worker.out ./... && go tool cover -html=../coverage-worker.out -o ../coverage-worker.html
	rm -f coverage-api.out coverage-db.out coverage-worker.out

# swagger
swagger:
	cd api-service && swag init -g cmd/main.go -o ./docs --parseDependency --parseInternal


# protofiles 
proto:
	protoc --go_out=. --go-grpc_out=. shared/proto/*.proto