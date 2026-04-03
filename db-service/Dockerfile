FROM golang:1.25.7-alpine AS builder

WORKDIR /app

COPY shared/ ./shared/
COPY db-service/ ./db-service/

WORKDIR /app/db-service

RUN go mod download

RUN go build -o main ./cmd/main.go
RUN wget -q -O /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.47/grpc_health_probe-linux-amd64 \
    && chmod +x /bin/grpc_health_probe

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/db-service/main .
COPY --from=builder /app/db-service/.env . 
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
CMD ["./main"]