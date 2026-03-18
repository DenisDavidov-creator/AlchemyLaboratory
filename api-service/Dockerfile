FROM golang:1.25.7-alpine AS builder

# Set the working directory to /app
WORKDIR /app

# 1. Copy the shared module first
# We copy it to /app/shared so it sits "next to" the service
COPY shared/ ./shared/

# 2. Copy the api-service
# We copy it to /app/api-service
COPY api-service/ ./api-service/

# 3. Change directory into the service
WORKDIR /app/api-service

# 4. Download dependencies
# Because we copied the files exactly as they are on your disk, 
# the "replace => ../shared" line in go.mod will work perfectly inside Docker!
RUN go mod download

# 5. Build the binary
RUN go build -o main ./cmd/main.go

# --- Stage 2: Run ---
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api-service/main .
COPY --from=builder /app/api-service/.env .  
# (Optional: Copy .env if you aren't using docker-compose environment vars)

CMD ["./main"]

