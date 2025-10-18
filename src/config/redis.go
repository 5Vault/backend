package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis() *redis.Client {
	var RedisClient *redis.Client
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "redis-15242.c240.us-east-1-3.ec2.redns.redis-cloud.com:15242",
		Password: os.Getenv("PASSWORD_REDIS"), // no password set
		DB:       0,                           // use default DB
	})

	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	fmt.Println("Successfully connected to Redis!")
	return RedisClient
}
