include .env
export 


# initialization
init:
	cp .env.example .env
	cp .env.example api-service.env
	cp .env.example db-service.env
	cp .env.example worker-service.env

# docker 
up:
	docker compose up -d --build

down: 
	docker compose down 

# Migrations
create-migrate:
	migrate create -ext sql -dir db/migrations -seq create_initial_schema

migrate-up:
	migrate -database ${DB_PATH} -path db/migrations up

migrate-down:
	migrate -database ${DB_PATH} -path db/migrations down


# mock generation 
mock:
	cd api-service && go generate ./...
	cd db-service && go generate ./...
	cd worker-service && go generate ./...


# testing in html 
test-cover:
	cd api-service && go test -coverprofile=../coverage-api.out ./... && go tool cover -html=../coverage-api.out -o ../coverage-api.html
	cd db-service && go test -coverprofile=../coverage-db.out ./... && go tool cover -html=../coverage-db.out -o ../coverage-db.html
	cd worker-service && go test -coverprofile=../coverage-worker.out ./... && go tool cover -html=../coverage-worker.out -o ../coverage-worker.html
	rm -f coverage-api.out coverage-db.out coverage-worker.out

# swagger
swagger:
	cd api-service && swag init -g cmd/main.go -o ./docs --parseDependency --parseInternal