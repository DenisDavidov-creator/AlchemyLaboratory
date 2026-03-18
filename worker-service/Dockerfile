FROM golang:1.25.7-alpine AS builder

WORKDIR /app

COPY shared/ ./shared/
COPY worker-service/ ./worker-service/

WORKDIR /app/worker-service

RUN go mod download
RUN go build -o main ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/worker-service/main .
COPY --from=builder /app/worker-service/.env .

CMD ["./main"]