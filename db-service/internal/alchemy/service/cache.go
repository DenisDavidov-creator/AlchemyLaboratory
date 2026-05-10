package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:generate mockery --name=RedisClientInterface
type RedisClientInterface interface {
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}
