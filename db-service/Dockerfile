FROM golang:1.25.7-alpine AS builder

WORKDIR /app

COPY shared/ ./shared/
COPY db-service/ ./db-service/

WORKDIR /app/db-service

RUN go mod download

RUN go build -o main ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/db-service/main .
COPY --from=builder /app/db-service/.env . 

CMD ["./main"]