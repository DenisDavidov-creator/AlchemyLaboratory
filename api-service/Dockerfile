FROM golang:1.25.7-alpine AS builder

WORKDIR /app
COPY shared/ ./shared/
COPY api-service/ ./api-service/

WORKDIR /app/api-service

RUN go mod download

RUN go build -o main ./cmd/

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api-service/main .
COPY --from=builder /app/api-service/.env .  

CMD ["./main"]

