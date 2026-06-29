package database

import (
	"backend/src/internal/logger"
	"context"
	"os"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func ConnectRedis() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		panic("REDIS_ADDR environment variable not set")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		logger.Error("failed to connect to redis", zap.String("addr", addr), zap.Error(err))
		panic("failed to connect to redis")
	}

	logger.Info("redis connected", zap.String("addr", addr))
	return client
}
