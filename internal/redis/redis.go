package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client
var Ctx = context.Background()

// InitRedis initializes Redis connection
func InitRedis(addr, password string, db int) (*redis.Client, error) {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       db,       // use default DB
	})

	// Test connection
	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Println("Redis connection established.")
	return Rdb, nil
}

// GetRedis returns the initialized Redis client instance
func GetRedis() *redis.Client {
	return Rdb
} 