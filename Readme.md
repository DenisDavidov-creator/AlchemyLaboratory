# Alchemical Laboratory
## a potion-brewing system 

## Architecture

![alt text](images/schema.png)

## Tech Stack
- Go 1.25.7
- PostgreSQL 17
- Docker / docker-compose
- sqlx, testcontainers-go, testify, mockery
- gRPC
- Redis
- Kafka (franz-go)

## How to run
```
make init      #initialization .env by example
make up
```
## swagger location 

[local](http://localhost:8090/)


## API Endpoints
- POST /ingredients
- GET  /ingredients
- PATCH  /ingredients/{id}
- POST /recipes
- GET  /recipes
- POST /brew
- GET  /brew/status

## Testing

```
make test                # run tests
```
```
make test-cover          # run with coverage report
```
```
make test-cover-html     # run with coverage report in html
```

## Roadmap
- [x] Monolith — single service, single database
- [x] Divide into microservices (api / db / worker)
- [x] Unit and integration tests
- [x] Docker containerization + graceful shutdown
- [x] Swagger documentation
- [x] HTTP → gRPC between services
- [x] Redis caching for recipes
- [x] Kafka for async worker jobs
- [x] Kafka retry + DLQ
- [x] gRPC health checks

## Extra tasks
- [x] Structured logging (slog)
- [x] Prometheus + Grafana metrics
- [x] Test coverage ≥ 60%
- [x] CI/CD pipeline (GitHub Actions) (only CI)
- [x] Stress tests (k6)



Assumptions:
 - When the API sends data to Kafka, we need to cover it into Transactional Outbox, or just when we can't send there set the job status to "failed"